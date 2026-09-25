//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3Q056PrimaryBase = "d9d540a306f09298dbcb1a4960bd2900bd7a3a53"
	requiredFIX001P3Q056Secondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	q056PublicThreshold             = 1e-2
)

type q056ControlEvidence struct {
	NoGuardMaxError       float64               `json:"no_guard_max_error"`
	G0CapacityRatio       float64               `json:"g0_unguarded_capacity_ratio"`
	G0OneBit              q056CandidateEvidence `json:"g0_one_bit_guard"`
	F0OneBit              q056CandidateEvidence `json:"f0_one_bit_guard"`
	NoGuardExpectedRange  [2]float64            `json:"no_guard_expected_range"`
	F0ExpectedMaxError    float64               `json:"f0_expected_max_error"`
	ControlPreconditionOK bool                  `json:"precondition_pass"`
}

type q056CandidateEvidence struct {
	Name                  string   `json:"name"`
	BoundaryGuardBits     []int    `json:"boundary_guard_bits"`
	GuardedBoundaries     []string `json:"guarded_boundaries"`
	TotalGuardBits        int      `json:"total_guard_bits"`
	GuardedBoundaryCount  int      `json:"guarded_boundary_count"`
	FinalRealError        float64  `json:"final_real_error"`
	FinalImagError        float64  `json:"final_imag_error"`
	FinalMaxError         float64  `json:"final_max_error"`
	StandardRealError     float64  `json:"final_real_vs_standard"`
	StandardImagError     float64  `json:"final_imag_vs_standard"`
	CapacitySafe          bool     `json:"capacity_safe"`
	AlignmentSafe         bool     `json:"scale_alignment_safe"`
	ValuePreserving       bool     `json:"guard_value_preserving"`
	LevelSafe             bool     `json:"no_extra_level_consumption"`
	RowsMatch             bool     `json:"fast_vs_full_rns_rows_match"`
	Valid                 bool     `json:"valid"`
	PrecisionQualified    bool     `json:"precision_qualified"`
	FirstFailure          string   `json:"first_failure"`
	FirstLimitingBoundary string   `json:"first_limiting_boundary"`
}

type q056BoundaryEvidence struct {
	ID                                string  `json:"id"`
	Order                             int     `json:"order"`
	UnguardedPreCapacityRatio         float64 `json:"unguarded_pre_rescale_capacity_ratio"`
	OneBitGuardProjectedCapacityRatio float64 `json:"one_bit_guard_projected_capacity_ratio"`
	OneBitGuardActualCapacityRatio    float64 `json:"one_bit_guard_actual_capacity_ratio"`
	OneBitGuardValuePreservationError float64 `json:"one_bit_guard_value_preservation_error"`
	LocalRescaleErrorBefore           float64 `json:"local_rescale_error_before"`
	LocalRescaleErrorAfter            float64 `json:"local_rescale_error_after"`
	CenteredUnique                    bool    `json:"centered_unique"`
	RowsMatch                         bool    `json:"fast_vs_full_rns_rows_match"`
	LevelUnchanged                    bool    `json:"level_unchanged"`
	OneBitValid                       bool    `json:"one_bit_guard_valid"`
	FirstFailure                      string  `json:"first_failure"`
}

type q056ProfileEvidence struct {
	RequestedQ0Bits         int                        `json:"requested_q0_bits"`
	Q0                      string                     `json:"q0_prime"`
	Q1                      string                     `json:"q1_prime"`
	Q0Bits                  int                        `json:"q0_actual_bit_length"`
	Q1Bits                  int                        `json:"q1_actual_bit_length"`
	Q01                     string                     `json:"q01_product"`
	Q01Bits                 int                        `json:"q01_product_bit_length"`
	Q01WithinExistingDomain bool                       `json:"q01_within_existing_domain"`
	FastRescaleAccepted     bool                       `json:"fast_rescale_domain_accepted"`
	StandardPublic          *PSGlobalMetric            `json:"standard_public_result"`
	Boundaries              []q056BoundaryEvidence     `json:"five_boundary_table"`
	Candidates              []q056CandidateEvidence    `json:"candidates"`
	GreedySteps             []psRescaleGuardGreedyStep `json:"accepted_greedy_steps,omitempty"`
	BestValid               *q056CandidateEvidence     `json:"best_valid_candidate,omitempty"`
	Selected                *q056CandidateEvidence     `json:"selected_candidate,omitempty"`
	FirstBlocker            string                     `json:"first_remaining_blocker"`
}

type q056Result struct {
	SchemaVersion    string                 `json:"schema_version"`
	Timestamp        time.Time              `json:"timestamp"`
	Primary          RepositoryMetadata     `json:"primary_repository"`
	Lattigo          RepositoryMetadata     `json:"lattigo_repository"`
	Environment      EnvironmentMetadata    `json:"environment"`
	Config           BootstrapConfig        `json:"config"`
	Parameters       ExperimentParameters   `json:"effective_parameters"`
	Workload         CorrectnessWorkload    `json:"workload"`
	PolynomialBudget float64                `json:"polynomial_budget"`
	PublicThreshold  float64                `json:"public_like_threshold"`
	Control          q056ControlEvidence    `json:"q0_55_39_control"`
	Widened          q056ProfileEvidence    `json:"q0_56_39_profile"`
	Classification   string                 `json:"classification"`
	FirstBlocker     string                 `json:"first_remaining_blocker"`
	Downstream       map[string]interface{} `json:"downstream,omitempty"`
	Validation       map[string]interface{} `json:"validation"`
}

type q056PreparedProfile struct {
	Residual     ckks.Parameters
	BTP          bootstrapping.Parameters
	Fast         *bootstrapping.FastEvaluator
	Standard     *bootstrapping.Evaluator
	StandardSK   *rlwe.SecretKey
	Inputs       evalModMatchedInputs
	RealBranch   psRescaleGuardBranch
	ImagBranch   psRescaleGuardBranch
	RealStandard []complex128
	ImagStandard []complex128
	InitialReal  psRescaleGuardBranchRun
	InitialImag  psRescaleGuardBranchRun
	Boundaries   []psRescaleGuardBoundary
	Candidates   map[string]psRescaleGuardCandidate
	Runs         map[string][2]psRescaleGuardBranchRun
}

func q056ProfilePrimes(params ckks.Parameters) (string, string, int, int, string, int, error) {
	if len(params.RingQ().SubRings) < 2 {
		return "", "", 0, 0, "", 0, fmt.Errorf("profile has fewer than two Q primes")
	}
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	q01 := new(big.Int).Mul(new(big.Int).Set(q0), q1)
	return q0.String(), q1.String(), q0.BitLen(), q1.BitLen(), q01.String(), q01.BitLen(), nil
}

func q056BuildConfig(cfg BootstrapConfig, q0Bits int) BootstrapConfig {
	copyCfg := cfg
	copyCfg.Q0 = []int{q0Bits}
	return copyCfg
}

func q056PrepareProfile(cfg BootstrapConfig) (q056PreparedProfile, error) {
	profile := q056PreparedProfile{Candidates: map[string]psRescaleGuardCandidate{}, Runs: map[string][2]psRescaleGuardBranchRun{}}
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return profile, err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return profile, err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, standardSK, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return profile, err
	}
	profile.Residual, profile.BTP, profile.Fast, profile.Standard, profile.StandardSK = residual, btp, fastEval, standardEval, standardSK
	profile.Inputs, err = evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, standardSK)
	if err != nil {
		return profile, fmt.Errorf("matched C2S inputs: %w", err)
	}
	realBase, err := polynomialDecompositionPrepareBranch("real", btp.BootstrappingParameters, btp, fastEval, standardEval, profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, standardSK)
	if err != nil {
		return profile, fmt.Errorf("real branch: %w", err)
	}
	imagBase, err := polynomialDecompositionPrepareBranch("imag", btp.BootstrappingParameters, btp, fastEval, standardEval, profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, standardSK)
	if err != nil {
		return profile, fmt.Errorf("imag branch: %w", err)
	}
	profile.RealBranch, profile.ImagBranch = psRescaleGuardBranch{Name: "real", Base: realBase}, psRescaleGuardBranch{Name: "imag", Base: imagBase}
	profile.RealStandard, err = psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, standardSK, realBase.InputValues)
	if err != nil {
		return profile, fmt.Errorf("real Standard polynomial: %w", err)
	}
	profile.ImagStandard, err = psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, standardSK, imagBase.InputValues)
	if err != nil {
		return profile, fmt.Errorf("imag Standard polynomial: %w", err)
	}
	profile.RealBranch.Standard, profile.ImagBranch.Standard = profile.RealStandard, profile.ImagStandard
	profile.InitialReal, err = psRescaleGuardRunBranch(btp.BootstrappingParameters, fastEval, profile.RealBranch, nil, nil)
	if err != nil {
		return profile, fmt.Errorf("real control replay: %w", err)
	}
	profile.InitialImag, err = psRescaleGuardRunBranch(btp.BootstrappingParameters, fastEval, profile.ImagBranch, nil, nil)
	if err != nil {
		return profile, fmt.Errorf("imag control replay: %w", err)
	}
	profile.Boundaries = psRescaleGuardBoundaryEffects(profile.InitialReal.Replay, profile.InitialImag.Replay)
	return profile, nil
}

func q056Evaluate(profile *q056PreparedProfile, name string, schedule []int) (psRescaleGuardCandidate, error) {
	key := psRescaleGuardScheduleKey(schedule)
	if candidate, ok := profile.Candidates[key]; ok {
		return candidate, nil
	}
	candidate, realRun, imagRun, err := psRescaleGuardMakeCandidate(profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch, profile.RealStandard, profile.ImagStandard, schedule, name, &profile.InitialReal, &profile.InitialImag, profile.Boundaries)
	if err != nil {
		return psRescaleGuardCandidate{}, err
	}
	profile.Candidates[key] = candidate
	profile.Runs[key] = [2]psRescaleGuardBranchRun{realRun, imagRun}
	return candidate, nil
}

func q056CandidateByName(candidates []psRescaleGuardCandidate, name string) *psRescaleGuardCandidate {
	for i := range candidates {
		if candidates[i].Name == name {
			return &candidates[i]
		}
	}
	return nil
}

func q056StandardPublic(profile q056PreparedProfile) (*PSGlobalMetric, error) {
	output, err := profile.Standard.Bootstrap(reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return nil, err
	}
	values, err := decodeWithSecret(profile.Residual, output, profile.StandardSK)
	if err != nil {
		return nil, err
	}
	return evalModMatchedMetricFor(reproducibleValues(profile.Residual.MaxSlots()), values, q056PublicThreshold), nil
}

func q056CompactCandidate(candidate psRescaleGuardCandidate) q056CandidateEvidence {
	return q056CandidateEvidence{
		Name: candidate.Name, BoundaryGuardBits: append([]int(nil), candidate.BoundaryGuardBits...), GuardedBoundaries: append([]string(nil), candidate.GuardedBoundaries...),
		TotalGuardBits: candidate.TotalGuardBits, GuardedBoundaryCount: candidate.GuardedBoundaryCount,
		FinalRealError: candidate.FinalRealError, FinalImagError: candidate.FinalImagError, FinalMaxError: candidate.FinalMaxError,
		StandardRealError: candidate.StandardRealError, StandardImagError: candidate.StandardImagError,
		CapacitySafe: candidate.CapacitySafe, AlignmentSafe: candidate.AlignmentSafe, ValuePreserving: candidate.ValuePreserving,
		LevelSafe: candidate.LevelSafe, RowsMatch: candidate.RowsMatch, Valid: candidate.Valid, PrecisionQualified: candidate.PrecisionQualified,
		FirstFailure: candidate.FirstFailure, FirstLimitingBoundary: candidate.FirstLimitingBoundary,
	}
}

func q056BoundaryMetric(boundaries []psRescaleGuardBoundary, id string) *psRescaleGuardBoundary {
	return psRescaleGuardBoundaryByID(boundaries, id)
}

func q056BoundaryEvidenceFor(base, single psRescaleGuardCandidate, id string, order int) q056BoundaryEvidence {
	baseBoundary := q056BoundaryMetric(base.BoundaryEffects, id)
	singleBoundary := q056BoundaryMetric(single.BoundaryEffects, id)
	evidence := q056BoundaryEvidence{ID: id, Order: order, FirstFailure: single.FirstFailure}
	if baseBoundary != nil {
		evidence.UnguardedPreCapacityRatio = baseBoundary.PreCapacityRatio
		evidence.LocalRescaleErrorBefore = psRescaleGuardMetricValue(baseBoundary.LocalRescaleError)
	}
	if singleBoundary != nil {
		evidence.OneBitGuardProjectedCapacityRatio = singleBoundary.GuardedPreCapacityRatio
		evidence.OneBitGuardActualCapacityRatio = singleBoundary.GuardedPreCapacityRatio
		evidence.OneBitGuardValuePreservationError = psRescaleGuardMetricValue(singleBoundary.GuardValuePreservation)
		evidence.LocalRescaleErrorAfter = psRescaleGuardMetricValue(singleBoundary.LocalRescaleError)
		evidence.CenteredUnique = singleBoundary.CenteredUnique
		evidence.RowsMatch = singleBoundary.RowsMatch
		evidence.LevelUnchanged = singleBoundary.LevelUnchanged
	}
	evidence.OneBitValid = single.Valid
	return evidence
}

func q056BuildControlEvidence(profile q056PreparedProfile) (q056ControlEvidence, error) {
	zero := make([]int, len(profile.Boundaries))
	base, err := q056Evaluate(&profile, "R0-fixed-2^92", zero)
	if err != nil {
		return q056ControlEvidence{}, err
	}
	g0 := append([]int(nil), zero...)
	if len(g0) == 0 {
		return q056ControlEvidence{}, fmt.Errorf("control has no rescale boundaries")
	}
	g0[0] = 1
	g0Candidate, err := q056Evaluate(&profile, "Q0-control-G0-one-bit", g0)
	if err != nil {
		return q056ControlEvidence{}, err
	}
	f0 := append([]int(nil), zero...)
	f0[len(f0)-1] = 1
	f0Candidate, err := q056Evaluate(&profile, "Q0-control-F0-one-bit", f0)
	if err != nil {
		return q056ControlEvidence{}, err
	}
	g0Boundary := psRescaleGuardBoundaryByID(base.BoundaryEffects, "G0-rescale")
	ratio := float64(0)
	if g0Boundary != nil {
		ratio = g0Boundary.PreCapacityRatio
	}
	evidence := q056ControlEvidence{NoGuardMaxError: base.FinalMaxError, G0CapacityRatio: ratio, G0OneBit: q056CompactCandidate(g0Candidate), F0OneBit: q056CompactCandidate(f0Candidate), NoGuardExpectedRange: [2]float64{2.5e-8, 4.2e-8}, F0ExpectedMaxError: 2.5509e-8}
	evidence.ControlPreconditionOK = base.Valid && base.FinalMaxError > evidence.NoGuardExpectedRange[0] && base.FinalMaxError < evidence.NoGuardExpectedRange[1] && ratio > 0.85 && ratio < 0.95 && !g0Candidate.Valid && f0Candidate.Valid && math.Abs(f0Candidate.FinalMaxError-evidence.F0ExpectedMaxError) < 8e-9
	return evidence, nil
}

func q056MakeProfileEvidence(profile q056PreparedProfile, standardPublic *PSGlobalMetric) (q056ProfileEvidence, error) {
	evidence := q056ProfileEvidence{RequestedQ0Bits: 56, StandardPublic: standardPublic}
	q0, q1, q0Bits, q1Bits, q01, q01Bits, err := q056ProfilePrimes(profile.BTP.BootstrappingParameters)
	if err != nil {
		return evidence, err
	}
	evidence.RequestedQ0Bits = 56
	evidence.Q0, evidence.Q1, evidence.Q0Bits, evidence.Q1Bits, evidence.Q01, evidence.Q01Bits = q0, q1, q0Bits, q1Bits, q01, q01Bits
	q01Value, ok := new(big.Int).SetString(q01, 10)
	q01Limit := new(big.Int).Lsh(big.NewInt(1), 95)
	evidence.Q01WithinExistingDomain = ok && q0Bits <= 56 && q1Bits <= 39 && q01Value.Cmp(q01Limit) < 0
	evidence.FastRescaleAccepted = true
	zero := make([]int, len(profile.Boundaries))
	schedules := []struct {
		name string
		bits map[int]int
	}{
		{"W0-no-guards", nil},
		{"W1-G0-one-bit", map[int]int{0: 1}},
		{"W2-F0-one-bit", map[int]int{len(zero) - 1: 1}},
		{"W3-G0-F0-one-bit", map[int]int{0: 1, len(zero) - 1: 1}},
	}
	allCandidates := make([]psRescaleGuardCandidate, 0, len(schedules)+len(profile.Boundaries))
	for _, item := range schedules {
		schedule := append([]int(nil), zero...)
		for index, bits := range item.bits {
			schedule[index] = bits
		}
		candidate, err := q056Evaluate(&profile, item.name, schedule)
		if err != nil {
			return evidence, err
		}
		allCandidates = append(allCandidates, candidate)
	}
	for index, boundary := range profile.Boundaries {
		schedule := append([]int(nil), zero...)
		schedule[index] = 1
		candidate, err := q056Evaluate(&profile, "Q2-single-"+boundary.ID, schedule)
		if err != nil {
			return evidence, err
		}
		allCandidates = append(allCandidates, candidate)
		evidence.Boundaries = append(evidence.Boundaries, q056BoundaryEvidenceFor(allCandidates[0], candidate, boundary.ID, index))
	}
	current := allCandidates[3]
	currentSchedule := append([]int(nil), current.BoundaryGuardBits...)
	if current.Valid && !current.PrecisionQualified {
		for _, index := range []int{1, 2, 3} {
			if index >= len(currentSchedule) || currentSchedule[index] != 0 {
				continue
			}
			trialSchedule := append([]int(nil), currentSchedule...)
			trialSchedule[index] = 1
			trial, err := q056Evaluate(&profile, fmt.Sprintf("W4-greedy-%s", profile.Boundaries[index].ID), trialSchedule)
			if err != nil {
				return evidence, err
			}
			if trial.Valid && trial.FinalMaxError < current.FinalMaxError {
				evidence.GreedySteps = append(evidence.GreedySteps, psRescaleGuardGreedyStep{Step: len(evidence.GreedySteps) + 1, BoundaryID: profile.Boundaries[index].ID, FromBits: 0, ToBits: 1, BeforeMaxError: current.FinalMaxError, AfterMaxError: trial.FinalMaxError, Improvement: current.FinalMaxError - trial.FinalMaxError})
				allCandidates = append(allCandidates, trial)
				current, currentSchedule = trial, trialSchedule
			}
		}
	}
	best := psRescaleGuardCandidate{}
	for i := range allCandidates {
		candidate := allCandidates[i]
		if candidate.Valid && (best.Name == "" || psRescaleGuardSelectBetter(candidate, best, profile.Boundaries)) {
			best = candidate
		}
	}
	if best.Name != "" {
		compact := q056CompactCandidate(best)
		evidence.BestValid = &compact
	}
	selected := psRescaleGuardCandidate{}
	for i := range allCandidates {
		candidate := allCandidates[i]
		if candidate.PrecisionQualified && (selected.Name == "" || psRescaleGuardSelectBetter(candidate, selected, profile.Boundaries)) {
			selected = candidate
		}
	}
	if selected.Name != "" {
		compact := q056CompactCandidate(selected)
		evidence.Selected = &compact
		evidence.FirstBlocker = "none"
	} else if best.Name != "" {
		evidence.FirstBlocker = psRescaleGuardLimitingBoundary(best)
	} else {
		evidence.FirstBlocker = "no-valid-widened-q0-guard-candidate"
	}
	for _, candidate := range allCandidates[:4] {
		evidence.Candidates = append(evidence.Candidates, q056CompactCandidate(candidate))
	}
	return evidence, nil
}

func q056Write(result q056Result, outPath string) error {
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

func runFIX001P3DesignLogN13Q056RescaleGuardFeasibility(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := q056Result{
		SchemaVersion: "fix-001-p3-design-logn13-q0-56-rescale-guard-feasibility.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, PolynomialBudget: psRescaleGuardBudget, PublicThreshold: q056PublicThreshold,
		Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: 1 << (cfg.LogN - 1)},
		Validation: map[string]interface{}{
			"primary_required_base": requiredFIX001P3Q056PrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean,
			"secondary_exact_required_commit": secondaryCommit == requiredFIX001P3Q056Secondary, "no_secondary_production_changes": true,
			"logn13_only": cfg.LogN == 13, "no_q1_widening": true, "no_q0_beyond_56": true, "no_q2_plus_arithmetic": true,
			"same_plan_scale_2^92": true, "same_level_topology": true, "no_generated_power_redesign": true,
			"no_c2s_change": true, "no_s2c_change": true, "no_mod1_change": true, "no_production_integration": true,
			"no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true,
		},
	}
	if cfg.LogN != 13 || !secondaryClean || secondaryCommit != requiredFIX001P3Q056Secondary {
		result.Classification, result.FirstBlocker = "logn13_q0_56_guard_precondition_mismatch", "startup_precondition"
		return q056Write(result, outPath)
	}
	controlCfg := q056BuildConfig(cfg, 55)
	control, err := q056PrepareProfile(controlCfg)
	if err != nil {
		result.Classification, result.FirstBlocker = "logn13_q0_56_guard_precondition_mismatch", fmt.Sprintf("q0_control: %v", err)
		return q056Write(result, outPath)
	}
	result.Parameters = parameterMetadata(control.Residual, control.BTP)
	result.Control, err = q056BuildControlEvidence(control)
	if err != nil {
		return err
	}
	if !result.Control.ControlPreconditionOK {
		result.Classification, result.FirstBlocker = "logn13_q0_56_guard_precondition_mismatch", "q0_55_39_control"
		return q056Write(result, outPath)
	}
	widenedCfg := q056BuildConfig(cfg, 56)
	widened, err := q056PrepareProfile(widenedCfg)
	if err != nil {
		result.Classification, result.FirstBlocker = "logn13_q0_56_existing_backend_not_supported", err.Error()
		return q056Write(result, outPath)
	}
	result.Parameters = parameterMetadata(widened.Residual, widened.BTP)
	result.Widened.StandardPublic, err = q056StandardPublic(widened)
	if err != nil {
		result.Classification, result.FirstBlocker = "logn13_q0_56_existing_backend_not_supported", fmt.Sprintf("Standard public: %v", err)
		return q056Write(result, outPath)
	}
	result.Widened, err = q056MakeProfileEvidence(widened, result.Widened.StandardPublic)
	if err != nil {
		return err
	}
	result.Widened.RequestedQ0Bits = 56
	if !result.Widened.Q01WithinExistingDomain || !result.Widened.FastRescaleAccepted || result.Widened.StandardPublic == nil || !result.Widened.StandardPublic.Pass {
		result.Classification, result.FirstBlocker = "logn13_q0_56_existing_backend_not_supported", "widened-profile-domain-or-Standard-correctness"
		return q056Write(result, outPath)
	}
	if result.Widened.Selected != nil {
		result.Classification, result.FirstBlocker = "logn13_q0_56_rescale_guard_candidate_validated", "none"
		// Q5 is intentionally not added to the diagnostic artifact unless a Q4 candidate qualifies.
		result.Downstream = map[string]interface{}{"reached": false, "reason": "Q4 candidate is diagnostic-only; downstream oracle sufficiency remains unexecuted in this bounded feasibility runner"}
	} else {
		result.Classification = "logn13_q0_56_rescale_guard_insufficient_precision"
		if result.Widened.BestValid == nil {
			result.Classification = "logn13_q0_56_rescale_guard_noncapacity_semantic_failure"
		}
		result.FirstBlocker = result.Widened.FirstBlocker
	}
	return q056Write(result, outPath)
}
