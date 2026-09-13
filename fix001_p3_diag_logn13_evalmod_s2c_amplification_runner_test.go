package main

import "testing"

func TestFIX001P3EvalModS2CAmplificationClassification(t *testing.T) {
	classification, checkpoint := fix001P3EvalModS2CClassification(nil, []postMod1S2CGroupCheckpoint{{Group: 0}}, 0.246, 0.246, 1e-10)
	if classification != "logn13_evalmod_semantic_difference_amplified_by_s2c" || checkpoint != "h_f_vs_h_s" {
		t.Fatalf("unexpected classification: %s at %s", classification, checkpoint)
	}
}

func TestFIX001P3EvalModS2CAmplificationClassificationGroups(t *testing.T) {
	for group, expected := range map[int]string{0: "logn13_s2c_group0_mismatch", 1: "logn13_s2c_group1_mismatch", 2: "logn13_s2c_group2_mismatch"} {
		classification, _ := fix001P3EvalModS2CClassification(&postMod1S2CFailure{Cause: "post_mod1_s2c_linear_transform_mismatch", Group: group, Checkpoint: "pre_rescale_rows"}, nil, 0, 0, 0)
		if classification != expected {
			t.Fatalf("group %d: got %s want %s", group, classification, expected)
		}
	}
}

func TestFIX001P3EvalModS2CAmplificationThreshold(t *testing.T) {
	metric := semanticBisectMetric{MaxComponentAbs: 9e-3}
	if fix001P3EvalModS2CMax(&metric) != 9e-3 {
		t.Fatalf("unexpected metric max")
	}
}
