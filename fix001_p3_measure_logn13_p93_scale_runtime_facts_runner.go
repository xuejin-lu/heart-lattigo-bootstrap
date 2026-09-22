package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const requiredFIX001P3MeasureSecondary = "40532b4dce5c7eeae2db5b0b6f21be64801ce923"

type fix001P3RuntimeResidual struct {
	MaxComponentAbs float64 `json:"max_component_abs"`
	MaxAbsComplex   float64 `json:"max_abs_complex"`
}

type fix001P3RuntimeMeasurementResult struct {
	SchemaVersion  string                   `json:"schema_version"`
	Timestamp      time.Time                `json:"timestamp"`
	Provenance     map[string]interface{}   `json:"provenance"`
	M0             map[string]interface{}   `json:"M0_baseline"`
	M1             map[string]interface{}   `json:"M1_scale_down"`
	M2             map[string]interface{}   `json:"M2_mod_up"`
	M3             map[string]interface{}   `json:"M3_ps_alignment"`
	M4             map[string]interface{}   `json:"M4_metadata_snaps"`
	M5             map[string]interface{}   `json:"M5_normalized_mod1"`
	M6             []map[string]interface{} `json:"M6_c2s_restore"`
	Secondary      map[string]interface{}   `json:"secondary_clean_state"`
	Classification string                   `json:"classification"`
}

func fix001P3RuntimeResidualFor(params ckks.Parameters, before, after *rlwe.Ciphertext) (fix001P3RuntimeResidual, error) {
	a, err := psGlobalDecode(params, before)
	if err != nil {
		return fix001P3RuntimeResidual{}, err
	}
	b, err := psGlobalDecode(params, after)
	if err != nil {
		return fix001P3RuntimeResidual{}, err
	}
	metric := psGlobalMetric(a, b)
	return fix001P3RuntimeResidual{MaxComponentAbs: metric.MaxComponent, MaxAbsComplex: metric.MaxComplex}, nil
}

func fix001P3RuntimeScaleDownBoundary(params ckks.Parameters, source *rlwe.Ciphertext, messageRatio float64) *rlwe.Ciphertext {
	boundary := source.CopyNew()
	ringQ := params.RingQ()
	for boundary.Level() != 0 {
		current := rlwe.NewScale(ringQ.ModulusAtLevel[boundary.Level()]).Div(boundary.Scale)
		target := rlwe.NewScale(ringQ.SubRings[boundary.Level()].Modulus).Mul(rlwe.NewScale(messageRatio))
		if current.Cmp(target) < 0 {
			break
		}
		fastckks.Resize(boundary, boundary.Degree(), boundary.Level()-1, params.N())
	}
	return boundary
}

func fix001P3RuntimeRatioEvidence(ratio rlwe.Scale) map[string]interface{} {
	rounded, absMismatch, relative, exact := fix001P3ScaleRounding(ratio)
	return map[string]interface{}{
		"exact_ratio":                ratio.Value.Text('e', 80),
		"ratio_bigint":               rounded,
		"absolute_rounding_mismatch": absMismatch,
		"relative_rounding_mismatch": relative,
		"exact_integral":             exact,
	}
}

func fix001P3RuntimeScaleDownRow(params ckks.Parameters, eval *bootstrapping.FastEvaluator, source *rlwe.Ciphertext, sk *rlwe.SecretKey, label string) (map[string]interface{}, *rlwe.Ciphertext, error) {
	boundary := fix001P3RuntimeScaleDownBoundary(params, source, eval.Mod1Parameters.MessageRatio())
	current := rlwe.NewScale(params.RingQ().ModulusAtLevel[boundary.Level()]).Div(boundary.Scale)
	target := rlwe.NewScale(eval.Mod1Parameters.MessageRatio())
	ratio := current.Div(target)
	rounded := ratio.BigInt()
	residualBeforeAfter := boundary.CopyNew()
	beforeValues, err := psGlobalDecode(params, residualBeforeAfter)
	if err != nil {
		return nil, nil, err
	}
	if err = eval.FastCKKS.MulIntegerMaintained(residualBeforeAfter, rounded, residualBeforeAfter); err != nil {
		return nil, nil, err
	}
	residualBeforeAfter.Scale = residualBeforeAfter.Scale.Mul(rlwe.NewScale(rounded))
	afterValues, err := psGlobalDecode(params, residualBeforeAfter)
	if err != nil {
		return nil, nil, err
	}
	residualMetric := psGlobalMetric(beforeValues, afterValues)
	actual := source.CopyNew()
	if _, _, err = eval.ScaleDown(actual); err != nil {
		return nil, nil, err
	}
	row := map[string]interface{}{
		"label": label, "boundary_level": boundary.Level(), "current_message_ratio": current.Value.Text('e', 80),
		"target_message_ratio": target.Value.Text('e', 80), "scale_up": ratio.Value.Text('e', 80),
		"scale_up_bigint": rounded.String(), "actual_output_level": actual.Level(), "actual_output_scale": finalizationScaleString(actual.Scale),
		"semantic_before_after_max_component": residualMetric.MaxComponent, "semantic_before_after_max_complex": residualMetric.MaxComplex,
	}
	for key, value := range fix001P3RuntimeRatioEvidence(ratio) {
		row[key] = value
	}
	return row, actual, nil
}

func fix001P3RuntimeModUpEvidence(params ckks.Parameters, mod1Scale rlwe.Scale, messageRatio float64, source *rlwe.Ciphertext, sk *rlwe.SecretKey, label string, fastEval *fastckks.Evaluator, standardEval *bootstrapping.Evaluator, fast bool) (map[string]interface{}, error) {
	ratio := mod1Scale.Div(rlwe.NewScale(messageRatio)).Div(source.Scale)
	floatRatio := mod1Scale.Float64() / messageRatio / source.Scale.Float64()
	scalar := math.Round(floatRatio)
	if scalar < 1 {
		scalar = 1
	}
	scalarBig := new(big.Int).SetInt64(int64(scalar))
	before := source.CopyNew()
	after := source.CopyNew()
	beforeValues, err := psGlobalDecode(params, before)
	if err != nil {
		return nil, err
	}
	if floatRatio > 1 {
		if fast {
			err = fastEval.MulIntegerMaintained(after, scalarBig, after)
		} else {
			err = standardEval.Evaluator.Mul(after, scalarBig, after)
		}
		if err != nil {
			return nil, err
		}
		after.Scale = after.Scale.Mul(rlwe.NewScale(floatRatio))
	}
	afterValues, err := psGlobalDecode(params, after)
	if err != nil {
		return nil, err
	}
	residual := psGlobalMetric(beforeValues, afterValues)
	row := map[string]interface{}{
		"label": label, "scaling_factor": mod1Scale.Value.Text('e', 80), "message_ratio": messageRatio,
		"input_scale": finalizationScaleString(source.Scale), "ratio_after_float64": floatRatio, "scalar_round": scalar,
		"absolute_scalar_minus_exact": math.Abs(scalar - ratio.Float64()), "relative_scalar_minus_exact": math.Abs(scalar/ratio.Float64() - 1),
		"semantic_before_after_max_component": residual.MaxComponent, "semantic_before_after_max_complex": residual.MaxComplex,
		"physical_scalar_applied": floatRatio > 1,
	}
	row["exact_ratio"] = ratio.Value.Text('e', 80)
	return row, nil
}

func fix001P3RuntimeAlignmentRows(branch string, trace fix001P3ScaleMeasurementTrace) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(trace.Alignments))
	for _, observation := range trace.Alignments {
		rows = append(rows, map[string]interface{}{
			"operation_id": observation.OperationID, "branch": branch, "checkpoint": observation.Checkpoint,
			"numerator_scale": observation.NumeratorScale, "denominator_scale": observation.DenominatorScale,
			"exact_ratio": observation.ExactRatio, "ratio_bigint": observation.RatioBigInt,
			"relative_rounding_mismatch": observation.RelativeMismatch, "semantic_residual": observation.SemanticResidual,
			"exact_integral": observation.ExactIntegral,
		})
	}
	return rows
}

func fix001P3RuntimeSnapRows(branch string, trace fix001P3ScaleMeasurementTrace) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(trace.Snaps))
	for _, observation := range trace.Snaps {
		rows = append(rows, map[string]interface{}{
			"operation_id": observation.OperationID, "branch": branch, "checkpoint": observation.Checkpoint,
			"scale_a": observation.ScaleA, "scale_b": observation.ScaleB, "ratio_a_over_b": observation.RatioAOverB,
			"log2_delta": observation.Log2Delta, "relative_difference": observation.RelativeDifference,
			"interpretation_shift_magnitude": observation.InterpretationShift, "semantic_residual": observation.SemanticResidual,
		})
	}
	return rows
}

func fix001P3RuntimeMod1Ledger(path evalModMatchedPath, eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, planScale rlwe.Scale) map[string]interface{} {
	scalingFactor := eval.Mod1Parameters.ScalingFactor()
	initial := map[string]interface{}{}
	rounds := make([]map[string]interface{}, 0)
	final := map[string]interface{}{}
	reset := map[string]interface{}{}
	for _, row := range path.ScaleLedger {
		switch row["kind"] {
		case "initial_normalization":
			initial = row
		case "double_angle":
			rounds = append(rounds, row)
		case "final_pre_materialization":
			final = row
		case "post_reset":
			reset = row
		}
	}
	return map[string]interface{}{
		"evalmod_input_scale": finalizationScaleString(input.Scale), "scaling_factor": scalingFactor.Value.Text('e', 80),
		"target_scale": path.PolyMeta["target_scale"], "plan_scale": finalizationScaleString(planScale), "polynomial_output_scale": path.PolyMeta["fast_scale"],
		"k_in": initial["k_in"], "coherent_scale": initial["coherent_scale"], "coherent_target_ratio": initial["coherent_target_ratio"], "coherent_target_log2_delta": initial["coherent_target_log2_delta"],
		"rounds": rounds, "final": final, "after_caller_input_reset_and_public_default_scale": reset,
	}
}

func fix001P3MeasureRuntimeWrite(result fix001P3RuntimeMeasurementResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"provenance", result.Provenance},
		{"M0_baseline", result.M0}, {"M1_scale_down", result.M1}, {"M2_mod_up", result.M2}, {"M3_ps_alignment", result.M3},
		{"M4_metadata_snaps", result.M4}, {"M5_normalized_mod1", result.M5}, {"M6_c2s_restore", result.M6},
		{"secondary_clean_state", result.Secondary}, {"classification", result.Classification},
	}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, field := range fields {
		value, err := json.Marshal(field.value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&buf, "  %q: %s", field.name, value)
		if i+1 != len(fields) {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func runFIX001P3MeasureLogN13P93ScaleRuntimeFacts(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	secondaryStatus := gitOutput(backendRoot, "status", "--porcelain")
	if cfg.LogN != 13 || secondaryCommit != requiredFIX001P3MeasureSecondary || secondaryBranch != "fast-ckks" || secondaryStatus != "" {
		return fmt.Errorf("measurement precondition failed: logN=%d secondary=%s branch=%s status=%q", cfg.LogN, secondaryCommit, secondaryBranch, secondaryStatus)
	}
	residual, btp, err := NewBootstrapParametersFromConfig(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, standardSK, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return err
	}
	profile := q056PreparedProfile{Residual: residual, BTP: btp, Fast: fastEval, Standard: standardEval, StandardSK: standardSK}
	profile.Inputs, err = evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, standardSK, true)
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	fastSK := zeroSecret(params)
	workload := reproducibleValues(profile.Residual.MaxSlots())
	fastOutput, err := profile.Fast.Bootstrap(reproducibleInput(profile.Residual, profile.BTP).CopyNew())
	if err != nil {
		return err
	}
	standardOutput, err := profile.Standard.Bootstrap(reproducibleInput(profile.Residual, profile.BTP).CopyNew())
	if err != nil {
		return err
	}
	fastValues, err := decodeWithSecret(profile.Residual, fastOutput, fastSK)
	if err != nil {
		return err
	}
	standardValues, err := decodeWithSecret(profile.Residual, standardOutput, profile.StandardSK)
	if err != nil {
		return err
	}
	fastMetric, err := semanticBisectMetricFor(workload, fastValues, 1e-2)
	if err != nil {
		return err
	}
	standardMetric, err := semanticBisectMetricFor(workload, standardValues, 1e-2)
	if err != nil {
		return err
	}
	result := fix001P3RuntimeMeasurementResult{
		SchemaVersion: "fix-001-p3-measure-logn13-p93-scale-runtime-facts.v1", Timestamp: time.Now().UTC(),
		Provenance:     map[string]interface{}{"primary": gitMetadata(primaryRoot), "secondary": gitMetadata(backendRoot), "config": cfg, "log_n": 13, "q0_bits": 56, "ps_authority": "Q012-wide", "plan_scale": "2^93", "measurement_only": true},
		M0:             map[string]interface{}{"fast_exact_e2e_max_component": fastMetric.MaxComponentAbs, "standard_exact_e2e_max_component": standardMetric.MaxComponentAbs, "fast_final_scale": finalizationScaleString(fastOutput.Scale), "standard_final_scale": finalizationScaleString(standardOutput.Scale), "fast_requirement_value": 1e-2},
		Secondary:      map[string]interface{}{"branch": secondaryBranch, "head": secondaryCommit, "clean": true, "exact_required_head": secondaryCommit == requiredFIX001P3MeasureSecondary, "production_modified": false},
		Classification: "MEASUREMENT_COMPLETE",
	}
	fastInput := reproducibleInput(profile.Residual, profile.BTP)
	standardInput := reproducibleInput(profile.Residual, profile.BTP)
	fastPacked, _, _, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput.CopyNew()})
	if err != nil {
		return err
	}
	standardPacked, _, _, err := profile.Standard.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput.CopyNew()})
	if err != nil {
		return err
	}
	fastScaleDown, fastScaled, err := fix001P3RuntimeScaleDownRow(params, profile.Fast, &fastPacked[0], fastSK, "Fast")
	if err != nil {
		return err
	}
	standardBoundary := fix001P3RuntimeScaleDownBoundary(params, &standardPacked[0], profile.Standard.Mod1Evaluator.Parameters.MessageRatio())
	standardCurrent := rlwe.NewScale(params.RingQ().ModulusAtLevel[standardBoundary.Level()]).Div(standardBoundary.Scale)
	standardTarget := rlwe.NewScale(profile.Standard.Mod1Evaluator.Parameters.MessageRatio())
	standardRatio := standardCurrent.Div(standardTarget)
	standardScaled, _, err := profile.Standard.ScaleDown(&standardPacked[0])
	if err != nil {
		return err
	}
	standardResidual, err := fix001P3RuntimeResidualFor(params, standardBoundary, standardBoundary.CopyNew())
	if err != nil {
		return err
	}
	standardRow := map[string]interface{}{"label": "Standard", "boundary_level": standardBoundary.Level(), "current_message_ratio": standardCurrent.Value.Text('e', 80), "target_message_ratio": standardTarget.Value.Text('e', 80), "scale_up": standardRatio.Value.Text('e', 80), "scale_up_bigint": standardRatio.BigInt().String(), "actual_output_level": standardScaled.Level(), "actual_output_scale": finalizationScaleString(standardScaled.Scale), "semantic_before_after_max_component": standardResidual.MaxComponentAbs}
	for key, value := range fix001P3RuntimeRatioEvidence(standardRatio) {
		standardRow[key] = value
	}
	result.M1 = map[string]interface{}{"fast": fastScaleDown, "standard": standardRow}
	fastModUp, err := fix001P3RuntimeModUpEvidence(params, profile.Fast.Mod1Parameters.ScalingFactor(), profile.Fast.Mod1Parameters.MessageRatio(), fastScaled, fastSK, "Fast", profile.Fast.FastCKKS, nil, true)
	if err != nil {
		return err
	}
	standardModUp, err := fix001P3RuntimeModUpEvidence(params, profile.Standard.Mod1Evaluator.Parameters.ScalingFactor(), profile.Standard.Mod1Evaluator.Parameters.MessageRatio(), standardScaled, profile.StandardSK, "Standard", nil, profile.Standard, false)
	if err != nil {
		return err
	}
	result.M2 = map[string]interface{}{"fast": fastModUp, "standard": standardModUp}
	planScale := fix001P3P93InternalPlanScale()
	alignmentRows := []map[string]interface{}{}
	snapRows := []map[string]interface{}{}
	for _, branch := range []string{"real", "imag"} {
		input := profile.Inputs.FastReal
		if branch == "imag" {
			input = profile.Inputs.FastImag
		}
		trace := fix001P3ScaleMeasurementTrace{}
		if _, err := fix001P3P93ProductionPSTrace(params, profile.Fast, input, planScale, &trace); err != nil {
			return err
		}
		alignmentRows = append(alignmentRows, fix001P3RuntimeAlignmentRows(branch, trace)...)
		snapRows = append(snapRows, fix001P3RuntimeSnapRows(branch, trace)...)
	}
	maxAlignment, maxSnap := 0.0, 0.0
	worstAlignment, worstSnap := "none", "none"
	for _, row := range alignmentRows {
		if value, ok := row["relative_rounding_mismatch"].(float64); ok && value > maxAlignment {
			maxAlignment, worstAlignment = value, row["operation_id"].(string)
		}
	}
	for _, row := range snapRows {
		if value, ok := row["semantic_residual"].(float64); ok && value > maxSnap {
			maxSnap, worstSnap = value, row["operation_id"].(string)
		}
	}
	result.M3 = map[string]interface{}{"count": len(alignmentRows), "max_relative_mismatch": maxAlignment, "worst_operation_id": worstAlignment, "max_semantic_residual": maxFloatFromRows(alignmentRows, "semantic_residual"), "rows": alignmentRows}
	result.M4 = map[string]interface{}{"count": len(snapRows), "worst_semantic_residual": maxSnap, "worst_operation_id": worstSnap, "rows": snapRows}
	var restoreRows []map[string]interface{}
	inputs, err := evalModMatchedC2SInputs(cfg, profile.BTP, profile.Fast, profile.Standard, profile.StandardSK, &restoreRows, true)
	if err != nil {
		return err
	}
	realPath, err := evalModMatchedRunPathWithPlanScale(inputs.FastReal, inputs.OrdinaryReal, profile.Fast, profile.Standard, params, inputs.FastSK, inputs.StandardSK, planScale)
	if err != nil {
		return err
	}
	imagPath, err := evalModMatchedRunPathWithPlanScale(inputs.FastImag, inputs.OrdinaryImag, profile.Fast, profile.Standard, params, inputs.FastSK, inputs.StandardSK, planScale)
	if err != nil {
		return err
	}
	result.M5 = map[string]interface{}{"real": fix001P3RuntimeMod1Ledger(realPath, profile.Fast, inputs.FastReal, planScale), "imag": fix001P3RuntimeMod1Ledger(imagPath, profile.Fast, inputs.FastImag, planScale)}
	result.M6 = restoreRows
	return fix001P3MeasureRuntimeWrite(result, outPath)
}

func maxFloatFromRows(rows []map[string]interface{}, key string) float64 {
	max := 0.0
	for _, row := range rows {
		if value, ok := row[key].(float64); ok && value > max {
			max = value
		}
	}
	return max
}
