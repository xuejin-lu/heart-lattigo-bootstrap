package main

import "testing"

func TestExactOrderedPrefix(t *testing.T) {
	full := []uint64{17, 29, 41}
	for _, tc := range []struct {
		name string
		got  []uint64
		want bool
	}{
		{name: "exact prefix", got: []uint64{17, 29}, want: true},
		{name: "full equality", got: []uint64{17, 29, 41}, want: true},
		{name: "wrong order", got: []uint64{29, 17}, want: false},
		{name: "wrong value", got: []uint64{17, 31}, want: false},
		{name: "too long", got: []uint64{17, 29, 41, 53}, want: false},
		{name: "empty", got: nil, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := exactOrderedPrefix(tc.got, full); got != tc.want {
				t.Fatalf("exactOrderedPrefix(%v, %v) = %t, want %t", tc.got, full, got, tc.want)
			}
		})
	}
}

func TestRotateLeftOracle(t *testing.T) {
	input := []complex128{1 + 2i, 3 + 4i, 5 + 6i, 7 + 8i}
	want := []complex128{3 + 4i, 5 + 6i, 7 + 8i, 1 + 2i}
	got := rotateLeftOracle(input, 1)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rotateLeftOracle[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
