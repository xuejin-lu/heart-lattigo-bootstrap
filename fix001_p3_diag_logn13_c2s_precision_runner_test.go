package main

import "testing"

func TestFIX001P3C2SPrecisionClassification(t *testing.T) {
	tests := []struct {
		group int
		point string
		want  string
	}{
		{0, "linear_transform", "logn13_c2s_precision_first_breach_group0_linear_transform"},
		{0, "rescale", "logn13_c2s_precision_first_breach_group0_rescale"},
		{3, "linear_transform", "logn13_c2s_precision_first_breach_group3_linear_transform"},
		{3, "rescale", "logn13_c2s_precision_first_breach_group3_rescale"},
	}
	for _, test := range tests {
		if got := c2sPrecisionClassification(test.group, test.point); got != test.want {
			t.Fatalf("classification(%d, %q) = %q, want %q", test.group, test.point, got, test.want)
		}
	}
}
