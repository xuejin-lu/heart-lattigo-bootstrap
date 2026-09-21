package main

import "testing"

func TestFIX001P3P93S2CAttributionConstants(t *testing.T) {
	if fix001P3P93S2CThreshold != 1e-2 {
		t.Fatalf("threshold = %g, want 1e-2", fix001P3P93S2CThreshold)
	}
	if fix001P3P93S2CSecondaryCommit != "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797" {
		t.Fatal("unexpected Secondary provenance")
	}
	if fix001P3P93S2CDiffSHA256 == "" {
		t.Fatal("missing Secondary dirty-diff fingerprint")
	}
}

func TestFIX001P3P93S2CAttributionRatioMetric(t *testing.T) {
	if got := ratioMetric(2, 4); got != 0.5 {
		t.Fatalf("ratioMetric(2,4) = %g, want 0.5", got)
	}
	if got := ratioMetric(1, 0); got != 0 {
		t.Fatalf("ratioMetric(1,0) = %g, want 0", got)
	}
}
