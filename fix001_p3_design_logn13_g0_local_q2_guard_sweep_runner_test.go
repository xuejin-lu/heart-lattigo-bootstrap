package main

import "testing"

func TestFIX001P3G0LocalQ2GuardSweepValidity(t *testing.T) {
	evidence := fix001P3LocalQ2BranchEvidence{
		Expansion:   fix001P3LocalQ2Expansion{Pass: true},
		Guard:       fix001P3LocalQ2Guard{Value: &PSGlobalMetric{Pass: true}, Q012After: fix001P3LocalQ2Capacity{Unique: true}, RowsMatch: true},
		Rescale:     fix001P3LocalQ2Rescale{InputLevel: 8, OutputLevel: 7, RowsMatch: true, RoundedDivisionMatch: true},
		Contraction: fix001P3LocalQ2Contraction{Pass: true},
	}
	if !fix001P3LocalQ2SweepBranchValid(evidence) {
		t.Fatal("valid G0 local-q2 evidence was rejected")
	}
	evidence.Guard.Q012After.Unique = false
	if fix001P3LocalQ2SweepBranchValid(evidence) {
		t.Fatal("G0 capacity failure was accepted")
	}
}

func TestFIX001P3G0LocalQ2GuardSweepCandidates(t *testing.T) {
	got := fix001P3G0Operations()["g0_candidate_guard_bits"]
	want := []int{2, 3, 4}
	values, ok := got.([]int)
	if !ok || len(values) != len(want) {
		t.Fatalf("G0 sweep candidates = %#v, want %#v", got, want)
	}
	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("G0 sweep candidate %d = %d, want %d", i, values[i], want[i])
		}
	}
}

func TestFIX001P3G0LocalQ2GuardSweepClassification(t *testing.T) {
	valid := fix001P3G0LocalQ2GuardSweepCandidate{SystemSufficient: true}
	if !valid.SystemSufficient {
		t.Fatal("validated downstream candidate must mark system sufficient")
	}
	if g0LocalQ2PublicThreshold != 1e-2 {
		t.Fatalf("public threshold = %g, want 1e-2", g0LocalQ2PublicThreshold)
	}
}
