package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestFIX001P3RuntimeScaleRoundingRecordsNonIntegralRatio(t *testing.T) {
	rounded, absolute, relative, exact := fix001P3ScaleRounding(rlwe.NewScale(new(big.Float).SetFloat64(3.25)))
	if rounded != "3" {
		t.Fatalf("rounded=%s, want 3", rounded)
	}
	if absolute != 0.25 {
		t.Fatalf("absolute mismatch=%g, want 0.25", absolute)
	}
	if relative != 1.0/13.0 {
		t.Fatalf("relative mismatch=%g, want %g", relative, 1.0/13.0)
	}
	if exact {
		t.Fatal("non-integral ratio reported as exact")
	}
}

func TestFIX001P3RuntimeScaleRoundingRecordsExactRatio(t *testing.T) {
	rounded, absolute, relative, exact := fix001P3ScaleRounding(rlwe.NewScale(16))
	if rounded != "16" || absolute != 0 || relative != 0 || !exact {
		t.Fatalf("exact ratio evidence=(%s, %g, %g, %t), want (16, 0, 0, true)", rounded, absolute, relative, exact)
	}
}
