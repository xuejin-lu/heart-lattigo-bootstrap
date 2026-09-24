package main

import (
	"encoding/json"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func bigInts(values ...int64) []*big.Int {
	result := make([]*big.Int, len(values))
	for i, value := range values {
		result[i] = big.NewInt(value)
	}
	return result
}

func TestFastCenteredCRTRoundTripOneTwoAndThreeModuli(t *testing.T) {
	cases := []struct {
		name   string
		moduli []*big.Int
		values []int64
	}{
		{"one", bigInts(5), []int64{-2, -1, 0, 1, 2}},
		{"two", bigInts(5, 7), []int64{-17, -16, -1, 0, 1, 16, 17}},
		{"three", bigInts(5, 7, 11), []int64{-192, -191, -1, 0, 1, 191, 192}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, integer := range test.values {
				want := big.NewInt(integer)
				residues, err := FastResiduesFor(want, test.moduli)
				if err != nil {
					t.Fatal(err)
				}
				got, err := FastCenteredCRT(test.moduli, residues)
				if err != nil {
					t.Fatal(err)
				}
				if got.Value.Cmp(want) != 0 {
					t.Fatalf("round trip(%s)=%s, want %s", test.name, got.Value, want)
				}
				if got.Negative != (integer < 0) {
					t.Fatalf("negative=%t for %d", got.Negative, integer)
				}
			}
		})
	}
	if _, err := FastCenteredCRT(bigInts(6, 9), bigInts(1, 2)); err == nil {
		t.Fatal("non-coprime moduli were accepted")
	}
}

func TestFastResidueConsistencyDetectsOneCorruptResidue(t *testing.T) {
	moduli := bigInts(5, 7, 11)
	residues, err := FastResiduesFor(big.NewInt(-73), moduli)
	if err != nil {
		t.Fatal(err)
	}
	if !FastCheckResidueConsistency(big.NewInt(-73), moduli, residues) {
		t.Fatal("valid residues were rejected")
	}
	residues[1].Add(residues[1], big.NewInt(1)).Mod(residues[1], moduli[1])
	if FastCheckResidueConsistency(big.NewInt(-73), moduli, residues) {
		t.Fatal("one corrupted residue was accepted")
	}
}

func TestFastLogicalCongruenceAcceptsDifferentPositiveAndNegativeLifts(t *testing.T) {
	modulus := big.NewInt(35)
	reference := big.NewInt(4)
	for _, k := range []int64{-7, -1, 0, 1, 9} {
		value := new(big.Int).Add(reference, new(big.Int).Mul(big.NewInt(k), modulus))
		if !FastCheckLogicalCongruence(value, reference, modulus) {
			t.Fatalf("valid lift %s rejected", value)
		}
		if k != 0 && value.Cmp(reference) == 0 {
			t.Fatalf("nonzero lift k=%d unexpectedly equals reference", k)
		}
	}
}

func TestFastCapacityHeadroomBoundariesAndZero(t *testing.T) {
	product := big.NewInt(35)
	for _, test := range []struct {
		bound int64
		pass  bool
	}{{10, true}, {17, true}, {18, false}} {
		result, err := FastCapacityEvidence(big.NewInt(test.bound), product)
		if err != nil {
			t.Fatal(err)
		}
		if result.Unique != test.pass {
			t.Fatalf("bound=%d unique=%t, want %t", test.bound, result.Unique, test.pass)
		}
		if result.HeadroomBits == nil {
			t.Fatalf("bound=%d has no finite headroom", test.bound)
		}
		if result.CenteredCapacityBitLength != 5 {
			t.Fatalf("centered capacity bit length=%d, want 5", result.CenteredCapacityBitLength)
		}
	}
	zero, err := FastCapacityEvidence(big.NewInt(0), product)
	if err != nil {
		t.Fatal(err)
	}
	if !zero.Unique || !zero.UnboundedHeadroom || zero.HeadroomBits != nil {
		t.Fatalf("zero-bound evidence=%+v", zero)
	}
	nearBoundary, err := FastCapacityEvidence(big.NewInt(34), big.NewInt(70))
	if err != nil {
		t.Fatal(err)
	}
	exactBoundary, err := FastCapacityEvidence(big.NewInt(35), big.NewInt(70))
	if err != nil {
		t.Fatal(err)
	}
	if !nearBoundary.Unique || exactBoundary.Unique {
		t.Fatalf("strict centered boundary results: near=%t exact=%t", nearBoundary.Unique, exactBoundary.Unique)
	}
}

func TestFastLogicalRescaleTheoremAcrossIndependentChains(t *testing.T) {
	chains := [][]*big.Int{
		bigInts(5, 7, 11, 13),
		bigInts(17, 19, 23),
		bigInts(101, 103, 107),
	}
	for chainIndex, chain := range chains {
		for level := 1; level < len(chain); level++ {
			qLevel := chain[level]
			qPrevious, err := FastLogicalModulus(chain, level-1)
			if err != nil {
				t.Fatal(err)
			}
			qCurrent, err := FastLogicalModulus(chain, level)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range []int64{-213, -1, 0, 5, 217} {
				for _, k := range []int64{-3, -1, 0, 2, 7} {
					reference := big.NewInt(c)
					value := new(big.Int).Add(reference, new(big.Int).Mul(big.NewInt(k), qCurrent))
					y, err := FastLogicalRescaleOracle(value, qLevel)
					if err != nil {
						t.Fatal(err)
					}
					yReference, err := FastLogicalRescaleOracle(reference, qLevel)
					if err != nil {
						t.Fatal(err)
					}
					if !FastCheckLogicalCongruence(y, yReference, qPrevious) {
						t.Fatalf("chain=%d level=%d c=%d k=%d: rescale congruence failed", chainIndex, level, c, k)
					}
				}
			}
		}
	}
	positive, _ := FastLogicalRescaleOracle(big.NewInt(8), big.NewInt(5))
	negative, _ := FastLogicalRescaleOracle(big.NewInt(-8), big.NewInt(5))
	if positive.Int64() != 2 || negative.Int64() != -2 {
		t.Fatalf("nearest signed results: +8/5=%s, -8/5=%s", positive, negative)
	}
}

func TestFastStorageContractionOracleLegalAndIllegalCases(t *testing.T) {
	threeToTwo, err := FastStorageContractionOracle(big.NewInt(17), bigInts(5, 7))
	if err != nil {
		t.Fatal(err)
	}
	if !threeToTwo.Allowed || threeToTwo.RoundTrip.Cmp(big.NewInt(17)) != 0 {
		t.Fatalf("3-to-2 contraction failed: %+v", threeToTwo)
	}
	twoToOne, err := FastStorageContractionOracle(big.NewInt(2), bigInts(5))
	if err != nil {
		t.Fatal(err)
	}
	if !twoToOne.Allowed || twoToOne.RoundTrip.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("2-to-1 contraction failed: %+v", twoToOne)
	}
	illegal, err := FastStorageContractionOracle(big.NewInt(3), bigInts(5))
	if err != nil {
		t.Fatal(err)
	}
	if illegal.Allowed || illegal.RoundTrip != nil {
		t.Fatalf("insufficient target basis contraction was allowed: %+v", illegal)
	}
}

func TestFastModUpCanonicalizationUnifiesDifferentLifts(t *testing.T) {
	oldQ := big.NewInt(35)
	addedLogicalPrime := big.NewInt(101)
	x := new(big.Int).Add(big.NewInt(3), new(big.Int).Mul(big.NewInt(2), oldQ))
	y := new(big.Int).Sub(big.NewInt(3), new(big.Int).Mul(big.NewInt(3), oldQ))
	cx, err := FastModUpCanonicalRepresentative(x, oldQ)
	if err != nil {
		t.Fatal(err)
	}
	cy, err := FastModUpCanonicalRepresentative(y, oldQ)
	if err != nil {
		t.Fatal(err)
	}
	if cx.Cmp(cy) != 0 || cx.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("canonical representatives differ: %s and %s", cx, cy)
	}
	naiveX := new(big.Int).Mod(x, addedLogicalPrime)
	naiveY := new(big.Int).Mod(y, addedLogicalPrime)
	if naiveX.Cmp(naiveY) == 0 {
		t.Fatalf("naive arbitrary-lift extension unexpectedly agrees: %s", naiveX)
	}
	canonicalExtended := new(big.Int).Mod(cx, addedLogicalPrime)
	if canonicalExtended.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("canonical extension=%s, want 3", canonicalExtended)
	}
}

func TestFastMeasurementSummaryFindsEarliestInvariantFailure(t *testing.T) {
	checkpoints := []FastMeasurementCheckpoint{
		{OperationID: "cp-0", Invariants: []FastInvariantCheck{{ID: FastInvariantLogicalCongruence, Status: FastMeasurementPass}}, Availability: []FastMeasurementAvailability{{Measurement: "optional", Status: string(FastMeasurementUnavailable)}}, Semantic: &FastSemanticSummary{Error: &FastDistributionStats{AbsMax: 9}, ThresholdPass: boolPtr(false)}},
		{OperationID: "cp-1", Invariants: []FastInvariantCheck{{ID: FastInvariantResidueConsistency, Status: FastMeasurementPass}}},
		{OperationID: "cp-2", Invariants: []FastInvariantCheck{{ID: FastInvariantCenteredUnique, Status: FastMeasurementFail}}},
		{OperationID: "cp-3", Invariants: []FastInvariantCheck{{ID: FastInvariantScaleTransition, Status: FastMeasurementFail}}},
	}
	summary := FastSummarizeMeasurement(checkpoints)
	if summary.FirstFailedCheckpoint != "cp-2" || summary.FirstFailedInvariant != string(FastInvariantCenteredUnique) {
		t.Fatalf("first failure=(%s,%s)", summary.FirstFailedCheckpoint, summary.FirstFailedInvariant)
	}
	if summary.E2EError != nil {
		t.Fatalf("non-E2E semantic threshold error was reported as E2E: %g", *summary.E2EError)
	}
	endToEnd := FastSummarizeMeasurement([]FastMeasurementCheckpoint{{OperationID: "e2e", Semantic: &FastSemanticSummary{EndToEnd: true, Error: &FastDistributionStats{AbsMax: 0.5}}}})
	if endToEnd.E2EError == nil || *endToEnd.E2EError != 0.5 {
		t.Fatalf("E2E error summary=%v, want 0.5", endToEnd.E2EError)
	}
}

func TestFastMeasurementModesOffLightAndFull(t *testing.T) {
	var expensiveCalls int
	off, err := NewFastMeasurementCollector(FastMeasurementOff, FastMeasurementProvenance{})
	if err != nil {
		t.Fatal(err)
	}
	off.Capture(func(FastMeasurementMode) FastMeasurementCheckpoint {
		expensiveCalls++
		return FastMeasurementCheckpoint{OperationID: "off"}
	})
	if expensiveCalls != 0 || len(off.Run().Checkpoints) != 0 {
		t.Fatal("OFF mode invoked capture work or stored checkpoints")
	}

	build := func(mode FastMeasurementMode) FastMeasurementCheckpoint {
		checkpoint := FastMeasurementCheckpoint{
			OperationID: "mode-checkpoint", Stage: "synthetic", Storage: &FastStorageState{ActiveStorageWidth: 2, MaxAbsXBitLength: 19, MaxAbsXDecimal: "123456", ResidueConsistencyChecked: boolPtr(true)},
			BeforeStorage: &FastStorageState{ActiveStorageWidth: 3, MaxAbsXDecimal: "42", ResidueConsistencyChecked: boolPtr(true)},
			ExactOracles:  &FastExactOracleEvidence{LogicalCongruence: &FastIntegerCheckSummary{CheckedCount: 1}},
			Semantic:      &FastSemanticSummary{Signal: &FastDistributionStats{Count: 1}},
			Invariants: []FastInvariantCheck{
				{ID: FastInvariantCenteredUnique, Status: FastMeasurementPass},
				{ID: FastInvariantLogicalCongruence, Status: FastMeasurementPass},
			},
		}
		if mode == FastMeasurementFull {
			expensiveCalls++
		}
		return checkpoint
	}
	light, err := NewFastMeasurementCollector(FastMeasurementLight, FastMeasurementProvenance{})
	if err != nil {
		t.Fatal(err)
	}
	light.Capture(build)
	lightCheckpoint := light.Run().Checkpoints[0]
	if expensiveCalls != 0 || len(lightCheckpoint.Invariants) != 1 || len(lightCheckpoint.Availability) != 3 {
		t.Fatalf("LIGHT mode result=%+v expensive calls=%d", lightCheckpoint, expensiveCalls)
	}
	if lightCheckpoint.Storage.MaxAbsXDecimal != "" || lightCheckpoint.Storage.ResidueConsistencyChecked != nil || lightCheckpoint.BeforeStorage.MaxAbsXDecimal != "" || lightCheckpoint.BeforeStorage.ResidueConsistencyChecked != nil || lightCheckpoint.ExactOracles != nil || lightCheckpoint.Semantic != nil {
		t.Fatal("LIGHT retained FULL-only reconstruction or semantic details")
	}
	if lightCheckpoint.Sequence != 1 {
		t.Fatalf("collector sequence=%d, want 1", lightCheckpoint.Sequence)
	}
	full, err := NewFastMeasurementCollector(FastMeasurementFull, FastMeasurementProvenance{})
	if err != nil {
		t.Fatal(err)
	}
	full.Capture(build)
	fullCheckpoint := full.Run().Checkpoints[0]
	if expensiveCalls != 1 || len(fullCheckpoint.Invariants) != 2 || len(fullCheckpoint.Availability) != 0 || fullCheckpoint.ExactOracles == nil || fullCheckpoint.Semantic == nil {
		t.Fatalf("FULL mode result=%+v expensive calls=%d", fullCheckpoint, expensiveCalls)
	}
}

func TestFastMeasurementStatisticsDeterministicAndCompact(t *testing.T) {
	values := []float64{1, 2, 3, 4}
	stats, err := FastSummarizeValues(values)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Count != 4 || stats.Min != 1 || stats.Max != 4 || stats.Mean != 2.5 || math.Abs(stats.StdDev-math.Sqrt(1.25)) > 1e-14 || math.Abs(stats.RMS-math.Sqrt(7.5)) > 1e-14 || stats.AbsMax != 4 {
		t.Fatalf("unexpected distribution stats: %+v", stats)
	}
	if stats.P50 != 2.5 || math.Abs(stats.P90-3.7) > 1e-14 || math.Abs(stats.P99-3.97) > 1e-14 || math.Abs(stats.P999-3.997) > 1e-14 {
		t.Fatalf("unexpected quantiles: %+v", stats)
	}
	first, err := json.Marshal(stats)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := json.Marshal(stats)
	if string(first) != string(second) {
		t.Fatalf("statistics serialization is nondeterministic: %s != %s", first, second)
	}
	if string(first) == "" || strings.Contains(string(first), "[1,2,3,4]") {
		t.Fatal("serialized stats unexpectedly retain the raw input vector")
	}
	residualStats, err := FastSummarizeValues([]float64{-2, 1, 4})
	if err != nil {
		t.Fatal(err)
	}
	if residualStats.AbsP50 != 2 || math.Abs(residualStats.AbsP90-3.6) > 1e-14 {
		t.Fatalf("unexpected absolute residual quantiles: %+v", residualStats)
	}
	sample := FastDeterministicSample(values, 3)
	wantSample := []FastNumericSample{{Index: 0, Value: 1}, {Index: 1, Value: 2}, {Index: 3, Value: 4}}
	if !reflect.DeepEqual(sample, wantSample) {
		t.Fatalf("deterministic sample=%+v, want %+v", sample, wantSample)
	}
	if _, err := FastSummarizeValues(nil); err == nil {
		t.Fatal("empty statistics input was accepted")
	}
}

func boolPtr(value bool) *bool { return &value }
