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

type NormalizedReferenceCheckpoint struct {
	Level         int                          `json:"level"`
	Degree        int                          `json:"degree"`
	Scale         string                       `json:"scale"`
	Rows          []FinalizationRowFingerprint `json:"q0_q1_rows"`
	ExactCapacity NormalizedExactCapacity      `json:"exact_coefficient_capacity,omitempty"`
}

type NormalizedRoundCheckpoint struct {
	Round                     int                          `json:"round"`
	Schedule                  NormalizedScaleSchedule      `json:"schedule"`
	InputSemantic             *PSGlobalMetric              `json:"input_normalized_semantic"`
	InputSquareBound          NormalizedSquareBound        `json:"input_square_bound"`
	SquareSemantic            *PSGlobalMetric              `json:"square_semantic"`
	AfterMultiplierSemantic   *PSGlobalMetric              `json:"after_a_semantic"`
	Constant                  string                       `json:"normalized_constant"`
	AfterConstantSemantic     *PSGlobalMetric              `json:"after_constant_semantic"`
	PreRescaleCapacity        NormalizedExactCapacity      `json:"pre_rescale_exact_reference_capacity"`
	SquareRowsMatchReference  bool                         `json:"square_rows_match_reference"`
	AfterMultiplierRowsMatch  bool                         `json:"after_multiplier_rows_match_reference"`
	AfterConstantRowsMatch    bool                         `json:"after_constant_rows_match_reference"`
	PreRescaleRowsMatch       bool                         `json:"pre_rescale_rows_match_reference"`
	PreRescaleFastRows        []FinalizationRowFingerprint `json:"pre_rescale_fast_rows"`
	PreRescaleReferenceRows   []FinalizationRowFingerprint `json:"pre_rescale_reference_rows"`
	PostRescaleSemantic       *PSGlobalMetric              `json:"post_rescale_semantic"`
	PostRescaleRowsMatch      bool                         `json:"post_rescale_rows_match_reference"`
	PostRescaleScaleMatch     bool                         `json:"post_rescale_scale_match"`
	PostRescaleEffectiveScale string                       `json:"post_rescale_effective_scale"`
	PostRescaleFastRows       []FinalizationRowFingerprint `json:"post_rescale_fast_rows"`
	PostRescaleReferenceRows  []FinalizationRowFingerprint `json:"post_rescale_reference_rows"`
}

type NormalizedFinalEvidence struct {
	K3                        string                       `json:"k3"`
	BeforeResetScale          string                       `json:"before_reset_scale"`
	FastRestoredRowsMatch     bool                         `json:"fast_restored_rows_match_r_coherent"`
	FastRestoredRows          []FinalizationRowFingerprint `json:"fast_restored_rows"`
	RCoherentRows             []FinalizationRowFingerprint `json:"r_coherent_rows"`
	FastVsRCoherentSemantic   *PSGlobalMetric              `json:"fast_vs_r_coherent_semantic"`
	RCoherentVsExactTarget    *PSGlobalMetric              `json:"r_coherent_vs_r_exact_target_semantic"`
	FastVsExactTargetSemantic *PSGlobalMetric              `json:"fast_vs_r_exact_target_semantic"`
	FastFinalRowsAfterReset   []FinalizationRowFingerprint `json:"fast_final_rows_after_reset"`
	RCoherentFinalRows        []FinalizationRowFingerprint `json:"r_coherent_final_rows_after_reset"`
	RowsMatchAfterReset       bool                         `json:"rows_match_after_reset"`
	FinalSemanticPass         bool                         `json:"final_semantic_pass"`
	ExactTargetCompatibility  bool                         `json:"exact_target_compatibility_pass"`
}

type NormalizedResult struct {
	SchemaVersion            string                      `json:"schema_version"`
	Timestamp                time.Time                   `json:"timestamp"`
	Primary                  RepositoryMetadata          `json:"primary_repository"`
	Lattigo                  RepositoryMetadata          `json:"lattigo_repository"`
	Environment              EnvironmentMetadata         `json:"environment"`
	Config                   BootstrapConfig             `json:"config"`
	Parameters               ExperimentParameters        `json:"effective_parameters"`
	Workload                 CorrectnessWorkload         `json:"workload"`
	AuthoritativeOracle      string                      `json:"authoritative_poly_oracle"`
	Threshold                float64                     `json:"threshold"`
	WorkingScaleW            string                      `json:"working_scale_w"`
	CoherentInitialScale     string                      `json:"coherent_initial_scale"`
	ExactTargetScale         string                      `json:"exact_standard_target_scale"`
	InitialScaleLog2Delta    float64                     `json:"initial_coherent_vs_exact_target_log2_delta"`
	PromotionMultiplier      string                      `json:"promotion_multiplier"`
	CanonicalLow             NormalizedStateEvidence     `json:"canonical_low"`
	CanonicalNormalizedInput *PSGlobalMetric             `json:"canonical_normalized_input_semantic"`
	Schedules                []NormalizedScaleSchedule   `json:"schedules"`
	Rounds                   []NormalizedRoundCheckpoint `json:"rounds"`
	Final                    NormalizedFinalEvidence     `json:"final"`
	DesignMechanism          string                      `json:"design_mechanism"`
	FirstSupportedCause      string                      `json:"first_supported_cause"`
	Validation               map[string]interface{}      `json:"validation"`
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

func normalizedFullSquareStep(params ckks.Parameters, input *rlwe.Ciphertext, factor *big.Int, constant float64) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	if input == nil || !input.IsNTT {
		return nil, nil, fmt.Errorf("normalized full reference requires an NTT input")
	}
	level := input.Level()
	ringQ := params.RingQ().AtLevel(level)
	product := ckks.NewCiphertext(params, 2, level)
	*product.MetaData = *input.MetaData
	product.IsNTT, product.IsMontgomery = input.IsNTT, input.IsMontgomery
	if input.IsMontgomery {
		ringQ.MulCoeffsMontgomery(input.Value[0], input.Value[0], product.Value[0])
	} else {
		ringQ.MulCoeffsBarrett(input.Value[0], input.Value[0], product.Value[0])
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

func normalizedFirstRowMismatch(left, right *rlwe.Ciphertext) string {
	if left == nil || right == nil {
		return "nil ciphertext"
	}
	for component := 0; component <= 1 && component < len(left.Value) && component < len(right.Value); component++ {
		for limb := 0; limb < 2 && limb < len(left.Value[component].Coeffs) && limb < len(right.Value[component].Coeffs); limb++ {
			leftRow, rightRow := left.Value[component].Coeffs[limb], right.Value[component].Coeffs[limb]
			for index := 0; index < len(leftRow) && index < len(rightRow); index++ {
				if leftRow[index] != rightRow[index] {
					return fmt.Sprintf("component=%d limb=%d index=%d fast=%d reference=%d", component, limb, index, leftRow[index], rightRow[index])
				}
			}
		}
	}
	return "none"
}

func normalizedScaleRatioString(scale, divisor rlwe.Scale) string {
	ratio := new(big.Float).Quo(new(big.Float).SetPrec(512).Set(&scale.Value), new(big.Float).SetPrec(512).Set(&divisor.Value))
	return ratio.Text('e', 80)
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
		SchemaVersion         string                      `json:"schema_version"`
		Timestamp             time.Time                   `json:"timestamp"`
		Primary               RepositoryMetadata          `json:"primary_repository"`
		Lattigo               RepositoryMetadata          `json:"lattigo_repository"`
		AuthoritativeOracle   string                      `json:"authoritative_poly_oracle"`
		Threshold             float64                     `json:"threshold"`
		WorkingScaleW         string                      `json:"working_scale_w"`
		CoherentInitialScale  string                      `json:"coherent_initial_scale"`
		ExactTargetScale      string                      `json:"exact_standard_target_scale"`
		InitialScaleLog2Delta float64                     `json:"initial_coherent_vs_exact_target_log2_delta"`
		PromotionMultiplier   string                      `json:"promotion_multiplier"`
		CanonicalLow          NormalizedStateEvidence     `json:"canonical_low"`
		Schedules             []NormalizedScaleSchedule   `json:"schedules"`
		Rounds                []NormalizedRoundCheckpoint `json:"rounds"`
		Final                 NormalizedFinalEvidence     `json:"final"`
		DesignMechanism       string                      `json:"design_mechanism"`
		FirstSupportedCause   string                      `json:"first_supported_cause"`
		Validation            map[string]interface{}      `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.AuthoritativeOracle, result.Threshold, result.WorkingScaleW, result.CoherentInitialScale, result.ExactTargetScale, result.InitialScaleLog2Delta, result.PromotionMultiplier, result.CanonicalLow, result.Schedules, result.Rounds, result.Final, result.DesignMechanism, result.FirstSupportedCause, result.Validation}
	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-RECURRENCE-logN13-summary.json"), summaryData, 0o644)
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
		result := NormalizedResult{SchemaVersion: "fix-001-p3-design-double-angle-normalized-recurrence.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, CanonicalLow: lowEvidence, FirstSupportedCause: "normalized_double_angle_precondition_mismatch", Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == ""}}
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
		result := NormalizedResult{SchemaVersion: "fix-001-p3-design-double-angle-normalized-recurrence.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, CanonicalLow: lowEvidence, CanonicalNormalizedInput: psGlobalMetric(canonicalNormalized, fastNormalizedDecoded), FirstSupportedCause: "normalized_double_angle_precondition_mismatch", Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == ""}}
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
		SchemaVersion: "fix-001-p3-design-double-angle-normalized-recurrence.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(params, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, WorkingScaleW: finalizationScaleString(low.Scale), CoherentInitialScale: finalizationScaleString(coherentInitialScale), ExactTargetScale: finalizationScaleString(exactTarget), InitialScaleLog2Delta: normalizedLog2Deviation(coherentInitialScale, exactTarget), PromotionMultiplier: multiplier.String(), CanonicalLow: lowEvidence, CanonicalNormalizedInput: psGlobalMetric(canonicalNormalized, fastNormalizedDecoded), DesignMechanism: "metadata_normalized_double_angle_with_power_of_two_value_factors",
		Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "canonical_root_hash_match": state.Root.RowsMatch([]string{"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf", "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"}), "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_coeffs_to_slots_or_later_bootstrap": true},
	}
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
			result.FirstSupportedCause = "normalized_double_angle_square_capacity_failure"
			return normalizedWriteResult(result, outPath)
		}

		zSquareHP := make([]doubleAngleHP, len(zExpected))
		for i, value := range zExpected {
			zSquareHP[i] = doubleAngleHPSquare(value)
		}
		zSquare := doubleAngleHPToVector(zSquareHP)
		fastSquareInput := fastNormalized.CopyNew()
		if err := state.Eval.FastCKKS.MulRelin(fastNormalized, fastNormalized, fastNormalized); err != nil {
			return err
		}
		squareDecoded, err := psGlobalDecode(params, fastNormalized)
		if err != nil {
			return err
		}
		checkpoint.SquareSemantic = psGlobalMetric(zSquare, squareDecoded)

		factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
		if err := state.Eval.FastCKKS.MulIntegerMaintained(fastNormalized, factor, fastNormalized); err != nil {
			return err
		}
		zAfterAHP := doubleAngleHPScale(zSquareHP, new(big.Float).SetInt(factor))
		zAfterA := doubleAngleHPToVector(zAfterAHP)
		aDecoded, err := psGlobalDecode(params, fastNormalized)
		if err != nil {
			return err
		}
		checkpoint.AfterMultiplierSemantic = psGlobalMetric(zAfterA, aDecoded)

		constantValue, _ := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
		checkpoint.Constant = schedule.Constant
		if err := state.Eval.FastCKKS.Add(fastNormalized, -constantValue, fastNormalized); err != nil {
			return err
		}
		zNextHP := make([]doubleAngleHP, len(zAfterAHP))
		for i, value := range zAfterAHP {
			zNextHP[i] = doubleAngleHPOffset(value, constantValue)
		}
		zNext := doubleAngleHPToVector(zNextHP)
		constantDecoded, err := psGlobalDecode(params, fastNormalized)
		if err != nil {
			return err
		}
		checkpoint.AfterConstantSemantic = psGlobalMetric(zNext, constantDecoded)

		fullPre, fullProduct, err := normalizedFullSquareStep(params, rNorm, factor, constantValue)
		if err != nil {
			return err
		}
		fullSquare := normalizedFullZeroC2(params, fullProduct)
		fullAfterA := normalizedFullIntegerMultiply(params, fullSquare, factor)
		fullAfterConstant := normalizedFullSubtractConstant(params, fullAfterA, constantValue)
		checkpoint.SquareRowsMatchReference = normalizedFastRowsMatchFull(fastNormalized, fullSquare)
		checkpoint.AfterMultiplierRowsMatch = normalizedFastRowsMatchFull(fastNormalized, fullAfterA)
		checkpoint.AfterConstantRowsMatch = normalizedFastRowsMatchFull(fastNormalized, fullAfterConstant)
		if round == 0 {
			result.Validation["first_square_row_mismatch"] = normalizedFirstRowMismatch(fastNormalized, fullSquare)
			result.Validation["square_input_rows_match_reference"] = normalizedFastRowsMatchFull(fastSquareInput, rNorm)
		}
		checkpoint.PreRescaleCapacity, err = normalizedExactCapacity(params, fullPre)
		if err != nil {
			return err
		}
		checkpoint.PreRescaleFastRows = finalizationEvidence(fastNormalized).Rows
		checkpoint.PreRescaleReferenceRows = finalizationEvidence(fullPre).Rows
		checkpoint.PreRescaleRowsMatch = normalizedFastRowsMatchFull(fastNormalized, fullPre)
		if !checkpoint.PreRescaleCapacity.Pass {
			result.Rounds = append(result.Rounds, checkpoint)
			result.FirstSupportedCause = "normalized_double_angle_pre_rescale_capacity_failure"
			return normalizedWriteResult(result, outPath)
		}
		if !checkpoint.PreRescaleRowsMatch {
			result.Rounds = append(result.Rounds, checkpoint)
			result.FirstSupportedCause = "normalized_double_angle_arithmetic_mismatch"
			return normalizedWriteResult(result, outPath)
		}
		rNorm = fullPre

		if err := state.Eval.FastCKKS.Rescale(fastNormalized, fastNormalized); err != nil {
			return err
		}
		fullPre, err = normalizedFullRescale(params, fullPre)
		if err != nil {
			return err
		}
		rNorm = fullPre
		postRescaleDecoded, err := psGlobalDecode(params, fastNormalized)
		if err != nil {
			return err
		}
		checkpoint.PostRescaleSemantic = psGlobalMetric(zNext, postRescaleDecoded)
		checkpoint.PostRescaleRowsMatch = normalizedFastRowsMatchFull(fastNormalized, fullPre)
		checkpoint.PostRescaleScaleMatch = fastNormalized.Level() == schedule.NextLevel && finalizationScaleString(fastNormalized.Scale) == schedule.ScaleOut
		checkpoint.PostRescaleEffectiveScale = finalizationScaleString(fastNormalized.Scale.Div(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))))
		checkpoint.PostRescaleFastRows = finalizationEvidence(fastNormalized).Rows
		checkpoint.PostRescaleReferenceRows = finalizationEvidence(fullPre).Rows
		result.Rounds = append(result.Rounds, checkpoint)
		if !checkpoint.InputSemantic.Pass || !checkpoint.SquareSemantic.Pass || !checkpoint.AfterMultiplierSemantic.Pass || !checkpoint.AfterConstantSemantic.Pass || !checkpoint.PostRescaleSemantic.Pass || !checkpoint.PostRescaleRowsMatch || !checkpoint.PostRescaleScaleMatch {
			result.FirstSupportedCause = "normalized_double_angle_precision_failure"
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
	k3 := new(big.Int).Lsh(big.NewInt(1), uint(kIn))
	if err := state.Eval.FastCKKS.MulIntegerMaintained(fastNormalized, k3, fastNormalized); err != nil {
		return err
	}
	fastRestoredRows := finalizationEvidence(fastNormalized).Rows
	coherentRows := finalizationEvidence(rCoherent).Rows
	fastRestoredRowsMatch := doubleAngleRowsMatch(fastRestoredRows, finalizationEvidence(rCoherent).Rows)
	inputScale := state.InputScaleValue
	fastNormalized.Scale = inputScale
	rCoherent.Scale = inputScale
	rExactTarget.Scale = inputScale
	fastFinal, err := psGlobalDecode(params, fastNormalized)
	if err != nil {
		return err
	}
	coherentFinal, err := psGlobalDecode(params, rCoherent)
	if err != nil {
		return err
	}
	exactFinal, err := psGlobalDecode(params, rExactTarget)
	if err != nil {
		return err
	}
	final := NormalizedFinalEvidence{K3: k3.String(), BeforeResetScale: finalizationScaleString(fastBeforeRestore.Scale), FastRestoredRowsMatch: fastRestoredRowsMatch, FastRestoredRows: fastRestoredRows, RCoherentRows: coherentRows, FastVsRCoherentSemantic: psGlobalMetric(coherentFinal, fastFinal), RCoherentVsExactTarget: psGlobalMetric(coherentFinal, exactFinal), FastVsExactTargetSemantic: psGlobalMetric(exactFinal, fastFinal), FastFinalRowsAfterReset: finalizationEvidence(fastNormalized).Rows, RCoherentFinalRows: finalizationEvidence(rCoherent).Rows}
	final.RowsMatchAfterReset = doubleAngleRowsMatch(final.FastFinalRowsAfterReset, final.RCoherentFinalRows)
	final.FinalSemanticPass = final.FastVsRCoherentSemantic.Pass
	final.ExactTargetCompatibility = final.RCoherentVsExactTarget.Pass
	result.Final = final
	result.DesignMechanism = "metadata_normalized_double_angle_with_power_of_two_value_factors"
	if !fastRestoredRowsMatch || !final.RowsMatchAfterReset || !final.FinalSemanticPass {
		result.FirstSupportedCause = "normalized_double_angle_arithmetic_mismatch"
	} else if !final.ExactTargetCompatibility {
		result.FirstSupportedCause = "normalized_double_angle_exact_target_compatibility_failure"
	} else {
		result.FirstSupportedCause = "normalized_double_angle_recurrence_validated"
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
