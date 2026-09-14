package main

import "testing"

func TestFIX001P3QuantizationAwareEvalModClassification(t *testing.T) {
	pass := func(value bool) interface{} { return &PSGlobalMetric{Pass: value} }
	cases := map[string]map[string]interface{}{
		"C_PS":        {"public_like": pass(true)},
		"C_DA":        {"public_like": pass(false)},
		"C_BOTH_CKKS": {"public_like": pass(false)},
		"C_MATH":      {"public_like": pass(false)},
	}
	classification, _ := fix001P3QuantizationAwareClassification(cases, &PSGlobalMetric{})
	if classification != "logn13_evalmod_residual_ps_input_blocker" {
		t.Fatalf("classification = %q", classification)
	}
}

func TestFIX001P3QuantizationAwareEvalModRepresentationFloor(t *testing.T) {
	cases := map[string]map[string]interface{}{
		"C_PS":        {"public_like": &PSGlobalMetric{Pass: false}},
		"C_DA":        {"public_like": &PSGlobalMetric{Pass: false}},
		"C_BOTH_CKKS": {"public_like": &PSGlobalMetric{Pass: false}},
		"C_MATH":      {"public_like": &PSGlobalMetric{Pass: true}},
	}
	classification, _ := fix001P3QuantizationAwareClassification(cases, &PSGlobalMetric{})
	if classification != "logn13_evalmod_residual_ckks_representation_floor" {
		t.Fatalf("classification = %q", classification)
	}
}

func TestFIX001P3QuantizationAwareEvalModVectorClosure(t *testing.T) {
	a := []complex128{1 + 2i, -3 + 4i}
	b := []complex128{2 - 1i, 5 + 2i}
	c := []complex128{-1 + 3i, 4 - 2i}
	d := []complex128{0.5 - 1i, -2 + 0.25i}
	e := []complex128{3 + 1i, -1 - 2i}
	total := fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(a, b), c), d), e)
	closure := fix001P3QuantizationAwareSub(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(fix001P3QuantizationAwareAdd(a, b), c), d), e), total)
	metric := fix001P3QuantizationAwareMetric(make([]complex128, len(closure)), closure, fix001P3QuantizationAwareClosureTolerance)
	if !metric.Pass || metric.MaxComponent != 0 {
		t.Fatalf("closure metric = %+v", metric)
	}
}
