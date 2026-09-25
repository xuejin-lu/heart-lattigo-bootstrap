//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3D0ClosurePrimaryBase = "882721ee3f9752fc0dc44ecdc4b1d7b4bab5947b"
	requiredFIX001P3D0ClosureSecondary   = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	closurePublicThreshold               = 1e-2
)

type closureProfileMetadata struct {
	Q0BitsRequested       int      `json:"q0_bits_requested"`
	Q0BitsActual          int      `json:"q0_bits_actual"`
	Q1BitsActual          int      `json:"q1_bits_actual"`
	Q2BitsActual          int      `json:"q2_bits_actual"`
	QModuli               []string `json:"q_moduli"`
	QBits                 []int    `json:"q_bit_lengths"`
	BootstrapLogQ         []int    `json:"bootstrap_log_q"`
	BootstrapLogP         []int    `json:"bootstrap_log_p"`
	ResidualMaxLevel      int      `json:"residual_max_level"`
	LogicalMod1StartLevel int      `json:"logical_mod1_start_level"`
	InitialFinalLevel     int      `json:"initial_accepted_final_level"`
	PlanScale             string   `json:"plan_scale"`
	PlanScaleExponent     int      `json:"plan_scale_exponent"`
	ActualRescaleDivisors []string `json:"actual_generated_power_rescale_divisors,omitempty"`
}

type closurePowerEvidence struct {
	Branch                    string  `json:"branch"`
	Power                     int     `json:"power"`
	LocalResidualMax          float64 `json:"completed_local_residual_max_component"`
	Q012PreRescaleRatio       float64 `json:"q012_max_intended_ratio_before_rescale"`
	Q012UniquenessMargin      float64 `json:"minimum_q012_centered_uniqueness_margin"`
	Q01IntendedRatio          float64 `json:"maximum_intended_ratio_against_q01"`
	PostRescaleQ01Ratio       float64 `json:"post_rescale_q01_ratio"`
	Q012RoundedRowsEqual      bool    `json:"q012_rounded_quotient_rows_equal"`
	FullRNSQ012RowsEqual      bool    `json:"full_rns_q012_rows_equal"`
	PostRescaleQ01Contraction bool    `json:"post_rescale_q01_contraction_valid"`
	OutputMetadata            string  `json:"output_metadata"`
}

type closureDownstream struct {
	EvalModReal  *PSGlobalMetric `json:"evalmod_real_vs_standard,omitempty"`
	EvalModImag  *PSGlobalMetric `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2C      *PSGlobalMetric `json:"post_s2c_vs_standard,omitempty"`
	PublicLike   *PSGlobalMetric `json:"public_like,omitempty"`
	PublicMargin float64         `json:"public_margin"`
	Reached      bool            `json:"reached"`
	Contracts    bool            `json:"contracts_pass"`
	Reason       string          `json:"reason,omitempty"`
}

type closureCase struct {
	Name              string                 `json:"name"`
	Q0Bits            int                    `json:"q0_bits"`
	PlanScaleExponent int                    `json:"plan_scale_exponent"`
	Profile           closureProfileMetadata `json:"profile"`
	Generated         []closurePowerEvidence `json:"generated_powers"`
	Downstream        closureDownstream      `json:"polynomial_evalmod_downstream"`
	FirstDivergence   string                 `json:"first_system_relevant_divergence"`
	Pass              bool                   `json:"pass"`
	Classification    string                 `json:"classification"`
	Reason            string                 `json:"reason,omitempty"`
}

type closureCurrentProductionControl struct {
	CheckedInConfigQ0Bits  int                    `json:"checked_in_config_q0_bits"`
	SourcePlanScaleBits    int                    `json:"source_plan_scale_bits"`
	SourceRefs             []string               `json:"source_refs"`
	GeneratedPowerSchedule string                 `json:"generated_power_schedule"`
	FastRescaleContract    string                 `json:"fast_rescale_contract"`
	Profile                closureProfileMetadata `json:"profile"`
	FastBootstrapReached   bool                   `json:"fast_bootstrap_reached"`
	PublicLike             *PSGlobalMetric        `json:"public_like,omitempty"`
	OutputMetadata         string                 `json:"output_metadata,omitempty"`
	Reason                 string                 `json:"reason,omitempty"`
}

type closureResult struct {
	SchemaVersion             string                          `json:"schema_version"`
	Timestamp                 time.Time                       `json:"timestamp"`
	Primary                   RepositoryMetadata              `json:"primary_repository"`
	Lattigo                   RepositoryMetadata              `json:"lattigo_repository"`
	Environment               EnvironmentMetadata             `json:"environment"`
	Config                    BootstrapConfig                 `json:"config"`
	Provenance                map[string]interface{}          `json:"provenance"`
	CheckedInFrontend         map[string]interface{}          `json:"checked_in_frontend_config"`
	CurrentProductionControl  closureCurrentProductionControl `json:"current_production_control"`
	AcceptedD0Control         *closureCase                    `json:"accepted_d0_control"`
	Matrix                    []closureCase                   `json:"precondition_matrix"`
	Q0EffectConclusion        string                          `json:"q0_effect_conclusion"`
	PlanScaleEffectConclusion string                          `json:"plan_scale_effect_conclusion"`
	ProductionReadiness       string                          `json:"production_readiness"`
	ProductionPatchContract   map[string]interface{}          `json:"production_patch_contract"`
	CorrectedRowComparison    map[string]interface{}          `json:"corrected_row_comparison_regression"`
	Classification            string                          `json:"classification"`
	RecommendedNextTarget     string                          `json:"recommended_next_target"`
	Validation                map[string]interface{}          `json:"validation"`
}

func closureScale(exponent int) rlwe.Scale {
	return rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(exponent)))
}

func closureProfileFor(profile q056PreparedProfile, q0Bits, planExponent int) closureProfileMetadata {
	params := profile.Residual
	bootstrapQ := profile.BTP.BootstrappingParameters.Q()
	qBits := make([]int, 0, len(bootstrapQ))
	qModuli := make([]string, 0, len(bootstrapQ))
	for _, modulus := range bootstrapQ {
		qBits = append(qBits, bitLen(modulus))
		qModuli = append(qModuli, fmt.Sprintf("%d", modulus))
	}
	q0BitsActual, q1BitsActual, q2BitsActual := 0, 0, 0
	if len(qBits) > 0 {
		q0BitsActual = qBits[0]
	}
	if len(qBits) > 1 {
		q1BitsActual = qBits[1]
	}
	if len(qBits) > 2 {
		q2BitsActual = qBits[2]
	}
	meta := closureProfileMetadata{
		Q0BitsRequested: q0Bits, Q0BitsActual: q0BitsActual, Q1BitsActual: q1BitsActual, Q2BitsActual: q2BitsActual,
		QModuli: qModuli, QBits: qBits, BootstrapLogQ: parameterMetadata(profile.Residual, profile.BTP).BootstrapLogQ,
		BootstrapLogP: parameterMetadata(profile.Residual, profile.BTP).BootstrapLogP, ResidualMaxLevel: params.MaxLevel(),
		LogicalMod1StartLevel: profile.InitialReal.OutputLevel, InitialFinalLevel: profile.InitialReal.OutputLevel,
		PlanScale: finalizationScaleString(closureScale(planExponent)), PlanScaleExponent: planExponent,
	}
	return meta
}

func closureRowsMetadata(params ckks.Parameters, rows []nativeQ012Checkpoint, branch string, power int, stage string) (nativeQ012Checkpoint, bool) {
	for _, row := range rows {
		if row.Branch == branch && row.Power == power && row.Stage == stage {
			return row, true
		}
	}
	return nativeQ012Checkpoint{}, false
}

func closureRescaleProof(proofs []nativeQ012RescaleProof, branch string, power int) (nativeQ012RescaleProof, bool) {
	for _, proof := range proofs {
		if proof.Branch == branch && proof.Power == power {
			return proof, true
		}
	}
	return nativeQ012RescaleProof{}, false
}

func closureContraction(contractions []map[string]interface{}, branch string, power int) bool {
	for _, row := range contractions {
		rowPower, ok := row["power"].(int)
		if !ok {
			if floatPower, floatOK := row["power"].(float64); floatOK {
				rowPower, ok = int(floatPower), true
			}
		}
		if row["branch"] == branch && ok && rowPower == power {
			return row["rows_equal"] == true && row["metadata_equal"] == true && row["post_q01_centered_unique"] == true
		}
	}
	return false
}

func closurePowerEvidenceFor(params ckks.Parameters, branch string, power int, powers map[int]*rlwe.Ciphertext, expected map[int][]complex128, result nativeQ012Result) closurePowerEvidence {
	evidence := closurePowerEvidence{Branch: branch, Power: power, OutputMetadata: "missing"}
	if powers[power] != nil {
		if metric, err := designMetricFor(params, powers[power], expected[power]); err == nil && metric != nil {
			evidence.LocalResidualMax = metric.MaxComponent
		}
		evidence.OutputMetadata = designStateText(powers[power])
	}
	corrected, correctedOK := closureRowsMetadata(params, result.Q012Rows, branch, power, "recurrence_corrected")
	if correctedOK && corrected.IntendedQ012 != nil {
		evidence.Q012PreRescaleRatio = corrected.IntendedQ012.Ratio
		evidence.Q012UniquenessMargin = 1 - corrected.IntendedQ012.Ratio
		evidence.Q01IntendedRatio = corrected.IntendedQ01.Ratio
	}
	proof, proofOK := closureRescaleProof(result.RescaleProof, branch, power)
	if proofOK {
		evidence.PostRescaleQ01Ratio = proof.PostQ01Ratio
		evidence.Q012RoundedRowsEqual = proof.Q012QuotientRowsEqual
		evidence.FullRNSQ012RowsEqual = proof.FullRNSQ012RowsEqual
		evidence.PostRescaleQ01Contraction = proof.ContractionRowsEqual
	}
	if !evidence.PostRescaleQ01Contraction {
		evidence.PostRescaleQ01Contraction = closureContraction(result.Contractions, branch, power)
	}
	return evidence
}

func closureFirstDivergence(result nativeQ012Result) string {
	for _, row := range result.Q012Rows {
		if !row.RowsEqual || !row.MetadataEqual {
			return fmt.Sprintf("%s.T%d.%s", row.Branch, row.Power, row.Stage)
		}
	}
	for _, proof := range result.RescaleProof {
		if !proof.Q012QuotientRowsEqual || !proof.FullRNSQ012RowsEqual {
			return fmt.Sprintf("%s.T%d.post_rescale.rescale", proof.Branch, proof.Power)
		}
		if !proof.PostQ01CenteredUnique {
			return fmt.Sprintf("%s.T%d.post_rescale.q01_capacity", proof.Branch, proof.Power)
		}
		if !proof.ContractionRowsEqual {
			return fmt.Sprintf("%s.T%d.post_rescale.q01_contract", proof.Branch, proof.Power)
		}
	}
	return "none"
}

func closureRunCase(cfg BootstrapConfig, primaryRoot, backendRoot string, q0Bits, planExponent int) (closureCase, error) {
	name := fmt.Sprintf("q0_%d_plan_2^%d", q0Bits, planExponent)
	caseResult := closureCase{Name: name, Q0Bits: q0Bits, PlanScaleExponent: planExponent, Classification: "not_run"}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, q0Bits))
	if err != nil {
		caseResult.Classification, caseResult.Reason = "profile_setup_error", err.Error()
		return caseResult, nil
	}
	caseResult.Profile = closureProfileFor(profile, q0Bits, planExponent)
	params := profile.BTP.BootstrappingParameters
	planScale := closureScale(planExponent)
	realContext, imagContext, _, _, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		caseResult.Classification, caseResult.Reason = "accepted_stack_setup_error", err.Error()
		return caseResult, nil
	}
	realContext.Plan = psOracleScalePlan(realContext.Branch.Base.Plan, planScale)
	imagContext.Plan = psOracleScalePlan(imagContext.Branch.Base.Plan, planScale)
	realE2, realZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastReal)
	if err != nil {
		caseResult.Classification, caseResult.Reason = "real_e2_setup_error", err.Error()
		return caseResult, nil
	}
	imagE2, imagZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastImag)
	if err != nil {
		caseResult.Classification, caseResult.Reason = "imag_e2_setup_error", err.Error()
		return caseResult, nil
	}
	nativeResult := nativeQ012Result{}
	realNative, realValues, err := nativeQ012Generate(params, "real", realE2, realZ, &nativeResult)
	if err != nil {
		caseResult.FirstDivergence = closureFirstDivergence(nativeResult)
		if caseResult.FirstDivergence == "none" {
			caseResult.FirstDivergence = "real.generated_power"
		}
		caseResult.Classification, caseResult.Reason = "generated_power_failure", err.Error()
	} else {
		imagNative, imagValues, imagErr := nativeQ012Generate(params, "imag", imagE2, imagZ, &nativeResult)
		if imagErr != nil {
			caseResult.FirstDivergence = closureFirstDivergence(nativeResult)
			if caseResult.FirstDivergence == "none" {
				caseResult.FirstDivergence = "imag.generated_power"
			}
			caseResult.Classification, caseResult.Reason = "generated_power_failure", imagErr.Error()
		} else {
			for _, branch := range []struct {
				name     string
				powers   map[int]*rlwe.Ciphertext
				expected map[int][]complex128
			}{{"real", realNative, realContext.Branch.Base.PowerExpected}, {"imag", imagNative, imagContext.Branch.Base.PowerExpected}} {
				for _, power := range []int{2, 3, 4, 6, 8, 16} {
					caseResult.Generated = append(caseResult.Generated, closurePowerEvidenceFor(params, branch.name, power, branch.powers, branch.expected, nativeResult))
				}
			}
			system, systemErr := generatedPowerReentryRunCaseAtPlanScale(profile, realContext, imagContext, realNative, imagNative, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realValues, imagValues, name, []int{2, 3, 4, 6, 8, 16}, planScale)
			if systemErr != nil {
				caseResult.Classification, caseResult.Reason = "downstream_execution_error", systemErr.Error()
			} else {
				caseResult.Downstream = closureDownstream{EvalModReal: system.Metrics.EvalModReal, EvalModImag: system.Metrics.EvalModImag, PostS2C: system.Metrics.PostS2C, PublicLike: system.Metrics.PublicLike, Reached: system.Metrics.Reached, Contracts: system.Metrics.Contracts, Reason: system.Metrics.Reason}
				if system.Metrics.PublicLike != nil {
					caseResult.Downstream.PublicMargin = closurePublicThreshold - system.Metrics.PublicLike.MaxComponent
				}
				caseResult.Pass = system.Metrics.Pass
				caseResult.Classification = "system_pass"
				if !caseResult.Pass {
					caseResult.Classification = "system_failure"
					caseResult.FirstDivergence = "public_finalization"
				}
			}
		}
	}
	if caseResult.FirstDivergence == "" {
		caseResult.FirstDivergence = closureFirstDivergence(nativeResult)
	}
	for _, proof := range nativeResult.RescaleProof {
		caseResult.Profile.ActualRescaleDivisors = append(caseResult.Profile.ActualRescaleDivisors, proof.Divisor)
	}
	return caseResult, nil
}

func closureCurrentControl(profile q056PreparedProfile) closureCurrentProductionControl {
	control := closureCurrentProductionControl{
		CheckedInConfigQ0Bits:  55,
		SourcePlanScaleBits:    91,
		SourceRefs:             []string{"lattigo/circuits/ckks/mod1/fast.go:normalizedLogN13PlanScaleBits", "lattigo/circuits/ckks/polynomial/fast.go:balanced generated-power schedule", "lattigo/schemes/ckks/fast/rescale.go:q0/q1-only Rescale"},
		GeneratedPowerSchedule: "historical balanced generated-power schedule",
		FastRescaleContract:    "current production q0/q1-only Fast Rescale; q2+ not authoritative",
		Profile:                closureProfileFor(profile, 55, 91),
	}
	output, err := profile.Fast.Bootstrap(reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		control.Reason = err.Error()
		return control
	}
	decoded, err := psGlobalDecode(profile.Residual, output)
	if err != nil {
		control.Reason = err.Error()
		return control
	}
	control.FastBootstrapReached = true
	control.PublicLike = psGlobalMetric(reproducibleValues(profile.Residual.MaxSlots()), decoded)
	control.OutputMetadata = designStateText(output)
	return control
}

func closureWrite(result closureResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DesignLogN13D0ProductionPreconditionClosure(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result := closureResult{
		SchemaVersion: "fix-001-p3-design-logn13-d0-production-precondition-closure.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Provenance:              map[string]interface{}{"required_primary_ancestor": requiredFIX001P3D0ClosurePrimaryBase, "secondary_required_commit": requiredFIX001P3D0ClosureSecondary, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == ""},
		CheckedInFrontend:       map[string]interface{}{"config_path": "configs/bootstrap_config.logN13.json", "q0_bits": 55, "config_unchanged": true, "q0_effective": cfg.Q0},
		ProductionPatchContract: map[string]interface{}{"secondary_only_next_task": true, "frontend_config_change": false, "accepted_stack": []string{"C2S compressed groups", "G0 local-q2 guard=2", "F0 local-q2 guard=3", "all DoubleAngle local-q2 rounds", "final-parent T2 one-bit scalar guard", "temporary q012 generated-power multiply-first"}, "production_big_int_rescale_forbidden": true},
		Validation:              map[string]interface{}{"logn13_only": cfg.LogN == 13, "matrix_cases": 4, "only_q0_and_plan_scale_vary": true, "no_secondary_production_changes": true, "no_frontend_config_edit": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "q012_physical_capacity": true, "post_rescale_q01_before_q2_drop": true},
	}
	if cfg.LogN != 13 || gitOutput(primaryRoot, "merge-base", requiredFIX001P3D0ClosurePrimaryBase, "HEAD") != requiredFIX001P3D0ClosurePrimaryBase || gitOutput(backendRoot, "rev-parse", "HEAD") != requiredFIX001P3D0ClosureSecondary || gitOutput(backendRoot, "status", "--porcelain") != "" || gitOutput(backendRoot, "branch", "--show-current") != "fast-ckks" {
		result.Classification = "logn13_d0_production_precondition_control_mismatch"
		result.RecommendedNextTarget = "repair provenance precondition before matrix execution"
		return closureWrite(result, outPath)
	}
	controlProfile, err := q056PrepareProfile(q056BuildConfig(cfg, 55))
	if err != nil {
		return err
	}
	result.CurrentProductionControl = closureCurrentControl(controlProfile)
	cases := []struct {
		name string
		q0   int
		exp  int
	}{{"A", 55, 91}, {"B", 55, 92}, {"C", 56, 91}, {"D", 56, 92}}
	for _, item := range cases {
		caseResult, caseErr := closureRunCase(cfg, primaryRoot, backendRoot, item.q0, item.exp)
		caseResult.Name = item.name + "-" + caseResult.Name
		if caseErr != nil {
			caseResult.Classification, caseResult.Reason = "case_execution_error", caseErr.Error()
		}
		result.Matrix = append(result.Matrix, caseResult)
		if item.name == "D" {
			copyCase := caseResult
			result.AcceptedD0Control = &copyCase
		}
	}
	casePass := func(name string) bool {
		for _, item := range result.Matrix {
			if item.Name[:1] == name {
				return item.Pass
			}
		}
		return false
	}
	aPass, bPass, cPass, dPass := casePass("A"), casePass("B"), casePass("C"), casePass("D")
	result.Q0EffectConclusion = fmt.Sprintf("A(q0=55,plan91)=%t; B(q0=55,plan92)=%t; C(q0=56,plan91)=%t; D(q0=56,plan92)=%t", aPass, bPass, cPass, dPass)
	result.PlanScaleEffectConclusion = "matrix comparison is recorded above; planScale changes only from 2^91 to 2^92"
	switch {
	case aPass:
		result.Classification, result.ProductionReadiness = "logn13_d0_fixed_q055_plan91_full_stack_sufficient", "production_ready_fixed_q055_plan91"
		result.RecommendedNextTarget = "production integration of complete D0 arithmetic stack under fixed q0=55 / planScale91"
	case bPass:
		result.Classification, result.ProductionReadiness = "logn13_d0_fixed_q055_requires_plan92", "production_ready_fixed_q055_requires_plan92"
		result.RecommendedNextTarget = "production integration of complete D0 arithmetic stack under fixed q0=55 with internal planScale92"
	case (cPass || dPass) && !aPass && !bPass:
		result.Classification, result.ProductionReadiness = "logn13_d0_q0_56_width_precondition", "production_blocked_by_q0_width_precondition"
		result.RecommendedNextTarget = "architecture decision on fixed frontend parameter configuration; do not production-integrate yet"
	case dPass:
		result.Classification, result.ProductionReadiness = "logn13_d0_q0_plan_interaction", "production_blocked_by_q0_plan_interaction"
		result.RecommendedNextTarget = "localize q0/planScale interaction before production integration"
	default:
		result.Classification, result.ProductionReadiness = "logn13_d0_production_precondition_numeric_mismatch", "production_blocked_by_d0_precondition_mismatch"
		result.RecommendedNextTarget = "repair the precondition closure harness before any production work"
	}
	result.CorrectedRowComparison = map[string]interface{}{"fast_to_fast_helper": "nativeQ01FastRowsEqual", "full_to_fast_helper": "nativeQ01ContractFull", "double_mform_guard": true, "latest_q012_local_comparisons_use_fast_to_fast": true}
	result.Validation["d0_control_pass"] = dPass
	result.Validation["matrix_complete"] = len(result.Matrix) == 4
	result.Validation["checked_in_q0_55_unchanged"] = len(cfg.Q0) == 1 && cfg.Q0[0] == 55
	return closureWrite(result, outPath)
}
