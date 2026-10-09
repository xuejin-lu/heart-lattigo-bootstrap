package main

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestRescaleScaleClosureUsesLogicalQ1(t *testing.T) {
	defaultScale := rlwe.NewScale(math.Exp2(defaultScaleLog2))
	for _, q1 := range []uint64{68719476731, 549755813881} {
		q1Scale := rlwe.NewScale(q1)
		productScale := defaultScale.Mul(q1Scale)
		rescaled := productScale.Div(q1Scale)
		if rescaled.Cmp(defaultScale) != 0 {
			t.Fatalf("product scale / logical q1 = %s, want default scale %s", rescaled.BigInt(), defaultScale.BigInt())
		}
	}
}

func TestCanonicalInputsAndWorkloadHashAreDeterministic(t *testing.T) {
	a, b, c := deterministicInputs(1 << 12)
	if got := hashComplexValues(a); got != canonicalInputSHA256 {
		t.Fatalf("canonical seed input hash = %s, want %s", got, canonicalInputSHA256)
	}
	if len(a) != 1<<12 || len(b) != len(a) || len(c) != len(a) {
		t.Fatalf("unexpected slot lengths: %d, %d, %d", len(a), len(b), len(c))
	}
	if hashWorkload(a, b, c, 68719476731) != hashWorkload(a, b, c, 68719476731) {
		t.Fatal("identical workload hashes differ")
	}
	if hashWorkload(a, b, c, 68719476731) == hashWorkload(a, b, c, 68719476729) {
		t.Fatal("workload hash omitted the q1 encoding scale")
	}
	if cmplx.Abs(a[0]-a[1]) == 0 || cmplx.Abs(b[0]-b[1]) == 0 {
		t.Fatal("input vectors must be deterministic and nonconstant")
	}
}

func TestCompareValuesRejectsEmptyMismatchedAndNonFiniteSamples(t *testing.T) {
	for name, pair := range map[string][2][]complex128{
		"empty":     {nil, nil},
		"length":    {{1, 2}, {1}},
		"nonfinite": {{complex(math.NaN(), 0)}, {0}},
	} {
		if got := compareValues(pair[0], pair[1]); got.Finite {
			t.Errorf("%s comparison was accepted: %+v", name, got)
		}
	}
	if got := compareValues([]complex128{1, 2}, []complex128{1, 2}); !got.Finite || got.MaxComplexError != 0 {
		t.Fatalf("equal nonempty samples should pass: %+v", got)
	}
}

func TestStripDecodedDoesNotMutateRawRun(t *testing.T) {
	original := runEvidence{Checkpoints: []checkpoint{{ID: "rotate", Decoded: []complexValue{{Real: 1, Imag: 2}}}}}
	compact := stripDecoded(original)
	if len(original.Checkpoints[0].Decoded) != 1 {
		t.Fatal("compacting evidence mutated the raw checkpoint slice")
	}
	if len(compact.Checkpoints[0].Decoded) != 0 {
		t.Fatal("compact evidence retained decoded samples")
	}
}

func TestExactOrderedPrefixRejectsMismatchAndEmpty(t *testing.T) {
	if !exactOrderedPrefix([]uint64{11, 13}, []uint64{11, 13, 17}) {
		t.Fatal("valid exact ordered prefix was rejected")
	}
	if exactOrderedPrefix([]uint64{11, 17}, []uint64{11, 13, 17}) {
		t.Fatal("non-prefix was accepted")
	}
	if exactOrderedPrefix(nil, []uint64{11}) {
		t.Fatal("empty prefix was accepted")
	}
}
