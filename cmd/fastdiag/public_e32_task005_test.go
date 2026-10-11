package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTask005BaselineIsPinnedAndExhausted(t *testing.T) {
	root, err := findPrimaryRootFromWorkingDirectory()
	require.NoError(t, err)
	baseline, _, err := loadPublicE32Task005Baseline(root)
	require.NoError(t, err)
	require.Equal(t, "CURRENT_LOGN16_FAST_SPEED_VERIFIED", baseline.Classification)
	require.Equal(t, 16, baseline.ReservedCalls)
	require.Equal(t, 16, baseline.ActualCalls)
	require.Equal(t, e32Task005PrimaryCommit, baseline.PrimaryCommit)
	require.Equal(t, standardFastBaselineCommit, baseline.StandardCommit)
	require.Equal(t, e32LogN16ProductionFastCommit, baseline.FastCommit)
	require.InDelta(t, 6425.27575, baseline.StandardWarmMedianMS, 0.001)
	require.InDelta(t, 663.878208, baseline.FastWarmMedianMS, 0.001)
	require.InDelta(t, 4.09656347139061e-8, baseline.FastVsStandardRMSE, 1e-15)
	require.InDelta(t, 4.345682976228596e-7, baseline.FastVsStandardMax, 1e-14)
}

func TestTask005ReferenceVectorsMatchCommittedOutputHashesWhenProvided(t *testing.T) {
	vectorsPath := os.Getenv("FASTDIAG_TASK005_FAST_VECTORS")
	if vectorsPath == "" {
		t.Skip("set FASTDIAG_TASK005_FAST_VECTORS to validate the preserved 42 MB Task005 vector artifact")
	}
	root, err := findPrimaryRootFromWorkingDirectory()
	require.NoError(t, err)
	manifest := publicE32FixtureManifest{
		Profile: "logn16-e32-public-native", BackendCommit: e32LogN16ProductionFastCommit,
		ConfigSHA256: logN16ConfigSHA, InputSHA256: logN16InputSHA,
		WorkloadSHA256: logN16WorkloadSHA, QPSHA256: logN16QPSHA,
	}
	info, err := validateTask005FastReference(root, vectorsPath, manifest)
	require.NoError(t, err)
	require.Equal(t, 6, len(info.Vectors.BootstrapOutputs))
	require.Len(t, info.Vectors.Checkpoints, 8)
	require.NotEmpty(t, info.VectorsSHA256)
	require.NotEmpty(t, info.ResultSHA256)
}

func findPrimaryRootFromWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findPrimaryRoot(wd)
}
