package numericalmetrics

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestCompareKnownSignalErrorPair(t *testing.T) {
	metric := Compare([]complex128{1}, []complex128{0.5})
	if metric.Status != Finite || metric.SNRDB == nil || math.Abs(*metric.SNRDB-20*math.Log10(2)) > 1e-12 {
		t.Fatalf("unexpected SNR result: %+v", metric)
	}
	if metric.SignalPower == nil || *metric.SignalPower != 1 || metric.SignalRMS == nil || *metric.SignalRMS != 1 {
		t.Fatalf("unexpected signal metrics: %+v", metric)
	}
	if metric.NoisePower == nil || *metric.NoisePower != 0.25 || metric.NoiseRMSE == nil || *metric.NoiseRMSE != 0.5 {
		t.Fatalf("unexpected noise metrics: %+v", metric)
	}
}

func TestCompareIsScaleInvariant(t *testing.T) {
	reference := []complex128{1 + 2i, -3 + 4i}
	output := []complex128{1.25 + 1.5i, -2.5 + 3.75i}
	base := Compare(reference, output)
	scaledReference := make([]complex128, len(reference))
	scaledOutput := make([]complex128, len(output))
	for i := range reference {
		scaledReference[i] = reference[i] * 7
		scaledOutput[i] = output[i] * 7
	}
	scaled := Compare(scaledReference, scaledOutput)
	if base.SNRDB == nil || scaled.SNRDB == nil || math.Abs(*base.SNRDB-*scaled.SNRDB) > 1e-12 {
		t.Fatalf("SNR changed under common scaling: base=%+v scaled=%+v", base, scaled)
	}
}

func TestExactEqualityHasStrictJSONInfinityStatus(t *testing.T) {
	metric := Compare([]complex128{2 - 3i}, []complex128{2 - 3i})
	if metric.Status != PositiveInfinity || metric.SNRDB != nil || metric.NoiseRMSE == nil || *metric.NoiseRMSE != 0 {
		t.Fatalf("unexpected exact-equality result: %+v", metric)
	}
	encoded, err := json.Marshal(metric)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "NaN") || strings.Contains(string(encoded), "Inf") || !strings.Contains(string(encoded), `"snr_db":null`) {
		t.Fatalf("non-strict JSON or missing nullable SNR value: %s", encoded)
	}
}

func TestZeroSignalIsExplicitlyUndefined(t *testing.T) {
	metric := Compare([]complex128{0, 0}, []complex128{0, 1i})
	if metric.Status != UndefinedZeroSignal || metric.SNRDB != nil || metric.SignalPower == nil || *metric.SignalPower != 0 {
		t.Fatalf("unexpected zero-signal result: %+v", metric)
	}
}

func TestNonComparableVectorsAreExplicit(t *testing.T) {
	for _, metric := range []SNR{Compare(nil, nil), Compare([]complex128{1}, nil), Compare([]complex128{complex(math.NaN(), 0)}, []complex128{0})} {
		if metric.Status != NotComparable || metric.Reason == "" || metric.SNRDB != nil {
			t.Fatalf("unexpected non-comparable result: %+v", metric)
		}
	}
}
