package main

import (
	"math/big"
	"testing"

	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestFIX001P3PSMixedScalePrecisionSchedule(t *testing.T) {
	base := commonpolynomial.PatersonStockmeyerPolynomial{Value: make([]commonpolynomial.Polynomial, 4)}
	schedule, err := psMixedScalePlan(base, []int{92, 93, 92, 94})
	if err != nil {
		t.Fatal(err)
	}
	for i, exponent := range []int{92, 93, 92, 94} {
		want := precisionSweepScale(exponent)
		if !schedule.Value[i].Scale.Equal(want) {
			t.Fatalf("block %d scale = %s, want %s", i, finalizationScaleString(schedule.Value[i].Scale), finalizationScaleString(want))
		}
	}
}

func TestFIX001P3PSMixedScalePrecisionExactAlignment(t *testing.T) {
	left := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 92))
	right := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 93))
	_, integer, fractional, exact := psMixedScaleRatio(right, left)
	frac, ok := new(big.Float).SetString(fractional)
	if !exact || integer != "2" || !ok || frac.Sign() != 0 {
		t.Fatalf("ratio = integer %s fractional %s exact %v", integer, fractional, exact)
	}
}

func TestFIX001P3PSMixedScalePrecisionClassification(t *testing.T) {
	if psMixedScaleBudget != 1.2e-8 {
		t.Fatalf("budget = %g, want 1.2e-8", psMixedScaleBudget)
	}
	if psMixedScalePublicThreshold != 1e-2 {
		t.Fatalf("public threshold = %g, want 1e-2", psMixedScalePublicThreshold)
	}
}
