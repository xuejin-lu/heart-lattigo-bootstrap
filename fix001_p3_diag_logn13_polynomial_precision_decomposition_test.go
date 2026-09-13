package main

import (
	"math"
	"testing"
)

func TestFIX001P3PolynomialPrecisionDecompositionPowerSet(t *testing.T) {
	want := []int{1, 2, 3, 4, 6, 8, 16}
	if len(polynomialDecompositionPowerDegrees) != len(want) {
		t.Fatalf("power set length = %d, want %d", len(polynomialDecompositionPowerDegrees), len(want))
	}
	for i, n := range want {
		if polynomialDecompositionPowerDegrees[i] != n {
			t.Fatalf("power[%d] = %d, want %d", i, polynomialDecompositionPowerDegrees[i], n)
		}
	}
}

func TestFIX001P3PolynomialPrecisionDecompositionResidualClosure(t *testing.T) {
	ref := []complex128{complex(0.25, -0.5), complex(-0.125, 0.75)}
	hActual := []complex128{complex(0.2, -0.4), complex(-0.1, 0.7)}
	hOracle := []complex128{complex(0.22, -0.45), complex(-0.11, 0.72)}
	fActual := []complex128{complex(0.3, -0.6), complex(-0.2, 0.8)}
	closure := make([]complex128, len(ref))
	for i := range closure {
		lhs := fActual[i] - ref[i]
		rhs := (fActual[i] - hActual[i]) + (hActual[i] - hOracle[i]) + (hOracle[i] - ref[i])
		closure[i] = lhs - rhs
	}
	metric := polynomialDecompositionMetric(make([]complex128, len(closure)), closure, polynomialDecompositionOracleTarget)
	if !metric.Pass || metric.MaxComponent > 1e-12 {
		t.Fatalf("decomposition closure = %g, want numerical zero", metric.MaxComponent)
	}
}

func TestFIX001P3PolynomialPrecisionDecompositionBudget(t *testing.T) {
	if math.Abs(polynomialDecompositionBudget-1.2e-8) > 0 {
		t.Fatalf("budget = %g, want 1.2e-8", polynomialDecompositionBudget)
	}
	if math.Abs(polynomialDecompositionOracleTarget-1e-10) > 0 {
		t.Fatalf("oracle target = %g, want 1e-10", polynomialDecompositionOracleTarget)
	}
}
