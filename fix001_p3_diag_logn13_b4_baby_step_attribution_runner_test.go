package main

import "testing"

func TestFIX001P3B4BabyStepAttributionProvenance(t *testing.T) {
	if requiredFIX001P3B4Primary != "33ff9bd8392685e31433a4956ea36309de4fd9de" {
		t.Fatalf("unexpected Primary provenance: %s", requiredFIX001P3B4Primary)
	}
	if requiredFIX001P3B4Secondary != "25e70430b10cd4af37ad4cf94cd994912e970e3f" {
		t.Fatalf("unexpected Secondary provenance: %s", requiredFIX001P3B4Secondary)
	}
}

func TestFIX001P3B4BabyStepAttributionClassification(t *testing.T) {
	semantics := map[string]b4BabySemantics{}
	resets := map[string][]b4BabyReset{
		"real": {{ID: "B4-term-4", Pass: false}, {ID: "B4-term-2", Pass: true}},
	}
	classification, _, target := b4BabyClassify(true, semantics, resets, map[string]b4BabyLedger{})
	if classification != "logn13_b4_single_scalar_operation_blocker" || target != "B4-term-2" {
		t.Fatalf("unexpected classification=%s target=%s", classification, target)
	}
}
