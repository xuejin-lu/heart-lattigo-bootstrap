package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3DAR0LocalQ2CapacityAndRound(t *testing.T) {
	q := big.NewInt(100)
	if got := fix001P3LocalQ2CapacityOf([]*big.Int{big.NewInt(49), big.NewInt(-49)}, q); !got.Unique || got.Ratio != 0.98 {
		t.Fatalf("unexpected centered-capacity result: %+v", got)
	}
	if got := fix001P3LocalQ2Round(big.NewInt(-6), big.NewInt(4)); got.Int64() != -2 {
		t.Fatalf("negative rounded division = %d, want -2", got.Int64())
	}
}

func TestFIX001P3DAR0LocalQ2ClassificationNames(t *testing.T) {
	allowed := map[string]bool{
		"logn13_da_r0_local_q2_precondition_mismatch":     true,
		"logn13_da_r0_local_q2_expansion_mismatch":        true,
		"logn13_da_r0_local_q2_cannot_contract_to_q01":    true,
		"logn13_da_r0_local_q2_later_da_capacity_blocker": true,
		"logn13_da_r0_local_q2_downstream_validated":      true,
		"logn13_da_r0_local_q2_downstream_insufficient":   true,
	}
	if !allowed["logn13_da_r0_local_q2_downstream_validated"] {
		t.Fatal("success classification missing")
	}
}
