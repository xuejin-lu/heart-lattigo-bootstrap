package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"reflect"
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

func runPublicTripletCompare(args []string) error {
	flags := flag.NewFlagSet("perfprobe compare-public-triplet", flag.ContinueOnError)
	standardPath := flags.String("standard", "", "single-call genuine Standard Bootstrap JSON")
	standardVectorsPath := flags.String("standard-vectors", "", "single-call genuine Standard decoded vectors JSON")
	fastPath := flags.String("fast", "", "single-call Fast Bootstrap with the hook disabled JSON")
	fastVectorsPath := flags.String("fast-vectors", "", "single-call Fast hook-disabled decoded vectors JSON")
	fastNoisePath := flags.String("fast-noise", "", "single-call Fast Bootstrap with fixed-sigma hook enabled JSON")
	fastNoiseVectorsPath := flags.String("fast-noise-vectors", "", "single-call Fast fixed-sigma decoded vectors JSON")
	outPath := flags.String("out", "", "exclusive compact triplet result JSON")
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
		SchemaVersion: "fast-bootstrap-noise-fidelity-triplet.v1", Status: "TRIPLET_COMPLETE_EXPLORATORY", GeneratedAt: time.Now().UTC(),
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
		Interpretation:             "Each arm's precision is computed against the same canonical plaintext supplied to Bootstrap; pairwise output RMSE is a separate comparison. A complete triplet is exploratory evidence, not proof of general noise fidelity.",
	}
	if err := writeExclusiveJSON(*outPath, doc); err != nil {
		return err
	}
	return nil
}

func validatePublicBootstrapTripletArm(doc publicNativeDocument, vectors publicVectorDocument, backend string, noiseEnabled bool) error {
	if doc.Mode != "public-native-bootstrap-once" || doc.SchemaVersion != "fast-standard-public-native-bootstrap-once.v1" ||
		doc.Status != "PASS_PUBLIC_BOOTSTRAP_ONCE" || doc.BootstrapBudget != 1 || doc.BootstrapCalls != 1 ||
		len(doc.BootstrapResults) != 1 || doc.Backend != backend || doc.Profile != "logn13" || doc.EphemeralSecretWeight != 32 ||
		!doc.PrimaryClean || !doc.BackendClean || !validSHA256Hex(doc.BinarySHA256) || doc.PreflightPairSHA256 == "" {
		return errors.New("artifact does not satisfy the exact one-call LogN13/E32 provenance contract")
	}
	if err := validatePublicBackendPin(backend, doc.BackendCommit, doc.Profile); err != nil {
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
	if standard.BackendCommit != publicStandardSHA || fast.BackendCommit != publicBootstrapNoiseFastSHA || fastNoise.BackendCommit != publicBootstrapNoiseFastSHA ||
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
		Error:           publicTripletErrorMetrics{Stats: stats, Tails: tails, RMSE: stats.RMS, PrecisionBits: precision, PrecisionBitsStatus: precisionStatus},
		BootstrapTimeNS: result.Timing.ElapsedNS, AllocBytes: result.Timing.AllocBytes, Allocs: result.Timing.Allocs,
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
