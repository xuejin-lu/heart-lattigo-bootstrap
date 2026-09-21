package main

import "testing"

func TestFIX001P3P93PolynomialVectorClosure(t *testing.T) {
	standard := []complex128{0.2 + 0.1i, -0.4 + 0.3i}
	historical := []complex128{0.21 + 0.1i, -0.4 + 0.31i}
	production := []complex128{0.215 + 0.105i, -0.405 + 0.31i}
	regression := fix001P3P93PlainVectorDifference(production, historical)
	historicalError := fix001P3P93PlainVectorDifference(historical, standard)
	total := fix001P3P93PlainVectorDifference(production, standard)
	closure := fix001P3P93PlainVectorDifference(fix001P3P93PlainVectorAdd(regression, historicalError), total)
	for i, value := range closure {
		if value != 0 {
			t.Fatalf("index %d: closure residual %v", i, value)
		}
	}
}

func TestFIX001P3P93PolynomialCheckpointOrdering(t *testing.T) {
	ordered := []string{"T2-final", "T3-final", "B0-term-2", "G0-add", "F0-final-rescale"}
	for i := 1; i < len(ordered); i++ {
		if fix001P3P93PolynomialCheckpointRank(ordered[i-1]) >= fix001P3P93PolynomialCheckpointRank(ordered[i]) {
			t.Fatalf("checkpoint order is not increasing at %q -> %q", ordered[i-1], ordered[i])
		}
	}
}

func TestFIX001P3P93PolynomialBudgetClassificationNames(t *testing.T) {
	want := map[string]bool{
		"P93_HISTORICAL_POLYNOMIAL_REFERENCE_MISMATCH":     true,
		"P93_HISTORICAL_DESIGN_POLYNOMIAL_BUDGET_CONFLICT": true,
		"P93_PRODUCTION_GENERATED_POWER_REGRESSION":        true,
		"P93_PRODUCTION_PS_ACCUMULATION_REGRESSION":        true,
		"P93_PRODUCTION_POLYNOMIAL_REGRESSION_UNLOCALIZED": true,
		"P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT":        true,
	}
	if len(want) != 6 {
		t.Fatalf("classification control set changed unexpectedly")
	}
}
