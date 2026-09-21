package main

import "testing"

func TestFIX001P3P93InternalFirstDivergenceUsesInternalBudget(t *testing.T) {
	branches := map[string]fix001P3P93InternalBranch{
		"real": {Checkpoints: []fix001P3P93InternalCheckpoint{
			{Name: "polynomial_output_after_rescale", Metric: &semanticBisectMetric{MaxComponentAbs: 5e-7}},
			{Name: "pre_double_angle", Operation: "normalized coherent scale", Metric: &semanticBisectMetric{MaxComponentAbs: 0.7}},
			{Name: "internal_final_restore", Metric: &semanticBisectMetric{MaxComponentAbs: 0.004}},
		}},
		"imag": {Checkpoints: []fix001P3P93InternalCheckpoint{}},
	}
	_, material := firstFIX001P3P93InternalDivergence(branches, 1e-12, fix001P3P93InternalBudget)
	if material["checkpoint"] != "pre_double_angle" {
		t.Fatalf("first material checkpoint = %v, want pre_double_angle", material["checkpoint"])
	}
	if material["threshold"] != fix001P3P93InternalBudget {
		t.Fatalf("material threshold = %v, want %v", material["threshold"], fix001P3P93InternalBudget)
	}
}
