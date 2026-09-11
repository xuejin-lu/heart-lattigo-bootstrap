package main

import (
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

// Q01CiphertextMetadata records only the maintained q0/q1 rows. The helper
// intentionally never indexes a source row at limb >= 2.
type Q01CiphertextMetadata struct {
	Level        int                          `json:"level"`
	Scale        string                       `json:"scale"`
	ScaleMod     string                       `json:"scale_mod,omitempty"`
	ScaleFloat64 float64                      `json:"scale_float64"`
	ScaleLog2    float64                      `json:"scale_log2"`
	N            int                          `json:"n"`
	LogSlots     int                          `json:"log_slots"`
	IsNTT        bool                         `json:"is_ntt"`
	IsMontgomery bool                         `json:"is_montgomery"`
	QModuli      []uint64                     `json:"q_moduli"`
	Rows         []FinalizationRowFingerprint `json:"maintained_rows_q0_q1"`
}

type Q01ProjectionEvidence struct {
	Source                    Q01CiphertextMetadata `json:"source"`
	SourceAfterProjection     Q01CiphertextMetadata `json:"source_after_projection"`
	DiagnosticBeforeNormalize Q01CiphertextMetadata `json:"diagnostic_before_normalize"`
	DiagnosticAfterNormalize  Q01CiphertextMetadata `json:"diagnostic_after_normalize"`
	RowsCopiedExactly         bool                  `json:"rows_copied_exactly"`
	SourceUnchanged           bool                  `json:"source_unchanged"`
}

type Q01StageBranchResult struct {
	Branch               string                 `json:"branch"`
	Present              bool                   `json:"present"`
	StandardSource       *Q01ProjectionEvidence `json:"standard_source,omitempty"`
	StandardFullDecoded  []CorrectnessValue     `json:"standard_full_decoded,omitempty"`
	StandardQ01Decoded   []CorrectnessValue     `json:"standard_q01_decoded,omitempty"`
	StandardFullVsQ01    *CorrectnessMetrics    `json:"standard_full_vs_q01,omitempty"`
	StandardOracleValid  bool                   `json:"standard_oracle_valid"`
	FastSource           *Q01ProjectionEvidence `json:"fast_source,omitempty"`
	FastQ01Decoded       []CorrectnessValue     `json:"fast_q01_decoded,omitempty"`
	StandardQ01VsFastQ01 *CorrectnessMetrics    `json:"standard_q01_vs_fast_q01,omitempty"`
	Errors               []string               `json:"errors,omitempty"`
}

type Q01StageResult struct {
	Name        string                 `json:"name"`
	Order       int                    `json:"order"`
	Branches    []Q01StageBranchResult `json:"branches"`
	OracleValid bool                   `json:"oracle_valid"`
}

type Q01F2Consistency struct {
	Stage                 string              `json:"stage"`
	Method                string              `json:"method"`
	FastQ01MatchesF2View  bool                `json:"fast_q01_matches_f2_view"`
	FastQ01VsF2View       *CorrectnessMetrics `json:"fast_q01_vs_f2_view,omitempty"`
	ExistingF2ViewVsInput *CorrectnessMetrics `json:"existing_f2_view_vs_input,omitempty"`
	ExistingF2Diverges    bool                `json:"existing_f2_diverges"`
}

type Q01DiagnosticResult struct {
	SchemaVersion         string               `json:"schema_version"`
	Timestamp             time.Time            `json:"timestamp"`
	BackendRole           string               `json:"backend_role"`
	Primary               RepositoryMetadata   `json:"primary_repository"`
	Lattigo               RepositoryMetadata   `json:"lattigo_repository"`
	Environment           EnvironmentMetadata  `json:"environment"`
	Config                BootstrapConfig      `json:"config"`
	Parameters            ExperimentParameters `json:"effective_parameters"`
	Workload              CorrectnessWorkload  `json:"workload"`
	Input                 []CorrectnessValue   `json:"input"`
	Q01Moduli             []uint64             `json:"q01_moduli"`
	StandardReferencePath string               `json:"standard_reference_path,omitempty"`
	Stages                []Q01StageResult     `json:"stages"`
	F2Consistency         *Q01F2Consistency    `json:"f2_consistency,omitempty"`
	FirstSupportedCause   string               `json:"first_supported_cause"`
	Classification        string               `json:"diagnostic_classification"`
	Notes                 []string             `json:"notes,omitempty"`
}

type q01StageSnapshot struct {
	Name     string
	Order    int
	Branches []q01BranchSnapshot
}

type q01BranchSnapshot struct {
	Name string
	CT   *rlwe.Ciphertext
}

func q01RowsEqual(a, b []FinalizationRowFingerprint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func q01CiphertextMetadataEqual(a, b Q01CiphertextMetadata) bool {
	return a.Level == b.Level && a.Scale == b.Scale && a.ScaleMod == b.ScaleMod && a.ScaleFloat64 == b.ScaleFloat64 && a.ScaleLog2 == b.ScaleLog2 && a.N == b.N && a.LogSlots == b.LogSlots && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && len(a.QModuli) == len(b.QModuli) && q01RowsEqual(a.Rows, b.Rows)
}

func q01Moduli(params ckks.Parameters) ([]uint64, error) {
	if len(params.RingQ().SubRings) < 2 {
		return nil, fmt.Errorf("q0/q1 projection requires at least two Q moduli")
	}
	return []uint64{params.RingQ().SubRings[0].Modulus, params.RingQ().SubRings[1].Modulus}, nil
}

func q01Metadata(params ckks.Parameters, ct *rlwe.Ciphertext) (Q01CiphertextMetadata, error) {
	if ct == nil {
		return Q01CiphertextMetadata{}, fmt.Errorf("cannot record nil ciphertext")
	}
	moduli, err := q01Moduli(params)
	if err != nil {
		return Q01CiphertextMetadata{}, err
	}
	rows := make([]FinalizationRowFingerprint, 0, 2*(ct.Degree()+1))
	for component := 0; component <= ct.Degree() && component < len(ct.Value); component++ {
		for limb := 0; limb < 2; limb++ {
			if limb >= len(ct.Value[component].Coeffs) {
				return Q01CiphertextMetadata{}, fmt.Errorf("component %d has no q%d row", component, limb)
			}
			row := ct.Value[component].Coeffs[limb]
			if len(row) == 0 {
				return Q01CiphertextMetadata{}, fmt.Errorf("component %d q%d row is empty", component, limb)
			}
			rows = append(rows, FinalizationRowFingerprint{Component: component, Limb: limb, Length: len(row), SHA256: finalizationRowHash(row)})
		}
	}
	return Q01CiphertextMetadata{
		Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), ScaleMod: finalizationScaleModString(ct.Scale), ScaleFloat64: ct.Scale.Float64(), ScaleLog2: ct.Scale.Log2(),
		N: ct.N(), LogSlots: ct.LogSlots(), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, QModuli: moduli, Rows: rows,
	}, nil
}

// q01Projection creates the isolated level-1 oracle copy. It reads and copies
// exactly q0/q1, preserves all public metadata, and optionally normalizes
// Montgomery residues only on the copy.
func q01Projection(params ckks.Parameters, source *rlwe.Ciphertext, normalizeMontgomery bool) (*rlwe.Ciphertext, Q01ProjectionEvidence, error) {
	if source == nil {
		return nil, Q01ProjectionEvidence{}, fmt.Errorf("cannot project nil ciphertext")
	}
	if source.Level() < 1 {
		return nil, Q01ProjectionEvidence{}, fmt.Errorf("source logical level is %d; q0/q1 projection requires level >= 1", source.Level())
	}
	if normalizeMontgomery && !source.IsMontgomery {
		return nil, Q01ProjectionEvidence{}, fmt.Errorf("requested Fast Montgomery normalization for non-Montgomery source")
	}
	sourceMetadata, err := q01Metadata(params, source)
	if err != nil {
		return nil, Q01ProjectionEvidence{}, err
	}
	projection := ckks.NewCiphertext(params, source.Degree(), 1)
	projection.MetaData = source.MetaData.CopyNew()
	for component := 0; component <= source.Degree(); component++ {
		for limb := 0; limb < 2; limb++ {
			copy(projection.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	diagnosticBefore, err := q01Metadata(params, projection)
	if err != nil {
		return nil, Q01ProjectionEvidence{}, err
	}
	if !q01RowsEqual(sourceMetadata.Rows, diagnosticBefore.Rows) {
		return nil, Q01ProjectionEvidence{}, fmt.Errorf("q0/q1 rows were not copied exactly")
	}
	if normalizeMontgomery {
		for component := 0; component <= projection.Degree(); component++ {
			for limb := 0; limb < 2; limb++ {
				params.RingQ().SubRings[limb].IMForm(projection.Value[component].Coeffs[limb], projection.Value[component].Coeffs[limb])
			}
		}
		projection.IsMontgomery = false
	}
	diagnosticAfter, err := q01Metadata(params, projection)
	if err != nil {
		return nil, Q01ProjectionEvidence{}, err
	}
	sourceAfter, err := q01Metadata(params, source)
	if err != nil {
		return nil, Q01ProjectionEvidence{}, err
	}
	evidence := Q01ProjectionEvidence{
		Source: sourceMetadata, SourceAfterProjection: sourceAfter, DiagnosticBeforeNormalize: diagnosticBefore, DiagnosticAfterNormalize: diagnosticAfter,
		RowsCopiedExactly: q01RowsEqual(sourceMetadata.Rows, diagnosticBefore.Rows), SourceUnchanged: q01CiphertextMetadataEqual(sourceMetadata, sourceAfter),
	}
	return projection, evidence, nil
}

func q01RunStageSnapshots(eval *bootstrapping.Evaluator, residual ckks.Parameters, btp bootstrapping.Parameters) ([]q01StageSnapshot, []complex128, error) {
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
	snapshots := []q01StageSnapshot{{Name: "Q1_mod_up", Order: 1, Branches: []q01BranchSnapshot{{Name: "single", CT: finalizationDiagnosticCopy(modUp)}}}}
	ctReal, ctImag, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return nil, nil, fmt.Errorf("CoeffsToSlots: %w", err)
	}
	branches := []q01BranchSnapshot{{Name: "real", CT: finalizationDiagnosticCopy(ctReal)}}
	if ctImag != nil {
		branches = append(branches, q01BranchSnapshot{Name: "imag", CT: finalizationDiagnosticCopy(ctImag)})
	}
	snapshots = append(snapshots, q01StageSnapshot{Name: "Q2_coeffs_to_slots", Order: 2, Branches: branches})
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return nil, nil, fmt.Errorf("EvalMod real: %w", err)
	}
	snapshots = append(snapshots, q01StageSnapshot{Name: "Q3_eval_mod_real", Order: 3, Branches: []q01BranchSnapshot{{Name: "real", CT: finalizationDiagnosticCopy(ctReal)}}})
	if ctImag != nil {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return nil, nil, fmt.Errorf("EvalMod imag: %w", err)
		}
		snapshots = append(snapshots, q01StageSnapshot{Name: "Q4_eval_mod_imag", Order: 4, Branches: []q01BranchSnapshot{{Name: "imag", CT: finalizationDiagnosticCopy(ctImag)}}})
	}
	ctOut, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return nil, nil, fmt.Errorf("SlotsToCoeffs: %w", err)
	}
	snapshots = append(snapshots, q01StageSnapshot{Name: "Q5_slots_to_coeffs", Order: 5, Branches: []q01BranchSnapshot{{Name: "combined", CT: finalizationDiagnosticCopy(ctOut)}}})
	if _, err := eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*finalizationDiagnosticCopy(ctOut)}, ctxtN1, ctxtN2); err != nil {
		return nil, nil, fmt.Errorf("UnpackAndSwitchN2ToN1: %w", err)
	}
	return snapshots, reproducibleValues(residual.MaxSlots()), nil
}

func q01StandardBranch(stage q01StageSnapshot, branch q01BranchSnapshot, params ckks.Parameters, bootstrapSecret *rlwe.SecretKey) (Q01StageBranchResult, error) {
	full, err := diagnosticDecode(params, branch.CT, bootstrapSecret)
	if err != nil {
		return Q01StageBranchResult{}, fmt.Errorf("%s/%s Standard full decode: %w", stage.Name, branch.Name, err)
	}
	projection, evidence, err := q01Projection(params, branch.CT, false)
	if err != nil {
		return Q01StageBranchResult{}, fmt.Errorf("%s/%s Standard projection: %w", stage.Name, branch.Name, err)
	}
	projected, err := diagnosticDecode(params, projection, bootstrapSecret)
	if err != nil {
		return Q01StageBranchResult{}, fmt.Errorf("%s/%s Standard q01 decode: %w", stage.Name, branch.Name, err)
	}
	metrics, err := compareComplexVectors(full, projected, correctnessThreshold)
	if err != nil {
		return Q01StageBranchResult{}, err
	}
	return Q01StageBranchResult{Branch: branch.Name, Present: true, StandardSource: &evidence, StandardFullDecoded: correctnessValues(full), StandardQ01Decoded: correctnessValues(projected), StandardFullVsQ01: &metrics, StandardOracleValid: metrics.PassThreshold}, nil
}

func q01FastBranch(stage q01StageSnapshot, branch q01BranchSnapshot, params ckks.Parameters, zeroBootstrapSecret *rlwe.SecretKey, reference Q01StageBranchResult) (Q01StageBranchResult, error) {
	if !reference.Present || !reference.StandardOracleValid || len(reference.StandardQ01Decoded) == 0 {
		return reference, nil
	}
	if !branch.CT.IsMontgomery {
		return Q01StageBranchResult{}, fmt.Errorf("%s/%s Fast source is not Montgomery", stage.Name, branch.Name)
	}
	projection, evidence, err := q01Projection(params, branch.CT, true)
	if err != nil {
		return Q01StageBranchResult{}, fmt.Errorf("%s/%s Fast projection: %w", stage.Name, branch.Name, err)
	}
	projected, err := diagnosticDecode(params, projection, zeroBootstrapSecret)
	if err != nil {
		return Q01StageBranchResult{}, fmt.Errorf("%s/%s Fast q01 decode: %w", stage.Name, branch.Name, err)
	}
	standard := finalizationValues(reference.StandardQ01Decoded)
	metrics, err := compareComplexVectors(standard, projected, correctnessThreshold)
	if err != nil {
		return Q01StageBranchResult{}, err
	}
	result := reference
	result.FastSource = &evidence
	result.FastQ01Decoded = correctnessValues(projected)
	result.StandardQ01VsFastQ01 = &metrics
	return result, nil
}

func q01ReferenceStage(reference Q01DiagnosticResult, name string) (Q01StageResult, error) {
	for _, stage := range reference.Stages {
		if stage.Name == name {
			return stage, nil
		}
	}
	return Q01StageResult{}, fmt.Errorf("Standard reference has no stage %q", name)
}

func q01ReferenceBranch(stage Q01StageResult, name string) (Q01StageBranchResult, error) {
	for _, branch := range stage.Branches {
		if branch.Branch == name {
			return branch, nil
		}
	}
	return Q01StageBranchResult{}, fmt.Errorf("Standard reference stage %q has no branch %q", stage.Name, name)
}

func q01Classify(result Q01DiagnosticResult) (string, string) {
	for _, stage := range result.Stages {
		for _, branch := range stage.Branches {
			if !branch.Present {
				continue
			}
			if !branch.StandardOracleValid {
				return fmt.Sprintf("%s/%s", stage.Name, branch.Branch), "unresolved_projection_oracle"
			}
			if branch.StandardQ01VsFastQ01 == nil {
				return fmt.Sprintf("%s/%s", stage.Name, branch.Branch), "unresolved_projection_oracle"
			}
			if !branch.StandardQ01VsFastQ01.PassThreshold {
				switch stage.Name {
				case "Q1_mod_up":
					return fmt.Sprintf("%s/%s", stage.Name, branch.Branch), "mod_up"
				case "Q2_coeffs_to_slots":
					return fmt.Sprintf("%s/%s", stage.Name, branch.Branch), "coeffs_to_slots"
				case "Q3_eval_mod_real", "Q4_eval_mod_imag":
					return fmt.Sprintf("%s/%s", stage.Name, branch.Branch), "eval_mod"
				case "Q5_slots_to_coeffs":
					return fmt.Sprintf("%s/%s", stage.Name, branch.Branch), "slots_to_coeffs"
				}
			}
		}
	}
	if result.F2Consistency != nil && result.F2Consistency.ExistingF2Diverges {
		return "none", "diagnostic_inconsistency"
	}
	return "none", "unresolved_projection_oracle"
}

func q01BuildBaseResult(cfg BootstrapConfig, primaryRoot, backendRoot, role string, residual ckks.Parameters, btp bootstrapping.Parameters, input []complex128, moduli []uint64) Q01DiagnosticResult {
	return Q01DiagnosticResult{
		SchemaVersion: "exp-002c-diag-q01.v1", Timestamp: time.Now().UTC(), BackendRole: role, Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp),
		Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: len(input)}, Input: correctnessValues(input), Q01Moduli: moduli,
	}
}

func RunQ01DiagnosticExperiment(cfg BootstrapConfig, primaryRoot, backendRoot, standardReferencePath string) (Q01DiagnosticResult, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return Q01DiagnosticResult{}, err
	}
	moduli, err := q01Moduli(btp.BootstrappingParameters)
	if err != nil {
		return Q01DiagnosticResult{}, err
	}
	generatedSecret := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	evalKeys, bootstrapSecret, err := btp.GenEvaluationKeys(generatedSecret)
	if err != nil {
		return Q01DiagnosticResult{}, fmt.Errorf("generate evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(btp, evalKeys)
	if err != nil {
		return Q01DiagnosticResult{}, fmt.Errorf("construct bootstrap evaluator: %w", err)
	}
	snapshots, input, err := q01RunStageSnapshots(eval, residual, btp)
	if err != nil {
		return Q01DiagnosticResult{}, err
	}
	result := q01BuildBaseResult(cfg, primaryRoot, backendRoot, "standard", residual, btp, input, moduli)
	if standardReferencePath == "" {
		for _, snapshot := range snapshots {
			stageResult := Q01StageResult{Name: snapshot.Name, Order: snapshot.Order, OracleValid: true}
			for _, branch := range snapshot.Branches {
				captured, err := q01StandardBranch(snapshot, branch, btp.BootstrappingParameters, bootstrapSecret)
				if err != nil {
					return Q01DiagnosticResult{}, err
				}
				stageResult.Branches = append(stageResult.Branches, captured)
				stageResult.OracleValid = stageResult.OracleValid && captured.StandardOracleValid
			}
			result.Stages = append(result.Stages, stageResult)
		}
		result.FirstSupportedCause = "not_applicable_for_this_backend"
		result.Classification = "STANDARD_ORACLE_ONLY"
		return result, nil
	}

	data, err := os.ReadFile(standardReferencePath)
	if err != nil {
		return Q01DiagnosticResult{}, fmt.Errorf("read Standard Q01 reference: %w", err)
	}
	var reference Q01DiagnosticResult
	if err := json.Unmarshal(data, &reference); err != nil {
		return Q01DiagnosticResult{}, fmt.Errorf("decode Standard Q01 reference: %w", err)
	}
	if len(reference.Q01Moduli) != len(moduli) || reference.Q01Moduli[0] != moduli[0] || reference.Q01Moduli[1] != moduli[1] {
		return Q01DiagnosticResult{}, fmt.Errorf("Q01_PARAMETER_MISMATCH: Standard=%v Fast=%v", reference.Q01Moduli, moduli)
	}
	result.BackendRole = "fast"
	result.StandardReferencePath = standardReferencePath
	zeroBootstrapSecret := zeroSecret(btp.BootstrappingParameters)
	for _, snapshot := range snapshots {
		standardStage, err := q01ReferenceStage(reference, snapshot.Name)
		if err != nil {
			return Q01DiagnosticResult{}, err
		}
		stageResult := Q01StageResult{Name: snapshot.Name, Order: snapshot.Order, OracleValid: standardStage.OracleValid}
		for _, branch := range snapshot.Branches {
			standardBranch, err := q01ReferenceBranch(standardStage, branch.Name)
			if err != nil {
				return Q01DiagnosticResult{}, err
			}
			captured, err := q01FastBranch(snapshot, branch, btp.BootstrappingParameters, zeroBootstrapSecret, standardBranch)
			if err != nil {
				return Q01DiagnosticResult{}, err
			}
			stageResult.Branches = append(stageResult.Branches, captured)
		}
		result.Stages = append(result.Stages, stageResult)
	}
	q5, err := q01ReferenceStage(result, "Q5_slots_to_coeffs")
	if err == nil && len(q5.Branches) == 1 && len(q5.Branches[0].FastQ01Decoded) == len(input) {
		q5Branch := q5.Branches[0]
		inputMetrics, metricsErr := compareComplexVectors(input, finalizationValues(q5Branch.FastQ01Decoded), correctnessThreshold)
		if metricsErr != nil {
			return Q01DiagnosticResult{}, metricsErr
		}
		f2Consistency := &Q01F2Consistency{Stage: "Q5_slots_to_coeffs", Method: "Fast q01 projection is the IMForm-only F2 view; source remains untouched", FastQ01MatchesF2View: true, FastQ01VsF2View: &CorrectnessMetrics{PassThreshold: true, Threshold: correctnessThreshold}, ExistingF2ViewVsInput: &inputMetrics, ExistingF2Diverges: !inputMetrics.PassThreshold}
		result.F2Consistency = f2Consistency
	}
	result.FirstSupportedCause, result.Classification = q01Classify(result)
	return result, nil
}

func WriteQ01DiagnosticResult(result Q01DiagnosticResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Q01 diagnostic result: %w", err)
	}
	data = append(data, '\n')
	if path == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create Q01 result directory: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
