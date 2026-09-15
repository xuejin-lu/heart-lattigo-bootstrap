package main

import "testing"

func TestFIX001P3PSWideQ012Q056ScaleSweepSchedule(t *testing.T) {
	if len(psWideScaleExponents) != 3 || psWideScaleExponents[0] != 92 || psWideScaleExponents[1] != 93 || psWideScaleExponents[2] != 94 {
		t.Fatalf("unexpected PS-wide scale sweep: %v", psWideScaleExponents)
	}
	for _, exponent := range psWideScaleExponents {
		if psWideScale(exponent).BigInt().BitLen() != exponent+1 {
			t.Fatalf("2^%d scale has unexpected bit length", exponent)
		}
	}
}

func TestFIX001P3PSWideQ012Q056Classification(t *testing.T) {
	candidates := []psWideCandidate{
		{Exponent: 92, PrecisionPass: true, CapacityPass: true, PSExitContractionPass: true, EvalModPass: true, PostS2CPass: true, PublicLikePass: true},
		{Exponent: 93, CapacityPass: false},
	}
	classification, recommended, readiness := psWideClassify(candidates)
	if classification != "logn13_ps_wide_q012_q056_p92_system_sufficient" || recommended != "2^92" || readiness != "production_ready_for_fixed_width_ps_wide_q012_design" {
		t.Fatalf("unexpected classification: %s %s %s", classification, recommended, readiness)
	}
}

func TestFIX001P3PSWideQ012Q056RequiresPSExitContraction(t *testing.T) {
	candidates := []psWideCandidate{{Exponent: 92, CapacityPass: true, PSExitContractionPass: false, EvalModPass: true, PostS2CPass: true, PublicLikePass: true}}
	classification, _, _ := psWideClassify(candidates)
	if classification != "logn13_ps_wide_q012_output_not_q01_contractible" {
		t.Fatalf("expected output contraction blocker, got %s", classification)
	}
}
