package main

import (
	"math"
	"testing"
)

func TestCompareComplexVectorsMetrics(t *testing.T) {
	reference := []complex128{1 + 2i, -3 - 4i, 0}
	actual := []complex128{1.003 + 1.996i, -2.99 - 4.01i, 0.001 + 0.002i}
	metrics, err := compareComplexVectors(reference, actual, 1e-2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(metrics.MaxAbsReal-0.01) > 1e-12 || math.Abs(metrics.MaxAbsImag-0.01) > 1e-12 || math.Abs(metrics.MaxComponentAbs-0.01) > 1e-12 {
		t.Fatalf("unexpected component maxima: %#v", metrics)
	}
	if len(metrics.MaxComponentIndices) != 2 || metrics.MaxComponentIndices[0] != 1 || metrics.MaxComponentIndices[1] != 1 {
		t.Fatalf("unexpected component indices: %v", metrics.MaxComponentIndices)
	}
	if metrics.MaxAbsComplex <= 0 || metrics.RMSEComplex <= 0 || metrics.PrecisionBits == nil {
		t.Fatalf("missing complex metrics: %#v", metrics)
	}
	if !metrics.PassThreshold {
		t.Fatalf("expected threshold pass: %#v", metrics)
	}
}

func TestCompareComplexVectorsExactZero(t *testing.T) {
	values := []complex128{0, 1 - 2i}
	metrics, err := compareComplexVectors(values, values, 1e-2)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.MaxAbsComplex != 0 || metrics.PrecisionBits != nil || metrics.PrecisionBitsReason == "" {
		t.Fatalf("unexpected exact-zero metrics: %#v", metrics)
	}
	if !metrics.PassThreshold || math.Signbit(metrics.MaxComponentAbs) {
		t.Fatalf("unexpected exact-zero threshold result: %#v", metrics)
	}
}
