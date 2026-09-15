package main

import "testing"

func TestFIX001P3DiagPublicFinalization(t *testing.T) {
	cfg, err := LoadBootstrapConfig("configs/bootstrap_config.logN13.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := runFIX001P3DiagPublicFinalization(cfg, ".", "../lattigo", "results/FIX-001-P3-DIAG-PUBLIC-FINALIZATION-summary.json"); err != nil {
		t.Fatal(err)
	}
}
