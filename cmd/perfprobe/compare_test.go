package main

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestFormalPairAllowsDistinctCiphertextMetadataButRejectsLegacyOrigins(t *testing.T) {
	original := []complexValue{{Real: 0.125}}
	inputHash := perfmeasure.Fingerprint([]complex128{0.125})
	makeDocument := func(backend, kind, metadata string, c1Nonzero bool, trials int) document {
		state := ciphertextState{Level: 0, Degree: 1, LogCols: 0}
		constructor := fastDirectConstructor
		if backend == "standard" {
			constructor = standardNativeConstructor
		}
		evidence := inputRecord{Kind: kind, Constructor: constructor, State: state, MetadataSHA256: metadata,
			C1SHA256: "c1-0", C1Nonzero: c1Nonzero, PreDecodedSHA256: inputHash, InputQualityLimit: inputQualityLimit}
		commit := pinnedFastSHA
		if backend == "standard" {
			commit = pinnedStandardSHA
		}
		doc := document{SchemaVersion: "fast-standard-perfprobe.v2", Backend: backend, BackendCommit: commit, Profile: "test", PrimaryCommit: "same",
			ConfigPath: "same", ConfigSHA256: "same", InputSHA256: inputHash, InputMetadataSHA256: metadata,
			InputState: state, InputKind: kind, InputEvidence: evidence,
			Parameters: effectiveParameters{InputSlots: 1}, Trials: make([]trial, trials)}
		for i := range doc.Trials {
			trialEvidence := evidence
			if c1Nonzero {
				trialEvidence.C1SHA256 = string(rune('a' + i))
			}
			doc.Trials[i] = trial{FreshKeyTrial: c1Nonzero, PreDecodedSHA256: inputHash, InputEvidence: trialEvidence}
		}
		doc.InputEvidence = doc.Trials[0].InputEvidence
		return doc
	}
	standard := makeDocument("standard", standardNativeInputKind, "standard-metadata", true, 3)
	fast := makeDocument("fast", fastDirectInputKind, "fast-metadata", false, 2)
	standardVectors := probeVectors{Original: original, Trials: make([][]complexValue, 3), PreTrials: make([][]complexValue, 3)}
	fastVectors := probeVectors{Original: original, Trials: make([][]complexValue, 2), PreTrials: make([][]complexValue, 2)}
	for i := range standardVectors.PreTrials {
		standardVectors.PreTrials[i] = original
	}
	for i := range fastVectors.PreTrials {
		fastVectors.PreTrials[i] = original
	}
	if err := validatePairedInputs(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("distinct ciphertext metadata should be valid: %v", err)
	}
	legacy := standard
	legacy.InputKind = "c0=encoded-message,c1=0"
	if err := validatePairedInputs(legacy, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("formal comparison accepted legacy synthetic Standard input")
	}
	legacy = standard
	legacy.SchemaVersion = "fast-standard-perfprobe.v1"
	if err := validatePairedInputs(legacy, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("formal comparison accepted legacy schema")
	}
	legacy = standard
	legacy.Trials[0].InputEvidence.C1Nonzero = false
	if err := validatePairedInputs(legacy, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("formal comparison accepted zero-c1 Standard trial")
	}
}

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

func TestFastdiagCapacityAuditReadsStrictTwoBLessThanSQ(t *testing.T) {
	var evidence fastdiagCapacityEvidence
	if err := json.Unmarshal([]byte(`{"evalmod_qprefix_capacity_audit":[{"checkpoint":"real/evalmod-entry","strict_2b_lt_sq":true}]}`), &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.EvalModCapacityAudit) != 1 || !evidence.EvalModCapacityAudit[0].StrictFit {
		t.Fatalf("capacity evidence=%+v, want strict capacity pass", evidence.EvalModCapacityAudit)
	}
}
