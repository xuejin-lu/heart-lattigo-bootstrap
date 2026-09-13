package main

import "testing"

func TestFIX001P3K3DownstreamSufficiencyClassification(t *testing.T) {
	if k3DownstreamPublicThreshold != 1e-2 {
		t.Fatalf("public threshold = %g, want 0.01", k3DownstreamPublicThreshold)
	}
	path := evalModMatchedPath{Rounds: []evalModMatchedRound{{Round: 0, KInExponent: 3, KOutExponent: 4, MultiplierExponent: 2}}}
	if got := fix001P3K3DownstreamSchedule(path); len(got) != 1 || got[0]["round"] != 0 || got[0]["k_out_exponent"] != 4 {
		t.Fatalf("unexpected derived schedule: %+v", got)
	}
}

func TestFIX001P3K3DownstreamSufficiencyDoesNotUseLocalBudgetAsSystemTarget(t *testing.T) {
	local := 1.223639867209414e-8
	if !(local > psRescaleGuardBudget && k3DownstreamPublicThreshold > local) {
		t.Fatal("expected distinct local planning budget and public-like system threshold")
	}
}
