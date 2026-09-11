package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type DiagnosticCiphertextMetadata struct {
	Branches     int     `json:"branches"`
	Level        int     `json:"level"`
	ScaleLog2    float64 `json:"scale_log2"`
	IsNTT        bool    `json:"is_ntt"`
	IsMontgomery bool    `json:"is_montgomery"`
}

type DiagnosticStageRecord struct {
	Stage                  string                         `json:"stage"`
	Quantity               string                         `json:"quantity"`
	Present                bool                           `json:"present"`
	Method                 string                         `json:"method"`
	Metadata               []DiagnosticCiphertextMetadata `json:"metadata"`
	ErrScaleLog2           *float64                       `json:"err_scale_log2"`
	GeneratedSecretDecoded [][]CorrectnessValue           `json:"generated_secret_decoded"`
	ZeroSecretDecoded      [][]CorrectnessValue           `json:"zero_secret_decoded"`
	DecodeErrors           []string                       `json:"decode_errors,omitempty"`
}

type DiagnosticTraceContext struct {
	ResidualLogN                int    `json:"residual_log_n"`
	BootstrapLogN               int    `json:"bootstrap_log_n"`
	BootstrapLogSlots           int    `json:"bootstrap_log_slots"`
	TraceGap                    int    `json:"trace_gap"`
	TraceAutomorphismIterations int    `json:"trace_automorphism_iterations"`
	PackingLogPackCTs           int    `json:"packing_log_pack_cts"`
	PackingWork                 string `json:"packing_work"`
	RingDegreeSwitch            string `json:"ring_degree_switch"`
	TraceWork                   string `json:"trace_work"`
}

type DiagnosticResult struct {
	SchemaVersion string                    `json:"schema_version"`
	Timestamp     time.Time                 `json:"timestamp"`
	Primary       RepositoryMetadata        `json:"primary_repository"`
	Lattigo       RepositoryMetadata        `json:"lattigo_repository"`
	Environment   EnvironmentMetadata       `json:"environment"`
	Config        BootstrapConfig           `json:"config"`
	Parameters    ExperimentParameters      `json:"effective_parameters"`
	Workload      CorrectnessWorkload       `json:"workload"`
	TraceContext  DiagnosticTraceContext    `json:"trace_context"`
	Stages        []DiagnosticStageRecord   `json:"stages"`
	OutputMeta    CorrectnessOutputMetadata `json:"output_metadata"`
}

func diagnosticMetadata(cts []*rlwe.Ciphertext) []DiagnosticCiphertextMetadata {
	metadata := make([]DiagnosticCiphertextMetadata, len(cts))
	for i, ct := range cts {
		if ct == nil {
			continue
		}
		metadata[i] = DiagnosticCiphertextMetadata{
			Branches:     len(cts),
			Level:        ct.Level(),
			ScaleLog2:    ct.Scale.Log2(),
			IsNTT:        ct.IsNTT,
			IsMontgomery: ct.IsMontgomery,
		}
	}
	return metadata
}

func diagnosticDecode(params ckks.Parameters, ct *rlwe.Ciphertext, sk *rlwe.SecretKey) (values []complex128, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("ordinary decrypt-and-CKKS-decode panic: %v", recovered)
		}
	}()
	return decodeWithSecret(params, ct, sk)
}

func diagnosticRecord(params ckks.Parameters, stage, quantity string, cts []*rlwe.Ciphertext, generatedSecret, zeroSecretKey *rlwe.SecretKey, errScale *rlwe.Scale) DiagnosticStageRecord {
	record := DiagnosticStageRecord{
		Stage:                  stage,
		Quantity:               quantity,
		Present:                true,
		Method:                 "ordinary_decrypt_and_ckks_decode",
		Metadata:               diagnosticMetadata(cts),
		GeneratedSecretDecoded: make([][]CorrectnessValue, len(cts)),
		ZeroSecretDecoded:      make([][]CorrectnessValue, len(cts)),
	}
	if errScale != nil {
		value := errScale.Log2()
		record.ErrScaleLog2 = &value
	}
	for i, ct := range cts {
		generated, generatedErr := diagnosticDecode(params, ct, generatedSecret)
		zero, zeroErr := diagnosticDecode(params, ct, zeroSecretKey)
		if generatedErr != nil {
			record.DecodeErrors = append(record.DecodeErrors, fmt.Sprintf("branch %d generated_secret: %v", i, generatedErr))
		} else {
			record.GeneratedSecretDecoded[i] = correctnessValues(generated)
		}
		if zeroErr != nil {
			record.DecodeErrors = append(record.DecodeErrors, fmt.Sprintf("branch %d zero_secret: %v", i, zeroErr))
		} else {
			record.ZeroSecretDecoded[i] = correctnessValues(zero)
		}
	}
	if len(record.DecodeErrors) > 0 {
		record.Method = "not_directly_comparable"
	}
	return record
}

func diagnosticTraceContext(residual ckks.Parameters, btp bootstrapping.Parameters) DiagnosticTraceContext {
	logN := btp.LogMaxSlots()
	traceGap := 1 << (btp.BootstrappingParameters.LogN() - logN - 1)
	traceIterations := btp.BootstrappingParameters.LogN() - 1 - logN
	packingLogPackCTs := btp.LogMaxSlots() - logN
	return DiagnosticTraceContext{
		ResidualLogN:                residual.LogN(),
		BootstrapLogN:               btp.BootstrappingParameters.LogN(),
		BootstrapLogSlots:           logN,
		TraceGap:                    traceGap,
		TraceAutomorphismIterations: traceIterations,
		PackingLogPackCTs:           packingLogPackCTs,
		PackingWork:                 "TRIVIAL: full slots and one ciphertext imply logPackCTs=0",
		RingDegreeSwitch:            "TRIVIAL: residual N equals bootstrap N",
		TraceWork:                   "TRIVIAL: Trace gap=1 and automorphism iteration count=0",
	}
}

func RunDiagnosticExperiment(cfg BootstrapConfig, primaryRoot, backendRoot string) (DiagnosticResult, error) {
	residual, btpParams, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return DiagnosticResult{}, err
	}
	generatedSecret := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	evalKeys, bootstrapSecret, err := btpParams.GenEvaluationKeys(generatedSecret)
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("generate evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(btpParams, evalKeys)
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("construct bootstrap evaluator: %w", err)
	}
	zeroSecretKey := zeroSecret(residual)
	zeroBootstrapSecretKey := zeroSecret(btpParams.BootstrappingParameters)

	input := reproducibleInput(residual, btpParams)
	inputRecord := diagnosticRecord(residual, "input_sanity", "deterministic full-slot input", []*rlwe.Ciphertext{input}, generatedSecret, zeroSecretKey, nil)
	stages := []DiagnosticStageRecord{inputRecord}

	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input})
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("PackAndSwitchN1ToN2: %w", err)
	}
	stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "pack_and_switch_n1_to_n2", "ciphertext boundary", []*rlwe.Ciphertext{&packed[0]}, bootstrapSecret, zeroBootstrapSecretKey, nil))

	scaled, errScale, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("ScaleDown: %w", err)
	}
	packed[0] = *scaled
	stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "scale_down", "ciphertext boundary", []*rlwe.Ciphertext{scaled}, bootstrapSecret, zeroBootstrapSecretKey, errScale))

	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("ModUp: %w", err)
	}
	packed[0] = *modUp
	stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "mod_up", "ciphertext boundary including Trace", []*rlwe.Ciphertext{modUp}, bootstrapSecret, zeroBootstrapSecretKey, nil))

	ctReal, ctImag, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("CoeffsToSlots: %w", err)
	}
	c2s := []*rlwe.Ciphertext{ctReal}
	if ctImag != nil {
		c2s = append(c2s, ctImag)
	}
	stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "coeffs_to_slots", "real and imaginary DFT branches", c2s, bootstrapSecret, zeroBootstrapSecretKey, nil))

	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("EvalMod real: %w", err)
	}
	stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "eval_mod_real", "real DFT branch", []*rlwe.Ciphertext{ctReal}, bootstrapSecret, zeroBootstrapSecretKey, nil))
	if ctImag != nil {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return DiagnosticResult{}, fmt.Errorf("EvalMod imag: %w", err)
		}
		stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "eval_mod_imag", "imaginary DFT branch", []*rlwe.Ciphertext{ctImag}, bootstrapSecret, zeroBootstrapSecretKey, nil))
	}

	ctOut, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("SlotsToCoeffs: %w", err)
	}
	stages = append(stages, diagnosticRecord(btpParams.BootstrappingParameters, "slots_to_coeffs", "ciphertext boundary before public unpack", []*rlwe.Ciphertext{ctOut}, bootstrapSecret, zeroBootstrapSecretKey, nil))

	unpacked, err := eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*ctOut}, ctxtN1, ctxtN2)
	if err != nil {
		return DiagnosticResult{}, fmt.Errorf("UnpackAndSwitchN2ToN1: %w", err)
	}
	final := make([]*rlwe.Ciphertext, len(unpacked))
	for i := range unpacked {
		final[i] = &unpacked[i]
	}
	stages = append(stages, diagnosticRecord(residual, "unpack_and_switch_n2_to_n1", "final public stage output", final, generatedSecret, zeroSecretKey, nil))

	return DiagnosticResult{
		SchemaVersion: "exp-002c-diag.v1",
		Timestamp:     time.Now().UTC(),
		Primary:       gitMetadata(primaryRoot),
		Lattigo:       gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{
			GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU(),
		},
		Config: cfg, Parameters: parameterMetadata(residual, btpParams),
		Workload: CorrectnessWorkload{
			Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots(),
		},
		TraceContext: diagnosticTraceContext(residual, btpParams),
		Stages:       stages,
		OutputMeta:   CorrectnessOutputMetadata{Level: unpacked[0].Level(), ScaleLog2: unpacked[0].Scale.Log2(), IsNTT: unpacked[0].IsNTT, IsMontgomery: unpacked[0].IsMontgomery},
	}, nil
}

func WriteDiagnosticResult(result DiagnosticResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode diagnostic result: %w", err)
	}
	if path == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create diagnostic result directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write diagnostic result: %w", err)
	}
	return nil
}
