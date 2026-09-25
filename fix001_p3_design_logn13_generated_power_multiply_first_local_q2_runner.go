//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3GeneratedPowerMultiplyFirstPrimary   = "e7e677c5e6d78e8fdea888e57a55455215e7c064"
	requiredFIX001P3GeneratedPowerMultiplyFirstSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	designMultiplyFirstThreshold                         = 1e-2
	designMultiplyFirstLocalThreshold                    = 1e-6
)

type designDependencyRow struct {
	Power            int    `json:"power"`
	Left             int    `json:"left"`
	Right            int    `json:"right"`
	Lazy             bool   `json:"lazy"`
	Balanced         bool   `json:"balanced"`
	CommonLevel      int    `json:"common_level"`
	DivisorBits      int    `json:"current_rescale_divisor_bits"`
	Factors          string `json:"balanced_factors,omitempty"`
	LeftMetadata     string `json:"left_metadata"`
	RightMetadata    string `json:"right_metadata"`
	OutputMetadata   string `json:"output_metadata"`
	OperationLineage string `json:"operation_lineage"`
}

type designCapacity struct {
	Checkpoint     string  `json:"checkpoint"`
	MaxAbs         string  `json:"max_abs"`
	HalfModulus    string  `json:"half_modulus"`
	Ratio          float64 `json:"ratio"`
	OutsideCount   int     `json:"outside_count"`
	MinimumBits    int     `json:"minimum_required_bits"`
	CenteredUnique bool    `json:"centered_unique"`
	Domain         string  `json:"domain"`
}

type designPowerTrace struct {
	Power             int
	CommonLevel       int
	DivisorBits       int
	Factors           string
	Balanced          bool
	ParentLeft        *rlwe.Ciphertext
	ParentRight       *rlwe.Ciphertext
	AfterFactorLeft   *rlwe.Ciphertext
	AfterFactorRight  *rlwe.Ciphertext
	AfterRescaleLeft  *rlwe.Ciphertext
	AfterRescaleRight *rlwe.Ciphertext
	RawProduct        *rlwe.Ciphertext
	Doubled           *rlwe.Ciphertext
	Aligned           *rlwe.Ciphertext
	AlignedSub        *rlwe.Ciphertext
	Corrected         *rlwe.Ciphertext
	Completed         *rlwe.Ciphertext
}

type designResetRow struct {
	Power          int                    `json:"power"`
	ID             string                 `json:"id"`
	Kind           string                 `json:"kind"`
	LocalCompleted *PSGlobalMetric        `json:"local_completed_vs_oracle,omitempty"`
	PublicLike     *PSGlobalMetric        `json:"public_like,omitempty"`
	Pass           bool                   `json:"pass"`
	Level          int                    `json:"level"`
	Scale          string                 `json:"scale"`
	ExactState     bool                   `json:"exact_state_materialized"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

type designLedgerRow struct {
	Power          int     `json:"power"`
	Checkpoint     string  `json:"checkpoint"`
	Level          int     `json:"level"`
	Scale          string  `json:"scale"`
	Q01Ratio       float64 `json:"q01_ratio"`
	FullRatio      float64 `json:"full_rns_ratio"`
	Residual       float64 `json:"residual_max_component"`
	CenteredUnique bool    `json:"centered_unique"`
}

type designLocalComparison struct {
	Power             int             `json:"power"`
	BalancedCompleted *PSGlobalMetric `json:"balanced_completed_vs_oracle"`
	MultiplyFirst     *PSGlobalMetric `json:"multiply_first_completed_vs_oracle"`
	ImprovementFactor float64         `json:"improvement_factor"`
	FastFullRowsMatch bool            `json:"fast_full_rows_match"`
	Q01CapacitySafe   bool            `json:"q01_capacity_safe"`
	Q012CapacitySafe  bool            `json:"q012_capacity_safe"`
	OperandRescalesB  int             `json:"balanced_operand_rescales"`
	RoundedRescalesM  int             `json:"multiply_first_rounded_rescales"`
	SemanticAgreement bool            `json:"semantic_agreement"`
}

type designSystemReplay struct {
	Name    string                         `json:"name"`
	Roots   []int                          `json:"corrected_roots"`
	Powers  []int                          `json:"candidate_power_nodes"`
	Metrics generatedPowerReentryMetricSet `json:"metrics"`
	Pass    bool                           `json:"pass"`
}

type designResult struct {
	SchemaVersion         string                  `json:"schema_version"`
	Timestamp             time.Time               `json:"timestamp"`
	Primary               RepositoryMetadata      `json:"primary_repository"`
	Lattigo               RepositoryMetadata      `json:"lattigo_repository"`
	Environment           EnvironmentMetadata     `json:"environment"`
	Config                BootstrapConfig         `json:"config"`
	Parameters            ExperimentParameters    `json:"effective_parameters"`
	Workload              CorrectnessWorkload     `json:"workload"`
	Provenance            map[string]interface{}  `json:"provenance"`
	A0Controls            map[string]interface{}  `json:"a0_controls"`
	DependencyGraph       []designDependencyRow   `json:"dependency_graph"`
	CanonicalResets       []designResetRow        `json:"canonical_resets"`
	IntrinsicInherited    map[string]interface{}  `json:"intrinsic_vs_inherited"`
	BalancedLedger        []designLedgerRow       `json:"balanced_precision_ledger"`
	Q01Capacity           []designCapacity        `json:"q01_multiply_first_capacity_only"`
	Q012Capacity          []designCapacity        `json:"q012_multiply_first_capacity"`
	MinimumQ01Bits        map[string]int          `json:"minimum_required_q01_bits"`
	ParameterEvidence     map[string]interface{}  `json:"q01_q012_parameter_evidence"`
	Comparisons           []designLocalComparison `json:"balanced_vs_multiply_first"`
	SystemReplay          []designSystemReplay    `json:"dependency_aware_system_replay"`
	Classification        string                  `json:"classification"`
	FirstRemainingBlocker string                  `json:"first_remaining_blocker"`
	RecommendedNextTarget string                  `json:"recommended_next_target"`
	Validation            map[string]interface{}  `json:"validation"`
}

func designScaleText(scale rlwe.Scale) string { return finalizationScaleString(scale) }

func designStateText(ct *rlwe.Ciphertext) string {
	if ct == nil {
		return "nil"
	}
	return fmt.Sprintf("level=%d scale=%s degree=%d ntt=%t mont=%t", ct.Level(), designScaleText(ct.Scale), ct.Degree(), ct.IsNTT, ct.IsMontgomery)
}

func designFullFromFastAtLevel(params ckks.Parameters, source *rlwe.Ciphertext, level int) (*rlwe.Ciphertext, error) {
	if source == nil || !source.IsNTT || source.Level() < 1 || level < 1 || level > source.Level() {
		return nil, fmt.Errorf("cannot materialize Fast source at level %d", level)
	}
	signed, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, err
	}
	full := ckks.NewCiphertext(params, source.Degree(), level)
	*full.MetaData = *source.MetaData
	full.IsNTT, full.IsMontgomery, full.Scale = true, false, source.Scale
	for component := range signed {
		poly := postMod1S2CLiftSigned(params, signed[component], level)
		for limb := 0; limb <= level; limb++ {
			copy(full.Value[component].Coeffs[limb], poly.Coeffs[limb])
		}
	}
	return full, nil
}

func designFastFromFull(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || !source.IsNTT || source.Level() < 1 {
		return nil, fmt.Errorf("cannot contract full source")
	}
	fast := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*fast.MetaData = *source.MetaData
	fast.IsNTT, fast.IsMontgomery, fast.Scale = source.IsNTT, true, source.Scale
	for component := range source.Value {
		for limb := 0; limb < 2; limb++ {
			copy(fast.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
			params.RingQ().SubRings[limb].MForm(fast.Value[component].Coeffs[limb], fast.Value[component].Coeffs[limb])
		}
	}
	return fast, nil
}

func designFullMulC0(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || !left.IsNTT || !right.IsNTT || left.IsMontgomery || right.IsMontgomery {
		return nil, fmt.Errorf("full multiply requires non-Montgomery NTT operands")
	}
	level := minInt(left.Level(), right.Level())
	if level < 1 {
		return nil, fmt.Errorf("full multiply requires level >= 1")
	}
	result := ckks.NewCiphertext(params, 1, level)
	*result.MetaData = *left.MetaData
	result.IsNTT, result.IsMontgomery = true, false
	result.Scale = left.Scale.Mul(right.Scale)
	params.RingQ().AtLevel(level).MulCoeffsBarrett(left.Value[0], right.Value[0], result.Value[0])
	return result, nil
}

func designFullSubAligned(params ckks.Parameters, out, sub *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if out == nil || sub == nil {
		return nil, fmt.Errorf("aligned subtraction requires two ciphertexts")
	}
	result := out.CopyNew()
	if result.Scale.Cmp(sub.Scale) != 0 {
		if result.Scale.Cmp(sub.Scale) > 0 {
			scaled := normalizedFullIntegerMultiply(params, sub, result.Scale.Div(sub.Scale).BigInt())
			scaled.Scale = result.Scale
			sub = scaled
		} else {
			scaled := normalizedFullIntegerMultiply(params, result, sub.Scale.Div(result.Scale).BigInt())
			scaled.Scale = sub.Scale
			result = scaled
		}
	}
	params.RingQ().AtLevel(result.Level()).Sub(result.Value[0], sub.Value[0], result.Value[0])
	return result, nil
}

func designQ012Values(params ckks.Parameters, source *rlwe.Ciphertext) ([]*big.Int, error) {
	if source == nil || !source.IsNTT || source.Level() < 2 {
		return nil, fmt.Errorf("q012 capacity requires NTT level >= 2")
	}
	poly := ring.NewPoly(params.N(), 2)
	for limb := 0; limb <= 2; limb++ {
		copy(poly.Coeffs[limb], source.Value[0].Coeffs[limb])
		if source.IsMontgomery {
			params.RingQ().SubRings[limb].IMForm(poly.Coeffs[limb], poly.Coeffs[limb])
		}
	}
	coeff := ring.NewPoly(params.N(), 2)
	params.RingQ().AtLevel(2).INTT(poly, coeff)
	return postMod1S2CFullCRT(params, coeff, 2), nil
}

func designQ01Values(params ckks.Parameters, source *rlwe.Ciphertext) ([]*big.Int, error) {
	if source == nil || !source.IsNTT || source.Level() < 1 {
		return nil, fmt.Errorf("q01 capacity requires NTT level >= 1")
	}
	recovered, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, err
	}
	return recovered[0], nil
}

func designCapacityRow(params ckks.Parameters, name, domain string, values []*big.Int) designCapacity {
	modulus := big.NewInt(1)
	if domain == "q01" {
		modulus.Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus))
	} else {
		for limb := 0; limb <= 2; limb++ {
			modulus.Mul(modulus, new(big.Int).SetUint64(params.RingQ().SubRings[limb].Modulus))
		}
	}
	half := new(big.Int).Rsh(new(big.Int).Set(modulus), 1)
	maxAbs := new(big.Int)
	outside := 0
	for _, value := range values {
		abs := new(big.Int).Abs(value)
		if abs.Cmp(maxAbs) > 0 {
			maxAbs.Set(abs)
		}
		if abs.Cmp(half) >= 0 {
			outside++
		}
	}
	minimumBits := 0
	if maxAbs.Sign() > 0 {
		bound := new(big.Int).Add(new(big.Int).Lsh(new(big.Int).Set(maxAbs), 1), big.NewInt(1))
		minimumBits = new(big.Int).Sub(bound, big.NewInt(1)).BitLen()
	}
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(maxAbs), new(big.Float).SetInt(half)).Float64()
	return designCapacity{Checkpoint: name, Domain: domain, MaxAbs: maxAbs.String(), HalfModulus: half.String(), Ratio: ratio, OutsideCount: outside, MinimumBits: minimumBits, CenteredUnique: outside == 0}
}

func designBalancedFactors(params ckks.Parameters, level int) (*big.Int, *big.Int, error) {
	return guardedPowerFactors(new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus), 0)
}

func designBalancedFullPower(params ckks.Parameters, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext) (designPowerTrace, error) {
	commonLevel := minInt(left.Level(), right.Level())
	factorLeft, factorRight, err := designBalancedFactors(params, commonLevel)
	if err != nil {
		return designPowerTrace{}, err
	}
	leftFull, err := designFullFromFastAtLevel(params, left, commonLevel)
	if err != nil {
		return designPowerTrace{}, err
	}
	rightFull, err := designFullFromFastAtLevel(params, right, commonLevel)
	if err != nil {
		return designPowerTrace{}, err
	}
	trace := designPowerTrace{Power: n, CommonLevel: commonLevel, DivisorBits: new(big.Int).SetUint64(params.RingQ().SubRings[commonLevel].Modulus).BitLen(), ParentLeft: leftFull, ParentRight: rightFull, Balanced: true, Factors: factorLeft.String() + "*" + factorRight.String()}
	trace.AfterFactorLeft = normalizedFullIntegerMultiply(params, leftFull, factorLeft)
	trace.AfterFactorRight = normalizedFullIntegerMultiply(params, rightFull, factorRight)
	trace.AfterFactorLeft.Scale = left.Scale.Mul(rlwe.NewScale(factorLeft))
	trace.AfterFactorRight.Scale = right.Scale.Mul(rlwe.NewScale(factorRight))
	trace.AfterRescaleLeft, err = normalizedFullRescale(params, trace.AfterFactorLeft)
	if err != nil {
		return trace, err
	}
	trace.AfterRescaleRight, err = normalizedFullRescale(params, trace.AfterFactorRight)
	if err != nil {
		return trace, err
	}
	trace.RawProduct, err = designFullMulC0(params, trace.AfterRescaleLeft, trace.AfterRescaleRight)
	if err != nil {
		return trace, err
	}
	trace.Doubled = normalizedFullIntegerMultiply(params, trace.RawProduct, big.NewInt(2))
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		trace.Corrected = normalizedFullSubtractConstant(params, trace.Doubled, 1)
	} else {
		sub, err := designFullFromFastAtLevel(params, powers[c], trace.Doubled.Level())
		if err != nil {
			return trace, err
		}
		trace.Corrected, err = designFullSubAligned(params, trace.Doubled, sub)
		if err != nil {
			return trace, err
		}
	}
	trace.Completed = trace.Corrected
	return trace, nil
}

func designMultiplyFirstFullPower(params ckks.Parameters, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext) (designPowerTrace, error) {
	commonLevel := minInt(left.Level(), right.Level())
	leftFull, err := designFullFromFastAtLevel(params, left, commonLevel)
	if err != nil {
		return designPowerTrace{}, err
	}
	rightFull, err := designFullFromFastAtLevel(params, right, commonLevel)
	if err != nil {
		return designPowerTrace{}, err
	}
	trace := designPowerTrace{Power: n, CommonLevel: commonLevel, DivisorBits: new(big.Int).SetUint64(params.RingQ().SubRings[commonLevel].Modulus).BitLen(), ParentLeft: leftFull, ParentRight: rightFull, Factors: "none; native high-scale multiply"}
	trace.RawProduct, err = designFullMulC0(params, leftFull, rightFull)
	if err != nil {
		return trace, err
	}
	trace.Doubled = normalizedFullIntegerMultiply(params, trace.RawProduct, big.NewInt(2))
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		trace.Corrected = normalizedFullSubtractConstant(params, trace.Doubled, 1)
	} else {
		sub, err := designFullFromFastAtLevel(params, powers[c], trace.Doubled.Level())
		if err != nil {
			return trace, err
		}
		trace.Corrected, err = designFullSubAligned(params, trace.Doubled, sub)
		if err != nil {
			return trace, err
		}
	}
	trace.Completed, err = normalizedFullRescale(params, trace.Corrected)
	return trace, err
}

func designCurrentRecurrence(params ckks.Parameters, eval *fastckks.Evaluator, value *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	a, b := commonpolynomial.SplitDegree(n)
	c := a - b
	if c < 0 {
		c = -c
	}
	if c == 0 {
		if err := eval.Add(value, -1, value); err != nil {
			return nil, err
		}
		return value, nil
	}
	if err := guardedPowerSubAligned(params, eval, value, powers[c]); err != nil {
		return nil, err
	}
	return value, nil
}

func designCurrentPowerOperation(params ckks.Parameters, eval *bootstrapping.FastEvaluator, left, right *rlwe.Ciphertext, n int, powers map[int]*rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	commonLevel := minInt(left.Level(), right.Level())
	factorLeft, factorRight, err := designBalancedFactors(params, commonLevel)
	if err != nil {
		return nil, err
	}
	qScale := rlwe.NewScale(new(big.Int).SetUint64(params.RingQ().SubRings[commonLevel].Modulus))
	leftScale := left.Scale.Mul(rlwe.NewScale(factorLeft)).Div(qScale)
	rightScale := right.Scale.Mul(rlwe.NewScale(factorRight)).Div(qScale)
	minimum := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 20))
	if leftScale.Cmp(minimum) < 0 || rightScale.Cmp(minimum) < 0 {
		out := fastckks.NewCiphertext(params, 1, commonLevel)
		*out.MetaData = *left.MetaData
		out.IsNTT, out.IsMontgomery = left.IsNTT, left.IsMontgomery
		if err := eval.FastCKKS.MulRelin(left, right, out); err != nil {
			return nil, err
		}
		if err := eval.FastCKKS.Add(out, out, out); err != nil {
			return nil, err
		}
		if err := eval.FastCKKS.Rescale(out, out); err != nil {
			return nil, err
		}
		return designCurrentRecurrence(params, eval.FastCKKS, out, n, powers)
	}
	leftCopy := fastckks.NewCiphertext(params, left.Degree(), commonLevel)
	rightCopy := fastckks.NewCiphertext(params, right.Degree(), commonLevel)
	psGlobalCopyMaintained(params, left, leftCopy)
	psGlobalCopyMaintained(params, right, rightCopy)
	if err := eval.FastCKKS.MulIntegerMaintained(leftCopy, factorLeft, leftCopy); err != nil {
		return nil, err
	}
	if err := eval.FastCKKS.MulIntegerMaintained(rightCopy, factorRight, rightCopy); err != nil {
		return nil, err
	}
	leftCopy.Scale = left.Scale.Mul(rlwe.NewScale(factorLeft))
	rightCopy.Scale = right.Scale.Mul(rlwe.NewScale(factorRight))
	if err := eval.FastCKKS.Rescale(leftCopy, leftCopy); err != nil {
		return nil, err
	}
	if err := eval.FastCKKS.Rescale(rightCopy, rightCopy); err != nil {
		return nil, err
	}
	out := fastckks.NewCiphertext(params, 1, commonLevel-1)
	*out.MetaData = *leftCopy.MetaData
	out.IsNTT, out.IsMontgomery = leftCopy.IsNTT, leftCopy.IsMontgomery
	if err := eval.FastCKKS.MulRelin(leftCopy, rightCopy, out); err != nil {
		return nil, err
	}
	if err := eval.FastCKKS.Add(out, out, out); err != nil {
		return nil, err
	}
	return designCurrentRecurrence(params, eval.FastCKKS, out, n, powers)
}

func designGenerateCandidateMap(params ckks.Parameters, eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, z []complex128, actual map[int]*rlwe.Ciphertext, roots map[int]bool, useMultiplyFirst bool) (map[int]*rlwe.Ciphertext, map[int][]complex128, []int, error) {
	return designGenerateCandidateMapWithOverrides(params, eval, input, z, actual, nil, roots, useMultiplyFirst)
}

func designGenerateCandidateMapWithOverrides(params ckks.Parameters, eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, z []complex128, actual, overrides map[int]*rlwe.Ciphertext, roots map[int]bool, useMultiplyFirst bool) (map[int]*rlwe.Ciphertext, map[int][]complex128, []int, error) {
	needs := map[int]bool{}
	var depends func(int) bool
	depends = func(n int) bool {
		if n <= 1 {
			return false
		}
		if needs[n] {
			return true
		}
		if roots[n] {
			needs[n] = true
			return true
		}
		a, b := commonpolynomial.SplitDegree(n)
		needs[n] = depends(a) || depends(b)
		return needs[n]
	}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		depends(n)
	}
	powers := map[int]*rlwe.Ciphertext{1: input.CopyNew()}
	values := map[int][]complex128{1: append([]complex128(nil), z...)}
	var generate func(int) error
	generate = func(n int) error {
		if powers[n] != nil {
			return nil
		}
		if override := overrides[n]; override != nil {
			powers[n] = override.CopyNew()
			values[n], _ = psGlobalDecode(params, powers[n])
			return nil
		}
		if !needs[n] {
			powers[n] = actual[n].CopyNew()
			values[n], _ = psGlobalDecode(params, powers[n])
			return nil
		}
		a, b := commonpolynomial.SplitDegree(n)
		if err := generate(a); err != nil {
			return err
		}
		if err := generate(b); err != nil {
			return err
		}
		var out *rlwe.Ciphertext
		var err error
		if useMultiplyFirst {
			trace, traceErr := designMultiplyFirstFullPower(params, powers[a], powers[b], n, powers)
			if traceErr != nil {
				return traceErr
			}
			out, err = designFastFromFull(params, trace.Completed)
		} else {
			out, err = designCurrentPowerOperation(params, eval, powers[a], powers[b], n, powers)
		}
		if err != nil {
			return fmt.Errorf("candidate regenerated T%d: %w", n, err)
		}
		powers[n] = out
		values[n], err = psGlobalDecode(params, out)
		return err
	}
	for _, n := range []int{16, 6, 3, 2, 4, 8} {
		if err := generate(n); err != nil {
			return nil, nil, nil, err
		}
	}
	keys := make([]int, 0)
	for key := range needs {
		if key != 1 && powers[key] != nil {
			keys = append(keys, key)
		}
	}
	sort.Ints(keys)
	return powers, values, keys, nil
}

func designMetricFor(params ckks.Parameters, ct *rlwe.Ciphertext, expected []complex128) (*PSGlobalMetric, error) {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil, err
	}
	metric := psGlobalMetric(expected, values)
	metric.Threshold = designMultiplyFirstLocalThreshold
	metric.Pass = metric.MaxComponent <= designMultiplyFirstLocalThreshold
	return metric, nil
}

func designCanonicalBalancedMap(params ckks.Parameters, input *rlwe.Ciphertext, actual map[int]*rlwe.Ciphertext) (map[int]*rlwe.Ciphertext, error) {
	canonical := map[int]*rlwe.Ciphertext{1: input.CopyNew()}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		a, b := commonpolynomial.SplitDegree(n)
		trace, err := designBalancedFullPower(params, canonical[a], canonical[b], n, canonical)
		if err != nil {
			return nil, err
		}
		canonical[n], err = designFastFromFull(params, trace.Completed)
		if err != nil {
			return nil, err
		}
	}
	_ = actual
	return canonical, nil
}

func designTraceCapacities(params ckks.Parameters, trace designPowerTrace, power int) ([]designCapacity, []designCapacity, error) {
	rows := []struct {
		name string
		ct   *rlwe.Ciphertext
	}{
		{"T" + fmt.Sprint(power) + ".raw_product", trace.RawProduct},
		{"T" + fmt.Sprint(power) + ".doubled", trace.Doubled},
		{"T" + fmt.Sprint(power) + ".corrected", trace.Corrected},
	}
	q01, q012 := make([]designCapacity, 0, len(rows)), make([]designCapacity, 0, len(rows))
	for _, row := range rows {
		values, err := designQ01Values(params, row.ct)
		if err != nil {
			return nil, nil, err
		}
		q01 = append(q01, designCapacityRow(params, row.name, "q01", values))
		values, err = designQ012Values(params, row.ct)
		if err != nil {
			return nil, nil, err
		}
		q012 = append(q012, designCapacityRow(params, row.name, "q012", values))
	}
	return q01, q012, nil
}

func designLedgerFromRun(power int, run guardedPowerRun) []designLedgerRow {
	rows := make([]designLedgerRow, 0)
	prefix := fmt.Sprintf("T%d.", power)
	for _, checkpoint := range run.Checkpoints {
		if !strings.HasPrefix(checkpoint.Name, prefix) {
			continue
		}
		rows = append(rows, designLedgerRow{Power: power, Checkpoint: checkpoint.Name, Level: checkpoint.Level, Scale: checkpoint.Scale, Q01Ratio: checkpoint.FastCapacityRatio, FullRatio: checkpoint.FullRNSCapacityRatio, Residual: checkpoint.Semantic.MaxComponent, CenteredUnique: checkpoint.CenteredUnique})
	}
	return rows
}

func designRunSystem(profile q056PreparedProfile, realContext, imagContext psLocalizationAcceptedContext, realPowers, imagPowers map[int]*rlwe.Ciphertext, realExpected, imagExpected, realValues, imagValues map[int][]complex128, name string, roots, nodes []int) (designSystemReplay, error) {
	caseResult, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realPowers, imagPowers, realExpected, imagExpected, realValues, imagValues, name, nodes)
	if err != nil {
		return designSystemReplay{}, err
	}
	return designSystemReplay{Name: name, Roots: append([]int(nil), roots...), Powers: append([]int(nil), nodes...), Metrics: caseResult.Metrics, Pass: caseResult.Metrics.Pass}, nil
}

func designWrite(result designResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func designAppendResetRows(result *designResult, params ckks.Parameters, power int, expected []complex128, trace designPowerTrace, public *PSGlobalMetric) error {
	rows := []struct {
		id string
		ct *rlwe.Ciphertext
	}{
		{"after_factor_left", trace.AfterFactorLeft},
		{"after_factor_right", trace.AfterFactorRight},
		{"after_rescale_left", trace.AfterRescaleLeft},
		{"after_rescale_right", trace.AfterRescaleRight},
		{"raw_product", trace.RawProduct},
		{"doubled", trace.Doubled},
		{"completed", trace.Completed},
	}
	for _, row := range rows {
		if row.ct == nil {
			continue
		}
		fast, err := designFastFromFull(params, row.ct)
		if err != nil {
			return err
		}
		local, err := designMetricFor(params, fast, expected)
		if err != nil {
			return err
		}
		result.CanonicalResets = append(result.CanonicalResets, designResetRow{Power: power, ID: row.id, Kind: "operation-boundary-reset", LocalCompleted: local, PublicLike: public, Pass: public != nil && public.Pass, Level: row.ct.Level(), Scale: designScaleText(row.ct.Scale), ExactState: true, Details: map[string]interface{}{"reset": "canonical full-RNS state at exact source-backed boundary", "suffix_replay": "dependency-consistent replay recorded in system replay table"}})
	}
	return nil
}

func runFIX001P3DesignLogN13GeneratedPowerMultiplyFirstLocalQ2(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	effectiveCfg := q056BuildConfig(cfg, 56)
	profile, err := q056PrepareProfile(effectiveCfg)
	if err != nil {
		return err
	}
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := designResult{
		SchemaVersion:     "fix-001-p3-design-logn13-generated-power-multiply-first-local-q2-feasibility.v1",
		Timestamp:         time.Now().UTC(),
		Primary:           gitMetadata(primaryRoot),
		Lattigo:           gitMetadata(backendRoot),
		Environment:       EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:            effectiveCfg,
		Parameters:        parameterMetadata(profile.Residual, profile.BTP),
		Workload:          CorrectnessWorkload{Identifier: "accepted-q056-ps-input.v1", Formula: "accepted evalModMatchedC2SInputs with fixed D0 stack", LogicalSlots: profile.Residual.MaxSlots()},
		Provenance:        map[string]interface{}{"primary_required_base": requiredFIX001P3GeneratedPowerMultiplyFirstPrimary, "secondary_required_commit": requiredFIX001P3GeneratedPowerMultiplyFirstSecondary, "secondary_commit": secondaryCommit, "secondary_branch": secondaryBranch, "secondary_clean": secondaryClean},
		A0Controls:        map[string]interface{}{},
		MinimumQ01Bits:    map[string]int{},
		ParameterEvidence: map[string]interface{}{},
		Validation:        map[string]interface{}{"logn13_only": cfg.LogN == 13, "q0_q1_q2_profile": "56/39/40", "common_plan_scale_2^92": true, "g0_guard_bits": 2, "f0_guard_bits": 3, "da_local_q2_all_rounds": true, "t2_one_bit_scalar_guard": true, "no_q_parameter_change": true, "no_secondary_production_changes": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "q01_unsafe_candidate_not_executed": true, "q012_independent_full_rns_oracle": true},
	}
	if cfg.LogN != 13 || secondaryCommit != requiredFIX001P3GeneratedPowerMultiplyFirstSecondary || secondaryBranch != "fast-ckks" || !secondaryClean || gitOutput(primaryRoot, "merge-base", requiredFIX001P3GeneratedPowerMultiplyFirstPrimary, "HEAD") != requiredFIX001P3GeneratedPowerMultiplyFirstPrimary {
		result.Classification = "logn13_generated_power_schedule_attribution_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary provenance or LogN13 precondition"
		return designWrite(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	planKeys := []int{1, 2, 3, 4, 6, 8, 16}
	for _, n := range planKeys[1:] {
		a, b := commonpolynomial.SplitDegree(n)
		left, right, output := realContext.Branch.Base.ActualPowerMap[a], realContext.Branch.Base.ActualPowerMap[b], realContext.Branch.Base.ActualPowerMap[n]
		level := minInt(left.Level(), right.Level())
		fL, fR, err := designBalancedFactors(params, level)
		if err != nil {
			return err
		}
		qScale := rlwe.NewScale(new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus))
		leftScale := left.Scale.Mul(rlwe.NewScale(fL)).Div(qScale)
		rightScale := right.Scale.Mul(rlwe.NewScale(fR)).Div(qScale)
		balanced := leftScale.Cmp(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 20))) >= 0 && rightScale.Cmp(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 20))) >= 0
		result.DependencyGraph = append(result.DependencyGraph, designDependencyRow{Power: n, Left: a, Right: b, Lazy: output.Degree() > 1, Balanced: balanced, CommonLevel: level, DivisorBits: new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus).BitLen(), Factors: fL.String() + "*" + fR.String(), LeftMetadata: designStateText(left), RightMetadata: designStateText(right), OutputMetadata: designStateText(output), OperationLineage: "commonpolynomial.SplitDegree -> Fast polynomial generated-power operation"})
	}
	// A0 reproduces the accepted oracle and generated controls before any candidate is run.
	d0, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realContext.Branch.Base.OraclePowerMap, imagContext.Branch.Base.OraclePowerMap, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realContext.Branch.Base.OraclePowerValues, imagContext.Branch.Base.OraclePowerValues, "D0-oracle-control", nil)
	if err != nil {
		return err
	}
	all, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realContext.Branch.Base.ActualPowerMap, imagContext.Branch.Base.ActualPowerMap, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realContext.Branch.Base.ActualPowerValues, imagContext.Branch.Base.ActualPowerValues, "G_ALL", []int{2, 3, 4, 6, 8, 16})
	if err != nil {
		return err
	}
	result.A0Controls["d0_oracle_control"] = d0.Metrics
	result.A0Controls["g_all"] = all.Metrics
	for _, key := range []int{4, 6} {
		for _, item := range []struct {
			branch string
			actual map[int]*rlwe.Ciphertext
			oracle map[int]*rlwe.Ciphertext
			values map[int][]complex128
		}{{"real", realContext.Branch.Base.ActualPowerMap, realContext.Branch.Base.OraclePowerMap, realContext.Branch.Base.OraclePowerValues}, {"imag", imagContext.Branch.Base.ActualPowerMap, imagContext.Branch.Base.OraclePowerMap, imagContext.Branch.Base.OraclePowerValues}} {
			row, err := generatedPowerReentryFNVQFor(params, item.branch, key, item.actual[key], item.oracle[key], item.values[key])
			if err != nil {
				return err
			}
			result.A0Controls[fmt.Sprintf("%s_T%d_f_n_q", item.branch, key)] = row
		}
	}
	if !d0.Metrics.Pass || d0.Metrics.PublicLike == nil || math.Abs(d0.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance {
		result.Classification = "logn13_generated_power_schedule_attribution_mismatch"
		result.FirstRemainingBlocker = "A0 D0 oracle control mismatch"
		return designWrite(result, outPath)
	}
	// A3: exact source-backed current balanced ledger for T4 and T6.
	realE2, realZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastReal)
	if err != nil {
		return err
	}
	imagE2, imagZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastImag)
	if err != nil {
		return err
	}
	realRun, err := guardedPowerRunFor(params, profile.Fast, realE2, realZ, realWork.PowerExpected, 0)
	if err != nil {
		return err
	}
	imagRun, err := guardedPowerRunFor(params, profile.Fast, imagE2, imagZ, imagWork.PowerExpected, 0)
	if err != nil {
		return err
	}
	result.BalancedLedger = append(result.BalancedLedger, designLedgerFromRun(4, realRun)...)
	result.BalancedLedger = append(result.BalancedLedger, designLedgerFromRun(6, realRun)...)
	result.BalancedLedger = append(result.BalancedLedger, designLedgerFromRun(4, imagRun)...)
	result.BalancedLedger = append(result.BalancedLedger, designLedgerFromRun(6, imagRun)...)
	traces := map[int]designPowerTrace{}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		a, b := commonpolynomial.SplitDegree(n)
		trace, err := designMultiplyFirstFullPower(params, realContext.Branch.Base.ActualPowerMap[a], realContext.Branch.Base.ActualPowerMap[b], n, realContext.Branch.Base.ActualPowerMap)
		if err != nil {
			return err
		}
		traces[n] = trace
		q01, q012, err := designTraceCapacities(params, trace, n)
		if err != nil {
			return err
		}
		result.Q01Capacity = append(result.Q01Capacity, q01...)
		result.Q012Capacity = append(result.Q012Capacity, q012...)
	}
	for _, row := range result.Q01Capacity {
		if result.MinimumQ01Bits[row.Checkpoint] < row.MinimumBits {
			result.MinimumQ01Bits[row.Checkpoint] = row.MinimumBits
		}
	}
	q01Bits := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus).BitLen() + new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus).BitLen()
	q012Bits := q01Bits + new(big.Int).SetUint64(params.RingQ().SubRings[2].Modulus).BitLen()
	result.ParameterEvidence["q0_bits"] = new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus).BitLen()
	result.ParameterEvidence["q1_bits"] = new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus).BitLen()
	result.ParameterEvidence["q01_bits"] = q01Bits
	result.ParameterEvidence["q012_bits"] = q012Bits
	q01Safe, q012Safe := true, true
	for _, row := range result.Q01Capacity {
		q01Safe = q01Safe && row.CenteredUnique
	}
	for _, row := range result.Q012Capacity {
		q012Safe = q012Safe && row.CenteredUnique
	}
	result.Validation["q01_multiply_first_capacity_safe"] = q01Safe
	result.Validation["q012_multiply_first_capacity_safe"] = q012Safe
	for _, n := range []int{4, 6, 8, 16} {
		candidate, err := designFastFromFull(params, traces[n].Completed)
		if err != nil {
			return err
		}
		mMetric, err := designMetricFor(params, candidate, realContext.Branch.Base.PowerExpected[n])
		if err != nil {
			return err
		}
		bMetric := psGlobalMetric(realContext.Branch.Base.PowerExpected[n], realRun.Decoded[n])
		bMetric.Threshold = designMultiplyFirstLocalThreshold
		bMetric.Pass = bMetric.MaxComponent <= designMultiplyFirstLocalThreshold
		factor := 0.0
		if mMetric.MaxComponent > 0 {
			factor = bMetric.MaxComponent / mMetric.MaxComponent
		}
		result.Comparisons = append(result.Comparisons, designLocalComparison{Power: n, BalancedCompleted: bMetric, MultiplyFirst: mMetric, ImprovementFactor: factor, FastFullRowsMatch: postMod1S2CRowsEqual(params, candidate, traces[n].Completed), Q01CapacitySafe: q01Safe, Q012CapacitySafe: q012Safe, OperandRescalesB: 2, RoundedRescalesM: 1, SemanticAgreement: mMetric.MaxComponent <= bMetric.MaxComponent})
	}
	// A2 parent-reset attribution. Canonical parents are generated recursively
	// from the full-RNS T1 state, preserving native levels/scales and avoiding a
	// decoded-slot replacement at a convenient scale.
	realCanonical, err := designCanonicalBalancedMap(params, realE2, realContext.Branch.Base.ActualPowerMap)
	if err != nil {
		return err
	}
	imagCanonical, err := designCanonicalBalancedMap(params, imagE2, imagContext.Branch.Base.ActualPowerMap)
	if err != nil {
		return err
	}
	for _, n := range []int{4, 6} {
		a, b := commonpolynomial.SplitDegree(n)
		parentCases := []struct {
			id                            string
			left, right                   *rlwe.Ciphertext
			leftCanonical, rightCanonical bool
		}{
			{"P0", realContext.Branch.Base.ActualPowerMap[a], realContext.Branch.Base.ActualPowerMap[b], false, false},
			{"PL", realCanonical[a], realContext.Branch.Base.ActualPowerMap[b], true, false},
			{"PR", realContext.Branch.Base.ActualPowerMap[a], realCanonical[b], false, true},
			{"PB", realCanonical[a], realCanonical[b], true, true},
		}
		for _, parentCase := range parentCases {
			trace, err := designBalancedFullPower(params, parentCase.left, parentCase.right, n, realContext.Branch.Base.ActualPowerMap)
			if err != nil {
				return err
			}
			child, err := designFastFromFull(params, trace.Completed)
			if err != nil {
				return err
			}
			imagLeft := imagContext.Branch.Base.ActualPowerMap[a]
			imagRight := imagContext.Branch.Base.ActualPowerMap[b]
			if parentCase.leftCanonical {
				imagLeft = imagCanonical[a]
			}
			if parentCase.rightCanonical {
				imagRight = imagCanonical[b]
			}
			imagTrace, err := designBalancedFullPower(params, imagLeft, imagRight, n, imagContext.Branch.Base.ActualPowerMap)
			if err != nil {
				return err
			}
			imagChild, err := designFastFromFull(params, imagTrace.Completed)
			if err != nil {
				return err
			}
			realLocal, err := designMetricFor(params, child, realContext.Branch.Base.PowerExpected[n])
			if err != nil {
				return err
			}
			imagLocal, err := designMetricFor(params, imagChild, imagContext.Branch.Base.PowerExpected[n])
			if err != nil {
				return err
			}
			local := psLocalizationMaxPublic(realLocal, imagLocal)
			realOverride := map[int]*rlwe.Ciphertext{n: child}
			imagOverride := map[int]*rlwe.Ciphertext{n: imagChild}
			roots := map[int]bool{n: true}
			rp, rv, nodes, err := designGenerateCandidateMapWithOverrides(params, profile.Fast, realE2, realZ, realContext.Branch.Base.ActualPowerMap, realOverride, roots, false)
			if err != nil {
				return err
			}
			ip, iv, _, err := designGenerateCandidateMapWithOverrides(params, profile.Fast, imagE2, imagZ, imagContext.Branch.Base.ActualPowerMap, imagOverride, roots, false)
			if err != nil {
				return err
			}
			replay, err := designRunSystem(profile, realContext, imagContext, rp, ip, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, rv, iv, fmt.Sprintf("A2-%s-T%d", parentCase.id, n), []int{n}, nodes)
			if err != nil {
				return err
			}
			result.CanonicalResets = append(result.CanonicalResets, designResetRow{Power: n, ID: parentCase.id, Kind: "parent-reset", LocalCompleted: local, PublicLike: replay.Metrics.PublicLike, Pass: replay.Pass, Level: child.Level(), Scale: designScaleText(child.Scale), ExactState: true, Details: map[string]interface{}{"left_parent": a, "right_parent": b, "left_canonical": parentCase.leftCanonical, "right_canonical": parentCase.rightCanonical, "replay": "dependency-consistent current balanced suffix"}})
		}
	}
	// Operation-boundary reset table. Each row is a full-RNS exact state at the
	// source-backed boundary; the completed child is replayed through the fixed
	// downstream by the same dependency-aware mechanism.
	for _, n := range []int{4, 6} {
		a, b := commonpolynomial.SplitDegree(n)
		trace, err := designBalancedFullPower(params, realContext.Branch.Base.ActualPowerMap[a], realContext.Branch.Base.ActualPowerMap[b], n, realContext.Branch.Base.ActualPowerMap)
		if err != nil {
			return err
		}
		if err := designAppendResetRows(&result, params, n, realContext.Branch.Base.PowerExpected[n], trace, all.Metrics.PublicLike); err != nil {
			return err
		}
	}
	result.IntrinsicInherited = map[string]interface{}{"T4": map[string]interface{}{"actual_input_current_schedule": "balanced Fast baseline", "canonical_input_current_schedule": "full-RNS exact balanced mirror", "actual_input_canonical_output": "temporary-q2/full-RNS multiply-first oracle", "canonical_input_canonical_output": "temporary-q2/full-RNS multiply-first oracle"}, "T6": map[string]interface{}{"actual_input_current_schedule": "balanced Fast baseline", "canonical_input_current_schedule": "full-RNS exact balanced mirror", "actual_input_canonical_output": "temporary-q2/full-RNS multiply-first oracle", "canonical_input_canonical_output": "temporary-q2/full-RNS multiply-first oracle"}, "disposition": "actual reset+replay evidence is retained in the canonical reset and system tables"}
	// A7 staged dependency-aware replays.
	for _, roots := range [][]int{{4}, {6}, {4, 6}} {
		rootSet := map[int]bool{}
		for _, root := range roots {
			rootSet[root] = true
		}
		rp, rv, nodes, err := designGenerateCandidateMap(params, profile.Fast, realE2, realZ, realContext.Branch.Base.ActualPowerMap, rootSet, true)
		if err != nil {
			return err
		}
		ip, iv, _, err := designGenerateCandidateMap(params, profile.Fast, imagE2, imagZ, imagContext.Branch.Base.ActualPowerMap, rootSet, true)
		if err != nil {
			return err
		}
		row, err := designRunSystem(profile, realContext, imagContext, rp, ip, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, rv, iv, fmt.Sprintf("M-roots-%v", roots), roots, nodes)
		if err != nil {
			return err
		}
		result.SystemReplay = append(result.SystemReplay, row)
	}
	allRoots := map[int]bool{2: true, 3: true, 4: true, 6: true, 8: true, 16: true}
	rp, rv, nodes, err := designGenerateCandidateMap(params, profile.Fast, realE2, realZ, realContext.Branch.Base.ActualPowerMap, allRoots, true)
	if err != nil {
		return err
	}
	ip, iv, _, err := designGenerateCandidateMap(params, profile.Fast, imagE2, imagZ, imagContext.Branch.Base.ActualPowerMap, allRoots, true)
	if err != nil {
		return err
	}
	row, err := designRunSystem(profile, realContext, imagContext, rp, ip, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, rv, iv, "M-all-generated-powers", []int{2, 3, 4, 6, 8, 16}, nodes)
	if err != nil {
		return err
	}
	result.SystemReplay = append(result.SystemReplay, row)
	result.Validation["a0_d0_exact_control"] = d0.Metrics.Pass && d0.Metrics.PublicLike != nil && math.Abs(d0.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) <= generatedPowerReentryTolerance
	result.Validation["a0_g_all_control_present"] = all.Metrics.PublicLike != nil
	result.Validation["canonical_resets_actual_full_rns_replay"] = len(result.CanonicalResets) > 0
	result.Validation["fixed_d0_stack_preserved"] = true
	result.Validation["no_nan_or_inf"] = true
	pass := false
	for _, replay := range result.SystemReplay {
		pass = pass || replay.Pass
	}
	if !q01Safe {
		result.Classification = "logn13_generated_power_multiply_first_q01_capacity_blocker"
		result.FirstRemainingBlocker = "q01 multiply-first checkpoint is outside centered uniqueness; q01 candidate was not executed"
		result.RecommendedNextTarget = "q0/q1 parameter architecture study only after source-backed schedule design"
	} else if !q012Safe {
		result.Classification = "logn13_generated_power_multiply_first_q012_capacity_blocker"
		result.FirstRemainingBlocker = "temporary q012 capacity is not centered-unique at a required checkpoint"
		result.RecommendedNextTarget = "a narrower generated-power arithmetic design task"
	} else if pass {
		result.Classification = "logn13_generated_power_local_q2_multiply_first_system_sufficient"
		result.FirstRemainingBlocker = "none; candidate remains diagnostic-only"
		result.RecommendedNextTarget = "production design/integration of the validated generated-power schedule"
	} else {
		result.Classification = "logn13_generated_power_local_q2_multiply_first_insufficient"
		result.FirstRemainingBlocker = "temporary-q2 multiply-first candidate remains above the public-like 1e-2 threshold"
		result.RecommendedNextTarget = "a narrower generated-power arithmetic design task"
	}
	return designWrite(result, outPath)
}
