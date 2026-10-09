// Command fast-dropin-api-lifecycle-audit runs the same ordinary CKKS API
// lifecycle against whichever Lattigo module is selected by the build.
// It intentionally does not call Bootstrap.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/bits"
	"math/cmplx"
	"os"
	"strconv"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const canonicalConfigSHA256 = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"

type config struct {
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
}

type parameterEvidence struct {
	ResidualLogN          int     `json:"residual_log_n"`
	ResidualMaxLevel      int     `json:"residual_max_level"`
	ResidualScaleLog2     float64 `json:"residual_scale_log2"`
	BootstrapLogN         int     `json:"bootstrap_log_n"`
	BootstrapMaxLevel     int     `json:"bootstrap_max_level"`
	LogSlots              int     `json:"log_slots"`
	Slots                 int     `json:"slots"`
	QPrimeBits            []int   `json:"q_prime_bits"`
	PPrimeBits            []int   `json:"p_prime_bits"`
	QPrimesSHA256         string  `json:"q_primes_sha256"`
	PPrimesSHA256         string  `json:"p_primes_sha256"`
	EphemeralSecretWeight int     `json:"ephemeral_secret_weight"`
	CircuitOrder          string  `json:"circuit_order"`
}

type ciphertextEvidence struct {
	Level              int     `json:"level"`
	ScaleLog2          float64 `json:"scale_log2"`
	Degree             int     `json:"degree"`
	N                  int     `json:"n"`
	ComponentCount     int     `json:"component_count"`
	IsNTT              bool    `json:"is_ntt"`
	IsMontgomery       bool    `json:"is_montgomery"`
	C1NonzeroCount     int     `json:"c1_nonzero_count"`
	C1CoefficientCount int     `json:"c1_coefficient_count"`
}

type decryptionEvidence struct {
	Slots           int     `json:"slots"`
	Finite          bool    `json:"finite"`
	ComplexRMSE     float64 `json:"complex_rmse"`
	MaxComplexError float64 `json:"max_complex_error"`
}

type evidence struct {
	SchemaVersion                string             `json:"schema_version"`
	ConfigSHA256                 string             `json:"config_sha256"`
	InputSHA256                  string             `json:"input_sha256"`
	Parameters                   parameterEvidence  `json:"parameters"`
	SecretKeyLevelQ              int                `json:"secret_key_level_q"`
	SecretKeyLevelP              int                `json:"secret_key_level_p"`
	Ciphertext                   ciphertextEvidence `json:"ciphertext"`
	EvaluationKeysGenerated      bool               `json:"evaluation_keys_generated"`
	DenseToSparseKeyPresent      bool               `json:"dense_to_sparse_key_present"`
	SparseToDenseKeyPresent      bool               `json:"sparse_to_dense_key_present"`
	EvaluatorConstructed         bool               `json:"evaluator_constructed"`
	EmbeddedCKKSEvaluatorPresent bool               `json:"embedded_ckks_evaluator_present"`
	Decryption                   decryptionEvidence `json:"decryption"`
	BootstrapCalls               int                `json:"bootstrap_calls"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "canonical CKKS bootstrap configuration")
	flag.Parse()
	if *configPath == "" {
		return fmt.Errorf("-config is required")
	}

	configBytes, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	configSHA := hash(configBytes)
	if configSHA != canonicalConfigSHA256 {
		return fmt.Errorf("canonical LogN13 config SHA mismatch: %s", configSHA)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	residual, params, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return err
	}
	values := deterministicInput(1 << logSlots)
	inputSHA := inputHash(values)

	encoder := ckks.NewEncoder(residual)
	plaintext := ckks.NewPlaintext(residual, 0)
	plaintext.IsNTT = true
	plaintext.IsMontgomery = false
	plaintext.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err = encoder.Encode(values, plaintext); err != nil {
		return fmt.Errorf("encode original message: %w", err)
	}

	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	ct, err := rlwe.NewEncryptor(residual, sk).EncryptNew(plaintext)
	if err != nil {
		return fmt.Errorf("native EncryptNew: %w", err)
	}
	if ct.Degree() != 1 || len(ct.Value) != 2 {
		return fmt.Errorf("EncryptNew returned degree %d with %d components", ct.Degree(), len(ct.Value))
	}
	c1Nonzero, c1Count := nonzeroCount(ct.Value[1].Coeffs)
	if c1Nonzero == 0 {
		return fmt.Errorf("native EncryptNew produced a trivial c1")
	}

	keys, _, err := params.GenEvaluationKeys(sk)
	if err != nil {
		return fmt.Errorf("GenEvaluationKeys: %w", err)
	}
	if keys == nil {
		return fmt.Errorf("GenEvaluationKeys returned nil keys without an error")
	}
	eval, err := bootstrapping.NewEvaluator(params, keys)
	if err != nil {
		return fmt.Errorf("bootstrapping.NewEvaluator: %w", err)
	}
	if eval == nil {
		return fmt.Errorf("bootstrapping.NewEvaluator returned nil without an error")
	}

	decoded := make([]complex128, len(values))
	decodedPlaintext := rlwe.NewDecryptor(residual, sk).DecryptNew(ct)
	if err = encoder.Decode(decodedPlaintext, decoded); err != nil {
		return fmt.Errorf("native DecryptNew/decode: %w", err)
	}
	decryption := compare(values, decoded)
	if !decryption.Finite {
		return fmt.Errorf("native decryption produced non-finite values")
	}

	bootstrap := params.BootstrappingParameters
	qHash, qBits := primeEvidence(bootstrap.Q())
	pHash, pBits := primeEvidence(bootstrap.P())
	result := evidence{
		SchemaVersion: "fast-dropin-api-lifecycle-audit.v1",
		ConfigSHA256:  configSHA,
		InputSHA256:   inputSHA,
		Parameters: parameterEvidence{
			ResidualLogN: residual.LogN(), ResidualMaxLevel: residual.MaxLevel(),
			ResidualScaleLog2: residual.DefaultScale().Log2(), BootstrapLogN: bootstrap.LogN(),
			BootstrapMaxLevel: bootstrap.MaxLevel(), LogSlots: logSlots, Slots: len(values),
			QPrimeBits: qBits, PPrimeBits: pBits, QPrimesSHA256: qHash, PPrimesSHA256: pHash,
			EphemeralSecretWeight: params.EphemeralSecretWeight, CircuitOrder: "ModUpThenEncode",
		},
		SecretKeyLevelQ: sk.LevelQ(), SecretKeyLevelP: sk.LevelP(),
		Ciphertext: ciphertextEvidence{
			Level: ct.Level(), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(), N: ct.N(),
			ComponentCount: len(ct.Value), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
			C1NonzeroCount: c1Nonzero, C1CoefficientCount: c1Count,
		},
		EvaluationKeysGenerated:      keys != nil,
		DenseToSparseKeyPresent:      keys.EvkDenseToSparse != nil,
		SparseToDenseKeyPresent:      keys.EvkSparseToDense != nil,
		EvaluatorConstructed:         eval != nil,
		EmbeddedCKKSEvaluatorPresent: eval.Evaluator != nil,
		Decryption:                   decryption,
		BootstrapCalls:               0,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	fmt.Println(string(encoded))
	return nil
}

func parametersFromConfig(cfg config) (ckks.Parameters, bootstrapping.Parameters, int, error) {
	if cfg.LogN != 13 || cfg.LogDefaultScale != 45 || cfg.SecretHamming != 192 ||
		len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || len(cfg.QSlotsToCoeffs) != 3 ||
		len(cfg.QCoeffsToSlots) != 4 || len(cfg.P) != 5 || len(cfg.SlotsToCoeffsDFT) != 3 ||
		len(cfg.CoeffsToSlotsDFT) != 4 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 ||
		cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("config does not match the pinned canonical LogN13 profile")
	}
	for _, scale := range cfg.QSlotsToCoeffs {
		if scale != 39 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected SlotsToCoeffs prime scale %d", scale)
		}
	}
	for _, scale := range cfg.QCoeffsToSlots {
		if scale != 56 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected CoeffsToSlots prime scale %d", scale)
		}
	}
	for _, scale := range cfg.P {
		if scale != 61 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected P prime scale %d", scale)
		}
	}
	residualQ := []int{cfg.Q0[0], cfg.QSlotsToCoeffs[0]}
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: cfg.LogN, LogQ: residualQ, LogDefaultScale: cfg.LogDefaultScale,
		Xs: ring.Ternary{H: cfg.SecretHamming},
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("residual parameters: %w", err)
	}
	logSlots := cfg.LogSlots
	if logSlots < 0 {
		logSlots = residual.LogMaxSlots()
	}
	c2s, err := factorization(cfg.CoeffsToSlotsDFT, cfg.QCoeffsToSlots)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, err
	}
	s2c, err := factorization(cfg.SlotsToCoeffsDFT, cfg.QSlotsToCoeffs)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, err
	}
	logN, logScale, degree, doubleAngle, k := cfg.LogN, cfg.Mod1LogScale, cfg.Mod1Degree, cfg.Mod1DoubleAngle, cfg.Mod1K
	logMessageRatio, inverseDegree, zero := cfg.LogMessageRatio, cfg.Mod1InvDegree, 0
	params, err := bootstrapping.NewParametersFromLiteral(residual, bootstrapping.ParametersLiteral{
		LogN: &logN, LogP: append([]int(nil), cfg.P...), Xs: ring.Ternary{H: cfg.SecretHamming},
		LogSlots: &logSlots, CoeffsToSlotsFactorizationDepthAndLogScales: c2s,
		SlotsToCoeffsFactorizationDepthAndLogScales: s2c, EvalModLogScale: &logScale,
		EphemeralSecretWeight: &zero, Mod1Type: mod1.CosDiscrete, LogMessageRatio: &logMessageRatio,
		K: &k, Mod1Degree: &degree, DoubleAngle: &doubleAngle, Mod1InvDegree: &inverseDegree,
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("bootstrap parameters: %w", err)
	}
	params.CircuitOrder = bootstrapping.ModUpThenEncode
	params.ResidualParameters = residual
	params.EphemeralSecretWeight = 32
	return residual, params, logSlots, nil
}

func factorization(depths, scales []int) ([][]int, error) {
	if len(depths) == 0 || len(depths) != len(scales) {
		return nil, fmt.Errorf("DFT depths and prime scales must have the same non-zero length")
	}
	result := make([][]int, len(depths))
	for i, depth := range depths {
		if depth <= 0 || scales[i] <= 0 {
			return nil, fmt.Errorf("invalid DFT factorization at index %d", i)
		}
		result[i] = make([]int, depth)
		for j := range result[i] {
			result[i][j] = scales[i]
		}
	}
	return result, nil
}

func deterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for i := range values {
		values[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return values
}

func inputHash(values []complex128) string {
	hash := sha256.New()
	var encoded [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(encoded[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(encoded[8:], math.Float64bits(imag(value)))
		_, _ = hash.Write(encoded[:])
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func primeEvidence(primes []uint64) (string, []int) {
	strings := make([]string, len(primes))
	primeBits := make([]int, len(primes))
	for i, prime := range primes {
		strings[i] = strconv.FormatUint(prime, 10)
		primeBits[i] = bits.Len64(prime)
	}
	encoded, _ := json.Marshal(strings)
	return hash(encoded), primeBits
}

func nonzeroCount(rows [][]uint64) (count, total int) {
	for _, row := range rows {
		for _, coefficient := range row {
			total++
			if coefficient != 0 {
				count++
			}
		}
	}
	return
}

func compare(expected, actual []complex128) decryptionEvidence {
	result := decryptionEvidence{Slots: len(actual), Finite: len(expected) == len(actual)}
	var squaredError float64
	for i, value := range actual {
		if math.IsNaN(real(value)) || math.IsInf(real(value), 0) || math.IsNaN(imag(value)) || math.IsInf(imag(value), 0) {
			result.Finite = false
		}
		difference := cmplx.Abs(value - expected[i])
		squaredError += difference * difference
		if difference > result.MaxComplexError {
			result.MaxComplexError = difference
		}
	}
	if result.Slots != 0 {
		result.ComplexRMSE = math.Sqrt(squaredError / float64(result.Slots))
	}
	return result
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
