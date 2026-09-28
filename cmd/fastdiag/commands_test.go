package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnavailableHistoricalRefOmitsEmptyRescaleShares(t *testing.T) {
	doc := CompareDocument{
		Profile:   "p93-q55",
		Baseline:  RefTrace{RequestedRef: "old", Status: hooksUnavailable},
		Candidate: RefTrace{RequestedRef: "new", Status: "READY"},
	}

	got := renderCompareSummary(doc)
	require.Contains(t, got, "No matched event deltas were computed.")
	require.NotContains(t, got, "Rescale time shares")
	require.NotContains(t, got, "0.000 ms")
}
