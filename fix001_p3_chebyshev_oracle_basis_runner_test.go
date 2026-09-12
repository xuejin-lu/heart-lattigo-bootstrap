package main

import (
	"math"
	"testing"

	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

func TestFIX001P3RawChebyshevOracleUsesPreprocessedVariable(t *testing.T) {
	coeffs := make([]*bignum.Complex, 3)
	coeffs[0] = bignum.ToComplex(1.0, 128)
	coeffs[1] = bignum.ToComplex(2.0, 128)
	coeffs[2] = bignum.ToComplex(3.0, 128)
	poly := bignum.NewPolynomial(bignum.Chebyshev, coeffs, [2]float64{-2, 2})
	got := chebyshevRawOracle(poly, []complex128{0.25})[0]
	// 1 + 2*T1(0.25) + 3*T2(0.25) = 1 + .5 + 3*(2*.25^2-1).
	want := -1.125
	if math.Abs(real(got)-want) > 1e-12 || math.Abs(imag(got)) > 1e-12 {
		t.Fatalf("raw Chebyshev oracle = %v, want %v", got, want)
	}
}
