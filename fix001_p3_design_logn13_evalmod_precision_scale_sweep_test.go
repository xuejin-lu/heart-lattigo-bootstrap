package main

import (
	"math"
	"testing"
)

func TestFIX001P3EvalModPrecisionScaleSweepPlanScales(t *testing.T) {
	if len(precisionSweepPlanExponents) != 10 {
		t.Fatalf("candidate count = %d, want 10", len(precisionSweepPlanExponents))
	}
	for i, exponent := range precisionSweepPlanExponents {
		want := 91 + i
		if exponent != want {
			t.Fatalf("candidate[%d] exponent = %d, want %d", i, exponent, want)
		}
	}
}

func TestFIX001P3EvalModPrecisionScaleSweepCapacity(t *testing.T) {
	capacity := precisionSweepCapacityAccumulator()
	precisionSweepAddCapacity(&capacity, "first", 1.25, 0, true, true)
	precisionSweepAddCapacity(&capacity, "limiting", 0.75, 2, false, false)
	if capacity.MinimumRatio != 0.75 || capacity.WorstCheckpoint != "limiting" {
		t.Fatalf("capacity minimum = %v at %q, want 0.75 at limiting", capacity.MinimumRatio, capacity.WorstCheckpoint)
	}
	if capacity.OutsideCount != 2 || capacity.AllCenteredUnique || capacity.AllRowsMatch {
		t.Fatalf("capacity aggregate = %+v, want outside=2 and both pass flags false", capacity)
	}
}

func TestFIX001P3EvalModPrecisionScaleSweepClassification(t *testing.T) {
	path := evalModMatchedPath{FirstFailure: "none"}
	final := evalModMatchedFinal{FastRowsMatchNormalized: true, RestoreRowsMatch: true, FastVsStandard: &PSGlobalMetric{MaxComponent: 1e-5, Pass: true}}
	classification, checkpoint := precisionSweepPathClassification(path, final)
	if classification != "normalized_candidate_pass" || checkpoint != "none" {
		t.Fatalf("classification = %q at %q, want normalized_candidate_pass at none", classification, checkpoint)
	}

	final.FastVsStandard = &PSGlobalMetric{MaxComponent: math.Inf(1), Pass: false}
	classification, checkpoint = precisionSweepPathClassification(path, final)
	if classification != "polynomial_semantic_precision_failure" || checkpoint != "final.fast_vs_standard" {
		t.Fatalf("classification = %q at %q, want polynomial semantic failure", classification, checkpoint)
	}
}
