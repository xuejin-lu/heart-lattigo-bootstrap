package main

import "testing"

func TestFIX001P3S2CErrorDecompositionVectorClosure(t *testing.T) {
	a := []complex128{1 + 2i, -3 + 4i}
	b := []complex128{0.25 - 1i, 2 + 0.5i}
	c := []complex128{-0.5 + 0.25i, 1 - 2i}
	observed := fix001P3S2CSubtract(fix001P3S2CAdd3(a, b, c), make([]complex128, len(a)))
	if len(observed) != len(a) || observed[0] != a[0]+b[0]+c[0] || observed[1] != a[1]+b[1]+c[1] {
		t.Fatal("vector closure helper changed the component-wise sum")
	}
}

func TestFIX001P3S2CErrorDecompositionThreshold(t *testing.T) {
	metric := fix001P3S2CMetric([]complex128{0, 0}, []complex128{1e-11, -1e-11}, fix001P3S2CDecompositionTolerance)
	if !metric.Pass {
		t.Fatalf("metric should pass the decomposition tolerance: %+v", metric)
	}
	if fix001P3S2CDecompositionTolerance != 1e-10 {
		t.Fatalf("decomposition tolerance = %g, want 1e-10", fix001P3S2CDecompositionTolerance)
	}
}

func TestFIX001P3S2CErrorDecompositionClassificationNames(t *testing.T) {
	if requiredFIX001P3S2CErrorDecompositionSecondary != "25e70430b10cd4af37ad4cf94cd994912e970e3f" {
		t.Fatal("unexpected Secondary provenance constant")
	}
}
