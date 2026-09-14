package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3QuantizationAwarePrimary   = "dbb2e96fe2b6366d6870ccf827dda87668141feb"
	requiredFIX001P3QuantizationAwareSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	fix001P3QuantizationAwareClosureTolerance  = 1e-10
	fix001P3QuantizationAwarePublicThreshold   = 1e-2
	fix001P3QuantizationAwarePostS2CThreshold  = 3.125e-4
	fix001P3QuantizationAwareSourcePrecision   = uint(160)
)

type fix001P3QuantizationAwareResult struct {
	SchemaVersion         string                            `json:"schema_version"`
	Timestamp             time.Time                         `json:"timestamp"`
	Primary               RepositoryMetadata                `json:"primary_repository"`
	Lattigo               RepositoryMetadata                `json:"lattigo_repository"`
	Environment           EnvironmentMetadata               `json:"environment"`
	Config                BootstrapConfig                   `json:"config"`
	Parameters            ExperimentParameters              `json:"effective_parameters"`
	Provenance            map[string]interface{}            `json:"provenance"`
	Q0                    map[string]interface{}            `json:"q0_control"`
	Canonical             map[string]interface{}            `json:"canonical_ckks_oracle"`
	Causal                map[string]interface{}            `json:"five_component_evalmod_decomposition"`
	Closure               *PSGlobalMetric                   `json:"closure_residual"`
	Checkpoints           []map[string]interface{}          `json:"da_checkpoint_attribution"`
	Counterfactuals       map[string]map[string]interface{} `json:"counterfactuals"`
	Projected             map[string]interface{}            `json:"s2c_projected_components"`
	Classification        string                            `json:"classification"`
	FirstRemainingBlocker string                            `json:"first_remaining_blocker"`
	RequiredPostS2C       float64                           `json:"required_post_s2c_threshold"`
	PublicThreshold       float64                           `json:"public_like_threshold"`
	Validation            map[string]interface{}            `json:"validation"`
}

func fix001P3QuantizationAwareWrite(result fix001P3QuantizationAwareResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func fix001P3QuantizationAwareMaterialize(params ckks.Parameters, source *rlwe.Ciphertext, values []*bignum.Complex, precision uint) (*rlwe.Ciphertext, error) {
	plaintext := ckks.NewPlaintext(params, source.Level())
	*plaintext.MetaData = *source.MetaData
	plaintext.IsMontgomery = false
	if err := ckks.NewEncoder(params, precision).Encode(values, plaintext); err != nil {
		return nil, err
	}
	ciphertext := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*ciphertext.MetaData = *plaintext.MetaData
	for limb := range plaintext.Value.Coeffs {
		copy(ciphertext.Value[0].Coeffs[limb], plaintext.Value.Coeffs[limb])
	}
	return ciphertext, nil
}

func fix001P3QuantizationAwareValues(values []complex128, precision uint) []*bignum.Complex {
	precise := make([]*bignum.Complex, len(values))
	for i, value := range values {
		precise[i] = bignum.ToComplex(value, precision)
	}
	return precise
}

func fix001P3QuantizationAwareMetric(reference, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(reference, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func fix001P3QuantizationAwareSub(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

func fix001P3QuantizationAwareAdd(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func fix001P3QuantizationAwareCombine(realValues, imagValues []complex128) []complex128 {
	out := make([]complex128, len(realValues))
	for i := range out {
		out[i] = realValues[i] + 1i*imagValues[i]
	}
	return out
}

func fix001P3QuantizationAwareDecode(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	values, err := fix001P3HighPrecisionOracleDecodeFull(params, ct, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, err
	}
	return fix001P3HighPrecisionToComplex(values), nil
}

func fix001P3QuantizationAwareFullS2C(profile q056PreparedProfile, realCT, imagCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	params := profile.BTP.BootstrappingParameters
	stages, _, err := fix001P3S2CTraceFull(params, profile.Fast.S2CDFTMatrix, profile.Standard.DFTEvaluator, realCT, imagCT)
	if err != nil {
		return nil, err
	}
	return stages[len(stages)-1], nil
}

func fix001P3QuantizationAwareFinalize(profile q056PreparedProfile, input, core *rlwe.Ciphertext) ([]complex128, error) {
	coreForFast := core.CopyNew()
	if !coreForFast.IsMontgomery {
		maintained := coreForFast.Level() + 1
		if maintained > 2 {
			maintained = 2
		}
		for component := range coreForFast.Value {
			for limb := 0; limb < maintained; limb++ {
				profile.BTP.BootstrappingParameters.RingQ().SubRings[limb].MForm(coreForFast.Value[component].Coeffs[limb], coreForFast.Value[component].Coeffs[limb])
			}
		}
		coreForFast.IsMontgomery = true
	}
	_, ctxtN1, ctxtN2, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, err
	}
	unpacked, err := profile.Fast.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*coreForFast}, ctxtN1, ctxtN2)
	if err != nil {
		return nil, err
	}
	if len(unpacked) != 1 {
		return nil, fmt.Errorf("expected one unpacked ciphertext, got %d", len(unpacked))
	}
	_, final, err := fix001P3E2EFinalizeOracle(&unpacked[0], profile.BTP)
	if err != nil {
		return nil, err
	}
	return decodeWithSecret(profile.Residual, final, zeroSecret(profile.Residual))
}

func fix001P3QuantizationAwareDownstream(profile q056PreparedProfile, baselineCore, baselineEvalMod []complex128, input, realCT, imagCT *rlwe.Ciphertext, evalModValues []complex128) (map[string]interface{}, *rlwe.Ciphertext, error) {
	params := profile.BTP.BootstrappingParameters
	core, err := fix001P3QuantizationAwareFullS2C(profile, realCT, imagCT)
	if err != nil {
		return nil, nil, err
	}
	coreValues, err := fix001P3QuantizationAwareDecode(params, core)
	if err != nil {
		return nil, nil, err
	}
	finalValues, err := fix001P3QuantizationAwareFinalize(profile, input, core)
	if err != nil {
		return nil, nil, err
	}
	return map[string]interface{}{
		"evalmod_vs_standard":  fix001P3QuantizationAwareMetric(baselineEvalMod, evalModValues, fix001P3QuantizationAwarePublicThreshold),
		"post_s2c_vs_standard": fix001P3QuantizationAwareMetric(baselineCore, coreValues, fix001P3QuantizationAwarePostS2CThreshold),
		"public_like":          fix001P3QuantizationAwareMetric(reproducibleValues(profile.Residual.MaxSlots()), finalValues, fix001P3QuantizationAwarePublicThreshold),
		"final_metadata_pass":  core.Level() >= 0,
	}, core, nil
}

func fix001P3QuantizationAwareProject(profile q056PreparedProfile, realTemplate, imagTemplate *rlwe.Ciphertext, vector []complex128) ([]complex128, *rlwe.Ciphertext, error) {
	realValues := make([]complex128, len(vector))
	imagValues := make([]complex128, len(vector))
	for i, value := range vector {
		realValues[i] = complex(real(value), 0)
		imagValues[i] = complex(imag(value), 0)
	}
	realCT, err := fix001P3QuantizationAwareMaterialize(profile.BTP.BootstrappingParameters, realTemplate, fix001P3QuantizationAwareValues(realValues, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, nil, err
	}
	imagCT, err := fix001P3QuantizationAwareMaterialize(profile.BTP.BootstrappingParameters, imagTemplate, fix001P3QuantizationAwareValues(imagValues, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, nil, err
	}
	core, err := fix001P3QuantizationAwareFullS2C(profile, realCT, imagCT)
	if err != nil {
		return nil, nil, err
	}
	values, err := fix001P3QuantizationAwareDecode(profile.BTP.BootstrappingParameters, core)
	if err != nil {
		return nil, nil, err
	}
	return values, core, nil
}

func fix001P3QuantizationAwareCheckpointTable(params ckks.Parameters, fastRun, qRun fix001P3EvalModResidualRun, idealQ, idealO map[string][]complex128) ([]map[string]interface{}, error) {
	rows := make([]map[string]interface{}, 0, 15)
	for round := 0; round < 3; round++ {
		for _, checkpoint := range []string{"input", "square", "after_multiplier", "after_constant", "post_rescale"} {
			key := fmt.Sprintf("round%d.%s", round, checkpoint)
			fastStage, ok := fastRun.Stages[key]
			if !ok {
				return nil, fmt.Errorf("missing Fast stage %s", key)
			}
			qStage, ok := qRun.Stages[key]
			if !ok {
				return nil, fmt.Errorf("missing canonical stage %s", key)
			}
			nf, err := fix001P3EvalModResidualDecode(params, fastStage.Normalized)
			if err != nil {
				return nil, err
			}
			nq, err := fix001P3QuantizationAwareDecode(params, qStage.Normalized)
			if err != nil {
				return nil, err
			}
			row := map[string]interface{}{
				"round": round, "checkpoint": checkpoint,
				"ps_arithmetic_excess_b_nf_minus_nq": fix001P3QuantizationAwareMetric(make([]complex128, len(nf)), fix001P3QuantizationAwareSub(nf, nq), fix001P3QuantizationAwarePublicThreshold),
				"da_ckks_arithmetic_c_nq_minus_iq":   fix001P3QuantizationAwareMetric(make([]complex128, len(nq)), fix001P3QuantizationAwareSub(nq, idealQ[key]), fix001P3QuantizationAwarePublicThreshold),
				"representation_d_iq_minus_io":       fix001P3QuantizationAwareMetric(make([]complex128, len(nq)), fix001P3QuantizationAwareSub(idealQ[key], idealO[key]), fix001P3QuantizationAwarePublicThreshold),
				"accepted_contracts":                 fix001P3EvalModResidualRowsAndCapacity(params, fastStage),
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func fix001P3QuantizationAwareClassification(cases map[string]map[string]interface{}, implementation *PSGlobalMetric) (string, string) {
	if implementation.MaxComponent > fix001P3QuantizationAwareClosureTolerance {
		return "logn13_evalmod_residual_fast_implementation_mismatch", "A = E_F-N_F is materially nonzero despite accepted exact contracts"
	}
	psPass := cases["C_PS"]["public_like"].(*PSGlobalMetric).Pass
	daPass := cases["C_DA"]["public_like"].(*PSGlobalMetric).Pass
	bothCKKSPass := cases["C_BOTH_CKKS"]["public_like"].(*PSGlobalMetric).Pass
	mathPass := cases["C_MATH"]["public_like"].(*PSGlobalMetric).Pass
	if psPass && !daPass {
		return "logn13_evalmod_residual_ps_input_blocker", "C_PS passes while C_DA fails"
	}
	if daPass && !psPass {
		return "logn13_evalmod_residual_da_arithmetic_blocker", "C_DA passes while C_PS fails"
	}
	if psPass && daPass {
		return "logn13_evalmod_residual_multiple_single_fix_options", "C_PS and C_DA both pass"
	}
	if !psPass && !daPass && bothCKKSPass {
		return "logn13_evalmod_residual_joint_ps_da_blocker", "C_PS and C_DA fail while C_BOTH_CKKS passes"
	}
	if !bothCKKSPass && mathPass {
		return "logn13_evalmod_residual_ckks_representation_floor", "C_BOTH_CKKS fails while C_MATH passes"
	}
	return "logn13_evalmod_residual_mathematical_chain_floor", "C_MATH fails"
}

func fix001P3QuantizationAwareWorstContribution(projected map[string][]complex128, index int, part string) map[string]interface{} {
	result := map[string]interface{}{"index": index, "component": part}
	for name, values := range projected {
		if index < 0 || index >= len(values) {
			continue
		}
		if part == "imag" {
			result[name] = imag(values[index])
		} else {
			result[name] = real(values[index])
		}
	}
	return result
}

func runFIX001P3DiagLogN13QuantizationAwareEvalModDecomposition(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3QuantizationAwarePrimary, "HEAD") == requiredFIX001P3QuantizationAwarePrimary
	result := fix001P3QuantizationAwareResult{
		SchemaVersion: "fix-001-p3-diag-logn13-quantization-aware-evalmod-decomposition.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		RequiredPostS2C: fix001P3QuantizationAwarePostS2CThreshold, PublicThreshold: fix001P3QuantizationAwarePublicThreshold,
		Provenance: map[string]interface{}{"primary_required_base": requiredFIX001P3QuantizationAwarePrimary, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_required_commit": requiredFIX001P3QuantizationAwareSecondary, "secondary_commit": secondaryCommit, "secondary_branch": secondaryBranch, "secondary_clean": secondaryClean},
		Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_production_changes": true, "no_candidate_retuning": true, "no_q_parameter_changes": true, "no_s2c_changes": true, "no_generated_power_redesign": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
	}
	if cfg.LogN != 13 || !primaryBaseAncestor || secondaryCommit != requiredFIX001P3QuantizationAwareSecondary || secondaryBranch != "fast-ckks" || !secondaryClean {
		result.Classification = "logn13_quantization_aware_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = "startup provenance or Secondary precondition"
		return fix001P3QuantizationAwareWrite(result, outPath)
	}

	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	acceptedRow, candidate, err := fix001P3G0RunCandidate(profile, 2)
	if err != nil {
		return err
	}
	if candidate.realFinal == nil || candidate.imagFinal == nil {
		result.Classification = "logn13_quantization_aware_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = "accepted candidate did not produce both P_F branches"
		return fix001P3QuantizationAwareWrite(result, outPath)
	}
	realPath, imagPath, err := fix001P3S2CAcceptedPaths(profile, candidate)
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	realFinal, err := evalModMatchedFinalEvidence(realPath, params, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, params, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realPath, imagPath, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	metadataPass, _ := publicEvidence.Metadata["pass"].(bool)
	q0Contracts := acceptedRow.ValidLocalQ2 && realPath.FirstFailure == "none" && imagPath.FirstFailure == "none" && realFinal.RestoreRowsMatch && imagFinal.RestoreRowsMatch && realFinal.NormalizedBeforeRestoreCapacity != nil && realFinal.NormalizedBeforeRestoreCapacity.Pass && imagFinal.NormalizedBeforeRestoreCapacity != nil && imagFinal.NormalizedBeforeRestoreCapacity.Pass && realFinal.FinalRestoreCapacity != nil && realFinal.FinalRestoreCapacity.Pass && imagFinal.FinalRestoreCapacity != nil && imagFinal.FinalRestoreCapacity.Pass && metadataPass && publicEvidence.InputUnchanged
	result.Q0 = map[string]interface{}{"executed_candidate": map[string]interface{}{"name": "G0-local-q2-2-F0-local-q2-3", "g0_guard_bits": 2, "f0_guard_bits": 3, "all_da_rounds_local_q2": true, "real_first_failure": realPath.FirstFailure, "imag_first_failure": imagPath.FirstFailure}, "evalmod_real_vs_standard": realFinal.FastVsStandard, "evalmod_imag_vs_standard": imagFinal.FastVsStandard, "post_s2c_vs_standard": publicEvidence.PostS2C, "public_like": publicEvidence.PublicLike, "final_metadata": publicEvidence.Metadata, "exact_contracts": q0Contracts}
	result.Validation["q0_control_pass"] = q0Contracts
	if !q0Contracts || publicEvidence.PublicLike == nil || !fix001P3EvalModResidualWithin(realFinal.FastVsStandard.MaxComponent, 6.261306805777143e-5, 2e-5) || !fix001P3EvalModResidualWithin(imagFinal.FastVsStandard.MaxComponent, 4.647793870173522e-5, 2e-5) || !fix001P3EvalModResidualWithin(publicEvidence.PostS2C.MaxComponent, 3.3385298628089955e-4, 2e-5) {
		result.Classification = "logn13_quantization_aware_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = "Q0 accepted control mismatch"
		return fix001P3QuantizationAwareWrite(result, outPath)
	}

	pfReal, err := fix001P3EvalModResidualDecode(params, candidate.realFinal)
	if err != nil {
		return err
	}
	pfImag, err := fix001P3EvalModResidualDecode(params, candidate.imagFinal)
	if err != nil {
		return err
	}
	inputReal, err := fix001P3HighPrecisionPreprocessedInput(params, profile.Fast, profile.Inputs.FastReal, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return err
	}
	inputImag, err := fix001P3HighPrecisionPreprocessedInput(params, profile.Fast, profile.Inputs.FastImag, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return err
	}
	poly := profile.Fast.Mod1Parameters.Mod1Poly
	poRealHP := fix001P3HighPrecisionChebyshevOracle(poly, inputReal, fix001P3QuantizationAwareSourcePrecision)
	poImagHP := fix001P3HighPrecisionChebyshevOracle(poly, inputImag, fix001P3QuantizationAwareSourcePrecision)
	qReal, err := fix001P3QuantizationAwareMaterialize(params, candidate.realFinal, poRealHP, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return err
	}
	qImag, err := fix001P3QuantizationAwareMaterialize(params, candidate.imagFinal, poImagHP, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return err
	}
	qReal80, err := fix001P3QuantizationAwareMaterialize(params, candidate.realFinal, poRealHP, 80)
	if err != nil {
		return err
	}
	qImag80, err := fix001P3QuantizationAwareMaterialize(params, candidate.imagFinal, poImagHP, 80)
	if err != nil {
		return err
	}
	qReal160Hash := fix001P3HighPrecisionOracleHash(qReal)
	qImag160Hash := fix001P3HighPrecisionOracleHash(qImag)
	qStable := qReal160Hash == fix001P3HighPrecisionOracleHash(qReal80) && qImag160Hash == fix001P3HighPrecisionOracleHash(qImag80)
	qRealHP, err := fix001P3HighPrecisionOracleDecodeFull(params, qReal, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return err
	}
	qImagHP, err := fix001P3HighPrecisionOracleDecodeFull(params, qImag, fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return err
	}
	poReal, poImag := fix001P3HighPrecisionToComplex(poRealHP), fix001P3HighPrecisionToComplex(poImagHP)
	qRealValues, qImagValues := fix001P3HighPrecisionToComplex(qRealHP), fix001P3HighPrecisionToComplex(qImagHP)
	qRealQuant := fix001P3HighPrecisionMetric(poRealHP, qRealHP, fix001P3QuantizationAwareClosureTolerance)
	qImagQuant := fix001P3HighPrecisionMetric(poImagHP, qImagHP, fix001P3QuantizationAwareClosureTolerance)
	realReporting := fix001P3HighPrecisionReportingFloor(qRealHP, fix001P3QuantizationAwareSourcePrecision)
	imagReporting := fix001P3HighPrecisionReportingFloor(qImagHP, fix001P3QuantizationAwareSourcePrecision)
	pFRealQ := fix001P3QuantizationAwareMetric(qRealValues, pfReal, fix001P3QuantizationAwareClosureTolerance)
	pFImagQ := fix001P3QuantizationAwareMetric(qImagValues, pfImag, fix001P3QuantizationAwareClosureTolerance)
	pFRealO := fix001P3QuantizationAwareMetric(poReal, pfReal, fix001P3QuantizationAwareClosureTolerance)
	pFImagO := fix001P3QuantizationAwareMetric(poImag, pfImag, fix001P3QuantizationAwareClosureTolerance)
	delta := candidate.realFinal.Scale.Float64()
	qUnit := 1 / delta
	theoretical := math.Sqrt(float64(candidate.realFinal.N())) * qUnit
	measured := math.Max(qRealQuant.MaxComponent, qImagQuant.MaxComponent)
	logicalMeta := fix001P3HighPrecisionMetadataMatch(candidate.realFinal, qReal)
	result.Canonical = map[string]interface{}{"source_precision_bits": fix001P3QuantizationAwareSourcePrecision, "encoder_precision_bits": fix001P3QuantizationAwareSourcePrecision, "p_q_minus_p_o_real": qRealQuant, "p_q_minus_p_o_imag": qImagQuant, "p_f_minus_p_q_real": pFRealQ, "p_f_minus_p_q_imag": pFImagQ, "p_f_minus_p_o_real": pFRealO, "p_f_minus_p_o_imag": pFImagO, "reporting_floor_real": realReporting, "reporting_floor_imag": imagReporting, "one_over_delta": qUnit, "sqrt_n_over_delta": theoretical, "measured_over_sqrt_n_over_delta": measured / theoretical, "logical_metadata_match": logicalMeta, "materialization_deterministic": qStable, "encoder_80_160_integer_state_stable": qStable, "p_q_real_hash": qReal160Hash, "p_q_imag_hash": qImag160Hash}
	if !qStable {
		result.Classification = "logn13_canonical_ckks_oracle_materialization_unstable"
		result.FirstRemainingBlocker = "80-bit and 160-bit encoder materializations differ"
		return fix001P3QuantizationAwareWrite(result, outPath)
	}

	fastRealRun, err := fix001P3EvalModResidualRunState(profile, profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, candidate.realFinal, candidate)
	if err != nil {
		return err
	}
	fastImagRun, err := fix001P3EvalModResidualRunState(profile, profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, candidate.imagFinal, candidate)
	if err != nil {
		return err
	}
	qRealRun, err := fix001P3EvalModResidualFullNormalizedRun(params, qReal, fastRealRun.WorkingScale, fastRealRun.InitialK, fastRealRun.InitialSqrt, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	qImagRun, err := fix001P3EvalModResidualFullNormalizedRun(params, qImag, fastImagRun.WorkingScale, fastImagRun.InitialK, fastImagRun.InitialSqrt, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealQRealStages, idealQReal, err := fix001P3EvalModResidualIdealNormalized(qRealValues, fastRealRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealQImagStages, idealQImag, err := fix001P3EvalModResidualIdealNormalized(qImagValues, fastImagRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealORealStages, idealOReal, err := fix001P3EvalModResidualIdealNormalized(poReal, fastRealRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealOImagStages, idealOImag, err := fix001P3EvalModResidualIdealNormalized(poImag, fastImagRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	_, idealFReal, err := fix001P3EvalModResidualIdealNormalized(pfReal, fastRealRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	_, idealFImag, err := fix001P3EvalModResidualIdealNormalized(pfImag, fastImagRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealQStages := map[string][]complex128{}
	idealOStages := map[string][]complex128{}
	for key, value := range idealQRealStages {
		idealQStages[key] = fix001P3QuantizationAwareCombine(value, idealQImagStages[key])
	}
	for key, value := range idealORealStages {
		idealOStages[key] = fix001P3QuantizationAwareCombine(value, idealOImagStages[key])
	}
	result.Checkpoints, err = fix001P3QuantizationAwareCheckpointTable(params, fastRealRun, qRealRun, idealQRealStages, idealORealStages)
	if err != nil {
		return err
	}
	imagRows, err := fix001P3QuantizationAwareCheckpointTable(params, fastImagRun, qImagRun, idealQImagStages, idealOImagStages)
	if err != nil {
		return err
	}
	for _, row := range result.Checkpoints {
		row["branch"] = "real"
	}
	for _, row := range imagRows {
		row["branch"] = "imag"
		result.Checkpoints = append(result.Checkpoints, row)
	}

	efReal, err := fix001P3EvalModResidualDecode(params, realPath.FastFinal)
	if err != nil {
		return err
	}
	efImag, err := fix001P3EvalModResidualDecode(params, imagPath.FastFinal)
	if err != nil {
		return err
	}
	nfReal, err := fix001P3EvalModResidualDecode(params, fastRealRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	nfImag, err := fix001P3EvalModResidualDecode(params, fastImagRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	nqReal, err := fix001P3QuantizationAwareDecode(params, qRealRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	nqImag, err := fix001P3QuantizationAwareDecode(params, qImagRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	_, esReal, err := semanticBisectView(realPath.StandardFinal, params, profile.StandardSK)
	if err != nil {
		return err
	}
	_, esImag, err := semanticBisectView(imagPath.StandardFinal, params, profile.StandardSK)
	if err != nil {
		return err
	}
	ef := fix001P3QuantizationAwareCombine(efReal, efImag)
	nf := fix001P3QuantizationAwareCombine(nfReal, nfImag)
	nq := fix001P3QuantizationAwareCombine(nqReal, nqImag)
	iq := fix001P3QuantizationAwareCombine(idealQReal, idealQImag)
	io := fix001P3QuantizationAwareCombine(idealOReal, idealOImag)
	es := fix001P3QuantizationAwareCombine(esReal, esImag)
	a := fix001P3QuantizationAwareSub(ef, nf)
	b := fix001P3QuantizationAwareSub(nf, nq)
	c := fix001P3QuantizationAwareSub(nq, iq)
	d := fix001P3QuantizationAwareSub(iq, io)
	e := fix001P3QuantizationAwareSub(io, es)
	total := fix001P3QuantizationAwareSub(ef, es)
	closure := fix001P3QuantizationAwareSub(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(a, b), c), d), e), total)
	result.Closure = fix001P3QuantizationAwareMetric(make([]complex128, len(closure)), closure, fix001P3QuantizationAwareClosureTolerance)
	result.Causal = map[string]interface{}{"A_fast_implementation_EF_minus_NF": fix001P3QuantizationAwareMetric(make([]complex128, len(a)), a, fix001P3QuantizationAwarePublicThreshold), "B_ps_excess_NF_minus_NQ": fix001P3QuantizationAwareMetric(make([]complex128, len(b)), b, fix001P3QuantizationAwarePublicThreshold), "C_da_ckks_NQ_minus_IQ": fix001P3QuantizationAwareMetric(make([]complex128, len(c)), c, fix001P3QuantizationAwarePublicThreshold), "D_representation_IQ_minus_IO": fix001P3QuantizationAwareMetric(make([]complex128, len(d)), d, fix001P3QuantizationAwarePublicThreshold), "E_reference_IO_minus_ES": fix001P3QuantizationAwareMetric(make([]complex128, len(e)), e, fix001P3QuantizationAwarePublicThreshold), "T_observed_EF_minus_ES": fix001P3QuantizationAwareMetric(make([]complex128, len(total)), total, fix001P3QuantizationAwarePublicThreshold)}

	baselineCoreCT, err := profile.Standard.SlotsToCoeffs(realPath.StandardFinal.CopyNew(), imagPath.StandardFinal.CopyNew())
	if err != nil {
		return err
	}
	baselineCore, err := decodeWithSecret(params, baselineCoreCT, profile.StandardSK)
	if err != nil {
		return err
	}
	baselineEvalMod := es
	input := reproducibleInput(profile.Residual, profile.BTP)
	cases := map[string]map[string]interface{}{}
	actualEvalMetric := fix001P3QuantizationAwareMetric(es, ef, fix001P3QuantizationAwarePublicThreshold)
	cases["C0"] = map[string]interface{}{"evalmod_vs_standard": actualEvalMetric, "post_s2c_vs_standard": publicEvidence.PostS2C, "public_like": publicEvidence.PublicLike, "final_metadata_pass": publicEvidence.Metadata["pass"] == true}
	caseDefs := []struct {
		name      string
		realCT    *rlwe.Ciphertext
		imagCT    *rlwe.Ciphertext
		evalValue []complex128
	}{
		{"C_PS", qRealRun.State.Path.NormalizedFinal, qImagRun.State.Path.NormalizedFinal, nq},
	}
	for _, def := range caseDefs {
		caseResult, _, err := fix001P3QuantizationAwareDownstream(profile, baselineCore, baselineEvalMod, input, def.realCT, def.imagCT, def.evalValue)
		if err != nil {
			return err
		}
		cases[def.name] = caseResult
	}
	idealCases := []struct {
		name      string
		realValue []complex128
		imagValue []complex128
		evalValue []complex128
	}{
		{"C_DA", idealFReal, idealFImag, fix001P3QuantizationAwareCombine(idealFReal, idealFImag)},
		{"C_BOTH_CKKS", idealQReal, idealQImag, iq},
		{"C_MATH", idealOReal, idealOImag, io},
	}
	for _, def := range idealCases {
		realCT, err := fix001P3QuantizationAwareMaterialize(params, qRealRun.State.Path.NormalizedFinal, fix001P3QuantizationAwareValues(def.realValue, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
		if err != nil {
			return err
		}
		imagCT, err := fix001P3QuantizationAwareMaterialize(params, qImagRun.State.Path.NormalizedFinal, fix001P3QuantizationAwareValues(def.imagValue, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
		if err != nil {
			return err
		}
		caseResult, _, err := fix001P3QuantizationAwareDownstream(profile, baselineCore, baselineEvalMod, input, realCT, imagCT, def.evalValue)
		if err != nil {
			return err
		}
		cases[def.name] = caseResult
	}
	result.Counterfactuals = cases

	componentVectors := map[string][]complex128{"A": a, "B": b, "C": c, "D": d, "E": e}
	projectedValues := map[string][]complex128{}
	projectedEvidence := map[string]interface{}{}
	for name, vector := range componentVectors {
		values, _, err := fix001P3QuantizationAwareProject(profile, realPath.FastFinal, imagPath.FastFinal, vector)
		if err != nil {
			return err
		}
		projectedValues[name] = values
		projectedEvidence[name] = fix001P3QuantizationAwareMetric(make([]complex128, len(values)), values, fix001P3QuantizationAwarePostS2CThreshold)
	}
	totalProjected, totalCore, err := fix001P3QuantizationAwareProject(profile, realPath.FastFinal, imagPath.FastFinal, total)
	if err != nil {
		return err
	}
	projectedClosure := fix001P3QuantizationAwareSub(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(projectedValues["A"], projectedValues["B"]), projectedValues["C"]), projectedValues["D"]), projectedValues["E"]), totalProjected)
	worstIndex, worstPart := publicEvidence.PublicLike.WorstIndex, publicEvidence.PublicLike.WorstPart
	projectedEvidence["closure_after_s2c"] = fix001P3QuantizationAwareMetric(make([]complex128, len(projectedClosure)), projectedClosure, fix001P3QuantizationAwareClosureTolerance)
	projectedEvidence["contribution_at_actual_worst_public_slot"] = fix001P3QuantizationAwareWorstContribution(projectedValues, worstIndex, worstPart)
	actualValues, actualCore, err := fix001P3QuantizationAwareProject(profile, realPath.FastFinal, imagPath.FastFinal, ef)
	if err != nil {
		return err
	}
	actualFinalValues, err := fix001P3QuantizationAwareFinalize(profile, input, actualCore)
	if err != nil {
		return err
	}
	removedEvidence := map[string]interface{}{}
	for name, componentVector := range componentVectors {
		remainingVector := fix001P3QuantizationAwareSub(ef, componentVector)
		_, removedCore, err := fix001P3QuantizationAwareProject(profile, realPath.FastFinal, imagPath.FastFinal, remainingVector)
		if err != nil {
			return err
		}
		removedValues, err := fix001P3QuantizationAwareFinalize(profile, input, removedCore)
		if err != nil {
			return err
		}
		removedEvidence[name] = map[string]interface{}{
			"public_like_if_removed": fix001P3QuantizationAwareMetric(reproducibleValues(profile.Residual.MaxSlots()), removedValues, fix001P3QuantizationAwarePublicThreshold),
			"removed_core_level":     removedCore.Level(),
		}
	}
	removedEvidence["actual_public_like"] = fix001P3QuantizationAwareMetric(reproducibleValues(profile.Residual.MaxSlots()), actualFinalValues, fix001P3QuantizationAwarePublicThreshold)
	removedEvidence["actual_post_s2c"] = fix001P3QuantizationAwareMetric(make([]complex128, len(actualValues)), actualValues, fix001P3QuantizationAwarePostS2CThreshold)
	removedEvidence["total_projection_closure_core_level"] = totalCore.Level()
	projectedEvidence["public_like_if_component_removed_individually"] = removedEvidence
	result.Projected = projectedEvidence
	result.Validation["closure_pass"] = result.Closure.MaxComponent <= fix001P3QuantizationAwareClosureTolerance
	result.Validation["s2c_projected_closure_pass"] = projectedEvidence["closure_after_s2c"].(*PSGlobalMetric).MaxComponent <= fix001P3QuantizationAwareClosureTolerance
	result.Validation["materialization_reporting_floor_pass"] = realReporting.MaxComponent <= 1e-12 && imagReporting.MaxComponent <= 1e-12
	result.Classification, result.FirstRemainingBlocker = fix001P3QuantizationAwareClassification(cases, result.Causal["A_fast_implementation_EF_minus_NF"].(*PSGlobalMetric))
	result.Validation["q2_exact_contracts"] = fastRealRun.State.Path.FastRowsMatch && fastRealRun.State.Path.RestoreRowsMatch && fastImagRun.State.Path.FastRowsMatch && fastImagRun.State.Path.RestoreRowsMatch
	return fix001P3QuantizationAwareWrite(result, outPath)
}
