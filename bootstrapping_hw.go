package main

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"sort"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	lt "github.com/tuneinsight/lattigo/v6/circuits/ckks/lintrans"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	ckkspoly "github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/ring/ringqp"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

// SlimBootstrapperHW evaluates the slim CKKS bootstrapping order while routing
// the low-level modulus/arithmetic steps through the hardware golden units.
//
// The current functional-unit set covers NTT/INTT, base extension by centered
// broadcast, element-wise modular arithmetic, rotations/key-switching, and
// rescale. DFT matrix evaluation and EvalMod polynomial constants are therefore
// delegated to Lattigo's bootstrapping evaluator until matching plaintext-mul
// and polynomial-evaluation units are added.
type SlimBootstrapperHW struct {
	Ref              *bootstrapping.Evaluator
	S2C              dft.Matrix
	C2S              dft.Matrix
	ForceHardwareDFT bool
	Verbose          bool
}

func NewSlimBootstrapperHW(ref *bootstrapping.Evaluator) *SlimBootstrapperHW {
	return &SlimBootstrapperHW{Ref: ref, S2C: ref.S2CDFTMatrix, C2S: ref.C2SDFTMatrix}
}

// Bootstrap evaluates the slim bootstrapping order:
// SlotsToCoeffs -> ScaleDown -> ModUp -> CoeffsToSlots -> EvalMod.
func (btp *SlimBootstrapperHW) Bootstrap(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if btp == nil || btp.Ref == nil {
		return nil, fmt.Errorf("nil slim bootstrapper")
	}

	btp.trace("Step 1/6 SlotsToCoeffs: hardware DFT")
	ct, err := btp.SlotsToCoeffs(ctIn.CopyNew(), nil)
	if err != nil {
		return nil, fmt.Errorf("slots-to-coeffs: %w", err)
	}

	btp.trace("Step 2/6 ScaleDown: hardware scalar/rescale path")
	if ct, _, err = btp.ScaleDown(ct); err != nil {
		return nil, fmt.Errorf("hardware scale-down: %w", err)
	}

	btp.trace("Step 3/6 ModUp: hardware centered broadcast + NTT")
	if ct, err = btp.ModUp(ct); err != nil {
		return nil, fmt.Errorf("hardware mod-up: %w", err)
	}

	var ctReal, ctImag *rlwe.Ciphertext
	btp.trace("Step 4/6 CoeffsToSlots: hardware inverse DFT")
	if ctReal, ctImag, err = btp.CoeffsToSlots(ct); err != nil {
		return nil, fmt.Errorf("coeffs-to-slots: %w", err)
	}

	btp.trace("Step 5/6 EvalMod(real): modular reduction")
	if ctReal, err = btp.EvalMod(ctReal); err != nil {
		return nil, fmt.Errorf("eval-mod real: %w", err)
	}

	if ctImag != nil {
		btp.trace("Step 5/6 EvalMod(imag): modular reduction")
		if ctImag, err = btp.EvalMod(ctImag); err != nil {
			return nil, fmt.Errorf("eval-mod imag: %w", err)
		}

		btp.trace("Step 6/6 Recombine real + i*imag")
		hwEval := newHardwareCKKSEvaluator(btp)
		if err = hwEval.Mul(ctImag, 1i, ctImag); err != nil {
			return nil, fmt.Errorf("recombine imag: %w", err)
		}

		if err = hwEval.Add(ctReal, ctImag, ctReal); err != nil {
			return nil, fmt.Errorf("recombine real+imag: %w", err)
		}
	}

	ctReal.Scale = btp.Ref.ResidualParameters.DefaultScale()
	return ctReal, nil
}

func (btp *SlimBootstrapperHW) SlotsToCoeffs(ctReal, ctImag *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if ctImag != nil {
		hwEval := newHardwareCKKSEvaluator(btp)
		if err := hwEval.Mul(ctImag, 1i, ctImag); err != nil {
			return nil, err
		}
		if err := hwEval.Add(ctImag, ctReal, ctImag); err != nil {
			return nil, err
		}
		return btp.dft(ctImag, btp.S2C, "SlotsToCoeffs")
	}
	return btp.dft(ctReal, btp.S2C, "SlotsToCoeffs")
}

func (btp *SlimBootstrapperHW) CoeffsToSlots(ctIn *rlwe.Ciphertext) (ctReal, ctImag *rlwe.Ciphertext, err error) {
	zV, err := btp.dft(ctIn, btp.C2S, "CoeffsToSlots")
	if err != nil {
		return nil, nil, err
	}

	if btp.C2S.Format != dft.RepackImagAsReal && btp.C2S.Format != dft.SplitRealAndImag {
		return zV, nil, nil
	}

	ctReal, err = btp.Conjugate(zV)
	if err != nil {
		return nil, nil, err
	}

	tmp := rlwe.NewCiphertext(btp.Ref.BootstrappingParameters, 1, zV.Level())
	*tmp.MetaData = *zV.MetaData
	tmpHW := EvalSub(btp.Ref.BootstrappingParameters.RingQ().AtLevel(zV.Level()), rlweToHWCiphertext(btp.Ref.BootstrappingParameters.RingQ(), zV), rlweToHWCiphertext(btp.Ref.BootstrappingParameters.RingQ(), ctReal))
	for i := range tmpHW.Value {
		tmp.Value[i] = tmpHW.Value[i]
	}

	if btp.C2S.LogSlots == btp.Ref.BootstrappingParameters.LogMaxSlots() {
		ctImag = tmp
	} else {
		ctImag = nil
	}

	// Scalar CKKS constants are encoded by Lattigo; the surrounding add/sub,
	// rotations, plaintext diagonal products and rescale are hardware modeled.
	if ctImag != nil {
		hwEval := newHardwareCKKSEvaluator(btp)
		if err = hwEval.Mul(ctImag, -1i, ctImag); err != nil {
			return nil, nil, err
		}
	}

	sumHW := EvalAdd(btp.Ref.BootstrappingParameters.RingQ().AtLevel(zV.Level()), rlweToHWCiphertext(btp.Ref.BootstrappingParameters.RingQ(), ctReal), rlweToHWCiphertext(btp.Ref.BootstrappingParameters.RingQ(), zV))
	for i := range sumHW.Value {
		ctReal.Value[i] = sumHW.Value[i]
	}

	return ctReal, ctImag, nil
}

func (btp *SlimBootstrapperHW) EvalMod(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	evm := btp.Ref.Mod1Parameters
	if ctIn.Level() < evm.LevelQ {
		return nil, fmt.Errorf("cannot hardware EvalMod: ct.Level()=%d < Mod1Parameters.LevelQ=%d", ctIn.Level(), evm.LevelQ)
	}

	ctWork := ctIn
	if ctWork.Level() > evm.LevelQ {
		ctWork = copyRLWECiphertextAtLevel(btp.Ref.BootstrappingParameters, ctWork, evm.LevelQ)
	}

	ctOut := ctWork.CopyNew()
	ctOut.Scale = evm.ScalingFactor()
	params := btp.Ref.BootstrappingParameters
	hwEval := newHardwareCKKSEvaluator(btp)

	Qi := params.Q()
	targetScale := ctOut.Scale
	for i := 0; i < evm.DoubleAngle; i++ {
		targetScale = targetScale.Mul(rlwe.NewScale(Qi[ctWork.Level()-evm.Mod1Poly.Depth()-evm.DoubleAngle+i+1]))
		targetScale.Value.Sqrt(&targetScale.Value)
	}

	if evm.Mod1Type == mod1.CosDiscrete || evm.Mod1Type == mod1.CosContinuous {
		offset := new(big.Float).Sub(&evm.Mod1Poly.B, &evm.Mod1Poly.A)
		offset.Mul(offset, new(big.Float).SetFloat64(evm.IntervalShrinkFactor()))
		offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
		if err := hwEval.Add(ctOut, offset, ctOut); err != nil {
			return nil, fmt.Errorf("hardware EvalMod offset: %w", err)
		}
	}

	scaling := complex(1, 0)
	sqrt2pi := complex(evm.Sqrt2Pi, 0)
	mod1Poly := evm.Mod1Poly
	if evm.Mod1InvPoly == nil {
		scaling = cmplx.Pow(scaling, complex(1/evm.IntervalShrinkFactor(), 0))
		mod1Poly = evm.Mod1Poly.Clone()
		scalingPowBig := bignum.NewComplex().SetComplex128(scaling)
		mul := bignum.NewComplexMultiplier().Mul
		for i := range mod1Poly.Coeffs {
			if mod1Poly.Coeffs[i] != nil {
				mul(mod1Poly.Coeffs[i], scalingPowBig, mod1Poly.Coeffs[i])
			}
		}
		sqrt2pi *= scaling
	}

	var err error
	polyEval := ckkspoly.Evaluator{
		Parameters: params,
		Evaluator:  commonPolynomialEvaluator(params, hwEval),
	}
	if ctOut, err = polyEval.Evaluate(ctOut, mod1Poly, targetScale); err != nil {
		return nil, fmt.Errorf("hardware EvalMod polynomial: %w", err)
	}

	for i := 0; i < evm.DoubleAngle; i++ {
		sqrt2pi *= sqrt2pi
		if err = hwEval.MulRelin(ctOut, ctOut, ctOut); err != nil {
			return nil, fmt.Errorf("hardware EvalMod double-angle mul: %w", err)
		}
		if err = hwEval.Add(ctOut, ctOut, ctOut); err != nil {
			return nil, fmt.Errorf("hardware EvalMod double-angle double: %w", err)
		}
		if err = hwEval.Add(ctOut, -sqrt2pi, ctOut); err != nil {
			return nil, fmt.Errorf("hardware EvalMod double-angle constant: %w", err)
		}
		if err = hwEval.Rescale(ctOut, ctOut); err != nil {
			return nil, fmt.Errorf("hardware EvalMod double-angle rescale: %w", err)
		}
	}

	if evm.Mod1InvPoly != nil {
		mod1InvPoly := evm.Mod1InvPoly.Clone()
		scalingBig := bignum.NewComplex().SetComplex128(scaling)
		mul := bignum.NewComplexMultiplier().Mul
		for i := range mod1InvPoly.Coeffs {
			if mod1InvPoly.Coeffs[i] != nil {
				mul(mod1InvPoly.Coeffs[i], scalingBig, mod1InvPoly.Coeffs[i])
			}
		}
		if ctOut, err = polyEval.Evaluate(ctOut, mod1InvPoly, ctOut.Scale); err != nil {
			return nil, fmt.Errorf("hardware EvalMod inverse polynomial: %w", err)
		}
	}

	ctOut.Scale = btp.Ref.BootstrappingParameters.DefaultScale()
	ctOut.Scale = btp.Ref.BootstrappingParameters.DefaultScale()
	return ctOut, nil
}

func (btp *SlimBootstrapperHW) Conjugate(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	galEl := btp.Ref.BootstrappingParameters.GaloisElementForComplexConjugation()
	gk, err := btp.Ref.MemEvaluationKeySet.GetGaloisKey(galEl)
	if err != nil {
		return nil, err
	}

	ringQ := btp.Ref.BootstrappingParameters.RingQ().AtLevel(ctIn.Level())
	ringP := btp.Ref.BootstrappingParameters.RingP().AtLevel(gk.LevelP())
	bconv := PrecomputeBasisConvConstants(ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1])

	ctHW := rlweToHWCiphertext(btp.Ref.BootstrappingParameters.RingQ(), ctIn)
	ks0Q, ks0P, ks1Q, ks1P := btp.KeySwitchQP(ringQ, ringP, ctHW.Value[1], &gk.EvaluationKey)
	ct0TimesP := ring.NewPoly(ringQ.N(), ringQ.Level())
	mulPolyByBigIntUnit(ringQ, ctHW.Value[0], ringP.ModulusAtLevel[ringP.Level()], ct0TimesP)
	addPolyUnit(ringQ, ks0Q, ct0TimesP, ks0Q)

	tmp0Q := ring.NewPoly(ringQ.N(), ringQ.Level())
	tmp1Q := ring.NewPoly(ringQ.N(), ringQ.Level())
	tmp0P := ring.NewPoly(ringP.N(), ringP.Level())
	tmp1P := ring.NewPoly(ringP.N(), ringP.Level())
	AutomorphismGaloisUnit(ringQ, ks0Q, galEl, tmp0Q)
	AutomorphismGaloisUnit(ringQ, ks1Q, galEl, tmp1Q)
	AutomorphismGaloisUnit(ringP, ks0P, galEl, tmp0P)
	AutomorphismGaloisUnit(ringP, ks1P, galEl, tmp1P)

	outHW := Ciphertext{Value: []ring.Poly{
		ModDownQPtoQNTTUnit(ringQ, ringP, ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1], bconv, tmp0Q, tmp0P, ringQ.N()),
		ModDownQPtoQNTTUnit(ringQ, ringP, ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1], bconv, tmp1Q, tmp1P, ringQ.N()),
	}}
	out := rlwe.NewCiphertext(btp.Ref.BootstrappingParameters, 1, ctIn.Level())
	*out.MetaData = *ctIn.MetaData
	for i := range outHW.Value {
		out.Value[i] = outHW.Value[i]
	}
	return out, nil
}

func (btp *SlimBootstrapperHW) dft(ctIn *rlwe.Ciphertext, mat dft.Matrix, name string) (*rlwe.Ciphertext, error) {
	opOut := ctIn.CopyNew()
	matrixIdx := 0

	for groupIdx, lvl := range mat.Levels {
		for range lvl {
			var in *rlwe.Ciphertext
			if matrixIdx == 0 {
				in = ctIn
			} else {
				in = opOut
			}

			btp.trace("  %s matrix %d: level=%d scale=2^%.2f", name, matrixIdx, in.Level(), math.Log2(in.Scale.Float64()))
			var err error
			opOut, err = btp.LinearTransform(in, mat.Matrices[matrixIdx])
			if err != nil {
				return nil, err
			}
			matrixIdx++
		}

		btp.trace("  %s rescale group %d", name, groupIdx)
		var err error
		opOut, err = btp.RescaleCiphertext(opOut)
		if err != nil {
			return nil, err
		}
	}

	opOut.LogDimensions = ctIn.LogDimensions
	return opOut, nil
}

func (btp *SlimBootstrapperHW) LinearTransform(ctIn *rlwe.Ciphertext, matrix lt.LinearTransformation) (*rlwe.Ciphertext, error) {
	if matrix.N1 != 0 {
		return nil, fmt.Errorf("hardware linear transform currently expects LogBSGSRatio=-1 (N1=0), got N1=%d", matrix.N1)
	}

	levelQ := ctIn.Level()
	if matrix.LevelQ < levelQ {
		levelQ = matrix.LevelQ
	}

	params := btp.Ref.BootstrappingParameters
	ringQ := params.RingQ().AtLevel(levelQ)
	ringP := params.RingP().AtLevel(matrix.LevelP)
	bconv := PrecomputeBasisConvConstants(ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1])

	ctHW := rlweToHWCiphertext(params.RingQ(), ctIn)
	acc0Q := ring.NewPoly(ringQ.N(), ringQ.Level())
	acc1Q := ring.NewPoly(ringQ.N(), ringQ.Level())
	acc0P := ring.NewPoly(ringP.N(), ringP.Level())
	acc1P := ring.NewPoly(ringP.N(), ringP.Level())
	hasNonZeroRotation := false

	keys := make([]int, 0, len(matrix.Vec))
	for k := range matrix.Vec {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	slots := 1 << matrix.LogDimensions.Cols
	for _, k := range keys {
		rot := k & (slots - 1)
		if rot == 0 {
			continue
		}

		galEl := params.GaloisElement(rot)
		gk, err := btp.Ref.MemEvaluationKeySet.GetGaloisKey(galEl)
		if err != nil {
			return nil, err
		}
		ks0Q, ks0P, ks1Q, ks1P := btp.KeySwitchQP(ringQ, ringP, ctHW.Value[1], &gk.EvaluationKey)

		ct0TimesP := ring.NewPoly(ringQ.N(), ringQ.Level())
		mulPolyByBigIntUnit(ringQ, ctHW.Value[0], ringP.ModulusAtLevel[ringP.Level()], ct0TimesP)
		addPolyUnit(ringQ, ks0Q, ct0TimesP, ks0Q)

		tmp0Q := ring.NewPoly(ringQ.N(), ringQ.Level())
		tmp1Q := ring.NewPoly(ringQ.N(), ringQ.Level())
		tmp0P := ring.NewPoly(ringP.N(), ringP.Level())
		tmp1P := ring.NewPoly(ringP.N(), ringP.Level())

		AutomorphismGaloisUnit(ringQ, ks0Q, galEl, tmp0Q)
		AutomorphismGaloisUnit(ringQ, ks1Q, galEl, tmp1Q)
		AutomorphismGaloisUnit(ringP, ks0P, galEl, tmp0P)
		AutomorphismGaloisUnit(ringP, ks1P, galEl, tmp1P)

		mulQPAdd(ringQ, ringP, matrix.Vec[k], tmp0Q, tmp0P, acc0Q, acc0P)
		mulQPAdd(ringQ, ringP, matrix.Vec[k], tmp1Q, tmp1P, acc1Q, acc1P)
		hasNonZeroRotation = true
	}

	if hasNonZeroRotation {
		acc0Q = ModDownQPtoQNTTUnit(ringQ, ringP, ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1], bconv, acc0Q, acc0P, ringQ.N())
		acc1Q = ModDownQPtoQNTTUnit(ringQ, ringP, ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1], bconv, acc1Q, acc1P, ringQ.N())
	}

	if pt, ok := matrix.Vec[0]; ok {
		zeroProd := EvalMulPlaintext(ringQ, ctHW, pt)
		sum := EvalAdd(ringQ, Ciphertext{Value: []ring.Poly{acc0Q, acc1Q}}, zeroProd)
		acc0Q = sum.Value[0]
		acc1Q = sum.Value[1]
	}

	out := rlwe.NewCiphertext(params, 1, levelQ)
	*out.MetaData = *ctIn.MetaData
	out.Scale = ctIn.Scale.Mul(matrix.Scale)
	out.Value[0] = acc0Q
	out.Value[1] = acc1Q
	return out, nil
}

func (btp *SlimBootstrapperHW) KeySwitchQP(
	ringQ, ringP *ring.Ring,
	c2 ring.Poly,
	evk *rlwe.EvaluationKey,
) (acc0Q, acc0P, acc1Q, acc1P ring.Poly) {
	N := ringQ.N()
	nQ := ringQ.Level() + 1
	nP := ringP.Level() + 1
	evkQ, evkP := ExtractEvaluationKey(evk)
	nDigits := (nQ + nP - 1) / nP
	if nDigits > len(evkQ) {
		nDigits = len(evkQ)
	}

	acc0Q = newPoly(ringQ, nQ)
	acc1Q = newPoly(ringQ, nQ)
	acc0P = ring.NewPoly(N, nP-1)
	acc1P = ring.NewPoly(N, nP-1)

	c2Coeff := newPoly(ringQ, nQ)
	NTTUnit(ringQ, c2, 0, c2Coeff)

	for d := 0; d < nDigits; d++ {
		digitQ := newPoly(ringQ, nQ)
		digitP := ring.NewPoly(N, nP-1)
		btp.Ref.Evaluator.DecomposeSingleNTT(ringQ.Level(), ringP.Level(), ringP.Level()+1, d, c2, c2Coeff, digitQ, digitP)

		for limb, q := range ringQ.ModuliChain()[:nQ] {
			for j := 0; j < N; j++ {
				ewuOut := ElementWiseUnit(OpAccQ,
					acc0Q.Coeffs[limb][j],
					acc1Q.Coeffs[limb][j],
					evkQ[d][0].Coeffs[limb][j],
					digitQ.Coeffs[limb][j],
					0,
					evkQ[d][1].Coeffs[limb][j],
					1, q)
				acc0Q.Coeffs[limb][j] = ewuOut.Out0
				acc1Q.Coeffs[limb][j] = ewuOut.Out1
			}
		}

		for limb, p := range ringP.ModuliChain()[:nP] {
			for j := 0; j < N; j++ {
				ewuOut := ElementWiseUnit(OpAccP,
					acc0P.Coeffs[limb][j],
					acc1P.Coeffs[limb][j],
					evkP[d][0].Coeffs[limb][j],
					0,
					digitP.Coeffs[limb][j],
					evkP[d][1].Coeffs[limb][j],
					1, p)
				acc0P.Coeffs[limb][j] = ewuOut.Out0
				acc1P.Coeffs[limb][j] = ewuOut.Out1
			}
		}
	}

	return
}

func mulQPAdd(ringQ, ringP *ring.Ring, pt ringqp.Poly, inQ, inP, accQ, accP ring.Poly) {
	Q := ringQ.ModuliChain()[:ringQ.Level()+1]
	P := ringP.ModuliChain()[:ringP.Level()+1]

	for limb, qi := range Q {
		for i := 0; i < ringQ.N(); i++ {
			ewuOut := ElementWiseUnit(OpAccP, accQ.Coeffs[limb][i], 0, inQ.Coeffs[limb][i], 0, pt.Q.Coeffs[limb][i], 0, 1, qi)
			accQ.Coeffs[limb][i] = ewuOut.Out0
		}
	}

	for limb, pi := range P {
		for i := 0; i < ringP.N(); i++ {
			ewuOut := ElementWiseUnit(OpAccP, accP.Coeffs[limb][i], 0, inP.Coeffs[limb][i], 0, pt.P.Coeffs[limb][i], 0, 1, pi)
			accP.Coeffs[limb][i] = ewuOut.Out0
		}
	}
}

func (btp *SlimBootstrapperHW) RescaleCiphertext(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if ctIn.Level() == 0 {
		return nil, fmt.Errorf("cannot hardware rescale level-0 ciphertext")
	}
	params := btp.Ref.BootstrappingParameters
	ringQ := params.RingQ().AtLevel(ctIn.Level())
	hw := Rescale(ringQ, rlweToHWCiphertext(params.RingQ(), ctIn))
	out := rlwe.NewCiphertext(params, ctIn.Degree(), ctIn.Level()-1)
	*out.MetaData = *ctIn.MetaData
	for i := range hw.Value {
		out.Value[i] = hw.Value[i]
	}
	out.Scale = ctIn.Scale.Div(rlwe.NewScale(ringQ.ModuliChain()[ctIn.Level()]))
	return out, nil
}

func (btp *SlimBootstrapperHW) trace(format string, args ...any) {
	if btp != nil && btp.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// ScaleDown mirrors bootstrapping.Evaluator.ScaleDown, using ElementWiseUnit
// for the integer scalar multiply and the existing Rescale datapath when levels
// must be consumed.
func (btp *SlimBootstrapperHW) ScaleDown(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Scale, error) {
	params := btp.Ref.BootstrappingParameters
	ringQ := params.RingQ()

	ctOut := ctIn.CopyNew()

	currentMessageRatio := rlwe.NewScale(ringQ.ModulusAtLevel[ctOut.Level()])
	currentMessageRatio = currentMessageRatio.Div(ctOut.Scale)
	targetMessageRatio := rlwe.NewScale(btp.Ref.Mod1Parameters.MessageRatio())
	scaleUp := currentMessageRatio.Div(targetMessageRatio)

	if scaleUp.Cmp(rlwe.NewScale(0.5)) == -1 {
		return nil, nil, fmt.Errorf("initial Q/Scale = %f < 0.5*Q[0]/MessageRatio = %f", currentMessageRatio.Float64(), targetMessageRatio.Float64())
	}

	scaleUpBig := scaleUp.BigInt()
	if scaleUpBig.BitLen() > 63 {
		return nil, nil, fmt.Errorf("hardware scalar path expects <=63-bit scalar, got %d bits", scaleUpBig.BitLen())
	}

	scaleUpUint := scaleUpBig.Uint64()
	for i := range ctOut.Value {
		mulScalarPolyUnit(ringQ.AtLevel(ctOut.Level()), ctOut.Value[i], scaleUpUint)
	}
	ctOut.Scale = ctOut.Scale.Mul(rlwe.NewScale(scaleUpBig))

	targetScale := new(big.Float).SetPrec(256).SetInt(ringQ.ModulusAtLevel[0])
	targetScale.Quo(targetScale, new(big.Float).SetFloat64(btp.Ref.Mod1Parameters.MessageRatio()))

	rescaled := false
	for ctOut.Level() != 0 {
		hw := rlweToHWCiphertext(ringQ, ctOut)
		hw = Rescale(ringQ.AtLevel(ctOut.Level()), hw)

		rescaledCt := rlwe.NewCiphertext(params, ctOut.Degree(), ctOut.Level()-1)
		for i := range hw.Value {
			rescaledCt.Value[i] = hw.Value[i]
		}
		*rescaledCt.MetaData = *ctOut.MetaData
		ctOut = rescaledCt
		rescaled = true
	}

	if rescaled {
		ctOut.Scale = rlwe.NewScale(targetScale)
	}
	errScale := ctOut.Scale.Div(rlwe.NewScale(targetScale))
	return ctOut, &errScale, nil
}

// ModUp mirrors the non-sparse path of bootstrapping.Evaluator.ModUp. It uses
// NTTUnit for domain conversion and centered coefficient broadcast for q0 -> Q.
func (btp *SlimBootstrapperHW) ModUp(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	params := btp.Ref.BootstrappingParameters
	if btp.Ref.EvkDenseToSparse != nil || btp.Ref.EvkSparseToDense != nil {
		return nil, fmt.Errorf("sparse bootstrapping keys are not supported by the current hardware ModUp path; set EphemeralSecretWeight to 0")
	}
	if btp.Ref.CoeffsToSlotsParameters.LogSlots != params.LogMaxSlots() {
		return nil, fmt.Errorf("hardware ModUp trace is only implemented for full packing: logSlots=%d logMaxSlots=%d", btp.Ref.CoeffsToSlotsParameters.LogSlots, params.LogMaxSlots())
	}

	ctOut := ctIn.CopyNew()
	ringQ0 := params.RingQ().AtLevel(ctOut.Level())
	for i := range ctOut.Value {
		NTTUnit(ringQ0, ctOut.Value[i], 0, ctOut.Value[i])
	}

	ctOut.Resize(ctOut.Degree(), params.MaxLevel())
	ringQ := params.RingQ().AtLevel(params.MaxLevel())
	Q := ringQ.ModuliChain()
	q0 := Q[0]
	halfQ0 := q0 >> 1
	N := ringQ.N()

	for polyIdx := range ctOut.Value {
		for coeffIdx := 0; coeffIdx < N; coeffIdx++ {
			coeff := ctOut.Value[polyIdx].Coeffs[0][coeffIdx]
			neg := coeff >= halfQ0
			if neg {
				coeff = q0 - coeff
			}

			for limb := 1; limb < len(Q); limb++ {
				reduced := coeff % Q[limb]
				if neg && reduced != 0 {
					reduced = Q[limb] - reduced
				}
				ctOut.Value[polyIdx].Coeffs[limb][coeffIdx] = reduced
			}
		}
	}

	for i := range ctOut.Value {
		NTTUnit(ringQ, ctOut.Value[i], 1, ctOut.Value[i])
	}

	if scale := (btp.Ref.Mod1Parameters.ScalingFactor().Float64() / btp.Ref.Mod1Parameters.MessageRatio()) / ctOut.Scale.Float64(); scale > 1 {
		scalar := uint64(math.Round(scale))
		for i := range ctOut.Value {
			mulScalarPolyUnit(ringQ, ctOut.Value[i], scalar)
		}
		ctOut.Scale = ctOut.Scale.Mul(rlwe.NewScale(scale))
	}

	return ctOut, nil
}

func rlweToHWCiphertext(ringQ *ring.Ring, ct *rlwe.Ciphertext) Ciphertext {
	level := ct.Level()
	out := Ciphertext{Value: make([]ring.Poly, len(ct.Value))}
	for i := range ct.Value {
		out.Value[i] = deepCopyPoly(ringQ.AtLevel(level), level+1, ct.Value[i])
	}
	return out
}

func mulScalarPolyUnit(r *ring.Ring, pol ring.Poly, scalar uint64) {
	Q := r.ModuliChain()[:r.Level()+1]
	scalarByLimb := make([]uint64, len(Q))
	for i, qi := range Q {
		scalarByLimb[i] = scalar % qi
	}

	for limb, qi := range Q {
		for coeffIdx := 0; coeffIdx < r.N(); coeffIdx++ {
			out := ElementWiseUnit(OpMAD,
				0,
				0,
				pol.Coeffs[limb][coeffIdx],
				0,
				scalarByLimb[limb],
				0,
				0,
				qi)
			pol.Coeffs[limb][coeffIdx] = out.Out0
		}
	}
}

func mulPolyByBigIntUnit(r *ring.Ring, in ring.Poly, scalar *big.Int, out ring.Poly) {
	Q := r.ModuliChain()[:r.Level()+1]
	scalarQi := new(big.Int)
	for limb, qi := range Q {
		scalarQi.Mod(scalar, new(big.Int).SetUint64(qi))
		for coeffIdx := 0; coeffIdx < r.N(); coeffIdx++ {
			ewuOut := ElementWiseUnit(OpMAD,
				0,
				0,
				in.Coeffs[limb][coeffIdx],
				0,
				scalarQi.Uint64(),
				0,
				0,
				qi)
			out.Coeffs[limb][coeffIdx] = ewuOut.Out0
		}
	}
}

func addPolyUnit(r *ring.Ring, lhs, rhs, out ring.Poly) {
	Q := r.ModuliChain()[:r.Level()+1]
	for limb, qi := range Q {
		for coeffIdx := 0; coeffIdx < r.N(); coeffIdx++ {
			out.Coeffs[limb][coeffIdx] = ring.CRed(lhs.Coeffs[limb][coeffIdx]+rhs.Coeffs[limb][coeffIdx], qi)
		}
	}
}
