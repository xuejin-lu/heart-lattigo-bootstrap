package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3T2GuardPrimary   = "490d3f3e6402f52b8c31d62a51b3e4e862954fc3"
	requiredFIX001P3T2GuardSecondary = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
	t2GuardPublicThreshold           = 1e-2
	t2GuardSemanticThreshold         = 1e-10
)

type t2GuardCapacityEvidence struct {
	Fast                     postMod1S2CCapacity `json:"fast_q01_capacity"`
	FullRNS                  postMod1S2CCapacity `json:"full_rns_capacity"`
	IndependentQ01           postMod1S2CCapacity `json:"independent_q01_capacity"`
	FastVsFullRowsMatch      bool                `json:"fast_vs_full_rns_rows_match"`
	IndependentRecoveryMatch bool                `json:"independent_q01_recovery_match"`
	CenteredUnique           bool                `json:"centered_unique"`
}

type t2GuardState struct {
	State    b4BabyState             `json:"state"`
	Capacity t2GuardCapacityEvidence `json:"capacity"`
}

type t2GuardOperation struct {
	Operation       b4BabyOperation `json:"operation"`
	PreState        t2GuardState    `json:"pre_state"`
	PostState       t2GuardState    `json:"post_state"`
	ContractedState t2GuardState    `json:"contracted_state"`
	ValidContracts  bool            `json:"valid_contracts"`
}

type t2GuardContraction struct {
	EvenCount                   int                 `json:"even_centered_coefficient_count"`
	OddCount                    int                 `json:"odd_centered_coefficient_count"`
	OddFraction                 float64             `json:"odd_centered_coefficient_fraction"`
	MaxAbsRoundingErrorInteger  int                 `json:"max_abs_rounding_error_integer_units"`
	IndependentRecoveryMatch    bool                `json:"independent_q01_reconstruction_match"`
	RoundedDivideOracleMatch    bool                `json:"rounded_divide_oracle_match"`
	OutputMetadataMatchesNative bool                `json:"output_metadata_matches_native"`
	OutputCapacity              postMod1S2CCapacity `json:"output_capacity"`
	OutputRowsMatch             bool                `json:"output_rows_match"`
	Valid                       bool                `json:"valid"`
}

type t2GuardLocalResult struct {
	B4GuardedVsCanonical        *PSGlobalMetric `json:"guarded_b4_vs_canonical"`
	NativeVsCanonical           *PSGlobalMetric `json:"native_b4_vs_canonical"`
	ImprovementFactor           float64         `json:"improvement_factor"`
	ResidualDirectionSimilarity float64         `json:"residual_direction_similarity"`
	GuardedCoefficientErrorReal string          `json:"guarded_coefficient_quantization_error_real"`
	GuardedCoefficientErrorImag string          `json:"guarded_coefficient_quantization_error_imag"`
}

type t2GuardCandidate struct {
	Name              string                         `json:"name"`
	Status            string                         `json:"status"`
	SkipReason        string                         `json:"skip_reason,omitempty"`
	GuardedOperations map[string]*t2GuardOperation   `json:"guarded_operations,omitempty"`
	Contractions      map[string]*t2GuardContraction `json:"contractions,omitempty"`
	Local             map[string]*t2GuardLocalResult `json:"local_precision,omitempty"`
	Downstream        *b4BabyDownstream              `json:"downstream,omitempty"`
	PublicMargin      float64                        `json:"public_margin,omitempty"`
	ValidContracts    bool                           `json:"valid_arithmetic_contracts"`
	PublicPass        bool                           `json:"public_like_pass"`
}

type t2GuardResult struct {
	SchemaVersion         string                         `json:"schema_version"`
	Timestamp             time.Time                      `json:"timestamp"`
	Primary               RepositoryMetadata             `json:"primary_repository"`
	Lattigo               RepositoryMetadata             `json:"lattigo_repository"`
	Environment           EnvironmentMetadata            `json:"environment"`
	Config                BootstrapConfig                `json:"config"`
	Parameters            ExperimentParameters           `json:"effective_parameters"`
	Provenance            map[string]interface{}         `json:"provenance"`
	G0Controls            map[string]interface{}         `json:"g0_controls"`
	NativeT2              map[string]t2GuardOperation    `json:"native_t2_operation"`
	G1                    *t2GuardCandidate              `json:"g1_t2_guard"`
	G1Downstream          *b4BabyDownstream              `json:"g1_downstream,omitempty"`
	X4CapacityOnly        map[string]postMod1S2CCapacity `json:"x4_capacity_only"`
	G2                    *t2GuardCandidate              `json:"g2_t6_t2_guard,omitempty"`
	Classification        string                         `json:"classification"`
	FirstRemainingBlocker string                         `json:"first_remaining_blocker"`
	RecommendedNextTarget string                         `json:"recommended_next_target"`
	Validation            map[string]interface{}         `json:"validation"`
}

type t2GuardBranchRun struct {
	Operation           *t2GuardOperation
	Contraction         *t2GuardContraction
	Ciphertext          *rlwe.Ciphertext
	PostGuardCiphertext *rlwe.Ciphertext
}

func t2GuardSignedCRT(q0, q1, r0, r1 uint64) *big.Int {
	q0Big := new(big.Int).SetUint64(q0)
	q1Big := new(big.Int).SetUint64(q1)
	value := new(big.Int).SetUint64(r0)
	delta := new(big.Int).Sub(new(big.Int).SetUint64(r1), new(big.Int).Mod(new(big.Int).Set(value), q1Big))
	delta.Mod(delta, q1Big)
	inverse := new(big.Int).ModInverse(new(big.Int).Mod(new(big.Int).Set(q0Big), q1Big), q1Big)
	delta.Mul(delta, inverse).Mod(delta, q1Big)
	value.Add(value, new(big.Int).Mul(q0Big, delta))
	q01 := new(big.Int).Mul(q0Big, q1Big)
	if value.Cmp(new(big.Int).Rsh(new(big.Int).Set(q01), 1)) > 0 {
		value.Sub(value, q01)
	}
	return value
}

// t2GuardIndependentRecoverQ01 is intentionally independent of the backend
// CRT reconstruction. It shares only the NTT inverse required to expose the
// authoritative q0/q1 coefficient rows, then performs its own big.Int CRT.
func t2GuardIndependentRecoverQ01(params ckks.Parameters, source *rlwe.Ciphertext) ([][]*big.Int, error) {
	if source == nil || source.Level() < 1 || !source.IsNTT {
		return nil, fmt.Errorf("independent q0/q1 recovery requires NTT level >= 1")
	}
	ringQ := params.RingQ().AtLevel(source.Level())
	q0 := params.RingQ().SubRings[0].Modulus
	q1 := params.RingQ().SubRings[1].Modulus
	result := make([][]*big.Int, source.Degree()+1)
	for component := range source.Value {
		coeff := ring.NewPoly(params.N(), 1)
		if err := fastckks.FastPartialINTT(ringQ, source.Value[component], coeff); err != nil {
			return nil, err
		}
		if source.IsMontgomery {
			params.RingQ().SubRings[0].IMForm(coeff.Coeffs[0], coeff.Coeffs[0])
			params.RingQ().SubRings[1].IMForm(coeff.Coeffs[1], coeff.Coeffs[1])
		}
		values := make([]*big.Int, params.N())
		for i := range values {
			values[i] = t2GuardSignedCRT(q0, q1, coeff.Coeffs[0][i], coeff.Coeffs[1][i])
		}
		result[component] = values
	}
	return result, nil
}

func t2GuardBigIntsEqual(a, b [][]*big.Int) bool {
	if len(a) != len(b) {
		return false
	}
	for component := range a {
		if len(a[component]) != len(b[component]) {
			return false
		}
		for i := range a[component] {
			if a[component][i].Cmp(b[component][i]) != 0 {
				return false
			}
		}
	}
	return true
}

func t2GuardFlatten(values [][]*big.Int) []*big.Int {
	result := make([]*big.Int, 0)
	for _, component := range values {
		result = append(result, component...)
	}
	return result
}

func t2GuardCapacity(params ckks.Parameters, source *rlwe.Ciphertext) (t2GuardCapacityEvidence, [][]*big.Int, error) {
	independent, err := t2GuardIndependentRecoverQ01(params, source)
	if err != nil {
		return t2GuardCapacityEvidence{}, nil, err
	}
	backend, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return t2GuardCapacityEvidence{}, nil, err
	}
	full, fullCapacity, err := postMod1S2CLiftFull(params, source)
	if err != nil {
		return t2GuardCapacityEvidence{}, nil, err
	}
	fastCapacity, err := postMod1S2CCapacityFromFastCiphertext(params, source)
	if err != nil {
		return t2GuardCapacityEvidence{}, nil, err
	}
	independentCapacity := postMod1S2CCapacityFromValues(params, t2GuardFlatten(independent))
	recoveryMatch := t2GuardBigIntsEqual(independent, backend)
	rowsMatch := postMod1S2CRowsEqual(params, source, full)
	return t2GuardCapacityEvidence{Fast: fastCapacity, FullRNS: fullCapacity, IndependentQ01: independentCapacity, FastVsFullRowsMatch: rowsMatch, IndependentRecoveryMatch: recoveryMatch, CenteredUnique: fastCapacity.Pass && fullCapacity.Pass && independentCapacity.Pass && rowsMatch && recoveryMatch}, independent, nil
}

func t2GuardX4Capacity(params ckks.Parameters, source *rlwe.Ciphertext) (postMod1S2CCapacity, error) {
	values, err := t2GuardIndependentRecoverQ01(params, source)
	if err != nil {
		return postMod1S2CCapacity{}, err
	}
	for component := range values {
		for i := range values[component] {
			values[component][i] = new(big.Int).Lsh(new(big.Int).Set(values[component][i]), 1)
		}
	}
	return postMod1S2CCapacityFromValues(params, t2GuardFlatten(values)), nil
}

func t2GuardStateEvidence(params ckks.Parameters, source *rlwe.Ciphertext, values []complex128) (t2GuardState, error) {
	capacity, _, err := t2GuardCapacity(params, source)
	if err != nil {
		return t2GuardState{}, err
	}
	return t2GuardState{State: b4BabyStateEvidence(params, source, values), Capacity: capacity}, nil
}

func t2GuardMetadataEqual(a, b *rlwe.Ciphertext) bool {
	return a != nil && b != nil && a.Level() == b.Level() && a.Degree() == b.Degree() && a.IsNTT == b.IsNTT && a.IsMontgomery == b.IsMontgomery && a.Scale.Equal(b.Scale) && a.N() == b.N()
}

func t2GuardRoundSigned(value *big.Int) *big.Int {
	return guardedPowerRoundSigned(value, 1)
}

func t2GuardContract(params ckks.Parameters, source, native *rlwe.Ciphertext) (*rlwe.Ciphertext, *t2GuardContraction, error) {
	values, err := t2GuardIndependentRecoverQ01(params, source)
	if err != nil {
		return nil, nil, err
	}
	backend, err := postMod1S2CRecoverQ01(params, source)
	if err != nil {
		return nil, nil, err
	}
	evidence := &t2GuardContraction{IndependentRecoveryMatch: t2GuardBigIntsEqual(values, backend)}
	if !evidence.IndependentRecoveryMatch {
		evidence.Valid = false
		return nil, evidence, nil
	}
	result := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*result.MetaData = *source.MetaData
	result.IsNTT, result.IsMontgomery = source.IsNTT, source.IsMontgomery
	result.Scale = source.Scale.Div(rlwe.NewScale(2))
	maxError := 0
	for component, row := range values {
		even, odd := 0, 0
		rounded := make([]*big.Int, len(row))
		for i, value := range row {
			if new(big.Int).And(new(big.Int).Set(value), big.NewInt(1)).Sign() == 0 {
				even++
			} else {
				odd++
			}
			rounded[i] = t2GuardRoundSigned(value)
			error := new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).Set(rounded[i]), 1), value)
			if abs := new(big.Int).Abs(error).Int64(); int(abs) > maxError {
				maxError = int(abs)
			}
		}
		evidence.EvenCount += even
		evidence.OddCount += odd
		normalNTT := postMod1S2CLiftSigned(params, rounded, source.Level())
		for limb := 0; limb < 2; limb++ {
			params.RingQ().SubRings[limb].MForm(normalNTT.Coeffs[limb], result.Value[component].Coeffs[limb])
		}
	}
	evidence.MaxAbsRoundingErrorInteger = maxError
	evidence.OddFraction = float64(evidence.OddCount) / float64(evidence.EvenCount+evidence.OddCount)
	oracle, err := guardedPowerRoundDivPow2Maintained(params, source, 1)
	if err != nil {
		return nil, nil, err
	}
	evidence.RoundedDivideOracleMatch = b4BabyExactState(result, oracle)
	evidence.OutputMetadataMatchesNative = t2GuardMetadataEqual(result, native)
	evidence.OutputRowsMatch = postMod1S2CRowsEqual(params, result, oracle)
	evidence.OutputCapacity, err = postMod1S2CCapacityFromFastCiphertext(params, result)
	if err != nil {
		return nil, nil, err
	}
	evidence.Valid = evidence.IndependentRecoveryMatch && evidence.RoundedDivideOracleMatch && evidence.OutputMetadataMatchesNative && evidence.OutputRowsMatch && evidence.OutputCapacity.Pass
	return result, evidence, nil
}

func t2GuardApply(params ckks.Parameters, eval *fastckks.Evaluator, power *rlwe.Ciphertext, coefficient *bignum.Complex, native *rlwe.Ciphertext, nativePost *rlwe.Ciphertext, id string) (t2GuardBranchRun, error) {
	guarded := native.CopyNew()
	if err := eval.MulIntegerMaintained(guarded, big.NewInt(2), guarded); err != nil {
		return t2GuardBranchRun{}, err
	}
	guarded.Scale = native.Scale.Mul(rlwe.NewScale(2))
	guardedBeforeAdd := guarded.CopyNew()
	preCapacity, _, err := t2GuardCapacity(params, guarded)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	encoding, err := psMixedScalarEncodingScale(params, power.Scale, guarded.Scale, coefficient, power.Level())
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	if err := eval.MulThenAdd(power, coefficient, guarded); err != nil {
		return t2GuardBranchRun{}, err
	}
	postCapacity, _, err := t2GuardCapacity(params, guarded)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	operation := b4BabyOperationRow(params, id, "guarded_scalar_mul_then_add", 2, coefficient, power, guardedBeforeAdd.Scale, guarded.Scale, encoding, false, "2", true)
	operation.ResultCapacityRatio = postCapacity.IndependentQ01.MaxAbsOverQ01Half
	operation.ScalarCenteredUnique = preCapacity.IndependentQ01.Pass && postCapacity.IndependentQ01.Pass
	preValues, err := psGlobalDecode(params, guardedBeforeAdd)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	postValues, err := psGlobalDecode(params, guarded)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	preState, err := t2GuardStateEvidence(params, guardedBeforeAdd, preValues)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	postState, err := t2GuardStateEvidence(params, guarded, postValues)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	contracted, contraction, err := t2GuardContract(params, guarded, nativePost)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	if contracted == nil {
		return t2GuardBranchRun{Operation: &t2GuardOperation{Operation: operation, PreState: preState, PostState: postState, ValidContracts: false}, Contraction: contraction, PostGuardCiphertext: guarded}, nil
	}
	contractedValues, err := psGlobalDecode(params, contracted)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	contractedState, err := t2GuardStateEvidence(params, contracted, contractedValues)
	if err != nil {
		return t2GuardBranchRun{}, err
	}
	valid := preCapacity.CenteredUnique && postCapacity.CenteredUnique && contraction.Valid
	return t2GuardBranchRun{Operation: &t2GuardOperation{Operation: operation, PreState: preState, PostState: postState, ContractedState: contractedState, ValidContracts: valid}, Contraction: contraction, Ciphertext: contracted, PostGuardCiphertext: guarded}, nil
}

func t2GuardNativeOperation(params ckks.Parameters, eval *fastckks.Evaluator, context psLocalizationAcceptedContext, pre, post *rlwe.Ciphertext, valuesPre, valuesPost []complex128) (t2GuardOperation, error) {
	ops, _, err := b4BabyLedgerRows(params, eval, context)
	if err != nil {
		return t2GuardOperation{}, err
	}
	var operation b4BabyOperation
	for _, row := range ops {
		if row.ID == "B4-term-2" {
			operation = row
		}
	}
	if operation.ID == "" {
		return t2GuardOperation{}, fmt.Errorf("native T2 operation not found")
	}
	preState, err := t2GuardStateEvidence(params, pre, valuesPre)
	if err != nil {
		return t2GuardOperation{}, err
	}
	postState, err := t2GuardStateEvidence(params, post, valuesPost)
	if err != nil {
		return t2GuardOperation{}, err
	}
	return t2GuardOperation{Operation: operation, PreState: preState, PostState: postState, ValidContracts: preState.Capacity.CenteredUnique && postState.Capacity.CenteredUnique}, nil
}

func t2GuardLocal(params ckks.Parameters, guarded, native, canonical *rlwe.Ciphertext, guardedCoefficient b4BabyOperation) (*t2GuardLocalResult, error) {
	guardedValues, err := psGlobalDecode(params, guarded)
	if err != nil {
		return nil, err
	}
	nativeValues, err := psGlobalDecode(params, native)
	if err != nil {
		return nil, err
	}
	canonicalValues, err := psLocalizationCanonicalValues(params, canonical)
	if err != nil {
		return nil, err
	}
	guardedMetric := psLocalizationMetric(canonicalValues, guardedValues, t2GuardSemanticThreshold)
	nativeMetric := psLocalizationMetric(canonicalValues, nativeValues, t2GuardSemanticThreshold)
	improvement := 0.0
	if guardedMetric.MaxComponent == 0 {
		improvement = math.Inf(1)
	} else {
		improvement = nativeMetric.MaxComponent / guardedMetric.MaxComponent
	}
	return &t2GuardLocalResult{B4GuardedVsCanonical: guardedMetric, NativeVsCanonical: nativeMetric, ImprovementFactor: improvement, ResidualDirectionSimilarity: psLocalizationDirection(subtractComplex(guardedValues, canonicalValues), subtractComplex(nativeValues, canonicalValues)), GuardedCoefficientErrorReal: guardedCoefficient.QuantizationErrorReal, GuardedCoefficientErrorImag: guardedCoefficient.QuantizationErrorImag}, nil
}

func subtractComplex(a, b []complex128) []complex128 {
	result := make([]complex128, len(a))
	for i := range a {
		result[i] = a[i] - b[i]
	}
	return result
}

func t2GuardG2Replay(params ckks.Parameters, eval *fastckks.Evaluator, context psLocalizationAcceptedContext, guardT6, guardT2 bool) (*rlwe.Ciphertext, map[string]*t2GuardOperation, map[string]*t2GuardContraction, error) {
	poly := context.Plan.Value[4]
	powers := context.Branch.Base.OraclePowerMap
	powerOne := powers[1]
	if powerOne == nil {
		return nil, nil, nil, fmt.Errorf("G2 missing power 1")
	}
	value := fastckks.NewCiphertext(params, 1, poly.Level)
	*value.MetaData = *powerOne.MetaData
	value.IsNTT, value.IsMontgomery, value.Scale = powerOne.IsNTT, powerOne.IsMontgomery, poly.Scale
	psGlobalZero(value)
	operations := map[string]*t2GuardOperation{}
	contractions := map[string]*t2GuardContraction{}
	if poly.IsEven {
		if err := eval.Add(value, poly.Coeffs[0], value); err != nil {
			return nil, nil, nil, err
		}
	}
	for key := poly.Degree(); key > 0; key-- {
		if (poly.IsEven || poly.IsOdd) && ((key&1 == 0 && !poly.IsEven) || (key&1 == 1 && !poly.IsOdd)) {
			continue
		}
		coefficient := poly.Coeffs[key]
		if coefficient == nil {
			continue
		}
		power := powers[key]
		if power == nil {
			return nil, nil, nil, fmt.Errorf("G2 missing power %d", key)
		}
		if power.Scale.Cmp(value.Scale) > 0 {
			ratio := power.Scale.Div(value.Scale).BigInt()
			if err := eval.MulIntegerMaintained(value, ratio, value); err != nil {
				return nil, nil, nil, err
			}
			value.Scale = power.Scale
		}
		if !coefficient.IsInt() && power.Scale.Cmp(value.Scale) >= 0 {
			encoding, err := psMixedScalarEncodingScale(params, power.Scale, value.Scale, coefficient, power.Level())
			if err != nil {
				return nil, nil, nil, err
			}
			if err := eval.MulIntegerMaintained(value, encoding.BigInt(), value); err != nil {
				return nil, nil, nil, err
			}
			value.Scale = value.Scale.Mul(encoding)
		}
		guard := (key == 6 && guardT6) || (key == 2 && guardT2)
		if guard {
			native := value.CopyNew()
			if err := eval.MulThenAdd(power, coefficient, native); err != nil {
				return nil, nil, nil, err
			}
			run, err := t2GuardApply(params, eval, power, coefficient, value, native, fmt.Sprintf("B4-term-%d", key))
			if err != nil {
				return nil, nil, nil, err
			}
			if run.Ciphertext == nil {
				return nil, operations, contractions, nil
			}
			operations[fmt.Sprintf("B4-term-%d", key)] = run.Operation
			contractions[fmt.Sprintf("B4-term-%d", key)] = run.Contraction
			value = run.Ciphertext
			continue
		}
		if err := eval.MulThenAdd(power, coefficient, value); err != nil {
			return nil, nil, nil, err
		}
	}
	return value, operations, contractions, nil
}

func t2GuardReplaySuffix(profile q056PreparedProfile, context psLocalizationAcceptedContext, b4 *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	replay, err := psRescaleGuardReplayPSWithBoundaryOverride(profile.BTP.BootstrappingParameters, profile.Fast.FastCKKS, context.Plan, context.Branch.Base.OraclePowerMap, context.Branch.Base.PowerExpected, context.Branch.Base.OraclePowerValues, context.Branch.Base.InputValues, precisionSweepScale(psLocalizationPlanScaleExponent), context.GuardMap, context.FinalOverride, context.BoundaryOverride, psGlobalReplayReset{ID: "B4-term-2", Ciphertext: b4})
	if err != nil {
		return nil, err
	}
	if replay.Final == nil {
		return nil, fmt.Errorf("B4-term-2 suffix replay returned nil final")
	}
	return replay.Final, nil
}

func t2GuardWrite(result t2GuardResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func runFIX001P3DesignLogN13B4T2LocalScalarGuard(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	params := profile.BTP.BootstrappingParameters
	primaryAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3T2GuardPrimary, "HEAD") == requiredFIX001P3T2GuardPrimary
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryBranch := gitOutput(backendRoot, "branch", "--show-current")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := t2GuardResult{SchemaVersion: "fix-001-p3-design-logn13-b4-t2-local-scalar-guard.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(profile.Residual, profile.BTP), Provenance: map[string]interface{}{"primary_required_base": requiredFIX001P3T2GuardPrimary, "primary_base_ancestor": primaryAncestor, "secondary_required_commit": requiredFIX001P3T2GuardSecondary, "secondary_commit": secondaryCommit, "secondary_branch": secondaryBranch, "secondary_clean": secondaryClean}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "q0_q1_q2_profile": "56/39/40", "common_plan_scale_2^92": true, "no_coefficient_change": true, "no_plan_scale_change": true, "no_q_parameter_changes": true, "no_generated_power_redesign": true, "no_g0_da_s2c_change": true, "no_secondary_production_changes": true, "no_production_integration": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if cfg.LogN != 13 || !primaryAncestor || secondaryCommit != requiredFIX001P3T2GuardSecondary || secondaryBranch != "fast-ckks" || !secondaryClean {
		result.Classification = "logn13_b4_t2_scalar_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary provenance or LogN13 precondition mismatch"
		return t2GuardWrite(result, outPath)
	}
	realContext, imagContext, realWork, imagWork, candidate, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return err
	}
	actualDownstream, err := g0MergeDownstream(profile, realWork.ActualCiphertext, imagWork.ActualCiphertext)
	if err != nil {
		return err
	}
	result.G0Controls = map[string]interface{}{"actual_public_like": actualDownstream.PublicLike, "actual_post_s2c": actualDownstream.PostS2C, "metadata_correct": actualDownstream.MetadataCorrect, "contracts_pass": actualDownstream.ContractsPass, "accepted_candidate": candidate.Name}
	realG0, err := g0MergeOperand(profile, realContext)
	if err != nil {
		return err
	}
	imagG0, err := g0MergeOperand(profile, imagContext)
	if err != nil {
		return err
	}
	ca, err := g0MergeRunCase(profile, realG0, imagG0, "CA", realG0.CanonicalA, realG0.Snapshot.Product, imagG0.CanonicalA, imagG0.Snapshot.Product)
	if err != nil {
		return err
	}
	result.G0Controls["canonical_parent_public_like"] = ca.Downstream.PublicLike
	result.G0Controls["g0_parent_alpha"] = func() float64 { _, alpha, _ := g0MergeAlphaSweep(profile, realG0, imagG0, "parent"); return alpha }()
	result.Validation["g0_precondition_pass"] = actualDownstream.PublicLike != nil && math.Abs(actualDownstream.PublicLike.MaxComponent-0.010683237260415292) < 1e-12 && ca.Downstream.PublicLike != nil && math.Abs(ca.Downstream.PublicLike.MaxComponent-0.0032351181652570046) < 1e-12
	if !result.Validation["g0_precondition_pass"].(bool) {
		result.Classification = "logn13_b4_t2_scalar_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "G0 actual or canonical-parent public-like control materially changed"
		return t2GuardWrite(result, outPath)
	}
	realCheck, realOK := b4BabyFindCheckpoint(realContext.Run.Replay.Checks, "B4-term-2")
	imagCheck, imagOK := b4BabyFindCheckpoint(imagContext.Run.Replay.Checks, "B4-term-2")
	realPre, realPreOK := b4BabyFindCheckpoint(realContext.Run.Replay.Checks, "B4-term-4")
	imagPre, imagPreOK := b4BabyFindCheckpoint(imagContext.Run.Replay.Checks, "B4-term-4")
	if !realOK || !imagOK || !realPreOK || !imagPreOK {
		result.Classification = "logn13_b4_t2_scalar_guard_precondition_mismatch"
		result.FirstRemainingBlocker = "accepted B4 replay did not expose native T2 boundaries"
		return t2GuardWrite(result, outPath)
	}
	realCanonical, realCanonicalValues, _, _, err := psLocalizationCanonical(params, realCheck)
	if err != nil {
		return err
	}
	imagCanonical, imagCanonicalValues, _, _, err := psLocalizationCanonical(params, imagCheck)
	if err != nil {
		return err
	}
	realPreValues, err := psGlobalDecode(params, realPre.ciphertext)
	if err != nil {
		return err
	}
	imagPreValues, err := psGlobalDecode(params, imagPre.ciphertext)
	if err != nil {
		return err
	}
	realNativeValues, err := psGlobalDecode(params, realCheck.ciphertext)
	if err != nil {
		return err
	}
	imagNativeValues, err := psGlobalDecode(params, imagCheck.ciphertext)
	if err != nil {
		return err
	}
	realNativeState := b4BabyStateEvidence(params, realCheck.ciphertext, realNativeValues)
	imagNativeState := b4BabyStateEvidence(params, imagCheck.ciphertext, imagNativeValues)
	result.G0Controls["b4_actual_vs_canonical_real"] = psLocalizationMetric(realCanonicalValues, realNativeValues, t2GuardSemanticThreshold)
	result.G0Controls["b4_actual_vs_canonical_imag"] = psLocalizationMetric(imagCanonicalValues, imagNativeValues, t2GuardSemanticThreshold)
	result.G0Controls["b4_final_q01_capacity_real"] = realNativeState.Capacity
	result.G0Controls["b4_final_q01_capacity_imag"] = imagNativeState.Capacity
	result.G0Controls["b4_parent_identity_real"] = b4BabyExactState(realCheck.ciphertext, realContext.Run.Replay.G0Merge.A)
	result.G0Controls["b4_parent_identity_imag"] = b4BabyExactState(imagCheck.ciphertext, imagContext.Run.Replay.G0Merge.A)
	result.Validation["g0_b4_parent_identity_exact"] = b4BabyExactState(realCheck.ciphertext, realContext.Run.Replay.G0Merge.A) && b4BabyExactState(imagCheck.ciphertext, imagContext.Run.Replay.G0Merge.A)
	result.NativeT2 = map[string]t2GuardOperation{}
	result.NativeT2["real"], err = t2GuardNativeOperation(params, profile.Fast.FastCKKS, realContext, realPre.ciphertext, realCheck.ciphertext, realPreValues, realNativeValues)
	if err != nil {
		return err
	}
	result.NativeT2["imag"], err = t2GuardNativeOperation(params, profile.Fast.FastCKKS, imagContext, imagPre.ciphertext, imagCheck.ciphertext, imagPreValues, imagNativeValues)
	if err != nil {
		return err
	}
	result.G1 = &t2GuardCandidate{Name: "G1-T2-one-bit-local-guard", Status: "executed", GuardedOperations: map[string]*t2GuardOperation{}, Contractions: map[string]*t2GuardContraction{}, Local: map[string]*t2GuardLocalResult{}}
	guardRuns := map[string]t2GuardBranchRun{}
	for _, item := range []struct {
		name    string
		context psLocalizationAcceptedContext
		pre     *rlwe.Ciphertext
		native  *rlwe.Ciphertext
	}{{"real", realContext, realPre.ciphertext, realCheck.ciphertext}, {"imag", imagContext, imagPre.ciphertext, imagCheck.ciphertext}} {
		poly := item.context.Plan.Value[4]
		power := item.context.Branch.Base.OraclePowerMap[2]
		run, e := t2GuardApply(params, profile.Fast.FastCKKS, power, poly.Coeffs[2], item.pre, item.native, "B4-term-2")
		if e != nil {
			return e
		}
		guardRuns[item.name] = run
		result.G1.GuardedOperations[item.name] = run.Operation
		result.G1.Contractions[item.name] = run.Contraction
		if run.Ciphertext != nil {
			canonical := realCanonical
			if item.name == "imag" {
				canonical = imagCanonical
			}
			local, e := t2GuardLocal(params, run.Ciphertext, item.native, canonical, run.Operation.Operation)
			if e != nil {
				return e
			}
			result.G1.Local[item.name] = local
		}
	}
	result.X4CapacityOnly = map[string]postMod1S2CCapacity{}
	for name, run := range guardRuns {
		if run.Operation != nil && run.PostGuardCiphertext != nil {
			result.X4CapacityOnly[name], err = t2GuardX4Capacity(params, run.PostGuardCiphertext)
			if err != nil {
				return err
			}
		}
	}
	result.Validation["g1_capacity_contracts"] = true
	result.Validation["g4_contraction_contracts"] = true
	result.Validation["g1_contraction_oracle_match"] = true
	for name, run := range guardRuns {
		if run.Operation == nil || !run.Operation.PreState.Capacity.CenteredUnique || !run.Operation.PostState.Capacity.CenteredUnique {
			result.Validation["g1_capacity_contracts"] = false
		}
		if run.Contraction == nil || !run.Contraction.Valid {
			result.Validation["g4_contraction_contracts"] = false
			result.Validation["g1_contraction_oracle_match"] = false
		}
		_ = name
	}
	result.G1.ValidContracts = result.Validation["g1_capacity_contracts"].(bool)
	result.G1.ValidContracts = result.G1.ValidContracts && result.Validation["g4_contraction_contracts"].(bool)
	result.G2 = &t2GuardCandidate{Name: "G2-T6-plus-T2-one-bit-local-guards", Status: "skipped"}
	if guardRuns["real"].Ciphertext != nil && guardRuns["imag"].Ciphertext != nil {
		realFinal, e := t2GuardReplaySuffix(profile, realContext, guardRuns["real"].Ciphertext)
		if e != nil {
			return e
		}
		imagFinal, e := t2GuardReplaySuffix(profile, imagContext, guardRuns["imag"].Ciphertext)
		if e != nil {
			return e
		}
		downstream, e := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psLocalizationPlanScaleExponent), realFinal, imagFinal)
		if e != nil {
			return e
		}
		result.G1Downstream = b4BabyCombineDownstream(downstream, downstream)
		result.G1.Downstream = result.G1Downstream
		if downstream.PublicLike != nil {
			result.G1.PublicMargin = t2GuardPublicThreshold - downstream.PublicLike.MaxComponent
			result.G1.PublicPass = downstream.PublicLike.Pass
		}
	}
	if !result.Validation["g1_contraction_oracle_match"].(bool) {
		result.G2.SkipReason = "G1 contraction oracle mismatch; conditional G2 is not authorized"
		result.Classification = "logn13_b4_t2_scalar_guard_contraction_mismatch"
		result.FirstRemainingBlocker = "independent q0/q1 reconstruction or centered rounded divide-by-2 oracle mismatch"
		result.RecommendedNextTarget = "repair the diagnostic contraction before considering production integration"
	} else if !result.Validation["g1_capacity_contracts"].(bool) {
		result.G2.SkipReason = "G1 capacity/row contract failed; conditional G2 is not authorized"
		result.Classification = "logn13_b4_t2_one_bit_scalar_guard_capacity_blocker"
		result.FirstRemainingBlocker = "one-bit guarded T2 state failed centered-capacity or row contract"
		result.RecommendedNextTarget = "repair the one-bit capacity-preserving guard design"
	} else if result.G1.ValidContracts && result.G1.PublicPass && result.G1.Downstream != nil && result.G1.Downstream.ContractsPass {
		result.G1.Status = "pass"
		result.Classification = "logn13_b4_t2_one_bit_scalar_guard_system_sufficient"
		result.FirstRemainingBlocker = "none"
		result.RecommendedNextTarget = "production integration of this exact local arithmetic mechanism"
		result.G2.SkipReason = "G1 passed the public threshold with all arithmetic contracts; minimal candidate was sufficient"
		result.Validation["g7_not_needed_minimal_candidate_passed"] = true
	} else if result.G1Downstream != nil && !result.G1Downstream.ContractsPass {
		result.G2.SkipReason = "G1 downstream contracts did not pass; conditional G2 is not authorized"
		result.Classification = "logn13_b4_local_scalar_guard_insufficient_precision"
		result.FirstRemainingBlocker = "G1 arithmetic contracts passed but downstream contracts did not"
		result.RecommendedNextTarget = "retain the guard as diagnostic evidence; do not production-integrate"
	} else {
		result.G2 = &t2GuardCandidate{Name: "G2-T6-plus-T2-one-bit-local-guards", Status: "executed", GuardedOperations: map[string]*t2GuardOperation{}, Contractions: map[string]*t2GuardContraction{}, Local: map[string]*t2GuardLocalResult{}}
		g2Real, opReal, conReal, e := t2GuardG2Replay(params, profile.Fast.FastCKKS, realContext, true, true)
		if e != nil {
			return e
		}
		g2Imag, opImag, conImag, e := t2GuardG2Replay(params, profile.Fast.FastCKKS, imagContext, true, true)
		if e != nil {
			return e
		}
		for id, op := range opReal {
			result.G2.GuardedOperations["real/"+id] = op
		}
		for id, op := range opImag {
			result.G2.GuardedOperations["imag/"+id] = op
		}
		for id, con := range conReal {
			result.G2.Contractions["real/"+id] = con
		}
		for id, con := range conImag {
			result.G2.Contractions["imag/"+id] = con
		}
		if g2Real != nil && g2Imag != nil {
			realFinal, e := t2GuardReplaySuffix(profile, realContext, g2Real)
			if e != nil {
				return e
			}
			imagFinal, e := t2GuardReplaySuffix(profile, imagContext, g2Imag)
			if e != nil {
				return e
			}
			downstream, e := psLocalizationRunAcceptedDownstream(profile, precisionSweepScale(psLocalizationPlanScaleExponent), realFinal, imagFinal)
			if e != nil {
				return e
			}
			result.G2.Downstream = b4BabyCombineDownstream(downstream, downstream)
			if downstream.PublicLike != nil {
				result.G2.PublicMargin = t2GuardPublicThreshold - downstream.PublicLike.MaxComponent
				result.G2.PublicPass = downstream.PublicLike.Pass
			}
		}
		result.G2.ValidContracts = g2Real != nil && g2Imag != nil
		for _, con := range result.G2.Contractions {
			result.G2.ValidContracts = result.G2.ValidContracts && con != nil && con.Valid
		}
		if result.G2.ValidContracts && result.G2.PublicPass && result.G2.Downstream != nil && result.G2.Downstream.ContractsPass {
			result.G2.Status = "pass"
			result.Classification = "logn13_b4_t6_t2_one_bit_scalar_guards_system_sufficient"
			result.FirstRemainingBlocker = "none"
			result.RecommendedNextTarget = "production integration of the exact T6+T2 local arithmetic mechanisms"
		} else {
			result.G2.Status = "insufficient"
			result.Classification = "logn13_b4_local_scalar_guard_insufficient_precision"
			result.FirstRemainingBlocker = "G1 and conditional G2 evidence-backed guards remain above the public threshold"
			result.RecommendedNextTarget = "do not production-integrate; design a new evidence-backed local precision mechanism"
		}
	}
	result.Validation["g1_public_like_pass"] = result.G1.PublicPass
	result.Validation["g6_downstream_replayed"] = result.G1Downstream != nil
	result.Validation["x4_capacity_only_not_executed"] = true
	result.Validation["g7_conditional_candidate_handled"] = result.G2 != nil || result.Validation["g7_not_needed_minimal_candidate_passed"] == true
	return t2GuardWrite(result, outPath)
}
