package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	ckkslintrans "github.com/tuneinsight/lattigo/v6/circuits/ckks/lintrans"
	commonlintrans "github.com/tuneinsight/lattigo/v6/circuits/common/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredC2SGroup0AliasPrimaryBase = "da13958650fbf923ffd5049f741610a8e3143c53"
	requiredC2SGroup0AliasSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
	predecessorGroup0LinearError      = 0.04690968130449508
	predecessorGroup0RescaleError     = 0.046909681304491926
	predecessorGroup0RescaleIncrement = -3.15e-15
)

type c2sAliasGiantGroup struct {
	GiantRotation int   `json:"giant_rotation"`
	BabyRotations []int `json:"baby_rotations"`
}

type c2sAliasStructure struct {
	N1                    int                  `json:"n1"`
	Path                  string               `json:"path"`
	DiagonalCount         int                  `json:"diagonal_count"`
	SortedDiagonalIndices []int                `json:"sorted_diagonal_indices"`
	GiantGroups           []c2sAliasGiantGroup `json:"giant_groups,omitempty"`
	MatrixScale           string               `json:"matrix_scale"`
	MatrixScaleLog2       float64              `json:"matrix_scale_log2"`
	LevelQ                int                  `json:"level_q"`
	LevelP                int                  `json:"level_p"`
	LogDimensions         string               `json:"log_dimensions"`
	InputLevel            int                  `json:"input_level"`
	InputScale            string               `json:"input_scale"`
}

type c2sAliasCheckpoint struct {
	Order                    int                   `json:"order"`
	Checkpoint               string                `json:"checkpoint"`
	Kind                     string                `json:"kind"`
	Diagonal                 int                   `json:"diagonal,omitempty"`
	BabyRotation             int                   `json:"baby_rotation,omitempty"`
	GiantRotation            int                   `json:"giant_rotation,omitempty"`
	FastQ01Capacity          postMod1S2CCapacity   `json:"fast_q01_capacity"`
	StandardFullRNSCapacity  postMod1S2CCapacity   `json:"standard_full_rns_capacity"`
	FullRNSCenteredQ01Unique bool                  `json:"full_rns_centered_q01_unique"`
	FastResiduesMatch        bool                  `json:"fast_residues_match_standard_mod_q01"`
	Semantic                 *semanticBisectMetric `json:"semantic,omitempty"`
}

type c2sAliasCompression struct {
	WorstCheckpoint         string  `json:"worst_checkpoint"`
	WorstExactCoefficient   string  `json:"worst_exact_coefficient"`
	WorstRatioToQ01Half     float64 `json:"worst_ratio_to_q01_half"`
	KMin                    int     `json:"k_min"`
	KSafe                   int     `json:"k_safe"`
	RequiredBeforeOperation string  `json:"required_before_operation"`
}

type c2sAliasResult struct {
	SchemaVersion                     string                      `json:"schema_version"`
	Timestamp                         time.Time                   `json:"timestamp"`
	Primary                           RepositoryMetadata          `json:"primary_repository"`
	Lattigo                           RepositoryMetadata          `json:"lattigo_repository"`
	Environment                       EnvironmentMetadata         `json:"environment"`
	Config                            BootstrapConfig             `json:"config"`
	Parameters                        ExperimentParameters        `json:"effective_parameters"`
	Workload                          CorrectnessWorkload         `json:"workload"`
	StandardProof                     semanticBisectStandardProof `json:"standard_control_proof"`
	ModUp                             *c2sPrecisionCheckpoint     `json:"modup_precondition,omitempty"`
	ModUpMetadataMatch                bool                        `json:"modup_metadata_match"`
	ModUpNoInputMutation              bool                        `json:"modup_no_input_mutation"`
	StandardEndToEnd                  *semanticBisectMetric       `json:"standard_end_to_end_semantic,omitempty"`
	DerivedC2SBudget                  float64                     `json:"derived_c2s_budget"`
	StandardNativeVsStandardOnFast    *semanticBisectMetric       `json:"standard_native_vs_standard_on_fast,omitempty"`
	C2SRealInputDelta                 *semanticBisectMetric       `json:"c2s_real_input_delta,omitempty"`
	C2SImagInputDelta                 *semanticBisectMetric       `json:"c2s_imag_input_delta,omitempty"`
	CausalSufficiency                 bool                        `json:"c2s_error_alone_sufficient_for_final_failure"`
	MatchedInputAsymmetry             bool                        `json:"matched_input_asymmetry"`
	FastEvalModMatchedInputUnresolved bool                        `json:"fast_evalmod_matched_input_divergence_remains_unresolved"`
	CompletedLinearCheckpoint         *c2sPrecisionCheckpoint     `json:"group0_completed_linear_transform"`
	CompletedRescaleCheckpoint        *c2sPrecisionCheckpoint     `json:"group0_completed_rescale"`
	Group0RescaleIncrement            float64                     `json:"group0_rescale_increment_max_component"`
	Structure                         c2sAliasStructure           `json:"group0_transform_structure"`
	Checkpoints                       []c2sAliasCheckpoint        `json:"ordered_checkpoints"`
	Compression                       *c2sAliasCompression        `json:"compression_requirement,omitempty"`
	FirstFailingCheckpoint            string                      `json:"first_failing_checkpoint"`
	FirstSupportedCause               string                      `json:"first_supported_cause"`
	Validation                        map[string]interface{}      `json:"validation"`
}

func c2sAliasWrite(result c2sAliasResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func c2sAliasStructureFromMatrix(matrix ckkslintrans.LinearTransformation, input *rlwe.Ciphertext) c2sAliasStructure {
	slots := 1 << matrix.LogDimensions.Cols
	diagonals := make([]int, 0, len(matrix.Vec))
	for diagonal := range matrix.Vec {
		diagonals = append(diagonals, diagonal&(slots-1))
	}
	sort.Ints(diagonals)
	structure := c2sAliasStructure{
		N1: matrix.N1, Path: "BSGS", DiagonalCount: len(matrix.Vec), SortedDiagonalIndices: diagsUnique(diagonals), MatrixScale: finalizationScaleString(matrix.Scale), MatrixScaleLog2: matrix.Scale.Log2(), LevelQ: matrix.LevelQ, LevelP: matrix.LevelP,
		LogDimensions: fmt.Sprintf("rows=%d,cols=%d", matrix.LogDimensions.Rows, matrix.LogDimensions.Cols), InputLevel: input.Level(), InputScale: finalizationScaleString(input.Scale),
	}
	if matrix.N1 == 0 {
		structure.Path = "direct"
		return structure
	}
	index, _, giantRotations := commonlintrans.LinearTransformation(matrix).BSGSIndex()
	sort.Ints(giantRotations)
	giantKeys := make([]int, 0, len(index))
	for giant := range index {
		giantKeys = append(giantKeys, giant)
	}
	sort.Ints(giantKeys)
	for _, giant := range giantKeys {
		baby := append([]int(nil), index[giant]...)
		sort.Ints(baby)
		structure.GiantGroups = append(structure.GiantGroups, c2sAliasGiantGroup{GiantRotation: giant, BabyRotations: baby})
	}
	return structure
}

func diagsUnique(diagonals []int) []int {
	if len(diagonals) < 2 {
		return diagonals
	}
	result := diagonals[:1]
	for _, diagonal := range diagonals[1:] {
		if diagonal != result[len(result)-1] {
			result = append(result, diagonal)
		}
	}
	return result
}

func c2sAliasPlaintext(params ckks.Parameters, matrix ckkslintrans.LinearTransformation, diagonal, level int) (*rlwe.Plaintext, error) {
	slots := 1 << matrix.LogDimensions.Cols
	value, ok := matrix.Vec[diagonal]
	if !ok {
		value, ok = matrix.Vec[diagonal-slots]
	}
	if !ok {
		return nil, fmt.Errorf("matrix diagonal %d is unavailable", diagonal)
	}
	pt := ckks.NewPlaintext(params, level)
	*pt.MetaData = *matrix.MetaData
	for limb := 0; limb <= level && limb < len(value.Q.Coeffs); limb++ {
		copy(pt.Value.Coeffs[limb], value.Q.Coeffs[limb])
	}
	return pt, nil
}

func c2sAliasRotate(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *fastckks.Evaluator, standard, fast *rlwe.Ciphertext, rotation int) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	standardOut := ckks.NewCiphertext(params, 1, standard.Level())
	fastOut := fastckks.NewCiphertext(params, 1, fast.Level())
	if rotation == 0 {
		*standardOut = *standard.CopyNew()
		*fastOut = *fast.CopyNew()
		return standardOut, fastOut, nil
	}
	*standardOut.MetaData = *standard.MetaData
	*fastOut.MetaData = *fast.MetaData
	if err := standardEval.Evaluator.Automorphism(standard, params.GaloisElement(rotation), standardOut); err != nil {
		return nil, nil, err
	}
	if err := fastEval.Automorphism(fast, fastOut, params.GaloisElement(rotation)); err != nil {
		return nil, nil, err
	}
	return standardOut, fastOut, nil
}

func c2sAliasMultiply(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *fastckks.Evaluator, standard, fast *rlwe.Ciphertext, matrix ckkslintrans.LinearTransformation, diagonal int) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	fastPT, err := c2sAliasPlaintext(params, matrix, diagonal, fast.Level())
	if err != nil {
		return nil, nil, err
	}
	standardPT, err := c2sAliasPlaintext(params, matrix, diagonal, standard.Level())
	if err != nil {
		return nil, nil, err
	}
	// Encoded linear-transform diagonals are stored in NTT/Montgomery form
	// for the Fast path. Standard Evaluator.Mul expects a non-Montgomery
	// plaintext and applies MForm itself before multiplying.
	for limb := 0; limb <= standard.Level() && limb < len(standardPT.Value.Coeffs); limb++ {
		params.RingQ().SubRings[limb].IMForm(standardPT.Value.Coeffs[limb], standardPT.Value.Coeffs[limb])
	}
	standardPT.IsMontgomery = false
	standardOut := ckks.NewCiphertext(params, 1, standard.Level())
	fastOut := fastckks.NewCiphertext(params, 1, fast.Level())
	*standardOut.MetaData = *standard.MetaData
	*fastOut.MetaData = *fast.MetaData
	if err := standardEval.Evaluator.Mul(standard, standardPT, standardOut); err != nil {
		return nil, nil, err
	}
	if err := fastEval.Mul(fast, fastPT, fastOut); err != nil {
		return nil, nil, err
	}
	return standardOut, fastOut, nil
}

func c2sAliasRowsMatch(params ckks.Parameters, standard, fast *rlwe.Ciphertext) (bool, error) {
	standardRows, err := postMod1S2CRecoverQ01(params, standard)
	if err != nil {
		return false, err
	}
	fastRows, err := postMod1S2CRecoverQ01(params, fast)
	if err != nil {
		return false, err
	}
	if len(standardRows) != len(fastRows) {
		return false, nil
	}
	for component := range standardRows {
		if len(standardRows[component]) != len(fastRows[component]) {
			return false, nil
		}
		for i := range standardRows[component] {
			if standardRows[component][i].Cmp(fastRows[component][i]) != 0 {
				return false, nil
			}
		}
	}
	return true, nil
}

func c2sAliasCheckpointCapture(name, kind string, order, diagonal, babyRotation, giantRotation int, params ckks.Parameters, standard, fast *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, budget float64) (c2sAliasCheckpoint, error) {
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
	if err != nil {
		return c2sAliasCheckpoint{}, err
	}
	standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standard)
	if err != nil {
		return c2sAliasCheckpoint{}, err
	}
	residuesMatch, err := c2sAliasRowsMatch(params, standard, fast)
	if err != nil {
		return c2sAliasCheckpoint{}, err
	}
	checkpoint := c2sAliasCheckpoint{Order: order, Checkpoint: name, Kind: kind, Diagonal: diagonal, BabyRotation: babyRotation, GiantRotation: giantRotation, FastQ01Capacity: fastCapacity, StandardFullRNSCapacity: standardCapacity, FullRNSCenteredQ01Unique: standardCapacity.Pass, FastResiduesMatch: residuesMatch}
	if standardCapacity.Pass {
		_, fastValues, err := semanticBisectView(fast, params, fastSK)
		if err != nil {
			return c2sAliasCheckpoint{}, err
		}
		standardValues, err := decodeWithSecret(params, standard, standardSK)
		if err != nil {
			return c2sAliasCheckpoint{}, err
		}
		metric, err := semanticBisectMetricFor(standardValues, fastValues, budget)
		if err != nil {
			return c2sAliasCheckpoint{}, err
		}
		checkpoint.Semantic = &metric
	}
	return checkpoint, nil
}

func c2sAliasCheckpointFailure(checkpoint c2sAliasCheckpoint, mismatch, alias string) string {
	if !checkpoint.FullRNSCenteredQ01Unique && checkpoint.FastResiduesMatch {
		return alias
	}
	if !checkpoint.FastResiduesMatch || checkpoint.Semantic != nil && !checkpoint.Semantic.Pass {
		return mismatch
	}
	return ""
}

func c2sAliasCompressionFor(checkpoint c2sAliasCheckpoint, requiredBefore string) c2sAliasCompression {
	ratio := checkpoint.StandardFullRNSCapacity.MaxAbsOverQ01Half
	kMin := int(math.Ceil(math.Log2(ratio)))
	if kMin < 1 {
		kMin = 1
	}
	return c2sAliasCompression{WorstCheckpoint: checkpoint.Checkpoint, WorstExactCoefficient: checkpoint.StandardFullRNSCapacity.MaxAbsExactCoefficient, WorstRatioToQ01Half: ratio, KMin: kMin, KSafe: kMin + 1, RequiredBeforeOperation: requiredBefore}
}

func c2sAliasResolveDiagonal(matrix ckkslintrans.LinearTransformation, key int) int {
	slots := 1 << matrix.LogDimensions.Cols
	if _, ok := matrix.Vec[key]; ok {
		return key
	}
	if _, ok := matrix.Vec[key-slots]; ok {
		return key - slots
	}
	return key
}

func c2sAliasReplayGroup0(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, standardInput, fastInput *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, budget float64) ([]c2sAliasCheckpoint, string, *c2sAliasCompression, error) {
	matrix := fastEval.C2SDFTMatrix.Matrices[0]
	structure := c2sAliasStructureFromMatrix(matrix, fastInput)
	_ = structure
	checkpoints := make([]c2sAliasCheckpoint, 0, len(matrix.Vec)*3)
	order := 0
	compression := (*c2sAliasCompression)(nil)
	record := func(checkpoint c2sAliasCheckpoint, mismatch, alias, before string) (string, error) {
		checkpoints = append(checkpoints, checkpoint)
		failure := c2sAliasCheckpointFailure(checkpoint, mismatch, alias)
		if failure != "" && compression == nil && !checkpoint.FullRNSCenteredQ01Unique && checkpoint.FastResiduesMatch {
			value := c2sAliasCompressionFor(checkpoint, before)
			compression = &value
		}
		return failure, nil
	}
	addRotation := func(standard, fast *rlwe.Ciphertext, rotation int, kind string) (string, error) {
		checkpoint, err := c2sAliasCheckpointCapture(fmt.Sprintf("%s_%d", kind, rotation), kind, order, 0, rotation, 0, params, standard, fast, standardSK, fastSK, budget)
		if err != nil {
			return "", err
		}
		order++
		return record(checkpoint, "logn13_c2s_group0_rotation_mismatch", "logn13_c2s_group0_rotation_mismatch", "rotation")
	}

	if matrix.N1 == 0 {
		rotations := make([]int, 0, len(matrix.Vec))
		slots := 1 << matrix.LogDimensions.Cols
		for diagonal := range matrix.Vec {
			rotations = append(rotations, diagonal&(slots-1))
		}
		sort.Ints(rotations)
		rotations = diagsUnique(rotations)
		standardRotations := make(map[int]*rlwe.Ciphertext)
		fastRotations := make(map[int]*rlwe.Ciphertext)
		for _, rotation := range rotations {
			standardRotation, fastRotation, err := c2sAliasRotate(params, standardEval, fastEval.FastCKKS, standardInput, fastInput, rotation)
			if err != nil {
				return checkpoints, "", compression, err
			}
			standardRotations[rotation], fastRotations[rotation] = standardRotation, fastRotation
			failure, err := addRotation(standardRotation, fastRotation, rotation, "direct_rotation")
			if err != nil || failure != "" {
				return checkpoints, failure, compression, err
			}
		}
		var standardSum, fastSum *rlwe.Ciphertext
		for _, diagonal := range rotations {
			standardTerm, fastTerm, err := c2sAliasMultiply(params, standardEval, fastEval.FastCKKS, standardRotations[diagonal], fastRotations[diagonal], matrix, c2sAliasResolveDiagonal(matrix, diagonal))
			if err != nil {
				return checkpoints, "", compression, err
			}
			checkpoint, err := c2sAliasCheckpointCapture(fmt.Sprintf("direct_diagonal_%d", diagonal), "diagonal_product", order, diagonal, diagonal, 0, params, standardTerm, fastTerm, standardSK, fastSK, budget)
			if err != nil {
				return checkpoints, "", compression, err
			}
			order++
			failure, err := record(checkpoint, "logn13_c2s_group0_diagonal_product_mismatch", "logn13_c2s_group0_single_term_q01_alias", "term_multiplication")
			if err != nil || failure != "" {
				return checkpoints, failure, compression, err
			}
			if standardSum == nil {
				standardSum, fastSum = standardTerm, fastTerm
			} else {
				if err := standardEval.Evaluator.Add(standardSum, standardTerm, standardSum); err != nil {
					return checkpoints, "", compression, err
				}
				if err := fastEval.FastCKKS.Add(fastSum, fastTerm, fastSum); err != nil {
					return checkpoints, "", compression, err
				}
			}
			checkpoint, err = c2sAliasCheckpointCapture(fmt.Sprintf("direct_outer_partial_%d", diagonal), "outer_accumulation", order, diagonal, diagonal, 0, params, standardSum, fastSum, standardSK, fastSK, budget)
			if err != nil {
				return checkpoints, "", compression, err
			}
			order++
			failure, err = record(checkpoint, "logn13_c2s_group0_outer_accumulation_mismatch", "logn13_c2s_group0_outer_accumulation_q01_alias", "accumulation")
			if err != nil || failure != "" {
				return checkpoints, failure, compression, err
			}
		}
		return checkpoints, "", compression, nil
	}

	index, _, babyRotations := commonlintrans.LinearTransformation(matrix).BSGSIndex()
	sort.Ints(babyRotations)
	standardRotations := make(map[int]*rlwe.Ciphertext)
	fastRotations := make(map[int]*rlwe.Ciphertext)
	for _, rotation := range babyRotations {
		standardRotation, fastRotation, err := c2sAliasRotate(params, standardEval, fastEval.FastCKKS, standardInput, fastInput, rotation)
		if err != nil {
			return checkpoints, "", compression, err
		}
		standardRotations[rotation], fastRotations[rotation] = standardRotation, fastRotation
		failure, err := addRotation(standardRotation, fastRotation, rotation, "baby_rotation")
		if err != nil || failure != "" {
			return checkpoints, failure, compression, err
		}
	}
	var standardSum, fastSum *rlwe.Ciphertext
	giantKeys := make([]int, 0, len(index))
	for giant := range index {
		giantKeys = append(giantKeys, giant)
	}
	sort.Ints(giantKeys)
	for _, giant := range giantKeys {
		baby := append([]int(nil), index[giant]...)
		sort.Ints(baby)
		var standardInner, fastInner *rlwe.Ciphertext
		for _, babyRotation := range baby {
			key := c2sAliasResolveDiagonal(matrix, giant+babyRotation)
			standardTerm, fastTerm, err := c2sAliasMultiply(params, standardEval, fastEval.FastCKKS, standardRotations[babyRotation], fastRotations[babyRotation], matrix, key)
			if err != nil {
				return checkpoints, "", compression, err
			}
			checkpoint, err := c2sAliasCheckpointCapture(fmt.Sprintf("giant_%d_diagonal_%d", giant, key), "diagonal_product", order, key, babyRotation, giant, params, standardTerm, fastTerm, standardSK, fastSK, budget)
			if err != nil {
				return checkpoints, "", compression, err
			}
			order++
			failure, err := record(checkpoint, "logn13_c2s_group0_diagonal_product_mismatch", "logn13_c2s_group0_single_term_q01_alias", "term_multiplication")
			if err != nil || failure != "" {
				return checkpoints, failure, compression, err
			}
			if standardInner == nil {
				standardInner, fastInner = standardTerm, fastTerm
			} else {
				if err := standardEval.Evaluator.Add(standardInner, standardTerm, standardInner); err != nil {
					return checkpoints, "", compression, err
				}
				if err := fastEval.FastCKKS.Add(fastInner, fastTerm, fastInner); err != nil {
					return checkpoints, "", compression, err
				}
			}
			checkpoint, err = c2sAliasCheckpointCapture(fmt.Sprintf("giant_%d_inner_partial_%d", giant, babyRotation), "inner_accumulation", order, key, babyRotation, giant, params, standardInner, fastInner, standardSK, fastSK, budget)
			if err != nil {
				return checkpoints, "", compression, err
			}
			order++
			failure, err = record(checkpoint, "logn13_c2s_group0_inner_accumulation_mismatch", "logn13_c2s_group0_inner_accumulation_q01_alias", "accumulation")
			if err != nil || failure != "" {
				return checkpoints, failure, compression, err
			}
		}
		standardOuter, fastOuter := standardInner, fastInner
		if giant != 0 {
			var err error
			standardOuter, fastOuter, err = c2sAliasRotate(params, standardEval, fastEval.FastCKKS, standardInner, fastInner, giant)
			if err != nil {
				return checkpoints, "", compression, err
			}
			checkpoint, err := c2sAliasCheckpointCapture(fmt.Sprintf("giant_rotation_%d", giant), "giant_rotation", order, 0, 0, giant, params, standardOuter, fastOuter, standardSK, fastSK, budget)
			if err != nil {
				return checkpoints, "", compression, err
			}
			order++
			failure, err := record(checkpoint, "logn13_c2s_group0_giant_rotation_mismatch", "logn13_c2s_group0_giant_rotation_mismatch", "rotation")
			if err != nil || failure != "" {
				return checkpoints, failure, compression, err
			}
		}
		if standardSum == nil {
			standardSum, fastSum = standardOuter, fastOuter
		} else {
			if err := standardEval.Evaluator.Add(standardSum, standardOuter, standardSum); err != nil {
				return checkpoints, "", compression, err
			}
			if err := fastEval.FastCKKS.Add(fastSum, fastOuter, fastSum); err != nil {
				return checkpoints, "", compression, err
			}
		}
		checkpoint, err := c2sAliasCheckpointCapture(fmt.Sprintf("giant_%d_outer_partial", giant), "outer_accumulation", order, 0, 0, giant, params, standardSum, fastSum, standardSK, fastSK, budget)
		if err != nil {
			return checkpoints, "", compression, err
		}
		order++
		failure, err := record(checkpoint, "logn13_c2s_group0_outer_accumulation_mismatch", "logn13_c2s_group0_outer_accumulation_q01_alias", "accumulation")
		if err != nil || failure != "" {
			return checkpoints, failure, compression, err
		}
	}
	return checkpoints, "", compression, nil
}

func runFIX001P3DiagLogN13C2SGroup0LinearAlias(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	params := btp.BootstrappingParameters
	result := c2sAliasResult{
		SchemaVersion: "fix-001-p3-diag-logn13-c2s-group0-linear-alias.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		StandardProof:          semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: standardEval.Evaluator == nil && standardEval.DFTEvaluator == nil && standardEval.Mod1Evaluator == nil, OrdinaryStageCallsUsed: true},
		FirstFailingCheckpoint: "none", FirstSupportedCause: "logn13_c2s_group0_linear_precondition_mismatch", MatchedInputAsymmetry: false, FastEvalModMatchedInputUnresolved: true,
		Validation: map[string]interface{}{"primary_required_base": requiredC2SGroup0AliasPrimaryBase, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredC2SGroup0AliasSecondary, "no_secondary_production_changes": true, "no_matrix_generation_or_parameter_changes": true, "no_production_fix": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_later_c2s_groups": true, "no_evalmod_fix": true},
	}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil {
		return c2sAliasWrite(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return c2sAliasWrite(result, outPath)
	}
	standardControlDecoded, err := decodeWithSecret(residual, standardControl, skN1)
	if err != nil {
		return err
	}
	standardMetric, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), standardControlDecoded, c2sPrecisionFinalThreshold)
	if err != nil {
		return err
	}
	result.StandardEndToEnd = &standardMetric
	if !standardMetric.Pass {
		return c2sAliasWrite(result, outPath)
	}
	fastInput := reproducibleInput(residual, btp)
	standardInput := reproducibleInput(residual, btp)
	fastBefore, standardBefore := fastInput.CopyNew(), standardInput.CopyNew()
	fastPacked, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*fastInput.CopyNew()})
	if err != nil {
		return err
	}
	standardPacked, _, _, err := standardEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*standardInput.CopyNew()})
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
	fastSK := zeroSecret(params)
	fastMetadata, fastValues, err := semanticBisectView(fastModUp, params, fastSK)
	if err != nil {
		return err
	}
	standardMetadata, standardValues, err := c2sPrecisionFullView(standardModUp, params, skN2)
	if err != nil {
		return err
	}
	modUpMetric, err := semanticBisectMetricFor(standardValues, fastValues, c2sPrecisionInputThreshold)
	if err != nil {
		return err
	}
	result.ModUpMetadataMatch = fastMetadata.SourceLevel == standardMetadata.SourceLevel && fastMetadata.Scale == standardMetadata.Scale && fastMetadata.Degree == standardMetadata.Degree && fastMetadata.N == standardMetadata.N && fastMetadata.LogDimensions == standardMetadata.LogDimensions && fastMetadata.IsNTT == standardMetadata.IsNTT && fastMetadata.IsMontgomery && !standardMetadata.IsMontgomery
	result.ModUpNoInputMutation = fix001P3E2EInputUnchanged(fastBefore, fastInput) && fix001P3E2EInputUnchanged(standardBefore, standardInput)
	result.ModUp = &c2sPrecisionCheckpoint{Checkpoint: "mod_up", Fast: fastMetadata, Standard: standardMetadata, FastVsStandard: modUpMetric}
	if !result.ModUpMetadataMatch || !result.ModUpNoInputMutation || !modUpMetric.Pass {
		result.FirstFailingCheckpoint = "mod_up"
		return c2sAliasWrite(result, outPath)
	}

	fastReal, fastImag, err := fastEval.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	if len(fastEval.C2SDFTMatrix.Matrices) == 0 || len(standardEval.C2SDFTMatrix.Matrices) == 0 {
		return c2sAliasWrite(result, outPath)
	}
	result.Structure = c2sAliasStructureFromMatrix(fastEval.C2SDFTMatrix.Matrices[0], fastModUp)
	if result.Structure.N1 == 0 {
		result.Structure.Path = "direct"
	}
	standardFromFast, err := evalModStandardFromFast(params, fastReal, func() []complex128 { _, values, _ := semanticBisectView(fastReal, params, fastSK); return values }(), skN2)
	if err != nil {
		return err
	}
	standardFastEvidence, err := evalModMatchedInput("standard_from_fast_c2s_real", fastReal, standardFromFast, params, fastSK, skN2)
	if err != nil {
		return err
	}
	if !standardFastEvidence.Pass {
		return c2sAliasWrite(result, outPath)
	}
	standardNativeEvalMod, err := standardEval.EvalMod(standardReal.CopyNew())
	if err != nil {
		return err
	}
	standardOnFastEvalMod, err := standardEval.EvalMod(standardFromFast.CopyNew())
	if err != nil {
		return err
	}
	standardNativeValues, err := decodeWithSecret(params, standardNativeEvalMod, skN2)
	if err != nil {
		return err
	}
	standardOnFastValues, err := decodeWithSecret(params, standardOnFastEvalMod, skN2)
	if err != nil {
		return err
	}
	standardSensitivity, err := semanticBisectMetricFor(standardNativeValues, standardOnFastValues, c2sPrecisionFinalThreshold)
	if err != nil {
		return err
	}
	result.StandardNativeVsStandardOnFast = &standardSensitivity
	standardRealValues, err := decodeWithSecret(params, standardReal, skN2)
	if err != nil {
		return err
	}
	realInputDelta, err := semanticBisectMetricFor(standardRealValues, func() []complex128 { _, values, _ := semanticBisectView(fastReal, params, fastSK); return values }(), 1)
	if err != nil {
		return err
	}
	result.C2SRealInputDelta = &realInputDelta
	if fastImag != nil && standardImag != nil {
		fastImagValues, decodeErr := func() ([]complex128, error) {
			_, values, err := semanticBisectView(fastImag, params, fastSK)
			return values, err
		}()
		if decodeErr != nil {
			return decodeErr
		}
		standardImagValues, decodeErr := decodeWithSecret(params, standardImag, skN2)
		if decodeErr != nil {
			return decodeErr
		}
		imagInputDelta, metricErr := semanticBisectMetricFor(standardImagValues, fastImagValues, 1)
		if metricErr != nil {
			return metricErr
		}
		result.C2SImagInputDelta = &imagInputDelta
	}
	if realInputDelta.MaxComponentAbs > 0 {
		result.DerivedC2SBudget = c2sPrecisionFinalThreshold / (standardSensitivity.MaxComponentAbs / realInputDelta.MaxComponentAbs)
	}
	result.CausalSufficiency = result.DerivedC2SBudget > 0 && standardSensitivity.MaxComponentAbs > c2sPrecisionFinalThreshold && realInputDelta.MaxComponentAbs > result.DerivedC2SBudget

	fastTransform := fastEval.DFTEvaluator.FastEvaluator()
	fastGroupOut := fastckks.NewCiphertext(params, 1, fastModUp.Level())
	*fastGroupOut.MetaData = *fastModUp.MetaData
	standardGroupOut := ckks.NewCiphertext(params, 1, standardModUp.Level())
	*standardGroupOut.MetaData = *standardModUp.MetaData
	if err := fastTransform.LinearTransform(fastModUp.CopyNew(), commonlintrans.LinearTransformation(fastEval.C2SDFTMatrix.Matrices[0]), fastGroupOut); err != nil {
		return err
	}
	if err := standardEval.DFTEvaluator.LTEvaluator.Evaluate(standardModUp.CopyNew(), standardEval.C2SDFTMatrix.Matrices[0], standardGroupOut); err != nil {
		return err
	}
	completedLinear, err := c2sPrecisionCapture("group0_completed_linear_transform", fastGroupOut, standardGroupOut, params, fastSK, skN2, standardValues, result.DerivedC2SBudget)
	if err != nil {
		return err
	}
	result.CompletedLinearCheckpoint = &completedLinear
	if err := fastTransform.Rescale(fastGroupOut, fastGroupOut); err != nil {
		return err
	}
	if err := standardEval.DFTEvaluator.Rescale(standardGroupOut, standardGroupOut); err != nil {
		return err
	}
	completedRescale, err := c2sPrecisionCapture("group0_completed_rescale", fastGroupOut, standardGroupOut, params, fastSK, skN2, standardValues, result.DerivedC2SBudget)
	if err != nil {
		return err
	}
	result.CompletedRescaleCheckpoint = &completedRescale
	result.Group0RescaleIncrement = completedRescale.FastVsStandard.MaxComponentAbs - completedLinear.FastVsStandard.MaxComponentAbs
	if math.Abs(completedLinear.FastVsStandard.MaxComponentAbs-predecessorGroup0LinearError) > 1e-3 || math.Abs(completedRescale.FastVsStandard.MaxComponentAbs-predecessorGroup0RescaleError) > 1e-3 || math.Abs(result.Group0RescaleIncrement-predecessorGroup0RescaleIncrement) > 1e-6 || completedLinear.StandardFullRNSCapacity == nil || completedLinear.StandardFullRNSCapacity.Pass || !completedLinear.FastQ01Capacity.Pass {
		result.FirstFailingCheckpoint = "group0_completed_linear_transform"
		return c2sAliasWrite(result, outPath)
	}

	checkpoints, classification, compression, err := c2sAliasReplayGroup0(params, standardEval, fastEval, standardModUp, fastModUp, skN2, fastSK, result.DerivedC2SBudget)
	if err != nil {
		return err
	}
	result.Checkpoints, result.Compression = checkpoints, compression
	result.FirstSupportedCause = classification
	if classification == "" {
		result.FirstSupportedCause = "logn13_c2s_group0_linear_no_internal_divergence"
	}
	for _, checkpoint := range checkpoints {
		if checkpoint.Kind == "diagonal_product" || checkpoint.Kind == "inner_accumulation" || checkpoint.Kind == "outer_accumulation" || checkpoint.Kind == "giant_rotation" || checkpoint.Kind == "direct_rotation" || checkpoint.Kind == "baby_rotation" {
			if !checkpoint.FastResiduesMatch {
				result.FirstFailingCheckpoint = checkpoint.Checkpoint
				break
			}
			if !checkpoint.FullRNSCenteredQ01Unique && checkpoint.FastResiduesMatch {
				result.FirstFailingCheckpoint = checkpoint.Checkpoint
				break
			}
			if checkpoint.Semantic != nil && !checkpoint.Semantic.Pass {
				result.FirstFailingCheckpoint = checkpoint.Checkpoint
				break
			}
		}
	}
	return c2sAliasWrite(result, outPath)
}
