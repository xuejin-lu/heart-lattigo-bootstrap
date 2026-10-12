// This overlay replaces only external/BTS/main.go. It retains the original
// BTS configuration and ordinary public CKKS lifecycle, and adds bounded
// sample repetition, durable reservations, and f64le result capture.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	btsTask                    = "BTS-NOTRACE-WARM-AND-NUMERICAL-008"
	btsBackendStandard         = "standard"
	btsBackendFast             = "fast"
	btsSourceManifestSHA       = "53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda"
	btsExpectedInputSHA        = "5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373"
	btsCallBudget              = 6
	btsSlotLimbs               = 15
	btsSeed              int64 = 0
	btsAmplitude               = 1.0
)

func main() {
	if err := runBTSMeasurement(); err != nil {
		fmt.Fprintln(os.Stderr, "BTS Batch008 measurement adapter:", err)
		os.Exit(1)
	}
}

func runBTSMeasurement() error {
	backend := os.Getenv("BTS_BACKEND")
	backendCommit := os.Getenv("BTS_BACKEND_COMMIT")
	binarySHA := os.Getenv("BTS_BINARY_SHA256")
	adapterSHA := os.Getenv("BTS_MEASUREMENT_ADAPTER_SHA256")
	primaryCommit := os.Getenv("BTS_PRIMARY_COMMIT")
	manifestSHA := os.Getenv("BTS_SOURCE_MANIFEST_SHA256")
	outputDir := os.Getenv("BTS_OUTPUT_DIR")
	summaryPath := os.Getenv("BTS_RUN_SUMMARY")
	journalBase := os.Getenv("BTS_ATTEMPT_JOURNAL_BASE")
	if backend != btsBackendStandard && backend != btsBackendFast {
		return fmt.Errorf("BTS_BACKEND must be standard or fast, got %q", backend)
	}
	if backendCommit == "" || binarySHA == "" || adapterSHA == "" || primaryCommit == "" || manifestSHA != btsSourceManifestSHA {
		return errors.New("missing or invalid build/source provenance environment")
	}
	if outputDir == "" || summaryPath == "" || journalBase == "" {
		return errors.New("BTS_OUTPUT_DIR, BTS_RUN_SUMMARY, and BTS_ATTEMPT_JOURNAL_BASE are required")
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return err
	}
	if err := ensureAbsent(summaryPath); err != nil {
		return err
	}

	cfg := BTSConfig{
		LogN: 16, LogDefaultScale: 32, Q0Log: 42, S2CLogQ: 32,
		EvalModLogQ: 42, EvalModLimbs: 8, SlotLogQ: 32, SlotLimbs: btsSlotLimbs,
		C2SLogQ: 42, LogP: []int{42, 42, 42, 42, 42}, H: 192,
		EphemeralH: 32, LogMessageRatio: 10,
		S2CKs: []int{4, 5, 6}, S2CN1s: []int{8, 256, 8192},
		C2SKs: []int{5, 5, 5}, C2SN1s: []int{8192, 512, 16}, C2SIntervalK: 16,
		Mod1Type: MyCosContinuous, Mod1Degree: 31, DoubleAngle: 3, K: 16,
		Mode: BTSModePlain,
	}
	params, err := cfg.GenParams()
	if err != nil {
		return fmt.Errorf("construct fixed BTS parameters: %w", err)
	}
	if params.LogN() != 16 || params.N() != 1<<16 || params.MaxSlots() != 1<<15 || cfg.SlotLimbs != 15 || cfg.Mode != BTSModePlain {
		return fmt.Errorf("effective BTS profile mismatch: LogN=%d N=%d slots=%d SlotLimbs=%d mode=%s", params.LogN(), params.N(), params.MaxSlots(), cfg.SlotLimbs, cfg.Mode)
	}

	startedAt := time.Now().UTC()
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	run := perfmeasure.BTSRunDocument{
		SchemaVersion:            perfmeasure.BTSNoTrace008SchemaVersion,
		Task:                     btsTask,
		Status:                   "RUNNING",
		Backend:                  backend,
		BackendCommit:            backendCommit,
		BinarySHA256:             binarySHA,
		BTSManifestSHA256:        manifestSHA,
		MeasurementAdapterSHA256: adapterSHA,
		PrimaryCommit:            primaryCommit,
		OriginalSourceRoot:       workingDir,
		StartedAt:                startedAt,
		UpdatedAt:                startedAt,
		MeasurementProcesses:     1,
		BootstrapBudget:          btsCallBudget,
		Environment: perfmeasure.BTSMeasurementEnvironment{
			GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
			CPU: perfmeasure.CPUModel(), NumCPU: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0),
		},
		Parameters: perfmeasure.BTSMeasurementParameters{
			LogN: params.LogN(), RingN: params.N(), Slots: params.MaxSlots(),
			SlotLimbs: cfg.SlotLimbs, Mode: "BTSModePlain", Seed: btsSeed,
			Amplitude: btsAmplitude, EphemeralH: cfg.EphemeralH,
			QPrimes: append([]uint64(nil), params.Q()...), PPrimes: append([]uint64(nil), params.P()...),
			InputLevel: cfg.S2CStartLevel(), InputScale: params.DefaultScale().Float64(),
			OutputLevel: cfg.OutputLevel(), OutputScale: params.DefaultScale().Float64(),
		},
	}
	if err := persistRunSummary(summaryPath, run, true); err != nil {
		return err
	}

	kgen := rlwe.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyNew()
	encoder := ckks.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, sk)
	decryptor := rlwe.NewDecryptor(params, sk)
	fmt.Printf("params: LogN=%d N=%d MaxLevel=L%d slots=%d SlotLimbs=%d mode=%s\n", params.LogN(), params.N(), params.MaxLevel(), params.MaxSlots(), cfg.SlotLimbs, cfg.Mode)

	newBTSStart := time.Now()
	bts, err := NewBTS(cfg, params, sk)
	run.NewBTSNS = time.Since(newBTSStart).Nanoseconds()
	if err != nil {
		return failRun(summaryPath, &run, fmt.Errorf("NewBTS: %w", err))
	}
	fmt.Printf("[NewBTS] %s (%.6f s) κ=%.3f\n", time.Duration(run.NewBTSNS), float64(run.NewBTSNS)/1e9, bts.Kappa)

	slots := params.MaxSlots()
	rng := rand.New(rand.NewSource(btsSeed))
	want := make([]float64, slots)
	for i := range want {
		want[i] = btsAmplitude * (2*rng.Float64() - 1)
	}
	pt := ckks.NewPlaintext(params, cfg.S2CStartLevel())
	if err := encoder.Encode(want, pt); err != nil {
		return failRun(summaryPath, &run, fmt.Errorf("encode original input: %w", err))
	}
	input, err := encryptor.EncryptNew(pt)
	if err != nil {
		return failRun(summaryPath, &run, fmt.Errorf("encrypt original input: %w", err))
	}
	inputBytes := encodeFloat64Vector(want)
	inputSHA := digest(inputBytes)
	if inputSHA != btsExpectedInputSHA {
		return failRun(summaryPath, &run, fmt.Errorf("original input fingerprint mismatch: got %s want %s", inputSHA, btsExpectedInputSHA))
	}
	inputPath := filepath.Join(outputDir, "input.f64le")
	if err := writeExclusiveSynced(inputPath, inputBytes); err != nil {
		return failRun(summaryPath, &run, fmt.Errorf("write original input vector: %w", err))
	}
	run.Input = perfmeasure.BTSVectorArtifact{
		Path: inputPath, Encoding: "IEEE754-f64-little-endian", SHA256: inputSHA,
		Slots: slots, FiniteSlots: finiteCount(want), Level: input.Level(), Scale: input.Scale.Float64(),
	}
	run.Parameters.InputLevel = input.Level()
	run.Parameters.InputScale = input.Scale.Float64()
	if err := persistRunSummary(summaryPath, run, false); err != nil {
		return err
	}
	fmt.Printf("[enc] L%d scale=%.12g slots=%d input_sha256=%s\n", input.Level(), input.Scale.Float64(), slots, inputSHA)

	budget := &perfmeasure.BootstrapAttemptBudget{
		Limit: btsCallBudget, JournalBase: journalBase,
		Metadata: map[string]string{
			"task": btsTask, "backend": backend, "backend_commit": backendCommit,
			"binary_sha256": binarySHA, "bts_source_manifest_sha256": manifestSHA,
			"measurement_adapter_sha256": adapterSHA, "primary_commit": primaryCommit,
			"input_vector_sha256": inputSHA, "command": strings.Join(os.Args, " "),
		},
	}
	for index := 0; index < btsCallBudget; index++ {
		phase := fmt.Sprintf("warm_bootstrap_%02d", index)
		if index == 0 {
			phase = "first_cold_bootstrap"
		}
		callInput := input.CopyNew()
		var output *rlwe.Ciphertext
		var bootstrapErr error
		elapsedNS, invokeErr := budget.Invoke(phase, func() error {
			output, bootstrapErr = bts.Bootstrap(callInput)
			return bootstrapErr
		})
		run.BootstrapReservations = budget.Attempts
		if invokeErr != nil {
			return failRun(summaryPath, &run, fmt.Errorf("%s: %w", phase, invokeErr))
		}
		if output == nil {
			return failRun(summaryPath, &run, fmt.Errorf("%s returned a nil ciphertext", phase))
		}

		plainOut := decryptor.DecryptNew(output)
		got := make([]float64, slots)
		if err := encoder.Decode(plainOut, got); err != nil {
			return failRun(summaryPath, &run, fmt.Errorf("%s decode: %w", phase, err))
		}
		if len(got) != slots || finiteCount(got) != slots {
			return failRun(summaryPath, &run, fmt.Errorf("%s decoded %d/%d finite slots", phase, finiteCount(got), slots))
		}
		metric := numericalmetrics.CompareRealVectors(want, got)
		if metric.Status != "FINITE" {
			return failRun(summaryPath, &run, fmt.Errorf("%s input-oracle comparison: %s: %s", phase, metric.Status, metric.Reason))
		}
		vectorBytes := encodeFloat64Vector(got)
		vectorPath := filepath.Join(outputDir, fmt.Sprintf("sample-%02d-output.f64le", index+1))
		if err := writeExclusiveSynced(vectorPath, vectorBytes); err != nil {
			return failRun(summaryPath, &run, fmt.Errorf("%s write decoded vector: %w", phase, err))
		}
		sample := perfmeasure.BTSBootstrapSample{
			AttemptIndex: index + 1, Phase: phase, BootstrapElapsedNS: elapsedNS,
			InputLevel: input.Level(), InputScale: input.Scale.Float64(),
			OutputCiphertextLevel: output.Level(), OutputCiphertextScale: output.Scale.Float64(),
			DecodedPlaintextLevel: plainOut.Level(), DecodedPlaintextScale: plainOut.Scale.Float64(),
			DecodedSlots: len(got), FiniteDecodedSlots: slots,
			VectorPath: vectorPath, VectorEncoding: "IEEE754-f64-little-endian", VectorSHA256: digest(vectorBytes),
			MaxAbsErrorVsInput: metric.MaxAbs, MeanAbsErrorVsInput: metric.MeanAbs, WorstErrorSlot: metric.WorstIndex,
		}
		run.Samples = append(run.Samples, sample)
		run.UpdatedAt = time.Now().UTC()
		if err := persistRunSummary(summaryPath, run, false); err != nil {
			return err
		}
		fmt.Printf("[Bootstrap %d/6 %s] %s (%d ns) -> ciphertext L%d scale=%.12g plaintext L%d scale=%.12g; max|err|=%.6e mean|err|=%.6e @slot %d; vector_sha256=%s\n",
			index+1, phase, time.Duration(elapsedNS), elapsedNS, output.Level(), output.Scale.Float64(), plainOut.Level(), plainOut.Scale.Float64(),
			metric.MaxAbs, metric.MeanAbs, metric.WorstIndex, sample.VectorSHA256)
	}
	run.Status = "COMPLETE"
	run.BootstrapReservations = budget.Attempts
	run.UpdatedAt = time.Now().UTC()
	if err := persistRunSummary(summaryPath, run, false); err != nil {
		return err
	}
	return nil
}

func failRun(path string, run *perfmeasure.BTSRunDocument, cause error) error {
	run.Status = "FAILED"
	run.Failure = cause.Error()
	run.UpdatedAt = time.Now().UTC()
	if err := persistRunSummary(path, *run, false); err != nil {
		return errors.Join(cause, fmt.Errorf("persist failure status: %w", err))
	}
	return cause
}

func persistRunSummary(path string, run perfmeasure.BTSRunDocument, exclusive bool) error {
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if exclusive {
		return writeExclusiveSynced(path, data)
	}
	temporary := path + ".tmp"
	if err := writeExclusiveSynced(temporary, data); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	return nil
}

func ensureAbsent(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("refusing to overwrite existing artifact %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeExclusiveSynced(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func encodeFloat64Vector(values []float64) []byte {
	data := make([]byte, len(values)*8)
	for index, value := range values {
		binary.LittleEndian.PutUint64(data[index*8:], math.Float64bits(value))
	}
	return data
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func finiteCount(values []float64) int {
	count := 0
	for _, value := range values {
		if !math.IsNaN(value) && !math.IsInf(value, 0) {
			count++
		}
	}
	return count
}
