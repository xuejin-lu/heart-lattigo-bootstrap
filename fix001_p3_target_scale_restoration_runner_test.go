package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestFIX001P3TargetScaleNearestIntegerEvidence(t *testing.T) {
	evidence, multiplier, matched := targetScaleScaleEvidence(rlwe.NewScale(8), rlwe.NewScale(25))
	if multiplier.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("nearest target-scale integer = %s, want 3", multiplier)
	}
	if evidence.M != "3" || !evidence.NearestIntegerPositive {
		t.Fatalf("unexpected scale evidence: %+v", evidence)
	}
	if matched.Cmp(rlwe.NewScale(24)) != 0 {
		t.Fatalf("matched scale = %s, want 24", matched.Value.Text('f', -1))
	}
}

func TestFIX001P3TargetScaleCapacityRejectsProspectiveWrap(t *testing.T) {
	data := targetScaleCenteredData{
		Q0:      big.NewInt(5),
		Q1:      big.NewInt(7),
		Q01:     big.NewInt(35),
		Q01Half: big.NewInt(17),
		Values:  [][]*big.Int{{big.NewInt(10)}, {big.NewInt(0)}},
	}
	if evidence := targetScaleCapacity(data, big.NewInt(2)); evidence.Pass || evidence.OutsideCount != 1 {
		t.Fatalf("capacity evidence = %+v, want one prospective outside coefficient", evidence)
	}
}
