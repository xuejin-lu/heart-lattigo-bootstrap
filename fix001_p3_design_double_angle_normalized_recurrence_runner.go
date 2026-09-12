package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type NormalizedScaleSchedule struct {
	Round             int      `json:"round"`
	LevelIn           int      `json:"level_in"`
	ScaleIn           string   `json:"scale_in"`
	WorkingScale      string   `json:"working_scale_w"`
	KIn               string   `json:"k_in"`
	KInExponent       int      `json:"k_in_exponent"`
	DroppedModuli     []uint64 `json:"dropped_moduli"`
	ScaleOut          string   `json:"scale_out"`
	KOut              string   `json:"k_out"`
	KOutExponent      int      `json:"k_out_exponent"`
	A                 string   `json:"a"`
	AExponent         int      `json:"a_exponent"`
	Constant          string   `json:"c_i_over_k_out"`
	EffectiveScaleIn  string   `json:"effective_scale_in"`
	EffectiveScaleOut string   `json:"effective_scale_out"`
	Log2DeviationIn   float64  `json:"log2_deviation_in"`
	Log2DeviationOut  float64  `json:"log2_deviation_out"`
	NextLevel         int      `json:"next_level"`
}

type NormalizedSquareBound struct {
	N         int     `json:"n"`
	B         string  `json:"max_abs_c0_coefficient"`
	NSquaredB string  `json:"n_times_b_squared"`
	Q01Half   string  `json:"q01_half"`
	Ratio     float64 `json:"bound_over_q01_half"`
	Pass      bool    `json:"pass"`
}

type NormalizedExactCapacity struct {
	MaxAbs       string  `json:"max_abs_exact_c0_coefficient"`
	Q01Half      string  `json:"q01_half"`
	Ratio        float64 `json:"max_abs_over_q01_half"`
	OutsideCount int     `json:"outside_count"`
	Pass         bool    `json:"pass_exact_q01_capacity"`
}

type NormalizedFinalRestoreCapacity struct {
	B            string  `json:"max_abs_c0_coefficient"`
	K3           string  `json:"k3"`
	RestoreBound string  `json:"k3_times_b"`
	Q01Half      string  `json:"q01_half"`
	Ratio        float64 `json:"k3_times_b_over_q01_half"`
	OutsideCount int     `json:"outside_count"`
	Pass         bool    `json:"pass_final_restore_capacity"`
}

type NormalizedStageComparison struct {
	Match               bool   `json:"match"`
	FastLevel           int    `json:"fast_level"`
	ReferenceLevel      int    `json:"reference_level"`
	FastScale           string `json:"fast_scale"`
	ReferenceScale      string `json:"reference_scale"`
	FastIsNTT           bool   `json:"fast_is_ntt"`
	ReferenceIsNTT      bool   `json:"reference_is_ntt"`
	FastIsMontgomery    bool   `json:"fast_is_montgomery"`
	ReferenceMontgomery bool   `json:"reference_is_montgomery"`
	HasMismatch         bool   `json:"has_mismatch"`
	Component           int    `json:"component,omitempty"`
	Limb                int    `json:"limb,omitempty"`
	Index               int    `json:"index,omitempty"`
	FastValue           uint64 `json:"fast_value,omitempty"`
	ReferenceValue      uint64 `json:"reference_value,omitempty"`
}

type NormalizedRoundCheckpoint struct {
	Round                     int                          `json:"round"`
	Schedule                  NormalizedScaleSchedule      `json:"schedule"`
	InputRowsMatchReference   bool                         `json:"input_rows_match_reference"`
	InputComparison           NormalizedStageComparison    `json:"input_comparison"`
	InputSemantic             *PSGlobalMetric              `json:"input_normalized_semantic"`
	InputSquareBound          NormalizedSquareBound        `json:"input_square_bound"`
	SquareSemantic            *PSGlobalMetric              `json:"square_semantic"`
	SquareComparison          NormalizedStageComparison    `json:"square_comparison"`
	AfterMultiplierSemantic   *PSGlobalMetric              `json:"after_a_semantic"`
	AfterMultiplierComparison NormalizedStageComparison    `json:"after_a_comparison"`
	Constant                  string                       `json:"normalized_constant"`
	AfterConstantSemantic     *PSGlobalMetric              `json:"after_constant_semantic"`
	AfterConstantComparison   NormalizedStageComparison    `json:"after_constant_comparison"`
	PreRescaleCapacity        NormalizedExactCapacity      `json:"pre_rescale_exact_reference_capacity"`
	SquareRowsMatchReference  bool                         `json:"square_rows_match_reference"`
	AfterMultiplierRowsMatch  bool                         `json:"after_multiplier_rows_match_reference"`
	AfterConstantRowsMatch    bool                         `json:"after_constant_rows_match_reference"`
	PreRescaleRowsMatch       bool                         `json:"pre_rescale_rows_match_reference"`
	PreRescaleFastRows        []FinalizationRowFingerprint `json:"pre_rescale_fast_rows"`
	PreRescaleReferenceRows   []FinalizationRowFingerprint `json:"pre_rescale_reference_rows"`
	PostRescaleSemantic       *PSGlobalMetric              `json:"post_rescale_semantic"`
	PostRescaleComparison     NormalizedStageComparison    `json:"post_rescale_comparison"`
	PostRescaleRowsMatch      bool                         `json:"post_rescale_rows_match_reference"`
	PostRescaleScaleMatch     bool                         `json:"post_rescale_scale_match"`
	PostRescaleEffectiveScale string                       `json:"post_rescale_effective_scale"`
	PostRescaleFastRows       []FinalizationRowFingerprint `json:"post_rescale_fast_rows"`
	PostRescaleReferenceRows  []FinalizationRowFingerprint `json:"post_rescale_reference_rows"`
}

type NormalizedFinalEvidence struct {
	K3                                      string                         `json:"k3"`
	BeforeResetScale                        string                         `json:"before_reset_scale"`
	RestoreCapacity                         NormalizedFinalRestoreCapacity `json:"final_restore_capacity"`
	FastBeforeRestoreRows                   []FinalizationRowFingerprint   `json:"fast_before_restore_rows"`
	RNormBeforeRestoreRows                  []FinalizationRowFingerprint   `json:"r_norm_before_restore_rows"`
	FastRestoredRows                        []FinalizationRowFingerprint   `json:"fast_restored_rows"`
	RNormRestoredRows                       []FinalizationRowFingerprint   `json:"r_norm_restored_rows"`
	NormalizedRestoredRowsMatch             bool                           `json:"normalized_restored_rows_match"`
	NormalizedRestoredComparison            NormalizedStageComparison      `json:"normalized_restored_comparison"`
	FastRestoreMetadataPass                 bool                           `json:"fast_restore_metadata_pass"`
	RNormRestoreMetadataPass                bool                           `json:"r_norm_restore_metadata_pass"`
	FastRestoreC1Zero                       bool                           `json:"fast_restore_c1_zero"`
	RNormRestoreC1Zero                      bool                           `json:"r_norm_restore_c1_zero"`
	FastVsNormalizedRestoredSemantic        *PSGlobalMetric                `json:"fast_vs_normalized_restored_semantic"`
	NormalizedRestoredVsRCoherentSemantic   *PSGlobalMetric                `json:"normalized_restored_vs_r_coherent_semantic"`
	NormalizedRestoredVsExactTargetSemantic *PSGlobalMetric                `json:"normalized_restored_vs_exact_target_semantic"`
	FastVsRCoherentSemantic                 *PSGlobalMetric                `json:"fast_vs_r_coherent_semantic"`
	RCoherentVsExactTarget                  *PSGlobalMetric                `json:"r_coherent_vs_r_exact_target_semantic"`
	FastVsExactTargetSemantic               *PSGlobalMetric                `json:"fast_vs_r_exact_target_semantic"`
	FastFinalRowsAfterReset                 []FinalizationRowFingerprint   `json:"fast_final_rows_after_reset"`
	NormalizedReferenceFinalRowsAfterReset  []FinalizationRowFingerprint   `json:"normalized_reference_final_rows_after_reset"`
	RowsMatchAfterReset                     bool                           `json:"rows_match_after_reset"`
	FastRowsUnchangedOnReset                bool                           `json:"fast_rows_unchanged_on_metadata_reset"`
	NormalizedReferenceRowsUnchangedOnReset bool                           `json:"normalized_reference_rows_unchanged_on_metadata_reset"`
	FastFinalLevelDegreeUnchanged           bool                           `json:"fast_final_level_degree_unchanged"`
	NormalizedFinalLevelDegreeUnchanged     bool                           `json:"normalized_final_level_degree_unchanged"`
	FinalScaleResetSemantic                 *PSGlobalMetric                `json:"final_scale_reset_semantic"`
	NormalizedFinalScaleResetSemantic       *PSGlobalMetric                `json:"normalized_final_scale_reset_semantic"`
	FinalScaleResetPass                     bool                           `json:"final_scale_reset_pass"`
	FinalSemanticPass                       bool                           `json:"final_semantic_pass"`
	ExactTargetCompatibility                bool                           `json:"exact_target_compatibility_pass"`
	RCoherentRows                           []FinalizationRowFingerprint   `json:"r_coherent_rows_semantic_only"`
	RCoherentFinalRows                      []FinalizationRowFingerprint   `json:"r_coherent_final_rows_semantic_only"`
	OldRCoherentRowEqualityDisposition      string                         `json:"old_r_coherent_row_equality_disposition"`
}

type NormalizedResult struct {
	SchemaVersion                     string                      `json:"schema_version"`
	Timestamp                         time.Time                   `json:"timestamp"`
	Primary                           RepositoryMetadata          `json:"primary_repository"`
	Lattigo                           RepositoryMetadata          `json:"lattigo_repository"`
	Environment                       EnvironmentMetadata         `json:"environment"`
	Config                            BootstrapConfig             `json:"config"`
	Parameters                        ExperimentParameters        `json:"effective_parameters"`
	Workload                          CorrectnessWorkload         `json:"workload"`
	AuthoritativeOracle               string                      `json:"authoritative_poly_oracle"`
	Threshold                         float64                     `json:"threshold"`
	WorkingScaleW                     string                      `json:"working_scale_w"`
	CoherentInitialScale              string                      `json:"coherent_initial_scale"`
	ExactTargetScale                  string                      `json:"exact_standard_target_scale"`
	InitialScaleLog2Delta             float64                     `json:"initial_coherent_vs_exact_target_log2_delta"`
	PromotionMultiplier               string                      `json:"promotion_multiplier"`
	CanonicalLow                      NormalizedStateEvidence     `json:"canonical_low"`
	CanonicalNormalizedInput          *PSGlobalMetric             `json:"canonical_normalized_input_semantic"`
	Schedules                         []NormalizedScaleSchedule   `json:"schedules"`
	Rounds                            []NormalizedRoundCheckpoint `json:"rounds"`
	Final                             NormalizedFinalEvidence     `json:"final"`
	DesignMechanism                   string                      `json:"design_mechanism"`
	PreviousClassificationDisposition string                      `json:"previous_classification_disposition"`
	FirstFailingRound                 string                      `json:"first_failing_round"`
	FirstFailingCheckpoint            string                      `json:"first_failing_checkpoint"`
	FirstSupportedCause               string                      `json:"first_supported_cause"`
	Validation                        map[string]interface{}      `json:"validation"`
}

type NormalizedStateEvidence struct {
	Level    int                          `json:"level"`
	Degree   int                          `json:"degree"`
	Scale    string                       `json:"scale"`
	Semantic *PSGlobalMetric              `json:"semantic_vs_y0"`
	Rows     []FinalizationRowFingerprint `json:"q0_q1_rows"`
	C1Zero   bool                         `json:"maintained_c1_exactly_zero"`
	C0MaxAbs string                       `json:"c0_max_abs_centered"`
}

func normalizedLog2Deviation(actual, target rlwe.Scale) float64 {
	delta := actual.Log2() - target.Log2()
	if math.IsNaN(delta) || math.IsInf(delta, 0) {
		return math.MaxFloat64
	}
	return math.Abs(delta)
}

func normalizedPowerOfTwoExponent(scale, working rlwe.Scale) (int, error) {
	ratio := new(big.Float).Quo(new(big.Float).SetPrec(512).Set(&scale.Value), new(big.Float).SetPrec(512).Set(&working.Value))
	ratioFloat, _ := ratio.Float64()
	if !(ratioFloat > 0) || math.IsInf(ratioFloat, 0) || math.IsNaN(ratioFloat) {
		return 0, fmt.Errorf("invalid scale ratio for power-of-two schedule: %s", ratio.Text('e', 40))
	}
	return int(math.Round(math.Log2(ratioFloat))), nil
}

func normalizedScaleSchedule(round, level int, scale, working rlwe.Scale, kIn int, c float64, params ckks.Parameters) (NormalizedScaleSchedule, error) {
	nextScale, nextLevel, dropped := doubleAngleDroppedScale(scale.Mul(scale), level, params)
	kOut, err := normalizedPowerOfTwoExponent(nextScale, working)
	if err != nil {
		return NormalizedScaleSchedule{}, err
	}
	if kIn < 0 || kOut < 0 {
		return NormalizedScaleSchedule{}, fmt.Errorf("negative normalized scale exponent: kIn=%d kOut=%d", kIn, kOut)
	}
	aExponent := 1 + 2*kIn - kOut
	if aExponent < 0 {
		return NormalizedScaleSchedule{}, fmt.Errorf("normalized recurrence multiplier is not an integer power of two: a=%d", aExponent)
	}
	kInValue := new(big.Int).Lsh(big.NewInt(1), uint(kIn))
	kOutValue := new(big.Int).Lsh(big.NewInt(1), uint(kOut))
	aValue := new(big.Int).Lsh(big.NewInt(1), uint(aExponent))
	constant := new(big.Float).Quo(new(big.Float).SetPrec(512).SetFloat64(c), new(big.Float).SetPrec(512).SetInt(kOutValue))
	effectiveIn := scale.Div(rlwe.NewScale(kInValue))
	effectiveOut := nextScale.Div(rlwe.NewScale(kOutValue))
	return NormalizedScaleSchedule{
		Round: round, LevelIn: level, ScaleIn: finalizationScaleString(scale), WorkingScale: finalizationScaleString(working),
		KIn: kInValue.String(), KInExponent: kIn, DroppedModuli: dropped, ScaleOut: finalizationScaleString(nextScale),
		KOut: kOutValue.String(), KOutExponent: kOut, A: aValue.String(), AExponent: aExponent,
		Constant: constant.Text('e', 80), EffectiveScaleIn: finalizationScaleString(effectiveIn), EffectiveScaleOut: finalizationScaleString(effectiveOut),
		Log2DeviationIn: normalizedLog2Deviation(effectiveIn, working), Log2DeviationOut: normalizedLog2Deviation(effectiveOut, working), NextLevel: nextLevel,
	}, nil
}

func normalizedSquareBound(data targetScaleCenteredData, n int) NormalizedSquareBound {
	max := mulAliasMaxAbs(data.Values[0])
	product := new(big.Int).Mul(big.NewInt(int64(n)), new(big.Int).Mul(max, max))
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(product), new(big.Float).SetInt(data.Q01Half)).Float64()
	return NormalizedSquareBound{N: n, B: max.String(), NSquaredB: product.String(), Q01Half: data.Q01Half.String(), Ratio: ratio, Pass: product.Cmp(data.Q01Half) < 0}
}

func normalizedFullCRTCentered(residues []uint64, moduli []uint64) *big.Int {
	modulus := big.NewInt(1)
	for _, qi := range moduli {
		modulus.Mul(modulus, new(big.Int).SetUint64(qi))
	}
	value := new(big.Int)
	for i, qi := range moduli {
		mi := new(big.Int).Quo(new(big.Int).Set(modulus), new(big.Int).SetUint64(qi))
		inverse := new(big.Int).ModInverse(mi, new(big.Int).SetUint64(qi))
		term := new(big.Int).Mul(new(big.Int).SetUint64(residues[i]), mi)
		term.Mul(term, inverse)
		value.Add(value, term)
	}
	value.Mod(value, modulus)
	half := new(big.Int).Quo(new(big.Int).Set(modulus), big.NewInt(2))
	if value.Cmp(half) >= 0 {
		value.Sub(value, modulus)
	}
	return value
}

func normalizedFullCoefficientValues(params ckks.Parameters, ct *rlwe.Ciphertext) ([][]*big.Int, error) {
	if ct == nil || !ct.IsNTT {
		return nil, fmt.Errorf("full-RNS coefficient view requires an NTT ciphertext")
	}
	ringQ := params.RingQ().AtLevel(ct.Level())
	source := ct.CopyNew()
	if source.IsMontgomery {
		for component := range source.Value {
			ringQ.IMForm(source.Value[component], source.Value[component])
		}
	}
	moduli := make([]uint64, ct.Level()+1)
	for i := range moduli {
		moduli[i] = ringQ.SubRings[i].Modulus
	}
	values := make([][]*big.Int, 0, ct.Degree()+1)
	for component := 0; component <= ct.Degree(); component++ {
		coeff := ring.NewPoly(ringQ.N(), ct.Level())
		ringQ.INTT(source.Value[component], coeff)
		componentValues := make([]*big.Int, ringQ.N())
		for index := range componentValues {
			residues := make([]uint64, len(moduli))
			for limb := range moduli {
				residues[limb] = coeff.Coeffs[limb][index]
			}
			componentValues[index] = normalizedFullCRTCentered(residues, moduli)
		}
		values = append(values, componentValues)
	}
	return values, nil
}

func normalizedExactCapacity(params ckks.Parameters, ct *rlwe.Ciphertext) (NormalizedExactCapacity, error) {
	values, err := normalizedFullCoefficientValues(params, ct)
	if err != nil {
		return NormalizedExactCapacity{}, err
	}
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	q01Half := new(big.Int).Quo(new(big.Int).Mul(q0, q1), big.NewInt(2))
	max := new(big.Int)
	outside := 0
	for _, value := range values[0] {
		absolute := new(big.Int).Abs(value)
		if absolute.Cmp(max) > 0 {
			max.Set(absolute)
		}
		if absolute.Cmp(q01Half) >= 0 {
			outside++
		}
	}
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(max), new(big.Float).SetInt(q01Half)).Float64()
	return NormalizedExactCapacity{MaxAbs: max.String(), Q01Half: q01Half.String(), Ratio: ratio, OutsideCount: outside, Pass: outside == 0}, nil
}

func normalizedFinalRestoreCapacity(data targetScaleCenteredData, k3 *big.Int) NormalizedFinalRestoreCapacity {
	b := mulAliasMaxAbs(data.Values[0])
	bound := new(big.Int).Mul(new(big.Int).Set(b), k3)
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(bound), new(big.Float).SetInt(data.Q01Half)).Float64()
	outside := 0
	if bound.Cmp(data.Q01Half) >= 0 {
		outside = 1
	}
	return NormalizedFinalRestoreCapacity{B: b.String(), K3: k3.String(), RestoreBound: bound.String(), Q01Half: data.Q01Half.String(), Ratio: ratio, OutsideCount: outside, Pass: outside == 0}
}

func normalizedFullFromLow(params ckks.Parameters, low *rlwe.Ciphertext, data targetScaleCenteredData) (*rlwe.Ciphertext, error) {
	level := low.Level()
	ringQ := params.RingQ().AtLevel(level)
	full := ckks.NewCiphertext(params, 1, level)
	*full.MetaData = *low.MetaData
	full.IsNTT = low.IsNTT
	full.IsMontgomery = low.IsMontgomery
	if !full.IsNTT {
		return nil, fmt.Errorf("normalized full reference requires an NTT low state")
	}
	coeff0 := ring.NewPoly(ringQ.N(), level)
	for index, value := range data.Values[0] {
		for limb := 0; limb <= level; limb++ {
			coeff0.Coeffs[limb][index] = new(big.Int).Mod(new(big.Int).Set(value), new(big.Int).SetUint64(ringQ.SubRings[limb].Modulus)).Uint64()
		}
	}
	ringQ.NTT(coeff0, full.Value[0])
	if full.IsMontgomery {
		ringQ.MForm(full.Value[0], full.Value[0])
	}
	return full, nil
}

func normalizedFullZeroC2(params ckks.Parameters, product *rlwe.Ciphertext) *rlwe.Ciphertext {
	result := ckks.NewCiphertext(params, 1, product.Level())
	*result.MetaData = *product.MetaData
	result.IsNTT, result.IsMontgomery = product.IsNTT, product.IsMontgomery
	for component := 0; component < 2; component++ {
		copy(result.Value[component].Coeffs[0], product.Value[component].Coeffs[0])
		for limb := 1; limb <= product.Level(); limb++ {
			copy(result.Value[component].Coeffs[limb], product.Value[component].Coeffs[limb])
		}
	}
	return result
}

func normalizedFullIntegerMultiply(params ckks.Parameters, input *rlwe.Ciphertext, factor *big.Int) *rlwe.Ciphertext {
	result := ckks.NewCiphertext(params, input.Degree(), input.Level())
	*result.MetaData = *input.MetaData
	result.IsNTT, result.IsMontgomery = input.IsNTT, input.IsMontgomery
	ringQ := params.RingQ().AtLevel(input.Level())
	for component := range input.Value {
		if !input.IsMontgomery {
			ringQ.MulScalar(input.Value[component], factor.Uint64(), result.Value[component])
			continue
		}
		for limb := 0; limb <= input.Level(); limb++ {
			subring := ringQ.SubRings[limb]
			scalar := new(big.Int).Mod(new(big.Int).Set(factor), new(big.Int).SetUint64(subring.Modulus)).Uint64()
			subring.MulScalarMontgomery(input.Value[component].Coeffs[limb], ring.MForm(scalar, subring.Modulus, subring.BRedConstant), result.Value[component].Coeffs[limb])
		}
	}
	return result
}

func normalizedFullSubtractConstant(params ckks.Parameters, input *rlwe.Ciphertext, constant float64) *rlwe.Ciphertext {
	result := input.CopyNew()
	constantScaled := new(big.Float).Mul(new(big.Float).SetPrec(512).SetFloat64(constant), new(big.Float).SetPrec(512).Set(&input.Scale.Value))
	if constantScaled.Sign() > 0 {
		constantScaled.Add(constantScaled, new(big.Float).SetFloat64(0.5))
	} else if constantScaled.Sign() < 0 {
		constantScaled.Sub(constantScaled, new(big.Float).SetFloat64(0.5))
	}
	constantInteger := new(big.Int)
	constantScaled.Int(constantInteger)
	ringQ := params.RingQ().AtLevel(input.Level())
	for limb := 0; limb <= input.Level(); limb++ {
		subring := ringQ.SubRings[limb]
		scalar := new(big.Int).Mod(new(big.Int).Set(constantInteger), new(big.Int).SetUint64(subring.Modulus)).Uint64()
		if input.IsMontgomery {
			scalar = ring.MForm(scalar, subring.Modulus, subring.BRedConstant)
		}
		subring.SubScalar(result.Value[0].Coeffs[limb], scalar, result.Value[0].Coeffs[limb])
	}
	return result
}

func normalizedFullSquare(params ckks.Parameters, input *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if input == nil || !input.IsNTT {
		return nil, fmt.Errorf("normalized full reference requires an NTT input")
	}
	level := input.Level()
	ringQ := params.RingQ().AtLevel(level)
	product := ckks.NewCiphertext(params, 2, level)
	*product.MetaData = *input.MetaData
	product.IsNTT, product.IsMontgomery = input.IsNTT, input.IsMontgomery
	product.Scale = input.Scale.Mul(input.Scale)
	if input.IsMontgomery {
		ringQ.MulCoeffsMontgomery(input.Value[0], input.Value[0], product.Value[0])
	} else {
		ringQ.MulCoeffsBarrett(input.Value[0], input.Value[0], product.Value[0])
	}
	return product, nil
}

func normalizedFullSquareStep(params ckks.Parameters, input *rlwe.Ciphertext, factor *big.Int, constant float64) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	product, err := normalizedFullSquare(params, input)
	if err != nil {
		return nil, nil, err
	}
	linear := normalizedFullZeroC2(params, product)
	factored := normalizedFullIntegerMultiply(params, linear, factor)
	preRescale := normalizedFullSubtractConstant(params, factored, constant)
	return preRescale, product, nil
}

func normalizedFullRescale(params ckks.Parameters, preRescale *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if preRescale == nil || !preRescale.IsNTT {
		return nil, fmt.Errorf("normalized full rescale requires an NTT ciphertext")
	}
	nbRescales := params.LevelsConsumedPerRescaling()
	if preRescale.Level() < nbRescales {
		return nil, fmt.Errorf("normalized full rescale has insufficient level")
	}
	oldLevel := preRescale.Level()
	targetLevel := oldLevel - nbRescales
	ringOld := params.RingQ().AtLevel(oldLevel)
	source := preRescale.CopyNew()
	if source.IsMontgomery {
		for component := range source.Value {
			ringOld.IMForm(source.Value[component], source.Value[component])
		}
	}
	result := ckks.NewCiphertext(params, preRescale.Degree(), targetLevel)
	*result.MetaData = *preRescale.MetaData
	result.IsNTT, result.IsMontgomery = preRescale.IsNTT, preRescale.IsMontgomery
	buff := ring.NewPoly(ringOld.N(), oldLevel)
	for component := range source.Value {
		ringOld.DivRoundByLastModulusManyNTT(nbRescales, source.Value[component], buff, result.Value[component])
	}
	result.Scale = preRescale.Scale
	for level := oldLevel; level > targetLevel; level-- {
		result.Scale = result.Scale.Div(rlwe.NewScale(params.RingQ().SubRings[level].Modulus))
	}
	if result.IsMontgomery {
		params.RingQ().AtLevel(targetLevel).MForm(result.Value[0], result.Value[0])
		if result.Degree() >= 1 {
			params.RingQ().AtLevel(targetLevel).MForm(result.Value[1], result.Value[1])
		}
	}
	return result, nil
}

func normalizedFastRowsMatchFull(fast, full *rlwe.Ciphertext) bool {
	return doubleAngleRowsMatch(finalizationEvidence(fast).Rows, finalizationEvidence(full).Rows)
}

func normalizedSetFailure(result *NormalizedResult, round, checkpoint, cause string) {
	result.FirstFailingRound = round
	result.FirstFailingCheckpoint = checkpoint
	result.FirstSupportedCause = cause
}

func normalizedStageComparison(fast, reference *rlwe.Ciphertext) NormalizedStageComparison {
	comparison := NormalizedStageComparison{
		Match:               normalizedFastRowsMatchFull(fast, reference),
		FastLevel:           fast.Level(),
		ReferenceLevel:      reference.Level(),
		FastScale:           finalizationScaleString(fast.Scale),
		ReferenceScale:      finalizationScaleString(reference.Scale),
		FastIsNTT:           fast.IsNTT,
		ReferenceIsNTT:      reference.IsNTT,
		FastIsMontgomery:    fast.IsMontgomery,
		ReferenceMontgomery: reference.IsMontgomery,
	}
	for component := 0; component <= 1 && component < len(fast.Value) && component < len(reference.Value); component++ {
		for limb := 0; limb < 2 && limb < len(fast.Value[component].Coeffs) && limb < len(reference.Value[component].Coeffs); limb++ {
			fastRow, referenceRow := fast.Value[component].Coeffs[limb], reference.Value[component].Coeffs[limb]
			if len(fastRow) != len(referenceRow) {
				comparison.HasMismatch = true
				comparison.Component, comparison.Limb, comparison.Index = component, limb, minInt(len(fastRow), len(referenceRow))
				comparison.FastValue, comparison.ReferenceValue = firstRowValue(fastRow, comparison.Index), firstRowValue(referenceRow, comparison.Index)
				return comparison
			}
			for index := range fastRow {
				if fastRow[index] != referenceRow[index] {
					comparison.HasMismatch = true
					comparison.Component, comparison.Limb, comparison.Index = component, limb, index
					comparison.FastValue, comparison.ReferenceValue = fastRow[index], referenceRow[index]
					return comparison
				}
			}
		}
	}
	return comparison
}

func firstRowValue(row []uint64, index int) uint64 {
	if index >= 0 && index < len(row) {
		return row[index]
	}
	return 0
}

func normalizedWriteResult(result NormalizedResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for normalized DoubleAngle design diagnostic")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	summary := struct {
		SchemaVersion          string                      `json:"schema_version"`
		Timestamp              time.Time                   `json:"timestamp"`
		Primary                RepositoryMetadata          `json:"primary_repository"`
		Lattigo                RepositoryMetadata          `json:"lattigo_repository"`
		AuthoritativeOracle    string                      `json:"authoritative_poly_oracle"`
		Threshold              float64                     `json:"threshold"`
		WorkingScaleW          string                      `json:"working_scale_w"`
		CoherentInitialScale   string                      `json:"coherent_initial_scale"`
		ExactTargetScale       string                      `json:"exact_standard_target_scale"`
		InitialScaleLog2Delta  float64                     `json:"initial_coherent_vs_exact_target_log2_delta"`
		PromotionMultiplier    string                      `json:"promotion_multiplier"`
		CanonicalLow           NormalizedStateEvidence     `json:"canonical_low"`
		Schedules              []NormalizedScaleSchedule   `json:"schedules"`
		Rounds                 []NormalizedRoundCheckpoint `json:"rounds"`
		Final                  NormalizedFinalEvidence     `json:"final"`
		DesignMechanism        string                      `json:"design_mechanism"`
		PreviousDisposition    string                      `json:"previous_classification_disposition"`
		FirstFailingRound      string                      `json:"first_failing_round"`
		FirstFailingCheckpoint string                      `json:"first_failing_checkpoint"`
		FirstSupportedCause    string                      `json:"first_supported_cause"`
		Validation             map[string]interface{}      `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.AuthoritativeOracle, result.Threshold, result.WorkingScaleW, result.CoherentInitialScale, result.ExactTargetScale, result.InitialScaleLog2Delta, result.PromotionMultiplier, result.CanonicalLow, result.Schedules, result.Rounds, result.Final, result.DesignMechanism, result.PreviousClassificationDisposition, result.FirstFailingRound, result.FirstFailingCheckpoint, result.FirstSupportedCause, result.Validation}
	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DESIGN-DOUBLE-ANGLE-FINAL-RESTORE-REFERENCE-FIX-logN13-summary.json"), summaryData, 0o644)
}

func runFIX001P3DesignNormalizedRecurrence(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	params := state.Params.BootstrappingParameters
	y0 := chebyshevRawOracle(state.Poly, state.Z)
	low := psGlobalCopy(params, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(low, low); err != nil {
		return err
	}
	lowDecoded, err := psGlobalDecode(params, low)
	if err != nil {
		return err
	}
	lowView, err := targetScaleCoefficientView(params, low)
	if err != nil {
		return err
	}
	lowData, err := targetScaleCenteredRows(params, lowView)
	if err != nil {
		return err
	}
	lowC1, err := mulAliasZeroCheck(params, low, 1)
	if err != nil {
		return err
	}
	lowEvidence := NormalizedStateEvidence{Level: low.Level(), Degree: low.Degree(), Scale: finalizationScaleString(low.Scale), Semantic: psGlobalMetric(y0, lowDecoded), Rows: finalizationEvidence(low).Rows, C1Zero: lowC1.CoefficientDomainExactlyZero && lowC1.NTTMontgomeryExactlyZero, C0MaxAbs: mulAliasMaxAbs(lowData.Values[0]).String()}
	if low.Level() != 7 || low.Degree() != 1 || !lowEvidence.Semantic.Pass || !lowEvidence.C1Zero {
		result := NormalizedResult{SchemaVersion: "fix-001-p3-design-double-angle-final-restore-reference-fix.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, CanonicalLow: lowEvidence, PreviousClassificationDisposition: "superseded_by_wrong_final_restore_reference", FirstFailingRound: "canonical", FirstFailingCheckpoint: "canonical_low", FirstSupportedCause: "normalized_double_angle_final_reference_precondition_mismatch", Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == ""}}
		return normalizedWriteResult(result, outPath)
	}

	multiplier := new(big.Int).Lsh(big.NewInt(1), 29)
	coherentInitialScale := low.Scale.Mul(rlwe.NewScale(multiplier))
	canonicalNormalized := make([]complex128, len(y0))
	for i, value := range y0 {
		canonicalNormalized[i] = value / complex(float64(multiplier.Int64()), 0)
	}
	fastNormalized := low.CopyNew()
	fastNormalized.Scale = coherentInitialScale
	fastNormalizedDecoded, err := psGlobalDecode(params, fastNormalized)
	if err != nil {
		return err
	}
	if psGlobalMetric(canonicalNormalized, fastNormalizedDecoded).MaxComponent > correctnessThreshold {
		result := NormalizedResult{SchemaVersion: "fix-001-p3-design-double-angle-final-restore-reference-fix.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, CanonicalLow: lowEvidence, CanonicalNormalizedInput: psGlobalMetric(canonicalNormalized, fastNormalizedDecoded), PreviousClassificationDisposition: "superseded_by_wrong_final_restore_reference", FirstFailingRound: "canonical", FirstFailingCheckpoint: "canonical_normalized_input", FirstSupportedCause: "normalized_double_angle_final_reference_precondition_mismatch", Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == ""}}
		return normalizedWriteResult(result, outPath)
	}

	exactTarget := state.Target
	coherentBase, err := normalizedFullFromLow(params, low, lowData)
	if err != nil {
		return err
	}
	coherentBase.Scale = low.Scale
	coherentPromoted := normalizedFullIntegerMultiply(params, coherentBase, multiplier)
	coherentPromoted.Scale = coherentInitialScale
	rCoherent := coherentPromoted
	rExactTarget := rCoherent.CopyNew()
	rExactTarget.Scale = exactTarget
	rNorm, err := normalizedFullFromLow(params, low, lowData)
	if err != nil {
		return err
	}
	rNorm.Scale = coherentInitialScale

	result := NormalizedResult{
		SchemaVersion: "fix-001-p3-design-double-angle-final-restore-reference-fix.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(params, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, WorkingScaleW: finalizationScaleString(low.Scale), CoherentInitialScale: finalizationScaleString(coherentInitialScale), ExactTargetScale: finalizationScaleString(exactTarget), InitialScaleLog2Delta: normalizedLog2Deviation(coherentInitialScale, exactTarget), PromotionMultiplier: multiplier.String(), CanonicalLow: lowEvidence, CanonicalNormalizedInput: psGlobalMetric(canonicalNormalized, fastNormalizedDecoded), DesignMechanism: "metadata_normalized_double_angle_with_power_of_two_value_factors", PreviousClassificationDisposition: "superseded_by_wrong_final_restore_reference",
		Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "canonical_root_hash_match": state.Root.RowsMatch([]string{"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf", "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"}), "previous_h0_stale_checkpoint_object": true, "previous_h0_missing_squared_scale": true, "corrected_reference_product_scale": true, "immutable_fast_stage_snapshots": true, "final_exact_row_oracle": "rNormRestored", "r_coherent_row_equality_required": false, "previous_classification_disposition": "superseded_by_wrong_final_restore_reference", "old_r_coherent_row_equality_disposition": "invalid_due_to_independent_rescale_rounding_histories", "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_coeffs_to_slots_or_later_bootstrap": true},
	}
	result.FirstFailingRound = "none"
	result.FirstFailingCheckpoint = "none"
	result.Validation["low_representation"] = fmt.Sprintf("is_ntt=%t,is_montgomery=%t", low.IsNTT, low.IsMontgomery)
	result.Validation["normalized_initial_rows_match_low"] = normalizedFastRowsMatchFull(low, rNorm)

	currentLevel := fastNormalized.Level()
	currentScale := coherentInitialScale
	kIn := 29
	zExpected := doubleAngleHPScale(doubleAngleHPVector(y0), new(big.Float).Quo(new(big.Float).SetPrec(doubleAnglePrecision).SetFloat64(1), new(big.Float).SetInt(multiplier)))
	sqrt2pi := state.Eval.Mod1Parameters.Sqrt2Pi
	for round := 0; round < state.Eval.Mod1Parameters.DoubleAngle; round++ {
		sqrt2pi *= sqrt2pi
		schedule, err := normalizedScaleSchedule(round, currentLevel, currentScale, low.Scale, kIn, sqrt2pi, params)
		if err != nil {
			return err
		}
		result.Schedules = append(result.Schedules, schedule)
		zBefore := doubleAngleHPToVector(zExpected)
		inputDecoded, err := psGlobalDecode(params, fastNormalized)
		if err != nil {
			return err
		}
		inputView, err := targetScaleCoefficientView(params, fastNormalized)
		if err != nil {
			return err
		}
		inputData, err := targetScaleCenteredRows(params, inputView)
		if err != nil {
			return err
		}
		checkpoint := NormalizedRoundCheckpoint{Round: round, Schedule: schedule, InputSemantic: psGlobalMetric(zBefore, inputDecoded), InputSquareBound: normalizedSquareBound(inputData, params.N())}
		if !checkpoint.InputSquareBound.Pass {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "pre_square_capacity", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}

		fastInput := fastNormalized.CopyNew()
		checkpoint.InputComparison = normalizedStageComparison(fastInput, rNorm)
		checkpoint.InputRowsMatchReference = checkpoint.InputComparison.Match
		if !checkpoint.InputRowsMatchReference {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "input_rows", "normalized_double_angle_reference_fix_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}

		zSquareHP := make([]doubleAngleHP, len(zExpected))
		for i, value := range zExpected {
			zSquareHP[i] = doubleAngleHPSquare(value)
		}
		zSquare := doubleAngleHPToVector(zSquareHP)
		if err := state.Eval.FastCKKS.MulRelin(fastNormalized, fastNormalized, fastNormalized); err != nil {
			return err
		}
		fastSquare := fastNormalized.CopyNew()
		refProduct, err := normalizedFullSquare(params, rNorm)
		if err != nil {
			return err
		}
		refSquare := normalizedFullZeroC2(params, refProduct)
		checkpoint.SquareComparison = normalizedStageComparison(fastSquare, refSquare)
		checkpoint.SquareRowsMatchReference = checkpoint.SquareComparison.Match
		squareDecoded, err := psGlobalDecode(params, fastSquare)
		if err != nil {
			return err
		}
		checkpoint.SquareSemantic = psGlobalMetric(zSquare, squareDecoded)
		if !checkpoint.SquareComparison.Match {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "square", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}
		if !checkpoint.SquareSemantic.Pass {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "square_semantic", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}

		factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
		if err := state.Eval.FastCKKS.MulIntegerMaintained(fastNormalized, factor, fastNormalized); err != nil {
			return err
		}
		fastAfterA := fastNormalized.CopyNew()
		refAfterA := normalizedFullIntegerMultiply(params, refSquare, factor)
		checkpoint.AfterMultiplierComparison = normalizedStageComparison(fastAfterA, refAfterA)
		checkpoint.AfterMultiplierRowsMatch = checkpoint.AfterMultiplierComparison.Match
		zAfterAHP := doubleAngleHPScale(zSquareHP, new(big.Float).SetInt(factor))
		zAfterA := doubleAngleHPToVector(zAfterAHP)
		aDecoded, err := psGlobalDecode(params, fastAfterA)
		if err != nil {
			return err
		}
		checkpoint.AfterMultiplierSemantic = psGlobalMetric(zAfterA, aDecoded)
		if !checkpoint.AfterMultiplierComparison.Match {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "after_multiplier", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}
		if !checkpoint.AfterMultiplierSemantic.Pass {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "after_multiplier_semantic", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}

		constantValue, _ := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
		checkpoint.Constant = schedule.Constant
		if err := state.Eval.FastCKKS.Add(fastNormalized, -constantValue, fastNormalized); err != nil {
			return err
		}
		fastAfterConstant := fastNormalized.CopyNew()
		refAfterConstant := normalizedFullSubtractConstant(params, refAfterA, constantValue)
		checkpoint.AfterConstantComparison = normalizedStageComparison(fastAfterConstant, refAfterConstant)
		checkpoint.AfterConstantRowsMatch = checkpoint.AfterConstantComparison.Match
		zNextHP := make([]doubleAngleHP, len(zAfterAHP))
		for i, value := range zAfterAHP {
			zNextHP[i] = doubleAngleHPOffset(value, constantValue)
		}
		zNext := doubleAngleHPToVector(zNextHP)
		constantDecoded, err := psGlobalDecode(params, fastAfterConstant)
		if err != nil {
			return err
		}
		checkpoint.AfterConstantSemantic = psGlobalMetric(zNext, constantDecoded)
		checkpoint.PreRescaleCapacity, err = normalizedExactCapacity(params, refAfterConstant)
		if err != nil {
			return err
		}
		checkpoint.PreRescaleFastRows = finalizationEvidence(fastAfterConstant).Rows
		checkpoint.PreRescaleReferenceRows = finalizationEvidence(refAfterConstant).Rows
		checkpoint.PreRescaleRowsMatch = checkpoint.AfterConstantComparison.Match
		if !checkpoint.PreRescaleCapacity.Pass {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "pre_rescale_capacity", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}
		if !checkpoint.AfterConstantComparison.Match {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "after_constant", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}
		if !checkpoint.AfterConstantSemantic.Pass {
			result.Rounds = append(result.Rounds, checkpoint)
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "after_constant_semantic", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}

		if err := state.Eval.FastCKKS.Rescale(fastNormalized, fastNormalized); err != nil {
			return err
		}
		fastPostRescale := fastNormalized.CopyNew()
		refPostRescale, err := normalizedFullRescale(params, refAfterConstant)
		if err != nil {
			return err
		}
		rNorm = refPostRescale
		postRescaleDecoded, err := psGlobalDecode(params, fastPostRescale)
		if err != nil {
			return err
		}
		checkpoint.PostRescaleSemantic = psGlobalMetric(zNext, postRescaleDecoded)
		checkpoint.PostRescaleComparison = normalizedStageComparison(fastPostRescale, refPostRescale)
		checkpoint.PostRescaleRowsMatch = checkpoint.PostRescaleComparison.Match
		checkpoint.PostRescaleScaleMatch = fastPostRescale.Level() == schedule.NextLevel && finalizationScaleString(fastPostRescale.Scale) == schedule.ScaleOut
		checkpoint.PostRescaleEffectiveScale = finalizationScaleString(fastPostRescale.Scale.Div(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))))
		checkpoint.PostRescaleFastRows = finalizationEvidence(fastPostRescale).Rows
		checkpoint.PostRescaleReferenceRows = finalizationEvidence(refPostRescale).Rows
		result.Rounds = append(result.Rounds, checkpoint)
		if !checkpoint.PostRescaleComparison.Match || !checkpoint.PostRescaleScaleMatch {
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "post_rescale", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}
		if !checkpoint.InputSemantic.Pass || !checkpoint.PostRescaleSemantic.Pass {
			normalizedSetFailure(&result, fmt.Sprintf("%d", round), "post_rescale_semantic", "normalized_double_angle_final_reference_precondition_mismatch")
			return normalizedWriteResult(result, outPath)
		}

		coherentPre, _, err := normalizedFullSquareStep(params, rCoherent, big.NewInt(2), sqrt2pi)
		if err != nil {
			return err
		}
		exactPre, _, err := normalizedFullSquareStep(params, rExactTarget, big.NewInt(2), sqrt2pi)
		if err != nil {
			return err
		}
		coherentPre, err = normalizedFullRescale(params, coherentPre)
		if err != nil {
			return err
		}
		exactPre, err = normalizedFullRescale(params, exactPre)
		if err != nil {
			return err
		}
		rCoherent, rExactTarget = coherentPre, exactPre

		zExpected = zNextHP
		currentLevel = schedule.NextLevel
		currentScale = scheduleScaleValue(schedule, params, currentLevel, currentScale)
		kIn = schedule.KOutExponent
	}

	fastBeforeRestore := fastNormalized.CopyNew()
	rNormBeforeRestore := rNorm.CopyNew()
	k3 := new(big.Int).Lsh(big.NewInt(1), uint(kIn))
	restoreView, err := targetScaleCoefficientView(params, fastBeforeRestore)
	if err != nil {
		return err
	}
	restoreData, err := targetScaleCenteredRows(params, restoreView)
	if err != nil {
		return err
	}
	final := NormalizedFinalEvidence{
		K3: k3.String(), BeforeResetScale: finalizationScaleString(fastBeforeRestore.Scale),
		RestoreCapacity:                    normalizedFinalRestoreCapacity(restoreData, k3),
		FastBeforeRestoreRows:              finalizationEvidence(fastBeforeRestore).Rows,
		RNormBeforeRestoreRows:             finalizationEvidence(rNormBeforeRestore).Rows,
		RCoherentRows:                      finalizationEvidence(rCoherent).Rows,
		OldRCoherentRowEqualityDisposition: "invalid_due_to_independent_rescale_rounding_histories",
	}
	if !final.RestoreCapacity.Pass {
		result.Final = final
		normalizedSetFailure(&result, "final", "restore_capacity", "normalized_double_angle_final_restore_capacity_failure")
		return normalizedWriteResult(result, outPath)
	}

	fastRestored := fastBeforeRestore.CopyNew()
	if err := state.Eval.FastCKKS.MulIntegerMaintained(fastBeforeRestore, k3, fastRestored); err != nil {
		return err
	}
	rNormRestored := normalizedFullIntegerMultiply(params, rNormBeforeRestore, k3)
	final.FastRestoredRows = finalizationEvidence(fastRestored).Rows
	final.RNormRestoredRows = finalizationEvidence(rNormRestored).Rows
	final.NormalizedRestoredComparison = normalizedStageComparison(fastRestored, rNormRestored)
	final.NormalizedRestoredRowsMatch = final.NormalizedRestoredComparison.Match
	final.FastRestoreMetadataPass = fastRestored.Level() == fastBeforeRestore.Level() && fastRestored.Degree() == fastBeforeRestore.Degree() && fastRestored.IsNTT == fastBeforeRestore.IsNTT && fastRestored.IsMontgomery == fastBeforeRestore.IsMontgomery && fastRestored.Scale.Equal(fastBeforeRestore.Scale)
	final.RNormRestoreMetadataPass = rNormRestored.Level() == rNormBeforeRestore.Level() && rNormRestored.Degree() == rNormBeforeRestore.Degree() && rNormRestored.IsNTT == rNormBeforeRestore.IsNTT && rNormRestored.IsMontgomery == rNormBeforeRestore.IsMontgomery && rNormRestored.Scale.Equal(rNormBeforeRestore.Scale)
	fastRestoreC1, err := mulAliasZeroCheck(params, fastRestored, 1)
	if err != nil {
		return err
	}
	rNormRestoreC1, err := mulAliasZeroCheck(params, rNormRestored, 1)
	if err != nil {
		return err
	}
	final.FastRestoreC1Zero = fastRestoreC1.CoefficientDomainExactlyZero && fastRestoreC1.NTTMontgomeryExactlyZero
	final.RNormRestoreC1Zero = rNormRestoreC1.CoefficientDomainExactlyZero && rNormRestoreC1.NTTMontgomeryExactlyZero
	if !final.NormalizedRestoredRowsMatch || !final.FastRestoreMetadataPass || !final.RNormRestoreMetadataPass || !final.FastRestoreC1Zero || !final.RNormRestoreC1Zero {
		result.Final = final
		normalizedSetFailure(&result, "final", "k3_restore", "normalized_double_angle_final_restore_primitive_mismatch")
		return normalizedWriteResult(result, outPath)
	}

	inputScale := state.InputScaleValue
	fastRestoredBeforeResetRows := finalizationEvidence(fastRestored).Rows
	rNormRestoredBeforeResetRows := finalizationEvidence(rNormRestored).Rows
	fastRestoredDecoded, err := psGlobalDecode(params, fastRestored)
	if err != nil {
		return err
	}
	rNormRestoredDecoded, err := psGlobalDecode(params, rNormRestored)
	if err != nil {
		return err
	}
	rCoherentBeforeReset := rCoherent.CopyNew()
	rExactTargetBeforeReset := rExactTarget.CopyNew()
	coherentBeforeReset, err := psGlobalDecode(params, rCoherentBeforeReset)
	if err != nil {
		return err
	}
	exactBeforeReset, err := psGlobalDecode(params, rExactTargetBeforeReset)
	if err != nil {
		return err
	}
	final.FastVsNormalizedRestoredSemantic = psGlobalMetric(rNormRestoredDecoded, fastRestoredDecoded)
	final.NormalizedRestoredVsRCoherentSemantic = psGlobalMetric(coherentBeforeReset, rNormRestoredDecoded)
	final.NormalizedRestoredVsExactTargetSemantic = psGlobalMetric(exactBeforeReset, rNormRestoredDecoded)
	final.FastVsRCoherentSemantic = psGlobalMetric(coherentBeforeReset, fastRestoredDecoded)
	final.RCoherentVsExactTarget = psGlobalMetric(coherentBeforeReset, exactBeforeReset)
	final.FastVsExactTargetSemantic = psGlobalMetric(exactBeforeReset, fastRestoredDecoded)
	final.ExactTargetCompatibility = final.RCoherentVsExactTarget.Pass
	if !final.FastVsNormalizedRestoredSemantic.Pass || !final.NormalizedRestoredVsRCoherentSemantic.Pass || !final.NormalizedRestoredVsExactTargetSemantic.Pass || !final.FastVsRCoherentSemantic.Pass || !final.RCoherentVsExactTarget.Pass || !final.FastVsExactTargetSemantic.Pass {
		result.Final = final
		normalizedSetFailure(&result, "final", "final_semantic", "normalized_double_angle_final_restore_semantic_failure")
		return normalizedWriteResult(result, outPath)
	}

	resetRatio, _ := new(big.Float).Quo(new(big.Float).Set(&fastRestored.Scale.Value), new(big.Float).Set(&inputScale.Value)).Float64()
	expectedReset := make([]complex128, len(fastRestoredDecoded))
	for index, value := range fastRestoredDecoded {
		expectedReset[index] = value * complex(resetRatio, 0)
	}
	expectedNormalizedReset := make([]complex128, len(rNormRestoredDecoded))
	for index, value := range rNormRestoredDecoded {
		expectedNormalizedReset[index] = value * complex(resetRatio, 0)
	}
	fastRestored.Scale = inputScale
	rNormRestored.Scale = inputScale
	rCoherent.Scale = inputScale
	rExactTarget.Scale = inputScale
	fastFinal, err := psGlobalDecode(params, fastRestored)
	if err != nil {
		return err
	}
	normalizedFinal, err := psGlobalDecode(params, rNormRestored)
	if err != nil {
		return err
	}
	final.FastFinalRowsAfterReset = finalizationEvidence(fastRestored).Rows
	final.NormalizedReferenceFinalRowsAfterReset = finalizationEvidence(rNormRestored).Rows
	final.RowsMatchAfterReset = doubleAngleRowsMatch(final.FastFinalRowsAfterReset, final.NormalizedReferenceFinalRowsAfterReset)
	final.FastRowsUnchangedOnReset = doubleAngleRowsMatch(final.FastFinalRowsAfterReset, fastRestoredBeforeResetRows)
	final.NormalizedReferenceRowsUnchangedOnReset = doubleAngleRowsMatch(final.NormalizedReferenceFinalRowsAfterReset, rNormRestoredBeforeResetRows)
	final.FastFinalLevelDegreeUnchanged = fastRestored.Level() == fastBeforeRestore.Level() && fastRestored.Degree() == fastBeforeRestore.Degree()
	final.NormalizedFinalLevelDegreeUnchanged = rNormRestored.Level() == rNormBeforeRestore.Level() && rNormRestored.Degree() == rNormBeforeRestore.Degree()
	final.FinalScaleResetSemantic = psGlobalMetric(expectedReset, fastFinal)
	final.NormalizedFinalScaleResetSemantic = psGlobalMetric(expectedNormalizedReset, normalizedFinal)
	final.FinalScaleResetPass = final.RowsMatchAfterReset && final.FastRowsUnchangedOnReset && final.NormalizedReferenceRowsUnchangedOnReset && final.FastFinalLevelDegreeUnchanged && final.NormalizedFinalLevelDegreeUnchanged && fastRestored.Scale.Equal(inputScale) && rNormRestored.Scale.Equal(inputScale) && final.FinalScaleResetSemantic.Pass && final.NormalizedFinalScaleResetSemantic.Pass
	final.RCoherentFinalRows = finalizationEvidence(rCoherent).Rows
	final.FinalSemanticPass = final.FastVsNormalizedRestoredSemantic.Pass && final.NormalizedRestoredVsRCoherentSemantic.Pass && final.NormalizedRestoredVsExactTargetSemantic.Pass && final.FastVsRCoherentSemantic.Pass && final.RCoherentVsExactTarget.Pass && final.FastVsExactTargetSemantic.Pass
	result.Final = final
	if !final.FinalScaleResetPass {
		normalizedSetFailure(&result, "final", "scale_reset", "normalized_double_angle_final_scale_reset_failure")
	} else if !final.FinalSemanticPass {
		normalizedSetFailure(&result, "final", "final_semantic", "normalized_double_angle_final_restore_semantic_failure")
	} else {
		normalizedSetFailure(&result, "none", "none", "normalized_double_angle_recurrence_validated_after_final_reference_fix")
	}
	return normalizedWriteResult(result, outPath)
}

func scheduleScaleValue(schedule NormalizedScaleSchedule, params ckks.Parameters, level int, previous rlwe.Scale) rlwe.Scale {
	result := previous.Mul(previous)
	for i, modulus := range schedule.DroppedModuli {
		_ = i
		_ = level
		result = result.Div(rlwe.NewScale(modulus))
	}
	return result
}
