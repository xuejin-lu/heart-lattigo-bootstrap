package main

import (
	"math"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestFIX001P3DoubleAngleHighPrecisionSquare(t *testing.T) {
	value := doubleAngleHPFromComplex(complex(1.25, -0.5))
	squared := doubleAngleHPSquare(value)
	realValue, _ := squared.Real.Float64()
	imagValue, _ := squared.Imag.Float64()
	if math.Abs(realValue-1.3125) > 1e-15 || math.Abs(imagValue+1.25) > 1e-15 {
		t.Fatalf("high-precision square = (%g,%g), want (1.3125,-1.25)", realValue, imagValue)
	}
}

func TestFIX001P3DoubleAngleScaleDeltaExactIsSerializable(t *testing.T) {
	scale := rlwe.NewScale(1 << 40)
	if got := doubleAngleScaleDelta(scale, scale); math.IsInf(got, 0) || math.IsNaN(got) || got != math.MaxFloat64 {
		t.Fatalf("exact scale delta = %v, want finite exact-match sentinel", got)
	}
}
