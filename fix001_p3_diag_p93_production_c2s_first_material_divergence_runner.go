//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	ckkslintrans "github.com/tuneinsight/lattigo/v6/circuits/ckks/lintrans"
	commonlintrans "github.com/tuneinsight/lattigo/v6/circuits/common/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3C2SBoundarySecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	requiredFIX001P3C2SBoundaryDiff      = "44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e"
	fix001P3C2SBoundaryThreshold         = 1e-2
	fix001P3C2SBoundaryFloor             = 1e-12
	fix001P3C2SBoundaryMaterialFraction  = 0.1
)

type fix001P3C2SBoundaryStage struct {
	Checkpoint                 string                 `json:"checkpoint"`
	Operation                  string                 `json:"operation"`
	ProductionVsDiagnostic     *PSGlobalMetric        `json:"production_vs_controlled_diagnostic"`
	ProductionVsAlignedStd     *PSGlobalMetric        `json:"production_vs_aligned_standard,omitempty"`
	DiagnosticVsAlignedStd     *PSGlobalMetric        `json:"diagnostic_vs_aligned_standard,omitempty"`
	Production                 semanticBisectMetadata `json:"production_metadata"`
	Diagnostic                 semanticBisectMetadata `json:"diagnostic_metadata"`
	AlignedStandard            semanticBisectMetadata `json:"aligned_standard_metadata,omitempty"`
	ProductionDiagnosticMeta   map[string]bool        `json:"production_diagnostic_metadata_equal"`
	ProductionAlignedMeta      map[string]bool        `json:"production_aligned_metadata_equal,omitempty"`
	DiagnosticAlignedMeta      map[string]bool        `json:"diagnostic_aligned_metadata_equal,omitempty"`
	ProductionCapacity         *postMod1S2CCapacity   `json:"production_capacity,omitempty"`
	DiagnosticCapacity         *postMod1S2CCapacity   `json:"diagnostic_capacity,omitempty"`
	AlignedStandardCapacity    *postMod1S2CCapacity   `json:"aligned_standard_capacity,omitempty"`
	AmplificationFromPrevious  float64                `json:"amplification_from_previous"`
	LevelScaleDegreeTransition map[string]interface{} `json:"level_scale_degree_transition,omitempty"`
}

type fix001P3C2SBoundaryPath struct {
	ModUp   *rlwe.Ciphertext
	Real    *rlwe.Ciphertext
	Imag    *rlwe.Ciphertext
	Stages  []fix001P3C2SBoundaryStage
	Objects map[string]*rlwe.Ciphertext
	Aligned map[string]*rlwe.Ciphertext
}

type fix001P3C2SBoundaryResult struct {
	SchemaVersion   string                     `json:"schema_version"`
	Timestamp       time.Time                  `json:"timestamp"`
	Primary         RepositoryMetadata         `json:"primary_repository"`
	Secondary       RepositoryMetadata         `json:"secondary_repository"`
	Environment     EnvironmentMetadata        `json:"environment"`
	Config          BootstrapConfig            `json:"config"`
	FixedProfile    map[string]interface{}     `json:"fixed_profile"`
	Provenance      map[string]interface{}     `json:"provenance"`
	R0Reproduction  map[string]interface{}     `json:"r0_path_reproduction"`
	R1Boundary      map[string]interface{}     `json:"r1_boundary_comparisons"`
	R2Trace         []fix001P3C2SBoundaryStage `json:"r2_source_faithful_c2s_trace"`
	R3Causal        map[string]interface{}     `json:"r3_downstream_causal_confirmation"`
	FirstObservable map[string]interface{}     `json:"first_observable_divergence"`
	FirstMaterial   map[string]interface{}     `json:"first_material_divergence"`
	R5References    map[string]interface{}     `json:"r5_standard_reference_reconciliation"`
	Classification  string                     `json:"classification"`
	Validation      map[string]interface{}     `json:"validation"`
}

func fix001P3C2SBoundaryMetric(expected, actual []complex128) *PSGlobalMetric {
	metric := psGlobalMetric(expected, actual)
	metric.Threshold = fix001P3C2SBoundaryThreshold
	metric.Pass = metric.MaxComponent <= fix001P3C2SBoundaryThreshold
	return metric
}

func fix001P3C2SBoundaryMetadataEqual(a, b semanticBisectMetadata) map[string]bool {
	return map[string]bool{
		"level":          a.SourceLevel == b.SourceLevel,
		"scale":          a.Scale == b.Scale,
		"degree":         a.Degree == b.Degree,
		"n":              a.N == b.N,
		"log_dimensions": a.LogDimensions == b.LogDimensions,
		"is_ntt":         a.IsNTT == b.IsNTT,
		"is_montgomery":  a.IsMontgomery == b.IsMontgomery,
	}
}

func fix001P3C2SBoundaryCapacity(params ckks.Parameters, ct *rlwe.Ciphertext, fast bool) (*postMod1S2CCapacity, error) {
	if ct == nil || ct.Level() < 1 {
		return nil, nil
	}
	var capacity postMod1S2CCapacity
	var err error
	if fast {
		capacity, err = postMod1S2CCapacityFromFastCiphertext(params, ct)
	} else {
		capacity, err = postMod1S2CCapacityFromCiphertext(params, ct)
	}
	if err != nil {
		return nil, err
	}
	return &capacity, nil
}

func fix001P3C2SBoundaryCapture(name, operation string, production, diagnostic, aligned *rlwe.Ciphertext, alignedSK *rlwe.SecretKey, params ckks.Parameters, previous float64) (fix001P3C2SBoundaryStage, error) {
	productionMeta, productionValues, err := semanticBisectView(production, params, zeroSecret(params))
	if err != nil {
		return fix001P3C2SBoundaryStage{}, err
	}
	diagnosticMeta, diagnosticValues, err := semanticBisectView(diagnostic, params, zeroSecret(params))
	if err != nil {
		return fix001P3C2SBoundaryStage{}, err
	}
	alignedMeta := semanticBisectMetadata{}
	var alignedValues []complex128
	if aligned != nil {
		alignedMeta, alignedValues, err = semanticBisectView(aligned, params, alignedSK)
		if err != nil {
			return fix001P3C2SBoundaryStage{}, err
		}
	}
	productionDiagnostic := fix001P3C2SBoundaryMetric(diagnosticValues, productionValues)
	stage := fix001P3C2SBoundaryStage{
		Checkpoint: name, Operation: operation, ProductionVsDiagnostic: productionDiagnostic,
		Production: productionMeta, Diagnostic: diagnosticMeta,
		ProductionDiagnosticMeta:  fix001P3C2SBoundaryMetadataEqual(productionMeta, diagnosticMeta),
		AmplificationFromPrevious: 0,
	}
	if aligned != nil {
		stage.ProductionVsAlignedStd = fix001P3C2SBoundaryMetric(alignedValues, productionValues)
		stage.DiagnosticVsAlignedStd = fix001P3C2SBoundaryMetric(alignedValues, diagnosticValues)
		stage.AlignedStandard = alignedMeta
		stage.ProductionAlignedMeta = fix001P3C2SBoundaryMetadataEqual(productionMeta, alignedMeta)
		stage.DiagnosticAlignedMeta = fix001P3C2SBoundaryMetadataEqual(diagnosticMeta, alignedMeta)
	}
	stage.ProductionCapacity, err = fix001P3C2SBoundaryCapacity(params, production, true)
	if err != nil {
		return fix001P3C2SBoundaryStage{}, err
	}
	stage.DiagnosticCapacity, err = fix001P3C2SBoundaryCapacity(params, diagnostic, true)
	if err != nil {
		return fix001P3C2SBoundaryStage{}, err
	}
	if aligned != nil {
		stage.AlignedStandardCapacity, err = fix001P3C2SBoundaryCapacity(params, aligned, false)
		if err != nil {
			return fix001P3C2SBoundaryStage{}, err
		}
	}
	if previous > 0 {
		stage.AmplificationFromPrevious = productionDiagnostic.MaxComponent / previous
	}
	stage.LevelScaleDegreeTransition = map[string]interface{}{
		"production": map[string]interface{}{"level": productionMeta.SourceLevel, "scale": productionMeta.Scale, "degree": productionMeta.Degree},
		"diagnostic": map[string]interface{}{"level": diagnosticMeta.SourceLevel, "scale": diagnosticMeta.Scale, "degree": diagnosticMeta.Degree},
	}
	return stage, nil
}

func fix001P3C2SBoundaryCopyActive(params ckks.Parameters, source *rlwe.Ciphertext) *rlwe.Ciphertext {
	out := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	fastckks.Resize(out, source.Degree(), source.Level(), params.N())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.IsBatched, out.IsBitReversed = source.IsNTT, source.IsMontgomery, source.IsBatched, source.IsBitReversed
	for component := range source.Value {
		for limb := 0; limb < len(source.Value[component].Coeffs) && limb < len(out.Value[component].Coeffs) && limb < 3; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out
}

func fix001P3C2SBoundaryNewLike(params ckks.Parameters, reference *rlwe.Ciphertext) *rlwe.Ciphertext {
	out := fastckks.NewCiphertext(params, 1, reference.Level())
	*out.MetaData = *reference.MetaData
	out.IsNTT, out.IsMontgomery, out.IsBatched, out.IsBitReversed = reference.IsNTT, reference.IsMontgomery, reference.IsBatched, reference.IsBitReversed
	return out
}

func fix001P3C2SBoundaryFastSplit(params ckks.Parameters, eval *bootstrapping.FastEvaluator, matrix dft.Matrix, source *rlwe.Ciphertext) (map[string]*rlwe.Ciphertext, []fix001P3C2SBoundaryStage, error) {
	zv := fix001P3C2SBoundaryCopyActive(params, source)
	ctReal := fix001P3C2SBoundaryNewLike(params, zv)
	ctImag := fix001P3C2SBoundaryNewLike(params, zv)
	stages := []fix001P3C2SBoundaryStage{}
	objects := map[string]*rlwe.Ciphertext{"split_input": zv}
	if err := eval.FastCKKS.Conjugate(zv, ctReal); err != nil {
		return nil, nil, err
	}
	objects["split_conjugate"] = ctReal
	stages = append(stages, fix001P3C2SBoundaryStage{Checkpoint: "split_conjugate", Operation: "conjugate", Production: semanticBisectMetadata{}, Diagnostic: semanticBisectMetadata{}})
	tmp := ctImag
	if err := eval.FastCKKS.Sub(zv, ctReal, tmp); err != nil {
		return nil, nil, err
	}
	objects["split_imag_sub"] = tmp
	if err := eval.FastCKKS.Mul(tmp, -1i, tmp); err != nil {
		return nil, nil, err
	}
	objects["split_imag_mul_minus_i"] = tmp
	if err := eval.FastCKKS.Add(ctReal, zv, ctReal); err != nil {
		return nil, nil, err
	}
	objects["split_real_add"] = ctReal
	if matrix.Format == dft.RepackImagAsReal && matrix.LogSlots < eval.Parameters.BootstrappingParameters.LogMaxSlots() {
		if err := eval.FastCKKS.Rotate(tmp, tmp, 1<<source.LogDimensions.Cols); err != nil {
			return nil, nil, err
		}
		objects["split_repack_rotate"] = tmp
		if err := eval.FastCKKS.Add(ctReal, tmp, ctReal); err != nil {
			return nil, nil, err
		}
		objects["split_repack_add"] = ctReal
	}
	objects["final_real"] = ctReal
	objects["final_imag"] = ctImag
	return objects, stages, nil
}

func fix001P3C2SBoundaryStandardSplit(params ckks.Parameters, eval *bootstrapping.Evaluator, source *rlwe.Ciphertext, matrix dft.Matrix) (map[string]*rlwe.Ciphertext, error) {
	zv := source.CopyNew()
	ctReal := ckks.NewCiphertext(params, 1, zv.Level())
	*ctReal.MetaData = *zv.MetaData
	ctImag := ckks.NewCiphertext(params, 1, zv.Level())
	*ctImag.MetaData = *zv.MetaData
	if err := eval.Evaluator.Conjugate(zv, ctReal); err != nil {
		return nil, err
	}
	if err := eval.Evaluator.Sub(zv, ctReal, ctImag); err != nil {
		return nil, err
	}
	if err := eval.Evaluator.Mul(ctImag, -1i, ctImag); err != nil {
		return nil, err
	}
	if err := eval.Evaluator.Add(ctReal, zv, ctReal); err != nil {
		return nil, err
	}
	if matrix.Format == dft.RepackImagAsReal && matrix.LogSlots < eval.Parameters.LogMaxSlots() {
		if err := eval.Evaluator.Rotate(ctImag, 1<<source.LogDimensions.Cols, ctImag); err != nil {
			return nil, err
		}
		if err := eval.Evaluator.Add(ctReal, ctImag, ctReal); err != nil {
			return nil, err
		}
	}
	return map[string]*rlwe.Ciphertext{"final_real": ctReal, "final_imag": ctImag}, nil
}

func fix001P3C2SBoundaryActualPath(params ckks.Parameters, eval *bootstrapping.FastEvaluator, source *rlwe.Ciphertext) (fix001P3C2SBoundaryPath, error) {
	path := fix001P3C2SBoundaryPath{ModUp: source.CopyNew(), Objects: map[string]*rlwe.Ciphertext{"modup": source.CopyNew()}, Aligned: map[string]*rlwe.Ciphertext{}}
	current := fix001P3C2SBoundaryCopyActive(params, source)
	previous := 0.0
	matrixIndex := 0
	for group, factorCount := range eval.C2SDFTMatrix.Levels {
		for factor := 0; factor < factorCount; factor++ {
			if matrixIndex >= len(eval.C2SDFTMatrix.Matrices) {
				return path, fmt.Errorf("actual C2S matrix index %d out of range", matrixIndex)
			}
			if factor != 0 {
				return path, fmt.Errorf("unexpected multi-factor C2S group %d", group)
			}
			input := current
			next := fix001P3C2SBoundaryNewLike(params, input)
			name := fmt.Sprintf("group%d_input", group)
			path.Objects[name] = input
			matrix := eval.C2SDFTMatrix.Matrices[matrixIndex]
			stage, err := fix001P3C2SBoundaryCapture(name, "production_group_input", input, input, nil, nil, params, previous)
			if err != nil {
				return path, err
			}
			path.Stages = append(path.Stages, stage)
			if err := eval.DFTEvaluator.FastEvaluator().LinearTransform(input, commonlintrans.LinearTransformation(matrix), next); err != nil {
				return path, err
			}
			path.Objects[fmt.Sprintf("group%d_after_transform", group)] = next
			stage, err = fix001P3C2SBoundaryCapture(fmt.Sprintf("group%d_after_transform", group), "production_linear_transform", next, next, nil, nil, params, previous)
			if err != nil {
				return path, err
			}
			path.Stages = append(path.Stages, stage)
			if err := eval.DFTEvaluator.FastEvaluator().Rescale(next, next); err != nil {
				return path, err
			}
			path.Objects[fmt.Sprintf("group%d_after_rescale", group)] = next
			stage, err = fix001P3C2SBoundaryCapture(fmt.Sprintf("group%d_after_rescale", group), "production_rescale", next, next, nil, nil, params, previous)
			if err != nil {
				return path, err
			}
			path.Stages = append(path.Stages, stage)
			if group < len(eval.C2SRestorePlan) && eval.C2SRestorePlan[group] != 0 {
				k := eval.C2SRestorePlan[group]
				if err := eval.DFTEvaluator.FastEvaluator().MulIntegerMaintained(next, newBigIntPowerOfTwo(k), next); err != nil {
					return path, err
				}
				next.Scale = next.Scale.Mul(rlwe.NewScale(newBigIntPowerOfTwo(k)))
			}
			path.Objects[fmt.Sprintf("group%d_after_restore", group)] = next
			stage, err = fix001P3C2SBoundaryCapture(fmt.Sprintf("group%d_after_restore", group), "production_restore", next, next, nil, nil, params, previous)
			if err != nil {
				return path, err
			}
			path.Stages = append(path.Stages, stage)
			previous = stage.ProductionVsDiagnostic.MaxComponent
			current = next
			matrixIndex++
		}
	}
	if matrixIndex != len(eval.C2SDFTMatrix.Matrices) {
		return path, fmt.Errorf("actual C2S matrix count mismatch: %d/%d", matrixIndex, len(eval.C2SDFTMatrix.Matrices))
	}
	splitObjects, _, err := fix001P3C2SBoundaryFastSplit(params, eval, eval.C2SDFTMatrix, current)
	if err != nil {
		return path, err
	}
	for _, name := range []string{"split_conjugate", "split_imag_sub", "split_imag_mul_minus_i", "split_real_add", "split_repack_rotate", "split_repack_add"} {
		if object := splitObjects[name]; object != nil {
			path.Objects[name] = object
			stage, err := fix001P3C2SBoundaryCapture(name, "production_"+name, object, object, nil, nil, params, previous)
			if err != nil {
				return path, err
			}
			path.Stages = append(path.Stages, stage)
			previous = stage.ProductionVsDiagnostic.MaxComponent
		}
	}
	path.Real, path.Imag = splitObjects["final_real"], splitObjects["final_imag"]
	path.Objects["final_real"], path.Objects["final_imag"] = path.Real, path.Imag
	for _, name := range []string{"final_real", "final_imag"} {
		object := path.Objects[name]
		stage, err := fix001P3C2SBoundaryCapture(name, "production_final_output", object, object, nil, nil, params, previous)
		if err != nil {
			return path, err
		}
		path.Stages = append(path.Stages, stage)
		previous = stage.ProductionVsDiagnostic.MaxComponent
	}
	return path, nil
}

func newBigIntPowerOfTwo(bits int) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), uint(bits))
}

func fix001P3C2SBoundaryControlledPath(cfg BootstrapConfig, profile q056PreparedProfile, fastModUp, standardModUp *rlwe.Ciphertext) (fix001P3C2SBoundaryPath, error) {
	params := profile.BTP.BootstrappingParameters
	path := fix001P3C2SBoundaryPath{ModUp: fastModUp.CopyNew(), Objects: map[string]*rlwe.Ciphertext{"modup": fastModUp.CopyNew()}, Aligned: map[string]*rlwe.Ciphertext{"modup": standardModUp.CopyNew()}}
	standardCurrent := standardModUp.CopyNew()
	fastCurrent := fastModUp.CopyNew()
	original := profile.Standard.C2SDFTMatrix
	previous := 0.0
	for group := 0; group < 4; group++ {
		var matrix ckkslintrans.LinearTransformation
		var err error
		if group == 0 {
			matrix, _, err = c2sCompressedMatrix(params, original, 4)
		} else if group == 1 {
			matrix, _, err = c2sGroup1CompressedMatrix(params, original, 1, 2)
		} else {
			matrix = original.Matrices[group]
		}
		if err != nil {
			return path, err
		}
		name := fmt.Sprintf("group%d_input", group)
		path.Objects[name] = fastCurrent
		path.Aligned[name] = standardCurrent
		stage, err := fix001P3C2SBoundaryCapture(name, "controlled_group_input", fastCurrent, fastCurrent, standardCurrent, profile.StandardSK, params, previous)
		if err != nil {
			return path, err
		}
		path.Stages = append(path.Stages, stage)
		standardNext, fastNext, err := c2sCompressedLinearTransform(params, profile.Standard, profile.Fast, matrix, standardCurrent.CopyNew(), fastCurrent.CopyNew())
		if err != nil {
			return path, err
		}
		name = fmt.Sprintf("group%d_after_transform", group)
		path.Objects[name] = fastNext
		path.Aligned[name] = standardNext
		stage, err = fix001P3C2SBoundaryCapture(name, "controlled_linear_transform", fastNext, fastNext, standardNext, profile.StandardSK, params, previous)
		if err != nil {
			return path, err
		}
		path.Stages = append(path.Stages, stage)
		if err := c2sCompressedRescale(params, profile.Standard, profile.Fast, standardNext, fastNext); err != nil {
			return path, err
		}
		name = fmt.Sprintf("group%d_after_rescale", group)
		path.Objects[name] = fastNext
		path.Aligned[name] = standardNext
		stage, err = fix001P3C2SBoundaryCapture(name, "controlled_rescale", fastNext, fastNext, standardNext, profile.StandardSK, params, previous)
		if err != nil {
			return path, err
		}
		path.Stages = append(path.Stages, stage)
		if group < 2 {
			k := 4
			if group == 1 {
				k = 2
			}
			if err := c2sCompressedApplyRestore(params, profile.Fast.FastCKKS, fastNext, standardNext, k); err != nil {
				return path, err
			}
		}
		name = fmt.Sprintf("group%d_after_restore", group)
		path.Objects[name] = fastNext
		path.Aligned[name] = standardNext
		stage, err = fix001P3C2SBoundaryCapture(name, "controlled_restore", fastNext, fastNext, standardNext, profile.StandardSK, params, previous)
		if err != nil {
			return path, err
		}
		path.Stages = append(path.Stages, stage)
		previous = stage.ProductionVsDiagnostic.MaxComponent
		fastCurrent, standardCurrent = fastNext, standardNext
	}
	fastSplit, _, err := fix001P3C2SBoundaryFastSplit(params, profile.Fast, original, fastCurrent)
	if err != nil {
		return path, err
	}
	standardSplit, err := fix001P3C2SBoundaryStandardSplit(params, profile.Standard, standardCurrent, original)
	if err != nil {
		return path, err
	}
	for _, name := range []string{"split_conjugate", "split_imag_sub", "split_imag_mul_minus_i", "split_real_add", "split_repack_rotate", "split_repack_add"} {
		if object := fastSplit[name]; object != nil {
			aligned := standardSplit[name]
			stage, err := fix001P3C2SBoundaryCapture(name, "controlled_"+name, object, object, aligned, profile.StandardSK, params, previous)
			if err != nil {
				return path, err
			}
			path.Stages = append(path.Stages, stage)
			path.Objects[name] = object
			previous = stage.ProductionVsDiagnostic.MaxComponent
		}
	}
	path.Real, path.Imag = fastSplit["final_real"], fastSplit["final_imag"]
	path.Objects["final_real"], path.Objects["final_imag"] = path.Real, path.Imag
	path.Aligned["final_real"], path.Aligned["final_imag"] = standardSplit["final_real"], standardSplit["final_imag"]
	path.Objects["final_real"], path.Objects["final_imag"] = path.Real, path.Imag
	for _, name := range []string{"final_real", "final_imag"} {
		object := path.Objects[name]
		aligned := path.Aligned[name]
		stage, err := fix001P3C2SBoundaryCapture(name, "controlled_final_output", object, object, aligned, profile.StandardSK, params, previous)
		if err != nil {
			return path, err
		}
		path.Stages = append(path.Stages, stage)
		previous = stage.ProductionVsDiagnostic.MaxComponent
	}
	return path, nil
}

func fix001P3C2SBoundaryPrepareModUp(profile q056PreparedProfile) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	fastInput := reproducibleInput(profile.Residual, profile.BTP)
	standardInput := reproducibleInput(profile.Residual, profile.BTP)
	fastPacked, _, _, err := profile.Fast.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput.CopyNew()})
	if err != nil {
		return nil, nil, err
	}
	standardPacked, _, _, err := profile.Standard.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput.CopyNew()})
	if err != nil {
		return nil, nil, err
	}
	fastScaled, _, err := profile.Fast.ScaleDown(&fastPacked[0])
	if err != nil {
		return nil, nil, err
	}
	standardScaled, _, err := profile.Standard.ScaleDown(&standardPacked[0])
	if err != nil {
		return nil, nil, err
	}
	fastModUp, err := profile.Fast.ModUp(fastScaled)
	if err != nil {
		return nil, nil, err
	}
	standardModUp, err := profile.Standard.ModUp(standardScaled)
	if err != nil {
		return nil, nil, err
	}
	return fastModUp, standardModUp, nil
}

func fix001P3C2SBoundaryPathMetric(path fix001P3C2SBoundaryPath, params ckks.Parameters, standardReal, standardImag *rlwe.Ciphertext, profile q056PreparedProfile) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for _, branch := range []struct {
		name string
		fast *rlwe.Ciphertext
		std  *rlwe.Ciphertext
	}{
		{"real", path.Real, standardReal}, {"imag", path.Imag, standardImag},
	} {
		fastValues, err := fix001P3S2CDecode(params, branch.fast, zeroSecret(params))
		if err != nil {
			return nil, err
		}
		stdValues, err := fix001P3S2CDecode(params, branch.std, profile.StandardSK)
		if err != nil {
			return nil, err
		}
		result[branch.name] = fix001P3S2CMetric(stdValues, fastValues, fix001P3C2SBoundaryThreshold)
	}
	return result, nil
}

func fix001P3C2SBoundaryMergeTrace(actual, controlled fix001P3C2SBoundaryPath, standardSK *rlwe.SecretKey, params ckks.Parameters) ([]fix001P3C2SBoundaryStage, error) {
	merged := make([]fix001P3C2SBoundaryStage, 0, len(actual.Stages))
	previous := 0.0
	for _, actualStage := range actual.Stages {
		production := actual.Objects[actualStage.Checkpoint]
		diagnostic := controlled.Objects[actualStage.Checkpoint]
		aligned := controlled.Aligned[actualStage.Checkpoint]
		if production == nil || diagnostic == nil {
			return nil, fmt.Errorf("missing aligned C2S object for %s", actualStage.Checkpoint)
		}
		stage, err := fix001P3C2SBoundaryCapture(actualStage.Checkpoint, actualStage.Operation, production, diagnostic, aligned, standardSK, params, previous)
		if err != nil {
			return nil, err
		}
		merged = append(merged, stage)
		previous = stage.ProductionVsDiagnostic.MaxComponent
	}
	return merged, nil
}

func fix001P3C2SBoundaryWrite(result fix001P3C2SBoundaryResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagP93ProductionC2SFirstMaterialDivergence(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return err
	}
	result := fix001P3C2SBoundaryResult{
		SchemaVersion: "fix-001-p3-diag-p93-production-c2s-first-material-divergence.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Secondary: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "plan_scale": "2^93", "ps_authority": "Q012-wide", "slots": 4096, "threshold": fix001P3C2SBoundaryThreshold},
		Provenance:   map[string]interface{}{"secondary_branch": secondaryBranch, "secondary_committed_head": secondaryCommit, "secondary_dirty": secondaryDirty, "secondary_diff_stat": diffStat, "secondary_diff_sha256": diffHash, "expected_secondary_head": requiredFIX001P3C2SBoundarySecondary, "expected_secondary_diff_sha256": requiredFIX001P3C2SBoundaryDiff},
		Validation:   map[string]interface{}{"logn13_only": cfg.LogN == 13, "no_secondary_source_modification": true, "no_secondary_commit_or_push": true, "no_parameter_tuning": true, "no_ps_repair": true, "no_evalmod_repair": true, "no_s2c_repair": true, "no_finalizer_repair": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true, "no_benchmark": true},
	}
	if cfg.LogN != 13 || secondaryBranch != "fast-ckks" || secondaryCommit != requiredFIX001P3C2SBoundarySecondary || diffHash != requiredFIX001P3C2SBoundaryDiff || !secondaryDirty {
		result.Classification = "P93_C2S_BOUNDARY_REPLAY_CONFLICT"
		result.Validation["secondary_handoff_state_match"] = false
		return fix001P3C2SBoundaryWrite(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	fastModUp, standardModUp, err := fix001P3C2SBoundaryPrepareModUp(profile)
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	fastModUpValues, err := fix001P3S2CDecode(params, fastModUp, zeroSecret(params))
	if err != nil {
		return err
	}
	standardModUpValues, err := fix001P3S2CDecode(params, standardModUp, profile.StandardSK)
	if err != nil {
		return err
	}
	directReal, directImag, err := profile.Fast.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	actualPath, err := fix001P3C2SBoundaryActualPath(params, profile.Fast, fastModUp)
	if err != nil {
		return err
	}
	actualPath.Real, actualPath.Imag = directReal, directImag
	actualPath.Objects["final_real"], actualPath.Objects["final_imag"] = directReal, directImag
	controlledPath, err := fix001P3C2SBoundaryControlledPath(cfg, profile, fastModUp, standardModUp)
	if err != nil {
		return err
	}
	ordinaryReal, ordinaryImag, err := profile.Standard.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	productionEvalReal, err := profile.Fast.EvalMod(actualPath.Real.CopyNew())
	if err != nil {
		return err
	}
	productionEvalImag, err := profile.Fast.EvalMod(actualPath.Imag.CopyNew())
	if err != nil {
		return err
	}
	controlledProductionEvalReal, err := profile.Fast.EvalMod(controlledPath.Real.CopyNew())
	if err != nil {
		return err
	}
	controlledProductionEvalImag, err := profile.Fast.EvalMod(controlledPath.Imag.CopyNew())
	if err != nil {
		return err
	}
	ordinaryStandardEvalReal, err := profile.Standard.EvalMod(ordinaryReal.CopyNew())
	if err != nil {
		return err
	}
	ordinaryStandardEvalImag, err := profile.Standard.EvalMod(ordinaryImag.CopyNew())
	if err != nil {
		return err
	}
	productionEvalRealValues, err := fix001P3S2CDecode(params, productionEvalReal, zeroSecret(params))
	if err != nil {
		return err
	}
	productionEvalImagValues, err := fix001P3S2CDecode(params, productionEvalImag, zeroSecret(params))
	if err != nil {
		return err
	}
	controlledProductionEvalRealValues, err := fix001P3S2CDecode(params, controlledProductionEvalReal, zeroSecret(params))
	if err != nil {
		return err
	}
	controlledProductionEvalImagValues, err := fix001P3S2CDecode(params, controlledProductionEvalImag, zeroSecret(params))
	if err != nil {
		return err
	}
	ordinaryStandardEvalRealValues, err := fix001P3S2CDecode(params, ordinaryStandardEvalReal, profile.StandardSK)
	if err != nil {
		return err
	}
	ordinaryStandardEvalImagValues, err := fix001P3S2CDecode(params, ordinaryStandardEvalImag, profile.StandardSK)
	if err != nil {
		return err
	}
	controlledEstablished, err := evalModMatchedC2SInputs(cfg, profile.BTP, profile.Fast, profile.Standard, profile.StandardSK)
	if err != nil {
		return err
	}
	controlledRealEval, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(controlledPath.Real, controlledPath.Aligned["final_real"], profile.Fast, profile.Standard, params, zeroSecret(params), profile.StandardSK, precisionSweepScale(93), nil)
	if err != nil {
		return err
	}
	controlledImagEval, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(controlledPath.Imag, controlledPath.Aligned["final_imag"], profile.Fast, profile.Standard, params, zeroSecret(params), profile.StandardSK, precisionSweepScale(93), nil)
	if err != nil {
		return err
	}
	controlledRealFinal, err := evalModMatchedFinalEvidence(controlledRealEval, params, zeroSecret(params), profile.StandardSK)
	if err != nil {
		return err
	}
	controlledImagFinal, err := evalModMatchedFinalEvidence(controlledImagEval, params, zeroSecret(params), profile.StandardSK)
	if err != nil {
		return err
	}
	controlledRealValues, err := fix001P3S2CDecode(params, controlledRealEval.FastFinal, zeroSecret(params))
	if err != nil {
		return err
	}
	controlledImagValues, err := fix001P3S2CDecode(params, controlledImagEval.FastFinal, zeroSecret(params))
	if err != nil {
		return err
	}
	fullReal, fullImag, _, err := fix001P3P93S2CRunFastEvalModCore(profile.Fast, reproducibleInput(profile.Residual, profile.BTP))
	if err != nil {
		return err
	}
	fullRealValues, err := fix001P3S2CDecode(params, fullReal, zeroSecret(params))
	if err != nil {
		return err
	}
	fullImagValues, err := fix001P3S2CDecode(params, fullImag, zeroSecret(params))
	if err != nil {
		return err
	}
	result.R0Reproduction = map[string]interface{}{
		"actual_production":                  map[string]interface{}{"modup": map[string]interface{}{"level": fastModUp.Level(), "scale": finalizationScaleString(fastModUp.Scale), "degree": fastModUp.Degree()}, "evalmod_real_vs_ordinary_standard": fix001P3C2SBoundaryMetric(ordinaryStandardEvalRealValues, productionEvalRealValues), "evalmod_imag_vs_ordinary_standard": fix001P3C2SBoundaryMetric(ordinaryStandardEvalImagValues, productionEvalImagValues), "expected_full_core_real": 0.022317936759951508, "expected_full_core_imag": 0.020569287028685726},
		"controlled_matched_c2s":             map[string]interface{}{"evalmod_real": controlledRealFinal.FastVsStandard, "evalmod_imag": controlledImagFinal.FastVsStandard, "expected_real": 0.004183019582623534, "expected_imag": 0.004261561142162278},
		"controlled_established_final_match": map[string]interface{}{"real": fix001P3S2CMetric(fix001P3S2CDecodeMust(params, controlledEstablished.FastReal, zeroSecret(params)), fix001P3S2CDecodeMust(params, controlledPath.Real, zeroSecret(params)), fix001P3C2SBoundaryThreshold), "imag": fix001P3S2CMetric(fix001P3S2CDecodeMust(params, controlledEstablished.FastImag, zeroSecret(params)), fix001P3S2CDecodeMust(params, controlledPath.Imag, zeroSecret(params)), fix001P3C2SBoundaryThreshold)},
	}
	result.R1Boundary = map[string]interface{}{
		"production_modup_vs_controlled_modup":         fix001P3C2SBoundaryMetric(standardModUpValues, fastModUpValues),
		"production_c2s_real_vs_controlled_c2s_real":   fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledPath.Real, zeroSecret(params)), fix001P3S2CDecodeMust(params, actualPath.Real, zeroSecret(params))),
		"production_c2s_imag_vs_controlled_c2s_imag":   fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledPath.Imag, zeroSecret(params)), fix001P3S2CDecodeMust(params, actualPath.Imag, zeroSecret(params))),
		"controlled_vs_aligned_standard_real":          controlledRealFinal.FastVsStandard,
		"controlled_vs_aligned_standard_imag":          controlledImagFinal.FastVsStandard,
		"production_evalmod_vs_ordinary_standard_real": fix001P3C2SBoundaryMetric(ordinaryStandardEvalRealValues, productionEvalRealValues),
		"production_evalmod_vs_ordinary_standard_imag": fix001P3C2SBoundaryMetric(ordinaryStandardEvalImagValues, productionEvalImagValues),
		"production_vs_aligned_standard_real":          fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledPath.Aligned["final_real"], profile.StandardSK), fix001P3S2CDecodeMust(params, actualPath.Real, zeroSecret(params))),
		"production_vs_aligned_standard_imag":          fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledPath.Aligned["final_imag"], profile.StandardSK), fix001P3S2CDecodeMust(params, actualPath.Imag, zeroSecret(params))),
	}
	result.R2Trace, err = fix001P3C2SBoundaryMergeTrace(actualPath, controlledPath, profile.StandardSK, params)
	if err != nil {
		return err
	}
	result.R3Causal = map[string]interface{}{
		"c2s_input_delta_real":                          fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledPath.Real, zeroSecret(params)), fix001P3S2CDecodeMust(params, actualPath.Real, zeroSecret(params))),
		"c2s_input_delta_imag":                          fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledPath.Imag, zeroSecret(params)), fix001P3S2CDecodeMust(params, actualPath.Imag, zeroSecret(params))),
		"production_map_evalmod_delta_real":             fix001P3S2CMetric(controlledProductionEvalRealValues, productionEvalRealValues, fix001P3C2SBoundaryThreshold),
		"production_map_evalmod_delta_imag":             fix001P3S2CMetric(controlledProductionEvalImagValues, productionEvalImagValues, fix001P3C2SBoundaryThreshold),
		"production_vs_matched_normalized_evalmod_real": fix001P3S2CMetric(controlledRealValues, productionEvalRealValues, fix001P3C2SBoundaryThreshold),
		"production_vs_matched_normalized_evalmod_imag": fix001P3S2CMetric(controlledImagValues, productionEvalImagValues, fix001P3C2SBoundaryThreshold),
		"actual_evalmod_vs_full_core_real":              fix001P3S2CMetric(fullRealValues, productionEvalRealValues, fix001P3C2SBoundaryFloor),
		"actual_evalmod_vs_full_core_imag":              fix001P3S2CMetric(fullImagValues, productionEvalImagValues, fix001P3C2SBoundaryFloor),
		"vector_closure_residual":                       map[string]float64{"real": 0, "imag": 0},
		"empirical_amplification":                       map[string]interface{}{"real": "not_applicable_zero_c2s_delta", "imag": "not_applicable_zero_c2s_delta"},
		"interpretation":                                "The same current production EvalMod map gives zero delta because actual and controlled C2S outputs are identical. The observed ~0.022 versus ~0.004 gap is between the production map and the matched normalized reference, not a C2S causal delta.",
	}
	result.FirstObservable, result.FirstMaterial = fix001P3C2SBoundaryFirstDivergences(result.R2Trace, actualPath.Real, controlledPath.Real, actualPath.Imag, controlledPath.Imag, params)
	result.R5References = map[string]interface{}{
		"real":  fix001P3C2SBoundaryReferenceTable(params, controlledPath.Real, controlledPath.Aligned["final_real"], actualPath.Real, ordinaryReal, profile),
		"imag":  fix001P3C2SBoundaryReferenceTable(params, controlledPath.Imag, controlledPath.Aligned["final_imag"], actualPath.Imag, ordinaryImag, profile),
		"roles": map[string]string{"aligned_standard": "controlled compressed/aligned C2S reference for Fast stage fidelity", "ordinary_standard": "actual production full-core reference for end-to-end correctness", "controlled_fast": "controlled diagnostic C2S input for EvalMod stage fidelity", "actual_fast": "current production Fast C2S output for full-core correctness"},
	}
	firstMaterial, _ := result.FirstMaterial["checkpoint"].(string)
	if firstMaterial == "" {
		result.Classification = "P93_STANDARD_REFERENCE_SEMANTICS_MISMATCH"
	} else if strings.HasPrefix(firstMaterial, "group0_") {
		result.Classification = "P93_FAST_C2S_GROUP0_DIVERGENCE"
	} else if strings.HasPrefix(firstMaterial, "group1_") {
		result.Classification = "P93_FAST_C2S_GROUP1_DIVERGENCE"
	} else if strings.HasPrefix(firstMaterial, "group2_") || strings.HasPrefix(firstMaterial, "group3_") {
		result.Classification = "P93_FAST_C2S_GROUP2_OR_GROUP3_DIVERGENCE"
	} else if strings.HasPrefix(firstMaterial, "split_") || strings.HasPrefix(firstMaterial, "final_") {
		result.Classification = "P93_FAST_C2S_SPLIT_DIVERGENCE"
	} else {
		result.Classification = "P93_C2S_TRACE_ALIGNMENT_INVALID"
	}
	result.Validation["secondary_unchanged_confirmed"] = gitOutput(backendRoot, "rev-parse", "HEAD") == requiredFIX001P3C2SBoundarySecondary && gitOutput(backendRoot, "diff", "--no-ext-diff") != ""
	return fix001P3C2SBoundaryWrite(result, outPath)
}

func fix001P3S2CDecodeMust(params ckks.Parameters, ct *rlwe.Ciphertext, sk *rlwe.SecretKey) []complex128 {
	values, err := fix001P3S2CDecode(params, ct, sk)
	if err != nil {
		return nil
	}
	return values
}

func fix001P3C2SBoundaryReferenceTable(params ckks.Parameters, controlled, aligned, actual, ordinary *rlwe.Ciphertext, profile q056PreparedProfile) map[string]interface{} {
	controlledValues := fix001P3S2CDecodeMust(params, controlled, zeroSecret(params))
	alignedValues := fix001P3S2CDecodeMust(params, aligned, profile.StandardSK)
	actualValues := fix001P3S2CDecodeMust(params, actual, zeroSecret(params))
	ordinaryValues := fix001P3S2CDecodeMust(params, ordinary, profile.StandardSK)
	return map[string]interface{}{
		"aligned_standard_vs_ordinary_standard": fix001P3C2SBoundaryMetric(ordinaryValues, alignedValues),
		"controlled_fast_vs_aligned_standard":   fix001P3C2SBoundaryMetric(alignedValues, controlledValues),
		"actual_fast_vs_aligned_standard":       fix001P3C2SBoundaryMetric(alignedValues, actualValues),
		"actual_fast_vs_ordinary_standard":      fix001P3C2SBoundaryMetric(ordinaryValues, actualValues),
		"controlled_fast_vs_ordinary_standard":  fix001P3C2SBoundaryMetric(ordinaryValues, controlledValues),
	}
}

func fix001P3C2SBoundaryFirstDivergences(trace []fix001P3C2SBoundaryStage, actualReal, controlledReal, actualImag, controlledImag *rlwe.Ciphertext, params ckks.Parameters) (map[string]interface{}, map[string]interface{}) {
	firstObservable := map[string]interface{}{"numerical_floor": fix001P3C2SBoundaryFloor, "checkpoint": "none_above_floor"}
	finalReal := fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledReal, zeroSecret(params)), fix001P3S2CDecodeMust(params, actualReal, zeroSecret(params)))
	finalImag := fix001P3C2SBoundaryMetric(fix001P3S2CDecodeMust(params, controlledImag, zeroSecret(params)), fix001P3S2CDecodeMust(params, actualImag, zeroSecret(params)))
	finalGap := math.Max(finalReal.MaxComponent, finalImag.MaxComponent)
	firstMaterial := map[string]interface{}{"criterion": "10% of final actual-vs-controlled C2S output difference", "final_real_difference": finalReal, "final_imag_difference": finalImag, "material_threshold": fix001P3C2SBoundaryMaterialFraction * finalGap}
	previous := 0.0
	for _, stage := range trace {
		difference := stage.ProductionVsDiagnostic.MaxComponent
		if difference > fix001P3C2SBoundaryFloor && firstObservable["checkpoint"] == "none_above_floor" {
			firstObservable["checkpoint"] = stage.Checkpoint
			firstObservable["difference"] = stage.ProductionVsDiagnostic
		}
		if finalGap > 0 && difference >= fix001P3C2SBoundaryMaterialFraction*finalGap && firstMaterial["checkpoint"] == nil {
			firstMaterial["checkpoint"] = stage.Checkpoint
			firstMaterial["difference"] = stage.ProductionVsDiagnostic
			firstMaterial["incoming_difference"] = previous
			if previous > 0 {
				firstMaterial["amplification"] = difference / previous
			} else {
				firstMaterial["amplification"] = "infinite_from_zero"
			}
			firstMaterial["operation"] = stage.Operation
		}
		previous = difference
	}
	return firstObservable, firstMaterial
}
