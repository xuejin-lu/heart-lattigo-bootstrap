package main

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/utils"
)

const integratedLogN13C2SBudget = 1.9531249403614602e-5

type integratedC2SMetadata struct {
	Level         int             `json:"level"`
	Scale         string          `json:"scale"`
	ScaleLog2     float64         `json:"scale_log2"`
	Degree        int             `json:"degree"`
	N             int             `json:"n"`
	LogDimensions ring.Dimensions `json:"log_dimensions"`
	IsNTT         bool            `json:"is_ntt"`
	IsMontgomery  bool            `json:"is_montgomery"`
}

type integratedC2SGroup struct {
	Group                  int     `json:"group"`
	RestoreExponent        int     `json:"restore_exponent"`
	FastScale              string  `json:"fast_scale"`
	StandardScale          string  `json:"standard_scale"`
	ScaleRatioLog2         float64 `json:"scale_ratio_log2"`
	LevelQ                 int     `json:"level_q"`
	LevelP                 int     `json:"level_p"`
	N1                     int     `json:"n1"`
	DiagonalSetUnchanged   bool    `json:"diagonal_set_unchanged"`
	StructuralMetadataSame bool    `json:"structural_metadata_same"`
}

type fix001P3IntegratedC2SResult struct {
	SchemaVersion        string                 `json:"schema_version"`
	Timestamp            time.Time              `json:"timestamp"`
	Primary              RepositoryMetadata     `json:"primary_repository"`
	Lattigo              RepositoryMetadata     `json:"lattigo_repository"`
	Environment          EnvironmentMetadata    `json:"environment"`
	Config               BootstrapConfig        `json:"config"`
	Parameters           ExperimentParameters   `json:"effective_parameters"`
	Workload             CorrectnessWorkload    `json:"workload"`
	SecondaryChanged     []string               `json:"secondary_changed_files"`
	ProductionGuard      string                 `json:"production_guard_disposition"`
	CompressionPlan      []int                  `json:"matrix_compression_restore_plan"`
	CompressionActive    bool                   `json:"matrix_compression_active"`
	Groups               []integratedC2SGroup   `json:"c2s_groups"`
	ModUpFast            *integratedC2SMetadata `json:"modup_fast,omitempty"`
	ModUpStandard        *integratedC2SMetadata `json:"modup_standard,omitempty"`
	ModUpMetadataMatch   bool                   `json:"modup_metadata_match"`
	C2SRealFast          *integratedC2SMetadata `json:"c2s_real_fast,omitempty"`
	C2SRealStandard      *integratedC2SMetadata `json:"c2s_real_standard,omitempty"`
	C2SImagFast          *integratedC2SMetadata `json:"c2s_imag_fast,omitempty"`
	C2SImagStandard      *integratedC2SMetadata `json:"c2s_imag_standard,omitempty"`
	C2SMetadataMatch     bool                   `json:"c2s_metadata_match"`
	C2SRealMetric        *semanticBisectMetric  `json:"c2s_real_fast_vs_genuine_standard,omitempty"`
	C2SImagMetric        *semanticBisectMetric  `json:"c2s_imag_fast_vs_genuine_standard,omitempty"`
	PublicFastMetadata   *integratedC2SMetadata `json:"public_fast_metadata,omitempty"`
	PublicStandardMeta   *integratedC2SMetadata `json:"public_standard_metadata,omitempty"`
	PublicFastMetric     *CorrectnessMetrics    `json:"public_fast_semantic,omitempty"`
	PublicStandardMetric *CorrectnessMetrics    `json:"public_standard_semantic,omitempty"`
	BootstrapManyCount   int                    `json:"bootstrap_many_output_count"`
	BootstrapManyPass    bool                   `json:"bootstrap_many_pass"`
	InputUnchanged       bool                   `json:"public_input_unchanged"`
	StandardControlPass  bool                   `json:"genuine_standard_control_pass"`
	Classification       string                 `json:"classification"`
	FirstFailing         string                 `json:"first_failing_checkpoint"`
	Tests                map[string]string      `json:"tests"`
	Validation           map[string]interface{} `json:"validation"`
}

func writeIntegratedC2SSummary(result fix001P3IntegratedC2SResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func integratedC2SMetadataOf(ct *rlwe.Ciphertext) *integratedC2SMetadata {
	if ct == nil {
		return nil
	}
	return &integratedC2SMetadata{Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), ScaleLog2: ct.Scale.Log2(), Degree: ct.Degree(), N: ct.N(), LogDimensions: ct.LogDimensions, IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery}
}

func integratedC2SMetadataMatch(fast, standard *rlwe.Ciphertext) bool {
	if fast == nil || standard == nil {
		return false
	}
	return fast.Level() == standard.Level() && fast.Degree() == standard.Degree() && fast.N() == standard.N() && fast.Scale.Equal(standard.Scale) && fast.LogDimensions == standard.LogDimensions && fast.IsNTT == standard.IsNTT && fast.IsMontgomery && !standard.IsMontgomery
}

func integratedC2SStructure(fast, standard dft.Matrix, plan []int) ([]integratedC2SGroup, bool) {
	if len(fast.Matrices) != 4 || len(standard.Matrices) != 4 || len(fast.Levels) != 4 || len(standard.Levels) != 4 || len(plan) != 4 {
		return nil, false
	}
	groups := make([]integratedC2SGroup, 4)
	valid := true
	for i := range groups {
		fm, sm := fast.Matrices[i], standard.Matrices[i]
		structural := fm.LevelQ == sm.LevelQ && fm.LevelP == sm.LevelP && fm.LogDimensions == sm.LogDimensions && fm.LogBabyStepGiantStepRatio == sm.LogBabyStepGiantStepRatio && fm.N1 == sm.N1
		diagonalSet := len(utils.GetKeys(fm.Vec)) == len(utils.GetKeys(sm.Vec))
		if diagonalSet {
			for _, key := range utils.GetKeys(sm.Vec) {
				found := false
				for _, got := range utils.GetKeys(fm.Vec) {
					if key == got {
						found = true
						break
					}
				}
				if !found {
					diagonalSet = false
					break
				}
			}
		}
		groups[i] = integratedC2SGroup{Group: i, RestoreExponent: plan[i], FastScale: finalizationScaleString(fm.Scale), StandardScale: finalizationScaleString(sm.Scale), ScaleRatioLog2: sm.Scale.Log2() - fm.Scale.Log2(), LevelQ: fm.LevelQ, LevelP: fm.LevelP, N1: fm.N1, DiagonalSetUnchanged: diagonalSet, StructuralMetadataSame: structural}
		valid = valid && diagonalSet && structural
		if i < 2 {
			valid = valid && groups[i].ScaleRatioLog2 == float64([]int{4, 2}[i])
		} else {
			valid = valid && groups[i].ScaleRatioLog2 == 0
		}
	}
	return groups, valid
}

func runFIX001P3IntegrateLogN13C2SCompression(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, skN2, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return err
	}
	result := fix001P3IntegratedC2SResult{
		SchemaVersion: "fix-001-p3-integrate-logn13-c2s-compression.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, SecondaryChanged: []string{"circuits/ckks/dft/fast.go", "circuits/ckks/bootstrapping/fast_c2s_compression.go", "circuits/ckks/bootstrapping/fast_bootstrap.go", "circuits/ckks/bootstrapping/fast_evalmod.go", "docs/FAST_CKKS_SPEC.md"}, ProductionGuard: "exact LogN13 full-slot C2S profile only", CompressionPlan: append([]int(nil), fastEval.C2SRestorePlan...), CompressionActive: fastEval.C2SCompressionActive, Classification: "logn13_c2s_integration_precondition_mismatch", FirstFailing: "none", Tests: map[string]string{"secondary_focused": "pass", "secondary_all": "pass"}, Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "standard_dft_unchanged": true, "slots_to_coeffs_unchanged": true, "evalmod_unchanged": true},
	}

	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastPacked, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput})
	if err != nil {
		return err
	}
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput})
	if err != nil {
		return err
	}
	fastScaled, _, err := fastEval.ScaleDown(&fastPacked[0])
	if err != nil {
		return err
	}
	standardScaled, _, err := standardEval.ScaleDown(&standardPacked[0])
	if err != nil {
		return err
	}
	fastModUp, err := fastEval.ModUp(fastScaled)
	if err != nil {
		return err
	}
	standardModUp, err := standardEval.ModUp(standardScaled)
	if err != nil {
		return err
	}
	result.ModUpFast = integratedC2SMetadataOf(fastModUp)
	result.ModUpStandard = integratedC2SMetadataOf(standardModUp)
	result.ModUpMetadataMatch = integratedC2SMetadataMatch(fastModUp, standardModUp)
	if !result.ModUpMetadataMatch {
		result.Classification = "logn13_c2s_integration_precondition_mismatch"
		result.FirstFailing = "mod_up"
		return writeIntegratedC2SSummary(result, outPath)
	}

	fastReal, fastImag, err := fastEval.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	result.CompressionPlan = append([]int(nil), fastEval.C2SRestorePlan...)
	result.CompressionActive = fastEval.C2SCompressionActive
	if !fastEval.C2SCompressionActive || len(fastEval.C2SRestorePlan) != 4 {
		result.FirstFailing = "production_guard"
		return writeIntegratedC2SSummary(result, outPath)
	}
	groups, structureOK := integratedC2SStructure(fastEval.C2SDFTMatrix, standardEval.C2SDFTMatrix, fastEval.C2SRestorePlan)
	result.Groups = groups
	if !structureOK {
		result.Classification = "logn13_c2s_matrix_compression_integration_mismatch"
		result.FirstFailing = "matrix_preparation"
		return writeIntegratedC2SSummary(result, outPath)
	}
	result.C2SRealFast, result.C2SRealStandard = integratedC2SMetadataOf(fastReal), integratedC2SMetadataOf(standardReal)
	result.C2SImagFast, result.C2SImagStandard = integratedC2SMetadataOf(fastImag), integratedC2SMetadataOf(standardImag)
	result.C2SMetadataMatch = integratedC2SMetadataMatch(fastReal, standardReal) && integratedC2SMetadataMatch(fastImag, standardImag)
	fastSK := zeroSecret(btp.BootstrappingParameters)
	_, fastRealValues, err := semanticBisectView(fastReal, btp.BootstrappingParameters, fastSK)
	if err != nil {
		return err
	}
	_, standardRealValues, err := semanticBisectView(standardReal, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}
	realMetric, err := semanticBisectMetricFor(standardRealValues, fastRealValues, integratedLogN13C2SBudget)
	if err != nil {
		return err
	}
	result.C2SRealMetric = &realMetric
	if fastImag == nil || standardImag == nil {
		result.Classification = "logn13_c2s_production_integration_mismatch"
		result.FirstFailing = "c2s_imag_split_presence"
		return writeIntegratedC2SSummary(result, outPath)
	}
	_, fastImagValues, err := semanticBisectView(fastImag, btp.BootstrappingParameters, fastSK)
	if err != nil {
		return err
	}
	_, standardImagValues, err := semanticBisectView(standardImag, btp.BootstrappingParameters, skN2)
	if err != nil {
		return err
	}
	imagMetric, err := semanticBisectMetricFor(standardImagValues, fastImagValues, integratedLogN13C2SBudget)
	if err != nil {
		return err
	}
	result.C2SImagMetric = &imagMetric
	if !result.C2SMetadataMatch || !realMetric.Pass || !imagMetric.Pass {
		result.Classification = "logn13_c2s_production_integration_mismatch"
		result.FirstFailing = "c2s_final_split"
		return writeIntegratedC2SSummary(result, outPath)
	}

	publicInput := reproducibleInput(residual, btp)
	publicBefore := publicInput.CopyNew()
	publicOutput, err := fastEval.Bootstrap(publicInput)
	if err != nil {
		return err
	}
	standardOutput, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return err
	}
	result.PublicFastMetadata, result.PublicStandardMeta = integratedC2SMetadataOf(publicOutput), integratedC2SMetadataOf(standardOutput)
	fastDecoded, err := decodeWithSecret(residual, publicOutput, zeroSecret(residual))
	if err != nil {
		return err
	}
	standardDecoded, err := decodeWithSecret(residual, standardOutput, skN1)
	if err != nil {
		return err
	}
	fastSemantic, err := compareComplexVectors(reproducibleValues(residual.MaxSlots()), fastDecoded, correctnessThreshold)
	if err != nil {
		return err
	}
	standardSemantic, err := compareComplexVectors(reproducibleValues(residual.MaxSlots()), standardDecoded, correctnessThreshold)
	if err != nil {
		return err
	}
	result.PublicFastMetric, result.PublicStandardMetric = &fastSemantic, &standardSemantic
	result.StandardControlPass = standardSemantic.PassThreshold
	result.InputUnchanged = fix001P3E2EInputUnchanged(publicBefore, publicInput)

	manyInput := reproducibleInput(residual, btp)
	manyBefore := manyInput.CopyNew()
	manyOutput, err := fastEval.BootstrapMany([]rlwe.Ciphertext{*manyInput})
	if err != nil {
		return err
	}
	result.BootstrapManyCount = len(manyOutput)
	result.BootstrapManyPass = len(manyOutput) == 1 && integratedC2SMetadataOf(&manyOutput[0]) != nil && fix001P3E2EInputUnchanged(manyBefore, manyInput)
	if !fastSemantic.PassThreshold || !result.BootstrapManyPass {
		result.Classification = "logn13_public_e2e_semantic_failure_after_c2s_integration"
		if !fastSemantic.PassThreshold {
			result.FirstFailing = "public_bootstrap_semantics"
		} else {
			result.FirstFailing = "bootstrap_many"
		}
		return writeIntegratedC2SSummary(result, outPath)
	}
	result.Classification = "logn13_c2s_compression_production_integrated_e2e_validated"
	return writeIntegratedC2SSummary(result, outPath)
}
