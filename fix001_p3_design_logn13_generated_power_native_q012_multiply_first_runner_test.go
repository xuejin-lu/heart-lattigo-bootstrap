package main

import "testing"

func TestFIX001P3GeneratedPowerNativeQ012MultiplyFirstClassification(t *testing.T) {
	allowed := map[string]bool{
		"logn13_generated_power_native_q012_multiply_first_system_sufficient": true,
		"logn13_generated_power_native_q012_capacity_blocker":                 true,
		"logn13_generated_power_native_q012_contraction_blocker":              true,
		"logn13_generated_power_native_q012_operation_mismatch":               true,
		"logn13_generated_power_native_q012_rescale_mismatch":                 true,
		"logn13_generated_power_native_q012_numeric_insufficient":             true,
		"logn13_generated_power_q012_precondition_mismatch":                   true,
	}
	if len(allowed) != 7 {
		t.Fatal("native q012 classification set changed unexpectedly")
	}
}
