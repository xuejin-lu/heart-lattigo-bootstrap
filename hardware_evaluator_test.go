package main

import (
	"testing"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	ckkspoly "github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	commonpoly "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

func TestHardwareEvaluatorScalarOps(t *testing.T) {
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            11,
		LogQ:            []int{50, 45, 45, 45, 45, 45},
		LogP:            []int{50},
		LogDefaultScale: 45,
	})
	if err != nil {
		t.Fatal(err)
	}

	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()
	rlk := kgen.GenRelinearizationKeyNew(sk)
	evk := rlwe.NewMemEvaluationKeySet(rlk)
	refEval := ckks.NewEvaluator(params, evk)
	btp := &SlimBootstrapperHW{Ref: &bootstrapping.Evaluator{
		Parameters: bootstrapping.Parameters{
			ResidualParameters:      params,
			BootstrappingParameters: params,
		},
		Evaluator:      refEval,
		EvaluationKeys: &bootstrapping.EvaluationKeys{MemEvaluationKeySet: evk},
	}}
	hwEval := newHardwareCKKSEvaluator(btp)

	encoder := ckks.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)
	values := make([]complex128, params.MaxSlots())
	for i := range values {
		values[i] = complex(float64(i%7)/16, float64(i%5)/32)
	}
	pt := ckks.NewPlaintext(params, params.MaxLevel())
	if err := encoder.Encode(values, pt); err != nil {
		t.Fatal(err)
	}
	ct, err := encryptor.EncryptNew(pt)
	if err != nil {
		t.Fatal(err)
	}

	check := func(name string, got, want *rlwe.Ciphertext) {
		t.Helper()
		if got.Level() != want.Level() || got.Degree() != want.Degree() {
			t.Fatalf("%s metadata mismatch", name)
		}
		for d := range got.Value {
			for limb := range got.Value[d].Coeffs {
				for i := range got.Value[d].Coeffs[limb] {
					if got.Value[d].Coeffs[limb][i] != want.Value[d].Coeffs[limb][i] {
						t.Fatalf("%s mismatch d=%d limb=%d i=%d got=%x want=%x", name, d, limb, i, got.Value[d].Coeffs[limb][i], want.Value[d].Coeffs[limb][i])
					}
				}
			}
		}
	}

	for _, scalar := range []complex128{1i, -1i, 0.125, -0.25 + 0.5i} {
		got := rlwe.NewCiphertext(params, 1, ct.Level())
		want := rlwe.NewCiphertext(params, 1, ct.Level())
		if err := hwEval.Mul(ct, scalar, got); err != nil {
			t.Fatal(err)
		}
		if err := refEval.Mul(ct, scalar, want); err != nil {
			t.Fatal(err)
		}
		check("mul scalar", got, want)
	}

	for _, scalar := range []complex128{-1, 0.125, -0.25 + 0.5i} {
		got := rlwe.NewCiphertext(params, 1, ct.Level())
		want := rlwe.NewCiphertext(params, 1, ct.Level())
		if err := hwEval.Add(ct, scalar, got); err != nil {
			t.Fatal(err)
		}
		if err := refEval.Add(ct, scalar, want); err != nil {
			t.Fatal(err)
		}
		check("add scalar", got, want)
	}

	gotMul := rlwe.NewCiphertext(params, 1, ct.Level())
	wantMul := rlwe.NewCiphertext(params, 1, ct.Level())
	if err := hwEval.MulRelin(ct, ct, gotMul); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulRelin(ct, ct, wantMul); err != nil {
		t.Fatal(err)
	}
	check("mul relin", gotMul, wantMul)

	gotTensor := rlwe.NewCiphertext(params, 2, ct.Level())
	wantTensor := rlwe.NewCiphertext(params, 2, ct.Level())
	if err := hwEval.Mul(ct, ct, gotTensor); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Mul(ct, ct, wantTensor); err != nil {
		t.Fatal(err)
	}
	check("mul tensor", gotTensor, wantTensor)

	gotInPlaceMul := ct.CopyNew()
	wantInPlaceMul := ct.CopyNew()
	if err := hwEval.Mul(gotInPlaceMul, ct, gotInPlaceMul); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Mul(wantInPlaceMul, ct, wantInPlaceMul); err != nil {
		t.Fatal(err)
	}
	check("mul tensor in-place", gotInPlaceMul, wantInPlaceMul)

	gotRelin := rlwe.NewCiphertext(params, 1, gotTensor.Level())
	wantRelin := rlwe.NewCiphertext(params, 1, wantTensor.Level())
	if err := hwEval.Relinearize(gotTensor, gotRelin); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Relinearize(wantTensor, wantRelin); err != nil {
		t.Fatal(err)
	}
	check("relinearize", gotRelin, wantRelin)

	gotRelinInPlace := gotTensor.CopyNew()
	wantRelinInPlace := wantTensor.CopyNew()
	if err := hwEval.Relinearize(gotRelinInPlace, gotRelinInPlace); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Relinearize(wantRelinInPlace, wantRelinInPlace); err != nil {
		t.Fatal(err)
	}
	check("relinearize in-place", gotRelinInPlace, wantRelinInPlace)

	gotRescale := rlwe.NewCiphertext(params, 1, gotMul.Level()-1)
	wantRescale := rlwe.NewCiphertext(params, 1, wantMul.Level()-1)
	if err := hwEval.Rescale(gotMul, gotRescale); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Rescale(wantMul, wantRescale); err != nil {
		t.Fatal(err)
	}
	check("rescale", gotRescale, wantRescale)

	gotRescaleDeg2 := rlwe.NewCiphertext(params, 2, gotTensor.Level()-1)
	wantRescaleDeg2 := rlwe.NewCiphertext(params, 2, wantTensor.Level()-1)
	if err := hwEval.Rescale(gotTensor, gotRescaleDeg2); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Rescale(wantTensor, wantRescaleDeg2); err != nil {
		t.Fatal(err)
	}
	check("rescale degree2", gotRescaleDeg2, wantRescaleDeg2)

	gotPow4 := rlwe.NewCiphertext(params, 1, gotRescale.Level())
	wantPow4 := rlwe.NewCiphertext(params, 1, wantRescale.Level())
	if err := hwEval.MulRelin(gotRescale, gotRescale, gotPow4); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulRelin(wantRescale, wantRescale, wantPow4); err != nil {
		t.Fatal(err)
	}
	check("pow4 mul relin", gotPow4, wantPow4)

	gotMixedTensor := rlwe.NewCiphertext(params, 2, gotRescale.Level())
	wantMixedTensor := rlwe.NewCiphertext(params, 2, wantRescale.Level())
	if err := hwEval.Mul(gotRescale, ct, gotMixedTensor); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Mul(wantRescale, ct, wantMixedTensor); err != nil {
		t.Fatal(err)
	}
	check("mixed-level mul tensor", gotMixedTensor, wantMixedTensor)

	gotMixedMulRelin := rlwe.NewCiphertext(params, 1, gotRescale.Level())
	wantMixedMulRelin := rlwe.NewCiphertext(params, 1, wantRescale.Level())
	if err := hwEval.MulRelin(gotRescale, ct, gotMixedMulRelin); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulRelin(wantRescale, ct, wantMixedMulRelin); err != nil {
		t.Fatal(err)
	}
	check("mixed-level mul relin", gotMixedMulRelin, wantMixedMulRelin)

	gotMixedAdd := gotRescale.CopyNew()
	wantMixedAdd := wantRescale.CopyNew()
	if err := hwEval.Add(gotMixedAdd, ct, gotMixedAdd); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Add(wantMixedAdd, ct, wantMixedAdd); err != nil {
		t.Fatal(err)
	}
	check("mixed-level add", gotMixedAdd, wantMixedAdd)

	pbHW := commonpoly.NewPowerBasis(ct.CopyNew(), bignum.Chebyshev)
	pbRef := commonpoly.NewPowerBasis(ct.CopyNew(), bignum.Chebyshev)
	if err := pbHW.GenPower(4, false, hwEval); err != nil {
		t.Fatal(err)
	}
	if err := pbRef.GenPower(4, false, refEval); err != nil {
		t.Fatal(err)
	}
	check("chebyshev power 4", pbHW.Value[4], pbRef.Value[4])
	if err := pbHW.GenPower(3, false, hwEval); err != nil {
		t.Fatal(err)
	}
	if err := pbRef.GenPower(3, false, refEval); err != nil {
		t.Fatal(err)
	}
	check("chebyshev power 3", pbHW.Value[3], pbRef.Value[3])

	gotMTA := ct.CopyNew()
	wantMTA := ct.CopyNew()
	coeff := bignum.NewComplex().SetComplex128(0.125)
	if err := hwEval.MulThenAdd(ct, coeff, gotMTA); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, coeff, wantMTA); err != nil {
		t.Fatal(err)
	}
	check("mul then add scalar", gotMTA, wantMTA)

	gotMTAZero := rlwe.NewCiphertext(params, 1, ct.Level())
	wantMTAZero := rlwe.NewCiphertext(params, 1, ct.Level())
	*gotMTAZero.MetaData = *ct.MetaData
	*wantMTAZero.MetaData = *ct.MetaData
	gotMTAZero.Scale = ct.Scale
	wantMTAZero.Scale = ct.Scale
	if err := hwEval.MulThenAdd(ct, coeff, gotMTAZero); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, coeff, wantMTAZero); err != nil {
		t.Fatal(err)
	}
	check("mul then add scalar zero", gotMTAZero, wantMTAZero)

	gotMTAZeroQuot := rlwe.NewCiphertext(params, 1, ct.Level())
	wantMTAZeroQuot := rlwe.NewCiphertext(params, 1, ct.Level())
	*gotMTAZeroQuot.MetaData = *ct.MetaData
	*wantMTAZeroQuot.MetaData = *ct.MetaData
	gotMTAZeroQuot.Scale = ct.Scale.Mul(rlwe.NewScale(params.Q()[ct.Level()]))
	wantMTAZeroQuot.Scale = ct.Scale.Mul(rlwe.NewScale(params.Q()[ct.Level()]))
	if err := hwEval.MulThenAdd(ct, coeff, gotMTAZeroQuot); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, coeff, wantMTAZeroQuot); err != nil {
		t.Fatal(err)
	}
	check("mul then add scalar zero quotient", gotMTAZeroQuot, wantMTAZeroQuot)

	gotMTAZeroLower := rlwe.NewCiphertext(params, 1, 3)
	wantMTAZeroLower := rlwe.NewCiphertext(params, 1, 3)
	*gotMTAZeroLower.MetaData = *ct.MetaData
	*wantMTAZeroLower.MetaData = *ct.MetaData
	gotMTAZeroLower.Scale = ct.Scale.Mul(rlwe.NewScale(params.Q()[3]))
	wantMTAZeroLower.Scale = ct.Scale.Mul(rlwe.NewScale(params.Q()[3]))
	if err := hwEval.MulThenAdd(ct, coeff, gotMTAZeroLower); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, coeff, wantMTAZeroLower); err != nil {
		t.Fatal(err)
	}
	check("mul then add scalar lower quotient", gotMTAZeroLower, wantMTAZeroLower)

	gotMTAZeroSqrt := rlwe.NewCiphertext(params, 1, 3)
	wantMTAZeroSqrt := rlwe.NewCiphertext(params, 1, 3)
	*gotMTAZeroSqrt.MetaData = *ct.MetaData
	*wantMTAZeroSqrt.MetaData = *ct.MetaData
	gotMTAZeroSqrt.Scale = ct.Scale.Mul(rlwe.NewScale(params.Q()[3]))
	gotMTAZeroSqrt.Scale.Value.Sqrt(&gotMTAZeroSqrt.Scale.Value)
	gotMTAZeroSqrt.Scale = gotMTAZeroSqrt.Scale.Mul(ct.Scale)
	wantMTAZeroSqrt.Scale = gotMTAZeroSqrt.Scale
	if err := hwEval.MulThenAdd(ct, coeff, gotMTAZeroSqrt); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, coeff, wantMTAZeroSqrt); err != nil {
		t.Fatal(err)
	}
	check("mul then add scalar sqrt quotient", gotMTAZeroSqrt, wantMTAZeroSqrt)

	gotMTAQuot := ct.CopyNew()
	wantMTAQuot := ct.CopyNew()
	level := ct.Level()
	qScale := rlwe.NewScale(params.RingQ().AtLevel(level).ModuliChain()[level])
	qScaleInt := qScale.BigInt()
	if err := hwEval.Mul(gotMTAQuot, qScaleInt, gotMTAQuot); err != nil {
		t.Fatal(err)
	}
	gotMTAQuot.Scale = gotMTAQuot.Scale.Mul(qScale)
	if err := refEval.Mul(wantMTAQuot, qScaleInt, wantMTAQuot); err != nil {
		t.Fatal(err)
	}
	wantMTAQuot.Scale = wantMTAQuot.Scale.Mul(qScale)
	if err := hwEval.MulThenAdd(ct, 0.125, gotMTAQuot); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, 0.125, wantMTAQuot); err != nil {
		t.Fatal(err)
	}
	check("mul then add scalar quotient", gotMTAQuot, wantMTAQuot)

	gotMTADeg2 := gotTensor.CopyNew()
	wantMTADeg2 := wantTensor.CopyNew()
	if err := hwEval.MulThenAdd(ct, coeff, gotMTADeg2); err != nil {
		t.Fatal(err)
	}
	if err := refEval.MulThenAdd(ct, coeff, wantMTADeg2); err != nil {
		t.Fatal(err)
	}
	check("mul then add degree2", gotMTADeg2, wantMTADeg2)

	gotAddDeg2 := gotTensor.CopyNew()
	wantAddDeg2 := wantTensor.CopyNew()
	if err := hwEval.Add(gotAddDeg2, -1, gotAddDeg2); err != nil {
		t.Fatal(err)
	}
	if err := refEval.Add(wantAddDeg2, -1, wantAddDeg2); err != nil {
		t.Fatal(err)
	}
	check("add scalar degree2", gotAddDeg2, wantAddDeg2)

	polCoeffs := []float64{0.1, 0.2, -0.05, 0.025, 0.01, -0.005, 0.0025, -0.00125, 0.000625, -0.0003, 0.00015, -0.00007, 0.00003, -0.000015, 0.000007, -0.000003, 0.000001}
	for len(polCoeffs) < 31 {
		polCoeffs = append(polCoeffs, polCoeffs[len(polCoeffs)-1]*-0.5)
	}
	for i := 1; i < len(polCoeffs); i += 2 {
		polCoeffs[i] = 0
	}
	pol := bignum.NewPolynomial(bignum.Chebyshev, polCoeffs, [2]float64{-1, 1})
	hwPolyEval := commonpoly.Evaluator[*bignum.Complex]{
		Evaluator:         hwEval,
		CoefficientGetter: hardwareCoefficientGetter{values: make([]*bignum.Complex, params.MaxSlots())},
	}
	refPolyEval := ckkspoly.NewEvaluator(params, refEval)
	hwCKKSPolyEval := ckkspoly.Evaluator{
		Parameters: params,
		Evaluator:  hwPolyEval,
	}
	polyTargetScale := ct.Scale.Mul(rlwe.NewScale(params.Q()[ct.Level()-1]))
	gotPoly, err := hwCKKSPolyEval.Evaluate(ct.CopyNew(), pol, polyTargetScale)
	if err != nil {
		t.Fatal(err)
	}
	wantPoly, err := refPolyEval.Evaluate(ct.CopyNew(), pol, polyTargetScale)
	if err != nil {
		t.Fatal(err)
	}
	check("polynomial", gotPoly, wantPoly)
}
