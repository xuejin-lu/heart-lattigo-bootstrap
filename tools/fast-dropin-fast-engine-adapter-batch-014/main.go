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
	logN             = 13
	logSlots         = 4
	defaultScaleLog2 = 45
	standardCommit   = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	catastrophicMax  = 1.0
)

var levels = []int{1, 3}

type profile struct {
	LogN       int   `json:"log_n"`
	LogQ       []int `json:"log_q_bits"`
	LogP       []int `json:"log_p_bits"`
	ScaleLog2  int   `json:"default_scale_log2"`
	LogSlots   int   `json:"log_slots"`
	TestLevels []int `json:"test_levels"`
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

type inputEvidence struct {
	Name   string `json:"name"`
	State  state  `json:"state"`
	Oracle metric `json:"plaintext_oracle"`
}

type state struct {
	Level                     int     `json:"level"`
	Degree                    int     `json:"degree"`
	ScaleLog2                 float64 `json:"scale_log2"`
	IsNTT                     bool    `json:"is_ntt"`
	IsMontgomery              bool    `json:"is_montgomery"`
	C1Zero                    bool    `json:"c1_zero"`
	RowsPerComponent          []int   `json:"rows_per_component"`
	AllActiveRowsMaterialized bool    `json:"all_active_q_rows_materialized"`
}

type checkpoint struct {
	ID        string         `json:"id"`
	API       string         `json:"api"`
	Level     int            `json:"level"`
	ScaleLog2 float64        `json:"scale_log2"`
	Degree    int            `json:"degree"`
	State     state          `json:"output_state"`
	Oracle    metric         `json:"plaintext_oracle"`
	Decoded   []complexValue `json:"decoded_values,omitempty"`
	Status    string         `json:"status"`
}

type runEvidence struct {
	SchemaVersion       string          `json:"schema_version"`
	Status              string          `json:"status"`
	CreatedUTC          string          `json:"created_utc"`
	GoVersion           string          `json:"go_version"`
	OS                  string          `json:"os"`
	Architecture        string          `json:"architecture"`
	PrimaryCommit       string          `json:"primary_commit"`
	PrimaryDirty        bool            `json:"primary_dirty"`
	BackendCommit       string          `json:"backend_commit"`
	BackendRef          string          `json:"backend_ref"`
	BackendDirty        bool            `json:"backend_dirty"`
	FastZeroSecret      bool            `json:"fast_zero_secret_capability"`
	BackendMode         string          `json:"backend_mode"`
	PublicEvaluator     string          `json:"public_evaluator"`
	Profile             profile         `json:"profile"`
	ProfileSHA256       string          `json:"profile_sha256"`
	EffectiveQPSHA256   string          `json:"effective_qp_sha256"`
	Q                   []uint64        `json:"q_primes"`
	P                   []uint64        `json:"p_primes"`
	FrontendSHA256      string          `json:"shared_frontend_sha256"`
	TestSHA256          string          `json:"shared_test_source_sha256"`
	InputSHA256         string          `json:"deterministic_input_sha256"`
	EncryptionLifecycle string          `json:"encryption_lifecycle"`
	Inputs              []inputEvidence `json:"encrypted_inputs"`
	Checkpoints         []checkpoint    `json:"checkpoints"`
	BootstrapCalls      int             `json:"bootstrap_calls"`
	StopReason          string          `json:"stop_reason,omitempty"`
}

type pairedCheckpoint struct {
	ID                 string  `json:"id"`
	StandardRMSE       float64 `json:"standard_complex_rmse"`
	FastRMSE           float64 `json:"fast_complex_rmse"`
	FastVsStandardRMSE float64 `json:"fast_vs_standard_complex_rmse"`
	FastVsStandardMax  float64 `json:"fast_vs_standard_max_complex_error"`
	StandardLevel      int     `json:"standard_level"`
	FastLevel          int     `json:"fast_level"`
	StandardScaleLog2  float64 `json:"standard_scale_log2"`
	FastScaleLog2      float64 `json:"fast_scale_log2"`
}

type combinedEvidence struct {
	SchemaVersion   string             `json:"schema_version"`
	CreatedUTC      string             `json:"created_utc"`
	Status          string             `json:"status"`
	PrimaryCommit   string             `json:"primary_commit_during_runs"`
	PrimaryDirty    bool               `json:"primary_dirty_during_runs"`
	FrontendSHA256  string             `json:"identical_frontend_sha256"`
	TestSHA256      string             `json:"identical_test_source_sha256"`
	Profile         profile            `json:"profile"`
	InputSHA256     string             `json:"deterministic_input_sha256"`
	BootstrapCalls  int                `json:"bootstrap_calls"`
	Standard        runEvidence        `json:"standard"`
	Fast            runEvidence        `json:"fast"`
	Paired          []pairedCheckpoint `json:"paired_checkpoints"`
	DispatchProof   string             `json:"dispatch_proof"`
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
	standardPath := flag.String("combine-standard", "", "combine Standard run")
	fastPath := flag.String("combine-fast", "", "combine Fast run")
	flag.Parse()
	if *standardPath != "" || *fastPath != "" {
		if *standardPath == "" || *fastPath == "" || *out == "" {
			return fmt.Errorf("combine mode requires -combine-standard, -combine-fast and -out")
		}
		return combineRuns(*standardPath, *fastPath, *out)
	}
	if *out == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return fmt.Errorf("-out, -primary-commit, -backend-commit and -backend-ref are required")
	}
	run, runErr := runPublicComposition()
	run.PrimaryCommit, run.PrimaryDirty = *primaryCommit, *primaryDirty
	run.BackendCommit, run.BackendRef, run.BackendDirty = *backendCommit, *backendRef, *backendDirty
	if b, err := os.ReadFile("tools/fast-dropin-fast-engine-adapter-batch-014/main.go"); err == nil {
		run.FrontendSHA256 = hashBytes(b)
	} else if runErr == nil {
		runErr = fmt.Errorf("hash runner source: %w", err)
	}
	if b, err := os.ReadFile("tools/fast-dropin-fast-engine-adapter-batch-014/main_test.go"); err == nil {
		run.TestSHA256 = hashBytes(b)
	} else if runErr == nil {
		runErr = fmt.Errorf("hash test source: %w", err)
	}
	if err := writeJSON(*out, run); err != nil {
		return err
	}
	return runErr
}

func runPublicComposition() (runEvidence, error) {
	used := profile{LogN: logN, LogQ: []int{55, 39, 40, 39}, LogP: []int{60}, ScaleLog2: defaultScaleLog2, LogSlots: logSlots, TestLevels: append([]int(nil), levels...)}
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: used.LogN, LogQ: used.LogQ, LogP: used.LogP, LogDefaultScale: used.ScaleLog2, Xs: ring.Ternary{H: 192}})
	if err != nil {
		return runEvidence{}, fmt.Errorf("construct frozen LogN13 profile: %w", err)
	}
	if params.MaxLevel() != 3 || len(params.Q()) != 4 || len(params.P()) != 1 {
		return runEvidence{}, fmt.Errorf("frozen profile mismatch: maxLevel=%d Q=%d P=%d", params.MaxLevel(), len(params.Q()), len(params.P()))
	}
	_, fast := any(params).(interface{ FastCKKSZeroSecretSimulation() })
	profileJSON, _ := json.Marshal(used)
	qpJSON, _ := json.Marshal(struct{ Q, P []uint64 }{params.Q(), params.P()})
	valuesA, valuesB := deterministicInputs(1 << logSlots)
	inputHash := hashInputs(valuesA, valuesB)
	result := runEvidence{
		SchemaVersion: "fast-dropin-fast-engine-adapter-batch-014-run.v1", Status: "running", CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		FastZeroSecret: fast, BackendMode: map[bool]string{true: "Fast zero-secret simulation", false: "genuine Standard CKKS"}[fast],
		PublicEvaluator: "ckks.NewEvaluator(params, nil)", Profile: used, ProfileSHA256: hashBytes(profileJSON), EffectiveQPSHA256: hashBytes(qpJSON),
		Q: params.Q(), P: params.P(), InputSHA256: inputHash,
		EncryptionLifecycle: "ckks.NewKeyGenerator -> GenSecretKeyNew; ckks.NewPlaintext/Encoder.Encode; rlwe.NewEncryptor(params, sk).EncryptNew; public ckks.NewEvaluator AddNew/SubNew; ordinary rlwe.NewDecryptor.DecryptNew/ckks.Encoder.Decode",
		BootstrapCalls:      0,
	}
	encoder := ckks.NewEncoder(params)
	keygen := ckks.NewKeyGenerator(params)
	sk := keygen.GenSecretKeyNew()
	evaluator := ckks.NewEvaluator(params, nil)
	decryptor := rlwe.NewDecryptor(params, sk)
	if evaluator == nil || evaluator.Evaluator == nil {
		return result, fmt.Errorf("public ckks.NewEvaluator was not constructed")
	}
	for _, level := range levels {
		a, err := encrypt(params, encoder, sk, valuesA, level)
		if err != nil {
			return result, fmt.Errorf("encrypt a L%d: %w", level, err)
		}
		b, err := encrypt(params, encoder, sk, valuesB, level)
		if err != nil {
			return result, fmt.Errorf("encrypt b L%d: %w", level, err)
		}
		if err := recordInput(&result, encoder, decryptor, "a-l"+itoa(level), a, valuesA, fast); err != nil {
			return result, err
		}
		if err := recordInput(&result, encoder, decryptor, "b-l"+itoa(level), b, valuesB, fast); err != nil {
			return result, err
		}
		if err := appendCheckpoint(&result, encoder, decryptor, "addnew-l"+itoa(level), "AddNew", evaluator.AddNew, a, b, add(valuesA, valuesB)); err != nil {
			return result, err
		}
		sum, err := evaluator.AddNew(a, b)
		if err != nil {
			return result, fmt.Errorf("composition AddNew L%d: %w", level, err)
		}
		if err := appendCheckpoint(&result, encoder, decryptor, "subnew-l"+itoa(level), "SubNew", evaluator.SubNew, a, b, subtract(valuesA, valuesB)); err != nil {
			return result, err
		}
		recovered, err := evaluator.SubNew(sum, b)
		if err != nil {
			return result, fmt.Errorf("composition SubNew(AddNew(a,b),b) L%d: %w", level, err)
		}
		if err := recordExistingOutput(&result, encoder, decryptor, "add-sub-recovery-l"+itoa(level), "SubNew(AddNew(a,b), b)", recovered, valuesA, level, 1, a.Scale); err != nil {
			return result, err
		}
	}
	result.Status = "public_add_sub_composition_passed"
	return result, nil
}

func encrypt(params ckks.Parameters, encoder *ckks.Encoder, sk *rlwe.SecretKey, values []complex128, level int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, level)
	pt.Scale = rlwe.NewScale(math.Exp2(defaultScaleLog2))
	pt.IsNTT = true
	pt.IsMontgomery = false
	pt.LogDimensions = ring.Dimensions{Rows: 1, Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(pt)
}

func appendCheckpoint(result *runEvidence, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, id, api string, call func(*rlwe.Ciphertext, rlwe.Operand) (*rlwe.Ciphertext, error), a, b *rlwe.Ciphertext, want []complex128) error {
	out, err := call(a, b)
	if err != nil {
		return fmt.Errorf("%s: %w", id, err)
	}
	return recordExistingOutput(result, encoder, decryptor, id, api, out, want, min(a.Level(), b.Level()), max(a.Degree(), b.Degree()), a.Scale.Max(b.Scale))
}

func recordInput(result *runEvidence, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, name string, ct *rlwe.Ciphertext, want []complex128, fast bool) error {
	info := summarize(ct)
	decoded := make([]complex128, len(want))
	if err := encoder.Decode(decryptor.DecryptNew(ct), decoded); err != nil {
		return fmt.Errorf("%s DecryptNew/Decode: %w", name, err)
	}
	oracle := compare(want, decoded)
	if !oracle.Finite || oracle.MaxComplex > catastrophicMax {
		return fmt.Errorf("%s input oracle failed: max error %.12g", name, oracle.MaxComplex)
	}
	if fast && !info.C1Zero {
		return fmt.Errorf("%s Fast ordinary EncryptNew did not preserve zero-c1 policy", name)
	}
	if info.Level != ct.Level() || !info.AllActiveRowsMaterialized {
		return fmt.Errorf("%s ordinary input has incomplete active-Q backing", name)
	}
	result.Inputs = append(result.Inputs, inputEvidence{Name: name, State: info, Oracle: oracle})
	return nil
}

func recordExistingOutput(result *runEvidence, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, id, api string, ct *rlwe.Ciphertext, want []complex128, expectedLevel, expectedDegree int, expectedScale rlwe.Scale) error {
	info := summarize(ct)
	decoded := make([]complex128, len(want))
	if err := encoder.Decode(decryptor.DecryptNew(ct), decoded); err != nil {
		return fmt.Errorf("%s DecryptNew/Decode: %w", id, err)
	}
	measurement := compare(want, decoded)
	status := "passed"
	if !measurement.Finite || measurement.MaxComplex > catastrophicMax {
		status = "failed_oracle"
	}
	if info.Level != expectedLevel || info.Degree != expectedDegree || ct.Scale.Cmp(expectedScale) != 0 || !info.AllActiveRowsMaterialized || !info.IsNTT || info.IsMontgomery {
		status = "failed_state_or_backing"
	}
	if result.FastZeroSecret && !info.C1Zero {
		status = "failed_c1_policy"
	}
	decodedValues := make([]complexValue, len(decoded))
	for i, value := range decoded {
		decodedValues[i] = complexValue{Real: real(value), Imag: imag(value)}
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{ID: id, API: api, Level: info.Level, ScaleLog2: info.ScaleLog2, Degree: info.Degree, State: info, Oracle: measurement, Decoded: decodedValues, Status: status})
	if status != "passed" {
		return fmt.Errorf("%s failed: status=%s maxError=%.12g", id, status, measurement.MaxComplex)
	}
	return nil
}

func summarize(ct *rlwe.Ciphertext) state {
	out := state{Level: ct.Level(), Degree: ct.Degree(), ScaleLog2: ct.Scale.Log2(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, C1Zero: true, AllActiveRowsMaterialized: true}
	for _, poly := range ct.Value {
		rows := 0
		for q := 0; q <= ct.Level() && q < len(poly.Coeffs); q++ {
			if len(poly.Coeffs[q]) > 0 {
				rows++
			}
		}
		out.RowsPerComponent = append(out.RowsPerComponent, rows)
		if rows != ct.Level()+1 {
			out.AllActiveRowsMaterialized = false
		}
	}
	if len(ct.Value) > 1 {
		for q := 0; q <= ct.Level() && q < len(ct.Value[1].Coeffs); q++ {
			for _, coefficient := range ct.Value[1].Coeffs[q] {
				if coefficient != 0 {
					out.C1Zero = false
					break
				}
			}
		}
	}
	return out
}

func combineRuns(standardPath, fastPath, outPath string) error {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return err
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return err
	}
	if standard.FastZeroSecret || !fast.FastZeroSecret {
		return fmt.Errorf("expected genuine Standard and Fast zero-secret capability runs")
	}
	if standard.FrontendSHA256 != fast.FrontendSHA256 || standard.TestSHA256 != fast.TestSHA256 || standard.ProfileSHA256 != fast.ProfileSHA256 || standard.InputSHA256 != fast.InputSHA256 || standard.EffectiveQPSHA256 != fast.EffectiveQPSHA256 {
		return fmt.Errorf("backend runs differ in frontend, profile, effective Q/P, or original inputs")
	}
	if standard.BackendCommit != standardCommit || standard.BackendDirty || fast.BackendDirty {
		return fmt.Errorf("backend provenance is not clean at the pinned Standard / clean Fast states")
	}
	if standard.Status != "public_add_sub_composition_passed" || fast.Status != "public_add_sub_composition_passed" || len(standard.Inputs) != 4 || len(fast.Inputs) != 4 {
		return fmt.Errorf("one backend did not pass the expected ordinary encryption/Add/Sub checks")
	}
	combined := combinedEvidence{SchemaVersion: "fast-dropin-fast-engine-adapter-batch-014-combined.v1", CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano), Status: "paired_public_add_sub_composition_passed", PrimaryCommit: standard.PrimaryCommit, PrimaryDirty: standard.PrimaryDirty, FrontendSHA256: standard.FrontendSHA256, TestSHA256: standard.TestSHA256, Profile: standard.Profile, InputSHA256: standard.InputSHA256, BootstrapCalls: standard.BootstrapCalls + fast.BootstrapCalls, Standard: standard, Fast: fast,
		DispatchProof:   "Focused Secondary TestPublicFastZeroSecretAddSubUsesSharedQPrefixCoreAndDecodes instruments the shared AddSubCore and verifies public AddNew/SubNew hit it at Levels 1 and 3; the Level-5 structural test verifies the four-row capped path. The independent Standard build does not implement the Fast capability marker.",
		CompactBoundary: "At the frozen LogN13 profile, max logical Level is 3 and all active Q rows fit in the four-row prefix. The Level-5 test proves Add/Sub compaction structurally; public MulRelin, Rescale, and Rotate fail closed at the first missing active row q4 rather than reading dormant rows. No >cap decoded oracle is claimed.",
	}
	if len(standard.Checkpoints) != len(fast.Checkpoints) {
		return fmt.Errorf("checkpoint count differs: Standard=%d Fast=%d", len(standard.Checkpoints), len(fast.Checkpoints))
	}
	for i := range standard.Checkpoints {
		sc, fc := standard.Checkpoints[i], fast.Checkpoints[i]
		if sc.ID != fc.ID || sc.Level != fc.Level || sc.Degree != fc.Degree || math.Abs(sc.ScaleLog2-fc.ScaleLog2) > 1e-9 {
			return fmt.Errorf("checkpoint state mismatch at index %d: Standard=%+v Fast=%+v", i, sc, fc)
		}
		if sc.Status != "passed" || fc.Status != "passed" {
			return fmt.Errorf("checkpoint %s did not pass: Standard=%s Fast=%s", sc.ID, sc.Status, fc.Status)
		}
		paired := compare(valuesFromEvidence(sc.Decoded), valuesFromEvidence(fc.Decoded))
		if !paired.Finite {
			return fmt.Errorf("checkpoint %s has non-finite paired comparison", sc.ID)
		}
		combined.Paired = append(combined.Paired, pairedCheckpoint{ID: sc.ID, StandardRMSE: sc.Oracle.ComplexRMSE, FastRMSE: fc.Oracle.ComplexRMSE, FastVsStandardRMSE: paired.ComplexRMSE, FastVsStandardMax: paired.MaxComplex, StandardLevel: sc.Level, FastLevel: fc.Level, StandardScaleLog2: sc.ScaleLog2, FastScaleLog2: fc.ScaleLog2})
	}
	for i := range combined.Standard.Checkpoints {
		combined.Standard.Checkpoints[i].Decoded = nil
	}
	for i := range combined.Fast.Checkpoints {
		combined.Fast.Checkpoints[i].Decoded = nil
	}
	return writeJSON(outPath, combined)
}

func valuesFromEvidence(values []complexValue) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func deterministicInputs(slots int) ([]complex128, []complex128) {
	a, b := make([]complex128, slots), make([]complex128, slots)
	for i := range a {
		a[i] = complex(float64(i%7-3)/32, float64((2*i)%5-2)/64)
		b[i] = complex(float64(i%5-2)/16, float64(i%3-1)/32)
	}
	return a, b
}

func add(a, b []complex128) []complex128 {
	return zip(a, b, func(x, y complex128) complex128 { return x + y })
}
func subtract(a, b []complex128) []complex128 {
	return zip(a, b, func(x, y complex128) complex128 { return x - y })
}
func zip(a, b []complex128, f func(complex128, complex128) complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = f(a[i], b[i])
	}
	return out
}

func compare(want, got []complex128) metric {
	out := metric{Finite: len(want) == len(got)}
	var sum float64
	for i := range want {
		if !finite(real(got[i])) || !finite(imag(got[i])) {
			out.Finite = false
		}
		d := cmplx.Abs(got[i] - want[i])
		sum += d * d
		out.MaxComplex = math.Max(out.MaxComplex, d)
	}
	if len(want) > 0 {
		out.ComplexRMSE = math.Sqrt(sum / float64(len(want)))
	}
	if !finite(out.ComplexRMSE) || !finite(out.MaxComplex) {
		out.Finite = false
	}
	return out
}

func hashInputs(inputs ...[]complex128) string {
	h := sha256.New()
	var b [16]byte
	for _, values := range inputs {
		binary.LittleEndian.PutUint64(b[:8], uint64(len(values)))
		_, _ = h.Write(b[:8])
		for _, v := range values {
			binary.LittleEndian.PutUint64(b[:8], math.Float64bits(real(v)))
			binary.LittleEndian.PutUint64(b[8:], math.Float64bits(imag(v)))
			_, _ = h.Write(b[:])
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
func hashBytes(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func finite(x float64) bool     { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func itoa(i int) string         { return fmt.Sprintf("%d", i) }
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
