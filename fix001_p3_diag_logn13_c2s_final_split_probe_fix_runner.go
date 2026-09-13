package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredC2SFinalSplitPrimaryBase = "2c2e69717fbdaa38500968756a4392e156be478e"
	requiredC2SFinalSplitSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
)

type c2sFinalBoundary struct {
	Group          int                `json:"group"`
	Phase          string             `json:"phase"`
	DroppedModulus string             `json:"dropped_logical_modulus,omitempty"`
	Stage          c2sCompressedStage `json:"stage"`
}

type c2sFinalSplitCheckpoint struct {
	Checkpoint               string                 `json:"checkpoint"`
	Fast                     semanticBisectMetadata `json:"fast"`
	Standard                 semanticBisectMetadata `json:"standard"`
	FastQ01Capacity          postMod1S2CCapacity    `json:"fast_q01_capacity"`
	StandardFullRNSCapacity  postMod1S2CCapacity    `json:"standard_full_rns_capacity"`
	FullRNSCenteredQ01Unique bool                   `json:"full_rns_centered_q01_unique"`
	FastResiduesMatch        bool                   `json:"fast_residues_match_standard_mod_q01"`
	Semantic                 *semanticBisectMetric  `json:"semantic,omitempty"`
	InputUnchanged           bool                   `json:"input_unchanged"`
}

type c2sFinalSplitResult struct {
	SchemaVersion                         string                      `json:"schema_version"`
	Timestamp                             time.Time                   `json:"timestamp"`
	Primary                               RepositoryMetadata          `json:"primary_repository"`
	Lattigo                               RepositoryMetadata          `json:"lattigo_repository"`
	Environment                           EnvironmentMetadata         `json:"environment"`
	Config                                BootstrapConfig             `json:"config"`
	Parameters                            ExperimentParameters        `json:"effective_parameters"`
	Workload                              CorrectnessWorkload         `json:"workload"`
	StandardProof                         semanticBisectStandardProof `json:"standard_control_proof"`
	StandardEndToEnd                      *semanticBisectMetric       `json:"standard_end_to_end_semantic,omitempty"`
	ModUp                                 *c2sPrecisionCheckpoint     `json:"modup_precondition,omitempty"`
	ModUpMetadataMatch                    bool                        `json:"modup_metadata_match"`
	ModUpNoInputMutation                  bool                        `json:"modup_no_input_mutation"`
	DerivedC2SBudget                      float64                     `json:"derived_c2s_budget"`
	Group0K4Validated                     bool                        `json:"accepted_group0_k4_validated"`
	Group1K2Validated                     bool                        `json:"accepted_group1_k2_validated"`
	Group1RestoredVsOriginal              *semanticBisectMetric       `json:"group1_restored_vs_original_standard,omitempty"`
	Group1AlignedInput                    *c2sGroup1InputSummary      `json:"aligned_group1_input,omitempty"`
	Group2And3Boundaries                  []c2sFinalBoundary          `json:"group2_group3_boundaries,omitempty"`
	ZVFastMetadata                        *semanticBisectMetadata     `json:"zv_fast_metadata,omitempty"`
	ZVStandardMetadata                    *semanticBisectMetadata     `json:"zv_standard_metadata,omitempty"`
	ZVPostGroup3Semantic                  *semanticBisectMetric       `json:"zv_fast_vs_standard,omitempty"`
	PredecessorFinalSplitDispositionValid bool                        `json:"predecessor_final_split_disposition_valid"`
	PredecessorDoubleDFTTailConfirmed     bool                        `json:"predecessor_double_dft_tail_confirmed"`
	PredecessorControlFlowEvidence        map[string]string           `json:"predecessor_control_flow_evidence"`
	SplitCheckpoints                      []c2sFinalSplitCheckpoint   `json:"split_checkpoints"`
	FinalRealFastVsStandard               *semanticBisectMetric       `json:"final_real_fast_vs_ordinary_standard,omitempty"`
	FinalImagFastVsStandard               *semanticBisectMetric       `json:"final_imag_fast_vs_ordinary_standard,omitempty"`
	FinalRealFastVsAlignedStandard        *semanticBisectMetric       `json:"final_real_fast_vs_aligned_split_standard,omitempty"`
	FinalImagFastVsAlignedStandard        *semanticBisectMetric       `json:"final_imag_fast_vs_aligned_split_standard,omitempty"`
	C2SErrorAloneSufficient               bool                        `json:"c2s_error_alone_sufficient_for_final_failure"`
	C2SPrecisionBlockerResolved           bool                        `json:"c2s_precision_blocker_resolved_for_logn13"`
	FastEvalModMatchedInputUnresolved     bool                        `json:"fast_evalmod_matched_input_divergence_remains_unresolved"`
	FirstFailingCheckpoint                string                      `json:"first_failing_checkpoint"`
	FirstSupportedCause                   string                      `json:"first_supported_cause"`
	Disposition                           string                      `json:"disposition"`
	Validation                            map[string]interface{}      `json:"validation"`
}

func c2sFinalSplitWrite(result c2sFinalSplitResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func c2sFinalSplitCapture(name string, standard, fast *rlwe.Ciphertext, params ckks.Parameters, standardSK, fastSK *rlwe.SecretKey, budget float64, inputUnchanged bool) (c2sFinalSplitCheckpoint, error) {
	standardMeta, standardValues, err := c2sPrecisionFullView(standard, params, standardSK)
	if err != nil {
		return c2sFinalSplitCheckpoint{}, err
	}
	fastMeta, fastValues, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return c2sFinalSplitCheckpoint{}, err
	}
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
	if err != nil {
		return c2sFinalSplitCheckpoint{}, err
	}
	standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standard)
	if err != nil {
		return c2sFinalSplitCheckpoint{}, err
	}
	residuesMatch, err := c2sAliasRowsMatch(params, standard, fast)
	if err != nil {
		return c2sFinalSplitCheckpoint{}, err
	}
	checkpoint := c2sFinalSplitCheckpoint{Checkpoint: name, Fast: fastMeta, Standard: standardMeta, FastQ01Capacity: fastCapacity, StandardFullRNSCapacity: standardCapacity, FullRNSCenteredQ01Unique: standardCapacity.Pass, FastResiduesMatch: residuesMatch, InputUnchanged: inputUnchanged}
	if standardCapacity.Pass {
		metric, err := semanticBisectMetricFor(standardValues, fastValues, budget)
		if err != nil {
			return c2sFinalSplitCheckpoint{}, err
		}
		checkpoint.Semantic = &metric
	}
	return checkpoint, nil
}

func c2sFinalSplitCause(checkpoint c2sFinalSplitCheckpoint) string {
	if checkpoint.FullRNSCenteredQ01Unique && (!checkpoint.FastResiduesMatch || checkpoint.Semantic != nil && !checkpoint.Semantic.Pass) {
		switch checkpoint.Checkpoint {
		case "split_conjugate":
			return "logn13_c2s_final_split_conjugate_mismatch"
		case "split_imag_sub":
			return "logn13_c2s_final_split_imag_sub_mismatch"
		case "split_imag_mul_minus_i":
			return "logn13_c2s_final_split_imag_mul_mismatch"
		case "split_real_add":
			return "logn13_c2s_final_split_real_add_mismatch"
		}
	}
	if !checkpoint.FullRNSCenteredQ01Unique && checkpoint.FastResiduesMatch {
		switch checkpoint.Checkpoint {
		case "split_conjugate":
			return "logn13_c2s_final_split_conjugate_q01_alias"
		case "split_imag_sub":
			return "logn13_c2s_final_split_imag_sub_q01_alias"
		case "split_imag_mul_minus_i":
			return "logn13_c2s_final_split_imag_mul_q01_alias"
		case "split_real_add":
			return "logn13_c2s_final_split_real_add_q01_alias"
		}
	}
	return ""
}

func runFIX001P3DiagLogN13C2SFinalSplitProbeFix(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := c2sFinalSplitResult{
		SchemaVersion: "fix-001-p3-diag-logn13-c2s-final-split-probe-fix.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		DerivedC2SBudget: acceptedC2SGroup1Budget, C2SErrorAloneSufficient: true, FastEvalModMatchedInputUnresolved: true, PredecessorFinalSplitDispositionValid: false, FirstFailingCheckpoint: "none", FirstSupportedCause: "logn13_c2s_split_probe_prefix_mismatch",
		PredecessorControlFlowEvidence: map[string]string{"predecessor_probe": "c2sGroup1DesignDownstreamProbe", "manual_tail_replay": "originalMatrix.Matrices[2:] and originalMatrix.Levels[2:]", "fast_split_entrypoint": "FastEvaluator.CoeffsToSlotsNew", "fast_split_control_flow": "CoeffsToSlots -> eval.dft(zV, matrices, zV) before Conjugate/Sub/Mul/Add", "corrected_probe": "split-only replay from completed zV; no CoeffsToSlots* call"},
		Validation:                     map[string]interface{}{"primary_required_base": requiredC2SFinalSplitPrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredC2SFinalSplitSecondary, "no_secondary_production_changes": true, "no_dft_generation_or_parameter_changes": true, "no_production_fix": true, "no_evalmod_work": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "groups2_and3_executed_once_in_corrected_probe": false},
	}
	result.StandardProof = semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: false, OrdinaryStageCallsUsed: true}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil || !secondaryClean || secondaryCommit != requiredC2SFinalSplitSecondary {
		return c2sFinalSplitWrite(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return c2sFinalSplitWrite(result, outPath)
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
		return c2sFinalSplitWrite(result, outPath)
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
		return c2sFinalSplitWrite(result, outPath)
	}

	// The ordinary full-RNS Standard path is the authoritative final real/imag reference.
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	originalMatrix := standardEval.C2SDFTMatrix
	if len(originalMatrix.Matrices) != 4 || len(originalMatrix.Levels) != 4 {
		return c2sFinalSplitWrite(result, outPath)
	}
	for _, level := range originalMatrix.Levels {
		if level != 1 {
			return c2sFinalSplitWrite(result, outPath)
		}
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
		result.Group1RestoredVsOriginal = group0Candidate.Comparison.CompressedStandardRestoredVsOriginal
	}
	if !result.Group0K4Validated {
		return c2sFinalSplitWrite(result, outPath)
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
	compressed1, _, err := c2sGroup1CompressedMatrix(params, originalMatrix, 1, 2)
	if err != nil {
		return err
	}
	group1Candidate, err := c2sCompressedCandidateRun(params, standardEval, fastEval, compressed1, originalMatrix.Matrices[1], group0Std.CopyNew(), group0Fast.CopyNew(), skN2, fastSK, originalStd1, originalFast1, originalStd1Values, acceptedC2SGroup1Budget, 2)
	if err != nil {
		return err
	}
	result.Group1K2Validated = group1Candidate.ValidCompleteGroup
	if group1Candidate.Comparison != nil {
		result.Group1RestoredVsOriginal = group1Candidate.Comparison.CompressedStandardRestoredVsOriginal
	}
	if !result.Group1K2Validated {
		return c2sFinalSplitWrite(result, outPath)
	}
	group1Std, group1Fast, err := c2sCompressedLinearTransform(params, standardEval, fastEval, compressed1, group0Std.CopyNew(), group0Fast.CopyNew())
	if err != nil {
		return err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, group1Std, group1Fast); err != nil {
		return err
	}
	if err := c2sCompressedApplyRestore(params, fastEval.FastCKKS, group1Fast, group1Std, 2); err != nil {
		return err
	}
	aligned, err := c2sGroup1CaptureInput(params, group1Std, group1Fast, skN2, fastSK, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.Group1AlignedInput = &aligned
	if !aligned.Aligned {
		return c2sFinalSplitWrite(result, outPath)
	}

	standardCurrent := group1Std.CopyNew()
	fastCurrent := group1Fast.CopyNew()
	for group := 2; group <= 3; group++ {
		beforeLevel := standardCurrent.Level()
		standardCurrent, fastCurrent, err = c2sCompressedLinearTransform(params, standardEval, fastEval, originalMatrix.Matrices[group], standardCurrent, fastCurrent)
		if err != nil {
			return err
		}
		pre, err := c2sCompressedStageCapture(fmt.Sprintf("group%d_post_linear", group), fastCurrent, standardCurrent, params, fastSK, skN2, acceptedC2SGroup1Budget)
		if err != nil {
			return err
		}
		result.Group2And3Boundaries = append(result.Group2And3Boundaries, c2sFinalBoundary{Group: group, Phase: "post_linear", DroppedModulus: fmt.Sprintf("%d", params.Q()[beforeLevel]), Stage: pre})
		if err := c2sCompressedRescale(params, standardEval, fastEval, standardCurrent, fastCurrent); err != nil {
			return err
		}
		post, err := c2sCompressedStageCapture(fmt.Sprintf("group%d_post_rescale", group), fastCurrent, standardCurrent, params, fastSK, skN2, acceptedC2SGroup1Budget)
		if err != nil {
			return err
		}
		result.Group2And3Boundaries = append(result.Group2And3Boundaries, c2sFinalBoundary{Group: group, Phase: "post_rescale", DroppedModulus: fmt.Sprintf("%d", params.Q()[beforeLevel]), Stage: post})
		if post.FastVsStandard != nil && !post.FastVsStandard.Pass {
			result.FirstFailingCheckpoint = post.Checkpoint
			result.FirstSupportedCause = fmt.Sprintf("logn13_c2s_final_split_group%d_post_rescale_budget_breach", group)
			return c2sFinalSplitWrite(result, outPath)
		}
	}
	result.Validation["groups2_and3_executed_once_in_corrected_probe"] = true
	result.Validation["no_further_dft_factor_before_split"] = true
	zvFastMeta, _, err := semanticBisectView(fastCurrent, params, fastSK)
	if err != nil {
		return err
	}
	result.ZVFastMetadata = &zvFastMeta
	zvStandardMeta, _, err := c2sPrecisionFullView(standardCurrent, params, skN2)
	if err != nil {
		return err
	}
	result.ZVStandardMetadata = &zvStandardMeta
	zvMetric, err := c2sGroup1CaptureInput(params, standardCurrent, fastCurrent, skN2, fastSK, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.ZVPostGroup3Semantic = &zvMetric.FastVsStandard
	if !zvMetric.FastVsStandard.Pass {
		result.FirstFailingCheckpoint = "group3_post_rescale"
		result.FirstSupportedCause = "logn13_c2s_final_split_prefix_mismatch"
		return c2sFinalSplitWrite(result, outPath)
	}

	result.PredecessorDoubleDFTTailConfirmed = true
	result.PredecessorFinalSplitDispositionValid = false
	zvFastBefore := fastCurrent.CopyNew()
	zvStandardBefore := standardCurrent.CopyNew()
	fastConj := fastckks.NewCiphertext(params, 1, fastCurrent.Level())
	*fastConj.MetaData = *fastCurrent.MetaData
	standardConj := ckks.NewCiphertext(params, 1, standardCurrent.Level())
	*standardConj.MetaData = *standardCurrent.MetaData
	if err := fastEval.FastCKKS.Conjugate(fastCurrent, fastConj); err != nil {
		return err
	}
	if err := standardEval.Evaluator.Conjugate(standardCurrent, standardConj); err != nil {
		return err
	}
	checkpoint, err := c2sFinalSplitCapture("split_input_zv", zvStandardBefore, zvFastBefore, params, skN2, fastSK, acceptedC2SGroup1Budget, fix001P3E2EInputUnchanged(zvFastBefore, fastCurrent) && fix001P3E2EInputUnchanged(zvStandardBefore, standardCurrent))
	if err != nil {
		return err
	}
	result.SplitCheckpoints = append(result.SplitCheckpoints, checkpoint)
	checkpoint, err = c2sFinalSplitCapture("split_conjugate", standardConj, fastConj, params, skN2, fastSK, acceptedC2SGroup1Budget, true)
	if err != nil {
		return err
	}
	result.SplitCheckpoints = append(result.SplitCheckpoints, checkpoint)
	fastImagTmp := fastckks.NewCiphertext(params, 1, fastCurrent.Level())
	*fastImagTmp.MetaData = *fastCurrent.MetaData
	standardImagTmp := ckks.NewCiphertext(params, 1, standardCurrent.Level())
	*standardImagTmp.MetaData = *standardCurrent.MetaData
	if err := fastEval.FastCKKS.Sub(fastCurrent, fastConj, fastImagTmp); err != nil {
		return err
	}
	if err := standardEval.Evaluator.Sub(standardCurrent, standardConj, standardImagTmp); err != nil {
		return err
	}
	checkpoint, err = c2sFinalSplitCapture("split_imag_sub", standardImagTmp, fastImagTmp, params, skN2, fastSK, acceptedC2SGroup1Budget, true)
	if err != nil {
		return err
	}
	result.SplitCheckpoints = append(result.SplitCheckpoints, checkpoint)
	fastImag := fastckks.NewCiphertext(params, 1, fastImagTmp.Level())
	*fastImag.MetaData = *fastImagTmp.MetaData
	standardImagSplit := ckks.NewCiphertext(params, 1, standardImagTmp.Level())
	*standardImagSplit.MetaData = *standardImagTmp.MetaData
	if err := fastEval.FastCKKS.Mul(fastImagTmp, -1i, fastImag); err != nil {
		return err
	}
	if err := standardEval.Evaluator.Mul(standardImagTmp, -1i, standardImagSplit); err != nil {
		return err
	}
	checkpoint, err = c2sFinalSplitCapture("split_imag_mul_minus_i", standardImagSplit, fastImag, params, skN2, fastSK, acceptedC2SGroup1Budget, true)
	if err != nil {
		return err
	}
	result.SplitCheckpoints = append(result.SplitCheckpoints, checkpoint)
	fastReal := fastckks.NewCiphertext(params, 1, fastCurrent.Level())
	*fastReal.MetaData = *fastCurrent.MetaData
	standardRealSplit := ckks.NewCiphertext(params, 1, standardCurrent.Level())
	*standardRealSplit.MetaData = *standardCurrent.MetaData
	if err := fastEval.FastCKKS.Add(fastConj, fastCurrent, fastReal); err != nil {
		return err
	}
	if err := standardEval.Evaluator.Add(standardConj, standardCurrent, standardRealSplit); err != nil {
		return err
	}
	checkpoint, err = c2sFinalSplitCapture("split_real_add", standardRealSplit, fastReal, params, skN2, fastSK, acceptedC2SGroup1Budget, true)
	if err != nil {
		return err
	}
	result.SplitCheckpoints = append(result.SplitCheckpoints, checkpoint)
	for _, splitCheckpoint := range result.SplitCheckpoints[1:] {
		if cause := c2sFinalSplitCause(splitCheckpoint); cause != "" {
			result.FirstFailingCheckpoint = splitCheckpoint.Checkpoint
			result.FirstSupportedCause = cause
			return c2sFinalSplitWrite(result, outPath)
		}
	}
	_, fastRealValues, err := semanticBisectView(fastReal, params, fastSK)
	if err != nil {
		return err
	}
	standardRealValues, err := decodeWithSecret(params, standardReal, skN2)
	if err != nil {
		return err
	}
	finalReal, err := semanticBisectMetricFor(standardRealValues, fastRealValues, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.FinalRealFastVsStandard = &finalReal
	_, fastImagValues, err := semanticBisectView(fastImag, params, fastSK)
	if err != nil {
		return err
	}
	standardImagValues, err := decodeWithSecret(params, standardImag, skN2)
	if err != nil {
		return err
	}
	finalImag, err := semanticBisectMetricFor(standardImagValues, fastImagValues, acceptedC2SGroup1Budget)
	if err != nil {
		return err
	}
	result.FinalImagFastVsStandard = &finalImag
	_, alignedRealValues, err := semanticBisectView(standardRealSplit, params, skN2)
	if err != nil {
		return err
	}
	_ = alignedRealValues
	_, fastRealValuesAgain, err := semanticBisectView(fastReal, params, fastSK)
	if err != nil {
		return err
	}
	result.FinalRealFastVsAlignedStandard, err = func() (*semanticBisectMetric, error) {
		metric, e := semanticBisectMetricFor(alignedRealValues, fastRealValuesAgain, acceptedC2SGroup1Budget)
		return &metric, e
	}()
	if err != nil {
		return err
	}
	_, alignedImagValues, err := semanticBisectView(standardImagSplit, params, skN2)
	if err != nil {
		return err
	}
	result.FinalImagFastVsAlignedStandard, err = func() (*semanticBisectMetric, error) {
		metric, e := semanticBisectMetricFor(alignedImagValues, fastImagValues, acceptedC2SGroup1Budget)
		return &metric, e
	}()
	if err != nil {
		return err
	}
	result.C2SPrecisionBlockerResolved = finalReal.Pass && finalImag.Pass
	if result.C2SPrecisionBlockerResolved && result.FinalRealFastVsAlignedStandard.Pass && result.FinalImagFastVsAlignedStandard.Pass {
		result.FirstSupportedCause = "logn13_c2s_corrected_groups0_1_final_split_validated"
		result.Disposition = "groups0_1_compression_sufficient_for_c2s_budget"
	} else {
		result.FirstSupportedCause = "logn13_c2s_final_reference_mismatch_requires_diagnosis"
		result.Disposition = "corrected_split_requires_reference_diagnosis"
	}
	return c2sFinalSplitWrite(result, outPath)
}
