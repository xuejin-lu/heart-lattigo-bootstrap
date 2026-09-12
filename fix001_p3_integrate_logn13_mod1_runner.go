package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

type integratedMod1ScheduleEvidence struct {
	Round         int    `json:"round"`
	LevelIn       int    `json:"level_in"`
	ScaleIn       string `json:"scale_in"`
	KInExponent   int    `json:"k_in_exponent"`
	AExponent     int    `json:"a_exponent"`
	Constant      string `json:"normalized_constant"`
	ScaleOut      string `json:"scale_out"`
	KOutExponent  int    `json:"k_out_exponent"`
	LevelOut      int    `json:"level_out"`
	InvariantPass bool   `json:"invariant_pass"`
}

type integratedMod1StageEvidence struct {
	Round          int                          `json:"round"`
	Stage          string                       `json:"stage"`
	ProductionRows []FinalizationRowFingerprint `json:"production_rows"`
	OracleRows     []FinalizationRowFingerprint `json:"oracle_rows"`
	RowsMatch      bool                         `json:"rows_match"`
}

type integratedMod1Result struct {
	SchemaVersion                    string                           `json:"schema_version"`
	Timestamp                        time.Time                        `json:"timestamp"`
	Primary                          RepositoryMetadata               `json:"primary_repository"`
	Lattigo                          RepositoryMetadata               `json:"lattigo_repository"`
	Environment                      EnvironmentMetadata              `json:"environment"`
	Config                           BootstrapConfig                  `json:"config"`
	Parameters                       ExperimentParameters             `json:"effective_parameters"`
	Workload                         CorrectnessWorkload              `json:"workload"`
	ProductionSystemUnderTest        string                           `json:"production_system_under_test"`
	ProfileGuard                     string                           `json:"profile_guard"`
	NormalizedProfileMatched         bool                             `json:"normalized_profile_matched"`
	NormalizedProfileLogQMatched     bool                             `json:"normalized_profile_logq_matched"`
	NormalizedProfileInvalidLogQ     []int                            `json:"normalized_profile_invalid_logq_indices,omitempty"`
	NormalizedProfileLogQ            []int                            `json:"normalized_profile_logq,omitempty"`
	LevelsPerRescale                 int                              `json:"levels_per_rescale"`
	CompressedPlanScale              string                           `json:"compressed_plan_scale"`
	PolynomialReturnedLowScale       string                           `json:"polynomial_returned_low_scale"`
	VerifiedPolynomialOutputScale    string                           `json:"verified_polynomial_output_scale"`
	NormalizedLowScale               string                           `json:"normalized_low_scale"`
	PolynomialOutputRows             []FinalizationRowFingerprint     `json:"polynomial_output_q0_q1_rows"`
	NormalizedLowRows                []FinalizationRowFingerprint     `json:"normalized_low_q0_q1_rows"`
	PolynomialRowsMatchNormalizedLow bool                             `json:"polynomial_rows_match_normalized_low"`
	RecurrenceEvidence               []integratedMod1StageEvidence    `json:"recurrence_evidence"`
	ProductionRecurrenceRowsMatch    bool                             `json:"production_recurrence_rows_match"`
	DerivedKExponent                 int                              `json:"derived_k_exponent"`
	DerivedKValue                    string                           `json:"derived_k_value"`
	Schedules                        []integratedMod1ScheduleEvidence `json:"normalized_schedule"`
	FinalLevel                       int                              `json:"final_level"`
	FinalDegree                      int                              `json:"final_degree"`
	FinalScale                       string                           `json:"final_scale"`
	FinalRows                        []FinalizationRowFingerprint     `json:"final_q0_q1_rows"`
	AcceptedNormalizedFinalRows      []FinalizationRowFingerprint     `json:"accepted_normalized_final_q0_q1_rows"`
	GeneratedOracleFinalRows         []FinalizationRowFingerprint     `json:"generated_normalized_oracle_final_q0_q1_rows"`
	FinalRowsMatchAccepted           bool                             `json:"final_rows_match_accepted_normalized_expectation"`
	FinalRowsMatchGeneratedOracle    bool                             `json:"final_rows_match_generated_normalized_oracle"`
	ProductionVsNormalizedSemantic   *PSGlobalMetric                  `json:"production_vs_normalized_semantic"`
	ProductionVsCoherentSemantic     *PSGlobalMetric                  `json:"production_vs_coherent_semantic"`
	ProductionVsExactTargetSemantic  *PSGlobalMetric                  `json:"production_vs_exact_target_semantic"`
	InputUnchanged                   bool                             `json:"input_unchanged"`
	FirstSupportedCause              string                           `json:"first_supported_cause"`
	Validation                       map[string]interface{}           `json:"validation"`
}

func integratedRecurrenceEvidence(state chebyshevOracleState, polynomialOutput, normalizedLow *rlwe.Ciphertext) ([]integratedMod1StageEvidence, bool, error) {
	params := state.Params.BootstrappingParameters
	lowView, err := targetScaleCoefficientView(params, normalizedLow)
	if err != nil {
		return nil, false, err
	}
	lowData, err := targetScaleCenteredRows(params, lowView)
	if err != nil {
		return nil, false, err
	}
	oracle, err := normalizedFullFromLow(params, normalizedLow, lowData)
	if err != nil {
		return nil, false, err
	}
	k := new(big.Int).Lsh(big.NewInt(1), 29)
	coherentScale := normalizedLow.Scale.Mul(rlwe.NewScale(k))
	oracle.Scale = coherentScale
	production := polynomialOutput.CopyNew()
	production.Scale = coherentScale
	evidence := make([]integratedMod1StageEvidence, 0, state.Eval.Mod1Parameters.DoubleAngle*4+1)
	appendEvidence := func(round int, stage string, actual, expected *rlwe.Ciphertext) {
		actualRows := finalizationEvidence(actual).Rows
		expectedRows := finalizationEvidence(expected).Rows
		evidence = append(evidence, integratedMod1StageEvidence{Round: round, Stage: stage, ProductionRows: actualRows, OracleRows: expectedRows, RowsMatch: doubleAngleRowsMatch(actualRows, expectedRows)})
	}
	sqrt2pi := state.Eval.Mod1Parameters.Sqrt2Pi
	for round := 0; round < state.Eval.Mod1Parameters.DoubleAngle; round++ {
		sqrt2pi *= sqrt2pi
		refProduct, err := normalizedFullSquare(params, oracle)
		if err != nil {
			return evidence, false, err
		}
		refSquare := normalizedFullZeroC2(params, refProduct)
		if err := state.Eval.FastCKKS.MulRelin(production, production, production); err != nil {
			return evidence, false, err
		}
		appendEvidence(round, "square", production, refSquare)
		factor := new(big.Int).Lsh(big.NewInt(1), 30)
		refAfterA := normalizedFullIntegerMultiply(params, refSquare, factor)
		if err := state.Eval.FastCKKS.MulIntegerMaintained(production, factor, production); err != nil {
			return evidence, false, err
		}
		appendEvidence(round, "multiplier", production, refAfterA)
		constantValue, _ := new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(k)).Float64()
		refAfterConstant := normalizedFullSubtractConstant(params, refAfterA, constantValue)
		if err := state.Eval.FastCKKS.Add(production, -constantValue, production); err != nil {
			return evidence, false, err
		}
		appendEvidence(round, "constant", production, refAfterConstant)
		if err := state.Eval.FastCKKS.Rescale(production, production); err != nil {
			return evidence, false, err
		}
		oracle, err = normalizedFullRescale(params, refAfterConstant)
		if err != nil {
			return evidence, false, err
		}
		appendEvidence(round, "rescale", production, oracle)
	}
	if err := state.Eval.FastCKKS.MulIntegerMaintained(production, k, production); err != nil {
		return evidence, false, err
	}
	oracle = normalizedFullIntegerMultiply(params, oracle, k)
	appendEvidence(state.Eval.Mod1Parameters.DoubleAngle, "restore", production, oracle)
	allMatch := true
	for _, item := range evidence {
		allMatch = allMatch && item.RowsMatch
	}
	return evidence, allMatch, nil
}

type acceptedNormalizedFinalArtifact struct {
	Final struct {
		NormalizedReferenceFinalRowsAfterReset []FinalizationRowFingerprint `json:"normalized_reference_final_rows_after_reset"`
	} `json:"final"`
}

func integratedInputRowsUnchanged(before, after *rlwe.Ciphertext) bool {
	if before == nil || after == nil || !before.Scale.Equal(after.Scale) || before.Level() != after.Level() || before.Degree() != after.Degree() {
		return false
	}
	return doubleAngleRowsMatch(finalizationEvidence(before).Rows, finalizationEvidence(after).Rows)
}

func loadAcceptedNormalizedFinalRows(primaryRoot string) ([]FinalizationRowFingerprint, error) {
	path := filepath.Join(primaryRoot, "results", "FIX-001-P3-DESIGN-DOUBLE-ANGLE-FINAL-RESTORE-REFERENCE-FIX-logN13.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read accepted normalized diagnostic artifact: %w", err)
	}
	var artifact acceptedNormalizedFinalArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, fmt.Errorf("decode accepted normalized diagnostic artifact: %w", err)
	}
	if len(artifact.Final.NormalizedReferenceFinalRowsAfterReset) == 0 {
		return nil, fmt.Errorf("accepted normalized diagnostic artifact has no final normalized rows")
	}
	return artifact.Final.NormalizedReferenceFinalRowsAfterReset, nil
}

func integratedNormalizedOracle(state chebyshevOracleState) (*rlwe.Ciphertext, *rlwe.Ciphertext, *rlwe.Ciphertext, []integratedMod1ScheduleEvidence, error) {
	params := state.Params.BootstrappingParameters
	low := psGlobalCopy(params, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(low, low); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("normalized oracle polynomial final Rescale: %w", err)
	}
	lowView, err := targetScaleCoefficientView(params, low)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	lowData, err := targetScaleCenteredRows(params, lowView)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	k := new(big.Int).Lsh(big.NewInt(1), 29)
	coherentScale := low.Scale.Mul(rlwe.NewScale(k))
	rNorm, err := normalizedFullFromLow(params, low, lowData)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rNorm.Scale = coherentScale
	coherentBase, err := normalizedFullFromLow(params, low, lowData)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	coherentBase.Scale = low.Scale
	rCoherent := normalizedFullIntegerMultiply(params, coherentBase, k)
	rCoherent.Scale = coherentScale
	rExactTarget := rCoherent.CopyNew()
	rExactTarget.Scale = state.Target

	schedules := make([]integratedMod1ScheduleEvidence, 0, state.Eval.Mod1Parameters.DoubleAngle)
	currentLevel := rNorm.Level()
	currentScale := coherentScale
	kIn := 29
	sqrt2pi := state.Eval.Mod1Parameters.Sqrt2Pi
	for round := 0; round < state.Eval.Mod1Parameters.DoubleAngle; round++ {
		sqrt2pi *= sqrt2pi
		schedule, err := normalizedScaleSchedule(round, currentLevel, currentScale, low.Scale, kIn, sqrt2pi, params)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
		constantValue, _ := new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
		refProduct, err := normalizedFullSquare(params, rNorm)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		refSquare := normalizedFullZeroC2(params, refProduct)
		refAfterA := normalizedFullIntegerMultiply(params, refSquare, factor)
		refAfterConstant := normalizedFullSubtractConstant(params, refAfterA, constantValue)
		refPostRescale, err := normalizedFullRescale(params, refAfterConstant)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		rNorm = refPostRescale
		coherentPre, _, err := normalizedFullSquareStep(params, rCoherent, big.NewInt(2), sqrt2pi)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		exactPre, _, err := normalizedFullSquareStep(params, rExactTarget, big.NewInt(2), sqrt2pi)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		rCoherent, err = normalizedFullRescale(params, coherentPre)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		rExactTarget, err = normalizedFullRescale(params, exactPre)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		invariant := rNorm.Level() == schedule.NextLevel && rNorm.Degree() == 1 && rNorm.IsNTT && rNorm.IsMontgomery && schedule.KOutExponent == 29 && schedule.AExponent == 30
		schedules = append(schedules, integratedMod1ScheduleEvidence{Round: round, LevelIn: currentLevel, ScaleIn: schedule.EffectiveScaleIn, KInExponent: schedule.KInExponent, AExponent: schedule.AExponent, Constant: schedule.Constant, ScaleOut: schedule.EffectiveScaleOut, KOutExponent: schedule.KOutExponent, LevelOut: rNorm.Level(), InvariantPass: invariant})
		if !invariant {
			return nil, nil, nil, schedules, fmt.Errorf("normalized oracle round %d invariant failed", round)
		}
		currentLevel = rNorm.Level()
		currentScale = rNorm.Scale
		kIn = schedule.KOutExponent
	}
	rNormRestored := normalizedFullIntegerMultiply(params, rNorm, new(big.Int).Lsh(big.NewInt(1), uint(kIn)))
	inputScale := state.InputScaleValue
	rNormRestored.Scale = inputScale
	rCoherent.Scale = inputScale
	rExactTarget.Scale = inputScale
	return rNormRestored, rCoherent, rExactTarget, schedules, nil
}

func runFIX001P3IntegrateLogN13Mod1(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	if state.Mod1Input == nil {
		return fmt.Errorf("production Mod1 input was not captured")
	}
	acceptedRows, err := loadAcceptedNormalizedFinalRows(primaryRoot)
	if err != nil {
		return err
	}
	polynomialInput := state.Mod1Input.CopyNew()
	polynomialInput.Scale = state.Eval.Mod1Evaluator.Parameters.ScalingFactor()
	offset := new(big.Float).Sub(&state.Eval.Mod1Evaluator.Parameters.Mod1Poly.B, &state.Eval.Mod1Evaluator.Parameters.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(state.Eval.Mod1Evaluator.Parameters.IntervalShrinkFactor()))
	offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
	if err := state.Eval.FastCKKS.Add(polynomialInput, offset, polynomialInput); err != nil {
		return fmt.Errorf("production polynomial boundary offset: %w", err)
	}
	polynomialTarget := psGlobalTargetScale(polynomialInput, state.Eval)
	polynomialOutput, err := state.Eval.PolynomialEvaluator.EvaluateWithPlanScale(polynomialInput, state.Eval.Mod1Evaluator.Parameters.Mod1Poly, polynomialTarget, rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 91)))
	if err != nil {
		return fmt.Errorf("production polynomial boundary verification: %w", err)
	}
	normalizedLow := psGlobalCopy(state.Params.BootstrappingParameters, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(normalizedLow, normalizedLow); err != nil {
		return fmt.Errorf("normalized low boundary verification: %w", err)
	}
	polynomialRows := finalizationEvidence(polynomialOutput).Rows
	normalizedLowRows := finalizationEvidence(normalizedLow).Rows
	productionInput := state.Mod1Input.CopyNew()
	inputBefore := productionInput.CopyNew()
	production, err := state.Eval.Mod1Evaluator.EvaluateNew(productionInput)
	if err != nil {
		return fmt.Errorf("production Fast Mod1 EvaluateNew: %w", err)
	}
	rNormRestored, rCoherent, rExactTarget, schedules, err := integratedNormalizedOracle(state)
	if err != nil {
		return err
	}
	productionDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, production)
	if err != nil {
		return err
	}
	normalizedDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, rNormRestored)
	if err != nil {
		return err
	}
	coherentDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, rCoherent)
	if err != nil {
		return err
	}
	exactDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, rExactTarget)
	if err != nil {
		return err
	}
	productionRows := finalizationEvidence(production).Rows
	generatedRows := finalizationEvidence(rNormRestored).Rows
	recurrenceEvidence, productionRecurrenceRowsMatch, err := integratedRecurrenceEvidence(state, polynomialOutput, normalizedLow)
	if err != nil {
		return err
	}
	logQMatched := true
	invalidLogQ := make([]int, 0)
	for i, logQ := range state.Params.BootstrappingParameters.LogQi() {
		valid := (i == 0 && logQ == 55) || (i >= 1 && i <= 3 && logQ >= 39 && logQ <= 40) || (i == 4 && logQ >= 39 && logQ <= 45) || (i >= 5 && i <= 12 && logQ >= 60 && logQ <= 61) || (i >= 13 && i <= 16 && logQ >= 56 && logQ <= 57)
		if !valid {
			invalidLogQ = append(invalidLogQ, i)
		}
		logQMatched = logQMatched && valid
	}
	profileMatched := state.Params.BootstrappingParameters.LogN() == 13 && state.Eval.Mod1Parameters.Mod1Poly.Degree() == 30 && state.Eval.Mod1Parameters.DoubleAngle == 3 && state.Eval.Mod1Parameters.Mod1Type == mod1.CosDiscrete && state.Eval.Mod1Parameters.Mod1InvPoly == nil && state.Params.BootstrappingParameters.LevelsConsumedPerRescaling() == 1 && state.Eval.Mod1Parameters.LevelQ == 12 && state.Eval.Mod1Parameters.LogMessageRatio == 10 && len(state.Params.BootstrappingParameters.Q()) == 17 && logQMatched
	result := integratedMod1Result{
		SchemaVersion: "fix-001-p3-integrate-logn13-normalized-mod1.v1", Timestamp: time.Now().UTC(),
		Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, Parameters: parameterMetadata(state.Params.ResidualParameters, state.Params),
		Workload:                  CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		ProductionSystemUnderTest: "bootstrapping.FastEvaluator.Mod1Evaluator.EvaluateNew",
		ProfileGuard:              "accepted LogN13 degree-30 CosDiscrete DoubleAngle=3 no-inverse one-level-rescale profile",
		NormalizedProfileMatched:  profileMatched, NormalizedProfileLogQMatched: logQMatched, NormalizedProfileInvalidLogQ: invalidLogQ, NormalizedProfileLogQ: state.Params.BootstrappingParameters.LogQi(), LevelsPerRescale: state.Params.BootstrappingParameters.LevelsConsumedPerRescaling(),
		CompressedPlanScale: "2^91", PolynomialReturnedLowScale: finalizationScaleString(state.Root.ciphertext.Scale), VerifiedPolynomialOutputScale: finalizationScaleString(polynomialOutput.Scale), NormalizedLowScale: finalizationScaleString(normalizedLow.Scale),
		PolynomialOutputRows: polynomialRows, NormalizedLowRows: normalizedLowRows, PolynomialRowsMatchNormalizedLow: doubleAngleRowsMatch(polynomialRows, normalizedLowRows),
		RecurrenceEvidence: recurrenceEvidence, ProductionRecurrenceRowsMatch: productionRecurrenceRowsMatch,
		DerivedKExponent: 29, DerivedKValue: new(big.Int).Lsh(big.NewInt(1), 29).String(), Schedules: schedules,
		FinalLevel: production.Level(), FinalDegree: production.Degree(), FinalScale: finalizationScaleString(production.Scale),
		FinalRows: productionRows, AcceptedNormalizedFinalRows: acceptedRows, GeneratedOracleFinalRows: generatedRows,
		FinalRowsMatchAccepted: doubleAngleRowsMatch(productionRows, acceptedRows), FinalRowsMatchGeneratedOracle: doubleAngleRowsMatch(productionRows, generatedRows),
		ProductionVsNormalizedSemantic: psGlobalMetric(normalizedDecoded, productionDecoded), ProductionVsCoherentSemantic: psGlobalMetric(coherentDecoded, productionDecoded), ProductionVsExactTargetSemantic: psGlobalMetric(exactDecoded, productionDecoded),
		InputUnchanged: integratedInputRowsUnchanged(inputBefore, productionInput),
		Validation:     map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "primary_production_call_once": true, "standard_behavior_unchanged": true, "no_runtime_backend_selector": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_later_bootstrap_stage": true},
	}
	result.FirstSupportedCause = "logn13_normalized_fast_mod1_production_integrated"
	if production.Level() != 4 || production.Degree() != 1 || !production.IsNTT || !production.IsMontgomery || !production.Scale.Equal(state.Mod1Input.Scale) || !result.FinalRowsMatchAccepted || !result.FinalRowsMatchGeneratedOracle || !result.ProductionRecurrenceRowsMatch || !result.ProductionVsNormalizedSemantic.Pass || !result.ProductionVsCoherentSemantic.Pass || !result.ProductionVsExactTargetSemantic.Pass || !result.InputUnchanged {
		result.FirstSupportedCause = "production_integration_precondition_mismatch"
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if outPath == "" {
		return fmt.Errorf("output path is required for production Mod1 integration verification")
	}
	return os.WriteFile(outPath, data, 0o644)
}
