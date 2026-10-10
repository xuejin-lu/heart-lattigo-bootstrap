package main

import (
	"errors"
	"fmt"
	"sort"
)

type publicE32EventClosure struct {
	Sequence             uint64  `json:"sequence"`
	ParentSequence       uint64  `json:"parent_sequence,omitempty"`
	Scope                string  `json:"scope"`
	Name                 string  `json:"name"`
	Component            string  `json:"component,omitempty"`
	InclusiveNS          float64 `json:"inclusive_ns"`
	ImmediateChildrenNS  float64 `json:"immediate_children_ns"`
	UnattributedNS       float64 `json:"unattributed_ns"`
	UnattributedFraction float64 `json:"unattributed_fraction_of_event"`
	InclusiveRootShare   float64 `json:"inclusive_root_share"`
}

type publicE32ParetoEntry struct {
	Rank        int      `json:"rank"`
	Sequence    uint64   `json:"sequence"`
	Scope       string   `json:"scope"`
	Name        string   `json:"name"`
	Component   string   `json:"component,omitempty"`
	ExclusiveNS float64  `json:"exclusive_ns"`
	RootShare   *float64 `json:"share_of_event_tree_root,omitempty"`
}

type publicE32ParetoGroup struct {
	RootSequence  uint64                 `json:"root_sequence"`
	RootScope     string                 `json:"root_scope"`
	RootName      string                 `json:"root_name"`
	RootElapsedNS float64                `json:"root_elapsed_ns"`
	Entries       []publicE32ParetoEntry `json:"positive_exclusive_top10"`
}

type publicE32AmdahlScenario struct {
	Stage                     string   `json:"stage"`
	ObservedRootFraction      float64  `json:"observed_root_fraction"`
	TwoXStageOverallSpeedup   *float64 `json:"two_x_stage_overall_speedup,omitempty"`
	FourXStageOverallSpeedup  *float64 `json:"four_x_stage_overall_speedup,omitempty"`
	StageEliminatedUpperBound *float64 `json:"stage_eliminated_upper_bound,omitempty"`
	Interpretation            string   `json:"interpretation"`
}

type publicE32TraceAnalysis struct {
	RootElapsedNS            float64                   `json:"root_elapsed_ns"`
	RootImmediateChildrenNS  float64                   `json:"root_immediate_children_ns"`
	RootUnattributedNS       float64                   `json:"root_unattributed_ns"`
	RootUnattributedFraction float64                   `json:"root_unattributed_fraction"`
	Closures                 []publicE32EventClosure   `json:"event_closures"`
	PositiveExclusivePareto  []publicE32ParetoGroup    `json:"positive_exclusive_pareto_by_event_root"`
	AmdahlScenarios          []publicE32AmdahlScenario `json:"stage_amdahl_scenarios"`
}

func analyzePublicE32Events(events []Event) (publicE32TraceAnalysis, error) {
	if len(events) == 0 {
		return publicE32TraceAnalysis{}, errors.New("cannot analyze an empty E32 event trace")
	}
	bySequence := make(map[uint64]Event, len(events))
	children := make(map[uint64][]Event, len(events))
	var root Event
	for _, event := range events {
		if event.Sequence == 0 || event.ElapsedNS < 0 {
			return publicE32TraceAnalysis{}, fmt.Errorf("event %q/%q has zero sequence or negative elapsed time", event.Scope, event.Name)
		}
		if _, exists := bySequence[event.Sequence]; exists {
			return publicE32TraceAnalysis{}, fmt.Errorf("duplicate event sequence %d", event.Sequence)
		}
		bySequence[event.Sequence] = event
		if event.ParentSequence != 0 {
			children[event.ParentSequence] = append(children[event.ParentSequence], event)
		}
		if event.Scope == "stage" && event.Name == "bootstrap" {
			if root.Sequence != 0 || event.ParentSequence != 0 {
				return publicE32TraceAnalysis{}, errors.New("trace must contain exactly one root Bootstrap event")
			}
			root = event
		}
	}
	if root.Sequence == 0 {
		return publicE32TraceAnalysis{}, errors.New("trace is missing root Bootstrap event")
	}
	for _, event := range events {
		if event.ParentSequence != 0 {
			if _, exists := bySequence[event.ParentSequence]; !exists {
				return publicE32TraceAnalysis{}, fmt.Errorf("event %d references missing parent %d", event.Sequence, event.ParentSequence)
			}
		}
	}
	treeRootBySequence := make(map[uint64]uint64, len(events))
	for _, event := range events {
		current := event
		steps := 0
		for current.ParentSequence != 0 {
			current = bySequence[current.ParentSequence]
			steps++
			if steps > len(events) {
				return publicE32TraceAnalysis{}, fmt.Errorf("event %d belongs to a cyclic event-parent chain", event.Sequence)
			}
		}
		treeRootBySequence[event.Sequence] = current.Sequence
	}
	rootElapsed := float64(root.ElapsedNS)
	rootChildren := sumEventElapsed(children[root.Sequence])
	rootResidual := rootElapsed - rootChildren
	analysis := publicE32TraceAnalysis{
		RootElapsedNS: rootElapsed, RootImmediateChildrenNS: rootChildren,
		RootUnattributedNS: rootResidual, Closures: make([]publicE32EventClosure, 0, len(events)),
		AmdahlScenarios: make([]publicE32AmdahlScenario, 0),
	}
	if rootElapsed > 0 {
		analysis.RootUnattributedFraction = rootResidual / rootElapsed
	}
	paretoGroups := make(map[uint64]*publicE32ParetoGroup)
	ordered := append([]Event(nil), events...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	for _, event := range ordered {
		childNS := sumEventElapsed(children[event.Sequence])
		unattributedNS := float64(event.ElapsedNS) - childNS
		closure := publicE32EventClosure{
			Sequence: event.Sequence, ParentSequence: event.ParentSequence,
			Scope: event.Scope, Name: event.Name, Component: event.Component,
			InclusiveNS: float64(event.ElapsedNS), ImmediateChildrenNS: childNS,
			UnattributedNS: unattributedNS, InclusiveRootShare: 0,
		}
		if event.ElapsedNS > 0 {
			closure.UnattributedFraction = unattributedNS / float64(event.ElapsedNS)
		}
		if rootElapsed > 0 {
			closure.InclusiveRootShare = float64(event.ElapsedNS) / rootElapsed
		}
		analysis.Closures = append(analysis.Closures, closure)
		if unattributedNS > 0 {
			rootSequence := treeRootBySequence[event.Sequence]
			rootEvent := bySequence[rootSequence]
			group := paretoGroups[rootSequence]
			if group == nil {
				group = &publicE32ParetoGroup{RootSequence: rootSequence, RootScope: rootEvent.Scope, RootName: rootEvent.Name, RootElapsedNS: float64(rootEvent.ElapsedNS)}
				paretoGroups[rootSequence] = group
			}
			var rootShare *float64
			if rootEvent.ElapsedNS > 0 {
				share := unattributedNS / float64(rootEvent.ElapsedNS)
				rootShare = &share
			}
			group.Entries = append(group.Entries, publicE32ParetoEntry{
				Sequence: event.Sequence, Scope: event.Scope, Name: event.Name, Component: event.Component,
				ExclusiveNS: unattributedNS, RootShare: rootShare,
			})
		}
		if event.Scope == "stage" && event.Name != "bootstrap" && event.ParentSequence == root.Sequence && rootElapsed > 0 {
			fraction := float64(event.ElapsedNS) / rootElapsed
			scenario := publicE32AmdahlScenario{Stage: event.Name, ObservedRootFraction: fraction,
				Interpretation: "hypothetical stage-only acceleration from observed inclusive fraction; not a performance prediction"}
			if fraction >= 0 && fraction <= 1 {
				twoX, fourX := amdahlOverallSpeedup(fraction, 2), amdahlOverallSpeedup(fraction, 4)
				scenario.TwoXStageOverallSpeedup = &twoX
				scenario.FourXStageOverallSpeedup = &fourX
				if fraction < 1 {
					upper := 1 / (1 - fraction)
					scenario.StageEliminatedUpperBound = &upper
				}
			} else {
				scenario.Interpretation = "not computed: observed stage fraction is outside [0,1]; preserve measured overlap/closure discrepancy"
			}
			analysis.AmdahlScenarios = append(analysis.AmdahlScenarios, scenario)
		}
	}
	rootSequences := make([]uint64, 0, len(paretoGroups))
	for sequence := range paretoGroups {
		rootSequences = append(rootSequences, sequence)
	}
	sort.Slice(rootSequences, func(i, j int) bool { return rootSequences[i] < rootSequences[j] })
	for _, sequence := range rootSequences {
		group := paretoGroups[sequence]
		sort.Slice(group.Entries, func(i, j int) bool {
			if group.Entries[i].ExclusiveNS == group.Entries[j].ExclusiveNS {
				return group.Entries[i].Sequence < group.Entries[j].Sequence
			}
			return group.Entries[i].ExclusiveNS > group.Entries[j].ExclusiveNS
		})
		if len(group.Entries) > 10 {
			group.Entries = group.Entries[:10]
		}
		for i := range group.Entries {
			group.Entries[i].Rank = i + 1
		}
		analysis.PositiveExclusivePareto = append(analysis.PositiveExclusivePareto, *group)
	}
	return analysis, nil
}

func sumEventElapsed(events []Event) float64 {
	var sum float64
	for _, event := range events {
		sum += float64(event.ElapsedNS)
	}
	return sum
}

func amdahlOverallSpeedup(fraction, stageSpeedup float64) float64 {
	return 1 / ((1 - fraction) + fraction/stageSpeedup)
}
