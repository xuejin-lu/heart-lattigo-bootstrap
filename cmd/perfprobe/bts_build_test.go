package main

import (
	"strings"
	"testing"
)

func TestBTSModfileUsesIsolatedNestedModuleAndPinnedLocalReplacements(t *testing.T) {
	original := "module mys2c\n\ngo 1.22\n\nrequire github.com/tuneinsight/lattigo/v6 v6.0.0\n"
	got, err := rewriteBTSModfile(original, "/primary/root", "/secondary/root")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"module " + bts008NestedModulePath,
		"replace github.com/xuejin-lu/heart-lattigo-bootstrap => /primary/root",
		"replace github.com/tuneinsight/lattigo/v6 => /secondary/root",
		"require github.com/xuejin-lu/heart-lattigo-bootstrap v0.0.0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rewritten modfile missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "module mys2c") {
		t.Fatalf("original standalone BTS module was not isolated:\n%s", got)
	}
}

func TestBTSOriginalSourceManifestMatchesApprovedBatch008(t *testing.T) {
	got, err := btsOriginalSourceManifest("../..")
	if err != nil {
		t.Fatal(err)
	}
	if got != bts008SourceManifestSHA {
		t.Fatalf("source manifest=%s, want %s", got, bts008SourceManifestSHA)
	}
}

func TestBTSBuildHashesEveryLinkedPrimaryMeasurementSource(t *testing.T) {
	hashes, err := hashPrimaryMeasurementSources("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(hashes) == 0 {
		t.Fatal("linked Primary measurement sources were not fingerprinted")
	}
	for _, required := range []string{
		"internal/perfmeasure/bootstrap_budget.go",
		"internal/numericalmetrics/snr.go",
	} {
		if hashes[required] == "" {
			t.Errorf("source hash missing for %s", required)
		}
	}
}
