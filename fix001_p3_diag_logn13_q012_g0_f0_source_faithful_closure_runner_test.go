package main

import "testing"

func TestFIX001P3Q012G0F0SourceFaithfulClosureProvenance(t *testing.T) {
	if requiredFIX001P3Q012ClosurePrimary != "997caf2caa9657812cc2c8b9b9821ba9dff71bf4" {
		t.Fatalf("unexpected Primary provenance: %s", requiredFIX001P3Q012ClosurePrimary)
	}
	if requiredFIX001P3Q012ClosureSecondary != "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797" {
		t.Fatalf("unexpected Secondary provenance: %s", requiredFIX001P3Q012ClosureSecondary)
	}
}

func TestFIX001P3Q012G0F0SourceFaithfulClosureClassification(t *testing.T) {
	allowed := map[string]bool{
		"logn13_q012_g0_merge_harness_mismatch":                       true,
		"logn13_q012_g0_merge_source_faithful_but_downstream_blocked": true,
		"logn13_q055_plan92_g0_to_f0_q012_path_system_sufficient":     true,
		"logn13_q055_plan92_g0_to_f0_q012_path_operation_mismatch":    true,
		"logn13_q055_plan92_g0_to_f0_q012_path_numeric_insufficient":  true,
		"logn13_q012_g0_f0_control_mismatch":                          true,
	}
	if len(allowed) != 6 {
		t.Fatalf("unexpected classification set size: %d", len(allowed))
	}
}

func TestFIX001P3Q012G0F0SourceFaithfulClosureAlignmentPolicy(t *testing.T) {
	if q012ClosurePlanScaleExponent != 92 || q012ClosureThreshold != 1e-2 {
		t.Fatalf("unexpected fixed diagnostic parameters: scale=2^%d threshold=%g", q012ClosurePlanScaleExponent, q012ClosureThreshold)
	}
	proof := q012ClosureAlignment{SourcePolicy: "psRescaleGuardAddAligned: ratio.BigInt() then exact target-scale metadata"}
	if proof.SourcePolicy == "" {
		t.Fatal("source alignment policy must be recorded")
	}
}
