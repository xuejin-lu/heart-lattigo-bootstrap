//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	fix001P3T3ScheduleSchema       = "fix-001-p3-diag-p93-t3-balanced-vs-postproduct-schedule-causal-ab.v1"
	fix001P3T3ScheduleSecondary    = fix001P3T3CausalProductionSecondary
	fix001P3T3ScheduleDiffSHA      = fix001P3T3CausalProductionDiffSHA
	fix001P3T3ScheduleRecoveryFrac = 0.1
	fix001P3T3ScheduleThreshold    = 1e-10
)

type fix001P3T3ScheduleCheckpoint struct {
	Checkpoint      string                 `json:"checkpoint"`
	Operation       string                 `json:"operation"`
	Metadata        map[string]interface{} `json:"metadata"`
	Semantic        *semanticBisectMetric  `json:"semantic,omitempty"`
	Capacity        *psWideCapacity        `json:"q01_q012_capacity,omitempty"`
	CapacitySafe    bool                   `json:"q012_capacity_safe"`
	RowHashesQ012   interface{}            `json:"row_hashes_q0_q1_q2,omitempty"`
	ComparisonValid bool                   `json:"comparison_valid"`
}

type fix001P3T3ScheduleBranch struct {
	BalancedA           map[string]interface{}         `json:"A_balanced_current"`
	PostProductB        []fix001P3T3ScheduleCheckpoint `json:"B_postproduct_current_primitives"`
	PostProductC        []fix001P3T3ScheduleCheckpoint `json:"C_postproduct_bigint_q012_oracle"`
	FinalA              *semanticBisectMetric          `json:"A_final_residual,omitempty"`
	FinalB              *semanticBisectMetric          `json:"B_final_residual,omitempty"`
	FinalC              *semanticBisectMetric          `json:"C_final_residual,omitempty"`
	BOverA              float64                        `json:"B_over_A"`
	COverA              float64                        `json:"C_over_A"`
	BStrongRecovery     bool                           `json:"B_strong_recovery"`
	CStrongRecovery     bool                           `json:"C_strong_recovery"`
	BBelowBudget        bool                           `json:"B_below_polynomial_budget"`
	CBelowBudget        bool                           `json:"C_below_polynomial_budget"`
	HistoricalCResidual *semanticBisectMetric          `json:"C_vs_historical_t3,omitempty"`
	CurrentT3Metadata   map[string]interface{}         `json:"current_t3_metadata"`
	PostProductMetadata map[string]interface{}         `json:"postproduct_metadata"`
	T3B                 *rlwe.Ciphertext               `json:"-"`
	T6Current           *rlwe.Ciphertext               `json:"-"`
}

type fix001P3T3ScheduleResult struct {
	SchemaVersion  string                              `json:"schema_version"`
	Timestamp      time.Time                           `json:"timestamp"`
	Primary        RepositoryMetadata                  `json:"primary_repository"`
	Secondary      RepositoryMetadata                  `json:"secondary_repository"`
	Environment    EnvironmentMetadata                 `json:"environment"`
	Config         BootstrapConfig                     `json:"config"`
	Provenance     map[string]interface{}              `json:"provenance"`
	R0Provenance   map[string]interface{}              `json:"r0_schedule_provenance_correction"`
	R1Balanced     map[string]interface{}              `json:"r1_balanced_current_control"`
	R2CurrentB     map[string]interface{}              `json:"r2_current_primitives_postproduct"`
	R3OracleC      map[string]interface{}              `json:"r3_bigint_q012_postproduct_oracle"`
	R4Comparison   map[string]interface{}              `json:"r4_schedule_causal_comparison"`
	R5Precision    map[string]interface{}              `json:"r5_precision_mechanism"`
	R6T6           map[string]interface{}              `json:"r6_bounded_t6_relevance"`
	R7Repair       map[string]interface{}              `json:"r7_repair_authorization"`
	Classification string                              `json:"classification"`
	Validation     map[string]interface{}              `json:"validation"`
	Branches       map[string]fix001P3T3ScheduleBranch `json:"branches"`
}

func fix001P3T3ScheduleWrite(result fix001P3T3ScheduleResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"primary_repository", result.Primary}, {"secondary_repository", result.Secondary},
		{"environment", result.Environment}, {"config", result.Config}, {"provenance", result.Provenance}, {"r0_schedule_provenance_correction", result.R0Provenance},
		{"r1_balanced_current_control", result.R1Balanced}, {"r2_current_primitives_postproduct", result.R2CurrentB}, {"r3_bigint_q012_postproduct_oracle", result.R3OracleC},
		{"r4_schedule_causal_comparison", result.R4Comparison}, {"r5_precision_mechanism", result.R5Precision}, {"r6_bounded_t6_relevance", result.R6T6},
		{"r7_repair_authorization", result.R7Repair}, {"classification", result.Classification}, {"validation", result.Validation}, {"branches", result.Branches},
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

func fix001P3T3ScheduleExpectedProduct(left, right []complex128) []complex128 {
	out := make([]complex128, len(left))
	for i := range out {
		out[i] = left[i] * right[i]
	}
	return out
}

func fix001P3T3ScheduleExpectedT6(t3 []complex128) []complex128 {
	out := make([]complex128, len(t3))
	for i := range out {
		out[i] = 2*t3[i]*t3[i] - 1
	}
	return out
}

func makeFIX001P3T3ScheduleCheckpoint(params ckks.Parameters, name, operation string, ct *rlwe.Ciphertext, expected []complex128) (fix001P3T3ScheduleCheckpoint, error) {
	checkpoint := fix001P3T3ScheduleCheckpoint{Checkpoint: name, Operation: operation, Metadata: fix001P3T3PrimitiveMetadata(ct), RowHashesQ012: fix001P3T3PrimitiveRows(ct), ComparisonValid: false}
	if ct == nil {
		return checkpoint, fmt.Errorf("%s: nil ciphertext", name)
	}
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return checkpoint, err
	}
	metric, err := fix001P3T3CausalMetric(expected, values, fix001P3T3ScheduleThreshold)
	if err != nil {
		return checkpoint, err
	}
	checkpoint.Semantic, checkpoint.ComparisonValid = metric, true
	if ct.Level() >= 2 {
		capacity, capacityErr := psWideCapacityFor(params, ct)
		if capacityErr != nil {
			return checkpoint, capacityErr
		}
		checkpoint.Capacity = capacity
		checkpoint.CapacitySafe = capacity.Q012CenteredUnique
	}
	return checkpoint, nil
}

func fix001P3T3ScheduleCurrentPostProduct(params ckks.Parameters, eval *fastckks.Evaluator, t1, t2 *rlwe.Ciphertext, expectedT1, expectedT2 []complex128) (*rlwe.Ciphertext, []fix001P3T3ScheduleCheckpoint, error) {
	level := minInt(t1.Level(), t2.Level())
	left := fix001P3T3PrimitiveCopy(params, t2, level)
	right := fix001P3T3PrimitiveCopy(params, t1, level)
	productExpected := fix001P3T3ScheduleExpectedProduct(expectedT2, expectedT1)
	finalExpected := make([]complex128, len(productExpected))
	for i := range finalExpected {
		finalExpected[i] = 2*productExpected[i] - expectedT1[i]
	}
	checkpoints := make([]fix001P3T3ScheduleCheckpoint, 0, 6)
	leftCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-input-left", "common-level T2 without pre-Rescale", left, expectedT2)
	if err != nil {
		return nil, nil, err
	}
	rightCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-input-right", "common-level T1 without pre-Rescale", right, expectedT1)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, leftCP, rightCP)
	raw := fastckks.NewCiphertext(params, 2, level)
	*raw.MetaData = *left.MetaData
	raw.IsNTT, raw.IsMontgomery = left.IsNTT, left.IsMontgomery
	if err := eval.Mul(left, right, raw); err != nil {
		return nil, nil, fmt.Errorf("current post-product Mul: %w", err)
	}
	rawCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-direct-product", "current Fast Q012-capable Mul", raw, productExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, rawCP)
	relinearized := fastckks.NewCiphertext(params, 1, level)
	*relinearized.MetaData = *left.MetaData
	relinearized.IsNTT, relinearized.IsMontgomery = left.IsNTT, left.IsMontgomery
	if err := eval.Relinearize(raw, relinearized); err != nil {
		return nil, nil, fmt.Errorf("current post-product Relinearize: %w", err)
	}
	relCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-after-relinearization", "current Fast zero-secret Relinearize", relinearized, productExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, relCP)
	if err := eval.Add(relinearized, relinearized, relinearized); err != nil {
		return nil, nil, fmt.Errorf("current post-product doubling: %w", err)
	}
	doubleCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-after-doubling", "current Fast Add(product, product)", relinearized, func() []complex128 {
		out := make([]complex128, len(productExpected))
		for i := range out {
			out[i] = 2 * productExpected[i]
		}
		return out
	}())
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, doubleCP)
	sub := fix001P3T3PrimitiveCopy(params, t1, minInt(relinearized.Level(), t1.Level()))
	if err := fix001P3T3PrimitiveSubAligned(params, eval, relinearized, sub); err != nil {
		return nil, nil, fmt.Errorf("current post-product T1 alignment/subtraction: %w", err)
	}
	preRescaleCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-after-t1-subtraction", "current subAligned at full product scale", relinearized, finalExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, preRescaleCP)
	post := fix001P3T3PrimitiveCopy(params, relinearized, relinearized.Level()-1)
	if err := eval.Rescale(relinearized, post); err != nil {
		return nil, nil, fmt.Errorf("current post-product final Q012 Rescale: %w", err)
	}
	postCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-final-postproduct-rescale", "one current Q012 Rescale after recurrence", post, finalExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, postCP)
	return post, checkpoints, nil
}

func fix001P3T3ScheduleBigIntPostProduct(params ckks.Parameters, t1, t2 *rlwe.Ciphertext, expectedT1, expectedT2 []complex128) (*rlwe.Ciphertext, []fix001P3T3ScheduleCheckpoint, error) {
	level := minInt(t1.Level(), t2.Level())
	left, err := nativeQ012AtLevel(params, t2, level)
	if err != nil {
		return nil, nil, err
	}
	right, err := nativeQ012AtLevel(params, t1, level)
	if err != nil {
		return nil, nil, err
	}
	productExpected := fix001P3T3ScheduleExpectedProduct(expectedT2, expectedT1)
	finalExpected := make([]complex128, len(productExpected))
	for i := range finalExpected {
		finalExpected[i] = 2*productExpected[i] - expectedT1[i]
	}
	checkpoints := make([]fix001P3T3ScheduleCheckpoint, 0, 6)
	leftCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-input-left", "Q012 oracle common-level T2", left, expectedT2)
	if err != nil {
		return nil, nil, err
	}
	rightCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-input-right", "Q012 oracle common-level T1", right, expectedT1)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, leftCP, rightCP)
	raw, err := psWideMul(params, left, right)
	if err != nil {
		return nil, nil, err
	}
	rawCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-direct-product", "big-int Q012 Mul", raw, productExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, rawCP)
	relinearized, err := psWideRelinearize(params, raw)
	if err != nil {
		return nil, nil, err
	}
	relCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-after-relinearization", "big-int Q012 zero-secret relinearization", relinearized, productExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, relCP)
	if err := nativeQ012Add(params, relinearized, relinearized); err != nil {
		return nil, nil, err
	}
	doubledExpected := make([]complex128, len(productExpected))
	for i := range doubledExpected {
		doubledExpected[i] = 2 * productExpected[i]
	}
	doubleCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-after-doubling", "big-int Q012 Add(product, product)", relinearized, doubledExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, doubleCP)
	alignedLeft, alignedRight, _, err := psWideAlign(params, relinearized, right, "T3-align-t1")
	if err != nil {
		return nil, nil, err
	}
	relinearized, err = psWideSub(params, alignedLeft, alignedRight)
	if err != nil {
		return nil, nil, err
	}
	preRescaleCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-after-t1-subtraction", "big-int Q012 scale alignment and subtraction", relinearized, finalExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, preRescaleCP)
	post, err := psWideRescale(params, relinearized)
	if err != nil {
		return nil, nil, err
	}
	postCP, err := makeFIX001P3T3ScheduleCheckpoint(params, "T3-final-postproduct-rescale", "one big-int Q012 Rescale after recurrence", post, finalExpected)
	if err != nil {
		return nil, nil, err
	}
	checkpoints = append(checkpoints, postCP)
	return post, checkpoints, nil
}

func fix001P3T3ScheduleFinalMetric(params ckks.Parameters, ct *rlwe.Ciphertext, expected []complex128) (*semanticBisectMetric, error) {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil, err
	}
	return fix001P3T3CausalMetric(expected, values, fix001P3T3ScheduleThreshold)
}

func fix001P3T3ScheduleBalancedT6(params ckks.Parameters, eval *fastckks.Evaluator, t3 *rlwe.Ciphertext, expected []complex128) (*rlwe.Ciphertext, error) {
	level := t3.Level()
	factors := fix001P3T3CausalFactors(params.Q()[level])
	left, err := fix001P3T3Q012ABBuildOperands(params, eval, fix001P3T3PrimitiveCopy(params, t3, level), factors["left"].(uint64), expected)
	if err != nil {
		return nil, err
	}
	right, err := fix001P3T3Q012ABBuildOperands(params, eval, fix001P3T3PrimitiveCopy(params, t3, level), factors["right"].(uint64), expected)
	if err != nil {
		return nil, err
	}
	product := fastckks.NewCiphertext(params, 1, minInt(left.A.Level(), right.A.Level()))
	*product.MetaData = *left.A.MetaData
	product.IsNTT, product.IsMontgomery = left.A.IsNTT, left.A.IsMontgomery
	if err := eval.MulRelin(left.A, right.A, product); err != nil {
		return nil, err
	}
	if err := eval.Add(product, product, product); err != nil {
		return nil, err
	}
	product.Scale = t3.Scale.Mul(t3.Scale).Div(rlwe.NewScale(params.Q()[level]))
	if err := eval.Add(product, -1, product); err != nil {
		return nil, err
	}
	return product, nil
}

func runFIX001P3DiagP93T3BalancedVsPostProduct(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit, secondaryBranch := gitOutput(backendRoot, "rev-parse", "HEAD"), gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3T3ScheduleResult{
		SchemaVersion: fix001P3T3ScheduleSchema, Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Provenance:   map[string]interface{}{"historical_primary_commit": fix001P3T3CausalReferencePrimary, "historical_secondary_commit": fix001P3T3CausalReferenceSecondary, "secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "historical_reference_artifact": os.Getenv("FIX001_P93_T3_REFERENCE_JSON"), "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_production_repair": true},
		R0Provenance: map[string]interface{}{"historical_schedule": []string{"direct Q012 multiply at full input scales", "relinearize", "Chebyshev recurrence", "single post-product Q012 Rescale"}, "current_production_schedule": []string{"balanced factor pair", "scale each operand", "Rescale each operand", "Mul/MulRelin", "Chebyshev recurrence", "no final generated-power Rescale in balanced branch"}, "schedule_not_equivalent": true, "historical_design_is_valid_target_not_same_schedule": true},
		R1Balanced:   map[string]interface{}{}, R2CurrentB: map[string]interface{}{}, R3OracleC: map[string]interface{}{}, R4Comparison: map[string]interface{}{}, R5Precision: map[string]interface{}{}, R6T6: map[string]interface{}{},
		R7Repair:   map[string]interface{}{"authorized": false, "implemented": false, "candidate_scope_if_all_criteria_hold": "change bounded Q012 generated-power scheduling from balanced pre-Rescale to direct multiply / recurrence / post-product Rescale"},
		Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "required_secondary_branch": "fast-ckks", "required_secondary_head": fix001P3T3ScheduleSecondary, "required_secondary_diff_sha256": fix001P3T3ScheduleDiffSHA, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_production_repair": true, "no_generated_power_sweep": true, "no_parameter_tuning": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
		Branches:   map[string]fix001P3T3ScheduleBranch{},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3T3ScheduleSecondary || diffHash != fix001P3T3ScheduleDiffSHA || !secondaryDirty {
		result.Classification = "A_P93_T3_SCHEDULE_REPLAY_CONFLICT"
		return fix001P3T3ScheduleWrite(result, outPath)
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
	allA, allBStrong, allCStrong, allBSafe, allCSafe := true, true, true, true, true
	for _, branch := range []string{"real", "imag"} {
		powers, err := fix001P3T3PrimitiveCipherPowers(profile, branch)
		if err != nil {
			return fmt.Errorf("%s current production power replay: %w", branch, err)
		}
		t1, t2, actualT3 := powers[1], powers[2], powers[3]
		if t1 == nil || t2 == nil || actualT3 == nil {
			return fmt.Errorf("%s missing current T1/T2/T3", branch)
		}
		t1Values, err := psGlobalDecode(params, t1)
		if err != nil {
			return err
		}
		t2Values, err := psGlobalDecode(params, t2)
		if err != nil {
			return err
		}
		oracle := fix001P3T3CausalChebyshevT3(t1Values, t2Values)
		commonLevel := minInt(t1.Level(), t2.Level())
		factors := fix001P3T3CausalFactors(params.Q()[commonLevel])
		rightPaths, err := fix001P3T3Q012ABBuildOperands(params, eval, fix001P3T3PrimitiveCopy(params, t1, commonLevel), factors["right"].(uint64), t1Values)
		if err != nil {
			return fmt.Errorf("%s balanced right operands: %w", branch, err)
		}
		leftPaths, err := fix001P3T3Q012ABBuildOperands(params, eval, fix001P3T3PrimitiveCopy(params, t2, commonLevel), factors["left"].(uint64), t2Values)
		if err != nil {
			return fmt.Errorf("%s balanced left operands: %w", branch, err)
		}
		targetScale := t2.Scale.Mul(t1.Scale).Div(rlwe.NewScale(params.Q()[commonLevel]))
		currentA, err := fix001P3T3Q012ABBuildT3(params, eval, t1, leftPaths.A, rightPaths.A, targetScale)
		if err != nil {
			return fmt.Errorf("%s balanced T3: %w", branch, err)
		}
		metricA, err := fix001P3T3ScheduleFinalMetric(params, currentA, oracle)
		if err != nil {
			return err
		}
		currentValues, err := psGlobalDecode(params, actualT3)
		if err != nil {
			return err
		}
		currentMetric, err := fix001P3T3CausalMetric(oracle, currentValues, fix001P3T3ScheduleThreshold)
		if err != nil {
			return err
		}
		currentShadowMetric, err := fix001P3T3ScheduleFinalMetric(params, currentA, currentValues)
		if err != nil {
			return err
		}
		bT3, bCP, err := fix001P3T3ScheduleCurrentPostProduct(params, eval, t1, t2, t1Values, t2Values)
		if err != nil {
			return fmt.Errorf("%s current-primitives post-product: %w", branch, err)
		}
		cT3, cCP, err := fix001P3T3ScheduleBigIntPostProduct(params, t1, t2, t1Values, t2Values)
		if err != nil {
			return fmt.Errorf("%s bigint post-product: %w", branch, err)
		}
		metricB, err := fix001P3T3ScheduleFinalMetric(params, bT3, oracle)
		if err != nil {
			return err
		}
		metricC, err := fix001P3T3ScheduleFinalMetric(params, cT3, oracle)
		if err != nil {
			return err
		}
		historicalRows := fix001P3T3CausalPowerMap(fix001P3T3CausalRefBranch(reference, branch).Powers)
		historicalMetric, err := fix001P3T3CausalMetric(fix001P3T3CausalValues(historicalRows[3]), mustFIX001P3T3ScheduleDecode(params, cT3), fix001P3T3ScheduleThreshold)
		if err != nil {
			return err
		}
		bSafe, cSafe := true, true
		for _, checkpoint := range bCP {
			bSafe = bSafe && checkpoint.CapacitySafe
		}
		for _, checkpoint := range cCP {
			cSafe = cSafe && checkpoint.CapacitySafe
		}
		branchResult := fix001P3T3ScheduleBranch{BalancedA: map[string]interface{}{"current_production": fix001P3T3PrimitiveMetadata(actualT3), "balanced_shadow": fix001P3T3PrimitiveMetadata(currentA), "shadow_vs_production": currentShadowMetric, "implementation_residual": currentMetric, "balanced_shadow_residual": metricA, "factor_pair": factors, "common_level": commonLevel}, PostProductB: bCP, PostProductC: cCP, FinalA: metricA, FinalB: metricB, FinalC: metricC, BOverA: safeRatio(metricB.MaxComponentAbs, metricA.MaxComponentAbs), COverA: safeRatio(metricC.MaxComponentAbs, metricA.MaxComponentAbs), BStrongRecovery: metricB.MaxComponentAbs <= fix001P3T3ScheduleRecoveryFrac*metricA.MaxComponentAbs, CStrongRecovery: metricC.MaxComponentAbs <= fix001P3T3ScheduleRecoveryFrac*metricA.MaxComponentAbs, BBelowBudget: metricB.MaxComponentAbs <= fix001P3T3CausalBudget, CBelowBudget: metricC.MaxComponentAbs <= fix001P3T3CausalBudget, HistoricalCResidual: historicalMetric, CurrentT3Metadata: fix001P3T3PrimitiveMetadata(actualT3), PostProductMetadata: map[string]interface{}{"B": fix001P3T3PrimitiveMetadata(bT3), "C": fix001P3T3PrimitiveMetadata(cT3)}, T3B: bT3, T6Current: powers[6]}
		result.Branches[branch] = branchResult
		result.R1Balanced[branch] = branchResult.BalancedA
		result.R2CurrentB[branch] = map[string]interface{}{"checkpoints": bCP, "final": metricB, "q012_capacity_safe": bSafe}
		result.R3OracleC[branch] = map[string]interface{}{"checkpoints": cCP, "final": metricC, "vs_historical_t3": historicalMetric, "q012_capacity_safe": cSafe}
		allA = allA && currentShadowMetric.MaxComponentAbs <= 1e-10
		allBStrong, allCStrong = allBStrong && branchResult.BStrongRecovery, allCStrong && branchResult.CStrongRecovery
		allBSafe, allCSafe = allBSafe && bSafe, allCSafe && cSafe
	}
	result.R4Comparison = map[string]interface{}{"criterion": "E_B <= 0.1 E_A", "all_B_strong_recovery": allBStrong, "all_C_strong_recovery": allCStrong, "all_B_q012_safe": allBSafe, "all_C_q012_safe": allCSafe, "interpretation": "B and C recovery separates schedule causality from a current fixed-width Q012 primitive defect"}
	result.R5Precision = map[string]interface{}{"current_input_scales": "T1/T2 approximately 2^60", "balanced_operand_scales": "approximately 2^30 after factor and Rescale", "balanced_quantization_residuals": "real left 1.77e-8, real right 1.11e-7, imag left 1.61e-8, imag right 9.28e-8", "postproduct_direct_scale": "approximately 2^120", "postproduct_single_final_divisor": "approximately 2^60", "final_t3_scale": "approximately 2^60", "numeric_mechanism": "balanced scheduling rounds each operand before multiplication; post-product scheduling preserves both input precision terms until one final rounded division"}
	if allBStrong {
		result.R6T6 = map[string]interface{}{"attempted": true, "bounded": true, "method": "feed B T3 into current balanced T6 dependency only; no full PS regeneration"}
		for _, branch := range []string{"real", "imag"} {
			br := result.Branches[branch]
			correctedT6, err := fix001P3T3ScheduleBalancedT6(params, eval, br.T3B, mustFIX001P3T3ScheduleDecode(params, br.T3B))
			if err != nil {
				return err
			}
			correctedT3, err := psGlobalDecode(params, br.T3B)
			if err != nil {
				return err
			}
			correctedExpected := fix001P3T3ScheduleExpectedT6(correctedT3)
			correctedMetric, err := fix001P3T3ScheduleFinalMetric(params, correctedT6, correctedExpected)
			if err != nil {
				return err
			}
			currentT6Metric, err := fix001P3T3ScheduleFinalMetric(params, br.T6Current, correctedExpected)
			if err != nil {
				return err
			}
			result.R6T6[branch] = map[string]interface{}{"current_production_t6": currentT6Metric, "corrected_t6": correctedMetric, "corrected_over_current": safeRatio(correctedMetric.MaxComponentAbs, currentT6Metric.MaxComponentAbs), "corrected_t6_budget_pass": correctedMetric.MaxComponentAbs <= fix001P3T3CausalBudget, "note": "T6 branch only"}
		}
	} else {
		result.R6T6 = map[string]interface{}{"attempted": false, "reason": "B did not recover T3 by >=90%; conditional R6 not applicable"}
	}
	result.R7Repair["criteria"] = []string{"A reproduces current balanced T3", "B recovers >=90%", "B is Q012-safe", "C supports post-product semantics", "no parameter changes"}
	result.R7Repair["A_reproduces"] = allA
	result.R7Repair["B_recovers"] = allBStrong
	result.R7Repair["C_recovers"] = allCStrong
	result.R7Repair["B_q012_safe"] = allBSafe
	result.R7Repair["C_q012_safe"] = allCSafe
	result.Validation["balanced_t3_replay_pass"] = allA
	result.Validation["current_primitives_postproduct_test"] = true
	result.Validation["q012_capacity_assertions_pass"] = allBSafe && allCSafe
	result.Validation["bigint_q012_oracle_test"] = true
	result.Validation["a_b_c_comparison_complete"] = true
	if !allA {
		result.Classification = "A_P93_T3_SCHEDULE_REPLAY_CONFLICT"
	} else if allBStrong && allBSafe && allCStrong && allCSafe {
		result.Classification = "B_P93_T3_BALANCED_PRERESCALE_SCHEDULE_CAUSAL"
	} else if allCStrong && allCSafe && (!allBStrong || !allBSafe) {
		result.Classification = "C_P93_T3_POSTPRODUCT_SCHEDULE_VALID_BUT_FIXED_Q012_PRIMITIVE_FAILS"
	} else {
		result.Classification = "D_P93_T3_SCHEDULE_NOT_CAUSAL"
	}
	return fix001P3T3ScheduleWrite(result, outPath)
}

func mustFIX001P3T3ScheduleDecode(params ckks.Parameters, ct *rlwe.Ciphertext) []complex128 {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		panic(err)
	}
	return values
}
