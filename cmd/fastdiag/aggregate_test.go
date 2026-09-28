package main

import (
	"strings"
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

func TestNestedRescaleClosureUsesOnlyImmediateChildren(t *testing.T) {
	makeRun := func(candidate bool) TraceRun {
		timeNS := func(baseline, changed int64) int64 {
			if candidate {
				return changed
			}
			return baseline
		}
		event := func(name string, baseline, changed int64, sequence, parent uint64, component string) Event {
			return Event{Scope: "rescale", Name: name, ElapsedNS: timeNS(baseline, changed), Sequence: sequence, ParentSequence: parent, Component: component}
		}
		return TraceRun{Events: []Event{
			event("rescale", 1000, 1100, 1, 0, ""),
			event("preflight", 400, 450, 2, 1, ""),
			event("materialization", 600, 650, 3, 1, ""),
			event("prefix_to_coefficient", 50, 55, 4, 2, "c0"),
			event("coefficient_loop", 350, 395, 5, 2, "c0"),
			event("reconstruct_center_round_capacity", 350, 395, 6, 5, "c0"),
			event("prefix_to_coefficient", 50, 55, 7, 3, "c0"),
			event("coefficient_loop", 500, 540, 8, 3, "c0"),
			event("reconstruct_center_round_capacity", 300, 330, 9, 8, "c0"),
			event("residue_materialization", 200, 210, 10, 8, "c0"),
			event("ntt_montgomery_restore", 50, 55, 11, 3, "c0"),
		}}
	}

	_, closures := compareEvents([]TraceRun{makeRun(false)}, []TraceRun{makeRun(true)})
	require.Len(t, closures, 5)
	byName := map[string][]ParentClosure{}
	for _, closure := range closures {
		parts := strings.Split(closure.ParentKey, "|")
		require.GreaterOrEqual(t, len(parts), 2)
		byName[parts[1]] = append(byName[parts[1]], closure)
		require.Zero(t, closure.ClosureResidualNS, "parent %s must close over only its direct children", closure.ParentKey)
	}
	require.Equal(t, float64(100), byName["rescale"][0].ChildrenDeltaNS)
	require.Equal(t, float64(50), byName["preflight"][0].ChildrenDeltaNS)
	require.Equal(t, float64(50), byName["materialization"][0].ChildrenDeltaNS)
	require.Len(t, byName["coefficient_loop"], 2)
	loopChildren := []float64{byName["coefficient_loop"][0].ChildrenDeltaNS, byName["coefficient_loop"][1].ChildrenDeltaNS}
	require.ElementsMatch(t, []float64{40, 45}, loopChildren)
}
