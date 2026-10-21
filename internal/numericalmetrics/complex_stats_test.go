package numericalmetrics

import (
	"math"
	"testing"
)

func TestSummarizeComplexMagnitudeTailsUsesNearestRank(t *testing.T) {
	values := []complex128{1, 2, 3, 4, 5, 6, 7, 8, 9, 100}
	got := SummarizeComplexMagnitudeTails(values)
	if got.Status != "FINITE" || got.Count != 10 || got.P95 != 100 || got.P99 != 100 || got.P999 != 100 || got.MaxAbs != 100 {
		t.Fatalf("unexpected nearest-rank tails: %+v", got)
	}
	got = SummarizeComplexMagnitudeTails([]complex128{1, 2, 3, 4})
	if got.P95 != 4 || got.P99 != 4 || got.P999 != 4 || got.MaxAbs != 4 {
		t.Fatalf("small-vector tails should use nearest rank: %+v", got)
	}
}

func TestSummarizeComplexMagnitudeTailsRejectsInvalidVectors(t *testing.T) {
	if got := SummarizeComplexMagnitudeTails(nil); got.Status != "NOT_COMPARABLE" {
		t.Fatalf("empty vector status = %q, want NOT_COMPARABLE", got.Status)
	}
	if got := SummarizeComplexMagnitudeTails([]complex128{1, complex(math.NaN(), 0)}); got.Status != "NOT_COMPARABLE" {
		t.Fatalf("non-finite vector status = %q, want NOT_COMPARABLE", got.Status)
	}
}
