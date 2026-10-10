package main

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

type outputSmokeStub struct {
	name           string
	bootstrapCalls int
	decodeCalls    int
	output         *rlwe.Ciphertext
	decoded        []complex128
	bootstrapPanic any
}

func (s *outputSmokeStub) Name() string { return s.name }

func (s *outputSmokeStub) EvaluatorPath() string { return "test-evaluator" }

func (s *outputSmokeStub) DecodePath() string { return "test-decoder" }

func (s *outputSmokeStub) Bootstrap(*rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	s.bootstrapCalls++
	if s.bootstrapPanic != nil {
		panic(s.bootstrapPanic)
	}
	return s.output, nil
}

func (s *outputSmokeStub) Decode(*rlwe.Ciphertext) ([]complex128, error) {
	s.decodeCalls++
	return append([]complex128(nil), s.decoded...), nil
}

func (s *outputSmokeStub) PrefixRows(ct *rlwe.Ciphertext) (int, error) {
	return ct.Level() + 1, nil
}

func outputSmokeTestFixture(t *testing.T, backendName string, decoded []complex128) (*outputSmokeStub, *rlwe.Ciphertext, []complex128, []complex128, inputRecord, []uint64, int) {
	t.Helper()
	cfg, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	residual, params, _, err := perfmeasure.ParametersFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	original := []complex128{0.125}
	preDecoded := []complex128{0.125}
	input := ckks.NewCiphertext(residual, 1, 0)
	input.LogDimensions.Rows, input.LogDimensions.Cols = 0, 0
	q := params.BootstrappingParameters.Q()
	state := stateOf(input, q, 1)
	kind, constructor, c1Nonzero := fastDirectInputKind, fastDirectConstructor, false
	if backendName == "standard" {
		kind, constructor, c1Nonzero = standardNativeInputKind, standardNativeConstructor, true
	}
	evidence := inputRecord{Kind: kind, Constructor: constructor, State: state, MetadataSHA256: "input-metadata-sha",
		C1SHA256: "input-c1-sha", C1Nonzero: c1Nonzero, PreDecodedSHA256: perfmeasure.Fingerprint(preDecoded),
		MaxComplexDeviation: 0, InputQualityLimit: inputQualityLimit}
	output := ckks.NewCiphertext(params.BootstrappingParameters, 1, params.BootstrappingParameters.MaxLevel())
	output.LogDimensions = input.LogDimensions
	backend := &outputSmokeStub{name: backendName, output: output, decoded: decoded}
	return backend, input, original, preDecoded, evidence, q, params.BootstrappingParameters.MaxLevel()
}

func TestOutputSmokeCallsPublicBootstrapOnceAndComputesBothComparisons(t *testing.T) {
	backend, input, original, preDecoded, evidence, q, maxLevel := outputSmokeTestFixture(t, "standard", []complex128{0.13})
	result := executeOutputSmoke(backend, input, original, preDecoded, evidence, 0, maxLevel, q)
	if result.FailureStage != "" {
		t.Fatalf("output smoke failed at %s: %s", result.FailureStage, result.FailureError)
	}
	if backend.bootstrapCalls != 1 || backend.decodeCalls != 1 {
		t.Fatalf("Bootstrap/decode calls=%d/%d, want exactly 1/1", backend.bootstrapCalls, backend.decodeCalls)
	}
	if result.OutputMetadata == nil || result.OutputMetadata.DecodedSlotCount != 1 || result.OutputMetadata.ComponentCount != 2 {
		t.Fatalf("output metadata was not validated and recorded: %+v", result.OutputMetadata)
	}
	if result.OutputAgainstOriginal == nil || result.PreBootstrapVsPostBootstrap == nil {
		t.Fatal("missing output-vs-original or pre-vs-post metrics")
	}
	if math.Abs(result.OutputAgainstOriginal.Error.ComplexRMSE-0.005) > 1e-12 || math.Abs(result.PreBootstrapVsPostBootstrap.Error.MaxComplexDifference-0.005) > 1e-12 {
		t.Fatalf("unexpected decoded-domain metrics: original=%+v pre/post=%+v", result.OutputAgainstOriginal, result.PreBootstrapVsPostBootstrap)
	}
}

func TestOutputSmokeRejectsInvalidOutputMetadata(t *testing.T) {
	backend, input, original, preDecoded, evidence, q, maxLevel := outputSmokeTestFixture(t, "standard", []complex128{0.13})
	backend.output.Value = backend.output.Value[:1]
	result := executeOutputSmoke(backend, input, original, preDecoded, evidence, 0, maxLevel, q)
	if result.FailureStage != "output_metadata_validation" || backend.bootstrapCalls != 1 || backend.decodeCalls != 0 {
		t.Fatalf("invalid degree result=%+v calls bootstrap/decode=%d/%d", result, backend.bootstrapCalls, backend.decodeCalls)
	}
}

func TestOutputSmokeRejectsNonFiniteDecodedOutput(t *testing.T) {
	backend, input, original, preDecoded, evidence, q, maxLevel := outputSmokeTestFixture(t, "fast", []complex128{complex(math.NaN(), 0)})
	result := executeOutputSmoke(backend, input, original, preDecoded, evidence, 0, maxLevel, q)
	if result.FailureStage != "output_decode_validation" || backend.bootstrapCalls != 1 || result.LastVerifiedStage != "output_metadata" {
		t.Fatalf("non-finite decoded result=%+v bootstrap calls=%d", result, backend.bootstrapCalls)
	}
}

func TestOutputSmokeRecordsBootstrapPanicAtExactStage(t *testing.T) {
	backend, input, original, preDecoded, evidence, q, maxLevel := outputSmokeTestFixture(t, "standard", []complex128{0.13})
	backend.bootstrapPanic = "test panic"
	result := executeOutputSmoke(backend, input, original, preDecoded, evidence, 0, maxLevel, q)
	if result.FailureStage != "public_bootstrap" || result.BootstrapCalls != 1 || result.BootstrapStatus != "FAILED" || result.LastVerifiedStage != "input_preflight" {
		t.Fatalf("Bootstrap panic evidence=%+v", result)
	}
	if result.FailureError == "" {
		t.Fatal("Bootstrap panic detail was not captured")
	}
}

func TestOutputSmokeRejectsLegacySyntheticStandardInput(t *testing.T) {
	backend, input, original, preDecoded, evidence, q, maxLevel := outputSmokeTestFixture(t, "standard", []complex128{0.13})
	evidence.Kind = "c0=encoded-message,c1=0"
	result := executeOutputSmoke(backend, input, original, preDecoded, evidence, 0, maxLevel, q)
	if result.FailureStage != "input_provenance_validation" || backend.bootstrapCalls != 0 {
		t.Fatalf("legacy synthetic Standard input was not refused before Bootstrap: result=%+v calls=%d", result, backend.bootstrapCalls)
	}
}

func TestOutputSmokeDoesNotRequireTimingCampaignOptions(t *testing.T) {
	if err := validateExecutionLimits(cliOptions{outputSmoke: true, warmup: 0, repetitions: 1, standardTrials: 1}); err == nil {
		t.Fatal("output smoke accepted an implicit unbounded Bootstrap attempt")
	}
	if err := validateExecutionLimits(cliOptions{outputSmoke: true, warmup: 0, repetitions: 1, standardTrials: 1, bootstrapBudget: 1}); err != nil {
		t.Fatalf("one-call output smoke budget was rejected: %v", err)
	}
	if err := validateExecutionLimits(cliOptions{warmup: 0, repetitions: 1, standardTrials: 1}); err == nil {
		t.Fatal("full measurement unexpectedly lost its seven-repetition precondition")
	}
	if err := validateExecutionLimits(cliOptions{warmup: 1, repetitions: 7, standardTrials: 3, bootstrapBudget: 11}); err != nil {
		t.Fatalf("bounded full measurement attempt budget was rejected: %v", err)
	}
}
