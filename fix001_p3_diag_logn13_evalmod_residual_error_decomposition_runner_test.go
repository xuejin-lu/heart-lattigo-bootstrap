package main

import "testing"

func TestFIX001P3EvalModResidualDecompositionClassification(t *testing.T) {
	checks := []struct {
		name      string
		hPS, hDA  bool
		hBoth     bool
		wantClass string
	}{
		{"ps", true, false, false, "logn13_evalmod_residual_ps_input_blocker"},
		{"da", false, true, false, "logn13_evalmod_residual_da_arithmetic_blocker"},
		{"either", true, true, true, "logn13_evalmod_residual_multiple_single_fix_options"},
		{"joint", false, false, true, "logn13_evalmod_residual_joint_ps_da_blocker"},
		{"floor", false, false, false, "logn13_evalmod_residual_mathematical_chain_floor"},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			got, _ := fix001P3EvalModResidualClassify(check.hPS, check.hDA, check.hBoth, false, true)
			if got != check.wantClass {
				t.Fatalf("classification = %q, want %q", got, check.wantClass)
			}
		})
	}
}

func TestFIX001P3EvalModResidualImplementationMismatchPrecedesCounterfactual(t *testing.T) {
	got, _ := fix001P3EvalModResidualClassify(true, true, true, true, true)
	if got != "logn13_evalmod_residual_fast_implementation_mismatch" {
		t.Fatalf("classification = %q, want implementation mismatch", got)
	}
}

func TestFIX001P3EvalModResidualVectorClosure(t *testing.T) {
	a := []complex128{1 + 2i, -3 + 4i}
	b := []complex128{0.5 - 1i, 2 + 3i}
	c := []complex128{-2 + 0.25i, 1 - 2i}
	d := []complex128{4 - 1.25i, -2 + 0.5i}
	left := fix001P3EvalModResidualAdd(fix001P3EvalModResidualAdd(a, b), fix001P3EvalModResidualAdd(c, d))
	right := fix001P3EvalModResidualAdd(fix001P3EvalModResidualAdd(a, b), fix001P3EvalModResidualAdd(c, d))
	metric := fix001P3EvalModResidualMetric(left, right, fix001P3EvalModResidualClosureTolerance)
	if !metric.Pass || metric.MaxComponent != 0 {
		t.Fatalf("closure metric = %#v, want exact zero", metric)
	}
}
