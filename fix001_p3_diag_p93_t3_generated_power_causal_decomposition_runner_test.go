package main

import (
	"math/cmplx"
	"testing"
)

func TestFIX001P3T3ChebyshevRecurrence(t *testing.T) {
	t1 := []complex128{0.25 + 0.1i, -0.4 + 0.2i}
	t2 := fix001P3T3CausalChebyshevT2(t1)
	t3 := fix001P3T3CausalChebyshevT3(t1, t2)
	for i := range t1 {
		wantT2 := 2*t1[i]*t1[i] - 1
		wantT3 := 2*wantT2*t1[i] - t1[i]
		if cmplx.Abs(t2[i]-wantT2) > 1e-15 || cmplx.Abs(t3[i]-wantT3) > 1e-15 {
			t.Fatalf("index %d: got T2=%v T3=%v, want T2=%v T3=%v", i, t2[i], t3[i], wantT2, wantT3)
		}
	}
}

func TestFIX001P3T3CausalVectorClosure(t *testing.T) {
	h := []complex128{0.1 + 0.2i, -0.3 + 0.4i}
	p := []complex128{0.11 + 0.2i, -0.3 + 0.41i}
	oh := []complex128{0.12 + 0.21i, -0.29 + 0.39i}
	op := []complex128{0.13 + 0.22i, -0.28 + 0.4i}
	input := fix001P3PlaintextVectorDifference(op, oh)
	implementation := fix001P3PlaintextVectorDifference(fix001P3PlaintextVectorDifference(p, op), fix001P3PlaintextVectorDifference(h, oh))
	observed := fix001P3PlaintextVectorDifference(p, h)
	closure := fix001P3PlaintextVectorDifference(observed, fix001P3PlaintextVectorAdd(input, implementation))
	for i, value := range closure {
		if value != 0 {
			t.Fatalf("index %d: closure residual %v", i, value)
		}
	}
}

func TestFIX001P3T3BalancedFactors(t *testing.T) {
	q := uint64(1152921504607223809)
	factors := fix001P3T3CausalFactors(q)
	left := factors["left"].(uint64)
	right := factors["right"].(uint64)
	if left == 0 || right == 0 {
		t.Fatalf("invalid factors: %#v", factors)
	}
	if left > right*2 || right > left*2 {
		t.Fatalf("factors are not balanced: %d, %d", left, right)
	}
}
