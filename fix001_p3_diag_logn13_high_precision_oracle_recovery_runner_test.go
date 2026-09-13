package main

import (
	"testing"

	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

func TestFIX001P3HighPrecisionOraclePrecisionSweep(t *testing.T) {
	want := []uint{80, 96, 128, 160}
	if len(fix001P3HighPrecisionOraclePrecisions) != len(want) {
		t.Fatalf("precision count mismatch: got %v want %v", fix001P3HighPrecisionOraclePrecisions, want)
	}
	for i := range want {
		if fix001P3HighPrecisionOraclePrecisions[i] != want[i] {
			t.Fatalf("precision[%d] = %d, want %d", i, fix001P3HighPrecisionOraclePrecisions[i], want[i])
		}
	}
}

func TestFIX001P3HighPrecisionOracleMetric(t *testing.T) {
	expected := []*bignum.Complex{bignum.ToComplex(complex(1, -2), 128)}
	actual := []*bignum.Complex{bignum.ToComplex(complex(1+5e-12, -2-7e-12), 128)}
	metric := fix001P3HighPrecisionMetric(expected, actual, 1e-10)
	if !metric.Pass || metric.MaxComponent < 6e-12 {
		t.Fatalf("unexpected high-precision metric: %+v", metric)
	}
}

func TestFIX001P3HighPrecisionOracleMetadataMatch(t *testing.T) {
	metadata := map[string]interface{}{"level": true, "scale": true, "degree": true, "n": true, "log_dimensions": true, "is_ntt": true, "logical_metadata_match_pass": true}
	if !fix001P3HighPrecisionMetadataPass(metadata) {
		t.Fatal("all-true metadata should pass")
	}
	metadata["logical_metadata_match_pass"] = false
	if fix001P3HighPrecisionMetadataPass(metadata) {
		t.Fatal("mismatched metadata should fail")
	}
}
