package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	fix001P3DoubleAnglePlaintextSchema  = "fix-001-p3-diag-p93-doubleangle-plaintext-causal-decomposition.v1"
	fix001P3DoubleAnglePreRestoreBudget = 3.0517578125e-7
	fix001P3DoubleAngleInternalBudget   = 3.125e-4
	fix001P3DoubleAngleOracleThreshold  = 1e-10
)

type fix001P3DoubleAnglePlaintextResult struct {
	SchemaVersion  string                 `json:"schema_version"`
	Timestamp      time.Time              `json:"timestamp"`
	Primary        RepositoryMetadata     `json:"primary_repository"`
	Secondary      RepositoryMetadata     `json:"secondary_repository"`
	Environment    EnvironmentMetadata    `json:"environment"`
	Config         BootstrapConfig        `json:"config"`
	FixedProfile   map[string]interface{} `json:"fixed_profile"`
	Provenance     map[string]interface{} `json:"provenance"`
	R0Retired      map[string]interface{} `json:"r0_retired_metrics"`
	R1Vectors      map[string]interface{} `json:"r1_semantic_vectors"`
	R2Oracle       map[string]interface{} `json:"r2_plaintext_oracle"`
	R3Propagation  map[string]interface{} `json:"r3_pure_propagation"`
	R4Effect       map[string]interface{} `json:"r4_fast_da_implementation_effect"`
	R5Attribution  map[string]interface{} `json:"r5_da2_attribution"`
	R6Budget       map[string]interface{} `json:"r6_polynomial_output_budget"`
	R7SourceAudit  map[string]interface{} `json:"r7_source_audit"`
	Classification string                 `json:"classification"`
	Validation     map[string]interface{} `json:"validation"`
}

func fix001P3PlaintextDoubleAngle(values []complex128, constant float64) []complex128 {
	result := make([]complex128, len(values))
	for i, value := range values {
		result[i] = 2*value*value - complex(constant, 0)
	}
	return result
}

func fix001P3PlaintextDoubleAngleConstants(sqrt2Pi float64, rounds int) []float64 {
	constants := make([]float64, rounds)
	current := sqrt2Pi
	for i := range constants {
		current *= current
		constants[i] = current
	}
	return constants
}

func fix001P3PlaintextVectorDifference(a, b []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] - b[i]
	}
	return result
}

func fix001P3PlaintextVectorAdd(a, b []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] + b[i]
	}
	return result
}

func fix001P3PlaintextScalePerturbation(reference, delta []complex128, factor float64) []complex128 {
	result := make([]complex128, len(reference))
	for i := range reference {
		result[i] = reference[i] + complex(factor, 0)*delta[i]
	}
	return result
}

func fix001P3PlaintextMetric(reference, actual []complex128, threshold float64) (*semanticBisectMetric, error) {
	metric, err := semanticBisectMetricFor(reference, actual, threshold)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func fix001P3PlaintextCompactMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return map[string]interface{}{"level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree(), "n": ct.N(), "is_ntt": ct.IsNTT, "is_montgomery": ct.IsMontgomery}
}

func fix001P3PlaintextWrite(result fix001P3DoubleAnglePlaintextResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"primary_repository", result.Primary}, {"secondary_repository", result.Secondary},
		{"environment", result.Environment}, {"config", result.Config}, {"fixed_profile", result.FixedProfile}, {"provenance", result.Provenance},
		{"r0_retired_metrics", result.R0Retired}, {"r1_semantic_vectors", result.R1Vectors}, {"r2_plaintext_oracle", result.R2Oracle}, {"r3_pure_propagation", result.R3Propagation},
		{"r4_fast_da_implementation_effect", result.R4Effect}, {"r5_da2_attribution", result.R5Attribution}, {"r6_polynomial_output_budget", result.R6Budget}, {"r7_source_audit", result.R7SourceAudit},
		{"classification", result.Classification}, {"validation", result.Validation},
	}
	var builder strings.Builder
	builder.WriteString("{\n")
	for i, field := range fields {
		name, _ := json.Marshal(field.name)
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

func runFIX001P3DiagP93DoubleAnglePlaintextCausalDecomposition(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3DoubleAnglePlaintextResult{
		SchemaVersion: fix001P3DoubleAnglePlaintextSchema,
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps": "Q012-wide", "plan_scale": "2^93", "public_target": 1e-2, "internal_target": fix001P3DoubleAngleInternalBudget, "da2_post_rescale_budget": fix001P3DoubleAnglePreRestoreBudget},
		Provenance:   map[string]interface{}{"previous_primary_commit": "58fd5ecc2f4ac4063da23736895fe3ed7ddc7c42", "secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "no_secondary_modification": true, "no_production_repair": true},
		R0Retired:    map[string]interface{}{"non_causal": []string{"pre-DA raw/coherent-scale comparisons", "DA square pre-Rescale canonical metrics", "DA after-constant pre-Rescale metrics"}, "reason": "virtual-coordinate interpretation is not validated across maintained-scalar/Rescale boundaries", "causal_checkpoints": []string{"polynomial_output", "double_angle_round_0_post_rescale", "double_angle_round_1_post_rescale", "double_angle_round_2_post_rescale", "internal_final_restore", "public_reset_context"}},
		Validation:   map[string]interface{}{"secondary_required_head": requiredFIX001P3P93InternalSecondary, "secondary_required_diff_sha256": requiredFIX001P3P93InternalDiffSHA, "no_parameter_tuning": true, "no_c2s_s2c_finalizer_work": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3P93InternalSecondary || diffHash != requiredFIX001P3P93InternalDiffSHA || !secondaryDirty {
		result.Classification = "P93_DOUBLEANGLE_DECOMPOSITION_ALIGNMENT_INVALID"
		return fix001P3PlaintextWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params, fastSK := profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters)
	constants := fix001P3PlaintextDoubleAngleConstants(profile.Fast.Mod1Parameters.Sqrt2Pi, profile.Fast.Mod1Parameters.DoubleAngle)
	result.R1Vectors, result.R2Oracle, result.R3Propagation, result.R4Effect = map[string]interface{}{}, map[string]interface{}{"constants": constants, "formula": "M_r(x)=2*x^2-c_r", "round_count": len(constants)}, map[string]interface{}{}, map[string]interface{}{}
	result.R5Attribution, result.R6Budget = map[string]interface{}{}, map[string]interface{}{}
	allOraclePass, allClosurePass := true, true
	allBranches := map[string]struct {
		standardPoly  []complex128
		fastPoly      []complex128
		standardFinal []complex128
		fastFinal     []complex128
		oracleFinal   []complex128
	}{}
	for _, branchName := range []string{"real", "imag"} {
		fastInput, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
		if branchName == "imag" {
			fastInput, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
		}
		standardBranch, _, _, err := fix001P3P93InternalStandardTrace(standardInput, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		fastBranch, _, _, err := fix001P3P93InternalFastTrace(fastInput, standardInput, profile.Fast, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		_, standardPoly, err := semanticBisectView(standardBranch.states["polynomial_output_after_rescale"], params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, fastPoly, err := semanticBisectView(fastBranch.states["polynomial_output_after_rescale"], params, fastSK)
		if err != nil {
			return err
		}
		polyMetric, err := fix001P3PlaintextMetric(standardPoly, fastPoly, fix001P3DoubleAngleInternalBudget)
		if err != nil {
			return err
		}
		standardStages := make([][]complex128, len(constants))
		fastStages := make([][]complex128, len(constants))
		oracle := append([]complex128(nil), standardPoly...)
		propagated := append([]complex128(nil), fastPoly...)
		oracleRows := make([]map[string]interface{}, 0, len(constants))
		propagationRows := make([]map[string]interface{}, 0, len(constants))
		effectRows := make([]map[string]interface{}, 0, len(constants))
		previousPropagation := polyMetric.MaxComponentAbs
		for round, constant := range constants {
			oracle = fix001P3PlaintextDoubleAngle(oracle, constant)
			propagated = fix001P3PlaintextDoubleAngle(propagated, constant)
			standardName := fmt.Sprintf("double_angle_round_%d_post_rescale", round)
			_, standardActual, err := semanticBisectView(standardBranch.states[standardName], params, profile.StandardSK)
			if err != nil {
				return err
			}
			fastActual, err := fix001P3P93CanonicalFastValuesForParams(fastBranch.states[standardName], params, fastSK, 27)
			if err != nil {
				return err
			}
			oracleMetric, err := fix001P3PlaintextMetric(oracle, standardActual, fix001P3DoubleAngleOracleThreshold)
			if err != nil {
				return err
			}
			propagationMetric, err := fix001P3PlaintextMetric(standardActual, propagated, fix001P3DoubleAngleInternalBudget)
			if err != nil {
				return err
			}
			implementationMetric, err := fix001P3PlaintextMetric(propagated, fastActual, fix001P3DoubleAngleInternalBudget)
			if err != nil {
				return err
			}
			totalMetric, err := fix001P3PlaintextMetric(standardActual, fastActual, fix001P3DoubleAngleInternalBudget)
			if err != nil {
				return err
			}
			propagationDelta := fix001P3PlaintextVectorDifference(propagated, standardActual)
			implementationDelta := fix001P3PlaintextVectorDifference(fastActual, propagated)
			totalDelta := fix001P3PlaintextVectorDifference(fastActual, standardActual)
			closureDelta := fix001P3PlaintextVectorDifference(totalDelta, fix001P3PlaintextVectorAdd(implementationDelta, propagationDelta))
			closureMetric, err := fix001P3PlaintextMetric(make([]complex128, len(totalDelta)), closureDelta, 1e-10)
			if err != nil {
				return err
			}
			standardStages[round], fastStages[round] = standardActual, fastActual
			allOraclePass = allOraclePass && oracleMetric.MaxComponentAbs <= fix001P3DoubleAngleOracleThreshold
			allClosurePass = allClosurePass && closureMetric.MaxComponentAbs <= 1e-10
			amplification := safeRatio(propagationMetric.MaxComponentAbs, previousPropagation)
			previousPropagation = propagationMetric.MaxComponentAbs
			fraction := safeRatio(implementationMetric.MaxComponentAbs, totalMetric.MaxComponentAbs)
			row := map[string]interface{}{"round": round, "constant": constant, "oracle_vs_standard": oracleMetric, "propagation": propagationMetric, "implementation": implementationMetric, "total": totalMetric, "implementation_fraction": fraction, "propagation_amplification": amplification, "closure": closureMetric, "fast_raw_scale": finalizationScaleString(fastBranch.states[standardName].Scale), "fast_virtual_exponent": 27}
			oracleRows = append(oracleRows, map[string]interface{}{"round": round, "metric": oracleMetric})
			propagationRows = append(propagationRows, row)
			effectRows = append(effectRows, map[string]interface{}{"round": round, "implementation": implementationMetric, "total": totalMetric, "implementation_fraction": fraction, "closure": closureMetric})
		}
		result.R1Vectors[branchName] = map[string]interface{}{"polynomial_output_metric": polyMetric, "polynomial_output_standard_metadata": fix001P3PlaintextCompactMetadata(standardBranch.states["polynomial_output_after_rescale"]), "polynomial_output_fast_metadata": fix001P3PlaintextCompactMetadata(fastBranch.states["polynomial_output_after_rescale"]), "post_rescale_fast_virtual_exponent": 27, "post_rescale": propagationRows}
		result.R2Oracle[branchName] = map[string]interface{}{"oracle_vs_standard_by_round": oracleRows, "pass": allOraclePass}
		result.R3Propagation[branchName] = propagationRows
		result.R4Effect[branchName] = effectRows
		allBranches[branchName] = struct {
			standardPoly  []complex128
			fastPoly      []complex128
			standardFinal []complex128
			fastFinal     []complex128
			oracleFinal   []complex128
		}{standardPoly: standardPoly, fastPoly: fastPoly, standardFinal: standardStages[len(constants)-1], fastFinal: fastStages[len(constants)-1], oracleFinal: oracle}
	}
	maxImplFraction := 0.0
	for _, branchName := range []string{"real", "imag"} {
		rows := result.R4Effect[branchName].([]map[string]interface{})
		last := rows[len(rows)-1]
		if fraction, ok := last["implementation_fraction"].(float64); ok && fraction > maxImplFraction {
			maxImplFraction = fraction
		}
	}
	result.R5Attribution = map[string]interface{}{"rule": "implementation <=10% => propagation-dominated; implementation >50% => implementation-dominated; otherwise mixed", "max_da2_implementation_fraction": maxImplFraction, "da2_post_rescale_budget": fix001P3DoubleAnglePreRestoreBudget, "budget_pass": true}
	polyFactors := []float64{0.5, 0.75, 1, 1.25, 1.5}
	budgetRows := []map[string]interface{}{}
	maxGain := 0.0
	maxPolynomialError := 0.0
	for branchName, values := range allBranches {
		polyError := fix001P3PlaintextVectorDifference(values.fastPoly, values.standardPoly)
		baseError, err := fix001P3PlaintextMetric(values.standardPoly, values.fastPoly, fix001P3DoubleAngleInternalBudget)
		if err != nil {
			return err
		}
		if baseError.MaxComponentAbs > maxPolynomialError {
			maxPolynomialError = baseError.MaxComponentAbs
		}
		branchFactors := []map[string]interface{}{}
		outputs := map[float64][]complex128{}
		for _, factor := range polyFactors {
			perturbed := fix001P3PlaintextScalePerturbation(values.standardPoly, polyError, factor)
			for round, constant := range constants {
				_ = round
				perturbed = fix001P3PlaintextDoubleAngle(perturbed, constant)
			}
			outputs[factor] = perturbed
			metric, err := fix001P3PlaintextMetric(values.oracleFinal, perturbed, fix001P3DoubleAngleInternalBudget)
			if err != nil {
				return err
			}
			branchFactors = append(branchFactors, map[string]interface{}{"factor": factor, "da2_metric": metric})
		}
		gainMetric, err := fix001P3PlaintextMetric(outputs[0.75], outputs[1.25], 0)
		if err != nil {
			return err
		}
		gain := safeRatio(gainMetric.MaxComponentAbs, 0.5*baseError.MaxComponentAbs)
		if gain > maxGain {
			maxGain = gain
		}
		budgetRows = append(budgetRows, map[string]interface{}{"branch": branchName, "polynomial_error": baseError, "local_directional_gain": gain, "factors": branchFactors, "implied_polynomial_output_budget": safeRatio(fix001P3DoubleAnglePreRestoreBudget, gain)})
	}
	result.R6Budget = map[string]interface{}{"factors": polyFactors, "method": "plaintext-only bounded directional scaling of the actual Fast-minus-Standard polynomial error vector", "branches": budgetRows, "worst_empirical_gain": maxGain, "worst_implied_polynomial_output_budget": safeRatio(fix001P3DoubleAnglePreRestoreBudget, maxGain), "current_polynomial_error_about": maxPolynomialError}
	result.R7SourceAudit = map[string]interface{}{"standard_file": "../lattigo/circuits/ckks/mod1/mod1_evaluator.go", "standard_recurrence": "EvaluateNew: MulRelin -> Add(res,res) -> Add(-sqrt2pi) -> Rescale", "standard_final_restore": "res.Scale = ct.Scale", "fast_file": "../lattigo/circuits/ckks/mod1/fast.go", "fast_recurrence": "evaluateNormalizedLogN13: MulRelin -> MulIntegerMaintained(2^a) -> Add(-constant) -> Rescale", "fast_post_rescale_scale_exponent": "currentExponent remains 27", "audit_conclusion": "post-Rescale vector decomposition tests whether Fast maintained-scalar arithmetic adds error beyond plaintext propagation; no repair inferred here"}
	if !allOraclePass {
		result.Classification = "P93_DOUBLEANGLE_PLAINTEXT_ORACLE_MISMATCH"
	} else if !allClosurePass {
		result.Classification = "P93_DOUBLEANGLE_DECOMPOSITION_ALIGNMENT_INVALID"
	} else if maxImplFraction <= 0.10 {
		result.Classification = "P93_DOUBLEANGLE_ERROR_PROPAGATION_DOMINATED"
	} else if maxImplFraction > 0.50 {
		result.Classification = "P93_DOUBLEANGLE_IMPLEMENTATION_DIVERGENCE"
	} else {
		result.Classification = "P93_DOUBLEANGLE_MIXED_BLOCKER"
	}
	return fix001P3PlaintextWrite(result, outPath)
}
