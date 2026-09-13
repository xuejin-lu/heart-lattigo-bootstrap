package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3GuardedPowerPrecisionFactors(t *testing.T) {
	q := new(big.Int).SetUint64(1<<60 + 123)
	left, right, err := guardedPowerFactors(q, 4)
	if err != nil {
		t.Fatal(err)
	}
	if left.Sign() <= 0 || right.Sign() <= 0 {
		t.Fatalf("invalid factors: %s, %s", left, right)
	}
	want := new(big.Int).Lsh(new(big.Int).Set(q), 4)
	got := new(big.Int).Mul(left, right)
	if new(big.Int).Abs(new(big.Int).Sub(got, want)).Sign() < 0 {
		t.Fatalf("factor product error must be non-negative")
	}
	left2, right2, err := guardedPowerFactors(q, 4)
	if err != nil || left.Cmp(left2) != 0 || right.Cmp(right2) != 0 {
		t.Fatalf("factor selection is not deterministic: (%s,%s) vs (%s,%s)", left, right, left2, right2)
	}
}

func TestFIX001P3GuardedPowerPrecisionRoundDiv(t *testing.T) {
	for _, test := range []struct {
		value int64
		want  int64
	}{
		{value: 0, want: 0},
		{value: 7, want: 2},
		{value: 8, want: 2},
		{value: 9, want: 2},
		{value: -7, want: -2},
		{value: -8, want: -2},
		{value: -9, want: -2},
	} {
		got := guardedPowerRoundSigned(big.NewInt(test.value), 2)
		if got.Cmp(big.NewInt(test.want)) != 0 {
			t.Fatalf("round(%d/4) = %s, want %d", test.value, got, test.want)
		}
	}
	if got := guardedPowerRoundSigned(big.NewInt(3), 0); got.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("g=0 is not a no-op: %s", got)
	}
}

func TestFIX001P3GuardedPowerPrecisionCandidateRange(t *testing.T) {
	if len(guardedPowerExponents) != 13 || guardedPowerExponents[0] != 0 || guardedPowerExponents[len(guardedPowerExponents)-1] != 12 {
		t.Fatalf("guard sweep = %v, want 0..12", guardedPowerExponents)
	}
}
