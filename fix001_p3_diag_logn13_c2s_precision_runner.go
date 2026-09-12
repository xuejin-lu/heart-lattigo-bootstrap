package main

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	commonlintrans "github.com/tuneinsight/lattigo/v6/circuits/common/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredC2SPrecisionPrimaryBase = "c0a058ac2e5caafa9cdc9eb63e68cd43269c329f"
	requiredC2SPrecisionSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
	c2sPrecisionInputThreshold      = 1e-9
	c2sPrecisionFinalThreshold      = 1e-2
)

type c2sPrecisionCheckpoint struct {
	Checkpoint                string                 `json:"checkpoint"`
	Fast                      semanticBisectMetadata `json:"fast"`
	Standard                  semanticBisectMetadata `json:"standard"`
	FastVsStandard            semanticBisectMetric   `json:"fast_vs_standard"`
	FastVsCommonModUp         *semanticBisectMetric  `json:"fast_vs_common_modup,omitempty"`
	StandardVsCommonModUp     *semanticBisectMetric  `json:"standard_vs_common_modup,omitempty"`
	FastQ01Capacity           *postMod1S2CCapacity   `json:"fast_q01_capacity,omitempty"`
	StandardFullRNSCapacity   *postMod1S2CCapacity   `json:"standard_full_rns_capacity,omitempty"`
	StandardQ01ProjectionSafe bool                   `json:"standard_q01_projection_safe"`
}

type c2sPrecisionGroup struct {
	Group                  int                    `json:"group"`
	MatrixIndex            int                    `json:"matrix_index"`
	MatrixScale            string                 `json:"matrix_scale"`
	MatrixScaleLog2        float64                `json:"matrix_scale_log2"`
	MatrixLevelQ           int                    `json:"matrix_level_q"`
	MatrixLevelP           int                    `json:"matrix_level_p"`
	DroppedModulus         string                 `json:"dropped_logical_modulus"`
	DroppedModulusBits     int                    `json:"dropped_logical_modulus_bits"`
	Input                  c2sPrecisionCheckpoint `json:"input"`
	AfterLinearTransform   c2sPrecisionCheckpoint `json:"after_linear_transform"`
	AfterRescale           c2sPrecisionCheckpoint `json:"after_rescale"`
	LinearIncrementMax     float64                `json:"linear_increment_max_component"`
	RescaleIncrementMax    float64                `json:"rescale_increment_max_component"`
	FirstBudgetBreachPoint string                 `json:"first_budget_breach_point"`
}

type c2sPrecisionResult struct {
	SchemaVersion                       string                       `json:"schema_version"`
	Timestamp                           time.Time                    `json:"timestamp"`
	Primary                             RepositoryMetadata           `json:"primary_repository"`
	Lattigo                             RepositoryMetadata           `json:"lattigo_repository"`
	Environment                         EnvironmentMetadata          `json:"environment"`
	Config                              BootstrapConfig              `json:"config"`
	Parameters                          ExperimentParameters         `json:"effective_parameters"`
	Workload                            CorrectnessWorkload          `json:"workload"`
	PreviousRawStandardSquareProjection string                       `json:"previous_raw_standard_square_projection_disposition"`
	MatchedInputAsymmetry               bool                         `json:"matched_input_asymmetry"`
	DownstreamEvalModIssueUnresolved    bool                         `json:"downstream_evalmod_issue_unresolved"`
	StandardProof                       semanticBisectStandardProof  `json:"standard_control_proof"`
	StandardEndToEnd                    *semanticBisectMetric        `json:"standard_end_to_end_semantic,omitempty"`
	ModUp                               *c2sPrecisionCheckpoint      `json:"modup_precondition,omitempty"`
	ModUpMetadataMatch                  bool                         `json:"modup_metadata_match"`
	ModUpNoInputMutation                bool                         `json:"modup_no_input_mutation"`
	MatrixLevels                        []int                        `json:"c2s_matrix_levels"`
	MatrixScheduleValid                 bool                         `json:"c2s_matrix_schedule_valid"`
	StandardFastReencode                *evalModMatchedInputEvidence `json:"standard_from_fast_c2s_real"`
	StandardNativeVsStandardOnFast      *semanticBisectMetric        `json:"standard_native_vs_standard_on_fast"`
	C2SRealInputDelta                   *semanticBisectMetric        `json:"c2s_real_input_delta"`
	C2SImagInputDelta                   *semanticBisectMetric        `json:"c2s_imag_input_delta,omitempty"`
	Amplification                       float64                      `json:"derived_amplification"`
	C2SBudget                           float64                      `json:"derived_c2s_budget"`
	CausalSufficiency                   bool                         `json:"c2s_error_alone_sufficient_for_final_failure"`
	Groups                              []c2sPrecisionGroup          `json:"groups"`
	FinalReal                           *c2sPrecisionCheckpoint      `json:"final_real_split,omitempty"`
	FinalImag                           *c2sPrecisionCheckpoint      `json:"final_imag_split,omitempty"`
	FirstFailingCheckpoint              string                       `json:"first_failing_checkpoint"`
	FirstSupportedCause                 string                       `json:"first_supported_cause"`
	Validation                          map[string]interface{}       `json:"validation"`
}

func c2sPrecisionWrite(result c2sPrecisionResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func c2sPrecisionFullView(source *rlwe.Ciphertext, params ckks.Parameters, sk *rlwe.SecretKey) (semanticBisectMetadata, []complex128, error) {
	decoded, err := decodeWithSecret(params, source, sk)
	if err != nil {
		return semanticBisectMetadata{}, nil, err
	}
	return semanticBisectMetadata{
		SourceLevel: source.Level(), DecodeLevel: source.Level(), Scale: finalizationScaleString(source.Scale), ScaleLog2: source.Scale.Log2(), Degree: source.Degree(), N: source.N(), LogDimensions: source.LogDimensions,
		IsNTT: source.IsNTT, IsMontgomery: source.IsMontgomery, ProjectedMontgomery: source.IsMontgomery,
	}, decoded, nil
}

func c2sPrecisionCapture(name string, fast, standard *rlwe.Ciphertext, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, common []complex128, budget float64) (c2sPrecisionCheckpoint, error) {
	fastMetadata, fastValues, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return c2sPrecisionCheckpoint{}, err
	}
	standardMetadata, standardValues, err := c2sPrecisionFullView(standard, params, standardSK)
	if err != nil {
		return c2sPrecisionCheckpoint{}, err
	}
	metric, err := semanticBisectMetricFor(standardValues, fastValues, budget)
	if err != nil {
		return c2sPrecisionCheckpoint{}, err
	}
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
	if err != nil {
		return c2sPrecisionCheckpoint{}, err
	}
	standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standard)
	if err != nil {
		return c2sPrecisionCheckpoint{}, err
	}
	checkpoint := c2sPrecisionCheckpoint{
		Checkpoint: name, Fast: fastMetadata, Standard: standardMetadata, FastVsStandard: metric,
		FastQ01Capacity: &fastCapacity, StandardFullRNSCapacity: &standardCapacity, StandardQ01ProjectionSafe: standardCapacity.Pass,
	}
	if common != nil {
		fastVsCommon, err := semanticBisectMetricFor(common, fastValues, budget)
		if err != nil {
			return c2sPrecisionCheckpoint{}, err
		}
		standardVsCommon, err := semanticBisectMetricFor(common, standardValues, budget)
		if err != nil {
			return c2sPrecisionCheckpoint{}, err
		}
		checkpoint.FastVsCommonModUp = &fastVsCommon
		checkpoint.StandardVsCommonModUp = &standardVsCommon
	}
	return checkpoint, nil
}

func c2sPrecisionScheduleValid(fast, standard dft.Matrix) bool {
	if len(fast.Levels) != 4 || len(standard.Levels) != 4 || len(fast.Matrices) != 4 || len(standard.Matrices) != 4 {
		return false
	}
	for i := range fast.Levels {
		if fast.Levels[i] != 1 || standard.Levels[i] != 1 || fast.Levels[i] != standard.Levels[i] {
			return false
		}
	}
	return true
}

func c2sPrecisionClassification(group int, point string) string {
	return fmt.Sprintf("logn13_c2s_precision_first_breach_group%d_%s", group, point)
}

func c2sPrecisionReplay(params ckks.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, fastInput, standardInput *rlwe.Ciphertext, fastSK, standardSK *rlwe.SecretKey, common []complex128, budget float64) ([]c2sPrecisionGroup, string, error) {
	fastCurrent := fastInput.CopyNew()
	standardCurrent := standardInput.CopyNew()
	fastTransform := fastEval.DFTEvaluator.FastEvaluator()
	groups := make([]c2sPrecisionGroup, 0, len(fastEval.C2SDFTMatrix.Levels))
	matrixIndex := 0
	for groupIndex, factorCount := range fastEval.C2SDFTMatrix.Levels {
		if factorCount != 1 || matrixIndex >= len(fastEval.C2SDFTMatrix.Matrices) || matrixIndex >= len(standardEval.C2SDFTMatrix.Matrices) {
			return groups, "", fmt.Errorf("C2S matrix schedule mismatch at group %d", groupIndex)
		}
		matrix := fastEval.C2SDFTMatrix.Matrices[matrixIndex]
		standardMatrix := standardEval.C2SDFTMatrix.Matrices[matrixIndex]
		beforeLevel := fastCurrent.Level()
		group := c2sPrecisionGroup{
			Group: groupIndex, MatrixIndex: matrixIndex, MatrixScale: finalizationScaleString(matrix.Scale), MatrixScaleLog2: matrix.Scale.Log2(), MatrixLevelQ: matrix.LevelQ, MatrixLevelP: matrix.LevelP,
			DroppedModulus: fmt.Sprintf("%d", params.Q()[beforeLevel]), DroppedModulusBits: bits.Len64(params.Q()[beforeLevel]),
		}
		var err error
		group.Input, err = c2sPrecisionCapture(fmt.Sprintf("group_%d_input", groupIndex), fastCurrent, standardCurrent, params, fastSK, standardSK, common, budget)
		if err != nil {
			return groups, "", err
		}
		fastNext := fastckks.NewCiphertext(params, 1, fastCurrent.Level())
		*fastNext.MetaData = *fastCurrent.MetaData
		if err = fastTransform.LinearTransform(fastCurrent, commonlintrans.LinearTransformation(matrix), fastNext); err != nil {
			return groups, "", err
		}
		standardNext := ckks.NewCiphertext(params, 1, standardCurrent.Level())
		*standardNext.MetaData = *standardCurrent.MetaData
		if err = standardEval.DFTEvaluator.LTEvaluator.Evaluate(standardCurrent, standardMatrix, standardNext); err != nil {
			return groups, "", err
		}
		group.AfterLinearTransform, err = c2sPrecisionCapture(fmt.Sprintf("group_%d_after_linear_transform", groupIndex), fastNext, standardNext, params, fastSK, standardSK, common, budget)
		if err != nil {
			return groups, "", err
		}
		if err = fastTransform.Rescale(fastNext, fastNext); err != nil {
			return groups, "", err
		}
		if err = standardEval.DFTEvaluator.Rescale(standardNext, standardNext); err != nil {
			return groups, "", err
		}
		group.AfterRescale, err = c2sPrecisionCapture(fmt.Sprintf("group_%d_after_rescale", groupIndex), fastNext, standardNext, params, fastSK, standardSK, common, budget)
		if err != nil {
			return groups, "", err
		}
		group.LinearIncrementMax = group.AfterLinearTransform.FastVsStandard.MaxComponentAbs - group.Input.FastVsStandard.MaxComponentAbs
		group.RescaleIncrementMax = group.AfterRescale.FastVsStandard.MaxComponentAbs - group.AfterLinearTransform.FastVsStandard.MaxComponentAbs
		if group.AfterLinearTransform.FastVsStandard.MaxComponentAbs > budget {
			group.FirstBudgetBreachPoint = "linear_transform"
		} else if group.AfterRescale.FastVsStandard.MaxComponentAbs > budget {
			group.FirstBudgetBreachPoint = "rescale"
		}
		groups = append(groups, group)
		if group.FirstBudgetBreachPoint != "" {
			return groups, c2sPrecisionClassification(groupIndex, group.FirstBudgetBreachPoint), nil
		}
		fastCurrent, standardCurrent = fastNext, standardNext
		matrixIndex++
	}
	if matrixIndex != len(fastEval.C2SDFTMatrix.Matrices) {
		return groups, "", fmt.Errorf("C2S matrix factorization has unused matrices")
	}
	return groups, "", nil
}

func runFIX001P3DiagLogN13C2SPrecision(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := c2sPrecisionResult{
		SchemaVersion: "fix-001-p3-diag-logn13-c2s-precision-budget.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		PreviousRawStandardSquareProjection: "invalid_as_q01_oracle_after_centered_capacity_loss", MatchedInputAsymmetry: false, DownstreamEvalModIssueUnresolved: true,
		StandardProof:          semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: standardEval.Evaluator == nil && standardEval.DFTEvaluator == nil && standardEval.Mod1Evaluator == nil, OrdinaryStageCallsUsed: true},
		FirstFailingCheckpoint: "none", FirstSupportedCause: "logn13_c2s_precision_standard_control_failure",
		Validation: map[string]interface{}{"primary_required_base": requiredC2SPrecisionPrimaryBase, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredC2SPrecisionSecondary, "no_secondary_production_changes": true, "no_c2s_parameter_or_matrix_changes": true, "no_evalmod_fix": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_unpack_or_finalization": true},
	}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil {
		return c2sPrecisionWrite(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return c2sPrecisionWrite(result, outPath)
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
		return c2sPrecisionWrite(result, outPath)
	}
	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastInputBefore := fastInput.CopyNew()
	standardInputBefore := standardInput.CopyNew()
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
	fastModUpMetadata, fastModUpValues, err := semanticBisectView(fastModUp, params, fastSK)
	if err != nil {
		return err
	}
	standardModUpMetadata, standardModUpValues, err := c2sPrecisionFullView(standardModUp, params, skN2)
	if err != nil {
		return err
	}
	modUpMetric, err := semanticBisectMetricFor(standardModUpValues, fastModUpValues, c2sPrecisionInputThreshold)
	if err != nil {
		return err
	}
	fastInputUnchanged := fix001P3E2EInputUnchanged(fastInputBefore, fastInput)
	standardInputUnchanged := fix001P3E2EInputUnchanged(standardInputBefore, standardInput)
	result.ModUpNoInputMutation = fastInputUnchanged && standardInputUnchanged
	result.ModUpMetadataMatch = fastModUpMetadata.SourceLevel == standardModUpMetadata.SourceLevel && fastModUpMetadata.Scale == standardModUpMetadata.Scale && fastModUpMetadata.Degree == standardModUpMetadata.Degree && fastModUpMetadata.N == standardModUpMetadata.N && fastModUpMetadata.LogDimensions == standardModUpMetadata.LogDimensions && fastModUpMetadata.IsNTT == standardModUpMetadata.IsNTT && fastModUpMetadata.IsMontgomery && !standardModUpMetadata.IsMontgomery
	modUpPoint := c2sPrecisionCheckpoint{Checkpoint: "mod_up", Fast: fastModUpMetadata, Standard: standardModUpMetadata, FastVsStandard: modUpMetric}
	result.ModUp = &modUpPoint
	if !result.ModUpMetadataMatch || !result.ModUpNoInputMutation || !modUpMetric.Pass {
		result.FirstFailingCheckpoint = "mod_up"
		result.FirstSupportedCause = "logn13_c2s_precision_modup_precondition_failure"
		return c2sPrecisionWrite(result, outPath)
	}

	fastReal, fastImag, err := fastEval.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	if !c2sPrecisionScheduleValid(fastEval.C2SDFTMatrix, standardEval.C2SDFTMatrix) {
		result.FirstFailingCheckpoint = "c2s_matrix_schedule"
		result.FirstSupportedCause = "logn13_c2s_precision_modup_precondition_failure"
		return c2sPrecisionWrite(result, outPath)
	}
	result.MatrixLevels = append([]int(nil), fastEval.C2SDFTMatrix.Levels...)
	result.MatrixScheduleValid = true
	_, fastRealValues, err := semanticBisectView(fastReal, params, fastSK)
	if err != nil {
		return err
	}
	standardFromFast, err := evalModStandardFromFast(params, fastReal, fastRealValues, skN2)
	if err != nil {
		return err
	}
	standardFastEvidence, err := evalModMatchedInput("standard_from_fast_c2s_real", fastReal, standardFromFast, params, fastSK, skN2)
	if err != nil {
		return err
	}
	result.StandardFastReencode = &standardFastEvidence
	if !standardFastEvidence.Pass {
		result.FirstSupportedCause = "logn13_c2s_precision_budget_construction_failure"
		return c2sPrecisionWrite(result, outPath)
	}
	standardNativeEvalMod, err := standardEval.EvalMod(standardReal.CopyNew())
	if err != nil {
		return err
	}
	standardOnFastEvalMod, err := standardEval.EvalMod(standardFromFast.CopyNew())
	if err != nil {
		return err
	}
	standardNativeValues, err := decodeWithSecret(params, standardNativeEvalMod, skN2)
	if err != nil {
		return err
	}
	standardOnFastValues, err := decodeWithSecret(params, standardOnFastEvalMod, skN2)
	if err != nil {
		return err
	}
	standardSensitivity, err := semanticBisectMetricFor(standardNativeValues, standardOnFastValues, c2sPrecisionFinalThreshold)
	if err != nil {
		return err
	}
	result.StandardNativeVsStandardOnFast = &standardSensitivity
	standardRealValues, err := decodeWithSecret(params, standardReal, skN2)
	if err != nil {
		return err
	}
	realInputDelta, err := semanticBisectMetricFor(standardRealValues, fastRealValues, 1)
	if err != nil {
		return err
	}
	result.C2SRealInputDelta = &realInputDelta
	if fastImag != nil && standardImag != nil {
		fastImagValues, decodeErr := func() ([]complex128, error) {
			_, values, err := semanticBisectView(fastImag, params, fastSK)
			return values, err
		}()
		if decodeErr != nil {
			return decodeErr
		}
		standardImagValues, decodeErr := decodeWithSecret(params, standardImag, skN2)
		if decodeErr != nil {
			return decodeErr
		}
		imagInputDelta, metricErr := semanticBisectMetricFor(standardImagValues, fastImagValues, 1)
		if metricErr != nil {
			return metricErr
		}
		result.C2SImagInputDelta = &imagInputDelta
	}
	if realInputDelta.MaxComponentAbs > 0 {
		result.Amplification = standardSensitivity.MaxComponentAbs / realInputDelta.MaxComponentAbs
		result.C2SBudget = c2sPrecisionFinalThreshold / result.Amplification
	}
	if result.Amplification <= 0 || result.C2SBudget <= 0 || standardFastEvidence.ReconstructionMetric.MaxComponentAbs > 1e-6 {
		result.FirstSupportedCause = "logn13_c2s_precision_budget_construction_failure"
		return c2sPrecisionWrite(result, outPath)
	}
	result.CausalSufficiency = standardSensitivity.MaxComponentAbs > c2sPrecisionFinalThreshold && realInputDelta.MaxComponentAbs > result.C2SBudget

	groups, classification, err := c2sPrecisionReplay(params, fastEval, standardEval, fastModUp, standardModUp, fastSK, skN2, standardModUpValues, result.C2SBudget)
	if err != nil {
		return err
	}
	result.Groups = groups
	result.FirstSupportedCause = classification
	if classification == "" {
		finalReal, err := c2sPrecisionCapture("real_imag_split_real", fastReal, standardReal, params, fastSK, skN2, standardModUpValues, result.C2SBudget)
		if err != nil {
			return err
		}
		result.FinalReal = &finalReal
		if fastImag != nil && standardImag != nil {
			finalImag, err := c2sPrecisionCapture("real_imag_split_imag", fastImag, standardImag, params, fastSK, skN2, standardModUpValues, result.C2SBudget)
			if err != nil {
				return err
			}
			result.FinalImag = &finalImag
		}
		if finalReal.FastVsStandard.MaxComponentAbs > result.C2SBudget || result.FinalImag != nil && result.FinalImag.FastVsStandard.MaxComponentAbs > result.C2SBudget {
			result.FirstFailingCheckpoint = "real_imag_split"
			result.FirstSupportedCause = "logn13_c2s_precision_first_breach_real_imag_split"
		} else {
			result.FirstSupportedCause = "logn13_c2s_precision_budget_satisfied"
		}
	}
	if result.FirstFailingCheckpoint == "none" && result.FirstSupportedCause != "logn13_c2s_precision_budget_satisfied" {
		for _, group := range result.Groups {
			if group.FirstBudgetBreachPoint != "" {
				result.FirstFailingCheckpoint = fmt.Sprintf("group_%d_%s", group.Group, group.FirstBudgetBreachPoint)
				break
			}
		}
	}
	return c2sPrecisionWrite(result, outPath)
}
