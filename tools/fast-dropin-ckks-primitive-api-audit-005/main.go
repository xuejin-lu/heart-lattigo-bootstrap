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
	logN       = 13
	logSlots   = 4
	inputLevel = 3
	logScale   = 45
	rotation   = 1
	// This is only a catastrophic-stop guard, not an acceptance threshold.
	catastrophicOracleError = 1.0
)

type profile struct {
	LogN          int   `json:"log_n"`
	LogQ          []int `json:"log_q"`
	LogP          []int `json:"log_p"`
	LogDefault    int   `json:"log_default_scale"`
	LogSlots      int   `json:"log_slots"`
	InputLevel    int   `json:"input_level"`
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
	Level          int     `json:"level"`
	ScaleLog2      float64 `json:"scale_log2"`
	Degree         int     `json:"degree"`
	IsNTT          bool    `json:"is_ntt"`
	IsMontgomery   bool    `json:"is_montgomery"`
	C1Zero         bool    `json:"c1_zero"`
	C1Nonzero      int     `json:"c1_nonzero_coefficients"`
	C1Coefficients int     `json:"c1_coefficient_count"`
}

type stageEvidence struct {
	Operation string          `json:"operation"`
	Status    string          `json:"status"`
	Input     *ciphertextInfo `json:"input,omitempty"`
	Output    *ciphertextInfo `json:"output,omitempty"`
	Oracle    *metric         `json:"decoded_vs_plaintext_oracle,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type evidence struct {
	SchemaVersion             string           `json:"schema_version"`
	Status                    string           `json:"status"`
	TimestampUTC              string           `json:"timestamp_utc"`
	GoVersion                 string           `json:"go_version"`
	OS                        string           `json:"os"`
	Architecture              string           `json:"architecture"`
	PrimaryCommit             string           `json:"primary_commit"`
	PrimaryDirty              bool             `json:"primary_dirty"`
	BackendCommit             string           `json:"backend_commit"`
	BackendRef                string           `json:"backend_ref"`
	BackendDirty              bool             `json:"backend_dirty"`
	FrontendSHA256            string           `json:"frontend_sha256"`
	Profile                   profile          `json:"profile"`
	ProfileSHA256             string           `json:"profile_sha256"`
	QPrimes                   []uint64         `json:"q_primes"`
	PPrimes                   []uint64         `json:"p_primes"`
	InputSHA256               string           `json:"input_sha256"`
	InputCiphertexts          []ciphertextInfo `json:"input_ciphertexts"`
	EvaluatorConstructor      string           `json:"evaluator_constructor"`
	UnderlyingEvaluator       string           `json:"underlying_evaluator"`
	FastEvaluatorConstructed  bool             `json:"fast_evaluator_constructed"`
	EvaluationKeysGeneratedBy string           `json:"evaluation_keys_generated_by"`
	FastKeyGeneratorUsed      bool             `json:"fast_key_generator_used"`
	BootstrapCalls            int              `json:"bootstrap_calls"`
	Stages                    []stageEvidence  `json:"stages"`
	StopReason                string           `json:"stop_reason,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	outPath := flag.String("out", "", "local JSON evidence output")
	primaryCommit := flag.String("primary-commit", "", "Primary commit used for this run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether the Primary worktree was dirty")
	backendCommit := flag.String("backend-commit", "", "Lattigo implementation commit")
	backendRef := flag.String("backend-ref", "", "Lattigo branch or pinned ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether the backend worktree was dirty")
	flag.Parse()
	if *outPath == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return fmt.Errorf("-out, -primary-commit, -backend-commit, and -backend-ref are required")
	}

	result, runErr := runAudit(*primaryCommit, *primaryDirty, *backendCommit, *backendRef, *backendDirty)
	if err := writeJSON(*outPath, result); err != nil {
		return err
	}
	return runErr
}

func runAudit(primaryCommit string, primaryDirty bool, backendCommit, backendRef string, backendDirty bool) (evidence, error) {
	profileUsed := profile{
		LogN: logN, LogQ: []int{55, 39, 40, 39}, LogP: []int{60}, LogDefault: logScale,
		LogSlots: logSlots, InputLevel: inputLevel, SecretHamming: 192, Rotation: rotation,
	}
	profileBytes, _ := json.Marshal(profileUsed)
	profileSHA := hashBytes(profileBytes)
	frontendBytes, err := os.ReadFile("tools/fast-dropin-ckks-primitive-api-audit-005/main.go")
	if err != nil {
		return evidence{}, fmt.Errorf("read shared frontend source: %w", err)
	}
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: logN, LogQ: append([]int(nil), profileUsed.LogQ...), LogP: append([]int(nil), profileUsed.LogP...), LogDefaultScale: logScale,
		Xs: ring.Ternary{H: profileUsed.SecretHamming},
	})
	if err != nil {
		return evidence{}, fmt.Errorf("create fixed CKKS parameters: %w", err)
	}
	encoder := ckks.NewEncoder(params)
	valuesA, valuesB := deterministicInputs(1 << logSlots)
	inputHash := hashInputs(valuesA, valuesB)
	keygen := ckks.NewKeyGenerator(params)
	sk := keygen.GenSecretKeyNew()
	encryptor := rlwe.NewEncryptor(params, sk)
	decryptor := rlwe.NewDecryptor(params, sk)
	relinKey := keygen.GenRelinearizationKeyNew(sk)
	galEl := params.GaloisElementForRotation(rotation)
	galoisKey := keygen.GenGaloisKeyNew(galEl, sk)
	keySet := rlwe.NewMemEvaluationKeySet(relinKey, galoisKey)
	eval := ckks.NewEvaluator(params, keySet)
	if eval == nil || eval.Evaluator == nil {
		return evidence{}, fmt.Errorf("public ckks.NewEvaluator did not construct its embedded RLWE evaluator")
	}

	result := evidence{
		SchemaVersion: "fast-dropin-ckks-primitive-api-audit-005.v1",
		Status:        "running", TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		PrimaryCommit: primaryCommit, PrimaryDirty: primaryDirty,
		BackendCommit: backendCommit, BackendRef: backendRef, BackendDirty: backendDirty,
		FrontendSHA256: hashBytes(frontendBytes), Profile: profileUsed, ProfileSHA256: profileSHA,
		QPrimes: params.Q(), PPrimes: params.P(), InputSHA256: inputHash,
		EvaluatorConstructor:      "ckks.NewEvaluator(params, evk)",
		UnderlyingEvaluator:       "*rlwe.Evaluator (constructed by ckks.NewEvaluator in both dependency trees)",
		FastEvaluatorConstructed:  false,
		EvaluationKeysGeneratedBy: "ckks.NewKeyGenerator (ordinary RLWE key layout; no Fast key API)",
		FastKeyGeneratorUsed:      false,
		BootstrapCalls:            0,
	}

	ctA, err := encryptValues(params, encoder, encryptor, valuesA, inputLevel, logScale)
	if err != nil {
		return result, fmt.Errorf("encrypt Add operand A: %w", err)
	}
	ctB, err := encryptValues(params, encoder, encryptor, valuesB, inputLevel, logScale)
	if err != nil {
		return result, fmt.Errorf("encrypt Add operand B: %w", err)
	}
	result.InputCiphertexts = append(result.InputCiphertexts, ciphertextSummary(ctA), ciphertextSummary(ctB))
	zeroSecretSimulation := allZero(ctA.Value[1].Coeffs) && allZero(ctB.Value[1].Coeffs)
	if stop := appendStage(&result, "EncryptNew/Add operand A", nil, ctA, valuesA, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}
	if stop := appendStage(&result, "EncryptNew/Add operand B", nil, ctB, valuesB, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}

	addOut, err := eval.AddNew(ctA, ctB)
	if err != nil {
		return failStage(result, "Add", err)
	}
	if stop := appendStage(&result, "Add", nil, addOut, addValues(valuesA, valuesB), encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}

	// Use fresh public EncryptNew ciphertexts for the independent multiplication inputs.
	ctMulA, err := encryptValues(params, encoder, encryptor, valuesA, inputLevel, logScale)
	if err != nil {
		return result, fmt.Errorf("encrypt MulRelin operand A: %w", err)
	}
	ctMulB, err := encryptValues(params, encoder, encryptor, valuesB, inputLevel, logScale)
	if err != nil {
		return result, fmt.Errorf("encrypt MulRelin operand B: %w", err)
	}
	result.InputCiphertexts = append(result.InputCiphertexts, ciphertextSummary(ctMulA), ciphertextSummary(ctMulB))
	if stop := appendStage(&result, "EncryptNew/MulRelin operand A", nil, ctMulA, valuesA, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}
	if stop := appendStage(&result, "EncryptNew/MulRelin operand B", nil, ctMulB, valuesB, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}
	product, err := eval.MulRelinNew(ctMulA, ctMulB)
	if err != nil {
		return failStage(result, "MulRelin", err)
	}
	productOracle := multiplyValues(valuesA, valuesB)
	if stop := appendStage(&result, "MulRelin", nil, product, productOracle, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}

	rescaled := ckks.NewCiphertext(params, product.Degree(), product.Level()-1)
	if err = eval.Rescale(product, rescaled); err != nil {
		return failStage(result, "Rescale", err)
	}
	if stop := appendStage(&result, "Rescale", product, rescaled, productOracle, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}

	rotated, err := eval.RotateNew(rescaled, rotation)
	if err != nil {
		return failStage(result, "Rotate", err)
	}
	rotationOracle := rotateLeft(productOracle, rotation)
	if stop := appendStage(&result, "Rotate", rescaled, rotated, rotationOracle, encoder, decryptor, zeroSecretSimulation); stop != nil {
		return stoppedResult(stop)
	}
	result.Status = "all_operations_passed_through_ckks_new_evaluator"
	return result, nil
}

func encryptValues(params ckks.Parameters, encoder *ckks.Encoder, encryptor *rlwe.Encryptor, values []complex128, level int, scaleLog2 int) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, level)
	pt.Scale = rlwe.NewScale(math.Exp2(float64(scaleLog2)))
	pt.IsNTT = true
	pt.IsMontgomery = false
	pt.LogDimensions = ring.Dimensions{Rows: 1, Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return encryptor.EncryptNew(pt)
}

func appendStage(result *evidence, name string, input, output *rlwe.Ciphertext, oracle []complex128, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, zeroSecretSimulation bool) *evidence {
	stage := stageEvidence{Operation: name, Status: "completed"}
	if input != nil {
		inputInfo := ciphertextSummary(input)
		stage.Input = &inputInfo
	}
	outputInfo := ciphertextSummary(output)
	stage.Output = &outputInfo
	if zeroSecretSimulation && !outputInfo.C1Zero {
		stage.Status = "stopped_nonzero_c1"
		stage.Error = "zero-c1 input invariant was lost; refusing to decrypt this Fast-simulation result"
		result.Stages = append(result.Stages, stage)
		result.Status = "stopped_at_" + name
		result.StopReason = stage.Error
		return result
	}
	decoded := make([]complex128, len(oracle))
	if err := encoder.Decode(decryptor.DecryptNew(output), decoded); err != nil {
		stage.Status = "decode_failed"
		stage.Error = err.Error()
		result.Stages = append(result.Stages, stage)
		result.Status = "stopped_at_" + name
		result.StopReason = "public DecryptNew/Decode failed at " + name
		return result
	}
	measurement := compare(oracle, decoded)
	stage.Oracle = &measurement
	if !measurement.Finite {
		stage.Status = "stopped_non_finite"
		stage.Error = "non-finite decrypted output or error aggregate"
		result.Stages = append(result.Stages, stage)
		result.Status = "stopped_at_" + name
		result.StopReason = stage.Error
		return result
	}
	if measurement.MaxComplexErr > catastrophicOracleError {
		stage.Status = "stopped_catastrophic_oracle_mismatch"
		stage.Error = fmt.Sprintf("max complex oracle error %.12g exceeded catastrophic-stop guard %.12g", measurement.MaxComplexErr, catastrophicOracleError)
		result.Stages = append(result.Stages, stage)
		result.Status = "stopped_at_" + name
		result.StopReason = stage.Error
		return result
	}
	result.Stages = append(result.Stages, stage)
	return nil
}

func failStage(result evidence, operation string, err error) (evidence, error) {
	result.Status = "stopped_at_" + operation
	result.StopReason = err.Error()
	result.Stages = append(result.Stages, stageEvidence{Operation: operation, Status: "operation_failed", Error: err.Error()})
	return result, fmt.Errorf("%s failed: %w", operation, err)
}

func stoppedResult(result *evidence) (evidence, error) {
	return *result, fmt.Errorf("audit stopped: %s", result.StopReason)
}

func ciphertextSummary(ct *rlwe.Ciphertext) ciphertextInfo {
	nonzero, total := 0, 0
	if ct != nil && len(ct.Value) > 1 {
		for _, row := range ct.Value[1].Coeffs {
			for _, coefficient := range row {
				total++
				if coefficient != 0 {
					nonzero++
				}
			}
		}
	}
	return ciphertextInfo{
		Level: ct.Level(), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(),
		IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		C1Zero: nonzero == 0, C1Nonzero: nonzero, C1Coefficients: total,
	}
}

func deterministicInputs(slots int) ([]complex128, []complex128) {
	a, b := make([]complex128, slots), make([]complex128, slots)
	for i := 0; i < slots; i++ {
		a[i] = complex(float64(i%7-3)/32, float64((2*i)%5-2)/64)
		b[i] = complex(float64(i%5-2)/16, float64(i%3-1)/32)
	}
	return a, b
}

func addValues(a, b []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] + b[i]
	}
	return result
}

func multiplyValues(a, b []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] * b[i]
	}
	return result
}

func rotateLeft(values []complex128, amount int) []complex128 {
	result := make([]complex128, len(values))
	for i := range values {
		result[i] = values[(i+amount)%len(values)]
	}
	return result
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

func allZero(rows [][]uint64) bool {
	for _, row := range rows {
		for _, coefficient := range row {
			if coefficient != 0 {
				return false
			}
		}
	}
	return true
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
