package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFIX001P3GenuineStandardVsMatchedReconciliationConstants(t *testing.T) {
	if fix001P3GenuineStandardReconciliationThreshold != 1e-2 {
		t.Fatalf("threshold = %v, want 1e-2", fix001P3GenuineStandardReconciliationThreshold)
	}
	if fix001P3GenuineStandardReconciliationMaterial != 0.10 {
		t.Fatalf("material fraction = %v, want 0.10", fix001P3GenuineStandardReconciliationMaterial)
	}
}

func TestFIX001P3GenuineStandardVsMatchedReconciliationClassificationNames(t *testing.T) {
	want := map[string]bool{
		"P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT":                       true,
		"P93_MATCHED_STANDARD_INPUT_LINEAGE_MISMATCH":                         true,
		"P93_MATCHED_STANDARD_EVALMOD_REFERENCE_MISMATCH":                     true,
		"P93_MATCHED_STANDARD_S2C_REFERENCE_MISMATCH":                         true,
		"P93_MATCHED_REFERENCE_VALID_BUT_FAST_DIVERGES_FROM_GENUINE_STANDARD": true,
		"P93_MATCHED_REFERENCE_PROXY_NOT_VALID_FOR_E2E":                       true,
		"P93_REFERENCE_RECONCILIATION_UNRESOLVED":                             true,
	}
	if len(want) != 7 {
		t.Fatalf("classification set has %d entries, want 7", len(want))
	}
}

func TestFIX001P3GenuineStandardPublicVsStagedConsistency(t *testing.T) {
	cfg, err := LoadBootstrapConfig("configs/bootstrap_config.logN13.json")
	if err != nil {
		t.Fatal(err)
	}
	outPath := t.TempDir() + "/reconciliation-summary.json"
	if err := runFIX001P3DiagP93GenuineStandardVsMatchedReconciliation(cfg, ".", "../lattigo", outPath); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var result fix001P3GenuineStandardReconciliationResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Classification != "P93_MATCHED_STANDARD_EVALMOD_REFERENCE_MISMATCH" {
		t.Fatalf("classification = %q", result.Classification)
	}
	if result.R0Standard["public_structural_equal_staged"] != true {
		t.Fatal("Genuine Standard public and staged outputs are not structurally equal")
	}
}
