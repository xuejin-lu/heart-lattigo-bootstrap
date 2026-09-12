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

const targetScaleRestorationThreshold = correctnessThreshold

type TargetScaleRestorationScaleEvidence struct {
	S                            string  `json:"compressed_output_scale"`
	T                            string  `json:"public_target_scale"`
	R                            string  `json:"target_over_compressed_ratio"`
	M                            string  `json:"chosen_nearest_positive_integer"`
	RelativeIntegerApproximation string  `json:"relative_integer_approximation_error"`
	SM                           string  `json:"compressed_scale_times_integer"`
	Log2Delta                    float64 `json:"log2_delta_sm_to_target"`
	NearestIntegerPositive       bool    `json:"nearest_integer_positive"`
}

type TargetScaleComponentCapacity struct {
	Component         int     `json:"component"`
	MaxAbsCentered    string  `json:"max_abs_centered_coefficient"`
	MaxAbsOverQ01Half float64 `json:"max_abs_over_q01_half"`
	OutsideCount      int     `json:"outside_count"`
}

type TargetScaleCapacityEvidence struct {
	Q0                    string                         `json:"q0"`
	Q1                    string                         `json:"q1"`
	Q01                   string                         `json:"q01"`
	Q01Half               string                         `json:"q01_half"`
	Components            []TargetScaleComponentCapacity `json:"per_component"`
	MaxAbsPrePromotion    string                         `json:"max_abs_pre_promotion"`
	MaxAbsProspective     string                         `json:"max_abs_prospective_promoted"`
	MaxProspectiveRatio   float64                        `json:"max_prospective_capacity_ratio"`
	OutsideCount          int                            `json:"outside_count"`
	WorstComponent        int                            `json:"worst_component"`
	WorstIndex            int                            `json:"worst_index"`
	WorstCenteredValue    string                         `json:"worst_centered_value"`
	WorstProspectiveValue string                         `json:"worst_prospective_value"`
	Pass                  bool                           `json:"pass_centered_capacity"`
}

type TargetScaleModularCheck struct {
	Pass          bool   `json:"pass"`
	MismatchCount int    `json:"mismatch_count"`
	FirstMismatch string `json:"first_mismatch,omitempty"`
}

type TargetScaleRestorationCheckpoint struct {
	State                   string                       `json:"state"`
	Level                   int                          `json:"level"`
	Degree                  int                          `json:"degree"`
	Scale                   string                       `json:"scale"`
	Rows                    []FinalizationRowFingerprint `json:"q0_q1_rows"`
	CorrectedOracle         *PSGlobalMetric              `json:"corrected_oracle,omitempty"`
	ScaledOracle            *PSGlobalMetric              `json:"scaled_oracle,omitempty"`
	NormalizedOracle        *PSGlobalMetric              `json:"normalized_oracle,omitempty"`
	PerturbationFromP0      *PSGlobalMetric              `json:"perturbation_from_p0,omitempty"`
	PerturbationFromP3      *PSGlobalMetric              `json:"perturbation_from_p3,omitempty"`
	IndependentModularCheck *TargetScaleModularCheck     `json:"independent_modular_check,omitempty"`
}

type TargetScaleRestorationResult struct {
	SchemaVersion       string                              `json:"schema_version"`
	Timestamp           time.Time                           `json:"timestamp"`
	Primary             RepositoryMetadata                  `json:"primary_repository"`
	Lattigo             RepositoryMetadata                  `json:"lattigo_repository"`
	Environment         EnvironmentMetadata                 `json:"environment"`
	Config              BootstrapConfig                     `json:"config"`
	Parameters          ExperimentParameters                `json:"effective_parameters"`
	Workload            CorrectnessWorkload                 `json:"workload"`
	AuthoritativeOracle string                              `json:"authoritative_poly_oracle"`
	Threshold           float64                             `json:"threshold"`
	Scale               TargetScaleRestorationScaleEvidence `json:"scale_restoration"`
	PreFinal            TargetScaleRestorationCheckpoint    `json:"canonical_pre_final_rescale"`
	P0                  TargetScaleRestorationCheckpoint    `json:"p0_baseline_compressed_output"`
	P0Capacity          TargetScaleCapacityEvidence         `json:"p0_capacity"`
	P1Capacity          TargetScaleCapacityEvidence         `json:"p1_prospective_capacity"`
	P2                  TargetScaleRestorationCheckpoint    `json:"p2_integer_promotion"`
	P3                  TargetScaleRestorationCheckpoint    `json:"p3_matched_scale"`
	P4                  TargetScaleRestorationCheckpoint    `json:"p4_exact_target_metadata"`
	P5Invariants        map[string]bool                     `json:"p5_restored_state_invariants"`
	FirstSupportedCause string                              `json:"first_supported_cause"`
	Validation          map[string]interface{}              `json:"validation"`
}

type targetScaleCenteredData struct {
	Q0      *big.Int
	Q1      *big.Int
	Q01     *big.Int
	Q01Half *big.Int
	Values  [][]*big.Int
}

func targetScaleFloatText(value *big.Float) string {
	return value.Text('e', 80)
}

func targetScaleIntText(value *big.Int) string {
	return value.Text(10)
}

func targetScaleCenteredRows(params ckks.Parameters, ct *rlwe.Ciphertext) (targetScaleCenteredData, error) {
	if ct == nil || ct.Level() < 1 || ct.Degree() < 1 || len(ct.Value) < 2 {
		return targetScaleCenteredData{}, fmt.Errorf("target-scale capacity requires a degree-one q0/q1 ciphertext")
	}
	if len(params.RingQ().SubRings) < 2 {
		return targetScaleCenteredData{}, fmt.Errorf("target-scale capacity requires q0 and q1")
	}
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	q01 := new(big.Int).Mul(q0, q1)
	q01Half := new(big.Int).Quo(new(big.Int).Set(q01), big.NewInt(2))
	q0Inverse := new(big.Int).ModInverse(q0, q1)
	if q0Inverse == nil {
		return targetScaleCenteredData{}, fmt.Errorf("q0 has no inverse modulo q1")
	}
	values := make([][]*big.Int, 0, ct.Degree()+1)
	for component := 0; component <= ct.Degree(); component++ {
		if len(ct.Value[component].Coeffs) < 2 || len(ct.Value[component].Coeffs[0]) != len(ct.Value[component].Coeffs[1]) {
			return targetScaleCenteredData{}, fmt.Errorf("component %d has invalid q0/q1 rows", component)
		}
		componentValues := make([]*big.Int, len(ct.Value[component].Coeffs[0]))
		for index := range componentValues {
			r0 := new(big.Int).SetUint64(ct.Value[component].Coeffs[0][index])
			r1 := new(big.Int).SetUint64(ct.Value[component].Coeffs[1][index])
			t := new(big.Int).Sub(r1, r0)
			t.Mod(t, q1)
			t.Mul(t, q0Inverse)
			t.Mod(t, q1)
			value := new(big.Int).Mul(q0, t)
			value.Add(value, r0)
			if value.Cmp(q01Half) >= 0 {
				value.Sub(value, q01)
			}
			componentValues[index] = value
		}
		values = append(values, componentValues)
	}
	return targetScaleCenteredData{Q0: q0, Q1: q1, Q01: q01, Q01Half: q01Half, Values: values}, nil
}

func targetScaleCapacity(data targetScaleCenteredData, multiplier *big.Int) TargetScaleCapacityEvidence {
	result := TargetScaleCapacityEvidence{Q0: targetScaleIntText(data.Q0), Q1: targetScaleIntText(data.Q1), Q01: targetScaleIntText(data.Q01), Q01Half: targetScaleIntText(data.Q01Half), WorstComponent: -1, WorstIndex: -1, Pass: true}
	if multiplier == nil {
		multiplier = big.NewInt(1)
	}
	maxPre := new(big.Int)
	maxProspective := new(big.Int)
	for component, values := range data.Values {
		componentMax := new(big.Int)
		componentOutside := 0
		for index, value := range values {
			absValue := new(big.Int).Abs(value)
			if absValue.Cmp(maxPre) > 0 {
				maxPre.Set(absValue)
			}
			if absValue.Cmp(componentMax) > 0 {
				componentMax.Set(absValue)
			}
			prospective := new(big.Int).Mul(value, multiplier)
			absProspective := new(big.Int).Abs(prospective)
			if absProspective.Cmp(maxProspective) > 0 {
				maxProspective.Set(absProspective)
				result.WorstComponent = component
				result.WorstIndex = index
				result.WorstCenteredValue = targetScaleIntText(value)
				result.WorstProspectiveValue = targetScaleIntText(prospective)
			}
			if absProspective.Cmp(data.Q01Half) >= 0 {
				componentOutside++
				result.OutsideCount++
			}
		}
		componentRatio, _ := new(big.Float).Quo(new(big.Float).SetInt(componentMax), new(big.Float).SetInt(data.Q01Half)).Float64()
		result.Components = append(result.Components, TargetScaleComponentCapacity{Component: component, MaxAbsCentered: targetScaleIntText(componentMax), MaxAbsOverQ01Half: componentRatio, OutsideCount: componentOutside})
	}
	result.MaxAbsPrePromotion = targetScaleIntText(maxPre)
	result.MaxAbsProspective = targetScaleIntText(maxProspective)
	result.MaxProspectiveRatio, _ = new(big.Float).Quo(new(big.Float).SetInt(maxProspective), new(big.Float).SetInt(data.Q01Half)).Float64()
	result.Pass = result.OutsideCount == 0
	return result
}

func targetScaleMetric(expected, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(expected, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func targetScaleScaleEvidence(s, t rlwe.Scale) (TargetScaleRestorationScaleEvidence, *big.Int, rlwe.Scale) {
	sHigh := new(big.Float).SetPrec(512).Set(&s.Value)
	tHigh := new(big.Float).SetPrec(512).Set(&t.Value)
	ratio := new(big.Float).Quo(new(big.Float).SetPrec(512).Set(tHigh), sHigh)
	multiplier := new(big.Int)
	new(big.Float).Add(ratio, new(big.Float).SetFloat64(0.5)).Int(multiplier)
	relative := new(big.Float).Quo(new(big.Float).Sub(new(big.Float).SetPrec(512).SetInt(multiplier), ratio), ratio)
	relative.Abs(relative)
	sTimesM := new(big.Float).Mul(sHigh, new(big.Float).SetPrec(512).SetInt(multiplier))
	matched := rlwe.NewScale(sTimesM)
	evidence := TargetScaleRestorationScaleEvidence{
		S: targetScaleFloatText(sHigh), T: targetScaleFloatText(tHigh), R: targetScaleFloatText(ratio), M: targetScaleIntText(multiplier),
		RelativeIntegerApproximation: targetScaleFloatText(relative), SM: targetScaleFloatText(sTimesM),
		Log2Delta: matched.Log2Delta(t), NearestIntegerPositive: multiplier.Sign() > 0,
	}
	return evidence, multiplier, matched
}

func targetScaleScaled(values []complex128, factor float64) []complex128 {
	result := make([]complex128, len(values))
	for i, value := range values {
		result[i] = value * complex(factor, 0)
	}
	return result
}

func targetScaleRowsModularCheck(params ckks.Parameters, before, after *rlwe.Ciphertext, multiplier *big.Int) (*TargetScaleModularCheck, error) {
	beforeProjection, _, err := q01Projection(params, before, true)
	if err != nil {
		return nil, err
	}
	afterProjection, _, err := q01Projection(params, after, true)
	if err != nil {
		return nil, err
	}
	result := &TargetScaleModularCheck{Pass: true}
	for component := 0; component <= beforeProjection.Degree(); component++ {
		for limb := 0; limb < 2; limb++ {
			modulus := new(big.Int).SetUint64(params.RingQ().SubRings[limb].Modulus)
			factor := new(big.Int).Mod(new(big.Int).Set(multiplier), modulus)
			beforeRow := beforeProjection.Value[component].Coeffs[limb]
			afterRow := afterProjection.Value[component].Coeffs[limb]
			for index := range beforeRow {
				want := new(big.Int).Mul(new(big.Int).SetUint64(beforeRow[index]), factor)
				want.Mod(want, modulus)
				if afterRow[index] != want.Uint64() {
					result.Pass = false
					result.MismatchCount++
					if result.FirstMismatch == "" {
						result.FirstMismatch = fmt.Sprintf("component=%d limb=%d index=%d expected=%d actual=%d", component, limb, index, want.Uint64(), afterRow[index])
					}
				}
			}
		}
	}
	return result, nil
}

func targetScaleCheckpoint(state string, ct *rlwe.Ciphertext) TargetScaleRestorationCheckpoint {
	return TargetScaleRestorationCheckpoint{State: state, Level: ct.Level(), Degree: ct.Degree(), Scale: finalizationScaleString(ct.Scale), Rows: finalizationEvidence(ct).Rows}
}

func writeTargetScaleRestorationResult(result TargetScaleRestorationResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for target-scale restoration diagnostic")
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
		SchemaVersion       string                              `json:"schema_version"`
		Timestamp           time.Time                           `json:"timestamp"`
		Primary             RepositoryMetadata                  `json:"primary_repository"`
		Lattigo             RepositoryMetadata                  `json:"lattigo_repository"`
		AuthoritativeOracle string                              `json:"authoritative_poly_oracle"`
		Threshold           float64                             `json:"threshold"`
		Scale               TargetScaleRestorationScaleEvidence `json:"scale_restoration"`
		PreFinal            TargetScaleRestorationCheckpoint    `json:"canonical_pre_final_rescale"`
		P0                  TargetScaleRestorationCheckpoint    `json:"p0_baseline_compressed_output"`
		P0Capacity          TargetScaleCapacityEvidence         `json:"p0_capacity"`
		P1Capacity          TargetScaleCapacityEvidence         `json:"p1_prospective_capacity"`
		P2                  TargetScaleRestorationCheckpoint    `json:"p2_integer_promotion"`
		P3                  TargetScaleRestorationCheckpoint    `json:"p3_matched_scale"`
		P4                  TargetScaleRestorationCheckpoint    `json:"p4_exact_target_metadata"`
		P5Invariants        map[string]bool                     `json:"p5_restored_state_invariants"`
		FirstSupportedCause string                              `json:"first_supported_cause"`
		Validation          map[string]interface{}              `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.AuthoritativeOracle, result.Threshold, result.Scale, result.PreFinal, result.P0, result.P0Capacity, result.P1Capacity, result.P2, result.P3, result.P4, result.P5Invariants, result.FirstSupportedCause, result.Validation}
	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DIAG-TARGET-SCALE-RESTORATION-logN13-summary.json"), summaryData, 0o644)
}

func runFIX001P3TargetScaleRestoration(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	o2 := chebyshevRawOracle(state.Poly, state.Z)
	preMetric := psGlobalMetric(o2, state.Root.actual)
	preFinal := targetScaleCheckpoint("canonical_pre_final_fast_rescale", state.Root.ciphertext)
	preFinal.CorrectedOracle = preMetric
	post := psGlobalCopy(state.Params.BootstrappingParameters, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(post, post); err != nil {
		return err
	}
	postDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, post)
	if err != nil {
		return err
	}
	p0Metric := psGlobalMetric(o2, postDecoded)
	p0 := targetScaleCheckpoint("P0_post_final_fast_rescale", post)
	p0.CorrectedOracle = p0Metric
	result := TargetScaleRestorationResult{
		SchemaVersion: "fix-001-p3-diag-target-scale-restoration.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(state.Params.BootstrappingParameters, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: targetScaleRestorationThreshold, PreFinal: preFinal, P0: p0,
		Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "canonical_root_hash_match": state.Root.RowsMatch([]string{"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf", "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"}), "corrected_pre_final_pass": preMetric.Pass, "corrected_post_final_pass": p0Metric.Pass, "corrected_pre_final_max_component": preMetric.MaxComponent, "corrected_post_final_max_component": p0Metric.MaxComponent, "historical_o1_used_for_control": false, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_double_angle": true, "no_exp003": true},
	}
	if !result.Validation["canonical_root_hash_match"].(bool) || !p0Metric.Pass {
		result.FirstSupportedCause = "target_scale_restoration_precondition_mismatch"
		return writeTargetScaleRestorationResult(result, outPath)
	}

	p0Projection, _, err := q01Projection(state.Params.BootstrappingParameters, post, true)
	if err != nil {
		return err
	}
	centered, err := targetScaleCenteredRows(state.Params.BootstrappingParameters, p0Projection)
	if err != nil {
		return err
	}
	scaleEvidence, multiplier, matchedScale := targetScaleScaleEvidence(post.Scale, state.Target)
	result.Scale = scaleEvidence
	result.P0Capacity = targetScaleCapacity(centered, big.NewInt(1))
	p1Projection, _, err := q01Projection(state.Params.BootstrappingParameters, post, true)
	if err != nil {
		return err
	}
	centeredP1, err := targetScaleCenteredRows(state.Params.BootstrappingParameters, p1Projection)
	if err != nil {
		return err
	}
	result.P1Capacity = targetScaleCapacity(centeredP1, multiplier)
	if multiplier.Sign() <= 0 || !result.P1Capacity.Pass {
		result.FirstSupportedCause = "target_scale_restoration_centered_capacity_failure"
		return writeTargetScaleRestorationResult(result, outPath)
	}

	promoted := psGlobalCopy(state.Params.BootstrappingParameters, post)
	if err := state.Eval.FastCKKS.MulIntegerMaintained(promoted, multiplier, promoted); err != nil {
		return err
	}
	promotedDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	multiplierFloat, _ := new(big.Float).SetInt(multiplier).Float64()
	p2Scaled := targetScaleMetric(targetScaleScaled(o2, multiplierFloat), promotedDecoded, targetScaleRestorationThreshold*math.Max(1, multiplierFloat))
	p2Normalized := targetScaleMetric(o2, targetScaleScaled(promotedDecoded, 1/multiplierFloat), targetScaleRestorationThreshold)
	modularCheck, err := targetScaleRowsModularCheck(state.Params.BootstrappingParameters, post, promoted, multiplier)
	if err != nil {
		return err
	}
	p2 := targetScaleCheckpoint("P2_integer_promotion_at_original_scale", promoted)
	p2.ScaledOracle, p2.NormalizedOracle, p2.IndependentModularCheck = p2Scaled, p2Normalized, modularCheck
	result.P2 = p2
	if !p2Scaled.Pass || !p2Normalized.Pass || !modularCheck.Pass {
		result.FirstSupportedCause = "target_scale_restoration_integer_promotion_failure"
		return writeTargetScaleRestorationResult(result, outPath)
	}

	p3Decoded := promotedDecoded
	promoted.Scale = matchedScale
	p3Decoded, err = psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	p3 := targetScaleCheckpoint("P3_integer_promotion_with_matched_scale", promoted)
	p3.CorrectedOracle = psGlobalMetric(o2, p3Decoded)
	p3.PerturbationFromP0 = psGlobalMetric(postDecoded, p3Decoded)
	result.P3 = p3
	if !p3.CorrectedOracle.Pass {
		result.FirstSupportedCause = "target_scale_restoration_matched_scale_failure"
		return writeTargetScaleRestorationResult(result, outPath)
	}

	if scaleEvidence.Log2Delta < 32 {
		result.FirstSupportedCause = "target_scale_restoration_integer_approximation_too_coarse"
		return writeTargetScaleRestorationResult(result, outPath)
	}
	p4Before := append([]complex128(nil), p3Decoded...)
	promoted.Scale = state.Target
	p4Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	p4 := targetScaleCheckpoint("P4_exact_public_target_scale_metadata", promoted)
	p4.CorrectedOracle = psGlobalMetric(o2, p4Decoded)
	p4.PerturbationFromP3 = psGlobalMetric(p4Before, p4Decoded)
	result.P4 = p4
	if !p4.CorrectedOracle.Pass {
		result.FirstSupportedCause = "target_scale_restoration_metadata_normalization_failure"
		return writeTargetScaleRestorationResult(result, outPath)
	}

	p2Rows := result.P2.Rows
	p5 := map[string]bool{
		"scale_equals_exact_public_target": promoted.Scale.Equal(state.Target),
		"level_unchanged_from_p0":          promoted.Level() == post.Level(),
		"degree_unchanged_from_p0":         promoted.Degree() == post.Degree(),
		"q01_hashes_unchanged_p2_to_p3":    q01RowsEqual(p2Rows, result.P3.Rows),
		"q01_hashes_unchanged_p3_to_p4":    q01RowsEqual(result.P3.Rows, result.P4.Rows),
		"centered_capacity_valid":          result.P1Capacity.Pass,
		"corrected_semantic_error_pass":    p4.CorrectedOracle.Pass,
		"historical_o1_not_used":           true,
	}
	result.P5Invariants = p5
	result.FirstSupportedCause = "compressed_ps_2p91_target_scale_restoration_validated"
	for _, pass := range p5 {
		if !pass {
			result.FirstSupportedCause = "target_scale_restoration_metadata_normalization_failure"
			break
		}
	}
	return writeTargetScaleRestorationResult(result, outPath)
}
