package main

import (
	"encoding/json"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestBTSReportAcceptsOnlyPinnedSixSampleRun(t *testing.T) {
	standard := validBTSReportRun("standard", bts008StandardCommit)
	if err := validateBTSRun(&standard, "standard", bts008StandardCommit); err != nil {
		t.Fatalf("valid Standard run rejected: %v", err)
	}
	fast := validBTSReportRun("fast", bts008FastCommit)
	if err := validateBTSRun(&fast, "fast", bts008FastCommit); err != nil {
		t.Fatalf("valid Fast run rejected: %v", err)
	}
	if err := validateMatchedBTSRuns(standard, fast); err != nil {
		t.Fatalf("matched runs rejected: %v", err)
	}

	fast.Samples[0].DecodedPlaintextLevel = 18
	if err := validateBTSRun(&fast, "fast", bts008FastCommit); err == nil {
		t.Fatal("Fast run with a fabricated full-level plaintext was accepted")
	}
}

func TestBTS009ReportValidatesThreeMatchedArmsAndSharedFastBinary(t *testing.T) {
	standard := validBTS009ReportRun("standard", "off")
	fastOff := validBTS009ReportRun("fast", "off")
	fastNoise := validBTS009ReportRun("fast", "on")
	writeValidBTS009NoiseEvidence(t, &fastNoise, true)
	for _, item := range []struct {
		run       *perfmeasure.BTSRunDocument
		backend   string
		noiseMode string
	}{{&standard, "standard", "off"}, {&fastOff, "fast", "off"}, {&fastNoise, "fast", "on"}} {
		if err := validateBTS009Run(item.run, item.backend, item.noiseMode); err != nil {
			t.Fatalf("valid arm rejected: backend=%s noise=%s: %v", item.backend, item.noiseMode, err)
		}
	}
	if err := validateMatchedBTS009Runs(standard, fastOff, fastNoise); err != nil {
		t.Fatalf("valid tri-arm provenance/profile rejected: %v", err)
	}

	fastNoise.BinarySHA256 = "different-binary"
	if err := validateMatchedBTS009Runs(standard, fastOff, fastNoise); err == nil {
		t.Fatal("Fast OFF and Fast+Noise with different binary identities were accepted")
	}
	fastNoise = validBTS009ReportRun("fast", "on")
	disabled := false
	fastNoise.NoiseHookEnabled = &disabled
	if err := validateBTS009Run(&fastNoise, "fast", "on"); err == nil {
		t.Fatal("Fast+Noise run with an OFF Hook snapshot was accepted")
	}
}

func TestBTS009ReportRejectsNoiseCapacityEvidenceThatDoesNotStrictlyFit(t *testing.T) {
	run := validBTS009ReportRun("fast", "on")
	writeValidBTS009NoiseEvidence(t, &run, false)
	if err := validateBTS009Run(&run, "fast", "on"); err == nil || !strings.Contains(err.Error(), "capacity evidence is invalid") {
		t.Fatalf("Noise event with failed Q-prefix capacity was accepted: %v", err)
	}
}

func TestBTS009ReportRendersColdLabelOnce(t *testing.T) {
	rendered := renderBTS009Report(bts009ReportDocument{})
	if strings.Contains(rendered, "| cold cold (ms) |") {
		t.Fatal("cold timing row repeats the cold label")
	}
	if !strings.Contains(rendered, "| cold (ms) |") {
		t.Fatal("cold timing row is missing")
	}
}

func TestBTS009CapacityEvidenceRecomputesExactPrefixAndStrictBound(t *testing.T) {
	side := bts009NoiseSide{}
	side.Capacity.Rows = 4
	side.Capacity.StrictFit = true
	side.Capacity.MaxAbs = "100"
	side.Capacity.PrefixProduct = bts009QPrefixProduct(4).String()
	if !bts009CapacityEvidenceFits(26, side) {
		t.Fatal("valid four-row Q-prefix proof was rejected")
	}

	side.Capacity.PrefixProduct = "1000"
	if bts009CapacityEvidenceFits(26, side) {
		t.Fatal("capacity evidence with a fabricated Q-prefix product was accepted")
	}

	product := bts009QPrefixProduct(4)
	tooLarge := new(big.Int).Add(new(big.Int).Rsh(new(big.Int).Set(product), 1), big.NewInt(1))
	side.Capacity.PrefixProduct = product.String()
	side.Capacity.MaxAbs = tooLarge.String()
	if bts009CapacityEvidenceFits(26, side) {
		t.Fatal("capacity evidence outside the strict centered interval was accepted")
	}
}

func validBTS009ReportRun(backend, noiseMode string) perfmeasure.BTSRunDocument {
	commit := bts009FastCommit
	if backend == "standard" {
		commit = bts008StandardCommit
	}
	run := perfmeasure.BTSRunDocument{
		SchemaVersion: perfmeasure.BTSFastNoise009SchemaVersion,
		Task:          bts009Task, Status: "COMPLETE", Backend: backend, NoiseMode: noiseMode,
		BackendCommit: commit, BinarySHA256: "fast-binary-sha", BTSManifestSHA256: bts008SourceManifestSHA,
		MeasurementAdapterSHA256: "adapter-sha", PrimaryCommit: "primary-sha", ProfileSHA256: "profile-sha",
		ArmConfigSHA256: "arm-config-sha", Circuit: "S2C -> Mod1 -> Relabel", MeasurementProcesses: 1,
		BootstrapBudget: 6, BootstrapReservations: 6, NewBTSNS: 100,
		Environment: perfmeasure.BTSMeasurementEnvironment{GoVersion: "go1.26", OS: "darwin", Arch: "arm64", CPU: "Apple Silicon", NumCPU: 8, GOMAXPROCS: 8},
		Parameters: perfmeasure.BTSMeasurementParameters{
			LogN: 16, RingN: 1 << 16, Slots: 1 << 15, SlotLimbs: 15, Mode: "BTSModePlain", Seed: 0, Amplitude: 1, EphemeralH: 32,
			MaxLevel: 29, QCount: 30, PCount: 5, Mod1Level: 26, Mod1Scale: math.Pow(2, 42), Mod1Degree: 31, Mod1PSDepth: 5, DoubleAngle: 3,
			QPrimes: append([]uint64(nil), bts009OriginalQ...), PPrimes: append([]uint64(nil), bts009OriginalP...),
			InputLevel: 3, InputScale: math.Pow(2, 32), OutputLevel: 18, OutputScale: math.Pow(2, 32),
		},
		Input: perfmeasure.BTSVectorArtifact{Encoding: "IEEE754-f64-little-endian", SHA256: bts008ExpectedInputSHA, Slots: 1 << 15, FiniteSlots: 1 << 15, Level: 3, Scale: math.Pow(2, 32)},
	}
	if noiseMode == "on" {
		run.NoiseSigma, run.NoiseRootSeed = 4.004280012097432, 20261027
	}
	hookSupported, hookEnabled := backend == "fast", backend == "fast" && noiseMode == "on"
	run.NoiseHookSupported, run.NoiseHookEnabled = &hookSupported, &hookEnabled
	plainLevel := 18
	if backend == "fast" {
		plainLevel = 3
	}
	for index := 0; index < 6; index++ {
		phase := "first_cold_bootstrap"
		if index > 0 {
			phase = "warm_bootstrap_0" + string(rune('0'+index))
		}
		run.Samples = append(run.Samples, perfmeasure.BTSBootstrapSample{
			AttemptIndex: index + 1, Phase: phase, BootstrapElapsedNS: int64(index + 1), AllocatedBytes: 100, Allocations: 2,
			InputLevel: 3, InputScale: math.Pow(2, 32), OutputCiphertextLevel: 18, OutputCiphertextScale: math.Pow(2, 32),
			DecodedPlaintextLevel: plainLevel, DecodedPlaintextScale: math.Pow(2, 32), DecodedSlots: 1 << 15, FiniteDecodedSlots: 1 << 15,
		})
	}
	return run
}

func writeValidBTS009NoiseEvidence(t *testing.T, run *perfmeasure.BTSRunDocument, strictFit bool) {
	t.Helper()
	for index, sample := range run.Samples {
		event := bts009NoiseEvent{
			Operation: uint64(index + 1), Level: 26, Scale: "4398046511104",
			MultiplyCompleted: true,
			Left:              bts009NoiseSide{Seed: bts009ExpectedNoiseSeed(20261027, uint64(index+1), "left"), SampleCount: 1 << 16, CoefficientSHA256: strings.Repeat("a", 64)},
			Right:             bts009NoiseSide{Seed: bts009ExpectedNoiseSeed(20261027, uint64(index+1), "right"), SampleCount: 1 << 16, CoefficientSHA256: strings.Repeat("b", 64)},
		}
		event.Left.Capacity.Rows, event.Right.Capacity.Rows = 4, 4
		event.Left.Capacity.CheckedCoefficients, event.Right.Capacity.CheckedCoefficients = 1<<16, 1<<16
		event.Left.Capacity.MaxAbs, event.Right.Capacity.MaxAbs = "100", "100"
		prefixProduct := bts009QPrefixProduct(4).String()
		event.Left.Capacity.PrefixProduct, event.Right.Capacity.PrefixProduct = prefixProduct, prefixProduct
		event.Left.Capacity.StrictFit, event.Right.Capacity.StrictFit = strictFit, strictFit
		evidence := bts009NoiseCall{
			SchemaVersion: "bts-fast-noise-fidelity-009-noise-call.v1", Task: bts009Task,
			Phase: sample.Phase, AttemptIndex: index + 1, Sigma: 4.004280012097432, RootSeed: 20261027,
			OperationStart: uint64(index + 1), OperationEnd: uint64(index + 1), Events: []bts009NoiseEvent{event},
		}
		data, err := json.Marshal(evidence)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "noise.json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		run.NoiseEvidence = append(run.NoiseEvidence, perfmeasure.BTSNoiseEvidenceArtifact{
			AttemptIndex: index + 1, Phase: sample.Phase, Path: path, SHA256: digestBytes(data), Bytes: int64(len(data)),
			OperationStart: uint64(index + 1), OperationEnd: uint64(index + 1),
			CtCtMultiplyCalls: 1, InjectedPolynomials: 2, EventCount: 1,
		})
	}
}

func TestBTSReportMagnitudeIncludesDecimalOrderAndZeroHandling(t *testing.T) {
	metric := magnitude(0.00032)
	if metric.Scientific != "3.20000000e-04" || metric.Log10 == nil || math.Abs(*metric.Log10+3.494850021680094) > 1e-12 || metric.OrderOfMagnitude != "10^-4" {
		t.Fatalf("unexpected nonzero metric formatting: %+v", metric)
	}
	zero := magnitude(0)
	if zero.Log10 != nil || zero.OrderOfMagnitude != "ZERO" || zero.Scientific != "0.00000000e+00" {
		t.Fatalf("unexpected zero metric formatting: %+v", zero)
	}
}

func validBTSReportRun(backend, commit string) perfmeasure.BTSRunDocument {
	run := perfmeasure.BTSRunDocument{
		SchemaVersion: perfmeasure.BTSNoTrace008SchemaVersion,
		Task:          "BTS-NOTRACE-WARM-AND-NUMERICAL-008", Status: "COMPLETE",
		Backend: backend, BackendCommit: commit, BinarySHA256: "binary-sha",
		BTSManifestSHA256: bts008SourceManifestSHA, MeasurementAdapterSHA256: "adapter-sha",
		PrimaryCommit: "primary-sha", MeasurementProcesses: 1, BootstrapBudget: 6,
		BootstrapReservations: 6,
		Environment:           perfmeasure.BTSMeasurementEnvironment{GoVersion: "go1.26", OS: "darwin", Arch: "arm64", CPU: "Apple Silicon", NumCPU: 8, GOMAXPROCS: 8},
		Parameters: perfmeasure.BTSMeasurementParameters{
			LogN: 16, RingN: 1 << 16, Slots: 1 << 15, SlotLimbs: 15, Mode: "BTSModePlain",
			Seed: 0, Amplitude: 1, EphemeralH: 32, QPrimes: []uint64{1, 2}, PPrimes: []uint64{3},
			InputLevel: 3, InputScale: math.Pow(2, 32), OutputLevel: 18, OutputScale: math.Pow(2, 32),
		},
		Input: perfmeasure.BTSVectorArtifact{Encoding: "IEEE754-f64-little-endian", SHA256: bts008ExpectedInputSHA, Slots: 1 << 15, FiniteSlots: 1 << 15, Level: 3, Scale: math.Pow(2, 32)},
	}
	plainLevel := 18
	if backend == "fast" {
		plainLevel = 3
	}
	for index := 0; index < 6; index++ {
		phase := "first_cold_bootstrap"
		if index > 0 {
			phase = "warm_bootstrap_0" + string(rune('0'+index))
		}
		run.Samples = append(run.Samples, perfmeasure.BTSBootstrapSample{
			AttemptIndex: index + 1, Phase: phase, BootstrapElapsedNS: int64(index + 1),
			InputLevel: 3, InputScale: math.Pow(2, 32), OutputCiphertextLevel: 18,
			OutputCiphertextScale: math.Pow(2, 32), DecodedPlaintextLevel: plainLevel,
			DecodedPlaintextScale: math.Pow(2, 32), DecodedSlots: 1 << 15, FiniteDecodedSlots: 1 << 15,
		})
	}
	return run
}
