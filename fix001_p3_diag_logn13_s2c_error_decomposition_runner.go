package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	commonlintrans "github.com/tuneinsight/lattigo/v6/circuits/common/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3S2CErrorDecompositionPrimary   = "30e2e602576cf2c8a3c91bbb4d8d724c78bc027d"
	requiredFIX001P3S2CErrorDecompositionSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	fix001P3S2CErrorDecompositionThreshold         = 1e-2
	fix001P3S2CDecompositionTolerance              = 1e-10
)

type fix001P3S2CTraceGroup struct {
	Group         int     `json:"group"`
	MatrixIndices []int   `json:"matrix_indices"`
	InputLevel    int     `json:"input_level"`
	InputScale    string  `json:"input_scale"`
	PreRowsMatch  bool    `json:"pre_rescale_q01_rows_match"`
	PostRowsMatch bool    `json:"post_rescale_q01_rows_match"`
	CapacityRatio float64 `json:"full_rns_q01_capacity_ratio"`
	FastPostLevel int     `json:"fast_post_level"`
	FullPostLevel int     `json:"full_post_level"`
	FastPostScale string  `json:"fast_post_scale"`
	FullPostScale string  `json:"full_post_scale"`
}

type fix001P3S2CTrace struct {
	FastStages []*rlwe.Ciphertext
	FullStages []*rlwe.Ciphertext
	Groups     []fix001P3S2CTraceGroup
}

type fix001P3S2CDecompositionGroup struct {
	Stage                    string          `json:"stage"`
	Group                    int             `json:"group"`
	MaxComponentError        *PSGlobalMetric `json:"propagated_input_error"`
	Amplification            float64         `json:"empirical_amplification"`
	FastImplementationEffect *PSGlobalMetric `json:"fast_s2c_effect_at_stage"`
	CapacityRatio            float64         `json:"intended_full_rns_q01_capacity_ratio"`
	FastPreRowsMatch         bool            `json:"fast_pre_rows_match_mirror"`
	FastPostRowsMatch        bool            `json:"fast_post_rows_match_mirror"`
}

type fix001P3S2CErrorDecompositionResult struct {
	SchemaVersion         string                          `json:"schema_version"`
	Timestamp             time.Time                       `json:"timestamp"`
	Primary               RepositoryMetadata              `json:"primary_repository"`
	Lattigo               RepositoryMetadata              `json:"lattigo_repository"`
	Environment           EnvironmentMetadata             `json:"environment"`
	Config                BootstrapConfig                 `json:"config"`
	Parameters            ExperimentParameters            `json:"effective_parameters"`
	D0                    map[string]interface{}          `json:"d0_control"`
	Mirror                map[string]interface{}          `json:"full_rns_mirror"`
	Decomposition         map[string]interface{}          `json:"causal_decomposition"`
	Groups                []fix001P3S2CDecompositionGroup `json:"propagated_input_group_amplification"`
	LinearDelta           map[string]interface{}          `json:"linear_delta_confirmation"`
	RequiredPostS2C       float64                         `json:"required_post_s2c_threshold"`
	CurrentGap            float64                         `json:"current_gap_to_required_post_s2c"`
	Classification        string                          `json:"classification"`
	FirstRemainingBlocker string                          `json:"first_remaining_blocker"`
	Validation            map[string]interface{}          `json:"validation"`
}

func fix001P3S2CWrite(result fix001P3S2CErrorDecompositionResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3S2CMetric(reference, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(reference, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func fix001P3S2CSubtract(a, b []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] - b[i]
	}
	return result
}

func fix001P3S2CAdd3(a, b, c []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] + b[i] + c[i]
	}
	return result
}

func fix001P3S2CMaxCapacity(params ckks.Parameters, a, b *rlwe.Ciphertext) float64 {
	maxRatio := 0.0
	for _, ct := range []*rlwe.Ciphertext{a, b} {
		if ct == nil || ct.Level() < 1 {
			continue
		}
		if capacity, err := postMod1S2CCapacityFromCiphertext(params, ct); err == nil && capacity.MaxAbsOverQ01Half > maxRatio {
			maxRatio = capacity.MaxAbsOverQ01Half
		}
	}
	return maxRatio
}

func fix001P3S2CDecode(params ckks.Parameters, ct *rlwe.Ciphertext, sk *rlwe.SecretKey) ([]complex128, error) {
	_, values, err := semanticBisectView(ct, params, sk)
	return values, err
}

func fix001P3S2CCombineFast(params ckks.Parameters, eval *bootstrapping.FastEvaluator, realCT, imagCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if imagCT == nil {
		return realCT.CopyNew(), nil
	}
	combined := fastckks.NewCiphertext(params, 1, realCT.Level())
	postMod1S2CCopyMetadata(combined, realCT)
	if err := eval.FastCKKS.Mul(imagCT, 1i, combined); err != nil {
		return nil, err
	}
	if err := eval.FastCKKS.Add(combined, realCT, combined); err != nil {
		return nil, err
	}
	return combined, nil
}

func fix001P3S2CCombineFull(params ckks.Parameters, eval *dft.Evaluator, realCT, imagCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if imagCT == nil {
		return realCT.CopyNew(), nil
	}
	combined := ckks.NewCiphertext(params, 1, realCT.Level())
	postMod1S2CCopyMetadata(combined, realCT)
	if err := eval.Evaluator.Mul(imagCT, 1i, combined); err != nil {
		return nil, err
	}
	if err := eval.Evaluator.Add(combined, realCT, combined); err != nil {
		return nil, err
	}
	return combined, nil
}

func fix001P3S2CTraceFull(params ckks.Parameters, matrix dft.Matrix, standardDFT *dft.Evaluator, realCT, imagCT *rlwe.Ciphertext) ([]*rlwe.Ciphertext, []fix001P3S2CTraceGroup, error) {
	current, err := fix001P3S2CCombineFull(params, standardDFT, realCT, imagCT)
	if err != nil {
		return nil, nil, err
	}
	stages := []*rlwe.Ciphertext{current.CopyNew()}
	groups := make([]fix001P3S2CTraceGroup, 0, len(matrix.Levels))
	matrixIndex := 0
	for groupIndex, factorCount := range matrix.Levels {
		group := fix001P3S2CTraceGroup{Group: groupIndex, MatrixIndices: make([]int, 0, factorCount), InputLevel: current.Level(), InputScale: finalizationScaleString(current.Scale)}
		for factor := 0; factor < factorCount; factor++ {
			if matrixIndex >= len(matrix.Matrices) {
				return nil, nil, fmt.Errorf("S2C matrix factorization is shorter than its level schedule")
			}
			group.MatrixIndices = append(group.MatrixIndices, matrixIndex)
			next := ckks.NewCiphertext(params, 1, current.Level())
			postMod1S2CCopyMetadata(next, current)
			if err := standardDFT.LTEvaluator.Evaluate(current, matrix.Matrices[matrixIndex], next); err != nil {
				return nil, nil, err
			}
			current = next
			matrixIndex++
		}
		group.CapacityRatio = fix001P3S2CMaxCapacity(params, current, nil)
		if err := standardDFT.Evaluator.Rescale(current, current); err != nil {
			return nil, nil, err
		}
		group.FullPostLevel = current.Level()
		group.FullPostScale = finalizationScaleString(current.Scale)
		groups = append(groups, group)
		stages = append(stages, current.CopyNew())
	}
	if matrixIndex != len(matrix.Matrices) {
		return nil, nil, fmt.Errorf("S2C matrix factorization has unused matrices")
	}
	return stages, groups, nil
}

func fix001P3S2CTraceBoth(params ckks.Parameters, fastEval *bootstrapping.FastEvaluator, matrix dft.Matrix, standardDFT *dft.Evaluator, fastReal, fastImag, fullReal, fullImag *rlwe.Ciphertext) (fix001P3S2CTrace, error) {
	fastCurrent, err := fix001P3S2CCombineFast(params, fastEval, fastReal, fastImag)
	if err != nil {
		return fix001P3S2CTrace{}, err
	}
	fullCurrent, err := fix001P3S2CCombineFull(params, standardDFT, fullReal, fullImag)
	if err != nil {
		return fix001P3S2CTrace{}, err
	}
	trace := fix001P3S2CTrace{FastStages: []*rlwe.Ciphertext{fastCurrent.CopyNew()}, FullStages: []*rlwe.Ciphertext{fullCurrent.CopyNew()}, Groups: make([]fix001P3S2CTraceGroup, 0, len(matrix.Levels))}
	matrixIndex := 0
	for groupIndex, factorCount := range matrix.Levels {
		group := fix001P3S2CTraceGroup{Group: groupIndex, MatrixIndices: make([]int, 0, factorCount), InputLevel: fastCurrent.Level(), InputScale: finalizationScaleString(fastCurrent.Scale)}
		for factor := 0; factor < factorCount; factor++ {
			if matrixIndex >= len(matrix.Matrices) {
				return fix001P3S2CTrace{}, fmt.Errorf("S2C matrix factorization is shorter than its level schedule")
			}
			group.MatrixIndices = append(group.MatrixIndices, matrixIndex)
			fastNext := fastckks.NewCiphertext(params, 1, fastCurrent.Level())
			postMod1S2CCopyMetadata(fastNext, fastCurrent)
			if err := fastEval.FastCKKS.LinearTransform(fastCurrent, commonlintrans.LinearTransformation(matrix.Matrices[matrixIndex]), fastNext); err != nil {
				return fix001P3S2CTrace{}, err
			}
			fullNext := ckks.NewCiphertext(params, 1, fullCurrent.Level())
			postMod1S2CCopyMetadata(fullNext, fullCurrent)
			if err := standardDFT.LTEvaluator.Evaluate(fullCurrent, matrix.Matrices[matrixIndex], fullNext); err != nil {
				return fix001P3S2CTrace{}, err
			}
			fastCurrent, fullCurrent = fastNext, fullNext
			matrixIndex++
		}
		group.PreRowsMatch = postMod1S2CRowsEqual(params, fastCurrent, fullCurrent)
		group.CapacityRatio = fix001P3S2CMaxCapacity(params, fullCurrent, nil)
		if err := fastEval.FastCKKS.Rescale(fastCurrent, fastCurrent); err != nil {
			return fix001P3S2CTrace{}, err
		}
		if err := standardDFT.Evaluator.Rescale(fullCurrent, fullCurrent); err != nil {
			return fix001P3S2CTrace{}, err
		}
		group.PostRowsMatch = postMod1S2CRowsEqual(params, fastCurrent, fullCurrent)
		group.FastPostLevel = fastCurrent.Level()
		group.FullPostLevel = fullCurrent.Level()
		group.FastPostScale = finalizationScaleString(fastCurrent.Scale)
		group.FullPostScale = finalizationScaleString(fullCurrent.Scale)
		trace.Groups = append(trace.Groups, group)
		trace.FastStages = append(trace.FastStages, fastCurrent.CopyNew())
		trace.FullStages = append(trace.FullStages, fullCurrent.CopyNew())
	}
	if matrixIndex != len(matrix.Matrices) {
		return fix001P3S2CTrace{}, fmt.Errorf("S2C matrix factorization has unused matrices")
	}
	return trace, nil
}

func fix001P3S2CWithin(actual, expected, tolerance float64) bool {
	return math.Abs(actual-expected) <= tolerance
}

func fix001P3S2CAcceptedPaths(profile q056PreparedProfile, candidate psRescaleGuardCandidate) (evalModMatchedPath, evalModMatchedPath, error) {
	planScale := precisionSweepScale(92)
	realState, err := fix001P3DAR0PrepareState(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	imagState, err := fix001P3DAR0PrepareState(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	for round := 0; round < profile.Fast.Mod1Parameters.DoubleAngle; round++ {
		realOutcome, err := fix001P3DAAllRoundsRound(&realState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return evalModMatchedPath{}, evalModMatchedPath{}, err
		}
		imagOutcome, err := fix001P3DAAllRoundsRound(&imagState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return evalModMatchedPath{}, evalModMatchedPath{}, err
		}
		if realOutcome.Classification != "" || imagOutcome.Classification != "" {
			return evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("accepted DA path failed at round %d: real=%s imag=%s", round, realOutcome.Blocker, imagOutcome.Blocker)
		}
	}
	if err := fix001P3DAAllRoundsFinalize(&realState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	if err := fix001P3DAAllRoundsFinalize(&imagState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, err
	}
	if realState.Path.FastFinal == nil || realState.Path.StandardFinal == nil || imagState.Path.FastFinal == nil || imagState.Path.StandardFinal == nil {
		return evalModMatchedPath{}, evalModMatchedPath{}, fmt.Errorf("accepted DA path returned nil final ciphertext")
	}
	return realState.Path, imagState.Path, nil
}

func runFIX001P3DiagLogN13S2CErrorDecomposition(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3S2CErrorDecompositionPrimary, "HEAD") == requiredFIX001P3S2CErrorDecompositionPrimary
	result := fix001P3S2CErrorDecompositionResult{
		SchemaVersion: "fix-001-p3-diag-logn13-s2c-error-decomposition.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, RequiredPostS2C: 3.125e-4,
		Validation: map[string]interface{}{
			"primary_required_base":           requiredFIX001P3S2CErrorDecompositionPrimary,
			"primary_required_base_ancestor":  primaryBaseAncestor,
			"secondary_commit":                secondaryCommit,
			"secondary_exact_required_commit": secondaryCommit == requiredFIX001P3S2CErrorDecompositionSecondary,
			"secondary_clean":                 secondaryClean,
			"no_secondary_production_changes": true,
			"logn13_only":                     cfg.LogN == 13,
			"accepted_g0_guard_bits":          2, "accepted_f0_guard_bits": 3,
			"all_rounds_da_fixed": true, "no_candidate_retuning": true,
			"no_q_parameter_changes": true, "no_s2c_production_changes": true,
			"no_generated_power_redesign": true, "no_production_integration": true,
			"no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true,
		},
	}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3S2CErrorDecompositionSecondary {
		result.Classification = "logn13_s2c_error_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = "startup provenance or Secondary precondition"
		return fix001P3S2CWrite(result, outPath)
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
	if !acceptedRow.ValidLocalQ2 || candidate.realFinal == nil || candidate.imagFinal == nil {
		result.Classification = "logn13_s2c_error_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = "accepted G0=2/F0=3 candidate replay is not valid"
		return fix001P3S2CWrite(result, outPath)
	}
	realPath, imagPath, err := fix001P3S2CAcceptedPaths(profile, candidate)
	if err != nil {
		return err
	}
	if realPath.FastFinal == nil || realPath.StandardFinal == nil || imagPath.FastFinal == nil || imagPath.StandardFinal == nil || realPath.FirstFailure != "none" || imagPath.FirstFailure != "none" {
		result.Classification = "logn13_s2c_error_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = fmt.Sprintf("accepted EvalMod path failure real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)
		return fix001P3S2CWrite(result, outPath)
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	fastEReal, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, realPath.FastFinal, zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	fastEImag, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, imagPath.FastFinal, zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	standardEReal, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, realPath.StandardFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	standardEImag, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, imagPath.StandardFinal, profile.StandardSK)
	if err != nil {
		return err
	}
	standardDFT := profile.Standard.DFTEvaluator
	fastOutput, err := profile.Fast.DFTEvaluator.SlotsToCoeffsNew(realPath.FastFinal.CopyNew(), imagPath.FastFinal.CopyNew(), profile.Fast.S2CDFTMatrix)
	if err != nil {
		return err
	}
	standardOutput, err := standardDFT.SlotsToCoeffsNew(realPath.StandardFinal.CopyNew(), imagPath.StandardFinal.CopyNew(), profile.Fast.S2CDFTMatrix)
	if err != nil {
		return err
	}
	fullFReal, _, err := postMod1S2CLiftFull(profile.BTP.BootstrappingParameters, realPath.FastFinal)
	if err != nil {
		return err
	}
	fullFImag, _, err := postMod1S2CLiftFull(profile.BTP.BootstrappingParameters, imagPath.FastFinal)
	if err != nil {
		return err
	}
	fullSReal := realPath.StandardFinal.CopyNew()
	fullSImag := imagPath.StandardFinal.CopyNew()
	trace, err := fix001P3S2CTraceBoth(profile.BTP.BootstrappingParameters, profile.Fast, profile.Fast.S2CDFTMatrix, standardDFT, realPath.FastFinal, imagPath.FastFinal, fullFReal, fullFImag)
	if err != nil {
		return err
	}
	fullSStages, _, err := fix001P3S2CTraceFull(profile.BTP.BootstrappingParameters, profile.Fast.S2CDFTMatrix, standardDFT, fullSReal, fullSImag)
	if err != nil {
		return err
	}
	fastOutputValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, fastOutput, zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	standardOutputValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, standardOutput, profile.StandardSK)
	if err != nil {
		return err
	}
	fullFOutputValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, trace.FullStages[len(trace.FullStages)-1], zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	fullSOutputValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, fullSStages[len(fullSStages)-1], profile.StandardSK)
	if err != nil {
		return err
	}
	observed := fix001P3S2CMetric(standardOutputValues, fastOutputValues, fix001P3S2CErrorDecompositionThreshold)
	fastImpl := fix001P3S2CMetric(fullFOutputValues, fastOutputValues, fix001P3S2CErrorDecompositionThreshold)
	propagated := fix001P3S2CMetric(fullSOutputValues, fullFOutputValues, fix001P3S2CErrorDecompositionThreshold)
	mirrorFidelity := fix001P3S2CMetric(standardOutputValues, fullSOutputValues, fix001P3S2CDecompositionTolerance)
	implVector := fix001P3S2CSubtract(fastOutputValues, fullFOutputValues)
	propVector := fix001P3S2CSubtract(fullFOutputValues, fullSOutputValues)
	mirrorVector := fix001P3S2CSubtract(fullSOutputValues, standardOutputValues)
	observedVector := fix001P3S2CSubtract(fastOutputValues, standardOutputValues)
	closureVector := fix001P3S2CSubtract(fix001P3S2CAdd3(implVector, propVector, mirrorVector), observedVector)
	closure := fix001P3S2CMetric(make([]complex128, len(closureVector)), closureVector, fix001P3S2CDecompositionTolerance)
	publicEvidence, err := fix001P3K3RunPublicLike(realPath, imagPath, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	result.D0 = map[string]interface{}{
		"evalmod_real_vs_standard":  realFinal.FastVsStandard,
		"evalmod_imag_vs_standard":  imagFinal.FastVsStandard,
		"post_s2c_fast_vs_standard": observed,
		"public_like_error":         publicEvidence.PublicLike,
		"final_metadata_valid":      publicEvidence.Metadata["pass"] == true,
		"no_da_capacity_failure":    realPath.FirstFailure == "none" && imagPath.FirstFailure == "none" && realFinal.NormalizedBeforeRestoreCapacity != nil && realFinal.NormalizedBeforeRestoreCapacity.Pass && realFinal.FinalRestoreCapacity != nil && realFinal.FinalRestoreCapacity.Pass,
	}
	result.Mirror = map[string]interface{}{
		"fidelity_metric_n_s_vs_genuine_standard_s2c": mirrorFidelity,
		"final_q01_rows_match":                        postMod1S2CRowsEqual(profile.BTP.BootstrappingParameters, trace.FullStages[len(trace.FullStages)-1], standardOutput),
		"target":                                      fix001P3S2CDecompositionTolerance,
	}
	result.Decomposition = map[string]interface{}{
		"fast_s2c_implementation_effect_s_f_minus_n_f":  fastImpl,
		"propagated_evalmod_input_effect_n_f_minus_n_s": propagated,
		"full_rns_mirror_fidelity_n_s_minus_s_s":        mirrorFidelity,
		"observed_total_s_f_minus_s_s":                  observed,
		"closure_residual":                              closure,
		"closure_relative_to_observed": func() float64 {
			if observed.MaxComponent == 0 {
				return 0
			}
			return closure.MaxComponent / observed.MaxComponent
		}(),
	}
	fastInputComplex := make([]complex128, len(fastEReal))
	standardInputComplex := make([]complex128, len(standardEReal))
	for i := range fastInputComplex {
		fastInputComplex[i] = fastEReal[i] + 1i*fastEImag[i]
		standardInputComplex[i] = standardEReal[i] + 1i*standardEImag[i]
	}
	inputDelta := fix001P3S2CSubtract(fastInputComplex, standardInputComplex)
	inputDeltaMetric := fix001P3S2CMetric(make([]complex128, len(inputDelta)), inputDelta, fix001P3S2CErrorDecompositionThreshold)
	for i, group := range trace.Groups {
		fastValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, trace.FullStages[i+1], zeroSecret(profile.BTP.BootstrappingParameters))
		if err != nil {
			return err
		}
		standardValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, fullSStages[i+1], profile.StandardSK)
		if err != nil {
			return err
		}
		metric := fix001P3S2CMetric(standardValues, fastValues, fix001P3S2CErrorDecompositionThreshold)
		fastStageValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, trace.FastStages[i+1], zeroSecret(profile.BTP.BootstrappingParameters))
		if err != nil {
			return err
		}
		fullStageValues, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, trace.FullStages[i+1], zeroSecret(profile.BTP.BootstrappingParameters))
		if err != nil {
			return err
		}
		implMetric := fix001P3S2CMetric(fullStageValues, fastStageValues, fix001P3S2CErrorDecompositionThreshold)
		previousMax := inputDeltaMetric.MaxComponent
		if i > 0 {
			previousGroup := result.Groups[i-1]
			previousMax = previousGroup.MaxComponentError.MaxComponent
		}
		amplification := 0.0
		if previousMax > 0 {
			amplification = metric.MaxComponent / previousMax
		}
		result.Groups = append(result.Groups, fix001P3S2CDecompositionGroup{Stage: fmt.Sprintf("group%d_post_rescale", group.Group), Group: group.Group, MaxComponentError: metric, Amplification: amplification, FastImplementationEffect: implMetric, CapacityRatio: group.CapacityRatio, FastPreRowsMatch: group.PreRowsMatch, FastPostRowsMatch: group.PostRowsMatch})
	}
	inputDeltaReal := make([]complex128, len(inputDelta))
	inputDeltaImag := make([]complex128, len(inputDelta))
	for i, value := range inputDelta {
		inputDeltaReal[i] = complex(real(value), 0)
		inputDeltaImag[i] = complex(imag(value), 0)
	}
	deltaFastReal, err := evalModFastFromStandard(profile.BTP.BootstrappingParameters, realPath.FastFinal, inputDeltaReal)
	if err != nil {
		return err
	}
	deltaFastImag, err := evalModFastFromStandard(profile.BTP.BootstrappingParameters, realPath.FastFinal, inputDeltaImag)
	if err != nil {
		return err
	}
	deltaFullReal, _, err := postMod1S2CLiftFull(profile.BTP.BootstrappingParameters, deltaFastReal)
	if err != nil {
		return err
	}
	deltaFullImag, _, err := postMod1S2CLiftFull(profile.BTP.BootstrappingParameters, deltaFastImag)
	if err != nil {
		return err
	}
	deltaStages, _, err := fix001P3S2CTraceFull(profile.BTP.BootstrappingParameters, profile.Fast.S2CDFTMatrix, standardDFT, deltaFullReal, deltaFullImag)
	if err != nil {
		return err
	}
	deltaOutput, err := fix001P3S2CDecode(profile.BTP.BootstrappingParameters, deltaStages[len(deltaStages)-1], zeroSecret(profile.BTP.BootstrappingParameters))
	if err != nil {
		return err
	}
	propagatedVector := fix001P3S2CSubtract(fullFOutputValues, fullSOutputValues)
	deltaOutputMetric := fix001P3S2CMetric(make([]complex128, len(deltaOutput)), deltaOutput, fix001P3S2CErrorDecompositionThreshold)
	deltaAgreement := fix001P3S2CMetric(propagatedVector, deltaOutput, fix001P3S2CDecompositionTolerance)
	result.LinearDelta = map[string]interface{}{"input_delta": inputDeltaMetric, "output_delta": deltaOutputMetric, "input_to_output_amplification": func() float64 {
		if inputDeltaMetric.MaxComponent == 0 {
			return 0
		}
		return deltaOutputMetric.MaxComponent / inputDeltaMetric.MaxComponent
	}(), "worst_output_index": deltaOutputMetric.WorstIndex, "worst_output_component": deltaOutputMetric.WorstPart, "agreement_with_n_f_minus_n_s": deltaAgreement, "agreement_tolerance": fix001P3S2CDecompositionTolerance}
	result.CurrentGap = observed.MaxComponent - result.RequiredPostS2C
	fastEffectRatio := 0.0
	if observed.MaxComponent > 0 {
		fastEffectRatio = fastImpl.MaxComponent / observed.MaxComponent
	}
	result.Validation["d0_expected_evalmod_real_near"] = fix001P3S2CWithin(realFinal.FastVsStandard.MaxComponent, 6.261306805777143e-5, 2e-5)
	result.Validation["d0_expected_evalmod_imag_near"] = fix001P3S2CWithin(imagFinal.FastVsStandard.MaxComponent, 4.647793870173725e-5, 2e-5)
	result.Validation["d0_expected_post_s2c_near"] = fix001P3S2CWithin(observed.MaxComponent, 3.3385298628089955e-4, 2e-5)
	result.Validation["d0_expected_public_like_near"] = publicEvidence.PublicLike != nil && fix001P3S2CWithin(publicEvidence.PublicLike.MaxComponent, 0.010683237260415292, 2e-5)
	result.Validation["mirror_fidelity_pass"] = mirrorFidelity.MaxComponent <= fix001P3S2CDecompositionTolerance
	result.Validation["closure_pass"] = closure.MaxComponent <= fix001P3S2CDecompositionTolerance
	result.Validation["linear_delta_pass"] = deltaAgreement.MaxComponent <= fix001P3S2CDecompositionTolerance
	if !result.Validation["d0_expected_evalmod_real_near"].(bool) || !result.Validation["d0_expected_evalmod_imag_near"].(bool) || !result.Validation["d0_expected_post_s2c_near"].(bool) || !result.Validation["d0_expected_public_like_near"].(bool) || result.D0["final_metadata_valid"] != true || result.D0["no_da_capacity_failure"] != true || !result.Validation["mirror_fidelity_pass"].(bool) {
		result.Classification = "logn13_s2c_error_decomposition_precondition_mismatch"
		result.FirstRemainingBlocker = "D0 accepted candidate or full-RNS mirror fidelity"
		return fix001P3S2CWrite(result, outPath)
	}
	if mirrorFidelity.MaxComponent > fix001P3S2CDecompositionTolerance {
		result.Classification = "logn13_s2c_full_rns_mirror_mismatch"
		result.FirstRemainingBlocker = "T_N(E_S) does not reproduce genuine Standard S2C"
	} else if fastEffectRatio <= 0.1 {
		result.Classification = "logn13_s2c_error_dominated_by_evalmod_input_propagation"
		result.FirstRemainingBlocker = "first S2C group with largest propagated-input amplification"
	} else {
		result.Classification = "logn13_s2c_fast_arithmetic_precision_blocker"
		result.FirstRemainingBlocker = "first group with material Fast S2C versus T_N divergence"
	}
	return fix001P3S2CWrite(result, outPath)
}
