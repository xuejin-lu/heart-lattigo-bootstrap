package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	requiredFIX001P3PrecisionSweepPrimaryBase = "72cc6a44d4ce25b11f1e77cbd77d018f21e4aa68"
	requiredFIX001P3PrecisionSweepSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	precisionSweepPublicThreshold             = 1e-2
	precisionSweepLocalTarget                 = 1e-4
	precisionSweepPolynomialTarget            = 2e-8
	precisionSweepC2SBudget                   = 1.9531249403614602e-5
)

var precisionSweepPlanExponents = []int{91, 92, 93, 94, 95, 96, 97, 98, 99, 100}

type precisionSweepCapacity struct {
	MinimumRatio      float64 `json:"minimum_capacity_ratio"`
	WorstCheckpoint   string  `json:"worst_capacity_checkpoint"`
	OutsideCount      int     `json:"outside_count"`
	AllCenteredUnique bool    `json:"all_centered_unique"`
	AllRowsMatch      bool    `json:"all_fast_vs_full_rns_rows_match"`
}

type precisionSweepRound struct {
	Round              int `json:"round"`
	KInExponent        int `json:"k_in_exponent"`
	KOutExponent       int `json:"k_out_exponent"`
	MultiplierExponent int `json:"multiplier_exponent"`
}

type precisionSweepCandidate struct {
	Exponent                  int                    `json:"plan_scale_exponent"`
	PlanScale                 string                 `json:"plan_scale"`
	PolynomialOutputScale     string                 `json:"polynomial_output_scale,omitempty"`
	PolynomialCapacityRatio   float64                `json:"polynomial_capacity_ratio,omitempty"`
	PolynomialCapacityPass    bool                   `json:"polynomial_capacity_pass"`
	PolynomialVsStandard      *PSGlobalMetric        `json:"compressed_polynomial_vs_standard,omitempty"`
	PolynomialTargetPass      bool                   `json:"polynomial_design_target_pass"`
	Capacity                  precisionSweepCapacity `json:"capacity"`
	KSchedule                 []precisionSweepRound  `json:"derived_k_schedule,omitempty"`
	FinalRestoreKExponent     string                 `json:"final_restore_k_exponent,omitempty"`
	RealEvalMod               *PSGlobalMetric        `json:"real_evalmod_vs_standard,omitempty"`
	ImagEvalMod               *PSGlobalMetric        `json:"imag_evalmod_vs_standard,omitempty"`
	RealNormalizedEffect      *PSGlobalMetric        `json:"real_fast_vs_normalized,omitempty"`
	ImagNormalizedEffect      *PSGlobalMetric        `json:"imag_fast_vs_normalized,omitempty"`
	CoherentRealVsStandard    *PSGlobalMetric        `json:"real_coherent_vs_standard,omitempty"`
	CoherentImagVsStandard    *PSGlobalMetric        `json:"imag_coherent_vs_standard,omitempty"`
	PostS2CImplementation     *PSGlobalMetric        `json:"post_s2c_implementation_effect,omitempty"`
	PostS2CVsStandard         *PSGlobalMetric        `json:"post_s2c_vs_standard,omitempty"`
	PublicLike                *PSGlobalMetric        `json:"public_like_vs_message,omitempty"`
	StandardPublicLike        *PSGlobalMetric        `json:"standard_public_like_vs_message,omitempty"`
	PublicLikeMetadataCorrect bool                   `json:"public_like_metadata_correct"`
	PublicLikeInputUnchanged  bool                   `json:"public_like_input_unchanged"`
	PrecisionQualified        bool                   `json:"precision_qualified"`
	PublicLikeAccepted        bool                   `json:"public_like_accepted"`
	FirstFailingCheckpoint    string                 `json:"first_failing_checkpoint"`
	Classification            string                 `json:"classification"`
}

type precisionSweepResult struct {
	SchemaVersion             string                    `json:"schema_version"`
	Timestamp                 time.Time                 `json:"timestamp"`
	Primary                   RepositoryMetadata        `json:"primary_repository"`
	Lattigo                   RepositoryMetadata        `json:"lattigo_repository"`
	Environment               EnvironmentMetadata       `json:"environment"`
	Config                    BootstrapConfig           `json:"config"`
	Parameters                ExperimentParameters      `json:"effective_parameters"`
	Workload                  CorrectnessWorkload       `json:"workload"`
	PublicThreshold           float64                   `json:"public_threshold"`
	MeasuredAmplification     float64                   `json:"measured_current_s2c_amplification"`
	DerivedLocalBudget        float64                   `json:"derived_no_margin_local_evalmod_budget"`
	FixedLocalTarget          float64                   `json:"fixed_evalmod_local_design_target"`
	FixedPolynomialTarget     float64                   `json:"fixed_polynomial_design_target"`
	CurrentC2SReal            *PSGlobalMetric           `json:"current_c2s_real,omitempty"`
	CurrentC2SImag            *PSGlobalMetric           `json:"current_c2s_imag,omitempty"`
	CurrentEvalModReal        *PSGlobalMetric           `json:"current_evalmod_real,omitempty"`
	CurrentEvalModImag        *PSGlobalMetric           `json:"current_evalmod_imag,omitempty"`
	CurrentPolynomialDelta    *PSGlobalMetric           `json:"current_polynomial_vs_standard,omitempty"`
	CurrentCoherentRealDelta  *PSGlobalMetric           `json:"current_coherent_real_vs_standard,omitempty"`
	CurrentCoherentImagDelta  *PSGlobalMetric           `json:"current_coherent_imag_vs_standard,omitempty"`
	CurrentPolyToEvalModRatio float64                   `json:"current_polynomial_to_evalmod_amplification"`
	DerivedPolynomialBudget   float64                   `json:"derived_polynomial_delta_budget_for_local_target"`
	StandardPublic            *PSGlobalMetric           `json:"genuine_standard_public,omitempty"`
	FastPublic                *PSGlobalMetric           `json:"current_fast_public,omitempty"`
	FastPublicMetadataCorrect bool                      `json:"current_fast_public_metadata_correct"`
	Candidates                []precisionSweepCandidate `json:"candidates"`
	PreferredExponent         int                       `json:"preferred_plan_scale_exponent,omitempty"`
	Classification            string                    `json:"classification"`
	FirstLimitingCheckpoint   string                    `json:"first_limiting_checkpoint"`
	Validation                map[string]interface{}    `json:"validation"`
}

func precisionSweepWrite(result precisionSweepResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func precisionSweepScale(exponent int) rlwe.Scale {
	return rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(exponent)))
}

func precisionSweepCapacityAccumulator() precisionSweepCapacity {
	return precisionSweepCapacity{MinimumRatio: math.Inf(1), AllCenteredUnique: true, AllRowsMatch: true}
}

func precisionSweepAddCapacity(out *precisionSweepCapacity, checkpoint string, ratio float64, outside int, pass, rowsMatch bool) {
	if ratio < out.MinimumRatio {
		out.MinimumRatio = ratio
		out.WorstCheckpoint = checkpoint
	}
	out.OutsideCount += outside
	out.AllCenteredUnique = out.AllCenteredUnique && pass
	out.AllRowsMatch = out.AllRowsMatch && rowsMatch
}

func precisionSweepAddCheckpoint(out *precisionSweepCapacity, name string, checkpoint evalModMatchedCheckpoint) {
	if checkpoint.FastQ01Capacity != nil {
		precisionSweepAddCapacity(out, name+".fast_q01", checkpoint.FastQ01Capacity.MaxAbsOverQ01Half, checkpoint.FastQ01Capacity.OutsideCount, checkpoint.FastQ01Capacity.Pass, checkpoint.RowsMatch)
	}
	if checkpoint.NormalizedCapacity != nil {
		precisionSweepAddCapacity(out, name+".full_rns", checkpoint.NormalizedCapacity.Ratio, checkpoint.NormalizedCapacity.OutsideCount, checkpoint.NormalizedCapacity.Pass, checkpoint.RowsMatch)
	}
}

func precisionSweepPathEvidence(path evalModMatchedPath) (precisionSweepCapacity, []precisionSweepRound, string) {
	out := precisionSweepCapacityAccumulator()
	for _, checkpoint := range path.Prefix {
		precisionSweepAddCheckpoint(&out, checkpoint.Name, checkpoint)
	}
	for _, round := range path.Rounds {
		precisionSweepAddCheckpoint(&out, fmt.Sprintf("round%d.input", round.Round), round.Input)
		precisionSweepAddCheckpoint(&out, fmt.Sprintf("round%d.square", round.Round), round.Square)
		precisionSweepAddCheckpoint(&out, fmt.Sprintf("round%d.after_multiplier", round.Round), round.AfterMultiplier)
		precisionSweepAddCheckpoint(&out, fmt.Sprintf("round%d.after_constant", round.Round), round.AfterConstant)
		precisionSweepAddCheckpoint(&out, fmt.Sprintf("round%d.post_rescale", round.Round), round.PostRescale)
	}
	if path.PolyCapacity != nil {
		precisionSweepAddCapacity(&out, "polynomial.normalized_full_rns", path.PolyCapacity.Ratio, path.PolyCapacity.OutsideCount, path.PolyCapacity.Pass, path.FastRowsMatch)
	}
	if path.FinalRestoreCapacity != nil {
		precisionSweepAddCapacity(&out, "final_restore", path.FinalRestoreCapacity.Ratio, path.FinalRestoreCapacity.OutsideCount, path.FinalRestoreCapacity.Pass, path.RestoreRowsMatch)
	}
	if math.IsInf(out.MinimumRatio, 1) {
		out.MinimumRatio = 0
	}
	schedule := make([]precisionSweepRound, 0, len(path.Rounds))
	for _, round := range path.Rounds {
		schedule = append(schedule, precisionSweepRound{Round: round.Round, KInExponent: round.KInExponent, KOutExponent: round.KOutExponent, MultiplierExponent: round.MultiplierExponent})
	}
	finalK := ""
	if path.FinalRestoreCapacity != nil {
		finalK = path.FinalRestoreCapacity.K3
	}
	return out, schedule, finalK
}

func precisionSweepPathClassification(path evalModMatchedPath, final evalModMatchedFinal) (string, string) {
	if path.FirstFailure != "none" {
		if strings.Contains(path.FirstFailure, "restore_capacity") || strings.Contains(path.FirstFailure, "normalized_capacity") || strings.Contains(path.FirstFailure, "pre_rescale_capacity") {
			return "normalized_candidate_capacity_failure", path.FirstFailure
		}
		if strings.Contains(path.FirstFailure, "polynomial") || strings.Contains(path.FirstFailure, "capacity") {
			return "polynomial_plan_capacity_failure", path.FirstFailure
		}
		return "polynomial_fast_vs_full_rns_mismatch", path.FirstFailure
	}
	if !final.FastRowsMatchNormalized || !final.RestoreRowsMatch {
		return "normalized_candidate_primitive_mismatch", "final.rows"
	}
	if final.FastVsStandard == nil || !final.FastVsStandard.Pass {
		return "polynomial_semantic_precision_failure", "final.fast_vs_standard"
	}
	return "normalized_candidate_pass", "none"
}

func precisionSweepPublicLike(pathReal, pathImag evalModMatchedPath, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params bootstrapping.Parameters, residualN int, skN2 *rlwe.SecretKey, input *rlwe.Ciphertext) (*PSGlobalMetric, *PSGlobalMetric, *PSGlobalMetric, bool, bool, error) {
	fastCore, err := fastEval.SlotsToCoeffs(pathReal.FastFinal.CopyNew(), pathImag.FastFinal.CopyNew())
	if err != nil {
		return nil, nil, nil, false, false, err
	}
	standardCore, err := standardEval.SlotsToCoeffs(pathReal.StandardFinal.CopyNew(), pathImag.StandardFinal.CopyNew())
	if err != nil {
		return nil, nil, nil, false, false, err
	}
	fastCoreValues, err := psGlobalDecode(params.BootstrappingParameters, fastCore)
	if err != nil {
		return nil, nil, nil, false, false, err
	}
	standardCoreValues, err := decodeWithSecret(params.BootstrappingParameters, standardCore, skN2)
	if err != nil {
		return nil, nil, nil, false, false, err
	}
	postS2C := evalModMatchedMetricFor(standardCoreValues, fastCoreValues, precisionSweepPublicThreshold)
	message := reproducibleValues(residualN)
	standardLike := evalModMatchedMetricFor(message, standardCoreValues, precisionSweepPublicThreshold)
	fastBefore := input.CopyNew()
	_, ctxtN1, ctxtN2, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, nil, nil, false, false, err
	}
	unpacked, err := fastEval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*fastCore.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil || len(unpacked) != 1 {
		if err == nil {
			err = fmt.Errorf("expected one unpacked ciphertext, got %d", len(unpacked))
		}
		return nil, standardLike, postS2C, false, false, err
	}
	_, final, err := fix001P3E2EFinalizeOracle(&unpacked[0], params)
	if err != nil {
		return nil, standardLike, postS2C, false, false, err
	}
	finalValues, err := decodeWithSecret(params.ResidualParameters, final, zeroSecret(params.ResidualParameters))
	if err != nil {
		return nil, standardLike, postS2C, false, false, err
	}
	publicLike := evalModMatchedMetricFor(message, finalValues, precisionSweepPublicThreshold)
	publicMetadata := final.Level() == params.ResidualParameters.MaxLevel() && final.Degree() == 1 && final.N() == params.ResidualParameters.N() && final.IsNTT && !final.IsMontgomery && final.Scale.Equal(params.ResidualParameters.DefaultScale())
	inputUnchanged := fix001P3E2EInputUnchanged(fastBefore, input)
	return publicLike, standardLike, postS2C, publicMetadata, inputUnchanged, nil
}

func runFIX001P3DesignLogN13EvalModPrecisionScaleSweep(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := precisionSweepResult{
		SchemaVersion: "fix-001-p3-design-logn13-evalmod-precision-scale-sweep.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		PublicThreshold: precisionSweepPublicThreshold, FixedLocalTarget: precisionSweepLocalTarget, FixedPolynomialTarget: precisionSweepPolynomialTarget,
		Validation:     map[string]interface{}{"primary_required_base": requiredFIX001P3PrecisionSweepPrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3PrecisionSweepSecondary, "no_secondary_production_changes": true, "no_production_fix": true, "no_c2s_change": true, "no_s2c_change": true, "no_parameter_retuning": true, "no_raw_standard_double_angle_q01_oracle": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
		Classification: "logn13_evalmod_precision_sweep_precondition_mismatch",
	}
	if !secondaryClean || secondaryCommit != requiredFIX001P3PrecisionSweepSecondary {
		return precisionSweepWrite(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return precisionSweepWrite(result, outPath)
	}
	standardControlValues, err := decodeWithSecret(residual, standardControl, skN1)
	if err != nil {
		return err
	}
	standardMetric := evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), standardControlValues, precisionSweepPublicThreshold)
	result.StandardPublic = standardMetric
	fastPublic, err := fastEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return precisionSweepWrite(result, outPath)
	}
	fastPublicValues, err := decodeWithSecret(residual, fastPublic, zeroSecret(residual))
	if err != nil {
		return err
	}
	fastMetric := evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), fastPublicValues, precisionSweepPublicThreshold)
	result.FastPublic = fastMetric
	result.FastPublicMetadataCorrect = fastPublic.Level() == residual.MaxLevel() && fastPublic.Degree() == 1 && fastPublic.N() == residual.N() && fastPublic.IsNTT && !fastPublic.IsMontgomery && fastPublic.Scale.Equal(residual.DefaultScale())
	if !standardMetric.Pass || fastMetric.Pass {
		return precisionSweepWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, skN2)
	if err != nil {
		return err
	}
	result.CurrentC2SReal, result.CurrentC2SImag = inputs.FastVsOrdinaryReal, inputs.FastVsOrdinaryImag
	if result.CurrentC2SReal == nil || result.CurrentC2SImag == nil || !result.CurrentC2SReal.Pass || !result.CurrentC2SImag.Pass {
		return precisionSweepWrite(result, outPath)
	}
	params := btp.BootstrappingParameters
	result.Candidates = make([]precisionSweepCandidate, 0, len(precisionSweepPlanExponents))
	var baselineReal, baselineImag evalModMatchedPath
	var baselinePublicLike *PSGlobalMetric
	for _, exponent := range precisionSweepPlanExponents {
		planScale := precisionSweepScale(exponent)
		candidate := precisionSweepCandidate{Exponent: exponent, PlanScale: finalizationScaleString(planScale), Classification: "polynomial_plan_capacity_failure", FirstFailingCheckpoint: "none"}
		realPath, realErr := evalModMatchedRunPathWithPlanScale(inputs.FastReal, inputs.OrdinaryReal, fastEval, standardEval, params, inputs.FastSK, inputs.StandardSK, planScale)
		if realErr != nil {
			return realErr
		}
		imagPath, imagErr := evalModMatchedRunPathWithPlanScale(inputs.FastImag, inputs.OrdinaryImag, fastEval, standardEval, params, inputs.FastSK, inputs.StandardSK, planScale)
		if imagErr != nil {
			return imagErr
		}
		if exponent == 91 {
			baselineReal, baselineImag = realPath, imagPath
		}
		realFinal := evalModMatchedFinal{}
		if realPath.FastFinal != nil {
			realFinal, err = evalModMatchedFinalEvidence(realPath, params, inputs.FastSK, inputs.StandardSK)
			if err != nil {
				return err
			}
		}
		imagFinal := evalModMatchedFinal{}
		if imagPath.FastFinal != nil {
			imagFinal, err = evalModMatchedFinalEvidence(imagPath, params, inputs.FastSK, inputs.StandardSK)
			if err != nil {
				return err
			}
		}
		realCapacity, realSchedule, realFinalK := precisionSweepPathEvidence(realPath)
		imagCapacity, _, _ := precisionSweepPathEvidence(imagPath)
		if imagCapacity.MinimumRatio < realCapacity.MinimumRatio {
			realCapacity.MinimumRatio, realCapacity.WorstCheckpoint = imagCapacity.MinimumRatio, imagCapacity.WorstCheckpoint
		}
		realCapacity.OutsideCount += imagCapacity.OutsideCount
		realCapacity.AllCenteredUnique = realCapacity.AllCenteredUnique && imagCapacity.AllCenteredUnique
		realCapacity.AllRowsMatch = realCapacity.AllRowsMatch && imagCapacity.AllRowsMatch
		candidate.Capacity, candidate.KSchedule, candidate.FinalRestoreKExponent = realCapacity, realSchedule, realFinalK
		candidate.PolynomialOutputScale = realPath.PolyMeta["fast_scale"]
		if realPath.PolyCapacity != nil {
			candidate.PolynomialCapacityRatio, candidate.PolynomialCapacityPass = realPath.PolyCapacity.Ratio, realPath.PolyCapacity.Pass
		}
		candidate.PolynomialVsStandard = realPath.Polynomial["fast_vs_standard"]
		candidate.PolynomialTargetPass = candidate.PolynomialVsStandard != nil && candidate.PolynomialVsStandard.MaxComponent <= precisionSweepPolynomialTarget
		candidate.RealEvalMod, candidate.ImagEvalMod = realFinal.FastVsStandard, imagFinal.FastVsStandard
		candidate.RealNormalizedEffect, candidate.ImagNormalizedEffect = realFinal.FastVsNormalized, imagFinal.FastVsNormalized
		candidate.CoherentRealVsStandard, candidate.CoherentImagVsStandard = realFinal.CoherentVsStandard, imagFinal.CoherentVsStandard
		candidate.PrecisionQualified = realPath.FirstFailure == "none" && imagPath.FirstFailure == "none" && candidate.Capacity.AllCenteredUnique && candidate.Capacity.AllRowsMatch && candidate.RealEvalMod != nil && candidate.ImagEvalMod != nil && candidate.RealEvalMod.MaxComponent <= precisionSweepLocalTarget && candidate.ImagEvalMod.MaxComponent <= precisionSweepLocalTarget
		candidate.Classification, candidate.FirstFailingCheckpoint = precisionSweepPathClassification(realPath, realFinal)
		if imagPath.FirstFailure != "none" && candidate.FirstFailingCheckpoint == "none" {
			candidate.Classification, candidate.FirstFailingCheckpoint = precisionSweepPathClassification(imagPath, imagFinal)
		}
		if exponent == 91 || candidate.PrecisionQualified {
			publicLike, standardLike, postS2C, metadataCorrect, inputUnchanged, publicErr := precisionSweepPublicLike(realPath, imagPath, fastEval, standardEval, btp, residual.MaxSlots(), skN2, reproducibleInput(residual, btp))
			if publicErr != nil {
				return publicErr
			}
			candidate.PublicLike, candidate.StandardPublicLike, candidate.PostS2CVsStandard = publicLike, standardLike, postS2C
			candidate.PublicLikeMetadataCorrect, candidate.PublicLikeInputUnchanged = metadataCorrect, inputUnchanged
			candidate.PublicLikeAccepted = publicLike != nil && publicLike.MaxComponent <= precisionSweepPublicThreshold && metadataCorrect && inputUnchanged
			if exponent == 91 {
				baselinePublicLike = publicLike
			}
			if candidate.PrecisionQualified {
				state := chebyshevOracleState{Params: btp, Eval: fastEval}
				fullReal, _, liftErr := postMod1S2CLiftFull(params, realPath.FastFinal)
				if liftErr != nil {
					return liftErr
				}
				fullImag, _, liftErr := postMod1S2CLiftFull(params, imagPath.FastFinal)
				if liftErr != nil {
					return liftErr
				}
				standardDFT, dftErr := newPostMod1StandardDFT(params, fastEval.S2CDFTMatrix)
				if dftErr != nil {
					return dftErr
				}
				implementationOut, _, _, replayFailure, replayErr := postMod1S2CReplay(state, realPath.FastFinal, imagPath.FastFinal, fullReal, fullImag, standardDFT)
				if replayErr != nil {
					return replayErr
				}
				if replayFailure != nil {
					candidate.PrecisionQualified = false
					candidate.PublicLikeAccepted = false
					candidate.Classification = "normalized_candidate_primitive_mismatch"
					candidate.FirstFailingCheckpoint = fmt.Sprintf("post_s2c.%s", replayFailure.Checkpoint)
				} else {
					implementationValues, decodeErr := decodeWithSecret(params, implementationOut, zeroSecret(params))
					if decodeErr != nil {
						return decodeErr
					}
					standardValues, decodeErr := decodeWithSecret(params, fullReal, skN2)
					if decodeErr != nil {
						return decodeErr
					}
					candidate.PostS2CImplementation = evalModMatchedMetricFor(standardValues, implementationValues, precisionSweepPublicThreshold)
				}
			}
		}
		result.Candidates = append(result.Candidates, candidate)
	}

	if baselineReal.FastFinal != nil {
		baselineFinal, err := evalModMatchedFinalEvidence(baselineReal, params, inputs.FastSK, inputs.StandardSK)
		if err != nil {
			return err
		}
		baselineImagFinal, err := evalModMatchedFinalEvidence(baselineImag, params, inputs.FastSK, inputs.StandardSK)
		if err != nil {
			return err
		}
		result.CurrentEvalModReal, result.CurrentEvalModImag = baselineFinal.FastVsStandard, baselineImagFinal.FastVsStandard
		result.CurrentPolynomialDelta = baselineReal.Polynomial["fast_vs_standard"]
		result.CurrentCoherentRealDelta, result.CurrentCoherentImagDelta = baselineFinal.CoherentVsStandard, baselineImagFinal.CoherentVsStandard
		if result.CurrentEvalModReal != nil && result.CurrentPolynomialDelta != nil && result.CurrentPolynomialDelta.MaxComponent > 0 {
			result.CurrentPolyToEvalModRatio = result.CurrentEvalModReal.MaxComponent / result.CurrentPolynomialDelta.MaxComponent
		}
		result.DerivedPolynomialBudget = precisionSweepLocalTarget
		if result.CurrentPolyToEvalModRatio > 0 {
			result.DerivedPolynomialBudget = precisionSweepLocalTarget / result.CurrentPolyToEvalModRatio
		}
	}
	if baselinePublicLike != nil && result.CurrentEvalModReal != nil && result.CurrentEvalModImag != nil {
		denominator := math.Max(result.CurrentEvalModReal.MaxComponent, result.CurrentEvalModImag.MaxComponent)
		if denominator > 0 {
			result.MeasuredAmplification = baselinePublicLike.MaxComponent / denominator
		}
	}
	result.DerivedLocalBudget = result.PublicThreshold / result.MeasuredAmplification
	preferred := -1
	for _, candidate := range result.Candidates {
		if candidate.PublicLikeAccepted && (preferred < 0 || candidate.Exponent < preferred) {
			preferred = candidate.Exponent
		}
	}
	if preferred >= 0 {
		result.PreferredExponent = preferred
		result.Classification = "logn13_evalmod_common_plan_scale_candidate_validated"
		result.FirstLimitingCheckpoint = "none"
	} else {
		result.Classification = precisionSweepOverallClassification(result.Candidates)
		result.FirstLimitingCheckpoint = precisionSweepFirstLimitingCheckpoint(result.Candidates)
	}
	return precisionSweepWrite(result, outPath)
}

func precisionSweepOverallClassification(candidates []precisionSweepCandidate) string {
	if len(candidates) == 0 {
		return "logn13_evalmod_precision_sweep_precondition_mismatch"
	}
	anyNormalizedFailure, anyPSFailure, anySafe := false, false, false
	for _, candidate := range candidates {
		if candidate.Classification == "normalized_candidate_capacity_failure" {
			anyNormalizedFailure = true
		}
		if candidate.Classification == "polynomial_plan_capacity_failure" {
			anyPSFailure = true
		}
		if candidate.Capacity.AllCenteredUnique && candidate.Capacity.AllRowsMatch && candidate.FirstFailingCheckpoint == "none" {
			anySafe = true
		}
	}
	if !anySafe && anyNormalizedFailure {
		return "logn13_evalmod_common_scale_blocked_by_normalized_da_capacity"
	}
	if !anySafe && anyPSFailure {
		return "logn13_evalmod_common_scale_blocked_by_ps_capacity"
	}
	return "logn13_evalmod_common_scale_insufficient_precision"
}

func precisionSweepFirstLimitingCheckpoint(candidates []precisionSweepCandidate) string {
	for _, candidate := range candidates {
		if candidate.FirstFailingCheckpoint != "none" {
			return candidate.FirstFailingCheckpoint
		}
	}
	return "public_like_threshold"
}
