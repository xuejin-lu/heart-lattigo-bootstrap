package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"
	"reflect"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const fixedSigmaAblationManifestSchema = "fixed-sigma-propagation-ablation-manifest.v1"
const fixedSigmaDirectMulAblationManifestSchema = "fixed-sigma-direct-ctct-ablation-manifest.v1"

const (
	fixedSigmaAblationManifestMaxBytes = 1 << 20
	fixedSigmaAblationResultMaxBytes   = 8 << 20
	fixedSigmaAblationVectorsMaxBytes  = 32 << 20
)

type fixedSigmaAblationArtifactRef struct {
	Result  string `json:"result"`
	Vectors string `json:"vectors"`
}

type fixedSigmaAblationSeedPair struct {
	Stage1Seed    int64                                    `json:"stage1_seed"`
	Arms          map[string]fixedSigmaAblationArtifactRef `json:"arms"`
	BaselineReuse *fixedSigmaAblationBaselineReuse         `json:"baseline_reuse,omitempty"`
}

type fixedSigmaAblationBaselineReuse struct {
	OriginSlot       string                        `json:"origin_slot"`
	OriginStage1Seed int64                         `json:"origin_stage1_seed"`
	OriginStage2Seed int64                         `json:"origin_stage2_seed"`
	WitnessSlot      string                        `json:"equivalence_witness_slot"`
	Witness          fixedSigmaAblationArtifactRef `json:"equivalence_witness"`
}

type fixedSigmaAblationWorkload struct {
	Name      string                        `json:"name"`
	CVariant  string                        `json:"c_input_variant"`
	Standard  fixedSigmaAblationArtifactRef `json:"standard_off_off"`
	SeedPairs []fixedSigmaAblationSeedPair  `json:"seed_pairs"`
}

type fixedSigmaAblationManifest struct {
	SchemaVersion string                       `json:"schema_version"`
	Workloads     []fixedSigmaAblationWorkload `json:"workloads"`
}

type fixedSigmaVectorAnalysis struct {
	Statistics numericalmetrics.ComplexVectorStats    `json:"statistics"`
	Tails      numericalmetrics.ComplexMagnitudeTails `json:"magnitude_tails"`
}

type fixedSigmaAblationArmAnalysis struct {
	Stage1Incremental fixedSigmaVectorAnalysis `json:"stage1_incremental_residual"`
	Stage1Total       fixedSigmaVectorAnalysis `json:"stage1_total_oracle_error"`
	Stage2Incremental fixedSigmaVectorAnalysis `json:"stage2_incremental_residual"`
	Stage2Total       fixedSigmaVectorAnalysis `json:"stage2_total_oracle_error"`
}

type fixedSigmaAblationPropagation struct {
	Stage1InjectedDelta          fixedSigmaVectorAnalysis           `json:"stage1_injected_delta"`
	Stage2PropagatedP1           fixedSigmaVectorAnalysis           `json:"stage2_propagated_p1"`
	Stage2IncrementalONOFF       fixedSigmaVectorAnalysis           `json:"stage2_incremental_residual_on_off"`
	Stage2IncrementalOFFOFF      fixedSigmaVectorAnalysis           `json:"stage2_incremental_residual_off_off"`
	FirstOrderPrediction         fixedSigmaVectorAnalysis           `json:"first_order_prediction_delta1_times_c_decoded"`
	FirstOrderResidual           fixedSigmaVectorAnalysis           `json:"p1_minus_first_order_prediction"`
	DecompositionResidual        fixedSigmaVectorAnalysis           `json:"p1_minus_full_decomposition"`
	P1PredictionCovariance       numericalmetrics.ComplexCovariance `json:"p1_prediction_covariance"`
	Stage2OperandDecodedSHA256   string                             `json:"stage2_operand_decoded_sha256"`
	Stage2OperandMatchedAllArms  bool                               `json:"stage2_operand_matched_all_fast_arms"`
	FastActiveInputRowsMatched   bool                               `json:"fast_active_input_rows_matched_all_arms"`
	DecompositionTolerance       string                             `json:"decomposition_tolerance"`
	MaxDecompositionResidual     float64                            `json:"max_decomposition_residual"`
	MaxFloatingPointTolerance    float64                            `json:"max_floating_point_tolerance"`
	DecompositionToleranceFails  int                                `json:"decomposition_tolerance_failures"`
	InjectedDeltaEvidenceMatched bool                               `json:"stage1_helper_delta_matches_ablation"`
}

type fixedSigmaAblationSecondCheck struct {
	ONONMinusOFFON   fixedSigmaVectorAnalysis           `json:"on_on_minus_off_on"`
	DifferenceFromP1 fixedSigmaVectorAnalysis           `json:"difference_from_p1"`
	CovarianceWithP1 numericalmetrics.ComplexCovariance `json:"covariance_with_p1"`
}

type fixedSigmaAblationSeedPairResult struct {
	Stage1Seed     int64                                    `json:"stage1_seed"`
	Stage2Seed     int64                                    `json:"stage2_seed"`
	FastBinarySHA  string                                   `json:"fast_binary_sha256"`
	BaselineOrigin string                                   `json:"baseline_origin,omitempty"`
	BaselineReuse  *fixedSigmaAblationBaselineReuseEvidence `json:"baseline_reuse_evidence,omitempty"`
	ArtifactSHA256 map[string]map[string]string             `json:"artifact_sha256"`
	Arms           map[string]fixedSigmaAblationArmAnalysis `json:"fast_arm_operation_metrics"`
	Standard       fixedSigmaAblationArmAnalysis            `json:"standard_off_off_operation_metrics"`
	Propagation    fixedSigmaAblationPropagation            `json:"stage1_to_stage2_propagation"`
	SecondCheck    fixedSigmaAblationSecondCheck            `json:"on_on_context_check"`
}

type fixedSigmaAblationBaselineReuseEvidence struct {
	OriginSlot                  string   `json:"baseline_origin"`
	OriginStageSeeds            [2]int64 `json:"baseline_source_stage_seeds"`
	RequestedStageSeeds         [2]int64 `json:"requested_pair_stage_seeds"`
	OriginBinarySHA256          string   `json:"baseline_binary_sha256"`
	OriginResultSHA256          string   `json:"baseline_result_sha256"`
	OriginVectorsSHA256         string   `json:"baseline_vectors_sha256"`
	WitnessSlot                 string   `json:"equivalence_witness_origin"`
	WitnessStageSeeds           [2]int64 `json:"equivalence_witness_stage_seeds"`
	WitnessBinarySHA256         string   `json:"equivalence_witness_binary_sha256"`
	WitnessResultSHA256         string   `json:"equivalence_witness_result_sha256"`
	WitnessVectorsSHA256        string   `json:"equivalence_witness_vectors_sha256"`
	DecodedVectorDocumentsMatch bool     `json:"decoded_vector_documents_match"`
	CheckpointStatesMatch       bool     `json:"checkpoint_states_match"`
	FastInputRowsMatch          bool     `json:"fast_input_active_rows_match"`
	NoiseInjectionCount         int      `json:"baseline_noise_injection_count"`
}

const (
	fixedSigmaDirectMulReuseTargetSeed    = int64(20261105)
	fixedSigmaDirectMulF9BinarySHA256     = "a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff"
	fixedSigmaDirectMulF9ResultSHA256     = "c3e5172c44688f62d4087149e485f0f2f4f597747dc1c748f0bac2d8e28be94a"
	fixedSigmaDirectMulF13BinarySHA256    = "bde01f61c36f70dc70939eb12d72f3a42ce07a925942904939b54e516cafae17"
	fixedSigmaDirectMulF13ResultSHA256    = "27370f575519e061a73584994d1a30e52acd65d9e85be19c6b87731a3a930f69"
	fixedSigmaDirectMulSharedVectorSHA256 = "b1abc549359dbb3c3b293c9e5de56b1e76c7888fd35e9c973357f62de3d61f4e"
)

type fixedSigmaAblationWorkloadResult struct {
	Name              string                                   `json:"name"`
	CVariant          string                                   `json:"c_input_variant"`
	InputHashes       map[string]string                        `json:"input_message_sha256"`
	WorkloadSHA256    string                                   `json:"workload_sha256"`
	Capacity          perfmeasure.PublicCapacityPlan           `json:"strict_capacity_plan"`
	DirectMulCapacity *perfmeasure.PublicDirectMulCapacityPlan `json:"direct_mul_strict_capacity_plan,omitempty"`
	StandardBinarySHA string                                   `json:"standard_binary_sha256"`
	SeedPairs         []fixedSigmaAblationSeedPairResult       `json:"seed_pairs"`
	TransferScreen    *fixedSigmaAblationTransferScreen        `json:"stagewise_rms_transfer_screen,omitempty"`
}

type fixedSigmaStageRMSComparison struct {
	StandardRMS          float64 `json:"standard_rms"`
	FastRMS              float64 `json:"fast_rms"`
	RelativeDifference   float64 `json:"relative_difference_fraction"`
	Comparable           bool    `json:"comparable"`
	WithinTenPercentBand bool    `json:"within_10_percent_band"`
	Reason               string  `json:"reason,omitempty"`
}

type fixedSigmaArmRMSComparison struct {
	Stage1 fixedSigmaStageRMSComparison `json:"stage1"`
	Stage2 fixedSigmaStageRMSComparison `json:"stage2"`
}

type fixedSigmaSeedRMSComparison struct {
	Stage1Seed int64                                 `json:"stage1_seed"`
	Stage2Seed int64                                 `json:"stage2_seed"`
	Arms       map[string]fixedSigmaArmRMSComparison `json:"arms"`
}

type fixedSigmaAblationTransferScreen struct {
	RelativeDifferenceLimit float64                       `json:"relative_difference_limit_fraction"`
	Criterion               string                        `json:"criterion"`
	SeedPairs               []fixedSigmaSeedRMSComparison `json:"seed_pairs"`
	AllOnOnWithinBand       bool                          `json:"all_on_on_seed_arms_within_band"`
}

type fixedSigmaAblationComparison struct {
	SchemaVersion     string                             `json:"schema_version"`
	GeneratedAt       time.Time                          `json:"generated_at"`
	Status            string                             `json:"status"`
	Classification    string                             `json:"classification"`
	FrozenSigma       float64                            `json:"frozen_sigma_coefficient"`
	BootstrapCalls    int                                `json:"actual_bootstrap_calls"`
	FastChainRuns     int                                `json:"fast_chain_runs"`
	StandardChainRuns int                                `json:"standard_chain_runs"`
	Workloads         []fixedSigmaAblationWorkloadResult `json:"workloads"`
	Limitations       []string                           `json:"limitations"`
}

type loadedFixedSigmaAblationArtifact struct {
	doc        publicNativeDocument
	vectors    publicVectorDocument
	resultSHA  string
	vectorsSHA string
}

func runPublicFixedSigmaAblationCompare(args []string) error {
	flags := flag.NewFlagSet("perfprobe compare-public-ablation", flag.ContinueOnError)
	manifestPath := flags.String("manifest", "", "bounded JSON manifest naming the accepted Standard and four Fast arms")
	outPath := flags.String("out", "", "compact propagation analysis JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *manifestPath == "" || *outPath == "" {
		return errors.New("--manifest and --out are required")
	}
	var manifest fixedSigmaAblationManifest
	if err := readJSONBounded(*manifestPath, &manifest, fixedSigmaAblationManifestMaxBytes); err != nil {
		return fmt.Errorf("read ablation manifest: %w", err)
	}
	directMul := manifest.SchemaVersion == fixedSigmaDirectMulAblationManifestSchema
	if (!directMul && manifest.SchemaVersion != fixedSigmaAblationManifestSchema) || len(manifest.Workloads) == 0 || len(manifest.Workloads) > 2 {
		return errors.New("ablation manifest must use a supported versioned schema and contain one or two workloads")
	}
	result := fixedSigmaAblationComparison{
		SchemaVersion: "fixed-sigma-propagation-ablation-analysis.v1", GeneratedAt: time.Now().UTC(),
		Status: "PASS_VALIDATED_DIAGNOSTIC_ANALYSIS", Classification: "DIAGNOSTIC_ONLY_NO_DISTRIBUTIONAL_OR_SECURITY_CLAIM",
		FrozenSigma: fixedSigmaChainCoefficientSigma,
		Limitations: []string{
			"All four Fast arms are paired by the deterministic zero-secret EncryptNew encoding and identical public operation sequence; coefficient-noise seeds are explicit per stage.",
			"P1 and the first-order product prediction are decoded slot vectors; their difference includes CKKS operation rounding and is not attributed to a single cause.",
			"Slots are correlated coordinates, not independent trials. Tail quantiles are descriptive, and no Gaussianity or Standard/Fast security equivalence is claimed.",
			"The genuine Standard reference is independently encrypted and is used for matched-workload operation-residual context, not as an ablation arm.",
		},
	}
	if directMul {
		if len(manifest.Workloads) != 2 {
			return errors.New("direct-ctct manifest must contain both direct-base and direct-c4 workloads")
		}
		result.SchemaVersion = "fixed-sigma-direct-ctct-ablation-analysis.v1"
		result.Classification = "SCOPED_RMS_TRANSFER_SCREEN_FAIL"
		result.Limitations = append(result.Limitations,
			"The ±10% RMS screen is descriptive and applies to Stage1 and Stage2 of every ON/ON seed arm for both declared direct workloads; it is not statistical or tail equivalence.",
			"Propagation decomposition residual is checked against a float64 roundoff envelope derived from gamma_64; no CKKS primitive is inferred from the residual difference.")
	}
	fastArtifactIDs := make(map[string]struct{})
	standardCount := 0
	decompositionFailed := false
	seenNames, seenVariants := map[string]bool{}, map[string]bool{}
	for _, workload := range manifest.Workloads {
		if workload.Name == "" || seenNames[workload.Name] || seenVariants[workload.CVariant] {
			return errors.New("ablation workload names and C variants must be non-empty and unique")
		}
		if directMul {
			if workload.CVariant != "direct-base" && workload.CVariant != "direct-c4" {
				return fmt.Errorf("workload %s has unsupported direct-ctct variant %q", workload.Name, workload.CVariant)
			}
		} else if workload.CVariant != "canonical" && workload.CVariant != "stronger-c4" {
			return fmt.Errorf("workload %s has unsupported C variant %q", workload.Name, workload.CVariant)
		}
		if len(workload.SeedPairs) == 0 || len(workload.SeedPairs) > 2 || (directMul && len(workload.SeedPairs) != 2) {
			return fmt.Errorf("workload %s must contain one or two seed pairs", workload.Name)
		}
		seenNames[workload.Name], seenVariants[workload.CVariant] = true, true
		standard, err := loadFixedSigmaAblationArtifact(workload.Standard)
		if err != nil {
			return fmt.Errorf("workload %s Standard reference: %w", workload.Name, err)
		}
		if standard.doc.Backend != "standard" || standard.doc.FixedSigmaChain.AblationArm != "OFF/OFF" || standard.doc.FixedSigmaChain.CVariant != workload.CVariant {
			return fmt.Errorf("workload %s Standard reference is not its native OFF/OFF artifact for the declared fixture", workload.Name)
		}
		standardCount++
		workResult := fixedSigmaAblationWorkloadResult{
			Name: workload.Name, CVariant: workload.CVariant,
			InputHashes:    standard.doc.FixedSigmaChain.InputMessageSHA256,
			WorkloadSHA256: standard.doc.WorkloadSHA256, Capacity: standard.doc.Capacity,
			StandardBinarySHA: standard.doc.BinarySHA256,
		}
		if directMul {
			workResult.DirectMulCapacity = standard.doc.DirectMulCapacity
		}
		seenSeeds := map[int64]bool{}
		for pairIndex, seedPair := range workload.SeedPairs {
			if seedPair.Stage1Seed == int64(^uint64(0)>>1) || seenSeeds[seedPair.Stage1Seed] || len(seedPair.Arms) != 4 {
				return fmt.Errorf("workload %s has duplicate/overflowing stage1 seed or incomplete four-arm set", workload.Name)
			}
			if directMul {
				wantSeed := int64(20261101 + 2*pairIndex)
				if workload.CVariant == "direct-c4" {
					wantSeed = int64(20261105 + 2*pairIndex)
				}
				if seedPair.Stage1Seed != wantSeed {
					return fmt.Errorf("workload %s seed pair %d uses stage1 seed %d, want frozen seed %d", workload.Name, pairIndex+1, seedPair.Stage1Seed, wantSeed)
				}
			}
			seenSeeds[seedPair.Stage1Seed] = true
			pairResult, err := analyzeFixedSigmaAblationSeedPair(standard, workload.CVariant, seedPair)
			if err != nil {
				return fmt.Errorf("workload %s seed %d: %w", workload.Name, seedPair.Stage1Seed, err)
			}
			workResult.SeedPairs = append(workResult.SeedPairs, pairResult)
			decompositionFailed = decompositionFailed || pairResult.Propagation.DecompositionToleranceFails != 0
			addFixedSigmaAblationFastArtifactIDs(fastArtifactIDs, pairResult)
		}
		if directMul {
			workResult.TransferScreen = computeFixedSigmaDirectMulTransferScreen(workResult)
		}
		result.Workloads = append(result.Workloads, workResult)
	}
	if directMul {
		seenBase, seenC4, allOnOnPass := false, false, true
		for _, workload := range result.Workloads {
			seenBase = seenBase || workload.CVariant == "direct-base"
			seenC4 = seenC4 || workload.CVariant == "direct-c4"
			allOnOnPass = allOnOnPass && workload.TransferScreen != nil && workload.TransferScreen.AllOnOnWithinBand
		}
		if !seenBase || !seenC4 {
			return errors.New("direct-ctct manifest must include direct-base and direct-c4")
		}
		if allOnOnPass {
			result.Classification = "SCOPED_RMS_TRANSFER_SCREEN_PASS"
		}
		if decompositionFailed {
			result.Status = "FAIL_PROPAGATION_DECOMPOSITION_ROUNDOFF_CHECK"
			result.Classification = "BATCH_BLOCKED_NEEDS_WEB_REVIEW"
		}
	}
	fastCount := len(fastArtifactIDs)
	if standardCount > 2 || fastCount > 16 {
		return fmt.Errorf("manifest exceeds charter execution ceilings: Standard=%d Fast=%d", standardCount, fastCount)
	}
	result.StandardChainRuns, result.FastChainRuns = standardCount, fastCount
	if err := writeExclusiveJSON(*outPath, result); err != nil {
		return err
	}
	fmt.Printf("fixed-sigma ablation workloads=%d standard_chains=%d fast_chains=%d bootstrap_calls=0 result=%s\n", len(result.Workloads), standardCount, fastCount, *outPath)
	return nil
}

func addFixedSigmaAblationFastArtifactIDs(ids map[string]struct{}, pair fixedSigmaAblationSeedPairResult) {
	for arm, hashes := range pair.ArtifactSHA256 {
		if arm == "standard" || hashes["result"] == "" || hashes["vectors"] == "" {
			continue
		}
		ids[hashes["result"]+":"+hashes["vectors"]] = struct{}{}
	}
	if reuse := pair.BaselineReuse; reuse != nil && reuse.WitnessResultSHA256 != "" && reuse.WitnessVectorsSHA256 != "" {
		ids[reuse.WitnessResultSHA256+":"+reuse.WitnessVectorsSHA256] = struct{}{}
	}
}

func computeFixedSigmaDirectMulTransferScreen(workload fixedSigmaAblationWorkloadResult) *fixedSigmaAblationTransferScreen {
	screen := &fixedSigmaAblationTransferScreen{
		RelativeDifferenceLimit: 0.10,
		Criterion:               "For each workload, every ON/ON seed pair must have Stage1 and Stage2 Fast incremental-residual RMS within 10% relative difference of the genuine Standard reference; all other arms are reported descriptively.",
		AllOnOnWithinBand:       len(workload.SeedPairs) == 2,
	}
	compare := func(standardRMS, fastRMS float64) fixedSigmaStageRMSComparison {
		result := fixedSigmaStageRMSComparison{StandardRMS: standardRMS, FastRMS: fastRMS}
		if standardRMS <= 0 || math.IsNaN(standardRMS) || math.IsInf(standardRMS, 0) ||
			fastRMS < 0 || math.IsNaN(fastRMS) || math.IsInf(fastRMS, 0) {
			result.Reason = "RMS must be finite and Standard RMS must be positive"
			return result
		}
		result.Comparable = true
		result.RelativeDifference = math.Abs(fastRMS-standardRMS) / standardRMS
		result.WithinTenPercentBand = result.RelativeDifference <= screen.RelativeDifferenceLimit
		return result
	}
	for _, seedPair := range workload.SeedPairs {
		pair := fixedSigmaSeedRMSComparison{Stage1Seed: seedPair.Stage1Seed, Stage2Seed: seedPair.Stage2Seed, Arms: make(map[string]fixedSigmaArmRMSComparison, 4)}
		for _, arm := range []string{"OFF/OFF", "ON/OFF", "OFF/ON", "ON/ON"} {
			metrics, ok := seedPair.Arms[arm]
			if !ok {
				pair.Arms[arm] = fixedSigmaArmRMSComparison{
					Stage1: fixedSigmaStageRMSComparison{Reason: "arm metrics are missing"},
					Stage2: fixedSigmaStageRMSComparison{Reason: "arm metrics are missing"},
				}
				if arm == "ON/ON" {
					screen.AllOnOnWithinBand = false
				}
				continue
			}
			comparison := fixedSigmaArmRMSComparison{
				Stage1: compare(seedPair.Standard.Stage1Incremental.Statistics.RMS, metrics.Stage1Incremental.Statistics.RMS),
				Stage2: compare(seedPair.Standard.Stage2Incremental.Statistics.RMS, metrics.Stage2Incremental.Statistics.RMS),
			}
			pair.Arms[arm] = comparison
			if arm == "ON/ON" && (!comparison.Stage1.Comparable || !comparison.Stage1.WithinTenPercentBand ||
				!comparison.Stage2.Comparable || !comparison.Stage2.WithinTenPercentBand) {
				screen.AllOnOnWithinBand = false
			}
		}
		screen.SeedPairs = append(screen.SeedPairs, pair)
	}
	return screen
}

func loadFixedSigmaAblationArtifact(ref fixedSigmaAblationArtifactRef) (loadedFixedSigmaAblationArtifact, error) {
	if ref.Result == "" || ref.Vectors == "" {
		return loadedFixedSigmaAblationArtifact{}, errors.New("each artifact needs result and vectors paths")
	}
	var loaded loadedFixedSigmaAblationArtifact
	if err := readJSONBounded(ref.Result, &loaded.doc, fixedSigmaAblationResultMaxBytes); err != nil {
		return loaded, err
	}
	if err := readJSONBounded(ref.Vectors, &loaded.vectors, fixedSigmaAblationVectorsMaxBytes); err != nil {
		return loaded, err
	}
	var err error
	if loaded.resultSHA, err = sha256File(ref.Result); err != nil {
		return loaded, err
	}
	if loaded.vectorsSHA, err = sha256File(ref.Vectors); err != nil {
		return loaded, err
	}
	backend := loaded.doc.Backend
	if backend != "fast" && backend != "standard" {
		return loaded, fmt.Errorf("unexpected backend %q", backend)
	}
	var validationErr error
	if directMulWorkloadIDFromDocument(loaded.doc) {
		validationErr = validatePublicFixedSigmaDirectMulArtifact(loaded.doc, loaded.vectors, backend)
	} else {
		validationErr = validatePublicFixedSigmaChainArtifact(loaded.doc, loaded.vectors, backend)
	}
	if validationErr != nil {
		return loaded, validationErr
	}
	return loaded, nil
}

// readJSONBounded checks artifact metadata before reading and caps the read in
// case a file grows or changes between the metadata check and the read.
func readJSONBounded(path string, dst any, maxBytes int64) error {
	if maxBytes <= 0 {
		return errors.New("JSON artifact size limit must be positive")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("JSON artifact %s is not a regular file", path)
	}
	if info.Size() > maxBytes {
		return fmt.Errorf("JSON artifact %s is %d bytes, exceeding the %d-byte inspection limit", path, info.Size(), maxBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > maxBytes {
		return fmt.Errorf("JSON artifact %s grew beyond the %d-byte inspection limit", path, maxBytes)
	}
	return json.Unmarshal(data, dst)
}

func analyzeFixedSigmaAblationSeedPair(standard loadedFixedSigmaAblationArtifact, cVariant string, seedPair fixedSigmaAblationSeedPair) (fixedSigmaAblationSeedPairResult, error) {
	const arms = "OFF/OFF,ON/OFF,OFF/ON,ON/ON"
	wantedArms := []string{"OFF/OFF", "ON/OFF", "OFF/ON", "ON/ON"}
	fastRuns := make(map[string]loadedFixedSigmaAblationArtifact, len(wantedArms))
	result := fixedSigmaAblationSeedPairResult{
		Stage1Seed: seedPair.Stage1Seed, Stage2Seed: seedPair.Stage1Seed + 1,
		ArtifactSHA256: make(map[string]map[string]string, len(wantedArms)+1),
		Arms:           make(map[string]fixedSigmaAblationArmAnalysis, len(wantedArms)),
		Standard:       summarizeFixedSigmaAblationArtifact(standard),
	}
	result.ArtifactSHA256["standard"] = map[string]string{"result": standard.resultSHA, "vectors": standard.vectorsSHA}
	var reusedBaseline *loadedFixedSigmaAblationArtifact
	if seedPair.BaselineReuse != nil {
		if !directMulWorkloadIDFromDocument(standard.doc) || cVariant != "direct-c4" || seedPair.Stage1Seed != fixedSigmaDirectMulReuseTargetSeed {
			return result, errors.New("OFF/OFF baseline reuse is authorized only for the direct-c4 first frozen seed pair")
		}
		baselineRef, ok := seedPair.Arms["OFF/OFF"]
		if !ok {
			return result, errors.New("reusable OFF/OFF baseline artifact is missing")
		}
		baseline, err := loadFixedSigmaAblationArtifact(baselineRef)
		if err != nil {
			return result, fmt.Errorf("reusable OFF/OFF baseline: %w", err)
		}
		witness, err := loadFixedSigmaAblationArtifact(seedPair.BaselineReuse.Witness)
		if err != nil {
			return result, fmt.Errorf("baseline equivalence witness: %w", err)
		}
		reuseEvidence, err := validateFixedSigmaDirectMulF9F13BaselineReuse(
			baseline, witness, cVariant, seedPair.Stage1Seed, *seedPair.BaselineReuse,
		)
		if err != nil {
			return result, err
		}
		reusedBaseline = &baseline
		result.BaselineOrigin = reuseEvidence.OriginSlot
		result.BaselineReuse = &reuseEvidence
	}
	for _, arm := range wantedArms {
		ref, ok := seedPair.Arms[arm]
		if !ok {
			return result, fmt.Errorf("missing arm %s (wanted %s)", arm, arms)
		}
		var loaded loadedFixedSigmaAblationArtifact
		var err error
		if arm == "OFF/OFF" && reusedBaseline != nil {
			loaded = *reusedBaseline
		} else {
			loaded, err = loadFixedSigmaAblationArtifact(ref)
			if err != nil {
				return result, fmt.Errorf("arm %s: %w", arm, err)
			}
		}
		wantSeeds := [2]int64{seedPair.Stage1Seed, seedPair.Stage1Seed + 1}
		if loaded.doc.Backend != "fast" || loaded.doc.FixedSigmaChain.AblationArm != arm || loaded.doc.FixedSigmaChain.CVariant != cVariant ||
			(arm == "OFF/OFF" && reusedBaseline == nil && loaded.doc.FixedSigmaChain.StageSeeds != wantSeeds) ||
			(arm != "OFF/OFF" && loaded.doc.FixedSigmaChain.StageSeeds != wantSeeds) {
			return result, fmt.Errorf("arm %s does not match Fast, C variant, and declared seed pair", arm)
		}
		var pairErr error
		if directMulWorkloadIDFromDocument(standard.doc) {
			pairErr = validatePublicDirectMulPairArtifacts(standard.doc, loaded.doc, standard.vectors, loaded.vectors)
		} else {
			pairErr = validatePublicFixedSigmaChainPairArtifacts(standard.doc, loaded.doc, standard.vectors, loaded.vectors)
		}
		if pairErr != nil {
			return result, fmt.Errorf("arm %s Standard/workload pair: %w", arm, pairErr)
		}
		if !publicEnvironmentsMatch(standard.doc, loaded.doc) {
			return result, fmt.Errorf("arm %s does not match the Standard reference execution environment", arm)
		}
		if len(fastRuns) != 0 {
			previous := fastRuns[wantedArms[0]]
			if loaded.doc.PrimaryCommit != previous.doc.PrimaryCommit || loaded.doc.PrimarySourceSHA256 != previous.doc.PrimarySourceSHA256 ||
				loaded.doc.BackendCommit != previous.doc.BackendCommit || loaded.doc.BinarySHA256 != previous.doc.BinarySHA256 ||
				loaded.doc.WorkloadSHA256 != previous.doc.WorkloadSHA256 || loaded.doc.ConfigSHA256 != previous.doc.ConfigSHA256 ||
				loaded.doc.QPSHA256 != previous.doc.QPSHA256 || loaded.doc.InputSHA256 != previous.doc.InputSHA256 ||
				!reflect.DeepEqual(loaded.doc.Parameters, previous.doc.Parameters) {
				return result, fmt.Errorf("arm %s differs in source, binary, workload or effective parameters from the shared seed-pair arms", arm)
			}
			if !publicEnvironmentsMatch(loaded.doc, previous.doc) {
				return result, fmt.Errorf("arm %s does not match the shared seed-pair execution environment", arm)
			}
			if !reflect.DeepEqual(loaded.doc.FixedSigmaChain.FastInputRowsSHA256, previous.doc.FixedSigmaChain.FastInputRowsSHA256) {
				return result, fmt.Errorf("arm %s does not reuse identical Fast input ciphertext active rows", arm)
			}
		}
		fastRuns[arm] = loaded
		result.Arms[arm] = summarizeFixedSigmaAblationArtifact(loaded)
		result.ArtifactSHA256[arm] = map[string]string{"result": loaded.resultSHA, "vectors": loaded.vectorsSHA}
	}
	result.FastBinarySHA = fastRuns["OFF/OFF"].doc.BinarySHA256
	if err := validateAblationFastInputs(fastRuns); err != nil {
		return result, err
	}
	propagation, err := computeFixedSigmaAblationPropagation(fastRuns)
	if err != nil {
		return result, err
	}
	result.Propagation = propagation
	second, err := computeFixedSigmaAblationSecondCheck(fastRuns)
	if err != nil {
		return result, err
	}
	result.SecondCheck = second
	return result, nil
}

func validateFixedSigmaDirectMulF9F13BaselineReuse(
	baseline, witness loadedFixedSigmaAblationArtifact,
	cVariant string,
	targetStage1Seed int64,
	reuse fixedSigmaAblationBaselineReuse,
) (fixedSigmaAblationBaselineReuseEvidence, error) {
	const (
		f9Stage1Seed  = int64(20261105)
		f9Stage2Seed  = int64(20261106)
		f13Stage1Seed = int64(20261107)
		f13Stage2Seed = int64(20261108)
	)
	if cVariant != "direct-c4" || targetStage1Seed != fixedSigmaDirectMulReuseTargetSeed ||
		reuse.OriginSlot != "F13" || reuse.WitnessSlot != "F9" ||
		reuse.OriginStage1Seed != f13Stage1Seed || reuse.OriginStage2Seed != f13Stage2Seed {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("baseline reuse metadata is outside the single authorized F9/F13 direct-c4 seed pair")
	}
	if baseline.doc.Backend != "fast" || witness.doc.Backend != "fast" ||
		baseline.doc.FixedSigmaChain.AblationArm != "OFF/OFF" || witness.doc.FixedSigmaChain.AblationArm != "OFF/OFF" ||
		baseline.doc.FixedSigmaChain.CVariant != cVariant || witness.doc.FixedSigmaChain.CVariant != cVariant ||
		baseline.doc.FixedSigmaChain.StageSeeds != [2]int64{f13Stage1Seed, f13Stage2Seed} ||
		witness.doc.FixedSigmaChain.StageSeeds != [2]int64{f9Stage1Seed, f9Stage2Seed} ||
		baseline.doc.FixedSigmaChain.LogicalNoiseInjections != 0 || witness.doc.FixedSigmaChain.LogicalNoiseInjections != 0 {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("F9/F13 reusable baselines must be OFF/OFF, zero-injection captures with their original source seeds")
	}
	if baseline.doc.BinarySHA256 != fixedSigmaDirectMulF13BinarySHA256 || witness.doc.BinarySHA256 != fixedSigmaDirectMulF9BinarySHA256 ||
		baseline.resultSHA != fixedSigmaDirectMulF13ResultSHA256 || witness.resultSHA != fixedSigmaDirectMulF9ResultSHA256 ||
		baseline.vectorsSHA != fixedSigmaDirectMulSharedVectorSHA256 || witness.vectorsSHA != fixedSigmaDirectMulSharedVectorSHA256 {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("F9/F13 artifact or binary SHA-256 differs from the reviewed deterministic reuse evidence")
	}
	if baseline.doc.Mode != witness.doc.Mode || baseline.doc.SchemaVersion != witness.doc.SchemaVersion ||
		baseline.doc.Status != witness.doc.Status || baseline.doc.PrimaryCommit != witness.doc.PrimaryCommit ||
		baseline.doc.PrimarySourceSHA256 != witness.doc.PrimarySourceSHA256 || baseline.doc.BackendCommit != witness.doc.BackendCommit ||
		baseline.doc.BackendRef != witness.doc.BackendRef || baseline.doc.BackendCheckoutRef != witness.doc.BackendCheckoutRef ||
		!baseline.doc.PrimaryClean || !witness.doc.PrimaryClean || !baseline.doc.BackendClean || !witness.doc.BackendClean ||
		baseline.doc.ConfigSHA256 != witness.doc.ConfigSHA256 || baseline.doc.QPSHA256 != witness.doc.QPSHA256 ||
		baseline.doc.InputSHA256 != witness.doc.InputSHA256 || baseline.doc.WorkloadSHA256 != witness.doc.WorkloadSHA256 ||
		!reflect.DeepEqual(baseline.doc.Parameters, witness.doc.Parameters) ||
		!reflect.DeepEqual(baseline.doc.DirectMulCapacity, witness.doc.DirectMulCapacity) ||
		!reflect.DeepEqual(baseline.doc.FixedSigmaChain.InputMessageSHA256, witness.doc.FixedSigmaChain.InputMessageSHA256) ||
		!publicEnvironmentsMatch(baseline.doc, witness.doc) {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("F9/F13 baseline source, inputs, parameters, environment or capacity evidence differs")
	}
	if len(baseline.doc.FixedSigmaChain.FastInputRowsSHA256) != 24 ||
		!reflect.DeepEqual(baseline.doc.FixedSigmaChain.FastInputRowsSHA256, witness.doc.FixedSigmaChain.FastInputRowsSHA256) {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("F9/F13 baseline input active-row hashes are missing or differ")
	}
	if !reflect.DeepEqual(baseline.doc.Checkpoints, witness.doc.Checkpoints) {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("F9/F13 checkpoint Level/Scale/domain/status/active-row evidence differs")
	}
	if !reflect.DeepEqual(baseline.vectors, witness.vectors) || baseline.vectorsSHA != witness.vectorsSHA {
		return fixedSigmaAblationBaselineReuseEvidence{}, errors.New("F9/F13 complete decoded vector documents or their SHA-256 differ")
	}
	return fixedSigmaAblationBaselineReuseEvidence{
		OriginSlot: "F13", OriginStageSeeds: baseline.doc.FixedSigmaChain.StageSeeds,
		RequestedStageSeeds: [2]int64{targetStage1Seed, targetStage1Seed + 1},
		OriginBinarySHA256:  baseline.doc.BinarySHA256,
		OriginResultSHA256:  baseline.resultSHA, OriginVectorsSHA256: baseline.vectorsSHA,
		WitnessSlot: "F9", WitnessStageSeeds: witness.doc.FixedSigmaChain.StageSeeds,
		WitnessBinarySHA256: witness.doc.BinarySHA256,
		WitnessResultSHA256: witness.resultSHA, WitnessVectorsSHA256: witness.vectorsSHA,
		DecodedVectorDocumentsMatch: true, CheckpointStatesMatch: true,
		FastInputRowsMatch: true, NoiseInjectionCount: 0,
	}, nil
}

func summarizeFixedSigmaAblationArtifact(artifact loadedFixedSigmaAblationArtifact) fixedSigmaAblationArmAnalysis {
	stage := artifact.vectors.FixedSigmaChain.Stages
	return fixedSigmaAblationArmAnalysis{
		Stage1Incremental: summarizeFixedSigmaVector(decodeComplexValues(stage["stage1"].IncrementalResidual)),
		Stage1Total:       summarizeFixedSigmaVector(decodeComplexValues(stage["stage1"].TotalError)),
		Stage2Incremental: summarizeFixedSigmaVector(decodeComplexValues(stage["stage2"].IncrementalResidual)),
		Stage2Total:       summarizeFixedSigmaVector(decodeComplexValues(stage["stage2"].TotalError)),
	}
}

func summarizeFixedSigmaVector(values []complex128) fixedSigmaVectorAnalysis {
	return fixedSigmaVectorAnalysis{
		Statistics: numericalmetrics.SummarizeComplexVector(values),
		Tails:      numericalmetrics.SummarizeComplexMagnitudeTails(values),
	}
}

func validateAblationFastInputs(runs map[string]loadedFixedSigmaAblationArtifact) error {
	baseline := runs["OFF/OFF"]
	checkpoints := []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_ab_level5", "encrypt_c_q4_level4"}
	if directMulWorkloadIDFromDocument(baseline.doc) {
		checkpoints = []string{"encrypt_a_level5", "encrypt_b_q5_level5", "encrypt_c_q4_level4"}
	}
	for _, arm := range []string{"ON/OFF", "OFF/ON", "ON/ON"} {
		candidate := runs[arm]
		for _, checkpoint := range checkpoints {
			baseValues, baseOK := baseline.vectors.Checkpoints[checkpoint]
			armValues, armOK := candidate.vectors.Checkpoints[checkpoint]
			if !baseOK || !armOK || !complexVectorsEqual(decodeComplexValues(baseValues), decodeComplexValues(armValues)) {
				return fmt.Errorf("arm %s changed shared decoded input checkpoint %s", arm, checkpoint)
			}
		}
		if !reflect.DeepEqual(baseline.doc.FixedSigmaChain.InputMessageSHA256, candidate.doc.FixedSigmaChain.InputMessageSHA256) {
			return fmt.Errorf("arm %s changed original plaintext input hashes", arm)
		}
	}
	return nil
}

func computeFixedSigmaAblationPropagation(runs map[string]loadedFixedSigmaAblationArtifact) (fixedSigmaAblationPropagation, error) {
	base := runs["OFF/OFF"].vectors
	onOff := runs["ON/OFF"].vectors
	stage1Delta := subtractComplexVectors(
		decodeComplexValues(onOff.Checkpoints["chain_stage1_output_level4"]),
		decodeComplexValues(base.Checkpoints["chain_stage1_output_level4"]),
	)
	helperDelta := decodeComplexValues(onOff.FixedSigmaChain.Stages["stage1"].DecodedNoiseDelta)
	stage2P1 := subtractComplexVectors(
		decodeComplexValues(onOff.Checkpoints["chain_stage2_output_level3"]),
		decodeComplexValues(base.Checkpoints["chain_stage2_output_level3"]),
	)
	cDecoded := decodeComplexValues(base.Checkpoints["encrypt_c_q4_level4"])
	stage1On := decodeComplexValues(onOff.Checkpoints["chain_stage1_output_level4"])
	stage1Off := decodeComplexValues(base.Checkpoints["chain_stage1_output_level4"])
	stage2On := decodeComplexValues(onOff.Checkpoints["chain_stage2_output_level3"])
	stage2Off := decodeComplexValues(base.Checkpoints["chain_stage2_output_level3"])
	prediction := multiplyComplexVectors(stage1Delta, cDecoded)
	firstOrderResidual := subtractComplexVectors(stage2P1, prediction)
	r2a := subtractComplexVectors(stage2On, multiplyComplexVectors(stage1On, cDecoded))
	r2b := subtractComplexVectors(stage2Off, multiplyComplexVectors(stage1Off, cDecoded))
	decompositionRight := addComplexVectors(prediction, subtractComplexVectors(r2a, r2b))
	decompositionResidual := subtractComplexVectors(stage2P1, decompositionRight)
	if len(stage1Delta) == 0 || len(stage2P1) != len(stage1Delta) || len(prediction) != len(stage1Delta) ||
		len(r2a) != len(stage1Delta) || len(r2b) != len(stage1Delta) || len(decompositionResidual) != len(stage1Delta) ||
		!allFinite(stage1Delta) || !allFinite(stage2P1) || !allFinite(prediction) || !allFinite(r2a) || !allFinite(r2b) || !allFinite(decompositionResidual) {
		return fixedSigmaAblationPropagation{}, errors.New("decoded propagation vectors are empty, mismatched or non-finite")
	}
	matchedDelta := complexVectorsEqual(stage1Delta, helperDelta)
	if !matchedDelta {
		return fixedSigmaAblationPropagation{}, fmt.Errorf("ON/OFF decoded stage1 difference %v does not exactly match the helper's recorded injected delta %v", stage1Delta, helperDelta)
	}
	stage2OperandMatched := true
	fastActiveRowsMatched := true
	for _, arm := range []string{"ON/OFF", "OFF/ON", "ON/ON"} {
		candidate := runs[arm]
		if !complexVectorsEqual(cDecoded, decodeComplexValues(candidate.vectors.Checkpoints["encrypt_c_q4_level4"])) ||
			!reflect.DeepEqual(runs["OFF/OFF"].doc.FixedSigmaChain.FastInputRowsSHA256, candidate.doc.FixedSigmaChain.FastInputRowsSHA256) {
			stage2OperandMatched, fastActiveRowsMatched = false, false
		}
	}
	if !stage2OperandMatched || !fastActiveRowsMatched {
		return fixedSigmaAblationPropagation{}, errors.New("Stage2 C input vector or authoritative Fast input rows differ across paired arms")
	}
	const unitRoundoff = 0x1p-52
	const operations = 64.0
	gamma64 := operations * unitRoundoff / (1 - operations*unitRoundoff)
	maxResidual, maxTolerance, toleranceFailures := 0.0, 0.0, 0
	for index := range decompositionResidual {
		residualMagnitude := cmplx.Abs(decompositionResidual[index])
		bound := gamma64 * (cmplx.Abs(stage2P1[index]) + cmplx.Abs(prediction[index]) + cmplx.Abs(r2a[index]) + cmplx.Abs(r2b[index]))
		if residualMagnitude > maxResidual {
			maxResidual = residualMagnitude
		}
		if bound > maxTolerance {
			maxTolerance = bound
		}
		if residualMagnitude > bound {
			toleranceFailures++
		}
	}
	return fixedSigmaAblationPropagation{
		Stage1InjectedDelta:          summarizeFixedSigmaVector(stage1Delta),
		Stage2PropagatedP1:           summarizeFixedSigmaVector(stage2P1),
		Stage2IncrementalONOFF:       summarizeFixedSigmaVector(r2a),
		Stage2IncrementalOFFOFF:      summarizeFixedSigmaVector(r2b),
		FirstOrderPrediction:         summarizeFixedSigmaVector(prediction),
		FirstOrderResidual:           summarizeFixedSigmaVector(firstOrderResidual),
		DecompositionResidual:        summarizeFixedSigmaVector(decompositionResidual),
		P1PredictionCovariance:       numericalmetrics.CompareComplexCovariance(stage2P1, prediction),
		Stage2OperandDecodedSHA256:   perfmeasure.Fingerprint(cDecoded),
		Stage2OperandMatchedAllArms:  stage2OperandMatched,
		FastActiveInputRowsMatched:   fastActiveRowsMatched,
		DecompositionTolerance:       "gamma_64=64u/(1-64u), u=2^-52; per-slot tolerance=gamma_64*(|P1|+|d1*C|+|r2a|+|r2b|)",
		MaxDecompositionResidual:     maxResidual,
		MaxFloatingPointTolerance:    maxTolerance,
		DecompositionToleranceFails:  toleranceFailures,
		InjectedDeltaEvidenceMatched: matchedDelta,
	}, nil
}

func addComplexVectors(x, y []complex128) []complex128 {
	if len(x) == 0 || len(x) != len(y) {
		return nil
	}
	out := make([]complex128, len(x))
	for index := range x {
		out[index] = x[index] + y[index]
	}
	return out
}

func computeFixedSigmaAblationSecondCheck(runs map[string]loadedFixedSigmaAblationArtifact) (fixedSigmaAblationSecondCheck, error) {
	offOn := runs["OFF/ON"].vectors
	onOn := runs["ON/ON"].vectors
	onOnMinusOffOn := subtractComplexVectors(
		decodeComplexValues(onOn.Checkpoints["chain_stage2_output_level3"]),
		decodeComplexValues(offOn.Checkpoints["chain_stage2_output_level3"]),
	)
	p1 := subtractComplexVectors(
		decodeComplexValues(runs["ON/OFF"].vectors.Checkpoints["chain_stage2_output_level3"]),
		decodeComplexValues(runs["OFF/OFF"].vectors.Checkpoints["chain_stage2_output_level3"]),
	)
	difference := subtractComplexVectors(onOnMinusOffOn, p1)
	if len(difference) == 0 || !allFinite(onOnMinusOffOn) || !allFinite(difference) {
		return fixedSigmaAblationSecondCheck{}, errors.New("ON/ON context comparison is empty or non-finite")
	}
	return fixedSigmaAblationSecondCheck{
		ONONMinusOFFON:   summarizeFixedSigmaVector(onOnMinusOffOn),
		DifferenceFromP1: summarizeFixedSigmaVector(difference),
		CovarianceWithP1: numericalmetrics.CompareComplexCovariance(onOnMinusOffOn, p1),
	}, nil
}
