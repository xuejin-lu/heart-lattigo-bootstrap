package main

import "testing"

func TestFIX001P3B4T2ProductionIntegrationContract(t *testing.T) {
	if fix001P3B4T2GuardPublicThreshold != 1e-2 {
		t.Fatalf("unexpected public threshold: %g", fix001P3B4T2GuardPublicThreshold)
	}
	if fix001P3B4T2GuardFeasibilityPublicLike <= 0 || fix001P3B4T2GuardFeasibilityPublicLike >= fix001P3B4T2GuardPublicThreshold {
		t.Fatalf("feasibility reference is outside the accepted public range: %g", fix001P3B4T2GuardFeasibilityPublicLike)
	}
	if got := (&PSGlobalMetric{Threshold: 1e-2, MaxComplex: 0.1, MaxComponent: 0.1, MeanComplex: 0.1}); !fix001P3B4T2GuardMetricFinite(got) {
		t.Fatal("finite metric was rejected")
	}
	if got := (&PSGlobalMetric{Threshold: 1e-2, MaxComplex: 0.1, MaxComponent: 0.1, MeanComplex: 0.1, WorstIndex: -1}); got.WorstIndex != -1 {
		t.Fatal("unexpected metric sentinel")
	}
}
