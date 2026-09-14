package main

import "testing"

func TestFIX001P3G0MergeAttributionClassification(t *testing.T) {
	if requiredFIX001P3G0MergePrimary != "13d54221f7feac63c860b3bb6cf81227ceb65240" {
		t.Fatalf("unexpected Primary provenance: %s", requiredFIX001P3G0MergePrimary)
	}
	if requiredFIX001P3G0MergeSecondary != "25e70430b10cd4af37ad4cf94cd994912e970e3f" {
		t.Fatalf("unexpected Secondary provenance: %s", requiredFIX001P3G0MergeSecondary)
	}
}

func TestFIX001P3G0MergeAttributionThresholds(t *testing.T) {
	if g0MergePublicThreshold != 1e-2 || g0MergeSemanticThreshold != 1e-10 {
		t.Fatalf("unexpected G0 merge thresholds: public=%g semantic=%g", g0MergePublicThreshold, g0MergeSemanticThreshold)
	}
}
