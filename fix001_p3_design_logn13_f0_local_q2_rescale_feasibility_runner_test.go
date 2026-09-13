package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3F0LocalQ2RescaleCapacity(t *testing.T) {
	q01 := bigIntForTest("100")
	if got := fix001P3LocalQ2CapacityOf([]*big.Int{bigIntForTest("49")}, q01); !got.Unique || got.Ratio != 0.98 {
		t.Fatalf("unexpected Q01 capacity: %+v", got)
	}
}

func TestFIX001P3F0LocalQ2RescaleRound(t *testing.T) {
	if got := fix001P3LocalQ2Round(bigIntForTest("-6"), bigIntForTest("4")); got.Int64() != -2 {
		t.Fatalf("round(-6/4) = %s, want -2", got)
	}
	if got := fix001P3LocalQ2Round(bigIntForTest("6"), bigIntForTest("4")); got.Int64() != 2 {
		t.Fatalf("round(6/4) = %s, want 2", got)
	}
}

func bigIntForTest(value string) *big.Int {
	result, ok := new(big.Int).SetString(value, 10)
	if !ok {
		panic("invalid test integer")
	}
	return result
}
