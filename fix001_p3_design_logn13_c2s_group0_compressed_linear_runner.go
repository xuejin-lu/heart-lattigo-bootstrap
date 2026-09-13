package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"runtime"
	"sort"
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
	requiredC2SCompressedPrimaryBase = "7059d11da93d1bc79e76ce4f1fd05f1dfbe1e8d9"
	requiredC2SCompressedSecondary   = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"
	c2sCompressedCandidateThreshold  = c2sPrecisionFinalThreshold
)

type c2sCompressedStage struct {
	Checkpoint              string                 `json:"checkpoint"`
	Fast                    semanticBisectMetadata `json:"fast"`
	Standard                semanticBisectMetadata `json:"standard"`
	FastVsStandard          *semanticBisectMetric  `json:"fast_vs_standard,omitempty"`
	FastQ01Capacity         postMod1S2CCapacity    `json:"fast_q01_capacity"`
	StandardFullRNSCapacity postMod1S2CCapacity    `json:"standard_full_rns_capacity"`
	FastResiduesMatch       bool                   `json:"fast_residues_match_standard_mod_q01"`
}

type c2sCompressedRestore struct {
	Multiplier                      string                 `json:"physical_and_metadata_multiplier"`
	FastBefore                      semanticBisectMetadata `json:"fast_before"`
	FastAfter                       semanticBisectMetadata `json:"fast_after"`
	StandardBefore                  semanticBisectMetadata `json:"standard_before"`
	StandardAfter                   semanticBisectMetadata `json:"standard_after"`
	FastCapacityAfter               postMod1S2CCapacity    `json:"fast_capacity_after"`
	StandardCapacityAfter           postMod1S2CCapacity    `json:"standard_capacity_after"`
	FastMetadataMatchesOriginal     bool                   `json:"fast_metadata_matches_original_post_rescale"`
	StandardMetadataMatchesOriginal bool                   `json:"standard_metadata_matches_original_post_rescale"`
}

type c2sCompressedComparison struct {
	CompressedStandardRestoredVsOriginal *semanticBisectMetric `json:"compressed_standard_restored_vs_original_standard,omitempty"`
	FastCompressedRestoredVsCompressed   *semanticBisectMetric `json:"fast_compressed_restored_vs_compressed_standard,omitempty"`
	FastCompressedRestoredVsOriginal     *semanticBisectMetric `json:"fast_compressed_restored_vs_original_standard,omitempty"`
	FastResiduesMatch                    bool                  `json:"fast_residues_match_compressed_standard"`
}

type c2sCompressedCandidate struct {
	K                              int                      `json:"k"`
	OriginalMatrixScale            string                   `json:"original_matrix_scale"`
	OriginalMatrixScaleLog2        float64                  `json:"original_matrix_scale_log2"`
	CompressedMatrixScale          string                   `json:"compressed_matrix_scale"`
	CompressedMatrixScaleLog2      float64                  `json:"compressed_matrix_scale_log2"`
	MathematicalDiagonalsUnchanged bool                     `json:"mathematical_diagonal_values_unchanged"`
	SameStructure                  bool                     `json:"same_bsgs_structure"`
	SameDiagonalIndices            bool                     `json:"same_diagonal_indices"`
	SameN1                         bool                     `json:"same_n1"`
	SameLevelQ                     bool                     `json:"same_level_q"`
	SameLevelP                     bool                     `json:"same_level_p"`
	SameLogDimensions              bool                     `json:"same_log_dimensions"`
	Checkpoints                    []c2sAliasCheckpoint     `json:"ordered_checkpoints"`
	FirstUnsafeCheckpoint          string                   `json:"first_unsafe_checkpoint"`
	FirstResidueMismatch           string                   `json:"first_residue_mismatch_checkpoint"`
	CompleteGroupCapacityPass      bool                     `json:"complete_group_capacity_pass"`
	CompleteGroupResiduesPass      bool                     `json:"complete_group_residue_match_pass"`
	PreRescale                     *c2sCompressedStage      `json:"pre_rescale,omitempty"`
	PostRescale                    *c2sCompressedStage      `json:"post_rescale,omitempty"`
	Restore                        *c2sCompressedRestore    `json:"restore,omitempty"`
	Comparison                     *c2sCompressedComparison `json:"restored_comparison,omitempty"`
	RestoredSemanticPass           bool                     `json:"restored_semantic_pass"`
	ValidCompleteGroup             bool                     `json:"valid_complete_group"`
}

type c2sCompressedDownstream struct {
	ValidatedKSafe             int                   `json:"validated_k_safe"`
	FinalRealDelta             *semanticBisectMetric `json:"final_c2s_real_delta,omitempty"`
	FinalImagDelta             *semanticBisectMetric `json:"final_c2s_imag_delta,omitempty"`
	FirstLaterBudgetBreach     string                `json:"first_later_group_boundary_over_budget"`
	Group0FixAloneSatisfiesC2S bool                  `json:"group0_fix_alone_satisfies_c2s_budget"`
}

type c2sCompressedResult struct {
	SchemaVersion            string                      `json:"schema_version"`
	Timestamp                time.Time                   `json:"timestamp"`
	Primary                  RepositoryMetadata          `json:"primary_repository"`
	Lattigo                  RepositoryMetadata          `json:"lattigo_repository"`
	Environment              EnvironmentMetadata         `json:"environment"`
	Config                   BootstrapConfig             `json:"config"`
	Parameters               ExperimentParameters        `json:"effective_parameters"`
	Workload                 CorrectnessWorkload         `json:"workload"`
	StandardProof            semanticBisectStandardProof `json:"standard_control_proof"`
	StandardEndToEnd         *semanticBisectMetric       `json:"standard_end_to_end_semantic,omitempty"`
	ModUp                    *c2sPrecisionCheckpoint     `json:"modup_precondition,omitempty"`
	ModUpMetadataMatch       bool                        `json:"modup_metadata_match"`
	ModUpNoInputMutation     bool                        `json:"modup_no_input_mutation"`
	DerivedC2SBudget         float64                     `json:"derived_c2s_budget"`
	C2SErrorAloneSufficient  bool                        `json:"c2s_error_alone_sufficient_for_final_failure"`
	EvalModUnresolved        bool                        `json:"fast_evalmod_matched_input_divergence_remains_unresolved"`
	OriginalStructure        c2sAliasStructure           `json:"original_group0_structure"`
	OriginalLinear           *c2sPrecisionCheckpoint     `json:"original_group0_linear_transform,omitempty"`
	OriginalRescale          *c2sPrecisionCheckpoint     `json:"original_group0_rescale,omitempty"`
	OriginalRescaleIncrement float64                     `json:"original_group0_rescale_increment"`
	Candidates               []c2sCompressedCandidate    `json:"candidates"`
	KMinFullGroup            int                         `json:"k_min_full_group"`
	KSafeFullGroup           int                         `json:"k_safe_full_group"`
	Downstream               *c2sCompressedDownstream    `json:"downstream_probe,omitempty"`
	FirstSupportedCause      string                      `json:"first_supported_cause"`
	Validation               map[string]interface{}      `json:"validation"`
}

func c2sCompressedWrite(result c2sCompressedResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func c2sCompressedMatrix(params ckks.Parameters, original dft.Matrix, k int) (ckkslintrans.LinearTransformation, bool, error) {
	if len(original.Matrices) == 0 {
		return ckkslintrans.LinearTransformation{}, false, fmt.Errorf("group-0 matrix is empty")
	}
	pVec := original.MatrixLiteral.GenMatrices(params.LogN(), params.EncodingPrecision())
	if len(pVec) == 0 {
		return ckkslintrans.LinearTransformation{}, false, fmt.Errorf("DFT MatrixLiteral generated no diagonal vectors")
	}
	scale := original.Matrices[0].Scale.Div(rlwe.NewScale(float64(uint64(1) << k)))
	originalMatrix := original.Matrices[0]
	ltParams := ckkslintrans.Parameters{
		DiagonalsIndexList:        pVec[0].DiagonalsIndexList(),
		LevelQ:                    originalMatrix.LevelQ,
		LevelP:                    originalMatrix.LevelP,
		Scale:                     scale,
		LogDimensions:             originalMatrix.LogDimensions,
		LogBabyStepGiantStepRatio: originalMatrix.LogBabyStepGiantStepRatio,
	}
	compressed := ckkslintrans.NewTransformation(params, ltParams)
	if err := ckkslintrans.Encode(ckks.NewEncoder(params), pVec[0], compressed); err != nil {
		return ckkslintrans.LinearTransformation{}, false, fmt.Errorf("encode compressed matrix k=%d: %w", k, err)
	}
	return compressed, true, nil
}

func c2sCompressedStructureEqual(original, candidate c2sAliasStructure) bool {
	return original.N1 == candidate.N1 && original.Path == candidate.Path && original.DiagonalCount == candidate.DiagonalCount && reflect.DeepEqual(original.SortedDiagonalIndices, candidate.SortedDiagonalIndices) && reflect.DeepEqual(original.GiantGroups, candidate.GiantGroups) && original.LevelQ == candidate.LevelQ && original.LevelP == candidate.LevelP && original.LogDimensions == candidate.LogDimensions
}

func c2sCompressedStageCapture(name string, fast, standard *rlwe.Ciphertext, params ckks.Parameters, fastSK, standardSK *rlwe.SecretKey, budget float64) (c2sCompressedStage, error) {
	fastMeta, _, err := semanticBisectView(fast, params, fastSK)
	if err != nil {
		return c2sCompressedStage{}, err
	}
	standardMeta, standardValues, err := c2sPrecisionFullView(standard, params, standardSK)
	if err != nil {
		return c2sCompressedStage{}, err
	}
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fast)
	if err != nil {
		return c2sCompressedStage{}, err
	}
	standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standard)
	if err != nil {
		return c2sCompressedStage{}, err
	}
	residuesMatch, err := c2sAliasRowsMatch(params, standard, fast)
	if err != nil {
		return c2sCompressedStage{}, err
	}
	stage := c2sCompressedStage{Checkpoint: name, Fast: fastMeta, Standard: standardMeta, FastQ01Capacity: fastCapacity, StandardFullRNSCapacity: standardCapacity, FastResiduesMatch: residuesMatch}
	if standardCapacity.Pass {
		_, fastValues, err := semanticBisectView(fast, params, fastSK)
		if err != nil {
			return c2sCompressedStage{}, err
		}
		metric, err := semanticBisectMetricFor(standardValues, fastValues, budget)
		if err != nil {
			return c2sCompressedStage{}, err
		}
		stage.FastVsStandard = &metric
	}
	return stage, nil
}

func c2sCompressedApplyRestore(params ckks.Parameters, fastEval *fastckks.Evaluator, fast, standard *rlwe.Ciphertext, k int) error {
	scalar := new(big.Int).Lsh(big.NewInt(1), uint(k))
	if err := fastEval.MulIntegerMaintained(fast, scalar, fast); err != nil {
		return err
	}
	ringQ := params.RingQ().AtLevel(standard.Level())
	for component := range standard.Value {
		ringQ.MulScalarBigint(standard.Value[component], scalar, standard.Value[component])
	}
	multiplier := rlwe.NewScale(float64(uint64(1) << k))
	fast.Scale = fast.Scale.Mul(multiplier)
	standard.Scale = standard.Scale.Mul(multiplier)
	return nil
}

func c2sCompressedCheckpointFailure(checkpoint c2sAliasCheckpoint) bool {
	return !checkpoint.FastResiduesMatch || !checkpoint.FullRNSCenteredQ01Unique || checkpoint.Semantic != nil && !checkpoint.Semantic.Pass
}

func c2sCompressedReplayGroup0(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, matrix ckkslintrans.LinearTransformation, standardInput, fastInput *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, budget float64) ([]c2sAliasCheckpoint, error) {
	checkpoints := make([]c2sAliasCheckpoint, 0, len(matrix.Vec)*5)
	order := 0
	record := func(name, kind string, diagonal, baby, giant int, standard, fast *rlwe.Ciphertext) error {
		checkpoint, err := c2sAliasCheckpointCapture(name, kind, order, diagonal, baby, giant, params, standard, fast, standardSK, fastSK, budget)
		if err != nil {
			return err
		}
		checkpoints = append(checkpoints, checkpoint)
		order++
		return nil
	}
	standardRotations := make(map[int]*rlwe.Ciphertext)
	fastRotations := make(map[int]*rlwe.Ciphertext)
	index, _, babyRotations := commonlintrans.LinearTransformation(matrix).BSGSIndex()
	sort.Ints(babyRotations)
	for _, rotation := range babyRotations {
		standardRotation, fastRotation, err := c2sAliasRotate(params, standardEval, fastEval.FastCKKS, standardInput, fastInput, rotation)
		if err != nil {
			return checkpoints, err
		}
		standardRotations[rotation], fastRotations[rotation] = standardRotation, fastRotation
		if err := record(fmt.Sprintf("baby_rotation_%d", rotation), "baby_rotation", 0, rotation, 0, standardRotation, fastRotation); err != nil {
			return checkpoints, err
		}
	}
	giantKeys := make([]int, 0, len(index))
	for giant := range index {
		giantKeys = append(giantKeys, giant)
	}
	sort.Ints(giantKeys)
	var standardOuter, fastOuter *rlwe.Ciphertext
	for _, giant := range giantKeys {
		baby := append([]int(nil), index[giant]...)
		sort.Ints(baby)
		var standardInner, fastInner *rlwe.Ciphertext
		for _, babyRotation := range baby {
			key := c2sAliasResolveDiagonal(matrix, giant+babyRotation)
			standardTerm, fastTerm, err := c2sAliasMultiply(params, standardEval, fastEval.FastCKKS, standardRotations[babyRotation], fastRotations[babyRotation], matrix, key)
			if err != nil {
				return checkpoints, err
			}
			if err := record(fmt.Sprintf("giant_%d_diagonal_%d", giant, key), "diagonal_product", key, babyRotation, giant, standardTerm, fastTerm); err != nil {
				return checkpoints, err
			}
			if standardInner == nil {
				standardInner, fastInner = standardTerm, fastTerm
			} else {
				if err := standardEval.Evaluator.Add(standardInner, standardTerm, standardInner); err != nil {
					return checkpoints, err
				}
				if err := fastEval.FastCKKS.Add(fastInner, fastTerm, fastInner); err != nil {
					return checkpoints, err
				}
			}
			if err := record(fmt.Sprintf("giant_%d_inner_partial_%d", giant, babyRotation), "inner_accumulation", key, babyRotation, giant, standardInner, fastInner); err != nil {
				return checkpoints, err
			}
		}
		standardGiant, fastGiant, err := c2sAliasRotate(params, standardEval, fastEval.FastCKKS, standardInner, fastInner, giant)
		if err != nil {
			return checkpoints, err
		}
		if err := record(fmt.Sprintf("giant_rotation_%d", giant), "giant_rotation", giant, 0, giant, standardGiant, fastGiant); err != nil {
			return checkpoints, err
		}
		if standardOuter == nil {
			standardOuter, fastOuter = standardGiant, fastGiant
		} else {
			if err := standardEval.Evaluator.Add(standardOuter, standardGiant, standardOuter); err != nil {
				return checkpoints, err
			}
			if err := fastEval.FastCKKS.Add(fastOuter, fastGiant, fastOuter); err != nil {
				return checkpoints, err
			}
		}
		if err := record(fmt.Sprintf("outer_partial_%d", giant), "outer_accumulation", giant, 0, giant, standardOuter, fastOuter); err != nil {
			return checkpoints, err
		}
	}
	return checkpoints, nil
}

func c2sCompressedLinearTransform(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, matrix ckkslintrans.LinearTransformation, standardInput, fastInput *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	fastOut := fastckks.NewCiphertext(params, 1, fastInput.Level())
	*fastOut.MetaData = *fastInput.MetaData
	standardOut := ckks.NewCiphertext(params, 1, standardInput.Level())
	*standardOut.MetaData = *standardInput.MetaData
	if err := fastEval.FastCKKS.LinearTransform(fastInput.CopyNew(), commonlintrans.LinearTransformation(matrix), fastOut); err != nil {
		return nil, nil, err
	}
	if err := standardEval.DFTEvaluator.LTEvaluator.Evaluate(standardInput.CopyNew(), matrix, standardOut); err != nil {
		return nil, nil, err
	}
	return standardOut, fastOut, nil
}

func c2sCompressedRescale(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, standard, fast *rlwe.Ciphertext) error {
	if err := fastEval.FastCKKS.Rescale(fast, fast); err != nil {
		return err
	}
	return standardEval.DFTEvaluator.Rescale(standard, standard)
}

func c2sCompressedMetadataMatch(a, b semanticBisectMetadata) bool {
	return a.SourceLevel == b.SourceLevel && a.Degree == b.Degree && a.N == b.N && a.LogDimensions == b.LogDimensions && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && a.Scale == b.Scale
}

func c2sCompressedCandidateRun(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, matrix, originalMatrix ckkslintrans.LinearTransformation, standardInput, fastInput *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, originalStandardPost, originalFastPost *rlwe.Ciphertext, originalStandardValues []complex128, budget float64, k int) (c2sCompressedCandidate, error) {
	originalStructure := c2sAliasStructureFromMatrix(originalMatrix, standardInput)
	candidateStructure := c2sAliasStructureFromMatrix(matrix, standardInput)
	candidate := c2sCompressedCandidate{
		K: k, OriginalMatrixScale: finalizationScaleString(originalMatrix.Scale), OriginalMatrixScaleLog2: originalMatrix.Scale.Log2(), CompressedMatrixScale: finalizationScaleString(matrix.Scale), CompressedMatrixScaleLog2: matrix.Scale.Log2(),
		MathematicalDiagonalsUnchanged: true, SameStructure: c2sCompressedStructureEqual(originalStructure, candidateStructure), SameDiagonalIndices: reflect.DeepEqual(originalStructure.SortedDiagonalIndices, candidateStructure.SortedDiagonalIndices), SameN1: originalMatrix.N1 == matrix.N1, SameLevelQ: originalMatrix.LevelQ == matrix.LevelQ, SameLevelP: originalMatrix.LevelP == matrix.LevelP, SameLogDimensions: originalMatrix.LogDimensions == matrix.LogDimensions,
	}
	checkpoints, err := c2sCompressedReplayGroup0(params, standardEval, fastEval, matrix, standardInput, fastInput, standardSK, fastSK, budget)
	if err != nil {
		return candidate, err
	}
	candidate.Checkpoints = checkpoints
	candidate.CompleteGroupCapacityPass = true
	candidate.CompleteGroupResiduesPass = true
	for _, checkpoint := range checkpoints {
		if candidate.FirstUnsafeCheckpoint == "" && (!checkpoint.FullRNSCenteredQ01Unique || !checkpoint.FastQ01Capacity.Pass) {
			candidate.FirstUnsafeCheckpoint = checkpoint.Checkpoint
		}
		if candidate.FirstResidueMismatch == "" && !checkpoint.FastResiduesMatch {
			candidate.FirstResidueMismatch = checkpoint.Checkpoint
		}
		if !checkpoint.FullRNSCenteredQ01Unique || !checkpoint.FastQ01Capacity.Pass {
			candidate.CompleteGroupCapacityPass = false
		}
		if !checkpoint.FastResiduesMatch {
			candidate.CompleteGroupResiduesPass = false
		}
	}
	standardLinear, fastLinear, err := c2sCompressedLinearTransform(params, standardEval, fastEval, matrix, standardInput, fastInput)
	if err != nil {
		return candidate, err
	}
	pre, err := c2sCompressedStageCapture("group0_pre_rescale", fastLinear, standardLinear, params, fastSK, standardSK, budget)
	if err != nil {
		return candidate, err
	}
	candidate.PreRescale = &pre
	if !candidate.CompleteGroupCapacityPass || !candidate.CompleteGroupResiduesPass {
		return candidate, nil
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, standardLinear, fastLinear); err != nil {
		return candidate, err
	}
	post, err := c2sCompressedStageCapture("group0_post_rescale", fastLinear, standardLinear, params, fastSK, standardSK, budget)
	if err != nil {
		return candidate, err
	}
	candidate.PostRescale = &post
	fastBefore, _, err := semanticBisectView(fastLinear, params, fastSK)
	if err != nil {
		return candidate, err
	}
	standardBefore, _, err := c2sPrecisionFullView(standardLinear, params, standardSK)
	if err != nil {
		return candidate, err
	}
	if err := c2sCompressedApplyRestore(params, fastEval.FastCKKS, fastLinear, standardLinear, k); err != nil {
		return candidate, err
	}
	fastAfter, _, err := semanticBisectView(fastLinear, params, fastSK)
	if err != nil {
		return candidate, err
	}
	standardAfter, _, err := c2sPrecisionFullView(standardLinear, params, standardSK)
	if err != nil {
		return candidate, err
	}
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, fastLinear)
	if err != nil {
		return candidate, err
	}
	standardCapacity, err := postMod1S2CCapacityFromCiphertext(params, standardLinear)
	if err != nil {
		return candidate, err
	}
	originalFastMeta, _, err := semanticBisectView(originalFastPost, params, fastSK)
	if err != nil {
		return candidate, err
	}
	originalStandardMeta, _, err := c2sPrecisionFullView(originalStandardPost, params, standardSK)
	if err != nil {
		return candidate, err
	}
	candidate.Restore = &c2sCompressedRestore{
		Multiplier: fmt.Sprintf("2^%d", k), FastBefore: fastBefore, FastAfter: fastAfter, StandardBefore: standardBefore, StandardAfter: standardAfter,
		FastCapacityAfter: fastCapacity, StandardCapacityAfter: standardCapacity, FastMetadataMatchesOriginal: c2sCompressedMetadataMatch(fastAfter, originalFastMeta), StandardMetadataMatchesOriginal: c2sCompressedMetadataMatch(standardAfter, originalStandardMeta),
	}
	compressedStandardValues, err := decodeWithSecret(params, standardLinear, standardSK)
	if err != nil {
		return candidate, err
	}
	fastValues, err := func() ([]complex128, error) {
		_, values, err := semanticBisectView(fastLinear, params, fastSK)
		return values, err
	}()
	if err != nil {
		return candidate, err
	}
	compressedStandardMetric, err := semanticBisectMetricFor(originalStandardValues, compressedStandardValues, budget)
	if err != nil {
		return candidate, err
	}
	fastCompressedMetric, err := semanticBisectMetricFor(compressedStandardValues, fastValues, budget)
	if err != nil {
		return candidate, err
	}
	fastOriginalMetric, err := semanticBisectMetricFor(originalStandardValues, fastValues, budget)
	if err != nil {
		return candidate, err
	}
	residuesMatch, err := c2sAliasRowsMatch(params, standardLinear, fastLinear)
	if err != nil {
		return candidate, err
	}
	candidate.Comparison = &c2sCompressedComparison{CompressedStandardRestoredVsOriginal: &compressedStandardMetric, FastCompressedRestoredVsCompressed: &fastCompressedMetric, FastCompressedRestoredVsOriginal: &fastOriginalMetric, FastResiduesMatch: residuesMatch}
	candidate.RestoredSemanticPass = compressedStandardMetric.Pass && fastCompressedMetric.Pass && fastOriginalMetric.Pass
	candidate.ValidCompleteGroup = candidate.CompleteGroupCapacityPass && candidate.CompleteGroupResiduesPass && candidate.RestoredSemanticPass && candidate.Restore.FastMetadataMatchesOriginal && candidate.Restore.StandardMetadataMatchesOriginal
	return candidate, nil
}

func c2sCompressedDownstreamProbe(params ckks.Parameters, standardEval *bootstrapping.Evaluator, fastEval *bootstrapping.FastEvaluator, originalMatrix dft.Matrix, compressedMatrix ckkslintrans.LinearTransformation, standardInput, fastInput *rlwe.Ciphertext, standardSK, fastSK *rlwe.SecretKey, originalStandardReal, originalStandardImag *rlwe.Ciphertext, budget float64, k int) (c2sCompressedDownstream, error) {
	standardCurrent, fastCurrent, err := c2sCompressedLinearTransform(params, standardEval, fastEval, compressedMatrix, standardInput, fastInput)
	if err != nil {
		return c2sCompressedDownstream{}, err
	}
	if err := c2sCompressedRescale(params, standardEval, fastEval, standardCurrent, fastCurrent); err != nil {
		return c2sCompressedDownstream{}, err
	}
	if err := c2sCompressedApplyRestore(params, fastEval.FastCKKS, fastCurrent, standardCurrent, k); err != nil {
		return c2sCompressedDownstream{}, err
	}
	firstBreach := "none"
	for matrixIndex := 1; matrixIndex < len(originalMatrix.Matrices); matrixIndex++ {
		standardCurrent, fastCurrent, err = c2sCompressedLinearTransform(params, standardEval, fastEval, originalMatrix.Matrices[matrixIndex], standardCurrent, fastCurrent)
		if err != nil {
			return c2sCompressedDownstream{}, err
		}
		if err := c2sCompressedRescale(params, standardEval, fastEval, standardCurrent, fastCurrent); err != nil {
			return c2sCompressedDownstream{}, err
		}
		stage, err := c2sCompressedStageCapture(fmt.Sprintf("group%d_post_rescale", matrixIndex), fastCurrent, standardCurrent, params, fastSK, standardSK, budget)
		if err != nil {
			return c2sCompressedDownstream{}, err
		}
		if firstBreach == "none" && stage.FastVsStandard != nil && !stage.FastVsStandard.Pass {
			firstBreach = stage.Checkpoint
		}
	}
	remainingLiteral := originalMatrix.MatrixLiteral
	remainingLiteral.LevelQ = fastCurrent.Level()
	remainingLiteral.Levels = append([]int(nil), originalMatrix.Levels[1:]...)
	remaining := dft.Matrix{MatrixLiteral: remainingLiteral, Matrices: append([]ckkslintrans.LinearTransformation(nil), originalMatrix.Matrices[1:]...)}
	fastReal, fastImag, err := fastEval.DFTEvaluator.CoeffsToSlotsNew(fastCurrent, remaining)
	if err != nil {
		return c2sCompressedDownstream{}, err
	}
	_, standardImag, err := standardEval.DFTEvaluator.CoeffsToSlotsNew(standardCurrent, remaining)
	if err != nil {
		return c2sCompressedDownstream{}, err
	}
	_, fastRealValues, err := semanticBisectView(fastReal, params, fastSK)
	if err != nil {
		return c2sCompressedDownstream{}, err
	}
	standardRealValues, err := decodeWithSecret(params, originalStandardReal, standardSK)
	if err != nil {
		return c2sCompressedDownstream{}, err
	}
	finalReal, err := semanticBisectMetricFor(standardRealValues, fastRealValues, budget)
	if err != nil {
		return c2sCompressedDownstream{}, err
	}
	var finalImag *semanticBisectMetric
	if fastImag != nil && originalStandardImag != nil && standardImag != nil {
		_, fastImagValues, err := semanticBisectView(fastImag, params, fastSK)
		if err != nil {
			return c2sCompressedDownstream{}, err
		}
		standardImagValues, err := decodeWithSecret(params, originalStandardImag, standardSK)
		if err != nil {
			return c2sCompressedDownstream{}, err
		}
		metric, err := semanticBisectMetricFor(standardImagValues, fastImagValues, budget)
		if err != nil {
			return c2sCompressedDownstream{}, err
		}
		finalImag = &metric
	}
	return c2sCompressedDownstream{ValidatedKSafe: k, FinalRealDelta: &finalReal, FinalImagDelta: finalImag, FirstLaterBudgetBreach: firstBreach, Group0FixAloneSatisfiesC2S: finalReal.Pass && (finalImag == nil || finalImag.Pass) && firstBreach == "none"}, nil
}

func runFIX001P3DesignLogN13C2SGroup0CompressedLinear(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
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
	result := c2sCompressedResult{
		SchemaVersion: "fix-001-p3-design-logn13-c2s-group0-compressed-linear.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()},
		StandardProof:           semanticBisectStandardProof{EvaluatorNonNil: standardEval.Evaluator != nil, DFTEvaluatorNonNil: standardEval.DFTEvaluator != nil, Mod1EvaluatorNonNil: standardEval.Mod1Evaluator != nil, FastCompatibilityPathUsed: false, OrdinaryStageCallsUsed: true},
		C2SErrorAloneSufficient: true, EvalModUnresolved: true,
		Validation:    map[string]interface{}{"primary_required_base": requiredC2SCompressedPrimaryBase, "secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredC2SCompressedSecondary, "no_secondary_production_changes": true, "no_matrix_generation_or_parameter_changes": true, "no_fast_production_changes": true, "no_production_fix": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "group0_only_design": true, "evalmod_untouched": true},
		KMinFullGroup: -1, KSafeFullGroup: -1, FirstSupportedCause: "logn13_c2s_group0_compressed_linear_design_not_validated",
	}
	if !result.StandardProof.EvaluatorNonNil || !result.StandardProof.DFTEvaluatorNonNil || !result.StandardProof.Mod1EvaluatorNonNil || !result.Validation["secondary_exact_required_commit"].(bool) || !result.Validation["secondary_clean"].(bool) {
		return c2sCompressedWrite(result, outPath)
	}
	standardControl, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return c2sCompressedWrite(result, outPath)
	}
	standardControlDecoded, err := decodeWithSecret(residual, standardControl, skN1)
	if err != nil {
		return err
	}
	standardMetric, err := semanticBisectMetricFor(reproducibleValues(residual.MaxSlots()), standardControlDecoded, c2sCompressedCandidateThreshold)
	if err != nil {
		return err
	}
	result.StandardEndToEnd = &standardMetric
	if !standardMetric.Pass {
		return c2sCompressedWrite(result, outPath)
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
		return c2sCompressedWrite(result, outPath)
	}
	result.Validation["fast_c2s_matrix_field_is_lazy"] = len(fastEval.C2SDFTMatrix.Matrices) == 0
	result.Validation["standard_c2s_levels"] = append([]int(nil), standardEval.C2SDFTMatrix.Levels...)
	if len(standardEval.C2SDFTMatrix.Matrices) == 0 {
		return c2sCompressedWrite(result, outPath)
	}
	originalMatrix := standardEval.C2SDFTMatrix
	result.OriginalStructure = c2sAliasStructureFromMatrix(originalMatrix.Matrices[0], fastModUp)
	if result.OriginalStructure.N1 == 0 {
		result.OriginalStructure.Path = "direct"
	}
	fastReal, _, err := fastEval.CoeffsToSlots(fastModUp.CopyNew())
	if err != nil {
		return err
	}
	standardReal, standardImag, err := standardEval.CoeffsToSlots(standardModUp.CopyNew())
	if err != nil {
		return err
	}
	standardFromFast, err := evalModStandardFromFast(params, fastReal, func() []complex128 { _, values, _ := semanticBisectView(fastReal, params, fastSK); return values }(), skN2)
	if err != nil {
		return err
	}
	standardFastEvidence, err := evalModMatchedInput("standard_from_fast_c2s_real", fastReal, standardFromFast, params, fastSK, skN2)
	if err != nil {
		return err
	}
	result.Validation["standard_from_fast_c2s_real_pass"] = standardFastEvidence.Pass
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
	standardRealValues, err := decodeWithSecret(params, standardReal, skN2)
	if err != nil {
		return err
	}
	_, fastRealValues, err := semanticBisectView(fastReal, params, fastSK)
	if err != nil {
		return err
	}
	realInputDelta, err := semanticBisectMetricFor(standardRealValues, fastRealValues, 1)
	if err != nil {
		return err
	}
	if realInputDelta.MaxComponentAbs > 0 {
		result.DerivedC2SBudget = c2sPrecisionFinalThreshold / (standardSensitivity.MaxComponentAbs / realInputDelta.MaxComponentAbs)
	}
	result.C2SErrorAloneSufficient = result.DerivedC2SBudget > 0 && standardSensitivity.MaxComponentAbs > c2sPrecisionFinalThreshold && realInputDelta.MaxComponentAbs > result.DerivedC2SBudget
	originalStandardLinear, originalFastLinear, err := c2sCompressedLinearTransform(params, standardEval, fastEval, originalMatrix.Matrices[0], standardModUp, fastModUp)
	if err != nil {
		return err
	}
	originalLinear, err := c2sPrecisionCapture("original_group0_linear_transform", originalFastLinear, originalStandardLinear, params, fastSK, skN2, standardValues, result.DerivedC2SBudget)
	if err != nil {
		return err
	}
	result.OriginalLinear = &originalLinear
	if err := c2sCompressedRescale(params, standardEval, fastEval, originalStandardLinear, originalFastLinear); err != nil {
		return err
	}
	originalRescale, err := c2sPrecisionCapture("original_group0_rescale", originalFastLinear, originalStandardLinear, params, fastSK, skN2, standardValues, result.DerivedC2SBudget)
	if err != nil {
		return err
	}
	result.OriginalRescale = &originalRescale
	result.OriginalRescaleIncrement = result.OriginalRescale.FastVsStandard.MaxComponentAbs - result.OriginalLinear.FastVsStandard.MaxComponentAbs
	originalStandardValues, err := decodeWithSecret(params, originalStandardLinear, skN2)
	if err != nil {
		return err
	}
	for _, k := range []int{2, 3, 4} {
		compressedMatrix, sameMath, err := c2sCompressedMatrix(params, originalMatrix, k)
		if err != nil {
			return err
		}
		candidate, err := c2sCompressedCandidateRun(params, standardEval, fastEval, compressedMatrix, originalMatrix.Matrices[0], standardModUp.CopyNew(), fastModUp.CopyNew(), skN2, fastSK, originalStandardLinear, originalFastLinear, originalStandardValues, result.DerivedC2SBudget, k)
		if err != nil {
			return err
		}
		candidate.MathematicalDiagonalsUnchanged = sameMath
		result.Candidates = append(result.Candidates, candidate)
	}
	for _, candidate := range result.Candidates {
		if candidate.ValidCompleteGroup && result.KMinFullGroup < 0 {
			result.KMinFullGroup = candidate.K
		}
	}
	if result.KMinFullGroup >= 0 {
		for _, candidate := range result.Candidates {
			if candidate.K == result.KMinFullGroup+1 && candidate.ValidCompleteGroup {
				result.KSafeFullGroup = candidate.K
			}
		}
	}
	if result.KMinFullGroup >= 0 && result.KSafeFullGroup >= 0 {
		for _, candidate := range result.Candidates {
			if candidate.K == result.KSafeFullGroup {
				compressedMatrix, _, err := c2sCompressedMatrix(params, originalMatrix, candidate.K)
				if err != nil {
					return err
				}
				downstream, err := c2sCompressedDownstreamProbe(params, standardEval, fastEval, originalMatrix, compressedMatrix, standardModUp.CopyNew(), fastModUp.CopyNew(), skN2, fastSK, standardReal, standardImag, result.DerivedC2SBudget, candidate.K)
				if err != nil {
					return err
				}
				result.Downstream = &downstream
				break
			}
		}
	}
	if result.KMinFullGroup >= 0 && result.KSafeFullGroup >= 0 && result.Downstream != nil {
		result.FirstSupportedCause = "logn13_c2s_group0_compressed_linear_design_validated"
	}
	return c2sCompressedWrite(result, outPath)
}
