// Command fast-dropin-public-rotate-bootstrap-composition-010 runs the same
// public CKKS frontend against the dependency selected by GOWORK. Preflight
// mode never invokes Bootstrap; bootstrap mode invokes it at most once.
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
	"reflect"
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
	rotation               = 1
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
	Repetitions      int   `json:"repetitions"`
	Warmup           int   `json:"warmup"`
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

type ciphertextEvidence struct {
	Level              int      `json:"level"`
	ScaleLog2          float64  `json:"scale_log2"`
	Degree             int      `json:"degree"`
	N                  int      `json:"n"`
	ComponentCount     int      `json:"component_count"`
	RowsPerComponent   []int    `json:"rows_per_component"`
	ActiveRows         int      `json:"active_rows"`
	FullActiveQ        bool     `json:"full_active_q"`
	IsNTT              bool     `json:"is_ntt"`
	IsMontgomery       bool     `json:"is_montgomery"`
	IsBatched          bool     `json:"is_batched"`
	IsBitReversed      bool     `json:"is_bit_reversed"`
	LogDimensionsRows  int      `json:"log_dimensions_rows"`
	LogDimensionsCols  int      `json:"log_dimensions_cols"`
	LogSlots           int      `json:"log_slots"`
	C1Zero             bool     `json:"c1_zero"`
	C1NonzeroCount     int      `json:"c1_nonzero_count"`
	C1CoefficientCount int      `json:"c1_coefficient_count"`
	C0RowSHA256        []string `json:"c0_row_sha256"`
	C1RowSHA256        []string `json:"c1_row_sha256"`
}

type parameterEvidence struct {
	ResidualLogN          int     `json:"residual_log_n"`
	ResidualMaxLevel      int     `json:"residual_max_level"`
	ResidualScaleLog2     float64 `json:"residual_scale_log2"`
	BootstrapLogN         int     `json:"bootstrap_log_n"`
	BootstrapMaxLevel     int     `json:"bootstrap_max_level"`
	LogSlots              int     `json:"log_slots"`
	Slots                 int     `json:"slots"`
	ResidualQBits         []int   `json:"residual_q_prime_bits"`
	BootstrapQBits        []int   `json:"bootstrap_q_prime_bits"`
	BootstrapPBits        []int   `json:"bootstrap_p_prime_bits"`
	BootstrapQSHA256      string  `json:"bootstrap_q_primes_sha256"`
	BootstrapPSHA256      string  `json:"bootstrap_p_primes_sha256"`
	EphemeralSecretWeight int     `json:"ephemeral_secret_weight"`
	CircuitOrder          string  `json:"circuit_order"`
}

type keyLayoutEvidence struct {
	GeneratedBy             string `json:"generated_by"`
	GaloisKeyCount          int    `json:"galois_key_count"`
	StandardLayoutKeyCount  int    `json:"standard_layout_key_count"`
	FastLayoutKeyCount      int    `json:"fast_layout_key_count"`
	RotationInBootstrapPlan bool   `json:"rotation_in_bootstrap_key_plan"`
	RelinearizationPresent  bool   `json:"relinearization_key_present"`
}

type runEvidence struct {
	SchemaVersion                 string              `json:"schema_version"`
	Phase                         string              `json:"phase"`
	Status                        string              `json:"status"`
	TimestampUTC                  string              `json:"timestamp_utc"`
	GoVersion                     string              `json:"go_version"`
	OS                            string              `json:"os"`
	Architecture                  string              `json:"architecture"`
	PrimaryCommit                 string              `json:"primary_commit"`
	PrimaryDirty                  bool                `json:"primary_dirty"`
	PrimarySourceSHA256           string              `json:"primary_source_sha256"`
	BackendCommit                 string              `json:"backend_commit"`
	BackendRef                    string              `json:"backend_ref"`
	BackendDirty                  bool                `json:"backend_dirty"`
	ConfigSHA256                  string              `json:"config_sha256"`
	InputSHA256                   string              `json:"input_sha256"`
	DecodedInputSHA256            string              `json:"decoded_input_sha256,omitempty"`
	DecodedRotateSHA256           string              `json:"decoded_rotate_sha256,omitempty"`
	DecodedBootstrapSHA256        string              `json:"decoded_bootstrap_sha256,omitempty"`
	Parameters                    parameterEvidence   `json:"parameters"`
	RotateK                       int                 `json:"rotate_k"`
	RotateGaloisElement           uint64              `json:"rotate_galois_element"`
	BootstrapGaloisElement        uint64              `json:"bootstrap_galois_element"`
	RotateElementMatchesBootstrap bool                `json:"rotate_element_matches_bootstrap_parameters"`
	EvaluationKeys                keyLayoutEvidence   `json:"evaluation_keys"`
	FastBootstrapSelected         bool                `json:"fast_bootstrap_selected"`
	RotateAPI                     string              `json:"rotate_api"`
	RotateGetGaloisKeyCalls       int                 `json:"rotate_get_galois_key_calls"`
	RotateGetGaloisKeyListCalls   int                 `json:"rotate_get_galois_key_list_calls"`
	InputCiphertext               *ciphertextEvidence `json:"input_ciphertext,omitempty"`
	InputUnchangedByRotate        bool                `json:"input_unchanged_by_rotate"`
	RotatedCiphertext             *ciphertextEvidence `json:"rotated_ciphertext,omitempty"`
	ImmediateDecryption           *metric             `json:"immediate_decryption,omitempty"`
	RotateVsOracle                *metric             `json:"rotate_vs_rotated_oracle,omitempty"`
	BootstrapCalls                int                 `json:"bootstrap_calls"`
	BootstrapInputRotated         bool                `json:"bootstrap_input_is_rotated_ciphertext"`
	BootstrapOutput               *ciphertextEvidence `json:"bootstrap_output,omitempty"`
	BootstrapVsOracle             *metric             `json:"bootstrap_vs_rotated_oracle,omitempty"`
	Error                         string              `json:"error,omitempty"`
	InputValues                   []complexValue      `json:"input_values,omitempty"`
	RotatedDecodedValues          []complexValue      `json:"rotated_decoded_values,omitempty"`
	BootstrapDecodedValues        []complexValue      `json:"bootstrap_decoded_values,omitempty"`
}

type backendSummary struct {
	BackendCommit           string              `json:"backend_commit"`
	BackendRef              string              `json:"backend_ref"`
	BackendDirty            bool                `json:"backend_dirty"`
	FastBootstrapSelected   bool                `json:"fast_bootstrap_selected"`
	EvaluationKeys          keyLayoutEvidence   `json:"evaluation_keys"`
	RotateGetGaloisKeyCalls int                 `json:"rotate_get_galois_key_calls"`
	Parameters              parameterEvidence   `json:"parameters"`
	InputCiphertext         *ciphertextEvidence `json:"input_ciphertext"`
	RotatedCiphertext       *ciphertextEvidence `json:"rotated_ciphertext"`
	BootstrapOutput         *ciphertextEvidence `json:"bootstrap_output"`
	ImmediateDecryption     *metric             `json:"immediate_decryption"`
	RotateVsOracle          *metric             `json:"rotate_vs_rotated_oracle"`
	BootstrapVsOracle       *metric             `json:"bootstrap_vs_rotated_oracle"`
	DecodedInputSHA256      string              `json:"decoded_input_sha256"`
	DecodedRotateSHA256     string              `json:"decoded_rotate_sha256"`
	DecodedBootstrapSHA256  string              `json:"decoded_bootstrap_sha256"`
}

type comparisonEvidence struct {
	SchemaVersion           string         `json:"schema_version"`
	TimestampUTC            string         `json:"timestamp_utc"`
	PrimaryCommit           string         `json:"primary_commit"`
	PrimarySourceSHA256     string         `json:"primary_source_sha256"`
	ConfigSHA256            string         `json:"config_sha256"`
	InputSHA256             string         `json:"input_sha256"`
	RotateK                 int            `json:"rotate_k"`
	RotateGaloisElement     uint64         `json:"rotate_galois_element"`
	BootstrapCallsStandard  int            `json:"bootstrap_calls_standard"`
	BootstrapCallsFast      int            `json:"bootstrap_calls_fast"`
	Standard                backendSummary `json:"standard"`
	Fast                    backendSummary `json:"fast"`
	RotatedFastVsStandard   metric         `json:"rotated_fast_vs_standard"`
	BootstrapFastVsStandard metric         `json:"bootstrap_fast_vs_standard"`
}

type trackingEvaluationKeySet struct {
	rlwe.EvaluationKeySet
	galoisKeyCalls     int
	galoisKeyListCalls int
}

func (keys *trackingEvaluationKeySet) GetGaloisKey(galEl uint64) (*rlwe.GaloisKey, error) {
	keys.galoisKeyCalls++
	return keys.EvaluationKeySet.GetGaloisKey(galEl)
}

func (keys *trackingEvaluationKeySet) GetGaloisKeysList() []uint64 {
	keys.galoisKeyListCalls++
	return keys.EvaluationKeySet.GetGaloisKeysList()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	phase := flag.String("phase", "preflight", "preflight, bootstrap, or compare")
	configPath := flag.String("config", "", "canonical LogN13 E32 bootstrap config")
	outPath := flag.String("out", "", "evidence output path")
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
	return runErr
}

func runBackend(phase, configPath, primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty bool) (result runEvidence, runErr error) {
	result = runEvidence{
		SchemaVersion: "fast-dropin-public-rotate-bootstrap-composition-010.v1",
		Phase:         phase, Status: "preflight_started", TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		PrimaryCommit: primaryCommit, PrimaryDirty: primaryDirty,
		BackendCommit: backendCommit, BackendRef: backendRef, BackendDirty: backendDirty,
		RotateK: rotation, RotateAPI: "ckks.NewEvaluator(...).RotateNew(ciphertext, k)",
	}
	defer func() {
		if runErr != nil {
			if phase == "bootstrap" && result.BootstrapCalls == 1 {
				result.Status = "bootstrap_failed"
			} else {
				result.Status = "preflight_failed"
			}
			result.Error = runErr.Error()
		}
	}()

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return result, fmt.Errorf("read config: %w", err)
	}
	result.ConfigSHA256 = hashBytes(configBytes)
	if result.ConfigSHA256 != canonicalConfigSHA256 {
		return result, fmt.Errorf("canonical LogN13 config SHA mismatch: %s", result.ConfigSHA256)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return result, fmt.Errorf("decode config: %w", err)
	}
	residual, btpParams, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return result, err
	}
	values := deterministicInput(1 << logSlots)
	expected := rotateLeftOracle(values, rotation)
	result.InputSHA256 = hashComplexValues(values)
	sourceBytes, err := os.ReadFile("tools/fast-dropin-public-rotate-bootstrap-composition-010/main.go")
	if err != nil {
		return result, fmt.Errorf("read shared frontend source: %w", err)
	}
	result.PrimarySourceSHA256 = hashBytes(sourceBytes)
	result.Parameters = parameterSummary(residual, btpParams, logSlots, len(values))

	rotateGalEl := residual.GaloisElement(rotation)
	bootstrapGalEl := btpParams.BootstrappingParameters.GaloisElement(rotation)
	result.RotateGaloisElement = rotateGalEl
	result.BootstrapGaloisElement = bootstrapGalEl
	result.RotateElementMatchesBootstrap = rotateGalEl == bootstrapGalEl
	result.EvaluationKeys = keyLayoutEvidence{
		GeneratedBy:             "bootstrapping.Parameters.GenEvaluationKeys(sk)",
		RotationInBootstrapPlan: containsGaloisElement(btpParams.GaloisElements(btpParams.BootstrappingParameters), bootstrapGalEl),
	}
	if residual.MaxLevel() != 1 || residual.LogN() != 13 || logSlots != residual.LogMaxSlots() {
		return result, fmt.Errorf("canonical profile mismatch: residual LogN=%d MaxLevel=%d LogSlots=%d", residual.LogN(), residual.MaxLevel(), logSlots)
	}
	if !result.RotateElementMatchesBootstrap {
		return result, fmt.Errorf("public Rotate and Bootstrap key-plan Galois elements differ: rotate=%d bootstrap=%d", rotateGalEl, bootstrapGalEl)
	}
	if !result.EvaluationKeys.RotationInBootstrapPlan {
		return result, fmt.Errorf("rotation Galois element %d is absent from the canonical Bootstrap evaluation-key plan", bootstrapGalEl)
	}

	encoder := ckks.NewEncoder(residual)
	plaintext := ckks.NewPlaintext(residual, 0)
	plaintext.Scale = rlwe.NewScale(math.Exp2(float64(cfg.LogDefaultScale)))
	plaintext.IsNTT = true
	plaintext.IsMontgomery = false
	plaintext.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err = encoder.Encode(values, plaintext); err != nil {
		return result, fmt.Errorf("public CKKS Encode: %w", err)
	}
	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	ct, err := rlwe.NewEncryptor(residual, sk).EncryptNew(plaintext)
	if err != nil {
		return result, fmt.Errorf("public secret-key EncryptNew: %w", err)
	}
	if err = validateBootstrapBoundaryCiphertext("EncryptNew", ct, residual); err != nil {
		return result, err
	}
	if ct.Level() != 0 || ct.Scale.Log2() != float64(cfg.LogDefaultScale) || !ct.IsNTT || ct.IsMontgomery {
		return result, fmt.Errorf("EncryptNew preflight contract mismatch: level=%d log2Scale=%.12g NTT=%t Montgomery=%t", ct.Level(), ct.Scale.Log2(), ct.IsNTT, ct.IsMontgomery)
	}
	inputC1Nonzero, inputC1Count := nonzeroCount(ct.Value[1].Coeffs)
	if inputC1Count != residual.N() {
		return result, fmt.Errorf("expected exactly one active q0 row in c1, got %d coefficients", inputC1Count)
	}
	result.InputCiphertext = ciphertextSummary(ct, inputC1Nonzero, inputC1Count)
	result.InputValues = toComplexValues(values)
	decodedInput := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(ct), decodedInput); err != nil {
		return result, fmt.Errorf("public DecryptNew/Decode before Rotate: %w", err)
	}
	inputMetric := compareValues(values, decodedInput)
	result.ImmediateDecryption = &inputMetric
	result.DecodedInputSHA256 = hashComplexValues(decodedInput)
	if !inputMetric.Finite || inputMetric.ComplexRMSE > preflightRMSETolerance {
		return result, fmt.Errorf("EncryptNew immediate-decrypt RMSE %.12g exceeds prior 004 preflight tolerance %.12g", inputMetric.ComplexRMSE, preflightRMSETolerance)
	}

	keys, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return result, fmt.Errorf("public Bootstrap GenEvaluationKeys: %w", err)
	}
	if keys == nil || keys.MemEvaluationKeySet == nil {
		return result, fmt.Errorf("GenEvaluationKeys returned nil or incomplete evaluation keys")
	}
	result.EvaluationKeys.GaloisKeyCount = len(keys.GaloisKeys)
	result.EvaluationKeys.StandardLayoutKeyCount, result.EvaluationKeys.FastLayoutKeyCount = countKeyLayouts(keys)
	result.EvaluationKeys.RelinearizationPresent = keys.RelinearizationKey != nil
	bootstrapEval, err := bootstrapping.NewEvaluator(btpParams, keys)
	if err != nil {
		return result, fmt.Errorf("public bootstrapping.NewEvaluator: %w", err)
	}
	if bootstrapEval == nil {
		return result, fmt.Errorf("bootstrapping.NewEvaluator returned nil")
	}
	result.FastBootstrapSelected = bootstrapEval.Evaluator == nil
	if result.FastBootstrapSelected && inputC1Nonzero != 0 {
		return result, fmt.Errorf("Fast native secret-key EncryptNew produced %d nonzero c1 coefficients; expected its pinned zero-c1 public lifecycle", inputC1Nonzero)
	}
	if !result.FastBootstrapSelected && inputC1Nonzero == 0 {
		return result, fmt.Errorf("genuine Standard native secret-key EncryptNew produced zero c1; expected ordinary Standard encryption")
	}

	trackingKeys := &trackingEvaluationKeySet{EvaluationKeySet: keys.MemEvaluationKeySet}
	rotateEval := ckks.NewEvaluator(residual, trackingKeys)
	trackingKeys.galoisKeyCalls, trackingKeys.galoisKeyListCalls = 0, 0
	if rotateEval == nil {
		return result, fmt.Errorf("public ckks.NewEvaluator returned nil")
	}
	result.InputUnchangedByRotate = true
	inputBeforeRotate := ciphertextSummary(ct, inputC1Nonzero, inputC1Count)
	rotated, err := rotateEval.RotateNew(ct, rotation)
	result.RotateGetGaloisKeyCalls = trackingKeys.galoisKeyCalls
	result.RotateGetGaloisKeyListCalls = trackingKeys.galoisKeyListCalls
	if err != nil {
		return result, fmt.Errorf("public ckks.Evaluator.RotateNew: %w", err)
	}
	if result.FastBootstrapSelected && result.RotateGetGaloisKeyCalls != 0 {
		return result, fmt.Errorf("Fast public Rotate unexpectedly called GetGaloisKey %d times", result.RotateGetGaloisKeyCalls)
	}
	if err = validateBootstrapBoundaryCiphertext("RotateNew output", rotated, residual); err != nil {
		return result, err
	}
	rotatedC1Nonzero, rotatedC1Count := nonzeroCount(rotated.Value[1].Coeffs)
	if result.FastBootstrapSelected && rotatedC1Nonzero != 0 {
		return result, fmt.Errorf("Fast public Rotate produced %d nonzero c1 coefficients", rotatedC1Nonzero)
	}
	result.RotatedCiphertext = ciphertextSummary(rotated, rotatedC1Nonzero, rotatedC1Count)
	inputC1NonzeroAfter, inputC1CountAfter := nonzeroCount(ct.Value[1].Coeffs)
	inputAfterRotate := ciphertextSummary(ct, inputC1NonzeroAfter, inputC1CountAfter)
	result.InputUnchangedByRotate = ciphertextEvidenceEqual(inputBeforeRotate, inputAfterRotate)
	if !result.InputUnchangedByRotate {
		return result, fmt.Errorf("RotateNew mutated its input ciphertext")
	}
	if !ct.Scale.Equal(rotated.Scale) || ct.Level() != rotated.Level() || ct.Degree() != rotated.Degree() ||
		ct.IsNTT != rotated.IsNTT || ct.IsMontgomery != rotated.IsMontgomery || !ct.MetaData.Equal(rotated.MetaData) {
		return result, fmt.Errorf("RotateNew changed Level/Scale/metadata/domain at the public boundary")
	}
	decodedRotate := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(rotated), decodedRotate); err != nil {
		return result, fmt.Errorf("public DecryptNew/Decode after RotateNew: %w", err)
	}
	rotateMetric := compareValues(expected, decodedRotate)
	result.RotateVsOracle = &rotateMetric
	result.DecodedRotateSHA256 = hashComplexValues(decodedRotate)
	result.RotatedDecodedValues = toComplexValues(decodedRotate)
	if !rotateMetric.Finite || rotateMetric.ComplexRMSE > preflightRMSETolerance {
		return result, fmt.Errorf("RotateNew oracle RMSE %.12g exceeds prior 004 preflight tolerance %.12g", rotateMetric.ComplexRMSE, preflightRMSETolerance)
	}
	result.Status = "preflight_passed"
	if phase == "preflight" {
		return result, nil
	}

	result.BootstrapCalls = 1
	result.BootstrapInputRotated = true
	output, err := bootstrapEval.Bootstrap(rotated)
	if err != nil {
		return result, fmt.Errorf("single public Bootstrap call: %w", err)
	}
	if err = validateBootstrapOutput("Bootstrap output", output, residual); err != nil {
		return result, err
	}
	outputC1Nonzero, outputC1Count := nonzeroCount(output.Value[1].Coeffs)
	result.BootstrapOutput = ciphertextSummary(output, outputC1Nonzero, outputC1Count)
	if result.FastBootstrapSelected && outputC1Nonzero != 0 {
		return result, fmt.Errorf("Fast Bootstrap output has %d nonzero c1 coefficients", outputC1Nonzero)
	}
	decodedBootstrap := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(output), decodedBootstrap); err != nil {
		return result, fmt.Errorf("public DecryptNew/Decode after Bootstrap: %w", err)
	}
	bootstrapMetric := compareValues(expected, decodedBootstrap)
	result.BootstrapVsOracle = &bootstrapMetric
	result.DecodedBootstrapSHA256 = hashComplexValues(decodedBootstrap)
	result.BootstrapDecodedValues = toComplexValues(decodedBootstrap)
	if !bootstrapMetric.Finite {
		return result, fmt.Errorf("Bootstrap-vs-rotated-oracle metrics are non-finite")
	}
	result.Status = "bootstrap_completed"
	return result, nil
}

func validateBootstrapBoundaryCiphertext(name string, ct *rlwe.Ciphertext, params ckks.Parameters) error {
	if ct == nil || ct.MetaData == nil {
		return fmt.Errorf("%s ciphertext or metadata is nil", name)
	}
	if ct.Degree() != 1 || ct.Level() != 0 || len(ct.Value) != 2 {
		return fmt.Errorf("%s expected degree-one Level-0 ciphertext, got degree=%d level=%d components=%d", name, ct.Degree(), ct.Level(), len(ct.Value))
	}
	if ct.N() != params.N() || !ct.IsNTT || ct.IsMontgomery || ct.Scale.Cmp(rlwe.NewScale(0)) != 1 {
		return fmt.Errorf("%s has unsupported ring degree/domain/scale", name)
	}
	if len(ct.Value[0].Coeffs) != 1 || len(ct.Value[1].Coeffs) != 1 ||
		len(ct.Value[0].Coeffs[0]) != params.N() || len(ct.Value[1].Coeffs[0]) != params.N() {
		return fmt.Errorf("%s must have one fully materialized q0 row in each component", name)
	}
	return nil
}

func parametersFromConfig(cfg config) (ckks.Parameters, bootstrapping.Parameters, int, error) {
	if cfg.LogN != 13 || cfg.LogDefaultScale != 45 || cfg.SecretHamming != 192 ||
		len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || len(cfg.QSlotsToCoeffs) != 3 ||
		len(cfg.QCoeffsToSlots) != 4 || len(cfg.P) != 5 || len(cfg.SlotsToCoeffsDFT) != 3 ||
		len(cfg.CoeffsToSlotsDFT) != 4 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 ||
		cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("config does not match pinned canonical LogN13 E32 profile")
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
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: cfg.LogN, LogQ: []int{cfg.Q0[0], cfg.QSlotsToCoeffs[0]}, LogDefaultScale: cfg.LogDefaultScale,
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
	logN, logScale, degree := cfg.LogN, cfg.Mod1LogScale, cfg.Mod1Degree
	doubleAngle, k := cfg.Mod1DoubleAngle, cfg.Mod1K
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
	_, residualQBits := primeEvidence(residual.Q())
	return parameterEvidence{
		ResidualLogN: residual.LogN(), ResidualMaxLevel: residual.MaxLevel(), ResidualScaleLog2: residual.DefaultScale().Log2(),
		BootstrapLogN: bootstrap.LogN(), BootstrapMaxLevel: bootstrap.MaxLevel(), LogSlots: logSlots, Slots: slots,
		ResidualQBits: residualQBits, BootstrapQBits: qBits, BootstrapPBits: pBits,
		BootstrapQSHA256: qHash, BootstrapPSHA256: pHash,
		EphemeralSecretWeight: btp.EphemeralSecretWeight, CircuitOrder: "ModUpThenEncode",
	}
}

func ciphertextSummary(ct *rlwe.Ciphertext, c1Nonzero, c1Count int) *ciphertextEvidence {
	result := &ciphertextEvidence{
		Level: ct.Level(), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(), N: ct.N(), ComponentCount: len(ct.Value),
		ActiveRows: ct.Level() + 1, IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		IsBatched: ct.IsBatched, IsBitReversed: ct.IsBitReversed,
		LogDimensionsRows: ct.LogDimensions.Rows, LogDimensionsCols: ct.LogDimensions.Cols, LogSlots: ct.LogSlots(),
		C1Zero: c1Nonzero == 0, C1NonzeroCount: c1Nonzero, C1CoefficientCount: c1Count,
		FullActiveQ: true, RowsPerComponent: make([]int, len(ct.Value)),
	}
	for component := range ct.Value {
		result.RowsPerComponent[component] = len(ct.Value[component].Coeffs)
		if len(ct.Value[component].Coeffs) != ct.Level()+1 {
			result.FullActiveQ = false
		}
	}
	if len(ct.Value) >= 2 {
		result.C0RowSHA256 = hashPolyRows(ct.Value[0].Coeffs)
		result.C1RowSHA256 = hashPolyRows(ct.Value[1].Coeffs)
	}
	return result
}

func countKeyLayouts(keys *bootstrapping.EvaluationKeys) (standard, fast int) {
	for _, key := range keys.GaloisKeys {
		if key == nil {
			continue
		}
		layout := reflect.ValueOf(key).Elem().FieldByName("Layout")
		if !layout.IsValid() {
			// The pinned Standard API predates explicit layout metadata and
			// materializes ordinary Standard-layout Galois keys.
			standard++
			continue
		}
		var layoutCode int64
		switch layout.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			layoutCode = layout.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			layoutCode = int64(layout.Uint())
		default:
			continue
		}
		switch layoutCode {
		case 0:
			standard++
		case 1:
			fast++
		}
	}
	return
}

func validateBootstrapOutput(name string, ct *rlwe.Ciphertext, residual ckks.Parameters) error {
	if ct == nil || ct.MetaData == nil || ct.Degree() != 1 || ct.N() != residual.N() {
		return fmt.Errorf("%s has invalid public CKKS ciphertext structure", name)
	}
	if ct.Level() < 0 || ct.Level() > residual.MaxLevel() || !ct.IsNTT || ct.IsMontgomery {
		return fmt.Errorf("%s has invalid level/domain", name)
	}
	activeRows := ct.Level() + 1
	for component := range ct.Value {
		if len(ct.Value[component].Coeffs) < activeRows {
			return fmt.Errorf("%s component %d lacks full active-Q rows", name, component)
		}
		for row := 0; row < activeRows; row++ {
			if len(ct.Value[component].Coeffs[row]) != residual.N() {
				return fmt.Errorf("%s component %d q%d row has invalid backing", name, component, row)
			}
		}
	}
	return nil
}

func deterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for i := range values {
		values[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return values
}

func rotateLeftOracle(values []complex128, k int) []complex128 {
	rotated := make([]complex128, len(values))
	shift := k % len(values)
	if shift < 0 {
		shift += len(values)
	}
	for i := range rotated {
		rotated[i] = values[(i+shift)%len(values)]
	}
	return rotated
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
	if standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 || standard.ConfigSHA256 != fast.ConfigSHA256 || standard.InputSHA256 != fast.InputSHA256 {
		return comparisonEvidence{}, fmt.Errorf("Standard/Fast source, config, and input hashes must match")
	}
	if standard.Parameters.BootstrapQSHA256 != fast.Parameters.BootstrapQSHA256 || standard.Parameters.BootstrapPSHA256 != fast.Parameters.BootstrapPSHA256 {
		return comparisonEvidence{}, fmt.Errorf("Standard/Fast generated Q/P prime hashes differ")
	}
	if standard.FastBootstrapSelected || !fast.FastBootstrapSelected {
		return comparisonEvidence{}, fmt.Errorf("expected genuine Standard and Fast public Bootstrap dispatch")
	}
	if standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 || !standard.BootstrapInputRotated || !fast.BootstrapInputRotated {
		return comparisonEvidence{}, fmt.Errorf("expected exactly one Bootstrap per backend, each consuming the rotated ciphertext")
	}
	if fast.RotateGetGaloisKeyCalls != 0 || fast.BootstrapOutput == nil || !fast.BootstrapOutput.C1Zero {
		return comparisonEvidence{}, fmt.Errorf("Fast keyless Rotate or zero-c1 Bootstrap output contract failed")
	}
	standardRotated := fromComplexValues(standard.RotatedDecodedValues)
	fastRotated := fromComplexValues(fast.RotatedDecodedValues)
	standardBootstrap := fromComplexValues(standard.BootstrapDecodedValues)
	fastBootstrap := fromComplexValues(fast.BootstrapDecodedValues)
	if len(standardRotated) == 0 || len(standardRotated) != len(fastRotated) || len(standardBootstrap) != len(fastBootstrap) {
		return comparisonEvidence{}, fmt.Errorf("Standard/Fast decoded output lengths differ or are empty")
	}
	result := comparisonEvidence{
		SchemaVersion: "fast-dropin-public-rotate-bootstrap-composition-010.compare.v1",
		TimestampUTC:  time.Now().UTC().Format(time.RFC3339Nano),
		PrimaryCommit: fast.PrimaryCommit, PrimarySourceSHA256: fast.PrimarySourceSHA256,
		ConfigSHA256: fast.ConfigSHA256, InputSHA256: fast.InputSHA256,
		RotateK: rotation, RotateGaloisElement: fast.RotateGaloisElement,
		BootstrapCallsStandard: standard.BootstrapCalls, BootstrapCallsFast: fast.BootstrapCalls,
		Standard: summarizeRun(standard), Fast: summarizeRun(fast),
		RotatedFastVsStandard:   compareValues(standardRotated, fastRotated),
		BootstrapFastVsStandard: compareValues(standardBootstrap, fastBootstrap),
	}
	return result, nil
}

func summarizeRun(run runEvidence) backendSummary {
	return backendSummary{
		BackendCommit: run.BackendCommit, BackendRef: run.BackendRef, BackendDirty: run.BackendDirty,
		FastBootstrapSelected: run.FastBootstrapSelected, EvaluationKeys: run.EvaluationKeys,
		RotateGetGaloisKeyCalls: run.RotateGetGaloisKeyCalls, Parameters: run.Parameters,
		InputCiphertext: run.InputCiphertext, RotatedCiphertext: run.RotatedCiphertext,
		BootstrapOutput: run.BootstrapOutput, ImmediateDecryption: run.ImmediateDecryption,
		RotateVsOracle: run.RotateVsOracle, BootstrapVsOracle: run.BootstrapVsOracle,
		DecodedInputSHA256: run.DecodedInputSHA256, DecodedRotateSHA256: run.DecodedRotateSHA256,
		DecodedBootstrapSHA256: run.DecodedBootstrapSHA256,
	}
}

func containsGaloisElement(values []uint64, target uint64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func ciphertextEvidenceEqual(a, b *ciphertextEvidence) bool {
	if a == nil || b == nil {
		return a == b
	}
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}

func hashPolyRows(rows [][]uint64) []string {
	hashes := make([]string, len(rows))
	var encoded [8]byte
	for row := range rows {
		h := sha256.New()
		for _, value := range rows[row] {
			binary.LittleEndian.PutUint64(encoded[:], value)
			_, _ = h.Write(encoded[:])
		}
		hashes[row] = hex.EncodeToString(h.Sum(nil))
	}
	return hashes
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
		return fmt.Errorf("encode evidence: %w", err)
	}
	encoded = append(encoded, '\n')
	if err = os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("write evidence %q: %w", path, err)
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
