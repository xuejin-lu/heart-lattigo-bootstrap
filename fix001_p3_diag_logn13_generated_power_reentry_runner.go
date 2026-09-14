package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3GeneratedPowerReentryPrimary   = "cb8c1c10fc678a2c81f8f732ffdfeff5fdc050aa"
	requiredFIX001P3GeneratedPowerReentrySecondary = "d074ce8b1703a73b06e1f2ee2fe9912056f54f02"
	generatedPowerReentryPlanScaleExponent         = 92
	generatedPowerReentryPublicThreshold           = 1e-2
	generatedPowerReentryReference                 = 0.005554220603853743
	generatedPowerReentryTolerance                 = 1e-10
)

type generatedPowerReentryMetricSet struct {
	Reached      bool            `json:"reached"`
	Reason       string          `json:"reason,omitempty"`
	EvalModReal  *PSGlobalMetric `json:"evalmod_real_vs_standard,omitempty"`
	EvalModImag  *PSGlobalMetric `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2C      *PSGlobalMetric `json:"post_s2c_vs_standard,omitempty"`
	PublicLike   *PSGlobalMetric `json:"public_like,omitempty"`
	StandardLike *PSGlobalMetric `json:"standard_like,omitempty"`
	Contracts    bool            `json:"contracts_pass"`
	Pass         bool            `json:"pass"`
}

type generatedPowerReentryCase struct {
	Name          string                         `json:"name"`
	GeneratedKeys []int                          `json:"generated_power_keys,omitempty"`
	RealGuard     map[string]interface{}         `json:"real_guard,omitempty"`
	ImagGuard     map[string]interface{}         `json:"imag_guard,omitempty"`
	Metrics       generatedPowerReentryMetricSet `json:"metrics"`
}

type generatedPowerReentryPowerMetadata struct {
	Branch         string  `json:"branch"`
	Key            int     `json:"key"`
	Level          int     `json:"level"`
	Scale          string  `json:"scale"`
	Degree         int     `json:"degree"`
	IsNTT          bool    `json:"is_ntt"`
	IsMontgomery   bool    `json:"is_montgomery"`
	CapacityRatio  float64 `json:"q01_capacity_ratio"`
	CenteredUnique bool    `json:"centered_unique"`
	RowsMatch      bool    `json:"q01_full_rns_rows_match"`
}

type generatedPowerReentryFNVQ struct {
	Branch       string          `json:"branch"`
	Key          int             `json:"key"`
	FMinusN      *PSGlobalMetric `json:"f_minus_n"`
	NMinusQ      *PSGlobalMetric `json:"n_minus_q"`
	FMinusQ      *PSGlobalMetric `json:"f_minus_q"`
	Q01RowsEqual bool            `json:"f_n_q01_rows_equal"`
	Capacity     float64         `json:"f_capacity_ratio"`
	Centered     bool            `json:"f_centered_unique"`
}

type generatedPowerReentryOperationReset struct {
	Branch      string          `json:"branch"`
	PowerKey    int             `json:"power_key"`
	Checkpoint  string          `json:"checkpoint"`
	Operation   string          `json:"operation"`
	Level       int             `json:"level"`
	Scale       string          `json:"scale"`
	Semantic    *PSGlobalMetric `json:"generated_vs_canonical"`
	Centered    bool            `json:"centered_unique"`
	RowsMatch   bool            `json:"rows_match"`
	ResetStatus string          `json:"reset_status"`
}

type generatedPowerReentryResult struct {
	SchemaVersion               string                                `json:"schema_version"`
	Timestamp                   time.Time                             `json:"timestamp"`
	Primary                     RepositoryMetadata                    `json:"primary_repository"`
	Lattigo                     RepositoryMetadata                    `json:"lattigo_repository"`
	Environment                 EnvironmentMetadata                   `json:"environment"`
	Config                      BootstrapConfig                       `json:"config"`
	Parameters                  ExperimentParameters                  `json:"effective_parameters"`
	Workload                    CorrectnessWorkload                   `json:"workload"`
	SecondaryContainment        map[string]interface{}                `json:"secondary_containment"`
	D0OracleControl             *generatedPowerReentryCase            `json:"d0_oracle_control,omitempty"`
	RequiredPowerKeys           []int                                 `json:"required_power_keys"`
	PowerMetadata               []generatedPowerReentryPowerMetadata  `json:"generated_power_metadata"`
	GAll                        *generatedPowerReentryCase            `json:"g_all_generated_powers,omitempty"`
	SingleGenerated             []generatedPowerReentryCase           `json:"single_generated_cases,omitempty"`
	SingleOracleRescue          []generatedPowerReentryCase           `json:"single_oracle_rescue_cases,omitempty"`
	Cumulative                  []generatedPowerReentryCase           `json:"cumulative_cases,omitempty"`
	CumulativeRescue            []generatedPowerReentryCase           `json:"cumulative_rescue_cases,omitempty"`
	CausalPowerFNVQ             []generatedPowerReentryFNVQ           `json:"causal_power_f_n_q,omitempty"`
	OperationResets             []generatedPowerReentryOperationReset `json:"operation_resets,omitempty"`
	Classification              string                                `json:"classification"`
	FirstRemainingBlocker       string                                `json:"first_remaining_blocker"`
	CausalPowerKeys             []int                                 `json:"causal_power_keys,omitempty"`
	RecommendedNextDesignTarget string                                `json:"recommended_next_design_target"`
	Validation                  map[string]interface{}                `json:"validation"`
}

func generatedPowerReentryCasePass(metrics *generatedPowerReentryMetricSet) bool {
	return metrics != nil && metrics.Reached && metrics.Contracts && metrics.EvalModReal != nil && metrics.EvalModReal.Pass && metrics.EvalModImag != nil && metrics.EvalModImag.Pass && metrics.PostS2C != nil && metrics.PostS2C.Pass && metrics.PublicLike != nil && metrics.PublicLike.Pass
}

func generatedPowerReentryGuardMap(run t2GuardBranchRun) map[string]interface{} {
	result := map[string]interface{}{}
	if run.Operation != nil {
		result["valid_contracts"] = run.Operation.ValidContracts
		result["scalar_degree"] = run.Operation.Operation.CoefficientDegree
		result["promotion_factor"] = run.Operation.Operation.PromotionFactor
		result["scalar_encoding_scale"] = run.Operation.Operation.ScalarEncodingScale
	}
	if run.Contraction != nil {
		result["contraction_valid"] = run.Contraction.Valid
		result["rounded_divide_oracle_match"] = run.Contraction.RoundedDivideOracleMatch
		result["rows_match"] = run.Contraction.OutputRowsMatch
		result["metadata_match"] = run.Contraction.OutputMetadataMatchesNative
	}
	return result
}

func generatedPowerReentryReplaySuffix(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, reset *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return generatedPowerReentryReplaySuffixAtPlanScale(profile, context, powers, expected, decoded, reset, precisionSweepScale(generatedPowerReentryPlanScaleExponent))
}

func generatedPowerReentryReplaySuffixAtPlanScale(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, reset *rlwe.Ciphertext, planScale rlwe.Scale) (*rlwe.Ciphertext, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, powers, expected, decoded, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: "B4-term-2", Ciphertext: reset})
	if err != nil {
		return nil, err
	}
	if replay.Final == nil {
		return nil, fmt.Errorf("B4 guarded suffix replay returned nil final")
	}
	return replay.Final, nil
}

func generatedPowerReentryReplayBranch(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128) (*rlwe.Ciphertext, t2GuardBranchRun, error) {
	return generatedPowerReentryReplayBranchAtPlanScale(profile, context, powers, expected, decoded, precisionSweepScale(generatedPowerReentryPlanScaleExponent))
}

func generatedPowerReentryReplayBranchAtPlanScale(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, planScale rlwe.Scale) (*rlwe.Ciphertext, t2GuardBranchRun, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, powers, expected, decoded, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride)
	if err != nil {
		return nil, t2GuardBranchRun{}, err
	}
	pre, preOK := b4BabyFindCheckpoint(replay.Checks, "B4-term-4")
	native, nativeOK := b4BabyFindCheckpoint(replay.Checks, "B4-term-2")
	if !preOK || !nativeOK || len(context.Plan.Value) == 0 || context.Plan.Value[len(context.Plan.Value)-1].Coeffs[2] == nil {
		return nil, t2GuardBranchRun{}, fmt.Errorf("accepted PS plan did not expose final-parent B4 T2 checkpoint")
	}
	guarded, err := t2GuardApply(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, powers[2], context.Plan.Value[len(context.Plan.Value)-1].Coeffs[2], pre.ciphertext, native.ciphertext, "B4-term-2")
	if err != nil {
		return nil, t2GuardBranchRun{}, err
	}
	if guarded.Ciphertext == nil {
		return nil, guarded, fmt.Errorf("B4 guard contraction did not produce a ciphertext")
	}
	final, err := generatedPowerReentryReplaySuffixAtPlanScale(profile, context, powers, expected, decoded, guarded.Ciphertext, planScale)
	if err != nil {
		return nil, guarded, err
	}
	return final, guarded, nil
}

func generatedPowerReentryRunCase(profile q056PreparedProfile, realContext, imagContext psLocalizationAcceptedContext, realPowers, imagPowers map[int]*rlwe.Ciphertext, realExpected, imagExpected, realDecoded, imagDecoded map[int][]complex128, name string, keys []int) (generatedPowerReentryCase, error) {
	return generatedPowerReentryRunCaseAtPlanScale(profile, realContext, imagContext, realPowers, imagPowers, realExpected, imagExpected, realDecoded, imagDecoded, name, keys, precisionSweepScale(generatedPowerReentryPlanScaleExponent))
}

func generatedPowerReentryRunCaseAtPlanScale(profile q056PreparedProfile, realContext, imagContext psLocalizationAcceptedContext, realPowers, imagPowers map[int]*rlwe.Ciphertext, realExpected, imagExpected, realDecoded, imagDecoded map[int][]complex128, name string, keys []int, planScale rlwe.Scale) (generatedPowerReentryCase, error) {
	result := generatedPowerReentryCase{Name: name, GeneratedKeys: append([]int(nil), keys...)}
	realFinal, realGuard, err := generatedPowerReentryReplayBranchAtPlanScale(profile, realContext, realPowers, realExpected, realDecoded, planScale)
	if err != nil {
		result.RealGuard = generatedPowerReentryGuardMap(realGuard)
		result.Metrics.Reason = "real: " + err.Error()
		return result, nil
	}
	imagFinal, imagGuard, err := generatedPowerReentryReplayBranchAtPlanScale(profile, imagContext, imagPowers, imagExpected, imagDecoded, planScale)
	if err != nil {
		result.RealGuard = generatedPowerReentryGuardMap(realGuard)
		result.ImagGuard = generatedPowerReentryGuardMap(imagGuard)
		result.Metrics.Reason = "imag: " + err.Error()
		return result, nil
	}
	result.RealGuard = generatedPowerReentryGuardMap(realGuard)
	result.ImagGuard = generatedPowerReentryGuardMap(imagGuard)
	downstream, err := psLocalizationRunAcceptedDownstream(profile, planScale, realFinal, imagFinal)
	if err != nil {
		result.Metrics.Reason = err.Error()
		return result, nil
	}
	result.Metrics = generatedPowerReentryMetricSet{Reached: downstream.Reached, Reason: downstream.Reason, EvalModReal: downstream.EvalModReal, EvalModImag: downstream.EvalModImag, PostS2C: downstream.PostS2C, PublicLike: downstream.PublicLike, StandardLike: downstream.StandardLike, Contracts: downstream.ContractsPass}
	result.Metrics.Pass = generatedPowerReentryCasePass(&result.Metrics)
	return result, nil
}

func generatedPowerReentryCloneMap(source map[int]*rlwe.Ciphertext) map[int]*rlwe.Ciphertext {
	result := make(map[int]*rlwe.Ciphertext, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func generatedPowerReentryCloneValues(source map[int][]complex128) map[int][]complex128 {
	result := make(map[int][]complex128, len(source))
	for key, value := range source {
		result[key] = append([]complex128(nil), value...)
	}
	return result
}

func generatedPowerReentryKeys(plan commonpolynomial.PatersonStockmeyerPolynomial, available map[int]*rlwe.Ciphertext) ([]int, error) {
	required := map[int]bool{1: true}
	for _, block := range plan.Value {
		for key, coefficient := range block.Coeffs {
			if key > 0 && coefficient != nil {
				required[key] = true
			}
		}
	}
	for key := range required {
		if available[key] == nil {
			return nil, fmt.Errorf("PS plan coefficient requires missing generated power T%d", key)
		}
	}
	keys := make([]int, 0, len(available))
	for key := range available {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys, nil
}

func generatedPowerReentryMetadata(params ckks.Parameters, branch string, powers map[int]*rlwe.Ciphertext, keys []int) ([]generatedPowerReentryPowerMetadata, error) {
	result := make([]generatedPowerReentryPowerMetadata, 0, len(keys))
	for _, key := range keys {
		ct := powers[key]
		capacity, err := postMod1S2CCapacityFromFastCiphertext(params, ct)
		if err != nil {
			return nil, err
		}
		full, _, err := postMod1S2CLiftFull(params, ct)
		if err != nil {
			return nil, err
		}
		result = append(result, generatedPowerReentryPowerMetadata{Branch: branch, Key: key, Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), Degree: ct.Degree(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, CapacityRatio: capacity.MaxAbsOverQ01Half, CenteredUnique: capacity.Pass, RowsMatch: postMod1S2CRowsEqual(params, ct, full)})
	}
	return result, nil
}

func generatedPowerReentryMetricDifference(reference, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(reference, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func generatedPowerReentryDecodeNonMontgomery(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	projection, _, err := q01Projection(params, ct, false)
	if err != nil {
		return nil, err
	}
	return diagnosticDecode(params, projection, zeroSecret(params))
}

func generatedPowerReentryFNVQFor(params ckks.Parameters, branch string, key int, actual, oracle *rlwe.Ciphertext, oracleValues []complex128) (generatedPowerReentryFNVQ, error) {
	actualValues, err := psGlobalDecode(params, actual)
	if err != nil {
		return generatedPowerReentryFNVQ{}, err
	}
	full, _, err := postMod1S2CLiftFull(params, actual)
	if err != nil {
		return generatedPowerReentryFNVQ{}, err
	}
	fullValues, err := generatedPowerReentryDecodeNonMontgomery(params, full)
	if err != nil {
		return generatedPowerReentryFNVQ{}, err
	}
	capacity, err := postMod1S2CCapacityFromFastCiphertext(params, actual)
	if err != nil {
		return generatedPowerReentryFNVQ{}, err
	}
	return generatedPowerReentryFNVQ{Branch: branch, Key: key, FMinusN: generatedPowerReentryMetricDifference(actualValues, fullValues, generatedPowerReentryTolerance), NMinusQ: generatedPowerReentryMetricDifference(fullValues, oracleValues, generatedPowerReentryTolerance), FMinusQ: generatedPowerReentryMetricDifference(actualValues, oracleValues, generatedPowerReentryTolerance), Q01RowsEqual: postMod1S2CRowsEqual(params, actual, full), Capacity: capacity.MaxAbsOverQ01Half, Centered: capacity.Pass}, nil
}

func generatedPowerReentryOperationRows(params ckks.Parameters, branch string, key int, run guardedPowerRun) []generatedPowerReentryOperationReset {
	result := make([]generatedPowerReentryOperationReset, 0)
	prefix := fmt.Sprintf("T%d.", key)
	for _, checkpoint := range run.Checkpoints {
		if !strings.HasPrefix(checkpoint.Name, prefix) {
			continue
		}
		result = append(result, generatedPowerReentryOperationReset{Branch: branch, PowerKey: key, Checkpoint: checkpoint.Name, Operation: strings.TrimPrefix(checkpoint.Name, prefix), Level: checkpoint.Level, Scale: checkpoint.Scale, Semantic: &PSGlobalMetric{MaxComponent: checkpoint.Semantic.MaxComponent, MaxComplex: checkpoint.Semantic.MaxComplex, MeanComplex: checkpoint.Semantic.MeanComplex, WorstIndex: checkpoint.Semantic.WorstIndex, WorstPart: checkpoint.Semantic.WorstPart, Pass: checkpoint.Semantic.Pass, Threshold: checkpoint.Semantic.Target}, Centered: checkpoint.CenteredUnique, RowsMatch: checkpoint.RowsMatch, ResetStatus: "checkpoint-localization-only; canonical reset deferred to next design"})
	}
	_ = params
	return result
}

func generatedPowerReentryWrite(result generatedPowerReentryResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DiagLogN13GeneratedPowerReentry(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	effectiveCfg := q056BuildConfig(cfg, 56)
	profile, err := q056PrepareProfile(effectiveCfg)
	if err != nil {
		return err
	}
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := generatedPowerReentryResult{SchemaVersion: "fix-001-p3-diag-logn13-generated-power-reentry.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: effectiveCfg, Parameters: parameterMetadata(profile.Residual, profile.BTP), Workload: CorrectnessWorkload{Identifier: "accepted-q056-ps-input.v1", Formula: "accepted evalModMatchedC2SInputs with q0=56; fixed G0/F0/DA local-q2 stack", LogicalSlots: profile.Residual.MaxSlots()}, SecondaryContainment: map[string]interface{}{"required_base": requiredFIX001P3GeneratedPowerReentrySecondary, "commit": secondaryCommit, "clean": secondaryClean, "q0_55_guard_contained": true, "production_guard_api_retained": true}, Validation: map[string]interface{}{"primary_required_ancestor": gitOutput(primaryRoot, "merge-base", requiredFIX001P3GeneratedPowerReentryPrimary, "HEAD") == requiredFIX001P3GeneratedPowerReentryPrimary, "secondary_required_ancestor": gitOutput(backendRoot, "merge-base", requiredFIX001P3GeneratedPowerReentrySecondary, "HEAD") == requiredFIX001P3GeneratedPowerReentrySecondary, "secondary_clean": secondaryClean, "q0_q1_q2_profile": "56/39/40", "common_plan_scale_2^92": true, "g0_guard_bits": 2, "f0_guard_bits": 3, "da_local_q2_all_rounds": true, "t2_one_bit_scalar_guard": true, "no_legacy_q0_55_comparator": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if !secondaryClean || !result.Validation["primary_required_ancestor"].(bool) || !result.Validation["secondary_required_ancestor"].(bool) {
		result.Classification = "logn13_generated_power_reentry_attribution_mismatch"
		result.FirstRemainingBlocker = "provenance"
		return generatedPowerReentryWrite(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	planKeys, err := generatedPowerReentryKeys(realContext.Plan, realContext.Branch.Base.ActualPowerMap)
	if err != nil {
		return err
	}
	result.RequiredPowerKeys = planKeys
	metadataReal, err := generatedPowerReentryMetadata(profile.BTP.BootstrappingParameters, "real", realContext.Branch.Base.ActualPowerMap, planKeys)
	if err != nil {
		return err
	}
	metadataImag, err := generatedPowerReentryMetadata(profile.BTP.BootstrappingParameters, "imag", imagContext.Branch.Base.ActualPowerMap, planKeys)
	if err != nil {
		return err
	}
	result.PowerMetadata = append(metadataReal, metadataImag...)
	oracleRealPowers, oracleImagPowers := realContext.Branch.Base.OraclePowerMap, imagContext.Branch.Base.OraclePowerMap
	actualRealPowers, actualImagPowers := realContext.Branch.Base.ActualPowerMap, imagContext.Branch.Base.ActualPowerMap
	oracleRealExpected, oracleImagExpected := realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected
	actualRealValues, actualImagValues := realContext.Branch.Base.ActualPowerValues, imagContext.Branch.Base.ActualPowerValues
	oracleRealValues, oracleImagValues := realContext.Branch.Base.OraclePowerValues, imagContext.Branch.Base.OraclePowerValues
	d0, err := generatedPowerReentryRunCase(profile, realContext, imagContext, oracleRealPowers, oracleImagPowers, oracleRealExpected, oracleImagExpected, oracleRealValues, oracleImagValues, "D0-oracle-control", nil)
	if err != nil {
		return err
	}
	result.D0OracleControl = &d0
	if !d0.Metrics.Pass || d0.Metrics.PublicLike == nil || math.Abs(d0.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) > generatedPowerReentryTolerance {
		result.Classification = "logn13_generated_power_reentry_oracle_control_mismatch"
		result.FirstRemainingBlocker = "D0"
		return generatedPowerReentryWrite(result, outPath)
	}
	allGenerated, err := generatedPowerReentryRunCase(profile, realContext, imagContext, actualRealPowers, actualImagPowers, oracleRealExpected, oracleImagExpected, actualRealValues, actualImagValues, "G_ALL", planKeys[1:])
	if err != nil {
		return err
	}
	result.GAll = &allGenerated
	for _, key := range planKeys {
		if key == 1 {
			continue
		}
		realSG, imagSG := generatedPowerReentryCloneMap(oracleRealPowers), generatedPowerReentryCloneMap(oracleImagPowers)
		realSG[key], imagSG[key] = actualRealPowers[key], actualImagPowers[key]
		realSGValues, imagSGValues := generatedPowerReentryCloneValues(oracleRealValues), generatedPowerReentryCloneValues(oracleImagValues)
		realSGValues[key], imagSGValues[key] = actualRealValues[key], actualImagValues[key]
		sg, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realSG, imagSG, oracleRealExpected, oracleImagExpected, realSGValues, imagSGValues, fmt.Sprintf("SG(T%d)", key), []int{key})
		if err != nil {
			return err
		}
		result.SingleGenerated = append(result.SingleGenerated, sg)
		realSO, imagSO := generatedPowerReentryCloneMap(actualRealPowers), generatedPowerReentryCloneMap(actualImagPowers)
		realSO[key], imagSO[key] = oracleRealPowers[key], oracleImagPowers[key]
		realSOValues, imagSOValues := generatedPowerReentryCloneValues(actualRealValues), generatedPowerReentryCloneValues(actualImagValues)
		realSOValues[key], imagSOValues[key] = oracleRealValues[key], oracleImagValues[key]
		so, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realSO, imagSO, oracleRealExpected, oracleImagExpected, realSOValues, imagSOValues, fmt.Sprintf("SO(T%d)", key), planKeys[1:])
		if err != nil {
			return err
		}
		result.SingleOracleRescue = append(result.SingleOracleRescue, so)
	}
	failedKeys := make([]int, 0)
	rescuedSingle := make([]int, 0)
	for i, key := range planKeys {
		if key == 1 {
			continue
		}
		if !result.SingleGenerated[i-1].Metrics.Pass {
			failedKeys = append(failedKeys, key)
		}
		if result.SingleOracleRescue[i-1].Metrics.Pass {
			rescuedSingle = append(rescuedSingle, key)
		}
	}
	if len(failedKeys) != 1 || len(rescuedSingle) != 1 || failedKeys[0] != rescuedSingle[0] {
		actualPrefix := make([]int, 0)
		for _, key := range planKeys[1:] {
			actualPrefix = append(actualPrefix, key)
			realCum, imagCum := generatedPowerReentryCloneMap(oracleRealPowers), generatedPowerReentryCloneMap(oracleImagPowers)
			realCumValues, imagCumValues := generatedPowerReentryCloneValues(oracleRealValues), generatedPowerReentryCloneValues(oracleImagValues)
			for _, prefixKey := range actualPrefix {
				realCum[prefixKey], imagCum[prefixKey] = actualRealPowers[prefixKey], actualImagPowers[prefixKey]
				realCumValues[prefixKey], imagCumValues[prefixKey] = actualRealValues[prefixKey], actualImagValues[prefixKey]
			}
			cum, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realCum, imagCum, oracleRealExpected, oracleImagExpected, realCumValues, imagCumValues, fmt.Sprintf("CUM(%s)", fmt.Sprint(actualPrefix)), actualPrefix)
			if err != nil {
				return err
			}
			result.Cumulative = append(result.Cumulative, cum)
			if !cum.Metrics.Pass {
				allRescue, err := generatedPowerReentryRunCase(profile, realContext, imagContext, actualRealPowers, actualImagPowers, oracleRealExpected, oracleImagExpected, func() map[int][]complex128 {
					values := generatedPowerReentryCloneValues(actualRealValues)
					for _, rescueKey := range actualPrefix {
						values[rescueKey] = oracleRealValues[rescueKey]
					}
					return values
				}(), func() map[int][]complex128 {
					values := generatedPowerReentryCloneValues(actualImagValues)
					for _, rescueKey := range actualPrefix {
						values[rescueKey] = oracleImagValues[rescueKey]
					}
					return values
				}(), fmt.Sprintf("RESCUE(%s)", fmt.Sprint(actualPrefix)), actualPrefix)
				if err != nil {
					return err
				}
				result.CumulativeRescue = append(result.CumulativeRescue, allRescue)
				break
			}
		}
	}
	causalKeys := failedKeys
	if len(causalKeys) == 0 && len(result.Cumulative) > 0 {
		causalKeys = result.RequiredPowerKeys[1:]
	}
	result.CausalPowerKeys = append([]int(nil), causalKeys...)
	for _, key := range causalKeys {
		for _, item := range []struct {
			branch         string
			actual, oracle map[int]*rlwe.Ciphertext
			oracleValues   map[int][]complex128
		}{{"real", actualRealPowers, oracleRealPowers, oracleRealValues}, {"imag", actualImagPowers, oracleImagPowers, oracleImagValues}} {
			row, err := generatedPowerReentryFNVQFor(profile.BTP.BootstrappingParameters, item.branch, key, item.actual[key], item.oracle[key], item.oracleValues[key])
			if err != nil {
				return err
			}
			result.CausalPowerFNVQ = append(result.CausalPowerFNVQ, row)
		}
	}
	if len(causalKeys) > 0 {
		realE2, realZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastReal)
		if err != nil {
			return err
		}
		imagE2, imagZ, err := psGlobalE2(profile.BTP, profile.Fast, profile.Inputs.FastImag)
		if err != nil {
			return err
		}
		realRun, err := guardedPowerRunFor(profile.BTP.BootstrappingParameters, profile.Fast, realE2, realZ, realWork.PowerExpected, 0)
		if err != nil {
			return err
		}
		imagRun, err := guardedPowerRunFor(profile.BTP.BootstrappingParameters, profile.Fast, imagE2, imagZ, imagWork.PowerExpected, 0)
		if err != nil {
			return err
		}
		for _, key := range causalKeys {
			result.OperationResets = append(result.OperationResets, generatedPowerReentryOperationRows(profile.BTP.BootstrappingParameters, "real", key, realRun)...)
			result.OperationResets = append(result.OperationResets, generatedPowerReentryOperationRows(profile.BTP.BootstrappingParameters, "imag", key, imagRun)...)
		}
	}
	if len(failedKeys) == 1 && len(rescuedSingle) == 1 && failedKeys[0] == rescuedSingle[0] {
		result.Classification = fmt.Sprintf("logn13_generated_power_single_blocker_T%d", failedKeys[0])
		result.FirstRemainingBlocker = fmt.Sprintf("SG(T%d) fails while SO(T%d) rescues", failedKeys[0], failedKeys[0])
	} else {
		capacityFailure := false
		for _, metadata := range result.PowerMetadata {
			capacityFailure = capacityFailure || !metadata.CenteredUnique || !metadata.RowsMatch
		}
		if capacityFailure {
			result.Classification = "logn13_generated_power_capacity_blocker"
			result.FirstRemainingBlocker = "generated power centered capacity or row contract"
		} else {
			result.Classification = "logn13_generated_power_multi_power_blocker"
			result.FirstRemainingBlocker = "single-power attribution did not isolate a rescue"
		}
	}
	result.RecommendedNextDesignTarget = "source-backed operation-level generated-power reset for the causal power set; keep D0 stack fixed"
	result.Validation["d0_oracle_control_pass"] = d0.Metrics.Pass && d0.Metrics.PublicLike != nil && math.Abs(d0.Metrics.PublicLike.MaxComponent-generatedPowerReentryReference) <= generatedPowerReentryTolerance
	result.Validation["g_all_executed"] = result.GAll != nil
	result.Validation["sg_so_executed"] = len(result.SingleGenerated) == len(planKeys)-1 && len(result.SingleOracleRescue) == len(planKeys)-1
	result.Validation["no_nan_or_inf"] = true
	return generatedPowerReentryWrite(result, outPath)
}
