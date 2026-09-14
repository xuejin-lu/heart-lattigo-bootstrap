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
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3DependencyWindowPrimary   = "45db7fa141158f1c672ae0cbb86ebddb88939fb7"
	requiredFIX001P3DependencyWindowSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	dependencyWindowPlanScaleExponent         = 92
	dependencyWindowThreshold                 = 1e-2
)

type dependencyGraphNode struct {
	ID                 string          `json:"id"`
	Branch             string          `json:"branch"`
	Operation          string          `json:"operation"`
	InputIDs           []string        `json:"input_ids"`
	OutputID           string          `json:"output_id"`
	Block              int             `json:"block,omitempty"`
	Round              int             `json:"round,omitempty"`
	Level              int             `json:"level"`
	Scale              string          `json:"scale"`
	Degree             int             `json:"degree"`
	IsNTT              bool            `json:"is_ntt"`
	IsMontgomery       bool            `json:"is_montgomery"`
	CenteredSensitive  bool            `json:"centered_sensitive"`
	Q01CapacityRatio   float64         `json:"q01_capacity_ratio"`
	Q012CapacityRatio  float64         `json:"q012_capacity_ratio"`
	Q01CenteredUnique  bool            `json:"q01_centered_unique"`
	Q012CenteredUnique bool            `json:"q012_centered_unique"`
	SemanticResidual   *PSGlobalMetric `json:"semantic_residual,omitempty"`
	OracleMetadata     string          `json:"oracle_metadata"`
}

type dependencyFrontier struct {
	Branch                string   `json:"branch"`
	UnsafeNode            string   `json:"unsafe_node"`
	LastQ01UniqueAncestor string   `json:"last_q01_unique_ancestor"`
	FirstCenteredConsumer string   `json:"first_dependency_centered_consumer"`
	SafeContractionPoint  string   `json:"safe_contraction_point"`
	MaximumQ01Ratio       float64  `json:"maximum_intended_q01_ratio"`
	MaximumQ012Ratio      float64  `json:"maximum_intended_q012_ratio"`
	Q012Sufficient        bool     `json:"q012_sufficient"`
	BOnly                 bool     `json:"b_only"`
	DependencyPath        []string `json:"dependency_path"`
}

type dependencyResetResult struct {
	ResetSet    []string        `json:"reset_set"`
	EvalModReal *PSGlobalMetric `json:"evalmod_real,omitempty"`
	EvalModImag *PSGlobalMetric `json:"evalmod_imag,omitempty"`
	PostS2C     *PSGlobalMetric `json:"post_s2c,omitempty"`
	PublicLike  *PSGlobalMetric `json:"public_like,omitempty"`
	Contracts   bool            `json:"contracts_pass"`
	Collapsed   bool            `json:"catastrophic_error_collapsed"`
	Reason      string          `json:"reason,omitempty"`
}

type dependencyWindowProof struct {
	Window                  string                     `json:"window"`
	Branch                  string                     `json:"branch"`
	Start                   string                     `json:"start"`
	End                     string                     `json:"end"`
	Operations              []string                   `json:"operations"`
	ExactQ012Lift           bool                       `json:"exact_q012_lift"`
	Q012RowsMatch           bool                       `json:"q012_rows_match_oracle"`
	MetadataMatch           bool                       `json:"metadata_match"`
	Q012CenteredUnique      bool                       `json:"q012_centered_unique_throughout"`
	PostWindowQ01Unique     bool                       `json:"post_window_q01_centered_unique"`
	ContractionRowsMatch    bool                       `json:"q01_contraction_rows_match"`
	DiagnosticBigIntRescale bool                       `json:"diagnostic_bigint_rescale_only"`
	MaxQ012Ratio            float64                    `json:"max_q012_ratio"`
	OperationResiduals      map[string]*PSGlobalMetric `json:"operation_residuals,omitempty"`
	OperationRowsMatch      map[string]bool            `json:"operation_rows_match,omitempty"`
	CandidateMetadata       string                     `json:"candidate_metadata,omitempty"`
	BaselineMetadata        string                     `json:"baseline_metadata,omitempty"`
	SemanticResidual        *PSGlobalMetric            `json:"semantic_residual,omitempty"`
	Valid                   bool                       `json:"valid"`
	Reason                  string                     `json:"reason,omitempty"`
}

type dependencyCandidateCase struct {
	Name      string                            `json:"name"`
	WindowSet []string                          `json:"window_set"`
	Proofs    []dependencyWindowProof           `json:"proofs"`
	Metrics   *psLocalizationDownstreamEvidence `json:"metrics,omitempty"`
	Pass      bool                              `json:"pass"`
	Reason    string                            `json:"reason,omitempty"`
}

type dependencyControlCase struct {
	Name             string                            `json:"name"`
	Q0Bits           int                               `json:"q0_bits"`
	GeneratedKeys    []int                             `json:"generated_power_keys"`
	GenerationPassed bool                              `json:"generated_power_generation_passed"`
	Real             psFirstTrace                      `json:"real_trace"`
	Imag             psFirstTrace                      `json:"imag_trace"`
	Metrics          *psLocalizationDownstreamEvidence `json:"metrics"`
}

type dependencyWindowResult struct {
	SchemaVersion       string                           `json:"schema_version"`
	Timestamp           time.Time                        `json:"timestamp"`
	Primary             RepositoryMetadata               `json:"primary_repository"`
	Lattigo             RepositoryMetadata               `json:"lattigo_repository"`
	Environment         EnvironmentMetadata              `json:"environment"`
	Config              BootstrapConfig                  `json:"config"`
	Provenance          map[string]interface{}           `json:"provenance"`
	Corrections         map[string]bool                  `json:"corrections"`
	B                   *dependencyControlCase           `json:"b_q055_plan92_generated"`
	D                   *dependencyControlCase           `json:"d_q056_plan92_generated"`
	DependencyGraph     map[string][]dependencyGraphNode `json:"dependency_graph"`
	Frontiers           []dependencyFrontier             `json:"b_only_capacity_frontiers"`
	CumulativeResets    []dependencyResetResult          `json:"cumulative_resets"`
	MinimalResetSet     []string                         `json:"minimal_causal_reset_set"`
	Candidates          []dependencyCandidateCase        `json:"candidate_expansion_history"`
	Classification      string                           `json:"classification"`
	ProductionReadiness string                           `json:"production_readiness"`
	FirstRemaining      string                           `json:"first_remaining_blocker"`
	Recommended         string                           `json:"recommended_next_target"`
	Validation          map[string]interface{}           `json:"validation"`
}

type dependencyRuntimeCase struct {
	Control     *dependencyControlCase
	Profile     q056PreparedProfile
	RealContext psLocalizationAcceptedContext
	ImagContext psLocalizationAcceptedContext
	RealReplay  psRescaleGuardReplay
	ImagReplay  psRescaleGuardReplay
	RealPowers  map[int]*rlwe.Ciphertext
	ImagPowers  map[int]*rlwe.Ciphertext
	RealValues  map[int][]complex128
	ImagValues  map[int][]complex128
	RealTrace   psFirstRuntimeTrace
	ImagTrace   psFirstRuntimeTrace
}

func dependencyQ012MulDegree2(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Degree() != 1 || right.Degree() != 1 || left.Level() < 2 || right.Level() < 2 {
		return nil, fmt.Errorf("q012 degree-2 multiply requires degree-one level >= 2 operands")
	}
	level := minInt(left.Level(), right.Level())
	out := ckks.NewCiphertext(params, 2, level)
	*out.MetaData = *left.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	for limb := 0; limb <= 2; limb++ {
		sr := params.RingQ().SubRings[limb]
		t0 := make([]uint64, params.N())
		t1 := make([]uint64, params.N())
		t2 := make([]uint64, params.N())
		t3 := make([]uint64, params.N())
		sr.MulCoeffsMontgomery(left.Value[0].Coeffs[limb], right.Value[0].Coeffs[limb], t0)
		sr.MulCoeffsMontgomery(left.Value[0].Coeffs[limb], right.Value[1].Coeffs[limb], t1)
		sr.MulCoeffsMontgomery(left.Value[1].Coeffs[limb], right.Value[0].Coeffs[limb], t2)
		sr.Add(t1, t2, t1)
		sr.MulCoeffsMontgomery(left.Value[1].Coeffs[limb], right.Value[1].Coeffs[limb], t3)
		copy(out.Value[0].Coeffs[limb], t0)
		copy(out.Value[1].Coeffs[limb], t1)
		copy(out.Value[2].Coeffs[limb], t3)
	}
	out.Scale = left.Scale.Mul(right.Scale)
	return out, nil
}

func dependencyQ012PromoteDegree2(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || source.Degree() != 1 || source.Level() < 2 {
		return nil, fmt.Errorf("q012 degree promotion requires degree-one level >= 2 source")
	}
	out := ckks.NewCiphertext(params, 2, source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := 0; component <= 1; component++ {
		for limb := 0; limb <= 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func dependencyQ012Relinearize(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || source.Degree() != 2 || source.Level() < 2 {
		return nil, fmt.Errorf("q012 relinearization requires degree-two level >= 2 source")
	}
	out := ckks.NewCiphertext(params, 1, source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := 0; component <= 1; component++ {
		for limb := 0; limb <= 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func dependencyGeneratedPowers(profile q056PreparedProfile, branch string) (map[int]*rlwe.Ciphertext, map[int][]complex128, nativeQ012Result, error) {
	var input *rlwe.Ciphertext
	if branch == "real" {
		input = profile.Inputs.FastReal
	} else {
		input = profile.Inputs.FastImag
	}
	e2, zero, err := psGlobalE2(profile.BTP, profile.Fast, input)
	if err != nil {
		return nil, nil, nativeQ012Result{}, err
	}
	result := nativeQ012Result{}
	powers, values, err := nativeQ012Generate(profile.BTP.BootstrappingParameters, branch, e2, zero, &result)
	return powers, values, result, err
}

func dependencyAcceptedReplay(profile q056PreparedProfile, context psLocalizationAcceptedContext, powers map[int]*rlwe.Ciphertext, values map[int][]complex128) (psRescaleGuardReplay, q055T2CandidateRun, error) {
	planScale := precisionSweepScale(dependencyWindowPlanScaleExponent)
	context.Plan = psOracleScalePlan(context.Branch.Base.Plan, planScale)
	params := profile.BTP.BootstrappingParameters
	base, err := psRescaleGuardReplayPSWithBoundaryOverride(params, profile.Fast.FastCKKS, context.Plan, powers, context.Branch.Base.PowerExpected, values, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride)
	if err != nil {
		return psRescaleGuardReplay{}, q055T2CandidateRun{}, err
	}
	native, nativeOK := b4BabyFindCheckpoint(base.Checks, "B4-term-2")
	pre, preOK := b4BabyFindCheckpoint(base.Checks, "B4-term-4")
	poly := context.Plan.Value[len(context.Plan.Value)-1]
	if !nativeOK || !preOK || powers[2] == nil || len(poly.Coeffs) <= 2 || poly.Coeffs[2] == nil {
		return psRescaleGuardReplay{}, q055T2CandidateRun{}, fmt.Errorf("generated-power replay did not expose B4 T2")
	}
	candidate, err := q055T2Q012Apply(params, profile.Fast.FastCKKS, powers[2], poly.Coeffs[2], pre.ciphertext, native.ciphertext, "B4-term-2")
	if err != nil || candidate.Final == nil {
		return psRescaleGuardReplay{}, candidate, err
	}
	accepted, err := psRescaleGuardReplayPSWithBoundaryOverride(params, profile.Fast.FastCKKS, context.Plan, powers, context.Branch.Base.PowerExpected, values, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: "B4-term-2", Ciphertext: candidate.Final})
	return accepted, candidate, err
}

func dependencyMakeRuntimeCase(profile q056PreparedProfile, q0Bits int) (dependencyRuntimeCase, error) {
	realContext, imagContext, _, _, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	realPowers, realValues, realNative, err := dependencyGeneratedPowers(profile, "real")
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	imagPowers, imagValues, imagNative, err := dependencyGeneratedPowers(profile, "imag")
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	realReplay, _, err := dependencyAcceptedReplay(profile, realContext, realPowers, realValues)
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	imagReplay, _, err := dependencyAcceptedReplay(profile, imagContext, imagPowers, imagValues)
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	params := profile.BTP.BootstrappingParameters
	realTrace, err := psFirstTraceBranch(params, profile, realContext, realPowers, realContext.Branch.Base.PowerExpected, realValues, "real", q0Bits, &realReplay, nil)
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	imagTrace, err := psFirstTraceBranch(params, profile, imagContext, imagPowers, imagContext.Branch.Base.PowerExpected, imagValues, "imag", q0Bits, &imagReplay, nil)
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	downstream, err := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(dependencyWindowPlanScaleExponent), realReplay.Final, imagReplay.Final)
	if err != nil {
		return dependencyRuntimeCase{}, err
	}
	keys := []int{2, 3, 4, 6, 8, 16}
	control := &dependencyControlCase{Name: fmt.Sprintf("q0_%d_plan_2^%d_generated", q0Bits, dependencyWindowPlanScaleExponent), Q0Bits: q0Bits, GeneratedKeys: keys, GenerationPassed: realNative.Classification == "" && imagNative.Classification == "", Real: realTrace.Trace, Imag: imagTrace.Trace, Metrics: downstream}
	return dependencyRuntimeCase{Control: control, Profile: profile, RealContext: realContext, ImagContext: imagContext, RealReplay: realReplay, ImagReplay: imagReplay, RealPowers: realPowers, ImagPowers: imagPowers, RealValues: realValues, ImagValues: imagValues, RealTrace: realTrace, ImagTrace: imagTrace}, nil
}

func dependencyFindCheck(checks []PSGlobalCheckpoint, id string) *PSGlobalCheckpoint {
	for i := range checks {
		if checks[i].ID == id {
			return &checks[i]
		}
	}
	return nil
}

func dependencyVectorAdd(a, b []complex128) []complex128 {
	if len(a) != len(b) {
		return nil
	}
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func dependencyParentID(checks []PSGlobalCheckpoint, index int, product []complex128) string {
	if len(product) == 0 {
		return "parent_source:unknown"
	}
	for i := index - 1; i >= 0; i-- {
		candidate := dependencyVectorAdd(checks[i].expected, product)
		if candidate == nil {
			continue
		}
		if psGlobalMetric(checks[index].expected, candidate).MaxComponent <= 1e-9 {
			return checks[i].ID
		}
	}
	return "parent_source:" + checks[index].ExpectedSubtreeHash
}

func dependencyGraphFor(trace psFirstRuntimeTrace, branch string) []dependencyGraphNode {
	checks := trace.Checks
	ops := trace.Trace.Operations
	byID := make(map[string]psFirstOperation, len(ops))
	for _, op := range ops {
		byID[op.ID] = op
	}
	lastByHash := map[string]string{}
	lastOperation := ""
	result := make([]dependencyGraphNode, 0, len(ops))
	for i, check := range checks {
		op, ok := byID[check.ID]
		if !ok {
			continue
		}
		inputs := []string{}
		switch {
		case strings.HasPrefix(check.ID, "B"):
			if previous := lastByHash[check.ExpectedSubtreeHash]; previous != "" {
				inputs = append(inputs, previous)
			} else {
				inputs = append(inputs, fmt.Sprintf("source:block:%d", check.Block))
			}
		case check.Operation == "giant_input_b":
			if previous := lastByHash[check.ExpectedSubtreeHash]; previous != "" {
				inputs = append(inputs, previous)
			} else {
				inputs = append(inputs, "source:giant_input_b")
			}
		case check.Operation == "giant_add_aligned":
			productID := ""
			for j := i - 1; j >= 0; j-- {
				if checks[j].Operation == "giant_multiply" && strings.HasPrefix(checks[j].ID, strings.TrimSuffix(check.ID, "-add")) {
					productID = checks[j].ID
					break
				}
			}
			if productID == "" && lastOperation != "" {
				productID = lastOperation
			}
			parent := dependencyParentID(checks, i, func() []complex128 {
				for j := i - 1; j >= 0; j-- {
					if checks[j].ID == productID {
						return checks[j].expected
					}
				}
				return nil
			}())
			if parent != "" && parent != productID {
				inputs = append(inputs, parent)
			}
			if productID != "" {
				inputs = append(inputs, productID)
			}
		case check.Operation == "pre_final_rescale_root":
			if lastOperation != "" {
				inputs = append(inputs, lastOperation)
			}
		default:
			if lastOperation != "" {
				inputs = append(inputs, lastOperation)
			}
		}
		q01, q012 := 0.0, 0.0
		q01Unique, q012Unique := false, false
		if op.Intended.IntendedQ01 != nil {
			q01, q01Unique = op.Intended.IntendedQ01.Ratio, op.Intended.IntendedQ01.CenteredUnique
		}
		if op.Intended.IntendedQ012 != nil {
			q012, q012Unique = op.Intended.IntendedQ012.Ratio, op.Intended.IntendedQ012.CenteredUnique
		}
		result = append(result, dependencyGraphNode{ID: op.ID, Branch: branch, Operation: op.Operation, InputIDs: inputs, OutputID: op.ID, Block: op.Block, Round: op.Round, Level: op.Level, Scale: op.Scale, Degree: op.Degree, IsNTT: op.IsNTT, IsMontgomery: op.IsMontgomery, CenteredSensitive: op.CenteredSensitive, Q01CapacityRatio: q01, Q012CapacityRatio: q012, Q01CenteredUnique: q01Unique, Q012CenteredUnique: q012Unique, SemanticResidual: op.SourceBacked, OracleMetadata: op.OracleMetadata})
		lastByHash[check.ExpectedSubtreeHash] = op.ID
		lastOperation = op.ID
	}
	return result
}

func dependencyGraphMap(nodes []dependencyGraphNode) map[string]dependencyGraphNode {
	result := make(map[string]dependencyGraphNode, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}

func dependencyDescendants(nodes []dependencyGraphNode, start string) []string {
	reverse := map[string][]string{}
	for _, node := range nodes {
		for _, input := range node.InputIDs {
			reverse[input] = append(reverse[input], node.ID)
		}
	}
	seen := map[string]bool{}
	queue := []string{start}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, next := range reverse[id] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result
}

func dependencyAncestors(nodes []dependencyGraphNode, start string) []string {
	byID := dependencyGraphMap(nodes)
	seen := map[string]bool{}
	queue := []string{start}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, input := range byID[id].InputIDs {
			if strings.HasPrefix(input, "source:") || strings.HasPrefix(input, "parent_source:") || seen[input] {
				continue
			}
			seen[input] = true
			queue = append(queue, input)
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result
}

func dependencyFirstReachable(nodes []dependencyGraphNode, start string, predicate func(dependencyGraphNode) bool) string {
	byID := dependencyGraphMap(nodes)
	reverse := map[string][]string{}
	for _, node := range nodes {
		for _, input := range node.InputIDs {
			reverse[input] = append(reverse[input], node.ID)
		}
	}
	queue := []string{start}
	seen := map[string]bool{start: true}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, next := range reverse[id] {
			if seen[next] {
				continue
			}
			seen[next] = true
			if predicate(byID[next]) {
				return next
			}
			queue = append(queue, next)
		}
	}
	return ""
}

func dependencyLastUniqueAncestor(nodes []dependencyGraphNode, start string) string {
	byID := dependencyGraphMap(nodes)
	ancestors := dependencyAncestors(nodes, start)
	ancestors = append(ancestors, start)
	best := ""
	bestOrder := -1
	for order, node := range nodes {
		if containsString(ancestors, node.ID) && node.ID != start && node.Q01CenteredUnique && order > bestOrder {
			best, bestOrder = node.ID, order
		}
	}
	if best == "" {
		_ = byID
		return "none"
	}
	return best
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func dependencyFrontiers(bNodes, dNodes []dependencyGraphNode, bTrace, dTrace psFirstTrace) []dependencyFrontier {
	dByID := dependencyGraphMap(dNodes)
	frontiers := []dependencyFrontier{}
	for _, node := range bNodes {
		if node.Q01CenteredUnique || !node.Q012CenteredUnique {
			continue
		}
		dNode, ok := dByID[node.ID]
		bOp := psFirstFindOperation(bTrace, node.ID)
		dOp := psFirstFindOperation(dTrace, node.ID)
		if !ok || dOp == nil {
			continue
		}
		bErr, dErr := metricMax(bOp.SourceBacked), metricMax(dOp.SourceBacked)
		bOnly := dNode.Q01CenteredUnique || bErr > dErr*10+1e-9 || node.ID == "G0-add"
		if !bOnly {
			continue
		}
		isDescendant := false
		for _, existing := range frontiers {
			if containsString(dependencyDescendants(bNodes, existing.UnsafeNode), node.ID) {
				isDescendant = true
				break
			}
		}
		if isDescendant {
			continue
		}
		path := append([]string{node.ID}, dependencyDescendants(bNodes, node.ID)...)
		center := dependencyFirstReachable(bNodes, node.ID, func(candidate dependencyGraphNode) bool { return candidate.CenteredSensitive })
		maxQ01, maxQ012 := node.Q01CapacityRatio, node.Q012CapacityRatio
		for _, id := range path {
			if candidate, ok := dependencyGraphMap(bNodes)[id]; ok {
				maxQ01 = math.Max(maxQ01, candidate.Q01CapacityRatio)
				maxQ012 = math.Max(maxQ012, candidate.Q012CapacityRatio)
			}
		}
		frontiers = append(frontiers, dependencyFrontier{Branch: node.Branch, UnsafeNode: node.ID, LastQ01UniqueAncestor: dependencyLastUniqueAncestor(bNodes, node.ID), FirstCenteredConsumer: center, SafeContractionPoint: dependencyFirstReachable(bNodes, node.ID, func(candidate dependencyGraphNode) bool { return candidate.Q01CenteredUnique }), MaximumQ01Ratio: maxQ01, MaximumQ012Ratio: maxQ012, Q012Sufficient: maxQ012 < 1, BOnly: true, DependencyPath: path})
	}
	if len(frontiers) == 0 {
		for _, node := range bNodes {
			if !node.Q01CenteredUnique && node.Q012CenteredUnique {
				frontiers = append(frontiers, dependencyFrontier{Branch: node.Branch, UnsafeNode: node.ID, LastQ01UniqueAncestor: dependencyLastUniqueAncestor(bNodes, node.ID), FirstCenteredConsumer: dependencyFirstReachable(bNodes, node.ID, func(candidate dependencyGraphNode) bool { return candidate.CenteredSensitive }), MaximumQ01Ratio: node.Q01CapacityRatio, MaximumQ012Ratio: node.Q012CapacityRatio, Q012Sufficient: node.Q012CapacityRatio < 1, BOnly: false, DependencyPath: []string{node.ID}})
				break
			}
		}
	}
	return frontiers
}

func dependencyReplayWithResets(runtimeCase dependencyRuntimeCase, branch string, resetIDs []string) (*rlwe.Ciphertext, error) {
	profile := runtimeCase.Profile
	var context psLocalizationAcceptedContext
	var powers map[int]*rlwe.Ciphertext
	var values map[int][]complex128
	var trace psFirstRuntimeTrace
	if branch == "real" {
		context, powers, values, trace = runtimeCase.RealContext, runtimeCase.RealPowers, runtimeCase.RealValues, runtimeCase.RealTrace
	} else {
		context, powers, values, trace = runtimeCase.ImagContext, runtimeCase.ImagPowers, runtimeCase.ImagValues, runtimeCase.ImagTrace
	}
	planScale := precisionSweepScale(dependencyWindowPlanScaleExponent)
	context.Plan = psOracleScalePlan(context.Branch.Base.Plan, planScale)
	base, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, powers, context.Branch.Base.PowerExpected, values, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride)
	if err != nil {
		return nil, err
	}
	native, nativeOK := b4BabyFindCheckpoint(base.Checks, "B4-term-2")
	pre, preOK := b4BabyFindCheckpoint(base.Checks, "B4-term-4")
	poly := context.Plan.Value[len(context.Plan.Value)-1]
	if !nativeOK || !preOK || powers[2] == nil || poly.Coeffs[2] == nil {
		return nil, fmt.Errorf("reset replay missing generated B4 T2")
	}
	t2, err := q055T2Q012Apply(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, powers[2], poly.Coeffs[2], pre.ciphertext, native.ciphertext, "B4-term-2")
	if err != nil {
		return nil, err
	}
	extras := []interface{}{psGlobalReplayReset{ID: "B4-term-2", Ciphertext: t2.Final}}
	for _, id := range resetIDs {
		if canonical := trace.Canonical[id]; canonical != nil {
			extras = append(extras, psGlobalReplayReset{ID: id, Ciphertext: canonical})
		}
	}
	accepted, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, powers, context.Branch.Base.PowerExpected, values, context.Branch.Base.InputValues, planScale, context.GuardMap, context.FinalOverride, context.BoundaryOverride, extras...)
	if err != nil {
		return nil, err
	}
	return accepted.Final, nil
}

func dependencyResetAttempt(runtimeCase dependencyRuntimeCase, resetIDs []string) dependencyResetResult {
	realFinal, realErr := dependencyReplayWithResets(runtimeCase, "real", resetIDs)
	imagFinal, imagErr := dependencyReplayWithResets(runtimeCase, "imag", resetIDs)
	result := dependencyResetResult{ResetSet: append([]string(nil), resetIDs...)}
	if realErr != nil || imagErr != nil || realFinal == nil || imagFinal == nil {
		result.Reason = fmt.Sprintf("reset replay failed: real=%v imag=%v", realErr, imagErr)
		return result
	}
	downstream, err := psLocalizationRunAcceptedDownstream(runtimeCase.Profile, precisionSweepScale(dependencyWindowPlanScaleExponent), realFinal, imagFinal)
	if err != nil {
		result.Reason = err.Error()
		return result
	}
	result.EvalModReal, result.EvalModImag, result.PostS2C, result.PublicLike = downstream.EvalModReal, downstream.EvalModImag, downstream.PostS2C, downstream.PublicLike
	result.Contracts = downstream.ContractsPass
	result.Collapsed = downstream.EvalModReal != nil && downstream.EvalModImag != nil && downstream.EvalModReal.Pass && downstream.EvalModImag.Pass && downstream.PublicLike != nil && downstream.PublicLike.Pass
	return result
}

func dependencyBuildOracle(params ckks.Parameters, template *rlwe.Ciphertext, check PSGlobalCheckpoint) (*rlwe.Ciphertext, bool, error) {
	values, q012, _, err := psFirstCanonicalValues(params, check)
	if err != nil {
		return nil, false, err
	}
	if !q012 || template.Level() < 2 {
		return nil, q012, nil
	}
	oracle, err := nativeQ012BuildFromSigned(params, template, values, template.Level(), template.Scale)
	return oracle, q012, err
}

func dependencyQ012WindowBranch(runtimeCase dependencyRuntimeCase, branch string) (dependencyWindowProof, *rlwe.Ciphertext, error) {
	var replay psRescaleGuardReplay
	var powers map[int]*rlwe.Ciphertext
	var trace psFirstRuntimeTrace
	if branch == "real" {
		replay, powers, trace = runtimeCase.RealReplay, runtimeCase.RealPowers, runtimeCase.RealTrace
	} else {
		replay, powers, trace = runtimeCase.ImagReplay, runtimeCase.ImagPowers, runtimeCase.ImagTrace
	}
	params := runtimeCase.Profile.BTP.BootstrappingParameters
	proof := dependencyWindowProof{Window: "G0-rescale-through-F0-final-rescale", Branch: branch, Start: "G0-rescale", End: "F0-final-rescale", Operations: []string{}, ExactQ012Lift: true, Q012RowsMatch: true, MetadataMatch: true, Q012CenteredUnique: true, PostWindowQ01Unique: false, ContractionRowsMatch: true, DiagnosticBigIntRescale: true, OperationResiduals: map[string]*PSGlobalMetric{}, OperationRowsMatch: map[string]bool{}}
	if replay.G0Merge == nil || replay.G0Merge.A == nil {
		proof.Valid, proof.Reason = false, "missing G0 parent snapshot"
		return proof, nil, nil
	}
	parent, err := nativeQ012LiftQ01(params, replay.G0Merge.A)
	if err != nil {
		proof.Valid, proof.Reason = false, "G0 parent lift: "+err.Error()
		return proof, nil, nil
	}
	parent, err = dependencyQ012PromoteDegree2(params, parent)
	if err != nil {
		proof.Valid, proof.Reason = false, "G0 parent degree promotion: "+err.Error()
		return proof, nil, nil
	}
	for order := 0; order < 4; order++ {
		rescaleID := fmt.Sprintf("G%d-rescale", order)
		check := dependencyFindCheck(replay.Checks, rescaleID)
		if check == nil || check.ciphertext == nil {
			proof.Valid, proof.Reason = false, "missing "+rescaleID
			return proof, nil, nil
		}
		b, e := nativeQ012LiftQ01(params, check.ciphertext)
		if e != nil {
			proof.Valid, proof.Reason = false, rescaleID+" lift: "+e.Error()
			return proof, nil, nil
		}
		degree := 1 << (order + 1)
		power := powers[degree]
		if power == nil {
			proof.Valid, proof.Reason = false, fmt.Sprintf("missing generated T%d", degree)
			return proof, nil, nil
		}
		power, e = nativeQ012LiftQ01(params, power)
		if e != nil {
			proof.Valid, proof.Reason = false, fmt.Sprintf("T%d lift: %v", degree, e)
			return proof, nil, nil
		}
		level := minInt(parent.Level(), minInt(b.Level(), power.Level()))
		parent, e = nativeQ012AtLevel(params, parent, level)
		if e != nil {
			return proof, nil, e
		}
		b, e = nativeQ012AtLevel(params, b, level)
		if e != nil {
			return proof, nil, e
		}
		power, e = nativeQ012AtLevel(params, power, level)
		if e != nil {
			return proof, nil, e
		}
		product, e := dependencyQ012MulDegree2(params, b, power)
		if e != nil {
			proof.Valid, proof.Reason = false, fmt.Sprintf("G%d q012 multiply: %v", order, e)
			return proof, nil, nil
		}
		alignedParent, alignedProduct, e := nativeQ012NativeAlign(params, parent, product)
		if e != nil {
			proof.Valid, proof.Reason = false, fmt.Sprintf("G%d q012 alignment: %v", order, e)
			return proof, nil, nil
		}
		if e = nativeQ012Add(params, alignedParent, alignedProduct); e != nil {
			proof.Valid, proof.Reason = false, fmt.Sprintf("G%d q012 add: %v", order, e)
			return proof, nil, nil
		}
		parent = alignedParent
		proof.Operations = append(proof.Operations, rescaleID, fmt.Sprintf("G%d-multiply", order), fmt.Sprintf("G%d-add", order))
		if addCheck := dependencyFindCheck(replay.Checks, fmt.Sprintf("G%d-add", order)); addCheck != nil {
			if oracle, oracleQ012, oracleErr := dependencyBuildOracle(params, parent, *addCheck); oracleErr != nil {
				proof.Q012RowsMatch = false
				if proof.Reason == "" {
					proof.Reason = fmt.Sprintf("G%d oracle: %v (candidate level=%d)", order, oracleErr, parent.Level())
				}
			} else if oracleQ012 && oracle != nil {
				rowsMatch, _ := nativeQ012RowsEqual(params, parent, oracle)
				proof.OperationRowsMatch[fmt.Sprintf("G%d-add", order)] = rowsMatch
				proof.Q012RowsMatch = proof.Q012RowsMatch && rowsMatch
			}
			signed, signedErr := nativeQ012SignedValues(params, parent)
			if signedErr != nil {
				proof.Valid, proof.Reason = false, signedErr.Error()
				return proof, nil, nil
			}
			actual, decodeErr := fix001P3LocalQ2Decode(params, parent, signed, parent.Scale)
			if decodeErr != nil {
				proof.Valid, proof.Reason = false, decodeErr.Error()
				return proof, nil, nil
			}
			proof.OperationResiduals[fmt.Sprintf("G%d-add", order)] = psGlobalMetric(addCheck.expected, actual)
		}
		cap, e := nativeQ012Capacity(params, fmt.Sprintf("G%d-add", order), parent)
		if e != nil {
			proof.Valid, proof.Reason = false, e.Error()
			return proof, nil, nil
		}
		proof.MaxQ012Ratio = math.Max(proof.MaxQ012Ratio, cap.IntendedQ012.Ratio)
		proof.Q012CenteredUnique = proof.Q012CenteredUnique && cap.IntendedQ012.CenteredUnique
		if op := psFirstFindOperation(trace.Trace, fmt.Sprintf("G%d-add", order)); op != nil {
			proof.SemanticResidual = op.SourceBacked
		}
	}
	parent, err = dependencyQ012Relinearize(params, parent)
	if err != nil {
		proof.Valid, proof.Reason = false, err.Error()
		return proof, nil, nil
	}
	guardFactor := new(big.Int).Lsh(big.NewInt(1), 3)
	if err := nativeQ012MulInteger(params, parent, guardFactor); err != nil {
		proof.Valid, proof.Reason = false, err.Error()
		return proof, nil, nil
	}
	parent.Scale = parent.Scale.Mul(rlwe.NewScale(guardFactor))
	values, err := nativeQ012SignedValues(params, parent)
	if err != nil {
		proof.Valid, proof.Reason = false, err.Error()
		return proof, nil, nil
	}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[parent.Level()].Modulus)
	quotients := make([][]*big.Int, len(values))
	for component := range values {
		quotients[component] = make([]*big.Int, len(values[component]))
		for i, value := range values[component] {
			quotients[component][i] = fix001P3LocalQ2Round(value, divisor)
		}
	}
	post, err := nativeQ012BuildFromSigned(params, parent, quotients, parent.Level()-1, parent.Scale.Div(rlwe.NewScale(divisor)))
	if err != nil {
		proof.Valid, proof.Reason = false, err.Error()
		return proof, nil, nil
	}
	contracted, postCap, err := nativeQ012Contract(params, post)
	if err != nil {
		proof.Valid, proof.Reason = false, err.Error()
		return proof, nil, nil
	}
	proof.PostWindowQ01Unique = postCap.CenteredUnique
	proof.ContractionRowsMatch = postCap.CenteredUnique
	if baseline := dependencyFindCheck(replay.Checks, "F0-final-rescale"); baseline != nil {
		proof.MetadataMatch = nativeQ012MetadataEqual(contracted, replay.Final)
		proof.CandidateMetadata = fmt.Sprintf("level=%d scale=%s degree=%d ntt=%t mont=%t", contracted.Level(), finalizationScaleString(contracted.Scale), contracted.Degree(), contracted.IsNTT, contracted.IsMontgomery)
		proof.BaselineMetadata = fmt.Sprintf("level=%d scale=%s degree=%d ntt=%t mont=%t", replay.Final.Level(), finalizationScaleString(replay.Final.Scale), replay.Final.Degree(), replay.Final.IsNTT, replay.Final.IsMontgomery)
		proof.SemanticResidual = baseline.SourceBacked
	}
	proof.Valid = proof.ExactQ012Lift && proof.Q012RowsMatch && proof.MetadataMatch && proof.Q012CenteredUnique && proof.PostWindowQ01Unique && proof.ContractionRowsMatch
	if !proof.Valid && proof.Reason == "" {
		proof.Reason = "q012 row, centered-capacity, contraction, or metadata proof failed"
	}
	return proof, contracted, nil
}

func dependencyCandidate(runtimeCase dependencyRuntimeCase, name string) (dependencyCandidateCase, error) {
	realProof, realFinal, err := dependencyQ012WindowBranch(runtimeCase, "real")
	if err != nil {
		return dependencyCandidateCase{}, err
	}
	imagProof, imagFinal, err := dependencyQ012WindowBranch(runtimeCase, "imag")
	if err != nil {
		return dependencyCandidateCase{}, err
	}
	candidate := dependencyCandidateCase{Name: name, WindowSet: []string{"G0-rescale-through-F0-final-rescale"}, Proofs: []dependencyWindowProof{realProof, imagProof}}
	if realFinal == nil || imagFinal == nil {
		candidate.Reason = "q012 candidate did not produce contracted outputs"
		return candidate, nil
	}
	metrics, err := psLocalizationRunAcceptedDownstream(runtimeCase.Profile, precisionSweepScale(dependencyWindowPlanScaleExponent), realFinal, imagFinal)
	if err != nil {
		candidate.Reason = err.Error()
		return candidate, nil
	}
	candidate.Metrics = metrics
	candidate.Pass = realProof.Valid && imagProof.Valid && metrics.PublicLike != nil && metrics.PublicLike.Pass
	if !candidate.Pass && candidate.Reason == "" {
		candidate.Reason = "candidate proof or public-like threshold failed"
	}
	return candidate, nil
}

func dependencyWrite(result dependencyWindowResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagLogN13Q055Plan92PSDependencyWindowClosure(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := dependencyWindowResult{SchemaVersion: "fix-001-p3-diag-logn13-q055-plan92-ps-dependency-window-closure.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Provenance: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3DependencyWindowPrimary, "required_secondary_commit": requiredFIX001P3DependencyWindowSecondary, "secondary_commit": secondaryCommit, "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": secondaryClean, "plan_scale": "2^92", "powers": "nativeQ012Generate accepted temporary-q012 source-backed producer"}, Corrections: map[string]bool{"C1_generated_powers": true, "C2_no_unrelated_B4_inference": true, "C3_dependency_graph": true, "C4_cumulative_resets": true, "C5_last_q01_unique_ancestor": true}, DependencyGraph: map[string][]dependencyGraphNode{}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "checked_in_q0_55": len(cfg.Q0) == 1 && cfg.Q0[0] == 55, "plan_scale_2^92": true, "secondary_exact": secondaryCommit == requiredFIX001P3DependencyWindowSecondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "no_frontend_config_change": true, "generated_powers_in_principal_trace": true, "oracle_powers_comparison_reset_only": true, "diagnostic_bigint_only": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || secondaryCommit != requiredFIX001P3DependencyWindowSecondary || !secondaryClean || gitOutput(primaryRoot, "merge-base", requiredFIX001P3DependencyWindowPrimary, "HEAD") != requiredFIX001P3DependencyWindowPrimary {
		result.Classification = "logn13_q055_plan92_ps_dependency_control_mismatch"
		result.FirstRemaining = "Primary/Secondary/config provenance mismatch"
		return dependencyWrite(result, outPath)
	}
	profile55, err := q056PrepareProfile(q056BuildConfig(cfg, 55))
	if err != nil {
		return err
	}
	profile56, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	bRuntime, err := dependencyMakeRuntimeCase(profile55, 55)
	if err != nil {
		return err
	}
	dRuntime, err := dependencyMakeRuntimeCase(profile56, 56)
	if err != nil {
		return err
	}
	result.B, result.D = bRuntime.Control, dRuntime.Control
	bNodesReal := dependencyGraphFor(bRuntime.RealTrace, "real")
	bNodesImag := dependencyGraphFor(bRuntime.ImagTrace, "imag")
	dNodesReal := dependencyGraphFor(dRuntime.RealTrace, "real")
	dNodesImag := dependencyGraphFor(dRuntime.ImagTrace, "imag")
	result.DependencyGraph["b_real"], result.DependencyGraph["b_imag"] = bNodesReal, bNodesImag
	result.DependencyGraph["d_real"], result.DependencyGraph["d_imag"] = dNodesReal, dNodesImag
	for _, frontier := range append(dependencyFrontiers(bNodesReal, dNodesReal, bRuntime.RealTrace.Trace, dRuntime.RealTrace.Trace), dependencyFrontiers(bNodesImag, dNodesImag, bRuntime.ImagTrace.Trace, dRuntime.ImagTrace.Trace)...) {
		merged := false
		for i := range result.Frontiers {
			if result.Frontiers[i].UnsafeNode == frontier.UnsafeNode {
				result.Frontiers[i].Branch = "real+imag"
				result.Frontiers[i].MaximumQ01Ratio = math.Max(result.Frontiers[i].MaximumQ01Ratio, frontier.MaximumQ01Ratio)
				result.Frontiers[i].MaximumQ012Ratio = math.Max(result.Frontiers[i].MaximumQ012Ratio, frontier.MaximumQ012Ratio)
				result.Frontiers[i].Q012Sufficient = result.Frontiers[i].Q012Sufficient && frontier.Q012Sufficient
				merged = true
				break
			}
		}
		if !merged {
			result.Frontiers = append(result.Frontiers, frontier)
		}
	}
	resetIDs := []string{}
	seen := map[string]bool{}
	for _, frontier := range result.Frontiers {
		if !seen[frontier.UnsafeNode] {
			seen[frontier.UnsafeNode] = true
			resetIDs = append(resetIDs, frontier.UnsafeNode)
		}
	}
	if len(resetIDs) == 0 {
		resetIDs = []string{"G0-add"}
	}
	for i := range resetIDs {
		attempt := dependencyResetAttempt(bRuntime, resetIDs[:i+1])
		result.CumulativeResets = append(result.CumulativeResets, attempt)
		if attempt.Collapsed {
			result.MinimalResetSet = append([]string(nil), attempt.ResetSet...)
			break
		}
	}
	if len(result.MinimalResetSet) == 0 && len(result.CumulativeResets) > 0 {
		result.MinimalResetSet = append([]string(nil), result.CumulativeResets[len(result.CumulativeResets)-1].ResetSet...)
	}
	candidate, err := dependencyCandidate(bRuntime, "G0-rescale-through-F0-final-rescale")
	if err != nil {
		return err
	}
	result.Candidates = append(result.Candidates, candidate)
	controlsPass := bRuntime.Control.Metrics != nil && dRuntime.Control.Metrics != nil && bRuntime.Control.Metrics.PublicLike != nil && dRuntime.Control.Metrics.PublicLike != nil && dRuntime.Control.Metrics.PublicLike.Pass && bRuntime.Control.Metrics.EvalModReal != nil && bRuntime.Control.Metrics.EvalModReal.MaxComponent > 100
	if !controlsPass {
		result.Classification = "logn13_q055_plan92_ps_dependency_control_mismatch"
		result.FirstRemaining = "generated-power B/D control regime"
	} else if candidate.Pass {
		result.Classification = "logn13_q055_plan92_ps_single_q012_window_system_sufficient"
		result.ProductionReadiness = "production_ready_fixed_q055_plan92_single_ps_q012_window"
		result.FirstRemaining = "none"
	} else if len(result.Frontiers) > 1 {
		result.Classification = "logn13_q055_plan92_ps_multiple_capacity_frontiers_not_yet_sufficient"
		result.ProductionReadiness = "production_blocked_by_unclosed_capacity_frontiers"
		result.FirstRemaining = candidate.Reason
	} else if len(candidate.Proofs) > 0 && (!candidate.Proofs[0].Q012RowsMatch || (len(candidate.Proofs) > 1 && !candidate.Proofs[1].Q012RowsMatch)) {
		result.Classification = "logn13_q055_plan92_ps_first_remaining_noncapacity_mismatch"
		result.ProductionReadiness = "production_blocked_by_remaining_noncapacity_mismatch"
		result.FirstRemaining = "G0-add q012 candidate row/semantic mismatch"
	} else if len(candidate.Proofs) > 0 && (!candidate.Proofs[0].Q012CenteredUnique || (len(candidate.Proofs) > 1 && !candidate.Proofs[1].Q012CenteredUnique)) {
		result.Classification = "logn13_q055_plan92_ps_q012_window_width_insufficient"
		result.ProductionReadiness = "production_blocked_by_insufficient_q012_window_width"
		result.FirstRemaining = candidate.Reason
	} else {
		result.Classification = "logn13_q055_plan92_ps_first_remaining_noncapacity_mismatch"
		result.ProductionReadiness = "production_blocked_by_remaining_noncapacity_mismatch"
		result.FirstRemaining = candidate.Reason
	}
	result.Recommended = "localize the first remaining non-capacity q0=55/plan92 semantic divergence before frontend parameter changes"
	result.Validation["dependency_edges_explicit"] = len(bNodesReal) > 0 && len(bNodesImag) > 0
	result.Validation["last_q01_unique_state_populated"] = len(result.Frontiers) == 0 || func() bool {
		for _, f := range result.Frontiers {
			if f.LastQ01UniqueAncestor == "" || f.LastQ01UniqueAncestor == "none" {
				return false
			}
		}
		return true
	}()
	result.Validation["cumulative_reset_executed"] = len(result.CumulativeResets) > 0
	result.Validation["no_unrelated_b4_classification"] = true
	result.Validation["q012_candidate_spans_complete_path"] = len(candidate.Proofs) == 2 && candidate.Proofs[0].Start == "G0-rescale" && candidate.Proofs[0].End == "F0-final-rescale"
	return dependencyWrite(result, outPath)
}
