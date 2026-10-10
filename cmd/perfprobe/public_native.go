package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"slices"
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
	PhysicalRowLengths [][]int              `json:"physical_row_lengths_by_component"`
	RowSHA256          [][]string           `json:"active_row_sha256_by_component"`
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
	Capacity                perfmeasure.PublicCapacityPlan  `json:"capacity_plan"`
	NumericalGate           float64                         `json:"max_complex_numerical_gate"`
	BootstrapBudget         int                             `json:"bootstrap_budget"`
	BootstrapCalls          int                             `json:"actual_bootstrap_calls"`
	ExecutionPhases         []phaseTiming                   `json:"execution_phases"`
	Checkpoints             []publicCheckpoint              `json:"checkpoints"`
	Limitations             []string                        `json:"limitations"`
}

type publicVectorDocument struct {
	SchemaVersion  string                    `json:"schema_version"`
	Mode           string                    `json:"mode"`
	Backend        string                    `json:"backend"`
	BackendCommit  string                    `json:"backend_commit"`
	PrimaryCommit  string                    `json:"primary_commit"`
	SourceSHA256   string                    `json:"primary_measurement_source_sha256"`
	ConfigSHA256   string                    `json:"config_sha256"`
	QPSHA256       string                    `json:"qp_sha256"`
	InputSHA256    string                    `json:"input_sha256"`
	WorkloadSHA256 string                    `json:"workload_sha256"`
	E              int                       `json:"ephemeral_secret_weight"`
	Checkpoints    map[string][]complexValue `json:"checkpoints"`
}

type publicPairCheckpoint struct {
	Name          string          `json:"name"`
	StateMatched  bool            `json:"level_scale_degree_matched"`
	FastState     ciphertextState `json:"fast_state"`
	StandardState ciphertextState `json:"standard_state"`
	Metrics       vectorMetrics   `json:"fast_vs_standard"`
	Pass          bool            `json:"pass"`
}

type publicPairDocument struct {
	SchemaVersion  string                 `json:"schema_version"`
	GeneratedAt    time.Time              `json:"generated_at"`
	Status         string                 `json:"status"`
	NumericalGate  float64                `json:"max_complex_gate"`
	FastCommit     string                 `json:"fast_backend_commit"`
	StandardCommit string                 `json:"standard_backend_commit"`
	PrimaryCommit  string                 `json:"primary_commit"`
	SourceSHA256   string                 `json:"primary_measurement_source_sha256"`
	InputSHA256    string                 `json:"input_sha256"`
	WorkloadSHA256 string                 `json:"workload_sha256"`
	ConfigSHA256   string                 `json:"config_sha256"`
	QPSHA256       string                 `json:"qp_sha256"`
	E              int                    `json:"ephemeral_secret_weight"`
	BootstrapCalls int                    `json:"actual_bootstrap_calls"`
	Checkpoints    []publicPairCheckpoint `json:"checkpoints"`
	Limitations    []string               `json:"limitations"`
}

func runPublicNativePreflight(opts cliOptions, configHash string, residual ckks.Parameters, params bootstrapping.Parameters, effective perfmeasure.EffectiveParameters) error {
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
	addCheckpoint := func(name string, ct *rlwe.Ciphertext, expected []complex128) error {
		decoded, phase, decodeErr := measurePublicPhase("native_decrypt_decode_"+name, func() ([]complex128, error) {
			return backend.decode(ct)
		})
		if decodeErr != nil {
			return fmt.Errorf("native Level/Scale decode %s: %w", name, decodeErr)
		}
		phaseTimings = append(phaseTimings, phase)
		cp, cpErr := makePublicCheckpoint(backend, name, ct, decoded, expected, q)
		if cpErr != nil {
			return cpErr
		}
		if !cp.OraclePass {
			return fmt.Errorf("%s cleartext oracle max %.12g exceeds %.12g", name, cp.Oracle.MaxComplexDifference, publicNumericalGate)
		}
		checkpoints = append(checkpoints, cp)
		vectorMap[name] = encodeVector(decoded)
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

	if err := addCheckpoint("encrypt_a_level5", ctA, valuesA); err != nil {
		return err
	}
	if err := addCheckpoint("encrypt_b_level5", ctB, valuesB); err != nil {
		return err
	}
	if err := addCheckpoint("encrypt_c_q5_level5", ctC, valuesC); err != nil {
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
	if err := addCheckpoint("add_level5", add, expectedAdd); err != nil {
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
	if err := addCheckpoint("mulrelin_level5", product, expectedProduct); err != nil {
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
	if err := addCheckpoint("rescale_q5_level4", rescaled, expectedProduct); err != nil {
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
	if err := addCheckpoint("rotate_level4", rotated, expectedRotated); err != nil {
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
	if err := addCheckpoint("drop_level0", dropped, expectedRotated); err != nil {
		return err
	}

	// Bootstrap is deliberately never invoked in Batch021. Public evaluator
	// construction/dispatch is verified, while first/cold and later/warm phases
	// remain explicitly unavailable under the zero-call budget.
	phaseTimings = append(phaseTimings,
		phaseTiming{Phase: "first_cold_bootstrap", Available: false, Reason: "Batch021 enforces a zero actual-Bootstrap budget"},
		phaseTiming{Phase: "later_warm_bootstrap", Available: false, Reason: "Batch021 enforces a zero actual-Bootstrap budget"},
	)
	inputKind, constructor := publicInputContract(backend.name)
	doc := publicNativeDocument{
		SchemaVersion: "fast-standard-public-native-preflight.v1", Mode: "public-native", Status: "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN",
		Timestamp: time.Now().UTC(), Profile: opts.profile, Backend: backend.name, BackendCommit: secondaryCommit,
		BackendRef: opts.backendRef, BackendCheckoutRef: secondaryRef, BackendPath: secondaryRoot, BackendClean: true, PrimaryPath: primaryRoot,
		PrimaryCommit: primaryCommit, PrimaryClean: true, PrimarySourceSHA256: sourceHash, BuildTag: "perf_" + backend.name,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
		ConfigPath: opts.config, ConfigSHA256: configHash, QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash,
		Parameters: effective, EphemeralSecretWeight: params.EphemeralSecretWeight, InputSlots: effective.InputSlots,
		InputKind: inputKind, InputConstructor: constructor, InputC1Nonzero: c1Nonzero, PublicEvaluatorDispatch: backend.dispatch,
		Capacity: capacity, NumericalGate: publicNumericalGate, BootstrapBudget: 0, BootstrapCalls: 0,
		ExecutionPhases: phaseTimings, Checkpoints: checkpoints,
		Limitations: []string{
			"This is a public pre-Bootstrap lifecycle only; no Bootstrap method was called, so public Bootstrap acceptance/output and cold/warm timings remain unverified.",
			"Fast zero-c1 is the pinned implementation's declared zero-secret simulation mode, not secure public-key encryption; Standard uses native generated-secret encryption.",
			"Existing fastdiag P93/E0 internal traces are not claimed as E32 Standard-comparable stages; in-circuit tracing is unavailable under the zero-Bootstrap budget.",
		},
	}
	vectorDoc := publicVectorDocument{
		SchemaVersion: "fast-standard-public-native-vectors.v1", Mode: "public-native", Backend: backend.name,
		BackendCommit: secondaryCommit, PrimaryCommit: primaryCommit, SourceSHA256: sourceHash, ConfigSHA256: configHash,
		QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash, E: params.EphemeralSecretWeight, Checkpoints: vectorMap,
	}
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		return err
	}
	if err := writeExclusiveJSON(opts.vectorsOut, vectorDoc); err != nil {
		return err
	}
	fmt.Printf("%s %s public-native preflight=%s checkpoints=%d bootstrap_calls=0\n", opts.profile, backend.name, opts.out, len(checkpoints))
	return nil
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

func (b *publicBackend) decode(ct *rlwe.Ciphertext) ([]complex128, error) {
	if ct == nil || ct.MetaData == nil || ct.Level() < 0 || ct.Degree() < 1 {
		return nil, errors.New("native decode requires a non-nil ciphertext with metadata")
	}
	params, secret := b.full, b.bootstrapSecret
	if ct.Level() <= b.residual.MaxLevel() {
		params, secret = b.residual, b.secret
	}
	plain := rlwe.NewDecryptor(params, secret).DecryptNew(ct)
	decoded := make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(plain, decoded); err != nil {
		return nil, err
	}
	if len(decoded) < b.maxSlots {
		return nil, errors.New("native Decode returned fewer slots than the frozen fixture")
	}
	return decoded[:b.maxSlots], nil
}

func makePublicCheckpoint(backend *publicBackend, name string, ct *rlwe.Ciphertext, decoded, expected []complex128, q []uint64) (publicCheckpoint, error) {
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
		for component, polynomial := range ct.Value {
			coeffRows := polynomial.Coeffs
			for row := rows; row < len(coeffRows); row++ {
				if len(coeffRows[row]) != 0 {
					return publicCheckpoint{}, fmt.Errorf("%s Fast component %d retains dormant q%d row (%d coefficients)", name, component, row, len(coeffRows[row]))
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
	if err := validatePublicPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
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
		BootstrapCalls: 0, Limitations: []string{"Only the public pre-Bootstrap chain is compared; no public Bootstrap acceptance or Bootstrap output metric is claimed."},
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
	for _, checkpoints := range [][]publicCheckpoint{standard.Checkpoints, fast.Checkpoints} {
		for _, checkpoint := range checkpoints {
			if !checkpoint.OraclePass || checkpoint.Oracle.MaxComplexDifference > publicNumericalGate {
				return fmt.Errorf("checkpoint %s failed its per-backend cleartext oracle", checkpoint.Name)
			}
		}
	}
	return nil
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
