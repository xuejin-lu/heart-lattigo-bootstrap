package main

import "testing"

func TestFIX001P3PSResidualLocalizationClassificationNames(t *testing.T) {
	if requiredFIX001P3PSLocalizationSecondary != "25e70430b10cd4af37ad4cf94cd994912e970e3f" {
		t.Fatalf("unexpected Secondary provenance: %s", requiredFIX001P3PSLocalizationSecondary)
	}
	if psLocalizationRegion("G3-add") != "G3" || psLocalizationRegion("B1-term-4") != "B1" || psLocalizationRegion("F0-root") != "F0" {
		t.Fatalf("unexpected PS region parsing")
	}
}

func TestFIX001P3PSResidualLocalizationDirection(t *testing.T) {
	a := []complex128{1, 2i, -1}
	if got := psLocalizationDirection(a, a); got < 1-1e-12 {
		t.Fatalf("identical residual direction got %g", got)
	}
	if got := psLocalizationDirection(a, []complex128{-1, -2i, 1}); got < 1-1e-12 {
		t.Fatalf("opposite residual direction should be collinear, got %g", got)
	}
}

func TestFIX001P3PSResidualLocalizationAlphaInterpolation(t *testing.T) {
	a := []complex128{1 + 2i, 3 - 4i}
	b := []complex128{5 + 6i, 7 - 8i}
	got := psLocalizationInterpolate(a, b, 0.5)
	if got[0] != 3+4i || got[1] != 5-6i {
		t.Fatalf("unexpected midpoint: %#v", got)
	}
}
