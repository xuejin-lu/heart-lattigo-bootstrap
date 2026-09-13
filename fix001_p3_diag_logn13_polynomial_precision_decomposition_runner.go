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
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3PolynomialDecompositionPrimaryBase = "571cd81854ec8b1834c33162cc40810d832d3cd3"
	requiredFIX001P3PolynomialDecompositionSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	polynomialDecompositionPlanScaleExponent           = 91
	polynomialDecompositionBudget                      = 1.2e-8
	polynomialDecompositionOracleTarget                = 1e-10
	polynomialDecompositionPowerTarget                 = 1e-10
	polynomialDecompositionPublicThreshold             = 1e-2
)

var polynomialDecompositionPowerDegrees = []int{1, 2, 3, 4, 6, 8, 16}

type polynomialDecompositionPowerSummary struct {
	N                 int             `json:"n"`
	Level             int             `json:"level"`
	Scale             string          `json:"scale"`
	SourceError       *PSGlobalMetric `json:"source_semantic_error"`
	OracleInjectError *PSGlobalMetric `json:"oracle_injection_error"`
	CapacityRatio     float64         `json:"q01_capacity_ratio"`
	CenteredUnique    bool            `json:"centered_unique"`
	OracleRowsMatch   bool            `json:"oracle_full_rns_q01_rows_match"`
}

type polynomialDecompositionCheckpoint struct {
	FirstCheckpoint         string          `json:"first_checkpoint_exceeding_budget"`
	LocalError              *PSGlobalMetric `json:"local_arithmetic_error"`
	CumulativeError         *PSGlobalMetric `json:"cumulative_oracle_path_error"`
	WorstCapacityCheckpoint string          `json:"worst_capacity_checkpoint"`
	WorstCapacityRatio      float64         `json:"worst_capacity_ratio"`
}

type polynomialDecompositionSensitivity struct {
	PowerN             int             `json:"replaced_power_n"`
	RealError          *PSGlobalMetric `json:"real_error_vs_reference"`
	ImagError          *PSGlobalMetric `json:"imag_error_vs_reference"`
	RealErrorReduction float64         `json:"real_error_reduction"`
	ImagErrorReduction float64         `json:"imag_error_reduction"`
}

type polynomialDecompositionDownstream struct {
	Reached         bool            `json:"reached"`
	Reason          string          `json:"reason,omitempty"`
	EvalModReal     *PSGlobalMetric `json:"evalmod_real_vs_standard,omitempty"`
	EvalModImag     *PSGlobalMetric `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2C         *PSGlobalMetric `json:"post_s2c_vs_standard,omitempty"`
	PublicLike      *PSGlobalMetric `json:"public_like_vs_message,omitempty"`
	StandardLike    *PSGlobalMetric `json:"standard_like_vs_message,omitempty"`
	MetadataCorrect bool            `json:"public_metadata_correct"`
	InputUnchanged  bool            `json:"input_unchanged"`
}

type polynomialDecompositionBranchSummary struct {
	Branch                  string                                `json:"branch"`
	D1SourceVsStandard      *PSGlobalMetric                       `json:"d1_source_oracle_vs_standard"`
	ActualPowers            []polynomialDecompositionPowerSummary `json:"actual_production_powers"`
	HActual                 *PSGlobalMetric                       `json:"h_actual_vs_reference"`
	HOracle                 *PSGlobalMetric                       `json:"h_oracle_vs_reference"`
	FActual                 *PSGlobalMetric                       `json:"f_actual_vs_reference"`
	FOracle                 *PSGlobalMetric                       `json:"f_oracle_vs_reference"`
	PowerSemanticResidual   *PSGlobalMetric                       `json:"power_semantic_residual"`
	FastPSActualResidual    *PSGlobalMetric                       `json:"fast_ps_arithmetic_residual_on_actual_powers"`
	FastPSOracleResidual    *PSGlobalMetric                       `json:"fast_ps_arithmetic_floor_with_oracle_powers"`
	OracleReferenceResidual *PSGlobalMetric                       `json:"oracle_reference_residual"`
	DecompositionClosure    *PSGlobalMetric                       `json:"decomposition_closure_residual"`
	FActualVsFOracle        *PSGlobalMetric                       `json:"f_actual_vs_f_oracle"`
	OraclePSCheckpoint      polynomialDecompositionCheckpoint     `json:"oracle_ps_checkpoint"`
	Sensitivity             []polynomialDecompositionSensitivity  `json:"one_power_replacement_sensitivity"`
}

type polynomialDecompositionResult struct {
	SchemaVersion         string                                          `json:"schema_version"`
	Timestamp             time.Time                                       `json:"timestamp"`
	Primary               RepositoryMetadata                              `json:"primary_repository"`
	Lattigo               RepositoryMetadata                              `json:"lattigo_repository"`
	Environment           EnvironmentMetadata                             `json:"environment"`
	Config                BootstrapConfig                                 `json:"config"`
	Parameters            ExperimentParameters                            `json:"effective_parameters"`
	Workload              CorrectnessWorkload                             `json:"workload"`
	PlanScale             string                                          `json:"production_ps_plan_scale"`
	Baseline              map[string]interface{}                          `json:"baseline_reproduction"`
	Branches              map[string]polynomialDecompositionBranchSummary `json:"branches"`
	Downstream            *polynomialDecompositionDownstream              `json:"d7_downstream_oracle_intervention,omitempty"`
	Classification        string                                          `json:"classification"`
	FirstSupportedBlocker string                                          `json:"first_supported_blocker"`
	CausalDisposition     string                                          `json:"causal_disposition"`
	Validation            map[string]interface{}                          `json:"validation"`
}

type polynomialDecompositionBranchWork struct {
	Summary           polynomialDecompositionBranchSummary
	InputValues       []complex128
	ReferenceValues   []complex128
	ActualPowerMap    map[int]*rlwe.Ciphertext
	OraclePowerMap    map[int]*rlwe.Ciphertext
	PowerExpected     map[int][]complex128
	ActualPowerValues map[int][]complex128
	OraclePowerValues map[int][]complex128
	ActualOutput      []complex128
	OracleOutput      []complex128
	HActualValues     []complex128
	HOracleValues     []complex128
	ActualCiphertext  *rlwe.Ciphertext
	OracleCiphertext  *rlwe.Ciphertext
	Plan              commonpolynomial.PatersonStockmeyerPolynomial
	PlanScale         rlwe.Scale
	ActualChecks      []PSGlobalCheckpoint
	OracleChecks      []PSGlobalCheckpoint
}

type polynomialNumericNode struct {
	Degree int
	Value  []complex128
}

func polynomialDecompositionMetric(expected, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(expected, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func polynomialDecompositionMetricCopy(source *PSGlobalMetric, threshold float64) *PSGlobalMetric {
	if source == nil {
		return nil
	}
	copyMetric := *source
	copyMetric.Threshold = threshold
	copyMetric.Pass = copyMetric.MaxComponent <= threshold
	return &copyMetric
}

func polynomialDecompositionResidualMetric(left, right []complex128, threshold float64) *PSGlobalMetric {
	if len(left) != len(right) {
		return polynomialDecompositionMetric(nil, nil, threshold)
	}
	difference := make([]complex128, len(left))
	for i := range difference {
		difference[i] = left[i] - right[i]
	}
	return polynomialDecompositionMetric(make([]complex128, len(difference)), difference, threshold)
}

func polynomialDecompositionComponentMetric(expected, actual []complex128, threshold float64, realComponent bool) *PSGlobalMetric {
	if len(expected) != len(actual) {
		return polynomialDecompositionMetric(nil, nil, threshold)
	}
	left := make([]complex128, len(expected))
	right := make([]complex128, len(actual))
	for i := range expected {
		if realComponent {
			left[i] = complex(real(expected[i]), 0)
			right[i] = complex(real(actual[i]), 0)
		} else {
			left[i] = complex(0, imag(expected[i]))
			right[i] = complex(0, imag(actual[i]))
		}
	}
	return polynomialDecompositionMetric(left, right, threshold)
}

func polynomialDecompositionVectorAddScaled(dst, value []complex128, scalar complex128) {
	for i := range dst {
		dst[i] += scalar * value[i]
	}
}

func polynomialDecompositionVectorMul(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] * b[i]
	}
	return out
}

func polynomialDecompositionVectorAdd(a, b []complex128) []complex128 {
	out := append([]complex128(nil), a...)
	for i := range out {
		out[i] += b[i]
	}
	return out
}

func polynomialDecompositionNumericPS(plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int][]complex128, input []complex128) ([]complex128, error) {
	baby := make([]*polynomialNumericNode, len(plan.Value))
	for i := range plan.Value {
		p := plan.Value[i]
		value := make([]complex128, len(input))
		if p.IsEven {
			if p.Coeffs[0] == nil {
				return nil, fmt.Errorf("numeric PS block %d has nil constant", i)
			}
			polynomialDecompositionVectorAddScaled(value, makeOnes(len(input)), psGlobalCoeff(p.Coeffs[0]))
		}
		for key := p.Degree(); key > 0; key-- {
			if (p.IsEven || p.IsOdd) && ((key&1 == 0 && !p.IsEven) || (key&1 == 1 && !p.IsOdd)) {
				continue
			}
			if p.Coeffs[key] == nil {
				continue
			}
			power, ok := powers[key]
			if !ok {
				return nil, fmt.Errorf("numeric PS missing T%d", key)
			}
			polynomialDecompositionVectorAddScaled(value, power, psGlobalCoeff(p.Coeffs[key]))
		}
		baby[len(plan.Value)-i-1] = &polynomialNumericNode{Degree: p.Degree(), Value: value}
	}
	for len(baby) != 1 {
		giant := make([]int, len(baby))
		for i := 0; i < len(baby); i++ {
			if i == len(baby)-1 {
				giant[i] = 2
			} else if baby[i].Degree == baby[i+1].Degree {
				giant[i] = 1
				i++
			}
		}
		for i := 0; i < len(baby); i++ {
			if giant[i] == 2 {
				baby[i].Degree = baby[i-1].Degree
				continue
			}
			if giant[i] != 1 {
				continue
			}
			a, b := baby[i], baby[i+1]
			degree := 1 << bits.Len64(uint64(a.Degree))
			power, ok := powers[degree]
			if !ok {
				return nil, fmt.Errorf("numeric PS missing giant T%d", degree)
			}
			parent := polynomialDecompositionVectorAdd(a.Value, polynomialDecompositionVectorMul(b.Value, power))
			b.Degree = 2*degree - 1
			b.Value = parent
			baby[i] = nil
			i++
		}
		kept := baby[:0]
		for _, node := range baby {
			if node != nil {
				kept = append(kept, node)
			}
		}
		baby = kept
	}
	return baby[0].Value, nil
}

func polynomialDecompositionPlan(eval *bootstrapping.FastEvaluator, powers map[int]*rlwe.Ciphertext, targetScale rlwe.Scale) (commonpolynomial.PatersonStockmeyerPolynomial, rlwe.Scale, error) {
	input, ok := powers[1]
	if !ok || input == nil {
		return commonpolynomial.PatersonStockmeyerPolynomial{}, rlwe.Scale{}, fmt.Errorf("production powers do not contain T1")
	}
	plan := commonpolynomial.NewPolynomial(eval.Mod1Parameters.Mod1Poly).PatersonStockmeyerPolynomial(eval.FastCKKS.GetRLWEParameters(), input.Level(), input.Scale, targetScale, psGlobalSim{params: eval.Parameters.BootstrappingParameters, levels: eval.Parameters.BootstrappingParameters.LevelsConsumedPerRescaling()})
	planScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), polynomialDecompositionPlanScaleExponent))
	for i := range plan.Value {
		plan.Value[i].Scale = planScale
	}
	return plan, planScale, nil
}

func polynomialDecompositionOraclePower(params ckks.Parameters, source *rlwe.Ciphertext, expected []complex128) (*rlwe.Ciphertext, []complex128, bool, error) {
	oracle, err := evalModFastFromStandard(params, source, expected)
	if err != nil {
		return nil, nil, false, err
	}
	decoded, err := psGlobalDecode(params, oracle)
	if err != nil {
		return nil, nil, false, err
	}
	metric := polynomialDecompositionMetric(expected, decoded, polynomialDecompositionOracleTarget)
	full, _, err := postMod1S2CLiftFull(params, oracle)
	if err != nil {
		return nil, decoded, metric.Pass, err
	}
	return oracle, decoded, metric.Pass && postMod1S2CRowsEqual(params, oracle, full), nil
}

func polynomialDecompositionCapacity(params ckks.Parameters, ct *rlwe.Ciphertext) (float64, bool, error) {
	if ct == nil || ct.Level() < 1 {
		return 0, true, nil
	}
	capacity, err := postMod1S2CCapacityFromFastCiphertext(params, ct)
	if err != nil {
		return 0, false, err
	}
	return capacity.MaxAbsOverQ01Half, capacity.Pass, nil
}

func polynomialDecompositionReplay(params ckks.Parameters, eval *bootstrapping.FastEvaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, input []complex128, planScale rlwe.Scale) ([]complex128, *rlwe.Ciphertext, []PSGlobalCheckpoint, error) {
	checks, _, root, _, err := psGlobalReplay(params, eval.FastCKKS, plan, powers, expected, decoded, input, planScale)
	if err != nil {
		return nil, nil, checks, err
	}
	if root.ciphertext == nil {
		return nil, nil, checks, fmt.Errorf("PS replay returned nil root ciphertext")
	}
	output := root.ciphertext.CopyNew()
	if err := eval.FastCKKS.Rescale(output, output); err != nil {
		return nil, nil, checks, err
	}
	values, err := psGlobalDecode(params, output)
	if err != nil {
		return nil, nil, checks, err
	}
	return values, output, checks, nil
}

func polynomialDecompositionOracleCheckpoint(params ckks.Parameters, eval *bootstrapping.FastEvaluator, checks []PSGlobalCheckpoint, output *rlwe.Ciphertext, reference []complex128, budget float64) (polynomialDecompositionCheckpoint, error) {
	result := polynomialDecompositionCheckpoint{}
	for _, check := range checks {
		if check.SourceBacked != nil && result.FirstCheckpoint == "" && check.SourceBacked.MaxComponent > budget {
			result.FirstCheckpoint = check.ID
			result.CumulativeError = polynomialDecompositionMetricCopy(check.SourceBacked, budget)
			if check.LocalConsistency != nil {
				result.LocalError = polynomialDecompositionMetricCopy(check.LocalConsistency, budget)
			}
		}
		if check.ciphertext != nil {
			ratio, _, err := polynomialDecompositionCapacity(params, check.ciphertext)
			if err != nil {
				return result, err
			}
			if ratio > result.WorstCapacityRatio {
				result.WorstCapacityRatio, result.WorstCapacityCheckpoint = ratio, check.ID
			}
		}
	}
	if output != nil {
		ratio, _, err := polynomialDecompositionCapacity(params, output)
		if err != nil {
			return result, err
		}
		if ratio > result.WorstCapacityRatio {
			result.WorstCapacityRatio, result.WorstCapacityCheckpoint = ratio, "F0-final-rescale"
		}
		finalMetric := polynomialDecompositionMetric(reference, mustPolyDecode(params, output), budget)
		if result.FirstCheckpoint == "" && finalMetric.MaxComponent > budget {
			result.FirstCheckpoint = "F0-final-rescale"
			result.CumulativeError = finalMetric
		}
	}
	_ = eval
	return result, nil
}

func mustPolyDecode(params ckks.Parameters, ct *rlwe.Ciphertext) []complex128 {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil
	}
	return values
}

func polynomialDecompositionPrepareBranch(name string, params ckks.Parameters, btp bootstrapping.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, fastInput, standardInput *rlwe.Ciphertext, standardSK *rlwe.SecretKey) (polynomialDecompositionBranchWork, error) {
	work := polynomialDecompositionBranchWork{Summary: polynomialDecompositionBranchSummary{Branch: name}, ActualPowerMap: map[int]*rlwe.Ciphertext{}, OraclePowerMap: map[int]*rlwe.Ciphertext{}, PowerExpected: map[int][]complex128{}, ActualPowerValues: map[int][]complex128{}, OraclePowerValues: map[int][]complex128{}}
	e2, inputValues, err := psGlobalE2(btp, fastEval, fastInput)
	if err != nil {
		return work, fmt.Errorf("%s E2: %w", name, err)
	}
	work.InputValues = inputValues
	work.ReferenceValues = chebyshevRawOracle(fastEval.Mod1Parameters.Mod1Poly, inputValues)
	standardCurrent := standardInput.CopyNew()
	standardCurrent.Scale = standardEval.Mod1Evaluator.Parameters.ScalingFactor()
	if err := standardEval.Evaluator.Add(standardCurrent, evalModCausalOffset(fastEval), standardCurrent); err != nil {
		return work, fmt.Errorf("%s Standard preprocessing: %w", name, err)
	}
	targetScale := psGlobalTargetScale(e2, fastEval)
	standardPoly, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardCurrent, standardEval.Mod1Evaluator.Parameters.Mod1Poly.Clone(), targetScale)
	if err != nil {
		return work, fmt.Errorf("%s Standard polynomial: %w", name, err)
	}
	_, standardValues, err := semanticBisectView(standardPoly, params, standardSK)
	if err != nil {
		return work, fmt.Errorf("%s Standard polynomial decode: %w", name, err)
	}
	work.Summary.D1SourceVsStandard = polynomialDecompositionMetric(work.ReferenceValues, standardValues, polynomialDecompositionOracleTarget)
	productionPowers, _, err := fastEval.PolynomialEvaluator.DiagnosticGeneratePowers(e2, fastEval.Mod1Parameters.Mod1Poly, targetScale)
	if err != nil {
		return work, fmt.Errorf("%s production powers: %w", name, err)
	}
	sort.Slice(productionPowers, func(i, j int) bool { return productionPowers[i].N < productionPowers[j].N })
	memo := map[int][]complex128{}
	for _, snapshot := range productionPowers {
		if snapshot.Ciphertext == nil {
			return work, fmt.Errorf("%s T%d production power is nil", name, snapshot.N)
		}
		decoded, err := psGlobalDecode(params, snapshot.Ciphertext)
		if err != nil {
			return work, fmt.Errorf("%s T%d decode: %w", name, snapshot.N, err)
		}
		expected := fix001ExpectedPower(snapshot.N, inputValues, memo)
		capacityRatio, centered, err := polynomialDecompositionCapacity(params, snapshot.Ciphertext)
		if err != nil {
			return work, fmt.Errorf("%s T%d capacity: %w", name, snapshot.N, err)
		}
		oracle, oracleDecoded, oracleOK, err := polynomialDecompositionOraclePower(params, snapshot.Ciphertext, expected)
		if err != nil {
			return work, fmt.Errorf("%s T%d oracle injection: %w", name, snapshot.N, err)
		}
		work.ActualPowerMap[snapshot.N] = snapshot.Ciphertext
		work.OraclePowerMap[snapshot.N] = oracle
		work.PowerExpected[snapshot.N] = expected
		work.ActualPowerValues[snapshot.N] = decoded
		work.OraclePowerValues[snapshot.N] = oracleDecoded
		work.Summary.ActualPowers = append(work.Summary.ActualPowers, polynomialDecompositionPowerSummary{N: snapshot.N, Level: snapshot.Ciphertext.Level(), Scale: finalizationScaleString(snapshot.Ciphertext.Scale), SourceError: polynomialDecompositionMetric(expected, decoded, polynomialDecompositionPowerTarget), OracleInjectError: polynomialDecompositionMetric(expected, oracleDecoded, polynomialDecompositionOracleTarget), CapacityRatio: capacityRatio, CenteredUnique: centered, OracleRowsMatch: oracleOK})
	}
	if len(work.ActualPowerMap) != len(polynomialDecompositionPowerDegrees) {
		return work, fmt.Errorf("%s production power set incomplete: got %d", name, len(work.ActualPowerMap))
	}
	work.Plan, work.PlanScale, err = polynomialDecompositionPlan(fastEval, work.ActualPowerMap, targetScale)
	if err != nil {
		return work, fmt.Errorf("%s PS plan: %w", name, err)
	}
	work.HActualValues, err = polynomialDecompositionNumericPS(work.Plan, work.ActualPowerValues, work.InputValues)
	if err != nil {
		return work, fmt.Errorf("%s H_actual: %w", name, err)
	}
	work.HOracleValues, err = polynomialDecompositionNumericPS(work.Plan, work.PowerExpected, work.InputValues)
	if err != nil {
		return work, fmt.Errorf("%s H_oracle: %w", name, err)
	}
	work.ActualOutput, work.ActualCiphertext, work.ActualChecks, err = polynomialDecompositionReplay(params, fastEval, work.Plan, work.ActualPowerMap, work.PowerExpected, work.ActualPowerValues, work.InputValues, work.PlanScale)
	if err != nil {
		return work, fmt.Errorf("%s F_actual: %w", name, err)
	}
	work.OracleOutput, work.OracleCiphertext, work.OracleChecks, err = polynomialDecompositionReplay(params, fastEval, work.Plan, work.OraclePowerMap, work.PowerExpected, work.OraclePowerValues, work.InputValues, work.PlanScale)
	if err != nil {
		return work, fmt.Errorf("%s F_oracle: %w", name, err)
	}
	work.Summary.HActual = polynomialDecompositionMetric(work.ReferenceValues, work.HActualValues, polynomialDecompositionBudget)
	work.Summary.HOracle = polynomialDecompositionMetric(work.ReferenceValues, work.HOracleValues, polynomialDecompositionBudget)
	work.Summary.FActual = polynomialDecompositionMetric(work.ReferenceValues, work.ActualOutput, polynomialDecompositionBudget)
	work.Summary.FOracle = polynomialDecompositionMetric(work.ReferenceValues, work.OracleOutput, polynomialDecompositionBudget)
	work.Summary.PowerSemanticResidual = polynomialDecompositionResidualMetric(work.HActualValues, work.HOracleValues, polynomialDecompositionBudget)
	work.Summary.FastPSActualResidual = polynomialDecompositionResidualMetric(work.ActualOutput, work.HActualValues, polynomialDecompositionBudget)
	work.Summary.FastPSOracleResidual = polynomialDecompositionResidualMetric(work.OracleOutput, work.HOracleValues, polynomialDecompositionBudget)
	work.Summary.OracleReferenceResidual = work.Summary.HOracle
	closure := make([]complex128, len(work.ReferenceValues))
	for i := range closure {
		lhs := work.ActualOutput[i] - work.ReferenceValues[i]
		rhs := (work.ActualOutput[i] - work.HActualValues[i]) + (work.HActualValues[i] - work.HOracleValues[i]) + (work.HOracleValues[i] - work.ReferenceValues[i])
		closure[i] = lhs - rhs
	}
	work.Summary.DecompositionClosure = polynomialDecompositionMetric(make([]complex128, len(closure)), closure, polynomialDecompositionOracleTarget)
	work.Summary.FActualVsFOracle = polynomialDecompositionResidualMetric(work.ActualOutput, work.OracleOutput, polynomialDecompositionBudget)
	work.Summary.OraclePSCheckpoint, err = polynomialDecompositionOracleCheckpoint(params, fastEval, work.OracleChecks, work.OracleCiphertext, work.HOracleValues, polynomialDecompositionBudget)
	if err != nil {
		return work, fmt.Errorf("%s oracle PS checkpoint: %w", name, err)
	}
	baselineRealError := polynomialDecompositionComponentMetric(work.ReferenceValues, work.HActualValues, polynomialDecompositionBudget, true).MaxComponent
	baselineImagError := polynomialDecompositionComponentMetric(work.ReferenceValues, work.HActualValues, polynomialDecompositionBudget, false).MaxComponent
	for _, powerN := range []int{2, 3, 4, 6, 8, 16} {
		replaced := make(map[int][]complex128, len(work.ActualPowerValues))
		for n, values := range work.ActualPowerValues {
			replaced[n] = values
		}
		replaced[powerN] = work.PowerExpected[powerN]
		values, err := polynomialDecompositionNumericPS(work.Plan, replaced, work.InputValues)
		if err != nil {
			return work, fmt.Errorf("%s sensitivity T%d: %w", name, powerN, err)
		}
		realMetric := polynomialDecompositionComponentMetric(work.ReferenceValues, values, polynomialDecompositionBudget, true)
		imagMetric := polynomialDecompositionComponentMetric(work.ReferenceValues, values, polynomialDecompositionBudget, false)
		work.Summary.Sensitivity = append(work.Summary.Sensitivity, polynomialDecompositionSensitivity{PowerN: powerN, RealError: realMetric, ImagError: imagMetric, RealErrorReduction: baselineRealError - realMetric.MaxComponent, ImagErrorReduction: baselineImagError - imagMetric.MaxComponent})
	}
	return work, nil
}

func polynomialDecompositionBaseline(cfg BootstrapConfig, btp bootstrapping.Parameters, residual ckks.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, standardSK *rlwe.SecretKey, inputs evalModMatchedInputs) (map[string]interface{}, error) {
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return nil, err
	}
	standardValues, err := decodeWithSecret(residual, standardControl, standardSK)
	if err != nil {
		return nil, err
	}
	fastControl, err := fastEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return nil, err
	}
	fastValues, err := decodeWithSecret(residual, fastControl, zeroSecret(residual))
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"standard_public":              evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), standardValues, polynomialDecompositionPublicThreshold),
		"fast_public":                  evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), fastValues, polynomialDecompositionPublicThreshold),
		"fast_public_metadata_correct": fastControl.Level() == residual.MaxLevel() && fastControl.Degree() == 1 && fastControl.N() == residual.N() && fastControl.IsNTT && !fastControl.IsMontgomery && fastControl.Scale.Equal(residual.DefaultScale()),
	}
	_ = cfg
	return result, nil
}

func polynomialDecompositionDownstreamRun(btp bootstrapping.Parameters, residual ckks.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, inputs evalModMatchedInputs, realWork, imagWork polynomialDecompositionBranchWork, standardSK *rlwe.SecretKey) (*polynomialDecompositionDownstream, error) {
	result := &polynomialDecompositionDownstream{}
	if realWork.Summary.FOracle == nil || imagWork.Summary.FOracle == nil || !realWork.Summary.FOracle.Pass || !imagWork.Summary.FOracle.Pass {
		result.Reason = "F_oracle polynomial error exceeds 1.2e-8; D7 not authorized"
		return result, nil
	}
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastReal, inputs.OrdinaryReal, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, realWork.PlanScale, realWork.OracleCiphertext)
	if err != nil {
		return nil, err
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastImag, inputs.OrdinaryImag, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, imagWork.PlanScale, imagWork.OracleCiphertext)
	if err != nil {
		return nil, err
	}
	if realPath.FastFinal == nil || imagPath.FastFinal == nil {
		result.Reason = fmt.Sprintf("oracle downstream path stopped: real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)
		return result, nil
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return nil, err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return nil, err
	}
	publicLike, standardLike, postS2C, metadata, unchanged, err := precisionSweepPublicLike(realPath, imagPath, fastEval, standardEval, btp, residual.MaxSlots(), standardSK, reproducibleInput(residual, btp))
	if err != nil {
		return nil, err
	}
	result.Reached = true
	result.EvalModReal, result.EvalModImag = realFinal.FastVsStandard, imagFinal.FastVsStandard
	result.PostS2C, result.PublicLike, result.StandardLike = postS2C, publicLike, standardLike
	result.MetadataCorrect, result.InputUnchanged = metadata, unchanged
	return result, nil
}

func polynomialDecompositionClassify(result *polynomialDecompositionResult, realWork, imagWork polynomialDecompositionBranchWork) {
	baselineOK, _ := result.Baseline["production_path_reproduced"].(bool)
	if !baselineOK || result.Baseline["standard_public"].(*PSGlobalMetric).Pass == false || result.Baseline["fast_public"].(*PSGlobalMetric).Pass || realWork.Summary.D1SourceVsStandard == nil || imagWork.Summary.D1SourceVsStandard == nil {
		result.Classification = "logn13_polynomial_precision_decomposition_precondition_mismatch"
		result.FirstSupportedBlocker = "D0"
		return
	}
	if !realWork.Summary.D1SourceVsStandard.Pass || !imagWork.Summary.D1SourceVsStandard.Pass {
		result.Classification = "logn13_polynomial_oracle_alignment_mismatch"
		result.FirstSupportedBlocker = "D1.source_vs_standard"
		return
	}
	for _, work := range []polynomialDecompositionBranchWork{realWork, imagWork} {
		for _, power := range work.Summary.ActualPowers {
			if !power.OracleInjectError.Pass || !power.OracleRowsMatch {
				result.Classification = "logn13_oracle_power_injection_capacity_or_domain_failure"
				result.FirstSupportedBlocker = fmt.Sprintf("%s.T%d.oracle_injection", work.Summary.Branch, power.N)
				return
			}
		}
		if !work.Summary.DecompositionClosure.Pass {
			result.Classification = "logn13_polynomial_precision_interaction_not_isolated"
			result.FirstSupportedBlocker = work.Summary.Branch + ".decomposition_closure"
			return
		}
	}
	hActualFailure := realWork.Summary.HActual.MaxComponent > polynomialDecompositionBudget || imagWork.Summary.HActual.MaxComponent > polynomialDecompositionBudget
	fOraclePass := realWork.Summary.FOracle.MaxComponent <= polynomialDecompositionBudget && imagWork.Summary.FOracle.MaxComponent <= polynomialDecompositionBudget
	if hActualFailure && fOraclePass {
		result.Classification = "logn13_polynomial_precision_generated_power_blocker"
		result.FirstSupportedBlocker = "H_actual"
	} else if !hActualFailure && !fOraclePass {
		result.Classification = "logn13_polynomial_precision_ps_arithmetic_blocker"
		result.FirstSupportedBlocker = "F_oracle"
	} else if hActualFailure && !fOraclePass {
		result.Classification = "logn13_polynomial_precision_joint_power_and_ps_blocker"
		result.FirstSupportedBlocker = "H_actual_and_F_oracle"
	} else {
		result.Classification = "logn13_polynomial_precision_interaction_not_isolated"
		result.FirstSupportedBlocker = "isolated_paths"
	}
	result.CausalDisposition = fmt.Sprintf("H_actual real=%g imag=%g; F_oracle real=%g imag=%g; budget=%g", realWork.Summary.HActual.MaxComponent, imagWork.Summary.HActual.MaxComponent, realWork.Summary.FOracle.MaxComponent, imagWork.Summary.FOracle.MaxComponent, polynomialDecompositionBudget)
}

func polynomialDecompositionWrite(result polynomialDecompositionResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DiagLogN13PolynomialPrecisionDecomposition(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, standardSK, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return err
	}
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := polynomialDecompositionResult{SchemaVersion: "fix-001-p3-diag-logn13-polynomial-precision-decomposition.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, PlanScale: finalizationScaleString(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), polynomialDecompositionPlanScaleExponent))), Branches: map[string]polynomialDecompositionBranchSummary{}, Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3PolynomialDecompositionPrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3PolynomialDecompositionSecondary, "no_secondary_production_changes": true, "no_production_integration": true, "no_guarded_power_production_change": true, "no_mixed_or_per_block_scale": true, "no_plan_scale_change": true, "no_c2s_change": true, "no_s2c_change": true, "no_mod1_parameter_retuning": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_invalid_raw_standard_da_q01_projection": true}}
	if !secondaryClean || secondaryCommit != requiredFIX001P3PolynomialDecompositionSecondary {
		result.Classification = "logn13_polynomial_precision_decomposition_precondition_mismatch"
		result.FirstSupportedBlocker = "secondary_provenance"
		return polynomialDecompositionWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, standardSK)
	if err != nil {
		return err
	}
	baseline, err := polynomialDecompositionBaseline(cfg, btp, residual, fastEval, standardEval, standardSK, inputs)
	if err != nil {
		return err
	}
	result.Baseline = baseline
	realWork, err := polynomialDecompositionPrepareBranch("real", btp.BootstrappingParameters, btp, fastEval, standardEval, inputs.FastReal, inputs.OrdinaryReal, standardSK)
	if err != nil {
		return err
	}
	imagWork, err := polynomialDecompositionPrepareBranch("imag", btp.BootstrappingParameters, btp, fastEval, standardEval, inputs.FastImag, inputs.OrdinaryImag, standardSK)
	if err != nil {
		return err
	}
	result.Branches["real"] = realWork.Summary
	result.Branches["imag"] = imagWork.Summary
	realPath, err := evalModMatchedRunPath(inputs.FastReal, inputs.OrdinaryReal, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return err
	}
	imagPath, err := evalModMatchedRunPath(inputs.FastImag, inputs.OrdinaryImag, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return err
	}
	result.Baseline["production_polynomial_real"] = realPath.Polynomial["fast_vs_standard"]
	result.Baseline["production_polynomial_imag"] = imagPath.Polynomial["fast_vs_standard"]
	productionEvalModOK := false
	if realPath.FastFinal != nil && imagPath.FastFinal != nil {
		realFinal, finalErr := evalModMatchedFinalEvidence(realPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
		if finalErr != nil {
			return finalErr
		}
		imagFinal, finalErr := evalModMatchedFinalEvidence(imagPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
		if finalErr != nil {
			return finalErr
		}
		result.Baseline["production_evalmod_real"] = realFinal.FastVsStandard
		result.Baseline["production_evalmod_imag"] = imagFinal.FastVsStandard
		productionEvalModOK = realFinal.FastVsStandard != nil && imagFinal.FastVsStandard != nil && realFinal.FastVsStandard.MaxComponent > 1e-3 && realFinal.FastVsStandard.MaxComponent < 1e-2 && imagFinal.FastVsStandard.MaxComponent > 1e-3 && imagFinal.FastVsStandard.MaxComponent < 1e-2
	}
	realPolynomial, _ := result.Baseline["production_polynomial_real"].(*PSGlobalMetric)
	imagPolynomial, _ := result.Baseline["production_polynomial_imag"].(*PSGlobalMetric)
	result.Baseline["production_path_reproduced"] = realPolynomial != nil && imagPolynomial != nil && realPolynomial.MaxComponent < 2e-6 && imagPolynomial.MaxComponent < 2e-6 && productionEvalModOK
	polynomialDecompositionClassify(&result, realWork, imagWork)
	if result.Classification == "logn13_polynomial_precision_generated_power_blocker" || result.Classification == "logn13_polynomial_precision_ps_arithmetic_blocker" || result.Classification == "logn13_polynomial_precision_joint_power_and_ps_blocker" {
		result.Downstream, err = polynomialDecompositionDownstreamRun(btp, residual, fastEval, standardEval, inputs, realWork, imagWork, standardSK)
		if err != nil {
			return err
		}
	}
	return polynomialDecompositionWrite(result, outPath)
}
