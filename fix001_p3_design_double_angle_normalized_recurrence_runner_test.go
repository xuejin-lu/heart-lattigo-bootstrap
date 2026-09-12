package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestFIX001P3NormalizedPowerOfTwoExponent(t *testing.T) {
	working := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 31))
	scale := working.Mul(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 29)))
	exponent, err := normalizedPowerOfTwoExponent(scale, working)
	if err != nil {
		t.Fatalf("normalizedPowerOfTwoExponent returned error: %v", err)
	}
	if exponent != 29 {
		t.Fatalf("unexpected exponent: got %d, want 29", exponent)
	}
}

func TestFIX001P3NormalizedSquareBound(t *testing.T) {
	data := targetScaleCenteredData{
		Q01Half: big.NewInt(1000),
		Values:  [][]*big.Int{{big.NewInt(3), big.NewInt(-4)}},
	}
	bound := normalizedSquareBound(data, 10)
	if !bound.Pass {
		t.Fatalf("expected square bound to pass: %+v", bound)
	}
	if bound.NSquaredB != "160" {
		t.Fatalf("unexpected N*B^2: got %s, want 160", bound.NSquaredB)
	}
}

func TestFIX001P3NormalizedFullCRTCentered(t *testing.T) {
	value := normalizedFullCRTCentered([]uint64{14, 16}, []uint64{17, 19})
	if value.Cmp(big.NewInt(-3)) != 0 {
		t.Fatalf("unexpected centered CRT value: got %s, want -3", value)
	}
}

func TestFIX001P3NormalizedLog2DeviationIsAbsolute(t *testing.T) {
	actual := rlwe.NewScale(big.NewInt(2))
	target := rlwe.NewScale(big.NewInt(4))
	if got := normalizedLog2Deviation(actual, target); got != 1 {
		t.Fatalf("unexpected absolute log2 deviation: got %v, want 1", got)
	}
}
