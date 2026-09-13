package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	requiredFIX001P3K3DownstreamPrimaryBase = "04db11434229ef97d372922cd3ee3f516d37014c"
	requiredFIX001P3K3DownstreamSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	k3DownstreamPublicThreshold             = 1e-2
)

type fix001P3K3DownstreamResult struct {
	SchemaVersion       string                 `json:"schema_version"`
	Timestamp           time.Time              `json:"timestamp"`
	Primary             RepositoryMetadata     `json:"primary_repository"`
	Lattigo             RepositoryMetadata     `json:"lattigo_repository"`
	Environment         EnvironmentMetadata    `json:"environment"`
	Config              BootstrapConfig        `json:"config"`
	Parameters          ExperimentParameters   `json:"effective_parameters"`
	D0Control           map[string]interface{} `json:"d0_k3_control"`
	Normalization       map[string]interface{} `json:"derived_normalization"`
	EvalMod             map[string]interface{} `json:"evalmod"`
	S2CFinalization     map[string]interface{} `json:"s2c_unpack_finalization"`
	FinalMetadata       map[string]interface{} `json:"final_metadata"`
	LocalBudget         map[string]interface{} `json:"local_budget"`
	PublicLikeThreshold float64                `json:"public_like_threshold"`
	PublicLikePass      bool                   `json:"public_like_threshold_pass"`
	SystemSufficient    bool                   `json:"ps_arithmetic_system_sufficient"`
	Classification      string                 `json:"classification"`
	FirstBlocker        string                 `json:"first_remaining_blocker"`
	Validation          map[string]interface{} `json:"validation"`
}

func fix001P3K3DownstreamWrite(result fix001P3K3DownstreamResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3K3DownstreamSchedule(path evalModMatchedPath) []map[string]interface{} {
	schedules := make([]map[string]interface{}, 0, len(path.Rounds))
	for _, round := range path.Rounds {
		schedules = append(schedules, map[string]interface{}{"round": round.Round, "k_in_exponent": round.KInExponent, "k_out_exponent": round.KOutExponent, "multiplier_exponent": round.MultiplierExponent, "normalized_constant": round.NormalizedConstant})
	}
	return schedules
}

func fix001P3K3DownstreamPathSummary(path evalModMatchedPath) map[string]interface{} {
	prefix := make([]map[string]interface{}, 0, len(path.Prefix))
	for _, checkpoint := range path.Prefix {
		prefix = append(prefix, map[string]interface{}{
			"name":               checkpoint.Name,
			"round":              checkpoint.Round,
			"fast_q01_safe":      checkpoint.FastQ01Safe,
			"normalized_safe":    checkpoint.NormalizedSafe,
			"rows_match":         checkpoint.RowsMatch,
			"fast_vs_normalized": checkpoint.FastVsNormalized,
			"normalized_vs_std":  checkpoint.NormalizedVsStd,
		})
	}
	rounds := make([]map[string]interface{}, 0, len(path.Rounds))
	for _, round := range path.Rounds {
		checkpoints := make([]map[string]interface{}, 0, 5)
		for _, checkpoint := range []evalModMatchedCheckpoint{round.Input, round.Square, round.AfterMultiplier, round.AfterConstant, round.PostRescale} {
			if checkpoint.Name == "" {
				continue
			}
			checkpoints = append(checkpoints, map[string]interface{}{
				"name":            checkpoint.Name,
				"round":           checkpoint.Round,
				"fast_q01_safe":   checkpoint.FastQ01Safe,
				"normalized_safe": checkpoint.NormalizedSafe,
				"rows_match":      checkpoint.RowsMatch,
			})
		}
		rounds = append(rounds, map[string]interface{}{
			"round":               round.Round,
			"k_in_exponent":       round.KInExponent,
			"k_out_exponent":      round.KOutExponent,
			"multiplier_exponent": round.MultiplierExponent,
			"checkpoints":         checkpoints,
		})
	}
	return map[string]interface{}{
		"first_failure":                  path.FirstFailure,
		"fast_final_present":             path.FastFinal != nil,
		"normalized_final_present":       path.NormalizedFinal != nil,
		"coherent_final_present":         path.CoherentFinal != nil,
		"standard_final_present":         path.StandardFinal != nil,
		"fast_rows_match":                path.FastRowsMatch,
		"restore_rows_match":             path.RestoreRowsMatch,
		"scale_reset_rows_unchanged":     path.ScaleResetRowsUnchanged,
		"level_degree_unchanged":         path.LevelDegreeUnchanged,
		"poly_meta":                      path.PolyMeta,
		"normalized_polynomial_capacity": path.PolyCapacity,
		"prefix":                         prefix,
		"rounds":                         rounds,
	}
}

func fix001P3K3DownstreamMetadata(ct *rlwe.Ciphertext, slots int) map[string]interface{} {
	if ct == nil {
		return map[string]interface{}{"present": false}
	}
	return map[string]interface{}{"present": true, "level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree(), "n": ct.N(), "slots": slots, "log_dimensions": ct.LogDimensions, "is_ntt": ct.IsNTT, "is_montgomery": ct.IsMontgomery}
}

type fix001P3K3PublicLikeEvidence struct {
	PostS2C        *PSGlobalMetric
	StandardLike   *PSGlobalMetric
	PublicLike     *PSGlobalMetric
	FinalVsStdCore *PSGlobalMetric
	Metadata       map[string]interface{}
	InputUnchanged bool
}

func fix001P3K3RunPublicLike(pathReal, pathImag evalModMatchedPath, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params bootstrapping.Parameters, residualN int, skN2 *rlwe.SecretKey, input *rlwe.Ciphertext) (fix001P3K3PublicLikeEvidence, error) {
	fastCore, err := fastEval.SlotsToCoeffs(pathReal.FastFinal.CopyNew(), pathImag.FastFinal.CopyNew())
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	standardCore, err := standardEval.SlotsToCoeffs(pathReal.StandardFinal.CopyNew(), pathImag.StandardFinal.CopyNew())
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	fastCoreValues, err := psGlobalDecode(params.BootstrappingParameters, fastCore)
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	standardCoreValues, err := decodeWithSecret(params.BootstrappingParameters, standardCore, skN2)
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	postS2C := evalModMatchedMetricFor(standardCoreValues, fastCoreValues, k3DownstreamPublicThreshold)
	message := reproducibleValues(residualN)
	standardLike := evalModMatchedMetricFor(message, standardCoreValues, k3DownstreamPublicThreshold)
	fastBefore := input.CopyNew()
	_, ctxtN1, ctxtN2, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	unpacked, err := fastEval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*fastCore.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil || len(unpacked) != 1 {
		if err == nil {
			err = fmt.Errorf("expected one unpacked ciphertext, got %d", len(unpacked))
		}
		return fix001P3K3PublicLikeEvidence{}, err
	}
	_, final, err := fix001P3E2EFinalizeOracle(&unpacked[0], params)
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	finalValues, err := decodeWithSecret(params.ResidualParameters, final, zeroSecret(params.ResidualParameters))
	if err != nil {
		return fix001P3K3PublicLikeEvidence{}, err
	}
	publicLike := evalModMatchedMetricFor(message, finalValues, k3DownstreamPublicThreshold)
	finalVsStandardCore := evalModMatchedMetricFor(standardCoreValues, finalValues, k3DownstreamPublicThreshold)
	metadata := fix001P3K3DownstreamMetadata(final, params.ResidualParameters.MaxSlots())
	metadata["expected_level"] = params.ResidualParameters.MaxLevel()
	metadata["expected_degree"] = 1
	metadata["expected_n"] = params.ResidualParameters.N()
	metadata["expected_is_ntt"] = true
	metadata["expected_is_montgomery"] = false
	metadata["expected_scale"] = finalizationScaleString(params.ResidualParameters.DefaultScale())
	metadata["pass"] = final.Level() == params.ResidualParameters.MaxLevel() && final.Degree() == 1 && final.N() == params.ResidualParameters.N() && final.IsNTT && !final.IsMontgomery && final.Scale.Equal(params.ResidualParameters.DefaultScale())
	return fix001P3K3PublicLikeEvidence{PostS2C: postS2C, StandardLike: standardLike, PublicLike: publicLike, FinalVsStdCore: finalVsStandardCore, Metadata: metadata, InputUnchanged: fix001P3E2EInputUnchanged(fastBefore, input)}, nil
}

func runFIX001P3DiagLogN13K3DownstreamSufficiency(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3K3DownstreamPrimaryBase, "HEAD") == requiredFIX001P3K3DownstreamPrimaryBase
	result := fix001P3K3DownstreamResult{SchemaVersion: "fix-001-p3-diag-logn13-k3-downstream-sufficiency.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, PublicLikeThreshold: k3DownstreamPublicThreshold, Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3K3DownstreamPrimaryBase, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3K3DownstreamSecondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "logn13_only": cfg.LogN == 13, "fixed_k3_candidate": true, "q0_q1_q2_profile": "56/39/40", "g0_guard_bits": 1, "local_q2_only_at_f0": true, "f0_guard_bits": 3, "common_plan_scale_2^92": true, "validated_oracle_powers": true, "no_candidate_retuning": true, "no_generated_power_redesign": true, "no_q3_plus": true, "no_extra_q_levels": true, "no_c2s_s2c_mod1_change": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3K3DownstreamSecondary {
		result.Classification, result.FirstBlocker = "logn13_k3_downstream_precondition_mismatch", "startup provenance or Secondary precondition"
		return fix001P3K3DownstreamWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	var realEvidence, imagEvidence fix001P3LocalQ2BranchEvidence
	candidate, realRun, imagRun, err := psRescaleGuardMakeCandidateWithOverrides(profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch, profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, 3}, "D0-K3-G0-one-local-q2-F0-three", &profile.InitialReal, &profile.InitialImag, profile.Boundaries, fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &realEvidence, 3), fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &imagEvidence, 3))
	if err != nil {
		return err
	}
	d0Valid := fix001P3LocalQ2SweepBranchValid(realEvidence) && fix001P3LocalQ2SweepBranchValid(imagEvidence) && candidate.FinalRealError > 1.2e-8 && candidate.FinalRealError < 1.25e-8 && candidate.FinalImagError > 1.1e-8 && candidate.FinalImagError < 1.2e-8
	result.D0Control = map[string]interface{}{"candidate": q056CompactCandidate(candidate), "real_expansion": realEvidence.Expansion, "real_guard": realEvidence.Guard, "real_rescale": realEvidence.Rescale, "real_contraction": realEvidence.Contraction, "imag_expansion": imagEvidence.Expansion, "imag_guard": imagEvidence.Guard, "imag_rescale": imagEvidence.Rescale, "imag_contraction": imagEvidence.Contraction, "accepted_real_error": 1.223639867209414e-8, "accepted_imag_error": 1.1510743691545144e-8, "valid": d0Valid}
	if !d0Valid {
		result.Classification, result.FirstBlocker = "logn13_k3_downstream_precondition_mismatch", "D0 k=3 local-q2 control"
		return fix001P3K3DownstreamWrite(result, outPath)
	}
	planScale := precisionSweepScale(92)
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return err
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return err
	}
	result.Normalization = map[string]interface{}{
		"plan_scale":     finalizationScaleString(planScale),
		"real_path":      fix001P3K3DownstreamPathSummary(realPath),
		"imag_path":      fix001P3K3DownstreamPathSummary(imagPath),
		"real_schedules": fix001P3K3DownstreamSchedule(realPath),
		"imag_schedules": fix001P3K3DownstreamSchedule(imagPath),
	}
	if realPath.FastFinal == nil || imagPath.FastFinal == nil {
		result.Classification, result.FirstBlocker = "logn13_k3_downstream_insufficient", fmt.Sprintf("normalized downstream path returned nil final: real=%t imag=%t", realPath.FastFinal != nil, imagPath.FastFinal != nil)
		return fix001P3K3DownstreamWrite(result, outPath)
	}
	result.Normalization["real_poly_meta"] = realPath.PolyMeta
	result.Normalization["imag_poly_meta"] = imagPath.PolyMeta
	result.Normalization["real_first_failure"] = realPath.FirstFailure
	result.Normalization["imag_first_failure"] = imagPath.FirstFailure
	result.Normalization["real_output_level"] = realPath.FastFinal.Level()
	result.Normalization["imag_output_level"] = imagPath.FastFinal.Level()
	result.Normalization["real_output_scale"] = finalizationScaleString(realPath.FastFinal.Scale)
	result.Normalization["imag_output_scale"] = finalizationScaleString(imagPath.FastFinal.Scale)
	if realPath.FirstFailure != "none" || imagPath.FirstFailure != "none" || !realPath.FastRowsMatch || !realPath.RestoreRowsMatch || !imagPath.FastRowsMatch || !imagPath.RestoreRowsMatch {
		result.Classification, result.FirstBlocker = "logn13_k3_downstream_insufficient", fmt.Sprintf("normalized downstream path real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)
		return fix001P3K3DownstreamWrite(result, outPath)
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	result.EvalMod = map[string]interface{}{"real_fast_vs_standard": realFinal.FastVsStandard, "imag_fast_vs_standard": imagFinal.FastVsStandard, "real_fast_vs_normalized": realFinal.FastVsNormalized, "imag_fast_vs_normalized": imagFinal.FastVsNormalized, "real_normalized_vs_coherent": realFinal.NormalizedVsCoherent, "imag_normalized_vs_coherent": imagFinal.NormalizedVsCoherent, "real_restore_rows_match": realFinal.RestoreRowsMatch, "imag_restore_rows_match": imagFinal.RestoreRowsMatch}
	publicEvidence, err := fix001P3K3RunPublicLike(realPath, imagPath, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	result.S2CFinalization = map[string]interface{}{"post_s2c_fast_vs_standard": publicEvidence.PostS2C, "unpack_finalization_fast_vs_standard_core": publicEvidence.FinalVsStdCore, "standard_like": publicEvidence.StandardLike, "public_like": publicEvidence.PublicLike, "input_unchanged": publicEvidence.InputUnchanged}
	result.FinalMetadata = publicEvidence.Metadata
	result.LocalBudget = map[string]interface{}{"target": psRescaleGuardBudget, "real": candidate.FinalRealError, "imag": candidate.FinalImagError, "pass": candidate.FinalRealError <= psRescaleGuardBudget && candidate.FinalImagError <= psRescaleGuardBudget, "explicitly_not_system_target": true}
	result.PublicLikePass = publicEvidence.PublicLike != nil && publicEvidence.PublicLike.Pass
	result.SystemSufficient = result.PublicLikePass && publicEvidence.Metadata["pass"] == true && publicEvidence.InputUnchanged && realFinal.FastVsStandard.Pass && imagFinal.FastVsStandard.Pass
	if result.SystemSufficient {
		result.Classification, result.FirstBlocker = "logn13_k3_downstream_sufficient_despite_local_budget_miss", "generated-power error remains a separate blocker"
	} else {
		result.Classification = "logn13_k3_downstream_insufficient"
		result.FirstBlocker = "public-like system threshold or downstream semantic/metadata contract"
	}
	_ = realRun
	_ = imagRun
	return fix001P3K3DownstreamWrite(result, outPath)
}
