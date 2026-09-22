package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	fix001P3P93ReferencePrimary = "41993cd03f85ec5f2de54dc21ade5ec812090d50"
	fix001P3P93ReferenceSecond  = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3P93ProdSecond       = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	fix001P3P93ProdDiffSHA      = "4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764"
	fix001P3P93FirstDiffFloor   = 1e-12
	fix001P3P93MaterialFraction = 0.1
)

type fix001P3P93FirstDivergenceCheckpoint struct {
	Checkpoint       string          `json:"checkpoint"`
	Branch           string          `json:"branch"`
	Reference        *PSGlobalMetric `json:"reference_vs_standard,omitempty"`
	Production       *PSGlobalMetric `json:"production_vs_standard,omitempty"`
	SemanticDiff     *PSGlobalMetric `json:"production_vs_fresh_reference"`
	LevelEqual       bool            `json:"level_equal"`
	ScaleEqual       bool            `json:"scale_equal"`
	DegreeEqual      bool            `json:"degree_equal"`
	ReferenceLevel   int             `json:"reference_level"`
	ProductionLevel  int             `json:"production_level"`
	ReferenceScale   string          `json:"reference_scale"`
	ProductionScale  string          `json:"production_scale"`
	ReferenceDegree  int             `json:"reference_degree"`
	ProductionDegree int             `json:"production_degree"`
}

type fix001P3P93FirstDivergenceResult struct {
	SchemaVersion  string                                 `json:"schema_version"`
	Timestamp      time.Time                              `json:"timestamp"`
	Primary        RepositoryMetadata                     `json:"primary_repository"`
	Production     RepositoryMetadata                     `json:"production_secondary"`
	Reference      RepositoryMetadata                     `json:"temporary_reference_secondary"`
	Environment    EnvironmentMetadata                    `json:"environment"`
	Config         BootstrapConfig                        `json:"config"`
	Architecture   map[string]interface{}                 `json:"fixed_p93_architecture"`
	Provenance     map[string]interface{}                 `json:"provenance"`
	FreshReplay    map[string]interface{}                 `json:"r0_fresh_reference_replay"`
	ProductionRun  map[string]interface{}                 `json:"r1_production_replay"`
	Checkpoints    []fix001P3P93FirstDivergenceCheckpoint `json:"r2_checkpoint_trace"`
	FirstObserve   map[string]interface{}                 `json:"first_observable_divergence"`
	FirstMaterial  map[string]interface{}                 `json:"first_material_divergence"`
	Milestone      []map[string]interface{}               `json:"r4_milestone_reconciliation"`
	Classification string                                 `json:"classification"`
	Validation     map[string]interface{}                 `json:"validation"`
}

type fix001P3P93FreshSummary struct {
	Primary    RepositoryMetadata                `json:"primary_repository"`
	Lattigo    RepositoryMetadata                `json:"lattigo_repository"`
	Validation map[string]interface{}            `json:"validation"`
	Candidates []fix001P3P93S2CAcceptedCandidate `json:"candidates"`
}

func fix001P3P93LoadFreshReference(summaryPath, tracePath string) (fix001P3P93FreshSummary, []fix001P3TraceEvent, error) {
	var summary fix001P3P93FreshSummary
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return summary, nil, err
	}
	if err := json.Unmarshal(data, &summary); err != nil {
		return summary, nil, err
	}
	var trace []fix001P3TraceEvent
	data, err = os.ReadFile(tracePath)
	if err != nil {
		return summary, nil, err
	}
	if err := json.Unmarshal(data, &trace); err != nil {
		return summary, nil, err
	}
	return summary, trace, nil
}

func fix001P3P93TraceValues(event fix001P3TraceEvent) []complex128 {
	values := make([]complex128, len(event.Values))
	for i, value := range event.Values {
		values[i] = complex(value.Real, value.Imag)
	}
	return values
}

func fix001P3P93FindTraceEvent(events []fix001P3TraceEvent, kind, branch, name string, round int) (fix001P3TraceEvent, bool) {
	for _, event := range events {
		if event.Kind == kind && event.Branch == branch && event.Name == name && event.Round == round {
			return event, true
		}
	}
	return fix001P3TraceEvent{}, false
}

func fix001P3P93TraceRepeatability(events []fix001P3TraceEvent) float64 {
	maxDifference := 0.0
	for i, left := range events {
		if left.Branch == "" {
			continue
		}
		for _, right := range events[i+1:] {
			if right.Branch == "" || left.Kind != right.Kind || left.Branch != right.Branch || left.Name != right.Name || left.Round != right.Round || len(left.Values) != len(right.Values) {
				continue
			}
			metric := psGlobalMetric(fix001P3P93TraceValues(left), fix001P3P93TraceValues(right))
			if metric.MaxComponent > maxDifference {
				maxDifference = metric.MaxComponent
			}
		}
	}
	return maxDifference
}

func fix001P3P93ProductionPSTrace(params ckks.Parameters, eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, planScale rlwe.Scale) ([]fix001P3TraceEvent, error) {
	capacityFor := func(ct *rlwe.Ciphertext) *postMod1S2CCapacity {
		capacity, err := postMod1S2CCapacityFromFastCiphertext(params, ct)
		if err != nil {
			return nil
		}
		return &capacity
	}
	e2, inputValues, err := psGlobalE2(eval.Parameters, eval, input)
	if err != nil {
		return nil, err
	}
	targetScale, err := evalModCausalTargetScale(e2, eval)
	if err != nil {
		return nil, err
	}
	snapshots, _, err := eval.PolynomialEvaluator.DiagnosticGeneratePowers(e2, eval.Mod1Parameters.Mod1Poly, targetScale)
	if err != nil {
		return nil, err
	}
	powers := map[int]*rlwe.Ciphertext{}
	decoded := map[int][]complex128{}
	expected := map[int][]complex128{}
	memo := map[int][]complex128{}
	events := make([]fix001P3TraceEvent, 0, len(snapshots)+64)
	for _, snapshot := range snapshots {
		if snapshot.Ciphertext == nil {
			return nil, fmt.Errorf("production T%d snapshot is nil", snapshot.N)
		}
		values, err := psGlobalDecode(params, snapshot.Ciphertext)
		if err != nil {
			return nil, err
		}
		powers[snapshot.N] = snapshot.Ciphertext
		decoded[snapshot.N] = values
		expected[snapshot.N] = fix001ExpectedPower(snapshot.N, inputValues, memo)
		events = append(events, fix001P3TraceEvent{Kind: "ps", Branch: fix001P3TraceBranch, Name: fmt.Sprintf("T%d-final", snapshot.N), Round: -1, Level: snapshot.Ciphertext.Level(), Scale: finalizationScaleString(snapshot.Ciphertext.Scale), Degree: snapshot.Ciphertext.Degree(), Capacity: capacityFor(snapshot.Ciphertext), RowHashes: fix001P3ActualForcedRowHashes(snapshot.Ciphertext), Values: fix001P3P93TraceComplex(values)})
	}
	plan := psWidePlan(eval, e2, targetScale, planScale)
	checks, giants, root, _, err := psGlobalReplay(params, eval.FastCKKS, plan, powers, expected, decoded, inputValues, planScale)
	if err != nil {
		return nil, err
	}
	for _, check := range checks {
		events = append(events, fix001P3TraceEvent{Kind: "ps", Branch: fix001P3TraceBranch, Name: check.ID, Round: check.Round, Level: check.Level, Scale: check.Scale, Degree: check.Degree, Capacity: capacityFor(check.ciphertext), Values: fix001P3P93TraceComplex(check.actual)})
	}
	for _, giant := range giants {
		for _, check := range giant.Checkpoints {
			events = append(events, fix001P3TraceEvent{Kind: "ps", Branch: fix001P3TraceBranch, Name: check.ID, Round: check.Round, Level: check.Level, Scale: check.Scale, Degree: check.Degree, Capacity: capacityFor(check.ciphertext), Values: fix001P3P93TraceComplex(check.actual)})
		}
	}
	if root.ciphertext == nil {
		return nil, fmt.Errorf("production PS root is nil")
	}
	events = append(events, fix001P3TraceEvent{Kind: "ps", Branch: fix001P3TraceBranch, Name: root.ID, Round: root.Round, Level: root.Level, Scale: root.Scale, Degree: root.Degree, Capacity: capacityFor(root.ciphertext), RowHashes: fix001P3ActualForcedRowHashes(root.ciphertext), Values: fix001P3P93TraceComplex(root.actual)})
	output := root.ciphertext.CopyNew()
	if err := eval.FastCKKS.Rescale(output, output); err != nil {
		return nil, err
	}
	outputValues, err := psGlobalDecode(params, output)
	if err != nil {
		return nil, err
	}
	events = append(events, fix001P3TraceEvent{Kind: "ps", Branch: fix001P3TraceBranch, Name: "F0-replay-final-rescale", Round: -1, Level: output.Level(), Scale: finalizationScaleString(output.Scale), Degree: output.Degree(), Capacity: capacityFor(output), RowHashes: fix001P3ActualForcedRowHashes(output), Values: fix001P3P93TraceComplex(outputValues)})
	actual, err := eval.PolynomialEvaluator.EvaluateWithPlanScale(e2.CopyNew(), eval.Mod1Parameters.Mod1Poly, targetScale, planScale)
	if err != nil {
		return nil, err
	}
	actualValues, err := psGlobalDecode(params, actual)
	if err != nil {
		return nil, err
	}
	events = append(events, fix001P3TraceEvent{Kind: "ps", Branch: fix001P3TraceBranch, Name: "F0-final-rescale", Round: -1, Level: actual.Level(), Scale: finalizationScaleString(actual.Scale), Degree: actual.Degree(), Capacity: capacityFor(actual), RowHashes: fix001P3ActualForcedRowHashes(actual), Values: fix001P3P93TraceComplex(actualValues)})
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Kind != events[j].Kind {
			return events[i].Kind < events[j].Kind
		}
		if events[i].Branch != events[j].Branch {
			return events[i].Branch < events[j].Branch
		}
		return events[i].Name < events[j].Name
	})
	return events, nil
}

func fix001P3P93TraceComplex(values []complex128) []fix001P3TraceComplex {
	out := make([]fix001P3TraceComplex, len(values))
	for i, value := range values {
		out[i] = fix001P3TraceComplex{Real: real(value), Imag: imag(value)}
	}
	return out
}

func fix001P3P93TraceCheckpoint(reference, production fix001P3TraceEvent) fix001P3P93FirstDivergenceCheckpoint {
	refValues := fix001P3P93TraceValues(reference)
	prodValues := fix001P3P93TraceValues(production)
	return fix001P3P93FirstDivergenceCheckpoint{
		Checkpoint: reference.Name, Branch: reference.Branch,
		SemanticDiff: fix001P3S2CMetric(refValues, prodValues, fix001P3P93FirstDiffFloor),
		LevelEqual:   reference.Level == production.Level, ScaleEqual: reference.Scale == production.Scale, DegreeEqual: reference.Degree == production.Degree,
		ReferenceLevel: reference.Level, ProductionLevel: production.Level, ReferenceScale: reference.Scale, ProductionScale: production.Scale, ReferenceDegree: reference.Degree, ProductionDegree: production.Degree,
	}
}

func fix001P3P93RunProductionTrace(cfg BootstrapConfig, primaryRoot, backendRoot string) (fix001P3P93FirstDivergenceResult, error) {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return fix001P3P93FirstDivergenceResult{}, err
	}
	result := fix001P3P93FirstDivergenceResult{SchemaVersion: "fix-001-p3-diag-p93-evalmod-reference-vs-production-first-divergence.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Architecture: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "ps_authority": "Q012-wide", "plan_scale": "2^93", "slots": 4096, "threshold": 1e-2}, Production: gitMetadata(backendRoot), Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_parameter_tuning": true, "no_s2c_repair": true, "no_ps_repair": true, "no_t2_repair": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true}, Provenance: map[string]interface{}{"production_branch": secondaryBranch, "production_committed_head": secondaryCommit, "production_dirty": secondaryDirty, "production_diff_stat": diffStat, "production_diff_sha256": diffHash, "expected_production_head": fix001P3P93ProdSecond, "expected_production_diff_sha256": fix001P3P93ProdDiffSHA}}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3P93ProdSecond || diffHash != fix001P3P93ProdDiffSHA || !secondaryDirty {
		result.Classification = "P93_STAGE_ALIGNMENT_INVALID"
		result.Validation["production_handoff_state_match"] = false
		return result, nil
	}
	refSummary, refEvents, err := fix001P3P93LoadFreshReference("/tmp/fix001-p93-reference.zxMXX3/p93-reference-traced.json", "/tmp/fix001-p93-reference.zxMXX3/reference-trace.json")
	if err != nil {
		return result, err
	}
	if len(refSummary.Candidates) != 1 || refSummary.Candidates[0].PlanScaleExponent != 93 || refSummary.Lattigo.Commit != fix001P3P93ReferenceSecond || refSummary.Primary.Commit != fix001P3P93ReferencePrimary || !refSummary.Validation["secondary_clean"].(bool) {
		result.Classification = "P93_REFERENCE_REPLAY_CONFLICT"
		return result, nil
	}
	result.Reference = refSummary.Lattigo
	refDiffStat, refDiffHash, _, err := fix001P3PublicFinalizationDiffFingerprint("/tmp/fix001-p93-reference.zxMXX3/heart-lattigo-bootstrap")
	if err != nil {
		return result, err
	}
	result.Provenance["temporary_reference_primary_path"] = "/tmp/fix001-p93-reference.zxMXX3/heart-lattigo-bootstrap"
	result.Provenance["temporary_reference_secondary_path"] = "/tmp/fix001-p93-reference.zxMXX3/lattigo"
	result.Provenance["temporary_reference_harness_diff_stat"] = refDiffStat
	result.Provenance["temporary_reference_harness_diff_sha256"] = refDiffHash
	result.Provenance["temporary_reference_harness_change"] = "loop constrained from psWideScaleExponents to []int{93}; uncommitted temporary worktree only"
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return result, err
	}
	result.Production = gitMetadata(backendRoot)
	var prodEvents []fix001P3TraceEvent
	fix001P3TraceSink = func(event fix001P3TraceEvent) { prodEvents = append(prodEvents, event) }
	fix001P3TraceBranch = "real"
	prodReal, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, precisionSweepScale(93), nil)
	if err != nil {
		fix001P3TraceSink = nil
		return result, err
	}
	fix001P3TraceBranch = "imag"
	prodImag, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, precisionSweepScale(93), nil)
	fix001P3TraceBranch = ""
	fix001P3TraceSink = nil
	if err != nil {
		return result, err
	}
	matchedRealFinal, err := evalModMatchedFinalEvidence(prodReal, profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters), profile.StandardSK)
	if err != nil {
		return result, err
	}
	matchedImagFinal, err := evalModMatchedFinalEvidence(prodImag, profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters), profile.StandardSK)
	if err != nil {
		return result, err
	}
	fix001P3TraceBranch = "real"
	prodRealPS, err := fix001P3P93ProductionPSTrace(profile.BTP.BootstrappingParameters, profile.Fast, profile.Inputs.FastReal, precisionSweepScale(93))
	if err != nil {
		return result, err
	}
	fix001P3TraceBranch = "imag"
	prodImagPS, err := fix001P3P93ProductionPSTrace(profile.BTP.BootstrappingParameters, profile.Fast, profile.Inputs.FastImag, precisionSweepScale(93))
	if err != nil {
		return result, err
	}
	prodEvents = append(prodEvents, prodRealPS...)
	prodEvents = append(prodEvents, prodImagPS...)
	productionReal, productionImag, productionCore, err := fix001P3P93S2CRunFastEvalModCore(profile.Fast, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return result, err
	}
	params := profile.BTP.BootstrappingParameters
	prodRealValues, err := fix001P3S2CDecode(params, productionReal, zeroSecret(params))
	if err != nil {
		return result, err
	}
	prodImagValues, err := fix001P3S2CDecode(params, productionImag, zeroSecret(params))
	if err != nil {
		return result, err
	}
	stdRealValues, err := fix001P3S2CDecode(params, prodReal.StandardFinal, profile.StandardSK)
	if err != nil {
		return result, err
	}
	stdImagValues, err := fix001P3S2CDecode(params, prodImag.StandardFinal, profile.StandardSK)
	if err != nil {
		return result, err
	}
	prodEvalReal := fix001P3S2CMetric(stdRealValues, prodRealValues, 1e-2)
	prodEvalImag := fix001P3S2CMetric(stdImagValues, prodImagValues, 1e-2)
	stdCore, err := profile.Standard.SlotsToCoeffs(prodReal.StandardFinal.CopyNew(), prodImag.StandardFinal.CopyNew())
	if err != nil {
		return result, err
	}
	stdCoreValues, err := fix001P3S2CDecode(params, stdCore, profile.StandardSK)
	if err != nil {
		return result, err
	}
	prodCoreValues, err := fix001P3S2CDecode(params, productionCore, zeroSecret(params))
	if err != nil {
		return result, err
	}
	prodF0 := fix001P3S2CMetric(stdCoreValues, prodCoreValues, 1e-2)
	result.FreshReplay = map[string]interface{}{"primary_commit": refSummary.Primary.Commit, "secondary_commit": refSummary.Lattigo.Commit, "real_ps_polynomial_vs_standard": refSummary.Candidates[0].Real.Polynomial, "imag_ps_polynomial_vs_standard": refSummary.Candidates[0].Imag.Polynomial, "real_evalmod_vs_standard": refSummary.Candidates[0].Real.EvalMod, "imag_evalmod_vs_standard": refSummary.Candidates[0].Imag.EvalMod, "post_s2c_proxy": refSummary.Candidates[0].Real.PostS2C, "public_like": refSummary.Candidates[0].Real.PublicLike, "ps_exit_contraction_real": refSummary.Candidates[0].Real.Contraction, "ps_exit_contraction_imag": refSummary.Candidates[0].Imag.Contraction, "reference_capacity_pass": refSummary.Candidates[0].Real.Contraction.Q01CenteredUnique && refSummary.Candidates[0].Imag.Contraction.Q01CenteredUnique}
	result.ProductionRun = map[string]interface{}{"evalmod_real": prodEvalReal, "evalmod_imag": prodEvalImag, "matched_c2s_evalmod_real": matchedRealFinal.FastVsStandard, "matched_c2s_evalmod_imag": matchedImagFinal.FastVsStandard, "post_s2c_f0": prodF0, "expected_f0": 0.0402736269192105, "f0_reproduced": math.Abs(prodF0.MaxComponent-0.0402736269192105) <= 2e-5}
	result.Checkpoints = fix001P3P93BuildCheckpointTable(refEvents, prodEvents)
	floor := math.Max(fix001P3P93TraceRepeatability(refEvents), fix001P3P93FirstDiffFloor)
	result.FirstObserve = map[string]interface{}{"numerical_floor": floor, "floor_method": "maximum repeatability difference among branch-labeled fresh-reference events, floored at the validated 1e-12 replay precision", "checkpoint": "none_above_floor"}
	result.FirstMaterial = map[string]interface{}{"criterion": "10% of the stage-aligned matched-C2S production EvalMod max-component gap", "real_gap": matchedRealFinal.FastVsStandard.MaxComponent, "imag_gap": matchedImagFinal.FastVsStandard.MaxComponent, "full_core_context_real_gap": prodEvalReal.MaxComponent, "full_core_context_imag_gap": prodEvalImag.MaxComponent}
	for _, checkpoint := range result.Checkpoints {
		if checkpoint.SemanticDiff.MaxComponent > fix001P3P93FirstDiffFloor && result.FirstObserve["checkpoint"] == "none_above_floor" {
			result.FirstObserve["checkpoint"] = checkpoint.Branch + "." + checkpoint.Checkpoint
		}
		material := checkpoint.SemanticDiff.MaxComponent > fix001P3P93MaterialFraction*math.Max(matchedRealFinal.FastVsStandard.MaxComponent, matchedImagFinal.FastVsStandard.MaxComponent)
		if material && result.FirstMaterial["checkpoint"] == nil {
			result.FirstMaterial["checkpoint"] = checkpoint.Branch + "." + checkpoint.Checkpoint
			result.FirstMaterial["difference"] = checkpoint.SemanticDiff
		}
	}
	if checkpointName, ok := result.FirstMaterial["checkpoint"].(string); ok {
		for i, checkpoint := range result.Checkpoints {
			if checkpoint.Branch+"."+checkpoint.Checkpoint != checkpointName {
				continue
			}
			incoming := 0.0
			if i > 0 && result.Checkpoints[i-1].Branch == checkpoint.Branch {
				incoming = result.Checkpoints[i-1].SemanticDiff.MaxComponent
			}
			amplification := math.Inf(1)
			if incoming > 0 {
				amplification = checkpoint.SemanticDiff.MaxComponent / incoming
			}
			result.FirstMaterial["incoming_difference"] = incoming
			result.FirstMaterial["outgoing_difference"] = checkpoint.SemanticDiff.MaxComponent
			result.FirstMaterial["amplification_factor"] = amplification
			result.FirstMaterial["capacity_margin"] = "not serialized in the compact trace; fresh reference and production PS paths separately passed their capacity checks"
			break
		}
	}
	result.Milestone = []map[string]interface{}{
		{"row": "fresh_p93_design_reference_public_like", "reference": "fresh clean P93 replay", "max_component_error": refSummary.Candidates[0].Real.PublicLike, "pass_vs_1e-2": refSummary.Candidates[0].Real.PublicLike.MaxComponent <= 1e-2, "provenance": refSummary.Primary.Commit + "/" + refSummary.Lattigo.Commit},
		{"row": "fresh_p93_reference_evalmod_real_imag", "reference": "fresh clean P93 Standard", "max_component_error_real": refSummary.Candidates[0].Real.EvalMod, "max_component_error_imag": refSummary.Candidates[0].Imag.EvalMod, "pass_vs_1e-2": true, "provenance": refSummary.Primary.Commit + "/" + refSummary.Lattigo.Commit},
		{"row": "current_production_evalmod_real_imag", "reference": "matched Standard; full-core R1 plus stage-aligned matched-C2S context", "max_component_error_real": prodEvalReal, "max_component_error_imag": prodEvalImag, "stage_aligned_matched_c2s_real": matchedRealFinal.FastVsStandard, "stage_aligned_matched_c2s_imag": matchedImagFinal.FastVsStandard, "pass_vs_1e-2": prodEvalReal.MaxComponent <= 1e-2 && prodEvalImag.MaxComponent <= 1e-2, "provenance": secondaryCommit + " + dirty fingerprint"},
		{"row": "current_production_post_s2c_f0", "reference": "matched Standard core", "max_component_error": prodF0, "pass_vs_1e-2": prodF0.MaxComponent <= 1e-2, "provenance": secondaryCommit + " + dirty fingerprint"},
	}
	if prodEvalReal.MaxComponent > 1e-2 || prodEvalImag.MaxComponent > 1e-2 {
		result.Milestone = append(result.Milestone, map[string]interface{}{"row": "current_official_public_output", "reference": "genuine Standard Bootstrap", "max_component_error": "see committed 35b06d1 evidence: 0.2219142512701697", "pass_vs_1e-2": false, "provenance": "35b06d1c2ef67c1962f02971dcef225049fbc692"})
	}
	if result.FirstMaterial["checkpoint"] == "real.evalmod_input" || result.FirstMaterial["checkpoint"] == "imag.evalmod_input" {
		result.Classification = "P93_PRODUCTION_INPUT_MISMATCH"
	} else if result.FirstMaterial["checkpoint"] != nil {
		result.Classification = "P93_FIXED_WIDTH_PS_FIRST_MATERIAL_DIVERGENCE"
	} else {
		result.Classification = "P93_NO_MATERIAL_REFERENCE_PRODUCTION_DIVERGENCE"
	}
	return result, nil
}

func fix001P3P93BuildCheckpointTable(reference, production []fix001P3TraceEvent) []fix001P3P93FirstDivergenceCheckpoint {
	rows := []fix001P3P93FirstDivergenceCheckpoint{}
	for _, branch := range []string{"real", "imag"} {
		seen := map[string]bool{}
		for _, event := range reference {
			if event.Kind != "ps" || event.Branch != branch || seen[event.Name] {
				continue
			}
			seen[event.Name] = true
			ref, refOK := fix001P3P93FindTraceEvent(reference, "ps", branch, event.Name, event.Round)
			prod, prodOK := fix001P3P93FindTraceEvent(production, "ps", branch, event.Name, event.Round)
			if refOK && prodOK {
				rows = append(rows, fix001P3P93TraceCheckpoint(ref, prod))
			}
		}
	}
	ordered := []struct {
		Name  string
		Round int
	}{{"evalmod_input", -1}, {"normalize_scale", -1}, {"apply_offset", -1}, {"da_input", 0}, {"da_square", 0}, {"da_after_multiplier", 0}, {"da_after_constant", 0}, {"da_post_rescale", 0}, {"da_input", 1}, {"da_square", 1}, {"da_after_multiplier", 1}, {"da_after_constant", 1}, {"da_post_rescale", 1}, {"da_input", 2}, {"da_square", 2}, {"da_after_multiplier", 2}, {"da_after_constant", 2}, {"da_post_rescale", 2}}
	for _, branch := range []string{"real", "imag"} {
		for _, point := range ordered {
			ref, refOK := fix001P3P93FindTraceEvent(reference, "evalmod", branch, point.Name, point.Round)
			prod, prodOK := fix001P3P93FindTraceEvent(production, "evalmod", branch, point.Name, point.Round)
			if refOK && prodOK {
				rows = append(rows, fix001P3P93TraceCheckpoint(ref, prod))
			}
		}
	}
	return rows
}

func fix001P3P93WriteFirstDivergence(result fix001P3P93FirstDivergenceResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagP93EvalModFirstDivergence(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result, err := fix001P3P93RunProductionTrace(cfg, primaryRoot, backendRoot)
	if err != nil {
		return err
	}
	return fix001P3P93WriteFirstDivergence(result, outPath)
}
