package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type RepositoryMetadata struct {
	Path   string `json:"path"`
	Commit string `json:"commit,omitempty"`
	Ref    string `json:"ref,omitempty"`
	Dirty  bool   `json:"dirty"`
}

type EnvironmentMetadata struct {
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"architecture"`
	CPU       string `json:"cpu,omitempty"`
	CPUs      int    `json:"logical_cpus"`
}

type BootstrapMeasurement struct {
	Repetition  int    `json:"repetition"`
	ElapsedNS   int64  `json:"elapsed_ns"`
	AllocBytes  uint64 `json:"alloc_bytes"`
	Allocs      uint64 `json:"allocs"`
	OutputLevel int    `json:"output_level"`
}

type ExperimentResult struct {
	SchemaVersion string                 `json:"schema_version"`
	Timestamp     time.Time              `json:"timestamp"`
	Primary       RepositoryMetadata     `json:"primary_repository"`
	Lattigo       RepositoryMetadata     `json:"lattigo_repository"`
	Environment   EnvironmentMetadata    `json:"environment"`
	Config        BootstrapConfig        `json:"config"`
	Parameters    ExperimentParameters   `json:"effective_parameters"`
	Warmup        int                    `json:"warmup"`
	Repetitions   int                    `json:"repetitions"`
	Measurements  []BootstrapMeasurement `json:"measurements"`
}

type ExperimentParameters struct {
	ResidualLogN      int    `json:"residual_log_n"`
	ResidualMaxLevel  int    `json:"residual_max_level"`
	BootstrapLogN     int    `json:"bootstrap_log_n"`
	BootstrapMaxLevel int    `json:"bootstrap_max_level"`
	BootstrapLogSlots int    `json:"bootstrap_log_slots"`
	BootstrapLogQ     []int  `json:"bootstrap_log_q"`
	BootstrapLogP     []int  `json:"bootstrap_log_p"`
	CircuitOrder      string `json:"circuit_order"`
}

func RunExperiment(cfg BootstrapConfig, primaryRoot, backendRoot string) (ExperimentResult, error) {
	if cfg.Repetitions <= 0 {
		return ExperimentResult{}, fmt.Errorf("repetitions must be positive")
	}
	if cfg.Warmup < 0 {
		return ExperimentResult{}, fmt.Errorf("warmup cannot be negative")
	}

	params, btpParams, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return ExperimentResult{}, err
	}

	// The key is generated from the residual/input parameters. Standard uses it
	// for normal evaluation keys; Fast consumes the same public construction
	// path through its backend compatibility adapter.
	sk := rlwe.NewKeyGenerator(params).GenSecretKeyNew()
	evalKeys, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return ExperimentResult{}, fmt.Errorf("generate evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(btpParams, evalKeys)
	if err != nil {
		return ExperimentResult{}, fmt.Errorf("construct bootstrap evaluator: %w", err)
	}

	for i := 0; i < cfg.Warmup; i++ {
		if err := runBootstrap(eval, params, btpParams); err != nil {
			return ExperimentResult{}, fmt.Errorf("warm-up %d: %w", i+1, err)
		}
	}

	result := ExperimentResult{
		SchemaVersion: "exp-000.v1",
		Timestamp:     time.Now().UTC(),
		Primary:       gitMetadata(primaryRoot),
		Lattigo:       gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			CPU:       cpuModel(),
			CPUs:      runtime.NumCPU(),
		},
		Config:       cfg,
		Parameters:   parameterMetadata(params, btpParams),
		Warmup:       cfg.Warmup,
		Repetitions:  cfg.Repetitions,
		Measurements: make([]BootstrapMeasurement, 0, cfg.Repetitions),
	}

	for i := 0; i < cfg.Repetitions; i++ {
		ct := reproducibleInput(params, btpParams)
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		start := time.Now()
		out, err := eval.Bootstrap(ct)
		elapsed := time.Since(start)
		runtime.ReadMemStats(&after)
		if err != nil {
			return ExperimentResult{}, fmt.Errorf("bootstrap repetition %d: %w", i+1, err)
		}
		result.Measurements = append(result.Measurements, BootstrapMeasurement{
			Repetition:  i + 1,
			ElapsedNS:   elapsed.Nanoseconds(),
			AllocBytes:  after.TotalAlloc - before.TotalAlloc,
			Allocs:      after.Mallocs - before.Mallocs,
			OutputLevel: out.Level(),
		})
	}
	return result, nil
}

func runBootstrap(eval bootstrapping.Bootstrapper, params ckks.Parameters, btpParams bootstrapping.Parameters) error {
	_, err := eval.Bootstrap(reproducibleInput(params, btpParams))
	return err
}

func reproducibleInput(params ckks.Parameters, btpParams bootstrapping.Parameters) *rlwe.Ciphertext {
	values := reproducibleValues(params.MaxSlots())
	encoder := ckks.NewEncoder(params)
	pt := ckks.NewPlaintext(params, 0)
	pt.IsNTT = true
	pt.IsMontgomery = false
	pt.LogDimensions = ring.Dimensions{Cols: btpParams.LogMaxSlots()}
	if err := encoder.Encode(values, pt); err != nil {
		panic(fmt.Sprintf("encode reproducible input: %v", err))
	}
	ct := ckks.NewCiphertext(params, 1, 0)
	*ct.MetaData = *pt.MetaData
	ct.Value[0].Copy(pt.Value)
	ct.Value[1].Zero()
	ct.IsNTT = pt.IsNTT
	ct.IsMontgomery = pt.IsMontgomery
	return ct
}

func reproducibleValues(slotCount int) []complex128 {
	values := make([]complex128, slotCount)
	for i := range values {
		values[i] = complex(float64((i%7)-3)/16, float64((i%5)-2)/32)
	}
	return values
}

func parameterMetadata(params ckks.Parameters, btp bootstrapping.Parameters) ExperimentParameters {
	logQ := make([]int, len(btp.BootstrappingParameters.Q()))
	for i, q := range btp.BootstrappingParameters.Q() {
		logQ[i] = bitLen(q)
	}
	logP := make([]int, len(btp.BootstrappingParameters.P()))
	for i, p := range btp.BootstrappingParameters.P() {
		logP[i] = bitLen(p)
	}
	return ExperimentParameters{
		ResidualLogN:      params.LogN(),
		ResidualMaxLevel:  params.MaxLevel(),
		BootstrapLogN:     btp.BootstrappingParameters.LogN(),
		BootstrapMaxLevel: btp.BootstrappingParameters.MaxLevel(),
		BootstrapLogSlots: btp.LogMaxSlots(),
		BootstrapLogQ:     logQ,
		BootstrapLogP:     logP,
		CircuitOrder:      "ModUpThenEncode",
	}
}

func bitLen(value uint64) int {
	bits := 0
	for value > 0 {
		value >>= 1
		bits++
	}
	return bits
}

func gitMetadata(path string) RepositoryMetadata {
	metadata := RepositoryMetadata{Path: path}
	metadata.Commit = gitOutput(path, "rev-parse", "HEAD")
	metadata.Ref = gitOutput(path, "branch", "--show-current")
	status := gitOutput(path, "status", "--porcelain")
	metadata.Dirty = status != ""
	return metadata
}

func gitOutput(path string, args ...string) string {
	command := exec.Command("git", append([]string{"-C", path}, args...)...)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func cpuModel() string {
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}

func WriteExperimentResult(result ExperimentResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	if path == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create result directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	return nil
}
