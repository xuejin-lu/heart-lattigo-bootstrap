package main

import "testing"

func TestFIX001P3Q056F0TwoBitGuardSchedule(t *testing.T) {
	boundaries := []psRescaleGuardBoundary{{ID: "G0-rescale"}, {ID: "G1-rescale"}, {ID: "G2-rescale"}, {ID: "G3-rescale"}, {ID: "F0-final-rescale"}}
	got := psRescaleGuardScheduleName([]int{1, 0, 0, 0, 2}, boundaries, "C3")
	want := "C3-G0-rescale=2^1+F0-final-rescale=2^2"
	if got != want {
		t.Fatalf("schedule = %q, want %q", got, want)
	}
}

func TestFIX001P3Q056F0TwoBitGuardEvidence(t *testing.T) {
	candidate := psRescaleGuardCandidate{
		Valid: true, CapacitySafe: true, AlignmentSafe: true, ValuePreserving: true, LevelSafe: true, RowsMatch: true,
		BoundaryEffects: []psRescaleGuardBoundary{{
			ID: "F0-final-rescale", GuardBits: 2, GuardedPreCapacityRatio: 0.779,
			GuardValuePreservation: &PSGlobalMetric{Pass: true, MaxComponent: 0},
			LocalRescaleError:      &PSGlobalMetric{MaxComponent: 1e-9}, PostCumulativeError: &PSGlobalMetric{MaxComponent: 2e-9},
			CenteredUnique: true, RowsMatch: true, LevelUnchanged: true,
		}},
	}
	evidence := q056F0Evidence(candidate, 2)
	if !evidence.Valid || evidence.CapacityRatio != 0.779 || !evidence.CenteredUnique || !evidence.RowsMatch || !evidence.LevelUnchanged {
		t.Fatalf("unexpected two-bit evidence: %+v", evidence)
	}
}
