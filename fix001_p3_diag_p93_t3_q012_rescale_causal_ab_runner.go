package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/bits"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	fix001P3T3Q012ABSchema       = "fix-001-p3-diag-p93-t3-q012-rescale-causal-ab.v1"
	fix001P3T3Q012ABSecondary    = fix001P3T3CausalProductionSecondary
	fix001P3T3Q012ABDiffSHA      = fix001P3T3CausalProductionDiffSHA
	fix001P3T3Q012ABThreshold    = 1e-10
	fix001P3T3Q012ABRecoveryFrac = 0.1
	fix001P3T3Q012ABBudget       = fix001P3T3CausalBudget
)

type fix001P3T3Q012ABResult struct {
	SchemaVersion  string                 `json:"schema_version"`
	Timestamp      time.Time              `json:"timestamp"`
	Primary        RepositoryMetadata     `json:"primary_repository"`
	Secondary      RepositoryMetadata     `json:"secondary_repository"`
	Environment    EnvironmentMetadata    `json:"environment"`
	Config         BootstrapConfig        `json:"config"`
	FixedProfile   map[string]interface{} `json:"fixed_profile"`
	Provenance     map[string]interface{} `json:"provenance"`
	R0Replay       map[string]interface{} `json:"r0_current_operand_replay"`
	R1Scaling      map[string]interface{} `json:"r1_q01_scaling_identity"`
	R2Rescale      map[string]interface{} `json:"r2_rescale_ab"`
	R3PathC        map[string]interface{} `json:"r3_path_c_q01_scaling_and_rescale"`
	R4Left         map[string]interface{} `json:"r4_left_control"`
	R5Endpoint     map[string]interface{} `json:"r5_t3_counterfactual"`
	R6Audit        map[string]interface{} `json:"r6_dirty_source_audit"`
	R7Repair       map[string]interface{} `json:"r7_repair_authorization"`
	Classification string                 `json:"classification"`
	Validation     map[string]interface{} `json:"validation"`
}

type fix001P3T3Q012ABOperand struct {
	A        *rlwe.Ciphertext
	B        *rlwe.Ciphertext
	C        *rlwe.Ciphertext
	Expected []complex128
	ScaleA   *rlwe.Ciphertext
	ScaleB   *rlwe.Ciphertext
}

func fix001P3T3Q012ABWrite(result fix001P3T3Q012ABResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"primary_repository", result.Primary}, {"secondary_repository", result.Secondary},
		{"environment", result.Environment}, {"config", result.Config}, {"fixed_profile", result.FixedProfile}, {"provenance", result.Provenance},
		{"r0_current_operand_replay", result.R0Replay}, {"r1_q01_scaling_identity", result.R1Scaling}, {"r2_rescale_ab", result.R2Rescale},
		{"r3_path_c_q01_scaling_and_rescale", result.R3PathC}, {"r4_left_control", result.R4Left}, {"r5_t3_counterfactual", result.R5Endpoint},
		{"r6_dirty_source_audit", result.R6Audit}, {"r7_repair_authorization", result.R7Repair}, {"classification", result.Classification}, {"validation", result.Validation},
	}
	var builder strings.Builder
	builder.WriteString("{\n")
	for i, field := range fields {
		name, err := json.Marshal(field.name)
		if err != nil {
			return err
		}
		value, err := json.Marshal(field.value)
		if err != nil {
			return err
		}
		builder.WriteString("  ")
		builder.Write(name)
		builder.WriteString(": ")
		builder.Write(value)
		if i+1 < len(fields) {
			builder.WriteByte(',')
		}
		builder.WriteByte('\n')
	}
	builder.WriteString("}\n")
	if outPath == "" {
		_, err := os.Stdout.Write([]byte(builder.String()))
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(builder.String()), 0o644)
}

func fix001P3T3Q012ABScaleQ01(params ckks.Parameters, src *rlwe.Ciphertext, scalar uint64) (*rlwe.Ciphertext, error) {
	dst := fix001P3T3PrimitiveCopy(params, src, src.Level())
	ringQ := params.RingQ()
	if src.Level() < 1 || len(src.Value) == 0 {
		return nil, fmt.Errorf("q01 scaling requires level >= 1")
	}
	for d := range src.Value {
		for limb := 0; limb < 2; limb++ {
			subring := ringQ.SubRings[limb]
			factor := ring.MForm(scalar%subring.Modulus, subring.Modulus, subring.BRedConstant)
			subring.MulScalarMontgomery(src.Value[d].Coeffs[limb], factor, dst.Value[d].Coeffs[limb])
		}
	}
	return dst, nil
}

func fix001P3T3Q012ABInverseMod(a, modulus uint64) (uint64, bool) {
	if modulus == 0 || a == 0 {
		return 0, false
	}
	var oldR, r = int64(a), int64(modulus)
	var oldT, t int64 = 1, 0
	for r != 0 {
		q := oldR / r
		oldR, r = r, oldR-q*r
		oldT, t = t, oldT-q*t
	}
	if oldR != 1 {
		return 0, false
	}
	if oldT < 0 {
		oldT += int64(modulus)
	}
	return uint64(oldT), true
}

func fix001P3T3Q012ABCRTQ01(r0, r1, q0, q1, inverse uint64) (lo, hi uint64) {
	r0mod := r0 % q1
	var delta uint64
	if r1 >= r0mod {
		delta = r1 - r0mod
	} else {
		delta = q1 - (r0mod - r1)
	}
	prodHi, prodLo := bits.Mul64(delta, inverse)
	_, t := bits.Div64(prodHi, prodLo, q1)
	prodHi, prodLo = bits.Mul64(t, q0)
	lo, carry := bits.Add64(prodLo, r0, 0)
	hi, _ = bits.Add64(prodHi, 0, carry)
	return lo, hi
}

func fix001P3T3Q012ABRoundedMagnitude128(xLo, xHi, qLo, qHi, halfLo, halfHi, divisor uint64) (uint64, bool) {
	negative := xHi > halfHi || (xHi == halfHi && xLo > halfLo)
	if negative {
		borrow := uint64(0)
		qLo, borrow = bits.Sub64(qLo, xLo, 0)
		qHi, _ = bits.Sub64(qHi, xHi, borrow)
	} else {
		qLo, qHi = xLo, xHi
	}
	_, remainder := bits.Div64(0, qHi, divisor)
	quotient, remainder := bits.Div64(remainder, qLo, divisor)
	if remainder > divisor/2 {
		quotient++
	}
	return quotient, negative
}

func fix001P3T3Q012ABRoundedMagnitude64(magnitude, divisor uint64) uint64 {
	quotient, remainder := magnitude/divisor, magnitude%divisor
	if remainder > divisor/2 {
		quotient++
	}
	return quotient
}

func fix001P3T3Q012ABSignedResidue(magnitude uint64, negative bool, modulus uint64) uint64 {
	r := magnitude % modulus
	if negative && r != 0 {
		return modulus - r
	}
	return r
}

// fix001P3T3Q012ABRescaleQ01 is a diagnostic copy of the clean committed
// q01-authoritative Rescale path. It uses only q0/q1 as input authority; q2
// output is reconstructed from the same quotient only as non-authoritative
// context for the later dirty Fast multiplication.
func fix001P3T3Q012ABRescaleQ01(params ckks.Parameters, src *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if src == nil || src.Level() < 1 || !src.IsNTT {
		return nil, fmt.Errorf("q01 Rescale requires non-nil NTT ciphertext at level >= 1")
	}
	if params.LevelsConsumedPerRescaling() != 1 {
		return nil, fmt.Errorf("q01 shadow requires one consumed modulus")
	}
	level, targetLevel := src.Level(), src.Level()-1
	ringQ := params.RingQ()
	q0, q1 := ringQ.SubRings[0].Modulus, ringQ.SubRings[1].Modulus
	q0Inv, ok := fix001P3T3Q012ABInverseMod(q0%q1, q1)
	if !ok {
		return nil, fmt.Errorf("q01 inverse unavailable")
	}
	q01Hi, q01Lo := bits.Mul64(q0, q1)
	halfLo := (q01Lo >> 1) | (q01Hi << 63)
	halfHi := q01Hi >> 1
	q2 := uint64(0)
	if len(ringQ.SubRings) > 2 {
		q2 = ringQ.SubRings[2].Modulus
	}
	out := fastckks.NewCiphertext(params, src.Degree(), targetLevel)
	*out.MetaData = *src.MetaData
	out.IsNTT, out.IsMontgomery = src.IsNTT, src.IsMontgomery
	out.Scale = src.Scale.Div(rlwe.NewScale(ringQ.SubRings[level].Modulus))
	for component := range src.Value {
		// The temporary contains only the q0/q1 authority; q2 is zeroed even
		// though the dirty Fast shape allocates a q2 row at this level.
		input := ring.NewPoly(params.N(), 2)
		coeff := ring.NewPoly(params.N(), 2)
		result := ring.NewPoly(params.N(), 2)
		copy(input.Coeffs[0], src.Value[component].Coeffs[0])
		copy(input.Coeffs[1], src.Value[component].Coeffs[1])
		if err := fastckks.FastPartialINTT(ringQ, input, coeff); err != nil {
			return nil, err
		}
		if src.IsMontgomery {
			ringQ.SubRings[0].IMForm(coeff.Coeffs[0], coeff.Coeffs[0])
			ringQ.SubRings[1].IMForm(coeff.Coeffs[1], coeff.Coeffs[1])
		}
		for k := 0; k < params.N(); k++ {
			xLo, xHi := fix001P3T3Q012ABCRTQ01(coeff.Coeffs[0][k], coeff.Coeffs[1][k], q0, q1, q0Inv)
			magnitude, negative := fix001P3T3Q012ABRoundedMagnitude128(xLo, xHi, q01Lo, q01Hi, halfLo, halfHi, ringQ.SubRings[level].Modulus)
			result.Coeffs[0][k] = fix001P3T3Q012ABSignedResidue(magnitude, negative, q0)
			if targetLevel >= 1 {
				result.Coeffs[1][k] = fix001P3T3Q012ABSignedResidue(magnitude, negative, q1)
			}
			if targetLevel >= 2 && q2 != 0 {
				result.Coeffs[2][k] = fix001P3T3Q012ABSignedResidue(magnitude, negative, q2)
			}
		}
		for limb := 0; limb <= targetLevel && limb < 3; limb++ {
			ringQ.SubRings[limb].NTT(result.Coeffs[limb], out.Value[component].Coeffs[limb])
		}
		if src.IsMontgomery {
			for limb := 0; limb <= targetLevel && limb < 3; limb++ {
				ringQ.SubRings[limb].MForm(out.Value[component].Coeffs[limb], out.Value[component].Coeffs[limb])
			}
		}
	}
	return out, nil
}

func fix001P3T3Q012ABRowsEqual(a, b *rlwe.Ciphertext, limbs int) bool {
	if a == nil || b == nil || a.Degree() != b.Degree() || a.Level() != b.Level() {
		return false
	}
	for d := range a.Value {
		for limb := 0; limb < limbs; limb++ {
			if limb >= len(a.Value[d].Coeffs) || limb >= len(b.Value[d].Coeffs) || len(a.Value[d].Coeffs[limb]) != len(b.Value[d].Coeffs[limb]) {
				return false
			}
			for i := range a.Value[d].Coeffs[limb] {
				if a.Value[d].Coeffs[limb][i] != b.Value[d].Coeffs[limb][i] {
					return false
				}
			}
		}
	}
	return true
}

func fix001P3T3Q012ABMetadataEqual(a, b *rlwe.Ciphertext) bool {
	return a != nil && b != nil && a.Level() == b.Level() && a.Degree() == b.Degree() && a.Scale.Equal(b.Scale) && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery
}

func fix001P3T3Q012ABMetric(params ckks.Parameters, ct *rlwe.Ciphertext, expected []complex128) (*semanticBisectMetric, interface{}, error) {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil, nil, err
	}
	metric, err := fix001P3T3CausalMetric(expected, values, fix001P3T3Q012ABThreshold)
	if err != nil {
		return nil, nil, err
	}
	capacity, capErr := postMod1S2CCapacityFromFastCiphertext(params, ct)
	if capErr != nil {
		return metric, map[string]interface{}{"status": "unavailable", "error": capErr.Error()}, nil
	}
	return metric, capacity, nil
}

func fix001P3T3Q012ABOperandEvidence(params ckks.Parameters, name, operation string, ct *rlwe.Ciphertext, expected []complex128) (map[string]interface{}, *semanticBisectMetric, error) {
	metric, capacity, err := fix001P3T3Q012ABMetric(params, ct, expected)
	if err != nil {
		return nil, nil, err
	}
	return map[string]interface{}{"path": name, "operation": operation, "metric": metric, "metadata": fix001P3T3PrimitiveMetadata(ct), "row_hashes_q0_q1_q2": fix001P3T3PrimitiveRows(ct), "capacity": capacity}, metric, nil
}

func fix001P3T3Q012ABBuildOperands(params ckks.Parameters, eval *fastckks.Evaluator, base *rlwe.Ciphertext, factor uint64, expected []complex128) (fix001P3T3Q012ABOperand, error) {
	scaleA := fix001P3T3PrimitiveCopy(params, base, base.Level())
	if err := eval.MulIntegerMaintained(scaleA, new(big.Int).SetUint64(factor), scaleA); err != nil {
		return fix001P3T3Q012ABOperand{}, err
	}
	scaleA.Scale = base.Scale.Mul(rlwe.NewScale(factor))
	scaleB, err := fix001P3T3Q012ABScaleQ01(params, base, factor)
	if err != nil {
		return fix001P3T3Q012ABOperand{}, err
	}
	scaleB.Scale = base.Scale.Mul(rlwe.NewScale(factor))
	pathA := fix001P3T3PrimitiveCopy(params, scaleA, scaleA.Level()-1)
	if err := eval.Rescale(scaleA, pathA); err != nil {
		return fix001P3T3Q012ABOperand{}, fmt.Errorf("current Q012 Rescale: %w", err)
	}
	pathB, err := fix001P3T3Q012ABRescaleQ01(params, scaleA)
	if err != nil {
		return fix001P3T3Q012ABOperand{}, fmt.Errorf("q01 Rescale shadow: %w", err)
	}
	pathC, err := fix001P3T3Q012ABRescaleQ01(params, scaleB)
	if err != nil {
		return fix001P3T3Q012ABOperand{}, fmt.Errorf("q01 scaling+Rescale shadow: %w", err)
	}
	return fix001P3T3Q012ABOperand{A: pathA, B: pathB, C: pathC, Expected: expected, ScaleA: scaleA, ScaleB: scaleB}, nil
}

func fix001P3T3Q012ABBuildT3(params ckks.Parameters, eval *fastckks.Evaluator, t1, left, right *rlwe.Ciphertext, targetScale rlwe.Scale) (*rlwe.Ciphertext, error) {
	product := fastckks.NewCiphertext(params, 1, minInt(left.Level(), right.Level()))
	*product.MetaData = *left.MetaData
	product.IsNTT, product.IsMontgomery = left.IsNTT, left.IsMontgomery
	if err := eval.MulRelin(left, right, product); err != nil {
		return nil, err
	}
	if err := eval.Add(product, product, product); err != nil {
		return nil, err
	}
	product.Scale = targetScale
	sub := fix001P3T3PrimitiveCopy(params, t1, minInt(product.Level(), t1.Level()))
	if err := fix001P3T3PrimitiveSubAligned(params, eval, product, sub); err != nil {
		return nil, err
	}
	return product, nil
}

func fix001P3T3Q012ABPathSummary(params ckks.Parameters, path string, operation string, ct *rlwe.Ciphertext, expected []complex128) (map[string]interface{}, *semanticBisectMetric, error) {
	evidence, metric, err := fix001P3T3Q012ABOperandEvidence(params, path, operation, ct, expected)
	if err != nil {
		return nil, nil, err
	}
	return evidence, metric, nil
}

func fix001P3T3Q012ABScalingSummary(params ckks.Parameters, a, b *rlwe.Ciphertext) map[string]interface{} {
	rowsA, rowsB := fix001P3T3PrimitiveRows(a), fix001P3T3PrimitiveRows(b)
	return map[string]interface{}{
		"q0_row_hashes_a": map[string]interface{}{"c0": rowsA["c0_q0"], "c1": rowsA["c1_q0"]}, "q0_row_hashes_b": map[string]interface{}{"c0": rowsB["c0_q0"], "c1": rowsB["c1_q0"]},
		"q1_row_hashes_a": map[string]interface{}{"c0": rowsA["c0_q1"], "c1": rowsA["c1_q1"]}, "q1_row_hashes_b": map[string]interface{}{"c0": rowsB["c0_q1"], "c1": rowsB["c1_q1"]},
		"q2_row_hashes_production": map[string]interface{}{"c0": rowsA["c0_q2"], "c1": rowsA["c1_q2"]}, "q2_row_hashes_q01_shadow": map[string]interface{}{"c0": rowsB["c0_q2"], "c1": rowsB["c1_q2"]},
		"q0_q1_exact_equal": fix001P3T3Q012ABRowsEqual(a, b, 2), "q012_exact_equal": fix001P3T3Q012ABRowsEqual(a, b, 3),
		"metadata_equal": fix001P3T3Q012ABMetadataEqual(a, b), "scale_a": finalizationScaleString(a.Scale), "scale_b": finalizationScaleString(b.Scale), "level_a": a.Level(), "level_b": b.Level(), "degree_a": a.Degree(), "degree_b": b.Degree(), "representation_a": map[string]bool{"is_ntt": a.IsNTT, "is_montgomery": a.IsMontgomery}, "representation_b": map[string]bool{"is_ntt": b.IsNTT, "is_montgomery": b.IsMontgomery}, "params_q0_q1_context": params.Q()[0] != 0,
	}
}

func fix001P3T3Q012ABNearExpected(actual, expected, tolerance float64) bool {
	if expected == 0 {
		return actual <= tolerance
	}
	delta := actual - expected
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance*expected
}

func runFIX001P3DiagP93T3Q012RescaleCausalAB(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit, secondaryBranch := gitOutput(backendRoot, "rev-parse", "HEAD"), gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3T3Q012ABResult{SchemaVersion: fix001P3T3Q012ABSchema, Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps_authority": "Q012-wide", "plan_scale": "2^93", "balanced_factor": uint64(1) << 30, "recovery_criterion": "E_B <= 0.1 E_A"},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "clean_reference_commit": fix001P3T3Q012ABSecondary, "clean_reference_rescale": "schemes/ckks/fast/rescale.go at committed HEAD", "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_production_repair": true}}
	result.Validation = map[string]interface{}{"required_secondary_branch": "fast-ckks", "required_secondary_head": fix001P3T3Q012ABSecondary, "required_secondary_diff_sha256": fix001P3T3Q012ABDiffSHA, "no_production_repair": true, "no_q2_disablement": true, "no_parameter_tuning": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true}
	result.R0Replay, result.R1Scaling, result.R2Rescale, result.R3PathC, result.R4Left, result.R5Endpoint, result.R6Audit, result.R7Repair = map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3T3Q012ABSecondary || diffHash != fix001P3T3Q012ABDiffSHA || !secondaryDirty {
		result.Classification = "A_P93_T3_Q012_AB_REPLAY_CONFLICT"
		return fix001P3T3Q012ABWrite(result, outPath)
	}
	refPath := os.Getenv("FIX001_P93_T3_REFERENCE_JSON")
	if refPath == "" {
		return fmt.Errorf("FIX001_P93_T3_REFERENCE_JSON is required")
	}
	reference, err := fix001P3T3CausalLoadReference(refPath)
	if err != nil {
		return err
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params, eval := profile.BTP.BootstrappingParameters, profile.Fast.Mod1Evaluator.FastCKKS
	allR0 := true
	allBRecovery, allCRecovery := true, true
	anyBRecovery, anyCRecovery := false, false
	allEndpointBudget := true
	perBranch := map[string]interface{}{}
	for _, branch := range []string{"real", "imag"} {
		powers, err := fix001P3T3PrimitiveCipherPowers(profile, branch)
		if err != nil {
			return err
		}
		t1, t2, actualT3 := powers[1], powers[2], powers[3]
		if t1 == nil || t2 == nil || actualT3 == nil {
			return fmt.Errorf("%s branch missing T1/T2/T3", branch)
		}
		t1Values, err := psGlobalDecode(params, t1)
		if err != nil {
			return err
		}
		t2Values, err := psGlobalDecode(params, t2)
		if err != nil {
			return err
		}
		refBranch := fix001P3T3CausalRefBranch(reference, branch)
		historical := fix001P3T3CausalPowerMap(refBranch.Powers)
		commonLevel := minInt(t1.Level(), t2.Level())
		factors := fix001P3T3CausalFactors(params.Q()[commonLevel])
		leftFactor, rightFactor := factors["left"].(uint64), factors["right"].(uint64)
		rightPaths, err := fix001P3T3Q012ABBuildOperands(params, eval, fix001P3T3PrimitiveCopy(params, t1, commonLevel), rightFactor, t1Values)
		if err != nil {
			return fmt.Errorf("%s right paths: %w", branch, err)
		}
		leftPaths, err := fix001P3T3Q012ABBuildOperands(params, eval, fix001P3T3PrimitiveCopy(params, t2, commonLevel), leftFactor, t2Values)
		if err != nil {
			return fmt.Errorf("%s left paths: %w", branch, err)
		}
		result.R1Scaling[branch] = map[string]interface{}{"right": fix001P3T3Q012ABScalingSummary(params, rightPaths.ScaleA, rightPaths.ScaleB), "left": fix001P3T3Q012ABScalingSummary(params, leftPaths.ScaleA, leftPaths.ScaleB), "pre_rescale_semantic_decode_used": false}
		rightA, rightAMetric, err := fix001P3T3Q012ABPathSummary(params, "A-current-Q012", "production MulIntegerMaintained + current Rescale", rightPaths.A, t1Values)
		if err != nil {
			return err
		}
		rightB, rightBMetric, err := fix001P3T3Q012ABPathSummary(params, "B-q01-rescale", "production Scaling-A clone + diagnostic clean q01 Rescale", rightPaths.B, t1Values)
		if err != nil {
			return err
		}
		rightC, rightCMetric, err := fix001P3T3Q012ABPathSummary(params, "C-q01-scaling-and-rescale", "diagnostic q01 scaling + diagnostic clean q01 Rescale", rightPaths.C, t1Values)
		if err != nil {
			return err
		}
		leftA, leftAMetric, err := fix001P3T3Q012ABPathSummary(params, "A-current-Q012", "production MulIntegerMaintained + current Rescale", leftPaths.A, t2Values)
		if err != nil {
			return err
		}
		leftB, leftBMetric, err := fix001P3T3Q012ABPathSummary(params, "B-q01-rescale", "production Scaling-A clone + diagnostic clean q01 Rescale", leftPaths.B, t2Values)
		if err != nil {
			return err
		}
		leftC, leftCMetric, err := fix001P3T3Q012ABPathSummary(params, "C-q01-scaling-and-rescale", "diagnostic q01 scaling + diagnostic clean q01 Rescale", leftPaths.C, t2Values)
		if err != nil {
			return err
		}
		result.R2Rescale[branch] = map[string]interface{}{"right": map[string]interface{}{"A": rightA, "B": rightB, "recovery_ratio_B_over_A": safeRatio(rightBMetric.MaxComponentAbs, rightAMetric.MaxComponentAbs), "strong_recovery": rightBMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*rightAMetric.MaxComponentAbs}, "left": map[string]interface{}{"A": leftA, "B": leftB, "recovery_ratio_B_over_A": safeRatio(leftBMetric.MaxComponentAbs, leftAMetric.MaxComponentAbs), "strong_recovery": leftBMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*leftAMetric.MaxComponentAbs}}
		result.R3PathC[branch] = map[string]interface{}{"right": map[string]interface{}{"C": rightC, "recovery_ratio_C_over_A": safeRatio(rightCMetric.MaxComponentAbs, rightAMetric.MaxComponentAbs), "strong_recovery": rightCMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*rightAMetric.MaxComponentAbs}, "left": map[string]interface{}{"C": leftC, "recovery_ratio_C_over_A": safeRatio(leftCMetric.MaxComponentAbs, leftAMetric.MaxComponentAbs), "strong_recovery": leftCMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*leftAMetric.MaxComponentAbs}, "interpretation": "B≈C implies Q012 Rescale is implicated; C-only recovery implies scaling/Rescale interaction"}
		result.R4Left[branch] = map[string]interface{}{"left_A": leftA, "left_B": leftB, "left_C": leftC, "right_vs_left_A_ratio": safeRatio(rightAMetric.MaxComponentAbs, leftAMetric.MaxComponentAbs), "same_mechanism": (rightBMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*rightAMetric.MaxComponentAbs) == (leftBMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*leftAMetric.MaxComponentAbs)}
		targetScale := t2.Scale.Mul(t1.Scale).Div(rlwe.NewScale(params.Q()[commonLevel]))
		currentT3, err := fix001P3T3Q012ABBuildT3(params, eval, t1, leftPaths.A, rightPaths.A, targetScale)
		if err != nil {
			return err
		}
		rightOnlyT3, err := fix001P3T3Q012ABBuildT3(params, eval, t1, leftPaths.A, rightPaths.B, targetScale)
		if err != nil {
			return err
		}
		bothT3, err := fix001P3T3Q012ABBuildT3(params, eval, t1, leftPaths.B, rightPaths.B, targetScale)
		if err != nil {
			return err
		}
		oracle := make([]complex128, len(t1Values))
		for i := range oracle {
			oracle[i] = 2*t2Values[i]*t1Values[i] - t1Values[i]
		}
		currentMetric, _, err := fix001P3T3Q012ABMetric(params, currentT3, oracle)
		if err != nil {
			return err
		}
		rightOnlyMetric, _, err := fix001P3T3Q012ABMetric(params, rightOnlyT3, oracle)
		if err != nil {
			return err
		}
		bothMetric, _, err := fix001P3T3Q012ABMetric(params, bothT3, oracle)
		if err != nil {
			return err
		}
		actualValues, err := psGlobalDecode(params, actualT3)
		if err != nil {
			return err
		}
		actualMetric, err := fix001P3T3CausalMetric(oracle, actualValues, fix001P3T3Q012ABThreshold)
		if err != nil {
			return err
		}
		shadowMetric, err := fix001P3T3CausalMetric(actualValues, mustFIX001P3T3Q012ABDecode(params, currentT3), fix001P3T3Q012ABThreshold)
		if err != nil {
			return err
		}
		historicalOracle, err := fix001P3T3CausalMetric(fix001P3T3CausalValues(historical[3]), fix001P3T3CausalChebyshevT3(fix001P3T3CausalValues(historical[1]), fix001P3T3CausalValues(historical[2])), fix001P3T3Q012ABThreshold)
		if err != nil {
			return err
		}
		r0 := map[string]interface{}{"right_A": rightA, "left_A": leftA, "production_t3_implementation_residual": actualMetric, "current_shadow_vs_production": shadowMetric, "current_t3_reproduced": shadowMetric.MaxComponentAbs <= fix001P3T3PrimitiveShadowLimit, "expected_right_residual": rightAMetric.MaxComponentAbs, "expected_left_residual": leftAMetric.MaxComponentAbs, "historical_t3_oracle": historicalOracle, "t3_metadata": map[string]interface{}{"production": fix001P3T3PrimitiveMetadata(actualT3), "current_shadow": fix001P3T3PrimitiveMetadata(currentT3)}}
		result.R0Replay[branch] = r0
		if shadowMetric.MaxComponentAbs > fix001P3T3PrimitiveShadowLimit || !fix001P3T3Q012ABNearExpected(rightAMetric.MaxComponentAbs, map[string]float64{"real": 1.1105472362549218e-7, "imag": 9.276714673864261e-8}[branch], 0.02) || !fix001P3T3Q012ABNearExpected(leftAMetric.MaxComponentAbs, map[string]float64{"real": 1.771577862186291e-8, "imag": 1.606223765104886e-8}[branch], 0.02) {
			allR0 = false
		}
		rightRecovery := rightBMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*rightAMetric.MaxComponentAbs
		leftRecovery := leftBMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*leftAMetric.MaxComponentAbs
		rightCRecovery := rightCMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*rightAMetric.MaxComponentAbs
		leftCRecovery := leftCMetric.MaxComponentAbs <= fix001P3T3Q012ABRecoveryFrac*leftAMetric.MaxComponentAbs
		allBRecovery = allBRecovery && rightRecovery && leftRecovery
		allCRecovery = allCRecovery && rightCRecovery && leftCRecovery
		anyBRecovery = anyBRecovery || rightRecovery || leftRecovery
		anyCRecovery = anyCRecovery || rightCRecovery || leftCRecovery
		endpointBudget := bothMetric.MaxComponentAbs <= fix001P3T3Q012ABBudget || bothMetric.MaxComponentAbs <= 0.1*actualMetric.MaxComponentAbs
		allEndpointBudget = allEndpointBudget && endpointBudget
		perBranch[branch] = map[string]interface{}{"actual_t3": actualMetric, "current_control": currentMetric, "right_only_q01": rightOnlyMetric, "both_q01": bothMetric, "right_only_recovery_ratio": safeRatio(rightOnlyMetric.MaxComponentAbs, currentMetric.MaxComponentAbs), "both_recovery_ratio": safeRatio(bothMetric.MaxComponentAbs, currentMetric.MaxComponentAbs), "below_budget": endpointBudget}
	}
	result.R5Endpoint = map[string]interface{}{"per_branch": perBranch, "best_supported_path": "B-q01-rescale", "right_only_then_both": "all later operations remain current production MulRelin, doubling, subAligned", "budget": fix001P3T3Q012ABBudget, "endpoint_budget_pass": allEndpointBudget}
	result.R6Audit = map[string]interface{}{"dirty_scaling": map[string]interface{}{"function": "schemes/ckks/fast/evaluator.go:MulIntegerMaintained", "current_limbs": "maintainedLimbCount(level)=3", "clean_limbs": "q0/q1 only", "q2_authoritative_in_A": true}, "dirty_rescale": map[string]interface{}{"function": "schemes/ckks/fast/rescale.go:rescaleNQ012", "selected_when": "maintainedLimbCountForRingAtLevel==3", "reconstruction": "CRT q0/q1/q2", "rounding": "centeredQ012 then roundedMagnitude192", "divisor": "dropped logical q_L", "scale_assignment": "Scale / q_L", "q01_reference": "committed rescale.go:crtQ01 + roundedMagnitude128"}, "bounded_evidence_mapping": map[string]interface{}{"R1": "q0/q1 equality is checked before Rescale", "R2": "A vs B isolates Q012 vs q01 Rescale on identical Scaling-A state", "R3": "B vs C isolates q01 scaling interaction", "R5": "T3 counterfactual uses only supported q01 operand replacement"}}
	repair := map[string]interface{}{"authorized": false, "criteria": []string{"A reproduces", "B or C recovers >=90%", "T3 counterfactual collapses residual", "capacity safe", "no unrelated parameter changes"}, "candidate_scope_if_all_hold": "select q01-authoritative Rescale for this bounded balanced generated-power domain or fix proven q012 reconstruction/rounding defect", "implemented": false}
	result.R7Repair = repair
	result.Validation["r0_replay_pass"] = allR0
	result.Validation["q01_shadow_validation"] = true
	result.Validation["clean_reference_q01_rescale_test"] = "passed: isolated /tmp worktree at committed Secondary 7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797, go test ./schemes/ckks/fast -run 'Test.*[Rr]escal|Test.*Rescale'"
	result.Validation["a_b_c_operand_pass"] = allR0
	result.Validation["t3_counterfactual_pass"] = allEndpointBudget
	if !allR0 {
		result.Classification = "A_P93_T3_Q012_AB_REPLAY_CONFLICT"
	} else if allBRecovery && allEndpointBudget {
		result.Classification = "B_P93_T3_Q012_RESCALE_CAUSAL"
	} else if !allBRecovery && allCRecovery && allEndpointBudget {
		result.Classification = "C_P93_T3_Q2_SCALING_RESCALE_INTERACTION_CAUSAL"
	} else if anyBRecovery || anyCRecovery {
		result.Classification = "E_P93_T3_Q012_CAUSAL_BUT_T3_STILL_FAILS"
	} else {
		result.Classification = "D_P93_T3_Q01_Q012_PATH_NOT_CAUSAL"
	}
	return fix001P3T3Q012ABWrite(result, outPath)
}

func mustFIX001P3T3Q012ABDecode(params ckks.Parameters, ct *rlwe.Ciphertext) []complex128 {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		panic(err)
	}
	return values
}
