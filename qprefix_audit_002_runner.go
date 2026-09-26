package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	ltcommon "github.com/tuneinsight/lattigo/v6/circuits/common/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const qprefixAudit002SummaryPath = "results/QPREFIX-AUDIT-002-summary.md"

type qprefixAudit002Result struct {
	SchemaVersion     string                       `json:"schema_version"`
	Timestamp         time.Time                    `json:"timestamp_utc"`
	Primary           RepositoryMetadata           `json:"primary_repository"`
	Lattigo           RepositoryMetadata           `json:"lattigo_repository"`
	Environment       EnvironmentMetadata          `json:"environment"`
	Config            BootstrapConfig              `json:"config"`
	Parameters        ExperimentParameters         `json:"effective_parameters"`
	Workload          CorrectnessWorkload          `json:"workload"`
	QValues           []string                     `json:"q_values"`
	RestorePlan       []int                        `json:"c2s_restore_plan"`
	MaintainedPrefix  string                       `json:"maintained_prefix"`
	PolicyPrefix      string                       `json:"qprefix_v2_policy_prefix"`
	InitialInputProof qprefixAudit002InitialProof  `json:"group0_input_uniqueness_proof"`
	Checkpoints       []qprefixAudit002Checkpoint  `json:"checkpoints"`
	LinearTransforms  []qprefixAudit002LinearProof `json:"linear_transform_bound_proofs"`
	Rescales          []qprefixAudit002Rescale     `json:"rescale_recurrence"`
	Restores          []qprefixAudit002Restore     `json:"restore_recurrence"`
	HelperEquivalence qprefixAudit002Equivalence   `json:"production_helper_equivalence"`
	Classification    string                       `json:"classification"`
	FirstFailingPoint string                       `json:"first_failing_checkpoint,omitempty"`
	StopReason        string                       `json:"stop_reason,omitempty"`
	Notes             []string                     `json:"notes"`
}

type qprefixAudit002Checkpoint struct {
	Name            string                     `json:"name"`
	Level           int                        `json:"level"`
	Scale           string                     `json:"scale"`
	Degree          int                        `json:"degree"`
	IsNTT           bool                       `json:"is_ntt"`
	IsMontgomery    bool                       `json:"is_montgomery"`
	MaintainedRows  int                        `json:"maintained_rows"`
	PhysicalPrefix  string                     `json:"strongest_physical_prefix"`
	PolicyPrefix    string                     `json:"qprefix_v2_policy_prefix"`
	QValues         []string                   `json:"q_values_used"`
	PhysicalProduct string                     `json:"physical_prefix_product"`
	PolicyProduct   string                     `json:"policy_prefix_product"`
	RowSHA256       string                     `json:"maintained_row_sha256"`
	Components      []qprefixAudit002Component `json:"components"`
}

type qprefixAudit002Component struct {
	Component            int     `json:"component"`
	MaxAbsCoefficient    string  `json:"max_abs_coefficient"`
	TwiceMaxAbs          string  `json:"twice_max_abs"`
	PhysicalPrefixStrict bool    `json:"physical_prefix_strict_capacity"`
	PolicyRho            float64 `json:"policy_rho_2b_over_s_q"`
	PolicyPrefixStrict   bool    `json:"policy_prefix_strict_capacity"`
	ProvenMaxAbs         string  `json:"proven_max_abs_bound"`
	TwiceProvenMaxAbs    string  `json:"twice_proven_max_abs_bound"`
	ProvenPhysicalStrict bool    `json:"proven_bound_physical_prefix_strict"`
	ProvenPolicyStrict   bool    `json:"proven_bound_policy_prefix_strict"`
	ObservedWithinProven bool    `json:"observed_within_proven_bound"`
}

type qprefixAudit002InitialProof struct {
	Source                    string   `json:"source"`
	SourceQ0                  string   `json:"source_q0"`
	ModUpIntegerScalar        string   `json:"modup_integer_scalar"`
	FullSlotTraceGap          int      `json:"full_slot_trace_gap"`
	ScaleDownMaxAbsQ0         []string `json:"scaledown_q0_centered_max_abs"`
	ProvenGroup0InputMaxAbs   []string `json:"proven_group0_input_max_abs"`
	ObservedGroup0InputMaxAbs []string `json:"observed_group0_input_max_abs"`
	StrictQ012                []bool   `json:"strict_q012_capacity"`
	ExactModUpLiftMatch       bool     `json:"exact_modup_lift_match"`
	Pass                      bool     `json:"pass"`
}

type qprefixAudit002LinearProof struct {
	Group                      int      `json:"group"`
	FactorIndex                int      `json:"factor_index"`
	DiagonalCount              int      `json:"diagonal_count"`
	DiagonalCoefficientL1Sum   string   `json:"sum_diagonal_coefficient_l1"`
	MaxDiagonalCoefficientAbs  string   `json:"max_diagonal_coefficient_abs"`
	DiagonalCoefficientsUnique bool     `json:"diagonal_coefficients_strict_q012"`
	InputProvenMaxAbs          []string `json:"input_proven_max_abs"`
	PredictedOutputMaxAbs      []string `json:"predicted_output_max_abs"`
	ObservedOutputMaxAbs       []string `json:"observed_output_max_abs"`
	PredictedOutputStrictQ012  []bool   `json:"predicted_output_strict_q012"`
	ObservedWithinBound        []bool   `json:"observed_within_predicted_bound"`
	Pass                       bool     `json:"pass"`
}

type qprefixAudit002Rescale struct {
	Group                     int      `json:"group"`
	SourceCheckpoint          string   `json:"source_checkpoint"`
	SourceLevel               int      `json:"source_level"`
	TargetLevel               int      `json:"target_level"`
	Divisor                   string   `json:"exact_logical_divisor_q_level"`
	ScaleBefore               string   `json:"scale_before"`
	ScaleAfter                string   `json:"scale_after"`
	ScaleTransitionExact      bool     `json:"scale_transition_exact"`
	PredictedMaxAbs           []string `json:"predicted_post_rescale_max_abs"`
	ObservedMaxAbs            []string `json:"observed_post_rescale_max_abs"`
	PredictedFromProvenBound  []string `json:"predicted_from_proven_source_bound"`
	ObservedWithinProvenBound []bool   `json:"observed_within_proven_bound"`
	PerComponentPass          []bool   `json:"per_component_recurrence_pass"`
	Pass                      bool     `json:"pass"`
}

type qprefixAudit002Restore struct {
	Group                     int      `json:"group"`
	Exponent                  int      `json:"restore_exponent"`
	Factor                    string   `json:"integer_factor"`
	LevelBefore               int      `json:"level_before"`
	LevelAfter                int      `json:"level_after"`
	ScaleBefore               string   `json:"scale_before"`
	ScaleAfter                string   `json:"scale_after"`
	ScaleTransitionExact      bool     `json:"scale_transition_exact"`
	PredictedMaxAbs           []string `json:"predicted_post_restore_max_abs"`
	ObservedMaxAbs            []string `json:"observed_post_restore_max_abs"`
	PredictedFromProvenBound  []string `json:"predicted_from_proven_source_bound"`
	ObservedWithinProvenBound []bool   `json:"observed_within_proven_bound"`
	PerComponentPass          []bool   `json:"per_component_recurrence_pass"`
	Pass                      bool     `json:"pass"`
}

type qprefixAudit002Equivalence struct {
	Attempted               bool   `json:"attempted"`
	ManualFinalLevel        int    `json:"manual_final_level,omitempty"`
	CombinedRealLevel       int    `json:"combined_real_level,omitempty"`
	ManualImagLevel         int    `json:"manual_imag_level,omitempty"`
	CombinedImagLevel       int    `json:"combined_imag_level,omitempty"`
	ManualRealRowsExact     bool   `json:"manual_real_maintained_rows_exact"`
	ManualImagRowsExact     bool   `json:"manual_imag_maintained_rows_exact"`
	ManualRealMetadataExact bool   `json:"manual_real_metadata_exact"`
	ManualImagMetadataExact bool   `json:"manual_imag_metadata_exact"`
	ManualRealRowSHA256     string `json:"manual_real_row_sha256,omitempty"`
	CombinedRealRowSHA256   string `json:"combined_real_row_sha256,omitempty"`
	ManualImagRowSHA256     string `json:"manual_imag_row_sha256,omitempty"`
	CombinedImagRowSHA256   string `json:"combined_imag_row_sha256,omitempty"`
	Pass                    bool   `json:"pass"`
}

// RunQPREFIXAudit002 records the current Fast C2S raw-output, Rescale, and
// restore chain. It intentionally uses only the explicitly maintained Fast
// residue rows and does not execute EvalMod or S2C.
func RunQPREFIXAudit002(cfg BootstrapConfig, primaryRoot, backendRoot string) (qprefixAudit002Result, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return qprefixAudit002Result{}, err
	}
	if err := qprefixAudit002ValidateProfile(cfg, btp); err != nil {
		return qprefixAudit002Result{}, err
	}

	result := qprefixAudit002Result{
		SchemaVersion:    "qprefix-audit-002-c2s-raw-rescale.v1",
		Timestamp:        time.Now().UTC(),
		Primary:          gitMetadata(primaryRoot),
		Lattigo:          gitMetadata(backendRoot),
		Environment:      EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:           cfg,
		Parameters:       parameterMetadata(residual, btp),
		Workload:         CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		RestorePlan:      []int{4, 2, 0, 0},
		MaintainedPrefix: "Q012",
		PolicyPrefix:     "Q0123",
		Classification:   "QPREFIX_C2S_EVIDENCE_INCOMPLETE",
		Notes: []string{
			"Only current fast-qprefix C2S is measured; EvalMod and S2C are outside scope.",
			"Coefficient bounds use centered CRT reconstruction from physically maintained q0/q1/q2 rows; dormant q3+ rows are never read.",
			"The strict Q012 centered-capacity check is stronger than Q0123 policy capacity at the recorded C2S levels.",
		},
	}

	params := btp.BootstrappingParameters
	q := params.Q()
	for _, modulus := range q {
		result.QValues = append(result.QValues, new(big.Int).SetUint64(modulus).String())
	}

	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return qprefixAudit002Result{}, fmt.Errorf("construct current Fast evaluator: %w", err)
	}
	input := reproducibleInput(residual, btp)
	scaled, _, err := fastEval.ScaleDown(input)
	if err != nil {
		return qprefixAudit002Result{}, fmt.Errorf("Fast ScaleDown: %w", err)
	}
	modUp, err := fastEval.ModUp(scaled)
	if err != nil {
		return qprefixAudit002Result{}, fmt.Errorf("Fast ModUp: %w", err)
	}
	initialProof, currentProvenBounds, err := qprefixAudit002InitialModUpProof(scaled, modUp, fastEval, params, btp.LogMaxSlots())
	if err != nil {
		return result, fmt.Errorf("prove group-0 ModUp input lift: %w", err)
	}
	result.InitialInputProof = initialProof
	if !initialProof.Pass {
		result.Classification = "QPREFIX_C2S_EVIDENCE_INCOMPLETE"
		result.FirstFailingPoint = "group_0_input"
		result.StopReason = "the q0-centered ScaleDown to full-slot ModUp source lift could not be proven exactly under Q012"
		return result, nil
	}

	// The ordinary combined path is retained only as an end-to-end control. It
	// receives a byte-identical copy of the same ModUp output as the manual path.
	controlInput := modUp.CopyNew()
	controlReal, controlImag, err := fastEval.CoeffsToSlots(controlInput)
	if err != nil {
		return qprefixAudit002Result{}, fmt.Errorf("combined Fast C2S control: %w", err)
	}
	if !qprefixAudit002SameRows(modUp, controlInput, params) || !qprefixAudit002SameMetadata(modUp, controlInput) {
		result.Classification = "QPREFIX_C2S_MEASUREMENT_CONFLICT"
		result.StopReason = "combined C2S helper modified its input; identical group-0 baseline could not be confirmed"
		result.FirstFailingPoint = "production_helper_input"
		return result, nil
	}
	if !fastEval.C2SCompressionActive || !qprefixAudit002IntsEqual(fastEval.C2SRestorePlan, result.RestorePlan) {
		return qprefixAudit002Result{}, fmt.Errorf("active Fast C2S compression/restore plan differs from frozen profile: active=%t plan=%v", fastEval.C2SCompressionActive, fastEval.C2SRestorePlan)
	}

	state := modUp.CopyNew()
	matrixIdx := 0
groupLoop:
	for group, factors := range fastEval.C2SDFTMatrix.Levels {
		inputName := fmt.Sprintf("group_%d_input", group)
		inputCheckpoint, capacityPass, err := qprefixAudit002Capture(inputName, state, params, currentProvenBounds)
		if err != nil {
			return result, fmt.Errorf("capture %s: %w", inputName, err)
		}
		result.Checkpoints = append(result.Checkpoints, inputCheckpoint)
		if !capacityPass {
			result.Classification = "QPREFIX_C2S_CAPACITY_FAIL"
			result.FirstFailingPoint = inputName
			result.StopReason = "strict centered capacity failed at a C2S group input"
			break groupLoop
		}
		if !qprefixAudit002ObservedWithinProven(inputCheckpoint) {
			result.Classification = "QPREFIX_C2S_MEASUREMENT_CONFLICT"
			result.FirstFailingPoint = inputName
			result.StopReason = "group input observed bound exceeds the carried proven bound"
			break groupLoop
		}
		if !qprefixAudit002ProvenBoundStrict(inputCheckpoint) {
			result.Classification = "QPREFIX_C2S_EVIDENCE_INCOMPLETE"
			result.FirstFailingPoint = inputName
			result.StopReason = "group input proven bound does not establish strict Q012 centered uniqueness"
			break groupLoop
		}

		if factors == 0 {
			return result, fmt.Errorf("C2S group %d has no factor matrices", group)
		}
		var raw qprefixAudit002Checkpoint
		for factorOffset := range factors {
			if matrixIdx >= len(fastEval.C2SDFTMatrix.Matrices) {
				return result, fmt.Errorf("C2S group %d references missing factor matrix %d", group, matrixIdx)
			}
			matrix := ltcommon.LinearTransformation(fastEval.C2SDFTMatrix.Matrices[matrixIdx])
			l1Sum, maxDiagonalCoefficient, diagonalCount, diagonalUnique, err := qprefixAudit002LinearTransformL1(matrix, params)
			if err != nil {
				return result, fmt.Errorf("C2S group %d factor %d diagonal L1 bound: %w", group, matrixIdx, err)
			}
			predictedRawBounds := make([]*big.Int, len(currentProvenBounds))
			for component, inputBound := range currentProvenBounds {
				predictedRawBounds[component] = new(big.Int).Mul(inputBound, l1Sum)
			}
			if err := fastEval.FastCKKS.LinearTransform(state, matrix, state); err != nil {
				return result, fmt.Errorf("C2S group %d factor %d LinearTransform: %w", group, matrixIdx, err)
			}
			factorName := fmt.Sprintf("group_%d_raw_linear_transform", group)
			if factorOffset != factors-1 {
				factorName = fmt.Sprintf("group_%d_factor_%d_post_linear_transform", group, factorOffset)
			}
			factorCheckpoint, factorCapacityPass, err := qprefixAudit002Capture(factorName, state, params, predictedRawBounds)
			if err != nil {
				return result, fmt.Errorf("capture %s: %w", factorName, err)
			}
			result.Checkpoints = append(result.Checkpoints, factorCheckpoint)
			linearProof := qprefixAudit002BuildLinearProof(group, matrixIdx, diagonalCount, l1Sum, maxDiagonalCoefficient, diagonalUnique, currentProvenBounds, predictedRawBounds, factorCheckpoint)
			result.LinearTransforms = append(result.LinearTransforms, linearProof)
			matrixIdx++
			raw = factorCheckpoint
			if !factorCapacityPass {
				result.Classification = "QPREFIX_C2S_CAPACITY_FAIL"
				result.FirstFailingPoint = factorName
				result.StopReason = "strict centered capacity failed at a raw LinearTransform output"
				break groupLoop
			}
			if !linearProof.Pass {
				result.FirstFailingPoint = factorName
				if !qprefixAudit002ObservedWithinProven(factorCheckpoint) {
					result.Classification = "QPREFIX_C2S_MEASUREMENT_CONFLICT"
					result.StopReason = "observed LinearTransform output exceeds its conservative diagonal L1 bound"
				} else if !diagonalUnique || !qprefixAudit002ProvenBoundStrict(factorCheckpoint) {
					result.Classification = "QPREFIX_C2S_EVIDENCE_INCOMPLETE"
					result.StopReason = "the conservative diagonal L1 output bound does not prove strict Q012 centered uniqueness"
				}
				break groupLoop
			}
			currentProvenBounds = predictedRawBounds
		}

		divisor := params.RingQ().SubRings[state.Level()].Modulus
		scaleBeforeRescale := state.Scale
		provenPostRescale := qprefixAudit002RoundedDivideBounds(currentProvenBounds, divisor)
		if err := fastEval.FastCKKS.Rescale(state, state); err != nil {
			return result, fmt.Errorf("C2S group %d Fast Rescale: %w", group, err)
		}
		postName := fmt.Sprintf("group_%d_post_rescale", group)
		post, postCapacityPass, err := qprefixAudit002Capture(postName, state, params, provenPostRescale)
		if err != nil {
			return result, fmt.Errorf("capture %s: %w", postName, err)
		}
		result.Checkpoints = append(result.Checkpoints, post)
		recurrence := qprefixAudit002BuildRescale(group, raw, post, divisor, scaleBeforeRescale, state.Scale, provenPostRescale)
		result.Rescales = append(result.Rescales, recurrence)
		if !recurrence.Pass {
			result.Classification = "QPREFIX_C2S_MEASUREMENT_CONFLICT"
			result.FirstFailingPoint = postName
			result.StopReason = "observed Rescale output violates the exact or proven rounded-division recurrence"
			break groupLoop
		}
		if !postCapacityPass {
			result.Classification = "QPREFIX_C2S_CAPACITY_FAIL"
			result.FirstFailingPoint = postName
			result.StopReason = "strict centered capacity failed after Rescale"
			break groupLoop
		}
		if !qprefixAudit002ProofWithinCapacity(post) {
			result.Classification = "QPREFIX_C2S_EVIDENCE_INCOMPLETE"
			result.FirstFailingPoint = postName
			result.StopReason = "proven post-Rescale bound does not establish strict Q012 centered uniqueness"
			break groupLoop
		}
		currentProvenBounds = provenPostRescale

		restoreExponent := fastEval.C2SRestorePlan[group]
		if restoreExponent != 0 {
			factor := new(big.Int).Lsh(big.NewInt(1), uint(restoreExponent))
			provenRestored := qprefixAudit002ScaleBounds(currentProvenBounds, factor)
			levelBefore := state.Level()
			scaleBefore := finalizationScaleString(state.Scale)
			scaleBeforeValue := state.Scale
			if err := fastEval.FastCKKS.MulIntegerMaintained(state, factor, state); err != nil {
				return result, fmt.Errorf("C2S group %d maintained restore: %w", group, err)
			}
			state.Scale = state.Scale.Mul(rlwe.NewScale(factor))
			restoreName := fmt.Sprintf("group_%d_post_restore", group)
			restored, restoreCapacityPass, err := qprefixAudit002Capture(restoreName, state, params, provenRestored)
			if err != nil {
				return result, fmt.Errorf("capture %s: %w", restoreName, err)
			}
			result.Checkpoints = append(result.Checkpoints, restored)
			restore := qprefixAudit002BuildRestore(group, restoreExponent, levelBefore, scaleBefore, scaleBeforeValue, state, post, restored, provenRestored)
			result.Restores = append(result.Restores, restore)
			if !restore.Pass || state.Level() != levelBefore {
				result.Classification = "QPREFIX_C2S_MEASUREMENT_CONFLICT"
				result.FirstFailingPoint = restoreName
				result.StopReason = "restore recurrence failed or restore consumed a logical level"
				break groupLoop
			}
			if !restoreCapacityPass {
				result.Classification = "QPREFIX_C2S_CAPACITY_FAIL"
				result.FirstFailingPoint = restoreName
				result.StopReason = "strict centered capacity failed after integer restore"
				break groupLoop
			}
			if !qprefixAudit002ProofWithinCapacity(restored) {
				result.Classification = "QPREFIX_C2S_EVIDENCE_INCOMPLETE"
				result.FirstFailingPoint = restoreName
				result.StopReason = "proven post-restore bound does not establish strict Q012 centered uniqueness"
				break groupLoop
			}
			currentProvenBounds = provenRestored
		}
	}
	if matrixIdx != len(fastEval.C2SDFTMatrix.Matrices) && result.StopReason == "" {
		return result, fmt.Errorf("manually stepped %d of %d C2S factors", matrixIdx, len(fastEval.C2SDFTMatrix.Matrices))
	}
	if result.StopReason != "" {
		return result, nil
	}

	manualReal, manualImag, err := qprefixAudit002Split(fastEval.FastCKKS, state, fastEval.C2SDFTMatrix)
	if err != nil {
		return result, fmt.Errorf("manual C2S split: %w", err)
	}
	result.HelperEquivalence = qprefixAudit002CompareHelper(manualReal, manualImag, controlReal, controlImag, params)
	result.HelperEquivalence.Attempted = true
	if !result.HelperEquivalence.Pass {
		result.Classification = "QPREFIX_C2S_MEASUREMENT_CONFLICT"
		result.FirstFailingPoint = "production_helper_equivalence"
		result.StopReason = "manual four-group DFT/split does not match combined production C2S helper"
		return result, nil
	}
	if len(result.Rescales) != 4 || len(result.Checkpoints) < 12 {
		result.Classification = "QPREFIX_C2S_EVIDENCE_INCOMPLETE"
		result.StopReason = fmt.Sprintf("expected four complete C2S groups; got %d Rescale records and %d checkpoints", len(result.Rescales), len(result.Checkpoints))
		return result, nil
	}
	result.Classification = "QPREFIX_C2S_CAPACITY_PROVEN"
	return result, nil
}

func qprefixAudit002InitialModUpProof(scaled, modUp *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator, params ckks.Parameters, logSlots int) (qprefixAudit002InitialProof, []*big.Int, error) {
	q := params.Q()
	if len(q) < 3 {
		return qprefixAudit002InitialProof{}, nil, fmt.Errorf("initial ModUp proof requires q0/q1/q2")
	}
	traceGap := 1 << (params.LogN() - logSlots - 1)
	proof := qprefixAudit002InitialProof{
		Source:              "Level-0 ScaleDown q0-centered coefficients, exact Fast ModUp basis lift, and identity full-slot Trace",
		SourceQ0:            new(big.Int).SetUint64(q[0]).String(),
		FullSlotTraceGap:    traceGap,
		ExactModUpLiftMatch: traceGap == 1,
	}
	scaleMultiplier := (eval.Mod1Parameters.ScalingFactor().Float64() / eval.Mod1Parameters.MessageRatio()) / scaled.Scale.Float64()
	integerScalar := big.NewInt(1)
	if scaleMultiplier > 1 {
		integerScalar.SetUint64(uint64(math.Round(scaleMultiplier)))
	}
	proof.ModUpIntegerScalar = integerScalar.String()
	crt, err := qprefixAudit002NewCRT(q[:3])
	if err != nil {
		return proof, nil, err
	}
	proven := make([]*big.Int, len(scaled.Value))
	for component := range scaled.Value {
		sourceRows, err := qprefixAudit002CoefficientRows(scaled, component, params, 1)
		if err != nil {
			return proof, nil, fmt.Errorf("ScaleDown component %d: %w", component, err)
		}
		outputRows, err := qprefixAudit002CoefficientRows(modUp, component, params, 3)
		if err != nil {
			return proof, nil, fmt.Errorf("ModUp component %d: %w", component, err)
		}
		maxSource := new(big.Int)
		maxOutput := new(big.Int)
		maxObservedOutput := new(big.Int)
		matches := true
		for i, residue := range sourceRows[0] {
			source := new(big.Int).SetUint64(residue)
			if source.Cmp(new(big.Int).Rsh(new(big.Int).SetUint64(q[0]), 1)) >= 0 {
				source.Sub(source, new(big.Int).SetUint64(q[0]))
			}
			sourceAbs := new(big.Int).Abs(new(big.Int).Set(source))
			if sourceAbs.Cmp(maxSource) > 0 {
				maxSource.Set(sourceAbs)
			}
			expected := new(big.Int).Mul(source, integerScalar)
			absExpected := new(big.Int).Abs(new(big.Int).Set(expected))
			if absExpected.Cmp(maxOutput) > 0 {
				maxOutput.Set(absExpected)
			}
			actual := crt.centered(outputRows[0][i], outputRows[1][i], outputRows[2][i])
			actualAbs := new(big.Int).Abs(new(big.Int).Set(actual))
			if actualAbs.Cmp(maxObservedOutput) > 0 {
				maxObservedOutput.Set(actualAbs)
			}
			if actual.Cmp(expected) != 0 {
				matches = false
			}
		}
		strict := new(big.Int).Lsh(new(big.Int).Set(maxOutput), 1).Cmp(crt.modulus) < 0
		proof.ScaleDownMaxAbsQ0 = append(proof.ScaleDownMaxAbsQ0, maxSource.String())
		proof.ProvenGroup0InputMaxAbs = append(proof.ProvenGroup0InputMaxAbs, maxOutput.String())
		proof.ObservedGroup0InputMaxAbs = append(proof.ObservedGroup0InputMaxAbs, maxObservedOutput.String())
		proof.StrictQ012 = append(proof.StrictQ012, strict)
		proof.ExactModUpLiftMatch = proof.ExactModUpLiftMatch && matches
		proven[component] = maxOutput
	}
	proof.Pass = proof.ExactModUpLiftMatch && traceGap == 1
	for _, strict := range proof.StrictQ012 {
		proof.Pass = proof.Pass && strict
	}
	return proof, proven, nil
}

func qprefixAudit002ValidateProfile(cfg BootstrapConfig, btp bootstrapping.Parameters) error {
	if cfg.LogN != 13 || cfg.LogDefaultScale != 45 || cfg.LogSlots != -1 && cfg.LogSlots != 12 {
		return fmt.Errorf("QPREFIX-AUDIT-002 requires the LogN13/P93 profile; got LogN=%d defaultScale=%d logSlots=%d", cfg.LogN, cfg.LogDefaultScale, cfg.LogSlots)
	}
	if len(cfg.Q0) != 1 || cfg.Q0[0] != 56 {
		return fmt.Errorf("QPREFIX-AUDIT-002 requires the accepted residual q0=56 control profile; got %v", cfg.Q0)
	}
	if cfg.SecretHamming != 192 || cfg.Mod1LogScale != 60 || cfg.Mod1Degree != 30 || cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 || cfg.Mod1InvDegree != 0 {
		return fmt.Errorf("QPREFIX-AUDIT-002 requires the accepted P93 EvalMod configuration")
	}
	if !qprefixAudit002IntsEqual(cfg.QSlotsToCoeffs, []int{39, 39, 39}) || !qprefixAudit002IntsEqual(cfg.P, []int{61, 61, 61, 61, 61}) || !qprefixAudit002IntsEqual(cfg.SlotsToCoeffsDFT, []int{1, 1, 1}) || !qprefixAudit002IntsEqual(cfg.CoeffsToSlotsDFT, []int{1, 1, 1, 1}) {
		return fmt.Errorf("QPREFIX-AUDIT-002 requires the accepted P93 S2C/P/C2S factorization")
	}
	if len(cfg.QCoeffsToSlots) != 4 {
		return fmt.Errorf("QPREFIX-AUDIT-002 requires four C2S q0=56 groups; got %v", cfg.QCoeffsToSlots)
	}
	for i, bits := range cfg.QCoeffsToSlots {
		if bits != 56 {
			return fmt.Errorf("QPREFIX-AUDIT-002 requires C2S group %d to use q0=56; got %d", i, bits)
		}
	}
	q := btp.BootstrappingParameters.Q()
	if len(q) < 4 || bitLen(q[0]) != 56 || bitLen(q[1]) != 39 || bitLen(q[2]) != 40 {
		return fmt.Errorf("QPREFIX-AUDIT-002 requires the current q0/q1/q2 profile 56/39/40; got bit lengths %v", modulusBitLengths(q))
	}
	return nil
}

func modulusBitLengths(values []uint64) []int {
	result := make([]int, len(values))
	for i, value := range values {
		result[i] = bitLen(value)
	}
	return result
}

func qprefixAudit002Capture(name string, ct *rlwe.Ciphertext, params ckks.Parameters, provenBounds []*big.Int) (qprefixAudit002Checkpoint, bool, error) {
	if ct == nil || ct.MetaData == nil || ct.Degree() != 1 {
		return qprefixAudit002Checkpoint{}, false, fmt.Errorf("checkpoint requires a non-nil degree-one ciphertext")
	}
	level := ct.Level()
	maintained := fastckks.MaintainedLimbCount(&params, level)
	if maintained < 3 || level < 3 {
		return qprefixAudit002Checkpoint{}, false, fmt.Errorf("checkpoint level %d has only %d maintained residues; Q012 is required", level, maintained)
	}
	if len(provenBounds) != len(ct.Value) {
		return qprefixAudit002Checkpoint{}, false, fmt.Errorf("checkpoint has %d proven component bounds for %d ciphertext components", len(provenBounds), len(ct.Value))
	}
	q := params.Q()
	if len(q) <= level || len(q) < 4 {
		return qprefixAudit002Checkpoint{}, false, fmt.Errorf("checkpoint level %d lacks required q values", level)
	}
	qValues := make([]string, level+1)
	for i := range qValues {
		qValues[i] = new(big.Int).SetUint64(q[i]).String()
	}
	physicalProduct := new(big.Int).Mul(new(big.Int).SetUint64(q[0]), new(big.Int).SetUint64(q[1]))
	physicalProduct.Mul(physicalProduct, new(big.Int).SetUint64(q[2]))
	policyCount := level + 1
	if policyCount > 4 {
		policyCount = 4
	}
	policyProduct := big.NewInt(1)
	for i := 0; i < policyCount; i++ {
		policyProduct.Mul(policyProduct, new(big.Int).SetUint64(q[i]))
	}

	bounds := make([]qprefixAudit002Component, len(ct.Value))
	allStrict := true
	for component := range ct.Value {
		rows, err := qprefixAudit002CoefficientRows(ct, component, params, maintained)
		if err != nil {
			return qprefixAudit002Checkpoint{}, false, err
		}
		maxAbs, err := qprefixAudit002MaxCenteredQ012(rows, q[:3])
		if err != nil {
			return qprefixAudit002Checkpoint{}, false, fmt.Errorf("component %d centered Q012 reconstruction: %w", component, err)
		}
		twice := new(big.Int).Lsh(new(big.Int).Set(maxAbs), 1)
		policyRho := qprefixAudit002Ratio(twice, policyProduct)
		physicalStrict := twice.Cmp(physicalProduct) < 0
		policyStrict := twice.Cmp(policyProduct) < 0
		allStrict = allStrict && physicalStrict && policyStrict
		proven := provenBounds[component]
		if proven == nil || proven.Sign() < 0 {
			return qprefixAudit002Checkpoint{}, false, fmt.Errorf("component %d has an invalid proven bound", component)
		}
		twiceProven := new(big.Int).Lsh(new(big.Int).Set(proven), 1)
		bounds[component] = qprefixAudit002Component{
			Component: component, MaxAbsCoefficient: maxAbs.String(), TwiceMaxAbs: twice.String(), PhysicalPrefixStrict: physicalStrict, PolicyRho: policyRho, PolicyPrefixStrict: policyStrict,
			ProvenMaxAbs: proven.String(), TwiceProvenMaxAbs: twiceProven.String(), ProvenPhysicalStrict: twiceProven.Cmp(physicalProduct) < 0,
			ProvenPolicyStrict: twiceProven.Cmp(policyProduct) < 0, ObservedWithinProven: maxAbs.Cmp(proven) <= 0,
		}
	}
	return qprefixAudit002Checkpoint{
		Name: name, Level: level, Scale: finalizationScaleString(ct.Scale), Degree: ct.Degree(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		MaintainedRows: maintained, PhysicalPrefix: "Q012", PolicyPrefix: fmt.Sprintf("Q0..Q%d", policyCount-1),
		QValues: qValues, PhysicalProduct: physicalProduct.String(), PolicyProduct: policyProduct.String(), RowSHA256: qprefixAudit002RowHash(ct, params, maintained), Components: bounds,
	}, allStrict, nil
}

func qprefixAudit002ProofWithinCapacity(checkpoint qprefixAudit002Checkpoint) bool {
	return qprefixAudit002ObservedWithinProven(checkpoint) && qprefixAudit002ProvenBoundStrict(checkpoint)
}

func qprefixAudit002ObservedWithinProven(checkpoint qprefixAudit002Checkpoint) bool {
	if len(checkpoint.Components) == 0 {
		return false
	}
	for _, component := range checkpoint.Components {
		if !component.ObservedWithinProven {
			return false
		}
	}
	return true
}

func qprefixAudit002ProvenBoundStrict(checkpoint qprefixAudit002Checkpoint) bool {
	if len(checkpoint.Components) == 0 {
		return false
	}
	for _, component := range checkpoint.Components {
		if !component.ProvenPhysicalStrict || !component.ProvenPolicyStrict {
			return false
		}
	}
	return true
}

func qprefixAudit002CoefficientRows(ct *rlwe.Ciphertext, component int, params ckks.Parameters, maintained int) ([][]uint64, error) {
	if component < 0 || component >= len(ct.Value) {
		return nil, fmt.Errorf("invalid component %d", component)
	}
	rows := make([][]uint64, maintained)
	for limb := 0; limb < maintained; limb++ {
		if limb >= len(ct.Value[component].Coeffs) || len(ct.Value[component].Coeffs[limb]) != params.N() {
			return nil, fmt.Errorf("component %d maintained limb q%d is absent", component, limb)
		}
		rows[limb] = append([]uint64(nil), ct.Value[component].Coeffs[limb]...)
		if ct.IsMontgomery {
			params.RingQ().SubRings[limb].IMForm(rows[limb], rows[limb])
		}
		if ct.IsNTT {
			params.RingQ().SubRings[limb].INTT(rows[limb], rows[limb])
		}
	}
	return rows, nil
}

func qprefixAudit002MaxCenteredQ012(rows [][]uint64, q []uint64) (*big.Int, error) {
	if len(rows) < 3 || len(q) < 3 || len(rows[0]) == 0 || len(rows[1]) != len(rows[0]) || len(rows[2]) != len(rows[0]) {
		return nil, fmt.Errorf("Q012 reconstruction requires three equal-length residue rows")
	}
	crt, err := qprefixAudit002NewCRT(q)
	if err != nil {
		return nil, err
	}
	maxAbs := new(big.Int)
	for i := range rows[0] {
		absolute := new(big.Int).Abs(crt.centered(rows[0][i], rows[1][i], rows[2][i]))
		if absolute.Cmp(maxAbs) > 0 {
			maxAbs.Set(absolute)
		}
	}
	return maxAbs, nil
}

type qprefixAudit002CRT struct {
	q0, q1, q2, q01, modulus, inverse01, inverse012, half *big.Int
}

func qprefixAudit002NewCRT(q []uint64) (qprefixAudit002CRT, error) {
	if len(q) < 3 {
		return qprefixAudit002CRT{}, fmt.Errorf("Q012 CRT requires three moduli")
	}
	q0, q1, q2 := new(big.Int).SetUint64(q[0]), new(big.Int).SetUint64(q[1]), new(big.Int).SetUint64(q[2])
	q01 := new(big.Int).Mul(new(big.Int).Set(q0), q1)
	modulus := new(big.Int).Mul(new(big.Int).Set(q01), q2)
	inverse01 := new(big.Int).ModInverse(q0, q1)
	inverse012 := new(big.Int).ModInverse(q01, q2)
	if inverse01 == nil || inverse012 == nil {
		return qprefixAudit002CRT{}, fmt.Errorf("Q012 moduli are not pairwise coprime")
	}
	return qprefixAudit002CRT{q0: q0, q1: q1, q2: q2, q01: q01, modulus: modulus, inverse01: inverse01, inverse012: inverse012, half: new(big.Int).Rsh(new(big.Int).Set(modulus), 1)}, nil
}

func (crt qprefixAudit002CRT) centered(r0, r1, r2 uint64) *big.Int {
	t1 := new(big.Int).Sub(new(big.Int).SetUint64(r1), new(big.Int).SetUint64(r0))
	t1.Mul(t1, crt.inverse01).Mod(t1, crt.q1)
	x01 := new(big.Int).SetUint64(r0)
	x01.Add(x01, new(big.Int).Mul(crt.q0, t1))
	t2 := new(big.Int).Sub(new(big.Int).SetUint64(r2), new(big.Int).Mod(new(big.Int).Set(x01), crt.q2))
	t2.Mul(t2, crt.inverse012).Mod(t2, crt.q2)
	x := new(big.Int).Add(x01, new(big.Int).Mul(crt.q01, t2))
	if x.Cmp(crt.half) > 0 {
		x.Sub(x, crt.modulus)
	}
	return x
}

func qprefixAudit002SumCenteredQ012(rows [][]uint64, q []uint64) (*big.Int, *big.Int, error) {
	if len(rows) < 3 || len(q) < 3 || len(rows[0]) == 0 || len(rows[1]) != len(rows[0]) || len(rows[2]) != len(rows[0]) {
		return nil, nil, fmt.Errorf("Q012 reconstruction requires three equal-length residue rows")
	}
	crt, err := qprefixAudit002NewCRT(q)
	if err != nil {
		return nil, nil, err
	}
	sum, maxAbs := new(big.Int), new(big.Int)
	for i := range rows[0] {
		absolute := new(big.Int).Abs(crt.centered(rows[0][i], rows[1][i], rows[2][i]))
		sum.Add(sum, absolute)
		if absolute.Cmp(maxAbs) > 0 {
			maxAbs.Set(absolute)
		}
	}
	return sum, maxAbs, nil
}

func qprefixAudit002LinearTransformL1(matrix ltcommon.LinearTransformation, params ckks.Parameters) (sum, maxCoefficient *big.Int, diagonalCount int, unique bool, err error) {
	if matrix.MetaData == nil || len(matrix.Vec) == 0 {
		return nil, nil, 0, false, fmt.Errorf("linear transform has no metadata or diagonals")
	}
	q := params.Q()
	crt, err := qprefixAudit002NewCRT(q[:3])
	if err != nil {
		return nil, nil, 0, false, err
	}
	sum, maxCoefficient = new(big.Int), new(big.Int)
	unique = true
	for _, plaintext := range matrix.Vec {
		if len(plaintext.Q.Coeffs) < 3 {
			return nil, nil, 0, false, fmt.Errorf("encoded diagonal has fewer than three Q rows")
		}
		rows := make([][]uint64, 3)
		for limb := range rows {
			if len(plaintext.Q.Coeffs[limb]) != params.N() {
				return nil, nil, 0, false, fmt.Errorf("encoded diagonal q%d row is absent", limb)
			}
			rows[limb] = append([]uint64(nil), plaintext.Q.Coeffs[limb]...)
			if matrix.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(rows[limb], rows[limb])
			}
			if matrix.IsNTT {
				params.RingQ().SubRings[limb].INTT(rows[limb], rows[limb])
			}
		}
		diagonalL1, diagonalMax, err := qprefixAudit002SumCenteredQ012(rows, q[:3])
		if err != nil {
			return nil, nil, 0, false, err
		}
		sum.Add(sum, diagonalL1)
		if diagonalMax.Cmp(maxCoefficient) > 0 {
			maxCoefficient.Set(diagonalMax)
		}
		unique = unique && new(big.Int).Lsh(new(big.Int).Set(diagonalMax), 1).Cmp(crt.modulus) < 0
		diagonalCount++
	}
	return sum, maxCoefficient, diagonalCount, unique, nil
}

func qprefixAudit002RoundedDivideBounds(bounds []*big.Int, divisor uint64) []*big.Int {
	d := new(big.Int).SetUint64(divisor)
	if d.Sign() <= 0 {
		return nil
	}
	addition := new(big.Int).Rsh(new(big.Int).Sub(new(big.Int).Set(d), big.NewInt(1)), 1)
	result := make([]*big.Int, len(bounds))
	for i, bound := range bounds {
		result[i] = new(big.Int).Div(new(big.Int).Add(new(big.Int).Set(bound), addition), d)
	}
	return result
}

func qprefixAudit002ScaleBounds(bounds []*big.Int, factor *big.Int) []*big.Int {
	result := make([]*big.Int, len(bounds))
	for i, bound := range bounds {
		result[i] = new(big.Int).Mul(bound, factor)
	}
	return result
}

func qprefixAudit002BuildLinearProof(group, factorIndex, diagonalCount int, l1Sum, maxDiagonalCoefficient *big.Int, diagonalUnique bool, input, predicted []*big.Int, checkpoint qprefixAudit002Checkpoint) qprefixAudit002LinearProof {
	proof := qprefixAudit002LinearProof{
		Group: group, FactorIndex: factorIndex, DiagonalCount: diagonalCount,
		DiagonalCoefficientL1Sum: l1Sum.String(), MaxDiagonalCoefficientAbs: maxDiagonalCoefficient.String(),
		DiagonalCoefficientsUnique: diagonalUnique, Pass: diagonalUnique && qprefixAudit002ProofWithinCapacity(checkpoint),
	}
	for component := range input {
		proof.InputProvenMaxAbs = append(proof.InputProvenMaxAbs, input[component].String())
		proof.PredictedOutputMaxAbs = append(proof.PredictedOutputMaxAbs, predicted[component].String())
		proof.PredictedOutputStrictQ012 = append(proof.PredictedOutputStrictQ012, checkpoint.Components[component].ProvenPhysicalStrict)
		proof.ObservedOutputMaxAbs = append(proof.ObservedOutputMaxAbs, checkpoint.Components[component].MaxAbsCoefficient)
		proof.ObservedWithinBound = append(proof.ObservedWithinBound, checkpoint.Components[component].ObservedWithinProven)
	}
	return proof
}

func qprefixAudit002Ratio(numerator, denominator *big.Int) float64 {
	if denominator.Sign() == 0 {
		return 0
	}
	n := new(big.Float).SetPrec(256).SetInt(numerator)
	d := new(big.Float).SetPrec(256).SetInt(denominator)
	ratio, _ := new(big.Float).SetPrec(256).Quo(n, d).Float64()
	return ratio
}

func qprefixAudit002RowHash(ct *rlwe.Ciphertext, params ckks.Parameters, maintained int) string {
	hash := sha256.New()
	var word [8]byte
	for component := range ct.Value {
		binary.LittleEndian.PutUint64(word[:], uint64(component))
		_, _ = hash.Write(word[:])
		for limb := 0; limb < maintained; limb++ {
			binary.LittleEndian.PutUint64(word[:], uint64(limb))
			_, _ = hash.Write(word[:])
			for _, residue := range ct.Value[component].Coeffs[limb] {
				binary.LittleEndian.PutUint64(word[:], residue)
				_, _ = hash.Write(word[:])
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func qprefixAudit002BuildRescale(group int, raw, post qprefixAudit002Checkpoint, divisor uint64, scaleBefore, scaleAfter rlwe.Scale, provenPost []*big.Int) qprefixAudit002Rescale {
	expectedScale := scaleBefore.Div(rlwe.NewScale(divisor))
	scaleExact := scaleAfter.Equal(expectedScale)
	record := qprefixAudit002Rescale{Group: group, SourceCheckpoint: raw.Name, SourceLevel: raw.Level, TargetLevel: post.Level, Divisor: new(big.Int).SetUint64(divisor).String(), ScaleBefore: finalizationScaleString(scaleBefore), ScaleAfter: finalizationScaleString(scaleAfter), ScaleTransitionExact: scaleExact, Pass: raw.Level == post.Level+1 && scaleExact}
	divisorBig := new(big.Int).SetUint64(divisor)
	for i, source := range raw.Components {
		bound, _ := new(big.Int).SetString(source.MaxAbsCoefficient, 10)
		observed, _ := new(big.Int).SetString(post.Components[i].MaxAbsCoefficient, 10)
		numerator := new(big.Int).Add(bound, new(big.Int).Rsh(new(big.Int).Sub(new(big.Int).Set(divisorBig), big.NewInt(1)), 1))
		predicted := numerator.Div(numerator, divisorBig)
		proven := provenPost[i]
		withinProven := observed.Cmp(proven) <= 0 && post.Components[i].ProvenMaxAbs == proven.String()
		pass := observed.Cmp(predicted) <= 0 && withinProven
		record.PredictedMaxAbs = append(record.PredictedMaxAbs, predicted.String())
		record.ObservedMaxAbs = append(record.ObservedMaxAbs, observed.String())
		record.PredictedFromProvenBound = append(record.PredictedFromProvenBound, proven.String())
		record.ObservedWithinProvenBound = append(record.ObservedWithinProvenBound, withinProven)
		record.PerComponentPass = append(record.PerComponentPass, pass)
		record.Pass = record.Pass && pass
	}
	return record
}

func qprefixAudit002BuildRestore(group, exponent, levelBefore int, scaleBefore string, scaleBeforeValue rlwe.Scale, state *rlwe.Ciphertext, post, restored qprefixAudit002Checkpoint, provenRestored []*big.Int) qprefixAudit002Restore {
	factor := new(big.Int).Lsh(big.NewInt(1), uint(exponent))
	expectedScale := scaleBeforeValue.Mul(rlwe.NewScale(factor))
	scaleExact := state.Scale.Equal(expectedScale)
	record := qprefixAudit002Restore{Group: group, Exponent: exponent, Factor: factor.String(), LevelBefore: levelBefore, LevelAfter: state.Level(), ScaleBefore: scaleBefore, ScaleAfter: finalizationScaleString(state.Scale), ScaleTransitionExact: scaleExact, Pass: levelBefore == state.Level() && scaleExact}
	for i, source := range post.Components {
		bound, _ := new(big.Int).SetString(source.MaxAbsCoefficient, 10)
		observed, _ := new(big.Int).SetString(restored.Components[i].MaxAbsCoefficient, 10)
		predicted := new(big.Int).Mul(bound, factor)
		proven := provenRestored[i]
		withinProven := observed.Cmp(proven) <= 0 && restored.Components[i].ProvenMaxAbs == proven.String()
		pass := observed.Cmp(predicted) <= 0 && withinProven
		record.PredictedMaxAbs = append(record.PredictedMaxAbs, predicted.String())
		record.ObservedMaxAbs = append(record.ObservedMaxAbs, observed.String())
		record.PredictedFromProvenBound = append(record.PredictedFromProvenBound, proven.String())
		record.ObservedWithinProvenBound = append(record.ObservedWithinProvenBound, withinProven)
		record.PerComponentPass = append(record.PerComponentPass, pass)
		record.Pass = record.Pass && pass
	}
	return record
}

func qprefixAudit002Split(fastEval *fastckks.Evaluator, state *rlwe.Ciphertext, matrix dft.Matrix) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	if matrix.Format == dft.Standard {
		return state.CopyNew(), nil, nil
	}
	if matrix.Format != dft.SplitRealAndImag && matrix.Format != dft.RepackImagAsReal {
		return nil, nil, fmt.Errorf("unsupported C2S matrix format %d", matrix.Format)
	}
	realOut, imagOut := state.CopyNew(), state.CopyNew()
	if err := fastEval.Conjugate(state, realOut); err != nil {
		return nil, nil, err
	}
	if err := fastEval.Sub(state, realOut, imagOut); err != nil {
		return nil, nil, err
	}
	if err := fastEval.Mul(imagOut, complex(0, -1), imagOut); err != nil {
		return nil, nil, err
	}
	if err := fastEval.Add(realOut, state, realOut); err != nil {
		return nil, nil, err
	}
	if matrix.Format == dft.RepackImagAsReal && matrix.LogSlots < fastEval.Parameters.LogMaxSlots() {
		if err := fastEval.Rotate(imagOut, imagOut, 1<<state.LogDimensions.Cols); err != nil {
			return nil, nil, err
		}
		if err := fastEval.Add(realOut, imagOut, realOut); err != nil {
			return nil, nil, err
		}
	}
	return realOut, imagOut, nil
}

func qprefixAudit002CompareHelper(manualReal, manualImag, controlReal, controlImag *rlwe.Ciphertext, params ckks.Parameters) qprefixAudit002Equivalence {
	result := qprefixAudit002Equivalence{Attempted: true}
	if manualReal == nil || controlReal == nil {
		return result
	}
	result.ManualFinalLevel = manualReal.Level()
	result.CombinedRealLevel = controlReal.Level()
	result.ManualRealRowsExact = qprefixAudit002SameRows(manualReal, controlReal, params)
	result.ManualRealMetadataExact = qprefixAudit002SameMetadata(manualReal, controlReal)
	result.ManualRealRowSHA256 = qprefixAudit002RowHash(manualReal, params, fastckks.MaintainedLimbCount(&params, manualReal.Level()))
	result.CombinedRealRowSHA256 = qprefixAudit002RowHash(controlReal, params, fastckks.MaintainedLimbCount(&params, controlReal.Level()))
	if manualImag == nil || controlImag == nil {
		result.ManualImagRowsExact = manualImag == nil && controlImag == nil
		result.ManualImagMetadataExact = result.ManualImagRowsExact
	} else {
		result.ManualImagLevel = manualImag.Level()
		result.CombinedImagLevel = controlImag.Level()
		result.ManualImagRowsExact = qprefixAudit002SameRows(manualImag, controlImag, params)
		result.ManualImagMetadataExact = qprefixAudit002SameMetadata(manualImag, controlImag)
		result.ManualImagRowSHA256 = qprefixAudit002RowHash(manualImag, params, fastckks.MaintainedLimbCount(&params, manualImag.Level()))
		result.CombinedImagRowSHA256 = qprefixAudit002RowHash(controlImag, params, fastckks.MaintainedLimbCount(&params, controlImag.Level()))
	}
	result.Pass = result.ManualRealRowsExact && result.ManualImagRowsExact && result.ManualRealMetadataExact && result.ManualImagMetadataExact && result.ManualFinalLevel == result.CombinedRealLevel
	if controlImag != nil {
		result.Pass = result.Pass && result.ManualImagLevel == result.CombinedImagLevel
	}
	return result
}

func qprefixAudit002SameRows(left, right *rlwe.Ciphertext, params ckks.Parameters) bool {
	if left == nil || right == nil || len(left.Value) != len(right.Value) || left.Level() != right.Level() {
		return false
	}
	maintained := fastckks.MaintainedLimbCount(&params, left.Level())
	if maintained != fastckks.MaintainedLimbCount(&params, right.Level()) {
		return false
	}
	for component := range left.Value {
		for limb := 0; limb < maintained; limb++ {
			if limb >= len(left.Value[component].Coeffs) || limb >= len(right.Value[component].Coeffs) || len(left.Value[component].Coeffs[limb]) != len(right.Value[component].Coeffs[limb]) {
				return false
			}
			for i, value := range left.Value[component].Coeffs[limb] {
				if value != right.Value[component].Coeffs[limb][i] {
					return false
				}
			}
		}
	}
	return true
}

func qprefixAudit002SameMetadata(left, right *rlwe.Ciphertext) bool {
	return left != nil && right != nil && left.Degree() == right.Degree() && left.Level() == right.Level() && left.Scale.Equal(right.Scale) && left.IsNTT == right.IsNTT && left.IsMontgomery == right.IsMontgomery && left.IsBatched == right.IsBatched && left.IsBitReversed == right.IsBitReversed && left.LogDimensions == right.LogDimensions
}

func qprefixAudit002IntsEqual(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func WriteQPREFIXAudit002(result qprefixAudit002Result, rawPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode QPREFIX-AUDIT-002 result: %w", err)
	}
	if rawPath == "" {
		rawPath = "results/QPREFIX-AUDIT-002-C2S-RAW-RESCALE.json"
	}
	if err := os.MkdirAll(filepath.Dir(rawPath), 0o755); err != nil {
		return fmt.Errorf("create QPREFIX-AUDIT-002 raw directory: %w", err)
	}
	if err := os.WriteFile(rawPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write QPREFIX-AUDIT-002 raw result: %w", err)
	}
	summary := qprefixAudit002Summary(result)
	if err := os.MkdirAll(filepath.Dir(qprefixAudit002SummaryPath), 0o755); err != nil {
		return fmt.Errorf("create QPREFIX-AUDIT-002 summary directory: %w", err)
	}
	if err := os.WriteFile(qprefixAudit002SummaryPath, []byte(summary), 0o644); err != nil {
		return fmt.Errorf("write QPREFIX-AUDIT-002 summary: %w", err)
	}
	return nil
}

func qprefixAudit002Summary(result qprefixAudit002Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# QPREFIX-AUDIT-002 C2S raw/Rescale summary\n\n")
	fmt.Fprintf(&b, "- Classification: `%s`\n- Primary: `%s` (%s, dirty=%t)\n- Secondary: `%s` (%s, dirty=%t)\n- Workload: `%s` (%s)\n- q0..q3: `%s`\n- Restore plan: `%v`\n", result.Classification, result.Primary.Commit, result.Primary.Ref, result.Primary.Dirty, result.Lattigo.Commit, result.Lattigo.Ref, result.Lattigo.Dirty, result.Workload.Identifier, result.Workload.Formula, strings.Join(result.QValues[:min(4, len(result.QValues))], ", "), result.RestorePlan)
	if result.FirstFailingPoint != "" {
		fmt.Fprintf(&b, "- First failing point: `%s` — %s\n", result.FirstFailingPoint, result.StopReason)
	}
	fmt.Fprintf(&b, "\n## Initial ModUp lift proof\n\nSource `%s`, q0 `%s`, ModUp integer scalar `%s`, full-slot trace gap `%d`; ScaleDown q0 max `%s`; proven/observed group-0 bounds `%s` / `%s`; exact lift match `%t`; strict Q012 per component `%v`; pass `%t`.\n", result.InitialInputProof.Source, result.InitialInputProof.SourceQ0, result.InitialInputProof.ModUpIntegerScalar, result.InitialInputProof.FullSlotTraceGap, strings.Join(result.InitialInputProof.ScaleDownMaxAbsQ0, "/"), strings.Join(result.InitialInputProof.ProvenGroup0InputMaxAbs, "/"), strings.Join(result.InitialInputProof.ObservedGroup0InputMaxAbs, "/"), result.InitialInputProof.ExactModUpLiftMatch, result.InitialInputProof.StrictQ012, result.InitialInputProof.Pass)
	b.WriteString("\n## Checkpoints\n\n| Checkpoint | Level | Scale | Degree | Rows | B(c0/c1) observed | B(c0/c1) proven | max ρ(policy) |\n|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, checkpoint := range result.Checkpoints {
		b0, b1, rho := "—", "—", float64(0)
		p0, p1 := "—", "—"
		if len(checkpoint.Components) > 0 {
			b0 = checkpoint.Components[0].MaxAbsCoefficient
			p0 = checkpoint.Components[0].ProvenMaxAbs
		}
		if len(checkpoint.Components) > 1 {
			b1 = checkpoint.Components[1].MaxAbsCoefficient
			p1 = checkpoint.Components[1].ProvenMaxAbs
		}
		for _, component := range checkpoint.Components {
			if component.PolicyRho > rho {
				rho = component.PolicyRho
			}
		}
		fmt.Fprintf(&b, "| %s | %d | %s | %d | %d | %s/%s | %s/%s | %.4g |\n", checkpoint.Name, checkpoint.Level, checkpoint.Scale, checkpoint.Degree, checkpoint.MaintainedRows, b0, b1, p0, p1, rho)
	}
	b.WriteString("\n## LinearTransform diagonal L1 proofs\n\n| Group/factor | Diagonals | Σ diagonal coefficient L1 | max diagonal coefficient | input proven B(c0/c1) | output bound B(c0/c1) | observed within bound | strict Q012 | Pass |\n|---|---:|---:|---:|---:|---:|---:|:---:|:---:|\n")
	for _, record := range result.LinearTransforms {
		strict := true
		for _, ok := range record.PredictedOutputStrictQ012 {
			strict = strict && ok
		}
		fmt.Fprintf(&b, "| %d/%d | %d | %s | %s | %s | %s | %t | %t | %t |\n", record.Group, record.FactorIndex, record.DiagonalCount, record.DiagonalCoefficientL1Sum, record.MaxDiagonalCoefficientAbs, strings.Join(record.InputProvenMaxAbs, "/"), strings.Join(record.PredictedOutputMaxAbs, "/"), qprefixAudit002AllTrue(record.ObservedWithinBound), strict, record.Pass)
	}
	b.WriteString("\n## Rescale recurrence\n\n| Group | Source L→target L | Divisor | Scale before→after exact | Predicted B(c0/c1) | Observed B(c0/c1) | Proven B(c0/c1) | Pass |\n|---:|---:|---:|:---:|---:|---:|---:|:---:|\n")
	for _, record := range result.Rescales {
		fmt.Fprintf(&b, "| %d | %d→%d | %s | %s→%s (%t) | %s | %s | %s | %t |\n", record.Group, record.SourceLevel, record.TargetLevel, record.Divisor, record.ScaleBefore, record.ScaleAfter, record.ScaleTransitionExact, strings.Join(record.PredictedMaxAbs, "/"), strings.Join(record.ObservedMaxAbs, "/"), strings.Join(record.PredictedFromProvenBound, "/"), record.Pass)
	}
	b.WriteString("\n## Restore recurrence\n\n| Group | Factor | Level before→after | Scale before→after exact | Predicted B(c0/c1) | Observed B(c0/c1) | Proven B(c0/c1) | Pass |\n|---:|---:|---:|:---:|---:|---:|---:|:---:|\n")
	for _, record := range result.Restores {
		fmt.Fprintf(&b, "| %d | %s | %d→%d | %s→%s (%t) | %s | %s | %s | %t |\n", record.Group, record.Factor, record.LevelBefore, record.LevelAfter, record.ScaleBefore, record.ScaleAfter, record.ScaleTransitionExact, strings.Join(record.PredictedMaxAbs, "/"), strings.Join(record.ObservedMaxAbs, "/"), strings.Join(record.PredictedFromProvenBound, "/"), record.Pass)
	}
	fmt.Fprintf(&b, "\n## Production helper equivalence\n\nAttempted: `%t`; exact maintained rows real/imag: `%t`/`%t`; metadata real/imag: `%t`/`%t`; overall: `%t`.\n", result.HelperEquivalence.Attempted, result.HelperEquivalence.ManualRealRowsExact, result.HelperEquivalence.ManualImagRowsExact, result.HelperEquivalence.ManualRealMetadataExact, result.HelperEquivalence.ManualImagMetadataExact, result.HelperEquivalence.Pass)
	return b.String()
}

func qprefixAudit002AllTrue(values []bool) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}
