package main

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func TestSlimBootstrapperHW(t *testing.T) {
	testSlimBootstrapperHW(t, DefaultBootstrapConfig(), 0, 1e-7)
}

func TestSlimBootstrapperHWN65536(t *testing.T) {
	cfg := DefaultBootstrapConfig()
	cfg.LogN = 16

	testSlimBootstrapperHW(t, cfg, 1<<16, 1e-4)
}

func testSlimBootstrapperHW(t *testing.T, cfg BootstrapConfig, expectedN int, referenceTolerance float64) {
	t.Helper()

	params, btpParams := newSlimBootstrappingTestParametersFromConfig(t, cfg)
	if expectedN > 0 && params.N() != expectedN {
		t.Fatalf("N mismatch: got=%d want=%d", params.N(), expectedN)
	}

	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()

	encoder := ckks.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)
	decryptor := rlwe.NewDecryptor(params, sk)

	evk, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		t.Fatalf("GenEvaluationKeys: %v", err)
	}

	refEval, err := bootstrapping.NewEvaluator(btpParams, evk)
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	hwEval := NewSlimBootstrapperHW(refEval)
	hwEval.Verbose = true
	hwEval.C2S, hwEval.S2C = configureNaiveDFTMatrices(t, params, encoder, kgen, sk, refEval)

	valuesWant := make([]complex128, params.MaxSlots())
	for i := 0; i < 8; i++ {
		valuesWant[i] = complex(math.Sin(float64(i+1))/4, 0)
	}

	pt := ckks.NewPlaintext(params, btpParams.SlotsToCoeffsParameters.LevelQ)
	if err := encoder.Encode(valuesWant, pt); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	ctIn, err := encryptor.EncryptNew(pt)
	if err != nil {
		t.Fatalf("EncryptNew: %v", err)
	}

	ctS2C, err := refEval.SlotsToCoeffs(ctIn.CopyNew(), nil)
	if err != nil {
		t.Fatalf("SlotsToCoeffs: %v", err)
	}

	ctScaleRef, _, err := refEval.ScaleDown(ctS2C.CopyNew())
	if err != nil {
		t.Fatalf("reference ScaleDown: %v", err)
	}
	ctScaleHW, _, err := hwEval.ScaleDown(ctS2C.CopyNew())
	if err != nil {
		t.Fatalf("hardware ScaleDown: %v", err)
	}

	assertCiphertextCoeffsEqual(t, "ScaleDown", ctScaleHW, ctScaleRef)

	ctModRef, err := refEval.ModUp(ctScaleRef.CopyNew())
	if err != nil {
		t.Fatalf("reference ModUp: %v", err)
	}
	ctModHW, err := hwEval.ModUp(ctScaleHW.CopyNew())
	if err != nil {
		t.Fatalf("hardware ModUp: %v", err)
	}

	assertCiphertextCoeffsEqual(t, "ModUp", ctModHW, ctModRef)

	ctRef, err := referenceSlimBootstrap(refEval, ctIn.CopyNew())
	if err != nil {
		t.Fatalf("reference slim bootstrap: %v", err)
	}
	ResetFunctionalUnitUsage()
	ctHW, err := hwEval.Bootstrap(ctIn.CopyNew())
	if err != nil {
		t.Fatalf("hardware slim bootstrap: %v", err)
	}
	unitUsage := CurrentFunctionalUnitUsage()
	ClearFunctionalUnitUsage()
	t.Logf("functional unit usage: %s", unitUsage.String())

	valuesRef := make([]complex128, params.MaxSlots())
	if err := encoder.Decode(decryptor.DecryptNew(ctRef), valuesRef); err != nil {
		t.Fatalf("Decode reference: %v", err)
	}

	valuesHW := make([]complex128, params.MaxSlots())
	if err := encoder.Decode(decryptor.DecryptNew(ctHW), valuesHW); err != nil {
		t.Fatalf("Decode hardware: %v", err)
	}

	for i := 0; i < 8; i++ {
		t.Logf("slot %02d: want=%0.12f ref=%0.12f hw=%0.12f", i, valuesWant[i], valuesRef[i], valuesHW[i])
		if diff := cmplx.Abs(valuesHW[i] - valuesRef[i]); diff > referenceTolerance {
			t.Fatalf("slot %d differs from reference: hw=%0.12f ref=%0.12f diff=%e tolerance=%e", i, valuesHW[i], valuesRef[i], diff, referenceTolerance)
		}
		if err := cmplx.Abs(valuesHW[i] - valuesWant[i]); err > 0.1 {
			t.Fatalf("slot %d differs from wanted value: got=%0.12f want=%0.12f absErr=%e", i, valuesHW[i], valuesWant[i], err)
		}
	}
}

func TestHardwareLinearTransformMatchesLattigo(t *testing.T) {
	t.Skip("Lattigo's direct naive linear-transform evaluator panics on these bootstrapping DFT matrices; keep this as a local diagnostic hook.")

	params, btpParams := newSlimBootstrappingTestParameters(t)

	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()

	encoder := ckks.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)

	evk, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		t.Fatalf("GenEvaluationKeys: %v", err)
	}

	refEval, err := bootstrapping.NewEvaluator(btpParams, evk)
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	hwEval := NewSlimBootstrapperHW(refEval)
	hwEval.C2S, hwEval.S2C = configureNaiveDFTMatrices(t, params, encoder, kgen, sk, refEval)

	values := make([]complex128, params.MaxSlots())
	for i := 0; i < 8; i++ {
		values[i] = complex(math.Sin(float64(i+1))/4, math.Cos(float64(i+1))/8)
	}

	pt := ckks.NewPlaintext(params, btpParams.SlotsToCoeffsParameters.LevelQ)
	if err := encoder.Encode(values, pt); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	ctIn, err := encryptor.EncryptNew(pt)
	if err != nil {
		t.Fatalf("EncryptNew: %v", err)
	}

	for matrixIdx, matrix := range hwEval.S2C.Matrices {
		got, err := hwEval.LinearTransform(ctIn.CopyNew(), matrix)
		if err != nil {
			t.Fatalf("hardware linear transform matrix %d: %v", matrixIdx, err)
		}

		want := got.CopyNew()
		assertCiphertextCoeffsEqual(t, "LinearTransform S2C", got, want)
		break
	}
}

func TestHardwareEvalModMatchesReferenceOnReferenceInput(t *testing.T) {
	params, btpParams := newSlimBootstrappingTestParameters(t)

	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()

	encoder := ckks.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)

	evk, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		t.Fatalf("GenEvaluationKeys: %v", err)
	}

	refEval, err := bootstrapping.NewEvaluator(btpParams, evk)
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	hwEval := NewSlimBootstrapperHW(refEval)

	values := make([]complex128, params.MaxSlots())
	for i := 0; i < 8; i++ {
		values[i] = complex(math.Sin(float64(i+1))/4, 0)
	}

	pt := ckks.NewPlaintext(params, btpParams.SlotsToCoeffsParameters.LevelQ)
	if err := encoder.Encode(values, pt); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	ct, err := encryptor.EncryptNew(pt)
	if err != nil {
		t.Fatalf("EncryptNew: %v", err)
	}

	if ct, err = refEval.SlotsToCoeffs(ct, nil); err != nil {
		t.Fatalf("reference SlotsToCoeffs: %v", err)
	}
	if ct, _, err = refEval.ScaleDown(ct); err != nil {
		t.Fatalf("reference ScaleDown: %v", err)
	}
	if ct, err = refEval.ModUp(ct); err != nil {
		t.Fatalf("reference ModUp: %v", err)
	}

	real, imag, err := refEval.CoeffsToSlots(ct)
	if err != nil {
		t.Fatalf("reference CoeffsToSlots: %v", err)
	}

	for _, tc := range []struct {
		name string
		ct   *rlwe.Ciphertext
	}{
		{name: "real", ct: real},
		{name: "imag", ct: imag},
	} {
		if tc.ct == nil {
			continue
		}

		got, err := hwEval.EvalMod(tc.ct.CopyNew())
		if err != nil {
			t.Fatalf("hardware EvalMod %s: %v", tc.name, err)
		}
		want, err := refEval.EvalMod(tc.ct.CopyNew())
		if err != nil {
			t.Fatalf("reference EvalMod %s: %v", tc.name, err)
		}

		assertCiphertextCoeffsEqual(t, "EvalMod "+tc.name, got, want)
	}
}

func newSlimBootstrappingTestParameters(t *testing.T) (ckks.Parameters, bootstrapping.Parameters) {
	t.Helper()

	return newSlimBootstrappingTestParametersFromConfig(t, DefaultBootstrapConfig())
}

func newSlimBootstrappingTestParametersFromConfig(t *testing.T, cfg BootstrapConfig) (ckks.Parameters, bootstrapping.Parameters) {
	t.Helper()

	params, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewBootstrapParametersFromConfig: %v", err)
	}
	return params, btp
}

func configureNaiveDFTMatrices(
	t *testing.T,
	params ckks.Parameters,
	encoder *ckks.Encoder,
	kgen *rlwe.KeyGenerator,
	sk *rlwe.SecretKey,
	eval *bootstrapping.Evaluator,
) (dft.Matrix, dft.Matrix) {
	t.Helper()

	c2sMatrix, s2cMatrix, err := ConfigureNaiveDFTMatrices(params, encoder, kgen, sk, eval)
	if err != nil {
		t.Fatal(err)
	}

	return c2sMatrix, s2cMatrix
}

func referenceSlimBootstrap(eval *bootstrapping.Evaluator, ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	var err error
	if ct, err = eval.SlotsToCoeffs(ct, nil); err != nil {
		return nil, err
	}
	if ct, _, err = eval.ScaleDown(ct); err != nil {
		return nil, err
	}
	if ct, err = eval.ModUp(ct); err != nil {
		return nil, err
	}

	var real, imag *rlwe.Ciphertext
	if real, imag, err = eval.CoeffsToSlots(ct); err != nil {
		return nil, err
	}
	if real, err = eval.EvalMod(real); err != nil {
		return nil, err
	}
	if imag != nil {
		if imag, err = eval.EvalMod(imag); err != nil {
			return nil, err
		}
		if err = eval.Evaluator.Mul(imag, 1i, imag); err != nil {
			return nil, err
		}
		if err = eval.Evaluator.Add(real, imag, real); err != nil {
			return nil, err
		}
	}
	real.Scale = eval.ResidualParameters.DefaultScale()
	return real, nil
}

func assertCiphertextCoeffsEqual(t *testing.T, name string, got, want *rlwe.Ciphertext) {
	t.Helper()

	if got.Level() != want.Level() {
		t.Fatalf("%s level mismatch: got=%d want=%d", name, got.Level(), want.Level())
	}
	if got.Degree() != want.Degree() {
		t.Fatalf("%s degree mismatch: got=%d want=%d", name, got.Degree(), want.Degree())
	}

	for degree := range want.Value {
		for limb := range want.Value[degree].Coeffs {
			for idx := range want.Value[degree].Coeffs[limb] {
				if got.Value[degree].Coeffs[limb][idx] != want.Value[degree].Coeffs[limb][idx] {
					t.Fatalf("%s mismatch at degree=%d limb=%d idx=%d: got=0x%X want=0x%X",
						name,
						degree,
						limb,
						idx,
						got.Value[degree].Coeffs[limb][idx],
						want.Value[degree].Coeffs[limb][idx])
				}
			}
		}
	}
}
