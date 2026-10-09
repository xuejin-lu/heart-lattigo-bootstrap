package main

import "testing"

func TestExactOrderedPrefix(t *testing.T) {
	full := []uint64{17, 29, 41}
	for _, tc := range []struct {
		name string
		got  []uint64
		want bool
	}{
		{name: "exact prefix", got: []uint64{17, 29}, want: true},
		{name: "full equality", got: []uint64{17, 29, 41}, want: true},
		{name: "wrong order", got: []uint64{29, 17}, want: false},
		{name: "wrong value", got: []uint64{17, 31}, want: false},
		{name: "too long", got: []uint64{17, 29, 41, 53}, want: false},
		{name: "empty", got: nil, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := exactOrderedPrefix(tc.got, full); got != tc.want {
				t.Fatalf("exactOrderedPrefix(%v, %v) = %t, want %t", tc.got, full, got, tc.want)
			}
		})
	}
}

func TestRotateLeftOracle(t *testing.T) {
	input := []complex128{1 + 2i, 3 + 4i, 5 + 6i, 7 + 8i}
	want := []complex128{3 + 4i, 5 + 6i, 7 + 8i, 1 + 2i}
	got := rotateLeftOracle(input, 1)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rotateLeftOracle[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestValidateComparablePreflightEvidence(t *testing.T) {
	standard, fast := validComparableRuns("preflight")
	if err := validateComparableRuns(standard, fast); err != nil {
		t.Fatalf("valid preflight evidence rejected: %v", err)
	}

	fast.RotateGetGaloisKeyCalls = 1
	if err := validateComparableRuns(standard, fast); err == nil {
		t.Fatal("Fast evidence with a Galois-key lookup was accepted")
	}

	standard, fast = validComparableRuns("preflight")
	fast.BackendCommit = "un-pinned"
	if err := validateComparableRuns(standard, fast); err == nil {
		t.Fatal("un-pinned Fast backend evidence was accepted")
	}
}

func TestValidateComparableBootstrapEvidenceRequiresSingleCall(t *testing.T) {
	standard, fast := validComparableRuns("bootstrap")
	if err := validateComparableRuns(standard, fast); err != nil {
		t.Fatalf("valid single-call Bootstrap evidence rejected: %v", err)
	}

	fast.BootstrapCalls = 2
	if err := validateComparableRuns(standard, fast); err == nil {
		t.Fatal("Fast evidence with more than one Bootstrap call was accepted")
	}
}

func TestSummarizeRunPreservesDispatchAndProvenance(t *testing.T) {
	standard, _ := validComparableRuns("preflight")
	standard.PrimaryDirty = true
	standard.BackendDirty = true
	standard.RotateGetGaloisKeyCalls = 1
	standard.RotateGaloisElement = rotationGaloisElement
	summary := summarizeRun(standard)
	if summary.PrimaryCommit != standard.PrimaryCommit || !summary.PrimaryDirty || summary.BackendCommit != standard.BackendCommit ||
		!summary.BackendDirty || summary.RotateGaloisElement != rotationGaloisElement || summary.EvaluationKeys.RotationGaloisElement != rotationGaloisElement ||
		summary.RotateGaloisKeyCalls != 1 || summary.BootstrapCalls != 0 {
		t.Fatalf("summary dropped provenance or dispatch evidence: %+v", summary)
	}
}

func validComparableRuns(phase string) (runEvidence, runEvidence) {
	makeRun := func(fast bool) runEvidence {
		backendCommit, backendRef, keyLayout := pinnedStandardCommit, "pinned-standard-"+pinnedStandardCommit, "implicit-standard"
		standardKeys, fastKeys, keyCalls := 30, 0, 1
		c1Zero, c1Nonzero := false, 8192
		if fast {
			backendCommit, backendRef, keyLayout = pinnedFastCommit, "fast-qprefix", "fast"
			standardKeys, fastKeys, keyCalls = 0, 30, 0
			c1Zero, c1Nonzero = true, 0
		}
		ciphertext := func() *ciphertextEvidence {
			return &ciphertextEvidence{
				Level: 0, ScaleLog2: 45, Degree: 1, N: 8192, ComponentCount: 2,
				RowsPerComponent: []int{1, 1}, ActiveRows: 1, FullActiveQ: true,
				IsNTT: true, IsBatched: true, LogDimensionsCols: 12, LogSlots: 12,
				C1Zero: c1Zero, C1NonzeroCount: c1Nonzero, C1CoefficientCount: 8192,
				C0RowSHA256: []string{"c0"}, C1RowSHA256: []string{"c1"},
			}
		}
		run := runEvidence{
			Phase: phase, Status: "preflight_passed", PrimaryCommit: "primary", FrontendSHA256: "frontend",
			BackendCommit: backendCommit, BackendRef: backendRef,
			ConfigSHA256: canonicalConfigSHA256, InputSHA256: canonicalInputSHA256,
			ResidualN: 8192, BootstrapN: 8192, ResidualRingType: "Standard", BootstrapRingType: "Standard",
			ResidualQPrimeBits: []int{55, 39}, BootstrapQPrimeBits: []int{55, 39, 40, 39, 40, 60, 60, 61, 60, 60, 60, 61, 61, 56, 57, 56, 56},
			BootstrapPPrimeBits: []int{61, 61, 62, 61, 62}, ResidualQSHA256: "residual-q", BootstrapQSHA256: "bootstrap-q", BootstrapPSHA256: "bootstrap-p",
			ResidualQIsExactBootstrapPrefix: true, Q0PrimeEqual: true, QPrimePrefixCount: 2,
			ResidualMaxLevel: 1, BootstrapMaxLevel: 16, LogSlots: 12, Slots: 4096,
			RotateK: rotation, RotateGaloisElement: rotationGaloisElement, RotateEvaluatorParameters: "btpParams.BootstrappingParameters",
			FastBootstrapSelected: fast,
			EvaluationKeys: keyEvidence{
				GeneratedBy: "btpParams.GenEvaluationKeys(sk)", GaloisKeyCount: 30,
				StandardLayoutKeyCount: standardKeys, FastLayoutKeyCount: fastKeys,
				RotationGaloisElement: rotationGaloisElement, RotationKeyPresent: true, RotationKeyLayout: keyLayout,
				RotationKeyLevelP: 4, EvaluatorMaxLevelP: 4, RelinearizationKeyPresent: true, SecretQ0Preserved: true,
				InputSecretLevelQ: 1, InputSecretLevelP: -1, BootstrapSecretLevelQ: 16, BootstrapSecretLevelP: 4,
			},
			RotateGetGaloisKeyCalls: keyCalls, InputCiphertext: ciphertext(), RotatedCiphertext: ciphertext(),
			InputUnchangedByRotate: true, ImmediateDecryption: &metric{Finite: true}, RotateVsOracle: &metric{Finite: true},
			RotatedDecodedValues: make([]complexValue, 4096),
		}
		if phase == "bootstrap" {
			run.Status = "bootstrap_completed"
			run.BootstrapCalls, run.BootstrapInputIsRotated = 1, true
			run.BootstrapOutput, run.BootstrapVsOracle = ciphertext(), &metric{Finite: true}
			run.BootstrapDecodedValues = make([]complexValue, 4096)
		}
		return run
	}
	return makeRun(false), makeRun(true)
}
