package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func TestFIX001P3Q055Plan92T2LocalQ2GuardCapacityRegression(t *testing.T) {
	capacity := []q055T2CapacityEvidence{
		{State: "promoted_accumulator_after_multiply_by_2", IntendedQ01OutsideCount: 3, IntendedQ012OutsideCount: 0},
		{State: "intended_physical_post_add_before_divide_by_2", IntendedQ01OutsideCount: 1, IntendedQ012OutsideCount: 0},
	}
	if !q055T2CapacityHasTransientBlocker(capacity) {
		t.Fatal("expected the intended post-add q01 capacity blocker to be detected")
	}
	if q055T2CapacityHasTransientBlocker([]q055T2CapacityEvidence{{State: "intended_physical_post_add_before_divide_by_2", IntendedQ01OutsideCount: 1, IntendedQ012OutsideCount: 1}}) {
		t.Fatal("q012 capacity must clear the transient blocker")
	}
}

func TestFIX001P3Q055Plan92T2LocalQ2GuardFirstDivergenceRegression(t *testing.T) {
	downstream := &b4BabyDownstream{
		EvalModReal: &PSGlobalMetric{MaxComponent: 100},
		PostS2C:     &PSGlobalMetric{MaxComponent: 100},
		PublicLike:  &PSGlobalMetric{MaxComponent: 1000},
	}
	if got := q055T2FirstDivergence(downstream); got != "EvalMod.real" {
		t.Fatalf("first divergence = %q, want EvalMod.real", got)
	}
}

func TestFIX001P3Q055Plan92T2LocalQ2GuardDivideByTwoRows(t *testing.T) {
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: 4, LogQ: []int{50, 50, 50}, LogDefaultScale: 30})
	if err != nil {
		t.Fatal(err)
	}
	source := ckks.NewCiphertext(params, 1, 2)
	values := make([][]*big.Int, 2)
	for component := range values {
		values[component] = make([]*big.Int, params.N())
		for i := range values[component] {
			value := int64((component+1)*(i+3)*17 + 1)
			if (component+i)%3 == 0 {
				value = -value
			}
			values[component][i] = big.NewInt(value)
		}
	}
	source, err = nativeQ012BuildFromSigned(params, source, values, 2, rlwe.NewScale(1))
	if err != nil {
		t.Fatal(err)
	}
	sourceValues, err := nativeQ012SignedValues(params, source)
	if err != nil {
		t.Fatal(err)
	}
	if sourceValues[0][0].Cmp(values[0][0]) != 0 {
		t.Fatalf("source round-trip mismatch: got=%s want=%s", sourceValues[0][0], values[0][0])
	}
	got, err := q055T2DivideByTwoQ012(params, source)
	if err != nil {
		t.Fatal(err)
	}
	wantValues := q055T2RoundValues(values)
	want, err := nativeQ012BuildFromSigned(params, source, wantValues, 2, source.Scale.Div(rlwe.NewScale(2)))
	if err != nil {
		t.Fatal(err)
	}
	if equal, limbs := nativeQ012RowsEqual(params, got, want); !equal {
		t.Fatalf("q012 divide-by-two rows mismatch: limbs=%v", limbs)
	}
}
