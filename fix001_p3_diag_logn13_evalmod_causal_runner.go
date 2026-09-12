package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredEvalModCausalPrimaryBase = "da9db6c8efa9fec30243241bbb58c0da3afe18f5"
	requiredEvalModCausalSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
	evalModMatchedInputThreshold     = 1e-6
	predecessorC2SRealMaxComponent   = 0.0003098126393859551
)

type evalModMatchedInputEvidence struct {
	Name                 string                 `json:"name"`
	Source               semanticBisectMetadata `json:"source"`
	Reconstructed        semanticBisectMetadata `json:"reconstructed"`
	ReconstructionMetric semanticBisectMetric   `json:"reconstruction_metric"`
	Bound                float64                `json:"bound"`
	Pass                 bool                   `json:"pass"`
}

type evalModCausalOutput struct {
	Name                        string                       `json:"name"`
	Metadata                    semanticBisectMetadata       `json:"metadata"`
	Rows                        []FinalizationRowFingerprint `json:"rows_q0_q1"`
	AcceptedNormalizedRowsMatch *bool                        `json:"accepted_normalized_rows_match,omitempty"`
}

type evalModCausalComparisons struct {
	FastNativeVsStdNative semanticBisectMetric `json:"fast_native_vs_std_native"`
	FastNativeVsStdOnFast semanticBisectMetric `json:"fast_native_vs_std_on_fast"`
	StdNativeVsFastOnStd  semanticBisectMetric `json:"std_native_vs_fast_on_std"`
	StdNativeVsStdOnFast  semanticBisectMetric `json:"std_native_vs_std_on_fast"`
	FastNativeVsFastOnStd semanticBisectMetric `json:"fast_native_vs_fast_on_std"`
}

type evalModInternalCheckpoint struct {
	Checkpoint         string                 `json:"checkpoint"`
	Round              int                    `json:"round"`
	KInExponent        int                    `json:"k_in_exponent,omitempty"`
	KOutExponent       int                    `json:"k_out_exponent,omitempty"`
	MultiplierExponent int                    `json:"multiplier_exponent,omitempty"`
	NormalizedConstant float64                `json:"normalized_constant,omitempty"`
	Fast               semanticBisectMetadata `json:"fast"`
	Standard           semanticBisectMetadata `json:"standard"`
	Metric             semanticBisectMetric   `json:"fast_vs_standard"`
	FastCapacity       *postMod1S2CCapacity   `json:"fast_q01_capacity,omitempty"`
	StandardCapacity   *postMod1S2CCapacity   `json:"standard_full_rns_capacity,omitempty"`
}

type evalModCausalResult struct {
	SchemaVersion                  string                       `json:"schema_version"`
	Timestamp                      time.Time                    `json:"timestamp"`
	Primary                        RepositoryMetadata           `json:"primary_repository"`
	Lattigo                        RepositoryMetadata           `json:"lattigo_repository"`
	Environment                    EnvironmentMetadata          `json:"environment"`
	Config                         BootstrapConfig              `json:"config"`
	Parameters                     ExperimentParameters         `json:"effective_parameters"`
	Workload                       CorrectnessWorkload          `json:"workload"`
	StandardProof                  semanticBisectStandardProof  `json:"standard_control_proof"`
	StandardEndToEnd               *semanticBisectMetric        `json:"standard_end_to_end_semantic,omitempty"`
	C2SReal                        *semanticBisectStage         `json:"c2s_real_predecessor_reproduction,omitempty"`
	NativeEvalMod                  *semanticBisectMetric        `json:"native_evalmod_predecessor_reproduction,omitempty"`
	StandardFromFast               *evalModMatchedInputEvidence `json:"standard_from_fast_semantic_input,omitempty"`
	FastFromStandard               *evalModMatchedInputEvidence `json:"fast_from_standard_semantic_input,omitempty"`
	Outputs                        []evalModCausalOutput        `json:"evalmod_outputs,omitempty"`
	Comparisons                    *evalModCausalComparisons    `json:"comparisons,omitempty"`
	AmplificationRatio             float64                      `json:"standard_c2s_input_amplification_ratio,omitempty"`
	ProvisionalMatchedInputVerdict string                       `json:"provisional_matched_input_verdict"`
	MatchedInputAsymmetry          bool                         `json:"matched_input_asymmetry"`
	NativeInputsUnchanged          bool                         `json:"native_c2s_inputs_unchanged"`
	InternalCheckpoints            []evalModInternalCheckpoint  `json:"internal_checkpoints,omitempty"`
	DerivedInitialKExponent        int                          `json:"derived_initial_k_exponent,omitempty"`
	DerivedFinalKExponent          int                          `json:"derived_final_k_exponent,omitempty"`
	FirstFailingInternalCheckpoint string                       `json:"first_failing_internal_checkpoint"`
	FirstSupportedCause            string                       `json:"first_supported_cause"`
	Validation                     map[string]interface{}       `json:"validation"`
}

func evalModCausalWrite(result evalModCausalResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func evalModEncodePlaintext(params ckks.Parameters, values []complex128, source *rlwe.Ciphertext) (*rlwe.Plaintext, error) {
	if source == nil || source.MetaData == nil {
		return nil, fmt.Errorf("matched-input source cannot be nil")
	}
	plaintext := ckks.NewPlaintext(params, source.Level())
	*plaintext.MetaData = *source.MetaData
	plaintext.IsNTT = true
	plaintext.IsMontgomery = false
	if err := ckks.NewEncoder(params).Encode(values, plaintext); err != nil {
		return nil, err
	}
	return plaintext, nil
}

func evalModStandardFromFast(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128, sk *rlwe.SecretKey) (*rlwe.Ciphertext, error) {
	plaintext, err := evalModEncodePlaintext(params, values, source)
	if err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(plaintext)
}

func evalModFastFromStandard(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128) (*rlwe.Ciphertext, error) {
	plaintext, err := evalModEncodePlaintext(params, values, source)
	if err != nil {
		return nil, err
	}
	ciphertext := fastckks.NewCiphertext(params, 1, source.Level())
	*ciphertext.MetaData = *plaintext.MetaData
	ciphertext.IsNTT = true
	ciphertext.IsMontgomery = true
	for limb := 0; limb < 2; limb++ {
		copy(ciphertext.Value[0].Coeffs[limb], plaintext.Value.Coeffs[limb])
		params.RingQ().SubRings[limb].MForm(ciphertext.Value[0].Coeffs[limb], ciphertext.Value[0].Coeffs[limb])
		ciphertext.Value[1].Coeffs[limb] = make([]uint64, len(ciphertext.Value[1].Coeffs[limb]))
	}
	return ciphertext, nil
}

func evalModMatchedInput(name string, source, reconstructed *rlwe.Ciphertext, params ckks.Parameters, sourceSK, reconstructedSK *rlwe.SecretKey) (evalModMatchedInputEvidence, error) {
	sourceMetadata, sourceValues, err := semanticBisectView(source, params, sourceSK)
	if err != nil {
		return evalModMatchedInputEvidence{}, err
	}
	reconstructedMetadata, reconstructedValues, err := semanticBisectView(reconstructed, params, reconstructedSK)
	if err != nil {
		return evalModMatchedInputEvidence{}, err
	}
	metric, err := semanticBisectMetricFor(sourceValues, reconstructedValues, evalModMatchedInputThreshold)
	if err != nil {
		return evalModMatchedInputEvidence{}, err
	}
	return evalModMatchedInputEvidence{Name: name, Source: sourceMetadata, Reconstructed: reconstructedMetadata, ReconstructionMetric: metric, Bound: evalModMatchedInputThreshold, Pass: metric.Pass}, nil
}

func evalModCausalOutputEvidence(name string, ciphertext *rlwe.Ciphertext, params ckks.Parameters, sk *rlwe.SecretKey, accepted *bool) (evalModCausalOutput, []complex128, error) {
	metadata, decoded, err := semanticBisectView(ciphertext, params, sk)
	if err != nil {
		return evalModCausalOutput{}, nil, err
	}
	return evalModCausalOutput{Name: name, Metadata: metadata, Rows: finalizationEvidence(ciphertext).Rows, AcceptedNormalizedRowsMatch: accepted}, decoded, nil
}

func evalModCausalMetric(reference, actual []complex128) (semanticBisectMetric, error) {
	return semanticBisectMetricFor(reference, actual, correctnessThreshold)
}

func evalModCausalClassification(comparisons evalModCausalComparisons) (classification, verdict string, asymmetry bool) {
	asymmetry = comparisons.FastNativeVsStdOnFast.Pass != comparisons.StdNativeVsFastOnStd.Pass
	if comparisons.FastNativeVsStdOnFast.Pass && comparisons.StdNativeVsFastOnStd.Pass && !comparisons.FastNativeVsStdNative.Pass && !comparisons.StdNativeVsStdOnFast.Pass {
		return "logn13_evalmod_divergence_explained_by_c2s_input_sensitivity", "c2s_input_sensitivity", asymmetry
	}
	return "", "fast_evalmod_matched_input_divergence", asymmetry
}

func evalModCausalTargetScale(input *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator) (rlwe.Scale, error) {
	params := eval.Parameters.BootstrappingParameters
	mod1Params := eval.Mod1Parameters
	target := mod1Params.ScalingFactor()
	for i := 0; i < mod1Params.DoubleAngle; i++ {
		index := input.Level() - mod1Params.Mod1Poly.Depth() - mod1Params.DoubleAngle + i + 1
		if index < 0 || index >= len(params.Q()) {
			return rlwe.Scale{}, fmt.Errorf("Mod1 target-scale Q index %d is outside the parameter chain", index)
		}
		target = target.Mul(rlwe.NewScale(params.Q()[index]))
		target.Value.Sqrt(&target.Value)
	}
	return target, nil
}

func evalModCausalOffset(eval *bootstrapping.FastEvaluator) *big.Float {
	mod1Params := eval.Mod1Parameters
	offset := new(big.Float).Sub(&mod1Params.Mod1Poly.B, &mod1Params.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(mod1Params.IntervalShrinkFactor()))
	return offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
}

func evalModInternalEvidence(checkpoint string, fast, standard *rlwe.Ciphertext, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, withCapacity bool) (evalModInternalCheckpoint, error) {
	return evalModInternalScaledEvidence(checkpoint, fast, standard, params, fastSK, standardSK, 1, withCapacity)
}

func evalModInternalScaledEvidence(checkpoint string, fast, standard *rlwe.Ciphertext, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, fastSemanticMultiplier float64, withCapacity bool) (evalModInternalCheckpoint, error) {
	fastMetadata, fastValues, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return evalModInternalCheckpoint{}, err
	}
	standardMetadata, standardValues, err := semanticBisectView(standard, params, standardSK)
	if err != nil {
		return evalModInternalCheckpoint{}, err
	}
	for i := range fastValues {
		fastValues[i] *= complex(fastSemanticMultiplier, 0)
	}
	metric, err := evalModCausalMetric(standardValues, fastValues)
	if err != nil {
		return evalModInternalCheckpoint{}, err
	}
	evidence := evalModInternalCheckpoint{Checkpoint: checkpoint, Round: -1, Fast: fastMetadata, Standard: standardMetadata, Metric: metric}
	if withCapacity {
		fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
		if err != nil {
			return evalModInternalCheckpoint{}, err
		}
		standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standard)
		if err != nil {
			return evalModInternalCheckpoint{}, err
		}
		evidence.FastCapacity = &fastCapacity
		evidence.StandardCapacity = &standardCapacity
	}
	return evidence, nil
}

func runFIX001P3DiagLogN13EvalModCausal(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := evalModCausalResult{
		SchemaVersion: "fix-001-p3-diag-logn13-evalmod-causal-localize.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		StandardProof:                  semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: standardEval.Evaluator == nil && standardEval.DFTEvaluator == nil && standardEval.Mod1Evaluator == nil, OrdinaryStageCallsUsed: true},
		FirstFailingInternalCheckpoint: "none", FirstSupportedCause: "logn13_evalmod_causal_precondition_mismatch",
		Validation: map[string]interface{}{"primary_required_base": requiredEvalModCausalPrimaryBase, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredEvalModCausalSecondary, "no_secondary_changes": true, "no_production_fix": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_unpack_or_finalization": true},
	}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil {
		return evalModCausalWrite(result, outPath)
	}

	standardControlOutput, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return evalModCausalWrite(result, outPath)
	}
	standardControlDecoded, err := decodeWithSecret(residual, standardControlOutput, skN1)
	if err != nil {
		return err
	}
	standardE2E, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), standardControlDecoded, correctnessThreshold)
	if err != nil {
		return err
	}
	result.StandardEndToEnd = &standardE2E
	if !standardE2E.Pass {
		return evalModCausalWrite(result, outPath)
	}

	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastPacked, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput})
	if err != nil {
		return err
	}
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput})
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
	fastReal, _, err := fastEval.CoeffsToSlots(fastModUp)
	if err != nil {
		return err
	}
	standardReal, _, err := standardEval.CoeffsToSlots(standardModUp)
	if err != nil {
		return err
	}
	zeroBootstrap := zeroSecret(btp.BootstrappingParameters)
	c2sStage, c2sPass, err := semanticBisectStageEvidence("coeffs_to_slots_real", fastReal, standardReal, btp.BootstrappingParameters, zeroBootstrap, skN2, nil, false)
	if err != nil {
		return err
	}
	result.C2SReal = &c2sStage
	if !c2sPass {
		return evalModCausalWrite(result, outPath)
	}
	_, fastC2SValues, err := semanticBisectView(fastReal, btp.BootstrappingParameters, zeroBootstrap)
	if err != nil {
		return err
	}
	_, standardC2SValues, err := semanticBisectView(standardReal, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}

	fastNativeBefore := fastReal.CopyNew()
	standardNativeBefore := standardReal.CopyNew()
	standardFromFast, err := evalModStandardFromFast(btp.BootstrappingParameters, fastReal, fastC2SValues, skN2)
	if err != nil {
		result.FirstSupportedCause = "logn13_evalmod_matched_input_construction_failure"
		return evalModCausalWrite(result, outPath)
	}
	fastFromStandard, err := evalModFastFromStandard(btp.BootstrappingParameters, standardReal, standardC2SValues)
	if err != nil {
		result.FirstSupportedCause = "logn13_evalmod_matched_input_construction_failure"
		return evalModCausalWrite(result, outPath)
	}
	standardFromFastEvidence, err := evalModMatchedInput("standard_from_fast_semantic_input", fastReal, standardFromFast, btp.BootstrappingParameters, zeroBootstrap, skN2)
	if err != nil {
		return err
	}
	fastFromStandardEvidence, err := evalModMatchedInput("fast_from_standard_semantic_input", standardReal, fastFromStandard, btp.BootstrappingParameters, skN2, zeroBootstrap)
	if err != nil {
		return err
	}
	result.StandardFromFast = &standardFromFastEvidence
	result.FastFromStandard = &fastFromStandardEvidence
	if !standardFromFastEvidence.Pass || !fastFromStandardEvidence.Pass {
		result.FirstSupportedCause = "logn13_evalmod_matched_input_construction_failure"
		return evalModCausalWrite(result, outPath)
	}

	stdNative, err := standardEval.EvalMod(standardReal.CopyNew())
	if err != nil {
		return err
	}
	fastNative, err := fastEval.EvalMod(fastReal.CopyNew())
	if err != nil {
		return err
	}
	stdOnFast, err := standardEval.EvalMod(standardFromFast.CopyNew())
	if err != nil {
		return err
	}
	fastOnStd, err := fastEval.EvalMod(fastFromStandard.CopyNew())
	if err != nil {
		return err
	}
	result.NativeInputsUnchanged = fix001P3E2EInputUnchanged(fastNativeBefore, fastReal) && fix001P3E2EInputUnchanged(standardNativeBefore, standardReal)
	acceptedRows, err := loadAcceptedNormalizedFinalRows(primaryRoot)
	if err != nil {
		return err
	}
	fastAccepted := doubleAngleRowsMatch(finalizationEvidence(fastNative).Rows, acceptedRows)
	stdNativeEvidence, stdNativeValues, err := evalModCausalOutputEvidence("STD_NATIVE", stdNative, btp.BootstrappingParameters, skN2, nil)
	if err != nil {
		return err
	}
	fastNativeEvidence, fastNativeValues, err := evalModCausalOutputEvidence("FAST_NATIVE", fastNative, btp.BootstrappingParameters, zeroBootstrap, &fastAccepted)
	if err != nil {
		return err
	}
	stdOnFastEvidence, stdOnFastValues, err := evalModCausalOutputEvidence("STD_ON_FAST", stdOnFast, btp.BootstrappingParameters, skN2, nil)
	if err != nil {
		return err
	}
	fastOnStdEvidence, fastOnStdValues, err := evalModCausalOutputEvidence("FAST_ON_STD", fastOnStd, btp.BootstrappingParameters, zeroBootstrap, nil)
	if err != nil {
		return err
	}
	result.Outputs = []evalModCausalOutput{stdNativeEvidence, fastNativeEvidence, stdOnFastEvidence, fastOnStdEvidence}

	comparisons := evalModCausalComparisons{}
	if comparisons.FastNativeVsStdNative, err = evalModCausalMetric(stdNativeValues, fastNativeValues); err != nil {
		return err
	}
	if comparisons.FastNativeVsStdOnFast, err = evalModCausalMetric(stdOnFastValues, fastNativeValues); err != nil {
		return err
	}
	if comparisons.StdNativeVsFastOnStd, err = evalModCausalMetric(stdNativeValues, fastOnStdValues); err != nil {
		return err
	}
	if comparisons.StdNativeVsStdOnFast, err = evalModCausalMetric(stdNativeValues, stdOnFastValues); err != nil {
		return err
	}
	if comparisons.FastNativeVsFastOnStd, err = evalModCausalMetric(fastNativeValues, fastOnStdValues); err != nil {
		return err
	}
	result.Comparisons = &comparisons
	result.NativeEvalMod = &comparisons.FastNativeVsStdNative
	if comparisons.FastNativeVsStdNative.Pass || !result.NativeInputsUnchanged {
		return evalModCausalWrite(result, outPath)
	}
	classification, verdict, asymmetry := evalModCausalClassification(comparisons)
	result.ProvisionalMatchedInputVerdict = verdict
	result.MatchedInputAsymmetry = asymmetry
	if predecessorC2SRealMaxComponent != 0 {
		result.AmplificationRatio = comparisons.StdNativeVsStdOnFast.MaxComponentAbs / predecessorC2SRealMaxComponent
	}
	if classification != "" {
		result.FirstSupportedCause = classification
		return evalModCausalWrite(result, outPath)
	}

	params := btp.BootstrappingParameters
	standardInternal := standardReal.CopyNew()
	fastInternal := fastFromStandard.CopyNew()
	standardInputScale := standardInternal.Scale
	fastInputScale := fastInternal.Scale
	standardInternal.Scale = standardEval.Mod1Evaluator.Parameters.ScalingFactor()
	fastInternal.Scale = fastEval.Mod1Parameters.ScalingFactor()
	offset := evalModCausalOffset(fastEval)
	if err := standardEval.Evaluator.Add(standardInternal, offset, standardInternal); err != nil {
		return err
	}
	if err := fastEval.FastCKKS.Add(fastInternal, offset, fastInternal); err != nil {
		return err
	}
	normalizeEvidence, err := evalModInternalEvidence("normalize_offset", fastInternal, standardInternal, params, zeroBootstrap, skN2, false)
	if err != nil {
		return err
	}
	result.InternalCheckpoints = append(result.InternalCheckpoints, normalizeEvidence)
	if !normalizeEvidence.Metric.Pass {
		result.FirstFailingInternalCheckpoint = "normalize_offset"
		result.FirstSupportedCause = "logn13_evalmod_first_internal_divergence_normalize_offset"
		return evalModCausalWrite(result, outPath)
	}

	targetScale, err := evalModCausalTargetScale(fastInternal, fastEval)
	if err != nil {
		return err
	}
	standardPolynomial, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardInternal, standardEval.Mod1Evaluator.Parameters.Mod1Poly.Clone(), targetScale)
	if err != nil {
		return err
	}
	planScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 91))
	fastPolynomial, err := fastEval.PolynomialEvaluator.EvaluateWithPlanScale(fastInternal, fastEval.Mod1Parameters.Mod1Poly, targetScale, planScale)
	if err != nil {
		return err
	}
	polynomialEvidence, err := evalModInternalEvidence("polynomial", fastPolynomial, standardPolynomial, params, zeroBootstrap, skN2, true)
	if err != nil {
		return err
	}
	result.InternalCheckpoints = append(result.InternalCheckpoints, polynomialEvidence)
	if !polynomialEvidence.Metric.Pass {
		result.FirstFailingInternalCheckpoint = "polynomial"
		result.FirstSupportedCause = "logn13_evalmod_first_internal_divergence_polynomial"
		return evalModCausalWrite(result, outPath)
	}

	workingScale := fastPolynomial.Scale
	currentKExponent, err := normalizedPowerOfTwoExponent(targetScale, workingScale)
	if err != nil {
		return err
	}
	if currentKExponent < 0 {
		return fmt.Errorf("negative initial normalized exponent: %d", currentKExponent)
	}
	result.DerivedInitialKExponent = currentKExponent
	initialK := new(big.Int).Lsh(big.NewInt(1), uint(currentKExponent))
	fastPolynomial.Scale = fastPolynomial.Scale.Mul(rlwe.NewScale(initialK))
	initialEvidence, err := evalModInternalScaledEvidence("initial_normalization", fastPolynomial, standardPolynomial, params, zeroBootstrap, skN2, float64FromPowerOfTwo(currentKExponent), false)
	if err != nil {
		return err
	}
	initialEvidence.KInExponent = currentKExponent
	result.InternalCheckpoints = append(result.InternalCheckpoints, initialEvidence)
	if !initialEvidence.Metric.Pass {
		result.FirstFailingInternalCheckpoint = "initial_normalization"
		result.FirstSupportedCause = "logn13_evalmod_first_internal_divergence_initial_normalization"
		return evalModCausalWrite(result, outPath)
	}

	fastCurrent := fastPolynomial
	standardCurrent := standardPolynomial
	sqrt2pi := fastEval.Mod1Parameters.Sqrt2Pi
	for round := 0; round < fastEval.Mod1Parameters.DoubleAngle; round++ {
		beforeLevel := fastCurrent.Level()
		preSquare, err := evalModInternalScaledEvidence(fmt.Sprintf("da_round%d_pre_square_capacity", round), fastCurrent, standardCurrent, params, zeroBootstrap, skN2, float64FromPowerOfTwo(currentKExponent), true)
		if err != nil {
			return err
		}
		preSquare.Round = round
		preSquare.KInExponent = currentKExponent
		result.InternalCheckpoints = append(result.InternalCheckpoints, preSquare)
		if !preSquare.FastCapacity.Pass || !preSquare.StandardCapacity.Pass {
			result.FirstFailingInternalCheckpoint = preSquare.Checkpoint
			result.FirstSupportedCause = "logn13_evalmod_internal_q01_capacity_failure"
			return evalModCausalWrite(result, outPath)
		}

		nextScale := fastCurrent.Scale.Mul(fastCurrent.Scale).Div(rlwe.NewScale(params.Q()[beforeLevel]))
		nextKExponent, err := normalizedPowerOfTwoExponent(nextScale, workingScale)
		if err != nil {
			return err
		}
		if nextKExponent < 0 {
			return fmt.Errorf("negative normalized exponent in round %d: %d", round, nextKExponent)
		}
		multiplierExponent := 1 + 2*currentKExponent - nextKExponent
		if multiplierExponent < 0 {
			return fmt.Errorf("negative multiplier exponent in round %d: %d", round, multiplierExponent)
		}

		if err := fastEval.FastCKKS.MulRelin(fastCurrent, fastCurrent, fastCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.MulRelin(standardCurrent, standardCurrent, standardCurrent); err != nil {
			return err
		}
		squareEvidence, err := evalModInternalScaledEvidence(fmt.Sprintf("da_round%d_square", round), fastCurrent, standardCurrent, params, zeroBootstrap, skN2, float64FromPowerOfTwo(2*currentKExponent), true)
		if err != nil {
			return err
		}
		squareEvidence.Round = round
		squareEvidence.KInExponent = currentKExponent
		squareEvidence.KOutExponent = nextKExponent
		squareEvidence.MultiplierExponent = multiplierExponent
		result.InternalCheckpoints = append(result.InternalCheckpoints, squareEvidence)
		if !squareEvidence.FastCapacity.Pass || !squareEvidence.StandardCapacity.Pass {
			result.FirstFailingInternalCheckpoint = squareEvidence.Checkpoint
			result.FirstSupportedCause = "logn13_evalmod_internal_q01_capacity_failure"
			return evalModCausalWrite(result, outPath)
		}
		if !squareEvidence.Metric.Pass {
			result.FirstFailingInternalCheckpoint = squareEvidence.Checkpoint
			result.FirstSupportedCause = fmt.Sprintf("logn13_evalmod_first_internal_divergence_da_round%d_square", round)
			return evalModCausalWrite(result, outPath)
		}

		factor := new(big.Int).Lsh(big.NewInt(1), uint(multiplierExponent))
		if err := fastEval.FastCKKS.MulIntegerMaintained(fastCurrent, factor, fastCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.Add(standardCurrent, standardCurrent, standardCurrent); err != nil {
			return err
		}
		multiplierEvidence, err := evalModInternalScaledEvidence(fmt.Sprintf("da_round%d_multiplier", round), fastCurrent, standardCurrent, params, zeroBootstrap, skN2, float64FromPowerOfTwo(nextKExponent), false)
		if err != nil {
			return err
		}
		multiplierEvidence.Round = round
		multiplierEvidence.KInExponent = currentKExponent
		multiplierEvidence.KOutExponent = nextKExponent
		multiplierEvidence.MultiplierExponent = multiplierExponent
		result.InternalCheckpoints = append(result.InternalCheckpoints, multiplierEvidence)
		if !multiplierEvidence.Metric.Pass {
			result.FirstFailingInternalCheckpoint = multiplierEvidence.Checkpoint
			result.FirstSupportedCause = fmt.Sprintf("logn13_evalmod_first_internal_divergence_da_round%d_multiplier", round)
			return evalModCausalWrite(result, outPath)
		}

		sqrt2pi *= sqrt2pi
		normalizedConstant, _ := new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(nextKExponent)))).Float64()
		if err := fastEval.FastCKKS.Add(fastCurrent, -normalizedConstant, fastCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.Add(standardCurrent, -sqrt2pi, standardCurrent); err != nil {
			return err
		}
		constantEvidence, err := evalModInternalScaledEvidence(fmt.Sprintf("da_round%d_constant", round), fastCurrent, standardCurrent, params, zeroBootstrap, skN2, float64FromPowerOfTwo(nextKExponent), true)
		if err != nil {
			return err
		}
		constantEvidence.Round = round
		constantEvidence.KInExponent = currentKExponent
		constantEvidence.KOutExponent = nextKExponent
		constantEvidence.MultiplierExponent = multiplierExponent
		constantEvidence.NormalizedConstant = normalizedConstant
		result.InternalCheckpoints = append(result.InternalCheckpoints, constantEvidence)
		if !constantEvidence.FastCapacity.Pass || !constantEvidence.StandardCapacity.Pass {
			result.FirstFailingInternalCheckpoint = constantEvidence.Checkpoint
			result.FirstSupportedCause = "logn13_evalmod_internal_q01_capacity_failure"
			return evalModCausalWrite(result, outPath)
		}
		if !constantEvidence.Metric.Pass {
			result.FirstFailingInternalCheckpoint = constantEvidence.Checkpoint
			result.FirstSupportedCause = fmt.Sprintf("logn13_evalmod_first_internal_divergence_da_round%d_constant", round)
			return evalModCausalWrite(result, outPath)
		}

		if err := fastEval.FastCKKS.Rescale(fastCurrent, fastCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.Rescale(standardCurrent, standardCurrent); err != nil {
			return err
		}
		rescaleEvidence, err := evalModInternalScaledEvidence(fmt.Sprintf("da_round%d_rescale", round), fastCurrent, standardCurrent, params, zeroBootstrap, skN2, float64FromPowerOfTwo(nextKExponent), false)
		if err != nil {
			return err
		}
		rescaleEvidence.Round = round
		rescaleEvidence.KInExponent = currentKExponent
		rescaleEvidence.KOutExponent = nextKExponent
		rescaleEvidence.MultiplierExponent = multiplierExponent
		result.InternalCheckpoints = append(result.InternalCheckpoints, rescaleEvidence)
		if !rescaleEvidence.Metric.Pass {
			result.FirstFailingInternalCheckpoint = rescaleEvidence.Checkpoint
			result.FirstSupportedCause = fmt.Sprintf("logn13_evalmod_first_internal_divergence_da_round%d_rescale", round)
			return evalModCausalWrite(result, outPath)
		}
		currentKExponent = nextKExponent
	}

	result.DerivedFinalKExponent = currentKExponent
	finalK := new(big.Int).Lsh(big.NewInt(1), uint(currentKExponent))
	if err := fastEval.FastCKKS.MulIntegerMaintained(fastCurrent, finalK, fastCurrent); err != nil {
		return err
	}
	finalRestoreEvidence, err := evalModInternalEvidence("final_restore", fastCurrent, standardCurrent, params, zeroBootstrap, skN2, false)
	if err != nil {
		return err
	}
	finalRestoreEvidence.KInExponent = currentKExponent
	result.InternalCheckpoints = append(result.InternalCheckpoints, finalRestoreEvidence)
	if !finalRestoreEvidence.Metric.Pass {
		result.FirstFailingInternalCheckpoint = "final_restore"
		result.FirstSupportedCause = "logn13_evalmod_first_internal_divergence_final_restore"
		return evalModCausalWrite(result, outPath)
	}

	fastCurrent.Scale = fastInputScale
	standardCurrent.Scale = standardInputScale
	finalScaleEvidence, err := evalModInternalEvidence("final_scale_reset", fastCurrent, standardCurrent, params, zeroBootstrap, skN2, false)
	if err != nil {
		return err
	}
	finalScaleEvidence.KInExponent = currentKExponent
	result.InternalCheckpoints = append(result.InternalCheckpoints, finalScaleEvidence)
	if !finalScaleEvidence.Metric.Pass {
		result.FirstFailingInternalCheckpoint = "final_scale_reset"
		result.FirstSupportedCause = "logn13_evalmod_first_internal_divergence_final_scale_reset"
		return evalModCausalWrite(result, outPath)
	}

	result.FirstSupportedCause = "logn13_evalmod_internal_replay_oracle_mismatch"
	return evalModCausalWrite(result, outPath)
}

func float64FromPowerOfTwo(exponent int) float64 {
	value, _ := new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(exponent))).Float64()
	return value
}
