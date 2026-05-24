package main

import (
	"math/big"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/ring/ringqp"
)

// Ciphertext represents a polynomial array.
type Ciphertext struct {
	Value []ring.Poly
}

func newPoly(r *ring.Ring, nLimbs int) ring.Poly {
	return ring.NewPoly(r.N(), nLimbs-1)
}

func newCiphertext(r *ring.Ring, degree int) Ciphertext {
	value := make([]ring.Poly, degree+1)
	for i := range value {
		value[i] = ring.NewPoly(r.N(), r.Level())
	}
	return Ciphertext{Value: value}
}

func copyCiphertext(r *ring.Ring, ct Ciphertext) Ciphertext {
	out := newCiphertext(r, len(ct.Value)-1)
	for i := range ct.Value {
		for limb := 0; limb <= r.Level(); limb++ {
			copy(out.Value[i].Coeffs[limb], ct.Value[i].Coeffs[limb])
		}
	}
	return out
}

func ExtractEvaluationKey(evk *rlwe.EvaluationKey) (evkQ, evkP [][2]ring.Poly) {
	gadget := evk.GadgetCiphertext
	nDigits := len(gadget.Value)
	evkQ = make([][2]ring.Poly, nDigits)
	evkP = make([][2]ring.Poly, nDigits)

	for d := 0; d < nDigits; d++ {
		digitVec := gadget.Value[d][0]
		evkQ[d][0] = digitVec[0].Q
		evkQ[d][1] = digitVec[1].Q
		evkP[d][0] = digitVec[0].P
		evkP[d][1] = digitVec[1].P
	}

	return
}

// -------------------------------------------------------------------------
// 0. Simple ciphertext operations
// -------------------------------------------------------------------------

func EvalAdd(ringQ *ring.Ring, ct0, ct1 Ciphertext) Ciphertext {
	out := newCiphertext(ringQ, len(ct0.Value)-1)
	N := ringQ.N()
	Q := ringQ.ModuliChain()[:ringQ.Level()+1]

	for degree := range out.Value {
		for limb, qi := range Q {
			for i := 0; i < N; i++ {
				out.Value[degree].Coeffs[limb][i] = ring.CRed(ct0.Value[degree].Coeffs[limb][i]+ct1.Value[degree].Coeffs[limb][i], qi)
			}
		}
	}
	return out
}

func EvalSub(ringQ *ring.Ring, ct0, ct1 Ciphertext) Ciphertext {
	out := newCiphertext(ringQ, len(ct0.Value)-1)
	N := ringQ.N()
	Q := ringQ.ModuliChain()[:ringQ.Level()+1]

	for degree := range out.Value {
		for limb, qi := range Q {
			for i := 0; i < N; i++ {
				out.Value[degree].Coeffs[limb][i] = ring.CRed(ct0.Value[degree].Coeffs[limb][i]+qi-ct1.Value[degree].Coeffs[limb][i], qi)
			}
		}
	}
	return out
}

// EvalMulPlaintext multiplies a degree-1 ciphertext by an encoded plaintext
// diagonal. The plaintext is expected in NTT/Montgomery form, as Lattigo stores
// linear-transformation diagonals.
func EvalMulPlaintext(ringQ *ring.Ring, ct Ciphertext, pt ringqp.Poly) Ciphertext {
	out := newCiphertext(ringQ, len(ct.Value)-1)
	N := ringQ.N()
	Q := ringQ.ModuliChain()[:ringQ.Level()+1]

	for degree := range ct.Value {
		for limb, qi := range Q {
			qInv := mredParams(qi)
			for i := 0; i < N; i++ {
				out.Value[degree].Coeffs[limb][i] = mRed(pt.Q.Coeffs[limb][i], ct.Value[degree].Coeffs[limb][i], qi, qInv)
			}
		}
	}
	return out
}

func EvalMulPlaintextAdd(ringQ *ring.Ring, ct Ciphertext, pt ringqp.Poly, acc Ciphertext) Ciphertext {
	prod := EvalMulPlaintext(ringQ, ct, pt)
	return EvalAdd(ringQ, acc, prod)
}

// -------------------------------------------------------------------------
// 1. Arithmetic Operations
// -------------------------------------------------------------------------

func EvalMult(
	ringQ, ringP *ring.Ring,
	ct0, ct1 Ciphertext,
	evkQ, evkP [][2]ring.Poly,
	bconv BasisConvConstants,
) Ciphertext {
	cttmp := EvalTensor(ringQ, ct0, ct1)
	return Relinearize(ringQ, ringP, cttmp, evkQ, evkP, bconv)
}

func EvalTensor(ringQ *ring.Ring, ct0, ct1 Ciphertext) Ciphertext {
	N := ringQ.N()
	nQ := ringQ.Level() + 1

	cttmp := Ciphertext{Value: []ring.Poly{
		newPoly(ringQ, nQ),
		newPoly(ringQ, nQ),
		newPoly(ringQ, nQ),
	}}

	for i := 0; i < nQ; i++ {
		q := ringQ.ModuliChain()[i]
		for j := 0; j < N; j++ {
			ewuOut := ElementWiseUnit(OpTensor,
				ct0.Value[0].Coeffs[i][j],
				ct1.Value[0].Coeffs[i][j],
				ct0.Value[1].Coeffs[i][j],
				ct1.Value[1].Coeffs[i][j],
				0, 0, 0, q)

			cttmp.Value[0].Coeffs[i][j] = ewuOut.Out0
			cttmp.Value[1].Coeffs[i][j] = ewuOut.Out1
			cttmp.Value[2].Coeffs[i][j] = ewuOut.Out2
		}
	}

	return cttmp
}

// -------------------------------------------------------------------------
// 2. Key Switching — fully self-contained using functional units
// -------------------------------------------------------------------------

func ModUpQtoPUnit(Q, P []uint64, c BasisConvConstants, digitQ ring.Poly, N int) ring.Poly {
	digitP := ring.NewPoly(N, len(P)-1)
	BaseConversionUnit(Q, P, c.QoverQiInvQi, c.QoverQiModP, c.MinusQModP, digitQ, digitP, N)
	return digitP
}

func ModDownQPtoQNTTUnit(
	ringQ, ringP *ring.Ring,
	Q, P []uint64,
	c BasisConvConstants,
	accQ, accP ring.Poly,
	N int,
) ring.Poly {
	nQ := len(Q)
	nP := len(P)

	// Step 1: INTT P-part to Coefficient Domain
	accPCoeff := ring.NewPoly(N, nP-1)
	NTTUnit(ringP, accP, 0, accPCoeff)

	// Step 2: Base Convert P -> Q
	accPQCoeff := newPoly(ringQ, nQ)
	BaseConversionUnit(P, Q, c.PoverPjInvPj, c.PoverPjModQ, c.MinusPModQ, accPCoeff, accPQCoeff, N)

	// Step 3: NTT the converted part back to Frequency Domain
	accPQNtt := newPoly(ringQ, nQ)
	NTTUnit(ringQ, accPQCoeff, 1, accPQNtt)

	// Step 4: Subtract and Scale by P^{-1}
	out := newPoly(ringQ, nQ)
	for i := 0; i < nQ; i++ {
		qi := Q[i]
		qInv := mredParams(qi)
		pInv := c.PInvModQi[i] // Montgomery Form
		for k := 0; k < N; k++ {
			xQ := accQ.Coeffs[i][k]
			xP := accPQNtt.Coeffs[i][k]

			// Standard Modular Subtraction
			ewuOut := ElementWiseUnit(OpModD, xQ, 0, 0, 0, xP, 0, 1, qi)
			// MRed(Standard diff, Montgomery PInv) -> Standard output
			out.Coeffs[i][k] = mRed(ewuOut.Out0, pInv, qi, qInv)
		}
	}

	return out
}

func KeySwitch(
	ringQ, ringP *ring.Ring,
	c2 ring.Poly,
	evkQ, evkP [][2]ring.Poly,
	bconv BasisConvConstants,
) (ring.Poly, ring.Poly) {
	N := ringQ.N()
	nQ := ringQ.Level() + 1
	nP := 0
	if ringP != nil {
		nP = ringP.Level() + 1
	}
	nDigits := len(evkQ)

	Q := ringQ.ModuliChain()[:ringQ.Level()+1]
	P := []uint64{}
	if ringP != nil {
		P = ringP.ModuliChain()[:ringP.Level()+1]
	}
	if nDigits > len(Q) {
		nDigits = len(Q)
	}

	// Accumulators (NTT+Montgomery domain)
	acc0Q := newPoly(ringQ, nQ)
	acc1Q := newPoly(ringQ, nQ)
	acc0P := ring.NewPoly(N, nP-1)
	acc1P := ring.NewPoly(N, nP-1)

	// Step 1: INTT c2 to coefficient domain
	c2Coeff := newPoly(ringQ, nQ)
	NTTUnit(ringQ, c2, 0, c2Coeff)

	for d := 0; d < nDigits; d++ {
		qd := Q[d]
		halfQd := qd >> 1

		// Step 2: Digit extraction + centered broadcast to all Q-limbs
		digitQ := newPoly(ringQ, nQ)
		for i := 0; i < N; i++ {
			val := c2Coeff.Coeffs[d][i]
			isNeg := val > halfQd
			var absVal uint64
			if isNeg {
				absVal = qd - val
			} else {
				absVal = val
			}
			for j := 0; j < nQ; j++ {
				qj := Q[j]
				if isNeg {
					red := absVal % qj
					if red == 0 {
						digitQ.Coeffs[j][i] = 0
					} else {
						digitQ.Coeffs[j][i] = qj - red
					}
				} else {
					digitQ.Coeffs[j][i] = absVal % qj
				}
			}
		}

		// Step 3: ModUp Q → P using functional unit
		digitP := ring.NewPoly(N, nP-1)
		if nP > 0 {
			digitP = ModUpQtoPUnit(Q, P, bconv, digitQ, N)
		}

		// Step 4: NTT both
		NTTUnit(ringQ, digitQ, 1, digitQ)
		if nP > 0 && ringP != nil {
			NTTUnit(ringP, digitP, 1, digitP)
		}

		// Step 5: EWU Accumulation — Q limbs
		for limb := 0; limb < nQ; limb++ {
			q := Q[limb]
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

		// Step 5b: EWU Accumulation — P limbs
		if nP > 0 {
			for limb := 0; limb < nP; limb++ {
				p := P[limb]
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
	}

	// Step 6: ModDown QP → Q using our own functional unit
	out0 := ModDownQPtoQNTTUnit(ringQ, ringP, Q, P, bconv, acc0Q, acc0P, N)
	out1 := ModDownQPtoQNTTUnit(ringQ, ringP, Q, P, bconv, acc1Q, acc1P, N)

	return out0, out1
}

func KeySwitchQP(
	ringQ, ringP *ring.Ring,
	c2 ring.Poly,
	evkQ, evkP [][2]ring.Poly,
	bconv BasisConvConstants,
) (acc0Q, acc0P, acc1Q, acc1P ring.Poly) {
	N := ringQ.N()
	nQ := ringQ.Level() + 1
	nP := ringP.Level() + 1
	nDigits := len(evkQ)

	Q := ringQ.ModuliChain()[:nQ]
	P := ringP.ModuliChain()[:nP]
	if nDigits > len(Q) {
		nDigits = len(Q)
	}

	acc0Q = newPoly(ringQ, nQ)
	acc1Q = newPoly(ringQ, nQ)
	acc0P = ring.NewPoly(N, nP-1)
	acc1P = ring.NewPoly(N, nP-1)

	c2Coeff := newPoly(ringQ, nQ)
	NTTUnit(ringQ, c2, 0, c2Coeff)

	for d := 0; d < nDigits; d++ {
		qd := Q[d]
		halfQd := qd >> 1

		digitQ := newPoly(ringQ, nQ)
		for i := 0; i < N; i++ {
			val := c2Coeff.Coeffs[d][i]
			isNeg := val > halfQd
			var absVal uint64
			if isNeg {
				absVal = qd - val
			} else {
				absVal = val
			}
			for j := 0; j < nQ; j++ {
				qj := Q[j]
				if isNeg {
					red := absVal % qj
					if red == 0 {
						digitQ.Coeffs[j][i] = 0
					} else {
						digitQ.Coeffs[j][i] = qj - red
					}
				} else {
					digitQ.Coeffs[j][i] = absVal % qj
				}
			}
		}

		digitP := ModUpQtoPUnit(Q, P, bconv, digitQ, N)

		NTTUnit(ringQ, digitQ, 1, digitQ)
		NTTUnit(ringP, digitP, 1, digitP)

		for limb := 0; limb < nQ; limb++ {
			q := Q[limb]
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

		for limb := 0; limb < nP; limb++ {
			p := P[limb]
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

// Relinearize reduces degree-2 (D0, D1, D2) → degree-1 (c0, c1)
func Relinearize(
	ringQ, ringP *ring.Ring,
	ctIn Ciphertext,
	evkQ, evkP [][2]ring.Poly,
	bconv BasisConvConstants,
) Ciphertext {
	ks0, ks1 := KeySwitch(ringQ, ringP, ctIn.Value[2], evkQ, evkP, bconv)

	nQ := ringQ.Level() + 1
	out := Ciphertext{Value: []ring.Poly{
		newPoly(ringQ, nQ),
		newPoly(ringQ, nQ),
	}}
	N := ringQ.N()
	Qchain := ringQ.ModuliChain()[:ringQ.Level()+1]

	for limb := 0; limb < nQ; limb++ {
		q := Qchain[limb]
		for i := 0; i < N; i++ {
			ewuOut := ElementWiseUnit(OpMAD,
				ctIn.Value[0].Coeffs[limb][i],
				ctIn.Value[1].Coeffs[limb][i],
				ks0.Coeffs[limb][i],
				ks1.Coeffs[limb][i],
				1, 1, 1, q)
			out.Value[0].Coeffs[limb][i] = ewuOut.Out0
			out.Value[1].Coeffs[limb][i] = ewuOut.Out1
		}
	}

	return out
}

// -------------------------------------------------------------------------
// 3. Data Movement (Rotations)
// -------------------------------------------------------------------------

func EvalRotate(
	ringQ, ringP *ring.Ring,
	ctIn Ciphertext,
	slots int,
	rotKeyQ, rotKeyP [][2]ring.Poly,
	bconv BasisConvConstants,
) Ciphertext {
	nQ := ringQ.Level() + 1
	N := ringQ.N()

	// Step 1: GadgetProduct (KeySwitch) on c1 FIRST
	// This evaluates c1 * \psi^{-1}(s)
	ks0, ks1 := KeySwitch(ringQ, ringP, ctIn.Value[1], rotKeyQ, rotKeyP, bconv)

	// Step 2: Add original c0 to the ks0 result
	tmp0 := newPoly(ringQ, nQ)
	for limb := 0; limb < nQ; limb++ {
		q := ringQ.ModuliChain()[limb]
		for i := 0; i < N; i++ {
			// Hardware: Simple modular adder
			tmp0.Coeffs[limb][i] = ring.CRed(ctIn.Value[0].Coeffs[limb][i]+ks0.Coeffs[limb][i], q)
		}
	}

	out := Ciphertext{Value: []ring.Poly{
		newPoly(ringQ, nQ),
		newPoly(ringQ, nQ),
	}}

	// Step 3: Apply Automorphism to both polynomials
	AutomorphismUnit(ringQ, tmp0, slots, out.Value[0])
	AutomorphismUnit(ringQ, ks1, slots, out.Value[1])

	return out
}

// -------------------------------------------------------------------------
// 4. Rescaling (Modulus Reduction)
// -------------------------------------------------------------------------

// Rescale performs RNS division by the last modulus q_L to reduce the scale.
// It maps to INTT -> Centered Broadcast -> NTT -> OpModD pipeline.
func Rescale(ringQ *ring.Ring, ctIn Ciphertext) Ciphertext {
	N := ringQ.N()
	nQIn := len(ctIn.Value[0].Coeffs)
	L := nQIn - 1 // The index of the modulus being dropped
	nQOut := L    // The number of moduli in the resulting ciphertext

	ringQOut := ringQ.AtLevel(nQOut - 1)

	Q := ringQ.ModuliChain()
	qL := Q[L]
	halfQL := qL >> 1

	// Precompute q_L^{-1} mod q_j for the EWU scaling step
	qLInvModQj := make([]uint64, nQOut)
	for j := 0; j < nQOut; j++ {
		qj := Q[j]
		qL_mod_qj := new(big.Int).SetUint64(qL % qj)
		inv := new(big.Int).ModInverse(qL_mod_qj, new(big.Int).SetUint64(qj))
		qLInvModQj[j] = inv.Uint64()
	}

	ctOut := Ciphertext{Value: make([]ring.Poly, len(ctIn.Value))}
	for i := range ctOut.Value {
		ctOut.Value[i] = newPoly(ringQOut, nQOut)
	}

	for polyIdx := range ctIn.Value {

		// 1. INTT the entire polynomial to the coefficient domain (Uses original Ring)
		cInCoeff := newPoly(ringQ, nQIn)
		NTTUnit(ringQ, ctIn.Value[polyIdx], 0, cInCoeff)

		// 2. Hardware RNS Broadcast Unit: Extract L-th limb and center it
		remCoeff := newPoly(ringQOut, nQOut)
		for i := 0; i < N; i++ {
			val := cInCoeff.Coeffs[L][i]
			isNeg := val > halfQL
			var absVal uint64
			if isNeg {
				absVal = qL - val
			} else {
				absVal = val
			}

			// Broadcast to all remaining j limbs
			for j := 0; j < nQOut; j++ {
				qj := Q[j]
				if isNeg {
					red := absVal % qj
					if red == 0 {
						remCoeff.Coeffs[j][i] = 0
					} else {
						remCoeff.Coeffs[j][i] = qj - red
					}
				} else {
					remCoeff.Coeffs[j][i] = absVal % qj
				}
			}
		}

		// 3. NTT the broadcasted remainders back to the frequency domain
		remNTT := newPoly(ringQOut, nQOut)

		NTTUnit(ringQOut, remCoeff, 1, remNTT)

		// 4. EWU Datapath (OpModD): c_out = (c_in - remNTT) * q_L^{-1} mod q_j
		for j := 0; j < nQOut; j++ {
			qj := Q[j]
			inv := qLInvModQj[j]
			for i := 0; i < N; i++ {
				in0 := ctIn.Value[polyIdx].Coeffs[j][i] // Original NTT limb
				ni := remNTT.Coeffs[j][i]               // Broadcasted NTT remainder

				// Execute OpModD (Subtraction & Multiplication)
				ewuOut := ElementWiseUnit(OpModD, in0, 0, 0, 0, ni, 0, inv, qj)
				ctOut.Value[polyIdx].Coeffs[j][i] = ewuOut.Out0
			}
		}
	}

	return ctOut
}

func EvalAutomorphism(
	ringQ, ringP *ring.Ring,
	ctIn Ciphertext,
	galEl uint64,
	galKeyQ, galKeyP [][2]ring.Poly,
	bconv BasisConvConstants,
) Ciphertext {
	nQ := ringQ.Level() + 1
	N := ringQ.N()

	ks0, ks1 := KeySwitch(ringQ, ringP, ctIn.Value[1], galKeyQ, galKeyP, bconv)

	tmp0 := ring.NewPoly(N, nQ-1)
	for limb := 0; limb < nQ; limb++ {
		q := ringQ.ModuliChain()[limb]
		for i := 0; i < N; i++ {
			tmp0.Coeffs[limb][i] = ring.CRed(ctIn.Value[0].Coeffs[limb][i]+ks0.Coeffs[limb][i], q)
		}
	}

	out := Ciphertext{Value: []ring.Poly{
		ring.NewPoly(N, nQ-1),
		ring.NewPoly(N, nQ-1),
	}}

	AutomorphismGaloisUnit(ringQ, tmp0, galEl, out.Value[0])
	AutomorphismGaloisUnit(ringQ, ks1, galEl, out.Value[1])

	return out
}
