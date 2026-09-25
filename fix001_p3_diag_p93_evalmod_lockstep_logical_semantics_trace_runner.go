//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const fix001P3P93LogicalTraceSchema = "fix-001-p3-diag-p93-evalmod-lockstep-logical-semantics-trace.v1"

type fix001P3P93LogicalCheckpoint struct {
	Name           string                 `json:"name"`
	Standard       semanticBisectMetadata `json:"standard"`
	FastRaw        semanticBisectMetadata `json:"fast_raw"`
	FastVirtualExp int                    `json:"fast_virtual_exponent"`
	Metric         *semanticBisectMetric  `json:"canonical_fast_vs_standard,omitempty"`
	RawDebugMetric *semanticBisectMetric  `json:"raw_coordinate_debug,omitempty"`
	FastCapacity   *postMod1S2CCapacity   `json:"fast_q012_capacity,omitempty"`
	LogicalValid   bool                   `json:"logical_comparison_valid"`
	StandardOp     string                 `json:"standard_operation"`
	FastOp         string                 `json:"fast_operation"`
}

type fix001P3P93LogicalBranch struct {
	Name               string                         `json:"name"`
	Checkpoints        []fix001P3P93LogicalCheckpoint `json:"checkpoints"`
	GeneratedPowers    []map[string]interface{}       `json:"generated_powers,omitempty"`
	PSPlan             map[string]interface{}         `json:"ps_plan,omitempty"`
	StandardReplay     *semanticBisectMetric          `json:"standard_replay_vs_actual"`
	FastReplay         *semanticBisectMetric          `json:"fast_replay_vs_actual"`
	StandardReplayPass bool                           `json:"standard_replay_pass"`
	FastReplayPass     bool                           `json:"fast_replay_pass"`
	InternalFinal      *semanticBisectMetric          `json:"internal_final_fast_vs_standard"`
	PublicLike         *semanticBisectMetric          `json:"public_like_fast_vs_standard"`
	Factor32           float64                        `json:"public_to_internal_max_component_ratio"`
}

type fix001P3P93LogicalResult struct {
	SchemaVersion   string                              `json:"schema_version"`
	Timestamp       time.Time                           `json:"timestamp"`
	Primary         RepositoryMetadata                  `json:"primary_repository"`
	Secondary       RepositoryMetadata                  `json:"secondary_repository"`
	Environment     EnvironmentMetadata                 `json:"environment"`
	Config          BootstrapConfig                     `json:"config"`
	FixedProfile    map[string]interface{}              `json:"fixed_profile"`
	R0Defect        map[string]interface{}              `json:"r0_previous_trace_defect"`
	R0C2S           map[string]interface{}              `json:"r0_c2s_precondition"`
	Branches        map[string]fix001P3P93LogicalBranch `json:"branches"`
	FirstObservable map[string]interface{}              `json:"first_observable_divergence"`
	FirstMaterial   map[string]interface{}              `json:"first_material_divergence"`
	Factor32        map[string]interface{}              `json:"factor_32_public_mapping"`
	Classification  string                              `json:"classification"`
	SourceAudit     map[string]interface{}              `json:"source_audit"`
	Validation      map[string]interface{}              `json:"validation"`
}

type fix001P3P93LogicalPair struct {
	Name       string
	Standard   string
	Exponent   int
	StandardOp string
	FastOp     string
}

func fix001P3P93CanonicalFastValuesForParams(ct *rlwe.Ciphertext, params ckks.Parameters, sk *rlwe.SecretKey, exponent int) ([]complex128, error) {
	_, values, err := semanticBisectView(ct, params, sk)
	if err != nil {
		return nil, err
	}
	if exponent == 0 {
		return values, nil
	}
	for i := range values {
		values[i] = fix001P3P93CanonicalValue(values[i], exponent)
	}
	return values, nil
}

func fix001P3P93CanonicalValue(value complex128, exponent int) complex128 {
	factor := math.Ldexp(1, exponent)
	return complex(real(value)*factor, imag(value)*factor)
}

func fix001P3P93LogicalPairs() []fix001P3P93LogicalPair {
	pairs := []fix001P3P93LogicalPair{
		{Name: "input", Standard: "evalmod_input", Exponent: 0, StandardOp: "input", FastOp: "input"},
		{Name: "normalize_scale", Standard: "normalize_scale", Exponent: 0, StandardOp: "assign Mod1 ScalingFactor", FastOp: "assign Mod1 ScalingFactor"},
		{Name: "offset", Standard: "apply_offset", Exponent: 0, StandardOp: "CosDiscrete offset Add", FastOp: "CosDiscrete offset Add"},
		{Name: "polynomial_input", Standard: "polynomial_input", Exponent: 0, StandardOp: "polynomial input", FastOp: "polynomial input"},
		{Name: "polynomial_output", Standard: "polynomial_output_after_rescale", Exponent: 0, StandardOp: "Standard PS output", FastOp: "Fast PS output"},
		{Name: "pre_double_angle_logical", Standard: "polynomial_output_after_rescale", Exponent: 27, StandardOp: "same logical polynomial output", FastOp: "metadata-only coherent scale; e=27"},
	}
	for round := 0; round < 3; round++ {
		pairs = append(pairs,
			fix001P3P93LogicalPair{Name: fmt.Sprintf("double_angle_round_%d_square", round), Standard: fmt.Sprintf("double_angle_round_%d_square", round), Exponent: 54, StandardOp: "MulRelin square", FastOp: "MulRelin square; e=2*27"},
			fix001P3P93LogicalPair{Name: fmt.Sprintf("double_angle_round_%d_after_constant", round), Standard: fmt.Sprintf("double_angle_round_%d_after_constant", round), Exponent: 27, StandardOp: "double + constant", FastOp: "maintained multiplier + constant; e=27"},
			fix001P3P93LogicalPair{Name: fmt.Sprintf("double_angle_round_%d_post_rescale", round), Standard: fmt.Sprintf("double_angle_round_%d_post_rescale", round), Exponent: 27, StandardOp: "Rescale", FastOp: "Rescale; e=27"},
		)
	}
	pairs = append(pairs,
		fix001P3P93LogicalPair{Name: "internal_final_restore", Standard: "internal_final_restore", Exponent: 0, StandardOp: "restore caller input Scale", FastOp: "materialize 2^e and restore input Scale; e=0"},
		fix001P3P93LogicalPair{Name: "public_reset_context", Standard: "public_reset_context", Exponent: 0, StandardOp: "DefaultScale context only", FastOp: "DefaultScale context only"},
	)
	return pairs
}

func fix001P3P93FastStateName(pairName string) string {
	switch pairName {
	case "input":
		return "evalmod_input"
	case "offset":
		return "apply_offset"
	case "polynomial_output":
		return "polynomial_output_after_rescale"
	case "pre_double_angle_logical":
		return "pre_double_angle"
	}
	return pairName
}

func fix001P3P93LogicalCheckpointForPair(pair fix001P3P93LogicalPair, standard, fast *rlwe.Ciphertext, params ckks.Parameters, standardSK, fastSK *rlwe.SecretKey) (fix001P3P93LogicalCheckpoint, error) {
	result := fix001P3P93LogicalCheckpoint{Name: pair.Name, FastVirtualExp: pair.Exponent, StandardOp: pair.StandardOp, FastOp: pair.FastOp}
	stdMeta, stdValues, err := semanticBisectView(standard, params, standardSK)
	if err != nil {
		return result, err
	}
	fastMeta, _, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return result, err
	}
	result.Standard, result.FastRaw = stdMeta, fastMeta
	if fast.Level() >= 1 {
		if capacity, capacityErr := postMod1S2CCapacityFromFastCiphertext(params, fast); capacityErr == nil {
			result.FastCapacity = &capacity
		}
	}
	if pair.Name == "public_reset_context" {
		return result, nil
	}
	fastValues, err := fix001P3P93CanonicalFastValuesForParams(fast, params, fastSK, pair.Exponent)
	if err != nil {
		return result, err
	}
	metric, err := semanticBisectMetricFor(stdValues, fastValues, fix001P3P93InternalBudget)
	if err != nil {
		return result, err
	}
	result.Metric = &metric
	result.LogicalValid = stdMeta.N == fastMeta.N && stdMeta.Degree == fastMeta.Degree && len(stdValues) == len(fastValues)
	return result, nil
}

func fix001P3P93LogicalPowerSummary(powers []fix001P3P93InternalPower) []map[string]interface{} {
	wanted := map[int]bool{2: true, 3: true, 4: true, 6: true, 8: true, 16: true}
	rows := make([]map[string]interface{}, 0, len(wanted))
	for _, power := range powers {
		if !wanted[power.Power] {
			continue
		}
		row := map[string]interface{}{"power": power.Power, "comparable": power.Metric != nil}
		if power.Metric != nil {
			row["metric"] = power.Metric
		}
		rows = append(rows, row)
	}
	return rows
}

func fix001P3P93LogicalFirst(branches map[string]fix001P3P93LogicalBranch, threshold float64) (map[string]interface{}, map[string]interface{}) {
	firstObservable := map[string]interface{}{"checkpoint": "none", "threshold": 1e-12}
	firstMaterial := map[string]interface{}{"checkpoint": "none", "threshold": threshold}
	for _, branchName := range []string{"real", "imag"} {
		branch := branches[branchName]
		for index, checkpoint := range branch.Checkpoints {
			if checkpoint.Metric == nil || !checkpoint.LogicalValid {
				continue
			}
			difference := checkpoint.Metric.MaxComponentAbs
			if firstObservable["checkpoint"] == "none" && difference > 1e-12 {
				firstObservable = map[string]interface{}{"branch": branchName, "checkpoint": checkpoint.Name, "metric": checkpoint.Metric, "fast_virtual_exponent": checkpoint.FastVirtualExp}
			}
			if firstMaterial["checkpoint"] == "none" && difference >= threshold {
				incoming := 0.0
				if index > 0 && branch.Checkpoints[index-1].Metric != nil {
					incoming = branch.Checkpoints[index-1].Metric.MaxComponentAbs
				}
				firstMaterial = map[string]interface{}{"branch": branchName, "checkpoint": checkpoint.Name, "incoming_max_component": incoming, "outgoing_max_component": difference, "metric": checkpoint.Metric, "fast_virtual_exponent": checkpoint.FastVirtualExp, "level": checkpoint.FastRaw.SourceLevel, "scale": checkpoint.FastRaw.Scale, "degree": checkpoint.FastRaw.Degree, "fast_q012_capacity": checkpoint.FastCapacity, "operation": checkpoint.FastOp, "threshold": threshold}
			}
		}
	}
	return firstObservable, firstMaterial
}

func fix001P3P93LogicalClassification(firstMaterial map[string]interface{}) string {
	checkpoint, _ := firstMaterial["checkpoint"].(string)
	switch {
	case checkpoint == "none":
		return "H"
	case checkpoint == "polynomial_output" || checkpoint == "pre_double_angle_logical":
		return "C"
	case strings.HasPrefix(checkpoint, "double_angle_round_0"):
		return "D"
	case strings.HasPrefix(checkpoint, "double_angle_round_1"):
		return "E"
	case strings.HasPrefix(checkpoint, "double_angle_round_2"):
		return "F"
	case checkpoint == "internal_final_restore":
		return "G"
	default:
		return "I"
	}
}

func fix001P3P93LogicalWrite(result fix001P3P93LogicalResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if len(data) > 300*160 {
		data, err = json.Marshal(result)
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagP93EvalModLockstepLogicalSemantics(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3P93LogicalResult{
		SchemaVersion: fix001P3P93LogicalTraceSchema,
		Timestamp:     time.Now().UTC(),
		Primary:       gitMetadata(primaryRoot),
		Secondary:     gitMetadata(backendRoot),
		Environment:   EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:        cfg,
		FixedProfile:  map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps": "Q012-wide", "plan_scale": "2^93", "material_threshold_internal": fix001P3P93InternalBudget, "public_threshold": fix001P3P93InternalPublicThreshold},
		R0Defect:      map[string]interface{}{"classification": "invalid_raw_coordinate_comparison", "prior_runner": "fix001_p3_diag_p93_fast_vs_genuine_standard_evalmod_internal_bisect_runner.go", "fast_coherent_scale_lines": "324-332", "fast_double_angle_lines": "335-369", "paired_standard_lines": "276, 313-318", "defect": "Fast advanced through DoubleAngle while the prior paired Standard object remained at polynomial output; Fast metadata-only coherent-scale reassignment was compared in raw coordinates", "old_pattern": map[string]float64{"polynomial_output_max_component": 5.07761187318323e-7, "raw_pre_double_angle_max_component": 0.7794741220173813, "next_raw_square_max_component": 7.156194948164926e-7}},
		Validation:    map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "secondary_required_head": requiredFIX001P3P93InternalSecondary, "secondary_required_diff_sha256": requiredFIX001P3P93InternalDiffSHA, "secondary_not_modified": true, "no_parameter_tuning": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "raw_metrics_not_used_for_causality": true},
		Branches:      map[string]fix001P3P93LogicalBranch{},
		Factor32:      map[string]interface{}{"expected_ratio": 32, "contract": "public wrappers reset DefaultScale after internal restore"},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3P93InternalSecondary || diffHash != requiredFIX001P3P93InternalDiffSHA || !secondaryDirty {
		result.Classification = "I"
		return fix001P3P93LogicalWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params, fastSK := profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters)
	pairs := fix001P3P93LogicalPairs()
	for _, branchName := range []string{"real", "imag"} {
		fastInput, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
		if branchName == "imag" {
			fastInput, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
		}
		stdMeta, stdValues, err := semanticBisectView(standardInput, params, profile.StandardSK)
		if err != nil {
			return err
		}
		fastMeta, fastValues, err := semanticBisectView(fastInput, params, fastSK)
		if err != nil {
			return err
		}
		c2sMetric, err := semanticBisectMetricFor(stdValues, fastValues, fix001P3P93InternalC2SThreshold)
		if err != nil {
			return err
		}
		if result.R0C2S == nil {
			result.R0C2S = map[string]interface{}{}
		}
		result.R0C2S[branchName] = map[string]interface{}{"standard": stdMeta, "fast": fastMeta, "metric": c2sMetric, "pass": c2sMetric.Pass}
		if !c2sMetric.Pass {
			result.Classification = "A"
			return fix001P3P93LogicalWrite(result, outPath)
		}
		stdBranch, stdActual, stdReplay, err := fix001P3P93InternalStandardTrace(standardInput, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		fastBranch, fastActual, fastReplay, err := fix001P3P93InternalFastTrace(fastInput, standardInput, profile.Fast, profile.Standard, params, profile.StandardSK, fastSK)
		if err != nil {
			return err
		}
		stdActualMeta, stdActualValues, err := semanticBisectView(stdActual, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, stdReplayValues, err := semanticBisectView(stdReplay, params, profile.StandardSK)
		if err != nil {
			return err
		}
		stdReplayMetric, err := semanticBisectMetricFor(stdActualValues, stdReplayValues, 1e-10)
		if err != nil {
			return err
		}
		fastActualMeta, fastActualValues, err := semanticBisectView(fastActual, params, fastSK)
		if err != nil {
			return err
		}
		_, fastReplayValues, err := semanticBisectView(fastReplay, params, fastSK)
		if err != nil {
			return err
		}
		fastReplayMetric, err := semanticBisectMetricFor(fastActualValues, fastReplayValues, 1e-10)
		if err != nil {
			return err
		}
		branch := fix001P3P93LogicalBranch{Name: branchName, Checkpoints: make([]fix001P3P93LogicalCheckpoint, 0, len(pairs)), StandardReplay: &stdReplayMetric, FastReplay: &fastReplayMetric, StandardReplayPass: stdReplayMetric.Pass, FastReplayPass: fastReplayMetric.Pass}
		for _, pair := range pairs {
			standardCT := stdBranch.states[pair.Standard]
			fastCT := fastBranch.states[fix001P3P93FastStateName(pair.Name)]
			if standardCT == nil || fastCT == nil {
				return fmt.Errorf("missing lockstep state %s: standard=%t fast=%t", pair.Name, standardCT != nil, fastCT != nil)
			}
			checkpoint, err := fix001P3P93LogicalCheckpointForPair(pair, standardCT, fastCT, params, profile.StandardSK, fastSK)
			if err != nil {
				return fmt.Errorf("checkpoint %s: %w", pair.Name, err)
			}
			branch.Checkpoints = append(branch.Checkpoints, checkpoint)
		}
		for _, checkpoint := range branch.Checkpoints {
			if checkpoint.Name == "internal_final_restore" {
				branch.InternalFinal = checkpoint.Metric
			}
		}
		targetScale, err := fix001P3P93InternalTargetScale(fastInput, profile.Fast.Mod1Parameters, params)
		if err != nil {
			return err
		}
		powers, plan, err := fix001P3P93InternalPowerTrace(stdBranch.states["polynomial_input"], fastBranch.states["polynomial_input"], profile.Fast, profile.Standard, params, profile.StandardSK, fastSK, targetScale, profile.Fast.Mod1Parameters.Mod1Poly)
		if err != nil {
			return err
		}
		branch.GeneratedPowers = fix001P3P93LogicalPowerSummary(powers)
		branch.PSPlan = map[string]interface{}{"degree": plan["degree"], "base": plan["base"], "level": plan["level"], "scale": plan["scale"], "fast_generated_power_count": plan["fast_generated_power_count"], "standard_generated_power_count": plan["standard_generated_power_count"]}
		stdPublic, err := profile.Standard.EvalMod(standardInput.CopyNew())
		if err != nil {
			return err
		}
		fastPublic, err := profile.Fast.EvalMod(fastInput.CopyNew())
		if err != nil {
			return err
		}
		_, stdPublicValues, err := semanticBisectView(stdPublic, params, profile.StandardSK)
		if err != nil {
			return err
		}
		_, fastPublicValues, err := semanticBisectView(fastPublic, params, fastSK)
		if err != nil {
			return err
		}
		publicMetric, err := semanticBisectMetricFor(stdPublicValues, fastPublicValues, fix001P3P93InternalPublicThreshold)
		if err != nil {
			return err
		}
		branch.PublicLike = &publicMetric
		internalMetric, err := semanticBisectMetricFor(stdActualValues, fastActualValues, fix001P3P93InternalBudget)
		if err != nil {
			return err
		}
		branch.InternalFinal = &internalMetric
		branch.Factor32 = safeRatio(publicMetric.MaxComponentAbs, internalMetric.MaxComponentAbs)
		result.Factor32[branchName] = map[string]interface{}{"internal_metric": internalMetric, "public_metric": publicMetric, "ratio": branch.Factor32, "standard_internal_scale": stdActualMeta.Scale, "fast_internal_scale": fastActualMeta.Scale, "standard_public_scale": finalizationScaleString(stdPublic.Scale), "fast_public_scale": finalizationScaleString(fastPublic.Scale)}
		result.Branches[branchName] = branch
	}
	result.FirstObservable, result.FirstMaterial = fix001P3P93LogicalFirst(result.Branches, fix001P3P93InternalBudget)
	result.Classification = fix001P3P93LogicalClassification(result.FirstMaterial)
	result.SourceAudit = map[string]interface{}{"standard_internal": "lattigo/circuits/ckks/mod1/mod1_evaluator.go:EvaluateNew", "fast_internal": "lattigo/circuits/ckks/mod1/fast.go:EvaluateNew/evaluateNormalizedLogN13", "standard_double_angle": "MulRelin -> Add -> Add(constant) -> Rescale", "fast_double_angle": "MulRelin -> MulIntegerMaintained(2^a) -> Add(constant) -> Rescale", "canonical_fast": "decoded Fast raw values multiplied by 2^e in diagnostic memory only", "logical_equations": "pre-DA e=27; square e=54; multiplier+constant/post-Rescale e=27; final restore e=0", "public_wrapper": "both wrappers reset DefaultScale after internal restore"}
	allReplay := true
	for _, branch := range result.Branches {
		allReplay = allReplay && branch.StandardReplayPass && branch.FastReplayPass
	}
	if !allReplay {
		result.Classification = "I"
	}
	return fix001P3P93LogicalWrite(result, outPath)
}
