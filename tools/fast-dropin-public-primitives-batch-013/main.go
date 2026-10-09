package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	logN                      = 13
	logSlots                  = 4
	defaultScaleLog2          = 45
	standardCommit            = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	fastCommit                = "00ac70ba136d190fa31bbb26c2f51d003a221634"
	catastrophicOracleStopMax = 1.0 // Inherited fail-stop guard from 005; not an acceptance threshold.
)

var testLevels = []int{3, 2}

type profile struct {
	LogN          int   `json:"log_n"`
	LogQ          []int `json:"log_q_bits"`
	LogP          []int `json:"log_p_bits"`
	ScaleLog2     int   `json:"default_scale_log2"`
	LogSlots      int   `json:"log_slots"`
	TestLevels    []int `json:"test_levels"`
	SecretHamming int   `json:"secret_hamming_weight"`
	Rotation      int   `json:"rotation_left"`
}

type metric struct {
	Finite        bool    `json:"finite"`
	ComplexRMSE   float64 `json:"complex_rmse"`
	MaxComplexErr float64 `json:"max_complex_error"`
	MaxRealErr    float64 `json:"max_real_error"`
	MaxImagErr    float64 `json:"max_imag_error"`
}

type ciphertextInfo struct {
	Level                  int             `json:"level"`
	Degree                 int             `json:"degree"`
	ScaleLog2              float64         `json:"scale_log2"`
	LogDimensions          ring.Dimensions `json:"log_dimensions"`
	IsBatched              bool            `json:"is_batched"`
	IsNTT                  bool            `json:"is_ntt"`
	IsMontgomery           bool            `json:"is_montgomery"`
	RowsPerComponent       []int           `json:"rows_per_component"`
	ActiveRowsMaterialized bool            `json:"all_active_q_rows_materialized"`
	C1Zero                 bool            `json:"c1_zero"`
	C1NonzeroCoefficients  int             `json:"c1_nonzero_coefficients"`
	C1ActiveCoefficients   int             `json:"c1_active_coefficients"`
}

type inputEvidence struct {
	Name   string         `json:"name"`
	State  ciphertextInfo `json:"state"`
	Oracle metric         `json:"decoded_vs_plaintext_oracle"`
}

type keyLookups struct {
	Relinearization int `json:"relinearization_key"`
	Galois          int `json:"galois_key"`
}

type operationEvidence struct {
	ID                 string          `json:"id"`
	API                string          `json:"api"`
	Overload           string          `json:"overload"`
	Path               string          `json:"actual_path"`
	Status             string          `json:"status"`
	Inputs             []string        `json:"inputs,omitempty"`
	Output             *ciphertextInfo `json:"output,omitempty"`
	Oracle             *metric         `json:"decoded_vs_plaintext_oracle,omitempty"`
	ExpectedLevel      *int            `json:"expected_level,omitempty"`
	ExpectedDegree     *int            `json:"expected_degree,omitempty"`
	ExpectedScaleLog2  *float64        `json:"expected_scale_log2,omitempty"`
	ScaleRule          string          `json:"scale_rule,omitempty"`
	FullActiveQBacking bool            `json:"full_active_q_backing_checked"`
	InputUnchanged     bool            `json:"non_aliased_inputs_unchanged"`
	OutputAlias        string          `json:"output_alias,omitempty"`
	OutputUnchanged    *bool           `json:"rejected_output_unchanged,omitempty"`
	C1PolicyPassed     bool            `json:"c1_policy_passed"`
	StateChecksPassed  bool            `json:"level_scale_degree_domain_checks_passed"`
	KeyLookupDelta     keyLookups      `json:"key_lookup_delta"`
	Error              string          `json:"error,omitempty"`
	Note               string          `json:"note,omitempty"`
}

type evidence struct {
	SchemaVersion      string              `json:"schema_version"`
	Status             string              `json:"status"`
	TimestampUTC       string              `json:"timestamp_utc"`
	GoVersion          string              `json:"go_version"`
	OS                 string              `json:"os"`
	Architecture       string              `json:"architecture"`
	PrimaryCommit      string              `json:"primary_commit"`
	PrimaryDirty       bool                `json:"primary_dirty"`
	BackendCommit      string              `json:"backend_commit"`
	BackendRef         string              `json:"backend_ref"`
	BackendDirty       bool                `json:"backend_dirty"`
	FastCapability     bool                `json:"fast_zero_secret_capability_detected"`
	BackendMode        string              `json:"backend_mode"`
	C1Policy           string              `json:"c1_policy"`
	Evaluator          string              `json:"public_evaluator"`
	Profile            profile             `json:"profile"`
	ProfileSHA256      string              `json:"profile_sha256"`
	ParameterSHA256    string              `json:"effective_qp_sha256"`
	QPrimes            []uint64            `json:"q_primes"`
	PPrimes            []uint64            `json:"p_primes"`
	FrontendSHA256     string              `json:"shared_frontend_sha256"`
	TestSourceSHA256   string              `json:"shared_test_source_sha256"`
	InputSHA256        string              `json:"deterministic_input_sha256"`
	Inputs             []inputEvidence     `json:"encrypted_inputs"`
	RelinKeyLayout     string              `json:"generated_relinearization_key_layout"`
	GaloisKeyLayout    string              `json:"generated_galois_key_layout"`
	KeyGeneration      string              `json:"evaluation_key_generation"`
	RelinKeyAvailable  bool                `json:"ordinary_relinearization_key_available_to_evaluator"`
	GaloisKeyAvailable bool                `json:"ordinary_galois_key_available_to_evaluator"`
	TotalKeyLookups    keyLookups          `json:"total_key_lookup_counts"`
	Operations         []operationEvidence `json:"operations"`
	BootstrapCalls     int                 `json:"bootstrap_calls"`
	StopReason         string              `json:"stop_reason,omitempty"`
}

type pairedOperation struct {
	ID               string     `json:"id"`
	StandardStatus   string     `json:"standard_status"`
	FastStatus       string     `json:"fast_status"`
	StandardRMSE     *float64   `json:"standard_complex_rmse,omitempty"`
	FastRMSE         *float64   `json:"fast_complex_rmse,omitempty"`
	StandardMaxError *float64   `json:"standard_max_complex_error,omitempty"`
	FastMaxError     *float64   `json:"fast_max_complex_error,omitempty"`
	FastPath         string     `json:"fast_actual_path"`
	FastKeyLookups   keyLookups `json:"fast_key_lookup_delta"`
}

type backendRun struct {
	Commit          string              `json:"commit"`
	Ref             string              `json:"ref"`
	Dirty           bool                `json:"dirty"`
	Mode            string              `json:"mode"`
	C1Policy        string              `json:"c1_policy"`
	RelinKeyLayout  string              `json:"relinearization_key_layout"`
	GaloisKeyLayout string              `json:"galois_key_layout"`
	KeyGeneration   string              `json:"evaluation_key_generation"`
	TotalLookups    keyLookups          `json:"total_key_lookup_counts"`
	Inputs          []inputEvidence     `json:"inputs"`
	Operations      []operationEvidence `json:"operations"`
}

type combinedEvidence struct {
	SchemaVersion    string            `json:"schema_version"`
	Status           string            `json:"status"`
	CreatedUTC       string            `json:"created_utc"`
	GoVersion        string            `json:"go_version"`
	OS               string            `json:"os"`
	Architecture     string            `json:"architecture"`
	PrimaryCommit    string            `json:"primary_commit_during_runs"`
	PrimaryDirty     bool              `json:"primary_dirty_during_runs"`
	FrontendSHA256   string            `json:"identical_frontend_sha256"`
	TestSourceSHA256 string            `json:"identical_test_source_sha256"`
	Profile          profile           `json:"profile"`
	ProfileSHA256    string            `json:"profile_sha256"`
	ParameterSHA     string            `json:"effective_qp_sha256"`
	QPrimes          []uint64          `json:"q_primes"`
	PPrimes          []uint64          `json:"p_primes"`
	InputSHA256      string            `json:"deterministic_input_sha256"`
	BootstrapCalls   int               `json:"bootstrap_calls"`
	Standard         backendRun        `json:"standard"`
	Fast             backendRun        `json:"fast"`
	Paired           []pairedOperation `json:"paired_checkpoints"`
}

type namedCiphertext struct {
	name string
	ct   *rlwe.Ciphertext
}

type inputSnapshot struct {
	name string
	ct   *rlwe.Ciphertext
	data []byte
}

type trackingKeySet struct {
	inner        rlwe.EvaluationKeySet
	relinLookup  int
	galoisLookup int
}

func (t *trackingKeySet) GetGaloisKey(galEl uint64) (*rlwe.GaloisKey, error) {
	t.galoisLookup++
	return t.inner.GetGaloisKey(galEl)
}

func (t *trackingKeySet) GetGaloisKeysList() []uint64 {
	return t.inner.GetGaloisKeysList()
}

func (t *trackingKeySet) GetRelinearizationKey() (*rlwe.RelinearizationKey, error) {
	t.relinLookup++
	return t.inner.GetRelinearizationKey()
}

func (t *trackingKeySet) counts() keyLookups {
	return keyLookups{Relinearization: t.relinLookup, Galois: t.galoisLookup}
}

type runContext struct {
	params    ckks.Parameters
	encoder   *ckks.Encoder
	decryptor *rlwe.Decryptor
	evaluator *ckks.Evaluator
	keys      *trackingKeySet
	secret    *rlwe.SecretKey
	inputs    map[string]*rlwe.Ciphertext
	fast      bool
	result    evidence
}

func main() {
	if err := runCLI(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() error {
	outPath := flag.String("out", "", "compact per-backend JSON evidence output")
	primaryCommit := flag.String("primary-commit", "", "Primary commit used for this run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty during this run")
	backendCommit := flag.String("backend-commit", "", "pinned Lattigo commit used for this run")
	backendRef := flag.String("backend-ref", "", "Lattigo ref or worktree label")
	backendDirty := flag.Bool("backend-dirty", false, "whether the Lattigo worktree was dirty")
	combineStandard := flag.String("combine-standard", "", "combine mode: Standard evidence JSON input")
	combineFast := flag.String("combine-fast", "", "combine mode: Fast evidence JSON input")
	flag.Parse()

	if *combineStandard != "" || *combineFast != "" {
		if *combineStandard == "" || *combineFast == "" || *outPath == "" {
			return fmt.Errorf("combine mode requires -combine-standard, -combine-fast, and -out")
		}
		return combineEvidenceFiles(*combineStandard, *combineFast, *outPath)
	}
	if *outPath == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return fmt.Errorf("-out, -primary-commit, -backend-commit, and -backend-ref are required")
	}

	result, runErr := runPrimitiveSuite()
	result.PrimaryCommit = *primaryCommit
	result.PrimaryDirty = *primaryDirty
	result.BackendCommit = *backendCommit
	result.BackendRef = *backendRef
	result.BackendDirty = *backendDirty
	if source, err := os.ReadFile("tools/fast-dropin-public-primitives-batch-013/main.go"); err == nil {
		result.FrontendSHA256 = hashBytes(source)
	} else if runErr == nil {
		runErr = fmt.Errorf("hash shared frontend source: %w", err)
	}
	if source, err := os.ReadFile("tools/fast-dropin-public-primitives-batch-013/main_test.go"); err == nil {
		result.TestSourceSHA256 = hashBytes(source)
	} else if runErr == nil {
		runErr = fmt.Errorf("hash shared test source: %w", err)
	}
	if err := writeJSON(*outPath, result); err != nil {
		return err
	}
	return runErr
}

func runPrimitiveSuite() (evidence, error) {
	profileUsed := profile{
		LogN: logN, LogQ: []int{55, 39, 40, 39}, LogP: []int{60},
		ScaleLog2: defaultScaleLog2, LogSlots: logSlots,
		TestLevels: append([]int(nil), testLevels...), SecretHamming: 192, Rotation: 1,
	}
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: profileUsed.LogN, LogQ: append([]int(nil), profileUsed.LogQ...),
		LogP: append([]int(nil), profileUsed.LogP...), LogDefaultScale: profileUsed.ScaleLog2,
		Xs: ring.Ternary{H: profileUsed.SecretHamming},
	})
	if err != nil {
		return evidence{}, fmt.Errorf("create frozen CKKS parameters: %w", err)
	}
	if params.MaxLevel() != 3 || len(params.Q()) != 4 || len(params.P()) != 1 {
		return evidence{}, fmt.Errorf("unexpected effective CKKS Q/P profile: maxLevel=%d Q=%d P=%d", params.MaxLevel(), len(params.Q()), len(params.P()))
	}

	_, fastCapability := any(params).(interface{ FastCKKSZeroSecretSimulation() })
	mode, c1Policy := "genuine-standard-encryption", "nonzero"
	if fastCapability {
		mode, c1Policy = "fast-p0-zero-secret-simulation", "zero"
	}
	encoder := ckks.NewEncoder(params)
	keygen := ckks.NewKeyGenerator(params)
	sk := keygen.GenSecretKeyNew()
	relinKey := keygen.GenRelinearizationKeyNew(sk)
	galoisKey := keygen.GenGaloisKeyNew(params.GaloisElementForRotation(1), sk)
	relinLayout, relinLayoutOK := describeKeyLayout(relinKey)
	galoisLayout, galoisLayoutOK := describeKeyLayout(galoisKey)
	if !relinLayoutOK || !galoisLayoutOK {
		return evidence{}, fmt.Errorf("ckks.NewKeyGenerator produced a non-Standard evaluation-key layout: relin=%s galois=%s", relinLayout, galoisLayout)
	}
	keys := &trackingKeySet{inner: rlwe.NewMemEvaluationKeySet(relinKey, galoisKey)}
	evaluator := ckks.NewEvaluator(params, keys)
	decryptor := rlwe.NewDecryptor(params, sk)
	if evaluator == nil || evaluator.Evaluator == nil {
		return evidence{}, fmt.Errorf("ckks.NewEvaluator did not construct the public evaluator")
	}
	profileJSON, _ := json.Marshal(profileUsed)
	qpJSON, _ := json.Marshal(struct {
		Q []uint64 `json:"q"`
		P []uint64 `json:"p"`
	}{params.Q(), params.P()})
	valuesA, valuesB, plain := deterministicInputs(1 << logSlots)
	inputHash := hashInputs(valuesA, valuesB, plain)
	result := evidence{
		SchemaVersion: "fast-dropin-public-primitives-batch-013.v1", Status: "running",
		TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano), GoVersion: runtime.Version(),
		OS: runtime.GOOS, Architecture: runtime.GOARCH,
		FastCapability: fastCapability, BackendMode: mode, C1Policy: c1Policy,
		Evaluator: "ckks.NewEvaluator(params, instrumented ordinary rlwe.EvaluationKeySet)",
		Profile:   profileUsed, ProfileSHA256: hashBytes(profileJSON), ParameterSHA256: hashBytes(qpJSON),
		QPrimes: params.Q(), PPrimes: params.P(), InputSHA256: inputHash,
		RelinKeyLayout: relinLayout, GaloisKeyLayout: galoisLayout,
		KeyGeneration:     "ckks.NewKeyGenerator(params): native Standard-layout relinearization and Galois keys",
		RelinKeyAvailable: true, GaloisKeyAvailable: true, BootstrapCalls: 0,
	}
	ctx := runContext{params: params, encoder: encoder, decryptor: decryptor, evaluator: evaluator, keys: keys, secret: sk, inputs: map[string]*rlwe.Ciphertext{}, fast: fastCapability, result: result}

	for _, level := range testLevels {
		_, err := ctx.encryptAndRecord(fmt.Sprintf("encrypt-l%d-a", level), valuesA, level)
		if err != nil {
			return ctx.result, err
		}
		_, err = ctx.encryptAndRecord(fmt.Sprintf("encrypt-l%d-b", level), valuesB, level)
		if err != nil {
			return ctx.result, err
		}
	}
	ct3a, ct3b := ctx.input("encrypt-l3-a"), ctx.input("encrypt-l3-b")
	ct2a, ct2b := ctx.input("encrypt-l2-a"), ctx.input("encrypt-l2-b")
	if err := ctx.runAddSub(ct3a, ct3b, valuesA, valuesB, 3); err != nil {
		return ctx.result, err
	}
	if err := ctx.runAddSub(ct2a, ct2b, valuesA, valuesB, 2); err != nil {
		return ctx.result, err
	}
	if err := ctx.runMulRelin(ct3a, ct3b, valuesA, valuesB, 3); err != nil {
		return ctx.result, err
	}
	if err := ctx.runMulRelin(ct2a, ct2b, valuesA, valuesB, 2); err != nil {
		return ctx.result, err
	}
	if err := ctx.runNegativeControls(ct3a, ct3b); err != nil {
		return ctx.result, err
	}
	if err := ctx.runRescaleAndComposition(ct3a, ct3b, valuesA, valuesB); err != nil {
		return ctx.result, err
	}

	ctx.result.TotalKeyLookups = keys.counts()
	if fastCapability && (keys.relinLookup != 0 || keys.galoisLookup != 0) {
		return ctx.result, fmt.Errorf("Fast zero-secret P0 path performed native evaluation-key lookups: %+v", keys.counts())
	}
	if !fastCapability && (keys.relinLookup == 0 || keys.galoisLookup == 0) {
		return ctx.result, fmt.Errorf("Standard path did not demonstrate native relin and Galois key access: %+v", keys.counts())
	}
	ctx.result.Status = "all_p0_public_primitive_checkpoints_passed"
	return ctx.result, nil
}

func (ctx *runContext) encryptAndRecord(name string, values []complex128, level int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(ctx.params, level)
	pt.Scale = rlwe.NewScale(math.Exp2(defaultScaleLog2))
	pt.IsNTT = true
	pt.IsMontgomery = false
	pt.LogDimensions = ring.Dimensions{Rows: 1, Cols: logSlots}
	if err := ctx.encoder.Encode(values, pt); err != nil {
		return nil, ctx.fail(name, err)
	}
	ct, err := rlwe.NewEncryptor(ctx.params, ctx.secret).EncryptNew(pt)
	if err != nil {
		return nil, ctx.fail(name, err)
	}
	info := summarizeCiphertext(ct)
	measurement, err := ctx.oracle(ct, values)
	input := inputEvidence{Name: name, State: info, Oracle: measurement}
	ctx.result.Inputs = append(ctx.result.Inputs, input)
	if err != nil {
		return nil, ctx.fail(name, err)
	}
	if err := ctx.validateCiphertext(ct, level, 1, pt.Scale); err != nil {
		return nil, ctx.fail(name, err)
	}
	if !ctx.c1MatchesPolicy(info) {
		return nil, ctx.fail(name, fmt.Errorf("EncryptNew c1 violates %s backend policy", ctx.result.C1Policy))
	}
	ctx.inputs[name] = ct
	return ct, nil
}

func (ctx *runContext) input(name string) *rlwe.Ciphertext {
	return ctx.inputs[name]
}

func (ctx *runContext) runAddSub(a, b *rlwe.Ciphertext, valuesA, valuesB []complex128, level int) error {
	addWant := addValues(valuesA, valuesB)
	subWant := subValues(valuesA, valuesB)
	addScale := a.Scale

	if level == 3 {
		if err := ctx.perform("addnew-ctct-l3", "AddNew", "ciphertext/ciphertext", "public_add_sub", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, addWant, level, 1, addScale, "unchanged", "", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.AddNew(a, b)
		}); err != nil {
			return err
		}
		if err := ctx.perform("subnew-ctct-l3", "SubNew", "ciphertext/ciphertext", "public_add_sub", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, subWant, level, 1, addScale, "unchanged", "", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.SubNew(a, b)
		}); err != nil {
			return err
		}
	} else {
		outAdd := ckks.NewCiphertext(ctx.params, 1, level)
		if err := ctx.perform("add-ctct-l2", "Add", "ciphertext/ciphertext", "public_add_sub", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, addWant, level, 1, addScale, "unchanged", "", func() (*rlwe.Ciphertext, error) {
			return outAdd, ctx.evaluator.Add(a, b, outAdd)
		}); err != nil {
			return err
		}
		outSub := ckks.NewCiphertext(ctx.params, 1, level)
		if err := ctx.perform("sub-ctct-l2", "Sub", "ciphertext/ciphertext", "public_add_sub", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, subWant, level, 1, addScale, "unchanged", "", func() (*rlwe.Ciphertext, error) {
			return outSub, ctx.evaluator.Sub(a, b, outSub)
		}); err != nil {
			return err
		}
	}

	if level == 3 {
		constant := complex(0.125, -0.0625)
		if err := ctx.perform("add-scalar-l3", "AddNew", "complex128 scalar", "public_add_sub", "", []namedCiphertext{{"a", a}}, []namedCiphertext{{"a", a}}, addScalar(valuesA, constant), level, 1, addScale, "unchanged", "supported scalar overload", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.AddNew(a, constant)
		}); err != nil {
			return err
		}
		if err := ctx.perform("sub-vector-l3", "SubNew", "[]complex128 plaintext", "public_add_sub", "", []namedCiphertext{{"a", a}}, []namedCiphertext{{"a", a}}, subValues(valuesA, deterministicPlaintext(len(valuesA))), level, 1, addScale, "unchanged", "supported encoded-plaintext overload", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.SubNew(a, deterministicPlaintext(len(valuesA)))
		}); err != nil {
			return err
		}
		aliasAdd := a.CopyNew()
		if err := ctx.perform("add-alias-l3", "Add", "ciphertext/ciphertext", "public_add_sub", "op0", []namedCiphertext{{"a", aliasAdd}, {"b", b}}, []namedCiphertext{{"b", b}}, addWant, level, 1, addScale, "unchanged", "opOut aliases op0; op0 mutation is intentional", func() (*rlwe.Ciphertext, error) {
			return aliasAdd, ctx.evaluator.Add(aliasAdd, b, aliasAdd)
		}); err != nil {
			return err
		}
		aliasSub := a.CopyNew()
		if err := ctx.perform("sub-alias-l3", "Sub", "ciphertext/ciphertext", "public_add_sub", "op0", []namedCiphertext{{"a", aliasSub}, {"b", b}}, []namedCiphertext{{"b", b}}, subWant, level, 1, addScale, "unchanged", "opOut aliases op0; op0 mutation is intentional", func() (*rlwe.Ciphertext, error) {
			return aliasSub, ctx.evaluator.Sub(aliasSub, b, aliasSub)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *runContext) runMulRelin(a, b *rlwe.Ciphertext, valuesA, valuesB []complex128, level int) error {
	productScale := a.Scale.Mul(b.Scale)
	productWant := multiplyValues(valuesA, valuesB)
	if level == 3 {
		if err := ctx.perform("mulrelinnew-ctct-l3", "MulRelinNew", "ciphertext/ciphertext", "mulrelin_ctct", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, productWant, level, 1, productScale, "multiply scales", "independent pointwise complex product oracle", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.MulRelinNew(a, b)
		}); err != nil {
			return err
		}
		squareScale := a.Scale.Mul(a.Scale)
		if err := ctx.perform("mulrelinnew-square-l3", "MulRelinNew", "ciphertext squared", "mulrelin_ctct", "", []namedCiphertext{{"a", a}}, []namedCiphertext{{"a", a}}, multiplyValues(valuesA, valuesA), level, 1, squareScale, "square scales", "independent complex square oracle", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.MulRelinNew(a, a)
		}); err != nil {
			return err
		}
		aliasA := a.CopyNew()
		if err := ctx.perform("mulrelin-alias-op0-l3", "MulRelin", "ciphertext/ciphertext", "mulrelin_ctct", "op0", []namedCiphertext{{"a", aliasA}, {"b", b}}, []namedCiphertext{{"b", b}}, productWant, level, 1, productScale, "multiply scales", "opOut aliases op0; supported in-place product path", func() (*rlwe.Ciphertext, error) {
			return aliasA, ctx.evaluator.MulRelin(aliasA, b, aliasA)
		}); err != nil {
			return err
		}
		if err := ctx.perform("mulrelin-scalar-l3", "MulRelinNew", "int64 scalar", "mulrelin_plaintext_scalar", "", []namedCiphertext{{"a", a}}, []namedCiphertext{{"a", a}}, scaleValues(valuesA, 2), level, 1, a.Scale, "Gaussian-integer scalar: scale unchanged", "public method delegates supported scalar operand to Mul; no relin key is required", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.MulRelinNew(a, int64(2))
		}); err != nil {
			return err
		}
		plain := deterministicPlaintext(len(valuesA))
		plainScale := rlwe.NewScale(ctx.params.Q()[level])
		if err := ctx.perform("mulrelin-vector-l3", "MulRelinNew", "[]complex128 plaintext", "mulrelin_plaintext_scalar", "", []namedCiphertext{{"a", a}}, []namedCiphertext{{"a", a}}, multiplyValues(valuesA, plain), level, 1, a.Scale.Mul(plainScale), "encoded plaintext scale is q[level]", "public method delegates supported vector operand to generic Mul", func() (*rlwe.Ciphertext, error) {
			return ctx.evaluator.MulRelinNew(a, plain)
		}); err != nil {
			return err
		}
	} else {
		out := ckks.NewCiphertext(ctx.params, 1, level)
		if err := ctx.perform("mulrelin-ctct-l2", "MulRelin", "ciphertext/ciphertext", "mulrelin_ctct", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, productWant, level, 1, productScale, "multiply scales", "preallocated output; independent product oracle", func() (*rlwe.Ciphertext, error) {
			return out, ctx.evaluator.MulRelin(a, b, out)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *runContext) runNegativeControls(a, b *rlwe.Ciphertext) error {
	if !ctx.fast {
		ctx.result.Operations = append(ctx.result.Operations, operationEvidence{
			ID: "mulrelin-invalid-c1", API: "MulRelin", Overload: "ciphertext/ciphertext with mutated nonzero c1",
			Path:   "not applicable: this negative control targets Fast zero-secret fail-closed semantics",
			Status: "not_applicable_to_genuine_standard", C1PolicyPassed: true, StateChecksPassed: true,
			Note: "Standard ciphertexts legitimately have nonzero c1; no invalid-c1 rejection is asserted for Standard.",
		})
	} else {
		invalid := a.CopyNew()
		invalid.Value[1].Coeffs[0][0] = 1
		out := a.CopyNew()
		invalidBefore, err := invalid.MarshalBinary()
		if err != nil {
			return ctx.fail("mulrelin-invalid-c1", err)
		}
		bBefore, err := b.MarshalBinary()
		if err != nil {
			return ctx.fail("mulrelin-invalid-c1", err)
		}
		outBefore, err := out.MarshalBinary()
		if err != nil {
			return ctx.fail("mulrelin-invalid-c1", err)
		}
		countsBefore := ctx.keys.counts()
		callErr := ctx.evaluator.MulRelin(invalid, b, out)
		countsAfter := ctx.keys.counts()
		invalidAfter, err := invalid.MarshalBinary()
		if err != nil {
			return ctx.fail("mulrelin-invalid-c1", err)
		}
		bAfter, err := b.MarshalBinary()
		if err != nil {
			return ctx.fail("mulrelin-invalid-c1", err)
		}
		outAfter, err := out.MarshalBinary()
		if err != nil {
			return ctx.fail("mulrelin-invalid-c1", err)
		}
		unchanged := bytes.Equal(invalidBefore, invalidAfter) && bytes.Equal(bBefore, bAfter)
		outUnchanged := bytes.Equal(outBefore, outAfter)
		op := operationEvidence{
			ID: "mulrelin-invalid-c1", API: "MulRelin", Overload: "ciphertext/ciphertext with one mutated active c1 residue",
			Path: pathFor(ctx.fast, "mulrelin_ctct"), Inputs: []string{"invalid-a-l3", "b-l3"},
			Status: "passed_expected_rejection", InputUnchanged: unchanged, OutputUnchanged: &outUnchanged,
			FullActiveQBacking: summarizeCiphertext(invalid).ActiveRowsMaterialized && summarizeCiphertext(out).ActiveRowsMaterialized,
			C1PolicyPassed:     callErr != nil && stringsContains(callErr.Error(), "nonzero c1"),
			StateChecksPassed:  unchanged && outUnchanged,
			KeyLookupDelta:     keyLookups{Relinearization: countsAfter.Relinearization - countsBefore.Relinearization, Galois: countsAfter.Galois - countsBefore.Galois},
			Note:               "A genuine KeyLayoutStandard relin key was present in the instrumented keyset; rejection must precede key lookup and output mutation.",
		}
		if callErr == nil {
			op.Status, op.Error = "failed_expected_rejection", "Fast zero-secret MulRelin accepted nonzero c1"
		} else if !op.C1PolicyPassed {
			op.Status, op.Error = "failed_wrong_rejection", callErr.Error()
		}
		if !unchanged || !outUnchanged {
			op.Status, op.Error = "failed_transactionality", "rejected Fast MulRelin modified an input or output"
		}
		if countsAfter.Relinearization != countsBefore.Relinearization {
			op.Status, op.Error = "failed_keyless_guard", "rejected Fast MulRelin queried a native relin key"
		}
		ctx.result.Operations = append(ctx.result.Operations, op)
		if op.Status != "passed_expected_rejection" || !op.C1PolicyPassed || !op.StateChecksPassed {
			return ctx.fail(op.ID, fmt.Errorf("negative control failed: %s", op.Error))
		}
	}

	countsBefore := ctx.keys.counts()
	inputBefore, err := a.MarshalBinary()
	if err != nil {
		return ctx.fail("mulrelin-invalid-operand", err)
	}
	_, callErr := ctx.evaluator.MulRelinNew(a, struct{}{})
	countsAfter := ctx.keys.counts()
	inputAfter, err := a.MarshalBinary()
	if err != nil {
		return ctx.fail("mulrelin-invalid-operand", err)
	}
	unchanged := bytes.Equal(inputBefore, inputAfter)
	op := operationEvidence{
		ID: "mulrelin-invalid-operand", API: "MulRelinNew", Overload: "unsupported struct{} operand",
		Path: pathFor(ctx.fast, "mulrelin_plaintext_scalar"), Inputs: []string{"a-l3"},
		Status: "passed_expected_rejection", InputUnchanged: unchanged,
		KeyLookupDelta: keyLookups{Relinearization: countsAfter.Relinearization - countsBefore.Relinearization, Galois: countsAfter.Galois - countsBefore.Galois},
		C1PolicyPassed: true, StateChecksPassed: unchanged,
		Note: "An unrecognized operand remains an error; no unsupported overload was invented.",
	}
	if callErr == nil || !unchanged {
		op.Status, op.Error = "failed_expected_rejection", "unsupported operand was accepted or mutated its input"
	}
	ctx.result.Operations = append(ctx.result.Operations, op)
	if op.Status != "passed_expected_rejection" {
		return ctx.fail(op.ID, fmt.Errorf("negative unsupported-overload control failed: %s", op.Error))
	}
	return nil
}

func (ctx *runContext) runRescaleAndComposition(a, b *rlwe.Ciphertext, valuesA, valuesB []complex128) error {
	// Isolated public Rescale: use product scale 2^90 at input Level 3 and consume the logical q3.
	var product *rlwe.Ciphertext
	if err := ctx.perform("rescale-source-mulrelin-l3", "MulRelinNew", "ciphertext/ciphertext", "mulrelin_ctct", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, multiplyValues(valuesA, valuesB), 3, 1, a.Scale.Mul(b.Scale), "multiply scales", "source checkpoint for isolated Rescale", func() (*rlwe.Ciphertext, error) {
		var err error
		product, err = ctx.evaluator.MulRelinNew(a, b)
		return product, err
	}); err != nil {
		return err
	}
	productOracle := multiplyValues(valuesA, valuesB)
	q3 := ctx.params.Q()[product.Level()]
	expectedScale := product.Scale.Div(rlwe.NewScale(q3))
	out := ckks.NewCiphertext(ctx.params, 1, product.Level()-1)
	if err := ctx.perform("rescale-l3-q3", "Rescale", "ciphertext/ciphertext result", "rescale", "", []namedCiphertext{{"product", product}}, []namedCiphertext{{"product", product}}, productOracle, product.Level()-1, 1, expectedScale, "input scale / actual logical q3", fmt.Sprintf("consumes logical q3=%d; P0 input has all four active rows", q3), func() (*rlwe.Ciphertext, error) {
		return out, ctx.evaluator.Rescale(product, out)
	}); err != nil {
		return err
	}

	// The pinned public API supports RescaleTo; select a fixed target that consumes exactly q3.
	var productTo *rlwe.Ciphertext
	if err := ctx.perform("rescaleto-source-mulrelin-l3", "MulRelinNew", "ciphertext/ciphertext", "mulrelin_ctct", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, productOracle, 3, 1, a.Scale.Mul(b.Scale), "multiply scales", "source checkpoint for RescaleTo", func() (*rlwe.Ciphertext, error) {
		var err error
		productTo, err = ctx.evaluator.MulRelinNew(a, b)
		return productTo, err
	}); err != nil {
		return err
	}
	minScale := rlwe.NewScale(math.Exp2(50))
	toExpected := productTo.Scale.Div(rlwe.NewScale(ctx.params.Q()[productTo.Level()]))
	outTo := ckks.NewCiphertext(ctx.params, 1, productTo.Level())
	if err := ctx.perform("rescaleto-l3-q3", "RescaleTo", "minScale=2^50", "rescale", "", []namedCiphertext{{"product", productTo}}, []namedCiphertext{{"product", productTo}}, productOracle, productTo.Level()-1, 1, toExpected, "consume q3 once; stop above minScale/2", "RescaleTo is present in both pinned public APIs", func() (*rlwe.Ciphertext, error) {
		return outTo, ctx.evaluator.RescaleTo(productTo, minScale, outTo)
	}); err != nil {
		return err
	}

	// One source-identical public chain: Add -> MulRelin -> Rescale -> Rotate.
	addWant := addValues(valuesA, valuesB)
	var chainAdd *rlwe.Ciphertext
	if err := ctx.perform("chain-add-l3", "AddNew", "ciphertext/ciphertext", "public_add_sub", "", []namedCiphertext{{"a", a}, {"b", b}}, []namedCiphertext{{"a", a}, {"b", b}}, addWant, 3, 1, a.Scale, "unchanged", "composition checkpoint 1/4", func() (*rlwe.Ciphertext, error) {
		var err error
		chainAdd, err = ctx.evaluator.AddNew(a, b)
		return chainAdd, err
	}); err != nil {
		return err
	}
	chainProductScale := chainAdd.Scale.Mul(a.Scale)
	chainOracle := multiplyValues(addWant, valuesA)
	var chainProduct *rlwe.Ciphertext
	if err := ctx.perform("chain-mulrelin-l3", "MulRelinNew", "ciphertext/ciphertext", "mulrelin_ctct", "", []namedCiphertext{{"sum", chainAdd}, {"a", a}}, []namedCiphertext{{"sum", chainAdd}, {"a", a}}, chainOracle, 3, 1, chainProductScale, "multiply scales", "composition checkpoint 2/4", func() (*rlwe.Ciphertext, error) {
		var err error
		chainProduct, err = ctx.evaluator.MulRelinNew(chainAdd, a)
		return chainProduct, err
	}); err != nil {
		return err
	}
	chainRescaleScale := chainProduct.Scale.Div(rlwe.NewScale(ctx.params.Q()[3]))
	chainRescaled := ckks.NewCiphertext(ctx.params, 1, 2)
	if err := ctx.perform("chain-rescale-l3-q3", "Rescale", "ciphertext/ciphertext result", "rescale", "", []namedCiphertext{{"product", chainProduct}}, []namedCiphertext{{"product", chainProduct}}, chainOracle, 2, 1, chainRescaleScale, "input scale / actual logical q3", "composition checkpoint 3/4", func() (*rlwe.Ciphertext, error) {
		return chainRescaled, ctx.evaluator.Rescale(chainProduct, chainRescaled)
	}); err != nil {
		return err
	}
	rotatedOracle := rotateLeft(chainOracle, 1)
	if err := ctx.perform("chain-rotate-l2", "RotateNew", "left rotation by 1", "rotate", "", []namedCiphertext{{"rescaled", chainRescaled}}, []namedCiphertext{{"rescaled", chainRescaled}}, rotatedOracle, 2, 1, chainRescaleScale, "unchanged", "composition checkpoint 4/4; Standard key-backed vs Fast shared-core branch", func() (*rlwe.Ciphertext, error) {
		return ctx.evaluator.RotateNew(chainRescaled, 1)
	}); err != nil {
		return err
	}
	return nil
}

func (ctx *runContext) perform(id, api, overload, pathKind, alias string, inputs, immutable []namedCiphertext, oracle []complex128, level, degree int, scale rlwe.Scale, scaleRule, note string, call func() (*rlwe.Ciphertext, error)) error {
	op := operationEvidence{
		ID: id, API: api, Overload: overload, Path: pathFor(ctx.fast, pathKind), Status: "running",
		ExpectedLevel: intPtr(level), ExpectedDegree: intPtr(degree), ExpectedScaleLog2: floatPtr(scale.Log2()),
		ScaleRule: scaleRule, OutputAlias: alias, Note: note,
	}
	for _, input := range inputs {
		op.Inputs = append(op.Inputs, input.name)
	}
	snapshots, err := snapshotInputs(immutable)
	if err != nil {
		return ctx.fail(id, err)
	}
	countsBefore := ctx.keys.counts()
	output, callErr := call()
	countsAfter := ctx.keys.counts()
	op.KeyLookupDelta = keyLookups{Relinearization: countsAfter.Relinearization - countsBefore.Relinearization, Galois: countsAfter.Galois - countsBefore.Galois}
	if callErr != nil {
		op.Status, op.Error = "operation_failed", callErr.Error()
		ctx.result.Operations = append(ctx.result.Operations, op)
		return ctx.fail(id, callErr)
	}
	if output == nil {
		op.Status, op.Error = "operation_failed", "API returned nil output without an error"
		ctx.result.Operations = append(ctx.result.Operations, op)
		return ctx.fail(id, fmt.Errorf("%s returned nil output", api))
	}
	info := summarizeCiphertext(output)
	op.Output = &info
	op.FullActiveQBacking = info.ActiveRowsMaterialized && len(info.RowsPerComponent) == degree+1
	op.C1PolicyPassed = ctx.c1MatchesPolicy(info)
	op.StateChecksPassed = ctx.validateCiphertext(output, level, degree, scale) == nil
	op.InputUnchanged = snapshotsUnchanged(snapshots)
	measurement, oracleErr := ctx.oracle(output, oracle)
	op.Oracle = &measurement
	op.Status = "passed"
	if !op.FullActiveQBacking {
		op.Status, op.Error = "failed_backing_invariant", "P0 output does not have exactly all active logical-Q rows materialized"
	} else if !op.C1PolicyPassed {
		op.Status, op.Error = "failed_c1_invariant", "output c1 does not match the backend's documented mode"
	} else if !op.StateChecksPassed {
		op.Status, op.Error = "failed_state_invariant", "output Level/Scale/Degree/metadata/domain differs from the expected CKKS contract"
	} else if !op.InputUnchanged {
		op.Status, op.Error = "failed_input_immutability", "a non-aliased input changed during the public operation"
	} else if oracleErr != nil {
		op.Status, op.Error = "failed_oracle", oracleErr.Error()
	}
	ctx.result.Operations = append(ctx.result.Operations, op)
	if op.Status != "passed" {
		return ctx.fail(id, fmt.Errorf("%s", op.Error))
	}
	return nil
}

func (ctx *runContext) validateCiphertext(ct *rlwe.Ciphertext, level, degree int, scale rlwe.Scale) error {
	if ct == nil || ct.MetaData == nil {
		return fmt.Errorf("ciphertext or metadata is nil")
	}
	info := summarizeCiphertext(ct)
	if info.Level != level || info.Degree != degree {
		return fmt.Errorf("got level/degree %d/%d, want %d/%d", info.Level, info.Degree, level, degree)
	}
	if ct.Scale.Cmp(scale) != 0 {
		return fmt.Errorf("got scale %.12g, want %.12g", ct.Scale.Log2(), scale.Log2())
	}
	if info.LogDimensions != (ring.Dimensions{Rows: 1, Cols: logSlots}) {
		return fmt.Errorf("unexpected log dimensions: %+v", info.LogDimensions)
	}
	if !info.ActiveRowsMaterialized || len(info.RowsPerComponent) != degree+1 {
		return fmt.Errorf("active logical Q backing is incomplete: rows=%v", info.RowsPerComponent)
	}
	if !info.IsNTT || info.IsMontgomery {
		return fmt.Errorf("unexpected ring domain/representation: NTT=%t Montgomery=%t", info.IsNTT, info.IsMontgomery)
	}
	return nil
}

func (ctx *runContext) c1MatchesPolicy(info ciphertextInfo) bool {
	if ctx.fast {
		return info.C1Zero
	}
	return !info.C1Zero
}

func (ctx *runContext) oracle(ct *rlwe.Ciphertext, expected []complex128) (metric, error) {
	decoded := make([]complex128, len(expected))
	if err := ctx.encoder.Decode(ctx.decryptor.DecryptNew(ct), decoded); err != nil {
		return metric{}, fmt.Errorf("DecryptNew/Decode: %w", err)
	}
	measurement := compare(expected, decoded)
	if !measurement.Finite {
		return measurement, fmt.Errorf("oracle produced non-finite values or aggregate")
	}
	if measurement.MaxComplexErr > catastrophicOracleStopMax {
		return measurement, fmt.Errorf("max complex oracle error %.12g exceeds inherited catastrophic stop guard %.12g", measurement.MaxComplexErr, catastrophicOracleStopMax)
	}
	return measurement, nil
}

func (ctx *runContext) fail(stage string, err error) error {
	ctx.result.Status = "stopped_at_" + stage
	ctx.result.StopReason = err.Error()
	return fmt.Errorf("%s: %w", stage, err)
}

func pathFor(fast bool, kind string) string {
	if !fast {
		switch kind {
		case "public_add_sub":
			return "genuine Standard ckks.Evaluator -> generic full-active-Q CKKS ring operations"
		case "mulrelin_ctct":
			return "genuine Standard ckks.Evaluator.MulRelin -> native relin-key lookup + GadgetProduct"
		case "mulrelin_plaintext_scalar":
			return "genuine Standard ckks.Evaluator.MulRelin -> public Mul scalar/plaintext overload; no relin lookup"
		case "rescale":
			return "genuine Standard ckks.Evaluator.Rescale/RescaleTo -> logical full-active-Q CKKS path"
		case "rotate":
			return "genuine Standard ckks.Evaluator.Rotate -> RLWE Automorphism + native Galois-key lookup"
		}
	}
	switch kind {
	case "public_add_sub":
		return "Fast P0 public ckks.Evaluator.Add/Sub -> generic full-active-Q CKKS path (not explicit fast.Evaluator kernel)"
	case "mulrelin_ctct":
		return "Fast P0 ckks.Evaluator.MulRelin -> mulRelinFastCKKSZeroSecret -> existing product kernel with relin=false; no native key lookup"
	case "mulrelin_plaintext_scalar":
		return "Fast P0 public ckks.Evaluator.MulRelin -> generic Mul scalar/plaintext overload over full-active-Q rows"
	case "rescale":
		return "Fast P0 public ckks.Evaluator.Rescale/RescaleTo -> generic full-active-Q CKKS path; all logical rows active"
	case "rotate":
		return "Fast P0 ckks.Evaluator.Rotate -> rotateFastCKKSZeroSecret -> shared fastcore.ApplyRows; no Galois-key lookup"
	default:
		return "unclassified public CKKS path"
	}
}

func summarizeCiphertext(ct *rlwe.Ciphertext) ciphertextInfo {
	info := ciphertextInfo{ActiveRowsMaterialized: ct != nil && ct.MetaData != nil}
	if ct == nil {
		return info
	}
	info.Level = ct.Level()
	info.Degree = ct.Degree()
	info.ScaleLog2 = ct.Scale.Log2()
	info.IsNTT = ct.IsNTT
	info.IsMontgomery = ct.IsMontgomery
	if ct.MetaData != nil {
		info.LogDimensions = ct.LogDimensions
		info.IsBatched = ct.IsBatched
	}
	for component := range ct.Value {
		poly := ct.Value[component]
		info.RowsPerComponent = append(info.RowsPerComponent, len(poly.Coeffs))
		if len(poly.Coeffs) != ct.Level()+1 {
			info.ActiveRowsMaterialized = false
		}
		for q := 0; q <= ct.Level(); q++ {
			if q >= len(poly.Coeffs) || len(poly.Coeffs[q]) == 0 {
				info.ActiveRowsMaterialized = false
				continue
			}
			if len(poly.Coeffs[q]) != 1<<logN {
				info.ActiveRowsMaterialized = false
			}
		}
	}
	if len(ct.Value) > 1 {
		for q := 0; q <= ct.Level() && q < len(ct.Value[1].Coeffs); q++ {
			for _, coefficient := range ct.Value[1].Coeffs[q] {
				info.C1ActiveCoefficients++
				if coefficient != 0 {
					info.C1NonzeroCoefficients++
				}
			}
		}
	}
	info.C1Zero = info.C1NonzeroCoefficients == 0
	return info
}

func snapshotInputs(inputs []namedCiphertext) ([]inputSnapshot, error) {
	snapshots := make([]inputSnapshot, 0, len(inputs))
	for _, input := range inputs {
		if input.ct == nil {
			return nil, fmt.Errorf("input %s is nil", input.name)
		}
		data, err := input.ct.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("snapshot input %s: %w", input.name, err)
		}
		snapshots = append(snapshots, inputSnapshot{name: input.name, ct: input.ct, data: data})
	}
	return snapshots, nil
}

func snapshotsUnchanged(snapshots []inputSnapshot) bool {
	for _, snapshot := range snapshots {
		if snapshot.ct == nil {
			return false
		}
		data, err := snapshot.ct.MarshalBinary()
		if err != nil || !bytes.Equal(snapshot.data, data) {
			return false
		}
	}
	return true
}

func deterministicInputs(slots int) (a, b, plain []complex128) {
	a, b, plain = make([]complex128, slots), make([]complex128, slots), make([]complex128, slots)
	for i := 0; i < slots; i++ {
		a[i] = complex(float64(i%7-3)/32, float64((2*i)%5-2)/64)
		b[i] = complex(float64(i%5-2)/16, float64(i%3-1)/32)
		plain[i] = complex(float64(i%3-1)/128, float64((3*i)%7-3)/256)
	}
	return
}

func deterministicPlaintext(slots int) []complex128 {
	_, _, values := deterministicInputs(slots)
	return values
}

func addValues(a, b []complex128) []complex128 {
	return zipValues(a, b, func(x, y complex128) complex128 { return x + y })
}
func subValues(a, b []complex128) []complex128 {
	return zipValues(a, b, func(x, y complex128) complex128 { return x - y })
}
func multiplyValues(a, b []complex128) []complex128 {
	return zipValues(a, b, func(x, y complex128) complex128 { return x * y })
}
func scaleValues(a []complex128, scale float64) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = complex(real(a[i])*scale, imag(a[i])*scale)
	}
	return out
}
func addScalar(a []complex128, scalar complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] + scalar
	}
	return out
}
func rotateLeft(values []complex128, amount int) []complex128 {
	out := make([]complex128, len(values))
	for i := range values {
		out[i] = values[(i+amount)%len(values)]
	}
	return out
}
func zipValues(a, b []complex128, f func(complex128, complex128) complex128) []complex128 {
	if len(a) != len(b) {
		return nil
	}
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = f(a[i], b[i])
	}
	return out
}

func compare(reference, actual []complex128) metric {
	result := metric{Finite: len(reference) == len(actual)}
	var sumSquares float64
	count := min(len(reference), len(actual))
	for i := 0; i < count; i++ {
		if !finite(real(reference[i])) || !finite(imag(reference[i])) || !finite(real(actual[i])) || !finite(imag(actual[i])) {
			result.Finite = false
		}
		difference := actual[i] - reference[i]
		magnitude := cmplx.Abs(difference)
		sumSquares += magnitude * magnitude
		result.MaxComplexErr = math.Max(result.MaxComplexErr, magnitude)
		result.MaxRealErr = math.Max(result.MaxRealErr, math.Abs(real(difference)))
		result.MaxImagErr = math.Max(result.MaxImagErr, math.Abs(imag(difference)))
	}
	if count > 0 {
		result.ComplexRMSE = math.Sqrt(sumSquares / float64(count))
	}
	if !finite(sumSquares) || !finite(result.ComplexRMSE) {
		result.Finite = false
	}
	return result
}

func hashInputs(inputs ...[]complex128) string {
	h := sha256.New()
	var encoded [16]byte
	for _, values := range inputs {
		binary.LittleEndian.PutUint64(encoded[:8], uint64(len(values)))
		_, _ = h.Write(encoded[:8])
		for _, value := range values {
			binary.LittleEndian.PutUint64(encoded[:8], math.Float64bits(real(value)))
			binary.LittleEndian.PutUint64(encoded[8:], math.Float64bits(imag(value)))
			_, _ = h.Write(encoded[:])
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashBytes(values []byte) string {
	sum := sha256.Sum256(values)
	return hex.EncodeToString(sum[:])
}

func describeKeyLayout(key any) (string, bool) {
	value := reflect.ValueOf(key)
	if !value.IsValid() {
		return "nil", false
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "nil", false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return fmt.Sprintf("unexpected key type %T", key), false
	}
	layout := value.FieldByName("Layout")
	if !layout.IsValid() {
		// Genuine Standard v6.2.0 predates the explicit Fast/Standard layout field.
		// Its key type is the native Standard representation by construction.
		return "native Standard key type (no Layout field in pinned API)", true
	}
	if layout.Kind() < reflect.Int || layout.Kind() > reflect.Int64 {
		return "unknown Layout field type", false
	}
	if layout.Int() != 0 {
		return fmt.Sprintf("non-Standard layout value %d", layout.Int()), false
	}
	return "KeyLayoutStandard", true
}

func finite(value float64) bool               { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func intPtr(v int) *int                       { return &v }
func floatPtr(v float64) *float64             { return &v }
func stringsContains(value, part string) bool { return bytes.Contains([]byte(value), []byte(part)) }

func writeJSON(path string, value any) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func combineEvidenceFiles(standardPath, fastPath, outPath string) error {
	var standard, fast evidence
	if err := readJSON(standardPath, &standard); err != nil {
		return fmt.Errorf("read Standard evidence: %w", err)
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return fmt.Errorf("read Fast evidence: %w", err)
	}
	if standard.Status != "all_p0_public_primitive_checkpoints_passed" || fast.Status != "all_p0_public_primitive_checkpoints_passed" {
		return fmt.Errorf("cannot combine incomplete backend runs: Standard=%s Fast=%s", standard.Status, fast.Status)
	}
	if standard.FastCapability || !fast.FastCapability || standard.BackendCommit != standardCommit || fast.BackendCommit != fastCommit {
		return fmt.Errorf("backend identity or immutable commit pin does not match the charter")
	}
	if standard.BackendMode != "genuine-standard-encryption" || fast.BackendMode != "fast-p0-zero-secret-simulation" {
		return fmt.Errorf("backend lifecycle classification does not match the charter")
	}
	if standard.FrontendSHA256 == "" || standard.FrontendSHA256 != fast.FrontendSHA256 || standard.TestSourceSHA256 == "" || standard.TestSourceSHA256 != fast.TestSourceSHA256 || standard.ProfileSHA256 != fast.ProfileSHA256 || standard.ParameterSHA256 != fast.ParameterSHA256 || standard.InputSHA256 != fast.InputSHA256 {
		return fmt.Errorf("Standard/Fast source, parameter, profile, or input provenance differs")
	}
	if standard.BackendDirty || fast.BackendDirty {
		return fmt.Errorf("a pinned Lattigo worktree was dirty during the run")
	}
	if standard.TotalKeyLookups.Relinearization == 0 || standard.TotalKeyLookups.Galois == 0 || fast.TotalKeyLookups.Relinearization != 0 || fast.TotalKeyLookups.Galois != 0 {
		return fmt.Errorf("native Standard vs keyless Fast lookup evidence is inconsistent")
	}
	if standard.GoVersion != fast.GoVersion || standard.OS != fast.OS || standard.Architecture != fast.Architecture || standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryDirty != fast.PrimaryDirty {
		return fmt.Errorf("Standard/Fast run environment or Primary provenance differs")
	}
	if standard.BootstrapCalls != 0 || fast.BootstrapCalls != 0 {
		return fmt.Errorf("Bootstrap call budget exceeded")
	}
	if len(standard.Operations) != len(fast.Operations) {
		return fmt.Errorf("backend checkpoint count differs: %d vs %d", len(standard.Operations), len(fast.Operations))
	}
	paired := make([]pairedOperation, 0, len(standard.Operations))
	for i := range standard.Operations {
		s, f := standard.Operations[i], fast.Operations[i]
		if s.ID != f.ID {
			return fmt.Errorf("checkpoint order differs at %d: %s vs %s", i, s.ID, f.ID)
		}
		pair := pairedOperation{ID: s.ID, StandardStatus: s.Status, FastStatus: f.Status, FastPath: f.Path, FastKeyLookups: f.KeyLookupDelta}
		if s.Oracle != nil {
			pair.StandardRMSE = floatPtr(s.Oracle.ComplexRMSE)
			pair.StandardMaxError = floatPtr(s.Oracle.MaxComplexErr)
		}
		if f.Oracle != nil {
			pair.FastRMSE = floatPtr(f.Oracle.ComplexRMSE)
			pair.FastMaxError = floatPtr(f.Oracle.MaxComplexErr)
		}
		paired = append(paired, pair)
	}
	combined := combinedEvidence{
		SchemaVersion: "fast-dropin-public-primitives-batch-013-combined.v1", Status: "paired_source_identical_p0_coverage_complete",
		CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano), GoVersion: standard.GoVersion, OS: standard.OS, Architecture: standard.Architecture,
		PrimaryCommit: standard.PrimaryCommit, PrimaryDirty: standard.PrimaryDirty,
		FrontendSHA256: standard.FrontendSHA256, TestSourceSHA256: standard.TestSourceSHA256, Profile: standard.Profile, ProfileSHA256: standard.ProfileSHA256,
		ParameterSHA: standard.ParameterSHA256, QPrimes: standard.QPrimes, PPrimes: standard.PPrimes,
		InputSHA256: standard.InputSHA256, BootstrapCalls: 0,
		Standard: backendRun{Commit: standard.BackendCommit, Ref: standard.BackendRef, Dirty: standard.BackendDirty, Mode: standard.BackendMode, C1Policy: standard.C1Policy, RelinKeyLayout: standard.RelinKeyLayout, GaloisKeyLayout: standard.GaloisKeyLayout, KeyGeneration: standard.KeyGeneration, TotalLookups: standard.TotalKeyLookups, Inputs: standard.Inputs, Operations: standard.Operations},
		Fast:     backendRun{Commit: fast.BackendCommit, Ref: fast.BackendRef, Dirty: fast.BackendDirty, Mode: fast.BackendMode, C1Policy: fast.C1Policy, RelinKeyLayout: fast.RelinKeyLayout, GaloisKeyLayout: fast.GaloisKeyLayout, KeyGeneration: fast.KeyGeneration, TotalLookups: fast.TotalKeyLookups, Inputs: fast.Inputs, Operations: fast.Operations},
		Paired:   paired,
	}
	return writeJSON(outPath, combined)
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
