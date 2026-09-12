package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	commonlintrans "github.com/tuneinsight/lattigo/v6/circuits/common/lintrans"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const requiredPostMod1SecondaryCommit = "ed8b19fc1b51fdf37383f9e196f3a5bed2c03219"

type postMod1S2CCapacity struct {
	MaxAbsExactCoefficient string  `json:"max_abs_exact_coefficient"`
	Q01Half                string  `json:"q01_half"`
	MaxAbsOverQ01Half      float64 `json:"max_abs_over_q01_half"`
	OutsideCount           int     `json:"outside_count"`
	Pass                   bool    `json:"pass_unique_centered_reconstruction"`
}

type postMod1S2CGroupCheckpoint struct {
	Group                   int                          `json:"group"`
	MatrixIndices           []int                        `json:"matrix_indices"`
	InputLevel              int                          `json:"input_level"`
	InputScale              string                       `json:"input_scale"`
	FastPreRescaleRows      []FinalizationRowFingerprint `json:"fast_pre_rescale_rows_normalized"`
	ReferencePreRescaleRows []FinalizationRowFingerprint `json:"reference_pre_rescale_rows"`
	PreRescaleRowsMatch     bool                         `json:"pre_rescale_rows_match"`
	Capacity                postMod1S2CCapacity          `json:"capacity"`
	FastPostLevel           int                          `json:"fast_post_rescale_level"`
	FastPostScale           string                       `json:"fast_post_rescale_scale"`
	FastPostRescaleRows     []FinalizationRowFingerprint `json:"fast_post_rescale_rows_normalized"`
	ReferencePostLevel      int                          `json:"reference_post_rescale_level"`
	ReferencePostScale      string                       `json:"reference_post_rescale_scale"`
	ReferencePostRows       []FinalizationRowFingerprint `json:"reference_post_rescale_rows"`
	PostRescaleRowsMatch    bool                         `json:"post_rescale_rows_match"`
}

type postMod1S2CFailure struct {
	Cause      string
	Group      int
	Checkpoint string
}

func (e *postMod1S2CFailure) Error() string {
	return fmt.Sprintf("%s at group=%d checkpoint=%s", e.Cause, e.Group, e.Checkpoint)
}

type postMod1S2CResult struct {
	SchemaVersion                string                          `json:"schema_version"`
	Timestamp                    time.Time                       `json:"timestamp"`
	Primary                      RepositoryMetadata              `json:"primary_repository"`
	Lattigo                      RepositoryMetadata              `json:"lattigo_repository"`
	Environment                  EnvironmentMetadata             `json:"environment"`
	Config                       BootstrapConfig                 `json:"config"`
	Parameters                   ExperimentParameters            `json:"effective_parameters"`
	Workload                     CorrectnessWorkload             `json:"workload"`
	ProductionPrefix             string                          `json:"production_prefix"`
	EvalModReal                  FinalizationCiphertextEvidence  `json:"production_eval_mod_real"`
	EvalModImag                  *FinalizationCiphertextEvidence `json:"production_eval_mod_imag,omitempty"`
	ImagPresent                  bool                            `json:"ct_imag_present"`
	AcceptedMod1Rows             []FinalizationRowFingerprint    `json:"accepted_mod1_final_rows"`
	Mod1RowsMatch                bool                            `json:"accepted_mod1_rows_match"`
	S2CMatrixLevelQ              int                             `json:"s2c_matrix_level_q"`
	S2CMatrixLevels              []int                           `json:"s2c_matrix_levels"`
	S2CMatrixCount               int                             `json:"s2c_matrix_count"`
	S2CInputReal                 FinalizationCiphertextEvidence  `json:"s2c_input_real"`
	S2CInputImag                 *FinalizationCiphertextEvidence `json:"s2c_input_imag,omitempty"`
	InputCapacityReal            postMod1S2CCapacity             `json:"s2c_input_capacity_real"`
	InputCapacityImag            *postMod1S2CCapacity            `json:"s2c_input_capacity_imag,omitempty"`
	ProductionOutput             *FinalizationCiphertextEvidence `json:"production_s2c_output,omitempty"`
	ReferenceOutput              *FinalizationCiphertextEvidence `json:"standard_full_rns_s2c_output,omitempty"`
	FastReplayMatchesProduction  bool                            `json:"fast_replay_matches_production"`
	ReferenceReplayMatchesPublic bool                            `json:"reference_replay_matches_standard_public_output"`
	ReplayMatchesPublic          bool                            `json:"reference_replay_matches_standard_public"`
	Groups                       []postMod1S2CGroupCheckpoint    `json:"groups"`
	FinalRowsMatch               bool                            `json:"s2c_final_rows_match"`
	FinalMetadataMatch           bool                            `json:"s2c_final_metadata_match"`
	Semantic                     *PSGlobalMetric                 `json:"s2c_semantic"`
	InputUnchanged               bool                            `json:"input_unchanged"`
	FirstFailingGroup            string                          `json:"first_failing_group"`
	FirstFailingPoint            string                          `json:"first_failing_checkpoint"`
	FirstSupportedCause          string                          `json:"first_supported_cause"`
	Validation                   map[string]interface{}          `json:"validation"`
}

func postMod1S2CCapacityFromValues(params ckks.Parameters, values []*big.Int) postMod1S2CCapacity {
	q01 := new(big.Int).Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus))
	q01Half := new(big.Int).Rsh(new(big.Int).Set(q01), 1)
	maxAbs := big.NewInt(0)
	outside := 0
	for _, value := range values {
		abs := new(big.Int).Abs(value)
		if abs.Cmp(maxAbs) > 0 {
			maxAbs.Set(abs)
		}
		if abs.Cmp(q01Half) >= 0 {
			outside++
		}
	}
	ratio, _ := new(big.Float).Quo(new(big.Float).SetInt(maxAbs), new(big.Float).SetInt(q01Half)).Float64()
	return postMod1S2CCapacity{MaxAbsExactCoefficient: maxAbs.String(), Q01Half: q01Half.String(), MaxAbsOverQ01Half: ratio, OutsideCount: outside, Pass: outside == 0}
}

func postMod1S2CRecoverQ01(params ckks.Parameters, source *rlwe.Ciphertext) ([][]*big.Int, error) {
	if source.Level() < 1 || !source.IsNTT {
		return nil, fmt.Errorf("q0/q1 lift requires NTT level >= 1")
	}
	ringQ := params.RingQ().AtLevel(source.Level())
	backend := fastckks.NewRNSBackend(params.RingQ().AtLevel(1), 1)
	recovered := make([][]*big.Int, source.Degree()+1)
	for component := range source.Value {
		coeff := ring.NewPoly(params.N(), 1)
		if err := fastckks.FastPartialINTT(ringQ, source.Value[component], coeff); err != nil {
			return nil, err
		}
		if source.IsMontgomery {
			for limb := 0; limb < 2; limb++ {
				params.RingQ().SubRings[limb].IMForm(coeff.Coeffs[limb], coeff.Coeffs[limb])
			}
		}
		signed, err := backend.ReconstructQ0Q1(&coeff)
		if err != nil {
			return nil, err
		}
		recovered[component] = signed
	}
	return recovered, nil
}

func postMod1S2CLiftSigned(params ckks.Parameters, signed []*big.Int, level int) ring.Poly {
	ringQ := params.RingQ().AtLevel(level)
	coeff := ring.NewPoly(params.N(), level)
	for limb := 0; limb <= level; limb++ {
		modulus := new(big.Int).SetUint64(ringQ.SubRings[limb].Modulus)
		for i, value := range signed {
			coeff.Coeffs[limb][i] = new(big.Int).Mod(new(big.Int).Set(value), modulus).Uint64()
		}
	}
	normalNTT := ring.NewPoly(params.N(), level)
	ringQ.NTT(coeff, normalNTT)
	return normalNTT
}

func postMod1S2CFullCRT(params ckks.Parameters, coeff ring.Poly, level int) []*big.Int {
	ringQ := params.RingQ().AtLevel(level)
	values := make([]*big.Int, params.N())
	for i := range values {
		value := new(big.Int).SetUint64(coeff.Coeffs[0][i])
		currentModulus := new(big.Int).SetUint64(ringQ.SubRings[0].Modulus)
		for limb := 1; limb <= level; limb++ {
			qi := new(big.Int).SetUint64(ringQ.SubRings[limb].Modulus)
			residue := new(big.Int).SetUint64(coeff.Coeffs[limb][i])
			delta := new(big.Int).Sub(residue, new(big.Int).Mod(new(big.Int).Set(value), qi))
			delta.Mod(delta, qi)
			inverse := new(big.Int).ModInverse(new(big.Int).Mod(new(big.Int).Set(currentModulus), qi), qi)
			delta.Mul(delta, inverse).Mod(delta, qi)
			value.Add(value, new(big.Int).Mul(currentModulus, delta))
			currentModulus.Mul(currentModulus, qi)
		}
		half := new(big.Int).Rsh(new(big.Int).Set(currentModulus), 1)
		if value.Cmp(half) > 0 {
			value.Sub(value, currentModulus)
		}
		values[i] = value
	}
	return values
}

func postMod1S2CCapacityFromCiphertext(params ckks.Parameters, ct *rlwe.Ciphertext) (postMod1S2CCapacity, error) {
	if ct == nil || ct.Level() < 1 || !ct.IsNTT {
		return postMod1S2CCapacity{}, fmt.Errorf("capacity reference requires an NTT ciphertext at level >= 1")
	}
	level := ct.Level()
	ringQ := params.RingQ().AtLevel(level)
	values := make([]*big.Int, 0, params.N()*(ct.Degree()+1))
	for component := 0; component <= ct.Degree(); component++ {
		poly := ring.NewPoly(params.N(), level)
		for limb := 0; limb <= level; limb++ {
			copy(poly.Coeffs[limb], ct.Value[component].Coeffs[limb])
			if ct.IsMontgomery {
				params.RingQ().SubRings[limb].IMForm(poly.Coeffs[limb], poly.Coeffs[limb])
			}
		}
		coeff := ring.NewPoly(params.N(), level)
		ringQ.INTT(poly, coeff)
		full := postMod1S2CFullCRT(params, coeff, level)
		values = append(values, full...)
	}
	return postMod1S2CCapacityFromValues(params, values), nil
}

func postMod1S2CCapacityFromFastCiphertext(params ckks.Parameters, source *rlwe.Ciphertext) (postMod1S2CCapacity, error) {
	recovered, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return postMod1S2CCapacity{}, err
	}
	values := make([]*big.Int, 0, params.N()*(source.Degree()+1))
	for _, component := range recovered {
		values = append(values, component...)
	}
	return postMod1S2CCapacityFromValues(params, values), nil
}

func postMod1S2CLiftFull(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, postMod1S2CCapacity, error) {
	recovered, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, postMod1S2CCapacity{}, err
	}
	values := make([]*big.Int, 0, params.N()*(source.Degree()+1))
	full := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*full.MetaData = *source.MetaData
	full.IsNTT = true
	full.IsMontgomery = false
	for component, signed := range recovered {
		values = append(values, signed...)
		normalNTT := postMod1S2CLiftSigned(params, signed, source.Level())
		for limb := 0; limb <= source.Level(); limb++ {
			copy(full.Value[component].Coeffs[limb], normalNTT.Coeffs[limb])
		}
	}
	return full, postMod1S2CCapacityFromValues(params, values), nil
}

func postMod1S2CFastNormalizedCopy(params ckks.Parameters, source *rlwe.Ciphertext) *rlwe.Ciphertext {
	copyCT := source.CopyNew()
	if source.IsMontgomery {
		for component := range copyCT.Value {
			for limb := 0; limb < 2; limb++ {
				params.RingQ().SubRings[limb].IMForm(copyCT.Value[component].Coeffs[limb], copyCT.Value[component].Coeffs[limb])
			}
		}
		copyCT.IsMontgomery = false
	}
	return copyCT
}

func postMod1S2CRowsEqual(params ckks.Parameters, fast, reference *rlwe.Ciphertext) bool {
	fastNormal := postMod1S2CFastNormalizedCopy(params, fast)
	if fastNormal.Level() != reference.Level() || fastNormal.Degree() != reference.Degree() {
		return false
	}
	for component := 0; component <= fastNormal.Degree(); component++ {
		for limb := 0; limb < 2; limb++ {
			if !finalizationRowComparison(component, limb, fastNormal.Value[component].Coeffs[limb], reference.Value[component].Coeffs[limb]).Equal {
				return false
			}
		}
	}
	return true
}

func postMod1S2CPreparedFastCopy(source *rlwe.Ciphertext) *rlwe.Ciphertext {
	if source == nil {
		return nil
	}
	return source.CopyNew()
}

func postMod1S2CPreparedFullCopy(source *rlwe.Ciphertext) *rlwe.Ciphertext {
	if source == nil {
		return nil
	}
	return source.CopyNew()
}

func postMod1S2CCopyMetadata(dst, source *rlwe.Ciphertext) {
	*dst.MetaData = *source.MetaData
	dst.IsNTT = source.IsNTT
	dst.IsMontgomery = source.IsMontgomery
	dst.IsBatched = source.IsBatched
	dst.IsBitReversed = source.IsBitReversed
	dst.Scale = source.Scale
}

func newPostMod1StandardDFT(params ckks.Parameters, matrix dft.Matrix) (*dft.Evaluator, error) {
	kgen := rlwe.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyNew()
	galEls := matrix.GaloisElements(params)
	keys := rlwe.NewMemEvaluationKeySet(nil, kgen.GenGaloisKeysNew(galEls, sk)...)
	return dft.NewEvaluator(params, ckks.NewEvaluator(params, keys)), nil
}

func postMod1S2CReplay(state chebyshevOracleState, fastReal, fastImag, fullReal, fullImag *rlwe.Ciphertext, standardDFT *dft.Evaluator) (*rlwe.Ciphertext, *rlwe.Ciphertext, []postMod1S2CGroupCheckpoint, *postMod1S2CFailure, error) {
	params := state.Params.BootstrappingParameters
	fastEval := state.Eval.DFTEvaluator.FastEvaluator()
	fastCurrent := postMod1S2CPreparedFastCopy(fastReal)
	fullCurrent := postMod1S2CPreparedFullCopy(fullReal)
	if fastImag != nil {
		fastCombined := fastckks.NewCiphertext(params, 1, fastReal.Level())
		postMod1S2CCopyMetadata(fastCombined, fastReal)
		if err := fastEval.Mul(fastImag, 1i, fastCombined); err != nil {
			return nil, nil, nil, nil, err
		}
		if err := fastEval.Add(fastCombined, fastReal, fastCombined); err != nil {
			return nil, nil, nil, nil, err
		}
		fastCurrent = fastCombined

		fullCombined := ckks.NewCiphertext(params, 1, fullReal.Level())
		postMod1S2CCopyMetadata(fullCombined, fullReal)
		if err := standardDFT.Evaluator.Mul(fullImag, 1i, fullCombined); err != nil {
			return nil, nil, nil, nil, err
		}
		if err := standardDFT.Evaluator.Add(fullCombined, fullReal, fullCombined); err != nil {
			return nil, nil, nil, nil, err
		}
		fullCurrent = fullCombined
	}

	groups := make([]postMod1S2CGroupCheckpoint, 0, len(state.Eval.S2CDFTMatrix.Levels))
	matrixIndex := 0
	for groupIndex, factorCount := range state.Eval.S2CDFTMatrix.Levels {
		checkpoint := postMod1S2CGroupCheckpoint{Group: groupIndex, MatrixIndices: make([]int, 0, factorCount), InputLevel: fastCurrent.Level(), InputScale: finalizationScaleString(fastCurrent.Scale)}
		for range factorCount {
			if matrixIndex >= len(state.Eval.S2CDFTMatrix.Matrices) {
				return nil, nil, groups, nil, fmt.Errorf("S2C matrix factorization is shorter than its level schedule")
			}
			checkpoint.MatrixIndices = append(checkpoint.MatrixIndices, matrixIndex)
			matrix := state.Eval.S2CDFTMatrix.Matrices[matrixIndex]
			fastNext := fastckks.NewCiphertext(params, 1, fastCurrent.Level())
			postMod1S2CCopyMetadata(fastNext, fastCurrent)
			if err := fastEval.LinearTransform(fastCurrent, commonlintrans.LinearTransformation(matrix), fastNext); err != nil {
				return nil, nil, groups, nil, err
			}
			fullNext := ckks.NewCiphertext(params, 1, fullCurrent.Level())
			postMod1S2CCopyMetadata(fullNext, fullCurrent)
			if err := standardDFT.LTEvaluator.Evaluate(fullCurrent, matrix, fullNext); err != nil {
				return nil, nil, groups, nil, err
			}
			fastCurrent, fullCurrent = fastNext, fullNext
			matrixIndex++
		}

		checkpoint.FastPreRescaleRows = finalizationEvidence(postMod1S2CFastNormalizedCopy(params, fastCurrent)).Rows
		checkpoint.ReferencePreRescaleRows = finalizationEvidence(fullCurrent).Rows
		checkpoint.PreRescaleRowsMatch = postMod1S2CRowsEqual(params, fastCurrent, fullCurrent)
		checkpoint.Capacity, _ = postMod1S2CCapacityFromCiphertext(params, fullCurrent)
		if !checkpoint.Capacity.Pass {
			groups = append(groups, checkpoint)
			return nil, nil, groups, &postMod1S2CFailure{Cause: "post_mod1_s2c_group_capacity_failure", Group: groupIndex, Checkpoint: "pre_rescale_capacity"}, nil
		}
		if !checkpoint.PreRescaleRowsMatch {
			groups = append(groups, checkpoint)
			return nil, nil, groups, &postMod1S2CFailure{Cause: "post_mod1_s2c_linear_transform_mismatch", Group: groupIndex, Checkpoint: "pre_rescale_rows"}, nil
		}
		if err := fastEval.Rescale(fastCurrent, fastCurrent); err != nil {
			return nil, nil, groups, nil, err
		}
		if err := standardDFT.Evaluator.Rescale(fullCurrent, fullCurrent); err != nil {
			return nil, nil, groups, nil, err
		}
		checkpoint.FastPostLevel = fastCurrent.Level()
		checkpoint.FastPostScale = finalizationScaleString(fastCurrent.Scale)
		checkpoint.FastPostRescaleRows = finalizationEvidence(postMod1S2CFastNormalizedCopy(params, fastCurrent)).Rows
		checkpoint.ReferencePostLevel = fullCurrent.Level()
		checkpoint.ReferencePostScale = finalizationScaleString(fullCurrent.Scale)
		checkpoint.ReferencePostRows = finalizationEvidence(fullCurrent).Rows
		checkpoint.PostRescaleRowsMatch = postMod1S2CRowsEqual(params, fastCurrent, fullCurrent)
		groups = append(groups, checkpoint)
		if !checkpoint.PostRescaleRowsMatch {
			return nil, nil, groups, &postMod1S2CFailure{Cause: "post_mod1_s2c_rescale_mismatch", Group: groupIndex, Checkpoint: "post_rescale_rows"}, nil
		}
	}
	if matrixIndex != len(state.Eval.S2CDFTMatrix.Matrices) {
		return nil, nil, groups, nil, fmt.Errorf("S2C matrix factorization has unused matrices")
	}
	return fastCurrent, fullCurrent, groups, nil, nil
}

func postMod1S2CWrite(result postMod1S2CResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for post-Mod1 S2C diagnostic")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DiagLogN13PostMod1S2C(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	if state.Mod1Input == nil || state.Eval.DFTEvaluator == nil || state.Eval.Mod1Evaluator == nil {
		return fmt.Errorf("post_mod1_s2c_precondition_mismatch: production prefix was not initialized")
	}
	result := postMod1S2CResult{
		SchemaVersion: "fix-001-p3-diag-logn13-post-mod1-s2c.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Parameters: parameterMetadata(state.Params.ResidualParameters, state.Params), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + production EvalMod", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32", LogicalSlots: 4096},
		ProductionPrefix: "ScaleDown -> ModUp -> CoeffsToSlots -> EvalMod -> production SlotsToCoeffs",
		S2CMatrixLevelQ:  state.Eval.S2CDFTMatrix.LevelQ, S2CMatrixLevels: append([]int(nil), state.Eval.S2CDFTMatrix.Levels...), S2CMatrixCount: len(state.Eval.S2CDFTMatrix.Matrices),
		FirstFailingGroup: "none", FirstFailingPoint: "none", FirstSupportedCause: "post_mod1_s2c_precondition_mismatch",
		Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "secondary_exact_required_commit": gitOutput(backendRoot, "rev-parse", "HEAD") == requiredPostMod1SecondaryCommit, "no_secondary_production_changes": true, "no_unpack_or_finalization": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
	}
	acceptedRows, err := loadAcceptedNormalizedFinalRows(primaryRoot)
	if err != nil {
		return err
	}
	ctReal, err := state.Eval.EvalMod(state.Mod1Input.CopyNew())
	if err != nil {
		result.FirstSupportedCause = "post_mod1_s2c_precondition_mismatch"
		return postMod1S2CWrite(result, outPath)
	}
	var ctImag *rlwe.Ciphertext
	if state.Mod1ImagInput != nil {
		ctImag, err = state.Eval.EvalMod(state.Mod1ImagInput.CopyNew())
		if err != nil {
			result.FirstSupportedCause = "post_mod1_s2c_precondition_mismatch"
			return postMod1S2CWrite(result, outPath)
		}
	}
	result.EvalModReal = finalizationEvidence(ctReal)
	if ctImag != nil {
		evidence := finalizationEvidence(ctImag)
		result.EvalModImag = &evidence
	}
	result.ImagPresent = ctImag != nil
	result.AcceptedMod1Rows = acceptedRows
	result.Mod1RowsMatch = doubleAngleRowsMatch(result.EvalModReal.Rows, acceptedRows)
	result.S2CInputReal = result.EvalModReal
	if ctImag != nil {
		imagEvidence := finalizationEvidence(ctImag)
		result.S2CInputImag = &imagEvidence
	}
	result.InputCapacityReal, err = postMod1S2CCapacityFromFastCiphertext(state.Params.BootstrappingParameters, ctReal)
	if err != nil {
		return err
	}
	if ctImag != nil {
		imagCapacity, capacityErr := postMod1S2CCapacityFromFastCiphertext(state.Params.BootstrappingParameters, ctImag)
		if capacityErr != nil {
			return capacityErr
		}
		result.InputCapacityImag = &imagCapacity
	}
	if !result.InputCapacityReal.Pass || (result.InputCapacityImag != nil && !result.InputCapacityImag.Pass) {
		result.FirstSupportedCause = "post_mod1_s2c_input_capacity_unresolved"
		return postMod1S2CWrite(result, outPath)
	}
	if !result.Mod1RowsMatch || ctReal.Level() != 4 || ctReal.Degree() != 1 || !ctReal.IsNTT || !ctReal.IsMontgomery {
		return postMod1S2CWrite(result, outPath)
	}

	params := state.Params.BootstrappingParameters
	fullReal, _, err := postMod1S2CLiftFull(params, ctReal)
	if err != nil {
		return err
	}
	var fullImag *rlwe.Ciphertext
	if ctImag != nil {
		fullImag, _, err = postMod1S2CLiftFull(params, ctImag)
		if err != nil {
			return err
		}
	}
	standardDFT, err := newPostMod1StandardDFT(params, state.Eval.S2CDFTMatrix)
	if err != nil {
		return err
	}
	productionRealBefore := ctReal.CopyNew()
	productionImagBefore := postMod1S2CPreparedFastCopy(ctImag)
	productionInputReal := postMod1S2CPreparedFastCopy(ctReal)
	productionInputImag := postMod1S2CPreparedFastCopy(ctImag)
	productionOutput, err := state.Eval.DFTEvaluator.SlotsToCoeffsNew(productionInputReal, productionInputImag, state.Eval.S2CDFTMatrix)
	if err != nil {
		result.FirstSupportedCause = "post_mod1_s2c_production_call_failure"
		return postMod1S2CWrite(result, outPath)
	}
	productionEvidence := finalizationEvidence(productionOutput)
	result.ProductionOutput = &productionEvidence
	result.InputUnchanged = integratedInputRowsUnchanged(productionRealBefore, productionInputReal)
	if ctImag != nil {
		result.InputUnchanged = result.InputUnchanged && integratedInputRowsUnchanged(productionImagBefore, productionInputImag)
	}
	standardOutput, err := standardDFT.SlotsToCoeffsNew(postMod1S2CPreparedFullCopy(fullReal), postMod1S2CPreparedFullCopy(fullImag), state.Eval.S2CDFTMatrix)
	if err != nil {
		return err
	}
	standardEvidence := finalizationEvidence(standardOutput)
	result.ReferenceOutput = &standardEvidence
	fastReplay, referenceReplay, groups, failure, err := postMod1S2CReplay(state, ctReal, ctImag, fullReal, fullImag, standardDFT)
	if err != nil {
		return err
	}
	result.Groups = groups
	if failure != nil {
		result.FirstSupportedCause = failure.Cause
		result.FirstFailingGroup = fmt.Sprintf("%d", failure.Group)
		result.FirstFailingPoint = failure.Checkpoint
		return postMod1S2CWrite(result, outPath)
	}
	result.FastReplayMatchesProduction = postMod1S2CRowsEqual(params, fastReplay, productionOutput)
	result.ReferenceReplayMatchesPublic = postMod1S2CRowsEqual(params, referenceReplay, standardOutput)
	result.ReplayMatchesPublic = result.FastReplayMatchesProduction && result.ReferenceReplayMatchesPublic
	result.FinalRowsMatch = postMod1S2CRowsEqual(params, productionOutput, standardOutput)
	result.FinalMetadataMatch = productionOutput.Level() == standardOutput.Level() && productionOutput.Degree() == standardOutput.Degree() && productionOutput.Scale.Equal(standardOutput.Scale) && productionOutput.LogDimensions == standardOutput.LogDimensions
	productionDecoded, err := psGlobalDecode(params, productionOutput)
	if err != nil {
		return err
	}
	referenceDecoded, err := decodeWithSecret(params, standardOutput, zeroSecret(params))
	if err != nil {
		return err
	}
	result.Semantic = psGlobalMetric(referenceDecoded, productionDecoded)
	if productionOutput.Level() != 1 || productionOutput.Degree() != 1 || !productionOutput.IsNTT || !productionOutput.IsMontgomery {
		result.FirstSupportedCause = "post_mod1_s2c_final_metadata_mismatch"
	} else if !result.FinalMetadataMatch || !result.FinalRowsMatch {
		result.FirstSupportedCause = "post_mod1_s2c_final_metadata_mismatch"
	} else if result.Semantic == nil || !result.Semantic.Pass {
		result.FirstSupportedCause = "post_mod1_s2c_semantic_failure"
	} else if !result.InputUnchanged {
		result.FirstSupportedCause = "post_mod1_s2c_final_metadata_mismatch"
	} else {
		result.FirstSupportedCause = "logn13_post_mod1_s2c_core_output_validated"
	}
	return postMod1S2CWrite(result, outPath)
}
