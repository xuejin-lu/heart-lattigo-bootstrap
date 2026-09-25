//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	fix001P3P93S2CSecondaryCommit = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3P93S2CDiffSHA256      = "44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e"
	fix001P3P93S2CExpectedF0      = 0.0402736269192105
	fix001P3P93S2CThreshold       = 1e-2
)

type fix001P3P93S2CGroupEvidence struct {
	Stage                string          `json:"stage"`
	InputPropagation     *PSGlobalMetric `json:"input_propagation"`
	ImplementationEffect *PSGlobalMetric `json:"implementation_effect"`
	Amplification        float64         `json:"empirical_amplification"`
	FastPreRowsMatch     bool            `json:"fast_pre_rows_match_mirror"`
	FastPostRowsMatch    bool            `json:"fast_post_rows_match_mirror"`
	LevelEqual           bool            `json:"level_equal"`
	ScaleEqual           bool            `json:"scale_equal"`
	DegreeEqual          bool            `json:"degree_equal"`
	FirstMaterial        bool            `json:"first_material_implementation_effect"`
}

type fix001P3P93S2CResult struct {
	SchemaVersion  string                        `json:"schema_version"`
	Timestamp      time.Time                     `json:"timestamp"`
	Primary        RepositoryMetadata            `json:"primary_repository"`
	Lattigo        RepositoryMetadata            `json:"lattigo_repository"`
	Environment    EnvironmentMetadata           `json:"environment"`
	Config         BootstrapConfig               `json:"config"`
	Parameters     ExperimentParameters          `json:"effective_parameters"`
	Architecture   map[string]interface{}        `json:"fixed_p93_architecture"`
	Provenance     map[string]interface{}        `json:"secondary_provenance"`
	R0             map[string]interface{}        `json:"r0_reproduction"`
	Mirror         map[string]interface{}        `json:"r2_mirror_validation"`
	Decomposition  map[string]interface{}        `json:"r3_causal_decomposition"`
	Groups         []fix001P3P93S2CGroupEvidence `json:"r4_group_attribution"`
	LinearDelta    map[string]interface{}        `json:"r5_linear_delta_confirmation"`
	Reconciliation []map[string]interface{}      `json:"r6_reconciliation"`
	Classification string                        `json:"classification"`
	FirstBlocker   string                        `json:"first_remaining_blocker"`
	Validation     map[string]interface{}        `json:"validation"`
}

type fix001P3P93S2CAcceptedCandidate struct {
	PlanScaleExponent int `json:"plan_scale_exponent"`
	Real              struct {
		Polynomial         *PSGlobalMetric        `json:"ps_polynomial_vs_standard"`
		PublicLike         *PSGlobalMetric        `json:"public_like"`
		StandardPublicLike *PSGlobalMetric        `json:"standard_public_like"`
		EvalMod            *PSGlobalMetric        `json:"evalmod_vs_standard"`
		PostS2C            *PSGlobalMetric        `json:"post_s2c_vs_standard"`
		Contraction        psWideContractionProof `json:"ps_exit_contraction"`
	} `json:"real"`
	Imag struct {
		Polynomial         *PSGlobalMetric        `json:"ps_polynomial_vs_standard"`
		PublicLike         *PSGlobalMetric        `json:"public_like"`
		StandardPublicLike *PSGlobalMetric        `json:"standard_public_like"`
		EvalMod            *PSGlobalMetric        `json:"evalmod_vs_standard"`
		PostS2C            *PSGlobalMetric        `json:"post_s2c_vs_standard"`
		Contraction        psWideContractionProof `json:"ps_exit_contraction"`
	} `json:"imag"`
}

type fix001P3P93S2CAcceptedArtifact struct {
	Candidates []fix001P3P93S2CAcceptedCandidate `json:"candidates"`
}

func fix001P3P93S2CLoadAcceptedArtifact() (fix001P3P93S2CAcceptedCandidate, error) {
	data, err := os.ReadFile("results/FIX-001-P3-DESIGN-LOGN13-PS-WIDE-Q012-Q056-SCALE-SWEEP-summary.json")
	if err != nil {
		return fix001P3P93S2CAcceptedCandidate{}, err
	}
	var artifact fix001P3P93S2CAcceptedArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return fix001P3P93S2CAcceptedCandidate{}, err
	}
	for _, candidate := range artifact.Candidates {
		if candidate.PlanScaleExponent == 93 {
			return candidate, nil
		}
	}
	return fix001P3P93S2CAcceptedCandidate{}, fmt.Errorf("accepted P93 candidate is missing from committed design evidence")
}

func fix001P3P93S2CWrite(result fix001P3P93S2CResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func fix001P3P93S2CReferencePaths(cfg BootstrapConfig, profile q056PreparedProfile) (evalModMatchedPath, evalModMatchedPath, error) {
	effective := q056BuildConfig(cfg, 56)
	realBranch, realContracted, err := psWideRunBranch(effective, "real", profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.BTP, profile.Fast, profile.Standard, profile.StandardSK, 93)
	if err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	if realContracted == nil || realBranch.FirstFailure != "none" {
		return evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("P93 reference real path failed: %s", realBranch.FirstFailure)
	}
	imagBranch, imagContracted, err := psWideRunBranch(effective, "imag", profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.BTP, profile.Fast, profile.Standard, profile.StandardSK, 93)
	if err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	if imagContracted == nil || imagBranch.FirstFailure != "none" {
		return evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("P93 reference imag path failed: %s", imagBranch.FirstFailure)
	}
	params := profile.BTP.BootstrappingParameters
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, params, profile.Inputs.FastSK, profile.StandardSK, precisionSweepScale(93), realContracted)
	if err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, params, profile.Inputs.FastSK, profile.StandardSK, precisionSweepScale(93), imagContracted)
	if err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	if realPath.FastFinal == nil || realPath.StandardFinal == nil || imagPath.FastFinal == nil || imagPath.StandardFinal == nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("P93 reference path returned nil final ciphertext")
	}
	return realPath, imagPath, nil
}

func fix001P3P93S2CMetadataEqual(a, b *rlwe.Ciphertext) (bool, bool, bool) {
	if a == nil || b == nil {
		return false, false, false
	}
	return a.Level() == b.Level(), a.Scale.Equal(b.Scale), a.Degree() == b.Degree()
}

func fix001P3P93S2CRunFastEvalModCore(eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, nil, nil, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return nil, nil, nil, err
	}
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return nil, nil, nil, err
	}
	real, imag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return nil, nil, nil, err
	}
	real, err = eval.EvalMod(real)
	if err != nil {
		return nil, nil, nil, err
	}
	if imag != nil {
		imag, err = eval.EvalMod(imag)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	core, err := eval.SlotsToCoeffs(real, imag)
	if err != nil {
		return nil, nil, nil, err
	}
	return real, imag, core, nil
}

func runFIX001P3DiagP93S2CAttribution(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	if cfg.LogN != 13 {
		return fmt.Errorf("P93 S2C attribution requires LogN13")
	}
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3P93S2CResult{
		SchemaVersion: "fix-001-p3-diag-p93-s2c-attribution-and-reference-reconciliation.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Architecture: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "ps_authority": "Q012-wide", "plan_scale": "2^93", "slots": 4096, "threshold": fix001P3P93S2CThreshold},
		Provenance:   map[string]interface{}{"branch": secondaryBranch, "committed_head": secondaryCommit, "dirty": secondaryDirty, "diff_stat": diffStat, "diff_sha256": diffHash, "expected_committed_head": fix001P3P93S2CSecondaryCommit, "expected_diff_sha256": fix001P3P93S2CDiffSHA256},
		Validation:   map[string]interface{}{"logn13_only": true, "no_secondary_source_modification": true, "no_parameter_tuning": true, "no_t2_repair": true, "no_ps_repair": true, "no_power_replacement_sweep": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
	}
	if secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3P93S2CSecondaryCommit || diffHash != fix001P3P93S2CDiffSHA256 || !secondaryDirty {
		result.Classification = "SECONDARY_HANDOFF_STATE_MISMATCH"
		result.FirstBlocker = "secondary branch, committed HEAD, dirty status, or diff fingerprint changed"
		return fix001P3P93S2CWrite(result, outPath)
	}
	accepted, err := fix001P3P93S2CLoadAcceptedArtifact()
	if err != nil {
		return err
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	inputs := profile.Inputs
	refPublic, refStandardLike, refPostS2C := accepted.Real.PublicLike, accepted.Real.StandardPublicLike, accepted.Real.PostS2C
	prodReal, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastReal, inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, inputs.FastSK, profile.StandardSK, precisionSweepScale(93), nil)
	if err != nil {
		return err
	}
	prodImag, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastImag, inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, inputs.FastSK, profile.StandardSK, precisionSweepScale(93), nil)
	if err != nil {
		return err
	}
	if prodReal.FastFinal == nil || prodReal.StandardFinal == nil || prodImag.FastFinal == nil || prodImag.StandardFinal == nil {
		return fmt.Errorf("current production P93 path returned nil final ciphertext")
	}
	productionReal, productionImag, prodFastCore, err := fix001P3P93S2CRunFastEvalModCore(profile.Fast, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	productionRealValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, productionReal, zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	productionImagValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, productionImag, zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	matchedStandardRealValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, prodReal.StandardFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedStandardImagValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, prodImag.StandardFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	productionEvalModReal := fix001P3S2CMetric(matchedStandardRealValues, productionRealValues, fix001P3P93S2CThreshold)
	productionEvalModImag := fix001P3S2CMetric(matchedStandardImagValues, productionImagValues, fix001P3P93S2CThreshold)
	refStandardCore, err := profile.Standard.SlotsToCoeffs(prodReal.StandardFinal.CopyNew(), prodImag.StandardFinal.CopyNew())
	if err != nil {
		return err
	}
	refStandardValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, refStandardCore, profile.StandardSK)
	if err != nil {
		return err
	}
	prodFastValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, prodFastCore, zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	prodF0Metric := fix001P3S2CMetric(refStandardValues, prodFastValues, fix001P3P93S2CThreshold)
	result.R0 = map[string]interface{}{
		"accepted_p93_design_proxy":         map[string]interface{}{"method": "committed FIX-001-P3-DESIGN-LOGN13-PS-WIDE-Q012-Q056-SCALE-SWEEP P93 compact evidence; dirty production backend is not reused for R0-A", "source_artifact": "results/FIX-001-P3-DESIGN-LOGN13-PS-WIDE-Q012-Q056-SCALE-SWEEP-summary.json", "evalmod_real": accepted.Real.EvalMod, "evalmod_imag": accepted.Imag.EvalMod, "post_s2c_reference_proxy": refPostS2C, "public_like": refPublic, "standard_public_like": refStandardLike, "historical_public_like": 0.009705381393898434},
		"current_dirty_production_boundary": map[string]interface{}{"method": "fix001P3P93S2CRunFastEvalModCore: Pack/ScaleDown/ModUp/C2S/EvalMod/SlotsToCoeffs; F0 compared with matched StandardFinal -> Standard SlotsToCoeffs", "evalmod_real": productionEvalModReal, "evalmod_imag": productionEvalModImag, "post_s2c_f0": prodF0Metric, "expected_f0": fix001P3P93S2CExpectedF0, "f0_reproduced": math.Abs(prodF0Metric.MaxComponent-fix001P3P93S2CExpectedF0) <= 2e-5},
	}
	if refPublic == nil || refPublic.MaxComponent > fix001P3P93S2CThreshold || math.Abs(refPublic.MaxComponent-0.009705381393898434) > 0.003 || math.Abs(prodF0Metric.MaxComponent-fix001P3P93S2CExpectedF0) > 2e-5 {
		if refPublic == nil || refPublic.MaxComponent > fix001P3P93S2CThreshold || math.Abs(refPublic.MaxComponent-0.009705381393898434) > 0.003 {
			result.Classification = "P93_REFERENCE_REPRODUCTION_CONFLICT"
			result.FirstBlocker = "accepted P93 design proxy did not reproduce on the documented side of the 1e-2 gate"
		} else {
			result.Classification = "P93_PRODUCTION_BOUNDARY_REPRODUCTION_CONFLICT"
			result.FirstBlocker = "current dirty production F0 did not reproduce the committed boundary"
		}
		return fix001P3P93S2CWrite(result, outPath)
	}
	params := profile.BTP.BootstrappingParameters
	fullRefReal, _, err := postMod1S2CLiftFull(params, prodReal.StandardFinal)
	if err != nil {
		return err
	}
	fullRefImag, _, err := postMod1S2CLiftFull(params, prodImag.StandardFinal)
	if err != nil {
		return err
	}
	fullProdReal, _, err := postMod1S2CLiftFull(params, productionReal)
	if err != nil {
		return err
	}
	fullProdImag, _, err := postMod1S2CLiftFull(params, productionImag)
	if err != nil {
		return err
	}
	trace, err := fix001P3S2CTraceBoth(params, profile.Fast, profile.Fast.S2CDFTMatrix, profile.Standard.DFTEvaluator, productionReal, productionImag, fullProdReal, fullProdImag)
	if err != nil {
		return err
	}
	refTrace, _, err := fix001P3S2CTraceFull(params, profile.Fast.S2CDFTMatrix, profile.Standard.DFTEvaluator, fullRefReal, fullRefImag)
	if err != nil {
		return err
	}
	fastOutputValues, err := fix001P3S2CDecode(params, trace.FastStages[len(trace.FastStages)-1], zeroSecret(params))
	if err != nil {
		return err
	}
	fullProdValues, err := fix001P3S2CDecode(params, trace.FullStages[len(trace.FullStages)-1], zeroSecret(params))
	if err != nil {
		return err
	}
	refOutputValues, err := fix001P3S2CDecode(params, refTrace[len(refTrace)-1], profile.StandardSK)
	if err != nil {
		return err
	}
	observed := fix001P3S2CMetric(refOutputValues, fastOutputValues, fix001P3P93S2CThreshold)
	implementation := fix001P3S2CMetric(fullProdValues, fastOutputValues, fix001P3P93S2CThreshold)
	propagation := fix001P3S2CMetric(refOutputValues, fullProdValues, fix001P3P93S2CThreshold)
	refMirrorValues, err := fix001P3S2CDecode(params, refTrace[len(refTrace)-1], zeroSecret(params))
	if err != nil {
		return err
	}
	fullProdInputValues, err := fix001P3S2CDecode(params, fullProdReal, zeroSecret(params))
	if err != nil {
		return err
	}
	prodFastInputValues, err := fix001P3S2CDecode(params, productionReal, zeroSecret(params))
	if err != nil {
		return err
	}
	refMirror := fix001P3S2CMetric(refOutputValues, refMirrorValues, 1e-10)
	prodInputMirror := fix001P3S2CMetric(fullProdInputValues, prodFastInputValues, 1e-10)
	fullProdCombined, err := fix001P3S2CCombineFull(params, profile.Standard.DFTEvaluator, fullProdReal, fullProdImag)
	if err != nil {
		return err
	}
	implVector := fix001P3S2CSubtract(fastOutputValues, fullProdValues)
	propVector := fix001P3S2CSubtract(fullProdValues, refOutputValues)
	observedVector := fix001P3S2CSubtract(fastOutputValues, refOutputValues)
	closureVector := fix001P3S2CSubtract(fix001P3S2CAdd3(implVector, propVector, make([]complex128, len(implVector))), observedVector)
	closure := fix001P3S2CMetric(make([]complex128, len(closureVector)), closureVector, 1e-10)
	result.Mirror = map[string]interface{}{"reference_full_rns_mirror_fidelity": refMirror, "production_input_full_rns_mirror_fidelity": prodInputMirror, "tolerance": 1e-10, "reference_rows_match": postMod1S2CRowsEqual(params, refTrace[len(refTrace)-1], refStandardCore), "production_rows_match": postMod1S2CRowsEqual(params, trace.FullStages[0], fullProdCombined)}
	result.Decomposition = map[string]interface{}{"upstream_input_propagation_n_f_minus_s_r": propagation, "fast_s2c_implementation_effect_s_f_minus_n_f": implementation, "observed_total_s_f_minus_s_r": observed, "closure_residual": closure, "implementation_relative_to_observed": ratioMetric(implementation.MaxComponent, observed.MaxComponent), "propagation_relative_to_observed": ratioMetric(propagation.MaxComponent, observed.MaxComponent)}

	inputDelta := make([]complex128, len(productionRealValues))
	for i := range inputDelta {
		inputDelta[i] = (productionRealValues[i] - matchedStandardRealValues[i]) + 1i*(productionImagValues[i]-matchedStandardImagValues[i])
	}
	deltaMetric := fix001P3S2CMetric(make([]complex128, len(inputDelta)), inputDelta, fix001P3P93S2CThreshold)
	deltaReal := make([]complex128, len(inputDelta))
	deltaImag := make([]complex128, len(inputDelta))
	for i, value := range inputDelta {
		deltaReal[i] = complex(real(value), 0)
		deltaImag[i] = complex(imag(value), 0)
	}
	dFastReal, err := evalModFastFromStandard(params, productionReal, deltaReal)
	if err != nil {
		return err
	}
	dFastImag, err := evalModFastFromStandard(params, productionImag, deltaImag)
	if err != nil {
		return err
	}
	dFullReal, _, err := postMod1S2CLiftFull(params, dFastReal)
	if err != nil {
		return err
	}
	dFullImag, _, err := postMod1S2CLiftFull(params, dFastImag)
	if err != nil {
		return err
	}
	deltaStages, _, err := fix001P3S2CTraceFull(params, profile.Fast.S2CDFTMatrix, profile.Standard.DFTEvaluator, dFullReal, dFullImag)
	if err != nil {
		return err
	}
	deltaOutput, err := fix001P3S2CDecode(params, deltaStages[len(deltaStages)-1], zeroSecret(params))
	if err != nil {
		return err
	}
	deltaAgreement := fix001P3S2CMetric(propVector, deltaOutput, 1e-10)
	result.LinearDelta = map[string]interface{}{"input_delta": deltaMetric, "output_delta": fix001P3S2CMetric(make([]complex128, len(deltaOutput)), deltaOutput, fix001P3P93S2CThreshold), "agreement_with_n_f_minus_s_r": deltaAgreement, "agreement_tolerance": 1e-10}
	for i, group := range trace.Groups {
		fastStage := trace.FastStages[i+1]
		fullStage := trace.FullStages[i+1]
		refStage := refTrace[i+1]
		fastValues, err := fix001P3S2CDecode(params, fastStage, zeroSecret(params))
		if err != nil {
			return err
		}
		fullValues, err := fix001P3S2CDecode(params, fullStage, zeroSecret(params))
		if err != nil {
			return err
		}
		refValues, err := fix001P3S2CDecode(params, refStage, profile.StandardSK)
		if err != nil {
			return err
		}
		inputMetric := fix001P3S2CMetric(refValues, fullValues, fix001P3P93S2CThreshold)
		implMetric := fix001P3S2CMetric(fullValues, fastValues, fix001P3P93S2CThreshold)
		previous := deltaMetric.MaxComponent
		if i > 0 {
			previous = result.Groups[i-1].InputPropagation.MaxComponent
		}
		amplification := 0.0
		if previous > 0 {
			amplification = inputMetric.MaxComponent / previous
		}
		levelEqual, scaleEqual, degreeEqual := fix001P3P93S2CMetadataEqual(fastStage, fullStage)
		result.Groups = append(result.Groups, fix001P3P93S2CGroupEvidence{Stage: fmt.Sprintf("group%d_post_rescale", group.Group), InputPropagation: inputMetric, ImplementationEffect: implMetric, Amplification: amplification, FastPreRowsMatch: group.PreRowsMatch, FastPostRowsMatch: group.PostRowsMatch, LevelEqual: levelEqual, ScaleEqual: scaleEqual, DegreeEqual: degreeEqual, FirstMaterial: implMetric.MaxComponent > 0.1*observed.MaxComponent})
	}
	standardPublic, err := profile.Standard.Bootstrap(reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	standardPublicValues, err := decodeWithSecret(profile.Residual, standardPublic, profile.StandardSK)
	if err != nil {
		return err
	}
	official, err := profile.Fast.Bootstrap(reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	officialValues, err := decodeWithSecret(profile.Residual, official, zeroSecret(profile.Residual))
	if err != nil {
		return err
	}
	officialMetric := fix001P3S2CMetric(standardPublicValues, officialValues, fix001P3P93S2CThreshold)
	result.Reconciliation = []map[string]interface{}{
		{"row": "accepted_p93_reference_evalmod_to_reference_s2c", "reference": "reproduced accepted P93 design proxy", "max_component_error": refPublic, "pass_vs_1e-2": refPublic != nil && refPublic.MaxComponent <= fix001P3P93S2CThreshold, "changed_from_previous": "P93 PS-wide design arithmetic and ordinary reference proxy"},
		{"row": "current_production_evalmod_to_reference_s2c", "reference": "matched P93 StandardFinal -> Standard S2C", "max_component_error": propagation, "pass_vs_1e-2": propagation.MaxComponent <= fix001P3P93S2CThreshold, "changed_from_previous": "only EvalMod semantics changed"},
		{"row": "current_production_evalmod_to_fast_production_s2c", "reference": "matched P93 StandardFinal -> Fast S2C", "max_component_error": observed, "pass_vs_1e-2": observed.MaxComponent <= fix001P3P93S2CThreshold, "changed_from_previous": "Fast S2C implementation added after controlling input"},
		{"row": "current_official_public_output", "reference": "genuine Standard Bootstrap", "max_component_error": officialMetric, "pass_vs_1e-2": officialMetric.MaxComponent <= fix001P3P93S2CThreshold, "changed_from_previous": "context only; finalization not modified"},
	}
	implRatio := ratioMetric(implementation.MaxComponent, observed.MaxComponent)
	propRatio := ratioMetric(propagation.MaxComponent, observed.MaxComponent)
	if refMirror.MaxComponent > 1e-10 || prodInputMirror.MaxComponent > 1e-10 {
		result.Classification, result.FirstBlocker = "P93_S2C_FULL_RNS_MIRROR_MISMATCH", "stage-aligned full-RNS mirror exceeded justified numerical floor"
	} else if closure.MaxComponent > 1e-10 || deltaAgreement.MaxComponent > 1e-10 {
		result.Classification, result.FirstBlocker = "P93_S2C_FULL_RNS_MIRROR_MISMATCH", "causal closure or linear-delta confirmation exceeded numerical floor"
	} else if implRatio > 0.1 && propRatio > 0.1 {
		result.Classification, result.FirstBlocker = "P93_MIXED_EVALMOD_AND_S2C_BLOCKERS", "both upstream propagation and Fast S2C effect are independently material"
	} else if implRatio > 0.1 || implementation.MaxComponent > fix001P3P93S2CThreshold {
		result.Classification, result.FirstBlocker = "P93_FAST_S2C_IMPLEMENTATION_DIVERGENCE", "first material Fast S2C implementation effect"
	} else {
		result.Classification, result.FirstBlocker = "P93_EVALMOD_ERROR_AMPLIFIED_BY_S2C", "upstream production EvalMod error dominates after linear S2C control"
	}
	result.Validation["reference_proxy_pass"] = refPublic.MaxComponent <= fix001P3P93S2CThreshold
	result.Validation["f0_boundary_reproduced"] = true
	result.Validation["mirror_fidelity_pass"] = refMirror.MaxComponent <= 1e-10 && prodInputMirror.MaxComponent <= 1e-10
	result.Validation["closure_pass"] = closure.MaxComponent <= 1e-10
	result.Validation["linear_delta_pass"] = deltaAgreement.MaxComponent <= 1e-10
	return fix001P3P93S2CWrite(result, outPath)
}

func ratioMetric(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
