//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"
)

const (
	requiredFIX001P3Q056F0TwoBitPrimaryBase = "29562ce5998c77a27584ed1f8251ab1a5b446c06"
	requiredFIX001P3Q056F0TwoBitSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
)

type q056F0GuardEvidence struct {
	GuardBits                  int     `json:"guard_bits"`
	CapacityRatio              float64 `json:"capacity_ratio"`
	GuardValuePreservation     float64 `json:"guard_value_preservation_error"`
	LocalRescaleError          float64 `json:"local_rescale_error"`
	PostRescaleCumulativeError float64 `json:"post_rescale_cumulative_error"`
	CenteredUnique             bool    `json:"centered_unique"`
	RowsMatch                  bool    `json:"fast_vs_full_rns_rows_match"`
	LevelUnchanged             bool    `json:"level_unchanged"`
	AlignmentSafe              bool    `json:"scale_alignment_safe"`
	Valid                      bool    `json:"valid"`
	FirstFailure               string  `json:"first_failure"`
}

type q056F0TwoBitDownstream struct {
	Reached                   bool            `json:"reached"`
	Reason                    string          `json:"reason,omitempty"`
	EvalModRealVsStandard     *PSGlobalMetric `json:"evalmod_real_vs_standard,omitempty"`
	EvalModImagVsStandard     *PSGlobalMetric `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2C                   *PSGlobalMetric `json:"post_s2c_semantic_error,omitempty"`
	PublicLike                *PSGlobalMetric `json:"public_like_max_component_error,omitempty"`
	StandardLike              *PSGlobalMetric `json:"standard_public_like,omitempty"`
	MetadataCorrect           bool            `json:"metadata_correct"`
	InputUnchanged            bool            `json:"input_unchanged"`
	PSArithmeticBlockerClosed bool            `json:"ps_arithmetic_blocker_independently_closed"`
}

type q056F0TwoBitResult struct {
	SchemaVersion      string                  `json:"schema_version"`
	Timestamp          time.Time               `json:"timestamp"`
	Primary            RepositoryMetadata      `json:"primary_repository"`
	Lattigo            RepositoryMetadata      `json:"lattigo_repository"`
	Environment        EnvironmentMetadata     `json:"environment"`
	Config             BootstrapConfig         `json:"config"`
	Parameters         ExperimentParameters    `json:"effective_parameters"`
	Workload           CorrectnessWorkload     `json:"workload"`
	PolynomialBudget   float64                 `json:"polynomial_budget"`
	PublicThreshold    float64                 `json:"public_like_threshold"`
	T0Control          map[string]interface{}  `json:"t0_control"`
	F0Guards           []q056F0GuardEvidence   `json:"f0_guard_0_1_2_bit_table"`
	Candidates         []q056CandidateEvidence `json:"candidates_c0_c3"`
	SelectedC3         *q056CandidateEvidence  `json:"selected_c3,omitempty"`
	Downstream         q056F0TwoBitDownstream  `json:"downstream"`
	Classification     string                  `json:"classification"`
	FirstBlocker       string                  `json:"first_remaining_blocker"`
	PSArithmeticClosed bool                    `json:"ps_arithmetic_blocker_independently_closed"`
	Validation         map[string]interface{}  `json:"validation"`
}

func q056F0Boundary(candidate psRescaleGuardCandidate) *psRescaleGuardBoundary {
	return psRescaleGuardBoundaryByID(candidate.BoundaryEffects, "F0-final-rescale")
}

func q056F0Evidence(candidate psRescaleGuardCandidate, guardBits int) q056F0GuardEvidence {
	boundary := q056F0Boundary(candidate)
	evidence := q056F0GuardEvidence{GuardBits: guardBits, AlignmentSafe: candidate.AlignmentSafe, Valid: candidate.Valid, FirstFailure: candidate.FirstFailure}
	if boundary == nil {
		evidence.FirstFailure = "F0-final-rescale.missing"
		evidence.Valid = false
		return evidence
	}
	evidence.CapacityRatio = boundary.GuardedPreCapacityRatio
	evidence.GuardValuePreservation = psRescaleGuardMetricValue(boundary.GuardValuePreservation)
	evidence.LocalRescaleError = psRescaleGuardMetricValue(boundary.LocalRescaleError)
	evidence.PostRescaleCumulativeError = psRescaleGuardMetricValue(boundary.PostCumulativeError)
	evidence.CenteredUnique = boundary.CenteredUnique
	evidence.RowsMatch = boundary.RowsMatch
	evidence.LevelUnchanged = boundary.LevelUnchanged
	if !evidence.CenteredUnique || !evidence.RowsMatch {
		evidence.Valid = false
	}
	return evidence
}

func q056F0RunDownstream(profile q056PreparedProfile, candidate psRescaleGuardCandidate) (q056F0TwoBitDownstream, error) {
	if candidate.realFinal == nil || candidate.imagFinal == nil {
		return q056F0TwoBitDownstream{Reason: "C3 final ciphertext unavailable"}, nil
	}
	planScale := precisionSweepScale(92)
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return q056F0TwoBitDownstream{}, fmt.Errorf("real normalized DoubleAngle: %w", err)
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return q056F0TwoBitDownstream{}, fmt.Errorf("imag normalized DoubleAngle: %w", err)
	}
	if realPath.FastFinal == nil || imagPath.FastFinal == nil {
		return q056F0TwoBitDownstream{Reason: fmt.Sprintf("normalized path stopped: real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)}, nil
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return q056F0TwoBitDownstream{}, fmt.Errorf("real final evidence: %w", err)
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return q056F0TwoBitDownstream{}, fmt.Errorf("imag final evidence: %w", err)
	}
	publicLike, standardLike, postS2C, metadata, unchanged, err := precisionSweepPublicLike(realPath, imagPath, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return q056F0TwoBitDownstream{}, fmt.Errorf("public-like path: %w", err)
	}
	return q056F0TwoBitDownstream{
		Reached: true, EvalModRealVsStandard: realFinal.FastVsStandard, EvalModImagVsStandard: imagFinal.FastVsStandard,
		PostS2C: postS2C, PublicLike: publicLike, StandardLike: standardLike, MetadataCorrect: metadata,
		InputUnchanged: unchanged, PSArithmeticBlockerClosed: publicLike != nil && publicLike.Pass,
	}, nil
}

func q056F0Write(result q056F0TwoBitResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if outPath == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DesignLogN13Q056F0TwoBitGuard(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := q056F0TwoBitResult{
		SchemaVersion: "fix-001-p3-design-logn13-q0-56-f0-two-bit-guard.v1", Timestamp: time.Now().UTC(),
		Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, PolynomialBudget: psRescaleGuardBudget, PublicThreshold: q056PublicThreshold,
		Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: 1 << (cfg.LogN - 1)},
		Validation: map[string]interface{}{
			"primary_required_base": requiredFIX001P3Q056F0TwoBitPrimaryBase, "secondary_commit": secondaryCommit,
			"secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3Q056F0TwoBitSecondary,
			"no_secondary_production_changes": true, "logn13_only": cfg.LogN == 13, "q0_56_q1_39_only": true,
			"no_q0_beyond_56": true, "no_q1_widening": true, "no_q2_plus_arithmetic": true, "no_generated_power_redesign": true,
			"common_plan_scale_2^92": true, "same_level_topology": true, "no_c2s_change": true, "no_s2c_change": true,
			"no_mod1_change": true, "no_extra_q_levels": true, "no_production_integration": true,
			"no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true,
		},
	}
	if cfg.LogN != 13 || !secondaryClean || secondaryCommit != requiredFIX001P3Q056F0TwoBitSecondary {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_precondition_mismatch", "startup_precondition"
		return q056F0Write(result, outPath)
	}
	profileCfg := q056BuildConfig(cfg, 56)
	profile, err := q056PrepareProfile(profileCfg)
	if err != nil {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_precondition_mismatch", err.Error()
		return q056F0Write(result, outPath)
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	standardPublic, err := q056StandardPublic(profile)
	if err != nil {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_precondition_mismatch", fmt.Sprintf("Standard public: %v", err)
		return q056F0Write(result, outPath)
	}
	zero := make([]int, len(profile.Boundaries))
	if len(zero) < 5 {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_precondition_mismatch", "unexpected boundary count"
		return q056F0Write(result, outPath)
	}
	controlSchedule := []int{1, 0, 0, 0, 1}
	c0, err := q056Evaluate(&profile, "C0-no-guards", zero)
	if err != nil {
		return err
	}
	c1, err := q056Evaluate(&profile, "C1-G0-one-F0-one", controlSchedule)
	if err != nil {
		return err
	}
	c2Schedule := []int{0, 0, 0, 0, 2}
	c2, err := q056Evaluate(&profile, "C2-F0-two-only", c2Schedule)
	if err != nil {
		return err
	}
	c3Schedule := []int{1, 0, 0, 0, 2}
	c3, err := q056Evaluate(&profile, "C3-G0-one-F0-two", c3Schedule)
	if err != nil {
		return err
	}
	result.T0Control = map[string]interface{}{
		"standard_public": standardPublic, "c1_real_error": c1.FinalRealError, "c1_imag_error": c1.FinalImagError,
		"c1_max_error": c1.FinalMaxError, "c1_expected_real": 1.185708509e-8, "c1_expected_imag": 1.241886349e-8,
		"c1_expected_max": 1.241886349e-8, "c1_contracts_pass": c1.Valid,
	}
	for _, guardBits := range []int{0, 1, 2} {
		schedule := append([]int(nil), zero...)
		schedule[len(schedule)-1] = guardBits
		candidate, err := q056Evaluate(&profile, fmt.Sprintf("F0-%d-bit", guardBits), schedule)
		if err != nil {
			return err
		}
		result.F0Guards = append(result.F0Guards, q056F0Evidence(candidate, guardBits))
	}
	result.Candidates = []q056CandidateEvidence{q056CompactCandidate(c0), q056CompactCandidate(c1), q056CompactCandidate(c2), q056CompactCandidate(c3)}
	for _, candidate := range result.Candidates {
		if candidate.Name == c3.Name {
			copyCandidate := candidate
			result.SelectedC3 = &copyCandidate
		}
	}
	precondition := standardPublic.Pass && c1.Valid && math.Abs(c1.FinalRealError-1.185708509e-8) < 1e-10 && math.Abs(c1.FinalImagError-1.241886349e-8) < 1e-10 && math.Abs(c1.FinalMaxError-1.241886349e-8) < 1e-10
	result.T0Control["precondition_pass"] = precondition
	if !precondition {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_precondition_mismatch", "T0 control"
		return q056F0Write(result, outPath)
	}
	if !c3.CapacitySafe || !c3.ValuePreserving {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_guard_capacity_failure", c3.FirstFailure
		return q056F0Write(result, outPath)
	}
	if !c3.AlignmentSafe {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_guard_alignment_failure", c3.FirstFailure
		return q056F0Write(result, outPath)
	}
	if !c3.PrecisionQualified {
		result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_guard_insufficient_precision", "C3 polynomial budget"
		return q056F0Write(result, outPath)
	}
	downstream, err := q056F0RunDownstream(profile, c3)
	if err != nil {
		return err
	}
	result.Downstream = downstream
	result.PSArithmeticClosed = downstream.PSArithmeticBlockerClosed
	result.Classification, result.FirstBlocker = "logn13_q0_56_f0_two_bit_guard_candidate_validated", "none"
	if !downstream.Reached {
		result.FirstBlocker = downstream.Reason
	}
	return q056F0Write(result, outPath)
}
