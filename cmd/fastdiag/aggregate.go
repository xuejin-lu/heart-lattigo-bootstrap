package main

import (
	"fmt"
	"sort"
	"strings"
)

type eventSample struct {
	elapsedNS int64
	parentKey string
	event     Event
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return sorted[middle-1]/2 + sorted[middle]/2
}

func aggregateEvents(runs []TraceRun) []EventMedian {
	samples := keyedEventSamples(runs)
	keys := make([]string, 0, len(samples))
	for key := range samples {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	medians := make([]EventMedian, 0, len(keys))
	for _, key := range keys {
		items := samples[key]
		durations := make([]float64, len(items))
		for i, sample := range items {
			durations[i] = float64(sample.elapsedNS)
		}
		event := items[0].event
		medians = append(medians, EventMedian{
			Key: key, ParentKey: items[0].parentKey, Scope: event.Scope, Name: event.Name, Power: event.Power, SplitA: event.SplitA, SplitB: event.SplitB,
			Component: event.Component, Count: len(items), MedianNS: median(durations),
		})
	}
	return medians
}

// aggregateSingleTraceEventTypes keeps a compact count/median by event and
// parent type for diagnostic traces. It intentionally does not merge distinct
// nesting relationships or infer disjoint time shares.
func aggregateSingleTraceEventTypes(events []Event) []EventMedian {
	bySequence := make(map[uint64]Event, len(events))
	for _, event := range events {
		bySequence[event.Sequence] = event
	}
	type groupedEvents struct {
		event     Event
		parent    Event
		hasParent bool
		durations []float64
	}
	groups := make(map[string]*groupedEvents)
	for _, event := range events {
		var parent Event
		hasParent := event.ParentSequence != 0
		if hasParent {
			parent = bySequence[event.ParentSequence]
		}
		parentType := ""
		if hasParent {
			parentType = eventBaseKey(parent)
		}
		key := eventBaseKey(event) + "<-" + parentType
		group := groups[key]
		if group == nil {
			group = &groupedEvents{event: event, parent: parent, hasParent: hasParent}
			groups[key] = group
		}
		group.durations = append(group.durations, float64(event.ElapsedNS))
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]EventMedian, 0, len(keys))
	for _, key := range keys {
		group := groups[key]
		event := group.event
		parentKey := ""
		if group.hasParent {
			parentKey = eventKindLabel(group.parent)
		}
		out = append(out, EventMedian{
			Key: key, ParentKey: parentKey, Scope: event.Scope, Name: event.Name,
			Power: event.Power, SplitA: event.SplitA, SplitB: event.SplitB, Component: event.Component,
			Count: len(group.durations), MedianNS: median(group.durations),
		})
	}
	return out
}

func eventKindLabel(event Event) string {
	label := event.Scope + "/" + event.Name
	if event.Power != nil {
		label += fmt.Sprintf("/T%d", *event.Power)
	}
	if event.Component != "" {
		label += "/" + event.Component
	}
	return label
}

func compareEvents(baselineRuns, candidateRuns []TraceRun) ([]EventDelta, []ParentClosure) {
	baseline := keyedEventSamples(baselineRuns)
	candidate := keyedEventSamples(candidateRuns)
	keys := make(map[string]struct{}, len(baseline)+len(candidate))
	for key := range baseline {
		keys[key] = struct{}{}
	}
	for key := range candidate {
		keys[key] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)

	deltas := make([]EventDelta, 0, len(ordered))
	deltaByKey := make(map[string]float64, len(ordered))
	parentByKey := make(map[string]string, len(ordered))
	for _, key := range ordered {
		baseValues := sampleDurations(baseline[key])
		candidateValues := sampleDurations(candidate[key])
		baseMedian, candidateMedian := median(baseValues), median(candidateValues)
		delta := candidateMedian - baseMedian
		deltaByKey[key] = delta
		items := candidate[key]
		if len(items) == 0 {
			items = baseline[key]
		}
		if len(items) == 0 {
			continue
		}
		event := items[0].event
		if items[0].parentKey != "" {
			parentByKey[key] = items[0].parentKey
		}
		var ratio *float64
		if baseMedian != 0 {
			value := candidateMedian / baseMedian
			ratio = &value
		}
		deltas = append(deltas, EventDelta{
			Key: key, Scope: event.Scope, Name: event.Name, Power: event.Power, SplitA: event.SplitA, SplitB: event.SplitB, Component: event.Component,
			BaselineMedianNS: baseMedian, CandidateMedianNS: candidateMedian,
			Ratio: ratio, DeltaNS: delta, SampleCount: min(len(baseValues), len(candidateValues)),
		})
	}

	for i := range deltas {
		parentKey := parentByKey[deltas[i].Key]
		parentDelta, ok := deltaByKey[parentKey]
		if !ok || parentDelta == 0 {
			continue
		}
		contribution := deltas[i].DeltaNS / parentDelta
		deltas[i].ContributionToParent = &contribution
	}

	children := make(map[string]float64)
	for key, parent := range parentByKey {
		children[parent] += deltaByKey[key]
	}
	parentKeys := make([]string, 0, len(children))
	for key := range children {
		parentKeys = append(parentKeys, key)
	}
	sort.Strings(parentKeys)
	closures := make([]ParentClosure, 0, len(parentKeys))
	for _, parentKey := range parentKeys {
		parentDelta, ok := deltaByKey[parentKey]
		if !ok {
			continue
		}
		residual := parentDelta - children[parentKey]
		closure := ParentClosure{ParentKey: parentKey, ParentDeltaNS: parentDelta, ChildrenDeltaNS: children[parentKey], ClosureResidualNS: residual}
		if parentDelta != 0 {
			fraction := residual / parentDelta
			closure.ResidualFraction = &fraction
		}
		closures = append(closures, closure)
	}
	return deltas, closures
}

func compareRescaleShares(baselineRuns, candidateRuns []TraceRun) RescaleShares {
	baselineTotal, baselinePreflight, baselineMaterialization := rescaleTotals(baselineRuns)
	candidateTotal, candidatePreflight, candidateMaterialization := rescaleTotals(candidateRuns)
	return RescaleShares{
		BaselineRescaleNS: baselineTotal, CandidateRescaleNS: candidateTotal,
		BaselinePreflightShare: baselinePreflight, CandidatePreflightShare: candidatePreflight,
		BaselineMaterializationShare: baselineMaterialization, CandidateMaterializationShare: candidateMaterialization,
	}
}

func rescaleTotals(runs []TraceRun) (totalMedian float64, preflightShare, materializationShare *float64) {
	var totals, preflightShares, materializationShares []float64
	for _, run := range runs {
		var total, preflight, materialization float64
		for _, event := range run.Events {
			if event.Scope != "rescale" {
				continue
			}
			switch event.Name {
			case "rescale":
				total += float64(event.ElapsedNS)
			case "preflight":
				preflight += float64(event.ElapsedNS)
			case "materialization":
				materialization += float64(event.ElapsedNS)
			}
		}
		if total == 0 {
			continue
		}
		totals = append(totals, total)
		preflightShares = append(preflightShares, preflight/total)
		materializationShares = append(materializationShares, materialization/total)
	}
	if len(totals) == 0 {
		return 0, nil, nil
	}
	return median(totals), floatPointer(median(preflightShares)), floatPointer(median(materializationShares))
}

func floatPointer(value float64) *float64 { return &value }

func keyedEventSamples(runs []TraceRun) map[string][]eventSample {
	out := make(map[string][]eventSample)
	for _, run := range runs {
		occurrences := make(map[string]int)
		keysBySequence := make(map[uint64]string, len(run.Events))
		stableKeys := make([]string, len(run.Events))
		for i, event := range run.Events {
			baseKey := eventBaseKey(event)
			occurrences[baseKey]++
			stableKey := fmt.Sprintf("%s#%d", baseKey, occurrences[baseKey])
			stableKeys[i] = stableKey
			keysBySequence[event.Sequence] = stableKey
		}
		for i, event := range run.Events {
			parentKey := ""
			if event.ParentSequence != 0 {
				parentKey = keysBySequence[event.ParentSequence]
			}
			key := stableKeys[i]
			out[key] = append(out[key], eventSample{elapsedNS: event.ElapsedNS, parentKey: parentKey, event: event})
		}
	}
	return out
}

func eventBaseKey(event Event) string {
	return strings.Join([]string{
		event.Scope, event.Name, optionalInt(event.Power), event.Component,
		optionalInt(event.LevelIn), optionalInt(event.LevelOut), optionalInt(event.RowsIn), optionalInt(event.RowsOut),
		optionalInt(event.DegreeIn), optionalInt(event.DegreeOut), optionalBool(event.InPlace),
		optionalInt(event.SplitA), optionalInt(event.SplitB),
		optionalFloat(event.ScaleLog2In), optionalFloat(event.ScaleLog2Out),
	}, "|")
}

func optionalInt(value *int) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprint(*value)
}

func optionalBool(value *bool) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprint(*value)
}

func optionalFloat(value *float64) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%.12g", *value)
}

func sampleDurations(samples []eventSample) []float64 {
	values := make([]float64, len(samples))
	for i, sample := range samples {
		values[i] = float64(sample.elapsedNS)
	}
	return values
}
