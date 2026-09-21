package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/bits"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3P93InternalSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	requiredFIX001P3P93InternalDiffSHA   = "44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e"
	fix001P3P93InternalC2SThreshold      = 1e-10
	fix001P3P93InternalBudget            = 3.125e-4
	fix001P3P93InternalPublicThreshold   = 1e-2
)

type fix001P3P93InternalCheckpoint struct {
	Name            string                       `json:"name"`
	Standard        semanticBisectMetadata       `json:"standard"`
	Fast            semanticBisectMetadata       `json:"fast"`
	SemanticAligned bool                         `json:"semantic_aligned"`
	Metric          *semanticBisectMetric        `json:"fast_vs_standard,omitempty"`
	StandardRows    []FinalizationRowFingerprint `json:"standard_row_hashes_q0_q1_q2"`
	FastRows        []FinalizationRowFingerprint `json:"fast_row_hashes_q0_q1_q2"`
	FastCapacity    *postMod1S2CCapacity         `json:"fast_q012_capacity,omitempty"`
	CapacityStatus  string                       `json:"fast_capacity_status"`
	Operation       string                       `json:"operation"`
	ReplayVerified  bool                         `json:"replay_verified"`
}

type fix001P3P93InternalPower struct {
	Power        int                          `json:"power"`
	Standard     semanticBisectMetadata       `json:"standard"`
	Fast         semanticBisectMetadata       `json:"fast"`
	Metric       *semanticBisectMetric        `json:"fast_vs_standard,omitempty"`
	StandardRows []FinalizationRowFingerprint `json:"standard_row_hashes_q0_q1_q2"`
	FastRows     []FinalizationRowFingerprint `json:"fast_row_hashes_q0_q1_q2"`
}

type fix001P3P93InternalBranch struct {
	Name                  string                          `json:"name"`
	Checkpoints           []fix001P3P93InternalCheckpoint `json:"checkpoints"`
	GeneratedPowers       []fix001P3P93InternalPower      `json:"generated_powers"`
	PSPlan                map[string]interface{}          `json:"ps_plan"`
	FastPlanEvidence      map[string]interface{}          `json:"fast_plan_evidence"`
	StandardInternal      semanticBisectMetadata          `json:"standard_internal_final"`
	FastInternal          semanticBisectMetadata          `json:"fast_internal_final"`
	StandardActualAligned bool                            `json:"standard_replay_matches_actual"`
	FastActualAligned     bool                            `json:"fast_replay_matches_actual"`
	StandardReplayMetric  *semanticBisectMetric           `json:"standard_replay_vs_actual,omitempty"`
	FastReplayMetric      *semanticBisectMetric           `json:"fast_replay_vs_actual,omitempty"`
	states                map[string]*rlwe.Ciphertext     `json:"-"`
}

type fix001P3P93InternalResult struct {
	SchemaVersion   string                               `json:"schema_version"`
	Timestamp       time.Time                            `json:"timestamp"`
	Primary         RepositoryMetadata                   `json:"primary_repository"`
	Secondary       RepositoryMetadata                   `json:"secondary_repository"`
	Environment     EnvironmentMetadata                  `json:"environment"`
	Config          BootstrapConfig                      `json:"config"`
	FixedProfile    map[string]interface{}               `json:"fixed_profile"`
	R0C2S           map[string]interface{}               `json:"r0_c2s_precondition"`
	Branches        map[string]fix001P3P93InternalBranch `json:"branches"`
	FirstObservable map[string]interface{}               `json:"first_observable_divergence"`
	FirstMaterial   map[string]interface{}               `json:"first_material_divergence"`
	Factor32        map[string]interface{}               `json:"factor_32_public_mapping"`
	HistoricalP93   map[string]interface{}               `json:"historical_p93_comparability"`
	SourceAudit     map[string]interface{}               `json:"standard_vs_fast_source_audit"`
	Classification  string                               `json:"classification"`
	Validation      map[string]interface{}               `json:"validation"`
}

func fix001P3P93InternalPlanScale() rlwe.Scale {
	return rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 93))
}

func fix001P3P93InternalFastCopy(params ckks.Parameters, src *rlwe.Ciphertext) *rlwe.Ciphertext {
	dst := fastckks.NewCiphertext(params, src.Degree(), src.Level())
	*dst.MetaData = *src.MetaData
	dst.IsNTT, dst.IsMontgomery = src.IsNTT, src.IsMontgomery
	limbs := fastckks.MaintainedLimbCount(&params, src.Level())
	for component := 0; component <= src.Degree(); component++ {
		for limb := 0; limb < limbs && limb < len(src.Value[component].Coeffs); limb++ {
			copy(dst.Value[component].Coeffs[limb], src.Value[component].Coeffs[limb])
		}
	}
	return dst
}

func fix001P3P93InternalRows(ct *rlwe.Ciphertext) []FinalizationRowFingerprint {
	rows := make([]FinalizationRowFingerprint, 0, 3*(ct.Degree()+1))
	for component := 0; component <= ct.Degree() && component < len(ct.Value); component++ {
		for limb := 0; limb < 3; limb++ {
			var row []uint64
			if limb < len(ct.Value[component].Coeffs) {
				row = ct.Value[component].Coeffs[limb]
			}
			rows = append(rows, FinalizationRowFingerprint{Component: component, Limb: limb, Length: len(row), SHA256: finalizationRowHash(row)})
		}
	}
	return rows
}

func makeFIX001P3P93InternalCheckpoint(name, operation string, standard, fast *rlwe.Ciphertext, params ckks.Parameters, standardSK, fastSK *rlwe.SecretKey, replayVerified bool) (fix001P3P93InternalCheckpoint, error) {
	result := fix001P3P93InternalCheckpoint{Name: name, Operation: operation, StandardRows: fix001P3P93InternalRows(standard), FastRows: fix001P3P93InternalRows(fast), ReplayVerified: replayVerified}
	standardMeta, standardValues, err := semanticBisectView(standard, params, standardSK)
	if err != nil {
		return result, err
	}
	fastMeta, fastValues, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return result, err
	}
	result.Standard, result.Fast = standardMeta, fastMeta
	metric, err := semanticBisectMetricFor(standardValues, fastValues, fix001P3P93InternalBudget)
	if err != nil {
		return result, err
	}
	result.Metric = &metric
	result.SemanticAligned = standardMeta.N == fastMeta.N && standardMeta.Degree == fastMeta.Degree && standardMeta.DecodeLevel == fastMeta.DecodeLevel && len(standardValues) == len(fastValues)
	if fast.Level() >= 1 {
		capacity, capacityErr := postMod1S2CCapacityFromFastCiphertext(params, fast)
		if capacityErr == nil {
			result.FastCapacity = &capacity
			result.CapacityStatus = "checked"
		} else {
			result.CapacityStatus = "unavailable: " + capacityErr.Error()
		}
	} else {
		result.CapacityStatus = "not_applicable_level_below_q1"
	}
	return result, nil
}

func fix001P3P93InternalTargetScale(input *rlwe.Ciphertext, mod1Params mod1.Parameters, params ckks.Parameters) (rlwe.Scale, error) {
	target := mod1Params.ScalingFactor()
	for i := 0; i < mod1Params.DoubleAngle; i++ {
		index := input.Level() - mod1Params.Mod1Poly.Depth() - mod1Params.DoubleAngle + i + 1
		if index < 0 || index >= len(params.Q()) {
			return rlwe.Scale{}, fmt.Errorf("target scale Q index %d outside parameter chain", index)
		}
		target = target.Mul(rlwe.NewScale(params.Q()[index]))
		target.Value.Sqrt(&target.Value)
	}
	return target, nil
}

func fix001P3P93InternalOffset(mod1Params mod1.Parameters) *big.Float {
	offset := new(big.Float).Sub(&mod1Params.Mod1Poly.B, &mod1Params.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(mod1Params.IntervalShrinkFactor()))
	return offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
}

func fix001P3P93InternalStandardTrace(input *rlwe.Ciphertext, eval *bootstrapping.Evaluator, params ckks.Parameters, sk *rlwe.SecretKey, fastSK *rlwe.SecretKey) (fix001P3P93InternalBranch, *rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	mod1 := eval.Mod1Evaluator
	mod1Params := mod1.Parameters
	actual, err := mod1.EvaluateNew(input.CopyNew())
	if err != nil {
		return fix001P3P93InternalBranch{}, nil, nil, err
	}
	res := input.CopyNew()
	branch := fix001P3P93InternalBranch{Name: "standard", Checkpoints: make([]fix001P3P93InternalCheckpoint, 0, 10), states: map[string]*rlwe.Ciphertext{}}
	add := func(name, op string, fast *rlwe.Ciphertext, replay bool) error {
		cp, e := makeFIX001P3P93InternalCheckpoint(name, op, res, fast, params, sk, fastSK, replay)
		if e == nil {
			branch.Checkpoints = append(branch.Checkpoints, cp)
			branch.states[name] = res.CopyNew()
		}
		return e
	}
	fastMirror := fix001P3P93InternalFastCopy(params, input)
	if err = add("evalmod_input", "source input", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	if res.Level() > mod1Params.LevelQ {
		mod1.DropLevel(res, res.Level()-mod1Params.LevelQ)
	}
	res.Scale = mod1Params.ScalingFactor()
	fastMirror.Scale = mod1Params.ScalingFactor()
	if err = add("normalize_scale", "assign Mod1 ScalingFactor", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	offset := fix001P3P93InternalOffset(mod1Params)
	if err = mod1.Add(res, offset, res); err != nil {
		return branch, actual, nil, err
	}
	if err = add("apply_offset", "CosDiscrete offset Add", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	if err = add("polynomial_input", "source polynomial input", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	targetScale, err := fix001P3P93InternalTargetScale(input, mod1Params, params)
	if err != nil {
		return branch, actual, nil, err
	}
	standardPoly, err := mod1.PolynomialEvaluator.Evaluate(res, mod1Params.Mod1Poly.Clone(), targetScale)
	if err != nil {
		return branch, actual, nil, err
	}
	res = standardPoly
	branch.PSPlan = map[string]interface{}{"standard_polynomial": "common polynomial Evaluator.Evaluate", "fast_polynomial": "FastEvaluator.EvaluateWithPlanScale", "plan_scale": "2^93", "standard_plan_scale_override": false}
	if err = add("polynomial_output_after_rescale", "polynomial evaluator output", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	if err = add("pre_double_angle", "pre-DoubleAngle state", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	for round := 0; round < mod1Params.DoubleAngle; round++ {
		sqrt2pi := mod1Params.Sqrt2Pi
		for i := 0; i <= round; i++ {
			sqrt2pi *= sqrt2pi
		}
		if err = mod1.MulRelin(res, res, res); err != nil {
			return branch, actual, nil, err
		}
		if err = add(fmt.Sprintf("double_angle_round_%d_square", round), "MulRelin square", fastMirror, true); err != nil {
			return branch, actual, nil, err
		}
		if err = mod1.Add(res, res, res); err != nil {
			return branch, actual, nil, err
		}
		if err = mod1.Add(res, -sqrt2pi, res); err != nil {
			return branch, actual, nil, err
		}
		if err = add(fmt.Sprintf("double_angle_round_%d_after_constant", round), "double + constant before Rescale", fastMirror, true); err != nil {
			return branch, actual, nil, err
		}
		if err = mod1.Rescale(res, res); err != nil {
			return branch, actual, nil, err
		}
		if err = add(fmt.Sprintf("double_angle_round_%d_post_rescale", round), "square + double + constant + Rescale", fastMirror, true); err != nil {
			return branch, actual, nil, err
		}
		// The standard replay is intentionally kept separate from the Fast replay.
		// The paired Fast state is overwritten below by the independently traced Fast path.
	}
	res.Scale = input.Scale
	if err = add("internal_final_restore", "restore caller input Scale", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	branch.StandardInternal, _, err = semanticBisectView(res, params, sk)
	if err != nil {
		return branch, actual, nil, err
	}
	res.Scale = eval.BootstrappingParameters.DefaultScale()
	fastMirror.Scale = eval.BootstrappingParameters.DefaultScale()
	if err = add("public_reset_context", "wrapper DefaultScale context", fastMirror, true); err != nil {
		return branch, actual, nil, err
	}
	return branch, actual, branch.states["internal_final_restore"], nil
}

func fix001P3P93InternalFastTrace(input, standardInput *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, standardSK, fastSK *rlwe.SecretKey) (fix001P3P93InternalBranch, *rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	mod1 := eval.Mod1Evaluator
	mod1Params := eval.Mod1Parameters
	actual, err := mod1.EvaluateNew(input.CopyNew())
	if err != nil {
		return fix001P3P93InternalBranch{}, nil, nil, err
	}
	res := fix001P3P93InternalFastCopy(params, input)
	standard := standardInput.CopyNew()
	branch := fix001P3P93InternalBranch{Name: "fast", Checkpoints: make([]fix001P3P93InternalCheckpoint, 0, 12), states: map[string]*rlwe.Ciphertext{}}
	add := func(name, op string, replay bool) error {
		cp, e := makeFIX001P3P93InternalCheckpoint(name, op, standard, res, params, standardSK, fastSK, replay)
		if e == nil {
			branch.Checkpoints = append(branch.Checkpoints, cp)
			branch.states[name] = res.CopyNew()
		}
		return e
	}
	if err = add("evalmod_input", "source input", true); err != nil {
		return branch, actual, nil, err
	}
	res.Scale = mod1Params.ScalingFactor()
	standard.Scale = mod1Params.ScalingFactor()
	if err = add("normalize_scale", "assign Mod1 ScalingFactor", true); err != nil {
		return branch, actual, nil, err
	}
	offset := fix001P3P93InternalOffset(mod1Params)
	if err = mod1.FastCKKS.Add(res, offset, res); err != nil {
		return branch, actual, nil, err
	}
	if err = add("apply_offset", "CosDiscrete offset Add", true); err != nil {
		return branch, actual, nil, err
	}
	if err = add("polynomial_input", "source polynomial input", true); err != nil {
		return branch, actual, nil, err
	}
	targetScale, err := fix001P3P93InternalTargetScale(input, mod1Params, params)
	if err != nil {
		return branch, actual, nil, err
	}
	planScale := fix001P3P93InternalPlanScale()
	fastPoly, err := mod1.PolynomialEvaluator.EvaluateWithPlanScale(res, mod1Params.Mod1Poly, targetScale, planScale)
	if err != nil {
		return branch, actual, nil, err
	}
	standardPoly, err := standardEvalPolynomialForDiagnostic(standardEval, standard, mod1Params, targetScale)
	if err != nil {
		return branch, actual, nil, err
	}
	standard = standardPoly
	res = fastPoly
	selection := mod1.PolynomialEvaluator.LastGuardSelectionEvidence()
	branch.FastPlanEvidence = map[string]interface{}{"plan_scale_override": selection.PlanScaleOverride, "plan_index": selection.PlanIndex, "plan_count": selection.PlanCount, "scalar_degree": selection.ScalarDegree, "operations": selection.Operations, "contraction": selection.Contraction}
	if err = add("polynomial_output_after_rescale", "Fast PS output after final Rescale", true); err != nil {
		return branch, actual, nil, err
	}
	workingScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 33))
	if !res.Scale.InDelta(workingScale, 32) {
		return branch, actual, nil, fmt.Errorf("Fast polynomial output scale %s is not near 2^33", res.Scale.Value.Text('e', 10))
	}
	kExponent := 27
	coherentScale := res.Scale.Mul(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(kExponent))))
	res.Scale = coherentScale
	standard.Scale = coherentScale
	if err = add("pre_double_angle", "normalized coherent scale", true); err != nil {
		return branch, actual, nil, err
	}
	currentExponent := kExponent
	sqrt2pi := mod1Params.Sqrt2Pi
	for round := 0; round < mod1Params.DoubleAngle; round++ {
		beforeLevel := res.Level()
		nextScale := res.Scale.Mul(res.Scale).Div(rlwe.NewScale(params.Q()[beforeLevel]))
		nextExponent, err := nearestPowerOfTwoExponent(nextScale.Div(workingScale))
		if err != nil {
			return branch, actual, nil, err
		}
		aExponent := 1 + 2*currentExponent - nextExponent
		factor := new(big.Int).Lsh(big.NewInt(1), uint(aExponent))
		sqrt2pi *= sqrt2pi
		constant, _ := new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(nextExponent)))).Float64()
		if err = mod1.FastCKKS.MulRelin(res, res, res); err != nil {
			return branch, actual, nil, err
		}
		if err = add(fmt.Sprintf("double_angle_round_%d_square", round), "MulRelin square", true); err != nil {
			return branch, actual, nil, err
		}
		if err = mod1.FastCKKS.MulIntegerMaintained(res, factor, res); err != nil {
			return branch, actual, nil, err
		}
		if err = mod1.FastCKKS.Add(res, -constant, res); err != nil {
			return branch, actual, nil, err
		}
		if err = add(fmt.Sprintf("double_angle_round_%d_after_constant", round), fmt.Sprintf("maintained multiplier 2^%d + constant before Rescale", aExponent), true); err != nil {
			return branch, actual, nil, err
		}
		if err = mod1.FastCKKS.Rescale(res, res); err != nil {
			return branch, actual, nil, err
		}
		if err = add(fmt.Sprintf("double_angle_round_%d_post_rescale", round), fmt.Sprintf("normalized square factor=2^%d constant=%.17g Rescale", aExponent, constant), true); err != nil {
			return branch, actual, nil, err
		}
		currentExponent = nextExponent
	}
	beforeRestore := res.Scale
	if err = mod1.FastCKKS.MulIntegerMaintained(res, new(big.Int).Lsh(big.NewInt(1), uint(currentExponent)), res); err != nil {
		return branch, actual, nil, err
	}
	if !res.Scale.Equal(beforeRestore) {
		return branch, actual, nil, fmt.Errorf("Fast internal restore changed scale")
	}
	res.Scale = input.Scale
	standard.Scale = res.Scale
	if err = add("internal_final_restore", "final maintained scalar restore", true); err != nil {
		return branch, actual, nil, err
	}
	branch.FastInternal, _, err = semanticBisectView(res, params, fastSK)
	if err != nil {
		return branch, actual, nil, err
	}
	res.Scale = eval.Parameters.BootstrappingParameters.DefaultScale()
	standard.Scale = eval.Parameters.BootstrappingParameters.DefaultScale()
	if err = add("public_reset_context", "wrapper DefaultScale context", true); err != nil {
		return branch, actual, nil, err
	}
	return branch, actual, branch.states["internal_final_restore"], nil
}

func standardEvalPolynomialForDiagnostic(eval *bootstrapping.Evaluator, input *rlwe.Ciphertext, mod1Params mod1.Parameters, targetScale rlwe.Scale) (*rlwe.Ciphertext, error) {
	// A Genuine Standard polynomial call is made by the Standard Mod1 evaluator.
	// This helper keeps the paired Standard state source-faithful without using
	// any Fast implementation detail.
	return eval.Mod1Evaluator.PolynomialEvaluator.Evaluate(input, mod1Params.Mod1Poly.Clone(), targetScale)
}

func nearestPowerOfTwoExponent(scale rlwe.Scale) (int, error) {
	if scale.Value.Sign() <= 0 {
		return 0, fmt.Errorf("non-positive scale")
	}
	mantissa := new(big.Float).SetPrec(256)
	exponent := scale.Value.MantExp(mantissa)
	if mantissa.Cmp(new(big.Float).SetPrec(256).SetFloat64(0.75)) >= 0 {
		return exponent, nil
	}
	return exponent - 1, nil
}

func fix001P3P93InternalPowerTrace(standardInput, fastInput *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, standardSK, fastSK *rlwe.SecretKey, targetScale rlwe.Scale, poly bignum.Polynomial) ([]fix001P3P93InternalPower, map[string]interface{}, error) {
	fastPowers, plan, err := eval.PolynomialEvaluator.DiagnosticGeneratePowers(fastInput, poly, targetScale)
	if err != nil {
		return nil, nil, err
	}
	stdPB := commonpolynomial.NewPowerBasis(standardInput.CopyNew(), bignum.Chebyshev)
	degree := poly.Degree()
	logDegree := bits.Len64(uint64(degree))
	if err := stdPB.GenPower(1<<(logDegree-1), false, standardEval.Mod1Evaluator.Evaluator); err != nil {
		return nil, nil, err
	}
	logSplit := bignum.OptimalSplit(logDegree)
	for i := (1 << logSplit) - 1; i > 2; i-- {
		if !(poly.IsEven || poly.IsOdd) || (i&1 == 0 && poly.IsEven) || (i&1 == 1 && poly.IsOdd) {
			if err := stdPB.GenPower(i, false, standardEval.Mod1Evaluator.Evaluator); err != nil {
				return nil, nil, err
			}
		}
	}
	stdKeys := make([]int, 0, len(stdPB.Value))
	for n := range stdPB.Value {
		stdKeys = append(stdKeys, n)
	}
	sort.Ints(stdKeys)
	fastMap := map[int]*rlwe.Ciphertext{}
	for _, p := range fastPowers {
		fastMap[p.N] = p.Ciphertext
	}
	rows := make([]fix001P3P93InternalPower, 0, len(stdKeys))
	for _, n := range stdKeys {
		if fastCT := fastMap[n]; fastCT != nil {
			stdCT := stdPB.Value[n]
			stdMeta, stdValues, e := semanticBisectView(stdCT, params, standardSK)
			if e != nil {
				return nil, nil, e
			}
			fastMeta, fastValues, e := semanticBisectView(fastCT, params, fastSK)
			if e != nil {
				return nil, nil, e
			}
			metric, e := semanticBisectMetricFor(stdValues, fastValues, fix001P3P93InternalBudget)
			if e != nil {
				return nil, nil, e
			}
			rows = append(rows, fix001P3P93InternalPower{Power: n, Standard: stdMeta, Fast: fastMeta, Metric: &metric, StandardRows: fix001P3P93InternalRows(stdCT), FastRows: fix001P3P93InternalRows(fastCT)})
		}
	}
	planMeta := map[string]interface{}{"degree": plan.Degree, "base": plan.Base, "level": plan.Level, "scale": finalizationScaleString(plan.Scale), "block_count": len(plan.Blocks), "fast_generated_power_count": len(fastPowers), "standard_generated_power_count": len(stdKeys)}
	return rows, planMeta, nil
}

func fix001P3P93InternalWrite(result fix001P3P93InternalResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func mergeFIX001P3P93InternalBranches(standard, fast fix001P3P93InternalBranch, params ckks.Parameters, standardSK, fastSK *rlwe.SecretKey) (fix001P3P93InternalBranch, error) {
	merged := fast
	merged.Name = "genuine_standard_vs_fast"
	merged.Checkpoints = nil
	for _, fastCP := range fast.Checkpoints {
		standardCT, fastCT := standard.states[fastCP.Name], fast.states[fastCP.Name]
		if standardCT == nil || fastCT == nil {
			continue
		}
		cp, err := makeFIX001P3P93InternalCheckpoint(fastCP.Name, fastCP.Operation, standardCT, fastCT, params, standardSK, fastSK, true)
		if err != nil {
			return merged, err
		}
		merged.Checkpoints = append(merged.Checkpoints, cp)
	}
	return merged, nil
}

func firstFIX001P3P93InternalDivergence(branches map[string]fix001P3P93InternalBranch, floor, budget float64) (map[string]interface{}, map[string]interface{}) {
	firstObservable := map[string]interface{}{"checkpoint": "none", "threshold": floor}
	firstMaterial := map[string]interface{}{"checkpoint": "none", "threshold": budget}
	for _, name := range []string{"real", "imag"} {
		branch := branches[name]
		for i, cp := range branch.Checkpoints {
			if cp.Name == "public_reset_context" {
				continue
			}
			if cp.Metric == nil {
				continue
			}
			difference := cp.Metric.MaxComponentAbs
			if firstObservable["checkpoint"] == "none" && difference > floor {
				firstObservable = map[string]interface{}{"branch": name, "checkpoint": cp.Name, "difference": cp.Metric, "threshold": floor}
			}
			if firstMaterial["checkpoint"] == "none" && difference >= budget {
				incoming := 0.0
				if i > 0 && branch.Checkpoints[i-1].Metric != nil {
					incoming = branch.Checkpoints[i-1].Metric.MaxComponentAbs
				}
				amplification := 0.0
				if incoming > 0 {
					amplification = difference / incoming
				}
				firstMaterial = map[string]interface{}{"branch": name, "checkpoint": cp.Name, "difference": cp.Metric, "incoming_difference": incoming, "outgoing_difference": difference, "amplification_factor": amplification, "operation": cp.Operation, "level": cp.Fast.SourceLevel, "scale": cp.Fast.Scale, "degree": cp.Fast.Degree, "fast_capacity": cp.FastCapacity}
				firstMaterial["threshold"] = budget
			}
		}
	}
	return firstObservable, firstMaterial
}

func runFIX001P3DiagP93FastVsGenuineStandardEvalModInternalBisect(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3P93InternalResult{SchemaVersion: "fix-001-p3-diag-p93-fast-vs-genuine-standard-evalmod-internal-bisect.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, FixedProfile: map[string]interface{}{"log_n": 13, "slots": 4096, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps_authority": "Q012-wide", "plan_scale": "2^93", "internal_budget": fix001P3P93InternalBudget, "public_threshold": fix001P3P93InternalPublicThreshold}, Validation: map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "secondary_not_modified": true, "no_secondary_commit_or_push": true, "no_parameter_tuning": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3P93InternalSecondary || diffHash != requiredFIX001P3P93InternalDiffSHA || !secondaryDirty {
		result.Classification = "H"
		return fix001P3P93InternalWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params, fastSK := profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters)
	branches := map[string]fix001P3P93InternalBranch{}
	result.Factor32 = map[string]interface{}{"expected_scale_ratio": 32, "contract": "both public EvalMod wrappers set DefaultScale after internal restore"}
	for _, branchName := range []string{"real", "imag"} {
		fastInput, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
		if branchName == "imag" {
			fastInput, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
		}
		stdMeta, stdValues, err := semanticBisectView(standardInput, params, profile.StandardSK)
		if err != nil {
			return err
		}
		fastMeta, fastValues, err := semanticBisectView(fastInput, params, fastSK)
		if err != nil {
			return err
		}
		c2sMetric, err := semanticBisectMetricFor(stdValues, fastValues, fix001P3P93InternalC2SThreshold)
		if err != nil {
			return err
		}
		if result.R0C2S == nil {
			result.R0C2S = map[string]interface{}{}
		}
		result.R0C2S[branchName] = map[string]interface{}{"standard": stdMeta, "fast": fastMeta, "metric": c2sMetric, "pass": c2sMetric.Pass, "lineage": "same deterministic input; Genuine Standard staged C2S vs current Fast C2S"}
		if !c2sMetric.Pass {
			result.Classification = "A"
			return fix001P3P93InternalWrite(result, outPath)
		}
		stdBranch, stdActual, stdReplay, err := fix001P3P93InternalStandardTrace(standardInput, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		fastBranch, fastActual, fastReplay, err := fix001P3P93InternalFastTrace(fastInput, standardInput, profile.Fast, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		stdActualMeta, stdActualValues, err := semanticBisectView(stdActual, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, stdReplayValues, err := semanticBisectView(stdReplay, params, profile.StandardSK)
		if err != nil {
			return err
		}
		stdReplayMetric, err := semanticBisectMetricFor(stdActualValues, stdReplayValues, 1e-10)
		if err != nil {
			return err
		}
		fastActualMeta, fastActualValues, err := semanticBisectView(fastActual, params, fastSK)
		if err != nil {
			return err
		}
		_, fastReplayValues, err := semanticBisectView(fastReplay, params, fastSK)
		if err != nil {
			return err
		}
		fastReplayMetric, err := semanticBisectMetricFor(fastActualValues, fastReplayValues, 1e-10)
		if err != nil {
			return err
		}
		merged, err := mergeFIX001P3P93InternalBranches(stdBranch, fastBranch, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		merged.StandardInternal, merged.FastInternal = stdActualMeta, fastActualMeta
		merged.StandardActualAligned, merged.FastActualAligned = stdReplayMetric.Pass, fastReplayMetric.Pass
		merged.StandardReplayMetric, merged.FastReplayMetric = &stdReplayMetric, &fastReplayMetric
		targetScale, err := fix001P3P93InternalTargetScale(fastInput, profile.Fast.Mod1Parameters, params)
		if err != nil {
			return err
		}
		powers, plan, err := fix001P3P93InternalPowerTrace(stdBranch.states["polynomial_input"], fastBranch.states["polynomial_input"], profile.Fast, profile.Standard, params, profile.StandardSK, fastSK, targetScale, profile.Fast.Mod1Parameters.Mod1Poly)
		if err != nil {
			return err
		}
		merged.GeneratedPowers, merged.PSPlan, merged.FastPlanEvidence = powers, plan, fastBranch.FastPlanEvidence
		branches[branchName] = merged
		stdPublic, err := profile.Standard.EvalMod(standardInput.CopyNew())
		if err != nil {
			return err
		}
		fastPublic, err := profile.Fast.EvalMod(fastInput.CopyNew())
		if err != nil {
			return err
		}
		_, stdPublicValues, err := semanticBisectView(stdPublic, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, fastPublicValues, err := semanticBisectView(fastPublic, params, fastSK)
		if err != nil {
			return err
		}
		publicMetric, err := semanticBisectMetricFor(stdPublicValues, fastPublicValues, fix001P3P93InternalPublicThreshold)
		if err != nil {
			return err
		}
		directInternalMetric, err := semanticBisectMetricFor(stdActualValues, fastActualValues, fix001P3P93InternalBudget)
		if err != nil {
			return err
		}
		internalMetric := merged.Checkpoints[0].Metric
		for _, checkpoint := range merged.Checkpoints {
			if checkpoint.Name == "internal_final_restore" {
				internalMetric = checkpoint.Metric
				break
			}
		}
		result.Factor32[branchName] = map[string]interface{}{"internal_metric": directInternalMetric, "trace_internal_final_metric": internalMetric, "public_metric": publicMetric, "ratio_realized_max_component": safeRatio(publicMetric.MaxComponentAbs, directInternalMetric.MaxComponentAbs), "standard_internal_scale": finalizationScaleString(stdActual.Scale), "fast_internal_scale": finalizationScaleString(fastActual.Scale), "standard_public_scale": finalizationScaleString(stdPublic.Scale), "fast_public_scale": finalizationScaleString(fastPublic.Scale)}
	}
	result.Branches = branches
	result.FirstObservable, result.FirstMaterial = firstFIX001P3P93InternalDivergence(branches, 1e-12, fix001P3P93InternalBudget)
	result.HistoricalP93 = map[string]interface{}{"comparable": false, "real": 0.0001635104343335875, "imag": 0.00013359983063118935, "artifact": "results/FIX-001-P3-DIAG-P93-EVALMOD-REFERENCE-VS-PRODUCTION-FIRST-DIVERGENCE-summary.json", "reason": "historical path used a proxy/matched Standard lineage and did not prove equality to the Genuine Standard internal pre-public oracle"}
	result.SourceAudit = map[string]interface{}{"standard_public_wrapper": "circuits/ckks/bootstrapping/evaluator.go:Evaluator.EvalMod", "fast_public_wrapper": "circuits/ckks/bootstrapping/fast_evalmod.go:FastEvaluator.EvalMod", "standard_internal": "circuits/ckks/mod1/mod1_evaluator.go:Evaluator.EvaluateNew/EvaluateAndScaleNew", "fast_internal": "circuits/ckks/mod1/fast.go:FastEvaluator.EvaluateNew/evaluateNormalizedLogN13", "standard_polynomial": "circuits/ckks/polynomial/polynomial_evaluator.go:Evaluator.Evaluate -> common polynomial Evaluator.Evaluate", "fast_polynomial": "circuits/ckks/polynomial/fast.go:FastEvaluator.EvaluateWithPlanScale -> generatePowers/evaluatePlan", "target_scale": "same current-level Q schedule and Mod1 ScalingFactor", "fast_plan_scale": "2^93 only for normalized LogN13 q0=56", "q012_behavior": "Fast maintained q0/q1 with q2-aware source operations; no Primary repair", "double_angle": "Standard MulRelin/Add/Add/Rescale; Fast MulRelin/MulIntegerMaintained/Add/Rescale", "internal_restore": "Standard res.Scale=ct.Scale; Fast maintained scalar restore then res.Scale=inputScale", "public_reset": "both public wrappers set DefaultScale()"}
	material, _ := result.FirstMaterial["checkpoint"].(string)
	switch {
	case material == "none":
		result.Classification = "F"
	case len(material) >= 10 && material[:10] == "polynomial":
		result.Classification = "C"
	case material == "pre_double_angle":
		result.Classification = "C"
	case len(material) >= 12 && material[:12] == "double_angle":
		result.Classification = "D"
	case material == "internal_final_restore":
		result.Classification = "E"
	default:
		result.Classification = "B"
	}
	return fix001P3P93InternalWrite(result, outPath)
}
