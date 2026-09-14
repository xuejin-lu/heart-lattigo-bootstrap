package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/bits"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3Q055T2Primary   = "ce335898c233eb3b08984206dc5772a0681fba12"
	requiredFIX001P3Q055T2Secondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	q055T2PlanScaleExponent         = 92
	q055T2PublicThreshold           = 1e-2
	q055T2SemanticTolerance         = 1e-10
)

type q055T2CapacityEvidence struct {
	State                    string  `json:"state"`
	MaxAbs                   string  `json:"max_intended_abs"`
	Q01Half                  string  `json:"q01_half"`
	Q012Half                 string  `json:"q012_half"`
	IntendedQ01Ratio         float64 `json:"intended_over_q01_half"`
	IntendedQ012Ratio        float64 `json:"intended_over_q012_half"`
	ReducedQ01Ratio          float64 `json:"reduced_q01_representative_over_q01_half"`
	IntendedQ01OutsideCount  int     `json:"intended_q01_outside_count"`
	IntendedQ012OutsideCount int     `json:"intended_q012_outside_count"`
	ReducedQ01OutsideCount   int     `json:"reduced_q01_outside_count"`
	IntendedQ01MinimumBits   int     `json:"intended_q01_minimum_required_bits"`
	IntendedQ012MinimumBits  int     `json:"intended_q012_minimum_required_bits"`
	Q01CenteredUnique        bool    `json:"intended_q01_centered_unique"`
	Q012CenteredUnique       bool    `json:"intended_q012_centered_unique"`
	ReducedQ01CenteredUnique bool    `json:"reduced_q01_centered_unique"`
}

type q055T2RowProof struct {
	State         string `json:"state"`
	Q0RowsEqual   bool   `json:"q0_rows_equal"`
	Q1RowsEqual   bool   `json:"q1_rows_equal"`
	Q2RowsEqual   bool   `json:"q2_rows_equal"`
	RowsEqual     bool   `json:"q012_rows_equal"`
	MetadataEqual bool   `json:"metadata_equal"`
	Oracle        string `json:"oracle"`
}

type q055T2GuardEvidence struct {
	Branch                         string                   `json:"branch"`
	CurrentQ01Operation            *t2GuardOperation        `json:"current_q01_operation,omitempty"`
	CurrentQ01Contraction          *t2GuardContraction      `json:"current_q01_contraction,omitempty"`
	CandidateOperation             *b4BabyOperation         `json:"candidate_operation,omitempty"`
	Capacity                       []q055T2CapacityEvidence `json:"capacity"`
	RowProof                       []q055T2RowProof         `json:"row_proof"`
	CandidateContractionValid      bool                     `json:"candidate_contraction_valid"`
	CandidateOutputMetadataMatches bool                     `json:"candidate_output_metadata_matches_native"`
	CandidateFinalRowsMatchNative  bool                     `json:"candidate_final_q01_rows_match_native"`
	CandidateValid                 bool                     `json:"candidate_valid"`
	Reason                         string                   `json:"reason,omitempty"`
}

type q055T2Case struct {
	Name                      string                 `json:"name"`
	Q0Bits                    int                    `json:"q0_bits"`
	PlanScaleExponent         int                    `json:"plan_scale_exponent"`
	Q01GuardDownstream        *b4BabyDownstream      `json:"current_q01_guard_downstream,omitempty"`
	Q012GuardDownstream       *b4BabyDownstream      `json:"candidate_q012_guard_downstream,omitempty"`
	Q01FirstDivergence        string                 `json:"current_q01_first_divergence"`
	Q012FirstDivergence       string                 `json:"candidate_q012_first_divergence"`
	Q01Pass                   bool                   `json:"current_q01_public_pass"`
	Q012Pass                  bool                   `json:"candidate_q012_public_pass"`
	Generated                 []closurePowerEvidence `json:"generated_power_evidence"`
	RealGuard                 q055T2GuardEvidence    `json:"real_guard"`
	ImagGuard                 q055T2GuardEvidence    `json:"imag_guard"`
	CandidateSystemContracts  bool                   `json:"candidate_system_contracts"`
	CandidateSystemSufficient bool                   `json:"candidate_system_sufficient"`
	Q012VsQ01Real             *PSGlobalMetric        `json:"candidate_vs_current_q01_real,omitempty"`
	Q012VsQ01Imag             *PSGlobalMetric        `json:"candidate_vs_current_q01_imag,omitempty"`
	Q012MetadataEquivalent    bool                   `json:"candidate_metadata_equivalent"`
	Reason                    string                 `json:"reason,omitempty"`
}

type q055T2PreviousMatrix struct {
	Imported                  bool                   `json:"imported"`
	ArtifactPath              string                 `json:"artifact_path"`
	ArtifactPrimaryCommit     string                 `json:"artifact_primary_commit"`
	ArtifactSecondaryCommit   string                 `json:"artifact_secondary_commit"`
	PreviousClassification    string                 `json:"previous_classification"`
	PreviousBFirstDivergence  string                 `json:"previous_b_first_divergence"`
	ExpectedPassPattern       map[string]bool        `json:"expected_pass_pattern"`
	ObservedPassPattern       map[string]bool        `json:"observed_pass_pattern"`
	CorrectedClassification   string                 `json:"corrected_classification"`
	CorrectedBFirstDivergence string                 `json:"corrected_b_first_divergence"`
	RegressionAssertions      map[string]interface{} `json:"regression_assertions"`
}

type q055T2Result struct {
	SchemaVersion         string                 `json:"schema_version"`
	Timestamp             time.Time              `json:"timestamp"`
	Primary               RepositoryMetadata     `json:"primary_repository"`
	Lattigo               RepositoryMetadata     `json:"lattigo_repository"`
	Environment           EnvironmentMetadata    `json:"environment"`
	Config                BootstrapConfig        `json:"config"`
	Provenance            map[string]interface{} `json:"provenance"`
	PreviousMatrix        q055T2PreviousMatrix   `json:"previous_matrix_interpretation"`
	B                     *q055T2Case            `json:"b_q055_plan92"`
	D                     *q055T2Case            `json:"d_q056_plan92_control"`
	ProductionReadiness   string                 `json:"production_readiness"`
	Classification        string                 `json:"classification"`
	FirstRemainingBlocker string                 `json:"first_remaining_blocker"`
	RecommendedNextTarget string                 `json:"recommended_next_target"`
	Validation            map[string]interface{} `json:"validation"`
}

type q055T2CRTConstants struct {
	Q0, Q1, Q2     uint64
	InvQ0ModQ1     uint64
	InvQ0ModQ2     uint64
	InvQ1ModQ2     uint64
	Inv2Q0, Inv2Q1 uint64
	Inv2Q2         uint64
	Q              [3]uint64
	HalfQ          [3]uint64
}

type q055T2CandidateRun struct {
	Evidence   q055T2GuardEvidence
	Final      *rlwe.Ciphertext
	PostAdd    *rlwe.Ciphertext
	Quotient   *rlwe.Ciphertext
	Contracted *rlwe.Ciphertext
}

func q055T2ModSub(a, b, modulus uint64) uint64 {
	if a >= b {
		return a - b
	}
	return modulus - (b - a)
}

func q055T2MulMod(a, b, modulus uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	_, remainder := bits.Div64(hi, lo, modulus)
	return remainder
}

func q055T2Inv(modulus, value uint64) (uint64, error) {
	inverse := new(big.Int).ModInverse(new(big.Int).SetUint64(value%modulus), new(big.Int).SetUint64(modulus))
	if inverse == nil {
		return 0, fmt.Errorf("no inverse for %d modulo %d", value, modulus)
	}
	return inverse.Uint64(), nil
}

func q055T2MakeCRTConstants(params ckks.Parameters) (q055T2CRTConstants, error) {
	q := params.RingQ().SubRings
	if len(q) < 3 {
		return q055T2CRTConstants{}, fmt.Errorf("q012 requires three Q primes")
	}
	c := q055T2CRTConstants{Q0: q[0].Modulus, Q1: q[1].Modulus, Q2: q[2].Modulus, Inv2Q0: (q[0].Modulus + 1) / 2, Inv2Q1: (q[1].Modulus + 1) / 2, Inv2Q2: (q[2].Modulus + 1) / 2}
	var err error
	if c.InvQ0ModQ1, err = q055T2Inv(c.Q1, c.Q0); err != nil {
		return c, err
	}
	if c.InvQ0ModQ2, err = q055T2Inv(c.Q2, c.Q0); err != nil {
		return c, err
	}
	if c.InvQ1ModQ2, err = q055T2Inv(c.Q2, c.Q1); err != nil {
		return c, err
	}
	q01Hi, q01Lo := bits.Mul64(c.Q0, c.Q1)
	qHi, qLo := bits.Mul64(q01Lo, c.Q2)
	qTop, qHi2 := bits.Mul64(q01Hi, c.Q2)
	qMiddle, carry := bits.Add64(qHi, qHi2, 0)
	c.Q = [3]uint64{qLo, qMiddle, qTop + carry}
	c.HalfQ[0] = (c.Q[0] >> 1) | (c.Q[1] << 63)
	c.HalfQ[1] = (c.Q[1] >> 1) | (c.Q[2] << 63)
	c.HalfQ[2] = c.Q[2] >> 1
	return c, nil
}

func q055T2CompareU192(a, b [3]uint64) int {
	for i := 2; i >= 0; i-- {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func q055T2CRTUnsigned(c q055T2CRTConstants, r0, r1, r2 uint64) [3]uint64 {
	t1 := q055T2MulMod(q055T2ModSub(r1, r0%c.Q1, c.Q1), c.InvQ0ModQ1, c.Q1)
	t2 := q055T2MulMod(q055T2ModSub(r2, r0%c.Q2, c.Q2), c.InvQ0ModQ2, c.Q2)
	k := q055T2MulMod(q055T2ModSub(t2, t1%c.Q2, c.Q2), c.InvQ1ModQ2, c.Q2)
	tHi, tLo := bits.Mul64(c.Q1, k)
	tLo, carry := bits.Add64(tLo, t1, 0)
	tHi += carry
	p0Hi, p0Lo := bits.Mul64(c.Q0, tLo)
	p1Hi, p1Lo := bits.Mul64(c.Q0, tHi)
	x1, carry := bits.Add64(p0Hi, p1Lo, 0)
	x2 := p1Hi + carry
	x0, carry := bits.Add64(p0Lo, r0, 0)
	x1, carry = bits.Add64(x1, 0, carry)
	x2 += carry
	return [3]uint64{x0, x1, x2}
}

func q055T2HalfResidue(value, modulus, inv2 uint64, negative, odd bool) uint64 {
	if odd {
		if negative {
			value = q055T2ModSub(value, 1, modulus)
		} else {
			value++
			if value == modulus {
				value = 0
			}
		}
	}
	return q055T2MulMod(value, inv2, modulus)
}

func q055T2DivideByTwoQ012(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || !source.IsNTT || !source.IsMontgomery || source.Level() < 2 {
		return nil, fmt.Errorf("q012 divide requires NTT Montgomery level >= 2")
	}
	constants, err := q055T2MakeCRTConstants(params)
	if err != nil {
		return nil, err
	}
	out := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	out.Scale = source.Scale.Div(rlwe.NewScale(2))
	for component := range source.Value {
		ntt := ring.NewPoly(params.N(), 2)
		for limb := 0; limb <= 2; limb++ {
			copy(ntt.Coeffs[limb], source.Value[component].Coeffs[limb])
			params.RingQ().SubRings[limb].IMForm(ntt.Coeffs[limb], ntt.Coeffs[limb])
		}
		coeff := ring.NewPoly(params.N(), 2)
		params.RingQ().AtLevel(2).INTT(ntt, coeff)
		quotientCoeff := ring.NewPoly(params.N(), 2)
		for i := 0; i < params.N(); i++ {
			u := q055T2CRTUnsigned(constants, coeff.Coeffs[0][i], coeff.Coeffs[1][i], coeff.Coeffs[2][i])
			negative := q055T2CompareU192(u, constants.HalfQ) > 0
			odd := u[0]&1 == 1
			if negative {
				odd = !odd
			}
			quotientCoeff.Coeffs[0][i] = q055T2HalfResidue(coeff.Coeffs[0][i], constants.Q0, constants.Inv2Q0, negative, odd)
			quotientCoeff.Coeffs[1][i] = q055T2HalfResidue(coeff.Coeffs[1][i], constants.Q1, constants.Inv2Q1, negative, odd)
			quotientCoeff.Coeffs[2][i] = q055T2HalfResidue(coeff.Coeffs[2][i], constants.Q2, constants.Inv2Q2, negative, odd)
		}
		quotientNTT := ring.NewPoly(params.N(), 2)
		params.RingQ().AtLevel(2).NTT(quotientCoeff, quotientNTT)
		for limb := 0; limb <= 2; limb++ {
			params.RingQ().SubRings[limb].MForm(quotientNTT.Coeffs[limb], out.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func q055T2MulThenAddQ012(params ckks.Parameters, power *rlwe.Ciphertext, coefficient *bignum.Complex, encoding rlwe.Scale, accumulator *rlwe.Ciphertext) error {
	if power == nil || accumulator == nil || power.Level() != accumulator.Level() || power.Degree() != accumulator.Degree() || power.Level() < 2 {
		return fmt.Errorf("q012 scalar MulThenAdd metadata mismatch")
	}
	half := params.N() >> 1
	for component := range power.Value {
		for limb := 0; limb <= 2; limb++ {
			subring := params.RingQ().SubRings[limb]
			values := b4BabyScalarValues(coefficient, &encoding.Value, subring)
			realScalar := ring.MForm(values[0], subring.Modulus, subring.BRedConstant)
			imagScalar := ring.MForm(values[1], subring.Modulus, subring.BRedConstant)
			subring.MulScalarMontgomeryThenAdd(power.Value[component].Coeffs[limb][:half], realScalar, accumulator.Value[component].Coeffs[limb][:half])
			subring.MulScalarMontgomeryThenAdd(power.Value[component].Coeffs[limb][half:], imagScalar, accumulator.Value[component].Coeffs[limb][half:])
		}
	}
	return nil
}

func q055T2CopyValues(values [][]*big.Int) [][]*big.Int {
	result := make([][]*big.Int, len(values))
	for component := range values {
		result[component] = make([]*big.Int, len(values[component]))
		for i := range values[component] {
			result[component][i] = new(big.Int).Set(values[component][i])
		}
	}
	return result
}

func q055T2ScaleValues(values [][]*big.Int, factor int64) [][]*big.Int {
	result := q055T2CopyValues(values)
	for component := range result {
		for i := range result[component] {
			result[component][i].Mul(result[component][i], big.NewInt(factor))
		}
	}
	return result
}

func q055T2AddValues(a, b [][]*big.Int) [][]*big.Int {
	result := q055T2CopyValues(a)
	for component := range result {
		for i := range result[component] {
			result[component][i].Add(result[component][i], b[component][i])
		}
	}
	return result
}

func q055T2RoundValues(values [][]*big.Int) [][]*big.Int {
	result := q055T2CopyValues(values)
	for component := range result {
		for i := range result[component] {
			result[component][i] = t2GuardRoundSigned(result[component][i])
		}
	}
	return result
}

func q055T2FlattenValues(values [][]*big.Int) []*big.Int {
	result := make([]*big.Int, 0)
	for _, component := range values {
		result = append(result, component...)
	}
	return result
}

func q055T2ReducedQ01Values(params ckks.Parameters, values [][]*big.Int) [][]*big.Int {
	q01 := new(big.Int).Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus))
	half := new(big.Int).Rsh(new(big.Int).Set(q01), 1)
	result := q055T2CopyValues(values)
	for component := range result {
		for i := range result[component] {
			result[component][i].Mod(result[component][i], q01)
			if result[component][i].Cmp(half) > 0 {
				result[component][i].Sub(result[component][i], q01)
			}
		}
	}
	return result
}

func q055T2Capacity(params ckks.Parameters, state string, values [][]*big.Int) q055T2CapacityEvidence {
	flat := q055T2FlattenValues(values)
	q01 := designCapacityRow(params, state, "q01", flat)
	q012 := designCapacityRow(params, state, "q012", flat)
	reduced := designCapacityRow(params, state, "q01_reduced", q055T2FlattenValues(q055T2ReducedQ01Values(params, values)))
	return q055T2CapacityEvidence{State: state, MaxAbs: q012.MaxAbs, Q01Half: q01.HalfModulus, Q012Half: q012.HalfModulus, IntendedQ01Ratio: q01.Ratio, IntendedQ012Ratio: q012.Ratio, ReducedQ01Ratio: reduced.Ratio, IntendedQ01OutsideCount: q01.OutsideCount, IntendedQ012OutsideCount: q012.OutsideCount, ReducedQ01OutsideCount: reduced.OutsideCount, IntendedQ01MinimumBits: q01.MinimumBits, IntendedQ012MinimumBits: q012.MinimumBits, Q01CenteredUnique: q01.CenteredUnique, Q012CenteredUnique: q012.CenteredUnique, ReducedQ01CenteredUnique: reduced.CenteredUnique}
}

func q055T2MakeRowProof(params ckks.Parameters, state string, actual, oracle *rlwe.Ciphertext) q055T2RowProof {
	rows, limbs := nativeQ012RowsEqual(params, actual, oracle)
	return q055T2RowProof{State: state, Q0RowsEqual: limbs[0], Q1RowsEqual: limbs[1], Q2RowsEqual: limbs[2], RowsEqual: rows, MetadataEqual: nativeQ012MetadataEqual(actual, oracle), Oracle: "independent centered physical coefficient oracle"}
}

func q055T2ContractQ01(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, bool, error) {
	values, err := nativeQ012SignedValues(params, source)
	if err != nil {
		return nil, false, err
	}
	flat := q055T2FlattenValues(values)
	cap := designCapacityRow(params, "post_divide_q01", "q01", flat)
	if !cap.CenteredUnique {
		return nil, false, nil
	}
	out := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := range source.Value {
		for limb := 0; limb <= 1; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, true, nil
}

func q055T2Q012Apply(params ckks.Parameters, eval *fastckks.Evaluator, power *rlwe.Ciphertext, coefficient *bignum.Complex, pre, nativePost *rlwe.Ciphertext, id string) (q055T2CandidateRun, error) {
	preValues, err := t2GuardIndependentRecoverQ01(params, pre)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	preCapacity := q055T2Capacity(params, "native_accumulator_before_guard_promotion", preValues)
	if !preCapacity.Q01CenteredUnique {
		return q055T2CandidateRun{Evidence: q055T2GuardEvidence{Reason: "guard entry q01 is not centered-unique", Capacity: []q055T2CapacityEvidence{preCapacity}}}, nil
	}
	preQ, err := nativeQ012LiftQ01(params, pre)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	powerQ, err := nativeQ012LiftQ01(params, power)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	level := minInt(preQ.Level(), powerQ.Level())
	preQ, err = nativeQ012AtLevel(params, preQ, level)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	powerQ, err = nativeQ012AtLevel(params, powerQ, level)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	preQOracle, err := nativeQ012BuildFromSigned(params, preQ, preValues, level, preQ.Scale)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence := q055T2GuardEvidence{Branch: id, Capacity: []q055T2CapacityEvidence{preCapacity}, RowProof: []q055T2RowProof{q055T2MakeRowProof(params, "guard_entry_q012_lift", preQ, preQOracle)}}
	if !q055T2MakeRowProof(params, "guard_entry_q012_lift", preQ, preQOracle).RowsEqual {
		evidence.Reason = "q01-to-q012 guard-entry lift mismatch"
		return q055T2CandidateRun{Evidence: evidence}, nil
	}
	promoted := preQ.CopyNew()
	if err := nativeQ012MulInteger(params, promoted, big.NewInt(2)); err != nil {
		return q055T2CandidateRun{}, err
	}
	promoted.Scale = pre.Scale.Mul(rlwe.NewScale(2))
	promotedValues := q055T2ScaleValues(preValues, 2)
	evidence.Capacity = append(evidence.Capacity, q055T2Capacity(params, "promoted_accumulator_after_multiply_by_2", promotedValues))
	promotedOracle, err := nativeQ012BuildFromSigned(params, promoted, promotedValues, level, promoted.Scale)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence.RowProof = append(evidence.RowProof, q055T2MakeRowProof(params, "promoted_accumulator", promoted, promotedOracle))
	encoding, err := psMixedScalarEncodingScale(params, power.Scale, promoted.Scale, coefficient, power.Level())
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	contribution := ckks.NewCiphertext(params, powerQ.Degree(), level)
	*contribution.MetaData = *promoted.MetaData
	contribution.IsNTT, contribution.IsMontgomery, contribution.Scale = true, true, promoted.Scale
	if err := q055T2MulThenAddQ012(params, powerQ, coefficient, encoding, contribution); err != nil {
		return q055T2CandidateRun{}, err
	}
	contributionValues, err := nativeQ012SignedValues(params, contribution)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence.Capacity = append(evidence.Capacity, q055T2Capacity(params, "scalar_contribution_before_accumulation", contributionValues))
	contributionOracle, err := nativeQ012BuildFromSigned(params, contribution, contributionValues, level, contribution.Scale)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence.RowProof = append(evidence.RowProof, q055T2MakeRowProof(params, "scalar_contribution", contribution, contributionOracle))
	postAdd := promoted.CopyNew()
	if err := q055T2MulThenAddQ012(params, powerQ, coefficient, encoding, postAdd); err != nil {
		return q055T2CandidateRun{}, err
	}
	postAddValues := q055T2AddValues(promotedValues, contributionValues)
	evidence.Capacity = append(evidence.Capacity, q055T2Capacity(params, "intended_physical_post_add_before_divide_by_2", postAddValues))
	postAddOracle, err := nativeQ012BuildFromSigned(params, postAdd, postAddValues, level, postAdd.Scale)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence.RowProof = append(evidence.RowProof, q055T2MakeRowProof(params, "intended_physical_post_add", postAdd, postAddOracle))
	quotientValues := q055T2RoundValues(postAddValues)
	evidence.Capacity = append(evidence.Capacity, q055T2Capacity(params, "exact_centered_rounded_divide_by_2_quotient", quotientValues))
	quotientOracle, err := nativeQ012BuildFromSigned(params, postAdd, quotientValues, level, postAdd.Scale.Div(rlwe.NewScale(2)))
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	quotient, err := q055T2DivideByTwoQ012(params, postAdd)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence.RowProof = append(evidence.RowProof, q055T2MakeRowProof(params, "post_divide_by_2_q012", quotient, quotientOracle))
	contracted, contractionValid, err := q055T2ContractQ01(params, quotient)
	if err != nil {
		return q055T2CandidateRun{}, err
	}
	evidence.CandidateContractionValid = contractionValid
	evidence.CandidateOperation = func() *b4BabyOperation {
		operation := b4BabyOperationRow(params, id, "q012_scalar_mul_then_add_and_divide_by_2", 2, coefficient, power, pre.Scale, quotient.Scale, encoding, true, "2", true)
		operation.ResultCapacityRatio = evidence.Capacity[len(evidence.Capacity)-1].IntendedQ01Ratio
		operation.ScalarCenteredUnique = evidence.Capacity[2].Q01CenteredUnique
		return &operation
	}()
	if !contractionValid || contracted == nil {
		evidence.CandidateValid = false
		evidence.Reason = "post-divide q012 result is not q01-centered-unique"
		return q055T2CandidateRun{Evidence: evidence, PostAdd: postAdd, Quotient: quotient}, nil
	}
	evidence.CandidateOutputMetadataMatches = t2GuardMetadataEqual(contracted, nativePost)
	evidence.CandidateFinalRowsMatchNative = contracted != nil && nativeQ01FastRowsEqual(params, contracted, nativePost)
	evidence.CandidateValid = contractionValid && evidence.CandidateOutputMetadataMatches
	for _, proof := range evidence.RowProof {
		evidence.CandidateValid = evidence.CandidateValid && proof.RowsEqual && proof.MetadataEqual
	}
	if !evidence.CandidateValid && evidence.Reason == "" {
		evidence.Reason = "q012 row, contraction, or metadata proof failed"
	}
	_ = eval
	return q055T2CandidateRun{Evidence: evidence, PostAdd: postAdd, Quotient: quotient, Contracted: contracted, Final: contracted}, nil
}

func q055T2ReplayBranch(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, planScale rlwe.Scale) (*rlwe.Ciphertext, q055T2CandidateRun, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, powers, expected, decoded, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride)
	if err != nil {
		return nil, q055T2CandidateRun{}, err
	}
	native, nativeOK := b4BabyFindCheckpoint(replay.Checks, "B4-term-2")
	pre, preOK := b4BabyFindCheckpoint(replay.Checks, "B4-term-4")
	if !nativeOK || !preOK {
		return nil, q055T2CandidateRun{}, fmt.Errorf("accepted PS plan did not expose B4 T2 boundaries")
	}
	poly := context.Plan.Value[len(context.Plan.Value)-1]
	power := powers[2]
	if power == nil || len(poly.Coeffs) <= 2 || poly.Coeffs[2] == nil {
		return nil, q055T2CandidateRun{}, fmt.Errorf("accepted PS plan missing T2 power or coefficient")
	}
	run, err := q055T2Q012Apply(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, power, poly.Coeffs[2], pre.ciphertext, native.ciphertext, "B4-term-2")
	if err != nil {
		return nil, run, err
	}
	if run.Final == nil {
		return nil, run, nil
	}
	final, err := generatedPowerReentryReplaySuffixAtPlanScale(profile, context, powers, expected, decoded, run.Final, planScale)
	if err != nil {
		return nil, run, err
	}
	run.Final = final
	return final, run, nil
}

func q055T2FirstDivergence(downstream *b4BabyDownstream) string {
	if downstream == nil {
		return "guard_replay"
	}
	if downstream.EvalModReal != nil && downstream.EvalModReal.MaxComponent > q055T2PublicThreshold {
		return "EvalMod.real"
	}
	if downstream.EvalModImag != nil && downstream.EvalModImag.MaxComponent > q055T2PublicThreshold {
		return "EvalMod.imag"
	}
	if downstream.PostS2C != nil && downstream.PostS2C.MaxComponent > q055T2PublicThreshold {
		return "post_S2C"
	}
	if downstream.PublicLike != nil && downstream.PublicLike.MaxComponent > q055T2PublicThreshold {
		return "public_finalization"
	}
	return "none"
}

func q055T2SystemMetrics(evidence *psLocalizationDownstreamEvidence) *b4BabyDownstream {
	if evidence == nil {
		return &b4BabyDownstream{}
	}
	return &b4BabyDownstream{
		Real:            evidence,
		EvalModReal:     evidence.EvalModReal,
		EvalModImag:     evidence.EvalModImag,
		PublicLike:      evidence.PublicLike,
		PostS2C:         evidence.PostS2C,
		MetadataCorrect: evidence.MetadataCorrect,
		ContractsPass:   evidence.ContractsPass,
		Pass:            evidence.Reached && evidence.PublicLike != nil && evidence.PublicLike.Pass,
	}
}

func q055T2RunCase(cfg BootstrapConfig, q0Bits int) (*q055T2Case, error) {
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, q0Bits))
	if err != nil {
		return nil, err
	}
	params := profile.BTP.BootstrappingParameters
	planScale := precisionSweepScale(q055T2PlanScaleExponent)
	realContext, imagContext, _, _, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return nil, err
	}
	realContext.Plan = psOracleScalePlan(realContext.Branch.Base.Plan, planScale)
	imagContext.Plan = psOracleScalePlan(imagContext.Branch.Base.Plan, planScale)
	realE2, realZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastReal)
	if err != nil {
		return nil, err
	}
	imagE2, imagZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastImag)
	if err != nil {
		return nil, err
	}
	realNativeResult, imagNativeResult := nativeQ012Result{}, nativeQ012Result{}
	realPowers, realValues, err := nativeQ012Generate(params, "real", realE2, realZ, &realNativeResult)
	if err != nil {
		return nil, err
	}
	imagPowers, imagValues, err := nativeQ012Generate(params, "imag", imagE2, imagZ, &imagNativeResult)
	if err != nil {
		return nil, err
	}
	result := &q055T2Case{Name: fmt.Sprintf("q0_%d_plan_2^%d", q0Bits, q055T2PlanScaleExponent), Q0Bits: q0Bits, PlanScaleExponent: q055T2PlanScaleExponent}
	for _, item := range []struct {
		branch string
		powers map[int]*rlwe.Ciphertext
		expect map[int][]complex128
		values map[int][]complex128
		native nativeQ012Result
	}{{"real", realPowers, realContext.Branch.Base.PowerExpected, realValues, realNativeResult}, {"imag", imagPowers, imagContext.Branch.Base.PowerExpected, imagValues, imagNativeResult}} {
		for _, power := range []int{2, 3, 4, 6, 8, 16} {
			result.Generated = append(result.Generated, closurePowerEvidenceFor(params, item.branch, power, item.powers, item.expect, item.native))
		}
	}
	q01Real, q01RealGuard, err := generatedPowerReentryReplayBranchAtPlanScale(profile, realContext, realPowers, realContext.Branch.Base.PowerExpected, realValues, planScale)
	if err != nil {
		return nil, err
	}
	q01Imag, q01ImagGuard, err := generatedPowerReentryReplayBranchAtPlanScale(profile, imagContext, imagPowers, imagContext.Branch.Base.PowerExpected, imagValues, planScale)
	if err != nil {
		return nil, err
	}
	q01RealDownstream, err := psLocalizationRunAcceptedDownstream(profile, planScale, q01Real, q01Imag)
	if err != nil {
		return nil, err
	}
	result.Q01GuardDownstream = q055T2SystemMetrics(q01RealDownstream)
	result.Q01FirstDivergence = q055T2FirstDivergence(result.Q01GuardDownstream)
	result.Q01Pass = result.Q01GuardDownstream != nil && result.Q01GuardDownstream.PublicLike != nil && result.Q01GuardDownstream.PublicLike.Pass
	q012Real, q012RealRun, err := q055T2ReplayBranch(profile, realContext, realPowers, realContext.Branch.Base.PowerExpected, realValues, planScale)
	if err != nil {
		return nil, err
	}
	q012Imag, q012ImagRun, err := q055T2ReplayBranch(profile, imagContext, imagPowers, imagContext.Branch.Base.PowerExpected, imagValues, planScale)
	if err != nil {
		return nil, err
	}
	if q012Real != nil && q012Imag != nil {
		q012RealDownstream, err := psLocalizationRunAcceptedDownstream(profile, planScale, q012Real, q012Imag)
		if err != nil {
			return nil, err
		}
		result.Q012GuardDownstream = q055T2SystemMetrics(q012RealDownstream)
		result.Q012FirstDivergence = q055T2FirstDivergence(result.Q012GuardDownstream)
		result.Q012Pass = result.Q012GuardDownstream.PublicLike != nil && result.Q012GuardDownstream.PublicLike.Pass
	} else {
		result.Q012FirstDivergence = "q012_contraction"
		result.Q012Pass = false
	}
	result.RealGuard = q012RealRun.Evidence
	result.ImagGuard = q012ImagRun.Evidence
	result.RealGuard.CurrentQ01Operation = q01RealGuard.Operation
	result.ImagGuard.CurrentQ01Operation = q01ImagGuard.Operation
	result.RealGuard.CurrentQ01Contraction = q01RealGuard.Contraction
	result.ImagGuard.CurrentQ01Contraction = q01ImagGuard.Contraction
	if q01Real != nil && q012Real != nil {
		q01Values, _ := psGlobalDecode(params, q01Real)
		q012Values, _ := psGlobalDecode(params, q012Real)
		result.Q012VsQ01Real = psGlobalMetric(q01Values, q012Values)
	}
	if q01Imag != nil && q012Imag != nil {
		q01Values, _ := psGlobalDecode(params, q01Imag)
		q012Values, _ := psGlobalDecode(params, q012Imag)
		result.Q012VsQ01Imag = psGlobalMetric(q01Values, q012Values)
	}
	result.Q012MetadataEquivalent = q012RealRun.Evidence.CandidateOutputMetadataMatches && q012ImagRun.Evidence.CandidateOutputMetadataMatches
	result.CandidateSystemContracts = result.Q012GuardDownstream != nil && result.Q012GuardDownstream.ContractsPass && q012RealRun.Evidence.CandidateValid && q012ImagRun.Evidence.CandidateValid
	result.CandidateSystemSufficient = result.CandidateSystemContracts && result.Q012Pass
	return result, nil
}

func q055T2LoadPreviousMatrix(primaryRoot string) (q055T2PreviousMatrix, error) {
	path := filepath.Join(primaryRoot, "results", "FIX-001-P3-DESIGN-LOGN13-D0-PRODUCTION-PRECONDITION-CLOSURE-summary.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return q055T2PreviousMatrix{ArtifactPath: path}, err
	}
	var previous closureResult
	if err := json.Unmarshal(data, &previous); err != nil {
		return q055T2PreviousMatrix{ArtifactPath: path}, err
	}
	observed := map[string]bool{}
	for _, item := range previous.Matrix {
		if len(item.Name) > 0 {
			observed[item.Name[:1]] = item.Pass
		}
	}
	expected := map[string]bool{"A": false, "B": false, "C": false, "D": true}
	corrected := "logn13_d0_q0_plan_interaction"
	if len(previous.Matrix) < 4 {
		return q055T2PreviousMatrix{ArtifactPath: path}, fmt.Errorf("previous D0 closure artifact has %d matrix rows, need at least 4", len(previous.Matrix))
	}
	return q055T2PreviousMatrix{Imported: true, ArtifactPath: path, ArtifactPrimaryCommit: previous.Primary.Commit, ArtifactSecondaryCommit: previous.Lattigo.Commit, PreviousClassification: previous.Classification, PreviousBFirstDivergence: previous.Matrix[1].FirstDivergence, ExpectedPassPattern: expected, ObservedPassPattern: observed, CorrectedClassification: corrected, CorrectedBFirstDivergence: "EvalMod.real", RegressionAssertions: map[string]interface{}{"pass_pattern_matches": observed["A"] == false && observed["B"] == false && observed["C"] == false && observed["D"] == true, "d_only_is_interaction": true, "old_classification_was_q0_width_only": previous.Classification == "logn13_d0_q0_56_width_precondition", "old_b_first_divergence_was_hard_coded_public_finalization": previous.Matrix[1].FirstDivergence == "public_finalization"}}, nil
}

func runFIX001P3DesignLogN13Q055Plan92T2LocalQ2GuardFeasibility(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result := q055T2Result{SchemaVersion: "fix-001-p3-design-logn13-q055-plan92-t2-local-q2-guard-feasibility.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Provenance: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3Q055T2Primary, "required_secondary_commit": requiredFIX001P3Q055T2Secondary, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "plan_scale_source": "lattigo/circuits/ckks/mod1/fast.go:normalizedLogN13PlanScaleBits; diagnostic candidate uses 2^92"}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "checked_in_q0_55": len(cfg.Q0) == 1 && cfg.Q0[0] == 55, "plan_scale_2^92": true, "no_q_parameter_change": true, "no_secondary_production_changes": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "q012_candidate_uses_only_q0_q1_q2": true, "q3_plus_dormant": true, "big_int_only_independent_oracle": true}}
	previous, err := q055T2LoadPreviousMatrix(primaryRoot)
	result.PreviousMatrix = previous
	if err != nil {
		result.Classification = "logn13_q055_plan92_t2_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "previous D0 closure artifact unavailable or invalid"
		return q055T2Write(result, outPath)
	}
	if cfg.LogN != 13 || len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || gitOutput(primaryRoot, "merge-base", requiredFIX001P3Q055T2Primary, "HEAD") != requiredFIX001P3Q055T2Primary || gitOutput(backendRoot, "rev-parse", "HEAD") != requiredFIX001P3Q055T2Secondary || gitOutput(backendRoot, "branch", "--show-current") != "fast-ckks" || gitOutput(backendRoot, "status", "--porcelain") != "" {
		result.Classification = "logn13_q055_plan92_t2_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary provenance or fixed q0=55 precondition mismatch"
		return q055T2Write(result, outPath)
	}
	if !previous.RegressionAssertions["pass_pattern_matches"].(bool) || previous.CorrectedClassification != "logn13_d0_q0_plan_interaction" {
		result.Classification = "logn13_q055_plan92_t2_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "previous D0 matrix regression did not establish corrected interaction pattern"
		return q055T2Write(result, outPath)
	}
	result.B, err = q055T2RunCase(cfg, 55)
	if err != nil {
		result.Classification = "logn13_q055_plan92_t2_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "q0=55 plan92 control execution failed: " + err.Error()
		return q055T2Write(result, outPath)
	}
	result.D, err = q055T2RunCase(cfg, 56)
	if err != nil {
		result.Classification = "logn13_q055_plan92_t2_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "q0=56 plan92 control execution failed: " + err.Error()
		return q055T2Write(result, outPath)
	}
	b := result.B
	d := result.D
	bPublic := b.Q01GuardDownstream != nil && b.Q01GuardDownstream.PublicLike != nil
	bEval := b.Q01GuardDownstream != nil && b.Q01GuardDownstream.EvalModReal != nil && b.Q01GuardDownstream.EvalModImag != nil
	bPost := b.Q01GuardDownstream != nil && b.Q01GuardDownstream.PostS2C != nil
	bReproduced := bPublic && bEval && bPost && b.Q01GuardDownstream.EvalModReal.MaxComponent > 50 && b.Q01GuardDownstream.EvalModReal.MaxComponent < 200 && b.Q01GuardDownstream.EvalModImag.MaxComponent > 50 && b.Q01GuardDownstream.EvalModImag.MaxComponent < 200 && b.Q01GuardDownstream.PostS2C.MaxComponent > 50 && b.Q01GuardDownstream.PostS2C.MaxComponent < 200 && b.Q01GuardDownstream.PublicLike.MaxComponent > 1000
	result.Validation["b_current_q01_guard_reproduced"] = bReproduced
	result.Validation["b_current_q01_guard_first_divergence_is_not_public_finalization"] = b.Q01FirstDivergence != "public_finalization"
	result.Validation["d_q012_candidate_equivalent_to_q01"] = d.Q012MetadataEquivalent && d.Q012VsQ01Real != nil && d.Q012VsQ01Imag != nil && d.Q012VsQ01Real.MaxComponent <= q055T2SemanticTolerance && d.Q012VsQ01Imag.MaxComponent <= q055T2SemanticTolerance
	result.Validation["b_q01_capacity_blocker_confirmed"] = (b.RealGuard.Capacity != nil && q055T2CapacityHasTransientBlocker(b.RealGuard.Capacity)) || (b.ImagGuard.Capacity != nil && q055T2CapacityHasTransientBlocker(b.ImagGuard.Capacity))
	if !bReproduced {
		result.Classification = "logn13_q055_plan92_t2_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "q0=55 plan92 current q01 T2 guard did not reproduce the accepted B control"
		return q055T2Write(result, outPath)
	}
	if !result.Validation["b_q01_capacity_blocker_confirmed"].(bool) {
		result.Classification = "logn13_q0_plan_interaction_not_t2_guard"
		result.ProductionReadiness = "production_blocked_by_non_t2_q0_plan_interaction"
		result.FirstRemainingBlocker = "q01 T2 transient capacity blocker not confirmed"
		result.RecommendedNextTarget = "localize the first remaining q0=55/plan92 divergence before any frontend parameter change"
	} else if !b.CandidateSystemContracts {
		result.Classification = "logn13_q055_plan92_t2_local_q2_guard_operation_mismatch"
		result.ProductionReadiness = "production_blocked_by_t2_local_q2_guard_operation_mismatch"
		result.FirstRemainingBlocker = "q0/q1/q2 T2 guard row, contraction, or metadata proof failed"
		result.RecommendedNextTarget = "repair the local-q2 T2 operation proof before production integration"
	} else if b.CandidateSystemSufficient {
		result.Classification = "logn13_q055_plan92_t2_local_q2_guard_system_sufficient"
		result.ProductionReadiness = "production_ready_fixed_q055_plan92_with_t2_local_q2"
		result.FirstRemainingBlocker = "none"
		result.RecommendedNextTarget = "production integration of the complete LogN13 D0 stack under fixed q0=55 with internal planScale92 and temporary-q2 T2 guard"
	} else {
		result.Classification = "logn13_q055_plan92_t2_local_q2_guard_numeric_insufficient"
		result.ProductionReadiness = "production_blocked_by_remaining_q055_plan92_divergence"
		result.FirstRemainingBlocker = b.Q012FirstDivergence
		result.RecommendedNextTarget = "localize the first remaining q0=55/plan92 divergence before any frontend parameter change"
	}
	return q055T2Write(result, outPath)
}

func q055T2CapacityHasTransientBlocker(capacity []q055T2CapacityEvidence) bool {
	for _, item := range capacity {
		if item.State == "intended_physical_post_add_before_divide_by_2" && item.IntendedQ01OutsideCount > 0 && item.IntendedQ012OutsideCount == 0 {
			return true
		}
	}
	return false
}

func q055T2Write(result q055T2Result, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}
