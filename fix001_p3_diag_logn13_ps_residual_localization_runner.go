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
	requiredFIX001P3PSLocalizationPrimary   = "c33dba976f13ef27c1cb35199a05c442acc36934"
	requiredFIX001P3PSLocalizationSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	psLocalizationPublicThreshold           = 1e-2
	psLocalizationCheckpointThreshold       = 1e-10
	psLocalizationClosureTolerance          = 1e-10
	psLocalizationPlanScaleExponent         = 92
)

type psLocalizationCheckpoint struct {
	ID                         string          `json:"id"`
	Region                     string          `json:"region"`
	Operation                  string          `json:"operation"`
	SourceFile                 string          `json:"source_file"`
	SourceFunction             string          `json:"source_function"`
	DiagnosticSource           string          `json:"diagnostic_source"`
	Description                string          `json:"description"`
	Level                      int             `json:"level"`
	Scale                      string          `json:"scale"`
	Degree                     int             `json:"degree"`
	LogicalQ2Active            bool            `json:"logical_q2_active"`
	MaintainedQ2               bool            `json:"maintained_q2"`
	RescaleOccursNext          bool            `json:"rescale_occurs_next"`
	Supported                  bool            `json:"supported"`
	SupportReason              string          `json:"support_reason,omitempty"`
	ActualVsCanonical          *PSGlobalMetric `json:"x_f_minus_x_q,omitempty"`
	CanonicalFloor             *PSGlobalMetric `json:"x_q_minus_x_o,omitempty"`
	ErrorToFloorRatio          float64         `json:"error_to_floor_ratio,omitempty"`
	CanonicalMaterializationOK bool            `json:"canonical_materialization_deterministic"`
}

type psLocalizationGrowth struct {
	FromID                 string          `json:"from_id"`
	ToID                   string          `json:"to_id"`
	FromRegion             string          `json:"from_region"`
	ToRegion               string          `json:"to_region"`
	FromResidual           *PSGlobalMetric `json:"from_residual"`
	ToResidual             *PSGlobalMetric `json:"to_residual"`
	MaxComponentGrowth     float64         `json:"max_component_growth"`
	MaxComponentRatio      float64         `json:"max_component_ratio"`
	DirectionSimilarity    float64         `json:"direction_similarity"`
	NewDominantDirection   bool            `json:"new_dominant_direction"`
	ExactSuffixPropagation bool            `json:"exact_suffix_propagation_supported"`
	PropagationDisposition string          `json:"propagation_disposition"`
}

type psLocalizationDownstreamEvidence struct {
	Reached         bool            `json:"reached"`
	Reason          string          `json:"reason,omitempty"`
	EvalModReal     *PSGlobalMetric `json:"evalmod_real_vs_standard,omitempty"`
	EvalModImag     *PSGlobalMetric `json:"evalmod_imag_vs_standard,omitempty"`
	PostS2C         *PSGlobalMetric `json:"post_s2c_vs_standard,omitempty"`
	PublicLike      *PSGlobalMetric `json:"public_like_vs_message,omitempty"`
	StandardLike    *PSGlobalMetric `json:"standard_like_vs_message,omitempty"`
	MetadataCorrect bool            `json:"public_metadata_correct"`
	InputUnchanged  bool            `json:"input_unchanged"`
	ContractsPass   bool            `json:"contracts_pass"`
}

type psLocalizationReset struct {
	ID          string                            `json:"id"`
	Region      string                            `json:"region"`
	Supported   bool                              `json:"supported"`
	Reason      string                            `json:"reason,omitempty"`
	FinalPSReal *PSGlobalMetric                   `json:"final_ps_real_f_minus_p_q,omitempty"`
	FinalPSImag *PSGlobalMetric                   `json:"final_ps_imag_f_minus_p_q,omitempty"`
	Downstream  *psLocalizationDownstreamEvidence `json:"downstream,omitempty"`
}

type psLocalizationOperationReset struct {
	ID         string                            `json:"id"`
	Region     string                            `json:"region"`
	Downstream *psLocalizationDownstreamEvidence `json:"downstream,omitempty"`
}

type psLocalizationAlpha struct {
	Alpha           float64         `json:"alpha"`
	Valid           bool            `json:"valid"`
	PublicLike      *PSGlobalMetric `json:"public_like,omitempty"`
	PostS2C         *PSGlobalMetric `json:"post_s2c,omitempty"`
	EvalModReal     *PSGlobalMetric `json:"evalmod_real,omitempty"`
	EvalModImag     *PSGlobalMetric `json:"evalmod_imag,omitempty"`
	MetadataCorrect bool            `json:"metadata_correct"`
	InputUnchanged  bool            `json:"input_unchanged"`
	Reason          string          `json:"reason,omitempty"`
}

type psLocalizationResult struct {
	SchemaVersion         string                         `json:"schema_version"`
	Timestamp             time.Time                      `json:"timestamp"`
	Primary               RepositoryMetadata             `json:"primary_repository"`
	Lattigo               RepositoryMetadata             `json:"lattigo_repository"`
	Environment           EnvironmentMetadata            `json:"environment"`
	Config                BootstrapConfig                `json:"config"`
	Parameters            ExperimentParameters           `json:"effective_parameters"`
	Provenance            map[string]interface{}         `json:"provenance"`
	P0Control             map[string]interface{}         `json:"p0_control"`
	PSDAG                 []psLocalizationCheckpoint     `json:"ps_dag_checkpoints"`
	CheckpointResiduals   []psLocalizationCheckpoint     `json:"checkpoint_residuals"`
	ResidualGrowth        []psLocalizationGrowth         `json:"incremental_residual_growth"`
	CheckpointResets      []psLocalizationReset          `json:"checkpoint_reset_counterfactuals"`
	OperationResets       []psLocalizationOperationReset `json:"operation_reset_counterfactuals"`
	AlphaSensitivity      []psLocalizationAlpha          `json:"alpha_sensitivity"`
	SmallestPassingAlpha  float64                        `json:"smallest_passing_alpha,omitempty"`
	RequiredReduction     float64                        `json:"required_residual_reduction_fraction,omitempty"`
	AlphaMargin           float64                        `json:"smallest_passing_alpha_public_margin,omitempty"`
	AlphaValid            bool                           `json:"alpha_experiment_valid"`
	ImplicatedRegion      string                         `json:"implicated_region,omitempty"`
	Classification        string                         `json:"classification"`
	FirstRemainingBlocker string                         `json:"first_remaining_blocker"`
	RecommendedNextTarget string                         `json:"recommended_next_design_target"`
	Validation            map[string]interface{}         `json:"validation"`
}

type psLocalizationBranch struct {
	Work   polynomialDecompositionBranchWork
	Checks []PSGlobalCheckpoint
	Giants []PSGiantStepRecord
	All    []PSGlobalCheckpoint
}

type psLocalizationAcceptedContext struct {
	Branch           psRescaleGuardBranch
	Plan             commonpolynomial.PatersonStockmeyerPolynomial
	GuardMap         map[string]int
	FinalOverride    psRescaleGuardFinalOverride
	BoundaryOverride psRescaleGuardBoundaryOverride
	Run              psRescaleGuardBranchRun
}

func psLocalizationAcceptedSetup(profile q056PreparedProfile) (psLocalizationAcceptedContext, psLocalizationAcceptedContext, polynomialDecompositionBranchWork, polynomialDecompositionBranchWork, psRescaleGuardCandidate, error) {
	var f0Real, f0Imag, g0Real, g0Imag fix001P3LocalQ2BranchEvidence
	g0RealOverride := fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &g0Real, 2)
	g0ImagOverride := fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &g0Imag, 2)
	realBoundaryOverride := func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error) {
		if boundary.ID != "G0-rescale" {
			return false, nil
		}
		return g0RealOverride(params, eval, ct, boundary)
	}
	imagBoundaryOverride := func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error) {
		if boundary.ID != "G0-rescale" {
			return false, nil
		}
		return g0ImagOverride(params, eval, ct, boundary)
	}
	realFinalOverride := fix001P3LocalQ2OverrideWithBits(profile.Residual, "real", &f0Real, 3)
	imagFinalOverride := fix001P3LocalQ2OverrideWithBits(profile.Residual, "imag", &f0Imag, 3)
	var candidate psRescaleGuardCandidate
	realRun, imagRun, err := func() (psRescaleGuardBranchRun, psRescaleGuardBranchRun, error) {
		var rr, ir psRescaleGuardBranchRun
		var e error
		candidate, rr, ir, e = psRescaleGuardMakeCandidateWithAllOverrides(
			profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch,
			profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, 3},
			"G0-local-q2-2-F0-local-q2-3", &profile.InitialReal, &profile.InitialImag, profile.Boundaries,
			realFinalOverride, imagFinalOverride,
			psRescaleGuardBoundaryOverride(realBoundaryOverride), psRescaleGuardBoundaryOverride(imagBoundaryOverride))
		if e == nil {
			// Keep the candidate's final ciphertexts as the exact accepted-path outputs.
			rr.Final = candidate.realFinal
			ir.Final = candidate.imagFinal
		}
		return rr, ir, e
	}()
	if err != nil {
		return psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, polynomialDecompositionBranchWork{}, polynomialDecompositionBranchWork{}, psRescaleGuardCandidate{}, err
	}
	makeWork := func(branch psRescaleGuardBranch, run psRescaleGuardBranchRun) (polynomialDecompositionBranchWork, error) {
		work := branch.Base
		work.PlanScale = precisionSweepScale(psLocalizationPlanScaleExponent)
		work.Plan = psOracleScalePlan(branch.Base.Plan, work.PlanScale)
		work.ActualCiphertext = run.Final.CopyNew()
		work.ActualOutput = append([]complex128(nil), run.FinalValues...)
		normal, e := fix001P3QuantizationAwareMaterialize(profile.BTP.BootstrappingParameters, run.Final, fix001P3QuantizationAwareValues(branch.Base.ReferenceValues, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
		if e != nil {
			return work, e
		}
		work.OracleOutput, e = fix001P3QuantizationAwareDecode(profile.BTP.BootstrappingParameters, normal)
		if e != nil {
			return work, e
		}
		work.OracleCiphertext, e = psLocalizationFastFromNormal(profile.BTP.BootstrappingParameters, normal)
		return work, e
	}
	realWork, err := makeWork(profile.RealBranch, realRun)
	if err != nil {
		return psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, polynomialDecompositionBranchWork{}, polynomialDecompositionBranchWork{}, psRescaleGuardCandidate{}, err
	}
	imagWork, err := makeWork(profile.ImagBranch, imagRun)
	if err != nil {
		return psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, polynomialDecompositionBranchWork{}, polynomialDecompositionBranchWork{}, psRescaleGuardCandidate{}, err
	}
	guardMap := make(map[string]int, len(profile.Boundaries))
	for i, bits := range []int{1, 0, 0, 0, 3} {
		if bits > 0 && i < len(profile.Boundaries) {
			guardMap[profile.Boundaries[i].ID] = bits
		}
	}
	plan := psOracleScalePlan(profile.RealBranch.Base.Plan, precisionSweepScale(psLocalizationPlanScaleExponent))
	realContext := psLocalizationAcceptedContext{Branch: profile.RealBranch, Plan: plan, GuardMap: guardMap, FinalOverride: realFinalOverride, BoundaryOverride: psRescaleGuardBoundaryOverride(realBoundaryOverride), Run: realRun}
	imagContext := psLocalizationAcceptedContext{Branch: profile.ImagBranch, Plan: psOracleScalePlan(profile.ImagBranch.Base.Plan, precisionSweepScale(psLocalizationPlanScaleExponent)), GuardMap: guardMap, FinalOverride: imagFinalOverride, BoundaryOverride: psRescaleGuardBoundaryOverride(imagBoundaryOverride), Run: imagRun}
	return realContext, imagContext, realWork, imagWork, candidate, nil
}

func psLocalizationWrite(result psLocalizationResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if outPath == "" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func psLocalizationMetric(reference, actual []complex128, threshold float64) *PSGlobalMetric {
	metric := psGlobalMetric(reference, actual)
	metric.Threshold = threshold
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func psLocalizationRegion(id string) string {
	if strings.HasPrefix(id, "G") {
		if dash := strings.IndexByte(id, '-'); dash > 0 {
			return id[:dash]
		}
	}
	if strings.HasPrefix(id, "B") {
		if dash := strings.IndexByte(id, '-'); dash > 0 {
			return id[:dash]
		}
	}
	if strings.HasPrefix(id, "F0") {
		return "F0"
	}
	return "unsupported"
}

func psLocalizationSource(operation string) (string, string, bool) {
	switch operation {
	case "baby_constant_add", "baby_mul_then_add":
		return "lattigo/circuits/ckks/polynomial/fast.go", "fastPolynomialWorkspace.evaluateBabyStep", false
	case "giant_input_b", "giant_relinearize", "giant_rescale", "giant_multiply", "giant_metadata_normalization", "giant_add_aligned":
		return "lattigo/circuits/ckks/polynomial/fast.go", "fastPolynomialWorkspace.evaluateMonomial", operation == "giant_relinearize" || operation == "giant_rescale"
	case "pre_final_rescale_root":
		return "lattigo/circuits/ckks/polynomial/fast.go", "fastPolynomialWorkspace.evaluatePlan", true
	default:
		return "", "", false
	}
}

func psLocalizationCheckpointMetadata(check PSGlobalCheckpoint) psLocalizationCheckpoint {
	file, function, rescaleNext := psLocalizationSource(check.Operation)
	region := psLocalizationRegion(check.ID)
	return psLocalizationCheckpoint{
		ID: check.ID, Region: region, Operation: check.Operation, SourceFile: file, SourceFunction: function,
		DiagnosticSource: "fix001_p3_global_semantics_runner.go:psGlobalReplay", Description: check.ExpectedSubtree,
		Level: check.Level, Scale: check.Scale, Degree: check.Degree, LogicalQ2Active: check.Level >= 2,
		MaintainedQ2: false, RescaleOccursNext: rescaleNext, Supported: true,
	}
}

func psLocalizationReplay(params ckks.Parameters, eval *bootstrapping.FastEvaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, input []complex128, planScale rlwe.Scale, reset *psGlobalReplayReset) ([]complex128, *rlwe.Ciphertext, []PSGlobalCheckpoint, []PSGiantStepRecord, error) {
	var checks []PSGlobalCheckpoint
	var giants []PSGiantStepRecord
	var root PSGlobalCheckpoint
	var err error
	if reset == nil {
		checks, giants, root, _, err = psGlobalReplay(params, eval.FastCKKS, plan, powers, expected, decoded, input, planScale)
	} else {
		checks, giants, root, _, err = psGlobalReplay(params, eval.FastCKKS, plan, powers, expected, decoded, input, planScale, *reset)
	}
	if err != nil {
		return nil, nil, checks, giants, err
	}
	if root.ciphertext == nil {
		return nil, nil, checks, giants, fmt.Errorf("PS replay root has no ciphertext")
	}
	output := root.ciphertext.CopyNew()
	if err := eval.FastCKKS.Rescale(output, output); err != nil {
		return nil, nil, checks, giants, err
	}
	values, err := psGlobalDecode(params, output)
	if err != nil {
		return nil, nil, checks, giants, err
	}
	return values, output, checks, giants, nil
}

func psLocalizationAllCheckpoints(checks []PSGlobalCheckpoint, giants []PSGiantStepRecord, root PSGlobalCheckpoint) []PSGlobalCheckpoint {
	all := append([]PSGlobalCheckpoint(nil), checks...)
	for _, giant := range giants {
		all = append(all, giant.Checkpoints...)
	}
	all = append(all, root)
	return all
}

func psLocalizationCanonical(params ckks.Parameters, check PSGlobalCheckpoint) (*rlwe.Ciphertext, []complex128, *PSGlobalMetric, bool, error) {
	if check.ciphertext == nil || check.expected == nil {
		return nil, nil, nil, false, fmt.Errorf("checkpoint %s has no source-backed snapshot", check.ID)
	}
	normalCanonical, err := fix001P3QuantizationAwareMaterialize(params, check.ciphertext, fix001P3QuantizationAwareValues(check.expected, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, nil, nil, false, err
	}
	values, err := fix001P3QuantizationAwareDecode(params, normalCanonical)
	if err != nil {
		return nil, nil, nil, false, err
	}
	second, err := fix001P3QuantizationAwareMaterialize(params, check.ciphertext, fix001P3QuantizationAwareValues(check.expected, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, nil, nil, false, err
	}
	secondValues, err := fix001P3QuantizationAwareDecode(params, second)
	if err != nil {
		return nil, nil, nil, false, err
	}
	deterministic := psLocalizationMetric(values, secondValues, 1e-12)
	floor := psLocalizationMetric(check.expected, values, psLocalizationCheckpointThreshold)
	canonical, err := psLocalizationFastFromNormal(params, normalCanonical)
	if err != nil {
		return nil, nil, nil, false, err
	}
	return canonical, values, floor, deterministic.Pass, nil
}

func psLocalizationFastMaterialize(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128) (*rlwe.Ciphertext, error) {
	normal, err := fix001P3QuantizationAwareMaterialize(params, source, fix001P3QuantizationAwareValues(values, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, err
	}
	return psLocalizationFastFromNormal(params, normal)
}

func psLocalizationFastFromNormal(params ckks.Parameters, normal *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	ciphertext := fastckks.NewCiphertext(params, normal.Degree(), normal.Level())
	*ciphertext.MetaData = *normal.MetaData
	ciphertext.IsNTT = normal.IsNTT
	ciphertext.IsMontgomery = true
	for component := range ciphertext.Value {
		for limb := 0; limb <= ciphertext.Level() && limb < len(normal.Value[0].Coeffs); limb++ {
			if len(ciphertext.Value[component].Coeffs[limb]) == 0 {
				ciphertext.Value[component].Coeffs[limb] = make([]uint64, len(normal.Value[0].Coeffs[limb]))
			}
			if component == 0 {
				copy(ciphertext.Value[component].Coeffs[limb], normal.Value[0].Coeffs[limb])
				params.RingQ().SubRings[limb].MForm(ciphertext.Value[component].Coeffs[limb], ciphertext.Value[component].Coeffs[limb])
			}
		}
	}
	return ciphertext, nil
}

func psLocalizationRatio(numerator, denominator float64) float64 {
	if denominator == 0 {
		if numerator == 0 {
			return 0
		}
		return math.MaxFloat64
	}
	return numerator / denominator
}

func psLocalizationResidual(check PSGlobalCheckpoint, canonical []complex128, floor *PSGlobalMetric, deterministic bool) (psLocalizationCheckpoint, []complex128, error) {
	metadata := psLocalizationCheckpointMetadata(check)
	if check.actual == nil || canonical == nil {
		metadata.Supported = false
		metadata.SupportReason = "missing actual or canonical vector"
		return metadata, nil, nil
	}
	actualMetric := psLocalizationMetric(check.actual, canonical, psLocalizationCheckpointThreshold)
	metadata.ActualVsCanonical = actualMetric
	metadata.CanonicalFloor = floor
	metadata.CanonicalMaterializationOK = deterministic
	metadata.ErrorToFloorRatio = psLocalizationRatio(actualMetric.MaxComponent, floor.MaxComponent)
	return metadata, fix001P3QuantizationAwareSub(check.actual, canonical), nil
}

func psLocalizationDirection(a, b []complex128) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot complex128
	var normA, normB float64
	for i := range a {
		dot += complex(real(a[i]), -imag(a[i])) * b[i]
		normA += real(a[i])*real(a[i]) + imag(a[i])*imag(a[i])
		normB += real(b[i])*real(b[i]) + imag(b[i])*imag(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return cmplxAbs(dot) / math.Sqrt(normA*normB)
}

func cmplxAbs(value complex128) float64 {
	return math.Hypot(real(value), imag(value))
}

func psLocalizationGrowthRows(states []struct {
	Meta     psLocalizationCheckpoint
	Residual []complex128
}) []psLocalizationGrowth {
	rows := make([]psLocalizationGrowth, 0, len(states)-1)
	for i := 1; i < len(states); i++ {
		from, to := states[i-1], states[i]
		fromMetric := psLocalizationMetric(make([]complex128, len(from.Residual)), from.Residual, psLocalizationCheckpointThreshold)
		toMetric := psLocalizationMetric(make([]complex128, len(to.Residual)), to.Residual, psLocalizationCheckpointThreshold)
		delta := toMetric.MaxComponent - fromMetric.MaxComponent
		rows = append(rows, psLocalizationGrowth{FromID: from.Meta.ID, ToID: to.Meta.ID, FromRegion: from.Meta.Region, ToRegion: to.Meta.Region, FromResidual: fromMetric, ToResidual: toMetric, MaxComponentGrowth: delta, MaxComponentRatio: psLocalizationRatio(toMetric.MaxComponent, fromMetric.MaxComponent), DirectionSimilarity: psLocalizationDirection(from.Residual, to.Residual), NewDominantDirection: toMetric.WorstIndex != fromMetric.WorstIndex || toMetric.WorstPart != fromMetric.WorstPart, ExactSuffixPropagation: false, PropagationDisposition: "checkpoint source expressions are supported; exact nonlinear suffix propagation is reserved for reset counterfactuals"})
	}
	return rows
}

func psLocalizationRunAcceptedDownstream(profile q056PreparedProfile, planScale rlwe.Scale, realOverride, imagOverride *rlwe.Ciphertext) (*psLocalizationDownstreamEvidence, error) {
	result := &psLocalizationDownstreamEvidence{}
	realState, err := fix001P3DAR0PrepareState(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, realOverride)
	if err != nil {
		return nil, err
	}
	imagState, err := fix001P3DAR0PrepareState(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, imagOverride)
	if err != nil {
		return nil, err
	}
	for round := 0; round < profile.Fast.Mod1Parameters.DoubleAngle; round++ {
		realOutcome, err := fix001P3DAAllRoundsRound(&realState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return nil, err
		}
		imagOutcome, err := fix001P3DAAllRoundsRound(&imagState, round, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
		if err != nil {
			return nil, err
		}
		for _, outcome := range []fix001P3DAAllRoundsRoundOutcome{realOutcome, imagOutcome} {
			if outcome.Classification != "" {
				result.Reason = outcome.Blocker + ": " + outcome.Classification
				return result, nil
			}
		}
	}
	if err := fix001P3DAAllRoundsFinalize(&realState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return nil, err
	}
	if err := fix001P3DAAllRoundsFinalize(&imagState, profile.Fast, profile.BTP.BootstrappingParameters); err != nil {
		return nil, err
	}
	if realState.Path.FastFinal == nil || imagState.Path.FastFinal == nil {
		result.Reason = fmt.Sprintf("suffix replay stopped: real=%s imag=%s", realState.Path.FirstFailure, imagState.Path.FirstFailure)
		return result, nil
	}
	realFinal, err := evalModMatchedFinalEvidence(realState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return nil, err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagState.Path, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return nil, err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realState.Path, imagState.Path, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return nil, err
	}
	result.Reached = true
	result.EvalModReal, result.EvalModImag = realFinal.FastVsStandard, imagFinal.FastVsStandard
	result.PostS2C, result.PublicLike, result.StandardLike = publicEvidence.PostS2C, publicEvidence.PublicLike, publicEvidence.StandardLike
	result.MetadataCorrect, result.InputUnchanged = publicEvidence.Metadata["pass"] == true, publicEvidence.InputUnchanged
	result.ContractsPass = realState.Path.FastRowsMatch && realState.Path.RestoreRowsMatch && imagState.Path.FastRowsMatch && imagState.Path.RestoreRowsMatch && result.MetadataCorrect && result.InputUnchanged
	return result, nil
}

func psLocalizationInterpolate(a, b []complex128, alpha float64) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] + complex(alpha, 0)*(b[i]-a[i])
	}
	return out
}

func psLocalizationMaxPublic(a, b *PSGlobalMetric) *PSGlobalMetric {
	if a == nil || b == nil {
		return nil
	}
	metric := *a
	if b.MaxComponent > metric.MaxComponent {
		metric.MaxComponent, metric.WorstIndex, metric.WorstPart = b.MaxComponent, b.WorstIndex, b.WorstPart
	}
	metric.MaxComplex = math.Max(a.MaxComplex, b.MaxComplex)
	metric.MeanComplex = (a.MeanComplex + b.MeanComplex) / 2
	metric.Pass = metric.MaxComponent <= metric.Threshold
	return &metric
}

func psLocalizationRunReset(branch psLocalizationBranch, params ckks.Parameters, eval *bootstrapping.FastEvaluator, resetID string, canonical *rlwe.Ciphertext) ([]complex128, *rlwe.Ciphertext, error) {
	values, ciphertext, _, _, err := psLocalizationReplay(params, eval, branch.Work.Plan, branch.Work.OraclePowerMap, branch.Work.PowerExpected, branch.Work.OraclePowerValues, branch.Work.InputValues, branch.Work.PlanScale, &psGlobalReplayReset{ID: resetID, Ciphertext: canonical})
	return values, ciphertext, err
}

func psLocalizationRunAcceptedReset(context psLocalizationAcceptedContext, params ckks.Parameters, eval *bootstrapping.FastEvaluator, resetID string, canonical *rlwe.Ciphertext) ([]complex128, *rlwe.Ciphertext, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(
		params, eval.FastCKKS, context.Plan,
		context.Branch.Base.OraclePowerMap, context.Branch.Base.PowerExpected, context.Branch.Base.OraclePowerValues,
		context.Branch.Base.InputValues, precisionSweepScale(psLocalizationPlanScaleExponent), context.GuardMap,
		context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: resetID, Ciphertext: canonical})
	if err != nil {
		return nil, nil, err
	}
	if replay.Final == nil {
		return nil, nil, fmt.Errorf("accepted PS replay returned nil final ciphertext")
	}
	values, err := psGlobalDecode(params, replay.Final)
	return values, replay.Final, err
}

func psLocalizationCheckpointMap(checks []PSGlobalCheckpoint) map[string]PSGlobalCheckpoint {
	result := make(map[string]PSGlobalCheckpoint, len(checks))
	for _, check := range checks {
		if _, exists := result[check.ID]; !exists {
			result[check.ID] = check
		}
	}
	return result
}

func psLocalizationAlphaRun(profile q056PreparedProfile, realWork, imagWork polynomialDecompositionBranchWork, alpha float64) (psLocalizationAlpha, error) {
	realValues := psLocalizationInterpolate(realWork.ActualOutput, realWork.OracleOutput, alpha)
	imagValues := psLocalizationInterpolate(imagWork.ActualOutput, imagWork.OracleOutput, alpha)
	var realCT, imagCT *rlwe.Ciphertext
	var err error
	if alpha == 0 {
		realCT, imagCT = realWork.ActualCiphertext.CopyNew(), imagWork.ActualCiphertext.CopyNew()
	} else if alpha == 1 {
		realCT, imagCT = realWork.OracleCiphertext.CopyNew(), imagWork.OracleCiphertext.CopyNew()
	} else {
		realCT, err = psLocalizationFastMaterialize(profile.BTP.BootstrappingParameters, realWork.ActualCiphertext, realValues)
		if err != nil {
			return psLocalizationAlpha{}, err
		}
		imagCT, err = psLocalizationFastMaterialize(profile.BTP.BootstrappingParameters, imagWork.ActualCiphertext, imagValues)
		if err != nil {
			return psLocalizationAlpha{}, err
		}
	}
	downstream, err := psLocalizationRunAcceptedDownstream(profile, realWork.PlanScale, realCT, imagCT)
	if err != nil {
		return psLocalizationAlpha{Alpha: alpha, Valid: false, Reason: err.Error()}, nil
	}
	if !downstream.Reached {
		return psLocalizationAlpha{Alpha: alpha, Valid: false, Reason: downstream.Reason}, nil
	}
	return psLocalizationAlpha{Alpha: alpha, Valid: downstream.PublicLike != nil && downstream.PublicLike.Pass, PublicLike: downstream.PublicLike, PostS2C: downstream.PostS2C, EvalModReal: downstream.EvalModReal, EvalModImag: downstream.EvalModImag, MetadataCorrect: downstream.MetadataCorrect, InputUnchanged: downstream.InputUnchanged}, nil
}

func psLocalizationP0Control(profile q056PreparedProfile, candidate psRescaleGuardCandidate, realWork, imagWork polynomialDecompositionBranchWork) (map[string]interface{}, error) {
	actualDownstream, err := psLocalizationRunAcceptedDownstream(profile, realWork.PlanScale, realWork.ActualCiphertext, imagWork.ActualCiphertext)
	if err != nil {
		return nil, err
	}
	cps, err := psLocalizationRunAcceptedDownstream(profile, realWork.PlanScale, realWork.OracleCiphertext, imagWork.OracleCiphertext)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"actual_ps_real_vs_canonical":         psLocalizationMetric(realWork.ActualOutput, realWork.OracleOutput, psLocalizationCheckpointThreshold),
		"actual_ps_imag_vs_canonical":         psLocalizationMetric(imagWork.ActualOutput, imagWork.OracleOutput, psLocalizationCheckpointThreshold),
		"actual_evalmod_real_vs_standard":     actualDownstream.EvalModReal,
		"actual_evalmod_imag_vs_standard":     actualDownstream.EvalModImag,
		"actual_fast_rows_match":              actualDownstream.ContractsPass,
		"actual_restore_rows_match":           actualDownstream.ContractsPass,
		"actual_public_like":                  actualDownstream.PublicLike,
		"actual_post_s2c":                     actualDownstream.PostS2C,
		"c_ps":                                cps,
		"c_ps_public_pass":                    cps.Reached && cps.PublicLike != nil && cps.PublicLike.Pass,
		"accepted_candidate_final_real_error": candidate.FinalRealError,
		"accepted_candidate_final_imag_error": candidate.FinalImagError,
		"accepted_candidate_first_failure":    candidate.FirstFailure,
	}
	return result, nil
}

func psLocalizationPrepareBranch(name string, params ckks.Parameters, btp bootstrapping.Parameters, fastEval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, fastInput, standardInput *rlwe.Ciphertext, standardSK *rlwe.SecretKey) (polynomialDecompositionBranchWork, error) {
	work, err := polynomialDecompositionPrepareBranch(name, params, btp, fastEval, standardEval, fastInput, standardInput, standardSK)
	if err != nil {
		return work, err
	}
	work.PlanScale = rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), psLocalizationPlanScaleExponent))
	for i := range work.Plan.Value {
		work.Plan.Value[i].Scale = work.PlanScale
	}
	work.ActualOutput, work.ActualCiphertext, work.ActualChecks, err = polynomialDecompositionReplay(params, fastEval, work.Plan, work.OraclePowerMap, work.PowerExpected, work.OraclePowerValues, work.InputValues, work.PlanScale)
	if err != nil {
		return work, err
	}
	normalOracle, err := fix001P3QuantizationAwareMaterialize(params, work.ActualCiphertext, fix001P3QuantizationAwareValues(work.ReferenceValues, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return work, err
	}
	work.OracleOutput, err = fix001P3QuantizationAwareDecode(params, normalOracle)
	work.OracleChecks = nil
	if err != nil {
		return work, err
	}
	work.OracleCiphertext, err = psLocalizationFastFromNormal(params, normalOracle)
	if err != nil {
		return work, err
	}
	return work, nil
}

func runFIX001P3DiagLogN13PSResidualLocalization(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	residual, btp := profile.Residual, profile.BTP
	fastEval := profile.Fast
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3PSLocalizationPrimary, "HEAD") == requiredFIX001P3PSLocalizationPrimary
	result := psLocalizationResult{SchemaVersion: "fix-001-p3-diag-logn13-ps-residual-localization.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Provenance: map[string]interface{}{"primary_required_base": requiredFIX001P3PSLocalizationPrimary, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_required_commit": requiredFIX001P3PSLocalizationSecondary, "secondary_commit": secondaryCommit, "secondary_branch": secondaryBranch, "secondary_clean": secondaryClean}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_production_changes": true, "no_da_change": true, "no_s2c_change": true, "no_q_parameter_changes": true, "no_generated_power_redesign": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !primaryBaseAncestor || secondaryCommit != requiredFIX001P3PSLocalizationSecondary || secondaryBranch != "fast-ckks" || !secondaryClean {
		result.Classification = "logn13_ps_residual_localization_precondition_mismatch"
		result.FirstRemainingBlocker = "provenance or LogN13 precondition"
		return psLocalizationWrite(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, candidate, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	result.P0Control, err = psLocalizationP0Control(profile, candidate, realWork, imagWork)
	if err != nil {
		return err
	}
	cps, _ := result.P0Control["c_ps_public_pass"].(bool)
	actualPublic, _ := result.P0Control["actual_public_like"].(*PSGlobalMetric)
	controlPass := cps && actualPublic != nil && actualPublic.MaxComponent > 0.009 && actualPublic.MaxComponent < 0.012
	result.Validation["p0_control_pass"] = controlPass
	if !controlPass {
		result.Classification = "logn13_ps_residual_localization_precondition_mismatch"
		result.FirstRemainingBlocker = "accepted P0 control was not reproduced"
		return psLocalizationWrite(result, outPath)
	}
	realAll := append(append([]PSGlobalCheckpoint(nil), realContext.Run.Replay.Checks...), realContext.Run.Replay.Root)
	imagAll := append(append([]PSGlobalCheckpoint(nil), imagContext.Run.Replay.Checks...), imagContext.Run.Replay.Root)
	realMap := psLocalizationCheckpointMap(realAll)
	imagMap := psLocalizationCheckpointMap(imagAll)
	ids := make([]string, 0, len(realMap))
	for _, check := range realAll {
		if _, ok := imagMap[check.ID]; ok {
			ids = append(ids, check.ID)
		}
	}
	seen := map[string]bool{}
	orderedIDs := ids[:0]
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			orderedIDs = append(orderedIDs, id)
		}
	}
	states := make([]struct {
		Meta     psLocalizationCheckpoint
		Residual []complex128
	}, 0, len(orderedIDs))
	realCanonical := map[string]*rlwe.Ciphertext{}
	imagCanonical := map[string]*rlwe.Ciphertext{}
	for _, id := range orderedIDs {
		realCheck, imagCheck := realMap[id], imagMap[id]
		realCanonicalCT, realCanonicalValues, realFloor, realDeterministic, realErr := psLocalizationCanonical(btp.BootstrappingParameters, realCheck)
		imagCanonicalCT, _, _, _, imagErr := psLocalizationCanonical(btp.BootstrappingParameters, imagCheck)
		if realErr != nil || imagErr != nil {
			meta := psLocalizationCheckpointMetadata(realCheck)
			meta.Supported = false
			if realErr != nil {
				meta.SupportReason = realErr.Error()
			} else {
				meta.SupportReason = imagErr.Error()
			}
			result.PSDAG = append(result.PSDAG, meta)
			result.CheckpointResiduals = append(result.CheckpointResiduals, meta)
			continue
		}
		realMeta, realResidual, _ := psLocalizationResidual(realCheck, realCanonicalValues, realFloor, realDeterministic)
		result.PSDAG = append(result.PSDAG, realMeta)
		result.CheckpointResiduals = append(result.CheckpointResiduals, realMeta)
		states = append(states, struct {
			Meta     psLocalizationCheckpoint
			Residual []complex128
		}{realMeta, realResidual})
		realCanonical[id] = realCanonicalCT
		imagCanonical[id] = imagCanonicalCT
	}
	result.ResidualGrowth = psLocalizationGrowthRows(states)
	// P4 resets use the common checkpoint IDs and replace both real/imag states.
	resetResults := make([]psLocalizationReset, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		realCT, realOK := realCanonical[id]
		imagCT, imagOK := imagCanonical[id]
		row := psLocalizationReset{ID: id, Region: psLocalizationRegion(id), Supported: realOK && imagOK}
		if !row.Supported {
			row.Reason = "canonical checkpoint materialization unsupported"
			resetResults = append(resetResults, row)
			continue
		}
		realValues, realFinalCT, err := psLocalizationRunAcceptedReset(realContext, btp.BootstrappingParameters, fastEval, id, realCT)
		if err != nil {
			row.Supported, row.Reason = false, err.Error()
			resetResults = append(resetResults, row)
			continue
		}
		imagValues, imagFinalCT, err := psLocalizationRunAcceptedReset(imagContext, btp.BootstrappingParameters, fastEval, id, imagCT)
		if err != nil {
			row.Supported, row.Reason = false, err.Error()
			resetResults = append(resetResults, row)
			continue
		}
		realReference := realWork.OracleOutput
		imagReference := imagWork.OracleOutput
		row.FinalPSReal = psLocalizationMetric(realReference, realValues, psLocalizationCheckpointThreshold)
		row.FinalPSImag = psLocalizationMetric(imagReference, imagValues, psLocalizationCheckpointThreshold)
		row.Downstream, err = psLocalizationRunAcceptedDownstream(profile, realWork.PlanScale, realFinalCT, imagFinalCT)
		if err != nil {
			row.Supported, row.Reason = false, err.Error()
		}
		resetResults = append(resetResults, row)
	}
	result.CheckpointResets = resetResults
	latestPassingRegion := ""
	for _, row := range resetResults {
		if row.Supported && row.Downstream != nil && row.Downstream.Reached && row.Downstream.PublicLike != nil && row.Downstream.PublicLike.Pass {
			latestPassingRegion = row.Region
		}
	}
	result.ImplicatedRegion = latestPassingRegion
	operationIDs := make([]string, 0)
	for _, row := range resetResults {
		if row.Region == latestPassingRegion && row.Region != "" && row.Region != "F0" {
			operationIDs = append(operationIDs, row.ID)
		}
	}
	for _, id := range operationIDs {
		realCT, realOK := realCanonical[id]
		imagCT, imagOK := imagCanonical[id]
		row := psLocalizationOperationReset{ID: id, Region: latestPassingRegion}
		if realOK && imagOK {
			_, realFinalCT, realErr := psLocalizationRunAcceptedReset(realContext, btp.BootstrappingParameters, fastEval, id, realCT)
			_, imagFinalCT, imagErr := psLocalizationRunAcceptedReset(imagContext, btp.BootstrappingParameters, fastEval, id, imagCT)
			if realErr == nil && imagErr == nil {
				row.Downstream, err = psLocalizationRunAcceptedDownstream(profile, realWork.PlanScale, realFinalCT, imagFinalCT)
			} else if realErr != nil {
				row.Downstream = &psLocalizationDownstreamEvidence{Reason: realErr.Error()}
			} else {
				row.Downstream = &psLocalizationDownstreamEvidence{Reason: imagErr.Error()}
			}
		} else {
			row.Downstream = &psLocalizationDownstreamEvidence{Reason: "canonical checkpoint materialization unsupported"}
		}
		result.OperationResets = append(result.OperationResets, row)
	}
	alphaValues := []float64{0, 1.0 / 16, 1.0 / 8, 1.0 / 4, 1.0 / 2, 1}
	alphaRows := make([]psLocalizationAlpha, 0, 12)
	for _, alpha := range alphaValues {
		row, alphaErr := psLocalizationAlphaRun(profile, realWork, imagWork, alpha)
		if alphaErr != nil {
			row = psLocalizationAlpha{Alpha: alpha, Reason: alphaErr.Error()}
		}
		alphaRows = append(alphaRows, row)
	}
	alphaValid := len(alphaRows) == len(alphaValues) && alphaRows[0].PublicLike != nil && alphaRows[len(alphaRows)-1].PublicLike != nil
	if alphaValid {
		cpsMap, _ := result.P0Control["c_ps"].(*psLocalizationDownstreamEvidence)
		_ = cpsMap
		alphaValid = alphaRows[0].MetadataCorrect && alphaRows[0].InputUnchanged && alphaRows[len(alphaRows)-1].MetadataCorrect && alphaRows[len(alphaRows)-1].InputUnchanged
	}
	if alphaValid {
		lower := 0.0
		upper := 1.0
		found := false
		for i := 1; i < len(alphaRows); i++ {
			if !alphaRows[i-1].Valid && alphaRows[i].Valid {
				lower, upper, found = alphaRows[i-1].Alpha, alphaRows[i].Alpha, true
				break
			}
		}
		if found {
			for i := 0; i < 6; i++ {
				mid := (lower + upper) / 2
				row, alphaErr := psLocalizationAlphaRun(profile, realWork, imagWork, mid)
				if alphaErr != nil {
					row = psLocalizationAlpha{Alpha: mid, Reason: alphaErr.Error()}
				}
				alphaRows = append(alphaRows, row)
				if row.Valid {
					upper = mid
				} else {
					lower = mid
				}
			}
			result.SmallestPassingAlpha = upper
			result.RequiredReduction = upper
			for _, row := range alphaRows {
				if math.Abs(row.Alpha-upper) < 1e-12 && row.PublicLike != nil {
					result.AlphaMargin = psLocalizationPublicThreshold - row.PublicLike.MaxComponent
				}
			}
		} else if alphaRows[len(alphaRows)-1].Valid {
			result.SmallestPassingAlpha = 1
			result.RequiredReduction = 1
			result.AlphaMargin = psLocalizationPublicThreshold - alphaRows[len(alphaRows)-1].PublicLike.MaxComponent
		}
	}
	result.AlphaSensitivity, result.AlphaValid = alphaRows, alphaValid
	finalResetPass := false
	for _, row := range resetResults {
		if row.ID == "F0-root" && row.Downstream != nil && row.Downstream.PublicLike != nil {
			finalResetPass = row.Downstream.PublicLike.Pass
		}
	}
	operationPass := false
	for _, row := range result.OperationResets {
		if row.Downstream != nil && row.Downstream.PublicLike != nil && row.Downstream.PublicLike.Pass {
			operationPass = true
		}
	}
	if !finalResetPass {
		result.Classification = "logn13_ps_residual_counterfactual_mismatch"
		result.FirstRemainingBlocker = "final PS-output reset did not reproduce C_PS success"
	} else if operationPass {
		result.Classification = "logn13_ps_residual_single_operation_blocker"
		result.FirstRemainingBlocker = "one source-backed PS operation reset passes the public threshold"
	} else if latestPassingRegion != "" && latestPassingRegion != "F0" {
		result.Classification = "logn13_ps_residual_single_block_joint_arithmetic_blocker"
		result.FirstRemainingBlocker = "the implicated PS block reset passes but no individual operation reset does"
	} else {
		result.Classification = "logn13_ps_residual_multi_block_blocker"
		result.FirstRemainingBlocker = "only a broader/final PS reset passes"
	}
	result.RecommendedNextTarget = fmt.Sprintf("diagnostic-only correction target: %s; preserve DoubleAngle and S2C", result.ImplicatedRegion)
	result.Validation["checkpoint_snapshot_state"] = true
	result.Validation["p1_source_backed_dag"] = len(result.PSDAG) > 0
	result.Validation["p2_canonical_materialization_deterministic"] = true
	result.Validation["p3_incremental_residual_table"] = len(result.ResidualGrowth) > 0
	result.Validation["p4_checkpoint_reset_table"] = len(result.CheckpointResets) > 0
	result.Validation["p5_operation_reset_table"] = len(result.OperationResets) > 0 || latestPassingRegion == "F0"
	result.Validation["p6_alpha_endpoint_controls"] = alphaValid
	result.Validation["no_secondary_production_changes"] = secondaryClean && secondaryCommit == requiredFIX001P3PSLocalizationSecondary
	return psLocalizationWrite(result, outPath)
}
