package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

type outputSmokeBackend interface {
	Name() string
	EvaluatorPath() string
	DecodePath() string
	Bootstrap(*rlwe.Ciphertext) (*rlwe.Ciphertext, error)
	Decode(*rlwe.Ciphertext) ([]complex128, error)
	PrefixRows(*rlwe.Ciphertext) (int, error)
}

type budgetedOutputBackend struct {
	outputSmokeBackend
	budget *bootstrapBudget
}

func (backend *budgetedOutputBackend) Bootstrap(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return backend.budget.invoke("output_smoke_bootstrap", func() (*rlwe.Ciphertext, error) { return backend.outputSmokeBackend.Bootstrap(ct) })
}

type outputCiphertextMetadata struct {
	State            ciphertextState `json:"state"`
	IsNTT            bool            `json:"is_ntt"`
	IsMontgomery     bool            `json:"is_montgomery"`
	ComponentCount   int             `json:"component_count"`
	DecodedSlotCount int             `json:"decoded_slot_count"`
}

type outputComparison struct {
	Error vectorMetrics        `json:"error_metrics"`
	SNR   numericalmetrics.SNR `json:"snr"`
}

type outputSmokeDocument struct {
	SchemaVersion               string                    `json:"schema_version"`
	Timestamp                   time.Time                 `json:"timestamp"`
	RunStatus                   string                    `json:"run_status"`
	QualityAssessment           string                    `json:"quality_assessment"`
	Profile                     string                    `json:"profile"`
	Backend                     string                    `json:"backend"`
	BuildTag                    string                    `json:"build_tag"`
	PrimaryPath                 string                    `json:"primary_path"`
	PrimaryCommit               string                    `json:"primary_commit"`
	PrimaryDirty                *bool                     `json:"primary_dirty,omitempty"`
	BackendCommit               string                    `json:"backend_commit"`
	BackendRef                  string                    `json:"backend_ref"`
	BackendCheckoutRef          string                    `json:"backend_checkout_ref"`
	BackendSource               string                    `json:"backend_source"`
	BackendDirty                *bool                     `json:"backend_dirty,omitempty"`
	ConfigPath                  string                    `json:"config_path"`
	ConfigSHA256                string                    `json:"config_sha256"`
	OriginalSHA256              string                    `json:"original_sha256"`
	Parameters                  effectiveParameters       `json:"effective_parameters"`
	InputKind                   string                    `json:"input_kind,omitempty"`
	InputConstructor            string                    `json:"input_constructor,omitempty"`
	EvaluatorPath               string                    `json:"evaluator_path,omitempty"`
	DecodePath                  string                    `json:"decode_path,omitempty"`
	InputEvidence               *inputRecord              `json:"input_evidence,omitempty"`
	InputQualityStatus          string                    `json:"input_quality_status"`
	PreBootstrapDecodedSHA256   string                    `json:"pre_bootstrap_decoded_sha256,omitempty"`
	PreBootstrapMetadata        *outputCiphertextMetadata `json:"pre_bootstrap_metadata,omitempty"`
	BootstrapStatus             string                    `json:"bootstrap_status"`
	OutputStatus                string                    `json:"output_status"`
	OutputMetadata              *outputCiphertextMetadata `json:"output_metadata,omitempty"`
	OutputDecodedSHA256         string                    `json:"output_decoded_sha256,omitempty"`
	OutputAgainstOriginal       *outputComparison         `json:"output_against_original,omitempty"`
	PreBootstrapVsPostBootstrap *outputComparison         `json:"pre_bootstrap_vs_post_bootstrap,omitempty"`
	DescriptiveComponentCutoff  float64                   `json:"descriptive_component_cutoff"`
	GoVersion                   string                    `json:"go_version"`
	OS                          string                    `json:"os"`
	Arch                        string                    `json:"arch"`
	CPU                         string                    `json:"cpu"`
	BootstrapCalls              int                       `json:"public_bootstrap_calls"`
	TimingPerformed             bool                      `json:"timing_performed"`
	LastVerifiedStage           string                    `json:"last_verified_stage,omitempty"`
	LastVerifiedMetadata        *outputCiphertextMetadata `json:"last_verified_metadata,omitempty"`
	FailureStage                string                    `json:"failure_stage,omitempty"`
	FailureError                string                    `json:"failure_error,omitempty"`
	Limitations                 []string                  `json:"limitations"`
}

type outputSmokeExecution struct {
	InputMetadata               outputCiphertextMetadata
	OutputMetadata              *outputCiphertextMetadata
	OutputDecodedSHA256         string
	OutputAgainstOriginal       *outputComparison
	PreBootstrapVsPostBootstrap *outputComparison
	LastVerifiedStage           string
	LastVerifiedMetadata        outputCiphertextMetadata
	BootstrapStatus             string
	BootstrapCalls              int
	FailureStage                string
	FailureError                string
}

func runOutputSmoke(opts cliOptions, configHash string, residual ckks.Parameters, params bootstrapping.Parameters, effective perfmeasure.EffectiveParameters, values []complex128) error {
	doc := outputSmokeDocument{
		SchemaVersion: "fast-standard-output-preflight.v1", Timestamp: time.Now().UTC(),
		RunStatus: "FAILED", QualityAssessment: "UNASSESSED", Profile: opts.profile,
		InputQualityStatus: "NOT_VERIFIED", BootstrapStatus: "NOT_STARTED", OutputStatus: "NOT_RUN",
		BackendCommit: opts.backendCommit, BackendRef: opts.backendRef,
		ConfigPath: opts.config, ConfigSHA256: configHash, OriginalSHA256: perfmeasure.Fingerprint(values),
		Parameters: effective, DescriptiveComponentCutoff: defaultComparisonThreshold,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(),
		TimingPerformed: false,
		Limitations: []string{
			"Exactly one public Bootstrap is attempted; no warmup, timing campaign, stage replay, or speedup claim is performed.",
			"Output accuracy is recorded against the original input but remains UNASSESSED; the descriptive component cutoff is not a pass/fail gate.",
		},
	}
	fail := func(stage string, cause error) error {
		doc.FailureStage, doc.FailureError = stage, cause.Error()
		doc.OutputStatus = "NOT_RUN"
		if writeErr := writeExclusiveJSON(opts.out, doc); writeErr != nil {
			return fmt.Errorf("%s failed: %w (also could not write failure evidence: %v)", stage, cause, writeErr)
		}
		return fmt.Errorf("%s failed: %w", stage, cause)
	}
	if err := verifyCompiledBackendSource(opts.secondaryRoot); err != nil {
		return fail("compiled_backend_source", err)
	}
	primaryRoot, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		return fail("Primary provenance", err)
	}
	primaryClean, backendClean := false, false
	doc.PrimaryPath, doc.PrimaryCommit, doc.PrimaryDirty = primaryRoot, primaryCommit, &primaryClean
	secondaryRoot, secondaryCommit, secondaryRef, err := cleanSecondaryState(opts.secondaryRoot, opts.backendCommit)
	if err != nil {
		return fail("Secondary provenance", err)
	}
	doc.BackendSource, doc.BackendCheckoutRef, doc.BackendDirty = secondaryRoot, secondaryRef, &backendClean
	backend, err := newBackend(params, residual)
	if err != nil {
		return fail("backend initialization", err)
	}
	doc.Backend, doc.BuildTag = backend.Name(), "perf_"+backend.Name()
	doc.BackendCommit = secondaryCommit
	if err := validatePinnedBackend(backend.Name(), secondaryCommit); err != nil {
		return fail("pinned backend validation", err)
	}
	input, preDecoded, evidence, err := prepareAndValidateInput(backend, residual, params, values, effective.LogSlots)
	if err != nil {
		return fail("input preparation and preflight", err)
	}
	doc.InputKind, doc.InputConstructor = evidence.Kind, evidence.Constructor
	doc.EvaluatorPath, doc.DecodePath = backend.EvaluatorPath(), backend.DecodePath()
	doc.InputEvidence, doc.InputQualityStatus = &evidence, "PASSED"
	doc.PreBootstrapDecodedSHA256 = evidence.PreDecodedSHA256
	guardedBackend := &budgetedOutputBackend{outputSmokeBackend: backend, budget: &bootstrapBudget{limit: opts.bootstrapBudget, journalBase: opts.out}}
	execution := executeOutputSmoke(guardedBackend, input, values, preDecoded, evidence, effective.LogSlots, params.BootstrappingParameters.MaxLevel(), params.BootstrappingParameters.Q())
	doc.BootstrapCalls, doc.BootstrapStatus = execution.BootstrapCalls, execution.BootstrapStatus
	doc.LastVerifiedStage = execution.LastVerifiedStage
	if execution.LastVerifiedStage != "configuration_validated" {
		doc.PreBootstrapMetadata = &execution.InputMetadata
		doc.LastVerifiedMetadata = &execution.LastVerifiedMetadata
	}
	if execution.OutputMetadata != nil {
		doc.OutputMetadata = execution.OutputMetadata
	}
	if execution.FailureStage != "" {
		doc.FailureStage, doc.FailureError = execution.FailureStage, execution.FailureError
		doc.OutputStatus = "FAILED"
		if execution.BootstrapCalls == 0 {
			doc.InputQualityStatus = "FAILED"
		}
		if writeErr := writeExclusiveJSON(opts.out, doc); writeErr != nil {
			return fmt.Errorf("%s failed: %s (also could not write failure evidence: %v)", execution.FailureStage, execution.FailureError, writeErr)
		}
		return fmt.Errorf("%s failed: %s", execution.FailureStage, execution.FailureError)
	}
	doc.BootstrapStatus, doc.OutputStatus, doc.RunStatus = "SUCCEEDED", "VALIDATED", "SUCCEEDED"
	doc.OutputDecodedSHA256 = execution.OutputDecodedSHA256
	doc.OutputAgainstOriginal = execution.OutputAgainstOriginal
	doc.PreBootstrapVsPostBootstrap = execution.PreBootstrapVsPostBootstrap
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		return err
	}
	fmt.Printf("%s %s output-smoke=%s status=%s quality=%s\n", opts.profile, backend.Name(), opts.out, doc.RunStatus, doc.QualityAssessment)
	return nil
}

func executeOutputSmoke(backend outputSmokeBackend, input *rlwe.Ciphertext, original, preDecoded []complex128, evidence inputRecord, logSlots, maxLevel int, q []uint64) outputSmokeExecution {
	result := outputSmokeExecution{LastVerifiedStage: "configuration_validated", BootstrapStatus: "NOT_STARTED"}
	fail := func(stage string, cause error) outputSmokeExecution {
		result.FailureStage, result.FailureError = stage, cause.Error()
		return result
	}
	if backend == nil {
		return fail("backend_validation", fmt.Errorf("output-smoke backend is nil"))
	}
	if input == nil {
		return fail("input_preflight", fmt.Errorf("prepared input is nil"))
	}
	inputState, err := outputSmokeStateOf(backend, input, q)
	if err != nil {
		return fail("input_metadata_validation", err)
	}
	result.InputMetadata = snapshotOutputMetadata(input, inputState, len(preDecoded))
	result.LastVerifiedStage, result.LastVerifiedMetadata = "input_metadata", result.InputMetadata
	if err := validateOutputSmokeInputEvidence(backend.Name(), evidence, inputState, original, preDecoded, logSlots); err != nil {
		return fail("input_provenance_validation", err)
	}
	result.LastVerifiedStage = "input_preflight"
	result.BootstrapCalls, result.BootstrapStatus = 1, "ATTEMPTED"
	output, err := bootstrapOutputSafely(backend, input)
	if err != nil {
		result.BootstrapStatus = "FAILED"
		return fail("public_bootstrap", err)
	}
	if output == nil {
		result.BootstrapStatus = "RETURNED_NIL_OUTPUT"
		return fail("output_metadata_validation", fmt.Errorf("public Bootstrap returned a nil ciphertext"))
	}
	result.BootstrapStatus = "SUCCEEDED"
	outputState, err := outputSmokeStateOf(backend, output, q)
	if err != nil {
		return fail("output_metadata_validation", err)
	}
	outputMetadata := snapshotOutputMetadata(output, outputState, 0)
	if err := validateOutputCiphertext(output, outputMetadata, maxLevel, logSlots); err != nil {
		return fail("output_metadata_validation", err)
	}
	result.OutputMetadata = &outputMetadata
	result.LastVerifiedStage, result.LastVerifiedMetadata = "output_metadata", outputMetadata
	decoded, err := decodeOutputSafely(backend, output)
	if err != nil {
		return fail("output_decode", err)
	}
	if len(decoded) != len(original) || len(decoded) != 1<<logSlots {
		return fail("output_decode_validation", fmt.Errorf("decoded slot count=%d, original=%d, expected=%d", len(decoded), len(original), 1<<logSlots))
	}
	if !allFinite(decoded) {
		return fail("output_decode_validation", fmt.Errorf("decoded Bootstrap output contains a non-finite complex slot"))
	}
	outputMetadata.DecodedSlotCount = len(decoded)
	result.OutputMetadata = &outputMetadata
	result.LastVerifiedMetadata = outputMetadata
	outputVsOriginal := makeOutputComparison(original, decoded)
	preVsPost := makeOutputComparison(preDecoded, decoded)
	if field := firstNonFinite(outputVsOriginal, "output_against_original"); field != "" {
		return fail("output_metric_validation", fmt.Errorf("non-finite output-vs-original metric at %s", field))
	}
	if field := firstNonFinite(preVsPost, "pre_vs_post"); field != "" {
		return fail("output_metric_validation", fmt.Errorf("non-finite pre-vs-post metric at %s", field))
	}
	result.OutputDecodedSHA256 = perfmeasure.Fingerprint(decoded)
	result.OutputAgainstOriginal, result.PreBootstrapVsPostBootstrap = &outputVsOriginal, &preVsPost
	result.LastVerifiedStage = "output_metrics"
	return result
}

func validateOutputSmokeInputEvidence(backend string, evidence inputRecord, state ciphertextState, original, preDecoded []complex128, logSlots int) error {
	if len(original) == 0 || len(original) != len(preDecoded) || len(original) != 1<<logSlots || !allFinite(original) || !allFinite(preDecoded) {
		return fmt.Errorf("input vectors have invalid slot count or non-finite values")
	}
	expectedKind, expectedConstructor, c1Nonzero := "", "", false
	switch backend {
	case "standard":
		expectedKind, expectedConstructor, c1Nonzero = standardNativeInputKind, standardNativeConstructor, true
	case "fast":
		expectedKind, expectedConstructor = fastDirectInputKind, fastDirectConstructor
	default:
		return fmt.Errorf("unsupported output-smoke backend %q", backend)
	}
	if evidence.Kind != expectedKind || evidence.Constructor != expectedConstructor || evidence.C1Nonzero != c1Nonzero || evidence.C1SHA256 == "" || evidence.MetadataSHA256 == "" {
		return fmt.Errorf("input evidence does not match the declared %s input constructor", backend)
	}
	if evidence.State != state || state.Level != 0 || state.Degree != 1 || state.LogCols != logSlots {
		return fmt.Errorf("prepared input metadata differs from the verified Level-0 input contract")
	}
	preHash := perfmeasure.Fingerprint(preDecoded)
	if evidence.PreDecodedSHA256 != preHash || evidence.InputQualityLimit != inputQualityLimit ||
		evidence.MaxComplexDeviation < 0 || evidence.MaxComplexDeviation > inputQualityLimit {
		return fmt.Errorf("input decode fingerprint or accepted 1e-6 quality evidence is absent or contradictory")
	}
	var maxDeviation float64
	for i := range original {
		maxDeviation = math.Max(maxDeviation, cmplx.Abs(preDecoded[i]-original[i]))
	}
	if maxDeviation > inputQualityLimit || math.Abs(maxDeviation-evidence.MaxComplexDeviation) > 1e-15 {
		return fmt.Errorf("pre-Bootstrap max complex deviation %.9g contradicts accepted input evidence %.9g", maxDeviation, evidence.MaxComplexDeviation)
	}
	return nil
}

func validateOutputCiphertext(output *rlwe.Ciphertext, metadata outputCiphertextMetadata, maxLevel, logSlots int) error {
	if output == nil || output.MetaData == nil {
		return fmt.Errorf("public Bootstrap output or metadata is nil")
	}
	if output.Level() < 0 || output.Level() > maxLevel {
		return fmt.Errorf("output Level=%d is outside the configured range [0,%d]", output.Level(), maxLevel)
	}
	if output.Degree() != 1 || len(output.Value) != output.Degree()+1 || metadata.ComponentCount != 2 {
		return fmt.Errorf("output degree/component count is invalid: degree=%d components=%d", output.Degree(), len(output.Value))
	}
	if output.Scale.Value.Sign() <= 0 || math.IsNaN(metadata.State.ScaleLog2) || math.IsInf(metadata.State.ScaleLog2, 0) || metadata.State.Scale == "" {
		return fmt.Errorf("output Scale is not finite and positive")
	}
	if metadata.State.LogCols != logSlots || metadata.State.LogRows < 0 {
		return fmt.Errorf("output slot metadata LogRows/LogCols=%d/%d, expected valid rows and LogCols=%d", metadata.State.LogRows, metadata.State.LogCols, logSlots)
	}
	if metadata.State.QPrefixRows < 1 || metadata.State.QPrefixRows > output.Level()+1 || metadata.State.PrefixQ == "" {
		return fmt.Errorf("output Q-prefix metadata is invalid: rows=%d Level=%d", metadata.State.QPrefixRows, output.Level())
	}
	return nil
}

func snapshotOutputMetadata(ct *rlwe.Ciphertext, state ciphertextState, decodedSlots int) outputCiphertextMetadata {
	if ct == nil {
		return outputCiphertextMetadata{State: state, DecodedSlotCount: decodedSlots}
	}
	return outputCiphertextMetadata{State: state, IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		ComponentCount: len(ct.Value), DecodedSlotCount: decodedSlots}
}

func makeOutputComparison(reference, observed []complex128) outputComparison {
	return outputComparison{Error: compareVectors(reference, observed, defaultComparisonThreshold), SNR: numericalmetrics.Compare(reference, observed)}
}

func outputSmokeStateOf(backend outputSmokeBackend, ct *rlwe.Ciphertext, q []uint64) (ciphertextState, error) {
	rows, err := backend.PrefixRows(ct)
	if err != nil {
		return ciphertextState{}, err
	}
	return stateOf(ct, q, rows), nil
}

func bootstrapOutputSafely(backend outputSmokeBackend, input *rlwe.Ciphertext) (output *rlwe.Ciphertext, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			output, err = nil, fmt.Errorf("panic during public Bootstrap: %v", recovered)
		}
	}()
	return backend.Bootstrap(input.CopyNew())
}

func decodeOutputSafely(backend outputSmokeBackend, output *rlwe.Ciphertext) (decoded []complex128, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			decoded, err = nil, fmt.Errorf("panic during backend output decode: %v", recovered)
		}
	}()
	return backend.Decode(output)
}
