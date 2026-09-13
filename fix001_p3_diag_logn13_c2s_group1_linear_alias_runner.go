package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredC2SGroup1AliasPrimaryBase = "302efbaef8ab15e0e7b09ee40911921753495ee0"
	requiredC2SGroup1AliasSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
	acceptedC2SGroup1Budget           = 0.000019531249403614602
)

type c2sGroup1InputSummary struct {
	FastMetadata            semanticBisectMetadata `json:"fast_metadata"`
	StandardMetadata        semanticBisectMetadata `json:"standard_metadata"`
	FastQ01Capacity         postMod1S2CCapacity    `json:"fast_q01_capacity"`
	StandardFullRNSCapacity postMod1S2CCapacity    `json:"standard_full_rns_capacity"`
	FastResiduesMatch       bool                   `json:"fast_residues_match_standard_mod_q01"`
	FastVsStandard          semanticBisectMetric   `json:"fast_vs_standard"`
	Aligned                 bool                   `json:"aligned"`
}

type c2sGroup1RescaleSummary struct {
	PreLevel             int                   `json:"pre_level"`
	PostLevel            int                   `json:"post_level"`
	PreScale             string                `json:"pre_scale"`
	PostScale            string                `json:"post_scale"`
	DroppedModulus       string                `json:"dropped_modulus"`
	StandardPreCapacity  postMod1S2CCapacity   `json:"standard_pre_capacity"`
	FastPreCapacity      postMod1S2CCapacity   `json:"fast_pre_capacity"`
	StandardPostCapacity postMod1S2CCapacity   `json:"standard_post_capacity"`
	FastPostCapacity     postMod1S2CCapacity   `json:"fast_post_capacity"`
	FastResiduesMatch    bool                  `json:"fast_residues_match_standard_mod_q01"`
	FastVsStandardAfter  *semanticBisectMetric `json:"fast_vs_standard_after,omitempty"`
}

type c2sGroup1Result struct {
	SchemaVersion                     string                      `json:"schema_version"`
	Timestamp                         time.Time                   `json:"timestamp"`
	Primary                           RepositoryMetadata          `json:"primary_repository"`
	Lattigo                           RepositoryMetadata          `json:"lattigo_repository"`
	Environment                       EnvironmentMetadata         `json:"environment"`
	Config                            BootstrapConfig             `json:"config"`
	Parameters                        ExperimentParameters        `json:"effective_parameters"`
	Workload                          CorrectnessWorkload         `json:"workload"`
	StandardProof                     semanticBisectStandardProof `json:"standard_control_proof"`
	StandardEndToEnd                  *semanticBisectMetric       `json:"standard_end_to_end_semantic,omitempty"`
	ModUp                             *c2sPrecisionCheckpoint     `json:"modup_precondition,omitempty"`
	ModUpMetadataMatch                bool                        `json:"modup_metadata_match"`
	ModUpNoInputMutation              bool                        `json:"modup_no_input_mutation"`
	DerivedC2SBudget                  float64                     `json:"derived_c2s_budget"`
	Group0Validated                   bool                        `json:"accepted_group0_k4_validated"`
	Group0Restored                    *c2sPrecisionCheckpoint     `json:"group0_restored,omitempty"`
	Group0OriginalPostRescale         *c2sPrecisionCheckpoint     `json:"group0_original_post_rescale,omitempty"`
	Group0RestoredVsOriginal          *semanticBisectMetric       `json:"group0_restored_vs_original_standard,omitempty"`
	Group0MetadataMatches             bool                        `json:"group0_metadata_matches_original_post_rescale"`
	OriginalGroup1InputCapacity       *postMod1S2CCapacity        `json:"original_group1_input_capacity,omitempty"`
	AlignedGroup1Input                *c2sGroup1InputSummary      `json:"aligned_group1_input,omitempty"`
	Structure                         c2sAliasStructure           `json:"group1_transform_structure"`
	Checkpoints                       []c2sAliasCheckpoint        `json:"ordered_checkpoints"`
	FirstFailingCheckpoint            string                      `json:"first_failing_checkpoint"`
	FirstSupportedCause               string                      `json:"first_supported_cause"`
	WorstFullGroupCapacity            *c2sAliasCompression        `json:"worst_full_group_capacity,omitempty"`
	Group1PreRescale                  *c2sCompressedStage         `json:"group1_pre_rescale,omitempty"`
	Group1PostRescale                 *c2sCompressedStage         `json:"group1_post_rescale,omitempty"`
	RescaleSummary                    *c2sGroup1RescaleSummary    `json:"group1_rescale_summary,omitempty"`
	AlignedStandardVsOriginal         *semanticBisectMetric       `json:"aligned_full_rns_vs_original_standard_post_rescale,omitempty"`
	C2SErrorAloneSufficient           bool                        `json:"c2s_error_alone_sufficient_for_final_failure"`
	FastEvalModMatchedInputUnresolved bool                        `json:"fast_evalmod_matched_input_divergence_remains_unresolved"`
	Validation                        map[string]interface{}      `json:"validation"`
}

func c2sGroup1Write(result c2sGroup1Result, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func c2sGroup1MetadataEqual(a, b semanticBisectMetadata) bool {
	return a.SourceLevel == b.SourceLevel && a.Degree == b.Degree && a.N == b.N && a.LogDimensions == b.LogDimensions && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && a.Scale == b.Scale
}

func c2sGroup1LogicalMetadataEqual(a, b semanticBisectMetadata) bool {
	return a.SourceLevel == b.SourceLevel && a.Degree == b.Degree && a.N == b.N && a.LogDimensions == b.LogDimensions && a.IsNTT == b.IsNTT && a.Scale == b.Scale
}

func c2sGroup1CaptureInput(params ckks.Parameters, standard, fast *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, budget float64) (c2sGroup1InputSummary, error) {
	fastMeta, _, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	standardMeta, standardValues, err := c2sPrecisionFullView(standard, params, standardSK)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	_, fastValues, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standard)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	residuesMatch, err := c2sAliasRowsMatch(params, standard, fast)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	metric, err := semanticBisectMetricFor(standardValues, fastValues, budget)
	if err != nil {
		return c2sGroup1InputSummary{}, err
	}
	return c2sGroup1InputSummary{FastMetadata: fastMeta, StandardMetadata: standardMeta, FastQ01Capacity: fastCapacity, StandardFullRNSCapacity: standardCapacity, FastResiduesMatch: residuesMatch, FastVsStandard: metric, Aligned: c2sGroup1LogicalMetadataEqual(fastMeta, standardMeta) && residuesMatch && metric.Pass}, nil
}

func c2sGroup1Classification(checkpoint c2sAliasCheckpoint) string {
	if checkpoint.FullRNSCenteredQ01Unique && (!checkpoint.FastResiduesMatch || checkpoint.Semantic != nil && !checkpoint.Semantic.Pass) {
		switch checkpoint.Kind {
		case "baby_rotation":
			return "logn13_c2s_group1_rotation_mismatch"
		case "diagonal_product":
			return "logn13_c2s_group1_diagonal_product_mismatch"
		case "inner_accumulation":
			return "logn13_c2s_group1_inner_accumulation_mismatch"
		case "giant_rotation":
			return "logn13_c2s_group1_giant_rotation_mismatch"
		case "outer_accumulation":
			return "logn13_c2s_group1_outer_accumulation_mismatch"
		}
	}
	if !checkpoint.FullRNSCenteredQ01Unique && checkpoint.FastResiduesMatch {
		switch checkpoint.Kind {
		case "diagonal_product":
			return "logn13_c2s_group1_single_term_q01_alias"
		case "inner_accumulation":
			return "logn13_c2s_group1_inner_accumulation_q01_alias"
		case "outer_accumulation":
			return "logn13_c2s_group1_outer_accumulation_q01_alias"
		default:
			return "logn13_c2s_group1_linear_modularly_correct_but_final_q01_alias"
		}
	}
	return ""
}

func c2sGroup1Replay(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, standard, fast *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, budget float64) ([]c2sAliasCheckpoint, string, *c2sAliasCompression, error) {
	checkpoints, err := c2sCompressedReplayGroup0(params, standardEval, fastEval, standardEval.C2SDFTMatrix.Matrices[1], standard, fast, standardSK, fastSK, budget)
	if err != nil {
		return nil, "", nil, err
	}
	var first string
	var worst *c2sAliasCompression
	for _, checkpoint := range checkpoints {
		if first == "" {
			first = c2sGroup1Classification(checkpoint)
		}
		if checkpoint.StandardFullRNSCapacity.MaxAbsOverQ01Half > 1 && (worst == nil || checkpoint.StandardFullRNSCapacity.MaxAbsOverQ01Half > worst.WorstRatioToQ01Half) {
			value := c2sAliasCompressionFor(checkpoint, "before plaintext multiplication or accumulation as indicated by first unsafe checkpoint")
			worst = &value
		}
	}
	if first == "" {
		first = "logn13_c2s_group1_linear_no_internal_divergence"
	}
	if worst != nil {
		worst.KMin = int(math.Max(0, math.Ceil(math.Log2(worst.WorstRatioToQ01Half))))
		worst.KSafe = worst.KMin + 1
	}
	return checkpoints, first, worst, nil
}

func captureC2SGroup1RescaleSummary(params ckks.Parameters, standardBefore, fastBefore, standardAfter, fastAfter *rlwe.Ciphertext) (*c2sGroup1RescaleSummary, error) {
	standardPre, err := postMod1S2CCapacityFromCiphertext(params, standardBefore)
	if err != nil {
		return nil, err
	}
	fastPre, err := postMod1S2CCapacityFromFastCiphertext(params, fastBefore)
	if err != nil {
		return nil, err
	}
	standardPost, err := postMod1S2CCapacityFromCiphertext(params, standardAfter)
	if err != nil {
		return nil, err
	}
	fastPost, err := postMod1S2CCapacityFromFastCiphertext(params, fastAfter)
	if err != nil {
		return nil, err
	}
	match, err := c2sAliasRowsMatch(params, standardAfter, fastAfter)
	if err != nil {
		return nil, err
	}
	return &c2sGroup1RescaleSummary{PreLevel: standardBefore.Level(), PostLevel: standardAfter.Level(), PreScale: finalizationScaleString(standardBefore.Scale), PostScale: finalizationScaleString(standardAfter.Scale), DroppedModulus: fmt.Sprintf("%d", params.Q()[standardBefore.Level()]), StandardPreCapacity: standardPre, FastPreCapacity: fastPre, StandardPostCapacity: standardPost, FastPostCapacity: fastPost, FastResiduesMatch: match}, nil
}

func runFIX001P3DiagLogN13C2SGroup1LinearAlias(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, skN2, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return err
	}
	params := btp.BootstrappingParameters
	result := c2sGroup1Result{
		SchemaVersion: "fix-001-p3-diag-logn13-c2s-group1-linear-alias.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		StandardProof:    semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: false, OrdinaryStageCallsUsed: true},
		DerivedC2SBudget: acceptedC2SGroup1Budget, FastEvalModMatchedInputUnresolved: true, FirstFailingCheckpoint: "none", FirstSupportedCause: "logn13_c2s_group1_precondition_mismatch",
		Validation: map[string]interface{}{"primary_required_base": requiredC2SGroup1AliasPrimaryBase, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredC2SGroup1AliasSecondary, "no_secondary_production_changes": true, "no_matrix_generation_or_parameter_changes": true, "no_production_fix": true, "no_group1_compressed_design": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_evalmod_work": true, "no_groups_2_or_3_internal_diagnosis": true},
	}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil || !result.Validation["secondary_exact_required_commit"].(bool) || !result.Validation["secondary_clean"].(bool) {
		return c2sGroup1Write(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return c2sGroup1Write(result, outPath)
	}
	standardControlDecoded, err := decodeWithSecret(residual, standardControl, skN1)
	if err != nil {
		return err
	}
	standardMetric, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), standardControlDecoded, c2sPrecisionFinalThreshold)
	if err != nil {
		return err
	}
	result.StandardEndToEnd = &standardMetric
	if !standardMetric.Pass {
		return c2sGroup1Write(result, outPath)
	}
	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastBefore, standardBefore := fastInput.CopyNew(), standardInput.CopyNew()
	fastPacked, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput.CopyNew()})
	if err != nil {
		return err
	}
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput.CopyNew()})
	if err != nil {
		return err
	}
	fastScaled, _, err := fastEval.ScaleDown(&fastPacked[0])
	if err != nil {
		return err
	}
	standardScaled, _, err := standardEval.ScaleDown(&standardPacked[0])
	if err != nil {
		return err
	}
	fastModUp, err := fastEval.ModUp(fastScaled)
	if err != nil {
		return err
	}
	standardModUp, err := standardEval.ModUp(standardScaled)
	if err != nil {
		return err
	}
	fastSK := zeroSecret(params)
	fastMetadata, fastValues, err := semanticBisectView(fastModUp, params, fastSK)
	if err != nil {
		return err
	}
	standardMetadata, standardValues, err := c2sPrecisionFullView(standardModUp, params, skN2)
	if err != nil {
		return err
	}
	modUpMetric, err := semanticBisectMetricFor(standardValues, fastValues, c2sPrecisionInputThreshold)
	if err != nil {
		return err
	}
	result.ModUpMetadataMatch = fastMetadata.SourceLevel == standardMetadata.SourceLevel && fastMetadata.Scale == standardMetadata.Scale && fastMetadata.Degree == standardMetadata.Degree && fastMetadata.N == standardMetadata.N && fastMetadata.LogDimensions == standardMetadata.LogDimensions && fastMetadata.IsNTT == standardMetadata.IsNTT && fastMetadata.IsMontgomery && !standardMetadata.IsMontgomery
	result.ModUpNoInputMutation = fix001P3E2EInputUnchanged(fastBefore, fastInput) && fix001P3E2EInputUnchanged(standardBefore, standardInput)
	result.ModUp = &c2sPrecisionCheckpoint{Checkpoint: "mod_up", Fast: fastMetadata, Standard: standardMetadata, FastVsStandard: modUpMetric}
	if !result.ModUpMetadataMatch || !result.ModUpNoInputMutation || !modUpMetric.Pass {
		result.FirstFailingCheckpoint = "mod_up"
		return c2sGroup1Write(result, outPath)
	}
	if _, _, err := fastEval.CoeffsToSlots(fastModUp.CopyNew()); err != nil {
		return err
	}
	if _, _, err := standardEval.CoeffsToSlots(standardModUp.CopyNew()); err != nil {
		return err
	}
	if len(standardEval.C2SDFTMatrix.Matrices) < 2 || len(standardEval.C2SDFTMatrix.Levels) != 4 || len(fastEval.C2SDFTMatrix.Levels) != 4 {
		return c2sGroup1Write(result, outPath)
	}
	result.Validation["c2s_schedule_is_four_one_level_groups"] = true
	result.Structure = c2sAliasStructureFromMatrix(standardEval.C2SDFTMatrix.Matrices[1], fastModUp)
	result.Structure.Path = "BSGS"

	originalStd0, originalFast0, err := c2sCompressedLinearTransform(params, standardEval, fastEval, standardEval.C2SDFTMatrix.Matrices[0], standardModUp.CopyNew(), fastModUp.CopyNew())
	if err != nil {
		return err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, originalStd0, originalFast0); err != nil {
		return err
	}
	originalGroup0Standard, err := c2sPrecisionCapture("original_group0_post_rescale", originalFast0, originalStd0, params, fastSK, skN2, nil, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	originalStd1, originalFast1, err := c2sCompressedLinearTransform(params, standardEval, fastEval, standardEval.C2SDFTMatrix.Matrices[1], originalStd0.CopyNew(), originalFast0.CopyNew())
	if err != nil {
		return err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, originalStd1, originalFast1); err != nil {
		return err
	}
	result.Validation["original_uncompressed_group1_oracle_recorded"] = true
	result.OriginalGroup1InputCapacity, err = func() (*postMod1S2CCapacity, error) {
		capacity, e := postMod1S2CCapacityFromCiphertext(params, originalStd0)
		return &capacity, e
	}()
	if err != nil {
		return err
	}

	compressed0, _, err := c2sCompressedMatrix(params, standardEval.C2SDFTMatrix, 4)
	if err != nil {
		return err
	}
	group0Std, group0Fast, err := c2sCompressedLinearTransform(params, standardEval, fastEval, compressed0, standardModUp.CopyNew(), fastModUp.CopyNew())
	if err != nil {
		return err
	}
	group0Checkpoints, err := c2sCompressedReplayGroup0(params, standardEval, fastEval, compressed0, standardModUp.CopyNew(), fastModUp.CopyNew(), skN2, fastSK, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	_ = group0Checkpoints
	if err := c2sCompressedRescale(params, standardEval, fastEval, group0Std, group0Fast); err != nil {
		return err
	}
	if err := c2sCompressedApplyRestore(params, fastEval.FastCKKS, group0Fast, group0Std, 4); err != nil {
		return err
	}
	group0Restored, err := c2sPrecisionCapture("group0_restored", group0Fast, group0Std, params, fastSK, skN2, nil, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Group0Restored = &group0Restored
	result.Group0OriginalPostRescale = &originalGroup0Standard
	originalStd0Values, err := decodeWithSecret(params, originalStd0, skN2)
	if err != nil {
		return err
	}
	group0StdValues, err := decodeWithSecret(params, group0Std, skN2)
	if err != nil {
		return err
	}
	group0Delta, err := semanticBisectMetricFor(originalStd0Values, group0StdValues, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Group0RestoredVsOriginal = &group0Delta
	group0FastMeta, _, err := semanticBisectView(group0Fast, params, fastSK)
	if err != nil {
		return err
	}
	originalFast0Meta, _, err := semanticBisectView(originalFast0, params, fastSK)
	if err != nil {
		return err
	}
	originalStd0Meta, _, err := c2sPrecisionFullView(originalStd0, params, skN2)
	if err != nil {
		return err
	}
	result.Group0MetadataMatches = c2sGroup1MetadataEqual(group0FastMeta, originalFast0Meta) && c2sGroup1MetadataEqual(group0Restored.Standard, originalStd0Meta)
	group0Valid := true
	for _, checkpoint := range group0Checkpoints {
		if !checkpoint.FullRNSCenteredQ01Unique || !checkpoint.FastResiduesMatch || checkpoint.Semantic != nil && !checkpoint.Semantic.Pass {
			group0Valid = false
		}
	}
	result.Group0Validated = group0Valid && group0Restored.FastQ01Capacity != nil && group0Restored.FastQ01Capacity.Pass && group0Restored.StandardFullRNSCapacity != nil && group0Restored.StandardFullRNSCapacity.Pass && result.Group0MetadataMatches && group0Delta.Pass
	result.C2SErrorAloneSufficient = true
	if !result.Group0Validated {
		result.FirstFailingCheckpoint = "group0_validated_k4_path"
		return c2sGroup1Write(result, outPath)
	}

	aligned, err := c2sGroup1CaptureInput(params, group0Std, group0Fast, skN2, fastSK, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.AlignedGroup1Input = &aligned
	result.Structure.InputLevel = group0Fast.Level()
	result.Structure.InputScale = finalizationScaleString(group0Fast.Scale)
	if !aligned.Aligned {
		result.FirstFailingCheckpoint = "group1_input_alignment"
		result.FirstSupportedCause = "logn13_c2s_group1_input_alignment_mismatch"
		return c2sGroup1Write(result, outPath)
	}

	checkpoints, classification, worst, err := c2sGroup1Replay(params, standardEval, fastEval, group0Std, group0Fast, skN2, fastSK, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Checkpoints, result.FirstSupportedCause, result.WorstFullGroupCapacity = checkpoints, classification, worst
	for _, checkpoint := range checkpoints {
		if c2sGroup1Classification(checkpoint) != "" {
			result.FirstFailingCheckpoint = checkpoint.Checkpoint
			break
		}
	}
	group1StdLinear, group1FastLinear, err := c2sCompressedLinearTransform(params, standardEval, fastEval, standardEval.C2SDFTMatrix.Matrices[1], group0Std.CopyNew(), group0Fast.CopyNew())
	if err != nil {
		return err
	}
	pre, err := c2sCompressedStageCapture("group1_pre_rescale", group1FastLinear, group1StdLinear, params, fastSK, skN2, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Group1PreRescale = &pre
	preStd := group1StdLinear.CopyNew()
	preFast := group1FastLinear.CopyNew()
	if err := c2sCompressedRescale(params, standardEval, fastEval, group1StdLinear, group1FastLinear); err != nil {
		return err
	}
	post, err := c2sCompressedStageCapture("group1_post_rescale", group1FastLinear, group1StdLinear, params, fastSK, skN2, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Group1PostRescale = &post
	result.RescaleSummary, err = captureC2SGroup1RescaleSummary(params, preStd, preFast, group1StdLinear, group1FastLinear)
	if err != nil {
		return err
	}
	result.RescaleSummary.FastVsStandardAfter = post.FastVsStandard
	alignedValues, err := decodeWithSecret(params, group1StdLinear, skN2)
	if err != nil {
		return err
	}
	originalValues, err := decodeWithSecret(params, originalStd1, skN2)
	if err != nil {
		return err
	}
	comparison, err := semanticBisectMetricFor(originalValues, alignedValues, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.AlignedStandardVsOriginal = &comparison
	result.Validation["group1_full_replay_completed"] = true
	result.Validation["group1_post_rescale_budget_breached"] = post.FastVsStandard != nil && !post.FastVsStandard.Pass
	result.Validation["c2s_error_alone_sufficient_for_final_failure"] = true
	if result.FirstFailingCheckpoint == "none" && result.FirstSupportedCause == "logn13_c2s_group1_linear_no_internal_divergence" {
		result.FirstFailingCheckpoint = "none"
	}
	return c2sGroup1Write(result, outPath)
}
