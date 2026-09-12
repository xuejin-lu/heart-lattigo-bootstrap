package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type DoubleAngleCheckpoint struct {
	Round             int                          `json:"round"`
	Checkpoint        string                       `json:"checkpoint"`
	Operation         string                       `json:"operation"`
	Level             int                          `json:"level"`
	Degree            int                          `json:"degree"`
	Scale             string                       `json:"scale"`
	ExpectedScale     string                       `json:"expected_scale,omitempty"`
	ScaleLog2Delta    float64                      `json:"scale_log2_delta,omitempty"`
	PreLevel          int                          `json:"pre_level,omitempty"`
	PostLevel         int                          `json:"post_level,omitempty"`
	ExpectedPostLevel int                          `json:"expected_post_level,omitempty"`
	DroppedModuli     []uint64                     `json:"dropped_moduli,omitempty"`
	Constant          string                       `json:"constant,omitempty"`
	Semantic          *PSGlobalMetric              `json:"semantic"`
	ErrorDelta        float64                      `json:"error_delta,omitempty"`
	ErrorRatio        float64                      `json:"error_ratio,omitempty"`
	Capacity          TargetScaleCapacityEvidence  `json:"capacity"`
	Rows              []FinalizationRowFingerprint `json:"q0_q1_rows"`
}

type DoubleAngleFinalEvidence struct {
	Level                   int                          `json:"level"`
	Degree                  int                          `json:"degree"`
	Scale                   string                       `json:"scale"`
	ScalingFactor           string                       `json:"mod1_scaling_factor"`
	ScaleLog2DeltaToScaling float64                      `json:"scale_log2_delta_to_mod1_scaling_factor"`
	Semantic                *PSGlobalMetric              `json:"semantic"`
	Capacity                TargetScaleCapacityEvidence  `json:"capacity"`
	Rows                    []FinalizationRowFingerprint `json:"q0_q1_rows"`
}

type DoubleAngleResetEvidence struct {
	ScaleBeforeReset         string                       `json:"scale_before_reset"`
	InputScale               string                       `json:"original_mod1_input_scale"`
	ScaleRatio               string                       `json:"scale_before_reset_over_input_scale"`
	ScaleAfterReset          string                       `json:"scale_after_reset"`
	LevelBefore              int                          `json:"level_before"`
	LevelAfter               int                          `json:"level_after"`
	DegreeBefore             int                          `json:"degree_before"`
	DegreeAfter              int                          `json:"degree_after"`
	RowsUnchanged            bool                         `json:"q0_q1_rows_unchanged"`
	MetadataExpectedSemantic *PSGlobalMetric              `json:"metadata_expected_semantic"`
	PerturbationFromBefore   *PSGlobalMetric              `json:"decoded_perturbation_from_before_reset"`
	Rows                     []FinalizationRowFingerprint `json:"q0_q1_rows"`
}

type DoubleAngleResult struct {
	SchemaVersion       string                       `json:"schema_version"`
	Timestamp           time.Time                    `json:"timestamp"`
	Primary             RepositoryMetadata           `json:"primary_repository"`
	Lattigo             RepositoryMetadata           `json:"lattigo_repository"`
	Environment         EnvironmentMetadata          `json:"environment"`
	Config              BootstrapConfig              `json:"config"`
	Parameters          ExperimentParameters         `json:"effective_parameters"`
	Workload            CorrectnessWorkload          `json:"workload"`
	AuthoritativeOracle string                       `json:"authoritative_poly_oracle"`
	Threshold           float64                      `json:"threshold"`
	DoubleAngleCount    int                          `json:"double_angle_count"`
	Sqrt2PiInitial      string                       `json:"sqrt2pi_initial"`
	InputScale          string                       `json:"original_mod1_input_scale"`
	P4Level             int                          `json:"p4_level"`
	P4Degree            int                          `json:"p4_degree"`
	P4Scale             string                       `json:"p4_scale"`
	P4Semantic          *PSGlobalMetric              `json:"p4_semantic"`
	P4Capacity          TargetScaleCapacityEvidence  `json:"p4_capacity"`
	P4Rows              []FinalizationRowFingerprint `json:"p4_q0_q1_rows"`
	P4Precondition      map[string]bool              `json:"p4_precondition"`
	Rounds              []DoubleAngleCheckpoint      `json:"rounds"`
	FinalBeforeReset    DoubleAngleFinalEvidence     `json:"da_f0_before_metadata_reset"`
	FinalScaleReset     DoubleAngleResetEvidence     `json:"da_f1_metadata_reset"`
	FirstFailingRound   string                       `json:"first_failing_round"`
	FirstFailingPoint   string                       `json:"first_failing_checkpoint"`
	LastPassingPoint    string                       `json:"last_passing_checkpoint"`
	FirstSupportedCause string                       `json:"first_supported_cause"`
	Validation          map[string]interface{}       `json:"validation"`
}

type doubleAngleHP struct {
	Real *big.Float
	Imag *big.Float
}

const doubleAnglePrecision = uint(256)

func newDoubleAngleFloat(value float64) *big.Float {
	return new(big.Float).SetPrec(doubleAnglePrecision).SetFloat64(value)
}

func cloneDoubleAngleFloat(value *big.Float) *big.Float {
	return new(big.Float).SetPrec(doubleAnglePrecision).Set(value)
}

func doubleAngleHPFromComplex(value complex128) doubleAngleHP {
	return doubleAngleHP{Real: newDoubleAngleFloat(real(value)), Imag: newDoubleAngleFloat(imag(value))}
}

func doubleAngleHPSquare(value doubleAngleHP) doubleAngleHP {
	rr := new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Real, value.Real)
	ii := new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Imag, value.Imag)
	ri := new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Real, value.Imag)
	return doubleAngleHP{Real: new(big.Float).SetPrec(doubleAnglePrecision).Sub(rr, ii), Imag: new(big.Float).SetPrec(doubleAnglePrecision).Add(ri, ri)}
}

func doubleAngleHPDouble(value doubleAngleHP) doubleAngleHP {
	two := newDoubleAngleFloat(2)
	return doubleAngleHP{Real: new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Real, two), Imag: new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Imag, two)}
}

func doubleAngleHPOffset(value doubleAngleHP, constant float64) doubleAngleHP {
	return doubleAngleHP{Real: new(big.Float).SetPrec(doubleAnglePrecision).Sub(value.Real, newDoubleAngleFloat(constant)), Imag: cloneDoubleAngleFloat(value.Imag)}
}

func doubleAngleHPMulReal(value doubleAngleHP, factor *big.Float) doubleAngleHP {
	return doubleAngleHP{Real: new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Real, factor), Imag: new(big.Float).SetPrec(doubleAnglePrecision).Mul(value.Imag, factor)}
}

func doubleAngleHPVector(values []complex128) []doubleAngleHP {
	result := make([]doubleAngleHP, len(values))
	for i, value := range values {
		result[i] = doubleAngleHPFromComplex(value)
	}
	return result
}

func doubleAngleHPToVector(values []doubleAngleHP) []complex128 {
	result := make([]complex128, len(values))
	for i, value := range values {
		r, _ := value.Real.Float64()
		im, _ := value.Imag.Float64()
		result[i] = complex(r, im)
	}
	return result
}

func doubleAngleHPScale(values []doubleAngleHP, factor *big.Float) []doubleAngleHP {
	result := make([]doubleAngleHP, len(values))
	for i, value := range values {
		result[i] = doubleAngleHPMulReal(value, factor)
	}
	return result
}

func doubleAngleCapacity(params ckks.Parameters, ct *rlwe.Ciphertext) (TargetScaleCapacityEvidence, error) {
	coefficientView, err := targetScaleCoefficientView(params, ct)
	if err != nil {
		return TargetScaleCapacityEvidence{}, err
	}
	data, err := targetScaleCenteredRows(params, coefficientView)
	if err != nil {
		return TargetScaleCapacityEvidence{}, err
	}
	return targetScaleCapacity(data, big.NewInt(1)), nil
}

func doubleAngleExpectedScaleString(scale rlwe.Scale) string {
	return finalizationScaleString(scale)
}

func doubleAngleScaleDelta(actual, expected rlwe.Scale) float64 {
	delta := actual.Log2Delta(expected)
	if math.IsInf(delta, 0) || math.IsNaN(delta) {
		if actual.Equal(expected) {
			return math.MaxFloat64
		}
		return 0
	}
	return delta
}

func doubleAngleCheckpoint(round int, name, operation string, ct *rlwe.Ciphertext, expected []complex128, actual []complex128, expectedScale *rlwe.Scale, previous *PSGlobalMetric, capacity TargetScaleCapacityEvidence) DoubleAngleCheckpoint {
	semantic := psGlobalMetric(expected, actual)
	checkpoint := DoubleAngleCheckpoint{Round: round, Checkpoint: name, Operation: operation, Level: ct.Level(), Degree: ct.Degree(), Scale: finalizationScaleString(ct.Scale), Semantic: semantic, Capacity: capacity, Rows: finalizationEvidence(ct).Rows}
	if expectedScale != nil {
		checkpoint.ExpectedScale = doubleAngleExpectedScaleString(*expectedScale)
		checkpoint.ScaleLog2Delta = doubleAngleScaleDelta(ct.Scale, *expectedScale)
	}
	if previous != nil {
		checkpoint.ErrorDelta = semantic.MaxComponent - previous.MaxComponent
		if previous.MaxComponent > 0 {
			checkpoint.ErrorRatio = semantic.MaxComponent / previous.MaxComponent
		}
	}
	return checkpoint
}

func doubleAngleDroppedScale(scale rlwe.Scale, level int, params ckks.Parameters) (rlwe.Scale, int, []uint64) {
	result := scale
	dropped := make([]uint64, 0, params.LevelsConsumedPerRescaling())
	for i := 0; i < params.LevelsConsumedPerRescaling(); i++ {
		modulus := params.Q()[level-i]
		dropped = append(dropped, modulus)
		result = result.Div(rlwe.NewScale(modulus))
	}
	return result, level - params.LevelsConsumedPerRescaling(), dropped
}

func doubleAngleExpectedP4Rows() []FinalizationRowFingerprint {
	return []FinalizationRowFingerprint{
		{Component: 0, Limb: 0, Length: 8192, SHA256: "8dc052eed24139c55c8472ec44701be29a3ee15177262243b773d4e71a3ce65d"},
		{Component: 0, Limb: 1, Length: 8192, SHA256: "64b0bd95c83a93ab48fdff695261d43bd834d053c45fadfff895429b492ce53b"},
		{Component: 1, Limb: 0, Length: 8192, SHA256: "de2f256064a0af797747c2b97505dc0b9f3df0de4f489eac731c23ae9ca9cc31"},
		{Component: 1, Limb: 1, Length: 8192, SHA256: "de2f256064a0af797747c2b97505dc0b9f3df0de4f489eac731c23ae9ca9cc31"},
	}
}

func doubleAngleRowsMatch(a, b []FinalizationRowFingerprint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func doubleAngleWriteResult(result DoubleAngleResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for DoubleAngle diagnostic")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	summary := struct {
		SchemaVersion       string                      `json:"schema_version"`
		Timestamp           time.Time                   `json:"timestamp"`
		Primary             RepositoryMetadata          `json:"primary_repository"`
		Lattigo             RepositoryMetadata          `json:"lattigo_repository"`
		AuthoritativeOracle string                      `json:"authoritative_poly_oracle"`
		Threshold           float64                     `json:"threshold"`
		DoubleAngleCount    int                         `json:"double_angle_count"`
		Sqrt2PiInitial      string                      `json:"sqrt2pi_initial"`
		InputScale          string                      `json:"original_mod1_input_scale"`
		P4Semantic          *PSGlobalMetric             `json:"p4_semantic"`
		P4Capacity          TargetScaleCapacityEvidence `json:"p4_capacity"`
		P4Precondition      map[string]bool             `json:"p4_precondition"`
		Rounds              []DoubleAngleCheckpoint     `json:"rounds"`
		FinalBeforeReset    DoubleAngleFinalEvidence    `json:"da_f0_before_metadata_reset"`
		FinalScaleReset     DoubleAngleResetEvidence    `json:"da_f1_metadata_reset"`
		FirstFailingRound   string                      `json:"first_failing_round"`
		FirstFailingPoint   string                      `json:"first_failing_checkpoint"`
		LastPassingPoint    string                      `json:"last_passing_checkpoint"`
		FirstSupportedCause string                      `json:"first_supported_cause"`
		Validation          map[string]interface{}      `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.AuthoritativeOracle, result.Threshold, result.DoubleAngleCount, result.Sqrt2PiInitial, result.InputScale, result.P4Semantic, result.P4Capacity, result.P4Precondition, result.Rounds, result.FinalBeforeReset, result.FinalScaleReset, result.FirstFailingRound, result.FirstFailingPoint, result.LastPassingPoint, result.FirstSupportedCause, result.Validation}
	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DIAG-DOUBLE-ANGLE-logN13-summary.json"), summaryData, 0o644)
}

func runFIX001P3DoubleAngle(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	o2 := chebyshevRawOracle(state.Poly, state.Z)
	post := psGlobalCopy(state.Params.BootstrappingParameters, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(post, post); err != nil {
		return err
	}
	postDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, post)
	if err != nil {
		return err
	}
	postMetric := psGlobalMetric(o2, postDecoded)
	scaleEvidence, multiplier, matchedScale := targetScaleScaleEvidence(post.Scale, state.Target)
	promoted := psGlobalCopy(state.Params.BootstrappingParameters, post)
	if err := state.Eval.FastCKKS.MulIntegerMaintained(promoted, multiplier, promoted); err != nil {
		return err
	}
	promoted.Scale = matchedScale
	promotedDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	promoted.Scale = state.Target
	p4Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	p4Capacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	p4Rows := finalizationEvidence(promoted).Rows
	p4Metric := psGlobalMetric(o2, p4Decoded)
	p4Precondition := map[string]bool{
		"exact_target_scale":           promoted.Scale.Equal(state.Target),
		"expected_level":               promoted.Level() == 7,
		"expected_degree_one":          promoted.Degree() == 1,
		"prior_p4_rows_match":          doubleAngleRowsMatch(p4Rows, doubleAngleExpectedP4Rows()),
		"p4_semantic_pass":             p4Metric.Pass,
		"p4_capacity_pass":             p4Capacity.Pass,
		"input_scale_captured":         state.InputScaleValue.Value.Sign() > 0,
		"canonical_root_hash_match":    state.Root.RowsMatch([]string{"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf", "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"}),
		"target_scale_log2_delta_pass": scaleEvidence.Log2Delta >= 32,
	}
	result := DoubleAngleResult{
		SchemaVersion: "fix-001-p3-diag-double-angle.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(state.Params.BootstrappingParameters, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, DoubleAngleCount: state.Eval.Mod1Parameters.DoubleAngle,
		Sqrt2PiInitial: fmt.Sprintf("%.17g", state.Eval.Mod1Parameters.Sqrt2Pi), InputScale: finalizationScaleString(state.InputScaleValue), P4Level: promoted.Level(), P4Degree: promoted.Degree(), P4Scale: finalizationScaleString(promoted.Scale), P4Semantic: p4Metric, P4Capacity: p4Capacity, P4Rows: p4Rows, P4Precondition: p4Precondition,
		Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "corrected_pre_final_max_component": psGlobalMetric(o2, state.Root.actual).MaxComponent, "corrected_post_final_max_component": postMetric.MaxComponent, "historical_o1_used_for_control": false, "no_logn16": true, "no_lower_scale_sweep": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_later_bootstrap_stage": true},
	}
	_ = promotedDecoded
	if !allBoolMap(p4Precondition) {
		result.FirstSupportedCause = "double_angle_precondition_mismatch"
		result.FirstFailingRound = "none"
		result.FirstFailingPoint = "precondition"
		result.LastPassingPoint = "none"
		return doubleAngleWriteResult(result, outPath)
	}

	res := promoted
	yHP := doubleAngleHPVector(o2)
	lastMetric := p4Metric
	sqrt2pi := state.Eval.Mod1Parameters.Sqrt2Pi
	for round := 0; round < state.Eval.Mod1Parameters.DoubleAngle; round++ {
		y := doubleAngleHPToVector(yHP)
		inputDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		inputCapacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d0 := doubleAngleCheckpoint(round, fmt.Sprintf("D0.%d", round), "round_input", res, y, inputDecoded, nil, lastMetric, inputCapacity)
		result.Rounds = append(result.Rounds, d0)
		result.LastPassingPoint = fmt.Sprintf("D0.%d", round)
		if !d0.Semantic.Pass {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_round_input_already_wrong", fmt.Sprint(round), d0.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}
		if !d0.Capacity.Pass {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_round_input_capacity_failure", fmt.Sprint(round), d0.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}

		mHP := make([]doubleAngleHP, len(yHP))
		for i, value := range yHP {
			mHP[i] = doubleAngleHPSquare(value)
		}
		m := doubleAngleHPToVector(mHP)
		d1Scale := res.Scale.Mul(res.Scale)
		if err := state.Eval.FastCKKS.MulRelin(res, res, res); err != nil {
			return err
		}
		d1Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d1Capacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d1 := doubleAngleCheckpoint(round, fmt.Sprintf("D1.%d", round), "FastCKKS.MulRelin", res, m, d1Decoded, &d1Scale, lastMetric, d1Capacity)
		result.Rounds = append(result.Rounds, d1)
		if !d1.Semantic.Pass {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_multiply_failure", fmt.Sprint(round), d1.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}

		dHP := make([]doubleAngleHP, len(mHP))
		for i, value := range mHP {
			dHP[i] = doubleAngleHPDouble(value)
		}
		d := doubleAngleHPToVector(dHP)
		if err := state.Eval.FastCKKS.Add(res, res, res); err != nil {
			return err
		}
		d2Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d2Capacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d2 := doubleAngleCheckpoint(round, fmt.Sprintf("D2.%d", round), "FastCKKS.Add_double", res, d, d2Decoded, &d1Scale, d1.Semantic, d2Capacity)
		result.Rounds = append(result.Rounds, d2)
		if !d2.Semantic.Pass {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_doubling_failure", fmt.Sprint(round), d2.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}

		sqrt2pi *= sqrt2pi
		oHP := make([]doubleAngleHP, len(dHP))
		for i, value := range dHP {
			oHP[i] = doubleAngleHPOffset(value, sqrt2pi)
		}
		o := doubleAngleHPToVector(oHP)
		if err := state.Eval.FastCKKS.Add(res, -sqrt2pi, res); err != nil {
			return err
		}
		d3Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d3Capacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d3 := doubleAngleCheckpoint(round, fmt.Sprintf("D3.%d", round), "FastCKKS.Add_negative_sqrt2pi", res, o, d3Decoded, &d1Scale, d2.Semantic, d3Capacity)
		d3.Constant = fmt.Sprintf("%.17g", sqrt2pi)
		result.Rounds = append(result.Rounds, d3)
		if !d3.Semantic.Pass {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_offset_failure", fmt.Sprint(round), d3.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}
		if !d3.Capacity.Pass {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_pre_rescale_centered_capacity_failure", fmt.Sprint(round), d3.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}

		beforeLevel := res.Level()
		expectedD4Scale, expectedD4Level, dropped := doubleAngleDroppedScale(res.Scale, beforeLevel, state.Params.BootstrappingParameters)
		if err := state.Eval.FastCKKS.Rescale(res, res); err != nil {
			return err
		}
		d4Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d4Capacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, res)
		if err != nil {
			return err
		}
		d4 := doubleAngleCheckpoint(round, fmt.Sprintf("D4.%d", round), "FastCKKS.Rescale", res, o, d4Decoded, &expectedD4Scale, d3.Semantic, d4Capacity)
		d4.PreLevel, d4.PostLevel, d4.ExpectedPostLevel, d4.DroppedModuli = beforeLevel, res.Level(), expectedD4Level, dropped
		result.Rounds = append(result.Rounds, d4)
		if !d4.Semantic.Pass || !d4.Capacity.Pass || res.Level() != expectedD4Level {
			result.FirstSupportedCause, result.FirstFailingRound, result.FirstFailingPoint = "double_angle_rescale_failure", fmt.Sprint(round), d4.Checkpoint
			return doubleAngleWriteResult(result, outPath)
		}
		lastMetric = d4.Semantic
		yHP = oHP
		result.LastPassingPoint = d4.Checkpoint
	}

	finalDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
	if err != nil {
		return err
	}
	finalCapacity, err := doubleAngleCapacity(state.Params.BootstrappingParameters, res)
	if err != nil {
		return err
	}
	finalY := doubleAngleHPToVector(yHP)
	finalMetric := psGlobalMetric(finalY, finalDecoded)
	scalingFactor := state.Eval.Mod1Parameters.ScalingFactor()
	result.FinalBeforeReset = DoubleAngleFinalEvidence{Level: res.Level(), Degree: res.Degree(), Scale: finalizationScaleString(res.Scale), ScalingFactor: finalizationScaleString(scalingFactor), ScaleLog2DeltaToScaling: doubleAngleScaleDelta(res.Scale, scalingFactor), Semantic: finalMetric, Capacity: finalCapacity, Rows: finalizationEvidence(res).Rows}
	if !finalMetric.Pass || !finalCapacity.Pass || res.Degree() != 1 {
		result.FirstSupportedCause = "double_angle_accumulated_drift_failure"
		return doubleAngleWriteResult(result, outPath)
	}

	beforeResetScale := res.Scale
	beforeResetLevel, beforeResetDegree := res.Level(), res.Degree()
	beforeResetRows := finalizationEvidence(res).Rows
	ratio := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetPrec(doubleAnglePrecision).Set(&beforeResetScale.Value), new(big.Float).SetPrec(doubleAnglePrecision).Set(&state.InputScaleValue.Value))
	resetExpected := doubleAngleHPToVector(doubleAngleHPScale(yHP, ratio))
	res.Scale = state.InputScaleValue
	resetDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, res)
	if err != nil {
		return err
	}
	resetMetric := psGlobalMetric(resetExpected, resetDecoded)
	perturbation := psGlobalMetric(finalDecoded, resetDecoded)
	resetRows := finalizationEvidence(res).Rows
	result.FinalScaleReset = DoubleAngleResetEvidence{ScaleBeforeReset: finalizationScaleString(beforeResetScale), InputScale: finalizationScaleString(state.InputScaleValue), ScaleRatio: ratio.Text('e', 80), ScaleAfterReset: finalizationScaleString(res.Scale), LevelBefore: beforeResetLevel, LevelAfter: res.Level(), DegreeBefore: beforeResetDegree, DegreeAfter: res.Degree(), RowsUnchanged: doubleAngleRowsMatch(beforeResetRows, resetRows), MetadataExpectedSemantic: resetMetric, PerturbationFromBefore: perturbation, Rows: resetRows}
	if !resetMetric.Pass || !result.FinalScaleReset.RowsUnchanged || !res.Scale.Equal(state.InputScaleValue) || res.Level() != beforeResetLevel || res.Degree() != beforeResetDegree {
		result.FirstSupportedCause = "double_angle_final_scale_reset_failure"
		return doubleAngleWriteResult(result, outPath)
	}
	result.FirstSupportedCause = "logn13_double_angle_and_scale_reset_validated"
	return doubleAngleWriteResult(result, outPath)
}

func allBoolMap(values map[string]bool) bool {
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}
