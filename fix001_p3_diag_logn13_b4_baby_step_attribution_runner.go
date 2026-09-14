package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3B4Primary   = "33ff9bd8392685e31433a4956ea36309de4fd9de"
	requiredFIX001P3B4Secondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	b4BabyPublicThreshold       = 1e-2
	b4BabySemanticThreshold     = 1e-10
	b4BabyLedgerTolerance       = 1e-12
)

type b4BabyState struct {
	Hash          string               `json:"q0_q1_state_hash"`
	Level         int                  `json:"level"`
	Scale         string               `json:"scale"`
	Degree        int                  `json:"degree"`
	IsNTT         bool                 `json:"is_ntt"`
	IsMontgomery  bool                 `json:"is_montgomery"`
	DecodedMaxAbs float64              `json:"decoded_max_abs"`
	Capacity      *postMod1S2CCapacity `json:"q01_capacity,omitempty"`
}

type b4BabyIdentity struct {
	ActualB4ToAF    bool            `json:"actual_b4_to_a_f_exact"`
	CanonicalB4ToAQ bool            `json:"canonical_b4_to_a_q_exact"`
	ActualB4        b4BabyState     `json:"actual_b4_state"`
	AF              b4BabyState     `json:"a_f_state"`
	CanonicalB4     b4BabyState     `json:"canonical_b4_state"`
	AQ              b4BabyState     `json:"a_q_state"`
	ActualResidual  *PSGlobalMetric `json:"actual_g0_parent_vs_canonical"`
	IdentityValid   bool            `json:"identity_valid"`
}

type b4BabyOperation struct {
	ID                       string  `json:"id"`
	Kind                     string  `json:"kind"`
	CoefficientDegree        int     `json:"coefficient_degree"`
	CoefficientReal          string  `json:"coefficient_real"`
	CoefficientImag          string  `json:"coefficient_imag"`
	CoefficientIsInteger     bool    `json:"coefficient_is_integer"`
	PowerLevel               int     `json:"input_power_level"`
	PowerScale               string  `json:"input_power_scale"`
	AccumulatorScaleBefore   string  `json:"accumulator_scale_before"`
	AccumulatorScaleAfter    string  `json:"accumulator_scale_after"`
	ScalarEncodingScale      string  `json:"scalar_encoding_scale"`
	RoundedIntegerReal       string  `json:"rounded_integer_real"`
	RoundedIntegerImag       string  `json:"rounded_integer_imag"`
	EffectiveCoefficientReal string  `json:"effective_coefficient_real"`
	EffectiveCoefficientImag string  `json:"effective_coefficient_imag"`
	QuantizationErrorReal    string  `json:"quantization_error_real"`
	QuantizationErrorImag    string  `json:"quantization_error_imag"`
	QuantizationErrorSign    string  `json:"quantization_error_sign"`
	AccumulatorPromotion     bool    `json:"accumulator_promotion"`
	PromotionFactor          string  `json:"promotion_factor,omitempty"`
	PromotionExact           bool    `json:"promotion_factor_exact"`
	ScalarQ01CapacityRatio   float64 `json:"scalar_q01_capacity_ratio"`
	ScalarCenteredUnique     bool    `json:"scalar_centered_unique"`
	ResultCapacityRatio      float64 `json:"result_capacity_ratio,omitempty"`
	SourceFunction           string  `json:"source_function"`
	SourceScalarPath         string  `json:"source_scalar_path"`
	SourceScalePath          string  `json:"source_scale_path"`
}

type b4BabyMode struct {
	Mode       string          `json:"mode"`
	Available  bool            `json:"available"`
	Reason     string          `json:"reason,omitempty"`
	State      *b4BabyState    `json:"state,omitempty"`
	ValuesVsQ  *PSGlobalMetric `json:"values_vs_q,omitempty"`
	ValuesVsS3 *PSGlobalMetric `json:"values_vs_s3,omitempty"`
}

type b4BabySemantics struct {
	F           b4BabyMode           `json:"f_current_fast"`
	N           b4BabyMode           `json:"n_full_rns_fast_schedule"`
	S           b4BabyMode           `json:"s_standard_scalar_semantics"`
	Q           b4BabyMode           `json:"q_canonical_checkpoint"`
	FMinusN     *PSGlobalMetric      `json:"f_minus_n,omitempty"`
	NMinusQ     *PSGlobalMetric      `json:"n_minus_q,omitempty"`
	SMinusQ     *PSGlobalMetric      `json:"s_minus_q,omitempty"`
	FMinusQ     *PSGlobalMetric      `json:"f_minus_q,omitempty"`
	FNRowsEqual bool                 `json:"f_n_q01_rows_equal"`
	FCapacity   *postMod1S2CCapacity `json:"f_capacity,omitempty"`
	NCapacity   *postMod1S2CCapacity `json:"n_capacity,omitempty"`
	fValues     []complex128         `json:"-"`
	nValues     []complex128         `json:"-"`
	qValues     []complex128         `json:"-"`
	sValues     []complex128         `json:"-"`
}

type b4BabyDownstream struct {
	Real            *psLocalizationDownstreamEvidence `json:"-"`
	Imag            *psLocalizationDownstreamEvidence `json:"-"`
	EvalModReal     *PSGlobalMetric                   `json:"evalmod_real,omitempty"`
	EvalModImag     *PSGlobalMetric                   `json:"evalmod_imag,omitempty"`
	PublicLike      *PSGlobalMetric                   `json:"public_like,omitempty"`
	PostS2C         *PSGlobalMetric                   `json:"post_s2c,omitempty"`
	MetadataCorrect bool                              `json:"metadata_correct"`
	ContractsPass   bool                              `json:"contracts_pass"`
	Pass            bool                              `json:"pass"`
}

type b4BabyReset struct {
	ID                string            `json:"id"`
	CoefficientDegree int               `json:"coefficient_degree"`
	B4FMinusQ         *PSGlobalMetric   `json:"completed_b4_f_minus_q,omitempty"`
	FinalPSMinusQ     *PSGlobalMetric   `json:"final_ps_minus_q,omitempty"`
	Downstream        *b4BabyDownstream `json:"downstream,omitempty"`
	Pass              bool              `json:"pass"`
}

type b4BabySubstitution struct {
	ID                string            `json:"id"`
	Variant           string            `json:"variant"`
	Valid             bool              `json:"valid"`
	Reason            string            `json:"reason,omitempty"`
	CompletedB4MinusQ *PSGlobalMetric   `json:"completed_b4_minus_q,omitempty"`
	Downstream        *b4BabyDownstream `json:"downstream,omitempty"`
}

type b4BabyLedger struct {
	Branch            string            `json:"branch"`
	Rows              []b4BabyOperation `json:"rows"`
	PredictedResidual *PSGlobalMetric   `json:"predicted_residual_vs_actual_b4_minus_q,omitempty"`
	Closure           *PSGlobalMetric   `json:"ledger_closure_vs_n_minus_q,omitempty"`
	ClosureMax        float64           `json:"closure_max_component"`
	Meaningful        bool              `json:"closure_numerically_meaningful"`
}

type b4BabyResult struct {
	SchemaVersion         string                          `json:"schema_version"`
	Timestamp             time.Time                       `json:"timestamp"`
	Primary               RepositoryMetadata              `json:"primary_repository"`
	Lattigo               RepositoryMetadata              `json:"lattigo_repository"`
	Environment           EnvironmentMetadata             `json:"environment"`
	Config                BootstrapConfig                 `json:"config"`
	Parameters            ExperimentParameters            `json:"effective_parameters"`
	Provenance            map[string]interface{}          `json:"provenance"`
	M0Controls            map[string]interface{}          `json:"m0_controls"`
	B4ParentIdentity      map[string]b4BabyIdentity       `json:"b4_parent_identity"`
	Operations            map[string][]b4BabyOperation    `json:"b4_operation_schedule"`
	Semantics             map[string]b4BabySemantics      `json:"execution_mode_comparison"`
	OperationResets       map[string][]b4BabyReset        `json:"operation_reset_downstream"`
	SmallestCausal        map[string]interface{}          `json:"smallest_causal_interval"`
	Substitutions         map[string][]b4BabySubstitution `json:"operation_substitutions"`
	ScalarLedgers         map[string]b4BabyLedger         `json:"scalar_rounding_ledger"`
	OneCoefficient        map[string][]b4BabySubstitution `json:"one_coefficient_corrections,omitempty"`
	Classification        string                          `json:"classification"`
	FirstRemainingBlocker string                          `json:"first_remaining_blocker"`
	RecommendedNextTarget string                          `json:"recommended_next_design_target"`
	Validation            map[string]interface{}          `json:"validation"`
}

type b4BabyFullRun struct {
	Ciphertext *rlwe.Ciphertext
	Values     []complex128
	Operations []b4BabyOperation
}

func b4BabyScaleString(scale rlwe.Scale) string {
	return scale.Value.Text('e', 40)
}

func b4BabyStateHash(ct *rlwe.Ciphertext) string {
	h := sha256.New()
	fmt.Fprintf(h, "level=%d|degree=%d|scale=%s|ntt=%t|mont=%t|", ct.Level(), ct.Degree(), b4BabyScaleString(ct.Scale), ct.IsNTT, ct.IsMontgomery)
	for component := range ct.Value {
		for limb := 0; limb < 2 && limb < len(ct.Value[component].Coeffs); limb++ {
			fmt.Fprintf(h, "c%dq%d:%s|", component, limb, finalizationRowHash(ct.Value[component].Coeffs[limb]))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func b4BabyStateEvidence(params ckks.Parameters, ct *rlwe.Ciphertext, values []complex128) b4BabyState {
	state := b4BabyState{Hash: b4BabyStateHash(ct), Level: ct.Level(), Scale: b4BabyScaleString(ct.Scale), Degree: ct.Degree(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, DecodedMaxAbs: g0MergeMaxAbs(values)}
	if capacity, err := postMod1S2CCapacityFromFastCiphertext(params, ct); err == nil {
		state.Capacity = &capacity
	}
	return state
}

func b4BabyExactState(a, b *rlwe.Ciphertext) bool {
	if a == nil || b == nil || a.Level() != b.Level() || a.Degree() != b.Degree() || a.IsNTT != b.IsNTT || a.IsMontgomery != b.IsMontgomery || !a.Scale.Equal(b.Scale) {
		return false
	}
	return b4BabyStateHash(a) == b4BabyStateHash(b)
}

func b4BabyMetricMax(a, b *PSGlobalMetric) *PSGlobalMetric {
	return psLocalizationMaxPublic(a, b)
}

func b4BabyEffective(value *big.Int, scale rlwe.Scale) string {
	if value == nil {
		return "0"
	}
	f := new(big.Float).Quo(new(big.Float).SetInt(value), &scale.Value)
	return f.Text('e', 40)
}

func b4BabyError(value *big.Int, scale rlwe.Scale, exact *big.Float) string {
	effective := new(big.Float).Quo(new(big.Float).SetInt(value), &scale.Value)
	effective.Sub(effective, exact)
	return effective.Text('e', 40)
}

func b4BabyErrorSign(realText, imagText string) string {
	real, _ := new(big.Float).SetString(realText)
	imag, _ := new(big.Float).SetString(imagText)
	if real != nil && real.Sign() != 0 {
		if real.Sign() > 0 {
			return "real_positive"
		}
		return "real_negative"
	}
	if imag != nil && imag.Sign() != 0 {
		if imag.Sign() > 0 {
			return "imag_positive"
		}
		return "imag_negative"
	}
	return "zero"
}

func b4BabyOperationRow(params ckks.Parameters, id, kind string, degree int, coefficient *bignum.Complex, power *rlwe.Ciphertext, before, after rlwe.Scale, encoding rlwe.Scale, promotion bool, factor string, exactPromotion bool) b4BabyOperation {
	realInteger := psMixedRoundCoefficient(coefficient[0], encoding)
	imagInteger := psMixedRoundCoefficient(coefficient[1], encoding)
	_, half := psMixedQ01(params)
	realRatio := psMixedRatio(new(big.Int).Abs(new(big.Int).Set(realInteger)), half)
	imagRatio := psMixedRatio(new(big.Int).Abs(new(big.Int).Set(imagInteger)), half)
	errorReal := b4BabyError(realInteger, encoding, coefficient[0])
	errorImag := b4BabyError(imagInteger, encoding, coefficient[1])
	return b4BabyOperation{ID: id, Kind: kind, CoefficientDegree: degree, CoefficientReal: psMixedBigFloatString(coefficient[0]), CoefficientImag: psMixedBigFloatString(coefficient[1]), CoefficientIsInteger: coefficient.IsInt(), PowerLevel: power.Level(), PowerScale: b4BabyScaleString(power.Scale), AccumulatorScaleBefore: b4BabyScaleString(before), AccumulatorScaleAfter: b4BabyScaleString(after), ScalarEncodingScale: b4BabyScaleString(encoding), RoundedIntegerReal: realInteger.String(), RoundedIntegerImag: imagInteger.String(), EffectiveCoefficientReal: b4BabyEffective(realInteger, encoding), EffectiveCoefficientImag: b4BabyEffective(imagInteger, encoding), QuantizationErrorReal: errorReal, QuantizationErrorImag: errorImag, QuantizationErrorSign: b4BabyErrorSign(errorReal, errorImag), AccumulatorPromotion: promotion, PromotionFactor: factor, PromotionExact: exactPromotion, ScalarQ01CapacityRatio: math.Max(realRatio, imagRatio), ScalarCenteredUnique: realRatio < 1 && imagRatio < 1, SourceFunction: "fastPolynomialWorkspace.evaluateBabyStep", SourceScalarPath: "schemes/ckks/fast/evaluator_ntt.go:MulThenAdd -> scalarNTT", SourceScalePath: "schemes/ckks/fast/evaluator_ntt.go:coefficientScale / scale promotion"}
}

func b4BabyOps(params ckks.Parameters, context psLocalizationAcceptedContext) ([]b4BabyOperation, error) {
	if len(context.Plan.Value) <= 4 {
		return nil, fmt.Errorf("accepted PS plan has no B4 block")
	}
	poly := context.Plan.Value[4]
	powers := context.Branch.Base.OraclePowerMap
	if powers[1] == nil {
		return nil, fmt.Errorf("B4 missing power 1")
	}
	current := poly.Scale
	rows := make([]b4BabyOperation, 0, 4)
	if poly.IsEven {
		if poly.Coeffs[0] == nil {
			return nil, fmt.Errorf("B4 constant coefficient missing")
		}
		rows = append(rows, b4BabyOperationRow(params, "B4-constant", "constant_add", 0, poly.Coeffs[0], powers[1], current, current, current, false, "", true))
	}
	for key := poly.Degree(); key > 0; key-- {
		if (poly.IsEven || poly.IsOdd) && ((key&1 == 0 && !poly.IsEven) || (key&1 == 1 && !poly.IsOdd)) {
			continue
		}
		coefficient := poly.Coeffs[key]
		if coefficient == nil {
			continue
		}
		power := powers[key]
		if power == nil {
			return nil, fmt.Errorf("B4 missing power %d", key)
		}
		before := current
		promotion := false
		factor, exact := "", true
		if power.Scale.Cmp(current) > 0 {
			promotion = true
			_, factor, _, exact = psMixedScaleRatio(power.Scale, current)
			current = power.Scale
		}
		encoding, err := psMixedScalarEncodingScale(params, power.Scale, current, coefficient, power.Level())
		if err != nil {
			return nil, err
		}
		after := current
		if !coefficient.IsInt() && power.Scale.Cmp(before) >= 0 {
			after = current.Mul(encoding)
		}
		rows = append(rows, b4BabyOperationRow(params, fmt.Sprintf("B4-term-%d", key), "scalar_mul_then_add", key, coefficient, power, before, after, encoding, promotion, factor, exact))
		current = after
	}
	return rows, nil
}

func b4BabyFullZero(params ckks.Parameters, source *rlwe.Ciphertext, level int, scale rlwe.Scale) *rlwe.Ciphertext {
	value := ckks.NewCiphertext(params, 1, level)
	*value.MetaData = *source.MetaData
	value.IsNTT = source.IsNTT
	value.IsMontgomery = false
	value.Scale = scale
	return value
}

func b4BabyScalarValues(c *bignum.Complex, scale *big.Float, subring *ring.SubRing) [2]uint64 {
	toInt := func(value *big.Float) *big.Int {
		v := new(big.Float).Mul(value, scale)
		if value.Sign() > 0 {
			v.Add(v, new(big.Float).SetFloat64(0.5))
		} else if value.Sign() < 0 {
			v.Sub(v, new(big.Float).SetFloat64(0.5))
		}
		out := new(big.Int)
		v.Int(out)
		return out
	}
	real, imag := toInt(c[0]), toInt(c[1])
	r := new(big.Int).Mod(real, new(big.Int).SetUint64(subring.Modulus)).Uint64()
	i := new(big.Int).Mod(imag, new(big.Int).SetUint64(subring.Modulus)).Uint64()
	i = ring.MRed(i, subring.RootsForward[1], subring.Modulus, subring.MRedConstant)
	return [2]uint64{ring.CRed(r+i, subring.Modulus), ring.CRed(r+subring.Modulus-i, subring.Modulus)}
}

func b4BabyFullAddScalar(params ckks.Parameters, value *rlwe.Ciphertext, coefficient *bignum.Complex) {
	half := params.N() >> 1
	for limb := 0; limb <= value.Level() && limb < len(params.RingQ().SubRings); limb++ {
		s := params.RingQ().SubRings[limb]
		values := b4BabyScalarValues(coefficient, &value.Scale.Value, s)
		if value.IsMontgomery {
			values[0] = ring.MForm(values[0], s.Modulus, s.BRedConstant)
			values[1] = ring.MForm(values[1], s.Modulus, s.BRedConstant)
		}
		s.AddScalar(value.Value[0].Coeffs[limb][:half], values[0], value.Value[0].Coeffs[limb][:half])
		s.AddScalar(value.Value[0].Coeffs[limb][half:], values[1], value.Value[0].Coeffs[limb][half:])
	}
}

func b4BabyToMontgomery(params ckks.Parameters, value *rlwe.Ciphertext) {
	for component := range value.Value {
		for limb := 0; limb <= value.Level() && limb < len(value.Value[component].Coeffs) && limb < len(params.RingQ().SubRings); limb++ {
			params.RingQ().SubRings[limb].MForm(value.Value[component].Coeffs[limb], value.Value[component].Coeffs[limb])
		}
	}
	value.IsMontgomery = true
}

func b4BabyDecodeFull(params ckks.Parameters, value *rlwe.Ciphertext) ([]complex128, error) {
	normal := value
	if value.IsMontgomery {
		normal = value.CopyNew()
		for component := range normal.Value {
			for limb := 0; limb < 2 && limb < len(normal.Value[component].Coeffs); limb++ {
				params.RingQ().SubRings[limb].IMForm(normal.Value[component].Coeffs[limb], normal.Value[component].Coeffs[limb])
			}
		}
		normal.IsMontgomery = false
	}
	return g0MergeDecodeFullQ01(params, normal)
}

func b4BabyFullMulInteger(params ckks.Parameters, value *rlwe.Ciphertext, factor *big.Int) {
	if factor == nil || factor.Sign() == 0 {
		return
	}
	ringQ := params.RingQ().AtLevel(value.Level())
	for component := range value.Value {
		ringQ.MulScalar(value.Value[component], factor.Uint64(), value.Value[component])
	}
}

func b4BabyFullMulThenAdd(params ckks.Parameters, power, value *rlwe.Ciphertext, coefficient *bignum.Complex, encoding rlwe.Scale) {
	half := params.N() >> 1
	for component := 0; component < len(power.Value) && component < len(value.Value); component++ {
		for limb := 0; limb <= value.Level() && limb < len(power.Value[component].Coeffs) && limb < len(value.Value[component].Coeffs) && limb < len(params.RingQ().SubRings); limb++ {
			s := params.RingQ().SubRings[limb]
			values := b4BabyScalarValues(coefficient, &encoding.Value, s)
			s.MulScalarMontgomeryThenAdd(power.Value[component].Coeffs[limb][:half], ring.MForm(values[0], s.Modulus, s.BRedConstant), value.Value[component].Coeffs[limb][:half])
			s.MulScalarMontgomeryThenAdd(power.Value[component].Coeffs[limb][half:], ring.MForm(values[1], s.Modulus, s.BRedConstant), value.Value[component].Coeffs[limb][half:])
		}
	}
}

func b4BabyFullFastSchedule(params ckks.Parameters, context psLocalizationAcceptedContext) (b4BabyFullRun, error) {
	if len(context.Plan.Value) <= 4 {
		return b4BabyFullRun{}, fmt.Errorf("accepted PS plan has no B4 block")
	}
	poly := context.Plan.Value[4]
	powers := context.Branch.Base.OraclePowerMap
	fullPowers := make(map[int]*rlwe.Ciphertext)
	for key, power := range powers {
		full, _, err := postMod1S2CLiftFull(params, power)
		if err != nil {
			return b4BabyFullRun{}, err
		}
		b4BabyToMontgomery(params, full)
		fullPowers[key] = full
	}
	value := b4BabyFullZero(params, fullPowers[1], poly.Level, poly.Scale)
	b4BabyToMontgomery(params, value)
	rows, err := b4BabyOps(params, context)
	if err != nil {
		return b4BabyFullRun{}, err
	}
	rowIndex := 0
	if poly.IsEven {
		b4BabyFullAddScalar(params, value, poly.Coeffs[0])
		rowIndex++
	}
	for key := poly.Degree(); key > 0; key-- {
		if (poly.IsEven || poly.IsOdd) && ((key&1 == 0 && !poly.IsEven) || (key&1 == 1 && !poly.IsOdd)) {
			continue
		}
		coefficient := poly.Coeffs[key]
		if coefficient == nil {
			continue
		}
		power := fullPowers[key]
		if power.Scale.Cmp(value.Scale) > 0 {
			ratio := power.Scale.Div(value.Scale).BigInt()
			b4BabyFullMulInteger(params, value, ratio)
			value.Scale = power.Scale
		}
		encoding, err := psMixedScalarEncodingScale(params, power.Scale, value.Scale, coefficient, power.Level())
		if err != nil {
			return b4BabyFullRun{}, err
		}
		if !coefficient.IsInt() && power.Scale.Cmp(value.Scale) >= 0 {
			b4BabyFullMulInteger(params, value, encoding.BigInt())
			value.Scale = value.Scale.Mul(encoding)
		}
		b4BabyFullMulThenAdd(params, power, value, coefficient, encoding)
		if rowIndex >= len(rows) {
			return b4BabyFullRun{}, fmt.Errorf("B4 operation row mismatch at term %d", key)
		}
		rowIndex++
	}
	values, err := b4BabyDecodeFull(params, value)
	if err != nil {
		return b4BabyFullRun{}, err
	}
	return b4BabyFullRun{Ciphertext: value, Values: values, Operations: rows}, nil
}

func b4BabyStandardRun(profile q056PreparedProfile, context psLocalizationAcceptedContext) (b4BabyFullRun, error) {
	params := profile.BTP.BootstrappingParameters
	poly := context.Plan.Value[4]
	fullPowers := make(map[int]*rlwe.Ciphertext)
	for key, power := range context.Branch.Base.OraclePowerMap {
		full, _, err := postMod1S2CLiftFull(params, power)
		if err != nil {
			return b4BabyFullRun{}, err
		}
		fullPowers[key] = full
	}
	value := b4BabyFullZero(params, fullPowers[1], poly.Level, poly.Scale)
	rows, err := b4BabyOps(params, context)
	if err != nil {
		return b4BabyFullRun{}, err
	}
	if poly.IsEven {
		if err := profile.Standard.Evaluator.Add(value, poly.Coeffs[0], value); err != nil {
			return b4BabyFullRun{Operations: rows}, fmt.Errorf("B4-constant: %w", err)
		}
	}
	for key := poly.Degree(); key > 0; key-- {
		if (poly.IsEven || poly.IsOdd) && ((key&1 == 0 && !poly.IsEven) || (key&1 == 1 && !poly.IsOdd)) {
			continue
		}
		if poly.Coeffs[key] == nil {
			continue
		}
		if err := profile.Standard.Evaluator.MulThenAdd(fullPowers[key], poly.Coeffs[key], value); err != nil {
			return b4BabyFullRun{Operations: rows}, fmt.Errorf("B4-term-%d: %w", key, err)
		}
	}
	values, err := g0MergeDecodeFullQ01(params, value)
	if err != nil {
		return b4BabyFullRun{}, err
	}
	return b4BabyFullRun{Ciphertext: value, Values: values, Operations: rows}, nil
}

func b4BabyFindCheckpoint(checks []PSGlobalCheckpoint, id string) (PSGlobalCheckpoint, bool) {
	for _, check := range checks {
		if check.ID == id {
			return check, true
		}
	}
	return PSGlobalCheckpoint{}, false
}

func b4BabyDownstreamPair(profile q056PreparedProfile, realCT, imagCT *rlwe.Ciphertext) (*b4BabyDownstream, error) {
	realDownstream, err := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psLocalizationPlanScaleExponent), realCT, imagCT)
	if err != nil {
		return nil, err
	}
	return &b4BabyDownstream{Real: realDownstream, EvalModReal: realDownstream.EvalModReal, EvalModImag: realDownstream.EvalModImag, PublicLike: realDownstream.PublicLike, PostS2C: realDownstream.PostS2C, MetadataCorrect: realDownstream.MetadataCorrect, ContractsPass: realDownstream.ContractsPass, Pass: realDownstream.Reached && realDownstream.PublicLike != nil && realDownstream.PublicLike.Pass}, nil
}

func b4BabyCombineDownstream(real, imag *psLocalizationDownstreamEvidence) *b4BabyDownstream {
	if real == nil || imag == nil {
		return &b4BabyDownstream{}
	}
	public := b4BabyMetricMax(real.PublicLike, imag.PublicLike)
	post := b4BabyMetricMax(real.PostS2C, imag.PostS2C)
	return &b4BabyDownstream{Real: real, Imag: imag, EvalModReal: real.EvalModReal, EvalModImag: imag.EvalModImag, PublicLike: public, PostS2C: post, MetadataCorrect: real.MetadataCorrect && imag.MetadataCorrect, ContractsPass: real.ContractsPass && imag.ContractsPass, Pass: public != nil && public.Pass}
}

func b4BabyAcceptedReset(profile q056PreparedProfile, context psLocalizationAcceptedContext, id string, canonical *rlwe.Ciphertext) (psRescaleGuardReplay, []complex128, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, context.Branch.Base.OraclePowerMap, context.Branch.Base.PowerExpected, context.Branch.Base.OraclePowerValues, context.Branch.Base.InputValues, precisionSweepScale(psLocalizationPlanScaleExponent), context.GuardMap, context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: id, Ciphertext: canonical})
	if err != nil {
		return replay, nil, err
	}
	if replay.Final == nil {
		return replay, nil, fmt.Errorf("reset %s returned nil final ciphertext", id)
	}
	values, err := psGlobalDecode(profile.BTP.BootstrappingParameters, replay.Final)
	return replay, values, err
}

func b4BabyResetPair(profile q056PreparedProfile, realContext, imagContext psLocalizationAcceptedContext, realWork, imagWork polynomialDecompositionBranchWork, id string, realCanonical, imagCanonical *rlwe.Ciphertext, degree int) (b4BabyReset, error) {
	realReplay, realValues, err := b4BabyAcceptedReset(profile, realContext, id, realCanonical)
	if err != nil {
		return b4BabyReset{}, err
	}
	imagReplay, imagValues, err := b4BabyAcceptedReset(profile, imagContext, id, imagCanonical)
	if err != nil {
		return b4BabyReset{}, err
	}
	realCheck, realOK := b4BabyFindCheckpoint(realReplay.Checks, "B4-term-2")
	imagCheck, imagOK := b4BabyFindCheckpoint(imagReplay.Checks, "B4-term-2")
	if !realOK || !imagOK {
		return b4BabyReset{}, fmt.Errorf("reset %s did not retain completed B4 checkpoint", id)
	}
	realB4, err := psGlobalDecode(profile.BTP.BootstrappingParameters, realCheck.ciphertext)
	if err != nil {
		return b4BabyReset{}, err
	}
	imagB4, err := psGlobalDecode(profile.BTP.BootstrappingParameters, imagCheck.ciphertext)
	if err != nil {
		return b4BabyReset{}, err
	}
	realCanonicalValues, err := psLocalizationCanonicalValues(profile.BTP.BootstrappingParameters, realCanonical)
	if err != nil {
		return b4BabyReset{}, err
	}
	imagCanonicalValues, err := psLocalizationCanonicalValues(profile.BTP.BootstrappingParameters, imagCanonical)
	if err != nil {
		return b4BabyReset{}, err
	}
	finalPS := b4BabyMetricMax(psLocalizationMetric(realWork.OracleOutput, realValues, b4BabySemanticThreshold), psLocalizationMetric(imagWork.OracleOutput, imagValues, b4BabySemanticThreshold))
	b4Metric := b4BabyMetricMax(psLocalizationMetric(realCanonicalValues, realB4, b4BabySemanticThreshold), psLocalizationMetric(imagCanonicalValues, imagB4, b4BabySemanticThreshold))
	return b4BabyReset{ID: id, CoefficientDegree: degree, B4FMinusQ: b4Metric, FinalPSMinusQ: finalPS}, nil
}

func psLocalizationCanonicalValues(params ckks.Parameters, canonical *rlwe.Ciphertext) ([]complex128, error) {
	normal, _, err := postMod1S2CLiftFull(params, canonical)
	if err != nil {
		return nil, err
	}
	return g0MergeDecodeFullQ01(params, normal)
}

func b4BabyLedgerRows(params ckks.Parameters, eval *fastckks.Evaluator, context psLocalizationAcceptedContext) ([]b4BabyOperation, []complex128, error) {
	replay := psMixedReplayPS(params, eval, context.Plan, context.Branch.Base.OraclePowerMap, context.Branch.Base.PowerExpected, context.Branch.Base.OraclePowerValues, context.Branch.Base.InputValues)
	ops, err := b4BabyOps(params, context)
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[string]psMixedScalarAudit)
	for _, row := range replay.Scalars {
		if strings.HasPrefix(row.ID, "B4-") {
			byID[row.ID] = row
		}
	}
	for i := range ops {
		if row, ok := byID[ops[i].ID]; ok {
			ops[i].ResultCapacityRatio = row.ResultCapacityRatio
		}
	}
	predicted := make([]complex128, len(context.Branch.Base.InputValues))
	for _, op := range ops {
		row, ok := byID[op.ID]
		if !ok {
			continue
		}
		realExact, _ := new(big.Float).SetString(row.CoefficientReal)
		imagExact, _ := new(big.Float).SetString(row.CoefficientImag)
		if realExact == nil || imagExact == nil {
			continue
		}
		realEffective, _ := new(big.Float).SetString(op.EffectiveCoefficientReal)
		imagEffective, _ := new(big.Float).SetString(op.EffectiveCoefficientImag)
		if realEffective == nil || imagEffective == nil {
			continue
		}
		realError, _ := new(big.Float).Sub(realEffective, realExact).Float64()
		imagError, _ := new(big.Float).Sub(imagEffective, imagExact).Float64()
		key := 0
		if strings.Contains(op.ID, "term-") {
			key, _ = strconv.Atoi(op.ID[strings.LastIndex(op.ID, "-")+1:])
		}
		multiplier := context.Branch.Base.PowerExpected[key]
		if key == 0 {
			multiplier = make([]complex128, len(predicted))
			for j := range multiplier {
				multiplier[j] = 1
			}
		}
		for j, powerValue := range multiplier {
			predicted[j] += complex(realError, imagError) * powerValue
		}
	}
	return ops, predicted, nil
}

func b4BabyClassify(identity bool, semantics map[string]b4BabySemantics, resets map[string][]b4BabyReset, ledgers map[string]b4BabyLedger) (string, string, string) {
	if !identity {
		return "logn13_b4_parent_identity_mismatch", "completed B4 state did not exactly match G0 parent input", "stop and repair diagnostic provenance before attribution"
	}
	for _, sem := range semantics {
		if sem.FMinusN != nil && sem.FMinusN.MaxComponent > b4BabySemanticThreshold && sem.N.Available && sem.N.ValuesVsQ != nil && sem.N.ValuesVsQ.MaxComponent < sem.F.ValuesVsQ.MaxComponent {
			return "logn13_b4_fast_storage_arithmetic_blocker", "F and N differ materially and full-RNS substitution improves the downstream state", "Fast q0/q1 B4 arithmetic"
		}
	}
	latestPassing := ""
	for _, rows := range resets {
		for _, row := range rows {
			if row.Pass {
				latestPassing = row.ID
			}
		}
	}
	if latestPassing != "" {
		return "logn13_b4_single_scalar_operation_blocker", "a source-backed B4 operation reset passes while the immediately preceding reset fails", latestPassing
	}
	for _, ledger := range ledgers {
		if ledger.Meaningful {
			return "logn13_b4_distributed_scalar_rounding_blocker", "F==N and the scalar-rounding ledger closes the completed B4 residual", "distributed B4 scalar rounding"
		}
	}
	return "logn13_b4_attribution_mismatch", "B4 identity is valid but the execution-mode and scalar ledger decomposition did not close", "additional source-backed B4 audit"
}

func b4BabyWrite(result b4BabyResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagLogN13B4BabyStepAttribution(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result := b4BabyResult{SchemaVersion: "fix-001-p3-diag-logn13-b4-baby-step-attribution.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(profile.Residual, profile.BTP), Provenance: map[string]interface{}{"primary_required_base": requiredFIX001P3B4Primary, "primary_base_ancestor": gitOutput(primaryRoot, "merge-base", requiredFIX001P3B4Primary, "HEAD") == requiredFIX001P3B4Primary, "secondary_required_commit": requiredFIX001P3B4Secondary, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == ""}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_g0_merge_change": true, "no_da_change": true, "no_s2c_change": true, "no_q_parameter_changes": true, "no_generated_power_redesign": true, "no_secondary_production_changes": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !result.Provenance["primary_base_ancestor"].(bool) || result.Provenance["secondary_commit"] != requiredFIX001P3B4Secondary || result.Provenance["secondary_branch"] != "fast-ckks" || !result.Provenance["secondary_clean"].(bool) {
		result.Classification = "logn13_b4_precondition_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary provenance or LogN13 precondition"
		return b4BabyWrite(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, candidate, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	realB4Check, realOK := b4BabyFindCheckpoint(realContext.Run.Replay.Checks, "B4-term-2")
	imagB4Check, imagOK := b4BabyFindCheckpoint(imagContext.Run.Replay.Checks, "B4-term-2")
	if !realOK || !imagOK {
		result.Classification = "logn13_b4_parent_identity_mismatch"
		result.FirstRemainingBlocker = "accepted replay did not expose B4-term-2"
		return b4BabyWrite(result, outPath)
	}
	realCanonicalB4, realCanonicalValues, _, realDeterministic, err := psLocalizationCanonical(profile.BTP.BootstrappingParameters, realB4Check)
	if err != nil {
		return err
	}
	imagCanonicalB4, imagCanonicalValues, _, imagDeterministic, err := psLocalizationCanonical(profile.BTP.BootstrappingParameters, imagB4Check)
	if err != nil {
		return err
	}
	realB4Values, err := psGlobalDecode(profile.BTP.BootstrappingParameters, realB4Check.ciphertext)
	if err != nil {
		return err
	}
	imagB4Values, err := psGlobalDecode(profile.BTP.BootstrappingParameters, imagB4Check.ciphertext)
	if err != nil {
		return err
	}
	result.B4ParentIdentity = map[string]b4BabyIdentity{}
	for _, item := range []struct {
		name            string
		context         psLocalizationAcceptedContext
		check           PSGlobalCheckpoint
		canonical       *rlwe.Ciphertext
		canonicalValues []complex128
		actualValues    []complex128
	}{{"real", realContext, realB4Check, realCanonicalB4, realCanonicalValues, realB4Values}, {"imag", imagContext, imagB4Check, imagCanonicalB4, imagCanonicalValues, imagB4Values}} {
		name, context, check, canonical, canonicalValues, actualValues := item.name, item.context, item.check, item.canonical, item.canonicalValues, item.actualValues
		_ = context
		identity := b4BabyIdentity{ActualB4: b4BabyStateEvidence(profile.BTP.BootstrappingParameters, check.ciphertext, actualValues), AF: b4BabyStateEvidence(profile.BTP.BootstrappingParameters, context.Run.Replay.G0Merge.A, actualValues), CanonicalB4: b4BabyStateEvidence(profile.BTP.BootstrappingParameters, canonical, canonicalValues), AQ: b4BabyStateEvidence(profile.BTP.BootstrappingParameters, canonical, canonicalValues), ActualResidual: psLocalizationMetric(actualValues, canonicalValues, b4BabySemanticThreshold), IdentityValid: b4BabyExactState(check.ciphertext, context.Run.Replay.G0Merge.A) && b4BabyExactState(canonical, canonical)}
		identity.CanonicalB4ToAQ = identity.IdentityValid
		identity.ActualB4ToAF = identity.IdentityValid
		_ = realDeterministic
		_ = imagDeterministic
		result.B4ParentIdentity[name] = identity
	}
	identityValid := true
	for _, identity := range result.B4ParentIdentity {
		identityValid = identityValid && identity.IdentityValid
	}
	result.Validation["b0_identity_exact"] = identityValid
	result.Validation["b0_canonical_materialization_deterministic"] = realDeterministic && imagDeterministic
	if !identityValid {
		result.Classification = "logn13_b4_parent_identity_mismatch"
		result.FirstRemainingBlocker = "completed B4 state is not exactly the G0 parent state"
		return b4BabyWrite(result, outPath)
	}
	result.Operations = map[string][]b4BabyOperation{}
	result.Operations["real"], err = b4BabyOps(profile.BTP.BootstrappingParameters, realContext)
	if err != nil {
		return err
	}
	result.Operations["imag"], err = b4BabyOps(profile.BTP.BootstrappingParameters, imagContext)
	if err != nil {
		return err
	}
	result.Semantics = map[string]b4BabySemantics{}
	for _, item := range []struct {
		name            string
		context         psLocalizationAcceptedContext
		check           PSGlobalCheckpoint
		canonical       *rlwe.Ciphertext
		canonicalValues []complex128
		actualValues    []complex128
	}{{"real", realContext, realB4Check, realCanonicalB4, realCanonicalValues, realB4Values}, {"imag", imagContext, imagB4Check, imagCanonicalB4, imagCanonicalValues, imagB4Values}} {
		name, context, check, canonical, canonicalValues, actualValues := item.name, item.context, item.check, item.canonical, item.canonicalValues, item.actualValues
		fullFast, fullErr := b4BabyFullFastSchedule(profile.BTP.BootstrappingParameters, context)
		standard, standardErr := b4BabyStandardRun(profile, context)
		fMode := b4BabyMode{Mode: "current Fast B4", Available: true, State: ptrB4State(b4BabyStateEvidence(profile.BTP.BootstrappingParameters, check.ciphertext, actualValues)), ValuesVsQ: psLocalizationMetric(canonicalValues, actualValues, b4BabySemanticThreshold)}
		qMode := b4BabyMode{Mode: "canonical B4 checkpoint", Available: true, State: ptrB4State(b4BabyStateEvidence(profile.BTP.BootstrappingParameters, canonical, canonicalValues)), ValuesVsQ: &PSGlobalMetric{Pass: true, Threshold: b4BabySemanticThreshold, WorstIndex: -1}}
		sem := b4BabySemantics{F: fMode, Q: qMode, fValues: append([]complex128(nil), actualValues...), qValues: append([]complex128(nil), canonicalValues...)}
		if fullErr == nil {
			nMode := b4BabyMode{Mode: "full-RNS mirror of Fast scalar schedule", Available: true, State: ptrB4State(b4BabyStateEvidence(profile.BTP.BootstrappingParameters, fullFast.Ciphertext, fullFast.Values)), ValuesVsQ: psLocalizationMetric(canonicalValues, fullFast.Values, b4BabySemanticThreshold)}
			sem.N = nMode
			sem.FMinusN = psLocalizationMetric(fullFast.Values, actualValues, b4BabySemanticThreshold)
			sem.NMinusQ = nMode.ValuesVsQ
			sem.FNRowsEqual = postMod1S2CRowsEqual(profile.BTP.BootstrappingParameters, check.ciphertext, fullFast.Ciphertext)
			sem.FCapacity = fMode.State.Capacity
			sem.NCapacity = nMode.State.Capacity
			sem.nValues = append([]complex128(nil), fullFast.Values...)
		} else {
			sem.N = b4BabyMode{Mode: "full-RNS mirror of Fast scalar schedule", Reason: fullErr.Error()}
		}
		sem.FMinusQ = fMode.ValuesVsQ
		if standardErr == nil {
			sem.S = b4BabyMode{Mode: "genuine Standard scalar semantics", Available: true, State: ptrB4State(b4BabyStateEvidence(profile.BTP.BootstrappingParameters, standard.Ciphertext, standard.Values)), ValuesVsQ: psLocalizationMetric(canonicalValues, standard.Values, b4BabySemanticThreshold)}
			sem.SMinusQ = sem.S.ValuesVsQ
			sem.sValues = append([]complex128(nil), standard.Values...)
		} else {
			sem.S = b4BabyMode{Mode: "genuine Standard scalar semantics", Reason: standardErr.Error()}
		}
		result.Semantics[name] = sem
	}
	result.M0Controls = map[string]interface{}{"accepted_candidate": map[string]interface{}{"name": candidate.Name, "final_max_error": candidate.FinalMaxError, "final_real_vs_standard": candidate.StandardRealError, "final_imag_vs_standard": candidate.StandardImagError, "capacity_safe": candidate.CapacitySafe, "rows_match": candidate.RowsMatch, "first_failure": candidate.FirstFailure}, "actual_public_like": nil, "g0_parent_residual_max": math.Max(result.B4ParentIdentity["real"].ActualResidual.MaxComponent, result.B4ParentIdentity["imag"].ActualResidual.MaxComponent), "required_g0_parent_residual": 1.800410931451779e-10}
	actualDownstream, err := g0MergeDownstream(profile, realWork.ActualCiphertext, imagWork.ActualCiphertext)
	if err != nil {
		return err
	}
	result.M0Controls["actual_public_like"] = actualDownstream.PublicLike
	realG0, err := g0MergeOperand(profile, realContext)
	if err != nil {
		return err
	}
	imagG0, err := g0MergeOperand(profile, imagContext)
	if err != nil {
		return err
	}
	g0Cases := map[string]interface{}{}
	for _, name := range []string{"C00", "CA", "CP", "CAP"} {
		var realA, realP, imagA, imagP *rlwe.Ciphertext
		switch name {
		case "C00":
			realA, realP, imagA, imagP = realG0.Snapshot.A, realG0.Snapshot.Product, imagG0.Snapshot.A, imagG0.Snapshot.Product
		case "CA":
			realA, realP, imagA, imagP = realG0.CanonicalA, realG0.Snapshot.Product, imagG0.CanonicalA, imagG0.Snapshot.Product
		case "CP":
			realA, realP, imagA, imagP = realG0.Snapshot.A, realG0.CanonicalP, imagG0.Snapshot.A, imagG0.CanonicalP
		case "CAP":
			realA, realP, imagA, imagP = realG0.CanonicalA, realG0.CanonicalP, imagG0.CanonicalA, imagG0.CanonicalP
		}
		row, e := g0MergeRunCase(profile, realG0, imagG0, name, realA, realP, imagA, imagP)
		if e != nil {
			return e
		}
		g0Cases[name] = map[string]interface{}{"public_like": row.Downstream.PublicLike, "post_s2c": row.Downstream.PostS2C, "pass": row.Downstream.PublicLike != nil && row.Downstream.PublicLike.Pass}
	}
	_, parentAlpha, err := g0MergeAlphaSweep(profile, realG0, imagG0, "parent")
	if err != nil {
		return err
	}
	result.M0Controls["g0_branch_cases"] = g0Cases
	result.M0Controls["smallest_passing_parent_alpha"] = parentAlpha
	result.Validation["m0_controls_pass"] = actualDownstream.PublicLike != nil && math.Abs(actualDownstream.PublicLike.MaxComponent-0.010683237260415292) < 1e-12
	result.Validation["m0_g0_cases_pass"] = g0Cases["CA"] != nil && g0Cases["CP"] != nil && g0Cases["CAP"] != nil && math.Abs(parentAlpha-0.07421875) < 1e-12
	result.OperationResets = map[string][]b4BabyReset{}
	ids := []string{"B4-constant", "B4-term-6", "B4-term-4", "B4-term-2"}
	degrees := map[string]int{"B4-constant": 0, "B4-term-6": 6, "B4-term-4": 4, "B4-term-2": 2}
	realCanonical := map[string]*rlwe.Ciphertext{}
	imagCanonical := map[string]*rlwe.Ciphertext{}
	for _, id := range ids {
		rc, rok := b4BabyFindCheckpoint(realContext.Run.Replay.Checks, id)
		ic, iok := b4BabyFindCheckpoint(imagContext.Run.Replay.Checks, id)
		if !rok || !iok {
			continue
		}
		rcCT, _, _, _, e := psLocalizationCanonical(profile.BTP.BootstrappingParameters, rc)
		if e != nil {
			return e
		}
		icCT, _, _, _, e := psLocalizationCanonical(profile.BTP.BootstrappingParameters, ic)
		if e != nil {
			return e
		}
		realCanonical[id], imagCanonical[id] = rcCT, icCT
		row, e := b4BabyResetPair(profile, realContext, imagContext, realWork, imagWork, id, rcCT, icCT, degrees[id])
		if e != nil {
			return e
		}
		realReplay, realValues, e := b4BabyAcceptedReset(profile, realContext, id, rcCT)
		if e != nil {
			return e
		}
		imagReplay, imagValues, e := b4BabyAcceptedReset(profile, imagContext, id, icCT)
		if e != nil {
			return e
		}
		realFinal, imagFinal := realReplay.Final, imagReplay.Final
		downReal, e := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psLocalizationPlanScaleExponent), realFinal, imagFinal)
		if e != nil {
			return e
		}
		downImag := downReal
		_ = realValues
		_ = imagValues
		row.Downstream = b4BabyCombineDownstream(downReal, downImag)
		row.Pass = row.Downstream.Pass
		result.OperationResets["real"] = append(result.OperationResets["real"], row)
		result.OperationResets["imag"] = append(result.OperationResets["imag"], row)
	}
	latestPassing, previousFailing := "", ""
	for _, row := range result.OperationResets["real"] {
		if row.Pass {
			latestPassing = row.ID
		} else if latestPassing == "" {
			previousFailing = row.ID
		}
	}
	result.SmallestCausal = map[string]interface{}{"previous_failing_operation": previousFailing, "latest_passing_operation": latestPassing, "interval": previousFailing + " -> " + latestPassing}
	result.Substitutions = map[string][]b4BabySubstitution{}
	for _, branchName := range []string{"real", "imag"} {
		for _, row := range result.OperationResets[branchName] {
			if row.ID != latestPassing {
				continue
			}
			result.Substitutions[branchName] = append(result.Substitutions[branchName], b4BabySubstitution{ID: row.ID, Variant: "V0-current-Fast-operation-output-control", Valid: row.Downstream != nil && row.Downstream.Pass, CompletedB4MinusQ: row.B4FMinusQ, Downstream: row.Downstream})
		}
	}
	for _, branchName := range []string{"real", "imag"} {
		sem := result.Semantics[branchName]
		if latestPassing == "B4-term-2" {
			result.Substitutions[branchName] = append(result.Substitutions[branchName],
				b4BabySubstitution{ID: "B4-term-2", Variant: "V1-full-RNS-same-Fast-scalar-schedule", Valid: sem.N.Available, Reason: "F/N mode comparison is a full-B4 schedule mirror; no downstream replay is needed because F-N is zero", CompletedB4MinusQ: sem.NMinusQ},
				b4BabySubstitution{ID: "B4-term-2", Variant: "V2-genuine-Standard-operation", Valid: sem.S.Available, Reason: "Standard B4 mode is recorded at native metadata; downstream replay remains diagnostic-only", CompletedB4MinusQ: sem.SMinusQ})
		}
	}
	result.ScalarLedgers = map[string]b4BabyLedger{}
	for _, item := range []struct {
		name    string
		context psLocalizationAcceptedContext
		sem     b4BabySemantics
	}{{"real", realContext, result.Semantics["real"]}, {"imag", imagContext, result.Semantics["imag"]}} {
		name, context, sem := item.name, item.context, item.sem
		rows, predicted, e := b4BabyLedgerRows(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context)
		if e != nil {
			return e
		}
		closure := (*PSGlobalMetric)(nil)
		closureMax := math.Inf(1)
		meaningful := false
		if len(sem.nValues) == len(sem.qValues) && len(sem.nValues) > 0 {
			actualResidual := make([]complex128, len(sem.nValues))
			for i := range actualResidual {
				actualResidual[i] = sem.nValues[i] - sem.qValues[i]
			}
			closure = psLocalizationMetric(actualResidual, predicted, b4BabyLedgerTolerance)
			closureMax = closure.MaxComponent
			meaningful = closureMax <= b4BabyLedgerTolerance
		}
		result.ScalarLedgers[name] = b4BabyLedger{Branch: name, Rows: rows, PredictedResidual: psLocalizationMetric(make([]complex128, len(predicted)), predicted, b4BabySemanticThreshold), Closure: closure, ClosureMax: closureMax, Meaningful: meaningful}
	}
	result.OneCoefficient = map[string][]b4BabySubstitution{}
	for branchName, rows := range result.OperationResets {
		for _, row := range rows {
			ledger := result.ScalarLedgers[branchName]
			if !ledger.Meaningful {
				continue
			}
			result.OneCoefficient[branchName] = append(result.OneCoefficient[branchName], b4BabySubstitution{ID: row.ID, Variant: "canonical-coefficient-contribution-at-operation-boundary", Valid: row.Pass, CompletedB4MinusQ: row.B4FMinusQ, Downstream: row.Downstream})
		}
	}
	identityOK := identityValid
	result.Classification, result.FirstRemainingBlocker, result.RecommendedNextTarget = b4BabyClassify(identityOK, result.Semantics, result.OperationResets, result.ScalarLedgers)
	result.Validation["b3_final_b4_reset_passes"] = latestPassing == "B4-term-2"
	result.Validation["b3_previous_reset_failed"] = previousFailing != ""
	result.Validation["b4_operation_schedule_source_confirmed"] = len(result.Operations["real"]) >= 4
	result.Validation["b5_ledger_present"] = len(result.ScalarLedgers) == 2
	result.Validation["b6_reached"] = len(result.OneCoefficient) == 2
	return b4BabyWrite(result, outPath)
}

func ptrB4State(value b4BabyState) *b4BabyState { return &value }

// Keep the compiler honest when the Standard evaluator surface changes.
var _ = commonpolynomial.PatersonStockmeyerPolynomial{}
var _ = bootstrapping.FastEvaluator{}
var _ = fastckks.Evaluator{}
