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

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3NativeQ01PrimaryBase = "904f0460e36b69add5491f7c7006c72153970789"
	requiredFIX001P3NativeQ01Secondary   = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	nativeQ01PublicThreshold             = 1e-2
	nativeQ01LocalThreshold              = 1e-6
)

type nativeQ01Checkpoint struct {
	Branch               string          `json:"branch"`
	Power                int             `json:"power"`
	Stage                string          `json:"stage"`
	NativeMetadata       string          `json:"native_metadata"`
	MirrorMetadata       string          `json:"mirror_metadata"`
	RowsEqual            bool            `json:"q0_q1_rows_equal"`
	MetadataEqual        bool            `json:"metadata_equal"`
	Capacity             *designCapacity `json:"mirror_capacity,omitempty"`
	DecodedMaxDifference float64         `json:"decoded_max_component_difference"`
}

type nativeQ01RescaleProof struct {
	Branch                            string `json:"branch"`
	Power                             int    `json:"power"`
	Divisor                           string `json:"divisor"`
	InputLevel                        int    `json:"input_level"`
	OutputLevel                       int    `json:"output_level"`
	RoundedQuotientRowsEqual          bool   `json:"rounded_quotient_rows_equal"`
	NativeQuotientRowsEqual           bool   `json:"native_quotient_rows_equal"`
	MirrorQuotientRowsEqual           bool   `json:"mirror_quotient_rows_equal"`
	MaximumIntegerRoundingDiscrepancy int    `json:"maximum_integer_rounding_discrepancy"`
	PreCenteredUnique                 bool   `json:"pre_rescale_centered_unique"`
	PostCenteredUnique                bool   `json:"post_rescale_centered_unique"`
	ScaleEqual                        bool   `json:"scale_equal"`
	LevelDecrementOne                 bool   `json:"level_decrement_one"`
	FirstMismatchIndex                int    `json:"first_mismatch_index,omitempty"`
	FirstExpected                     string `json:"first_expected,omitempty"`
	FirstMirror                       string `json:"first_mirror,omitempty"`
}

type nativeQ01LocalComparison struct {
	Branch                       string          `json:"branch"`
	Power                        int             `json:"power"`
	OldBalanced                  *PSGlobalMetric `json:"old_balanced_vs_canonical"`
	FullRNSMultiplyFirst         *PSGlobalMetric `json:"full_rns_multiply_first_vs_canonical"`
	NativeQ01MultiplyFirst       *PSGlobalMetric `json:"native_q01_multiply_first_vs_canonical"`
	NativeVsFullRNSMaxDifference float64         `json:"native_vs_full_rns_max_component_difference"`
	RowsEqual                    bool            `json:"post_rescale_rows_equal"`
}

type nativeQ01Result struct {
	SchemaVersion         string                     `json:"schema_version"`
	Timestamp             time.Time                  `json:"timestamp"`
	Primary               RepositoryMetadata         `json:"primary_repository"`
	Lattigo               RepositoryMetadata         `json:"lattigo_repository"`
	Environment           EnvironmentMetadata        `json:"environment"`
	Config                BootstrapConfig            `json:"config"`
	Parameters            ExperimentParameters       `json:"effective_parameters"`
	Workload              CorrectnessWorkload        `json:"workload"`
	Provenance            map[string]interface{}     `json:"provenance"`
	Q0Controls            map[string]interface{}     `json:"q0_controls"`
	DependencyGraph       []designDependencyRow      `json:"dependency_graph"`
	Q01Capacity           []designCapacity           `json:"q01_capacity_table"`
	Q012Capacity          []designCapacity           `json:"q012_capacity_table"`
	Checkpoints           []nativeQ01Checkpoint      `json:"native_vs_full_checkpoint_rows"`
	RescaleProof          []nativeQ01RescaleProof    `json:"rescale_quotient_oracle_table"`
	LocalComparisons      []nativeQ01LocalComparison `json:"local_power_comparisons"`
	System                map[string]interface{}     `json:"all_generated_native_system"`
	ParameterEvidence     map[string]interface{}     `json:"q0_q1_headroom_evidence"`
	Classification        string                     `json:"classification"`
	FirstRemainingBlocker string                     `json:"first_remaining_blocker"`
	RecommendedNextTarget string                     `json:"recommended_next_target"`
	Validation            map[string]interface{}     `json:"validation"`
}

type nativeQ01Stop struct{ Classification, Checkpoint string }

func (e *nativeQ01Stop) Error() string { return e.Classification + " at " + e.Checkpoint }

func nativeQ01Write(result nativeQ01Result, path string) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func nativeQ01FullAtLevel(params ckks.Parameters, source *rlwe.Ciphertext, level int) (*rlwe.Ciphertext, error) {
	if source == nil || !source.IsNTT || source.IsMontgomery || level < 0 || level > source.Level() {
		return nil, fmt.Errorf("full mirror level contraction precondition failed")
	}
	out := ckks.NewCiphertext(params, source.Degree(), level)
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery = source.IsNTT, source.IsMontgomery
	for component := range out.Value {
		for limb := 0; limb <= level; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func nativeQ01ContractFull(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || !source.IsNTT || source.Level() < 0 {
		return nil, fmt.Errorf("cannot contract full mirror")
	}
	out := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, true, source.Scale
	for component := range source.Value {
		for limb := 0; limb <= source.Level() && limb < 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
			params.RingQ().SubRings[limb].MForm(out.Value[component].Coeffs[limb], out.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func nativeQ01RowsEqual(params ckks.Parameters, a, b *rlwe.Ciphertext) bool {
	if a == nil || b == nil || a.Level() != b.Level() || a.Degree() != b.Degree() {
		return false
	}
	for component := 0; component <= a.Degree(); component++ {
		for limb := 0; limb <= a.Level() && limb < 2; limb++ {
			arow := append([]uint64(nil), a.Value[component].Coeffs[limb]...)
			brow := append([]uint64(nil), b.Value[component].Coeffs[limb]...)
			if a.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(arow, arow)
			}
			if b.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(brow, brow)
			}
			if !finalizationRowComparison(component, limb, arow, brow).Equal {
				return false
			}
		}
	}
	return true
}

func nativeQ01MetadataEqual(a, b *rlwe.Ciphertext) bool {
	return a != nil && b != nil && a.Level() == b.Level() && a.Degree() == b.Degree() && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && a.Scale.Equal(b.Scale)
}

func nativeQ01DecodedDifference(params ckks.Parameters, a, b *rlwe.Ciphertext) float64 {
	av, errA := psGlobalDecode(params, a)
	bv, errB := psGlobalDecode(params, b)
	if errA != nil || errB != nil || len(av) != len(bv) {
		return 0
	}
	max := 0.0
	for i := range av {
		d := math.Hypot(real(av[i]-bv[i]), imag(av[i]-bv[i]))
		if d > max {
			max = d
		}
	}
	return max
}

func nativeQ01CheckpointFor(params ckks.Parameters, branch string, power int, stage string, native, mirror *rlwe.Ciphertext) (nativeQ01Checkpoint, error) {
	contracted, err := nativeQ01ContractFull(params, mirror)
	if err != nil {
		return nativeQ01Checkpoint{}, err
	}
	capacity := (*designCapacity)(nil)
	if mirror.Level() >= 1 {
		values, e := designQ01Values(params, mirror)
		if e != nil {
			return nativeQ01Checkpoint{}, e
		}
		row := designCapacityRow(params, fmt.Sprintf("%s.T%d.%s", branch, power, stage), "q01", values)
		capacity = &row
	}
	return nativeQ01Checkpoint{Branch: branch, Power: power, Stage: stage, NativeMetadata: designStateText(native), MirrorMetadata: designStateText(contracted), RowsEqual: nativeQ01RowsEqual(params, native, contracted), MetadataEqual: nativeQ01MetadataEqual(native, contracted), Capacity: capacity, DecodedMaxDifference: nativeQ01DecodedDifference(params, native, contracted)}, nil
}

func nativeQ01FullAlign(params ckks.Parameters, out, sub *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	if out == nil || sub == nil {
		return nil, nil, fmt.Errorf("full recurrence alignment nil operand")
	}
	level := minInt(out.Level(), sub.Level())
	o, err := nativeQ01FullAtLevel(params, out, level)
	if err != nil {
		return nil, nil, err
	}
	s, err := nativeQ01FullAtLevel(params, sub, level)
	if err != nil {
		return nil, nil, err
	}
	if o.Scale.Equal(s.Scale) {
		return o, s, nil
	}
	if o.Scale.Cmp(s.Scale) > 0 {
		s = normalizedFullIntegerMultiply(params, s, o.Scale.Div(s.Scale).BigInt())
		s.Scale = o.Scale
		return o, s, nil
	}
	o = normalizedFullIntegerMultiply(params, o, s.Scale.Div(o.Scale).BigInt())
	o.Scale = s.Scale
	return o, s, nil
}

func nativeQ01NativeAlign(params ckks.Parameters, eval *fastckks.Evaluator, out, sub *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	level := minInt(out.Level(), sub.Level())
	if out.Scale.Equal(sub.Scale) {
		if out.Level() != level {
			source := out
			out = fastckks.NewCiphertext(params, source.Degree(), level)
			psGlobalCopyMaintained(params, source, out)
		}
		return out, sub, nil
	}
	if out.Scale.Cmp(sub.Scale) > 0 {
		scratch := fastckks.NewCiphertext(params, sub.Degree(), level)
		psGlobalCopyMaintained(params, sub, scratch)
		if err := eval.MulIntegerMaintained(scratch, out.Scale.Div(sub.Scale).BigInt(), scratch); err != nil {
			return nil, nil, err
		}
		scratch.Scale = out.Scale
		return out, scratch, nil
	}
	scratch := fastckks.NewCiphertext(params, out.Degree(), level)
	psGlobalCopyMaintained(params, out, scratch)
	if err := eval.MulIntegerMaintained(scratch, sub.Scale.Div(out.Scale).BigInt(), scratch); err != nil {
		return nil, nil, err
	}
	scratch.Scale = sub.Scale
	return scratch, sub, nil
}

func nativeQ01FullPower(params ckks.Parameters, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext) (designPowerTrace, error) {
	level := minInt(left.Level(), right.Level())
	l, err := designFullFromFastAtLevel(params, left, level)
	if err != nil {
		return designPowerTrace{}, err
	}
	r, err := designFullFromFastAtLevel(params, right, level)
	if err != nil {
		return designPowerTrace{}, err
	}
	trace := designPowerTrace{Power: n, CommonLevel: level, DivisorBits: new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus).BitLen(), ParentLeft: l, ParentRight: r, Factors: "none; native q01 oracle"}
	trace.RawProduct, err = designFullMulC0(params, l, r)
	if err != nil {
		return trace, err
	}
	trace.Doubled = normalizedFullIntegerMultiply(params, trace.RawProduct, big.NewInt(2))
	trace.Doubled.Scale = trace.RawProduct.Scale
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		trace.Aligned = trace.Doubled.CopyNew()
		trace.Corrected = normalizedFullSubtractConstant(params, trace.Doubled, 1)
	} else {
		sub, e := designFullFromFastAtLevel(params, powers[c], trace.Doubled.Level())
		if e != nil {
			return trace, e
		}
		trace.Aligned, trace.AlignedSub, err = nativeQ01FullAlign(params, trace.Doubled, sub)
		if err != nil {
			return trace, err
		}
		trace.Corrected = trace.Aligned.CopyNew()
		params.RingQ().AtLevel(trace.Corrected.Level()).Sub(trace.Corrected.Value[0], trace.AlignedSub.Value[0], trace.Corrected.Value[0])
	}
	trace.Completed, err = normalizedFullRescale(params, trace.Corrected)
	return trace, err
}

func nativeQ01CapacityRows(params ckks.Parameters, trace designPowerTrace, branch string) ([]designCapacity, []designCapacity, error) {
	rows := []struct {
		name string
		ct   *rlwe.Ciphertext
	}{{"raw_product", trace.RawProduct}, {"doubled", trace.Doubled}}
	if trace.Aligned != nil {
		rows = append(rows, struct {
			name string
			ct   *rlwe.Ciphertext
		}{"recurrence_aligned", trace.Aligned})
	}
	rows = append(rows, struct {
		name string
		ct   *rlwe.Ciphertext
	}{"recurrence_corrected", trace.Corrected})
	q01, q012 := []designCapacity{}, []designCapacity{}
	for _, row := range rows {
		name := fmt.Sprintf("%s.T%d.%s", branch, trace.Power, row.name)
		values, err := designQ01Values(params, row.ct)
		if err != nil {
			return nil, nil, err
		}
		q01 = append(q01, designCapacityRow(params, name, "q01", values))
		if row.ct.Level() >= 2 {
			values, err = designQ012Values(params, row.ct)
			if err != nil {
				return nil, nil, err
			}
			q012 = append(q012, designCapacityRow(params, name, "q012", values))
		}
	}
	return q01, q012, nil
}

func nativeQ01StopWrite(result *nativeQ01Result, stop *nativeQ01Stop, outPath string) error {
	result.Classification = stop.Classification
	result.FirstRemainingBlocker = stop.Checkpoint
	result.RecommendedNextTarget = "localize the first native q0/q1 blocker before production integration"
	result.Validation["q01_capacity_checked_before_native_operations"] = true
	result.Validation["no_unsafe_q01_operation_executed"] = true
	result.Validation["native_candidate_stopped_at_first_hard_blocker"] = true
	result.Validation["rescale_oracle_match"] = stop.Classification != "logn13_generated_power_native_q01_rescale_mismatch"
	return nativeQ01Write(*result, outPath)
}

func nativeQ01RescaleProofFor(params ckks.Parameters, branch string, n int, trace designPowerTrace, native *rlwe.Ciphertext) (nativeQ01RescaleProof, error) {
	before, err := postMod1S2CRecoverQ01(params, trace.Corrected)
	if err != nil {
		return nativeQ01RescaleProof{}, err
	}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[trace.Corrected.Level()].Modulus)
	match, nativeMatch, maxDiff := true, true, 0
	firstIndex := -1
	firstExpected, firstMirror := "", ""
	postCenteredUnique := true
	if native.Level() >= 1 {
		after, e := postMod1S2CRecoverQ01(params, trace.Completed)
		if e != nil {
			return nativeQ01RescaleProof{}, e
		}
		nativeAfter, e := postMod1S2CRecoverQ01(params, native)
		if e != nil {
			return nativeQ01RescaleProof{}, e
		}
		for component := range before {
			if component >= len(after) || component >= len(nativeAfter) || len(before[component]) != len(after[component]) || len(before[component]) != len(nativeAfter[component]) {
				match = false
				nativeMatch = false
				continue
			}
			for i, value := range before[component] {
				expected := fix001P3LocalQ2Round(value, divisor)
				diff := new(big.Int).Sub(expected, after[component][i])
				if diff.Sign() < 0 {
					diff.Neg(diff)
				}
				if diff.Sign() != 0 {
					match = false
					if firstIndex < 0 {
						firstIndex = i
						firstExpected, firstMirror = expected.String(), after[component][i].String()
					}
					if diff.IsInt64() && int(diff.Int64()) > maxDiff {
						maxDiff = int(diff.Int64())
					}
				}
				nativeDiff := new(big.Int).Sub(expected, nativeAfter[component][i])
				if nativeDiff.Sign() < 0 {
					nativeDiff.Neg(nativeDiff)
				}
				if nativeDiff.Sign() != 0 {
					nativeMatch = false
				}
			}
		}
		if values, e := designQ01Values(params, trace.Completed); e == nil {
			postCenteredUnique = designCapacityRow(params, fmt.Sprintf("%s.T%d.post_rescale", branch, n), "q01", values).CenteredUnique
		}
		match = match && nativeQ01RowsEqual(params, native, mustNativeQ01Contract(params, trace.Completed))
	} else {
		q0 := params.RingQ().SubRings[0]
		for component := range before {
			expected := make([]*big.Int, len(before[component]))
			for i, value := range before[component] {
				expected[i] = fix001P3LocalQ2Round(value, divisor)
			}
			poly := postMod1S2CLiftSigned(params, expected, 0)
			actual := append([]uint64(nil), native.Value[component].Coeffs[0]...)
			if native.IsMontgomery {
				q0.IMForm(actual, actual)
			}
			if !finalizationRowComparison(component, 0, poly.Coeffs[0], actual).Equal {
				match = false
				nativeMatch = false
			}
		}
	}
	return nativeQ01RescaleProof{Branch: branch, Power: n, Divisor: divisor.String(), InputLevel: trace.Corrected.Level(), OutputLevel: native.Level(), RoundedQuotientRowsEqual: nativeMatch && match && nativeQ01RowsEqual(params, native, mustNativeQ01Contract(params, trace.Completed)), NativeQuotientRowsEqual: nativeMatch, MirrorQuotientRowsEqual: match, MaximumIntegerRoundingDiscrepancy: maxDiff, PreCenteredUnique: true, PostCenteredUnique: postCenteredUnique, ScaleEqual: native.Scale.Equal(trace.Completed.Scale), LevelDecrementOne: native.Level() == trace.Corrected.Level()-params.LevelsConsumedPerRescaling(), FirstMismatchIndex: firstIndex, FirstExpected: firstExpected, FirstMirror: firstMirror}, nil
}

func mustNativeQ01Contract(params ckks.Parameters, source *rlwe.Ciphertext) *rlwe.Ciphertext {
	out, _ := nativeQ01ContractFull(params, source)
	return out
}

func nativeQ01PowerOperation(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch string, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext, trace designPowerTrace, result *nativeQ01Result) (*rlwe.Ciphertext, error) {
	for _, row := range []struct {
		stage string
		ct    *rlwe.Ciphertext
	}{{"parent_left", left}, {"parent_right", right}} {
		var ref *rlwe.Ciphertext
		if row.stage == "parent_left" {
			ref = trace.ParentLeft
		} else {
			ref = trace.ParentRight
		}
		cp, err := nativeQ01CheckpointFor(params, branch, n, row.stage, row.ct, ref)
		if err != nil {
			return nil, err
		}
		result.Checkpoints = append(result.Checkpoints, cp)
	}
	q01, q012, err := nativeQ01CapacityRows(params, trace, branch)
	if err != nil {
		return nil, err
	}
	result.Q01Capacity = append(result.Q01Capacity, q01...)
	result.Q012Capacity = append(result.Q012Capacity, q012...)
	for _, row := range q01 {
		if !row.CenteredUnique {
			return nil, &nativeQ01Stop{"logn13_generated_power_native_q01_capacity_regression", row.Checkpoint}
		}
	}
	out := fastckks.NewCiphertext(params, 1, trace.CommonLevel)
	*out.MetaData = *left.MetaData
	out.IsNTT, out.IsMontgomery = left.IsNTT, left.IsMontgomery
	if err := eval.FastCKKS.MulRelin(left, right, out); err != nil {
		return nil, err
	}
	cp, err := nativeQ01CheckpointFor(params, branch, n, "raw_product", out, trace.RawProduct)
	if err != nil {
		return nil, err
	}
	result.Checkpoints = append(result.Checkpoints, cp)
	if !q01[0].CenteredUnique {
		return nil, &nativeQ01Stop{"logn13_generated_power_native_q01_capacity_regression", q01[0].Checkpoint}
	}
	if err := eval.FastCKKS.Add(out, out, out); err != nil {
		return nil, err
	}
	cp, err = nativeQ01CheckpointFor(params, branch, n, "doubled", out, trace.Doubled)
	if err != nil {
		return nil, err
	}
	result.Checkpoints = append(result.Checkpoints, cp)
	if !q01[1].CenteredUnique {
		return nil, &nativeQ01Stop{"logn13_generated_power_native_q01_capacity_regression", q01[1].Checkpoint}
	}
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		if err := eval.FastCKKS.Add(out, -1, out); err != nil {
			return nil, err
		}
	} else {
		aligned, sub, e := nativeQ01NativeAlign(params, eval.FastCKKS, out, powers[c])
		if e != nil {
			return nil, e
		}
		alignedIndex := 2
		if trace.Aligned != nil && len(q01) > 2 {
			alignedIndex = 2
		}
		if !q01[alignedIndex].CenteredUnique {
			return nil, &nativeQ01Stop{"logn13_generated_power_native_q01_capacity_regression", q01[alignedIndex].Checkpoint}
		}
		cp, e = nativeQ01CheckpointFor(params, branch, n, "recurrence_aligned", aligned, trace.Aligned)
		if e != nil {
			return nil, e
		}
		result.Checkpoints = append(result.Checkpoints, cp)
		if err := eval.FastCKKS.Sub(aligned, sub, aligned); err != nil {
			return nil, err
		}
		out = aligned
	}
	cp, err = nativeQ01CheckpointFor(params, branch, n, "recurrence_corrected", out, trace.Corrected)
	if err != nil {
		return nil, err
	}
	result.Checkpoints = append(result.Checkpoints, cp)
	last := q01[len(q01)-1]
	if !last.CenteredUnique {
		return nil, &nativeQ01Stop{"logn13_generated_power_native_q01_capacity_regression", last.Checkpoint}
	}
	if err := eval.FastCKKS.Rescale(out, out); err != nil {
		return nil, err
	}
	cp, err = nativeQ01CheckpointFor(params, branch, n, "post_rescale", out, trace.Completed)
	if err != nil {
		return nil, err
	}
	result.Checkpoints = append(result.Checkpoints, cp)
	proof, err := nativeQ01RescaleProofFor(params, branch, n, trace, out)
	if err != nil {
		return nil, err
	}
	result.RescaleProof = append(result.RescaleProof, proof)
	if !proof.RoundedQuotientRowsEqual {
		return nil, &nativeQ01Stop{"logn13_generated_power_native_q01_rescale_mismatch", fmt.Sprintf("%s.T%d.post_rescale", branch, n)}
	}
	return out, nil
}

func nativeQ01Generate(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch string, input *rlwe.Ciphertext, z []complex128, result *nativeQ01Result) (map[int]*rlwe.Ciphertext, map[int][]complex128, error) {
	powers := map[int]*rlwe.Ciphertext{1: input.CopyNew()}
	values := map[int][]complex128{1: append([]complex128(nil), z...)}
	var generate func(int) error
	generate = func(n int) error {
		if powers[n] != nil {
			return nil
		}
		a, b := commonpolynomial.SplitDegree(n)
		if err := generate(a); err != nil {
			return err
		}
		if err := generate(b); err != nil {
			return err
		}
		trace, err := nativeQ01FullPower(params, powers[a], powers[b], n, powers)
		if err != nil {
			return err
		}
		out, err := nativeQ01PowerOperation(params, eval, branch, powers[a], powers[b], n, powers, trace, result)
		if err != nil {
			return err
		}
		powers[n] = out
		values[n], err = psGlobalDecode(params, out)
		return err
	}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		if err := generate(n); err != nil {
			return powers, values, err
		}
	}
	return powers, values, nil
}

func runFIX001P3DesignLogN13GeneratedPowerNativeQ01MultiplyFirst(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	effectiveCfg := q056BuildConfig(cfg, 56)
	profile, err := q056PrepareProfile(effectiveCfg)
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := nativeQ01Result{SchemaVersion: "fix-001-p3-design-logn13-generated-power-native-q01-multiply-first-feasibility.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: effectiveCfg, Parameters: parameterMetadata(profile.Residual, profile.BTP), Workload: CorrectnessWorkload{Identifier: "accepted-q056-ps-input.v1", Formula: "accepted evalModMatchedC2SInputs with fixed D0 stack", LogicalSlots: profile.Residual.MaxSlots()}, Provenance: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3NativeQ01PrimaryBase, "secondary_required_commit": requiredFIX001P3NativeQ01Secondary, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean}, Q0Controls: map[string]interface{}{}, System: map[string]interface{}{}, ParameterEvidence: map[string]interface{}{}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "q0_q1_q2_profile": "56/39/40", "common_plan_scale_2^92": true, "no_q_parameter_change": true, "no_secondary_production_changes": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "native_fast_operations": "MulRelin/Add/Rescale", "q2_native_arithmetic": false}}
	if cfg.LogN != 13 || secondaryCommit != requiredFIX001P3NativeQ01Secondary || !secondaryClean || gitOutput(primaryRoot, "merge-base", requiredFIX001P3NativeQ01PrimaryBase, "HEAD") != requiredFIX001P3NativeQ01PrimaryBase {
		result.Classification = "logn13_generated_power_native_q01_precondition_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary provenance or LogN13 precondition"
		return nativeQ01Write(result, outPath)
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
	// Q0: reproduce the accepted D0, old balanced, and full-RNS controls.
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
	if d0.Metrics.PublicLike == nil || fullCase.Metrics.PublicLike == nil || math.Abs(d0.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance || math.Abs(fullCase.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance {
		result.Classification = "logn13_generated_power_native_q01_precondition_mismatch"
		result.FirstRemainingBlocker = "Q0 accepted control mismatch"
		return nativeQ01Write(result, outPath)
	}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		a, b := commonpolynomial.SplitDegree(n)
		left, right, out := realContext.Branch.Base.ActualPowerMap[a], realContext.Branch.Base.ActualPowerMap[b], realContext.Branch.Base.ActualPowerMap[n]
		level := minInt(left.Level(), right.Level())
		fL, fR, e := designBalancedFactors(params, level)
		if e != nil {
			return e
		}
		result.DependencyGraph = append(result.DependencyGraph, designDependencyRow{Power: n, Left: a, Right: b, Lazy: out.Degree() > 1, Balanced: true, CommonLevel: level, DivisorBits: new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus).BitLen(), Factors: fL.String() + "*" + fR.String(), LeftMetadata: designStateText(left), RightMetadata: designStateText(right), OutputMetadata: designStateText(out), OperationLineage: "commonpolynomial.SplitDegree -> native Fast q01 MulRelin -> Add -> recurrence correction -> one Rescale"})
	}
	result.ParameterEvidence["q0_bits"] = new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus).BitLen()
	result.ParameterEvidence["q1_bits"] = new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus).BitLen()
	result.ParameterEvidence["q01_bits"] = new(big.Int).Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)).BitLen()
	nativeR, nativeRV, err := nativeQ01Generate(params, profile.Fast, "real", realE2, realZ, &result)
	if err != nil {
		if stop, ok := err.(*nativeQ01Stop); ok {
			return nativeQ01StopWrite(&result, stop, outPath)
		}
		return err
	}
	nativeI, nativeIV, err := nativeQ01Generate(params, profile.Fast, "imag", imagE2, imagZ, &result)
	if err != nil {
		if stop, ok := err.(*nativeQ01Stop); ok {
			return nativeQ01StopWrite(&result, stop, outPath)
		}
		return err
	}
	for _, branch := range []struct {
		name              string
		native, full, old map[int]*rlwe.Ciphertext
		expected          map[int][]complex128
	}{{"real", nativeR, fullR, realContext.Branch.Base.ActualPowerMap, realContext.Branch.Base.PowerExpected}, {"imag", nativeI, fullI, imagContext.Branch.Base.ActualPowerMap, imagContext.Branch.Base.PowerExpected}} {
		for _, n := range []int{2, 3, 4, 6, 8, 16} {
			oldM, _ := designMetricFor(params, branch.old[n], branch.expected[n])
			fullM, _ := designMetricFor(params, branch.full[n], branch.expected[n])
			nativeM, _ := designMetricFor(params, branch.native[n], branch.expected[n])
			fullVals, _ := psGlobalDecode(params, branch.full[n])
			nativeVals, _ := psGlobalDecode(params, branch.native[n])
			delta := 0.0
			for i := range nativeVals {
				if d := math.Hypot(real(nativeVals[i]-fullVals[i]), imag(nativeVals[i]-fullVals[i])); d > delta {
					delta = d
				}
			}
			result.LocalComparisons = append(result.LocalComparisons, nativeQ01LocalComparison{Branch: branch.name, Power: n, OldBalanced: oldM, FullRNSMultiplyFirst: fullM, NativeQ01MultiplyFirst: nativeM, NativeVsFullRNSMaxDifference: delta, RowsEqual: nativeQ01RowsEqual(params, branch.native[n], branch.full[n])})
		}
	}
	system, err := generatedPowerReentryRunCase(profile, realContext, imagContext, nativeR, nativeI, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, nativeRV, nativeIV, "M-all-generated-native-q01", []int{2, 3, 4, 6, 8, 16})
	if err != nil {
		return err
	}
	result.System["name"] = system.Name
	result.System["metrics"] = system.Metrics
	result.System["pass"] = system.Metrics.Pass
	result.System["native_full_rns_public_like_difference"] = func() float64 {
		if system.Metrics.PublicLike == nil || fullCase.Metrics.PublicLike == nil {
			return 0
		}
		return math.Abs(system.Metrics.PublicLike.MaxComponent - fullCase.Metrics.PublicLike.MaxComponent)
	}()
	maxRatio := 0.0
	minBits := 0
	checkpoint := ""
	allRows, allMeta, allRescale := true, true, true
	for _, row := range result.Q01Capacity {
		if row.Ratio > maxRatio {
			maxRatio, checkpoint = row.Ratio, row.Checkpoint
		}
		if row.MinimumBits > minBits {
			minBits = row.MinimumBits
		}
	}
	result.ParameterEvidence["maximum_native_q01_multiply_first_ratio"] = maxRatio
	result.ParameterEvidence["headroom_fraction"] = 1 - maxRatio
	result.ParameterEvidence["minimum_required_q01_bits"] = minBits
	result.ParameterEvidence["minimum_checkpoint"] = checkpoint
	for _, cp := range result.Checkpoints {
		allRows = allRows && cp.RowsEqual
		allMeta = allMeta && cp.MetadataEqual
	}
	for _, p := range result.RescaleProof {
		allRescale = allRescale && p.RoundedQuotientRowsEqual && p.PreCenteredUnique && p.PostCenteredUnique && p.ScaleEqual && p.LevelDecrementOne
	}
	result.Validation["q01_capacity_checked_before_native_operations"] = true
	result.Validation["q01_capacity_safe"] = allRows
	result.Validation["native_full_rns_rows_match"] = allRows
	result.Validation["metadata_contracts_match"] = allMeta
	result.Validation["rescale_oracle_match"] = allRescale
	result.Validation["d0_exact_control"] = true
	result.Validation["full_rns_multiply_first_control"] = true
	result.Validation["no_unsafe_q01_operation_executed"] = true
	if !allRows || !allMeta {
		result.Classification = "logn13_generated_power_native_q01_operation_mismatch"
		result.FirstRemainingBlocker = "first native/full-RNS checkpoint row or metadata divergence"
	} else if !allRescale {
		result.Classification = "logn13_generated_power_native_q01_rescale_mismatch"
		result.FirstRemainingBlocker = "native Fast Rescale differs from independent q01 rounded quotient"
	} else if system.Metrics.PublicLike == nil || !system.Metrics.PublicLike.Pass {
		result.Classification = "logn13_generated_power_native_q01_numeric_insufficient"
		result.FirstRemainingBlocker = "all-generated native q01 system public-like threshold"
	} else {
		result.Classification = "logn13_generated_power_native_q01_multiply_first_system_sufficient"
		result.FirstRemainingBlocker = "q0/q1 remains sufficient but tight; no production blocker in this diagnostic"
	}
	if result.Classification == "logn13_generated_power_native_q01_multiply_first_system_sufficient" {
		result.RecommendedNextTarget = "production integration of native q01 generated-power multiply-first schedule"
	} else {
		result.RecommendedNextTarget = "localize the first native q01 divergent primitive before production integration"
	}
	_ = realWork
	_ = imagWork
	return nativeQ01Write(result, outPath)
}
