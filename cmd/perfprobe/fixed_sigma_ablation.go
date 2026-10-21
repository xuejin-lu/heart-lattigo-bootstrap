package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const fixedSigmaAblationManifestSchema = "fixed-sigma-propagation-ablation-manifest.v1"

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
	Stage1Seed int64                                    `json:"stage1_seed"`
	Arms       map[string]fixedSigmaAblationArtifactRef `json:"arms"`
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
	FirstOrderPrediction         fixedSigmaVectorAnalysis           `json:"first_order_prediction_delta1_times_c_decoded"`
	FirstOrderResidual           fixedSigmaVectorAnalysis           `json:"p1_minus_first_order_prediction"`
	P1PredictionCovariance       numericalmetrics.ComplexCovariance `json:"p1_prediction_covariance"`
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
	ArtifactSHA256 map[string]map[string]string             `json:"artifact_sha256"`
	Arms           map[string]fixedSigmaAblationArmAnalysis `json:"fast_arm_operation_metrics"`
	Standard       fixedSigmaAblationArmAnalysis            `json:"standard_off_off_operation_metrics"`
	Propagation    fixedSigmaAblationPropagation            `json:"stage1_to_stage2_propagation"`
	SecondCheck    fixedSigmaAblationSecondCheck            `json:"on_on_context_check"`
}

type fixedSigmaAblationWorkloadResult struct {
	Name              string                             `json:"name"`
	CVariant          string                             `json:"c_input_variant"`
	InputHashes       map[string]string                  `json:"input_message_sha256"`
	WorkloadSHA256    string                             `json:"workload_sha256"`
	Capacity          perfmeasure.PublicCapacityPlan     `json:"strict_capacity_plan"`
	StandardBinarySHA string                             `json:"standard_binary_sha256"`
	SeedPairs         []fixedSigmaAblationSeedPairResult `json:"seed_pairs"`
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
	if manifest.SchemaVersion != fixedSigmaAblationManifestSchema || len(manifest.Workloads) == 0 || len(manifest.Workloads) > 2 {
		return errors.New("ablation manifest must use the v1 schema and contain one or two workloads")
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
	fastCount, standardCount := 0, 0
	seenNames, seenVariants := map[string]bool{}, map[string]bool{}
	for _, workload := range manifest.Workloads {
		if workload.Name == "" || seenNames[workload.Name] || seenVariants[workload.CVariant] {
			return errors.New("ablation workload names and C variants must be non-empty and unique")
		}
		if workload.CVariant != "canonical" && workload.CVariant != "stronger-c4" {
			return fmt.Errorf("workload %s has unsupported C variant %q", workload.Name, workload.CVariant)
		}
		if len(workload.SeedPairs) == 0 || len(workload.SeedPairs) > 2 {
			return fmt.Errorf("workload %s must contain one or two seed pairs", workload.Name)
		}
		seenNames[workload.Name], seenVariants[workload.CVariant] = true, true
		standard, err := loadFixedSigmaAblationArtifact(workload.Standard)
		if err != nil {
			return fmt.Errorf("workload %s Standard reference: %w", workload.Name, err)
		}
		if standard.doc.Backend != "standard" || standard.doc.FixedSigmaChain.AblationArm != "OFF/OFF" || standard.doc.FixedSigmaChain.CVariant != workload.CVariant {
			return fmt.Errorf("workload %s Standard reference is not its native OFF/OFF v4 artifact", workload.Name)
		}
		standardCount++
		workResult := fixedSigmaAblationWorkloadResult{
			Name: workload.Name, CVariant: workload.CVariant,
			InputHashes:    standard.doc.FixedSigmaChain.InputMessageSHA256,
			WorkloadSHA256: standard.doc.WorkloadSHA256, Capacity: standard.doc.Capacity,
			StandardBinarySHA: standard.doc.BinarySHA256,
		}
		seenSeeds := map[int64]bool{}
		for _, seedPair := range workload.SeedPairs {
			if seedPair.Stage1Seed == int64(^uint64(0)>>1) || seenSeeds[seedPair.Stage1Seed] || len(seedPair.Arms) != 4 {
				return fmt.Errorf("workload %s has duplicate/overflowing stage1 seed or incomplete four-arm set", workload.Name)
			}
			seenSeeds[seedPair.Stage1Seed] = true
			pairResult, err := analyzeFixedSigmaAblationSeedPair(standard, workload.CVariant, seedPair)
			if err != nil {
				return fmt.Errorf("workload %s seed %d: %w", workload.Name, seedPair.Stage1Seed, err)
			}
			workResult.SeedPairs = append(workResult.SeedPairs, pairResult)
			fastCount += len(seedPair.Arms)
		}
		result.Workloads = append(result.Workloads, workResult)
	}
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
	if err := validatePublicFixedSigmaChainArtifact(loaded.doc, loaded.vectors, backend); err != nil {
		return loaded, err
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
	for _, arm := range wantedArms {
		ref, ok := seedPair.Arms[arm]
		if !ok {
			return result, fmt.Errorf("missing arm %s (wanted %s)", arm, arms)
		}
		loaded, err := loadFixedSigmaAblationArtifact(ref)
		if err != nil {
			return result, fmt.Errorf("arm %s: %w", arm, err)
		}
		if loaded.doc.Backend != "fast" || loaded.doc.FixedSigmaChain.AblationArm != arm || loaded.doc.FixedSigmaChain.CVariant != cVariant ||
			loaded.doc.FixedSigmaChain.StageSeeds != [2]int64{seedPair.Stage1Seed, seedPair.Stage1Seed + 1} {
			return result, fmt.Errorf("arm %s does not match Fast, C variant, and declared seed pair", arm)
		}
		if err := validatePublicFixedSigmaChainPairArtifacts(standard.doc, loaded.doc, standard.vectors, loaded.vectors); err != nil {
			return result, fmt.Errorf("arm %s Standard/workload pair: %w", arm, err)
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
	for _, arm := range []string{"ON/OFF", "OFF/ON", "ON/ON"} {
		candidate := runs[arm]
		for _, checkpoint := range []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_ab_level5", "encrypt_c_q4_level4"} {
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
	prediction := multiplyComplexVectors(stage1Delta, cDecoded)
	firstOrderResidual := subtractComplexVectors(stage2P1, prediction)
	if len(stage1Delta) == 0 || len(stage2P1) != len(stage1Delta) || len(prediction) != len(stage1Delta) || !allFinite(stage1Delta) || !allFinite(stage2P1) || !allFinite(prediction) {
		return fixedSigmaAblationPropagation{}, errors.New("decoded propagation vectors are empty, mismatched or non-finite")
	}
	matchedDelta := complexVectorsEqual(stage1Delta, helperDelta)
	if !matchedDelta {
		return fixedSigmaAblationPropagation{}, fmt.Errorf("ON/OFF decoded stage1 difference %v does not exactly match the helper's recorded injected delta %v", stage1Delta, helperDelta)
	}
	return fixedSigmaAblationPropagation{
		Stage1InjectedDelta:          summarizeFixedSigmaVector(stage1Delta),
		Stage2PropagatedP1:           summarizeFixedSigmaVector(stage2P1),
		FirstOrderPrediction:         summarizeFixedSigmaVector(prediction),
		FirstOrderResidual:           summarizeFixedSigmaVector(firstOrderResidual),
		P1PredictionCovariance:       numericalmetrics.CompareComplexCovariance(stage2P1, prediction),
		InjectedDeltaEvidenceMatched: matchedDelta,
	}, nil
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
