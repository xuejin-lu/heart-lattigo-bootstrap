package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type MulAliasZeroCheck struct {
	Component                    int  `json:"component"`
	CoefficientZeroCount         int  `json:"coefficient_domain_zero_count"`
	CoefficientNonzeroCount      int  `json:"coefficient_domain_nonzero_count"`
	NTTMontgomeryZeroCount       int  `json:"ntt_montgomery_zero_count"`
	NTTMontgomeryNonzeroCount    int  `json:"ntt_montgomery_nonzero_count"`
	CoefficientDomainExactlyZero bool `json:"coefficient_domain_exactly_zero"`
	NTTMontgomeryExactlyZero     bool `json:"ntt_montgomery_exactly_zero"`
}

type MulAliasContractEvidence struct {
	MaintainedC1Zero       bool              `json:"maintained_c1_zero"`
	C1                     MulAliasZeroCheck `json:"c1"`
	PrimaryDecodePath      string            `json:"primary_decode_path"`
	FastMulRelinProvenance string            `json:"fast_mul_relin_provenance"`
	FastTruncateProvenance string            `json:"fast_truncate_provenance"`
}

type MulAliasBoundEvidence struct {
	N              int     `json:"n"`
	B              string  `json:"max_abs_c0_coefficient"`
	ProductBound   string  `json:"n_times_b_squared"`
	Q01Half        string  `json:"q01_half"`
	Ratio          float64 `json:"bound_over_q01_half"`
	UniquenessPass bool    `json:"exact_product_uniqueness_bound_pass"`
}

type MulAliasStateEvidence struct {
	State    string                       `json:"state"`
	Level    int                          `json:"level"`
	Degree   int                          `json:"degree"`
	Scale    string                       `json:"scale"`
	Semantic *PSGlobalMetric              `json:"semantic_vs_y0"`
	Rows     []FinalizationRowFingerprint `json:"q0_q1_rows"`
	Contract MulAliasContractEvidence     `json:"zero_secret_contract"`
	Bound    MulAliasBoundEvidence        `json:"c0_product_bound"`
	C0MaxAbs string                       `json:"c0_max_abs_centered"`
}

type MulAliasOperationEvidence struct {
	Operation           string                       `json:"operation"`
	Level               int                          `json:"level"`
	Degree              int                          `json:"degree"`
	Scale               string                       `json:"scale"`
	Rows                []FinalizationRowFingerprint `json:"q0_q1_rows"`
	Semantic            *PSGlobalMetric              `json:"semantic_vs_y0_squared"`
	C1Zero              *MulAliasZeroCheck           `json:"c1_zero"`
	C2Zero              *MulAliasZeroCheck           `json:"c2_zero"`
	PostModularCapacity TargetScaleCapacityEvidence  `json:"post_modular_centered_capacity"`
}

type MulAliasRelinearizationEvidence struct {
	Relinearize        MulAliasOperationEvidence `json:"relinearize"`
	DirectMulRelin     MulAliasOperationEvidence `json:"direct_mul_relin"`
	C0RowsMatchProduct bool                      `json:"relinearize_c0_rows_match_mul_product"`
	RowsMatchDirect    bool                      `json:"relinearize_rows_match_direct_mul_relin"`
	SemanticAgreement  *PSGlobalMetric           `json:"relinearize_vs_direct_semantic"`
}

type MulAliasFreshQ2Witness struct {
	Q2                   uint64 `json:"q2"`
	MismatchCount        int    `json:"mismatch_count"`
	FirstMismatchIndex   int    `json:"first_mismatch_index"`
	FirstCenteredQ01     string `json:"first_centered_q01_candidate"`
	FirstCandidateModQ2  string `json:"first_candidate_mod_q2"`
	FirstFreshQ2Product  string `json:"first_fresh_q2_product_residue"`
	PostModularOnlyLabel string `json:"candidate_label"`
}

type MulAliasPromotionEvidence struct {
	ExpectedMultiplier string                             `json:"expected_multiplier"`
	ScaleL             string                             `json:"low_scale"`
	ScaleH             string                             `json:"high_scale"`
	C0MaxAbsL          string                             `json:"low_c0_max_abs_centered"`
	C0MaxAbsH          string                             `json:"high_c0_max_abs_centered"`
	CoefficientCheck   *TargetScaleCoefficientDomainCheck `json:"coefficient_promotion_check"`
}

type MulAliasResult struct {
	SchemaVersion       string                          `json:"schema_version"`
	Timestamp           time.Time                       `json:"timestamp"`
	Primary             RepositoryMetadata              `json:"primary_repository"`
	Lattigo             RepositoryMetadata              `json:"lattigo_repository"`
	Environment         EnvironmentMetadata             `json:"environment"`
	Config              BootstrapConfig                 `json:"config"`
	Parameters          ExperimentParameters            `json:"effective_parameters"`
	Workload            CorrectnessWorkload             `json:"workload"`
	AuthoritativeOracle string                          `json:"authoritative_poly_oracle"`
	Threshold           float64                         `json:"threshold"`
	CanonicalLow        MulAliasStateEvidence           `json:"state_l_low_compressed"`
	CanonicalHigh       MulAliasStateEvidence           `json:"state_h_high_restored"`
	Promotion           MulAliasPromotionEvidence       `json:"physical_scale_promotion"`
	LowMul              MulAliasOperationEvidence       `json:"low_mul"`
	LowRelinearization  MulAliasRelinearizationEvidence `json:"low_relinearization"`
	HighMul             MulAliasOperationEvidence       `json:"high_mul"`
	HighRelinearization MulAliasRelinearizationEvidence `json:"high_relinearization"`
	FreshQ2Witness      MulAliasFreshQ2Witness          `json:"high_fresh_q2_alias_witness"`
	RootMechanism       string                          `json:"root_mechanism"`
	FirstSupportedCause string                          `json:"first_supported_cause"`
	Validation          map[string]interface{}          `json:"validation"`
}

func mulAliasCountRows(rows [][]uint64) (zero, nonzero int) {
	for _, row := range rows {
		for _, value := range row {
			if value == 0 {
				zero++
			} else {
				nonzero++
			}
		}
	}
	return zero, nonzero
}

func mulAliasZeroCheck(params ckks.Parameters, ct *rlwe.Ciphertext, component int) (MulAliasZeroCheck, error) {
	if ct == nil || component < 0 || component > ct.Degree() {
		return MulAliasZeroCheck{}, fmt.Errorf("invalid zero check component %d", component)
	}
	view, err := targetScaleCoefficientView(params, ct)
	if err != nil {
		return MulAliasZeroCheck{}, err
	}
	coefficientZero, coefficientNonzero := mulAliasCountRows(view.Value[component].Coeffs[:2])
	nttZero, nttNonzero := mulAliasCountRows(ct.Value[component].Coeffs[:2])
	return MulAliasZeroCheck{
		Component: component, CoefficientZeroCount: coefficientZero, CoefficientNonzeroCount: coefficientNonzero,
		NTTMontgomeryZeroCount: nttZero, NTTMontgomeryNonzeroCount: nttNonzero,
		CoefficientDomainExactlyZero: coefficientNonzero == 0, NTTMontgomeryExactlyZero: nttNonzero == 0,
	}, nil
}

func mulAliasMaxAbs(values []*big.Int) *big.Int {
	result := new(big.Int)
	for _, value := range values {
		if new(big.Int).Abs(value).Cmp(result) > 0 {
			result.Set(new(big.Int).Abs(value))
		}
	}
	return result
}

func mulAliasBound(data targetScaleCenteredData, n int) MulAliasBoundEvidence {
	max := mulAliasMaxAbs(data.Values[0])
	product := new(big.Int).Mul(big.NewInt(int64(n)), new(big.Int).Mul(max, max))
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(product), new(big.Float).SetInt(data.Q01Half)).Float64()
	return MulAliasBoundEvidence{N: n, B: max.String(), ProductBound: product.String(), Q01Half: data.Q01Half.String(), Ratio: ratio, UniquenessPass: product.Cmp(data.Q01Half) < 0}
}

func mulAliasContract(params ckks.Parameters, ct *rlwe.Ciphertext) (MulAliasContractEvidence, error) {
	check, err := mulAliasZeroCheck(params, ct, 1)
	if err != nil {
		return MulAliasContractEvidence{}, err
	}
	return MulAliasContractEvidence{
		MaintainedC1Zero:       check.CoefficientDomainExactlyZero && check.NTTMontgomeryExactlyZero,
		C1:                     check,
		PrimaryDecodePath:      "psGlobalDecode -> q01Projection -> diagnosticDecode -> decodeWithSecret(zeroSecret(params))",
		FastMulRelinProvenance: "lattigo/schemes/ckks/fast/evaluator_ntt.go: Evaluator.MulRelin/mulElement",
		FastTruncateProvenance: "lattigo/schemes/ckks/fast/truncate.go: FastTruncateDegree2To1 (zero-secret s=0 contract)",
	}, nil
}

func mulAliasState(params ckks.Parameters, stateName string, ct *rlwe.Ciphertext, expected []complex128) (MulAliasStateEvidence, targetScaleCenteredData, error) {
	decoded, err := psGlobalDecode(params, ct)
	if err != nil {
		return MulAliasStateEvidence{}, targetScaleCenteredData{}, err
	}
	view, err := targetScaleCoefficientView(params, ct)
	if err != nil {
		return MulAliasStateEvidence{}, targetScaleCenteredData{}, err
	}
	data, err := targetScaleCenteredRows(params, view)
	if err != nil {
		return MulAliasStateEvidence{}, targetScaleCenteredData{}, err
	}
	contract, err := mulAliasContract(params, ct)
	if err != nil {
		return MulAliasStateEvidence{}, targetScaleCenteredData{}, err
	}
	max := mulAliasMaxAbs(data.Values[0])
	return MulAliasStateEvidence{
		State: stateName, Level: ct.Level(), Degree: ct.Degree(), Scale: finalizationScaleString(ct.Scale),
		Semantic: psGlobalMetric(expected, decoded), Rows: finalizationEvidence(ct).Rows, Contract: contract,
		Bound: mulAliasBound(data, params.N()), C0MaxAbs: max.String(),
	}, data, nil
}

func mulAliasOperation(params ckks.Parameters, eval *fastckks.Evaluator, operation string, input *rlwe.Ciphertext, expected []complex128) (*rlwe.Ciphertext, MulAliasOperationEvidence, error) {
	output := fastckks.NewCiphertext(params.Parameters, 2, input.Level())
	output.IsNTT, output.IsMontgomery = input.IsNTT, input.IsMontgomery
	if err := eval.Mul(input, input, output); err != nil {
		return nil, MulAliasOperationEvidence{}, err
	}
	decoded, err := psGlobalDecode(params, output)
	if err != nil {
		return nil, MulAliasOperationEvidence{}, err
	}
	capacity, err := doubleAngleCapacity(params, output)
	if err != nil {
		return nil, MulAliasOperationEvidence{}, err
	}
	evidence := MulAliasOperationEvidence{Operation: operation, Level: output.Level(), Degree: output.Degree(), Scale: finalizationScaleString(output.Scale), Rows: finalizationEvidence(output).Rows, Semantic: psGlobalMetric(expected, decoded), PostModularCapacity: capacity}
	for component := 1; component <= 2; component++ {
		check, err := mulAliasZeroCheck(params, output, component)
		if err != nil {
			return nil, MulAliasOperationEvidence{}, err
		}
		if component == 1 {
			evidence.C1Zero = &check
		} else {
			evidence.C2Zero = &check
		}
	}
	return output, evidence, nil
}

func mulAliasRelinearization(params ckks.Parameters, eval *fastckks.Evaluator, label string, product, input *rlwe.Ciphertext, expected []complex128) (MulAliasRelinearizationEvidence, error) {
	relinearized := fastckks.NewCiphertext(params.Parameters, 1, product.Level())
	relinearized.IsNTT, relinearized.IsMontgomery = product.IsNTT, product.IsMontgomery
	if err := eval.Relinearize(product, relinearized); err != nil {
		return MulAliasRelinearizationEvidence{}, err
	}
	direct := fastckks.NewCiphertext(params.Parameters, 1, input.Level())
	direct.IsNTT, direct.IsMontgomery = input.IsNTT, input.IsMontgomery
	if err := eval.MulRelin(input, input, direct); err != nil {
		return MulAliasRelinearizationEvidence{}, err
	}
	relinDecoded, err := psGlobalDecode(params, relinearized)
	if err != nil {
		return MulAliasRelinearizationEvidence{}, err
	}
	directDecoded, err := psGlobalDecode(params, direct)
	if err != nil {
		return MulAliasRelinearizationEvidence{}, err
	}
	relinCapacity, err := doubleAngleCapacity(params, relinearized)
	if err != nil {
		return MulAliasRelinearizationEvidence{}, err
	}
	directCapacity, err := doubleAngleCapacity(params, direct)
	if err != nil {
		return MulAliasRelinearizationEvidence{}, err
	}
	relinEvidence := MulAliasOperationEvidence{Operation: fmt.Sprintf("FastCKKS.Relinearize(%s-MUL)", label), Level: relinearized.Level(), Degree: relinearized.Degree(), Scale: finalizationScaleString(relinearized.Scale), Rows: finalizationEvidence(relinearized).Rows, Semantic: psGlobalMetric(expected, relinDecoded), PostModularCapacity: relinCapacity}
	directEvidence := MulAliasOperationEvidence{Operation: fmt.Sprintf("FastCKKS.MulRelin(%s,%s)", label, label), Level: direct.Level(), Degree: direct.Degree(), Scale: finalizationScaleString(direct.Scale), Rows: finalizationEvidence(direct).Rows, Semantic: psGlobalMetric(expected, directDecoded), PostModularCapacity: directCapacity}
	return MulAliasRelinearizationEvidence{
		Relinearize: relinEvidence, DirectMulRelin: directEvidence,
		C0RowsMatchProduct: finalizationEvidence(product).Rows[0] == relinEvidence.Rows[0],
		RowsMatchDirect:    doubleAngleRowsMatch(relinEvidence.Rows, directEvidence.Rows),
		SemanticAgreement:  psGlobalMetric(relinDecoded, directDecoded),
	}, nil
}

func mulAliasFreshQ2Square(params ckks.Parameters, values []*big.Int) ([]uint64, error) {
	ringQ := params.RingQ()
	if len(ringQ.SubRings) < 3 {
		return nil, fmt.Errorf("fresh q2 witness requires at least q0, q1 and q2")
	}
	subring := ringQ.SubRings[2]
	input := make([]uint64, len(values))
	q2 := new(big.Int).SetUint64(subring.Modulus)
	for i, value := range values {
		input[i] = new(big.Int).Mod(new(big.Int).Set(value), q2).Uint64()
	}
	ntt := make([]uint64, len(values))
	productNTT := make([]uint64, len(values))
	product := make([]uint64, len(values))
	subring.NTT(input, ntt)
	subring.MForm(ntt, ntt)
	subring.MulCoeffsMontgomery(ntt, ntt, productNTT)
	subring.INTT(productNTT, product)
	return product, nil
}

func mulAliasQ2Witness(params ckks.Parameters, highProduct *rlwe.Ciphertext, highValues []*big.Int) (MulAliasFreshQ2Witness, error) {
	view, err := targetScaleCoefficientView(params, highProduct)
	if err != nil {
		return MulAliasFreshQ2Witness{}, err
	}
	data, err := targetScaleCenteredRows(params, view)
	if err != nil {
		return MulAliasFreshQ2Witness{}, err
	}
	q2Product, err := mulAliasFreshQ2Square(params, highValues)
	if err != nil {
		return MulAliasFreshQ2Witness{}, err
	}
	q2 := new(big.Int).SetUint64(params.RingQ().SubRings[2].Modulus)
	witness := MulAliasFreshQ2Witness{Q2: q2.Uint64(), FirstMismatchIndex: -1, PostModularOnlyLabel: "H-MUL c0 centered q0/q1 representative after modular multiplication; not an exact-product claim"}
	for index, candidate := range data.Values[0] {
		candidateMod := new(big.Int).Mod(new(big.Int).Set(candidate), q2)
		fresh := new(big.Int).SetUint64(q2Product[index])
		if candidateMod.Cmp(fresh) != 0 {
			witness.MismatchCount++
			if witness.FirstMismatchIndex < 0 {
				witness.FirstMismatchIndex = index
				witness.FirstCenteredQ01 = candidate.String()
				witness.FirstCandidateModQ2 = candidateMod.String()
				witness.FirstFreshQ2Product = fresh.String()
			}
		}
	}
	return witness, nil
}

func mulAliasWriteResult(result MulAliasResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for DoubleAngle Mul alias diagnostic")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	summary := struct {
		SchemaVersion       string                          `json:"schema_version"`
		Timestamp           time.Time                       `json:"timestamp"`
		Primary             RepositoryMetadata              `json:"primary_repository"`
		Lattigo             RepositoryMetadata              `json:"lattigo_repository"`
		AuthoritativeOracle string                          `json:"authoritative_poly_oracle"`
		Threshold           float64                         `json:"threshold"`
		CanonicalLow        MulAliasStateEvidence           `json:"state_l_low_compressed"`
		CanonicalHigh       MulAliasStateEvidence           `json:"state_h_high_restored"`
		Promotion           MulAliasPromotionEvidence       `json:"physical_scale_promotion"`
		LowMul              MulAliasOperationEvidence       `json:"low_mul"`
		LowRelinearization  MulAliasRelinearizationEvidence `json:"low_relinearization"`
		HighMul             MulAliasOperationEvidence       `json:"high_mul"`
		HighRelinearization MulAliasRelinearizationEvidence `json:"high_relinearization"`
		FreshQ2Witness      MulAliasFreshQ2Witness          `json:"high_fresh_q2_alias_witness"`
		RootMechanism       string                          `json:"root_mechanism"`
		FirstSupportedCause string                          `json:"first_supported_cause"`
		Validation          map[string]interface{}          `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.AuthoritativeOracle, result.Threshold, result.CanonicalLow, result.CanonicalHigh, result.Promotion, result.LowMul, result.LowRelinearization, result.HighMul, result.HighRelinearization, result.FreshQ2Witness, result.RootMechanism, result.FirstSupportedCause, result.Validation}
	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DIAG-DOUBLE-ANGLE-MUL-ALIAS-logN13-summary.json"), summaryData, 0o644)
}

func runFIX001P3DoubleAngleMulAlias(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	y0 := chebyshevRawOracle(state.Poly, state.Z)
	params := state.Params.BootstrappingParameters

	low := psGlobalCopy(params, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(low, low); err != nil {
		return err
	}
	lowEvidence, lowData, err := mulAliasState(params, "L", low, y0)
	if err != nil {
		return err
	}
	lowScaleEvidence, multiplier, matchedScale := targetScaleScaleEvidence(low.Scale, state.Target)
	_ = lowScaleEvidence
	high := psGlobalCopy(params, low)
	if err := state.Eval.FastCKKS.MulIntegerMaintained(high, multiplier, high); err != nil {
		return err
	}
	high.Scale = matchedScale
	high.Scale = state.Target
	highEvidence, highData, err := mulAliasState(params, "H", high, y0)
	if err != nil {
		return err
	}
	promotion := MulAliasPromotionEvidence{ExpectedMultiplier: multiplier.String(), ScaleL: finalizationScaleString(low.Scale), ScaleH: finalizationScaleString(high.Scale), C0MaxAbsL: lowEvidence.C0MaxAbs, C0MaxAbsH: highEvidence.C0MaxAbs, CoefficientCheck: targetScaleCoefficientDomainCheck(lowData, &highData, multiplier)}

	result := MulAliasResult{
		SchemaVersion: "fix-001-p3-diag-double-angle-mul-alias.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(params, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, CanonicalLow: lowEvidence, CanonicalHigh: highEvidence, Promotion: promotion,
		Validation: map[string]interface{}{
			"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "",
			"canonical_root_hash_match": state.Root.RowsMatch([]string{"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf", "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"}),
			"low_high_same_y0":          lowEvidence.Semantic.Pass && highEvidence.Semantic.Pass, "low_scale_evidence": lowScaleEvidence,
			"no_logn16": true, "no_lower_candidate_sweep": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_double_angle_after_first_square": true,
		},
	}
	if !result.Validation["canonical_root_hash_match"].(bool) || !promotion.CoefficientCheck.Pass || !lowEvidence.Semantic.Pass || !highEvidence.Semantic.Pass {
		result.FirstSupportedCause = "double_angle_mul_alias_precondition_mismatch"
		return mulAliasWriteResult(result, outPath)
	}
	if !highEvidence.Contract.MaintainedC1Zero {
		result.FirstSupportedCause = "double_angle_mul_zero_secret_precondition_mismatch"
		return mulAliasWriteResult(result, outPath)
	}

	squareExpected := make([]complex128, len(y0))
	for i, value := range y0 {
		squareExpected[i] = value * value
	}
	lowProduct, lowMul, err := mulAliasOperation(params, state.Eval.FastCKKS, "FastCKKS.Mul(L,L)", low, squareExpected)
	if err != nil {
		return err
	}
	result.LowMul = lowMul
	result.LowRelinearization, err = mulAliasRelinearization(params, state.Eval.FastCKKS, "L", lowProduct, low, squareExpected)
	if err != nil {
		return err
	}
	if lowEvidence.Bound.UniquenessPass && !lowMul.Semantic.Pass {
		result.FirstSupportedCause = "double_angle_mul_primitive_failure_precedes_target_scale_alias_question"
		return mulAliasWriteResult(result, outPath)
	}

	highProduct, highMul, err := mulAliasOperation(params, state.Eval.FastCKKS, "FastCKKS.Mul(H,H)", high, squareExpected)
	if err != nil {
		return err
	}
	result.HighMul = highMul
	result.HighRelinearization, err = mulAliasRelinearization(params, state.Eval.FastCKKS, "H", highProduct, high, squareExpected)
	if err != nil {
		return err
	}
	result.FreshQ2Witness, err = mulAliasQ2Witness(params, highProduct, highData.Values[0])
	if err != nil {
		return err
	}
	result.Validation["h_bound_inconclusive"] = !highEvidence.Bound.UniquenessPass
	result.Validation["h_mul_failed_before_relinearization"] = !highMul.Semantic.Pass
	result.Validation["h_relinearize_equals_direct"] = result.HighRelinearization.RowsMatchDirect && result.HighRelinearization.SemanticAgreement.Pass
	result.Validation["l_no_alias_bound_pass"] = lowEvidence.Bound.UniquenessPass
	result.Validation["l_mul_semantic_pass"] = lowMul.Semantic.Pass
	result.Validation["h_q01_alias_witness"] = result.FreshQ2Witness.MismatchCount > 0

	if !highMul.Semantic.Pass {
		if highEvidence.Bound.UniquenessPass {
			result.FirstSupportedCause = "double_angle_mul_primitive_failure_without_alias"
		} else if result.FreshQ2Witness.MismatchCount > 0 && lowEvidence.Bound.UniquenessPass && lowMul.Semantic.Pass && result.HighRelinearization.RowsMatchDirect && result.HighRelinearization.SemanticAgreement.Pass {
			result.RootMechanism = "physical_target_scale_promotion_exceeds_q01_square_uniqueness_capacity"
			result.FirstSupportedCause = "double_angle_mul_q01_alias_confirmed"
		} else if result.FreshQ2Witness.MismatchCount == 0 {
			result.FirstSupportedCause = "double_angle_mul_capacity_unresolved"
		} else {
			result.FirstSupportedCause = "double_angle_mul_primitive_failure_without_alias"
		}
	} else if !result.HighRelinearization.Relinearize.Semantic.Pass || !result.HighRelinearization.DirectMulRelin.Semantic.Pass || !result.HighRelinearization.RowsMatchDirect {
		result.FirstSupportedCause = "double_angle_relinearization_truncation_failure"
	} else {
		result.FirstSupportedCause = "double_angle_mul_capacity_unresolved"
	}
	return mulAliasWriteResult(result, outPath)
}
