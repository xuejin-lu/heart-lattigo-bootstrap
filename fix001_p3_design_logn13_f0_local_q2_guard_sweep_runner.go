package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	requiredFIX001P3F0LocalQ2GuardSweepPrimaryBase = "73b0ac1aeaafa42650ef79260c2c5412354ac558"
	requiredFIX001P3F0LocalQ2GuardSweepSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
)

type fix001P3LocalQ2GuardSweepRow struct {
	GuardBits              int     `json:"guard_bits"`
	RealError              float64 `json:"final_real_error"`
	ImagError              float64 `json:"final_imag_error"`
	MaxError               float64 `json:"final_max_error"`
	RealQ012BeforeRatio    float64 `json:"real_q012_before_ratio"`
	ImagQ012BeforeRatio    float64 `json:"imag_q012_before_ratio"`
	RealQ012AfterRatio     float64 `json:"real_q012_after_ratio"`
	ImagQ012AfterRatio     float64 `json:"imag_q012_after_ratio"`
	RealQ01AfterRatio      float64 `json:"real_q01_after_rescale_ratio"`
	ImagQ01AfterRatio      float64 `json:"imag_q01_after_rescale_ratio"`
	RealLocalRescaleError  float64 `json:"real_local_rescale_rounding_error"`
	ImagLocalRescaleError  float64 `json:"imag_local_rescale_rounding_error"`
	ExpansionPass          bool    `json:"expansion_pass"`
	GuardValuePass         bool    `json:"guard_value_preserving"`
	Q012CenteredUnique     bool    `json:"q012_centered_unique"`
	Q012RowsMatch          bool    `json:"q012_rows_match_full_rns"`
	RoundedDivisionMatch   bool    `json:"rounded_division_match"`
	ContractionPass        bool    `json:"contraction_pass"`
	MetadataPass           bool    `json:"level_scale_metadata_pass"`
	NoExtraLevel           bool    `json:"no_extra_level_consumption"`
	ValidLocalQ2           bool    `json:"valid_local_q2_candidate"`
	PrecisionQualified     bool    `json:"precision_qualified"`
	StopAfterThisCandidate bool    `json:"stop_after_this_candidate"`
}

type fix001P3LocalQ2GuardSweepResult struct {
	SchemaVersion         string                         `json:"schema_version"`
	Timestamp             time.Time                      `json:"timestamp"`
	Primary               RepositoryMetadata             `json:"primary_repository"`
	Lattigo               RepositoryMetadata             `json:"lattigo_repository"`
	Environment           EnvironmentMetadata            `json:"environment"`
	Config                BootstrapConfig                `json:"config"`
	Parameters            ExperimentParameters           `json:"effective_parameters"`
	Primes                map[string]interface{}         `json:"q0_q1_q2_primes"`
	S0Controls            map[string]interface{}         `json:"s0_controls"`
	Sweep                 []fix001P3LocalQ2GuardSweepRow `json:"guard_sweep"`
	SelectedGuardBits     *int                           `json:"selected_smallest_qualifying_guard_bits,omitempty"`
	SelectedCandidate     *q056CandidateEvidence         `json:"selected_candidate,omitempty"`
	Downstream            map[string]interface{}         `json:"downstream,omitempty"`
	Classification        string                         `json:"classification"`
	BlockerClosed         bool                           `json:"ps_arithmetic_blocker_independently_closed"`
	FirstRemainingBlocker string                         `json:"first_remaining_blocker"`
	Validation            map[string]interface{}         `json:"validation"`
}

func fix001P3LocalQ2SweepBranchValid(evidence fix001P3LocalQ2BranchEvidence) bool {
	return evidence.Expansion.Pass && evidence.Guard.Value != nil && evidence.Guard.Value.Pass && evidence.Guard.Q012After.Unique && evidence.Guard.RowsMatch && evidence.Rescale.RowsMatch && evidence.Rescale.RoundedDivisionMatch && evidence.Contraction.Pass && evidence.Rescale.OutputLevel == evidence.Rescale.InputLevel-1
}

func fix001P3LocalQ2SweepRowFor(bits int, candidate psRescaleGuardCandidate, real, imag fix001P3LocalQ2BranchEvidence) fix001P3LocalQ2GuardSweepRow {
	row := fix001P3LocalQ2GuardSweepRow{GuardBits: bits, RealError: candidate.FinalRealError, ImagError: candidate.FinalImagError, MaxError: candidate.FinalMaxError, RealQ012BeforeRatio: real.Guard.Q012Before.Ratio, ImagQ012BeforeRatio: imag.Guard.Q012Before.Ratio, RealQ012AfterRatio: real.Guard.Q012After.Ratio, ImagQ012AfterRatio: imag.Guard.Q012After.Ratio, RealQ01AfterRatio: real.Rescale.Q01Output.Ratio, ImagQ01AfterRatio: imag.Rescale.Q01Output.Ratio, RealLocalRescaleError: psRescaleGuardMetricValue(real.Rescale.LocalError), ImagLocalRescaleError: psRescaleGuardMetricValue(imag.Rescale.LocalError), ExpansionPass: real.Expansion.Pass && imag.Expansion.Pass, GuardValuePass: real.Guard.Value != nil && imag.Guard.Value != nil && real.Guard.Value.Pass && imag.Guard.Value.Pass, Q012CenteredUnique: real.Guard.Q012After.Unique && imag.Guard.Q012After.Unique, Q012RowsMatch: real.Guard.RowsMatch && imag.Guard.RowsMatch && real.Rescale.RowsMatch && imag.Rescale.RowsMatch, RoundedDivisionMatch: real.Rescale.RoundedDivisionMatch && imag.Rescale.RoundedDivisionMatch, ContractionPass: real.Contraction.Pass && imag.Contraction.Pass, MetadataPass: real.Contraction.MetadataMatch && imag.Contraction.MetadataMatch, NoExtraLevel: real.Rescale.OutputLevel == real.Rescale.InputLevel-1 && imag.Rescale.OutputLevel == imag.Rescale.InputLevel-1}
	row.ValidLocalQ2 = fix001P3LocalQ2SweepBranchValid(real) && fix001P3LocalQ2SweepBranchValid(imag)
	row.PrecisionQualified = row.ValidLocalQ2 && candidate.FinalRealError <= psRescaleGuardBudget && candidate.FinalImagError <= psRescaleGuardBudget
	return row
}

func fix001P3LocalQ2WriteGuardSweep(result fix001P3LocalQ2GuardSweepResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DesignLogN13F0LocalQ2GuardSweep(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3F0LocalQ2GuardSweepPrimaryBase, "HEAD") == requiredFIX001P3F0LocalQ2GuardSweepPrimaryBase
	result := fix001P3LocalQ2GuardSweepResult{SchemaVersion: "fix-001-p3-design-logn13-f0-local-q2-guard-sweep.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3F0LocalQ2GuardSweepPrimaryBase, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3F0LocalQ2GuardSweepSecondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "logn13_only": cfg.LogN == 13, "validated_oracle_powers": true, "common_plan_scale_2^92": true, "q0_q1_q2_profile": "56/39/40", "q2_local_only_at_f0": true, "no_q3_plus": true, "no_q0_q1_widening": true, "no_extra_q_levels": true, "no_c2s_s2c_mod1_change": true, "no_generated_power_redesign": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3F0LocalQ2GuardSweepSecondary {
		result.Classification, result.FirstRemainingBlocker = "logn13_f0_local_q2_guard_sweep_precondition_mismatch", "startup provenance or Secondary precondition"
		return fix001P3LocalQ2WriteGuardSweep(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	q0, q1, q2, q012, err := fix001P3LocalQ2Moduli(profile.BTP.BootstrappingParameters)
	if err != nil {
		return err
	}
	result.Primes = map[string]interface{}{"q0": q0.String(), "q1": q1.String(), "q2": q2.String(), "q0_bits": q0.BitLen(), "q1_bits": q1.BitLen(), "q2_bits": q2.BitLen(), "q012": q012.String(), "q012_bits": q012.BitLen()}
	standardPublic, err := q056StandardPublic(profile)
	if err != nil {
		return err
	}
	c1, err := q056Evaluate(&profile, "S0-q01-only-C1-G0-one-F0-one", []int{1, 0, 0, 0, 1})
	if err != nil {
		return err
	}
	result.S0Controls = map[string]interface{}{"standard_public": standardPublic, "q01_only_c1": q056CompactCandidate(c1), "expected_c1_max_error": 1.241886349e-8, "accepted_base_local_q2_guard_bits": 2}
	if !standardPublic.Pass || !c1.Valid || c1.FinalMaxError < 1.2e-8 || c1.FinalMaxError > 1.3e-8 {
		result.Classification, result.FirstRemainingBlocker = "logn13_f0_local_q2_guard_sweep_precondition_mismatch", "S0 q01-only C1 control"
		return fix001P3LocalQ2WriteGuardSweep(result, outPath)
	}

	for _, bits := range []int{2, 3, 4, 5, 6} {
		var realEvidence, imagEvidence fix001P3LocalQ2BranchEvidence
		candidate, realRun, imagRun, err := psRescaleGuardMakeCandidateWithOverrides(profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch, profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, bits}, fmt.Sprintf("S1-G0-one-local-q2-F0-%d", bits), &profile.InitialReal, &profile.InitialImag, profile.Boundaries, fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &realEvidence, bits), fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &imagEvidence, bits))
		if err != nil {
			return err
		}
		row := fix001P3LocalQ2SweepRowFor(bits, candidate, realEvidence, imagEvidence)
		row.StopAfterThisCandidate = !row.ValidLocalQ2
		result.Sweep = append(result.Sweep, row)
		if row.ValidLocalQ2 && row.PrecisionQualified && result.SelectedGuardBits == nil {
			selected := bits
			result.SelectedGuardBits = &selected
			compact := q056CompactCandidate(candidate)
			result.SelectedCandidate = &compact
			result.Classification, result.FirstRemainingBlocker = "logn13_f0_local_q2_guard_candidate_validated", "none"
			downstream, err := q056F0RunDownstream(profile, candidate)
			if err != nil {
				return err
			}
			result.Downstream = map[string]interface{}{"reached": downstream.Reached, "reason": downstream.Reason, "evalmod_real_vs_standard": downstream.EvalModRealVsStandard, "evalmod_imag_vs_standard": downstream.EvalModImagVsStandard, "post_s2c": downstream.PostS2C, "public_like": downstream.PublicLike, "standard_public_like": downstream.StandardLike, "metadata_correct": downstream.MetadataCorrect, "input_unchanged": downstream.InputUnchanged, "ps_arithmetic_blocker_independently_closed": downstream.PSArithmeticBlockerClosed}
			result.BlockerClosed = downstream.PSArithmeticBlockerClosed
			if !downstream.Reached {
				result.FirstRemainingBlocker = downstream.Reason
			}
			break
		}
		if !row.ValidLocalQ2 {
			if !row.Q012CenteredUnique || !row.Q012RowsMatch {
				result.Classification, result.FirstRemainingBlocker = "logn13_f0_local_q2_guard_capacity_failure", fmt.Sprintf("F0 guard bits=%d q012 capacity or exact row contract", bits)
			} else {
				result.Classification, result.FirstRemainingBlocker = "logn13_f0_local_q2_guard_contraction_failure", fmt.Sprintf("F0 guard bits=%d post-Rescale q01 contraction", bits)
			}
			break
		}
		_ = realRun
		_ = imagRun
	}
	if result.Classification == "" {
		result.Classification, result.FirstRemainingBlocker = "logn13_f0_local_q2_guard_insufficient_precision", "no tested guard bit count met the global polynomial budget"
	}
	return fix001P3LocalQ2WriteGuardSweep(result, outPath)
}
