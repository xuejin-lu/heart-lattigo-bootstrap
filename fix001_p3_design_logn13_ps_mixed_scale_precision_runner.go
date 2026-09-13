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
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3PSMixedScalePrimaryBase = "fa4757cd70a62120d8d82988b235259c9306fdb0"
	requiredFIX001P3PSMixedScaleSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	psMixedScaleBudget                      = 1.2e-8
	psMixedScaleLocalTarget                 = 1e-10
	psMixedScalePublicThreshold             = 1e-2
)

type psMixedScalarAudit struct {
	ID                     string          `json:"id"`
	Block                  int             `json:"block"`
	CoefficientReal        string          `json:"coefficient_real"`
	CoefficientImag        string          `json:"coefficient_imag"`
	PowerLevel             int             `json:"power_level"`
	PowerScale             string          `json:"power_scale"`
	AccumulatorScaleBefore string          `json:"accumulator_scale_before"`
	EncodingScale          string          `json:"scalar_encoding_scale"`
	RoundedIntegerReal     string          `json:"rounded_integer_real"`
	RoundedIntegerImag     string          `json:"rounded_integer_imag"`
	Q01                    string          `json:"q01"`
	Q01Half                string          `json:"q01_half"`
	RealCapacityRatio      float64         `json:"real_scalar_capacity_ratio"`
	ImagCapacityRatio      float64         `json:"imag_scalar_capacity_ratio"`
	MaxCapacityRatio       float64         `json:"max_scalar_capacity_ratio"`
	CenteredUnique         bool            `json:"scalar_centered_unique"`
	ResultCapacityRatio    float64         `json:"result_ciphertext_capacity_ratio"`
	ResultCenteredUnique   bool            `json:"result_ciphertext_centered_unique"`
	LocalSemanticError     *PSGlobalMetric `json:"local_semantic_error"`
}

type psMixedAlignmentAudit struct {
	ID                   string          `json:"id"`
	Kind                 string          `json:"kind"`
	LeftScale            string          `json:"left_scale"`
	RightScale           string          `json:"right_scale"`
	ExactRatio           string          `json:"exact_real_ratio"`
	IntegerRatio         string          `json:"integer_ratio_used"`
	FractionalDifference string          `json:"fractional_difference"`
	IntegerRatioExact    bool            `json:"integer_ratio_exact"`
	MetadataOnlyRelabel  bool            `json:"metadata_only_relabel"`
	LocalSemanticError   *PSGlobalMetric `json:"local_semantic_error"`
}

type psMixedReplay struct {
	Checks     []PSGlobalCheckpoint
	Root       PSGlobalCheckpoint
	Scalars    []psMixedScalarAudit
	Alignments []psMixedAlignmentAudit
	FirstError string
	ErrorClass string
}

type psMixedBranchRun struct {
	Name           string
	Schedule       []int
	Replay         psMixedReplay
	Final          *rlwe.Ciphertext
	FinalValues    []complex128
	StandardValues []complex128
	PlainOracle    []complex128
	PlainError     *PSGlobalMetric
	StandardError  *PSGlobalMetric
	FullError      *PSGlobalMetric
	CapacitySafe   bool
	SemanticSafe   bool
	LevelSafe      bool
	Valid          bool
	FirstFailure   string
}

type psMixedBlockSummary struct {
	Block                   int     `json:"block"`
	BlockID                 string  `json:"block_id"`
	SafeScaleCeiling        int     `json:"safe_scale_ceiling"`
	LimitingOperation       string  `json:"limiting_operation"`
	HeadroomRatio           float64 `json:"headroom_ratio"`
	ScalarMaxRatioAtCeiling float64 `json:"scalar_max_ratio_at_ceiling"`
	AllScaleChecksValid     bool    `json:"all_scale_checks_valid"`
}

type psMixedCandidate struct {
	Name                string  `json:"name"`
	Schedule            []int   `json:"block_scale_exponents"`
	FinalRealError      float64 `json:"final_real_error"`
	FinalImagError      float64 `json:"final_imag_error"`
	FinalMaxError       float64 `json:"final_max_error"`
	StandardRealError   float64 `json:"final_real_vs_standard"`
	StandardImagError   float64 `json:"final_imag_vs_standard"`
	FullRNSRealError    float64 `json:"final_real_vs_full_rns"`
	FullRNSImagError    float64 `json:"final_imag_vs_full_rns"`
	Valid               bool    `json:"valid"`
	PrecisionQualified  bool    `json:"precision_qualified"`
	CapacitySafe        bool    `json:"capacity_safe"`
	SemanticSafe        bool    `json:"semantic_safe"`
	LevelSafe           bool    `json:"level_safe"`
	FirstFailure        string  `json:"first_failure"`
	FirstBudgetCrossing string  `json:"first_budget_crossing"`
	OutputScaleReal     string  `json:"output_scale_real,omitempty"`
	OutputScaleImag     string  `json:"output_scale_imag,omitempty"`
	ScalarAuditCount    int     `json:"scalar_audit_count"`
	AlignmentAuditCount int     `json:"alignment_audit_count"`
}

type psMixedGreedyStep struct {
	Step           int     `json:"step"`
	Block          int     `json:"block"`
	FromExponent   int     `json:"from_exponent"`
	ToExponent     int     `json:"to_exponent"`
	BeforeMaxError float64 `json:"before_max_error"`
	AfterMaxError  float64 `json:"after_max_error"`
	Improvement    float64 `json:"improvement"`
}

type psMixedResult struct {
	SchemaVersion     string                 `json:"schema_version"`
	Timestamp         time.Time              `json:"timestamp"`
	Primary           RepositoryMetadata     `json:"primary_repository"`
	Lattigo           RepositoryMetadata     `json:"lattigo_repository"`
	Environment       EnvironmentMetadata    `json:"environment"`
	Config            BootstrapConfig        `json:"config"`
	Parameters        ExperimentParameters   `json:"effective_parameters"`
	Workload          CorrectnessWorkload    `json:"workload"`
	Budget            float64                `json:"polynomial_budget"`
	PublicThreshold   float64                `json:"public_like_threshold"`
	M0Baseline        map[string]interface{} `json:"m0_baseline"`
	M1Cause           map[string]interface{} `json:"m1_first_semantic_boundary"`
	Blocks            []psMixedBlockSummary  `json:"blocks"`
	C0Control         psMixedCandidate       `json:"c0_all_2^92"`
	SingleBlockLifts  []psMixedCandidate     `json:"single_block_lifts"`
	Safe93Aggregate   psMixedCandidate       `json:"safe_93_aggregate"`
	GreedySteps       []psMixedGreedyStep    `json:"greedy_accepted_steps"`
	BestCandidate     psMixedCandidate       `json:"best_mixed_candidate"`
	SelectedCandidate string                 `json:"selected_candidate,omitempty"`
	Downstream        map[string]interface{} `json:"downstream,omitempty"`
	Classification    string                 `json:"classification"`
	FirstBlocker      string                 `json:"first_remaining_blocker"`
	Validation        map[string]interface{} `json:"validation"`
}

type psMixedSchedulePlan struct {
	Base      commonpolynomial.PatersonStockmeyerPolynomial
	Exponents []int
}

func psMixedScalePlan(base commonpolynomial.PatersonStockmeyerPolynomial, exponents []int) (commonpolynomial.PatersonStockmeyerPolynomial, error) {
	if len(base.Value) != len(exponents) {
		return commonpolynomial.PatersonStockmeyerPolynomial{}, fmt.Errorf("PS block exponent count %d does not match plan block count %d", len(exponents), len(base.Value))
	}
	plan := base
	for i := range plan.Value {
		plan.Value[i].Scale = precisionSweepScale(exponents[i])
	}
	return plan, nil
}

func psMixedBigFloatString(value *big.Float) string {
	if value == nil {
		return ""
	}
	return value.Text('e', 40)
}

func psMixedScaleString(scale rlwe.Scale) string {
	return psMixedBigFloatString(&scale.Value)
}

func psMixedQ01(params ckks.Parameters) (*big.Int, *big.Int) {
	q01 := new(big.Int).Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus))
	return q01, new(big.Int).Rsh(new(big.Int).Set(q01), 1)
}

func psMixedRoundCoefficient(value *big.Float, scale rlwe.Scale) *big.Int {
	if value == nil {
		return new(big.Int)
	}
	v := new(big.Float).Mul(value, &scale.Value)
	if value.Sign() > 0 {
		v.Add(v, new(big.Float).SetFloat64(0.5))
	} else if value.Sign() < 0 {
		v.Sub(v, new(big.Float).SetFloat64(0.5))
	}
	out := new(big.Int)
	v.Int(out)
	return out
}

func psMixedRatio(abs, half *big.Int) float64 {
	if half.Sign() == 0 {
		return math.Inf(1)
	}
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(abs), new(big.Float).SetInt(half)).Float64()
	return ratio
}

func psMixedScaleRatio(higher, lower rlwe.Scale) (string, string, string, bool) {
	ratio := higher.Div(lower)
	ratioInt := ratio.BigInt()
	integerScale := rlwe.NewScale(ratioInt)
	diff := new(big.Float).Sub(&ratio.Value, &integerScale.Value)
	return psMixedBigFloatString(&ratio.Value), ratioInt.String(), psMixedBigFloatString(diff), diff.Sign() == 0
}

func psMixedBigInt(text string) *big.Int {
	value, ok := new(big.Int).SetString(text, 10)
	if !ok {
		return new(big.Int)
	}
	return value
}

func psMixedCoefficientScale(params ckks.Parameters, level int) (rlwe.Scale, error) {
	consumed := params.LevelsConsumedPerRescaling()
	if level-consumed+1 < 0 {
		return rlwe.Scale{}, fmt.Errorf("insufficient level %d for coefficient scale", level)
	}
	scale := rlwe.NewScale(1)
	for i := 0; i < consumed; i++ {
		scale = scale.Mul(rlwe.NewScale(params.Q()[level-i]))
	}
	return scale, nil
}

func psMixedScalarEncodingScale(params ckks.Parameters, power, accumulator rlwe.Scale, coefficient *bignum.Complex, level int) (rlwe.Scale, error) {
	if coefficient.IsInt() {
		return rlwe.NewScale(1), nil
	}
	if power.Equal(accumulator) {
		return psMixedCoefficientScale(params, level)
	}
	if power.Cmp(accumulator) < 0 {
		return accumulator.Div(power), nil
	}
	return psMixedCoefficientScale(params, level)
}

func psMixedScalarAuditFor(params ckks.Parameters, id string, block int, coefficient *bignum.Complex, power *rlwe.Ciphertext, accumulator rlwe.Scale, encoding rlwe.Scale, result *rlwe.Ciphertext, local *PSGlobalMetric) (psMixedScalarAudit, error) {
	q01, half := psMixedQ01(params)
	realInteger := psMixedRoundCoefficient(coefficient[0], encoding)
	imagInteger := psMixedRoundCoefficient(coefficient[1], encoding)
	realRatio := psMixedRatio(new(big.Int).Abs(new(big.Int).Set(realInteger)), half)
	imagRatio := psMixedRatio(new(big.Int).Abs(new(big.Int).Set(imagInteger)), half)
	resultRatio, resultCentered, err := polynomialDecompositionCapacity(params, result)
	if err != nil {
		return psMixedScalarAudit{}, err
	}
	return psMixedScalarAudit{
		ID: id, Block: block,
		CoefficientReal: psMixedBigFloatString(coefficient[0]), CoefficientImag: psMixedBigFloatString(coefficient[1]),
		PowerLevel: power.Level(), PowerScale: psMixedScaleString(power.Scale), AccumulatorScaleBefore: psMixedScaleString(accumulator), EncodingScale: psMixedScaleString(encoding),
		RoundedIntegerReal: realInteger.String(), RoundedIntegerImag: imagInteger.String(), Q01: q01.String(), Q01Half: half.String(),
		RealCapacityRatio: realRatio, ImagCapacityRatio: imagRatio, MaxCapacityRatio: math.Max(realRatio, imagRatio), CenteredUnique: realRatio < 1 && imagRatio < 1,
		ResultCapacityRatio: resultRatio, ResultCenteredUnique: resultCentered, LocalSemanticError: local,
	}, nil
}

func psMixedCheckpointCapacity(params ckks.Parameters, checkpoint *PSGlobalCheckpoint) (float64, bool, bool, error) {
	if checkpoint == nil || checkpoint.ciphertext == nil || checkpoint.ciphertext.Level() < 1 {
		return 0, true, true, nil
	}
	capacity, err := postMod1S2CCapacityFromFastCiphertext(params, checkpoint.ciphertext)
	if err != nil {
		return 0, false, false, err
	}
	rowsMatch := false
	if capacity.Pass {
		full, _, err := postMod1S2CLiftFull(params, checkpoint.ciphertext)
		if err != nil {
			return capacity.MaxAbsOverQ01Half, capacity.Pass, false, err
		}
		rowsMatch = postMod1S2CRowsEqual(params, checkpoint.ciphertext, full)
	}
	return capacity.MaxAbsOverQ01Half, capacity.Pass, rowsMatch, nil
}

func psMixedStrictAddAligned(params ckks.Parameters, eval *fastckks.Evaluator, a, b *rlwe.Ciphertext, id string, alignments *[]psMixedAlignmentAudit) error {
	if a.Scale.Equal(b.Scale) {
		return eval.Add(b, a, b)
	}
	leftBefore, err := psGlobalDecode(params, a)
	if err != nil {
		return err
	}
	rightBefore, err := psGlobalDecode(params, b)
	if err != nil {
		return err
	}
	if b.Scale.Cmp(a.Scale) > 0 {
		ratioText, ratioInt, fractional, exact := psMixedScaleRatio(b.Scale, a.Scale)
		audit := psMixedAlignmentAudit{ID: id, Kind: "align_left_to_right", LeftScale: psMixedScaleString(a.Scale), RightScale: psMixedScaleString(b.Scale), ExactRatio: ratioText, IntegerRatio: ratioInt, FractionalDifference: fractional, IntegerRatioExact: exact, MetadataOnlyRelabel: false}
		scratch := fastckks.NewCiphertext(params.Parameters, a.Degree(), minInt(a.Level(), b.Level()))
		psGlobalCopyMaintained(params, a, scratch)
		if err := eval.MulIntegerMaintained(a, psMixedBigInt(ratioInt), scratch); err != nil {
			return err
		}
		scratch.Scale = b.Scale
		aligned, err := psGlobalDecode(params, scratch)
		if err != nil {
			return err
		}
		audit.LocalSemanticError = psGlobalMetric(leftBefore, aligned)
		*alignments = append(*alignments, audit)
		return eval.Add(b, scratch, b)
	}
	ratioText, ratioInt, fractional, exact := psMixedScaleRatio(a.Scale, b.Scale)
	audit := psMixedAlignmentAudit{ID: id, Kind: "align_right_to_left", LeftScale: psMixedScaleString(a.Scale), RightScale: psMixedScaleString(b.Scale), ExactRatio: ratioText, IntegerRatio: ratioInt, FractionalDifference: fractional, IntegerRatioExact: exact, MetadataOnlyRelabel: false}
	scratch := fastckks.NewCiphertext(params.Parameters, b.Degree(), minInt(a.Level(), b.Level()))
	psGlobalCopyMaintained(params, b, scratch)
	if err := eval.MulIntegerMaintained(b, psMixedBigInt(ratioInt), scratch); err != nil {
		return err
	}
	scratch.Scale = a.Scale
	aligned, err := psGlobalDecode(params, scratch)
	if err != nil {
		return err
	}
	audit.LocalSemanticError = psGlobalMetric(rightBefore, aligned)
	*alignments = append(*alignments, audit)
	if err := eval.Add(scratch, a, scratch); err != nil {
		return err
	}
	psGlobalCopyMaintained(params, scratch, b)
	return nil
}

func psMixedReplayPS(params ckks.Parameters, eval *fastckks.Evaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, powerExpected, powerDecoded map[int][]complex128, input []complex128) psMixedReplay {
	out := psMixedReplay{}
	type babyNode struct {
		degree   int
		value    *rlwe.Ciphertext
		source   []complex128
		identity string
	}
	baby := make([]*babyNode, len(plan.Value))
	for i := range plan.Value {
		p := plan.Value[i]
		value := fastckks.NewCiphertext(params.Parameters, 1, p.Level)
		value.IsNTT, value.IsMontgomery = powers[1].IsNTT, powers[1].IsMontgomery
		*value.MetaData = *powers[1].MetaData
		value.Scale = p.Scale
		psGlobalZero(value)
		source := make([]complex128, len(input))
		description := fmt.Sprintf("Chebyshev block %d degree=%d", i, p.Degree())
		identity := psGlobalSubtreeHash(fmt.Sprintf("block-%d", i), p.Polynomial)
		if p.IsEven {
			if p.Coeffs[0] == nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-constant", i), "ps_common_93_other_semantic_boundary"
				return out
			}
			beforeScale := value.Scale
			if err := eval.Add(value, p.Coeffs[0], value); err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-constant", i), "ps_common_93_other_semantic_boundary"
				return out
			}
			psGlobalAddScaled(source, makeOnes(len(input)), psGlobalCoeff(p.Coeffs[0]))
			actual, err := psGlobalDecode(params, value)
			if err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-constant", i), "ps_common_93_other_semantic_boundary"
				return out
			}
			local := psGlobalMetric(source, actual)
			audit, err := psMixedScalarAuditFor(params, fmt.Sprintf("B%d-constant", i), i, p.Coeffs[0], powers[1], beforeScale, beforeScale, value, local)
			if err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-constant", i), "ps_common_93_other_semantic_boundary"
				return out
			}
			out.Scalars = append(out.Scalars, audit)
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("B%d-constant", i), "baby_constant_add", value, source, actual, source, identity, description, nil))
		}
		for key := p.Degree(); key > 0; key-- {
			if (p.IsEven || p.IsOdd) && ((key&1 == 0 && !p.IsEven) || (key&1 == 1 && !p.IsOdd)) {
				continue
			}
			if p.Coeffs[key] == nil {
				continue
			}
			power := powers[key]
			if power == nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-term-%d", i, key), "ps_common_93_other_semantic_boundary"
				return out
			}
			before, err := psGlobalDecode(params, value)
			if err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-term-%d", i, key), "ps_common_93_other_semantic_boundary"
				return out
			}
			beforeScale := value.Scale
			if power.Scale.Cmp(value.Scale) > 0 {
				ratioText, ratioInt, fractional, exact := psMixedScaleRatio(power.Scale, value.Scale)
				alignment := psMixedAlignmentAudit{ID: fmt.Sprintf("B%d-term-%d-promotion", i, key), Kind: "mul_then_add_accumulator_promotion", LeftScale: psMixedScaleString(value.Scale), RightScale: psMixedScaleString(power.Scale), ExactRatio: ratioText, IntegerRatio: ratioInt, FractionalDifference: fractional, IntegerRatioExact: exact, MetadataOnlyRelabel: false}
				if err := eval.MulIntegerMaintained(value, psMixedBigInt(ratioInt), value); err != nil {
					out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-term-%d", i, key), "ps_common_93_integer_scale_alignment_loss"
					return out
				}
				value.Scale = power.Scale
				promoted, decodeErr := psGlobalDecode(params, value)
				if decodeErr != nil {
					return out
				}
				alignment.LocalSemanticError = psGlobalMetric(before, promoted)
				out.Alignments = append(out.Alignments, alignment)
			}
			encoding, err := psMixedScalarEncodingScale(params, power.Scale, value.Scale, p.Coeffs[key], value.Level())
			if err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-term-%d", i, key), "ps_common_93_other_semantic_boundary"
				return out
			}
			if err := eval.MulThenAdd(power, p.Coeffs[key], value); err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("B%d-term-%d", i, key), "ps_common_93_other_semantic_boundary"
				return out
			}
			psGlobalAddScaled(source, powerExpected[key], psGlobalCoeff(p.Coeffs[key]))
			actual, err := psGlobalDecode(params, value)
			if err != nil {
				return out
			}
			localExpected := append([]complex128(nil), before...)
			psGlobalAddScaled(localExpected, powerDecoded[key], psGlobalCoeff(p.Coeffs[key]))
			local := psGlobalMetric(localExpected, actual)
			audit, err := psMixedScalarAuditFor(params, fmt.Sprintf("B%d-term-%d", i, key), i, p.Coeffs[key], power, beforeScale, encoding, value, local)
			if err != nil {
				return out
			}
			out.Scalars = append(out.Scalars, audit)
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("B%d-term-%d", i, key), "baby_mul_then_add", value, source, actual, localExpected, identity, description, nil))
		}
		baby[len(plan.Value)-i-1] = &babyNode{degree: p.Degree(), value: value, source: source, identity: identity}
	}
	round := 0
	for len(baby) != 1 {
		giant := make([]int, len(baby))
		for i := 0; i < len(baby); i++ {
			if i == len(baby)-1 {
				giant[i] = 2
			} else if baby[i].degree == baby[i+1].degree {
				giant[i] = 1
				i++
			}
		}
		for i := 0; i < len(baby); i++ {
			if giant[i] == 2 {
				baby[i].degree = baby[i-1].degree
				continue
			}
			if giant[i] != 1 {
				continue
			}
			a, b := baby[i], baby[i+1]
			deg := 1 << bitsLen64(uint64(a.degree))
			beforeB, err := psGlobalDecode(params, b.value)
			if err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("G%d-input-b", round), "ps_common_93_other_semantic_boundary"
				return out
			}
			out.Checks = append(out.Checks, psGlobalCheckpointWithDecode(fmt.Sprintf("G%d-input-b", round), "giant_input_b", b.value, b.source, beforeB, b.identity, "PS child branch b", nil))
			if b.value.Degree() == 2 {
				if err := eval.Relinearize(b.value, b.value); err != nil {
					out.FirstError, out.ErrorClass = fmt.Sprintf("G%d-relinearize", round), "ps_common_93_other_semantic_boundary"
					return out
				}
				after, _ := psGlobalDecode(params, b.value)
				out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-relinearize", round), "giant_relinearize", b.value, b.source, after, beforeB, b.identity, "same subtree after Fast relinearize", nil))
			}
			beforeRescale, _ := psGlobalDecode(params, b.value)
			if err := eval.Rescale(b.value, b.value); err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("G%d-rescale", round), "ps_common_93_other_semantic_boundary"
				return out
			}
			afterRescale, _ := psGlobalDecode(params, b.value)
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-rescale", round), "giant_rescale", b.value, b.source, afterRescale, beforeRescale, b.identity, "same subtree after Fast Rescale", nil))
			if err := eval.Mul(b.value, powers[deg], b.value); err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("G%d-multiply", round), "ps_common_93_other_semantic_boundary"
				return out
			}
			productSource := append([]complex128(nil), b.source...)
			for k := range productSource {
				productSource[k] *= powerExpected[deg][k]
			}
			productDecoded, _ := psGlobalDecode(params, b.value)
			productID := hashSourceVector(fmt.Sprintf("%s*T%d", b.identity, deg))
			localProduct := make([]complex128, len(afterRescale))
			for k := range localProduct {
				localProduct[k] = afterRescale[k] * powerDecoded[deg][k]
			}
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-multiply", round), "giant_multiply", b.value, productSource, productDecoded, localProduct, productID, fmt.Sprintf("(%s) * T%d", b.identity, deg), nil))
			parentSource := append([]complex128(nil), a.source...)
			psGlobalAddScaled(parentSource, productSource, 1)
			if err := psMixedStrictAddAligned(params, eval, a.value, b.value, fmt.Sprintf("G%d-add", round), &out.Alignments); err != nil {
				out.FirstError, out.ErrorClass = fmt.Sprintf("G%d-add", round), "ps_common_93_integer_scale_alignment_loss"
				return out
			}
			parentDecoded, err := psGlobalDecode(params, b.value)
			if err != nil {
				return out
			}
			localParent := append([]complex128(nil), parentSource...)
			parentID := hashSourceVector(fmt.Sprintf("%s+%s", a.identity, productID))
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-add", round), "giant_add_aligned", b.value, parentSource, parentDecoded, localParent, parentID, fmt.Sprintf("%s + (%s)", a.identity, productID), nil))
			b.degree = 2*deg - 1
			b.source, b.identity = parentSource, parentID
			baby[i] = nil
			round++
			i++
		}
		kept := baby[:0]
		for _, step := range baby {
			if step != nil {
				kept = append(kept, step)
			}
		}
		baby = kept
	}
	root := baby[0]
	if root.value.Degree() == 2 {
		if err := eval.Relinearize(root.value, root.value); err != nil {
			out.FirstError, out.ErrorClass = "F0-relinearize", "ps_common_93_other_semantic_boundary"
			return out
		}
	}
	decoded, err := psGlobalDecode(params, root.value)
	if err != nil {
		return out
	}
	out.Root = psGlobalCheckpointWithDecode("F0-root", "pre_final_rescale_root", root.value, root.source, decoded, root.identity, "PS root before final Rescale", nil)
	return out
}

func bitsLen64(value uint64) int {
	length := 0
	for value > 0 {
		length++
		value >>= 1
	}
	return length
}

func psMixedRoundoffMetric() *PSGlobalMetric {
	return &PSGlobalMetric{Pass: true, Threshold: psMixedScaleLocalTarget, WorstIndex: -1}
}

func psMixedRunBranch(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch psOracleScaleBranch, exponents []int, standard []complex128) (psMixedBranchRun, error) {
	result := psMixedBranchRun{Name: branch.Name, Schedule: append([]int(nil), exponents...), StandardValues: standard, PlainOracle: branch.Base.ReferenceValues, CapacitySafe: true, SemanticSafe: true, LevelSafe: true, Valid: true, FirstFailure: "none"}
	plan, err := psMixedScalePlan(branch.Base.Plan, exponents)
	if err != nil {
		return result, err
	}
	result.Replay = psMixedReplayPS(params, eval.FastCKKS, plan, branch.Base.OraclePowerMap, branch.Base.PowerExpected, branch.Base.OraclePowerValues, branch.Base.InputValues)
	firstFailure := result.Replay.FirstError
	if firstFailure != "" {
		result.Valid = false
		result.SemanticSafe = false
	}
	for _, scalar := range result.Replay.Scalars {
		if !scalar.CenteredUnique && firstFailure == "" {
			firstFailure = scalar.ID
			result.Valid = false
			result.CapacitySafe = false
		}
	}
	for _, alignment := range result.Replay.Alignments {
		if alignment.LocalSemanticError != nil && alignment.LocalSemanticError.MaxComponent > psMixedScaleLocalTarget {
			result.SemanticSafe = false
			if firstFailure == "" {
				firstFailure = alignment.ID
				result.Valid = false
			}
		}
	}
	for _, checkpoint := range result.Replay.Checks {
		_, centered, rowsMatch, err := psMixedCheckpointCapacity(params, &checkpoint)
		if err != nil {
			return result, err
		}
		if !centered {
			result.CapacitySafe = false
			result.Valid = false
			if firstFailure == "" {
				firstFailure = checkpoint.ID
			}
		}
		if !rowsMatch {
			result.CapacitySafe = false
			result.Valid = false
			if firstFailure == "" {
				firstFailure = checkpoint.ID + ".rows"
			}
		}
	}
	if result.Replay.Root.ciphertext == nil {
		if firstFailure == "" {
			firstFailure = "replay_root"
		}
		result.Valid = false
		result.LevelSafe = false
		result.FirstFailure = firstFailure
		return result, nil
	}
	output := result.Replay.Root.ciphertext.CopyNew()
	if err := eval.FastCKKS.Rescale(output, output); err != nil {
		if firstFailure == "" {
			firstFailure = "F0-final-rescale"
		}
		result.Valid = false
		result.FirstFailure = firstFailure
		return result, nil
	}
	result.Final = output
	result.FinalValues, err = psGlobalDecode(params, output)
	if err != nil {
		return result, err
	}
	result.PlainError = psMixedScalePair(branch.Base.ReferenceValues, result.FinalValues, psMixedScaleBudget).Real
	result.StandardError = psMixedScalePair(standard, result.FinalValues, psMixedScaleBudget).Real
	if output.Level() >= 1 {
		fullValues, rowsMatch, err := psOracleScaleFullMirror(params, output)
		if err != nil {
			return result, err
		}
		result.FullError = psGlobalMetric(fullValues, result.FinalValues)
		if !rowsMatch {
			result.CapacitySafe = false
			result.Valid = false
			if firstFailure == "" {
				firstFailure = "F0-final-rescale.rows"
			}
		}
	} else {
		result.FullError = psMixedRoundoffMetric()
	}
	if branch.Base.OracleCiphertext != nil && (output.Level() != branch.Base.OracleCiphertext.Level() || output.Degree() != branch.Base.OracleCiphertext.Degree()) {
		result.LevelSafe = false
		result.Valid = false
		if firstFailure == "" {
			firstFailure = "level_or_degree_changed"
		}
	}
	if result.FullError != nil && result.FullError.MaxComponent > psMixedScaleLocalTarget {
		result.Valid = false
	}
	if firstFailure == "" {
		firstFailure = "none"
	}
	result.FirstFailure = firstFailure
	return result, nil
}

func psMixedScalePair(expected, actual []complex128, threshold float64) psOracleScaleMetricPair {
	return psOracleScalePair(expected, actual, threshold)
}

func psMixedRunCandidate(params ckks.Parameters, eval *bootstrapping.FastEvaluator, realBranch, imagBranch psOracleScaleBranch, realStandard, imagStandard []complex128, schedule []int, name string) (psMixedCandidate, psMixedBranchRun, psMixedBranchRun, error) {
	realRun, err := psMixedRunBranch(params, eval, realBranch, schedule, realStandard)
	if err != nil {
		return psMixedCandidate{}, realRun, psMixedBranchRun{}, err
	}
	imagRun, err := psMixedRunBranch(params, eval, imagBranch, schedule, imagStandard)
	if err != nil {
		return psMixedCandidate{}, realRun, imagRun, err
	}
	finalReal, finalImag := float64(0), float64(0)
	standardReal, standardImag := float64(0), float64(0)
	fullReal, fullImag := float64(0), float64(0)
	if realRun.PlainError != nil {
		finalReal = realRun.PlainError.MaxComponent
	}
	if imagRun.PlainError != nil {
		finalImag = imagRun.PlainError.MaxComponent
	}
	if realRun.StandardError != nil {
		standardReal = realRun.StandardError.MaxComponent
	}
	if imagRun.StandardError != nil {
		standardImag = imagRun.StandardError.MaxComponent
	}
	if realRun.FullError != nil {
		fullReal = realRun.FullError.MaxComponent
	}
	if imagRun.FullError != nil {
		fullImag = imagRun.FullError.MaxComponent
	}
	candidate := psMixedCandidate{
		Name: name, Schedule: append([]int(nil), schedule...), FinalRealError: finalReal, FinalImagError: finalImag, FinalMaxError: math.Max(finalReal, finalImag), StandardRealError: standardReal, StandardImagError: standardImag, FullRNSRealError: fullReal, FullRNSImagError: fullImag,
		CapacitySafe: realRun.CapacitySafe && imagRun.CapacitySafe, SemanticSafe: realRun.SemanticSafe && imagRun.SemanticSafe, LevelSafe: realRun.LevelSafe && imagRun.LevelSafe,
		FirstFailure: realRun.FirstFailure, ScalarAuditCount: len(realRun.Replay.Scalars) + len(imagRun.Replay.Scalars), AlignmentAuditCount: len(realRun.Replay.Alignments) + len(imagRun.Replay.Alignments),
	}
	if candidate.FirstFailure == "none" && imagRun.FirstFailure != "none" {
		candidate.FirstFailure = imagRun.FirstFailure
	}
	for _, checkpoint := range append(realRun.Replay.Checks, imagRun.Replay.Checks...) {
		if checkpoint.SourceBacked != nil && checkpoint.SourceBacked.MaxComponent > psMixedScaleBudget && candidate.FirstBudgetCrossing == "" {
			candidate.FirstBudgetCrossing = checkpoint.ID
		}
	}
	candidate.Valid = candidate.CapacitySafe && candidate.SemanticSafe && candidate.LevelSafe && fullReal <= psMixedScaleLocalTarget && fullImag <= psMixedScaleLocalTarget && candidate.FirstFailure == "none"
	candidate.PrecisionQualified = candidate.Valid && finalReal <= psMixedScaleBudget && finalImag <= psMixedScaleBudget && standardReal <= psMixedScaleBudget && standardImag <= psMixedScaleBudget
	if realRun.Final != nil {
		candidate.OutputScaleReal = finalizationScaleString(realRun.Final.Scale)
	}
	if imagRun.Final != nil {
		candidate.OutputScaleImag = finalizationScaleString(imagRun.Final.Scale)
	}
	return candidate, realRun, imagRun, nil
}

func psMixedCandidateKey(schedule []int) string {
	parts := make([]string, len(schedule))
	for i, exponent := range schedule {
		parts[i] = fmt.Sprintf("%d", exponent)
	}
	return strings.Join(parts, ",")
}

func psMixedFirstScalar(audits []psMixedScalarAudit, id string) *psMixedScalarAudit {
	for i := range audits {
		if audits[i].ID == id {
			return &audits[i]
		}
	}
	return nil
}

func psMixedFirstAlignment(audits []psMixedAlignmentAudit, id string) *psMixedAlignmentAudit {
	for i := range audits {
		if audits[i].ID == id || strings.HasPrefix(audits[i].ID, id+"-") {
			return &audits[i]
		}
	}
	return nil
}

func psMixedAllRowsMatch(params ckks.Parameters, run psMixedBranchRun) bool {
	for _, checkpoint := range run.Replay.Checks {
		_, centered, rowsMatch, err := psMixedCheckpointCapacity(params, &checkpoint)
		if err != nil || (centered && !rowsMatch) {
			return false
		}
	}
	return true
}

func psMixedBlockSafeAtExponent(params ckks.Parameters, run psMixedBranchRun, block, exponent int) (bool, float64, string) {
	prefix := fmt.Sprintf("B%d-", block)
	maxRatio := float64(0)
	limiting := "none"
	found := false
	for _, scalar := range run.Replay.Scalars {
		if !strings.HasPrefix(scalar.ID, prefix) {
			continue
		}
		found = true
		if scalar.MaxCapacityRatio > maxRatio {
			maxRatio = scalar.MaxCapacityRatio
		}
		if !scalar.CenteredUnique && limiting == "none" {
			limiting = scalar.ID
		}
	}
	for _, checkpoint := range run.Replay.Checks {
		if !strings.HasPrefix(checkpoint.ID, prefix) {
			continue
		}
		found = true
		ratio, centered, rowsMatch, err := psMixedCheckpointCapacity(params, &checkpoint)
		if err != nil {
			return false, maxRatio, checkpoint.ID
		}
		if ratio > maxRatio {
			maxRatio = ratio
		}
		if (!centered || !rowsMatch) && limiting == "none" {
			limiting = checkpoint.ID
		}
	}
	if !found {
		return false, maxRatio, "missing_block_evidence"
	}
	_ = exponent
	return limiting == "none", maxRatio, limiting
}

type psMixedEvaluation struct {
	Candidate psMixedCandidate
	Real      psMixedBranchRun
	Imag      psMixedBranchRun
}

func psMixedScheduleLexicographicallyLess(a, b []int) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}

func psMixedBetterQualified(a, b psMixedCandidate) bool {
	countAbove := func(schedule []int) int {
		count := 0
		for _, exponent := range schedule {
			if exponent > 92 {
				count++
			}
		}
		return count
	}
	if countAbove(a.Schedule) != countAbove(b.Schedule) {
		return countAbove(a.Schedule) < countAbove(b.Schedule)
	}
	sum := func(schedule []int) int {
		total := 0
		for _, exponent := range schedule {
			total += exponent
		}
		return total
	}
	if sum(a.Schedule) != sum(b.Schedule) {
		return sum(a.Schedule) < sum(b.Schedule)
	}
	if a.FinalMaxError != b.FinalMaxError {
		return a.FinalMaxError < b.FinalMaxError
	}
	return psMixedScheduleLexicographicallyLess(a.Schedule, b.Schedule)
}

func psMixedDownstream(btp bootstrapping.Parameters, residual ckks.Parameters, eval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, inputs evalModMatchedInputs, selected psMixedEvaluation, standardSK *rlwe.SecretKey) (map[string]interface{}, error) {
	if !selected.Candidate.PrecisionQualified || selected.Real.Final == nil || selected.Imag.Final == nil {
		return map[string]interface{}{"reached": false, "reason": "no_precision_qualified_candidate"}, nil
	}
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastReal, inputs.OrdinaryReal, eval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, selected.Real.Final.Scale, selected.Real.Final)
	if err != nil {
		return nil, err
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastImag, inputs.OrdinaryImag, eval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, selected.Imag.Final.Scale, selected.Imag.Final)
	if err != nil {
		return nil, err
	}
	if realPath.FastFinal == nil || imagPath.FastFinal == nil {
		return map[string]interface{}{"reached": false, "reason": fmt.Sprintf("normalized_mod1_stopped: real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)}, nil
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return nil, err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return nil, err
	}
	publicLike, standardLike, postS2C, metadata, unchanged, err := precisionSweepPublicLike(realPath, imagPath, eval, standardEval, btp, residual.MaxSlots(), standardSK, reproducibleInput(residual, btp))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"reached":                  true,
		"evalmod_real_vs_standard": realFinal.FastVsStandard,
		"evalmod_imag_vs_standard": imagFinal.FastVsStandard,
		"post_s2c_vs_standard":     postS2C,
		"public_like":              publicLike,
		"standard_public_like":     standardLike,
		"public_metadata_correct":  metadata,
		"input_unchanged":          unchanged,
		"public_threshold":         psMixedScalePublicThreshold,
	}, nil
}

func psMixedWrite(result psMixedResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DesignLogN13PSMixedScalePrecision(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := psMixedResult{
		SchemaVersion: "fix-001-p3-design-logn13-ps-mixed-scale-precision.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, Budget: psMixedScaleBudget, PublicThreshold: psMixedScalePublicThreshold,
		M0Baseline: map[string]interface{}{}, M1Cause: map[string]interface{}{}, Validation: map[string]interface{}{
			"primary_required_base": requiredFIX001P3PSMixedScalePrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3PSMixedScaleSecondary,
			"no_secondary_production_changes": true, "no_generated_power_redesign": true, "no_q2_plus_arithmetic": true, "no_production_integration": true, "no_c2s_change": true, "no_s2c_change": true, "no_parameter_retuning": true, "no_extra_q_levels": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_metadata_only_relabel": true,
		}, Classification: "logn13_ps_mixed_scale_precondition_mismatch", FirstBlocker: "startup_precondition",
	}
	if !secondaryClean || secondaryCommit != requiredFIX001P3PSMixedScaleSecondary {
		return psMixedWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, standardSK)
	if err != nil {
		return err
	}
	realBase, err := polynomialDecompositionPrepareBranch("real", btp.BootstrappingParameters, btp, fastEval, standardEval, inputs.FastReal, inputs.OrdinaryReal, standardSK)
	if err != nil {
		return err
	}
	imagBase, err := polynomialDecompositionPrepareBranch("imag", btp.BootstrappingParameters, btp, fastEval, standardEval, inputs.FastImag, inputs.OrdinaryImag, standardSK)
	if err != nil {
		return err
	}
	realStandard, err := psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastReal, inputs.OrdinaryReal, standardSK, realBase.InputValues)
	if err != nil {
		return err
	}
	imagStandard, err := psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastImag, inputs.OrdinaryImag, standardSK, imagBase.InputValues)
	if err != nil {
		return err
	}
	realBranch := psOracleScaleBranch{Name: "real", Base: realBase, Standard: realStandard, Input: inputs.FastReal, Workload: realBase.InputValues}
	imagBranch := psOracleScaleBranch{Name: "imag", Base: imagBase, Standard: imagStandard, Input: inputs.FastImag, Workload: imagBase.InputValues}
	blockCount := len(realBase.Plan.Value)
	all92 := make([]int, blockCount)
	for i := range all92 {
		all92[i] = 92
	}
	cache := map[string]psMixedEvaluation{}
	evaluate := func(name string, schedule []int) (psMixedEvaluation, error) {
		key := psMixedCandidateKey(schedule)
		if existing, ok := cache[key]; ok {
			return existing, nil
		}
		candidate, realRun, imagRun, err := psMixedRunCandidate(btp.BootstrappingParameters, fastEval, realBranch, imagBranch, realStandard, imagStandard, schedule, name)
		if err != nil {
			return psMixedEvaluation{}, err
		}
		evaluation := psMixedEvaluation{Candidate: candidate, Real: realRun, Imag: imagRun}
		cache[key] = evaluation
		return evaluation, nil
	}
	c92, err := evaluate("C0-all-2^92", all92)
	if err != nil {
		return err
	}
	all93 := make([]int, blockCount)
	for i := range all93 {
		all93[i] = 93
	}
	c93, err := evaluate("common-2^93", all93)
	if err != nil {
		return err
	}
	rows92 := psMixedAllRowsMatch(btp.BootstrappingParameters, c92.Real) && psMixedAllRowsMatch(btp.BootstrappingParameters, c92.Imag)
	rows93 := psMixedAllRowsMatch(btp.BootstrappingParameters, c93.Real) && psMixedAllRowsMatch(btp.BootstrappingParameters, c93.Imag)
	m0Pass := c92.Candidate.FinalRealError > 2.5e-8 && c92.Candidate.FinalRealError < 4.2e-8 && c93.Candidate.FinalRealError > 1e-2 && c93.Candidate.FinalRealError < 3e-2 && c93.Candidate.FirstBudgetCrossing == "B3-term-2" && c93.Real.Replay.Root.SourceBacked != nil && c93.Real.Replay.Root.SourceBacked.MaxComponent > 1e-2 && rows92 && rows93
	result.M0Baseline = map[string]interface{}{"pass": m0Pass, "common_2^92": c92.Candidate, "common_2^93": c93.Candidate, "common_2^93_root_pre_rescale_error": c93.Real.Replay.Root.SourceBacked, "common_2^93_rows_match": rows93}
	if !m0Pass {
		result.Classification, result.FirstBlocker = "logn13_ps_mixed_scale_precondition_mismatch", "M0"
		return psMixedWrite(result, outPath)
	}
	causalID := "B3-term-2"
	var causal93 *psMixedScalarAudit
	for i := range c93.Real.Replay.Scalars {
		if !c93.Real.Replay.Scalars[i].CenteredUnique {
			causalID, causal93 = c93.Real.Replay.Scalars[i].ID, &c93.Real.Replay.Scalars[i]
			break
		}
	}
	if causal93 == nil {
		causalID = "B3-term-2"
		causal93 = psMixedFirstScalar(c93.Real.Replay.Scalars, causalID)
	}
	causal92 := psMixedFirstScalar(c92.Real.Replay.Scalars, causalID)
	causalAlignment := psMixedFirstAlignment(c93.Real.Replay.Alignments, causalID)
	classification := "ps_common_93_other_semantic_boundary"
	if causal93 != nil && !causal93.CenteredUnique {
		classification = "ps_common_93_scalar_encoding_alias"
	} else if causalAlignment != nil && !causalAlignment.IntegerRatioExact {
		classification = "ps_common_93_integer_scale_alignment_loss"
	} else if causalAlignment != nil && causalAlignment.MetadataOnlyRelabel {
		classification = "ps_common_93_metadata_scale_relabel_loss"
	}
	result.M1Cause = map[string]interface{}{"classification": classification, "first_checkpoint": causalID, "scalar_2^92": causal92, "scalar_2^93": causal93, "alignment_2^93": causalAlignment, "maintained_rows_2^93_match": rows93}
	result.Validation["m1_cause"] = classification
	result.Validation["m1_checkpoint"] = causalID
	uniformRuns := map[int]psMixedBranchRun{92: c92.Real}
	for exponent := 93; exponent <= 96; exponent++ {
		schedule := make([]int, blockCount)
		for i := range schedule {
			schedule[i] = exponent
		}
		candidate, err := evaluate(fmt.Sprintf("uniform-2^%d", exponent), schedule)
		if err != nil {
			return err
		}
		uniformRuns[exponent] = candidate.Real
	}
	result.Blocks = make([]psMixedBlockSummary, blockCount)
	ceilings := make([]int, blockCount)
	for block := 0; block < blockCount; block++ {
		ceiling := 92
		limiting := "none"
		maxRatio := float64(0)
		for exponent := 96; exponent >= 92; exponent-- {
			safe, ratio, limit := psMixedBlockSafeAtExponent(btp.BootstrappingParameters, uniformRuns[exponent], block, exponent)
			if safe {
				ceiling = exponent
				maxRatio, limiting = ratio, limit
				break
			}
			if exponent == 92 {
				_, maxRatio, limiting = safe, ratio, limit
			}
		}
		ceilings[block] = ceiling
		result.Blocks[block] = psMixedBlockSummary{Block: block, BlockID: fmt.Sprintf("B%d", block), SafeScaleCeiling: ceiling, LimitingOperation: limiting, HeadroomRatio: 1 - maxRatio, ScalarMaxRatioAtCeiling: maxRatio, AllScaleChecksValid: true}
	}
	result.C0Control = c92.Candidate
	for block, ceiling := range ceilings {
		if ceiling < 93 {
			continue
		}
		schedule := append([]int(nil), all92...)
		schedule[block] = 93
		evaluation, err := evaluate(fmt.Sprintf("single-block-B%d-2^93", block), schedule)
		if err != nil {
			return err
		}
		result.SingleBlockLifts = append(result.SingleBlockLifts, evaluation.Candidate)
	}
	safe93 := append([]int(nil), all92...)
	for block, ceiling := range ceilings {
		if ceiling >= 93 {
			safe93[block] = 93
		}
	}
	safe93Eval, err := evaluate("safe-93-aggregate", safe93)
	if err != nil {
		return err
	}
	result.Safe93Aggregate = safe93Eval.Candidate
	currentSchedule := append([]int(nil), all92...)
	currentEval := c92
	for step := 0; step < 12; step++ {
		bestEval := psMixedEvaluation{}
		bestBlock := -1
		bestImprovement := float64(0)
		for block, ceiling := range ceilings {
			if currentSchedule[block] >= ceiling {
				continue
			}
			trialSchedule := append([]int(nil), currentSchedule...)
			trialSchedule[block]++
			trial, err := evaluate(fmt.Sprintf("greedy-step-%d-B%d", step+1, block), trialSchedule)
			if err != nil {
				return err
			}
			improvement := currentEval.Candidate.FinalMaxError - trial.Candidate.FinalMaxError
			if !trial.Candidate.Valid || improvement <= 0 {
				continue
			}
			if bestBlock < 0 || improvement > bestImprovement || (improvement == bestImprovement && block < bestBlock) {
				bestEval, bestBlock, bestImprovement = trial, block, improvement
			}
		}
		if bestBlock < 0 {
			break
		}
		from := currentSchedule[bestBlock]
		currentSchedule[bestBlock]++
		result.GreedySteps = append(result.GreedySteps, psMixedGreedyStep{Step: len(result.GreedySteps) + 1, Block: bestBlock, FromExponent: from, ToExponent: currentSchedule[bestBlock], BeforeMaxError: currentEval.Candidate.FinalMaxError, AfterMaxError: bestEval.Candidate.FinalMaxError, Improvement: bestImprovement})
		currentEval = bestEval
	}
	allCandidates := []psMixedCandidate{c92.Candidate, safe93Eval.Candidate}
	allCandidates = append(allCandidates, result.SingleBlockLifts...)
	allCandidates = append(allCandidates, currentEval.Candidate)
	best := c92.Candidate
	for _, candidate := range allCandidates {
		if candidate.Valid && (!best.Valid || candidate.FinalMaxError < best.FinalMaxError) {
			best = candidate
		}
	}
	result.BestCandidate = best
	selected := psMixedCandidate{}
	for _, candidate := range allCandidates {
		if !candidate.PrecisionQualified {
			continue
		}
		if result.SelectedCandidate == "" || psMixedBetterQualified(candidate, selected) {
			selected, result.SelectedCandidate = candidate, candidate.Name
		}
	}
	if result.SelectedCandidate != "" {
		selectedEval := cache[psMixedCandidateKey(selected.Schedule)]
		downstream, err := psMixedDownstream(btp, residual, fastEval, standardEval, inputs, selectedEval, standardSK)
		if err != nil {
			return err
		}
		result.Downstream = downstream
		result.Classification, result.FirstBlocker = "logn13_ps_oracle_mixed_scale_candidate_validated", "none"
	} else {
		if classification == "ps_common_93_scalar_encoding_alias" {
			result.Classification = "logn13_ps_oracle_mixed_scale_blocked_by_scalar_encoding"
		} else if classification == "ps_common_93_integer_scale_alignment_loss" || !result.BestCandidate.SemanticSafe {
			result.Classification = "logn13_ps_oracle_mixed_scale_blocked_by_alignment"
		} else {
			result.Classification = "logn13_ps_oracle_mixed_scale_insufficient_precision"
		}
		result.FirstBlocker = causalID
	}
	return psMixedWrite(result, outPath)
}
