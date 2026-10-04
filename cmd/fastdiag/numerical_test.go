package main

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
)

func TestFastStandardCanonicalInputFingerprint(t *testing.T) {
	require.Equal(t, fastStandardInputSHA256, fastStandardInputFingerprint(fastStandardP93Values()))
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
