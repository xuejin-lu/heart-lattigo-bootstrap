package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3SemanticBisectPrimaryBase = "47150c02b7d82e3f766c853e405c2d5cc445d45d"
	requiredFIX001P3SemanticBisectSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
)

type semanticBisectMetric struct {
	MaxComponentAbs float64 `json:"max_component_abs"`
	MaxAbsComplex   float64 `json:"max_abs_complex"`
	MeanAbsComplex  float64 `json:"mean_abs_complex"`
	WorstIndex      int     `json:"worst_index"`
	WorstComponent  string  `json:"worst_component"`
	Threshold       float64 `json:"threshold"`
	Pass            bool    `json:"pass"`
}

type semanticBisectMetadata struct {
	SourceLevel         int             `json:"source_level"`
	DecodeLevel         int             `json:"decode_level"`
	Scale               string          `json:"scale"`
	ScaleLog2           float64         `json:"scale_log2"`
	Degree              int             `json:"degree"`
	N                   int             `json:"n"`
	LogDimensions       ring.Dimensions `json:"log_dimensions"`
	IsNTT               bool            `json:"is_ntt"`
	IsMontgomery        bool            `json:"is_montgomery"`
	ProjectedMontgomery bool            `json:"projected_is_montgomery"`
}

type semanticBisectStage struct {
	Stage           string                       `json:"stage"`
	Fast            semanticBisectMetadata       `json:"fast"`
	Standard        semanticBisectMetadata       `json:"standard"`
	MetadataMatch   bool                         `json:"metadata_match"`
	FastVsStandard  *semanticBisectMetric        `json:"fast_vs_standard,omitempty"`
	FastVsInput     *semanticBisectMetric        `json:"fast_vs_input,omitempty"`
	StandardVsInput *semanticBisectMetric        `json:"standard_vs_input,omitempty"`
	FastPresent     bool                         `json:"fast_present"`
	StandardPresent bool                         `json:"standard_present"`
	FastRows        []FinalizationRowFingerprint `json:"fast_rows_q0_q1,omitempty"`
	StandardRows    []FinalizationRowFingerprint `json:"standard_rows_q0_q1,omitempty"`
}

type semanticBisectStandardProof struct {
	EvaluatorNonNil           bool `json:"evaluator_non_nil"`
	DFTEvaluatorNonNil        bool `json:"dft_evaluator_non_nil"`
	Mod1EvaluatorNonNil       bool `json:"mod1_evaluator_non_nil"`
	FastCompatibilityPathUsed bool `json:"fast_compatibility_path_used"`
	OrdinaryStageCallsUsed    bool `json:"ordinary_stage_calls_used"`
}

type semanticBisectResult struct {
	SchemaVersion                         string                      `json:"schema_version"`
	Timestamp                             time.Time                   `json:"timestamp"`
	Primary                               RepositoryMetadata          `json:"primary_repository"`
	Lattigo                               RepositoryMetadata          `json:"lattigo_repository"`
	Environment                           EnvironmentMetadata         `json:"environment"`
	Config                                BootstrapConfig             `json:"config"`
	Parameters                            ExperimentParameters        `json:"effective_parameters"`
	Workload                              CorrectnessWorkload         `json:"workload"`
	StandardProof                         semanticBisectStandardProof `json:"standard_control_proof"`
	PrecisionMode                         string                      `json:"residual_precision_mode"`
	IterationsParametersPresent           bool                        `json:"iterations_parameters_present"`
	FastScaleDownErrScale                 string                      `json:"fast_scaledown_err_scale"`
	StandardScaleDownErrScale             string                      `json:"standard_scaledown_err_scale"`
	StandardDiffScaleCorrectionApplicable bool                        `json:"standard_diffscale_correction_applicable"`
	StandardEndToEnd                      *semanticBisectMetric       `json:"standard_end_to_end_semantic,omitempty"`
	Stages                                []semanticBisectStage       `json:"stages"`
	FirstFailingStage                     string                      `json:"first_failing_stage"`
	FirstSupportedCause                   string                      `json:"first_supported_cause"`
	Validation                            map[string]interface{}      `json:"validation"`
}

func semanticBisectMetricFor(reference, actual []complex128, threshold float64) (semanticBisectMetric, error) {
	if len(reference) != len(actual) || len(reference) == 0 {
		return semanticBisectMetric{}, fmt.Errorf("semantic vector length mismatch: reference=%d actual=%d", len(reference), len(actual))
	}
	metric := semanticBisectMetric{WorstIndex: -1, Threshold: threshold}
	for i := range reference {
		if math.IsNaN(real(reference[i])) || math.IsNaN(imag(reference[i])) || math.IsInf(real(reference[i]), 0) || math.IsInf(imag(reference[i]), 0) || math.IsNaN(real(actual[i])) || math.IsNaN(imag(actual[i])) || math.IsInf(real(actual[i]), 0) || math.IsInf(imag(actual[i]), 0) {
			return semanticBisectMetric{}, fmt.Errorf("non-finite semantic value at index %d", i)
		}
		delta := actual[i] - reference[i]
		absComplex := cmplx.Abs(delta)
		absReal := math.Abs(real(delta))
		absImag := math.Abs(imag(delta))
		metric.MeanAbsComplex += absComplex
		if absComplex > metric.MaxAbsComplex {
			metric.MaxAbsComplex = absComplex
		}
		if absReal > metric.MaxComponentAbs {
			metric.MaxComponentAbs = absReal
			metric.WorstIndex = i
			metric.WorstComponent = "real"
		}
		if absImag > metric.MaxComponentAbs {
			metric.MaxComponentAbs = absImag
			metric.WorstIndex = i
			metric.WorstComponent = "imag"
		}
	}
	metric.MeanAbsComplex /= float64(len(reference))
	metric.Pass = metric.MaxComponentAbs <= threshold
	return metric, nil
}

func semanticBisectProject(source *rlwe.Ciphertext, params ckks.Parameters) (*rlwe.Ciphertext, int, error) {
	if source == nil || source.MetaData == nil {
		return nil, 0, fmt.Errorf("cannot project nil ciphertext")
	}
	targetLevel := source.Level()
	if targetLevel > 1 {
		targetLevel = 1
	}
	if targetLevel < 0 || targetLevel > params.MaxLevel() {
		return nil, 0, fmt.Errorf("projection level %d is outside target parameters", targetLevel)
	}
	projected := ckks.NewCiphertext(params, source.Degree(), targetLevel)
	*projected.MetaData = *source.MetaData
	for component := 0; component <= source.Degree(); component++ {
		if component >= len(source.Value) || component >= len(projected.Value) {
			return nil, 0, fmt.Errorf("projection component %d is unavailable", component)
		}
		for limb := 0; limb <= targetLevel && limb < 2; limb++ {
			if limb >= len(source.Value[component].Coeffs) || limb >= len(projected.Value[component].Coeffs) {
				return nil, 0, fmt.Errorf("projection maintained limb %d is unavailable", limb)
			}
			copy(projected.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
			if source.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(projected.Value[component].Coeffs[limb], projected.Value[component].Coeffs[limb])
			}
		}
	}
	projected.IsNTT = source.IsNTT
	projected.IsMontgomery = false
	return projected, targetLevel, nil
}

func semanticBisectView(source *rlwe.Ciphertext, params ckks.Parameters, sk *rlwe.SecretKey) (semanticBisectMetadata, []complex128, error) {
	projected, decodeLevel, err := semanticBisectProject(source, params)
	if err != nil {
		return semanticBisectMetadata{}, nil, err
	}
	decoded, err := decodeWithSecret(params, projected, sk)
	if err != nil {
		return semanticBisectMetadata{}, nil, err
	}
	return semanticBisectMetadata{
		SourceLevel: source.Level(), DecodeLevel: decodeLevel, Scale: finalizationScaleString(source.Scale), ScaleLog2: source.Scale.Log2(), Degree: source.Degree(), N: source.N(), LogDimensions: source.LogDimensions, IsNTT: source.IsNTT, IsMontgomery: source.IsMontgomery, ProjectedMontgomery: projected.IsMontgomery,
	}, decoded, nil
}

func semanticBisectMetadataMatch(fast, standard semanticBisectMetadata) bool {
	return fast.SourceLevel == standard.SourceLevel && fast.DecodeLevel == standard.DecodeLevel && fast.Scale == standard.Scale && fast.Degree == standard.Degree && fast.N == standard.N && fast.LogDimensions == standard.LogDimensions && fast.IsNTT == standard.IsNTT
}

func semanticBisectStageEvidence(name string, fast, standard *rlwe.Ciphertext, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, input []complex128, compareInput bool) (semanticBisectStage, bool, error) {
	stage := semanticBisectStage{Stage: name, FastPresent: fast != nil, StandardPresent: standard != nil}
	if fast == nil || standard == nil {
		return stage, fast == nil && standard == nil, nil
	}
	fastMetadata, fastDecoded, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return stage, false, err
	}
	standardMetadata, standardDecoded, err := semanticBisectView(standard, params, standardSK)
	if err != nil {
		return stage, false, err
	}
	stage.Fast = fastMetadata
	stage.Standard = standardMetadata
	stage.MetadataMatch = semanticBisectMetadataMatch(fastMetadata, standardMetadata)
	stage.FastRows = finalizationEvidence(fast).Rows
	stage.StandardRows = finalizationEvidence(standard).Rows
	fastVsStandard, err := semanticBisectMetricFor(standardDecoded, fastDecoded, correctnessThreshold)
	if err != nil {
		return stage, false, err
	}
	stage.FastVsStandard = &fastVsStandard
	if compareInput {
		fastVsInput, err := semanticBisectMetricFor(input, fastDecoded, correctnessThreshold)
		if err != nil {
			return stage, false, err
		}
		standardVsInput, err := semanticBisectMetricFor(input, standardDecoded, correctnessThreshold)
		if err != nil {
			return stage, false, err
		}
		stage.FastVsInput = &fastVsInput
		stage.StandardVsInput = &standardVsInput
	}
	return stage, stage.MetadataMatch && fastVsStandard.Pass, nil
}

func semanticBisectScaleString(scale *rlwe.Scale) string {
	if scale == nil {
		return "nil"
	}
	return finalizationScaleString(*scale)
}

func buildFIX001P3GenuineStandardEvaluator(btp bootstrapping.Parameters, skN1 *rlwe.SecretKey) (*bootstrapping.Evaluator, *rlwe.SecretKey, error) {
	paramsN2 := btp.BootstrappingParameters
	if btp.ResidualParameters.N() != paramsN2.N() {
		return nil, nil, fmt.Errorf("semantic bisect Standard constructor requires N1 == N2")
	}
	skN2 := rlwe.NewSecretKey(paramsN2)
	buffer := paramsN2.RingQ().NewPoly()
	rlwe.ExtendBasisSmallNormAndCenterNTTMontgomery(paramsN2.RingQ(), paramsN2.RingQ(), skN1.Value.Q, buffer, skN2.Value.Q)
	rlwe.ExtendBasisSmallNormAndCenterNTTMontgomery(paramsN2.RingQ(), paramsN2.RingP(), skN1.Value.Q, buffer, skN2.Value.P)
	kgen := rlwe.NewKeyGenerator(paramsN2)
	rlk := kgen.GenRelinearizationKeyNew(skN2)
	galEls := append(btp.GaloisElements(paramsN2), paramsN2.GaloisElementForComplexConjugation())
	gks := kgen.GenGaloisKeysNew(galEls, skN2)
	keys := &bootstrapping.EvaluationKeys{MemEvaluationKeySet: rlwe.NewMemEvaluationKeySet(rlk, gks...)}
	eval, err := bootstrapping.NewEvaluator(btp, keys)
	if err != nil {
		return nil, nil, err
	}
	return eval, skN2, nil
}

func semanticBisectWrite(result semanticBisectResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DiagLogN13SemanticBisect(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := semanticBisectResult{
		SchemaVersion: "fix-001-p3-diag-logn13-semantic-bisect.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		StandardProof: semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: standardEval.Evaluator == nil && standardEval.DFTEvaluator == nil && standardEval.Mod1Evaluator == nil, OrdinaryStageCallsUsed: true},
		PrecisionMode: "PREC64", IterationsParametersPresent: btp.IterationsParameters != nil, StandardDiffScaleCorrectionApplicable: false, FirstFailingStage: "none", FirstSupportedCause: "logn13_semantic_bisect_standard_control_construction_failure",
		Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3SemanticBisectPrimaryBase, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredFIX001P3SemanticBisectSecondary, "no_secondary_production_changes": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
	}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil {
		return semanticBisectWrite(result, outPath)
	}

	standardControlInput := reproducibleInput(residual, btp)
	standardControlOutput, err := standardEval.Bootstrap(standardControlInput)
	if err != nil {
		result.FirstSupportedCause = "logn13_semantic_bisect_standard_control_failure"
		return semanticBisectWrite(result, outPath)
	}
	standardControlDecoded, err := decodeWithSecret(residual, standardControlOutput, skN1)
	if err != nil {
		result.FirstSupportedCause = "logn13_semantic_bisect_standard_control_failure"
		return semanticBisectWrite(result, outPath)
	}
	standardMetric, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), standardControlDecoded, correctnessThreshold)
	if err != nil {
		return err
	}
	result.StandardEndToEnd = &standardMetric
	if !standardMetric.Pass {
		result.FirstSupportedCause = "logn13_semantic_bisect_standard_control_failure"
		return semanticBisectWrite(result, outPath)
	}

	fastSKResidual := zeroSecret(residual)
	fastSKBootstrap := zeroSecret(btp.BootstrappingParameters)
	inputValues := reproducibleValues(residual.MaxSlots())
	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	inputStage, inputPass, err := semanticBisectStageEvidence("input", fastInput, standardInput, residual, fastSKResidual, skN1, inputValues, true)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, inputStage)
	if !inputPass {
		result.FirstFailingStage = "input"
		result.FirstSupportedCause = "logn13_semantic_bisect_input_precondition_failure"
		return semanticBisectWrite(result, outPath)
	}

	fastPacked, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput})
	if err != nil {
		return err
	}
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput})
	if err != nil {
		return err
	}
	packStage, packPass, err := semanticBisectStageEvidence("pack_and_switch_n1_to_n2", &fastPacked[0], &standardPacked[0], btp.BootstrappingParameters, fastSKBootstrap, skN2, inputValues, true)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, packStage)
	if !packPass {
		result.FirstFailingStage = "pack_and_switch_n1_to_n2"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_modup"
		return semanticBisectWrite(result, outPath)
	}

	fastScaled, fastErrScale, err := fastEval.ScaleDown(&fastPacked[0])
	if err != nil {
		return err
	}
	standardScaled, standardErrScale, err := standardEval.ScaleDown(&standardPacked[0])
	if err != nil {
		return err
	}
	result.FastScaleDownErrScale = semanticBisectScaleString(fastErrScale)
	result.StandardScaleDownErrScale = semanticBisectScaleString(standardErrScale)
	scaledStage, scaledPass, err := semanticBisectStageEvidence("scale_down", fastScaled, standardScaled, btp.BootstrappingParameters, fastSKBootstrap, skN2, inputValues, true)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, scaledStage)
	if !scaledPass {
		result.FirstFailingStage = "scale_down"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_scaledown"
		return semanticBisectWrite(result, outPath)
	}

	fastModUp, err := fastEval.ModUp(fastScaled)
	if err != nil {
		return err
	}
	standardModUp, err := standardEval.ModUp(standardScaled)
	if err != nil {
		return err
	}
	modUpStage, modUpPass, err := semanticBisectStageEvidence("mod_up", fastModUp, standardModUp, btp.BootstrappingParameters, fastSKBootstrap, skN2, inputValues, true)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, modUpStage)
	if !modUpPass {
		result.FirstFailingStage = "mod_up"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_modup"
		return semanticBisectWrite(result, outPath)
	}

	fastReal, fastImag, err := fastEval.CoeffsToSlots(fastModUp)
	if err != nil {
		return err
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp)
	if err != nil {
		return err
	}
	if (fastReal == nil) != (standardReal == nil) || (fastImag == nil) != (standardImag == nil) {
		result.FirstFailingStage = "coeffs_to_slots_branch_structure"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_c2s_real"
		return semanticBisectWrite(result, outPath)
	}
	c2sRealStage, c2sRealPass, err := semanticBisectStageEvidence("coeffs_to_slots_real", fastReal, standardReal, btp.BootstrappingParameters, fastSKBootstrap, skN2, nil, false)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, c2sRealStage)
	if !c2sRealPass {
		result.FirstFailingStage = "coeffs_to_slots_real"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_c2s_real"
		return semanticBisectWrite(result, outPath)
	}
	if fastImag != nil {
		c2sImagStage, c2sImagPass, err := semanticBisectStageEvidence("coeffs_to_slots_imag", fastImag, standardImag, btp.BootstrappingParameters, fastSKBootstrap, skN2, nil, false)
		if err != nil {
			return err
		}
		result.Stages = append(result.Stages, c2sImagStage)
		if !c2sImagPass {
			result.FirstFailingStage = "coeffs_to_slots_imag"
			result.FirstSupportedCause = "logn13_semantic_first_divergence_c2s_imag"
			return semanticBisectWrite(result, outPath)
		}
	}

	fastReal, err = fastEval.EvalMod(fastReal)
	if err != nil {
		return err
	}
	standardReal, err = standardEval.EvalMod(standardReal)
	if err != nil {
		return err
	}
	evalModRealStage, evalModRealPass, err := semanticBisectStageEvidence("eval_mod_real", fastReal, standardReal, btp.BootstrappingParameters, fastSKBootstrap, skN2, nil, false)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, evalModRealStage)
	if !evalModRealPass {
		result.FirstFailingStage = "eval_mod_real"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_evalmod_real"
		return semanticBisectWrite(result, outPath)
	}
	if fastImag != nil {
		fastImag, err = fastEval.EvalMod(fastImag)
		if err != nil {
			return err
		}
		standardImag, err = standardEval.EvalMod(standardImag)
		if err != nil {
			return err
		}
		evalModImagStage, evalModImagPass, err := semanticBisectStageEvidence("eval_mod_imag", fastImag, standardImag, btp.BootstrappingParameters, fastSKBootstrap, skN2, nil, false)
		if err != nil {
			return err
		}
		result.Stages = append(result.Stages, evalModImagStage)
		if !evalModImagPass {
			result.FirstFailingStage = "eval_mod_imag"
			result.FirstSupportedCause = "logn13_semantic_first_divergence_evalmod_imag"
			return semanticBisectWrite(result, outPath)
		}
	}

	fastCore, err := fastEval.SlotsToCoeffs(fastReal, fastImag)
	if err != nil {
		return err
	}
	standardCore, err := standardEval.SlotsToCoeffs(standardReal, standardImag)
	if err != nil {
		return err
	}
	coreStage, corePass, err := semanticBisectStageEvidence("slots_to_coeffs_core_output", fastCore, standardCore, btp.BootstrappingParameters, fastSKBootstrap, skN2, inputValues, true)
	if err != nil {
		return err
	}
	result.Stages = append(result.Stages, coreStage)
	if !corePass {
		result.FirstFailingStage = "slots_to_coeffs_core_output"
		result.FirstSupportedCause = "logn13_semantic_first_divergence_s2c_core"
		return semanticBisectWrite(result, outPath)
	}

	result.FirstSupportedCause = "logn13_semantic_bisect_no_internal_divergence_oracle_mismatch"
	return semanticBisectWrite(result, outPath)
}
