package main

import (
	"math"
	"testing"

	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

func TestFIX001ExpectedPowerT0IsOne(t *testing.T) {
	inputs := []complex128{complex(-2.5, 1.25), complex(0.25, -0.5), complex(3, 0)}
	got := fix001ExpectedPower(0, inputs, make(map[int][]complex128))
	for i, value := range got {
		if value != 1 {
			t.Fatalf("T0[%d] = %v, want 1+0i", i, value)
		}
	}
}

func TestFIX001ExpectedPowerT1IsInput(t *testing.T) {
	inputs := []complex128{complex(-2.5, 1.25), complex(0.25, -0.5), complex(3, 0)}
	got := fix001ExpectedPower(1, inputs, make(map[int][]complex128))
	for i := range inputs {
		if got[i] != inputs[i] {
			t.Fatalf("T1[%d] = %v, want %v", i, got[i], inputs[i])
		}
	}
}

func TestFIX001ExpectedPowerT2MatchesClosedForm(t *testing.T) {
	inputs := []complex128{complex(-0.75, 0), complex(0.25, 0), complex(0.5, 0.125), complex(-0.2, -0.3)}
	got := fix001ExpectedPower(2, inputs, make(map[int][]complex128))
	for i, x := range inputs {
		want := 2*x*x - 1
		if math.Hypot(real(got[i]-want), imag(got[i]-want)) > 1e-14 {
			t.Fatalf("T2[%d] = %v, want %v", i, got[i], want)
		}
	}
}

func TestFIX001ExpectedPowerT3MatchesClosedForm(t *testing.T) {
	inputs := []complex128{complex(-0.75, 0), complex(0.25, 0), complex(0.5, 0.125), complex(-0.2, -0.3)}
	got := fix001ExpectedPower(3, inputs, make(map[int][]complex128))
	for i, x := range inputs {
		want := 4*x*x*x - 3*x
		if math.Hypot(real(got[i]-want), imag(got[i]-want)) > 1e-14 {
			t.Fatalf("T3[%d] = %v, want %v", i, got[i], want)
		}
	}
}

func TestFIX001ExpectedPowerMatchesBignumChebyshev(t *testing.T) {
	inputs := []complex128{complex(-0.75, 0), complex(0.25, 0), complex(0.5, 0.125), complex(-0.2, -0.3)}
	memo := make(map[int][]complex128)
	for n := 0; n <= 16; n++ {
		got := fix001ExpectedPower(n, inputs, memo)
		coeffs := make([]float64, n+1)
		coeffs[n] = 1
		poly := bignum.NewPolynomial(bignum.Chebyshev, coeffs, [2]float64{-1, 1})
		for i, x := range inputs {
			value := poly.Evaluate(x)
			realValue, _ := value[0].Float64()
			imagValue, _ := value[1].Float64()
			want := complex(realValue, imagValue)
			if math.Hypot(real(got[i]-want), imag(got[i]-want)) > 1e-10 {
				t.Fatalf("T%d[%d] = %v, bignum Chebyshev = %v", n, i, got[i], want)
			}
		}
	}
}
