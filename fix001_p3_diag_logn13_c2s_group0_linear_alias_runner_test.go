package main

import "testing"

func TestFIX001P3C2SGroup0LinearAliasClassification(t *testing.T) {
	unsafeButMatching := c2sAliasCheckpoint{
		FullRNSCenteredQ01Unique: false,
		FastResiduesMatch:        true,
	}
	if got := c2sAliasCheckpointFailure(unsafeButMatching, "mismatch", "alias"); got != "alias" {
		t.Fatalf("unsafe matching checkpoint = %q, want alias", got)
	}
	unsafeMismatch := c2sAliasCheckpoint{
		FullRNSCenteredQ01Unique: false,
		FastResiduesMatch:        false,
	}
	if got := c2sAliasCheckpointFailure(unsafeMismatch, "mismatch", "alias"); got != "mismatch" {
		t.Fatalf("unsafe mismatching checkpoint = %q, want mismatch", got)
	}
	safe := c2sAliasCheckpoint{
		FullRNSCenteredQ01Unique: true,
		FastResiduesMatch:        true,
	}
	if got := c2sAliasCheckpointFailure(safe, "mismatch", "alias"); got != "" {
		t.Fatalf("safe checkpoint = %q, want no failure", got)
	}
}

func TestFIX001P3C2SGroup0LinearAliasCompressionBound(t *testing.T) {
	checkpoint := c2sAliasCheckpoint{
		Checkpoint: "giant_0_diagonal_0",
		StandardFullRNSCapacity: postMod1S2CCapacity{
			MaxAbsExactCoefficient: "24095776414275219558662852736",
			MaxAbsOverQ01Half:      2.4330520098231347,
		},
	}
	got := c2sAliasCompressionFor(checkpoint, "term_multiplication")
	if got.KMin != 2 || got.KSafe != 3 {
		t.Fatalf("compression bound = k_min %d, k_safe %d, want 2, 3", got.KMin, got.KSafe)
	}
	if got.WorstCheckpoint != checkpoint.Checkpoint || got.RequiredBeforeOperation != "term_multiplication" {
		t.Fatalf("compression provenance = %#v", got)
	}
}

