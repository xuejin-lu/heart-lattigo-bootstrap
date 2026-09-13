package main

import "testing"

func TestFIX001P3PSOracleScalePrecisionExponentSet(t *testing.T) {
	want := []int{91, 92, 93, 94, 95, 96}
	if len(psOracleScalePlanExponents) != len(want) {
		t.Fatalf("candidate count = %d, want %d", len(psOracleScalePlanExponents), len(want))
	}
	for i, exponent := range want {
		if psOracleScalePlanExponents[i] != exponent {
			t.Fatalf("candidate[%d] = %d, want %d", i, psOracleScalePlanExponents[i], exponent)
		}
	}
}

func TestFIX001P3PSOracleScalePrecisionCapacityClasses(t *testing.T) {
	cases := map[string]string{
		"B0-constant":      "ps_oracle_baby_capacity_failure",
		"G0-rescale":       "ps_oracle_giant_capacity_failure",
		"F0-final-rescale": "ps_oracle_final_rescale_capacity_failure",
	}
	for checkpoint, want := range cases {
		if got := psOracleScaleFailureClass(checkpoint, true); got != want {
			t.Fatalf("failure class for %s = %q, want %q", checkpoint, got, want)
		}
	}
	if got := psOracleScaleFailureClass("G0-add", false); got != "ps_oracle_primitive_mismatch" {
		t.Fatalf("primitive class = %q", got)
	}
}

func TestFIX001P3PSOracleScalePrecisionBudget(t *testing.T) {
	if psOracleScaleBudget != 1.2e-8 {
		t.Fatalf("budget = %g, want 1.2e-8", psOracleScaleBudget)
	}
	if psOracleScalePublicThreshold != 1e-2 {
		t.Fatalf("public threshold = %g, want 1e-2", psOracleScalePublicThreshold)
	}
}
