package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	publicStandardSHA        = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	publicFastSHA            = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac"
	publicConfigSHA          = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"
	publicInputSHA           = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
	publicWorkloadSHA        = "00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74"
	publicQPSHA              = "1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b"
	publicNumericalGate      = 1e-6
	publicTerminalLevel      = 0
	publicMulRescaleInputLvl = 5
)

type publicBackend struct {
	name            string
	residual        ckks.Parameters
	full            ckks.Parameters
	params          bootstrapping.Parameters
	secret          *rlwe.SecretKey
	bootstrapSecret *rlwe.SecretKey
	evaluator       *ckks.Evaluator
	bootstrapEval   *bootstrapping.Evaluator
	dispatch        string
	construction    []phaseTiming
	maxSlots        int
}

type publicCheckpoint struct {
	Name               string               `json:"name"`
	State              ciphertextState      `json:"state"`
	DecodePath         string               `json:"decode_path"`
	PhysicalRowLengths [][]int              `json:"physical_row_lengths_by_component"`
	RowSHA256          [][]string           `json:"active_row_sha256_by_component"`
	FastCompactPrefix  bool                 `json:"fast_compact_prefix_layout"`
	C1Nonzero          bool                 `json:"c1_nonzero"`
	DecodedSHA256      string               `json:"decoded_sha256"`
	Oracle             vectorMetrics        `json:"oracle_metrics"`
	OracleSNR          numericalmetrics.SNR `json:"oracle_snr"`
	OraclePass         bool                 `json:"oracle_pass"`
}

type publicNativeDocument struct {
	SchemaVersion           string                          `json:"schema_version"`
	Mode                    string                          `json:"mode"`
	Status                  string                          `json:"status"`
	Timestamp               time.Time                       `json:"timestamp"`
	Profile                 string                          `json:"profile"`
	Backend                 string                          `json:"backend"`
	BackendCommit           string                          `json:"backend_commit"`
	BackendRef              string                          `json:"backend_ref"`
	BackendCheckoutRef      string                          `json:"backend_checkout_ref"`
	BackendPath             string                          `json:"backend_path"`
	BackendClean            bool                            `json:"backend_clean"`
	PrimaryPath             string                          `json:"primary_path"`
	PrimaryCommit           string                          `json:"primary_commit"`
	PrimaryClean            bool                            `json:"primary_clean"`
	PrimarySourceSHA256     string                          `json:"primary_measurement_source_sha256"`
	BuildTag                string                          `json:"build_tag"`
	GoVersion               string                          `json:"go_version"`
	OS                      string                          `json:"os"`
	Arch                    string                          `json:"arch"`
	CPU                     string                          `json:"cpu"`
	NumCPU                  int                             `json:"num_cpu"`
	GOMAXPROCS              int                             `json:"gomaxprocs"`
	GOGC                    string                          `json:"gogc"`
	GOMEMLIMIT              string                          `json:"gomemlimit"`
	GODEBUG                 string                          `json:"godebug"`
	ConfigPath              string                          `json:"config_path"`
	ConfigSHA256            string                          `json:"config_sha256"`
	QPSHA256                string                          `json:"qp_sha256"`
	InputSHA256             string                          `json:"canonical_input_sha256"`
	WorkloadSHA256          string                          `json:"workload_sha256"`
	Parameters              perfmeasure.EffectiveParameters `json:"effective_parameters"`
	EphemeralSecretWeight   int                             `json:"ephemeral_secret_weight"`
	InputSlots              int                             `json:"input_slots"`
	InputKind               string                          `json:"input_kind"`
	InputConstructor        string                          `json:"input_constructor"`
	InputC1Nonzero          bool                            `json:"input_c1_nonzero"`
	PublicEvaluatorDispatch string                          `json:"public_evaluator_dispatch"`
	PreflightPairSHA256     string                          `json:"preflight_pair_sha256,omitempty"`
	Capacity                perfmeasure.PublicCapacityPlan  `json:"capacity_plan"`
	NumericalGate           float64                         `json:"max_complex_numerical_gate"`
	BootstrapBudget         int                             `json:"bootstrap_budget"`
	BootstrapCalls          int                             `json:"actual_bootstrap_calls"`
	ExecutionPhases         []phaseTiming                   `json:"execution_phases"`
	Checkpoints             []publicCheckpoint              `json:"checkpoints"`
	BootstrapResults        []publicBootstrapResult         `json:"bootstrap_results,omitempty"`
	Limitations             []string                        `json:"limitations"`
}

type publicVectorDocument struct {
	SchemaVersion    string                    `json:"schema_version"`
	Mode             string                    `json:"mode"`
	Backend          string                    `json:"backend"`
	BackendCommit    string                    `json:"backend_commit"`
	PrimaryCommit    string                    `json:"primary_commit"`
	SourceSHA256     string                    `json:"primary_measurement_source_sha256"`
	ConfigSHA256     string                    `json:"config_sha256"`
	QPSHA256         string                    `json:"qp_sha256"`
	InputSHA256      string                    `json:"input_sha256"`
	WorkloadSHA256   string                    `json:"workload_sha256"`
	E                int                       `json:"ephemeral_secret_weight"`
	Checkpoints      map[string][]complexValue `json:"checkpoints"`
	BootstrapOutputs map[string][]complexValue `json:"bootstrap_outputs,omitempty"`
}

type publicBootstrapResult struct {
	Phase                 string           `json:"phase"`
	InputCiphertextSHA256 string           `json:"input_ciphertext_sha256"`
	Timing                phaseTiming      `json:"timing"`
	Output                publicCheckpoint `json:"output_checkpoint"`
}

type publicTraceFixtureManifest struct {
	SchemaVersion               string                          `json:"schema_version"`
	GeneratedAt                 time.Time                       `json:"generated_at"`
	Profile                     string                          `json:"profile"`
	Backend                     string                          `json:"backend"`
	BackendCommit               string                          `json:"backend_commit"`
	PrimaryCommit               string                          `json:"primary_commit"`
	PrimarySourceSHA256         string                          `json:"primary_measurement_source_sha256"`
	ConfigSHA256                string                          `json:"config_sha256"`
	QPSHA256                    string                          `json:"qp_sha256"`
	InputSHA256                 string                          `json:"canonical_input_sha256"`
	WorkloadSHA256              string                          `json:"workload_sha256"`
	ParametersFile              string                          `json:"parameters_file"`
	ParametersSHA256            string                          `json:"parameters_sha256"`
	CiphertextFile              string                          `json:"ciphertext_file"`
	CiphertextSHA256            string                          `json:"ciphertext_sha256"`
	CiphertextFingerprintSHA256 string                          `json:"ciphertext_fingerprint_sha256"`
	DecodedSHA256               string                          `json:"decoded_sha256"`
	State                       ciphertextState                 `json:"ciphertext_state"`
	PhysicalRowLengths          [][]int                         `json:"physical_row_lengths"`
	IsNTT                       bool                            `json:"is_ntt"`
	IsMontgomery                bool                            `json:"is_montgomery"`
	C1Nonzero                   bool                            `json:"c1_nonzero"`
	NumericalGate               float64                         `json:"max_complex_numerical_gate"`
	EphemeralSecretWeight       int                             `json:"ephemeral_secret_weight"`
	Parameters                  perfmeasure.EffectiveParameters `json:"effective_parameters"`
}

type publicBootstrapFailureDocument struct {
	SchemaVersion       string                  `json:"schema_version"`
	Status              string                  `json:"status"`
	Backend             string                  `json:"backend"`
	BackendCommit       string                  `json:"backend_commit"`
	PrimaryCommit       string                  `json:"primary_commit"`
	PrimarySourceSHA256 string                  `json:"primary_measurement_source_sha256"`
	PreflightPairSHA256 string                  `json:"preflight_pair_sha256"`
	ConfigSHA256        string                  `json:"config_sha256"`
	QPSHA256            string                  `json:"qp_sha256"`
	InputSHA256         string                  `json:"input_sha256"`
	WorkloadSHA256      string                  `json:"workload_sha256"`
	BootstrapBudget     int                     `json:"bootstrap_budget"`
	BootstrapCalls      int                     `json:"reserved_bootstrap_calls"`
	FailedPhase         string                  `json:"failed_phase,omitempty"`
	Failure             string                  `json:"failure"`
	ExecutionPhases     []phaseTiming           `json:"execution_phases,omitempty"`
	CompletedResults    []publicBootstrapResult `json:"completed_results,omitempty"`
}

type publicPairCheckpoint struct {
	Name          string          `json:"name"`
	StateMatched  bool            `json:"level_scale_degree_matched"`
	FastState     ciphertextState `json:"fast_state"`
	StandardState ciphertextState `json:"standard_state"`
	Metrics       vectorMetrics   `json:"fast_vs_standard"`
	Pass          bool            `json:"pass"`
}

type publicPairBootstrapComparison struct {
	Phase         string          `json:"phase"`
	StateMatched  bool            `json:"level_scale_degree_matched"`
	FastState     ciphertextState `json:"fast_state"`
	StandardState ciphertextState `json:"standard_state"`
	Metrics       vectorMetrics   `json:"fast_vs_standard"`
	Pass          bool            `json:"pass"`
}

type publicPairDocument struct {
	SchemaVersion      string                          `json:"schema_version"`
	GeneratedAt        time.Time                       `json:"generated_at"`
	Status             string                          `json:"status"`
	NumericalGate      float64                         `json:"max_complex_gate"`
	FastCommit         string                          `json:"fast_backend_commit"`
	StandardCommit     string                          `json:"standard_backend_commit"`
	PrimaryCommit      string                          `json:"primary_commit"`
	SourceSHA256       string                          `json:"primary_measurement_source_sha256"`
	InputSHA256        string                          `json:"input_sha256"`
	WorkloadSHA256     string                          `json:"workload_sha256"`
	ConfigSHA256       string                          `json:"config_sha256"`
	QPSHA256           string                          `json:"qp_sha256"`
	E                  int                             `json:"ephemeral_secret_weight"`
	BootstrapCalls     int                             `json:"actual_bootstrap_calls"`
	MatchedEnvironment bool                            `json:"matched_environment"`
	Checkpoints        []publicPairCheckpoint          `json:"checkpoints"`
	Bootstrap          []publicPairBootstrapComparison `json:"bootstrap_comparisons,omitempty"`
	Limitations        []string                        `json:"limitations"`
}

func runPublicNative(opts cliOptions, configHash string, residual ckks.Parameters, params bootstrapping.Parameters, effective perfmeasure.EffectiveParameters) error {
	if err := ensurePublicNativeOutputsAvailable(opts); err != nil {
		return err
	}
	if opts.profile != "logn13" || opts.ephemeralSecretWeight != 32 || effective.InputSlots != 1<<12 {
		return errors.New("public-native Batch021 fixture is frozen to LogN13, 4096 slots, and E=32")
	}
	if configHash != publicConfigSHA {
		return fmt.Errorf("canonical Batch019/020 config SHA mismatch: got %s want %s", configHash, publicConfigSHA)
	}
	if err := validatePublicEphemeralWeight(opts.ephemeralSecretWeight); err != nil {
		return err
	}
	if err := verifyCompiledBackendSource(opts.secondaryRoot); err != nil {
		return err
	}
	primaryRoot, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		return fmt.Errorf("Primary provenance: %w", err)
	}
	secondaryRoot, secondaryCommit, secondaryRef, err := cleanSecondaryState(opts.secondaryRoot, opts.backendCommit)
	if err != nil {
		return fmt.Errorf("backend provenance: %w", err)
	}
	if err := validatePublicBackendPin(publicBackendName(), secondaryCommit); err != nil {
		return err
	}
	sourceHash, err := primaryMeasurementSourceFingerprint(primaryRoot)
	if err != nil {
		return err
	}
	valuesA, valuesB, valuesC, err := perfmeasure.PublicMulRescaleWorkload(effective.InputSlots)
	if err != nil {
		return err
	}
	inputHash := perfmeasure.Fingerprint(valuesA)
	if inputHash != publicInputSHA {
		return fmt.Errorf("canonical Batch019/020 input fingerprint mismatch: %s", inputHash)
	}
	full := params.BootstrappingParameters
	q, p := full.Q(), full.P()
	qpBytes, err := json.Marshal(struct{ Q, P []uint64 }{q, p})
	if err != nil {
		return err
	}
	qpSum := sha256.Sum256(qpBytes)
	qpHash := hex.EncodeToString(qpSum[:])
	if qpHash != publicQPSHA {
		return fmt.Errorf("canonical Batch019/020 Q/P fingerprint mismatch: %s", qpHash)
	}
	q5Scale := rlwe.NewScale(q[5])
	workloadHash, err := perfmeasure.PublicMulRescaleWorkloadFingerprint(valuesA, valuesB, valuesC, q5Scale.BigInt().String())
	if err != nil {
		return err
	}
	if workloadHash != publicWorkloadSHA {
		return fmt.Errorf("canonical Batch019 workload fingerprint mismatch: %s", workloadHash)
	}
	capacity, err := perfmeasure.BoundPublicMulRescaleWorkload(valuesA, valuesB, valuesC, q, full.N())
	if err != nil {
		return err
	}
	preflightPairSHA := ""
	if opts.publicBootstrap || opts.publicRepeatability {
		preflightPairSHA, err = validatePublicPreflightGate(opts.preflightPair, primaryCommit, sourceHash, configHash, qpHash, inputHash, workloadHash)
		if err != nil {
			return fmt.Errorf("public Bootstrap preflight gate: %w", err)
		}
	}

	backend, err := newPublicBackend(params, residual)
	if err != nil {
		return err
	}
	if backend.name != publicBackendName() {
		return fmt.Errorf("build tag/backend mismatch: adapter=%s selected=%s", backend.name, publicBackendName())
	}
	defaultScale := residual.DefaultScale()
	if full.LogDefaultScale() != 45 || residual.MaxLevel() != 1 || full.MaxLevel() != 16 || len(q) <= 5 {
		return errors.New("effective public-native parameters do not match frozen LogN13/E32 fixture")
	}
	phaseTimings := append([]phaseTiming(nil), backend.construction...)
	checkpoints := make([]publicCheckpoint, 0, 8)
	vectorMap := make(map[string][]complexValue, 8)
	decodedByCheckpoint := make(map[string][]complex128, 8)
	addCheckpoint := func(name string, ct *rlwe.Ciphertext, expected []complex128, fastCompactOutput bool) error {
		var decodePath string
		decoded, phase, decodeErr := measurePublicPhase("checkpoint_decode_"+name, func() ([]complex128, error) {
			values, path, err := backend.decode(ct)
			decodePath = path
			return values, err
		})
		if decodeErr != nil {
			return fmt.Errorf("native Level/Scale decode %s: %w", name, decodeErr)
		}
		phaseTimings = append(phaseTimings, phase)
		cp, cpErr := makePublicCheckpoint(backend, name, ct, decoded, expected, q, fastCompactOutput)
		if cpErr != nil {
			return cpErr
		}
		cp.DecodePath = decodePath
		if !cp.OraclePass {
			return fmt.Errorf("%s cleartext oracle max %.12g exceeds %.12g", name, cp.Oracle.MaxComplexDifference, publicNumericalGate)
		}
		checkpoints = append(checkpoints, cp)
		vectorMap[name] = encodeVector(decoded)
		decodedByCheckpoint[name] = append([]complex128(nil), decoded...)
		return nil
	}

	ctA, phase, err := backend.encryptAtLevel(valuesA, defaultScale, publicMulRescaleInputLvl, effective.LogSlots)
	phase.Phase = "public_encrypt_a_level5"
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public EncryptNew(A): %w", err)
	}
	ctB, phase, err := backend.encryptAtLevel(valuesB, defaultScale, publicMulRescaleInputLvl, effective.LogSlots)
	phase.Phase = "public_encrypt_b_level5"
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public EncryptNew(B): %w", err)
	}
	ctC, phase, err := backend.encryptAtLevel(valuesC, q5Scale, publicMulRescaleInputLvl, effective.LogSlots)
	phase.Phase = "public_encrypt_c_q5_level5"
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public EncryptNew(C,q5 scale): %w", err)
	}
	for _, item := range []struct {
		name string
		ct   *rlwe.Ciphertext
	}{{"A", ctA}, {"B", ctB}, {"C", ctC}} {
		name, ct := item.name, item.ct
		if ct == nil || ct.Level() != publicMulRescaleInputLvl || ct.Degree() != 1 || !ct.IsNTT || ct.IsMontgomery || ct.LogDimensions.Cols != effective.LogSlots {
			return fmt.Errorf("public EncryptNew(%s) violates Level5/degree/domain/dimension contract", name)
		}
	}
	if !ctA.Scale.Equal(defaultScale) || !ctB.Scale.Equal(defaultScale) || !ctC.Scale.Equal(q5Scale) {
		return errors.New("public EncryptNew inputs do not have A/B=2^45 and C=q5 Scale")
	}
	c1Nonzero, err := isCiphertextComponentNonzero(ctA, 1)
	if err != nil {
		return err
	}
	if backend.name == "fast" && c1Nonzero {
		return errors.New("Fast public EncryptNew did not preserve the approved zero-secret c1=0 mode")
	}
	if backend.name == "standard" && !c1Nonzero {
		return errors.New("genuine Standard EncryptNew unexpectedly produced a zero-c1 input")
	}

	if err := addCheckpoint("encrypt_a_level5", ctA, valuesA, false); err != nil {
		return err
	}
	if err := addCheckpoint("encrypt_b_level5", ctB, valuesB, false); err != nil {
		return err
	}
	if err := addCheckpoint("encrypt_c_q5_level5", ctC, valuesC, false); err != nil {
		return err
	}
	add, phase, err := measurePublicPhase("public_AddNew", func() (*rlwe.Ciphertext, error) {
		return backend.evaluator.AddNew(ctA, ctB)
	})
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public AddNew: %w", err)
	}
	if add.Level() != 5 || !add.Scale.Equal(defaultScale) {
		return errors.New("AddNew changed expected Level5/Scale2^45")
	}
	expectedAdd := publicAdd(valuesA, valuesB)
	if err := addCheckpoint("add_level5", add, expectedAdd, true); err != nil {
		return err
	}

	product, phase, err := measurePublicPhase("public_MulRelinNew", func() (*rlwe.Ciphertext, error) {
		return backend.evaluator.MulRelinNew(add, ctC)
	})
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public MulRelinNew: %w", err)
	}
	mulScale := rlwe.NewScale(new(big.Int).Mul(defaultScale.BigInt(), q5Scale.BigInt()))
	if product.Level() != 5 || !product.Scale.Equal(mulScale) {
		return errors.New("MulRelinNew changed expected Level5/2^45*q5 Scale")
	}
	expectedProduct := publicMul(expectedAdd, valuesC)
	if err := addCheckpoint("mulrelin_level5", product, expectedProduct, true); err != nil {
		return err
	}

	rescaled, phase, err := measurePublicPhase("public_Rescale_q5", func() (*rlwe.Ciphertext, error) {
		out := ckks.NewCiphertext(full, 1, 4)
		if rescaleErr := backend.evaluator.Rescale(product, out); rescaleErr != nil {
			return nil, rescaleErr
		}
		return out, nil
	})
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public Rescale(q5): %w", err)
	}
	if rescaled.Level() != 4 || !rescaled.Scale.Equal(defaultScale) {
		return errors.New("Rescale(q5) did not yield Level4/Scale2^45")
	}
	if err := addCheckpoint("rescale_q5_level4", rescaled, expectedProduct, true); err != nil {
		return err
	}

	rotated, phase, err := measurePublicPhase("public_RotateNew_1", func() (*rlwe.Ciphertext, error) {
		return backend.evaluator.RotateNew(rescaled, 1)
	})
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public RotateNew(1): %w", err)
	}
	if rotated.Level() != 4 || !rotated.Scale.Equal(defaultScale) {
		return errors.New("RotateNew changed expected Level4/Scale2^45")
	}
	expectedRotated := publicRotateLeft(expectedProduct, 1)
	if err := addCheckpoint("rotate_level4", rotated, expectedRotated, true); err != nil {
		return err
	}

	dropped, phase, err := measurePublicPhase("public_DropLevelNew_4", func() (*rlwe.Ciphertext, error) {
		return backend.evaluator.DropLevelNew(rotated, rotated.Level()), nil
	})
	phaseTimings = append(phaseTimings, phase)
	if err != nil {
		return fmt.Errorf("public DropLevelNew(4): %w", err)
	}
	if dropped.Level() != publicTerminalLevel || !dropped.Scale.Equal(defaultScale) {
		return errors.New("DropLevelNew did not yield Level0/Scale2^45")
	}
	if err := addCheckpoint("drop_level0", dropped, expectedRotated, true); err != nil {
		return err
	}
	if opts.traceFixtureOut != "" {
		if backend.name != "fast" {
			return errors.New("E32 diagnostic trace fixture export is Fast-only")
		}
		if err := writePublicTraceFixture(opts.traceFixtureOut, params, dropped, checkpoints[len(checkpoints)-1], effective,
			primaryCommit, sourceHash, secondaryCommit, configHash, qpHash, inputHash, workloadHash); err != nil {
			return fmt.Errorf("export E32 public trace fixture: %w", err)
		}
	}

	bootstrapResults := []publicBootstrapResult(nil)
	bootstrapOutputMap := map[string][]complexValue(nil)
	bootstrapCalls := 0
	schemaVersion, mode, status := "fast-standard-public-native-preflight.v1", "public-native", "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN"
	vectorSchema := "fast-standard-public-native-vectors.v1"
	limitations := []string{
		"This is a public pre-Bootstrap lifecycle only; no Bootstrap method was called, so public Bootstrap acceptance/output and cold/warm timings remain unverified.",
		"Fast zero-c1 is the pinned implementation's declared zero-secret simulation mode, not secure public-key encryption; Standard uses native generated-secret encryption.",
		"Existing fastdiag P93/E0 traces are not E32 evidence. This public-native run performs no in-circuit tracing; the separate test-only E32 trace is a distinct diagnostic lane.",
	}
	if opts.publicBootstrap || opts.publicRepeatability {
		preBootstrapDecoded := decodedByCheckpoint["drop_level0"]
		budget := &bootstrapBudget{limit: opts.bootstrapBudget, journalBase: opts.out}
		var attemptPhases []phaseTiming
		bootstrapResults, bootstrapOutputMap, attemptPhases, err = runPublicBootstrapAttempts(
			dropped,
			budget,
			func(input *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
				return backend.bootstrapEval.Bootstrap(input)
			},
			func(name string, output *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
				if output == nil || output.Level() != residual.MaxLevel() || !output.Scale.Equal(defaultScale) || output.Degree() != 1 || !output.IsNTT || output.IsMontgomery || output.LogDimensions.Cols != effective.LogSlots {
					return publicCheckpoint{}, nil, fmt.Errorf("%s output violates native Level1/Scale2^45/degree/domain/dimension contract", name)
				}
				decoded, decodePath, decodeErr := backend.decodeNativeLevel1(output)
				if decodeErr != nil {
					return publicCheckpoint{}, nil, fmt.Errorf("%s native Level1 DecryptNew/Decode: %w", name, decodeErr)
				}
				checkpoint, checkpointErr := makePublicCheckpoint(backend, name, output, decoded, preBootstrapDecoded, q, true)
				if checkpointErr != nil {
					return publicCheckpoint{}, nil, checkpointErr
				}
				checkpoint.DecodePath = decodePath
				if !checkpoint.OraclePass {
					return publicCheckpoint{}, nil, fmt.Errorf("%s native plaintext oracle max %.12g exceeds %.12g", name, checkpoint.Oracle.MaxComplexDifference, publicNumericalGate)
				}
				return checkpoint, decoded, nil
			},
		)
		if err != nil {
			failedPhase := "pre-call-validation"
			var attemptErr *publicBootstrapAttemptError
			if errors.As(err, &attemptErr) {
				failedPhase = attemptErr.Phase
			}
			failure := publicBootstrapFailureDocument{
				SchemaVersion: "fast-standard-public-native-bootstrap-failure.v1", Status: "STOPPED_NO_RETRY",
				Backend: backend.name, BackendCommit: secondaryCommit, PrimaryCommit: primaryCommit,
				PrimarySourceSHA256: sourceHash, PreflightPairSHA256: preflightPairSHA,
				ConfigSHA256: configHash, QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash,
				BootstrapBudget: opts.bootstrapBudget, BootstrapCalls: budget.attempts, FailedPhase: failedPhase,
				Failure: err.Error(), ExecutionPhases: attemptPhases, CompletedResults: bootstrapResults,
			}
			failurePath := opts.out + ".failure.json"
			if writeErr := writeExclusiveJSON(failurePath, failure); writeErr != nil {
				return fmt.Errorf("public E32 Bootstrap measurement stopped after %d/%d reserved calls: %w (also could not persist failure evidence: %v)", budget.attempts, budget.limit, err, writeErr)
			}
			return fmt.Errorf("public E32 Bootstrap measurement stopped after %d/%d reserved calls: %w", budget.attempts, budget.limit, err)
		}
		bootstrapCalls = budget.attempts
		phaseTimings = append(phaseTimings, attemptPhases...)
		if opts.publicRepeatability {
			schemaVersion, mode, status = "fast-standard-public-native-repeatability.v1", "public-native-repeatability", "PASS_PUBLIC_BOOTSTRAP_1COLD_5WARM"
			vectorSchema = "fast-standard-public-native-repeatability-vectors.v1"
			limitations = []string{
				"One cold plus five warm samples describe this pinned LogN13/E32 profile; they do not establish secure-FHE equivalence or general hardware performance.",
				"Standard uses genuine native key generation/encryption and Fast uses the pinned intentionally insecure zero-secret mode.",
				"Fast internal trace is a separate instrumented diagnostic lane; its timing is not included in these uninstrumented samples.",
			}
		} else {
			schemaVersion, mode, status = "fast-standard-public-native-bootstrap.v1", "public-native-bootstrap", "PASS_PUBLIC_BOOTSTRAP_COLD_WARM"
			vectorSchema = "fast-standard-public-native-bootstrap-vectors.v1"
			limitations = []string{
				"Exactly one cold and one warm Bootstrap sample are descriptive observations, not a statistically stable speedup estimate.",
				"Standard uses genuine native key generation/encryption and Fast uses the pinned intentionally insecure zero-secret mode; this does not establish security equivalence.",
				"Fast E32 in-circuit internal tracing remains separate from these uninstrumented public-call timings.",
			}
		}
	} else {
		phaseTimings = append(phaseTimings,
			phaseTiming{Phase: "first_cold_bootstrap", Available: false, Reason: "zero-call public-native preflight does not invoke Bootstrap"},
			phaseTiming{Phase: "later_warm_bootstrap", Available: false, Reason: "zero-call public-native preflight does not invoke Bootstrap"},
		)
	}
	inputKind, constructor := publicInputContract(backend.name)
	doc := publicNativeDocument{
		SchemaVersion: schemaVersion, Mode: mode, Status: status,
		Timestamp: time.Now().UTC(), Profile: opts.profile, Backend: backend.name, BackendCommit: secondaryCommit,
		BackendRef: opts.backendRef, BackendCheckoutRef: secondaryRef, BackendPath: secondaryRoot, BackendClean: true, PrimaryPath: primaryRoot,
		PrimaryCommit: primaryCommit, PrimaryClean: true, PrimarySourceSHA256: sourceHash, BuildTag: "perf_" + backend.name,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
		CPU: cpuModel(), NumCPU: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0),
		GOGC: environmentSetting("GOGC"), GOMEMLIMIT: environmentSetting("GOMEMLIMIT"), GODEBUG: environmentSetting("GODEBUG"),
		ConfigPath: opts.config, ConfigSHA256: configHash, QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash,
		Parameters: effective, EphemeralSecretWeight: params.EphemeralSecretWeight, InputSlots: effective.InputSlots,
		InputKind: inputKind, InputConstructor: constructor, InputC1Nonzero: c1Nonzero, PublicEvaluatorDispatch: backend.dispatch,
		PreflightPairSHA256: preflightPairSHA, Capacity: capacity, NumericalGate: publicNumericalGate,
		BootstrapBudget: opts.bootstrapBudget, BootstrapCalls: bootstrapCalls,
		ExecutionPhases: phaseTimings, Checkpoints: checkpoints, BootstrapResults: bootstrapResults,
		Limitations: limitations,
	}
	vectorDoc := publicVectorDocument{
		SchemaVersion: vectorSchema, Mode: mode, Backend: backend.name,
		BackendCommit: secondaryCommit, PrimaryCommit: primaryCommit, SourceSHA256: sourceHash, ConfigSHA256: configHash,
		QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash, E: params.EphemeralSecretWeight,
		Checkpoints: vectorMap, BootstrapOutputs: bootstrapOutputMap,
	}
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		return err
	}
	if err := writeExclusiveJSON(opts.vectorsOut, vectorDoc); err != nil {
		return err
	}
	fmt.Printf("%s %s mode=%s result=%s checkpoints=%d bootstrap_calls=%d\n", opts.profile, backend.name, mode, opts.out, len(checkpoints), bootstrapCalls)
	return nil
}

func ensurePublicNativeOutputsAvailable(opts cliOptions) error {
	paths := []string{opts.out, opts.vectorsOut}
	if opts.publicBootstrap || opts.publicRepeatability {
		paths = append(paths, opts.out+".failure.json")
		for index := 1; index <= opts.bootstrapBudget; index++ {
			paths = append(paths, fmt.Sprintf("%s.bootstrap-attempt-%02d.json", opts.out, index))
		}
	}
	if opts.traceFixtureOut != "" {
		ciphertextPath, parametersPath, manifestPath := publicTraceFixtureArtifactPaths(opts.traceFixtureOut)
		paths = append(paths, ciphertextPath, parametersPath, manifestPath)
	}
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve public-native output path %q: %w", path, err)
		}
		absolute = filepath.Clean(absolute)
		parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
		if err != nil {
			return fmt.Errorf("resolve public-native output directory for %s: %w", absolute, err)
		}
		absolute = filepath.Join(parent, filepath.Base(absolute))
		if _, exists := seen[absolute]; exists {
			return fmt.Errorf("public-native output paths alias the same file: %s", absolute)
		}
		seen[absolute] = struct{}{}
		if _, err := os.Lstat(absolute); err == nil {
			return fmt.Errorf("public-native output path already exists: %s", absolute)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect public-native output path %s: %w", absolute, err)
		}
	}
	return nil
}

func writePublicTraceFixture(
	base string,
	params bootstrapping.Parameters,
	ciphertext *rlwe.Ciphertext,
	checkpoint publicCheckpoint,
	effective perfmeasure.EffectiveParameters,
	primaryCommit, sourceHash, backendCommit, configHash, qpHash, inputHash, workloadHash string,
) error {
	if ciphertext == nil || ciphertext.MetaData == nil || checkpoint.Name != "drop_level0" || checkpoint.State.Level != 0 ||
		checkpoint.C1Nonzero || !ciphertext.IsNTT || ciphertext.IsMontgomery || ciphertext.Degree() != 1 || ciphertext.Level() != 0 {
		return errors.New("held Fast trace fixture is not the validated compact Level0 zero-c1 public ciphertext")
	}
	parameterBytes, err := params.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal public bootstrapping parameters: %w", err)
	}
	ciphertextBytes, err := ciphertext.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal held public ciphertext: %w", err)
	}
	fingerprint, err := publicCiphertextFingerprint(ciphertext)
	if err != nil {
		return err
	}
	parameterSum, ciphertextSum := sha256.Sum256(parameterBytes), sha256.Sum256(ciphertextBytes)
	rowLengths := make([][]int, len(ciphertext.Value))
	for component := range ciphertext.Value {
		rowLengths[component] = make([]int, len(ciphertext.Value[component].Coeffs))
		for row := range ciphertext.Value[component].Coeffs {
			rowLengths[component][row] = len(ciphertext.Value[component].Coeffs[row])
		}
	}
	ciphertextPath, parameterPath, manifestPath := publicTraceFixtureArtifactPaths(base)
	parameterName, ciphertextName := filepath.Base(parameterPath), filepath.Base(ciphertextPath)
	if err := writeExclusiveBytes(parameterPath, parameterBytes, 0o600); err != nil {
		return fmt.Errorf("write parameter manifest: %w", err)
	}
	if err := writeExclusiveBytes(ciphertextPath, ciphertextBytes, 0o600); err != nil {
		return fmt.Errorf("write held ciphertext: %w", err)
	}
	manifest := publicTraceFixtureManifest{
		SchemaVersion: "fast-public-e32-trace-fixture.v1", GeneratedAt: time.Now().UTC(), Profile: "logn13-e32-public-native",
		Backend: "fast", BackendCommit: backendCommit, PrimaryCommit: primaryCommit, PrimarySourceSHA256: sourceHash,
		ConfigSHA256: configHash, QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash,
		ParametersFile: parameterName, ParametersSHA256: hex.EncodeToString(parameterSum[:]),
		CiphertextFile: ciphertextName, CiphertextSHA256: hex.EncodeToString(ciphertextSum[:]),
		CiphertextFingerprintSHA256: fingerprint, DecodedSHA256: checkpoint.DecodedSHA256, State: checkpoint.State,
		PhysicalRowLengths: rowLengths, IsNTT: ciphertext.IsNTT, IsMontgomery: ciphertext.IsMontgomery,
		C1Nonzero: checkpoint.C1Nonzero, NumericalGate: publicNumericalGate, EphemeralSecretWeight: 32, Parameters: effective,
	}
	return writeExclusiveJSON(manifestPath, manifest)
}

func publicTraceFixtureArtifactPaths(base string) (ciphertext, parameters, manifest string) {
	return base + ".ciphertext.bin", base + ".parameters.bin", base + ".manifest.json"
}

func writeExclusiveBytes(path string, data []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

type publicBootstrapOutputValidator func(name string, output *rlwe.Ciphertext) (publicCheckpoint, []complex128, error)

type publicBootstrapAttemptError struct {
	Phase string
	Err   error
}

func (err *publicBootstrapAttemptError) Error() string { return err.Err.Error() }

func (err *publicBootstrapAttemptError) Unwrap() error { return err.Err }

func runTwoPublicBootstrapAttempts(
	input *rlwe.Ciphertext,
	budget *bootstrapBudget,
	call func(*rlwe.Ciphertext) (*rlwe.Ciphertext, error),
	validate publicBootstrapOutputValidator,
) ([]publicBootstrapResult, map[string][]complexValue, []phaseTiming, error) {
	if budget == nil || budget.limit != 2 {
		return nil, nil, nil, errors.New("cold/warm public Bootstrap requires an exact two-attempt budget")
	}
	return runPublicBootstrapAttempts(input, budget, call, validate)
}

func publicBootstrapAttemptSpecs(limit int) ([]struct{ phase, name string }, error) {
	if limit != 2 && limit != 6 {
		return nil, fmt.Errorf("public Bootstrap attempts require an exact budget of 2 or 6, got %d", limit)
	}
	attempts := make([]struct{ phase, name string }, limit)
	attempts[0] = struct{ phase, name string }{phase: "first_cold_bootstrap", name: "bootstrap_first_cold"}
	if limit == 2 {
		attempts[1] = struct{ phase, name string }{phase: "later_warm_bootstrap", name: "bootstrap_later_warm"}
		return attempts, nil
	}
	for index := 1; index < limit; index++ {
		attempts[index] = struct{ phase, name string }{
			phase: fmt.Sprintf("warm_bootstrap_%02d", index), name: fmt.Sprintf("bootstrap_warm_%02d", index),
		}
	}
	return attempts, nil
}

func runPublicBootstrapAttempts(
	input *rlwe.Ciphertext,
	budget *bootstrapBudget,
	call func(*rlwe.Ciphertext) (*rlwe.Ciphertext, error),
	validate publicBootstrapOutputValidator,
) ([]publicBootstrapResult, map[string][]complexValue, []phaseTiming, error) {
	if input == nil || budget == nil || call == nil || validate == nil {
		return nil, nil, nil, errors.New("public Bootstrap requires an input, validator, caller, and explicit attempt budget")
	}
	attempts, err := publicBootstrapAttemptSpecs(budget.limit)
	if err != nil {
		return nil, nil, nil, err
	}
	inputSHA256, err := publicCiphertextFingerprint(input)
	if err != nil {
		return nil, nil, nil, err
	}
	results := make([]publicBootstrapResult, 0, budget.limit)
	outputs := make(map[string][]complexValue, budget.limit)
	phases := make([]phaseTiming, 0, budget.limit)
	for _, attempt := range attempts {
		callInput := input.CopyNew()
		copySHA256, err := publicCiphertextFingerprint(callInput)
		if err != nil {
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		if copySHA256 != inputSHA256 {
			err := fmt.Errorf("%s input copy differs from the held pre-Bootstrap ciphertext", attempt.phase)
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		var timing phaseTiming
		var output *rlwe.Ciphertext
		output, err = budget.invoke(attempt.phase, func() (*rlwe.Ciphertext, error) {
			measuredOutput, measured, callErr := measurePublicPhase(attempt.phase, func() (*rlwe.Ciphertext, error) {
				return call(callInput)
			})
			timing = measured
			return measuredOutput, callErr
		})
		if timing.Available {
			phases = append(phases, timing)
		}
		if err != nil {
			if afterSHA256, fingerprintErr := publicCiphertextFingerprint(input); fingerprintErr != nil {
				err = fmt.Errorf("%s Bootstrap failed (%v); held input fingerprint check failed: %w", attempt.phase, err, fingerprintErr)
			} else if afterSHA256 != inputSHA256 {
				err = fmt.Errorf("%s Bootstrap failed (%v) and mutated the held pre-Bootstrap input", attempt.phase, err)
			} else {
				err = fmt.Errorf("%s Bootstrap attempt: %w", attempt.phase, err)
			}
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		afterSHA256, err := publicCiphertextFingerprint(input)
		if err != nil {
			err = fmt.Errorf("%s held-input verification: %w", attempt.phase, err)
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		if afterSHA256 != inputSHA256 {
			err = fmt.Errorf("%s Bootstrap mutated the held pre-Bootstrap input", attempt.phase)
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		checkpoint, decoded, err := validate(attempt.name, output)
		if err != nil {
			err = fmt.Errorf("%s output validation: %w", attempt.phase, err)
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		if checkpoint.Name != attempt.name || len(decoded) != 1<<12 || !allFinite(decoded) {
			err = fmt.Errorf("%s output validator returned an invalid checkpoint/vector", attempt.phase)
			return results, outputs, phases, &publicBootstrapAttemptError{Phase: attempt.phase, Err: err}
		}
		results = append(results, publicBootstrapResult{Phase: attempt.phase, InputCiphertextSHA256: inputSHA256, Timing: timing, Output: checkpoint})
		outputs[checkpoint.Name] = encodeVector(decoded)
	}
	if budget.attempts != budget.limit {
		return results, outputs, phases, fmt.Errorf("public Bootstrap consumed %d calls, want exactly %d", budget.attempts, budget.limit)
	}
	return results, outputs, phases, nil
}

func publicCiphertextFingerprint(ct *rlwe.Ciphertext) (string, error) {
	if ct == nil || ct.MetaData == nil || ct.Level() < 0 || ct.Degree() < 0 {
		return "", errors.New("ciphertext fingerprint requires valid ciphertext metadata")
	}
	h := sha256.New()
	var encoded [8]byte
	writeUint64 := func(value uint64) {
		binary.LittleEndian.PutUint64(encoded[:], value)
		_, _ = h.Write(encoded[:])
	}
	writeString := func(value string) {
		writeUint64(uint64(len(value)))
		_, _ = h.Write([]byte(value))
	}
	writeUint64(uint64(ct.Level()))
	writeUint64(uint64(ct.Degree()))
	writeString(ct.Scale.Value.Text('e', 80))
	writeUint64(uint64(ct.LogDimensions.Rows))
	writeUint64(uint64(ct.LogDimensions.Cols))
	if ct.IsNTT {
		writeUint64(1)
	} else {
		writeUint64(0)
	}
	if ct.IsMontgomery {
		writeUint64(1)
	} else {
		writeUint64(0)
	}
	writeUint64(uint64(len(ct.Value)))
	for component, polynomial := range ct.Value {
		writeUint64(uint64(component))
		writeUint64(uint64(len(polynomial.Coeffs)))
		for row, coefficients := range polynomial.Coeffs {
			writeUint64(uint64(row))
			writeUint64(uint64(len(coefficients)))
			for _, coefficient := range coefficients {
				writeUint64(coefficient)
			}
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func validatePublicPreflightGate(path, primaryCommit, sourceSHA256, configSHA256, qpSHA256, inputSHA256, workloadSHA256 string) (string, error) {
	var pair publicPairDocument
	if err := readJSON(path, &pair); err != nil {
		return "", fmt.Errorf("read passing compare-public pair %s: %w", path, err)
	}
	if pair.SchemaVersion != "fast-standard-public-native-pair.v1" || pair.Status != "PASS" || !pair.MatchedEnvironment || pair.BootstrapCalls != 0 || len(pair.Bootstrap) != 0 || pair.E != 32 || pair.NumericalGate != publicNumericalGate {
		return "", errors.New("preflight pair is not a passing zero-Bootstrap LogN13/E32 pair")
	}
	if pair.PrimaryCommit != primaryCommit || pair.SourceSHA256 != sourceSHA256 || pair.StandardCommit != publicStandardSHA || pair.FastCommit != publicFastSHA ||
		pair.ConfigSHA256 != configSHA256 || pair.ConfigSHA256 != publicConfigSHA || pair.QPSHA256 != qpSHA256 || pair.QPSHA256 != publicQPSHA ||
		pair.InputSHA256 != inputSHA256 || pair.InputSHA256 != publicInputSHA || pair.WorkloadSHA256 != workloadSHA256 || pair.WorkloadSHA256 != publicWorkloadSHA {
		return "", errors.New("preflight pair provenance does not match the current Primary source, pins, config, Q/P, input, and workload")
	}
	wantNames := []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"}
	if len(pair.Checkpoints) != len(wantNames) {
		return "", errors.New("preflight pair does not contain the exact eight canonical checkpoints")
	}
	seen := make(map[string]bool, len(pair.Checkpoints))
	for _, checkpoint := range pair.Checkpoints {
		if !slices.Contains(wantNames, checkpoint.Name) || seen[checkpoint.Name] || !checkpoint.Pass || !checkpoint.StateMatched || !finitePublicPairMetrics(checkpoint.Metrics) || checkpoint.Metrics.MaxComplexDifference < 0 || checkpoint.Metrics.MaxComplexDifference > publicNumericalGate {
			return "", fmt.Errorf("preflight pair checkpoint %q is missing, duplicate, or failed a fixed gate", checkpoint.Name)
		}
		seen[checkpoint.Name] = true
	}
	for _, name := range wantNames {
		if !seen[name] {
			return "", fmt.Errorf("preflight pair is missing checkpoint %q", name)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func finitePublicPairMetrics(metrics vectorMetrics) bool {
	for _, value := range []float64{
		metrics.ComplexRMSE, metrics.MaxComplexDifference, metrics.MaxRealDifference,
		metrics.MaxImagDifference, metrics.MedianPrecisionBits,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return metrics.ComplexRMSE >= 0 && metrics.MaxRealDifference >= 0 && metrics.MaxImagDifference >= 0
}

func newPublicBackend(params bootstrapping.Parameters, residual ckks.Parameters) (*publicBackend, error) {
	if err := validatePublicEphemeralWeight(params.EphemeralSecretWeight); err != nil {
		return nil, err
	}
	full := params.BootstrappingParameters
	b := &publicBackend{name: publicBackendName(), residual: residual, full: full, params: params}
	keygen := rlwe.NewKeyGenerator(residual)
	phase, err := measurePhase("secret_key_generation", func() error {
		b.secret = keygen.GenSecretKeyNew()
		return nil
	})
	b.construction = append(b.construction, phase)
	if err != nil {
		return nil, fmt.Errorf("generate native secret key: %w", err)
	}
	var keys *bootstrapping.EvaluationKeys
	phase, err = measurePhase("public_GenEvaluationKeys", func() error {
		var err error
		keys, b.bootstrapSecret, err = params.GenEvaluationKeys(b.secret)
		return err
	})
	b.construction = append(b.construction, phase)
	if err != nil {
		return nil, fmt.Errorf("public GenEvaluationKeys: %w", err)
	}
	if keys == nil || keys.MemEvaluationKeySet == nil || b.bootstrapSecret == nil {
		return nil, errors.New("public GenEvaluationKeys returned incomplete evaluation-key/secret material")
	}
	if b.bootstrapSecret.LevelQ() != full.MaxLevel() {
		return nil, fmt.Errorf("public GenEvaluationKeys returned secret LevelQ=%d, want full Q level %d", b.bootstrapSecret.LevelQ(), full.MaxLevel())
	}
	if !secretPrefixEqual(b.secret, b.bootstrapSecret, residual.MaxLevel()+1) {
		return nil, errors.New("public GenEvaluationKeys extended secret does not preserve residual Q-prefix")
	}
	phase, err = measurePhase("public_ckks_evaluator_construction", func() error {
		b.evaluator = ckks.NewEvaluator(full, keys.MemEvaluationKeySet)
		return nil
	})
	b.construction = append(b.construction, phase)
	if err != nil {
		return nil, err
	}
	var bootstrapEval *bootstrapping.Evaluator
	phase, err = measurePhase("public_bootstrapping_evaluator_construction", func() error {
		var constructErr error
		bootstrapEval, constructErr = bootstrapping.NewEvaluator(params, keys)
		return constructErr
	})
	b.construction = append(b.construction, phase)
	if err != nil {
		return nil, err
	}
	b.dispatch, err = publicEvaluatorDispatch(bootstrapEval)
	if err != nil {
		return nil, err
	}
	b.bootstrapEval = bootstrapEval
	b.maxSlots = residual.MaxSlots()
	return b, nil
}

func (b *publicBackend) encryptAtLevel(values []complex128, scale rlwe.Scale, level, logSlots int) (*rlwe.Ciphertext, phaseTiming, error) {
	var ct *rlwe.Ciphertext
	phase, err := measurePhase("public_EncryptNew", func() error {
		pt := ckks.NewPlaintext(b.full, level)
		pt.Scale, pt.IsNTT, pt.LogDimensions = scale, true, ring.Dimensions{Cols: logSlots}
		if encodeErr := ckks.NewEncoder(b.full).Encode(values, pt); encodeErr != nil {
			return encodeErr
		}
		var encryptErr error
		ct, encryptErr = rlwe.NewEncryptor(b.full, b.bootstrapSecret).EncryptNew(pt)
		return encryptErr
	})
	return ct, phase, err
}

func (b *publicBackend) decode(ct *rlwe.Ciphertext) ([]complex128, string, error) {
	if ct == nil || ct.MetaData == nil || ct.Level() < 0 || ct.Degree() < 1 {
		return nil, "", errors.New("native decode requires a non-nil ciphertext with metadata")
	}
	params, secret := b.full, b.bootstrapSecret
	if ct.Level() <= b.residual.MaxLevel() {
		params, secret = b.residual, b.secret
	}
	if b.name == "fast" {
		c1Nonzero, err := isCiphertextComponentNonzero(ct, 1)
		if err != nil {
			return nil, "", err
		}
		if c1Nonzero {
			return nil, "", errors.New("Fast zero-secret checkpoint has a nonzero c1 component")
		}
	}
	if b.name == "fast" && ct.Level() > 0 {
		if len(ct.Value) != 2 || !ct.IsNTT || ct.IsMontgomery {
			return nil, "", errors.New("Fast Q-prefix checkpoint requires degree-one NTT non-Montgomery ciphertext storage")
		}
		rows := min(ct.Level()+1, 4)
		prefixLevel := rows - 1
		plain := ckks.NewPlaintext(b.full, prefixLevel)
		*plain.MetaData = *ct.MetaData
		if len(ct.Value[0].Coeffs) < rows || len(plain.Value.Coeffs) < rows {
			return nil, "", errors.New("Fast Q-prefix decode has incomplete authoritative c0 rows")
		}
		for row := 0; row < rows; row++ {
			if len(ct.Value[0].Coeffs[row]) != ct.N() || len(plain.Value.Coeffs[row]) != ct.N() {
				return nil, "", fmt.Errorf("Fast Q-prefix decode q%d row length does not match N=%d", row, ct.N())
			}
			copy(plain.Value.Coeffs[row], ct.Value[0].Coeffs[row])
		}
		decoded := make([]complex128, b.full.MaxSlots())
		if err := ckks.NewEncoder(b.full).Decode(plain, decoded); err != nil {
			return nil, "", err
		}
		if len(decoded) < b.maxSlots {
			return nil, "", errors.New("Fast Q-prefix Decode returned fewer slots than the frozen fixture")
		}
		return decoded[:b.maxSlots], "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)", nil
	}
	plain := rlwe.NewDecryptor(params, secret).DecryptNew(ct)
	decoded := make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(plain, decoded); err != nil {
		return nil, "", err
	}
	if len(decoded) < b.maxSlots {
		return nil, "", errors.New("native Decode returned fewer slots than the frozen fixture")
	}
	return decoded[:b.maxSlots], "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode", nil
}

func (b *publicBackend) decodeNativeLevel1(ct *rlwe.Ciphertext) ([]complex128, string, error) {
	if ct == nil || ct.MetaData == nil || ct.Level() != 1 || ct.Degree() != 1 || len(ct.Value) != 2 {
		return nil, "", errors.New("native Bootstrap output decode requires a degree-one Level1 ciphertext with metadata")
	}
	if b.residual.MaxLevel() != 1 || b.secret == nil {
		return nil, "", errors.New("native Level1 decode requires the frozen residual secret/profile")
	}
	if b.name == "fast" {
		c1Nonzero, err := isCiphertextComponentNonzero(ct, 1)
		if err != nil {
			return nil, "", err
		}
		if c1Nonzero {
			return nil, "", errors.New("Fast public Bootstrap output violates zero-secret c1=0 mode")
		}
	}
	for component, polynomial := range ct.Value {
		if len(polynomial.Coeffs) < 2 {
			return nil, "", fmt.Errorf("native Level1 output component %d lacks q0/q1 rows", component)
		}
		for row := 0; row < 2; row++ {
			if len(polynomial.Coeffs[row]) != ct.N() {
				return nil, "", fmt.Errorf("native Level1 output component %d q%d row has invalid length", component, row)
			}
		}
		if b.name == "fast" {
			for row := 2; row < len(polynomial.Coeffs); row++ {
				if len(polynomial.Coeffs[row]) != 0 {
					return nil, "", fmt.Errorf("Fast Level1 output component %d retains non-authoritative q%d row", component, row)
				}
			}
		}
	}
	plain := rlwe.NewDecryptor(b.residual, b.secret).DecryptNew(ct)
	decoded := make([]complex128, b.residual.MaxSlots())
	if err := ckks.NewEncoder(b.residual).Decode(plain, decoded); err != nil {
		return nil, "", err
	}
	if len(decoded) < b.maxSlots {
		return nil, "", errors.New("native Level1 Decode returned fewer slots than the frozen fixture")
	}
	path := "rlwe.NewDecryptor(residual parameters, generated secret).DecryptNew -> ckks.NewEncoder.Decode"
	if b.name == "fast" {
		path = "rlwe.NewDecryptor(residual parameters, matching zero-secret-mode secret).DecryptNew -> ckks.NewEncoder.Decode"
	}
	return decoded[:b.maxSlots], path, nil
}

func makePublicCheckpoint(backend *publicBackend, name string, ct *rlwe.Ciphertext, decoded, expected []complex128, q []uint64, fastCompactOutput bool) (publicCheckpoint, error) {
	if ct == nil || ct.MetaData == nil || ct.Degree() != 1 || len(ct.Value) != 2 || ct.Level() < 0 || ct.Level() >= len(q) {
		return publicCheckpoint{}, fmt.Errorf("%s has invalid ciphertext shape/metadata", name)
	}
	rows := min(ct.Level()+1, 4)
	if rows < 1 {
		return publicCheckpoint{}, fmt.Errorf("%s has invalid Q-prefix width %d", name, rows)
	}
	cp := publicCheckpoint{Name: name, State: stateOf(ct, q, rows), PhysicalRowLengths: make([][]int, len(ct.Value)), RowSHA256: make([][]string, len(ct.Value))}
	for component := range ct.Value {
		coeffRows := ct.Value[component].Coeffs
		if len(coeffRows) < rows {
			return publicCheckpoint{}, fmt.Errorf("%s component %d has %d rows, needs %d", name, component, len(coeffRows), rows)
		}
		cp.PhysicalRowLengths[component] = make([]int, len(coeffRows))
		cp.RowSHA256[component] = make([]string, rows)
		for row, coeffs := range coeffRows {
			cp.PhysicalRowLengths[component][row] = len(coeffs)
		}
		for row := 0; row < rows; row++ {
			if len(coeffRows[row]) != ct.N() {
				return publicCheckpoint{}, fmt.Errorf("%s component %d q%d row length=%d, want N=%d", name, component, row, len(coeffRows[row]), ct.N())
			}
			cp.RowSHA256[component][row] = hashUint64Row(coeffRows[row])
		}
	}
	if backend.name == "fast" {
		cp.FastCompactPrefix = fastCompactOutput
		if fastCompactOutput {
			for component, polynomial := range ct.Value {
				coeffRows := polynomial.Coeffs
				for row := rows; row < len(coeffRows); row++ {
					if len(coeffRows[row]) != 0 {
						return publicCheckpoint{}, fmt.Errorf("%s Fast compact output retains dormant q%d row in component %d (%d coefficients)", name, row, component, len(coeffRows[row]))
					}
				}
			}
		}
	}
	var err error
	cp.C1Nonzero, err = isCiphertextComponentNonzero(ct, 1)
	if err != nil {
		return publicCheckpoint{}, err
	}
	if backend.name == "fast" && cp.C1Nonzero {
		return publicCheckpoint{}, fmt.Errorf("%s Fast zero-secret state has nonzero c1", name)
	}
	if len(decoded) == 0 || len(decoded) != len(expected) || !allFinite(decoded) || !allFinite(expected) {
		return publicCheckpoint{}, fmt.Errorf("%s has empty, mismatched, or non-finite decoded/oracle vectors", name)
	}
	cp.Oracle = compareVectors(expected, decoded, publicNumericalGate)
	cp.OracleSNR = numericalmetrics.Compare(expected, decoded)
	cp.OraclePass = cp.Oracle.MaxComplexDifference <= publicNumericalGate
	cp.DecodedSHA256 = perfmeasure.Fingerprint(decoded)
	return cp, nil
}

func runPublicCompare(args []string) error {
	flags := flag.NewFlagSet("perfprobe compare-public", flag.ContinueOnError)
	standardPath := flags.String("standard", "", "Standard public-native preflight JSON")
	fastPath := flags.String("fast", "", "Fast public-native preflight JSON")
	standardVectorsPath := flags.String("standard-vectors", "", "Standard ephemeral vectors JSON")
	fastVectorsPath := flags.String("fast-vectors", "", "Fast ephemeral vectors JSON")
	outPath := flags.String("out", "", "compact paired preflight JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *standardPath == "" || *fastPath == "" || *standardVectorsPath == "" || *fastVectorsPath == "" || *outPath == "" {
		return errors.New("--standard, --fast, --standard-vectors, --fast-vectors, and --out are required")
	}
	var standard, fast publicNativeDocument
	var standardVectors, fastVectors publicVectorDocument
	if err := readJSON(*standardPath, &standard); err != nil {
		return err
	}
	if err := readJSON(*fastPath, &fast); err != nil {
		return err
	}
	if err := readJSON(*standardVectorsPath, &standardVectors); err != nil {
		return err
	}
	if err := readJSON(*fastVectorsPath, &fastVectors); err != nil {
		return err
	}
	bootstrapMode := standard.Mode == "public-native-bootstrap" || standard.Mode == "public-native-repeatability" ||
		fast.Mode == "public-native-bootstrap" || fast.Mode == "public-native-repeatability"
	if bootstrapMode {
		if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
			return err
		}
	} else if err := validatePublicPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		return err
	}
	standardStates, fastStates := map[string]publicCheckpoint{}, map[string]publicCheckpoint{}
	for _, cp := range standard.Checkpoints {
		standardStates[cp.Name] = cp
	}
	for _, cp := range fast.Checkpoints {
		fastStates[cp.Name] = cp
	}
	if len(standardStates) == 0 || len(standard.Checkpoints) != len(standardStates) || len(fast.Checkpoints) != len(fastStates) ||
		len(standardStates) != len(fastStates) || len(standardVectors.Checkpoints) != len(standardStates) || len(fastVectors.Checkpoints) != len(fastStates) {
		return errors.New("paired artifacts have empty or mismatched checkpoint coverage")
	}
	pair := publicPairDocument{
		SchemaVersion: "fast-standard-public-native-pair.v1", GeneratedAt: time.Now().UTC(), Status: "PASS",
		NumericalGate: publicNumericalGate, FastCommit: fast.BackendCommit, StandardCommit: standard.BackendCommit,
		PrimaryCommit: standard.PrimaryCommit, SourceSHA256: standard.PrimarySourceSHA256, InputSHA256: standard.InputSHA256,
		WorkloadSHA256: standard.WorkloadSHA256, ConfigSHA256: standard.ConfigSHA256, QPSHA256: standard.QPSHA256, E: 32,
		BootstrapCalls: 0, MatchedEnvironment: publicEnvironmentsMatch(standard, fast),
	}
	if bootstrapMode {
		pair.BootstrapCalls = standard.BootstrapCalls + fast.BootstrapCalls
		if standard.Mode == "public-native-repeatability" {
			pair.Limitations = []string{
				"One cold plus five warm public Bootstrap samples characterize only this pinned LogN13/E32 profile and environment.",
				"Standard uses genuine native encryption while Fast uses the pinned intentionally insecure zero-secret mode; security equivalence is not claimed.",
			}
		} else {
			pair.Limitations = []string{
				"Exactly one cold and one warm Bootstrap sample per backend are descriptive, not a statistically stable speedup estimate.",
				"Standard uses genuine native encryption while Fast uses the pinned intentionally insecure zero-secret mode; security equivalence is not claimed.",
			}
		}
	} else {
		pair.Limitations = []string{"Only the public pre-Bootstrap chain is compared; no public Bootstrap acceptance or Bootstrap output metric is claimed."}
	}
	names := make([]string, 0, len(standardStates))
	for name := range standardStates {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		standardState, standardOK := standardStates[name]
		fastState, fastOK := fastStates[name]
		standardValues, standardValuesOK := standardVectors.Checkpoints[name]
		fastValues, fastValuesOK := fastVectors.Checkpoints[name]
		if !standardOK || !fastOK || !standardValuesOK || !fastValuesOK || len(standardValues) == 0 || len(fastValues) == 0 {
			return fmt.Errorf("checkpoint %s is absent or has an empty decoded vector", name)
		}
		standardComplex, fastComplex := decodeComplexValues(standardValues), decodeComplexValues(fastValues)
		if !allFinite(standardComplex) || !allFinite(fastComplex) || len(standardComplex) != len(fastComplex) {
			return fmt.Errorf("checkpoint %s has non-finite or mismatched paired vectors", name)
		}
		stateMatched := standardState.State.Level == fastState.State.Level && standardState.State.Degree == fastState.State.Degree && standardState.State.Scale == fastState.State.Scale
		metrics := compareVectors(standardComplex, fastComplex, publicNumericalGate)
		passed := stateMatched && metrics.MaxComplexDifference <= publicNumericalGate
		if !passed {
			pair.Status = "FAIL"
		}
		pair.Checkpoints = append(pair.Checkpoints, publicPairCheckpoint{Name: name, StateMatched: stateMatched, FastState: fastState.State, StandardState: standardState.State, Metrics: metrics, Pass: passed})
	}
	if bootstrapMode {
		for index, standardResult := range standard.BootstrapResults {
			fastResult := fast.BootstrapResults[index]
			phase := standardResult.Phase
			name := standardResult.Output.Name
			standardValues := decodeComplexValues(standardVectors.BootstrapOutputs[name])
			fastValues := decodeComplexValues(fastVectors.BootstrapOutputs[name])
			stateMatched := standardResult.Output.State.Level == fastResult.Output.State.Level &&
				standardResult.Output.State.Degree == fastResult.Output.State.Degree &&
				standardResult.Output.State.Scale == fastResult.Output.State.Scale
			metrics := compareVectors(standardValues, fastValues, publicNumericalGate)
			passed := stateMatched && metrics.MaxComplexDifference <= publicNumericalGate
			if !passed {
				pair.Status = "FAIL"
			}
			pair.Bootstrap = append(pair.Bootstrap, publicPairBootstrapComparison{
				Phase: phase, StateMatched: stateMatched, FastState: fastResult.Output.State,
				StandardState: standardResult.Output.State, Metrics: metrics, Pass: passed,
			})
		}
	}
	if !pair.MatchedEnvironment {
		pair.Status = "FAIL"
	}
	if err := writeExclusiveJSON(*outPath, pair); err != nil {
		return err
	}
	if pair.Status != "PASS" {
		return errors.New("one or more paired public checkpoints exceeded the fixed 1e-6 gate or metadata did not match")
	}
	return nil
}

func validatePublicPairArtifacts(standard, fast publicNativeDocument, standardVectors, fastVectors publicVectorDocument) error {
	if standard.SchemaVersion != "fast-standard-public-native-preflight.v1" || fast.SchemaVersion != standard.SchemaVersion || standard.Status != "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN" || fast.Status != standard.Status {
		return errors.New("paired inputs are not successful public-native preflight artifacts")
	}
	if standard.Backend != "standard" || fast.Backend != "fast" || standard.Profile != "logn13" || fast.Profile != standard.Profile ||
		standard.EphemeralSecretWeight != 32 || fast.EphemeralSecretWeight != 32 || standard.InputSlots != 1<<12 || fast.InputSlots != standard.InputSlots ||
		standard.BootstrapCalls != 0 || fast.BootstrapCalls != 0 || standard.BootstrapBudget != 0 || fast.BootstrapBudget != 0 ||
		!standard.PrimaryClean || !fast.PrimaryClean || !standard.BackendClean || !fast.BackendClean {
		return errors.New("paired artifacts violate pinned LogN13/E32 clean-provenance zero-Bootstrap contract")
	}
	if err := validatePublicBackendPin(standard.Backend, standard.BackendCommit); err != nil {
		return err
	}
	if err := validatePublicBackendPin(fast.Backend, fast.BackendCommit); err != nil {
		return err
	}
	if standard.BuildTag != "perf_standard" || fast.BuildTag != "perf_fast" ||
		standard.PrimaryCommit == "" || standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimarySourceSHA256 == "" || standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 ||
		standard.ConfigSHA256 != publicConfigSHA || fast.ConfigSHA256 != publicConfigSHA || standard.QPSHA256 != publicQPSHA || fast.QPSHA256 != publicQPSHA ||
		standard.InputSHA256 != publicInputSHA || fast.InputSHA256 != publicInputSHA || standard.WorkloadSHA256 != publicWorkloadSHA || fast.WorkloadSHA256 != publicWorkloadSHA ||
		!reflect.DeepEqual(standard.Parameters, fast.Parameters) {
		return errors.New("paired artifacts have mismatched or noncanonical Primary source, pins, config, parameters, input, or workload provenance")
	}
	if standardVectors.SchemaVersion != "fast-standard-public-native-vectors.v1" || fastVectors.SchemaVersion != standardVectors.SchemaVersion ||
		standardVectors.Mode != "public-native" || fastVectors.Mode != "public-native" ||
		standardVectors.SourceSHA256 != standard.PrimarySourceSHA256 || fastVectors.SourceSHA256 != fast.PrimarySourceSHA256 ||
		standardVectors.PrimaryCommit != standard.PrimaryCommit || fastVectors.PrimaryCommit != fast.PrimaryCommit ||
		standardVectors.BackendCommit != standard.BackendCommit || fastVectors.BackendCommit != fast.BackendCommit ||
		standardVectors.ConfigSHA256 != standard.ConfigSHA256 || fastVectors.ConfigSHA256 != fast.ConfigSHA256 ||
		standardVectors.QPSHA256 != standard.QPSHA256 || fastVectors.QPSHA256 != fast.QPSHA256 ||
		standardVectors.InputSHA256 != standard.InputSHA256 || fastVectors.InputSHA256 != fast.InputSHA256 ||
		standardVectors.WorkloadSHA256 != standard.WorkloadSHA256 || fastVectors.WorkloadSHA256 != fast.WorkloadSHA256 ||
		standardVectors.E != 32 || fastVectors.E != 32 || standardVectors.Backend != "standard" || fastVectors.Backend != "fast" {
		return errors.New("paired vector artifacts have mismatched source/input/workload provenance")
	}
	wantNames := []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"}
	if !samePublicCheckpointNames(standard.Checkpoints, wantNames) || !samePublicCheckpointNames(fast.Checkpoints, wantNames) {
		return errors.New("paired artifacts do not contain the exact canonical pre-Bootstrap checkpoint set")
	}
	if !reflect.DeepEqual(standard.Capacity, fast.Capacity) {
		return errors.New("paired artifacts have mismatched Q0123/q0 capacity evidence")
	}
	if err := validatePublicCapacity(standard.Capacity); err != nil {
		return fmt.Errorf("Standard capacity evidence: %w", err)
	}
	if err := validatePublicCapacity(fast.Capacity); err != nil {
		return fmt.Errorf("Fast capacity evidence: %w", err)
	}
	oracles, err := publicOracleCheckpoints(1 << 12)
	if err != nil {
		return err
	}
	if err := validatePublicBackendEvidence(standard, standardVectors, oracles); err != nil {
		return fmt.Errorf("Standard public evidence: %w", err)
	}
	if err := validatePublicBackendEvidence(fast, fastVectors, oracles); err != nil {
		return fmt.Errorf("Fast public evidence: %w", err)
	}
	for _, checkpoint := range standard.Checkpoints {
		if checkpoint.FastCompactPrefix || checkpoint.DecodePath != "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode" || !checkpoint.C1Nonzero {
			return errors.New("Standard artifact has an unexpected compact-storage or native-decrypt declaration")
		}
	}
	if !standard.InputC1Nonzero || fast.InputC1Nonzero {
		return errors.New("paired inputs do not preserve native Standard and Fast zero-secret c1 provenance")
	}
	for _, checkpoint := range fast.Checkpoints {
		wantCompact := checkpoint.Name != "encrypt_a_level5" && checkpoint.Name != "encrypt_b_level5" && checkpoint.Name != "encrypt_c_q5_level5"
		if checkpoint.FastCompactPrefix != wantCompact || checkpoint.C1Nonzero {
			return fmt.Errorf("Fast checkpoint %s has unexpected compact-prefix declaration", checkpoint.Name)
		}
		wantDecode := "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)"
		if checkpoint.Name == "drop_level0" {
			wantDecode = "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode"
		}
		if checkpoint.DecodePath != wantDecode {
			return fmt.Errorf("Fast checkpoint %s has unexpected decode path", checkpoint.Name)
		}
	}
	return nil
}

func publicEnvironmentsMatch(a, b publicNativeDocument) bool {
	return a.GoVersion == b.GoVersion && a.OS == b.OS && a.Arch == b.Arch && a.CPU == b.CPU &&
		a.NumCPU == b.NumCPU && a.GOMAXPROCS == b.GOMAXPROCS && a.GOGC == b.GOGC &&
		a.GOMEMLIMIT == b.GOMEMLIMIT && a.GODEBUG == b.GODEBUG
}

func validatePublicBootstrapPairArtifacts(standard, fast publicNativeDocument, standardVectors, fastVectors publicVectorDocument) error {
	if standard.Mode != fast.Mode || (standard.Mode != "public-native-bootstrap" && standard.Mode != "public-native-repeatability") {
		return errors.New("paired Bootstrap artifacts do not share an approved public Bootstrap mode")
	}
	wantBudget, wantSchema, wantStatus, wantVectorSchema := 2,
		"fast-standard-public-native-bootstrap.v1", "PASS_PUBLIC_BOOTSTRAP_COLD_WARM", "fast-standard-public-native-bootstrap-vectors.v1"
	if standard.Mode == "public-native-repeatability" {
		wantBudget, wantSchema, wantStatus, wantVectorSchema = 6,
			"fast-standard-public-native-repeatability.v1", "PASS_PUBLIC_BOOTSTRAP_1COLD_5WARM", "fast-standard-public-native-repeatability-vectors.v1"
	}
	if standard.SchemaVersion != wantSchema || fast.SchemaVersion != wantSchema ||
		standard.Status != wantStatus || fast.Status != wantStatus ||
		standard.BootstrapBudget != wantBudget || fast.BootstrapBudget != wantBudget ||
		standard.BootstrapCalls != wantBudget || fast.BootstrapCalls != wantBudget ||
		standard.NumericalGate != publicNumericalGate || fast.NumericalGate != publicNumericalGate ||
		standard.PreflightPairSHA256 == "" || standard.PreflightPairSHA256 != fast.PreflightPairSHA256 {
		return errors.New("paired Bootstrap artifacts violate their exact bounded, preflight-bound contract")
	}
	if standardVectors.SchemaVersion != wantVectorSchema || fastVectors.SchemaVersion != wantVectorSchema ||
		standardVectors.Mode != standard.Mode || fastVectors.Mode != standardVectors.Mode {
		return errors.New("paired Bootstrap vector artifacts have the wrong schema or mode")
	}
	standardBase, fastBase := standard, fast
	standardVectorBase, fastVectorBase := standardVectors, fastVectors
	for _, base := range []*publicNativeDocument{&standardBase, &fastBase} {
		base.SchemaVersion = "fast-standard-public-native-preflight.v1"
		base.Mode = "public-native"
		base.Status = "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN"
		base.BootstrapBudget, base.BootstrapCalls = 0, 0
		base.BootstrapResults = nil
		base.PreflightPairSHA256 = ""
	}
	for _, base := range []*publicVectorDocument{&standardVectorBase, &fastVectorBase} {
		base.SchemaVersion = "fast-standard-public-native-vectors.v1"
		base.Mode = "public-native"
		base.BootstrapOutputs = nil
	}
	if err := validatePublicPairArtifacts(standardBase, fastBase, standardVectorBase, fastVectorBase); err != nil {
		return fmt.Errorf("pre-Bootstrap evidence: %w", err)
	}
	if err := validatePublicBootstrapOutputEvidence(standard, standardVectors, standardBase, standardVectorBase); err != nil {
		return fmt.Errorf("Standard Bootstrap evidence: %w", err)
	}
	if err := validatePublicBootstrapOutputEvidence(fast, fastVectors, fastBase, fastVectorBase); err != nil {
		return fmt.Errorf("Fast Bootstrap evidence: %w", err)
	}
	return nil
}

func validatePublicBootstrapOutputEvidence(doc publicNativeDocument, vectors publicVectorDocument, base publicNativeDocument, baseVectors publicVectorDocument) error {
	wantPhases, err := publicBootstrapAttemptSpecs(doc.BootstrapBudget)
	if err != nil {
		return err
	}
	if len(doc.BootstrapResults) != len(wantPhases) || len(vectors.BootstrapOutputs) != len(wantPhases) {
		return errors.New("Bootstrap evidence must contain exactly cold and warm outputs")
	}
	baseInput, ok := baseVectors.Checkpoints["drop_level0"]
	if !ok || len(baseInput) != 1<<12 {
		return errors.New("held pre-Bootstrap Level0 decoded vector is missing or has an invalid length")
	}
	input := decodeComplexValues(baseInput)
	if !allFinite(input) {
		return errors.New("held pre-Bootstrap Level0 vector contains NaN or Inf")
	}
	inputFingerprint := ""
	for index, want := range wantPhases {
		result := doc.BootstrapResults[index]
		if result.Phase != want.phase || result.Timing.Phase != want.phase || !result.Timing.Available || result.Timing.Samples != 1 || result.InputCiphertextSHA256 == "" {
			return fmt.Errorf("Bootstrap result %d is missing its named measured phase/input fingerprint", index)
		}
		if inputFingerprint == "" {
			inputFingerprint = result.InputCiphertextSHA256
		} else if result.InputCiphertextSHA256 != inputFingerprint {
			return errors.New("cold and warm Bootstrap did not use copies of the same held ciphertext")
		}
		checkpoint := result.Output
		if checkpoint.Name != want.name {
			return fmt.Errorf("Bootstrap phase %s has output checkpoint %q, want %q", want.phase, checkpoint.Name, want.name)
		}
		values, exists := vectors.BootstrapOutputs[checkpoint.Name]
		if !exists || len(values) != 1<<12 {
			return fmt.Errorf("Bootstrap output %s has %d decoded slots, want exactly %d", checkpoint.Name, len(values), 1<<12)
		}
		decoded := decodeComplexValues(values)
		if !allFinite(decoded) || checkpoint.DecodedSHA256 == "" || perfmeasure.Fingerprint(decoded) != checkpoint.DecodedSHA256 {
			return fmt.Errorf("Bootstrap output %s has non-finite data or a mismatched DecodedSHA256", checkpoint.Name)
		}
		metrics := compareVectors(input, decoded, publicNumericalGate)
		snr := numericalmetrics.Compare(input, decoded)
		passed := metrics.MaxComplexDifference <= publicNumericalGate
		if !passed || checkpoint.OraclePass != passed || !reflect.DeepEqual(checkpoint.Oracle, metrics) || !reflect.DeepEqual(checkpoint.OracleSNR, snr) {
			return fmt.Errorf("Bootstrap output %s fails or misreports its native plaintext oracle", checkpoint.Name)
		}
		if doc.Backend == "standard" {
			if checkpoint.DecodePath != "rlwe.NewDecryptor(residual parameters, generated secret).DecryptNew -> ckks.NewEncoder.Decode" || !checkpoint.C1Nonzero {
				return fmt.Errorf("Standard Bootstrap output %s lacks genuine native Level1 decryption provenance", checkpoint.Name)
			}
		} else if checkpoint.DecodePath != "rlwe.NewDecryptor(residual parameters, matching zero-secret-mode secret).DecryptNew -> ckks.NewEncoder.Decode" || checkpoint.C1Nonzero {
			return fmt.Errorf("Fast Bootstrap output %s lacks native Level1 zero-secret decryption provenance", checkpoint.Name)
		}
		if err := validatePublicCheckpointState(doc, checkpoint); err != nil {
			return err
		}
	}
	for name := range vectors.BootstrapOutputs {
		if !isPublicBootstrapCheckpoint(name, doc.BootstrapBudget) {
			return fmt.Errorf("unexpected Bootstrap vector checkpoint %q", name)
		}
	}
	return nil
}

func publicBootstrapResultByPhase(results []publicBootstrapResult, phase string) publicBootstrapResult {
	for _, result := range results {
		if result.Phase == phase {
			return result
		}
	}
	return publicBootstrapResult{}
}

func validatePublicBackendEvidence(doc publicNativeDocument, vectors publicVectorDocument, oracles map[string][]complex128) error {
	wantNames := []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"}
	if len(vectors.Checkpoints) != len(wantNames) {
		return fmt.Errorf("decoded vector checkpoint count=%d, want %d", len(vectors.Checkpoints), len(wantNames))
	}
	for name := range vectors.Checkpoints {
		if !slices.Contains(wantNames, name) {
			return fmt.Errorf("unexpected decoded vector checkpoint %q", name)
		}
	}
	if doc.InputSlots != 1<<12 || doc.Parameters.InputSlots != doc.InputSlots || doc.Parameters.RingN != 1<<13 {
		return errors.New("public evidence does not describe the frozen 4096-slot LogN13 profile")
	}
	if err := validatePublicProfileEvidence(doc); err != nil {
		return err
	}
	if err := validatePublicCapacity(doc.Capacity); err != nil {
		return err
	}
	if doc.Backend == "fast" && doc.InputC1Nonzero {
		return errors.New("Fast input reports nonzero c1 for the approved zero-secret mode")
	}
	for _, checkpoint := range doc.Checkpoints {
		values, ok := vectors.Checkpoints[checkpoint.Name]
		if !ok || len(values) != 1<<12 {
			return fmt.Errorf("checkpoint %s decoded vector has %d slots, want exactly %d", checkpoint.Name, len(values), 1<<12)
		}
		decoded := decodeComplexValues(values)
		if !allFinite(decoded) {
			return fmt.Errorf("checkpoint %s decoded vector contains NaN or Inf", checkpoint.Name)
		}
		if checkpoint.DecodedSHA256 == "" || perfmeasure.Fingerprint(decoded) != checkpoint.DecodedSHA256 {
			return fmt.Errorf("checkpoint %s decoded vector does not match DecodedSHA256", checkpoint.Name)
		}
		oracle, ok := oracles[checkpoint.Name]
		if !ok || len(oracle) != len(decoded) || !allFinite(oracle) {
			return fmt.Errorf("checkpoint %s has no canonical finite plaintext oracle", checkpoint.Name)
		}
		recomputed := compareVectors(oracle, decoded, publicNumericalGate)
		recomputedSNR := numericalmetrics.Compare(oracle, decoded)
		passed := recomputed.MaxComplexDifference <= publicNumericalGate
		if !passed || checkpoint.OraclePass != passed || !reflect.DeepEqual(checkpoint.Oracle, recomputed) || !reflect.DeepEqual(checkpoint.OracleSNR, recomputedSNR) {
			return fmt.Errorf("checkpoint %s stored plaintext-oracle metrics do not match the decoded vector", checkpoint.Name)
		}
		if err := validatePublicCheckpointState(doc, checkpoint); err != nil {
			return err
		}
	}
	return nil
}

func validatePublicProfileEvidence(doc publicNativeDocument) error {
	parsePrimes := func(label string, values []string) ([]uint64, error) {
		primes := make([]uint64, len(values))
		for index, value := range values {
			prime, err := strconv.ParseUint(value, 10, 64)
			if err != nil || prime == 0 {
				return nil, fmt.Errorf("invalid %s[%d] modulus metadata", label, index)
			}
			primes[index] = prime
		}
		return primes, nil
	}
	q, err := parsePrimes("Q", doc.Parameters.QPrimes)
	if err != nil {
		return err
	}
	p, err := parsePrimes("P", doc.Parameters.PPrimes)
	if err != nil {
		return err
	}
	if len(q) != len(doc.Parameters.QChainBits) || len(p) != len(doc.Parameters.PBits) || len(q) <= publicMulRescaleInputLvl {
		return errors.New("public profile modulus counts disagree with effective Q/P metadata")
	}
	qpBytes, err := json.Marshal(struct{ Q, P []uint64 }{q, p})
	if err != nil {
		return err
	}
	qpHash := sha256.Sum256(qpBytes)
	if hex.EncodeToString(qpHash[:]) != doc.QPSHA256 || doc.QPSHA256 != publicQPSHA {
		return errors.New("recorded Q/P modulus rows do not match the canonical Q/P fingerprint")
	}
	a, b, c, err := perfmeasure.PublicMulRescaleWorkload(doc.InputSlots)
	if err != nil {
		return err
	}
	if perfmeasure.Fingerprint(a) != doc.InputSHA256 || doc.InputSHA256 != publicInputSHA {
		return errors.New("recorded public input does not match the canonical workload fingerprint")
	}
	workloadHash, err := perfmeasure.PublicMulRescaleWorkloadFingerprint(a, b, c, new(big.Int).SetUint64(q[publicMulRescaleInputLvl]).String())
	if err != nil {
		return err
	}
	if workloadHash != doc.WorkloadSHA256 || doc.WorkloadSHA256 != publicWorkloadSHA {
		return errors.New("recorded public workload does not match the canonical workload fingerprint")
	}
	capacity, err := perfmeasure.BoundPublicMulRescaleWorkload(a, b, c, q, doc.Parameters.RingN)
	if err != nil {
		return fmt.Errorf("recompute canonical Q0123/q0 capacity: %w", err)
	}
	if !reflect.DeepEqual(capacity, doc.Capacity) {
		return errors.New("recorded Q0123/q0 capacity plan does not match recomputed canonical bounds")
	}
	return nil
}

func validatePublicCheckpointState(doc publicNativeDocument, checkpoint publicCheckpoint) error {
	state := checkpoint.State
	if state.Degree != 1 || state.Level < 0 || state.Scale == "" {
		return fmt.Errorf("checkpoint %s has invalid Level/Scale/Degree metadata", checkpoint.Name)
	}
	wantLevels := map[string]int{
		"encrypt_a_level5": 5, "encrypt_b_level5": 5, "encrypt_c_q5_level5": 5,
		"add_level5": 5, "mulrelin_level5": 5, "rescale_q5_level4": 4,
		"rotate_level4": 4, "drop_level0": 0,
	}
	wantLevel, knownCheckpoint := wantLevels[checkpoint.Name]
	if isPublicBootstrapCheckpoint(checkpoint.Name, doc.BootstrapBudget) {
		wantLevel, knownCheckpoint = 1, true
	} else if !knownCheckpoint {
		return fmt.Errorf("unknown public checkpoint %q for Level validation", checkpoint.Name)
	}
	if state.Level != wantLevel || state.QPrefixRows != min(state.Level+1, 4) {
		return fmt.Errorf("checkpoint %s has unexpected Level/Q-prefix row metadata", checkpoint.Name)
	}
	wantScale, err := expectedPublicCheckpointScale(doc.Parameters, checkpoint.Name)
	if err != nil {
		return err
	}
	if state.Scale != wantScale {
		return fmt.Errorf("checkpoint %s Scale=%s, want frozen workload Scale %s", checkpoint.Name, state.Scale, wantScale)
	}
	if len(checkpoint.PhysicalRowLengths) != 2 || len(checkpoint.RowSHA256) != 2 {
		return fmt.Errorf("checkpoint %s lacks two-component row authority evidence", checkpoint.Name)
	}
	wantCompact := doc.Backend == "fast" && checkpoint.Name != "encrypt_a_level5" && checkpoint.Name != "encrypt_b_level5" && checkpoint.Name != "encrypt_c_q5_level5"
	if checkpoint.FastCompactPrefix != wantCompact {
		return fmt.Errorf("checkpoint %s has unexpected Fast Q-prefix layout declaration", checkpoint.Name)
	}
	activeRows := state.QPrefixRows
	for component := 0; component < 2; component++ {
		rowLengths, rowHashes := checkpoint.PhysicalRowLengths[component], checkpoint.RowSHA256[component]
		if len(rowLengths) < activeRows || len(rowHashes) != activeRows {
			return fmt.Errorf("checkpoint %s component %d has incomplete active-row evidence", checkpoint.Name, component)
		}
		for row := 0; row < activeRows; row++ {
			if rowLengths[row] != doc.Parameters.RingN || rowHashes[row] == "" {
				return fmt.Errorf("checkpoint %s component %d q%d active row evidence is invalid", checkpoint.Name, component, row)
			}
		}
		if doc.Backend == "fast" {
			for row := activeRows; row < len(rowLengths); row++ {
				if wantCompact && rowLengths[row] != 0 {
					return fmt.Errorf("checkpoint %s Fast compact component %d retains unauthorized q%d row", checkpoint.Name, component, row)
				}
				if !wantCompact && rowLengths[row] != doc.Parameters.RingN {
					return fmt.Errorf("checkpoint %s Fast native input has unexpected q%d row length", checkpoint.Name, row)
				}
			}
		}
	}
	return nil
}

func expectedPublicCheckpointScale(params effectiveParameters, checkpoint string) (string, error) {
	if len(params.QPrimes) <= publicMulRescaleInputLvl {
		return "", errors.New("public profile is missing q5 for checkpoint Scale validation")
	}
	defaultValue, _, err := big.ParseFloat(params.DefaultScale, 10, rlwe.ScalePrecision, big.ToNearestEven)
	if err != nil {
		return "", fmt.Errorf("parse frozen default Scale: %w", err)
	}
	defaultInteger, _ := defaultValue.Int(nil)
	defaultScale := rlwe.NewScale(defaultInteger)
	q5, ok := new(big.Int).SetString(params.QPrimes[publicMulRescaleInputLvl], 10)
	if !ok || q5.Sign() <= 0 {
		return "", errors.New("public profile has invalid q5 metadata")
	}
	q5Scale := rlwe.NewScale(q5)
	var expected rlwe.Scale
	switch checkpoint {
	case "encrypt_c_q5_level5":
		expected = q5Scale
	case "mulrelin_level5":
		expected = rlwe.NewScale(new(big.Int).Mul(defaultScale.BigInt(), q5Scale.BigInt()))
	case "encrypt_a_level5", "encrypt_b_level5", "add_level5", "rescale_q5_level4", "rotate_level4", "drop_level0":
		expected = defaultScale
	default:
		if isPublicBootstrapCheckpoint(checkpoint, 6) || isPublicBootstrapCheckpoint(checkpoint, 2) {
			expected = defaultScale
		} else {
			return "", fmt.Errorf("unknown public checkpoint %q for Scale validation", checkpoint)
		}
	}
	return expected.Value.Text('e', 80), nil
}

func isPublicBootstrapCheckpoint(name string, budget int) bool {
	if name == "bootstrap_first_cold" {
		return budget == 2 || budget == 6
	}
	if budget == 2 {
		return name == "bootstrap_later_warm"
	}
	if budget != 6 || !strings.HasPrefix(name, "bootstrap_warm_") {
		return false
	}
	index, err := strconv.Atoi(strings.TrimPrefix(name, "bootstrap_warm_"))
	return err == nil && index >= 1 && index <= 5 && name == fmt.Sprintf("bootstrap_warm_%02d", index)
}

func validatePublicCapacity(capacity perfmeasure.PublicCapacityPlan) error {
	parsePositive := func(name, value string) (*big.Int, error) {
		bound, ok := new(big.Int).SetString(value, 10)
		if !ok || bound.Sign() <= 0 {
			return nil, fmt.Errorf("%s is not a positive integer", name)
		}
		return bound, nil
	}
	q0123, err := parsePositive("Q0123 product", capacity.Q0123)
	if err != nil {
		return err
	}
	q0, err := parsePositive("q0", capacity.Q0)
	if err != nil {
		return err
	}
	for name, raw := range map[string]string{
		"input A": capacity.InputA, "input B": capacity.InputB, "input C": capacity.InputC,
		"AddNew": capacity.Add, "MulRelinNew": capacity.MulRelin, "Rescale(q5)": capacity.Rescale,
	} {
		bound, err := parsePositive(name+" bound", raw)
		if err != nil {
			return err
		}
		if new(big.Int).Lsh(bound, 1).Cmp(q0123) >= 0 {
			return fmt.Errorf("%s bound does not fit strict Q0123 centered capacity", name)
		}
	}
	rescale, _ := new(big.Int).SetString(capacity.Rescale, 10)
	if new(big.Int).Lsh(rescale, 1).Cmp(q0) >= 0 {
		return errors.New("Rescale(q5) bound does not fit strict terminal q0 capacity")
	}
	if _, err := parsePositive("MulRelin divisor", capacity.MulDivisor); err != nil {
		return err
	}
	return nil
}

func publicOracleCheckpoints(slots int) (map[string][]complex128, error) {
	a, b, c, err := perfmeasure.PublicMulRescaleWorkload(slots)
	if err != nil {
		return nil, err
	}
	add := publicAdd(a, b)
	product := publicMul(add, c)
	rotated := publicRotateLeft(product, 1)
	return map[string][]complex128{
		"encrypt_a_level5":    a,
		"encrypt_b_level5":    b,
		"encrypt_c_q5_level5": c,
		"add_level5":          add,
		"mulrelin_level5":     product,
		"rescale_q5_level4":   product,
		"rotate_level4":       rotated,
		"drop_level0":         rotated,
	}, nil
}

func samePublicCheckpointNames(checkpoints []publicCheckpoint, want []string) bool {
	if len(checkpoints) != len(want) {
		return false
	}
	seen := make(map[string]bool, len(checkpoints))
	for _, checkpoint := range checkpoints {
		if seen[checkpoint.Name] {
			return false
		}
		seen[checkpoint.Name] = true
	}
	for _, name := range want {
		if !seen[name] {
			return false
		}
	}
	return true
}

func validatePublicBackendPin(backend, commit string) error {
	want := map[string]string{"standard": publicStandardSHA, "fast": publicFastSHA}[backend]
	if want == "" || commit != want {
		return fmt.Errorf("%s backend commit %s differs from approved Batch021 pin %s", backend, commit, want)
	}
	return nil
}

func publicInputContract(backend string) (kind, constructor string) {
	if backend == "fast" {
		return "fast_zero_secret_public_encrypt_v1", "rlwe.NewEncryptor(full, extended bootstrap secret).EncryptNew; pinned Fast CKKS zero-secret capability yields observed c1=0"
	}
	return "standard_native_rlwe_encrypt_v1", "rlwe.NewEncryptor(full, extended generated Standard secret).EncryptNew"
}

func measurePhase(name string, operation func() error) (phaseTiming, error) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	err := operation()
	elapsed := time.Since(started).Nanoseconds()
	runtime.ReadMemStats(&after)
	return phaseTiming{Phase: name, ElapsedNS: elapsed, AllocBytes: after.TotalAlloc - before.TotalAlloc, Allocs: after.Mallocs - before.Mallocs, Samples: 1, Available: true}, err
}

func environmentSetting(name string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return "runtime-default"
}

func measurePublicPhase[T any](name string, operation func() (T, error)) (T, phaseTiming, error) {
	var value T
	phase, err := measurePhase(name, func() error {
		var operationErr error
		value, operationErr = operation()
		return operationErr
	})
	return value, phase, err
}

func primaryMeasurementSourceFingerprint(root string) (string, error) {
	paths := []string{"cmd/perfprobe", "internal/perfmeasure", "internal/numericalmetrics", "go.mod", "go.sum"}
	h := sha256.New()
	for _, path := range paths {
		object, err := gitOutput(root, "rev-parse", "HEAD:"+path)
		if err != nil {
			return "", fmt.Errorf("measurement source fingerprint %s: %w", path, err)
		}
		_, _ = h.Write([]byte(path + "\x00" + object + "\x00"))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func secretPrefixEqual(a, b *rlwe.SecretKey, rows int) bool {
	if a == nil || b == nil || rows < 1 || len(a.Value.Q.Coeffs) < rows || len(b.Value.Q.Coeffs) < rows {
		return false
	}
	for row := 0; row < rows; row++ {
		if !slices.Equal(a.Value.Q.Coeffs[row], b.Value.Q.Coeffs[row]) {
			return false
		}
	}
	return true
}

func isCiphertextComponentNonzero(ct *rlwe.Ciphertext, component int) (bool, error) {
	if ct == nil || component < 0 || component >= len(ct.Value) {
		return false, errors.New("ciphertext component is unavailable")
	}
	for _, row := range ct.Value[component].Coeffs {
		for _, coefficient := range row {
			if coefficient != 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func hashUint64Row(values []uint64) string {
	h := sha256.New()
	var encoded [8]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(encoded[:], value)
		_, _ = h.Write(encoded[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func publicAdd(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] + b[i]
	}
	return out
}

func publicMul(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] * b[i]
	}
	return out
}

func publicRotateLeft(values []complex128, shift int) []complex128 {
	out := make([]complex128, len(values))
	if len(values) == 0 {
		return out
	}
	shift %= len(values)
	for i := range out {
		out[i] = values[(i+shift)%len(values)]
	}
	return out
}

func decodeComplexValues(values []complexValue) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func assertNoBootstrapPhaseCalls(actual int) error {
	if actual != 0 {
		return fmt.Errorf("zero-call preflight recorded %d Bootstrap calls", actual)
	}
	return nil
}
