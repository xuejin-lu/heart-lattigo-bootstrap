package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const correctnessThreshold = 1e-2

type CorrectnessValue struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type CorrectnessWorkload struct {
	Identifier   string `json:"identifier"`
	Formula      string `json:"formula"`
	LogicalSlots int    `json:"logical_slots"`
}

type CorrectnessOutputMetadata struct {
	Level        int     `json:"level"`
	ScaleLog2    float64 `json:"scale_log2"`
	IsNTT        bool    `json:"is_ntt"`
	IsMontgomery bool    `json:"is_montgomery"`
}

type CorrectnessMetrics struct {
	SampleCount          int      `json:"sample_count"`
	MaxAbsComplex        float64  `json:"max_abs_complex"`
	MeanAbsComplex       float64  `json:"mean_abs_complex"`
	RMSEComplex          float64  `json:"rmse_complex"`
	MaxAbsReal           float64  `json:"max_abs_real"`
	MaxAbsImag           float64  `json:"max_abs_imag"`
	MaxComponentAbs      float64  `json:"max_component_abs"`
	MaxAbsComplexIndices []int    `json:"max_abs_complex_indices"`
	MaxComponentIndices  []int    `json:"max_component_indices"`
	MaxAbsRealIndices    []int    `json:"max_abs_real_indices"`
	MaxAbsImagIndices    []int    `json:"max_abs_imag_indices"`
	PrecisionBits        *float64 `json:"precision_bits"`
	PrecisionBitsReason  string   `json:"precision_bits_reason,omitempty"`
	PassThreshold        bool     `json:"pass_threshold"`
	Threshold            float64  `json:"threshold"`
}

type CorrectnessResult struct {
	SchemaVersion string                    `json:"schema_version"`
	Timestamp     time.Time                 `json:"timestamp"`
	Primary       RepositoryMetadata        `json:"primary_repository"`
	Lattigo       RepositoryMetadata        `json:"lattigo_repository"`
	Environment   EnvironmentMetadata       `json:"environment"`
	Config        BootstrapConfig           `json:"config"`
	Parameters    ExperimentParameters      `json:"effective_parameters"`
	Workload      CorrectnessWorkload       `json:"workload"`
	Warmup        int                       `json:"warmup"`
	Repetitions   int                       `json:"repetitions"`
	Input         []CorrectnessValue        `json:"input"`
	Decoded       CorrectnessDecoded        `json:"decoded"`
	Errors        CorrectnessErrors         `json:"errors"`
	OutputMeta    CorrectnessOutputMetadata `json:"output_metadata"`
}

type CorrectnessDecoded struct {
	GeneratedSecret []CorrectnessValue `json:"generated_secret"`
	ZeroSecret      []CorrectnessValue `json:"zero_secret"`
}

type CorrectnessErrors struct {
	GeneratedSecretVsInput CorrectnessMetrics `json:"generated_secret_vs_input"`
	ZeroSecretVsInput      CorrectnessMetrics `json:"zero_secret_vs_input"`
}

func correctnessValues(values []complex128) []CorrectnessValue {
	result := make([]CorrectnessValue, len(values))
	for i, value := range values {
		result[i] = CorrectnessValue{Real: real(value), Imag: imag(value)}
	}
	return result
}

func zeroSecret(params ckks.Parameters) *rlwe.SecretKey {
	sk := rlwe.NewKeyGenerator(params).GenSecretKeyNew()
	for component := range sk.Value.Q.Coeffs {
		for coefficient := range sk.Value.Q.Coeffs[component] {
			sk.Value.Q.Coeffs[component][coefficient] = 0
		}
	}
	for component := range sk.Value.P.Coeffs {
		for coefficient := range sk.Value.P.Coeffs[component] {
			sk.Value.P.Coeffs[component][coefficient] = 0
		}
	}
	return sk
}

func decodeWithSecret(params ckks.Parameters, output *rlwe.Ciphertext, sk *rlwe.SecretKey) ([]complex128, error) {
	plaintext := rlwe.NewDecryptor(params, sk).DecryptNew(output)
	decoded := make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(plaintext, decoded); err != nil {
		return nil, fmt.Errorf("ordinary decrypt-and-decode: %w", err)
	}
	return decoded, nil
}

func compareComplexVectors(reference, actual []complex128, threshold float64) (CorrectnessMetrics, error) {
	if len(reference) != len(actual) {
		return CorrectnessMetrics{}, fmt.Errorf("vector length mismatch: reference=%d actual=%d", len(reference), len(actual))
	}
	if len(reference) == 0 {
		return CorrectnessMetrics{}, fmt.Errorf("cannot compare empty vectors")
	}

	metrics := CorrectnessMetrics{SampleCount: len(reference), Threshold: threshold}
	var squaredSum float64
	for i := range reference {
		delta := actual[i] - reference[i]
		absComplex := math.Hypot(real(delta), imag(delta))
		absReal := math.Abs(real(delta))
		absImag := math.Abs(imag(delta))
		metrics.MeanAbsComplex += absComplex
		squaredSum += absComplex * absComplex
		if absComplex > metrics.MaxAbsComplex {
			metrics.MaxAbsComplex = absComplex
			metrics.MaxAbsComplexIndices = []int{i}
		} else if absComplex == metrics.MaxAbsComplex {
			metrics.MaxAbsComplexIndices = append(metrics.MaxAbsComplexIndices, i)
		}
		if absReal > metrics.MaxAbsReal {
			metrics.MaxAbsReal = absReal
			metrics.MaxAbsRealIndices = []int{i}
		} else if absReal == metrics.MaxAbsReal {
			metrics.MaxAbsRealIndices = append(metrics.MaxAbsRealIndices, i)
		}
		if absImag > metrics.MaxAbsImag {
			metrics.MaxAbsImag = absImag
			metrics.MaxAbsImagIndices = []int{i}
		} else if absImag == metrics.MaxAbsImag {
			metrics.MaxAbsImagIndices = append(metrics.MaxAbsImagIndices, i)
		}
	}
	metrics.MeanAbsComplex /= float64(len(reference))
	metrics.RMSEComplex = math.Sqrt(squaredSum / float64(len(reference)))
	metrics.MaxComponentAbs = math.Max(metrics.MaxAbsReal, metrics.MaxAbsImag)
	if metrics.MaxAbsReal >= metrics.MaxAbsImag {
		metrics.MaxComponentIndices = append(metrics.MaxComponentIndices, metrics.MaxAbsRealIndices...)
	}
	if metrics.MaxAbsImag >= metrics.MaxAbsReal {
		metrics.MaxComponentIndices = append(metrics.MaxComponentIndices, metrics.MaxAbsImagIndices...)
	}
	if metrics.MaxAbsComplex == 0 {
		metrics.PrecisionBitsReason = "max_abs_complex is exactly zero"
	} else {
		precisionBits := -math.Log2(metrics.MaxAbsComplex)
		metrics.PrecisionBits = &precisionBits
	}
	metrics.PassThreshold = metrics.MaxAbsReal <= threshold && metrics.MaxAbsImag <= threshold
	return metrics, nil
}

func RunCorrectnessExperiment(cfg BootstrapConfig, primaryRoot, backendRoot string) (CorrectnessResult, error) {
	residual, btpParams, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return CorrectnessResult{}, err
	}
	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	evalKeys, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("generate evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(btpParams, evalKeys)
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("construct bootstrap evaluator: %w", err)
	}

	input := reproducibleValues(residual.MaxSlots())
	output, err := eval.Bootstrap(reproducibleInput(residual, btpParams))
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("bootstrap: %w", err)
	}

	// Both views deliberately use the ordinary public recovery path. The
	// interpretation of which view is authoritative belongs in summaries; the
	// execution path never branches on backend identity.
	generatedDecoded, err := decodeWithSecret(residual, output, sk)
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("generated-secret view: %w", err)
	}
	zeroDecoded, err := decodeWithSecret(residual, output, zeroSecret(residual))
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("zero-secret view: %w", err)
	}
	generatedMetrics, err := compareComplexVectors(input, generatedDecoded, correctnessThreshold)
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("generated-secret metrics: %w", err)
	}
	zeroMetrics, err := compareComplexVectors(input, zeroDecoded, correctnessThreshold)
	if err != nil {
		return CorrectnessResult{}, fmt.Errorf("zero-secret metrics: %w", err)
	}

	return CorrectnessResult{
		SchemaVersion: "exp-002c.v1",
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
		Config:     cfg,
		Parameters: parameterMetadata(residual, btpParams),
		Workload: CorrectnessWorkload{
			Identifier:   "reproducibleInput.v1",
			Formula:      "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i",
			LogicalSlots: len(input),
		},
		Warmup:      0,
		Repetitions: 1,
		Input:       correctnessValues(input),
		Decoded: CorrectnessDecoded{
			GeneratedSecret: correctnessValues(generatedDecoded),
			ZeroSecret:      correctnessValues(zeroDecoded),
		},
		Errors: CorrectnessErrors{
			GeneratedSecretVsInput: generatedMetrics,
			ZeroSecretVsInput:      zeroMetrics,
		},
		OutputMeta: CorrectnessOutputMetadata{
			Level:        output.Level(),
			ScaleLog2:    output.Scale.Log2(),
			IsNTT:        output.IsNTT,
			IsMontgomery: output.IsMontgomery,
		},
	}, nil
}

func WriteCorrectnessResult(result CorrectnessResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode correctness result: %w", err)
	}
	if path == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create correctness result directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write correctness result: %w", err)
	}
	return nil
}
