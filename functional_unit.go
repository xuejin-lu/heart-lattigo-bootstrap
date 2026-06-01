package main

import (
	"math"
	"math/big"
	"math/bits"
	"sync"

	"github.com/tuneinsight/lattigo/v6/ring"
)

// -------------------------------------------------------------------------
// Montgomery arithmetic helpers
// -------------------------------------------------------------------------

type montgomeryConstants struct {
	negQInv uint64
	bred    [2]uint64
}

var montgomeryConstantsCache sync.Map

func getMontgomeryConstants(q uint64) montgomeryConstants {
	if cached, ok := montgomeryConstantsCache.Load(q); ok {
		return cached.(montgomeryConstants)
	}

	constants := montgomeryConstants{
		negQInv: genNegMRedConstant(q),
		bred:    ring.GenBRedConstant(q),
	}
	actual, _ := montgomeryConstantsCache.LoadOrStore(q, constants)
	return actual.(montgomeryConstants)
}

func genNegMRedConstant(q uint64) uint64 {
	qInv := uint64(1)
	qPow := q
	for i := 0; i < 63; i++ {
		qInv *= qPow
		qPow *= qPow
	}
	return -qInv
}

func mredParams(q uint64) uint64 {
	return getMontgomeryConstants(q).negQInv
}

func mForm(a, q uint64) uint64 {
	return ring.MForm(a, q, getMontgomeryConstants(q).bred)
}

func mRed(x, y, q, qInv uint64) uint64 {
	hi, lo := bits.Mul64(x, y)
	m := lo * qInv
	mHi, mLo := bits.Mul64(m, q)
	_, carry := bits.Add64(lo, mLo, 0)
	u, _ := bits.Add64(hi, mHi, carry)
	if u >= q {
		u -= q
	}
	return u
}

// -------------------------------------------------------------------------
// 1. NTT Unit
// -------------------------------------------------------------------------

func NTTUnit(r *ring.Ring, polIn ring.Poly, isNTT int, polOut ring.Poly) {
	countNTTUnit(isNTT)
	if isNTT == 1 {
		r.NTT(polIn, polOut)
	} else {
		r.INTT(polIn, polOut)
	}
}

// -------------------------------------------------------------------------
// 2. Base Unit and its constants(precomputed, stored in hardware ROM )
// -------------------------------------------------------------------------

// BasisConvConstants holds the precomputed constants for ModUp and ModDown CRT trees.
type BasisConvConstants struct {
	// ModUp Q→P constants
	QoverQiInvQi []uint64   // (Q/q_i)^{-1} mod q_i  (Montgomery form)
	QoverQiModP  [][]uint64 // (Q/q_i) mod p_j        (Montgomery form)
	MinusQModP   []uint64   // (-Q) mod p_j           (Standard form) - Replaces 2D vTimes

	// ModDown P→Q constants
	PoverPjInvPj []uint64   // (P/p_j)^{-1} mod p_j  (Montgomery form)
	PoverPjModQ  [][]uint64 // (P/p_j) mod q_i        (Montgomery form)
	MinusPModQ   []uint64   // (-P) mod q_i           (Standard form) - Replaces 2D vTimes

	// ModDown Scaling constant
	PInvModQi []uint64 // P^{-1} mod q_i (Montgomery form)
}

func PrecomputeBasisConvConstants(Q, P []uint64) BasisConvConstants {
	nQ := len(Q)
	nP := len(P)
	var c BasisConvConstants

	Qfull := big.NewInt(1)
	for _, qi := range Q {
		Qfull.Mul(Qfull, new(big.Int).SetUint64(qi))
	}
	Pfull := big.NewInt(1)
	for _, pj := range P {
		Pfull.Mul(Pfull, new(big.Int).SetUint64(pj))
	}

	// --- Q -> P Constants ---
	c.QoverQiInvQi = make([]uint64, nQ)
	c.QoverQiModP = make([][]uint64, nP)
	c.MinusQModP = make([]uint64, nP)

	for j := 0; j < nP; j++ {
		pj := P[j]
		c.QoverQiModP[j] = make([]uint64, nQ)

		QmodPj := new(big.Int).Mod(Qfull, new(big.Int).SetUint64(pj)).Uint64()
		if QmodPj == 0 {
			c.MinusQModP[j] = 0
		} else {
			c.MinusQModP[j] = pj - QmodPj // Computes -Q mod p_j
		}
	}

	for i := 0; i < nQ; i++ {
		qi := Q[i]
		BigQi := new(big.Int).SetUint64(qi)
		QdivQi := new(big.Int).Div(Qfull, BigQi)

		QdivQiModQi := new(big.Int).Mod(QdivQi, BigQi)
		inv := new(big.Int).ModInverse(QdivQiModQi, BigQi)
		c.QoverQiInvQi[i] = mForm(inv.Uint64(), qi)

		for j := 0; j < nP; j++ {
			pj := P[j]
			QdivQiModPj := new(big.Int).Mod(QdivQi, new(big.Int).SetUint64(pj))
			c.QoverQiModP[j][i] = QdivQiModPj.Uint64() // Standard form for plain MAC tree
		}
	}

	// --- P -> Q Constants (For ModDown) ---
	c.PoverPjInvPj = make([]uint64, nP)
	c.PoverPjModQ = make([][]uint64, nQ)
	c.MinusPModQ = make([]uint64, nQ)

	for i := 0; i < nQ; i++ {
		qi := Q[i]
		c.PoverPjModQ[i] = make([]uint64, nP)

		PmodQi := new(big.Int).Mod(Pfull, new(big.Int).SetUint64(qi)).Uint64()
		if PmodQi == 0 {
			c.MinusPModQ[i] = 0
		} else {
			c.MinusPModQ[i] = qi - PmodQi // Computes -P mod q_i
		}
	}

	for j := 0; j < nP; j++ {
		pj := P[j]
		BigPj := new(big.Int).SetUint64(pj)
		PdivPj := new(big.Int).Div(Pfull, BigPj)

		PdivPjModPj := new(big.Int).Mod(PdivPj, BigPj)
		inv := new(big.Int).ModInverse(PdivPjModPj, BigPj)
		c.PoverPjInvPj[j] = mForm(inv.Uint64(), pj)

		for i := 0; i < nQ; i++ {
			qi := Q[i]
			PdivPjModQi := new(big.Int).Mod(PdivPj, new(big.Int).SetUint64(qi))
			c.PoverPjModQ[i][j] = PdivPjModQi.Uint64() // Standard form for plain MAC tree
		}
	}

	// --- ModDown Scaling Constants ---
	c.PInvModQi = make([]uint64, nQ)
	for i := 0; i < nQ; i++ {
		qi := Q[i]
		PmodQi := new(big.Int).Mod(Pfull, new(big.Int).SetUint64(qi))
		Pinv := new(big.Int).ModInverse(PmodQi, new(big.Int).SetUint64(qi))
		c.PInvModQi[i] = mForm(Pinv.Uint64(), qi)
	}

	return c
}

// BaseConversionUnit represents the optimized CRT reconstruction hardware.
func BaseConversionUnit(fromBase, toBase []uint64, inv []uint64, modTo [][]uint64, minusBaseModTo []uint64, digitFrom, digitTo ring.Poly, N int) {
	countBaseConversionUnit()
	nFrom := len(fromBase)
	nTo := len(toBase)

	// Lattigo reciprocal multiplier ROM configuration
	floatInv := make([]float64, nFrom)
	for i, q := range fromBase {
		floatInv[i] = 1.0 / float64(q)
	}

	for coeffIdx := 0; coeffIdx < N; coeffIdx++ {
		var y float64
		xi := make([]uint64, nFrom)

		// 1. Calculate x_i and Shenoy-Kumaresan Float Tracker
		for i := 0; i < nFrom; i++ {
			qi := fromBase[i]
			qInv := mredParams(qi)
			xi[i] = mRed(digitFrom.Coeffs[i][coeffIdx], inv[i], qi, qInv)
			y += float64(xi[i]) * floatInv[i]
		}

		v := uint64(math.Round(y))

		// 2. Hardware MAC Tree + Dynamic Correction logic

		for j := 0; j < nTo; j++ {
			pj := toBase[j]
			bigPj := new(big.Int).SetUint64(pj)

			// Wide accumulator (big.Int to handle multi-word intermediate)
			acc := new(big.Int)
			for i := 0; i < nFrom; i++ {
				// Plain multiply: xi[i] * modTo[j][i]
				term := new(big.Int).Mul(
					new(big.Int).SetUint64(xi[i]),
					new(big.Int).SetUint64(modTo[j][i]),
				)
				acc.Add(acc, term)
			}

			acc.Mod(acc, bigPj)

			// Add correction: v * (-Base mod p_j)
			correction := new(big.Int).Mul(
				new(big.Int).SetUint64(v),
				new(big.Int).SetUint64(minusBaseModTo[j]),
			)
			acc.Add(acc, correction)

			// Single modular reduction at the end
			acc.Mod(acc, bigPj)

			digitTo.Coeffs[j][coeffIdx] = acc.Uint64()
		}
	}
}

// -------------------------------------------------------------------------
// 3. Automorphism Unit
// -------------------------------------------------------------------------

func AutomorphismUnit(r *ring.Ring, polIn ring.Poly, rotateSlots int, polOut ring.Poly) {
	countAutomorphismUnit()
	gen := calculateGenerator(r.N(), rotateSlots)
	r.AutomorphismNTT(polIn, gen, polOut)
}

func AutomorphismGaloisUnit(r *ring.Ring, polIn ring.Poly, galEl uint64, polOut ring.Poly) {
	countAutomorphismUnit()
	r.AutomorphismNTT(polIn, galEl, polOut)
}

func calculateGenerator(N int, slots int) uint64 {
	return uint64(math.Pow(5, float64(slots)))
}

// -------------------------------------------------------------------------
// 4. Element-Wise Unit (EWU)
// -------------------------------------------------------------------------

const (
	OpTensor = iota
	OpAccQ
	OpAccP
	OpModD
	OpMAD
)

type EWUOutputs struct {
	Out0 uint64
	Out1 uint64
	Out2 uint64
}

func ElementWiseUnit(opCode int, in0, in1, in2, in3, ni, pi, ci, q uint64) EWUOutputs {
	countElementWiseUnit(opCode)
	var out EWUOutputs

	qInv := mredParams(q)

	modAdd := func(a, b uint64) uint64 { return ring.CRed(a+b, q) }
	modSub := func(a, b uint64) uint64 { return ring.CRed(a+q-b, q) }
	modMul := func(a, b uint64) uint64 { return mRed(a, b, q, qInv) }

	switch opCode {
	case OpTensor:
		in0 = mForm(in0, q)
		in2 = mForm(in2, q)
		out.Out2 = modMul(in2, in3)
		term1 := modMul(in0, in3)
		term2 := modMul(in1, in2)
		out.Out1 = modAdd(term1, term2)
		out.Out0 = modMul(in0, in1)

	case OpAccQ:
		ciMont := mForm(ci, q)
		term1 := modMul(in3, in2)
		term2 := modMul(ciMont, in0)
		out.Out0 = modAdd(term1, term2)
		term3 := modMul(in3, pi)
		term4 := modMul(ciMont, in1)
		out.Out1 = modAdd(term3, term4)

	case OpAccP:
		term1 := modMul(ni, in2)
		out.Out0 = modAdd(term1, in0)
		term2 := modMul(ni, pi)
		out.Out1 = modAdd(term2, in1)

	case OpModD:
		ciMont := mForm(ci, q)
		term1 := modMul(ciMont, in0)
		term2 := modMul(ciMont, ni)
		out.Out0 = modSub(term1, term2)

	case OpMAD:
		niMont := mForm(ni, q)
		ciMont := mForm(ci, q)
		term1 := modMul(niMont, in2)
		term2 := modMul(ciMont, in0)
		out.Out0 = modAdd(term1, term2)
		term3 := modMul(niMont, in3)
		term4 := modMul(ciMont, in1)
		out.Out1 = modAdd(term3, term4)
	}

	return out
}

// -------------------------------------------------------------------------
// 5. Double-prime Scaling Unit (DSU)
// -------------------------------------------------------------------------

func DoublePrimeScalingUnit(a, b, qi, qiPlus1, qj uint64) uint64 {
	countDoublePrimeScalingUnit()
	bigA := new(big.Int).SetUint64(a)
	bigB := new(big.Int).SetUint64(b)
	bigQi := new(big.Int).SetUint64(qi)
	bigQiPlus1 := new(big.Int).SetUint64(qiPlus1)
	bigQj := new(big.Int).SetUint64(qj)
	term1 := new(big.Int).Mul(bigA, bigQiPlus1)
	term2 := new(big.Int).Mul(bigB, bigQi)
	sum := new(big.Int).Add(term1, term2)
	result := new(big.Int).Mod(sum, bigQj)
	return result.Uint64()
}
