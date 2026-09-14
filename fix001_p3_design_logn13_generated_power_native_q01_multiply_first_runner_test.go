package main

import "testing"

func TestFIX001P3GeneratedPowerNativeQ01MultiplyFirstClassification(t *testing.T) {
	allowed := map[string]bool{
		"logn13_generated_power_native_q01_multiply_first_system_sufficient": true,
		"logn13_generated_power_native_q01_capacity_regression":              true,
		"logn13_generated_power_native_q01_operation_mismatch":               true,
		"logn13_generated_power_native_q01_rescale_mismatch":                 true,
		"logn13_generated_power_native_q01_numeric_insufficient":             true,
		"logn13_generated_power_native_q01_precondition_mismatch":            true,
	}
	if len(allowed) != 6 {
		t.Fatal("native q01 classification set changed unexpectedly")
	}
}
