package main

import "testing"

func TestFIX001P3Q056GuardConfigProfile(t *testing.T) {
	cfg := DefaultBootstrapConfig()
	control := q056BuildConfig(cfg, 55)
	widened := q056BuildConfig(cfg, 56)
	if len(control.Q0) != 1 || control.Q0[0] != 55 {
		t.Fatalf("control q0 = %#v, want [55]", control.Q0)
	}
	if len(widened.Q0) != 1 || widened.Q0[0] != 56 {
		t.Fatalf("widened q0 = %#v, want [56]", widened.Q0)
	}
	if widened.QSlotsToCoeffs[0] != cfg.QSlotsToCoeffs[0] || len(widened.QSlotsToCoeffs) != len(cfg.QSlotsToCoeffs) {
		t.Fatalf("q1 profile changed: got %#v, want %#v", widened.QSlotsToCoeffs, cfg.QSlotsToCoeffs)
	}
}

func TestFIX001P3Q056GuardCandidateSchedule(t *testing.T) {
	boundaries := []psRescaleGuardBoundary{{ID: "G0-rescale"}, {ID: "G1-rescale"}, {ID: "G2-rescale"}, {ID: "G3-rescale"}, {ID: "F0-final-rescale"}}
	got := psRescaleGuardScheduleName([]int{1, 0, 0, 0, 1}, boundaries, "W3")
	want := "W3-G0-rescale=2^1+F0-final-rescale=2^1"
	if got != want {
		t.Fatalf("schedule = %q, want %q", got, want)
	}
}
