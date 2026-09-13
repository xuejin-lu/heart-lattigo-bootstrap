package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3GuardedPowerPrimaryBase = "72cc6a44d4ce25b11f1e77cbd77d018f21e4aa68"
	requiredFIX001P3GuardedPowerSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	guardedPowerPlanScaleExponent           = 91
	guardedPowerSemanticTarget              = 2e-8
	guardedPowerEvalModTarget               = 1e-4
	guardedPowerPublicThreshold             = 1e-2
)

var guardedPowerExponents = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

type guardedPowerMetric struct {
	MaxComponent float64 `json:"max_component_abs"`
	MaxComplex   float64 `json:"max_abs_complex"`
	MeanComplex  float64 `json:"mean_abs_complex"`
	WorstIndex   int     `json:"worst_index"`
	WorstPart    string  `json:"worst_component"`
	Pass         bool    `json:"pass_target"`
	Target       float64 `json:"target"`
}

type guardedPowerCheckpoint struct {
	Name                 string              `json:"name"`
	Level                int                 `json:"level"`
	Scale                string              `json:"scale"`
	Degree               int                 `json:"degree"`
	FastCapacityRatio    float64             `json:"fast_q01_capacity_ratio"`
	FullRNSCapacityRatio float64             `json:"full_rns_capacity_ratio"`
	FastOutsideCount     int                 `json:"fast_q01_outside_count"`
	FullRNSOutsideCount  int                 `json:"full_rns_outside_count"`
	CenteredUnique       bool                `json:"centered_unique"`
	RowsMatch            bool                `json:"fast_vs_full_rns_rows_match"`
	Semantic             *guardedPowerMetric `json:"semantic_vs_source_oracle"`
}

type guardedPowerCandidate struct {
	GuardExponent            int                            `json:"guard_exponent"`
	FactorSummaries          map[string]string              `json:"selected_factor_summaries,omitempty"`
	FirstFailure             string                         `json:"first_failure"`
	Classification           string                         `json:"classification"`
	WorstCapacityCheckpoint  string                         `json:"worst_capacity_checkpoint"`
	WorstCapacityRatio       float64                        `json:"worst_capacity_ratio"`
	CapacityOutsideCount     int                            `json:"capacity_outside_count"`
	AllCenteredUnique        bool                           `json:"all_centered_unique"`
	AllRowsMatch             bool                           `json:"all_fast_vs_full_rns_rows_match"`
	GeneratedPowerErrors     map[string]*guardedPowerMetric `json:"generated_power_max_errors,omitempty"`
	GeneratedPowersPass      bool                           `json:"generated_powers_precision_guidance_pass"`
	PolynomialError          *PSGlobalMetric                `json:"compressed_polynomial_vs_standard,omitempty"`
	PolynomialOracleError    *PSGlobalMetric                `json:"compressed_polynomial_vs_source_oracle,omitempty"`
	EvalModRealError         *PSGlobalMetric                `json:"evalmod_real_vs_standard,omitempty"`
	EvalModImagError         *PSGlobalMetric                `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2CImplementation    *PSGlobalMetric                `json:"post_s2c_implementation_effect,omitempty"`
	PostS2CVsStandard        *PSGlobalMetric                `json:"post_s2c_vs_standard,omitempty"`
	PublicLikeError          *PSGlobalMetric                `json:"public_like_vs_message,omitempty"`
	PublicLikeMetadata       bool                           `json:"public_like_metadata_correct"`
	PublicLikeInputUnchanged bool                           `json:"public_like_input_unchanged"`
	PrecisionQualified       bool                           `json:"precision_qualified"`
	PublicLikeAccepted       bool                           `json:"public_like_accepted"`
	Checkpoints              []guardedPowerCheckpoint       `json:"-"`
}

type guardedPowerResult struct {
	SchemaVersion           string                  `json:"schema_version"`
	Timestamp               time.Time               `json:"timestamp"`
	Primary                 RepositoryMetadata      `json:"primary_repository"`
	Lattigo                 RepositoryMetadata      `json:"lattigo_repository"`
	Environment             EnvironmentMetadata     `json:"environment"`
	Config                  BootstrapConfig         `json:"config"`
	Parameters              ExperimentParameters    `json:"effective_parameters"`
	Workload                CorrectnessWorkload     `json:"workload"`
	PublicThreshold         float64                 `json:"public_threshold"`
	FixedPlanScale          string                  `json:"fixed_production_plan_scale"`
	Baseline                map[string]interface{}  `json:"baseline"`
	Candidates              []guardedPowerCandidate `json:"candidates"`
	SelectedGuardExponent   int                     `json:"selected_guard_exponent,omitempty"`
	CausalDisposition       string                  `json:"causal_disposition"`
	Classification          string                  `json:"classification"`
	FirstLimitingCheckpoint string                  `json:"first_limiting_checkpoint"`
	Validation              map[string]interface{}  `json:"validation"`
}

type guardedPowerRun struct {
	Powers          map[int]*rlwe.Ciphertext
	Expected        map[int][]complex128
	Decoded         map[int][]complex128
	Errors          map[int]*guardedPowerMetric
	Checkpoints     []guardedPowerCheckpoint
	FactorSummaries map[string]string
	FirstFailure    string
	Classification  string
	WorstRatio      float64
	WorstCheckpoint string
	OutsideCount    int
	AllCentered     bool
	AllRows         bool
}

type guardedPowerStop struct {
	Classification string
	Checkpoint     string
}

func (e *guardedPowerStop) Error() string { return e.Classification + " at " + e.Checkpoint }

func guardedPowerMetricFor(reference, actual []complex128) *guardedPowerMetric {
	metric := psGlobalMetric(reference, actual)
	return &guardedPowerMetric{MaxComponent: metric.MaxComponent, MaxComplex: metric.MaxComplex, MeanComplex: metric.MeanComplex, WorstIndex: metric.WorstIndex, WorstPart: metric.WorstPart, Pass: metric.MaxComponent <= guardedPowerSemanticTarget, Target: guardedPowerSemanticTarget}
}

func guardedPowerFactors(q *big.Int, guard int) (*big.Int, *big.Int, error) {
	if q == nil || q.Sign() <= 0 || guard < 0 {
		return nil, nil, fmt.Errorf("invalid guarded factor inputs")
	}
	target := new(big.Int).Lsh(new(big.Int).Set(q), uint(guard))
	root := new(big.Int).Sqrt(target)
	var bestLeft, bestRight, bestAbs, bestMax *big.Int
	for delta := -3; delta <= 3; delta++ {
		left := new(big.Int).Add(root, big.NewInt(int64(delta)))
		if left.Sign() <= 0 {
			continue
		}
		quotient, remainder := new(big.Int).QuoRem(target, left, new(big.Int))
		for _, right := range []*big.Int{quotient, new(big.Int).Add(quotient, big.NewInt(1))} {
			if right.Sign() <= 0 || (remainder.Sign() != 0 && right.Cmp(quotient) == 0) {
				continue
			}
			product := new(big.Int).Mul(left, right)
			abs := new(big.Int).Abs(new(big.Int).Sub(new(big.Int).Set(product), target))
			maxFactor := new(big.Int).Set(left)
			if right.Cmp(maxFactor) > 0 {
				maxFactor.Set(right)
			}
			better := bestAbs == nil || abs.Cmp(bestAbs) < 0
			if !better && abs.Cmp(bestAbs) == 0 {
				better = maxFactor.Cmp(bestMax) < 0 || (maxFactor.Cmp(bestMax) == 0 && (left.Cmp(bestLeft) < 0 || (left.Cmp(bestLeft) == 0 && right.Cmp(bestRight) < 0)))
			}
			if better {
				bestLeft, bestRight, bestAbs, bestMax = new(big.Int).Set(left), new(big.Int).Set(right), abs, maxFactor
			}
		}
	}
	if bestLeft == nil {
		return nil, nil, fmt.Errorf("cannot derive guarded factors for %s", target)
	}
	return bestLeft, bestRight, nil
}

func guardedPowerRoundSigned(value *big.Int, guard int) *big.Int {
	if guard == 0 {
		return new(big.Int).Set(value)
	}
	divisor := new(big.Int).Lsh(big.NewInt(1), uint(guard))
	abs := new(big.Int).Abs(new(big.Int).Set(value))
	quotient, remainder := new(big.Int).QuoRem(abs, divisor, new(big.Int))
	if new(big.Int).Lsh(new(big.Int).Set(remainder), 1).Cmp(divisor) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if value.Sign() < 0 {
		quotient.Neg(quotient)
	}
	return quotient
}

// guardedPowerRoundDivPow2Maintained performs exact signed, ties-away-from-zero
// rounding on the authoritative q0/q1 rows. It never consumes a level and
// deliberately does not inspect dormant q2+ rows.
func guardedPowerRoundDivPow2Maintained(params ckks.Parameters, source *rlwe.Ciphertext, guard int) (*rlwe.Ciphertext, error) {
	if source == nil || source.Level() < 1 || !source.IsNTT || !source.IsMontgomery {
		return nil, fmt.Errorf("guarded round-div requires NTT Montgomery level >= 1 ciphertext")
	}
	if guard == 0 {
		return source.CopyNew(), nil
	}
	recovered, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, err
	}
	q01 := new(big.Int).Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus))
	q01Half := new(big.Int).Rsh(new(big.Int).Set(q01), 1)
	result := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*result.MetaData = *source.MetaData
	result.IsNTT, result.IsMontgomery = source.IsNTT, source.IsMontgomery
	result.Scale = source.Scale.Div(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(guard))))
	for component, values := range recovered {
		for _, value := range values {
			if new(big.Int).Abs(value).Cmp(q01Half) >= 0 {
				return nil, fmt.Errorf("guarded round-div centered capacity failure")
			}
		}
		normalNTT := postMod1S2CLiftSigned(params, func() []*big.Int {
			out := make([]*big.Int, len(values))
			for i, value := range values {
				out[i] = guardedPowerRoundSigned(value, guard)
			}
			return out
		}(), source.Level())
		for limb := 0; limb < 2; limb++ {
			params.RingQ().SubRings[limb].MForm(normalNTT.Coeffs[limb], result.Value[component].Coeffs[limb])
		}
	}
	return result, nil
}

func guardedPowerSubAligned(params ckks.Parameters, eval *fastckks.Evaluator, out, sub *rlwe.Ciphertext) error {
	if out.Scale.Equal(sub.Scale) {
		return eval.Sub(out, sub, out)
	}
	if out.Scale.Cmp(sub.Scale) > 0 {
		scratch := fastckks.NewCiphertext(params, sub.Degree(), minInt(out.Level(), sub.Level()))
		psGlobalCopyMaintained(params, sub, scratch)
		if err := eval.MulIntegerMaintained(sub, out.Scale.Div(sub.Scale).BigInt(), scratch); err != nil {
			return err
		}
		scratch.Scale = out.Scale
		return eval.Sub(out, scratch, out)
	}
	scratch := fastckks.NewCiphertext(params, out.Degree(), minInt(out.Level(), sub.Level()))
	psGlobalCopyMaintained(params, out, scratch)
	if err := eval.MulIntegerMaintained(out, sub.Scale.Div(out.Scale).BigInt(), scratch); err != nil {
		return err
	}
	scratch.Scale = sub.Scale
	if err := eval.Sub(scratch, sub, scratch); err != nil {
		return err
	}
	psGlobalCopyMaintained(params, scratch, out)
	return nil
}

func guardedPowerCheckpointFor(params ckks.Parameters, name string, ct *rlwe.Ciphertext, expected []complex128) (guardedPowerCheckpoint, error) {
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, ct)
	if err != nil {
		return guardedPowerCheckpoint{}, fmt.Errorf("%s fast capacity: %w", name, err)
	}
	full, _, err := postMod1S2CLiftFull(params, ct)
	if err != nil {
		return guardedPowerCheckpoint{}, fmt.Errorf("%s full lift: %w", name, err)
	}
	fullCapacity, err := postMod1S2CCapacityFromCiphertext(params, full)
	if err != nil {
		return guardedPowerCheckpoint{}, fmt.Errorf("%s full capacity: %w", name, err)
	}
	decoded, err := psGlobalDecode(params, ct)
	if err != nil {
		return guardedPowerCheckpoint{}, fmt.Errorf("%s semantic decode: %w", name, err)
	}
	return guardedPowerCheckpoint{Name: name, Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), Degree: ct.Degree(), FastCapacityRatio: fastCapacity.MaxAbsOverQ01Half, FullRNSCapacityRatio: fullCapacity.MaxAbsOverQ01Half, FastOutsideCount: fastCapacity.OutsideCount, FullRNSOutsideCount: fullCapacity.OutsideCount, CenteredUnique: fastCapacity.Pass && fullCapacity.Pass, RowsMatch: postMod1S2CRowsEqual(params, ct, full), Semantic: guardedPowerMetricFor(expected, decoded)}, nil
}

func guardedPowerRecord(run *guardedPowerRun, params ckks.Parameters, name string, ct *rlwe.Ciphertext, expected []complex128) (*guardedPowerCheckpoint, error) {
	cp, err := guardedPowerCheckpointFor(params, name, ct, expected)
	if err != nil {
		return nil, err
	}
	run.Checkpoints = append(run.Checkpoints, cp)
	if cp.FastCapacityRatio > run.WorstRatio {
		run.WorstRatio, run.WorstCheckpoint = cp.FastCapacityRatio, name
	}
	if cp.FullRNSCapacityRatio > run.WorstRatio {
		run.WorstRatio, run.WorstCheckpoint = cp.FullRNSCapacityRatio, name
	}
	run.OutsideCount += cp.FastOutsideCount + cp.FullRNSOutsideCount
	run.AllCentered = run.AllCentered && cp.CenteredUnique
	run.AllRows = run.AllRows && cp.RowsMatch
	return &run.Checkpoints[len(run.Checkpoints)-1], nil
}

func guardedPowerExpectedMul(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] * b[i]
	}
	return out
}

func guardedPowerExpectedDouble(a []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = 2 * a[i]
	}
	return out
}

func guardedPowerRunFor(params ckks.Parameters, eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, z []complex128, polyExpected map[int][]complex128, guard int) (guardedPowerRun, error) {
	run := guardedPowerRun{Powers: map[int]*rlwe.Ciphertext{1: input.CopyNew()}, Expected: map[int][]complex128{1: append([]complex128(nil), z...)}, Decoded: map[int][]complex128{}, Errors: map[int]*guardedPowerMetric{}, FactorSummaries: map[string]string{}, WorstRatio: 0, AllCentered: true, AllRows: true}
	decoded, err := psGlobalDecode(params, input)
	if err != nil {
		return run, fmt.Errorf("T1 semantic decode: %w", err)
	}
	run.Decoded[1] = decoded
	if _, err = guardedPowerRecord(&run, params, "T1.final", input, z); err != nil {
		return run, fmt.Errorf("T1 checkpoint: %w", err)
	}
	var generate func(int, bool) error
	generate = func(n int, lazy bool) error {
		if run.Powers[n] != nil {
			return nil
		}
		a, b := commonpolynomial.SplitDegree(n)
		isPow2 := n&(n-1) == 0
		if err := generate(a, lazy && !isPow2); err != nil {
			return err
		}
		if err := generate(b, lazy && !isPow2); err != nil {
			return err
		}
		left, right := run.Powers[a], run.Powers[b]
		if !lazy && (left.Degree() > 1 || right.Degree() > 1) {
			return fmt.Errorf("guarded power %d reached non-relinearized strict operand", n)
		}
		commonLevel := minInt(left.Level(), right.Level())
		q := new(big.Int).SetUint64(params.RingQ().SubRings[commonLevel].Modulus)
		factorLeft, factorRight, err := guardedPowerFactors(q, guard)
		if err != nil {
			return err
		}
		run.FactorSummaries[fmt.Sprintf("T%d", n)] = fmt.Sprintf("left=%s,right=%s,target=%s", factorLeft, factorRight, new(big.Int).Lsh(new(big.Int).Set(q), uint(guard)))
		leftExpected, rightExpected := run.Expected[a], run.Expected[b]
		if _, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.input_left", n), left, leftExpected); err != nil {
			return err
		}
		if _, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.input_right", n), right, rightExpected); err != nil {
			return err
		}
		leftCopy := fastckks.NewCiphertext(params, left.Degree(), commonLevel)
		rightCopy := fastckks.NewCiphertext(params, right.Degree(), commonLevel)
		psGlobalCopyMaintained(params, left, leftCopy)
		psGlobalCopyMaintained(params, right, rightCopy)
		if err := eval.FastCKKS.MulIntegerMaintained(leftCopy, factorLeft, leftCopy); err != nil {
			return err
		}
		if err := eval.FastCKKS.MulIntegerMaintained(rightCopy, factorRight, rightCopy); err != nil {
			return err
		}
		leftCopy.Scale = left.Scale.Mul(rlwe.NewScale(factorLeft))
		rightCopy.Scale = right.Scale.Mul(rlwe.NewScale(factorRight))
		leftFactorExpected := leftExpected
		rightFactorExpected := rightExpected
		leftCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.after_factor_left", n), leftCopy, leftFactorExpected)
		if err != nil {
			return err
		}
		if !leftCP.CenteredUnique {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.after_factor_left", n), "guard_factor_pre_rescale_capacity_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		rightCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.after_factor_right", n), rightCopy, rightFactorExpected)
		if err != nil {
			return err
		}
		if !rightCP.CenteredUnique {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.after_factor_right", n), "guard_factor_pre_rescale_capacity_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		if err := eval.FastCKKS.Rescale(leftCopy, leftCopy); err != nil {
			return err
		}
		if err := eval.FastCKKS.Rescale(rightCopy, rightCopy); err != nil {
			return err
		}
		leftCP, err = guardedPowerRecord(&run, params, fmt.Sprintf("T%d.after_rescale_left", n), leftCopy, leftFactorExpected)
		if err != nil {
			return err
		}
		if !leftCP.CenteredUnique {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.after_rescale_left", n), "guard_factor_pre_rescale_capacity_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		rightCP, err = guardedPowerRecord(&run, params, fmt.Sprintf("T%d.after_rescale_right", n), rightCopy, rightFactorExpected)
		if err != nil {
			return err
		}
		if !rightCP.CenteredUnique {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.after_rescale_right", n), "guard_factor_pre_rescale_capacity_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		degree := 1
		if lazy {
			degree = 2
		}
		out := fastckks.NewCiphertext(params, degree, commonLevel-1)
		*out.MetaData = *leftCopy.MetaData
		if lazy {
			err = eval.FastCKKS.Mul(leftCopy, rightCopy, out)
		} else {
			err = eval.FastCKKS.MulRelin(leftCopy, rightCopy, out)
		}
		if err != nil {
			return err
		}
		rawExpected := guardedPowerExpectedMul(leftExpected, rightExpected)
		rawCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.raw_product", n), out, rawExpected)
		if err != nil {
			return err
		}
		if !rawCP.CenteredUnique {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.raw_product", n), "guard_product_capacity_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		if err := eval.FastCKKS.Add(out, out, out); err != nil {
			return err
		}
		doubledExpected := guardedPowerExpectedDouble(rawExpected)
		doubledCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.doubled_guarded", n), out, doubledExpected)
		if err != nil {
			return err
		}
		if !doubledCP.CenteredUnique {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.doubled_guarded", n), "guard_product_capacity_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		normalized, err := guardedPowerRoundDivPow2Maintained(params, out, guard)
		if err != nil {
			return err
		}
		normalizedCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.after_round_div", n), normalized, doubledExpected)
		if err != nil {
			return err
		}
		if !normalizedCP.RowsMatch {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.after_round_div", n), "guard_round_div_primitive_mismatch"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		c := a - b
		if c < 0 {
			c = -c
		}
		if c == 0 {
			if err := eval.FastCKKS.Add(normalized, -1, normalized); err != nil {
				return err
			}
		} else {
			if err := generate(c, lazy); err != nil {
				return err
			}
			if err := guardedPowerSubAligned(params, eval.FastCKKS, normalized, run.Powers[c]); err != nil {
				run.FirstFailure, run.Classification = fmt.Sprintf("T%d.recurrence_alignment", n), "guard_recurrence_alignment_failure"
				return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
			}
		}
		expected := polyExpected[n]
		if expected == nil {
			expected = fix001ExpectedPower(n, z, polyExpected)
		}
		recurrenceCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.after_recurrence", n), normalized, expected)
		if err != nil {
			return err
		}
		if !recurrenceCP.RowsMatch {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.after_recurrence", n), "guard_recurrence_alignment_failure"
			return &guardedPowerStop{Classification: run.Classification, Checkpoint: run.FirstFailure}
		}
		run.Powers[n], run.Expected[n] = normalized, expected
		decoded, err := psGlobalDecode(params, normalized)
		if err != nil {
			return err
		}
		run.Decoded[n] = decoded
		run.Errors[n] = guardedPowerMetricFor(expected, decoded)
		finalCP, err := guardedPowerRecord(&run, params, fmt.Sprintf("T%d.final", n), normalized, expected)
		if err != nil {
			return err
		}
		if !finalCP.Semantic.Pass && run.FirstFailure == "" {
			run.FirstFailure, run.Classification = fmt.Sprintf("T%d.final", n), "guard_generated_power_semantic_failure"
		}
		return nil
	}
	if err := generate(16, false); err != nil {
		if _, ok := err.(*guardedPowerStop); !ok {
			return run, fmt.Errorf("guarded dependency generation: %w", err)
		}
	}
	// The LogN13 degree-30 PS decomposition asks for the odd/non-power-of-two
	// branches separately after the power-of-two spine has been built.
	for _, n := range []int{6, 3} {
		if err := generate(n, false); err != nil {
			if _, ok := err.(*guardedPowerStop); !ok {
				return run, fmt.Errorf("guarded auxiliary dependency generation T%d: %w", n, err)
			}
			break
		}
	}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		if run.Errors[n] == nil {
			decoded, err := psGlobalDecode(params, run.Powers[n])
			if err != nil {
				return run, fmt.Errorf("T%d final semantic decode: %w", n, err)
			}
			run.Decoded[n] = decoded
			run.Errors[n] = guardedPowerMetricFor(run.Expected[n], decoded)
		}
	}
	if run.FirstFailure == "" {
		run.Classification = "guard_generated_powers_pass"
	}
	return run, nil
}

func guardedPowerPolynomial(eval *bootstrapping.FastEvaluator, params ckks.Parameters, input *rlwe.Ciphertext, polyExpected map[int][]complex128, z []complex128, guard int) (guardedPowerRun, *rlwe.Ciphertext, *PSGlobalMetric, error) {
	run, err := guardedPowerRunFor(params, eval, input, z, polyExpected, guard)
	if err != nil {
		return run, nil, nil, err
	}
	if run.FirstFailure != "" && !strings.HasPrefix(run.Classification, "guard_generated_power_semantic_failure") {
		return run, nil, nil, nil
	}
	poly := eval.Mod1Parameters.Mod1Poly
	target := psGlobalTargetScale(input, eval)
	sim := psGlobalSim{params: params, levels: params.LevelsConsumedPerRescaling()}
	plan := commonpolynomial.NewPolynomial(poly).PatersonStockmeyerPolynomial(eval.FastCKKS.GetRLWEParameters(), input.Level(), input.Scale, target, sim)
	planScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), guardedPowerPlanScaleExponent))
	for i := range plan.Value {
		plan.Value[i].Scale = planScale
	}
	_, _, root, _, err := psGlobalReplay(params, eval.FastCKKS, plan, run.Powers, run.Expected, run.Decoded, z, planScale)
	if err != nil {
		run.FirstFailure, run.Classification = "PS.replay", "guard_generated_power_semantic_failure"
		return run, nil, nil, nil
	}
	if root.ciphertext == nil {
		return run, nil, nil, fmt.Errorf("guarded PS root ciphertext is nil")
	}
	output := root.ciphertext.CopyNew()
	if err := eval.FastCKKS.Rescale(output, output); err != nil {
		return run, nil, nil, err
	}
	oracle := psGlobalPolyExpected(poly, z)
	decoded, err := psGlobalDecode(params, output)
	if err != nil {
		return run, nil, nil, fmt.Errorf("guarded PS output semantic decode: %w", err)
	}
	return run, output, evalModMatchedMetricFor(oracle, decoded, guardedPowerEvalModTarget), nil
}

func guardedPowerMetricMap(runs ...guardedPowerRun) map[string]*guardedPowerMetric {
	result := map[string]*guardedPowerMetric{}
	for _, run := range runs {
		for _, n := range []int{2, 3, 4, 6, 8, 16} {
			if run.Errors[n] != nil {
				result[fmt.Sprintf("T%d", n)] = run.Errors[n]
			}
		}
	}
	return result
}

func guardedPowerBaseline(cfg BootstrapConfig, primaryRoot, backendRoot string, residual ckks.Parameters, btp bootstrapping.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, inputs evalModMatchedInputs) (map[string]interface{}, evalModMatchedPath, evalModMatchedPath, error) {
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline Standard Bootstrap: %w", err)
	}
	standardValues, err := decodeWithSecret(residual, standardControl, inputs.StandardSK)
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline Standard decode: %w", err)
	}
	standardMetric := evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), standardValues, guardedPowerPublicThreshold)
	fastControl, err := fastEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline Fast Bootstrap: %w", err)
	}
	fastValues, err := decodeWithSecret(residual, fastControl, zeroSecret(residual))
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline Fast decode: %w", err)
	}
	fastMetric := evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), fastValues, guardedPowerPublicThreshold)
	realPath, err := evalModMatchedRunPath(inputs.FastReal, inputs.OrdinaryReal, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline real EvalMod path: %w", err)
	}
	imagPath, err := evalModMatchedRunPath(inputs.FastImag, inputs.OrdinaryImag, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline imag EvalMod path: %w", err)
	}
	if realPath.FastFinal == nil || imagPath.FastFinal == nil {
		return nil, realPath, imagPath, fmt.Errorf("baseline path stopped before final: real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline real final evidence: %w", err)
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK)
	if err != nil {
		return nil, evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("baseline imag final evidence: %w", err)
	}
	baseline := map[string]interface{}{"standard_public": standardMetric, "fast_public": fastMetric, "fast_public_metadata_correct": fastControl.Level() == residual.MaxLevel() && fastControl.Degree() == 1 && fastControl.N() == residual.N() && fastControl.IsNTT && !fastControl.IsMontgomery && fastControl.Scale.Equal(residual.DefaultScale()), "c2s_real": inputs.FastVsOrdinaryReal, "c2s_imag": inputs.FastVsOrdinaryImag, "evalmod_real": realFinal.FastVsStandard, "evalmod_imag": imagFinal.FastVsStandard, "polynomial_real": realPath.Polynomial["fast_vs_standard"], "polynomial_imag": imagPath.Polynomial["fast_vs_standard"]}
	return baseline, realPath, imagPath, nil
}

func runFIX001P3DesignLogN13GuardedPowerPrecision(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("load bootstrap parameters: %w", err)
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return fmt.Errorf("build Fast evaluator: %w", err)
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, skN2, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return fmt.Errorf("build genuine Standard evaluator: %w", err)
	}
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := guardedPowerResult{SchemaVersion: "fix-001-p3-design-logn13-guarded-power-precision.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, PublicThreshold: guardedPowerPublicThreshold, FixedPlanScale: finalizationScaleString(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), guardedPowerPlanScaleExponent))), SelectedGuardExponent: -1, Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3GuardedPowerPrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3GuardedPowerSecondary, "no_secondary_production_changes": true, "no_production_fix": true, "no_plan_scale_change": true, "no_c2s_change": true, "no_s2c_change": true, "no_parameter_retuning": true, "no_extra_level_consumption": true, "no_raw_standard_double_angle_q01_oracle": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if !secondaryClean || secondaryCommit != requiredFIX001P3GuardedPowerSecondary {
		result.Classification = "logn13_guarded_power_precondition_mismatch"
		return guardedPowerWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, skN2)
	if err != nil {
		return fmt.Errorf("matched C2S baseline: %w", err)
	}
	baseline, _, _, err := guardedPowerBaseline(cfg, primaryRoot, backendRoot, residual, btp, fastEval, standardEval, inputs)
	if err != nil {
		return fmt.Errorf("guarded-power G0 baseline: %w", err)
	}
	result.Baseline = baseline
	standardPublic, _ := baseline["standard_public"].(*PSGlobalMetric)
	fastPublic, _ := baseline["fast_public"].(*PSGlobalMetric)
	if standardPublic == nil || !standardPublic.Pass || fastPublic == nil || fastPublic.Pass || inputs.FastVsOrdinaryReal == nil || inputs.FastVsOrdinaryImag == nil || !inputs.FastVsOrdinaryReal.Pass || !inputs.FastVsOrdinaryImag.Pass {
		result.Classification = "logn13_guarded_power_precondition_mismatch"
		result.CausalDisposition = "baseline control did not reproduce accepted Standard/C2S/public behavior"
		return guardedPowerWrite(result, outPath)
	}
	poly := fastEval.Mod1Parameters.Mod1Poly
	e2Baseline, zBaseline, err := psGlobalE2(btp, fastEval, inputs.FastReal)
	if err != nil {
		return fmt.Errorf("baseline guarded power input: %w", err)
	}
	baselineTarget := psGlobalTargetScale(e2Baseline, fastEval)
	baselinePowers, _, err := fastEval.PolynomialEvaluator.DiagnosticGeneratePowers(e2Baseline, poly, baselineTarget)
	if err != nil {
		return fmt.Errorf("baseline production power generation: %w", err)
	}
	result.Baseline["source_backed_power_errors"] = map[string]*PSGlobalMetric{}
	baselineMemo := map[int][]complex128{}
	for _, power := range baselinePowers {
		decoded, err := psGlobalDecode(btp.BootstrappingParameters, power.Ciphertext)
		if err != nil {
			return fmt.Errorf("baseline T%d decode: %w", power.N, err)
		}
		expected := fix001ExpectedPower(power.N, zBaseline, baselineMemo)
		if power.N == 2 || power.N == 3 || power.N == 4 || power.N == 6 || power.N == 8 || power.N == 16 {
			result.Baseline["source_backed_power_errors"].(map[string]*PSGlobalMetric)[fmt.Sprintf("T%d", power.N)] = evalModMatchedMetricFor(expected, decoded, guardedPowerSemanticTarget)
		}
	}
	result.Candidates = make([]guardedPowerCandidate, 0, len(guardedPowerExponents))
	for _, guard := range guardedPowerExponents {
		candidate := guardedPowerCandidate{GuardExponent: guard, FirstFailure: "none", Classification: "guard_generated_powers_pass", GeneratedPowersPass: true, AllCenteredUnique: true, AllRowsMatch: true}
		e2Real, zReal, err := psGlobalE2(btp, fastEval, inputs.FastReal)
		if err != nil {
			return fmt.Errorf("guard %d real power input: %w", guard, err)
		}
		e2Imag, zImag, err := psGlobalE2(btp, fastEval, inputs.FastImag)
		if err != nil {
			return fmt.Errorf("guard %d imag power input: %w", guard, err)
		}
		realExpected := map[int][]complex128{}
		realMemo := map[int][]complex128{}
		for _, n := range []int{1, 2, 3, 4, 6, 8, 16} {
			realExpected[n] = fix001ExpectedPower(n, zReal, realMemo)
		}
		runReal, guardedReal, polyOracleMetric, err := guardedPowerPolynomial(fastEval, btp.BootstrappingParameters, e2Real, realExpected, zReal, guard)
		if err != nil {
			return fmt.Errorf("guard %d real guarded power replay: %w", guard, err)
		}
		imagExpected := map[int][]complex128{}
		for _, n := range []int{1, 2, 3, 4, 6, 8, 16} {
			imagExpected[n] = fix001ExpectedPower(n, zImag, imagExpected)
		}
		runImag, guardedImag, _, imagErr := guardedPowerPolynomial(fastEval, btp.BootstrappingParameters, e2Imag, imagExpected, zImag, guard)
		if imagErr != nil {
			return fmt.Errorf("guard %d imag guarded power replay: %w", guard, imagErr)
		}
		candidate.FactorSummaries = runReal.FactorSummaries
		candidate.Checkpoints = append(runReal.Checkpoints, runImag.Checkpoints...)
		candidate.GeneratedPowerErrors = guardedPowerMetricMap(runReal, runImag)
		candidate.GeneratedPowersPass = true
		for _, metric := range candidate.GeneratedPowerErrors {
			candidate.GeneratedPowersPass = candidate.GeneratedPowersPass && metric.Pass
		}
		candidate.AllCenteredUnique = runReal.AllCentered && runImag.AllCentered
		candidate.AllRowsMatch = runReal.AllRows && runImag.AllRows
		candidate.CapacityOutsideCount = runReal.OutsideCount + runImag.OutsideCount
		candidate.WorstCapacityRatio, candidate.WorstCapacityCheckpoint = runReal.WorstRatio, runReal.WorstCheckpoint
		if runImag.WorstRatio > candidate.WorstCapacityRatio {
			candidate.WorstCapacityRatio, candidate.WorstCapacityCheckpoint = runImag.WorstRatio, runImag.WorstCheckpoint
		}
		if runReal.FirstFailure != "" {
			candidate.FirstFailure, candidate.Classification = runReal.FirstFailure, runReal.Classification
		}
		if runImag.FirstFailure != "" && candidate.FirstFailure == "none" {
			candidate.FirstFailure, candidate.Classification = runImag.FirstFailure, runImag.Classification
		}
		if guardedReal == nil || guardedImag == nil || !candidate.AllCenteredUnique || !candidate.AllRowsMatch {
			if candidate.FirstFailure == "none" {
				candidate.FirstFailure, candidate.Classification = "generated_power_capacity", "guard_product_capacity_failure"
			}
			candidate.GeneratedPowersPass = false
			result.Candidates = append(result.Candidates, candidate)
			continue
		}
		candidate.PolynomialOracleError = polyOracleMetric
		fastPathReal, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastReal, inputs.OrdinaryReal, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK, rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), guardedPowerPlanScaleExponent)), guardedReal)
		if err != nil {
			return fmt.Errorf("guard %d real normalized EvalMod path: %w", guard, err)
		}
		fastPathImag, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastImag, inputs.OrdinaryImag, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK, rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), guardedPowerPlanScaleExponent)), guardedImag)
		if err != nil {
			return fmt.Errorf("guard %d imag normalized EvalMod path: %w", guard, err)
		}
		if fastPathReal.FastFinal == nil || fastPathImag.FastFinal == nil {
			candidate.PrecisionQualified = false
			candidate.PublicLikeAccepted = false
			candidate.FirstFailure = "normalized_double_angle"
			if fastPathReal.FirstFailure != "none" {
				candidate.FirstFailure = "real." + fastPathReal.FirstFailure
			} else if fastPathImag.FirstFailure != "none" {
				candidate.FirstFailure = "imag." + fastPathImag.FirstFailure
			}
			candidate.Classification = "guard_generated_power_semantic_failure"
			result.Candidates = append(result.Candidates, candidate)
			continue
		}
		candidate.PolynomialError = fastPathReal.Polynomial["fast_vs_standard"]
		realFinal, err := evalModMatchedFinalEvidence(fastPathReal, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK)
		if err != nil {
			return fmt.Errorf("guard %d real final evidence: %w", guard, err)
		}
		imagFinal, err := evalModMatchedFinalEvidence(fastPathImag, btp.BootstrappingParameters, inputs.FastSK, inputs.StandardSK)
		if err != nil {
			return fmt.Errorf("guard %d imag final evidence: %w", guard, err)
		}
		candidate.EvalModRealError, candidate.EvalModImagError = realFinal.FastVsStandard, imagFinal.FastVsStandard
		candidate.PrecisionQualified = candidate.EvalModRealError != nil && candidate.EvalModImagError != nil && candidate.EvalModRealError.MaxComponent <= guardedPowerEvalModTarget && candidate.EvalModImagError.MaxComponent <= guardedPowerEvalModTarget
		if runReal.FirstFailure != "" && candidate.Classification == "guard_generated_powers_pass" {
			candidate.Classification = runReal.Classification
		}
		if candidate.PrecisionQualified {
			publicLike, _, postS2C, metadataCorrect, inputUnchanged, err := precisionSweepPublicLike(fastPathReal, fastPathImag, fastEval, standardEval, btp, residual.MaxSlots(), skN2, reproducibleInput(residual, btp))
			if err != nil {
				return err
			}
			candidate.PublicLikeError, candidate.PostS2CVsStandard = publicLike, postS2C
			candidate.PublicLikeMetadata, candidate.PublicLikeInputUnchanged = metadataCorrect, inputUnchanged
			candidate.PublicLikeAccepted = publicLike != nil && publicLike.MaxComponent <= guardedPowerPublicThreshold && metadataCorrect && inputUnchanged
			fullReal, _, err := postMod1S2CLiftFull(btp.BootstrappingParameters, fastPathReal.FastFinal)
			if err != nil {
				return err
			}
			fullImag, _, err := postMod1S2CLiftFull(btp.BootstrappingParameters, fastPathImag.FastFinal)
			if err != nil {
				return err
			}
			standardDFT, err := newPostMod1StandardDFT(btp.BootstrappingParameters, fastEval.S2CDFTMatrix)
			if err != nil {
				return err
			}
			state := chebyshevOracleState{Params: btp, Eval: fastEval}
			impl, fullOut, _, failure, err := postMod1S2CReplay(state, fastPathReal.FastFinal, fastPathImag.FastFinal, fullReal, fullImag, standardDFT)
			if err != nil {
				return err
			}
			if failure != nil {
				candidate.PrecisionQualified, candidate.PublicLikeAccepted = false, false
				candidate.Classification, candidate.FirstFailure = "normalized_candidate_primitive_mismatch", failure.Checkpoint
			} else {
				implValues, err := decodeWithSecret(btp.BootstrappingParameters, impl, zeroSecret(btp.BootstrappingParameters))
				if err != nil {
					return err
				}
				fullValues, err := decodeWithSecret(btp.BootstrappingParameters, fullOut, zeroSecret(btp.BootstrappingParameters))
				if err != nil {
					return err
				}
				candidate.PostS2CImplementation = evalModMatchedMetricFor(fullValues, implValues, guardedPowerPublicThreshold)
			}
		}
		if candidate.Classification == "guard_generated_powers_pass" && !candidate.GeneratedPowersPass {
			candidate.Classification = "guard_generated_power_semantic_failure"
		}
		result.Candidates = append(result.Candidates, candidate)
	}
	result.Classification = "logn13_guarded_power_precision_improved_but_insufficient"
	result.CausalDisposition = "guarded generated-power replay completed without production integration; compare generated-power, PS, EvalMod and public-like error rows"
	result.FirstLimitingCheckpoint = "generated_power_semantic"
	bestT16, bestPoly := math.Inf(1), math.Inf(1)
	hasPolynomialFailure, hasEvalModFailure, hasPublicFailure := false, false, false
	for _, candidate := range result.Candidates {
		if metric := candidate.GeneratedPowerErrors["T16"]; metric != nil && metric.MaxComponent < bestT16 {
			bestT16 = metric.MaxComponent
		}
		if candidate.PolynomialError != nil && candidate.PolynomialError.MaxComponent < bestPoly {
			bestPoly = candidate.PolynomialError.MaxComponent
		}
		if candidate.PolynomialError != nil && candidate.PolynomialError.MaxComponent > 1.2e-8 {
			hasPolynomialFailure = true
		}
		if (candidate.EvalModRealError != nil && candidate.EvalModRealError.MaxComponent > guardedPowerEvalModTarget) || (candidate.EvalModImagError != nil && candidate.EvalModImagError.MaxComponent > guardedPowerEvalModTarget) {
			hasEvalModFailure = true
		}
		if candidate.PublicLikeError != nil && candidate.PublicLikeError.MaxComponent > guardedPowerPublicThreshold {
			hasPublicFailure = true
		}
		if candidate.PublicLikeAccepted && result.SelectedGuardExponent < 0 {
			result.SelectedGuardExponent, result.Classification, result.FirstLimitingCheckpoint = candidate.GuardExponent, "logn13_guarded_power_precision_candidate_validated", "none"
		}
	}
	if result.SelectedGuardExponent < 0 {
		switch {
		case hasPublicFailure:
			result.FirstLimitingCheckpoint = "public_like_threshold"
		case hasEvalModFailure:
			result.FirstLimitingCheckpoint = "evalmod_precision_threshold"
		case hasPolynomialFailure:
			result.FirstLimitingCheckpoint = "polynomial_precision_budget"
		}
	}
	if math.IsInf(bestT16, 1) {
		bestT16 = 0
	}
	if math.IsInf(bestPoly, 1) {
		bestPoly = 0
	}
	result.Validation["best_t16_error"] = bestT16
	result.Validation["best_polynomial_error"] = bestPoly
	return guardedPowerWrite(result, outPath)
}

func guardedPowerWrite(result guardedPowerResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}
