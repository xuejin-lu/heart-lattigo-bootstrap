package main

import (
	"fmt"
	"math"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	ckkspolynomial "github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
)

type numericalEvalModTrace struct {
	checkpoints            []NumericalStageCheckpoint
	scaleAudit             []NumericalScaleAudit
	powers                 []NumericalGeneratedPower
	plan                   NumericalPolynomialPlan
	fastReplayRMSE         float64
	standardReplayRMSE     float64
	fastReplayVerified     bool
	standardReplayVerified bool
}

func runNumericalEvalModBisect(
	branch string,
	fastInput, standardInput *rlwe.Ciphertext,
	actualFast, actualStandard []complex128,
	fastEval *bootstrapping.FastEvaluator,
	standardEval *bootstrapping.Evaluator,
	standardSK *rlwe.SecretKey,
	params ckks.Parameters,
) (numericalEvalModTrace, error) {
	if fastInput == nil || standardInput == nil || fastEval == nil || standardEval == nil || standardSK == nil {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod bisect requires both inputs/evaluators and a Standard key", branch)
	}
	if fastEval.Mod1Evaluator == nil || standardEval.Mod1Evaluator == nil {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod evaluator is unavailable", branch)
	}
	fastParams := fastEval.Mod1Evaluator.Parameters
	standardParams := standardEval.Mod1Evaluator.Parameters
	if fastParams.Mod1Type != mod1.CosDiscrete || standardParams.Mod1Type != mod1.CosDiscrete || fastParams.Mod1InvPoly != nil || standardParams.Mod1InvPoly != nil {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod source replay only supports the canonical CosDiscrete/no-inverse path", branch)
	}
	if fastInput.Level() != fastParams.LevelQ || standardInput.Level() != standardParams.LevelQ {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod level does not match Mod1 parameters", branch)
	}

	trace := numericalEvalModTrace{}
	var previousInternal *NumericalStageCheckpoint
	add := func(name string, fastCT, standardCT *rlwe.Ciphertext) (numericalStagePair, error) {
		fastSample := numericalStageSampleFromCiphertext(branch+"/"+name, fastCT, true)
		standardSample := numericalStageSampleFromCiphertext(branch+"/"+name, standardCT, false)
		pair := numericalStagePair{fast: fastSample, standard: standardSample}
		decodeNumericalStagePair(&pair, params, standardSK)
		checkpoint := numericalInternalCheckpoint(pair)
		if checkpoint.Comparable {
			checkpoint.AmplificationStatus = "NO_PREVIOUS_CHECKPOINT"
			checkpoint.DeltaSNRStatus = "NO_PREVIOUS_CHECKPOINT"
			if previousInternal != nil && previousInternal.D != nil && checkpoint.D != nil {
				if *previousInternal.D > 0 {
					factor := *checkpoint.D / *previousInternal.D
					if math.IsInf(factor, 0) || math.IsNaN(factor) {
						checkpoint.AmplificationStatus = "NON_FINITE"
					} else {
						checkpoint.Amplification = &factor
						checkpoint.AmplificationStatus = "FINITE"
					}
				} else if *checkpoint.D == 0 {
					checkpoint.AmplificationStatus = "UNDEFINED_ZERO_OVER_ZERO"
				} else {
					checkpoint.AmplificationStatus = "UNDEFINED_ZERO_PREVIOUS_D"
				}
				checkpoint.DeltaSNRDB, checkpoint.DeltaSNRStatus = numericalSNRDelta(checkpoint.StageReferenceSNR, previousInternal.StageReferenceSNR)
			}
			checkpointCopy := checkpoint
			previousInternal = &checkpointCopy
		}
		trace.checkpoints = append(trace.checkpoints, checkpoint)
		trace.scaleAudit = append(trace.scaleAudit,
			numericalScaleAuditForCT(branch+"/"+name+"/fast", fastCT),
			numericalScaleAuditForCT(branch+"/"+name+"/standard", standardCT),
		)
		return pair, nil
	}

	if _, err := add("input_before_normalization", fastInput.CopyNew(), standardInput.CopyNew()); err != nil {
		return trace, err
	}
	inputScale := fastInput.Scale
	fastRes := fastInput.CopyNew()
	standardRes := standardInput.CopyNew()
	fastRes.Scale = fastParams.ScalingFactor()
	standardRes.Scale = standardParams.ScalingFactor()
	if _, err := add("after_normalization", fastRes, standardRes); err != nil {
		return trace, err
	}
	targetScale, err := numericalEvalModTargetScale(fastParams, params, fastInput.Level())
	if err != nil {
		return trace, err
	}
	offset := numericalEvalModOffset(fastParams)
	fastRows, err := fastckks.QPrefixWidth(fastRes.Level())
	if err != nil {
		return trace, err
	}
	if err := fastEval.FastCKKS.AddScalarQPrefixRows(fastRes, offset, fastRows, fastRes); err != nil {
		return trace, fmt.Errorf("%s Fast Chebyshev offset replay: %w", branch, err)
	}
	if err := standardEval.Evaluator.Add(standardRes, offset, standardRes); err != nil {
		return trace, fmt.Errorf("%s Standard Chebyshev offset replay: %w", branch, err)
	}
	if _, err := add("after_chebyshev_offset", fastRes, standardRes); err != nil {
		return trace, err
	}
	polynomialInputPair, err := add("polynomial_input", fastRes, standardRes)
	if err != nil {
		return trace, err
	}
	if !polynomialInputPair.fast.comparable || !polynomialInputPair.standard.comparable {
		return trace, fmt.Errorf("%s polynomial input is not semantically comparable", branch)
	}

	planBits := 91
	planScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(planBits)))
	powers, plan, err := fastEval.PolynomialEvaluator.DiagnosticGeneratePowers(fastRes, fastParams.Mod1Poly, targetScale)
	if err != nil {
		return trace, fmt.Errorf("%s DiagnosticGeneratePowers: %w", branch, err)
	}
	trace.plan = NumericalPolynomialPlan{
		Degree: plan.Degree, Base: plan.Base, Level: plan.Level,
		ScaleLog2: plan.Scale.Log2(), ScaleExact: plan.Scale.Value.Text('e', 80), BlockCount: len(plan.Blocks),
	}
	trace.powers, err = numericalGeneratedPowerEvidence(powers, polynomialInputPair.standard.values, params)
	if err != nil {
		return trace, fmt.Errorf("%s generated-power semantic inspection: %w", branch, err)
	}
	standardPolynomial := standardParams.Mod1Poly.Clone()
	// EvaluateNew calls EvaluateAndScaleNew(ct, 1); with no inverse polynomial,
	// the Standard coefficient scaling is exactly one and leaves this clone unchanged.
	standardPolyOut, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardRes.CopyNew(), standardPolynomial, targetScale)
	if err != nil {
		return trace, fmt.Errorf("%s Standard polynomial replay: %w", branch, err)
	}
	fastPolyOut, err := fastEval.PolynomialEvaluator.EvaluateWithPlanScaleQPrefixRows(fastRes.CopyNew(), fastParams.Mod1Poly, targetScale, planScale, fastRows)
	if err != nil {
		return trace, fmt.Errorf("%s Fast polynomial replay: %w", branch, err)
	}
	if _, err := add("generated_power_polynomial_output", fastPolyOut, standardPolyOut); err != nil {
		return trace, err
	}
	if _, err := add("before_coherent_scale_transition", fastPolyOut.CopyNew(), standardPolyOut.CopyNew()); err != nil {
		return trace, err
	}
	workingExponent, kExponent, multiplierExponent := 31, 29, 30
	workingScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(workingExponent)))
	if fastPolyOut.Level() != 7 || fastPolyOut.Degree() != 1 || !fastPolyOut.IsNTT || !fastPolyOut.IsMontgomery || !fastPolyOut.Scale.InDelta(workingScale, 32) {
		return trace, fmt.Errorf("%s Fast polynomial output does not match source normalized-path invariant: level=%d scale=%s NTT=%t Montgomery=%t", branch, fastPolyOut.Level(), fastPolyOut.Scale.Value.Text('e', 20), fastPolyOut.IsNTT, fastPolyOut.IsMontgomery)
	}
	kIn, err := numericalNearestScaleExponent(targetScale.Div(fastPolyOut.Scale))
	if err != nil {
		return trace, err
	}
	if kIn != kExponent {
		return trace, fmt.Errorf("%s normalized coherent-scale kIn=%d, expected source value %d", branch, kIn, kExponent)
	}
	kValue := new(big.Int).Lsh(big.NewInt(1), uint(kIn))
	coherentScale := fastPolyOut.Scale.Mul(rlwe.NewScale(kValue))
	if !coherentScale.InDelta(targetScale, 32) {
		return trace, fmt.Errorf("%s normalized coherent scale is outside target tolerance", branch)
	}
	fastRes = fastPolyOut.CopyNew()
	fastRes.Scale = coherentScale
	trace.scaleAudit = append(trace.scaleAudit, NumericalScaleAudit{
		Checkpoint: branch + "/normalized_coherent_scale_transition",
		Scale:      coherentScale.Float64(), ScaleExact: coherentScale.Value.Text('e', 80), ScaleFloat64Finite: true, ScaleLog2: coherentScale.Log2(),
		PlanScaleBits: planBits, PlanScaleExact: planScale.Value.Text('e', 80), WorkingScaleBits: workingExponent,
		TargetScale: targetScale.Float64(), TargetScaleExact: targetScale.Value.Text('e', 80), TargetScaleFloat64Finite: true, TargetScaleLog2: targetScale.Log2(),
		Level: fastRes.Level(), KInExponent: kIn, MultiplierExponent: multiplierExponent,
	})
	if _, err := add("after_coherent_scale_transition", fastRes.CopyNew(), standardPolyOut.CopyNew()); err != nil {
		return trace, err
	}
	if _, err := add("before_double_angle_round_0", fastRes.CopyNew(), standardPolyOut.CopyNew()); err != nil {
		return trace, err
	}

	standardRes = standardPolyOut.CopyNew()
	standardSqrt2Pi := standardParams.Sqrt2Pi
	fastSqrt2Pi := fastParams.Sqrt2Pi
	currentExponent := kIn
	for round := 0; round < fastParams.DoubleAngle; round++ {
		beforeLevel := fastRes.Level()
		if beforeLevel < 1 || beforeLevel != standardRes.Level() {
			return trace, fmt.Errorf("%s DoubleAngle round %d level mismatch Fast=%d Standard=%d", branch, round, beforeLevel, standardRes.Level())
		}
		rows, err := fastckks.QPrefixWidth(beforeLevel)
		if err != nil {
			return trace, err
		}
		nextScale := fastRes.Scale.Mul(fastRes.Scale).Div(rlwe.NewScale(params.Q()[beforeLevel]))
		nextExponent, err := numericalNearestScaleExponent(nextScale.Div(workingScale))
		if err != nil {
			return trace, err
		}
		aExponent := 1 + 2*currentExponent - nextExponent
		if nextExponent != kExponent || aExponent != multiplierExponent {
			return trace, fmt.Errorf("%s DoubleAngle round %d source scale exponents changed: k=%d multiplier=%d", branch, round, nextExponent, aExponent)
		}
		standardSqrt2Pi *= standardSqrt2Pi
		fastSqrt2Pi *= fastSqrt2Pi
		if err := fastEval.FastCKKS.MulRelinElementQPrefixRows(fastRes, fastRes.El(), rows, fastRes); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d multiply: %w", branch, round, err)
		}
		if err := standardEval.Evaluator.MulRelin(standardRes, standardRes, standardRes); err != nil {
			return trace, fmt.Errorf("%s Standard DoubleAngle round %d multiply: %w", branch, round, err)
		}
		if _, err := add(fmt.Sprintf("double_angle_round_%d_after_multiply", round), fastRes.CopyNew(), standardRes.CopyNew()); err != nil {
			return trace, err
		}
		factor := new(big.Int).Lsh(big.NewInt(1), uint(aExponent))
		constant := new(big.Float).SetPrec(256).SetFloat64(fastSqrt2Pi)
		constant.Quo(constant, new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(nextExponent))))
		fastConstant, _ := constant.Float64()
		if err := fastEval.FastCKKS.MulIntegerQPrefixRows(fastRes, factor, rows, fastRes); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d scale multiplier: %w", branch, round, err)
		}
		if err := fastEval.FastCKKS.AddScalarQPrefixRows(fastRes, -fastConstant, rows, fastRes); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d constant: %w", branch, round, err)
		}
		if err := standardEval.Evaluator.Add(standardRes, standardRes, standardRes); err != nil {
			return trace, fmt.Errorf("%s Standard DoubleAngle round %d doubling: %w", branch, round, err)
		}
		if err := standardEval.Evaluator.Add(standardRes, -complex(standardSqrt2Pi, 0), standardRes); err != nil {
			return trace, fmt.Errorf("%s Standard DoubleAngle round %d constant: %w", branch, round, err)
		}
		if _, err := add(fmt.Sprintf("double_angle_round_%d_after_constant", round), fastRes.CopyNew(), standardRes.CopyNew()); err != nil {
			return trace, err
		}
		if err := fastEval.FastCKKS.RescaleQPrefixRows(fastRes, rows, fastRes); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d Rescale: %w", branch, round, err)
		}
		if err := standardEval.Evaluator.Rescale(standardRes, standardRes); err != nil {
			return trace, fmt.Errorf("%s Standard DoubleAngle round %d Rescale: %w", branch, round, err)
		}
		if _, err := add(fmt.Sprintf("double_angle_round_%d_after_rescale", round), fastRes.CopyNew(), standardRes.CopyNew()); err != nil {
			return trace, err
		}
		trace.scaleAudit = append(trace.scaleAudit, numericalScaleAuditForCT(fmt.Sprintf("%s/double_angle_round_%d_fast", branch, round), fastRes))
		trace.scaleAudit[len(trace.scaleAudit)-1].DoubleAngleRound = round
		trace.scaleAudit[len(trace.scaleAudit)-1].PlanScaleBits = planBits
		trace.scaleAudit[len(trace.scaleAudit)-1].PlanScaleExact = planScale.Value.Text('e', 80)
		trace.scaleAudit[len(trace.scaleAudit)-1].WorkingScaleBits = workingExponent
		trace.scaleAudit[len(trace.scaleAudit)-1].KInExponent = nextExponent
		trace.scaleAudit[len(trace.scaleAudit)-1].MultiplierExponent = aExponent
		currentExponent = nextExponent
	}
	if currentExponent != kExponent || fastRes.Level() != 4 {
		return trace, fmt.Errorf("%s Fast DoubleAngle final recurrence state differs from source invariant", branch)
	}
	rows, err := fastckks.QPrefixWidth(fastRes.Level())
	if err != nil {
		return trace, err
	}
	if err := fastEval.FastCKKS.MulIntegerQPrefixRows(fastRes, new(big.Int).Lsh(big.NewInt(1), uint(currentExponent)), rows, fastRes); err != nil {
		return trace, fmt.Errorf("%s Fast final normalized restore: %w", branch, err)
	}
	fastRes.Scale = inputScale
	standardRes.Scale = standardInput.Scale
	if _, err := add("evalmod_output_before_public_scale_reset", fastRes, standardRes); err != nil {
		return trace, err
	}
	fastPublic := fastRes.CopyNew()
	standardPublic := standardRes.CopyNew()
	fastPublic.Scale = params.DefaultScale()
	standardPublic.Scale = params.DefaultScale()
	publicPair, err := add("evalmod_output_after_public_scale_reset", fastPublic, standardPublic)
	if err != nil {
		return trace, err
	}
	if publicPair.fast.comparable && publicPair.standard.comparable {
		fastDiff, _ := numericalComparison(actualFast, publicPair.fast.values)
		standardDiff, _ := numericalComparison(actualStandard, publicPair.standard.values)
		trace.fastReplayRMSE = fastDiff.Complex.RMSE
		trace.standardReplayRMSE = standardDiff.Complex.RMSE
		trace.fastReplayVerified = firstNonFiniteNumber(fastDiff) == "" && fastDiff.Complex.RMSE <= 1e-12
		trace.standardReplayVerified = firstNonFiniteNumber(standardDiff) == "" && standardDiff.Complex.RMSE <= 1e-12
	}
	return trace, nil
}

func numericalInternalCheckpoint(pair numericalStagePair) NumericalStageCheckpoint {
	checkpoint := NumericalStageCheckpoint{
		Name: pair.fast.name, Fast: pair.fast.state, Standard: pair.standard.state,
		AmplificationStatus: "NOT_COMPARABLE", DeltaSNRStatus: "NOT_COMPARABLE",
		StageReferenceSNR: numericalmetrics.Unavailable("internal trace checkpoint is not comparable"),
	}
	if !pair.fast.comparable || !pair.standard.comparable {
		checkpoint.NotComparableReason = pair.fast.reason + "; " + pair.standard.reason
		return checkpoint
	}
	metrics, _ := numericalComparison(pair.standard.values, pair.fast.values)
	if firstNonFiniteNumber(metrics) != "" {
		checkpoint.NotComparableReason = "internal semantic comparison overflowed"
		return checkpoint
	}
	checkpoint.Comparable = true
	checkpoint.FastVsStandard = &metrics
	distance := metrics.Complex.RMSE
	checkpoint.D = &distance
	checkpoint.StageReferenceSNR = numericalmetrics.Compare(pair.standard.values, pair.fast.values)
	return checkpoint
}

func numericalScaleAuditForCT(name string, ct *rlwe.Ciphertext) NumericalScaleAudit {
	metadata := numericalMetadata(ct)
	return NumericalScaleAudit{
		Checkpoint: name, Scale: metadata.Scale, ScaleExact: metadata.ScaleExact,
		ScaleFloat64Finite: metadata.ScaleFloat64Finite, ScaleLog2: metadata.ScaleLog2, Level: metadata.Level,
	}
}

func numericalEvalModTargetScale(params mod1.Parameters, ckksParams ckks.Parameters, level int) (rlwe.Scale, error) {
	target := params.ScalingFactor()
	for i := 0; i < params.DoubleAngle; i++ {
		index := level - params.Mod1Poly.Depth() - params.DoubleAngle + i + 1
		if index < 0 || index >= len(ckksParams.Q()) {
			return rlwe.Scale{}, fmt.Errorf("EvalMod target-scale Q index %d is outside chain", index)
		}
		target = target.Mul(rlwe.NewScale(ckksParams.Q()[index]))
		target.Value.Sqrt(&target.Value)
	}
	return target, nil
}

func numericalEvalModOffset(params mod1.Parameters) *big.Float {
	offset := new(big.Float).Sub(&params.Mod1Poly.B, &params.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(params.IntervalShrinkFactor()))
	return offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
}

func numericalNearestScaleExponent(scale rlwe.Scale) (int, error) {
	if scale.Value.Sign() <= 0 {
		return 0, fmt.Errorf("scale ratio must be positive")
	}
	mantissa := new(big.Float).SetPrec(256)
	exponent := scale.Value.MantExp(mantissa)
	if mantissa.Cmp(new(big.Float).SetPrec(256).SetFloat64(0.75)) >= 0 {
		return exponent, nil
	}
	return exponent - 1, nil
}

func numericalGeneratedPowerEvidence(powers []ckkspolynomial.DiagnosticPower, referenceInput []complex128, params ckks.Parameters) ([]NumericalGeneratedPower, error) {
	out := make([]NumericalGeneratedPower, 0, len(powers))
	for _, power := range powers {
		if power.Ciphertext == nil || power.N < 1 {
			return nil, fmt.Errorf("generated power %d has invalid ciphertext or power number", power.N)
		}
		rows, err := fastckks.QPrefixWidth(power.Ciphertext.Level())
		if err != nil {
			return nil, fmt.Errorf("generated power %d Q-prefix: %w", power.N, err)
		}
		projected, err := projectNumericalStageCiphertext(params, power.Ciphertext, min(rows, power.Ciphertext.Level()+1)-1)
		if err != nil {
			return nil, fmt.Errorf("generated power %d projection: %w", power.N, err)
		}
		fastValues, err := fastStandardDecodeFast(params, projected)
		if err != nil {
			return nil, fmt.Errorf("generated power %d Fast semantic decode: %w", power.N, err)
		}
		reference, err := numericalChebyshevPower(referenceInput, power.N)
		if err != nil {
			return nil, fmt.Errorf("generated power %d reference: %w", power.N, err)
		}
		metrics, _ := numericalComparison(reference, fastValues)
		if field := firstNonFiniteNumber(metrics); field != "" {
			return nil, fmt.Errorf("generated power %d comparison overflowed at %s", power.N, field)
		}
		out = append(out, NumericalGeneratedPower{
			ReferenceKind:      "plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power",
			Power:              power.N,
			Level:              power.Ciphertext.Level(),
			ScaleLog2:          power.Ciphertext.Scale.Log2(),
			FastMaintainedRows: rows,
			RMSE:               metrics.Complex.RMSE,
			MaxComplexDiff:     metrics.Complex.Max,
		})
	}
	return out, nil
}

func numericalChebyshevPower(input []complex128, power int) ([]complex128, error) {
	if power < 1 || len(input) == 0 {
		return nil, fmt.Errorf("Chebyshev reference requires a positive power and non-empty input")
	}
	previous := make([]complex128, len(input))
	for i := range previous {
		previous[i] = 1
	}
	current := append([]complex128(nil), input...)
	if power == 1 {
		return current, validateNumericalVector(current)
	}
	for n := 2; n <= power; n++ {
		next := make([]complex128, len(input))
		for i, value := range input {
			next[i] = 2*value*current[i] - previous[i]
		}
		previous, current = current, next
	}
	if err := validateNumericalVector(current); err != nil {
		return nil, err
	}
	return current, nil
}
