package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const (
	requiredFIX001P3E2EPrimaryCommit   = "9e98a1a9a2a52c4a40ad8107582f7a3e966cc0bf"
	requiredFIX001P3E2ESecondaryCommit = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
)

var acceptedFIX001P3E2ECoreRows = []FinalizationRowFingerprint{
	{Component: 0, Limb: 0, Length: 8192, SHA256: "a6abfbab5dbbb0086baf0a359c41645a3ce75c70face718424e49de6ef177f85"},
	{Component: 0, Limb: 1, Length: 8192, SHA256: "46be461b5ec721c1db201fc32d80592913d185acdfce3ee8c3f844d12422afb9"},
	{Component: 1, Limb: 0, Length: 8192, SHA256: "de2f256064a0af797747c2b97505dc0b9f3df0de4f489eac731c23ae9ca9cc31"},
	{Component: 1, Limb: 1, Length: 8192, SHA256: "de2f256064a0af797747c2b97505dc0b9f3df0de4f489eac731c23ae9ca9cc31"},
}

type fix001P3E2EResult struct {
	SchemaVersion                 string                          `json:"schema_version"`
	Timestamp                     time.Time                       `json:"timestamp"`
	Primary                       RepositoryMetadata              `json:"primary_repository"`
	Lattigo                       RepositoryMetadata              `json:"lattigo_repository"`
	Environment                   EnvironmentMetadata             `json:"environment"`
	Config                        BootstrapConfig                 `json:"config"`
	Parameters                    ExperimentParameters            `json:"effective_parameters"`
	Workload                      CorrectnessWorkload             `json:"workload"`
	ProductionSystemUnderTest     string                          `json:"production_system_under_test"`
	PredecessorCoreProvenance     string                          `json:"predecessor_core_provenance"`
	PreviousFastReplayDisposition string                          `json:"previous_fast_replay_disposition"`
	CoreOutput                    *FinalizationCiphertextEvidence `json:"core_output,omitempty"`
	CorePreconditionRows          []FinalizationRowFingerprint    `json:"core_precondition_expected_rows"`
	CorePreconditionRowsMatch     bool                            `json:"core_precondition_rows_match"`
	CorePreconditionMetadataMatch bool                            `json:"core_precondition_metadata_match"`
	PackProfile                   map[string]interface{}          `json:"pack_profile"`
	OriginalInputUnchanged        bool                            `json:"original_input_unchanged"`
	UnpackInput                   *FinalizationCiphertextEvidence `json:"unpack_input,omitempty"`
	UnpackedOutput                *FinalizationCiphertextEvidence `json:"unpacked_output,omitempty"`
	ExpectedUnpackedOutput        *FinalizationCiphertextEvidence `json:"expected_unpacked_output,omitempty"`
	UnpackRowsMatch               bool                            `json:"unpack_rows_match"`
	UnpackMetadataMatch           bool                            `json:"unpack_metadata_match"`
	UnpackInputUnchanged          bool                            `json:"unpack_input_unchanged"`
	FinalizationBefore            *FinalizationCiphertextEvidence `json:"finalization_oracle_before,omitempty"`
	FinalizationAfter             *FinalizationCiphertextEvidence `json:"finalization_oracle_after,omitempty"`
	FinalizationRowsMatch         bool                            `json:"finalization_oracle_rows_match"`
	FinalizationMetadataMatch     bool                            `json:"finalization_oracle_metadata_match"`
	PublicBootstrapOutput         *FinalizationCiphertextEvidence `json:"public_bootstrap_output,omitempty"`
	PublicBootstrapRowsMatch      bool                            `json:"public_bootstrap_rows_match"`
	PublicBootstrapMetadataMatch  bool                            `json:"public_bootstrap_metadata_match"`
	PublicInputUnchanged          bool                            `json:"public_input_unchanged"`
	PublicBootstrapManyOutput     *FinalizationCiphertextEvidence `json:"public_bootstrap_many_output,omitempty"`
	BootstrapManyRowsMatch        bool                            `json:"bootstrap_many_rows_match"`
	BootstrapManyMetadataMatch    bool                            `json:"bootstrap_many_metadata_match"`
	BootstrapManyInputUnchanged   bool                            `json:"bootstrap_many_input_unchanged"`
	Semantic                      *CorrectnessMetrics             `json:"semantic,omitempty"`
	ManualOracleSemantic          *CorrectnessMetrics             `json:"manual_oracle_semantic,omitempty"`
	ManualVsPublicRowsMatch       bool                            `json:"manual_oracle_vs_public_rows_match"`
	FirstFailingCheckpoint        string                          `json:"first_failing_checkpoint"`
	FirstSupportedCause           string                          `json:"first_supported_cause"`
	Validation                    map[string]interface{}          `json:"validation"`
}

func fix001P3E2EWrite(result fix001P3E2EResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func fix001P3E2EMetadataEqual(a, b *rlwe.Ciphertext) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Level() == b.Level() && a.Degree() == b.Degree() && a.N() == b.N() && a.Scale.Equal(b.Scale) && a.LogDimensions == b.LogDimensions && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery
}

func fix001P3E2ERowsEqual(a, b *rlwe.Ciphertext) bool {
	if a == nil || b == nil || a.Level() != b.Level() || a.Degree() != b.Degree() {
		return false
	}
	return finalizationRowsEqual(compareFinalizationRows(a, b))
}

func fix001P3E2EExpectedUnpack(coreOut, input *rlwe.Ciphertext, params bootstrapping.Parameters) (*rlwe.Ciphertext, error) {
	residual := params.ResidualParameters
	bootstrap := params.BootstrappingParameters
	if residual.N() != bootstrap.N() {
		return nil, fmt.Errorf("expected unpack profile requires residual N == bootstrap N")
	}
	if input.LogSlots() != bootstrap.LogMaxDimensions().Cols {
		return nil, fmt.Errorf("expected unpack profile requires full-slot input: logSlots=%d maxCols=%d", input.LogSlots(), bootstrap.LogMaxDimensions().Cols)
	}
	expected := coreOut.CopyNew()
	// Source fastUnpack makes one copy when logPackCTs == 0, then the public
	// boundary sets only the column dimension to the original log-slot count.
	expected.LogDimensions.Cols = input.LogSlots()
	return expected, nil
}

func fix001P3E2EFinalizeOracle(input *rlwe.Ciphertext, params bootstrapping.Parameters) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	if input == nil || input.N() != params.ResidualParameters.N() || input.Degree() != 1 || input.Level() != params.ResidualParameters.MaxLevel() {
		return nil, nil, fmt.Errorf("finalization oracle input geometry does not match residual parameters")
	}
	if !input.IsNTT || !input.IsMontgomery {
		return nil, nil, fmt.Errorf("finalization oracle input must be NTT Montgomery")
	}
	before := input.CopyNew()
	after := input.CopyNew()
	ringQ := params.ResidualParameters.RingQ()
	maintained := after.Level() + 1
	if maintained > 2 {
		maintained = 2
	}
	for component := 0; component <= 1; component++ {
		for limb := 0; limb < maintained; limb++ {
			ringQ.SubRings[limb].IMForm(after.Value[component].Coeffs[limb], after.Value[component].Coeffs[limb])
		}
	}
	after.IsMontgomery = false
	after.Scale = params.ResidualParameters.DefaultScale()
	return before, after, nil
}

func fix001P3E2EFinalizationRowsMatch(before, after *rlwe.Ciphertext, params bootstrapping.Parameters) bool {
	if before == nil || after == nil {
		return false
	}
	expected := before.CopyNew()
	ringQ := params.ResidualParameters.RingQ()
	maintained := expected.Level() + 1
	if maintained > 2 {
		maintained = 2
	}
	for component := 0; component <= 1; component++ {
		for limb := 0; limb < maintained; limb++ {
			ringQ.SubRings[limb].IMForm(expected.Value[component].Coeffs[limb], expected.Value[component].Coeffs[limb])
		}
	}
	return fix001P3E2ERowsEqual(expected, after)
}

func fix001P3E2EInputUnchanged(before, after *rlwe.Ciphertext) bool {
	return before != nil && after != nil && before.Level() == after.Level() && before.Degree() == after.Degree() && before.N() == after.N() && before.Scale.Equal(after.Scale) && before.IsNTT == after.IsNTT && before.IsMontgomery == after.IsMontgomery && doubleAngleRowsMatch(finalizationEvidence(before).Rows, finalizationEvidence(after).Rows)
}

func runFIX001P3ValidateLogN13E2E(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	eval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return err
	}
	input := reproducibleInput(residual, btp)
	inputBefore := input.CopyNew()
	result := fix001P3E2EResult{
		SchemaVersion:                 "fix-001-p3-validate-logn13-end-to-end.v1",
		Timestamp:                     time.Now().UTC(),
		Primary:                       gitMetadata(primaryRoot),
		Lattigo:                       gitMetadata(backendRoot),
		Environment:                   EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:                        cfg,
		Parameters:                    parameterMetadata(residual, btp),
		Workload:                      CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		ProductionSystemUnderTest:     "FastEvaluator.Bootstrap(input)",
		PredecessorCoreProvenance:     requiredFIX001P3E2EPrimaryCommit,
		PreviousFastReplayDisposition: "invalid_comparison_due_to_asymmetric_montgomery_normalization",
		CorePreconditionRows:          append([]FinalizationRowFingerprint(nil), acceptedFIX001P3E2ECoreRows...),
		FirstFailingCheckpoint:        "none",
		FirstSupportedCause:           "logn13_e2e_core_precondition_mismatch",
		Validation: map[string]interface{}{
			"primary_required_base":           requiredFIX001P3E2EPrimaryCommit,
			"secondary_commit":                gitOutput(backendRoot, "rev-parse", "HEAD"),
			"secondary_clean":                 gitOutput(backendRoot, "status", "--porcelain") == "",
			"secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredFIX001P3E2ESecondaryCommit,
			"no_secondary_production_changes": true,
			"no_logn16":                       true,
			"no_benchmark":                    true,
			"no_gate_4_or_5":                  true,
			"no_exp003":                       true,
		},
	}

	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input})
	if err != nil {
		return fmt.Errorf("E1 PackAndSwitchN1ToN2: %w", err)
	}
	if len(packed) != 1 {
		return fmt.Errorf("E1 expected one packed ciphertext, got %d", len(packed))
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return fmt.Errorf("E1 ScaleDown: %w", err)
	}
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return fmt.Errorf("E1 ModUp: %w", err)
	}
	ctReal, ctImag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return fmt.Errorf("E1 CoeffsToSlots: %w", err)
	}
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return fmt.Errorf("E1 EvalMod real: %w", err)
	}
	if ctImag != nil {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return fmt.Errorf("E1 EvalMod imag: %w", err)
		}
	}
	coreOut, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return fmt.Errorf("E1 SlotsToCoeffs: %w", err)
	}
	coreEvidence := finalizationEvidence(coreOut)
	result.CoreOutput = &coreEvidence
	result.CorePreconditionRowsMatch = doubleAngleRowsMatch(coreEvidence.Rows, acceptedFIX001P3E2ECoreRows)
	result.OriginalInputUnchanged = fix001P3E2EInputUnchanged(inputBefore, input)
	result.CorePreconditionMetadataMatch = coreOut.Level() == residual.MaxLevel() && coreOut.Degree() == 1 && coreOut.N() == btp.BootstrappingParameters.N() && coreOut.IsNTT && coreOut.IsMontgomery && coreOut.Scale.Equal(residual.DefaultScale()) && coreOut.LogSlots() == input.LogSlots() && result.OriginalInputUnchanged
	result.PackProfile = map[string]interface{}{
		"input_n": input.N(), "bootstrap_n": btp.BootstrappingParameters.N(), "residual_n": residual.N(), "input_log_slots": input.LogSlots(),
		"bootstrap_log_max_columns": btp.LogMaxDimensions().Cols, "ctxt_n1_present": ctxtN1 != nil, "expected_ring_degree_switch": false,
	}
	if !result.CorePreconditionRowsMatch || !result.CorePreconditionMetadataMatch {
		return fix001P3E2EWrite(result, outPath)
	}

	coreBeforeUnpack := coreOut.CopyNew()
	unpacked, err := eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*coreOut.CopyNew()}, ctxtN1, ctxtN2)
	if err != nil {
		result.FirstSupportedCause = "logn13_e2e_unpack_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	if len(unpacked) != 1 {
		result.FirstSupportedCause = "logn13_e2e_unpack_mismatch"
		return fix001P3E2EWrite(result, outPath)
	}
	unpackedCT := &unpacked[0]
	unpackInput := finalizationEvidence(coreBeforeUnpack)
	unpackOutput := finalizationEvidence(unpackedCT)
	result.UnpackInput = &unpackInput
	result.UnpackedOutput = &unpackOutput
	expectedUnpacked, err := fix001P3E2EExpectedUnpack(coreBeforeUnpack, input, btp)
	if err != nil {
		result.FirstSupportedCause = "logn13_e2e_unpack_mismatch"
		return fix001P3E2EWrite(result, outPath)
	}
	expectedEvidence := finalizationEvidence(expectedUnpacked)
	result.ExpectedUnpackedOutput = &expectedEvidence
	result.UnpackRowsMatch = fix001P3E2ERowsEqual(unpackedCT, expectedUnpacked)
	result.UnpackMetadataMatch = fix001P3E2EMetadataEqual(unpackedCT, expectedUnpacked)
	result.UnpackInputUnchanged = fix001P3E2EInputUnchanged(coreBeforeUnpack, coreOut)
	if !result.UnpackRowsMatch || !result.UnpackMetadataMatch || !result.UnpackInputUnchanged {
		result.FirstSupportedCause = "logn13_e2e_unpack_mismatch"
		return fix001P3E2EWrite(result, outPath)
	}

	finalBefore, finalOracle, err := fix001P3E2EFinalizeOracle(unpackedCT, btp)
	if err != nil {
		result.FirstSupportedCause = "logn13_e2e_finalization_oracle_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	finalBeforeEvidence := finalizationEvidence(finalBefore)
	finalAfterEvidence := finalizationEvidence(finalOracle)
	result.FinalizationBefore = &finalBeforeEvidence
	result.FinalizationAfter = &finalAfterEvidence
	result.FinalizationRowsMatch = fix001P3E2EFinalizationRowsMatch(finalBefore, finalOracle, btp)
	result.FinalizationMetadataMatch = finalOracle.Level() == residual.MaxLevel() && finalOracle.Degree() == 1 && finalOracle.N() == residual.N() && finalOracle.IsNTT && !finalOracle.IsMontgomery && finalOracle.Scale.Equal(residual.DefaultScale()) && finalOracle.LogDimensions == finalBefore.LogDimensions
	if !result.FinalizationRowsMatch || !result.FinalizationMetadataMatch {
		result.FirstSupportedCause = "logn13_e2e_finalization_oracle_failure"
		return fix001P3E2EWrite(result, outPath)
	}

	publicInput := reproducibleInput(residual, btp)
	publicInputBefore := publicInput.CopyNew()
	publicOutput, err := eval.Bootstrap(publicInput)
	if err != nil {
		result.FirstSupportedCause = "logn13_e2e_public_bootstrap_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	publicEvidence := finalizationEvidence(publicOutput)
	result.PublicBootstrapOutput = &publicEvidence
	result.PublicBootstrapRowsMatch = fix001P3E2ERowsEqual(publicOutput, finalOracle)
	result.PublicBootstrapMetadataMatch = fix001P3E2EMetadataEqual(publicOutput, finalOracle)
	result.PublicInputUnchanged = fix001P3E2EInputUnchanged(publicInputBefore, publicInput)
	if !result.PublicBootstrapRowsMatch || !result.PublicBootstrapMetadataMatch || !result.PublicInputUnchanged {
		result.FirstSupportedCause = "logn13_e2e_public_structural_mismatch"
		return fix001P3E2EWrite(result, outPath)
	}

	decoded, err := decodeWithSecret(residual, publicOutput, zeroSecret(residual))
	if err != nil {
		result.FirstSupportedCause = "logn13_e2e_semantic_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	if err := chebyshevCheckFinite("public Bootstrap decoded", decoded); err != nil {
		result.FirstSupportedCause = "logn13_e2e_semantic_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	semantic, err := compareComplexVectors(reproducibleValues(residual.MaxSlots()), decoded, correctnessThreshold)
	if err != nil {
		return err
	}
	result.Semantic = &semantic
	manualDecoded, err := decodeWithSecret(residual, finalOracle, zeroSecret(residual))
	if err != nil {
		return err
	}
	manualSemantic, err := compareComplexVectors(reproducibleValues(residual.MaxSlots()), manualDecoded, correctnessThreshold)
	if err != nil {
		return err
	}
	result.ManualOracleSemantic = &manualSemantic
	semanticPass := semantic.PassThreshold
	if !semanticPass {
		result.FirstFailingCheckpoint = "E5: public_bootstrap_plaintext_semantics"
	}

	manyInput := reproducibleInput(residual, btp)
	manyInputBefore := manyInput.CopyNew()
	manyOutput, err := eval.BootstrapMany([]rlwe.Ciphertext{*manyInput})
	if err != nil {
		result.FirstSupportedCause = "logn13_e2e_bootstrap_many_consistency_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	if len(manyOutput) != 1 {
		result.FirstSupportedCause = "logn13_e2e_bootstrap_many_consistency_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	manyCT := &manyOutput[0]
	manyEvidence := finalizationEvidence(manyCT)
	result.PublicBootstrapManyOutput = &manyEvidence
	result.BootstrapManyRowsMatch = fix001P3E2ERowsEqual(manyCT, publicOutput)
	result.BootstrapManyMetadataMatch = fix001P3E2EMetadataEqual(manyCT, publicOutput)
	result.BootstrapManyInputUnchanged = fix001P3E2EInputUnchanged(manyInputBefore, manyInput)
	result.ManualVsPublicRowsMatch = result.PublicBootstrapRowsMatch
	if !result.BootstrapManyRowsMatch || !result.BootstrapManyMetadataMatch || !result.BootstrapManyInputUnchanged {
		result.FirstFailingCheckpoint = "E6: bootstrap_many_one_element_consistency"
		result.FirstSupportedCause = "logn13_e2e_bootstrap_many_consistency_failure"
		return fix001P3E2EWrite(result, outPath)
	}
	if !semanticPass {
		result.FirstSupportedCause = "logn13_e2e_semantic_failure"
		return fix001P3E2EWrite(result, outPath)
	}

	result.FirstSupportedCause = "logn13_fast_bootstrap_end_to_end_validated"
	return fix001P3E2EWrite(result, outPath)
}
