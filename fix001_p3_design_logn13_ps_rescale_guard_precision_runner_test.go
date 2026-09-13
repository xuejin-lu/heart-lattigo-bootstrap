package main

import "testing"

func TestFIX001P3PSRescaleGuardPrecisionSchedule(t *testing.T) {
	boundaries := []psRescaleGuardBoundary{{ID: "G0-rescale"}, {ID: "G1-rescale"}, {ID: "F0-final-rescale"}}
	got := psRescaleGuardScheduleName([]int{0, 1, 2}, boundaries, "single")
	want := "single-G1-rescale=2^1+F0-final-rescale=2^2"
	if got != want {
		t.Fatalf("schedule name = %q, want %q", got, want)
	}
}

func TestFIX001P3PSRescaleGuardPrecisionBudget(t *testing.T) {
	if psRescaleGuardBudget != 1.2e-8 {
		t.Fatalf("budget = %g, want 1.2e-8", psRescaleGuardBudget)
	}
	if psRescaleGuardValueTarget != 1e-10 {
		t.Fatalf("value target = %g, want 1e-10", psRescaleGuardValueTarget)
	}
}

func TestFIX001P3PSRescaleGuardPrecisionClassification(t *testing.T) {
	candidate := psRescaleGuardCandidate{Valid: true, TotalGuardBits: 1, GuardedBoundaryCount: 1, FinalMaxError: 1e-8}
	other := psRescaleGuardCandidate{Valid: true, TotalGuardBits: 2, GuardedBoundaryCount: 1, FinalMaxError: 1e-12}
	if !psRescaleGuardSelectBetter(candidate, other, nil) {
		t.Fatal("fewest guard bits must win deterministic ranking")
	}
}
