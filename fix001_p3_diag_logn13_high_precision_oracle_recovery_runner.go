//go:build !lattigo_standard

package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3HighPrecisionOraclePrimary   = "591c8fa5dabada8ac3e2577f5fc512a2fab85e3c"
	requiredFIX001P3HighPrecisionOracleSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	fix001P3HighPrecisionOracleInjectionLimit    = 1e-10
	fix001P3HighPrecisionOracleReportingLimit    = 1e-12
)

var fix001P3HighPrecisionOraclePrecisions = []uint{80, 96, 128, 160}

type fix001P3HighPrecisionOracleResult struct {
	SchemaVersion         string                            `json:"schema_version"`
	Timestamp             time.Time                         `json:"timestamp"`
	Primary               RepositoryMetadata                `json:"primary_repository"`
	Lattigo               RepositoryMetadata                `json:"lattigo_repository"`
	Environment           EnvironmentMetadata               `json:"environment"`
	Config                BootstrapConfig                   `json:"config"`
	Parameters            ExperimentParameters              `json:"effective_parameters"`
	Provenance            map[string]interface{}            `json:"provenance"`
	H0                    map[string]interface{}            `json:"h0_control"`
	PrecisionSweep        []map[string]interface{}          `json:"h2_precision_sweep"`
	SelectedPrecision     uint                              `json:"selected_precision_bits,omitempty"`
	SourceAndReporting    map[string]interface{}            `json:"source_and_reporting_floors,omitempty"`
	Causal                map[string]interface{}            `json:"four_state_evalmod_causal_decomposition,omitempty"`
	Checkpoints           []map[string]interface{}          `json:"da_checkpoint_attribution,omitempty"`
	Counterfactuals       map[string]map[string]interface{} `json:"downstream_counterfactuals,omitempty"`
	Projected             map[string]interface{}            `json:"s2c_projected_components,omitempty"`
	Classification        string                            `json:"classification"`
	FirstRemainingBlocker string                            `json:"first_remaining_blocker"`
	Validation            map[string]interface{}            `json:"validation"`
}

func fix001P3HighPrecisionOracleWrite(result fix001P3HighPrecisionOracleResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func fix001P3HighPrecisionOracleDecodeFull(params ckks.Parameters, ct *rlwe.Ciphertext, precision uint) ([]*bignum.Complex, error) {
	plaintext := rlwe.NewDecryptor(params, zeroSecret(params)).DecryptNew(ct)
	values := make([]*bignum.Complex, params.MaxSlots())
	for i := range values {
		values[i] = bignum.NewComplex().SetPrec(precision)
	}
	if err := ckks.NewEncoder(params, precision).Decode(plaintext, values); err != nil {
		return nil, err
	}
	return values, nil
}

func fix001P3HighPrecisionOracleDecodeProjected(params ckks.Parameters, ct *rlwe.Ciphertext, precision uint) ([]*bignum.Complex, error) {
	projected, _, err := semanticBisectProject(ct, params)
	if err != nil {
		return nil, err
	}
	plaintext := rlwe.NewDecryptor(params, zeroSecret(params)).DecryptNew(projected)
	values := make([]*bignum.Complex, params.MaxSlots())
	for i := range values {
		values[i] = bignum.NewComplex().SetPrec(precision)
	}
	if err := ckks.NewEncoder(params, precision).Decode(plaintext, values); err != nil {
		return nil, err
	}
	return values, nil
}

func fix001P3HighPrecisionPreprocessedInput(params ckks.Parameters, fastEval *bootstrapping.FastEvaluator, source *rlwe.Ciphertext, precision uint) ([]*bignum.Complex, error) {
	current := source.CopyNew()
	current.Scale = fastEval.Mod1Parameters.ScalingFactor()
	if err := fastEval.FastCKKS.Add(current, evalModCausalOffset(fastEval), current); err != nil {
		return nil, err
	}
	return fix001P3HighPrecisionOracleDecodeProjected(params, current, precision)
}

func fix001P3HighPrecisionChebyshevOracle(poly bignum.Polynomial, input []*bignum.Complex, precision uint) []*bignum.Complex {
	output := make([]*bignum.Complex, len(input))
	mul := bignum.NewComplexMultiplier()
	two := bignum.ToComplex(2, precision)
	for i, value := range input {
		x := value.Clone()
		previous := bignum.ToComplex(1, precision)
		term := x.Clone()
		result := bignum.NewComplex().SetPrec(precision)
		product := bignum.NewComplex().SetPrec(precision)
		twoX := bignum.NewComplex().SetPrec(precision)
		mul.Mul(two, x, twoX)
		if len(poly.Coeffs) > 0 && poly.Coeffs[0] != nil {
			mul.Mul(previous, poly.Coeffs[0], product)
			result.Add(result, product)
		}
		if len(poly.Coeffs) > 1 && poly.Coeffs[1] != nil {
			mul.Mul(term, poly.Coeffs[1], product)
			result.Add(result, product)
		}
		for k := 2; k < len(poly.Coeffs); k++ {
			next := bignum.NewComplex().SetPrec(precision)
			mul.Mul(twoX, term, next)
			next.Sub(next, previous)
			previous, term = term, next
			if poly.Coeffs[k] != nil {
				mul.Mul(term, poly.Coeffs[k], product)
				result.Add(result, product)
			}
		}
		output[i] = result
	}
	return output
}

func fix001P3HighPrecisionOracleCipher(params ckks.Parameters, source *rlwe.Ciphertext, values []*bignum.Complex) (*rlwe.Ciphertext, error) {
	plaintext := ckks.NewPlaintext(params, source.Level())
	*plaintext.MetaData = *source.MetaData
	// CKKS encoder input is coefficient-domain non-Montgomery; the source Fast
	// ciphertext may use a Montgomery execution representation. Preserve the
	// logical DA metadata while materializing the oracle in the decoder's domain.
	plaintext.IsMontgomery = false
	if err := ckks.NewEncoder(params, values[0][0].Prec()).Encode(values, plaintext); err != nil {
		return nil, err
	}
	ciphertext := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*ciphertext.MetaData = *plaintext.MetaData
	for component := range plaintext.Value.Coeffs {
		copy(ciphertext.Value[0].Coeffs[component], plaintext.Value.Coeffs[component])
	}
	return ciphertext, nil
}

func fix001P3HighPrecisionOracleHash(ct *rlwe.Ciphertext) string {
	h := sha256.New()
	var buf [8]byte
	for _, component := range ct.Value {
		for _, limb := range component.Coeffs {
			for _, value := range limb {
				binary.LittleEndian.PutUint64(buf[:], value)
				_, _ = h.Write(buf[:])
			}
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func fix001P3HighPrecisionMetric(expected, actual []*bignum.Complex, threshold float64) *PSGlobalMetric {
	metric := &PSGlobalMetric{WorstIndex: -1, Threshold: threshold}
	if len(expected) != len(actual) || len(expected) == 0 {
		return metric
	}
	for i := range expected {
		dr := new(big.Float).Sub(actual[i][0], expected[i][0])
		di := new(big.Float).Sub(actual[i][1], expected[i][1])
		ar, _ := new(big.Float).Abs(dr).Float64()
		ai, _ := new(big.Float).Abs(di).Float64()
		complexAbs := math.Hypot(ar, ai)
		metric.MeanComplex += complexAbs
		if complexAbs > metric.MaxComplex {
			metric.MaxComplex = complexAbs
		}
		if ar > metric.MaxComponent {
			metric.MaxComponent, metric.WorstIndex, metric.WorstPart = ar, i, "real"
		}
		if ai > metric.MaxComponent {
			metric.MaxComponent, metric.WorstIndex, metric.WorstPart = ai, i, "imag"
		}
	}
	metric.MeanComplex /= float64(len(expected))
	metric.Pass = metric.MaxComponent <= threshold
	return metric
}

func fix001P3HighPrecisionToComplex(values []*bignum.Complex) []complex128 {
	output := make([]complex128, len(values))
	for i, value := range values {
		output[i] = value.Complex128()
	}
	return output
}

func fix001P3HighPrecisionMetadataMatch(source, oracle *rlwe.Ciphertext) map[string]interface{} {
	logicalMatch := source.Level() == oracle.Level() && source.Scale.Equal(oracle.Scale) && source.Degree() == oracle.Degree() && source.N() == oracle.N() && source.LogDimensions == oracle.LogDimensions && source.IsNTT == oracle.IsNTT
	return map[string]interface{}{
		"level":                        source.Level() == oracle.Level(),
		"scale":                        source.Scale.Equal(oracle.Scale),
		"degree":                       source.Degree() == oracle.Degree(),
		"n":                            source.N() == oracle.N(),
		"log_dimensions":               source.LogDimensions == oracle.LogDimensions,
		"is_ntt":                       source.IsNTT == oracle.IsNTT,
		"logical_metadata_match_pass":  logicalMatch,
		"source_is_montgomery":         source.IsMontgomery,
		"oracle_is_montgomery":         oracle.IsMontgomery,
		"montgomery_domain_normalized": source.IsMontgomery != oracle.IsMontgomery,
	}
}

func fix001P3HighPrecisionMetadataPass(metadata map[string]interface{}) bool {
	matched, ok := metadata["logical_metadata_match_pass"].(bool)
	return ok && matched
}

func fix001P3HighPrecisionReportingFloor(values []*bignum.Complex, precision uint) *PSGlobalMetric {
	reencoded := make([]*bignum.Complex, len(values))
	for i, value := range values {
		reencoded[i] = bignum.ToComplex(value.Complex128(), precision)
	}
	return fix001P3HighPrecisionMetric(values, reencoded, fix001P3HighPrecisionOracleReportingLimit)
}

func fix001P3HighPrecisionSweepBranch(params ckks.Parameters, fastEval *bootstrapping.FastEvaluator, input, source *rlwe.Ciphertext, poly bignum.Polynomial, precision uint) (map[string]interface{}, error) {
	inputValues, err := fix001P3HighPrecisionPreprocessedInput(params, fastEval, input, precision)
	if err != nil {
		return nil, err
	}
	sourceOracle := fix001P3HighPrecisionChebyshevOracle(poly, inputValues, precision)
	oracle, err := fix001P3HighPrecisionOracleCipher(params, source, sourceOracle)
	if err != nil {
		return nil, err
	}
	repeat, err := fix001P3HighPrecisionOracleCipher(params, source, sourceOracle)
	if err != nil {
		return nil, err
	}
	decoded, err := fix001P3HighPrecisionOracleDecodeFull(params, oracle, precision)
	if err != nil {
		return nil, err
	}
	pF, err := fix001P3EvalModResidualDecode(params, source)
	if err != nil {
		return nil, err
	}
	source128 := fix001P3HighPrecisionToComplex(sourceOracle)
	decoded128 := fix001P3HighPrecisionToComplex(decoded)
	pFPOMetric := fix001P3EvalModResidualMetric(source128, pF, fix001P3EvalModResidualStageThreshold)
	injection := fix001P3HighPrecisionMetric(sourceOracle, decoded, fix001P3HighPrecisionOracleInjectionLimit)
	reporting := fix001P3HighPrecisionReportingFloor(decoded, precision)
	metadata := fix001P3HighPrecisionMetadataMatch(source, oracle)
	metadataPass := fix001P3HighPrecisionMetadataPass(metadata)
	deterministic := fix001P3HighPrecisionOracleHash(oracle) == fix001P3HighPrecisionOracleHash(repeat)
	return map[string]interface{}{
		"source_precision_bits":         precision,
		"encoder_precision_bits":        precision,
		"oracle_injection":              injection,
		"reporting_conversion_floor":    reporting,
		"p_f_minus_p_o":                 pFPOMetric,
		"metadata_match":                metadata,
		"metadata_match_pass":           metadataPass,
		"deterministic_materialization": deterministic,
		"qualified_h2":                  injection.Pass && injection.MaxComponent <= 0.1*pFPOMetric.MaxComponent && metadataPass && deterministic,
		"decoded_reporting_values":      len(decoded128),
	}, nil
}

func runFIX001P3DiagLogN13HighPrecisionOracleRecovery(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	primaryBaseAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3HighPrecisionOraclePrimary, "HEAD") == requiredFIX001P3HighPrecisionOraclePrimary
	result := fix001P3HighPrecisionOracleResult{
		SchemaVersion: "fix-001-p3-diag-logn13-high-precision-oracle-recovery.v1",
		Timestamp:     time.Now().UTC(),
		Primary:       gitMetadata(primaryRoot),
		Lattigo:       gitMetadata(backendRoot),
		Environment:   EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:        cfg,
		Provenance: map[string]interface{}{
			"primary_required_base":          requiredFIX001P3HighPrecisionOraclePrimary,
			"primary_required_base_ancestor": primaryBaseAncestor,
			"secondary_required_commit":      requiredFIX001P3HighPrecisionOracleSecondary,
			"secondary_commit":               secondaryCommit,
			"secondary_branch":               gitOutput(backendRoot, "branch", "--show-current"),
			"secondary_clean":                secondaryClean,
		},
		Validation: map[string]interface{}{
			"no_secondary_production_changes": true, "no_candidate_retuning": true, "no_q_parameter_changes": true,
			"no_g0_f0_da_behavior_changes": true, "no_s2c_changes": true, "no_generated_power_redesign": true,
			"no_production_integration": true, "no_logn16": true, "no_benchmark": true,
			"no_gate_4_or_5": true, "no_exp003": true, "logn13_only": cfg.LogN == 13,
		},
	}
	if cfg.LogN != 13 || !primaryBaseAncestor || !secondaryClean || secondaryCommit != requiredFIX001P3HighPrecisionOracleSecondary || gitOutput(backendRoot, "branch", "--show-current") != "fast-ckks" {
		result.Classification = "logn13_high_precision_oracle_control_provenance_mismatch"
		result.FirstRemainingBlocker = "startup provenance or Secondary precondition"
		return fix001P3HighPrecisionOracleWrite(result, outPath)
	}

	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	acceptedRow, candidate, err := fix001P3G0RunCandidate(profile, 2)
	if err != nil {
		return err
	}
	if candidate.realFinal == nil || candidate.imagFinal == nil {
		result.Classification = "logn13_high_precision_oracle_control_provenance_mismatch"
		result.FirstRemainingBlocker = "accepted G0=2/F0=3 candidate did not produce both PS outputs"
		return fix001P3HighPrecisionOracleWrite(result, outPath)
	}
	realPath, imagPath, err := fix001P3S2CAcceptedPaths(profile, candidate)
	if err != nil {
		return err
	}
	realFinal, err := evalModMatchedFinalEvidence(realPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	imagFinal, err := evalModMatchedFinalEvidence(imagPath, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK)
	if err != nil {
		return err
	}
	publicEvidence, err := fix001P3K3RunPublicLike(realPath, imagPath, profile.Fast, profile.Standard, profile.BTP, profile.Residual.MaxSlots(), profile.StandardSK, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	metadataPass, _ := publicEvidence.Metadata["pass"].(bool)
	contracts := acceptedRow.ValidLocalQ2 && realPath.FirstFailure == "none" && imagPath.FirstFailure == "none" && realFinal.RestoreRowsMatch && imagFinal.RestoreRowsMatch && realFinal.NormalizedBeforeRestoreCapacity != nil && realFinal.NormalizedBeforeRestoreCapacity.Pass && imagFinal.NormalizedBeforeRestoreCapacity != nil && imagFinal.NormalizedBeforeRestoreCapacity.Pass && realFinal.FinalRestoreCapacity != nil && realFinal.FinalRestoreCapacity.Pass && imagFinal.FinalRestoreCapacity != nil && imagFinal.FinalRestoreCapacity.Pass && metadataPass && publicEvidence.InputUnchanged
	result.H0 = map[string]interface{}{
		"executed_candidate": map[string]interface{}{
			"name": "G0-local-q2-2-F0-local-q2-3", "g0_guard_bits": 2, "f0_guard_bits": 3, "all_da_rounds_local_q2": true,
			"g0_row_valid_local_q2": acceptedRow.ValidLocalQ2, "real_first_failure": realPath.FirstFailure, "imag_first_failure": imagPath.FirstFailure,
		},
		"evalmod_real_vs_standard":                 realFinal.FastVsStandard,
		"evalmod_imag_vs_standard":                 imagFinal.FastVsStandard,
		"post_s2c_vs_standard":                     publicEvidence.PostS2C,
		"public_like_vs_standard":                  publicEvidence.PublicLike,
		"final_metadata":                           publicEvidence.Metadata,
		"exact_capacity_row_and_restore_contracts": contracts,
		"input_unchanged":                          publicEvidence.InputUnchanged,
	}
	result.Validation["h0_executed_schedule_verified"] = contracts
	if !contracts {
		result.Classification = "logn13_high_precision_oracle_control_provenance_mismatch"
		result.FirstRemainingBlocker = "accepted executed path did not satisfy corrected G0=2/F0=3 capacity/row/restore contracts"
		return fix001P3HighPrecisionOracleWrite(result, outPath)
	}

	params := profile.BTP.BootstrappingParameters
	poly := profile.Fast.Mod1Parameters.Mod1Poly
	result.PrecisionSweep = make([]map[string]interface{}, 0, len(fix001P3HighPrecisionOraclePrecisions))
	selected := uint(0)
	for _, precision := range fix001P3HighPrecisionOraclePrecisions {
		realSweep, err := fix001P3HighPrecisionSweepBranch(params, profile.Fast, profile.Inputs.FastReal, candidate.realFinal, poly, precision)
		if err != nil {
			return err
		}
		imagSweep, err := fix001P3HighPrecisionSweepBranch(params, profile.Fast, profile.Inputs.FastImag, candidate.imagFinal, poly, precision)
		if err != nil {
			return err
		}
		qualified := realSweep["qualified_h2"] == true && imagSweep["qualified_h2"] == true
		result.PrecisionSweep = append(result.PrecisionSweep, map[string]interface{}{"precision_bits": precision, "real": realSweep, "imag": imagSweep, "qualified_h2_both_branches": qualified})
		if qualified && selected == 0 {
			selected = precision
		}
	}
	result.SelectedPrecision = selected
	if selected == 0 {
		result.Classification = "logn13_high_precision_oracle_materialization_still_insufficient"
		result.FirstRemainingBlocker = "no tested precision (80/96/128/160 bits) satisfies both-branch injection, margin, metadata, and determinism gates"
		result.Validation["h2_qualified"] = false
		return fix001P3HighPrecisionOracleWrite(result, outPath)
	}
	result.Validation["h2_qualified"] = true
	result.Validation["selected_precision_bits"] = selected
	// H4-H6 resume from the qualified P_O in the follow-up portion of this runner.
	result.Classification = "logn13_high_precision_oracle_recovery_selected"
	result.FirstRemainingBlocker = "qualified oracle requires H4-H6 continuation"
	return fix001P3HighPrecisionOracleWrite(result, outPath)
}
