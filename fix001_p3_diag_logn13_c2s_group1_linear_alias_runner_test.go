package main

import "testing"

func TestFIX001P3C2SGroup1LinearClassification(t *testing.T) {
	checks := []struct {
		name   string
		kind   string
		unique bool
		match  bool
		want   string
	}{
		{"rotation", "baby_rotation", true, false, "logn13_c2s_group1_rotation_mismatch"},
		{"term-alias", "diagonal_product", false, true, "logn13_c2s_group1_single_term_q01_alias"},
		{"inner-alias", "inner_accumulation", false, true, "logn13_c2s_group1_inner_accumulation_q01_alias"},
		{"outer-alias", "outer_accumulation", false, true, "logn13_c2s_group1_outer_accumulation_q01_alias"},
	}
	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			got := c2sGroup1Classification(c2sAliasCheckpoint{Kind: tc.kind, FullRNSCenteredQ01Unique: tc.unique, FastResiduesMatch: tc.match})
			if got != tc.want {
				t.Fatalf("classification=%q, want %q", got, tc.want)
			}
		})
	}
}

func TestFIX001P3C2SGroup1WorstCapacityRecommendation(t *testing.T) {
	checkpoint := c2sAliasCheckpoint{Kind: "diagonal_product", FullRNSCenteredQ01Unique: false, FastResiduesMatch: true, StandardFullRNSCapacity: postMod1S2CCapacity{MaxAbsOverQ01Half: 5}}
	got := c2sGroup1Classification(checkpoint)
	if got != "logn13_c2s_group1_single_term_q01_alias" {
		t.Fatalf("alias classification=%q", got)
	}
}
