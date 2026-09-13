package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3DAAllRoundsPrimaryBase = "5070ca40194f29d75a0f5aa79e64905ec1af821b"
	requiredFIX001P3DAAllRoundsSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
)

type fix001P3DAAllRoundsResult struct {
	SchemaVersion    string                   `json:"schema_version"`
	Timestamp        time.Time                `json:"timestamp"`
	Primary          RepositoryMetadata       `json:"primary_repository"`
	Lattigo          RepositoryMetadata       `json:"lattigo_repository"`
	Environment      EnvironmentMetadata      `json:"environment"`
	Config           BootstrapConfig          `json:"config"`
	Parameters       ExperimentParameters     `json:"effective_parameters"`
	A0Control        map[string]interface{}   `json:"a0_control"`
	PerRound         []map[string]interface{} `json:"per_round"`
	Final            map[string]interface{}   `json:"final_metrics,omitempty"`
	Operations       map[string]interface{}   `json:"static_q2_costs"`
	Classification   string                   `json:"classification"`
	SystemSufficient bool                     `json:"ps_da_arithmetic_system_sufficient"`
	FirstBlocker     string                   `json:"first_remaining_blocker"`
	Validation       map[string]interface{}   `json:"validation"`
}

type fix001P3DAAllRoundsRoundOutcome struct {
	Evidence       map[string]interface{}
	Classification string
	Blocker        string
}

func fix001P3DAAllRoundsWrite(result fix001P3DAAllRoundsResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3DAAllRoundsCapacityTable(params ckks.Parameters, checkpoints ...struct {
	Name string
	CT   *rlwe.Ciphertext
}) []map[string]interface{} {
	table := make([]map[string]interface{}, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		row := map[string]interface{}{"checkpoint": checkpoint.Name}
		if checkpoint.CT != nil {
			if q01, err := fix001P3DAR0Capacity(params, checkpoint.CT, 1); err == nil {
				row["q01"] = q01
			}
			if checkpoint.CT.Level() >= 2 {
				if q012, err := fix001P3DAR0Capacity(params, checkpoint.CT, 2); err == nil {
					row["q012"] = q012
				}
			}
		}
		table = append(table, row)
	}
	return table
}

func fix001P3DAAllRoundsExpansion(params ckks.Parameters, source, intended *rlwe.Ciphertext) (map[string]interface{}, error) {
	signed, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, err
	}
	expanded, err := fix001P3DAR0ExpandQ2(params, source, signed)
	if err != nil {
		return nil, err
	}
	before, err := psGlobalDecode(params, source)
	if err != nil {
		return nil, err
	}
	after, err := fix001P3LocalQ2Decode(params, source, signed, source.Scale)
	if err != nil {
		return nil, err
	}
	semantic := psRescaleGuardMetric(before, after)
	q01Rows := fix001P3DAR0RowsAtLimb(params, expanded, source, 0) && fix001P3DAR0RowsAtLimb(params, expanded, source, 1)
	q2Rows := fix001P3DAR0RowsAtLimb(params, expanded, intended, 2)
	return map[string]interface{}{"input_q01_centered_unique": true, "q0_q1_rows_unchanged_bit_for_bit": q01Rows, "q2_matches_stage_aligned_full_rns": q2Rows, "semantic_change": semantic, "pass": semantic.Pass && q01Rows && q2Rows}, nil
}

func fix001P3DAAllRoundsRound(state *fix001P3DAR0LocalQ2State, round int, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey) (fix001P3DAAllRoundsRoundOutcome, error) {
	outcome := fix001P3DAAllRoundsRoundOutcome{Evidence: map[string]interface{}{"round": round}}
	sqrt2pi := state.Sqrt2Pi * state.Sqrt2Pi
	schedule, err := normalizedScaleSchedule(round, state.CurrentLevel, state.CurrentScale, state.WorkingScale, state.KIn, sqrt2pi, params)
	if err != nil {
		return outcome, err
	}
	inputCheckpoint, err := evalModMatchedCheckpointFor("da_input", round, state.FastNormalized, state.NormalizedCurrent, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return outcome, err
	}
	inputCap, err := fix001P3DAR0Capacity(params, state.NormalizedCurrent, 1)
	if err != nil {
		return outcome, err
	}
	if !inputCap.Unique || !inputCheckpoint.NormalizedSafe {
		outcome.Classification = "logn13_da_all_rounds_q01_round_input_capacity_blocker"
		outcome.Blocker = fmt.Sprintf("round%d.input", round)
		outcome.Evidence["activation"] = "none"
		outcome.Evidence["capacity_table"] = fix001P3DAAllRoundsCapacityTable(params, struct {
			Name string
			CT   *rlwe.Ciphertext
		}{"input", state.NormalizedCurrent})
		return outcome, nil
	}
	refProduct, err := normalizedFullSquare(params, state.NormalizedCurrent)
	if err != nil {
		return outcome, err
	}
	refSquare := normalizedFullZeroC2(params, refProduct)
	if err := fastEval.FastCKKS.MulRelin(state.FastNormalized, state.FastNormalized, state.FastNormalized); err != nil {
		return outcome, err
	}
	squareCheckpoint, err := evalModMatchedCheckpointFor("da_square", round, state.FastNormalized, refSquare, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return outcome, err
	}
	squareCap, err := fix001P3DAR0Capacity(params, refSquare, 1)
	if err != nil {
		return outcome, err
	}
	activation := "post_square"
	expansionSource := state.FastNormalized
	expansionIntended := refSquare
	if !squareCap.Unique || !squareCheckpoint.NormalizedSafe {
		if !inputCap.Unique {
			outcome.Classification = "logn13_da_all_rounds_q01_round_input_capacity_blocker"
			outcome.Blocker = fmt.Sprintf("round%d.input", round)
			return outcome, nil
		}
		activation = "pre_square"
		expansionSource = state.FastNormalized.CopyNew()
		expansionIntended = state.NormalizedCurrent
	}
	expansion, err := fix001P3DAAllRoundsExpansion(params, expansionSource, expansionIntended)
	if err != nil {
		return outcome, err
	}
	if !expansion["pass"].(bool) {
		outcome.Classification = "logn13_da_all_rounds_local_q2_expansion_mismatch"
		outcome.Blocker = fmt.Sprintf("round%d.%s.expansion", round, activation)
		outcome.Evidence["activation"] = activation
		outcome.Evidence["expansion"] = expansion
		return outcome, nil
	}
	factor := new(big.Int).Lsh(big.NewInt(1), uint(schedule.AExponent))
	refAfterMultiplier := normalizedFullIntegerMultiply(params, refSquare, factor)
	q01AfterMultiplier, err := fix001P3DAR0ContractQ01(params, refAfterMultiplier)
	if err != nil {
		return outcome, err
	}
	multiplierCheckpoint, err := evalModMatchedCheckpointFor("da_after_multiplier", round, q01AfterMultiplier, refAfterMultiplier, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return outcome, err
	}
	constantValue, _ := new(big.Float).SetPrec(doubleAnglePrecision).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(schedule.KOutExponent)))).Float64()
	refAfterConstant := normalizedFullSubtractConstant(params, refAfterMultiplier, constantValue)
	q01AfterConstant, err := fix001P3DAR0ContractQ01(params, refAfterConstant)
	if err != nil {
		return outcome, err
	}
	constantCheckpoint, err := evalModMatchedCheckpointFor("da_after_constant", round, q01AfterConstant, refAfterConstant, state.CoherentCurrent, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return outcome, err
	}
	refAfterRescale, err := normalizedFullRescale(params, refAfterConstant)
	if err != nil {
		return outcome, err
	}
	q01AfterRescale, err := fix001P3DAR0ContractQ01(params, refAfterRescale)
	if err != nil {
		return outcome, err
	}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[refAfterConstant.Level()].Modulus)
	roundedMatch, err := fix001P3DAR0RoundedDivisionMatches(params, refAfterConstant, refAfterRescale, divisor)
	if err != nil {
		return outcome, err
	}
	q012AfterRescale, err := fix001P3DAR0Capacity(params, refAfterRescale, 2)
	if err != nil {
		return outcome, err
	}
	q01AfterRescaleCap, err := fix001P3DAR0Capacity(params, refAfterRescale, 1)
	if err != nil {
		return outcome, err
	}
	capacityTable := fix001P3DAAllRoundsCapacityTable(params,
		struct {
			Name string
			CT   *rlwe.Ciphertext
		}{"input", state.NormalizedCurrent},
		struct {
			Name string
			CT   *rlwe.Ciphertext
		}{"square", refSquare},
		struct {
			Name string
			CT   *rlwe.Ciphertext
		}{"after_multiplier", refAfterMultiplier},
		struct {
			Name string
			CT   *rlwe.Ciphertext
		}{"after_constant", refAfterConstant},
		struct {
			Name string
			CT   *rlwe.Ciphertext
		}{"post_rescale", refAfterRescale},
	)
	if !multiplierCheckpoint.RowsMatch || !constantCheckpoint.RowsMatch || !q012AfterRescale.Unique {
		outcome.Classification = "logn13_da_all_rounds_local_q2_q012_capacity_failure"
		outcome.Blocker = fmt.Sprintf("round%d.q012", round)
	}
	if !roundedMatch {
		outcome.Classification = "logn13_da_all_rounds_local_q2_rescale_mismatch"
		outcome.Blocker = fmt.Sprintf("round%d.post_rescale.rounded_division", round)
	}
	contractDecoded, err := psGlobalDecode(params, q01AfterRescale)
	if err != nil {
		return outcome, err
	}
	q012Decoded, err := fix001P3LocalQ2Decode(params, refAfterRescale, mustFIX001P3DAR0Signed(params, refAfterRescale, 2), refAfterRescale.Scale)
	if err != nil {
		return outcome, err
	}
	contractMetric := psRescaleGuardMetric(q012Decoded, contractDecoded)
	contractRows := fix001P3DAR0RowsAtLimb(params, q01AfterRescale, refAfterRescale, 0) && fix001P3DAR0RowsAtLimb(params, q01AfterRescale, refAfterRescale, 1)
	contractPass := q01AfterRescaleCap.Unique && contractRows && contractMetric.Pass && q01AfterRescale.Level() == refAfterRescale.Level() && q01AfterRescale.Scale.Equal(refAfterRescale.Scale) && q01AfterRescale.IsNTT && q01AfterRescale.IsMontgomery
	if !contractPass && outcome.Classification == "" {
		outcome.Classification = "logn13_da_all_rounds_local_q2_cannot_contract_to_q01"
		outcome.Blocker = fmt.Sprintf("round%d.post_rescale.contraction", round)
	}
	outcome.Evidence["activation"] = activation
	outcome.Evidence["schedule"] = map[string]interface{}{"k_in_exponent": schedule.KInExponent, "k_out_exponent": schedule.KOutExponent, "multiplier_exponent": schedule.AExponent, "constant": constantValue}
	outcome.Evidence["input_rows_match"] = inputCheckpoint.RowsMatch
	outcome.Evidence["square_rows_match"] = squareCheckpoint.RowsMatch
	outcome.Evidence["expansion"] = expansion
	outcome.Evidence["capacity_table"] = capacityTable
	outcome.Evidence["multiplier_rows_match"] = multiplierCheckpoint.RowsMatch
	outcome.Evidence["constant_rows_match"] = constantCheckpoint.RowsMatch
	outcome.Evidence["rounded_division_match"] = roundedMatch
	outcome.Evidence["no_extra_level"] = refAfterRescale.Level() == refAfterConstant.Level()-params.LevelsConsumedPerRescaling()
	outcome.Evidence["contraction"] = map[string]interface{}{"q01_centered_unique": q01AfterRescaleCap.Unique, "q0_q1_rows_match_q012": contractRows, "semantic": contractMetric, "metadata_match": q01AfterRescale.Level() == refAfterRescale.Level() && q01AfterRescale.Scale.Equal(refAfterRescale.Scale) && q01AfterRescale.IsNTT && q01AfterRescale.IsMontgomery, "pass": contractPass}
	state.Path.Rounds = append(state.Path.Rounds, evalModMatchedRound{Round: round, KInExponent: schedule.KInExponent, KOutExponent: schedule.KOutExponent, MultiplierExponent: schedule.AExponent, NormalizedConstant: constantValue, Input: inputCheckpoint, Square: squareCheckpoint, AfterMultiplier: multiplierCheckpoint, AfterConstant: constantCheckpoint})
	if outcome.Classification != "" {
		return outcome, nil
	}
	postCheckpoint, err := evalModMatchedCheckpointFor("da_post_rescale", round, q01AfterRescale, refAfterRescale, refAfterRescale, params, fastSK, zeroSecret(params), zeroSecret(params))
	if err != nil {
		return outcome, err
	}
	state.Path.Rounds[len(state.Path.Rounds)-1].PostRescale = postCheckpoint
	if !postCheckpoint.RowsMatch || !postCheckpoint.NormalizedSafe {
		outcome.Classification = "logn13_da_all_rounds_local_q2_q012_capacity_failure"
		outcome.Blocker = fmt.Sprintf("round%d.post_rescale", round)
		return outcome, nil
	}
	state.FastNormalized = q01AfterRescale
	state.NormalizedCurrent = refAfterRescale
	coherentPre, _, err := normalizedFullSquareStep(params, state.CoherentCurrent, big.NewInt(2), sqrt2pi)
	if err != nil {
		return outcome, err
	}
	state.CoherentCurrent, err = normalizedFullRescale(params, coherentPre)
	if err != nil {
		return outcome, err
	}
	if err := standardEval.Evaluator.MulRelin(state.StandardCurrent, state.StandardCurrent, state.StandardCurrent); err != nil {
		return outcome, err
	}
	if err := standardEval.Evaluator.Add(state.StandardCurrent, state.StandardCurrent, state.StandardCurrent); err != nil {
		return outcome, err
	}
	if err := standardEval.Evaluator.Add(state.StandardCurrent, -sqrt2pi, state.StandardCurrent); err != nil {
		return outcome, err
	}
	if err := standardEval.Evaluator.Rescale(state.StandardCurrent, state.StandardCurrent); err != nil {
		return outcome, err
	}
	state.Sqrt2Pi = sqrt2pi
	state.CurrentLevel = schedule.NextLevel
	state.CurrentScale = scheduleScaleValue(schedule, params, state.CurrentLevel, state.CurrentScale)
	state.KIn = schedule.KOutExponent
	return outcome, nil
}

func fix001P3DAAllRoundsFinalize(state *fix001P3DAR0LocalQ2State, fastEval *bootstrapping.FastEvaluator, params ckks.Parameters) error {
	fastBeforeRestore := state.FastNormalized.CopyNew()
	normalizedBeforeRestore := state.NormalizedCurrent.CopyNew()
	finalK := new(big.Int).Lsh(big.NewInt(1), uint(state.KIn))
	beforeRestoreCapacity, err := normalizedExactCapacity(params, state.NormalizedCurrent)
	if err != nil {
		return err
	}
	restoreView, err := targetScaleCoefficientView(params, fastBeforeRestore)
	if err != nil {
		return err
	}
	restoreData, err := targetScaleCenteredRows(params, restoreView)
	if err != nil {
		return err
	}
	restoreCapacity := normalizedFinalRestoreCapacity(restoreData, finalK)
	state.Path.NormalizedBeforeRestoreCapacity = &beforeRestoreCapacity
	state.Path.FinalRestoreCapacity = &restoreCapacity
	if !beforeRestoreCapacity.Pass || !restoreCapacity.Pass {
		state.Path.FirstFailure = "final.restore_capacity"
		return nil
	}
	if err := fastEval.FastCKKS.MulIntegerMaintained(state.FastNormalized, finalK, state.FastNormalized); err != nil {
		return err
	}
	state.NormalizedCurrent = normalizedFullIntegerMultiply(params, state.NormalizedCurrent, finalK)
	fastRestoreRows := finalizationEvidence(state.FastNormalized).Rows
	normalizedRestoreRows := finalizationEvidence(state.NormalizedCurrent).Rows
	state.Path.FastRowsMatch = doubleAngleRowsMatch(fastRestoreRows, normalizedRestoreRows)
	state.Path.RestoreRowsMatch = state.Path.FastRowsMatch
	state.FastNormalized.Scale = state.OriginalFastScale
	state.NormalizedCurrent.Scale = state.OriginalFastScale
	state.StandardCurrent.Scale = state.OriginalStdScale
	state.CoherentCurrent.Scale = state.OriginalStdScale
	state.Path.ScaleResetRowsUnchanged = doubleAngleRowsMatch(fastRestoreRows, finalizationEvidence(state.FastNormalized).Rows) && doubleAngleRowsMatch(normalizedRestoreRows, finalizationEvidence(state.NormalizedCurrent).Rows)
	state.Path.LevelDegreeUnchanged = state.FastNormalized.Level() == fastBeforeRestore.Level() && state.FastNormalized.Degree() == fastBeforeRestore.Degree() && state.NormalizedCurrent.Level() == normalizedBeforeRestore.Level() && state.NormalizedCurrent.Degree() == normalizedBeforeRestore.Degree()
	state.Path.FastFinal, state.Path.NormalizedFinal, state.Path.CoherentFinal, state.Path.StandardFinal = state.FastNormalized, state.NormalizedCurrent, state.CoherentCurrent, state.StandardCurrent
	return nil
}

func runFIX001P3DesignLogN13DAAllRoundsLocalQ2(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3DAAllRoundsPrimaryBase, "HEAD") == requiredFIX001P3DAAllRoundsPrimaryBase
	result := fix001P3DAAllRoundsResult{SchemaVersion: "fix-001-p3-design-logn13-da-all-rounds-local-q2.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Operations: map[string]interface{}{"round_count": 3, "coefficient_transforms": 0, "q2_ntt": 0, "q2_intt": 0, "q2_limb_arithmetic_passes": 0, "q2_squares": 0, "q2_multipliers": 0, "q2_constants": 0, "q2_rescales": 0, "q2_discarded_after_each_valid_round": true}, Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3DAAllRoundsPrimaryBase, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_commit": secondaryCommit, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3DAAllRoundsSecondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "logn13_only": cfg.LogN == 13, "fixed_k3_candidate": true, "q0_q1_q2_profile": "56/39/40", "plan_scale_2^92": true, "g0_guard_bits": 1, "f0_local_q2_guard_bits": 3, "q2_bounded_per_round": true, "no_q0_q1_widening": true, "no_q3_plus": true, "no_extra_q_levels": true, "no_ps_retuning": true, "no_generated_power_redesign": true, "no_c2s_s2c_mod1_change": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3DAAllRoundsSecondary {
		result.Classification, result.FirstBlocker = "logn13_da_all_rounds_local_q2_precondition_mismatch", "startup provenance or Secondary precondition"
		return fix001P3DAAllRoundsWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	var realEvidence, imagEvidence fix001P3LocalQ2BranchEvidence
	candidate, _, _, err := psRescaleGuardMakeCandidateWithOverrides(profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch, profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, 3}, "D0-K3-G0-one-local-q2-F0-three", &profile.InitialReal, &profile.InitialImag, profile.Boundaries, fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &realEvidence, 3), fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &imagEvidence, 3))
	if err != nil {
		return err
	}
	planScale := precisionSweepScale(92)
	realControl, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return err
	}
	imagControl, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return err
	}
	result.A0Control = map[string]interface{}{"real": fix001P3DAR0Control(realControl), "imag": fix001P3DAR0Control(imagControl), "fixed_candidate": q056CompactCandidate(candidate), "round0_q2_evidence_from_previous_task": true}
	if !fix001P3DAR0Control(realControl)["reproduced"].(bool) || !fix001P3DAR0Control(imagControl)["reproduced"].(bool) {
		result.Classification, result.FirstBlocker = "logn13_da_all_rounds_local_q2_precondition_mismatch", "A0 round0 control not reproduced"
		return fix001P3DAAllRoundsWrite(result, outPath)
	}
	realState, err := fix001P3DAR0PrepareState(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.realFinal)
	if err != nil {
		return err
	}
	imagState, err := fix001P3DAR0PrepareState(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, candidate.imagFinal)
	if err != nil {
		return err
	}
	result.PerRound = make([]map[string]interface{}, 0, 3)
	for round := 0; round < profile.Fast.Mod1Parameters.DoubleAngle; round++ {
		realOutcome, err := fix001P3DAAllRoundsRound(&realState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return err
		}
		imagOutcome, err := fix001P3DAAllRoundsRound(&imagState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return err
		}
		result.PerRound = append(result.PerRound, map[string]interface{}{"round": round, "real": realOutcome.Evidence, "imag": imagOutcome.Evidence})
		for _, outcome := range []fix001P3DAAllRoundsRoundOutcome{realOutcome, imagOutcome} {
			if outcome.Classification != "" {
				result.Classification, result.FirstBlocker = outcome.Classification, outcome.Blocker
				return fix001P3DAAllRoundsWrite(result, outPath)
			}
		}
		result.Operations["coefficient_transforms"] = result.Operations["coefficient_transforms"].(int) + 2
		result.Operations["q2_ntt"] = result.Operations["q2_ntt"].(int) + 2
		result.Operations["q2_limb_arithmetic_passes"] = result.Operations["q2_limb_arithmetic_passes"].(int) + 6
		result.Operations["q2_squares"] = result.Operations["q2_squares"].(int) + 2
		result.Operations["q2_multipliers"] = result.Operations["q2_multipliers"].(int) + 2
		result.Operations["q2_constants"] = result.Operations["q2_constants"].(int) + 2
		result.Operations["q2_rescales"] = result.Operations["q2_rescales"].(int) + 2
	}
	if err := fix001P3DAAllRoundsFinalize(&realState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return err
	}
	if err := fix001P3DAAllRoundsFinalize(&imagState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return err
	}
	if realState.Path.FastFinal == nil || imagState.Path.FastFinal == nil || !realState.Path.FastRowsMatch || !realState.Path.RestoreRowsMatch || !imagState.Path.FastRowsMatch || !imagState.Path.RestoreRowsMatch {
		result.Classification, result.FirstBlocker = "logn13_da_all_rounds_local_q2_downstream_insufficient", "final restore row/metadata contract"
		return fix001P3DAAllRoundsWrite(result, outPath)
	}
	realFinal, err := evalModMatchedFinalEvidence(realState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realState.Path, imagState.Path, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	result.Final = map[string]interface{}{"evalmod_real_vs_standard": realFinal.FastVsStandard, "evalmod_imag_vs_standard": imagFinal.FastVsStandard, "post_s2c_fast_vs_standard": publicEvidence.PostS2C, "unpack_finalization_fast_vs_standard_core": publicEvidence.FinalVsStdCore, "standard_like": publicEvidence.StandardLike, "public_like": publicEvidence.PublicLike, "final_metadata": publicEvidence.Metadata, "input_unchanged": publicEvidence.InputUnchanged, "local_polynomial_budget": map[string]interface{}{"target": psRescaleGuardBudget, "real": candidate.FinalRealError, "imag": candidate.FinalImagError, "pass": candidate.FinalRealError <= psRescaleGuardBudget && candidate.FinalImagError <= psRescaleGuardBudget, "explicitly_not_system_target": true}}
	result.SystemSufficient = publicEvidence.PublicLike != nil && publicEvidence.PublicLike.Pass && publicEvidence.Metadata["pass"] == true && publicEvidence.InputUnchanged && realFinal.FastVsStandard.Pass && imagFinal.FastVsStandard.Pass
	if result.SystemSufficient {
		result.Classification, result.FirstBlocker = "logn13_da_all_rounds_local_q2_downstream_validated", "generated-power semantic error remains separately proven"
	} else {
		result.Classification, result.FirstBlocker = "logn13_da_all_rounds_local_q2_downstream_insufficient", "all DA rounds completed but downstream public-like contract failed"
	}
	return fix001P3DAAllRoundsWrite(result, outPath)
}
