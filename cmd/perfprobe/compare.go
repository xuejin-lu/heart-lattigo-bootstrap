package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const defaultComparisonThreshold = 1e-2

type probeVectors struct {
	Original    []complexValue            `json:"original"`
	PreTrials   [][]complexValue          `json:"pre_bootstrap_trials"`
	Trials      [][]complexValue          `json:"bootstrap_trials"`
	Checkpoints map[string][]complexValue `json:"checkpoints"`
}

type vectorMetrics struct {
	ComplexRMSE          float64 `json:"complex_rmse"`
	MaxComplexDifference float64 `json:"max_complex_difference"`
	MaxRealDifference    float64 `json:"max_real_difference"`
	MaxImagDifference    float64 `json:"max_imag_difference"`
	WorstSlot            int     `json:"worst_slot"`
	RealAboveThreshold   int     `json:"real_components_above_threshold"`
	ImagAboveThreshold   int     `json:"imag_components_above_threshold"`
	MedianPrecisionBits  float64 `json:"median_precision_bits"`
}

type trialComparison struct {
	FastTrial      int           `json:"fast_trial"`
	StandardTrial  int           `json:"standard_trial"`
	VsOriginal     vectorMetrics `json:"fast_vs_original"`
	FastVsStandard vectorMetrics `json:"fast_vs_standard"`
}

type stageComparison struct {
	Checkpoint          string               `json:"checkpoint"`
	Parent              string               `json:"parent,omitempty"`
	Fast                ciphertextState      `json:"fast_state"`
	Standard            ciphertextState      `json:"standard_state"`
	ComparableMetadata  bool                 `json:"comparable_metadata"`
	Comparable          bool                 `json:"comparable"`
	NotComparableReason string               `json:"not_comparable_reason,omitempty"`
	FastVsStandard      *vectorMetrics       `json:"fast_vs_standard,omitempty"`
	StageReferenceSNR   numericalmetrics.SNR `json:"stage_reference_snr"`
	Amplification       *float64             `json:"amplification,omitempty"`
	AmplificationStatus string               `json:"amplification_status"`
	DeltaSNRDB          *float64             `json:"delta_snr_db,omitempty"`
	DeltaSNRStatus      string               `json:"delta_snr_status"`
}

type pairedNumericalDocument struct {
	SchemaVersion                 string                 `json:"schema_version"`
	GeneratedAt                   time.Time              `json:"generated_at"`
	Profile                       string                 `json:"profile"`
	MeasurementPrimaryCommit      string                 `json:"measurement_primary_commit"`
	FastBackendCommit             string                 `json:"fast_backend_commit"`
	StandardBackendCommit         string                 `json:"standard_backend_commit"`
	FastBackendPath               string                 `json:"fast_evaluator_path"`
	StandardBackendPath           string                 `json:"standard_evaluator_path"`
	ConfigPath                    string                 `json:"config_path"`
	ConfigSHA256                  string                 `json:"config_sha256"`
	InputSHA256                   string                 `json:"input_sha256"`
	StandardInputMetadataSHA256   string                 `json:"standard_input_metadata_sha256"`
	FastInputMetadataSHA256       string                 `json:"fast_input_metadata_sha256"`
	StandardInputKind             string                 `json:"standard_input_kind"`
	FastInputKind                 string                 `json:"fast_input_kind"`
	Parameters                    effectiveParameters    `json:"effective_parameters"`
	GoVersion                     string                 `json:"go_version"`
	OS                            string                 `json:"os"`
	Arch                          string                 `json:"arch"`
	CPU                           string                 `json:"cpu"`
	MatchedEnvironment            bool                   `json:"matched_environment"`
	MatchedParameters             bool                   `json:"matched_parameters"`
	DescriptiveComponentThreshold float64                `json:"descriptive_component_threshold"`
	FormalEvidenceStatus          string                 `json:"formal_evidence_status"`
	QualityAssessment             string                 `json:"quality_assessment"`
	ReviewStatus                  string                 `json:"review_status"`
	QualityAssessmentReason       string                 `json:"quality_assessment_reason"`
	StandardTiming                timingResult           `json:"standard_full_bootstrap_timing"`
	FastTiming                    timingResult           `json:"fast_full_bootstrap_timing"`
	StandardStageTimings          []stageTiming          `json:"standard_stage_timings"`
	FastStageTimings              []stageTiming          `json:"fast_stage_timings"`
	MedianSpeedup                 float64                `json:"fast_speedup_median"`
	FastOriginalRMSE              []vectorMetrics        `json:"fast_vs_original_trials"`
	StandardOriginalRMSE          []vectorMetrics        `json:"standard_vs_original_trials"`
	FastBootstrapSNR              []numericalmetrics.SNR `json:"fast_bootstrap_snr_trials"`
	StandardBootstrapSNR          []numericalmetrics.SNR `json:"standard_bootstrap_snr_trials"`
	TrialComparisons              []trialComparison      `json:"fast_standard_trial_comparisons"`
	MedianFastStandardRMSE        float64                `json:"median_fast_standard_complex_rmse"`
	MaxFastStandardDifference     float64                `json:"max_fast_standard_complex_difference"`
	MaxRealComponentDifference    float64                `json:"max_fast_standard_real_difference"`
	MaxImagComponentDifference    float64                `json:"max_fast_standard_imag_difference"`
	RealThresholdViolations       int                    `json:"real_threshold_violations"`
	ImagThresholdViolations       int                    `json:"imag_threshold_violations"`
	StageComparisons              []stageComparison      `json:"stage_comparisons"`
	StageCoverage                 string                 `json:"stage_coverage_note"`
	SecurityCaveat                string                 `json:"security_caveat"`
}

func runCompare(args []string) error {
	flags := flag.NewFlagSet("perfprobe compare", flag.ContinueOnError)
	standardPath := flags.String("standard", "", "pinned Standard probe JSON")
	fastPath := flags.String("fast", "", "pinned Fast probe JSON")
	standardVectorsPath := flags.String("standard-vectors", "", "ephemeral Standard vector JSON")
	fastVectorsPath := flags.String("fast-vectors", "", "ephemeral Fast vector JSON")
	outputPath := flags.String("out", "", "paired numerical JSON output")
	reportPath := flags.String("report", "", "human-readable profile report output")
	threshold := flags.Float64("threshold", defaultComparisonThreshold, "descriptive per-component cutoff for metric counts; not a pass/fail gate")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *standardPath == "" || *fastPath == "" || *standardVectorsPath == "" || *fastVectorsPath == "" || *outputPath == "" || *reportPath == "" {
		return errors.New("--standard, --fast, --standard-vectors, --fast-vectors, --out, and --report are required")
	}
	if *threshold <= 0 || math.IsNaN(*threshold) || math.IsInf(*threshold, 0) {
		return errors.New("--threshold must be finite and positive")
	}

	var standard, fast document
	if err := readJSON(*standardPath, &standard); err != nil {
		return fmt.Errorf("read Standard probe: %w", err)
	}
	if err := readJSON(*fastPath, &fast); err != nil {
		return fmt.Errorf("read Fast probe: %w", err)
	}
	var standardVectors, fastVectors probeVectors
	if err := readJSON(*standardVectorsPath, &standardVectors); err != nil {
		return fmt.Errorf("read Standard vectors: %w", err)
	}
	if err := readJSON(*fastVectorsPath, &fastVectors); err != nil {
		return fmt.Errorf("read Fast vectors: %w", err)
	}
	if err := validatePairedInputs(standard, fast, standardVectors, fastVectors); err != nil {
		return err
	}
	comparison, err := buildPairedNumericalDocument(standard, fast, standardVectors, fastVectors, *threshold)
	if err != nil {
		return err
	}
	if field := firstNonFinite(comparison, "comparison"); field != "" {
		return fmt.Errorf("non-finite value at %s", field)
	}
	if err := writeExclusiveJSON(*outputPath, comparison); err != nil {
		return err
	}
	if err := writeExclusiveText(*reportPath, renderPairedReport(comparison)); err != nil {
		return err
	}
	fmt.Printf("%s paired evidence=%s report=%s quality=%s review=%s\n", comparison.Profile, *outputPath, *reportPath, comparison.QualityAssessment, comparison.ReviewStatus)
	return nil
}

func validatePairedInputs(standard, fast document, standardVectors, fastVectors probeVectors) error {
	if standard.SchemaVersion != "fast-standard-perfprobe.v2" || fast.SchemaVersion != "fast-standard-perfprobe.v2" {
		return errors.New("formal comparison requires v2 input provenance; historical/unknown probe records are diagnostic only")
	}
	if standard.Backend != "standard" || fast.Backend != "fast" {
		return fmt.Errorf("backend pair mismatch: Standard=%q Fast=%q", standard.Backend, fast.Backend)
	}
	if err := validatePinnedBackend(standard.Backend, standard.BackendCommit); err != nil {
		return err
	}
	if err := validatePinnedBackend(fast.Backend, fast.BackendCommit); err != nil {
		return err
	}
	if standard.Profile != fast.Profile || standard.PrimaryCommit != fast.PrimaryCommit {
		return fmt.Errorf("profile or Primary harness mismatch: standard=(%s,%s) fast=(%s,%s)", standard.Profile, standard.PrimaryCommit, fast.Profile, fast.PrimaryCommit)
	}
	if standard.PrimaryDirty || fast.PrimaryDirty || standard.SecondaryDirty || fast.SecondaryDirty {
		return errors.New("probe provenance reports a dirty Primary or Secondary worktree")
	}
	if standard.ConfigSHA256 != fast.ConfigSHA256 || standard.ConfigPath != fast.ConfigPath || standard.InputSHA256 != fast.InputSHA256 {
		return errors.New("Standard/Fast config or original input vector fingerprint differs")
	}
	if err := validateFormalInputProvenance(standard, standardNativeInputKind, true); err != nil {
		return fmt.Errorf("Standard input provenance: %w", err)
	}
	if err := validateFormalInputProvenance(fast, fastDirectInputKind, false); err != nil {
		return fmt.Errorf("Fast input provenance: %w", err)
	}
	if !equalEffectiveParameters(standard.Parameters, fast.Parameters) {
		return errors.New("Standard/Fast effective Q/P/profile parameters differ")
	}
	if standard.GoVersion != fast.GoVersion || standard.OS != fast.OS || standard.Arch != fast.Arch || standard.CPU != fast.CPU || standard.NumCPU != fast.NumCPU || standard.GOMAXPROCS != fast.GOMAXPROCS {
		return errors.New("Standard/Fast runtime environment differs")
	}
	if len(standard.Trials) < 3 || len(fast.Trials) < 2 || len(standardVectors.Trials) != len(standard.Trials) || len(fastVectors.Trials) != len(fast.Trials) {
		return fmt.Errorf("insufficient trial/vector counts: Standard trials/vectors=%d/%d Fast=%d/%d", len(standard.Trials), len(standardVectors.Trials), len(fast.Trials), len(fastVectors.Trials))
	}
	if len(standardVectors.PreTrials) != len(standard.Trials) || len(fastVectors.PreTrials) != len(fast.Trials) {
		return errors.New("pre-Bootstrap vector trial counts do not match input provenance")
	}
	if len(standardVectors.Original) == 0 || len(standardVectors.Original) != len(fastVectors.Original) || len(standardVectors.Original) != standard.Parameters.InputSlots {
		return fmt.Errorf("original input vector lengths do not match effective slots: Standard=%d Fast=%d effective=%d", len(standardVectors.Original), len(fastVectors.Original), standard.Parameters.InputSlots)
	}
	if vectorFingerprint(toComplex(standardVectors.Original)) != standard.InputSHA256 || vectorFingerprint(toComplex(fastVectors.Original)) != fast.InputSHA256 {
		return errors.New("vector artifact contents do not match the recorded input SHA-256")
	}
	if !allFinite(toComplex(standardVectors.Original)) || !allFinite(toComplex(fastVectors.Original)) {
		return errors.New("original input vectors contain non-finite values")
	}
	if err := validateBackendTrialVectors(standard, standardVectors); err != nil {
		return fmt.Errorf("Standard trial vectors: %w", err)
	}
	if err := validateBackendTrialVectors(fast, fastVectors); err != nil {
		return fmt.Errorf("Fast trial vectors: %w", err)
	}
	return nil
}

func validateBackendTrialVectors(doc document, vectors probeVectors) error {
	original := toComplex(vectors.Original)
	for i, record := range doc.Trials {
		pre, post := toComplex(vectors.PreTrials[i]), toComplex(vectors.Trials[i])
		if len(pre) != len(original) || len(post) != len(original) || !allFinite(pre) || !allFinite(post) {
			return fmt.Errorf("trial %d input/output vector has invalid length or non-finite values", i+1)
		}
		preHash, postHash := vectorFingerprint(pre), vectorFingerprint(post)
		if preHash != record.PreDecodedSHA256 || preHash != record.InputEvidence.PreDecodedSHA256 {
			return fmt.Errorf("trial %d pre-Bootstrap vector hash contradicts its provenance", i+1)
		}
		if postHash != record.PostDecodedSHA256 || record.OriginalSHA256 != doc.InputSHA256 {
			return fmt.Errorf("trial %d output/original vector hash contradicts its probe record", i+1)
		}
		var maxDeviation float64
		for slot := range original {
			maxDeviation = math.Max(maxDeviation, cmplx.Abs(pre[slot]-original[slot]))
		}
		if maxDeviation > inputQualityLimit || math.Abs(maxDeviation-record.InputEvidence.MaxComplexDeviation) > 1e-15 {
			return fmt.Errorf("trial %d pre-Bootstrap input deviation %.9g contradicts the accepted %.9g input gate/evidence", i+1, maxDeviation, record.InputEvidence.MaxComplexDeviation)
		}
	}
	return nil
}

func validateFormalInputProvenance(doc document, requiredKind string, requireNonzeroC1 bool) error {
	if len(doc.Trials) == 0 || doc.InputEvidence != doc.Trials[0].InputEvidence {
		return errors.New("representative input provenance differs from first trial")
	}
	requiredConstructor := fastDirectConstructor
	if requiredKind == standardNativeInputKind {
		requiredConstructor = standardNativeConstructor
	}
	if doc.InputKind != requiredKind || doc.InputEvidence.Kind != requiredKind || doc.InputEvidence.C1Nonzero != requireNonzeroC1 || doc.InputEvidence.Constructor != requiredConstructor {
		return fmt.Errorf("expected input kind %q with c1_nonzero=%t and a constructor path", requiredKind, requireNonzeroC1)
	}
	if doc.InputEvidence.MetadataSHA256 == "" || doc.InputEvidence.MetadataSHA256 != doc.InputMetadataSHA256 || doc.InputEvidence.State != doc.InputState || doc.InputEvidence.C1SHA256 == "" {
		return errors.New("representative input evidence does not match probe metadata")
	}
	seenC1 := make(map[string]bool, len(doc.Trials))
	for i, trial := range doc.Trials {
		evidence := trial.InputEvidence
		if evidence.Kind != requiredKind || evidence.Constructor != doc.InputEvidence.Constructor || evidence.C1Nonzero != requireNonzeroC1 || evidence.C1SHA256 == "" || evidence.PreDecodedSHA256 != trial.PreDecodedSHA256 {
			return fmt.Errorf("trial %d lacks matching input-origin evidence", i+1)
		}
		if evidence.State.Level != 0 || evidence.State.Degree != 1 || evidence.State.LogCols != doc.Parameters.LogSlots || evidence.MaxComplexDeviation < 0 || evidence.MaxComplexDeviation > inputQualityLimit || evidence.InputQualityLimit != inputQualityLimit {
			return fmt.Errorf("trial %d failed input metadata/quality preflight", i+1)
		}
		if requireNonzeroC1 {
			if !trial.FreshKeyTrial || seenC1[evidence.C1SHA256] {
				return fmt.Errorf("Standard trial %d has no independent-key/c1 evidence", i+1)
			}
			seenC1[evidence.C1SHA256] = true
		}
	}
	return nil
}

func buildPairedNumericalDocument(standard, fast document, standardVectors, fastVectors probeVectors, threshold float64) (pairedNumericalDocument, error) {
	if len(standardVectors.PreTrials) != len(standard.Trials) || len(fastVectors.PreTrials) != len(fast.Trials) {
		return pairedNumericalDocument{}, errors.New("pre-Bootstrap vector trial counts do not match probe records")
	}
	output := pairedNumericalDocument{
		SchemaVersion: "fast-standard-perfprobe-comparison.v3", GeneratedAt: time.Now().UTC(),
		Profile: fast.Profile, MeasurementPrimaryCommit: fast.PrimaryCommit,
		FastBackendCommit: fast.BackendCommit, StandardBackendCommit: standard.BackendCommit,
		FastBackendPath: fast.Trials[0].EvaluatorPath, StandardBackendPath: standard.Trials[0].EvaluatorPath,
		ConfigPath: fast.ConfigPath, ConfigSHA256: fast.ConfigSHA256,
		InputSHA256:                 fast.InputSHA256,
		StandardInputMetadataSHA256: standard.InputMetadataSHA256, FastInputMetadataSHA256: fast.InputMetadataSHA256,
		StandardInputKind: standard.InputKind, FastInputKind: fast.InputKind, Parameters: fast.Parameters,
		GoVersion: fast.GoVersion, OS: fast.OS, Arch: fast.Arch, CPU: fast.CPU,
		MatchedEnvironment: true, MatchedParameters: true, DescriptiveComponentThreshold: threshold,
		FormalEvidenceStatus:    "VALIDATED_V2_PAIRED_PROVENANCE",
		QualityAssessment:       "UNASSESSED",
		ReviewStatus:            "REQUIRES_WEB_REVIEW",
		QualityAssessmentReason: "No accepted formal output-quality pass/fail threshold is defined here; the component cutoff is descriptive only.",
		StandardTiming:          standard.Timing, FastTiming: fast.Timing,
		StandardStageTimings: standard.StageTimings, FastStageTimings: fast.StageTimings,
		MedianSpeedup:  standard.Timing.MedianNS / fast.Timing.MedianNS,
		StageCoverage:  "Public checkpoint comparisons use each probe's own backend-native decoded values and provenance. Internal PS/DoubleAngle stage replay is outside this formal paired artifact.",
		SecurityCaveat: "Standard inputs use genuine RLWE encryption under matching generated secrets. Fast inputs use intentionally insecure direct-encoded zero-secret/zero-a simulation, not native Fast public-key encryption. Input noise/security are not equivalent.",
	}
	output.FastBootstrapSNR = make([]numericalmetrics.SNR, len(fast.Trials))
	output.FastOriginalRMSE = make([]vectorMetrics, len(fast.Trials))
	for i, trial := range fast.Trials {
		pre, post := toComplex(fastVectors.PreTrials[i]), toComplex(fastVectors.Trials[i])
		actualSNR := numericalmetrics.Compare(pre, post)
		if !sameSNR(actualSNR, trial.BootstrapSNR) {
			return pairedNumericalDocument{}, fmt.Errorf("Fast trial %d stored SNR does not match the reusable decoded-domain SNR formula", i+1)
		}
		output.FastBootstrapSNR[i] = actualSNR
		output.FastOriginalRMSE[i] = compareVectors(toComplex(fastVectors.Original), post, threshold)
	}
	output.StandardBootstrapSNR = make([]numericalmetrics.SNR, len(standard.Trials))
	output.StandardOriginalRMSE = make([]vectorMetrics, len(standard.Trials))
	for i, trial := range standard.Trials {
		pre, post := toComplex(standardVectors.PreTrials[i]), toComplex(standardVectors.Trials[i])
		actualSNR := numericalmetrics.Compare(pre, post)
		if !sameSNR(actualSNR, trial.BootstrapSNR) {
			return pairedNumericalDocument{}, fmt.Errorf("Standard trial %d stored SNR does not match the reusable decoded-domain SNR formula", i+1)
		}
		output.StandardBootstrapSNR[i] = actualSNR
		output.StandardOriginalRMSE[i] = compareVectors(toComplex(standardVectors.Original), post, threshold)
	}

	var pairwiseRMSE []float64
	for fastIndex, fastTrial := range fastVectors.Trials {
		fastValues := toComplex(fastTrial)
		for standardIndex, standardTrial := range standardVectors.Trials {
			standardValues := toComplex(standardTrial)
			metric := compareVectors(standardValues, fastValues, threshold)
			output.TrialComparisons = append(output.TrialComparisons, trialComparison{
				FastTrial: fastIndex + 1, StandardTrial: standardIndex + 1,
				VsOriginal:     compareVectors(toComplex(standardVectors.Original), fastValues, threshold),
				FastVsStandard: metric,
			})
			pairwiseRMSE = append(pairwiseRMSE, metric.ComplexRMSE)
			output.MaxFastStandardDifference = math.Max(output.MaxFastStandardDifference, metric.MaxComplexDifference)
			output.MaxRealComponentDifference = math.Max(output.MaxRealComponentDifference, metric.MaxRealDifference)
			output.MaxImagComponentDifference = math.Max(output.MaxImagComponentDifference, metric.MaxImagDifference)
			output.RealThresholdViolations += metric.RealAboveThreshold
			output.ImagThresholdViolations += metric.ImagAboveThreshold
		}
	}
	sort.Float64s(pairwiseRMSE)
	output.MedianFastStandardRMSE = medianSorted(pairwiseRMSE)
	output.StageComparisons = buildStageComparisons(standard, fast, standardVectors, fastVectors, threshold)
	return output, nil
}

func buildStageComparisons(standard, fast document, standardVectors, fastVectors probeVectors, threshold float64) []stageComparison {
	parents := map[string]string{
		"scale_down": "input", "mod_up": "scale_down",
		"c2s_real": "mod_up", "c2s_imag": "mod_up",
		"evalmod_real": "c2s_real", "evalmod_imag": "c2s_imag",
		"final_public_output": "s2c",
	}
	standardStates, fastStates := checkpointStates(standard.Checkpoints), checkpointStates(fast.Checkpoints)
	var out []stageComparison
	errorsByName := make(map[string]vectorMetrics)
	snrsByName := make(map[string]numericalmetrics.SNR)
	var jointError vectorMetrics
	var jointSNR numericalmetrics.SNR
	jointComparable := false
	for _, name := range []string{"input", "scale_down", "mod_up", "c2s_real", "c2s_imag", "evalmod_real", "evalmod_imag", "s2c", "final_public_output"} {
		standardValues, okStandard := standardVectors.Checkpoints[name]
		fastValues, okFast := fastVectors.Checkpoints[name]
		standardState, okStandardState := standardStates[name]
		fastState, okFastState := fastStates[name]
		if !okStandard || !okFast || !okStandardState || !okFastState || !standardState.SemanticallyValid || !fastState.SemanticallyValid {
			continue
		}
		stage := stageComparison{
			Checkpoint: name, Parent: parents[name], Fast: fastState.State, Standard: standardState.State,
			ComparableMetadata:  comparableStageMetadata(fastState.State, standardState.State),
			AmplificationStatus: "NOT_COMPARABLE", DeltaSNRStatus: "NOT_COMPARABLE",
		}
		if fastState.State.QPrefixRows != standardState.State.QPrefixRows {
			stage.NotComparableReason = fmt.Sprintf("Fast maintains %d Q rows while Standard uses %d; decoded values require a common-Q projection", fastState.State.QPrefixRows, standardState.State.QPrefixRows)
			out = append(out, stage)
			continue
		}
		standardDecoded, fastDecoded := toComplex(standardValues), toComplex(fastValues)
		if len(standardDecoded) != len(fastDecoded) || !allFinite(standardDecoded) || !allFinite(fastDecoded) {
			stage.NotComparableReason = "checkpoint vectors are absent, have different lengths, or contain non-finite values"
			out = append(out, stage)
			continue
		}
		metric := compareVectors(standardDecoded, fastDecoded, threshold)
		if field := firstNonFinite(metric, "stage_metric"); field != "" {
			stage.NotComparableReason = "metric overflow at " + field
			out = append(out, stage)
			continue
		}
		snr := numericalmetrics.Compare(standardDecoded, fastDecoded)
		stage.Comparable, stage.FastVsStandard, stage.StageReferenceSNR = true, &metric, snr
		stage.AmplificationStatus, stage.DeltaSNRStatus = "NO_PARENT", "NO_PARENT"
		out = append(out, stage)
		errorsByName[name], snrsByName[name] = metric, snr
		if name == "evalmod_real" {
			jointComparable = true
		}
	}
	if jointComparable && errorsByName["evalmod_imag"].ComplexRMSE >= 0 && snrsByName["evalmod_imag"].Status != "" {
		standardJoined := append(toComplex(standardVectors.Checkpoints["evalmod_real"]), toComplex(standardVectors.Checkpoints["evalmod_imag"])...)
		fastJoined := append(toComplex(fastVectors.Checkpoints["evalmod_real"]), toComplex(fastVectors.Checkpoints["evalmod_imag"])...)
		jointError = compareVectors(standardJoined, fastJoined, threshold)
		jointSNR = numericalmetrics.Compare(standardJoined, fastJoined)
	}
	for i := range out {
		parentName := out[i].Parent
		if out[i].Checkpoint == "s2c" {
			parentName = "evalmod_real+evalmod_imag (joint)"
			out[i].Parent = parentName
			if out[i].Comparable && jointComparable && snrsByName["evalmod_imag"].Status != "" {
				setStageRelations(&out[i], out[i].FastVsStandard.ComplexRMSE, jointError.ComplexRMSE, out[i].StageReferenceSNR, jointSNR)
			} else if out[i].Comparable {
				out[i].AmplificationStatus, out[i].DeltaSNRStatus = "NOT_COMPARABLE_PARENT", "NOT_COMPARABLE_PARENT"
			}
			continue
		}
		if parentName == "" || !out[i].Comparable {
			continue
		}
		parentError, exists := errorsByName[parentName]
		parentSNR, hasSNR := snrsByName[parentName]
		if exists && hasSNR {
			setStageRelations(&out[i], out[i].FastVsStandard.ComplexRMSE, parentError.ComplexRMSE, out[i].StageReferenceSNR, parentSNR)
		}
	}
	return out
}

func setStageRelations(stage *stageComparison, childD, parentD float64, childSNR, parentSNR numericalmetrics.SNR) {
	switch {
	case parentD > 0:
		factor := childD / parentD
		stage.Amplification, stage.AmplificationStatus = &factor, "FINITE"
	case childD == 0:
		stage.AmplificationStatus = "UNDEFINED_ZERO_OVER_ZERO"
	default:
		stage.AmplificationStatus = "UNDEFINED_ZERO_PARENT"
	}
	stage.DeltaSNRDB, stage.DeltaSNRStatus = snrDelta(childSNR, parentSNR)
}

func snrDelta(child, parent numericalmetrics.SNR) (*float64, string) {
	if child.Status == numericalmetrics.Finite && parent.Status == numericalmetrics.Finite && child.SNRDB != nil && parent.SNRDB != nil {
		delta := *child.SNRDB - *parent.SNRDB
		return &delta, "FINITE"
	}
	if child.Status == numericalmetrics.PositiveInfinity && parent.Status == numericalmetrics.PositiveInfinity {
		return nil, "UNDEFINED_INFINITY_MINUS_INFINITY"
	}
	if child.Status == numericalmetrics.Finite && parent.Status == numericalmetrics.PositiveInfinity {
		return nil, "NEGATIVE_INFINITY"
	}
	if child.Status == numericalmetrics.PositiveInfinity && parent.Status == numericalmetrics.Finite {
		return nil, "POSITIVE_INFINITY"
	}
	return nil, "NOT_COMPARABLE"
}

func compareVectors(reference, output []complex128, threshold float64) vectorMetrics {
	if len(reference) == 0 || len(reference) != len(output) {
		return vectorMetrics{}
	}
	precision := make([]float64, len(reference))
	var squaredError float64
	var worstComplex float64
	var maxReal, maxImag float64
	worstSlot := 0
	var realViolations, imagViolations int
	for i, ref := range reference {
		difference := output[i] - ref
		realDifference, imagDifference := math.Abs(real(difference)), math.Abs(imag(difference))
		complexDifference := cmplx.Abs(difference)
		squaredError += complexDifference * complexDifference
		if complexDifference > worstComplex {
			worstComplex, worstSlot = complexDifference, i
		}
		maxReal, maxImag = math.Max(maxReal, realDifference), math.Max(maxImag, imagDifference)
		if realDifference > threshold {
			realViolations++
		}
		if imagDifference > threshold {
			imagViolations++
		}
		precision[i] = -math.Log2(math.Max(complexDifference, 1e-30))
	}
	sort.Float64s(precision)
	return vectorMetrics{
		ComplexRMSE:          math.Sqrt(squaredError / float64(len(reference))),
		MaxComplexDifference: worstComplex, MaxRealDifference: maxReal, MaxImagDifference: maxImag,
		WorstSlot: worstSlot, RealAboveThreshold: realViolations, ImagAboveThreshold: imagViolations,
		MedianPrecisionBits: medianSorted(precision),
	}
}

func allFinite(values []complex128) bool {
	for _, value := range values {
		if math.IsNaN(real(value)) || math.IsNaN(imag(value)) || math.IsInf(real(value), 0) || math.IsInf(imag(value), 0) {
			return false
		}
	}
	return true
}

func medianSorted(sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func checkpointStates(checkpoints []checkpoint) map[string]checkpoint {
	out := make(map[string]checkpoint, len(checkpoints))
	for _, item := range checkpoints {
		out[item.Name] = item
	}
	return out
}

func comparableStageMetadata(fast, standard ciphertextState) bool {
	return fast.Level == standard.Level && fast.Degree == standard.Degree &&
		math.Abs(fast.ScaleLog2-standard.ScaleLog2) <= 1e-9 &&
		fast.LogRows == standard.LogRows && fast.LogCols == standard.LogCols
}

func toComplex(values []complexValue) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func vectorFingerprint(values []complex128) string {
	return perfmeasure.Fingerprint(values)
}

func equalEffectiveParameters(a, b effectiveParameters) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return bytes.Equal(left, right)
}

func sameSNR(a, b numericalmetrics.SNR) bool {
	if a.Status != b.Status || a.Reason != b.Reason {
		return false
	}
	return sameFloatPtr(a.SignalPower, b.SignalPower) && sameFloatPtr(a.NoisePower, b.NoisePower) && sameFloatPtr(a.SNRDB, b.SNRDB)
}

func sameFloatPtr(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return math.Abs(*a-*b) <= 1e-12*math.Max(1, math.Abs(*a))
}

func readJSON(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func firstNonFinite(value any, root string) string {
	return firstNonFiniteValue(reflect.ValueOf(value), root)
}

func firstNonFiniteValue(value reflect.Value, path string) string {
	if !value.IsValid() {
		return ""
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return ""
		}
		return firstNonFiniteValue(value.Elem(), path)
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0) {
			return path
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			if found := firstNonFiniteValue(value.Field(i), path+"."+field.Name); found != "" {
				return found
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if found := firstNonFiniteValue(value.Index(i), fmt.Sprintf("%s[%d]", path, i)); found != "" {
				return found
			}
		}
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			if found := firstNonFiniteValue(iter.Value(), path+"["+fmt.Sprint(iter.Key().Interface())+"]"); found != "" {
				return found
			}
		}
	}
	return ""
}

func writeExclusiveText(path, contents string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(contents); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func renderPairedReport(doc pairedNumericalDocument) string {
	var out strings.Builder
	profileName := profileArtifactName(doc.Profile)
	fmt.Fprintf(&out, "# Fast/Standard perfprobe — %s\n\n", profileName)
	fmt.Fprintf(&out, "- Formal evidence status: `%s`\n- Quality assessment: `%s`\n- Review status: `%s`\n- Measurement Primary SHA: `%s`\n- Fast SHA: `%s`\n- Pinned Standard SHA: `%s`\n- Config: `%s` (`%s`)\n- Original input: `%s`\n- Standard input: `%s`; metadata SHA `%s`\n- Fast input: `%s`; metadata SHA `%s`\n- Effective LogN / LogSlots / slots: %d / %d / %d\n- q0 target / actual prime bits: %d / %d (prime `%s`)\n- Descriptive component cutoff (not a gate): `%.3g`\n- Environment: `%s`, `%s/%s`, CPU `%s`\n\n%s\n\n",
		doc.FormalEvidenceStatus, doc.QualityAssessment, doc.ReviewStatus, doc.MeasurementPrimaryCommit, doc.FastBackendCommit, doc.StandardBackendCommit,
		doc.ConfigPath, doc.ConfigSHA256, doc.InputSHA256,
		doc.StandardInputKind, doc.StandardInputMetadataSHA256, doc.FastInputKind, doc.FastInputMetadataSHA256,
		doc.Parameters.LogN, doc.Parameters.LogSlots, doc.Parameters.InputSlots, doc.Parameters.Q0Target, doc.Parameters.Q0Bits,
		doc.Parameters.QPrimes[0], doc.DescriptiveComponentThreshold, doc.GoVersion, doc.OS, doc.Arch, doc.CPU, doc.QualityAssessmentReason)
	out.WriteString("## Output versus original input — primary quality evidence\n\nThese per-backend measurements compare each public Bootstrap output with the same original workload. The descriptive component cutoff only counts components; it does not classify quality as pass/fail.\n\n| Backend | Trial | Complex RMSE | Max complex diff | Max real diff | Max imag diff | Real / imag components over descriptive cutoff |\n|---|---:|---:|---:|---:|---:|---:|\n")
	for i, metric := range doc.StandardOriginalRMSE {
		fmt.Fprintf(&out, "| Standard | %d | %.9g | %.9g | %.9g | %.9g | %d / %d |\n", i+1, metric.ComplexRMSE, metric.MaxComplexDifference, metric.MaxRealDifference, metric.MaxImagDifference, metric.RealAboveThreshold, metric.ImagAboveThreshold)
	}
	for i, metric := range doc.FastOriginalRMSE {
		fmt.Fprintf(&out, "| Fast | %d | %.9g | %.9g | %.9g | %.9g | %d / %d |\n", i+1, metric.ComplexRMSE, metric.MaxComplexDifference, metric.MaxRealDifference, metric.MaxImagDifference, metric.RealAboveThreshold, metric.ImagAboveThreshold)
	}
	out.WriteString("\n## Bootstrap SNR evidence\n\nSNR uses the reusable `numericalmetrics.Compare(preDecoded, postDecoded)` implementation. Standard trials use fresh native encryptions under independent generated secrets; Fast uses its current direct-encoded zero-secret simulation. These input semantics differ and are recorded above.\n\n| Backend | Trial | Status | SNR dB | Noise RMSE |\n|---|---:|---|---:|---:|\n")
	for i, metric := range doc.StandardBootstrapSNR {
		fmt.Fprintf(&out, "| Standard | %d | %s | %s | %s |\n", i+1, metric.Status, formatOptional(metric.SNRDB), formatOptional(metric.NoiseRMSE))
	}
	for i, metric := range doc.FastBootstrapSNR {
		fmt.Fprintf(&out, "| Fast | %d | %s | %s | %s |\n", i+1, metric.Status, formatOptional(metric.SNRDB), formatOptional(metric.NoiseRMSE))
	}
	fmt.Fprintf(&out, "\n## Full public Bootstrap timing\n\n| Backend | Median ms | Mean ms | Min ms | Max ms | Samples ns | Alloc B/op | Allocs/op |\n|---|---:|---:|---:|---:|---|---:|---:|\n| Standard | %.3f | %.3f | %.3f | %.3f | `%v` | %d | %d |\n| Fast | %.3f | %.3f | %.3f | %.3f | `%v` | %d | %d |\n\nMedian speedup: `%.4fx`.\n",
		doc.StandardTiming.MedianNS/1e6, doc.StandardTiming.MeanNS/1e6, float64(doc.StandardTiming.MinNS)/1e6, float64(doc.StandardTiming.MaxNS)/1e6,
		doc.StandardTiming.SamplesNS, doc.StandardTiming.BytesPerOp, doc.StandardTiming.AllocsPerOp,
		doc.FastTiming.MedianNS/1e6, doc.FastTiming.MeanNS/1e6, float64(doc.FastTiming.MinNS)/1e6, float64(doc.FastTiming.MaxNS)/1e6,
		doc.FastTiming.SamplesNS, doc.FastTiming.BytesPerOp, doc.FastTiming.AllocsPerOp, doc.MedianSpeedup)
	out.WriteString("\nStage timings are separate single-sample public-stage replays and are not summed into the full Bootstrap time.\n\n| Backend | Stage | Elapsed ms | Available | Reason |\n|---|---|---:|---|---|\n")
	for _, stage := range doc.StandardStageTimings {
		fmt.Fprintf(&out, "| Standard | %s | %.3f | %t | %s |\n", stage.Stage, float64(stage.ElapsedNS)/1e6, stage.Available, stage.Reason)
	}
	for _, stage := range doc.FastStageTimings {
		fmt.Fprintf(&out, "| Fast | %s | %.3f | %t | %s |\n", stage.Stage, float64(stage.ElapsedNS)/1e6, stage.Available, stage.Reason)
	}
	fmt.Fprintf(&out, "\n## Fast-versus-Standard pairwise comparison — secondary evidence\n\nFormal quality remains `%s`; pairwise differences are descriptive and require independent review. Median Fast-vs-Standard complex RMSE across %d trial pairs: `%.9g`; maximum complex difference: `%.9g`; max real / imag component difference: `%.9g` / `%.9g`; components over descriptive cutoff: %d / %d.\n\n| Fast trial | Standard trial | Pairwise complex RMSE | Max complex diff | Median precision bits | Fast-vs-original RMSE |\n|---:|---:|---:|---:|---:|---:|\n",
		doc.QualityAssessment, len(doc.TrialComparisons), doc.MedianFastStandardRMSE, doc.MaxFastStandardDifference,
		doc.MaxRealComponentDifference, doc.MaxImagComponentDifference, doc.RealThresholdViolations, doc.ImagThresholdViolations)
	for _, pair := range doc.TrialComparisons {
		fmt.Fprintf(&out, "| %d | %d | %.9g | %.9g | %.4f | %.9g |\n", pair.FastTrial, pair.StandardTrial,
			pair.FastVsStandard.ComplexRMSE, pair.FastVsStandard.MaxComplexDifference,
			pair.FastVsStandard.MedianPrecisionBits, pair.VsOriginal.ComplexRMSE)
	}
	out.WriteString("\n## Topology-aware public stage comparisons\n\n`C2S real` and `C2S imag` are parallel children of ModUp. EvalMod real/imag compare only with their matching C2S branch. S2C amplification uses the joint real+imag EvalMod error representation; final public output compares with S2C.\n\n| Checkpoint | L / scale (Fast; Standard) | Fast-vs-Standard RMSE | Max complex diff | Stage-reference SNR | Parent | Aᵢ | ΔSNRᵢ | Metadata match |\n|---|---|---:|---:|---:|---|---:|---:|---|\n")
	for _, stage := range doc.StageComparisons {
		metricText, maxText := "n/a", "n/a"
		if stage.FastVsStandard != nil {
			metricText = fmt.Sprintf("%.9g", stage.FastVsStandard.ComplexRMSE)
			maxText = fmt.Sprintf("%.9g", stage.FastVsStandard.MaxComplexDifference)
		}
		fmt.Fprintf(&out, "| %s | %d / %.4f; %d / %.4f | %s | %s | %s | %s | %s | %s | %t |\n",
			stage.Checkpoint, stage.Fast.Level, stage.Fast.ScaleLog2, stage.Standard.Level, stage.Standard.ScaleLog2,
			metricText, maxText, formatSNR(stage.StageReferenceSNR),
			stage.Parent, formatOptional(stage.Amplification), formatOptional(stage.DeltaSNRDB), stage.ComparableMetadata)
		if stage.NotComparableReason != "" {
			fmt.Fprintf(&out, "\n  Not comparable: %s\n", stage.NotComparableReason)
		}
	}
	fmt.Fprintf(&out, "\n## Evidence scope and interpretation limits\n\n%s\n\n%s\n", doc.StageCoverage, doc.SecurityCaveat)
	return out.String()
}

func profileArtifactName(profile string) string {
	if strings.HasPrefix(profile, "logn") {
		return "LogN" + strings.TrimPrefix(profile, "logn")
	}
	return profile
}

func formatOptional(value *float64) string {
	if value == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.9g", *value)
}

func formatSNR(value numericalmetrics.SNR) string {
	if value.SNRDB != nil {
		return fmt.Sprintf("%.6f dB / %s", *value.SNRDB, value.Status)
	}
	if value.Status == "" {
		return "NOT_COMPARABLE"
	}
	return string(value.Status)
}
