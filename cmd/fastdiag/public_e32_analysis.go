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

type publicE32EventTypeSummary struct {
	TreeRootSequence     uint64   `json:"tree_root_sequence"`
	TreeRootScope        string   `json:"tree_root_scope"`
	TreeRootName         string   `json:"tree_root_name"`
	ParentType           string   `json:"parent_type,omitempty"`
	Scope                string   `json:"scope"`
	Name                 string   `json:"name"`
	Component            string   `json:"component,omitempty"`
	Count                int      `json:"count"`
	InclusiveMedianNS    float64  `json:"inclusive_median_ns"`
	ChildrenMedianNS     float64  `json:"immediate_children_median_ns"`
	ExclusiveMedianNS    float64  `json:"exclusive_median_ns"`
	ExclusiveTotalNS     float64  `json:"exclusive_total_ns"`
	ExclusiveMinNS       float64  `json:"exclusive_min_ns"`
	ExclusiveMaxNS       float64  `json:"exclusive_max_ns"`
	ExclusiveRootShare   *float64 `json:"exclusive_share_of_this_event_tree_root,omitempty"`
	ExclusiveFractionMed float64  `json:"median_exclusive_fraction_of_event"`
}

type publicE32EventTypeParetoEntry struct {
	Rank              int      `json:"rank"`
	Scope             string   `json:"scope"`
	Name              string   `json:"name"`
	Component         string   `json:"component,omitempty"`
	ParentType        string   `json:"parent_type,omitempty"`
	Count             int      `json:"count"`
	ExclusiveTotalNS  float64  `json:"exclusive_total_ns"`
	ExclusiveMedianNS float64  `json:"exclusive_median_ns"`
	RootShare         *float64 `json:"share_of_this_event_tree_root,omitempty"`
}

type publicE32EventTypeParetoGroup struct {
	RootSequence  uint64                          `json:"root_sequence"`
	RootScope     string                          `json:"root_scope"`
	RootName      string                          `json:"root_name"`
	RootElapsedNS float64                         `json:"root_elapsed_ns"`
	Entries       []publicE32EventTypeParetoEntry `json:"positive_exclusive_top10_by_event_type"`
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
	RootElapsedNS            float64                         `json:"root_elapsed_ns"`
	RootImmediateChildrenNS  float64                         `json:"root_immediate_children_ns"`
	RootUnattributedNS       float64                         `json:"root_unattributed_ns"`
	RootUnattributedFraction float64                         `json:"root_unattributed_fraction"`
	Closures                 []publicE32EventClosure         `json:"event_closures"`
	PositiveExclusivePareto  []publicE32ParetoGroup          `json:"positive_exclusive_pareto_by_event_root"`
	EventTypeSummaries       []publicE32EventTypeSummary     `json:"event_type_summaries_by_independent_root"`
	PositiveExclusiveByType  []publicE32EventTypeParetoGroup `json:"positive_exclusive_pareto_by_event_type_root"`
	AmdahlScenarios          []publicE32AmdahlScenario       `json:"stage_amdahl_scenarios"`
}

type publicE32EventTypeSamples struct {
	event          Event
	parent         Event
	hasParent      bool
	root           Event
	inclusive      []float64
	children       []float64
	exclusive      []float64
	exclusiveRatio []float64
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
	treeRootBySequence := make(map[uint64]uint64, len(events))
	for _, event := range events {
		current := event
		steps := 0
		for current.ParentSequence != 0 {
			parent, exists := bySequence[current.ParentSequence]
			if !exists {
				return publicE32TraceAnalysis{}, fmt.Errorf("event %d references missing parent %d", current.Sequence, current.ParentSequence)
			}
			current = parent
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
		RootUnattributedNS:      rootResidual,
		Closures:                make([]publicE32EventClosure, 0, len(events)),
		EventTypeSummaries:      make([]publicE32EventTypeSummary, 0),
		PositiveExclusivePareto: make([]publicE32ParetoGroup, 0),
		PositiveExclusiveByType: make([]publicE32EventTypeParetoGroup, 0),
		AmdahlScenarios:         make([]publicE32AmdahlScenario, 0),
	}
	if rootElapsed > 0 {
		analysis.RootUnattributedFraction = rootResidual / rootElapsed
	}
	ordered := append([]Event(nil), events...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	paretoGroups := make(map[uint64]*publicE32ParetoGroup)
	typeGroups := make(map[string]*publicE32EventTypeSamples)
	for _, event := range ordered {
		rootEvent := bySequence[treeRootBySequence[event.Sequence]]
		childNS := sumEventElapsed(children[event.Sequence])
		inclusiveNS := float64(event.ElapsedNS)
		exclusiveNS := inclusiveNS - childNS
		closure := publicE32EventClosure{
			Sequence: event.Sequence, ParentSequence: event.ParentSequence,
			Scope: event.Scope, Name: event.Name, Component: event.Component,
			InclusiveNS: inclusiveNS, ImmediateChildrenNS: childNS,
			UnattributedNS: exclusiveNS,
		}
		if event.ElapsedNS > 0 {
			closure.UnattributedFraction = exclusiveNS / inclusiveNS
		}
		if rootEvent.ElapsedNS > 0 {
			closure.InclusiveRootShare = inclusiveNS / float64(rootEvent.ElapsedNS)
		}
		analysis.Closures = append(analysis.Closures, closure)

		if exclusiveNS > 0 {
			group := paretoGroups[rootEvent.Sequence]
			if group == nil {
				group = &publicE32ParetoGroup{RootSequence: rootEvent.Sequence, RootScope: rootEvent.Scope,
					RootName: rootEvent.Name, RootElapsedNS: float64(rootEvent.ElapsedNS)}
				paretoGroups[rootEvent.Sequence] = group
			}
			var rootShare *float64
			if rootEvent.ElapsedNS > 0 {
				share := exclusiveNS / float64(rootEvent.ElapsedNS)
				rootShare = &share
			}
			group.Entries = append(group.Entries, publicE32ParetoEntry{
				Sequence: event.Sequence, Scope: event.Scope, Name: event.Name, Component: event.Component,
				ExclusiveNS: exclusiveNS, RootShare: rootShare,
			})
		}

		var parent Event
		hasParent := event.ParentSequence != 0
		parentType := ""
		if hasParent {
			parent = bySequence[event.ParentSequence]
			parentType = eventKindLabel(parent)
		}
		groupKey := fmt.Sprintf("%d|%s<-%s", rootEvent.Sequence, eventKindLabel(event), parentType)
		typeGroup := typeGroups[groupKey]
		if typeGroup == nil {
			typeGroup = &publicE32EventTypeSamples{event: event, parent: parent, hasParent: hasParent, root: rootEvent}
			typeGroups[groupKey] = typeGroup
		}
		typeGroup.inclusive = append(typeGroup.inclusive, inclusiveNS)
		typeGroup.children = append(typeGroup.children, childNS)
		typeGroup.exclusive = append(typeGroup.exclusive, exclusiveNS)
		if inclusiveNS > 0 {
			typeGroup.exclusiveRatio = append(typeGroup.exclusiveRatio, exclusiveNS/inclusiveNS)
		}

		if event.Scope == "stage" && event.Name != "bootstrap" && event.ParentSequence == root.Sequence && rootElapsed > 0 {
			fraction := inclusiveNS / rootElapsed
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
	typeGroupKeys := make([]string, 0, len(typeGroups))
	for key := range typeGroups {
		typeGroupKeys = append(typeGroupKeys, key)
	}
	sort.Strings(typeGroupKeys)
	typeParetoGroups := make(map[uint64]*publicE32EventTypeParetoGroup)
	for _, key := range typeGroupKeys {
		group := typeGroups[key]
		exclusiveTotal := sumFloat64(group.exclusive)
		rootElapsedNS := float64(group.root.ElapsedNS)
		summary := publicE32EventTypeSummary{
			TreeRootSequence: group.root.Sequence, TreeRootScope: group.root.Scope, TreeRootName: group.root.Name,
			Scope: group.event.Scope, Name: group.event.Name, Component: group.event.Component,
			Count: len(group.exclusive), InclusiveMedianNS: median(group.inclusive), ChildrenMedianNS: median(group.children),
			ExclusiveMedianNS: median(group.exclusive), ExclusiveTotalNS: exclusiveTotal,
			ExclusiveMinNS: minFloat64(group.exclusive), ExclusiveMaxNS: maxFloat64(group.exclusive),
		}
		if group.hasParent {
			summary.ParentType = eventKindLabel(group.parent)
		}
		if len(group.exclusiveRatio) > 0 {
			summary.ExclusiveFractionMed = median(group.exclusiveRatio)
		}
		if rootElapsedNS > 0 {
			share := exclusiveTotal / rootElapsedNS
			summary.ExclusiveRootShare = &share
		}
		analysis.EventTypeSummaries = append(analysis.EventTypeSummaries, summary)

		if exclusiveTotal <= 0 {
			continue
		}
		pareto := typeParetoGroups[group.root.Sequence]
		if pareto == nil {
			pareto = &publicE32EventTypeParetoGroup{RootSequence: group.root.Sequence, RootScope: group.root.Scope,
				RootName: group.root.Name, RootElapsedNS: rootElapsedNS}
			typeParetoGroups[group.root.Sequence] = pareto
		}
		var rootShare *float64
		if rootElapsedNS > 0 {
			share := exclusiveTotal / rootElapsedNS
			rootShare = &share
		}
		entry := publicE32EventTypeParetoEntry{
			Scope: group.event.Scope, Name: group.event.Name, Component: group.event.Component,
			Count: len(group.exclusive), ExclusiveTotalNS: exclusiveTotal,
			ExclusiveMedianNS: median(group.exclusive), RootShare: rootShare,
		}
		if group.hasParent {
			entry.ParentType = eventKindLabel(group.parent)
		}
		pareto.Entries = append(pareto.Entries, entry)
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
	typeRootSequences := make([]uint64, 0, len(typeParetoGroups))
	for sequence := range typeParetoGroups {
		typeRootSequences = append(typeRootSequences, sequence)
	}
	sort.Slice(typeRootSequences, func(i, j int) bool { return typeRootSequences[i] < typeRootSequences[j] })
	for _, sequence := range typeRootSequences {
		group := typeParetoGroups[sequence]
		sort.Slice(group.Entries, func(i, j int) bool {
			if group.Entries[i].ExclusiveTotalNS == group.Entries[j].ExclusiveTotalNS {
				left := group.Entries[i].Scope + group.Entries[i].Name + group.Entries[i].Component + group.Entries[i].ParentType
				right := group.Entries[j].Scope + group.Entries[j].Name + group.Entries[j].Component + group.Entries[j].ParentType
				return left < right
			}
			return group.Entries[i].ExclusiveTotalNS > group.Entries[j].ExclusiveTotalNS
		})
		if len(group.Entries) > 10 {
			group.Entries = group.Entries[:10]
		}
		for i := range group.Entries {
			group.Entries[i].Rank = i + 1
		}
		analysis.PositiveExclusiveByType = append(analysis.PositiveExclusiveByType, *group)
	}
	return analysis, nil
}

func sumFloat64(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum
}

func minFloat64(values []float64) float64 {
	minValue := values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

func maxFloat64(values []float64) float64 {
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
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
