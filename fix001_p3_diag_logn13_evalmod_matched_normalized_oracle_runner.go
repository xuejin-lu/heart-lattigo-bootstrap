package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredEvalModMatchedPrimaryBase = "3e4bb1bf4ba877ea3516292486e6dd12b331f832"
	requiredEvalModMatchedSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
	evalModMatchedLocalThreshold      = 1e-2
	evalModMatchedStageThreshold      = 1e-6
)

type evalModMatchedMetric struct {
	Metric    *PSGlobalMetric `json:"metric"`
	Threshold float64         `json:"threshold"`
}

type evalModMatchedC2S struct {
	Level                    int                    `json:"level"`
	Degree                   int                    `json:"degree"`
	Scale                    string                 `json:"scale"`
	Fast                     semanticBisectMetadata `json:"fast"`
	AlignedStandard          semanticBisectMetadata `json:"aligned_standard"`
	OrdinaryStandard         semanticBisectMetadata `json:"ordinary_standard"`
	FastVsAlignedStandard    *PSGlobalMetric        `json:"fast_vs_aligned_standard"`
	FastVsOrdinaryStandard   *PSGlobalMetric        `json:"fast_vs_ordinary_standard"`
	FastVsAlignedImag        *PSGlobalMetric        `json:"fast_imag_vs_aligned_standard"`
	FastVsOrdinaryImag       *PSGlobalMetric        `json:"fast_imag_vs_ordinary_standard"`
	AlignedVsOrdinary        *PSGlobalMetric        `json:"aligned_vs_ordinary_standard"`
	FastQ01CapacityPass      bool                   `json:"fast_q01_capacity_pass"`
	FastImagQ01CapacityPass  bool                   `json:"fast_imag_q01_capacity_pass"`
	AlignedFullRNSCapacity   bool                   `json:"aligned_full_rns_capacity_pass"`
	NoFurtherDFT             bool                   `json:"no_further_dft_factor"`
	AcceptedGroup0K4         bool                   `json:"accepted_group0_k4"`
	AcceptedGroup1K2         bool                   `json:"accepted_group1_k2"`
	Groups2And3Once          bool                   `json:"groups2_and3_executed_once"`
	C2SPrecisionBlockerFixed bool                   `json:"c2s_precision_blocker_resolved"`
}

type evalModMatchedCheckpoint struct {
	Name               string                    `json:"name"`
	Round              int                       `json:"round,omitempty"`
	Fast               semanticBisectMetadata    `json:"fast"`
	Normalized         semanticBisectMetadata    `json:"normalized"`
	Standard           semanticBisectMetadata    `json:"standard,omitempty"`
	FastQ01Safe        bool                      `json:"fast_q01_capacity_safe"`
	NormalizedSafe     bool                      `json:"normalized_full_rns_capacity_safe"`
	FastQ01Capacity    *postMod1S2CCapacity      `json:"fast_q01_capacity,omitempty"`
	NormalizedCapacity *NormalizedExactCapacity  `json:"normalized_full_rns_capacity,omitempty"`
	RowsMatch          bool                      `json:"fast_vs_normalized_q01_rows_match"`
	RowComparison      NormalizedStageComparison `json:"first_row_mismatch"`
	FastVsNormalized   *PSGlobalMetric           `json:"fast_vs_normalized"`
	NormalizedVsStd    *PSGlobalMetric           `json:"normalized_vs_standard"`
}

type evalModMatchedRound struct {
	Round              int                      `json:"round"`
	KInExponent        int                      `json:"k_in_exponent"`
	KOutExponent       int                      `json:"k_out_exponent"`
	MultiplierExponent int                      `json:"multiplier_exponent"`
	NormalizedConstant float64                  `json:"normalized_constant"`
	Input              evalModMatchedCheckpoint `json:"input"`
	Square             evalModMatchedCheckpoint `json:"square"`
	AfterMultiplier    evalModMatchedCheckpoint `json:"after_multiplier"`
	AfterConstant      evalModMatchedCheckpoint `json:"after_constant"`
	PostRescale        evalModMatchedCheckpoint `json:"post_rescale"`
	NVsCPostRescale    *PSGlobalMetric          `json:"normalized_mapped_vs_coherent"`
	CVsSPostRescale    *PSGlobalMetric          `json:"coherent_vs_standard"`
}

type evalModMatchedFinal struct {
	FastVsNormalized                *PSGlobalMetric                 `json:"fast_vs_normalized"`
	NormalizedVsCoherent            *PSGlobalMetric                 `json:"normalized_vs_coherent"`
	CoherentVsStandard              *PSGlobalMetric                 `json:"coherent_vs_standard"`
	FastVsStandard                  *PSGlobalMetric                 `json:"fast_vs_standard"`
	NormalizedVsPlainOracle         *PSGlobalMetric                 `json:"normalized_vs_plaintext_oracle"`
	StandardVsPlainOracle           *PSGlobalMetric                 `json:"standard_vs_plaintext_oracle"`
	NormalizedBeforeRestoreCapacity *NormalizedExactCapacity        `json:"normalized_before_restore_capacity,omitempty"`
	FinalRestoreCapacity            *NormalizedFinalRestoreCapacity `json:"final_restore_capacity,omitempty"`
	FastRowsMatchNormalized         bool                            `json:"fast_rows_match_normalized"`
	RestoreRowsMatch                bool                            `json:"restore_rows_match"`
	ScaleResetRowsUnchanged         bool                            `json:"scale_reset_rows_unchanged"`
	LevelDegreeUnchanged            bool                            `json:"level_degree_unchanged"`
}

type evalModMatchedCausalEffects struct {
	FastImplementationEffect   *PSGlobalMetric `json:"fast_implementation_effect_f_vs_n"`
	NormalizedRecurrenceEffect *PSGlobalMetric `json:"normalized_recurrence_effect_n_vs_c"`
	CompressedPolynomialEffect *PSGlobalMetric `json:"compressed_polynomial_effect_c_vs_s"`
	TotalFastVsStandard        *PSGlobalMetric `json:"total_fast_vs_standard_f_vs_s"`
	Identity                   string          `json:"identity"`
}

type evalModMatchedResult struct {
	SchemaVersion          string                      `json:"schema_version"`
	Timestamp              time.Time                   `json:"timestamp"`
	Primary                RepositoryMetadata          `json:"primary_repository"`
	Lattigo                RepositoryMetadata          `json:"lattigo_repository"`
	Environment            EnvironmentMetadata         `json:"environment"`
	Config                 BootstrapConfig             `json:"config"`
	Parameters             ExperimentParameters        `json:"effective_parameters"`
	Workload               CorrectnessWorkload         `json:"workload"`
	StandardProof          semanticBisectStandardProof `json:"standard_control_proof"`
	CorrectedC2S           evalModMatchedC2S           `json:"corrected_c2s_matched_input"`
	TargetScale            string                      `json:"target_scale"`
	PlanScale              string                      `json:"plan_scale"`
	CompressedPolyScale    string                      `json:"compressed_polynomial_scale"`
	NormalizedPolyCapacity *NormalizedExactCapacity    `json:"normalized_polynomial_capacity,omitempty"`
	PlaintextOracle        string                      `json:"authoritative_plaintext_oracle"`
	Prefix                 []evalModMatchedCheckpoint  `json:"scale_normalization_offset_checkpoints"`
	Polynomial             map[string]*PSGlobalMetric  `json:"polynomial_semantic_metrics"`
	PolynomialMeta         map[string]string           `json:"polynomial_metadata"`
	Rounds                 []evalModMatchedRound       `json:"normalized_rounds"`
	FinalReal              evalModMatchedFinal         `json:"final_real"`
	FinalImag              evalModMatchedFinal         `json:"final_imag"`
	CausalEffects          evalModMatchedCausalEffects `json:"causal_effects_real"`
	Classification         string                      `json:"classification"`
	FirstFailing           string                      `json:"first_failing_stage_or_checkpoint"`
	Validation             map[string]interface{}      `json:"validation"`
}

type evalModMatchedInputs struct {
	FastReal, FastImag                     *rlwe.Ciphertext
	AlignedReal, AlignedImag               *rlwe.Ciphertext
	OrdinaryReal, OrdinaryImag             *rlwe.Ciphertext
	FastSK, StandardSK                     *rlwe.SecretKey
	FastVsAlignedReal, FastVsAlignedImag   *PSGlobalMetric
	FastVsOrdinaryReal, FastVsOrdinaryImag *PSGlobalMetric
	FastQ01Safe, FastImagQ01Safe           bool
	AlignedCapacitySafe                    bool
}

func evalModMatchedMetricFor(reference, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(reference, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func evalModMatchedScaleValues(values []complex128, factor *big.Int) []complex128 {
	result := make([]complex128, len(values))
	scale, _ := new(big.Float).SetInt(factor).Float64()
	for i, value := range values {
		result[i] = value * complex(scale, 0)
	}
	return result
}

func evalModMatchedPlainDoubleAngle(values []complex128, sqrt2pi float64, rounds int) []complex128 {
	result := append([]complex128(nil), values...)
	constant := sqrt2pi
	for round := 0; round < rounds; round++ {
		constant *= constant
		for i, value := range result {
			result[i] = 2*value*value - complex(constant, 0)
		}
	}
	return result
}

func evalModMatchedCheckpointFor(name string, round int, fast, normalized, standard *rlwe.Ciphertext, params ckks.Parameters, fastSK, normalizedSK, standardSK *rlwe.SecretKey) (evalModMatchedCheckpoint, error) {
	fastMeta, fastValues, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return evalModMatchedCheckpoint{}, err
	}
	normalizedMeta, normalizedValues, err := semanticBisectView(normalized, params, normalizedSK)
	if err != nil {
		return evalModMatchedCheckpoint{}, err
	}
	rowComparison := normalizedStageComparison(fast, normalized)
	checkpoint := evalModMatchedCheckpoint{Name: name, Round: round, Fast: fastMeta, Normalized: normalizedMeta, RowsMatch: rowComparison.Match, RowComparison: rowComparison}
	checkpoint.FastVsNormalized = evalModMatchedMetricFor(normalizedValues, fastValues, evalModMatchedStageThreshold)
	if standard != nil {
		standardMeta, standardValues, err := semanticBisectView(standard, params, standardSK)
		if err != nil {
			return evalModMatchedCheckpoint{}, err
		}
		checkpoint.Standard = standardMeta
		checkpoint.NormalizedVsStd = evalModMatchedMetricFor(standardValues, normalizedValues, evalModMatchedLocalThreshold)
	}
	if fast.Level() >= 1 {
		capacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
		if err != nil {
			return evalModMatchedCheckpoint{}, err
		}
		checkpoint.FastQ01Safe = capacity.Pass
		checkpoint.FastQ01Capacity = &capacity
	}
	if normalized.Level() >= 1 {
		capacity, err := normalizedExactCapacity(params, normalized)
		if err != nil {
			return evalModMatchedCheckpoint{}, err
		}
		checkpoint.NormalizedSafe = capacity.Pass
		checkpoint.NormalizedCapacity = &capacity
	}
	return checkpoint, nil
}

func evalModMatchedC2SInputs(cfg BootstrapConfig, btp bootstrapping.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, standardSK *rlwe.SecretKey) (evalModMatchedInputs, error) {
	params := btp.BootstrappingParameters
	residual := btp.ResidualParameters
	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastPacked, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput})
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput})
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	fastScaled, _, err := fastEval.ScaleDown(&fastPacked[0])
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	standardScaled, _, err := standardEval.ScaleDown(&standardPacked[0])
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	fastModUp, err := fastEval.ModUp(fastScaled)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	standardModUp, err := standardEval.ModUp(standardScaled)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	original := standardEval.C2SDFTMatrix
	if len(original.Matrices) != 4 || len(original.Levels) != 4 {
		return evalModMatchedInputs{}, fmt.Errorf("expected four C2S matrices")
	}
	for _, level := range original.Levels {
		if level != 1 {
			return evalModMatchedInputs{}, fmt.Errorf("unexpected C2S matrix level %d", level)
		}
	}
	fastSK := zeroSecret(params)
	ordinaryReal, ordinaryImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	compressed0, _, err := c2sCompressedMatrix(params, original, 4)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	stdCurrent, fastCurrent, err := c2sCompressedLinearTransform(params, standardEval, fastEval, compressed0, standardModUp.CopyNew(), fastModUp.CopyNew())
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	if err = c2sCompressedRescale(params, standardEval, fastEval, stdCurrent, fastCurrent); err != nil {
		return evalModMatchedInputs{}, err
	}
	if err = c2sCompressedApplyRestore(params, fastEval.FastCKKS, fastCurrent, stdCurrent, 4); err != nil {
		return evalModMatchedInputs{}, err
	}
	group0Std, group0Fast := stdCurrent, fastCurrent
	compressed1, _, err := c2sGroup1CompressedMatrix(params, original, 1, 2)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	stdCurrent, fastCurrent, err = c2sCompressedLinearTransform(params, standardEval, fastEval, compressed1, group0Std.CopyNew(), group0Fast.CopyNew())
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	if err = c2sCompressedRescale(params, standardEval, fastEval, stdCurrent, fastCurrent); err != nil {
		return evalModMatchedInputs{}, err
	}
	if err = c2sCompressedApplyRestore(params, fastEval.FastCKKS, fastCurrent, stdCurrent, 2); err != nil {
		return evalModMatchedInputs{}, err
	}
	for group := 2; group <= 3; group++ {
		stdCurrent, fastCurrent, err = c2sCompressedLinearTransform(params, standardEval, fastEval, original.Matrices[group], stdCurrent, fastCurrent)
		if err != nil {
			return evalModMatchedInputs{}, err
		}
		if err = c2sCompressedRescale(params, standardEval, fastEval, stdCurrent, fastCurrent); err != nil {
			return evalModMatchedInputs{}, err
		}
	}
	realSplit, imagSplit, err := evalModMatchedSplitOnly(params, standardEval, fastEval, stdCurrent, fastCurrent)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_, fastRealValues, err := semanticBisectView(realSplit.Fast, params, fastSK)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_, alignedRealValues, err := semanticBisectView(realSplit.Standard, params, standardSK)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_, ordinaryRealValues, err := semanticBisectView(ordinaryReal, params, standardSK)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_, fastImagValues, err := semanticBisectView(imagSplit.Fast, params, fastSK)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_, alignedImagValues, err := semanticBisectView(imagSplit.Standard, params, standardSK)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_, ordinaryImagValues, err := semanticBisectView(ordinaryImag, params, standardSK)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	fastRealMetric := evalModMatchedMetricFor(alignedRealValues, fastRealValues, acceptedC2SGroup1Budget)
	fastImagMetric := evalModMatchedMetricFor(alignedImagValues, fastImagValues, acceptedC2SGroup1Budget)
	fastRealOrdinaryMetric := evalModMatchedMetricFor(ordinaryRealValues, fastRealValues, acceptedC2SGroup1Budget)
	fastImagOrdinaryMetric := evalModMatchedMetricFor(ordinaryImagValues, fastImagValues, acceptedC2SGroup1Budget)
	alignedRealOrdinaryMetric := evalModMatchedMetricFor(ordinaryRealValues, alignedRealValues, acceptedC2SGroup1Budget)
	alignedImagOrdinaryMetric := evalModMatchedMetricFor(ordinaryImagValues, alignedImagValues, acceptedC2SGroup1Budget)
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, realSplit.Fast)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	alignedCapacity, err := postMod1S2CCapacityFromCiphertext(params, realSplit.Standard)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	fastImagCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, imagSplit.Fast)
	if err != nil {
		return evalModMatchedInputs{}, err
	}
	_ = alignedRealOrdinaryMetric
	_ = alignedImagOrdinaryMetric
	return evalModMatchedInputs{
		FastReal: realSplit.Fast, FastImag: imagSplit.Fast, AlignedReal: realSplit.Standard, AlignedImag: imagSplit.Standard,
		OrdinaryReal: ordinaryReal, OrdinaryImag: ordinaryImag, FastSK: fastSK, StandardSK: standardSK,
		FastVsAlignedReal: fastRealMetric, FastVsAlignedImag: fastImagMetric, FastVsOrdinaryReal: fastRealOrdinaryMetric, FastVsOrdinaryImag: fastImagOrdinaryMetric,
		FastQ01Safe: fastCapacity.Pass, FastImagQ01Safe: fastImagCapacity.Pass, AlignedCapacitySafe: alignedCapacity.Pass,
	}, nil
}

type evalModMatchedSplit struct{ Fast, Standard *rlwe.Ciphertext }

func evalModMatchedSplitOnly(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, standard, fast *rlwe.Ciphertext) (evalModMatchedSplit, evalModMatchedSplit, error) {
	fastConj := fastckks.NewCiphertext(params, 1, fast.Level())
	*fastConj.MetaData = *fast.MetaData
	standardConj := ckks.NewCiphertext(params, 1, standard.Level())
	*standardConj.MetaData = *standard.MetaData
	if err := fastEval.FastCKKS.Conjugate(fast, fastConj); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	if err := standardEval.Evaluator.Conjugate(standard, standardConj); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	fastImagSub := fastckks.NewCiphertext(params, 1, fast.Level())
	*fastImagSub.MetaData = *fast.MetaData
	standardImagSub := ckks.NewCiphertext(params, 1, standard.Level())
	*standardImagSub.MetaData = *standard.MetaData
	if err := fastEval.FastCKKS.Sub(fast, fastConj, fastImagSub); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	if err := standardEval.Evaluator.Sub(standard, standardConj, standardImagSub); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	fastImag := fastckks.NewCiphertext(params, 1, fast.Level())
	*fastImag.MetaData = *fast.MetaData
	standardImag := ckks.NewCiphertext(params, 1, standard.Level())
	*standardImag.MetaData = *standard.MetaData
	if err := fastEval.FastCKKS.Mul(fastImagSub, -1i, fastImag); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	if err := standardEval.Evaluator.Mul(standardImagSub, -1i, standardImag); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	fastReal := fastckks.NewCiphertext(params, 1, fast.Level())
	*fastReal.MetaData = *fast.MetaData
	standardReal := ckks.NewCiphertext(params, 1, standard.Level())
	*standardReal.MetaData = *standard.MetaData
	if err := fastEval.FastCKKS.Add(fastConj, fast, fastReal); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	if err := standardEval.Evaluator.Add(standardConj, standard, standardReal); err != nil {
		return evalModMatchedSplit{}, evalModMatchedSplit{}, err
	}
	return evalModMatchedSplit{Fast: fastReal, Standard: standardReal}, evalModMatchedSplit{Fast: fastImag, Standard: standardImag}, nil
}

type evalModMatchedPath struct {
	Prefix                          []evalModMatchedCheckpoint
	Polynomial                      map[string]*PSGlobalMetric
	PolyMeta                        map[string]string
	PolyCapacity                    *NormalizedExactCapacity
	Rounds                          []evalModMatchedRound
	FastFinal                       *rlwe.Ciphertext
	NormalizedFinal                 *rlwe.Ciphertext
	CoherentFinal                   *rlwe.Ciphertext
	StandardFinal                   *rlwe.Ciphertext
	PlainOracle                     []complex128
	PlainFinal                      []complex128
	FastPlain                       *PSGlobalMetric
	NormalizedPlain                 *PSGlobalMetric
	StandardPlain                   *PSGlobalMetric
	FastRowsMatch                   bool
	RestoreRowsMatch                bool
	ScaleResetRowsUnchanged         bool
	LevelDegreeUnchanged            bool
	FirstFailure                    string
	NormalizedBeforeRestoreCapacity *NormalizedExactCapacity
	FinalRestoreCapacity            *NormalizedFinalRestoreCapacity
}

func evalModMatchedAddFullConstant(params ckks.Parameters, input *rlwe.Ciphertext, constant float64) *rlwe.Ciphertext {
	return normalizedFullSubtractConstant(params, input, -constant)
}

func evalModMatchedRunPath(fastInput, standardInput *rlwe.Ciphertext, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey) (evalModMatchedPath, error) {
	return evalModMatchedRunPathWithPlanScale(fastInput, standardInput, fastEval, standardEval, params, fastSK, standardSK, rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 91)))
}

func evalModMatchedRunPathWithPlanScale(fastInput, standardInput *rlwe.Ciphertext, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, planScale rlwe.Scale) (evalModMatchedPath, error) {
	return evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastInput, standardInput, fastEval, standardEval, params, fastSK, standardSK, planScale, nil)
}

func evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastInput, standardInput *rlwe.Ciphertext, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, planScale rlwe.Scale, fastPolyOverride *rlwe.Ciphertext) (evalModMatchedPath, error) {
	path := evalModMatchedPath{Polynomial: map[string]*PSGlobalMetric{}, PolyMeta: map[string]string{}, FirstFailure: "none"}
	coefficientView, err := targetScaleCoefficientView(params, fastInput)
	if err != nil {
		return path, err
	}
	inputData, err := targetScaleCenteredRows(params, coefficientView)
	if err != nil {
		return path, err
	}
	normalizedInput, err := normalizedFullFromLow(params, fastInput, inputData)
	if err != nil {
		return path, err
	}
	fastCurrent := fastInput.CopyNew()
	standardCurrent := standardInput.CopyNew()
	originalFastScale := fastCurrent.Scale
	originalStandardScale := standardCurrent.Scale
	addPrefix := func(name string, round int) error {
		checkpoint, e := evalModMatchedCheckpointFor(name, round, fastCurrent, normalizedInput, standardCurrent, params, fastSK, zeroSecret(params), standardSK)
		if e != nil {
			return e
		}
		path.Prefix = append(path.Prefix, checkpoint)
		if !checkpoint.RowsMatch && path.FirstFailure == "none" {
			path.FirstFailure = name
		}
		if !checkpoint.NormalizedSafe && path.FirstFailure == "none" {
			path.FirstFailure = name + ".normalized_capacity"
		}
		return nil
	}
	if err := addPrefix("evalmod_input", -1); err != nil {
		return path, err
	}
	mod1 := fastEval.Mod1Parameters
	fastCurrent.Scale = mod1.ScalingFactor()
	normalizedInput.Scale = mod1.ScalingFactor()
	standardCurrent.Scale = standardEval.Mod1Evaluator.Parameters.ScalingFactor()
	if err := addPrefix("normalize_scale", -1); err != nil {
		return path, err
	}
	offset := evalModCausalOffset(fastEval)
	if err := fastEval.FastCKKS.Add(fastCurrent, offset, fastCurrent); err != nil {
		return path, err
	}
	normalizedInput = evalModMatchedAddFullConstant(params, normalizedInput, float64FromBigFloat(offset))
	if err := standardEval.Evaluator.Add(standardCurrent, offset, standardCurrent); err != nil {
		return path, err
	}
	if err := addPrefix("apply_offset", -1); err != nil {
		return path, err
	}
	_, preprocessedValues, err := semanticBisectView(fastCurrent, params, fastSK)
	if err != nil {
		return path, err
	}
	path.PlainOracle = chebyshevRawOracle(mod1.Mod1Poly, preprocessedValues)
	path.PlainFinal = evalModMatchedPlainDoubleAngle(path.PlainOracle, fastEval.Mod1Parameters.Sqrt2Pi, fastEval.Mod1Parameters.DoubleAngle)
	targetScale, err := evalModCausalTargetScale(fastCurrent, fastEval)
	if err != nil {
		return path, err
	}
	standardPoly, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardCurrent.CopyNew(), standardEval.Mod1Evaluator.Parameters.Mod1Poly.Clone(), targetScale)
	if err != nil {
		return path, err
	}
	var fastPoly *rlwe.Ciphertext
	if fastPolyOverride != nil {
		fastPoly = fastPolyOverride.CopyNew()
	} else {
		fastPoly, err = fastEval.PolynomialEvaluator.EvaluateWithPlanScale(fastCurrent.CopyNew(), fastEval.Mod1Parameters.Mod1Poly, targetScale, planScale)
		if err != nil {
			return path, err
		}
	}
	polyView, err := targetScaleCoefficientView(params, fastPoly)
	if err != nil {
		return path, err
	}
	polyData, err := targetScaleCenteredRows(params, polyView)
	if err != nil {
		return path, err
	}
	normalizedPoly, err := normalizedFullFromLow(params, fastPoly, polyData)
	if err != nil {
		return path, err
	}
	path.PolyMeta["fast_level"] = fmt.Sprintf("%d", fastPoly.Level())
	path.PolyMeta["normalized_level"] = fmt.Sprintf("%d", normalizedPoly.Level())
	path.PolyMeta["standard_level"] = fmt.Sprintf("%d", standardPoly.Level())
	path.PolyMeta["fast_scale"] = finalizationScaleString(fastPoly.Scale)
	path.PolyMeta["normalized_scale"] = finalizationScaleString(normalizedPoly.Scale)
	path.PolyMeta["standard_scale"] = finalizationScaleString(standardPoly.Scale)
	path.PolyMeta["plan_scale"] = finalizationScaleString(planScale)
	path.PolyMeta["target_scale"] = finalizationScaleString(targetScale)
	_, fastPolyValues, err := semanticBisectView(fastPoly, params, fastSK)
	if err != nil {
		return path, err
	}
	_, normalizedPolyValues, err := semanticBisectView(normalizedPoly, params, zeroSecret(params))
	if err != nil {
		return path, err
	}
	_, standardPolyValues, err := semanticBisectView(standardPoly, params, standardSK)
	if err != nil {
		return path, err
	}
	path.Polynomial["fast_vs_normalized"] = evalModMatchedMetricFor(normalizedPolyValues, fastPolyValues, evalModMatchedStageThreshold)
	path.Polynomial["normalized_vs_standard"] = evalModMatchedMetricFor(standardPolyValues, normalizedPolyValues, evalModMatchedLocalThreshold)
	path.Polynomial["fast_vs_standard"] = evalModMatchedMetricFor(standardPolyValues, fastPolyValues, evalModMatchedLocalThreshold)
	path.Polynomial["fast_vs_plaintext_oracle"] = evalModMatchedMetricFor(path.PlainOracle, fastPolyValues, evalModMatchedLocalThreshold)
	path.Polynomial["normalized_vs_plaintext_oracle"] = evalModMatchedMetricFor(path.PlainOracle, normalizedPolyValues, evalModMatchedLocalThreshold)
	path.Polynomial["standard_vs_plaintext_oracle"] = evalModMatchedMetricFor(path.PlainOracle, standardPolyValues, evalModMatchedLocalThreshold)
	polyCapacity, err := normalizedExactCapacity(params, normalizedPoly)
	if err != nil {
		return path, err
	}
	path.PolyMeta["normalized_capacity_pass"] = fmt.Sprintf("%t", polyCapacity.Pass)
	path.PolyCapacity = &polyCapacity
	if !polyCapacity.Pass && path.FirstFailure == "none" {
		path.FirstFailure = "polynomial.normalized_capacity"
		return path, nil
	}
	if !path.Polynomial["fast_vs_normalized"].Pass && path.FirstFailure == "none" {
		path.FirstFailure = "polynomial.fast_vs_normalized"
	}
	workingScale := fastPoly.Scale
	kExponent, err := normalizedPowerOfTwoExponent(targetScale, workingScale)
	if err != nil {
		return path, err
	}
	if kExponent < 0 {
		return path, fmt.Errorf("negative initial normalized exponent %d", kExponent)
	}
	initialK := new(big.Int).Lsh(big.NewInt(1), uint(kExponent))
	fastNormalized := fastPoly.CopyNew()
	fastNormalized.Scale = fastNormalized.Scale.Mul(rlwe.NewScale(initialK))
	normalizedCurrent := normalizedPoly.CopyNew()
	normalizedCurrent.Scale = normalizedCurrent.Scale.Mul(rlwe.NewScale(initialK))
	sqrt2pi := fastEval.Mod1Parameters.Sqrt2Pi
	currentLevel := fastNormalized.Level()
	currentScale := fastNormalized.Scale
	kIn := kExponent
	coherentMultiplier := initialK
	coherentCurrent := normalizedFullIntegerMultiply(params, normalizedPoly, coherentMultiplier)
	coherentCurrent.Scale = normalizedPoly.Scale.Mul(rlwe.NewScale(coherentMultiplier))
	standardPolyCurrent := standardPoly.CopyNew()
	for round := 0; round < fastEval.Mod1Parameters.DoubleAngle; round++ {
		sqrt2pi *= sqrt2pi
		schedule, err := normalizedScaleSchedule(round, currentLevel, currentScale, workingScale, kIn, sqrt2pi, params)
		if err != nil {
			return path, err
		}
		inputCheckpoint, err := evalModMatchedCheckpointFor("da_input", round, fastNormalized, normalizedCurrent, coherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return path, err
		}
		inputCheckpoint.FastQ01Safe = inputCheckpoint.FastQ01Safe
		if !inputCheckpoint.RowsMatch && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.input_rows", round)
		}
		if !inputCheckpoint.NormalizedSafe && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.input_capacity", round)
			path.Rounds = append(path.Rounds, evalModMatchedRound{Round: round, KInExponent: kIn, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint})
			return path, nil
		}
		if !inputCheckpoint.FastVsNormalized.Pass && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.input_semantic", round)
		}
		refProduct, err := normalizedFullSquare(params, normalizedCurrent)
		if err != nil {
			return path, err
		}
		refSquare := normalizedFullZeroC2(params, refProduct)
		if err := fastEval.FastCKKS.MulRelin(fastNormalized, fastNormalized, fastNormalized); err != nil {
			return path, err
		}
		squareCheckpoint, err := evalModMatchedCheckpointFor("da_square", round, fastNormalized, refSquare, coherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return path, err
		}
		if !squareCheckpoint.RowsMatch && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.square_rows", round)
		}
		if !squareCheckpoint.NormalizedSafe && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.square_capacity", round)
			path.Rounds = append(path.Rounds, evalModMatchedRound{Round: round, KInExponent: kIn, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint})
			return path, nil
		}
		if !squareCheckpoint.FastVsNormalized.Pass && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.square_semantic", round)
		}
		factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
		if err := fastEval.FastCKKS.MulIntegerMaintained(fastNormalized, factor, fastNormalized); err != nil {
			return path, err
		}
		refAfterA := normalizedFullIntegerMultiply(params, refSquare, factor)
		aCheckpoint, err := evalModMatchedCheckpointFor("da_after_multiplier", round, fastNormalized, refAfterA, coherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return path, err
		}
		if !aCheckpoint.RowsMatch && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.after_multiplier_rows", round)
		}
		if !aCheckpoint.NormalizedSafe && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.after_multiplier_capacity", round)
			path.Rounds = append(path.Rounds, evalModMatchedRound{Round: round, KInExponent: kIn, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint})
			return path, nil
		}
		constantValue, _ := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
		if err := fastEval.FastCKKS.Add(fastNormalized, -constantValue, fastNormalized); err != nil {
			return path, err
		}
		refAfterConstant := normalizedFullSubtractConstant(params, refAfterA, constantValue)
		constantCheckpoint, err := evalModMatchedCheckpointFor("da_after_constant", round, fastNormalized, refAfterConstant, coherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return path, err
		}
		if !constantCheckpoint.NormalizedSafe && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.constant_capacity", round)
			path.Rounds = append(path.Rounds, evalModMatchedRound{Round: round, KInExponent: kIn, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint, AfterConstant: constantCheckpoint})
			return path, nil
		}
		if !constantCheckpoint.RowsMatch && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.constant_rows", round)
		}
		if err := fastEval.FastCKKS.Rescale(fastNormalized, fastNormalized); err != nil {
			return path, err
		}
		normalizedCurrent, err = normalizedFullRescale(params, refAfterConstant)
		if err != nil {
			return path, err
		}
		coherentPre, _, err := normalizedFullSquareStep(params, coherentCurrent, big.NewInt(2), sqrt2pi)
		if err != nil {
			return path, err
		}
		coherentCurrent, err = normalizedFullRescale(params, coherentPre)
		if err != nil {
			return path, err
		}
		if err := standardEval.Evaluator.MulRelin(standardPolyCurrent, standardPolyCurrent, standardPolyCurrent); err != nil {
			return path, err
		}
		if err := standardEval.Evaluator.Add(standardPolyCurrent, standardPolyCurrent, standardPolyCurrent); err != nil {
			return path, err
		}
		if err := standardEval.Evaluator.Add(standardPolyCurrent, -sqrt2pi, standardPolyCurrent); err != nil {
			return path, err
		}
		if err := standardEval.Evaluator.Rescale(standardPolyCurrent, standardPolyCurrent); err != nil {
			return path, err
		}
		fastAfterRescale := fastNormalized.CopyNew()
		postCheckpoint, err := evalModMatchedCheckpointFor("da_post_rescale", round, fastAfterRescale, normalizedCurrent, coherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return path, err
		}
		if !postCheckpoint.RowsMatch && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.post_rescale_rows", round)
		}
		if !postCheckpoint.FastVsNormalized.Pass && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.post_rescale_semantic", round)
		}
		if !postCheckpoint.NormalizedSafe && path.FirstFailure == "none" {
			path.FirstFailure = fmt.Sprintf("round%d.post_rescale_capacity", round)
			path.Rounds = append(path.Rounds, evalModMatchedRound{Round: round, KInExponent: kIn, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint, AfterConstant: constantCheckpoint, PostRescale: postCheckpoint})
			return path, nil
		}
		_, nValues, err := semanticBisectView(normalizedCurrent, params, zeroSecret(params))
		if err != nil {
			return path, err
		}
		_, cValues, err := semanticBisectView(coherentCurrent, params, zeroSecret(params))
		if err != nil {
			return path, err
		}
		mappedN := evalModMatchedScaleValues(nValues, new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))
		nVsC := evalModMatchedMetricFor(cValues, mappedN, evalModMatchedLocalThreshold)
		_, sValues, err := semanticBisectView(standardPolyCurrent, params, standardSK)
		if err != nil {
			return path, err
		}
		cVsS := evalModMatchedMetricFor(sValues, cValues, evalModMatchedLocalThreshold)
		path.Rounds = append(path.Rounds, evalModMatchedRound{Round: round, KInExponent: kIn, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, NormalizedConstant: constantValue, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint, AfterConstant: constantCheckpoint, PostRescale: postCheckpoint, NVsCPostRescale: nVsC, CVsSPostRescale: cVsS})
		fastNormalized = fastAfterRescale
		currentLevel = schedule.NextLevel
		currentScale = scheduleScaleValue(schedule, params, currentLevel, currentScale)
		kIn = schedule.KOutExponent
	}
	fastBeforeRestore := fastNormalized.CopyNew()
	normalizedBeforeRestore := normalizedCurrent.CopyNew()
	coherentBeforeReset := coherentCurrent.CopyNew()
	standardBeforeReset := standardPolyCurrent.CopyNew()
	finalK := new(big.Int).Lsh(big.NewInt(1), uint(kIn))
	beforeRestoreCapacity, err := normalizedExactCapacity(params, normalizedCurrent)
	if err != nil {
		return path, err
	}
	restoreView, err := targetScaleCoefficientView(params, fastBeforeRestore)
	if err != nil {
		return path, err
	}
	restoreData, err := targetScaleCenteredRows(params, restoreView)
	if err != nil {
		return path, err
	}
	restoreCapacity := normalizedFinalRestoreCapacity(restoreData, finalK)
	path.NormalizedBeforeRestoreCapacity = &beforeRestoreCapacity
	path.FinalRestoreCapacity = &restoreCapacity
	if !beforeRestoreCapacity.Pass || !restoreCapacity.Pass {
		path.FirstFailure = "final.restore_capacity"
		return path, nil
	}
	if err := fastEval.FastCKKS.MulIntegerMaintained(fastNormalized, finalK, fastNormalized); err != nil {
		return path, err
	}
	normalizedCurrent = normalizedFullIntegerMultiply(params, normalizedCurrent, finalK)
	fastRestoreRows := finalizationEvidence(fastNormalized).Rows
	normalizedRestoreRows := finalizationEvidence(normalizedCurrent).Rows
	path.FastRowsMatch = doubleAngleRowsMatch(fastRestoreRows, normalizedRestoreRows)
	path.RestoreRowsMatch = path.FastRowsMatch
	fastNormalized.Scale = originalFastScale
	normalizedCurrent.Scale = originalFastScale
	standardPolyCurrent.Scale = originalStandardScale
	coherentCurrent.Scale = originalStandardScale
	path.ScaleResetRowsUnchanged = doubleAngleRowsMatch(fastRestoreRows, finalizationEvidence(fastNormalized).Rows) && doubleAngleRowsMatch(normalizedRestoreRows, finalizationEvidence(normalizedCurrent).Rows)
	path.LevelDegreeUnchanged = fastNormalized.Level() == fastBeforeRestore.Level() && fastNormalized.Degree() == fastBeforeRestore.Degree() && normalizedCurrent.Level() == normalizedBeforeRestore.Level() && normalizedCurrent.Degree() == normalizedBeforeRestore.Degree()
	path.FastFinal, path.NormalizedFinal, path.CoherentFinal, path.StandardFinal = fastNormalized, normalizedCurrent, coherentCurrent, standardPolyCurrent
	_ = standardBeforeReset
	_ = coherentBeforeReset
	_ = originalStandardScale
	_ = workingScale
	return path, nil
}

func float64FromBigFloat(value *big.Float) float64 {
	result, _ := value.Float64()
	return result
}

func evalModMatchedFinalEvidence(path evalModMatchedPath, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey) (evalModMatchedFinal, error) {
	_, fastValues, err := semanticBisectView(path.FastFinal, params, fastSK)
	if err != nil {
		return evalModMatchedFinal{}, err
	}
	_, normalizedValues, err := semanticBisectView(path.NormalizedFinal, params, zeroSecret(params))
	if err != nil {
		return evalModMatchedFinal{}, err
	}
	_, coherentValues, err := semanticBisectView(path.CoherentFinal, params, zeroSecret(params))
	if err != nil {
		return evalModMatchedFinal{}, err
	}
	_, standardValues, err := semanticBisectView(path.StandardFinal, params, standardSK)
	if err != nil {
		return evalModMatchedFinal{}, err
	}
	return evalModMatchedFinal{
		FastVsNormalized:                evalModMatchedMetricFor(normalizedValues, fastValues, evalModMatchedLocalThreshold),
		NormalizedVsCoherent:            evalModMatchedMetricFor(coherentValues, normalizedValues, evalModMatchedLocalThreshold),
		CoherentVsStandard:              evalModMatchedMetricFor(standardValues, coherentValues, evalModMatchedLocalThreshold),
		FastVsStandard:                  evalModMatchedMetricFor(standardValues, fastValues, evalModMatchedLocalThreshold),
		NormalizedVsPlainOracle:         evalModMatchedMetricFor(path.PlainFinal, normalizedValues, evalModMatchedLocalThreshold),
		StandardVsPlainOracle:           evalModMatchedMetricFor(path.PlainFinal, standardValues, evalModMatchedLocalThreshold),
		NormalizedBeforeRestoreCapacity: path.NormalizedBeforeRestoreCapacity,
		FinalRestoreCapacity:            path.FinalRestoreCapacity,
		FastRowsMatchNormalized:         path.FastRowsMatch, RestoreRowsMatch: path.RestoreRowsMatch, ScaleResetRowsUnchanged: path.ScaleResetRowsUnchanged, LevelDegreeUnchanged: path.LevelDegreeUnchanged,
	}, nil
}

func evalModMatchedClassification(real evalModMatchedPath, realFinal evalModMatchedFinal) (string, string) {
	if real.FirstFailure != "none" {
		if strings.Contains(real.FirstFailure, "capacity") {
			return "logn13_evalmod_normalized_reference_capacity_failure", real.FirstFailure
		}
		if len(real.Prefix) > 0 && real.Prefix[len(real.Prefix)-1].Name == real.FirstFailure {
			return "logn13_evalmod_fast_prefix_primitive_mismatch", real.FirstFailure
		}
		if real.Polynomial["fast_vs_normalized"] != nil && !real.Polynomial["fast_vs_normalized"].Pass {
			return "logn13_evalmod_fast_polynomial_primitive_mismatch", real.FirstFailure
		}
		if real.Polynomial["normalized_vs_plaintext_oracle"] != nil && !real.Polynomial["normalized_vs_plaintext_oracle"].Pass {
			return "logn13_evalmod_normalized_reference_capacity_failure", real.FirstFailure
		}
		return "logn13_evalmod_fast_normalized_primitive_mismatch", real.FirstFailure
	}
	if realFinal.FastVsStandard.Pass {
		return "logn13_evalmod_matched_input_divergence_not_reproduced_after_c2s_correction", "none"
	}
	if !realFinal.NormalizedVsCoherent.Pass {
		return "logn13_evalmod_normalized_recurrence_semantic_divergence", "final.normalized_vs_coherent"
	}
	if !realFinal.ScaleResetRowsUnchanged {
		return "logn13_evalmod_final_scale_reset_mismatch", "final.scale_reset"
	}
	if !realFinal.RestoreRowsMatch {
		return "logn13_evalmod_final_restore_mismatch", "final.restore"
	}
	if !realFinal.CoherentVsStandard.Pass {
		return "logn13_evalmod_compressed_polynomial_semantic_divergence", "final.coherent_vs_standard"
	}
	return "logn13_evalmod_semantic_cause_not_yet_isolated", "final.fast_vs_standard"
}

func evalModMatchedWrite(result evalModMatchedResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DiagLogN13EvalModMatchedNormalizedOracle(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := evalModMatchedResult{SchemaVersion: "fix-001-p3-diag-logn13-evalmod-matched-normalized-oracle.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, StandardProof: semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, OrdinaryStageCallsUsed: true}, PlaintextOracle: "raw_chebyshev_on_preprocessed_z", Validation: map[string]interface{}{"primary_required_base": requiredEvalModMatchedPrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredEvalModMatchedSecondary, "no_secondary_production_changes": true, "no_production_fix": true, "no_c2s_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_s2c_packing_unpacking_finalization": true, "raw_standard_da_q01_projection_disposition": "invalid_after_centered_capacity_loss", "raw_standard_da_q01_projection_avoided": true}}
	result.FirstFailing = "none"
	if !secondaryClean || secondaryCommit != requiredEvalModMatchedSecondary || !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil {
		result.Classification = "logn13_evalmod_matched_input_precondition_failure"
		result.FirstFailing = "startup_precondition"
		return evalModMatchedWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, skN2)
	if err != nil {
		return err
	}
	params := btp.BootstrappingParameters
	fastRealMeta, _, err := semanticBisectView(inputs.FastReal, params, inputs.FastSK)
	if err != nil {
		return err
	}
	alignedMeta, alignedValues, err := semanticBisectView(inputs.AlignedReal, params, inputs.StandardSK)
	if err != nil {
		return err
	}
	ordinaryMeta, ordinaryValues, err := semanticBisectView(inputs.OrdinaryReal, params, inputs.StandardSK)
	if err != nil {
		return err
	}
	_, fastValues, err := semanticBisectView(inputs.FastReal, params, inputs.FastSK)
	if err != nil {
		return err
	}
	inputC2S := evalModMatchedC2S{Level: inputs.FastReal.Level(), Degree: inputs.FastReal.Degree(), Scale: finalizationScaleString(inputs.FastReal.Scale), Fast: fastRealMeta, AlignedStandard: alignedMeta, OrdinaryStandard: ordinaryMeta, FastVsAlignedStandard: inputs.FastVsAlignedReal, FastVsOrdinaryStandard: inputs.FastVsOrdinaryReal, FastVsAlignedImag: inputs.FastVsAlignedImag, FastVsOrdinaryImag: inputs.FastVsOrdinaryImag, AlignedVsOrdinary: evalModMatchedMetricFor(ordinaryValues, alignedValues, acceptedC2SGroup1Budget), FastQ01CapacityPass: inputs.FastQ01Safe, FastImagQ01CapacityPass: inputs.FastImagQ01Safe, AlignedFullRNSCapacity: inputs.AlignedCapacitySafe, NoFurtherDFT: true, AcceptedGroup0K4: true, AcceptedGroup1K2: true, Groups2And3Once: true, C2SPrecisionBlockerFixed: inputs.FastVsAlignedReal.Pass && inputs.FastVsAlignedImag.Pass && inputs.FastQ01Safe && inputs.FastImagQ01Safe}
	result.CorrectedC2S = inputC2S
	if !inputC2S.C2SPrecisionBlockerFixed {
		result.Classification = "logn13_evalmod_matched_input_precondition_failure"
		result.FirstFailing = "corrected_c2s"
		return evalModMatchedWrite(result, outPath)
	}
	realPath, err := evalModMatchedRunPath(inputs.FastReal, inputs.OrdinaryReal, fastEval, standardEval, params, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return err
	}
	if realPath.FirstFailure != "none" {
		result.Classification, result.FirstFailing = evalModMatchedClassification(realPath, evalModMatchedFinal{})
		return evalModMatchedWrite(result, outPath)
	}
	imagPath, err := evalModMatchedRunPath(inputs.FastImag, inputs.OrdinaryImag, fastEval, standardEval, params, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return err
	}
	result.Prefix = realPath.Prefix
	result.Polynomial = realPath.Polynomial
	result.PolynomialMeta = realPath.PolyMeta
	result.Rounds = realPath.Rounds
	result.TargetScale = realPath.PolyMeta["target_scale"]
	result.PlanScale = realPath.PolyMeta["plan_scale"]
	result.CompressedPolyScale = realPath.PolyMeta["fast_scale"]
	result.NormalizedPolyCapacity = realPath.PolyCapacity
	realFinal, err := evalModMatchedFinalEvidence(realPath, params, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, params, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return err
	}
	result.FinalReal, result.FinalImag = realFinal, imagFinal
	result.CausalEffects = evalModMatchedCausalEffects{FastImplementationEffect: realFinal.FastVsNormalized, NormalizedRecurrenceEffect: realFinal.NormalizedVsCoherent, CompressedPolynomialEffect: realFinal.CoherentVsStandard, TotalFastVsStandard: realFinal.FastVsStandard, Identity: "F - S = (F - N) + (N - C) + (C - S); max-absolute metrics are recorded separately and are not algebraically additive"}
	classification, first := evalModMatchedClassification(realPath, realFinal)
	result.Classification, result.FirstFailing = classification, first
	result.Validation["real_final_pass"] = realFinal.FastVsStandard.Pass
	result.Validation["imag_final_pass"] = imagFinal.FastVsStandard.Pass
	result.Validation["c2s_precision_blocker_resolved_for_logn13"] = inputC2S.C2SPrecisionBlockerFixed
	result.Validation["normalized_reference_capacity_failure"] = realPath.FirstFailure == "polynomial.normalized_capacity"
	result.Validation["no_invalid_standard_double_angle_oracle"] = true
	_ = fastValues
	return evalModMatchedWrite(result, outPath)
}
