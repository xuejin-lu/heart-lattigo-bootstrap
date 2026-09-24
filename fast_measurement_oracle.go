package main

import (
	"errors"
	"fmt"
	"math"
	"math/big"
)

var (
	errFastMeasurementInvalidModulus = errors.New("modulus must be greater than one")
	errFastMeasurementBasisMismatch  = errors.New("moduli and residues length mismatch")
)

// FastCenteredCRT reconstructs the unique representative in (-S/2, S/2],
// where S is the product of pairwise-coprime moduli.
func FastCenteredCRT(moduli, residues []*big.Int) (FastCenteredCRTResult, error) {
	if len(moduli) == 0 || len(moduli) != len(residues) {
		return FastCenteredCRTResult{}, errFastMeasurementBasisMismatch
	}
	product := big.NewInt(1)
	for i, modulus := range moduli {
		if modulus == nil || modulus.Cmp(big.NewInt(1)) <= 0 {
			return FastCenteredCRTResult{}, fmt.Errorf("modulus %d: %w", i, errFastMeasurementInvalidModulus)
		}
		for j := 0; j < i; j++ {
			if new(big.Int).GCD(nil, nil, modulus, moduli[j]).Cmp(big.NewInt(1)) != 0 {
				return FastCenteredCRTResult{}, fmt.Errorf("moduli %d and %d are not coprime", i, j)
			}
		}
		product.Mul(product, modulus)
	}

	value := new(big.Int)
	for i, modulus := range moduli {
		if residues[i] == nil {
			return FastCenteredCRTResult{}, fmt.Errorf("residue %d cannot be nil", i)
		}
		partial := new(big.Int).Quo(new(big.Int).Set(product), modulus)
		inverse := new(big.Int).ModInverse(partial, modulus)
		if inverse == nil {
			return FastCenteredCRTResult{}, fmt.Errorf("modulus %d has no CRT inverse", i)
		}
		residue := new(big.Int).Mod(new(big.Int).Set(residues[i]), modulus)
		term := new(big.Int).Mul(residue, partial)
		term.Mul(term, inverse)
		value.Add(value, term)
	}
	value.Mod(value, product)
	if new(big.Int).Lsh(new(big.Int).Set(value), 1).Cmp(product) > 0 {
		value.Sub(value, product)
	}

	abs := new(big.Int).Abs(new(big.Int).Set(value))
	return FastCenteredCRTResult{
		Value: value, Product: new(big.Int).Set(product), Negative: value.Sign() < 0,
		AbsBitLength: abs.BitLen(),
	}, nil
}

// FastResiduesFor returns canonical residues of value in each positive modulus.
func FastResiduesFor(value *big.Int, moduli []*big.Int) ([]*big.Int, error) {
	if value == nil || len(moduli) == 0 {
		return nil, errFastMeasurementBasisMismatch
	}
	residues := make([]*big.Int, len(moduli))
	for i, modulus := range moduli {
		if modulus == nil || modulus.Cmp(big.NewInt(1)) <= 0 {
			return nil, fmt.Errorf("modulus %d: %w", i, errFastMeasurementInvalidModulus)
		}
		residues[i] = new(big.Int).Mod(new(big.Int).Set(value), modulus)
	}
	return residues, nil
}

// FastCheckResidueConsistency checks that every active residue represents X.
func FastCheckResidueConsistency(value *big.Int, moduli, residues []*big.Int) bool {
	if value == nil || len(moduli) == 0 || len(moduli) != len(residues) {
		return false
	}
	for i, modulus := range moduli {
		if modulus == nil || modulus.Cmp(big.NewInt(1)) <= 0 || residues[i] == nil {
			return false
		}
		want := new(big.Int).Mod(new(big.Int).Set(value), modulus)
		got := new(big.Int).Mod(new(big.Int).Set(residues[i]), modulus)
		if want.Cmp(got) != 0 {
			return false
		}
	}
	return true
}

// FastCheckLogicalCongruence verifies X = reference (mod logicalModulus).
func FastCheckLogicalCongruence(value, reference, logicalModulus *big.Int) bool {
	if value == nil || reference == nil || logicalModulus == nil || logicalModulus.Cmp(big.NewInt(1)) <= 0 {
		return false
	}
	delta := new(big.Int).Sub(value, reference)
	return new(big.Int).Mod(delta, logicalModulus).Sign() == 0
}

// FastLogicalModulus returns q0*q1*...*q_level. The input slice must contain
// every logical prime through level.
func FastLogicalModulus(logicalPrimes []*big.Int, level int) (*big.Int, error) {
	if level < 0 || level >= len(logicalPrimes) {
		return nil, fmt.Errorf("logical level %d is outside the provided chain", level)
	}
	product := big.NewInt(1)
	for i := 0; i <= level; i++ {
		if logicalPrimes[i] == nil || logicalPrimes[i].Cmp(big.NewInt(1)) <= 0 {
			return nil, fmt.Errorf("logical prime %d: %w", i, errFastMeasurementInvalidModulus)
		}
		for j := 0; j < i; j++ {
			if new(big.Int).GCD(nil, nil, logicalPrimes[i], logicalPrimes[j]).Cmp(big.NewInt(1)) != 0 {
				return nil, fmt.Errorf("logical moduli %d and %d are not coprime", i, j)
			}
		}
		product.Mul(product, logicalPrimes[i])
	}
	return product, nil
}

// FastCapacityEvidence checks the strict centered uniqueness condition
// 2*bound < storageProduct. HeadroomBits is presentation-only and nil when
// bound is zero (unbounded positive headroom).
func FastCapacityEvidence(bound, storageProduct *big.Int) (FastCapacityResult, error) {
	if bound == nil || bound.Sign() < 0 {
		return FastCapacityResult{}, errors.New("magnitude bound must be non-negative")
	}
	if storageProduct == nil || storageProduct.Cmp(big.NewInt(1)) <= 0 {
		return FastCapacityResult{}, errFastMeasurementInvalidModulus
	}
	unique := new(big.Int).Lsh(new(big.Int).Set(bound), 1).Cmp(storageProduct) < 0
	result := FastCapacityResult{
		Unique: unique, BoundBitLength: bound.BitLen(),
		StorageProductBitLength:   storageProduct.BitLen(),
		CenteredCapacityBitLength: new(big.Int).Rsh(new(big.Int).Set(storageProduct), 1).BitLen(),
		CenteredCapacityLog2:      log2BigInt(storageProduct) - 1,
	}
	if bound.Sign() == 0 {
		result.UnboundedHeadroom = true
		return result, nil
	}
	headroom := log2BigInt(storageProduct) - 1 - log2BigInt(bound)
	result.HeadroomBits = &headroom
	return result, nil
}

// FastLogicalRescaleOracle performs signed nearest division for a positive odd
// logical divisor. With odd q, exact half-way ties cannot occur; the rule
// matches the current Fast/Standard behavior of incrementing only when
// remainder > floor(q/2).
func FastLogicalRescaleOracle(value, logicalDivisor *big.Int) (*big.Int, error) {
	if value == nil {
		return nil, errors.New("rescale input cannot be nil")
	}
	if logicalDivisor == nil || logicalDivisor.Cmp(big.NewInt(1)) <= 0 || logicalDivisor.Bit(0) == 0 {
		return nil, errors.New("logical rescale divisor must be greater than one and odd")
	}
	negative := value.Sign() < 0
	magnitude := new(big.Int).Abs(new(big.Int).Set(value))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(magnitude, logicalDivisor, remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(logicalDivisor) > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if negative {
		quotient.Neg(quotient)
	}
	return quotient, nil
}

// FastStorageContractionOracle proves that target storage can uniquely
// represent value and, when legal, verifies the exact centered CRT round-trip.
func FastStorageContractionOracle(value *big.Int, targetModuli []*big.Int) (FastContractionResult, error) {
	if value == nil || len(targetModuli) == 0 {
		return FastContractionResult{}, errFastMeasurementBasisMismatch
	}
	product := big.NewInt(1)
	for i, modulus := range targetModuli {
		if modulus == nil || modulus.Cmp(big.NewInt(1)) <= 0 {
			return FastContractionResult{}, fmt.Errorf("modulus %d: %w", i, errFastMeasurementInvalidModulus)
		}
		for j := 0; j < i; j++ {
			if new(big.Int).GCD(nil, nil, modulus, targetModuli[j]).Cmp(big.NewInt(1)) != 0 {
				return FastContractionResult{}, fmt.Errorf("target moduli %d and %d are not coprime", i, j)
			}
		}
		product.Mul(product, modulus)
	}
	abs := new(big.Int).Abs(new(big.Int).Set(value))
	if new(big.Int).Lsh(new(big.Int).Set(abs), 1).Cmp(product) >= 0 {
		return FastContractionResult{Allowed: false, TargetProduct: product}, nil
	}
	residues, err := FastResiduesFor(value, targetModuli)
	if err != nil {
		return FastContractionResult{}, err
	}
	reconstructed, err := FastCenteredCRT(targetModuli, residues)
	if err != nil {
		return FastContractionResult{}, err
	}
	return FastContractionResult{
		Allowed: reconstructed.Value.Cmp(value) == 0, TargetProduct: product,
		Residues: residues, RoundTrip: new(big.Int).Set(reconstructed.Value),
	}, nil
}

// FastModUpCanonicalRepresentative returns Center(X mod logicalModulus) in
// (-Q/2, Q/2]. This must be the representative extended across ModUp.
func FastModUpCanonicalRepresentative(value, logicalModulus *big.Int) (*big.Int, error) {
	if value == nil || logicalModulus == nil || logicalModulus.Cmp(big.NewInt(1)) <= 0 {
		return nil, errFastMeasurementInvalidModulus
	}
	canonical := new(big.Int).Mod(new(big.Int).Set(value), logicalModulus)
	if new(big.Int).Lsh(new(big.Int).Set(canonical), 1).Cmp(logicalModulus) > 0 {
		canonical.Sub(canonical, logicalModulus)
	}
	return canonical, nil
}

func log2BigInt(value *big.Int) float64 {
	if value == nil || value.Sign() <= 0 {
		return math.Inf(-1)
	}
	abs := new(big.Int).Abs(new(big.Int).Set(value))
	shift := abs.BitLen() - 53
	if shift < 0 {
		shift = 0
	}
	mantissa := new(big.Int).Rsh(abs, uint(shift))
	f, _ := new(big.Float).SetInt(mantissa).Float64()
	return float64(shift) + math.Log2(f)
}
