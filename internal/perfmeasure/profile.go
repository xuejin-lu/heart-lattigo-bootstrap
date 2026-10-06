// Package perfmeasure provides the shared, backend-neutral workload and
// parameter construction used by matched Standard/Fast measurements.
package perfmeasure

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"os"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type Config struct {
	LogN             int   `json:"log_n"`
	LogDefaultScale  int   `json:"log_default_scale"`
	SecretHamming    int   `json:"secret_hamming"`
	Q0               []int `json:"q0"`
	QSlotsToCoeffs   []int `json:"q_slots_to_coeffs"`
	QCoeffsToSlots   []int `json:"q_coeffs_to_slots"`
	P                []int `json:"p"`
	SlotsToCoeffsDFT []int `json:"slots_to_coeffs_dft_levels"`
	CoeffsToSlotsDFT []int `json:"coeffs_to_slots_dft_levels"`
	LogSlots         int   `json:"log_slots"`
	Mod1LogScale     int   `json:"mod1_log_scale"`
	Mod1Degree       int   `json:"mod1_degree"`
	Mod1DoubleAngle  int   `json:"mod1_double_angle"`
	Mod1K            int   `json:"mod1_k"`
	LogMessageRatio  int   `json:"log_message_ratio"`
	Mod1InvDegree    int   `json:"mod1_inv_degree"`
	Repetitions      int   `json:"repetitions"`
	Warmup           int   `json:"warmup"`
}

type EffectiveParameters struct {
	LogN             int      `json:"log_n"`
	LogSlots         int      `json:"log_slots"`
	InputSlots       int      `json:"input_slots"`
	RingN            int      `json:"ring_n"`
	ResidualRingN    int      `json:"residual_ring_n"`
	Q0Target         int      `json:"q0_config_target_bits"`
	Q0Bits           int      `json:"q0_bits"`
	QChainBits       []int    `json:"q_chain_bits"`
	PBits            []int    `json:"p_bits"`
	QPrimes          []string `json:"q_primes"`
	PPrimes          []string `json:"p_primes"`
	DefaultScale     string   `json:"default_scale"`
	Mod1Scale        int      `json:"mod1_log_scale"`
	Mod1Degree       int      `json:"mod1_degree"`
	DoubleAngle      int      `json:"double_angle"`
	K                int      `json:"k"`
	LogMessageRatio  int      `json:"log_message_ratio"`
	QPrefixRowsAtMax int      `json:"q_prefix_rows_at_max_level"`
}

func LoadConfig(path string) (Config, [32]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, [32]byte{}, err
	}
	var cfg Config
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, [32]byte{}, err
	}
	if len(cfg.Q0) != 1 || len(cfg.P) == 0 || len(cfg.QSlotsToCoeffs) == 0 {
		return Config{}, [32]byte{}, errors.New("config must define q0, q_slots_to_coeffs, and P")
	}
	return cfg, sha256.Sum256(data), nil
}

func ParametersFromConfig(cfg Config) (ckks.Parameters, bootstrapping.Parameters, EffectiveParameters, error) {
	residualQ := []int{cfg.Q0[0], cfg.QSlotsToCoeffs[0]}
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: cfg.LogN, LogQ: residualQ, LogDefaultScale: cfg.LogDefaultScale,
		Xs: ring.Ternary{H: cfg.SecretHamming},
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, EffectiveParameters{}, fmt.Errorf("residual parameters: %w", err)
	}
	logSlots := cfg.LogSlots
	if logSlots < 0 {
		logSlots = residual.LogMaxSlots()
	}
	if logSlots < 0 || logSlots > residual.LogMaxSlots() {
		return ckks.Parameters{}, bootstrapping.Parameters{}, EffectiveParameters{}, fmt.Errorf("effective log_slots=%d outside [0,%d]", logSlots, residual.LogMaxSlots())
	}
	s2c, err := factorization(cfg.SlotsToCoeffsDFT, cfg.QSlotsToCoeffs)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, EffectiveParameters{}, fmt.Errorf("S2C factorization: %w", err)
	}
	c2s, err := factorization(cfg.CoeffsToSlotsDFT, cfg.QCoeffsToSlots)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, EffectiveParameters{}, fmt.Errorf("C2S factorization: %w", err)
	}
	logN, mod1Scale, degree, doubleAngle := cfg.LogN, cfg.Mod1LogScale, cfg.Mod1Degree, cfg.Mod1DoubleAngle
	k, ratio, invDegree, zero := cfg.Mod1K, cfg.LogMessageRatio, cfg.Mod1InvDegree, 0
	params, err := bootstrapping.NewParametersFromLiteral(residual, bootstrapping.ParametersLiteral{
		LogN: &logN, LogP: append([]int(nil), cfg.P...), Xs: ring.Ternary{H: cfg.SecretHamming}, LogSlots: &logSlots,
		CoeffsToSlotsFactorizationDepthAndLogScales: c2s, SlotsToCoeffsFactorizationDepthAndLogScales: s2c,
		EvalModLogScale: &mod1Scale, EphemeralSecretWeight: &zero, Mod1Type: mod1.CosDiscrete,
		LogMessageRatio: &ratio, K: &k, Mod1Degree: &degree, DoubleAngle: &doubleAngle, Mod1InvDegree: &invDegree,
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, EffectiveParameters{}, fmt.Errorf("bootstrapping parameters: %w", err)
	}
	params.CircuitOrder = bootstrapping.ModUpThenEncode
	params.ResidualParameters = residual
	q, p := params.BootstrappingParameters.Q(), params.BootstrappingParameters.P()
	defaultScale := residual.DefaultScale()
	effective := EffectiveParameters{
		LogN: params.BootstrappingParameters.LogN(), LogSlots: params.CoeffsToSlotsParameters.LogSlots,
		InputSlots: 1 << params.CoeffsToSlotsParameters.LogSlots, RingN: params.BootstrappingParameters.N(), ResidualRingN: residual.N(),
		Q0Target: cfg.Q0[0], Q0Bits: bits.Len64(q[0]), QChainBits: primeBits(q), PBits: primeBits(p), QPrimes: primeStrings(q), PPrimes: primeStrings(p),
		DefaultScale: defaultScale.Value.Text('e', 80), Mod1Scale: cfg.Mod1LogScale, Mod1Degree: cfg.Mod1Degree,
		DoubleAngle: cfg.Mod1DoubleAngle, K: cfg.Mod1K, LogMessageRatio: cfg.LogMessageRatio,
		QPrefixRowsAtMax: min(params.BootstrappingParameters.MaxLevel()+1, 4),
	}
	return residual, params, effective, nil
}

func DeterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for j := range values {
		values[j] = complex(float64(j%17-8)/256, float64((3*j)%13-6)/512)
	}
	return values
}

func Fingerprint(values []complex128) string {
	hash := sha256.New()
	var encoded [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(encoded[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(encoded[8:], math.Float64bits(imag(value)))
		_, _ = hash.Write(encoded[:])
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func EncodePlaintextLikeInput(params ckks.Parameters, logSlots int, values []complex128) (*rlwe.Ciphertext, error) {
	plain := ckks.NewPlaintext(params, 0)
	plain.IsNTT, plain.IsMontgomery = true, false
	plain.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err := ckks.NewEncoder(params).Encode(values, plain); err != nil {
		return nil, err
	}
	ct := ckks.NewCiphertext(params, 1, 0)
	*ct.MetaData = *plain.MetaData
	ct.Value[0].Copy(plain.Value)
	ct.Value[1].Zero()
	ct.IsNTT, ct.IsMontgomery = plain.IsNTT, plain.IsMontgomery
	return ct, nil
}

func factorization(depths, scales []int) ([][]int, error) {
	if len(depths) != len(scales) || len(depths) == 0 {
		return nil, errors.New("DFT factorization depths and scales must have equal non-zero lengths")
	}
	out := make([][]int, len(depths))
	for i := range depths {
		if depths[i] <= 0 || scales[i] <= 0 {
			return nil, fmt.Errorf("non-positive DFT factorization at group %d", i)
		}
		out[i] = make([]int, depths[i])
		for j := range out[i] {
			out[i][j] = scales[i]
		}
	}
	return out, nil
}

func primeBits(primes []uint64) []int {
	out := make([]int, len(primes))
	for i, prime := range primes {
		out[i] = bits.Len64(prime)
	}
	return out
}

func primeStrings(primes []uint64) []string {
	out := make([]string, len(primes))
	for i, prime := range primes {
		out[i] = new(big.Int).SetUint64(prime).String()
	}
	return out
}
