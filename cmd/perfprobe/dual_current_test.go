package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
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
	standard, fast := dualCurrentIntegrityPair(t)
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

func TestDualCurrentComparatorsRejectDecodedIntegrityFailures(t *testing.T) {
	standard, fast := dualCurrentIntegrityPair(t)
	if err := validateDualCurrentPairInputs(standard, fast); err != nil {
		t.Fatalf("valid independent artifacts rejected: %v", err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*dualCurrentDocument)
	}{
		{name: "wrong decoded vector hash", mutate: func(doc *dualCurrentDocument) { doc.Checkpoints[0].DecodedSHA256 = strings.Repeat("0", 64) }},
		{name: "wrong slot count", mutate: func(doc *dualCurrentDocument) {
			doc.Checkpoints[0].DecodedValues = doc.Checkpoints[0].DecodedValues[:len(doc.Checkpoints[0].DecodedValues)-1]
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := standard
			mutated.Checkpoints = append([]dualCurrentCheckpoint(nil), standard.Checkpoints...)
			test.mutate(&mutated)
			if err := validateDualCurrentPairInputs(mutated, fast); err == nil {
				t.Fatal("preflight comparator accepted corrupted decoded evidence")
			}
		})
	}
}

func TestDualCurrentComparatorsRejectWrongBackendAndReusedTerminalSource(t *testing.T) {
	standard, fast := dualCurrentIntegrityPair(t)
	wrongSource := fast
	wrongSource.BackendCommit = dualCurrentStandardCommit
	if err := validateDualCurrentPairInputs(standard, wrongSource); err == nil {
		t.Fatal("preflight comparator accepted a Fast artifact with the Standard source pin")
	}
	wrongSource = fast
	wrongSource.LattigoModulePath = standard.LattigoModulePath
	if err := validateDualCurrentPairInputs(standard, wrongSource); err == nil {
		t.Fatal("preflight comparator accepted a reused Standard module source path")
	}

	standardBootstrap, fastBootstrap := standard, fast
	standardBootstrap.Status, fastBootstrap.Status = "BOOTSTRAP_COMPLETE", "BOOTSTRAP_COMPLETE"
	standardBootstrap.Mode, fastBootstrap.Mode = "single_cold_public_bootstrap", "single_cold_public_bootstrap"
	standardBootstrap.BootstrapCalls, fastBootstrap.BootstrapCalls = 1, 1
	standardBootstrap.PreflightPairSHA256, fastBootstrap.PreflightPairSHA256 = "pair-sha", "pair-sha"
	standardBootstrap.OutputVsInput, fastBootstrap.OutputVsInput = &vectorMetrics{}, &vectorMetrics{}
	standardBootstrap.OutputAgainstInput, fastBootstrap.OutputAgainstInput = &numericalmetrics.SNR{}, &numericalmetrics.SNR{}
	standardBootstrap.Checkpoints = standardBootstrap.Checkpoints[:1]
	fastBootstrap.Checkpoints = fastBootstrap.Checkpoints[:1]
	standardBootstrap.Checkpoints[0].Name, fastBootstrap.Checkpoints[0].Name = "bootstrap_output", "bootstrap_output"
	input := perfmeasure.DeterministicInput(1 << 12)
	for _, doc := range []*dualCurrentDocument{&standardBootstrap, &fastBootstrap} {
		values := decodeDualValues(doc.Checkpoints[0].DecodedValues)
		metrics, snr := compareVectors(input, values, 1e-7), numericalmetrics.Compare(input, values)
		doc.OutputVsInput, doc.OutputAgainstInput = &metrics, &snr
	}

	root := t.TempDir()
	standardPath, fastPath := filepath.Join(root, "standard.json"), filepath.Join(root, "fast.json")
	if err := os.WriteFile(standardPath, []byte("standard"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fastPath, []byte("fast"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateDualCurrentBootstrapInputs(standardBootstrap, fastBootstrap, standardPath, fastPath); err != nil {
		t.Fatalf("valid terminal pair rejected: %v", err)
	}
	badMetrics := standardBootstrap
	stale := *badMetrics.OutputVsInput
	stale.ComplexRMSE = 1
	badMetrics.OutputVsInput = &stale
	if err := validateDualCurrentBootstrapInputs(badMetrics, fastBootstrap, standardPath, fastPath); err == nil {
		t.Fatal("terminal comparator accepted stale output-vs-input metrics")
	}
	if err := validateDualCurrentBootstrapInputs(fastBootstrap, standardBootstrap, standardPath, fastPath); err == nil {
		t.Fatal("terminal comparator accepted swapped Standard/Fast artifacts")
	}
	if err := validateDualCurrentBootstrapInputs(standardBootstrap, fastBootstrap, standardPath, standardPath); err == nil {
		t.Fatal("terminal comparator accepted a reused source artifact path")
	}
	reusedPath := filepath.Join(root, "same-inode.json")
	if err := os.Link(standardPath, reusedPath); err != nil {
		t.Fatal(err)
	}
	if err := validateDualCurrentBootstrapInputs(standardBootstrap, fastBootstrap, standardPath, reusedPath); err == nil {
		t.Fatal("terminal comparator accepted two paths to the same artifact inode")
	}
}

func TestDualCurrentDecodedBitwiseEqualityIsExplicit(t *testing.T) {
	standard := []complex128{complex(1, -2), complex(3, 4)}
	if !dualCurrentDecodedBitwiseEqual(standard, append([]complex128(nil), standard...)) {
		t.Fatal("identical decoded float64 values not reported bitwise equal")
	}
	changedSignZero := []complex128{complex(math.Copysign(0, -1), -2), complex(3, 4)}
	if dualCurrentDecodedBitwiseEqual(standard, changedSignZero) {
		t.Fatal("distinct IEEE-754 signed-zero encodings reported bitwise equal")
	}
}

func dualCurrentIntegrityPair(t *testing.T) (dualCurrentDocument, dualCurrentDocument) {
	t.Helper()
	values := perfmeasure.DeterministicInput(1 << 12)
	encoded := encodeVector(values)
	checkpoints := func() []dualCurrentCheckpoint {
		out := make([]dualCurrentCheckpoint, 0, 8)
		for _, name := range []string{"input", "scale_down", "mod_up", "c2s_real", "c2s_imag", "evalmod_real", "evalmod_imag", "s2c"} {
			out = append(out, dualCurrentCheckpoint{Name: name, DecodedSHA256: perfmeasure.Fingerprint(values), DecodedValues: append([]complexValue(nil), encoded...)})
		}
		return out
	}
	standard := dualCurrentDocument{
		Status: "PREFLIGHT_PASS", Mode: "zero_call_public_stage_preflight", Profile: "logn13",
		Backend: "standard", BackendCommit: dualCurrentStandardCommit, BackendPath: "/tmp/lattigo-standard", LattigoModulePath: "/tmp/lattigo-standard",
		BackendClean: true, BinarySHA256: strings.Repeat("1", 64), BuildTag: "perf_standard",
		PrimaryClean: true, PrimaryCommit: "primary", PrimarySourceSHA256: "source",
		ConfigSHA256: dualCurrentConfig13, InputSHA256: dualCurrentInput13, InputMetadataSHA256: "metadata", QPSHA256: dualCurrentQP13, InputSlots: 1 << 12,
		Parameters: perfmeasure.EffectiveParameters{
			LogN: 13, LogSlots: 12, InputSlots: 1 << 12, RingN: 1 << 13, ResidualRingN: 1 << 13,
			Q0Target: 55, Q0Bits: 55, QPrefixRowsAtMax: 4,
			QChainBits: make([]int, 17), PBits: make([]int, 5), QPrimes: make([]string, 17), PPrimes: make([]string, 5),
		}, Checkpoints: checkpoints(),
	}
	fast := standard
	fast.Backend, fast.BackendCommit = "fast", dualCurrentFastCommit
	fast.BackendPath, fast.LattigoModulePath = "/tmp/lattigo-fast", "/tmp/lattigo-fast"
	fast.BinarySHA256, fast.BuildTag = strings.Repeat("2", 64), "perf_fast"
	fast.Checkpoints = checkpoints()
	return standard, fast
}
