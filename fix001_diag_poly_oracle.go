package main

import commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"

// fix001ExpectedPower evaluates the plaintext Chebyshev recurrence used by
// the diagnostic power comparison. It deliberately includes T0(x)=1 so that
// equal-factor splits such as T2 use 2*T1*T1-T0.
func fix001ExpectedPower(n int, x []complex128, memo map[int][]complex128) []complex128 {
	if value, ok := memo[n]; ok {
		return value
	}
	if n == 0 {
		memo[n] = make([]complex128, len(x))
		for i := range memo[n] {
			memo[n][i] = 1
		}
		return memo[n]
	}
	if n == 1 {
		memo[n] = append([]complex128(nil), x...)
		return memo[n]
	}
	a, b := commonpolynomial.SplitDegree(n)
	va := fix001ExpectedPower(a, x, memo)
	vb := fix001ExpectedPower(b, x, memo)
	c := a - b
	if c < 0 {
		c = -c
	}
	vc := fix001ExpectedPower(c, x, memo)
	out := make([]complex128, len(x))
	for i := range out {
		out[i] = 2*va[i]*vb[i] - vc[i]
	}
	memo[n] = out
	return out
}
