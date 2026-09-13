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
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3DAR0LocalQ2PrimaryBase = "b4cdbb28ed8071d8d5dea3059bb0313ce73b5dc2"
	requiredFIX001P3DAR0LocalQ2Secondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
)

type fix001P3DAR0LocalQ2Result struct {
	SchemaVersion       string                        `json:"schema_version"`
	Timestamp           time.Time                     `json:"timestamp"`
	Primary             RepositoryMetadata            `json:"primary_repository"`
	Lattigo             RepositoryMetadata            `json:"lattigo_repository"`
	Environment         EnvironmentMetadata           `json:"environment"`
	Config              BootstrapConfig               `json:"config"`
	Parameters          ExperimentParameters          `json:"effective_parameters"`
	R0Control           map[string]interface{}        `json:"r0_capacity_control"`
	Expansion           map[string]interface{}        `json:"q01_to_q012_expansion"`
	R2Capacity          map[string]interface{}        `json:"round0_q012_multiplier_constant"`
	R3Rescale           map[string]interface{}        `json:"round0_q012_rescale"`
	Contraction         map[string]interface{}        `json:"contraction_to_q01"`
	LaterDA             []map[string]interface{}      `json:"later_da_checkpoints,omitempty"`
	Final               map[string]interface{}        `json:"final_metrics,omitempty"`
	Operations          fix001P3DAR0LocalQ2Operations `json:"static_q2_costs"`
	Classification      string                        `json:"classification"`
	SystemSufficient    bool                          `json:"ps_da_arithmetic_system_sufficient"`
	FirstRemainingBlock string                        `json:"first_remaining_blocker"`
	Validation          map[string]interface{}        `json:"validation"`
}

type fix001P3DAR0LocalQ2Operations struct {
	CoefficientTransforms int    `json:"q01_to_q012_coefficient_transforms"`
	Q2NTT                 int    `json:"q2_ntt_count"`
	Q2INTT                int    `json:"q2_intt_count"`
	Q2ArithmeticPasses    int    `json:"q2_arithmetic_passes_multiplier_constant_rescale"`
	RescaleBoundaries     int    `json:"q2_rescale_boundaries"`
	ActiveRegion          string `json:"active_region"`
}

type fix001P3DAR0LocalQ2State struct {
	Path              evalModMatchedPath
	FastNormalized    *rlwe.Ciphertext
	NormalizedCurrent *rlwe.Ciphertext
	CoherentCurrent   *rlwe.Ciphertext
	StandardCurrent   *rlwe.Ciphertext
	OriginalFastScale rlwe.Scale
	OriginalStdScale  rlwe.Scale
	WorkingScale      rlwe.Scale
	CurrentLevel      int
	CurrentScale      rlwe.Scale
	KIn               int
	Sqrt2Pi           float64
	TargetScale       rlwe.Scale
}

func fix001P3DAR0LocalQ2Write(result fix001P3DAR0LocalQ2Result, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3DAR0SignedValues(params ckks.Parameters, ct *rlwe.Ciphertext, level int) ([][]*big.Int, error) {
	if ct == nil || !ct.IsNTT || level < 0 || level > ct.Level() {
		return nil, fmt.Errorf("cannot reconstruct q012 values at level %d", level)
	}
	ringQ := params.RingQ().AtLevel(level)
	values := make([][]*big.Int, len(ct.Value))
	for component := range ct.Value {
		poly := ring.NewPoly(params.N(), level)
		for limb := 0; limb <= level; limb++ {
			copy(poly.Coeffs[limb], ct.Value[component].Coeffs[limb])
			if ct.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(poly.Coeffs[limb], poly.Coeffs[limb])
			}
		}
		coeff := ring.NewPoly(params.N(), level)
		ringQ.INTT(poly, coeff)
		values[component] = postMod1S2CFullCRT(params, coeff, level)
	}
	return values, nil
}

func fix001P3DAR0Capacity(params ckks.Parameters, ct *rlwe.Ciphertext, level int) (fix001P3LocalQ2Capacity, error) {
	values, err := fix001P3DAR0SignedValues(params, ct, level)
	if err != nil {
		return fix001P3LocalQ2Capacity{}, err
	}
	q := big.NewInt(1)
	for limb := 0; limb <= level; limb++ {
		q.Mul(q, new(big.Int).SetUint64(params.RingQ().SubRings[limb].Modulus))
	}
	flat := make([]*big.Int, 0)
	for _, component := range values {
		flat = append(flat, component...)
	}
	return fix001P3LocalQ2CapacityOf(flat, q), nil
}

func fix001P3DAR0Q01Values(params ckks.Parameters, ct *rlwe.Ciphertext) ([][]*big.Int, error) {
	return fix001P3DAR0SignedValues(params, ct, 1)
}

func fix001P3DAR0ExpandQ2(params ckks.Parameters, source *rlwe.Ciphertext, signed [][]*big.Int) (*rlwe.Ciphertext, error) {
	if source == nil || source.Level() < 2 || len(signed) != len(source.Value) {
		return nil, fmt.Errorf("q2 expansion requires a level >= 2 source")
	}
	expanded := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*expanded.MetaData = *source.MetaData
	expanded.IsNTT, expanded.IsMontgomery = true, true
	for component := range source.Value {
		for limb := 0; limb < 2; limb++ {
			copy(expanded.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
		normalNTT := postMod1S2CLiftSigned(params, signed[component], source.Level())
		params.RingQ().SubRings[2].MForm(normalNTT.Coeffs[2], expanded.Value[component].Coeffs[2])
	}
	return expanded, nil
}

func fix001P3DAR0ContractQ01(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	signed, err := fix001P3DAR0Q01Values(params, source)
	if err != nil {
		return nil, err
	}
	contract := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*contract.MetaData = *source.MetaData
	contract.IsNTT, contract.IsMontgomery = true, true
	for component := range signed {
		normalNTT := postMod1S2CLiftSigned(params, signed[component], source.Level())
		for limb := 0; limb < 2; limb++ {
			params.RingQ().SubRings[limb].MForm(normalNTT.Coeffs[limb], contract.Value[component].Coeffs[limb])
		}
	}
	return contract, nil
}

func fix001P3DAR0RowsAtLimb(params ckks.Parameters, a, b *rlwe.Ciphertext, limb int) bool {
	if a == nil || b == nil || a.Degree() != b.Degree() || limb > a.Level() || limb > b.Level() {
		return false
	}
	for component := 0; component <= a.Degree(); component++ {
		arow := append([]uint64(nil), a.Value[component].Coeffs[limb]...)
		brow := append([]uint64(nil), b.Value[component].Coeffs[limb]...)
		if a.IsMontgomery {
			params.RingQ().SubRings[limb].IMForm(arow, arow)
		}
		if b.IsMontgomery {
			params.RingQ().SubRings[limb].IMForm(brow, brow)
		}
		if !fix001P3LocalQ2RowEqual(arow, brow) {
			return false
		}
	}
	return true
}

func fix001P3DAR0RoundedDivisionMatches(params ckks.Parameters, before, after *rlwe.Ciphertext, divisor *big.Int) (bool, error) {
	beforeValues, err := fix001P3DAR0SignedValues(params, before, 2)
	if err != nil {
		return false, err
	}
	afterValues, err := fix001P3DAR0SignedValues(params, after, 2)
	if err != nil {
		return false, err
	}
	if len(beforeValues) != len(afterValues) {
		return false, nil
	}
	for component := range beforeValues {
		if len(beforeValues[component]) != len(afterValues[component]) {
			return false, nil
		}
		for i, value := range beforeValues[component] {
			if fix001P3LocalQ2Round(value, divisor).Cmp(afterValues[component][i]) != 0 {
				return false, nil
			}
		}
	}
	return true, nil
}

func fix001P3DAR0CheckpointSummary(checkpoint evalModMatchedCheckpoint) map[string]interface{} {
	return map[string]interface{}{
		"name":                     checkpoint.Name,
		"round":                    checkpoint.Round,
		"fast_q01_capacity_safe":   checkpoint.FastQ01Safe,
		"normalized_capacity_safe": checkpoint.NormalizedSafe,
		"rows_match":               checkpoint.RowsMatch,
		"fast_q01_capacity":        checkpoint.FastQ01Capacity,
		"normalized_capacity":      checkpoint.NormalizedCapacity,
	}
}

func fix001P3DAR0Control(path evalModMatchedPath) map[string]interface{} {
	control := map[string]interface{}{"first_failure": path.FirstFailure, "round0_present": len(path.Rounds) > 0}
	if len(path.Rounds) == 0 {
		return control
	}
	round := path.Rounds[0]
	control["checkpoints"] = []map[string]interface{}{
		fix001P3DAR0CheckpointSummary(round.Input),
		fix001P3DAR0CheckpointSummary(round.Square),
		fix001P3DAR0CheckpointSummary(round.AfterMultiplier),
	}
	control["input_square_rows_match"] = round.Input.RowsMatch && round.Square.RowsMatch
	control["after_multiplier_rows_match"] = round.AfterMultiplier.RowsMatch
	control["after_multiplier_capacity_failure"] = !round.AfterMultiplier.NormalizedSafe
	control["reproduced"] = path.FirstFailure == "round0.after_multiplier_capacity" && round.Input.NormalizedSafe && round.Square.NormalizedSafe && !round.AfterMultiplier.NormalizedSafe && round.Input.RowsMatch && round.Square.RowsMatch && round.AfterMultiplier.RowsMatch
	return control
}

func fix001P3DAR0LaterSummary(path evalModMatchedPath) map[string]interface{} {
	result := map[string]interface{}{"first_failure": path.FirstFailure, "rounds": make([]map[string]interface{}, 0)}
	rounds := result["rounds"].([]map[string]interface{})
	for _, round := range path.Rounds {
		if round.Round == 0 {
			continue
		}
		checkpoints := make([]map[string]interface{}, 0, 5)
		for _, checkpoint := range []evalModMatchedCheckpoint{round.Input, round.Square, round.AfterMultiplier, round.AfterConstant, round.PostRescale} {
			if checkpoint.Name != "" {
				checkpoints = append(checkpoints, fix001P3DAR0CheckpointSummary(checkpoint))
			}
		}
		rounds = append(rounds, map[string]interface{}{"round": round.Round, "k_in_exponent": round.KInExponent, "k_out_exponent": round.KOutExponent, "multiplier_exponent": round.MultiplierExponent, "checkpoints": checkpoints})
	}
	result["rounds"] = rounds
	return result
}

func fix001P3DAR0PrepareState(fastInput, standardInput *rlwe.Ciphertext, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, planScale rlwe.Scale, fastPolyOverride *rlwe.Ciphertext) (fix001P3DAR0LocalQ2State, error) {
	state := fix001P3DAR0LocalQ2State{Path: evalModMatchedPath{Polynomial: map[string]*PSGlobalMetric{}, PolyMeta: map[string]string{}, FirstFailure: "none"}}
	coefficientView, err := targetScaleCoefficientView(params, fastInput)
	if err != nil {
		return state, err
	}
	inputData, err := targetScaleCenteredRows(params, coefficientView)
	if err != nil {
		return state, err
	}
	normalizedInput, err := normalizedFullFromLow(params, fastInput, inputData)
	if err != nil {
		return state, err
	}
	fastCurrent := fastInput.CopyNew()
	standardCurrent := standardInput.CopyNew()
	state.OriginalFastScale, state.OriginalStdScale = fastCurrent.Scale, standardCurrent.Scale
	addPrefix := func(name string, round int) error {
		checkpoint, e := evalModMatchedCheckpointFor(name, round, fastCurrent, normalizedInput, standardCurrent, params, fastSK, zeroSecret(params), standardSK)
		if e != nil {
			return e
		}
		state.Path.Prefix = append(state.Path.Prefix, checkpoint)
		if !checkpoint.RowsMatch && state.Path.FirstFailure == "none" {
			state.Path.FirstFailure = name
		}
		if !checkpoint.NormalizedSafe && state.Path.FirstFailure == "none" {
			state.Path.FirstFailure = name + ".normalized_capacity"
		}
		return nil
	}
	if err := addPrefix("evalmod_input", -1); err != nil {
		return state, err
	}
	mod1 := fastEval.Mod1Parameters
	fastCurrent.Scale = mod1.ScalingFactor()
	normalizedInput.Scale = mod1.ScalingFactor()
	standardCurrent.Scale = standardEval.Mod1Evaluator.Parameters.ScalingFactor()
	if err := addPrefix("normalize_scale", -1); err != nil {
		return state, err
	}
	offset := evalModCausalOffset(fastEval)
	if err := fastEval.FastCKKS.Add(fastCurrent, offset, fastCurrent); err != nil {
		return state, err
	}
	normalizedInput = evalModMatchedAddFullConstant(params, normalizedInput, float64FromBigFloat(offset))
	if err := standardEval.Evaluator.Add(standardCurrent, offset, standardCurrent); err != nil {
		return state, err
	}
	if err := addPrefix("apply_offset", -1); err != nil {
		return state, err
	}
	_, preprocessedValues, err := semanticBisectView(fastCurrent, params, fastSK)
	if err != nil {
		return state, err
	}
	state.Path.PlainOracle = chebyshevRawOracle(mod1.Mod1Poly, preprocessedValues)
	state.Path.PlainFinal = evalModMatchedPlainDoubleAngle(state.Path.PlainOracle, fastEval.Mod1Parameters.Sqrt2Pi, fastEval.Mod1Parameters.DoubleAngle)
	targetScale, err := evalModCausalTargetScale(fastCurrent, fastEval)
	if err != nil {
		return state, err
	}
	standardPoly, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardCurrent.CopyNew(), standardEval.Mod1Evaluator.Parameters.Mod1Poly.Clone(), targetScale)
	if err != nil {
		return state, err
	}
	var fastPoly *rlwe.Ciphertext
	if fastPolyOverride != nil {
		fastPoly = fastPolyOverride.CopyNew()
	} else {
		fastPoly, err = fastEval.PolynomialEvaluator.EvaluateWithPlanScale(fastCurrent.CopyNew(), fastEval.Mod1Parameters.Mod1Poly, targetScale, planScale)
		if err != nil {
			return state, err
		}
	}
	polyView, err := targetScaleCoefficientView(params, fastPoly)
	if err != nil {
		return state, err
	}
	polyData, err := targetScaleCenteredRows(params, polyView)
	if err != nil {
		return state, err
	}
	normalizedPoly, err := normalizedFullFromLow(params, fastPoly, polyData)
	if err != nil {
		return state, err
	}
	state.Path.PolyMeta["fast_level"] = fmt.Sprintf("%d", fastPoly.Level())
	state.Path.PolyMeta["normalized_level"] = fmt.Sprintf("%d", normalizedPoly.Level())
	state.Path.PolyMeta["standard_level"] = fmt.Sprintf("%d", standardPoly.Level())
	state.Path.PolyMeta["fast_scale"] = finalizationScaleString(fastPoly.Scale)
	state.Path.PolyMeta["normalized_scale"] = finalizationScaleString(normalizedPoly.Scale)
	state.Path.PolyMeta["standard_scale"] = finalizationScaleString(standardPoly.Scale)
	state.Path.PolyMeta["plan_scale"] = finalizationScaleString(planScale)
	state.Path.PolyMeta["target_scale"] = finalizationScaleString(targetScale)
	_, fastPolyValues, err := semanticBisectView(fastPoly, params, fastSK)
	if err != nil {
		return state, err
	}
	_, normalizedPolyValues, err := semanticBisectView(normalizedPoly, params, zeroSecret(params))
	if err != nil {
		return state, err
	}
	_, standardPolyValues, err := semanticBisectView(standardPoly, params, standardSK)
	if err != nil {
		return state, err
	}
	state.Path.Polynomial["fast_vs_normalized"] = evalModMatchedMetricFor(normalizedPolyValues, fastPolyValues, evalModMatchedStageThreshold)
	state.Path.Polynomial["normalized_vs_standard"] = evalModMatchedMetricFor(standardPolyValues, normalizedPolyValues, evalModMatchedLocalThreshold)
	state.Path.Polynomial["fast_vs_standard"] = evalModMatchedMetricFor(standardPolyValues, fastPolyValues, evalModMatchedLocalThreshold)
	state.Path.Polynomial["fast_vs_plaintext_oracle"] = evalModMatchedMetricFor(state.Path.PlainOracle, fastPolyValues, evalModMatchedLocalThreshold)
	state.Path.Polynomial["normalized_vs_plaintext_oracle"] = evalModMatchedMetricFor(state.Path.PlainOracle, normalizedPolyValues, evalModMatchedLocalThreshold)
	state.Path.Polynomial["standard_vs_plaintext_oracle"] = evalModMatchedMetricFor(state.Path.PlainOracle, standardPolyValues, evalModMatchedLocalThreshold)
	polyCapacity, err := normalizedExactCapacity(params, normalizedPoly)
	if err != nil {
		return state, err
	}
	state.Path.PolyMeta["normalized_capacity_pass"] = fmt.Sprintf("%t", polyCapacity.Pass)
	state.Path.PolyCapacity = &polyCapacity
	if !polyCapacity.Pass {
		state.Path.FirstFailure = "polynomial.normalized_capacity"
		return state, nil
	}
	workingScale := fastPoly.Scale
	kExponent, err := normalizedPowerOfTwoExponent(targetScale, workingScale)
	if err != nil {
		return state, err
	}
	if kExponent < 0 {
		return state, fmt.Errorf("negative initial normalized exponent %d", kExponent)
	}
	initialK := new(big.Int).Lsh(big.NewInt(1), uint(kExponent))
	state.FastNormalized = fastPoly.CopyNew()
	state.FastNormalized.Scale = state.FastNormalized.Scale.Mul(rlwe.NewScale(initialK))
	state.NormalizedCurrent = normalizedPoly.CopyNew()
	state.NormalizedCurrent.Scale = state.NormalizedCurrent.Scale.Mul(rlwe.NewScale(initialK))
	state.CoherentCurrent = normalizedFullIntegerMultiply(params, normalizedPoly, initialK)
	state.CoherentCurrent.Scale = normalizedPoly.Scale.Mul(rlwe.NewScale(initialK))
	state.StandardCurrent = standardPoly.CopyNew()
	state.WorkingScale = workingScale
	state.CurrentLevel = state.FastNormalized.Level()
	state.CurrentScale = state.FastNormalized.Scale
	state.KIn = kExponent
	state.Sqrt2Pi = fastEval.Mod1Parameters.Sqrt2Pi
	state.TargetScale = targetScale
	return state, nil
}

func fix001P3DAR0CapacityRow(params ckks.Parameters, ct *rlwe.Ciphertext) map[string]interface{} {
	row := map[string]interface{}{}
	if ct == nil {
		return row
	}
	if q01, err := fix001P3DAR0Capacity(params, ct, 1); err == nil {
		row["q01"] = q01
	}
	if ct.Level() >= 2 {
		if q012, err := fix001P3DAR0Capacity(params, ct, 2); err == nil {
			row["q012"] = q012
		}
	}
	return row
}

func fix001P3DAR0Round0Q2(state fix001P3DAR0LocalQ2State, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey) (fix001P3DAR0LocalQ2State, map[string]interface{}, map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {
	sqrt2pi := state.Sqrt2Pi * state.Sqrt2Pi
	schedule, err := normalizedScaleSchedule(0, state.CurrentLevel, state.CurrentScale, state.WorkingScale, state.KIn, sqrt2pi, params)
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	inputCheckpoint, err := evalModMatchedCheckpointFor("da_input", 0, state.FastNormalized, state.NormalizedCurrent, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	refProduct, err := normalizedFullSquare(params, state.NormalizedCurrent)
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	refSquare := normalizedFullZeroC2(params, refProduct)
	if err := fastEval.FastCKKS.MulRelin(state.FastNormalized, state.FastNormalized, state.FastNormalized); err != nil {
		return state, nil, nil, nil, nil, err
	}
	squareCheckpoint, err := evalModMatchedCheckpointFor("da_square", 0, state.FastNormalized, refSquare, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	signed, err := postMod1S2CRecoverQ01(params, state.FastNormalized)
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	expanded, err := fix001P3DAR0ExpandQ2(params, state.FastNormalized, signed)
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	semanticBefore, err := psGlobalDecode(params, state.FastNormalized)
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	semanticExpanded, err := fix001P3LocalQ2Decode(params, state.FastNormalized, signed, state.FastNormalized.Scale)
	if err != nil {
		return state, nil, nil, nil, nil, err
	}
	expansionMetric := psRescaleGuardMetric(semanticBefore, semanticExpanded)
	q2RowsMatch := fix001P3DAR0RowsAtLimb(params, expanded, refSquare, 2)
	q01RowsUnchanged := fix001P3DAR0RowsAtLimb(params, expanded, state.FastNormalized, 0) && fix001P3DAR0RowsAtLimb(params, expanded, state.FastNormalized, 1)
	expansion := map[string]interface{}{"input_q01_centered_unique": true, "q0_q1_rows_unchanged_bit_for_bit": q01RowsUnchanged, "q2_matches_stage_aligned_full_rns": q2RowsMatch, "semantic_change": expansionMetric, "pass": expansionMetric.Pass && q01RowsUnchanged && q2RowsMatch}
	if !expansionMetric.Pass || !q01RowsUnchanged || !q2RowsMatch {
		return state, expansion, nil, nil, nil, nil
	}
	factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
	refAfterMultiplier := normalizedFullIntegerMultiply(params, refSquare, factor)
	q01AfterMultiplier, err := fix001P3DAR0ContractQ01(params, refAfterMultiplier)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	checkpointMultiplier, err := evalModMatchedCheckpointFor("da_after_multiplier", 0, q01AfterMultiplier, refAfterMultiplier, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	constantValue, _ := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
	refAfterConstant := normalizedFullSubtractConstant(params, refAfterMultiplier, constantValue)
	q01AfterConstant, err := fix001P3DAR0ContractQ01(params, refAfterConstant)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	checkpointConstant, err := evalModMatchedCheckpointFor("da_after_constant", 0, q01AfterConstant, refAfterConstant, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	refAfterRescale, err := normalizedFullRescale(params, refAfterConstant)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	q01AfterRescale, err := fix001P3DAR0ContractQ01(params, refAfterRescale)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	q012AfterMultiplier, err := fix001P3DAR0Capacity(params, refAfterMultiplier, 2)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	q012AfterConstant, err := fix001P3DAR0Capacity(params, refAfterConstant, 2)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	q01BeforeMultiplier, err := fix001P3DAR0Capacity(params, refSquare, 1)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	q01AfterMultiplierCap, err := fix001P3DAR0Capacity(params, refAfterMultiplier, 1)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	q01AfterConstantCap, err := fix001P3DAR0Capacity(params, refAfterConstant, 1)
	if err != nil {
		return state, expansion, nil, nil, nil, err
	}
	r2 := map[string]interface{}{"before_multiplier_q01": q01BeforeMultiplier, "after_multiplier_q01": q01AfterMultiplierCap, "after_multiplier_q012": q012AfterMultiplier, "after_multiplier_rows_match": checkpointMultiplier.RowsMatch, "after_multiplier_q012_centered_unique": q012AfterMultiplier.Unique, "after_constant_q01": q01AfterConstantCap, "after_constant_q012": q012AfterConstant, "after_constant_rows_match": checkpointConstant.RowsMatch, "after_constant_q012_centered_unique": q012AfterConstant.Unique, "multiplier_exponent": schedule.AExponent, "k_in_exponent": schedule.KInExponent, "k_out_exponent": schedule.KOutExponent, "constant": constantValue}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[refAfterConstant.Level()].Modulus)
	roundedMatch, err := fix001P3DAR0RoundedDivisionMatches(params, refAfterConstant, refAfterRescale, divisor)
	if err != nil {
		return state, expansion, r2, nil, nil, err
	}
	q012Output, err := fix001P3DAR0Capacity(params, refAfterRescale, 2)
	if err != nil {
		return state, expansion, r2, nil, nil, err
	}
	q01Output, err := fix001P3DAR0Capacity(params, refAfterRescale, 1)
	if err != nil {
		return state, expansion, r2, nil, nil, err
	}
	r3 := map[string]interface{}{"divisor": divisor.String(), "input_level": refAfterConstant.Level(), "output_level": refAfterRescale.Level(), "input_scale": finalizationScaleString(refAfterConstant.Scale), "output_scale": finalizationScaleString(refAfterRescale.Scale), "q012_output": q012Output, "q01_output": q01Output, "rounded_division_match": roundedMatch, "q012_rows_match_full_rns": fix001P3DAR0RowsAtLimb(params, refAfterRescale, refAfterRescale, 2), "no_extra_level": refAfterRescale.Level() == refAfterConstant.Level()-params.LevelsConsumedPerRescaling()}
	contractDecoded, err := psGlobalDecode(params, q01AfterRescale)
	if err != nil {
		return state, expansion, r2, r3, nil, err
	}
	q012Decoded, err := fix001P3LocalQ2Decode(params, refAfterRescale, mustFIX001P3DAR0Signed(params, refAfterRescale, 2), refAfterRescale.Scale)
	if err != nil {
		return state, expansion, r2, r3, nil, err
	}
	contractMetric := psRescaleGuardMetric(q012Decoded, contractDecoded)
	contractUnique := q01Output.Unique
	contractRows := fix001P3DAR0RowsAtLimb(params, q01AfterRescale, refAfterRescale, 0) && fix001P3DAR0RowsAtLimb(params, q01AfterRescale, refAfterRescale, 1)
	contraction := map[string]interface{}{"q01_centered_unique": contractUnique, "q0_q1_rows_match_q012": contractRows, "semantic": contractMetric, "metadata_match": q01AfterRescale.Level() == refAfterRescale.Level() && q01AfterRescale.Scale.Equal(refAfterRescale.Scale) && q01AfterRescale.IsNTT && q01AfterRescale.IsMontgomery, "pass": contractUnique && contractRows && contractMetric.Pass}
	if !contraction["pass"].(bool) || !roundedMatch || !q012Output.Unique {
		return state, expansion, r2, r3, contraction, nil
	}
	postCheckpoint, err := evalModMatchedCheckpointFor("da_post_rescale", 0, q01AfterRescale, refAfterRescale, refAfterRescale, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	state.Path.Rounds = append(state.Path.Rounds, evalModMatchedRound{Round: 0, KInExponent: schedule.KInExponent, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, NormalizedConstant: constantValue, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: checkpointMultiplier, AfterConstant: checkpointConstant, PostRescale: postCheckpoint})
	state.FastNormalized = q01AfterRescale
	state.NormalizedCurrent = refAfterRescale
	coherentPre, _, err := normalizedFullSquareStep(params, state.CoherentCurrent, big.NewInt(2), sqrt2pi)
	if err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	state.CoherentCurrent, err = normalizedFullRescale(params, coherentPre)
	if err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	if err := standardEval.Evaluator.MulRelin(state.StandardCurrent, state.StandardCurrent, state.StandardCurrent); err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	if err := standardEval.Evaluator.Add(state.StandardCurrent, state.StandardCurrent, state.StandardCurrent); err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	if err := standardEval.Evaluator.Add(state.StandardCurrent, -sqrt2pi, state.StandardCurrent); err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	if err := standardEval.Evaluator.Rescale(state.StandardCurrent, state.StandardCurrent); err != nil {
		return state, expansion, r2, r3, contraction, err
	}
	state.Sqrt2Pi = sqrt2pi
	state.CurrentLevel = schedule.NextLevel
	state.CurrentScale = scheduleScaleValue(schedule, params, state.CurrentLevel, state.CurrentScale)
	state.KIn = schedule.KOutExponent
	return state, expansion, r2, r3, contraction, nil
}

func mustFIX001P3DAR0Signed(params ckks.Parameters, ct *rlwe.Ciphertext, level int) [][]*big.Int {
	values, err := fix001P3DAR0SignedValues(params, ct, level)
	if err != nil {
		return nil
	}
	return values
}

func fix001P3DAR0ContinueOrdinary(state *fix001P3DAR0LocalQ2State, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey) error {
	for round := 1; round < fastEval.Mod1Parameters.DoubleAngle; round++ {
		state.Sqrt2Pi *= state.Sqrt2Pi
		schedule, err := normalizedScaleSchedule(round, state.CurrentLevel, state.CurrentScale, state.WorkingScale, state.KIn, state.Sqrt2Pi, params)
		if err != nil {
			return err
		}
		inputCheckpoint, err := evalModMatchedCheckpointFor("da_input", round, state.FastNormalized, state.NormalizedCurrent, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return err
		}
		refProduct, err := normalizedFullSquare(params, state.NormalizedCurrent)
		if err != nil {
			return err
		}
		refSquare := normalizedFullZeroC2(params, refProduct)
		if err := fastEval.FastCKKS.MulRelin(state.FastNormalized, state.FastNormalized, state.FastNormalized); err != nil {
			return err
		}
		squareCheckpoint, err := evalModMatchedCheckpointFor("da_square", round, state.FastNormalized, refSquare, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return err
		}
		if !inputCheckpoint.NormalizedSafe || !squareCheckpoint.NormalizedSafe {
			state.Path.FirstFailure = fmt.Sprintf("round%d.%s_capacity", round, map[bool]string{true: "input", false: "square"}[!inputCheckpoint.NormalizedSafe])
			state.Path.Rounds = append(state.Path.Rounds, evalModMatchedRound{Round: round, KInExponent: schedule.KInExponent, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint})
			return nil
		}
		if !inputCheckpoint.RowsMatch || !squareCheckpoint.RowsMatch {
			state.Path.FirstFailure = fmt.Sprintf("round%d.rows", round)
		}
		factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
		if err := fastEval.FastCKKS.MulIntegerMaintained(state.FastNormalized, factor, state.FastNormalized); err != nil {
			return err
		}
		refAfterA := normalizedFullIntegerMultiply(params, refSquare, factor)
		aCheckpoint, err := evalModMatchedCheckpointFor("da_after_multiplier", round, state.FastNormalized, refAfterA, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return err
		}
		if !aCheckpoint.NormalizedSafe {
			state.Path.FirstFailure = fmt.Sprintf("round%d.after_multiplier_capacity", round)
			state.Path.Rounds = append(state.Path.Rounds, evalModMatchedRound{Round: round, KInExponent: schedule.KInExponent, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint})
			return nil
		}
		constantValue, _ := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetFloat64(state.Sqrt2Pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
		if err := fastEval.FastCKKS.Add(state.FastNormalized, -constantValue, state.FastNormalized); err != nil {
			return err
		}
		refAfterConstant := normalizedFullSubtractConstant(params, refAfterA, constantValue)
		constantCheckpoint, err := evalModMatchedCheckpointFor("da_after_constant", round, state.FastNormalized, refAfterConstant, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return err
		}
		if !constantCheckpoint.NormalizedSafe {
			state.Path.FirstFailure = fmt.Sprintf("round%d.constant_capacity", round)
			state.Path.Rounds = append(state.Path.Rounds, evalModMatchedRound{Round: round, KInExponent: schedule.KInExponent, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint, AfterConstant: constantCheckpoint})
			return nil
		}
		if err := fastEval.FastCKKS.Rescale(state.FastNormalized, state.FastNormalized); err != nil {
			return err
		}
		normalizedCurrent, err := normalizedFullRescale(params, refAfterConstant)
		if err != nil {
			return err
		}
		coherentPre, _, err := normalizedFullSquareStep(params, state.CoherentCurrent, big.NewInt(2), state.Sqrt2Pi)
		if err != nil {
			return err
		}
		coherentCurrent, err := normalizedFullRescale(params, coherentPre)
		if err != nil {
			return err
		}
		if err := standardEval.Evaluator.MulRelin(state.StandardCurrent, state.StandardCurrent, state.StandardCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.Add(state.StandardCurrent, state.StandardCurrent, state.StandardCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.Add(state.StandardCurrent, -state.Sqrt2Pi, state.StandardCurrent); err != nil {
			return err
		}
		if err := standardEval.Evaluator.Rescale(state.StandardCurrent, state.StandardCurrent); err != nil {
			return err
		}
		postCheckpoint, err := evalModMatchedCheckpointFor("da_post_rescale", round, state.FastNormalized, normalizedCurrent, coherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
		if err != nil {
			return err
		}
		state.Path.Rounds = append(state.Path.Rounds, evalModMatchedRound{Round: round, KInExponent: schedule.KInExponent, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, NormalizedConstant: constantValue, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: aCheckpoint, AfterConstant: constantCheckpoint, PostRescale: postCheckpoint})
		if !postCheckpoint.NormalizedSafe {
			state.Path.FirstFailure = fmt.Sprintf("round%d.post_rescale_capacity", round)
			return nil
		}
		state.NormalizedCurrent, state.CoherentCurrent = normalizedCurrent, coherentCurrent
		state.CurrentLevel = schedule.NextLevel
		state.CurrentScale = scheduleScaleValue(schedule, params, state.CurrentLevel, state.CurrentScale)
		state.KIn = schedule.KOutExponent
	}
	fastBeforeRestore := state.FastNormalized.CopyNew()
	normalizedBeforeRestore := state.NormalizedCurrent.CopyNew()
	coherentBeforeReset := state.CoherentCurrent.CopyNew()
	standardBeforeReset := state.StandardCurrent.CopyNew()
	finalK := new(big.Int).Lsh(big.NewInt(1), uint(state.KIn))
	beforeRestoreCapacity, err := normalizedExactCapacity(params, state.NormalizedCurrent)
	if err != nil {
		return err
	}
	restoreView, err := targetScaleCoefficientView(params, fastBeforeRestore)
	if err != nil {
		return err
	}
	restoreData, err := targetScaleCenteredRows(params, restoreView)
	if err != nil {
		return err
	}
	restoreCapacity := normalizedFinalRestoreCapacity(restoreData, finalK)
	state.Path.NormalizedBeforeRestoreCapacity = &beforeRestoreCapacity
	state.Path.FinalRestoreCapacity = &restoreCapacity
	if !beforeRestoreCapacity.Pass || !restoreCapacity.Pass {
		state.Path.FirstFailure = "final.restore_capacity"
		return nil
	}
	if err := fastEval.FastCKKS.MulIntegerMaintained(state.FastNormalized, finalK, state.FastNormalized); err != nil {
		return err
	}
	state.NormalizedCurrent = normalizedFullIntegerMultiply(params, state.NormalizedCurrent, finalK)
	fastRestoreRows := finalizationEvidence(state.FastNormalized).Rows
	normalizedRestoreRows := finalizationEvidence(state.NormalizedCurrent).Rows
	state.Path.FastRowsMatch = doubleAngleRowsMatch(fastRestoreRows, normalizedRestoreRows)
	state.Path.RestoreRowsMatch = state.Path.FastRowsMatch
	state.FastNormalized.Scale = state.OriginalFastScale
	state.NormalizedCurrent.Scale = state.OriginalFastScale
	state.StandardCurrent.Scale = state.OriginalStdScale
	state.CoherentCurrent.Scale = state.OriginalStdScale
	state.Path.ScaleResetRowsUnchanged = doubleAngleRowsMatch(fastRestoreRows, finalizationEvidence(state.FastNormalized).Rows) && doubleAngleRowsMatch(normalizedRestoreRows, finalizationEvidence(state.NormalizedCurrent).Rows)
	state.Path.LevelDegreeUnchanged = state.FastNormalized.Level() == fastBeforeRestore.Level() && state.FastNormalized.Degree() == fastBeforeRestore.Degree() && state.NormalizedCurrent.Level() == normalizedBeforeRestore.Level() && state.NormalizedCurrent.Degree() == normalizedBeforeRestore.Degree()
	state.Path.FastFinal, state.Path.NormalizedFinal, state.Path.CoherentFinal, state.Path.StandardFinal = state.FastNormalized, state.NormalizedCurrent, state.CoherentCurrent, state.StandardCurrent
	_ = coherentBeforeReset
	_ = standardBeforeReset
	return nil
}

func runFIX001P3DesignLogN13DAR0LocalQ2Feasibility(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3DAR0LocalQ2PrimaryBase, "HEAD") == requiredFIX001P3DAR0LocalQ2PrimaryBase
	result := fix001P3DAR0LocalQ2Result{
		SchemaVersion: "fix-001-p3-design-logn13-da-r0-local-q2-feasibility.v1",
		Timestamp:     time.Now().UTC(),
		Primary:       gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Operations: fix001P3DAR0LocalQ2Operations{CoefficientTransforms: 2, Q2NTT: 2, Q2INTT: 0, Q2ArithmeticPasses: 6, RescaleBoundaries: 2, ActiveRegion: "DA round0 after-square through round0 post-Rescale only"},
		Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3DAR0LocalQ2PrimaryBase, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_commit": secondaryCommit, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3DAR0LocalQ2Secondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "logn13_only": cfg.LogN == 13, "fixed_k3_candidate": true, "q0_q1_q2_profile": "56/39/40", "normalized_schedule_fixed": true, "q2_local_only_at_da_r0": true, "no_q0_q1_widening": true, "no_q3_plus": true, "no_extra_q_levels": true, "no_ps_retuning": true, "no_generated_power_redesign": true, "no_c2s_s2c_mod1_change": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
	}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3DAR0LocalQ2Secondary {
		result.Classification, result.FirstRemainingBlock = "logn13_da_r0_local_q2_precondition_mismatch", "startup provenance or Secondary precondition"
		return fix001P3DAR0LocalQ2Write(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	var realEvidence, imagEvidence fix001P3LocalQ2BranchEvidence
	candidate, _, _, err := psRescaleGuardMakeCandidateWithOverrides(profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch, profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, 3}, "D0-K3-G0-one-local-q2-F0-three", &profile.InitialReal, &profile.InitialImag, profile.Boundaries, fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &realEvidence, 3), fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &imagEvidence, 3))
	if err != nil {
		return err
	}
	planScale := precisionSweepScale(92)
	realControlPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return err
	}
	imagControlPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return err
	}
	result.R0Control = map[string]interface{}{"real": fix001P3DAR0Control(realControlPath), "imag": fix001P3DAR0Control(imagControlPath), "fixed_candidate": q056CompactCandidate(candidate), "accepted_real_error": 1.223639867209414e-8, "accepted_imag_error": 1.1510743691545144e-8}
	if !fix001P3DAR0Control(realControlPath)["reproduced"].(bool) || !fix001P3DAR0Control(imagControlPath)["reproduced"].(bool) {
		result.Classification, result.FirstRemainingBlock = "logn13_da_r0_local_q2_precondition_mismatch", "R0 accepted after-multiplier capacity failure was not reproduced"
		return fix001P3DAR0LocalQ2Write(result, outPath)
	}
	realState, err := fix001P3DAR0PrepareState(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return err
	}
	imagState, err := fix001P3DAR0PrepareState(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return err
	}
	var realExpansion, imagExpansion, realR2, imagR2, realR3, imagR3, realContract, imagContract map[string]interface{}
	realState, realExpansion, realR2, realR3, realContract, err = fix001P3DAR0Round0Q2(realState, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagState, imagExpansion, imagR2, imagR3, imagContract, err = fix001P3DAR0Round0Q2(imagState, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	result.Expansion = map[string]interface{}{"real": realExpansion, "imag": imagExpansion}
	result.R2Capacity = map[string]interface{}{"real": realR2, "imag": imagR2}
	result.R3Rescale = map[string]interface{}{"real": realR3, "imag": imagR3}
	result.Contraction = map[string]interface{}{"real": realContract, "imag": imagContract}
	if realContract == nil || imagContract == nil || !realContract["pass"].(bool) || !imagContract["pass"].(bool) {
		result.Classification, result.FirstRemainingBlock = "logn13_da_r0_local_q2_cannot_contract_to_q01", "R4 post-round0 q01 contraction"
		return fix001P3DAR0LocalQ2Write(result, outPath)
	}
	if err := fix001P3DAR0ContinueOrdinary(&realState, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK); err != nil {
		return err
	}
	if err := fix001P3DAR0ContinueOrdinary(&imagState, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK); err != nil {
		return err
	}
	result.LaterDA = []map[string]interface{}{{"real": fix001P3DAR0LaterSummary(realState.Path), "imag": fix001P3DAR0LaterSummary(imagState.Path)}}
	if realState.Path.FastFinal == nil || imagState.Path.FastFinal == nil || realState.Path.FirstFailure != "none" || imagState.Path.FirstFailure != "none" || !realState.Path.FastRowsMatch || !realState.Path.RestoreRowsMatch || !imagState.Path.FastRowsMatch || !imagState.Path.RestoreRowsMatch {
		result.Classification, result.FirstRemainingBlock = "logn13_da_r0_local_q2_later_da_capacity_blocker", fmt.Sprintf("real=%s imag=%s", realState.Path.FirstFailure, imagState.Path.FirstFailure)
		return fix001P3DAR0LocalQ2Write(result, outPath)
	}
	realFinal, err := evalModMatchedFinalEvidence(realState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realState.Path, imagState.Path, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	result.Final = map[string]interface{}{"evalmod_real_vs_standard": realFinal.FastVsStandard, "evalmod_imag_vs_standard": imagFinal.FastVsStandard, "post_s2c_fast_vs_standard": publicEvidence.PostS2C, "unpack_finalization_fast_vs_standard_core": publicEvidence.FinalVsStdCore, "standard_like": publicEvidence.StandardLike, "public_like": publicEvidence.PublicLike, "final_metadata": publicEvidence.Metadata, "input_unchanged": publicEvidence.InputUnchanged}
	result.SystemSufficient = publicEvidence.PublicLike != nil && publicEvidence.PublicLike.Pass && publicEvidence.Metadata["pass"] == true && publicEvidence.InputUnchanged && realFinal.FastVsStandard.Pass && imagFinal.FastVsStandard.Pass
	if result.SystemSufficient {
		result.Classification, result.FirstRemainingBlock = "logn13_da_r0_local_q2_downstream_validated", "none"
	} else {
		result.Classification, result.FirstRemainingBlock = "logn13_da_r0_local_q2_downstream_insufficient", "DA completed but public-like or semantic contract failed"
	}
	return fix001P3DAR0LocalQ2Write(result, outPath)
}
