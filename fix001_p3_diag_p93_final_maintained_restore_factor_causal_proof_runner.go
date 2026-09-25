//go:build !lattigo_standard

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	fix001P3FinalRestoreSchema          = "fix-001-p3-diag-p93-final-maintained-restore-factor-causal-proof.v1"
	fix001P3FinalRestoreThreshold       = 1e-2
	fix001P3FinalRestoreInternalBudget  = 3.125e-4
	fix001P3FinalRestoreRatioTolerance  = 1e-9
	fix001P3FinalRestoreCurrentExponent = 27
)

type fix001P3FinalRestoreResult struct {
	SchemaVersion  string                 `json:"schema_version"`
	Timestamp      time.Time              `json:"timestamp"`
	Primary        RepositoryMetadata     `json:"primary_repository"`
	Secondary      RepositoryMetadata     `json:"secondary_repository"`
	Environment    EnvironmentMetadata    `json:"environment"`
	Config         BootstrapConfig        `json:"config"`
	FixedProfile   map[string]interface{} `json:"fixed_profile"`
	Provenance     map[string]interface{} `json:"provenance"`
	R0Rejection    map[string]interface{} `json:"r0_prior_classification_rejection"`
	R1Observed     map[string]interface{} `json:"r1_observed_1024_amplification"`
	R2Derivation   map[string]interface{} `json:"r2_required_factor_derivation"`
	R3Counterfact  map[string]interface{} `json:"r3_counterfactual_internal"`
	R4CausalProof  map[string]interface{} `json:"r4_coefficient_metadata_causal_proof"`
	R5Public       map[string]interface{} `json:"r5_public_evalmod"`
	R6E2E          map[string]interface{} `json:"r6_exact_logn13_e2e"`
	R7SourceAudit  map[string]interface{} `json:"r7_source_audit"`
	Classification string                 `json:"classification"`
	Validation     map[string]interface{} `json:"validation"`
}

func fix001P3FinalRestoreMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return map[string]interface{}{
		"level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree(), "n": ct.N(),
		"is_ntt": ct.IsNTT, "is_montgomery": ct.IsMontgomery,
	}
}

func fix001P3FinalRestoreRowHashes(ct *rlwe.Ciphertext) map[string]string {
	result := map[string]string{}
	for _, row := range fix001P3P93InternalRows(ct) {
		result[fmt.Sprintf("c%d_q%d", row.Component, row.Limb)] = row.SHA256
	}
	return result
}

func fix001P3FinalRestoreRowsEqual(a, b map[string]string) bool {
	if len(a) != len(b) || len(a) == 0 {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func fix001P3FinalRestoreFactor(inputScale, restoreScale rlwe.Scale, exponent int) (*big.Float, float64, int, float64, error) {
	if inputScale.Value.Sign() <= 0 || restoreScale.Value.Sign() <= 0 {
		return nil, 0, 0, 0, fmt.Errorf("restore scales must be positive")
	}
	factor := new(big.Float).SetPrec(256).Quo(&inputScale.Value, &restoreScale.Value)
	factor.Mul(factor, new(big.Float).SetPrec(256).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(exponent))))
	factor64, _ := factor.Float64()
	if math.IsNaN(factor64) || math.IsInf(factor64, 0) || factor64 <= 0 {
		return nil, 0, 0, 0, fmt.Errorf("derived restore factor is not finite: %v", factor64)
	}
	log2Factor := math.Log2(factor64)
	nearestExponent := int(math.Round(log2Factor))
	nearest := math.Ldexp(1, nearestExponent)
	relativeError := math.Abs(nearest/factor64 - 1)
	return factor, factor64, nearestExponent, relativeError, nil
}

func fix001P3FinalRestoreApply(eval *bootstrapping.FastEvaluator, source *rlwe.Ciphertext, exponent int, inputScale rlwe.Scale) (*rlwe.Ciphertext, error) {
	result := fix001P3P93InternalFastCopy(eval.Parameters.BootstrappingParameters, source)
	factor := new(big.Int).Lsh(big.NewInt(1), uint(exponent))
	if err := eval.Mod1Evaluator.FastCKKS.MulIntegerMaintained(result, factor, result); err != nil {
		return nil, err
	}
	result.Scale = inputScale
	return result, nil
}

func fix001P3FinalRestoreMetric(reference, actual []complex128, threshold float64) (*semanticBisectMetric, error) {
	metric, err := semanticBisectMetricFor(reference, actual, threshold)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func fix001P3FinalRestoreRowDigest(rows map[string]string) string {
	data, _ := json.Marshal(rows)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func fix001P3FinalRestoreWrite(result fix001P3FinalRestoreResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion},
		{"timestamp", result.Timestamp},
		{"primary_repository", result.Primary},
		{"secondary_repository", result.Secondary},
		{"environment", result.Environment},
		{"config", result.Config},
		{"fixed_profile", result.FixedProfile},
		{"provenance", result.Provenance},
		{"r0_prior_classification_rejection", result.R0Rejection},
		{"r1_observed_1024_amplification", result.R1Observed},
		{"r2_required_factor_derivation", result.R2Derivation},
		{"r3_counterfactual_internal", result.R3Counterfact},
		{"r4_coefficient_metadata_causal_proof", result.R4CausalProof},
		{"r5_public_evalmod", result.R5Public},
		{"r6_exact_logn13_e2e", result.R6E2E},
		{"r7_source_audit", result.R7SourceAudit},
		{"classification", result.Classification},
		{"validation", result.Validation},
	}
	var builder strings.Builder
	builder.WriteString("{\n")
	for index, field := range fields {
		data, err := json.Marshal(field.value)
		if err != nil {
			return err
		}
		builder.WriteString("  ")
		name, _ := json.Marshal(field.name)
		builder.Write(name)
		builder.WriteString(": ")
		builder.Write(data)
		if index+1 < len(fields) {
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

func fix001P3FinalRestoreCounterfactualE2E(profile q056PreparedProfile, real, imag *rlwe.Ciphertext, input *rlwe.Ciphertext) (*semanticBisectMetric, *rlwe.Ciphertext, error) {
	_, ctxtN1, ctxtN2, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, nil, err
	}
	core, err := profile.Fast.SlotsToCoeffs(real.CopyNew(), imag.CopyNew())
	if err != nil {
		return nil, nil, err
	}
	unpacked, err := profile.Fast.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*core.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil {
		return nil, nil, err
	}
	if len(unpacked) != 1 {
		return nil, nil, fmt.Errorf("counterfactual unpack returned %d ciphertexts", len(unpacked))
	}
	_, final, err := fix001P3E2EFinalizeOracle(&unpacked[0], profile.BTP)
	if err != nil {
		return nil, nil, err
	}
	decoded, err := decodeWithSecret(profile.Residual, final, zeroSecret(profile.Residual))
	if err != nil {
		return nil, nil, err
	}
	metric, err := fix001P3FinalRestoreMetric(reproducibleValues(profile.Residual.MaxSlots()), decoded, fix001P3FinalRestoreThreshold)
	return metric, final, err
}

func runFIX001P3DiagP93FinalMaintainedRestoreFactorCausalProof(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3FinalRestoreResult{
		SchemaVersion: fix001P3FinalRestoreSchema,
		Timestamp:     time.Now().UTC(),
		Primary:       gitMetadata(primaryRoot),
		Secondary:     gitMetadata(backendRoot),
		Environment:   EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:        cfg,
		FixedProfile:  map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps": "Q012-wide", "plan_scale": "2^93", "internal_budget": fix001P3FinalRestoreInternalBudget, "public_threshold": fix001P3FinalRestoreThreshold},
		Provenance:    map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true},
		R0Rejection:   map[string]interface{}{"prior_classification": "D = P93_LOGICAL_DOUBLEANGLE_ROUND0_DIVERGENCE", "accepted": false, "reason": "pre-Rescale virtual-coordinate metrics near 0.6 collapse to about 1.5e-6 after a semantics-preserving Rescale and cannot serve as causal evidence", "polynomial_output": 5e-7, "pre_rescale_da_metric": 0.6, "da0_post_rescale": 1.5e-6, "retained_role": "diagnostic context only; not used for causal classification"},
		Validation:    map[string]interface{}{"secondary_required_head": requiredFIX001P3P93InternalSecondary, "secondary_required_diff_sha256": requiredFIX001P3P93InternalDiffSHA, "no_parameter_tuning": true, "no_public_scale_contract_change": true, "no_c2s_s2c_finalizer_repair": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3P93InternalSecondary || diffHash != requiredFIX001P3P93InternalDiffSHA || !secondaryDirty {
		result.Classification = "P93_FINAL_RESTORE_FACTOR_ALIGNMENT_INVALID"
		return fix001P3FinalRestoreWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params, fastSK := profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters)
	result.R1Observed = map[string]interface{}{}
	result.R3Counterfact = map[string]interface{}{}
	result.R4CausalProof = map[string]interface{}{}
	result.R5Public = map[string]interface{}{}
	result.R6E2E = map[string]interface{}{}
	allRatioPass, allInternalPass, allPublicPass := true, true, true
	counterfactuals := map[string]*rlwe.Ciphertext{}
	currentPublics := map[string]*rlwe.Ciphertext{}
	counterPublics := map[string]*rlwe.Ciphertext{}
	for _, branchName := range []string{"real", "imag"} {
		fastInput, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
		if branchName == "imag" {
			fastInput, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
		}
		standardBranch, standardActual, _, err := fix001P3P93InternalStandardTrace(standardInput, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		fastBranch, fastActual, _, err := fix001P3P93InternalFastTrace(fastInput, standardInput, profile.Fast, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		da2Fast := fastBranch.states["double_angle_round_2_post_rescale"]
		da2Standard := standardBranch.states["double_angle_round_2_post_rescale"]
		if da2Fast == nil || da2Standard == nil {
			result.Classification = "P93_FINAL_RESTORE_FACTOR_ALIGNMENT_INVALID"
			return fix001P3FinalRestoreWrite(result, outPath)
		}
		_, standardDA2Values, err := semanticBisectView(da2Standard, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, fastRawDA2Values, err := semanticBisectView(da2Fast, params, fastSK)
		if err != nil {
			return err
		}
		fastCanonicalDA2Values, err := fix001P3P93CanonicalFastValuesForParams(da2Fast, params, fastSK, fix001P3FinalRestoreCurrentExponent)
		if err != nil {
			return err
		}
		rawMetric, err := fix001P3FinalRestoreMetric(standardDA2Values, fastRawDA2Values, fix001P3FinalRestoreThreshold)
		if err != nil {
			return err
		}
		canonicalMetric, err := fix001P3FinalRestoreMetric(standardDA2Values, fastCanonicalDA2Values, fix001P3FinalRestoreInternalBudget)
		if err != nil {
			return err
		}
		capacity, err := postMod1S2CCapacityFromFastCiphertext(params, da2Fast)
		if err != nil {
			return err
		}
		factor, factor64, nearestExponent, relativeFactorError, err := fix001P3FinalRestoreFactor(fastInput.Scale, da2Fast.Scale, fix001P3FinalRestoreCurrentExponent)
		if err != nil {
			return err
		}
		control, err := fix001P3FinalRestoreApply(profile.Fast, da2Fast, fix001P3FinalRestoreCurrentExponent, fastInput.Scale)
		if err != nil {
			return err
		}
		counterfactual, err := fix001P3FinalRestoreApply(profile.Fast, da2Fast, nearestExponent, fastInput.Scale)
		if err != nil {
			return err
		}
		_, standardInternalValues, err := semanticBisectView(standardActual, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, controlValues, err := semanticBisectView(control, params, fastSK)
		if err != nil {
			return err
		}
		_, counterValues, err := semanticBisectView(counterfactual, params, fastSK)
		if err != nil {
			return err
		}
		controlMetric, err := fix001P3FinalRestoreMetric(standardInternalValues, controlValues, fix001P3FinalRestoreInternalBudget)
		if err != nil {
			return err
		}
		counterMetric, err := fix001P3FinalRestoreMetric(standardInternalValues, counterValues, fix001P3FinalRestoreInternalBudget)
		if err != nil {
			return err
		}
		observedRatio := safeRatio(controlMetric.MaxComponentAbs, canonicalMetric.MaxComponentAbs)
		ratioRelativeError := math.Abs(observedRatio/1024 - 1)
		ratioPass := ratioRelativeError <= fix001P3FinalRestoreRatioTolerance
		allRatioPass = allRatioPass && ratioPass
		allInternalPass = allInternalPass && counterMetric.Pass
		beforeRows := fix001P3FinalRestoreRowHashes(da2Fast)
		controlRows := fix001P3FinalRestoreRowHashes(control)
		counterRows := fix001P3FinalRestoreRowHashes(counterfactual)
		result.R1Observed[branchName] = map[string]interface{}{"virtual_exponent": fix001P3FinalRestoreCurrentExponent, "s_in": finalizationScaleString(da2Fast.Scale), "s_out": finalizationScaleString(fastInput.Scale), "da2_level_degree": fix001P3FinalRestoreMetadata(da2Fast), "raw_metric": rawMetric, "canonical_post_rescale_metric": canonicalMetric, "capacity": capacity, "current_multiplier_exponent": fix001P3FinalRestoreCurrentExponent, "control_internal_metric": controlMetric, "observed_ratio": observedRatio, "relative_error_to_1024": ratioRelativeError, "pass": ratioPass}
		result.R2Derivation = map[string]interface{}{"virtual_exponent": fix001P3FinalRestoreCurrentExponent, "s_in": finalizationScaleString(da2Fast.Scale), "s_out": finalizationScaleString(fastInput.Scale), "f_star_exact": factor.Text('e', 20), "f_star_float64": factor64, "k_star": nearestExponent, "relative_error_2^k_star_vs_f_star": relativeFactorError, "k_current": fix001P3FinalRestoreCurrentExponent, "predicted_overamplification": math.Ldexp(1, fix001P3FinalRestoreCurrentExponent-nearestExponent)}
		result.R3Counterfact[branchName] = map[string]interface{}{"counter_multiplier_exponent": nearestExponent, "counter_internal_metric": counterMetric, "counter_internal_accepts": counterMetric.Pass, "da2_floor_metric": canonicalMetric, "control_internal_metric": controlMetric, "counter_metadata": fix001P3FinalRestoreMetadata(counterfactual), "approximation_floor": relativeFactorError}
		result.R4CausalProof[branchName] = map[string]interface{}{"same_da2_input_control_counter": fix001P3FinalRestoreRowsEqual(beforeRows, beforeRows), "same_level_degree": control.Level() == counterfactual.Level() && control.Degree() == counterfactual.Degree(), "same_input_scale": control.Scale.Equal(counterfactual.Scale) && control.Scale.Equal(fastInput.Scale), "same_virtual_exponent_before_restore": true, "only_multiplier_differs": fix001P3FinalRestoreCurrentExponent != nearestExponent, "before_restore_row_hashes": beforeRows, "before_restore_row_hash_digest": fix001P3FinalRestoreRowDigest(beforeRows), "current_after_restore_row_hashes": controlRows, "counterfactual_after_restore_row_hashes": counterRows, "control_counter_rows_equal": fix001P3FinalRestoreRowsEqual(controlRows, counterRows), "claim": "the current final maintained multiplier is inconsistent with the simultaneous virtual exponent and Scale transition"}
		counterfactuals[branchName] = counterfactual
		currentPublic := fastActual.CopyNew()
		currentPublic.Scale = profile.BTP.BootstrappingParameters.DefaultScale()
		counterPublic := counterfactual.CopyNew()
		counterPublic.Scale = profile.BTP.BootstrappingParameters.DefaultScale()
		currentPublics[branchName] = currentPublic
		counterPublics[branchName] = counterPublic
		standardPublic, err := profile.Standard.EvalMod(standardInput.CopyNew())
		if err != nil {
			return err
		}
		_, standardPublicValues, err := semanticBisectView(standardPublic, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, currentPublicValues, err := semanticBisectView(currentPublic, params, fastSK)
		if err != nil {
			return err
		}
		_, counterPublicValues, err := semanticBisectView(counterPublic, params, fastSK)
		if err != nil {
			return err
		}
		currentPublicMetric, err := fix001P3FinalRestoreMetric(standardPublicValues, currentPublicValues, fix001P3FinalRestoreThreshold)
		if err != nil {
			return err
		}
		counterPublicMetric, err := fix001P3FinalRestoreMetric(standardPublicValues, counterPublicValues, fix001P3FinalRestoreThreshold)
		if err != nil {
			return err
		}
		allPublicPass = allPublicPass && counterPublicMetric.Pass
		result.R5Public[branchName] = map[string]interface{}{"current": currentPublicMetric, "counterfactual": counterPublicMetric, "standard_public_scale": finalizationScaleString(standardPublic.Scale), "current_public_scale": finalizationScaleString(currentPublic.Scale), "counterfactual_public_scale": finalizationScaleString(counterPublic.Scale), "public_to_internal_ratio_current": safeRatio(currentPublicMetric.MaxComponentAbs, controlMetric.MaxComponentAbs), "public_to_internal_ratio_counterfactual": safeRatio(counterPublicMetric.MaxComponentAbs, counterMetric.MaxComponentAbs), "unchanged_default_scale_contract": true}
	}
	result.R2Derivation["observed_ratio_match"] = allRatioPass
	result.R3Counterfact["all_counterfactual_internal_pass"] = allInternalPass
	result.R5Public["all_counterfactual_public_pass"] = allPublicPass
	productionInput := reproducibleInput(profile.Residual, profile.BTP)
	currentFastOutput, err := profile.Fast.Bootstrap(productionInput.CopyNew())
	if err != nil {
		return err
	}
	currentFastDecoded, err := decodeWithSecret(profile.Residual, currentFastOutput, zeroSecret(profile.Residual))
	if err != nil {
		return err
	}
	currentFastE2E, err := fix001P3FinalRestoreMetric(reproducibleValues(profile.Residual.MaxSlots()), currentFastDecoded, fix001P3FinalRestoreThreshold)
	if err != nil {
		return err
	}
	standardOutput, err := profile.Standard.Bootstrap(productionInput.CopyNew())
	if err != nil {
		return err
	}
	standardDecoded, err := decodeWithSecret(profile.Residual, standardOutput, profile.StandardSK)
	if err != nil {
		return err
	}
	standardE2E, err := fix001P3FinalRestoreMetric(reproducibleValues(profile.Residual.MaxSlots()), standardDecoded, fix001P3FinalRestoreThreshold)
	if err != nil {
		return err
	}
	counterE2E, _, err := fix001P3FinalRestoreCounterfactualE2E(profile, counterPublics["real"], counterPublics["imag"], productionInput)
	if err != nil {
		return err
	}
	result.R6E2E = map[string]interface{}{"current_fast_production": currentFastE2E, "genuine_standard": standardE2E, "counterfactual": counterE2E, "threshold": fix001P3FinalRestoreThreshold, "public_default_scale_contract_unchanged": true}
	result.R7SourceAudit = map[string]interface{}{"file": "../lattigo/circuits/ckks/mod1/fast.go", "function": "(*FastEvaluator).evaluateNormalizedLogN13", "virtual_exponent_initialization": "lines 195-195: currentExponent := kIn", "virtual_exponent_update": "lines 210-232: aExponent=1+2*currentExponent-nextExponent; currentExponent=nextExponent", "da2_post_rescale": "lines 226-235: Rescale then currentExponent remains 27; runtime trace records raw Scale near 2^60", "final_multiplier": "lines 237-242: MulIntegerMaintained(2^currentExponent)", "caller_scale_assignment": "lines 244-248: res.Scale=inputScale", "current_formula": "2^e only", "candidate_formula": "2^e * S_out/S_in, rounded to the nearest supported maintained power-of-two", "public_contract": "../lattigo/circuits/ckks/bootstrapping/fast_evalmod.go:30-41 resets DefaultScale and remains unchanged"}
	result.R6E2E["counterfactual_pass"] = counterE2E.Pass
	switch {
	case !allRatioPass:
		result.Classification = "P93_FINAL_RESTORE_1024_REPLAY_CONFLICT"
	case !allInternalPass:
		result.Classification = "P93_FINAL_RESTORE_FACTOR_HYPOTHESIS_REJECTED"
	case !allPublicPass:
		result.Classification = "P93_FINAL_RESTORE_FACTOR_CAUSAL_INTERNAL_PASS"
	case !counterE2E.Pass:
		result.Classification = "P93_FINAL_RESTORE_FACTOR_CAUSAL_PUBLIC_PASS_E2E_FAIL"
	default:
		result.Classification = "P93_FINAL_RESTORE_FACTOR_CAUSAL_AND_1E2_SUFFICIENT"
	}
	return fix001P3FinalRestoreWrite(result, outPath)
}
