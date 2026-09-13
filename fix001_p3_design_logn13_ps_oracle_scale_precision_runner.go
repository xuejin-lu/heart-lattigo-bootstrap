package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3PSOracleScalePrimaryBase = "08e8c3b150d029fc961925a93b79437cb54ddb5b"
	requiredFIX001P3PSOracleScaleSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	psOracleScaleBudget                      = 1.2e-8
	psOracleScaleRoundoff                    = 1e-10
	psOracleScalePublicThreshold             = 1e-2
)

var psOracleScalePlanExponents = []int{91, 92, 93, 94, 95, 96}

type psOracleScaleMetricPair struct {
	Real *PSGlobalMetric `json:"real"`
	Imag *PSGlobalMetric `json:"imag"`
}

type psOracleScaleCandidate struct {
	Exponent                   int                     `json:"plan_scale_exponent"`
	PlanScale                  string                  `json:"plan_scale"`
	OutputScaleReal            string                  `json:"output_scale_real,omitempty"`
	OutputScaleImag            string                  `json:"output_scale_imag,omitempty"`
	FirstFailure               string                  `json:"first_failure"`
	FirstCumulativeBudgetCross string                  `json:"first_cumulative_budget_crossing"`
	WorstCapacityCheckpoint    string                  `json:"worst_capacity_checkpoint"`
	WorstCapacityRatio         float64                 `json:"worst_capacity_ratio"`
	CapacitySafe               bool                    `json:"capacity_safe"`
	AllCenteredUnique          bool                    `json:"all_centered_unique"`
	AllRowsMatch               bool                    `json:"all_fast_vs_full_rns_rows_match"`
	FastVsFullRNS              psOracleScaleMetricPair `json:"fast_vs_stage_aligned_full_rns"`
	F0PreError                 psOracleScaleMetricPair `json:"f0_final_rescale_pre_error"`
	F0LocalError               psOracleScaleMetricPair `json:"f0_final_rescale_local_error"`
	F0PostError                psOracleScaleMetricPair `json:"f0_final_rescale_post_error"`
	F0ErrorRatioToNextReal     float64                 `json:"f0_error_ratio_to_next_real,omitempty"`
	F0ErrorRatioToNextImag     float64                 `json:"f0_error_ratio_to_next_imag,omitempty"`
	FinalVsPlainOracle         psOracleScaleMetricPair `json:"final_vs_plaintext_polynomial_oracle"`
	FinalVsStandard            psOracleScaleMetricPair `json:"final_vs_genuine_standard"`
	OutputScale                string                  `json:"output_scale,omitempty"`
	CheckpointCount            int                     `json:"checkpoint_count"`
	PrecisionQualified         bool                    `json:"precision_qualified"`
	DownstreamReached          bool                    `json:"downstream_reached"`
	DownstreamReason           string                  `json:"downstream_reason,omitempty"`
	EvalModVsStandard          psOracleScaleMetricPair `json:"downstream_evalmod_vs_standard,omitempty"`
	PostS2CVsStandard          psOracleScaleMetricPair `json:"downstream_post_s2c_vs_standard,omitempty"`
	PublicLike                 *PSGlobalMetric         `json:"downstream_public_like,omitempty"`
	StandardPublicLike         *PSGlobalMetric         `json:"downstream_standard_public_like,omitempty"`
	PublicLikeMetadataCorrect  bool                    `json:"downstream_public_metadata_correct"`
	PublicLikeInputUnchanged   bool                    `json:"downstream_input_unchanged"`
	Classification             string                  `json:"classification"`
	firstFailureOrder          int
	realFinal                  *rlwe.Ciphertext
	imagFinal                  *rlwe.Ciphertext
}

type psOracleScaleResult struct {
	SchemaVersion        string                   `json:"schema_version"`
	Timestamp            time.Time                `json:"timestamp"`
	Primary              RepositoryMetadata       `json:"primary_repository"`
	Lattigo              RepositoryMetadata       `json:"lattigo_repository"`
	Environment          EnvironmentMetadata      `json:"environment"`
	Config               BootstrapConfig          `json:"config"`
	Parameters           ExperimentParameters     `json:"effective_parameters"`
	Workload             CorrectnessWorkload      `json:"workload"`
	Budget               float64                  `json:"polynomial_budget"`
	PublicThreshold      float64                  `json:"public_like_threshold"`
	BaselineReproduction map[string]interface{}   `json:"baseline_reproduction"`
	Candidates           []psOracleScaleCandidate `json:"candidates"`
	SelectedExponent     int                      `json:"selected_exponent,omitempty"`
	Classification       string                   `json:"classification"`
	FirstLimiting        string                   `json:"first_limiting_checkpoint"`
	Validation           map[string]interface{}   `json:"validation"`
}

type psOracleScaleBranch struct {
	Name     string
	Base     polynomialDecompositionBranchWork
	Standard []complex128
	Input    *rlwe.Ciphertext
	Workload []complex128
}

func psOracleScalePlan(base commonpolynomial.PatersonStockmeyerPolynomial, scale rlwe.Scale) commonpolynomial.PatersonStockmeyerPolynomial {
	plan := base
	for i := range plan.Value {
		plan.Value[i].Scale = scale
	}
	return plan
}

func psOracleScalePair(expected, actual []complex128, threshold float64) psOracleScaleMetricPair {
	return psOracleScaleMetricPair{
		Real: polynomialDecompositionComponentMetric(expected, actual, threshold, true),
		Imag: polynomialDecompositionComponentMetric(expected, actual, threshold, false),
	}
}

func psOracleScaleRoundoffPair() psOracleScaleMetricPair {
	metric := &PSGlobalMetric{Pass: true, Threshold: psOracleScaleRoundoff, WorstIndex: -1}
	return psOracleScaleMetricPair{Real: metric, Imag: metric}
}

func allPolynomialDecompositionOracleRowsMatch(powers []polynomialDecompositionPowerSummary) bool {
	for _, power := range powers {
		if !power.OracleRowsMatch {
			return false
		}
	}
	return true
}

func psOracleScaleWorstPair(a, b psOracleScaleMetricPair) psOracleScaleMetricPair {
	choose := func(x, y *PSGlobalMetric) *PSGlobalMetric {
		if x == nil {
			return y
		}
		if y == nil || x.MaxComponent >= y.MaxComponent {
			return x
		}
		return y
	}
	return psOracleScaleMetricPair{Real: choose(a.Real, b.Real), Imag: choose(a.Imag, b.Imag)}
}

func psOracleScaleCheckpointCapacity(params ckks.Parameters, checkpoint *PSGlobalCheckpoint) (float64, bool, bool, error) {
	if checkpoint == nil || checkpoint.ciphertext == nil || checkpoint.ciphertext.Level() < 1 {
		return 0, true, true, nil
	}
	capacity, err := postMod1S2CCapacityFromFastCiphertext(params, checkpoint.ciphertext)
	if err != nil {
		return 0, false, false, err
	}
	rowsMatch := false
	if capacity.Pass {
		full, _, err := postMod1S2CLiftFull(params, checkpoint.ciphertext)
		if err != nil {
			return capacity.MaxAbsOverQ01Half, capacity.Pass, false, err
		}
		rowsMatch = postMod1S2CRowsEqual(params, checkpoint.ciphertext, full)
	}
	return capacity.MaxAbsOverQ01Half, capacity.Pass, rowsMatch, nil
}

func psOracleScaleFailureClass(id string, capacity bool) string {
	if capacity {
		switch {
		case strings.HasPrefix(id, "B"):
			return "ps_oracle_baby_capacity_failure"
		case strings.HasPrefix(id, "G"):
			return "ps_oracle_giant_capacity_failure"
		case strings.HasPrefix(id, "F0"):
			return "ps_oracle_final_rescale_capacity_failure"
		default:
			return "ps_oracle_scalar_capacity_failure"
		}
	}
	return "ps_oracle_primitive_mismatch"
}

func psOracleScaleCandidateForBranch(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch psOracleScaleBranch, exponent int) (psOracleScaleCandidate, error) {
	planScale := precisionSweepScale(exponent)
	candidate := psOracleScaleCandidate{Exponent: exponent, PlanScale: finalizationScaleString(planScale), FirstFailure: "none", FirstCumulativeBudgetCross: "none", WorstCapacityRatio: 0, AllCenteredUnique: true, AllRowsMatch: true, Classification: "ps_oracle_stage_pass", firstFailureOrder: math.MaxInt}
	plan := psOracleScalePlan(branch.Base.Plan, planScale)
	checks, _, root, _, replayErr := psGlobalReplay(params, eval.FastCKKS, plan, branch.Base.OraclePowerMap, branch.Base.PowerExpected, branch.Base.OraclePowerValues, branch.Base.InputValues, planScale)
	candidate.CheckpointCount = len(checks) + 1
	attribution := true
	check := func(id string, cp *PSGlobalCheckpoint) error {
		if cp != nil && cp.SourceBacked != nil && cp.SourceBacked.MaxComponent > psOracleScaleBudget && candidate.FirstCumulativeBudgetCross == "none" {
			candidate.FirstCumulativeBudgetCross = fmt.Sprintf("%s:%s", branch.Name, id)
		}
		ratio, centered, rowsMatch, err := psOracleScaleCheckpointCapacity(params, cp)
		if err != nil {
			return err
		}
		if ratio > candidate.WorstCapacityRatio {
			candidate.WorstCapacityRatio, candidate.WorstCapacityCheckpoint = ratio, fmt.Sprintf("%s:%s", branch.Name, id)
		}
		if !centered && candidate.FirstFailure == "none" {
			candidate.FirstFailure = fmt.Sprintf("%s:%s", branch.Name, id)
			candidate.firstFailureOrder = len(checks)
			candidate.Classification = psOracleScaleFailureClass(id, true)
			attribution = false
		}
		candidate.AllCenteredUnique = candidate.AllCenteredUnique && centered
		if attribution {
			candidate.AllRowsMatch = candidate.AllRowsMatch && rowsMatch
			if !rowsMatch && candidate.FirstFailure == "none" {
				candidate.FirstFailure = fmt.Sprintf("%s:%s", branch.Name, id)
				candidate.firstFailureOrder = len(checks)
				candidate.Classification = psOracleScaleFailureClass(id, false)
			}
		}
		return nil
	}
	for i := range checks {
		if err := check(checks[i].ID, &checks[i]); err != nil {
			return candidate, err
		}
	}
	if replayErr != nil {
		if candidate.FirstFailure == "none" {
			candidate.FirstFailure = fmt.Sprintf("%s:replay", branch.Name)
			candidate.Classification = "ps_oracle_primitive_mismatch"
		}
		return candidate, nil
	}
	if root.ciphertext == nil {
		return candidate, fmt.Errorf("%s PS replay returned nil root", branch.Name)
	}
	if err := check("F0-root", &root); err != nil {
		return candidate, err
	}
	preValues, err := psGlobalDecode(params, root.ciphertext)
	if err != nil {
		return candidate, err
	}
	output := root.ciphertext.CopyNew()
	if err := eval.FastCKKS.Rescale(output, output); err != nil {
		if candidate.FirstFailure == "none" {
			candidate.FirstFailure = fmt.Sprintf("%s:F0-final-rescale", branch.Name)
			candidate.Classification = "ps_oracle_primitive_mismatch"
		}
		return candidate, nil
	}
	postValues, err := psGlobalDecode(params, output)
	if err != nil {
		return candidate, err
	}
	post := PSGlobalCheckpoint{ID: "F0-final-rescale", Operation: "final_rescale", Level: output.Level(), Degree: output.Degree(), Scale: finalizationScaleString(output.Scale), ciphertext: output, SourceBacked: polynomialDecompositionMetric(branch.Base.ReferenceValues, postValues, psOracleScaleBudget), LocalConsistency: polynomialDecompositionMetric(preValues, postValues, psOracleScaleBudget)}
	if err := check(post.ID, &post); err != nil {
		return candidate, err
	}
	candidate.F0PreError = psOracleScalePair(branch.Base.ReferenceValues, preValues, psOracleScaleBudget)
	candidate.F0LocalError = psOracleScalePair(preValues, postValues, psOracleScaleBudget)
	candidate.F0PostError = psOracleScalePair(branch.Base.ReferenceValues, postValues, psOracleScaleBudget)
	candidate.OutputScale = finalizationScaleString(output.Scale)
	candidate.OutputScaleReal, candidate.OutputScaleImag = candidate.OutputScale, candidate.OutputScale
	candidate.FinalVsPlainOracle = psOracleScalePair(branch.Base.ReferenceValues, postValues, psOracleScaleBudget)
	candidate.FinalVsStandard = psOracleScalePair(branch.Standard, postValues, psOracleScaleBudget)
	if output.Level() < 1 {
		candidate.FastVsFullRNS = psOracleScaleRoundoffPair()
	}
	candidate.realFinal = output
	if candidate.FirstFailure == "none" && candidate.AllCenteredUnique && candidate.AllRowsMatch {
		candidate.Classification = "ps_oracle_stage_pass"
	}
	return candidate, nil
}

func psOracleScaleMerge(real, imag psOracleScaleCandidate) psOracleScaleCandidate {
	out := real
	out.OutputScaleImag = imag.OutputScale
	out.F0PreError = psOracleScaleWorstPair(real.F0PreError, imag.F0PreError)
	out.F0LocalError = psOracleScaleWorstPair(real.F0LocalError, imag.F0LocalError)
	out.F0PostError = psOracleScaleWorstPair(real.F0PostError, imag.F0PostError)
	out.FinalVsPlainOracle = psOracleScaleWorstPair(real.FinalVsPlainOracle, imag.FinalVsPlainOracle)
	out.FinalVsStandard = psOracleScaleWorstPair(real.FinalVsStandard, imag.FinalVsStandard)
	out.FastVsFullRNS = psOracleScaleWorstPair(real.FastVsFullRNS, imag.FastVsFullRNS)
	out.WorstCapacityRatio = math.Max(real.WorstCapacityRatio, imag.WorstCapacityRatio)
	if real.WorstCapacityRatio >= imag.WorstCapacityRatio {
		out.WorstCapacityCheckpoint = real.WorstCapacityCheckpoint
	} else {
		out.WorstCapacityCheckpoint = imag.WorstCapacityCheckpoint
	}
	out.AllCenteredUnique = real.AllCenteredUnique && imag.AllCenteredUnique
	out.AllRowsMatch = real.AllRowsMatch && imag.AllRowsMatch
	out.CapacitySafe = out.AllCenteredUnique && out.AllRowsMatch
	out.CheckpointCount = real.CheckpointCount + imag.CheckpointCount
	if real.FirstCumulativeBudgetCross != "none" {
		out.FirstCumulativeBudgetCross = real.FirstCumulativeBudgetCross
	} else {
		out.FirstCumulativeBudgetCross = imag.FirstCumulativeBudgetCross
	}
	if imag.FirstFailure != "none" && (out.FirstFailure == "none" || imag.firstFailureOrder < real.firstFailureOrder) {
		out.FirstFailure, out.Classification = imag.FirstFailure, imag.Classification
	}
	if out.FirstFailure == "none" && out.CapacitySafe {
		out.Classification = "ps_oracle_stage_pass"
	}
	out.PrecisionQualified = out.CapacitySafe && out.FinalVsPlainOracle.Real != nil && out.FinalVsPlainOracle.Imag != nil && out.FinalVsStandard.Real != nil && out.FinalVsStandard.Imag != nil && out.FinalVsPlainOracle.Real.MaxComponent <= psOracleScaleBudget && out.FinalVsPlainOracle.Imag.MaxComponent <= psOracleScaleBudget && out.FinalVsStandard.Real.MaxComponent <= psOracleScaleBudget && out.FinalVsStandard.Imag.MaxComponent <= psOracleScaleBudget && out.FastVsFullRNS.Real.Pass && out.FastVsFullRNS.Imag.Pass
	out.realFinal, out.imagFinal = real.realFinal, imag.realFinal
	return out
}

func psOracleScaleStandardPolynomial(btp bootstrapping.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, params ckks.Parameters, fastInput, standardInput *rlwe.Ciphertext, standardSK *rlwe.SecretKey, inputValues []complex128) ([]complex128, error) {
	e2, _, err := psGlobalE2(btp, fastEval, fastInput)
	if err != nil {
		return nil, err
	}
	standardCurrent := standardInput.CopyNew()
	standardCurrent.Scale = standardEval.Mod1Evaluator.Parameters.ScalingFactor()
	if err := standardEval.Evaluator.Add(standardCurrent, evalModCausalOffset(fastEval), standardCurrent); err != nil {
		return nil, err
	}
	targetScale := psGlobalTargetScale(e2, fastEval)
	standardPoly, err := standardEval.Mod1Evaluator.PolynomialEvaluator.Evaluate(standardCurrent, standardEval.Mod1Evaluator.Parameters.Mod1Poly.Clone(), targetScale)
	if err != nil {
		return nil, err
	}
	_, values, err := semanticBisectView(standardPoly, params, standardSK)
	_ = inputValues
	return values, err
}

func psOracleScaleFullMirror(params ckks.Parameters, fast *rlwe.Ciphertext) ([]complex128, bool, error) {
	if fast == nil || fast.Level() < 1 {
		return nil, true, nil
	}
	full, _, err := postMod1S2CLiftFull(params, fast)
	if err != nil {
		return nil, false, err
	}
	_, values, err := semanticBisectView(full, params, zeroSecret(params))
	return values, postMod1S2CRowsEqual(params, fast, full), err
}

func psOracleScaleDownstream(btp bootstrapping.Parameters, residual ckks.Parameters, eval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, inputs evalModMatchedInputs, candidate *psOracleScaleCandidate, standardSK *rlwe.SecretKey) error {
	if !candidate.PrecisionQualified || candidate.realFinal == nil || candidate.imagFinal == nil {
		candidate.DownstreamReason = "candidate_not_ps_precision_qualified"
		return nil
	}
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastReal, inputs.OrdinaryReal, eval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, precisionSweepScale(candidate.Exponent), candidate.realFinal)
	if err != nil {
		return err
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastImag, inputs.OrdinaryImag, eval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, precisionSweepScale(candidate.Exponent), candidate.imagFinal)
	if err != nil {
		return err
	}
	if realPath.FastFinal == nil || imagPath.FastFinal == nil {
		candidate.DownstreamReason = fmt.Sprintf("normalized_mod1_stopped: real=%s imag=%s", realPath.FirstFailure, imagPath.FirstFailure)
		return nil
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, btp.BootstrappingParameters, inputs.FastSK, standardSK)
	if err != nil {
		return err
	}
	publicLike, standardLike, postS2C, metadata, unchanged, err := precisionSweepPublicLike(realPath, imagPath, eval, standardEval, btp, residual.MaxSlots(), standardSK, reproducibleInput(residual, btp))
	if err != nil {
		return err
	}
	candidate.DownstreamReached = true
	candidate.EvalModVsStandard = psOracleScaleMetricPair{Real: realFinal.FastVsStandard, Imag: imagFinal.FastVsStandard}
	candidate.PostS2CVsStandard = psOracleScaleMetricPair{Real: postS2C, Imag: postS2C}
	candidate.PublicLike, candidate.StandardPublicLike = publicLike, standardLike
	candidate.PublicLikeMetadataCorrect, candidate.PublicLikeInputUnchanged = metadata, unchanged
	return nil
}

func psOracleScaleWrite(result psOracleScaleResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DesignLogN13PSOracleScalePrecision(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
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
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := psOracleScaleResult{
		SchemaVersion: "fix-001-p3-design-logn13-ps-oracle-scale-precision.v1",
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, Budget: psOracleScaleBudget, PublicThreshold: psOracleScalePublicThreshold, SelectedExponent: -1,
		BaselineReproduction: map[string]interface{}{}, Validation: map[string]interface{}{
			"primary_required_base": requiredFIX001P3PSOracleScalePrimaryBase, "secondary_commit": secondaryCommit, "secondary_clean": secondaryClean, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3PSOracleScaleSecondary,
			"no_secondary_production_changes": true, "no_generated_power_redesign": true, "no_mixed_or_per_block_scale": true, "no_c2s_change": true, "no_s2c_change": true, "no_mod1_parameter_retuning": true, "no_extra_q_levels": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_production_integration": true, "no_invalid_raw_standard_da_q01_oracle": true,
		}, Classification: "logn13_ps_oracle_scale_precondition_mismatch", FirstLimiting: "startup_precondition",
	}
	if !secondaryClean || secondaryCommit != requiredFIX001P3PSOracleScaleSecondary {
		return psOracleScaleWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, fastEval, standardEval, standardSK)
	if err != nil {
		return err
	}
	realBase, err := polynomialDecompositionPrepareBranch("real", btp.BootstrappingParameters, btp, fastEval, standardEval, inputs.FastReal, inputs.OrdinaryReal, standardSK)
	if err != nil {
		return err
	}
	imagBase, err := polynomialDecompositionPrepareBranch("imag", btp.BootstrappingParameters, btp, fastEval, standardEval, inputs.FastImag, inputs.OrdinaryImag, standardSK)
	if err != nil {
		return err
	}
	baseline, err := polynomialDecompositionBaseline(cfg, btp, residual, fastEval, standardEval, standardSK, inputs)
	if err != nil {
		return err
	}
	result.BaselineReproduction["public_baseline"] = baseline
	result.BaselineReproduction["real_f_oracle"] = realBase.Summary.FOracle
	result.BaselineReproduction["imag_f_oracle"] = imagBase.Summary.FOracle
	result.BaselineReproduction["real_first_budget_crossing"] = realBase.Summary.OraclePSCheckpoint.FirstCheckpoint
	result.BaselineReproduction["imag_first_budget_crossing"] = imagBase.Summary.OraclePSCheckpoint.FirstCheckpoint
	result.BaselineReproduction["oracle_power_injection_roundoff_real"] = realBase.Summary.ActualPowers
	result.BaselineReproduction["oracle_power_injection_roundoff_imag"] = imagBase.Summary.ActualPowers
	result.BaselineReproduction["oracle_power_rows_match_real"] = allPolynomialDecompositionOracleRowsMatch(realBase.Summary.ActualPowers)
	result.BaselineReproduction["oracle_power_rows_match_imag"] = allPolynomialDecompositionOracleRowsMatch(imagBase.Summary.ActualPowers)
	p0 := realBase.Summary.D1SourceVsStandard != nil && imagBase.Summary.D1SourceVsStandard != nil && realBase.Summary.D1SourceVsStandard.MaxComponent <= psOracleScaleRoundoff && imagBase.Summary.D1SourceVsStandard.MaxComponent <= psOracleScaleRoundoff && realBase.Summary.FOracle != nil && imagBase.Summary.FOracle != nil && realBase.Summary.OraclePSCheckpoint.FirstCheckpoint == "F0-final-rescale" && imagBase.Summary.OraclePSCheckpoint.FirstCheckpoint == "F0-final-rescale"
	result.BaselineReproduction["p0_pass"] = p0
	if !p0 {
		result.Classification, result.FirstLimiting = "logn13_ps_oracle_scale_precondition_mismatch", "P0"
		return psOracleScaleWrite(result, outPath)
	}
	standardReal, err := psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastReal, inputs.OrdinaryReal, standardSK, realBase.InputValues)
	if err != nil {
		return err
	}
	standardImag, err := psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastImag, inputs.OrdinaryImag, standardSK, imagBase.InputValues)
	if err != nil {
		return err
	}
	branches := []psOracleScaleBranch{{Name: "real", Base: realBase, Standard: standardReal, Input: inputs.FastReal, Workload: realBase.InputValues}, {Name: "imag", Base: imagBase, Standard: standardImag, Input: inputs.FastImag, Workload: imagBase.InputValues}}
	for _, exponent := range psOracleScalePlanExponents {
		realCandidate, err := psOracleScaleCandidateForBranch(btp.BootstrappingParameters, fastEval, branches[0], exponent)
		if err != nil {
			return err
		}
		imagCandidate, err := psOracleScaleCandidateForBranch(btp.BootstrappingParameters, fastEval, branches[1], exponent)
		if err != nil {
			return err
		}
		merged := psOracleScaleMerge(realCandidate, imagCandidate)
		if realCandidate.realFinal != nil && realCandidate.realFinal.Level() >= 1 {
			fullValues, rowsMatch, err := psOracleScaleFullMirror(btp.BootstrappingParameters, realCandidate.realFinal)
			if err != nil {
				return err
			}
			merged.FastVsFullRNS.Real = psGlobalMetric(fullValues, mustPolyDecode(btp.BootstrappingParameters, realCandidate.realFinal))
			merged.AllRowsMatch = merged.AllRowsMatch && rowsMatch
		} else if realCandidate.realFinal != nil {
			merged.FastVsFullRNS.Real = psOracleScaleRoundoffPair().Real
		}
		if imagCandidate.realFinal != nil && imagCandidate.realFinal.Level() >= 1 {
			fullValues, rowsMatch, err := psOracleScaleFullMirror(btp.BootstrappingParameters, imagCandidate.realFinal)
			if err != nil {
				return err
			}
			merged.FastVsFullRNS.Imag = psGlobalMetric(fullValues, mustPolyDecode(btp.BootstrappingParameters, imagCandidate.realFinal))
			merged.AllRowsMatch = merged.AllRowsMatch && rowsMatch
		} else if imagCandidate.realFinal != nil {
			merged.FastVsFullRNS.Imag = psOracleScaleRoundoffPair().Imag
		}
		merged.CapacitySafe = merged.AllCenteredUnique && merged.AllRowsMatch
		merged.PrecisionQualified = merged.CapacitySafe && merged.FastVsFullRNS.Real != nil && merged.FastVsFullRNS.Imag != nil && merged.FastVsFullRNS.Real.MaxComponent <= psOracleScaleRoundoff && merged.FastVsFullRNS.Imag.MaxComponent <= psOracleScaleRoundoff && merged.FinalVsStandard.Real != nil && merged.FinalVsStandard.Imag != nil && merged.FinalVsStandard.Real.MaxComponent <= psOracleScaleBudget && merged.FinalVsStandard.Imag.MaxComponent <= psOracleScaleBudget && merged.FinalVsPlainOracle.Real.MaxComponent <= psOracleScaleBudget && merged.FinalVsPlainOracle.Imag.MaxComponent <= psOracleScaleBudget
		if merged.PrecisionQualified {
			if err := psOracleScaleDownstream(btp, residual, fastEval, standardEval, inputs, &merged, standardSK); err != nil {
				return err
			}
		}
		result.Candidates = append(result.Candidates, merged)
	}
	for i := range result.Candidates {
		if i+1 < len(result.Candidates) {
			result.Candidates[i].F0ErrorRatioToNextReal = result.Candidates[i].F0PostError.Real.MaxComponent / result.Candidates[i+1].F0PostError.Real.MaxComponent
			result.Candidates[i].F0ErrorRatioToNextImag = result.Candidates[i].F0PostError.Imag.MaxComponent / result.Candidates[i+1].F0PostError.Imag.MaxComponent
		}
		if result.Candidates[i].PrecisionQualified && result.SelectedExponent < 0 {
			result.SelectedExponent = result.Candidates[i].Exponent
		}
	}
	if result.SelectedExponent >= 0 {
		result.Classification, result.FirstLimiting = "logn13_ps_oracle_common_scale_candidate_validated", "none"
	} else {
		capacitySafe := false
		for _, candidate := range result.Candidates {
			capacitySafe = capacitySafe || candidate.CapacitySafe
		}
		if !capacitySafe {
			result.Classification = "logn13_ps_oracle_common_scale_blocked_by_capacity"
		} else {
			result.Classification = "logn13_ps_oracle_common_scale_insufficient_precision"
		}
		for _, candidate := range result.Candidates {
			if candidate.FirstFailure != "none" {
				result.FirstLimiting = candidate.FirstFailure
				break
			}
		}
		if result.FirstLimiting == "startup_precondition" {
			result.FirstLimiting = "final_polynomial_precision_budget"
		}
	}
	return psOracleScaleWrite(result, outPath)
}
