package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	task005ResultBase          = "FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005"
	logN16ConfigSHA            = "16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02"
	logN16InputSHA             = "659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf"
	logN16WorkloadSHA          = "1413c0f0edc3c6b52d7a71e4714efe70280854e1208f87640b741682b65fa2a9"
	logN16QPSHA                = "6c82ca724cbfdce2083b2ef2c18d53f81e658a6f28ff283e9a671d276710ec5f"
	standardFastBaselineCommit = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
)

type publicE32Task005Baseline struct {
	Classification         string  `json:"classification"`
	PrimaryCommit          string  `json:"primary_measurement_commit"`
	PrimarySourceSHA256    string  `json:"primary_measurement_source_sha256"`
	StandardCommit         string  `json:"standard_lattigo_commit"`
	FastCommit             string  `json:"fast_lattigo_commit"`
	ConfigSHA256           string  `json:"config_sha256"`
	InputSHA256            string  `json:"input_sha256"`
	WorkloadSHA256         string  `json:"workload_sha256"`
	QPSHA256               string  `json:"qp_sha256"`
	AttemptLedgerSHA256    string  `json:"attempt_ledger_sha256"`
	StandardResultSHA256   string  `json:"standard_result_sha256"`
	FastResultSHA256       string  `json:"fast_result_sha256"`
	BootstrapPairSHA256    string  `json:"bootstrap_pair_sha256"`
	ReservedCalls          int     `json:"reserved_calls"`
	ActualCalls            int     `json:"actual_calls"`
	StandardColdMS         float64 `json:"standard_cold_ms"`
	StandardWarmMedianMS   float64 `json:"standard_warm_median_ms"`
	FastColdMS             float64 `json:"fast_cold_ms"`
	FastWarmMedianMS       float64 `json:"fast_warm_median_ms"`
	FastWarmAllocatedBytes uint64  `json:"fast_warm_median_allocated_bytes"`
	FastWarmAllocations    uint64  `json:"fast_warm_median_allocations"`
	FastVsStandardRMSE     float64 `json:"fast_vs_standard_complex_rmse"`
	FastVsStandardMax      float64 `json:"fast_vs_standard_max_complex_difference"`
	NumericalGate          float64 `json:"numerical_gate"`
	ReferenceStatus        string  `json:"fast_reference_status"`
	ReferenceVectorsSHA256 string  `json:"fast_reference_vectors_sha256,omitempty"`
	ReferenceResultSHA256  string  `json:"fast_reference_result_sha256,omitempty"`
}

type task005RunArtifact struct {
	SchemaVersion         string                          `json:"schema_version"`
	Mode                  string                          `json:"mode"`
	Status                string                          `json:"status"`
	Profile               string                          `json:"profile"`
	Backend               string                          `json:"backend"`
	BackendCommit         string                          `json:"backend_commit"`
	PrimaryCommit         string                          `json:"primary_commit"`
	PrimarySourceSHA256   string                          `json:"primary_measurement_source_sha256"`
	BackendClean          bool                            `json:"backend_clean"`
	PrimaryClean          bool                            `json:"primary_clean"`
	ConfigSHA256          string                          `json:"config_sha256"`
	InputSHA256           string                          `json:"canonical_input_sha256"`
	WorkloadSHA256        string                          `json:"workload_sha256"`
	QPSHA256              string                          `json:"qp_sha256"`
	EffectiveParameters   perfmeasure.EffectiveParameters `json:"effective_parameters"`
	EphemeralSecretWeight int                             `json:"ephemeral_secret_weight"`
	InputSlots            int                             `json:"input_slots"`
	BootstrapBudget       int                             `json:"bootstrap_budget"`
	BootstrapCalls        int                             `json:"actual_bootstrap_calls"`
	NumericalGate         float64                         `json:"max_complex_numerical_gate"`
	BootstrapResults      []task005BootstrapResult        `json:"bootstrap_results"`
	Checkpoints           []task005Checkpoint             `json:"checkpoints"`
}

type task005BootstrapResult struct {
	Phase  string            `json:"phase"`
	Timing task005Timing     `json:"timing"`
	Output task005Checkpoint `json:"output_checkpoint"`
}

type task005Timing struct {
	ElapsedNS      int64  `json:"elapsed_ns"`
	AllocatedBytes uint64 `json:"allocated_bytes"`
	Allocations    uint64 `json:"allocations"`
	Available      bool   `json:"available"`
}

type task005Checkpoint struct {
	Name          string         `json:"name"`
	DecodedSHA256 string         `json:"decoded_sha256"`
	OraclePass    bool           `json:"oracle_pass"`
	Oracle        task005Metrics `json:"oracle_metrics"`
	State         task005State   `json:"state"`
}

type task005Metrics struct {
	ComplexRMSE          float64 `json:"complex_rmse"`
	MaxComplexDifference float64 `json:"max_complex_difference"`
}

type task005State struct {
	Level       int    `json:"level"`
	Degree      int    `json:"degree"`
	Scale       string `json:"scale"`
	QPrefixRows int    `json:"q_prefix_rows"`
}

type task005PairArtifact struct {
	SchemaVersion        string                      `json:"schema_version"`
	Status               string                      `json:"status"`
	ActualBootstrapCalls int                         `json:"actual_bootstrap_calls"`
	StandardCommit       string                      `json:"standard_backend_commit"`
	FastCommit           string                      `json:"fast_backend_commit"`
	PrimaryCommit        string                      `json:"primary_commit"`
	SourceSHA256         string                      `json:"primary_measurement_source_sha256"`
	ConfigSHA256         string                      `json:"config_sha256"`
	InputSHA256          string                      `json:"input_sha256"`
	WorkloadSHA256       string                      `json:"workload_sha256"`
	QPSHA256             string                      `json:"qp_sha256"`
	NumericalGate        float64                     `json:"max_complex_gate"`
	Comparisons          []task005PairBootstrapCheck `json:"bootstrap_comparisons"`
}

type task005PairBootstrapCheck struct {
	Phase        string         `json:"phase"`
	StateMatched bool           `json:"level_scale_degree_matched"`
	Metrics      task005Metrics `json:"fast_vs_standard"`
	Pass         bool           `json:"pass"`
}

type task005LedgerArtifact struct {
	Task       string `json:"task"`
	TotalCalls int    `json:"total_authorized_calls_across_004_and_005"`
	Task004    struct {
		Reserved int `json:"reserved_calls"`
		Actual   int `json:"actual_calls"`
	} `json:"task_004"`
	Task005 struct {
		Reserved int `json:"reserved_calls"`
		Actual   int `json:"actual_calls"`
	} `json:"task_005"`
}

type task005ReferenceInfo struct {
	Vectors       publicE32Vectors
	VectorsSHA256 string
	ResultSHA256  string
	Path          string
	ResultPath    string
}

func loadPublicE32Task005Baseline(primaryRoot string) (publicE32Task005Baseline, task005RunArtifact, error) {
	resultsRoot := filepath.Join(primaryRoot, "results")
	ledgerPath := filepath.Join(resultsRoot, task005ResultBase+"-attempt-ledger.json")
	standardPath := filepath.Join(resultsRoot, task005ResultBase+"-logn16-standard.json")
	fastPath := filepath.Join(resultsRoot, task005ResultBase+"-logn16-fast.json")
	pairPath := filepath.Join(resultsRoot, task005ResultBase+"-logn16-bootstrap-pair.json")
	var ledger task005LedgerArtifact
	if err := readTaskArtifact(ledgerPath, &ledger); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	if ledger.Task != "FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005" || ledger.TotalCalls != 16 ||
		ledger.Task004.Reserved != 0 || ledger.Task004.Actual != 0 || ledger.Task005.Reserved != 16 || ledger.Task005.Actual != 16 {
		return publicE32Task005Baseline{}, task005RunArtifact{}, errors.New("Task005 attempt ledger is not the verified immutable 16/16 state")
	}
	if err := validateTask005AttemptJournals(resultsRoot); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	var standard, fast task005RunArtifact
	if err := readTaskArtifact(standardPath, &standard); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	if err := readTaskArtifact(fastPath, &fast); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	var pair task005PairArtifact
	if err := readTaskArtifact(pairPath, &pair); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	if err := validateTask005BackendResult(standard, "standard", standardFastBaselineCommit); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	if err := validateTask005BackendResult(fast, "fast", e32LogN16ProductionFastCommit); err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	if pair.SchemaVersion != "fast-standard-public-native-pair.v1" || pair.Status != "PASS" || pair.ActualBootstrapCalls != 12 ||
		pair.StandardCommit != standardFastBaselineCommit || pair.FastCommit != e32LogN16ProductionFastCommit ||
		pair.PrimaryCommit != e32Task005PrimaryCommit || pair.SourceSHA256 != e32Task005SourceSHA256 ||
		pair.ConfigSHA256 != logN16ConfigSHA || pair.InputSHA256 != logN16InputSHA || pair.WorkloadSHA256 != logN16WorkloadSHA ||
		pair.QPSHA256 != logN16QPSHA || pair.NumericalGate != 1e-6 || len(pair.Comparisons) != 6 {
		return publicE32Task005Baseline{}, task005RunArtifact{}, errors.New("Task005 paired Bootstrap artifact does not match its pinned LogN16/E32 provenance")
	}
	for _, comparison := range pair.Comparisons {
		if !comparison.Pass || !comparison.StateMatched || comparison.Metrics.ComplexRMSE < 0 || comparison.Metrics.MaxComplexDifference > pair.NumericalGate ||
			!finiteTask005(comparison.Metrics.ComplexRMSE) || !finiteTask005(comparison.Metrics.MaxComplexDifference) {
			return publicE32Task005Baseline{}, task005RunArtifact{}, fmt.Errorf("Task005 paired output %s fails its stored numerical/state gate", comparison.Phase)
		}
	}
	standardWarm, standardCold, err := task005Timings(standard)
	if err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	fastWarm, fastCold, err := task005Timings(fast)
	if err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	ledgerBytes, ledgerSHA, err := taskArtifactBytes(ledgerPath)
	if err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	_ = ledgerBytes
	standardBytes, standardSHA, err := taskArtifactBytes(standardPath)
	if err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	_ = standardBytes
	fastBytes, fastSHA, err := taskArtifactBytes(fastPath)
	if err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	_ = fastBytes
	pairBytes, pairSHA, err := taskArtifactBytes(pairPath)
	if err != nil {
		return publicE32Task005Baseline{}, task005RunArtifact{}, err
	}
	_ = pairBytes
	worstRMSE, worstMax := 0.0, 0.0
	for _, item := range pair.Comparisons {
		worstRMSE = math.Max(worstRMSE, item.Metrics.ComplexRMSE)
		worstMax = math.Max(worstMax, item.Metrics.MaxComplexDifference)
	}
	fastWarmResults := make([]task005BootstrapResult, 0, 5)
	for _, result := range fast.BootstrapResults {
		if result.Phase != "first_cold_bootstrap" {
			fastWarmResults = append(fastWarmResults, result)
		}
	}
	medianIndex := (len(fastWarmResults) - 1) / 2
	sort.Slice(fastWarmResults, func(i, j int) bool { return fastWarmResults[i].Timing.ElapsedNS < fastWarmResults[j].Timing.ElapsedNS })
	medianFast := fastWarmResults[medianIndex].Timing
	baseline := publicE32Task005Baseline{
		Classification: "CURRENT_LOGN16_FAST_SPEED_VERIFIED", PrimaryCommit: e32Task005PrimaryCommit,
		PrimarySourceSHA256: e32Task005SourceSHA256, StandardCommit: standardFastBaselineCommit, FastCommit: e32LogN16ProductionFastCommit,
		ConfigSHA256: logN16ConfigSHA, InputSHA256: logN16InputSHA, WorkloadSHA256: logN16WorkloadSHA, QPSHA256: logN16QPSHA,
		AttemptLedgerSHA256: ledgerSHA, StandardResultSHA256: standardSHA, FastResultSHA256: fastSHA, BootstrapPairSHA256: pairSHA,
		ReservedCalls: 16, ActualCalls: 16, StandardColdMS: float64(standardCold.ElapsedNS) / 1e6,
		StandardWarmMedianMS: float64(standardWarm) / 1e6, FastColdMS: float64(fastCold.ElapsedNS) / 1e6,
		FastWarmMedianMS: float64(fastWarm) / 1e6, FastWarmAllocatedBytes: medianFast.AllocatedBytes,
		FastWarmAllocations: medianFast.Allocations, FastVsStandardRMSE: worstRMSE, FastVsStandardMax: worstMax,
		NumericalGate: pair.NumericalGate, ReferenceStatus: "TASK005_FAST_REFERENCE_VECTORS_PENDING_VALIDATION",
	}
	return baseline, fast, nil
}

func validateTask005BackendResult(doc task005RunArtifact, backend, pin string) error {
	wantMode := "public-native-repeatability"
	if doc.SchemaVersion != "fast-standard-public-native-repeatability.v1" || doc.Mode != wantMode || doc.Status != "PASS_PUBLIC_BOOTSTRAP_1COLD_5WARM" ||
		doc.Profile != "logn16" || doc.Backend != backend || doc.BackendCommit != pin || doc.PrimaryCommit != e32Task005PrimaryCommit ||
		doc.PrimarySourceSHA256 != e32Task005SourceSHA256 || !doc.PrimaryClean || !doc.BackendClean || doc.ConfigSHA256 != logN16ConfigSHA ||
		doc.InputSHA256 != logN16InputSHA || doc.WorkloadSHA256 != logN16WorkloadSHA || doc.QPSHA256 != logN16QPSHA ||
		doc.EphemeralSecretWeight != 32 || doc.InputSlots != 1<<15 || doc.BootstrapBudget != 6 || doc.BootstrapCalls != 6 ||
		doc.NumericalGate != 1e-6 || len(doc.BootstrapResults) != 6 || len(doc.Checkpoints) != 8 {
		return fmt.Errorf("Task005 %s artifact does not match its frozen LogN16/E32 six-call provenance", backend)
	}
	for index, result := range doc.BootstrapResults {
		wantPhase := "warm_bootstrap_" + fmt.Sprintf("%02d", index)
		if index == 0 {
			wantPhase = "first_cold_bootstrap"
		}
		if result.Phase != wantPhase || !result.Timing.Available || result.Timing.ElapsedNS <= 0 || !result.Output.OraclePass ||
			result.Output.State.Level != 1 || result.Output.State.Degree != 1 || result.Output.State.QPrefixRows != 2 ||
			result.Output.DecodedSHA256 == "" || result.Output.Oracle.MaxComplexDifference > doc.NumericalGate {
			return fmt.Errorf("Task005 %s result %d fails its recorded output/timing gate", backend, index+1)
		}
	}
	return nil
}

func task005Timings(doc task005RunArtifact) (warmMedianNS int64, cold task005Timing, err error) {
	warm := make([]int64, 0, 5)
	for _, result := range doc.BootstrapResults {
		if result.Phase == "first_cold_bootstrap" {
			cold = result.Timing
		} else {
			warm = append(warm, result.Timing.ElapsedNS)
		}
	}
	if cold.ElapsedNS <= 0 || len(warm) != 5 {
		return 0, task005Timing{}, errors.New("Task005 timing artifacts do not contain one cold and five warm samples")
	}
	sort.Slice(warm, func(i, j int) bool { return warm[i] < warm[j] })
	return warm[len(warm)/2], cold, nil
}

func validateTask005FastReference(primaryRoot, vectorsPath string, manifest publicE32FixtureManifest) (task005ReferenceInfo, error) {
	resultPath := filepath.Join(primaryRoot, "results", task005ResultBase+"-logn16-fast.json")
	var fast task005RunArtifact
	if err := readTaskArtifact(resultPath, &fast); err != nil {
		return task005ReferenceInfo{}, err
	}
	if err := validateTask005BackendResult(fast, "fast", e32LogN16ProductionFastCommit); err != nil {
		return task005ReferenceInfo{}, err
	}
	vectorBytes, vectorSHA, err := taskArtifactBytes(vectorsPath)
	if err != nil {
		return task005ReferenceInfo{}, err
	}
	var vectors publicE32Vectors
	if err := json.Unmarshal(vectorBytes, &vectors); err != nil {
		return task005ReferenceInfo{}, fmt.Errorf("decode Task005 Fast reference vectors: %w", err)
	}
	if vectors.SchemaVersion != "fast-standard-public-native-repeatability-vectors.v1" || vectors.Mode != "public-native-repeatability" ||
		vectors.Backend != "fast" || vectors.BackendCommit != e32LogN16ProductionFastCommit || vectors.PrimaryCommit != e32Task005PrimaryCommit ||
		vectors.SourceSHA256 != e32Task005SourceSHA256 || vectors.ConfigSHA256 != logN16ConfigSHA || vectors.QPSHA256 != logN16QPSHA ||
		vectors.InputSHA256 != logN16InputSHA || vectors.WorkloadSHA256 != logN16WorkloadSHA || vectors.EphemeralSecretWeight != 32 {
		return task005ReferenceInfo{}, errors.New("Task005 Fast reference vectors have mismatched profile or source provenance")
	}
	if manifest.Profile != "logn16-e32-public-native" || manifest.BackendCommit != e32LogN16ProductionFastCommit ||
		manifest.ConfigSHA256 != logN16ConfigSHA || manifest.InputSHA256 != logN16InputSHA || manifest.WorkloadSHA256 != logN16WorkloadSHA || manifest.QPSHA256 != logN16QPSHA {
		return task005ReferenceInfo{}, errors.New("Task005 Fast references are not compatible with the held LogN16 native fixture")
	}
	checkpoints := make(map[string]task005Checkpoint, len(fast.Checkpoints))
	for _, checkpoint := range fast.Checkpoints {
		checkpoints[checkpoint.Name] = checkpoint
	}
	if len(checkpoints) != 8 || len(vectors.Checkpoints) != 8 {
		return task005ReferenceInfo{}, errors.New("Task005 Fast references do not contain the exact eight public checkpoint vectors")
	}
	for name, expected := range checkpoints {
		values := vectors.Checkpoints[name]
		if len(values) != 1<<15 || expected.DecodedSHA256 == "" || fingerprintE32Vectors(values) != expected.DecodedSHA256 || !finiteE32Vectors(values) {
			return task005ReferenceInfo{}, fmt.Errorf("Task005 checkpoint vector %s does not match its committed decoded hash", name)
		}
	}
	if len(vectors.BootstrapOutputs) != 6 || len(fast.BootstrapResults) != 6 {
		return task005ReferenceInfo{}, errors.New("Task005 Fast reference must contain exactly six Bootstrap outputs")
	}
	for _, result := range fast.BootstrapResults {
		name := result.Output.Name
		values := vectors.BootstrapOutputs[name]
		if len(values) != 1<<15 || fingerprintE32Vectors(values) != result.Output.DecodedSHA256 || !finiteE32Vectors(values) {
			return task005ReferenceInfo{}, fmt.Errorf("Task005 Bootstrap vector %s does not match its committed decoded hash", name)
		}
	}
	resultBytes, resultSHA, err := taskArtifactBytes(resultPath)
	if err != nil {
		return task005ReferenceInfo{}, err
	}
	_ = resultBytes
	return task005ReferenceInfo{Vectors: vectors, VectorsSHA256: vectorSHA, ResultSHA256: resultSHA, Path: vectorsPath, ResultPath: resultPath}, nil
}

func validateTask005AttemptJournals(resultsRoot string) error {
	type expectedRun struct {
		name   string
		budget int
		phases []string
	}
	runs := []expectedRun{
		{"FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005-logn13-standard", 2, []string{"first_cold_bootstrap", "later_warm_bootstrap"}},
		{"FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005-logn13-fast", 2, []string{"first_cold_bootstrap", "later_warm_bootstrap"}},
		{"FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005-logn16-standard", 6, []string{"first_cold_bootstrap", "warm_bootstrap_01", "warm_bootstrap_02", "warm_bootstrap_03", "warm_bootstrap_04", "warm_bootstrap_05"}},
		{"FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005-logn16-fast", 6, []string{"first_cold_bootstrap", "warm_bootstrap_01", "warm_bootstrap_02", "warm_bootstrap_03", "warm_bootstrap_04", "warm_bootstrap_05"}},
	}
	count := 0
	for _, run := range runs {
		for index, phase := range run.phases {
			path := filepath.Join(resultsRoot, run.name+".json.bootstrap-attempt-"+fmt.Sprintf("%02d", index+1)+".json")
			var reservation struct {
				SchemaVersion string    `json:"schema_version"`
				Index         int       `json:"attempt_index"`
				Budget        int       `json:"bootstrap_budget"`
				Phase         string    `json:"phase"`
				ReservedAt    time.Time `json:"reserved_at"`
				Status        string    `json:"status"`
			}
			if err := readTaskArtifact(path, &reservation); err != nil {
				return fmt.Errorf("validate immutable Task005 attempt journal: %w", err)
			}
			if reservation.SchemaVersion != "perfprobe-bootstrap-attempt.v1" || reservation.Index != index+1 || reservation.Budget != run.budget ||
				reservation.Phase != phase || reservation.Status != "irrevocably_reserved_before_call" || reservation.ReservedAt.IsZero() {
				return fmt.Errorf("Task005 attempt journal %s does not match its recorded reservation", filepath.Base(path))
			}
			count++
		}
	}
	if count != 16 {
		return fmt.Errorf("Task005 journal audit found %d irrevocable reservations, want exactly 16", count)
	}
	return nil
}

func readTaskArtifact(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read required Task005 artifact %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode required Task005 artifact %s: %w", path, err)
	}
	return nil
}

func taskArtifactBytes(path string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read artifact %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func fingerprintE32Vectors(values []publicE32Vector) string {
	decoded := make([]complex128, len(values))
	for index, value := range values {
		decoded[index] = complex(value.Real, value.Imag)
	}
	return perfmeasure.Fingerprint(decoded)
}

func finiteE32Vectors(values []publicE32Vector) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if math.IsNaN(value.Real) || math.IsInf(value.Real, 0) || math.IsNaN(value.Imag) || math.IsInf(value.Imag, 0) {
			return false
		}
	}
	return true
}

func finiteTask005(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
