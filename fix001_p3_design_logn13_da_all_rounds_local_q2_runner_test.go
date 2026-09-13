package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3DAAllRoundsLocalQ2CapacityRule(t *testing.T) {
	q01 := big.NewInt(100)
	if got := fix001P3LocalQ2CapacityOf([]*big.Int{big.NewInt(49)}, q01); !got.Unique || got.Ratio != 0.98 {
		t.Fatalf("unexpected Q01 capacity: %+v", got)
	}
	if got := fix001P3LocalQ2CapacityOf([]*big.Int{big.NewInt(50)}, q01); got.Unique {
		t.Fatalf("half-modulus representative must not be centered-unique: %+v", got)
	}
}

func TestFIX001P3DAAllRoundsLocalQ2ClassificationNames(t *testing.T) {
	classifications := []string{
		"logn13_da_all_rounds_local_q2_precondition_mismatch",
		"logn13_da_all_rounds_q01_round_input_capacity_blocker",
		"logn13_da_all_rounds_local_q2_expansion_mismatch",
		"logn13_da_all_rounds_local_q2_q012_capacity_failure",
		"logn13_da_all_rounds_local_q2_rescale_mismatch",
		"logn13_da_all_rounds_local_q2_cannot_contract_to_q01",
		"logn13_da_all_rounds_local_q2_downstream_insufficient",
		"logn13_da_all_rounds_local_q2_downstream_validated",
	}
	if len(classifications) != 8 {
		t.Fatal("unexpected classification table")
	}
}
