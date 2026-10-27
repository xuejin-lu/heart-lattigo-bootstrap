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
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	btsTask008                 = "BTS-NOTRACE-WARM-AND-NUMERICAL-008"
	btsTask009                 = "BTS-FAST-NOISE-FIDELITY-009"
	btsBackendStandard         = "standard"
	btsBackendFast             = "fast"
	btsNoiseOff                = "off"
	btsNoiseOn                 = "on"
	btsSourceManifestSHA       = "53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda"
	btsExpectedInputSHA        = "5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373"
	btsCallBudget              = 6
	btsNoiseSigma              = 4.004280012097432
	btsSlotLimbs               = 15
	btsSeed              int64 = 0
	btsNoiseRootSeed     int64 = 20261027
	btsAmplitude               = 1.0
)

func main() {
	if err := runBTSMeasurement(); err != nil {
		fmt.Fprintln(os.Stderr, "BTS measurement adapter:", err)
		os.Exit(1)
	}
}

func runBTSMeasurement() error {
	task := os.Getenv("BTS_TASK")
	if task == "" {
		task = btsTask008
	}
	backend := os.Getenv("BTS_BACKEND")
	noiseMode := os.Getenv("BTS_NOISE_MODE")
	if task == btsTask008 {
		if noiseMode != "" && noiseMode != btsNoiseOff {
			return fmt.Errorf("Batch008 is historical Noise-OFF only, got BTS_NOISE_MODE=%q", noiseMode)
		}
		noiseMode = btsNoiseOff
	} else if task == btsTask009 {
		if noiseMode != btsNoiseOff && noiseMode != btsNoiseOn {
			return fmt.Errorf("BTS-009 requires explicit BTS_NOISE_MODE=off or on, got %q", noiseMode)
		}
		if backend == btsBackendStandard && noiseMode == btsNoiseOn {
			return errors.New("BTS-009 Standard arm cannot enable Fast-only Noise")
		}
	} else {
		return fmt.Errorf("unsupported BTS task %q", task)
	}
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
		Task:                     task,
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
	if task == btsTask009 {
		run.Parameters.MaxLevel = params.MaxLevel()
		run.Parameters.QCount, run.Parameters.PCount = len(params.Q()), len(params.P())
		run.Parameters.Mod1Level = params.MaxLevel() - len(cfg.C2SKs)
		run.Parameters.Mod1Scale = math.Exp2(float64(cfg.EvalModLogQ))
		run.Parameters.Mod1Degree = cfg.Mod1Degree
		run.Parameters.Mod1PSDepth = int(math.Ceil(math.Log2(float64(cfg.Mod1Degree))))
		run.Parameters.DoubleAngle = cfg.DoubleAngle
		run.Circuit = "S2C -> ScaleDown -> ModUp -> C2S -> Mod1 Chebyshev power basis and PS evaluation -> 3 double-angle rounds -> BTSModePlain relabel"
	}
	if task == btsTask009 {
		run.SchemaVersion = perfmeasure.BTSFastNoise009SchemaVersion
		run.NoiseMode = noiseMode
		if noiseMode == btsNoiseOn {
			run.NoiseSigma = btsNoiseSigma
			run.NoiseRootSeed = btsNoiseRootSeed
		}
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
	if task == btsTask009 && noiseMode == btsNoiseOn {
		if err := enableBTSFixedSigmaNoise(bts.mod1.eval, btsNoiseSigma, btsNoiseRootSeed); err != nil {
			return failRun(summaryPath, &run, fmt.Errorf("enable 009 Fast fixed-sigma Noise Hook before any Bootstrap: %w", err))
		}
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
	if task == btsTask009 {
		profileConfig := struct {
			Task       string                               `json:"task"`
			Parameters perfmeasure.BTSMeasurementParameters `json:"parameters"`
			InputSHA   string                               `json:"input_sha256"`
		}{Task: task, Parameters: run.Parameters, InputSHA: inputSHA}
		profileBytes, marshalErr := json.Marshal(profileConfig)
		if marshalErr != nil {
			return failRun(summaryPath, &run, fmt.Errorf("marshal frozen BTS-009 profile: %w", marshalErr))
		}
		run.ProfileSHA256 = digest(profileBytes)
		armConfig := struct {
			ProfileSHA string  `json:"profile_sha256"`
			NoiseMode  string  `json:"noise_mode"`
			Sigma      float64 `json:"noise_sigma"`
			RootSeed   int64   `json:"noise_root_seed"`
		}{run.ProfileSHA256, noiseMode, run.NoiseSigma, run.NoiseRootSeed}
		armBytes, marshalErr := json.Marshal(armConfig)
		if marshalErr != nil {
			return failRun(summaryPath, &run, fmt.Errorf("marshal BTS-009 arm configuration: %w", marshalErr))
		}
		run.ArmConfigSHA256 = digest(armBytes)
	}
	var noiseSnapshot btsNoiseSnapshot
	if task == btsTask009 {
		noiseSnapshot, err = readBTSFixedSigmaNoiseSnapshot(bts.mod1.eval)
		if err != nil {
			return failRun(summaryPath, &run, fmt.Errorf("read initial 009 Noise Hook state: %w", err))
		}
		run.NoiseHookSupported = noiseSnapshot.Supported
		run.NoiseHookEnabled = noiseSnapshot.Enabled
		if noiseSnapshot.CtCtMultiplyCalls != 0 || noiseSnapshot.InjectedPolynomials != 0 || len(noiseSnapshot.Events) != 0 {
			return failRun(summaryPath, &run, fmt.Errorf("009 Noise Hook must start with empty evidence: %+v", noiseSnapshot))
		}
		if noiseMode == btsNoiseOn && (!noiseSnapshot.Supported || !noiseSnapshot.Enabled || noiseSnapshot.Sigma != btsNoiseSigma || noiseSnapshot.RootSeed != btsNoiseRootSeed) {
			return failRun(summaryPath, &run, fmt.Errorf("009 Fast Noise Hook state mismatch before Bootstrap: %+v", noiseSnapshot))
		}
		if noiseMode == btsNoiseOff && noiseSnapshot.Enabled {
			return failRun(summaryPath, &run, errors.New("009 Noise-OFF evaluator unexpectedly has Noise Hook enabled"))
		}
	}
	if err := persistRunSummary(summaryPath, run, false); err != nil {
		return err
	}
	fmt.Printf("[enc] L%d scale=%.12g slots=%d input_sha256=%s\n", input.Level(), input.Scale.Float64(), slots, inputSHA)
	if os.Getenv("BTS_PREFLIGHT_ONLY") == "1" {
		run.Status = "ZERO_BOOTSTRAP_PREFLIGHT_OK"
		run.UpdatedAt = time.Now().UTC()
		if err := persistRunSummary(summaryPath, run, false); err != nil {
			return err
		}
		fmt.Printf("[preflight] task=%s backend=%s noise=%s bootstrap_calls=0 input_sha256=%s\n", task, backend, noiseMode, inputSHA)
		return nil
	}

	budget := &perfmeasure.BootstrapAttemptBudget{
		Limit: btsCallBudget, JournalBase: journalBase,
		Metadata: map[string]string{
			"task": task, "arm": noiseMode, "backend": backend, "backend_commit": backendCommit,
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
		var bootstrapElapsedNS int64
		var allocatedBytes, allocationCount uint64
		noiseEventMissing := false
		budgetElapsedNS, invokeErr := budget.Invoke(phase, func() error {
			var before, after runtime.MemStats
			var started time.Time
			if task == btsTask009 {
				runtime.ReadMemStats(&before)
				started = time.Now()
			}
			output, bootstrapErr = bts.Bootstrap(callInput)
			if task == btsTask009 {
				bootstrapElapsedNS = time.Since(started).Nanoseconds()
				runtime.ReadMemStats(&after)
				allocatedBytes = after.TotalAlloc - before.TotalAlloc
				allocationCount = after.Mallocs - before.Mallocs
			}
			return bootstrapErr
		})
		run.BootstrapReservations = budget.Attempts
		elapsedNS := budgetElapsedNS
		if task == btsTask009 && bootstrapElapsedNS > 0 {
			elapsedNS = bootstrapElapsedNS
		}
		if task == btsTask009 {
			currentNoiseSnapshot, snapshotErr := readBTSFixedSigmaNoiseSnapshot(bts.mod1.eval)
			if snapshotErr != nil {
				return failRun(summaryPath, &run, fmt.Errorf("%s read post-Bootstrap Noise Hook state: %w", phase, snapshotErr))
			}
			if noiseMode == btsNoiseOn {
				artifact, captureErr := captureBTSFixedSigmaNoiseDelta(outputDir, index+1, phase, noiseSnapshot, currentNoiseSnapshot)
				if artifact.Path != "" {
					run.NoiseEvidence = append(run.NoiseEvidence, artifact)
				}
				if persistErr := persistRunSummary(summaryPath, run, false); persistErr != nil {
					return failRun(summaryPath, &run, errors.Join(captureErr, fmt.Errorf("persist Noise event evidence: %w", persistErr)))
				}
				if captureErr != nil {
					return failRun(summaryPath, &run, fmt.Errorf("%s Noise event capture: %w", phase, captureErr))
				}
				noiseEventMissing = artifact.EventCount == 0
			}
			if noiseMode == btsNoiseOff && currentNoiseSnapshot.Supported &&
				(currentNoiseSnapshot.Enabled || currentNoiseSnapshot.CtCtMultiplyCalls != noiseSnapshot.CtCtMultiplyCalls || len(currentNoiseSnapshot.Events) != len(noiseSnapshot.Events)) {
				return failRun(summaryPath, &run, fmt.Errorf("%s Fast-OFF unexpectedly changed Noise Hook state", phase))
			}
			noiseSnapshot = currentNoiseSnapshot
		}
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
			AllocatedBytes: allocatedBytes, Allocations: allocationCount,
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
		if noiseEventMissing {
			return failRun(summaryPath, &run, fmt.Errorf("%s completed without any observed ct×ct Noise Hook events", phase))
		}
		if task == btsTask009 && noiseMode == btsNoiseOn && len(run.NoiseEvidence) != index+1 {
			return failRun(summaryPath, &run, fmt.Errorf("%s completed without the required per-call Noise evidence artifact", phase))
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

type btsNoiseSnapshot struct {
	Supported           bool              `json:"supported"`
	Enabled             bool              `json:"enabled"`
	Sigma               float64           `json:"sigma_coefficient,omitempty"`
	RootSeed            int64             `json:"root_seed,omitempty"`
	CtCtMultiplyCalls   uint64            `json:"ct_ct_multiply_calls"`
	InjectedPolynomials uint64            `json:"injected_polynomials"`
	Events              []json.RawMessage `json:"events"`
}

type btsNoiseCallEvidence struct {
	SchemaVersion  string            `json:"schema_version"`
	Task           string            `json:"task"`
	Phase          string            `json:"phase"`
	AttemptIndex   int               `json:"attempt_index"`
	Sigma          float64           `json:"sigma_coefficient"`
	RootSeed       int64             `json:"root_seed"`
	OperationStart uint64            `json:"operation_start"`
	OperationEnd   uint64            `json:"operation_end"`
	Events         []json.RawMessage `json:"events"`
}

func enableBTSFixedSigmaNoise(evaluator *ckks.Evaluator, sigma float64, rootSeed int64) error {
	if evaluator == nil {
		return errors.New("BTS Mod1 CKKS evaluator is nil")
	}
	method := reflect.ValueOf(evaluator).MethodByName("EnableFixedSigmaBootstrapNoiseDiagnostic")
	if !method.IsValid() {
		return errors.New("selected Lattigo public CKKS evaluator does not expose the approved Fixed-Sigma Noise Hook")
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if method.Type().NumIn() != 2 || method.Type().In(0).Kind() != reflect.Float64 || method.Type().In(1).Kind() != reflect.Int64 || method.Type().NumOut() != 1 || method.Type().Out(0) != errorType {
		return errors.New("selected Lattigo Noise Hook has an unexpected API signature")
	}
	result := method.Call([]reflect.Value{reflect.ValueOf(sigma), reflect.ValueOf(rootSeed)})[0]
	if !result.IsNil() {
		return result.Interface().(error)
	}
	return nil
}

func readBTSFixedSigmaNoiseSnapshot(evaluator *ckks.Evaluator) (btsNoiseSnapshot, error) {
	if evaluator == nil {
		return btsNoiseSnapshot{}, errors.New("BTS Mod1 CKKS evaluator is nil")
	}
	method := reflect.ValueOf(evaluator).MethodByName("FixedSigmaBootstrapNoiseDiagnosticSnapshot")
	if !method.IsValid() {
		return btsNoiseSnapshot{Events: []json.RawMessage{}}, nil
	}
	if method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return btsNoiseSnapshot{}, errors.New("selected Lattigo Noise Hook snapshot has an unexpected API signature")
	}
	result := method.Call(nil)[0].Interface()
	encoded, err := json.Marshal(result)
	if err != nil {
		return btsNoiseSnapshot{}, fmt.Errorf("marshal public Noise Hook snapshot: %w", err)
	}
	var snapshot btsNoiseSnapshot
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		return btsNoiseSnapshot{}, fmt.Errorf("decode public Noise Hook snapshot: %w", err)
	}
	snapshot.Supported = true
	if snapshot.Events == nil {
		snapshot.Events = []json.RawMessage{}
	}
	if uint64(len(snapshot.Events)) != snapshot.CtCtMultiplyCalls {
		return snapshot, fmt.Errorf("Noise Hook operation count %d differs from event count %d", snapshot.CtCtMultiplyCalls, len(snapshot.Events))
	}
	return snapshot, nil
}

func captureBTSFixedSigmaNoiseDelta(
	outputDir string,
	attemptIndex int,
	phase string,
	previous, current btsNoiseSnapshot,
) (perfmeasure.BTSNoiseEvidenceArtifact, error) {
	if !current.Supported || !current.Enabled || current.Sigma != btsNoiseSigma || current.RootSeed != btsNoiseRootSeed {
		return perfmeasure.BTSNoiseEvidenceArtifact{}, fmt.Errorf("Noise Hook snapshot is not enabled with the frozen configuration: %+v", current)
	}
	if current.CtCtMultiplyCalls < previous.CtCtMultiplyCalls || current.CtCtMultiplyCalls > uint64(len(current.Events)) || current.InjectedPolynomials < previous.InjectedPolynomials {
		return perfmeasure.BTSNoiseEvidenceArtifact{}, errors.New("Noise Hook operation counter moved backwards or exceeds event evidence")
	}
	start := previous.CtCtMultiplyCalls
	end := current.CtCtMultiplyCalls
	callEvents := append([]json.RawMessage(nil), current.Events[start:end]...)
	evidence := btsNoiseCallEvidence{
		SchemaVersion: "bts-fast-noise-fidelity-009-noise-call.v1",
		Task:          btsTask009, Phase: phase, AttemptIndex: attemptIndex,
		Sigma: current.Sigma, RootSeed: current.RootSeed,
		OperationStart: start + 1, OperationEnd: end, Events: callEvents,
	}
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return perfmeasure.BTSNoiseEvidenceArtifact{}, err
	}
	data = append(data, '\n')
	path := filepath.Join(outputDir, fmt.Sprintf("noise-events-%02d.json", attemptIndex))
	if err := writeExclusiveSynced(path, data); err != nil {
		return perfmeasure.BTSNoiseEvidenceArtifact{}, err
	}
	return perfmeasure.BTSNoiseEvidenceArtifact{
		AttemptIndex: attemptIndex, Phase: phase, Path: path, SHA256: digest(data), Bytes: int64(len(data)),
		OperationStart: start + 1, OperationEnd: end, CtCtMultiplyCalls: end - start,
		InjectedPolynomials: current.InjectedPolynomials - previous.InjectedPolynomials, EventCount: len(callEvents),
	}, nil
}
