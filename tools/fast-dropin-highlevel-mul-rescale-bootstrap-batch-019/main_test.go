package main

import (
	"errors"
	"math/big"
	"testing"
)

var stageAllocationSink []byte

func TestBatch019ExactBoundsAndCapacity(t *testing.T) {
	a, b, c := deterministicInputs(4096)
	q := []uint64{
		36028797018652673, 549755731969, 549756026881, 549755486209,
		549756174337, 1152921504606830593, 1152921504606748673,
		1152921504606994433, 1152921504606683137, 1152921504606601217,
		1152921504606584833, 1152921504607191041, 1152921504607223809,
		72057594037616641, 72057594038321153, 72057594037370881, 72057594037338113,
	}
	got, err := deriveBounds(a, b, c, q, 8192)
	if err != nil {
		t.Fatal(err)
	}
	want := boundPlan{
		InputA: "2348557866436", InputB: "1804822278899", InputC: "76957544167327219",
		Add: "4153380145335", MulRelin: "2618441203534382746900245490606080",
		Rescale: "2271135713118062", Q0123Product: "5986308565615587353347023386369277282933624144412673",
		Q0: "36028797018652673", MulDivisor: "1152921504606830593",
	}
	if got != want {
		t.Fatalf("exact bound plan mismatch:\n got: %#v\nwant: %#v", got, want)
	}
	if !strictCapacity(new(big.Int).SetUint64(1), new(big.Int).SetUint64(3)) {
		t.Fatal("strict capacity should accept 2B < S")
	}
	if strictCapacity(new(big.Int).SetUint64(2), new(big.Int).SetUint64(4)) {
		t.Fatal("strict capacity must reject equality")
	}
}

func TestBatch019DefaultScaleMismatchIsNotHidden(t *testing.T) {
	defaultScale := new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2)
	defaultProduct := new(big.Int).Mul(new(big.Int).Set(defaultScale), defaultScale)
	q5 := new(big.Int).SetUint64(1152921504606830593)
	wrong := new(big.Int).Quo(defaultProduct, q5)
	if wrong.Cmp(defaultScale) == 0 {
		t.Fatal("two default-scale inputs unexpectedly close to default after q5 Rescale")
	}
	correctedProduct := new(big.Int).Mul(new(big.Int).Set(defaultScale), q5)
	if new(big.Int).Quo(correctedProduct, q5).Cmp(defaultScale) != 0 {
		t.Fatal("per-input C Scale=q5 did not exactly close the public Scale equation")
	}
}

func TestBatch019IntegerSqrtCeiling(t *testing.T) {
	for _, tc := range []struct{ input, want int64 }{{0, 0}, {1, 1}, {2, 1}, {3, 1}, {4, 2}, {15, 3}, {16, 4}, {17, 4}} {
		got := integerSqrt(big.NewInt(tc.input))
		if got.Int64() != tc.want {
			t.Fatalf("integerSqrt(%d)=%s, want %d", tc.input, got, tc.want)
		}
	}
}

func TestBatch019MetricReportsSNR(t *testing.T) {
	ref := []complex128{1 + 2i, 3 - 4i}
	actual := []complex128{1 + 1e-8 + 2i, 3 - 4i + 1e-8i}
	got := compareValues(ref, actual)
	if !got.Finite || got.SNRDB == nil || got.SNRState != "finite" || got.SignalPower <= 0 || got.ErrorPower <= 0 {
		t.Fatalf("incomplete SNR evidence: %#v", got)
	}
	pair, err := pairedValues(ref, actual)
	if err != nil || pair.SNRDB == nil || !pair.Pass {
		t.Fatalf("paired SNR/gate evidence: pair=%#v err=%v", pair, err)
	}
}

func TestCaptureStageRecordsWallAndOperationAllocationsOnError(t *testing.T) {
	var result runEvidence
	wantErr := errors.New("expected stage failure")
	err := captureStage(&result, "probe", func() error {
		stageAllocationSink = make([]byte, 1024)
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("captureStage error = %v, want %v", err, wantErr)
	}
	if len(result.StageSamples) != 1 {
		t.Fatalf("recorded %d stage samples, want 1", len(result.StageSamples))
	}
	sample := result.StageSamples[0]
	if sample.Name != "probe" || sample.WallNS < 0 || sample.AllocBytes < 1024 || sample.Allocs == 0 {
		t.Fatalf("incomplete stage sample: %#v", sample)
	}
}
