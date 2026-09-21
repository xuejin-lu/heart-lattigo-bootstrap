package main

import "testing"

func TestFIX001P3DoubleAnglePlaintextOracle(t *testing.T) {
	input := []complex128{0.25 + 0.5i, -0.75 + 0.125i}
	constant := 0.1
	got := fix001P3PlaintextDoubleAngle(input, constant)
	for i, value := range input {
		want := 2*value*value - complex(constant, 0)
		if got[i] != want {
			t.Fatalf("index %d: got %v, want %v", i, got[i], want)
		}
	}
}

func TestFIX001P3DoubleAngleRuntimeConstants(t *testing.T) {
	got := fix001P3PlaintextDoubleAngleConstants(0.5, 3)
	want := []float64{0.25, 0.0625, 0.00390625}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("round %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestFIX001P3DoubleAngleVectorClosure(t *testing.T) {
	standard := []complex128{0.2 + 0.1i, -0.4 + 0.3i}
	propagated := []complex128{0.21 + 0.1i, -0.4 + 0.31i}
	fast := []complex128{0.215 + 0.105i, -0.405 + 0.31i}
	total := fix001P3PlaintextVectorDifference(fast, standard)
	decomposed := fix001P3PlaintextVectorAdd(
		fix001P3PlaintextVectorDifference(fast, propagated),
		fix001P3PlaintextVectorDifference(propagated, standard),
	)
	closure := fix001P3PlaintextVectorDifference(total, decomposed)
	for i, value := range closure {
		if value != 0 {
			t.Fatalf("index %d: closure residual %v", i, value)
		}
	}
}
