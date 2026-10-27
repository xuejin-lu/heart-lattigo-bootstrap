package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

type publicTripletErrorMetrics struct {
	Stats               numericalmetrics.ComplexVectorStats    `json:"complex_error_stats"`
	Tails               numericalmetrics.ComplexMagnitudeTails `json:"error_magnitude_tails"`
	RMSE                float64                                `json:"rmse"`
	PrecisionBits       *float64                               `json:"precision_bits,omitempty"`
	PrecisionBitsStatus string                                 `json:"precision_bits_status"`
}

type publicTripletArm struct {
	Backend              string                    `json:"backend"`
	BackendCommit        string                    `json:"backend_commit"`
	BinarySHA256         string                    `json:"binary_sha256"`
	BootstrapCalls       int                       `json:"bootstrap_calls"`
	SuccessfulSlots      int                       `json:"successful_slots"`
	BootstrapInputSHA256 string                    `json:"bootstrap_plaintext_input_sha256"`
	OutputSHA256         string                    `json:"output_sha256"`
	Error                publicTripletErrorMetrics `json:"error_vs_same_original_input"`
	NumericalGateMaxAbs  float64                   `json:"numerical_gate_max_abs_observed"`
	NumericalGate        float64                   `json:"numerical_gate_max_abs_threshold"`
	NumericalGatePass    bool                      `json:"numerical_gate_pass"`
	BootstrapTimingPhase string                    `json:"bootstrap_timing_phase"`
	BootstrapTimeNS      int64                     `json:"bootstrap_time_ns"`
	AllocBytes           uint64                    `json:"bootstrap_alloc_bytes"`
	Allocs               uint64                    `json:"bootstrap_allocs"`
	PeakMemoryStatus     string                    `json:"peak_memory_status"`
	CtCtMultiplyCalls    uint64                    `json:"ct_ct_multiply_calls"`
	InjectedPolynomials  uint64                    `json:"injected_polynomials"`
	SigmaCoefficient     float64                   `json:"sigma_coefficient,omitempty"`
}

type publicTripletPairwise struct {
	Left  string                                 `json:"left"`
	Right string                                 `json:"right"`
	RMSE  float64                                `json:"output_pairwise_rmse"`
	Stats numericalmetrics.ComplexVectorStats    `json:"output_pairwise_error_stats"`
	Tails numericalmetrics.ComplexMagnitudeTails `json:"output_pairwise_error_tails"`
}

type publicTripletDocument struct {
	SchemaVersion                          string                      `json:"schema_version"`
	Status                                 string                      `json:"status"`
	GeneratedAt                            time.Time                   `json:"generated_at"`
	Profile                                string                      `json:"profile"`
	PrimaryCommit                          string                      `json:"primary_commit"`
	PrimarySourceSHA256                    string                      `json:"primary_measurement_source_sha256"`
	StandardCommit                         string                      `json:"standard_commit"`
	FastCommit                             string                      `json:"fast_commit"`
	FastBinarySHA256                       string                      `json:"shared_fast_binary_sha256"`
	InputSHA256                            string                      `json:"canonical_workload_input_sha256"`
	BootstrapInputSHA256                   string                      `json:"bootstrap_plaintext_input_sha256"`
	ConfigSHA256                           string                      `json:"config_sha256"`
	QPSHA256                               string                      `json:"qp_sha256"`
	WorkloadSHA256                         string                      `json:"workload_sha256"`
	PreflightPairSHA256                    string                      `json:"preflight_pair_sha256"`
	SigmaCoefficient                       float64                     `json:"fixed_sigma_coefficient,omitempty"`
	NoiseRootSeed                          *int64                      `json:"noise_root_seed,omitempty"`
	Arms                                   map[string]publicTripletArm `json:"arms"`
	Pairwise                               []publicTripletPairwise     `json:"pairwise_output_errors"`
	FastPrecisionDistanceFromStandard      *float64                    `json:"fast_abs_precision_bit_distance_from_standard,omitempty"`
	FastNoisePrecisionDistanceFromStandard *float64                    `json:"fast_noise_abs_precision_bit_distance_from_standard,omitempty"`
	FastNoisePrecisionDeltaFromStandard    *float64                    `json:"fast_noise_precision_delta_from_standard,omitempty"`
	PrecisionDistanceImprovement           *float64                    `json:"absolute_precision_distance_improvement,omitempty"`
	PrecisionComparisonStatus              string                      `json:"precision_comparison_status"`
	FastSpeedupVsStandard                  float64                     `json:"fast_speedup_vs_standard"`
	FastNoiseSpeedupVsStandard             float64                     `json:"fast_noise_speedup_vs_standard"`
	FastNoiseTimeOverheadNS                int64                       `json:"fast_noise_time_overhead_ns_vs_fast"`
	FastNoiseTimeOverheadRatio             float64                     `json:"fast_noise_time_ratio_vs_fast"`
	MatchedHost                            bool                        `json:"matched_host"`
	Interpretation                         string                      `json:"interpretation"`
}

type publicTriplet002Sample struct {
	Name                  string                    `json:"sample_name"`
	Phase                 string                    `json:"phase"`
	InputCiphertextSHA256 string                    `json:"input_ciphertext_sha256"`
	OutputSHA256          string                    `json:"output_sha256"`
	Error                 publicTripletErrorMetrics `json:"error_vs_same_original_input"`
	MaxAbs                float64                   `json:"max_abs"`
	P99                   float64                   `json:"p99"`
	P999                  float64                   `json:"p99_9"`
	NumericalGate         float64                   `json:"observational_max_abs_gate"`
	NumericalGatePass     bool                      `json:"observational_gate_pass"`
	ElapsedNS             int64                     `json:"elapsed_ns"`
	AllocBytes            uint64                    `json:"allocated_bytes"`
	Allocs                uint64                    `json:"allocations"`
}

type publicTriplet002FloatSummary struct {
	Count        int     `json:"count"`
	Mean         float64 `json:"mean"`
	Median       float64 `json:"median"`
	SampleStdDev float64 `json:"sample_stddev"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
}

type publicTriplet002Arm struct {
	Backend                        string                       `json:"backend"`
	BackendCommit                  string                       `json:"backend_commit"`
	PrimaryEvidenceCommit          string                       `json:"primary_evidence_commit"`
	BinarySHA256                   string                       `json:"binary_sha256"`
	Samples                        []publicTriplet002Sample     `json:"all_six_samples"`
	Cold                           publicTriplet002Sample       `json:"cold_sample"`
	Warm                           []publicTriplet002Sample     `json:"five_warm_samples"`
	WarmElapsedNS                  publicTriplet002FloatSummary `json:"warm_elapsed_ns_summary"`
	WarmAllocBytes                 publicTriplet002FloatSummary `json:"warm_allocated_bytes_summary"`
	WarmAllocs                     publicTriplet002FloatSummary `json:"warm_allocations_summary"`
	WarmRMSE                       publicTriplet002FloatSummary `json:"warm_rmse_summary"`
	WarmPrecisionBits              publicTriplet002FloatSummary `json:"warm_global_precision_bits_summary"`
	WarmMaxAbs                     publicTriplet002FloatSummary `json:"warm_max_abs_summary"`
	WarmP99                        publicTriplet002FloatSummary `json:"warm_p99_summary"`
	WarmP999                       publicTriplet002FloatSummary `json:"warm_p99_9_summary"`
	EvaluatorConstructionPhases    []phaseTiming                `json:"evaluator_construction_phases"`
	NoiseHookConfiguration         *phaseTiming                 `json:"noise_hook_configuration_phase,omitempty"`
	EvaluatorConstructionElapsedNS int64                        `json:"evaluator_construction_elapsed_ns"`
	EvaluatorConstructionAllocByte uint64                       `json:"evaluator_construction_allocated_bytes"`
	EvaluatorConstructionAllocs    uint64                       `json:"evaluator_construction_allocations"`
	ConstructionPlusColdElapsedNS  int64                        `json:"construction_plus_cold_elapsed_ns"`
	ConstructionPlusColdAllocBytes uint64                       `json:"construction_plus_cold_allocated_bytes"`
	ConstructionPlusColdAllocs     uint64                       `json:"construction_plus_cold_allocations"`
}

type publicTriplet002PrecisionComparison struct {
	Phase                          string   `json:"phase"`
	StandardBits                   *float64 `json:"standard_global_precision_bits,omitempty"`
	FastBits                       *float64 `json:"fast_global_precision_bits,omitempty"`
	FastNoiseBits                  *float64 `json:"fast_noise_global_precision_bits,omitempty"`
	FastSignedBitsGapFromStandard  *float64 `json:"fast_signed_bits_gap_from_standard,omitempty"`
	NoiseSignedBitsGapFromStandard *float64 `json:"fast_noise_signed_bits_gap_from_standard,omitempty"`
	FastDistanceFromStandard       *float64 `json:"fast_abs_bits_distance_from_standard,omitempty"`
	NoiseDistanceFromStandard      *float64 `json:"fast_noise_abs_bits_distance_from_standard,omitempty"`
	NoiseDistanceImprovement       *float64 `json:"noise_distance_improvement_bits,omitempty"`
	NoiseReducedAbsoluteGap        *bool    `json:"noise_reduced_absolute_gap,omitempty"`
}

type publicTriplet002PairwiseOutput struct {
	Phase               string                `json:"phase"`
	StandardVsFast      publicTripletPairwise `json:"standard_vs_fast"`
	StandardVsFastNoise publicTripletPairwise `json:"standard_vs_fast_noise"`
	FastVsFastNoise     publicTripletPairwise `json:"fast_vs_fast_noise"`
}

type publicTriplet002NoiseEvidence struct {
	Enabled                  bool                          `json:"enabled"`
	Sigma                    float64                       `json:"sigma_coefficient"`
	RootSeed                 int64                         `json:"root_seed"`
	CtCtMultiplyCalls        uint64                        `json:"ct_ct_multiply_calls"`
	InjectedPolynomials      uint64                        `json:"injected_polynomials"`
	Events                   []bootstrapNoiseEventEvidence `json:"events"`
	InjectionSubphaseElapsed map[string]int64              `json:"injection_subphase_elapsed_ns"`
	TotalInjectionSubphaseNS int64                         `json:"total_injection_subphase_elapsed_ns"`
	PerInjectionAllocation   string                        `json:"per_injection_allocation_measurement"`
}

type publicTripletFidelity002Document struct {
	SchemaVersion                     string                                `json:"schema_version"`
	TaskID                            string                                `json:"task_id"`
	Status                            string                                `json:"status"`
	GeneratedAt                       time.Time                             `json:"generated_at"`
	Profile                           string                                `json:"profile"`
	PrimaryCommit                     string                                `json:"primary_commit"`
	PrimarySourceSHA256               string                                `json:"primary_measurement_source_sha256"`
	StandardCommit                    string                                `json:"standard_commit"`
	FastCommit                        string                                `json:"fast_commit"`
	FastBinarySHA256                  string                                `json:"shared_fast_binary_sha256"`
	InputSHA256                       string                                `json:"canonical_workload_input_sha256"`
	BootstrapInputSHA256              string                                `json:"bootstrap_plaintext_input_sha256"`
	ConfigSHA256                      string                                `json:"config_sha256"`
	QPSHA256                          string                                `json:"qp_sha256"`
	WorkloadSHA256                    string                                `json:"workload_sha256"`
	PreflightPairSHA256               string                                `json:"preflight_pair_sha256"`
	SigmaCoefficient                  float64                               `json:"fixed_sigma_coefficient"`
	NoiseRootSeed                     int64                                 `json:"noise_root_seed"`
	NumericalGateMaxAbs               float64                               `json:"observational_numerical_gate_max_abs"`
	Arms                              map[string]publicTriplet002Arm        `json:"arms"`
	PrecisionComparisonByPhase        []publicTriplet002PrecisionComparison `json:"global_precision_comparison_by_cold_warm_phase"`
	PairwiseOutputsByPhase            []publicTriplet002PairwiseOutput      `json:"pairwise_output_errors_by_cold_warm_phase"`
	NoiseImprovedPrecisionSamples     int                                   `json:"noise_improved_precision_sample_count"`
	NoiseWorsenedPrecisionSamples     int                                   `json:"noise_worsened_precision_sample_count"`
	NoiseEqualPrecisionSamples        int                                   `json:"noise_equal_precision_sample_count"`
	FastColdSpeedupVsStandard         float64                               `json:"fast_cold_speedup_vs_standard"`
	FastWarmMedianSpeedupVsStandard   float64                               `json:"fast_warm_median_speedup_vs_standard"`
	NoiseColdSpeedupVsStandard        float64                               `json:"fast_noise_cold_speedup_vs_standard"`
	NoiseWarmMedianSpeedupVsStandard  float64                               `json:"fast_noise_warm_median_speedup_vs_standard"`
	NoiseColdTimeOverheadNS           int64                                 `json:"fast_noise_cold_time_overhead_ns_vs_fast"`
	NoiseColdTimeOverheadRatio        float64                               `json:"fast_noise_cold_time_ratio_vs_fast"`
	NoiseWarmMedianOverheadNS         float64                               `json:"fast_noise_warm_median_time_overhead_ns_vs_fast"`
	NoiseWarmMedianOverheadRatio      float64                               `json:"fast_noise_warm_median_time_ratio_vs_fast"`
	NoiseColdAllocationOverheadBytes  int64                                 `json:"fast_noise_cold_allocation_overhead_bytes_vs_fast"`
	NoiseColdAllocationOverheadAllocs int64                                 `json:"fast_noise_cold_allocation_count_overhead_vs_fast"`
	NoiseWarmMedianAllocOverheadBytes float64                               `json:"fast_noise_warm_median_allocated_bytes_overhead_vs_fast"`
	NoiseWarmMedianAllocsOverhead     float64                               `json:"fast_noise_warm_median_allocation_count_overhead_vs_fast"`
	NoiseEvidence                     publicTriplet002NoiseEvidence         `json:"noise_evidence"`
	MatchedHost                       bool                                  `json:"matched_host"`
	Interpretation                    string                                `json:"interpretation"`
}

func runPublicTripletCompare(args []string) error {
	flags := flag.NewFlagSet("perfprobe compare-public-triplet", flag.ContinueOnError)
	standardPath := flags.String("standard", "", "single-call genuine Standard Bootstrap JSON")
	standardVectorsPath := flags.String("standard-vectors", "", "single-call genuine Standard decoded vectors JSON")
	fastPath := flags.String("fast", "", "single-call Fast Bootstrap with the hook disabled JSON")
	fastVectorsPath := flags.String("fast-vectors", "", "single-call Fast hook-disabled decoded vectors JSON")
	fastNoisePath := flags.String("fast-noise", "", "single-call Fast Bootstrap with fixed-sigma hook enabled JSON")
	fastNoiseVectorsPath := flags.String("fast-noise-vectors", "", "single-call Fast fixed-sigma decoded vectors JSON")
	preflightPairPath := flags.String("preflight-pair", "", "fresh zero-call task-bound Standard/Fast pair required by Noise Fidelity comparison")
	outPath := flags.String("out", "", "exclusive compact triplet result JSON")
	noiseFidelity002 := flags.Bool("noise-fidelity-002", false, "compare the task-specific 1-cold/5-warm captures from FAST-BOOTSTRAP-NOISE-FIDELITY-002")
	noiseFidelity003 := flags.Bool("noise-fidelity-003-logn16", false, "compare the task-specific 1-cold/5-warm captures from FAST-BOOTSTRAP-NOISE-FIDELITY-003-LOGN16")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *standardPath == "" || *standardVectorsPath == "" || *fastPath == "" || *fastVectorsPath == "" ||
		*fastNoisePath == "" || *fastNoiseVectorsPath == "" || *outPath == "" {
		return errors.New("--standard, --standard-vectors, --fast, --fast-vectors, --fast-noise, --fast-noise-vectors, and --out are required")
	}
	var standard, fast, fastNoise publicNativeDocument
	var standardVectors, fastVectors, fastNoiseVectors publicVectorDocument
	for _, item := range []struct {
		path string
		dst  any
	}{
		{*standardPath, &standard}, {*standardVectorsPath, &standardVectors},
		{*fastPath, &fast}, {*fastVectorsPath, &fastVectors},
		{*fastNoisePath, &fastNoise}, {*fastNoiseVectorsPath, &fastNoiseVectors},
	} {
		if err := readJSON(item.path, item.dst); err != nil {
			return fmt.Errorf("read triplet artifact %s: %w", item.path, err)
		}
	}
	if *noiseFidelity002 && *noiseFidelity003 {
		return errors.New("Noise Fidelity 002 and 003 triplet contracts are mutually exclusive")
	}
	if *noiseFidelity002 || *noiseFidelity003 {
		if *preflightPairPath == "" {
			return errors.New("task-scoped Noise Fidelity comparison requires --preflight-pair")
		}
		taskID := publicNoiseFidelity002TaskID
		if *noiseFidelity003 {
			taskID = publicNoiseFidelity003TaskID
		}
		return runPublicTripletCompareTask(standard, standardVectors, fast, fastVectors, fastNoise, fastNoiseVectors, *preflightPairPath, *outPath, taskID)
	}
	if *preflightPairPath != "" {
		return errors.New("--preflight-pair is reserved for a task-scoped Noise Fidelity comparison")
	}
	if err := validatePublicBootstrapTripletArm(standard, standardVectors, "standard", false); err != nil {
		return fmt.Errorf("Standard arm: %w", err)
	}
	if err := validatePublicBootstrapTripletArm(fast, fastVectors, "fast", false); err != nil {
		return fmt.Errorf("Fast arm: %w", err)
	}
	if err := validatePublicBootstrapTripletArm(fastNoise, fastNoiseVectors, "fast", true); err != nil {
		return fmt.Errorf("Fast+Noise arm: %w", err)
	}
	if err := validatePublicBootstrapTripletCompatibility(standard, fast, fastNoise, standardVectors, fastVectors, fastNoiseVectors); err != nil {
		return err
	}
	bootstrapInputs, err := publicOracleCheckpoints(standard.InputSlots)
	if err != nil {
		return err
	}
	plainInput := bootstrapInputs["drop_level0"]
	standardOutput := decodeComplexValues(standardVectors.BootstrapOutputs["bootstrap_once"])
	fastOutput := decodeComplexValues(fastVectors.BootstrapOutputs["bootstrap_once"])
	fastNoiseOutput := decodeComplexValues(fastNoiseVectors.BootstrapOutputs["bootstrap_once"])
	standardArm, err := summarizePublicTripletArm(standard, standardOutput, plainInput)
	if err != nil {
		return err
	}
	fastArm, err := summarizePublicTripletArm(fast, fastOutput, plainInput)
	if err != nil {
		return err
	}
	fastNoiseArm, err := summarizePublicTripletArm(fastNoise, fastNoiseOutput, plainInput)
	if err != nil {
		return err
	}
	var hookSnapshot bootstrapNoiseSnapshotEvidence
	if err := json.Unmarshal(fastNoise.BootstrapNoiseEvidence, &hookSnapshot); err != nil {
		return fmt.Errorf("decode Fast+Noise summary evidence: %w", err)
	}
	fastNoiseArm.CtCtMultiplyCalls = hookSnapshot.CtCtMultiplyCalls
	fastNoiseArm.InjectedPolynomials = hookSnapshot.InjectedPolynomials
	fastNoiseArm.SigmaCoefficient = hookSnapshot.Sigma
	precisionStatus := "FINITE"
	var fastDistance, fastNoiseDistance, noiseDelta, improvement *float64
	if standardArm.Error.PrecisionBits == nil || fastArm.Error.PrecisionBits == nil || fastNoiseArm.Error.PrecisionBits == nil {
		precisionStatus = "UNDEFINED_ZERO_OR_NONFINITE_RMSE_IN_AT_LEAST_ONE_ARM"
	} else {
		standardBits, fastBits, fastNoiseBits := *standardArm.Error.PrecisionBits, *fastArm.Error.PrecisionBits, *fastNoiseArm.Error.PrecisionBits
		fastDistanceValue, fastNoiseDistanceValue := math.Abs(fastBits-standardBits), math.Abs(fastNoiseBits-standardBits)
		noiseDeltaValue, improvementValue := fastNoiseBits-standardBits, fastDistanceValue-fastNoiseDistanceValue
		fastDistance, fastNoiseDistance = &fastDistanceValue, &fastNoiseDistanceValue
		noiseDelta, improvement = &noiseDeltaValue, &improvementValue
	}
	if standardArm.BootstrapTimeNS <= 0 || fastArm.BootstrapTimeNS <= 0 || fastNoiseArm.BootstrapTimeNS <= 0 {
		return errors.New("triplet timing ratios require three positive single-call timings")
	}
	rootSeed := bootstrapNoiseRootSeed(fastNoise.BootstrapNoiseEvidence)
	doc := publicTripletDocument{
		SchemaVersion: "fast-bootstrap-noise-fidelity-triplet.v2", Status: "TRIPLET_COMPLETE_EXPLORATORY", GeneratedAt: time.Now().UTC(),
		Profile: standard.Profile, PrimaryCommit: standard.PrimaryCommit, PrimarySourceSHA256: standard.PrimarySourceSHA256,
		StandardCommit: standard.BackendCommit, FastCommit: fast.BackendCommit, FastBinarySHA256: fast.BinarySHA256,
		InputSHA256: standard.InputSHA256, BootstrapInputSHA256: standard.BootstrapInputSHA256,
		ConfigSHA256: standard.ConfigSHA256, QPSHA256: standard.QPSHA256, WorkloadSHA256: standard.WorkloadSHA256,
		PreflightPairSHA256: standard.PreflightPairSHA256, SigmaCoefficient: publicBootstrapNoiseSigma, NoiseRootSeed: &rootSeed,
		Arms: map[string]publicTripletArm{"S": standardArm, "F": fastArm, "FN": fastNoiseArm},
		Pairwise: []publicTripletPairwise{
			makePublicTripletPairwise("S", "F", standardOutput, fastOutput),
			makePublicTripletPairwise("S", "FN", standardOutput, fastNoiseOutput),
			makePublicTripletPairwise("F", "FN", fastOutput, fastNoiseOutput),
		},
		FastPrecisionDistanceFromStandard: fastDistance, FastNoisePrecisionDistanceFromStandard: fastNoiseDistance,
		FastNoisePrecisionDeltaFromStandard: noiseDelta, PrecisionDistanceImprovement: improvement,
		PrecisionComparisonStatus:  precisionStatus,
		FastSpeedupVsStandard:      float64(standardArm.BootstrapTimeNS) / float64(fastArm.BootstrapTimeNS),
		FastNoiseSpeedupVsStandard: float64(standardArm.BootstrapTimeNS) / float64(fastNoiseArm.BootstrapTimeNS),
		FastNoiseTimeOverheadNS:    fastNoiseArm.BootstrapTimeNS - fastArm.BootstrapTimeNS,
		FastNoiseTimeOverheadRatio: float64(fastNoiseArm.BootstrapTimeNS) / float64(fastArm.BootstrapTimeNS),
		MatchedHost:                publicEnvironmentsMatch(standard, fast) && publicEnvironmentsMatch(standard, fastNoise),
		Interpretation:             "Each arm's precision is computed against the same canonical plaintext supplied to Bootstrap; pairwise output RMSE is a separate comparison. NumericalGatePass reports the observational MaxAbs gate per arm and does not reject finite structurally valid one-call outputs. A complete triplet is exploratory evidence, not proof of general noise fidelity.",
	}
	if err := writeExclusiveJSON(*outPath, doc); err != nil {
		return err
	}
	return nil
}

func runPublicTripletCompare002(
	standard publicNativeDocument, standardVectors publicVectorDocument,
	fast publicNativeDocument, fastVectors publicVectorDocument,
	fastNoise publicNativeDocument, fastNoiseVectors publicVectorDocument,
	preflightPairPath, outPath string,
) error {
	return runPublicTripletCompareTask(standard, standardVectors, fast, fastVectors, fastNoise, fastNoiseVectors, preflightPairPath, outPath, publicNoiseFidelity002TaskID)
}

func runPublicTripletCompareTask(
	standard publicNativeDocument, standardVectors publicVectorDocument,
	fast publicNativeDocument, fastVectors publicVectorDocument,
	fastNoise publicNativeDocument, fastNoiseVectors publicVectorDocument,
	preflightPairPath, outPath, taskID string,
) error {
	for label, pair := range []struct {
		standard publicNativeDocument
		fast     publicNativeDocument
		sv       publicVectorDocument
		fv       publicVectorDocument
	}{
		{standard, fast, standardVectors, fastVectors},
		{standard, fastNoise, standardVectors, fastNoiseVectors},
	} {
		if err := validatePublicBootstrapPairArtifacts(pair.standard, pair.fast, pair.sv, pair.fv); err != nil {
			return fmt.Errorf("%s pair %d: %w", taskID, label+1, err)
		}
	}
	if err := validatePublicTripletCompatibilityTask(standard, fast, fastNoise, standardVectors, fastVectors, fastNoiseVectors, taskID); err != nil {
		return err
	}
	if err := validateNoiseFidelityTripletPreflight(preflightPairPath, standard, fast, fastNoise, taskID); err != nil {
		return err
	}
	oracles, err := publicOracleCheckpoints(standard.InputSlots)
	if err != nil {
		return err
	}
	input := decodeComplexValues(standardVectors.BootstrapPlaintextInput)
	if len(input) != standard.InputSlots || !allFinite(input) || perfmeasure.Fingerprint(input) != standard.BootstrapInputSHA256 {
		return fmt.Errorf("%s comparator requires the canonical finite Bootstrap plaintext", taskID)
	}
	canonical := decodeComplexValues(encodeVector(oracles["drop_level0"]))
	if !reflect.DeepEqual(input, canonical) {
		return fmt.Errorf("%s Bootstrap plaintext differs from the frozen original workload input", taskID)
	}
	standardArm, err := summarizePublicTriplet002Arm(standard, standardVectors, input)
	if err != nil {
		return fmt.Errorf("Standard arm: %w", err)
	}
	fastArm, err := summarizePublicTriplet002Arm(fast, fastVectors, input)
	if err != nil {
		return fmt.Errorf("Fast arm: %w", err)
	}
	noiseArm, err := summarizePublicTriplet002Arm(fastNoise, fastNoiseVectors, input)
	if err != nil {
		return fmt.Errorf("Fast+Noise arm: %w", err)
	}
	noiseEvidence, err := summarizePublicTripletTaskNoise(fastNoise, taskID)
	if err != nil {
		return err
	}
	if standardArm.Samples[0].ElapsedNS <= 0 || fastArm.Samples[0].ElapsedNS <= 0 || noiseArm.Samples[0].ElapsedNS <= 0 {
		return fmt.Errorf("%s cold timing ratios require three positive measured calls", taskID)
	}
	comparison := make([]publicTriplet002PrecisionComparison, 0, 6)
	pairwise := make([]publicTriplet002PairwiseOutput, 0, 6)
	noiseImproved, noiseWorsened, noiseEqual := 0, 0, 0
	for index, phase := range []string{"cold", "warm_01", "warm_02", "warm_03", "warm_04", "warm_05"} {
		s, f, fn := standardArm.Samples[index].Error.PrecisionBits, fastArm.Samples[index].Error.PrecisionBits, noiseArm.Samples[index].Error.PrecisionBits
		item := publicTriplet002PrecisionComparison{Phase: phase, StandardBits: s, FastBits: f, FastNoiseBits: fn}
		if s != nil && f != nil && fn != nil {
			fs, fns := *f-*s, *fn-*s
			fd, nd := math.Abs(fs), math.Abs(fns)
			improvement := fd - nd
			reduced := improvement > 0
			item.FastSignedBitsGapFromStandard, item.NoiseSignedBitsGapFromStandard = &fs, &fns
			item.FastDistanceFromStandard, item.NoiseDistanceFromStandard, item.NoiseDistanceImprovement = &fd, &nd, &improvement
			item.NoiseReducedAbsoluteGap = &reduced
			switch {
			case improvement > 0:
				noiseImproved++
			case improvement < 0:
				noiseWorsened++
			default:
				noiseEqual++
			}
		}
		comparison = append(comparison, item)
		sv := decodeComplexValues(standardVectors.BootstrapOutputs[standardArm.Samples[index].Name])
		fv := decodeComplexValues(fastVectors.BootstrapOutputs[fastArm.Samples[index].Name])
		fnv := decodeComplexValues(fastNoiseVectors.BootstrapOutputs[noiseArm.Samples[index].Name])
		pairwise = append(pairwise, publicTriplet002PairwiseOutput{
			Phase: phase, StandardVsFast: makePublicTripletPairwise("S", "F", sv, fv),
			StandardVsFastNoise: makePublicTripletPairwise("S", "FN", sv, fnv),
			FastVsFastNoise:     makePublicTripletPairwise("F", "FN", fv, fnv),
		})
	}
	if noiseArm.NoiseHookConfiguration == nil || !noiseArm.NoiseHookConfiguration.Available || noiseArm.NoiseHookConfiguration.ElapsedNS <= 0 {
		return errors.New("Fast+Noise arm is missing its separately measured hook-configuration phase")
	}
	fastWarmMedian := fastArm.WarmElapsedNS.Median
	noiseWarmMedian := noiseArm.WarmElapsedNS.Median
	if standardArm.WarmElapsedNS.Median <= 0 || fastWarmMedian <= 0 || noiseWarmMedian <= 0 {
		return fmt.Errorf("%s warm speedup ratios require positive warm medians", taskID)
	}
	profileName := "logn13"
	fastPin := publicNoiseFidelity002FastSHA
	rootSeed := publicNoiseFidelity002RootSeed
	schemaSuffix := "002"
	if taskID == publicNoiseFidelity003TaskID {
		profileName, fastPin, rootSeed = "logn16", publicNoiseFidelity003FastSHA, publicNoiseFidelity003RootSeed
		schemaSuffix = "003-logn16"
	}
	doc := publicTripletFidelity002Document{
		SchemaVersion: fmt.Sprintf("fast-bootstrap-noise-fidelity-%s-triplet.v1", schemaSuffix), TaskID: taskID,
		Status: "TRIPLET_COMPLETE_EXPLORATORY", GeneratedAt: time.Now().UTC(), Profile: profileName,
		PrimaryCommit: standard.PrimaryCommit, PrimarySourceSHA256: standard.PrimarySourceSHA256,
		StandardCommit: standard.BackendCommit, FastCommit: fastPin, FastBinarySHA256: fast.BinarySHA256,
		InputSHA256: standard.InputSHA256, BootstrapInputSHA256: standard.BootstrapInputSHA256,
		ConfigSHA256: standard.ConfigSHA256, QPSHA256: standard.QPSHA256, WorkloadSHA256: standard.WorkloadSHA256,
		PreflightPairSHA256: standard.PreflightPairSHA256, SigmaCoefficient: publicBootstrapNoiseSigma,
		NoiseRootSeed: rootSeed, NumericalGateMaxAbs: publicNumericalGate,
		Arms:                              map[string]publicTriplet002Arm{"S": standardArm, "F": fastArm, "FN": noiseArm},
		PrecisionComparisonByPhase:        comparison,
		PairwiseOutputsByPhase:            pairwise,
		NoiseImprovedPrecisionSamples:     noiseImproved,
		NoiseWorsenedPrecisionSamples:     noiseWorsened,
		NoiseEqualPrecisionSamples:        noiseEqual,
		FastColdSpeedupVsStandard:         float64(standardArm.Cold.ElapsedNS) / float64(fastArm.Cold.ElapsedNS),
		FastWarmMedianSpeedupVsStandard:   standardArm.WarmElapsedNS.Median / fastArm.WarmElapsedNS.Median,
		NoiseColdSpeedupVsStandard:        float64(standardArm.Cold.ElapsedNS) / float64(noiseArm.Cold.ElapsedNS),
		NoiseWarmMedianSpeedupVsStandard:  standardArm.WarmElapsedNS.Median / noiseArm.WarmElapsedNS.Median,
		NoiseColdTimeOverheadNS:           noiseArm.Cold.ElapsedNS - fastArm.Cold.ElapsedNS,
		NoiseColdTimeOverheadRatio:        float64(noiseArm.Cold.ElapsedNS) / float64(fastArm.Cold.ElapsedNS),
		NoiseWarmMedianOverheadNS:         noiseArm.WarmElapsedNS.Median - fastArm.WarmElapsedNS.Median,
		NoiseWarmMedianOverheadRatio:      noiseArm.WarmElapsedNS.Median / fastArm.WarmElapsedNS.Median,
		NoiseColdAllocationOverheadBytes:  int64(noiseArm.Cold.AllocBytes) - int64(fastArm.Cold.AllocBytes),
		NoiseColdAllocationOverheadAllocs: int64(noiseArm.Cold.Allocs) - int64(fastArm.Cold.Allocs),
		NoiseWarmMedianAllocOverheadBytes: noiseArm.WarmAllocBytes.Median - fastArm.WarmAllocBytes.Median,
		NoiseWarmMedianAllocsOverhead:     noiseArm.WarmAllocs.Median - fastArm.WarmAllocs.Median,
		NoiseEvidence:                     noiseEvidence,
		MatchedHost:                       publicEnvironmentsMatch(standard, fast) && publicEnvironmentsMatch(standard, fastNoise),
		Interpretation:                    "Each sample is compared to the same canonical original Bootstrap plaintext. Global precision bits are -log2(vector RMSE), not the median of per-slot precision. The 1e-6 MaxAbs gate is observational only in this task; structure and provenance remain strict. F/FN full-call allocation/time differences estimate net hook overhead; helper subphase timers are separately recorded and per-injection MemStats remain disabled.",
	}
	if err := writeExclusiveJSON(outPath, doc); err != nil {
		return err
	}
	return nil
}

func validateNoiseFidelity002TripletPreflight(path string, standard, fast, fastNoise publicNativeDocument) error {
	return validateNoiseFidelityTripletPreflight(path, standard, fast, fastNoise, publicNoiseFidelity002TaskID)
}

func validateNoiseFidelityTripletPreflight(path string, standard, fast, fastNoise publicNativeDocument, taskID string) error {
	if path == "" {
		return fmt.Errorf("%s triplet requires the exact zero-call preflight pair artifact", taskID)
	}
	profileName := "logn13"
	if taskID == publicNoiseFidelity003TaskID {
		profileName = "logn16"
	}
	fastPin, err := noiseFidelityTaskFastPin(taskID)
	if err != nil {
		return err
	}
	profile, err := publicProfileFor(profileName)
	if err != nil {
		return err
	}
	pairSHA, err := validatePublicPreflightGateForTaskID(
		path, standard.PrimaryCommit, standard.PrimarySourceSHA256, profile, fastPin,
		standard.ConfigSHA256, standard.QPSHA256, standard.InputSHA256, standard.WorkloadSHA256,
		false, taskID, "standard", standard.BinarySHA256,
	)
	if err != nil {
		return fmt.Errorf("%s Standard preflight/binary provenance: %w", taskID, err)
	}
	fastPairSHA, err := validatePublicPreflightGateForTaskID(
		path, fast.PrimaryCommit, fast.PrimarySourceSHA256, profile, fastPin,
		fast.ConfigSHA256, fast.QPSHA256, fast.InputSHA256, fast.WorkloadSHA256,
		false, taskID, "fast", fast.BinarySHA256,
	)
	if err != nil {
		return fmt.Errorf("%s Fast preflight/binary provenance: %w", taskID, err)
	}
	if pairSHA != fastPairSHA || pairSHA != standard.PreflightPairSHA256 || pairSHA != fast.PreflightPairSHA256 || pairSHA != fastNoise.PreflightPairSHA256 {
		return fmt.Errorf("%s captures do not reference the same exact binary-bound preflight pair", taskID)
	}
	return nil
}

func validatePublicTripletCompatibility002(standard, fast, fastNoise publicNativeDocument, standardVectors, fastVectors, fastNoiseVectors publicVectorDocument) error {
	return validatePublicTripletCompatibilityTask(standard, fast, fastNoise, standardVectors, fastVectors, fastNoiseVectors, publicNoiseFidelity002TaskID)
}

func validatePublicTripletCompatibilityTask(standard, fast, fastNoise publicNativeDocument, standardVectors, fastVectors, fastNoiseVectors publicVectorDocument, taskID string) error {
	profileName := "logn13"
	if taskID == publicNoiseFidelity003TaskID {
		profileName = "logn16"
	}
	profile, err := publicProfileFor(profileName)
	if err != nil {
		return err
	}
	wantMode := "public-native-noise-fidelity-002-repeatability"
	if taskID == publicNoiseFidelity003TaskID {
		wantMode = "public-native-noise-fidelity-003-logn16-repeatability"
	}
	for _, doc := range []publicNativeDocument{standard, fast, fastNoise} {
		if doc.TaskID != taskID || doc.Mode != wantMode ||
			doc.BootstrapBudget != 6 || doc.BootstrapCalls != 6 || doc.Profile != profileName || doc.EphemeralSecretWeight != 32 ||
			!validGitCommitSHA(doc.PrimaryCommit) || !validGitCommitSHA(doc.PrimaryEvidenceCommit) ||
			doc.Parameters.LogN != profile.LogN || doc.Parameters.LogSlots != bits.Len(uint(profile.InputSlots))-1 ||
			doc.Parameters.InputSlots != profile.InputSlots || doc.Parameters.RingN != profile.RingN ||
			!effectiveScaleIs2To45(doc.Parameters.DefaultScale) ||
			doc.ConfigSHA256 != profile.ConfigSHA256 || doc.QPSHA256 != profile.QPSHA256 || doc.InputSHA256 != profile.InputSHA256 || doc.WorkloadSHA256 != profile.WorkloadSHA256 ||
			doc.BootstrapInputSHA256 == "" || doc.PreflightPairSHA256 == "" || doc.CapacityPolicy != profile.CapacityPolicy {
			return fmt.Errorf("%s arm violates its frozen task/profile/source workload contract: task=%q mode=%q budget=%d/%d profile=%q E=%d effective=(LogN%d,LogSlots%d,Scale%s) config=%s QP=%s input=%s workload=%s bootstrap_input=%s preflight=%t",
				taskID, doc.TaskID, doc.Mode, doc.BootstrapBudget, doc.BootstrapCalls, doc.Profile, doc.EphemeralSecretWeight,
				doc.Parameters.LogN, doc.Parameters.LogSlots, doc.Parameters.DefaultScale, doc.ConfigSHA256, doc.QPSHA256,
				doc.InputSHA256, doc.WorkloadSHA256, doc.BootstrapInputSHA256, doc.PreflightPairSHA256 != "")
		}
	}
	fastPin, err := noiseFidelityTaskFastPin(taskID)
	if err != nil {
		return err
	}
	if standard.BackendCommit != publicStandardSHA || fast.BackendCommit != fastPin || fastNoise.BackendCommit != fastPin ||
		!validSHA256Hex(fast.BinarySHA256) || fast.BinarySHA256 != fastNoise.BinarySHA256 ||
		len(fast.BootstrapNoiseEvidence) != 0 || len(standard.BootstrapNoiseEvidence) != 0 {
		return fmt.Errorf("%s source pins, shared Fast/FN binary, or OFF-arm hook state do not match", taskID)
	}
	if standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryCommit != fastNoise.PrimaryCommit ||
		standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 || standard.PrimarySourceSHA256 != fastNoise.PrimarySourceSHA256 ||
		standard.PreflightPairSHA256 != fast.PreflightPairSHA256 || standard.PreflightPairSHA256 != fastNoise.PreflightPairSHA256 ||
		!reflect.DeepEqual(standard.Parameters, fast.Parameters) || !reflect.DeepEqual(standard.Parameters, fastNoise.Parameters) ||
		!publicEnvironmentsMatch(standard, fast) || !publicEnvironmentsMatch(standard, fastNoise) {
		return fmt.Errorf("%s arms mismatch in source, preflight, effective parameters, or execution host", taskID)
	}
	for _, doc := range []publicNativeDocument{standard, fast} {
		for _, phase := range doc.ExecutionPhases {
			if phase.Phase == "bootstrap_noise_hook_configuration" {
				return fmt.Errorf("%s %s arm unexpectedly configured the Noise hook", taskID, doc.Backend)
			}
		}
	}
	noiseHookConfigurations := 0
	for _, phase := range fastNoise.ExecutionPhases {
		if phase.Phase == "bootstrap_noise_hook_configuration" {
			noiseHookConfigurations++
		}
	}
	if noiseHookConfigurations != 1 {
		return fmt.Errorf("%s Fast+Noise arm has %d hook-configuration phases, want exactly one", taskID, noiseHookConfigurations)
	}
	for _, vectors := range []publicVectorDocument{standardVectors, fastVectors, fastNoiseVectors} {
		if vectors.TaskID != taskID || vectors.Mode != wantMode ||
			vectors.BootstrapInputSHA256 != standard.BootstrapInputSHA256 || vectors.ConfigSHA256 != profile.ConfigSHA256 ||
			vectors.QPSHA256 != profile.QPSHA256 || vectors.InputSHA256 != profile.InputSHA256 || vectors.WorkloadSHA256 != profile.WorkloadSHA256 {
			return fmt.Errorf("%s vector provenance or canonical Bootstrap input is mismatched", taskID)
		}
	}
	if standard.BootstrapInputSHA256 != fast.BootstrapInputSHA256 || standard.BootstrapInputSHA256 != fastNoise.BootstrapInputSHA256 ||
		standardVectors.BootstrapInputSHA256 != fastVectors.BootstrapInputSHA256 || standardVectors.BootstrapInputSHA256 != fastNoiseVectors.BootstrapInputSHA256 ||
		!reflect.DeepEqual(standardVectors.BootstrapPlaintextInput, fastVectors.BootstrapPlaintextInput) ||
		!reflect.DeepEqual(standardVectors.BootstrapPlaintextInput, fastNoiseVectors.BootstrapPlaintextInput) {
		return fmt.Errorf("%s arms do not use the same Bootstrap plaintext vector", taskID)
	}
	return nil
}

func effectiveScaleIs2To45(value string) bool {
	scale, err := strconv.ParseFloat(value, 64)
	return err == nil && scale == math.Exp2(45)
}

func summarizePublicTriplet002Arm(doc publicNativeDocument, vectors publicVectorDocument, input []complex128) (publicTriplet002Arm, error) {
	wants, err := publicBootstrapAttemptSpecs(6)
	if err != nil {
		return publicTriplet002Arm{}, err
	}
	if len(doc.BootstrapResults) != 6 || len(vectors.BootstrapOutputs) != 6 {
		return publicTriplet002Arm{}, errors.New("exactly six validated Bootstrap outputs are required")
	}
	arm := publicTriplet002Arm{
		Backend: doc.Backend, BackendCommit: doc.BackendCommit, PrimaryEvidenceCommit: doc.PrimaryEvidenceCommit,
		BinarySHA256: doc.BinarySHA256, Samples: make([]publicTriplet002Sample, 0, 6), Warm: make([]publicTriplet002Sample, 0, 5),
	}
	for index, want := range wants {
		result := doc.BootstrapResults[index]
		if result.Phase != want.phase || result.Timing.Phase != want.phase || !result.Timing.Available || result.Timing.Samples != 1 || result.Timing.ElapsedNS <= 0 || result.InputCiphertextSHA256 == "" {
			return publicTriplet002Arm{}, fmt.Errorf("sample %d lacks its expected measured %s attempt evidence", index, want.phase)
		}
		values, ok := vectors.BootstrapOutputs[want.name]
		if !ok || len(values) != doc.InputSlots {
			return publicTriplet002Arm{}, fmt.Errorf("sample %s output has wrong slot count", want.name)
		}
		output := decodeComplexValues(values)
		if !allFinite(output) || result.Output.DecodedSHA256 == "" || perfmeasure.Fingerprint(output) != result.Output.DecodedSHA256 {
			return publicTriplet002Arm{}, fmt.Errorf("sample %s output is non-finite or has mismatched SHA", want.name)
		}
		errorVector := make([]complex128, len(output))
		for slot := range output {
			errorVector[slot] = output[slot] - input[slot]
		}
		stats := numericalmetrics.SummarizeComplexVector(errorVector)
		tails := numericalmetrics.SummarizeComplexMagnitudeTails(errorVector)
		if stats.Status != "FINITE" || tails.Status != "FINITE" || math.IsNaN(stats.RMS) || math.IsInf(stats.RMS, 0) {
			return publicTriplet002Arm{}, fmt.Errorf("sample %s error metrics are non-finite", want.name)
		}
		precisionStatus := "FINITE"
		var precision *float64
		if stats.RMS == 0 {
			precisionStatus = "UNDEFINED_ZERO_RMSE"
		} else {
			bits := -math.Log2(stats.RMS)
			if math.IsNaN(bits) || math.IsInf(bits, 0) {
				precisionStatus = "UNDEFINED_NONFINITE_PRECISION"
			} else {
				precision = &bits
			}
		}
		gatePass := stats.MaxAbs <= publicNumericalGate
		if result.Output.OraclePass != gatePass || result.Output.Oracle.MaxComplexDifference != stats.MaxAbs {
			return publicTriplet002Arm{}, fmt.Errorf("sample %s does not faithfully report the observational MaxAbs gate", want.name)
		}
		sample := publicTriplet002Sample{
			Name: want.name, Phase: want.phase, InputCiphertextSHA256: result.InputCiphertextSHA256,
			OutputSHA256: result.Output.DecodedSHA256,
			Error:        publicTripletErrorMetrics{Stats: stats, Tails: tails, RMSE: stats.RMS, PrecisionBits: precision, PrecisionBitsStatus: precisionStatus},
			MaxAbs:       stats.MaxAbs, P99: tails.P99, P999: tails.P999, NumericalGate: publicNumericalGate, NumericalGatePass: gatePass,
			ElapsedNS: result.Timing.ElapsedNS, AllocBytes: result.Timing.AllocBytes, Allocs: result.Timing.Allocs,
		}
		arm.Samples = append(arm.Samples, sample)
		if index == 0 {
			arm.Cold = sample
		} else {
			arm.Warm = append(arm.Warm, sample)
		}
	}
	if len(arm.Warm) != 5 {
		return publicTriplet002Arm{}, errors.New("five warm Bootstrap observations are required")
	}
	arm.WarmElapsedNS = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return float64(s.ElapsedNS) }))
	arm.WarmAllocBytes = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return float64(s.AllocBytes) }))
	arm.WarmAllocs = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return float64(s.Allocs) }))
	arm.WarmRMSE = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return s.Error.RMSE }))
	arm.WarmPrecisionBits = summarizeTriplet002Precision(arm.Warm)
	arm.WarmMaxAbs = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return s.MaxAbs }))
	arm.WarmP99 = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return s.P99 }))
	arm.WarmP999 = summarizeTriplet002Float(floatSampleValues(arm.Warm, func(s publicTriplet002Sample) float64 { return s.P999 }))
	constructionNames := map[string]bool{"public_ckks_evaluator_construction": true, "public_bootstrapping_evaluator_construction": true}
	seenConstruction := make(map[string]bool, len(constructionNames))
	seenNoiseHookConfiguration := false
	for _, phase := range doc.ExecutionPhases {
		if constructionNames[phase.Phase] {
			if seenConstruction[phase.Phase] {
				return publicTriplet002Arm{}, fmt.Errorf("evaluator construction phase %s was recorded more than once", phase.Phase)
			}
			seenConstruction[phase.Phase] = true
			if !phase.Available || phase.Samples != 1 {
				return publicTriplet002Arm{}, fmt.Errorf("evaluator construction phase %s is unavailable", phase.Phase)
			}
			arm.EvaluatorConstructionPhases = append(arm.EvaluatorConstructionPhases, phase)
			arm.EvaluatorConstructionElapsedNS += phase.ElapsedNS
			arm.EvaluatorConstructionAllocByte += phase.AllocBytes
			arm.EvaluatorConstructionAllocs += phase.Allocs
		} else if phase.Phase == "bootstrap_noise_hook_configuration" {
			if seenNoiseHookConfiguration {
				return publicTriplet002Arm{}, errors.New("Noise hook configuration phase was recorded more than once")
			}
			seenNoiseHookConfiguration = true
			phaseCopy := phase
			arm.NoiseHookConfiguration = &phaseCopy
		}
	}
	if len(seenConstruction) != len(constructionNames) || len(arm.EvaluatorConstructionPhases) != 2 {
		return publicTriplet002Arm{}, errors.New("both CKKS and Bootstrapping evaluator construction phases must be recorded")
	}
	arm.ConstructionPlusColdElapsedNS = arm.EvaluatorConstructionElapsedNS + arm.Cold.ElapsedNS
	arm.ConstructionPlusColdAllocBytes = arm.EvaluatorConstructionAllocByte + arm.Cold.AllocBytes
	arm.ConstructionPlusColdAllocs = arm.EvaluatorConstructionAllocs + arm.Cold.Allocs
	return arm, nil
}

func summarizePublicTriplet002Noise(doc publicNativeDocument) (publicTriplet002NoiseEvidence, error) {
	return summarizePublicTripletTaskNoise(doc, publicNoiseFidelity002TaskID)
}

func summarizePublicTripletTaskNoise(doc publicNativeDocument, taskID string) (publicTriplet002NoiseEvidence, error) {
	rootSeed := publicNoiseFidelity002RootSeed
	if taskID == publicNoiseFidelity003TaskID {
		rootSeed = publicNoiseFidelity003RootSeed
	}
	var snapshot bootstrapNoiseSnapshotEvidence
	if err := json.Unmarshal(doc.BootstrapNoiseEvidence, &snapshot); err != nil {
		return publicTriplet002NoiseEvidence{}, fmt.Errorf("decode Fast+Noise hook snapshot: %w", err)
	}
	qPrimes, err := parseEffectiveQPrimes(doc.Parameters)
	if err != nil {
		return publicTriplet002NoiseEvidence{}, err
	}
	if err := validateBootstrapNoiseSnapshot(doc.BootstrapNoiseEvidence, publicBootstrapNoiseSigma, rootSeed, doc.Parameters.RingN, qPrimes); err != nil {
		return publicTriplet002NoiseEvidence{}, fmt.Errorf("Fast+Noise hook evidence: %w", err)
	}
	result := publicTriplet002NoiseEvidence{
		Enabled: snapshot.Enabled, Sigma: snapshot.Sigma, RootSeed: snapshot.RootSeed,
		CtCtMultiplyCalls: snapshot.CtCtMultiplyCalls, InjectedPolynomials: snapshot.InjectedPolynomials,
		Events: snapshot.Events, InjectionSubphaseElapsed: make(map[string]int64),
		PerInjectionAllocation: "Per-injection MemStats collection is disabled. Net full-Bootstrap allocation overhead is reported from the paired Fast/Fast+Noise samples.",
	}
	for _, event := range snapshot.Events {
		if event.Left.Seed == 0 || event.Right.Seed == 0 || event.Left.Seed == event.Right.Seed {
			return publicTriplet002NoiseEvidence{}, fmt.Errorf("Noise event %d has invalid independent sampling seeds", event.Operation)
		}
		if event.Left.Seed != expectedBootstrapNoiseSamplingSeed(rootSeed, event.Operation, false) ||
			event.Right.Seed != expectedBootstrapNoiseSamplingSeed(rootSeed, event.Operation, true) {
			return publicTriplet002NoiseEvidence{}, fmt.Errorf("Noise event %d seeds do not match the frozen root/operation/side derivation", event.Operation)
		}
		for _, side := range []bootstrapNoisePolynomialEvidence{event.Left, event.Right} {
			if side.SamplingAlgorithm == "" || !strings.Contains(side.DispatchEvidence, "NTT") || !strings.Contains(side.DispatchEvidence, "MForm-mapped") || len(side.Phases) == 0 ||
				math.IsNaN(side.CoefficientMean) || math.IsInf(side.CoefficientMean, 0) ||
				math.IsNaN(side.CoefficientVariance) || math.IsInf(side.CoefficientVariance, 0) || side.CoefficientVariance < 0 || side.CoefficientMaxAbs == 0 {
				return publicTriplet002NoiseEvidence{}, fmt.Errorf("Noise event %d lacks reproducible sampling/summary/phase evidence", event.Operation)
			}
			wantPhases := []string{"sampling_hash_moments", "coefficient_to_qprefix_ntt", "candidate_c0_add", "qprefix_centered_capacity_check", "commit_c0"}
			if len(side.Phases) != len(wantPhases) {
				return publicTriplet002NoiseEvidence{}, fmt.Errorf("Noise event %d has incomplete phase coverage", event.Operation)
			}
			for index, phase := range side.Phases {
				if phase.Name != wantPhases[index] {
					return publicTriplet002NoiseEvidence{}, fmt.Errorf("Noise event %d phase order mismatch at %d: got %q want %q", event.Operation, index, phase.Name, wantPhases[index])
				}
				// A named commit_c0 measurement can legitimately round to 0 ns
				// because the operation is shorter than the timer's nanosecond
				// resolution. The phase entry and order still prove it was measured;
				// all other subphases must have a positive duration.
				if phase.ElapsedNS < 0 || (phase.ElapsedNS == 0 && phase.Name != "commit_c0") || phase.MemoryStatsCollected {
					return publicTriplet002NoiseEvidence{}, fmt.Errorf("Noise event %d phase %s is absent or collected per-operation MemStats", event.Operation, phase.Name)
				}
				result.InjectionSubphaseElapsed[phase.Name] += phase.ElapsedNS
				result.TotalInjectionSubphaseNS += phase.ElapsedNS
			}
		}
	}
	return result, nil
}

func floatSampleValues(samples []publicTriplet002Sample, value func(publicTriplet002Sample) float64) []float64 {
	values := make([]float64, len(samples))
	for index, sample := range samples {
		values[index] = value(sample)
	}
	return values
}

func summarizeTriplet002Precision(samples []publicTriplet002Sample) publicTriplet002FloatSummary {
	values := make([]float64, 0, len(samples))
	for _, sample := range samples {
		if sample.Error.PrecisionBits != nil {
			values = append(values, *sample.Error.PrecisionBits)
		}
	}
	return summarizeTriplet002Float(values)
}

func summarizeTriplet002Float(values []float64) publicTriplet002FloatSummary {
	if len(values) == 0 {
		return publicTriplet002FloatSummary{}
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	var total float64
	for _, value := range ordered {
		total += value
	}
	mean := total / float64(len(ordered))
	var squared float64
	for _, value := range ordered {
		delta := value - mean
		squared += delta * delta
	}
	stddev := 0.0
	if len(ordered) > 1 {
		stddev = math.Sqrt(squared / float64(len(ordered)-1))
	}
	return publicTriplet002FloatSummary{
		Count: len(ordered), Mean: mean, Median: medianSorted(ordered), SampleStdDev: stddev,
		Min: ordered[0], Max: ordered[len(ordered)-1],
	}
}

func validatePublicBootstrapTripletArm(doc publicNativeDocument, vectors publicVectorDocument, backend string, noiseEnabled bool) error {
	if doc.Mode != "public-native-bootstrap-once" || doc.SchemaVersion != "fast-standard-public-native-bootstrap-once.v2" ||
		doc.Status != publicBootstrapOnceObservedStatus || doc.BootstrapBudget != 1 || doc.BootstrapCalls != 1 ||
		len(doc.BootstrapResults) != 1 || doc.Backend != backend || doc.Profile != "logn13" || doc.EphemeralSecretWeight != 32 ||
		!doc.PrimaryClean || !doc.BackendClean || !validSHA256Hex(doc.BinarySHA256) || doc.PreflightPairSHA256 == "" {
		return errors.New("artifact does not satisfy the exact one-call LogN13/E32 provenance contract")
	}
	if err := validatePublicBackendPinForBootstrapOnce(backend, doc.BackendCommit, doc.Profile, backend == "fast"); err != nil {
		return err
	}
	if vectors.SchemaVersion != "fast-standard-public-native-bootstrap-once-vectors.v1" || vectors.Mode != doc.Mode ||
		vectors.Backend != doc.Backend || vectors.BackendCommit != doc.BackendCommit || vectors.PrimaryCommit != doc.PrimaryCommit ||
		vectors.SourceSHA256 != doc.PrimarySourceSHA256 || vectors.ConfigSHA256 != doc.ConfigSHA256 || vectors.QPSHA256 != doc.QPSHA256 ||
		vectors.InputSHA256 != doc.InputSHA256 || vectors.WorkloadSHA256 != doc.WorkloadSHA256 || vectors.E != doc.EphemeralSecretWeight ||
		vectors.BootstrapInputSHA256 != doc.BootstrapInputSHA256 {
		return errors.New("document/vector provenance fields do not match")
	}
	if noiseEnabled {
		if backend != "fast" {
			return errors.New("noise-enabled arm is not Fast")
		}
		var snapshot bootstrapNoiseSnapshotEvidence
		if len(doc.BootstrapNoiseEvidence) == 0 || jsonUnmarshalBootstrapNoise(doc.BootstrapNoiseEvidence, &snapshot) != nil ||
			!snapshot.Enabled || snapshot.Sigma != publicBootstrapNoiseSigma || snapshot.CtCtMultiplyCalls == 0 || snapshot.InjectedPolynomials != 2*snapshot.CtCtMultiplyCalls {
			return errors.New("noise-enabled artifact lacks fixed-sigma ct×ct event evidence")
		}
		qPrimes, err := parseEffectiveQPrimes(doc.Parameters)
		if err != nil {
			return err
		}
		if err := validateBootstrapNoiseSnapshot(doc.BootstrapNoiseEvidence, publicBootstrapNoiseSigma, snapshot.RootSeed, doc.Parameters.RingN, qPrimes); err != nil {
			return err
		}
	} else if len(doc.BootstrapNoiseEvidence) != 0 {
		return errors.New("noise-disabled arm unexpectedly records an enabled hook snapshot")
	}
	oracles, err := publicOracleCheckpoints(doc.InputSlots)
	if err != nil {
		return err
	}
	if err := validatePublicProfileEvidence(doc); err != nil {
		return err
	}
	bootstrapInput := oracles["drop_level0"]
	if len(vectors.BootstrapPlaintextInput) != len(bootstrapInput) || !reflect.DeepEqual(vectors.BootstrapPlaintextInput, encodeVector(bootstrapInput)) ||
		doc.BootstrapInputSHA256 != perfmeasure.Fingerprint(bootstrapInput) {
		return errors.New("artifact Bootstrap plaintext is not the canonical original message for this workload")
	}
	if err := validatePublicBackendEvidence(doc, vectors, oracles); err != nil {
		return err
	}
	base, baseVectors := doc, vectors
	base.SchemaVersion, base.Mode, base.Status = "fast-standard-public-native-preflight.v1", "public-native", "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN"
	base.BootstrapBudget, base.BootstrapCalls, base.BootstrapResults, base.BootstrapNoiseEvidence = 0, 0, nil, nil
	base.PreflightPairSHA256 = ""
	baseVectors.SchemaVersion, baseVectors.Mode = "fast-standard-public-native-vectors.v1", "public-native"
	baseVectors.BootstrapOutputs = nil
	if err := validatePublicBootstrapOutputEvidence(doc, vectors, base, baseVectors); err != nil {
		return err
	}
	return nil
}

func jsonUnmarshalBootstrapNoise(raw []byte, dst any) error {
	return json.Unmarshal(raw, dst)
}

func validatePublicBootstrapTripletCompatibility(standard, fast, fastNoise publicNativeDocument, standardVectors, fastVectors, fastNoiseVectors publicVectorDocument) error {
	if standard.BackendCommit != publicStandardSHA || fast.BackendCommit != publicBootstrapNoisePilotFastSHA || fastNoise.BackendCommit != publicBootstrapNoisePilotFastSHA ||
		fast.BinarySHA256 != fastNoise.BinarySHA256 {
		return errors.New("triplet source pins or shared Fast/Fast+Noise binary identity do not match")
	}
	if standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryCommit != fastNoise.PrimaryCommit ||
		standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 || standard.PrimarySourceSHA256 != fastNoise.PrimarySourceSHA256 ||
		standard.ProfileKind != fast.ProfileKind || standard.ProfileKind != fastNoise.ProfileKind ||
		standard.InputSHA256 != fast.InputSHA256 || standard.InputSHA256 != fastNoise.InputSHA256 ||
		standard.BootstrapInputSHA256 != fast.BootstrapInputSHA256 || standard.BootstrapInputSHA256 != fastNoise.BootstrapInputSHA256 ||
		standard.ConfigSHA256 != fast.ConfigSHA256 || standard.ConfigSHA256 != fastNoise.ConfigSHA256 ||
		standard.QPSHA256 != fast.QPSHA256 || standard.QPSHA256 != fastNoise.QPSHA256 ||
		standard.WorkloadSHA256 != fast.WorkloadSHA256 || standard.WorkloadSHA256 != fastNoise.WorkloadSHA256 ||
		standard.PreflightPairSHA256 != fast.PreflightPairSHA256 || standard.PreflightPairSHA256 != fastNoise.PreflightPairSHA256 ||
		!reflect.DeepEqual(standard.Parameters, fast.Parameters) || !reflect.DeepEqual(standard.Parameters, fastNoise.Parameters) ||
		!publicEnvironmentsMatch(standard, fast) || !publicEnvironmentsMatch(standard, fastNoise) {
		return errors.New("triplet input, effective parameters, preflight, source, or host provenance is mismatched")
	}
	if standardVectors.BootstrapInputSHA256 != fastVectors.BootstrapInputSHA256 ||
		standardVectors.BootstrapInputSHA256 != fastNoiseVectors.BootstrapInputSHA256 ||
		!reflect.DeepEqual(standardVectors.BootstrapPlaintextInput, fastVectors.BootstrapPlaintextInput) ||
		!reflect.DeepEqual(standardVectors.BootstrapPlaintextInput, fastNoiseVectors.BootstrapPlaintextInput) {
		return errors.New("triplet arms do not use the exact same original Bootstrap plaintext vector")
	}
	return nil
}

func summarizePublicTripletArm(doc publicNativeDocument, output, input []complex128) (publicTripletArm, error) {
	if len(output) != doc.InputSlots || len(input) != doc.InputSlots || !allFinite(output) || !allFinite(input) {
		return publicTripletArm{}, errors.New("triplet output and original input must be matching finite slot vectors")
	}
	errorVector := make([]complex128, len(output))
	for index := range output {
		errorVector[index] = output[index] - input[index]
	}
	stats := numericalmetrics.SummarizeComplexVector(errorVector)
	tails := numericalmetrics.SummarizeComplexMagnitudeTails(errorVector)
	if stats.Status != "FINITE" || tails.Status != "FINITE" {
		return publicTripletArm{}, errors.New("triplet error statistics are non-finite")
	}
	precisionStatus := "FINITE"
	var precision *float64
	if stats.RMS == 0 {
		precisionStatus = "UNDEFINED_ZERO_RMSE"
	} else {
		bits := -math.Log2(stats.RMS)
		if math.IsNaN(bits) || math.IsInf(bits, 0) {
			precisionStatus = "UNDEFINED_NONFINITE_PRECISION"
		} else {
			precision = &bits
		}
	}
	result := publicBootstrapResultByPhase(doc.BootstrapResults, "single_bootstrap")
	if result.Phase == "" || result.Timing.Samples != 1 || !result.Timing.Available || result.Timing.ElapsedNS <= 0 {
		return publicTripletArm{}, errors.New("single Bootstrap call timing evidence is missing")
	}
	return publicTripletArm{
		Backend: doc.Backend, BackendCommit: doc.BackendCommit, BinarySHA256: doc.BinarySHA256,
		BootstrapCalls: doc.BootstrapCalls, SuccessfulSlots: len(output), BootstrapInputSHA256: doc.BootstrapInputSHA256,
		OutputSHA256: result.Output.DecodedSHA256, BootstrapTimingPhase: result.Phase,
		Error:               publicTripletErrorMetrics{Stats: stats, Tails: tails, RMSE: stats.RMS, PrecisionBits: precision, PrecisionBitsStatus: precisionStatus},
		NumericalGateMaxAbs: result.Output.Oracle.MaxComplexDifference, NumericalGate: publicNumericalGate,
		NumericalGatePass: result.Output.OraclePass,
		BootstrapTimeNS:   result.Timing.ElapsedNS, AllocBytes: result.Timing.AllocBytes, Allocs: result.Timing.Allocs,
		PeakMemoryStatus: "UNAVAILABLE_NO_PEAK_MEMORY_METRIC_IN_PERFMEASURE",
	}, nil
}

func makePublicTripletPairwise(leftName, rightName string, left, right []complex128) publicTripletPairwise {
	errorVector := make([]complex128, len(left))
	for index := range left {
		errorVector[index] = left[index] - right[index]
	}
	stats := numericalmetrics.SummarizeComplexVector(errorVector)
	return publicTripletPairwise{Left: leftName, Right: rightName, RMSE: stats.RMS, Stats: stats, Tails: numericalmetrics.SummarizeComplexMagnitudeTails(errorVector)}
}

func bootstrapNoiseRootSeed(raw []byte) int64 {
	var snapshot bootstrapNoiseSnapshotEvidence
	if jsonUnmarshalBootstrapNoise(raw, &snapshot) != nil {
		return 0
	}
	return snapshot.RootSeed
}
