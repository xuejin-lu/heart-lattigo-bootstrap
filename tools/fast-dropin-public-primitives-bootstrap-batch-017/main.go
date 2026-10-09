// Command fast-dropin-public-primitives-bootstrap-batch-017 runs one shared
// public CKKS frontend against the dependency selected by GOWORK.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	canonicalConfigSHA256 = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"
	canonicalInputSHA256  = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
	standardCommit        = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	fastCommit            = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac"
	frontendPath          = "tools/fast-dropin-public-primitives-bootstrap-batch-017/main.go"
	defaultScaleLog2      = 45
	rotation              = 1
	addMaxError           = 1e-6
	primitiveMaxError     = 1e-4
	bootstrapMaxError     = 1e-6
	pairedMaxError        = 1e-6
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

type keyLookups struct {
	GaloisKey     int `json:"galois_key"`
	GaloisKeyList int `json:"galois_key_list"`
	Relinearize   int `json:"relinearization_key"`
}

type inputScaleEvidence struct {
	InputAInteger string  `json:"input_a_integer_scale"`
	InputBInteger string  `json:"input_b_integer_scale"`
	InputCInteger string  `json:"input_c_integer_scale"`
	InputCLog2    float64 `json:"input_c_log2_scale"`
}

type primitiveLookupEvidence struct {
	EvaluatorConstruction keyLookups `json:"evaluator_construction"`
	Add                   keyLookups `json:"add"`
	MulRelin              keyLookups `json:"mul_relin"`
	Rescale               keyLookups `json:"rescale"`
	Rotate                keyLookups `json:"rotate"`
}

type keyEvidence struct {
	GeneratedBy            string `json:"generated_by"`
	GaloisKeyCount         int    `json:"galois_key_count"`
	RelinearizationPresent bool   `json:"relinearization_key_present"`
	RotationGaloisElement  uint64 `json:"rotation_galois_element"`
	RotationKeyPresent     bool   `json:"rotation_key_present"`
	RotationKeyLayout      string `json:"rotation_key_layout"`
	RotationKeyLevelP      int    `json:"rotation_key_level_p"`
	EvaluatorMaxLevelP     int    `json:"evaluator_max_level_p"`
	InputSecretLevelQ      int    `json:"input_secret_level_q"`
	InputSecretLevelP      int    `json:"input_secret_level_p"`
	BootstrapSecretLevelQ  int    `json:"bootstrap_secret_level_q"`
	BootstrapSecretLevelP  int    `json:"bootstrap_secret_level_p"`
	SecretQ0Preserved      bool   `json:"secret_q0_preserved"`
}

type checkpoint struct {
	ID            string         `json:"id"`
	Operation     string         `json:"operation"`
	State         stateEvidence  `json:"state"`
	Plaintext     metric         `json:"plaintext_oracle"`
	DecodedSHA256 string         `json:"decoded_sha256"`
	Decoded       []complexValue `json:"decoded,omitempty"`
}

type stateEvidence struct {
	Level              int      `json:"level"`
	ScaleLog2          float64  `json:"scale_log2"`
	Degree             int      `json:"degree"`
	N                  int      `json:"n"`
	ComponentCount     int      `json:"component_count"`
	ActiveRows         int      `json:"active_rows"`
	RowsPerComponent   []int    `json:"rows_per_component"`
	FullActiveQ        bool     `json:"full_active_q"`
	IsNTT              bool     `json:"is_ntt"`
	IsMontgomery       bool     `json:"is_montgomery"`
	C1Zero             bool     `json:"c1_zero"`
	C1NonzeroCount     int      `json:"c1_nonzero_count"`
	C1CoefficientCount int      `json:"c1_coefficient_count"`
	C0ActiveRowSHA256  []string `json:"c0_active_row_sha256"`
	C1ActiveRowSHA256  []string `json:"c1_active_row_sha256"`
}

type runEvidence struct {
	SchemaVersion                 string                  `json:"schema_version"`
	Status                        string                  `json:"status"`
	TimestampUTC                  string                  `json:"timestamp_utc"`
	GoVersion                     string                  `json:"go_version"`
	OS                            string                  `json:"os"`
	Architecture                  string                  `json:"architecture"`
	PrimaryCommit                 string                  `json:"primary_commit"`
	PrimaryDirty                  bool                    `json:"primary_dirty"`
	FrontendSHA256                string                  `json:"frontend_sha256"`
	BackendCommit                 string                  `json:"backend_commit"`
	BackendRef                    string                  `json:"backend_ref"`
	BackendDirty                  bool                    `json:"backend_dirty"`
	ConfigSHA256                  string                  `json:"config_sha256"`
	OriginalInputSHA256           string                  `json:"canonical_012_seed_input_sha256"`
	WorkloadInputSHA256           string                  `json:"workload_input_sha256"`
	ProfileSHA256                 string                  `json:"effective_profile_sha256"`
	QPSHA256                      string                  `json:"effective_qp_sha256"`
	Q                             []uint64                `json:"q_primes"`
	P                             []uint64                `json:"p_primes"`
	ResidualMaxLevel              int                     `json:"residual_max_level"`
	BootstrapMaxLevel             int                     `json:"bootstrap_max_level"`
	LogSlots                      int                     `json:"log_slots"`
	Slots                         int                     `json:"slots"`
	EphemeralSecretWeight         int                     `json:"ephemeral_secret_weight"`
	FastZeroSecretCapability      bool                    `json:"fast_zero_secret_capability"`
	FastBootstrapSelected         bool                    `json:"fast_bootstrap_selected"`
	EvaluationKeys                keyEvidence             `json:"evaluation_keys"`
	PrimitiveKeyLookups           primitiveLookupEvidence `json:"primitive_key_lookups"`
	BootstrapKeyAccessObservation string                  `json:"bootstrap_key_access_observation,omitempty"`
	InputScales                   inputScaleEvidence      `json:"input_scales"`
	BootstrapInputFingerprint     string                  `json:"bootstrap_input_fingerprint,omitempty"`
	BootstrapCalls                int                     `json:"bootstrap_calls"`
	BootstrapAttempted            bool                    `json:"bootstrap_attempted"`
	Checkpoints                   []checkpoint            `json:"checkpoints"`
	Error                         string                  `json:"error,omitempty"`
}

type pairedCheckpoint struct {
	ID                    string  `json:"id"`
	Level                 int     `json:"level"`
	StandardScaleLog2     float64 `json:"standard_scale_log2"`
	FastScaleLog2         float64 `json:"fast_scale_log2"`
	StandardPlaintextRMSE float64 `json:"standard_plaintext_rmse"`
	FastPlaintextRMSE     float64 `json:"fast_plaintext_rmse"`
	FastVsStandard        metric  `json:"fast_vs_standard"`
	Pass                  bool    `json:"pass"`
}

type combinedEvidence struct {
	SchemaVersion       string             `json:"schema_version"`
	Status              string             `json:"status"`
	CreatedUTC          string             `json:"created_utc"`
	PrimaryCommit       string             `json:"primary_commit"`
	PrimaryDirty        bool               `json:"primary_dirty"`
	FrontendSHA256      string             `json:"shared_frontend_sha256"`
	ConfigSHA256        string             `json:"config_sha256"`
	WorkloadInputSHA256 string             `json:"workload_input_sha256"`
	ProfileSHA256       string             `json:"effective_profile_sha256"`
	QPSHA256            string             `json:"effective_qp_sha256"`
	BootstrapCalls      int                `json:"bootstrap_calls_total"`
	Standard            runEvidence        `json:"standard"`
	Fast                runEvidence        `json:"fast"`
	Paired              []pairedCheckpoint `json:"paired_checkpoints"`
	StopReason          string             `json:"stop_reason,omitempty"`
}

type trackingKeySet struct {
	rlwe.EvaluationKeySet
	lookups keyLookups
}

func (keys *trackingKeySet) GetGaloisKey(galEl uint64) (*rlwe.GaloisKey, error) {
	keys.lookups.GaloisKey++
	return keys.EvaluationKeySet.GetGaloisKey(galEl)
}

func (keys *trackingKeySet) GetGaloisKeysList() []uint64 {
	keys.lookups.GaloisKeyList++
	return keys.EvaluationKeySet.GetGaloisKeysList()
}

func (keys *trackingKeySet) GetRelinearizationKey() (*rlwe.RelinearizationKey, error) {
	keys.lookups.Relinearize++
	return keys.EvaluationKeySet.GetRelinearizationKey()
}

func main() {
	if err := runCLI(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() error {
	mode := flag.String("mode", "run", "run or compare")
	configPath := flag.String("config", "", "canonical LogN13 E32 config")
	outPath := flag.String("out", "", "raw run JSON or compact comparison JSON")
	primaryCommit := flag.String("primary-commit", "", "Primary revision used for this run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty")
	backendCommit := flag.String("backend-commit", "", "pinned Lattigo revision")
	backendRef := flag.String("backend-ref", "", "Lattigo branch or pinned ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether Lattigo was dirty")
	hold := flag.Bool("hold-for-bootstrap", false, "wait for exactly BOOTSTRAP after a passing preflight")
	standardPath := flag.String("standard-run", "", "raw Standard run JSON for compare mode")
	fastPath := flag.String("fast-run", "", "raw Fast run JSON for compare mode")
	flag.Parse()
	if *outPath == "" {
		return errors.New("-out is required")
	}
	if *mode == "compare" {
		if *standardPath == "" || *fastPath == "" {
			return errors.New("compare mode requires -standard-run and -fast-run")
		}
		result, err := compareRunFiles(*standardPath, *fastPath)
		if writeErr := writeJSON(*outPath, result); writeErr != nil {
			return writeErr
		}
		if err != nil {
			return err
		}
		return nil
	}
	if *mode != "run" {
		return fmt.Errorf("unsupported mode %q", *mode)
	}
	if *configPath == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return errors.New("run mode requires -config, -primary-commit, -backend-commit and -backend-ref")
	}
	if *backendCommit != standardCommit && *backendCommit != fastCommit {
		return fmt.Errorf("backend commit %q is outside the frozen pins", *backendCommit)
	}
	result, runErr := executeRun(*configPath, *primaryCommit, *primaryDirty, *backendCommit, *backendRef, *backendDirty, *hold)
	if err := writeJSON(*outPath, result); err != nil {
		return err
	}
	if runErr != nil {
		return runErr
	}
	return nil
}

func executeRun(configPath, primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty, hold bool) (result runEvidence, runErr error) {
	result = runEvidence{
		SchemaVersion: "fast-dropin-public-primitives-bootstrap-batch-017.run.v1",
		Status:        "started", TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		PrimaryCommit: primaryCommit, PrimaryDirty: primaryDirty,
		BackendCommit: backendCommit, BackendRef: backendRef, BackendDirty: backendDirty,
		BootstrapCalls: 0,
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			runErr = fmt.Errorf("runner panic: %v", recovered)
		}
		if runErr != nil {
			if result.BootstrapAttempted {
				result.Status = "bootstrap_failed"
			} else {
				result.Status = "primitive_preflight_failed"
			}
			result.Error = runErr.Error()
		}
	}()

	frontend, err := os.ReadFile(frontendPath)
	if err != nil {
		return result, fmt.Errorf("read shared frontend before execution: %w", err)
	}
	result.FrontendSHA256 = hashBytes(frontend)
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return result, fmt.Errorf("read canonical config: %w", err)
	}
	result.ConfigSHA256 = hashBytes(configBytes)
	if result.ConfigSHA256 != canonicalConfigSHA256 {
		return result, fmt.Errorf("canonical config hash mismatch: %s", result.ConfigSHA256)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return result, fmt.Errorf("decode canonical config: %w", err)
	}
	residual, btpParams, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return result, err
	}
	if residual.MaxLevel() != 1 || btpParams.BootstrappingParameters.MaxLevel() != 16 || logSlots != 12 || btpParams.EphemeralSecretWeight != 32 {
		return result, fmt.Errorf("frozen E32 profile mismatch: residual=%d bootstrap=%d logSlots=%d E=%d", residual.MaxLevel(), btpParams.BootstrappingParameters.MaxLevel(), logSlots, btpParams.EphemeralSecretWeight)
	}
	if !exactOrderedPrefix(residual.Q(), btpParams.BootstrappingParameters.Q()) || len(residual.Q()) != 2 || residual.Q()[1] != btpParams.BootstrappingParameters.Q()[1] {
		return result, errors.New("residual Q is not the exact two-prime prefix of Bootstrap Q")
	}
	result.Q, result.P = btpParams.BootstrappingParameters.Q(), btpParams.BootstrappingParameters.P()
	result.ResidualMaxLevel, result.BootstrapMaxLevel = residual.MaxLevel(), btpParams.BootstrappingParameters.MaxLevel()
	result.LogSlots, result.Slots = logSlots, 1<<logSlots
	result.EphemeralSecretWeight = btpParams.EphemeralSecretWeight
	_, result.FastZeroSecretCapability = any(btpParams.BootstrappingParameters).(interface{ FastCKKSZeroSecretSimulation() })
	profilePayload, _ := json.Marshal(struct {
		LogN, DefaultScaleLog2, LogSlots, ResidualMaxLevel, BootstrapMaxLevel, E int
		Q, P                                                                     []uint64
	}{cfg.LogN, cfg.LogDefaultScale, logSlots, residual.MaxLevel(), btpParams.BootstrappingParameters.MaxLevel(), btpParams.EphemeralSecretWeight, result.Q, result.P})
	result.ProfileSHA256 = hashBytes(profilePayload)
	qpPayload, _ := json.Marshal(struct{ Q, P []uint64 }{result.Q, result.P})
	result.QPSHA256 = hashBytes(qpPayload)

	a, b, c := deterministicInputs(result.Slots)
	result.OriginalInputSHA256 = hashComplexValues(deterministicInput(result.Slots))
	if result.OriginalInputSHA256 != canonicalInputSHA256 {
		return result, fmt.Errorf("canonical 012 seed input hash mismatch: %s", result.OriginalInputSHA256)
	}
	q1 := residual.Q()[1]
	defaultScale := rlwe.NewScale(math.Exp2(float64(cfg.LogDefaultScale)))
	q1Scale := rlwe.NewScale(q1)
	result.InputScales = inputScaleEvidence{
		InputAInteger: defaultScale.BigInt().String(), InputBInteger: defaultScale.BigInt().String(),
		InputCInteger: q1Scale.BigInt().String(), InputCLog2: q1Scale.Log2(),
	}
	result.WorkloadInputSHA256 = hashWorkload(a, b, c, q1)

	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	keys, bootstrapSecret, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return result, fmt.Errorf("public GenEvaluationKeys: %w", err)
	}
	if keys == nil || keys.MemEvaluationKeySet == nil || bootstrapSecret == nil {
		return result, errors.New("public GenEvaluationKeys returned incomplete evaluation keys")
	}
	result.EvaluationKeys = inspectKeys(keys, btpParams.BootstrappingParameters.GaloisElement(rotation), btpParams.BootstrappingParameters.MaxLevelP())
	result.EvaluationKeys.GeneratedBy = "btpParams.GenEvaluationKeys(sk)"
	result.EvaluationKeys.InputSecretLevelQ, result.EvaluationKeys.InputSecretLevelP = sk.LevelQ(), sk.LevelP()
	result.EvaluationKeys.BootstrapSecretLevelQ, result.EvaluationKeys.BootstrapSecretLevelP = bootstrapSecret.LevelQ(), bootstrapSecret.LevelP()
	result.EvaluationKeys.SecretQ0Preserved = equalSecretQ0(sk, bootstrapSecret)
	if !result.EvaluationKeys.RelinearizationPresent || !result.EvaluationKeys.RotationKeyPresent || !result.EvaluationKeys.SecretQ0Preserved {
		return result, errors.New("public Bootstrap key plan lacks relin/rotation material or same-secret q0 extension")
	}
	bootstrapEval, err := bootstrapping.NewEvaluator(btpParams, keys)
	if err != nil {
		return result, fmt.Errorf("public bootstrapping.NewEvaluator: %w", err)
	}
	if bootstrapEval == nil {
		return result, errors.New("public bootstrapping.NewEvaluator returned nil")
	}
	result.FastBootstrapSelected = bootstrapEval.Evaluator == nil
	if result.FastBootstrapSelected != result.FastZeroSecretCapability {
		return result, errors.New("public Bootstrap dispatch does not match backend capability")
	}
	if result.FastBootstrapSelected && result.EvaluationKeys.RotationKeyLayout != "fast" {
		return result, fmt.Errorf("Fast key plan has unexpected rotation-key layout %q", result.EvaluationKeys.RotationKeyLayout)
	}
	if !result.FastBootstrapSelected && result.EvaluationKeys.RotationKeyLayout != "standard" && result.EvaluationKeys.RotationKeyLayout != "implicit-standard" {
		return result, fmt.Errorf("Standard key plan has unexpected rotation-key layout %q", result.EvaluationKeys.RotationKeyLayout)
	}
	if !result.FastBootstrapSelected && result.EvaluationKeys.RotationKeyLevelP != result.EvaluationKeys.EvaluatorMaxLevelP {
		return result, fmt.Errorf("Standard rotation key P level %d does not match evaluator max P level %d", result.EvaluationKeys.RotationKeyLevelP, result.EvaluationKeys.EvaluatorMaxLevelP)
	}

	tracked := &trackingKeySet{EvaluationKeySet: keys.MemEvaluationKeySet}
	evaluator := ckks.NewEvaluator(btpParams.BootstrappingParameters, tracked)
	encoder := ckks.NewEncoder(residual)
	decryptor := rlwe.NewDecryptor(residual, sk)
	if evaluator == nil {
		return result, errors.New("public ckks.NewEvaluator returned nil")
	}
	constructionLookups := tracked.lookups
	tracked.lookups = keyLookups{}

	ctA, err := encrypt(residual, encoder, sk, a, defaultScale, logSlots)
	if err != nil {
		return result, fmt.Errorf("EncryptNew input A: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "input-a", "EncryptNew", ctA, a, 1, defaultScale, 1e-7, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	ctB, err := encrypt(residual, encoder, sk, b, defaultScale, logSlots)
	if err != nil {
		return result, fmt.Errorf("EncryptNew input B: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "input-b", "EncryptNew", ctB, b, 1, defaultScale, 1e-7, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	ctC, err := encrypt(residual, encoder, sk, c, q1Scale, logSlots)
	if err != nil {
		return result, fmt.Errorf("EncryptNew q1-scaled input C: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "input-c-q1-scale", "EncryptNew", ctC, c, 1, q1Scale, 1e-7, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	if result.FastBootstrapSelected && (!result.Checkpoints[0].State.C1Zero || !result.Checkpoints[1].State.C1Zero || !result.Checkpoints[2].State.C1Zero) {
		return result, errors.New("Fast public EncryptNew did not produce zero-c1 inputs")
	}
	if !result.FastBootstrapSelected && (result.Checkpoints[0].State.C1Zero || result.Checkpoints[1].State.C1Zero || result.Checkpoints[2].State.C1Zero) {
		return result, errors.New("genuine Standard public EncryptNew unexpectedly produced a zero-c1 input")
	}

	wantAdd := zip(a, b, func(x, y complex128) complex128 { return x + y })
	sum, err := evaluator.AddNew(ctA, ctB)
	if err != nil {
		return result, fmt.Errorf("public AddNew: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "add", "AddNew", sum, wantAdd, 1, defaultScale, addMaxError, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	lookupsAfterAdd := tracked.lookups

	wantProduct := zip(wantAdd, c, func(x, y complex128) complex128 { return x * y })
	productScale := sum.Scale.Mul(ctC.Scale)
	product, err := evaluator.MulRelinNew(sum, ctC)
	if err != nil {
		return result, fmt.Errorf("public ciphertext MulRelinNew: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "mul-relin", "MulRelinNew", product, wantProduct, 1, productScale, primitiveMaxError, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	lookupsAfterMul := tracked.lookups

	rescaleScale := productScale.Div(q1Scale)
	if rescaleScale.Cmp(defaultScale) != 0 {
		return result, fmt.Errorf("scale closure failed: MulRelin scale/q1=%s, frozen default scale=%s", rescaleScale.BigInt().String(), defaultScale.BigInt().String())
	}
	rescaled := ckks.NewCiphertext(btpParams.BootstrappingParameters, 1, 0)
	if err = evaluator.Rescale(product, rescaled); err != nil {
		return result, fmt.Errorf("public Rescale level 1 to 0: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "rescale", "Rescale", rescaled, wantProduct, 0, rescaleScale, primitiveMaxError, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	lookupsAfterRescale := tracked.lookups

	wantRotate := rotateLeft(wantProduct, rotation)
	rotated, err := evaluator.RotateNew(rescaled, rotation)
	if err != nil {
		return result, fmt.Errorf("public RotateNew(%d): %w", rotation, err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "rotate", "RotateNew", rotated, wantRotate, 0, defaultScale, primitiveMaxError, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	result.PrimitiveKeyLookups = primitiveLookupEvidence{
		EvaluatorConstruction: constructionLookups,
		Add:                   lookupsAfterAdd,
		MulRelin:              subtractLookups(lookupsAfterMul, lookupsAfterAdd),
		Rescale:               subtractLookups(lookupsAfterRescale, lookupsAfterMul),
		Rotate:                subtractLookups(tracked.lookups, lookupsAfterRescale),
	}
	if result.FastBootstrapSelected && (hasDirectKeyLookup(result.PrimitiveKeyLookups.Add) ||
		hasDirectKeyLookup(result.PrimitiveKeyLookups.MulRelin) ||
		hasDirectKeyLookup(result.PrimitiveKeyLookups.Rescale) ||
		hasDirectKeyLookup(result.PrimitiveKeyLookups.Rotate)) {
		return result, fmt.Errorf("Fast primitive chain unexpectedly retrieved keys: add=%+v mul=%+v rescale=%+v rotate=%+v", lookupsAfterAdd, lookupsAfterMul, lookupsAfterRescale, tracked.lookups)
	}
	if !result.FastBootstrapSelected && (result.PrimitiveKeyLookups.MulRelin.Relinearize < 1 || result.PrimitiveKeyLookups.Rotate.GaloisKey < 1) {
		return result, fmt.Errorf("Standard primitive chain did not consult native relin/rotation keys: mul=%+v rotate=%+v", result.PrimitiveKeyLookups.MulRelin, result.PrimitiveKeyLookups.Rotate)
	}
	result.Status = "primitive_preflight_passed"
	result.BootstrapInputFingerprint = hashState(result.Checkpoints[len(result.Checkpoints)-1].State)
	printPreflightEvent(result)
	if !hold {
		return result, nil
	}

	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, os.ErrClosed) {
		return result, fmt.Errorf("wait for Bootstrap authorization token: %w", err)
	}
	switch strings.TrimSpace(command) {
	case "STOP":
		result.Status = "primitive_preflight_passed_bootstrap_not_run"
		return result, nil
	case "BOOTSTRAP":
	default:
		result.Status = "primitive_preflight_passed_bootstrap_not_run"
		return result, fmt.Errorf("no Bootstrap invoked: expected exact BOOTSTRAP or STOP token, got %q", strings.TrimSpace(command))
	}

	result.BootstrapAttempted = true
	result.BootstrapCalls = 1
	if result.FastBootstrapSelected {
		result.BootstrapKeyAccessObservation = "source-traced Fast zero-secret public Bootstrap path; no Standard key-switch/relinearization path is selected; no per-call runtime key counter is exposed"
	} else {
		result.BootstrapKeyAccessObservation = "source-traced public Standard Bootstrap uses the generated EvaluationKeys; exact per-call lookups are not dynamically instrumented because the public key bundle embeds concrete MemEvaluationKeySet"
	}
	bootstrapOutput, err := bootstrapEval.Bootstrap(rotated)
	if err != nil {
		return result, fmt.Errorf("single public Bootstrap call: %w", err)
	}
	if err = addCheckpoint(&result, encoder, decryptor, "bootstrap", "bootstrapping.Evaluator.Bootstrap", bootstrapOutput, wantRotate, 1, defaultScale, bootstrapMaxError, result.FastBootstrapSelected, residual.N()); err != nil {
		return result, err
	}
	result.Status = "bootstrap_completed"
	return result, nil
}

func parametersFromConfig(cfg config) (ckks.Parameters, bootstrapping.Parameters, int, error) {
	if cfg.LogN != 13 || cfg.LogDefaultScale != defaultScaleLog2 || cfg.SecretHamming != 192 ||
		len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || len(cfg.QSlotsToCoeffs) != 3 ||
		len(cfg.QCoeffsToSlots) != 4 || len(cfg.P) != 5 || len(cfg.SlotsToCoeffsDFT) != 3 ||
		len(cfg.CoeffsToSlotsDFT) != 4 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 ||
		cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, errors.New("config is not the frozen LogN13 E32 profile")
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
		return nil, errors.New("DFT depth/scale lengths must match and be nonzero")
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

func encrypt(params ckks.Parameters, encoder *ckks.Encoder, sk *rlwe.SecretKey, values []complex128, scale rlwe.Scale, logSlots int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, 1)
	pt.Scale = scale
	pt.IsNTT = true
	pt.IsMontgomery = false
	pt.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(pt)
}

func addCheckpoint(result *runEvidence, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, id, operation string, ct *rlwe.Ciphertext, want []complex128, level int, scale rlwe.Scale, maxError float64, fast bool, n int) error {
	state, err := summarizeCiphertext(ct, fast, n)
	if err != nil {
		return fmt.Errorf("%s state: %w", id, err)
	}
	decoded := make([]complex128, len(want))
	if err = encoder.Decode(decryptor.DecryptNew(ct), decoded); err != nil {
		return fmt.Errorf("%s native DecryptNew/Decode: %w", id, err)
	}
	measurement := compareValues(want, decoded)
	if !measurement.Finite || measurement.MaxComplexError > maxError || state.Level != level || ct.Scale.Cmp(scale) != 0 || state.Degree != 1 || !state.IsNTT || state.IsMontgomery {
		return fmt.Errorf("%s failed CKKS oracle/state gate: max=%.12g gate=%.12g level=%d/%d scale=%s expected=%s", id, measurement.MaxComplexError, maxError, state.Level, level, ct.Scale.BigInt().String(), scale.BigInt().String())
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{
		ID: id, Operation: operation, State: state, Plaintext: measurement,
		DecodedSHA256: hashComplexValues(decoded), Decoded: toComplexValues(decoded),
	})
	return nil
}

func summarizeCiphertext(ct *rlwe.Ciphertext, fast bool, n int) (stateEvidence, error) {
	if ct == nil || ct.MetaData == nil || len(ct.Value) != 2 || ct.Degree() != 1 || ct.N() != n || ct.Level() < 0 {
		return stateEvidence{}, errors.New("ciphertext has invalid metadata, degree, component count, ring degree, or level")
	}
	active := ct.Level() + 1
	state := stateEvidence{
		Level: ct.Level(), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(), N: ct.N(), ComponentCount: len(ct.Value),
		ActiveRows: active, IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		RowsPerComponent: make([]int, len(ct.Value)), FullActiveQ: true,
	}
	for component := range ct.Value {
		rows := ct.Value[component].Coeffs
		state.RowsPerComponent[component] = len(rows)
		if len(rows) < active {
			return stateEvidence{}, fmt.Errorf("component %d has %d rows, needs %d active rows", component, len(rows), active)
		}
		for row := 0; row < active; row++ {
			if len(rows[row]) != n {
				state.FullActiveQ = false
				return stateEvidence{}, fmt.Errorf("component %d q%d row has %d coefficients, want %d", component, row, len(rows[row]), n)
			}
		}
	}
	if fast {
		wantRows := min(active, 4)
		if state.RowsPerComponent[0] != wantRows || state.RowsPerComponent[1] != wantRows {
			return stateEvidence{}, fmt.Errorf("Fast output allocated rows %v, want fixed Q-prefix width %d", state.RowsPerComponent, wantRows)
		}
	}
	state.C0ActiveRowSHA256 = hashRows(ct.Value[0].Coeffs[:active])
	state.C1ActiveRowSHA256 = hashRows(ct.Value[1].Coeffs[:active])
	state.C1CoefficientCount = active * n
	for row := 0; row < active; row++ {
		for _, coefficient := range ct.Value[1].Coeffs[row] {
			if coefficient != 0 {
				state.C1NonzeroCount++
			}
		}
	}
	state.C1Zero = state.C1NonzeroCount == 0
	if fast && !state.C1Zero {
		return stateEvidence{}, errors.New("Fast public ciphertext has nonzero authoritative c1")
	}
	return state, nil
}

func compareRunFiles(standardPath, fastPath string) (result combinedEvidence, resultErr error) {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return result, fmt.Errorf("read Standard run: %w", err)
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return result, fmt.Errorf("read Fast run: %w", err)
	}
	result = combinedEvidence{
		SchemaVersion: "fast-dropin-public-primitives-bootstrap-batch-017.combined.v1",
		CreatedUTC:    time.Now().UTC().Format(time.RFC3339Nano),
		PrimaryCommit: standard.PrimaryCommit, PrimaryDirty: standard.PrimaryDirty,
		FrontendSHA256: standard.FrontendSHA256, ConfigSHA256: standard.ConfigSHA256,
		WorkloadInputSHA256: standard.WorkloadInputSHA256, ProfileSHA256: standard.ProfileSHA256, QPSHA256: standard.QPSHA256,
		Standard: stripDecoded(standard), Fast: stripDecoded(fast),
		BootstrapCalls: standard.BootstrapCalls + fast.BootstrapCalls,
	}
	if err := validateComparableRuns(standard, fast); err != nil {
		result.Status, result.StopReason = "BATCH_BLOCKED_NEEDS_WEB_REVIEW", err.Error()
		return result, nil
	}
	passed := true
	for i := range standard.Checkpoints {
		sc, fc := standard.Checkpoints[i], fast.Checkpoints[i]
		paired := compareValues(fromComplexValues(sc.Decoded), fromComplexValues(fc.Decoded))
		checkpointPass := paired.Finite && paired.MaxComplexError <= pairedMaxError
		result.Paired = append(result.Paired, pairedCheckpoint{
			ID: sc.ID, Level: sc.State.Level, StandardScaleLog2: sc.State.ScaleLog2, FastScaleLog2: fc.State.ScaleLog2,
			StandardPlaintextRMSE: sc.Plaintext.ComplexRMSE, FastPlaintextRMSE: fc.Plaintext.ComplexRMSE,
			FastVsStandard: paired, Pass: checkpointPass,
		})
		passed = passed && checkpointPass
	}
	if standard.Status != "bootstrap_completed" || fast.Status != "bootstrap_completed" || standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 {
		passed = false
		result.StopReason = "both backends must complete exactly one Bootstrap after matched primitive preflights"
	}
	if passed {
		result.Status = "BATCH_COMPLETE_READY_FOR_WEB_REVIEW"
	} else {
		result.Status = "BATCH_BLOCKED_NEEDS_WEB_REVIEW"
		if result.StopReason == "" {
			result.StopReason = "one or more paired checkpoint maximum complex differences exceeded the fixed gate"
		}
	}
	return result, nil
}

func validateComparableRuns(standard, fast runEvidence) error {
	if standard.BackendCommit != standardCommit || fast.BackendCommit != fastCommit || standard.BackendDirty || fast.BackendDirty {
		return errors.New("backend commit/cleanliness does not match the frozen genuine Standard and Fast pins")
	}
	if standard.PrimaryCommit == "" || standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryDirty || fast.PrimaryDirty {
		return errors.New("paired runs do not share one clean Primary commit")
	}
	if standard.Status != "bootstrap_completed" || fast.Status != "bootstrap_completed" || standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 {
		return errors.New("both backend runs must contain exactly one completed Bootstrap call")
	}
	if standard.FrontendSHA256 == "" || standard.FrontendSHA256 != fast.FrontendSHA256 ||
		standard.ConfigSHA256 != canonicalConfigSHA256 || standard.ConfigSHA256 != fast.ConfigSHA256 ||
		standard.WorkloadInputSHA256 == "" || standard.WorkloadInputSHA256 != fast.WorkloadInputSHA256 ||
		standard.ProfileSHA256 != fast.ProfileSHA256 || standard.QPSHA256 != fast.QPSHA256 {
		return errors.New("frontend/config/workload/profile/QP provenance differs between backends")
	}
	if len(standard.Checkpoints) != len(fast.Checkpoints) || len(standard.Checkpoints) != 8 {
		return fmt.Errorf("checkpoint count mismatch: Standard=%d Fast=%d want=8", len(standard.Checkpoints), len(fast.Checkpoints))
	}
	for i := range standard.Checkpoints {
		sc, fc := standard.Checkpoints[i], fast.Checkpoints[i]
		if sc.ID != fc.ID || sc.State.Level != fc.State.Level || sc.State.Degree != fc.State.Degree ||
			!closeFloat(sc.State.ScaleLog2, fc.State.ScaleLog2, 1e-10) ||
			!sc.Plaintext.Finite || !fc.Plaintext.Finite || len(sc.Decoded) != standard.Slots || len(fc.Decoded) != fast.Slots {
			return fmt.Errorf("checkpoint %d is not matched or lacks finite full-slot decoded values", i)
		}
	}
	return nil
}

func printPreflightEvent(result runEvidence) {
	event := stripDecoded(result)
	data, _ := json.Marshal(struct {
		Event  string      `json:"event"`
		Result runEvidence `json:"result"`
	}{"primitive_preflight", event})
	fmt.Printf("PREFLIGHT_EVENT %s\n", data)
}

func stripDecoded(result runEvidence) runEvidence {
	result.Checkpoints = append([]checkpoint(nil), result.Checkpoints...)
	for i := range result.Checkpoints {
		result.Checkpoints[i].Decoded = nil
	}
	return result
}

func inspectKeys(keys *bootstrapping.EvaluationKeys, rotationGalEl uint64, evaluatorMaxP int) (result keyEvidence) {
	result.GaloisKeyCount = len(keys.GaloisKeys)
	result.EvaluatorMaxLevelP = evaluatorMaxP
	result.RelinearizationPresent = keys.RelinearizationKey != nil
	for _, key := range keys.GaloisKeys {
		if key == nil || key.GaloisElement != rotationGalEl {
			continue
		}
		result.RotationGaloisElement = rotationGalEl
		result.RotationKeyPresent = true
		result.RotationKeyLayout = keyLayout(key)
		result.RotationKeyLevelP = key.GadgetCiphertext.LevelP()
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

func compareValues(reference, actual []complex128) metric {
	result := metric{Finite: len(reference) == len(actual), SNRState: "finite"}
	if len(reference) != len(actual) || len(reference) == 0 {
		result.Finite, result.SNRState = false, "empty_or_length_mismatch"
		return result
	}
	var errorPower, signalPower float64
	for i := range reference {
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
	result.ComplexRMSE = math.Sqrt(errorPower / float64(len(reference)))
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

func deterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for i := range values {
		values[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return values
}

func deterministicInputs(slots int) (a, b, c []complex128) {
	a = deterministicInput(slots)
	b = make([]complex128, slots)
	c = append([]complex128(nil), a...)
	for i := range b {
		b[i] = complex(float64(i%19-9)/384, float64((5*i)%17-8)/768)
	}
	return
}

func zip(a, b []complex128, operation func(complex128, complex128) complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range result {
		result[i] = operation(a[i], b[i])
	}
	return result
}

func rotateLeft(values []complex128, k int) []complex128 {
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

func hashWorkload(a, b, c []complex128, q1 uint64) string {
	data := struct {
		A, B, C []complexValue
		CScale  string
	}{toComplexValues(a), toComplexValues(b), toComplexValues(c), new(big.Int).SetUint64(q1).String()}
	encoded, _ := json.Marshal(data)
	return hashBytes(encoded)
}

func hashComplexValues(values []complex128) string {
	h := sha256.New()
	var buffer [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(buffer[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(buffer[8:], math.Float64bits(imag(value)))
		_, _ = h.Write(buffer[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashRows(rows [][]uint64) []string {
	result := make([]string, len(rows))
	var buffer [8]byte
	for i, row := range rows {
		h := sha256.New()
		for _, coefficient := range row {
			binary.LittleEndian.PutUint64(buffer[:], coefficient)
			_, _ = h.Write(buffer[:])
		}
		result[i] = hex.EncodeToString(h.Sum(nil))
	}
	return result
}

func hashState(state stateEvidence) string {
	encoded, _ := json.Marshal(state)
	return hashBytes(encoded)
}

func hashBytes(value []byte) string {
	hash := sha256.Sum256(value)
	return hex.EncodeToString(hash[:])
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

func exactOrderedPrefix(prefix, full []uint64) bool {
	if len(prefix) > len(full) {
		return false
	}
	for i := range prefix {
		if prefix[i] != full[i] {
			return false
		}
	}
	return len(prefix) > 0
}

func equalSecretQ0(a, b *rlwe.SecretKey) bool {
	if a == nil || b == nil || len(a.Value.Q.Coeffs) == 0 || len(b.Value.Q.Coeffs) == 0 {
		return false
	}
	left, right := a.Value.Q.Coeffs[0], b.Value.Q.Coeffs[0]
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func subtractLookups(after, before keyLookups) keyLookups {
	return keyLookups{after.GaloisKey - before.GaloisKey, after.GaloisKeyList - before.GaloisKeyList, after.Relinearize - before.Relinearize}
}

func hasDirectKeyLookup(value keyLookups) bool {
	return value.GaloisKey != 0 || value.GaloisKeyList != 0 || value.Relinearize != 0
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func closeFloat(a, b, tolerance float64) bool { return math.Abs(a-b) <= tolerance }

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err = os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return nil
}
