package main

import (
	"math"
	"testing"
)

func TestCompareVectorsCountsComponentThresholds(t *testing.T) {
	metric := compareVectors([]complex128{0}, []complex128{0.02 + 0.03i}, 0.01)
	if math.Abs(metric.ComplexRMSE-math.Hypot(0.02, 0.03)) > 1e-12 {
		t.Fatalf("complex RMSE=%g, want %g", metric.ComplexRMSE, math.Hypot(0.02, 0.03))
	}
	if metric.RealAboveThreshold != 1 || metric.ImagAboveThreshold != 1 {
		t.Fatalf("threshold violations real/imag=%d/%d, want 1/1", metric.RealAboveThreshold, metric.ImagAboveThreshold)
	}
}

func TestStageTopologyKeepsParallelBranchesAndJointS2CParent(t *testing.T) {
	values := map[string][]complexValue{
		"input":               {{Real: 1}},
		"scale_down":          {{Real: 1}},
		"mod_up":              {{Real: 1}},
		"c2s_real":            {{Real: 2}},
		"c2s_imag":            {{Real: 3}},
		"evalmod_real":        {{Real: 2.1}},
		"evalmod_imag":        {{Real: 3.1}},
		"s2c":                 {{Real: 4}},
		"final_public_output": {{Real: 4}},
	}
	standardVectors := probeVectors{Checkpoints: values}
	fastValues := make(map[string][]complexValue, len(values))
	for name, vector := range values {
		copyVector := append([]complexValue(nil), vector...)
		fastValues[name] = copyVector
	}
	fastValues["evalmod_real"] = []complexValue{{Real: 2.2}}
	fastValues["evalmod_imag"] = []complexValue{{Real: 3.2}}
	fastValues["s2c"] = []complexValue{{Real: 4.2}}
	fastValues["final_public_output"] = []complexValue{{Real: 4.2}}
	fastVectors := probeVectors{Checkpoints: fastValues}
	var standard, fast document
	for name := range values {
		state := ciphertextState{Level: 3, Degree: 1, ScaleLog2: 45, LogCols: 12}
		standard.Checkpoints = append(standard.Checkpoints, checkpoint{Name: name, State: state, SemanticallyValid: true})
		fast.Checkpoints = append(fast.Checkpoints, checkpoint{Name: name, State: state, SemanticallyValid: true})
	}
	stages := buildStageComparisons(standard, fast, standardVectors, fastVectors, 0.01)
	byName := make(map[string]stageComparison, len(stages))
	for _, stage := range stages {
		byName[stage.Checkpoint] = stage
	}
	if byName["c2s_real"].Parent != "mod_up" || byName["c2s_imag"].Parent != "mod_up" {
		t.Fatalf("C2S branches do not share ModUp as parent: real=%q imag=%q", byName["c2s_real"].Parent, byName["c2s_imag"].Parent)
	}
	if byName["evalmod_real"].Parent != "c2s_real" || byName["evalmod_imag"].Parent != "c2s_imag" {
		t.Fatalf("EvalMod branch parents mismatch: real=%q imag=%q", byName["evalmod_real"].Parent, byName["evalmod_imag"].Parent)
	}
	if byName["s2c"].Parent != "evalmod_real+evalmod_imag (joint)" || byName["s2c"].Amplification == nil || math.Abs(*byName["s2c"].Amplification-2) > 1e-12 {
		t.Fatalf("S2C did not use the joint branch error parent: %+v", byName["s2c"])
	}
	if byName["final_public_output"].Parent != "s2c" {
		t.Fatalf("final output parent=%q, want s2c", byName["final_public_output"].Parent)
	}
}
