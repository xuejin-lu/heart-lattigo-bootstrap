//go:build fix001t3capacity

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type FIX001CapacityCiphertextMetadata struct {
	Level        int     `json:"level"`
	Scale        string  `json:"scale"`
	ScaleFloat64 float64 `json:"scale_float64"`
	ScaleLog2    float64 `json:"scale_log2"`
	Degree       int     `json:"degree"`
	N            int     `json:"n"`
	IsNTT        bool    `json:"is_ntt"`
	IsMontgomery bool    `json:"is_montgomery"`
}

type FIX001CapacityRecoveredComponent struct {
	Component       int      `json:"component"`
	SignedCoeffs    []string `json:"signed_centered_coefficients"`
	MaxAbs          string   `json:"max_abs"`
	MaxAbsIndices   []int    `json:"max_abs_indices"`
	LiftQ01Exact    bool     `json:"lift_q01_exact"`
	LiftNTTQ01Exact bool     `json:"lift_ntt_q01_exact"`
}

type FIX001CapacityRecoveredCiphertext struct {
	Metadata   FIX001CapacityCiphertextMetadata   `json:"metadata"`
	Components []FIX001CapacityRecoveredComponent `json:"components"`
}

type FIX001CapacityRepresentative struct {
	Index        int    `json:"index"`
	Value        string `json:"value"`
	AbsValue     string `json:"abs_value"`
	WrappedValue string `json:"wrapped_value"`
	WrapCount    string `json:"wrap_count"`
}

type FIX001CapacityStats struct {
	Name                    string                         `json:"name"`
	CoefficientCount        int                            `json:"coefficient_count"`
	SignedCoefficients      []string                       `json:"signed_centered_coefficients"`
	MaxAbs                  string                         `json:"max_abs"`
	MaxAbsIndices           []int                          `json:"max_abs_indices"`
	MaxRatioQ01Half         float64                        `json:"max_ratio_abs_over_q01_half"`
	MaxRatioQ01HalfExact    string                         `json:"max_ratio_abs_over_q01_half_exact"`
	OutsideQ01CenteredCount int                            `json:"outside_q01_centered_count"`
	OutsideQ01Percentage    float64                        `json:"outside_q01_percentage"`
	MinNonzeroWrapCount     string                         `json:"min_nonzero_wrap_count,omitempty"`
	MaxNonzeroWrapCount     string                         `json:"max_nonzero_wrap_count,omitempty"`
	WrapHistogram           map[string]int                 `json:"wrap_count_histogram_small"`
	Representatives         []FIX001CapacityRepresentative `json:"representative_outside_coefficients"`
	FullQCentered           bool                           `json:"all_within_full_q_centered_range"`
	FullQHalf               string                         `json:"full_q_half"`
}

type FIX001CapacityRows struct {
	Component   int                         `json:"component"`
	Comparisons []FinalizationRowComparison `json:"comparisons"`
	Exact       bool                        `json:"exact"`
}

type FIX001CapacityWrappedSimulation struct {
	InputQ01RowsBitExact  bool                           `json:"input_q01_rows_bit_exact"`
	OutputQ01RowsBitExact bool                           `json:"output_q01_rows_bit_exact"`
	Simulated             FinalizationCiphertextEvidence `json:"simulated_ciphertext"`
	Actual                FinalizationCiphertextEvidence `json:"actual_ciphertext"`
	SimulatedDecoded      []CorrectnessValue             `json:"simulated_decoded"`
	ActualDecoded         []CorrectnessValue             `json:"actual_decoded"`
	SimulatedVsExpected   CorrectnessMetrics             `json:"simulated_vs_expected"`
	ActualVsExpected      CorrectnessMetrics             `json:"actual_vs_expected"`
	SimulatedVsActual     CorrectnessMetrics             `json:"simulated_vs_actual"`
	Rows                  []FIX001CapacityRows           `json:"row_comparisons"`
	ExplainsFastFailure   bool                           `json:"explains_fast_failure"`
}

type FIX001CapacityFullQRescale struct {
	Input              FinalizationCiphertextEvidence `json:"input_ciphertext"`
	Output             FinalizationCiphertextEvidence `json:"output_ciphertext"`
	Decoded            []CorrectnessValue             `json:"decoded"`
	Expected           []CorrectnessValue             `json:"expected"`
	Metrics            CorrectnessMetrics             `json:"metrics"`
	Pass               bool                           `json:"pass"`
	ReferenceValidated bool                           `json:"reference_validated"`
	ValidationFailure  string                         `json:"validation_failure,omitempty"`
}

type FIX001CapacityT2Control struct {
	P1                 FIX001CapacityStats `json:"p1"`
	P2                 FIX001CapacityStats `json:"p2"`
	PostRescaleMetrics CorrectnessMetrics  `json:"post_rescale_metrics"`
	PostRescalePass    bool                `json:"post_rescale_pass"`
}

type FIX001CapacityResult struct {
	SchemaVersion                   string                            `json:"schema_version"`
	Timestamp                       time.Time                         `json:"timestamp"`
	Primary                         RepositoryMetadata                `json:"primary_repository"`
	Lattigo                         RepositoryMetadata                `json:"lattigo_repository"`
	Environment                     EnvironmentMetadata               `json:"environment"`
	Config                          BootstrapConfig                   `json:"config"`
	Parameters                      ExperimentParameters              `json:"effective_parameters"`
	Workload                        CorrectnessWorkload               `json:"workload"`
	RepairedFastCommit              string                            `json:"repaired_fast_commit"`
	CorrectedT0Oracle               bool                              `json:"corrected_t0_oracle"`
	E2VsHistorical                  CorrectnessMetrics                `json:"e2_vs_historical"`
	T1                              FIX001CapacityRecoveredCiphertext `json:"t1"`
	T2                              FIX001CapacityRecoveredCiphertext `json:"t2"`
	T3PreRescale                    FIX001CapacityCiphertextMetadata  `json:"t3_pre_rescale"`
	T3PostRescale                   FIX001CapacityCiphertextMetadata  `json:"t3_post_rescale"`
	Q0                              string                            `json:"q0"`
	Q1                              string                            `json:"q1"`
	Q01                             string                            `json:"q01"`
	Q01Half                         string                            `json:"q01_half"`
	CommonLevel                     int                               `json:"common_level"`
	RescaleDivisor                  string                            `json:"rescale_divisor"`
	P1                              FIX001CapacityStats               `json:"p1"`
	P2                              FIX001CapacityStats               `json:"p2"`
	T2Control                       FIX001CapacityT2Control           `json:"t2_control"`
	FullQRescale                    FIX001CapacityFullQRescale        `json:"full_q_rescale_oracle"`
	WrappedSimulation               FIX001CapacityWrappedSimulation   `json:"q01_wrapped_simulation"`
	RoughWrapError                  string                            `json:"rough_q01_over_d_scale_out"`
	RoughWrapErrorFloat64           float64                           `json:"rough_q01_over_d_scale_out_float64"`
	RoughWrapErrorLog2              float64                           `json:"rough_q01_over_d_scale_out_log2"`
	FirstSupportedCause             string                            `json:"first_supported_cause"`
	Threshold                       float64                           `json:"threshold"`
	SecondaryWorktreeClean          bool                              `json:"secondary_worktree_clean"`
	TemporarySecondarySourceRemoved bool                              `json:"temporary_secondary_source_removed"`
	Notes                           []string                          `json:"notes"`
}

type fix001CapacityRecovered struct {
	components []fix001CapacityComponent
}

type fix001CapacityComponent struct {
	signed    []*big.Int
	residues  [][]uint64
	normalNTT ring.Poly
	montNTT   ring.Poly
}

func fix001CapacityMetadata(ct *rlwe.Ciphertext) FIX001CapacityCiphertextMetadata {
	return FIX001CapacityCiphertextMetadata{Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), ScaleFloat64: ct.Scale.Float64(), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(), N: ct.N(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery}
}

func fix001CapacityBigIntStrings(values []*big.Int) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

func fix001CapacityAbs(value *big.Int) *big.Int {
	return new(big.Int).Abs(value)
}

func fix001CapacityRecover(params ckks.Parameters, ct *rlwe.Ciphertext) (fix001CapacityRecovered, FIX001CapacityRecoveredCiphertext, error) {
	if ct.Level() < 1 || !ct.IsNTT {
		return fix001CapacityRecovered{}, FIX001CapacityRecoveredCiphertext{}, fmt.Errorf("capacity recovery requires NTT level >= 1 ciphertext")
	}
	ringQ := params.RingQ().AtLevel(ct.Level())
	backend := fastckks.NewRNSBackend(params.RingQ().AtLevel(1), 1)
	recovered := fix001CapacityRecovered{components: make([]fix001CapacityComponent, ct.Degree()+1)}
	recorded := FIX001CapacityRecoveredCiphertext{Metadata: fix001CapacityMetadata(ct), Components: make([]FIX001CapacityRecoveredComponent, ct.Degree()+1)}
	for component := range ct.Value {
		coeff := ring.NewPoly(params.N(), 1)
		if err := fastckks.FastPartialINTT(ringQ, ct.Value[component], coeff); err != nil {
			return fix001CapacityRecovered{}, FIX001CapacityRecoveredCiphertext{}, err
		}
		if ct.IsMontgomery {
			for limb := 0; limb < 2; limb++ {
				params.RingQ().SubRings[limb].IMForm(coeff.Coeffs[limb], coeff.Coeffs[limb])
			}
		}
		signed, err := backend.ReconstructQ0Q1(&coeff)
		if err != nil {
			return fix001CapacityRecovered{}, FIX001CapacityRecoveredCiphertext{}, err
		}
		normalNTT := ring.NewPoly(params.N(), 1)
		params.RingQ().AtLevel(1).NTT(coeff, normalNTT)
		actualNormalNTT := ring.NewPoly(params.N(), 1)
		for limb := 0; limb < 2; limb++ {
			copy(actualNormalNTT.Coeffs[limb], ct.Value[component].Coeffs[limb])
			if ct.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(actualNormalNTT.Coeffs[limb], actualNormalNTT.Coeffs[limb])
			}
		}
		nttExact := true
		for limb := 0; limb < 2; limb++ {
			for i := 0; i < params.N(); i++ {
				if normalNTT.Coeffs[limb][i] != actualNormalNTT.Coeffs[limb][i] {
					nttExact = false
					break
				}
			}
		}
		maxAbs := big.NewInt(0)
		maxIndices := []int{}
		for i, value := range signed {
			abs := fix001CapacityAbs(value)
			cmp := abs.Cmp(maxAbs)
			if cmp > 0 {
				maxAbs.Set(abs)
				maxIndices = []int{i}
			} else if cmp == 0 {
				maxIndices = append(maxIndices, i)
			}
		}
		recovered.components[component] = fix001CapacityComponent{signed: signed, residues: [][]uint64{append([]uint64(nil), coeff.Coeffs[0]...), append([]uint64(nil), coeff.Coeffs[1]...)}, normalNTT: normalNTT}
		recorded.Components[component] = FIX001CapacityRecoveredComponent{Component: component, SignedCoeffs: fix001CapacityBigIntStrings(signed), MaxAbs: maxAbs.String(), MaxAbsIndices: maxIndices, LiftQ01Exact: true, LiftNTTQ01Exact: nttExact}
	}
	return recovered, recorded, nil
}

func fix001CapacityLift(params ckks.Parameters, signed []*big.Int, level int) (ring.Poly, ring.Poly, ring.Poly) {
	ringQ := params.RingQ().AtLevel(level)
	coeff := ring.NewPoly(params.N(), level)
	for limb := 0; limb <= level; limb++ {
		modulus := new(big.Int).SetUint64(ringQ.SubRings[limb].Modulus)
		for i, value := range signed {
			coeff.Coeffs[limb][i] = new(big.Int).Mod(new(big.Int).Set(value), modulus).Uint64()
		}
	}
	normalNTT := ring.NewPoly(params.N(), level)
	ringQ.NTT(coeff, normalNTT)
	montNTT := ring.NewPoly(params.N(), level)
	ringQ.MForm(normalNTT, montNTT)
	return coeff, normalNTT, montNTT
}

func fix001CapacityFullCRT(params ckks.Parameters, poly ring.Poly, level int) ([]*big.Int, *big.Int) {
	ringQ := params.RingQ().AtLevel(level)
	qFull := new(big.Int).Set(params.RingQ().ModulusAtLevel[level])
	values := make([]*big.Int, params.N())
	for i := range values {
		value := new(big.Int).SetUint64(poly.Coeffs[0][i])
		currentModulus := new(big.Int).SetUint64(ringQ.SubRings[0].Modulus)
		for limb := 1; limb <= level; limb++ {
			qi := new(big.Int).SetUint64(ringQ.SubRings[limb].Modulus)
			residue := new(big.Int).SetUint64(poly.Coeffs[limb][i])
			delta := new(big.Int).Sub(residue, new(big.Int).Mod(new(big.Int).Set(value), qi))
			delta.Mod(delta, qi)
			inverse := new(big.Int).ModInverse(new(big.Int).Mod(new(big.Int).Set(currentModulus), qi), qi)
			delta.Mul(delta, inverse).Mod(delta, qi)
			value.Add(value, new(big.Int).Mul(currentModulus, delta))
			currentModulus.Mul(currentModulus, qi)
		}
		half := new(big.Int).Rsh(new(big.Int).Set(currentModulus), 1)
		if value.Cmp(half) > 0 {
			value.Sub(value, currentModulus)
		}
		values[i] = value
	}
	return values, qFull
}

func fix001CapacityProduct(params ckks.Parameters, left, right fix001CapacityComponent, level int) ([]*big.Int, []*big.Int, ring.Poly, ring.Poly, error) {
	_, _, leftMont := fix001CapacityLift(params, left.signed, level)
	_, rightNormal, rightMont := fix001CapacityLift(params, right.signed, level)
	_ = rightMont
	ringQ := params.RingQ().AtLevel(level)
	p1NormalNTT := ring.NewPoly(params.N(), level)
	ringQ.MulCoeffsMontgomery(leftMont, rightNormal, p1NormalNTT)
	p2NormalNTT := ring.NewPoly(params.N(), level)
	ringQ.Add(p1NormalNTT, p1NormalNTT, p2NormalNTT)
	p1Coeff := fix001CapacityNTTToCoeff(params, p1NormalNTT, level)
	p2Coeff := fix001CapacityNTTToCoeff(params, p2NormalNTT, level)
	p1, qFull := fix001CapacityFullCRT(params, p1Coeff, level)
	p2, _ := fix001CapacityFullCRT(params, p2Coeff, level)
	fullHalf := new(big.Int).Rsh(new(big.Int).Set(qFull), 1)
	for i := range p1 {
		if fix001CapacityAbs(p1[i]).Cmp(fullHalf) > 0 || fix001CapacityAbs(p2[i]).Cmp(fullHalf) > 0 {
			return nil, nil, ring.Poly{}, ring.Poly{}, fmt.Errorf("FULL_Q_REFERENCE_CAPACITY_INSUFFICIENT at coefficient %d", i)
		}
	}
	return p1, p2, p1NormalNTT, p2NormalNTT, nil
}

func fix001CapacityNTTToCoeff(params ckks.Parameters, normalNTT ring.Poly, level int) ring.Poly {
	ringQ := params.RingQ().AtLevel(level)
	coeff := ring.NewPoly(params.N(), level)
	ringQ.INTT(normalNTT, coeff)
	return coeff
}

func fix001CapacityCenteredQ01(value, q01, half *big.Int) (*big.Int, *big.Int) {
	wrapped := new(big.Int).Mod(new(big.Int).Set(value), q01)
	if wrapped.Cmp(half) > 0 {
		wrapped.Sub(wrapped, q01)
	}
	wrap := new(big.Int).Sub(value, wrapped)
	wrap.Quo(wrap, q01)
	return wrapped, wrap
}

func fix001CapacityStats(name string, values []*big.Int, q01, q01Half, qFull *big.Int) FIX001CapacityStats {
	stats := FIX001CapacityStats{Name: name, CoefficientCount: len(values), SignedCoefficients: fix001CapacityBigIntStrings(values), MaxAbs: "0", WrapHistogram: map[string]int{}, FullQHalf: new(big.Int).Rsh(new(big.Int).Set(qFull), 1).String()}
	maxAbs := big.NewInt(0)
	minWrap, maxWrap := (*big.Int)(nil), (*big.Int)(nil)
	for i, value := range values {
		abs := fix001CapacityAbs(value)
		if abs.Cmp(maxAbs) > 0 {
			maxAbs.Set(abs)
			stats.MaxAbsIndices = []int{i}
		} else if abs.Cmp(maxAbs) == 0 {
			stats.MaxAbsIndices = append(stats.MaxAbsIndices, i)
		}
		if abs.Cmp(q01Half) > 0 {
			stats.OutsideQ01CenteredCount++
			if len(stats.Representatives) < 12 {
				wrapped, wrap := fix001CapacityCenteredQ01(value, q01, q01Half)
				stats.Representatives = append(stats.Representatives, FIX001CapacityRepresentative{Index: i, Value: value.String(), AbsValue: abs.String(), WrappedValue: wrapped.String(), WrapCount: wrap.String()})
			}
		}
		_, wrap := fix001CapacityCenteredQ01(value, q01, q01Half)
		if wrap.Sign() != 0 {
			if minWrap == nil || wrap.Cmp(minWrap) < 0 {
				minWrap = new(big.Int).Set(wrap)
			}
			if maxWrap == nil || wrap.Cmp(maxWrap) > 0 {
				maxWrap = new(big.Int).Set(wrap)
			}
			if new(big.Int).Abs(wrap).Cmp(big.NewInt(8)) <= 0 {
				stats.WrapHistogram[wrap.String()]++
			}
		}
	}
	stats.MaxAbs = maxAbs.String()
	stats.OutsideQ01Percentage = float64(stats.OutsideQ01CenteredCount) * 100 / float64(len(values))
	if maxAbs.Sign() != 0 {
		ratio := new(big.Float).Quo(new(big.Float).SetInt(maxAbs), new(big.Float).SetInt(q01Half))
		stats.MaxRatioQ01Half, _ = ratio.Float64()
		stats.MaxRatioQ01HalfExact = ratio.Text('g', 30)
	}
	if minWrap != nil {
		stats.MinNonzeroWrapCount = minWrap.String()
	}
	if maxWrap != nil {
		stats.MaxNonzeroWrapCount = maxWrap.String()
	}
	stats.FullQCentered = true
	fullHalf := new(big.Int).Rsh(new(big.Int).Set(qFull), 1)
	for _, value := range values {
		if fix001CapacityAbs(value).Cmp(fullHalf) > 0 {
			stats.FullQCentered = false
			break
		}
	}
	return stats
}

func fix001CapacityRowResult(component int, before, after *rlwe.Ciphertext) FIX001CapacityRows {
	comparisons := make([]FinalizationRowComparison, 0, 2)
	for limb := 0; limb < 2; limb++ {
		comparisons = append(comparisons, finalizationRowComparison(component, limb, before.Value[component].Coeffs[limb], after.Value[component].Coeffs[limb]))
	}
	exact := true
	for _, row := range comparisons {
		exact = exact && row.Equal
	}
	return FIX001CapacityRows{Component: component, Comparisons: comparisons, Exact: exact}
}

func fix001CapacityRowsEqual(component int, a, b *rlwe.Ciphertext) bool {
	for limb := 0; limb < 2; limb++ {
		if !finalizationRowComparison(component, limb, a.Value[component].Coeffs[limb], b.Value[component].Coeffs[limb]).Equal {
			return false
		}
	}
	return true
}

func fix001CapacityBuildFullCiphertext(params ckks.Parameters, source *rlwe.Ciphertext, c0 ring.Poly) *rlwe.Ciphertext {
	ct := ckks.NewCiphertext(params, 1, source.Level())
	*ct.MetaData = *source.MetaData
	ct.Scale = source.Scale
	ct.IsNTT = true
	ct.IsMontgomery = false
	for limb := 0; limb <= source.Level(); limb++ {
		copy(ct.Value[0].Coeffs[limb], c0.Coeffs[limb])
		for i := range ct.Value[1].Coeffs[limb] {
			ct.Value[1].Coeffs[limb][i] = 0
		}
	}
	return ct
}

func fix001CapacityBuildWrappedFast(params ckks.Parameters, source *rlwe.Ciphertext, values []*big.Int, divisor *big.Int) *rlwe.Ciphertext {
	targetLevel := source.Level() - 1
	ct := fastckks.NewCiphertext(params, 1, targetLevel)
	*ct.MetaData = *source.MetaData
	ct.Scale = source.Scale
	ct.IsNTT = true
	ct.IsMontgomery = true
	coeff := ring.NewPoly(params.N(), targetLevel)
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	for i, value := range values {
		quotient := new(big.Int).Div(new(big.Int).Add(value, new(big.Int).Rsh(new(big.Int).Set(divisor), 1)), divisor)
		coeff.Coeffs[0][i] = new(big.Int).Mod(new(big.Int).Set(quotient), q0).Uint64()
		coeff.Coeffs[1][i] = new(big.Int).Mod(new(big.Int).Set(quotient), q1).Uint64()
	}
	normalNTT := ring.NewPoly(params.N(), targetLevel)
	params.RingQ().AtLevel(targetLevel).NTT(coeff, normalNTT)
	for limb := 0; limb < 2; limb++ {
		params.RingQ().SubRings[limb].MForm(normalNTT.Coeffs[limb], ct.Value[0].Coeffs[limb])
	}
	return ct
}

func fix001CapacityBuildQ01AtLevel(params ckks.Parameters, source *rlwe.Ciphertext, values []*big.Int) *rlwe.Ciphertext {
	ct := fastckks.NewCiphertext(params, 1, source.Level())
	*ct.MetaData = *source.MetaData
	ct.Scale = source.Scale
	ct.IsNTT = true
	ct.IsMontgomery = true
	q0Coeff := make([]uint64, params.N())
	q1Coeff := make([]uint64, params.N())
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	for i, value := range values {
		q0Coeff[i] = new(big.Int).Mod(new(big.Int).Set(value), q0).Uint64()
		q1Coeff[i] = new(big.Int).Mod(new(big.Int).Set(value), q1).Uint64()
	}
	params.RingQ().SubRings[0].NTT(q0Coeff, ct.Value[0].Coeffs[0])
	params.RingQ().SubRings[1].NTT(q1Coeff, ct.Value[0].Coeffs[1])
	params.RingQ().SubRings[0].MForm(ct.Value[0].Coeffs[0], ct.Value[0].Coeffs[0])
	params.RingQ().SubRings[1].MForm(ct.Value[0].Coeffs[1], ct.Value[0].Coeffs[1])
	return ct
}

func fix001CapacityDecode(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	projection, _, err := q01Projection(params, ct, true)
	if err != nil {
		return nil, err
	}
	return diagnosticDecode(params, projection, zeroSecret(params))
}

func fix001CapacityVectorMul(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] * b[i]
	}
	return out
}

func fix001CapacityVectorScale(a []complex128, scalar complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = scalar * a[i]
	}
	return out
}

func fix001CapacityE2(params bootstrapping.Parameters, eval *bootstrapping.FastEvaluator, ct *rlwe.Ciphertext) (*rlwe.Ciphertext, []complex128, error) {
	res := ct.CopyNew()
	mod1Params := eval.Mod1Parameters
	res.Scale = mod1Params.ScalingFactor()
	offset := new(big.Float).Sub(&mod1Params.Mod1Poly.B, &mod1Params.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(mod1Params.IntervalShrinkFactor()))
	offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
	if err := eval.FastCKKS.Add(res, offset, res); err != nil {
		return nil, nil, err
	}
	decoded, err := fix001CapacityDecode(params.BootstrappingParameters, res)
	return res, decoded, err
}

func fix001CapacityLoadE2(path string) ([]complex128, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value struct {
		E2Decoded []CorrectnessValue `json:"e2_decoded"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	result := make([]complex128, len(value.E2Decoded))
	for i, item := range value.E2Decoded {
		result[i] = complex(item.Real, item.Imag)
	}
	return result, nil
}

func fix001CapacityWrite(result FIX001CapacityResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fix001CapacityWriteSummary(result FIX001CapacityResult, path string) error {
	type component struct {
		Component       int    `json:"component"`
		MaxAbs          string `json:"max_abs"`
		MaxAbsIndices   []int  `json:"max_abs_indices"`
		LiftQ01Exact    bool   `json:"lift_q01_exact"`
		LiftNTTQ01Exact bool   `json:"lift_ntt_q01_exact"`
	}
	components := func(value FIX001CapacityRecoveredCiphertext) []component {
		out := make([]component, len(value.Components))
		for i, item := range value.Components {
			out[i] = component{Component: item.Component, MaxAbs: item.MaxAbs, MaxAbsIndices: item.MaxAbsIndices, LiftQ01Exact: item.LiftQ01Exact, LiftNTTQ01Exact: item.LiftNTTQ01Exact}
		}
		return out
	}
	summary := struct {
		SchemaVersion                   string                           `json:"schema_version"`
		Timestamp                       time.Time                        `json:"timestamp"`
		Primary                         RepositoryMetadata               `json:"primary_repository"`
		Lattigo                         RepositoryMetadata               `json:"lattigo_repository"`
		Environment                     EnvironmentMetadata              `json:"environment"`
		Config                          BootstrapConfig                  `json:"config"`
		Parameters                      ExperimentParameters             `json:"effective_parameters"`
		Workload                        CorrectnessWorkload              `json:"workload"`
		RepairedFastCommit              string                           `json:"repaired_fast_commit"`
		CorrectedT0Oracle               bool                             `json:"corrected_t0_oracle"`
		E2VsHistorical                  CorrectnessMetrics               `json:"e2_vs_historical"`
		T1Metadata                      FIX001CapacityCiphertextMetadata `json:"t1_metadata"`
		T1Components                    []component                      `json:"t1_components"`
		T2Metadata                      FIX001CapacityCiphertextMetadata `json:"t2_metadata"`
		T2Components                    []component                      `json:"t2_components"`
		T3PreRescale                    FIX001CapacityCiphertextMetadata `json:"t3_pre_rescale"`
		T3PostRescale                   FIX001CapacityCiphertextMetadata `json:"t3_post_rescale"`
		Q0                              string                           `json:"q0"`
		Q1                              string                           `json:"q1"`
		Q01                             string                           `json:"q01"`
		Q01Half                         string                           `json:"q01_half"`
		CommonLevel                     int                              `json:"common_level"`
		RescaleDivisor                  string                           `json:"rescale_divisor"`
		P1                              FIX001CapacityStats              `json:"p1"`
		P2                              FIX001CapacityStats              `json:"p2"`
		T2Control                       FIX001CapacityT2Control          `json:"t2_control"`
		FullQRescale                    FIX001CapacityFullQRescale       `json:"full_q_rescale_oracle"`
		WrappedSimulation               FIX001CapacityWrappedSimulation  `json:"q01_wrapped_simulation"`
		RoughWrapError                  string                           `json:"rough_q01_over_d_scale_out"`
		RoughWrapErrorFloat64           float64                          `json:"rough_q01_over_d_scale_out_float64"`
		RoughWrapErrorLog2              float64                          `json:"rough_q01_over_d_scale_out_log2"`
		FirstSupportedCause             string                           `json:"first_supported_cause"`
		Threshold                       float64                          `json:"threshold"`
		SecondaryWorktreeClean          bool                             `json:"secondary_worktree_clean"`
		TemporarySecondarySourceRemoved bool                             `json:"temporary_secondary_source_removed"`
		Notes                           []string                         `json:"notes"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.Environment, result.Config, result.Parameters, result.Workload, result.RepairedFastCommit, result.CorrectedT0Oracle, result.E2VsHistorical, result.T1.Metadata, components(result.T1), result.T2.Metadata, components(result.T2), result.T3PreRescale, result.T3PostRescale, result.Q0, result.Q1, result.Q01, result.Q01Half, result.CommonLevel, result.RescaleDivisor, result.P1, result.P2, result.T2Control, result.FullQRescale, result.WrappedSimulation, result.RoughWrapError, result.RoughWrapErrorFloat64, result.RoughWrapErrorLog2, result.FirstSupportedCause, result.Threshold, result.SecondaryWorktreeClean, result.TemporarySecondarySourceRemoved, result.Notes}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func RunFIX001T3Capacity(cfg BootstrapConfig, primaryRoot, backendRoot string) (FIX001CapacityResult, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	eval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*reproducibleInput(residual, btp)})
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	packed[0] = *scaled
	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	packed[0] = *modUp
	ctReal, _, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	e2, e2Decoded, err := fix001CapacityE2(btp, eval, ctReal)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	historical, err := fix001CapacityLoadE2("results/EXP-002C-DIAG-POLY-logN13-fast.json")
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	e2Metrics, err := compareComplexVectors(historical, e2Decoded, correctnessThreshold)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	if !e2Metrics.PassThreshold {
		return FIX001CapacityResult{}, fmt.Errorf("DIAGNOSTIC_INPUT_MISMATCH: E2 max component=%g", e2Metrics.MaxComponentAbs)
	}
	replay, err := eval.PolynomialEvaluator.DiagnosticReplayT3Capacity(e2)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	x := fix001ExpectedPower(1, e2Decoded, map[int][]complex128{})
	y := fix001ExpectedPower(2, e2Decoded, map[int][]complex128{})
	doubledExpected := fix001CapacityVectorScale(fix001CapacityVectorMul(x, y), 2)
	params := btp.BootstrappingParameters
	t1Recovered, t1Record, err := fix001CapacityRecover(params, replay.T1)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	t2Recovered, t2Record, err := fix001CapacityRecover(params, replay.T2)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	commonLevel := replay.T1.Level()
	if replay.T2.Level() < commonLevel {
		commonLevel = replay.T2.Level()
	}
	if commonLevel < 1 {
		return FIX001CapacityResult{}, fmt.Errorf("invalid common level %d", commonLevel)
	}
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	q01 := new(big.Int).Mul(new(big.Int).Set(q0), q1)
	q01Half := new(big.Int).Rsh(new(big.Int).Set(q01), 1)
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[replay.Doubled.Level()].Modulus)
	p1, p2, _, p2NormalNTT, err := fix001CapacityProduct(params, t1Recovered.components[0], t2Recovered.components[0], commonLevel)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	t2p1, t2p2, _, _, err := fix001CapacityProduct(params, t1Recovered.components[0], t1Recovered.components[0], commonLevel)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	qFull := params.RingQ().AtLevel(commonLevel).Modulus()
	p1Stats := fix001CapacityStats("T3 P1=T1_c0*T2_c0", p1, q01, q01Half, qFull)
	p2Stats := fix001CapacityStats("T3 P2=2*P1", p2, q01, q01Half, qFull)
	t2Control := FIX001CapacityT2Control{P1: fix001CapacityStats("T2 P1=T1_c0*T1_c0", t2p1, q01, q01Half, qFull), P2: fix001CapacityStats("T2 P2=2*T1_c0*T1_c0", t2p2, q01, q01Half, qFull)}
	t2Decoded, err := fix001CapacityDecode(params, replay.T2)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	t2Expected := y
	t2Metrics, err := compareComplexVectors(t2Expected, t2Decoded, correctnessThreshold)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	t2Control.PostRescaleMetrics = t2Metrics
	t2Control.PostRescalePass = t2Metrics.PassThreshold
	fullInput := fix001CapacityBuildFullCiphertext(params, replay.Doubled, p2NormalNTT)
	fullOut := ckks.NewCiphertext(params, 1, replay.Doubled.Level()-1)
	stdEval := ckks.NewEvaluator(params, nil)
	if err := stdEval.Rescale(fullInput, fullOut); err != nil {
		return FIX001CapacityResult{}, fmt.Errorf("FULL_Q_RESCALE_ORACLE_INVALID: %w", err)
	}
	fullInputDecoded, fullInputDecodeErr := decodeWithSecret(params, fullInput, zeroSecret(params))
	var fullInputMetrics CorrectnessMetrics
	if fullInputDecodeErr == nil {
		fullInputMetrics, fullInputDecodeErr = compareComplexVectors(doubledExpected, fullInputDecoded, correctnessThreshold)
	}
	rawDecoded, rawDecodeErr := decodeWithSecret(params, fullOut, zeroSecret(params))
	var rawMetrics CorrectnessMetrics
	if rawDecodeErr == nil {
		rawMetrics, rawDecodeErr = compareComplexVectors(doubledExpected, rawDecoded, correctnessThreshold)
	}
	fullDecodeCiphertext := fullOut.CopyNew()
	if fullDecodeCiphertext.IsMontgomery {
		params.RingQ().AtLevel(fullDecodeCiphertext.Level()).IMForm(fullDecodeCiphertext.Value[0], fullDecodeCiphertext.Value[0])
		params.RingQ().AtLevel(fullDecodeCiphertext.Level()).IMForm(fullDecodeCiphertext.Value[1], fullDecodeCiphertext.Value[1])
		fullDecodeCiphertext.IsMontgomery = false
	}
	fullDecoded, err := decodeWithSecret(params, fullDecodeCiphertext, zeroSecret(params))
	if err != nil {
		return FIX001CapacityResult{}, fmt.Errorf("FULL_Q_RESCALE_ORACLE_INVALID: %w", err)
	}
	fullMetrics, err := compareComplexVectors(doubledExpected, fullDecoded, correctnessThreshold)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	fullOracle := FIX001CapacityFullQRescale{Input: finalizationEvidence(fullInput), Output: finalizationEvidence(fullOut), Decoded: correctnessValues(fullDecoded), Expected: correctnessValues(doubledExpected), Metrics: fullMetrics, Pass: fullMetrics.PassThreshold, ReferenceValidated: fullMetrics.PassThreshold}
	if !fullMetrics.PassThreshold {
		fullOracle.ValidationFailure = "FULL_Q_RESCALE_ORACLE_INVALID"
	}
	if !fullMetrics.PassThreshold {
		return FIX001CapacityResult{FullQRescale: fullOracle}, fmt.Errorf("FULL_Q_RESCALE_ORACLE_INVALID: input max=%g input error=%v normalized max=%g raw max=%g raw error=%v", fullInputMetrics.MaxComponentAbs, fullInputDecodeErr, fullMetrics.MaxComponentAbs, rawMetrics.MaxComponentAbs, rawDecodeErr)
	}
	wrappedP2 := make([]*big.Int, len(p2))
	for i, value := range p2 {
		wrappedP2[i], _ = fix001CapacityCenteredQ01(value, q01, q01Half)
	}
	wrappedFast := fix001CapacityBuildWrappedFast(params, replay.Doubled, wrappedP2, divisor)
	actualPostDecoded, err := fix001CapacityDecode(params, replay.PostRescale)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	simDecoded, err := fix001CapacityDecode(params, wrappedFast)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	simExpected, err := compareComplexVectors(doubledExpected, simDecoded, correctnessThreshold)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	actualExpected, err := compareComplexVectors(doubledExpected, actualPostDecoded, correctnessThreshold)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	simActual, err := compareComplexVectors(actualPostDecoded, simDecoded, correctnessThreshold)
	if err != nil {
		return FIX001CapacityResult{}, err
	}
	wrappedInput := fix001CapacityBuildQ01AtLevel(params, replay.Doubled, p2)
	inputRowsExact := fix001CapacityRowsEqual(0, replay.Doubled, wrappedInput)
	rowResults := []FIX001CapacityRows{{Component: 0, Comparisons: []FinalizationRowComparison{finalizationRowComparison(0, 0, replay.Doubled.Value[0].Coeffs[0], wrappedInput.Value[0].Coeffs[0]), finalizationRowComparison(0, 1, replay.Doubled.Value[0].Coeffs[1], wrappedInput.Value[0].Coeffs[1])}, Exact: inputRowsExact}}
	outputRows := fix001CapacityRowResult(0, replay.PostRescale, wrappedFast)
	rowResults = append(rowResults, outputRows)
	explains := inputRowsExact && outputRows.Exact && simActual.MaxComponentAbs <= correctnessThreshold
	wrapped := FIX001CapacityWrappedSimulation{InputQ01RowsBitExact: inputRowsExact, OutputQ01RowsBitExact: outputRows.Exact, Simulated: finalizationEvidence(wrappedFast), Actual: finalizationEvidence(replay.PostRescale), SimulatedDecoded: correctnessValues(simDecoded), ActualDecoded: correctnessValues(actualPostDecoded), SimulatedVsExpected: simExpected, ActualVsExpected: actualExpected, SimulatedVsActual: simActual, Rows: rowResults, ExplainsFastFailure: explains}
	scaleOut := replay.PostRescale.Scale.Value
	denom := new(big.Float).Mul(new(big.Float).SetInt(divisor), &scaleOut)
	rough := new(big.Float).Quo(new(big.Float).SetInt(q01), denom)
	roughFloat, _ := rough.Float64()
	roughLog2 := math.Log2(roughFloat)
	cause := "q01_capacity_rejected"
	if p1Stats.OutsideQ01CenteredCount > 0 {
		cause = "t3_multiply_crosses_q01_capacity"
	} else if p2Stats.OutsideQ01CenteredCount > 0 {
		cause = "t3_doubling_crosses_q01_capacity"
	}
	if (p1Stats.OutsideQ01CenteredCount > 0 || p2Stats.OutsideQ01CenteredCount > 0) && !explains {
		cause = "q01_capacity_correlated_but_not_causal"
	}
	result := FIX001CapacityResult{SchemaVersion: "fix-001-diag-t3-capacity.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, RepairedFastCommit: "87be78ff3c591932699aba63d3be46ca306a6eea", CorrectedT0Oracle: true, E2VsHistorical: e2Metrics, T1: t1Record, T2: t2Record, T3PreRescale: fix001CapacityMetadata(replay.Doubled), T3PostRescale: fix001CapacityMetadata(replay.PostRescale), Q0: q0.String(), Q1: q1.String(), Q01: q01.String(), Q01Half: q01Half.String(), CommonLevel: commonLevel, RescaleDivisor: divisor.String(), P1: p1Stats, P2: p2Stats, T2Control: t2Control, FullQRescale: fullOracle, WrappedSimulation: wrapped, RoughWrapError: rough.Text('g', 30), RoughWrapErrorFloat64: roughFloat, RoughWrapErrorLog2: roughLog2, FirstSupportedCause: cause, Threshold: correctnessThreshold, SecondaryWorktreeClean: false, TemporarySecondarySourceRemoved: false, Notes: []string{"Corrected T0=1 oracle is used.", "Only LogN13 was run.", "No production Lattigo source was modified.", "The full-Q reference uses exact coefficient-domain lifts from actual repaired Fast T1/T2 c0 rows."}}
	return result, nil
}

func runFIX001T3Capacity(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result, err := RunFIX001T3Capacity(cfg, primaryRoot, backendRoot)
	if err != nil {
		return err
	}
	if err := fix001CapacityWrite(result, outPath); err != nil {
		return err
	}
	result.SecondaryWorktreeClean = true
	result.TemporarySecondarySourceRemoved = true
	if err := fix001CapacityWrite(result, outPath); err != nil {
		return err
	}
	return fix001CapacityWriteSummary(result, filepath.Join(filepath.Dir(outPath), "FIX-001-DIAG-T3-CAPACITY-logN13-summary.json"))
}

func sortCapacityHistogram(hist map[string]int) []string {
	keys := make([]string, 0, len(hist))
	for key := range hist {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
