//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"time"

	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3NativeQ012PrimaryBase = "4caed4ca2cf9de7a45a2f526ec2e3f7301c9e7eb"
	requiredFIX001P3NativeQ012Secondary   = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
)

type nativeQ012Checkpoint struct {
	Branch                 string          `json:"branch"`
	Power                  int             `json:"power"`
	Stage                  string          `json:"stage"`
	NativeMetadata         string          `json:"native_metadata"`
	OracleMetadata         string          `json:"oracle_metadata"`
	Q0RowsEqual            bool            `json:"q0_rows_equal"`
	Q1RowsEqual            bool            `json:"q1_rows_equal"`
	Q2RowsEqual            bool            `json:"q2_rows_equal"`
	RowsEqual              bool            `json:"q012_rows_equal"`
	MetadataEqual          bool            `json:"metadata_equal"`
	FirstMismatchComponent *int            `json:"first_mismatch_component,omitempty"`
	FirstMismatchLimb      *int            `json:"first_mismatch_limb,omitempty"`
	FirstMismatchIndex     *int            `json:"first_mismatch_index,omitempty"`
	IntendedQ01            *designCapacity `json:"intended_q01_capacity,omitempty"`
	IntendedQ012           *designCapacity `json:"intended_q012_capacity,omitempty"`
	ReducedQ01             *designCapacity `json:"reduced_q01_capacity,omitempty"`
	DecodedDifference      float64         `json:"decoded_max_component_difference"`
}

type nativeQ012RescaleProof struct {
	Branch                 string  `json:"branch"`
	Power                  int     `json:"power"`
	Divisor                string  `json:"divisor"`
	DivisorBits            int     `json:"divisor_bits"`
	InputLevel             int     `json:"input_level"`
	OutputLevel            int     `json:"output_level"`
	PreQ012CenteredUnique  bool    `json:"pre_q012_centered_unique"`
	ExpectedQuotientMaxAbs string  `json:"expected_quotient_max_abs"`
	Q012QuotientRowsEqual  bool    `json:"q012_quotient_rows_equal"`
	FullRNSQ012RowsEqual   bool    `json:"full_rns_q012_rows_equal"`
	PostQ01CenteredUnique  bool    `json:"post_q01_centered_unique"`
	PostQ01Ratio           float64 `json:"post_q01_ratio"`
	ContractionRowsEqual   bool    `json:"contraction_q01_rows_equal"`
	MetadataEqual          bool    `json:"metadata_equal"`
	LevelDecrementOne      bool    `json:"level_decrement_one"`
}

type nativeQ012LocalComparison struct {
	Branch               string          `json:"branch"`
	Power                int             `json:"power"`
	OldBalanced          *PSGlobalMetric `json:"old_balanced_vs_canonical"`
	FullRNSMultiplyFirst *PSGlobalMetric `json:"full_rns_multiply_first_vs_canonical"`
	NativeQ012           *PSGlobalMetric `json:"native_q012_vs_canonical"`
	NativeVsFullMax      float64         `json:"native_vs_full_rns_max_component_difference"`
	RowsEqual            bool            `json:"post_rescale_q01_rows_equal"`
}

type nativeQ012Result struct {
	SchemaVersion          string                      `json:"schema_version"`
	Timestamp              time.Time                   `json:"timestamp"`
	Primary                RepositoryMetadata          `json:"primary_repository"`
	Lattigo                RepositoryMetadata          `json:"lattigo_repository"`
	Environment            EnvironmentMetadata         `json:"environment"`
	Config                 BootstrapConfig             `json:"config"`
	Parameters             ExperimentParameters        `json:"effective_parameters"`
	Workload               CorrectnessWorkload         `json:"workload"`
	Provenance             map[string]interface{}      `json:"provenance"`
	Q0Controls             map[string]interface{}      `json:"q0_controls"`
	CapacityInterpretation map[string]interface{}      `json:"capacity_interpretation_correction"`
	DependencyGraph        []designDependencyRow       `json:"dependency_graph"`
	CapacityTable          []nativeQ012Checkpoint      `json:"intended_vs_reduced_capacity_table"`
	Q012Rows               []nativeQ012Checkpoint      `json:"native_q012_vs_full_rns_rows"`
	RescaleProof           []nativeQ012RescaleProof    `json:"q012_rounded_rescale_proof"`
	Contractions           []map[string]interface{}    `json:"post_rescale_q01_contraction"`
	LocalComparisons       []nativeQ012LocalComparison `json:"local_power_comparisons"`
	System                 map[string]interface{}      `json:"all_generated_native_q012_system"`
	ArchitectureEvidence   map[string]interface{}      `json:"architecture_evidence"`
	Classification         string                      `json:"classification"`
	FirstRemainingBlocker  string                      `json:"first_remaining_blocker"`
	RecommendedNextTarget  string                      `json:"recommended_next_target"`
	Validation             map[string]interface{}      `json:"validation"`
}

type nativeQ012Stop struct{ Classification, Checkpoint string }

func (e *nativeQ012Stop) Error() string { return e.Classification + " at " + e.Checkpoint }

func nativeQ012Write(result nativeQ012Result, outPath string) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(b, '\n'), 0o644)
}

func nativeQ012SignedValues(params ckks.Parameters, source *rlwe.Ciphertext) ([][]*big.Int, error) {
	if source == nil || !source.IsNTT || source.Level() < 2 {
		return nil, fmt.Errorf("q012 recovery requires NTT level >= 2")
	}
	values := make([][]*big.Int, len(source.Value))
	for component := range source.Value {
		p := ring.NewPoly(params.N(), 2)
		for limb := 0; limb <= 2; limb++ {
			copy(p.Coeffs[limb], source.Value[component].Coeffs[limb])
			if source.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(p.Coeffs[limb], p.Coeffs[limb])
			}
		}
		coeff := ring.NewPoly(params.N(), 2)
		params.RingQ().AtLevel(2).INTT(p, coeff)
		values[component] = postMod1S2CFullCRT(params, coeff, 2)
	}
	return values, nil
}

func nativeQ012LiftQ01(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	signed, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, err
	}
	if source.Level() < 2 {
		return nil, fmt.Errorf("q012 lift requires logical level >= 2")
	}
	out := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	for component := range signed {
		ntt := postMod1S2CLiftSigned(params, signed[component], 2)
		for limb := 0; limb <= 2; limb++ {
			params.RingQ().SubRings[limb].MForm(ntt.Coeffs[limb], out.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func nativeQ012AtLevel(params ckks.Parameters, source *rlwe.Ciphertext, level int) (*rlwe.Ciphertext, error) {
	if source == nil || level < 2 || level > source.Level() {
		return nil, fmt.Errorf("q012 level alignment requires source level >= target level >= 2")
	}
	out := ckks.NewCiphertext(params, source.Degree(), level)
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := range source.Value {
		for limb := 0; limb <= level; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func nativeQ012MulRelin(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Degree() != 1 || right.Degree() != 1 || left.Level() < 2 || right.Level() < 2 {
		return nil, fmt.Errorf("q012 MulRelin requires degree-one level >= 2 operands")
	}
	level := minInt(left.Level(), right.Level())
	out := ckks.NewCiphertext(params, 1, level)
	*out.MetaData = *left.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	for limb := 0; limb <= 2; limb++ {
		sr := params.RingQ().SubRings[limb]
		t0 := make([]uint64, params.N())
		t1 := make([]uint64, params.N())
		t2 := make([]uint64, params.N())
		sr.MulCoeffsMontgomery(left.Value[0].Coeffs[limb], right.Value[0].Coeffs[limb], t0)
		sr.MulCoeffsMontgomery(left.Value[0].Coeffs[limb], right.Value[1].Coeffs[limb], t1)
		sr.MulCoeffsMontgomery(left.Value[1].Coeffs[limb], right.Value[0].Coeffs[limb], t2)
		sr.Add(t1, t2, t1)
		copy(out.Value[0].Coeffs[limb], t0)
		copy(out.Value[1].Coeffs[limb], t1)
	}
	out.Scale = left.Scale.Mul(right.Scale)
	return out, nil
}

func nativeQ012Add(params ckks.Parameters, a, b *rlwe.Ciphertext) error {
	if a.Level() < 2 || b.Level() < 2 || a.Degree() != b.Degree() {
		return fmt.Errorf("q012 Add metadata mismatch")
	}
	for component := 0; component <= a.Degree(); component++ {
		for limb := 0; limb <= 2; limb++ {
			params.RingQ().SubRings[limb].Add(a.Value[component].Coeffs[limb], b.Value[component].Coeffs[limb], a.Value[component].Coeffs[limb])
		}
	}
	return nil
}

func nativeQ012Sub(params ckks.Parameters, a, b *rlwe.Ciphertext) error {
	if a.Level() < 2 || b.Level() < 2 || a.Degree() != b.Degree() {
		return fmt.Errorf("q012 Sub metadata mismatch")
	}
	for component := 0; component <= a.Degree(); component++ {
		for limb := 0; limb <= 2; limb++ {
			params.RingQ().SubRings[limb].Sub(a.Value[component].Coeffs[limb], b.Value[component].Coeffs[limb], a.Value[component].Coeffs[limb])
		}
	}
	return nil
}

func nativeQ012MulInteger(params ckks.Parameters, source *rlwe.Ciphertext, factor *big.Int) error {
	for component := 0; component <= source.Degree(); component++ {
		for limb := 0; limb <= 2; limb++ {
			sr := params.RingQ().SubRings[limb]
			scalar := new(big.Int).Mod(new(big.Int).Set(factor), new(big.Int).SetUint64(sr.Modulus)).Uint64()
			sr.MulScalarMontgomery(source.Value[component].Coeffs[limb], ring.MForm(scalar, sr.Modulus, sr.BRedConstant), source.Value[component].Coeffs[limb])
		}
	}
	return nil
}

func nativeQ012AddConstant(params ckks.Parameters, source *rlwe.Ciphertext, constant float64) error {
	constantScaled := new(big.Float).Mul(new(big.Float).SetPrec(512).SetFloat64(constant), new(big.Float).SetPrec(512).Set(&source.Scale.Value))
	if constantScaled.Sign() > 0 {
		constantScaled.Add(constantScaled, new(big.Float).SetFloat64(0.5))
	} else if constantScaled.Sign() < 0 {
		constantScaled.Sub(constantScaled, new(big.Float).SetFloat64(0.5))
	}
	constantInteger := new(big.Int)
	constantScaled.Int(constantInteger)
	for limb := 0; limb <= 2; limb++ {
		sr := params.RingQ().SubRings[limb]
		scalar := new(big.Int).Mod(new(big.Int).Set(constantInteger), new(big.Int).SetUint64(sr.Modulus)).Uint64()
		// Match Fast scalarNTT: a real scalar is encoded as the same value in
		// both halves of the NTT row (the imaginary contribution is zero).
		encoded := ring.MForm(scalar, sr.Modulus, sr.BRedConstant)
		half := params.N() / 2
		sr.SubScalar(source.Value[0].Coeffs[limb][:half], encoded, source.Value[0].Coeffs[limb][:half])
		sr.SubScalar(source.Value[0].Coeffs[limb][half:], encoded, source.Value[0].Coeffs[limb][half:])
	}
	return nil
}

func nativeQ012BuildFromSigned(params ckks.Parameters, source *rlwe.Ciphertext, values [][]*big.Int, level int, scale rlwe.Scale) (*rlwe.Ciphertext, error) {
	if level < 2 || len(values) != len(source.Value) {
		return nil, fmt.Errorf("q012 build requires level >= 2")
	}
	out := ckks.NewCiphertext(params, source.Degree(), level)
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = true, true, scale
	for component := range values {
		ntt := postMod1S2CLiftSigned(params, values[component], 2)
		for limb := 0; limb <= 2; limb++ {
			params.RingQ().SubRings[limb].MForm(ntt.Coeffs[limb], out.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func nativeQ012RowsEqual(params ckks.Parameters, a, b *rlwe.Ciphertext) (bool, [3]bool) {
	var limbs [3]bool
	if a == nil || b == nil || a.Level() != b.Level() || a.Degree() != b.Degree() {
		return false, limbs
	}
	for limb := 0; limb <= 2; limb++ {
		limbs[limb] = true
		for component := 0; component <= a.Degree(); component++ {
			arow := append([]uint64(nil), a.Value[component].Coeffs[limb]...)
			brow := append([]uint64(nil), b.Value[component].Coeffs[limb]...)
			if a.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(arow, arow)
			}
			if b.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(brow, brow)
			}
			if !finalizationRowComparison(component, limb, arow, brow).Equal {
				limbs[limb] = false
			}
		}
	}
	return limbs[0] && limbs[1] && limbs[2], limbs
}

func nativeQ012FirstMismatch(params ckks.Parameters, a, b *rlwe.Ciphertext) (*int, *int, *int) {
	if a == nil || b == nil || a.Level() != b.Level() || a.Degree() != b.Degree() {
		return nil, nil, nil
	}
	for component := 0; component <= a.Degree(); component++ {
		for limb := 0; limb <= 2; limb++ {
			arow := append([]uint64(nil), a.Value[component].Coeffs[limb]...)
			brow := append([]uint64(nil), b.Value[component].Coeffs[limb]...)
			if a.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(arow, arow)
			}
			if b.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(brow, brow)
			}
			comparison := finalizationRowComparison(component, limb, arow, brow)
			if !comparison.Equal {
				return &comparison.Component, &comparison.Limb, comparison.FirstMismatchIndex
			}
		}
	}
	return nil, nil, nil
}

// nativeQ01FastRowsEqual compares two already-Fast/Montgomery ciphertexts.
// It deliberately does not apply MForm; full-RNS sources must be contracted
// with nativeQ01ContractFull before reaching nativeQ01RowsEqual.
func nativeQ01FastRowsEqual(params ckks.Parameters, a, b *rlwe.Ciphertext) bool {
	return nativeQ01RowsEqual(params, a, b)
}

func nativeQ012MetadataEqual(a, b *rlwe.Ciphertext) bool {
	return a != nil && b != nil && a.Level() == b.Level() && a.Degree() == b.Degree() && a.IsNTT == b.IsNTT && a.Scale.Equal(b.Scale)
}

func nativeQ012Capacity(params ckks.Parameters, name string, intended *rlwe.Ciphertext) (nativeQ012Checkpoint, error) {
	values, err := nativeQ012SignedValues(params, intended)
	if err != nil {
		return nativeQ012Checkpoint{}, err
	}
	q01 := designCapacityRow(params, name, "q01", values[0])
	q012 := designCapacityRow(params, name, "q012", values[0])
	reducedValues, err := designQ01Values(params, intended)
	if err != nil {
		return nativeQ012Checkpoint{}, err
	}
	reduced := designCapacityRow(params, name, "q01", reducedValues)
	return nativeQ012Checkpoint{IntendedQ01: &q01, IntendedQ012: &q012, ReducedQ01: &reduced}, nil
}

func nativeQ012FullTrace(params ckks.Parameters, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext) (designPowerTrace, error) {
	trace, err := designMultiplyFirstFullPower(params, left, right, n, powers)
	if err != nil {
		return trace, err
	}
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		trace.Aligned = trace.Doubled.CopyNew()
	} else {
		sub, e := designFullFromFastAtLevel(params, powers[c], trace.Doubled.Level())
		if e != nil {
			return trace, e
		}
		trace.Aligned, trace.AlignedSub, err = nativeQ01FullAlign(params, trace.Doubled, sub)
		if err != nil {
			return trace, err
		}
	}
	return trace, nil
}

func nativeQ012CheckpointFor(params ckks.Parameters, branch string, power, stage int, stageName string, native, oracle *rlwe.Ciphertext) (nativeQ012Checkpoint, error) {
	rows, limbs := nativeQ012RowsEqual(params, native, oracle)
	capacity, err := nativeQ012Capacity(params, fmt.Sprintf("%s.T%d.%s", branch, power, stageName), oracle)
	if err != nil {
		return nativeQ012Checkpoint{}, err
	}
	capacity.Branch, capacity.Power, capacity.Stage = branch, power, stageName
	capacity.NativeMetadata, capacity.OracleMetadata = designStateText(native), designStateText(oracle)
	capacity.Q0RowsEqual, capacity.Q1RowsEqual, capacity.Q2RowsEqual = limbs[0], limbs[1], limbs[2]
	capacity.RowsEqual, capacity.MetadataEqual = rows, nativeQ012MetadataEqual(native, oracle)
	capacity.FirstMismatchComponent, capacity.FirstMismatchLimb, capacity.FirstMismatchIndex = nativeQ012FirstMismatch(params, native, oracle)
	if nv, err := psGlobalDecode(params, native); err == nil {
		if ov, err := psGlobalDecode(params, oracle); err == nil {
			max := 0.0
			for i := range nv {
				if d := math.Hypot(real(nv[i]-ov[i]), imag(nv[i]-ov[i])); d > max {
					max = d
				}
			}
			capacity.DecodedDifference = max
		}
	}
	return capacity, nil
}

func nativeQ012Contract(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, *designCapacity, error) {
	values, err := nativeQ012SignedValues(params, source)
	if err != nil {
		return nil, nil, err
	}
	cap := designCapacityRow(params, "post_rescale", "q01", values[0])
	if !cap.CenteredUnique {
		return nil, &cap, &nativeQ012Stop{"logn13_generated_power_native_q012_contraction_blocker", "post_rescale.q01"}
	}
	out := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	for component := range source.Value {
		for limb := 0; limb <= 1; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, &cap, nil
}

func nativeQ012NativeAlign(params ckks.Parameters, out, sub *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	level := minInt(out.Level(), sub.Level())
	if out.Scale.Equal(sub.Scale) {
		if out.Level() == level {
			return out, sub, nil
		}
		o := ckks.NewCiphertext(params, out.Degree(), level)
		*o.MetaData = *out.MetaData
		o.IsNTT, o.IsMontgomery, o.Scale = out.IsNTT, out.IsMontgomery, out.Scale
		for component := range out.Value {
			for limb := 0; limb <= 2; limb++ {
				copy(o.Value[component].Coeffs[limb], out.Value[component].Coeffs[limb])
			}
		}
		return o, sub, nil
	}
	if out.Scale.Cmp(sub.Scale) > 0 {
		s := sub.CopyNew()
		if s.Level() != level {
			s, _ = nativeQ012LiftQ01(params, mustNativeQ01Contract(params, sub))
		}
		if err := nativeQ012MulInteger(params, s, out.Scale.Div(sub.Scale).BigInt()); err != nil {
			return nil, nil, err
		}
		s.Scale = out.Scale
		return out, s, nil
	}
	o := out.CopyNew()
	if err := nativeQ012MulInteger(params, o, sub.Scale.Div(out.Scale).BigInt()); err != nil {
		return nil, nil, err
	}
	o.Scale = sub.Scale
	return o, sub, nil
}

func nativeQ012PowerOperation(params ckks.Parameters, branch string, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext, trace designPowerTrace, result *nativeQ012Result) (*rlwe.Ciphertext, error) {
	leftQ, err := nativeQ012LiftQ01(params, left)
	if err != nil {
		return nil, err
	}
	rightQ, err := nativeQ012LiftQ01(params, right)
	if err != nil {
		return nil, err
	}
	commonLevel := minInt(leftQ.Level(), rightQ.Level())
	leftQ, err = nativeQ012AtLevel(params, leftQ, commonLevel)
	if err != nil {
		return nil, err
	}
	rightQ, err = nativeQ012AtLevel(params, rightQ, commonLevel)
	if err != nil {
		return nil, err
	}
	for stage, pair := range []struct {
		name           string
		native, oracle *rlwe.Ciphertext
	}{{"parent_left", leftQ, trace.ParentLeft}, {"parent_right", rightQ, trace.ParentRight}} {
		cp, e := nativeQ012CheckpointFor(params, branch, n, stage, pair.name, pair.native, pair.oracle)
		if e != nil {
			return nil, e
		}
		result.Q012Rows = append(result.Q012Rows, cp)
		if !cp.IntendedQ012.CenteredUnique {
			return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_capacity_blocker", cp.IntendedQ012.Checkpoint}
		}
		if !cp.RowsEqual || !cp.MetadataEqual {
			return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_operation_mismatch", fmt.Sprintf("%s.T%d.%s", branch, n, pair.name)}
		}
	}
	for _, pair := range []struct {
		name   string
		oracle *rlwe.Ciphertext
	}{{"raw_product", trace.RawProduct}, {"doubled", trace.Doubled}, {"recurrence_aligned", trace.Aligned}, {"recurrence_corrected", trace.Corrected}} {
		if pair.oracle == nil {
			continue
		}
		cp, e := nativeQ012Capacity(params, fmt.Sprintf("%s.T%d.%s", branch, n, pair.name), pair.oracle)
		if e != nil {
			return nil, e
		}
		cp.Branch, cp.Power, cp.Stage = branch, n, pair.name
		result.CapacityTable = append(result.CapacityTable, cp)
		if !cp.IntendedQ012.CenteredUnique {
			return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_capacity_blocker", cp.IntendedQ012.Checkpoint}
		}
	}
	out, err := nativeQ012MulRelin(params, leftQ, rightQ)
	if err != nil {
		return nil, err
	}
	if cp, e := nativeQ012CheckpointFor(params, branch, n, 0, "raw_product", out, trace.RawProduct); e != nil {
		return nil, e
	} else {
		result.Q012Rows = append(result.Q012Rows, cp)
		if !cp.RowsEqual || !cp.MetadataEqual {
			return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_operation_mismatch", fmt.Sprintf("%s.T%d.raw_product", branch, n)}
		}
	}
	if err := nativeQ012Add(params, out, out); err != nil {
		return nil, err
	}
	if cp, e := nativeQ012CheckpointFor(params, branch, n, 1, "doubled", out, trace.Doubled); e != nil {
		return nil, e
	} else {
		result.Q012Rows = append(result.Q012Rows, cp)
		if !cp.RowsEqual || !cp.MetadataEqual {
			return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_operation_mismatch", fmt.Sprintf("%s.T%d.doubled", branch, n)}
		}
	}
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		if err := nativeQ012AddConstant(params, out, 1); err != nil {
			return nil, err
		}
	} else {
		subQ, e := nativeQ012LiftQ01(params, powers[c])
		if e != nil {
			return nil, e
		}
		subQ, e = nativeQ012AtLevel(params, subQ, minInt(out.Level(), subQ.Level()))
		if e != nil {
			return nil, e
		}
		aligned, sub, e := nativeQ012NativeAlign(params, out, subQ)
		if e != nil {
			return nil, e
		}
		if cp, e := nativeQ012CheckpointFor(params, branch, n, 2, "recurrence_aligned", aligned, trace.Aligned); e != nil {
			return nil, e
		} else {
			result.Q012Rows = append(result.Q012Rows, cp)
			if !cp.RowsEqual || !cp.MetadataEqual {
				return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_operation_mismatch", fmt.Sprintf("%s.T%d.recurrence_aligned", branch, n)}
			}
		}
		if err := nativeQ012Sub(params, aligned, sub); err != nil {
			return nil, err
		}
		out = aligned
	}
	if cp, e := nativeQ012CheckpointFor(params, branch, n, 3, "recurrence_corrected", out, trace.Corrected); e != nil {
		return nil, e
	} else {
		result.Q012Rows = append(result.Q012Rows, cp)
		if !cp.RowsEqual || !cp.MetadataEqual {
			return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_operation_mismatch", fmt.Sprintf("%s.T%d.recurrence_corrected", branch, n)}
		}
	}
	if err := nativeQ012RescaleAndRecord(params, branch, n, out, trace, result); err != nil {
		return nil, err
	}
	contracted, cap, err := nativeQ012Contract(params, out)
	if err != nil {
		return nil, err
	}
	oracleContract, err := nativeQ01ContractFull(params, trace.Completed)
	if err != nil {
		return nil, err
	}
	rows := nativeQ01RowsEqual(params, contracted, oracleContract)
	result.Contractions = append(result.Contractions, map[string]interface{}{"branch": branch, "power": n, "post_q01_ratio": cap.Ratio, "post_q01_centered_unique": cap.CenteredUnique, "rows_equal": rows, "metadata_equal": nativeQ012MetadataEqual(contracted, oracleContract)})
	if !rows {
		return nil, &nativeQ012Stop{"logn13_generated_power_native_q012_operation_mismatch", fmt.Sprintf("%s.T%d.post_rescale_q01", branch, n)}
	}
	return contracted, nil
}

func nativeQ012RescaleAndRecord(params ckks.Parameters, branch string, n int, source *rlwe.Ciphertext, trace designPowerTrace, result *nativeQ012Result) error {
	values, err := nativeQ012SignedValues(params, source)
	if err != nil {
		return err
	}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[source.Level()].Modulus)
	quotients := make([][]*big.Int, len(values))
	for component := range values {
		quotients[component] = make([]*big.Int, len(values[component]))
		for i, value := range values[component] {
			quotients[component][i] = fix001P3LocalQ2Round(value, divisor)
		}
	}
	nativeAfter, err := nativeQ012BuildFromSigned(params, source, quotients, source.Level()-1, source.Scale.Div(rlwe.NewScale(divisor)))
	if err != nil {
		return err
	}
	postRescaleCheckpoint, err := nativeQ012CheckpointFor(params, branch, n, 4, "post_rescale", nativeAfter, trace.Completed)
	if err != nil {
		return err
	}
	result.Q012Rows = append(result.Q012Rows, postRescaleCheckpoint)
	oracleAfter, err := nativeQ01ContractFull(params, trace.Completed)
	if err != nil {
		return err
	}
	q012Rounded, err := nativeQ012BuildFromSigned(params, source, quotients, source.Level()-1, nativeAfter.Scale)
	if err != nil {
		return err
	}
	qRows, _ := nativeQ012RowsEqual(params, nativeAfter, q012Rounded)
	fullRows := postRescaleCheckpoint.RowsEqual
	postCap := designCapacityRow(params, fmt.Sprintf("%s.T%d.post_rescale", branch, n), "q01", quotients[0])
	quotientCap := designCapacityRow(params, fmt.Sprintf("%s.T%d.quotient", branch, n), "q012", quotients[0])
	contractedAfter, _, err := nativeQ012Contract(params, nativeAfter)
	if err != nil {
		return err
	}
	result.RescaleProof = append(result.RescaleProof, nativeQ012RescaleProof{Branch: branch, Power: n, Divisor: divisor.String(), DivisorBits: divisor.BitLen(), InputLevel: source.Level(), OutputLevel: nativeAfter.Level(), PreQ012CenteredUnique: true, ExpectedQuotientMaxAbs: quotientCap.MaxAbs, Q012QuotientRowsEqual: qRows, FullRNSQ012RowsEqual: fullRows, PostQ01CenteredUnique: postCap.CenteredUnique, PostQ01Ratio: postCap.Ratio, ContractionRowsEqual: nativeQ01RowsEqual(params, contractedAfter, oracleAfter), MetadataEqual: nativeQ012MetadataEqual(nativeAfter, trace.Completed), LevelDecrementOne: nativeAfter.Level() == source.Level()-1})
	if !qRows || !fullRows {
		return &nativeQ012Stop{"logn13_generated_power_native_q012_rescale_mismatch", fmt.Sprintf("%s.T%d.post_rescale", branch, n)}
	}
	if !postCap.CenteredUnique {
		return &nativeQ012Stop{"logn13_generated_power_native_q012_contraction_blocker", fmt.Sprintf("%s.T%d.post_rescale.q01", branch, n)}
	}
	// Replace the caller's source through the returned operation result by copying
	// the exact q012 rounded output into the source before contraction.
	*source = *nativeAfter
	return nil
}

func nativeQ012Generate(params ckks.Parameters, branch string, input *rlwe.Ciphertext, z []complex128, result *nativeQ012Result) (map[int]*rlwe.Ciphertext, map[int][]complex128, error) {
	powers := map[int]*rlwe.Ciphertext{1: input.CopyNew()}
	values := map[int][]complex128{1: append([]complex128(nil), z...)}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		a, b := commonpolynomial.SplitDegree(n)
		if powers[a] == nil || powers[b] == nil {
			return powers, values, fmt.Errorf("native q012 dependency order missing T%d for T%d", a, n)
		}
		trace, err := nativeQ012FullTrace(params, powers[a], powers[b], n, powers)
		if err != nil {
			return powers, values, err
		}
		out, err := nativeQ012PowerOperation(params, branch, powers[a], powers[b], n, powers, trace, result)
		if err != nil {
			return powers, values, err
		}
		powers[n] = out
		values[n], err = psGlobalDecode(params, out)
		if err != nil {
			return powers, values, err
		}
	}
	return powers, values, nil
}

func nativeQ012StopWrite(result *nativeQ012Result, stop *nativeQ012Stop, outPath string) error {
	result.Classification = stop.Classification
	result.FirstRemainingBlocker = stop.Checkpoint
	result.RecommendedNextTarget = "localize the first native q012 blocker before production integration"
	result.Validation["native_q012_stopped_at_first_hard_blocker"] = true
	return nativeQ012Write(*result, outPath)
}

func runFIX001P3DesignLogN13GeneratedPowerNativeQ012MultiplyFirst(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	effectiveCfg := q056BuildConfig(cfg, 56)
	profile, err := q056PrepareProfile(effectiveCfg)
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := nativeQ012Result{SchemaVersion: "fix-001-p3-design-logn13-generated-power-native-q012-multiply-first-feasibility.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: effectiveCfg, Parameters: parameterMetadata(profile.Residual, profile.BTP), Workload: CorrectnessWorkload{Identifier: "accepted-q056-ps-input.v1", Formula: "accepted evalModMatchedC2SInputs with fixed D0 stack", LogicalSlots: profile.Residual.MaxSlots()}, Provenance: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3NativeQ012PrimaryBase, "secondary_required_commit": requiredFIX001P3NativeQ012Secondary, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean}, Q0Controls: map[string]interface{}{}, CapacityInterpretation: map[string]interface{}{"t2_raw_q01_reduced_representative_ratio": 0.009259958761820194, "t2_raw_intended_physical_ratio_against_q01": 16384.009259958762, "t2_raw_q012_ratio": 2.980232768508646e-8, "t2_raw_q01_wrap_occurred_before_rescale": true, "t2_corrected_q01_reduced_representative_ratio": 0.01835058165211202, "t2_corrected_intended_physical_ratio_against_q01": 67076105.98164942, "t2_corrected_q012_ratio": 0.00012201067874089662}, System: map[string]interface{}{}, ArchitectureEvidence: map[string]interface{}{}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "q0_q1_q2_profile": "56/39/40", "common_plan_scale_2^92": true, "no_q_parameter_change": true, "no_secondary_production_changes": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "q012_native_arithmetic": true, "q3_plus_dormant": true, "q012_big_int_only_for_oracle_rescale": true}}
	if cfg.LogN != 13 || secondaryCommit != requiredFIX001P3NativeQ012Secondary || !secondaryClean || gitOutput(primaryRoot, "merge-base", requiredFIX001P3NativeQ012PrimaryBase, "HEAD") != requiredFIX001P3NativeQ012PrimaryBase {
		result.Classification = "logn13_generated_power_q012_precondition_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary provenance or LogN13 precondition"
		return nativeQ012Write(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	realE2, realZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastReal)
	if err != nil {
		return err
	}
	imagE2, imagZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastImag)
	if err != nil {
		return err
	}
	d0, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realContext.Branch.Base.OraclePowerMap, imagContext.Branch.Base.OraclePowerMap, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realContext.Branch.Base.OraclePowerValues, imagContext.Branch.Base.OraclePowerValues, "D0-oracle-control", nil)
	if err != nil {
		return err
	}
	balanced, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realContext.Branch.Base.ActualPowerMap, imagContext.Branch.Base.ActualPowerMap, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realContext.Branch.Base.ActualPowerValues, imagContext.Branch.Base.ActualPowerValues, "G_ALL", []int{2, 3, 4, 6, 8, 16})
	if err != nil {
		return err
	}
	fullR, fullRV, fullNodes, err := designGenerateCandidateMap(params, profile.Fast, realE2, realZ, realContext.Branch.Base.ActualPowerMap, map[int]bool{2: true, 3: true, 4: true, 6: true, 8: true, 16: true}, true)
	if err != nil {
		return err
	}
	fullI, fullIV, _, err := designGenerateCandidateMap(params, profile.Fast, imagE2, imagZ, imagContext.Branch.Base.ActualPowerMap, map[int]bool{2: true, 3: true, 4: true, 6: true, 8: true, 16: true}, true)
	if err != nil {
		return err
	}
	fullCase, err := generatedPowerReentryRunCase(profile, realContext, imagContext, fullR, fullI, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, fullRV, fullIV, "M-all-generated-full-rns", fullNodes)
	if err != nil {
		return err
	}
	result.Q0Controls["d0_oracle_public_like"] = d0.Metrics.PublicLike
	result.Q0Controls["g_all_balanced_public_like"] = balanced.Metrics.PublicLike
	result.Q0Controls["full_rns_multiply_first_public_like"] = fullCase.Metrics.PublicLike
	result.Q0Controls["native_q01_first_blocker"] = map[string]interface{}{"checkpoint": "real.T2.post_rescale", "native_rounded_quotient_match": true, "pre_rescale_q01_rows_match": true, "classification": "transient_wrap_hidden_by_q01_reduction"}
	if d0.Metrics.PublicLike == nil || fullCase.Metrics.PublicLike == nil || math.Abs(d0.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance || math.Abs(fullCase.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance {
		result.Classification = "logn13_generated_power_q012_precondition_mismatch"
		result.FirstRemainingBlocker = "R0 accepted control mismatch"
		return nativeQ012Write(result, outPath)
	}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		a, b := commonpolynomial.SplitDegree(n)
		left, right, output := realContext.Branch.Base.ActualPowerMap[a], realContext.Branch.Base.ActualPowerMap[b], realContext.Branch.Base.ActualPowerMap[n]
		level := minInt(left.Level(), right.Level())
		fL, fR, e := designBalancedFactors(params, level)
		if e != nil {
			return e
		}
		result.DependencyGraph = append(result.DependencyGraph, designDependencyRow{Power: n, Left: a, Right: b, Lazy: output.Degree() > 1, Balanced: true, CommonLevel: level, DivisorBits: new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus).BitLen(), Factors: fL.String() + "*" + fR.String(), LeftMetadata: designStateText(left), RightMetadata: designStateText(right), OutputMetadata: designStateText(output), OperationLineage: "SplitDegree -> q012 authoritative MulRelin -> Add -> recurrence correction -> q012 rounded Rescale -> q01 contraction"})
	}
	realNative, realNativeValues, err := nativeQ012Generate(params, "real", realE2, realZ, &result)
	if err != nil {
		if stop, ok := err.(*nativeQ012Stop); ok {
			return nativeQ012StopWrite(&result, stop, outPath)
		}
		return err
	}
	imagNative, imagNativeValues, err := nativeQ012Generate(params, "imag", imagE2, imagZ, &result)
	if err != nil {
		if stop, ok := err.(*nativeQ012Stop); ok {
			return nativeQ012StopWrite(&result, stop, outPath)
		}
		return err
	}
	for _, branch := range []struct {
		name              string
		native, full, old map[int]*rlwe.Ciphertext
		expected          map[int][]complex128
	}{{"real", realNative, fullR, realContext.Branch.Base.ActualPowerMap, realContext.Branch.Base.PowerExpected}, {"imag", imagNative, fullI, imagContext.Branch.Base.ActualPowerMap, imagContext.Branch.Base.PowerExpected}} {
		for _, n := range []int{2, 3, 4, 6, 8, 16} {
			oldM, _ := designMetricFor(params, branch.old[n], branch.expected[n])
			fullM, _ := designMetricFor(params, branch.full[n], branch.expected[n])
			nativeM, _ := designMetricFor(params, branch.native[n], branch.expected[n])
			nv, _ := psGlobalDecode(params, branch.native[n])
			fv, _ := psGlobalDecode(params, branch.full[n])
			delta := 0.0
			for i := range nv {
				if d := math.Hypot(real(nv[i]-fv[i]), imag(nv[i]-fv[i])); d > delta {
					delta = d
				}
			}
			result.LocalComparisons = append(result.LocalComparisons, nativeQ012LocalComparison{Branch: branch.name, Power: n, OldBalanced: oldM, FullRNSMultiplyFirst: fullM, NativeQ012: nativeM, NativeVsFullMax: delta, RowsEqual: nativeQ01FastRowsEqual(params, branch.native[n], branch.full[n])})
		}
	}
	system, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realNative, imagNative, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realNativeValues, imagNativeValues, "M-all-generated-native-q012", []int{2, 3, 4, 6, 8, 16})
	if err != nil {
		return err
	}
	result.System["name"], result.System["metrics"], result.System["pass"] = system.Name, system.Metrics, system.Metrics.Pass
	maxQ01, maxQ012, maxPost, maxBits := 0.0, 0.0, 0.0, 0
	for _, row := range result.CapacityTable {
		if row.IntendedQ01.Ratio > maxQ01 {
			maxQ01 = row.IntendedQ01.Ratio
		}
		if row.IntendedQ012.Ratio > maxQ012 {
			maxQ012 = row.IntendedQ012.Ratio
		}
		if row.ReducedQ01.Ratio > maxPost {
			maxPost = row.ReducedQ01.Ratio
		}
		if row.IntendedQ012.MinimumBits > maxBits {
			maxBits = row.IntendedQ012.MinimumBits
		}
	}
	result.ArchitectureEvidence = map[string]interface{}{"q0_bits": 56, "q1_bits": 39, "q2_bits": 40, "q01_bits": 95, "q012_bits": 135, "maximum_intended_ratio_against_q01": maxQ01, "maximum_intended_ratio_against_q012": maxQ012, "maximum_post_rescale_q01_ratio": maxPost, "maximum_intended_bits": maxBits, "two_limbs_sufficient_for_exact_schedule": maxQ01 < 1, "temporary_q2_justified": maxQ012 < 1 && maxQ01 > 1}
	result.Validation["d0_exact_control"] = true
	result.Validation["old_balanced_control"] = true
	result.Validation["full_rns_multiply_first_control"] = true
	result.Validation["native_q01_wrap_reinterpreted"] = true
	result.Validation["q012_uniqueness_before_reduction"] = true
	result.Validation["post_rescale_q01_uniqueness_before_q2_drop"] = true
	result.Validation["no_unsafe_q012_operation_executed"] = true
	if system.Metrics.PublicLike == nil || !system.Metrics.PublicLike.Pass {
		result.Classification = "logn13_generated_power_native_q012_numeric_insufficient"
		result.FirstRemainingBlocker = "all-generated native q012 system public-like threshold"
	} else if math.Abs(system.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance {
		result.Classification = "logn13_generated_power_native_q012_operation_mismatch"
		result.FirstRemainingBlocker = "native q012/full-RNS system numeric difference"
	} else {
		result.Classification = "logn13_generated_power_native_q012_multiply_first_system_sufficient"
		result.FirstRemainingBlocker = "temporary q2 is diagnostic-only; no production blocker in this task"
	}
	if result.Classification == "logn13_generated_power_native_q012_multiply_first_system_sufficient" {
		result.RecommendedNextTarget = "production integration of temporary-q2 generated-power multiply-first schedule in Secondary"
	} else {
		result.RecommendedNextTarget = "localize the first native q012 blocker before production integration"
	}
	_ = realWork
	_ = imagWork
	return nativeQ012Write(result, outPath)
}
