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

func fix001P3B4T2GuardDecodePublic(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	projection, _, err := q01Projection(params, ct, false)
	if err != nil {
		return nil, err
	}
	return diagnosticDecode(params, projection, zeroSecret(params))
}

const (
	fix001P3B4T2GuardPublicThreshold       = 1e-2
	fix001P3B4T2GuardFeasibilityPublicLike = 0.005554220603853743
)

type fix001P3B4T2GuardIntegrationResult struct {
	SchemaVersion         string                 `json:"schema_version"`
	Timestamp             time.Time              `json:"timestamp"`
	Primary               RepositoryMetadata     `json:"primary_repository"`
	Lattigo               RepositoryMetadata     `json:"lattigo_repository"`
	Environment           EnvironmentMetadata    `json:"environment"`
	Config                BootstrapConfig        `json:"config"`
	Parameters            ExperimentParameters   `json:"effective_parameters"`
	Workload              CorrectnessWorkload    `json:"workload"`
	Classification        string                 `json:"classification"`
	FirstFailingCheck     string                 `json:"first_failing_checkpoint"`
	FinalMetadata         map[string]interface{} `json:"production_final_metadata"`
	EvalModReal           *PSGlobalMetric        `json:"evalmod_real_vs_standard"`
	EvalModImag           *PSGlobalMetric        `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2C               *PSGlobalMetric        `json:"post_s2c_vs_standard"`
	PublicLike            *PSGlobalMetric        `json:"public_like"`
	StandardLike          *PSGlobalMetric        `json:"standard_like"`
	FinalVsStandard       *PSGlobalMetric        `json:"final_vs_standard,omitempty"`
	PublicThreshold       float64                `json:"public_like_threshold"`
	PublicMargin          float64                `json:"public_like_margin"`
	FeasibilityReference  float64                `json:"feasibility_guarded_public_like_reference"`
	FeasibilityDifference float64                `json:"production_vs_feasibility_difference"`
	GuardSelection        map[string]interface{} `json:"guard_structural_selection"`
	ContractionMetadata   map[string]interface{} `json:"guard_contraction_metadata"`
	Validation            map[string]interface{} `json:"validation"`
}

func fix001P3B4T2GuardMetricFinite(metric *PSGlobalMetric) bool {
	if metric == nil {
		return false
	}
	for _, value := range []float64{metric.Threshold, metric.MaxComplex, metric.MaxComponent, metric.MeanComplex} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func fix001P3B4T2GuardWrite(result fix001P3B4T2GuardIntegrationResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for FIX-001-P3 production integration")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3IntegrateLogN13B4T2OneBitScalarGuard(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	result := fix001P3B4T2GuardIntegrationResult{
		SchemaVersion: "fix-001-p3-integrate-logn13-b4-t2-one-bit-scalar-guard.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, Parameters: parameterMetadata(residual, btp),
		Workload:       CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		Classification: "logn13_b4_t2_guard_production_selection_mismatch", FirstFailingCheck: "startup",
		PublicThreshold: fix001P3B4T2GuardPublicThreshold, FeasibilityReference: fix001P3B4T2GuardFeasibilityPublicLike,
		Validation: map[string]interface{}{
			"ordinary_fast_bootstrap_path": true, "logn13_only": cfg.LogN == 13,
			"secondary_commit":        gitOutput(backendRoot, "rev-parse", "HEAD"),
			"secondary_clean":         gitOutput(backendRoot, "status", "--porcelain") == "",
			"no_diagnostic_injection": true, "no_logn16": true, "no_benchmark": true,
			"no_gate_4_or_5": true, "no_exp003": true, "no_later_bootstrap_stage": true,
		},
	}
	if cfg.LogN != 13 {
		return fix001P3B4T2GuardWrite(result, outPath)
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
	zeroN2 := zeroSecret(btp.BootstrappingParameters)

	// Exercise the same ordinary production stage calls used by Fast Bootstrap
	// to retain compact EvalMod and post-S2C evidence.
	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastPacked, fastN1, fastN2, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput.CopyNew()})
	if err != nil {
		return err
	}
	standardPacked, standardN1, standardN2, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput.CopyNew()})
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
	fastReal, fastImag, err := fastEval.CoeffsToSlots(fastModUp)
	if err != nil {
		return fmt.Errorf("ordinary Fast CoeffsToSlots: %w (level=%d is_ntt=%t is_montgomery=%t)", err, fastModUp.Level(), fastModUp.IsNTT, fastModUp.IsMontgomery)
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp)
	if err != nil {
		return fmt.Errorf("ordinary Standard CoeffsToSlots: %w (level=%d is_ntt=%t is_montgomery=%t)", err, standardModUp.Level(), standardModUp.IsNTT, standardModUp.IsMontgomery)
	}
	fastReal, err = fastEval.EvalMod(fastReal)
	if err != nil {
		return err
	}
	standardReal, err = standardEval.EvalMod(standardReal)
	if err != nil {
		return err
	}
	var fastImagValues, standardImagValues []complex128
	if fastImag != nil && standardImag != nil {
		fastImag, err = fastEval.EvalMod(fastImag)
		if err != nil {
			return err
		}
		standardImag, err = standardEval.EvalMod(standardImag)
		if err != nil {
			return err
		}
		_, fastImagValues, err = semanticBisectView(fastImag, btp.BootstrappingParameters, zeroN2)
		if err != nil {
			return err
		}
		_, standardImagValues, err = semanticBisectView(standardImag, btp.BootstrappingParameters, skN2)
		if err != nil {
			return err
		}
		result.EvalModImag = psGlobalMetric(standardImagValues, fastImagValues)
	}
	_, fastRealValues, err := semanticBisectView(fastReal, btp.BootstrappingParameters, zeroN2)
	if err != nil {
		return err
	}
	_, standardRealValues, err := semanticBisectView(standardReal, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}
	result.EvalModReal = psGlobalMetric(standardRealValues, fastRealValues)

	fastCore, err := fastEval.SlotsToCoeffs(fastReal, fastImag)
	if err != nil {
		return err
	}
	standardCore, err := standardEval.SlotsToCoeffs(standardReal, standardImag)
	if err != nil {
		return err
	}
	_, fastCoreValues, err := semanticBisectView(fastCore, btp.BootstrappingParameters, zeroN2)
	if err != nil {
		return err
	}
	_, standardCoreValues, err := semanticBisectView(standardCore, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}
	result.PostS2C = psGlobalMetric(standardCoreValues, fastCoreValues)

	// The final result is obtained from the ordinary public Bootstrap call,
	// rather than from an oracle or a diagnostic replay.
	fastPublicInput := reproducibleInput(residual, btp)
	fastPublicBefore := fastPublicInput.CopyNew()
	fastOutput, err := fastEval.Bootstrap(fastPublicInput)
	if err != nil {
		return err
	}
	standardOutput, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return err
	}
	fastFinalValues, err := fix001P3B4T2GuardDecodePublic(residual, fastOutput)
	if err != nil {
		return err
	}
	standardFinalValues, err := decodeWithSecret(residual, standardOutput, skN1)
	if err != nil {
		return err
	}
	message := reproducibleValues(residual.MaxSlots())
	result.PublicLike = psGlobalMetric(message, fastFinalValues)
	result.StandardLike = psGlobalMetric(message, standardFinalValues)
	result.FinalVsStandard = psGlobalMetric(standardFinalValues, fastFinalValues)
	if result.PublicLike != nil {
		result.PublicMargin = result.PublicThreshold - result.PublicLike.MaxComponent
		result.FeasibilityDifference = math.Abs(result.PublicLike.MaxComponent - result.FeasibilityReference)
	}
	selection := fastEval.PolynomialEvaluator.LastGuardSelectionEvidence()
	selectionPass := selection.PlanScaleOverride && selection.PlanCount > 0 && selection.PlanIndex == selection.PlanCount-1 && selection.ScalarDegree == 2 && selection.Operations == 1
	result.GuardSelection = map[string]interface{}{"plan_scale_override": selection.PlanScaleOverride, "plan_index": selection.PlanIndex, "plan_count": selection.PlanCount, "final_parent_selected": selection.PlanCount > 0 && selection.PlanIndex == selection.PlanCount-1, "scalar_degree": selection.ScalarDegree, "selected_operations": selection.Operations, "pass": selectionPass}
	metadataPass := fastOutput.Level() == residual.MaxLevel() && fastOutput.Degree() == 1 && fastOutput.IsNTT && !fastOutput.IsMontgomery && fastOutput.Scale.Equal(residual.DefaultScale())
	result.ContractionMetadata = map[string]interface{}{"contraction_completed": selection.Contraction, "final_level": fastOutput.Level(), "final_degree": fastOutput.Degree(), "final_is_ntt": fastOutput.IsNTT, "final_is_montgomery": fastOutput.IsMontgomery, "final_scale": finalizationScaleString(fastOutput.Scale), "pass": selection.Contraction && metadataPass}
	result.FinalMetadata = result.ContractionMetadata
	result.Validation["fast_input_unchanged"] = fix001P3E2EInputUnchanged(fastPublicBefore, fastPublicInput)
	result.Validation["compact_artifact"] = true
	result.Validation["no_nan_or_inf"] = fix001P3B4T2GuardMetricFinite(result.EvalModReal) && fix001P3B4T2GuardMetricFinite(result.EvalModImag) && fix001P3B4T2GuardMetricFinite(result.PostS2C) && fix001P3B4T2GuardMetricFinite(result.PublicLike)

	result.Classification = "logn13_b4_t2_one_bit_scalar_guard_production_integrated"
	result.FirstFailingCheck = "none"
	if !selectionPass {
		result.Classification, result.FirstFailingCheck = "logn13_b4_t2_guard_production_selection_mismatch", "guard_selection"
	} else if !result.ContractionMetadata["pass"].(bool) {
		result.Classification, result.FirstFailingCheck = "logn13_b4_t2_guard_production_contraction_mismatch", "guard_contraction_metadata"
	} else if result.EvalModReal == nil || !result.EvalModReal.Pass || (result.EvalModImag != nil && !result.EvalModImag.Pass) || result.PostS2C == nil || !result.PostS2C.Pass || result.FeasibilityDifference > 1e-10 || !result.Validation["no_nan_or_inf"].(bool) {
		result.Classification, result.FirstFailingCheck = "logn13_b4_t2_guard_production_numeric_mismatch", "evalmod_or_post_s2c_or_feasibility"
	} else if result.PublicLike == nil || !result.PublicLike.Pass {
		result.Classification, result.FirstFailingCheck = "logn13_b4_t2_guard_production_public_threshold_failure", "public_like"
	}
	result.Validation["classification_pass"] = result.Classification == "logn13_b4_t2_one_bit_scalar_guard_production_integrated"
	result.Validation["ordinary_fast_public_bootstrap_metadata"] = metadataPass
	result.Validation["input_unchanged"] = result.Validation["fast_input_unchanged"]
	_ = fastN1
	_ = fastN2
	_ = standardN1
	_ = standardN2
	return fix001P3B4T2GuardWrite(result, outPath)
}
