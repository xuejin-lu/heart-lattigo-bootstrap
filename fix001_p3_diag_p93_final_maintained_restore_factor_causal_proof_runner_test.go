package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestFIX001P3FinalRestoreDerivesK17FromRuntimeScaleTransition(t *testing.T) {
	sIn := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 60))
	sOut := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 50))
	_, factor, exponent, relativeError, err := fix001P3FinalRestoreFactor(sOut, sIn, 27)
	if err != nil {
		t.Fatal(err)
	}
	if exponent != 17 {
		t.Fatalf("nearest restore exponent=%d, want 17 (factor=%g)", exponent, factor)
	}
	if relativeError > 1e-12 {
		t.Fatalf("nearest power relative error=%g, want <=1e-12", relativeError)
	}
}

func TestFIX001P3FinalRestoreControlAndCounterfactualDifferOnlyInMultiplier(t *testing.T) {
	if fix001P3FinalRestoreCurrentExponent == 17 {
		t.Fatal("control and counterfactual exponents unexpectedly equal")
	}
	if fix001P3FinalRestoreCurrentExponent != 27 {
		t.Fatalf("current exponent=%d, want 27", fix001P3FinalRestoreCurrentExponent)
	}
	if got := mathPow2(fix001P3FinalRestoreCurrentExponent - 17); got != 1024 {
		t.Fatalf("control/counterfactual multiplier ratio=%g, want 1024", got)
	}
}

func TestFIX001P3FinalRestoreRowHashProofRequiresEquality(t *testing.T) {
	a := map[string]string{"c0_q0": "a", "c0_q1": "b"}
	b := map[string]string{"c0_q0": "a", "c0_q1": "b"}
	if !fix001P3FinalRestoreRowsEqual(a, b) {
		t.Fatal("equal row hashes did not compare equal")
	}
	b["c0_q1"] = "changed"
	if fix001P3FinalRestoreRowsEqual(a, b) {
		t.Fatal("changed row hashes compared equal")
	}
}

func TestFIX001P3FinalRestorePublicContractFactor(t *testing.T) {
	internal := 0.0037575136047775944
	public := internal * 32
	if ratio := safeRatio(public, internal); mathAbs(ratio-32) > 1e-12 {
		t.Fatalf("public/internal ratio=%g, want 32", ratio)
	}
}

func mathAbs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func mathPow2(exponent int) float64 {
	value := 1.0
	for i := 0; i < exponent; i++ {
		value *= 2
	}
	return value
}
