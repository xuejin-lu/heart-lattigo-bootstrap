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
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	configPath              = "configs/bootstrap_config.logN13.json"
	frontendPath            = "tools/fast-dropin-highlevel-bootstrap-batch-018/main.go"
	standardCommit          = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	fastCommit              = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac"
	canonicalConfigSHA256   = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"
	canonicalInputSHA256    = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
	canonicalQPSHA256       = "1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b"
	defaultScaleLog2        = 45
	bootstrapMaxError       = 1e-6
	rotation                = 1
	qPrefixCap              = 4
	runSchema               = "fast-dropin-highlevel-bootstrap-batch-018.run.v1"
	combinedSchema          = "fast-dropin-highlevel-bootstrap-batch-018.combined.v1"
	minimumBootstrapMessage = "bootstrap"
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
	Finite      bool     `json:"finite"`
	ComplexRMSE float64  `json:"complex_rmse"`
	MaxComplex  float64  `json:"max_complex_error"`
	SNRDB       *float64 `json:"snr_db,omitempty"`
	SNRState    string   `json:"snr_state"`
}

type state struct {
	Level                  int        `json:"level"`
	Degree                 int        `json:"degree"`
	Scale                  string     `json:"scale_integer"`
	ScaleLog2              float64    `json:"scale_log2"`
	N                      int        `json:"n"`
	NTT                    bool       `json:"is_ntt"`
	Montgomery             bool       `json:"is_montgomery"`
	AuthoritativeRows      int        `json:"qprefix_authoritative_rows"`
	RowLengths             [][]int    `json:"component_row_lengths"`
	AuthoritativeRowHashes [][]string `json:"component_qprefix_row_sha256"`
	C1Zero                 bool       `json:"c1_zero_on_authoritative_rows"`
}

type capacityEvidence struct {
	Rows             int      `json:"rows"`
	PrefixProduct    string   `json:"prefix_product"`
	IndependentBound []string `json:"independent_component_bounds"`
	ObserverMaxAbs   []string `json:"observer_centered_max_abs,omitempty"`
	StrictFit        bool     `json:"strict_2B_less_than_prefix"`
	ObserverInvoked  bool     `json:"existing_fast_capacity_observer_invoked"`
}

type checkpoint struct {
	ID       string            `json:"id"`
	API      string            `json:"public_api"`
	State    state             `json:"state"`
	Capacity *capacityEvidence `json:"capacity,omitempty"`
}

type runEvidence struct {
	SchemaVersion        string         `json:"schema_version"`
	Status               string         `json:"status"`
	Error                string         `json:"error,omitempty"`
	CreatedUTC           string         `json:"created_utc"`
	GoVersion            string         `json:"go_version"`
	OS                   string         `json:"os"`
	Architecture         string         `json:"architecture"`
	PrimaryCommit        string         `json:"primary_commit"`
	PrimaryDirty         bool           `json:"primary_dirty"`
	BackendCommit        string         `json:"backend_commit"`
	BackendRef           string         `json:"backend_ref"`
	BackendDirty         bool           `json:"backend_dirty"`
	FastZeroSecret       bool           `json:"fast_zero_secret"`
	PublicDispatch       string         `json:"public_bootstrap_dispatch"`
	ConfigSHA256         string         `json:"config_sha256"`
	FrontendSHA256       string         `json:"shared_frontend_sha256"`
	RunnerBundleSHA256   string         `json:"runner_bundle_sha256"`
	InputSHA256          string         `json:"canonical_input_sha256"`
	WorkloadSHA256       string         `json:"workload_sha256"`
	ProfileSHA256        string         `json:"effective_profile_sha256"`
	QPSHA256             string         `json:"effective_qp_sha256"`
	Q                    []uint64       `json:"q_primes"`
	P                    []uint64       `json:"p_primes"`
	LogN                 int            `json:"log_n"`
	LogSlots             int            `json:"log_slots"`
	Slots                int            `json:"slots"`
	ResidualMaxLevel     int            `json:"residual_max_level"`
	BootstrapMaxLevel    int            `json:"bootstrap_max_level"`
	E                    int            `json:"ephemeral_secret_weight"`
	DefaultScale         string         `json:"default_scale_integer"`
	KeysetIdentity       string         `json:"keyset_identity"`
	ResidualSecretLevel  int            `json:"residual_secret_level"`
	BootstrapSecretLevel int            `json:"bootstrap_secret_level"`
	SameLowSecret        bool           `json:"bootstrap_secret_matches_residual_q0_q1"`
	ResidualSecretHash   []string       `json:"residual_secret_q0_q1_sha256"`
	BootstrapSecretHash  []string       `json:"bootstrap_secret_q0_q1_sha256"`
	BootstrapCalls       int            `json:"bootstrap_calls"`
	BootstrapAttempted   bool           `json:"bootstrap_attempted"`
	BootstrapInputLevel  int            `json:"bootstrap_input_level"`
	BootstrapInputScale  string         `json:"bootstrap_input_scale_integer"`
	Checkpoints          []checkpoint   `json:"checkpoints"`
	PreBootstrapOracle   metric         `json:"pre_bootstrap_cleartext_oracle"`
	PreBootstrapSHA256   string         `json:"pre_bootstrap_decoded_sha256"`
	BootstrapOracle      *metric        `json:"bootstrap_cleartext_oracle,omitempty"`
	BootstrapSHA256      string         `json:"bootstrap_decoded_sha256,omitempty"`
	PreBootstrapDecoded  []complexValue `json:"pre_bootstrap_decoded_values,omitempty"`
	BootstrapDecoded     []complexValue `json:"bootstrap_decoded_values,omitempty"`
}

type pairedMetric struct {
	RMSE float64 `json:"rmse"`
	Max  float64 `json:"max_complex_error"`
	Pass bool    `json:"pass"`
}

type runSummary struct {
	BackendCommit      string       `json:"backend_commit"`
	BackendRef         string       `json:"backend_ref"`
	BackendDirty       bool         `json:"backend_dirty"`
	FastZeroSecret     bool         `json:"fast_zero_secret"`
	PublicDispatch     string       `json:"public_bootstrap_dispatch"`
	BootstrapCalls     int          `json:"bootstrap_calls"`
	BootstrapAttempted bool         `json:"bootstrap_attempted"`
	KeysetIdentity     string       `json:"keyset_identity"`
	SameLowSecret      bool         `json:"bootstrap_secret_matches_residual_q0_q1"`
	Checkpoints        []checkpoint `json:"checkpoints"`
	PreOracle          metric       `json:"pre_bootstrap_cleartext_oracle"`
	PreSHA256          string       `json:"pre_bootstrap_decoded_sha256"`
	BootstrapOracle    *metric      `json:"bootstrap_cleartext_oracle,omitempty"`
	BootstrapSHA256    string       `json:"bootstrap_decoded_sha256,omitempty"`
}

type combinedEvidence struct {
	SchemaVersion      string        `json:"schema_version"`
	Status             string        `json:"status"`
	CreatedUTC         string        `json:"created_utc"`
	PrimaryCommit      string        `json:"primary_code_commit"`
	PrimaryDirty       bool          `json:"primary_dirty_during_runs"`
	GoVersion          string        `json:"go_version"`
	OS                 string        `json:"os"`
	Architecture       string        `json:"architecture"`
	ConfigSHA256       string        `json:"config_sha256"`
	FrontendSHA256     string        `json:"shared_frontend_sha256"`
	RunnerBundleSHA256 string        `json:"runner_bundle_sha256"`
	InputSHA256        string        `json:"canonical_input_sha256"`
	WorkloadSHA256     string        `json:"workload_sha256"`
	ProfileSHA256      string        `json:"effective_profile_sha256"`
	QPSHA256           string        `json:"effective_qp_sha256"`
	Q                  []uint64      `json:"q_primes"`
	P                  []uint64      `json:"p_primes"`
	ResidualMaxLevel   int           `json:"residual_max_level"`
	BootstrapMaxLevel  int           `json:"bootstrap_max_level"`
	E                  int           `json:"ephemeral_secret_weight"`
	Slots              int           `json:"slots"`
	PreBootstrapPair   pairedMetric  `json:"pre_bootstrap_fast_vs_standard"`
	BootstrapPair      *pairedMetric `json:"bootstrap_fast_vs_standard,omitempty"`
	Standard           runSummary    `json:"standard"`
	Fast               runSummary    `json:"fast"`
	BootstrapCalls     int           `json:"bootstrap_calls_total"`
	FixedMaxErrorGate  float64       `json:"fixed_max_error_gate"`
}

type capacityAudit interface {
	Observe(name string, ct *rlwe.Ciphertext, rows int) (rowsSeen int, prefixProduct string, maxAbs []string, strictFit bool, err error)
}

type execution struct {
	result      runEvidence
	baptEval    *bootstrapping.Evaluator
	bootstrapIn *rlwe.Ciphertext
	decryptor   *rlwe.Decryptor
	encoder     *ckks.Encoder
	expected    []complex128
}

func main() {
	if err := runCLI(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() error {
	out := flag.String("out", "", "raw run JSON written outside the repository")
	primaryCommit := flag.String("primary-commit", "", "clean Primary code commit")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty at run start")
	backendCommit := flag.String("backend-commit", "", "pinned Lattigo commit")
	backendRef := flag.String("backend-ref", "", "pinned Lattigo ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether backend was dirty")
	combineStage := flag.String("combine-stage", "", "combine paired runs: preflight or final")
	standardRun := flag.String("standard-run", "", "Standard raw run JSON")
	fastRun := flag.String("fast-run", "", "Fast raw run JSON")
	flag.Parse()
	if *combineStage != "" {
		if *out == "" || *standardRun == "" || *fastRun == "" {
			return errors.New("combine mode requires -out, -standard-run and -fast-run")
		}
		return combineRuns(*combineStage, *standardRun, *fastRun, *out)
	}
	if *out == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return errors.New("run mode requires -out, -primary-commit, -backend-commit and -backend-ref")
	}
	result, err := prepareRun(*primaryCommit, *primaryDirty, *backendCommit, *backendRef, *backendDirty)
	if err != nil {
		result.result.Status = "preflight_failed"
		result.result.Error = err.Error()
		_ = writeJSON(*out, result.result)
		return err
	}
	result.result.Status = "preflight_passed_waiting_for_bootstrap"
	if err = writeJSON(*out, result.result); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "BATCH018_PREFLIGHT_READY — send exactly `bootstrap` to spend this backend's single Bootstrap allowance, or anything else to abort")
	line, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
	if readErr != nil || strings.TrimSpace(line) != minimumBootstrapMessage {
		result.result.Status = "preflight_passed_bootstrap_not_started"
		if readErr != nil {
			result.result.Error = "Bootstrap gate ended before explicit local release: " + readErr.Error()
		}
		if writeErr := writeJSON(*out, result.result); writeErr != nil {
			return writeErr
		}
		if readErr != nil && !errors.Is(readErr, os.ErrClosed) && !strings.Contains(readErr.Error(), "EOF") {
			return readErr
		}
		return nil
	}
	// Persist consumed budget before entering the one-shot public Bootstrap call.
	result.result.BootstrapCalls = 1
	result.result.BootstrapAttempted = true
	result.result.Status = "bootstrap_call_started"
	if err = writeJSON(*out, result.result); err != nil {
		return fmt.Errorf("persist consumed Bootstrap budget before call: %w", err)
	}
	bootstrapOut, callErr := result.baptEval.Bootstrap(result.bootstrapIn)
	if callErr != nil {
		result.result.Status = "bootstrap_failed_after_single_call"
		result.result.Error = callErr.Error()
		_ = writeJSON(*out, result.result)
		return fmt.Errorf("single public Bootstrap call: %w", callErr)
	}
	if err = appendBootstrapCheckpoint(result, bootstrapOut); err != nil {
		result.result.Status = "bootstrap_completed_output_gate_failed"
		result.result.Error = err.Error()
		_ = writeJSON(*out, result.result)
		return err
	}
	result.result.Status = "bootstrap_completed"
	return writeJSON(*out, result.result)
}

func prepareRun(primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty bool) (*execution, error) {
	result := runEvidence{
		SchemaVersion: runSchema, CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		PrimaryCommit: primaryCommit, PrimaryDirty: primaryDirty,
		BackendCommit: backendCommit, BackendRef: backendRef, BackendDirty: backendDirty,
		FastZeroSecret: isFastBackend(), BootstrapCalls: 0, BootstrapAttempted: false,
		KeysetIdentity: "same public GenEvaluationKeys result passed to both ckks.NewEvaluator and bootstrapping.NewEvaluator",
	}
	result.DefaultScale = new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2).String()
	result.LogN, result.LogSlots, result.Slots = 13, 12, 1<<12
	result.E = 32
	result.InputSHA256 = canonicalInputSHA256
	if primaryDirty || backendDirty {
		return &execution{result: result}, errors.New("Primary and backend must both be clean for the paired preflight")
	}
	if isFastBackend() {
		if backendCommit != fastCommit || backendRef != "fast-qprefix" {
			return &execution{result: result}, fmt.Errorf("Fast pin mismatch: got %s@%s", backendRef, backendCommit)
		}
	} else if backendCommit != standardCommit || backendRef != standardCommit {
		return &execution{result: result}, fmt.Errorf("Standard pin mismatch: got %s@%s", backendRef, backendCommit)
	}

	frontendBytes, err := os.ReadFile(frontendPath)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("read shared frontend: %w", err)
	}
	result.FrontendSHA256 = hashBytes(frontendBytes)
	result.RunnerBundleSHA256, err = runnerBundleHash()
	if err != nil {
		return &execution{result: result}, err
	}
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("read frozen config: %w", err)
	}
	result.ConfigSHA256 = hashBytes(configBytes)
	if result.ConfigSHA256 != canonicalConfigSHA256 {
		return &execution{result: result}, fmt.Errorf("canonical config SHA mismatch: %s", result.ConfigSHA256)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return &execution{result: result}, fmt.Errorf("decode frozen config: %w", err)
	}
	residual, btp, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return &execution{result: result}, err
	}
	full := btp.BootstrappingParameters
	if residual.MaxLevel() != 1 || full.MaxLevel() != 16 || logSlots != 12 || btp.EphemeralSecretWeight != 32 || full.LogDefaultScale() != defaultScaleLog2 {
		return &execution{result: result}, fmt.Errorf("frozen profile mismatch: residual=%d full=%d slots=%d E=%d scale=%d", residual.MaxLevel(), full.MaxLevel(), logSlots, btp.EphemeralSecretWeight, full.LogDefaultScale())
	}
	if len(residual.Q()) != 2 || !slices.Equal(residual.Q(), full.Q()[:2]) {
		return &execution{result: result}, errors.New("residual Q is not the exact q0/q1 prefix of the full Bootstrap parameters")
	}
	result.Q, result.P = full.Q(), full.P()
	result.ResidualMaxLevel, result.BootstrapMaxLevel = residual.MaxLevel(), full.MaxLevel()
	result.LogSlots, result.Slots = logSlots, 1<<logSlots
	qpJSON, _ := json.Marshal(struct{ Q, P []uint64 }{Q: result.Q, P: result.P})
	result.QPSHA256 = hashBytes(qpJSON)
	if result.QPSHA256 != canonicalQPSHA256 {
		return &execution{result: result}, fmt.Errorf("canonical generated Q/P chain mismatch: %s", result.QPSHA256)
	}
	profileJSON, _ := json.Marshal(struct {
		LogN, DefaultScaleLog2, LogSlots, ResidualMaxLevel, BootstrapMaxLevel, E int
		Q, P                                                                     []uint64
	}{cfg.LogN, cfg.LogDefaultScale, logSlots, residual.MaxLevel(), full.MaxLevel(), btp.EphemeralSecretWeight, result.Q, result.P})
	result.ProfileSHA256 = hashBytes(profileJSON)

	a, b := deterministicInputs(result.Slots)
	if hashComplexValues(deterministicInput(result.Slots)) != canonicalInputSHA256 {
		return &execution{result: result}, errors.New("canonical 017 original input hash mismatch")
	}
	workloadJSON, _ := json.Marshal(struct{ A, B []complexValue }{toValues(a), toValues(b)})
	result.WorkloadSHA256 = hashBytes(workloadJSON)
	boundA := encodedInputBound(a, math.Exp2(defaultScaleLog2))
	boundB := encodedInputBound(b, math.Exp2(defaultScaleLog2))
	boundSum := new(big.Int).Add(new(big.Int).Set(boundA), boundB)
	q0Capacity := new(big.Int).SetUint64(result.Q[0])
	if !strictCapacity(boundA, q0Capacity) || !strictCapacity(boundB, q0Capacity) || !strictCapacity(boundSum, q0Capacity) {
		return &execution{result: result}, errors.New("independent encoder/Add bound does not fit the target q0")
	}

	keygen := rlwe.NewKeyGenerator(residual)
	sk := keygen.GenSecretKeyNew()
	keys, bootstrapSecret, err := btp.GenEvaluationKeys(sk)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("public GenEvaluationKeys: %w", err)
	}
	if keys == nil || keys.MemEvaluationKeySet == nil || bootstrapSecret == nil {
		return &execution{result: result}, errors.New("public GenEvaluationKeys returned incomplete key material")
	}
	result.ResidualSecretLevel, result.BootstrapSecretLevel = sk.LevelQ(), bootstrapSecret.LevelQ()
	result.ResidualSecretHash = secretRowHashes(sk, 2)
	result.BootstrapSecretHash = secretRowHashes(bootstrapSecret, 2)
	result.SameLowSecret = slices.Equal(result.ResidualSecretHash, result.BootstrapSecretHash)
	if !result.SameLowSecret || bootstrapSecret.LevelQ() != full.MaxLevel() {
		return &execution{result: result}, errors.New("extended Bootstrap secret does not preserve residual q0/q1 or full-Q level")
	}
	result.PublicDispatch = "standard"
	if isFastBackend() {
		result.PublicDispatch = "Fast public bootstrap adapter selected"
	}
	encoder := ckks.NewEncoder(full)
	residualEncoder := ckks.NewEncoder(residual)
	decryptor := rlwe.NewDecryptor(residual, sk)
	evaluator := ckks.NewEvaluator(full, keys.MemEvaluationKeySet)
	bootstrapEval, err := bootstrapping.NewEvaluator(btp, keys)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("public bootstrapping.NewEvaluator: %w", err)
	}
	if bootstrapEval == nil {
		return &execution{result: result}, errors.New("public bootstrapping.NewEvaluator returned nil")
	}
	dispatchFast := bootstrapEval.Evaluator == nil
	if dispatchFast != isFastBackend() {
		return &execution{result: result}, errors.New("public Bootstrap dispatch did not match the pinned backend capability")
	}
	if isFastBackend() {
		result.PublicDispatch = "bootstrapping.NewEvaluator selected Fast through public GenEvaluationKeys capability"
	}
	result.KeysetIdentity = "one btpParams.GenEvaluationKeys result; its MemEvaluationKeySet is passed to ckks.NewEvaluator and the containing same keys object to bootstrapping.NewEvaluator"

	scale := rlwe.NewScale(math.Exp2(defaultScaleLog2))
	ctA, err := encryptAtLevel(full, encoder, bootstrapSecret, a, scale, 5, logSlots)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("public EncryptNew(A) at full-profile Level 5: %w", err)
	}
	ctB, err := encryptAtLevel(full, encoder, bootstrapSecret, b, scale, 5, logSlots)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("public EncryptNew(B) at full-profile Level 5: %w", err)
	}
	expectedAdd := addValues(a, b)
	expected := rotateLeft(expectedAdd, rotation)
	audit := makeCapacityAudit(full)
	if isFastBackend() != (audit != nil) {
		return &execution{result: result}, errors.New("Fast capacity adapter does not match backend build")
	}
	for _, item := range []struct {
		id    string
		ct    *rlwe.Ciphertext
		bound *big.Int
	}{{"encrypt_a_level5", ctA, boundA}, {"encrypt_b_level5", ctB, boundB}} {
		cap, capErr := capacityCheckpoint(audit, result.Q, item.id, item.ct, item.bound)
		if capErr != nil {
			return &execution{result: result}, capErr
		}
		st, stErr := summarizeState(item.ct, full.N())
		if stErr != nil {
			return &execution{result: result}, stErr
		}
		if isFastBackend() && !st.C1Zero {
			return &execution{result: result}, fmt.Errorf("Fast %s did not preserve zero c1 on authoritative q0..q3", item.id)
		}
		result.Checkpoints = append(result.Checkpoints, checkpoint{ID: item.id, API: "ckks.NewEncryptor(...).EncryptNew", State: st, Capacity: cap})
	}
	add, err := evaluator.AddNew(ctA, ctB)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("public AddNew at Level 5: %w", err)
	}
	if add.Level() != 5 || !add.Scale.Equal(scale) {
		return &execution{result: result}, errors.New("public AddNew changed the frozen Level/Scale")
	}
	addCap, err := capacityCheckpoint(audit, result.Q, "add_level5", add, boundSum)
	if err != nil {
		return &execution{result: result}, err
	}
	addState, err := summarizeState(add, full.N())
	if err != nil {
		return &execution{result: result}, err
	}
	if isFastBackend() && !compactAtPrefix(addState, full.N()) {
		return &execution{result: result}, errors.New("Fast AddNew did not leave q4+ dormant at Level 5")
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{ID: "add_level5", API: "ckks.Evaluator.AddNew", State: addState, Capacity: addCap})

	rotated, err := evaluator.RotateNew(add, rotation)
	if err != nil {
		return &execution{result: result}, fmt.Errorf("public RotateNew(%d) at Level 5: %w", rotation, err)
	}
	if rotated.Level() != 5 || !rotated.Scale.Equal(scale) {
		return &execution{result: result}, errors.New("public RotateNew changed the frozen Level/Scale")
	}
	rotateCap, err := capacityCheckpoint(audit, result.Q, "rotate_level5", rotated, boundSum)
	if err != nil {
		return &execution{result: result}, err
	}
	rotateState, err := summarizeState(rotated, full.N())
	if err != nil {
		return &execution{result: result}, err
	}
	if isFastBackend() && !compactAtPrefix(rotateState, full.N()) {
		return &execution{result: result}, errors.New("Fast RotateNew did not leave q4+ dormant at Level 5")
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{ID: "rotate_level5", API: "ckks.Evaluator.RotateNew(1)", State: rotateState, Capacity: rotateCap})

	// CKKS DropLevel is an ordinary public modulus projection with no rescale.
	// The precomputed independent bound proves the same centered lift fits q0.
	dropTarget := evaluator.DropLevelNew(rotated, rotated.Level())
	if dropTarget.Level() != 0 || !dropTarget.Scale.Equal(scale) {
		return &execution{result: result}, errors.New("public DropLevelNew did not produce the exact Level-0 / Scale-2^45 target")
	}
	dropCap, err := capacityCheckpoint(audit, result.Q, "drop_to_residual_level0", dropTarget, boundSum)
	if err != nil {
		return &execution{result: result}, err
	}
	dropState, err := summarizeState(dropTarget, residual.N())
	if err != nil {
		return &execution{result: result}, err
	}
	if isFastBackend() && !compactAtPrefix(dropState, residual.N()) {
		return &execution{result: result}, errors.New("Fast DropLevelNew output is not an exact q0-only public Level-0 layout")
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{ID: "drop_to_residual_level0", API: "ckks.Evaluator.DropLevelNew(5)", State: dropState, Capacity: dropCap})
	if dropTarget.Level() != 0 || dropTarget.Value[0].Level() != 0 || dropTarget.Scale.Cmp(scale) != 0 {
		return &execution{result: result}, errors.New("Bootstrap input does not match canonical residual q0 and Scale contract")
	}
	decoded := make([]complex128, result.Slots)
	if err = residualEncoder.Decode(decryptor.DecryptNew(dropTarget), decoded); err != nil {
		return &execution{result: result}, fmt.Errorf("native Level-0 DecryptNew/Decode before Bootstrap: %w", err)
	}
	oracle := compareValues(expected, decoded)
	if !oracle.Finite || oracle.MaxComplex > bootstrapMaxError {
		return &execution{result: result}, fmt.Errorf("pre-Bootstrap cleartext oracle failed: max=%g gate=%g", oracle.MaxComplex, bootstrapMaxError)
	}
	result.PreBootstrapOracle = oracle
	result.PreBootstrapSHA256 = hashComplexValues(decoded)
	result.PreBootstrapDecoded = toValues(decoded)
	result.BootstrapInputLevel = dropTarget.Level()
	result.BootstrapInputScale = dropTarget.Scale.BigInt().String()
	result.Status = "preflight_passed"
	return &execution{result: result, baptEval: bootstrapEval, bootstrapIn: dropTarget, decryptor: decryptor, encoder: residualEncoder, expected: expected}, nil
}

func appendBootstrapCheckpoint(run *execution, ct *rlwe.Ciphertext) error {
	if ct == nil {
		return errors.New("public Bootstrap returned nil ciphertext")
	}
	var cap *capacityEvidence
	st, err := summarizeState(ct, 1<<13)
	if err != nil {
		return err
	}
	if ct.Level() != 1 || ct.Scale.Cmp(rlwe.NewScale(math.Exp2(defaultScaleLog2))) != 0 {
		return fmt.Errorf("Bootstrap output Level/Scale mismatch: level=%d scale=%s", ct.Level(), ct.Scale.BigInt().String())
	}
	if isFastBackend() && (!compactAtPrefix(st, 1<<13) || !st.C1Zero) {
		return errors.New("Fast Bootstrap output violated residual Q-prefix row authority or zero-c1 semantics")
	}
	decoded := make([]complex128, len(run.expected))
	if err = run.encoder.Decode(run.decryptor.DecryptNew(ct), decoded); err != nil {
		return fmt.Errorf("native post-Bootstrap DecryptNew/Decode: %w", err)
	}
	oracle := compareValues(run.expected, decoded)
	if !oracle.Finite || oracle.MaxComplex > bootstrapMaxError {
		return fmt.Errorf("post-Bootstrap cleartext oracle failed: max=%g gate=%g", oracle.MaxComplex, bootstrapMaxError)
	}
	run.result.BootstrapOracle = &oracle
	run.result.BootstrapSHA256 = hashComplexValues(decoded)
	run.result.BootstrapDecoded = toValues(decoded)
	run.result.Checkpoints = append(run.result.Checkpoints, checkpoint{ID: "public_bootstrap_output", API: "bootstrapping.Evaluator.Bootstrap", State: st, Capacity: cap})
	return nil
}

func parametersFromConfig(cfg config) (ckks.Parameters, bootstrapping.Parameters, int, error) {
	if cfg.LogN != 13 || cfg.LogDefaultScale != defaultScaleLog2 || cfg.SecretHamming != 192 ||
		!slices.Equal(cfg.Q0, []int{55}) || !slices.Equal(cfg.QSlotsToCoeffs, []int{39, 39, 39}) ||
		!slices.Equal(cfg.QCoeffsToSlots, []int{56, 56, 56, 56}) || len(cfg.P) != 5 ||
		!slices.Equal(cfg.SlotsToCoeffsDFT, []int{1, 1, 1}) || !slices.Equal(cfg.CoeffsToSlotsDFT, []int{1, 1, 1, 1}) ||
		cfg.LogSlots != -1 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 || cfg.Mod1DoubleAngle != 3 ||
		cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 || cfg.Repetitions != 3 || cfg.Warmup != 1 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, errors.New("config is not the frozen canonical LogN13 E32 profile")
	}
	for _, v := range cfg.P {
		if v != 61 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected P prime bit-size %d", v)
		}
	}
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: cfg.LogN, LogQ: []int{cfg.Q0[0], cfg.QSlotsToCoeffs[0]}, LogDefaultScale: cfg.LogDefaultScale, Xs: ring.Ternary{H: cfg.SecretHamming}})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("construct residual parameters: %w", err)
	}
	logSlots := residual.LogMaxSlots()
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
	messageRatio, invDegree, zero := cfg.LogMessageRatio, cfg.Mod1InvDegree, 0
	btp, err := bootstrapping.NewParametersFromLiteral(residual, bootstrapping.ParametersLiteral{
		LogN: &logN, LogP: slices.Clone(cfg.P), Xs: ring.Ternary{H: cfg.SecretHamming}, LogSlots: &logSlots,
		CoeffsToSlotsFactorizationDepthAndLogScales: c2s, SlotsToCoeffsFactorizationDepthAndLogScales: s2c,
		EvalModLogScale: &logScale, EphemeralSecretWeight: &zero, Mod1Type: mod1.CosDiscrete,
		LogMessageRatio: &messageRatio, K: &k, Mod1Degree: &degree, DoubleAngle: &doubleAngle, Mod1InvDegree: &invDegree,
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("construct canonical Bootstrap parameters: %w", err)
	}
	btp.CircuitOrder, btp.ResidualParameters, btp.EphemeralSecretWeight = bootstrapping.ModUpThenEncode, residual, 32
	return residual, btp, logSlots, nil
}

func factorization(depths, scales []int) ([][]int, error) {
	if len(depths) == 0 || len(depths) != len(scales) {
		return nil, errors.New("DFT depth/scale lengths must match and be nonzero")
	}
	result := make([][]int, len(depths))
	for i, depth := range depths {
		if depth <= 0 || scales[i] <= 0 {
			return nil, fmt.Errorf("invalid DFT factorization entry %d", i)
		}
		result[i] = make([]int, depth)
		for j := range result[i] {
			result[i][j] = scales[i]
		}
	}
	return result, nil
}

func encryptAtLevel(params ckks.Parameters, encoder *ckks.Encoder, sk *rlwe.SecretKey, values []complex128, scale rlwe.Scale, level, logSlots int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, level)
	pt.Scale = scale
	pt.IsNTT = true
	pt.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(pt)
}

func capacityCheckpoint(audit capacityAudit, q []uint64, name string, ct *rlwe.Ciphertext, bound *big.Int) (*capacityEvidence, error) {
	if audit == nil {
		// The small bound describes the encoded Fast zero-secret polynomial, not
		// the randomized components of a genuine Standard RLWE ciphertext.
		return nil, nil
	}
	rows := qPrefixWidth(ct.Level())
	product := prefixProduct(q, rows)
	if !strictCapacity(bound, product) {
		return nil, fmt.Errorf("%s independent B=%s violates strict Q-prefix capacity for rows=%d product=%s", name, bound, rows, product)
	}
	evidence := &capacityEvidence{Rows: rows, PrefixProduct: product.String(), IndependentBound: []string{bound.String(), "0"}, StrictFit: true}
	seenRows, seenProduct, observed, strict, err := audit.Observe(name, ct, rows)
	if err != nil {
		return nil, fmt.Errorf("Fast capacity observer %s: %w", name, err)
	}
	if seenRows != rows || seenProduct != product.String() || !strict || len(observed) < 2 {
		return nil, fmt.Errorf("Fast capacity observer %s returned inconsistent rows/product/strict result", name)
	}
	for i, maxAbs := range observed[:2] {
		got, ok := new(big.Int).SetString(maxAbs, 10)
		limit := bound
		if i == 1 {
			limit = big.NewInt(0)
		}
		if !ok || got.Cmp(limit) > 0 {
			return nil, fmt.Errorf("Fast capacity observer %s component %d max=%s exceeds independent B=%s", name, i, maxAbs, limit)
		}
	}
	evidence.ObserverMaxAbs, evidence.ObserverInvoked = observed[:2], true
	return evidence, nil
}

func summarizeState(ct *rlwe.Ciphertext, n int) (state, error) {
	if ct == nil || ct.MetaData == nil || len(ct.Value) != 2 || ct.Degree() != 1 || ct.N() != n || ct.Level() < 0 {
		return state{}, errors.New("invalid public ciphertext shape or metadata")
	}
	rows := qPrefixWidth(ct.Level())
	out := state{Level: ct.Level(), Degree: ct.Degree(), Scale: ct.Scale.BigInt().String(), ScaleLog2: ct.Scale.Log2(), N: n, NTT: ct.IsNTT, Montgomery: ct.IsMontgomery, AuthoritativeRows: rows}
	out.RowLengths = make([][]int, len(ct.Value))
	out.AuthoritativeRowHashes = make([][]string, len(ct.Value))
	out.C1Zero = true
	for component := range ct.Value {
		polyRows := ct.Value[component].Coeffs
		out.RowLengths[component] = make([]int, len(polyRows))
		if len(polyRows) < rows {
			return state{}, fmt.Errorf("component %d has %d rows; Q-prefix needs %d", component, len(polyRows), rows)
		}
		for row, coefficients := range polyRows {
			out.RowLengths[component][row] = len(coefficients)
		}
		out.AuthoritativeRowHashes[component] = hashRows(polyRows[:rows])
		for row := 0; row < rows; row++ {
			if len(polyRows[row]) != n {
				return state{}, fmt.Errorf("component %d q%d backing=%d, want N=%d", component, row, len(polyRows[row]), n)
			}
			if component == 1 {
				for _, coefficient := range polyRows[row] {
					if coefficient != 0 {
						out.C1Zero = false
					}
				}
			}
		}
	}
	return out, nil
}

func compactAtPrefix(st state, n int) bool {
	for _, rows := range st.RowLengths {
		if len(rows) < st.AuthoritativeRows {
			return false
		}
		for i := 0; i < st.AuthoritativeRows; i++ {
			if rows[i] != n {
				return false
			}
		}
		for i := st.AuthoritativeRows; i < len(rows); i++ {
			if rows[i] != 0 {
				return false
			}
		}
	}
	return true
}

func combineRuns(stage, standardPath, fastPath, outPath string) error {
	if stage != "preflight" && stage != "final" {
		return fmt.Errorf("unsupported combine stage %q", stage)
	}
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return fmt.Errorf("read Standard run: %w", err)
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return fmt.Errorf("read Fast run: %w", err)
	}
	if err := validateRunPair(standard, fast); err != nil {
		return err
	}
	if stage == "preflight" {
		if standard.Status != "preflight_passed_waiting_for_bootstrap" || fast.Status != "preflight_passed_waiting_for_bootstrap" || standard.BootstrapCalls != 0 || fast.BootstrapCalls != 0 {
			return errors.New("both matched preflight runs must pass and remain before Bootstrap")
		}
	} else if standard.Status != "bootstrap_completed" || fast.Status != "bootstrap_completed" || standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 || !standard.BootstrapAttempted || !fast.BootstrapAttempted {
		return errors.New("final pair requires exactly one completed public Bootstrap call per backend")
	}
	prePair, err := pairedValues(fromValues(standard.PreBootstrapDecoded), fromValues(fast.PreBootstrapDecoded))
	if err != nil || !prePair.Pass {
		return fmt.Errorf("pre-Bootstrap Fast-vs-Standard gate failed: %+v (%v)", prePair, err)
	}
	out := combinedEvidence{
		SchemaVersion: combinedSchema, Status: "BATCH018_PREFLIGHT_PAIR_PASSED", CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano),
		PrimaryCommit: standard.PrimaryCommit, PrimaryDirty: standard.PrimaryDirty, GoVersion: standard.GoVersion, OS: standard.OS, Architecture: standard.Architecture,
		ConfigSHA256: standard.ConfigSHA256, FrontendSHA256: standard.FrontendSHA256, RunnerBundleSHA256: standard.RunnerBundleSHA256,
		InputSHA256: standard.InputSHA256, WorkloadSHA256: standard.WorkloadSHA256, ProfileSHA256: standard.ProfileSHA256,
		QPSHA256: standard.QPSHA256, Q: standard.Q, P: standard.P, ResidualMaxLevel: standard.ResidualMaxLevel,
		BootstrapMaxLevel: standard.BootstrapMaxLevel, E: standard.E, Slots: standard.Slots,
		PreBootstrapPair: prePair, Standard: summarizeRun(standard), Fast: summarizeRun(fast),
		BootstrapCalls: standard.BootstrapCalls + fast.BootstrapCalls, FixedMaxErrorGate: bootstrapMaxError,
	}
	if stage == "final" {
		fastPair, pairErr := pairedValues(fromValues(standard.BootstrapDecoded), fromValues(fast.BootstrapDecoded))
		if pairErr != nil || !fastPair.Pass {
			out.Status = "BATCH_BLOCKED_NEEDS_WEB_REVIEW"
			out.BootstrapPair = &fastPair
			if writeErr := writeJSON(outPath, out); writeErr != nil {
				return writeErr
			}
			return fmt.Errorf("Bootstrap Fast-vs-Standard gate failed: %+v (%v)", fastPair, pairErr)
		}
		out.Status = "BATCH_COMPLETE_READY_FOR_WEB_REVIEW"
		out.BootstrapPair = &fastPair
		out.BootstrapCalls = 2
	}
	return writeJSON(outPath, out)
}

func validateRunPair(standard, fast runEvidence) error {
	if standard.FastZeroSecret || !fast.FastZeroSecret || standard.BackendCommit != standardCommit || standard.BackendRef != standardCommit || fast.BackendCommit != fastCommit || fast.BackendRef != "fast-qprefix" {
		return errors.New("run backend identity/pins do not match genuine Standard and Fast")
	}
	if standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryDirty || fast.PrimaryDirty || standard.BackendDirty || fast.BackendDirty {
		return errors.New("paired runs do not have identical clean Primary/backend provenance")
	}
	for field, pair := range map[string][2]string{
		"config":   {standard.ConfigSHA256, fast.ConfigSHA256},
		"frontend": {standard.FrontendSHA256, fast.FrontendSHA256},
		"runner":   {standard.RunnerBundleSHA256, fast.RunnerBundleSHA256},
		"input":    {standard.InputSHA256, fast.InputSHA256},
		"workload": {standard.WorkloadSHA256, fast.WorkloadSHA256},
		"profile":  {standard.ProfileSHA256, fast.ProfileSHA256},
		"Q/P":      {standard.QPSHA256, fast.QPSHA256},
	} {
		if pair[0] == "" || pair[0] != pair[1] {
			return fmt.Errorf("paired %s provenance mismatch: %q != %q", field, pair[0], pair[1])
		}
	}
	if !slices.Equal(standard.Q, fast.Q) || !slices.Equal(standard.P, fast.P) || standard.ResidualMaxLevel != 1 || fast.ResidualMaxLevel != 1 || standard.BootstrapMaxLevel != 16 || fast.BootstrapMaxLevel != 16 || standard.E != 32 || fast.E != 32 || standard.Slots != 4096 || fast.Slots != 4096 {
		return errors.New("paired generated parameters/residual/Bootstrap profile mismatch")
	}
	if standard.BootstrapInputLevel != 0 || fast.BootstrapInputLevel != 0 || standard.BootstrapInputScale != "35184372088832" || fast.BootstrapInputScale != "35184372088832" {
		return errors.New("paired Bootstrap inputs do not have canonical residual Level 0 and exact Scale 2^45")
	}
	if !standard.SameLowSecret || !fast.SameLowSecret || standard.KeysetIdentity != fast.KeysetIdentity {
		return errors.New("paired runs changed the public Bootstrap keyset/low-level key lifecycle")
	}
	for _, run := range []runEvidence{standard, fast} {
		if !run.PreBootstrapOracle.Finite || run.PreBootstrapOracle.MaxComplex > bootstrapMaxError || len(run.PreBootstrapDecoded) != run.Slots || len(run.Checkpoints) < 5 {
			return fmt.Errorf("%s pre-Bootstrap native oracle or checkpoint set is incomplete", run.BackendRef)
		}
		for _, id := range []string{"encrypt_a_level5", "encrypt_b_level5", "add_level5", "rotate_level5"} {
			cp, ok := findCheckpoint(run.Checkpoints, id)
			if !ok || cp.State.Level != 5 || cp.State.Scale != "35184372088832" || cp.State.Degree != 1 || cp.State.AuthoritativeRows != qPrefixWidth(5) {
				return fmt.Errorf("%s is missing expected Level-5/Scale-2^45 checkpoint %s", run.BackendRef, id)
			}
		}
		final, ok := findCheckpoint(run.Checkpoints, "drop_to_residual_level0")
		if !ok || final.State.Level != 0 || final.State.Scale != "35184372088832" || final.State.Degree != 1 || final.State.AuthoritativeRows != 1 {
			return fmt.Errorf("%s is missing canonical Level-0/Scale-2^45 DropLevel checkpoint", run.BackendRef)
		}
	}
	for _, id := range []string{"encrypt_a_level5", "encrypt_b_level5", "add_level5", "rotate_level5", "drop_to_residual_level0"} {
		cp, ok := findCheckpoint(fast.Checkpoints, id)
		if !ok {
			return fmt.Errorf("Fast checkpoint %s is missing", id)
		}
		requireCompact := id != "encrypt_a_level5" && id != "encrypt_b_level5"
		if err := validateFastCheckpoint(cp, requireCompact, 1<<13); err != nil {
			return fmt.Errorf("Fast checkpoint %s: %w", id, err)
		}
	}
	return nil
}

func validateFastCheckpoint(cp checkpoint, requireCompact bool, n int) error {
	if cp.State.Level < 0 || cp.State.AuthoritativeRows != qPrefixWidth(cp.State.Level) {
		return errors.New("logical Level and authoritative Q-prefix width disagree")
	}
	if !cp.State.C1Zero {
		return errors.New("authoritative c1 rows are not zero")
	}
	if requireCompact && !compactAtPrefix(cp.State, n) {
		return errors.New("output did not leave non-authoritative Q rows dormant")
	}
	if cp.Capacity == nil || cp.Capacity.Rows != cp.State.AuthoritativeRows || !cp.Capacity.ObserverInvoked || !cp.Capacity.StrictFit {
		return errors.New("existing Q-prefix capacity observer did not prove strict fit")
	}
	return nil
}

func findCheckpoint(checkpoints []checkpoint, id string) (checkpoint, bool) {
	for _, item := range checkpoints {
		if item.ID == id {
			return item, true
		}
	}
	return checkpoint{}, false
}

func summarizeRun(run runEvidence) runSummary {
	return runSummary{
		BackendCommit: run.BackendCommit, BackendRef: run.BackendRef, BackendDirty: run.BackendDirty,
		FastZeroSecret: run.FastZeroSecret, PublicDispatch: run.PublicDispatch,
		BootstrapCalls: run.BootstrapCalls, BootstrapAttempted: run.BootstrapAttempted,
		KeysetIdentity: run.KeysetIdentity, SameLowSecret: run.SameLowSecret,
		Checkpoints: run.Checkpoints, PreOracle: run.PreBootstrapOracle, PreSHA256: run.PreBootstrapSHA256,
		BootstrapOracle: run.BootstrapOracle, BootstrapSHA256: run.BootstrapSHA256,
	}
}

func deterministicInput(slots int) []complex128 {
	values := make([]complex128, slots)
	for i := range values {
		values[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return values
}

func deterministicInputs(slots int) (a, b []complex128) {
	a = deterministicInput(slots)
	b = make([]complex128, slots)
	for i := range b {
		b[i] = complex(float64(i%19-9)/384, float64((5*i)%17-8)/768)
	}
	return
}

func encodedInputBound(values []complex128, scale float64) *big.Int {
	maximum := 0.0
	for _, value := range values {
		maximum = math.Max(maximum, cmplx.Abs(value))
	}
	return big.NewInt(int64(math.Ceil(2*maximum*scale + 1)))
}

func strictCapacity(bound, product *big.Int) bool {
	return bound != nil && product != nil && bound.Sign() >= 0 && product.Sign() > 0 && new(big.Int).Lsh(new(big.Int).Set(bound), 1).Cmp(product) < 0
}

func prefixProduct(q []uint64, rows int) *big.Int {
	out := big.NewInt(1)
	for _, qi := range q[:rows] {
		out.Mul(out, new(big.Int).SetUint64(qi))
	}
	return out
}

func qPrefixWidth(level int) int { return min(level+1, qPrefixCap) }

func addValues(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func rotateLeft(values []complex128, k int) []complex128 {
	out := make([]complex128, len(values))
	shift := k % len(values)
	for i := range out {
		out[i] = values[(i+shift)%len(values)]
	}
	return out
}

func compareValues(expected, actual []complex128) metric {
	if len(expected) == 0 || len(expected) != len(actual) {
		return metric{}
	}
	var errPower, signalPower float64
	result := metric{Finite: true}
	for i := range expected {
		delta := actual[i] - expected[i]
		magnitude := cmplx.Abs(delta)
		if math.IsNaN(magnitude) || math.IsInf(magnitude, 0) {
			result.Finite = false
			return result
		}
		errPower += magnitude * magnitude
		signalPower += cmplx.Abs(expected[i]) * cmplx.Abs(expected[i])
		result.MaxComplex = math.Max(result.MaxComplex, magnitude)
	}
	result.ComplexRMSE = math.Sqrt(errPower / float64(len(expected)))
	if errPower == 0 && signalPower > 0 {
		result.SNRState = "positive_infinity_zero_error"
	} else if signalPower == 0 && errPower > 0 {
		result.SNRState = "negative_infinity_zero_signal"
	} else if signalPower == 0 && errPower == 0 {
		result.SNRState = "undefined_zero_signal_and_error"
	} else {
		snr := 10 * math.Log10(signalPower/errPower)
		result.SNRDB = &snr
		result.SNRState = "finite"
	}
	result.Finite = !math.IsNaN(result.ComplexRMSE) && !math.IsInf(result.ComplexRMSE, 0)
	return result
}

func pairedValues(standard, fast []complex128) (pairedMetric, error) {
	if len(standard) != len(fast) || len(standard) == 0 {
		return pairedMetric{}, errors.New("paired decoded vectors are absent or have different lengths")
	}
	m := compareValues(standard, fast)
	return pairedMetric{RMSE: m.ComplexRMSE, Max: m.MaxComplex, Pass: m.Finite && m.MaxComplex <= bootstrapMaxError}, nil
}

func toValues(values []complex128) []complexValue {
	out := make([]complexValue, len(values))
	for i, value := range values {
		out[i] = complexValue{Real: real(value), Imag: imag(value)}
	}
	return out
}

func fromValues(values []complexValue) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func secretRowHashes(sk *rlwe.SecretKey, rows int) []string {
	if sk == nil || len(sk.Value.Q.Coeffs) < rows {
		return nil
	}
	return hashRows(sk.Value.Q.Coeffs[:rows])
}

func hashComplexValues(values []complex128) string {
	h := sha256.New()
	var buf [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(buf[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(buf[8:], math.Float64bits(imag(value)))
		_, _ = h.Write(buf[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashRows(rows [][]uint64) []string {
	out := make([]string, len(rows))
	var buf [8]byte
	for i, row := range rows {
		h := sha256.New()
		for _, value := range row {
			binary.LittleEndian.PutUint64(buf[:], value)
			_, _ = h.Write(buf[:])
		}
		out[i] = hex.EncodeToString(h.Sum(nil))
	}
	return out
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func runnerBundleHash() (string, error) {
	files := []string{frontendPath, "tools/fast-dropin-highlevel-bootstrap-batch-018/capacity_adapter_fast.go", "tools/fast-dropin-highlevel-bootstrap-batch-018/capacity_adapter_standard.go"}
	h := sha256.New()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read runner source %s: %w", file, err)
		}
		_, _ = fmt.Fprintf(h, "%s\x00", file)
		_, _ = h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." {
		if err = os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
