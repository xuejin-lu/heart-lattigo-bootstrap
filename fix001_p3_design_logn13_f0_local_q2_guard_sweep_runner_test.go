package main

import "testing"

func TestFIX001P3F0LocalQ2GuardSweepValidity(t *testing.T) {
	evidence := fix001P3LocalQ2BranchEvidence{
		Expansion:   fix001P3LocalQ2Expansion{Pass: true},
		Guard:       fix001P3LocalQ2Guard{Value: &PSGlobalMetric{Pass: true}, Q012After: fix001P3LocalQ2Capacity{Unique: true}, RowsMatch: true},
		Rescale:     fix001P3LocalQ2Rescale{InputLevel: 8, OutputLevel: 7, RowsMatch: true, RoundedDivisionMatch: true},
		Contraction: fix001P3LocalQ2Contraction{Pass: true},
	}
	if !fix001P3LocalQ2SweepBranchValid(evidence) {
		t.Fatal("valid local-q2 evidence was rejected")
	}
	evidence.Rescale.OutputLevel = 6
	if fix001P3LocalQ2SweepBranchValid(evidence) {
		t.Fatal("extra level consumption was accepted")
	}
}

func TestFIX001P3F0LocalQ2GuardSweepIgnoresIdealRescaleThreshold(t *testing.T) {
	evidence := fix001P3LocalQ2BranchEvidence{
		Expansion:   fix001P3LocalQ2Expansion{Pass: true},
		Guard:       fix001P3LocalQ2Guard{Value: &PSGlobalMetric{Pass: true}, Q012After: fix001P3LocalQ2Capacity{Unique: true}, RowsMatch: true},
		Rescale:     fix001P3LocalQ2Rescale{InputLevel: 8, OutputLevel: 7, RowsMatch: true, RoundedDivisionMatch: true, LocalError: &PSGlobalMetric{Pass: false, MaxComponent: 3e-9}},
		Contraction: fix001P3LocalQ2Contraction{Pass: true},
	}
	if !fix001P3LocalQ2SweepBranchValid(evidence) {
		t.Fatal("ideal-real-arithmetic rounding error incorrectly rejected local-q2 candidate")
	}
}
