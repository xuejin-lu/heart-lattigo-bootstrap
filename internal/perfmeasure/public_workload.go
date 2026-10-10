package perfmeasure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

// PublicMulRescaleWorkload is the frozen Batch019 public-operation fixture.
// A remains the canonical LogN13 input used by the earlier measurement path.
func PublicMulRescaleWorkload(slots int) (a, b, c []complex128, err error) {
	if slots <= 0 {
		return nil, nil, nil, errors.New("public workload requires a positive slot count")
	}
	a = DeterministicInput(slots)
	b, c = make([]complex128, slots), make([]complex128, slots)
	for i := range b {
		b[i] = complex(float64(i%19-9)/384, float64((5*i)%17-8)/768)
		c[i] = a[i]
	}
	return a, b, c, nil
}

// PublicMulRescaleWorkloadFingerprint hashes the exact ordered complex inputs
// and the scale of C without serializing large vectors into result metadata.
func PublicMulRescaleWorkloadFingerprint(a, b, c []complex128, cScale string) (string, error) {
	if len(a) == 0 || len(a) != len(b) || len(a) != len(c) || cScale == "" {
		return "", errors.New("public workload vectors must have matching non-zero lengths and a C scale")
	}
	toPairs := func(values []complex128) []struct {
		Real float64 `json:"real"`
		Imag float64 `json:"imag"`
	} {
		pairs := make([]struct {
			Real float64 `json:"real"`
			Imag float64 `json:"imag"`
		}, len(values))
		for i, value := range values {
			pairs[i].Real, pairs[i].Imag = real(value), imag(value)
		}
		return pairs
	}
	data, err := json.Marshal(struct {
		A, B, C []struct {
			Real float64 `json:"real"`
			Imag float64 `json:"imag"`
		}
		CScale string
	}{toPairs(a), toPairs(b), toPairs(c), cScale})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

type PublicCapacityPlan struct {
	InputA     string `json:"input_a_bound"`
	InputB     string `json:"input_b_bound"`
	InputC     string `json:"input_c_bound"`
	Add        string `json:"add_bound"`
	MulRelin   string `json:"mulrelin_bound"`
	Rescale    string `json:"rescale_bound"`
	Q0123      string `json:"q0123_product"`
	Q0         string `json:"q0"`
	MulDivisor string `json:"mul_divisor_q5"`
}

// BoundPublicMulRescaleWorkload derives the exact independent coefficient
// bounds used by the Batch019 fixture and enforces both Q0123 and terminal q0
// centered-capacity gates before the public chain is executed.
func BoundPublicMulRescaleWorkload(a, b, c []complex128, q []uint64, ringN int) (PublicCapacityPlan, error) {
	if len(a) == 0 || len(a) != len(b) || len(a) != len(c) || len(q) <= 5 || ringN <= 0 {
		return PublicCapacityPlan{}, errors.New("public capacity plan requires q5 and a positive ring degree")
	}
	ba, err := encodedInputBoundExact(a, new(big.Int).Lsh(big.NewInt(1), 45))
	if err != nil {
		return PublicCapacityPlan{}, fmt.Errorf("input A bound: %w", err)
	}
	bb, err := encodedInputBoundExact(b, new(big.Int).Lsh(big.NewInt(1), 45))
	if err != nil {
		return PublicCapacityPlan{}, fmt.Errorf("input B bound: %w", err)
	}
	bc, err := encodedInputBoundExact(c, new(big.Int).SetUint64(q[5]))
	if err != nil {
		return PublicCapacityPlan{}, fmt.Errorf("input C bound: %w", err)
	}
	badd := new(big.Int).Add(new(big.Int).Set(ba), bb)
	bmul := new(big.Int).Mul(big.NewInt(int64(ringN)), new(big.Int).Mul(new(big.Int).Set(badd), bc))
	q5 := new(big.Int).SetUint64(q[5])
	br := new(big.Int).Add(new(big.Int).Set(bmul), new(big.Int).Rsh(new(big.Int).Sub(new(big.Int).Set(q5), big.NewInt(1)), 1))
	br.Quo(br, q5)
	prefix := prefixProduct(q, 4)
	for _, item := range []struct {
		name  string
		bound *big.Int
	}{{"EncryptNew(A)", ba}, {"EncryptNew(B)", bb}, {"EncryptNew(C,q5 scale)", bc}, {"AddNew", badd}, {"MulRelinNew", bmul}, {"Rescale(q5)", br}} {
		if !strictCapacity(item.bound, prefix) {
			return PublicCapacityPlan{}, fmt.Errorf("Q0123 capacity failure at %s: B=%s S_Q0123=%s deficit=%s", item.name, item.bound, prefix, new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).Set(item.bound), 1), prefix))
		}
	}
	q0 := new(big.Int).SetUint64(q[0])
	if !strictCapacity(br, q0) {
		return PublicCapacityPlan{}, fmt.Errorf("terminal Level0 q0 capacity failure: B=%s q0=%s deficit=%s", br, q0, new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).Set(br), 1), q0))
	}
	return PublicCapacityPlan{
		InputA: ba.String(), InputB: bb.String(), InputC: bc.String(), Add: badd.String(),
		MulRelin: bmul.String(), Rescale: br.String(), Q0123: prefix.String(), Q0: q0.String(), MulDivisor: q5.String(),
	}, nil
}

func encodedInputBoundExact(values []complex128, scale *big.Int) (*big.Int, error) {
	var maxSquared *big.Rat
	for _, value := range values {
		r := new(big.Rat).SetFloat64(real(value))
		i := new(big.Rat).SetFloat64(imag(value))
		if r == nil || i == nil {
			return nil, errors.New("non-finite workload coordinate")
		}
		squared := new(big.Rat).Add(new(big.Rat).Mul(r, r), new(big.Rat).Mul(i, i))
		if maxSquared == nil || squared.Cmp(maxSquared) > 0 {
			maxSquared = squared
		}
	}
	if maxSquared == nil {
		return nil, errors.New("empty workload")
	}
	// ceil(2*scale*sqrt(maxSquared)+1) = ceil(sqrt(4*scale^2*maxSquared))+1.
	twiceScale := new(big.Int).Lsh(new(big.Int).Set(scale), 1)
	numerator := new(big.Int).Mul(new(big.Int).Set(twiceScale), twiceScale)
	numerator.Mul(numerator, maxSquared.Num())
	denominator := maxSquared.Denom()
	root := integerSqrt(new(big.Int).Quo(new(big.Int).Set(numerator), denominator))
	square := new(big.Int).Mul(new(big.Int).Set(root), root)
	square.Mul(square, denominator)
	if square.Cmp(numerator) < 0 {
		root.Add(root, big.NewInt(1))
	}
	return root.Add(root, big.NewInt(1)), nil
}

func integerSqrt(n *big.Int) *big.Int {
	if n.Sign() < 0 {
		panic("integerSqrt of negative input")
	}
	if n.Sign() == 0 {
		return big.NewInt(0)
	}
	x := new(big.Int).Lsh(big.NewInt(1), uint((n.BitLen()+1)/2))
	for {
		y := new(big.Int).Rsh(new(big.Int).Add(x, new(big.Int).Quo(n, x)), 1)
		if y.Cmp(x) >= 0 {
			return x
		}
		x = y
	}
}

func prefixProduct(q []uint64, rows int) *big.Int {
	out := big.NewInt(1)
	for _, prime := range q[:rows] {
		out.Mul(out, new(big.Int).SetUint64(prime))
	}
	return out
}

func strictCapacity(bound, product *big.Int) bool {
	return bound != nil && product != nil && bound.Sign() >= 0 && product.Sign() > 0 && new(big.Int).Lsh(new(big.Int).Set(bound), 1).Cmp(product) < 0
}
