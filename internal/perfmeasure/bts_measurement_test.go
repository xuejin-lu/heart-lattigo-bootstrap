package perfmeasure

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestSummarizeBTSRepeatedSamplesUsesSampleStandardDeviation(t *testing.T) {
	stats := SummarizeBTSRepeatedSamples([]int64{5, 1, 4, 2, 3})
	if stats.Count != 5 || stats.Mean != 3 || stats.Median != 3 || stats.Min != 1 || stats.Max != 5 {
		t.Fatalf("unexpected five-sample summary: %+v", stats)
	}
	if math.Abs(stats.SampleStandardDeviation-math.Sqrt(2.5)) > 1e-12 {
		t.Fatalf("unexpected sample standard deviation: %+v", stats)
	}
}

func TestSummarizeBTSRepeatedSamplesUsesMidpointForEvenMedian(t *testing.T) {
	stats := SummarizeBTSRepeatedSamples([]int64{8, 2, 4, 6})
	if stats.Count != 4 || stats.Mean != 5 || stats.Median != 5 || stats.Min != 2 || stats.Max != 8 {
		t.Fatalf("unexpected four-sample summary: %+v", stats)
	}
}

func TestBTS009OptionalParametersDoNotChangeHistoricalZeroValueJSON(t *testing.T) {
	encoded, err := json.Marshal(BTSMeasurementParameters{})
	if err != nil {
		t.Fatal(err)
	}
	for _, taskOnlyField := range []string{"max_level", "q_count", "p_count", "mod1_level", "mod1_scale", "mod1_chebyshev_degree", "mod1_ps_depth", "mod1_double_angle"} {
		if strings.Contains(string(encoded), `"`+taskOnlyField+`"`) {
			t.Fatalf("unset BTS-009 field %q changed historical JSON shape: %s", taskOnlyField, encoded)
		}
	}
}
