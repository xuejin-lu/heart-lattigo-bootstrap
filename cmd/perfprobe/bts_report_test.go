package main

import (
	"math"
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
