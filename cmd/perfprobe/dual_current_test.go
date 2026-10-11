package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestDualCurrentProfilesPinHistoricalDiagnosticInputs(t *testing.T) {
	for _, test := range []struct {
		profile, config, inputHash, metadataHash string
	}{
		{"logn13", "../../configs/bootstrap_config.logN13.json", dualCurrentInput13, dualCurrentMetadata13},
		{"logn16", "../../configs/bootstrap_config.logN16.json", dualCurrentInput16, dualCurrentMetadata16},
	} {
		t.Run(test.profile, func(t *testing.T) {
			residual, params, effective, values, configHash, qpHash, err := loadDualCurrentProfile(dualCurrentOptions{profile: test.profile, config: test.config})
			if err != nil {
				t.Fatal(err)
			}
			ct, err := encodeDualCurrentInput(residual, effective.LogSlots, values)
			if err != nil {
				t.Fatal(err)
			}
			state := stateOf(ct, params.BootstrappingParameters.Q(), ct.Level()+1)
			metadataHash := perfmeasure.CiphertextStateFingerprint(state)
			if got := perfmeasure.Fingerprint(values); got != test.inputHash {
				t.Fatalf("input fingerprint=%s, want %s", got, test.inputHash)
			}
			if metadataHash != test.metadataHash {
				t.Fatalf("metadata fingerprint=%s, want %s", metadataHash, test.metadataHash)
			}
			if err := validateDualCurrentInput(test.profile, configHash, test.inputHash, metadataHash, effective, ct); err != nil {
				t.Fatal(err)
			}
			if qpHash == "" || ct.Value[1].Coeffs[0][0] != 0 {
				t.Fatal("Q/P fingerprint missing or diagnostic c1 is not zero")
			}
		})
	}
}

func TestDualCurrentPairRejectsBackendOrPrimaryIdentityMismatch(t *testing.T) {
	standard := dualCurrentDocument{
		Status: "PREFLIGHT_PASS", Mode: "zero_call_public_stage_preflight", Backend: "standard",
		BackendCommit: dualCurrentStandardCommit, PrimaryClean: true, BackendClean: true,
		PrimaryCommit: "primary-sha", PrimarySourceSHA256: "source-hash", Profile: "logn13",
		ConfigSHA256: "config", InputSHA256: "input", InputMetadataSHA256: "metadata", QPSHA256: "qp",
		Parameters: perfmeasure.EffectiveParameters{LogN: 13, QPrimes: []string{"q0"}, PPrimes: []string{"p0"}},
	}
	fast := standard
	fast.Backend, fast.BackendCommit = "fast", dualCurrentFastCommit
	if err := validateDualCurrentPairInputs(standard, fast); err != nil {
		t.Fatalf("matching independent sources rejected: %v", err)
	}
	fast.BackendCommit = dualCurrentStandardCommit
	if err := validateDualCurrentPairInputs(standard, fast); err == nil {
		t.Fatal("a Fast artifact carrying the Standard source commit was accepted")
	}
	fast.BackendCommit = dualCurrentFastCommit
	fast.PrimarySourceSHA256 = "other-source"
	if err := validateDualCurrentPairInputs(standard, fast); err == nil {
		t.Fatal("different Primary measurement sources were accepted")
	}
}

func TestDualCurrentAttemptReservationIsExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attempt.jsonl")
	opts := dualCurrentOptions{profile: "logn13", backendCommit: dualCurrentStandardCommit}
	journal, err := reserveDualCurrentAttempt(path, opts, dualCurrentPair{}, "metadata")
	if err != nil {
		t.Fatal(err)
	}
	if err := finishDualCurrentAttempt(journal, "COMPLETED", "test-only reservation"); err != nil {
		t.Fatal(err)
	}
	if _, err := reserveDualCurrentAttempt(path, opts, dualCurrentPair{}, "metadata"); err == nil {
		t.Fatal("a second attempt reused the same exclusive journal")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("attempt journal is empty")
	}
}

func TestDualCurrentCheckpointCompareKeepsRealAndImagBranchesSeparate(t *testing.T) {
	standard := []dualCurrentCheckpoint{
		{Name: "c2s_real", DecodedValues: encodeVector([]complex128{1, 2})},
		{Name: "c2s_imag", DecodedValues: encodeVector([]complex128{3, 4})},
	}
	fast := []dualCurrentCheckpoint{
		{Name: "c2s_real", DecodedValues: encodeVector([]complex128{1.01, 2})},
		{Name: "c2s_imag", DecodedValues: encodeVector([]complex128{3, 4.01})},
	}
	comparisons := compareDualCurrentCheckpoints(standard, fast)
	if len(comparisons) != 2 || comparisons[0].Name != "c2s_real" || comparisons[1].Name != "c2s_imag" {
		t.Fatalf("unexpected paired branches: %+v", comparisons)
	}
	if !comparisons[0].Comparable || !comparisons[1].Comparable || comparisons[0].SNR.NoiseRMSE == nil || comparisons[1].SNR.NoiseRMSE == nil || comparisons[0].FastVsStd.ComplexRMSE == comparisons[1].FastVsStd.ComplexRMSE {
		t.Fatalf("branch-local metrics were not computed independently: %+v", comparisons)
	}
}

func TestDualCurrentCheckpointCompareRequiresMatchingLogicalMetadata(t *testing.T) {
	base := dualCurrentCheckpoint{
		Name: "evalmod_real", State: ciphertextState{Level: 7, Degree: 1, Scale: "1.25e+20", LogRows: 0, LogCols: 12, QPrefixRows: 4},
		IsNTT: true, IsMontgomery: false, DecodedValues: encodeVector([]complex128{1, 2}),
	}
	for _, test := range []struct {
		name   string
		mutate func(*dualCurrentCheckpoint)
	}{
		{name: "level", mutate: func(cp *dualCurrentCheckpoint) { cp.State.Level++ }},
		{name: "scale", mutate: func(cp *dualCurrentCheckpoint) { cp.State.Scale = "1.2501e+20" }},
		{name: "degree", mutate: func(cp *dualCurrentCheckpoint) { cp.State.Degree++ }},
		{name: "dimensions", mutate: func(cp *dualCurrentCheckpoint) { cp.State.LogCols++ }},
	} {
		t.Run(test.name, func(t *testing.T) {
			fast := base
			fast.DecodedValues = append([]complexValue(nil), base.DecodedValues...)
			test.mutate(&fast)
			comparisons := compareDualCurrentCheckpoints([]dualCurrentCheckpoint{base}, []dualCurrentCheckpoint{fast})
			if len(comparisons) != 1 || comparisons[0].Comparable || comparisons[0].LogicalStateMatched {
				t.Fatalf("metadata mismatch was accepted as comparable: %+v", comparisons)
			}
		})
	}

	// Q-prefix width is intentionally backend-specific and must not make
	// otherwise equivalent decoded public stages incomparable.
	fast := base
	fast.State.QPrefixRows = 16
	fast.State.PrefixQ = "different full-Q product"
	comparisons := compareDualCurrentCheckpoints([]dualCurrentCheckpoint{base}, []dualCurrentCheckpoint{fast})
	if len(comparisons) != 1 || !comparisons[0].Comparable || !comparisons[0].LogicalStateMatched {
		t.Fatalf("backend-specific Q-prefix metadata blocked comparison: %+v", comparisons)
	}

	// Backend-specific NTT/Montgomery flags are reported but may differ when
	// each backend adapter validly decodes its own storage representation.
	fast = base
	fast.IsMontgomery = true
	comparisons = compareDualCurrentCheckpoints([]dualCurrentCheckpoint{base}, []dualCurrentCheckpoint{fast})
	if len(comparisons) != 1 || !comparisons[0].Comparable || !comparisons[0].LogicalStateMatched || comparisons[0].RepresentationFlagsMatched {
		t.Fatalf("valid backend-specific storage flags blocked semantic comparison: %+v", comparisons)
	}
}

func TestDualCurrentLogN16StandardAuditRequiresLevel16ThenLevel15FullRows(t *testing.T) {
	shape := func(name string, level, rows int) dualCurrentCheckpoint {
		checkpoint := dualCurrentCheckpoint{Name: name, State: ciphertextState{Level: level, QPrefixRows: rows}}
		checkpoint.PhysicalRowLengths = make([][]int, 2)
		checkpoint.ActiveRowSHA256 = make([][]string, 2)
		for component := 0; component < 2; component++ {
			checkpoint.PhysicalRowLengths[component] = make([]int, rows)
			checkpoint.ActiveRowSHA256[component] = make([]string, rows)
			for row := 0; row < rows; row++ {
				checkpoint.PhysicalRowLengths[component][row] = 1 << 16
				checkpoint.ActiveRowSHA256[component][row] = "row-hash"
			}
		}
		for row := 0; row < rows; row++ {
			checkpoint.ActiveRowLabels = append(checkpoint.ActiveRowLabels, fmt.Sprintf("q%d", row))
		}
		return checkpoint
	}
	doc := dualCurrentDocument{Backend: "standard", Profile: "logn16", Parameters: perfmeasure.EffectiveParameters{RingN: 1 << 16, QPrimes: make([]string, 17)}, C2SAudit: dualCurrentC2SAudit{
		Applicable: true, MatrixGroupSize: 1, Before: shape("modup", 16, 17),
		AfterRescale: shape("rescale", 15, 16), AfterNextMatrix: shape("next", 15, 16),
	}}
	if err := validateDualCurrentStandardCapacityAudit(doc); err != nil {
		t.Fatalf("complete q0..q16 ModUp / q0..q15 post-Rescale audit rejected: %v", err)
	}
	doc.C2SAudit.Before.PhysicalRowLengths[1][16] = 0
	if err := validateDualCurrentStandardCapacityAudit(doc); err == nil {
		t.Fatal("Standard Level16 ModUp audit accepted an empty q16 row")
	}
}
