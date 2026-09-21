package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	fix001P3PublicScaleResetInputScaleBits   = 50
	fix001P3PublicScaleResetDefaultScaleBits = 45
	fix001P3PublicScaleResetCorrectionFloor  = 1e-10
	fix001P3PublicScaleResetThreshold        = 1e-2
)

type fix001P3PublicScaleResetDownstreamRun struct {
	Core       *rlwe.Ciphertext
	Unpacked   *rlwe.Ciphertext
	Manual     *rlwe.Ciphertext
	CoreValues []complex128
	Final      []complex128
}

type fix001P3PublicScaleResetResult struct {
	SchemaVersion  string                 `json:"schema_version"`
	Timestamp      time.Time              `json:"timestamp"`
	Primary        RepositoryMetadata     `json:"primary_repository"`
	Secondary      RepositoryMetadata     `json:"secondary_repository"`
	Environment    EnvironmentMetadata    `json:"environment"`
	Config         BootstrapConfig        `json:"config"`
	FixedProfile   map[string]interface{} `json:"fixed_profile"`
	Provenance     map[string]interface{} `json:"provenance"`
	R0Boundary     map[string]interface{} `json:"r0_public_scale_reset_boundary"`
	R1Correction   map[string]interface{} `json:"r1_metadata_only_correction"`
	R2EvalMod      map[string]interface{} `json:"r2_evalmod_corrected_vs_uncorrected"`
	R3Downstream   map[string]interface{} `json:"r3_downstream_production_replay"`
	R4Closure      map[string]interface{} `json:"r4_vector_causal_closure"`
	R5SourceAudit  map[string]interface{} `json:"r5_source_audit"`
	Classification string                 `json:"classification"`
	Validation     map[string]interface{} `json:"validation"`
}

func fix001P3PublicScaleResetWrite(result fix001P3PublicScaleResetResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func fix001P3PublicScaleResetMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return map[string]interface{}{
		"level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree(), "n": ct.N(),
		"is_ntt": ct.IsNTT, "is_montgomery": ct.IsMontgomery,
	}
}

func fix001P3PublicScaleResetMetric(reference, actual []complex128, threshold float64) (*semanticBisectMetric, error) {
	metric, err := semanticBisectMetricFor(reference, actual, threshold)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func fix001P3PublicScaleResetRowsEqual(a, b *rlwe.Ciphertext) bool {
	return allBoolValues(fix001P3ActualForcedHashEqual(fix001P3ActualForcedRowHashes(a), fix001P3ActualForcedRowHashes(b)))
}

func allBoolValues(values map[string]bool) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}

func safeRatio(numerator, denominator float64) float64 {
	if denominator == 0 {
		if numerator == 0 {
			return 0
		}
		return math.Inf(1)
	}
	return numerator / denominator
}

func runFIX001P3DiagP93EvalModPublicScaleResetCausalProof(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3PublicScaleResetResult{
		SchemaVersion: "fix-001-p3-diag-p93-evalmod-public-scale-reset-causal-proof.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "ps_authority": "Q012-wide", "plan_scale": "2^93", "slots": 4096, "threshold": fix001P3PublicScaleResetThreshold},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "expected_secondary_head": requiredFIX001P3ActualForcedSecondary, "expected_secondary_diff_sha256": requiredFIX001P3ActualForcedDiff},
		Validation:   map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_coefficient_correction": true, "no_q_or_plan_scale_tuning": true, "no_ps_repair": true, "no_c2s_repair": true, "no_s2c_repair": true, "no_finalizer_repair": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3ActualForcedSecondary || diffHash != requiredFIX001P3ActualForcedDiff || !secondaryDirty {
		result.Classification = "P93_PUBLIC_SCALE_RESET_REPLAY_CONFLICT"
		result.Validation["secondary_handoff_state_match"] = false
		return fix001P3PublicScaleResetWrite(result, outPath)
	}

	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	zero := zeroSecret(params)
	planScale := precisionSweepScale(93)
	productionInput := reproducibleInput(profile.Residual, profile.BTP)
	packed, _, _, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*productionInput.CopyNew()})
	if err != nil {
		return err
	}
	scaled, _, err := profile.Fast.ScaleDown(&packed[0])
	if err != nil {
		return err
	}
	modUp, err := profile.Fast.ModUp(scaled)
	if err != nil {
		return err
	}
	fastRealInput, fastImagInput, err := profile.Fast.CoeffsToSlots(modUp)
	if err != nil {
		return err
	}
	ordinaryReal, ordinaryImag := profile.Inputs.OrdinaryReal.CopyNew(), profile.Inputs.OrdinaryImag.CopyNew()
	actualReal, err := profile.Fast.EvalMod(fastRealInput.CopyNew())
	if err != nil {
		return err
	}
	actualImag, err := profile.Fast.EvalMod(fastImagInput.CopyNew())
	if err != nil {
		return err
	}
	internalReal, err := fix001P3ActualForcedReplayFastEvalMod(profile.Fast, fastRealInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	internalImag, err := fix001P3ActualForcedReplayFastEvalMod(profile.Fast, fastImagInput.CopyNew(), planScale)
	if err != nil {
		return err
	}

	actualRealRows := fix001P3ActualForcedRowHashes(actualReal)
	actualImagRows := fix001P3ActualForcedRowHashes(actualImag)
	internalRealRows := fix001P3ActualForcedRowHashes(internalReal)
	internalImagRows := fix001P3ActualForcedRowHashes(internalImag)
	r0RealRows := fix001P3ActualForcedHashEqual(actualRealRows, internalRealRows)
	r0ImagRows := fix001P3ActualForcedHashEqual(actualImagRows, internalImagRows)
	_, r0RealValues, err := semanticBisectView(internalReal, params, zero)
	if err != nil {
		return err
	}
	_, actualRealValues, err := semanticBisectView(actualReal, params, zero)
	if err != nil {
		return err
	}
	_, r0ImagValues, err := semanticBisectView(internalImag, params, zero)
	if err != nil {
		return err
	}
	_, actualImagValues, err := semanticBisectView(actualImag, params, zero)
	if err != nil {
		return err
	}
	r0RealMetric, err := fix001P3PublicScaleResetMetric(r0RealValues, actualRealValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	r0ImagMetric, err := fix001P3PublicScaleResetMetric(r0ImagValues, actualImagValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	result.R0Boundary = map[string]interface{}{
		"real":          map[string]interface{}{"internal": fix001P3PublicScaleResetMetadata(internalReal), "public": fix001P3PublicScaleResetMetadata(actualReal), "internal_scale_expected": "2^50", "public_scale_expected": "2^45", "scale_ratio": 32, "level_equal": internalReal.Level() == actualReal.Level(), "degree_equal": internalReal.Degree() == actualReal.Degree(), "row_hashes_equal": r0RealRows, "decoded_public_vs_internal": r0RealMetric},
		"imag":          map[string]interface{}{"internal": fix001P3PublicScaleResetMetadata(internalImag), "public": fix001P3PublicScaleResetMetadata(actualImag), "internal_scale_expected": "2^50", "public_scale_expected": "2^45", "scale_ratio": 32, "level_equal": internalImag.Level() == actualImag.Level(), "degree_equal": internalImag.Degree() == actualImag.Degree(), "row_hashes_equal": r0ImagRows, "decoded_public_vs_internal": r0ImagMetric},
		"metadata_only": true, "scale_expression": "ctOut.Scale = eval.Parameters.BootstrappingParameters.DefaultScale()",
	}

	correctedReal := actualReal.CopyNew()
	correctedImag := actualImag.CopyNew()
	correctedReal.Scale = fastRealInput.Scale
	correctedImag.Scale = fastImagInput.Scale
	forcedRealPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastRealInput.CopyNew(), ordinaryReal.CopyNew(), profile.Fast, profile.Standard, params, zero, profile.StandardSK, planScale, nil)
	if err != nil {
		return err
	}
	forcedImagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastImagInput.CopyNew(), ordinaryImag.CopyNew(), profile.Fast, profile.Standard, params, zero, profile.StandardSK, planScale, nil)
	if err != nil {
		return err
	}
	correctedRealValues, err := fix001P3S2CDecode(params, correctedReal, zero)
	if err != nil {
		return err
	}
	correctedImagValues, err := fix001P3S2CDecode(params, correctedImag, zero)
	if err != nil {
		return err
	}
	forcedRealValues, err := fix001P3S2CDecode(params, forcedRealPath.FastFinal, zero)
	if err != nil {
		return err
	}
	forcedImagValues, err := fix001P3S2CDecode(params, forcedImagPath.FastFinal, zero)
	if err != nil {
		return err
	}
	correctedRealForced, err := fix001P3PublicScaleResetMetric(forcedRealValues, correctedRealValues, fix001P3PublicScaleResetCorrectionFloor)
	if err != nil {
		return err
	}
	correctedImagForced, err := fix001P3PublicScaleResetMetric(forcedImagValues, correctedImagValues, fix001P3PublicScaleResetCorrectionFloor)
	if err != nil {
		return err
	}
	result.R1Correction = map[string]interface{}{
		"real":       map[string]interface{}{"before": fix001P3PublicScaleResetMetadata(actualReal), "after": fix001P3PublicScaleResetMetadata(correctedReal), "scale_only": true, "row_hashes_unchanged": fix001P3PublicScaleResetRowsEqual(actualReal, correctedReal), "corrected_vs_forced": correctedRealForced},
		"imag":       map[string]interface{}{"before": fix001P3PublicScaleResetMetadata(actualImag), "after": fix001P3PublicScaleResetMetadata(correctedImag), "scale_only": true, "row_hashes_unchanged": fix001P3PublicScaleResetRowsEqual(actualImag, correctedImag), "corrected_vs_forced": correctedImagForced},
		"correction": "corrected.Scale = original_C2S_input.Scale",
	}

	uncorrectedRealMetric, err := fix001P3PublicScaleResetMetric(forcedRealValues, actualRealValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	uncorrectedImagMetric, err := fix001P3PublicScaleResetMetric(forcedImagValues, actualImagValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	matchedRealStandard, err := fix001P3S2CDecode(params, forcedRealPath.StandardFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedImagStandard, err := fix001P3S2CDecode(params, forcedImagPath.StandardFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	correctedRealStandard, err := fix001P3PublicScaleResetMetric(matchedRealStandard, correctedRealValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	correctedImagStandard, err := fix001P3PublicScaleResetMetric(matchedImagStandard, correctedImagValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	uncorrectedRealStandard, err := fix001P3PublicScaleResetMetric(matchedRealStandard, actualRealValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	uncorrectedImagStandard, err := fix001P3PublicScaleResetMetric(matchedImagStandard, actualImagValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	result.R2EvalMod = map[string]interface{}{"real": map[string]interface{}{"corrected_vs_matched_standard": correctedRealStandard, "uncorrected_vs_matched_standard": uncorrectedRealStandard, "uncorrected_public_vs_forced": uncorrectedRealMetric}, "imag": map[string]interface{}{"corrected_vs_matched_standard": correctedImagStandard, "uncorrected_vs_matched_standard": uncorrectedImagStandard, "uncorrected_public_vs_forced": uncorrectedImagMetric}, "threshold": fix001P3PublicScaleResetThreshold, "expected_forced_real": 0.004183019582623534, "expected_forced_imag": 0.004261561142162278, "expected_uncorrected_real": 0.022317936759951508, "expected_uncorrected_imag": 0.020569287028685726}

	runDownstream := func(real, imag *rlwe.Ciphertext) (fix001P3PublicScaleResetDownstreamRun, error) {
		_, ctxtN1, ctxtN2, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*productionInput.CopyNew()})
		if err != nil {
			return fix001P3PublicScaleResetDownstreamRun{}, err
		}
		core, err := profile.Fast.SlotsToCoeffs(real.CopyNew(), imag.CopyNew())
		if err != nil {
			return fix001P3PublicScaleResetDownstreamRun{}, err
		}
		unpacked, err := profile.Fast.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*core.CopyNew()}, ctxtN1, ctxtN2)
		if err != nil {
			return fix001P3PublicScaleResetDownstreamRun{}, err
		}
		if len(unpacked) != 1 {
			return fix001P3PublicScaleResetDownstreamRun{}, fmt.Errorf("expected one unpacked ciphertext, got %d", len(unpacked))
		}
		_, manual, err := fix001P3E2EFinalizeOracle(&unpacked[0], profile.BTP)
		if err != nil {
			return fix001P3PublicScaleResetDownstreamRun{}, err
		}
		coreValues, err := psGlobalDecode(params, core)
		if err != nil {
			return fix001P3PublicScaleResetDownstreamRun{}, err
		}
		finalValues, err := decodeWithSecret(profile.Residual, manual, zeroSecret(profile.Residual))
		if err != nil {
			return fix001P3PublicScaleResetDownstreamRun{}, err
		}
		return fix001P3PublicScaleResetDownstreamRun{Core: core, Unpacked: &unpacked[0], Manual: manual, CoreValues: coreValues, Final: finalValues}, nil
	}
	uncorrectedDownstream, err := runDownstream(actualReal, actualImag)
	if err != nil {
		return err
	}
	correctedDownstream, err := runDownstream(correctedReal, correctedImag)
	if err != nil {
		return err
	}
	workloadValues := reproducibleValues(profile.Residual.MaxSlots())
	uncorrectedCoreMetric, err := fix001P3PublicScaleResetMetric(workloadValues, uncorrectedDownstream.CoreValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	correctedCoreMetric, err := fix001P3PublicScaleResetMetric(workloadValues, correctedDownstream.CoreValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	matchedStandardCore, err := profile.Standard.SlotsToCoeffs(forcedRealPath.StandardFinal.CopyNew(), forcedImagPath.StandardFinal.CopyNew())
	if err != nil {
		return err
	}
	matchedStandardCoreValues, err := fix001P3S2CDecode(params, matchedStandardCore, profile.StandardSK)
	if err != nil {
		return err
	}
	uncorrectedCoreVsMatchedStandard, err := fix001P3PublicScaleResetMetric(matchedStandardCoreValues, uncorrectedDownstream.CoreValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	correctedCoreVsMatchedStandard, err := fix001P3PublicScaleResetMetric(matchedStandardCoreValues, correctedDownstream.CoreValues, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	uncorrectedFinalMetric, err := fix001P3PublicScaleResetMetric(workloadValues, uncorrectedDownstream.Final, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	correctedFinalMetric, err := fix001P3PublicScaleResetMetric(workloadValues, correctedDownstream.Final, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	result.R3Downstream = map[string]interface{}{
		"uncorrected": map[string]interface{}{"evalmod_real": fix001P3PublicScaleResetMetadata(actualReal), "evalmod_imag": fix001P3PublicScaleResetMetadata(actualImag), "post_s2c_f0": uncorrectedCoreVsMatchedStandard, "post_s2c_f0_vs_workload": uncorrectedCoreMetric, "final_public_like": uncorrectedFinalMetric, "final_metadata": fix001P3PublicScaleResetMetadata(uncorrectedDownstream.Manual)},
		"corrected":   map[string]interface{}{"evalmod_real": fix001P3PublicScaleResetMetadata(correctedReal), "evalmod_imag": fix001P3PublicScaleResetMetadata(correctedImag), "post_s2c_f0": correctedCoreVsMatchedStandard, "post_s2c_f0_vs_workload": correctedCoreMetric, "final_public_like": correctedFinalMetric, "final_metadata": fix001P3PublicScaleResetMetadata(correctedDownstream.Manual)},
		"threshold":   fix001P3PublicScaleResetThreshold,
	}

	uncorrectedRealVector, err := fix001P3S2CDecode(params, actualReal, zero)
	if err != nil {
		return err
	}
	uncorrectedImagVector, err := fix001P3S2CDecode(params, actualImag, zero)
	if err != nil {
		return err
	}
	correctedRealVector, err := fix001P3S2CDecode(params, correctedReal, zero)
	if err != nil {
		return err
	}
	correctedImagVector, err := fix001P3S2CDecode(params, correctedImag, zero)
	if err != nil {
		return err
	}
	uncorrectedCoreVector, err := psGlobalDecode(params, uncorrectedDownstream.Core)
	if err != nil {
		return err
	}
	correctedCoreVector, err := psGlobalDecode(params, correctedDownstream.Core)
	if err != nil {
		return err
	}
	uncorrectedFinalVector, correctedFinalVector := uncorrectedDownstream.Final, correctedDownstream.Final
	evalModDeltaReal, err := fix001P3PublicScaleResetMetric(uncorrectedRealVector, correctedRealVector, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	evalModDeltaImag, err := fix001P3PublicScaleResetMetric(uncorrectedImagVector, correctedImagVector, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	postS2CDelta, err := fix001P3PublicScaleResetMetric(uncorrectedCoreVector, correctedCoreVector, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	finalDelta, err := fix001P3PublicScaleResetMetric(uncorrectedFinalVector, correctedFinalVector, fix001P3PublicScaleResetThreshold)
	if err != nil {
		return err
	}
	inputDelta := math.Max(evalModDeltaReal.MaxComponentAbs, evalModDeltaImag.MaxComponentAbs)
	result.R4Closure = map[string]interface{}{"evalmod_real_corrected_minus_uncorrected": evalModDeltaReal, "evalmod_imag_corrected_minus_uncorrected": evalModDeltaImag, "post_s2c_corrected_minus_uncorrected": postS2CDelta, "final_corrected_minus_uncorrected": finalDelta, "post_s2c_amplification": safeRatio(postS2CDelta.MaxComponentAbs, inputDelta), "final_amplification": safeRatio(finalDelta.MaxComponentAbs, inputDelta), "vector_based": true, "no_exact_linearity_claim": true}

	result.R5SourceAudit = map[string]interface{}{
		"public_wrapper_file":                               "circuits/ckks/bootstrapping/fast_evalmod.go",
		"public_wrapper_function":                           "(*FastEvaluator).EvalMod",
		"public_scale_assignment":                           "ctOut.Scale = eval.Parameters.BootstrappingParameters.DefaultScale()",
		"internal_mod1_file":                                "circuits/ckks/mod1/fast.go",
		"internal_input_scale_capture":                      "inputScale := ct.Scale",
		"internal_final_scale_assignment":                   "res.Scale = inputScale",
		"internal_final_scale":                              finalizationScaleString(fastRealInput.Scale),
		"residual_default_scale":                            finalizationScaleString(profile.BTP.BootstrappingParameters.DefaultScale()),
		"wrapper_overwrites_internal_scale_unconditionally": true,
		"documented_reason_observed":                        "wrapper comment says restore Bootstrap boundary public default scale; no separate API justification for replacing the internal restored/input scale was found",
		"smallest_candidate_repair":                         "preserve the internal restored/input Scale at this boundary, subject to Standard/public API compatibility review",
		"source_read_only":                                  true,
	}

	correctedPass := correctedRealStandard.Pass && correctedImagStandard.Pass && correctedFinalMetric.Pass
	if !r0RealRows["c0q0"] || !r0ImagRows["c0q0"] || !r0RealRows["c0q1"] || !r0ImagRows["c0q1"] || !r0RealRows["c0q2"] || !r0ImagRows["c0q2"] {
		result.Classification = "P93_PUBLIC_SCALE_RESET_NOT_METADATA_ONLY"
	} else if correctedRealForced.MaxComponentAbs > fix001P3PublicScaleResetCorrectionFloor || correctedImagForced.MaxComponentAbs > fix001P3PublicScaleResetCorrectionFloor {
		result.Classification = "P93_SCALE_METADATA_CORRECTION_INSUFFICIENT"
	} else if !correctedPass {
		result.Classification = "P93_SCALE_METADATA_CAUSAL_BUT_SYSTEM_STILL_FAILS_1E2"
	} else {
		result.Classification = "P93_PUBLIC_EVALMOD_SCALE_RESET_CAUSAL_AND_1E2_SUFFICIENT"
	}
	return fix001P3PublicScaleResetWrite(result, outPath)
}
