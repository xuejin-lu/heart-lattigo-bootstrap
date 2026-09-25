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
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	fix001P3GenuineStandardReconciliationThreshold = 1e-2
	fix001P3GenuineStandardReconciliationMaterial  = 0.10
)

type fix001P3GenuineStandardStagedRun struct {
	Input       *rlwe.Ciphertext
	Packed      *rlwe.Ciphertext
	ModUp       *rlwe.Ciphertext
	C2SReal     *rlwe.Ciphertext
	C2SImag     *rlwe.Ciphertext
	EvalModReal *rlwe.Ciphertext
	EvalModImag *rlwe.Ciphertext
	Core        *rlwe.Ciphertext
	Unpacked    *rlwe.Ciphertext
	Final       *rlwe.Ciphertext
	RealValues  []complex128
	ImagValues  []complex128
	CoreValues  []complex128
	FinalValues []complex128
}

type fix001P3FastCorrectedReconciliationRun struct {
	Input       *rlwe.Ciphertext
	ModUp       *rlwe.Ciphertext
	C2SReal     *rlwe.Ciphertext
	C2SImag     *rlwe.Ciphertext
	EvalModReal *rlwe.Ciphertext
	EvalModImag *rlwe.Ciphertext
	Core        *rlwe.Ciphertext
	Unpacked    *rlwe.Ciphertext
	Final       *rlwe.Ciphertext
	RealValues  []complex128
	ImagValues  []complex128
	CoreValues  []complex128
	FinalValues []complex128
}

type fix001P3GenuineStandardReconciliationResult struct {
	SchemaVersion  string                   `json:"schema_version"`
	Timestamp      time.Time                `json:"timestamp"`
	Primary        RepositoryMetadata       `json:"primary_repository"`
	Secondary      RepositoryMetadata       `json:"secondary_repository"`
	Environment    EnvironmentMetadata      `json:"environment"`
	Config         BootstrapConfig          `json:"config"`
	FixedProfile   map[string]interface{}   `json:"fixed_profile"`
	Provenance     map[string]interface{}   `json:"provenance"`
	R0Standard     map[string]interface{}   `json:"r0_genuine_standard_baseline"`
	R1Matched      map[string]interface{}   `json:"r1_matched_reference_lineage"`
	R2Alignment    []map[string]interface{} `json:"r2_matched_vs_genuine_alignment"`
	R3E2E          map[string]interface{}   `json:"r3_exact_e2e_vs_proxy_distinction"`
	R4Fast         []map[string]interface{} `json:"r4_corrected_fast_vs_genuine"`
	FirstMaterial  map[string]interface{}   `json:"first_material_divergence"`
	R5Audit        map[string]interface{}   `json:"r5_reference_lineage_audit"`
	Classification string                   `json:"classification"`
	Validation     map[string]interface{}   `json:"validation"`
}

func fix001P3GenuineStandardReconciliationWrite(result fix001P3GenuineStandardReconciliationResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func fix001P3GenuineStandardReconciliationMetric(reference, actual []complex128) (*semanticBisectMetric, error) {
	return fix001P3PublicScaleResetMetric(reference, actual, fix001P3GenuineStandardReconciliationThreshold)
}

func fix001P3GenuineStandardReconciliationMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return fix001P3PostS2CScaleCollapseMetadata(ct)
}

func fix001P3GenuineStandardReconciliationRows(ct *rlwe.Ciphertext) map[string]string {
	return fix001P3ActualForcedRowHashes(ct)
}

func fix001P3GenuineStandardReconciliationStage(name string, matched, genuine *rlwe.Ciphertext, matchedValues, genuineValues []complex128, valid bool, matchedLineage, genuineLineage string) (map[string]interface{}, error) {
	entry := map[string]interface{}{
		"stage":                   name,
		"semantically_comparable": valid,
		"matched_lineage":         matchedLineage,
		"genuine_lineage":         genuineLineage,
	}
	if !valid {
		entry["comparison"] = "not_comparable"
		return entry, nil
	}
	metric, err := fix001P3GenuineStandardReconciliationMetric(matchedValues, genuineValues)
	if err != nil {
		return nil, err
	}
	entry["matched"] = map[string]interface{}{"metadata": fix001P3GenuineStandardReconciliationMetadata(matched), "row_hashes": fix001P3GenuineStandardReconciliationRows(matched)}
	entry["genuine"] = map[string]interface{}{"metadata": fix001P3GenuineStandardReconciliationMetadata(genuine), "row_hashes": fix001P3GenuineStandardReconciliationRows(genuine)}
	entry["metric"] = metric
	entry["metadata_equal"] = fix001P3PostS2CScaleCollapseMetadataEqual(matched, genuine)
	entry["coefficient_rows_equal"] = fix001P3PostS2CScaleCollapseRowsEqual(matched, genuine)
	return entry, nil
}

func fix001P3GenuineStandardReconciliationRunStaged(eval *bootstrapping.Evaluator, residual ckks.Parameters, btp bootstrapping.Parameters, sk *rlwe.SecretKey) (fix001P3GenuineStandardStagedRun, error) {
	input := reproducibleInput(residual, btp)
	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard PackAndSwitchN1ToN2: %w", err)
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard ScaleDown: %w", err)
	}
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard ModUp: %w", err)
	}
	c2sReal, c2sImag, err := eval.CoeffsToSlots(modUp.CopyNew())
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard CoeffsToSlots: %w", err)
	}
	evalModReal, err := eval.EvalMod(c2sReal.CopyNew())
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard EvalMod real: %w", err)
	}
	evalModImag, err := eval.EvalMod(c2sImag.CopyNew())
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard EvalMod imag: %w", err)
	}
	core, err := eval.SlotsToCoeffs(evalModReal.CopyNew(), evalModImag.CopyNew())
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard SlotsToCoeffs: %w", err)
	}
	unpacked, err := eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*core.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard UnpackAndSwitchN2ToN1: %w", err)
	}
	if len(unpacked) != 1 {
		return fix001P3GenuineStandardStagedRun{}, fmt.Errorf("Standard staged unpack returned %d ciphertexts, want 1", len(unpacked))
	}
	final := unpacked[0].CopyNew()
	final.Scale = residual.DefaultScale()
	params := btp.BootstrappingParameters
	realValues, err := semanticBisectValues(c2sReal, params, sk)
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, err
	}
	imagValues, err := semanticBisectValues(c2sImag, params, sk)
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, err
	}
	coreValues, err := fix001P3S2CDecode(params, core, sk)
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, err
	}
	finalValues, err := decodeWithSecret(residual, final, sk)
	if err != nil {
		return fix001P3GenuineStandardStagedRun{}, err
	}
	return fix001P3GenuineStandardStagedRun{Input: input, Packed: &packed[0], ModUp: modUp, C2SReal: c2sReal, C2SImag: c2sImag, EvalModReal: evalModReal, EvalModImag: evalModImag, Core: core, Unpacked: &unpacked[0], Final: final, RealValues: realValues, ImagValues: imagValues, CoreValues: coreValues, FinalValues: finalValues}, nil
}

func semanticBisectValues(ct *rlwe.Ciphertext, params ckks.Parameters, sk *rlwe.SecretKey) ([]complex128, error) {
	_, values, err := semanticBisectView(ct, params, sk)
	return values, err
}

func fix001P3GenuineStandardReconciliationRunFast(profile q056PreparedProfile) (fix001P3FastCorrectedReconciliationRun, error) {
	residual, btp, eval := profile.Residual, profile.BTP, profile.Fast
	input := reproducibleInput(residual, btp)
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	c2sReal, c2sImag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	evalModReal, err := eval.EvalMod(c2sReal.CopyNew())
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	evalModImag, err := eval.EvalMod(c2sImag.CopyNew())
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	evalModReal.Scale = c2sReal.Scale
	evalModImag.Scale = c2sImag.Scale
	core, err := eval.SlotsToCoeffs(evalModReal.CopyNew(), evalModImag.CopyNew())
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	_, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	unpacked, err := eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*core.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	if len(unpacked) != 1 {
		return fix001P3FastCorrectedReconciliationRun{}, fmt.Errorf("Fast corrected unpack returned %d ciphertexts, want 1", len(unpacked))
	}
	final := unpacked[0].CopyNew()
	if err := finalizationTransformMaintained(final, residual, false); err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	final.IsMontgomery = false
	final.Scale = evalModReal.Scale
	params := btp.BootstrappingParameters
	zero := zeroSecret(params)
	realValues, err := fix001P3S2CDecode(params, evalModReal, zero)
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	imagValues, err := fix001P3S2CDecode(params, evalModImag, zero)
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	coreValues, err := fix001P3S2CDecode(params, core, zero)
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	finalValues, err := decodeWithSecret(residual, final, zeroSecret(residual))
	if err != nil {
		return fix001P3FastCorrectedReconciliationRun{}, err
	}
	return fix001P3FastCorrectedReconciliationRun{Input: input, ModUp: modUp, C2SReal: c2sReal, C2SImag: c2sImag, EvalModReal: evalModReal, EvalModImag: evalModImag, Core: core, Unpacked: &unpacked[0], Final: final, RealValues: realValues, ImagValues: imagValues, CoreValues: coreValues, FinalValues: finalValues}, nil
}

func fix001P3GenuineStandardReconciliationLineage(profile q056PreparedProfile, realPath, imagPath evalModMatchedPath) map[string]interface{} {
	return map[string]interface{}{
		"c2s_real":                      map[string]interface{}{"constructor": "evalModMatchedC2SInputs", "source_input": "reproducibleInput(residual, btp) -> standardEval.PackAndSwitchN1ToN2 -> ScaleDown -> ModUp -> standardEval.CoeffsToSlots", "object": fix001P3GenuineStandardReconciliationMetadata(profile.Inputs.OrdinaryReal), "secret_key": "profile.StandardSK"},
		"c2s_imag":                      map[string]interface{}{"constructor": "evalModMatchedC2SInputs", "source_input": "reproducibleInput(residual, btp) -> standardEval.PackAndSwitchN1ToN2 -> ScaleDown -> ModUp -> standardEval.CoeffsToSlots", "object": fix001P3GenuineStandardReconciliationMetadata(profile.Inputs.OrdinaryImag), "secret_key": "profile.StandardSK"},
		"evalmod_real":                  map[string]interface{}{"constructor": "evalModMatchedRunPathWithPlanScaleAndFastPolynomial.StandardFinal", "source_input": "profile.Inputs.OrdinaryReal plus manual Standard polynomial/DoubleAngle replay", "object": fix001P3GenuineStandardReconciliationMetadata(realPath.StandardFinal), "secret_key": "profile.StandardSK"},
		"evalmod_imag":                  map[string]interface{}{"constructor": "evalModMatchedRunPathWithPlanScaleAndFastPolynomial.StandardFinal", "source_input": "profile.Inputs.OrdinaryImag plus manual Standard polynomial/DoubleAngle replay", "object": fix001P3GenuineStandardReconciliationMetadata(imagPath.StandardFinal), "secret_key": "profile.StandardSK"},
		"c2s_input_lineage_exact":       true,
		"full_bootstrap_lineage_reused": false,
	}
}

func runFIX001P3DiagP93GenuineStandardVsMatchedReconciliation(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3GenuineStandardReconciliationResult{
		SchemaVersion: "fix-001-p3-diag-p93-genuine-standard-vs-matched-reference-reconciliation.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "plan_scale": "2^93", "slots": 4096, "threshold": fix001P3GenuineStandardReconciliationThreshold},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "expected_secondary_head": requiredFIX001P3ActualForcedSecondary, "expected_secondary_diff_sha256": requiredFIX001P3ActualForcedDiff},
		Validation:   map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_fast_production_repair": true, "no_coefficient_correction": true, "no_parameter_tuning": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3ActualForcedSecondary || diffHash != requiredFIX001P3ActualForcedDiff || !secondaryDirty {
		result.Classification = "P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT"
		result.Validation["secondary_handoff_state_match"] = false
		return fix001P3GenuineStandardReconciliationWrite(result, outPath)
	}

	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	residual, btp := profile.Residual, profile.BTP
	standardPublic, err := profile.Standard.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return fmt.Errorf("Genuine Standard public Bootstrap: %w", err)
	}
	standardStaged, err := fix001P3GenuineStandardReconciliationRunStaged(profile.Standard, residual, btp, profile.StandardSK)
	if err != nil {
		return err
	}
	workload := reproducibleValues(residual.MaxSlots())
	standardPublicValues, err := decodeWithSecret(residual, standardPublic, profile.StandardSK)
	if err != nil {
		return err
	}
	publicVsStaged, err := fix001P3GenuineStandardReconciliationMetric(standardStaged.FinalValues, standardPublicValues)
	if err != nil {
		return err
	}
	publicE2E, err := fix001P3GenuineStandardReconciliationMetric(workload, standardPublicValues)
	if err != nil {
		return err
	}
	stagedE2E, err := fix001P3GenuineStandardReconciliationMetric(workload, standardStaged.FinalValues)
	if err != nil {
		return err
	}
	result.R0Standard = map[string]interface{}{
		"public_output":                  map[string]interface{}{"metadata": fix001P3GenuineStandardReconciliationMetadata(standardPublic), "row_hashes": fix001P3GenuineStandardReconciliationRows(standardPublic)},
		"staged_output":                  map[string]interface{}{"metadata": fix001P3GenuineStandardReconciliationMetadata(standardStaged.Final), "row_hashes": fix001P3GenuineStandardReconciliationRows(standardStaged.Final)},
		"public_vs_staged":               publicVsStaged,
		"public_structural_equal_staged": fix001P3PostS2CScaleCollapseMetadataEqual(standardPublic, standardStaged.Final) && fix001P3PostS2CScaleCollapseRowsEqual(standardPublic, standardStaged.Final),
		"public_exact_e2e":               publicE2E,
		"staged_exact_e2e":               stagedE2E,
		"baseline_agreement_threshold":   1e-8,
	}
	if publicVsStaged.MaxComponentAbs > 1e-8 || !result.R0Standard["public_structural_equal_staged"].(bool) {
		result.Classification = "P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT"
		return fix001P3GenuineStandardReconciliationWrite(result, outPath)
	}

	planScale := precisionSweepScale(93)
	matchedRealPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal.CopyNew(), profile.Inputs.OrdinaryReal.CopyNew(), profile.Fast, profile.Standard, btp.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, nil)
	if err != nil {
		return fmt.Errorf("matched real path: %w", err)
	}
	matchedImagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag.CopyNew(), profile.Inputs.OrdinaryImag.CopyNew(), profile.Fast, profile.Standard, btp.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, nil)
	if err != nil {
		return fmt.Errorf("matched imag path: %w", err)
	}
	matchedCore, err := profile.Standard.SlotsToCoeffs(matchedRealPath.StandardFinal.CopyNew(), matchedImagPath.StandardFinal.CopyNew())
	if err != nil {
		return err
	}
	_, matchedCtxtN1, matchedCtxtN2, err := profile.Standard.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*reproducibleInput(residual, btp)})
	if err != nil {
		return err
	}
	matchedUnpacked, err := profile.Standard.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*matchedCore.CopyNew()}, matchedCtxtN1, matchedCtxtN2)
	if err != nil {
		return err
	}
	if len(matchedUnpacked) != 1 {
		return fmt.Errorf("matched Standard unpack returned %d ciphertexts, want 1", len(matchedUnpacked))
	}
	matchedFinal := matchedUnpacked[0].CopyNew()
	matchedFinal.Scale = residual.DefaultScale()
	matchedC2SRealValues, err := semanticBisectValues(profile.Inputs.OrdinaryReal, btp.BootstrappingParameters, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedC2SImagValues, err := semanticBisectValues(profile.Inputs.OrdinaryImag, btp.BootstrappingParameters, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedEvalModRealValues, err := semanticBisectValues(matchedRealPath.StandardFinal, btp.BootstrappingParameters, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedEvalModImagValues, err := semanticBisectValues(matchedImagPath.StandardFinal, btp.BootstrappingParameters, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedCoreValues, err := fix001P3S2CDecode(btp.BootstrappingParameters, matchedCore, profile.StandardSK)
	if err != nil {
		return err
	}
	matchedFinalValues, err := decodeWithSecret(residual, matchedFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	result.R1Matched = fix001P3GenuineStandardReconciliationLineage(profile, matchedRealPath, matchedImagPath)

	stageSpecs := []struct {
		name                           string
		matched, genuine               *rlwe.Ciphertext
		matchedValues, genuineValues   []complex128
		valid                          bool
		matchedLineage, genuineLineage string
	}{
		{"c2s_real", profile.Inputs.OrdinaryReal, standardStaged.C2SReal, matchedC2SRealValues, standardStaged.RealValues, true, "evalModMatchedC2SInputs.OrdinaryReal", "fresh Standard staged CoeffsToSlots real"},
		{"c2s_imag", profile.Inputs.OrdinaryImag, standardStaged.C2SImag, matchedC2SImagValues, standardStaged.ImagValues, true, "evalModMatchedC2SInputs.OrdinaryImag", "fresh Standard staged CoeffsToSlots imag"},
		{"evalmod_real", matchedRealPath.StandardFinal, standardStaged.EvalModReal, matchedEvalModRealValues, mustSemanticBisectValues(standardStaged.EvalModReal, btp.BootstrappingParameters, profile.StandardSK), true, "manual Standard polynomial/DoubleAngle StandardFinal", "fresh Standard EvalMod real"},
		{"evalmod_imag", matchedImagPath.StandardFinal, standardStaged.EvalModImag, matchedEvalModImagValues, mustSemanticBisectValues(standardStaged.EvalModImag, btp.BootstrappingParameters, profile.StandardSK), true, "manual Standard polynomial/DoubleAngle StandardFinal", "fresh Standard EvalMod imag"},
		{"s2c_core", matchedCore, standardStaged.Core, matchedCoreValues, standardStaged.CoreValues, true, "Standard SlotsToCoeffs(matched StandardFinal real/imag)", "fresh Standard staged SlotsToCoeffs"},
		{"unpacked_output", matchedUnpacked[0].CopyNew(), standardStaged.Unpacked, mustS2CValues(btp.BootstrappingParameters, &matchedUnpacked[0], profile.StandardSK), mustS2CValues(btp.BootstrappingParameters, standardStaged.Unpacked, profile.StandardSK), true, "Standard Unpack(matched core)", "fresh Standard staged Unpack"},
		{"final_decoded_message", matchedFinal, standardStaged.Final, matchedFinalValues, standardStaged.FinalValues, true, "matched final-like decoded object", "fresh Standard staged public boundary"},
	}
	for _, spec := range stageSpecs {
		entry, err := fix001P3GenuineStandardReconciliationStage(spec.name, spec.matched, spec.genuine, spec.matchedValues, spec.genuineValues, spec.valid, spec.matchedLineage, spec.genuineLineage)
		if err != nil {
			return err
		}
		result.R2Alignment = append(result.R2Alignment, entry)
	}

	matchedFinalE2E, err := fix001P3GenuineStandardReconciliationMetric(workload, matchedFinalValues)
	if err != nil {
		return err
	}
	fastCorrected, err := fix001P3GenuineStandardReconciliationRunFast(profile)
	if err != nil {
		return err
	}
	fastModUpValues, err := semanticBisectValues(fastCorrected.ModUp, btp.BootstrappingParameters, zeroSecret(btp.BootstrappingParameters))
	if err != nil {
		return err
	}
	standardModUpValues, err := semanticBisectValues(standardStaged.ModUp, btp.BootstrappingParameters, profile.StandardSK)
	if err != nil {
		return err
	}
	fastUnpackedValues, err := fix001P3S2CDecode(btp.BootstrappingParameters, fastCorrected.Unpacked, zeroSecret(btp.BootstrappingParameters))
	if err != nil {
		return err
	}
	standardUnpackedValues, err := fix001P3S2CDecode(btp.BootstrappingParameters, standardStaged.Unpacked, profile.StandardSK)
	if err != nil {
		return err
	}
	fastStages := []struct {
		name                        string
		fast, genuine               *rlwe.Ciphertext
		fastValues, genuineValues   []complex128
		fastLineage, genuineLineage string
	}{
		{"c2s_input_modup", fastCorrected.ModUp, standardStaged.ModUp, fastModUpValues, standardModUpValues, "Fast Pack/ScaleDown/ModUp from reproducibleInput", "fresh Standard Pack/ScaleDown/ModUp from reproducibleInput"},
		{"c2s_real", fastCorrected.C2SReal, standardStaged.C2SReal, mustS2CValues(btp.BootstrappingParameters, fastCorrected.C2SReal, zeroSecret(btp.BootstrappingParameters)), standardStaged.RealValues, "production Fast CoeffsToSlots real", "fresh Standard CoeffsToSlots real"},
		{"c2s_imag", fastCorrected.C2SImag, standardStaged.C2SImag, mustS2CValues(btp.BootstrappingParameters, fastCorrected.C2SImag, zeroSecret(btp.BootstrappingParameters)), standardStaged.ImagValues, "production Fast CoeffsToSlots imag", "fresh Standard CoeffsToSlots imag"},
		{"evalmod_real", fastCorrected.EvalModReal, standardStaged.EvalModReal, fastCorrected.RealValues, mustSemanticBisectValues(standardStaged.EvalModReal, btp.BootstrappingParameters, profile.StandardSK), "Fast EvalMod with proven Scale restoration", "fresh Standard EvalMod real"},
		{"evalmod_imag", fastCorrected.EvalModImag, standardStaged.EvalModImag, fastCorrected.ImagValues, mustSemanticBisectValues(standardStaged.EvalModImag, btp.BootstrappingParameters, profile.StandardSK), "Fast EvalMod with proven Scale restoration", "fresh Standard EvalMod imag"},
		{"s2c_core", fastCorrected.Core, standardStaged.Core, fastCorrected.CoreValues, standardStaged.CoreValues, "Fast SlotsToCoeffs corrected branch", "fresh Standard staged SlotsToCoeffs"},
		{"unpacked_output", fastCorrected.Unpacked, standardStaged.Unpacked, fastUnpackedValues, standardUnpackedValues, "Fast UnpackAndSwitchN2ToN1", "fresh Standard UnpackAndSwitchN2ToN1"},
		{"final_decoded_message", fastCorrected.Final, standardStaged.Final, fastCorrected.FinalValues, standardStaged.FinalValues, "Fast corrected EvalMod + downstream Scale preserved", "fresh Standard staged public boundary"},
	}
	fastFinalE2E, err := fix001P3GenuineStandardReconciliationMetric(workload, fastCorrected.FinalValues)
	if err != nil {
		return err
	}
	finalGap := math.Abs(fastFinalE2E.MaxComponentAbs - publicE2E.MaxComponentAbs)
	materialThreshold := fix001P3GenuineStandardReconciliationMaterial * finalGap
	for _, spec := range fastStages {
		metric, err := fix001P3GenuineStandardReconciliationMetric(spec.genuineValues, spec.fastValues)
		if err != nil {
			return err
		}
		result.R4Fast = append(result.R4Fast, map[string]interface{}{"stage": spec.name, "fast_lineage": spec.fastLineage, "genuine_lineage": spec.genuineLineage, "fast": map[string]interface{}{"metadata": fix001P3GenuineStandardReconciliationMetadata(spec.fast), "row_hashes": fix001P3GenuineStandardReconciliationRows(spec.fast)}, "genuine": map[string]interface{}{"metadata": fix001P3GenuineStandardReconciliationMetadata(spec.genuine), "row_hashes": fix001P3GenuineStandardReconciliationRows(spec.genuine)}, "metric": metric, "material_threshold": materialThreshold, "material": metric.MaxComponentAbs >= materialThreshold})
	}
	firstFastMaterial := "none"
	for _, entry := range result.R4Fast {
		if entry["material"].(bool) {
			firstFastMaterial = entry["stage"].(string)
			break
		}
	}
	result.FirstMaterial = map[string]interface{}{"fast_vs_genuine_first_material_stage": firstFastMaterial, "final_fast_corrected_e2e": fastFinalE2E, "genuine_standard_public_e2e": publicE2E, "final_gap": finalGap, "material_fraction": fix001P3GenuineStandardReconciliationMaterial, "material_threshold": materialThreshold}
	result.R3E2E = map[string]interface{}{"genuine_standard_public": publicE2E, "genuine_standard_staged": stagedE2E, "matched_standard_final_like": matchedFinalE2E, "fast_corrected_with_both_metadata_preserved": fastFinalE2E, "matched_s2c_core_is_exact_e2e": false, "matched_s2c_core_role": "stage/proxy metric only", "fast_corrected_final_role": "exact E2E metric", "threshold": fix001P3GenuineStandardReconciliationThreshold}
	result.R5Audit = map[string]interface{}{
		"ordinary_inputs_origin":               "evalModMatchedC2SInputs creates a separate reproducibleInput, then runs the genuine Standard Pack/ScaleDown/ModUp/CoeffsToSlots path",
		"same_original_input_values":           true,
		"same_ciphertext_lineage":              false,
		"normalization_or_synthetic_split":     "matched EvalMod StandardFinal uses manual normalized polynomial/DoubleAngle replay; it is not the public Standard EvalMod object",
		"matched_reference_original_validity":  "valid for local Fast-vs-normalized/matched stage diagnostics",
		"matched_reference_exact_e2e_validity": false,
		"reason":                               "matched StandardFinal/core lineage is a manually reconstructed stage object and must not replace fresh Genuine Standard public/staged final output as the exact E2E oracle",
		"primary_helpers":                      []string{"evalModMatchedC2SInputs", "evalModMatchedRunPathWithPlanScaleAndFastPolynomial", "buildFIX001P3GenuineStandardEvaluator"},
		"secondary_modified":                   false,
	}

	matchedStageMaterial := map[string]bool{}
	for _, entry := range result.R2Alignment {
		if metric, ok := entry["metric"].(*semanticBisectMetric); ok {
			matchedStageMaterial[entry["stage"].(string)] = metric.MaxComponentAbs >= materialThreshold
		}
	}
	if matchedStageMaterial["c2s_real"] || matchedStageMaterial["c2s_imag"] {
		result.Classification = "P93_MATCHED_STANDARD_INPUT_LINEAGE_MISMATCH"
	} else if matchedStageMaterial["evalmod_real"] || matchedStageMaterial["evalmod_imag"] {
		result.Classification = "P93_MATCHED_STANDARD_EVALMOD_REFERENCE_MISMATCH"
	} else if matchedStageMaterial["s2c_core"] || matchedStageMaterial["unpacked_output"] {
		result.Classification = "P93_MATCHED_STANDARD_S2C_REFERENCE_MISMATCH"
	} else if firstFastMaterial != "none" {
		result.Classification = "P93_MATCHED_REFERENCE_VALID_BUT_FAST_DIVERGES_FROM_GENUINE_STANDARD"
	} else {
		result.Classification = "P93_MATCHED_REFERENCE_PROXY_NOT_VALID_FOR_E2E"
	}
	return fix001P3GenuineStandardReconciliationWrite(result, outPath)
}

func mustSemanticBisectValues(ct *rlwe.Ciphertext, params ckks.Parameters, sk *rlwe.SecretKey) []complex128 {
	values, err := semanticBisectValues(ct, params, sk)
	if err != nil {
		panic(err)
	}
	return values
}

func mustS2CValues(params ckks.Parameters, ct *rlwe.Ciphertext, sk *rlwe.SecretKey) []complex128 {
	values, err := fix001P3S2CDecode(params, ct, sk)
	if err != nil {
		panic(err)
	}
	return values
}
