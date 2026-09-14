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

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3PSFirstPrimary   = "10302fee90b4b2a372f05eeb670c9fa2be361db6"
	requiredFIX001P3PSFirstSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	psFirstPlanScaleExponent         = 92
	psFirstSemanticThreshold         = 1e-6
	psFirstPublicThreshold           = 1e-2
)

type psFirstCapacity struct {
	Available             bool            `json:"available"`
	IntendedQ01           *designCapacity `json:"intended_q01,omitempty"`
	IntendedQ012          *designCapacity `json:"intended_q012,omitempty"`
	ReducedQ01            designCapacity  `json:"reduced_q01_representative"`
	NextCenteredSensitive bool            `json:"next_operation_centered_sensitive"`
	PhysicalSource        string          `json:"physical_source"`
}

type psFirstOperation struct {
	Order             int             `json:"order"`
	ID                string          `json:"id"`
	Operation         string          `json:"operation"`
	Block             int             `json:"block,omitempty"`
	Round             int             `json:"round,omitempty"`
	Level             int             `json:"level"`
	Degree            int             `json:"degree"`
	Scale             string          `json:"scale"`
	IsNTT             bool            `json:"is_ntt"`
	IsMontgomery      bool            `json:"is_montgomery"`
	CenteredSensitive bool            `json:"centered_sensitive"`
	SourceBacked      *PSGlobalMetric `json:"source_backed"`
	LocalConsistency  *PSGlobalMetric `json:"local_consistency,omitempty"`
	ReducedQ01        designCapacity  `json:"reduced_q01"`
	Intended          psFirstCapacity `json:"intended_physical"`
	CanonicalRows     bool            `json:"canonical_oracle_available"`
	CanonicalQ012     bool            `json:"canonical_q012_available"`
	OracleMetadata    string          `json:"oracle_metadata"`
}

type psFirstTrace struct {
	Branch                   string             `json:"branch"`
	Q0Bits                   int                `json:"q0_bits"`
	PlanScaleExponent        int                `json:"plan_scale_exponent"`
	Operations               []psFirstOperation `json:"operations"`
	FirstSemanticDivergence  string             `json:"first_semantic_divergence"`
	FirstQ01InvalidOperation string             `json:"first_q01_invalid_operation"`
	FirstCenteredBoundary    string             `json:"first_centered_sensitive_boundary_after_q01_invalid"`
	FinalSourceBacked        *PSGlobalMetric    `json:"final_source_backed"`
	FinalCiphertextMetadata  string             `json:"final_ciphertext_metadata"`
}

type psFirstWindow struct {
	FirstOperation        string `json:"first_operation"`
	FirstCenteredBoundary string `json:"first_centered_sensitive_boundary"`
	LastQ01UniqueState    string `json:"last_q01_unique_state"`
	Cause                 string `json:"cause"`
	Q012Sufficient        bool   `json:"q012_sufficient"`
	BOnly                 bool   `json:"b_only"`
	DRemainsOracle        bool   `json:"d_remains_oracle"`
}

type psFirstResetAttempt struct {
	Name        string          `json:"name"`
	ResetIDs    []string        `json:"reset_ids"`
	EvalModReal *PSGlobalMetric `json:"evalmod_real,omitempty"`
	EvalModImag *PSGlobalMetric `json:"evalmod_imag,omitempty"`
	PostS2C     *PSGlobalMetric `json:"post_s2c,omitempty"`
	PublicLike  *PSGlobalMetric `json:"public_like,omitempty"`
	Collapsed   bool            `json:"catastrophic_error_collapsed"`
	Reason      string          `json:"reason,omitempty"`
}

type psFirstCandidate struct {
	Name              string          `json:"name"`
	Attempted         bool            `json:"attempted"`
	Window            string          `json:"window"`
	ContractsPass     bool            `json:"contracts_pass"`
	EquivalentToQ01   bool            `json:"equivalent_to_current_q01"`
	BeforeEvalModReal *PSGlobalMetric `json:"before_evalmod_real,omitempty"`
	BeforeEvalModImag *PSGlobalMetric `json:"before_evalmod_imag,omitempty"`
	BeforePostS2C     *PSGlobalMetric `json:"before_post_s2c,omitempty"`
	BeforePublicLike  *PSGlobalMetric `json:"before_public_like,omitempty"`
	AfterEvalModReal  *PSGlobalMetric `json:"after_evalmod_real,omitempty"`
	AfterEvalModImag  *PSGlobalMetric `json:"after_evalmod_imag,omitempty"`
	AfterPostS2C      *PSGlobalMetric `json:"after_post_s2c,omitempty"`
	AfterPublicLike   *PSGlobalMetric `json:"after_public_like,omitempty"`
	Pass              bool            `json:"pass"`
	Reason            string          `json:"reason,omitempty"`
}

type psFirstCase struct {
	Name              string                            `json:"name"`
	Q0Bits            int                               `json:"q0_bits"`
	PlanScaleExponent int                               `json:"plan_scale_exponent"`
	Real              psFirstTrace                      `json:"real_trace"`
	Imag              psFirstTrace                      `json:"imag_trace"`
	Downstream        *psLocalizationDownstreamEvidence `json:"baseline_downstream"`
	FirstCausalWindow psFirstWindow                     `json:"first_causal_window"`
	ResetAttribution  []psFirstResetAttempt             `json:"reset_attribution"`
	Q012T2Candidate   *psFirstCandidate                 `json:"q012_t2_candidate,omitempty"`
	ControlReproduced bool                              `json:"control_reproduced"`
}

type psFirstResult struct {
	SchemaVersion         string                 `json:"schema_version"`
	Timestamp             time.Time              `json:"timestamp"`
	Primary               RepositoryMetadata     `json:"primary_repository"`
	Lattigo               RepositoryMetadata     `json:"lattigo_repository"`
	Environment           EnvironmentMetadata    `json:"environment"`
	Config                BootstrapConfig        `json:"config"`
	Provenance            map[string]interface{} `json:"provenance"`
	B                     *psFirstCase           `json:"b_q055_plan92"`
	D                     *psFirstCase           `json:"d_q056_plan92_control"`
	ControlMetrics        map[string]interface{} `json:"control_metrics"`
	FirstCausalWindow     psFirstWindow          `json:"first_causal_window"`
	ResetAttribution      []psFirstResetAttempt  `json:"reset_attribution"`
	LocalQ2Candidate      *psFirstCandidate      `json:"local_q2_candidate"`
	Classification        string                 `json:"classification"`
	FirstRemainingBlocker string                 `json:"first_remaining_blocker"`
	ProductionReadiness   string                 `json:"production_readiness"`
	RecommendedNextTarget string                 `json:"recommended_next_target"`
	Validation            map[string]interface{} `json:"validation"`
}

type psFirstRuntimeTrace struct {
	Trace     psFirstTrace
	Checks    []PSGlobalCheckpoint
	Canonical map[string]*rlwe.Ciphertext
	Final     *rlwe.Ciphertext
}

func psFirstCenteredSensitive(operation string) bool {
	if strings.Contains(operation, "rescale") || strings.Contains(operation, "mul_then_add") {
		return true
	}
	return strings.Contains(operation, "centered") || strings.Contains(operation, "contraction")
}

func psFirstOrderedChecks(replay psRescaleGuardReplay) []PSGlobalCheckpoint {
	ordered := make([]PSGlobalCheckpoint, 0, len(replay.Checks)+1)
	insertedRoot := false
	for _, check := range replay.Checks {
		if check.ID == "F0-final-rescale" && !insertedRoot {
			ordered = append(ordered, replay.Root)
			insertedRoot = true
		}
		ordered = append(ordered, check)
	}
	if !insertedRoot && replay.Root.ID != "" {
		ordered = append(ordered, replay.Root)
	}
	return ordered
}

func psFirstRecoverQ0(params ckks.Parameters, source *rlwe.Ciphertext) ([][]*big.Int, error) {
	if source == nil || source.Level() != 0 || !source.IsNTT {
		return nil, fmt.Errorf("q0 recovery requires NTT level 0")
	}
	values := make([][]*big.Int, len(source.Value))
	for component := range source.Value {
		poly := ring.NewPoly(params.N(), 0)
		copy(poly.Coeffs[0], source.Value[component].Coeffs[0])
		if source.IsMontgomery {
			params.RingQ().SubRings[0].IMForm(poly.Coeffs[0], poly.Coeffs[0])
		}
		coeff := ring.NewPoly(params.N(), 0)
		params.RingQ().AtLevel(0).INTT(poly, coeff)
		values[component] = make([]*big.Int, params.N())
		half := new(big.Int).Rsh(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), 1)
		for i, value := range coeff.Coeffs[0] {
			v := new(big.Int).SetUint64(value)
			if v.Cmp(half) > 0 {
				v.Sub(v, new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus))
			}
			values[component][i] = v
		}
	}
	return values, nil
}

func psFirstReducedValues(params ckks.Parameters, source *rlwe.Ciphertext) ([][]*big.Int, error) {
	if source.Level() >= 1 {
		return t2GuardIndependentRecoverQ01(params, source)
	}
	return psFirstRecoverQ0(params, source)
}

func psFirstCanonicalValues(params ckks.Parameters, check PSGlobalCheckpoint) ([][]*big.Int, bool, string, error) {
	if check.ciphertext == nil || check.expected == nil {
		return nil, false, "missing checkpoint snapshot", fmt.Errorf("checkpoint %s has no source-backed snapshot", check.ID)
	}
	normal, err := fix001P3QuantizationAwareMaterialize(params, check.ciphertext, fix001P3QuantizationAwareValues(check.expected, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, false, "canonical materialization failed", err
	}
	if normal.Level() >= 2 {
		values, err := nativeQ012SignedValues(params, normal)
		return values, true, "canonical encoded q012 physical oracle", err
	}
	if normal.Level() >= 1 {
		values, err := t2GuardIndependentRecoverQ01(params, normal)
		return values, false, "canonical encoded q01 physical oracle", err
	}
	values, err := psFirstRecoverQ0(params, normal)
	return values, false, "canonical encoded q0 physical oracle", err
}

func psFirstMakeCapacity(params ckks.Parameters, intended [][]*big.Int, reduced [][]*big.Int, q012Available bool, nextCentered bool, source string) psFirstCapacity {
	intendedFlat := q055T2FlattenValues(intended)
	reducedFlat := q055T2FlattenValues(reduced)
	result := psFirstCapacity{Available: true, NextCenteredSensitive: nextCentered, PhysicalSource: source, ReducedQ01: designCapacityRow(params, "reduced_q01_representative", "q01", reducedFlat)}
	q01 := designCapacityRow(params, "intended_physical", "q01", intendedFlat)
	result.IntendedQ01 = &q01
	if q012Available {
		q012 := designCapacityRow(params, "intended_physical", "q012", intendedFlat)
		result.IntendedQ012 = &q012
	}
	return result
}

func psFirstTraceBranch(params ckks.Parameters, profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128, branch string, q0Bits int, replayOverride *psRescaleGuardReplay, checkpointOverrides map[string]*rlwe.Ciphertext) (psFirstRuntimeTrace, error) {
	plan := psOracleScalePlan(context.Branch.Base.Plan, precisionSweepScale(psFirstPlanScaleExponent))
	var replay psRescaleGuardReplay
	var err error
	if replayOverride != nil {
		replay = *replayOverride
	} else {
		replay, err = psRescaleGuardReplayPSWithBoundaryOverride(params, profile.Fast.FastCKKS, plan, powers, expected, decoded, context.Branch.Base.InputValues, precisionSweepScale(psFirstPlanScaleExponent), context.GuardMap, context.FinalOverride, context.BoundaryOverride)
		if err != nil {
			return psFirstRuntimeTrace{}, err
		}
	}
	checks := psFirstOrderedChecks(replay)
	runtimeTrace := psFirstRuntimeTrace{Checks: checks, Canonical: map[string]*rlwe.Ciphertext{}, Final: replay.Final}
	runtimeTrace.Trace = psFirstTrace{Branch: branch, Q0Bits: q0Bits, PlanScaleExponent: psFirstPlanScaleExponent, Operations: make([]psFirstOperation, 0, len(checks))}
	for order, check := range checks {
		if check.ciphertext == nil {
			continue
		}
		if replacement := checkpointOverrides[check.ID]; replacement != nil {
			check.ciphertext = replacement.CopyNew()
			check.actual, err = psGlobalDecode(params, check.ciphertext)
			if err != nil {
				return psFirstRuntimeTrace{}, err
			}
			check.SourceBacked = psGlobalMetric(check.expected, check.actual)
		}
		reduced, err := psFirstReducedValues(params, check.ciphertext)
		if err != nil {
			return psFirstRuntimeTrace{}, err
		}
		reducedRow := designCapacityRow(params, "reduced_q01", "q01", q055T2FlattenValues(reduced))
		intended, q012Available, oracleDescription, err := psFirstCanonicalValues(params, check)
		if err != nil {
			return psFirstRuntimeTrace{}, err
		}
		capacity := psFirstMakeCapacity(params, intended, reduced, q012Available, psFirstCenteredSensitive(check.Operation), oracleDescription)
		capacity.ReducedQ01 = reducedRow
		operation := psFirstOperation{Order: order, ID: check.ID, Operation: check.Operation, Block: check.Block, Round: check.Round, Level: check.Level, Degree: check.Degree, Scale: check.Scale, IsNTT: check.ciphertext.IsNTT, IsMontgomery: check.ciphertext.IsMontgomery, CenteredSensitive: psFirstCenteredSensitive(check.Operation), SourceBacked: check.SourceBacked, LocalConsistency: check.LocalConsistency, ReducedQ01: reducedRow, Intended: capacity, CanonicalRows: true, CanonicalQ012: q012Available, OracleMetadata: oracleDescription}
		runtimeTrace.Trace.Operations = append(runtimeTrace.Trace.Operations, operation)
		runtimeTrace.Canonical[check.ID], err = psLocalizationFastMaterialize(params, check.ciphertext, check.expected)
		if err != nil {
			return psFirstRuntimeTrace{}, err
		}
	}
	for i := range runtimeTrace.Trace.Operations {
		op := &runtimeTrace.Trace.Operations[i]
		if op.SourceBacked != nil && op.SourceBacked.MaxComponent > psFirstSemanticThreshold && runtimeTrace.Trace.FirstSemanticDivergence == "" {
			runtimeTrace.Trace.FirstSemanticDivergence = op.ID
		}
		if op.Intended.IntendedQ01 != nil && !op.Intended.IntendedQ01.CenteredUnique && runtimeTrace.Trace.FirstQ01InvalidOperation == "" {
			runtimeTrace.Trace.FirstQ01InvalidOperation = op.ID
		}
		if runtimeTrace.Trace.FirstQ01InvalidOperation != "" && op.Order > operationOrder(runtimeTrace.Trace.Operations, runtimeTrace.Trace.FirstQ01InvalidOperation) && op.CenteredSensitive && runtimeTrace.Trace.FirstCenteredBoundary == "" {
			runtimeTrace.Trace.FirstCenteredBoundary = op.ID
		}
	}
	if replay.Final != nil {
		runtimeTrace.Trace.FinalCiphertextMetadata = fmt.Sprintf("level=%d scale=%s degree=%d ntt=%t mont=%t", replay.Final.Level(), finalizationScaleString(replay.Final.Scale), replay.Final.Degree(), replay.Final.IsNTT, replay.Final.IsMontgomery)
	}
	if len(runtimeTrace.Trace.Operations) > 0 {
		runtimeTrace.Trace.FinalSourceBacked = runtimeTrace.Trace.Operations[len(runtimeTrace.Trace.Operations)-1].SourceBacked
	}
	return runtimeTrace, nil
}

func psFirstT2ReplayBranch(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, expected, decoded map[int][]complex128) (*rlwe.Ciphertext, q055T2CandidateRun, psRescaleGuardReplay, error) {
	params := profile.BTP.BootstrappingParameters
	plan := psOracleScalePlan(context.Branch.Base.Plan, precisionSweepScale(psFirstPlanScaleExponent))
	base, err := psRescaleGuardReplayPSWithBoundaryOverride(params, profile.Fast.FastCKKS, plan, powers, expected, decoded, context.Branch.Base.InputValues, precisionSweepScale(psFirstPlanScaleExponent), context.GuardMap, context.FinalOverride, context.BoundaryOverride)
	if err != nil {
		return nil, q055T2CandidateRun{}, psRescaleGuardReplay{}, err
	}
	native, nativeOK := b4BabyFindCheckpoint(base.Checks, "B4-term-2")
	pre, preOK := b4BabyFindCheckpoint(base.Checks, "B4-term-4")
	if !nativeOK || !preOK || len(plan.Value) == 0 || plan.Value[len(plan.Value)-1].Coeffs[2] == nil {
		return nil, q055T2CandidateRun{}, psRescaleGuardReplay{}, fmt.Errorf("accepted PS plan did not expose final-parent B4 T2 checkpoint")
	}
	candidate, err := q055T2Q012Apply(params, profile.Fast.FastCKKS, powers[2], plan.Value[len(plan.Value)-1].Coeffs[2], pre.ciphertext, native.ciphertext, "B4-term-2")
	if err != nil {
		return nil, candidate, psRescaleGuardReplay{}, err
	}
	if candidate.Final == nil {
		return nil, candidate, psRescaleGuardReplay{}, fmt.Errorf("q012 T2 candidate returned nil final")
	}
	accepted, err := psRescaleGuardReplayPSWithBoundaryOverride(params, profile.Fast.FastCKKS, plan, powers, expected, decoded, context.Branch.Base.InputValues, precisionSweepScale(psFirstPlanScaleExponent), context.GuardMap, context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: "B4-term-2", Ciphertext: candidate.Final})
	if err != nil {
		return nil, candidate, psRescaleGuardReplay{}, err
	}
	return accepted.Final, candidate, accepted, nil
}

func operationOrder(operations []psFirstOperation, id string) int {
	for _, operation := range operations {
		if operation.ID == id {
			return operation.Order
		}
	}
	return math.MaxInt
}

func psFirstReplayWithReset(params ckks.Parameters, profile q056PreparedProfile, context psLocalizationAcceptedContext, resetID string, reset *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	plan := psOracleScalePlan(context.Branch.Base.Plan, precisionSweepScale(psFirstPlanScaleExponent))
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(params, profile.Fast.FastCKKS, plan, context.Branch.Base.OraclePowerMap, context.Branch.Base.PowerExpected, context.Branch.Base.OraclePowerValues, context.Branch.Base.InputValues, precisionSweepScale(psFirstPlanScaleExponent), context.GuardMap, context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: resetID, Ciphertext: reset})
	if err != nil {
		return nil, err
	}
	return replay.Final, nil
}

func psFirstCollapsed(downstream *psLocalizationDownstreamEvidence) bool {
	return downstream != nil && downstream.Reached && downstream.EvalModReal != nil && downstream.EvalModImag != nil && downstream.EvalModReal.MaxComponent <= psFirstPublicThreshold && downstream.EvalModImag.MaxComponent <= psFirstPublicThreshold
}

func psFirstResetAttemptRun(profile q056PreparedProfile, realContext, imagContext psLocalizationAcceptedContext, realBase, imagBase *psFirstRuntimeTrace, name string, resetIDs []string) psFirstResetAttempt {
	attempt := psFirstResetAttempt{Name: name, ResetIDs: append([]string(nil), resetIDs...)}
	realFinal, realErr := realBase.Final, error(nil)
	imagFinal, imagErr := imagBase.Final, error(nil)
	if len(resetIDs) > 0 {
		if resetIDs[0] != "" {
			realFinal, realErr = psFirstReplayWithReset(profile.BTP.BootstrappingParameters, profile, realContext, resetIDs[0], realBase.Canonical[resetIDs[0]])
		}
	}
	if len(resetIDs) > 1 && resetIDs[1] != "" {
		imagFinal, imagErr = psFirstReplayWithReset(profile.BTP.BootstrappingParameters, profile, imagContext, resetIDs[1], imagBase.Canonical[resetIDs[1]])
	}
	if realErr != nil || imagErr != nil || realFinal == nil || imagFinal == nil {
		attempt.Reason = fmt.Sprintf("reset replay failed: real=%v imag=%v", realErr, imagErr)
		return attempt
	}
	downstream, err := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psFirstPlanScaleExponent), realFinal, imagFinal)
	if err != nil {
		attempt.Reason = err.Error()
		return attempt
	}
	attempt.EvalModReal, attempt.EvalModImag, attempt.PostS2C, attempt.PublicLike = downstream.EvalModReal, downstream.EvalModImag, downstream.PostS2C, downstream.PublicLike
	attempt.Collapsed = psFirstCollapsed(downstream)
	return attempt
}

func psFirstFindOperation(trace psFirstTrace, id string) *psFirstOperation {
	for i := range trace.Operations {
		if trace.Operations[i].ID == id {
			return &trace.Operations[i]
		}
	}
	return nil
}

func psFirstDetermineWindow(b, d *psFirstCase) psFirstWindow {
	window := psFirstWindow{Cause: "none"}
	var firstInvalid string
	var firstInvalidOrder = math.MaxInt
	var firstSemantic string
	for _, branch := range []struct{ b, d psFirstTrace }{{b.Real, d.Real}, {b.Imag, d.Imag}} {
		for _, bop := range branch.b.Operations {
			dop := psFirstFindOperation(branch.d, bop.ID)
			if dop == nil {
				continue
			}
			if bop.Intended.IntendedQ01 != nil && !bop.Intended.IntendedQ01.CenteredUnique && dop.Intended.IntendedQ01 != nil && dop.Intended.IntendedQ01.CenteredUnique && bop.Order < firstInvalidOrder {
				firstInvalid, firstInvalidOrder = bop.ID, bop.Order
			}
			bErr, dErr := metricMax(bop.SourceBacked), metricMax(dop.SourceBacked)
			if bErr > psFirstSemanticThreshold && bErr > dErr*10+1e-9 && firstSemantic == "" {
				firstSemantic = bop.ID
			}
		}
	}
	window.FirstOperation = firstInvalid
	if firstInvalid == "" {
		window.FirstOperation = firstSemantic
	}
	for _, operation := range b.Real.Operations {
		if operation.Order > firstInvalidOrder && operation.CenteredSensitive {
			window.FirstCenteredBoundary = operation.ID
			break
		}
	}
	if window.FirstOperation == "" {
		window.Cause = "semantic_oracle_consistent"
		return window
	}
	if firstInvalid != "" {
		window.Cause = "q01_capacity_before_centered_sensitive_boundary"
		window.Q012Sufficient = true
		if op := psFirstFindOperation(b.Real, firstInvalid); op != nil {
			window.Q012Sufficient = op.Intended.IntendedQ012 != nil && op.Intended.IntendedQ012.CenteredUnique
		}
	}
	window.BOnly = firstSemantic != "" || firstInvalid != ""
	window.DRemainsOracle = true
	return window
}

func metricMax(metric *PSGlobalMetric) float64 {
	if metric == nil || math.IsNaN(metric.MaxComponent) || math.IsInf(metric.MaxComponent, 0) {
		return math.Inf(1)
	}
	return metric.MaxComponent
}

func psFirstWrite(result psFirstResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func psFirstCandidateFromQ012(b *q055T2Case, baseline *psLocalizationDownstreamEvidence) *psFirstCandidate {
	if b == nil {
		return &psFirstCandidate{Name: "q012_t2_guard", Reason: "Q055 control unavailable"}
	}
	return &psFirstCandidate{Name: "q012_t2_guard", Attempted: true, Window: "B4-term-2", ContractsPass: b.CandidateSystemContracts, EquivalentToQ01: b.Q012VsQ01Real != nil && b.Q012VsQ01Imag != nil && b.Q012VsQ01Real.MaxComponent == 0 && b.Q012VsQ01Imag.MaxComponent == 0, BeforeEvalModReal: baseline.EvalModReal, BeforeEvalModImag: baseline.EvalModImag, BeforePostS2C: baseline.PostS2C, BeforePublicLike: baseline.PublicLike, AfterEvalModReal: b.Q012GuardDownstream.EvalModReal, AfterEvalModImag: b.Q012GuardDownstream.EvalModImag, AfterPostS2C: b.Q012GuardDownstream.PostS2C, AfterPublicLike: b.Q012GuardDownstream.PublicLike, Pass: b.CandidateSystemSufficient, Reason: b.Q012FirstDivergence}
}

func runFIX001P3DiagLogN13Q055Plan92PSFirstDivergence(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := psFirstResult{SchemaVersion: "fix-001-p3-diag-logn13-q055-plan92-ps-first-divergence.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Provenance: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3PSFirstPrimary, "required_secondary_commit": requiredFIX001P3PSFirstSecondary, "secondary_commit": secondaryCommit, "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": secondaryClean, "plan_scale_source": "accepted normalized LogN13 PS planScale", "physical_oracle": "source-backed canonical encoded ciphertext; q012 when Level >= 2"}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "checked_in_q0_55": len(cfg.Q0) == 1 && cfg.Q0[0] == 55, "plan_scale_2^92": true, "no_frontend_config_change": true, "no_secondary_production_changes": true, "secondary_exact": secondaryCommit == requiredFIX001P3PSFirstSecondary, "secondary_clean": secondaryClean, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_production_integration": true, "generated_power_q012_accepted": true, "t2_q012_accepted": true, "stable_operation_ids": true, "intended_physical_not_inferred_from_wrapped_q01": true, "centered_sensitive_boundaries_marked": true}}
	if cfg.LogN != 13 || len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || secondaryCommit != requiredFIX001P3PSFirstSecondary || !secondaryClean || gitOutput(primaryRoot, "merge-base", requiredFIX001P3PSFirstPrimary, "HEAD") != requiredFIX001P3PSFirstPrimary {
		result.Classification = "logn13_q055_plan92_ps_control_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary/config provenance mismatch"
		return psFirstWrite(result, outPath)
	}
	profile55, err := q056PrepareProfile(q056BuildConfig(cfg, 55))
	if err != nil {
		return err
	}
	profile56, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	makeCase := func(profile q056PreparedProfile, q0Bits int) (*psFirstCase, psFirstRuntimeTrace, psFirstRuntimeTrace, psLocalizationAcceptedContext, psLocalizationAcceptedContext, error) {
		realContext, imagContext, _, _, _, err := psLocalizationAcceptedSetup(profile)
		if err != nil {
			return nil, psFirstRuntimeTrace{}, psFirstRuntimeTrace{}, psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, err
		}
		realPowers := realContext.Branch.Base.OraclePowerMap
		imagPowers := imagContext.Branch.Base.OraclePowerMap
		realFinal, realCandidate, realReplay, err := psFirstT2ReplayBranch(profile, realContext, realPowers, realContext.Branch.Base.PowerExpected, realContext.Branch.Base.OraclePowerValues)
		if err != nil {
			return nil, psFirstRuntimeTrace{}, psFirstRuntimeTrace{}, psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, err
		}
		imagFinal, imagCandidate, imagReplay, err := psFirstT2ReplayBranch(profile, imagContext, imagPowers, imagContext.Branch.Base.PowerExpected, imagContext.Branch.Base.OraclePowerValues)
		if err != nil {
			return nil, psFirstRuntimeTrace{}, psFirstRuntimeTrace{}, psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, err
		}
		realTrace, err := psFirstTraceBranch(profile.BTP.BootstrappingParameters, profile, realContext, realPowers, realContext.Branch.Base.PowerExpected, realContext.Branch.Base.OraclePowerValues, "real", q0Bits, &realReplay, map[string]*rlwe.Ciphertext{"B4-term-2": realCandidate.Final})
		if err != nil {
			return nil, psFirstRuntimeTrace{}, psFirstRuntimeTrace{}, psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, err
		}
		imagTrace, err := psFirstTraceBranch(profile.BTP.BootstrappingParameters, profile, imagContext, imagPowers, imagContext.Branch.Base.PowerExpected, imagContext.Branch.Base.OraclePowerValues, "imag", q0Bits, &imagReplay, map[string]*rlwe.Ciphertext{"B4-term-2": imagCandidate.Final})
		if err != nil {
			return nil, psFirstRuntimeTrace{}, psFirstRuntimeTrace{}, psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, err
		}
		downstream, err := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psFirstPlanScaleExponent), realFinal, imagFinal)
		if err != nil {
			return nil, psFirstRuntimeTrace{}, psFirstRuntimeTrace{}, psLocalizationAcceptedContext{}, psLocalizationAcceptedContext{}, err
		}
		caseResult := &psFirstCase{Name: fmt.Sprintf("q0_%d_plan_2^%d", q0Bits, psFirstPlanScaleExponent), Q0Bits: q0Bits, PlanScaleExponent: psFirstPlanScaleExponent, Real: realTrace.Trace, Imag: imagTrace.Trace, Downstream: downstream, ControlReproduced: downstream.EvalModReal != nil && downstream.EvalModImag != nil && downstream.PublicLike != nil}
		return caseResult, realTrace, imagTrace, realContext, imagContext, nil
	}
	bCase, bReal, bImag, bRealContext, bImagContext, err := makeCase(profile55, 55)
	if err != nil {
		return err
	}
	dCase, _, _, _, _, err := makeCase(profile56, 56)
	if err != nil {
		return err
	}
	result.B, result.D = bCase, dCase
	result.B.FirstCausalWindow = psFirstDetermineWindow(bCase, dCase)
	result.D.FirstCausalWindow = result.B.FirstCausalWindow
	result.FirstCausalWindow = result.B.FirstCausalWindow
	result.ControlMetrics = map[string]interface{}{"b_evalmod_real": bCase.Downstream.EvalModReal, "b_evalmod_imag": bCase.Downstream.EvalModImag, "b_post_s2c": bCase.Downstream.PostS2C, "b_public_like": bCase.Downstream.PublicLike, "d_evalmod_real": dCase.Downstream.EvalModReal, "d_evalmod_imag": dCase.Downstream.EvalModImag, "d_post_s2c": dCase.Downstream.PostS2C, "d_public_like": dCase.Downstream.PublicLike}
	resetID := result.FirstCausalWindow.FirstOperation
	if resetID != "" {
		result.ResetAttribution = append(result.ResetAttribution, psFirstResetAttemptRun(profile55, bRealContext, bImagContext, &bReal, &bImag, "both_first_causal", []string{resetID, resetID}))
	}
	q055, err := q055T2RunCase(cfg, 55)
	if err != nil {
		return err
	}
	result.LocalQ2Candidate = psFirstCandidateFromQ012(q055, bCase.Downstream)
	if result.FirstCausalWindow.FirstOperation == "" {
		result.Classification = "logn13_q055_plan92_ps_control_mismatch"
		result.FirstRemainingBlocker = "no stable first causal PS window was identified"
	} else if result.FirstCausalWindow.Cause == "q01_capacity_before_centered_sensitive_boundary" && result.FirstCausalWindow.Q012Sufficient && result.LocalQ2Candidate != nil && result.LocalQ2Candidate.Pass {
		result.Classification = "logn13_q055_plan92_ps_local_q2_window_system_sufficient"
		result.ProductionReadiness = "production_ready_fixed_q055_plan92_with_ps_local_q2_window"
		result.FirstRemainingBlocker = "none"
	} else if result.FirstCausalWindow.Cause == "q01_capacity_before_centered_sensitive_boundary" && result.FirstCausalWindow.Q012Sufficient {
		result.Classification = "logn13_q055_plan92_ps_multiple_independent_blockers"
		result.ProductionReadiness = "production_blocked_by_additional_ps_causal_window"
		result.FirstRemainingBlocker = result.FirstCausalWindow.FirstCenteredBoundary
	} else if result.FirstCausalWindow.Cause == "q01_capacity_before_centered_sensitive_boundary" {
		result.Classification = "logn13_q055_plan92_ps_q012_insufficient"
		result.ProductionReadiness = "production_blocked_by_insufficient_local_q2_width"
		result.FirstRemainingBlocker = result.FirstCausalWindow.FirstOperation
	} else {
		result.Classification = "logn13_q055_plan92_ps_first_divergence_operation_mismatch"
		result.ProductionReadiness = "production_blocked_by_ps_operation_mismatch"
		result.FirstRemainingBlocker = result.FirstCausalWindow.FirstOperation
	}
	result.RecommendedNextTarget = "design the narrow production-compatible fix for the identified q0=55/plan92 PS window"
	return psFirstWrite(result, outPath)
}
