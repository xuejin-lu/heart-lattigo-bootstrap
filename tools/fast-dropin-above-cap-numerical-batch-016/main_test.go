package main

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuneinsight/lattigo/v6/ring"
)

func TestStrictCapacityRejectsNearBoundary(t *testing.T) {
	product := big.NewInt(100)
	require.True(t, strictCapacity(big.NewInt(49), product))
	require.False(t, strictCapacity(big.NewInt(50), product), "2B == S_Q must fail")
	require.False(t, strictCapacity(big.NewInt(51), product))
	require.False(t, strictCapacity(big.NewInt(-1), product))
}

func TestCenteredCRTLiftUsesOnlyRequestedPrefixRows(t *testing.T) {
	q := []uint64{17, 19, 23, 29, 31}
	const rows = 4
	poly := ring.NewPoly(1, rows-1)
	want := big.NewInt(-1234)
	for row := 0; row < rows; row++ {
		qi := new(big.Int).SetUint64(q[row])
		residue := new(big.Int).Mod(new(big.Int).Set(want), qi)
		if residue.Sign() < 0 {
			residue.Add(residue, qi)
		}
		poly.Coeffs[row][0] = residue.Uint64()
	}
	got, product, err := centeredCRT(poly, q, rows, 0)
	require.NoError(t, err)
	require.Equal(t, want, got)
	require.True(t, strictCapacity(new(big.Int).Abs(new(big.Int).Set(got)), product))
	require.Len(t, poly.Coeffs, rows)
}

func TestEncodedInputBoundIsPositiveForFiniteInput(t *testing.T) {
	bound := encodedInputBound([]complex128{1.0 / 16, -1.0/32 + 1.0/64i}, 1<<45)
	require.Positive(t, bound.Sign())
	require.True(t, strictCapacity(bound, new(big.Int).Lsh(big.NewInt(1), 170)))
}

func TestFrozenLevel5BoundsFitActualQ0123AndInputQ0(t *testing.T) {
	params, err := frozenParameters()
	require.NoError(t, err)
	require.Equal(t, 5, params.MaxLevel())
	require.Len(t, params.Q(), 6)
	a, b := deterministicInputs(1 << logSlots)
	boundA := encodedInputBound(a, 1<<scaleLog2)
	boundB := encodedInputBound(b, 1<<scaleLog2)
	bounds, err := deriveBounds(params, boundA, boundB)
	require.NoError(t, err)
	q0 := new(big.Int).SetUint64(params.Q()[0])
	require.True(t, strictCapacity(boundA, q0))
	require.True(t, strictCapacity(boundB, q0))
	for _, bound := range []*big.Int{bounds.add, bounds.mul, bounds.rescale, bounds.rotate} {
		require.True(t, strictCapacity(bound, prefixProduct(params.Q(), 4)))
	}
}
