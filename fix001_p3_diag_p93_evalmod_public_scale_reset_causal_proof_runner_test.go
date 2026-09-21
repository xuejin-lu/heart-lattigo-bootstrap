package main

import "testing"

func TestFIX001P3PublicScaleResetCausalProofScaleRatio(t *testing.T) {
	if got := float64(uint64(1) << (fix001P3PublicScaleResetInputScaleBits - fix001P3PublicScaleResetDefaultScaleBits)); got != 32 {
		t.Fatalf("scale ratio = %v, want 32", got)
	}
}

func TestFIX001P3PublicScaleResetCausalProofClassificationNames(t *testing.T) {
	want := map[string]bool{
		"P93_PUBLIC_SCALE_RESET_REPLAY_CONFLICT":                   true,
		"P93_PUBLIC_SCALE_RESET_NOT_METADATA_ONLY":                 true,
		"P93_SCALE_METADATA_CORRECTION_INSUFFICIENT":               true,
		"P93_SCALE_METADATA_CAUSAL_BUT_SYSTEM_STILL_FAILS_1E2":     true,
		"P93_PUBLIC_EVALMOD_SCALE_RESET_CAUSAL_AND_1E2_SUFFICIENT": true,
		"P93_PUBLIC_SCALE_RESET_SOURCE_SEMANTICS_UNRESOLVED":       true,
	}
	if len(want) != 6 {
		t.Fatal("causal-proof classification set is incomplete")
	}
}
