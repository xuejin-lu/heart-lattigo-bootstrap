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
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type TargetScaleRepresentationMetadata struct {
	IsNTT        bool                         `json:"is_ntt"`
	IsMontgomery bool                         `json:"is_montgomery"`
	Level        int                          `json:"level"`
	Degree       int                          `json:"degree"`
	Scale        string                       `json:"scale"`
	Rows         []FinalizationRowFingerprint `json:"q0_q1_rows"`
}

type TargetScaleDomainRoundTrip struct {
	Attempted       bool   `json:"attempted"`
	Pass            bool   `json:"pass"`
	MismatchCount   int    `json:"mismatch_count"`
	FirstMismatch   string `json:"first_mismatch,omitempty"`
	SourceUnchanged bool   `json:"source_unchanged"`
}

func targetScaleRepresentationEqual(a, b TargetScaleRepresentationMetadata) bool {
	return a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && a.Level == b.Level && a.Degree == b.Degree && a.Scale == b.Scale && q01RowsEqual(a.Rows, b.Rows)
}

type TargetScaleCoefficientDomainCheck struct {
	Pass          bool   `json:"pass"`
	MismatchCount int    `json:"mismatch_count"`
	FirstMismatch string `json:"first_mismatch,omitempty"`
}

type TargetScaleHistoricalCapacityView struct {
	P0 TargetScaleCapacityEvidence `json:"p0"`
	P1 TargetScaleCapacityEvidence `json:"p1"`
}

type TargetScaleCapacityDomainResult struct {
	SchemaVersion             string                              `json:"schema_version"`
	Timestamp                 time.Time                           `json:"timestamp"`
	Primary                   RepositoryMetadata                  `json:"primary_repository"`
	Lattigo                   RepositoryMetadata                  `json:"lattigo_repository"`
	Environment               EnvironmentMetadata                 `json:"environment"`
	Config                    BootstrapConfig                     `json:"config"`
	Parameters                ExperimentParameters                `json:"effective_parameters"`
	Workload                  CorrectnessWorkload                 `json:"workload"`
	AuthoritativeOracle       string                              `json:"authoritative_poly_oracle"`
	Threshold                 float64                             `json:"threshold"`
	C0                        TargetScaleRepresentationMetadata   `json:"c0_post_final_representation"`
	HistoricalNTTView         TargetScaleHistoricalCapacityView   `json:"historical_ntt_domain_view"`
	C2CoefficientView         TargetScaleRepresentationMetadata   `json:"c2_coefficient_domain_representation"`
	C3RoundTrip               TargetScaleDomainRoundTrip          `json:"c3_domain_round_trip"`
	C4CorrectedP0             TargetScaleCapacityEvidence         `json:"c4_corrected_p0_capacity"`
	C5CorrectedP1             TargetScaleCapacityEvidence         `json:"c5_corrected_p1_capacity"`
	Scale                     TargetScaleRestorationScaleEvidence `json:"scale_restoration"`
	P2                        TargetScaleRestorationCheckpoint    `json:"p2_integer_promotion"`
	P2CoefficientDomainCheck  *TargetScaleCoefficientDomainCheck  `json:"p2_coefficient_domain_check,omitempty"`
	P3                        TargetScaleRestorationCheckpoint    `json:"p3_matched_scale"`
	P4                        TargetScaleRestorationCheckpoint    `json:"p4_exact_target_metadata"`
	P5Invariants              map[string]bool                     `json:"p5_restored_state_invariants,omitempty"`
	CapacityOracleDisposition string                              `json:"capacity_oracle_disposition"`
	FirstSupportedCause       string                              `json:"first_supported_cause"`
	Validation                map[string]interface{}              `json:"validation"`
}

func targetScaleRepresentation(ct *rlwe.Ciphertext) TargetScaleRepresentationMetadata {
	return TargetScaleRepresentationMetadata{IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, Level: ct.Level(), Degree: ct.Degree(), Scale: finalizationScaleString(ct.Scale), Rows: finalizationEvidence(ct).Rows}
}

func targetScaleCoefficientView(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || source.Level() < 1 || source.Degree() < 1 {
		return nil, fmt.Errorf("coefficient-domain target-scale view requires a degree-one Level >= 1 ciphertext")
	}
	if len(params.RingQ().SubRings) < 2 {
		return nil, fmt.Errorf("coefficient-domain target-scale view requires q0 and q1")
	}
	view := fastckks.NewCiphertext(params, source.Degree(), 1)
	view.MetaData = source.MetaData.CopyNew()
	view.IsNTT = false
	view.IsMontgomery = false
	ringQ := params.RingQ()
	for component := 0; component <= source.Degree(); component++ {
		if len(source.Value[component].Coeffs) < 2 {
			return nil, fmt.Errorf("component %d has no maintained q0/q1 rows", component)
		}
		coeff := ring.NewPoly(ringQ.N(), 1)
		if source.IsNTT {
			if err := fastckks.FastPartialINTT(ringQ, source.Value[component], coeff); err != nil {
				return nil, fmt.Errorf("FastPartialINTT component %d: %w", component, err)
			}
		} else {
			copy(coeff.Coeffs[0], source.Value[component].Coeffs[0])
			copy(coeff.Coeffs[1], source.Value[component].Coeffs[1])
		}
		if source.IsMontgomery {
			for limb := 0; limb < 2; limb++ {
				ringQ.SubRings[limb].IMForm(coeff.Coeffs[limb], coeff.Coeffs[limb])
			}
		}
		copy(view.Value[component].Coeffs[0], coeff.Coeffs[0])
		copy(view.Value[component].Coeffs[1], coeff.Coeffs[1])
	}
	return view, nil
}

func targetScaleDomainRoundTrip(params ckks.Parameters, source, coefficientView *rlwe.Ciphertext) (TargetScaleDomainRoundTrip, error) {
	result := TargetScaleDomainRoundTrip{Attempted: true, Pass: true, SourceUnchanged: true}
	if source == nil || coefficientView == nil {
		return result, fmt.Errorf("domain round-trip requires source and coefficient view")
	}
	sourceBefore := targetScaleRepresentation(source)
	ringQ := params.RingQ()
	reentered := fastckks.NewCiphertext(params, source.Degree(), 1)
	for component := 0; component <= source.Degree(); component++ {
		coeff := ring.NewPoly(ringQ.N(), 1)
		copy(coeff.Coeffs[0], coefficientView.Value[component].Coeffs[0])
		copy(coeff.Coeffs[1], coefficientView.Value[component].Coeffs[1])
		if source.IsMontgomery {
			for limb := 0; limb < 2; limb++ {
				ringQ.SubRings[limb].MForm(coeff.Coeffs[limb], coeff.Coeffs[limb])
			}
		}
		if source.IsNTT {
			if err := fastckks.FastPartialNTT(ringQ, coeff, reentered.Value[component]); err != nil {
				return result, fmt.Errorf("FastPartialNTT component %d: %w", component, err)
			}
		} else {
			copy(reentered.Value[component].Coeffs[0], coeff.Coeffs[0])
			copy(reentered.Value[component].Coeffs[1], coeff.Coeffs[1])
		}
		for limb := 0; limb < 2; limb++ {
			want := source.Value[component].Coeffs[limb]
			got := reentered.Value[component].Coeffs[limb]
			for index := range want {
				if want[index] != got[index] {
					result.Pass = false
					result.MismatchCount++
					if result.FirstMismatch == "" {
						result.FirstMismatch = fmt.Sprintf("component=%d limb=%d index=%d expected=%d actual=%d", component, limb, index, want[index], got[index])
					}
				}
			}
		}
	}
	result.SourceUnchanged = targetScaleRepresentationEqual(sourceBefore, targetScaleRepresentation(source))
	return result, nil
}

func targetScaleCoefficientDomainCheck(data targetScaleCenteredData, promoted *targetScaleCenteredData, multiplier *big.Int) *TargetScaleCoefficientDomainCheck {
	result := &TargetScaleCoefficientDomainCheck{Pass: true}
	for component, values := range data.Values {
		for index, value := range values {
			expected := new(big.Int).Mul(value, multiplier)
			expected.Mod(expected, data.Q01)
			if expected.Cmp(data.Q01Half) >= 0 {
				expected.Sub(expected, data.Q01)
			}
			if expected.Cmp(promoted.Values[component][index]) != 0 {
				result.Pass = false
				result.MismatchCount++
				if result.FirstMismatch == "" {
					result.FirstMismatch = fmt.Sprintf("component=%d index=%d expected=%s actual=%s", component, index, expected.String(), promoted.Values[component][index].String())
				}
			}
		}
	}
	return result
}

type TargetScaleCapacityDomainSummary struct {
	SchemaVersion             string                              `json:"schema_version"`
	Timestamp                 time.Time                           `json:"timestamp"`
	Primary                   RepositoryMetadata                  `json:"primary_repository"`
	Lattigo                   RepositoryMetadata                  `json:"lattigo_repository"`
	AuthoritativeOracle       string                              `json:"authoritative_poly_oracle"`
	Threshold                 float64                             `json:"threshold"`
	C0                        TargetScaleRepresentationMetadata   `json:"c0_post_final_representation"`
	HistoricalNTTView         TargetScaleHistoricalCapacityView   `json:"historical_ntt_domain_view"`
	C2CoefficientView         TargetScaleRepresentationMetadata   `json:"c2_coefficient_domain_representation"`
	C3RoundTrip               TargetScaleDomainRoundTrip          `json:"c3_domain_round_trip"`
	C4CorrectedP0             TargetScaleCapacityEvidence         `json:"c4_corrected_p0_capacity"`
	C5CorrectedP1             TargetScaleCapacityEvidence         `json:"c5_corrected_p1_capacity"`
	Scale                     TargetScaleRestorationScaleEvidence `json:"scale_restoration"`
	P2                        TargetScaleRestorationCheckpoint    `json:"p2_integer_promotion"`
	P2CoefficientDomainCheck  *TargetScaleCoefficientDomainCheck  `json:"p2_coefficient_domain_check,omitempty"`
	P3                        TargetScaleRestorationCheckpoint    `json:"p3_matched_scale"`
	P4                        TargetScaleRestorationCheckpoint    `json:"p4_exact_target_metadata"`
	P5Invariants              map[string]bool                     `json:"p5_restored_state_invariants,omitempty"`
	CapacityOracleDisposition string                              `json:"capacity_oracle_disposition"`
	FirstSupportedCause       string                              `json:"first_supported_cause"`
	Validation                map[string]interface{}              `json:"validation"`
}

func targetScaleCapacityDomainSummary(result TargetScaleCapacityDomainResult) TargetScaleCapacityDomainSummary {
	return TargetScaleCapacityDomainSummary{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.AuthoritativeOracle, result.Threshold, result.C0, result.HistoricalNTTView, result.C2CoefficientView, result.C3RoundTrip, result.C4CorrectedP0, result.C5CorrectedP1, result.Scale, result.P2, result.P2CoefficientDomainCheck, result.P3, result.P4, result.P5Invariants, result.CapacityOracleDisposition, result.FirstSupportedCause, result.Validation}
}

func writeTargetScaleCapacityDomainResult(result TargetScaleCapacityDomainResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for target-scale capacity-domain diagnostic")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	summaryData, err := json.MarshalIndent(targetScaleCapacityDomainSummary(result), "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DIAG-TARGET-SCALE-CAPACITY-DOMAIN-logN13-summary.json"), summaryData, 0o644)
}

func runFIX001P3TargetScaleCapacityDomain(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	o2 := chebyshevRawOracle(state.Poly, state.Z)
	preMetric := psGlobalMetric(o2, state.Root.actual)
	post := psGlobalCopy(state.Params.BootstrappingParameters, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(post, post); err != nil {
		return err
	}
	postDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, post)
	if err != nil {
		return err
	}
	postMetric := psGlobalMetric(o2, postDecoded)
	scaleEvidence, multiplier, matchedScale := targetScaleScaleEvidence(post.Scale, state.Target)
	result := TargetScaleCapacityDomainResult{
		SchemaVersion: "fix-001-p3-diag-target-scale-capacity-domain.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(state.Params.BootstrappingParameters, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096},
		AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", Threshold: correctnessThreshold, Scale: scaleEvidence,
		C0: targetScaleRepresentation(post), Validation: map[string]interface{}{
			"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "",
			"canonical_root_hash_match":         state.Root.RowsMatch([]string{"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf", "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"}),
			"corrected_pre_final_max_component": preMetric.MaxComponent, "corrected_post_final_max_component": postMetric.MaxComponent,
			"corrected_pre_final_pass": preMetric.Pass, "corrected_post_final_pass": postMetric.Pass,
			"historical_capacity_disposition_only": "invalid_ntt_domain_capacity_view", "historical_helper_consumed_ntt_domain": post.IsNTT,
			"no_logn16": true, "no_lower_scale_sweep": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_double_angle": true, "no_exp003": true,
		},
	}
	if !result.Validation["canonical_root_hash_match"].(bool) || !preMetric.Pass || !postMetric.Pass || post.Scale.Log2() < 30 || state.Target.Log2() < 58 || multiplier.Cmp(new(big.Int).Lsh(big.NewInt(1), 29)) != 0 || scaleEvidence.Log2Delta < 32 {
		result.CapacityOracleDisposition = "capacity_domain_precondition_mismatch"
		result.FirstSupportedCause = "TARGET_SCALE_CAPACITY_DOMAIN_PRECONDITION_MISMATCH"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}
	if !post.IsNTT {
		result.CapacityOracleDisposition = "capacity_domain_precondition_mismatch"
		result.FirstSupportedCause = "TARGET_SCALE_CAPACITY_DOMAIN_ASSUMPTION_MISMATCH"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}

	historicalProjection, _, err := q01Projection(state.Params.BootstrappingParameters, post, true)
	if err != nil {
		return err
	}
	historicalData, err := targetScaleCenteredRows(state.Params.BootstrappingParameters, historicalProjection)
	if err != nil {
		return err
	}
	result.HistoricalNTTView = TargetScaleHistoricalCapacityView{P0: targetScaleCapacity(historicalData, big.NewInt(1)), P1: targetScaleCapacity(historicalData, multiplier)}

	coefficientView, err := targetScaleCoefficientView(state.Params.BootstrappingParameters, post)
	if err != nil {
		return err
	}
	result.C2CoefficientView = targetScaleRepresentation(coefficientView)
	roundTrip, err := targetScaleDomainRoundTrip(state.Params.BootstrappingParameters, post, coefficientView)
	if err != nil {
		return err
	}
	result.C3RoundTrip = roundTrip
	coefficientData, err := targetScaleCenteredRows(state.Params.BootstrappingParameters, coefficientView)
	if err != nil {
		return err
	}
	result.C4CorrectedP0 = targetScaleCapacity(coefficientData, big.NewInt(1))
	if !roundTrip.Pass {
		result.CapacityOracleDisposition = "capacity_domain_transform_unresolved"
		result.FirstSupportedCause = "target_scale_capacity_domain_transform_failure"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}
	if !result.C4CorrectedP0.Pass {
		result.CapacityOracleDisposition = "capacity_domain_transform_unresolved"
		result.FirstSupportedCause = "target_scale_post_rescale_centered_capacity_failure"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}

	result.C5CorrectedP1 = targetScaleCapacity(coefficientData, multiplier)
	if !result.C5CorrectedP1.Pass {
		result.CapacityOracleDisposition = "historical_ntt_domain_capacity_invalid_corrected_domain_also_fails"
		result.FirstSupportedCause = "target_scale_restoration_centered_capacity_failure_confirmed"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}

	promoted := psGlobalCopy(state.Params.BootstrappingParameters, post)
	if err := state.Eval.FastCKKS.MulIntegerMaintained(promoted, multiplier, promoted); err != nil {
		return err
	}
	promotedDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	multiplierFloat, _ := new(big.Float).SetInt(multiplier).Float64()
	p2Scaled := targetScaleMetric(targetScaleScaled(o2, multiplierFloat), promotedDecoded, correctnessThreshold*maxFloat64(1, multiplierFloat))
	p2Normalized := targetScaleMetric(o2, targetScaleScaled(promotedDecoded, 1/multiplierFloat), correctnessThreshold)
	modularCheck, err := targetScaleRowsModularCheck(state.Params.BootstrappingParameters, post, promoted, multiplier)
	if err != nil {
		return err
	}
	promotedCoefficientView, err := targetScaleCoefficientView(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	promotedCoefficientData, err := targetScaleCenteredRows(state.Params.BootstrappingParameters, promotedCoefficientView)
	if err != nil {
		return err
	}
	coefficientCheck := targetScaleCoefficientDomainCheck(coefficientData, &promotedCoefficientData, multiplier)
	p2 := targetScaleCheckpoint("P2_integer_promotion_at_original_scale", promoted)
	p2.ScaledOracle, p2.NormalizedOracle, p2.IndependentModularCheck = p2Scaled, p2Normalized, modularCheck
	result.P2, result.P2CoefficientDomainCheck = p2, coefficientCheck
	if !p2Scaled.Pass || !p2Normalized.Pass || !modularCheck.Pass || !coefficientCheck.Pass {
		result.CapacityOracleDisposition = "historical_ntt_domain_capacity_invalid_corrected_domain_passed"
		result.FirstSupportedCause = "target_scale_restoration_integer_promotion_failure"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}

	promoted.Scale = matchedScale
	p3Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	p3 := targetScaleCheckpoint("P3_integer_promotion_with_matched_scale", promoted)
	p3.CorrectedOracle = psGlobalMetric(o2, p3Decoded)
	p3.PerturbationFromP0 = psGlobalMetric(postDecoded, p3Decoded)
	result.P3 = p3
	if !p3.CorrectedOracle.Pass {
		result.CapacityOracleDisposition = "historical_ntt_domain_capacity_invalid_corrected_domain_passed"
		result.FirstSupportedCause = "target_scale_restoration_matched_scale_failure"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}

	promoted.Scale = state.Target
	p4Decoded, err := psGlobalDecode(state.Params.BootstrappingParameters, promoted)
	if err != nil {
		return err
	}
	p4 := targetScaleCheckpoint("P4_exact_public_target_scale_metadata", promoted)
	p4.CorrectedOracle = psGlobalMetric(o2, p4Decoded)
	p4.PerturbationFromP3 = psGlobalMetric(p3Decoded, p4Decoded)
	result.P4 = p4
	if !p4.CorrectedOracle.Pass {
		result.CapacityOracleDisposition = "historical_ntt_domain_capacity_invalid_corrected_domain_passed"
		result.FirstSupportedCause = "target_scale_restoration_metadata_normalization_failure"
		return writeTargetScaleCapacityDomainResult(result, outPath)
	}

	result.P5Invariants = map[string]bool{
		"scale_equals_exact_public_target": promoted.Scale.Equal(state.Target),
		"level_unchanged_from_p0":          promoted.Level() == post.Level(),
		"degree_unchanged_from_p0":         promoted.Degree() == post.Degree(),
		"q01_hashes_unchanged_p2_to_p3":    q01RowsEqual(result.P2.Rows, result.P3.Rows),
		"q01_hashes_unchanged_p3_to_p4":    q01RowsEqual(result.P3.Rows, result.P4.Rows),
		"corrected_capacity_valid":         result.C5CorrectedP1.Pass,
		"corrected_semantic_error_pass":    p4.CorrectedOracle.Pass,
		"historical_o1_not_used":           true,
	}
	result.CapacityOracleDisposition = "historical_ntt_domain_capacity_invalid_corrected_domain_passed"
	result.FirstSupportedCause = "compressed_ps_2p91_target_scale_restoration_validated_after_capacity_domain_fix"
	for _, pass := range result.P5Invariants {
		if !pass {
			result.FirstSupportedCause = "target_scale_restoration_metadata_normalization_failure"
			break
		}
	}
	return writeTargetScaleCapacityDomainResult(result, outPath)
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
