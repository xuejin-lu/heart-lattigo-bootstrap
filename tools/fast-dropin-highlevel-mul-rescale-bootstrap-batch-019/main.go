// Command fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 validates one
// identical public CKKS arithmetic path against the pinned Standard and Fast
// implementations. Bootstrap is gated until both backend preflights pass.
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
	canonicalConfigPath = "configs/bootstrap_config.logN13.json"
	standardCommit      = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	fastCommit          = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac"
	canonicalConfigHash = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"
	canonicalInputHash  = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
	canonicalQpHash     = "1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b"
	defaultScaleLog2    = 45
	bootstrapErrorGate  = 1e-6
	rotation            = 1
	qPrefixCap          = 4
	runSchema           = "fast-dropin-highlevel-mul-rescale-batch-019.run.v1"
	combinedSchema      = "fast-dropin-highlevel-mul-rescale-batch-019.combined.v1"
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
	SignalPower float64  `json:"signal_power"`
	ErrorPower  float64  `json:"error_power"`
	SNRDB       *float64 `json:"snr_db,omitempty"`
	SNRState    string   `json:"snr_state"`
}

type pairedMetric struct {
	RMSE  float64  `json:"rmse"`
	Max   float64  `json:"max_complex_error"`
	SNRDB *float64 `json:"snr_db,omitempty"`
	Pass  bool     `json:"pass"`
}

type scalePlan struct {
	DefaultInput             string  `json:"default_input_scale_2^45"`
	InputC                   string  `json:"chosen_input_c_scale_q5"`
	MulRelin                 string  `json:"chosen_mulrelin_scale_2^45_times_q5"`
	PostRescale              string  `json:"chosen_post_rescale_scale"`
	NaiveTwoDefaultNumerator string  `json:"naive_two_default_scale_numerator_2^90"`
	NaiveRescaleDivisor      string  `json:"naive_rescale_divisor_q5"`
	NaivePostRescaleLog2     float64 `json:"naive_post_rescale_log2"`
}

type keyLookups struct {
	GaloisKey     int `json:"galois_key"`
	GaloisKeyList int `json:"galois_key_list"`
	Relinearize   int `json:"relinearization_key"`
}

type keyAccess struct {
	EvaluatorConstruction keyLookups `json:"evaluator_construction"`
	Add                   keyLookups `json:"add"`
	MulRelin              keyLookups `json:"mul_relin"`
	Rescale               keyLookups `json:"rescale"`
	Rotate                keyLookups `json:"rotate"`
	BootstrapInternals    string     `json:"bootstrap_internal_key_accesses"`
}

type boundPlan struct {
	InputA       string `json:"input_a_B"`
	InputB       string `json:"input_b_B"`
	InputC       string `json:"input_c_B"`
	Add          string `json:"add_B"`
	MulRelin     string `json:"mulrelin_B"`
	Rescale      string `json:"rescale_B"`
	Q0123Product string `json:"S_Q0123"`
	Q0           string `json:"q0"`
	MulDivisor   string `json:"logical_rescale_divisor_q5"`
}

type capacityEvidence struct {
	Rows            int      `json:"authoritative_rows"`
	PrefixProduct   string   `json:"prefix_product"`
	Bound           string   `json:"independent_B"`
	ObserverMaxAbs  []string `json:"observer_component_max_abs,omitempty"`
	StrictFit       bool     `json:"strict_2B_less_than_prefix"`
	ObserverInvoked bool     `json:"existing_fast_capacity_observer_invoked"`
}

type checkpoint struct {
	ID       string            `json:"id"`
	API      string            `json:"public_api"`
	Level    int               `json:"level"`
	Scale    string            `json:"scale_integer"`
	ScaleLog float64           `json:"scale_log2"`
	Rows     [][]int           `json:"component_row_lengths"`
	RowHash  [][]string        `json:"authoritative_row_sha256"`
	C1Zero   bool              `json:"c1_zero_on_authoritative_rows"`
	Compact  bool              `json:"fast_compact_prefix_layout"`
	Capacity *capacityEvidence `json:"capacity,omitempty"`
}

type runEvidence struct {
	SchemaVersion       string         `json:"schema_version"`
	Status              string         `json:"status"`
	Error               string         `json:"error,omitempty"`
	CreatedUTC          string         `json:"created_utc"`
	GoVersion           string         `json:"go_version"`
	OS                  string         `json:"os"`
	Architecture        string         `json:"architecture"`
	PrimaryCommit       string         `json:"primary_commit"`
	PrimaryDirty        bool           `json:"primary_dirty"`
	BackendCommit       string         `json:"backend_commit"`
	BackendRef          string         `json:"backend_ref"`
	BackendDirty        bool           `json:"backend_dirty"`
	FastZeroSecret      bool           `json:"fast_zero_secret"`
	BootstrapDispatch   string         `json:"public_bootstrap_dispatch"`
	ConfigSHA256        string         `json:"config_sha256"`
	FrontendSHA256      string         `json:"shared_frontend_sha256"`
	InputSHA256         string         `json:"canonical_input_sha256"`
	WorkloadSHA256      string         `json:"workload_sha256"`
	ExpectedSHA256      string         `json:"expected_output_sha256"`
	ProfileSHA256       string         `json:"effective_profile_sha256"`
	QPSHA256            string         `json:"effective_qp_sha256"`
	Q                   []uint64       `json:"q_primes"`
	P                   []uint64       `json:"p_primes"`
	ResidualMaxLevel    int            `json:"residual_max_level"`
	BootstrapMaxLevel   int            `json:"bootstrap_max_level"`
	E                   int            `json:"ephemeral_secret_weight"`
	Slots               int            `json:"slots"`
	DefaultScale        string         `json:"default_scale_integer"`
	InputCScale         string         `json:"input_c_scale_q5_integer"`
	MulScale            string         `json:"mulrelin_scale_integer"`
	RescaleScale        string         `json:"rescale_scale_integer"`
	Bounds              boundPlan      `json:"independent_bounds"`
	KeysetIdentity      string         `json:"keyset_identity"`
	BootstrapSecretFull bool           `json:"bootstrap_secret_full_q_domain"`
	BootstrapSecretLow  bool           `json:"bootstrap_secret_preserves_residual_q0_q1"`
	KeyAccess           keyAccess      `json:"primitive_key_accesses"`
	BootstrapCalls      int            `json:"bootstrap_calls"`
	BootstrapAttempted  bool           `json:"bootstrap_attempted"`
	Checkpoints         []checkpoint   `json:"checkpoints"`
	PreOracle           metric         `json:"pre_bootstrap_cleartext_oracle"`
	PreDecoded          []complexValue `json:"pre_bootstrap_decoded_values,omitempty"`
	BootstrapOracle     *metric        `json:"bootstrap_cleartext_oracle,omitempty"`
	BootstrapDecoded    []complexValue `json:"bootstrap_decoded_values,omitempty"`
}

type combinedEvidence struct {
	SchemaVersion    string        `json:"schema_version"`
	Status           string        `json:"status"`
	CreatedUTC       string        `json:"created_utc"`
	GoVersion        string        `json:"go_version"`
	OS               string        `json:"os"`
	Architecture     string        `json:"architecture"`
	PrimaryCommit    string        `json:"primary_commit"`
	PrimaryDirty     bool          `json:"primary_dirty_during_runs"`
	ConfigSHA256     string        `json:"config_sha256"`
	FrontendSHA256   string        `json:"shared_frontend_sha256"`
	InputSHA256      string        `json:"canonical_input_sha256"`
	WorkloadSHA256   string        `json:"workload_sha256"`
	ProfileSHA256    string        `json:"effective_profile_sha256"`
	QPSHA256         string        `json:"effective_qp_sha256"`
	Q                []uint64      `json:"q_primes"`
	P                []uint64      `json:"p_primes"`
	Bounds           boundPlan     `json:"independent_bounds"`
	ScalePlan        scalePlan     `json:"scale_feasibility"`
	PreBootstrapPair pairedMetric  `json:"pre_bootstrap_fast_vs_standard"`
	BootstrapPair    *pairedMetric `json:"bootstrap_fast_vs_standard,omitempty"`
	Standard         runSummary    `json:"standard"`
	Fast             runSummary    `json:"fast"`
	BootstrapCalls   int           `json:"bootstrap_calls_total"`
	BootstrapGate    float64       `json:"fixed_max_error_gate"`
}

type runSummary struct {
	Status              string       `json:"status"`
	BackendCommit       string       `json:"backend_commit"`
	BackendRef          string       `json:"backend_ref"`
	BackendDirty        bool         `json:"backend_dirty"`
	FastZeroSecret      bool         `json:"fast_zero_secret"`
	BootstrapDispatch   string       `json:"public_bootstrap_dispatch"`
	BootstrapCalls      int          `json:"bootstrap_calls"`
	BootstrapAttempted  bool         `json:"bootstrap_attempted"`
	KeysetIdentity      string       `json:"keyset_identity"`
	BootstrapSecretFull bool         `json:"bootstrap_secret_full_q_domain"`
	BootstrapSecretLow  bool         `json:"bootstrap_secret_preserves_residual_q0_q1"`
	InputCScale         string       `json:"input_c_scale_q5_integer"`
	MulScale            string       `json:"mulrelin_scale_integer"`
	RescaleScale        string       `json:"rescale_scale_integer"`
	KeyAccess           keyAccess    `json:"primitive_key_accesses"`
	Checkpoints         []checkpoint `json:"checkpoints"`
	PreOracle           metric       `json:"pre_bootstrap_cleartext_oracle"`
	BootstrapOracle     *metric      `json:"bootstrap_cleartext_oracle,omitempty"`
}

type capacityAudit interface {
	Observe(string, *rlwe.Ciphertext, int) (int, string, []string, bool, error)
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

type execution struct {
	result    runEvidence
	btpEval   *bootstrapping.Evaluator
	input     *rlwe.Ciphertext
	decryptor *rlwe.Decryptor
	encoder   *ckks.Encoder
	expected  []complex128
}

func main() {
	if err := runCLI(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() error {
	mode := flag.String("mode", "run", "run, combine-preflight, or combine-final")
	out := flag.String("out", "", "raw run or compact combined JSON")
	primaryCommit := flag.String("primary-commit", "", "committed clean Primary revision")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty during the run")
	backendCommit := flag.String("backend-commit", "", "pinned Lattigo revision")
	backendRef := flag.String("backend-ref", "", "pinned Lattigo ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether backend was dirty")
	standardRun := flag.String("standard-run", "", "raw Standard JSON")
	fastRun := flag.String("fast-run", "", "raw Fast JSON")
	flag.Parse()
	if *out == "" {
		return errors.New("-out is required")
	}
	if *mode == "combine-preflight" || *mode == "combine-final" {
		if *standardRun == "" || *fastRun == "" {
			return errors.New("combine mode requires -standard-run and -fast-run")
		}
		return combine(*mode, *standardRun, *fastRun, *out)
	}
	if *mode != "run" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return errors.New("run mode requires pinned Primary/backend provenance and -out")
	}
	run, err := prepareRun(*primaryCommit, *primaryDirty, *backendCommit, *backendRef, *backendDirty)
	if err != nil {
		if run != nil {
			run.result.Status, run.result.Error = "preflight_failed", err.Error()
			_ = writeJSON(*out, run.result)
		}
		return err
	}
	run.result.Status = "preflight_passed_waiting_for_bootstrap"
	if err = writeJSON(*out, run.result); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "BATCH019_PREFLIGHT_READY — after the paired preflight passes, send exactly `bootstrap` to spend this backend's single Bootstrap allowance; otherwise send `abort`")
	line, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
	if readErr != nil || strings.TrimSpace(line) != "bootstrap" {
		run.result.Status = "preflight_passed_bootstrap_not_started"
		if writeErr := writeJSON(*out, run.result); writeErr != nil {
			return writeErr
		}
		if readErr != nil && !strings.Contains(readErr.Error(), "EOF") {
			return readErr
		}
		return nil
	}
	run.result.BootstrapCalls, run.result.BootstrapAttempted = 1, true
	run.result.Status = "bootstrap_call_started"
	if err = writeJSON(*out, run.result); err != nil {
		return fmt.Errorf("persist one-shot Bootstrap budget before invocation: %w", err)
	}
	output, err := run.btpEval.Bootstrap(run.input)
	if err != nil {
		run.result.Status, run.result.Error = "bootstrap_failed_after_single_call", err.Error()
		_ = writeJSON(*out, run.result)
		return fmt.Errorf("single public Bootstrap call: %w", err)
	}
	if err = recordBootstrapOutput(run, output); err != nil {
		run.result.Status, run.result.Error = "bootstrap_output_gate_failed", err.Error()
		_ = writeJSON(*out, run.result)
		return err
	}
	run.result.Status = "bootstrap_completed"
	return writeJSON(*out, run.result)
}

func prepareRun(primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty bool) (*execution, error) {
	run := &execution{result: runEvidence{
		SchemaVersion: runSchema, CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		PrimaryCommit: primaryCommit, PrimaryDirty: primaryDirty,
		BackendCommit: backendCommit, BackendRef: backendRef, BackendDirty: backendDirty,
		FastZeroSecret: isFastBackend(), BootstrapCalls: 0,
		KeysetIdentity: "one public GenEvaluationKeys result passed to ckks.NewEvaluator and bootstrapping.NewEvaluator",
	}}
	if primaryDirty || backendDirty {
		return run, errors.New("Primary and pinned backend must both be clean")
	}
	if isFastBackend() {
		if backendCommit != fastCommit || backendRef != "fast-qprefix" {
			return run, fmt.Errorf("Fast provenance mismatch: %s@%s", backendRef, backendCommit)
		}
	} else if backendCommit != standardCommit || backendRef != standardCommit {
		return run, fmt.Errorf("Standard provenance mismatch: %s@%s", backendRef, backendCommit)
	}
	configBytes, err := os.ReadFile(canonicalConfigPath)
	if err != nil {
		return run, err
	}
	run.result.ConfigSHA256 = hashBytes(configBytes)
	if run.result.ConfigSHA256 != canonicalConfigHash {
		return run, fmt.Errorf("frozen config SHA mismatch: %s", run.result.ConfigSHA256)
	}
	var cfg config
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return run, err
	}
	residual, btp, logSlots, err := parametersFromConfig(cfg)
	if err != nil {
		return run, err
	}
	full := btp.BootstrappingParameters
	if residual.MaxLevel() != 1 || full.MaxLevel() != 16 || logSlots != 12 || btp.EphemeralSecretWeight != 32 || full.LogDefaultScale() != defaultScaleLog2 {
		return run, fmt.Errorf("canonical E32 profile mismatch: residual=%d full=%d slots=%d E=%d defaultScale=%d", residual.MaxLevel(), full.MaxLevel(), logSlots, btp.EphemeralSecretWeight, full.LogDefaultScale())
	}
	q, p := full.Q(), full.P()
	qpBytes, _ := json.Marshal(struct{ Q, P []uint64 }{q, p})
	run.result.QPSHA256 = hashBytes(qpBytes)
	if run.result.QPSHA256 != canonicalQpHash || len(q) <= 5 {
		return run, fmt.Errorf("canonical Q/P mismatch or missing q5: sha=%s q-count=%d", run.result.QPSHA256, len(q))
	}
	run.result.Q, run.result.P = q, p
	run.result.ResidualMaxLevel, run.result.BootstrapMaxLevel = residual.MaxLevel(), full.MaxLevel()
	run.result.E, run.result.Slots = btp.EphemeralSecretWeight, 1<<logSlots
	run.result.InputSHA256 = canonicalInputHash
	profileBytes, _ := json.Marshal(struct {
		LogN, DefaultScale, LogSlots, ResidualLevel, FullLevel, E int
		Q, P                                                      []uint64
	}{cfg.LogN, cfg.LogDefaultScale, logSlots, residual.MaxLevel(), full.MaxLevel(), btp.EphemeralSecretWeight, q, p})
	run.result.ProfileSHA256 = hashBytes(profileBytes)
	frontendBytes, err := os.ReadFile("tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019/main.go")
	if err != nil {
		return run, err
	}
	run.result.FrontendSHA256 = hashBytes(frontendBytes)
	if err = validateCanonicalInputs(run.result.Slots); err != nil {
		return run, err
	}
	a, b, c := deterministicInputs(run.result.Slots)
	q5 := q[5]
	defaultScale := new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2)
	cScale := new(big.Int).SetUint64(q5)
	mulScale := new(big.Int).Mul(new(big.Int).Set(defaultScale), cScale)
	rescaleScale := new(big.Int).Quo(new(big.Int).Set(mulScale), cScale)
	run.result.DefaultScale, run.result.InputCScale = defaultScale.String(), cScale.String()
	run.result.MulScale, run.result.RescaleScale = mulScale.String(), rescaleScale.String()
	if rescaleScale.Cmp(defaultScale) != 0 {
		return run, errors.New("q5 per-input Scale did not close MulRelin/Rescale to frozen 2^45")
	}
	workloadBytes, _ := json.Marshal(struct {
		A, B, C []complexValue
		CScale  string
	}{toValues(a), toValues(b), toValues(c), cScale.String()})
	run.result.WorkloadSHA256 = hashBytes(workloadBytes)
	run.result.ExpectedSHA256 = hashComplexValues(rotateLeft(multiply(addValues(a, b), c), rotation))
	run.result.Bounds, err = deriveBounds(a, b, c, q, full.N())
	if err != nil {
		return run, err
	}

	keygen := rlwe.NewKeyGenerator(residual)
	sk := keygen.GenSecretKeyNew()
	keys, bootstrapSecret, err := btp.GenEvaluationKeys(sk)
	if err != nil {
		return run, fmt.Errorf("public GenEvaluationKeys: %w", err)
	}
	if keys == nil || keys.MemEvaluationKeySet == nil || bootstrapSecret == nil {
		return run, errors.New("public GenEvaluationKeys returned incomplete key material")
	}
	run.result.BootstrapSecretFull = bootstrapSecret.LevelQ() == full.MaxLevel()
	if !run.result.BootstrapSecretFull {
		return run, fmt.Errorf("generated Bootstrap secret LevelQ=%d, full Q requires %d", bootstrapSecret.LevelQ(), full.MaxLevel())
	}
	run.result.BootstrapSecretLow = secretPrefixEqual(sk, bootstrapSecret, 2)
	if !run.result.BootstrapSecretLow {
		return run, errors.New("extended Bootstrap secret does not preserve residual q0/q1")
	}
	trackedKeys := &trackingKeySet{EvaluationKeySet: keys.MemEvaluationKeySet}
	evaluator := ckks.NewEvaluator(full, trackedKeys)
	run.result.KeyAccess.EvaluatorConstruction = trackedKeys.lookups
	trackedKeys.lookups = keyLookups{}
	btsEval, err := bootstrapping.NewEvaluator(btp, keys)
	if err != nil {
		return run, fmt.Errorf("public bootstrapping.NewEvaluator: %w", err)
	}
	if btsEval == nil || isFastBackend() != (btsEval.Evaluator == nil) {
		return run, errors.New("public Bootstrap dispatch did not match the pinned backend")
	}
	if isFastBackend() {
		run.result.BootstrapDispatch = "Fast public Bootstrap dispatch selected from frozen zero-secret capability"
	} else {
		run.result.BootstrapDispatch = "genuine Standard public Bootstrap evaluator"
	}
	run.result.KeyAccess.BootstrapInternals = "not runtime-instrumented: public Bootstrap uses its existing concrete EvaluationKeys/MemEvaluationKeySet; no claim of zero Bootstrap-internal key access"
	fullEncoder, residualEncoder := ckks.NewEncoder(full), ckks.NewEncoder(residual)
	decryptor := rlwe.NewDecryptor(residual, sk)
	defaultRLWEScale := rlwe.NewScale(defaultScale)
	q5Scale := rlwe.NewScale(q5)
	ctA, err := encryptAtLevel(full, fullEncoder, bootstrapSecret, a, defaultRLWEScale, 5, logSlots)
	if err != nil {
		return run, fmt.Errorf("public EncryptNew(A) Level5: %w", err)
	}
	ctB, err := encryptAtLevel(full, fullEncoder, bootstrapSecret, b, defaultRLWEScale, 5, logSlots)
	if err != nil {
		return run, fmt.Errorf("public EncryptNew(B) Level5: %w", err)
	}
	ctC, err := encryptAtLevel(full, fullEncoder, bootstrapSecret, c, q5Scale, 5, logSlots)
	if err != nil {
		return run, fmt.Errorf("public EncryptNew(C, Scale=q5) Level5: %w", err)
	}
	if ctA.Level() != 5 || ctB.Level() != 5 || ctC.Level() != 5 || ctA.Scale.Cmp(defaultRLWEScale) != 0 || ctB.Scale.Cmp(defaultRLWEScale) != 0 || ctC.Scale.Cmp(q5Scale) != 0 {
		return run, errors.New("public EncryptNew input Level/Scale mismatch")
	}
	audit := makeCapacityAudit(full)
	if isFastBackend() != (audit != nil) {
		return run, errors.New("Fast capacity observer adapter/backend mismatch")
	}
	for _, item := range []struct {
		id    string
		ct    *rlwe.Ciphertext
		bound string
	}{
		{"encrypt_a_level5", ctA, run.result.Bounds.InputA},
		{"encrypt_b_level5", ctB, run.result.Bounds.InputB},
		{"encrypt_c_q5_scale_level5", ctC, run.result.Bounds.InputC},
	} {
		cp, cpErr := makeCheckpoint(audit, q, item.id, "rlwe.NewEncryptor(...).EncryptNew", item.ct, item.bound, false)
		if cpErr != nil {
			return run, cpErr
		}
		if (isFastBackend() && !cp.C1Zero) || (!isFastBackend() && cp.C1Zero) {
			return run, fmt.Errorf("%s c1 lifecycle mismatch: Fast must remain zero-c1 and genuine Standard EncryptNew must be nonzero-c1", item.id)
		}
		run.result.Checkpoints = append(run.result.Checkpoints, cp)
	}

	keyCountsBefore := trackedKeys.lookups
	add, err := evaluator.AddNew(ctA, ctB)
	if err != nil {
		return run, fmt.Errorf("public AddNew: %w", err)
	}
	run.result.KeyAccess.Add = subtractLookups(trackedKeys.lookups, keyCountsBefore)
	if add.Level() != 5 || add.Scale.Cmp(defaultRLWEScale) != 0 {
		return run, errors.New("AddNew Level/Scale mismatch")
	}
	cp, err := makeCheckpoint(audit, q, "add_level5", "ckks.Evaluator.AddNew", add, run.result.Bounds.Add, true)
	if err != nil {
		return run, err
	}
	run.result.Checkpoints = append(run.result.Checkpoints, cp)

	keyCountsBefore = trackedKeys.lookups
	product, err := evaluator.MulRelinNew(add, ctC)
	if err != nil {
		return run, fmt.Errorf("public MulRelinNew: %w", err)
	}
	if product.Level() != 5 || product.Scale.Cmp(rlwe.NewScale(mulScale)) != 0 {
		return run, errors.New("MulRelinNew Level/Scale mismatch; expected default*q5")
	}
	run.result.KeyAccess.MulRelin = subtractLookups(trackedKeys.lookups, keyCountsBefore)
	cp, err = makeCheckpoint(audit, q, "mulrelin_level5", "ckks.Evaluator.MulRelinNew", product, run.result.Bounds.MulRelin, true)
	if err != nil {
		return run, err
	}
	run.result.Checkpoints = append(run.result.Checkpoints, cp)

	keyCountsBefore = trackedKeys.lookups
	rescaled := ckks.NewCiphertext(full, 1, 4)
	if err = evaluator.Rescale(product, rescaled); err != nil {
		return run, fmt.Errorf("public Rescale(logical q5): %w", err)
	}
	run.result.KeyAccess.Rescale = subtractLookups(trackedKeys.lookups, keyCountsBefore)
	if rescaled.Level() != 4 || rescaled.Scale.Cmp(defaultRLWEScale) != 0 {
		return run, errors.New("Rescale did not consume Level5/q5 and restore Scale=2^45")
	}
	cp, err = makeCheckpoint(audit, q, "rescale_q5_level4", "ckks.Evaluator.Rescale", rescaled, run.result.Bounds.Rescale, true)
	if err != nil {
		return run, err
	}
	run.result.Checkpoints = append(run.result.Checkpoints, cp)

	keyCountsBefore = trackedKeys.lookups
	rotated, err := evaluator.RotateNew(rescaled, rotation)
	if err != nil {
		return run, fmt.Errorf("public RotateNew(1): %w", err)
	}
	if rotated.Level() != 4 || rotated.Scale.Cmp(defaultRLWEScale) != 0 {
		return run, errors.New("RotateNew changed expected Level/Scale")
	}
	run.result.KeyAccess.Rotate = subtractLookups(trackedKeys.lookups, keyCountsBefore)
	cp, err = makeCheckpoint(audit, q, "rotate_level4", "ckks.Evaluator.RotateNew(1)", rotated, run.result.Bounds.Rescale, true)
	if err != nil {
		return run, err
	}
	run.result.Checkpoints = append(run.result.Checkpoints, cp)

	// This ordinary public DropLevel only projects the already-proven centered
	// lift. It does not alter Scale or perform a rescale.
	drop := evaluator.DropLevelNew(rotated, rotated.Level())
	if drop.Level() != 0 || drop.Scale.Cmp(defaultRLWEScale) != 0 {
		return run, errors.New("DropLevelNew did not yield Level0 / Scale2^45")
	}
	cp, err = makeCheckpoint(audit, q, "drop_level0", "ckks.Evaluator.DropLevelNew(4)", drop, run.result.Bounds.Rescale, true)
	if err != nil {
		return run, err
	}
	run.result.Checkpoints = append(run.result.Checkpoints, cp)
	if isFastBackend() && !cp.Compact {
		return run, errors.New("Fast DropLevelNew did not retain only q0 at Level0")
	}

	expected := rotateLeft(multiply(addValues(a, b), c), rotation)
	decoded := make([]complex128, len(expected))
	if err = residualEncoder.Decode(decryptor.DecryptNew(drop), decoded); err != nil {
		return run, fmt.Errorf("native residual Level0 DecryptNew/Decode: %w", err)
	}
	oracle := compareValues(expected, decoded)
	if !oracle.Finite || oracle.MaxComplex > bootstrapErrorGate {
		return run, fmt.Errorf("pre-Bootstrap cleartext oracle max %.12g exceeds fixed gate %.12g", oracle.MaxComplex, bootstrapErrorGate)
	}
	run.result.PreOracle, run.result.PreDecoded = oracle, toValues(decoded)
	return &execution{result: run.result, btpEval: btsEval, input: drop, decryptor: decryptor, encoder: residualEncoder, expected: expected}, nil
}

func recordBootstrapOutput(run *execution, output *rlwe.Ciphertext) error {
	if output == nil {
		return errors.New("public Bootstrap returned nil")
	}
	if output.Level() != 1 || output.Scale.Cmp(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2))) != 0 {
		return fmt.Errorf("Bootstrap output Level/Scale mismatch: L%d scale=%s", output.Level(), output.Scale.BigInt())
	}
	cp, err := makeCheckpoint(nil, run.result.Q, "public_bootstrap_output", "bootstrapping.Evaluator.Bootstrap", output, "", false)
	if err != nil {
		return err
	}
	if isFastBackend() && !cp.C1Zero {
		return errors.New("Fast Bootstrap output violated zero-c1 semantics")
	}
	run.result.Checkpoints = append(run.result.Checkpoints, cp)
	decoded := make([]complex128, len(run.expected))
	if err = run.encoder.Decode(run.decryptor.DecryptNew(output), decoded); err != nil {
		return fmt.Errorf("native Bootstrap-output DecryptNew/Decode: %w", err)
	}
	metric := compareValues(run.expected, decoded)
	if !metric.Finite || metric.MaxComplex > bootstrapErrorGate {
		return fmt.Errorf("Bootstrap cleartext oracle max %.12g exceeds fixed gate %.12g", metric.MaxComplex, bootstrapErrorGate)
	}
	run.result.BootstrapOracle, run.result.BootstrapDecoded = &metric, toValues(decoded)
	return nil
}

func combine(mode, standardPath, fastPath, outPath string) error {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return err
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return err
	}
	combined := combinedEvidence{
		SchemaVersion: combinedSchema, CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: standard.GoVersion, OS: standard.OS, Architecture: standard.Architecture,
		BootstrapGate: bootstrapErrorGate,
	}
	combined.Standard, combined.Fast = summarizeRun(standard), summarizeRun(fast)
	if err := validatePairProvenance(standard, fast); err != nil {
		combined.Status = "blocked_pair_provenance"
		_ = writeJSON(outPath, combined)
		return err
	}
	combined.PrimaryCommit, combined.PrimaryDirty = standard.PrimaryCommit, standard.PrimaryDirty
	combined.ConfigSHA256, combined.FrontendSHA256 = standard.ConfigSHA256, standard.FrontendSHA256
	combined.InputSHA256, combined.WorkloadSHA256 = standard.InputSHA256, standard.WorkloadSHA256
	combined.ProfileSHA256, combined.QPSHA256 = standard.ProfileSHA256, standard.QPSHA256
	combined.Q, combined.P, combined.Bounds = standard.Q, standard.P, standard.Bounds
	combined.ScalePlan = buildScalePlan(standard.Q[5], standard.DefaultScale, standard.InputCScale, standard.MulScale, standard.RescaleScale)
	if mode == "combine-preflight" {
		if standard.Status != "preflight_passed_waiting_for_bootstrap" || fast.Status != "preflight_passed_waiting_for_bootstrap" || standard.BootstrapCalls != 0 || fast.BootstrapCalls != 0 {
			combined.Status = "blocked_preflight"
			_ = writeJSON(outPath, combined)
			return fmt.Errorf("both zero-call preflights must pass before Bootstrap: standard=%s fast=%s calls=%d/%d", standard.Status, fast.Status, standard.BootstrapCalls, fast.BootstrapCalls)
		}
		pair, err := pairedValues(fromValues(standard.PreDecoded), fromValues(fast.PreDecoded))
		if err != nil || !pair.Pass || !standard.PreOracle.Finite || !fast.PreOracle.Finite || standard.PreOracle.MaxComplex > bootstrapErrorGate || fast.PreOracle.MaxComplex > bootstrapErrorGate {
			combined.Status = "blocked_prebootstrap_numerical_gate"
			_ = writeJSON(outPath, combined)
			if err != nil {
				return err
			}
			return fmt.Errorf("pre-Bootstrap paired/oracle max gate failed: paired=%g standard=%g fast=%g", pair.Max, standard.PreOracle.MaxComplex, fast.PreOracle.MaxComplex)
		}
		combined.PreBootstrapPair, combined.Status = pair, "BATCH019_PREFLIGHT_PASSED_PENDING_ONE_SHOT_BOOTSTRAP"
		return writeJSON(outPath, combined)
	}
	if mode != "combine-final" {
		return fmt.Errorf("unsupported combine mode %q", mode)
	}
	if standard.Status != "bootstrap_completed" || fast.Status != "bootstrap_completed" || standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 || standard.BootstrapOracle == nil || fast.BootstrapOracle == nil {
		combined.Status = "blocked_bootstrap_result"
		_ = writeJSON(outPath, combined)
		return fmt.Errorf("expected exactly one completed Bootstrap per backend: status=%s/%s calls=%d/%d", standard.Status, fast.Status, standard.BootstrapCalls, fast.BootstrapCalls)
	}
	prePair, err := pairedValues(fromValues(standard.PreDecoded), fromValues(fast.PreDecoded))
	if err != nil {
		return err
	}
	bootstrapPair, err := pairedValues(fromValues(standard.BootstrapDecoded), fromValues(fast.BootstrapDecoded))
	if err != nil {
		return err
	}
	if !prePair.Pass || !bootstrapPair.Pass {
		combined.Status = "blocked_paired_gate"
		_ = writeJSON(outPath, combined)
		return errors.New("paired backend error exceeded fixed 1e-6 max gate")
	}
	combined.PreBootstrapPair, combined.BootstrapPair = prePair, &bootstrapPair
	combined.BootstrapCalls = standard.BootstrapCalls + fast.BootstrapCalls
	combined.Status = "BATCH_COMPLETE_READY_FOR_WEB_REVIEW"
	return writeJSON(outPath, combined)
}

func summarizeRun(run runEvidence) runSummary {
	return runSummary{
		Status: run.Status, BackendCommit: run.BackendCommit, BackendRef: run.BackendRef,
		BackendDirty: run.BackendDirty, FastZeroSecret: run.FastZeroSecret,
		BootstrapDispatch: run.BootstrapDispatch, BootstrapCalls: run.BootstrapCalls,
		BootstrapAttempted: run.BootstrapAttempted, KeysetIdentity: run.KeysetIdentity,
		BootstrapSecretFull: run.BootstrapSecretFull, BootstrapSecretLow: run.BootstrapSecretLow,
		InputCScale: run.InputCScale, MulScale: run.MulScale, RescaleScale: run.RescaleScale,
		KeyAccess: run.KeyAccess, Checkpoints: run.Checkpoints, PreOracle: run.PreOracle,
		BootstrapOracle: run.BootstrapOracle,
	}
}

func buildScalePlan(q5 uint64, defaultScale, inputC, mulScale, postRescale string) scalePlan {
	naiveNumerator := new(big.Int).Lsh(new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2), defaultScaleLog2)
	naiveLog2 := 2*float64(defaultScaleLog2) - math.Log2(float64(q5))
	return scalePlan{
		DefaultInput: defaultScale, InputC: inputC, MulRelin: mulScale, PostRescale: postRescale,
		NaiveTwoDefaultNumerator: naiveNumerator.String(), NaiveRescaleDivisor: new(big.Int).SetUint64(q5).String(),
		NaivePostRescaleLog2: naiveLog2,
	}
}

func validatePairProvenance(standard, fast runEvidence) error {
	if standard.FastZeroSecret || !fast.FastZeroSecret || standard.BackendCommit != standardCommit || standard.BackendRef != standardCommit || fast.BackendCommit != fastCommit || fast.BackendRef != "fast-qprefix" {
		return errors.New("backend identity/ref is not the frozen genuine Standard and Fast pair")
	}
	if standard.PrimaryDirty || fast.PrimaryDirty || standard.BackendDirty || fast.BackendDirty || standard.PrimaryCommit != fast.PrimaryCommit {
		return errors.New("paired measurements require identical clean committed Primary and backend pins")
	}
	if !standard.BootstrapSecretFull || !fast.BootstrapSecretFull || !standard.BootstrapSecretLow || !fast.BootstrapSecretLow {
		return errors.New("paired runs did not prove the expected full-Q key domain and residual secret prefix")
	}
	if standard.ConfigSHA256 != fast.ConfigSHA256 || standard.FrontendSHA256 != fast.FrontendSHA256 || standard.InputSHA256 != fast.InputSHA256 || standard.WorkloadSHA256 != fast.WorkloadSHA256 || standard.ExpectedSHA256 != fast.ExpectedSHA256 || standard.ProfileSHA256 != fast.ProfileSHA256 || standard.QPSHA256 != fast.QPSHA256 {
		return errors.New("Standard/Fast config, shared frontend, workload, profile or Q/P provenance differs")
	}
	if standard.ConfigSHA256 != canonicalConfigHash || standard.InputSHA256 != canonicalInputHash || standard.QPSHA256 != canonicalQpHash ||
		standard.ResidualMaxLevel != 1 || fast.ResidualMaxLevel != 1 || standard.BootstrapMaxLevel != 16 || fast.BootstrapMaxLevel != 16 ||
		standard.E != 32 || fast.E != 32 || standard.Slots != 4096 || fast.Slots != 4096 {
		return errors.New("paired runs do not match the frozen canonical LogN13 E32 profile")
	}
	if !slices.Equal(standard.Q, fast.Q) || !slices.Equal(standard.P, fast.P) {
		return errors.New("Standard/Fast actual Q/P arrays differ")
	}
	if len(standard.Q) <= 5 {
		return errors.New("paired canonical Q chain is missing logical q5")
	}
	defaultScale := new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2)
	q5 := new(big.Int).SetUint64(standard.Q[5])
	wantMulScale := new(big.Int).Mul(new(big.Int).Set(defaultScale), q5)
	if standard.InputCScale != q5.String() || fast.InputCScale != q5.String() ||
		standard.MulScale != wantMulScale.String() || fast.MulScale != wantMulScale.String() ||
		standard.RescaleScale != defaultScale.String() || fast.RescaleScale != defaultScale.String() {
		return errors.New("paired runs do not satisfy the exact default*q5/q5 Scale closure")
	}
	if standard.GoVersion != fast.GoVersion || standard.OS != fast.OS || standard.Architecture != fast.Architecture {
		return errors.New("paired run environments differ")
	}
	if standard.KeyAccess.MulRelin.Relinearize != 1 || standard.KeyAccess.Rotate.GaloisKey != 1 ||
		fast.KeyAccess.MulRelin.Relinearize != 0 || fast.KeyAccess.MulRelin.GaloisKey != 0 || fast.KeyAccess.Rotate.GaloisKey != 0 {
		return errors.New("primitive key-access behavior differs from native Standard and zero-secret Fast contracts")
	}
	return nil
}

func makeCheckpoint(audit capacityAudit, q []uint64, id, api string, ct *rlwe.Ciphertext, bound string, requireCompact bool) (checkpoint, error) {
	if ct == nil || ct.MetaData == nil || len(ct.Value) != 2 || ct.Degree() != 1 || ct.Level() < 0 || ct.Level() >= len(q) {
		return checkpoint{}, fmt.Errorf("%s: invalid ciphertext shape/metadata", id)
	}
	rows := qPrefixWidth(ct.Level())
	cp := checkpoint{ID: id, API: api, Level: ct.Level(), Scale: ct.Scale.BigInt().String(), ScaleLog: ct.Scale.Log2(), C1Zero: true}
	cp.Rows, cp.RowHash = make([][]int, 2), make([][]string, 2)
	for component := range ct.Value {
		polyRows := ct.Value[component].Coeffs
		cp.Rows[component] = make([]int, len(polyRows))
		if len(polyRows) < rows {
			return checkpoint{}, fmt.Errorf("%s component %d has %d rows, needs %d", id, component, len(polyRows), rows)
		}
		cp.RowHash[component] = make([]string, rows)
		for row, coeffs := range polyRows {
			cp.Rows[component][row] = len(coeffs)
		}
		for row := 0; row < rows; row++ {
			if len(polyRows[row]) != ct.N() {
				return checkpoint{}, fmt.Errorf("%s component %d q%d backing length mismatch", id, component, row)
			}
			cp.RowHash[component][row] = hashUint64s(polyRows[row])
			if component == 1 {
				for _, v := range polyRows[row] {
					if v != 0 {
						cp.C1Zero = false
						break
					}
				}
			}
		}
	}
	cp.Compact = true
	for _, componentRows := range cp.Rows {
		for row := 0; row < rows; row++ {
			if componentRows[row] != ct.N() {
				cp.Compact = false
			}
		}
		for row := rows; row < len(componentRows); row++ {
			if componentRows[row] != 0 {
				cp.Compact = false
			}
		}
	}
	if requireCompact && isFastBackend() && !cp.Compact {
		return checkpoint{}, fmt.Errorf("%s Fast output allocated/retained dormant Q rows", id)
	}
	product := prefixProduct(q, rows)
	var independentBound *big.Int
	if bound != "" {
		var ok bool
		independentBound, ok = new(big.Int).SetString(bound, 10)
		if !ok {
			return checkpoint{}, fmt.Errorf("%s invalid independent B=%q", id, bound)
		}
		if !strictCapacity(independentBound, product) {
			return checkpoint{}, fmt.Errorf("%s independent 2B is not strictly below prefix product: B=%s S_Q=%s", id, independentBound, product)
		}
		if isFastBackend() {
			cp.Capacity = &capacityEvidence{Rows: rows, PrefixProduct: product.String(), Bound: bound, StrictFit: true}
		}
	}
	if audit != nil {
		if independentBound == nil {
			return checkpoint{}, fmt.Errorf("%s Fast capacity observer lacks an independent bound", id)
		}
		seenRows, seenProduct, maxAbs, strict, err := audit.Observe(id, ct, rows)
		if err != nil {
			return checkpoint{}, fmt.Errorf("%s Fast capacity observer: %w", id, err)
		}
		if seenRows != rows || seenProduct != product.String() || !strict || len(maxAbs) != 2 {
			return checkpoint{}, fmt.Errorf("%s Fast capacity observer returned inconsistent rows/product/fit", id)
		}
		c0, ok0 := new(big.Int).SetString(maxAbs[0], 10)
		c1, ok1 := new(big.Int).SetString(maxAbs[1], 10)
		if !ok0 || !ok1 || c0.Cmp(independentBound) > 0 || c1.Sign() != 0 {
			return checkpoint{}, fmt.Errorf("%s Fast observed centered component maxima exceed independent bound: %v", id, maxAbs)
		}
		if cp.Capacity == nil {
			cp.Capacity = &capacityEvidence{Rows: rows, PrefixProduct: product.String()}
		}
		cp.Capacity.ObserverMaxAbs, cp.Capacity.ObserverInvoked = maxAbs, true
	}
	return cp, nil
}

func deriveBounds(a, b, c []complex128, q []uint64, n int) (boundPlan, error) {
	if len(q) <= 5 {
		return boundPlan{}, errors.New("Q chain does not contain logical q5")
	}
	scale := new(big.Int).Lsh(big.NewInt(1), defaultScaleLog2)
	ba, err := encodedInputBoundExact(a, scale)
	if err != nil {
		return boundPlan{}, err
	}
	bb, err := encodedInputBoundExact(b, scale)
	if err != nil {
		return boundPlan{}, err
	}
	bc, err := encodedInputBoundExact(c, new(big.Int).SetUint64(q[5]))
	if err != nil {
		return boundPlan{}, err
	}
	badd := new(big.Int).Add(new(big.Int).Set(ba), bb)
	bmul := new(big.Int).Mul(big.NewInt(int64(n)), new(big.Int).Mul(new(big.Int).Set(badd), bc))
	q5 := new(big.Int).SetUint64(q[5])
	br := new(big.Int).Add(new(big.Int).Set(bmul), new(big.Int).Rsh(new(big.Int).Sub(new(big.Int).Set(q5), big.NewInt(1)), 1))
	br.Quo(br, q5)
	prefix := prefixProduct(q, 4)
	for name, bound := range map[string]*big.Int{"EncryptNew(A)": ba, "EncryptNew(B)": bb, "EncryptNew(C)": bc, "AddNew": badd, "MulRelinNew": bmul, "Rescale(q5)": br} {
		if !strictCapacity(bound, prefix) {
			return boundPlan{}, fmt.Errorf("A-stage capacity failure at %s: B=%s, S_Q0123=%s, deficit=%s", name, bound, prefix, new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).Set(bound), 1), prefix))
		}
	}
	q0 := new(big.Int).SetUint64(q[0])
	if !strictCapacity(br, q0) {
		return boundPlan{}, fmt.Errorf("A-stage terminal Level0 capacity failure after DropLevel: B=%s, q0=%s, deficit=%s", br, q0, new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).Set(br), 1), q0))
	}
	return boundPlan{InputA: ba.String(), InputB: bb.String(), InputC: bc.String(), Add: badd.String(), MulRelin: bmul.String(), Rescale: br.String(), Q0123Product: prefix.String(), Q0: q0.String(), MulDivisor: q5.String()}, nil
}

func encodedInputBoundExact(values []complex128, scale *big.Int) (*big.Int, error) {
	var maxSquared *big.Rat
	for _, v := range values {
		r := new(big.Rat).SetFloat64(real(v))
		i := new(big.Rat).SetFloat64(imag(v))
		if r == nil || i == nil {
			return nil, errors.New("non-finite workload coordinate")
		}
		rs := new(big.Rat).Mul(r, r)
		is := new(big.Rat).Mul(i, i)
		sq := new(big.Rat).Add(rs, is)
		if maxSquared == nil || sq.Cmp(maxSquared) > 0 {
			maxSquared = sq
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

func parametersFromConfig(cfg config) (ckks.Parameters, bootstrapping.Parameters, int, error) {
	if cfg.LogN != 13 || cfg.LogDefaultScale != defaultScaleLog2 || cfg.SecretHamming != 192 ||
		!slices.Equal(cfg.Q0, []int{55}) || !slices.Equal(cfg.QSlotsToCoeffs, []int{39, 39, 39}) ||
		!slices.Equal(cfg.QCoeffsToSlots, []int{56, 56, 56, 56}) || len(cfg.P) != 5 ||
		!slices.Equal(cfg.SlotsToCoeffsDFT, []int{1, 1, 1}) || !slices.Equal(cfg.CoeffsToSlotsDFT, []int{1, 1, 1, 1}) ||
		cfg.LogSlots != -1 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 || cfg.Mod1DoubleAngle != 3 ||
		cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 || cfg.Repetitions != 3 || cfg.Warmup != 1 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, errors.New("config is not the frozen canonical LogN13 E32 profile")
	}
	for _, primeBits := range cfg.P {
		if primeBits != 61 {
			return ckks.Parameters{}, bootstrapping.Parameters{}, 0, fmt.Errorf("unexpected P prime size %d", primeBits)
		}
	}
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: cfg.LogN, LogQ: []int{cfg.Q0[0], cfg.QSlotsToCoeffs[0]}, LogDefaultScale: cfg.LogDefaultScale, Xs: ring.Ternary{H: cfg.SecretHamming}})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, err
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
		return ckks.Parameters{}, bootstrapping.Parameters{}, 0, err
	}
	btp.CircuitOrder, btp.ResidualParameters, btp.EphemeralSecretWeight = bootstrapping.ModUpThenEncode, residual, 32
	return residual, btp, logSlots, nil
}

func factorization(depths, scales []int) ([][]int, error) {
	if len(depths) == 0 || len(depths) != len(scales) {
		return nil, errors.New("DFT depth and scale lengths must match")
	}
	out := make([][]int, len(depths))
	for i, depth := range depths {
		if depth <= 0 || scales[i] <= 0 {
			return nil, fmt.Errorf("invalid DFT factorization index %d", i)
		}
		out[i] = make([]int, depth)
		for j := range out[i] {
			out[i][j] = scales[i]
		}
	}
	return out, nil
}

func encryptAtLevel(params ckks.Parameters, encoder *ckks.Encoder, sk *rlwe.SecretKey, values []complex128, scale rlwe.Scale, level, logSlots int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, level)
	pt.Scale, pt.IsNTT, pt.LogDimensions = scale, true, ring.Dimensions{Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(pt)
}

func deterministicInput(slots int) []complex128 {
	out := make([]complex128, slots)
	for i := range out {
		out[i] = complex(float64(i%17-8)/256, float64((3*i)%13-6)/512)
	}
	return out
}

func deterministicInputs(slots int) (a, b, c []complex128) {
	a = deterministicInput(slots)
	b, c = make([]complex128, slots), make([]complex128, slots)
	for i := range b {
		b[i] = complex(float64(i%19-9)/384, float64((5*i)%17-8)/768)
		c[i] = a[i]
	}
	return
}

func validateCanonicalInputs(slots int) error {
	values := deterministicInput(slots)
	if hashComplexValues(values) != canonicalInputHash {
		return errors.New("deterministic input does not match the canonical 017 input SHA")
	}
	return nil
}

func qPrefixWidth(level int) int { return min(level+1, qPrefixCap) }

func prefixProduct(q []uint64, rows int) *big.Int {
	out := big.NewInt(1)
	for _, qi := range q[:rows] {
		out.Mul(out, new(big.Int).SetUint64(qi))
	}
	return out
}

func strictCapacity(bound, product *big.Int) bool {
	return bound != nil && product != nil && bound.Sign() >= 0 && product.Sign() > 0 && new(big.Int).Lsh(new(big.Int).Set(bound), 1).Cmp(product) < 0
}

func addValues(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] + b[i]
	}
	return out
}

func multiply(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] * b[i]
	}
	return out
}

func rotateLeft(values []complex128, k int) []complex128 {
	out := make([]complex128, len(values))
	shift := k % len(values)
	if shift < 0 {
		shift += len(values)
	}
	for i := range out {
		out[i] = values[(i+shift)%len(values)]
	}
	return out
}

func compareValues(expected, actual []complex128) metric {
	out := metric{Finite: len(expected) > 0 && len(expected) == len(actual), SNRState: "finite"}
	if !out.Finite {
		return out
	}
	var errorPower, signalPower float64
	for i := range expected {
		if math.IsNaN(real(expected[i])) || math.IsNaN(imag(expected[i])) || math.IsNaN(real(actual[i])) || math.IsNaN(imag(actual[i])) || math.IsInf(real(actual[i]), 0) || math.IsInf(imag(actual[i]), 0) {
			out.Finite = false
			continue
		}
		err := cmplx.Abs(expected[i] - actual[i])
		out.MaxComplex = math.Max(out.MaxComplex, err)
		errorPower += err * err
		signalPower += real(expected[i])*real(expected[i]) + imag(expected[i])*imag(expected[i])
	}
	out.ErrorPower, out.SignalPower = errorPower, signalPower
	out.ComplexRMSE = math.Sqrt(errorPower / float64(len(expected)))
	switch {
	case !out.Finite || math.IsNaN(errorPower) || math.IsInf(errorPower, 0) || math.IsNaN(signalPower) || math.IsInf(signalPower, 0):
		out.Finite, out.SNRState = false, "non_finite_aggregate"
	case errorPower == 0 && signalPower > 0:
		out.SNRState = "positive_infinity_zero_error"
	case errorPower == 0 && signalPower == 0:
		out.SNRState = "undefined_zero_signal_and_error"
	case signalPower == 0:
		out.SNRState = "negative_infinity_zero_signal"
	default:
		snr := 10 * math.Log10(signalPower/errorPower)
		out.SNRDB = &snr
	}
	return out
}

func pairedValues(standard, fast []complex128) (pairedMetric, error) {
	if len(standard) == 0 || len(standard) != len(fast) {
		return pairedMetric{}, errors.New("paired decoded slot arrays are empty or mismatched")
	}
	m := compareValues(standard, fast)
	return pairedMetric{RMSE: m.ComplexRMSE, Max: m.MaxComplex, SNRDB: m.SNRDB, Pass: m.Finite && m.MaxComplex <= bootstrapErrorGate}, nil
}

func subtractLookups(after, before keyLookups) keyLookups {
	return keyLookups{GaloisKey: after.GaloisKey - before.GaloisKey, GaloisKeyList: after.GaloisKeyList - before.GaloisKeyList, Relinearize: after.Relinearize - before.Relinearize}
}

func secretPrefixEqual(a, b *rlwe.SecretKey, rows int) bool {
	if a == nil || b == nil || len(a.Value.Q.Coeffs) < rows || len(b.Value.Q.Coeffs) < rows {
		return false
	}
	for row := 0; row < rows; row++ {
		if !slices.Equal(a.Value.Q.Coeffs[row], b.Value.Q.Coeffs[row]) {
			return false
		}
	}
	return true
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

func hashComplexValues(values []complex128) string {
	h := sha256.New()
	var bytes [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(bytes[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(bytes[8:], math.Float64bits(imag(value)))
		_, _ = h.Write(bytes[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashUint64s(values []uint64) string {
	h := sha256.New()
	var bytes [8]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(bytes[:], value)
		_, _ = h.Write(bytes[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashBytes(values []byte) string { h := sha256.Sum256(values); return hex.EncodeToString(h[:]) }

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
