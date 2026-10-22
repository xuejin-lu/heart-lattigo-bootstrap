package perfmeasure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
)

const PublicDirectMulWorkloadID = "direct-ctct-v1"

// PublicDirectMulWorkload returns the separately defined A, B and C messages
// for the direct ciphertext-by-ciphertext multiplication experiment. B reuses
// the maintained public Mul/Rescale workload values; its legacy Add operation
// is not part of this workload.
func PublicDirectMulWorkload(slots int, variant string) (a, b, c []complex128, err error) {
	if slots <= 0 {
		return nil, nil, nil, errors.New("direct-mul workload requires a positive slot count")
	}
	if variant != "direct-base" && variant != "direct-c4" {
		return nil, nil, nil, fmt.Errorf("unsupported direct-mul workload variant %q", variant)
	}
	a = DeterministicInput(slots)
	_, b, _, err = PublicMulRescaleWorkload(slots)
	if err != nil {
		return nil, nil, nil, err
	}
	c = make([]complex128, slots)
	for i := range c {
		c[i] = complex(float64((7*i+3)%23-11)/512, float64((11*i+5)%29-14)/1024)
		if variant == "direct-c4" {
			c[i] *= 4
		}
	}
	return a, b, c, nil
}

// PublicDirectMulInputsFingerprint gives the new workload an identity distinct
// from the historical Add-containing fixture while keeping vector data out of
// result metadata.
func PublicDirectMulInputsFingerprint(a, b, c []complex128) (string, error) {
	if len(a) == 0 || len(a) != len(b) || len(a) != len(c) {
		return "", errors.New("direct-mul inputs must have matching non-zero lengths")
	}
	for _, input := range []struct {
		name   string
		values []complex128
	}{{"A", a}, {"B", b}, {"C", c}} {
		for i, value := range input.values {
			if math.IsNaN(real(value)) || math.IsInf(real(value), 0) || math.IsNaN(imag(value)) || math.IsInf(imag(value), 0) {
				return "", fmt.Errorf("direct-mul input %s[%d] is non-finite", input.name, i)
			}
		}
	}
	data, err := json.Marshal(struct {
		SchemaVersion string `json:"schema_version"`
		WorkloadID    string `json:"workload_id"`
		Slots         int    `json:"slots"`
		A             string `json:"a_sha256"`
		B             string `json:"b_sha256"`
		C             string `json:"c_sha256"`
	}{"public-direct-mul-inputs.v1", PublicDirectMulWorkloadID, len(a), Fingerprint(a), Fingerprint(b), Fingerprint(c)})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// PublicDirectMulWorkloadFingerprint binds exact messages, the two Rescale
// divisors/scales, effective Q/P provenance and the versioned call schedule.
func PublicDirectMulWorkloadFingerprint(a, b, c []complex128, variant, deltaScale, q5, q4, qpSHA256 string) (string, error) {
	if variant != "direct-base" && variant != "direct-c4" || deltaScale == "" || q5 == "" || q4 == "" || !isSHA256Hex(qpSHA256) {
		return "", errors.New("direct-mul fingerprint requires a supported variant, exact scales, and a Q/P SHA-256")
	}
	inputsSHA256, err := PublicDirectMulInputsFingerprint(a, b, c)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(struct {
		SchemaVersion string `json:"schema_version"`
		WorkloadID    string `json:"workload_id"`
		Variant       string `json:"variant"`
		InputsSHA256  string `json:"inputs_sha256"`
		QPSHA256      string `json:"qp_sha256"`
		Stage1        string `json:"stage1"`
		Stage2        string `json:"stage2"`
	}{
		SchemaVersion: "public-direct-mul-workload.v1", WorkloadID: PublicDirectMulWorkloadID,
		Variant: variant, InputsSHA256: inputsSHA256, QPSHA256: qpSHA256,
		Stage1: "EncryptNew(A,Level5,Scale=" + deltaScale + ");EncryptNew(B,Level5,Scale=q5);MulRelinNew(A_ct,B_ct);Rescale(divisor=q5,output=Level4/Scale=" + deltaScale + ")",
		Stage2: "EncryptNew(C,Level4,Scale=q4);MulRelinNew(X,C_ct);Rescale(divisor=q4,output=Level3/Scale=" + deltaScale + ")",
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

type PublicDirectMulCapacityCheck struct {
	Operation string `json:"operation"`
	Level     int    `json:"level"`
	QRows     int    `json:"active_q_prefix_rows"`
	Bound     string `json:"centered_bound"`
	Capacity  string `json:"q_prefix_capacity"`
	StrictFit bool   `json:"strict_2b_lt_capacity"`
}

type PublicDirectMulStageCapacity struct {
	InputXBound       string `json:"input_x_bound"`
	InputYBound       string `json:"input_y_bound"`
	MulRelinBound     string `json:"mulrelin_bound"`
	RescaleDivisor    string `json:"rescale_divisor"`
	RescaleBound      string `json:"rescale_bound"`
	InputLevel        int    `json:"input_level"`
	OutputLevel       int    `json:"output_level"`
	ActiveQPrefixRows int    `json:"active_q_prefix_rows"`
	QPrefixCapacity   string `json:"q_prefix_capacity"`
}

type PublicDirectMulCapacityPlan struct {
	SchemaVersion string                         `json:"schema_version"`
	WorkloadID    string                         `json:"workload_id"`
	RingN         int                            `json:"ring_n"`
	Q4            string                         `json:"q4"`
	Q5            string                         `json:"q5"`
	Q0123         string                         `json:"q0123_product"`
	Stage1        PublicDirectMulStageCapacity   `json:"stage1"`
	Stage2        PublicDirectMulStageCapacity   `json:"stage2"`
	Checks        []PublicDirectMulCapacityCheck `json:"checks"`
	Status        string                         `json:"status"`
}

// BoundPublicDirectMulChain derives stage-by-stage centered-lift bounds for
// A×B followed by (A×B)×C. It reuses the maintained conservative encoded-input
// and negacyclic-product bounds, and requires strict fit at every Level/Q-prefix
// boundary touched by this Level5→4→3 schedule.
func BoundPublicDirectMulChain(a, b, c []complex128, q []uint64, ringN int) (PublicDirectMulCapacityPlan, error) {
	if len(a) == 0 || len(a) != len(b) || len(a) != len(c) || len(q) <= 5 || ringN <= 0 {
		return PublicDirectMulCapacityPlan{}, errors.New("direct-mul capacity plan requires matching inputs, q4/q5, and a positive ring degree")
	}
	if q[4] < 3 || q[5] < 3 || q[4]%2 == 0 || q[5]%2 == 0 {
		return PublicDirectMulCapacityPlan{}, errors.New("direct-mul q4/q5 Rescale divisors must be odd integers greater than two")
	}
	delta := new(big.Int).Lsh(big.NewInt(1), 45)
	q5, q4 := new(big.Int).SetUint64(q[5]), new(big.Int).SetUint64(q[4])
	ba, err := encodedInputBoundExact(a, delta)
	if err != nil {
		return PublicDirectMulCapacityPlan{}, fmt.Errorf("EncryptNew(A) bound: %w", err)
	}
	bb, err := encodedInputBoundExact(b, q5)
	if err != nil {
		return PublicDirectMulCapacityPlan{}, fmt.Errorf("EncryptNew(B,q5 scale) bound: %w", err)
	}
	bc, err := encodedInputBoundExact(c, q4)
	if err != nil {
		return PublicDirectMulCapacityPlan{}, fmt.Errorf("EncryptNew(C,q4 scale) bound: %w", err)
	}
	n := big.NewInt(int64(ringN))
	bMul1 := new(big.Int).Mul(n, new(big.Int).Mul(new(big.Int).Set(ba), bb))
	bRescale1 := rescaleBoundExact(bMul1, q5)
	bMul2 := new(big.Int).Mul(n, new(big.Int).Mul(new(big.Int).Set(bRescale1), bc))
	bRescale2 := rescaleBoundExact(bMul2, q4)

	capacityAt := func(level int) (*big.Int, int, error) {
		rows := min(level+1, 4)
		if rows <= 0 || len(q) < rows {
			return nil, rows, fmt.Errorf("Level%d has no valid Q-prefix rows", level)
		}
		return prefixProduct(q, rows), rows, nil
	}
	plan := PublicDirectMulCapacityPlan{
		SchemaVersion: "public-direct-mul-qprefix-capacity.v1", WorkloadID: PublicDirectMulWorkloadID,
		RingN: ringN, Q4: q4.String(), Q5: q5.String(), Q0123: prefixProduct(q, 4).String(),
		Stage1: PublicDirectMulStageCapacity{
			InputXBound: ba.String(), InputYBound: bb.String(), MulRelinBound: bMul1.String(),
			RescaleDivisor: q5.String(), RescaleBound: bRescale1.String(), InputLevel: 5, OutputLevel: 4,
			ActiveQPrefixRows: 4, QPrefixCapacity: prefixProduct(q, 4).String(),
		},
		Stage2: PublicDirectMulStageCapacity{
			InputXBound: bRescale1.String(), InputYBound: bc.String(), MulRelinBound: bMul2.String(),
			RescaleDivisor: q4.String(), RescaleBound: bRescale2.String(), InputLevel: 4, OutputLevel: 3,
			ActiveQPrefixRows: 4, QPrefixCapacity: prefixProduct(q, 4).String(),
		},
		Checks: make([]PublicDirectMulCapacityCheck, 0, 8), Status: "STRICT_QPREFIX_FIT",
	}
	checks := []struct {
		name  string
		level int
		bound *big.Int
	}{
		{"EncryptNew(A)", 5, ba}, {"EncryptNew(B,q5)", 5, bb},
		{"stage1.MulRelinNew", 5, bMul1}, {"stage1.Rescale(q5)", 4, bRescale1},
		{"EncryptNew(C,q4)", 4, bc}, {"stage2.input_X", 4, bRescale1},
		{"stage2.MulRelinNew", 4, bMul2}, {"stage2.Rescale(q4)", 3, bRescale2},
	}
	for _, check := range checks {
		capacity, rows, err := capacityAt(check.level)
		if err != nil {
			return PublicDirectMulCapacityPlan{}, err
		}
		fits := strictCapacity(check.bound, capacity)
		plan.Checks = append(plan.Checks, PublicDirectMulCapacityCheck{
			Operation: check.name, Level: check.level, QRows: rows,
			Bound: check.bound.String(), Capacity: capacity.String(), StrictFit: fits,
		})
		if !fits {
			plan.Status = "UNSUPPORTED_CAPACITY"
			return plan, fmt.Errorf("Q-prefix strict capacity failure at %s Level%d: 2B=%s capacity=%s", check.name, check.level,
				new(big.Int).Lsh(new(big.Int).Set(check.bound), 1), capacity)
		}
	}
	return plan, nil
}

func rescaleBoundExact(bound, divisor *big.Int) *big.Int {
	if divisor.Sign() <= 0 {
		return nil
	}
	nearestHalf := new(big.Int).Rsh(new(big.Int).Sub(new(big.Int).Set(divisor), big.NewInt(1)), 1)
	return new(big.Int).Quo(new(big.Int).Add(new(big.Int).Set(bound), nearestHalf), divisor)
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			if char < 'a' || char > 'f' {
				return false
			}
		}
	}
	return true
}
