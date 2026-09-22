package main

import (
	"math/cmplx"
	"testing"
)

func TestFIX001P3T3ScheduleExpectedProductAndT6(t *testing.T) {
	left := []complex128{complex(0.25, -0.5), complex(-0.75, 0.125)}
	right := []complex128{complex(-0.5, 0.25), complex(0.125, 0.75)}
	product := fix001P3T3ScheduleExpectedProduct(left, right)
	if len(product) != len(left) || product[0] != left[0]*right[0] || product[1] != left[1]*right[1] {
		t.Fatalf("unexpected product: %#v", product)
	}
	t3 := []complex128{complex(0.25, 0.5), complex(-0.125, 0.75)}
	t6 := fix001P3T3ScheduleExpectedT6(t3)
	for i := range t3 {
		want := 2*t3[i]*t3[i] - 1
		if cmplx.Abs(t6[i]-want) != 0 {
			t.Fatalf("T6[%d] = %v, want %v", i, t6[i], want)
		}
	}
}

func TestFIX001P3T3ScheduleProfileAndRecoveryContract(t *testing.T) {
	if fix001P3T3ScheduleRecoveryFrac != 0.1 {
		t.Fatalf("recovery fraction = %g, want 0.1", fix001P3T3ScheduleRecoveryFrac)
	}
	if fix001P3T3ScheduleSecondary != fix001P3T3CausalProductionSecondary {
		t.Fatalf("secondary commit changed: %s", fix001P3T3ScheduleSecondary)
	}
	if fix001P3T3ScheduleDiffSHA != fix001P3T3CausalProductionDiffSHA {
		t.Fatalf("secondary dirty fingerprint changed: %s", fix001P3T3ScheduleDiffSHA)
	}
}
