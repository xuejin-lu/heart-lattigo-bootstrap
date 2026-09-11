package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type FinalizationRowFingerprint struct {
	Component int    `json:"component"`
	Limb      int    `json:"limb"`
	Length    int    `json:"length"`
	SHA256    string `json:"sha256"`
}

type FinalizationCiphertextEvidence struct {
	Level               int                          `json:"level"`
	Scale               string                       `json:"scale"`
	ScaleMod            string                       `json:"scale_mod,omitempty"`
	ScaleFloat64        float64                      `json:"scale_float64"`
	ScaleLog2           float64                      `json:"scale_log2"`
	Degree              int                          `json:"degree"`
	N                   int                          `json:"n"`
	LogSlots            int                          `json:"log_slots"`
	IsNTT               bool                         `json:"is_ntt"`
	IsMontgomery        bool                         `json:"is_montgomery"`
	MaintainedLimbCount int                          `json:"maintained_limb_count"`
	Rows                []FinalizationRowFingerprint `json:"maintained_rows_q0_q1"`
}

type FinalizationRowComparison struct {
	Component          int    `json:"component"`
	Limb               int    `json:"limb"`
	Equal              bool   `json:"equal"`
	BeforeLength       int    `json:"before_length"`
	AfterLength        int    `json:"after_length"`
	BeforeSHA256       string `json:"before_sha256"`
	AfterSHA256        string `json:"after_sha256"`
	FirstMismatchIndex *int   `json:"first_mismatch_index,omitempty"`
}

type FinalizationBoundaryComparison struct {
	F0Metadata          FinalizationCiphertextEvidence `json:"f0_metadata"`
	F1Metadata          FinalizationCiphertextEvidence `json:"f1_metadata"`
	MetadataEqual       bool                           `json:"metadata_equal"`
	MaintainedRowsEqual bool                           `json:"maintained_rows_equal"`
	ExactEqual          bool                           `json:"exact_equal"`
	Rows                []FinalizationRowComparison    `json:"rows"`
	Classification      string                         `json:"classification"`
}

type FinalizationRoundTrip struct {
	Rows           []FinalizationRowComparison `json:"rows"`
	ExactEqual     bool                        `json:"exact_equal"`
	Classification string                      `json:"classification"`
}

type FinalizationLogicalEvidence struct {
	MaxAbsComplex        float64 `json:"max_abs_complex"`
	MeanAbsComplex       float64 `json:"mean_abs_complex"`
	RMSEComplex          float64 `json:"rmse_complex"`
	MaxAbsReal           float64 `json:"max_abs_real"`
	MaxAbsImag           float64 `json:"max_abs_imag"`
	MaxComponentAbs      float64 `json:"max_component_abs"`
	MaxAbsComplexIndices []int   `json:"max_abs_complex_indices"`
	MaxComponentIndices  []int   `json:"max_component_indices"`
	MaxAbsRealIndices    []int   `json:"max_abs_real_indices"`
	MaxAbsImagIndices    []int   `json:"max_abs_imag_indices"`
	PassThreshold        bool    `json:"pass_threshold"`
	Threshold            float64 `json:"threshold"`
}

type FinalizationCheckpoint struct {
	Name                   string                         `json:"name"`
	Present                bool                           `json:"present"`
	Method                 string                         `json:"method"`
	Ciphertext             FinalizationCiphertextEvidence `json:"ciphertext"`
	GeneratedSecretDecoded []CorrectnessValue             `json:"generated_secret_decoded,omitempty"`
	ZeroSecretDecoded      []CorrectnessValue             `json:"zero_secret_decoded,omitempty"`
	GeneratedVsInput       *FinalizationLogicalEvidence   `json:"generated_vs_input,omitempty"`
	ZeroVsInput            *FinalizationLogicalEvidence   `json:"zero_vs_input,omitempty"`
	ZeroVsStandard         *FinalizationLogicalEvidence   `json:"zero_vs_standard,omitempty"`
	DecodeErrors           []string                       `json:"decode_errors,omitempty"`
}

type FinalizationScaleEvidence struct {
	S2CScale                    string  `json:"s2c_scale"`
	S2CScaleFloat64             float64 `json:"s2c_scale_float64"`
	ResidualDefaultScale        string  `json:"residual_default_scale"`
	ResidualDefaultScaleFloat64 float64 `json:"residual_default_scale_float64"`
	Ratio                       float64 `json:"s2c_over_residual_default_ratio"`
	F2ScaleChanged              bool    `json:"f2_scale_changed"`
	F3ScaleChanged              bool    `json:"f3_scale_changed"`
}

type FinalizationDiagnosticResult struct {
	SchemaVersion            string                          `json:"schema_version"`
	Timestamp                time.Time                       `json:"timestamp"`
	Primary                  RepositoryMetadata              `json:"primary_repository"`
	Lattigo                  RepositoryMetadata              `json:"lattigo_repository"`
	Environment              EnvironmentMetadata             `json:"environment"`
	Config                   BootstrapConfig                 `json:"config"`
	Parameters               ExperimentParameters            `json:"effective_parameters"`
	Workload                 CorrectnessWorkload             `json:"workload"`
	Input                    []CorrectnessValue              `json:"input"`
	StandardReferencePath    string                          `json:"standard_reference_path,omitempty"`
	OfficialOutput           FinalizationCheckpoint          `json:"official_output"`
	F0SlotsToCoeffsRaw       *FinalizationCheckpoint         `json:"f0_slots_to_coeffs_raw,omitempty"`
	F1AfterUnpackRaw         *FinalizationCheckpoint         `json:"f1_after_unpack_raw,omitempty"`
	F2ImformOnly             *FinalizationCheckpoint         `json:"f2_imform_only,omitempty"`
	F2RRoundTrip             *FinalizationRoundTrip          `json:"f2r_imform_mform_roundtrip,omitempty"`
	F3FullFinalizationView   *FinalizationCheckpoint         `json:"f3_full_finalization_view,omitempty"`
	F4OfficialFastOutput     *FinalizationCheckpoint         `json:"f4_official_fast_output,omitempty"`
	UnpackBoundary           *FinalizationBoundaryComparison `json:"unpack_boundary,omitempty"`
	ScaleEvidence            *FinalizationScaleEvidence      `json:"scale_evidence,omitempty"`
	F3VsF4                   *FinalizationLogicalEvidence    `json:"f3_vs_f4,omitempty"`
	FinalizationMode         string                          `json:"finalization_mode"`
	FirstSupportedCause      string                          `json:"first_supported_cause"`
	DiagnosticClassification string                          `json:"diagnostic_classification"`
	Notes                    []string                        `json:"notes,omitempty"`
}

type finalizationStandardReference struct {
	OfficialOutput FinalizationCheckpoint `json:"official_output"`
}

func finalizationScaleString(scale rlwe.Scale) string {
	return scale.Value.Text('e', rlwe.ScalePrecisionLog10)
}

func finalizationScaleModString(scale rlwe.Scale) string {
	if scale.Mod == nil {
		return ""
	}
	return scale.Mod.String()
}

func finalizationDiagnosticCopy(ct *rlwe.Ciphertext) *rlwe.Ciphertext {
	if ct == nil {
		return nil
	}
	return ct.CopyNew()
}

func finalizationRowHash(row []uint64) string {
	hash := sha256.New()
	var encoded [8]byte
	for _, value := range row {
		binary.LittleEndian.PutUint64(encoded[:], value)
		_, _ = hash.Write(encoded[:])
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func finalizationEvidence(ct *rlwe.Ciphertext) FinalizationCiphertextEvidence {
	evidence := FinalizationCiphertextEvidence{
		Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), ScaleMod: finalizationScaleModString(ct.Scale), ScaleFloat64: ct.Scale.Float64(), ScaleLog2: ct.Scale.Log2(),
		Degree: ct.Degree(), N: ct.N(), LogSlots: ct.LogSlots(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		Rows: make([]FinalizationRowFingerprint, 0, 2*(ct.Degree()+1)),
	}
	if len(ct.Value) > 0 {
		for limb := 0; limb < 2; limb++ {
			if limb < len(ct.Value[0].Coeffs) && len(ct.Value[0].Coeffs[limb]) > 0 {
				evidence.MaintainedLimbCount++
			}
		}
	}
	for component := 0; component <= ct.Degree() && component < len(ct.Value); component++ {
		for limb := 0; limb < 2; limb++ {
			var row []uint64
			if limb < len(ct.Value[component].Coeffs) {
				row = ct.Value[component].Coeffs[limb]
			}
			evidence.Rows = append(evidence.Rows, FinalizationRowFingerprint{Component: component, Limb: limb, Length: len(row), SHA256: finalizationRowHash(row)})
		}
	}
	return evidence
}

func finalizationRowComparison(component, limb int, before, after []uint64) FinalizationRowComparison {
	comparison := FinalizationRowComparison{Component: component, Limb: limb, BeforeLength: len(before), AfterLength: len(after), BeforeSHA256: finalizationRowHash(before), AfterSHA256: finalizationRowHash(after), Equal: len(before) == len(after)}
	limit := len(before)
	if len(after) < limit {
		limit = len(after)
	}
	for i := 0; i < limit; i++ {
		if before[i] != after[i] {
			comparison.Equal = false
			index := i
			comparison.FirstMismatchIndex = &index
			break
		}
	}
	if comparison.Equal && len(before) != len(after) {
		index := limit
		comparison.FirstMismatchIndex = &index
	}
	return comparison
}

func compareFinalizationRows(before, after *rlwe.Ciphertext) []FinalizationRowComparison {
	rows := make([]FinalizationRowComparison, 0, 2*(before.Degree()+1))
	for component := 0; component <= before.Degree() && component < len(before.Value); component++ {
		for limb := 0; limb < 2; limb++ {
			var beforeRow, afterRow []uint64
			if limb < len(before.Value[component].Coeffs) {
				beforeRow = before.Value[component].Coeffs[limb]
			}
			if component < len(after.Value) && limb < len(after.Value[component].Coeffs) {
				afterRow = after.Value[component].Coeffs[limb]
			}
			rows = append(rows, finalizationRowComparison(component, limb, beforeRow, afterRow))
		}
	}
	return rows
}

func finalizationRowsEqual(rows []FinalizationRowComparison) bool {
	for _, row := range rows {
		if !row.Equal {
			return false
		}
	}
	return true
}

func finalizationMetadataEqual(before, after *rlwe.Ciphertext) bool {
	return before.Level() == after.Level() && before.Scale.Equal(after.Scale) && before.Degree() == after.Degree() && before.N() == after.N() && before.LogSlots() == after.LogSlots() && before.IsNTT == after.IsNTT && before.IsMontgomery == after.IsMontgomery
}

func finalizationBoundaryComparison(before, after *rlwe.Ciphertext) FinalizationBoundaryComparison {
	rows := compareFinalizationRows(before, after)
	rowsEqual := finalizationRowsEqual(rows)
	metadataEqual := finalizationMetadataEqual(before, after)
	classification := "MATCH"
	if !rowsEqual {
		classification = "UNPACK_BOUNDARY_CHANGED_VALUE"
	} else if !metadataEqual {
		classification = "UNPACK_BOUNDARY_CHANGED_METADATA"
	}
	return FinalizationBoundaryComparison{F0Metadata: finalizationEvidence(before), F1Metadata: finalizationEvidence(after), MetadataEqual: metadataEqual, MaintainedRowsEqual: rowsEqual, ExactEqual: metadataEqual && rowsEqual, Rows: rows, Classification: classification}
}

func finalizationTransformMaintained(ct *rlwe.Ciphertext, params ckks.Parameters, toMontgomery bool) error {
	if ct == nil || ct.Degree() < 1 || ct.Level() < 1 || len(ct.Value) < 2 {
		return fmt.Errorf("expected degree-one level-at-least-one ciphertext with c0/c1")
	}
	if len(ct.Value[0].Coeffs) < 2 || len(ct.Value[1].Coeffs) < 2 || len(ct.Value[0].Coeffs[0]) == 0 || len(ct.Value[1].Coeffs[0]) == 0 {
		return fmt.Errorf("expected maintained q0/q1 rows for c0/c1")
	}
	ringQ := params.RingQ()
	for component := 0; component <= 1; component++ {
		for limb := 0; limb < 2; limb++ {
			row := ct.Value[component].Coeffs[limb]
			if toMontgomery {
				ringQ.SubRings[limb].MForm(row, row)
			} else {
				ringQ.SubRings[limb].IMForm(row, row)
			}
		}
	}
	return nil
}

func finalizationRawViewFromPublicOutput(publicOutput, rawBefore *rlwe.Ciphertext, residual ckks.Parameters) (*rlwe.Ciphertext, error) {
	if publicOutput == nil || rawBefore == nil {
		return nil, fmt.Errorf("missing public output or raw pre-unpack ciphertext")
	}
	if !rawBefore.IsMontgomery || publicOutput.IsMontgomery {
		return nil, fmt.Errorf("cannot reconstruct raw F1: expected Montgomery F0 and non-Montgomery public output")
	}
	raw := finalizationDiagnosticCopy(publicOutput)
	if err := finalizationTransformMaintained(raw, residual, true); err != nil {
		return nil, fmt.Errorf("restore Montgomery rows for raw F1: %w", err)
	}
	raw.IsMontgomery = true
	raw.Scale = rawBefore.Scale
	return raw, nil
}

func finalizationValues(values []CorrectnessValue) []complex128 {
	result := make([]complex128, len(values))
	for i, value := range values {
		result[i] = complex(value.Real, value.Imag)
	}
	return result
}

func finalizationLogicalEvidence(reference, actual []complex128) (*FinalizationLogicalEvidence, error) {
	metrics, err := compareComplexVectors(reference, actual, correctnessThreshold)
	if err != nil {
		return nil, err
	}
	return &FinalizationLogicalEvidence{MaxAbsComplex: metrics.MaxAbsComplex, MeanAbsComplex: metrics.MeanAbsComplex, RMSEComplex: metrics.RMSEComplex, MaxAbsReal: metrics.MaxAbsReal, MaxAbsImag: metrics.MaxAbsImag, MaxComponentAbs: metrics.MaxComponentAbs, MaxAbsComplexIndices: metrics.MaxAbsComplexIndices, MaxComponentIndices: metrics.MaxComponentIndices, MaxAbsRealIndices: metrics.MaxAbsRealIndices, MaxAbsImagIndices: metrics.MaxAbsImagIndices, PassThreshold: metrics.PassThreshold, Threshold: metrics.Threshold}, nil
}

func finalizationCheckpoint(name, method string, ct *rlwe.Ciphertext, params ckks.Parameters, generatedSecret, zeroSecretKey *rlwe.SecretKey, input, standard []complex128) FinalizationCheckpoint {
	checkpoint := FinalizationCheckpoint{Name: name, Present: ct != nil, Method: method}
	if ct == nil {
		return checkpoint
	}
	checkpoint.Ciphertext = finalizationEvidence(ct)
	generated, generatedErr := diagnosticDecode(params, ct, generatedSecret)
	zero, zeroErr := diagnosticDecode(params, ct, zeroSecretKey)
	if generatedErr != nil {
		checkpoint.DecodeErrors = append(checkpoint.DecodeErrors, fmt.Sprintf("generated_secret: %v", generatedErr))
	} else {
		checkpoint.GeneratedSecretDecoded = correctnessValues(generated)
		if evidence, err := finalizationLogicalEvidence(input, generated); err == nil {
			checkpoint.GeneratedVsInput = evidence
		}
	}
	if zeroErr != nil {
		checkpoint.DecodeErrors = append(checkpoint.DecodeErrors, fmt.Sprintf("zero_secret: %v", zeroErr))
	} else {
		checkpoint.ZeroSecretDecoded = correctnessValues(zero)
		if evidence, err := finalizationLogicalEvidence(input, zero); err == nil {
			checkpoint.ZeroVsInput = evidence
		}
		if len(standard) == len(zero) {
			if evidence, err := finalizationLogicalEvidence(standard, zero); err == nil {
				checkpoint.ZeroVsStandard = evidence
			}
		}
	}
	return checkpoint
}

func runFinalizationStages(eval *bootstrapping.Evaluator, residual ckks.Parameters, btp bootstrapping.Parameters) (*rlwe.Ciphertext, func() ([]rlwe.Ciphertext, error), error) {
	input := reproducibleInput(residual, btp)
	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input})
	if err != nil {
		return nil, nil, fmt.Errorf("PackAndSwitchN1ToN2: %w", err)
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return nil, nil, fmt.Errorf("ScaleDown: %w", err)
	}
	packed[0] = *scaled
	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return nil, nil, fmt.Errorf("ModUp: %w", err)
	}
	packed[0] = *modUp
	ctReal, ctImag, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return nil, nil, fmt.Errorf("CoeffsToSlots: %w", err)
	}
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return nil, nil, fmt.Errorf("EvalMod real: %w", err)
	}
	if ctImag != nil {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return nil, nil, fmt.Errorf("EvalMod imag: %w", err)
		}
	}
	ctOut, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return nil, nil, fmt.Errorf("SlotsToCoeffs: %w", err)
	}
	unpack := func() ([]rlwe.Ciphertext, error) {
		return eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*finalizationDiagnosticCopy(ctOut)}, ctxtN1, ctxtN2)
	}
	return finalizationDiagnosticCopy(ctOut), unpack, nil
}

func readFinalizationStandardReference(path string) ([]complex128, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Standard reference: %w", err)
	}
	var reference finalizationStandardReference
	if err := json.Unmarshal(data, &reference); err != nil {
		return nil, fmt.Errorf("decode Standard reference: %w", err)
	}
	if len(reference.OfficialOutput.GeneratedSecretDecoded) == 0 {
		return nil, fmt.Errorf("Standard reference has no generated-secret decoded output")
	}
	return finalizationValues(reference.OfficialOutput.GeneratedSecretDecoded), nil
}

func RunFinalizationBoundaryExperiment(cfg BootstrapConfig, primaryRoot, backendRoot, standardReferencePath string) (FinalizationDiagnosticResult, error) {
	residual, btpParams, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return FinalizationDiagnosticResult{}, err
	}
	generatedSecret := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	evalKeys, _, err := btpParams.GenEvaluationKeys(generatedSecret)
	if err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("generate evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(btpParams, evalKeys)
	if err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("construct bootstrap evaluator: %w", err)
	}
	zeroSecretKey := zeroSecret(residual)
	input := reproducibleValues(residual.MaxSlots())
	standardReference, err := readFinalizationStandardReference(standardReferencePath)
	if err != nil {
		return FinalizationDiagnosticResult{}, err
	}
	officialOutput, err := eval.Bootstrap(reproducibleInput(residual, btpParams))
	if err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("official Bootstrap: %w", err)
	}
	officialCheckpoint := finalizationCheckpoint("official_public_output", "ordinary_public_bootstrap_output", officialOutput, residual, generatedSecret, zeroSecretKey, input, standardReference)
	result := FinalizationDiagnosticResult{
		SchemaVersion: "exp-002c-diag-p.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      cfg, Parameters: parameterMetadata(residual, btpParams), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		Input: correctnessValues(input), StandardReferencePath: standardReferencePath, OfficialOutput: officialCheckpoint,
		FinalizationMode:         "not_applicable: S2C output is not Montgomery",
		FirstSupportedCause:      "not_applicable_for_this_backend",
		DiagnosticClassification: "STANDARD_REFERENCE_ONLY",
	}

	f0, unpack, err := runFinalizationStages(eval, residual, btpParams)
	if err != nil {
		return FinalizationDiagnosticResult{}, err
	}
	if !f0.IsMontgomery {
		result.Notes = append(result.Notes, "SlotsToCoeffs output is non-Montgomery; Fast finalization checkpoints are not applicable for this Standard/reference execution.")
		return result, nil
	}
	result.FinalizationMode = "Fast public finalization diagnostic"
	result.DiagnosticClassification = "PENDING"
	result.FirstSupportedCause = "PENDING"
	result.F0SlotsToCoeffsRaw = ptrFinalizationCheckpoint(finalizationCheckpoint("F0_slots_to_coeffs_raw", "raw_pre_unpack_montgomery", f0, btpParams.BootstrappingParameters, generatedSecret, zeroSecret(btpParams.BootstrappingParameters), input, standardReference))
	unpacked, err := unpack()
	if err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("public UnpackAndSwitchN2ToN1: %w", err)
	}
	if len(unpacked) != 1 {
		return FinalizationDiagnosticResult{}, fmt.Errorf("public unpack returned %d ciphertexts, want 1", len(unpacked))
	}
	f1Public := &unpacked[0]
	f1Raw, err := finalizationRawViewFromPublicOutput(f1Public, f0, residual)
	if err != nil {
		return FinalizationDiagnosticResult{}, err
	}
	result.F1AfterUnpackRaw = ptrFinalizationCheckpoint(finalizationCheckpoint("F1_after_unpack_raw", "public_unpack_then_inverse_diagnostic_finalizer", f1Raw, residual, generatedSecret, zeroSecretKey, input, standardReference))
	boundary := finalizationBoundaryComparison(f0, f1Raw)
	result.UnpackBoundary = &boundary

	f2 := finalizationDiagnosticCopy(f1Raw)
	if err := finalizationTransformMaintained(f2, residual, false); err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("F2 IMForm-only: %w", err)
	}
	f2.IsMontgomery = false
	result.F2ImformOnly = ptrFinalizationCheckpoint(finalizationCheckpoint("F2_imform_only", "public_IMForm_only_no_scale_change", f2, residual, generatedSecret, zeroSecretKey, input, standardReference))

	f2r := finalizationDiagnosticCopy(f1Raw)
	if err := finalizationTransformMaintained(f2r, residual, false); err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("F2R IMForm: %w", err)
	}
	if err := finalizationTransformMaintained(f2r, residual, true); err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("F2R MForm: %w", err)
	}
	roundTrip := FinalizationRoundTrip{Rows: compareFinalizationRows(f1Raw, f2r)}
	roundTrip.ExactEqual = finalizationRowsEqual(roundTrip.Rows)
	roundTrip.Classification = "MATCH"
	if !roundTrip.ExactEqual {
		roundTrip.Classification = "MONTGOMERY_ROUNDTRIP_FAILURE"
	}
	result.F2RRoundTrip = &roundTrip

	f3 := finalizationDiagnosticCopy(f1Raw)
	if err := finalizationTransformMaintained(f3, residual, false); err != nil {
		return FinalizationDiagnosticResult{}, fmt.Errorf("F3 IMForm: %w", err)
	}
	f3.IsMontgomery = false
	f3.Scale = residual.DefaultScale()
	result.F3FullFinalizationView = ptrFinalizationCheckpoint(finalizationCheckpoint("F3_full_finalization_view", "public_IMForm_plus_residual_default_scale", f3, residual, generatedSecret, zeroSecretKey, input, standardReference))

	f4 := finalizationDiagnosticCopy(officialOutput)
	result.F4OfficialFastOutput = ptrFinalizationCheckpoint(finalizationCheckpoint("F4_official_fast_output", "ordinary_public_bootstrap_output", f4, residual, generatedSecret, zeroSecretKey, input, standardReference))
	if result.F3FullFinalizationView.ZeroVsInput != nil && result.F4OfficialFastOutput.ZeroVsInput != nil {
		if evidence, err := finalizationLogicalEvidence(finalizationValues(result.F4OfficialFastOutput.ZeroSecretDecoded), finalizationValues(result.F3FullFinalizationView.ZeroSecretDecoded)); err == nil {
			result.F3VsF4 = evidence
		}
	}
	result.ScaleEvidence = &FinalizationScaleEvidence{
		S2CScale: finalizationScaleString(f0.Scale), S2CScaleFloat64: f0.Scale.Float64(), ResidualDefaultScale: finalizationScaleString(residual.DefaultScale()), ResidualDefaultScaleFloat64: residual.DefaultScale().Float64(),
		Ratio: f0.Scale.Float64() / residual.DefaultScale().Float64(), F2ScaleChanged: !f2.Scale.Equal(f1Raw.Scale), F3ScaleChanged: !f3.Scale.Equal(f1Raw.Scale),
	}
	result.FirstSupportedCause, result.DiagnosticClassification = classifyFinalizationEvidence(result)
	return result, nil
}

func ptrFinalizationCheckpoint(value FinalizationCheckpoint) *FinalizationCheckpoint {
	return &value
}

func classifyFinalizationEvidence(result FinalizationDiagnosticResult) (string, string) {
	if result.UnpackBoundary == nil || !result.UnpackBoundary.ExactEqual {
		return "unpack_and_switch_n2_to_n1", "UNPACK_BOUNDARY_CHANGED_VALUE"
	}
	if result.F2RRoundTrip == nil || !result.F2RRoundTrip.ExactEqual {
		return "montgomery_representation_boundary", "MONTGOMERY_ROUNDTRIP_FAILURE"
	}
	f2Pass := result.F2ImformOnly != nil && result.F2ImformOnly.ZeroVsInput != nil && result.F2ImformOnly.ZeroVsInput.PassThreshold && result.F2ImformOnly.ZeroVsStandard != nil && result.F2ImformOnly.ZeroVsStandard.PassThreshold
	f3Pass := result.F3FullFinalizationView != nil && result.F3FullFinalizationView.ZeroVsInput != nil && result.F3FullFinalizationView.ZeroVsInput.PassThreshold && result.F3FullFinalizationView.ZeroVsStandard != nil && result.F3FullFinalizationView.ZeroVsStandard.PassThreshold
	f4Pass := result.F4OfficialFastOutput != nil && result.F4OfficialFastOutput.ZeroVsInput != nil && result.F4OfficialFastOutput.ZeroVsInput.PassThreshold && result.F4OfficialFastOutput.ZeroVsStandard != nil && result.F4OfficialFastOutput.ZeroVsStandard.PassThreshold
	if !f2Pass {
		return "at_or_before_slots_to_coeffs", "CASE_C_F2_DIVERGED_ROUNDTRIP_EXACT"
	}
	if f2Pass && !f3Pass {
		return "final_scale_restoration", "CASE_D_F2_PASS_F3_FAIL"
	}
	if f2Pass && f3Pass && !f4Pass {
		return "unresolved_public_path_mismatch", "CASE_E_F2_F3_PASS_F4_FAIL"
	}
	return "unresolved_previous_diagnostic_inconsistency", "CASE_F_ALL_FINALIZATION_VIEWS_PASS"
}

func WriteFinalizationDiagnosticResult(result FinalizationDiagnosticResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode finalization diagnostic result: %w", err)
	}
	data = append(data, '\n')
	if path == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create finalization result directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write finalization diagnostic result: %w", err)
	}
	return nil
}
