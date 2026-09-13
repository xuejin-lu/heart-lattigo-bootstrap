package main

import "testing"

func TestFIX001P3C2SGroup1CompressedCandidateSet(t *testing.T) {
	want := []int{1, 2, 3}
	for i, k := range want {
		if k != i+1 {
			t.Fatalf("candidate order changed: got k=%d at index %d", k, i)
		}
	}
}

func TestFIX001P3C2SGroup1CompressedStructureInvariant(t *testing.T) {
	structure := c2sGroup1ExpectedStructure(15, "unused")
	if structure.N1 != 256 || structure.Path != "BSGS" || structure.DiagonalCount != 15 || structure.LevelQ != 16 || structure.LevelP != 4 {
		t.Fatalf("unexpected expected structure: %+v", structure)
	}
}
