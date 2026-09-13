package main

import "testing"

func TestFIX001P3C2SGroup0CompressedCheckpointCapacity(t *testing.T) {
	safe := c2sAliasCheckpoint{FullRNSCenteredQ01Unique: true, FastResiduesMatch: true}
	if c2sCompressedCheckpointFailure(safe) {
		t.Fatal("safe checkpoint was classified as a failure")
	}
	unsafe := c2sAliasCheckpoint{FullRNSCenteredQ01Unique: false, FastResiduesMatch: true}
	if !c2sCompressedCheckpointFailure(unsafe) {
		t.Fatal("unsafe checkpoint was not classified as a failure")
	}
	mismatch := c2sAliasCheckpoint{FullRNSCenteredQ01Unique: true, FastResiduesMatch: false}
	if !c2sCompressedCheckpointFailure(mismatch) {
		t.Fatal("residue mismatch was not classified as a failure")
	}
}

func TestFIX001P3C2SGroup0CompressedStructureInvariant(t *testing.T) {
	original := c2sAliasStructure{
		N1: 1024, Path: "BSGS", DiagonalCount: 8,
		SortedDiagonalIndices: []int{0, 512, 1024, 1536},
		GiantGroups:           []c2sAliasGiantGroup{{GiantRotation: 0, BabyRotations: []int{0, 512}}},
		LevelQ:                16, LevelP: 4, LogDimensions: "rows=0,cols=12",
	}
	copyOfOriginal := original
	if !c2sCompressedStructureEqual(original, copyOfOriginal) {
		t.Fatal("identical structure was not accepted")
	}
	copyOfOriginal.N1 = 512
	if c2sCompressedStructureEqual(original, copyOfOriginal) {
		t.Fatal("changed BSGS N1 was accepted")
	}
}

func TestFIX001P3C2SGroup0CompressedMetadataRestoreContract(t *testing.T) {
	want := semanticBisectMetadata{SourceLevel: 15, Degree: 1, N: 8192, IsNTT: true, IsMontgomery: true, Scale: "scale"}
	if !c2sCompressedMetadataMatch(want, want) {
		t.Fatal("identical metadata did not match")
	}
	changed := want
	changed.SourceLevel = 14
	if c2sCompressedMetadataMatch(want, changed) {
		t.Fatal("changed level matched restore contract")
	}
}
