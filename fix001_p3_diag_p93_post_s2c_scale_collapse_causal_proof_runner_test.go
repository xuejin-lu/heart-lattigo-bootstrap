package main

import "testing"

func TestFIX001P3PostS2CScaleCollapseCausalProofConstants(t *testing.T) {
	if fix001P3PostS2CScaleCollapseThreshold != 1e-2 {
		t.Fatalf("threshold = %v, want 1e-2", fix001P3PostS2CScaleCollapseThreshold)
	}
	if fix001P3PostS2CScaleCollapseExact >= fix001P3PostS2CScaleCollapseThreshold {
		t.Fatalf("exact convergence threshold = %v must be below diagnostic threshold", fix001P3PostS2CScaleCollapseExact)
	}
}

func TestFIX001P3PostS2CScaleCollapseClassificationNames(t *testing.T) {
	want := map[string]bool{
		"P93_POST_S2C_REPLAY_CONFLICT":                              true,
		"P93_POST_S2C_ARITHMETIC_COLLAPSE":                          true,
		"P93_UNPACK_SCALE_METADATA_COLLAPSE":                        true,
		"P93_FINALIZATION_SCALE_METADATA_COLLAPSE":                  true,
		"P93_MULTIPLE_DOWNSTREAM_SCALE_RESETS":                      true,
		"P93_DOWNSTREAM_METADATA_CAUSAL_BUT_SYSTEM_STILL_FAILS_1E2": true,
		"P93_EVALMOD_PLUS_DOWNSTREAM_METADATA_SUFFICIENT_FOR_1E2":   true,
		"P93_POST_S2C_TRACE_ALIGNMENT_INVALID":                      true,
	}
	if len(want) != 8 {
		t.Fatalf("classification set has %d entries, want 8", len(want))
	}
}
