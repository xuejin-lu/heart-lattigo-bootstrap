package main

import (
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
)

type NumericalComplex struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type NumericalDimensions struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

type NumericalMetadata struct {
	Level              int                 `json:"level"`
	Degree             int                 `json:"degree"`
	Scale              float64             `json:"scale"`
	ScaleExact         string              `json:"scale_exact,omitempty"`
	ScaleFloat64Finite bool                `json:"scale_float64_finite"`
	ScaleLog2          float64             `json:"scale_log2"`
	IsNTT              bool                `json:"is_ntt"`
	IsMontgomery       bool                `json:"is_montgomery"`
	LogDimensions      NumericalDimensions `json:"log_dimensions"`
}

type NumericalPublicContract struct {
	LevelMatchesResidualMax  bool   `json:"level_matches_residual_max"`
	ScaleMatchesDefault      bool   `json:"scale_matches_default"`
	DimensionsMatchCanonical bool   `json:"dimensions_match_canonical"`
	Status                   string `json:"status"`
}

type NumericalErrorMetric struct {
	MeanAbsoluteError float64          `json:"mean_absolute_error"`
	RMSE              float64          `json:"rmse"`
	P50               float64          `json:"p50"`
	P95               float64          `json:"p95"`
	P99               float64          `json:"p99"`
	Max               float64          `json:"max"`
	WorstSlot         int              `json:"worst_slot"`
	WorstReference    NumericalComplex `json:"worst_reference"`
	WorstOutput       NumericalComplex `json:"worst_output"`
	WorstDifference   NumericalComplex `json:"worst_difference"`
}

type NumericalMetricSet struct {
	Real    NumericalErrorMetric `json:"real"`
	Imag    NumericalErrorMetric `json:"imag"`
	Complex NumericalErrorMetric `json:"complex"`
}

type NumericalThresholdAudit struct {
	Threshold        float64 `json:"threshold"`
	Comparisons      int     `json:"comparisons"`
	RealExceeding    int     `json:"real_exceeding"`
	RealCoordinates  int     `json:"real_coordinates"`
	RealFraction     float64 `json:"real_fraction"`
	ImagExceeding    int     `json:"imag_exceeding"`
	ImagCoordinates  int     `json:"imag_coordinates"`
	ImagFraction     float64 `json:"imag_fraction"`
	TotalExceeding   int     `json:"total_exceeding"`
	TotalCoordinates int     `json:"total_coordinates"`
	TotalFraction    float64 `json:"total_fraction"`
}

type NumericalPrecisionBits struct {
	ErrorFloor       float64 `json:"error_floor"`
	MeanBits         float64 `json:"mean_bits"`
	MedianBits       float64 `json:"median_bits"`
	P05Bits          float64 `json:"p05_bits"`
	MinimumBits      float64 `json:"minimum_bits"`
	MinimumSlot      int     `json:"minimum_slot"`
	MinimumSlotError float64 `json:"minimum_slot_error"`
}

type NumericalPrecisionVector struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
	L2   float64 `json:"l2"`
}

type NumericalCKKSPrecisionHelper struct {
	Minimum   NumericalPrecisionVector `json:"minimum_log2_precision"`
	Maximum   NumericalPrecisionVector `json:"maximum_log2_precision"`
	Average   NumericalPrecisionVector `json:"average_log2_precision"`
	Median    NumericalPrecisionVector `json:"median_log2_precision"`
	StdDev    NumericalPrecisionVector `json:"stddev_log2_precision"`
	Log2Scale float64                  `json:"log2_scale"`
}

type NumericalOutput struct {
	Index                                      int                          `json:"index"`
	Metadata                                   NumericalMetadata            `json:"metadata"`
	PublicContract                             NumericalPublicContract      `json:"public_bootstrap_contract"`
	VsOriginal                                 NumericalMetricSet           `json:"vs_original"`
	Threshold                                  NumericalThresholdAudit      `json:"threshold_audit"`
	PrecisionBits                              NumericalPrecisionBits       `json:"precision_bits"`
	CKKSHelper                                 NumericalCKKSPrecisionHelper `json:"ckks_precision_helper"`
	BootstrapSNR                               numericalmetrics.SNR         `json:"bootstrap_snr"`
	PreBootstrapVsCanonicalOriginalComplexRMSE float64                      `json:"pre_bootstrap_vs_canonical_original_complex_rmse"`
}

type NumericalStageState struct {
	Metadata       NumericalMetadata `json:"metadata"`
	AuthorityRows  int               `json:"authority_rows"`
	MaintainedRows int               `json:"maintained_rows,omitempty"`
}

type NumericalStageCheckpoint struct {
	Name                         string                  `json:"checkpoint"`
	Fast                         NumericalStageState     `json:"fast"`
	Standard                     NumericalStageState     `json:"standard"`
	Comparable                   bool                    `json:"comparable"`
	NotComparableReason          string                  `json:"not_comparable_reason,omitempty"`
	PrecisionVsCanonicalOriginal *NumericalPrecisionBits `json:"precision_vs_canonical_original,omitempty"`
	FastVsStandard               *NumericalMetricSet     `json:"fast_vs_standard,omitempty"`
	D                            *float64                `json:"d_i_rmse"`
	Amplification                *float64                `json:"a_i"`
	AmplificationStatus          string                  `json:"a_i_status"`
	StageReferenceSNR            numericalmetrics.SNR    `json:"stage_reference_snr"`
	DeltaSNRDB                   *float64                `json:"delta_snr_i_db"`
	DeltaSNRStatus               string                  `json:"delta_snr_i_status"`
	FirstObservable              bool                    `json:"first_observable,omitempty"`
	FirstMaterial                bool                    `json:"first_material,omitempty"`
}

type NumericalStageLockstep struct {
	Checkpoints                       []NumericalStageCheckpoint `json:"checkpoints"`
	FirstObservable                   string                     `json:"first_observable_checkpoint"`
	FirstObservableMaxDiff            *float64                   `json:"first_observable_max_diff"`
	FirstMaterial                     string                     `json:"first_material_checkpoint"`
	FirstMaterialMaxDiff              *float64                   `json:"first_material_max_diff"`
	MaterialThreshold                 float64                    `json:"material_threshold"`
	ObservableThreshold               float64                    `json:"observable_threshold"`
	Classification                    string                     `json:"classification"`
	FinalFastStandardRMSE             float64                    `json:"final_fast_standard_rmse"`
	S2CAmplificationFactor            *float64                   `json:"s2c_amplification_factor"`
	LargestRawAmplificationCheckpoint string                     `json:"largest_raw_amplification_checkpoint"`
	LargestRawAmplificationFactor     *float64                   `json:"largest_raw_amplification_factor"`
	LargestSNRDropCheckpoint          string                     `json:"largest_snr_drop_checkpoint"`
	LargestSNRDropDB                  *float64                   `json:"largest_snr_drop_db"`
	EvalModReplayVerified             map[string]bool            `json:"evalmod_replay_verified,omitempty"`
	EvalModReplayRMSE                 map[string]float64         `json:"evalmod_replay_rmse,omitempty"`
	EvalModInternal                   []NumericalStageCheckpoint `json:"evalmod_internal,omitempty"`
	ScaleAudit                        []NumericalScaleAudit      `json:"scale_audit,omitempty"`
	GeneratedPowerEvidence            []NumericalGeneratedPower  `json:"generated_power_evidence,omitempty"`
	PolynomialPlan                    *NumericalPolynomialPlan   `json:"polynomial_plan,omitempty"`
}

type NumericalScaleAudit struct {
	Checkpoint               string  `json:"checkpoint"`
	Scale                    float64 `json:"scale"`
	ScaleExact               string  `json:"scale_exact,omitempty"`
	ScaleFloat64Finite       bool    `json:"scale_float64_finite"`
	ScaleLog2                float64 `json:"scale_log2"`
	PlanScaleBits            int     `json:"plan_scale_bits,omitempty"`
	PlanScaleExact           string  `json:"plan_scale_exact,omitempty"`
	WorkingScaleBits         int     `json:"working_scale_bits,omitempty"`
	DoubleAngleRound         int     `json:"double_angle_round,omitempty"`
	KInExponent              int     `json:"k_in_exponent,omitempty"`
	MultiplierExponent       int     `json:"multiplier_exponent,omitempty"`
	TargetScale              float64 `json:"target_scale,omitempty"`
	TargetScaleExact         string  `json:"target_scale_exact,omitempty"`
	TargetScaleFloat64Finite bool    `json:"target_scale_float64_finite"`
	TargetScaleLog2          float64 `json:"target_scale_log2,omitempty"`
	Level                    int     `json:"level"`
}

type NumericalGeneratedPower struct {
	ReferenceKind      string  `json:"reference_kind"`
	Power              int     `json:"power"`
	Level              int     `json:"level"`
	ScaleLog2          float64 `json:"scale_log2"`
	FastMaintainedRows int     `json:"fast_maintained_rows"`
	RMSE               float64 `json:"rmse_fast_vs_reference"`
	MaxComplexDiff     float64 `json:"max_complex_diff_fast_vs_reference"`
}

type NumericalPolynomialPlan struct {
	Degree     int     `json:"degree"`
	Base       int     `json:"base"`
	Level      int     `json:"level"`
	ScaleLog2  float64 `json:"scale_log2"`
	ScaleExact string  `json:"scale_exact"`
	BlockCount int     `json:"block_count"`
}

type NumericalKeyTrial struct {
	Index               int    `json:"index"`
	FreshKeyGenerator   bool   `json:"fresh_key_generator"`
	FreshEvaluationKeys bool   `json:"fresh_evaluation_keys"`
	SecretKeyLevelP     int    `json:"secret_key_level_p"`
	UniqueSecretKey     bool   `json:"unique_secret_key_vs_prior_trials"`
	EvaluatorPath       string `json:"evaluator_path"`
}

type NumericalMetricAggregate struct {
	MeanAbsoluteError float64 `json:"median_mean_absolute_error_across_trials"`
	RMSE              float64 `json:"median_rmse_across_trials"`
	P50               float64 `json:"median_p50_across_trials"`
	P95               float64 `json:"median_p95_across_trials"`
	P99               float64 `json:"median_p99_across_trials"`
	Max               float64 `json:"median_max_across_trials"`
}

type NumericalAggregate struct {
	TrialCount              int                      `json:"trial_count"`
	Real                    NumericalMetricAggregate `json:"real"`
	Imag                    NumericalMetricAggregate `json:"imag"`
	Complex                 NumericalMetricAggregate `json:"complex"`
	MedianOfTrialMedianBits float64                  `json:"median_of_trial_median_precision_bits"`
	MedianOfTrialMeanBits   float64                  `json:"median_of_trial_mean_precision_bits"`
}

type NumericalMetadataComparison struct {
	Classification   string   `json:"classification"`
	Differences      []string `json:"differences,omitempty"`
	PublicContractOK bool     `json:"public_contract_ok"`
}

type NumericalPairwiseComparison struct {
	FastTrialIndex     int                            `json:"fast_trial_index"`
	StandardTrialIndex int                            `json:"standard_trial_index"`
	Metrics            NumericalMetricSet             `json:"metrics"`
	WorstExamples      NumericalPairwiseWorstExamples `json:"worst_slot_examples"`
	Threshold          NumericalThresholdAudit        `json:"threshold_audit"`
	Metadata           NumericalMetadataComparison    `json:"metadata_comparison"`
}

type NumericalPairwiseWorstExamples struct {
	Real    NumericalPairwiseWorstExample `json:"real"`
	Imag    NumericalPairwiseWorstExample `json:"imag"`
	Complex NumericalPairwiseWorstExample `json:"complex"`
}

type NumericalPairwiseWorstExample struct {
	Slot       int              `json:"slot"`
	Original   NumericalComplex `json:"original"`
	Fast       NumericalComplex `json:"fast"`
	Standard   NumericalComplex `json:"standard"`
	Difference NumericalComplex `json:"difference_fast_minus_standard"`
}

type NumericalFastDeterminism struct {
	ComparedRuns     int     `json:"compared_runs"`
	BitwiseIdentical bool    `json:"bitwise_identical"`
	MaxComplexDiff   float64 `json:"max_complex_diff"`
	MaxDiffSlot      int     `json:"max_diff_slot"`
}

type NumericalStandardSpread struct {
	TrialCount                   int                       `json:"trial_count"`
	PerSlotMeanFingerprintSHA256 string                    `json:"per_slot_mean_fingerprint_sha256"`
	RMSRealSpread                float64                   `json:"rms_real_spread"`
	RMSImagSpread                float64                   `json:"rms_imag_spread"`
	RMSComplexSpread             float64                   `json:"rms_complex_spread"`
	MaxComplexSpread             float64                   `json:"max_complex_spread_from_per_slot_mean"`
	MaxSpreadSlot                int                       `json:"max_spread_slot"`
	MaxSpreadTrialIndex          int                       `json:"max_spread_trial_index"`
	MaxSpreadMean                NumericalComplex          `json:"max_spread_per_slot_mean"`
	MaxSpreadOutput              NumericalComplex          `json:"max_spread_output"`
	MaxPairwiseComplexDiff       float64                   `json:"max_pairwise_complex_diff"`
	MaxPairwiseSlot              int                       `json:"max_pairwise_slot"`
	MaxPairwiseTrialA            int                       `json:"max_pairwise_trial_a"`
	MaxPairwiseTrialB            int                       `json:"max_pairwise_trial_b"`
	MaxPairwiseValueA            NumericalComplex          `json:"max_pairwise_value_a"`
	MaxPairwiseValueB            NumericalComplex          `json:"max_pairwise_value_b"`
	PerSlotMeanSamples           []NumericalMeanSlotSample `json:"per_slot_mean_samples"`
}

type NumericalMeanSlotSample struct {
	Slot int              `json:"slot"`
	Mean NumericalComplex `json:"mean"`
}

type NumericalClassificationChecks struct {
	FastVsOriginalNoCoordinateExceeds bool     `json:"fast_vs_original_no_coordinate_exceeds_1e2"`
	FastVsStandardNoCoordinateExceeds bool     `json:"fast_vs_standard_no_coordinate_exceeds_1e2"`
	MedianPrecisionDropWithinTwoBits  bool     `json:"median_precision_drop_within_2_bits"`
	FastRMSEWithinFourXOrNearZero     bool     `json:"fast_rmse_within_4x_or_standard_near_zero"`
	PublicMetadataContractRespected   bool     `json:"public_metadata_contract_respected"`
	StandardRMSENearZero              bool     `json:"standard_rmse_near_zero_for_ratio"`
	FastToStandardMedianPrecisionDrop float64  `json:"fast_to_standard_median_precision_drop_bits"`
	FastToStandardComplexRMSERatio    *float64 `json:"fast_to_standard_complex_rmse_ratio,omitempty"`
}

type NumericalWorkload struct {
	FingerprintSHA256 string `json:"fingerprint_sha256"`
	InputSlots        int    `json:"input_slots"`
	LogN              int    `json:"log_n"`
	LogSlots          int    `json:"log_slots"`
	QChainBits        []int  `json:"q_chain_bits"`
	PBits             []int  `json:"p_bits"`
	Q0Bits            int    `json:"q0_bits"`
	PolynomialDegree  int    `json:"polynomial_degree"`
	DoubleAngle       int    `json:"double_angle"`
	EvalModLogScale   int    `json:"eval_mod_log_scale"`
	LogMessageRatio   int    `json:"log_message_ratio"`
}

type NumericalDocument struct {
	SchemaVersion                 string                        `json:"schema_version"`
	Timestamp                     time.Time                     `json:"timestamp"`
	Profile                       string                        `json:"profile"`
	Primary                       RepositoryMetadata            `json:"primary_repository"`
	Secondary                     RepositoryMetadata            `json:"secondary_repository"`
	Environment                   EnvironmentMetadata           `json:"environment"`
	Workload                      NumericalWorkload             `json:"workload"`
	Execution                     map[string]int                `json:"execution"`
	Threshold                     float64                       `json:"correctness_threshold"`
	PrecisionErrorFloor           float64                       `json:"precision_error_floor"`
	StandardRMSENearZeroFloor     float64                       `json:"standard_rmse_near_zero_floor"`
	FastDeterminism               NumericalFastDeterminism      `json:"fast_repeat_determinism"`
	FastTrials                    []NumericalOutput             `json:"fast_trials"`
	StandardTrials                []NumericalOutput             `json:"standard_trials"`
	StandardKeyTrials             []NumericalKeyTrial           `json:"standard_key_trials"`
	FastAggregate                 NumericalAggregate            `json:"fast_aggregate"`
	StandardAggregate             NumericalAggregate            `json:"standard_aggregate"`
	FastVsOriginalThreshold       NumericalThresholdAudit       `json:"fast_vs_original_threshold_aggregate"`
	StandardVsOriginalThreshold   NumericalThresholdAudit       `json:"standard_vs_original_threshold_aggregate"`
	FastVsStandard                []NumericalPairwiseComparison `json:"fast_vs_standard_trials"`
	FastVsStandardThreshold       NumericalThresholdAudit       `json:"fast_vs_standard_threshold_aggregate"`
	StandardToStandardVariability NumericalStandardSpread       `json:"standard_to_standard_variability"`
	StageLockstep                 *NumericalStageLockstep       `json:"stage_lockstep,omitempty"`
	Classification                string                        `json:"classification"`
	ClassificationChecks          NumericalClassificationChecks `json:"classification_checks"`
	Limitations                   []string                      `json:"limitations"`
}
