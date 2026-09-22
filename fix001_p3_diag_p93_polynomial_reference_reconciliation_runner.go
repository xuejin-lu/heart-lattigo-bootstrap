package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	fix001P3P93PolynomialReconciliationSchema = "fix-001-p3-diag-p93-polynomial-reference-reconciliation-and-ps-regression-localization.v1"
	fix001P3P93PolynomialReferencePrimary     = "41993cd03f85ec5f2de54dc21ade5ec812090d50"
	fix001P3P93PolynomialReferenceSecondary   = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3P93PolynomialProductionSecondary  = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3P93PolynomialProductionDiffSHA    = "4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764"
	fix001P3P93PolynomialBudget               = 3.716228023823462e-8
	fix001P3P93DoubleAnglePreRestoreBudget    = 3.0517578125e-7
	fix001P3P93PolynomialOracleThreshold      = 1e-10
	fix001P3P93PolynomialReferenceDir         = "/tmp/fix001-p93-poly-reconcile.wrfCrs"
)

type fix001P3P93PolynomialReferenceResult struct {
	SchemaVersion  string                   `json:"schema_version"`
	Timestamp      time.Time                `json:"timestamp"`
	Primary        RepositoryMetadata       `json:"primary_repository"`
	Production     RepositoryMetadata       `json:"production_secondary"`
	Environment    EnvironmentMetadata      `json:"environment"`
	Config         BootstrapConfig          `json:"config"`
	FixedProfile   map[string]interface{}   `json:"fixed_profile"`
	Provenance     map[string]interface{}   `json:"provenance"`
	R0Standard     map[string]interface{}   `json:"r0_genuine_standard_polynomial_oracle"`
	R1Historical   map[string]interface{}   `json:"r1_fresh_historical_p93_replay"`
	R2Reconcile    map[string]interface{}   `json:"r2_historical_standard_reconciliation"`
	R3Budget       map[string]interface{}   `json:"r3_historical_design_budget"`
	R4Regression   map[string]interface{}   `json:"r4_current_production_regression"`
	R5Checkpoints  []map[string]interface{} `json:"r5_ps_checkpoint_localization"`
	R6Localization map[string]interface{}   `json:"r6_generated_power_vs_ps_accumulation"`
	R7Projection   map[string]interface{}   `json:"r7_double_angle_target_implication"`
	Classification string                   `json:"classification"`
	Validation     map[string]interface{}   `json:"validation"`
}

type fix001P3P93ReferenceSummary struct {
	Primary    RepositoryMetadata              `json:"primary_repository"`
	Lattigo    RepositoryMetadata              `json:"lattigo_repository"`
	Validation map[string]interface{}          `json:"validation"`
	Candidates []fix001P3P93ReferenceCandidate `json:"candidates"`
}

type fix001P3P93ReferenceCandidate struct {
	PlanScaleExponent int                        `json:"plan_scale_exponent"`
	Real              fix001P3P93ReferenceBranch `json:"real"`
	Imag              fix001P3P93ReferenceBranch `json:"imag"`
}

type fix001P3P93ReferenceBranch struct {
	GeneratedPowers []fix001P3P93ReferencePower      `json:"generated_powers"`
	PSCheckpoints   []fix001P3P93ReferenceCheckpoint `json:"ps_checkpoints"`
	PSPolynomial    *PSGlobalMetric                  `json:"ps_polynomial_vs_standard"`
	EvalMod         *PSGlobalMetric                  `json:"evalmod_vs_standard"`
	PostS2C         *PSGlobalMetric                  `json:"post_s2c_vs_standard"`
	PublicLike      *PSGlobalMetric                  `json:"public_like"`
}

type fix001P3P93ReferencePower struct {
	Power       int                              `json:"power"`
	Checkpoints []fix001P3P93ReferenceCheckpoint `json:"checkpoints"`
}

type fix001P3P93ReferenceCheckpoint struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	Level     int                    `json:"level"`
	Scale     string                 `json:"scale"`
	Degree    int                    `json:"degree"`
	Capacity  map[string]interface{} `json:"capacity"`
}

type fix001P3P93ReferencePolynomialRow struct {
	Branch   string                 `json:"branch"`
	Exponent int                    `json:"plan_scale_exponent"`
	Level    int                    `json:"level"`
	Scale    string                 `json:"scale"`
	Degree   int                    `json:"degree"`
	Fast     []fix001P3TraceComplex `json:"fast"`
	Standard []fix001P3TraceComplex `json:"standard"`
}

type fix001P3P93ReferencePSRow struct {
	Branch    string                 `json:"branch"`
	Exponent  int                    `json:"plan_scale_exponent"`
	Name      string                 `json:"name"`
	Operation string                 `json:"operation"`
	Level     int                    `json:"level"`
	Scale     string                 `json:"scale"`
	Degree    int                    `json:"degree"`
	Values    []fix001P3TraceComplex `json:"values"`
}

func fix001P3P93PolynomialWrite(result fix001P3P93PolynomialReferenceResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"primary_repository", result.Primary}, {"production_secondary", result.Production},
		{"environment", result.Environment}, {"config", result.Config}, {"fixed_profile", result.FixedProfile}, {"provenance", result.Provenance},
		{"r0_genuine_standard_polynomial_oracle", result.R0Standard}, {"r1_fresh_historical_p93_replay", result.R1Historical}, {"r2_historical_standard_reconciliation", result.R2Reconcile},
		{"r3_historical_design_budget", result.R3Budget}, {"r4_current_production_regression", result.R4Regression}, {"r5_ps_checkpoint_localization", result.R5Checkpoints},
		{"r6_generated_power_vs_ps_accumulation", result.R6Localization}, {"r7_double_angle_target_implication", result.R7Projection}, {"classification", result.Classification}, {"validation", result.Validation},
	}
	var builder strings.Builder
	builder.WriteString("{\n")
	for i, field := range fields {
		name, _ := json.Marshal(field.name)
		value, err := json.Marshal(field.value)
		if err != nil {
			return err
		}
		builder.WriteString("  ")
		builder.Write(name)
		builder.WriteString(": ")
		builder.Write(value)
		if i+1 < len(fields) {
			builder.WriteByte(',')
		}
		builder.WriteByte('\n')
	}
	builder.WriteString("}\n")
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(builder.String()), 0o644)
}

func fix001P3P93PolynomialMetric(reference, actual []complex128, threshold float64) (*semanticBisectMetric, error) {
	metric, err := semanticBisectMetricFor(reference, actual, threshold)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func fix001P3P93PolynomialTraceValues(values []fix001P3TraceComplex) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func fix001P3P93PolynomialReadJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func fix001P3P93PolynomialCapacitySummary(capacity interface{}) map[string]interface{} {
	summary := map[string]interface{}{"available": false}
	switch value := capacity.(type) {
	case *postMod1S2CCapacity:
		if value != nil {
			summary["available"] = true
			summary["max_abs_over_q01_half"] = value.MaxAbsOverQ01Half
			summary["pass_unique_centered_reconstruction"] = value.Pass
			summary["outside_count"] = value.OutsideCount
		}
	case map[string]interface{}:
		for _, key := range []string{"q012_max_abs_over_half", "q01_max_abs_over_half", "q012_centered_unique", "q01_centered_unique", "minimum_remaining_q012_bits", "outside_q012_count", "outside_q01_count"} {
			if value[key] != nil {
				summary[key] = value[key]
			}
		}
		if len(summary) > 1 {
			summary["available"] = true
		}
	}
	return summary
}

func fix001P3P93PolynomialHistoricalCapacity(candidate fix001P3P93ReferenceCandidate, branch, checkpoint string) map[string]interface{} {
	var source fix001P3P93ReferenceBranch
	if branch == "real" {
		source = candidate.Real
	} else {
		source = candidate.Imag
	}
	for _, power := range source.GeneratedPowers {
		for _, row := range power.Checkpoints {
			if row.ID == checkpoint {
				return fix001P3P93PolynomialCapacitySummary(row.Capacity)
			}
		}
	}
	for _, row := range source.PSCheckpoints {
		if row.ID == checkpoint {
			return fix001P3P93PolynomialCapacitySummary(row.Capacity)
		}
	}
	return map[string]interface{}{"available": false}
}

func fix001P3P93PolynomialEventMap(events []fix001P3TraceEvent, branch string) map[string]fix001P3TraceEvent {
	result := map[string]fix001P3TraceEvent{}
	for _, event := range events {
		if event.Kind == "ps" && event.Branch == branch {
			result[event.Name] = event
		}
	}
	return result
}

func fix001P3P93PolynomialReferenceMap(rows []fix001P3P93ReferencePSRow, branch string) map[string]fix001P3P93ReferencePSRow {
	result := map[string]fix001P3P93ReferencePSRow{}
	for _, row := range rows {
		if row.Branch == branch && row.Exponent == 93 {
			result[row.Name] = row
		}
	}
	return result
}

func fix001P3P93PolynomialCheckpointRank(name string) string {
	if strings.HasPrefix(name, "T") && strings.HasSuffix(name, "-final") {
		power := strings.TrimSuffix(strings.TrimPrefix(name, "T"), "-final")
		return fmt.Sprintf("0-%03s", power)
	}
	if strings.HasPrefix(name, "B") {
		parts := strings.Split(strings.TrimPrefix(name, "B"), "-")
		if len(parts) > 0 {
			block, _ := strconv.Atoi(parts[0])
			return fmt.Sprintf("1-%03d-%s", block, name)
		}
	}
	if strings.HasPrefix(name, "G") {
		parts := strings.Split(strings.TrimPrefix(name, "G"), "-")
		if len(parts) > 0 {
			round, _ := strconv.Atoi(parts[0])
			return fmt.Sprintf("2-%03d-%s", round, name)
		}
	}
	if strings.HasPrefix(name, "F0-") {
		return "3-" + name
	}
	return "9-" + name
}

func fix001P3P93PolynomialSortedCommonNames(reference map[string]fix001P3P93ReferencePSRow, productionEvents map[string]fix001P3TraceEvent) []string {
	names := []string{}
	for name := range reference {
		if _, ok := productionEvents[name]; ok {
			names = append(names, name)
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return fix001P3P93PolynomialCheckpointRank(names[i]) < fix001P3P93PolynomialCheckpointRank(names[j])
	})
	return names
}

func fix001P3P93PolynomialZeroMetric(values []complex128) (*semanticBisectMetric, error) {
	return fix001P3P93PolynomialMetric(make([]complex128, len(values)), values, 1e-10)
}

func fix001P3P93PolynomialRunBranch(profile q056PreparedProfile, branch string) (fix001P3P93InternalBranch, fix001P3P93InternalBranch, []fix001P3TraceEvent, error) {
	fastInput, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
	if branch == "imag" {
		fastInput, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
	}
	standard, _, _, err := fix001P3P93InternalStandardTrace(standardInput, profile.Standard, profile.BTP.BootstrappingParameters, profile.StandardSK, profile.Inputs.FastSK)
	if err != nil {
		return standard, fix001P3P93InternalBranch{}, nil, err
	}
	fast, _, _, err := fix001P3P93InternalFastTrace(fastInput, standardInput, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.StandardSK, profile.Inputs.FastSK)
	if err != nil {
		return standard, fast, nil, err
	}
	fix001P3TraceBranch = branch
	events, err := fix001P3P93ProductionPSTrace(profile.BTP.BootstrappingParameters, profile.Fast, fastInput, fix001P3P93InternalPlanScale())
	fix001P3TraceBranch = ""
	return standard, fast, events, err
}

func fix001P3P93PolynomialMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	if ct == nil {
		return map[string]interface{}{"available": false}
	}
	metadata := fix001P3PlaintextCompactMetadata(ct)
	metadata["row_hashes"] = fix001P3ActualForcedRowHashes(ct)
	return metadata
}

func runFIX001P3DiagP93PolynomialReferenceReconciliation(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3P93PolynomialReferenceResult{
		SchemaVersion: fix001P3P93PolynomialReconciliationSchema,
		Timestamp:     time.Now().UTC(), Primary: gitMetadata(primaryRoot), Production: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits": "~39", "q2_bits": "~40", "ps_authority": "Q012-wide", "plan_scale": "2^93", "polynomial_budget": fix001P3P93PolynomialBudget, "double_angle_pre_restore_budget": fix001P3P93DoubleAnglePreRestoreBudget, "oracle_threshold": fix001P3P93PolynomialOracleThreshold, "double_angle_directional_gain": 8.211976748833033},
		Provenance:   map[string]interface{}{"production_branch": secondaryBranch, "production_committed_head": secondaryCommit, "production_dirty": secondaryDirty, "production_diff_stat": diffStat, "production_diff_sha256": diffHash, "reference_primary_commit": fix001P3P93PolynomialReferencePrimary, "reference_secondary_commit": fix001P3P93PolynomialReferenceSecondary, "reference_worktree_method": "fresh detached temporary worktrees; temporary harness-only P93 scale restriction and vector/PS trace instrumentation", "reference_worktree_root": os.Getenv("FIX001_P93_REFERENCE_DIR")},
		Validation:   map[string]interface{}{"secondary_required_head": fix001P3P93PolynomialProductionSecondary, "secondary_required_diff_sha256": fix001P3P93PolynomialProductionDiffSHA, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_parameter_tuning": true, "no_generated_power_replacement": true, "no_ps_repair": true, "no_double_angle_repair": true, "no_c2s_s2c_finalizer_work": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if result.Provenance["reference_worktree_root"] == "" {
		result.Provenance["reference_worktree_root"] = fix001P3P93PolynomialReferenceDir
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3P93PolynomialProductionSecondary || diffHash != fix001P3P93PolynomialProductionDiffSHA || !secondaryDirty {
		result.Classification = "P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT"
		result.Validation["production_handoff_state_match"] = false
		return fix001P3P93PolynomialWrite(result, outPath)
	}

	refDir := result.Provenance["reference_worktree_root"].(string)
	var referenceSummary fix001P3P93ReferenceSummary
	var referencePolynomial []fix001P3P93ReferencePolynomialRow
	var referencePS []fix001P3P93ReferencePSRow
	if err := fix001P3P93PolynomialReadJSON(refDir+"/p93-reference-summary.json", &referenceSummary); err != nil {
		return err
	}
	if err := fix001P3P93PolynomialReadJSON(refDir+"/p93-polynomial-trace.json", &referencePolynomial); err != nil {
		return err
	}
	if err := fix001P3P93PolynomialReadJSON(refDir+"/p93-ps-trace.json", &referencePS); err != nil {
		return err
	}
	if len(referenceSummary.Candidates) != 1 || referenceSummary.Candidates[0].PlanScaleExponent != 93 || referenceSummary.Primary.Commit != fix001P3P93PolynomialReferencePrimary || referenceSummary.Lattigo.Commit != fix001P3P93PolynomialReferenceSecondary || referenceSummary.Lattigo.Dirty || referenceSummary.Validation["secondary_clean"] != true {
		result.Classification = "P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT"
		return fix001P3P93PolynomialWrite(result, outPath)
	}
	candidate := referenceSummary.Candidates[0]
	result.R1Historical = map[string]interface{}{"primary_commit": referenceSummary.Primary.Commit, "secondary_commit": referenceSummary.Lattigo.Commit, "secondary_clean": true, "plan_scale_exponent": candidate.PlanScaleExponent, "real": map[string]interface{}{"ps_polynomial_vs_historical_standard": candidate.Real.PSPolynomial, "evalmod_vs_historical_standard": candidate.Real.EvalMod, "post_s2c": candidate.Real.PostS2C, "public_like": candidate.Real.PublicLike}, "imag": map[string]interface{}{"ps_polynomial_vs_historical_standard": candidate.Imag.PSPolynomial, "evalmod_vs_historical_standard": candidate.Imag.EvalMod, "post_s2c": candidate.Imag.PostS2C, "public_like": candidate.Imag.PublicLike}, "trace_slots": 4096, "fresh_replay": true}
	if candidate.Real.PublicLike == nil || candidate.Imag.PublicLike == nil || candidate.Real.PublicLike.MaxComponent >= 1e-2 || candidate.Imag.PublicLike.MaxComponent >= 1e-2 {
		result.Classification = "P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT"
		return fix001P3P93PolynomialWrite(result, outPath)
	}

	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	branches := map[string]struct {
		standard fix001P3P93InternalBranch
		fast     fix001P3P93InternalBranch
		events   []fix001P3TraceEvent
	}{}
	for _, branch := range []string{"real", "imag"} {
		standard, fast, events, runErr := fix001P3P93PolynomialRunBranch(profile, branch)
		if runErr != nil {
			return runErr
		}
		branches[branch] = struct {
			standard fix001P3P93InternalBranch
			fast     fix001P3P93InternalBranch
			events   []fix001P3TraceEvent
		}{standard: standard, fast: fast, events: events}
	}

	standardAlignment := map[string]interface{}{}
	standardPolynomials := map[string][]complex128{}
	productionPolynomials := map[string][]complex128{}
	referencePolynomials := map[string]struct {
		fast, standard []complex128
	}{}
	for _, branch := range []string{"real", "imag"} {
		state := branches[branch]
		standardActualInput := profile.Inputs.OrdinaryReal
		if branch == "imag" {
			standardActualInput = profile.Inputs.OrdinaryImag
		}
		actual, actualErr := profile.Standard.Mod1Evaluator.EvaluateNew(standardActualInput.CopyNew())
		if actualErr != nil {
			return actualErr
		}
		_, actualValues, actualErr := semanticBisectView(actual, profile.BTP.BootstrappingParameters, profile.StandardSK)
		if actualErr != nil {
			return actualErr
		}
		_, replayValues, replayErr := semanticBisectView(state.standard.states["internal_final_restore"], profile.BTP.BootstrappingParameters, profile.StandardSK)
		if replayErr != nil {
			return replayErr
		}
		alignmentMetric, alignmentErr := fix001P3P93PolynomialMetric(actualValues, replayValues, fix001P3P93PolynomialOracleThreshold)
		if alignmentErr != nil {
			return alignmentErr
		}
		standardAlignment[branch] = map[string]interface{}{"actual_evalmod_internal": fix001P3P93PolynomialMetadata(actual), "replay_internal_final": fix001P3P93PolynomialMetadata(state.standard.states["internal_final_restore"]), "metric": alignmentMetric, "pass": alignmentMetric.MaxComponentAbs <= fix001P3P93PolynomialOracleThreshold, "lineage": "same deterministic profile OrdinaryReal/OrdinaryImag input through profile.Standard.Mod1Evaluator.EvaluateNew"}
		_, standardValues, err := semanticBisectView(state.standard.states["polynomial_output_after_rescale"], profile.BTP.BootstrappingParameters, profile.StandardSK)
		if err != nil {
			return err
		}
		_, productionValues, err := semanticBisectView(state.fast.states["polynomial_output_after_rescale"], profile.BTP.BootstrappingParameters, profile.Inputs.FastSK)
		if err != nil {
			return err
		}
		standardPolynomials[branch], productionPolynomials[branch] = standardValues, productionValues
		for _, row := range referencePolynomial {
			if row.Branch == branch && row.Exponent == 93 {
				referencePolynomials[branch] = struct{ fast, standard []complex128 }{fast: fix001P3P93PolynomialTraceValues(row.Fast), standard: fix001P3P93PolynomialTraceValues(row.Standard)}
			}
		}
	}
	result.R0Standard = map[string]interface{}{"lineage": "Genuine Standard Mod1Evaluator.EvaluateNew on the deterministic matched C2S inputs; replay uses the same input and source-faithful Standard polynomial/DoubleAngle operations", "branches": standardAlignment, "oracle_alignment_threshold": fix001P3P93PolynomialOracleThreshold}
	allOracleAligned := true
	for _, branch := range []string{"real", "imag"} {
		if !standardAlignment[branch].(map[string]interface{})["pass"].(bool) {
			allOracleAligned = false
		}
	}
	if !allOracleAligned {
		result.Classification = "P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT"
		return fix001P3P93PolynomialWrite(result, outPath)
	}

	polyRows := map[string]interface{}{}
	maxCurrent, maxHistorical := 0.0, 0.0
	referenceComparable, closurePass := true, true
	for _, branch := range []string{"real", "imag"} {
		ref := referencePolynomials[branch]
		if len(ref.fast) == 0 || len(ref.standard) == 0 {
			result.Classification = "P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT"
			return fix001P3P93PolynomialWrite(result, outPath)
		}
		currentVsStandard, err := fix001P3P93PolynomialMetric(standardPolynomials[branch], productionPolynomials[branch], fix001P3P93PolynomialBudget)
		if err != nil {
			return err
		}
		historicalVsStandard, err := fix001P3P93PolynomialMetric(standardPolynomials[branch], ref.fast, fix001P3P93PolynomialBudget)
		if err != nil {
			return err
		}
		productionVsHistorical, err := fix001P3P93PolynomialMetric(ref.fast, productionPolynomials[branch], fix001P3P93PolynomialBudget)
		if err != nil {
			return err
		}
		referenceReconciliation, err := fix001P3P93PolynomialMetric(ref.standard, standardPolynomials[branch], fix001P3P93PolynomialOracleThreshold)
		if err != nil {
			return err
		}
		regression := fix001P3P93PlainVectorDifference(productionPolynomials[branch], ref.fast)
		regressionReference := fix001P3P93PlainVectorDifference(ref.fast, standardPolynomials[branch])
		regressionTotal := fix001P3P93PlainVectorDifference(productionPolynomials[branch], standardPolynomials[branch])
		closure := fix001P3P93PlainVectorDifference(fix001P3P93PlainVectorAdd(regression, regressionReference), regressionTotal)
		closureMetric, err := fix001P3P93PolynomialZeroMetric(closure)
		if err != nil {
			return err
		}
		if currentVsStandard.MaxComponentAbs > maxCurrent {
			maxCurrent = currentVsStandard.MaxComponentAbs
		}
		if historicalVsStandard.MaxComponentAbs > maxHistorical {
			maxHistorical = historicalVsStandard.MaxComponentAbs
		}
		if referenceReconciliation.MaxComponentAbs > fix001P3P93PolynomialOracleThreshold {
			referenceComparable = false
		}
		if closureMetric.MaxComponentAbs > 1e-10 {
			closurePass = false
		}
		polyRows[branch] = map[string]interface{}{"genuine_standard_metadata": fix001P3P93PolynomialMetadata(branches[branch].standard.states["polynomial_output_after_rescale"]), "historical_fast_metadata": map[string]interface{}{"level": referencePolynomialLevel(referencePolynomial, branch), "scale": referencePolynomialScale(referencePolynomial, branch), "degree": referencePolynomialDegree(referencePolynomial, branch)}, "current_fast_vs_genuine_standard": currentVsStandard, "historical_fast_vs_genuine_standard": historicalVsStandard, "current_fast_vs_historical_fast": productionVsHistorical, "historical_standard_vs_genuine_standard": referenceReconciliation, "regression_vector_closure": closureMetric, "closure_pass": closureMetric.MaxComponentAbs <= 1e-10}
	}
	result.R2Reconcile = map[string]interface{}{"threshold": fix001P3P93PolynomialOracleThreshold, "branches": map[string]interface{}{"real": polyRows["real"].(map[string]interface{})["historical_standard_vs_genuine_standard"], "imag": polyRows["imag"].(map[string]interface{})["historical_standard_vs_genuine_standard"]}, "directly_comparable": referenceComparable}
	result.R3Budget = map[string]interface{}{"budget": fix001P3P93PolynomialBudget, "historical_fast_vs_genuine_standard": map[string]interface{}{"real": polyRows["real"].(map[string]interface{})["historical_fast_vs_genuine_standard"], "imag": polyRows["imag"].(map[string]interface{})["historical_fast_vs_genuine_standard"]}, "pass": maxHistorical <= fix001P3P93PolynomialBudget}
	result.R4Regression = map[string]interface{}{"polynomial_budget": fix001P3P93PolynomialBudget, "current_production_fast_vs_genuine_standard": map[string]interface{}{"real": polyRows["real"].(map[string]interface{})["current_fast_vs_genuine_standard"], "imag": polyRows["imag"].(map[string]interface{})["current_fast_vs_genuine_standard"]}, "current_fast_vs_historical_fast": map[string]interface{}{"real": polyRows["real"].(map[string]interface{})["current_fast_vs_historical_fast"], "imag": polyRows["imag"].(map[string]interface{})["current_fast_vs_historical_fast"]}, "regression_vector_closure": map[string]interface{}{"real": polyRows["real"].(map[string]interface{})["regression_vector_closure"], "imag": polyRows["imag"].(map[string]interface{})["regression_vector_closure"]}, "closure_target": 1e-10, "closure_pass": closurePass}
	if !referenceComparable {
		result.Classification = "P93_HISTORICAL_POLYNOMIAL_REFERENCE_MISMATCH"
		return fix001P3P93PolynomialWrite(result, outPath)
	}
	if !closurePass {
		result.Classification = "P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT"
		return fix001P3P93PolynomialWrite(result, outPath)
	}

	checkpointRows := []map[string]interface{}{}
	firstBudget := map[string]interface{}{"checkpoint": "none", "branch": "none", "difference": 0.0}
	for _, branch := range []string{"real", "imag"} {
		state := branches[branch]
		productionEvents := fix001P3P93PolynomialEventMap(state.events, branch)
		referenceEvents := fix001P3P93PolynomialReferenceMap(referencePS, branch)
		for _, name := range fix001P3P93PolynomialSortedCommonNames(referenceEvents, productionEvents) {
			refRow, prodEvent := referenceEvents[name], productionEvents[name]
			refValues := fix001P3P93PolynomialTraceValues(refRow.Values)
			prodValues := fix001P3P93TraceValues(prodEvent)
			metric, err := fix001P3P93PolynomialMetric(refValues, prodValues, fix001P3P93PolynomialBudget)
			if err != nil {
				return err
			}
			row := map[string]interface{}{"branch": branch, "checkpoint": name, "operation": prodEvent.Name, "comparison_valid": true, "production_vs_historical": metric, "production_metadata": map[string]interface{}{"level": prodEvent.Level, "scale": prodEvent.Scale, "degree": prodEvent.Degree}, "historical_metadata": map[string]interface{}{"level": refRow.Level, "scale": refRow.Scale, "degree": refRow.Degree}, "production_capacity": fix001P3P93PolynomialCapacitySummary(prodEvent.Capacity), "historical_capacity": fix001P3P93PolynomialHistoricalCapacity(candidate, branch, name), "genuine_standard_metric_available": false}
			checkpointRows = append(checkpointRows, row)
			if firstBudget["checkpoint"] == "none" && metric.MaxComponentAbs >= fix001P3P93PolynomialBudget {
				firstBudget = map[string]interface{}{"checkpoint": name, "branch": branch, "difference": metric.MaxComponentAbs, "operation": prodEvent.Name, "incoming_difference": 0.0, "outgoing_difference": metric.MaxComponentAbs, "amplification": 0.0, "capacity": row["production_capacity"]}
			}
		}
	}
	for i := range checkpointRows {
		if checkpointRows[i]["branch"] != firstBudget["branch"] {
			continue
		}
		if checkpointRows[i]["checkpoint"] == firstBudget["checkpoint"] {
			if i > 0 && checkpointRows[i-1]["branch"] == firstBudget["branch"] {
				incoming := checkpointRows[i-1]["production_vs_historical"].(*semanticBisectMetric).MaxComponentAbs
				firstBudget["incoming_difference"] = incoming
				if incoming > 0 {
					firstBudget["amplification"] = firstBudget["difference"].(float64) / incoming
				}
			}
			break
		}
	}
	result.R5Checkpoints = checkpointRows
	result.R6Localization = map[string]interface{}{"threshold": fix001P3P93PolynomialBudget, "first_budget_breaking_regression": firstBudget, "generated_power_boundary": strings.HasPrefix(firstBudget["checkpoint"].(string), "T"), "classification_rule": "first Tn-final budget break is generated-power regression; first B/G/F0 budget break is PS accumulation regression"}

	result.R7Projection = map[string]interface{}{"directional_gain": 8.211976748833033, "polynomial_budget": fix001P3P93PolynomialBudget, "da2_pre_restore_budget": fix001P3P93DoubleAnglePreRestoreBudget, "historical_projected_da2_real": maxHistorical * 8.211976748833033, "historical_projected_da2_pass": maxHistorical*8.211976748833033 <= fix001P3P93DoubleAnglePreRestoreBudget, "current_projected_da2": maxCurrent * 8.211976748833033, "current_projected_da2_pass": maxCurrent*8.211976748833033 <= fix001P3P93DoubleAnglePreRestoreBudget, "plaintext_only": true}
	if maxHistorical > fix001P3P93PolynomialBudget {
		result.Classification = "P93_HISTORICAL_DESIGN_POLYNOMIAL_BUDGET_CONFLICT"
	} else if firstBudget["checkpoint"] == "none" {
		result.Classification = "P93_PRODUCTION_POLYNOMIAL_REGRESSION_UNLOCALIZED"
	} else if strings.HasPrefix(firstBudget["checkpoint"].(string), "T") {
		result.Classification = "P93_PRODUCTION_GENERATED_POWER_REGRESSION"
	} else {
		result.Classification = "P93_PRODUCTION_PS_ACCUMULATION_REGRESSION"
	}
	return fix001P3P93PolynomialWrite(result, outPath)
}

func fix001P3P93PlainVectorDifference(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

func fix001P3P93PlainVectorAdd(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func referencePolynomialLevel(rows []fix001P3P93ReferencePolynomialRow, branch string) int {
	for _, row := range rows {
		if row.Branch == branch && row.Exponent == 93 {
			return row.Level
		}
	}
	return -1
}

func referencePolynomialScale(rows []fix001P3P93ReferencePolynomialRow, branch string) string {
	for _, row := range rows {
		if row.Branch == branch && row.Exponent == 93 {
			return row.Scale
		}
	}
	return ""
}

func referencePolynomialDegree(rows []fix001P3P93ReferencePolynomialRow, branch string) int {
	for _, row := range rows {
		if row.Branch == branch && row.Exponent == 93 {
			return row.Degree
		}
	}
	return -1
}
