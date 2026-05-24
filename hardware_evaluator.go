package main

import (
	"fmt"
	"math/big"
	"math/bits"

	commonpoly "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/ring/ringqp"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

type hardwareCKKSEvaluator struct {
	btp    *SlimBootstrapperHW
	params ckks.Parameters
	*rlwe.Evaluator
}

type hardwareCoefficientGetter struct {
	values []*bignum.Complex
}

func newHardwareCKKSEvaluator(btp *SlimBootstrapperHW) hardwareCKKSEvaluator {
	params := btp.Ref.BootstrappingParameters
	return hardwareCKKSEvaluator{
		btp:       btp,
		params:    params,
		Evaluator: btp.Ref.Evaluator.Evaluator,
	}
}

func commonPolynomialEvaluator(params ckks.Parameters, eval hardwareCKKSEvaluator) commonpoly.Evaluator[*bignum.Complex] {
	return commonpoly.Evaluator[*bignum.Complex]{
		Evaluator:         eval,
		CoefficientGetter: hardwareCoefficientGetter{values: make([]*bignum.Complex, params.MaxSlots())},
	}
}

func (h hardwareCKKSEvaluator) GetParameters() *ckks.Parameters {
	return &h.params
}

func (h hardwareCKKSEvaluator) GetRLWEParameters() *rlwe.Parameters {
	return &h.params.Parameters
}

func (h hardwareCoefficientGetter) GetVectorCoefficient(pol commonpoly.PolynomialVector, k int) (values []*bignum.Complex) {
	values = h.values
	if len(values) == 0 {
		values = make([]*bignum.Complex, 1)
	}
	for i := range values {
		values[i] = nil
	}
	for i, p := range pol.Value {
		for _, j := range pol.Mapping[i] {
			if j >= len(values) {
				next := make([]*bignum.Complex, j+1)
				copy(next, values)
				values = next
			}
			values[j] = p.Coeffs[k]
		}
	}
	return values
}

func (h hardwareCoefficientGetter) GetSingleCoefficient(pol commonpoly.Polynomial, k int) (value *bignum.Complex) {
	return pol.Coeffs[k]
}

func (h hardwareCKKSEvaluator) Add(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) error {
	switch op1 := op1.(type) {
	case rlwe.ElementInterface[ring.Poly]:
		return h.addElement(op0, op1.El(), opOut, false)
	case complex128, float64, int, int64, uint, uint64, *big.Int, *big.Float, *bignum.Complex:
		return h.addScalar(op0, bignum.ToComplex(op1, h.params.EncodingPrecision()), opOut, false)
	case []complex128, []float64, []*big.Float, []*bignum.Complex:
		pt, err := h.encodeOperand(op0, op1, op0.Scale)
		if err != nil {
			return fmt.Errorf("hardware Add encode: %w", err)
		}
		return h.addElement(op0, pt.El(), opOut, false)
	default:
		return fmt.Errorf("hardware Add: unsupported operand %T", op1)
	}
}

func (h hardwareCKKSEvaluator) AddNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (*rlwe.Ciphertext, error) {
	opOut := ckks.NewCiphertext(h.params, op0.Degree(), op0.Level())
	return opOut, h.Add(op0, op1, opOut)
}

func (h hardwareCKKSEvaluator) Sub(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) error {
	switch op1 := op1.(type) {
	case rlwe.ElementInterface[ring.Poly]:
		return h.addElement(op0, op1.El(), opOut, true)
	case complex128, float64, int, int64, uint, uint64, *big.Int, *big.Float, *bignum.Complex:
		return h.addScalar(op0, bignum.ToComplex(op1, h.params.EncodingPrecision()), opOut, true)
	case []complex128, []float64, []*big.Float, []*bignum.Complex:
		pt, err := h.encodeOperand(op0, op1, op0.Scale)
		if err != nil {
			return fmt.Errorf("hardware Sub encode: %w", err)
		}
		return h.addElement(op0, pt.El(), opOut, true)
	default:
		return fmt.Errorf("hardware Sub: unsupported operand %T", op1)
	}
}

func (h hardwareCKKSEvaluator) SubNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (*rlwe.Ciphertext, error) {
	opOut := ckks.NewCiphertext(h.params, op0.Degree(), op0.Level())
	return opOut, h.Sub(op0, op1, opOut)
}

func (h hardwareCKKSEvaluator) Mul(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) error {
	switch op1 := op1.(type) {
	case rlwe.ElementInterface[ring.Poly]:
		if op1.El().Degree() == 0 {
			return h.mulPlain(op0, op1.El(), opOut)
		}
		return h.mulCiphertexts(op0, op1.El(), opOut, false)
	case complex128, float64, int, int64, uint, uint64, *big.Int, *big.Float, *bignum.Complex:
		return h.mulScalar(op0, bignum.ToComplex(op1, h.params.EncodingPrecision()), opOut)
	case []complex128, []float64, []*big.Float, []*bignum.Complex:
		level := op0.Level()
		scale := rlwe.NewScale(h.params.RingQ().AtLevel(level).ModuliChain()[level])
		for i := 1; i < h.params.LevelsConsumedPerRescaling(); i++ {
			scale = scale.Mul(rlwe.NewScale(h.params.RingQ().AtLevel(level).ModuliChain()[level-i]))
		}
		pt, err := h.encodeOperand(op0, op1, scale)
		if err != nil {
			return fmt.Errorf("hardware Mul encode: %w", err)
		}
		return h.mulPlain(op0, pt.El(), opOut)
	default:
		return fmt.Errorf("hardware Mul: unsupported operand %T", op1)
	}
}

func (h hardwareCKKSEvaluator) MulNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (*rlwe.Ciphertext, error) {
	degree := op0.Degree()
	if op1El, ok := op1.(rlwe.ElementInterface[ring.Poly]); ok {
		if op1El.El().Degree() > 0 {
			degree = op0.Degree() + op1El.El().Degree()
		}
	}
	if degree > 2 {
		degree = 2
	}
	opOut := ckks.NewCiphertext(h.params, degree, op0.Level())
	return opOut, h.Mul(op0, op1, opOut)
}

func (h hardwareCKKSEvaluator) MulRelin(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) error {
	switch op1 := op1.(type) {
	case rlwe.ElementInterface[ring.Poly]:
		if op1.El().Degree() == 0 {
			return h.mulPlain(op0, op1.El(), opOut)
		}
		return h.mulCiphertexts(op0, op1.El(), opOut, true)
	default:
		return h.Mul(op0, op1, opOut)
	}
}

func (h hardwareCKKSEvaluator) MulRelinNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (*rlwe.Ciphertext, error) {
	opOut := ckks.NewCiphertext(h.params, 1, op0.Level())
	return opOut, h.MulRelin(op0, op1, opOut)
}

func (h hardwareCKKSEvaluator) MulThenAdd(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) error {
	if cmplx := operandAsComplex(op1, h.params.EncodingPrecision()); cmplx != nil {
		opOut.Resize(op0.Degree(), opOut.Level())
		var scaleRLWE rlwe.Scale
		switch op0.Scale.Cmp(opOut.Scale) {
		case 0:
			scaleRLWE = rlwe.NewScale(1)
			if !cmplx.IsInt() {
				level := minInt(op0.Level(), opOut.Level())
				scaleRLWE = rlwe.NewScale(h.params.RingQ().AtLevel(level).ModuliChain()[level])
				scaleInt := scaleRLWE.BigInt()
				if err := h.mulScalar(opOut, bignum.ToComplex(scaleInt, h.params.EncodingPrecision()), opOut); err != nil {
					return err
				}
				opOut.Scale = opOut.Scale.Mul(scaleRLWE)
			}
		case -1:
			scaleRLWE = opOut.Scale.Div(op0.Scale)
		default:
			return fmt.Errorf("hardware MulThenAdd: op0.Scale > opOut.Scale is not supported")
		}

		tmp := ckks.NewCiphertext(h.params, op0.Degree(), op0.Level())
		if err := h.mulScalarWithScale(op0, cmplx, scaleRLWE, tmp); err != nil {
			return err
		}
		tmp.Scale = opOut.Scale
		return h.Add(opOut, tmp, opOut)
	}

	if op0.Scale.Cmp(opOut.Scale) == 0 {
		if cmplx := operandAsComplex(op1, h.params.EncodingPrecision()); cmplx != nil && !cmplx.IsInt() {
			level := minInt(op0.Level(), opOut.Level())
			scale := rlwe.NewScale(h.params.RingQ().AtLevel(level).ModuliChain()[level])
			scaleInt := scale.BigInt()
			if err := h.mulScalar(opOut, bignum.ToComplex(scaleInt, h.params.EncodingPrecision()), opOut); err != nil {
				return err
			}
			opOut.Scale = opOut.Scale.Mul(scale)
		}
	}

	tmp, err := h.MulNew(op0, op1)
	if err != nil {
		return err
	}
	return h.Add(opOut, tmp, opOut)
}

func (h hardwareCKKSEvaluator) Relinearize(op0, opOut *rlwe.Ciphertext) error {
	if op0.Degree() <= 1 {
		copyCiphertextInto(opOut, op0)
		return nil
	}
	rlk, err := h.btp.Ref.MemEvaluationKeySet.GetRelinearizationKey()
	if err != nil {
		return err
	}
	level := op0.Level()
	ringQ := h.params.RingQ().AtLevel(level)
	ringP := h.params.RingP().AtLevel(rlk.LevelP())
	bconv := PrecomputeBasisConvConstants(ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1])
	ctHW := rlweToHWCiphertextAtLevel(h.params.RingQ(), op0, level)
	ks0Q, ks0P, ks1Q, ks1P := h.btp.KeySwitchQP(ringQ, ringP, ctHW.Value[2], &rlk.EvaluationKey)
	ks0 := ModDownQPtoQNTTUnit(ringQ, ringP, ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1], bconv, ks0Q, ks0P, ringQ.N())
	ks1 := ModDownQPtoQNTTUnit(ringQ, ringP, ringQ.ModuliChain()[:ringQ.Level()+1], ringP.ModuliChain()[:ringP.Level()+1], bconv, ks1Q, ks1P, ringQ.N())
	hw := Ciphertext{Value: []ring.Poly{newPoly(ringQ, level+1), newPoly(ringQ, level+1)}}
	for limb, q := range ringQ.ModuliChain()[:level+1] {
		for i := 0; i < ringQ.N(); i++ {
			ewuOut := ElementWiseUnit(OpMAD,
				ctHW.Value[0].Coeffs[limb][i],
				ctHW.Value[1].Coeffs[limb][i],
				ks0.Coeffs[limb][i],
				ks1.Coeffs[limb][i],
				1, 1, 1, q)
			hw.Value[0].Coeffs[limb][i] = ewuOut.Out0
			hw.Value[1].Coeffs[limb][i] = ewuOut.Out1
		}
	}
	out := hwToRLWECiphertext(h.params, hw, 1, level, op0.MetaData)
	out.Scale = op0.Scale
	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) Rescale(op0, opOut *rlwe.Ciphertext) error {
	out, err := h.btp.RescaleCiphertext(op0)
	if err != nil {
		return err
	}
	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) addElement(op0 *rlwe.Ciphertext, op1 *rlwe.Element[ring.Poly], opOut *rlwe.Ciphertext, sub bool) error {
	level := minInt(op0.Level(), op1.Level())
	degree := maxInt(op0.Degree(), op1.Degree())
	op0Work := copyRLWECiphertextAtLevel(h.params, op0, level)
	op1Work := copyElementAtLevel(h.params.RingQ(), op1, level)

	outScale := op0.Scale.Max(op1.Scale)
	switch op0.Scale.Cmp(op1.Scale) {
	case 1:
		ratioScale := op0.Scale.Div(op1.Scale)
		ratio, _ := ratioScale.Value.Int(nil)
		if ratio.Sign() > 0 {
			tmp := ckks.NewCiphertext(h.params, op1Work.Degree(), level)
			if err := h.mulScalar(&rlwe.Ciphertext{Element: *op1Work}, bignum.ToComplex(ratio, h.params.EncodingPrecision()), tmp); err != nil {
				return err
			}
			op1Work = tmp.El()
		}
	case -1:
		ratioScale := op1.Scale.Div(op0.Scale)
		ratio, _ := ratioScale.Value.Int(nil)
		if ratio.Sign() > 0 {
			tmp := ckks.NewCiphertext(h.params, op0Work.Degree(), level)
			if err := h.mulScalar(op0Work, bignum.ToComplex(ratio, h.params.EncodingPrecision()), tmp); err != nil {
				return err
			}
			op0Work = tmp
		}
	}

	out := ckks.NewCiphertext(h.params, degree, level)
	*out.MetaData = *op0.MetaData
	out.Scale = outScale
	ringQ := h.params.RingQ().AtLevel(level)
	Q := ringQ.ModuliChain()[:level+1]
	N := ringQ.N()

	for d := 0; d <= degree; d++ {
		for limb, qi := range Q {
			for i := 0; i < N; i++ {
				var a, b uint64
				if d <= op0Work.Degree() {
					a = op0Work.Value[d].Coeffs[limb][i]
				}
				if d <= op1Work.Degree() {
					b = op1Work.Value[d].Coeffs[limb][i]
				}
				if sub {
					out.Value[d].Coeffs[limb][i] = ring.CRed(a+qi-b, qi)
				} else {
					out.Value[d].Coeffs[limb][i] = ring.CRed(a+b, qi)
				}
			}
		}
	}

	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) addScalar(op0 *rlwe.Ciphertext, scalar *bignum.Complex, opOut *rlwe.Ciphertext, sub bool) error {
	level := op0.Level()
	out := copyRLWECiphertextAtLevel(h.params, op0, level)
	ringQ := h.params.RingQ().AtLevel(level)
	real, imag := bigComplexToRNSScalarHW(ringQ, &op0.Scale.Value, scalar)
	prepareDoubleRNSScalar(ringQ, real, imag)
	if sub {
		negateRNSScalar(ringQ, real)
		negateRNSScalar(ringQ, imag)
	}
	addDoubleRNSScalarUnit(ringQ, out.Value[0], real, imag)
	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) mulScalar(op0 *rlwe.Ciphertext, scalar *bignum.Complex, opOut *rlwe.Ciphertext) error {
	level := op0.Level()
	ringQ := h.params.RingQ().AtLevel(level)
	var scale rlwe.Scale
	if scalar.IsInt() {
		scale = rlwe.NewScale(1)
	} else {
		scale = rlwe.NewScale(ringQ.ModuliChain()[level])
		for i := 1; i < h.params.LevelsConsumedPerRescaling(); i++ {
			scale = scale.Mul(rlwe.NewScale(ringQ.ModuliChain()[level-i]))
		}
	}
	return h.mulScalarWithScale(op0, scalar, scale, opOut)
}

func (h hardwareCKKSEvaluator) mulScalarWithScale(op0 *rlwe.Ciphertext, scalar *bignum.Complex, scale rlwe.Scale, opOut *rlwe.Ciphertext) error {
	level := op0.Level()
	ringQ := h.params.RingQ().AtLevel(level)
	out := ckks.NewCiphertext(h.params, op0.Degree(), level)
	*out.MetaData = *op0.MetaData
	out.Scale = op0.Scale.Mul(scale)
	real, imag := bigComplexToRNSScalarHW(ringQ, &scale.Value, scalar)
	prepareDoubleRNSScalar(ringQ, real, imag)
	for d := range op0.Value {
		mulDoubleRNSScalarUnit(ringQ, op0.Value[d], real, imag, out.Value[d])
	}
	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) mulPlain(op0 *rlwe.Ciphertext, op1 *rlwe.Element[ring.Poly], opOut *rlwe.Ciphertext) error {
	level := minInt(op0.Level(), op1.Level())
	ringQ := h.params.RingQ().AtLevel(level)
	hw := EvalMulPlaintext(ringQ, rlweToHWCiphertextAtLevel(h.params.RingQ(), op0, level), ringqpFromPlain(op1, level))
	out := hwToRLWECiphertext(h.params, hw, op0.Degree(), level, op0.MetaData)
	out.Scale = op0.Scale.Mul(op1.Scale)
	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) mulCiphertexts(op0 *rlwe.Ciphertext, op1 *rlwe.Element[ring.Poly], opOut *rlwe.Ciphertext, relin bool) error {
	if op0.Degree() != 1 || op1.Degree() != 1 {
		return fmt.Errorf("hardware Mul: ciphertext-ciphertext multiply expects degree-1 inputs, got %d and %d", op0.Degree(), op1.Degree())
	}
	level := minInt(op0.Level(), op1.Level())
	ringQ := h.params.RingQ().AtLevel(level)
	ct0 := rlweToHWCiphertextAtLevel(h.params.RingQ(), op0, level)
	ct1 := elementToHWCiphertextAtLevel(h.params.RingQ(), op1, level)

	var hw Ciphertext
	degree := 2
	if relin {
		tensor := EvalTensor(ringQ, ct0, ct1)
		tmp := hwToRLWECiphertext(h.params, tensor, 2, level, op0.MetaData)
		tmp.Scale = op0.Scale.Mul(op1.Scale)
		return h.Relinearize(tmp, opOut)
	} else {
		hw = EvalTensor(ringQ, ct0, ct1)
	}

	out := hwToRLWECiphertext(h.params, hw, degree, level, op0.MetaData)
	out.Scale = op0.Scale.Mul(op1.Scale)
	copyCiphertextInto(opOut, out)
	return nil
}

func (h hardwareCKKSEvaluator) encodeOperand(op0 *rlwe.Ciphertext, values interface{}, scale rlwe.Scale) (*rlwe.Plaintext, error) {
	pt := ckks.NewPlaintext(h.params, op0.Level())
	*pt.MetaData = *op0.MetaData
	pt.Scale = scale
	if err := h.btp.Ref.Evaluator.Encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return pt, nil
}

func copyCiphertextInto(dst, src *rlwe.Ciphertext) {
	dst.Resize(src.Degree(), src.Level())
	*dst.MetaData = *src.MetaData
	for i := range src.Value {
		dst.Value[i] = src.Value[i]
	}
}

func copyRLWECiphertextAtLevel(params ckks.Parameters, ct *rlwe.Ciphertext, level int) *rlwe.Ciphertext {
	out := ckks.NewCiphertext(params, ct.Degree(), level)
	*out.MetaData = *ct.MetaData
	for d := range out.Value {
		for limb := 0; limb <= level; limb++ {
			copy(out.Value[d].Coeffs[limb], ct.Value[d].Coeffs[limb])
		}
	}
	return out
}

func copyElementAtLevel(ringQ *ring.Ring, el *rlwe.Element[ring.Poly], level int) *rlwe.Element[ring.Poly] {
	out := &rlwe.Element[ring.Poly]{
		Value:    make([]ring.Poly, el.Degree()+1),
		MetaData: el.MetaData.CopyNew(),
	}
	ringAtLevel := ringQ.AtLevel(level)
	for d := range out.Value {
		out.Value[d] = deepCopyPoly(ringAtLevel, level+1, el.Value[d])
	}
	return out
}

func hwToRLWECiphertext(params ckks.Parameters, hw Ciphertext, degree, level int, meta *rlwe.MetaData) *rlwe.Ciphertext {
	out := ckks.NewCiphertext(params, degree, level)
	if meta != nil {
		*out.MetaData = *meta
	}
	for i := range hw.Value {
		out.Value[i] = hw.Value[i]
	}
	return out
}

func rlweToHWCiphertextAtLevel(ringQ *ring.Ring, ct *rlwe.Ciphertext, level int) Ciphertext {
	out := Ciphertext{Value: make([]ring.Poly, len(ct.Value))}
	ringAtLevel := ringQ.AtLevel(level)
	for i := range ct.Value {
		out.Value[i] = deepCopyPoly(ringAtLevel, level+1, ct.Value[i])
	}
	return out
}

func elementToHWCiphertextAtLevel(ringQ *ring.Ring, el *rlwe.Element[ring.Poly], level int) Ciphertext {
	out := Ciphertext{Value: make([]ring.Poly, len(el.Value))}
	ringAtLevel := ringQ.AtLevel(level)
	for i := range el.Value {
		out.Value[i] = deepCopyPoly(ringAtLevel, level+1, el.Value[i])
	}
	return out
}

func ringqpFromPlain(el *rlwe.Element[ring.Poly], level int) ringqp.Poly {
	return ringqp.Poly{Q: copyPolyToLevel(el.Value[0], level)}
}

func copyPolyToLevel(src ring.Poly, level int) ring.Poly {
	dst := ring.NewPoly(len(src.Coeffs[0]), level)
	for limb := 0; limb <= level; limb++ {
		copy(dst.Coeffs[limb], src.Coeffs[limb])
	}
	return dst
}

func operandAsComplex(op rlwe.Operand, precision uint) *bignum.Complex {
	switch op := op.(type) {
	case complex128, float64, int, int64, uint, uint64, *big.Int, *big.Float, *bignum.Complex:
		return bignum.ToComplex(op, precision)
	default:
		return nil
	}
}

func bigComplexToRNSScalarHW(r *ring.Ring, scale *big.Float, cmplx *bignum.Complex) (real, imag ring.RNSScalar) {
	if scale == nil {
		scale = new(big.Float).SetFloat64(1)
	}

	realInt := new(big.Int)
	if cmplx[0] != nil {
		x := new(big.Float).Mul(cmplx[0], scale)
		if cmplx[0].Sign() > 0 {
			x.Add(x, new(big.Float).SetFloat64(0.5))
		} else if cmplx[0].Sign() < 0 {
			x.Sub(x, new(big.Float).SetFloat64(0.5))
		}
		x.Int(realInt)
	}

	imagInt := new(big.Int)
	if cmplx[1] != nil {
		x := new(big.Float).Mul(cmplx[1], scale)
		if cmplx[1].Sign() > 0 {
			x.Add(x, new(big.Float).SetFloat64(0.5))
		} else if cmplx[1].Sign() < 0 {
			x.Sub(x, new(big.Float).SetFloat64(0.5))
		}
		x.Int(imagInt)
	}

	return r.NewRNSScalarFromBigint(realInt), r.NewRNSScalarFromBigint(imagInt)
}

func prepareDoubleRNSScalar(r *ring.Ring, real, imag ring.RNSScalar) {
	for limb, s := range r.SubRings[:r.Level()+1] {
		imagRoot := mRed(imag[limb], s.RootsForward[1], s.Modulus, mredParams(s.Modulus))
		real[limb], imag[limb] = ring.CRed(real[limb]+imagRoot, s.Modulus), ring.CRed(real[limb]+s.Modulus-imagRoot, s.Modulus)
	}
}

func negateRNSScalar(r *ring.Ring, scalar ring.RNSScalar) {
	for limb, s := range r.SubRings[:r.Level()+1] {
		if scalar[limb] != 0 {
			scalar[limb] = s.Modulus - scalar[limb]
		}
	}
}

func addDoubleRNSScalarUnit(r *ring.Ring, pol ring.Poly, scalar0, scalar1 ring.RNSScalar) {
	countDoubleRNSAddUnit()
	nHalf := r.N() >> 1
	for limb, qi := range r.ModuliChain()[:r.Level()+1] {
		for i := 0; i < r.N(); i++ {
			scalar := scalar0[limb]
			if i >= nHalf {
				scalar = scalar1[limb]
			}
			pol.Coeffs[limb][i] = ring.CRed(pol.Coeffs[limb][i]+scalar, qi)
		}
	}
}

func mulDoubleRNSScalarUnit(r *ring.Ring, in ring.Poly, scalar0, scalar1 ring.RNSScalar, out ring.Poly) {
	countDoubleRNSMulUnit()
	nHalf := r.N() >> 1
	for limb, qi := range r.ModuliChain()[:r.Level()+1] {
		for i := 0; i < r.N(); i++ {
			scalar := scalar0[limb]
			if i >= nHalf {
				scalar = scalar1[limb]
			}
			ewuOut := ElementWiseUnit(OpMAD, 0, 0, in.Coeffs[limb][i], 0, scalar, 0, 0, qi)
			out.Coeffs[limb][i] = ewuOut.Out0
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type hardwareSimEvaluator struct {
	params                     ckks.Parameters
	levelsConsumedPerRescaling int
}

func (s hardwareSimEvaluator) PolynomialDepth(degree int) int {
	if degree <= 0 {
		panic(fmt.Errorf("invalid polynomial degree %d", degree))
	}
	return s.levelsConsumedPerRescaling * (bits.Len64(uint64(degree)) - 1)
}

func (s hardwareSimEvaluator) Rescale(op0 *commonpoly.SimOperand) {
	for i := 0; i < s.levelsConsumedPerRescaling; i++ {
		op0.Scale = op0.Scale.Div(rlwe.NewScale(s.params.Q()[op0.Level]))
		op0.Level--
	}
}

func (s hardwareSimEvaluator) MulNew(op0, op1 *commonpoly.SimOperand) *commonpoly.SimOperand {
	return &commonpoly.SimOperand{
		Level: minInt(op0.Level, op1.Level),
		Scale: op0.Scale.Mul(op1.Scale),
	}
}

func (s hardwareSimEvaluator) UpdateLevelAndScaleBabyStep(lead bool, tLevelOld int, tScaleOld rlwe.Scale) (int, rlwe.Scale) {
	tLevelNew := tLevelOld
	tScaleNew := tScaleOld
	if lead {
		for i := 0; i < s.levelsConsumedPerRescaling; i++ {
			tScaleNew = tScaleNew.Mul(rlwe.NewScale(s.params.Q()[tLevelNew-i]))
		}
	}
	return tLevelNew, tScaleNew
}

func (s hardwareSimEvaluator) UpdateLevelAndScaleGiantStep(lead bool, tLevelOld int, tScaleOld, xPowScale rlwe.Scale) (int, rlwe.Scale) {
	Q := s.params.Q()
	var qi *big.Int
	if lead {
		qi = new(big.Int).SetUint64(Q[tLevelOld])
		for i := 1; i < s.levelsConsumedPerRescaling; i++ {
			qi.Mul(qi, new(big.Int).SetUint64(Q[tLevelOld-i]))
		}
	} else {
		qi = new(big.Int).SetUint64(Q[tLevelOld+s.levelsConsumedPerRescaling])
		for i := 1; i < s.levelsConsumedPerRescaling; i++ {
			qi.Mul(qi, new(big.Int).SetUint64(Q[tLevelOld+s.levelsConsumedPerRescaling-i]))
		}
	}
	tLevelNew := tLevelOld + s.levelsConsumedPerRescaling
	tScaleNew := tScaleOld.Mul(rlwe.NewScale(qi)).Div(xPowScale)
	return tLevelNew, tScaleNew
}
