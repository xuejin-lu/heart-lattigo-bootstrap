// Command fast-dropin-zero-secret-encrypt-bridge-004 runs one unchanged CKKS
// frontend against the dependency selected by the build workspace. Its
// preflight phase never calls Bootstrap; the bootstrap phase calls it once.
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
	"runtime"
	"strconv"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	canonicalConfigSHA256  = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"
	preflightRMSETolerance = 1e-7
)

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

type complexValue struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type metric struct {
	Finite          bool     `json:"finite"`
	ComplexRMSE     float64  `json:"complex_rmse"`
	MaxComplexError float64  `json:"max_complex_error"`
	MaxRealError    float64  `json:"max_real_error"`
	MaxImagError    float64  `json:"max_imag_error"`
	SNRDB           *float64 `json:"snr_db,omitempty"`
	SNRState        string   `json:"snr_state"`
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
	C1Zero             bool    `json:"c1_zero"`
	C1NonzeroCount     int     `json:"c1_nonzero_count"`
	C1CoefficientCount int     `json:"c1_coefficient_count"`
}

type runEvidence struct {
	SchemaVersion           string              `json:"schema_version"`
	Phase                   string              `json:"phase"`
	Status                  string              `json:"status"`
	Repetitions             int                 `json:"repetitions"`
	WarmupRuns              int                 `json:"warmup_runs"`
	TimestampUTC            string              `json:"timestamp_utc"`
	GoVersion               string              `json:"go_version"`
	OS                      string              `json:"os"`
	Architecture            string              `json:"architecture"`
	PrimaryCommit           string              `json:"primary_commit"`
	PrimaryDirty            bool                `json:"primary_dirty"`
	PrimarySourceSHA256     string              `json:"primary_source_sha256"`
	BackendCommit           string              `json:"backend_commit"`
	BackendRef              string              `json:"backend_ref"`
	BackendDirty            bool                `json:"backend_dirty"`
	ConfigSHA256            string              `json:"config_sha256"`
	InputSHA256             string              `json:"input_sha256"`
	DecodedImmediateSHA256  string              `json:"decoded_immediate_sha256,omitempty"`
	DecodedBootstrapSHA256  string              `json:"decoded_bootstrap_sha256,omitempty"`
	Parameters              parameterEvidence   `json:"parameters"`
	SecretKeyLevelQ         int                 `json:"secret_key_level_q"`
	SecretKeyLevelP         int                 `json:"secret_key_level_p"`
	EvaluationKeysGenerated bool                `json:"evaluation_keys_generated"`
	DenseToSparseKeyPresent bool                `json:"dense_to_sparse_key_present"`
	SparseToDenseKeyPresent bool                `json:"sparse_to_dense_key_present"`
	EvaluatorConstructed    bool                `json:"evaluator_constructed"`
	FastEvaluatorSelected   bool                `json:"fast_evaluator_selected"`
	InputCiphertext         ciphertextEvidence  `json:"input_ciphertext"`
	ImmediateDecryption     metric              `json:"immediate_decryption"`
	BootstrapCalls          int                 `json:"bootstrap_calls"`
	BootstrapOutput         *ciphertextEvidence `json:"bootstrap_output,omitempty"`
	BootstrapVsOriginal     *metric             `json:"bootstrap_vs_original,omitempty"`
	BootstrapError          string              `json:"bootstrap_error,omitempty"`
	InputValues             []complexValue      `json:"input_values,omitempty"`
	ImmediateDecodedValues  []complexValue      `json:"immediate_decoded_values,omitempty"`
	BootstrapDecodedValues  []complexValue      `json:"bootstrap_decoded_values,omitempty"`
}

type comparisonEvidence struct {
	SchemaVersion       string `json:"schema_version"`
	TimestampUTC        string `json:"timestamp_utc"`
	PrimaryCommit       string `json:"primary_commit"`
	PrimarySourceSHA256 string `json:"primary_source_sha256"`
	StandardCommit      string `json:"standard_commit"`
	FastCommit          string `json:"fast_commit"`
	ConfigSHA256        string `json:"config_sha256"`
	InputSHA256         string `json:"input_sha256"`
	BootstrapCalls      struct {
		Standard int `json:"standard"`
		Fast     int `json:"fast"`
	} `json:"bootstrap_calls"`
	FastVsStandard metric `json:"fast_vs_standard"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	phase := flag.String("phase", "preflight", "preflight, bootstrap, or compare")
	configPath := flag.String("config", "", "canonical LogN13 bootstrap config")
	outPath := flag.String("out", "", "local evidence output path")
	primaryCommit := flag.String("primary-commit", "", "Primary commit used for this run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether the Primary source tree was dirty")
	backendCommit := flag.String("backend-commit", "", "Lattigo implementation commit")
	backendRef := flag.String("backend-ref", "", "Lattigo branch or pinned ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether the Lattigo source tree was dirty")
	standardResult := flag.String("standard-result", "", "Standard bootstrap evidence for compare phase")
	fastResult := flag.String("fast-result", "", "Fast bootstrap evidence for compare phase")
	flag.Parse()

	if *outPath == "" {
		return fmt.Errorf("-out is required")
	}
	if *phase == "compare" {
		if *standardResult == "" || *fastResult == "" {
			return fmt.Errorf("compare phase requires -standard-result and -fast-result")
		}
		result, err := compareRunFiles(*standardResult, *fastResult)
		if err != nil {
			return err
		}
		return writeJSON(*outPath, result)
	}
	if *phase != "preflight" && *phase != "bootstrap" {
		return fmt.Errorf("unsupported phase %q", *phase)
	}
	if *configPath == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return fmt.Errorf("run phase requires -config, -primary-commit, -backend-commit, and -backend-ref")
	}

	result, runErr := runBackend(*phase, *configPath, *primaryCommit, *primaryDirty, *backendCommit, *backendRef, *backendDirty)
	if err := writeJSON(*outPath, result); err != nil {
		return err
	}
	if runErr != nil {
		return runErr
	}
	return nil
}

func runBackend(phase, configPath, primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty bool) (runEvidence, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return runEvidence{}, fmt.Errorf("read config: %w", err)
	}
	configSHA := hashBytes(configBytes)
	if configSHA != canonicalConfigSHA256 {
		return runEvidence{}, fmt.Errorf("canonical LogN13 config SHA mismatch: %s", configSHA)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return runEvidence{}, fmt.Errorf("decode config: %w", err)
	}
	residual, btpParams, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return runEvidence{}, err
	}
	values := deterministicInput(1 << logSlots)
	inputSHA := hashComplexValues(values)
	sourceBytes, err := os.ReadFile("tools/fast-dropin-zero-secret-encrypt-bridge-004/main.go")
	if err != nil {
		return runEvidence{}, fmt.Errorf("read shared frontend source: %w", err)
	}

	encoder := ckks.NewEncoder(residual)
	plaintext := ckks.NewPlaintext(residual, 0)
	plaintext.IsNTT = true
	plaintext.IsMontgomery = false
	plaintext.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err = encoder.Encode(values, plaintext); err != nil {
		return runEvidence{}, fmt.Errorf("encode original message: %w", err)
	}

	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	ct, err := rlwe.NewEncryptor(residual, sk).EncryptNew(plaintext)
	if err != nil {
		return runEvidence{}, fmt.Errorf("ordinary secret-key EncryptNew: %w", err)
	}
	if ct.Degree() != 1 || len(ct.Value) != 2 {
		return runEvidence{}, fmt.Errorf("EncryptNew returned degree %d with %d components", ct.Degree(), len(ct.Value))
	}
	c1Nonzero, c1Count := nonzeroCount(ct.Value[1].Coeffs)
	keys, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return runEvidence{}, fmt.Errorf("GenEvaluationKeys: %w", err)
	}
	if keys == nil {
		return runEvidence{}, fmt.Errorf("GenEvaluationKeys returned nil keys without an error")
	}
	eval, err := bootstrapping.NewEvaluator(btpParams, keys)
	if err != nil {
		return runEvidence{}, fmt.Errorf("public bootstrapping.NewEvaluator: %w", err)
	}
	if eval == nil {
		return runEvidence{}, fmt.Errorf("NewEvaluator returned nil without an error")
	}

	decodedImmediate := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(ct), decodedImmediate); err != nil {
		return runEvidence{}, fmt.Errorf("immediate public DecryptNew/Decode: %w", err)
	}
	immediateMetric := compareValues(values, decodedImmediate)
	fastSelected := eval.Evaluator == nil
	result := runEvidence{
		SchemaVersion:           "fast-dropin-zero-secret-encrypt-bridge-004.v1",
		Phase:                   phase,
		Status:                  "preflight_passed",
		Repetitions:             1,
		WarmupRuns:              0,
		TimestampUTC:            time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion:               runtime.Version(),
		OS:                      runtime.GOOS,
		Architecture:            runtime.GOARCH,
		PrimaryCommit:           primaryCommit,
		PrimaryDirty:            primaryDirty,
		PrimarySourceSHA256:     hashBytes(sourceBytes),
		BackendCommit:           backendCommit,
		BackendRef:              backendRef,
		BackendDirty:            backendDirty,
		ConfigSHA256:            configSHA,
		InputSHA256:             inputSHA,
		DecodedImmediateSHA256:  hashComplexValues(decodedImmediate),
		Parameters:              parameterSummary(residual, btpParams, logSlots, len(values)),
		SecretKeyLevelQ:         sk.LevelQ(),
		SecretKeyLevelP:         sk.LevelP(),
		EvaluationKeysGenerated: true,
		DenseToSparseKeyPresent: keys.EvkDenseToSparse != nil,
		SparseToDenseKeyPresent: keys.EvkSparseToDense != nil,
		EvaluatorConstructed:    true,
		FastEvaluatorSelected:   fastSelected,
		InputCiphertext:         ciphertextSummary(ct, c1Nonzero, c1Count),
		ImmediateDecryption:     immediateMetric,
		InputValues:             toComplexValues(values),
		ImmediateDecodedValues:  toComplexValues(decodedImmediate),
	}
	if !immediateMetric.Finite || immediateMetric.ComplexRMSE > preflightRMSETolerance {
		result.Status = "preflight_failed"
		result.BootstrapError = fmt.Sprintf("immediate decrypt/decode RMSE %.12g exceeds preflight tolerance %.12g", immediateMetric.ComplexRMSE, preflightRMSETolerance)
		return result, fmt.Errorf("%s", result.BootstrapError)
	}
	if fastSelected && c1Nonzero != 0 {
		result.Status = "preflight_failed"
		result.BootstrapError = fmt.Sprintf("Fast evaluator selected but EncryptNew returned %d non-zero c1 coefficients", c1Nonzero)
		return result, fmt.Errorf("%s", result.BootstrapError)
	}
	if phase == "preflight" {
		return result, nil
	}

	result.BootstrapCalls = 1
	output, err := eval.Bootstrap(ct)
	if err != nil {
		result.Status = "bootstrap_failed"
		result.BootstrapError = err.Error()
		return result, fmt.Errorf("single public Bootstrap call: %w", err)
	}
	outC1Nonzero, outC1Count := nonzeroCount(output.Value[1].Coeffs)
	outputSummary := ciphertextSummary(output, outC1Nonzero, outC1Count)
	result.BootstrapOutput = &outputSummary
	if fastSelected && outC1Nonzero != 0 {
		result.Status = "bootstrap_output_c1_nonzero"
		result.BootstrapError = fmt.Sprintf("Fast Bootstrap output has %d non-zero c1 coefficients; ordinary DecryptNew semantics are not accepted as a zero-secret result", outC1Nonzero)
		return result, fmt.Errorf("%s", result.BootstrapError)
	}

	decodedBootstrap := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(output), decodedBootstrap); err != nil {
		result.Status = "bootstrap_decode_failed"
		result.BootstrapError = err.Error()
		return result, fmt.Errorf("public DecryptNew/Decode after Bootstrap: %w", err)
	}
	bootstrapMetric := compareValues(values, decodedBootstrap)
	result.Status = "bootstrap_completed"
	result.DecodedBootstrapSHA256 = hashComplexValues(decodedBootstrap)
	result.BootstrapVsOriginal = &bootstrapMetric
	result.BootstrapDecodedValues = toComplexValues(decodedBootstrap)
	return result, nil
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

func parameterSummary(residual ckks.Parameters, btp bootstrapping.Parameters, logSlots, slots int) parameterEvidence {
	bootstrap := btp.BootstrappingParameters
	qHash, qBits := primeEvidence(bootstrap.Q())
	pHash, pBits := primeEvidence(bootstrap.P())
	return parameterEvidence{
		ResidualLogN: residual.LogN(), ResidualMaxLevel: residual.MaxLevel(),
		ResidualScaleLog2: residual.DefaultScale().Log2(), BootstrapLogN: bootstrap.LogN(),
		BootstrapMaxLevel: bootstrap.MaxLevel(), LogSlots: logSlots, Slots: slots,
		QPrimeBits: qBits, PPrimeBits: pBits, QPrimesSHA256: qHash, PPrimesSHA256: pHash,
		EphemeralSecretWeight: btp.EphemeralSecretWeight, CircuitOrder: "ModUpThenEncode",
	}
}

func ciphertextSummary(ct *rlwe.Ciphertext, c1Nonzero, c1Count int) ciphertextEvidence {
	return ciphertextEvidence{
		Level: ct.Level(), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(), N: ct.N(),
		ComponentCount: len(ct.Value), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		C1Zero: c1Nonzero == 0, C1NonzeroCount: c1Nonzero, C1CoefficientCount: c1Count,
	}
}

func deterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for i := range values {
		values[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return values
}

func toComplexValues(values []complex128) []complexValue {
	result := make([]complexValue, len(values))
	for i, value := range values {
		result[i] = complexValue{Real: real(value), Imag: imag(value)}
	}
	return result
}

func fromComplexValues(values []complexValue) []complex128 {
	result := make([]complex128, len(values))
	for i, value := range values {
		result[i] = complex(value.Real, value.Imag)
	}
	return result
}

func compareValues(reference, actual []complex128) metric {
	result := metric{Finite: len(reference) == len(actual), SNRState: "finite"}
	var errorPower, signalPower float64
	count := len(reference)
	if len(actual) < count {
		count = len(actual)
	}
	for i := 0; i < count; i++ {
		if !finite(real(reference[i])) || !finite(imag(reference[i])) || !finite(real(actual[i])) || !finite(imag(actual[i])) {
			result.Finite = false
		}
		difference := actual[i] - reference[i]
		magnitude := cmplx.Abs(difference)
		errorPower += magnitude * magnitude
		signalPower += real(reference[i])*real(reference[i]) + imag(reference[i])*imag(reference[i])
		result.MaxRealError = math.Max(result.MaxRealError, math.Abs(real(difference)))
		result.MaxImagError = math.Max(result.MaxImagError, math.Abs(imag(difference)))
		result.MaxComplexError = math.Max(result.MaxComplexError, magnitude)
	}
	if count > 0 {
		result.ComplexRMSE = math.Sqrt(errorPower / float64(count))
	}
	if !result.Finite || !finite(errorPower) || !finite(signalPower) {
		result.Finite = false
		result.SNRState = "non_finite_values_or_aggregate"
		return result
	}
	switch {
	case errorPower == 0 && signalPower > 0:
		result.SNRState = "positive_infinity_zero_error"
	case errorPower == 0 && signalPower == 0:
		result.SNRState = "undefined_zero_signal_and_error"
	case signalPower == 0:
		result.SNRState = "negative_infinity_zero_signal"
	default:
		snr := 10 * math.Log10(signalPower/errorPower)
		result.SNRDB = &snr
	}
	return result
}

func compareRunFiles(standardPath, fastPath string) (comparisonEvidence, error) {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return comparisonEvidence{}, fmt.Errorf("read Standard evidence: %w", err)
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return comparisonEvidence{}, fmt.Errorf("read Fast evidence: %w", err)
	}
	if standard.Phase != "bootstrap" || fast.Phase != "bootstrap" || standard.Status != "bootstrap_completed" || fast.Status != "bootstrap_completed" {
		return comparisonEvidence{}, fmt.Errorf("both inputs must be completed bootstrap-phase evidence")
	}
	if standard.ConfigSHA256 != fast.ConfigSHA256 || standard.InputSHA256 != fast.InputSHA256 {
		return comparisonEvidence{}, fmt.Errorf("Standard and Fast provenance does not contain the same config and input hashes")
	}
	if standard.FastEvaluatorSelected || !fast.FastEvaluatorSelected {
		return comparisonEvidence{}, fmt.Errorf("public NewEvaluator did not select the expected Standard/Fast paths")
	}
	if standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 {
		return comparisonEvidence{}, fmt.Errorf("expected exactly one Bootstrap invocation per backend, got Standard=%d Fast=%d", standard.BootstrapCalls, fast.BootstrapCalls)
	}
	if fast.BootstrapOutput == nil || !fast.BootstrapOutput.C1Zero {
		return comparisonEvidence{}, fmt.Errorf("Fast output is not a verified zero-c1 ciphertext")
	}
	standardValues := fromComplexValues(standard.BootstrapDecodedValues)
	fastValues := fromComplexValues(fast.BootstrapDecodedValues)
	if len(standardValues) != len(fastValues) || len(standardValues) == 0 {
		return comparisonEvidence{}, fmt.Errorf("Standard and Fast decoded output lengths differ or are empty")
	}
	result := comparisonEvidence{
		SchemaVersion: "fast-dropin-zero-secret-encrypt-bridge-004.compare.v1",
		TimestampUTC:  time.Now().UTC().Format(time.RFC3339Nano),
		PrimaryCommit: fast.PrimaryCommit, PrimarySourceSHA256: fast.PrimarySourceSHA256,
		StandardCommit: standard.BackendCommit, FastCommit: fast.BackendCommit,
		ConfigSHA256: standard.ConfigSHA256, InputSHA256: standard.InputSHA256,
		FastVsStandard: compareValues(standardValues, fastValues),
	}
	result.BootstrapCalls.Standard = standard.BootstrapCalls
	result.BootstrapCalls.Fast = fast.BootstrapCalls
	return result, nil
}

func hashComplexValues(values []complex128) string {
	h := sha256.New()
	var encoded [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(encoded[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(encoded[8:], math.Float64bits(imag(value)))
		_, _ = h.Write(encoded[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func primeEvidence(primes []uint64) (string, []int) {
	strings := make([]string, len(primes))
	primeBits := make([]int, len(primes))
	for i, prime := range primes {
		strings[i] = strconv.FormatUint(prime, 10)
		primeBits[i] = bits.Len64(prime)
	}
	encoded, _ := json.Marshal(strings)
	return hashBytes(encoded), primeBits
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

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func writeJSON(path string, value interface{}) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode local evidence: %w", err)
	}
	encoded = append(encoded, '\n')
	if err = os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("write local evidence %q: %w", path, err)
	}
	return nil
}

func readJSON(path string, value interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
