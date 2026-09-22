package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3T3Q012ABCleanQ01CRTAndRound(t *testing.T) {
	q0, q1, divisor := uint64(17), uint64(13), uint64(7)
	inverse, ok := fix001P3T3Q012ABInverseMod(q0%q1, q1)
	if !ok {
		t.Fatal("q01 inverse was not available")
	}
	for _, tc := range []struct {
		value        int64
		wantNegative bool
	}{{value: 100}, {value: -100, wantNegative: true}} {
		value := big.NewInt(tc.value)
		r0 := new(big.Int).Mod(new(big.Int).Set(value), new(big.Int).SetUint64(q0)).Uint64()
		r1 := new(big.Int).Mod(new(big.Int).Set(value), new(big.Int).SetUint64(q1)).Uint64()
		lo, hi := fix001P3T3Q012ABCRTQ01(r0, r1, q0, q1, inverse)
		q01 := new(big.Int).Mul(new(big.Int).SetUint64(q0), new(big.Int).SetUint64(q1))
		half := new(big.Int).Rsh(new(big.Int).Set(q01), 1)
		x := new(big.Int).SetUint64(hi)
		x.Lsh(x, 64)
		x.Add(x, new(big.Int).SetUint64(lo))
		negative := x.Cmp(half) > 0
		if negative {
			x.Sub(q01, x)
		}
		quotient := new(big.Int).Quo(new(big.Int).Set(x), new(big.Int).SetUint64(divisor))
		remainder := new(big.Int).Mod(new(big.Int).Set(x), new(big.Int).SetUint64(divisor))
		if remainder.Cmp(new(big.Int).SetUint64(divisor/2)) > 0 {
			quotient.Add(quotient, big.NewInt(1))
		}
		if negative != tc.wantNegative || quotient.Uint64() != 14 {
			t.Fatalf("value=%d got magnitude=%d negative=%v, want magnitude=14 negative=%v", tc.value, quotient.Uint64(), negative, tc.wantNegative)
		}
	}
}

func TestFIX001P3T3Q012ABScalingIdentityRows(t *testing.T) {
	if !fix001P3T3Q012ABNearExpected(1.0, 1.0, 1e-12) || fix001P3T3Q012ABNearExpected(1.2, 1.0, 0.1) {
		t.Fatal("recovery comparison helper is inconsistent")
	}
}
