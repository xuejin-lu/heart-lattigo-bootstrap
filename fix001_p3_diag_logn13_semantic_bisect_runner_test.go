package main

import (
	"math"
	"testing"
)

func TestFIX001P3SemanticBisectMetric(t *testing.T) {
	reference := []complex128{1 + 2i, -3 + 4i}
	actual := []complex128{1.001 + 2i, -3 + 4.02i}
	metric, err := semanticBisectMetricFor(reference, actual, 1e-2)
	if err != nil {
		t.Fatal(err)
	}
	if metric.Pass {
		t.Fatalf("expected metric to fail threshold")
	}
	if metric.WorstIndex != 1 || metric.WorstComponent != "imag" {
		t.Fatalf("unexpected worst component: index=%d component=%q", metric.WorstIndex, metric.WorstComponent)
	}
	if math.Abs(metric.MaxComponentAbs-0.02) > 1e-12 {
		t.Fatalf("unexpected max component error: %v", metric.MaxComponentAbs)
	}
}
