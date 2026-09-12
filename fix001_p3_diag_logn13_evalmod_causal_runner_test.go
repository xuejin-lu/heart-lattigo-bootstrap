package main

import "testing"

func TestFIX001P3EvalModCausalClassification(t *testing.T) {
	comparisons := evalModCausalComparisons{
		FastNativeVsStdNative: semanticBisectMetric{Pass: false},
		FastNativeVsStdOnFast: semanticBisectMetric{Pass: true},
		StdNativeVsFastOnStd:  semanticBisectMetric{Pass: true},
		StdNativeVsStdOnFast:  semanticBisectMetric{Pass: false},
	}
	classification, verdict, asymmetry := evalModCausalClassification(comparisons)
	if classification != "logn13_evalmod_divergence_explained_by_c2s_input_sensitivity" {
		t.Fatalf("unexpected classification: %q", classification)
	}
	if verdict != "c2s_input_sensitivity" || asymmetry {
		t.Fatalf("unexpected verdict=%q asymmetry=%t", verdict, asymmetry)
	}
}

func TestFIX001P3EvalModCausalMatchedInputAsymmetry(t *testing.T) {
	comparisons := evalModCausalComparisons{
		FastNativeVsStdNative: semanticBisectMetric{Pass: false},
		FastNativeVsStdOnFast: semanticBisectMetric{Pass: true},
		StdNativeVsFastOnStd:  semanticBisectMetric{Pass: false},
		StdNativeVsStdOnFast:  semanticBisectMetric{Pass: false},
	}
	classification, verdict, asymmetry := evalModCausalClassification(comparisons)
	if classification != "" || verdict != "fast_evalmod_matched_input_divergence" || !asymmetry {
		t.Fatalf("unexpected classification=%q verdict=%q asymmetry=%t", classification, verdict, asymmetry)
	}
}
