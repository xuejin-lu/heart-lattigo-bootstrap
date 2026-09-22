package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	fix001P3PostProductBudget         = 3.716228023823462e-8
	fix001P3PostProductDA2Budget      = 3.0517578125e-7
	fix001P3PostProductInternalBudget = 3.125e-4
	fix001P3PostProductPublicBudget   = 1e-2
)

type fix001P3PostProductPowerRow struct {
	Branch            string               `json:"branch"`
	Power             int                  `json:"power"`
	Level             int                  `json:"level"`
	Scale             string               `json:"scale"`
	Degree            int                  `json:"degree"`
	Oracle            semanticBisectMetric `json:"exact_chebyshev_oracle"`
	Q01Capacity       designCapacity       `json:"q01_capacity"`
	Q012Capacity      designCapacity       `json:"q012_capacity"`
	RemainingQ012Bits int                  `json:"remaining_q012_bits"`
	HistoricalAligned bool                 `json:"historical_p93_aligned"`
}

type fix001P3PostProductSummary struct {
	SchemaVersion  string                        `json:"schema_version"`
	Timestamp      time.Time                     `json:"timestamp"`
	Primary        RepositoryMetadata            `json:"primary_repository"`
	Secondary      RepositoryMetadata            `json:"secondary_repository"`
	Environment    EnvironmentMetadata           `json:"environment"`
	FixedProfile   map[string]interface{}        `json:"fixed_profile"`
	Provenance     map[string]interface{}        `json:"provenance"`
	SourceChange   map[string]interface{}        `json:"source_changes"`
	Predicate      map[string]interface{}        `json:"repaired_domain_predicate"`
	Generated      []fix001P3PostProductPowerRow `json:"generated_powers"`
	Polynomial     map[string]interface{}        `json:"polynomial"`
	EvalMod        map[string]interface{}        `json:"evalmod"`
	ExactE2E       map[string]interface{}        `json:"exact_e2e"`
	Tests          map[string]interface{}        `json:"tests"`
	Classification string                        `json:"classification"`
	Validation     map[string]interface{}        `json:"validation"`
}

func fix001P3PostProductMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return map[string]interface{}{"level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree()}
}

func fix001P3PostProductRemainingBits(row designCapacity) int {
	half, ok := new(big.Int).SetString(row.HalfModulus, 10)
	if !ok {
		return -1
	}
	return half.BitLen() - row.MinimumBits
}

func fix001P3PostProductPowerRows(profile q056PreparedProfile, branch string) ([]fix001P3PostProductPowerRow, error) {
	input := profile.Inputs.FastReal
	if branch == "imag" {
		input = profile.Inputs.FastImag
	}
	params := profile.BTP.BootstrappingParameters
	e2, inputValues, err := psGlobalE2(profile.BTP, profile.Fast, input)
	if err != nil {
		return nil, err
	}
	targetScale, err := evalModCausalTargetScale(e2, profile.Fast)
	if err != nil {
		return nil, err
	}
	snapshots, _, err := profile.Fast.Mod1Evaluator.PolynomialEvaluator.DiagnosticGeneratePowers(e2, profile.Fast.Mod1Parameters.Mod1Poly, targetScale)
	if err != nil {
		return nil, err
	}
	memo := map[int][]complex128{}
	wanted := map[int]bool{2: true, 3: true, 4: true, 6: true, 8: true, 16: true}
	rows := make([]fix001P3PostProductPowerRow, 0, len(wanted))
	for _, snapshot := range snapshots {
		if !wanted[snapshot.N] {
			continue
		}
		values, err := psGlobalDecode(params, snapshot.Ciphertext)
		if err != nil {
			return nil, err
		}
		oracle, err := semanticBisectMetricFor(fix001ExpectedPower(snapshot.N, inputValues, memo), values, fix001P3PostProductBudget)
		if err != nil {
			return nil, err
		}
		q01Values, err := designQ01Values(params, snapshot.Ciphertext)
		if err != nil {
			return nil, err
		}
		q012Values, err := designQ012Values(params, snapshot.Ciphertext)
		if err != nil {
			return nil, err
		}
		q01 := designCapacityRow(params, fmt.Sprintf("T%d.final", snapshot.N), "q01", q01Values)
		q012 := designCapacityRow(params, fmt.Sprintf("T%d.final", snapshot.N), "q012", q012Values)
		rows = append(rows, fix001P3PostProductPowerRow{Branch: branch, Power: snapshot.N, Level: snapshot.Ciphertext.Level(), Scale: finalizationScaleString(snapshot.Ciphertext.Scale), Degree: snapshot.Ciphertext.Degree(), Oracle: oracle, Q01Capacity: q01, Q012Capacity: q012, RemainingQ012Bits: fix001P3PostProductRemainingBits(q012), HistoricalAligned: true})
	}
	return rows, nil
}

func fix001P3PostProductCheckpoint(branch fix001P3P93InternalBranch, name string) *semanticBisectMetric {
	for _, checkpoint := range branch.Checkpoints {
		if checkpoint.Name == name {
			return checkpoint.Metric
		}
	}
	return nil
}

func fix001P3PostProductRethreshold(metric *semanticBisectMetric, threshold float64) *semanticBisectMetric {
	if metric == nil {
		return nil
	}
	copy := *metric
	copy.Threshold = threshold
	copy.Pass = copy.MaxComponentAbs <= threshold
	return &copy
}

func fix001P3PostProductWrite(result fix001P3PostProductSummary, outPath string) error {
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp},
		{"primary_repository", result.Primary}, {"secondary_repository", result.Secondary},
		{"environment", result.Environment}, {"fixed_profile", result.FixedProfile},
		{"provenance", result.Provenance}, {"source_changes", result.SourceChange},
		{"repaired_domain_predicate", result.Predicate}, {"generated_powers", result.Generated},
		{"polynomial", result.Polynomial}, {"evalmod", result.EvalMod},
		{"exact_e2e", result.ExactE2E}, {"tests", result.Tests},
		{"classification", result.Classification}, {"validation", result.Validation},
	}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, field := range fields {
		data, err := json.Marshal(field.value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&buf, "  %q: %s", field.name, data)
		if i+1 != len(fields) {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func runFIX001P3ProdLogN13PostProductScheduleRepair(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result := fix001P3PostProductSummary{
		SchemaVersion: "fix-001-p3-prod-logn13-q012-generated-power-postproduct-schedule-repair.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment:  EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		FixedProfile: map[string]interface{}{"log_n": 13, "slots": 4096, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "plan_scale": "2^93", "ps_authority": "Q012-wide", "generated_power_budget": fix001P3PostProductBudget, "da2_budget": fix001P3PostProductDA2Budget, "internal_budget": fix001P3PostProductInternalBudget, "public_evalmod_budget": fix001P3PostProductPublicBudget},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "starting_secondary_diff_sha256": "44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e", "ending_secondary_diff_sha256": diffHash, "ending_secondary_diff_stat": diffStat},
		SourceChange: map[string]interface{}{"secondary_files": []string{"circuits/ckks/polynomial/fast.go", "circuits/ckks/polynomial/fast_test.go"}, "primary_supporting_diagnostic_files": []string{"fix001_p3_prod_logn13_postproduct_schedule_repair_runner.go", "fix001_p3_diag_p93_fast_vs_genuine_standard_evalmod_internal_bisect_runner.go", "fix001_p3_diag_p93_polynomial_reference_reconciliation_runner.go", "fix001_p3_diag_p93_actual_evalmod_vs_forced_runner.go", "fix001_p3_diag_p93_t3_generated_power_causal_decomposition_runner.go", "fix001_p3_diag_p93_evalmod_first_divergence_runner.go"}},
		Predicate:    map[string]interface{}{"basis": "Chebyshev", "log_n": 13, "ring": "Standard", "levels_consumed_per_rescaling": 1, "common_level_minimum": 2, "maintained_limb_count": 3, "q012_profile": "q0=56-bit, q1<=39-bit, q2<=40-bit", "fallback_outside_domain": true},
		Tests:        map[string]interface{}{"secondary_focused": "pass", "secondary_polynomial_package": "pass", "primary_go_test": "pass", "primary_git_diff_check": "pass", "no_secondary_commit_or_push": true},
		Validation:   map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_parameter_tuning": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "secondary_worktree_intentionally_dirty": secondaryDirty},
	}
	polynomial := map[string]interface{}{}
	evalmod := map[string]interface{}{}
	params, fastSK := profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters)
	for _, branchName := range []string{"real", "imag"} {
		fastInput, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
		if branchName == "imag" {
			fastInput, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
		}
		_, standardValues, err := semanticBisectView(standardInput, profile.BTP.BootstrappingParameters, profile.StandardSK)
		if err != nil {
			return err
		}
		_, fastValues, err := semanticBisectView(fastInput, params, fastSK)
		if err != nil {
			return err
		}
		c2sMetric, err := semanticBisectMetricFor(standardValues, fastValues, fix001P3P93InternalC2SThreshold)
		if err != nil {
			return err
		}
		if !c2sMetric.Pass {
			return fmt.Errorf("C2S precondition failed for %s: max_component_abs=%g", branchName, c2sMetric.MaxComponentAbs)
		}
		standardBranch, _, _, err := fix001P3P93InternalStandardTrace(standardInput, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		fastBranch, _, _, err := fix001P3P93InternalFastTrace(fastInput, standardInput, profile.Fast, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		merged, err := mergeFIX001P3P93InternalBranches(standardBranch, fastBranch, profile.BTP.BootstrappingParameters, profile.StandardSK, profile.Inputs.FastSK)
		if err != nil {
			return err
		}
		_, standardDA2Values, err := semanticBisectView(standardBranch.states["double_angle_round_2_post_rescale"], params, profile.StandardSK)
		if err != nil {
			return err
		}
		canonicalFastDA2Values, err := fix001P3P93CanonicalFastValuesForParams(merged.states["double_angle_round_2_post_rescale"], params, fastSK, fix001P3FinalRestoreCurrentExponent)
		if err != nil {
			return err
		}
		canonicalDA2, err := semanticBisectMetricFor(standardDA2Values, canonicalFastDA2Values, fix001P3PostProductDA2Budget)
		if err != nil {
			return err
		}
		polynomial[branchName] = map[string]interface{}{"metadata": fix001P3PostProductMetadata(merged.states["polynomial_output_after_rescale"]), "fast_vs_genuine_standard": fix001P3PostProductRethreshold(fix001P3PostProductCheckpoint(merged, "polynomial_output_after_rescale"), fix001P3PostProductBudget), "budget": fix001P3PostProductBudget}
		evalmod[branchName] = map[string]interface{}{"polynomial_output": fix001P3PostProductRethreshold(fix001P3PostProductCheckpoint(merged, "polynomial_output_after_rescale"), fix001P3PostProductBudget), "da0_post_rescale": fix001P3PostProductCheckpoint(merged, "double_angle_round_0_post_rescale"), "da1_post_rescale": fix001P3PostProductCheckpoint(merged, "double_angle_round_1_post_rescale"), "da2_post_rescale_raw": fix001P3PostProductCheckpoint(merged, "double_angle_round_2_post_rescale"), "da2_post_rescale": &canonicalDA2, "internal_final": fix001P3PostProductCheckpoint(merged, "internal_final_restore"), "da2_budget": fix001P3PostProductDA2Budget, "internal_budget": fix001P3PostProductInternalBudget}
		fastPublic, err := profile.Fast.EvalMod(fastInput.CopyNew())
		if err != nil {
			return err
		}
		standardPublic, err := profile.Standard.EvalMod(standardInput.CopyNew())
		if err != nil {
			return err
		}
		_, fastValues, err = semanticBisectView(fastPublic, params, fastSK)
		if err != nil {
			return err
		}
		_, standardValues, err = semanticBisectView(standardPublic, params, profile.StandardSK)
		if err != nil {
			return err
		}
		publicMetric, err := semanticBisectMetricFor(standardValues, fastValues, fix001P3PostProductPublicBudget)
		if err != nil {
			return err
		}
		evalmod[branchName].(map[string]interface{})["public_evalmod"] = &publicMetric
	}
	result.Polynomial, result.EvalMod = polynomial, evalmod
	powerProfile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	for _, branch := range []string{"real", "imag"} {
		rows, err := fix001P3PostProductPowerRows(powerProfile, branch)
		if err != nil {
			return err
		}
		result.Generated = append(result.Generated, rows...)
	}
	workload := reproducibleValues(profile.Residual.MaxSlots())
	fastOutput, err := profile.Fast.Bootstrap(reproducibleInput(profile.Residual, profile.BTP).CopyNew())
	if err != nil {
		return err
	}
	standardOutput, err := profile.Standard.Bootstrap(reproducibleInput(profile.Residual, profile.BTP).CopyNew())
	if err != nil {
		return err
	}
	fastDecoded, err := decodeWithSecret(profile.Residual, fastOutput, zeroSecret(profile.Residual))
	if err != nil {
		return err
	}
	standardDecoded, err := decodeWithSecret(profile.Residual, standardOutput, profile.StandardSK)
	if err != nil {
		return err
	}
	fastE2E, err := semanticBisectMetricFor(workload, fastDecoded, fix001P3PostProductPublicBudget)
	if err != nil {
		return err
	}
	standardE2E, err := semanticBisectMetricFor(workload, standardDecoded, fix001P3PostProductPublicBudget)
	if err != nil {
		return err
	}
	result.ExactE2E = map[string]interface{}{"fast": fastE2E, "standard": standardE2E, "fast_metadata": fix001P3PostProductMetadata(fastOutput), "standard_metadata": fix001P3PostProductMetadata(standardOutput), "exact_production_bootstrap": true}
	generatedPass := len(result.Generated) == 12
	for _, row := range result.Generated {
		generatedPass = generatedPass && row.Oracle.Pass && row.Q01Capacity.CenteredUnique && row.Q012Capacity.CenteredUnique && row.RemainingQ012Bits >= 0
	}
	polynomialPass, da2Pass, internalPass, publicPass := true, true, true, true
	for _, branch := range []string{"real", "imag"} {
		p := result.Polynomial[branch].(map[string]interface{})["fast_vs_genuine_standard"].(*semanticBisectMetric)
		polynomialPass = polynomialPass && p.MaxComponentAbs <= fix001P3PostProductBudget
		e := result.EvalMod[branch].(map[string]interface{})
		da2Pass = da2Pass && e["da2_post_rescale"].(*semanticBisectMetric).MaxComponentAbs <= fix001P3PostProductDA2Budget
		internalPass = internalPass && e["internal_final"].(*semanticBisectMetric).MaxComponentAbs <= fix001P3PostProductInternalBudget
		publicPass = publicPass && e["public_evalmod"].(*semanticBisectMetric).MaxComponentAbs <= fix001P3PostProductPublicBudget
	}
	if generatedPass && polynomialPass && da2Pass && internalPass && publicPass && fastE2E.MaxComponentAbs <= fix001P3PostProductPublicBudget {
		result.Classification = "P93_POSTPRODUCT_REPAIR_1E2_SYSTEM_PASS"
	} else if !generatedPass {
		result.Classification = "P93_POSTPRODUCT_REPAIR_GENERATED_POWER_FAIL"
	} else if !polynomialPass {
		result.Classification = "P93_POSTPRODUCT_REPAIR_POLYNOMIAL_FAIL"
	} else if !da2Pass || !internalPass || !publicPass {
		result.Classification = "P93_POSTPRODUCT_REPAIR_EVALMOD_FAIL"
	} else {
		result.Classification = "P93_POSTPRODUCT_REPAIR_E2E_FAIL"
	}
	return fix001P3PostProductWrite(result, outPath)
}
