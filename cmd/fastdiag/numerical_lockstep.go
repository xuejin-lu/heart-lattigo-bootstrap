package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
)

const (
	numericalObservableThreshold = 1e-8
	numericalMaterialThreshold   = 0.1 * 0.03640730397217224
)

type numericalStageSample struct {
	name       string
	state      NumericalStageState
	ciphertext *rlwe.Ciphertext
	values     []complex128
	comparable bool
	reason     string
}

type numericalStagePair struct {
	fast     numericalStageSample
	standard numericalStageSample
}

func runNumericalStageLockstep(
	fastEval *bootstrapping.FastEvaluator,
	standardEval *bootstrapping.Evaluator,
	standardSK *rlwe.SecretKey,
	btp bootstrapping.Parameters,
	input *rlwe.Ciphertext,
	original, fastPre, standardPre, fastFinal, standardFinal []complex128,
	fastFinalCT, standardFinalCT *rlwe.Ciphertext,
) (NumericalStageLockstep, error) {
	fastSamples, err := runFastNumericalStages(fastEval, input, fastPre, fastFinalCT, fastFinal)
	if err != nil {
		return NumericalStageLockstep{}, fmt.Errorf("Fast stages: %w", err)
	}
	standardSamples, err := runStandardNumericalStages(standardEval, input, standardPre, standardFinalCT, standardFinal)
	if err != nil {
		return NumericalStageLockstep{}, fmt.Errorf("Standard stages: %w", err)
	}
	if len(fastSamples) != len(standardSamples) {
		return NumericalStageLockstep{}, fmt.Errorf("Fast/Standard checkpoint count differs: %d != %d", len(fastSamples), len(standardSamples))
	}

	pairs := make([]numericalStagePair, len(fastSamples))
	for i := range fastSamples {
		if fastSamples[i].name != standardSamples[i].name {
			return NumericalStageLockstep{}, fmt.Errorf("checkpoint order differs at %d: %q != %q", i, fastSamples[i].name, standardSamples[i].name)
		}
		pairs[i] = numericalStagePair{fast: fastSamples[i], standard: standardSamples[i]}
		if pairs[i].fast.name != "input" && pairs[i].fast.name != "final_public_output" {
			decodeNumericalStagePair(&pairs[i], btp.BootstrappingParameters, standardSK)
		}
	}

	result := summarizeNumericalStageLockstep(pairs, original)
	result.ScaleAudit = numericalScaleAudit(fastEval, standardEval, btp.BootstrappingParameters, pairs)
	result.EvalModReplayVerified = make(map[string]bool)
	result.EvalModReplayRMSE = make(map[string]float64)
	tracedBranches := 0
	for _, branch := range []string{"real", "imag"} {
		inputPair, okInput := numericalPairByName(pairs, "c2s_"+branch)
		outputPair, okOutput := numericalPairByName(pairs, "evalmod_"+branch)
		if !okInput || !okOutput || inputPair.fast.ciphertext == nil || inputPair.standard.ciphertext == nil {
			continue
		}
		trace, err := runNumericalEvalModStageReplay(
			branch,
			inputPair.fast.ciphertext,
			inputPair.standard.ciphertext,
			outputPair.fast.values,
			outputPair.standard.values,
			fastEval,
			standardEval,
			standardSK,
			btp.BootstrappingParameters,
		)
		if err != nil {
			return NumericalStageLockstep{}, fmt.Errorf("EvalMod %s stage-equivalence replay: %w", branch, err)
		}
		result.EvalModReplayVerified[branch+".fast"] = trace.fastReplayVerified
		result.EvalModReplayVerified[branch+".standard"] = trace.standardReplayVerified
		result.EvalModReplayRMSE[branch+".fast"] = trace.fastReplayRMSE
		result.EvalModReplayRMSE[branch+".standard"] = trace.standardReplayRMSE
		result.EvalModInternal = append(result.EvalModInternal, trace.checkpoints...)
		result.EvalModCapacityAudit = append(result.EvalModCapacityAudit, trace.capacityAudit...)
		result.ScaleAudit = append(result.ScaleAudit, trace.scaleAudit...)
		result.GeneratedPowerEvidence = append(result.GeneratedPowerEvidence, trace.powers...)
		if branch == "real" || result.PolynomialPlan == nil {
			plan := trace.plan
			result.PolynomialPlan = &plan
		}
		if result.FirstMaterial == "evalmod_"+branch {
			result.Classification = numericalEvalModClassification(trace)
		}
		tracedBranches++
	}
	if tracedBranches == 0 {
		return NumericalStageLockstep{}, fmt.Errorf("no comparable EvalMod branch was available for direct Standard/Fast replay")
	}
	return result, nil
}

func numericalPairByName(pairs []numericalStagePair, name string) (numericalStagePair, bool) {
	for _, pair := range pairs {
		if pair.fast.name == name {
			return pair, true
		}
	}
	return numericalStagePair{}, false
}

func numericalEvalModClassification(trace numericalEvalModTrace) string {
	previousMax := 0.0
	for _, checkpoint := range trace.checkpoints {
		if !checkpoint.Comparable || checkpoint.FastVsStandard == nil {
			continue
		}
		maxDiff := checkpoint.FastVsStandard.Complex.Max
		if previousMax < numericalMaterialThreshold && maxDiff >= numericalMaterialThreshold {
			switch {
			case strings.Contains(checkpoint.Name, "normalization"), strings.Contains(checkpoint.Name, "offset"):
				return "FAST_STANDARD_FIRST_MATERIAL_EVALMOD_INPUT"
			case strings.Contains(checkpoint.Name, "double_angle"), strings.Contains(checkpoint.Name, "evalmod_output"):
				return "FAST_STANDARD_FIRST_MATERIAL_EVALMOD_DOUBLE_ANGLE"
			default:
				return "FAST_STANDARD_FIRST_MATERIAL_EVALMOD_POLYNOMIAL"
			}
		}
		previousMax = maxDiff
	}
	for _, power := range trace.powers {
		if power.MaxComplexDiffFastStandard >= numericalMaterialThreshold {
			return "FAST_STANDARD_FIRST_MATERIAL_EVALMOD_POLYNOMIAL"
		}
	}
	if !trace.fastReplayVerified || !trace.standardReplayVerified {
		return "FAST_STANDARD_EVALMOD_REPLAY_UNCLOSED"
	}
	return "FAST_STANDARD_EVALMOD_REPLAY_CLOSED"
}

func runFastNumericalStages(eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, pre []complex128, finalCT *rlwe.Ciphertext, final []complex128) ([]numericalStageSample, error) {
	if eval == nil || input == nil || finalCT == nil {
		return nil, fmt.Errorf("Fast evaluator, input, and final public ciphertext are required")
	}
	samples := []numericalStageSample{numericalStageSampleFromValues("input", input, pre, true, true, "")}
	packed, packN1, packN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, fmt.Errorf("PackAndSwitchN1ToN2: %w", err)
	}
	if len(packed) != 1 {
		return nil, fmt.Errorf("Fast packing returned %d ciphertexts, want 1", len(packed))
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return nil, fmt.Errorf("ScaleDown: %w", err)
	}
	sample := captureFastNumericalStage("scale_down", scaled)
	samples = append(samples, sample)
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return nil, fmt.Errorf("ModUp: %w", err)
	}
	sample = captureFastNumericalStage("mod_up", modUp)
	samples = append(samples, sample)
	ctReal, ctImag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return nil, fmt.Errorf("CoeffsToSlots: %w", err)
	}
	sample = captureFastNumericalStage("c2s_real", ctReal)
	samples = append(samples, sample)
	if ctImag == nil {
		samples = append(samples, numericalUnavailableStage("c2s_imag", "Fast C2S did not produce an imaginary branch"))
	} else {
		sample = captureFastNumericalStage("c2s_imag", ctImag)
		samples = append(samples, sample)
	}
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return nil, fmt.Errorf("EvalMod(real): %w", err)
	}
	sample = captureFastNumericalStage("evalmod_real", ctReal)
	samples = append(samples, sample)
	if ctImag == nil {
		samples = append(samples, numericalUnavailableStage("evalmod_imag", "Fast C2S imaginary branch was absent"))
	} else {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return nil, fmt.Errorf("EvalMod(imag): %w", err)
		}
		sample = captureFastNumericalStage("evalmod_imag", ctImag)
		samples = append(samples, sample)
	}
	ctOut, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return nil, fmt.Errorf("SlotsToCoeffs: %w", err)
	}
	sample = captureFastNumericalStage("s2c", ctOut)
	samples = append(samples, sample)
	_ = packN1
	_ = packN2
	_ = finalCT
	samples = append(samples, numericalStageSampleFromValues("final_public_output", finalCT, final, true, true, ""))
	return samples, nil
}

func runStandardNumericalStages(eval *bootstrapping.Evaluator, input *rlwe.Ciphertext, pre []complex128, finalCT *rlwe.Ciphertext, final []complex128) ([]numericalStageSample, error) {
	if eval == nil || input == nil || finalCT == nil {
		return nil, fmt.Errorf("Standard evaluator, input, and final public ciphertext are required")
	}
	samples := []numericalStageSample{numericalStageSampleFromValues("input", input, pre, false, true, "")}
	packed, packN1, packN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, fmt.Errorf("PackAndSwitchN1ToN2: %w", err)
	}
	if len(packed) != 1 {
		return nil, fmt.Errorf("Standard packing returned %d ciphertexts, want 1", len(packed))
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return nil, fmt.Errorf("ScaleDown: %w", err)
	}
	sample := captureStandardNumericalStage("scale_down", scaled)
	samples = append(samples, sample)
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return nil, fmt.Errorf("ModUp: %w", err)
	}
	sample = captureStandardNumericalStage("mod_up", modUp)
	samples = append(samples, sample)
	ctReal, ctImag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return nil, fmt.Errorf("CoeffsToSlots: %w", err)
	}
	sample = captureStandardNumericalStage("c2s_real", ctReal)
	samples = append(samples, sample)
	if ctImag == nil {
		samples = append(samples, numericalUnavailableStage("c2s_imag", "Standard C2S did not produce an imaginary branch"))
	} else {
		sample = captureStandardNumericalStage("c2s_imag", ctImag)
		samples = append(samples, sample)
	}
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return nil, fmt.Errorf("EvalMod(real): %w", err)
	}
	sample = captureStandardNumericalStage("evalmod_real", ctReal)
	samples = append(samples, sample)
	if ctImag == nil {
		samples = append(samples, numericalUnavailableStage("evalmod_imag", "Standard C2S imaginary branch was absent"))
	} else {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return nil, fmt.Errorf("EvalMod(imag): %w", err)
		}
		sample = captureStandardNumericalStage("evalmod_imag", ctImag)
		samples = append(samples, sample)
	}
	ctOut, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return nil, fmt.Errorf("SlotsToCoeffs: %w", err)
	}
	sample = captureStandardNumericalStage("s2c", ctOut)
	samples = append(samples, sample)
	_ = packN1
	_ = packN2
	samples = append(samples, numericalStageSampleFromValues("final_public_output", finalCT, final, false, true, ""))
	return samples, nil
}

func captureFastNumericalStage(name string, ct *rlwe.Ciphertext) numericalStageSample {
	if ct == nil {
		return numericalUnavailableStage(name, "stage returned nil ciphertext")
	}
	return numericalStageSampleFromCiphertext(name, ct, true)
}

func captureStandardNumericalStage(name string, ct *rlwe.Ciphertext) numericalStageSample {
	if ct == nil {
		return numericalUnavailableStage(name, "stage returned nil ciphertext")
	}
	return numericalStageSampleFromCiphertext(name, ct, false)
}

func decodeNumericalStagePair(pair *numericalStagePair, params ckks.Parameters, standardSK *rlwe.SecretKey) {
	if pair.fast.ciphertext == nil || pair.standard.ciphertext == nil {
		pair.fast.comparable, pair.standard.comparable = false, false
		pair.fast.reason, pair.standard.reason = "stage ciphertext missing", "stage ciphertext missing"
		return
	}
	fastRows, err := fastckks.QPrefixWidth(pair.fast.ciphertext.Level())
	if err != nil {
		pair.fast.comparable, pair.fast.reason = false, err.Error()
		return
	}
	standardRows := pair.standard.ciphertext.Level() + 1
	commonRows := min(fastRows, standardRows)
	if commonRows < 1 {
		pair.fast.comparable, pair.standard.comparable = false, false
		pair.fast.reason, pair.standard.reason = "no common authoritative Q rows", "no common authoritative Q rows"
		return
	}
	commonLevel := commonRows - 1
	fastProjected, err := projectNumericalStageCiphertext(params, pair.fast.ciphertext, commonLevel)
	if err != nil {
		pair.fast.comparable, pair.fast.reason = false, "Fast common-Q projection: "+err.Error()
		return
	}
	standardProjected, err := projectNumericalStageCiphertext(params, pair.standard.ciphertext, commonLevel)
	if err != nil {
		pair.standard.comparable, pair.standard.reason = false, "Standard common-Q projection: "+err.Error()
		return
	}
	pair.fast.values, err = fastStandardDecodeFast(params, fastProjected)
	if err != nil {
		pair.fast.comparable, pair.fast.reason = false, "Fast common-Q decode: "+err.Error()
		return
	}
	plain := rlwe.NewDecryptor(params, standardSK).DecryptNew(standardProjected)
	pair.standard.values = make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(plain, pair.standard.values); err != nil {
		pair.standard.comparable, pair.standard.reason = false, "Standard common-Q decode: "+err.Error()
		return
	}
	if err := validateNumericalVector(pair.standard.values); err != nil {
		pair.standard.comparable, pair.standard.reason = false, "Standard decoded vector: "+err.Error()
		return
	}
	pair.fast.comparable, pair.standard.comparable = true, true
}

func projectNumericalStageCiphertext(params ckks.Parameters, source *rlwe.Ciphertext, level int) (*rlwe.Ciphertext, error) {
	if level < 0 || level > source.Level() || level > params.MaxLevel() {
		return nil, fmt.Errorf("projection level %d is outside source/parameter levels", level)
	}
	projected := ckks.NewCiphertext(params, source.Degree(), level)
	*projected.MetaData = *source.MetaData
	projected.IsNTT, projected.IsMontgomery = source.IsNTT, source.IsMontgomery
	ringQ := params.RingQ()
	for component := 0; component <= source.Degree(); component++ {
		if component >= len(source.Value) || component >= len(projected.Value) {
			return nil, fmt.Errorf("component %d is unavailable", component)
		}
		for limb := 0; limb <= level; limb++ {
			if limb >= len(source.Value[component].Coeffs) || limb >= len(projected.Value[component].Coeffs) {
				return nil, fmt.Errorf("component %d row q%d is unavailable", component, limb)
			}
			copy(projected.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
			if source.IsMontgomery {
				ringQ.SubRings[limb].IMForm(projected.Value[component].Coeffs[limb], projected.Value[component].Coeffs[limb])
			}
		}
	}
	projected.IsMontgomery = false
	return projected, nil
}

func numericalStageSampleFromValues(name string, ct *rlwe.Ciphertext, values []complex128, fast, comparable bool, reason string) numericalStageSample {
	return numericalStageSample{name: name, state: numericalStageStateFromCiphertext(ct, fast), ciphertext: ct, values: values, comparable: comparable, reason: reason}
}

func numericalStageSampleFromCiphertext(name string, ct *rlwe.Ciphertext, fast bool) numericalStageSample {
	return numericalStageSample{name: name, state: numericalStageStateFromCiphertext(ct, fast), ciphertext: ct, comparable: ct != nil, reason: ""}
}

func numericalStageStateFromCiphertext(ct *rlwe.Ciphertext, fast bool) NumericalStageState {
	state := NumericalStageState{}
	if ct != nil {
		state.Metadata = numericalMetadata(ct)
		state.AuthorityRows = ct.Level() + 1
		if fast {
			if rows, err := fastckks.QPrefixWidth(ct.Level()); err == nil {
				state.AuthorityRows, state.MaintainedRows = rows, rows
			}
		}
	}
	return state
}

func numericalUnavailableStage(name, reason string) numericalStageSample {
	return numericalStageSample{name: name, comparable: false, reason: reason}
}

func summarizeNumericalStageLockstep(pairs []numericalStagePair, original []complex128) NumericalStageLockstep {
	result := NumericalStageLockstep{
		Checkpoints:       make([]NumericalStageCheckpoint, 0, len(pairs)),
		MaterialThreshold: numericalMaterialThreshold, ObservableThreshold: numericalObservableThreshold,
		Classification:                       "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED",
		CombinedBranchS2CAmplificationStatus: "NOT_COMPARABLE",
	}
	var previousMax *float64
	evalModFast, evalModStandard := make([]complex128, 0), make([]complex128, 0)
	evalModBranches := map[string]bool{}
	for _, pair := range pairs {
		checkpoint := NumericalStageCheckpoint{
			Name: pair.fast.name, Fast: pair.fast.state, Standard: pair.standard.state,
			StageReferenceSNR:   numericalmetrics.Unavailable("one or both stage values are not semantically comparable"),
			AmplificationStatus: "NOT_COMPARABLE", DeltaSNRStatus: "NOT_COMPARABLE",
		}
		if !pair.fast.comparable || !pair.standard.comparable || len(pair.fast.values) == 0 || len(pair.fast.values) != len(pair.standard.values) {
			checkpoint.NotComparableReason = pair.fast.reason
			if pair.standard.reason != "" {
				checkpoint.NotComparableReason += "; " + pair.standard.reason
			}
			result.Checkpoints = append(result.Checkpoints, checkpoint)
			previousMax = nil
			continue
		}
		metrics, _ := numericalComparison(pair.standard.values, pair.fast.values)
		checkpoint.StageReferenceSNR = numericalmetrics.Compare(pair.standard.values, pair.fast.values)
		if field := firstNonFiniteNumber(metrics); field != "" {
			checkpoint.NotComparableReason = "complex comparison overflowed at " + field
			checkpoint.StageReferenceSNR = numericalmetrics.Unavailable(checkpoint.NotComparableReason)
			result.Checkpoints = append(result.Checkpoints, checkpoint)
			previousMax = nil
			continue
		}
		checkpoint.Comparable = true
		checkpoint.FastVsStandard = &metrics
		distance := metrics.Complex.RMSE
		checkpoint.D = &distance
		maxDiff := metrics.Complex.Max
		if pair.fast.name == "input" || pair.fast.name == "final_public_output" {
			precision := numericalPrecisionBits(original, pair.fast.values)
			checkpoint.PrecisionVsCanonicalOriginal = &precision
		}
		if result.FirstObservable == "" && maxDiff >= numericalObservableThreshold {
			result.FirstObservable = checkpoint.Name
			result.FirstObservableMaxDiff = &maxDiff
			checkpoint.FirstObservable = true
		}
		if result.FirstMaterial == "" && previousMax != nil && *previousMax < numericalMaterialThreshold && maxDiff >= numericalMaterialThreshold {
			result.FirstMaterial = checkpoint.Name
			result.FirstMaterialMaxDiff = &maxDiff
			checkpoint.FirstMaterial = true
			result.Classification = numericalClassificationForCheckpoint(checkpoint.Name)
		}
		if checkpoint.Name == "evalmod_real" || checkpoint.Name == "evalmod_imag" {
			evalModFast = append(evalModFast, pair.fast.values...)
			evalModStandard = append(evalModStandard, pair.standard.values...)
			evalModBranches[strings.TrimPrefix(checkpoint.Name, "evalmod_")] = true
		}
		copyMax := maxDiff
		previousMax = &copyMax
		result.Checkpoints = append(result.Checkpoints, checkpoint)
	}
	if len(result.Checkpoints) > 0 {
		for i := len(result.Checkpoints) - 1; i >= 0; i-- {
			if result.Checkpoints[i].Name == "final_public_output" {
				if result.Checkpoints[i].D != nil {
					result.FinalFastStandardRMSE = *result.Checkpoints[i].D
				}
				break
			}
		}
	}
	var combinedEvalModSNR *numericalmetrics.SNR
	if evalModBranches["real"] && evalModBranches["imag"] && len(evalModFast) > 0 && len(evalModFast) == len(evalModStandard) {
		combinedEvalModSNR = setCombinedBranchS2CMetrics(&result, evalModStandard, evalModFast)
	}
	applyNumericalStageTopology(&result, combinedEvalModSNR)
	return result
}

func setCombinedBranchS2CMetrics(result *NumericalStageLockstep, evalModStandard, evalModFast []complex128) *numericalmetrics.SNR {
	if result == nil || len(evalModStandard) == 0 || len(evalModStandard) != len(evalModFast) {
		return nil
	}
	combined := numericalmetrics.Compare(evalModStandard, evalModFast)
	if combined.Status == numericalmetrics.NotComparable {
		return nil
	}
	combinedMetrics, _ := numericalComparison(evalModStandard, evalModFast)
	if firstNonFiniteNumber(combinedMetrics) != "" {
		return nil
	}
	jointD := combinedMetrics.Complex.RMSE
	result.CombinedBranchEvalModErrorRMSE = &jointD
	if s2c := numericalStageCheckpointByName(result, "s2c"); s2c != nil && s2c.D != nil {
		switch {
		case jointD > 0:
			factor := *s2c.D / jointD
			if math.IsNaN(factor) || math.IsInf(factor, 0) {
				result.CombinedBranchS2CAmplificationStatus = "NON_FINITE"
			} else {
				result.CombinedBranchS2CAmplification = &factor
				result.CombinedBranchS2CAmplificationStatus = "FINITE"
			}
		case *s2c.D > 0:
			result.CombinedBranchS2CAmplificationStatus = "POSITIVE_INFINITY"
		default:
			result.CombinedBranchS2CAmplificationStatus = "UNDEFINED_ZERO_OVER_ZERO"
		}
	}
	return &combined
}

func applyNumericalStageTopology(result *NumericalStageLockstep, combinedEvalModSNR *numericalmetrics.SNR) {
	if result == nil {
		return
	}
	result.LargestRawAmplificationCheckpoint = ""
	result.LargestRawAmplificationFactor = nil
	result.LargestSNRDropCheckpoint = ""
	result.LargestSNRDropDB = nil
	for i := range result.Checkpoints {
		checkpoint := &result.Checkpoints[i]
		checkpoint.Amplification = nil
		checkpoint.AmplificationStatus = "NOT_COMPARABLE"
		checkpoint.AmplificationParent = ""
		checkpoint.DeltaSNRDB = nil
		checkpoint.DeltaSNRStatus = "NOT_COMPARABLE"
		checkpoint.DeltaSNRParent = ""
	}

	parents := []struct{ child, parent string }{
		{"scale_down", "input"},
		{"mod_up", "scale_down"},
		{"c2s_real", "mod_up"},
		{"c2s_imag", "mod_up"},
		{"evalmod_real", "c2s_real"},
		{"evalmod_imag", "c2s_imag"},
		{"final_public_output", "s2c"},
	}
	for _, relation := range parents {
		child := numericalStageCheckpointByName(result, relation.child)
		parent := numericalStageCheckpointByName(result, relation.parent)
		if child == nil {
			continue
		}
		child.AmplificationParent = relation.parent
		child.DeltaSNRParent = relation.parent
		if parent == nil || !child.Comparable || !parent.Comparable || child.D == nil || parent.D == nil {
			continue
		}
		child.Amplification, child.AmplificationStatus = numericalErrorAmplification(*child.D, *parent.D)
		child.DeltaSNRDB, child.DeltaSNRStatus = numericalSNRDelta(child.StageReferenceSNR, parent.StageReferenceSNR)
	}

	if s2c := numericalStageCheckpointByName(result, "s2c"); s2c != nil {
		s2c.AmplificationStatus = "NOT_APPLICABLE_COMBINED_BRANCH"
		s2c.DeltaSNRParent = "evalmod_real+evalmod_imag (joint)"
		if combinedEvalModSNR != nil && s2c.Comparable {
			s2c.DeltaSNRDB, s2c.DeltaSNRStatus = numericalSNRDelta(s2c.StageReferenceSNR, *combinedEvalModSNR)
		}
	}

	var largestRaw, largestDrop *float64
	for i := range result.Checkpoints {
		checkpoint := &result.Checkpoints[i]
		if checkpoint.Amplification != nil && (largestRaw == nil || *checkpoint.Amplification > *largestRaw) {
			factor := *checkpoint.Amplification
			largestRaw = &factor
			result.LargestRawAmplificationCheckpoint = checkpoint.Name
		}
		if checkpoint.DeltaSNRDB != nil && *checkpoint.DeltaSNRDB < 0 && (largestDrop == nil || *checkpoint.DeltaSNRDB < *largestDrop) {
			drop := *checkpoint.DeltaSNRDB
			largestDrop = &drop
			result.LargestSNRDropCheckpoint = checkpoint.Name
		}
	}
	result.LargestRawAmplificationFactor = largestRaw
	result.LargestSNRDropDB = largestDrop
}

func numericalStageCheckpointByName(result *NumericalStageLockstep, name string) *NumericalStageCheckpoint {
	if result == nil {
		return nil
	}
	for i := range result.Checkpoints {
		if result.Checkpoints[i].Name == name {
			return &result.Checkpoints[i]
		}
	}
	return nil
}

func numericalErrorAmplification(currentD, parentD float64) (*float64, string) {
	switch {
	case parentD > 0:
		factor := currentD / parentD
		if math.IsNaN(factor) || math.IsInf(factor, 0) {
			return nil, "POSITIVE_INFINITY"
		}
		return &factor, "FINITE"
	case currentD > 0:
		return nil, "UNDEFINED_ZERO_PREVIOUS_D"
	default:
		return nil, "UNDEFINED_ZERO_OVER_ZERO"
	}
}

func numericalSNRDelta(current, previous numericalmetrics.SNR) (*float64, string) {
	if current.Status == numericalmetrics.Finite && previous.Status == numericalmetrics.Finite && current.SNRDB != nil && previous.SNRDB != nil {
		delta := *current.SNRDB - *previous.SNRDB
		return &delta, "FINITE"
	}
	if current.Status == numericalmetrics.Finite && previous.Status == numericalmetrics.PositiveInfinity {
		return nil, "NEGATIVE_INFINITY"
	}
	if current.Status == numericalmetrics.PositiveInfinity && previous.Status == numericalmetrics.Finite {
		return nil, "POSITIVE_INFINITY"
	}
	if current.Status == numericalmetrics.PositiveInfinity && previous.Status == numericalmetrics.PositiveInfinity {
		return nil, "UNDEFINED_INFINITY_MINUS_INFINITY"
	}
	return nil, "NOT_COMPARABLE"
}

func numericalClassificationForCheckpoint(name string) string {
	switch name {
	case "c2s_real", "c2s_imag":
		return "CURRENT_FAST_FIRST_MATERIAL_C2S"
	case "evalmod_real", "evalmod_imag":
		return "CURRENT_FAST_FIRST_MATERIAL_EVALMOD_POLYNOMIAL"
	case "s2c":
		return "CURRENT_FAST_FIRST_MATERIAL_S2C"
	default:
		return "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED"
	}
}

func numericalScaleAudit(fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, pairs []numericalStagePair) []NumericalScaleAudit {
	audit := make([]NumericalScaleAudit, 0, len(pairs)+3)
	for _, pair := range pairs {
		if !pair.fast.comparable {
			continue
		}
		for _, mode := range []struct {
			name  string
			state NumericalStageState
		}{{"fast", pair.fast.state}, {"standard", pair.standard.state}} {
			audit = append(audit, NumericalScaleAudit{
				Checkpoint: pair.fast.name + "/" + mode.name,
				Scale:      mode.state.Metadata.Scale, ScaleExact: mode.state.Metadata.ScaleExact,
				ScaleFloat64Finite: mode.state.Metadata.ScaleFloat64Finite,
				ScaleLog2:          mode.state.Metadata.ScaleLog2, Level: mode.state.Metadata.Level,
			})
		}
	}
	if fastEval == nil || standardEval == nil || fastEval.Mod1Evaluator == nil || standardEval.Mod1Evaluator == nil {
		return audit
	}
	mod1Params := fastEval.Mod1Evaluator.Parameters
	inputLevel := mod1Params.LevelQ
	target := mod1Params.ScalingFactor()
	for i := 0; i < mod1Params.DoubleAngle; i++ {
		index := inputLevel - mod1Params.Mod1Poly.Depth() - mod1Params.DoubleAngle + i + 1
		if index < 0 || index >= len(params.Q()) {
			break
		}
		target = target.Mul(rlwe.NewScale(params.Q()[index]))
		target.Value.Sqrt(&target.Value)
	}
	targetFloat := target.Float64()
	targetFloatFinite := !math.IsNaN(targetFloat) && !math.IsInf(targetFloat, 0)
	if !targetFloatFinite {
		targetFloat = 0
	}
	audit = append(audit, NumericalScaleAudit{
		Checkpoint: "evalmod/standard_target_scale",
		Scale:      targetFloat, ScaleExact: target.Value.Text('e', 80), ScaleFloat64Finite: targetFloatFinite, ScaleLog2: target.Log2(),
		TargetScale: targetFloat, TargetScaleExact: target.Value.Text('e', 80), TargetScaleFloat64Finite: targetFloatFinite, TargetScaleLog2: target.Log2(),
		Level: inputLevel,
	})
	return audit
}

func numericalScaleLogRatio(numerator, denominator rlwe.Scale) int {
	if numerator.Value.Sign() <= 0 || denominator.Value.Sign() <= 0 {
		return 0
	}
	ratio := numerator.Div(denominator)
	return int(math.Round(ratio.Log2()))
}
