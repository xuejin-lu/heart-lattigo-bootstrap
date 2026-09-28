package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMedian(t *testing.T) {
	require.Equal(t, 3.0, median([]float64{5, 1, 3}))
	require.Equal(t, 3.0, median([]float64{4, 2}))
	require.Zero(t, median(nil))
}

func TestCompareAggregationKeepsNegativeDeltaAndClosesParent(t *testing.T) {
	baseline := []TraceRun{{Events: []Event{
		{Scope: "stage", Name: "bootstrap", ElapsedNS: 100, Sequence: 1},
		{Scope: "stage", Name: "pack", ElapsedNS: 30, Sequence: 2, ParentSequence: 1},
		{Scope: "stage", Name: "evalmod", ElapsedNS: 70, Sequence: 3, ParentSequence: 1},
	}}}
	candidate := []TraceRun{{Events: []Event{
		{Scope: "stage", Name: "bootstrap", ElapsedNS: 120, Sequence: 1},
		{Scope: "stage", Name: "pack", ElapsedNS: 28, Sequence: 2, ParentSequence: 1},
		{Scope: "stage", Name: "evalmod", ElapsedNS: 95, Sequence: 3, ParentSequence: 1},
	}}}
	deltas, closures := compareEvents(baseline, candidate)
	require.Len(t, deltas, 3)
	var pack, bootstrap EventDelta
	for _, event := range deltas {
		switch event.Name {
		case "pack":
			pack = event
		case "bootstrap":
			bootstrap = event
		}
	}
	require.Equal(t, float64(-2), pack.DeltaNS)
	require.NotNil(t, pack.ContributionToParent)
	require.Equal(t, float64(-0.1), *pack.ContributionToParent)
	require.Equal(t, float64(20), bootstrap.DeltaNS)
	require.Len(t, closures, 1)
	require.Equal(t, float64(-3), closures[0].ClosureResidualNS)
	require.Equal(t, float64(-0.15), *closures[0].ResidualFraction)
}

func TestCompareKeepsRatiosUndefinedAtZeroBaseline(t *testing.T) {
	baseline := []TraceRun{{Events: []Event{{Scope: "power", Name: "op", ElapsedNS: 0, Sequence: 1}}}}
	candidate := []TraceRun{{Events: []Event{{Scope: "power", Name: "op", ElapsedNS: 4, Sequence: 1}}}}
	deltas, _ := compareEvents(baseline, candidate)
	require.Len(t, deltas, 1)
	require.Nil(t, deltas[0].Ratio)
	require.Equal(t, float64(4), deltas[0].DeltaNS)
}

func TestCompareReportsRescalePhaseShares(t *testing.T) {
	baseline := []TraceRun{{Events: []Event{
		{Scope: "rescale", Name: "rescale", ElapsedNS: 100, Sequence: 1},
		{Scope: "rescale", Name: "preflight", ElapsedNS: 20, Sequence: 2, ParentSequence: 1},
		{Scope: "rescale", Name: "materialization", ElapsedNS: 70, Sequence: 3, ParentSequence: 1},
	}}}
	candidate := []TraceRun{{Events: []Event{
		{Scope: "rescale", Name: "rescale", ElapsedNS: 200, Sequence: 1},
		{Scope: "rescale", Name: "preflight", ElapsedNS: 50, Sequence: 2, ParentSequence: 1},
		{Scope: "rescale", Name: "materialization", ElapsedNS: 100, Sequence: 3, ParentSequence: 1},
	}}}
	shares := compareRescaleShares(baseline, candidate)
	require.Equal(t, float64(100), shares.BaselineRescaleNS)
	require.Equal(t, float64(200), shares.CandidateRescaleNS)
	require.Equal(t, float64(0.2), *shares.BaselinePreflightShare)
	require.Equal(t, float64(0.7), *shares.BaselineMaterializationShare)
	require.Equal(t, float64(0.25), *shares.CandidatePreflightShare)
	require.Equal(t, float64(0.5), *shares.CandidateMaterializationShare)
}
