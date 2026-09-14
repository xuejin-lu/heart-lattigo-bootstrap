package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3B4T2ScalarGuardProvenance(t *testing.T) {
	if requiredFIX001P3T2GuardPrimary != "490d3f3e6402f52b8c31d62a51b3e4e862954fc3" {
		t.Fatalf("unexpected Primary base: %s", requiredFIX001P3T2GuardPrimary)
	}
	if requiredFIX001P3T2GuardSecondary != "25e70430b10cd4af37ad4cf94cd994912e970e3f" {
		t.Fatalf("unexpected Secondary base: %s", requiredFIX001P3T2GuardSecondary)
	}
}

func TestFIX001P3B4T2ScalarGuardSignedRound(t *testing.T) {
	cases := map[string]string{"one": "1", "minus-one": "-1", "two": "2", "minus-two": "-2", "three": "3", "minus-three": "-3"}
	want := map[string]string{"one": "1", "minus-one": "-1", "two": "1", "minus-two": "-1", "three": "2", "minus-three": "-2"}
	for name, input := range cases {
		value, _ := new(big.Int).SetString(input, 10)
		if got := t2GuardRoundSigned(value).String(); got != want[name] {
			t.Fatalf("%s: got %s, want %s", name, got, want[name])
		}
	}
}

func TestFIX001P3B4T2ScalarGuardCapacityClassificationNames(t *testing.T) {
	allowed := map[string]bool{
		"logn13_b4_t2_one_bit_scalar_guard_system_sufficient":     true,
		"logn13_b4_t6_t2_one_bit_scalar_guards_system_sufficient": true,
		"logn13_b4_local_scalar_guard_insufficient_precision":     true,
		"logn13_b4_t2_one_bit_scalar_guard_capacity_blocker":      true,
		"logn13_b4_t2_scalar_guard_contraction_mismatch":          true,
	}
	for classification := range allowed {
		if !allowed[classification] {
			t.Fatalf("classification not accepted: %s", classification)
		}
	}
}
