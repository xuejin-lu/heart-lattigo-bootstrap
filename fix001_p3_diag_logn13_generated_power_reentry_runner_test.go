package main

import "testing"

func TestFIX001P3GeneratedPowerReentryCasePassRequiresAllGates(t *testing.T) {
	metric := &PSGlobalMetric{Pass: true}
	caseResult := generatedPowerReentryMetricSet{Reached: true, Contracts: true, EvalModReal: metric, EvalModImag: metric, PostS2C: metric, PublicLike: metric}
	if !generatedPowerReentryCasePass(&caseResult) {
		t.Fatal("all accepted generated-power reentry gates should pass")
	}
	caseResult.Contracts = false
	if generatedPowerReentryCasePass(&caseResult) {
		t.Fatal("failed contract gate must fail generated-power reentry case")
	}
}

func TestFIX001P3GeneratedPowerReentryReference(t *testing.T) {
	if generatedPowerReentryReference <= 0 || generatedPowerReentryReference >= generatedPowerReentryPublicThreshold {
		t.Fatalf("unexpected accepted public-like reference: %g", generatedPowerReentryReference)
	}
}
