// Command fast-dropin-rotate-bootstrap-keyplan-011 exercises one unchanged
// public CKKS frontend against the dependency selected by GOWORK.
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
	canonicalInputSHA256   = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
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

type keyEvidence struct {
	GeneratedBy               string `json:"generated_by"`
	GaloisKeyCount            int    `json:"galois_key_count"`
	StandardLayoutKeyCount    int    `json:"standard_layout_key_count"`
	FastLayoutKeyCount        int    `json:"fast_layout_key_count"`
	RotationGaloisElement     uint64 `json:"rotation_galois_element"`
	RotationKeyPresent        bool   `json:"rotation_key_present"`
	RotationKeyLayout         string `json:"rotation_key_layout"`
	RotationKeyLevelP         int    `json:"rotation_key_level_p"`
	EvaluatorMaxLevelP        int    `json:"evaluator_max_level_p"`
	RelinearizationKeyPresent bool   `json:"relinearization_key_present"`
	SecretQ0Preserved         bool   `json:"secret_q0_preserved"`
	InputSecretLevelQ         int    `json:"input_secret_level_q"`
	InputSecretLevelP         int    `json:"input_secret_level_p"`
	BootstrapSecretLevelQ     int    `json:"bootstrap_secret_level_q"`
	BootstrapSecretLevelP     int    `json:"bootstrap_secret_level_p"`
}

type runEvidence struct {
	SchemaVersion                   string              `json:"schema_version"`
	Phase                           string              `json:"phase"`
	Status                          string              `json:"status"`
	TimestampUTC                    string              `json:"timestamp_utc"`
	GoVersion                       string              `json:"go_version"`
	OS                              string              `json:"os"`
	Architecture                    string              `json:"architecture"`
	PrimaryCommit                   string              `json:"primary_commit"`
	PrimaryDirty                    bool                `json:"primary_dirty"`
	FrontendSHA256                  string              `json:"frontend_sha256"`
	BackendCommit                   string              `json:"backend_commit"`
	BackendRef                      string              `json:"backend_ref"`
	BackendDirty                    bool                `json:"backend_dirty"`
	ConfigSHA256                    string              `json:"config_sha256"`
	InputSHA256                     string              `json:"input_sha256"`
	ResidualN                       int                 `json:"residual_n"`
	BootstrapN                      int                 `json:"bootstrap_n"`
	ResidualRingType                string              `json:"residual_ring_type"`
	BootstrapRingType               string              `json:"bootstrap_ring_type"`
	ResidualQPrimeBits              []int               `json:"residual_q_prime_bits"`
	BootstrapQPrimeBits             []int               `json:"bootstrap_q_prime_bits"`
	BootstrapPPrimeBits             []int               `json:"bootstrap_p_prime_bits"`
	ResidualQSHA256                 string              `json:"residual_q_sha256"`
	BootstrapQSHA256                string              `json:"bootstrap_q_sha256"`
	BootstrapPSHA256                string              `json:"bootstrap_p_sha256"`
	ResidualQIsExactBootstrapPrefix bool                `json:"residual_q_is_exact_bootstrap_prefix"`
	Q0PrimeEqual                    bool                `json:"q0_prime_equal"`
	QPrimePrefixCount               int                 `json:"q_prime_prefix_count"`
	ResidualMaxLevel                int                 `json:"residual_max_level"`
	BootstrapMaxLevel               int                 `json:"bootstrap_max_level"`
	LogSlots                        int                 `json:"log_slots"`
	Slots                           int                 `json:"slots"`
	RotateK                         int                 `json:"rotate_k"`
	RotateGaloisElement             uint64              `json:"rotate_galois_element"`
	RotateEvaluatorParameters       string              `json:"rotate_evaluator_parameters"`
	EvaluationKeys                  keyEvidence         `json:"evaluation_keys"`
	FastBootstrapSelected           bool                `json:"fast_bootstrap_selected"`
	RotateGetGaloisKeyCalls         int                 `json:"rotate_get_galois_key_calls"`
	RotateGetGaloisKeyListCalls     int                 `json:"rotate_get_galois_key_list_calls"`
	InputCiphertext                 *ciphertextEvidence `json:"input_ciphertext,omitempty"`
	InputUnchangedByRotate          bool                `json:"input_unchanged_by_rotate"`
	RotatedCiphertext               *ciphertextEvidence `json:"rotated_ciphertext,omitempty"`
	ImmediateDecryption             *metric             `json:"immediate_decryption,omitempty"`
	RotateVsOracle                  *metric             `json:"rotate_vs_rotated_oracle,omitempty"`
	DecodedInputSHA256              string              `json:"decoded_input_sha256,omitempty"`
	DecodedRotateSHA256             string              `json:"decoded_rotate_sha256,omitempty"`
	BootstrapCalls                  int                 `json:"bootstrap_calls"`
	BootstrapInputIsRotated         bool                `json:"bootstrap_input_is_rotated_ciphertext"`
	BootstrapOutput                 *ciphertextEvidence `json:"bootstrap_output,omitempty"`
	BootstrapVsOracle               *metric             `json:"bootstrap_vs_rotated_oracle,omitempty"`
	DecodedBootstrapSHA256          string              `json:"decoded_bootstrap_sha256,omitempty"`
	Error                           string              `json:"error,omitempty"`
	InputValues                     []complexValue      `json:"input_values,omitempty"`
	InputDecodedValues              []complexValue      `json:"input_decoded_values,omitempty"`
	RotatedDecodedValues            []complexValue      `json:"rotated_decoded_values,omitempty"`
	BootstrapDecodedValues          []complexValue      `json:"bootstrap_decoded_values,omitempty"`
}

type comparisonEvidence struct {
	SchemaVersion           string     `json:"schema_version"`
	TimestampUTC            string     `json:"timestamp_utc"`
	PrimaryCommit           string     `json:"primary_commit"`
	FrontendSHA256          string     `json:"frontend_sha256"`
	ConfigSHA256            string     `json:"config_sha256"`
	InputSHA256             string     `json:"input_sha256"`
	StandardBackendCommit   string     `json:"standard_backend_commit"`
	FastBackendCommit       string     `json:"fast_backend_commit"`
	BootstrapCallsStandard  int        `json:"bootstrap_calls_standard"`
	BootstrapCallsFast      int        `json:"bootstrap_calls_fast"`
	StandardStatus          string     `json:"standard_status"`
	FastStatus              string     `json:"fast_status"`
	Standard                runSummary `json:"standard"`
	Fast                    runSummary `json:"fast"`
	RotateFastVsStandard    *metric    `json:"rotate_fast_vs_standard,omitempty"`
	BootstrapFastVsStandard *metric    `json:"bootstrap_fast_vs_standard,omitempty"`
}

type runSummary struct {
	FastBootstrapSelected  bool                `json:"fast_bootstrap_selected"`
	EvaluationKeys         keyEvidence         `json:"evaluation_keys"`
	ResidualQSHA256        string              `json:"residual_q_sha256"`
	BootstrapQSHA256       string              `json:"bootstrap_q_sha256"`
	BootstrapPSHA256       string              `json:"bootstrap_p_sha256"`
	ResidualMaxLevel       int                 `json:"residual_max_level"`
	BootstrapMaxLevel      int                 `json:"bootstrap_max_level"`
	InputCiphertext        *ciphertextEvidence `json:"input_ciphertext,omitempty"`
	RotatedCiphertext      *ciphertextEvidence `json:"rotated_ciphertext,omitempty"`
	BootstrapOutput        *ciphertextEvidence `json:"bootstrap_output,omitempty"`
	ImmediateDecryption    *metric             `json:"immediate_decryption,omitempty"`
	RotateVsOracle         *metric             `json:"rotate_vs_rotated_oracle,omitempty"`
	BootstrapVsOracle      *metric             `json:"bootstrap_vs_rotated_oracle,omitempty"`
	DecodedInputSHA256     string              `json:"decoded_input_sha256,omitempty"`
	DecodedRotateSHA256    string              `json:"decoded_rotate_sha256,omitempty"`
	DecodedBootstrapSHA256 string              `json:"decoded_bootstrap_sha256,omitempty"`
}

type trackingKeySet struct {
	rlwe.EvaluationKeySet
	galoisKeyCalls     int
	galoisKeyListCalls int
}

func (keys *trackingKeySet) GetGaloisKey(galEl uint64) (*rlwe.GaloisKey, error) {
	keys.galoisKeyCalls++
	return keys.EvaluationKeySet.GetGaloisKey(galEl)
}

func (keys *trackingKeySet) GetGaloisKeysList() []uint64 {
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
	configPath := flag.String("config", "", "canonical LogN13 E32 config")
	outPath := flag.String("out", "", "evidence output path")
	primaryCommit := flag.String("primary-commit", "", "Primary commit used for this run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty")
	backendCommit := flag.String("backend-commit", "", "pinned Lattigo commit")
	backendRef := flag.String("backend-ref", "", "Lattigo branch or pinned ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether Lattigo was dirty")
	standardPath := flag.String("standard-result", "", "Standard run evidence for compare")
	fastPath := flag.String("fast-result", "", "Fast run evidence for compare")
	flag.Parse()
	if *outPath == "" {
		return fmt.Errorf("-out is required")
	}
	if *phase == "compare" {
		if *standardPath == "" || *fastPath == "" {
			return fmt.Errorf("compare phase requires -standard-result and -fast-result")
		}
		comparison, err := compareFiles(*standardPath, *fastPath)
		if err != nil {
			return err
		}
		return writeJSON(*outPath, comparison)
	}
	if *phase != "preflight" && *phase != "bootstrap" {
		return fmt.Errorf("unsupported phase %q", *phase)
	}
	if *configPath == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return fmt.Errorf("run phase requires config, Primary/backend commit and backend ref")
	}
	result, runErr := runBackend(*phase, *configPath, *primaryCommit, *primaryDirty, *backendCommit, *backendRef, *backendDirty)
	if err := writeJSON(*outPath, result); err != nil {
		return err
	}
	return runErr
}

func runBackend(phase, configPath, primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty bool) (result runEvidence, runErr error) {
	result = runEvidence{
		SchemaVersion: "fast-dropin-rotate-bootstrap-keyplan-011.v1", Phase: phase, Status: "preflight_started",
		TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano), GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		PrimaryCommit: primaryCommit, PrimaryDirty: primaryDirty, BackendCommit: backendCommit, BackendRef: backendRef, BackendDirty: backendDirty,
		RotateK: rotation, RotateEvaluatorParameters: "btpParams.BootstrappingParameters",
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			runErr = fmt.Errorf("panic: %v", recovered)
		}
		if runErr != nil {
			if phase == "bootstrap" && result.BootstrapCalls == 1 {
				result.Status = "bootstrap_failed"
			} else {
				result.Status = "preflight_failed"
			}
			result.Error = runErr.Error()
		}
	}()

	// Capture the exact shared source identity before parameter/key generation or
	// any CKKS operation, so recovered panics still have provenance.
	frontend, err := os.ReadFile("tools/fast-dropin-rotate-bootstrap-keyplan-011/main.go")
	if err != nil {
		return result, fmt.Errorf("read shared frontend source before execution: %w", err)
	}
	result.FrontendSHA256 = hashBytes(frontend)
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return result, fmt.Errorf("read canonical config: %w", err)
	}
	result.ConfigSHA256 = hashBytes(configBytes)
	if result.ConfigSHA256 != canonicalConfigSHA256 {
		return result, fmt.Errorf("canonical config SHA mismatch: %s", result.ConfigSHA256)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return result, fmt.Errorf("decode config: %w", err)
	}
	residual, btpParams, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return result, err
	}
	result.ResidualN, result.BootstrapN = residual.N(), btpParams.BootstrappingParameters.N()
	result.ResidualRingType, result.BootstrapRingType = residual.RingType().String(), btpParams.BootstrappingParameters.RingType().String()
	result.ResidualMaxLevel, result.BootstrapMaxLevel = residual.MaxLevel(), btpParams.BootstrappingParameters.MaxLevel()
	result.LogSlots, result.Slots = logSlots, 1<<logSlots
	residualQ, bootstrapQ, bootstrapP := residual.Q(), btpParams.BootstrappingParameters.Q(), btpParams.BootstrappingParameters.P()
	result.ResidualQSHA256, result.ResidualQPrimeBits = primeEvidence(residualQ)
	result.BootstrapQSHA256, result.BootstrapQPrimeBits = primeEvidence(bootstrapQ)
	result.BootstrapPSHA256, result.BootstrapPPrimeBits = primeEvidence(bootstrapP)
	result.ResidualQIsExactBootstrapPrefix = exactOrderedPrefix(residualQ, bootstrapQ)
	result.QPrimePrefixCount = len(residualQ)
	result.Q0PrimeEqual = len(residualQ) > 0 && len(bootstrapQ) > 0 && residualQ[0] == bootstrapQ[0]
	if result.ResidualN != result.BootstrapN || residual.RingType() != ring.Standard || btpParams.BootstrappingParameters.RingType() != ring.Standard {
		return result, fmt.Errorf("pre-Rotate compatibility failure: require same N and Standard ring (residual N=%d/type=%s, bootstrap N=%d/type=%s)", result.ResidualN, result.ResidualRingType, result.BootstrapN, result.BootstrapRingType)
	}
	if !result.ResidualQIsExactBootstrapPrefix || !result.Q0PrimeEqual {
		return result, fmt.Errorf("pre-Rotate compatibility failure: residual Q is not an exact ordered Bootstrap-Q prefix with identical q0")
	}
	if residual.MaxLevel() != 1 || btpParams.BootstrappingParameters.MaxLevel() != 16 || logSlots != 12 {
		return result, fmt.Errorf("canonical profile mismatch: residual level=%d bootstrap level=%d logSlots=%d", residual.MaxLevel(), btpParams.BootstrappingParameters.MaxLevel(), logSlots)
	}
	values := deterministicInput(1 << logSlots)
	if result.InputSHA256 = hashComplexValues(values); result.InputSHA256 != canonicalInputSHA256 {
		return result, fmt.Errorf("deterministic input SHA mismatch: %s", result.InputSHA256)
	}
	expectedRotate := rotateLeftOracle(values, rotation)
	result.RotateGaloisElement = residual.GaloisElement(rotation)
	bootstrapGalEl := btpParams.BootstrappingParameters.GaloisElement(rotation)
	if result.RotateGaloisElement != bootstrapGalEl {
		return result, fmt.Errorf("rotation Galois element differs across residual/bootstrap parameters: %d != %d", result.RotateGaloisElement, bootstrapGalEl)
	}

	// Generate one native residual secret and extend it through the public
	// Bootstrap key-generation API. No secret coefficients are serialized.
	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	keys, bootstrapSecret, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return result, fmt.Errorf("public GenEvaluationKeys: %w", err)
	}
	if keys == nil || keys.MemEvaluationKeySet == nil || bootstrapSecret == nil {
		return result, fmt.Errorf("GenEvaluationKeys returned incomplete evaluation keys/secret")
	}
	result.EvaluationKeys = inspectKeys(keys, result.RotateGaloisElement, btpParams.BootstrappingParameters.MaxLevelP())
	result.EvaluationKeys.GeneratedBy = "btpParams.GenEvaluationKeys(sk)"
	result.EvaluationKeys.InputSecretLevelQ, result.EvaluationKeys.InputSecretLevelP = sk.LevelQ(), sk.LevelP()
	result.EvaluationKeys.BootstrapSecretLevelQ, result.EvaluationKeys.BootstrapSecretLevelP = bootstrapSecret.LevelQ(), bootstrapSecret.LevelP()
	result.EvaluationKeys.SecretQ0Preserved = equalSecretQ0(sk, bootstrapSecret)
	if !result.EvaluationKeys.SecretQ0Preserved {
		return result, fmt.Errorf("GenEvaluationKeys did not preserve original residual secret q0 residues")
	}
	if !result.EvaluationKeys.RotationKeyPresent {
		return result, fmt.Errorf("Bootstrap evaluation-key plan lacks rotation Galois element %d", result.RotateGaloisElement)
	}
	bootstrapEval, err := bootstrapping.NewEvaluator(btpParams, keys)
	if err != nil {
		return result, fmt.Errorf("public bootstrapping.NewEvaluator: %w", err)
	}
	if bootstrapEval == nil {
		return result, fmt.Errorf("bootstrapping.NewEvaluator returned nil")
	}
	result.FastBootstrapSelected = bootstrapEval.Evaluator == nil
	if !result.FastBootstrapSelected && (result.EvaluationKeys.RotationKeyLayout != "implicit-standard" && result.EvaluationKeys.RotationKeyLayout != "standard" || result.EvaluationKeys.RotationKeyLevelP != result.EvaluationKeys.EvaluatorMaxLevelP) {
		return result, fmt.Errorf("Standard rotation key layout/P level incompatible with Bootstrap evaluator: layout=%s keyP=%d evaluatorMaxP=%d", result.EvaluationKeys.RotationKeyLayout, result.EvaluationKeys.RotationKeyLevelP, result.EvaluationKeys.EvaluatorMaxLevelP)
	}
	if result.FastBootstrapSelected && result.EvaluationKeys.RotationKeyLayout != "fast" {
		return result, fmt.Errorf("Fast evaluation-key plan did not expose Fast-layout Galois key: %s", result.EvaluationKeys.RotationKeyLayout)
	}

	encoder := ckks.NewEncoder(residual)
	plaintext := ckks.NewPlaintext(residual, 0)
	plaintext.Scale = rlwe.NewScale(math.Exp2(float64(cfg.LogDefaultScale)))
	plaintext.IsNTT, plaintext.IsMontgomery = true, false
	plaintext.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err = encoder.Encode(values, plaintext); err != nil {
		return result, fmt.Errorf("public CKKS Encode: %w", err)
	}
	ct, err := rlwe.NewEncryptor(residual, sk).EncryptNew(plaintext)
	if err != nil {
		return result, fmt.Errorf("public secret-key EncryptNew: %w", err)
	}
	if err = validateLevelZeroCiphertext("EncryptNew", ct, residual); err != nil {
		return result, err
	}
	inputC1Nonzero, inputC1Count := nonzeroCount(ct.Value[1].Coeffs)
	if !result.FastBootstrapSelected && inputC1Nonzero == 0 {
		return result, fmt.Errorf("genuine Standard EncryptNew produced zero c1")
	}
	if result.FastBootstrapSelected && inputC1Nonzero != 0 {
		return result, fmt.Errorf("Fast zero-secret EncryptNew produced %d nonzero c1 coefficients", inputC1Nonzero)
	}
	result.InputCiphertext = summarizeCiphertext(ct, inputC1Nonzero, inputC1Count)
	result.InputValues = toComplexValues(values)
	decodedInput := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(ct), decodedInput); err != nil {
		return result, fmt.Errorf("public DecryptNew/Decode before Rotate: %w", err)
	}
	immediateMetric := compareValues(values, decodedInput)
	result.ImmediateDecryption = &immediateMetric
	result.DecodedInputSHA256 = hashComplexValues(decodedInput)
	result.InputDecodedValues = toComplexValues(decodedInput)
	if !immediateMetric.Finite || immediateMetric.ComplexRMSE > preflightRMSETolerance {
		return result, fmt.Errorf("immediate-decryption RMSE %.12g exceeds fixed tolerance %.12g", immediateMetric.ComplexRMSE, preflightRMSETolerance)
	}

	tracker := &trackingKeySet{EvaluationKeySet: keys.MemEvaluationKeySet}
	rotateEval := ckks.NewEvaluator(btpParams.BootstrappingParameters, tracker)
	if rotateEval == nil {
		return result, fmt.Errorf("public ckks.NewEvaluator returned nil")
	}
	tracker.galoisKeyCalls, tracker.galoisKeyListCalls = 0, 0
	inputBefore := summarizeCiphertext(ct, inputC1Nonzero, inputC1Count)
	rotated, err := rotateEval.RotateNew(ct, rotation)
	result.RotateGetGaloisKeyCalls, result.RotateGetGaloisKeyListCalls = tracker.galoisKeyCalls, tracker.galoisKeyListCalls
	if err != nil {
		return result, fmt.Errorf("public ckks.Evaluator.RotateNew: %w", err)
	}
	if result.FastBootstrapSelected && (result.RotateGetGaloisKeyCalls != 0 || result.RotateGetGaloisKeyListCalls != 0) {
		return result, fmt.Errorf("Fast public Rotate performed Galois key lookups: key=%d list=%d", result.RotateGetGaloisKeyCalls, result.RotateGetGaloisKeyListCalls)
	}
	if err = validateLevelZeroCiphertext("RotateNew output", rotated, btpParams.BootstrappingParameters); err != nil {
		return result, err
	}
	rotatedC1Nonzero, rotatedC1Count := nonzeroCount(rotated.Value[1].Coeffs)
	if result.FastBootstrapSelected && rotatedC1Nonzero != 0 {
		return result, fmt.Errorf("Fast RotateNew output has %d nonzero c1 coefficients", rotatedC1Nonzero)
	}
	result.RotatedCiphertext = summarizeCiphertext(rotated, rotatedC1Nonzero, rotatedC1Count)
	inputAfterC1Nonzero, inputAfterC1Count := nonzeroCount(ct.Value[1].Coeffs)
	result.InputUnchangedByRotate = ciphertextEvidenceEqual(inputBefore, summarizeCiphertext(ct, inputAfterC1Nonzero, inputAfterC1Count))
	if !result.InputUnchangedByRotate {
		return result, fmt.Errorf("RotateNew mutated its input ciphertext")
	}
	if !ct.Scale.Equal(rotated.Scale) || ct.Level() != rotated.Level() || ct.Degree() != rotated.Degree() || !ct.MetaData.Equal(rotated.MetaData) || ct.IsNTT != rotated.IsNTT || ct.IsMontgomery != rotated.IsMontgomery {
		return result, fmt.Errorf("RotateNew changed Level/Scale/metadata/domain")
	}
	decodedRotate := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(rotated), decodedRotate); err != nil {
		return result, fmt.Errorf("public DecryptNew/Decode after RotateNew with original residual secret: %w", err)
	}
	rotateMetric := compareValues(expectedRotate, decodedRotate)
	result.RotateVsOracle, result.DecodedRotateSHA256 = &rotateMetric, hashComplexValues(decodedRotate)
	result.RotatedDecodedValues = toComplexValues(decodedRotate)
	if !rotateMetric.Finite || rotateMetric.ComplexRMSE > preflightRMSETolerance {
		return result, fmt.Errorf("RotateNew oracle RMSE %.12g exceeds fixed tolerance %.12g", rotateMetric.ComplexRMSE, preflightRMSETolerance)
	}
	result.Status = "preflight_passed"
	if phase == "preflight" {
		return result, nil
	}

	result.BootstrapCalls = 1
	result.BootstrapInputIsRotated = true
	bootstrapOutput, err := bootstrapEval.Bootstrap(rotated)
	if err != nil {
		return result, fmt.Errorf("single public Bootstrap call: %w", err)
	}
	if err = validateBootstrapOutput(bootstrapOutput, residual); err != nil {
		return result, err
	}
	outputC1Nonzero, outputC1Count := nonzeroCount(bootstrapOutput.Value[1].Coeffs)
	if result.FastBootstrapSelected && outputC1Nonzero != 0 {
		return result, fmt.Errorf("Fast Bootstrap output has %d nonzero c1 coefficients", outputC1Nonzero)
	}
	result.BootstrapOutput = summarizeCiphertext(bootstrapOutput, outputC1Nonzero, outputC1Count)
	decodedBootstrap := make([]complex128, len(values))
	if err = encoder.Decode(rlwe.NewDecryptor(residual, sk).DecryptNew(bootstrapOutput), decodedBootstrap); err != nil {
		return result, fmt.Errorf("public DecryptNew/Decode after Bootstrap: %w", err)
	}
	bootstrapMetric := compareValues(expectedRotate, decodedBootstrap)
	result.BootstrapVsOracle, result.DecodedBootstrapSHA256 = &bootstrapMetric, hashComplexValues(decodedBootstrap)
	result.BootstrapDecodedValues = toComplexValues(decodedBootstrap)
	if !bootstrapMetric.Finite {
		return result, fmt.Errorf("Bootstrap oracle metrics are non-finite")
	}
	result.Status = "bootstrap_completed"
	return result, nil
}

func parametersFromConfig(cfg config) (ckks.Parameters, bootstrapping.Parameters, int, error) {
	if cfg.LogN != 13 || cfg.LogDefaultScale != 45 || cfg.SecretHamming != 192 ||
		len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || len(cfg.QSlotsToCoeffs) != 3 ||
		len(cfg.QCoeffsToSlots) != 4 || len(cfg.P) != 5 || len(cfg.SlotsToCoeffsDFT) != 3 ||
		len(cfg.CoeffsToSlotsDFT) != 4 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 ||
		cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("config is not the frozen LogN13 E32 profile")
	}
	for _, value := range cfg.QSlotsToCoeffs {
		if value != 39 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected S2C prime scale %d", value)
		}
	}
	for _, value := range cfg.QCoeffsToSlots {
		if value != 56 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected C2S prime scale %d", value)
		}
	}
	for _, value := range cfg.P {
		if value != 61 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected P prime scale %d", value)
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
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("Bootstrap parameters: %w", err)
	}
	params.CircuitOrder, params.ResidualParameters, params.EphemeralSecretWeight = bootstrapping.ModUpThenEncode, residual, 32
	return residual, params, logSlots, nil
}

func factorization(depths, scales []int) ([][]int, error) {
	if len(depths) == 0 || len(depths) != len(scales) {
		return nil, fmt.Errorf("DFT depth/scale lengths must match and be nonzero")
	}
	result := make([][]int, len(depths))
	for i, depth := range depths {
		if depth <= 0 || scales[i] <= 0 {
			return nil, fmt.Errorf("invalid DFT factorization index %d", i)
		}
		result[i] = make([]int, depth)
		for j := range result[i] {
			result[i][j] = scales[i]
		}
	}
	return result, nil
}

func inspectKeys(keys *bootstrapping.EvaluationKeys, rotationGalEl uint64, evaluatorMaxP int) (result keyEvidence) {
	result.GaloisKeyCount = len(keys.GaloisKeys)
	result.EvaluatorMaxLevelP = evaluatorMaxP
	result.RelinearizationKeyPresent = keys.RelinearizationKey != nil
	for _, key := range keys.GaloisKeys {
		if key == nil {
			continue
		}
		layout := keyLayout(key)
		switch layout {
		case "standard", "implicit-standard":
			result.StandardLayoutKeyCount++
		case "fast":
			result.FastLayoutKeyCount++
		}
		if key.GaloisElement == rotationGalEl {
			result.RotationKeyPresent = true
			result.RotationKeyLayout = layout
			result.RotationKeyLevelP = key.GadgetCiphertext.LevelP()
		}
	}
	return
}

func keyLayout(key *rlwe.GaloisKey) string {
	field := reflect.ValueOf(key).Elem().FieldByName("Layout")
	if !field.IsValid() {
		return "implicit-standard"
	}
	var value int64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = int64(field.Uint())
	default:
		return "unknown"
	}
	switch value {
	case 0:
		return "standard"
	case 1:
		return "fast"
	default:
		return fmt.Sprintf("layout-%d", value)
	}
}

func validateLevelZeroCiphertext(name string, ct *rlwe.Ciphertext, params ckks.Parameters) error {
	if ct == nil || ct.MetaData == nil {
		return fmt.Errorf("%s ciphertext/metadata is nil", name)
	}
	if ct.Degree() != 1 || ct.Level() != 0 || len(ct.Value) != 2 || ct.N() != params.N() {
		return fmt.Errorf("%s expected degree-one Level-0 ciphertext in N=%d ring; got degree=%d level=%d components=%d N=%d", name, params.N(), ct.Degree(), ct.Level(), len(ct.Value), ct.N())
	}
	if ct.Scale.Log2() != 45 || !ct.IsNTT || ct.IsMontgomery || len(ct.Value[0].Coeffs) != 1 || len(ct.Value[1].Coeffs) != 1 {
		return fmt.Errorf("%s violated Scale/domain/one-active-q0-row contract", name)
	}
	for component := range ct.Value {
		if len(ct.Value[component].Coeffs[0]) != params.N() {
			return fmt.Errorf("%s component %d q0 row length %d != N %d", name, component, len(ct.Value[component].Coeffs[0]), params.N())
		}
	}
	return nil
}

func validateBootstrapOutput(ct *rlwe.Ciphertext, residual ckks.Parameters) error {
	if ct == nil || ct.MetaData == nil || ct.Degree() != 1 || ct.N() != residual.N() || ct.Level() < 0 || ct.Level() > residual.MaxLevel() || !ct.IsNTT || ct.IsMontgomery {
		return fmt.Errorf("Bootstrap output has invalid public CKKS structure/level/domain")
	}
	activeRows := ct.Level() + 1
	for component := range ct.Value {
		if len(ct.Value[component].Coeffs) < activeRows {
			return fmt.Errorf("Bootstrap output component %d lacks active Q rows", component)
		}
		for q := 0; q < activeRows; q++ {
			if len(ct.Value[component].Coeffs[q]) != residual.N() {
				return fmt.Errorf("Bootstrap output component %d q%d row backing invalid", component, q)
			}
		}
	}
	return nil
}

func summarizeCiphertext(ct *rlwe.Ciphertext, c1Nonzero, c1Count int) *ciphertextEvidence {
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
		if len(ct.Value[component].Coeffs) < ct.Level()+1 {
			result.FullActiveQ = false
			continue
		}
		for q := 0; q <= ct.Level(); q++ {
			if len(ct.Value[component].Coeffs[q]) != ct.N() {
				result.FullActiveQ = false
			}
		}
	}
	if len(ct.Value) >= 2 {
		result.C0RowSHA256 = hashPolyRows(ct.Value[0].Coeffs)
		result.C1RowSHA256 = hashPolyRows(ct.Value[1].Coeffs)
	}
	return result
}

func ciphertextEvidenceEqual(a, b *ciphertextEvidence) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}

func deterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for i := range values {
		values[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return values
}

func rotateLeftOracle(values []complex128, k int) []complex128 {
	result := make([]complex128, len(values))
	shift := k % len(values)
	if shift < 0 {
		shift += len(values)
	}
	for i := range result {
		result[i] = values[(i+shift)%len(values)]
	}
	return result
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
	count := min(len(reference), len(actual))
	var errorPower, signalPower float64
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
		result.Finite, result.SNRState = false, "non_finite_values_or_aggregate"
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

func compareFiles(standardPath, fastPath string) (comparisonEvidence, error) {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return comparisonEvidence{}, fmt.Errorf("read Standard evidence: %w", err)
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return comparisonEvidence{}, fmt.Errorf("read Fast evidence: %w", err)
	}
	if standard.Phase != fast.Phase || (standard.Phase != "preflight" && standard.Phase != "bootstrap") {
		return comparisonEvidence{}, fmt.Errorf("evidence phases do not match or are unsupported")
	}
	if standard.FrontendSHA256 == "" || standard.FrontendSHA256 != fast.FrontendSHA256 || standard.ConfigSHA256 != fast.ConfigSHA256 || standard.InputSHA256 != fast.InputSHA256 || standard.BootstrapQSHA256 != fast.BootstrapQSHA256 || standard.BootstrapPSHA256 != fast.BootstrapPSHA256 {
		return comparisonEvidence{}, fmt.Errorf("Standard/Fast frontend, config, input, Q or P identities differ")
	}
	comparison := comparisonEvidence{
		SchemaVersion: "fast-dropin-rotate-bootstrap-keyplan-011.compare.v1", TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano),
		PrimaryCommit: standard.PrimaryCommit, FrontendSHA256: standard.FrontendSHA256, ConfigSHA256: standard.ConfigSHA256, InputSHA256: standard.InputSHA256,
		StandardBackendCommit: standard.BackendCommit, FastBackendCommit: fast.BackendCommit,
		BootstrapCallsStandard: standard.BootstrapCalls, BootstrapCallsFast: fast.BootstrapCalls,
		StandardStatus: standard.Status, FastStatus: fast.Status,
		Standard: summarizeRun(standard), Fast: summarizeRun(fast),
	}
	if standard.Status == "preflight_passed" && fast.Status == "preflight_passed" {
		comparison.RotateFastVsStandard = metricPtr(compareValues(fromComplexValues(standard.RotatedDecodedValues), fromComplexValues(fast.RotatedDecodedValues)))
	}
	if standard.Status == "bootstrap_completed" && fast.Status == "bootstrap_completed" {
		comparison.BootstrapFastVsStandard = metricPtr(compareValues(fromComplexValues(standard.BootstrapDecodedValues), fromComplexValues(fast.BootstrapDecodedValues)))
	}
	return comparison, nil
}

func summarizeRun(run runEvidence) runSummary {
	return runSummary{
		FastBootstrapSelected: run.FastBootstrapSelected, EvaluationKeys: run.EvaluationKeys,
		ResidualQSHA256: run.ResidualQSHA256, BootstrapQSHA256: run.BootstrapQSHA256, BootstrapPSHA256: run.BootstrapPSHA256,
		ResidualMaxLevel: run.ResidualMaxLevel, BootstrapMaxLevel: run.BootstrapMaxLevel,
		InputCiphertext: run.InputCiphertext, RotatedCiphertext: run.RotatedCiphertext, BootstrapOutput: run.BootstrapOutput,
		ImmediateDecryption: run.ImmediateDecryption, RotateVsOracle: run.RotateVsOracle, BootstrapVsOracle: run.BootstrapVsOracle,
		DecodedInputSHA256: run.DecodedInputSHA256, DecodedRotateSHA256: run.DecodedRotateSHA256, DecodedBootstrapSHA256: run.DecodedBootstrapSHA256,
	}
}

func metricPtr(value metric) *metric { return &value }

func exactOrderedPrefix(prefix, full []uint64) bool {
	if len(prefix) == 0 || len(prefix) > len(full) {
		return false
	}
	for i := range prefix {
		if prefix[i] != full[i] {
			return false
		}
	}
	return true
}

func equalSecretQ0(a, b *rlwe.SecretKey) bool {
	if a == nil || b == nil || len(a.Value.Q.Coeffs) == 0 || len(b.Value.Q.Coeffs) == 0 || len(a.Value.Q.Coeffs[0]) != len(b.Value.Q.Coeffs[0]) {
		return false
	}
	for i := range a.Value.Q.Coeffs[0] {
		if a.Value.Q.Coeffs[0][i] != b.Value.Q.Coeffs[0][i] {
			return false
		}
	}
	return true
}

func hashPolyRows(rows [][]uint64) []string {
	result := make([]string, len(rows))
	var encoded [8]byte
	for row := range rows {
		h := sha256.New()
		for _, coefficient := range rows[row] {
			binary.LittleEndian.PutUint64(encoded[:], coefficient)
			_, _ = h.Write(encoded[:])
		}
		result[row] = hex.EncodeToString(h.Sum(nil))
	}
	return result
}

func primeEvidence(primes []uint64) (string, []int) {
	values := make([]string, len(primes))
	primeBits := make([]int, len(primes))
	for i, prime := range primes {
		values[i], primeBits[i] = strconv.FormatUint(prime, 10), bits.Len64(prime)
	}
	encoded, _ := json.Marshal(values)
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

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func writeJSON(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	encoded = append(encoded, '\n')
	if err = os.WriteFile(path, encoded, 0o644); err != nil {
		return fmt.Errorf("write evidence %q: %w", path, err)
	}
	return nil
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
