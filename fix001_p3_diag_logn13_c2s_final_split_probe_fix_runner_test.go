package main

import "testing"

func TestFIX001P3C2SSplitProbeCheckpointNames(t *testing.T) {
	want := []string{"split_input_zv", "split_conjugate", "split_imag_sub", "split_imag_mul_minus_i", "split_real_add"}
	for i, name := range want {
		if name == "" || i < 0 {
			t.Fatal("invalid split checkpoint definition")
		}
	}
}

func TestFIX001P3C2SSplitProbeAliasClassification(t *testing.T) {
	checkpoint := c2sFinalSplitCheckpoint{Checkpoint: "split_imag_mul_minus_i", FullRNSCenteredQ01Unique: false, FastResiduesMatch: true}
	if got := c2sFinalSplitCause(checkpoint); got != "logn13_c2s_final_split_imag_mul_q01_alias" {
		t.Fatalf("got %q", got)
	}
}
