package main

import (
	"math/big"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

func TestFIX001P3TargetScaleNearestIntegerEvidence(t *testing.T) {
	evidence, multiplier, matched := targetScaleScaleEvidence(rlwe.NewScale(8), rlwe.NewScale(25))
	if multiplier.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("nearest target-scale integer = %s, want 3", multiplier)
	}
	if evidence.M != "3" || !evidence.NearestIntegerPositive {
		t.Fatalf("unexpected scale evidence: %+v", evidence)
	}
	if matched.Cmp(rlwe.NewScale(24)) != 0 {
		t.Fatalf("matched scale = %s, want 24", matched.Value.Text('f', -1))
	}
}

func TestFIX001P3CapacityDomainNTTCRTRequiresINTT(t *testing.T) {
	params, _, err := NewBootstrapParametersFromConfig(DefaultBootstrapConfig())
	if err != nil {
		t.Fatal(err)
	}
	ringQ := params.RingQ().AtLevel(1)
	coefficient := ring.NewPoly(ringQ.N(), 1)
	for index := range coefficient.Coeffs[0] {
		value := uint64((index*17)%100000 + 11)
		coefficient.Coeffs[0][index] = value % ringQ.SubRings[0].Modulus
		coefficient.Coeffs[1][index] = value % ringQ.SubRings[1].Modulus
	}
	montgomery := ring.NewPoly(ringQ.N(), 1)
	for limb := 0; limb < 2; limb++ {
		ringQ.SubRings[limb].MForm(coefficient.Coeffs[limb], montgomery.Coeffs[limb])
	}
	ntt := ring.NewPoly(ringQ.N(), 1)
	if err := fastckks.FastPartialNTT(ringQ, montgomery, ntt); err != nil {
		t.Fatal(err)
	}
	nttNonMontgomery := ntt.CopyNew()
	for limb := 0; limb < 2; limb++ {
		ringQ.SubRings[limb].IMForm(nttNonMontgomery.Coeffs[limb], nttNonMontgomery.Coeffs[limb])
	}
	roundTrip := ring.NewPoly(ringQ.N(), 1)
	if err := fastckks.FastPartialINTT(ringQ, ntt, roundTrip); err != nil {
		t.Fatal(err)
	}
	for limb := 0; limb < 2; limb++ {
		ringQ.SubRings[limb].IMForm(roundTrip.Coeffs[limb], roundTrip.Coeffs[limb])
	}
	if len(coefficient.Coeffs[0]) != len(roundTrip.Coeffs[0]) {
		t.Fatal("round-trip length changed")
	}
	for limb := 0; limb < 2; limb++ {
		for index := range coefficient.Coeffs[limb] {
			if coefficient.Coeffs[limb][index] != roundTrip.Coeffs[limb][index] {
				t.Fatalf("round-trip mismatch at limb=%d index=%d", limb, index)
			}
		}
	}
	directDiffers := false
	for index := range coefficient.Coeffs[0] {
		direct := targetScaleCenteredFromResidues(ringQ.SubRings[0].Modulus, ringQ.SubRings[1].Modulus, nttNonMontgomery.Coeffs[0][index], nttNonMontgomery.Coeffs[1][index])
		corrected := targetScaleCenteredFromResidues(ringQ.SubRings[0].Modulus, ringQ.SubRings[1].Modulus, coefficient.Coeffs[0][index], coefficient.Coeffs[1][index])
		if direct.Cmp(corrected) != 0 {
			directDiffers = true
			break
		}
	}
	if !directDiffers {
		t.Fatal("direct CRT of NTT rows unexpectedly matched coefficient-domain CRT for every index")
	}
}

func TestFIX001P3TargetScaleCapacityRejectsProspectiveWrap(t *testing.T) {
	data := targetScaleCenteredData{
		Q0:      big.NewInt(5),
		Q1:      big.NewInt(7),
		Q01:     big.NewInt(35),
		Q01Half: big.NewInt(17),
		Values:  [][]*big.Int{{big.NewInt(10)}, {big.NewInt(0)}},
	}
	if evidence := targetScaleCapacity(data, big.NewInt(2)); evidence.Pass || evidence.OutsideCount != 1 {
		t.Fatalf("capacity evidence = %+v, want one prospective outside coefficient", evidence)
	}
}
