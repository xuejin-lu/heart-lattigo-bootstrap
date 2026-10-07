package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func formalPairFixture() (document, document, probeVectors, probeVectors) {
	original := []complexValue{{Real: 0.125}}
	originalHash := perfmeasure.Fingerprint([]complex128{0.125})
	state := ciphertextState{Level: 0, Degree: 1, LogCols: 0}
	params := effectiveParameters{InputSlots: 1, LogSlots: 0, Q0Target: 55, Q0Bits: 55, QPrimes: []string{"prime-q0"}}
	makeDocument := func(backend string, trialOutputs []complexValue) document {
		kind, constructor, c1Nonzero, commit, metadata := fastDirectInputKind, fastDirectConstructor, false, pinnedFastSHA, "fast-metadata"
		if backend == "standard" {
			kind, constructor, c1Nonzero, commit, metadata = standardNativeInputKind, standardNativeConstructor, true, pinnedStandardSHA, "standard-metadata"
		}
		evidence := inputRecord{Kind: kind, Constructor: constructor, State: state, MetadataSHA256: metadata,
			C1SHA256: backend + "-c1-1", C1Nonzero: c1Nonzero, PreDecodedSHA256: originalHash,
			MaxComplexDeviation: 0, InputQualityLimit: inputQualityLimit}
		doc := document{SchemaVersion: "fast-standard-perfprobe.v2", Backend: backend, BackendCommit: commit,
			Profile: "logn13", PrimaryCommit: "primary-sha", ConfigPath: "configs/test.json", ConfigSHA256: "config-sha",
			InputSHA256: originalHash, InputMetadataSHA256: metadata, InputState: state, InputKind: kind, InputEvidence: evidence,
			Parameters: params, GoVersion: "go-test", OS: "testos", Arch: "testarch", CPU: "testcpu", NumCPU: 1, GOMAXPROCS: 1,
			Timing: timingResult{MedianNS: 100, MeanNS: 100, MinNS: 100, MaxNS: 100}, Trials: make([]trial, len(trialOutputs))}
		for i, output := range trialOutputs {
			pre := []complex128{0.125}
			post := []complex128{complex(output.Real, output.Imag)}
			trialEvidence := evidence
			if c1Nonzero {
				trialEvidence.C1SHA256 = backend + "-c1-" + string(rune('1'+i))
			}
			trialEvidence.PreDecodedSHA256 = perfmeasure.Fingerprint(pre)
			doc.Trials[i] = trial{Index: i + 1, BootstrapSNR: numericalmetrics.Compare(pre, post), FreshKeyTrial: c1Nonzero,
				EvaluatorPath: backend + "-evaluator", OriginalSHA256: originalHash,
				PreDecodedSHA256: trialEvidence.PreDecodedSHA256, PostDecodedSHA256: perfmeasure.Fingerprint(post), InputEvidence: trialEvidence}
		}
		doc.InputEvidence = doc.Trials[0].InputEvidence
		return doc
	}
	standard := makeDocument("standard", []complexValue{{Real: 0.126}, {Real: 0.127}, {Real: 0.128}})
	fast := makeDocument("fast", []complexValue{{Real: 0.129}, {Real: 0.130}})
	pre := [][]complexValue{{{Real: 0.125}}, {{Real: 0.125}}, {{Real: 0.125}}}
	standardPost := [][]complexValue{{{Real: 0.126}}, {{Real: 0.127}}, {{Real: 0.128}}}
	fastPre := [][]complexValue{{{Real: 0.125}}, {{Real: 0.125}}}
	fastPost := [][]complexValue{{{Real: 0.129}}, {{Real: 0.130}}}
	return standard, fast,
		probeVectors{Original: original, PreTrials: pre, Trials: standardPost},
		probeVectors{Original: original, PreTrials: fastPre, Trials: fastPost}
}

func TestFormalPairRequiresV2NativeStandardAndFastInputEvidence(t *testing.T) {
	standard, fast, standardVectors, fastVectors := formalPairFixture()
	if err := validatePairedInputs(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("valid v2 pair rejected: %v", err)
	}
	if _, err := buildPairedNumericalDocument(standard, fast, standardVectors, fastVectors, defaultComparisonThreshold); err != nil {
		t.Fatalf("valid v2 pair could not build formal report without diagnostic input: %v", err)
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
	legacy.Trials = append([]trial(nil), standard.Trials...)
	legacy.Trials[0].InputEvidence.C1Nonzero = false
	if err := validatePairedInputs(legacy, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("formal comparison accepted zero-c1 Standard trial")
	}
}

func TestFormalPairRejectsAbsentOrContradictoryTrialProvenance(t *testing.T) {
	standard, fast, standardVectors, fastVectors := formalPairFixture()
	missing := standard
	missing.Trials = append([]trial(nil), standard.Trials...)
	missing.Trials[1].InputEvidence.C1SHA256 = ""
	if err := validatePairedInputs(missing, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("formal comparison accepted absent per-trial Standard C1 evidence")
	}
	contradictory := standardVectors
	contradictory.PreTrials = append([][]complexValue(nil), standardVectors.PreTrials...)
	contradictory.PreTrials[0] = []complexValue{{Real: 0.25}}
	if err := validatePairedInputs(standard, fast, contradictory, fastVectors); err == nil {
		t.Fatal("formal comparison accepted a pre-vector contradicting recorded provenance")
	}
}

func TestFormalReportKeepsOutputAgainstOriginalPrimaryAndOmitsLegacyDiagnostic(t *testing.T) {
	standard, fast, standardVectors, fastVectors := formalPairFixture()
	doc, err := buildPairedNumericalDocument(standard, fast, standardVectors, fastVectors, defaultComparisonThreshold)
	if err != nil {
		t.Fatal(err)
	}
	if doc.QualityAssessment != "UNASSESSED" || doc.ReviewStatus != "REQUIRES_WEB_REVIEW" {
		t.Fatalf("formal assessment=%s review=%s, want unassessed/review", doc.QualityAssessment, doc.ReviewStatus)
	}
	if len(doc.StandardOriginalRMSE) != 3 || len(doc.FastOriginalRMSE) != 2 || len(doc.TrialComparisons) != 6 {
		t.Fatalf("primary output/original or secondary pairwise evidence missing: Standard=%d Fast=%d pairs=%d", len(doc.StandardOriginalRMSE), len(doc.FastOriginalRMSE), len(doc.TrialComparisons))
	}
	report := renderPairedReport(doc)
	if strings.Index(report, "Output versus original input — primary") < 0 || strings.Index(report, "Fast-versus-Standard pairwise comparison — secondary") < 0 ||
		strings.Index(report, "Output versus original input — primary") > strings.Index(report, "Fast-versus-Standard pairwise comparison — secondary") {
		t.Fatal("report does not present output-versus-original evidence before pairwise evidence")
	}
	for _, forbidden := range []string{"FAST-STANDARD-PERF-REBASELINE-002", "Fastdiag classification", "Classification:", "fastdiag_stage_lockstep"} {
		if strings.Contains(report, forbidden) {
			t.Errorf("formal report contains legacy diagnostic reference %q", forbidden)
		}
	}
	if !strings.Contains(report, "Quality assessment: `UNASSESSED`") || !strings.Contains(report, "Review status: `REQUIRES_WEB_REVIEW`") {
		t.Fatal("formal report does not clearly state unassessed quality and required review")
	}
}

func TestRunCompareNeedsOnlyPinnedProbesAndTrialVectors(t *testing.T) {
	standard, fast, standardVectors, fastVectors := formalPairFixture()
	root := t.TempDir()
	writeJSON := func(name string, value any) string {
		t.Helper()
		path := filepath.Join(root, name)
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	standardPath := writeJSON("standard.json", standard)
	fastPath := writeJSON("fast.json", fast)
	standardVectorsPath := writeJSON("standard-vectors.json", standardVectors)
	fastVectorsPath := writeJSON("fast-vectors.json", fastVectors)
	outputPath := filepath.Join(root, "paired.json")
	reportPath := filepath.Join(root, "paired.md")
	if err := runCompare([]string{"--standard", standardPath, "--fast", fastPath,
		"--standard-vectors", standardVectorsPath, "--fast-vectors", fastVectorsPath,
		"--out", outputPath, "--report", reportPath}); err != nil {
		t.Fatalf("formal compare unexpectedly requires a diagnostic attachment: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("paired JSON was not created: %v", err)
	}
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("paired report was not created: %v", err)
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
		"input": {{Real: 1}}, "scale_down": {{Real: 1}}, "mod_up": {{Real: 1}},
		"c2s_real": {{Real: 2}}, "c2s_imag": {{Real: 3}},
		"evalmod_real": {{Real: 2.1}}, "evalmod_imag": {{Real: 3.1}},
		"s2c": {{Real: 4}}, "final_public_output": {{Real: 4}},
	}
	standardVectors := probeVectors{Checkpoints: values}
	fastValues := make(map[string][]complexValue, len(values))
	for name, vector := range values {
		fastValues[name] = append([]complexValue(nil), vector...)
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
