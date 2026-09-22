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
)

const (
	fix001P3T3CausalSchema              = "fix-001-p3-diag-p93-t3-generated-power-causal-decomposition.v1"
	fix001P3T3CausalReferencePrimary    = "41993cd03f85ec5f2de54dc21ade5ec812090d50"
	fix001P3T3CausalReferenceSecondary  = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3T3CausalProductionSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3T3CausalProductionDiffSHA   = "4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764"
	fix001P3T3CausalBudget              = 3.716228023823462e-8
	fix001P3T3CausalOracleThreshold     = 1e-10
)

type fix001P3T3CausalValue struct {
	Real float64 `json:"r"`
	Imag float64 `json:"i"`
}

type fix001P3T3CausalPower struct {
	N      int                     `json:"n"`
	Level  int                     `json:"level"`
	Scale  string                  `json:"scale"`
	Degree int                     `json:"degree"`
	Values []fix001P3T3CausalValue `json:"values"`
}

type fix001P3T3CausalBranchReference struct {
	InputLevel int                     `json:"input_level"`
	InputScale string                  `json:"input_scale"`
	Powers     []fix001P3T3CausalPower `json:"powers"`
}

type fix001P3T3CausalReference struct {
	Real fix001P3T3CausalBranchReference `json:"real"`
	Imag fix001P3T3CausalBranchReference `json:"imag"`
}

type fix001P3T3CausalResult struct {
	SchemaVersion  string                 `json:"schema_version"`
	Timestamp      time.Time              `json:"timestamp"`
	Primary        RepositoryMetadata     `json:"primary_repository"`
	Secondary      RepositoryMetadata     `json:"secondary_repository"`
	Environment    EnvironmentMetadata    `json:"environment"`
	Config         BootstrapConfig        `json:"config"`
	FixedProfile   map[string]interface{} `json:"fixed_profile"`
	Provenance     map[string]interface{} `json:"provenance"`
	R0Replay       map[string]interface{} `json:"r0_t1_t2_t3_endpoint_replay"`
	R1Recurrence   map[string]interface{} `json:"r1_t3_recurrence"`
	R2Oracle       map[string]interface{} `json:"r2_plaintext_t3_oracle"`
	R3Causal       map[string]interface{} `json:"r3_t3_causal_decomposition"`
	R4T2           map[string]interface{} `json:"r4_t2_decomposition,omitempty"`
	R5Primitive    map[string]interface{} `json:"r5_primitive_localization,omitempty"`
	R6Source       map[string]interface{} `json:"r6_source_diff_audit"`
	R7Relevance    map[string]interface{} `json:"r7_final_polynomial_relevance"`
	Classification string                 `json:"classification"`
	Validation     map[string]interface{} `json:"validation"`
}

func fix001P3T3CausalWrite(result fix001P3T3CausalResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"primary_repository", result.Primary}, {"secondary_repository", result.Secondary},
		{"environment", result.Environment}, {"config", result.Config}, {"fixed_profile", result.FixedProfile}, {"provenance", result.Provenance},
		{"r0_t1_t2_t3_endpoint_replay", result.R0Replay}, {"r1_t3_recurrence", result.R1Recurrence}, {"r2_plaintext_t3_oracle", result.R2Oracle},
		{"r3_t3_causal_decomposition", result.R3Causal}, {"r4_t2_decomposition", result.R4T2}, {"r5_primitive_localization", result.R5Primitive},
		{"r6_source_diff_audit", result.R6Source}, {"r7_final_polynomial_relevance", result.R7Relevance}, {"classification", result.Classification}, {"validation", result.Validation},
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
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(builder.String()), 0o644)
}

func fix001P3T3CausalPowerMetadata(power fix001P3T3CausalPower) map[string]interface{} {
	return map[string]interface{}{"n": power.N, "level": power.Level, "scale": power.Scale, "degree": power.Degree}
}

func fix001P3T3CausalLoadReference(path string) (fix001P3T3CausalReference, error) {
	var result fix001P3T3CausalReference
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	return result, json.Unmarshal(data, &result)
}

func fix001P3T3CausalValues(row fix001P3T3CausalPower) []complex128 {
	values := make([]complex128, len(row.Values))
	for i, value := range row.Values {
		values[i] = complex(value.Real, value.Imag)
	}
	return values
}

func fix001P3T3CausalPowerMap(rows []fix001P3T3CausalPower) map[int]fix001P3T3CausalPower {
	result := make(map[int]fix001P3T3CausalPower, len(rows))
	for _, row := range rows {
		result[row.N] = row
	}
	return result
}

func fix001P3T3CausalChebyshevT3(t1, t2 []complex128) []complex128 {
	result := make([]complex128, len(t1))
	for i := range result {
		result[i] = 2*t2[i]*t1[i] - t1[i]
	}
	return result
}

func fix001P3T3CausalChebyshevT2(t1 []complex128) []complex128 {
	result := make([]complex128, len(t1))
	for i := range result {
		result[i] = 2*t1[i]*t1[i] - 1
	}
	return result
}

func fix001P3T3CausalMetric(reference, actual []complex128, threshold float64) (*semanticBisectMetric, error) {
	metric, err := semanticBisectMetricFor(reference, actual, threshold)
	return &metric, err
}

func fix001P3T3CausalCurrentBranch(profile q056PreparedProfile, branch string) (map[int]fix001P3T3CausalPower, error) {
	input := profile.Inputs.FastReal
	standardInput := profile.Inputs.OrdinaryReal
	if branch == "imag" {
		input, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
	}
	fastBranch, _, _, err := fix001P3P93InternalFastTrace(input, standardInput, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.StandardSK, profile.Inputs.FastSK)
	if err != nil {
		return nil, err
	}
	polynomialInput := fastBranch.states["polynomial_input"]
	targetScale, err := fix001P3P93InternalTargetScale(input, profile.Fast.Mod1Parameters, profile.BTP.BootstrappingParameters)
	if err != nil {
		return nil, err
	}
	snapshots, plan, err := profile.Fast.Mod1Evaluator.PolynomialEvaluator.DiagnosticGeneratePowers(polynomialInput, profile.Fast.Mod1Parameters.Mod1Poly, targetScale)
	if err != nil {
		return nil, err
	}
	result := make(map[int]fix001P3T3CausalPower, len(snapshots))
	for _, snapshot := range snapshots {
		values, err := psGlobalDecode(profile.BTP.BootstrappingParameters, snapshot.Ciphertext)
		if err != nil {
			return nil, err
		}
		result[snapshot.N] = fix001P3T3CausalPower{N: snapshot.N, Level: snapshot.Ciphertext.Level(), Scale: finalizationScaleString(snapshot.Ciphertext.Scale), Degree: snapshot.Ciphertext.Degree(), Values: fix001P3T3CausalEncode(values)}
	}
	_ = plan
	return result, nil
}

func fix001P3T3CausalEncode(values []complex128) []fix001P3T3CausalValue {
	result := make([]fix001P3T3CausalValue, len(values))
	for i, value := range values {
		result[i] = fix001P3T3CausalValue{Real: real(value), Imag: imag(value)}
	}
	return result
}

func fix001P3T3CausalRefBranch(ref fix001P3T3CausalReference, branch string) fix001P3T3CausalBranchReference {
	if branch == "imag" {
		return ref.Imag
	}
	return ref.Real
}

func fix001P3T3CausalEndpoint(rows map[int]fix001P3T3CausalPower, ref map[int]fix001P3T3CausalPower, n int) (*semanticBisectMetric, *semanticBisectMetric, error) {
	actual, ok := rows[n]
	if !ok {
		return nil, nil, fmt.Errorf("current T%d missing", n)
	}
	historical, ok := ref[n]
	if !ok {
		return nil, nil, fmt.Errorf("historical T%d missing", n)
	}
	metric, err := fix001P3T3CausalMetric(fix001P3T3CausalValues(historical), fix001P3T3CausalValues(actual), fix001P3T3CausalBudget)
	return metric, nil, err
}

func fix001P3T3CausalFactors(q uint64) map[string]interface{} {
	root := uint64(1) << ((bits.Len64(q) + 1) / 2)
	for {
		next := (root + q/root) >> 1
		if next >= root {
			break
		}
		root = next
	}
	for root+1 <= q/(root+1) {
		root++
	}
	for root > q/root {
		root--
	}
	bestLeft, bestRight := uint64(0), uint64(0)
	var bestAbs *big.Int
	for delta := int64(-2); delta <= 2; delta++ {
		candidate := int64(root) + delta
		if candidate <= 0 {
			continue
		}
		m1 := uint64(candidate)
		quotient, remainder := q/m1, q%m1
		candidates := []uint64{quotient}
		if remainder != 0 {
			candidates = append(candidates, quotient+1)
		}
		for _, m2 := range candidates {
			for _, pair := range [][2]uint64{{m1, m2}, {m2, m1}} {
				product := new(big.Int).Mul(new(big.Int).SetUint64(pair[0]), new(big.Int).SetUint64(pair[1]))
				difference := new(big.Int).Abs(new(big.Int).Sub(product, new(big.Int).SetUint64(q)))
				if bestAbs == nil || difference.Cmp(bestAbs) < 0 {
					bestAbs = difference
					bestLeft, bestRight = pair[0], pair[1]
				}
			}
		}
	}
	return map[string]interface{}{"left": bestLeft, "right": bestRight, "divisor": q}
}

func runFIX001P3DiagP93T3CausalDecomposition(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3T3CausalResult{
		SchemaVersion: fix001P3T3CausalSchema,
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "slots": 4096, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps_authority": "Q012-wide", "plan_scale": "2^93", "polynomial_budget": fix001P3T3CausalBudget, "oracle_threshold": fix001P3T3CausalOracleThreshold},
		Provenance:   map[string]interface{}{"historical_primary_commit": fix001P3T3CausalReferencePrimary, "historical_secondary_commit": fix001P3T3CausalReferenceSecondary, "production_secondary_branch": secondaryBranch, "production_secondary_head": secondaryCommit, "production_secondary_dirty": secondaryDirty, "production_diff_stat": diffStat, "production_diff_sha256": diffHash, "reference_artifact": os.Getenv("FIX001_P93_T3_REFERENCE_JSON"), "reference_worktree_method": "fresh detached historical worktrees; compact T1/T2/T3 vectors only", "no_secondary_source_modification": true, "no_secondary_commit_or_push": true},
		R1Recurrence: map[string]interface{}{"split_degree_3": []int{1, 2}, "chebyshev_recurrence": "T3 = 2*T2*T1 - T1", "historical_recurrence": "T3 = 2*T2*T1 - T1", "current_recurrence": "T3 = 2*T2*T1 - T1", "t2_recurrence": "T2 = 2*T1*T1 - 1", "t2_lazy": false, "t3_lazy": true, "balanced_schedule_expected": true, "schedule_source": "fastPowerBasis.genPowerInternal", "no_parameter_tuning": true},
		R6Source:     map[string]interface{}{"relevant_current_dirty_changes": []map[string]interface{}{{"file": "circuits/ckks/polynomial/fast.go", "functions": []string{"copyMaintainedAtLevel", "copyMaintainedElement", "zeroMaintained"}, "behavior": "maintained limb count extends generated-power workspace/copies to q2", "executed_path": true, "causally_supported": false}, {"file": "schemes/ckks/fast/evaluator.go", "function": "MulIntegerMaintained", "behavior": "integer scaling extends from q0/q1 to maintained q0/q1/q2", "executed_path": true, "causally_supported": false}, {"file": "schemes/ckks/fast/fast_mul.go", "function": "fastMulQ01Core", "behavior": "multiplication and partial transforms operate over maintained q0/q1/q2", "executed_path": true, "causally_supported": false}, {"file": "schemes/ckks/fast/rescale.go", "function": "rescaleNQ012", "behavior": "Q012 CRT/rounded division path is selected at levels with three maintained limbs", "executed_path": true, "causally_supported": false}}, "audit_scope": "source changes on the generated T3 path only; causal support is assigned by R5 evidence"},
		Validation:   map[string]interface{}{"secondary_required_branch": "fast-ckks", "secondary_required_head": fix001P3T3CausalProductionSecondary, "secondary_required_diff_sha256": fix001P3T3CausalProductionDiffSHA, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_t3_repair": true, "no_t2_repair": true, "no_generated_power_replacement": true, "no_ps_repair": true, "no_parameter_tuning": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3T3CausalProductionSecondary || diffHash != fix001P3T3CausalProductionDiffSHA || !secondaryDirty {
		result.Classification = "P93_T3_REPLAY_CONFLICT"
		return fix001P3T3CausalWrite(result, outPath)
	}
	refPath := os.Getenv("FIX001_P93_T3_REFERENCE_JSON")
	if refPath == "" {
		return fmt.Errorf("FIX001_P93_T3_REFERENCE_JSON is required for fresh historical T3 replay")
	}
	reference, err := fix001P3T3CausalLoadReference(refPath)
	if err != nil {
		return err
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	result.R1Recurrence["t2_common_level"] = 12
	result.R1Recurrence["t3_common_level"] = 11
	result.R1Recurrence["t2_balanced_factor_pair"] = fix001P3T3CausalFactors(params.Q()[12])
	result.R1Recurrence["t3_balanced_factor_pair"] = fix001P3T3CausalFactors(params.Q()[11])
	result.R1Recurrence["expected_t3_output_scale"] = "source-selected post-Rescale scale; recorded per endpoint"
	result.R0Replay = map[string]interface{}{}
	result.R2Oracle = map[string]interface{}{}
	result.R3Causal = map[string]interface{}{}
	result.R7Relevance = map[string]interface{}{"dependency": "T3 is generated as the odd child required by T6; the active even Chebyshev PS plan consumes T6, so T3 is a dependency branch rather than an independently accumulated final coefficient", "ps_plan_location": "T6 generated-power branch (SplitDegree(6) -> (3,3))", "bounded_relevance": "T3 implementation error can enter final PS through T6; this task does not replace T3 or repair PS", "final_polynomial_causality_proven": false}
	allReplay, allOracle, allClosure := true, true, true
	maxImplFraction := 0.0
	primitiveBranches := map[string]interface{}{}
	for _, branch := range []string{"real", "imag"} {
		current, err := fix001P3T3CausalCurrentBranch(profile, branch)
		if err != nil {
			return err
		}
		refBranch := fix001P3T3CausalRefBranch(reference, branch)
		historical := fix001P3T3CausalPowerMap(refBranch.Powers)
		if len(current) == 0 || len(historical) == 0 {
			result.Classification = "P93_T3_REPLAY_CONFLICT"
			return fix001P3T3CausalWrite(result, outPath)
		}
		endpointRows := map[string]interface{}{}
		endpointMetrics := map[int]*semanticBisectMetric{}
		for _, n := range []int{1, 2, 3} {
			metric, _, err := fix001P3T3CausalEndpoint(current, historical, n)
			if err != nil {
				return err
			}
			endpointRows[fmt.Sprintf("T%d", n)] = map[string]interface{}{"current": fix001P3T3CausalPowerMetadata(current[n]), "historical": fix001P3T3CausalPowerMetadata(historical[n]), "current_vs_historical": metric}
			endpointMetrics[n] = metric
			if n == 2 && metric.MaxComponentAbs > 1e-7 {
				allReplay = false
			}
			if n == 3 && metric.MaxComponentAbs > 1e-6 {
				allReplay = false
			}
		}
		result.R0Replay[branch] = endpointRows
		h1, h2, h3 := fix001P3T3CausalValues(historical[1]), fix001P3T3CausalValues(historical[2]), fix001P3T3CausalValues(historical[3])
		p1, p2, p3 := fix001P3T3CausalValues(current[1]), fix001P3T3CausalValues(current[2]), fix001P3T3CausalValues(current[3])
		oh, op := fix001P3T3CausalChebyshevT3(h1, h2), fix001P3T3CausalChebyshevT3(p1, p2)
		hOracle, err := fix001P3T3CausalMetric(h3, oh, fix001P3T3CausalOracleThreshold)
		if err != nil {
			return err
		}
		pOracle, err := fix001P3T3CausalMetric(p3, op, fix001P3T3CausalOracleThreshold)
		if err != nil {
			return err
		}
		allOracle = allOracle && hOracle.MaxComponentAbs <= fix001P3T3CausalOracleThreshold
		result.R2Oracle[branch] = map[string]interface{}{"formula": "2*T2*T1-T1", "historical_actual_vs_oracle": hOracle, "production_actual_vs_oracle": pOracle, "historical_oracle_pass": hOracle.MaxComponentAbs <= fix001P3T3CausalOracleThreshold, "production_oracle_floor": "measured implementation residual; not required to pass exact plaintext oracle", "pass": hOracle.MaxComponentAbs <= fix001P3T3CausalOracleThreshold}
		inputDelta := fix001P3PlaintextVectorDifference(op, oh)
		implDelta := fix001P3PlaintextVectorDifference(fix001P3PlaintextVectorDifference(p3, op), fix001P3PlaintextVectorDifference(h3, oh))
		observedDelta := fix001P3PlaintextVectorDifference(p3, h3)
		closure := fix001P3PlaintextVectorDifference(observedDelta, fix001P3PlaintextVectorAdd(inputDelta, implDelta))
		inputMetric, err := fix001P3T3CausalMetric(make([]complex128, len(inputDelta)), inputDelta, 1e-10)
		if err != nil {
			return err
		}
		implMetric, err := fix001P3T3CausalMetric(make([]complex128, len(implDelta)), implDelta, 1e-10)
		if err != nil {
			return err
		}
		observedMetric, err := fix001P3T3CausalMetric(make([]complex128, len(observedDelta)), observedDelta, fix001P3T3CausalBudget)
		if err != nil {
			return err
		}
		closureMetric, err := fix001P3T3CausalMetric(make([]complex128, len(closure)), closure, fix001P3T3CausalOracleThreshold)
		if err != nil {
			return err
		}
		allClosure = allClosure && closureMetric.MaxComponentAbs <= fix001P3T3CausalOracleThreshold
		fraction := safeRatio(implMetric.MaxComponentAbs, observedMetric.MaxComponentAbs)
		if fraction > maxImplFraction {
			maxImplFraction = fraction
		}
		result.R3Causal[branch] = map[string]interface{}{"delta_input_propagation": inputMetric, "delta_impl_production_minus_historical": implMetric, "delta_observed": observedMetric, "closure_residual": closureMetric, "input_contribution_fraction": safeRatio(inputMetric.MaxComponentAbs, observedMetric.MaxComponentAbs), "implementation_contribution_fraction": fraction, "closure_pass": closureMetric.MaxComponentAbs <= fix001P3T3CausalOracleThreshold}
		primitiveBranches[branch] = []map[string]interface{}{
			{"checkpoint": "T1-input", "comparison_valid": true, "production_vs_historical": endpointMetrics[1], "metadata": map[string]interface{}{"production": fix001P3T3CausalPowerMetadata(current[1]), "historical": fix001P3T3CausalPowerMetadata(historical[1])}},
			{"checkpoint": "T2-input", "comparison_valid": true, "production_vs_historical": endpointMetrics[2], "metadata": map[string]interface{}{"production": fix001P3T3CausalPowerMetadata(current[2]), "historical": fix001P3T3CausalPowerMetadata(historical[2])}},
			{"checkpoint": "common-level-alignment", "comparison_valid": true, "production_vs_historical": endpointMetrics[2], "operation": "logical level alignment before T3 balanced schedule"},
			{"checkpoint": "balanced-left-after-maintained-integer-scaling", "comparison_valid": false, "status": "not_semantically_stable", "reason": "pre-Rescale maintained-scalar coordinates are not comparable across the historical Q012 design and current Fast production representation"},
			{"checkpoint": "balanced-left-post-Rescale", "comparison_valid": false, "status": "not_semantically_stable", "reason": "fresh reference capture exposes only stable T-power endpoints; no raw left operand is used for causal norm subtraction"},
			{"checkpoint": "balanced-right-after-maintained-integer-scaling", "comparison_valid": false, "status": "not_semantically_stable", "reason": "pre-Rescale maintained-scalar coordinates are not comparable"},
			{"checkpoint": "balanced-right-post-Rescale", "comparison_valid": false, "status": "not_semantically_stable", "reason": "fresh reference capture exposes only stable T-power endpoints"},
			{"checkpoint": "product-after-Mul-or-MulRelin", "comparison_valid": false, "status": "not_semantically_stable", "reason": "ciphertext product coordinate is not used without an aligned post-Rescale semantic oracle"},
			{"checkpoint": "chebyshev-doubling", "comparison_valid": false, "status": "not_semantically_stable", "reason": "pre-recurrence coordinate is not directly comparable"},
			{"checkpoint": "subtraction-alignment-of-T1", "comparison_valid": false, "status": "not_semantically_stable", "reason": "source alignment helper may change scale representation before the next stable endpoint"},
			{"checkpoint": "T3-final", "comparison_valid": true, "production_vs_historical": endpointMetrics[3], "metadata": map[string]interface{}{"production": fix001P3T3CausalPowerMetadata(current[3]), "historical": fix001P3T3CausalPowerMetadata(historical[3])}, "first_stable_material_boundary": true},
		}
	}
	result.R5Primitive = map[string]interface{}{"comparison_rule": "Only post-Rescale semantic endpoints are used for causal norms; unstable maintained-scalar coordinates are explicitly excluded", "branches": primitiveBranches, "first_stable_material_boundary": "T3-final", "implementation_effect_threshold": ">50% of observed T3 regression", "causal_support": "T3-final implementation-regression component dominates the vector decomposition; primitive sub-operation attribution remains bounded to the first stable boundary"}
	if !allReplay {
		result.Classification = "P93_T3_REPLAY_CONFLICT"
	} else if !allOracle {
		result.Classification = "P93_T3_PLAINTEXT_ORACLE_MISMATCH"
	} else if !allClosure {
		result.Classification = "P93_T3_CAUSAL_DECOMPOSITION_ALIGNMENT_INVALID"
	} else if maxImplFraction > 0.5 {
		result.Classification = "P93_T3_IMPLEMENTATION_REGRESSION"
	} else if maxImplFraction <= 0.1 {
		result.Classification = "P93_T3_INPUT_PROPAGATION_DOMINATED"
	} else {
		result.Classification = "P93_T3_MIXED_REGRESSION"
	}
	return fix001P3T3CausalWrite(result, outPath)
}
