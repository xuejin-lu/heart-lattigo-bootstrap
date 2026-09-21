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
	fix001P3PostS2CScaleCollapseThreshold = 1e-2
	fix001P3PostS2CScaleCollapseExact     = 1e-12
)

type fix001P3PostS2CScaleCollapseResult struct {
	SchemaVersion  string                   `json:"schema_version"`
	Timestamp      time.Time                `json:"timestamp"`
	Primary        RepositoryMetadata       `json:"primary_repository"`
	Secondary      RepositoryMetadata       `json:"secondary_repository"`
	Environment    EnvironmentMetadata      `json:"environment"`
	Config         BootstrapConfig          `json:"config"`
	FixedProfile   map[string]interface{}   `json:"fixed_profile"`
	Provenance     map[string]interface{}   `json:"provenance"`
	R0Replay       map[string]interface{}   `json:"r0_replay"`
	Checkpoints    []map[string]interface{} `json:"downstream_checkpoints"`
	FirstBoundary  map[string]interface{}   `json:"first_convergence_boundary"`
	R2Proof        map[string]interface{}   `json:"r2_metadata_vs_arithmetic"`
	R3Counterfact  map[string]interface{}   `json:"r3_metadata_counterfactual"`
	R4E2E          map[string]interface{}   `json:"r4_exact_e2e"`
	R5SourceAudit  map[string]interface{}   `json:"r5_source_audit"`
	Classification string                   `json:"classification"`
	Validation     map[string]interface{}   `json:"validation"`
}

type fix001P3PostS2CScaleCollapseBranch struct {
	Core              *rlwe.Ciphertext
	UnpackInput       *rlwe.Ciphertext
	Unpacked          *rlwe.Ciphertext
	FinalizationInput *rlwe.Ciphertext
	AfterIMForm       *rlwe.Ciphertext
	Final             *rlwe.Ciphertext
	PreservedFinal    *rlwe.Ciphertext
	CoreValues        []complex128
	UnpackedValues    []complex128
	FinalValues       []complex128
	PreservedValues   []complex128
}

func fix001P3PostS2CScaleCollapseWrite(result fix001P3PostS2CScaleCollapseResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func fix001P3PostS2CScaleCollapseMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return map[string]interface{}{
		"level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree(), "n": ct.N(),
		"log_slots": ct.LogSlots(), "is_ntt": ct.IsNTT, "is_montgomery": ct.IsMontgomery,
		"log_dimensions": map[string]int{"rows": ct.LogDimensions.Rows, "cols": ct.LogDimensions.Cols},
	}
}

func fix001P3PostS2CScaleCollapseMetadataEqual(a, b *rlwe.Ciphertext) bool {
	return a.Level() == b.Level() && a.Scale.Equal(b.Scale) && a.Degree() == b.Degree() && a.N() == b.N() && a.LogSlots() == b.LogSlots() && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && a.LogDimensions == b.LogDimensions
}

func fix001P3PostS2CScaleCollapseRowsEqual(a, b *rlwe.Ciphertext) bool {
	return allBoolValues(fix001P3ActualForcedHashEqual(fix001P3ActualForcedRowHashes(a), fix001P3ActualForcedRowHashes(b)))
}

func fix001P3PostS2CScaleCollapseCheckpoint(name string, corrected, uncorrected *rlwe.Ciphertext, correctedValues, uncorrectedValues []complex128) (map[string]interface{}, error) {
	difference, err := fix001P3PublicScaleResetMetric(uncorrectedValues, correctedValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return nil, err
	}
	correctedRows := fix001P3ActualForcedRowHashes(corrected)
	uncorrectedRows := fix001P3ActualForcedRowHashes(uncorrected)
	return map[string]interface{}{
		"name":                   name,
		"corrected":              map[string]interface{}{"metadata": fix001P3PostS2CScaleCollapseMetadata(corrected), "row_hashes": correctedRows},
		"uncorrected":            map[string]interface{}{"metadata": fix001P3PostS2CScaleCollapseMetadata(uncorrected), "row_hashes": uncorrectedRows},
		"c_vs_u":                 difference,
		"metadata_equal":         fix001P3PostS2CScaleCollapseMetadataEqual(corrected, uncorrected),
		"coefficient_rows_equal": fix001P3PostS2CScaleCollapseRowsEqual(corrected, uncorrected),
		"row_hashes_equal":       fix001P3ActualForcedHashEqual(correctedRows, uncorrectedRows),
	}, nil
}

func runFIX001P3DiagP93PostS2CScaleCollapseCausalProof(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3PostS2CScaleCollapseResult{
		SchemaVersion: "fix-001-p3-diag-p93-post-s2c-scale-collapse-causal-proof.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "ps_authority": "Q012-wide", "plan_scale": "2^93", "slots": 4096, "threshold": fix001P3PostS2CScaleCollapseThreshold},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "expected_secondary_head": requiredFIX001P3ActualForcedSecondary, "expected_secondary_diff_sha256": requiredFIX001P3ActualForcedDiff},
		Validation:   map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_coefficient_correction": true, "no_parameter_tuning": true, "no_ps_repair": true, "no_c2s_repair": true, "no_s2c_repair": true, "no_finalizer_repair": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3ActualForcedSecondary || diffHash != requiredFIX001P3ActualForcedDiff || !secondaryDirty {
		result.Classification = "P93_POST_S2C_REPLAY_CONFLICT"
		result.Validation["secondary_handoff_state_match"] = false
		return fix001P3PostS2CScaleCollapseWrite(result, outPath)
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
	actualReal, err := profile.Fast.EvalMod(fastRealInput.CopyNew())
	if err != nil {
		return err
	}
	actualImag, err := profile.Fast.EvalMod(fastImagInput.CopyNew())
	if err != nil {
		return err
	}
	correctedReal := actualReal.CopyNew()
	correctedImag := actualImag.CopyNew()
	correctedReal.Scale = fastRealInput.Scale
	correctedImag.Scale = fastImagInput.Scale
	ordinaryReal, ordinaryImag := profile.Inputs.OrdinaryReal.CopyNew(), profile.Inputs.OrdinaryImag.CopyNew()
	forcedRealPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastRealInput.CopyNew(), ordinaryReal.CopyNew(), profile.Fast, profile.Standard, params, zero, profile.StandardSK, planScale, nil)
	if err != nil {
		return err
	}
	forcedImagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastImagInput.CopyNew(), ordinaryImag.CopyNew(), profile.Fast, profile.Standard, params, zero, profile.StandardSK, planScale, nil)
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

	runBranch := func(real, imag *rlwe.Ciphertext) (fix001P3PostS2CScaleCollapseBranch, error) {
		_, ctxtN1, ctxtN2, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*productionInput.CopyNew()})
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		core, err := profile.Fast.SlotsToCoeffs(real.CopyNew(), imag.CopyNew())
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		unpackInput := core.CopyNew()
		unpacked, err := profile.Fast.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*unpackInput.CopyNew()}, ctxtN1, ctxtN2)
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		if len(unpacked) != 1 {
			return fix001P3PostS2CScaleCollapseBranch{}, fmt.Errorf("expected one unpacked ciphertext, got %d", len(unpacked))
		}
		finalizationInput := unpacked[0].CopyNew()
		before, ordinaryFinal, err := fix001P3E2EFinalizeOracle(&unpacked[0], profile.BTP)
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		if !fix001P3PostS2CScaleCollapseMetadataEqual(before, finalizationInput) {
			return fix001P3PostS2CScaleCollapseBranch{}, fmt.Errorf("finalization input copy changed metadata")
		}
		afterIMForm := finalizationInput.CopyNew()
		if err := finalizationTransformMaintained(afterIMForm, profile.Residual, false); err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		afterIMForm.IsMontgomery = false
		preservedFinal := afterIMForm.CopyNew()
		coreValues, err := psGlobalDecode(params, core)
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		unpackedValues, err := fix001P3S2CDecode(params, &unpacked[0], zero)
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		finalValues, err := decodeWithSecret(profile.Residual, ordinaryFinal, zeroSecret(profile.Residual))
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		preservedValues, err := decodeWithSecret(profile.Residual, preservedFinal, zeroSecret(profile.Residual))
		if err != nil {
			return fix001P3PostS2CScaleCollapseBranch{}, err
		}
		return fix001P3PostS2CScaleCollapseBranch{Core: core, UnpackInput: unpackInput, Unpacked: &unpacked[0], FinalizationInput: finalizationInput, AfterIMForm: afterIMForm, Final: ordinaryFinal, PreservedFinal: preservedFinal, CoreValues: coreValues, UnpackedValues: unpackedValues, FinalValues: finalValues, PreservedValues: preservedValues}, nil
	}

	uncorrected, err := runBranch(actualReal, actualImag)
	if err != nil {
		return err
	}
	corrected, err := runBranch(correctedReal, correctedImag)
	if err != nil {
		return err
	}
	workloadValues := reproducibleValues(profile.Residual.MaxSlots())
	uncorrectedCore, err := fix001P3PublicScaleResetMetric(matchedStandardCoreValues, uncorrected.CoreValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	correctedCore, err := fix001P3PublicScaleResetMetric(matchedStandardCoreValues, corrected.CoreValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	r0Replay := map[string]interface{}{
		"uncorrected_post_s2c_vs_matched_standard": uncorrectedCore,
		"corrected_post_s2c_vs_matched_standard":   correctedCore,
		"expected_uncorrected_max_component_abs":   0.0402736269192105,
		"expected_corrected_max_component_abs":     0.006934820352192803,
		"threshold":                                fix001P3PostS2CScaleCollapseThreshold,
		"replay_match":                             math.Abs(uncorrectedCore.MaxComponentAbs-0.0402736269192105) < 1e-9 && math.Abs(correctedCore.MaxComponentAbs-0.006934820352192803) < 1e-9,
	}
	result.R0Replay = r0Replay
	if !r0Replay["replay_match"].(bool) {
		result.Classification = "P93_POST_S2C_REPLAY_CONFLICT"
		return fix001P3PostS2CScaleCollapseWrite(result, outPath)
	}

	checkpoints := []struct {
		name   string
		c, u   *rlwe.Ciphertext
		cv, uv []complex128
	}{
		{"s2c_core_output", corrected.Core, uncorrected.Core, corrected.CoreValues, uncorrected.CoreValues},
		{"unpack_input", corrected.UnpackInput, uncorrected.UnpackInput, corrected.CoreValues, uncorrected.CoreValues},
		{"unpack_output", corrected.Unpacked, uncorrected.Unpacked, corrected.UnpackedValues, uncorrected.UnpackedValues},
		{"finalization_input", corrected.FinalizationInput, uncorrected.FinalizationInput, corrected.UnpackedValues, uncorrected.UnpackedValues},
		{"finalization_after_imform", corrected.AfterIMForm, uncorrected.AfterIMForm, corrected.PreservedValues, uncorrected.PreservedValues},
		{"finalization_after_scale_reset", corrected.Final, uncorrected.Final, corrected.FinalValues, uncorrected.FinalValues},
	}
	for _, checkpoint := range checkpoints {
		value, err := fix001P3PostS2CScaleCollapseCheckpoint(checkpoint.name, checkpoint.c, checkpoint.u, checkpoint.cv, checkpoint.uv)
		if err != nil {
			return err
		}
		result.Checkpoints = append(result.Checkpoints, value)
	}

	firstIndex := -1
	for i := 1; i < len(result.Checkpoints); i++ {
		previous := result.Checkpoints[i-1]
		current := result.Checkpoints[i]
		previousMetric := previous["c_vs_u"].(*semanticBisectMetric)
		currentMetric := current["c_vs_u"].(*semanticBisectMetric)
		previousMetadataEqual := previous["metadata_equal"].(bool)
		currentMetadataEqual := current["metadata_equal"].(bool)
		currentRowsEqual := current["coefficient_rows_equal"].(bool)
		if previousMetric.MaxComponentAbs > fix001P3PostS2CScaleCollapseExact && currentMetric.MaxComponentAbs <= fix001P3PostS2CScaleCollapseExact && currentMetadataEqual && currentRowsEqual {
			firstIndex = i
			_ = previousMetadataEqual
			break
		}
	}
	if firstIndex < 0 {
		for i := 1; i < len(result.Checkpoints); i++ {
			previous := result.Checkpoints[i-1]
			current := result.Checkpoints[i]
			if previous["metadata_equal"].(bool) == false && current["metadata_equal"].(bool) && current["coefficient_rows_equal"].(bool) {
				firstIndex = i
				break
			}
		}
	}
	if firstIndex < 0 {
		result.Classification = "P93_POST_S2C_TRACE_ALIGNMENT_INVALID"
		result.R2Proof = map[string]interface{}{"classification": "unresolved", "reason": "no downstream corrected-vs-uncorrected convergence boundary observed"}
		return fix001P3PostS2CScaleCollapseWrite(result, outPath)
	}
	first := result.Checkpoints[firstIndex]
	previous := result.Checkpoints[firstIndex-1]
	firstName := first["name"].(string)
	firstBoundaryClass := "P93_FINALIZATION_SCALE_METADATA_COLLAPSE"
	if firstName == "unpack_output" || firstName == "finalization_input" {
		firstBoundaryClass = "P93_UNPACK_SCALE_METADATA_COLLAPSE"
	}
	result.FirstBoundary = map[string]interface{}{"checkpoint": firstName, "previous_checkpoint": previous["name"], "operation": "diagnostic finalization scale normalization", "classification": firstBoundaryClass, "previous_c_vs_u": previous["c_vs_u"], "first_equal_c_vs_u": first["c_vs_u"], "metadata_equal_after": first["metadata_equal"], "coefficient_rows_equal_after": first["coefficient_rows_equal"]}
	result.R2Proof = map[string]interface{}{
		"classification":    "metadata_only",
		"first_boundary":    firstName,
		"before_rows_equal": previous["coefficient_rows_equal"], "after_rows_equal": first["coefficient_rows_equal"],
		"before_metadata_equal": previous["metadata_equal"], "after_metadata_equal": first["metadata_equal"],
		"no_coefficient_correction": true,
		"scale_only_change":         true,
		"evidence":                  "C/U maintained coefficient row hashes remain equal while the corrected incoming Scale is replaced by residual DefaultScale at finalization",
	}

	uncorrectedE2E, err := fix001P3PublicScaleResetMetric(workloadValues, uncorrected.FinalValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	correctedOnlyE2E, err := fix001P3PublicScaleResetMetric(workloadValues, corrected.FinalValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	correctedPreservedE2E, err := fix001P3PublicScaleResetMetric(workloadValues, corrected.PreservedValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	correctedPreservedVsCorrected, err := fix001P3PublicScaleResetMetric(corrected.FinalValues, corrected.PreservedValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	correctedPreservedVsUncorrected, err := fix001P3PublicScaleResetMetric(uncorrected.FinalValues, corrected.PreservedValues, fix001P3PostS2CScaleCollapseThreshold)
	if err != nil {
		return err
	}
	result.R3Counterfact = map[string]interface{}{
		"boundary":                               firstName,
		"corrected_incoming_scale":               finalizationScaleString(corrected.AfterIMForm.Scale),
		"ordinary_corrected_outgoing_scale":      finalizationScaleString(corrected.Final.Scale),
		"counterfactual_outgoing_scale":          finalizationScaleString(corrected.PreservedFinal.Scale),
		"rows_unchanged_by_scale_counterfactual": fix001P3PostS2CScaleCollapseRowsEqual(corrected.Final, corrected.PreservedFinal),
		"metadata_before":                        fix001P3PostS2CScaleCollapseMetadata(corrected.AfterIMForm),
		"ordinary_metadata_after":                fix001P3PostS2CScaleCollapseMetadata(corrected.Final),
		"counterfactual_metadata_after":          fix001P3PostS2CScaleCollapseMetadata(corrected.PreservedFinal),
		"counterfactual_vs_ordinary_corrected":   correctedPreservedVsCorrected,
		"counterfactual_vs_ordinary_uncorrected": correctedPreservedVsUncorrected,
	}
	result.R4E2E = map[string]interface{}{
		"message_domain":               "reproducible deterministic workload",
		"uncorrected_current_final":    uncorrectedE2E,
		"evalmod_corrected_only_final": correctedOnlyE2E,
		"evalmod_corrected_plus_downstream_metadata_preserved_final": correctedPreservedE2E,
		"threshold":                        fix001P3PostS2CScaleCollapseThreshold,
		"counterfactual_reaches_threshold": correctedPreservedE2E.Pass,
		"final_metadata":                   fix001P3PostS2CScaleCollapseMetadata(corrected.PreservedFinal),
	}
	result.R5SourceAudit = map[string]interface{}{
		"first_convergence_function":         "(*FastEvaluator).finalizeFastPublicCiphertext",
		"first_convergence_file":             "circuits/ckks/bootstrapping/fast_bootstrap.go",
		"first_convergence_lines":            "247-253",
		"unpack_function":                    "(*FastEvaluator).UnpackAndSwitchN2ToN1",
		"unpack_file":                        "circuits/ckks/bootstrapping/fast_packing.go",
		"unpack_lines":                       "322-370",
		"scale_overwrite_expression":         "ct.Scale = params.DefaultScale()",
		"incoming_corrected_scale":           finalizationScaleString(corrected.AfterIMForm.Scale),
		"outgoing_scale":                     finalizationScaleString(corrected.Final.Scale),
		"primary_oracle_helper":              "fix001P3E2EFinalizeOracle in fix001_p3_validate_logn13_e2e_runner.go",
		"primary_oracle_scale_expression":    "after.Scale = params.ResidualParameters.DefaultScale()",
		"standard_analogous_contract":        "Standard BootstrapMany also restores eval.ResidualParameters.DefaultScale after unpack; Standard Evaluator.UnpackAndSwitchN2ToN1 itself does not assign Scale",
		"operation_treats_scale_as_metadata": true,
		"smallest_candidate_repair":          "preserve the corrected incoming Scale through the Fast public finalization boundary, subject to public Standard compatibility review",
		"source_read_only":                   true,
	}
	if firstBoundaryClass == "P93_UNPACK_SCALE_METADATA_COLLAPSE" {
		result.Classification = "P93_UNPACK_SCALE_METADATA_COLLAPSE"
	} else if correctedPreservedE2E.Pass {
		result.Classification = "P93_EVALMOD_PLUS_DOWNSTREAM_METADATA_SUFFICIENT_FOR_1E2"
	} else {
		result.Classification = "P93_DOWNSTREAM_METADATA_CAUSAL_BUT_SYSTEM_STILL_FAILS_1E2"
	}
	return fix001P3PostS2CScaleCollapseWrite(result, outPath)
}
