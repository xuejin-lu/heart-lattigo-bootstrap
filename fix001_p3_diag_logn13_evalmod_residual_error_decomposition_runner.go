//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3EvalModResidualPrimary   = "02d472f7e5c02709bafd5ef67586a8337684c4f3"
	requiredFIX001P3EvalModResidualSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	fix001P3EvalModResidualStageThreshold    = 1e-6
	fix001P3EvalModResidualClosureTolerance  = 1e-10
	fix001P3EvalModResidualPublicThreshold   = 1e-2
	fix001P3EvalModResidualPostS2CThreshold  = 3.125e-4
)

type fix001P3EvalModResidualStage struct {
	Fast       *rlwe.Ciphertext
	Normalized *rlwe.Ciphertext
}

type fix001P3EvalModResidualRun struct {
	State        fix001P3DAR0LocalQ2State
	Stages       map[string]fix001P3EvalModResidualStage
	InitialLevel int
	InitialScale rlwe.Scale
	WorkingScale rlwe.Scale
	InitialK     int
	InitialSqrt  float64
}

type fix001P3EvalModResidualResult struct {
	SchemaVersion         string                            `json:"schema_version"`
	Timestamp             time.Time                         `json:"timestamp"`
	Primary               RepositoryMetadata                `json:"primary_repository"`
	Lattigo               RepositoryMetadata                `json:"lattigo_repository"`
	Environment           EnvironmentMetadata               `json:"environment"`
	Config                BootstrapConfig                   `json:"config"`
	Parameters            ExperimentParameters              `json:"effective_parameters"`
	E0                    map[string]interface{}            `json:"e0_control"`
	Polynomial            map[string]interface{}            `json:"polynomial_input"`
	Causal                map[string]interface{}            `json:"four_state_evalmod_causal_decomposition"`
	Closure               *PSGlobalMetric                   `json:"closure_residual"`
	Checkpoints           []map[string]interface{}          `json:"da_checkpoint_attribution"`
	Counterfactuals       map[string]map[string]interface{} `json:"downstream_counterfactuals"`
	Projected             map[string]interface{}            `json:"s2c_projected_components"`
	RequiredPostS2C       float64                           `json:"required_post_s2c_threshold"`
	PublicThreshold       float64                           `json:"public_like_threshold"`
	Classification        string                            `json:"classification"`
	FirstRemainingBlocker string                            `json:"first_remaining_blocker"`
	Validation            map[string]interface{}            `json:"validation"`
}

func fix001P3EvalModResidualWrite(result fix001P3EvalModResidualResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3EvalModResidualMetric(reference, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(reference, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func fix001P3EvalModResidualSub(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

func fix001P3EvalModResidualAdd(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func fix001P3EvalModResidualCombine(realValues, imagValues []complex128) []complex128 {
	out := make([]complex128, len(realValues))
	for i := range out {
		out[i] = realValues[i] + 1i*imagValues[i]
	}
	return out
}

func fix001P3EvalModResidualScale(values []complex128, factor *big.Float) []complex128 {
	return doubleAngleHPToVector(doubleAngleHPScale(doubleAngleHPVector(values), factor))
}

func fix001P3EvalModResidualIdealNormalized(values []complex128, run fix001P3EvalModResidualRun, params ckks.Parameters, rounds int) (map[string][]complex128, []complex128, error) {
	stages := make(map[string][]complex128)
	normalized := fix001P3EvalModResidualScale(values, new(big.Float).Quo(big.NewFloat(1), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(run.InitialK)))))
	currentLevel, currentScale, kIn := run.InitialLevel, run.InitialScale, run.InitialK
	sqrt2pi := run.InitialSqrt
	for round := 0; round < rounds; round++ {
		sqrt2pi *= sqrt2pi
		schedule, err := normalizedScaleSchedule(round, currentLevel, currentScale, run.WorkingScale, kIn, sqrt2pi, params)
		if err != nil {
			return nil, nil, err
		}
		stages[fmt.Sprintf("round%d.input", round)] = append([]complex128(nil), normalized...)
		square := make([]doubleAngleHP, len(normalized))
		for i, value := range doubleAngleHPVector(normalized) {
			square[i] = doubleAngleHPSquare(value)
		}
		squareValues := doubleAngleHPToVector(square)
		stages[fmt.Sprintf("round%d.square", round)] = squareValues
		aFactor := new(big.Float).SetPrec(doubleAnglePrecision).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent)))
		afterMultiplier := fix001P3EvalModResidualScale(squareValues, aFactor)
		stages[fmt.Sprintf("round%d.after_multiplier", round)] = afterMultiplier
		constant := new(big.Float).Quo(new(big.Float).SetPrec(doubleAnglePrecision).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent))))
		afterConstantHP := make([]doubleAngleHP, len(afterMultiplier))
		for i, value := range doubleAngleHPVector(afterMultiplier) {
			afterConstantHP[i] = doubleAngleHPOffset(value, float64FromBigFloat(constant))
		}
		normalized = doubleAngleHPToVector(afterConstantHP)
		stages[fmt.Sprintf("round%d.after_constant", round)] = append([]complex128(nil), normalized...)
		stages[fmt.Sprintf("round%d.post_rescale", round)] = append([]complex128(nil), normalized...)
		currentLevel = schedule.NextLevel
		currentScale = scheduleScaleValue(schedule, params, currentLevel, currentScale)
		kIn = schedule.KOutExponent
	}
	final := fix001P3EvalModResidualScale(normalized, new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(kIn))))
	stages["final"] = append([]complex128(nil), final...)
	return stages, final, nil
}

func fix001P3EvalModResidualRunState(profile q056PreparedProfile, fastInput, standardInput, fastPoly *rlwe.Ciphertext, candidate psRescaleGuardCandidate) (fix001P3EvalModResidualRun, error) {
	state, err := fix001P3DAR0PrepareState(fastInput, standardInput, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, precisionSweepScale(92), fastPoly)
	if err != nil {
		return fix001P3EvalModResidualRun{}, err
	}
	run := fix001P3EvalModResidualRun{State: state, Stages: make(map[string]fix001P3EvalModResidualStage), InitialLevel: state.CurrentLevel, InitialScale: state.CurrentScale, WorkingScale: state.WorkingScale, InitialK: state.KIn, InitialSqrt: state.Sqrt2Pi}
	trace := func(round int, checkpoint string, fast, normalized *rlwe.Ciphertext) {
		run.Stages[fmt.Sprintf("round%d.%s", round, checkpoint)] = fix001P3EvalModResidualStage{Fast: fast.CopyNew(), Normalized: normalized.CopyNew()}
	}
	for round := 0; round < profile.Fast.Mod1Parameters.DoubleAngle; round++ {
		outcome, err := fix001P3DAAllRoundsRoundWithTrace(&run.State, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, trace)
		if err != nil {
			return run, err
		}
		if outcome.Classification != "" {
			return run, fmt.Errorf("residual DA replay failed at round %d: %s", round, outcome.Blocker)
		}
	}
	if err := fix001P3DAAllRoundsFinalize(&run.State, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return run, err
	}
	if run.State.Path.FastFinal == nil || run.State.Path.NormalizedFinal == nil {
		return run, fmt.Errorf("residual DA replay returned nil final ciphertext")
	}
	run.Stages["final"] = fix001P3EvalModResidualStage{Fast: run.State.Path.FastFinal.CopyNew(), Normalized: run.State.Path.NormalizedFinal.CopyNew()}
	_ = candidate
	return run, nil
}

func fix001P3EvalModResidualOracleCipher(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128) (*rlwe.Ciphertext, error) {
	return evalModFastFromStandard(params, source, values)
}

func fix001P3EvalModResidualFullOracleCipher(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128) (*rlwe.Ciphertext, error) {
	plaintext, err := evalModEncodePlaintext(params, values, source)
	if err != nil {
		return nil, err
	}
	ciphertext := ckks.NewCiphertext(params, 1, source.Level())
	*ciphertext.MetaData = *plaintext.MetaData
	ciphertext.IsNTT, ciphertext.IsMontgomery = plaintext.IsNTT, plaintext.IsMontgomery
	for limb := range plaintext.Value.Coeffs {
		copy(ciphertext.Value[0].Coeffs[limb], plaintext.Value.Coeffs[limb])
	}
	return ciphertext, nil
}

func fix001P3EvalModResidualHighPrecisionOracleCipher(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128) (*rlwe.Ciphertext, error) {
	plaintext := ckks.NewPlaintext(params, source.Level())
	*plaintext.MetaData = *source.MetaData
	plaintext.IsNTT = true
	plaintext.IsMontgomery = false
	precise := make([]*bignum.Complex, len(values))
	for i, value := range values {
		precise[i] = bignum.ToComplex(value, 256)
	}
	if err := ckks.NewEncoder(params, 256).Encode(precise, plaintext); err != nil {
		return nil, err
	}
	ciphertext := ckks.NewCiphertext(params, 1, source.Level())
	*ciphertext.MetaData = *plaintext.MetaData
	ciphertext.IsNTT, ciphertext.IsMontgomery = plaintext.IsNTT, plaintext.IsMontgomery
	for limb := range plaintext.Value.Coeffs {
		copy(ciphertext.Value[0].Coeffs[limb], plaintext.Value.Coeffs[limb])
	}
	return ciphertext, nil
}

func fix001P3EvalModResidualFullNormalizedRun(params ckks.Parameters, source *rlwe.Ciphertext, workingScale rlwe.Scale, initialK int, initialSqrt float64, rounds int) (fix001P3EvalModResidualRun, error) {
	run := fix001P3EvalModResidualRun{Stages: make(map[string]fix001P3EvalModResidualStage), InitialLevel: source.Level(), InitialScale: source.Scale.Mul(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(initialK)))), WorkingScale: workingScale, InitialK: initialK, InitialSqrt: initialSqrt}
	current := source.CopyNew()
	current.Scale = run.InitialScale
	currentLevel, currentScale, kIn := current.Level(), current.Scale, initialK
	sqrt2pi := initialSqrt
	for round := 0; round < rounds; round++ {
		sqrt2pi *= sqrt2pi
		schedule, err := normalizedScaleSchedule(round, currentLevel, currentScale, workingScale, kIn, sqrt2pi, params)
		if err != nil {
			return run, err
		}
		run.Stages[fmt.Sprintf("round%d.input", round)] = fix001P3EvalModResidualStage{Normalized: current.CopyNew(), Fast: current.CopyNew()}
		preRescale, _, err := normalizedFullSquareStep(params, current, new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent)), float64FromBigFloat(new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent))))))
		if err != nil {
			return run, err
		}
		square, err := normalizedFullSquare(params, current)
		if err != nil {
			return run, err
		}
		run.Stages[fmt.Sprintf("round%d.square", round)] = fix001P3EvalModResidualStage{Normalized: square, Fast: square.CopyNew()}
		refAfterMultiplier := normalizedFullIntegerMultiply(params, square, new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent)))
		run.Stages[fmt.Sprintf("round%d.after_multiplier", round)] = fix001P3EvalModResidualStage{Normalized: refAfterMultiplier, Fast: refAfterMultiplier.CopyNew()}
		kOutValue := new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent))
		constant := float64FromBigFloat(new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(kOutValue)))
		refAfterConstant := normalizedFullSubtractConstant(params, refAfterMultiplier, constant)
		run.Stages[fmt.Sprintf("round%d.after_constant", round)] = fix001P3EvalModResidualStage{Normalized: refAfterConstant, Fast: refAfterConstant.CopyNew()}
		current, err = normalizedFullRescale(params, preRescale)
		if err != nil {
			return run, err
		}
		run.Stages[fmt.Sprintf("round%d.post_rescale", round)] = fix001P3EvalModResidualStage{Normalized: current.CopyNew(), Fast: current.CopyNew()}
		currentLevel, currentScale, kIn = schedule.NextLevel, scheduleScaleValue(schedule, params, schedule.NextLevel, currentScale), schedule.KOutExponent
	}
	finalK := new(big.Int).Lsh(big.NewInt(1), uint(kIn))
	current = normalizedFullIntegerMultiply(params, current, finalK)
	current.Scale = source.Scale
	run.State.Path.NormalizedFinal = current
	run.Stages["final"] = fix001P3EvalModResidualStage{Normalized: current.CopyNew(), Fast: current.CopyNew()}
	return run, nil
}

func fix001P3EvalModResidualDecode(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	_, values, err := semanticBisectView(ct, params, zeroSecret(params))
	return values, err
}

func fix001P3EvalModResidualFullDecode(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	plaintext := rlwe.NewDecryptor(params, zeroSecret(params)).DecryptNew(ct)
	precise := make([]*bignum.Complex, params.MaxSlots())
	for i := range precise {
		precise[i] = bignum.NewComplex().SetPrec(256)
	}
	if err := ckks.NewEncoder(params, 256).Decode(plaintext, precise); err != nil {
		return nil, err
	}
	values := make([]complex128, len(precise))
	for i, value := range precise {
		values[i] = value.Complex128()
	}
	return values, nil
}

func fix001P3EvalModResidualMetricForCiphertexts(params ckks.Parameters, reference, actual *rlwe.Ciphertext, threshold float64) (*PSGlobalMetric, error) {
	referenceValues, err := fix001P3EvalModResidualDecode(params, reference)
	if err != nil {
		return nil, err
	}
	actualValues, err := fix001P3EvalModResidualDecode(params, actual)
	if err != nil {
		return nil, err
	}
	return fix001P3EvalModResidualMetric(referenceValues, actualValues, threshold), nil
}

func fix001P3EvalModResidualRowsAndCapacity(params ckks.Parameters, stage fix001P3EvalModResidualStage) map[string]interface{} {
	row := map[string]interface{}{"rows_match": postMod1S2CRowsEqual(params, stage.Fast, stage.Normalized)}
	if stage.Fast.Level() >= 1 {
		if capacity, err := postMod1S2CCapacityFromFastCiphertext(params, stage.Fast); err == nil {
			row["fast_q01_capacity_pass"] = capacity.Pass
		}
	}
	if capacity, err := normalizedExactCapacity(params, stage.Normalized); err == nil {
		row["normalized_full_rns_capacity_pass"] = capacity.Pass
	}
	return row
}

func fix001P3EvalModResidualCheckpointTable(params ckks.Parameters, fastRun, oracleRun fix001P3EvalModResidualRun, ideal map[string][]complex128) ([]map[string]interface{}, error) {
	rows := make([]map[string]interface{}, 0)
	for round := 0; round < 3; round++ {
		for _, checkpoint := range []string{"input", "square", "after_multiplier", "after_constant", "post_rescale"} {
			key := fmt.Sprintf("round%d.%s", round, checkpoint)
			fastStage, fastOK := fastRun.Stages[key]
			oracleStage, oracleOK := oracleRun.Stages[key]
			if !fastOK || !oracleOK {
				return nil, fmt.Errorf("missing residual stage %s", key)
			}
			fastValues, err := fix001P3EvalModResidualDecode(params, fastStage.Normalized)
			if err != nil {
				return nil, err
			}
			oracleValues, err := fix001P3EvalModResidualFullDecode(params, oracleStage.Normalized)
			if err != nil {
				return nil, err
			}
			idealValues := ideal[key]
			rows = append(rows, map[string]interface{}{
				"round": round, "checkpoint": checkpoint,
				"normalized_da_ps_input_component":   fix001P3EvalModResidualMetric(oracleValues, fastValues, fix001P3EvalModResidualStageThreshold),
				"normalized_da_arithmetic_component": fix001P3EvalModResidualMetric(idealValues, oracleValues, fix001P3EvalModResidualStageThreshold),
				"fast_full_rns_contract":             fix001P3EvalModResidualRowsAndCapacity(params, fastStage),
			})
		}
	}
	return rows, nil
}

func fix001P3EvalModResidualDownstream(profile q056PreparedProfile, baselineCore []complex128, baselineEvalMod []complex128, input *rlwe.Ciphertext, realTemplate, imagTemplate *rlwe.Ciphertext, realValues, imagValues []complex128) (map[string]interface{}, error) {
	params := profile.BTP.BootstrappingParameters
	fastReal, err := fix001P3EvalModResidualOracleCipher(params, realTemplate, realValues)
	if err != nil {
		return nil, err
	}
	fastImag, err := fix001P3EvalModResidualOracleCipher(params, imagTemplate, imagValues)
	if err != nil {
		return nil, err
	}
	fastCore, err := profile.Fast.SlotsToCoeffs(fastReal.CopyNew(), fastImag.CopyNew())
	if err != nil {
		return nil, err
	}
	coreValues, err := fix001P3EvalModResidualDecode(params, fastCore)
	if err != nil {
		return nil, err
	}
	evalModValues := fix001P3EvalModResidualCombine(realValues, imagValues)
	evalModMetric := fix001P3EvalModResidualMetric(baselineEvalMod, evalModValues, fix001P3EvalModResidualPublicThreshold)
	postS2CMetric := fix001P3EvalModResidualMetric(baselineCore, coreValues, fix001P3EvalModResidualPostS2CThreshold)
	beforeInput := input.CopyNew()
	_, ctxtN1, ctxtN2, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, err
	}
	unpacked, err := profile.Fast.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*fastCore.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil || len(unpacked) != 1 {
		if err == nil {
			err = fmt.Errorf("expected one unpacked residual ciphertext, got %d", len(unpacked))
		}
		return nil, err
	}
	_, final, err := fix001P3E2EFinalizeOracle(&unpacked[0], profile.BTP)
	if err != nil {
		return nil, err
	}
	finalValues, err := decodeWithSecret(profile.Residual, final, zeroSecret(profile.Residual))
	if err != nil {
		return nil, err
	}
	publicLike := fix001P3EvalModResidualMetric(reproducibleValues(profile.Residual.MaxSlots()), finalValues, fix001P3EvalModResidualPublicThreshold)
	metadataPass := final.Level() == profile.Residual.MaxLevel() && final.Degree() == 1 && final.N() == profile.Residual.N() && final.IsNTT && !final.IsMontgomery && final.Scale.Equal(profile.Residual.DefaultScale())
	return map[string]interface{}{
		"evalmod_vs_standard": evalModMetric, "post_s2c_vs_standard": postS2CMetric, "public_like": publicLike,
		"final_metadata_pass": metadataPass, "input_unchanged": fix001P3E2EInputUnchanged(beforeInput, input),
		"core_values": coreValues, "final_values": finalValues,
	}, nil
}

func fix001P3EvalModResidualProjectS2C(profile q056PreparedProfile, realTemplate, imagTemplate *rlwe.Ciphertext, vector []complex128) ([]complex128, error) {
	params := profile.BTP.BootstrappingParameters
	realValues := make([]complex128, len(vector))
	imagValues := make([]complex128, len(vector))
	for i, value := range vector {
		realValues[i] = complex(real(value), 0)
		imagValues[i] = complex(imag(value), 0)
	}
	fastReal, err := fix001P3EvalModResidualOracleCipher(params, realTemplate, realValues)
	if err != nil {
		return nil, err
	}
	fastImag, err := fix001P3EvalModResidualOracleCipher(params, imagTemplate, imagValues)
	if err != nil {
		return nil, err
	}
	fullReal, _, err := postMod1S2CLiftFull(params, fastReal)
	if err != nil {
		return nil, err
	}
	fullImag, _, err := postMod1S2CLiftFull(params, fastImag)
	if err != nil {
		return nil, err
	}
	stages, _, err := fix001P3S2CTraceFull(params, profile.Fast.S2CDFTMatrix, profile.Standard.DFTEvaluator, fullReal, fullImag)
	if err != nil {
		return nil, err
	}
	return fix001P3EvalModResidualDecode(params, stages[len(stages)-1])
}

func fix001P3EvalModResidualClassify(hPS, hDA, hBoth bool, implementationMaterial bool, contracts bool) (string, string) {
	if implementationMaterial && contracts {
		return "logn13_evalmod_residual_fast_implementation_mismatch", "E_F-N_F is materially nonzero despite exact Fast/full-RNS checkpoint contracts"
	}
	if hPS && !hDA {
		return "logn13_evalmod_residual_ps_input_blocker", "PS-input propagation is the single-fix blocker"
	}
	if hDA && !hPS {
		return "logn13_evalmod_residual_da_arithmetic_blocker", "normalized DA arithmetic is the single-fix blocker"
	}
	if hPS && hDA {
		return "logn13_evalmod_residual_multiple_single_fix_options", "compare H_PS and H_DA public-like margins"
	}
	if hBoth {
		return "logn13_evalmod_residual_joint_ps_da_blocker", "both PS-input and DA-arithmetic components must be reduced"
	}
	return "logn13_evalmod_residual_mathematical_chain_floor", "H_BOTH remains above the public-like threshold"
}

func runFIX001P3DiagLogN13EvalModResidualDecomposition(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3EvalModResidualPrimary, "HEAD") == requiredFIX001P3EvalModResidualPrimary
	result := fix001P3EvalModResidualResult{
		SchemaVersion: "fix-001-p3-diag-logn13-evalmod-residual-error-decomposition.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		RequiredPostS2C: fix001P3EvalModResidualPostS2CThreshold, PublicThreshold: fix001P3EvalModResidualPublicThreshold,
		Validation: map[string]interface{}{
			"primary_required_base": requiredFIX001P3EvalModResidualPrimary, "primary_required_base_ancestor": primaryBaseAncestor,
			"secondary_commit": secondaryCommit, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3EvalModResidualSecondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true,
			"logn13_only": cfg.LogN == 13, "accepted_g0_guard_bits": 2, "accepted_f0_guard_bits": 3, "all_rounds_da_fixed": true,
			"no_candidate_retuning": true, "no_q_parameter_changes": true, "no_s2c_changes": true, "no_generated_power_redesign": true, "no_production_integration": true,
			"no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true,
		},
	}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3EvalModResidualSecondary {
		result.Classification, result.FirstRemainingBlocker = "logn13_evalmod_residual_decomposition_precondition_mismatch", "startup provenance or Secondary precondition"
		return fix001P3EvalModResidualWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	acceptedRow, candidate, err := fix001P3G0RunCandidate(profile, 2)
	if err != nil {
		return err
	}
	if !acceptedRow.ValidLocalQ2 || candidate.realFinal == nil || candidate.imagFinal == nil {
		result.Classification, result.FirstRemainingBlocker = "logn13_evalmod_residual_decomposition_precondition_mismatch", "accepted G0=2/F0=3 candidate replay"
		return fix001P3EvalModResidualWrite(result, outPath)
	}
	realPath, imagPath, err := fix001P3S2CAcceptedPaths(profile, candidate)
	if err != nil {
		return err
	}
	if realPath.FirstFailure != "none" || imagPath.FirstFailure != "none" || realPath.FastFinal == nil || imagPath.FastFinal == nil {
		result.Classification, result.FirstRemainingBlocker = "logn13_evalmod_residual_decomposition_precondition_mismatch", fmt.Sprintf("accepted DA path real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)
		return fix001P3EvalModResidualWrite(result, outPath)
	}
	params := profile.BTP.BootstrappingParameters
	realFinal, err := evalModMatchedFinalEvidence(realPath, params, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, params, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realPath, imagPath, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	result.E0 = map[string]interface{}{
		"accepted_candidate": q056CompactCandidate(candidate), "real_evalmod_vs_standard": realFinal.FastVsStandard, "imag_evalmod_vs_standard": imagFinal.FastVsStandard,
		"post_s2c_vs_standard": publicEvidence.PostS2C, "public_like": publicEvidence.PublicLike, "final_metadata": publicEvidence.Metadata,
		"no_capacity_or_row_failure": realPath.FirstFailure == "none" && imagPath.FirstFailure == "none" && realFinal.RestoreRowsMatch && imagFinal.RestoreRowsMatch,
	}
	if publicEvidence.PublicLike == nil || !fix001P3EvalModResidualWithin(realFinal.FastVsStandard.MaxComponent, 6.261306805777143e-5, 2e-5) || !fix001P3EvalModResidualWithin(imagFinal.FastVsStandard.MaxComponent, 4.647793870173522e-5, 2e-5) || !fix001P3EvalModResidualWithin(publicEvidence.PostS2C.MaxComponent, 3.3385298628089955e-4, 2e-5) {
		result.Classification, result.FirstRemainingBlocker = "logn13_evalmod_residual_decomposition_precondition_mismatch", "E0 accepted control mismatch"
		return fix001P3EvalModResidualWrite(result, outPath)
	}

	pfReal, err := fix001P3EvalModResidualDecode(params, candidate.realFinal)
	if err != nil {
		return err
	}
	pfImag, err := fix001P3EvalModResidualDecode(params, candidate.imagFinal)
	if err != nil {
		return err
	}
	poRealCT, err := fix001P3EvalModResidualHighPrecisionOracleCipher(params, candidate.realFinal, realPath.PlainOracle)
	if err != nil {
		return err
	}
	poImagCT, err := fix001P3EvalModResidualHighPrecisionOracleCipher(params, candidate.imagFinal, imagPath.PlainOracle)
	if err != nil {
		return err
	}
	poReal, err := fix001P3EvalModResidualFullDecode(params, poRealCT)
	if err != nil {
		return err
	}
	poImag, err := fix001P3EvalModResidualFullDecode(params, poImagCT)
	if err != nil {
		return err
	}
	result.Polynomial = map[string]interface{}{
		"real_p_f_minus_p_o":    fix001P3EvalModResidualMetric(realPath.PlainOracle, pfReal, fix001P3EvalModResidualStageThreshold),
		"imag_p_f_minus_p_o":    fix001P3EvalModResidualMetric(imagPath.PlainOracle, pfImag, fix001P3EvalModResidualStageThreshold),
		"oracle_injection_real": fix001P3EvalModResidualMetric(realPath.PlainOracle, poReal, fix001P3EvalModResidualClosureTolerance),
		"oracle_injection_imag": fix001P3EvalModResidualMetric(imagPath.PlainOracle, poImag, fix001P3EvalModResidualClosureTolerance),
		"real_input_metadata":   map[string]interface{}{"level": candidate.realFinal.Level(), "scale": finalizationScaleString(candidate.realFinal.Scale), "degree": candidate.realFinal.Degree()},
		"imag_input_metadata":   map[string]interface{}{"level": candidate.imagFinal.Level(), "scale": finalizationScaleString(candidate.imagFinal.Scale), "degree": candidate.imagFinal.Degree()},
	}
	oracleRealInjection := result.Polynomial["oracle_injection_real"].(*PSGlobalMetric)
	oracleImagInjection := result.Polynomial["oracle_injection_imag"].(*PSGlobalMetric)
	pfPoReal := result.Polynomial["real_p_f_minus_p_o"].(*PSGlobalMetric)
	pfPoImag := result.Polynomial["imag_p_f_minus_p_o"].(*PSGlobalMetric)
	oracleInjectionPass := oracleRealInjection.MaxComponent <= fix001P3EvalModResidualClosureTolerance && oracleImagInjection.MaxComponent <= fix001P3EvalModResidualClosureTolerance && oracleRealInjection.MaxComponent*10 <= pfPoReal.MaxComponent && oracleImagInjection.MaxComponent*10 <= pfPoImag.MaxComponent
	result.Validation["oracle_injection_ten_x_below_pf_po"] = oracleInjectionPass
	if !oracleInjectionPass {
		result.Classification, result.FirstRemainingBlocker = "logn13_evalmod_residual_oracle_injection_mismatch", "P_O full-RNS materialization floor is not <=1e-10 and 10x below P_F-P_O"
		return fix001P3EvalModResidualWrite(result, outPath)
	}
	fastRealRun, err := fix001P3EvalModResidualRunState(profile, profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, candidate.realFinal, candidate)
	if err != nil {
		return err
	}
	fastImagRun, err := fix001P3EvalModResidualRunState(profile, profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, candidate.imagFinal, candidate)
	if err != nil {
		return err
	}
	oracleRealRun, err := fix001P3EvalModResidualFullNormalizedRun(params, poRealCT, fastRealRun.WorkingScale, fastRealRun.InitialK, fastRealRun.InitialSqrt, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	oracleImagRun, err := fix001P3EvalModResidualFullNormalizedRun(params, poImagCT, fastImagRun.WorkingScale, fastImagRun.InitialK, fastImagRun.InitialSqrt, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealRealStages, idealReal, err := fix001P3EvalModResidualIdealNormalized(realPath.PlainOracle, fastRealRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealFRealStages, idealFReal, err := fix001P3EvalModResidualIdealNormalized(pfReal, fastRealRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealImagStages, idealImag, err := fix001P3EvalModResidualIdealNormalized(imagPath.PlainOracle, fastImagRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	idealFImagStages, idealFImag, err := fix001P3EvalModResidualIdealNormalized(pfImag, fastImagRun, params, profile.Fast.Mod1Parameters.DoubleAngle)
	if err != nil {
		return err
	}
	result.Checkpoints, err = fix001P3EvalModResidualCheckpointTable(params, fastRealRun, oracleRealRun, idealRealStages)
	if err != nil {
		return err
	}
	imagCheckpoints, err := fix001P3EvalModResidualCheckpointTable(params, fastImagRun, oracleImagRun, idealImagStages)
	if err != nil {
		return err
	}
	for _, row := range imagCheckpoints {
		row["branch"] = "imag"
		result.Checkpoints = append(result.Checkpoints, row)
	}
	for _, row := range result.Checkpoints[:len(result.Checkpoints)-len(imagCheckpoints)] {
		row["branch"] = "real"
	}

	efReal, err := fix001P3EvalModResidualDecode(params, realPath.FastFinal)
	if err != nil {
		return err
	}
	efImag, err := fix001P3EvalModResidualDecode(params, imagPath.FastFinal)
	if err != nil {
		return err
	}
	nfReal, err := fix001P3EvalModResidualDecode(params, fastRealRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	nfImag, err := fix001P3EvalModResidualDecode(params, fastImagRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	noReal, err := fix001P3EvalModResidualFullDecode(params, oracleRealRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	noImag, err := fix001P3EvalModResidualFullDecode(params, oracleImagRun.State.Path.NormalizedFinal)
	if err != nil {
		return err
	}
	_, esRealValues, err := semanticBisectView(realPath.StandardFinal, params, profile.StandardSK)
	if err != nil {
		return err
	}
	_, esImagValues, err := semanticBisectView(imagPath.StandardFinal, params, profile.StandardSK)
	if err != nil {
		return err
	}
	ef, nf, no, io, es := fix001P3EvalModResidualCombine(efReal, efImag), fix001P3EvalModResidualCombine(nfReal, nfImag), fix001P3EvalModResidualCombine(noReal, noImag), fix001P3EvalModResidualCombine(idealReal, idealImag), fix001P3EvalModResidualCombine(esRealValues, esImagValues)
	implVector := fix001P3EvalModResidualSub(ef, nf)
	psVector := fix001P3EvalModResidualSub(nf, no)
	daVector := fix001P3EvalModResidualSub(no, io)
	floorVector := fix001P3EvalModResidualSub(io, es)
	totalVector := fix001P3EvalModResidualSub(ef, es)
	closureVector := fix001P3EvalModResidualSub(fix001P3EvalModResidualAdd(fix001P3EvalModResidualAdd(fix001P3EvalModResidualAdd(implVector, psVector), daVector), floorVector), totalVector)
	result.Closure = fix001P3EvalModResidualMetric(make([]complex128, len(closureVector)), closureVector, fix001P3EvalModResidualClosureTolerance)
	result.Causal = map[string]interface{}{
		"e_f_minus_n_f_fast_local_q2_implementation": fix001P3EvalModResidualMetric(make([]complex128, len(implVector)), implVector, fix001P3EvalModResidualPublicThreshold),
		"n_f_minus_n_o_ps_input_propagation":         fix001P3EvalModResidualMetric(make([]complex128, len(psVector)), psVector, fix001P3EvalModResidualPublicThreshold),
		"n_o_minus_i_o_normalized_da_arithmetic":     fix001P3EvalModResidualMetric(make([]complex128, len(daVector)), daVector, fix001P3EvalModResidualPublicThreshold),
		"i_o_minus_e_s_reference_alignment_floor":    fix001P3EvalModResidualMetric(make([]complex128, len(floorVector)), floorVector, fix001P3EvalModResidualPublicThreshold),
		"observed_total_e_f_minus_e_s":               fix001P3EvalModResidualMetric(make([]complex128, len(totalVector)), totalVector, fix001P3EvalModResidualPublicThreshold),
	}
	allContracts := fastRealRun.State.Path.FastRowsMatch && fastRealRun.State.Path.RestoreRowsMatch && fastImagRun.State.Path.FastRowsMatch && fastImagRun.State.Path.RestoreRowsMatch
	baselineCoreCT, err := profile.Standard.SlotsToCoeffs(realPath.StandardFinal.CopyNew(), imagPath.StandardFinal.CopyNew())
	if err != nil {
		return err
	}
	baselineCore, err := decodeWithSecret(params, baselineCoreCT, profile.StandardSK)
	if err != nil {
		return err
	}
	baselineEvalMod := es
	input := reproducibleInput(profile.Residual, profile.BTP)
	cases := map[string]map[string]interface{}{}
	caseValues := map[string][2][]complex128{"H0": {efReal, efImag}, "H_PS": {noReal, noImag}, "H_DA": {idealFReal, idealFImag}, "H_BOTH": {idealReal, idealImag}}
	for name, values := range caseValues {
		caseResult, err := fix001P3EvalModResidualDownstream(profile, baselineCore, baselineEvalMod, input, realPath.FastFinal, imagPath.FastFinal, values[0], values[1])
		if err != nil {
			return err
		}
		delete(caseResult, "core_values")
		delete(caseResult, "final_values")
		cases[name] = caseResult
	}
	result.Counterfactuals = cases
	projected := map[string]interface{}{}
	for name, vector := range map[string][]complex128{"fast_implementation": implVector, "ps_input": psVector, "da_arithmetic": daVector, "reference_floor": floorVector} {
		projectedValues, err := fix001P3EvalModResidualProjectS2C(profile, realPath.FastFinal, imagPath.FastFinal, vector)
		if err != nil {
			return err
		}
		projected[name] = fix001P3EvalModResidualMetric(make([]complex128, len(projectedValues)), projectedValues, fix001P3EvalModResidualPostS2CThreshold)
	}
	actualCore := cases["H0"]
	_ = actualCore
	result.Projected = projected
	_ = idealFRealStages
	_ = idealFImagStages
	hPS := cases["H_PS"]["public_like"].(*PSGlobalMetric).Pass
	hDA := cases["H_DA"]["public_like"].(*PSGlobalMetric).Pass
	hBoth := cases["H_BOTH"]["public_like"].(*PSGlobalMetric).Pass
	implMetric := result.Causal["e_f_minus_n_f_fast_local_q2_implementation"].(*PSGlobalMetric)
	result.Classification, result.FirstRemainingBlocker = fix001P3EvalModResidualClassify(hPS, hDA, hBoth, implMetric.MaxComponent > fix001P3EvalModResidualClosureTolerance, allContracts)
	result.Validation["e0_control_pass"] = true
	result.Validation["oracle_injection_real_pass"] = result.Polynomial["oracle_injection_real"].(*PSGlobalMetric).MaxComponent <= fix001P3EvalModResidualClosureTolerance
	result.Validation["oracle_injection_imag_pass"] = result.Polynomial["oracle_injection_imag"].(*PSGlobalMetric).MaxComponent <= fix001P3EvalModResidualClosureTolerance
	result.Validation["fast_full_rns_rows_match_all_checkpoints"] = allContracts
	result.Validation["closure_pass"] = result.Closure.MaxComponent <= fix001P3EvalModResidualClosureTolerance
	result.Validation["h_ps_public_like_pass"] = hPS
	result.Validation["h_da_public_like_pass"] = hDA
	result.Validation["h_both_public_like_pass"] = hBoth
	result.Validation["required_post_s2c_threshold"] = fix001P3EvalModResidualPostS2CThreshold
	result.Validation["public_like_threshold"] = fix001P3EvalModResidualPublicThreshold
	return fix001P3EvalModResidualWrite(result, outPath)
}

func fix001P3EvalModResidualWithin(actual, expected, tolerance float64) bool {
	delta := actual - expected
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance
}
