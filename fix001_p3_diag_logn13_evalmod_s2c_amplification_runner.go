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
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3EvalModS2CAmplificationPrimary   = "f3161a96b7b3d6d1250abc9e7ea834f4ca98e393"
	requiredFIX001P3EvalModS2CAmplificationSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	fix001P3EvalModS2CAmplificationThreshold         = 1e-2
	fix001P3EvalModS2CC2SBudget                      = 1.9531249403614602e-5
)

type fix001P3EvalModS2CStage struct {
	Name             string                 `json:"name"`
	InputMetadata    *integratedC2SMetadata `json:"input_metadata,omitempty"`
	Metadata         *integratedC2SMetadata `json:"metadata,omitempty"`
	StandardMetadata *integratedC2SMetadata `json:"standard_metadata,omitempty"`
	Semantic         *semanticBisectMetric  `json:"semantic,omitempty"`
	RowsMatch        bool                   `json:"rows_match,omitempty"`
	Capacity         *postMod1S2CCapacity   `json:"capacity,omitempty"`
}

type fix001P3EvalModS2CGroup struct {
	Group                int                          `json:"group"`
	MatrixIndices        []int                        `json:"matrix_indices"`
	InputLevel           int                          `json:"input_level"`
	InputScale           string                       `json:"input_scale"`
	FastQ01Capacity      postMod1S2CCapacity          `json:"fast_q01_capacity"`
	FullRNSCapacity      postMod1S2CCapacity          `json:"full_rns_capacity"`
	PreRescaleRowsMatch  bool                         `json:"pre_rescale_rows_match"`
	PostRescaleRowsMatch bool                         `json:"post_rescale_rows_match"`
	FastPostLevel        int                          `json:"fast_post_level"`
	FullPostLevel        int                          `json:"full_post_level"`
	FastPostScale        string                       `json:"fast_post_scale"`
	FullPostScale        string                       `json:"full_post_scale"`
	FastPreRows          []FinalizationRowFingerprint `json:"fast_pre_rows,omitempty"`
	FullPreRows          []FinalizationRowFingerprint `json:"full_pre_rows,omitempty"`
	FastPostRows         []FinalizationRowFingerprint `json:"fast_post_rows,omitempty"`
	FullPostRows         []FinalizationRowFingerprint `json:"full_post_rows,omitempty"`
}

type fix001P3EvalModS2CAmplificationResult struct {
	SchemaVersion                 string                    `json:"schema_version"`
	Timestamp                     time.Time                 `json:"timestamp"`
	Primary                       RepositoryMetadata        `json:"primary_repository"`
	Lattigo                       RepositoryMetadata        `json:"lattigo_repository"`
	Environment                   EnvironmentMetadata       `json:"environment"`
	Config                        BootstrapConfig           `json:"config"`
	Parameters                    ExperimentParameters      `json:"effective_parameters"`
	Workload                      CorrectnessWorkload       `json:"workload"`
	StandardControl               *semanticBisectMetric     `json:"genuine_standard_public_bootstrap,omitempty"`
	FastPublic                    *semanticBisectMetric     `json:"fast_public_bootstrap,omitempty"`
	PublicFailure                 bool                      `json:"public_fast_failure_reproduced"`
	StandardPublicMetadataCorrect bool                      `json:"standard_public_metadata_correct"`
	FastPublicMetadataCorrect     bool                      `json:"fast_public_metadata_correct"`
	FastPublicInputUnchanged      bool                      `json:"fast_public_input_unchanged"`
	C2SReal                       *semanticBisectMetric     `json:"c2s_real_fast_vs_standard,omitempty"`
	C2SImag                       *semanticBisectMetric     `json:"c2s_imag_fast_vs_standard,omitempty"`
	C2SBudget                     float64                   `json:"c2s_budget"`
	EvalModReal                   *fix001P3EvalModS2CStage  `json:"evalmod_real,omitempty"`
	EvalModImag                   *fix001P3EvalModS2CStage  `json:"evalmod_imag,omitempty"`
	S2CReal                       []fix001P3EvalModS2CGroup `json:"s2c_real_groups,omitempty"`
	S2CImag                       []fix001P3EvalModS2CGroup `json:"s2c_imag_groups,omitempty"`
	S2CFinal                      *fix001P3EvalModS2CStage  `json:"s2c_final,omitempty"`
	HFReal                        *fix001P3EvalModS2CStage  `json:"h_f_real,omitempty"`
	HFImag                        *fix001P3EvalModS2CStage  `json:"h_f_imag,omitempty"`
	HSReal                        *fix001P3EvalModS2CStage  `json:"h_s_real,omitempty"`
	HSImag                        *fix001P3EvalModS2CStage  `json:"h_s_imag,omitempty"`
	EvalModInputDifferenceReal    float64                   `json:"evalmod_input_difference_real"`
	EvalModInputDifferenceImag    float64                   `json:"evalmod_input_difference_imag"`
	PostS2CDifferenceDueToEvalMod float64                   `json:"post_s2c_difference_due_to_evalmod_semantics"`
	AmplificationRatio            float64                   `json:"empirical_amplification_ratio"`
	FastS2CImplementationEffect   float64                   `json:"fast_s2c_implementation_effect"`
	UnpackEffect                  float64                   `json:"unpack_effect"`
	FinalizationEffect            float64                   `json:"finalization_effect"`
	TotalPublicFastError          float64                   `json:"total_public_fast_error"`
	CoreAfterS2C                  *fix001P3EvalModS2CStage  `json:"fast_core_after_s2c,omitempty"`
	AfterUnpack                   *fix001P3EvalModS2CStage  `json:"fast_after_unpack,omitempty"`
	ManualFinal                   *fix001P3EvalModS2CStage  `json:"manual_finalization,omitempty"`
	PublicFinal                   *fix001P3EvalModS2CStage  `json:"public_finalization,omitempty"`
	PublicEqualsManualFinal       bool                      `json:"public_equals_manual_finalization"`
	FirstFailingRound             string                    `json:"first_failing_round_or_checkpoint"`
	FirstSupportedCause           string                    `json:"classification"`
	PreviousEvalModDisposition    string                    `json:"local_evalmod_1e2_threshold_disposition"`
	Validation                    map[string]interface{}    `json:"validation"`
}

func fix001P3EvalModS2CWrite(result fix001P3EvalModS2CAmplificationResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3EvalModS2CStageOfCKKS(name string, ct *rlwe.Ciphertext, params bootstrapping.Parameters, sk *rlwe.SecretKey, expected []complex128) (*fix001P3EvalModS2CStage, error) {
	meta := integratedC2SMetadataOf(ct)
	stage := &fix001P3EvalModS2CStage{Name: name, Metadata: meta}
	if expected != nil {
		_, values, err := semanticBisectView(ct, params.BootstrappingParameters, sk)
		if err != nil {
			return nil, err
		}
		metric, err := semanticBisectMetricFor(expected, values, fix001P3EvalModS2CAmplificationThreshold)
		if err != nil {
			return nil, err
		}
		stage.Semantic = &metric
	}
	return stage, nil
}

func fix001P3EvalModS2CMetric(params bootstrapping.Parameters, reference, actual *rlwe.Ciphertext, referenceSK, actualSK *rlwe.SecretKey) (*semanticBisectMetric, error) {
	_, ref, err := semanticBisectView(reference, params.BootstrappingParameters, referenceSK)
	if err != nil {
		return nil, err
	}
	_, got, err := semanticBisectView(actual, params.BootstrappingParameters, actualSK)
	if err != nil {
		return nil, err
	}
	metric, err := semanticBisectMetricFor(ref, got, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func fix001P3EvalModS2CPlainMetric(reference, actual []complex128) (*semanticBisectMetric, error) {
	metric, err := semanticBisectMetricFor(reference, actual, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func fix001P3EvalModS2CMax(metric *semanticBisectMetric) float64 {
	if metric == nil {
		return math.NaN()
	}
	return metric.MaxComponentAbs
}

func fix001P3EvalModS2CGroupEvidence(groups []postMod1S2CGroupCheckpoint, params bootstrapping.Parameters, fast, full *rlwe.Ciphertext) []fix001P3EvalModS2CGroup {
	out := make([]fix001P3EvalModS2CGroup, 0, len(groups))
	for _, group := range groups {
		capacity := group.Capacity
		fastCapacity := capacity
		if fast != nil {
			if got, err := postMod1S2CCapacityFromFastCiphertext(params.BootstrappingParameters, fast); err == nil {
				fastCapacity = got
			}
		}
		out = append(out, fix001P3EvalModS2CGroup{
			Group: group.Group, MatrixIndices: append([]int(nil), group.MatrixIndices...), InputLevel: group.InputLevel, InputScale: group.InputScale,
			FastQ01Capacity: fastCapacity, FullRNSCapacity: capacity, PreRescaleRowsMatch: group.PreRescaleRowsMatch, PostRescaleRowsMatch: group.PostRescaleRowsMatch,
			FastPostLevel: group.FastPostLevel, FullPostLevel: group.ReferencePostLevel, FastPostScale: group.FastPostScale, FullPostScale: group.ReferencePostScale,
			FastPreRows: append([]FinalizationRowFingerprint(nil), group.FastPreRescaleRows...), FullPreRows: append([]FinalizationRowFingerprint(nil), group.ReferencePreRescaleRows...),
			FastPostRows: append([]FinalizationRowFingerprint(nil), group.FastPostRescaleRows...), FullPostRows: append([]FinalizationRowFingerprint(nil), group.ReferencePostRows...),
		})
	}
	return out
}

func fix001P3EvalModS2CStandardDFTForSecret(params ckks.Parameters, matrix dft.Matrix, sk *rlwe.SecretKey) (*dft.Evaluator, error) {
	kgen := rlwe.NewKeyGenerator(params)
	keys := rlwe.NewMemEvaluationKeySet(nil, kgen.GenGaloisKeysNew(matrix.GaloisElements(params), sk)...)
	return dft.NewEvaluator(params, ckks.NewEvaluator(params, keys)), nil
}

func fix001P3EvalModS2CClassification(failure *postMod1S2CFailure, groups []postMod1S2CGroupCheckpoint, hF, hS, s2cEffect float64) (string, string) {
	if failure != nil {
		if failure.Cause == "post_mod1_s2c_group_capacity_failure" {
			return "logn13_s2c_q01_capacity_alias", fmt.Sprintf("group%d.%s", failure.Group, failure.Checkpoint)
		}
		if failure.Cause == "post_mod1_s2c_linear_transform_mismatch" || failure.Cause == "post_mod1_s2c_rescale_mismatch" {
			if failure.Group == 0 {
				return "logn13_s2c_group0_mismatch", fmt.Sprintf("group%d.%s", failure.Group, failure.Checkpoint)
			}
			if failure.Group == 1 {
				return "logn13_s2c_group1_mismatch", fmt.Sprintf("group%d.%s", failure.Group, failure.Checkpoint)
			}
			return "logn13_s2c_group2_mismatch", fmt.Sprintf("group%d.%s", failure.Group, failure.Checkpoint)
		}
		return "logn13_fast_s2c_primitive_or_rescale_mismatch", failure.Checkpoint
	}
	if hF > fix001P3EvalModS2CAmplificationThreshold && s2cEffect <= fix001P3EvalModS2CAmplificationThreshold {
		return "logn13_evalmod_semantic_difference_amplified_by_s2c", "h_f_vs_h_s"
	}
	if len(groups) == 0 {
		return "logn13_post_c2s_public_failure_not_yet_isolated", "s2c_groups"
	}
	return "logn13_post_c2s_public_failure_not_yet_isolated", "causal_accounting"
}

func runFIX001P3DiagLogN13EvalModS2CAmplification(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, skN2, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return err
	}
	result := fix001P3EvalModS2CAmplificationResult{
		SchemaVersion: "fix-001-p3-diag-logn13-evalmod-s2c-amplification.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		C2SBudget: fix001P3EvalModS2CC2SBudget, FirstFailingRound: "none", FirstSupportedCause: "logn13_post_c2s_public_failure_not_yet_isolated", PreviousEvalModDisposition: "insufficient_for_e2e_causality",
		Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3EvalModS2CAmplificationPrimary, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredFIX001P3EvalModS2CAmplificationSecondary, "no_secondary_production_changes": true, "no_production_fix": true, "no_parameter_or_matrix_changes": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "raw_standard_double_angle_q01_oracle_avoided": true},
	}
	writePartial := func() error { return fix001P3EvalModS2CWrite(result, outPath) }
	if result.Validation["secondary_clean"] != true || result.Validation["secondary_exact_required_commit"] != true {
		return writePartial()
	}

	standardPublicInput := reproducibleInput(residual, btp)
	standardPublic, err := standardEval.Bootstrap(standardPublicInput)
	if err != nil {
		return writePartial()
	}
	standardValues, err := decodeWithSecret(residual, standardPublic, skN1)
	if err != nil {
		return err
	}
	standardMetric, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), standardValues, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return err
	}
	result.StandardControl = &standardMetric
	result.StandardPublicMetadataCorrect = standardPublic.Level() == residual.MaxLevel() && standardPublic.Degree() == 1 && standardPublic.N() == residual.N() && standardPublic.IsNTT && !standardPublic.IsMontgomery && standardPublic.Scale.Equal(residual.DefaultScale())
	if !standardMetric.Pass {
		result.FirstSupportedCause = "logn13_post_c2s_public_failure_precondition_mismatch"
		result.FirstFailingRound = "standard_public_bootstrap"
		return writePartial()
	}

	fastInput := reproducibleInput(residual, btp)
	fastPacked, ctxtN1, ctxtN2, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput})
	if err != nil {
		return err
	}
	fastScaled, _, err := fastEval.ScaleDown(&fastPacked[0])
	if err != nil {
		return err
	}
	fastModUp, err := fastEval.ModUp(fastScaled)
	if err != nil {
		return err
	}
	fastC2SReal, fastC2SImag, err := fastEval.CoeffsToSlots(fastModUp)
	if err != nil {
		return err
	}
	fastReal := fastC2SReal.CopyNew()
	fastImag := fastC2SImag.CopyNew()
	fastReal, err = fastEval.EvalMod(fastReal)
	if err != nil {
		return err
	}
	if fastImag != nil {
		fastImag, err = fastEval.EvalMod(fastImag)
		if err != nil {
			return err
		}
	}
	fastPublicInput := reproducibleInput(residual, btp)
	fastPublicBefore := fastPublicInput.CopyNew()
	fastPublic, err := fastEval.Bootstrap(fastPublicInput)
	if err != nil {
		return err
	}
	fastPublicValues, err := decodeWithSecret(residual, fastPublic, zeroSecret(residual))
	if err != nil {
		return err
	}
	fastPublicMetric, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), fastPublicValues, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return err
	}
	result.FastPublic, result.PublicFailure = &fastPublicMetric, !fastPublicMetric.Pass
	result.FastPublicMetadataCorrect = fastPublic.Level() == residual.MaxLevel() && fastPublic.Degree() == 1 && fastPublic.N() == residual.N() && fastPublic.IsNTT && !fastPublic.IsMontgomery && fastPublic.Scale.Equal(residual.DefaultScale())
	result.FastPublicInputUnchanged = fix001P3E2EInputUnchanged(fastPublicBefore, fastPublicInput)
	if !result.PublicFailure {
		result.FirstSupportedCause = "logn13_post_c2s_public_failure_precondition_mismatch"
		return writePartial()
	}

	standardInput := reproducibleInput(residual, btp)
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput})
	if err != nil {
		return err
	}
	standardScaled, _, err := standardEval.ScaleDown(&standardPacked[0])
	if err != nil {
		return err
	}
	standardModUp, err := standardEval.ModUp(standardScaled)
	if err != nil {
		return err
	}
	standardC2SReal, standardC2SImag, err := standardEval.CoeffsToSlots(standardModUp)
	if err != nil {
		return err
	}

	_, fastC2SRealValues, err := semanticBisectView(fastC2SReal, btp.BootstrappingParameters, zeroSecret(btp.BootstrappingParameters))
	if err != nil {
		return err
	}
	_, standardRealValues, err := semanticBisectView(standardC2SReal, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}
	c2sReal, err := semanticBisectMetricFor(standardRealValues, fastC2SRealValues, fix001P3EvalModS2CC2SBudget)
	if err != nil {
		return err
	}
	result.C2SReal = &c2sReal
	if fastC2SImag == nil || standardC2SImag == nil {
		return writePartial()
	}
	_, fastC2SImagValues, err := semanticBisectView(fastC2SImag, btp.BootstrappingParameters, zeroSecret(btp.BootstrappingParameters))
	if err != nil {
		return err
	}
	_, standardImagValues, err := semanticBisectView(standardC2SImag, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}
	c2sImag, err := semanticBisectMetricFor(standardImagValues, fastC2SImagValues, fix001P3EvalModS2CC2SBudget)
	if err != nil {
		return err
	}
	result.C2SImag = &c2sImag
	params := btp.BootstrappingParameters
	matchedStandardReal, _, err := postMod1S2CLiftFull(params, fastC2SReal)
	if err != nil {
		return err
	}
	matchedStandardImag, _, err := postMod1S2CLiftFull(params, fastC2SImag)
	if err != nil {
		return err
	}
	standardNativeReal, err := standardEval.EvalMod(matchedStandardReal)
	if err != nil {
		return err
	}
	standardNativeImag, err := standardEval.EvalMod(matchedStandardImag)
	if err != nil {
		return err
	}
	// Build the genuine Standard EvalMod control from the actual Fast C2S
	// semantic input. evalModMatchedRunPath uses centered full-RNS lifting and
	// the ordinary Standard polynomial/DoubleAngle path; it never treats raw
	// Standard q0/q1 DoubleAngle residues as an oracle.
	realPath, err := evalModMatchedRunPath(fastC2SReal, standardC2SReal, fastEval, standardEval, params, zeroSecret(params), skN2)
	if err != nil {
		return err
	}
	imagPath, err := evalModMatchedRunPath(fastC2SImag, standardC2SImag, fastEval, standardEval, params, zeroSecret(params), skN2)
	if err != nil {
		return err
	}
	if realPath.StandardFinal == nil || imagPath.StandardFinal == nil || realPath.FirstFailure != "none" || imagPath.FirstFailure != "none" {
		result.FirstSupportedCause = "logn13_post_c2s_public_failure_precondition_mismatch"
		result.FirstFailingRound = "matched_standard_evalmod"
		return writePartial()
	}

	result.EvalModReal, err = fix001P3EvalModS2CStageOfCKKS("fast", fastReal, btp, zeroSecret(btp.BootstrappingParameters), nil)
	if err != nil {
		return err
	}
	result.EvalModImag, err = fix001P3EvalModS2CStageOfCKKS("fast_imag", fastImag, btp, zeroSecret(btp.BootstrappingParameters), nil)
	if err != nil {
		return err
	}
	_, matchedFastRealValues, err := semanticBisectView(realPath.FastFinal, params, zeroSecret(params))
	if err != nil {
		return err
	}
	_, matchedStandardRealValues, err := semanticBisectView(realPath.StandardFinal, params, skN2)
	if err != nil {
		return err
	}
	matchedRealMetric, err := semanticBisectMetricFor(matchedStandardRealValues, matchedFastRealValues, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return err
	}
	_, matchedFastImagValues, err := semanticBisectView(imagPath.FastFinal, params, zeroSecret(params))
	if err != nil {
		return err
	}
	_, matchedStandardImagValues, err := semanticBisectView(imagPath.StandardFinal, params, skN2)
	if err != nil {
		return err
	}
	matchedImagMetric, err := semanticBisectMetricFor(matchedStandardImagValues, matchedFastImagValues, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return err
	}
	result.EvalModReal.Semantic = &matchedRealMetric
	result.EvalModImag.Semantic = &matchedImagMetric
	result.EvalModReal.InputMetadata = integratedC2SMetadataOf(fastC2SReal)
	result.EvalModImag.InputMetadata = integratedC2SMetadataOf(fastC2SImag)
	result.EvalModReal.StandardMetadata = integratedC2SMetadataOf(realPath.StandardFinal)
	result.EvalModImag.StandardMetadata = integratedC2SMetadataOf(imagPath.StandardFinal)
	result.EvalModInputDifferenceReal = result.EvalModReal.Semantic.MaxComponentAbs
	result.EvalModInputDifferenceImag = result.EvalModImag.Semantic.MaxComponentAbs

	fastCapacityReal, err := postMod1S2CCapacityFromFastCiphertext(params, fastReal)
	if err != nil {
		return err
	}
	fastCapacityImag, err := postMod1S2CCapacityFromFastCiphertext(params, fastImag)
	if err != nil {
		return err
	}
	result.EvalModReal.Capacity, result.EvalModImag.Capacity = &fastCapacityReal, &fastCapacityImag
	if !fastCapacityReal.Pass || !fastCapacityImag.Pass {
		result.FirstSupportedCause, result.FirstFailingRound = "logn13_s2c_q01_capacity_alias", "evalmod_output"
		return writePartial()
	}

	fullReal, _, err := postMod1S2CLiftFull(params, fastReal)
	if err != nil {
		return err
	}
	fullImag, _, err := postMod1S2CLiftFull(params, fastImag)
	if err != nil {
		return err
	}
	standardDFT, err := newPostMod1StandardDFT(params, fastEval.S2CDFTMatrix)
	if err != nil {
		return err
	}
	state := chebyshevOracleState{Params: btp, Eval: fastEval}
	fastReplay, fullReplay, groups, failure, err := postMod1S2CReplay(state, fastReal, fastImag, fullReal, fullImag, standardDFT)
	if err != nil {
		return err
	}
	result.S2CReal = fix001P3EvalModS2CGroupEvidence(groups, btp, fastReal, fullReal)
	result.S2CImag = append([]fix001P3EvalModS2CGroup(nil), result.S2CReal...)
	if failure != nil {
		result.FirstSupportedCause, result.FirstFailingRound = fix001P3EvalModS2CClassification(failure, groups, 0, 0, math.Inf(1))
		return writePartial()
	}
	_, fastReplayValues, err := semanticBisectView(fastReplay, params, zeroSecret(params))
	if err != nil {
		return err
	}
	_, fullReplayValues, err := semanticBisectView(fullReplay, params, zeroSecret(params))
	if err != nil {
		return err
	}
	s2cImplMetric, err := semanticBisectMetricFor(fullReplayValues, fastReplayValues, fix001P3EvalModS2CAmplificationThreshold)
	if err != nil {
		return err
	}
	result.FastS2CImplementationEffect = s2cImplMetric.MaxComponentAbs
	result.S2CFinal, err = fix001P3EvalModS2CStageOfCKKS("fast_replay_final", fastReplay, btp, zeroSecret(params), nil)
	if err != nil {
		return err
	}
	result.S2CFinal.RowsMatch = postMod1S2CRowsEqual(params, fastReplay, fullReplay)

	// H_F is the valid full-RNS lift of the actual Fast EvalMod output; H_S is
	// the genuine Standard EvalMod on the same matched C2S semantic input.
	hFReal := fullReal.CopyNew()
	hFImag := fullImag.CopyNew()
	hSReal := standardNativeReal.CopyNew()
	hSImag := standardNativeImag.CopyNew()
	secretDFT, err := fix001P3EvalModS2CStandardDFTForSecret(params, fastEval.S2CDFTMatrix, skN2)
	if err != nil {
		return err
	}
	hFOutReal, err := secretDFT.SlotsToCoeffsNew(hFReal, hFImag, fastEval.S2CDFTMatrix)
	if err != nil {
		return err
	}
	hSOutReal, err := secretDFT.SlotsToCoeffsNew(hSReal, hSImag, fastEval.S2CDFTMatrix)
	if err != nil {
		return err
	}
	hFValues, err := decodeWithSecret(params, hFOutReal, skN2)
	if err != nil {
		return err
	}
	hSValues, err := decodeWithSecret(params, hSOutReal, skN2)
	if err != nil {
		return err
	}
	hFMetric, err := fix001P3EvalModS2CPlainMetric(hSValues, hFValues)
	if err != nil {
		return err
	}
	result.PostS2CDifferenceDueToEvalMod = hFMetric.MaxComponentAbs
	result.HFReal, err = fix001P3EvalModS2CStageOfCKKS("h_f", hFOutReal, btp, skN2, reproducibleValues(residual.MaxSlots()))
	if err != nil {
		return err
	}
	result.HSReal, err = fix001P3EvalModS2CStageOfCKKS("h_s", hSOutReal, btp, skN2, reproducibleValues(residual.MaxSlots()))
	if err != nil {
		return err
	}
	result.HFReal.Semantic = hFMetric
	result.HFImag = result.HFReal
	result.HSImag = result.HSReal
	result.AmplificationRatio = 0
	denominator := math.Max(result.EvalModInputDifferenceReal, result.EvalModInputDifferenceImag)
	if denominator > 0 {
		result.AmplificationRatio = result.PostS2CDifferenceDueToEvalMod / denominator
	}

	coreOut, err := fastEval.SlotsToCoeffs(fastReal.CopyNew(), fastImag.CopyNew())
	if err != nil {
		return err
	}
	coreValues, err := psGlobalDecode(params, coreOut)
	if err != nil {
		return err
	}
	coreMetric, err := fix001P3EvalModS2CPlainMetric(reproducibleValues(residual.MaxSlots()), coreValues)
	if err != nil {
		return err
	}
	result.CoreAfterS2C, err = fix001P3EvalModS2CStageOfCKKS("fast_core_after_s2c", coreOut, btp, zeroSecret(params), nil)
	if err != nil {
		return err
	}
	result.CoreAfterS2C.Semantic = coreMetric
	coreBefore := coreOut.CopyNew()
	unpacked, err := fastEval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*coreOut.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil {
		return err
	}
	if len(unpacked) != 1 {
		return fmt.Errorf("expected one unpacked ciphertext, got %d", len(unpacked))
	}
	finalBefore, manualFinal, err := fix001P3E2EFinalizeOracle(&unpacked[0], btp)
	if err != nil {
		return err
	}
	_ = finalBefore
	unpackValues, err := psGlobalDecode(params, &unpacked[0])
	if err != nil {
		return err
	}
	unpackMetric, err := fix001P3EvalModS2CPlainMetric(reproducibleValues(residual.MaxSlots()), unpackValues)
	if err != nil {
		return err
	}
	result.AfterUnpack, err = fix001P3EvalModS2CStageOfCKKS("fast_after_unpack", &unpacked[0], btp, zeroSecret(params), nil)
	if err != nil {
		return err
	}
	result.AfterUnpack.Semantic = unpackMetric
	manualValues, err := decodeWithSecret(residual, manualFinal, zeroSecret(residual))
	if err != nil {
		return err
	}
	manualMetric, err := fix001P3EvalModS2CPlainMetric(reproducibleValues(residual.MaxSlots()), manualValues)
	if err != nil {
		return err
	}
	result.ManualFinal, err = fix001P3EvalModS2CStageOfCKKS("manual_finalization", manualFinal, btp, zeroSecret(params), nil)
	if err != nil {
		return err
	}
	result.ManualFinal.Semantic = manualMetric
	publicValues, err := decodeWithSecret(residual, fastPublic, zeroSecret(residual))
	if err != nil {
		return err
	}
	publicMetric, err := fix001P3EvalModS2CPlainMetric(reproducibleValues(residual.MaxSlots()), publicValues)
	if err != nil {
		return err
	}
	result.PublicFinal, err = fix001P3EvalModS2CStageOfCKKS("public_finalization", fastPublic, btp, zeroSecret(residual), nil)
	if err != nil {
		return err
	}
	result.PublicFinal.Semantic = publicMetric
	result.PublicEqualsManualFinal = fix001P3E2ERowsEqual(manualFinal, fastPublic)
	result.UnpackEffect = unpackMetric.MaxComponentAbs - coreMetric.MaxComponentAbs
	result.FinalizationEffect = publicMetric.MaxComponentAbs - unpackMetric.MaxComponentAbs
	result.TotalPublicFastError = publicMetric.MaxComponentAbs
	_ = coreBefore
	result.FirstSupportedCause, result.FirstFailingRound = fix001P3EvalModS2CClassification(nil, groups, result.PostS2CDifferenceDueToEvalMod, result.PostS2CDifferenceDueToEvalMod, result.FastS2CImplementationEffect)
	if result.FastS2CImplementationEffect > fix001P3EvalModS2CAmplificationThreshold {
		result.FirstSupportedCause = "logn13_fast_s2c_primitive_or_rescale_mismatch"
	}
	return fix001P3EvalModS2CWrite(result, outPath)
}
