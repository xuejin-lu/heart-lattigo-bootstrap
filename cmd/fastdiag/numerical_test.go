package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestFastStandardCanonicalInputFingerprint(t *testing.T) {
	require.Equal(t, fastStandardInputSHA256, fastStandardInputFingerprint(fastStandardP93Values()))
}

func TestValidateNumericalVectorUsesEffectiveProfileSlotCount(t *testing.T) {
	values := make([]complex128, 1<<15)
	require.NoError(t, validateNumericalVector(values, 1<<15))
	require.ErrorContains(t, validateNumericalVector(values, 1<<12), "want 4096")
}

func TestNumericalProfilesUseCanonicalEffectiveParameters(t *testing.T) {
	root := filepath.Join("..", "..")
	p93, p93Residual, err := fastStandardP93Parameters()
	require.NoError(t, err)
	sharedP93, sharedResidual, p93Values, p93Effective, p93Config, p93ConfigHash, err := numericalProfile(supportedProfile, root)
	require.NoError(t, err)
	require.Equal(t, p93.BootstrappingParameters.Q(), sharedP93.BootstrappingParameters.Q())
	require.Equal(t, p93.BootstrappingParameters.P(), sharedP93.BootstrappingParameters.P())
	require.True(t, p93Residual.Equal(&sharedResidual), "LogN13 config profile must reproduce the canonical P93 residual parameters")
	require.Equal(t, p93.BootstrappingParameters.LogN(), sharedP93.BootstrappingParameters.LogN())
	require.Equal(t, p93.BootstrappingParameters.MaxLevel(), sharedP93.BootstrappingParameters.MaxLevel())
	require.Equal(t, p93.BootstrappingParameters.DefaultScale(), sharedP93.BootstrappingParameters.DefaultScale())
	require.Equal(t, p93.CoeffsToSlotsParameters, sharedP93.CoeffsToSlotsParameters)
	require.Equal(t, p93.SlotsToCoeffsParameters, sharedP93.SlotsToCoeffsParameters)
	require.Equal(t, p93.Mod1ParametersLiteral, sharedP93.Mod1ParametersLiteral)
	require.Equal(t, p93.CircuitOrder, sharedP93.CircuitOrder)
	require.Equal(t, fastStandardInputSHA256, perfmeasure.Fingerprint(p93Values))
	require.Equal(t, 55, p93Effective.Q0Bits)
	require.Equal(t, "configs/bootstrap_config.logN13.json", p93Config)
	require.Equal(t, "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98", p93ConfigHash)
	require.Equal(t, fastStandardP93Values(), p93Values)
	require.Equal(t, fastStandardInputSHA256, perfmeasure.Fingerprint(p93Values))
	p93Input, err := fastStandardEncodedInput(p93Residual, sharedP93.CoeffsToSlotsParameters.LogSlots, p93Values)
	require.NoError(t, err)
	require.Equal(t, fastStandardInputMetadata13,
		perfmeasure.CiphertextStateFingerprint(perfmeasure.CiphertextStateOf(p93Input, sharedP93.BootstrappingParameters.Q(), 1)))

	logN16, logN16Residual, logN16Values, logN16Effective, config, configHash, err := numericalProfile(logN16Profile, root)
	require.NoError(t, err)
	require.Equal(t, 16, logN16.BootstrappingParameters.LogN())
	require.Equal(t, 15, logN16.CoeffsToSlotsParameters.LogSlots)
	require.Len(t, logN16Values, 1<<15)
	require.Equal(t, 56, logN16Effective.Q0Bits, "report the generated prime bit length, not the requested config scale")
	require.Equal(t, "configs/bootstrap_config.logN16.json", config)
	require.Equal(t, "16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02", configHash)
	require.Equal(t, fastStandardInputSHA256LogN16, perfmeasure.Fingerprint(logN16Values))
	logN16Input, err := fastStandardEncodedInput(logN16Residual, logN16.CoeffsToSlotsParameters.LogSlots, logN16Values)
	require.NoError(t, err)
	require.Equal(t, fastStandardInputMetadata16,
		perfmeasure.CiphertextStateFingerprint(perfmeasure.CiphertextStateOf(logN16Input, logN16.BootstrappingParameters.Q(), 1)))
}

func TestNumericalFastRevisionModesFailClosed(t *testing.T) {
	require.NoError(t, validateNumericalFastReference(numericalFastRevisionHistorical, "fast-qprefix", correctedFastCommitSHA, ""))
	require.ErrorContains(t, validateNumericalFastReference(numericalFastRevisionHistorical, "fast-qprefix", "2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f", ""), "historical-pinned")
	require.NoError(t, validateNumericalFastReference(numericalFastRevisionCurrent, "fast-qprefix", "current", "current"))
	require.ErrorContains(t, validateNumericalFastReference(numericalFastRevisionCurrent, "fast-qprefix", "local-ahead", "remote"), "equal to fetched origin")
	require.ErrorContains(t, validateNumericalFastReference(numericalFastRevisionCurrent, "", "current", "current"), "branch=")
}

func TestCurrentLogN16ClassificationUsesStrictGateAndCompleteBranches(t *testing.T) {
	checkpoint := func(name string, rmse float64, comparable bool) NumericalStageCheckpoint {
		return NumericalStageCheckpoint{Name: name, Comparable: comparable, FastVsStandard: &NumericalMetricSet{Complex: NumericalErrorMetric{RMSE: rmse}}}
	}
	stageCheckpoints := make([]NumericalStageCheckpoint, 0, 9)
	for _, name := range []string{"input", "scale_down", "mod_up", "c2s_real", "c2s_imag", "evalmod_real", "evalmod_imag", "s2c", "final_public_output"} {
		stageCheckpoints = append(stageCheckpoints, checkpoint(name, 5e-8, true))
	}
	internal := make([]NumericalStageCheckpoint, 0, 2*24)
	baseInternal := []string{
		"input_before_normalization", "after_normalization", "after_chebyshev_offset", "polynomial_input", "polynomial_output",
		"before_double_angle_round_0", "evalmod_output_before_public_scale_reset", "evalmod_output_after_public_scale_reset",
	}
	for round := 0; round < 3; round++ {
		baseInternal = append(baseInternal,
			fmt.Sprintf("double_angle_round_%d_after_multiply", round),
			fmt.Sprintf("double_angle_round_%d_after_add", round),
			fmt.Sprintf("double_angle_round_%d_after_constant", round),
			fmt.Sprintf("double_angle_round_%d_after_rescale", round),
		)
	}
	powers := make([]NumericalGeneratedPower, 0, 8)
	for _, branch := range []string{"real", "imag"} {
		for _, name := range baseInternal {
			internal = append(internal, checkpoint(branch+"/"+name, 5e-8, true))
		}
		for _, power := range []int{1, 2, 3, 6} {
			powers = append(powers, NumericalGeneratedPower{Branch: branch, Power: power, RMSEFastStandard: 5e-8})
		}
	}
	stage := NumericalStageLockstep{
		FinalFastStandardRMSE: 5e-8,
		Checkpoints:           stageCheckpoints, EvalModInternal: internal, GeneratedPowerEvidence: powers,
	}
	fast := NumericalAggregate{Complex: NumericalMetricAggregate{RMSE: 5e-8}}
	require.Equal(t, "LOGN16_CURRENT_NUMERICALLY_CLOSE", currentLogN16Classification(fast, stage))
	stage.EvalModInternal[1].FastVsStandard.Complex.RMSE = 1e-6
	require.Equal(t, "LOGN16_CURRENT_STILL_DIVERGES", currentLogN16Classification(fast, stage))
	stage.EvalModInternal[1].FastVsStandard.Complex.RMSE = 5e-8
	stage.Checkpoints[3].FastVsStandard.Complex.RMSE = 1e-6
	require.Equal(t, "LOGN16_CURRENT_STILL_DIVERGES", currentLogN16Classification(fast, stage))
	stage.Checkpoints[3].FastVsStandard.Complex.RMSE = 5e-8
	stage.Checkpoints[6].Comparable = false
	require.Equal(t, "LOGN16_CURRENT_REVALIDATION_BLOCKED", currentLogN16Classification(fast, stage))
}

func TestNumericalComparisonMetricsAndThresholdCoordinates(t *testing.T) {
	want := []complex128{0, 1 + 1i, -1 - 1i}
	have := []complex128{0.02 - 0.01i, 1.001 + 0.999i, -1 - 1i}
	metrics, threshold := numericalComparison(want, have)

	require.InDelta(t, 0.007, metrics.Real.MeanAbsoluteError, 1e-15)
	require.InDelta(t, 0.001, metrics.Real.P50, 1e-15)
	require.InDelta(t, 0.02, metrics.Real.Max, 1e-15)
	require.Equal(t, 0, metrics.Real.WorstSlot)
	require.Equal(t, 1, threshold.RealExceeding)
	require.Equal(t, 0, threshold.ImagExceeding, "threshold uses strict greater-than")
	require.Equal(t, 1, threshold.TotalExceeding)
	require.InDelta(t, 1.0/6, threshold.TotalFraction, 1e-15)
}

func TestReusableSNRNoiseRMSEMatchesExistingComplexRMSE(t *testing.T) {
	reference := []complex128{1 + 2i, -3 + 4i, 0.25 - 0.5i}
	output := []complex128{1.1 + 1.8i, -2.5 + 3.75i, 0.25 - 0.5i}
	metric := numericalmetrics.Compare(reference, output)
	legacy, _ := numericalComparison(reference, output)
	require.Equal(t, numericalmetrics.Finite, metric.Status)
	require.NotNil(t, metric.NoiseRMSE)
	require.InDelta(t, legacy.Complex.RMSE, *metric.NoiseRMSE, 1e-15)
}

func TestNumerical001JSONRemainsParseableAndLegacyMetricsRemainUnchanged(t *testing.T) {
	data, err := os.ReadFile("../../results/FAST-STANDARD-NUMERICAL-001-summary.json")
	require.NoError(t, err)
	var doc NumericalDocument
	require.NoError(t, json.Unmarshal(data, &doc))
	require.Equal(t, "fastdiag.numerical.v1", doc.SchemaVersion)
	require.NotEmpty(t, doc.FastTrials)
	require.NotEmpty(t, doc.StandardTrials)
	require.InDelta(t, 0.020892211083414026, doc.FastTrials[0].VsOriginal.Complex.RMSE, 1e-15)
	require.InDelta(t, 5.743515404716076, doc.FastTrials[0].PrecisionBits.MedianBits, 1e-12)
	require.InDelta(t, 1.1903936621664996e-9, doc.StandardTrials[0].VsOriginal.Complex.RMSE, 1e-20)
	require.InDelta(t, 30.795452055977087, doc.StandardTrials[0].PrecisionBits.MedianBits, 1e-12)
}

func TestPairwiseWorstSlotIncludesOriginalFastAndStandard(t *testing.T) {
	original := []complex128{1 + 2i, 3 + 4i}
	fast := []complex128{1.1 + 2.1i, 3.2 + 4.3i}
	standard := []complex128{1.01 + 2.01i, 3.1 + 4.1i}
	metrics, _ := numericalComparison(fast, standard)
	worst := pairwiseWorstExamples(original, fast, standard, metrics)

	require.Equal(t, 1, worst.Complex.Slot)
	require.Equal(t, numericalComplex(original[1]), worst.Complex.Original)
	require.Equal(t, numericalComplex(fast[1]), worst.Complex.Fast)
	require.Equal(t, numericalComplex(standard[1]), worst.Complex.Standard)
	require.Equal(t, numericalComplex(fast[1]-standard[1]), worst.Complex.Difference)
}

func TestNumericalPrecisionFloorAndQuantileInterpolation(t *testing.T) {
	precision := numericalPrecisionBits([]complex128{1, 2, 3}, []complex128{1, 2, 3})
	require.InDelta(t, -math.Log2(precisionErrorFloor), precision.MinimumBits, 1e-12)
	require.Equal(t, 0, precision.MinimumSlot)
	require.InDelta(t, 1.5, numericalQuantile([]float64{0, 1, 2, 3}, 0.5), 1e-15)
	require.InDelta(t, 0.15, numericalQuantile([]float64{0, 1, 2, 3}, 0.05), 1e-15)
}

func TestNumericalStandardSpreadTracksPerSlotMeans(t *testing.T) {
	trials := [][]complex128{
		{1 + 1i, 3 + 3i},
		{3 + 3i, 5 + 5i},
		{5 + 5i, 7 + 7i},
	}
	spread := numericalStandardSpread(trials)
	require.Equal(t, 3, spread.TrialCount)
	require.NotEmpty(t, spread.PerSlotMeanFingerprintSHA256)
	require.InDelta(t, math.Sqrt(8.0/3.0), spread.RMSRealSpread, 1e-15)
	require.InDelta(t, math.Sqrt(8.0/3.0), spread.RMSImagSpread, 1e-15)
	require.InDelta(t, 4*math.Sqrt2, spread.MaxPairwiseComplexDiff, 1e-15)
	require.Equal(t, 0, spread.MaxPairwiseSlot)

	identical := numericalStandardSpread([][]complex128{{0.1 + 0.2i}, {0.1 + 0.2i}, {0.1 + 0.2i}})
	require.Zero(t, identical.RMSComplexSpread)
	require.Zero(t, identical.MaxComplexSpread)
}

func TestStageMetricsFollowBranchTopology(t *testing.T) {
	checkpoints := []NumericalStageCheckpoint{
		syntheticStageCheckpoint("input", 0.5, 90),
		syntheticStageCheckpoint("scale_down", 1, 80),
		syntheticStageCheckpoint("mod_up", 2, 70),
		syntheticStageCheckpoint("c2s_real", 4, 60),
		syntheticStageCheckpoint("c2s_imag", 8, 50),
		syntheticStageCheckpoint("evalmod_real", 40, 40),
		syntheticStageCheckpoint("evalmod_imag", 80, 30),
		syntheticStageCheckpoint("s2c", 60, 20),
		syntheticStageCheckpoint("final_public_output", 120, 15),
	}
	jointSNR := syntheticSNR(35)
	result := NumericalStageLockstep{Checkpoints: checkpoints}
	applyNumericalStageTopology(&result, &jointSNR)

	require.Equal(t, "mod_up", stageCheckpoint(t, &result, "c2s_real").AmplificationParent)
	require.Equal(t, "mod_up", stageCheckpoint(t, &result, "c2s_imag").AmplificationParent)
	require.InDelta(t, 2, *stageCheckpoint(t, &result, "c2s_real").Amplification, 1e-15)
	require.InDelta(t, 4, *stageCheckpoint(t, &result, "c2s_imag").Amplification, 1e-15)
	require.Equal(t, "c2s_real", stageCheckpoint(t, &result, "evalmod_real").AmplificationParent)
	require.Equal(t, "c2s_imag", stageCheckpoint(t, &result, "evalmod_imag").AmplificationParent)
	require.InDelta(t, 10, *stageCheckpoint(t, &result, "evalmod_real").Amplification, 1e-15)
	require.InDelta(t, 10, *stageCheckpoint(t, &result, "evalmod_imag").Amplification, 1e-15)

	s2c := stageCheckpoint(t, &result, "s2c")
	require.Nil(t, s2c.Amplification)
	require.Equal(t, "NOT_APPLICABLE_COMBINED_BRANCH", s2c.AmplificationStatus)
	require.Equal(t, "evalmod_real+evalmod_imag (joint)", s2c.DeltaSNRParent)
	require.InDelta(t, -15, *s2c.DeltaSNRDB, 1e-15)
	require.Equal(t, "s2c", stageCheckpoint(t, &result, "final_public_output").AmplificationParent)
	require.InDelta(t, 2, *stageCheckpoint(t, &result, "final_public_output").Amplification, 1e-15)
	require.Equal(t, "evalmod_real", result.LargestRawAmplificationCheckpoint)
	require.InDelta(t, 10, *result.LargestRawAmplificationFactor, 1e-15)
}

func TestCombinedBranchS2CAmplificationUsesJointEvalModError(t *testing.T) {
	s2cD := 10.0
	result := NumericalStageLockstep{Checkpoints: []NumericalStageCheckpoint{{
		Name: "s2c", Comparable: true, D: &s2cD,
	}}}
	combinedSNR := setCombinedBranchS2CMetrics(
		&result,
		[]complex128{2, 4},
		[]complex128{3, 7},
	)
	require.NotNil(t, combinedSNR)
	applyNumericalStageTopology(&result, combinedSNR)
	require.Equal(t, numericalmetrics.Finite, combinedSNR.Status)
	require.InDelta(t, math.Sqrt(5), *result.CombinedBranchEvalModErrorRMSE, 1e-15)
	require.InDelta(t, 10/math.Sqrt(5), *result.CombinedBranchS2CAmplification, 1e-15)
	require.Equal(t, "FINITE", result.CombinedBranchS2CAmplificationStatus)

	data, err := json.Marshal(result)
	require.NoError(t, err)
	var encoded map[string]any
	require.NoError(t, json.Unmarshal(data, &encoded))
	require.Contains(t, encoded, "combined_branch_s2c_amplification_factor")
	require.NotContains(t, encoded, "s2c_amplification_factor")
	encodedCheckpoints := encoded["checkpoints"].([]any)
	encodedS2C := encodedCheckpoints[0].(map[string]any)
	require.Nil(t, encodedS2C["a_i"])
	require.Equal(t, "NOT_APPLICABLE_COMBINED_BRANCH", encodedS2C["a_i_status"])
}

func TestStageLockstepHealthyTerminalClassification(t *testing.T) {
	pairs := []numericalStagePair{
		numericalStagePairForTest("input", []complex128{1}, []complex128{1}, true),
		numericalStagePairForTest("final_public_output", []complex128{1}, []complex128{1 + 5e-9}, true),
	}

	result := summarizeNumericalStageLockstep(pairs, []complex128{1})

	require.Empty(t, result.FirstObservable)
	require.Empty(t, result.FirstMaterial)
	require.Equal(t, "CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE", result.Classification)
}

func TestStageLockstepObservableButNonMaterialClassification(t *testing.T) {
	pairs := []numericalStagePair{
		numericalStagePairForTest("input", []complex128{1}, []complex128{1}, true),
		numericalStagePairForTest("final_public_output", []complex128{1}, []complex128{1 + 5e-8}, true),
	}

	result := summarizeNumericalStageLockstep(pairs, []complex128{1})

	require.Equal(t, "final_public_output", result.FirstObservable)
	require.Empty(t, result.FirstMaterial)
	require.Equal(t, "CURRENT_FAST_OBSERVABLE_NON_MATERIAL_DIVERGENCE", result.Classification)
}

func TestStageLockstepFirstMaterialEvalModClassificationIsPreserved(t *testing.T) {
	pairs := []numericalStagePair{
		numericalStagePairForTest("input", []complex128{1}, []complex128{1}, true),
		numericalStagePairForTest("evalmod_real", []complex128{1}, []complex128{1 + 0.01}, true),
	}

	result := summarizeNumericalStageLockstep(pairs, []complex128{1})

	require.Equal(t, "evalmod_real", result.FirstMaterial)
	require.Equal(t, "CURRENT_FAST_FIRST_MATERIAL_EVALMOD_POLYNOMIAL", result.Classification)
}

func TestStageLockstepRequiresComparableFiniteFinalOutputForHealthyClassification(t *testing.T) {
	t.Run("missing final output", func(t *testing.T) {
		result := summarizeNumericalStageLockstep([]numericalStagePair{
			numericalStagePairForTest("input", []complex128{1}, []complex128{1}, true),
		}, []complex128{1})

		require.Equal(t, "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED", result.Classification)
	})

	t.Run("non-comparable final output", func(t *testing.T) {
		result := summarizeNumericalStageLockstep([]numericalStagePair{
			numericalStagePairForTest("input", []complex128{1}, []complex128{1}, true),
			numericalStagePairForTest("final_public_output", nil, nil, false),
		}, []complex128{1})

		require.False(t, stageCheckpoint(t, &result, "final_public_output").Comparable)
		require.Equal(t, "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED", result.Classification)
	})

	t.Run("non-finite final metrics", func(t *testing.T) {
		result := summarizeNumericalStageLockstep([]numericalStagePair{
			numericalStagePairForTest("input", []complex128{1}, []complex128{1}, true),
			numericalStagePairForTest("final_public_output", []complex128{1}, []complex128{complex(math.NaN(), 0)}, true),
		}, []complex128{1})

		require.Equal(t, "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED", result.Classification)
	})
}

func TestAcceptedCanonicalResultRendersConsistentStageClassification(t *testing.T) {
	data, err := os.ReadFile("../../results/FAST-STANDARD-NUMERICAL-FIX-001-GENERATED-POWER-ORDER.json")
	require.NoError(t, err)
	var doc NumericalDocument
	require.NoError(t, json.Unmarshal(data, &doc))
	require.NotNil(t, doc.StageLockstep)
	require.Equal(t, "FAST_STANDARD_NUMERICAL_CLOSE", doc.Classification)
	require.Equal(t, "CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE", doc.StageLockstep.Classification)
	require.InDelta(t, 1.4468089544345387e-10, doc.StageLockstep.FinalFastStandardRMSE, 1e-20)
	require.NotEmpty(t, doc.FastTrials)
	require.NotNil(t, doc.FastTrials[0].BootstrapSNR.SNRDB)
	require.InDelta(t, 144.71074997463654, *doc.FastTrials[0].BootstrapSNR.SNRDB, 1e-12)

	rendered := renderNumericalSummary(doc)
	require.Contains(t, rendered, "Current classification: `CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`")
	require.NotContains(t, rendered, "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED")
}

func numericalStagePairForTest(name string, standard, fast []complex128, comparable bool) numericalStagePair {
	return numericalStagePair{
		fast:     numericalStageSample{name: name, values: fast, comparable: comparable},
		standard: numericalStageSample{name: name, values: standard, comparable: comparable},
	}
}

func TestNumericalSummaryRendersCapacityAndGenuineStandardPowerEvidence(t *testing.T) {
	stage := NumericalStageLockstep{
		PolynomialPlan: &NumericalPolynomialPlan{Branch: "real", Degree: 30, Base: 4, Level: 12, ScaleLog2: 60, ScaleExact: "2^60"},
		GeneratedPowerEvidence: []NumericalGeneratedPower{{
			Branch: "real", ReferenceKind: "Fast-checkout Standard API PowerBasis.GenPower", Power: 2,
			FastLevel: 11, FastScaleLog2: 60, FastMaintainedRows: 4,
			StandardLevel: 11, StandardScaleLog2: 60, StandardAuthorityRows: 12,
			LevelScaleMatch: true, RMSEFastStandard: 1e-8, MaxComplexDiffFastStandard: 1e-7,
		}},
		EvalModCapacityAudit: []NumericalQPrefixCapacity{{
			Checkpoint: "real/evalmod-entry", Level: 12, Rows: 4, PrefixProduct: "12345",
			Degree: 1, MaxAbs: []string{"10", "0"}, StrictFit: true,
		}},
	}
	var output strings.Builder
	renderNumericalStageEvidence(&output, NumericalDocument{StageLockstep: &stage})

	require.Contains(t, output.String(), "Standard-equivalent Fast EvalMod Q-prefix capacity audit")
	require.Contains(t, output.String(), "real/evalmod-entry")
	require.Contains(t, output.String(), "12345")
	require.Contains(t, output.String(), "`[10 0]`")
	require.Contains(t, output.String(), "true")
	require.Contains(t, output.String(), "ordinary Standard CKKS evaluator compiled from the recorded Fast source checkout")
	require.Contains(t, output.String(), "11 / 60.000000")
	require.Contains(t, output.String(), "Complex RMSE Fast-vs-Standard")
	require.Contains(t, output.String(), "1.000000e-08")
}

func syntheticStageCheckpoint(name string, distance, snrDB float64) NumericalStageCheckpoint {
	return NumericalStageCheckpoint{
		Name: name, Comparable: true, D: &distance,
		StageReferenceSNR: syntheticSNR(snrDB),
	}
}

func syntheticSNR(snrDB float64) numericalmetrics.SNR {
	return numericalmetrics.SNR{Status: numericalmetrics.Finite, SNRDB: &snrDB}
}

func stageCheckpoint(t *testing.T, result *NumericalStageLockstep, name string) *NumericalStageCheckpoint {
	t.Helper()
	checkpoint := numericalStageCheckpointByName(result, name)
	require.NotNil(t, checkpoint)
	return checkpoint
}
