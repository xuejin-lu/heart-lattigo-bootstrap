package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	ckkslintrans "github.com/tuneinsight/lattigo/v6/circuits/ckks/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredC2SGroup1DesignPrimaryBase = "5c1c0d9a5e0a599370cc152faab56d91649518f8"
	requiredC2SGroup1DesignSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
)

type c2sGroup1DesignCandidateSummary struct {
	K                              int                      `json:"k"`
	OriginalMatrixScale            string                   `json:"original_matrix_scale"`
	OriginalMatrixScaleLog2        float64                  `json:"original_matrix_scale_log2"`
	CompressedMatrixScale          string                   `json:"compressed_matrix_scale"`
	CompressedMatrixScaleLog2      float64                  `json:"compressed_matrix_scale_log2"`
	MathematicalDiagonalsUnchanged bool                     `json:"mathematical_diagonal_values_unchanged"`
	SameStructure                  bool                     `json:"same_bsgs_structure"`
	SameDiagonalIndices            bool                     `json:"same_diagonal_indices"`
	SameN1                         bool                     `json:"same_n1"`
	SameLevelQ                     bool                     `json:"same_level_q"`
	SameLevelP                     bool                     `json:"same_level_p"`
	SameLogDimensions              bool                     `json:"same_log_dimensions"`
	CheckpointCount                int                      `json:"checkpoint_count"`
	FirstUnsafeCheckpoint          string                   `json:"first_unsafe_checkpoint"`
	FirstResidueMismatch           string                   `json:"first_residue_mismatch_checkpoint"`
	WorstCapacity                  *c2sAliasCompression     `json:"worst_full_group_capacity,omitempty"`
	CompleteGroupCapacityPass      bool                     `json:"complete_group_capacity_pass"`
	CompleteGroupResiduesPass      bool                     `json:"complete_group_residue_match_pass"`
	PreRescale                     *c2sCompressedStage      `json:"pre_rescale,omitempty"`
	PostRescale                    *c2sCompressedStage      `json:"post_rescale,omitempty"`
	Restore                        *c2sCompressedRestore    `json:"restore,omitempty"`
	Comparison                     *c2sCompressedComparison `json:"restored_comparison,omitempty"`
	RestoredSemanticPass           bool                     `json:"restored_semantic_pass"`
	ValidCompleteGroup             bool                     `json:"valid_complete_group"`
}

type c2sGroup1DesignDownstream struct {
	ValidatedKSafe                   int                   `json:"validated_k_safe"`
	CorrectedGroup1PostRescale       *semanticBisectMetric `json:"corrected_group1_post_rescale,omitempty"`
	FinalRealDelta                   *semanticBisectMetric `json:"final_c2s_real_delta,omitempty"`
	FinalImagDelta                   *semanticBisectMetric `json:"final_c2s_imag_delta,omitempty"`
	FirstLaterBudgetBreach           string                `json:"first_later_group_boundary_over_budget"`
	Groups0And1FixSatisfiesC2SBudget bool                  `json:"groups0_1_fix_satisfies_c2s_budget"`
	Disposition                      string                `json:"disposition"`
}

type c2sGroup1DesignResult struct {
	SchemaVersion                     string                            `json:"schema_version"`
	Timestamp                         time.Time                         `json:"timestamp"`
	Primary                           RepositoryMetadata                `json:"primary_repository"`
	Lattigo                           RepositoryMetadata                `json:"lattigo_repository"`
	Environment                       EnvironmentMetadata               `json:"environment"`
	Config                            BootstrapConfig                   `json:"config"`
	Parameters                        ExperimentParameters              `json:"effective_parameters"`
	Workload                          CorrectnessWorkload               `json:"workload"`
	StandardProof                     semanticBisectStandardProof       `json:"standard_control_proof"`
	StandardEndToEnd                  *semanticBisectMetric             `json:"standard_end_to_end_semantic,omitempty"`
	ModUp                             *c2sPrecisionCheckpoint           `json:"modup_precondition,omitempty"`
	ModUpMetadataMatch                bool                              `json:"modup_metadata_match"`
	ModUpNoInputMutation              bool                              `json:"modup_no_input_mutation"`
	DerivedC2SBudget                  float64                           `json:"derived_c2s_budget"`
	C2SErrorAloneSufficient           bool                              `json:"c2s_error_alone_sufficient_for_final_failure"`
	FastEvalModMatchedInputUnresolved bool                              `json:"fast_evalmod_matched_input_divergence_remains_unresolved"`
	Group0K4Validated                 bool                              `json:"accepted_group0_k4_validated"`
	Group0RestoredVsOriginal          *semanticBisectMetric             `json:"group0_restored_vs_original_standard,omitempty"`
	Group1AlignedInput                *c2sGroup1InputSummary            `json:"aligned_group1_input,omitempty"`
	Group1OriginalStructure           c2sAliasStructure                 `json:"group1_transform_structure"`
	Group1StructureMatchesPredecessor bool                              `json:"group1_structure_matches_predecessor"`
	OriginalGroup1InputCapacity       *postMod1S2CCapacity              `json:"original_group1_input_capacity,omitempty"`
	Candidates                        []c2sGroup1DesignCandidateSummary `json:"candidates"`
	KMinFullGroup                     int                               `json:"k_min_full_group"`
	KSafeFullGroup                    int                               `json:"k_safe_full_group"`
	Downstream                        *c2sGroup1DesignDownstream        `json:"downstream_probe,omitempty"`
	FirstSupportedCause               string                            `json:"first_supported_cause"`
	Validation                        map[string]interface{}            `json:"validation"`
}

func c2sGroup1DesignWrite(result c2sGroup1DesignResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func c2sGroup1CompressedMatrix(params ckks.Parameters, original dft.Matrix, matrixIndex, k int) (ckkslintrans.LinearTransformation, bool, error) {
	if matrixIndex < 0 || matrixIndex >= len(original.Matrices) {
		return ckkslintrans.LinearTransformation{}, false, fmt.Errorf("group-1 matrix index %d is unavailable", matrixIndex)
	}
	pVec := original.MatrixLiteral.GenMatrices(params.LogN(), params.EncodingPrecision())
	if matrixIndex >= len(pVec) {
		return ckkslintrans.LinearTransformation{}, false, fmt.Errorf("mathematical matrix index %d is unavailable", matrixIndex)
	}
	originalMatrix := original.Matrices[matrixIndex]
	scale := originalMatrix.Scale.Div(rlwe.NewScale(float64(uint64(1) << k)))
	ltParams := ckkslintrans.Parameters{DiagonalsIndexList: pVec[matrixIndex].DiagonalsIndexList(), LevelQ: originalMatrix.LevelQ, LevelP: originalMatrix.LevelP, Scale: scale, LogDimensions: originalMatrix.LogDimensions, LogBabyStepGiantStepRatio: originalMatrix.LogBabyStepGiantStepRatio}
	compressed := ckkslintrans.NewTransformation(params, ltParams)
	if err := ckkslintrans.Encode(ckks.NewEncoder(params), pVec[matrixIndex], compressed); err != nil {
		return ckkslintrans.LinearTransformation{}, false, fmt.Errorf("encode group-1 compressed matrix k=%d: %w", k, err)
	}
	return compressed, true, nil
}

func c2sGroup1ExpectedStructure(inputLevel int, inputScale string) c2sAliasStructure {
	return c2sAliasStructure{
		N1: 256, Path: "BSGS", DiagonalCount: 15,
		SortedDiagonalIndices: []int{0, 64, 128, 192, 256, 320, 384, 448, 3648, 3712, 3776, 3840, 3904, 3968, 4032},
		GiantGroups:           []c2sAliasGiantGroup{{GiantRotation: 0, BabyRotations: []int{0, 64, 128, 192}}, {GiantRotation: 256, BabyRotations: []int{0, 64, 128, 192}}, {GiantRotation: 3584, BabyRotations: []int{64, 128, 192}}, {GiantRotation: 3840, BabyRotations: []int{0, 64, 128, 192}}},
		LevelQ:                16, LevelP: 4, LogDimensions: "rows=0,cols=12", InputLevel: inputLevel, InputScale: inputScale,
	}
}

func c2sGroup1DesignCandidate(candidate c2sCompressedCandidate) c2sGroup1DesignCandidateSummary {
	summary := c2sGroup1DesignCandidateSummary{K: candidate.K, OriginalMatrixScale: candidate.OriginalMatrixScale, OriginalMatrixScaleLog2: candidate.OriginalMatrixScaleLog2, CompressedMatrixScale: candidate.CompressedMatrixScale, CompressedMatrixScaleLog2: candidate.CompressedMatrixScaleLog2, MathematicalDiagonalsUnchanged: candidate.MathematicalDiagonalsUnchanged, SameStructure: candidate.SameStructure, SameDiagonalIndices: candidate.SameDiagonalIndices, SameN1: candidate.SameN1, SameLevelQ: candidate.SameLevelQ, SameLevelP: candidate.SameLevelP, SameLogDimensions: candidate.SameLogDimensions, CheckpointCount: len(candidate.Checkpoints), CompleteGroupCapacityPass: candidate.CompleteGroupCapacityPass, CompleteGroupResiduesPass: candidate.CompleteGroupResiduesPass, PreRescale: candidate.PreRescale, PostRescale: candidate.PostRescale, Restore: candidate.Restore, Comparison: candidate.Comparison, RestoredSemanticPass: candidate.RestoredSemanticPass, ValidCompleteGroup: candidate.ValidCompleteGroup}
	if summary.PreRescale != nil {
		value := *summary.PreRescale
		value.Checkpoint = fmt.Sprintf("group1_k%d_pre_rescale", candidate.K)
		summary.PreRescale = &value
	}
	if summary.PostRescale != nil {
		value := *summary.PostRescale
		value.Checkpoint = fmt.Sprintf("group1_k%d_post_rescale", candidate.K)
		summary.PostRescale = &value
	}
	for _, checkpoint := range candidate.Checkpoints {
		if summary.FirstUnsafeCheckpoint == "" && !checkpoint.FullRNSCenteredQ01Unique {
			summary.FirstUnsafeCheckpoint = checkpoint.Checkpoint
		}
		if summary.FirstResidueMismatch == "" && !checkpoint.FastResiduesMatch {
			summary.FirstResidueMismatch = checkpoint.Checkpoint
		}
		if summary.WorstCapacity == nil || checkpoint.StandardFullRNSCapacity.MaxAbsOverQ01Half > summary.WorstCapacity.WorstRatioToQ01Half {
			value := c2sAliasCompressionFor(checkpoint, "before plaintext multiplication or accumulation as indicated by first unsafe checkpoint")
			if checkpoint.StandardFullRNSCapacity.MaxAbsOverQ01Half < 1 {
				value.KMin = 0
				value.KSafe = 1
			}
			summary.WorstCapacity = &value
		}
	}
	return summary
}

func c2sGroup1DesignDownstreamProbe(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, originalMatrix dft.Matrix, group0Standard, group0Fast *rlwe.Ciphertext, group1Matrix ckkslintrans.LinearTransformation, originalGroup1Standard *rlwe.Ciphertext, fastSK, standardSK *rlwe.SecretKey, originalStandardReal, originalStandardImag *rlwe.Ciphertext, budget float64, k int) (c2sGroup1DesignDownstream, error) {
	standardCurrent, fastCurrent, err := c2sCompressedLinearTransform(params, standardEval, fastEval, group1Matrix, group0Standard.CopyNew(), group0Fast.CopyNew())
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, standardCurrent, fastCurrent); err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	if err := c2sCompressedApplyRestore(params, fastEval.FastCKKS, fastCurrent, standardCurrent, k); err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	correctedValues, err := decodeWithSecret(params, standardCurrent, standardSK)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	originalValues, err := decodeWithSecret(params, originalGroup1Standard, standardSK)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	correctedGroup1, err := semanticBisectMetricFor(originalValues, correctedValues, budget)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	firstBreach := "none"
	for matrixIndex := 2; matrixIndex < len(originalMatrix.Matrices); matrixIndex++ {
		standardCurrent, fastCurrent, err = c2sCompressedLinearTransform(params, standardEval, fastEval, originalMatrix.Matrices[matrixIndex], standardCurrent, fastCurrent)
		if err != nil {
			return c2sGroup1DesignDownstream{}, err
		}
		if err := c2sCompressedRescale(params, standardEval, fastEval, standardCurrent, fastCurrent); err != nil {
			return c2sGroup1DesignDownstream{}, err
		}
		stage, err := c2sCompressedStageCapture(fmt.Sprintf("group%d_post_rescale", matrixIndex), fastCurrent, standardCurrent, params, fastSK, standardSK, budget)
		if err != nil {
			return c2sGroup1DesignDownstream{}, err
		}
		if firstBreach == "none" && stage.FastVsStandard != nil && !stage.FastVsStandard.Pass {
			firstBreach = stage.Checkpoint
		}
	}
	remainingLiteral := originalMatrix.MatrixLiteral
	remainingLiteral.LevelQ = fastCurrent.Level()
	remainingLiteral.Levels = append([]int(nil), originalMatrix.Levels[2:]...)
	remaining := dft.Matrix{MatrixLiteral: remainingLiteral, Matrices: append([]ckkslintrans.LinearTransformation(nil), originalMatrix.Matrices[2:]...)}
	fastReal, fastImag, err := fastEval.DFTEvaluator.CoeffsToSlotsNew(fastCurrent, remaining)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	_, fastRealValues, err := semanticBisectView(fastReal, params, fastSK)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	standardRealValues, err := decodeWithSecret(params, originalStandardReal, standardSK)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	finalReal, err := semanticBisectMetricFor(standardRealValues, fastRealValues, budget)
	if err != nil {
		return c2sGroup1DesignDownstream{}, err
	}
	var finalImag *semanticBisectMetric
	if fastImag != nil && originalStandardImag != nil {
		_, fastImagValues, err := semanticBisectView(fastImag, params, fastSK)
		if err != nil {
			return c2sGroup1DesignDownstream{}, err
		}
		standardImagValues, err := decodeWithSecret(params, originalStandardImag, standardSK)
		if err != nil {
			return c2sGroup1DesignDownstream{}, err
		}
		metric, err := semanticBisectMetricFor(standardImagValues, fastImagValues, budget)
		if err != nil {
			return c2sGroup1DesignDownstream{}, err
		}
		finalImag = &metric
	}
	satisfies := correctedGroup1.Pass && finalReal.Pass && (finalImag == nil || finalImag.Pass) && firstBreach == "none"
	disposition := "groups0_1_fix_valid_but_final_real_imag_split_breach"
	if satisfies {
		disposition = "groups0_1_fix_satisfies_c2s_budget"
	} else if firstBreach != "none" {
		disposition = fmt.Sprintf("groups0_1_fix_valid_but_next_budget_breach_%s", firstBreach[:len(firstBreach)-len("_post_rescale")])
	}
	return c2sGroup1DesignDownstream{ValidatedKSafe: k, CorrectedGroup1PostRescale: &correctedGroup1, FinalRealDelta: &finalReal, FinalImagDelta: finalImag, FirstLaterBudgetBreach: firstBreach, Groups0And1FixSatisfiesC2SBudget: satisfies, Disposition: disposition}, nil
}

func runFIX001P3DesignLogN13C2SGroup1CompressedLinear(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := c2sGroup1DesignResult{SchemaVersion: "fix-001-p3-design-logn13-c2s-group1-compressed-linear.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, StandardProof: semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: false, OrdinaryStageCallsUsed: true}, DerivedC2SBudget: acceptedC2SGroup1Budget, C2SErrorAloneSufficient: true, FastEvalModMatchedInputUnresolved: true, KMinFullGroup: -1, KSafeFullGroup: -1, FirstSupportedCause: "logn13_c2s_group1_compressed_linear_design_not_validated", Validation: map[string]interface{}{"primary_required_base": requiredC2SGroup1DesignPrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredC2SGroup1DesignSecondary, "no_secondary_production_changes": true, "no_matrix_generation_or_parameter_changes": true, "no_fast_production_changes": true, "no_production_fix": true, "no_internal_group2_or_3_diagnosis": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "evalmod_untouched": true}}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil || !secondaryClean || secondaryCommit != requiredC2SGroup1DesignSecondary {
		return c2sGroup1DesignWrite(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return c2sGroup1DesignWrite(result, outPath)
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
		result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
		return c2sGroup1DesignWrite(result, outPath)
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
		result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
		return c2sGroup1DesignWrite(result, outPath)
	}
	_, _, err = fastEval.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	if len(standardEval.C2SDFTMatrix.Matrices) != 4 || len(standardEval.C2SDFTMatrix.Levels) != 4 || len(fastEval.C2SDFTMatrix.Matrices) != 4 || len(fastEval.C2SDFTMatrix.Levels) != 4 {
		result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
		return c2sGroup1DesignWrite(result, outPath)
	}
	for _, level := range standardEval.C2SDFTMatrix.Levels {
		if level != 1 {
			result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
			return c2sGroup1DesignWrite(result, outPath)
		}
	}
	result.Validation["c2s_schedule_is_four_one_level_groups"] = true
	originalMatrix := standardEval.C2SDFTMatrix
	result.Group1OriginalStructure = c2sAliasStructureFromMatrix(originalMatrix.Matrices[1], fastModUp)
	result.Group1OriginalStructure.InputLevel = fastModUp.Level()
	result.Group1OriginalStructure.InputScale = finalizationScaleString(fastModUp.Scale)
	expected := c2sGroup1ExpectedStructure(fastModUp.Level(), finalizationScaleString(fastModUp.Scale))
	result.Group1StructureMatchesPredecessor = c2sCompressedStructureEqual(result.Group1OriginalStructure, expected)
	if !result.Group1StructureMatchesPredecessor {
		result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
		return c2sGroup1DesignWrite(result, outPath)
	}

	originalStd0, originalFast0, err := c2sCompressedLinearTransform(params, standardEval, fastEval, originalMatrix.Matrices[0], standardModUp.CopyNew(), fastModUp.CopyNew())
	if err != nil {
		return err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, originalStd0, originalFast0); err != nil {
		return err
	}
	originalStd1, originalFast1, err := c2sCompressedLinearTransform(params, standardEval, fastEval, originalMatrix.Matrices[1], originalStd0.CopyNew(), originalFast0.CopyNew())
	if err != nil {
		return err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, originalStd1, originalFast1); err != nil {
		return err
	}
	originalStd1Values, err := decodeWithSecret(params, originalStd1, skN2)
	if err != nil {
		return err
	}
	originalGroup1InputCapacity, err := postMod1S2CCapacityFromCiphertext(params, originalStd0)
	if err != nil {
		return err
	}
	result.OriginalGroup1InputCapacity = &originalGroup1InputCapacity
	compressed0, _, err := c2sCompressedMatrix(params, originalMatrix, 4)
	if err != nil {
		return err
	}
	group0Candidate, err := c2sCompressedCandidateRun(params, standardEval, fastEval, compressed0, originalMatrix.Matrices[0], standardModUp.CopyNew(), fastModUp.CopyNew(), skN2, fastSK, originalStd0, originalFast0, func() []complex128 { values, _ := decodeWithSecret(params, originalStd0, skN2); return values }(), acceptedC2SGroup1Budget, 4)
	if err != nil {
		return err
	}
	result.Group0K4Validated = group0Candidate.ValidCompleteGroup
	if group0Candidate.Comparison != nil {
		result.Group0RestoredVsOriginal = group0Candidate.Comparison.CompressedStandardRestoredVsOriginal
	}
	if !result.Group0K4Validated || result.Group0RestoredVsOriginal == nil || !result.Group0RestoredVsOriginal.Pass {
		result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
		return c2sGroup1DesignWrite(result, outPath)
	}
	group0Std, group0Fast, err := c2sCompressedLinearTransform(params, standardEval, fastEval, compressed0, standardModUp.CopyNew(), fastModUp.CopyNew())
	if err != nil {
		return err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, group0Std, group0Fast); err != nil {
		return err
	}
	if err := c2sCompressedApplyRestore(params, fastEval.FastCKKS, group0Fast, group0Std, 4); err != nil {
		return err
	}
	alignedInput, err := c2sGroup1CaptureInput(params, group0Std, group0Fast, skN2, fastSK, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Group1AlignedInput = &alignedInput
	result.Group1OriginalStructure.InputLevel = group0Fast.Level()
	result.Group1OriginalStructure.InputScale = finalizationScaleString(group0Fast.Scale)
	if !alignedInput.Aligned {
		result.FirstSupportedCause = "logn13_c2s_group1_compression_baseline_mismatch"
		return c2sGroup1DesignWrite(result, outPath)
	}

	for _, k := range []int{1, 2, 3} {
		compressedMatrix, sameMath, err := c2sGroup1CompressedMatrix(params, originalMatrix, 1, k)
		if err != nil {
			return err
		}
		candidate, err := c2sCompressedCandidateRun(params, standardEval, fastEval, compressedMatrix, originalMatrix.Matrices[1], group0Std.CopyNew(), group0Fast.CopyNew(), skN2, fastSK, originalStd1, originalFast1, originalStd1Values, acceptedC2SGroup1Budget, k)
		if err != nil {
			return err
		}
		candidate.MathematicalDiagonalsUnchanged = sameMath
		result.Candidates = append(result.Candidates, c2sGroup1DesignCandidate(candidate))
	}
	for _, candidate := range result.Candidates {
		if candidate.ValidCompleteGroup && result.KMinFullGroup < 0 {
			result.KMinFullGroup = candidate.K
		}
	}
	if result.KMinFullGroup >= 0 {
		for _, candidate := range result.Candidates {
			if candidate.K == result.KMinFullGroup+1 && candidate.ValidCompleteGroup {
				result.KSafeFullGroup = candidate.K
			}
		}
	}
	if result.KMinFullGroup >= 0 && result.KSafeFullGroup >= 0 {
		compressedSafe, _, err := c2sGroup1CompressedMatrix(params, originalMatrix, 1, result.KSafeFullGroup)
		if err != nil {
			return err
		}
		downstream, err := c2sGroup1DesignDownstreamProbe(params, standardEval, fastEval, originalMatrix, group0Std, group0Fast, compressedSafe, originalStd1, fastSK, skN2, standardReal, standardImag, acceptedC2SGroup1Budget, result.KSafeFullGroup)
		if err != nil {
			return err
		}
		result.Downstream = &downstream
		result.FirstSupportedCause = "logn13_c2s_group1_compressed_linear_design_validated"
	}
	return c2sGroup1DesignWrite(result, outPath)
}
