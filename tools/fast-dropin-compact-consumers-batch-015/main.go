package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	logN            = 13
	logSlots        = 4
	defaultScaleLog = 45
	standardCommit  = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
)

var levels = []int{1, 3}

type profile struct {
	LogN      int   `json:"log_n"`
	LogQ      []int `json:"log_q_bits"`
	LogP      []int `json:"log_p_bits"`
	ScaleLog2 int   `json:"default_scale_log2"`
	LogSlots  int   `json:"log_slots"`
	Levels    []int `json:"levels"`
}

type metric struct {
	Finite      bool    `json:"finite"`
	ComplexRMSE float64 `json:"complex_rmse"`
	MaxComplex  float64 `json:"max_complex_error"`
}

type complexValue struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type keyLookups struct {
	Relinearization int `json:"relinearization"`
	Galois          int `json:"galois"`
	GaloisList      int `json:"galois_key_list"`
}

type outputState struct {
	Level            int     `json:"level"`
	Degree           int     `json:"degree"`
	ScaleLog2        float64 `json:"scale_log2"`
	IsNTT            bool    `json:"is_ntt"`
	IsMontgomery     bool    `json:"is_montgomery"`
	C1Zero           bool    `json:"c1_zero"`
	RowsPerComponent []int   `json:"rows_per_component"`
}

type checkpoint struct {
	ID        string         `json:"id"`
	API       string         `json:"api"`
	State     outputState    `json:"state"`
	Plaintext metric         `json:"plaintext_oracle"`
	Decoded   []complexValue `json:"decoded,omitempty"`
	Status    string         `json:"status"`
}

type runEvidence struct {
	SchemaVersion  string       `json:"schema_version"`
	Status         string       `json:"status"`
	CreatedUTC     string       `json:"created_utc"`
	GoVersion      string       `json:"go_version"`
	OS             string       `json:"os"`
	Architecture   string       `json:"architecture"`
	PrimaryCommit  string       `json:"primary_commit"`
	PrimaryDirty   bool         `json:"primary_dirty"`
	BackendCommit  string       `json:"backend_commit"`
	BackendRef     string       `json:"backend_ref"`
	BackendDirty   bool         `json:"backend_dirty"`
	FastCapability bool         `json:"fast_zero_secret_capability"`
	KeyLookups     keyLookups   `json:"key_lookups"`
	Profile        profile      `json:"profile"`
	ProfileSHA256  string       `json:"profile_sha256"`
	QPSHA256       string       `json:"effective_qp_sha256"`
	Q              []uint64     `json:"q_primes"`
	P              []uint64     `json:"p_primes"`
	SourceSHA256   string       `json:"shared_runner_sha256"`
	InputSHA256    string       `json:"deterministic_input_sha256"`
	Lifecycle      string       `json:"lifecycle"`
	Checkpoints    []checkpoint `json:"checkpoints"`
	BootstrapCalls int          `json:"bootstrap_calls"`
}

type trackedKeySet struct {
	inner rlwe.EvaluationKeySet
	count keyLookups
}

func (keys *trackedKeySet) GetGaloisKey(galEl uint64) (*rlwe.GaloisKey, error) {
	keys.count.Galois++
	return keys.inner.GetGaloisKey(galEl)
}

func (keys *trackedKeySet) GetGaloisKeysList() []uint64 {
	keys.count.GaloisList++
	return keys.inner.GetGaloisKeysList()
}

func (keys *trackedKeySet) GetRelinearizationKey() (*rlwe.RelinearizationKey, error) {
	keys.count.Relinearization++
	return keys.inner.GetRelinearizationKey()
}

type backendSummary struct {
	Commit         string       `json:"commit"`
	Ref            string       `json:"ref"`
	Dirty          bool         `json:"dirty"`
	FastCapability bool         `json:"fast_zero_secret_capability"`
	KeyLookups     keyLookups   `json:"key_lookups"`
	Checkpoints    []checkpoint `json:"checkpoints"`
}

type pairedCheckpoint struct {
	ID                 string  `json:"id"`
	StandardRMSE       float64 `json:"standard_plaintext_rmse"`
	FastRMSE           float64 `json:"fast_plaintext_rmse"`
	FastVsStandardRMSE float64 `json:"fast_vs_standard_rmse"`
	FastVsStandardMax  float64 `json:"fast_vs_standard_max_complex_error"`
	Level              int     `json:"level"`
	StandardScaleLog2  float64 `json:"standard_scale_log2"`
	FastScaleLog2      float64 `json:"fast_scale_log2"`
}

type combinedEvidence struct {
	SchemaVersion   string             `json:"schema_version"`
	Status          string             `json:"status"`
	CreatedUTC      string             `json:"created_utc"`
	PrimaryCommit   string             `json:"primary_commit_during_runs"`
	PrimaryDirty    bool               `json:"primary_dirty_during_runs"`
	Profile         profile            `json:"profile"`
	ProfileSHA256   string             `json:"profile_sha256"`
	EffectiveQPSHA  string             `json:"effective_qp_sha256"`
	RunnerSHA256    string             `json:"shared_runner_sha256"`
	InputSHA256     string             `json:"deterministic_input_sha256"`
	BootstrapCalls  int                `json:"bootstrap_calls"`
	Standard        backendSummary     `json:"standard"`
	Fast            backendSummary     `json:"fast"`
	Paired          []pairedCheckpoint `json:"paired_checkpoints"`
	CompactBoundary string             `json:"compact_boundary"`
}

func main() {
	if err := runCLI(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() error {
	out := flag.String("out", "", "output JSON evidence")
	primaryCommit := flag.String("primary-commit", "", "Primary revision during run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty")
	backendCommit := flag.String("backend-commit", "", "Lattigo backend revision")
	backendRef := flag.String("backend-ref", "", "Lattigo backend ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether backend was dirty")
	combineStandard := flag.String("combine-standard", "", "Standard run JSON")
	combineFast := flag.String("combine-fast", "", "Fast run JSON")
	flag.Parse()
	if *combineStandard != "" || *combineFast != "" {
		if *combineStandard == "" || *combineFast == "" || *out == "" {
			return fmt.Errorf("combine mode requires -combine-standard, -combine-fast and -out")
		}
		return combine(*combineStandard, *combineFast, *out)
	}
	if *out == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return fmt.Errorf("-out, -primary-commit, -backend-commit and -backend-ref are required")
	}
	result, runErr := executeComposition()
	result.PrimaryCommit, result.PrimaryDirty = *primaryCommit, *primaryDirty
	result.BackendCommit, result.BackendRef, result.BackendDirty = *backendCommit, *backendRef, *backendDirty
	if source, err := os.ReadFile("tools/fast-dropin-compact-consumers-batch-015/main.go"); err == nil {
		result.SourceSHA256 = hashBytes(source)
	} else if runErr == nil {
		runErr = fmt.Errorf("hash shared runner: %w", err)
	}
	if err := writeJSON(*out, result); err != nil {
		return err
	}
	return runErr
}

func executeComposition() (runEvidence, error) {
	used := profile{LogN: logN, LogQ: []int{55, 39, 40, 39}, LogP: []int{60}, ScaleLog2: defaultScaleLog, LogSlots: logSlots, Levels: append([]int(nil), levels...)}
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: used.LogN, LogQ: used.LogQ, LogP: used.LogP, LogDefaultScale: used.ScaleLog2,
		Xs: ring.Ternary{H: 192},
	})
	if err != nil {
		return runEvidence{}, fmt.Errorf("construct frozen Batch 014 profile: %w", err)
	}
	if params.MaxLevel() != 3 || len(params.Q()) != 4 || len(params.P()) != 1 {
		return runEvidence{}, fmt.Errorf("profile mismatch: maxLevel=%d Q=%d P=%d", params.MaxLevel(), len(params.Q()), len(params.P()))
	}
	_, fast := any(params).(interface{ FastCKKSZeroSecretSimulation() })
	profileJSON, _ := json.Marshal(used)
	qpJSON, _ := json.Marshal(struct{ Q, P []uint64 }{params.Q(), params.P()})
	a, b := deterministicInputs(1 << logSlots)
	keys := &trackedKeySet{inner: rlwe.NewMemEvaluationKeySet(nil)}
	keygen := ckks.NewKeyGenerator(params)
	sk := keygen.GenSecretKeyNew()
	if !fast {
		relin := keygen.GenRelinearizationKeyNew(sk)
		galEl := params.GaloisElementForRotation(1)
		galois := keygen.GenGaloisKeyNew(galEl, sk)
		keys.inner = rlwe.NewMemEvaluationKeySet(relin, galois)
	} else {
		// Keys are deliberately present as a sentinel: Fast compact public
		// MulRelin/Rotate must not retrieve them.
		relin := keygen.GenRelinearizationKeyNew(sk)
		galEl := params.GaloisElementForRotation(1)
		galois := keygen.GenGaloisKeyNew(galEl, sk)
		keys.inner = rlwe.NewMemEvaluationKeySet(relin, galois)
	}
	result := runEvidence{
		SchemaVersion: "fast-dropin-compact-consumers-batch-015-run.v1", Status: "running",
		CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano), GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		FastCapability: fast, Profile: used, ProfileSHA256: hashBytes(profileJSON), QPSHA256: hashBytes(qpJSON),
		Q: params.Q(), P: params.P(), InputSHA256: hashInputs(a, b), BootstrapCalls: 0,
		Lifecycle: "same CKKS public source and profile; NewKeyGenerator/GenSecretKeyNew; NewPlaintext/Encoder.Encode; rlwe.NewEncryptor.EncryptNew; public ckks.NewEvaluator AddNew/MulRelinNew/Rescale/RotateNew; ordinary DecryptNew/Decode",
	}
	evaluator := ckks.NewEvaluator(params, keys)
	encoder := ckks.NewEncoder(params)
	decryptor := rlwe.NewDecryptor(params, sk)
	if evaluator == nil || evaluator.Evaluator == nil {
		return result, fmt.Errorf("public ckks.NewEvaluator was not constructed")
	}
	for _, level := range levels {
		ctA, err := encrypt(params, encoder, sk, a, level)
		if err != nil {
			return result, fmt.Errorf("encrypt a L%d: %w", level, err)
		}
		ctB, err := encrypt(params, encoder, sk, b, level)
		if err != nil {
			return result, fmt.Errorf("encrypt b L%d: %w", level, err)
		}
		if fast && (!summarize(ctA).C1Zero || !summarize(ctB).C1Zero) {
			return result, fmt.Errorf("Fast EncryptNew did not preserve the zero-c1 contract at Level %d", level)
		}
		wantAdd := zip(a, b, func(x, y complex128) complex128 { return x + y })
		wantMul := zip(wantAdd, b, func(x, y complex128) complex128 { return x * y })
		sum, err := evaluator.AddNew(ctA, ctB)
		if err != nil {
			return result, fmt.Errorf("AddNew L%d: %w", level, err)
		}
		if err := record(&result, encoder, decryptor, fmt.Sprintf("add-l%d", level), "AddNew", sum, wantAdd, level, 1, ctA.Scale, 1e-6, fast); err != nil {
			return result, err
		}
		product, err := evaluator.MulRelinNew(sum, ctB)
		if err != nil {
			return result, fmt.Errorf("MulRelinNew L%d: %w", level, err)
		}
		productScale := sum.Scale.Mul(ctB.Scale)
		if err := record(&result, encoder, decryptor, fmt.Sprintf("mul-relin-l%d", level), "MulRelinNew", product, wantMul, level, 1, productScale, 1e-4, fast); err != nil {
			return result, err
		}
		rescaled := ckks.NewCiphertext(params, 1, level-1)
		if err := evaluator.Rescale(product, rescaled); err != nil {
			return result, fmt.Errorf("Rescale L%d: %w", level, err)
		}
		rescaleScale := productScale.Div(rlwe.NewScale(params.Q()[level]))
		if err := record(&result, encoder, decryptor, fmt.Sprintf("rescale-l%d", level), "Rescale", rescaled, wantMul, level-1, 1, rescaleScale, 1e-4, fast); err != nil {
			return result, err
		}
		rotated, err := evaluator.RotateNew(rescaled, 1)
		if err != nil {
			return result, fmt.Errorf("RotateNew L%d: %w", level-1, err)
		}
		wantRotated := rotate(wantMul, 1)
		if err := record(&result, encoder, decryptor, fmt.Sprintf("rotate-l%d", level-1), "RotateNew", rotated, wantRotated, level-1, 1, rescaleScale, 1e-4, fast); err != nil {
			return result, err
		}
	}
	result.KeyLookups = keys.count
	result.Status = "public_compact_add_mul_relin_rescale_rotate_passed"
	return result, nil
}

func encrypt(params ckks.Parameters, encoder *ckks.Encoder, sk *rlwe.SecretKey, values []complex128, level int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, level)
	pt.Scale = rlwe.NewScale(math.Exp2(defaultScaleLog))
	pt.IsNTT = true
	pt.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(pt)
}

func record(result *runEvidence, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, id, api string, ct *rlwe.Ciphertext, want []complex128, level, degree int, scale rlwe.Scale, tolerance float64, fast bool) error {
	decoded := make([]complex128, len(want))
	if err := encoder.Decode(decryptor.DecryptNew(ct), decoded); err != nil {
		return fmt.Errorf("%s DecryptNew/Decode: %w", id, err)
	}
	measurement := compare(want, decoded)
	state := summarize(ct)
	status := "passed"
	if !measurement.Finite || measurement.MaxComplex > tolerance || state.Level != level || state.Degree != degree || ct.Scale.Cmp(scale) != 0 || !state.IsNTT || state.IsMontgomery {
		status = "failed_oracle_or_state"
	}
	if fast && !state.C1Zero {
		status = "failed_zero_c1_contract"
	}
	wireValues := make([]complexValue, len(decoded))
	for i, value := range decoded {
		wireValues[i] = complexValue{Real: real(value), Imag: imag(value)}
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{ID: id, API: api, State: state, Plaintext: measurement, Decoded: wireValues, Status: status})
	if status != "passed" {
		return fmt.Errorf("%s failed: status=%s Level=%d Degree=%d ScaleLog2=%.9g RMSE=%.9g max=%.9g tolerance=%.9g", id, status, state.Level, state.Degree, state.ScaleLog2, measurement.ComplexRMSE, measurement.MaxComplex, tolerance)
	}
	return nil
}

func summarize(ct *rlwe.Ciphertext) outputState {
	state := outputState{Level: ct.Level(), Degree: ct.Degree(), ScaleLog2: ct.Scale.Log2(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, C1Zero: true}
	for _, poly := range ct.Value {
		rows := 0
		for row := 0; row <= ct.Level() && row < len(poly.Coeffs); row++ {
			if len(poly.Coeffs[row]) > 0 {
				rows++
			}
		}
		state.RowsPerComponent = append(state.RowsPerComponent, rows)
	}
	if len(ct.Value) > 1 {
		for row := 0; row <= ct.Level() && row < len(ct.Value[1].Coeffs); row++ {
			for _, value := range ct.Value[1].Coeffs[row] {
				if value != 0 {
					state.C1Zero = false
					break
				}
			}
		}
	}
	return state
}

func combine(standardPath, fastPath, outPath string) error {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return err
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return err
	}
	if standard.FastCapability || !fast.FastCapability || standard.BackendCommit != standardCommit || standard.BackendDirty || fast.BackendDirty {
		return fmt.Errorf("backend provenance does not identify clean pinned Standard and clean Fast")
	}
	if standard.Status != "public_compact_add_mul_relin_rescale_rotate_passed" || fast.Status != standard.Status {
		return fmt.Errorf("a backend failed the composed public chain: Standard=%s Fast=%s", standard.Status, fast.Status)
	}
	if standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryDirty != fast.PrimaryDirty || standard.SourceSHA256 != fast.SourceSHA256 || standard.ProfileSHA256 != fast.ProfileSHA256 || standard.QPSHA256 != fast.QPSHA256 || standard.InputSHA256 != fast.InputSHA256 {
		return fmt.Errorf("paired runs differ in Primary state, runner, profile, effective Q/P, or deterministic inputs")
	}
	if len(standard.Checkpoints) != len(fast.Checkpoints) || len(standard.Checkpoints) != 8 {
		return fmt.Errorf("expected 8 matched checkpoints, Standard=%d Fast=%d", len(standard.Checkpoints), len(fast.Checkpoints))
	}
	combined := combinedEvidence{
		SchemaVersion: "fast-dropin-compact-consumers-batch-015-evidence.v1", Status: "paired_public_compact_chain_passed",
		CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano), PrimaryCommit: standard.PrimaryCommit, PrimaryDirty: standard.PrimaryDirty,
		Profile: standard.Profile, ProfileSHA256: standard.ProfileSHA256, EffectiveQPSHA: standard.QPSHA256,
		RunnerSHA256: standard.SourceSHA256, InputSHA256: standard.InputSHA256, BootstrapCalls: 0,
		CompactBoundary: "Numerical decoding is limited to accepted LogN13 Levels 1 and 3, where the maintained Q-prefix includes every active row. The separate Level-5 Secondary test is structural-only: q4/q5 remain dormant through Add/MulRelin/Rescale/Rotate. No Bootstrap, benchmark, or LogN16 run was made.",
	}
	combined.Standard = summaryOf(standard)
	combined.Fast = summaryOf(fast)
	for i := range standard.Checkpoints {
		sc, fc := standard.Checkpoints[i], fast.Checkpoints[i]
		if sc.ID != fc.ID || sc.Status != "passed" || fc.Status != "passed" || sc.State.Level != fc.State.Level || sc.State.Degree != fc.State.Degree || math.Abs(sc.State.ScaleLog2-fc.State.ScaleLog2) > 1e-9 {
			return fmt.Errorf("paired checkpoint mismatch at %d: Standard=%s/%d Fast=%s/%d", i, sc.ID, sc.State.Level, fc.ID, fc.State.Level)
		}
		paired := compare(valuesFromEvidence(sc.Decoded), valuesFromEvidence(fc.Decoded))
		if !paired.Finite {
			return fmt.Errorf("non-finite Fast-vs-Standard result at %s", sc.ID)
		}
		combined.Paired = append(combined.Paired, pairedCheckpoint{ID: sc.ID, StandardRMSE: sc.Plaintext.ComplexRMSE, FastRMSE: fc.Plaintext.ComplexRMSE, FastVsStandardRMSE: paired.ComplexRMSE, FastVsStandardMax: paired.MaxComplex, Level: sc.State.Level, StandardScaleLog2: sc.State.ScaleLog2, FastScaleLog2: fc.State.ScaleLog2})
	}
	if fast.KeyLookups.Relinearization != 0 || fast.KeyLookups.Galois != 0 {
		return fmt.Errorf("Fast public chain unexpectedly looked up evaluation keys: %+v", fast.KeyLookups)
	}
	return writeJSON(outPath, combined)
}

func summaryOf(result runEvidence) backendSummary {
	for i := range result.Checkpoints {
		result.Checkpoints[i].Decoded = nil
	}
	return backendSummary{Commit: result.BackendCommit, Ref: result.BackendRef, Dirty: result.BackendDirty, FastCapability: result.FastCapability, KeyLookups: result.KeyLookups, Checkpoints: result.Checkpoints}
}

func deterministicInputs(slots int) ([]complex128, []complex128) {
	a, b := make([]complex128, slots), make([]complex128, slots)
	for i := range a {
		a[i] = complex(float64(i%7-3)/32, float64((2*i)%5-2)/64)
		b[i] = complex(float64(i%5-2)/16, float64(i%3-1)/32)
	}
	return a, b
}

func rotate(values []complex128, k int) []complex128 {
	out := make([]complex128, len(values))
	shift := k % len(values)
	for i := range out {
		out[i] = values[(i+shift)%len(values)]
	}
	return out
}

func zip(a, b []complex128, operation func(complex128, complex128) complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = operation(a[i], b[i])
	}
	return out
}

func compare(want, got []complex128) metric {
	out := metric{Finite: len(want) == len(got)}
	var squares float64
	for i := range want {
		if !finite(real(got[i])) || !finite(imag(got[i])) {
			out.Finite = false
		}
		error := cmplx.Abs(got[i] - want[i])
		squares += error * error
		out.MaxComplex = math.Max(out.MaxComplex, error)
	}
	if len(want) > 0 {
		out.ComplexRMSE = math.Sqrt(squares / float64(len(want)))
	}
	if !finite(out.ComplexRMSE) || !finite(out.MaxComplex) {
		out.Finite = false
	}
	return out
}

func hashInputs(inputs ...[]complex128) string {
	h := sha256.New()
	var buf [16]byte
	for _, values := range inputs {
		binary.LittleEndian.PutUint64(buf[:8], uint64(len(values)))
		_, _ = h.Write(buf[:8])
		for _, value := range values {
			binary.LittleEndian.PutUint64(buf[:8], math.Float64bits(real(value)))
			binary.LittleEndian.PutUint64(buf[8:], math.Float64bits(imag(value)))
			_, _ = h.Write(buf[:])
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func valuesFromEvidence(values []complexValue) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
