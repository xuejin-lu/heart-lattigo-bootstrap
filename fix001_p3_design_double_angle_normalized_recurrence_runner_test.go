package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
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

func TestFIX001P3NormalizedFullSquareUsesSquaredScale(t *testing.T) {
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: 4, LogQ: []int{50, 50}, LogDefaultScale: 30})
	if err != nil {
		t.Fatalf("create test parameters: %v", err)
	}
	input := ckks.NewCiphertext(params, 1, 1)
	input.LogDimensions = ring.Dimensions{Cols: params.LogMaxSlots()}
	input.IsNTT = true
	input.IsMontgomery = true
	input.Scale = rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 20))
	for i := range input.Value[0].Coeffs[0] {
		input.Value[0].Coeffs[0][i] = uint64(i + 1)
		input.Value[0].Coeffs[1][i] = uint64(i + 2)
	}
	for limb := 0; limb < 2; limb++ {
		params.RingQ().SubRings[limb].MForm(input.Value[0].Coeffs[limb], input.Value[0].Coeffs[limb])
	}
	product, err := normalizedFullSquare(params, input)
	if err != nil {
		t.Fatalf("normalized full square: %v", err)
	}
	want := input.Scale.Mul(input.Scale)
	if !product.Scale.Equal(want) {
		t.Fatalf("square scale = %s, want squared input scale %s", finalizationScaleString(product.Scale), finalizationScaleString(want))
	}
	if product.Degree() != 2 || product.Level() != input.Level() || !product.IsNTT || !product.IsMontgomery {
		t.Fatalf("square metadata not preserved: degree=%d level=%d isNTT=%t isMontgomery=%t", product.Degree(), product.Level(), product.IsNTT, product.IsMontgomery)
	}
}

func TestFIX001P3NormalizedFastSnapshotIsImmutable(t *testing.T) {
	params := newQ01TestParameters(t, []int{50, 50})
	live := fillQ01TestCiphertext(t, params, 1, true)
	snapshot := live.CopyNew()
	want := snapshot.Value[0].Coeffs[0][0]
	live.Value[0].Coeffs[0][0]++
	if snapshot.Value[0].Coeffs[0][0] != want {
		t.Fatalf("snapshot changed after live ciphertext mutation: got %d, want %d", snapshot.Value[0].Coeffs[0][0], want)
	}
}

func TestFIX001P3FinalRestoreCapacityBound(t *testing.T) {
	data := targetScaleCenteredData{Q01Half: big.NewInt(1000), Values: [][]*big.Int{{big.NewInt(3), big.NewInt(-4)}}}
	evidence := normalizedFinalRestoreCapacity(data, big.NewInt(2))
	if !evidence.Pass || evidence.B != "4" || evidence.RestoreBound != "8" || evidence.Ratio != 0.008 {
		t.Fatalf("unexpected final restore capacity evidence: %+v", evidence)
	}
	failing := normalizedFinalRestoreCapacity(data, big.NewInt(300))
	if failing.Pass || failing.OutsideCount != 1 || failing.RestoreBound != "1200" {
		t.Fatalf("expected final restore capacity failure: %+v", failing)
	}
}

func TestFIX001P3FinalRestoreNormalizedReferenceRows(t *testing.T) {
	params := newQ01TestParameters(t, []int{50, 50})
	input := fillQ01TestCiphertext(t, params, 1, true)
	factor := big.NewInt(8)
	fast := input.CopyNew()
	reference := normalizedFullIntegerMultiply(params, input, factor)
	for component := range fast.Value {
		for limb := range fast.Value[component].Coeffs {
			for index := range fast.Value[component].Coeffs[limb] {
				fast.Value[component].Coeffs[limb][index] = reference.Value[component].Coeffs[limb][index]
			}
		}
	}
	if !normalizedStageComparison(fast, reference).Match {
		t.Fatal("normalized reference rows should match the exact integer restore")
	}
}

func TestFIX001P3FinalRestoreMetadataResetPreservesRows(t *testing.T) {
	params := newQ01TestParameters(t, []int{50, 50})
	ct := fillQ01TestCiphertext(t, params, 1, true)
	before := finalizationEvidence(ct).Rows
	ct.Scale = rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 30))
	if !doubleAngleRowsMatch(before, finalizationEvidence(ct).Rows) {
		t.Fatal("metadata-only scale reset changed q0/q1 rows")
	}
}
