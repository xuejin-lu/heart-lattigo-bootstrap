//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3G0LocalQ2GuardSweepPrimaryBase = "07e4e8e1c4a9fca8c0f4504f4fda4099a3d883de"
	requiredFIX001P3G0LocalQ2GuardSweepSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	g0LocalQ2PublicThreshold                       = 1e-2
)

type fix001P3G0LocalQ2GuardSweepCandidate struct {
	GuardBits        int                           `json:"guard_bits"`
	Candidate        q056CandidateEvidence         `json:"candidate"`
	RealEvidence     fix001P3LocalQ2BranchEvidence `json:"real_g0_local_q2"`
	ImagEvidence     fix001P3LocalQ2BranchEvidence `json:"imag_g0_local_q2"`
	ValidLocalQ2     bool                          `json:"valid_local_q2"`
	Downstream       map[string]interface{}        `json:"downstream,omitempty"`
	PublicLikePass   bool                          `json:"public_like_pass"`
	SystemSufficient bool                          `json:"system_sufficient"`
	FirstBlocker     string                        `json:"first_blocker"`
}

type fix001P3G0LocalQ2GuardSweepResult struct {
	SchemaVersion         string                                 `json:"schema_version"`
	Timestamp             time.Time                              `json:"timestamp"`
	Primary               RepositoryMetadata                     `json:"primary_repository"`
	Lattigo               RepositoryMetadata                     `json:"lattigo_repository"`
	Environment           EnvironmentMetadata                    `json:"environment"`
	Config                BootstrapConfig                        `json:"config"`
	Parameters            ExperimentParameters                   `json:"effective_parameters"`
	Primes                map[string]interface{}                 `json:"q0_q1_q2_primes"`
	Controls              map[string]interface{}                 `json:"controls"`
	Candidates            []fix001P3G0LocalQ2GuardSweepCandidate `json:"candidates"`
	SelectedGuardBits     *int                                   `json:"selected_smallest_qualifying_guard_bits,omitempty"`
	SelectedCandidate     *q056CandidateEvidence                 `json:"selected_candidate,omitempty"`
	Operations            map[string]interface{}                 `json:"static_q2_costs"`
	Classification        string                                 `json:"classification"`
	SystemSufficient      bool                                   `json:"ps_da_arithmetic_system_sufficient"`
	FirstRemainingBlocker string                                 `json:"first_remaining_blocker"`
	Validation            map[string]interface{}                 `json:"validation"`
}

func fix001P3G0Write(result fix001P3G0LocalQ2GuardSweepResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3G0Operations() map[string]interface{} {
	return map[string]interface{}{
		"g0_bounded_region": map[string]interface{}{
			"coefficient_transforms":    2,
			"q2_ntt":                    2,
			"q2_intt":                   0,
			"q2_limb_arithmetic_passes": 4,
			"q2_rescale_boundaries":     2,
		},
		"f0_fixed_region": map[string]interface{}{
			"coefficient_transforms":    2,
			"q2_ntt":                    2,
			"q2_intt":                   0,
			"q2_limb_arithmetic_passes": 3,
			"q2_rescale_boundaries":     1,
		},
		"all_rounds_double_angle": map[string]interface{}{
			"rounds":                    3,
			"coefficient_transforms":    6,
			"q2_ntt":                    6,
			"q2_intt":                   0,
			"q2_limb_arithmetic_passes": 18,
			"q2_rescale_boundaries":     6,
		},
		"total_fixed_plus_sweep_path": map[string]interface{}{
			"coefficient_transforms":    10,
			"q2_ntt":                    10,
			"q2_intt":                   0,
			"q2_limb_arithmetic_passes": 25,
			"q2_rescale_boundaries":     9,
		},
		"g0_candidate_guard_bits": []int{2, 3, 4},
	}
}

func fix001P3G0RunDownstream(profile q056PreparedProfile, candidate psRescaleGuardCandidate) (map[string]interface{}, error) {
	planScale := precisionSweepScale(92)
	realState, err := fix001P3DAR0PrepareState(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return nil, err
	}
	imagState, err := fix001P3DAR0PrepareState(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return nil, err
	}
	perRound := make([]map[string]interface{}, 0, profile.Fast.Mod1Parameters.DoubleAngle)
	for round := 0; round < profile.Fast.Mod1Parameters.DoubleAngle; round++ {
		realOutcome, err := fix001P3DAAllRoundsRound(&realState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return nil, err
		}
		imagOutcome, err := fix001P3DAAllRoundsRound(&imagState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return nil, err
		}
		perRound = append(perRound, map[string]interface{}{"round": round, "real": realOutcome.Evidence, "imag": imagOutcome.Evidence})
		for _, outcome := range []fix001P3DAAllRoundsRoundOutcome{realOutcome, imagOutcome} {
			if outcome.Classification != "" {
				return map[string]interface{}{"reached": false, "per_round": perRound, "classification": outcome.Classification, "first_blocker": outcome.Blocker}, nil
			}
		}
	}
	if err := fix001P3DAAllRoundsFinalize(&realState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return nil, err
	}
	if err := fix001P3DAAllRoundsFinalize(&imagState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return nil, err
	}
	if realState.Path.FastFinal == nil || imagState.Path.FastFinal == nil || !realState.Path.FastRowsMatch || !realState.Path.RestoreRowsMatch || !imagState.Path.FastRowsMatch || !imagState.Path.RestoreRowsMatch {
		return map[string]interface{}{"reached": false, "per_round": perRound, "classification": "logn13_g0_local_q2_guard_downstream_insufficient", "first_blocker": "final.restore_row_or_metadata_contract"}, nil
	}
	realFinal, err := evalModMatchedFinalEvidence(realState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return nil, err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return nil, err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realState.Path, imagState.Path, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return nil, err
	}
	publicPass := publicEvidence.PublicLike != nil && publicEvidence.PublicLike.Pass && publicEvidence.PublicLike.MaxComponent <= g0LocalQ2PublicThreshold
	metadataPass, _ := publicEvidence.Metadata["pass"].(bool)
	systemPass := publicPass && metadataPass && publicEvidence.InputUnchanged && realFinal.FastVsStandard != nil && imagFinal.FastVsStandard != nil && realFinal.FastVsStandard.Pass && imagFinal.FastVsStandard.Pass
	return map[string]interface{}{
		"reached":                                   true,
		"per_round":                                 perRound,
		"evalmod_real_vs_standard":                  realFinal.FastVsStandard,
		"evalmod_imag_vs_standard":                  imagFinal.FastVsStandard,
		"post_s2c_fast_vs_standard":                 publicEvidence.PostS2C,
		"unpack_finalization_fast_vs_standard_core": publicEvidence.FinalVsStdCore,
		"standard_like":                             publicEvidence.StandardLike,
		"public_like":                               publicEvidence.PublicLike,
		"final_metadata":                            publicEvidence.Metadata,
		"input_unchanged":                           publicEvidence.InputUnchanged,
		"public_like_threshold":                     g0LocalQ2PublicThreshold,
		"public_like_pass":                          publicPass,
		"system_sufficient":                         systemPass,
		"first_blocker": func() string {
			if systemPass {
				return "none"
			}
			return "public-like threshold or downstream semantic/metadata contract"
		}(),
	}, nil
}

func fix001P3G0RunCandidate(profile q056PreparedProfile, bits int) (fix001P3G0LocalQ2GuardSweepCandidate, psRescaleGuardCandidate, error) {
	var f0Real, f0Imag, g0Real, g0Imag fix001P3LocalQ2BranchEvidence
	g0RealOverride := fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &g0Real, bits)
	g0ImagOverride := fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &g0Imag, bits)
	realBoundaryOverride := func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error) {
		if boundary.ID != "G0-rescale" {
			return false, nil
		}
		return g0RealOverride(params, eval, ct, boundary)
	}
	imagBoundaryOverride := func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error) {
		if boundary.ID != "G0-rescale" {
			return false, nil
		}
		return g0ImagOverride(params, eval, ct, boundary)
	}
	candidate, _, _, err := psRescaleGuardMakeCandidateWithAllOverrides(
		profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch,
		profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, 3},
		fmt.Sprintf("G0-local-q2-%d-F0-local-q2-3", bits), &profile.InitialReal, &profile.InitialImag,
		profile.Boundaries,
		fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &f0Real, 3),
		fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &f0Imag, 3),
		psRescaleGuardBoundaryOverride(realBoundaryOverride),
		psRescaleGuardBoundaryOverride(imagBoundaryOverride),
	)
	if err != nil {
		return fix001P3G0LocalQ2GuardSweepCandidate{}, psRescaleGuardCandidate{}, err
	}
	row := fix001P3G0LocalQ2GuardSweepCandidate{GuardBits: bits, Candidate: q056CompactCandidate(candidate), RealEvidence: g0Real, ImagEvidence: g0Imag, ValidLocalQ2: fix001P3LocalQ2SweepBranchValid(g0Real) && fix001P3LocalQ2SweepBranchValid(g0Imag), FirstBlocker: candidate.FirstFailure}
	if !row.ValidLocalQ2 {
		row.FirstBlocker = fmt.Sprintf("G0 guard bits=%d local q2 expansion/capacity/contraction contract", bits)
		return row, candidate, nil
	}
	downstream, err := fix001P3G0RunDownstream(profile, candidate)
	if err != nil {
		return row, candidate, err
	}
	row.Downstream = downstream
	if reached, _ := downstream["reached"].(bool); reached {
		row.PublicLikePass, _ = downstream["public_like_pass"].(bool)
		row.SystemSufficient, _ = downstream["system_sufficient"].(bool)
	}
	if blocker, ok := downstream["first_blocker"].(string); ok && blocker != "" {
		row.FirstBlocker = blocker
	}
	return row, candidate, nil
}

func fix001P3G0ControlValid(row fix001P3G0LocalQ2GuardSweepCandidate) bool {
	if !row.ValidLocalQ2 || row.Downstream == nil {
		return false
	}
	if reached, _ := row.Downstream["reached"].(bool); !reached {
		return false
	}
	realMetric, realOK := row.Downstream["evalmod_real_vs_standard"].(*PSGlobalMetric)
	imagMetric, imagOK := row.Downstream["evalmod_imag_vs_standard"].(*PSGlobalMetric)
	metadata, metadataOK := row.Downstream["final_metadata"].(map[string]interface{})
	inputUnchanged, inputOK := row.Downstream["input_unchanged"].(bool)
	metadataPass, passOK := metadata["pass"].(bool)
	return realOK && imagOK && realMetric.Pass && imagMetric.Pass && metadataOK && passOK && metadataPass && inputOK && inputUnchanged
}

func runFIX001P3DesignLogN13G0LocalQ2GuardSweep(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3G0LocalQ2GuardSweepPrimaryBase, "HEAD") == requiredFIX001P3G0LocalQ2GuardSweepPrimaryBase
	result := fix001P3G0LocalQ2GuardSweepResult{
		SchemaVersion: "fix-001-p3-design-logn13-g0-local-q2-guard-sweep.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, Operations: fix001P3G0Operations(),
		Validation: map[string]interface{}{
			"primary_required_base":                  requiredFIX001P3G0LocalQ2GuardSweepPrimaryBase,
			"primary_required_base_ancestor":         primaryBaseAncestor,
			"secondary_commit":                       secondaryCommit,
			"secondary_exact_required_commit":        secondaryCommit == requiredFIX001P3G0LocalQ2GuardSweepSecondary,
			"secondary_clean":                        secondaryClean,
			"no_secondary_production_changes":        true,
			"logn13_only":                            cfg.LogN == 13,
			"q0_q1_q2_profile":                       "56/39/40",
			"plan_scale_2^92":                        true,
			"f0_local_q2_guard_bits":                 3,
			"g0_sweep_guard_bits":                    []int{2, 3, 4},
			"all_rounds_double_angle_local_q2_fixed": true,
			"no_q3_plus":                             true, "no_q0_q1_widening": true, "no_extra_q_levels": true,
			"no_ps_retuning": true, "no_generated_power_redesign": true,
			"no_c2s_s2c_mod1_change": true, "no_production_integration": true,
			"no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true,
		},
	}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3G0LocalQ2GuardSweepSecondary {
		result.Classification, result.FirstRemainingBlocker = "logn13_g0_local_q2_guard_sweep_precondition_mismatch", "startup provenance or Secondary precondition"
		return fix001P3G0Write(result, outPath)
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
	control, err := q056Evaluate(&profile, "G0-control-q01-only", []int{1, 0, 0, 0, 1})
	if err != nil {
		return err
	}
	result.Controls = map[string]interface{}{"standard_public": standardPublic, "q01_only_c1": q056CompactCandidate(control), "q01_only_control_pass": standardPublic.Pass && control.Valid}
	if !standardPublic.Pass || !control.Valid {
		result.Classification, result.FirstRemainingBlocker = "logn13_g0_local_q2_guard_sweep_precondition_mismatch", "q01-only C1 control"
		return fix001P3G0Write(result, outPath)
	}
	controlRow, controlCandidate, err := fix001P3G0RunCandidate(profile, 1)
	if err != nil {
		return err
	}
	result.Controls["g0_one_bit_local_q2_control"] = controlRow
	if !fix001P3G0ControlValid(controlRow) {
		result.Classification, result.FirstRemainingBlocker = "logn13_g0_local_q2_guard_sweep_precondition_mismatch", "G0 one-bit local-q2 downstream control"
		return fix001P3G0Write(result, outPath)
	}
	_ = controlCandidate
	for _, bits := range []int{2, 3, 4} {
		row, _, err := fix001P3G0RunCandidate(profile, bits)
		if err != nil {
			return err
		}
		result.Candidates = append(result.Candidates, row)
		if row.SystemSufficient && result.SelectedGuardBits == nil {
			selected := bits
			result.SelectedGuardBits = &selected
			selectedCandidate := row.Candidate
			result.SelectedCandidate = &selectedCandidate
		}
	}
	for _, row := range result.Candidates {
		if row.SystemSufficient {
			result.SystemSufficient = true
			result.Classification = "logn13_g0_local_q2_guard_downstream_validated"
			result.FirstRemainingBlocker = "none"
			break
		}
	}
	if !result.SystemSufficient {
		allCapacity := true
		anyContraction := false
		for _, row := range result.Candidates {
			if row.ValidLocalQ2 {
				allCapacity = false
			}
			if !row.RealEvidence.Contraction.Pass || !row.ImagEvidence.Contraction.Pass {
				anyContraction = true
			}
		}
		switch {
		case allCapacity:
			result.Classification = "logn13_g0_local_q2_guard_capacity_failure"
		case anyContraction:
			result.Classification = "logn13_g0_local_q2_guard_contraction_failure"
		default:
			result.Classification = "logn13_g0_local_q2_guard_insufficient_downstream_precision"
		}
		for _, row := range result.Candidates {
			if row.FirstBlocker != "" && row.FirstBlocker != "none" {
				result.FirstRemainingBlocker = row.FirstBlocker
				break
			}
		}
		if result.FirstRemainingBlocker == "" {
			result.FirstRemainingBlocker = "no G0 guard candidate met the downstream public-like threshold"
		}
	}
	return fix001P3G0Write(result, outPath)
}
