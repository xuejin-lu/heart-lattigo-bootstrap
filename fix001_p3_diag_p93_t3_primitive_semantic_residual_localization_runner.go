//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	fix001P3T3PrimitiveSchema       = "fix-001-p3-diag-p93-t3-primitive-semantic-residual-localization.v1"
	fix001P3T3PrimitiveSecondary    = fix001P3T3CausalProductionSecondary
	fix001P3T3PrimitiveDiffSHA      = fix001P3T3CausalProductionDiffSHA
	fix001P3T3PrimitiveShadowLimit  = 1e-10
	fix001P3T3PrimitiveMaterialFrac = 0.1
)

type fix001P3T3PrimitiveStage struct {
	Checkpoint      string                 `json:"checkpoint"`
	Operation       string                 `json:"operation"`
	ComparisonValid bool                   `json:"comparison_valid"`
	MaxSemanticDiff float64                `json:"max_semantic_difference"`
	Metric          *semanticBisectMetric  `json:"metric,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Capacity        interface{}            `json:"q012_capacity,omitempty"`
	CapacityStatus  string                 `json:"capacity_status,omitempty"`
	RowHashesQ012   interface{}            `json:"row_hashes_q0_q1_q2,omitempty"`
	InvalidReason   string                 `json:"invalid_reason,omitempty"`
}

type fix001P3T3PrimitiveResult struct {
	SchemaVersion    string                 `json:"schema_version"`
	Timestamp        time.Time              `json:"timestamp"`
	Primary          RepositoryMetadata     `json:"primary_repository"`
	Secondary        RepositoryMetadata     `json:"secondary_repository"`
	Environment      EnvironmentMetadata    `json:"environment"`
	Config           BootstrapConfig        `json:"config"`
	FixedProfile     map[string]interface{} `json:"fixed_profile"`
	Provenance       map[string]interface{} `json:"provenance"`
	R0Shadow         map[string]interface{} `json:"r0_source_faithful_t3_shadow"`
	R1Semantic       map[string]interface{} `json:"r1_plaintext_semantic_expectations"`
	R2Checkpoints    map[string]interface{} `json:"r2_stable_checkpoints"`
	R3Historical     map[string]interface{} `json:"r3_historical_clean_validation"`
	R4Material       map[string]interface{} `json:"r4_first_material_boundary"`
	R5Isolation      map[string]interface{} `json:"r5_conditional_primitive_isolation"`
	R6SourceAudit    map[string]interface{} `json:"r6_source_audit"`
	R7Relevance      map[string]interface{} `json:"r7_downstream_relevance"`
	R8Classification string                 `json:"classification"`
	Validation       map[string]interface{} `json:"validation"`
}

func fix001P3T3PrimitiveWrite(result fix001P3T3PrimitiveResult, outPath string) error {
	fields := []struct {
		name  string
		value interface{}
	}{
		{"schema_version", result.SchemaVersion}, {"timestamp", result.Timestamp}, {"primary_repository", result.Primary}, {"secondary_repository", result.Secondary},
		{"environment", result.Environment}, {"config", result.Config}, {"fixed_profile", result.FixedProfile}, {"provenance", result.Provenance},
		{"r0_source_faithful_t3_shadow", result.R0Shadow}, {"r1_plaintext_semantic_expectations", result.R1Semantic}, {"r2_stable_checkpoints", result.R2Checkpoints},
		{"r3_historical_clean_validation", result.R3Historical}, {"r4_first_material_boundary", result.R4Material}, {"r5_conditional_primitive_isolation", result.R5Isolation},
		{"r6_source_audit", result.R6SourceAudit}, {"r7_downstream_relevance", result.R7Relevance}, {"classification", result.R8Classification}, {"validation", result.Validation},
	}
	var builder strings.Builder
	builder.WriteString("{\n")
	for i, field := range fields {
		name, err := json.Marshal(field.name)
		if err != nil {
			return err
		}
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
	if outPath == "" {
		_, err := os.Stdout.Write([]byte(builder.String()))
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(builder.String()), 0o644)
}

func fix001P3T3PrimitiveCipherPowers(profile q056PreparedProfile, branch string) (map[int]*rlwe.Ciphertext, error) {
	input, standardInput := profile.Inputs.FastReal, profile.Inputs.OrdinaryReal
	if branch == "imag" {
		input, standardInput = profile.Inputs.FastImag, profile.Inputs.OrdinaryImag
	}
	fastBranch, _, _, err := fix001P3P93InternalFastTrace(input, standardInput, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.StandardSK, profile.Inputs.FastSK)
	if err != nil {
		return nil, err
	}
	polynomialInput := fastBranch.states["polynomial_input"]
	targetScale, err := fix001P3P93InternalTargetScale(input, profile.Fast.Mod1Parameters, profile.BTP.BootstrappingParameters)
	if err != nil {
		return nil, err
	}
	snapshots, _, err := profile.Fast.Mod1Evaluator.PolynomialEvaluator.DiagnosticGeneratePowers(polynomialInput, profile.Fast.Mod1Parameters.Mod1Poly, targetScale)
	if err != nil {
		return nil, err
	}
	powers := make(map[int]*rlwe.Ciphertext, len(snapshots))
	for _, snapshot := range snapshots {
		powers[snapshot.N] = snapshot.Ciphertext.CopyNew()
	}
	return powers, nil
}

func fix001P3T3PrimitiveMetadata(ct *rlwe.Ciphertext) map[string]interface{} {
	return map[string]interface{}{"level": ct.Level(), "scale": finalizationScaleString(ct.Scale), "degree": ct.Degree(), "is_ntt": ct.IsNTT, "is_montgomery": ct.IsMontgomery}
}

func fix001P3T3PrimitiveRows(ct *rlwe.Ciphertext) map[string]interface{} {
	rows := map[string]interface{}{}
	for component := 0; component <= ct.Degree() && component < len(ct.Value); component++ {
		for limb := 0; limb < 3; limb++ {
			var row []uint64
			if limb < len(ct.Value[component].Coeffs) {
				row = ct.Value[component].Coeffs[limb]
			}
			rows[fmt.Sprintf("c%d_q%d", component, limb)] = finalizationRowHash(row)
		}
	}
	return rows
}

func makeFIX001P3T3PrimitiveStage(params ckks.Parameters, name, operation string, ct *rlwe.Ciphertext, expected []complex128, valid bool, invalidReason string) fix001P3T3PrimitiveStage {
	stage := fix001P3T3PrimitiveStage{Checkpoint: name, Operation: operation, ComparisonValid: false, Metadata: fix001P3T3PrimitiveMetadata(ct), RowHashesQ012: fix001P3T3PrimitiveRows(ct), InvalidReason: invalidReason}
	if !valid {
		stage.CapacityStatus = "not_comparison_valid"
		return stage
	}
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		stage.InvalidReason, stage.CapacityStatus = "decode: "+err.Error(), "unavailable"
		return stage
	}
	metric, err := fix001P3T3CausalMetric(expected, values, fix001P3T3PrimitiveShadowLimit)
	if err != nil {
		stage.InvalidReason, stage.CapacityStatus = "metric: "+err.Error(), "unavailable"
		return stage
	}
	stage.ComparisonValid, stage.MaxSemanticDiff, stage.Metric = true, metric.MaxComponentAbs, metric
	capacity, capErr := postMod1S2CCapacityFromFastCiphertext(params, ct)
	if capErr != nil {
		stage.CapacityStatus = "unavailable: " + capErr.Error()
	} else {
		stage.Capacity, stage.CapacityStatus = capacity, "checked"
	}
	return stage
}

func fix001P3T3PrimitiveCopy(params ckks.Parameters, src *rlwe.Ciphertext, level int) *rlwe.Ciphertext {
	dst := fastckks.NewCiphertext(params, src.Degree(), level)
	*dst.MetaData = *src.MetaData
	dst.IsNTT, dst.IsMontgomery, dst.Scale = src.IsNTT, src.IsMontgomery, src.Scale
	fastckks.Resize(dst, src.Degree(), level, params.N())
	limbs := fastckks.MaintainedLimbCount(&params, level)
	for component := range src.Value {
		for limb := 0; limb < limbs && limb < len(src.Value[component].Coeffs); limb++ {
			copy(dst.Value[component].Coeffs[limb], src.Value[component].Coeffs[limb])
		}
	}
	return dst
}

func fix001P3T3PrimitiveSubAligned(params ckks.Parameters, eval *fastckks.Evaluator, out, sub *rlwe.Ciphertext) error {
	if out.Scale.Equal(sub.Scale) {
		return eval.Sub(out, sub, out)
	}
	if out.Scale.Cmp(sub.Scale) > 0 {
		scratch := fastckks.NewCiphertext(params, sub.Degree(), minInt(out.Level(), sub.Level()))
		*scratch.MetaData = *sub.MetaData
		scratch.IsNTT, scratch.IsMontgomery = sub.IsNTT, sub.IsMontgomery
		if err := eval.MulIntegerMaintained(sub, out.Scale.Div(sub.Scale).BigInt(), scratch); err != nil {
			return err
		}
		scratch.Scale = out.Scale
		return eval.Sub(out, scratch, out)
	}
	scratch := fastckks.NewCiphertext(params, out.Degree(), minInt(out.Level(), sub.Level()))
	*scratch.MetaData = *out.MetaData
	scratch.IsNTT, scratch.IsMontgomery = out.IsNTT, out.IsMontgomery
	if err := eval.MulIntegerMaintained(out, sub.Scale.Div(out.Scale).BigInt(), scratch); err != nil {
		return err
	}
	scratch.Scale = sub.Scale
	if err := eval.Sub(scratch, sub, scratch); err != nil {
		return err
	}
	copyOut := fix001P3P93InternalFastCopy(params, scratch)
	*out = *copyOut
	return nil
}

func fix001P3T3PrimitiveMap(stage fix001P3T3PrimitiveStage) map[string]interface{} {
	data := map[string]interface{}{"checkpoint": stage.Checkpoint, "operation": stage.Operation, "comparison_valid": stage.ComparisonValid, "max_semantic_difference": stage.MaxSemanticDiff, "metadata": stage.Metadata, "row_hashes_q0_q1_q2": stage.RowHashesQ012}
	if stage.Metric != nil {
		data["metric"] = stage.Metric
	}
	if stage.Capacity != nil {
		data["q012_capacity"] = stage.Capacity
	}
	if stage.CapacityStatus != "" {
		data["capacity_status"] = stage.CapacityStatus
	}
	if stage.InvalidReason != "" {
		data["invalid_reason"] = stage.InvalidReason
	}
	return data
}

func fix001P3T3PrimitiveMaxStage(stages []fix001P3T3PrimitiveStage) (string, float64) {
	for _, stage := range stages {
		if stage.ComparisonValid {
			return stage.Checkpoint, stage.MaxSemanticDiff
		}
	}
	return "", 0
}

func runFIX001P3DiagP93T3PrimitiveSemanticResidualLocalization(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit, secondaryBranch := gitOutput(backendRoot, "rev-parse", "HEAD"), gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3T3PrimitiveResult{SchemaVersion: fix001P3T3PrimitiveSchema, Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_q2_policy": "PS-wide Q012", "plan_scale": "2^93", "basis": "Chebyshev", "material_threshold_fraction": fix001P3T3PrimitiveMaterialFrac},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "historical_reference": os.Getenv("FIX001_P93_T3_REFERENCE_JSON"), "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_production_repair": true}}
	result.Validation = map[string]interface{}{"required_secondary_branch": "fast-ckks", "required_secondary_head": fix001P3T3PrimitiveSecondary, "required_secondary_diff_sha256": fix001P3T3PrimitiveDiffSHA, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_parameter_tuning": true}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != fix001P3T3PrimitiveSecondary || diffHash != fix001P3T3PrimitiveDiffSHA || !secondaryDirty {
		result.R8Classification = "A_P93_T3_SHADOW_REPLAY_CONFLICT"
		return fix001P3T3PrimitiveWrite(result, outPath)
	}
	refPath := os.Getenv("FIX001_P93_T3_REFERENCE_JSON")
	if refPath == "" {
		return fmt.Errorf("FIX001_P93_T3_REFERENCE_JSON is required")
	}
	reference, err := fix001P3T3CausalLoadReference(refPath)
	if err != nil {
		return err
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params, eval := profile.BTP.BootstrappingParameters, profile.Fast.Mod1Evaluator.FastCKKS
	result.R0Shadow, result.R1Semantic, result.R2Checkpoints, result.R3Historical, result.R4Material, result.R5Isolation, result.R6SourceAudit, result.R7Relevance = map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}
	result.R6SourceAudit = map[string]interface{}{"source": "lattigo/circuits/ckks/polynomial/fast.go:fastPowerBasis.genPowerInternal", "split_degree_3": []int{1, 2}, "balanced_operations": []string{"maintained_copy", "MulIntegerMaintained", "metadata_scale_update", "Rescale", "MulRelin", "Chebyshev_Add", "subAligned"}, "production_api_surface": "Fast evaluator only", "secondary_unchanged": true}
	result.R7Relevance = map[string]interface{}{"t3_consumer": "T6 generated-power branch", "downstream": "T3 feeds T6 and the even PS path", "final_polynomial_causality_proven": false}
	allShadow := true
	branchRows := map[string]interface{}{}
	firstMaterials := map[string]interface{}{}
	for _, branch := range []string{"real", "imag"} {
		powers, err := fix001P3T3PrimitiveCipherPowers(profile, branch)
		if err != nil {
			return err
		}
		t1, t2, actualT3 := powers[1], powers[2], powers[3]
		if t1 == nil || t2 == nil || actualT3 == nil {
			return fmt.Errorf("%s branch missing T1/T2/T3", branch)
		}
		t1Values, err := psGlobalDecode(params, t1)
		if err != nil {
			return err
		}
		t2Values, err := psGlobalDecode(params, t2)
		if err != nil {
			return err
		}
		expectedProduct := make([]complex128, len(t1Values))
		expectedDouble := make([]complex128, len(t1Values))
		expectedFinal := make([]complex128, len(t1Values))
		for i := range expectedProduct {
			expectedProduct[i] = t2Values[i] * t1Values[i]
			expectedDouble[i] = 2 * expectedProduct[i]
			expectedFinal[i] = expectedDouble[i] - t1Values[i]
		}
		result.R1Semantic[branch] = map[string]interface{}{"Lin": "x2", "Rin": "x1", "Lpost": "x2", "Rpost": "x1", "M": "x2*x1", "D": "2*x2*x1", "final_oracle": "2*x2*x1-x1", "source_values": "decoded exact current production T1/T2"}
		commonLevel := minInt(t1.Level(), t2.Level())
		left := fix001P3T3PrimitiveCopy(params, t2, commonLevel)
		right := fix001P3T3PrimitiveCopy(params, t1, commonLevel)
		stages := []fix001P3T3PrimitiveStage{
			fix001P3T3PrimitiveStageFromExpected(params, "T2-input", "production snapshot", t2, t2Values),
			fix001P3T3PrimitiveStageFromExpected(params, "T1-input", "production snapshot", t1, t1Values),
			fix001P3T3PrimitiveStageFromExpected(params, "left-copy-before-factor", "maintained copy at common level", left, t2Values),
			fix001P3T3PrimitiveStageFromExpected(params, "right-copy-before-factor", "maintained copy at common level", right, t1Values),
		}
		factors := fix001P3T3CausalFactors(params.Q()[commonLevel])
		leftFactor, rightFactor := factors["left"].(uint64), factors["right"].(uint64)
		if err := eval.MulIntegerMaintained(left, new(big.Int).SetUint64(leftFactor), left); err != nil {
			return fmt.Errorf("%s left MulIntegerMaintained: %w", branch, err)
		}
		left.Scale = t2.Scale.Mul(rlwe.NewScale(leftFactor))
		if err := eval.MulIntegerMaintained(right, new(big.Int).SetUint64(rightFactor), right); err != nil {
			return fmt.Errorf("%s right MulIntegerMaintained: %w", branch, err)
		}
		right.Scale = t1.Scale.Mul(rlwe.NewScale(rightFactor))
		stages = append(stages, makeFIX001P3T3PrimitiveStage(params, "left-after-maintained-factor", "MulIntegerMaintained (pre-Rescale; not a stable semantic checkpoint)", left, nil, false, "pre-Rescale maintained-scalar coordinates are excluded by spec"))
		stages = append(stages, makeFIX001P3T3PrimitiveStage(params, "right-after-maintained-factor", "MulIntegerMaintained (pre-Rescale; not a stable semantic checkpoint)", right, nil, false, "pre-Rescale maintained-scalar coordinates are excluded by spec"))
		if err := eval.Rescale(left, left); err != nil {
			return fmt.Errorf("%s left Rescale: %w", branch, err)
		}
		if err := eval.Rescale(right, right); err != nil {
			return fmt.Errorf("%s right Rescale: %w", branch, err)
		}
		stages = append(stages, fix001P3T3PrimitiveStageFromExpected(params, "left-post-Rescale", "Rescale", left, t2Values), fix001P3T3PrimitiveStageFromExpected(params, "right-post-Rescale", "Rescale", right, t1Values))
		var product *rlwe.Ciphertext
		lazy := commonpolynomial.NewPolynomial(profile.Fast.Mod1Parameters.Mod1Poly).Lazy
		product = fastckks.NewCiphertext(params, 1, minInt(left.Level(), right.Level()))
		if lazy {
			fastckks.Resize(product, 2, minInt(left.Level(), right.Level()), params.N())
		}
		*product.MetaData = *left.MetaData
		product.IsNTT, product.IsMontgomery = left.IsNTT, left.IsMontgomery
		if lazy {
			err = eval.Mul(left, right, product)
		} else {
			err = eval.MulRelin(left, right, product)
		}
		if err != nil {
			return fmt.Errorf("%s product (left mont=%v right mont=%v left ntt=%v right ntt=%v): %w", branch, left.IsMontgomery, right.IsMontgomery, left.IsNTT, right.IsNTT, err)
		}
		stages = append(stages, fix001P3T3PrimitiveStageFromExpected(params, "product-before-doubling", "strict MulRelin (commonPoly.Lazy=false)", product, expectedProduct))
		if err := eval.Add(product, product, product); err != nil {
			return fmt.Errorf("%s doubling: %w", branch, err)
		}
		product.Scale = t2.Scale.Mul(t1.Scale).Div(rlwe.NewScale(params.Q()[commonLevel]))
		stages = append(stages, fix001P3T3PrimitiveStageFromExpected(params, "after-doubling", "Chebyshev Add(product, product)", product, expectedDouble))
		sub := fix001P3T3PrimitiveCopy(params, t1, minInt(product.Level(), t1.Level()))
		if err := fix001P3T3PrimitiveSubAligned(params, eval, product, sub); err != nil {
			return fmt.Errorf("%s subAligned: %w", branch, err)
		}
		shadow := product
		stages = append(stages, fix001P3T3PrimitiveStageFromExpected(params, "after-t1-subtraction-alignment", "subAligned(out, T1)", shadow, expectedFinal))
		actualValues, err := psGlobalDecode(params, actualT3)
		if err != nil {
			return err
		}
		shadowValues, err := psGlobalDecode(params, shadow)
		if err != nil {
			return err
		}
		shadowMetric, err := fix001P3T3CausalMetric(actualValues, shadowValues, fix001P3T3PrimitiveShadowLimit)
		if err != nil {
			return err
		}
		metadataMatch := actualT3.Level() == shadow.Level() && actualT3.Degree() == shadow.Degree() && actualT3.Scale.Equal(shadow.Scale)
		stages = append(stages, fix001P3T3PrimitiveStageFromExpected(params, "T3-shadow-final", "source-faithful replay endpoint", shadow, expectedFinal))
		last := &stages[len(stages)-1]
		last.Metadata["production_metadata"] = fix001P3T3PrimitiveMetadata(actualT3)
		last.Metadata["metadata_match"] = metadataMatch
		last.Metadata["shadow_vs_production_max_component_abs"] = shadowMetric.MaxComponentAbs
		if shadowMetric.MaxComponentAbs > fix001P3T3PrimitiveShadowLimit || !metadataMatch {
			allShadow = false
		}
		result.R0Shadow[branch] = map[string]interface{}{"split_degree": []int{1, 2}, "common_level": commonLevel, "factors": factors, "lazy": false, "shadow_vs_production": shadowMetric, "metadata_match": metadataMatch, "shadow_metadata": fix001P3T3PrimitiveMetadata(shadow), "production_metadata": fix001P3T3PrimitiveMetadata(actualT3)}
		rows := make([]interface{}, 0, len(stages))
		for _, stage := range stages {
			rows = append(rows, fix001P3T3PrimitiveMap(stage))
		}
		result.R2Checkpoints[branch] = rows
		actualFormulaMetric, err := fix001P3T3CausalMetric(expectedFinal, actualValues, fix001P3T3PrimitiveShadowLimit)
		if err != nil {
			return err
		}
		first, firstIndex, threshold := "", -1, fix001P3T3PrimitiveMaterialFrac*actualFormulaMetric.MaxComponentAbs
		for index, stage := range stages {
			if stage.ComparisonValid && stage.Checkpoint != "T2-input" && stage.Checkpoint != "T1-input" && stage.MaxSemanticDiff > threshold {
				first, firstIndex = stage.Checkpoint, index
				break
			}
		}
		boundary := map[string]interface{}{}
		if firstIndex >= 0 {
			incomingIndex := firstIndex - 1
			for incomingIndex >= 0 && !stages[incomingIndex].ComparisonValid {
				incomingIndex--
			}
			if incomingIndex >= 0 {
				incoming := stages[incomingIndex]
				outgoing := stages[firstIndex]
				boundary = map[string]interface{}{"incoming_checkpoint": incoming.Checkpoint, "outgoing_checkpoint": outgoing.Checkpoint, "incoming_max_semantic_difference": incoming.MaxSemanticDiff, "outgoing_max_semantic_difference": outgoing.MaxSemanticDiff, "amplification": safeRatio(outgoing.MaxSemanticDiff, incoming.MaxSemanticDiff), "primitive": outgoing.Operation, "incoming_metadata": incoming.Metadata, "outgoing_metadata": outgoing.Metadata, "incoming_capacity": incoming.Capacity, "outgoing_capacity": outgoing.Capacity, "capacity_status": outgoing.CapacityStatus}
			}
		}
		firstMaterials[branch] = map[string]interface{}{"production_t3_implementation_residual": actualFormulaMetric, "material_threshold": threshold, "first_material_checkpoint": first, "boundary_evidence": boundary}
		branchRows[branch] = map[string]interface{}{"first_checkpoint": first, "stage_count": len(stages)}
		refBranch := fix001P3T3CausalRefBranch(reference, branch)
		historical := fix001P3T3CausalPowerMap(refBranch.Powers)
		historicalT3, historicalT2, historicalT1 := fix001P3T3CausalValues(historical[3]), fix001P3T3CausalValues(historical[2]), fix001P3T3CausalValues(historical[1])
		histOracle, err := fix001P3T3CausalMetric(historicalT3, fix001P3T3CausalChebyshevT3(historicalT1, historicalT2), fix001P3T3PrimitiveShadowLimit)
		if err != nil {
			return err
		}
		result.R3Historical[branch] = map[string]interface{}{"endpoint": "historical T3", "oracle_metric": histOracle, "post_rescale_primitive_artifact": "not captured in historical compact artifact; endpoint oracle is the clean validation boundary"}
	}
	result.R4Material = map[string]interface{}{"per_branch": firstMaterials, "order": []string{"left-post-Rescale", "right-post-Rescale", "product-before-doubling", "after-doubling", "after-t1-subtraction-alignment"}}
	result.R5Isolation = map[string]interface{}{"conditional": true, "status": "recorded only for first material boundary", "operand_path": []string{"maintained copy", "MulIntegerMaintained", "scale metadata update", "Rescale"}, "q012_path": true, "q01_vs_q012_shadow": "not run; bounded diagnostic did not alter production"}
	result.R8Classification = "F_MIXED"
	if !allShadow {
		result.R8Classification = "A_P93_T3_SHADOW_REPLAY_CONFLICT"
	} else {
		for _, value := range firstMaterials {
			if m, ok := value.(map[string]interface{}); ok && m["first_material_checkpoint"] == "left-post-Rescale" || ok && m["first_material_checkpoint"] == "right-post-Rescale" {
				result.R8Classification = "B_OPERAND_RESCALE_REGRESSION"
			}
		}
	}
	result.Validation["shadow_endpoint_pass"] = allShadow
	return fix001P3T3PrimitiveWrite(result, outPath)
}

func fix001P3T3PrimitiveStageFromExpected(params ckks.Parameters, name, operation string, ct *rlwe.Ciphertext, expected []complex128) fix001P3T3PrimitiveStage {
	return makeFIX001P3T3PrimitiveStage(params, name, operation, ct, expected, true, "")
}
