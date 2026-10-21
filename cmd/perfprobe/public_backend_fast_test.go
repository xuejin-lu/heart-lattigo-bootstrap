//go:build perf_fast

package main

import "testing"

func TestFixedSigmaNoisePhaseMeasurementsReflectsCompatibleEvidence(t *testing.T) {
	type phase struct {
		Name       string
		ElapsedNS  int64
		AllocBytes uint64
		Allocs     uint64
	}
	type result struct{ Phases []phase }
	want := []fixedSigmaNoisePhaseEvidence{
		{Name: "sampling_hash_moments", ElapsedNS: 10, AllocBytes: 20, Allocs: 30},
		{Name: "coefficient_to_qprefix_ntt", ElapsedNS: 40, AllocBytes: 50, Allocs: 60},
	}
	got := fixedSigmaNoisePhaseMeasurements(result{Phases: []phase{
		{Name: "sampling_hash_moments", ElapsedNS: 10, AllocBytes: 20, Allocs: 30},
		{Name: "coefficient_to_qprefix_ntt", ElapsedNS: 40, AllocBytes: 50, Allocs: 60},
	}})
	if len(got) != len(want) {
		t.Fatalf("mapped phase count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("mapped phase %d = %+v, want %+v", index, got[index], want[index])
		}
	}
	if got := fixedSigmaNoisePhaseMeasurements(struct{}{}); got != nil {
		t.Fatalf("missing phase field mapped to %+v", got)
	}
}

func TestFixedSigmaNoisePhaseMeasurementsRejectsIncompatibleFieldTypes(t *testing.T) {
	type badPhase struct {
		Name       string
		ElapsedNS  uint64
		AllocBytes uint64
		Allocs     uint64
	}
	if got := fixedSigmaNoisePhaseMeasurements(struct{ Phases []badPhase }{Phases: []badPhase{{}}}); got != nil {
		t.Fatalf("incompatible phase fields mapped to %+v", got)
	}
	if got := fixedSigmaNoisePhaseMeasurements(struct{ Phases []int }{Phases: []int{1}}); got != nil {
		t.Fatalf("non-struct phase entries mapped to %+v", got)
	}
}
