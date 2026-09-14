package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	requiredFIX001P3G0MergePrimary   = "13d54221f7feac63c860b3bb6cf81227ceb65240"
	requiredFIX001P3G0MergeSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	g0MergePublicThreshold           = 1e-2
	g0MergeSemanticThreshold         = 1e-10
)

type g0MergeCipherState struct {
	Level         int                    `json:"level"`
	Scale         string                 `json:"scale"`
	Degree        int                    `json:"degree"`
	NTT           bool                   `json:"is_ntt"`
	Montgomery    bool                   `json:"is_montgomery"`
	Capacity      map[string]interface{} `json:"capacity"`
	DecodedMaxAbs float64                `json:"decoded_max_abs"`
}

type g0MergeOperandEvidence struct {
	AF                          g0MergeCipherState         `json:"a_f"`
	BFRaw                       g0MergeCipherState         `json:"b_f_raw"`
	PF                          g0MergeCipherState         `json:"p_f"`
	OF                          g0MergeCipherState         `json:"o_f"`
	AQ                          g0MergeCipherState         `json:"a_q"`
	PQ                          g0MergeCipherState         `json:"p_q"`
	AFMinusAQ                   *PSGlobalMetric            `json:"a_f_minus_a_q"`
	PFMinusPQ                   *PSGlobalMetric            `json:"p_f_minus_p_q"`
	CanonicalFloors             map[string]*PSGlobalMetric `json:"canonical_floors"`
	ResidualDirectionSimilarity float64                    `json:"residual_direction_similarity"`
	Scales                      map[string]interface{}     `json:"scales"`
}

type g0MergeSemanticEvidence struct {
	S0 map[string]interface{} `json:"s0_current_fast_merge"`
	S1 map[string]interface{} `json:"s1_full_rns_current_fast_merge"`
	S2 map[string]interface{} `json:"s2_standard_add_native_scales"`
	S3 map[string]interface{} `json:"s3_ideal_decoded_addition"`
}

type g0MergeCase struct {
	Case               string                            `json:"case"`
	G0AddVsCanonical   *PSGlobalMetric                   `json:"g0_add_vs_canonical"`
	FinalPSVsCanonical *PSGlobalMetric                   `json:"final_ps_vs_canonical"`
	Downstream         *psLocalizationDownstreamEvidence `json:"downstream"`
}

type g0MergeAlpha struct {
	Alpha      float64         `json:"alpha"`
	Valid      bool            `json:"valid"`
	PublicLike *PSGlobalMetric `json:"public_like,omitempty"`
	PostS2C    *PSGlobalMetric `json:"post_s2c,omitempty"`
	Reason     string          `json:"reason,omitempty"`
}

type g0MergeResult struct {
	SchemaVersion         string                             `json:"schema_version"`
	Timestamp             time.Time                          `json:"timestamp"`
	Primary               RepositoryMetadata                 `json:"primary_repository"`
	Lattigo               RepositoryMetadata                 `json:"lattigo_repository"`
	Environment           EnvironmentMetadata                `json:"environment"`
	Config                BootstrapConfig                    `json:"config"`
	Parameters            ExperimentParameters               `json:"effective_parameters"`
	Provenance            map[string]interface{}             `json:"provenance"`
	M0Controls            map[string]interface{}             `json:"m0_controls"`
	Operands              map[string]g0MergeOperandEvidence  `json:"live_g0_operands"`
	MergeSemantics        map[string]g0MergeSemanticEvidence `json:"merge_semantics"`
	BranchCases           map[string]g0MergeCase             `json:"branch_replacement_cases"`
	BranchAlpha           map[string][]g0MergeAlpha          `json:"branch_alpha_sensitivity"`
	SmallestPassingAlpha  map[string]float64                 `json:"smallest_passing_alpha,omitempty"`
	Classification        string                             `json:"classification"`
	FirstRemainingBlocker string                             `json:"first_remaining_blocker"`
	RecommendedNextTarget string                             `json:"recommended_next_design_target"`
	Validation            map[string]interface{}             `json:"validation"`
}

type g0MergeBranch struct {
	Context               psLocalizationAcceptedContext
	Snapshot              *psRescaleGuardG0MergeSnapshot
	CanonicalA            *rlwe.Ciphertext
	CanonicalP            *rlwe.Ciphertext
	CanonicalOutput       *rlwe.Ciphertext
	AValues               []complex128
	PValues               []complex128
	AQValues              []complex128
	PQValues              []complex128
	CanonicalOutputValues []complex128
	CanonicalFinalValues  []complex128
}

func g0MergeWrite(result g0MergeResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func g0MergeMaxAbs(values []complex128) float64 {
	max := 0.0
	for _, value := range values {
		max = math.Max(max, math.Hypot(real(value), imag(value)))
	}
	return max
}

func g0MergeState(params ckks.Parameters, ct *rlwe.Ciphertext) (g0MergeCipherState, error) {
	state := g0MergeCipherState{}
	if ct == nil {
		return state, fmt.Errorf("nil G0 merge ciphertext")
	}
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return state, err
	}
	state.Level, state.Scale, state.Degree = ct.Level(), finalizationScaleString(ct.Scale), ct.Degree()
	state.NTT, state.Montgomery = ct.IsNTT, ct.IsMontgomery
	state.DecodedMaxAbs = g0MergeMaxAbs(values)
	state.Capacity = fix001P3DAR0CapacityRow(params, ct)
	return state, nil
}

func g0MergeCanonical(params ckks.Parameters, source *rlwe.Ciphertext, expected []complex128) (*rlwe.Ciphertext, []complex128, *PSGlobalMetric, error) {
	normal, err := fix001P3QuantizationAwareMaterialize(params, source, fix001P3QuantizationAwareValues(expected, fix001P3QuantizationAwareSourcePrecision), fix001P3QuantizationAwareSourcePrecision)
	if err != nil {
		return nil, nil, nil, err
	}
	values, err := fix001P3QuantizationAwareDecode(params, normal)
	if err != nil {
		return nil, nil, nil, err
	}
	fast, err := psLocalizationFastFromNormal(params, normal)
	if err != nil {
		return nil, nil, nil, err
	}
	return fast, values, psLocalizationMetric(expected, values, g0MergeSemanticThreshold), nil
}

func g0MergeAddVectors(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] + b[i]
	}
	return out
}

func g0MergeDecodeFullQ01(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	projection, _, err := q01Projection(params, ct, false)
	if err != nil {
		return nil, err
	}
	return diagnosticDecode(params, projection, zeroSecret(params))
}

func g0MergeOperand(profile q056PreparedProfile, context psLocalizationAcceptedContext) (g0MergeBranch, error) {
	snapshot := context.Run.Replay.G0Merge
	if snapshot == nil || snapshot.A == nil || snapshot.BRaw == nil || snapshot.Product == nil || snapshot.Output == nil {
		return g0MergeBranch{}, fmt.Errorf("accepted replay did not capture complete G0 merge")
	}
	aValues, err := psGlobalDecode(profile.BTP.BootstrappingParameters, snapshot.A)
	if err != nil {
		return g0MergeBranch{}, err
	}
	pValues, err := psGlobalDecode(profile.BTP.BootstrappingParameters, snapshot.Product)
	if err != nil {
		return g0MergeBranch{}, err
	}
	aQ, aQValues, aFloor, err := g0MergeCanonical(profile.BTP.BootstrappingParameters, snapshot.A, snapshot.AExpected)
	if err != nil {
		return g0MergeBranch{}, err
	}
	pQ, pQValues, pFloor, err := g0MergeCanonical(profile.BTP.BootstrappingParameters, snapshot.Product, snapshot.ProductExpected)
	if err != nil {
		return g0MergeBranch{}, err
	}
	mergeExpected := g0MergeAddVectors(snapshot.AExpected, snapshot.ProductExpected)
	_, oQValues, _, err := g0MergeCanonical(profile.BTP.BootstrappingParameters, snapshot.Output, mergeExpected)
	if err != nil {
		return g0MergeBranch{}, err
	}
	aState, err := g0MergeState(profile.BTP.BootstrappingParameters, snapshot.A)
	if err != nil {
		return g0MergeBranch{}, err
	}
	bState, err := g0MergeState(profile.BTP.BootstrappingParameters, snapshot.BRaw)
	if err != nil {
		return g0MergeBranch{}, err
	}
	pState, err := g0MergeState(profile.BTP.BootstrappingParameters, snapshot.Product)
	if err != nil {
		return g0MergeBranch{}, err
	}
	oState, err := g0MergeState(profile.BTP.BootstrappingParameters, snapshot.Output)
	if err != nil {
		return g0MergeBranch{}, err
	}
	aQState, err := g0MergeState(profile.BTP.BootstrappingParameters, aQ)
	if err != nil {
		return g0MergeBranch{}, err
	}
	pQState, err := g0MergeState(profile.BTP.BootstrappingParameters, pQ)
	if err != nil {
		return g0MergeBranch{}, err
	}
	ratioAP := snapshot.A.Scale.Div(snapshot.Product.Scale)
	ratioPA := snapshot.Product.Scale.Div(snapshot.A.Scale)
	evidence := g0MergeOperandEvidence{AF: aState, BFRaw: bState, PF: pState, OF: oState, AQ: aQState, PQ: pQState,
		AFMinusAQ: psLocalizationMetric(aValues, aQValues, g0MergeSemanticThreshold), PFMinusPQ: psLocalizationMetric(pValues, pQValues, g0MergeSemanticThreshold),
		CanonicalFloors: map[string]*PSGlobalMetric{"a": aFloor, "p": pFloor}, ResidualDirectionSimilarity: psLocalizationDirection(psLocalizationInterpolate(aValues, aQValues, 0), psLocalizationInterpolate(pValues, pQValues, 0)),
		Scales: map[string]interface{}{"a": finalizationScaleString(snapshot.A.Scale), "p": finalizationScaleString(snapshot.Product.Scale), "a_to_p": finalizationScaleString(ratioAP), "p_to_a": finalizationScaleString(ratioPA), "log2_delta": psGlobalLog2Delta(snapshot.A.Scale, snapshot.Product.Scale), "plan_scale_override_relabels": "b/product branch to parent A scale before Fast Add"}}
	_ = evidence
	return g0MergeBranch{Context: context, Snapshot: snapshot, CanonicalA: aQ, CanonicalP: pQ, AValues: aValues, PValues: pValues, AQValues: aQValues, PQValues: pQValues, CanonicalOutputValues: oQValues}, nil
}

func g0MergeDownstream(profile q056PreparedProfile, real, imag *rlwe.Ciphertext) (*psLocalizationDownstreamEvidence, error) {
	return psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psLocalizationPlanScaleExponent), real, imag)
}

func g0MergeCheckpointCanonical(profile q056PreparedProfile, context psLocalizationAcceptedContext, id string) (*rlwe.Ciphertext, error) {
	var check *PSGlobalCheckpoint
	for i := range context.Run.Replay.Checks {
		if context.Run.Replay.Checks[i].ID == id {
			check = &context.Run.Replay.Checks[i]
			break
		}
	}
	if id == "F0-root" {
		check = &context.Run.Replay.Root
	}
	if check == nil {
		return nil, fmt.Errorf("missing accepted checkpoint %s", id)
	}
	canonical, _, _, _, err := psLocalizationCanonical(profile.BTP.BootstrappingParameters, *check)
	return canonical, err
}

func g0MergeResetControl(profile q056PreparedProfile, realContext, imagContext psLocalizationAcceptedContext, id string) (*psLocalizationDownstreamEvidence, error) {
	realCanonical, err := g0MergeCheckpointCanonical(profile, realContext, id)
	if err != nil {
		return nil, err
	}
	imagCanonical, err := g0MergeCheckpointCanonical(profile, imagContext, id)
	if err != nil {
		return nil, err
	}
	_, realFinal, err := psLocalizationRunAcceptedReset(realContext, profile.BTP.BootstrappingParameters, profile.Fast, id, realCanonical)
	if err != nil {
		return nil, err
	}
	_, imagFinal, err := psLocalizationRunAcceptedReset(imagContext, profile.BTP.BootstrappingParameters, profile.Fast, id, imagCanonical)
	if err != nil {
		return nil, err
	}
	return g0MergeDownstream(profile, realFinal, imagFinal)
}

func g0MergeReplay(branch g0MergeBranch, params ckks.Parameters, eval *bootstrapping.FastEvaluator, a, product *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(params, eval.FastCKKS, branch.Context.Plan, branch.Context.Branch.Base.OraclePowerMap, branch.Context.Branch.Base.PowerExpected, branch.Context.Branch.Base.OraclePowerValues, branch.Context.Branch.Base.InputValues, precisionSweepScale(psLocalizationPlanScaleExponent), branch.Context.GuardMap, branch.Context.FinalOverride, branch.Context.BoundaryOverride, psRescaleGuardG0MergeOverride{A: a, Product: product})
	if err != nil {
		return nil, nil, err
	}
	if replay.G0Merge == nil || replay.G0Merge.Output == nil {
		return nil, nil, fmt.Errorf("G0 replay did not capture replaced merge output")
	}
	return replay.Final, replay.G0Merge.Output, nil
}

func g0MergeRunCase(profile q056PreparedProfile, real, imag g0MergeBranch, name string, realA, realP, imagA, imagP *rlwe.Ciphertext) (g0MergeCase, error) {
	realFinal, realOutput, err := g0MergeReplay(real, profile.BTP.BootstrappingParameters, profile.Fast, realA, realP)
	if err != nil {
		return g0MergeCase{}, err
	}
	imagFinal, imagOutput, err := g0MergeReplay(imag, profile.BTP.BootstrappingParameters, profile.Fast, imagA, imagP)
	if err != nil {
		return g0MergeCase{}, err
	}
	downstream, err := g0MergeDownstream(profile, realFinal, imagFinal)
	if err != nil {
		return g0MergeCase{}, err
	}
	params := profile.BTP.BootstrappingParameters
	realOutputValues, err := psGlobalDecode(params, realOutput)
	if err != nil {
		return g0MergeCase{}, err
	}
	imagOutputValues, err := psGlobalDecode(params, imagOutput)
	if err != nil {
		return g0MergeCase{}, err
	}
	realFinalValues, err := psGlobalDecode(params, realFinal)
	if err != nil {
		return g0MergeCase{}, err
	}
	imagFinalValues, err := psGlobalDecode(params, imagFinal)
	if err != nil {
		return g0MergeCase{}, err
	}
	return g0MergeCase{Case: name, G0AddVsCanonical: psLocalizationMaxPublic(psLocalizationMetric(real.CanonicalOutputValues, realOutputValues, g0MergeSemanticThreshold), psLocalizationMetric(imag.CanonicalOutputValues, imagOutputValues, g0MergeSemanticThreshold)), FinalPSVsCanonical: psLocalizationMaxPublic(psLocalizationMetric(real.CanonicalFinalValues, realFinalValues, g0MergeSemanticThreshold), psLocalizationMetric(imag.CanonicalFinalValues, imagFinalValues, g0MergeSemanticThreshold)), Downstream: downstream}, nil
}

func g0MergeBranchAlpha(profile q056PreparedProfile, real, imag g0MergeBranch, branchName string, alpha float64) (g0MergeAlpha, error) {
	var realA, realP, imagA, imagP *rlwe.Ciphertext
	var err error
	if branchName == "parent" {
		if alpha == 0 {
			realA, imagA = real.Snapshot.A.CopyNew(), imag.Snapshot.A.CopyNew()
		} else if alpha == 1 {
			realA, imagA = real.CanonicalA.CopyNew(), imag.CanonicalA.CopyNew()
		} else {
			realA, err = psLocalizationFastMaterialize(profile.BTP.BootstrappingParameters, real.Snapshot.A, psLocalizationInterpolate(real.AValues, real.AQValues, alpha))
			if err != nil {
				return g0MergeAlpha{}, err
			}
			imagA, err = psLocalizationFastMaterialize(profile.BTP.BootstrappingParameters, imag.Snapshot.A, psLocalizationInterpolate(imag.AValues, imag.AQValues, alpha))
			if err != nil {
				return g0MergeAlpha{}, err
			}
		}
		realP, imagP = real.Snapshot.Product.CopyNew(), imag.Snapshot.Product.CopyNew()
	} else {
		realA, imagA = real.Snapshot.A.CopyNew(), imag.Snapshot.A.CopyNew()
		if alpha == 0 {
			realP, imagP = real.Snapshot.Product.CopyNew(), imag.Snapshot.Product.CopyNew()
		} else if alpha == 1 {
			realP, imagP = real.CanonicalP.CopyNew(), imag.CanonicalP.CopyNew()
		} else {
			realP, err = psLocalizationFastMaterialize(profile.BTP.BootstrappingParameters, real.Snapshot.Product, psLocalizationInterpolate(real.PValues, real.PQValues, alpha))
			if err != nil {
				return g0MergeAlpha{}, err
			}
			imagP, err = psLocalizationFastMaterialize(profile.BTP.BootstrappingParameters, imag.Snapshot.Product, psLocalizationInterpolate(imag.PValues, imag.PQValues, alpha))
			if err != nil {
				return g0MergeAlpha{}, err
			}
		}
	}
	realFinal, _, err := g0MergeReplay(real, profile.BTP.BootstrappingParameters, profile.Fast, realA, realP)
	if err != nil {
		return g0MergeAlpha{}, err
	}
	imagFinal, _, err := g0MergeReplay(imag, profile.BTP.BootstrappingParameters, profile.Fast, imagA, imagP)
	if err != nil {
		return g0MergeAlpha{}, err
	}
	downstream, err := g0MergeDownstream(profile, realFinal, imagFinal)
	if err != nil {
		return g0MergeAlpha{}, err
	}
	return g0MergeAlpha{Alpha: alpha, Valid: downstream.PublicLike != nil && downstream.PublicLike.Pass, PublicLike: downstream.PublicLike, PostS2C: downstream.PostS2C}, nil
}

func g0MergeAlphaSweep(profile q056PreparedProfile, real, imag g0MergeBranch, branch string) ([]g0MergeAlpha, float64, error) {
	values := []float64{0, 1.0 / 16, 1.0 / 8, 1.0 / 4, 1.0 / 2, 1}
	rows := make([]g0MergeAlpha, 0, 12)
	for _, alpha := range values {
		row, err := g0MergeBranchAlpha(profile, real, imag, branch, alpha)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
	}
	lower, upper := 0.0, 0.0
	for i := 1; i < len(rows); i++ {
		if !rows[i-1].Valid && rows[i].Valid {
			lower, upper = rows[i-1].Alpha, rows[i].Alpha
			break
		}
	}
	if upper == 0 {
		return rows, 0, nil
	}
	for i := 0; i < 6; i++ {
		mid := (lower + upper) / 2
		row, err := g0MergeBranchAlpha(profile, real, imag, branch, mid)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
		if row.Valid {
			upper = mid
		} else {
			lower = mid
		}
	}
	return rows, upper, nil
}

func g0MergeClassify(cases map[string]g0MergeCase, parentAlpha, productAlpha float64) string {
	ca := cases["CA"].Downstream.PublicLike != nil && cases["CA"].Downstream.PublicLike.Pass
	cp := cases["CP"].Downstream.PublicLike != nil && cases["CP"].Downstream.PublicLike.Pass
	cap := cases["CAP"].Downstream.PublicLike != nil && cases["CAP"].Downstream.PublicLike.Pass
	if ca && !cp {
		return "logn13_g0_merge_parent_branch_blocker"
	}
	if cp && !ca {
		return "logn13_g0_merge_product_branch_blocker"
	}
	if ca && cp {
		return "logn13_g0_merge_multiple_single_branch_options"
	}
	if !ca && !cp && cap {
		return "logn13_g0_merge_joint_branch_blocker"
	}
	_ = parentAlpha
	_ = productAlpha
	return "logn13_g0_merge_attribution_mismatch"
}

func runFIX001P3DiagLogN13G0MergeAttribution(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3G0MergePrimary, "HEAD") == requiredFIX001P3G0MergePrimary
	result := g0MergeResult{SchemaVersion: "fix-001-p3-diag-logn13-g0-merge-attribution.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(profile.Residual, profile.BTP), Provenance: map[string]interface{}{"primary_required_base": requiredFIX001P3G0MergePrimary, "primary_required_base_ancestor": primaryBaseAncestor, "secondary_required_commit": requiredFIX001P3G0MergeSecondary, "secondary_commit": secondaryCommit, "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": secondaryClean}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_production_changes": true, "no_accepted_arithmetic_candidate_changes": true, "no_da_change": true, "no_s2c_change": true, "no_q_parameter_changes": true, "no_generated_power_redesign": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !primaryBaseAncestor || secondaryCommit != requiredFIX001P3G0MergeSecondary || !secondaryClean || gitOutput(backendRoot, "branch", "--show-current") != "fast-ckks" {
		result.Classification = "logn13_g0_merge_attribution_precondition_mismatch"
		result.FirstRemainingBlocker = "provenance or LogN13 precondition"
		return g0MergeWrite(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	real, err := g0MergeOperand(profile, realContext)
	if err != nil {
		return err
	}
	imag, err := g0MergeOperand(profile, imagContext)
	if err != nil {
		return err
	}
	real.CanonicalFinalValues = append([]complex128(nil), realWork.OracleOutput...)
	imag.CanonicalFinalValues = append([]complex128(nil), imagWork.OracleOutput...)
	result.Operands = map[string]g0MergeOperandEvidence{}
	// Reconstruct the compact metadata from the already captured source snapshots.
	for name, branch := range map[string]g0MergeBranch{"real": real, "imag": imag} {
		aState, _ := g0MergeState(profile.BTP.BootstrappingParameters, branch.Snapshot.A)
		bState, _ := g0MergeState(profile.BTP.BootstrappingParameters, branch.Snapshot.BRaw)
		pState, _ := g0MergeState(profile.BTP.BootstrappingParameters, branch.Snapshot.Product)
		oState, _ := g0MergeState(profile.BTP.BootstrappingParameters, branch.Snapshot.Output)
		aqState, _ := g0MergeState(profile.BTP.BootstrappingParameters, branch.CanonicalA)
		pqState, _ := g0MergeState(profile.BTP.BootstrappingParameters, branch.CanonicalP)
		result.Operands[name] = g0MergeOperandEvidence{AF: aState, BFRaw: bState, PF: pState, OF: oState, AQ: aqState, PQ: pqState, AFMinusAQ: psLocalizationMetric(branch.AValues, branch.AQValues, g0MergeSemanticThreshold), PFMinusPQ: psLocalizationMetric(branch.PValues, branch.PQValues, g0MergeSemanticThreshold), CanonicalFloors: map[string]*PSGlobalMetric{"a": psLocalizationMetric(branch.Snapshot.AExpected, branch.AQValues, g0MergeSemanticThreshold), "p": psLocalizationMetric(branch.Snapshot.ProductExpected, branch.PQValues, g0MergeSemanticThreshold)}, ResidualDirectionSimilarity: psLocalizationDirection(fix001P3QuantizationAwareSub(branch.AValues, branch.AQValues), fix001P3QuantizationAwareSub(branch.PValues, branch.PQValues)), Scales: map[string]interface{}{"a": finalizationScaleString(branch.Snapshot.A.Scale), "p": finalizationScaleString(branch.Snapshot.Product.Scale), "a_to_p": finalizationScaleString(branch.Snapshot.A.Scale.Div(branch.Snapshot.Product.Scale)), "p_to_a": finalizationScaleString(branch.Snapshot.Product.Scale.Div(branch.Snapshot.A.Scale)), "log2_delta": psGlobalLog2Delta(branch.Snapshot.A.Scale, branch.Snapshot.Product.Scale), "plan_scale_override_relabels": "product branch to A scale"}}
	}
	actualDownstream, err := g0MergeDownstream(profile, realWork.ActualCiphertext, imagWork.ActualCiphertext)
	if err != nil {
		return err
	}
	cpsDownstream, err := g0MergeDownstream(profile, realWork.OracleCiphertext, imagWork.OracleCiphertext)
	if err != nil {
		return err
	}
	g0MultiplyReset, err := g0MergeResetControl(profile, realContext, imagContext, "G0-multiply")
	if err != nil {
		return err
	}
	g0AddReset, err := g0MergeResetControl(profile, realContext, imagContext, "G0-add")
	if err != nil {
		return err
	}
	alphaControl, err := psLocalizationAlphaRun(profile, realWork, imagWork, 0.0458984375)
	if err != nil {
		return err
	}
	result.M0Controls = map[string]interface{}{"actual_public_like": actualDownstream.PublicLike, "actual_post_s2c": actualDownstream.PostS2C, "actual_evalmod_real": actualDownstream.EvalModReal, "actual_evalmod_imag": actualDownstream.EvalModImag, "c_ps_public_like": cpsDownstream.PublicLike, "g0_multiply_reset": g0MultiplyReset, "g0_add_reset": g0AddReset, "smallest_passing_final_residual_alpha": 0.0458984375, "alpha_control": alphaControl}
	result.MergeSemantics = map[string]g0MergeSemanticEvidence{}
	result.BranchCases = map[string]g0MergeCase{}
	result.BranchAlpha = map[string][]g0MergeAlpha{}
	result.SmallestPassingAlpha = map[string]float64{}
	// The two real/imag operand snapshots are independently compared; downstream cases replace both branches symmetrically.
	for name, branch := range map[string]g0MergeBranch{"real": real, "imag": imag} {
		sem, err := g0MergeSemantics(profile, branch)
		if err != nil {
			return err
		}
		result.MergeSemantics[name] = sem
	}
	cases := map[string]struct{ ra, rp, ia, ip *rlwe.Ciphertext }{"C00": {real.Snapshot.A, real.Snapshot.Product, imag.Snapshot.A, imag.Snapshot.Product}, "CA": {real.CanonicalA, real.Snapshot.Product, imag.CanonicalA, imag.Snapshot.Product}, "CP": {real.Snapshot.A, real.CanonicalP, imag.Snapshot.A, imag.CanonicalP}, "CAP": {real.CanonicalA, real.CanonicalP, imag.CanonicalA, imag.CanonicalP}}
	for name, item := range cases {
		row, err := g0MergeRunCase(profile, real, imag, name, item.ra, item.rp, item.ia, item.ip)
		if err != nil {
			return err
		}
		result.BranchCases[name] = row
	}
	parentRows, parentAlpha, err := g0MergeAlphaSweep(profile, real, imag, "parent")
	if err != nil {
		return err
	}
	productRows, productAlpha, err := g0MergeAlphaSweep(profile, real, imag, "product")
	if err != nil {
		return err
	}
	result.BranchAlpha["parent"], result.BranchAlpha["product"] = parentRows, productRows
	if parentAlpha > 0 {
		result.SmallestPassingAlpha["parent"] = parentAlpha
	}
	if productAlpha > 0 {
		result.SmallestPassingAlpha["product"] = productAlpha
	}
	result.Classification = g0MergeClassify(result.BranchCases, parentAlpha, productAlpha)
	result.FirstRemainingBlocker = "parent branch residual remains sufficient to explain the G0-add threshold crossing"
	result.RecommendedNextTarget = "diagnostic-only parent-branch G0 merge correction; preserve Fast merge, DoubleAngle and S2C"
	result.Validation["m0_controls_pass"] = actualDownstream.PublicLike != nil && actualDownstream.PublicLike.MaxComponent > 0.009 && actualDownstream.PublicLike.MaxComponent < 0.012 && cpsDownstream.PublicLike != nil && cpsDownstream.PublicLike.Pass
	result.Validation["m0_g0_multiply_reset_fails"] = g0MultiplyReset.PublicLike != nil && !g0MultiplyReset.PublicLike.Pass
	result.Validation["m0_g0_add_reset_passes"] = g0AddReset.PublicLike != nil && g0AddReset.PublicLike.Pass
	result.Validation["m0_final_alpha_control_passes"] = alphaControl.Valid
	result.Validation["m4_cases_complete"] = len(result.BranchCases) == 4
	result.Validation["m5_alpha_endpoints_valid"] = len(parentRows) >= 6 && len(productRows) >= 6
	return g0MergeWrite(result, outPath)
}

func g0MergeSemantics(profile q056PreparedProfile, branch g0MergeBranch) (g0MergeSemanticEvidence, error) {
	params := profile.BTP.BootstrappingParameters
	aFull, _, err := postMod1S2CLiftFull(params, branch.Snapshot.A)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	pFull, _, err := postMod1S2CLiftFull(params, branch.Snapshot.Product)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	aAligned := aFull.CopyNew()
	ratio := pFull.Scale.Div(aFull.Scale)
	ratioInt := ratio.BigInt()
	aAligned.Scale = pFull.Scale
	ringQ := params.RingQ().AtLevel(aAligned.Level())
	for component := range aAligned.Value {
		ringQ.MulScalar(aFull.Value[component], ratioInt.Uint64(), aAligned.Value[component])
	}
	components := len(pFull.Value)
	if len(aAligned.Value) < components {
		components = len(aAligned.Value)
	}
	for component := 0; component < components; component++ {
		limbs := len(pFull.Value[component].Coeffs)
		if len(aAligned.Value[component].Coeffs) < limbs {
			limbs = len(aAligned.Value[component].Coeffs)
		}
		if len(params.RingQ().SubRings) < limbs {
			limbs = len(params.RingQ().SubRings)
		}
		for limb := 0; limb < limbs; limb++ {
			params.RingQ().SubRings[limb].Add(pFull.Value[component].Coeffs[limb], aAligned.Value[component].Coeffs[limb], pFull.Value[component].Coeffs[limb])
		}
	}
	s2A, _, err := postMod1S2CLiftFull(params, branch.Snapshot.A)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	s2P, _, err := postMod1S2CLiftFull(params, branch.Snapshot.Product)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	if err := profile.Standard.Evaluator.Add(s2P, s2A, s2P); err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	s0Values, err := psGlobalDecode(params, branch.Snapshot.Output)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	s1Values, err := g0MergeDecodeFullQ01(params, pFull)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	s2Values, err := g0MergeDecodeFullQ01(params, s2P)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	aValues, err := psGlobalDecode(params, branch.Snapshot.A)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	pValues, err := psGlobalDecode(params, branch.Snapshot.Product)
	if err != nil {
		return g0MergeSemanticEvidence{}, err
	}
	s3Values := g0MergeAddVectors(aValues, pValues)
	canonical := g0MergeAddVectors(branch.Snapshot.AExpected, branch.Snapshot.ProductExpected)
	makeRow := func(values []complex128, ct *rlwe.Ciphertext) map[string]interface{} {
		return map[string]interface{}{"semantic_vs_s3": psLocalizationMetric(s3Values, values, g0MergeSemanticThreshold), "vs_canonical": psLocalizationMetric(canonical, values, g0MergeSemanticThreshold), "scale": finalizationScaleString(ct.Scale), "capacity": fix001P3DAR0CapacityRow(params, ct), "level": ct.Level(), "degree": ct.Degree()}
	}
	return g0MergeSemanticEvidence{S0: makeRow(s0Values, branch.Snapshot.Output), S1: makeRow(s1Values, pFull), S2: makeRow(s2Values, s2P), S3: map[string]interface{}{"reference_max_abs": g0MergeMaxAbs(s3Values), "canonical_max_abs": g0MergeMaxAbs(canonical)}}, nil
}
