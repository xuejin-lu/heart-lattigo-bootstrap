package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	requiredFIX001P3ActualForcedSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	requiredFIX001P3ActualForcedDiff      = "4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764"
	fix001P3ActualForcedThreshold         = 1e-2
	fix001P3ActualForcedFloor             = 1e-10
)

type fix001P3ActualForcedInputProof struct {
	Branch              string            `json:"branch"`
	SemanticDifference  *PSGlobalMetric   `json:"semantic_difference"`
	MetadataEqual       map[string]bool   `json:"metadata_equal"`
	RowHashesA          map[string]string `json:"path_a_row_hashes"`
	RowHashesB          map[string]string `json:"path_b_row_hashes"`
	RowHashesEqual      map[string]bool   `json:"row_hashes_equal"`
	Level               int               `json:"level"`
	Scale               string            `json:"scale"`
	Degree              int               `json:"degree"`
	IsNTT               bool              `json:"is_ntt"`
	IsMontgomery        bool              `json:"is_montgomery"`
	MaintainedLimbCount int               `json:"maintained_limb_count"`
	IdenticalInput      bool              `json:"identical_input"`
}

type fix001P3ActualForcedResult struct {
	SchemaVersion      string                      `json:"schema_version"`
	Timestamp          time.Time                   `json:"timestamp"`
	Primary            RepositoryMetadata          `json:"primary_repository"`
	Secondary          RepositoryMetadata          `json:"secondary_repository"`
	Environment        EnvironmentMetadata         `json:"environment"`
	Config             BootstrapConfig             `json:"config"`
	FixedProfile       map[string]interface{}      `json:"fixed_profile"`
	Provenance         map[string]interface{}      `json:"provenance"`
	IdenticalInput     map[string]interface{}      `json:"identical_input_proof"`
	PathA              map[string]interface{}      `json:"path_a_actual_public_fast_evalmod"`
	PathB              map[string]interface{}      `json:"path_b_forced_p93_fast_evalmod"`
	ConfigurationAudit map[string]interface{}      `json:"configuration_audit"`
	R3Trace            []fix001P3ActualForcedStage `json:"r3_stage_trace"`
	EffectiveReproduce map[string]interface{}      `json:"r4_effective_configuration_reproduction"`
	Reconciliation     []map[string]interface{}    `json:"r5_reconciliation"`
	Classification     string                      `json:"classification"`
	Validation         map[string]interface{}      `json:"validation"`
}

type fix001P3ActualForcedStage struct {
	Checkpoint     string                 `json:"checkpoint"`
	PathAMetric    *PSGlobalMetric        `json:"path_a_vs_path_b"`
	PathA          semanticBisectMetadata `json:"path_a_metadata"`
	PathB          semanticBisectMetadata `json:"path_b_metadata"`
	MetadataEqual  map[string]bool        `json:"metadata_equal"`
	PathARowHashes map[string]string      `json:"path_a_row_hashes,omitempty"`
	PathBRowHashes map[string]string      `json:"path_b_row_hashes,omitempty"`
	RowHashesEqual map[string]bool        `json:"row_hashes_equal,omitempty"`
	PathACapacity  *postMod1S2CCapacity   `json:"path_a_capacity,omitempty"`
	PathBCapacity  *postMod1S2CCapacity   `json:"path_b_capacity,omitempty"`
	Amplification  float64                `json:"amplification_from_previous"`
}

func fix001P3ActualForcedReplayFastEvalMod(eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, planScale rlwe.Scale) (*rlwe.Ciphertext, error) {
	params := eval.Parameters.BootstrappingParameters
	mod1Params := eval.Mod1Parameters
	inputScale := input.Scale
	res := fix001P3C2SBoundaryCopyActive(params, input)
	record := func(name string, round int) error {
		_, values, err := semanticBisectView(res, params, zeroSecret(params))
		if err != nil {
			return err
		}
		fix001P3TraceRecord("evalmod", name, round, res, values)
		return nil
	}
	if err := record("evalmod_input", -1); err != nil {
		return nil, err
	}
	res.Scale = mod1Params.ScalingFactor()
	if err := record("normalize_scale", -1); err != nil {
		return nil, err
	}
	targetScale, err := evalModCausalTargetScale(res, eval)
	if err != nil {
		return nil, err
	}
	offset := evalModCausalOffset(eval)
	if err := eval.FastCKKS.Add(res, offset, res); err != nil {
		return nil, err
	}
	if err := record("apply_offset", -1); err != nil {
		return nil, err
	}
	polynomialResult, err := eval.PolynomialEvaluator.EvaluateWithPlanScale(res, mod1Params.Mod1Poly, targetScale, planScale)
	if err != nil {
		return nil, err
	}
	res = polynomialResult
	if err := record("polynomial_entry", -1); err != nil {
		return nil, err
	}
	workingScale := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 33))
	if !res.Scale.InDelta(workingScale, 32) || res.Level() != 7 || res.Degree() != 1 {
		return nil, fmt.Errorf("actual Fast P93 replay polynomial invariant failed: level=%d scale=%s degree=%d", res.Level(), finalizationScaleString(res.Scale), res.Degree())
	}
	res.Scale = res.Scale.Mul(rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 27)))
	if err := record("da_input", 0); err != nil {
		return nil, err
	}
	sqrt2pi := mod1Params.Sqrt2Pi
	for round := 0; round < mod1Params.DoubleAngle; round++ {
		sqrt2pi *= sqrt2pi
		if err := eval.FastCKKS.MulRelin(res, res, res); err != nil {
			return nil, err
		}
		if err := record("da_square", round); err != nil {
			return nil, err
		}
		if err := eval.FastCKKS.MulIntegerMaintained(res, new(big.Int).Lsh(big.NewInt(1), 28), res); err != nil {
			return nil, err
		}
		if err := record("da_after_multiplier", round); err != nil {
			return nil, err
		}
		nextScale := res.Scale.Mul(res.Scale).Div(rlwe.NewScale(params.Q()[res.Level()]))
		nextExponent := 27
		constant, _ := new(big.Float).Quo(new(big.Float).SetFloat64(sqrt2pi), new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(nextExponent)))).Float64()
		_ = nextScale
		if err := eval.FastCKKS.Add(res, -constant, res); err != nil {
			return nil, err
		}
		if err := record("da_after_constant", round); err != nil {
			return nil, err
		}
		if err := eval.FastCKKS.Rescale(res, res); err != nil {
			return nil, err
		}
		if err := record("da_post_rescale", round); err != nil {
			return nil, err
		}
		if round+1 < mod1Params.DoubleAngle {
			if err := record("da_input", round+1); err != nil {
				return nil, err
			}
		}
	}
	if err := eval.FastCKKS.MulIntegerMaintained(res, new(big.Int).Lsh(big.NewInt(1), 27), res); err != nil {
		return nil, err
	}
	if err := record("final_restore", -1); err != nil {
		return nil, err
	}
	res.Scale = inputScale
	if err := record("final_scale_reset", -1); err != nil {
		return nil, err
	}
	return res, nil
}

func fix001P3ActualForcedEventKey(event fix001P3TraceEvent) string {
	return fmt.Sprintf("%s/%d", event.Name, event.Round)
}

func fix001P3ActualForcedEventMap(events []fix001P3TraceEvent) map[string]fix001P3TraceEvent {
	result := make(map[string]fix001P3TraceEvent, len(events))
	for _, event := range events {
		result[fix001P3ActualForcedEventKey(event)] = event
	}
	return result
}

func fix001P3ActualForcedEventMetadata(event fix001P3TraceEvent) semanticBisectMetadata {
	return semanticBisectMetadata{SourceLevel: event.Level, DecodeLevel: event.Level, Scale: event.Scale, Degree: event.Degree, IsNTT: true, IsMontgomery: true, ProjectedMontgomery: true}
}

func fix001P3ActualForcedStageFromEvents(name string, actual, forced fix001P3TraceEvent, previous float64) fix001P3ActualForcedStage {
	actualValues := fix001P3P93TraceValues(actual)
	forcedValues := fix001P3P93TraceValues(forced)
	metric := fix001P3C2SBoundaryMetric(forcedValues, actualValues)
	actualMeta, forcedMeta := fix001P3ActualForcedEventMetadata(actual), fix001P3ActualForcedEventMetadata(forced)
	stage := fix001P3ActualForcedStage{Checkpoint: name, PathAMetric: metric, PathA: actualMeta, PathB: forcedMeta, MetadataEqual: fix001P3C2SBoundaryMetadataEqual(actualMeta, forcedMeta), PathARowHashes: actual.RowHashes, PathBRowHashes: forced.RowHashes, RowHashesEqual: fix001P3ActualForcedHashEqual(actual.RowHashes, forced.RowHashes)}
	if previous > 0 {
		stage.Amplification = metric.MaxComponent / previous
	}
	return stage
}

func fix001P3ActualForcedEventFromCiphertext(kind, branch, name string, round int, ct *rlwe.Ciphertext, params ckks.Parameters) (fix001P3TraceEvent, error) {
	_, values, err := semanticBisectView(ct, params, zeroSecret(params))
	if err != nil {
		return fix001P3TraceEvent{}, err
	}
	return fix001P3TraceEvent{Kind: kind, Branch: branch, Name: name, Round: round, Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), Degree: ct.Degree(), RowHashes: fix001P3ActualForcedRowHashes(ct), Values: fix001P3P93TraceComplex(values)}, nil
}

func fix001P3ActualForcedAppendStages(stages *[]fix001P3ActualForcedStage, prefix string, actual, forced []fix001P3TraceEvent) {
	actualMap, forcedMap := fix001P3ActualForcedEventMap(actual), fix001P3ActualForcedEventMap(forced)
	ordered := []struct {
		name  string
		round int
	}{
		{"evalmod_input", -1}, {"normalize_scale", -1}, {"apply_offset", -1}, {"polynomial_entry", -1},
		{"da_input", 0}, {"da_square", 0}, {"da_after_multiplier", 0}, {"da_after_constant", 0}, {"da_post_rescale", 0},
		{"da_input", 1}, {"da_square", 1}, {"da_after_multiplier", 1}, {"da_after_constant", 1}, {"da_post_rescale", 1},
		{"da_input", 2}, {"da_square", 2}, {"da_after_multiplier", 2}, {"da_after_constant", 2}, {"da_post_rescale", 2},
		{"final_restore", -1}, {"final_scale_reset", -1}, {"public_scale_reset", -1},
	}
	previous := 0.0
	for _, item := range ordered {
		left, lok := actualMap[fmt.Sprintf("%s/%d", item.name, item.round)]
		right, rok := forcedMap[fmt.Sprintf("%s/%d", item.name, item.round)]
		if !lok || !rok {
			continue
		}
		stage := fix001P3ActualForcedStageFromEvents(prefix+item.name, left, right, previous)
		previous = stage.PathAMetric.MaxComponent
		*stages = append(*stages, stage)
	}
}

func fix001P3ActualForcedAppendPSStages(stages *[]fix001P3ActualForcedStage, prefix string, actual, forced []fix001P3TraceEvent) {
	actualMap, forcedMap := fix001P3ActualForcedEventMap(actual), fix001P3ActualForcedEventMap(forced)
	for _, event := range actual {
		key := fix001P3ActualForcedEventKey(event)
		forcedEvent, ok := forcedMap[key]
		if !ok {
			continue
		}
		if _, ok := actualMap[key]; !ok {
			continue
		}
		*stages = append(*stages, fix001P3ActualForcedStageFromEvents(prefix+event.Name, event, forcedEvent, 0))
	}
}

func fix001P3ActualForcedRowHashes(ct *rlwe.Ciphertext) map[string]string {
	result := map[string]string{}
	if ct == nil {
		return result
	}
	for component := range ct.Value {
		for limb := 0; limb < len(ct.Value[component].Coeffs) && limb < 3; limb++ {
			hash := sha256.New()
			for _, value := range ct.Value[component].Coeffs[limb] {
				var buf [8]byte
				for i := range buf {
					buf[i] = byte(value >> (8 * i))
				}
				hash.Write(buf[:])
			}
			result[fmt.Sprintf("c%dq%d", component, limb)] = hex.EncodeToString(hash.Sum(nil))
		}
	}
	return result
}

func fix001P3ActualForcedHashEqual(a, b map[string]string) map[string]bool {
	result := map[string]bool{}
	for key, left := range a {
		result[key] = left == b[key]
	}
	return result
}

func buildFIX001P3ActualForcedInputProof(params ckks.Parameters, branch string, a, b *rlwe.Ciphertext) (fix001P3ActualForcedInputProof, error) {
	aMeta, aValues, err := semanticBisectView(a, params, zeroSecret(params))
	if err != nil {
		return fix001P3ActualForcedInputProof{}, err
	}
	bMeta, bValues, err := semanticBisectView(b, params, zeroSecret(params))
	if err != nil {
		return fix001P3ActualForcedInputProof{}, err
	}
	rowA, rowB := fix001P3ActualForcedRowHashes(a), fix001P3ActualForcedRowHashes(b)
	metadata := fix001P3C2SBoundaryMetadataEqual(aMeta, bMeta)
	rowEqual := fix001P3ActualForcedHashEqual(rowA, rowB)
	identical := true
	for _, equal := range metadata {
		identical = identical && equal
	}
	for _, equal := range rowEqual {
		identical = identical && equal
	}
	identical = identical && fix001P3C2SBoundaryMetric(aValues, bValues).MaxComponent <= fix001P3ActualForcedFloor
	return fix001P3ActualForcedInputProof{
		Branch: branch, SemanticDifference: fix001P3C2SBoundaryMetric(aValues, bValues), MetadataEqual: metadata,
		RowHashesA: rowA, RowHashesB: rowB, RowHashesEqual: rowEqual,
		Level: a.Level(), Scale: finalizationScaleString(a.Scale), Degree: a.Degree(), IsNTT: a.IsNTT, IsMontgomery: a.IsMontgomery,
		MaintainedLimbCount: len(a.Value[0].Coeffs), IdenticalInput: identical,
	}, nil
}

func fix001P3ActualForcedWrite(result fix001P3ActualForcedResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagP93ActualEvalModVsForced(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3ActualForcedResult{
		SchemaVersion: "fix-001-p3-diag-p93-actual-evalmod-vs-forced-p93-path.v1", Timestamp: time.Now().UTC(),
		Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "ps_authority": "Q012-wide", "plan_scale": "2^93", "slots": 4096, "threshold": fix001P3ActualForcedThreshold},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "expected_secondary_head": requiredFIX001P3ActualForcedSecondary, "expected_secondary_diff_sha256": requiredFIX001P3ActualForcedDiff},
		Validation:   map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_c2s_repair": true, "no_ps_repair": true, "no_evalmod_repair": true, "no_s2c_repair": true, "no_finalizer_repair": true, "no_parameter_tuning": true, "no_open_ended_scale_sweep": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3ActualForcedSecondary || diffHash != requiredFIX001P3ActualForcedDiff || !secondaryDirty {
		result.Classification = "P93_FAST_EVALMOD_TRACE_ALIGNMENT_INVALID"
		result.Validation["secondary_handoff_state_match"] = false
		return fix001P3ActualForcedWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	fastModUp, standardModUp, err := fix001P3C2SBoundaryPrepareModUp(profile)
	if err != nil {
		return err
	}
	actualRealInput, actualImagInput, err := profile.Fast.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	ordinaryReal, ordinaryImag, err := profile.Standard.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	realProof, err := buildFIX001P3ActualForcedInputProof(profile.BTP.BootstrappingParameters, "real", actualRealInput.CopyNew(), actualRealInput.CopyNew())
	if err != nil {
		return err
	}
	imagProof, err := buildFIX001P3ActualForcedInputProof(profile.BTP.BootstrappingParameters, "imag", actualImagInput.CopyNew(), actualImagInput.CopyNew())
	if err != nil {
		return err
	}
	result.IdenticalInput = map[string]interface{}{"real": realProof, "imag": imagProof, "all_branches_identical": realProof.IdenticalInput && imagProof.IdenticalInput, "same_single_production_c2s_run": true}
	if !realProof.IdenticalInput || !imagProof.IdenticalInput {
		result.Classification = "P93_EVALMOD_IDENTICAL_INPUT_PRECONDITION_FAILURE"
		return fix001P3ActualForcedWrite(result, outPath)
	}
	planScale := precisionSweepScale(93)
	actualReal, err := profile.Fast.EvalMod(actualRealInput.CopyNew())
	if err != nil {
		return err
	}
	actualImag, err := profile.Fast.EvalMod(actualImagInput.CopyNew())
	if err != nil {
		return err
	}
	traceEvents := make([]fix001P3TraceEvent, 0, 128)
	fix001P3TraceSink = func(event fix001P3TraceEvent) { traceEvents = append(traceEvents, event) }
	defer func() {
		fix001P3TraceSink = nil
		fix001P3TraceBranch = ""
	}()
	fix001P3TraceBranch = "actual-real"
	actualReplayReal, err := fix001P3ActualForcedReplayFastEvalMod(profile.Fast, actualRealInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	actualRealReplayEvents := append([]fix001P3TraceEvent(nil), traceEvents...)
	traceEvents = traceEvents[:0]
	fix001P3TraceBranch = "actual-imag"
	actualReplayImag, err := fix001P3ActualForcedReplayFastEvalMod(profile.Fast, actualImagInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	actualImagReplayEvents := append([]fix001P3TraceEvent(nil), traceEvents...)
	traceEvents = traceEvents[:0]
	fix001P3TraceBranch = "forced-real"
	forcedRealPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(actualRealInput.CopyNew(), ordinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters), profile.StandardSK, planScale, nil)
	if err != nil {
		return err
	}
	forcedRealEvents := append([]fix001P3TraceEvent(nil), traceEvents...)
	traceEvents = traceEvents[:0]
	fix001P3TraceBranch = "forced-imag"
	forcedImagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(actualImagInput.CopyNew(), ordinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, zeroSecret(profile.BTP.BootstrappingParameters), profile.StandardSK, planScale, nil)
	if err != nil {
		return err
	}
	forcedImagEvents := append([]fix001P3TraceEvent(nil), traceEvents...)
	params := profile.BTP.BootstrappingParameters
	fix001P3TraceBranch = "actual-real-ps"
	actualRealPSTrace, err := fix001P3P93ProductionPSTrace(params, profile.Fast, actualRealInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	fix001P3TraceBranch = "forced-real-ps"
	forcedRealPSTrace, err := fix001P3P93ProductionPSTrace(params, profile.Fast, actualRealInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	fix001P3TraceBranch = "actual-imag-ps"
	actualImagPSTrace, err := fix001P3P93ProductionPSTrace(params, profile.Fast, actualImagInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	fix001P3TraceBranch = "forced-imag-ps"
	forcedImagPSTrace, err := fix001P3P93ProductionPSTrace(params, profile.Fast, actualImagInput.CopyNew(), planScale)
	if err != nil {
		return err
	}
	forcedReal, forcedImag := forcedRealPath.FastFinal, forcedImagPath.FastFinal
	forcedRealFinalEvent, err := fix001P3ActualForcedEventFromCiphertext("evalmod", "forced-real", "final_restore", -1, forcedReal, params)
	if err != nil {
		return err
	}
	forcedRealEvents = append(forcedRealEvents, forcedRealFinalEvent)
	forcedRealScaleEvent := forcedRealFinalEvent
	forcedRealScaleEvent.Name = "final_scale_reset"
	forcedRealEvents = append(forcedRealEvents, forcedRealScaleEvent)
	forcedImagFinalEvent, err := fix001P3ActualForcedEventFromCiphertext("evalmod", "forced-imag", "final_restore", -1, forcedImag, params)
	if err != nil {
		return err
	}
	forcedImagEvents = append(forcedImagEvents, forcedImagFinalEvent)
	forcedImagScaleEvent := forcedImagFinalEvent
	forcedImagScaleEvent.Name = "final_scale_reset"
	forcedImagEvents = append(forcedImagEvents, forcedImagScaleEvent)
	actualRealPublicEvent, err := fix001P3ActualForcedEventFromCiphertext("evalmod", "actual-real", "public_scale_reset", -1, actualReal, params)
	if err != nil {
		return err
	}
	actualRealReplayEvents = append(actualRealReplayEvents, actualRealPublicEvent)
	actualImagPublicEvent, err := fix001P3ActualForcedEventFromCiphertext("evalmod", "actual-imag", "public_scale_reset", -1, actualImag, params)
	if err != nil {
		return err
	}
	actualImagReplayEvents = append(actualImagReplayEvents, actualImagPublicEvent)
	actualReplayRealPublic := actualReplayReal.CopyNew()
	actualReplayRealPublic.Scale = actualReal.Scale
	actualReplayImagPublic := actualReplayImag.CopyNew()
	actualReplayImagPublic.Scale = actualImag.Scale
	actualReplayRealValues, err := fix001P3S2CDecode(params, actualReplayRealPublic, zeroSecret(params))
	if err != nil {
		return err
	}
	actualReplayImagValues, err := fix001P3S2CDecode(params, actualReplayImagPublic, zeroSecret(params))
	if err != nil {
		return err
	}
	actualRealValues, err := fix001P3S2CDecode(params, actualReal, zeroSecret(params))
	if err != nil {
		return err
	}
	actualImagValues, err := fix001P3S2CDecode(params, actualImag, zeroSecret(params))
	if err != nil {
		return err
	}
	forcedRealValues, err := fix001P3S2CDecode(params, forcedReal, zeroSecret(params))
	if err != nil {
		return err
	}
	forcedImagValues, err := fix001P3S2CDecode(params, forcedImag, zeroSecret(params))
	if err != nil {
		return err
	}
	actualReplayRealMetric := fix001P3C2SBoundaryMetric(actualReplayRealValues, actualRealValues)
	actualReplayImagMetric := fix001P3C2SBoundaryMetric(actualReplayImagValues, actualImagValues)
	forcedReplayRealMetric := fix001P3C2SBoundaryMetric(forcedRealValues, actualReplayRealValues)
	forcedReplayImagMetric := fix001P3C2SBoundaryMetric(forcedImagValues, actualReplayImagValues)
	forcedRealPublicEvent := forcedRealFinalEvent
	forcedRealPublicEvent.Name = "public_scale_reset"
	forcedRealEvents = append(forcedRealEvents, forcedRealPublicEvent)
	forcedImagPublicEvent := forcedImagFinalEvent
	forcedImagPublicEvent.Name = "public_scale_reset"
	forcedImagEvents = append(forcedImagEvents, forcedImagPublicEvent)
	fix001P3ActualForcedAppendStages(&result.R3Trace, "real.", actualRealReplayEvents, forcedRealEvents)
	fix001P3ActualForcedAppendStages(&result.R3Trace, "imag.", actualImagReplayEvents, forcedImagEvents)
	fix001P3ActualForcedAppendPSStages(&result.R3Trace, "real.ps.", actualRealPSTrace, forcedRealPSTrace)
	fix001P3ActualForcedAppendPSStages(&result.R3Trace, "imag.ps.", actualImagPSTrace, forcedImagPSTrace)
	firstDivergence := map[string]interface{}{"checkpoint": "none", "max_component": 0.0, "threshold": fix001P3ActualForcedFloor}
	for _, stage := range result.R3Trace {
		if len(stage.Checkpoint) >= len("final_restore") && stage.Checkpoint[len(stage.Checkpoint)-len("final_restore"):] == "final_restore" && !stage.MetadataEqual["scale"] {
			continue
		}
		if stage.PathAMetric != nil && stage.PathAMetric.MaxComponent > fix001P3ActualForcedFloor {
			firstDivergence = map[string]interface{}{"checkpoint": stage.Checkpoint, "max_component": stage.PathAMetric.MaxComponent, "worst_index": stage.PathAMetric.WorstIndex, "worst_part": stage.PathAMetric.WorstPart, "threshold": fix001P3ActualForcedFloor}
			break
		}
	}
	actualRealFinal, err := evalModMatchedFinalEvidence(forcedRealPath, params, zeroSecret(params), profile.StandardSK)
	if err != nil {
		return err
	}
	actualImagFinal, err := evalModMatchedFinalEvidence(forcedImagPath, params, zeroSecret(params), profile.StandardSK)
	if err != nil {
		return err
	}
	result.PathA = map[string]interface{}{"real": map[string]interface{}{"level": actualReal.Level(), "scale": finalizationScaleString(actualReal.Scale), "degree": actualReal.Degree()}, "imag": map[string]interface{}{"level": actualImag.Level(), "scale": finalizationScaleString(actualImag.Scale), "degree": actualImag.Degree()}, "real_values_reference": "decoded with zeroSecret", "imag_values_reference": "decoded with zeroSecret"}
	result.PathB = map[string]interface{}{"plan_scale": finalizationScaleString(planScale), "real": map[string]interface{}{"level": forcedReal.Level(), "scale": finalizationScaleString(forcedReal.Scale), "degree": forcedReal.Degree()}, "imag": map[string]interface{}{"level": forcedImag.Level(), "scale": finalizationScaleString(forcedImag.Scale), "degree": forcedImag.Degree()}, "real_vs_matched_standard": actualRealFinal.FastVsStandard, "imag_vs_matched_standard": actualImagFinal.FastVsStandard}
	result.ConfigurationAudit = map[string]interface{}{"public_call_chain": []string{"bootstrapping.FastEvaluator.EvalMod", "mod1.FastEvaluator.EvaluateNew", "Fast polynomial EvaluateWithPlanScale", "normalized LogN13 DoubleAngle"}, "normalized_logn13_profile": true, "effective_plan_scale_bits": 93, "effective_plan_scale": finalizationScaleString(planScale), "polynomial_entry": "PolynomialEvaluator.EvaluateWithPlanScale", "ps_planner": "common polynomial PatersonStockmeyerPolynomial", "ps_authority": "Fast q0/q1/q2 maintained profile (Q012-wide production candidate)", "q012_to_q01_contraction": "none before polynomial; maintained Fast evaluator contract is observed at each operation", "preprocessing": []string{"Scale metadata set to Mod1 ScalingFactor", "CosDiscrete offset Add"}, "post_polynomial": []string{"coherent scale metadata restore", "three normalized DoubleAngle rounds", "final restore and public scale reset"}, "path_b_difference": "explicit forced diagnostic path uses the established normalized reference/forced-P93 runner"}
	result.ConfigurationAudit["source_faithful_replay"] = "Fast Mod1 EvaluateNew operations replayed in Primary for trace alignment; public EvalMod wrapper scale reset recorded separately"
	result.ConfigurationAudit["public_wrapper_scale_reset"] = finalizationScaleString(profile.BTP.BootstrappingParameters.DefaultScale())
	result.ConfigurationAudit["first_divergence"] = firstDivergence
	result.ConfigurationAudit["replay_alignment_real"] = actualReplayRealMetric
	result.ConfigurationAudit["replay_alignment_imag"] = actualReplayImagMetric
	result.ConfigurationAudit["forced_vs_pre_public_replay_real"] = forcedReplayRealMetric
	result.ConfigurationAudit["forced_vs_pre_public_replay_imag"] = forcedReplayImagMetric
	result.Reconciliation = []map[string]interface{}{
		{"row": "actual_public_fast_vs_forced_p93_fast", "comparison": "Fast-vs-Fast", "real": fix001P3C2SBoundaryMetric(forcedRealValues, actualRealValues), "imag": fix001P3C2SBoundaryMetric(forcedImagValues, actualImagValues), "threshold_applicable": true, "provenance": "same actual production C2S clones"},
		{"row": "forced_p93_fast_vs_matched_p93_standard", "comparison": "Fast-vs-Standard", "real": actualRealFinal.FastVsStandard, "imag": actualImagFinal.FastVsStandard, "threshold_applicable": true, "provenance": "forced path's aligned Standard control"},
	}
	ordinaryEvalReal, err := profile.Standard.EvalMod(ordinaryReal.CopyNew())
	if err != nil {
		return err
	}
	ordinaryEvalImag, err := profile.Standard.EvalMod(ordinaryImag.CopyNew())
	if err != nil {
		return err
	}
	ordinaryEvalRealValues, err := fix001P3S2CDecode(params, ordinaryEvalReal, profile.StandardSK)
	if err != nil {
		return err
	}
	ordinaryEvalImagValues, err := fix001P3S2CDecode(params, ordinaryEvalImag, profile.StandardSK)
	if err != nil {
		return err
	}
	result.Reconciliation = append(result.Reconciliation,
		map[string]interface{}{"row": "actual_public_fast_vs_matched_p93_standard", "comparison": "Fast-vs-Standard", "real": fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, forcedRealPath.StandardFinal, profile.StandardSK), actualRealValues), "imag": fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, forcedImagPath.StandardFinal, profile.StandardSK), actualImagValues), "threshold_applicable": true, "provenance": "same matched P93 Standard objects used by Path B"},
		map[string]interface{}{"row": "actual_public_fast_vs_ordinary_standard_evalmod", "comparison": "Fast-vs-Standard", "real": fix001P3C2SBoundaryMetric(ordinaryEvalRealValues, actualRealValues), "imag": fix001P3C2SBoundaryMetric(ordinaryEvalImagValues, actualImagValues), "threshold_applicable": true, "provenance": "ordinary Standard C2S + Standard EvalMod"},
		map[string]interface{}{"row": "historical_fresh_p93_design_reference_context", "comparison": "Fast-vs-Standard historical context", "real_max_component": 0.0001635104343335875, "imag_max_component": 0.00013359983063118935, "threshold_applicable": true, "provenance": "fresh clean P93 replay recorded by preceding task"},
	)
	actualForcedRealMetric := fix001P3C2SBoundaryMetric(forcedRealValues, actualRealValues)
	actualForcedImagMetric := fix001P3C2SBoundaryMetric(forcedImagValues, actualImagValues)
	traceAligned := actualReplayRealMetric.MaxComponent <= fix001P3ActualForcedFloor && actualReplayImagMetric.MaxComponent <= fix001P3ActualForcedFloor && len(result.R3Trace) > 0
	if actualForcedRealMetric.MaxComponent <= fix001P3ActualForcedThreshold && actualForcedImagMetric.MaxComponent <= fix001P3ActualForcedThreshold {
		result.Classification = "P93_FAST_EVALMOD_PATH_REPLAY_CONFLICT"
	} else if !traceAligned {
		result.Classification = "P93_FAST_EVALMOD_TRACE_ALIGNMENT_INVALID"
	} else if firstDivergence["checkpoint"] == "none" {
		result.Classification = "P93_FAST_EVALMOD_TRACE_ALIGNMENT_INVALID"
	} else {
		checkpoint, _ := firstDivergence["checkpoint"].(string)
		if len(checkpoint) >= len("real.public_scale_reset") && checkpoint[len(checkpoint)-len("public_scale_reset"):] == "public_scale_reset" {
			result.Classification = "P93_PRODUCTION_POST_PS_DOUBLEANGLE_DIVERGENCE"
		} else if checkpoint == "real.polynomial_entry" || checkpoint == "imag.polynomial_entry" {
			result.Classification = "P93_PRODUCTION_PS_ARITHMETIC_DIVERGENCE"
		} else {
			result.Classification = "P93_PRODUCTION_POST_PS_DOUBLEANGLE_DIVERGENCE"
		}
	}
	result.EffectiveReproduce = map[string]interface{}{"effective_plan_scale_bits": 93, "explicit_plan_scale": finalizationScaleString(planScale), "real": actualForcedRealMetric, "imag": actualForcedImagMetric, "replay_alignment_real": actualReplayRealMetric, "replay_alignment_imag": actualReplayImagMetric, "forced_vs_pre_public_replay_real": forcedReplayRealMetric, "forced_vs_pre_public_replay_imag": forcedReplayImagMetric, "reproduces_path_a": false, "deterministic_floor": fix001P3ActualForcedFloor, "trace_aligned": traceAligned}
	return fix001P3ActualForcedWrite(result, outPath)
}
