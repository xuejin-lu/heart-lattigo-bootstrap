package main

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	ckkspolynomial "github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
)

type numericalEvalModTrace struct {
	checkpoints            []NumericalStageCheckpoint
	capacityAudit          []NumericalQPrefixCapacity
	scaleAudit             []NumericalScaleAudit
	powers                 []NumericalGeneratedPower
	plan                   NumericalPolynomialPlan
	fastReplayRMSE         float64
	standardReplayRMSE     float64
	fastReplayVerified     bool
	standardReplayVerified bool
}

func runNumericalEvalModStageReplay(
	branch string,
	fastInput, standardInput *rlwe.Ciphertext,
	actualFast, actualStandard []complex128,
	fastEval *bootstrapping.FastEvaluator,
	standardEval *bootstrapping.Evaluator,
	standardSK *rlwe.SecretKey,
	params ckks.Parameters,
) (numericalEvalModTrace, error) {
	if fastInput == nil || standardInput == nil || fastEval == nil || standardEval == nil || standardSK == nil {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod stage replay requires both inputs/evaluators and a Standard key", branch)
	}
	if fastEval.Mod1Evaluator == nil || standardEval.Mod1Evaluator == nil {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod evaluator is unavailable", branch)
	}
	fastParams := fastEval.Mod1Evaluator.Parameters
	standardParams := standardEval.Mod1Evaluator.Parameters
	if fastParams.Mod1Type != mod1.CosDiscrete || standardParams.Mod1Type != mod1.CosDiscrete || fastParams.Mod1InvPoly != nil || standardParams.Mod1InvPoly != nil {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod replay only supports the canonical CosDiscrete/no-inverse path", branch)
	}
	if fastInput.Level() != fastParams.LevelQ || standardInput.Level() != standardParams.LevelQ {
		return numericalEvalModTrace{}, fmt.Errorf("%s EvalMod level does not match Mod1 parameters", branch)
	}

	trace := numericalEvalModTrace{}
	fastEval.FastCKKS.SetQPrefixCapacityObserver(func(snapshot fastckks.QPrefixCapacitySnapshot) error {
		trace.capacityAudit = append(trace.capacityAudit, NumericalQPrefixCapacity{
			Checkpoint: branch + "/" + snapshot.Name, Level: snapshot.Level, Rows: snapshot.Rows,
			Scale: snapshot.Scale, Degree: snapshot.Degree, MaxAbs: append([]string(nil), snapshot.MaxAbs...),
			PrefixProduct: snapshot.PrefixProduct, StrictFit: snapshot.StrictFit,
		})
		return nil
	})
	defer fastEval.FastCKKS.SetQPrefixCapacityObserver(nil)
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
				checkpoint.AmplificationParent = previousInternal.Name
				checkpoint.DeltaSNRParent = previousInternal.Name
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
	fastRows, err := fastckks.QPrefixWidth(fastRes.Level())
	if err != nil {
		return trace, err
	}
	if err := fastEval.FastCKKS.ObserveQPrefixCapacity("evalmod-entry", fastRes, fastRows); err != nil {
		return trace, fmt.Errorf("%s Fast EvalMod input capacity: %w", branch, err)
	}
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

	powers, plan, err := fastEval.PolynomialEvaluator.DiagnosticGeneratePowers(fastRes, fastParams.Mod1Poly, targetScale)
	if err != nil {
		return trace, fmt.Errorf("%s DiagnosticGeneratePowers: %w", branch, err)
	}
	trace.plan = NumericalPolynomialPlan{
		Branch: branch, Degree: plan.Degree, Base: plan.Base, Level: plan.Level,
		ScaleLog2: plan.Scale.Log2(), ScaleExact: plan.Scale.Value.Text('e', 80), BlockCount: len(plan.Blocks),
	}
	standardPowers, err := numericalStandardGeneratedPowers(standardRes, standardParams.Mod1Poly, powers, standardEval.Evaluator)
	if err != nil {
		return trace, fmt.Errorf("%s genuine Standard generated powers: %w", branch, err)
	}
	trace.powers, err = numericalGeneratedPowerEvidence(branch, powers, standardPowers, standardSK, params)
	if err != nil {
		return trace, fmt.Errorf("%s generated-power Standard/Fast comparison: %w", branch, err)
	}
	standardPolynomial := standardParams.Mod1Poly.Clone()
	// EvaluateNew calls EvaluateAndScaleNew(ct, 1); with no inverse polynomial,
	// the Standard coefficient scaling is exactly one and leaves this clone unchanged.
	standardPolyOut, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardRes.CopyNew(), standardPolynomial, targetScale)
	if err != nil {
		return trace, fmt.Errorf("%s Standard polynomial replay: %w", branch, err)
	}
	fastPolyOut, err := fastEval.PolynomialEvaluator.EvaluateQPrefixRows(fastRes.CopyNew(), fastParams.Mod1Poly, targetScale, fastRows)
	if err != nil {
		return trace, fmt.Errorf("%s Fast polynomial replay: %w", branch, err)
	}
	if err := fastEval.FastCKKS.ObserveQPrefixCapacity("evalmod-polynomial-output", fastPolyOut, fastRows); err != nil {
		return trace, fmt.Errorf("%s Fast polynomial output capacity: %w", branch, err)
	}
	if _, err := add("polynomial_output", fastPolyOut, standardPolyOut); err != nil {
		return trace, err
	}
	fastRes = fastPolyOut.CopyNew()
	if _, err := add("before_double_angle_round_0", fastRes.CopyNew(), standardPolyOut.CopyNew()); err != nil {
		return trace, err
	}

	standardRes = standardPolyOut.CopyNew()
	standardSqrt2Pi := standardParams.Sqrt2Pi
	fastSqrt2Pi := fastParams.Sqrt2Pi
	for round := 0; round < fastParams.DoubleAngle; round++ {
		beforeLevel := fastRes.Level()
		if beforeLevel < 1 || beforeLevel != standardRes.Level() {
			return trace, fmt.Errorf("%s DoubleAngle round %d level mismatch Fast=%d Standard=%d", branch, round, beforeLevel, standardRes.Level())
		}
		rows, err := fastckks.QPrefixWidth(beforeLevel)
		if err != nil {
			return trace, err
		}
		if err := fastEval.FastCKKS.ObserveQPrefixCapacity(fmt.Sprintf("double-angle-%d-before", round), fastRes, rows); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d input capacity: %w", branch, round, err)
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
		if err := fastEval.FastCKKS.AddQPrefixRows(fastRes, fastRes, fastRes, rows); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d doubling: %w", branch, round, err)
		}
		if err := standardEval.Evaluator.Add(standardRes, standardRes, standardRes); err != nil {
			return trace, fmt.Errorf("%s Standard DoubleAngle round %d doubling: %w", branch, round, err)
		}
		if _, err := add(fmt.Sprintf("double_angle_round_%d_after_add", round), fastRes.CopyNew(), standardRes.CopyNew()); err != nil {
			return trace, err
		}
		if err := fastEval.FastCKKS.AddScalarQPrefixRows(fastRes, -fastSqrt2Pi, rows, fastRes); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d constant: %w", branch, round, err)
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
		rows, err = fastckks.QPrefixWidth(fastRes.Level())
		if err != nil {
			return trace, err
		}
		if err := fastEval.FastCKKS.ObserveQPrefixCapacity(fmt.Sprintf("double-angle-%d-after-rescale", round), fastRes, rows); err != nil {
			return trace, fmt.Errorf("%s Fast DoubleAngle round %d output capacity: %w", branch, round, err)
		}
		if _, err := add(fmt.Sprintf("double_angle_round_%d_after_rescale", round), fastRes.CopyNew(), standardRes.CopyNew()); err != nil {
			return trace, err
		}
		trace.scaleAudit = append(trace.scaleAudit, numericalScaleAuditForCT(fmt.Sprintf("%s/double_angle_round_%d_fast", branch, round), fastRes))
		trace.scaleAudit[len(trace.scaleAudit)-1].DoubleAngleRound = round
	}
	fastRes.Scale = inputScale
	standardRes.Scale = standardInput.Scale
	outputRows, err := fastckks.QPrefixWidth(fastRes.Level())
	if err != nil {
		return trace, err
	}
	if err := fastEval.FastCKKS.ObserveQPrefixCapacity("evalmod-output", fastRes, outputRows); err != nil {
		return trace, fmt.Errorf("%s Fast EvalMod output capacity: %w", branch, err)
	}
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

func numericalStandardGeneratedPowers(input *rlwe.Ciphertext, polynomial bignum.Polynomial, fastPowers []ckkspolynomial.DiagnosticPower, eval *ckks.Evaluator) (map[int]*rlwe.Ciphertext, error) {
	if input == nil || eval == nil || len(polynomial.Coeffs) == 0 {
		return nil, fmt.Errorf("Standard power replay requires an input, evaluator, and non-empty polynomial")
	}
	standardPolynomial := commonpolynomial.NewPolynomial(polynomial)
	logDegree := bits.Len64(uint64(standardPolynomial.Degree()))
	if logDegree == 0 {
		return nil, fmt.Errorf("Standard power replay requires positive polynomial degree")
	}
	powerBasis := commonpolynomial.NewPowerBasis(input, polynomial.Basis)
	if err := powerBasis.GenPower(1<<(logDegree-1), false, eval); err != nil {
		return nil, fmt.Errorf("Standard power %d: %w", 1<<(logDegree-1), err)
	}
	logSplit := bignum.OptimalSplit(logDegree)
	for n := (1 << logSplit) - 1; n > 2; n-- {
		if !(standardPolynomial.IsEven || standardPolynomial.IsOdd) ||
			(n&1 == 0 && standardPolynomial.IsEven) || (n&1 == 1 && standardPolynomial.IsOdd) {
			if err := powerBasis.GenPower(n, standardPolynomial.Lazy, eval); err != nil {
				return nil, fmt.Errorf("Standard power %d: %w", n, err)
			}
		}
	}

	standardPowers := make(map[int]*rlwe.Ciphertext, len(fastPowers))
	for _, fastPower := range fastPowers {
		if fastPower.N < 1 || fastPower.Ciphertext == nil {
			return nil, fmt.Errorf("Fast diagnostic power %d is invalid", fastPower.N)
		}
		standardPower := powerBasis.Value[fastPower.N]
		if standardPower == nil {
			return nil, fmt.Errorf("Standard PowerBasis did not generate required power T%d", fastPower.N)
		}
		standardPowers[fastPower.N] = standardPower
	}
	return standardPowers, nil
}

func numericalGeneratedPowerEvidence(branch string, powers []ckkspolynomial.DiagnosticPower, standardPowers map[int]*rlwe.Ciphertext, standardSK *rlwe.SecretKey, params ckks.Parameters) ([]NumericalGeneratedPower, error) {
	out := make([]NumericalGeneratedPower, 0, len(powers))
	for _, power := range powers {
		if power.Ciphertext == nil || power.N < 1 {
			return nil, fmt.Errorf("generated power %d has invalid ciphertext or power number", power.N)
		}
		standardPower := standardPowers[power.N]
		if standardPower == nil {
			return nil, fmt.Errorf("generated power T%d is missing its Standard ciphertext", power.N)
		}
		rows, err := fastckks.QPrefixWidth(power.Ciphertext.Level())
		if err != nil {
			return nil, fmt.Errorf("generated power %d Q-prefix: %w", power.N, err)
		}
		pair := numericalStagePair{
			fast:     numericalStageSampleFromCiphertext(branch+fmt.Sprintf("/T%d", power.N), power.Ciphertext, true),
			standard: numericalStageSampleFromCiphertext(branch+fmt.Sprintf("/T%d", power.N), standardPower, false),
		}
		decodeNumericalStagePair(&pair, params, standardSK)
		if !pair.fast.comparable || !pair.standard.comparable {
			return nil, fmt.Errorf("generated power T%d is not comparable: Fast=%s Standard=%s", power.N, pair.fast.reason, pair.standard.reason)
		}
		metrics, _ := numericalComparison(pair.standard.values, pair.fast.values)
		if field := firstNonFiniteNumber(metrics); field != "" {
			return nil, fmt.Errorf("generated power %d comparison overflowed at %s", power.N, field)
		}
		out = append(out, NumericalGeneratedPower{
			Branch:                     branch,
			ReferenceKind:              "genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator",
			Power:                      power.N,
			FastLevel:                  power.Ciphertext.Level(),
			FastScaleLog2:              power.Ciphertext.Scale.Log2(),
			FastScaleExact:             power.Ciphertext.Scale.Value.Text('e', 80),
			FastMaintainedRows:         rows,
			StandardLevel:              standardPower.Level(),
			StandardScaleLog2:          standardPower.Scale.Log2(),
			StandardScaleExact:         standardPower.Scale.Value.Text('e', 80),
			StandardAuthorityRows:      standardPower.Level() + 1,
			LevelScaleMatch:            power.Ciphertext.Level() == standardPower.Level() && power.Ciphertext.Scale.Equal(standardPower.Scale),
			RMSEFastStandard:           metrics.Complex.RMSE,
			MaxComplexDiffFastStandard: metrics.Complex.Max,
		})
	}
	return out, nil
}
