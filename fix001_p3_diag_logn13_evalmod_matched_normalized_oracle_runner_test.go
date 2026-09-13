package main

import "testing"

func TestFIX001P3EvalModMatchedNormalizedClassification(t *testing.T) {
	path := evalModMatchedPath{FirstFailure: "none"}
	final := evalModMatchedFinal{
		FastVsStandard:          &PSGlobalMetric{Pass: false},
		NormalizedVsCoherent:    &PSGlobalMetric{Pass: true},
		CoherentVsStandard:      &PSGlobalMetric{Pass: false},
		ScaleResetRowsUnchanged: true,
		RestoreRowsMatch:        true,
	}
	classification, first := evalModMatchedClassification(path, final)
	if classification != "logn13_evalmod_compressed_polynomial_semantic_divergence" || first != "final.coherent_vs_standard" {
		t.Fatalf("unexpected classification: %s (%s)", classification, first)
	}
}

func TestFIX001P3EvalModMatchedNormalizedThreshold(t *testing.T) {
	metric := evalModMatchedMetricFor([]complex128{0}, []complex128{complex(0.009, 0)}, evalModMatchedLocalThreshold)
	if !metric.Pass || metric.Threshold != 1e-2 {
		t.Fatalf("expected local threshold pass, got %+v", metric)
	}
}
