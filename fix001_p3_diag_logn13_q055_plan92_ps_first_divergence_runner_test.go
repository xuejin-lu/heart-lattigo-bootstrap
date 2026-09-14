package main

import "testing"

func TestFIX001P3Q055Plan92PSFirstDivergenceCenteredSensitive(t *testing.T) {
	for _, operation := range []string{"baby_mul_then_add", "giant_rescale", "final_rescale", "centered_contraction"} {
		if !psFirstCenteredSensitive(operation) {
			t.Fatalf("operation %q was not marked centered-sensitive", operation)
		}
	}
	if psFirstCenteredSensitive("giant_multiply") {
		t.Fatal("pure modular giant multiply must not be marked centered-sensitive")
	}
}

func TestFIX001P3Q055Plan92PSFirstDivergenceWindowCause(t *testing.T) {
	valid := designCapacity{CenteredUnique: true}
	invalid := designCapacity{CenteredUnique: false}
	b := &psFirstCase{Real: psFirstTrace{Operations: []psFirstOperation{{Order: 0, ID: "B4-term-2", Intended: psFirstCapacity{IntendedQ01: &invalid, IntendedQ012: &valid}}, {Order: 1, ID: "G0-rescale", CenteredSensitive: true, Intended: psFirstCapacity{IntendedQ01: &valid}}}}, Imag: psFirstTrace{Operations: []psFirstOperation{}}}
	d := &psFirstCase{Real: psFirstTrace{Operations: []psFirstOperation{{Order: 0, ID: "B4-term-2", Intended: psFirstCapacity{IntendedQ01: &valid, IntendedQ012: &valid}}, {Order: 1, ID: "G0-rescale", CenteredSensitive: true, Intended: psFirstCapacity{IntendedQ01: &valid}}}}, Imag: psFirstTrace{Operations: []psFirstOperation{}}}
	window := psFirstDetermineWindow(b, d)
	if window.FirstOperation != "B4-term-2" || window.FirstCenteredBoundary != "G0-rescale" || window.Cause != "q01_capacity_before_centered_sensitive_boundary" || !window.Q012Sufficient {
		t.Fatalf("unexpected first window: %+v", window)
	}
}
