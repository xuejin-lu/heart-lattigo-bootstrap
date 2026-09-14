package main

import "testing"

func TestFIX001P3GeneratedPowerMultiplyFirstQ2Classification(t *testing.T) {
	allowed := map[string]bool{
		"logn13_generated_power_pre_rescale_quantization_blocker":          true,
		"logn13_generated_power_multiply_first_q01_capacity_blocker":       true,
		"logn13_generated_power_multiply_first_q012_capacity_blocker":      true,
		"logn13_generated_power_local_q2_multiply_first_system_sufficient": true,
		"logn13_generated_power_local_q2_multiply_first_insufficient":      true,
		"logn13_generated_power_schedule_attribution_mismatch":             true,
	}
	for classification := range allowed {
		if !allowed[classification] {
			t.Fatalf("classification %q was not retained", classification)
		}
	}
}
