package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type StageMeasurement struct {
	Repetition       int     `json:"repetition"`
	Stage            string  `json:"stage"`
	Present          bool    `json:"present"`
	ElapsedNS        int64   `json:"elapsed_ns"`
	AllocBytes       uint64  `json:"alloc_bytes"`
	Allocs           uint64  `json:"allocs"`
	InputBranches    int     `json:"input_branches"`
	OutputBranches   int     `json:"output_branches"`
	InputLevel       int     `json:"input_level"`
	OutputLevel      int     `json:"output_level"`
	InputScaleLog2   float64 `json:"input_scale_log2"`
	OutputScaleLog2  float64 `json:"output_scale_log2"`
	InputIsNTT       bool    `json:"input_is_ntt"`
	OutputIsNTT      bool    `json:"output_is_ntt"`
	InputMontgomery  bool    `json:"input_is_montgomery"`
	OutputMontgomery bool    `json:"output_is_montgomery"`
}

type StagePipelineCheck struct {
	Repetition                int     `json:"repetition"`
	StagedOutputLevel         int     `json:"staged_output_level"`
	FullBootstrapLevel        int     `json:"full_bootstrap_output_level"`
	StagedOutputScaleLog2     float64 `json:"staged_output_scale_log2"`
	FullBootstrapScaleLog2    float64 `json:"full_bootstrap_output_scale_log2"`
	StagedIsNTT               bool    `json:"staged_is_ntt"`
	FullBootstrapIsNTT        bool    `json:"full_bootstrap_is_ntt"`
	StagedIsMontgomery        bool    `json:"staged_is_montgomery"`
	FullBootstrapIsMontgomery bool    `json:"full_bootstrap_is_montgomery"`
	LevelMatch                bool    `json:"level_match"`
	ScaleMatch                bool    `json:"scale_match"`
	DomainMatch               bool    `json:"domain_match"`
}

type StageExperimentResult struct {
	SchemaVersion string               `json:"schema_version"`
	Timestamp     time.Time            `json:"timestamp"`
	Primary       RepositoryMetadata   `json:"primary_repository"`
	Lattigo       RepositoryMetadata   `json:"lattigo_repository"`
	Environment   EnvironmentMetadata  `json:"environment"`
	Config        BootstrapConfig      `json:"config"`
	Parameters    ExperimentParameters `json:"effective_parameters"`
	Warmup        int                  `json:"warmup"`
	Repetitions   int                  `json:"repetitions"`
	Measurements  []StageMeasurement   `json:"measurements"`
	Correctness   []StagePipelineCheck `json:"correctness"`
}

type stageState struct {
	Branches   int
	Level      int
	ScaleLog2  float64
	IsNTT      bool
	Montgomery bool
}

func stageStateFromCiphertexts(cts []rlwe.Ciphertext) stageState {
	if len(cts) == 0 {
		return stageState{Level: -1}
	}
	ct := &cts[0]
	return stageState{Branches: len(cts), Level: ct.Level(), ScaleLog2: ct.Scale.Log2(), IsNTT: ct.IsNTT, Montgomery: ct.IsMontgomery}
}

func stageStateFromPointers(cts ...*rlwe.Ciphertext) stageState {
	branches := 0
	var first *rlwe.Ciphertext
	for _, ct := range cts {
		if ct != nil {
			branches++
			if first == nil {
				first = ct
			}
		}
	}
	if first == nil {
		return stageState{Level: -1}
	}
	return stageState{Branches: branches, Level: first.Level(), ScaleLog2: first.Scale.Log2(), IsNTT: first.IsNTT, Montgomery: first.IsMontgomery}
}

func stageMeasurementFromStats(stage string, repetition int, input, output stageState, elapsed time.Duration, before, after runtime.MemStats) StageMeasurement {
	return StageMeasurement{
		Repetition:       repetition,
		Stage:            stage,
		Present:          true,
		ElapsedNS:        elapsed.Nanoseconds(),
		AllocBytes:       after.TotalAlloc - before.TotalAlloc,
		Allocs:           after.Mallocs - before.Mallocs,
		InputBranches:    input.Branches,
		OutputBranches:   output.Branches,
		InputLevel:       input.Level,
		OutputLevel:      output.Level,
		InputScaleLog2:   input.ScaleLog2,
		OutputScaleLog2:  output.ScaleLog2,
		InputIsNTT:       input.IsNTT,
		OutputIsNTT:      output.IsNTT,
		InputMontgomery:  input.Montgomery,
		OutputMontgomery: output.Montgomery,
	}
}

func measureStageCall(stage string, repetition int, input stageState, call func() error, output func() stageState) (StageMeasurement, error) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	err := call()
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	if err != nil {
		return StageMeasurement{}, err
	}
	return stageMeasurementFromStats(stage, repetition, input, output(), elapsed, before, after), nil
}

func runStagePipeline(eval *bootstrapping.Evaluator, residual ckks.Parameters, btp bootstrapping.Parameters, repetition int, measured bool) ([]StageMeasurement, *rlwe.Ciphertext, error) {
	base := reproducibleInput(residual, btp)
	cts := []rlwe.Ciphertext{*base}
	measurements := make([]StageMeasurement, 0, 9)

	record := func(stage string, input stageState, call func() error, output func() stageState) error {
		if measured {
			measurement, err := measureStageCall(stage, repetition, input, call, output)
			if err != nil {
				return err
			}
			measurements = append(measurements, measurement)
			return nil
		}
		return call()
	}

	// The packing context types are intentionally private to the backend
	// package. Short declaration keeps them opaque while allowing them to be
	// passed to the next public stage method without frontend knowledge.
	packInput := stageStateFromCiphertexts(cts)
	var packBefore, packAfter runtime.MemStats
	var packStart time.Time
	if measured {
		runtime.ReadMemStats(&packBefore)
		packStart = time.Now()
	}
	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2(cts)
	packElapsed := time.Since(packStart)
	if measured {
		runtime.ReadMemStats(&packAfter)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("stage pack_and_switch_n1_to_n2: %w", err)
	}
	if measured {
		measurements = append(measurements, stageMeasurementFromStats("pack_and_switch_n1_to_n2", repetition, packInput, stageStateFromCiphertexts(packed), packElapsed, packBefore, packAfter))
	}

	var scaled *rlwe.Ciphertext
	input := stageStateFromCiphertexts(packed)
	if err := record("scale_down", input, func() error {
		var scaleErr error
		scaled, _, scaleErr = eval.ScaleDown(&packed[0])
		return scaleErr
	}, func() stageState { return stageStateFromPointers(scaled) }); err != nil {
		return nil, nil, fmt.Errorf("stage scale_down: %w", err)
	}
	packed[0] = *scaled

	var modUp *rlwe.Ciphertext
	input = stageStateFromCiphertexts(packed)
	if err := record("mod_up", input, func() error {
		var modUpErr error
		modUp, modUpErr = eval.ModUp(&packed[0])
		return modUpErr
	}, func() stageState { return stageStateFromPointers(modUp) }); err != nil {
		return nil, nil, fmt.Errorf("stage mod_up: %w", err)
	}
	packed[0] = *modUp

	var ctReal, ctImag *rlwe.Ciphertext
	input = stageStateFromCiphertexts(packed)
	if err := record("coeffs_to_slots", input, func() error {
		var dftErr error
		ctReal, ctImag, dftErr = eval.CoeffsToSlots(&packed[0])
		return dftErr
	}, func() stageState { return stageStateFromPointers(ctReal, ctImag) }); err != nil {
		return nil, nil, fmt.Errorf("stage coeffs_to_slots: %w", err)
	}

	input = stageStateFromPointers(ctReal)
	if err := record("eval_mod_real", input, func() error {
		var evalErr error
		ctReal, evalErr = eval.EvalMod(ctReal)
		return evalErr
	}, func() stageState { return stageStateFromPointers(ctReal) }); err != nil {
		return nil, nil, fmt.Errorf("stage eval_mod_real: %w", err)
	}

	if ctImag == nil {
		if measured {
			measurements = append(measurements, StageMeasurement{Repetition: repetition, Stage: "eval_mod_imag", InputLevel: -1, OutputLevel: -1})
		}
	} else {
		input = stageStateFromPointers(ctImag)
		if err := record("eval_mod_imag", input, func() error {
			var evalErr error
			ctImag, evalErr = eval.EvalMod(ctImag)
			return evalErr
		}, func() stageState { return stageStateFromPointers(ctImag) }); err != nil {
			return nil, nil, fmt.Errorf("stage eval_mod_imag: %w", err)
		}
	}

	var ctOut *rlwe.Ciphertext
	input = stageStateFromPointers(ctReal, ctImag)
	if err := record("slots_to_coeffs", input, func() error {
		var dftErr error
		ctOut, dftErr = eval.SlotsToCoeffs(ctReal, ctImag)
		return dftErr
	}, func() stageState { return stageStateFromPointers(ctOut) }); err != nil {
		return nil, nil, fmt.Errorf("stage slots_to_coeffs: %w", err)
	}

	var unpacked []rlwe.Ciphertext
	input = stageStateFromPointers(ctOut)
	if err := record("unpack_and_switch_n2_to_n1", input, func() error {
		var unpackErr error
		unpacked, unpackErr = eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*ctOut}, ctxtN1, ctxtN2)
		return unpackErr
	}, func() stageState { return stageStateFromCiphertexts(unpacked) }); err != nil {
		return nil, nil, fmt.Errorf("stage unpack_and_switch_n2_to_n1: %w", err)
	}
	if len(unpacked) != 1 {
		return nil, nil, fmt.Errorf("stage unpack_and_switch_n2_to_n1 returned %d ciphertexts, want 1", len(unpacked))
	}
	return measurements, &unpacked[0], nil
}

func RunStageExperiment(cfg BootstrapConfig, primaryRoot, backendRoot string) (StageExperimentResult, error) {
	if cfg.Repetitions <= 0 {
		return StageExperimentResult{}, fmt.Errorf("repetitions must be positive")
	}
	if cfg.Warmup < 0 {
		return StageExperimentResult{}, fmt.Errorf("warmup cannot be negative")
	}
	residual, btpParams, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return StageExperimentResult{}, err
	}
	sk := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	evalKeys, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return StageExperimentResult{}, fmt.Errorf("generate evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(btpParams, evalKeys)
	if err != nil {
		return StageExperimentResult{}, fmt.Errorf("construct bootstrap evaluator: %w", err)
	}

	for i := 0; i < cfg.Warmup; i++ {
		if _, _, err := runStagePipeline(eval, residual, btpParams, i+1, false); err != nil {
			return StageExperimentResult{}, fmt.Errorf("stage warm-up %d: %w", i+1, err)
		}
		if _, err := eval.Bootstrap(reproducibleInput(residual, btpParams)); err != nil {
			return StageExperimentResult{}, fmt.Errorf("full-bootstrap warm-up %d: %w", i+1, err)
		}
	}

	result := StageExperimentResult{
		SchemaVersion: "exp-002.v1",
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
		Parameters:   parameterMetadata(residual, btpParams),
		Warmup:       cfg.Warmup,
		Repetitions:  cfg.Repetitions,
		Measurements: make([]StageMeasurement, 0, cfg.Repetitions*9),
		Correctness:  make([]StagePipelineCheck, 0, cfg.Repetitions),
	}

	for i := 0; i < cfg.Repetitions; i++ {
		repetition := i + 1
		measurements, staged, err := runStagePipeline(eval, residual, btpParams, repetition, true)
		if err != nil {
			return StageExperimentResult{}, fmt.Errorf("staged pipeline repetition %d: %w", repetition, err)
		}
		result.Measurements = append(result.Measurements, measurements...)

		fullInput := reproducibleInput(residual, btpParams)
		fullInputState := stageStateFromPointers(fullInput)
		var fullOutput *rlwe.Ciphertext
		fullMeasurement, err := measureStageCall("full_bootstrap", repetition, fullInputState, func() error {
			var bootstrapErr error
			fullOutput, bootstrapErr = eval.Bootstrap(fullInput)
			return bootstrapErr
		}, func() stageState { return stageStateFromPointers(fullOutput) })
		if err != nil {
			return StageExperimentResult{}, fmt.Errorf("full bootstrap repetition %d: %w", repetition, err)
		}
		result.Measurements = append(result.Measurements, fullMeasurement)

		check := StagePipelineCheck{
			Repetition:                repetition,
			StagedOutputLevel:         staged.Level(),
			FullBootstrapLevel:        fullOutput.Level(),
			StagedOutputScaleLog2:     staged.Scale.Log2(),
			FullBootstrapScaleLog2:    fullOutput.Scale.Log2(),
			StagedIsNTT:               staged.IsNTT,
			FullBootstrapIsNTT:        fullOutput.IsNTT,
			StagedIsMontgomery:        staged.IsMontgomery,
			FullBootstrapIsMontgomery: fullOutput.IsMontgomery,
		}
		check.LevelMatch = check.StagedOutputLevel == check.FullBootstrapLevel
		check.ScaleMatch = staged.Scale.Equal(fullOutput.Scale)
		check.DomainMatch = check.StagedIsNTT == fullOutput.IsNTT && check.StagedIsMontgomery == fullOutput.IsMontgomery
		result.Correctness = append(result.Correctness, check)
		if !check.LevelMatch || !check.ScaleMatch || !check.DomainMatch {
			return StageExperimentResult{}, fmt.Errorf("repetition %d staged/full output metadata mismatch", repetition)
		}
	}
	return result, nil
}

func WriteStageExperimentResult(result StageExperimentResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode stage result: %w", err)
	}
	if path == "" {
		_, err = os.Stdout.Write(append(data, '\n'))
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write stage result: %w", err)
	}
	return nil
}
