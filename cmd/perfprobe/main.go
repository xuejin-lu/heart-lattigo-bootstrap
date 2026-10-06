package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const canonicalLogN13InputSHA256 = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"

type effectiveParameters = perfmeasure.EffectiveParameters

type ciphertextState struct {
	Level       int     `json:"level"`
	Degree      int     `json:"degree"`
	Scale       string  `json:"scale"`
	ScaleLog2   float64 `json:"scale_log2"`
	LogRows     int     `json:"log_rows"`
	LogCols     int     `json:"log_cols"`
	QPrefixRows int     `json:"q_prefix_rows"`
	PrefixQ     string  `json:"prefix_q"`
}

type checkpoint struct {
	Name              string          `json:"name"`
	State             ciphertextState `json:"state"`
	DecodedSHA256     string          `json:"decoded_sha256"`
	DecodeMethod      string          `json:"decode_method"`
	SemanticallyValid bool            `json:"semantically_valid"`
}

type trial struct {
	Index             int                  `json:"index"`
	BootstrapSNR      numericalmetrics.SNR `json:"bootstrap_snr"`
	FreshKeyTrial     bool                 `json:"fresh_key_trial"`
	SecretKeyLevelP   int                  `json:"secret_key_level_p,omitempty"`
	EvaluatorPath     string               `json:"evaluator_path"`
	Output            ciphertextState      `json:"output"`
	OriginalSHA256    string               `json:"original_sha256"`
	PreDecodedSHA256  string               `json:"pre_decoded_sha256"`
	PostDecodedSHA256 string               `json:"post_decoded_sha256"`
}

type timingResult struct {
	Warmup      int       `json:"warmup"`
	Repetitions int       `json:"repetitions"`
	SamplesNS   []int64   `json:"samples_ns"`
	MedianNS    float64   `json:"median_ns"`
	MeanNS      float64   `json:"mean_ns"`
	MinNS       int64     `json:"min_ns"`
	MaxNS       int64     `json:"max_ns"`
	BytesPerOp  uint64    `json:"bytes_per_op"`
	AllocsPerOp uint64    `json:"allocs_per_op"`
	MeasuredAt  time.Time `json:"measured_at"`
}

type stageTiming struct {
	Stage     string `json:"stage"`
	ElapsedNS int64  `json:"elapsed_ns"`
	Samples   int    `json:"samples"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type document struct {
	SchemaVersion       string              `json:"schema_version"`
	Timestamp           time.Time           `json:"timestamp"`
	Profile             string              `json:"profile"`
	Backend             string              `json:"backend"`
	BackendCommit       string              `json:"backend_commit"`
	PrimaryPath         string              `json:"primary_path"`
	PrimaryCommit       string              `json:"primary_commit"`
	PrimaryDirty        bool                `json:"primary_dirty"`
	SecondaryDirty      bool                `json:"secondary_dirty"`
	SecondaryPath       string              `json:"secondary_path"`
	SecondaryRef        string              `json:"secondary_ref"`
	ConfigSHA256        string              `json:"config_sha256"`
	ConfigPath          string              `json:"config_path"`
	GoVersion           string              `json:"go_version"`
	OS                  string              `json:"os"`
	Arch                string              `json:"arch"`
	CPU                 string              `json:"cpu"`
	NumCPU              int                 `json:"num_cpu"`
	GOMAXPROCS          int                 `json:"gomaxprocs"`
	InputSHA256         string              `json:"input_sha256"`
	InputMetadataSHA256 string              `json:"input_metadata_sha256"`
	InputState          ciphertextState     `json:"input_ciphertext_state"`
	InputKind           string              `json:"input_kind"`
	Parameters          effectiveParameters `json:"effective_parameters"`
	Timing              timingResult        `json:"full_bootstrap_timing"`
	StageTimings        []stageTiming       `json:"separate_stage_timing"`
	Trials              []trial             `json:"numerical_trials"`
	Checkpoints         []checkpoint        `json:"stage_checkpoints"`
	Limitations         []string            `json:"limitations"`
}

type complexValue struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type vectors struct {
	PreBootstrap []complexValue            `json:"pre_bootstrap"`
	PreTrials    [][]complexValue          `json:"pre_bootstrap_trials"`
	Original     []complexValue            `json:"original"`
	Trials       [][]complexValue          `json:"bootstrap_trials"`
	Checkpoints  map[string][]complexValue `json:"checkpoints"`
}

type cliOptions struct {
	profile        string
	config         string
	out            string
	vectorsOut     string
	backendCommit  string
	backendRef     string
	secondaryRoot  string
	warmup         int
	repetitions    int
	standardTrials int
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "perfprobe:", err)
		os.Exit(1)
	}
}

func run() error {
	var opts cliOptions
	flag.StringVar(&opts.profile, "profile", "", "logn13 or logn16")
	flag.StringVar(&opts.config, "config", "", "bootstrap config JSON")
	flag.StringVar(&opts.out, "out", "", "compact result JSON")
	flag.StringVar(&opts.vectorsOut, "vectors-out", "", "ephemeral decoded vectors JSON")
	flag.StringVar(&opts.backendCommit, "backend-commit", "", "exact Secondary commit used by this process")
	flag.StringVar(&opts.backendRef, "backend-ref", "", "human-readable pinned Secondary ref")
	flag.StringVar(&opts.secondaryRoot, "secondary-root", "", "clean detached Secondary worktree used by this process")
	flag.IntVar(&opts.warmup, "warmup", 1, "untimed full Bootstrap warmups")
	flag.IntVar(&opts.repetitions, "repetitions", 7, "timed full Bootstrap repetitions")
	flag.IntVar(&opts.standardTrials, "standard-trials", 3, "independent Standard secret/evaluation-key numerical trials")
	flag.Parse()
	if opts.profile != "logn13" && opts.profile != "logn16" {
		return fmt.Errorf("--profile must be logn13 or logn16")
	}
	if opts.config == "" || opts.out == "" || opts.vectorsOut == "" || opts.backendCommit == "" || opts.secondaryRoot == "" || opts.backendRef == "" {
		return errors.New("--config, --out, --vectors-out, --backend-commit, --backend-ref, and --secondary-root are required")
	}
	if opts.warmup < 1 || opts.repetitions < 7 || opts.standardTrials < 3 {
		return errors.New("measurement requires warmup >= 1, repetitions >= 7, and standard-trials >= 3")
	}
	cfg, configSum, err := perfmeasure.LoadConfig(opts.config)
	if err != nil {
		return err
	}
	configHash := hex.EncodeToString(configSum[:])
	profileLogN := 13
	if opts.profile == "logn16" {
		profileLogN = 16
	}
	if cfg.LogN != profileLogN {
		return fmt.Errorf("%s profile requires config log_n=%d, got %d", opts.profile, profileLogN, cfg.LogN)
	}
	residual, params, effective, err := perfmeasure.ParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	values := perfmeasure.DeterministicInput(effective.InputSlots)
	inputHash := perfmeasure.Fingerprint(values)
	if opts.profile == "logn13" && inputHash != canonicalLogN13InputSHA256 {
		return fmt.Errorf("canonical LogN13 input fingerprint mismatch: %s", inputHash)
	}
	input, err := perfmeasure.EncodePlaintextLikeInput(residual, effective.LogSlots, values)
	if err != nil {
		return err
	}
	backend, err := newBackend(params, residual)
	if err != nil {
		return err
	}
	inputState, err := backendStateOf(backend, input, params.BootstrappingParameters.Q())
	if err != nil {
		return fmt.Errorf("inspect input ciphertext metadata: %w", err)
	}
	inputMetadata, err := json.Marshal(inputState)
	if err != nil {
		return fmt.Errorf("marshal input ciphertext metadata: %w", err)
	}
	inputMetadataHash := sha256.Sum256(inputMetadata)
	preDecoded, err := backend.Decode(input)
	if err != nil {
		return fmt.Errorf("decode pre-Bootstrap: %w", err)
	}
	primaryRoot, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		return fmt.Errorf("Primary provenance: %w", err)
	}
	secondaryRoot, secondaryCommit, _, err := cleanSecondaryState(opts.secondaryRoot, opts.backendCommit)
	if err != nil {
		return fmt.Errorf("Secondary provenance: %w", err)
	}
	if secondaryCommit != opts.backendCommit {
		return fmt.Errorf("Secondary HEAD %s does not match pinned backend commit %s", secondaryCommit, opts.backendCommit)
	}

	measured, err := measureBootstrap(backend, input, opts.warmup, opts.repetitions)
	if err != nil {
		return err
	}
	numericalTrials := make([]trial, 0, max(opts.standardTrials, 2))
	trialVectors := make([][]complexValue, 0, cap(numericalTrials))
	preTrialVectors := make([][]complexValue, 0, cap(numericalTrials))
	standardKeys := make([]*rlwe.SecretKey, 0, opts.standardTrials)
	var representativeOutput *rlwe.Ciphertext
	if backend.Name() == "fast" {
		for i := 0; i < 2; i++ {
			out, err := backend.Bootstrap(input.CopyNew())
			if err != nil {
				return fmt.Errorf("Fast numerical trial %d: %w", i+1, err)
			}
			decoded, err := backend.Decode(out)
			if err != nil {
				return fmt.Errorf("decode Fast trial %d: %w", i+1, err)
			}
			record, err := makeTrial(i+1, backend, values, preDecoded, decoded, out, params.BootstrappingParameters.Q())
			if err != nil {
				return fmt.Errorf("Fast trial %d metadata: %w", i+1, err)
			}
			numericalTrials = append(numericalTrials, record)
			preTrialVectors = append(preTrialVectors, encodeVector(preDecoded))
			trialVectors = append(trialVectors, encodeVector(decoded))
			if representativeOutput == nil {
				representativeOutput = out.CopyNew()
			}
		}
	} else {
		for i := 0; i < opts.standardTrials; i++ {
			if i > 0 {
				backend, err = newBackend(params, residual)
				if err != nil {
					return fmt.Errorf("create independent Standard trial %d: %w", i+1, err)
				}
			}
			secret := backend.SecretKeyForTrial()
			if secret == nil {
				return fmt.Errorf("Standard trial %d did not expose its in-memory key for uniqueness validation", i+1)
			}
			for priorIndex, prior := range standardKeys {
				if secret.Equal(prior) {
					return fmt.Errorf("Standard trial %d generated a key identical to trial %d", i+1, priorIndex+1)
				}
			}
			standardKeys = append(standardKeys, secret)
			trialPreDecoded, err := backend.Decode(input)
			if err != nil {
				return fmt.Errorf("decode Standard pre-Bootstrap trial %d: %w", i+1, err)
			}
			out, err := backend.Bootstrap(input.CopyNew())
			if err != nil {
				return fmt.Errorf("Standard numerical trial %d: %w", i+1, err)
			}
			decoded, err := backend.Decode(out)
			if err != nil {
				return fmt.Errorf("decode Standard trial %d: %w", i+1, err)
			}
			record, err := makeTrial(i+1, backend, values, trialPreDecoded, decoded, out, params.BootstrappingParameters.Q())
			if err != nil {
				return fmt.Errorf("Standard trial %d metadata: %w", i+1, err)
			}
			numericalTrials = append(numericalTrials, record)
			preTrialVectors = append(preTrialVectors, encodeVector(trialPreDecoded))
			trialVectors = append(trialVectors, encodeVector(decoded))
			representativeOutput = out.CopyNew()
		}
	}
	stages, stageVectors, stageTimings, err := runStages(backend, input, representativeOutput, params.BootstrappingParameters.Q())
	if err != nil {
		return fmt.Errorf("stage replay: %w", err)
	}

	doc := document{
		SchemaVersion: "fast-standard-perfprobe.v1", Timestamp: time.Now().UTC(),
		Profile: opts.profile, Backend: backend.Name(), BackendCommit: opts.backendCommit,
		PrimaryPath: primaryRoot, PrimaryCommit: primaryCommit, PrimaryDirty: false, SecondaryDirty: false,
		SecondaryPath: secondaryRoot, SecondaryRef: opts.backendRef, ConfigSHA256: configHash, ConfigPath: opts.config,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
		CPU: cpuModel(), NumCPU: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0),
		InputSHA256: inputHash, InputMetadataSHA256: hex.EncodeToString(inputMetadataHash[:]), InputState: inputState,
		InputKind:  "c0=encoded-message,c1=0; specialized zero-a plaintext-like comparator",
		Parameters: effective, Timing: measured, StageTimings: stageTimings, Trials: numericalTrials, Checkpoints: stages,
		Limitations: []string{
			"Timing samples cover only the public Bootstrap call; key generation, setup, encoding, decoding, and analysis are outside the timed region.",
			"Stage checkpoints are a separate public-stage replay and are not summed to estimate total Bootstrap time.",
			"The input has c1=0 and does not represent an encrypted-input security or RLWE-noise experiment.",
		},
	}
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		return err
	}
	vectorDoc := vectors{
		PreBootstrap: encodeVector(preDecoded), PreTrials: preTrialVectors, Original: encodeVector(values),
		Trials: trialVectors, Checkpoints: stageVectors,
	}
	if err := writeExclusiveJSON(opts.vectorsOut, vectorDoc); err != nil {
		return err
	}
	fmt.Printf("%s %s timing=%s numerical_trials=%d checkpoints=%d\n", opts.profile, backend.Name(), opts.out, len(numericalTrials), len(stages))
	return nil
}

type backendAdapter interface {
	Name() string
	EvaluatorPath() string
	KeyTrialEvidence() (bool, int)
	SecretKeyForTrial() *rlwe.SecretKey
	Bootstrap(*rlwe.Ciphertext) (*rlwe.Ciphertext, error)
	Decode(*rlwe.Ciphertext) ([]complex128, error)
	DecodeStage(*rlwe.Ciphertext, string) ([]complex128, string, error)
	PrefixRows(*rlwe.Ciphertext) (int, error)
	Pack([]rlwe.Ciphertext) ([]rlwe.Ciphertext, error)
	ScaleDown(*rlwe.Ciphertext) (*rlwe.Ciphertext, error)
	ModUp(*rlwe.Ciphertext) (*rlwe.Ciphertext, error)
	CoeffsToSlots(*rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error)
	EvalMod(*rlwe.Ciphertext) (*rlwe.Ciphertext, error)
	SlotsToCoeffs(*rlwe.Ciphertext, *rlwe.Ciphertext) (*rlwe.Ciphertext, error)
}

func measureBootstrap(backend backendAdapter, input *rlwe.Ciphertext, warmup, repetitions int) (timingResult, error) {
	for i := 0; i < warmup; i++ {
		if _, err := backend.Bootstrap(input.CopyNew()); err != nil {
			return timingResult{}, fmt.Errorf("warmup %d: %w", i+1, err)
		}
	}
	runtime.GC()
	samples := make([]int64, repetitions)
	var totalBytes, totalAllocs uint64
	var before, after runtime.MemStats
	for i := range samples {
		ct := input.CopyNew()
		runtime.ReadMemStats(&before)
		start := time.Now()
		_, err := backend.Bootstrap(ct)
		elapsed := time.Since(start)
		runtime.ReadMemStats(&after)
		if err != nil {
			return timingResult{}, fmt.Errorf("timed repetition %d: %w", i+1, err)
		}
		samples[i] = elapsed.Nanoseconds()
		totalBytes += after.TotalAlloc - before.TotalAlloc
		totalAllocs += after.Mallocs - before.Mallocs
	}
	ordered := append([]int64(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	var total float64
	for _, sample := range samples {
		total += float64(sample)
	}
	median := float64(ordered[len(ordered)/2])
	if len(ordered)%2 == 0 {
		median = (float64(ordered[len(ordered)/2-1]) + float64(ordered[len(ordered)/2])) / 2
	}
	return timingResult{
		Warmup: warmup, Repetitions: repetitions, SamplesNS: samples, MedianNS: median,
		MeanNS: total / float64(repetitions), MinNS: ordered[0], MaxNS: ordered[len(ordered)-1],
		BytesPerOp: totalBytes / uint64(repetitions), AllocsPerOp: totalAllocs / uint64(repetitions),
		MeasuredAt: time.Now().UTC(),
	}, nil
}

func runStages(backend backendAdapter, input, final *rlwe.Ciphertext, q []uint64) ([]checkpoint, map[string][]complexValue, []stageTiming, error) {
	var checkpoints []checkpoint
	var timings []stageTiming
	vectorsByName := make(map[string][]complexValue)
	add := func(name string, ct *rlwe.Ciphertext) error {
		if ct == nil {
			checkpoints = append(checkpoints, checkpoint{Name: name, SemanticallyValid: false, DecodeMethod: backend.Name() + "-decode"})
			return nil
		}
		decoded, decodeMethod, err := backend.DecodeStage(ct, name)
		if err != nil {
			return fmt.Errorf("decode %s: %w", name, err)
		}
		state, err := backendStateOf(backend, ct, q)
		if err != nil {
			return fmt.Errorf("inspect %s metadata: %w", name, err)
		}
		checkpoints = append(checkpoints, checkpoint{Name: name, State: state, DecodedSHA256: perfmeasure.Fingerprint(decoded), DecodeMethod: decodeMethod, SemanticallyValid: true})
		if _, exists := vectorsByName[name]; !exists {
			vectorsByName[name] = encodeVector(decoded)
		}
		return nil
	}
	if err := add("input", input); err != nil {
		return nil, nil, nil, err
	}
	timeStage := func(name string, start time.Time) {
		timings = append(timings, stageTiming{Stage: name, ElapsedNS: time.Since(start).Nanoseconds(), Samples: 1, Available: true})
	}
	packedInput := []rlwe.Ciphertext{*input.CopyNew()}
	start := time.Now()
	packed, err := backend.Pack(packedInput)
	timeStage("pack_and_switch", start)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("pack: %w", err)
	}
	if len(packed) != 1 {
		return nil, nil, nil, fmt.Errorf("pack returned %d ciphertexts, expected one", len(packed))
	}
	start = time.Now()
	ct, err := backend.ScaleDown(&packed[0])
	timeStage("scale_down", start)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("scale down: %w", err)
	}
	if err := add("scale_down", ct); err != nil {
		return nil, nil, nil, err
	}
	start = time.Now()
	ct, err = backend.ModUp(ct)
	timeStage("mod_up", start)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("mod up: %w", err)
	}
	if err := add("mod_up", ct); err != nil {
		return nil, nil, nil, err
	}
	start = time.Now()
	realCT, imagCT, err := backend.CoeffsToSlots(ct)
	timeStage("coeffs_to_slots", start)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("C2S: %w", err)
	}
	if err := add("c2s_real", realCT); err != nil {
		return nil, nil, nil, err
	}
	if err := add("c2s_imag", imagCT); err != nil {
		return nil, nil, nil, err
	}
	start = time.Now()
	realCT, err = backend.EvalMod(realCT)
	timeStage("evalmod_real", start)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("EvalMod real: %w", err)
	}
	if err := add("evalmod_real", realCT); err != nil {
		return nil, nil, nil, err
	}
	if imagCT != nil {
		start = time.Now()
		imagCT, err = backend.EvalMod(imagCT)
		timeStage("evalmod_imag", start)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("EvalMod imag: %w", err)
		}
	} else {
		timings = append(timings, stageTiming{Stage: "evalmod_imag", Available: false, Reason: "C2S did not return an imaginary branch"})
	}
	if err := add("evalmod_imag", imagCT); err != nil {
		return nil, nil, nil, err
	}
	start = time.Now()
	ct, err = backend.SlotsToCoeffs(realCT, imagCT)
	timeStage("slots_to_coeffs", start)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("S2C: %w", err)
	}
	if err := add("s2c", ct); err != nil {
		return nil, nil, nil, err
	}
	if err := add("final_public_output", final); err != nil {
		return nil, nil, nil, err
	}
	return checkpoints, vectorsByName, timings, nil
}

func makeTrial(index int, backend backendAdapter, original, pre, post []complex128, output *rlwe.Ciphertext, q []uint64) (trial, error) {
	freshKeys, levelP := backend.KeyTrialEvidence()
	state, err := backendStateOf(backend, output, q)
	if err != nil {
		return trial{}, err
	}
	return trial{Index: index, BootstrapSNR: numericalmetrics.Compare(pre, post), FreshKeyTrial: freshKeys,
		SecretKeyLevelP: levelP, EvaluatorPath: backend.EvaluatorPath(), Output: state,
		OriginalSHA256: perfmeasure.Fingerprint(original), PreDecodedSHA256: perfmeasure.Fingerprint(pre), PostDecodedSHA256: perfmeasure.Fingerprint(post)}, nil
}

func backendStateOf(backend backendAdapter, ct *rlwe.Ciphertext, q []uint64) (ciphertextState, error) {
	rows, err := backend.PrefixRows(ct)
	if err != nil {
		return ciphertextState{}, err
	}
	return stateOf(ct, q, rows), nil
}

func stateOf(ct *rlwe.Ciphertext, q []uint64, prefixRows int) ciphertextState {
	if ct == nil {
		return ciphertextState{}
	}
	state := ciphertextState{Level: ct.Level(), Degree: ct.Degree(), Scale: ct.Scale.Value.Text('e', 80), ScaleLog2: ct.Scale.Log2(),
		LogRows: ct.LogDimensions.Rows, LogCols: ct.LogDimensions.Cols}
	if q != nil && ct.Level() >= 0 {
		rows := min(prefixRows, len(q))
		if rows > len(q) {
			rows = len(q)
		}
		if rows > 0 {
			product := big.NewInt(1)
			for _, prime := range q[:rows] {
				product.Mul(product, new(big.Int).SetUint64(prime))
			}
			state.QPrefixRows, state.PrefixQ = rows, product.String()
		}
	}
	return state
}

func encodeVector(values []complex128) []complexValue {
	out := make([]complexValue, len(values))
	for i, value := range values {
		out[i] = complexValue{Real: real(value), Imag: imag(value)}
	}
	return out
}

func cleanRepositoryState(path string) (string, string, error) {
	root, err := gitOutput(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", err
	}
	commit, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return "", "", err
	}
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return "", "", err
	}
	if status != "" {
		return "", "", fmt.Errorf("worktree is dirty:\n%s", status)
	}
	return root, commit, nil
}

func cleanSecondaryState(path, requestedCommit string) (string, string, string, error) {
	root, err := gitOutput(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", "", err
	}
	commit, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return "", "", "", err
	}
	if requestedCommit != "" && commit != requestedCommit {
		return "", commit, "", fmt.Errorf("HEAD=%s, requested=%s", commit, requestedCommit)
	}
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return "", "", "", err
	}
	if status != "" {
		return "", commit, "", fmt.Errorf("worktree is dirty:\n%s", status)
	}
	ref, _ := gitOutput(root, "branch", "--show-current")
	if ref == "" {
		ref = commit
	}
	return root, commit, ref, nil
}

func gitOutput(root string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", root}, args...)
	command := exec.Command("git", commandArgs...)
	data, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(data)), nil
}

func cpuModel() string {
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			if model := strings.TrimSpace(string(output)); model != "" {
				return model
			}
		}
	}
	return runtime.GOARCH
}

func writeExclusiveJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = f.Write(append(data, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
