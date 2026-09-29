package main

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
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
