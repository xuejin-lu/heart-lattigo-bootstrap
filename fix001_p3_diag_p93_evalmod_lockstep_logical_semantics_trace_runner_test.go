package main

import (
	"math"
	"testing"
)

func TestFIX001P3P93CanonicalFastLogicalExponent(t *testing.T) {
	values := []complex128{1.25 - 0.5i, -0.25 + 0.75i}
	for _, exponent := range []int{0, 27, 54} {
		factor := math.Ldexp(1, exponent)
		for _, value := range values {
			want := complex(real(value)*factor, imag(value)*factor)
			got := fix001P3P93CanonicalValue(value, exponent)
			if got != want {
				t.Fatalf("canonical exponent %d: got %v want %v", exponent, got, want)
			}
		}
	}
}

func TestFIX001P3P93LogicalPairsCarryFastExponentEquations(t *testing.T) {
	pairs := fix001P3P93LogicalPairs()
	want := map[string]int{
		"polynomial_output":                   0,
		"pre_double_angle_logical":            27,
		"double_angle_round_0_square":         54,
		"double_angle_round_0_after_constant": 27,
		"double_angle_round_2_post_rescale":   27,
		"internal_final_restore":              0,
	}
	for _, pair := range pairs {
		if expected, ok := want[pair.Name]; ok && pair.Exponent != expected {
			t.Fatalf("%s exponent=%d want %d", pair.Name, pair.Exponent, expected)
		}
	}
}

func TestFIX001P3P93LogicalClassification(t *testing.T) {
	cases := map[string]string{
		"polynomial_output":           "C",
		"double_angle_round_0_square": "D",
		"double_angle_round_1_square": "E",
		"double_angle_round_2_square": "F",
		"internal_final_restore":      "G",
		"none":                        "H",
	}
	for checkpoint, want := range cases {
		got := fix001P3P93LogicalClassification(map[string]interface{}{"checkpoint": checkpoint})
		if got != want {
			t.Fatalf("checkpoint %s classification=%s want %s", checkpoint, got, want)
		}
	}
}
