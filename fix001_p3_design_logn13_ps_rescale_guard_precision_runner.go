//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3PSRescaleGuardPrimaryBase = "1a41918bba9754eb30b43b6ce5ad6ee6682c5cc7"
	requiredFIX001P3PSRescaleGuardSecondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	psRescaleGuardBudget                      = 1.2e-8
	psRescaleGuardValueTarget                 = 1e-10
	psRescaleGuardPublicThreshold             = 1e-2
)

type psRescaleGuardMetricPair struct {
	Real *PSGlobalMetric `json:"real"`
	Imag *PSGlobalMetric `json:"imag"`
}

type psRescaleGuardAlignment struct {
	ID                   string          `json:"id"`
	LeftScale            string          `json:"left_scale"`
	RightScale           string          `json:"right_scale"`
	ExactRatio           string          `json:"exact_ratio"`
	IntegerRatio         string          `json:"integer_ratio_used"`
	FractionalDifference string          `json:"fractional_difference"`
	IntegerRatioExact    bool            `json:"integer_ratio_exact"`
	MetadataOnlyRelabel  bool            `json:"metadata_only_relabel"`
	LocalSemanticError   *PSGlobalMetric `json:"local_semantic_error"`
}

type psRescaleGuardBoundary struct {
	ID                        string          `json:"id"`
	Order                     int             `json:"order"`
	InputLevel                int             `json:"input_level"`
	InputScale                string          `json:"input_scale"`
	OutputLevel               int             `json:"output_level"`
	OutputScale               string          `json:"output_scale"`
	GuardBits                 int             `json:"guard_bits"`
	PreCumulativeError        *PSGlobalMetric `json:"pre_rescale_cumulative_error"`
	GuardValuePreservation    *PSGlobalMetric `json:"guard_value_preservation"`
	PreCapacityRatio          float64         `json:"pre_rescale_capacity_ratio"`
	GuardedPreCapacityRatio   float64         `json:"guarded_pre_rescale_capacity_ratio"`
	LocalRescaleError         *PSGlobalMetric `json:"local_rescale_error"`
	PostCumulativeError       *PSGlobalMetric `json:"post_rescale_cumulative_error"`
	FirstDownstreamCheckpoint string          `json:"first_downstream_checkpoint"`
	CenteredUnique            bool            `json:"centered_unique"`
	RowsMatch                 bool            `json:"fast_vs_full_rns_rows_match"`
	LevelUnchanged            bool            `json:"level_unchanged"`
}

type psRescaleGuardReplay struct {
	Checks     []PSGlobalCheckpoint
	Root       PSGlobalCheckpoint
	Final      *rlwe.Ciphertext
	Boundaries []psRescaleGuardBoundary
	Alignments []psRescaleGuardAlignment
	FirstError string
	G0Merge    *psRescaleGuardG0MergeSnapshot
}

type psRescaleGuardG0MergeSnapshot struct {
	A               *rlwe.Ciphertext
	BRaw            *rlwe.Ciphertext
	Product         *rlwe.Ciphertext
	Output          *rlwe.Ciphertext
	AExpected       []complex128
	ProductExpected []complex128
}

type psRescaleGuardG0MergeOverride struct {
	A       *rlwe.Ciphertext
	Product *rlwe.Ciphertext
}

type psRescaleGuardBranchRun struct {
	Name            string
	Replay          psRescaleGuardReplay
	Final           *rlwe.Ciphertext
	FinalValues     []complex128
	StandardValues  []complex128
	PlainError      *PSGlobalMetric
	StandardError   *PSGlobalMetric
	FullError       *PSGlobalMetric
	CapacitySafe    bool
	AlignmentSafe   bool
	ValuePreserving bool
	LevelSafe       bool
	RowsMatch       bool
	Valid           bool
	FirstFailure    string
	OutputLevel     int
	OutputScale     string
}

type psRescaleGuardCandidate struct {
	Name                      string                    `json:"name"`
	BoundaryGuardBits         []int                     `json:"boundary_guard_bits"`
	GuardedBoundaries         []string                  `json:"guarded_boundaries"`
	TotalGuardBits            int                       `json:"total_guard_bits"`
	GuardedBoundaryCount      int                       `json:"guarded_boundary_count"`
	FinalRealError            float64                   `json:"final_real_error"`
	FinalImagError            float64                   `json:"final_imag_error"`
	FinalMaxError             float64                   `json:"final_max_error"`
	StandardRealError         float64                   `json:"final_real_vs_standard"`
	StandardImagError         float64                   `json:"final_imag_vs_standard"`
	FullRNSRealError          float64                   `json:"final_real_vs_full_rns"`
	FullRNSImagError          float64                   `json:"final_imag_vs_full_rns"`
	CapacitySafe              bool                      `json:"capacity_safe"`
	AlignmentSafe             bool                      `json:"scale_alignment_safe"`
	ValuePreserving           bool                      `json:"guard_value_preserving"`
	LevelSafe                 bool                      `json:"no_extra_level_consumption"`
	RowsMatch                 bool                      `json:"fast_vs_full_rns_rows_match"`
	Valid                     bool                      `json:"valid"`
	PrecisionQualified        bool                      `json:"precision_qualified"`
	FirstFailure              string                    `json:"first_failure"`
	FirstLimitingBoundary     string                    `json:"first_limiting_boundary"`
	WorstGuardedCapacityRatio float64                   `json:"worst_guarded_capacity_ratio"`
	WorstGuardedCapacityCheck string                    `json:"worst_guarded_capacity_checkpoint"`
	BoundaryEffects           []psRescaleGuardBoundary  `json:"boundary_effects,omitempty"`
	AlignmentAudits           []psRescaleGuardAlignment `json:"alignment_audits,omitempty"`
	OutputScaleReal           string                    `json:"output_scale_real,omitempty"`
	OutputScaleImag           string                    `json:"output_scale_imag,omitempty"`
	DownstreamReached         bool                      `json:"downstream_reached"`
	DownstreamReason          string                    `json:"downstream_reason,omitempty"`
	EvalModVsStandard         *psRescaleGuardMetricPair `json:"downstream_evalmod_vs_standard,omitempty"`
	PostS2CVsStandard         *PSGlobalMetric           `json:"downstream_post_s2c_vs_standard,omitempty"`
	PublicLike                *PSGlobalMetric           `json:"downstream_public_like,omitempty"`
	StandardPublicLike        *PSGlobalMetric           `json:"downstream_standard_public_like,omitempty"`
	PublicLikeMetadataCorrect bool                      `json:"downstream_public_metadata_correct"`
	PublicLikeInputUnchanged  bool                      `json:"downstream_input_unchanged"`
	realFinal                 *rlwe.Ciphertext
	imagFinal                 *rlwe.Ciphertext
}

type psRescaleGuardSingleResult struct {
	BoundaryID              string  `json:"boundary_id"`
	GuardBits               int     `json:"guard_bits"`
	GuardedPreCapacityRatio float64 `json:"guarded_pre_rescale_capacity_ratio"`
	LocalErrorBeforeReal    float64 `json:"local_rescale_error_before_real"`
	LocalErrorBeforeImag    float64 `json:"local_rescale_error_before_imag"`
	LocalErrorAfterReal     float64 `json:"local_rescale_error_after_real"`
	LocalErrorAfterImag     float64 `json:"local_rescale_error_after_imag"`
	FinalRealError          float64 `json:"final_real_error"`
	FinalImagError          float64 `json:"final_imag_error"`
	FinalMaxError           float64 `json:"final_max_error"`
	AlignmentSafe           bool    `json:"scale_alignment_safe"`
	ValuePreserving         bool    `json:"guard_value_preserving"`
	LevelSafe               bool    `json:"no_extra_level_consumption"`
	Valid                   bool    `json:"valid"`
	FirstFailure            string  `json:"first_failure"`
}

type psRescaleGuardGreedyStep struct {
	Step           int     `json:"step"`
	BoundaryID     string  `json:"boundary_id"`
	FromBits       int     `json:"from_guard_bits"`
	ToBits         int     `json:"to_guard_bits"`
	BeforeMaxError float64 `json:"before_max_error"`
	AfterMaxError  float64 `json:"after_max_error"`
	Improvement    float64 `json:"improvement"`
}

type psRescaleGuardResult struct {
	SchemaVersion     string                       `json:"schema_version"`
	Timestamp         time.Time                    `json:"timestamp"`
	Primary           RepositoryMetadata           `json:"primary_repository"`
	Lattigo           RepositoryMetadata           `json:"lattigo_repository"`
	Environment       EnvironmentMetadata          `json:"environment"`
	Config            BootstrapConfig              `json:"config"`
	Parameters        ExperimentParameters         `json:"effective_parameters"`
	Workload          CorrectnessWorkload          `json:"workload"`
	Budget            float64                      `json:"polynomial_budget"`
	PublicThreshold   float64                      `json:"public_like_threshold"`
	R0Baseline        map[string]interface{}       `json:"r0_baseline"`
	RescaleBoundaries []psRescaleGuardBoundary     `json:"rescale_boundaries"`
	SingleBoundary    []psRescaleGuardSingleResult `json:"single_boundary_one_bit_results"`
	GreedySteps       []psRescaleGuardGreedyStep   `json:"accepted_greedy_guard_steps"`
	BestSchedule      psRescaleGuardCandidate      `json:"best_valid_schedule"`
	SelectedSchedule  string                       `json:"selected_schedule,omitempty"`
	Downstream        map[string]interface{}       `json:"downstream,omitempty"`
	Classification    string                       `json:"classification"`
	FirstLimiting     string                       `json:"first_limiting_boundary"`
	Validation        map[string]interface{}       `json:"validation"`
}

type psRescaleGuardBranch struct {
	Name     string
	Base     polynomialDecompositionBranchWork
	Standard []complex128
}

type psRescaleGuardFinalOverride func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error)
type psRescaleGuardBoundaryOverride func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error)

func psRescaleGuardScheduleKey(bits []int) string {
	parts := make([]string, len(bits))
	for i, value := range bits {
		parts[i] = fmt.Sprintf("%d", value)
	}
	return strings.Join(parts, ",")
}

func psRescaleGuardScheduleName(bits []int, boundaries []psRescaleGuardBoundary, prefix string) string {
	parts := make([]string, 0, len(bits))
	for i, value := range bits {
		if value > 0 {
			parts = append(parts, fmt.Sprintf("%s=2^%d", boundaries[i].ID, value))
		}
	}
	if len(parts) == 0 {
		return prefix + "-none"
	}
	return prefix + "-" + strings.Join(parts, "+")
}

func psRescaleGuardBoundaryIndex(boundaries []psRescaleGuardBoundary) map[string]int {
	index := make(map[string]int, len(boundaries))
	for i := range boundaries {
		index[boundaries[i].ID] = i
	}
	return index
}

func psRescaleGuardRunBranch(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch psRescaleGuardBranch, schedule []int, boundaries []psRescaleGuardBoundary) (psRescaleGuardBranchRun, error) {
	return psRescaleGuardRunBranchWithOverride(params, eval, branch, schedule, boundaries, nil)
}

func psRescaleGuardRunBranchWithOverride(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch psRescaleGuardBranch, schedule []int, boundaries []psRescaleGuardBoundary, override psRescaleGuardFinalOverride) (psRescaleGuardBranchRun, error) {
	return psRescaleGuardRunBranchWithAllOverrides(params, eval, branch, schedule, boundaries, override, nil)
}

func psRescaleGuardRunBranchWithAllOverrides(params ckks.Parameters, eval *bootstrapping.FastEvaluator, branch psRescaleGuardBranch, schedule []int, boundaries []psRescaleGuardBoundary, override psRescaleGuardFinalOverride, boundaryOverride psRescaleGuardBoundaryOverride) (psRescaleGuardBranchRun, error) {
	run := psRescaleGuardBranchRun{Name: branch.Name, CapacitySafe: true, AlignmentSafe: true, ValuePreserving: true, LevelSafe: true, RowsMatch: true, FirstFailure: "none"}
	guardMap := make(map[string]int, len(boundaries))
	for i, boundary := range boundaries {
		if i < len(schedule) && schedule[i] > 0 {
			guardMap[boundary.ID] = schedule[i]
		}
	}
	plan := psOracleScalePlan(branch.Base.Plan, precisionSweepScale(92))
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(params, eval.FastCKKS, plan, branch.Base.OraclePowerMap, branch.Base.PowerExpected, branch.Base.OraclePowerValues, branch.Base.InputValues, precisionSweepScale(92), guardMap, override, boundaryOverride)
	if err != nil {
		return run, err
	}
	run.Replay = replay
	for _, boundary := range replay.Boundaries {
		if !boundary.CenteredUnique || !boundary.RowsMatch {
			run.CapacitySafe = false
			run.RowsMatch = false
			if run.FirstFailure == "none" {
				run.FirstFailure = boundary.ID
			}
		}
		if boundary.GuardValuePreservation != nil && !boundary.GuardValuePreservation.Pass {
			run.ValuePreserving = false
			if run.FirstFailure == "none" {
				run.FirstFailure = boundary.ID + ".guard_value"
			}
		}
		if !boundary.LevelUnchanged {
			run.LevelSafe = false
			if run.FirstFailure == "none" {
				run.FirstFailure = boundary.ID + ".level"
			}
		}
	}
	for _, alignment := range replay.Alignments {
		if alignment.MetadataOnlyRelabel || alignment.LocalSemanticError == nil || !alignment.LocalSemanticError.Pass {
			run.AlignmentSafe = false
			if run.FirstFailure == "none" {
				run.FirstFailure = alignment.ID
			}
		}
	}
	if replay.FirstError != "" && run.FirstFailure == "none" {
		run.FirstFailure = replay.FirstError
	}
	if replay.Final == nil {
		return run, fmt.Errorf("rescale guard replay returned nil final ciphertext")
	}
	output := replay.Final.CopyNew()
	run.Final = output
	postValues, err := psGlobalDecode(params, output)
	if err != nil {
		return run, err
	}
	run.FinalValues = postValues
	run.PlainError = psRescaleGuardMetric(branch.Base.ReferenceValues, postValues)
	run.StandardValues = branch.Standard
	run.StandardError = psRescaleGuardMetric(branch.Standard, postValues)
	if output.Level() >= 1 {
		fullValues, rows, err := psOracleScaleFullMirror(params, output)
		if err != nil {
			return run, err
		}
		run.FullError = psRescaleGuardMetric(fullValues, postValues)
		run.RowsMatch = run.RowsMatch && rows
		if !rows && run.FirstFailure == "none" {
			run.FirstFailure = "F0-final-rescale.rows"
		}
	} else {
		run.FullError = &PSGlobalMetric{Pass: true, Threshold: psRescaleGuardValueTarget, WorstIndex: -1}
	}
	run.OutputLevel = output.Level()
	run.OutputScale = psRescaleGuardScaleString(output.Scale)
	run.CapacitySafe = run.CapacitySafe && run.RowsMatch
	run.Valid = run.CapacitySafe && run.AlignmentSafe && run.ValuePreserving && run.LevelSafe && run.FullError != nil && run.FullError.Pass && run.FirstFailure == "none"
	return run, nil
}

func psRescaleGuardBoundaryEffects(real, imag psRescaleGuardReplay) []psRescaleGuardBoundary {
	merged := make([]psRescaleGuardBoundary, len(real.Boundaries))
	for i, boundary := range real.Boundaries {
		merged[i] = boundary
		if i >= len(imag.Boundaries) {
			continue
		}
		other := imag.Boundaries[i]
		merged[i].PreCumulativeError = psRescaleGuardWorstMetric(boundary.PreCumulativeError, other.PreCumulativeError)
		merged[i].GuardValuePreservation = psRescaleGuardWorstMetric(boundary.GuardValuePreservation, other.GuardValuePreservation)
		merged[i].LocalRescaleError = psRescaleGuardWorstMetric(boundary.LocalRescaleError, other.LocalRescaleError)
		merged[i].PostCumulativeError = psRescaleGuardWorstMetric(boundary.PostCumulativeError, other.PostCumulativeError)
		merged[i].PreCapacityRatio = math.Max(boundary.PreCapacityRatio, other.PreCapacityRatio)
		merged[i].GuardedPreCapacityRatio = math.Max(boundary.GuardedPreCapacityRatio, other.GuardedPreCapacityRatio)
		merged[i].CenteredUnique = boundary.CenteredUnique && other.CenteredUnique
		merged[i].RowsMatch = boundary.RowsMatch && other.RowsMatch
		merged[i].LevelUnchanged = boundary.LevelUnchanged && other.LevelUnchanged
		if boundary.FirstDownstreamCheckpoint == "none" || (other.FirstDownstreamCheckpoint != "none" && other.FirstDownstreamCheckpoint < boundary.FirstDownstreamCheckpoint) {
			merged[i].FirstDownstreamCheckpoint = other.FirstDownstreamCheckpoint
		}
	}
	return merged
}

func psRescaleGuardWorstMetric(a, b *PSGlobalMetric) *PSGlobalMetric {
	if a == nil {
		return b
	}
	if b == nil || a.MaxComponent >= b.MaxComponent {
		return a
	}
	return b
}

func psRescaleGuardMakeCandidate(params ckks.Parameters, eval *bootstrapping.FastEvaluator, realBranch, imagBranch psRescaleGuardBranch, realStandard, imagStandard []complex128, schedule []int, name string, baselineReal, baselineImag *psRescaleGuardBranchRun, boundaries []psRescaleGuardBoundary) (psRescaleGuardCandidate, psRescaleGuardBranchRun, psRescaleGuardBranchRun, error) {
	return psRescaleGuardMakeCandidateWithOverrides(params, eval, realBranch, imagBranch, realStandard, imagStandard, schedule, name, baselineReal, baselineImag, boundaries, nil, nil)
}

func psRescaleGuardMakeCandidateWithOverrides(params ckks.Parameters, eval *bootstrapping.FastEvaluator, realBranch, imagBranch psRescaleGuardBranch, realStandard, imagStandard []complex128, schedule []int, name string, baselineReal, baselineImag *psRescaleGuardBranchRun, boundaries []psRescaleGuardBoundary, realOverride, imagOverride psRescaleGuardFinalOverride) (psRescaleGuardCandidate, psRescaleGuardBranchRun, psRescaleGuardBranchRun, error) {
	return psRescaleGuardMakeCandidateWithAllOverrides(params, eval, realBranch, imagBranch, realStandard, imagStandard, schedule, name, baselineReal, baselineImag, boundaries, realOverride, imagOverride, nil, nil)
}

func psRescaleGuardMakeCandidateWithAllOverrides(params ckks.Parameters, eval *bootstrapping.FastEvaluator, realBranch, imagBranch psRescaleGuardBranch, realStandard, imagStandard []complex128, schedule []int, name string, baselineReal, baselineImag *psRescaleGuardBranchRun, boundaries []psRescaleGuardBoundary, realOverride, imagOverride psRescaleGuardFinalOverride, realBoundaryOverride, imagBoundaryOverride psRescaleGuardBoundaryOverride) (psRescaleGuardCandidate, psRescaleGuardBranchRun, psRescaleGuardBranchRun, error) {
	realRun, err := psRescaleGuardRunBranchWithAllOverrides(params, eval, realBranch, schedule, boundaries, realOverride, realBoundaryOverride)
	if err != nil {
		return psRescaleGuardCandidate{}, realRun, psRescaleGuardBranchRun{}, err
	}
	imagRun, err := psRescaleGuardRunBranchWithAllOverrides(params, eval, imagBranch, schedule, boundaries, imagOverride, imagBoundaryOverride)
	if err != nil {
		return psRescaleGuardCandidate{}, realRun, imagRun, err
	}
	candidate := psRescaleGuardCandidate{Name: name, BoundaryGuardBits: append([]int(nil), schedule...), GuardedBoundaries: []string{}, FirstFailure: realRun.FirstFailure}
	for i, bits := range schedule {
		candidate.TotalGuardBits += bits
		if bits > 0 {
			candidate.GuardedBoundaryCount++
			candidate.GuardedBoundaries = append(candidate.GuardedBoundaries, boundaries[i].ID)
		}
	}
	if candidate.FirstFailure == "none" && imagRun.FirstFailure != "none" {
		candidate.FirstFailure = imagRun.FirstFailure
	}
	if realRun.PlainError != nil {
		candidate.FinalRealError = realRun.PlainError.MaxComponent
	}
	if imagRun.PlainError != nil {
		candidate.FinalImagError = imagRun.PlainError.MaxComponent
	}
	if realRun.StandardError != nil {
		candidate.StandardRealError = realRun.StandardError.MaxComponent
	}
	if imagRun.StandardError != nil {
		candidate.StandardImagError = imagRun.StandardError.MaxComponent
	}
	if realRun.FullError != nil {
		candidate.FullRNSRealError = realRun.FullError.MaxComponent
	}
	if imagRun.FullError != nil {
		candidate.FullRNSImagError = imagRun.FullError.MaxComponent
	}
	candidate.FinalMaxError = math.Max(candidate.FinalRealError, candidate.FinalImagError)
	candidate.CapacitySafe = realRun.CapacitySafe && imagRun.CapacitySafe
	candidate.AlignmentSafe = realRun.AlignmentSafe && imagRun.AlignmentSafe
	candidate.ValuePreserving = realRun.ValuePreserving && imagRun.ValuePreserving
	candidate.LevelSafe = realRun.LevelSafe && imagRun.LevelSafe
	candidate.RowsMatch = realRun.RowsMatch && imagRun.RowsMatch
	candidate.Valid = realRun.Valid && imagRun.Valid
	candidate.PrecisionQualified = candidate.Valid && candidate.FinalRealError <= psRescaleGuardBudget && candidate.FinalImagError <= psRescaleGuardBudget && candidate.StandardRealError <= psRescaleGuardBudget && candidate.StandardImagError <= psRescaleGuardBudget
	candidate.BoundaryEffects = psRescaleGuardBoundaryEffects(realRun.Replay, imagRun.Replay)
	candidate.AlignmentAudits = append(append([]psRescaleGuardAlignment(nil), realRun.Replay.Alignments...), imagRun.Replay.Alignments...)
	for _, boundary := range candidate.BoundaryEffects {
		if boundary.GuardBits > 0 && boundary.GuardedPreCapacityRatio > candidate.WorstGuardedCapacityRatio {
			candidate.WorstGuardedCapacityRatio = boundary.GuardedPreCapacityRatio
			candidate.WorstGuardedCapacityCheck = boundary.ID
		}
		if boundary.GuardBits > 0 && candidate.FirstLimitingBoundary == "" && (!boundary.CenteredUnique || !boundary.RowsMatch || (boundary.GuardValuePreservation != nil && !boundary.GuardValuePreservation.Pass)) {
			candidate.FirstLimitingBoundary = boundary.ID
		}
	}
	if candidate.FirstLimitingBoundary == "" && candidate.FirstFailure != "none" {
		candidate.FirstLimitingBoundary = candidate.FirstFailure
	}
	if baselineReal != nil && baselineImag != nil {
		candidate.LevelSafe = candidate.LevelSafe && realRun.OutputLevel == baselineReal.OutputLevel && imagRun.OutputLevel == baselineImag.OutputLevel
	}
	candidate.OutputScaleReal, candidate.OutputScaleImag = realRun.OutputScale, imagRun.OutputScale
	candidate.realFinal, candidate.imagFinal = realRun.Final, imagRun.Final
	return candidate, realRun, imagRun, nil
}

func psRescaleGuardScaleString(scale rlwe.Scale) string {
	return finalizationScaleString(scale)
}

func psRescaleGuardFractionalDifference(ratio rlwe.Scale, integer *big.Int) string {
	integerScale := rlwe.NewScale(integer)
	difference := new(big.Float).Sub(&ratio.Value, &integerScale.Value)
	return difference.Text('e', 40)
}

func psRescaleGuardMetric(expected, actual []complex128) *PSGlobalMetric {
	metric := psGlobalMetric(expected, actual)
	metric.Threshold = psRescaleGuardValueTarget
	metric.Pass = metric.MaxComponent <= psRescaleGuardValueTarget
	return metric
}

func psRescaleGuardCapacity(params ckks.Parameters, ct *rlwe.Ciphertext) (float64, bool, bool, error) {
	if ct == nil || ct.Level() < 1 {
		return 0, true, true, nil
	}
	capacity, err := postMod1S2CCapacityFromFastCiphertext(params, ct)
	if err != nil {
		return 0, false, false, err
	}
	rows := false
	if capacity.Pass {
		full, _, err := postMod1S2CLiftFull(params, ct)
		if err != nil {
			return capacity.MaxAbsOverQ01Half, false, false, err
		}
		rows = postMod1S2CRowsEqual(params, ct, full)
	}
	return capacity.MaxAbsOverQ01Half, capacity.Pass, rows, nil
}

func psRescaleGuardApply(eval *fastckks.Evaluator, params ckks.Parameters, ct *rlwe.Ciphertext, bits int) (*PSGlobalMetric, error) {
	if bits == 0 {
		return &PSGlobalMetric{Pass: true, Threshold: psRescaleGuardValueTarget, WorstIndex: -1}, nil
	}
	before, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil, err
	}
	factor := new(big.Int).Lsh(big.NewInt(1), uint(bits))
	if err := eval.MulIntegerMaintained(ct, factor, ct); err != nil {
		return nil, err
	}
	ct.Scale = ct.Scale.Mul(rlwe.NewScale(factor))
	after, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil, err
	}
	return psRescaleGuardMetric(before, after), nil
}

func psRescaleGuardAddAligned(params ckks.Parameters, eval *fastckks.Evaluator, a, b *rlwe.Ciphertext, id string, alignments *[]psRescaleGuardAlignment) error {
	if a.Scale.Equal(b.Scale) {
		return eval.Add(b, a, b)
	}
	left, err := psGlobalDecode(params, a)
	if err != nil {
		return err
	}
	right, err := psGlobalDecode(params, b)
	if err != nil {
		return err
	}
	if b.Scale.Cmp(a.Scale) > 0 {
		ratio := b.Scale.Div(a.Scale)
		ratioInt := ratio.BigInt()
		scratch := fastckks.NewCiphertext(params.Parameters, a.Degree(), minInt(a.Level(), b.Level()))
		psGlobalCopyMaintained(params, a, scratch)
		if err := eval.MulIntegerMaintained(a, ratioInt, scratch); err != nil {
			return err
		}
		scratch.Scale = b.Scale
		aligned, err := psGlobalDecode(params, scratch)
		if err != nil {
			return err
		}
		*alignments = append(*alignments, psRescaleGuardAlignment{ID: id, LeftScale: psRescaleGuardScaleString(a.Scale), RightScale: psRescaleGuardScaleString(b.Scale), ExactRatio: psRescaleGuardScaleString(ratio), IntegerRatio: ratioInt.String(), FractionalDifference: psRescaleGuardFractionalDifference(ratio, ratioInt), IntegerRatioExact: ratio.Equal(rlwe.NewScale(ratioInt)), MetadataOnlyRelabel: false, LocalSemanticError: psRescaleGuardMetric(left, aligned)})
		return eval.Add(b, scratch, b)
	}
	ratio := a.Scale.Div(b.Scale)
	ratioInt := ratio.BigInt()
	scratch := fastckks.NewCiphertext(params.Parameters, b.Degree(), minInt(a.Level(), b.Level()))
	psGlobalCopyMaintained(params, b, scratch)
	if err := eval.MulIntegerMaintained(b, ratioInt, scratch); err != nil {
		return err
	}
	scratch.Scale = a.Scale
	aligned, err := psGlobalDecode(params, scratch)
	if err != nil {
		return err
	}
	*alignments = append(*alignments, psRescaleGuardAlignment{ID: id, LeftScale: psRescaleGuardScaleString(a.Scale), RightScale: psRescaleGuardScaleString(b.Scale), ExactRatio: psRescaleGuardScaleString(ratio), IntegerRatio: ratioInt.String(), FractionalDifference: psRescaleGuardFractionalDifference(ratio, ratioInt), IntegerRatioExact: ratio.Equal(rlwe.NewScale(ratioInt)), MetadataOnlyRelabel: false, LocalSemanticError: psRescaleGuardMetric(right, aligned)})
	if err := eval.Add(scratch, a, scratch); err != nil {
		return err
	}
	psGlobalCopyMaintained(params, scratch, b)
	return nil
}

func psRescaleGuardReplayPS(params ckks.Parameters, eval *fastckks.Evaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, powerExpected, powerDecoded map[int][]complex128, input []complex128, candidate rlwe.Scale, guards map[string]int, override psRescaleGuardFinalOverride) (psRescaleGuardReplay, error) {
	return psRescaleGuardReplayPSWithBoundaryOverride(params, eval, plan, powers, powerExpected, powerDecoded, input, candidate, guards, override, nil)
}

func psRescaleGuardReplayPSWithBoundaryOverride(params ckks.Parameters, eval *fastckks.Evaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, powerExpected, powerDecoded map[int][]complex128, input []complex128, candidate rlwe.Scale, guards map[string]int, override psRescaleGuardFinalOverride, boundaryOverride psRescaleGuardBoundaryOverride, extras ...interface{}) (psRescaleGuardReplay, error) {
	out := psRescaleGuardReplay{}
	var mergeOverride *psRescaleGuardG0MergeOverride
	var resets []psGlobalReplayReset
	for _, extra := range extras {
		switch value := extra.(type) {
		case psGlobalReplayReset:
			resets = append(resets, value)
		case psRescaleGuardG0MergeOverride:
			mergeOverride = &value
		case *psRescaleGuardG0MergeOverride:
			mergeOverride = value
		}
	}
	applyReset := func(id string, current **rlwe.Ciphertext) {
		for _, reset := range resets {
			if id == reset.ID && reset.Ciphertext != nil {
				*current = reset.Ciphertext.CopyNew()
			}
		}
	}
	type node struct {
		degree   int
		value    *rlwe.Ciphertext
		source   []complex128
		identity string
	}
	baby := make([]*node, len(plan.Value))
	for i := range plan.Value {
		p := plan.Value[i]
		value := fastckks.NewCiphertext(params.Parameters, 1, p.Level)
		value.IsNTT, value.IsMontgomery = powers[1].IsNTT, powers[1].IsMontgomery
		*value.MetaData = *powers[1].MetaData
		value.Scale = candidate
		psGlobalZero(value)
		source := make([]complex128, len(input))
		identity := psGlobalSubtreeHash(fmt.Sprintf("block-%d", i), p.Polynomial)
		if p.IsEven {
			if p.Coeffs[0] == nil {
				return out, fmt.Errorf("nil block constant")
			}
			if err := eval.Add(value, p.Coeffs[0], value); err != nil {
				return out, err
			}
			psGlobalAddScaled(source, makeOnes(len(input)), psGlobalCoeff(p.Coeffs[0]))
			decoded, err := psGlobalDecode(params, value)
			if err != nil {
				return out, err
			}
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("B%d-constant", i), "baby_constant_add", value, source, decoded, source, identity, fmt.Sprintf("Chebyshev block %d degree=%d", i, p.Degree()), nil))
			applyReset(out.Checks[len(out.Checks)-1].ID, &value)
		}
		for key := p.Degree(); key > 0; key-- {
			if (p.IsEven || p.IsOdd) && ((key&1 == 0 && !p.IsEven) || (key&1 == 1 && !p.IsOdd)) {
				continue
			}
			if p.Coeffs[key] == nil {
				continue
			}
			power := powers[key]
			if power == nil {
				return out, fmt.Errorf("missing oracle power T%d", key)
			}
			before, err := psGlobalDecode(params, value)
			if err != nil {
				return out, err
			}
			if power.Scale.Cmp(value.Scale) > 0 {
				ratio := power.Scale.Div(value.Scale)
				ratioInt := ratio.BigInt()
				if err := eval.MulIntegerMaintained(value, ratioInt, value); err != nil {
					return out, err
				}
				value.Scale = power.Scale
			}
			if err := eval.MulThenAdd(powers[key], p.Coeffs[key], value); err != nil {
				return out, err
			}
			psGlobalAddScaled(source, powerExpected[key], psGlobalCoeff(p.Coeffs[key]))
			decoded, err := psGlobalDecode(params, value)
			if err != nil {
				return out, err
			}
			local := append([]complex128(nil), before...)
			psGlobalAddScaled(local, powerDecoded[key], psGlobalCoeff(p.Coeffs[key]))
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("B%d-term-%d", i, key), "baby_mul_then_add", value, source, decoded, local, identity, fmt.Sprintf("Chebyshev block %d degree=%d", i, p.Degree()), nil))
			applyReset(out.Checks[len(out.Checks)-1].ID, &value)
		}
		baby[len(plan.Value)-i-1] = &node{degree: p.Degree(), value: value, source: source, identity: identity}
	}
	order := 0
	for len(baby) != 1 {
		giant := make([]int, len(baby))
		for i := 0; i < len(baby); i++ {
			if i == len(baby)-1 {
				giant[i] = 2
			} else if baby[i].degree == baby[i+1].degree {
				giant[i] = 1
				i++
			}
		}
		for i := 0; i < len(baby); i++ {
			if giant[i] == 2 {
				baby[i].degree = baby[i-1].degree
				continue
			}
			if giant[i] != 1 {
				continue
			}
			a, b := baby[i], baby[i+1]
			if order == 0 && out.G0Merge == nil {
				out.G0Merge = &psRescaleGuardG0MergeSnapshot{A: a.value.CopyNew(), AExpected: append([]complex128(nil), a.source...)}
			}
			deg := 1 << bitLen(uint64(a.degree))
			aDecoded, err := psGlobalDecode(params, a.value)
			if err != nil {
				return out, err
			}
			beforeB, err := psGlobalDecode(params, b.value)
			if err != nil {
				return out, err
			}
			out.Checks = append(out.Checks, psGlobalCheckpointWithDecode(fmt.Sprintf("G%d-input-b", order), "giant_input_b", b.value, b.source, beforeB, b.identity, "PS child branch b", nil))
			if b.value.Degree() == 2 {
				if err := eval.Relinearize(b.value, b.value); err != nil {
					return out, err
				}
				after, err := psGlobalDecode(params, b.value)
				if err != nil {
					return out, err
				}
				out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-relinearize", order), "giant_relinearize", b.value, b.source, after, beforeB, b.identity, "same subtree after Fast relinearize", nil))
				applyReset(out.Checks[len(out.Checks)-1].ID, &b.value)
			}
			beforeRescale, err := psGlobalDecode(params, b.value)
			if err != nil {
				return out, err
			}
			id := fmt.Sprintf("G%d-rescale", order)
			boundary := psRescaleGuardBoundary{ID: id, Order: len(out.Boundaries), InputLevel: b.value.Level(), InputScale: psRescaleGuardScaleString(b.value.Scale), GuardBits: guards[id], FirstDownstreamCheckpoint: "none"}
			boundary.PreCumulativeError = psGlobalMetric(b.source, beforeRescale)
			boundary.PreCapacityRatio, boundary.CenteredUnique, boundary.RowsMatch, err = psRescaleGuardCapacity(params, b.value)
			if err != nil {
				return out, err
			}
			handled := false
			if boundaryOverride != nil {
				handled, err = boundaryOverride(params, eval, b.value, &boundary)
				if err != nil {
					return out, err
				}
			}
			var guardedBefore []complex128
			if !handled {
				boundary.GuardValuePreservation, err = psRescaleGuardApply(eval, params, b.value, guards[id])
				if err != nil {
					return out, err
				}
				boundary.GuardedPreCapacityRatio, boundary.CenteredUnique, boundary.RowsMatch, err = psRescaleGuardCapacity(params, b.value)
				if err != nil {
					return out, err
				}
				guardedBefore, err = psGlobalDecode(params, b.value)
				if err != nil {
					return out, err
				}
				if err := eval.Rescale(b.value, b.value); err != nil {
					return out, err
				}
			}
			if order == 0 && out.G0Merge != nil {
				out.G0Merge.BRaw = b.value.CopyNew()
			}
			afterRescale, err := psGlobalDecode(params, b.value)
			if err != nil {
				return out, err
			}
			if !handled {
				boundary.LocalRescaleError = psRescaleGuardMetric(guardedBefore, afterRescale)
			}
			boundary.PostCumulativeError = psGlobalMetric(b.source, afterRescale)
			boundary.OutputLevel = b.value.Level()
			boundary.OutputScale = psRescaleGuardScaleString(b.value.Scale)
			boundary.LevelUnchanged = boundary.OutputLevel == boundary.InputLevel-1
			out.Boundaries = append(out.Boundaries, boundary)
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(id, "giant_rescale", b.value, b.source, afterRescale, beforeRescale, b.identity, "same subtree after Fast Rescale", nil))
			applyReset(out.Checks[len(out.Checks)-1].ID, &b.value)
			if err := eval.Mul(b.value, powers[deg], b.value); err != nil {
				return out, err
			}
			productSource := append([]complex128(nil), b.source...)
			for k := range productSource {
				productSource[k] *= powerExpected[deg][k]
			}
			productDecoded, err := psGlobalDecode(params, b.value)
			if err != nil {
				return out, err
			}
			localProduct := make([]complex128, len(afterRescale))
			powerValues := powerDecoded[deg]
			for k := range localProduct {
				localProduct[k] = afterRescale[k] * powerValues[k]
			}
			productID := hashSourceVector(fmt.Sprintf("%s*T%d", b.identity, deg))
			if order == 0 && out.G0Merge != nil {
				out.G0Merge.Product = b.value.CopyNew()
				out.G0Merge.ProductExpected = append([]complex128(nil), productSource...)
			}
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-multiply", order), "giant_multiply", b.value, productSource, productDecoded, localProduct, productID, fmt.Sprintf("(%s) * T%d", b.identity, deg), nil))
			applyReset(out.Checks[len(out.Checks)-1].ID, &b.value)
			productForAdd := productDecoded
			parentSource := append([]complex128(nil), a.source...)
			psGlobalAddScaled(parentSource, productSource, 1)
			parentID := hashSourceVector(fmt.Sprintf("%s+%s", a.identity, productID))
			if order == 0 && mergeOverride != nil {
				if mergeOverride.A != nil {
					a.value = mergeOverride.A.CopyNew()
				}
				if mergeOverride.Product != nil {
					b.value = mergeOverride.Product.CopyNew()
				}
			}
			if err := psRescaleGuardAddAligned(params, eval, a.value, b.value, fmt.Sprintf("G%d-add", order), &out.Alignments); err != nil {
				return out, err
			}
			if order == 0 && out.G0Merge != nil {
				out.G0Merge.Output = b.value.CopyNew()
			}
			parentDecoded, err := psGlobalDecode(params, b.value)
			if err != nil {
				return out, err
			}
			localParent := append([]complex128(nil), aDecoded...)
			psGlobalAddScaled(localParent, productForAdd, 1)
			out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-add", order), "giant_add_aligned", b.value, parentSource, parentDecoded, localParent, parentID, fmt.Sprintf("%s + (%s)", a.identity, productID), nil))
			applyReset(out.Checks[len(out.Checks)-1].ID, &b.value)
			b.degree = 2*deg - 1
			b.source, b.identity = parentSource, parentID
			baby[i] = nil
			order++
			i++
		}
		kept := baby[:0]
		for _, step := range baby {
			if step != nil {
				kept = append(kept, step)
			}
		}
		baby = kept
	}
	root := baby[0]
	if root.value.Degree() == 2 {
		if err := eval.Relinearize(root.value, root.value); err != nil {
			return out, err
		}
	}
	decoded, err := psGlobalDecode(params, root.value)
	if err != nil {
		return out, err
	}
	rootBeforeFinal := root.value.CopyNew()
	out.Root = psGlobalCheckpointWithDecode("F0-root", "pre_final_rescale_root", rootBeforeFinal, root.source, decoded, root.identity, "PS root before final Rescale", nil)
	applyReset(out.Root.ID, &root.value)
	out.Root.ciphertext = root.value.CopyNew()
	out.Root.Level = root.value.Level()
	out.Root.Degree = root.value.Degree()
	out.Root.Scale = finalizationScaleString(root.value.Scale)
	out.Root.Rows = finalizationEvidence(root.value).Rows
	finalBefore := decoded
	finalID := "F0-final-rescale"
	finalBoundary := psRescaleGuardBoundary{ID: finalID, Order: len(out.Boundaries), InputLevel: root.value.Level(), InputScale: psRescaleGuardScaleString(root.value.Scale), GuardBits: guards[finalID], FirstDownstreamCheckpoint: "final_output"}
	finalBoundary.PreCumulativeError = psGlobalMetric(root.source, finalBefore)
	finalBoundary.PreCapacityRatio, finalBoundary.CenteredUnique, finalBoundary.RowsMatch, err = psRescaleGuardCapacity(params, root.value)
	if err != nil {
		return out, err
	}
	handled := false
	if override != nil {
		handled, err = override(params, eval, root.value, &finalBoundary)
		if err != nil {
			return out, err
		}
	}
	var guardedFinalBefore []complex128
	if !handled {
		finalBoundary.GuardValuePreservation, err = psRescaleGuardApply(eval, params, root.value, guards[finalID])
		if err != nil {
			return out, err
		}
		finalBoundary.GuardedPreCapacityRatio, finalBoundary.CenteredUnique, finalBoundary.RowsMatch, err = psRescaleGuardCapacity(params, root.value)
		if err != nil {
			return out, err
		}
		guardedFinalBefore, err = psGlobalDecode(params, root.value)
		if err != nil {
			return out, err
		}
		if err := eval.Rescale(root.value, root.value); err != nil {
			return out, err
		}
	}
	finalAfter, err := psGlobalDecode(params, root.value)
	if err != nil {
		return out, err
	}
	if !handled {
		finalBoundary.LocalRescaleError = psRescaleGuardMetric(guardedFinalBefore, finalAfter)
	}
	finalBoundary.PostCumulativeError = psGlobalMetric(root.source, finalAfter)
	finalBoundary.OutputLevel = root.value.Level()
	finalBoundary.OutputScale = psRescaleGuardScaleString(root.value.Scale)
	finalBoundary.LevelUnchanged = finalBoundary.OutputLevel == finalBoundary.InputLevel-1
	out.Boundaries = append(out.Boundaries, finalBoundary)
	out.Checks = append(out.Checks, psGlobalCheckpointWithLocal(finalID, "final_rescale", root.value, root.source, finalAfter, finalBefore, root.identity, "PS final output after Rescale", nil))
	applyReset(out.Checks[len(out.Checks)-1].ID, &root.value)
	out.Final = root.value.CopyNew()
	for i := range out.Boundaries {
		boundary := &out.Boundaries[i]
		for _, checkpoint := range out.Checks {
			if checkpoint.ID == boundary.ID || checkpoint.ID == "F0-root" {
				continue
			}
			if boundary.ID == "F0-final-rescale" {
				break
			}
			if strings.HasPrefix(checkpoint.ID, fmt.Sprintf("G%d-", boundary.Order)) && checkpoint.LocalConsistency != nil && checkpoint.LocalConsistency.MaxComponent > 0 {
				boundary.FirstDownstreamCheckpoint = checkpoint.ID
				break
			}
		}
	}
	return out, nil
}

func psRescaleGuardWrite(result psRescaleGuardResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func psRescaleGuardMetricValue(metric *PSGlobalMetric) float64 {
	if metric == nil {
		return 0
	}
	return metric.MaxComponent
}

func psRescaleGuardBoundaryByID(boundaries []psRescaleGuardBoundary, id string) *psRescaleGuardBoundary {
	for i := range boundaries {
		if boundaries[i].ID == id {
			return &boundaries[i]
		}
	}
	return nil
}

func psRescaleGuardSelectBetter(a, b psRescaleGuardCandidate, boundaries []psRescaleGuardBoundary) bool {
	if !a.Valid {
		return false
	}
	if !b.Valid {
		return true
	}
	if a.TotalGuardBits != b.TotalGuardBits {
		return a.TotalGuardBits < b.TotalGuardBits
	}
	if a.GuardedBoundaryCount != b.GuardedBoundaryCount {
		return a.GuardedBoundaryCount < b.GuardedBoundaryCount
	}
	if a.FinalMaxError != b.FinalMaxError {
		return a.FinalMaxError < b.FinalMaxError
	}
	for i := range a.BoundaryGuardBits {
		if a.BoundaryGuardBits[i] != b.BoundaryGuardBits[i] {
			return a.BoundaryGuardBits[i] < b.BoundaryGuardBits[i]
		}
	}
	_ = boundaries
	return a.Name < b.Name
}

func psRescaleGuardLimitingBoundary(candidate psRescaleGuardCandidate) string {
	if candidate.FirstLimitingBoundary != "" {
		return candidate.FirstLimitingBoundary
	}
	worst := ""
	worstValue := float64(-1)
	for _, boundary := range candidate.BoundaryEffects {
		value := psRescaleGuardMetricValue(boundary.LocalRescaleError)
		if value > worstValue {
			worstValue, worst = value, boundary.ID
		}
	}
	return worst
}

func psRescaleGuardSingleEvidence(boundaryID string, baseline, candidate psRescaleGuardCandidate) psRescaleGuardSingleResult {
	evidence := psRescaleGuardSingleResult{BoundaryID: boundaryID, GuardBits: 1, AlignmentSafe: candidate.AlignmentSafe, ValuePreserving: candidate.ValuePreserving, LevelSafe: candidate.LevelSafe, Valid: candidate.Valid, FirstFailure: candidate.FirstFailure, FinalRealError: candidate.FinalRealError, FinalImagError: candidate.FinalImagError, FinalMaxError: candidate.FinalMaxError}
	base := psRescaleGuardBoundaryByID(baseline.BoundaryEffects, boundaryID)
	guarded := psRescaleGuardBoundaryByID(candidate.BoundaryEffects, boundaryID)
	if base != nil {
		evidence.LocalErrorBeforeReal = psRescaleGuardMetricValue(base.LocalRescaleError)
		evidence.LocalErrorBeforeImag = psRescaleGuardMetricValue(base.LocalRescaleError)
	}
	if guarded != nil {
		evidence.GuardedPreCapacityRatio = guarded.GuardedPreCapacityRatio
		evidence.LocalErrorAfterReal = psRescaleGuardMetricValue(guarded.LocalRescaleError)
		evidence.LocalErrorAfterImag = evidence.LocalErrorAfterReal
	}
	return evidence
}

func runFIX001P3DesignLogN13PSRescaleGuardPrecision(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := psRescaleGuardResult{
		SchemaVersion:   "fix-001-p3-design-logn13-ps-rescale-guard-precision.v1",
		Timestamp:       time.Now().UTC(),
		Primary:         gitMetadata(primaryRoot),
		Lattigo:         gitMetadata(backendRoot),
		Environment:     EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:          cfg,
		Parameters:      parameterMetadata(residual, btp),
		Workload:        CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		Budget:          psRescaleGuardBudget,
		PublicThreshold: psRescaleGuardPublicThreshold,
		R0Baseline:      map[string]interface{}{},
		Validation: map[string]interface{}{
			"primary_required_base":           requiredFIX001P3PSRescaleGuardPrimaryBase,
			"secondary_commit":                secondaryCommit,
			"secondary_clean":                 secondaryClean,
			"secondary_exact_required_commit": secondaryCommit == requiredFIX001P3PSRescaleGuardSecondary,
			"no_secondary_production_changes": true,
			"no_generated_power_redesign":     true,
			"fixed_common_2^92_plan":          true,
			"no_mixed_or_raised_plan_scales":  true,
			"no_q0_q1_widening":               true,
			"no_q2_plus_arithmetic":           true,
			"no_c2s_change":                   true,
			"no_s2c_change":                   true,
			"no_mod1_parameter_retuning":      true,
			"no_extra_q_levels":               true,
			"no_metadata_only_relabel":        true,
			"no_production_integration":       true,
			"no_logn16":                       true,
			"no_benchmark":                    true,
			"no_gate_4_or_5":                  true,
			"no_exp003":                       true,
		},
		Classification: "logn13_ps_rescale_guard_precondition_mismatch",
		FirstLimiting:  "startup_precondition",
	}
	if !secondaryClean || secondaryCommit != requiredFIX001P3PSRescaleGuardSecondary {
		return psRescaleGuardWrite(result, outPath)
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
	realStandard, err := psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastReal, inputs.OrdinaryReal, standardSK, realBase.InputValues)
	if err != nil {
		return err
	}
	imagStandard, err := psOracleScaleStandardPolynomial(btp, fastEval, standardEval, btp.BootstrappingParameters, inputs.FastImag, inputs.OrdinaryImag, standardSK, imagBase.InputValues)
	if err != nil {
		return err
	}
	realBranch := psRescaleGuardBranch{Name: "real", Base: realBase, Standard: realStandard}
	imagBranch := psRescaleGuardBranch{Name: "imag", Base: imagBase, Standard: imagStandard}
	initialReal, err := psRescaleGuardRunBranch(btp.BootstrappingParameters, fastEval, realBranch, nil, nil)
	if err != nil {
		return err
	}
	initialImag, err := psRescaleGuardRunBranch(btp.BootstrappingParameters, fastEval, imagBranch, nil, nil)
	if err != nil {
		return err
	}
	boundaries := psRescaleGuardBoundaryEffects(initialReal.Replay, initialImag.Replay)
	allZero := make([]int, len(boundaries))
	cache := make(map[string]psRescaleGuardCandidate)
	evaluations := make(map[string][2]psRescaleGuardBranchRun)
	evaluate := func(name string, schedule []int) (psRescaleGuardCandidate, error) {
		key := psRescaleGuardScheduleKey(schedule)
		if existing, ok := cache[key]; ok {
			return existing, nil
		}
		candidate, realRun, imagRun, err := psRescaleGuardMakeCandidate(btp.BootstrappingParameters, fastEval, realBranch, imagBranch, realStandard, imagStandard, schedule, name, &initialReal, &initialImag, boundaries)
		if err != nil {
			return psRescaleGuardCandidate{}, err
		}
		cache[key] = candidate
		evaluations[key] = [2]psRescaleGuardBranchRun{realRun, imagRun}
		return candidate, nil
	}
	baseline, err := evaluate("R0-fixed-2^92", allZero)
	if err != nil {
		return err
	}
	r0Pass := baseline.FinalMaxError > 2.5e-8 && baseline.FinalMaxError < 4.2e-8 && baseline.CapacitySafe && baseline.AlignmentSafe && baseline.ValuePreserving && baseline.LevelSafe && baseline.RowsMatch && baseline.Valid
	result.R0Baseline = map[string]interface{}{"pass": r0Pass, "candidate": baseline, "expected_max_error_range": []float64{2.5e-8, 4.2e-8}}
	if !r0Pass {
		result.FirstLimiting = baseline.FirstFailure
		return psRescaleGuardWrite(result, outPath)
	}
	result.RescaleBoundaries = baseline.BoundaryEffects
	allSeen := []psRescaleGuardCandidate{baseline}
	for i, boundary := range boundaries {
		schedule := append([]int(nil), allZero...)
		schedule[i] = 1
		candidate, err := evaluate("R2-single-"+boundary.ID, schedule)
		if err != nil {
			return err
		}
		allSeen = append(allSeen, candidate)
		result.SingleBoundary = append(result.SingleBoundary, psRescaleGuardSingleEvidence(boundary.ID, baseline, candidate))
	}
	currentSchedule := append([]int(nil), allZero...)
	current := baseline
	for step := 0; step < 8; step++ {
		bestCandidate := psRescaleGuardCandidate{}
		bestBoundary := -1
		bestImprovement := float64(0)
		for i := range boundaries {
			if currentSchedule[i] >= 3 {
				continue
			}
			trialSchedule := append([]int(nil), currentSchedule...)
			trialSchedule[i]++
			trial, err := evaluate(fmt.Sprintf("R3-step-%d-%s", step+1, boundaries[i].ID), trialSchedule)
			if err != nil {
				return err
			}
			improvement := current.FinalMaxError - trial.FinalMaxError
			if !trial.Valid || improvement <= 0 {
				continue
			}
			if bestBoundary < 0 || improvement > bestImprovement {
				bestCandidate, bestBoundary, bestImprovement = trial, i, improvement
			}
		}
		if bestBoundary < 0 {
			break
		}
		from := currentSchedule[bestBoundary]
		currentSchedule[bestBoundary]++
		result.GreedySteps = append(result.GreedySteps, psRescaleGuardGreedyStep{Step: len(result.GreedySteps) + 1, BoundaryID: boundaries[bestBoundary].ID, FromBits: from, ToBits: currentSchedule[bestBoundary], BeforeMaxError: current.FinalMaxError, AfterMaxError: bestCandidate.FinalMaxError, Improvement: bestImprovement})
		current = bestCandidate
		allSeen = append(allSeen, current)
	}
	for _, candidate := range result.SingleBoundary {
		_ = candidate
	}
	best := baseline
	for _, candidate := range allSeen {
		if candidate.Valid && (!best.Valid || candidate.FinalMaxError < best.FinalMaxError || (candidate.FinalMaxError == best.FinalMaxError && psRescaleGuardSelectBetter(candidate, best, boundaries))) {
			best = candidate
		}
	}
	result.BestSchedule = best
	selected := psRescaleGuardCandidate{}
	for _, candidate := range allSeen {
		if candidate.PrecisionQualified && (result.SelectedSchedule == "" || psRescaleGuardSelectBetter(candidate, selected, boundaries)) {
			selected = candidate
			result.SelectedSchedule = candidate.Name
		}
	}
	if result.SelectedSchedule != "" {
		result.Classification = "logn13_ps_oracle_rescale_guard_candidate_validated"
		result.FirstLimiting = "none"
	} else {
		result.FirstLimiting = psRescaleGuardLimitingBoundary(best)
		result.Classification = "logn13_ps_rescale_guard_insufficient_precision"
		validGuard := false
		for _, candidate := range allSeen {
			if candidate.TotalGuardBits > 0 && candidate.Valid {
				validGuard = true
			}
		}
		if !validGuard {
			for _, candidate := range allSeen {
				if candidate.TotalGuardBits == 0 {
					continue
				}
				if !candidate.CapacitySafe || !candidate.ValuePreserving {
					result.Classification = "logn13_ps_rescale_guard_blocked_by_q01_capacity"
					break
				}
				if !candidate.AlignmentSafe {
					result.Classification = "logn13_ps_rescale_guard_blocked_by_scale_alignment"
				}
			}
		}
		if !validGuard && result.Classification == "logn13_ps_rescale_guard_insufficient_precision" {
			for _, candidate := range allSeen {
				if candidate.TotalGuardBits > 0 && !candidate.AlignmentSafe {
					result.Classification = "logn13_ps_rescale_guard_blocked_by_scale_alignment"
					break
				}
			}
		}
	}
	result.Validation["r0_pass"] = r0Pass
	result.Validation["boundary_count"] = len(boundaries)
	result.Validation["greedy_action_limit"] = 8
	result.Validation["max_guard_bits_per_boundary"] = 3
	return psRescaleGuardWrite(result, outPath)
}
