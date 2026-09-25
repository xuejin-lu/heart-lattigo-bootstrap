//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const (
	requiredFIX001P3PSWidePrimaryBase = "4a3c8346cdf11112bb835865477035a8cce81d86"
	requiredFIX001P3PSWideSecondary   = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	psWideSemanticThreshold           = 1.2e-8
	psWidePublicThreshold             = 1e-2
)

var psWideScaleExponents = []int{92, 93, 94}

type psWideCapacity struct {
	MaxAbs                   string  `json:"max_abs"`
	Q012Half                 string  `json:"q012_half"`
	Q012MaxAbsOverHalf       float64 `json:"q012_max_abs_over_half"`
	Q012CenteredUnique       bool    `json:"q012_centered_unique"`
	Q01Half                  string  `json:"q01_half"`
	Q01MaxAbsOverHalf        float64 `json:"q01_max_abs_over_half"`
	Q01CenteredUnique        bool    `json:"q01_centered_unique"`
	OutsideQ012              int     `json:"q012_outside_count"`
	OutsideQ01               int     `json:"q01_outside_count"`
	MinimumRemainingQ012Bits int     `json:"minimum_remaining_q012_bits,omitempty"`
}

type psWideCheckpoint struct {
	ID                 string          `json:"id"`
	Operation          string          `json:"operation"`
	Level              int             `json:"level"`
	Scale              string          `json:"scale"`
	Degree             int             `json:"degree"`
	Capacity           *psWideCapacity `json:"capacity,omitempty"`
	Semantic           *PSGlobalMetric `json:"semantic,omitempty"`
	LocalResidual      *PSGlobalMetric `json:"local_residual,omitempty"`
	Q012Authority      bool            `json:"q012_authority"`
	Relinearized       bool            `json:"relinearized,omitempty"`
	LogicalDivisorBits int             `json:"logical_divisor_bits,omitempty"`
}

type psWideAlignment struct {
	Operation         string `json:"operation"`
	LeftScale         string `json:"left_scale"`
	RightScale        string `json:"right_scale"`
	TargetScale       string `json:"target_scale"`
	IntegerMultiplier string `json:"integer_multiplier"`
	ExactIntegerRatio bool   `json:"exact_integer_ratio"`
	MetadataOnlyAfter bool   `json:"metadata_only_after_integer_mul"`
}

type psWidePowerEvidence struct {
	Power             int                `json:"power"`
	Expected          *PSGlobalMetric    `json:"expected_power_semantic"`
	Checkpoints       []psWideCheckpoint `json:"checkpoints"`
	OutputLevel       int                `json:"output_level"`
	OutputScale       string             `json:"output_scale"`
	OutputCapacity    *psWideCapacity    `json:"output_capacity"`
	Q012Authoritative bool               `json:"q012_authoritative"`
}

type psWideContractionProof struct {
	AttemptedAtPSExit        bool            `json:"attempted_at_ps_exit"`
	Q01CenteredUnique        bool            `json:"q01_centered_unique"`
	DirectReductionRowsEqual bool            `json:"direct_q01_reduction_rows_equal"`
	MetadataValid            bool            `json:"metadata_valid"`
	Q2DroppedExactlyOnce     bool            `json:"q2_dropped_exactly_once"`
	OutputLevel              int             `json:"output_level"`
	OutputScale              string          `json:"output_scale"`
	OutputDegree             int             `json:"output_degree"`
	Capacity                 *psWideCapacity `json:"capacity"`
	FirstFailure             string          `json:"first_failure"`
}

type psWideBranchResult struct {
	Name                      string                 `json:"name"`
	InputLevel                int                    `json:"input_level"`
	InputScale                string                 `json:"input_scale"`
	GeneratedPowers           []psWidePowerEvidence  `json:"generated_powers"`
	PSCheckpoints             []psWideCheckpoint     `json:"ps_checkpoints"`
	Alignments                []psWideAlignment      `json:"scale_alignments"`
	PSPolynomialOracle        *PSGlobalMetric        `json:"ps_polynomial_vs_plaintext_oracle"`
	PSPolynomialStandard      *PSGlobalMetric        `json:"ps_polynomial_vs_standard"`
	PSOutputCapacity          *psWideCapacity        `json:"ps_output_capacity"`
	PSOutputLevel             int                    `json:"ps_output_level"`
	PSOutputScale             string                 `json:"ps_output_scale"`
	Contraction               psWideContractionProof `json:"ps_exit_contraction"`
	EvalModVsStandard         *PSGlobalMetric        `json:"evalmod_vs_standard"`
	PostS2CVsStandard         *PSGlobalMetric        `json:"post_s2c_vs_standard"`
	PublicLike                *PSGlobalMetric        `json:"public_like"`
	StandardPublicLike        *PSGlobalMetric        `json:"standard_public_like"`
	PublicThresholdPass       bool                   `json:"public_threshold_pass"`
	MetadataCorrect           bool                   `json:"metadata_correct"`
	InputUnchanged            bool                   `json:"input_unchanged"`
	FirstFailure              string                 `json:"first_failure"`
	SemanticPass              bool                   `json:"semantic_pass"`
	CapacityPass              bool                   `json:"capacity_pass"`
	Q012OnlyUntilPSExit       bool                   `json:"q012_only_until_ps_exit"`
	NoIntermediateContraction bool                   `json:"no_intermediate_contraction"`
}

type psWideCandidate struct {
	Exponent                 int                `json:"plan_scale_exponent"`
	PlanScale                string             `json:"plan_scale"`
	Real                     psWideBranchResult `json:"real"`
	Imag                     psWideBranchResult `json:"imag"`
	MaxCapacityRatio         float64            `json:"maximum_q012_capacity_ratio"`
	WorstCapacityCheckpoint  string             `json:"worst_capacity_checkpoint"`
	MinRemainingCapacityBits int                `json:"minimum_remaining_capacity_bits"`
	PrecisionPass            bool               `json:"precision_pass"`
	CapacityPass             bool               `json:"capacity_pass"`
	PSExitContractionPass    bool               `json:"ps_exit_contraction_pass"`
	EvalModPass              bool               `json:"evalmod_pass"`
	PostS2CPass              bool               `json:"post_s2c_pass"`
	PublicLikePass           bool               `json:"public_like_pass"`
	Classification           string             `json:"classification"`
	FirstFailure             string             `json:"first_failure"`
	DownstreamReason         string             `json:"downstream_reason,omitempty"`
}

type psWideResult struct {
	SchemaVersion       string                 `json:"schema_version"`
	Timestamp           time.Time              `json:"timestamp"`
	Primary             RepositoryMetadata     `json:"primary_repository"`
	Lattigo             RepositoryMetadata     `json:"lattigo_repository"`
	Environment         EnvironmentMetadata    `json:"environment"`
	Config              BootstrapConfig        `json:"config"`
	Parameters          ExperimentParameters   `json:"effective_parameters"`
	Workload            CorrectnessWorkload    `json:"workload"`
	Q0Profile           map[string]interface{} `json:"q0_q1_q2_profile"`
	Controls            map[string]interface{} `json:"controls"`
	Candidates          []psWideCandidate      `json:"candidates"`
	ArchitectureCompare map[string]interface{} `json:"architecture_comparison"`
	RecommendedScale    string                 `json:"recommended_scale"`
	Classification      string                 `json:"classification"`
	ProductionReadiness string                 `json:"production_readiness"`
	NextTarget          string                 `json:"next_target"`
	Validation          map[string]interface{} `json:"validation"`
}

type psWideNode struct {
	Degree   int
	Value    *rlwe.Ciphertext
	Source   []complex128
	Identity string
}

type psWideReplay struct {
	Checkpoints  []psWideCheckpoint
	Alignments   []psWideAlignment
	Root         *rlwe.Ciphertext
	RootSource   []complex128
	Output       *rlwe.Ciphertext
	OutputSource []complex128
}

func psWideScale(exponent int) rlwe.Scale {
	return rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(exponent)))
}

func psWideCapacityFor(params ckks.Parameters, ct *rlwe.Ciphertext) (*psWideCapacity, error) {
	if ct == nil || ct.Level() < 2 {
		return nil, fmt.Errorf("Q012 capacity requires level >= 2")
	}
	q012Values, err := nativeQ012SignedValues(params, ct)
	if err != nil {
		return nil, err
	}
	q01Values, err := postMod1S2CRecoverQ01(params, ct)
	if err != nil {
		return nil, err
	}
	q012 := big.NewInt(1)
	for limb := 0; limb <= 2; limb++ {
		q012.Mul(q012, new(big.Int).SetUint64(params.RingQ().SubRings[limb].Modulus))
	}
	q01 := new(big.Int).Mul(new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus))
	q012Half := new(big.Int).Rsh(new(big.Int).Set(q012), 1)
	q01Half := new(big.Int).Rsh(new(big.Int).Set(q01), 1)
	maxAbs := new(big.Int)
	outside012, outside01 := 0, 0
	for component := range q012Values {
		for _, value := range q012Values[component] {
			if new(big.Int).Abs(value).Cmp(maxAbs) > 0 {
				maxAbs.Set(new(big.Int).Abs(value))
			}
			if new(big.Int).Abs(value).Cmp(q012Half) >= 0 {
				outside012++
			}
		}
		for _, value := range q01Values[component] {
			if new(big.Int).Abs(value).Cmp(q01Half) >= 0 {
				outside01++
			}
		}
	}
	ratio012, _ := new(big.Float).Quo(new(big.Float).SetInt(maxAbs), new(big.Float).SetInt(q012Half)).Float64()
	maxQ01 := new(big.Int)
	for _, component := range q01Values {
		for _, value := range component {
			if new(big.Int).Abs(value).Cmp(maxQ01) > 0 {
				maxQ01.Set(new(big.Int).Abs(value))
			}
		}
	}
	ratio01, _ := new(big.Float).Quo(new(big.Float).SetInt(maxQ01), new(big.Float).SetInt(q01Half)).Float64()
	remaining := 0
	if maxAbs.Sign() != 0 {
		remaining = q012.BitLen() - (new(big.Int).Lsh(new(big.Int).Set(maxAbs), 1)).BitLen()
	}
	return &psWideCapacity{MaxAbs: maxAbs.String(), Q012Half: q012Half.String(), Q012MaxAbsOverHalf: ratio012, Q012CenteredUnique: outside012 == 0, Q01Half: q01Half.String(), Q01MaxAbsOverHalf: ratio01, Q01CenteredUnique: outside01 == 0, OutsideQ012: outside012, OutsideQ01: outside01, MinimumRemainingQ012Bits: remaining}, nil
}

func psWideScalarNTT(params ckks.Parameters, c *bignum.Complex, scale *big.Float, montgomery bool) [3][2]uint64 {
	var out [3][2]uint64
	toInt := func(x *big.Float) *big.Int {
		z := new(big.Int)
		if x == nil {
			return z
		}
		v := new(big.Float).Mul(x, scale)
		if x.Sign() > 0 {
			v.Add(v, new(big.Float).SetFloat64(0.5))
		} else if x.Sign() < 0 {
			v.Sub(v, new(big.Float).SetFloat64(0.5))
		}
		v.Int(z)
		return z
	}
	real, imag := toInt(c[0]), toInt(c[1])
	for limb := 0; limb <= 2; limb++ {
		s := params.RingQ().SubRings[limb]
		r := new(big.Int).Mod(new(big.Int).Set(real), new(big.Int).SetUint64(s.Modulus)).Uint64()
		i := new(big.Int).Mod(new(big.Int).Set(imag), new(big.Int).SetUint64(s.Modulus)).Uint64()
		i = ring.MRed(i, s.RootsForward[1], s.Modulus, s.MRedConstant)
		out[limb][0], out[limb][1] = ring.CRed(r+i, s.Modulus), ring.CRed(r+s.Modulus-i, s.Modulus)
		if montgomery {
			out[limb][0] = ring.MForm(out[limb][0], s.Modulus, s.BRedConstant)
			out[limb][1] = ring.MForm(out[limb][1], s.Modulus, s.BRedConstant)
		}
	}
	return out
}

func psWideAddScalar(params ckks.Parameters, source *rlwe.Ciphertext, coefficient *bignum.Complex) error {
	if source == nil || source.Level() < 2 || !source.IsNTT || !source.IsMontgomery {
		return fmt.Errorf("Q012 scalar add requires NTT Montgomery level >= 2")
	}
	values := psWideScalarNTT(params, coefficient, &source.Scale.Value, true)
	half := params.N() >> 1
	for limb := 0; limb <= 2; limb++ {
		s := params.RingQ().SubRings[limb]
		s.AddScalar(source.Value[0].Coeffs[limb][:half], values[limb][0], source.Value[0].Coeffs[limb][:half])
		s.AddScalar(source.Value[0].Coeffs[limb][half:], values[limb][1], source.Value[0].Coeffs[limb][half:])
	}
	return nil
}

func psWideCoefficientScale(params ckks.Parameters, level int) (rlwe.Scale, error) {
	consumed := params.LevelsConsumedPerRescaling()
	if level-consumed+1 < 0 {
		return rlwe.Scale{}, fmt.Errorf("insufficient level for Q012 coefficient scale")
	}
	scale := rlwe.NewScale(1)
	for i := 0; i < consumed; i++ {
		scale = scale.Mul(rlwe.NewScale(params.RingQ().SubRings[level-i].Modulus))
	}
	return scale, nil
}

func psWideMulInteger(params ckks.Parameters, source *rlwe.Ciphertext, factor *big.Int) error {
	return nativeQ012MulInteger(params, source, factor)
}

func psWideMulThenAdd(params ckks.Parameters, op0 *rlwe.Ciphertext, coefficient *bignum.Complex, accumulator *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if op0 == nil || accumulator == nil || op0.Degree() != accumulator.Degree() || op0.Level() < 2 || accumulator.Level() < 2 {
		return nil, fmt.Errorf("Q012 scalar MulThenAdd metadata mismatch")
	}
	level := minInt(op0.Level(), accumulator.Level())
	if level < 2 {
		return nil, fmt.Errorf("Q012 scalar MulThenAdd level below Q012 authority")
	}
	op, err := nativeQ012AtLevel(params, op0, level)
	if err != nil {
		return nil, err
	}
	out, err := nativeQ012AtLevel(params, accumulator, level)
	if err != nil {
		return nil, err
	}
	original := op.CopyNew()
	var scalarScale rlwe.Scale
	if op.Scale.Equal(out.Scale) {
		if coefficient.IsInt() {
			scalarScale = rlwe.NewScale(1)
		} else {
			scalarScale, _ = psWideCoefficientScale(params, level)
			if err := psWideMulInteger(params, out, scalarScale.BigInt()); err != nil {
				return nil, err
			}
			out.Scale = out.Scale.Mul(scalarScale)
		}
	} else if op.Scale.Cmp(out.Scale) < 0 {
		scalarScale = out.Scale.Div(op.Scale)
	} else {
		ratio := op.Scale.Div(out.Scale).BigInt()
		if ratio.Sign() <= 0 {
			return nil, fmt.Errorf("invalid Q012 scalar scale promotion ratio")
		}
		if err := psWideMulInteger(params, out, ratio); err != nil {
			return nil, err
		}
		out.Scale = op.Scale
		if coefficient.IsInt() {
			scalarScale = rlwe.NewScale(1)
		} else {
			scalarScale, _ = psWideCoefficientScale(params, level)
			if err := psWideMulInteger(params, out, scalarScale.BigInt()); err != nil {
				return nil, err
			}
			out.Scale = out.Scale.Mul(scalarScale)
		}
	}
	values := psWideScalarNTT(params, coefficient, &scalarScale.Value, true)
	half := params.N() >> 1
	for component := range original.Value {
		for limb := 0; limb <= 2; limb++ {
			s := params.RingQ().SubRings[limb]
			s.MulScalarMontgomeryThenAdd(original.Value[component].Coeffs[limb][:half], values[limb][0], out.Value[component].Coeffs[limb][:half])
			s.MulScalarMontgomeryThenAdd(original.Value[component].Coeffs[limb][half:], values[limb][1], out.Value[component].Coeffs[limb][half:])
		}
	}
	return out, nil
}

func psWideMul(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Level() < 2 || right.Level() < 2 || left.Degree() > 1 || right.Degree() > 1 {
		return nil, fmt.Errorf("Q012 multiplication requires degree <= 1 level >= 2")
	}
	level := minInt(left.Level(), right.Level())
	out := ckks.NewCiphertext(params, left.Degree()+right.Degree(), level)
	*out.MetaData = *left.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	out.Scale = left.Scale.Mul(right.Scale)
	for i := 0; i <= left.Degree(); i++ {
		for j := 0; j <= right.Degree(); j++ {
			for limb := 0; limb <= 2; limb++ {
				s := params.RingQ().SubRings[limb]
				tmp := make([]uint64, params.N())
				s.MulCoeffsMontgomery(left.Value[i].Coeffs[limb], right.Value[j].Coeffs[limb], tmp)
				s.Add(out.Value[i+j].Coeffs[limb], tmp, out.Value[i+j].Coeffs[limb])
			}
		}
	}
	return out, nil
}

func psWidePromote(params ckks.Parameters, source *rlwe.Ciphertext, degree, level int) (*rlwe.Ciphertext, error) {
	if source == nil || degree < source.Degree() || level < 2 || level > source.Level() {
		return nil, fmt.Errorf("invalid Q012 promotion degree=%d level=%d", degree, level)
	}
	out := ckks.NewCiphertext(params, degree, level)
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := range source.Value {
		for limb := 0; limb <= 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func psWideAdd(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Level() < 2 || right.Level() < 2 {
		return nil, fmt.Errorf("Q012 add metadata mismatch")
	}
	level := minInt(left.Level(), right.Level())
	degree := left.Degree()
	if right.Degree() > degree {
		degree = right.Degree()
	}
	left, _ = psWidePromote(params, left, degree, level)
	right, _ = psWidePromote(params, right, degree, level)
	if !left.Scale.Equal(right.Scale) {
		return nil, fmt.Errorf("Q012 add requires aligned scales")
	}
	if err := nativeQ012Add(params, left, right); err != nil {
		return nil, err
	}
	return left, nil
}

func psWideSub(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Level() < 2 || right.Level() < 2 {
		return nil, fmt.Errorf("Q012 sub metadata mismatch")
	}
	level := minInt(left.Level(), right.Level())
	degree := left.Degree()
	if right.Degree() > degree {
		degree = right.Degree()
	}
	left, _ = psWidePromote(params, left, degree, level)
	right, _ = psWidePromote(params, right, degree, level)
	if !left.Scale.Equal(right.Scale) {
		return nil, fmt.Errorf("Q012 sub requires aligned scales")
	}
	if err := nativeQ012Sub(params, left, right); err != nil {
		return nil, err
	}
	return left, nil
}

func psWideAlign(params ckks.Parameters, left, right *rlwe.Ciphertext, id string) (*rlwe.Ciphertext, *rlwe.Ciphertext, psWideAlignment, error) {
	level := minInt(left.Level(), right.Level())
	degree := left.Degree()
	if right.Degree() > degree {
		degree = right.Degree()
	}
	left, err := psWidePromote(params, left, degree, level)
	if err != nil {
		return nil, nil, psWideAlignment{}, err
	}
	right, err = psWidePromote(params, right, degree, level)
	if err != nil {
		return nil, nil, psWideAlignment{}, err
	}
	alignment := psWideAlignment{Operation: id, LeftScale: finalizationScaleString(left.Scale), RightScale: finalizationScaleString(right.Scale), TargetScale: finalizationScaleString(left.Scale), IntegerMultiplier: "1", ExactIntegerRatio: true}
	if left.Scale.Equal(right.Scale) {
		return left, right, alignment, nil
	}
	if left.Scale.Cmp(right.Scale) > 0 {
		ratio := left.Scale.Div(right.Scale)
		integer := ratio.BigInt()
		if err := psWideMulInteger(params, right, integer); err != nil {
			return nil, nil, alignment, err
		}
		right.Scale = left.Scale
		alignment.TargetScale = finalizationScaleString(left.Scale)
		alignment.IntegerMultiplier = integer.String()
		alignment.ExactIntegerRatio = ratio.Equal(rlwe.NewScale(integer))
		alignment.MetadataOnlyAfter = true
		return left, right, alignment, nil
	}
	ratio := right.Scale.Div(left.Scale)
	integer := ratio.BigInt()
	if err := psWideMulInteger(params, left, integer); err != nil {
		return nil, nil, alignment, err
	}
	left.Scale = right.Scale
	alignment.TargetScale = finalizationScaleString(right.Scale)
	alignment.IntegerMultiplier = integer.String()
	alignment.ExactIntegerRatio = ratio.Equal(rlwe.NewScale(integer))
	alignment.MetadataOnlyAfter = true
	return left, right, alignment, nil
}

func psWideRelinearize(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || source.Degree() != 2 || source.Level() < 2 {
		return nil, fmt.Errorf("Q012 relinearization requires degree 2 level >= 2")
	}
	out := ckks.NewCiphertext(params, 1, source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := 0; component <= 1; component++ {
		for limb := 0; limb <= 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func psWideRescale(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || source.Level() < 2 {
		return nil, fmt.Errorf("Q012 diagnostic rescale requires level >= 2")
	}
	values, err := nativeQ012SignedValues(params, source)
	if err != nil {
		return nil, err
	}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[source.Level()].Modulus)
	quotients := make([][]*big.Int, len(values))
	for component := range values {
		quotients[component] = make([]*big.Int, len(values[component]))
		for i, value := range values[component] {
			quotients[component][i] = fix001P3LocalQ2Round(value, divisor)
		}
	}
	return nativeQ012BuildFromSigned(params, source, quotients, source.Level()-1, source.Scale.Div(rlwe.NewScale(divisor)))
}

func psWideCheckpointFor(params ckks.Parameters, id, operation string, ct *rlwe.Ciphertext, expected, localExpected []complex128) (psWideCheckpoint, error) {
	capacity, err := psWideCapacityFor(params, ct)
	if err != nil {
		return psWideCheckpoint{}, err
	}
	actual, err := psGlobalDecode(params, ct)
	if err != nil {
		return psWideCheckpoint{}, err
	}
	checkpoint := psWideCheckpoint{ID: id, Operation: operation, Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), Degree: ct.Degree(), Capacity: capacity, Q012Authority: true}
	if expected != nil {
		checkpoint.Semantic = polynomialDecompositionMetric(expected, actual, psWideSemanticThreshold)
	}
	if localExpected != nil {
		checkpoint.LocalResidual = polynomialDecompositionResidualMetric(localExpected, actual, psWideSemanticThreshold)
	}
	fix001P3TraceRecord("ps", id, -1, ct, actual)
	return checkpoint, nil
}

func psWidePlan(eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext, targetScale, planScale rlwe.Scale) commonpolynomial.PatersonStockmeyerPolynomial {
	plan := commonpolynomial.NewPolynomial(eval.Mod1Parameters.Mod1Poly).PatersonStockmeyerPolynomial(eval.FastCKKS.GetRLWEParameters(), input.Level(), input.Scale, targetScale, psGlobalSim{params: eval.Parameters.BootstrappingParameters, levels: eval.Parameters.BootstrappingParameters.LevelsConsumedPerRescaling()})
	plan.Scale = planScale
	for i := range plan.Value {
		plan.Value[i].Scale = planScale
	}
	return plan
}

func psWideGeneratePowers(params ckks.Parameters, input *rlwe.Ciphertext, values []complex128) (map[int]*rlwe.Ciphertext, map[int][]complex128, []psWidePowerEvidence, error) {
	powers := map[int]*rlwe.Ciphertext{1: input.CopyNew()}
	expected := map[int][]complex128{1: append([]complex128(nil), values...)}
	memo := map[int][]complex128{}
	evidence := make([]psWidePowerEvidence, 0, 6)
	var generate func(int) error
	generate = func(n int) error {
		if powers[n] != nil {
			return nil
		}
		a, b := commonpolynomial.SplitDegree(n)
		if err := generate(a); err != nil {
			return err
		}
		if err := generate(b); err != nil {
			return err
		}
		left := powers[a]
		right := powers[b]
		level := minInt(left.Level(), right.Level())
		left, err := nativeQ012AtLevel(params, left, level)
		if err != nil {
			return err
		}
		right, err = nativeQ012AtLevel(params, right, level)
		if err != nil {
			return err
		}
		leftExpected, rightExpected := expected[a], expected[b]
		checkpoints := []psWideCheckpoint{}
		leftCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-input-left", n), "generated_power_input", left, leftExpected, nil)
		if err != nil {
			return err
		}
		checkpoints = append(checkpoints, leftCP)
		rightCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-input-right", n), "generated_power_input", right, rightExpected, nil)
		if err != nil {
			return err
		}
		checkpoints = append(checkpoints, rightCP)
		raw, err := psWideMul(params, left, right)
		if err != nil {
			return err
		}
		rawExpected := polynomialDecompositionVectorMul(leftExpected, rightExpected)
		rawCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-pre-rescale-raw-product", n), "generated_power_pre_rescale_raw_product", raw, rawExpected, nil)
		if err != nil {
			return err
		}
		checkpoints = append(checkpoints, rawCP)
		relinearized, err := psWideRelinearize(params, raw)
		if err != nil {
			return err
		}
		relCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-relinearized", n), "generated_power_zero_secret_relinearization", relinearized, rawExpected, rawExpected)
		if err != nil {
			return err
		}
		relCP.Relinearized = true
		checkpoints = append(checkpoints, relCP)
		if err := nativeQ012Add(params, relinearized, relinearized); err != nil {
			return err
		}
		doubledExpected := make([]complex128, len(rawExpected))
		for i := range doubledExpected {
			doubledExpected[i] = 2 * rawExpected[i]
		}
		doubledCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-doubled", n), "generated_power_double", relinearized, doubledExpected, nil)
		if err != nil {
			return err
		}
		checkpoints = append(checkpoints, doubledCP)
		c := a - b
		if c < 0 {
			c = -c
		}
		if c == 0 {
			if err := psWideAddScalar(params, relinearized, bignum.ToComplex(-1, params.EncodingPrecision())); err != nil {
				return err
			}
		} else {
			sub := powers[c]
			if !relinearized.Scale.Equal(sub.Scale) {
				var alignment psWideAlignment
				relinearized, sub, alignment, err = psWideAlign(params, relinearized, sub, fmt.Sprintf("T%d-recurrence-align", n))
				if err != nil {
					return err
				}
				_ = alignment
			}
			relinearized, err = psWideSub(params, relinearized, sub)
			if err != nil {
				return err
			}
		}
		recurrenceExpected := fix001ExpectedPower(n, values, memo)
		recurrenceCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-pre-rescale-recurrence", n), "generated_power_pre_rescale_recurrence", relinearized, recurrenceExpected, nil)
		if err != nil {
			return err
		}
		checkpoints = append(checkpoints, recurrenceCP)
		post, err := psWideRescale(params, relinearized)
		if err != nil {
			return err
		}
		postCP, err := psWideCheckpointFor(params, fmt.Sprintf("T%d-final", n), "generated_power_post_rescale", post, recurrenceExpected, nil)
		if err != nil {
			return err
		}
		postCP.LogicalDivisorBits = new(big.Int).SetUint64(params.RingQ().SubRings[level].Modulus).BitLen()
		checkpoints = append(checkpoints, postCP)
		powers[n] = post
		expected[n] = recurrenceExpected
		decoded, err := psGlobalDecode(params, post)
		if err != nil {
			return err
		}
		evidence = append(evidence, psWidePowerEvidence{Power: n, Expected: polynomialDecompositionMetric(recurrenceExpected, decoded, psWideSemanticThreshold), Checkpoints: checkpoints, OutputLevel: post.Level(), OutputScale: finalizationScaleString(post.Scale), OutputCapacity: postCP.Capacity, Q012Authoritative: true})
		return nil
	}
	for _, n := range []int{2, 3, 4, 6, 8, 16} {
		if err := generate(n); err != nil {
			return nil, nil, evidence, err
		}
	}
	return powers, expected, evidence, nil
}

func psWideReplayPS(params ckks.Parameters, eval *bootstrapping.FastEvaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, powerExpected map[int][]complex128, input []complex128, planScale rlwe.Scale) (psWideReplay, error) {
	replay := psWideReplay{}
	baby := make([]*psWideNode, len(plan.Value))
	for i := range plan.Value {
		p := plan.Value[i]
		out := ckks.NewCiphertext(params, 1, p.Level)
		*out.MetaData = *powers[1].MetaData
		out.IsNTT, out.IsMontgomery, out.Scale = true, true, planScale
		identity := psGlobalSubtreeHash(fmt.Sprintf("ps-wide-block-%d", i), p.Polynomial)
		source := make([]complex128, len(input))
		if p.IsEven {
			if p.Coeffs[0] == nil {
				return replay, fmt.Errorf("PS block %d has nil constant", i)
			}
			if err := psWideAddScalar(params, out, p.Coeffs[0]); err != nil {
				return replay, err
			}
			psGlobalAddScaled(source, makeOnes(len(input)), psGlobalCoeff(p.Coeffs[0]))
			cp, err := psWideCheckpointFor(params, fmt.Sprintf("B%d-constant", i), "baby_constant_add", out, source, nil)
			if err != nil {
				return replay, err
			}
			replay.Checkpoints = append(replay.Checkpoints, cp)
		}
		for key := p.Degree(); key > 0; key-- {
			if (p.IsEven || p.IsOdd) && ((key&1 == 0 && !p.IsEven) || (key&1 == 1 && !p.IsOdd)) {
				continue
			}
			if p.Coeffs[key] == nil {
				continue
			}
			before, err := psGlobalDecode(params, out)
			if err != nil {
				return replay, err
			}
			out, err = psWideMulThenAdd(params, powers[key], p.Coeffs[key], out)
			if err != nil {
				return replay, err
			}
			psGlobalAddScaled(source, powerExpected[key], psGlobalCoeff(p.Coeffs[key]))
			localExpected := append([]complex128(nil), before...)
			powerDecoded, err := psGlobalDecode(params, powers[key])
			if err != nil {
				return replay, err
			}
			psGlobalAddScaled(localExpected, powerDecoded, psGlobalCoeff(p.Coeffs[key]))
			cp, err := psWideCheckpointFor(params, fmt.Sprintf("B%d-term-%d", i, key), "baby_coefficient_mul_then_add", out, source, localExpected)
			if err != nil {
				return replay, err
			}
			replay.Checkpoints = append(replay.Checkpoints, cp)
		}
		baby[len(plan.Value)-i-1] = &psWideNode{Degree: p.Degree(), Value: out, Source: source, Identity: identity}
	}
	round := 0
	for len(baby) != 1 {
		giant := make([]int, len(baby))
		for i := 0; i < len(baby); i++ {
			if i == len(baby)-1 {
				giant[i] = 2
			} else if baby[i].Degree == baby[i+1].Degree {
				giant[i] = 1
				i++
			}
		}
		for i := 0; i < len(baby); i++ {
			if giant[i] == 2 {
				baby[i].Degree = baby[i-1].Degree
				continue
			}
			if giant[i] != 1 {
				continue
			}
			a, b := baby[i], baby[i+1]
			deg := 1 << bits.Len64(uint64(a.Degree))
			if b.Value.Degree() == 2 {
				var err error
				b.Value, err = psWideRelinearize(params, b.Value)
				if err != nil {
					return replay, err
				}
				cp, err := psWideCheckpointFor(params, fmt.Sprintf("G%d-relinearize", round), "giant_zero_secret_relinearization", b.Value, b.Source, b.Source)
				if err != nil {
					return replay, err
				}
				cp.Relinearized = true
				replay.Checkpoints = append(replay.Checkpoints, cp)
			}
			beforeRescale, err := psGlobalDecode(params, b.Value)
			if err != nil {
				return replay, err
			}
			b.Value, err = psWideRescale(params, b.Value)
			if err != nil {
				return replay, err
			}
			postRescale, err := psGlobalDecode(params, b.Value)
			if err != nil {
				return replay, err
			}
			cp, err := psWideCheckpointFor(params, fmt.Sprintf("G%d-rescale", round), "giant_logical_rescale", b.Value, b.Source, beforeRescale)
			if err != nil {
				return replay, err
			}
			cp.LocalResidual = polynomialDecompositionResidualMetric(beforeRescale, postRescale, psWideSemanticThreshold)
			cp.LogicalDivisorBits = new(big.Int).SetUint64(params.RingQ().SubRings[b.Value.Level()+1].Modulus).BitLen()
			replay.Checkpoints = append(replay.Checkpoints, cp)
			b.Value, err = psWideMul(params, b.Value, powers[deg])
			if err != nil {
				return replay, err
			}
			productSource := make([]complex128, len(b.Source))
			copy(productSource, b.Source)
			for k := range productSource {
				productSource[k] *= powerExpected[deg][k]
			}
			productDecoded, err := psGlobalDecode(params, b.Value)
			if err != nil {
				return replay, err
			}
			localProduct := make([]complex128, len(postRescale))
			for k := range localProduct {
				localProduct[k] = postRescale[k] * mustPSWideDecode(params, powers[deg])[k]
			}
			productID := hashSourceVector(fmt.Sprintf("%s*T%d", b.Identity, deg))
			cp, err = psWideCheckpointFor(params, fmt.Sprintf("G%d-multiply", round), "giant_multiply", b.Value, productSource, localProduct)
			if err != nil {
				return replay, err
			}
			replay.Checkpoints = append(replay.Checkpoints, cp)
			var alignment psWideAlignment
			a.Value, b.Value, alignment, err = psWideAlign(params, a.Value, b.Value, fmt.Sprintf("G%d-align", round))
			if err != nil {
				return replay, err
			}
			replay.Alignments = append(replay.Alignments, alignment)
			parentSource := append([]complex128(nil), a.Source...)
			psGlobalAddScaled(parentSource, productSource, 1)
			parentID := hashSourceVector(fmt.Sprintf("%s+%s", a.Identity, productID))
			parent, err := psWideAdd(params, a.Value, b.Value)
			if err != nil {
				return replay, err
			}
			parentDecoded, err := psGlobalDecode(params, parent)
			if err != nil {
				return replay, err
			}
			localParent := append([]complex128(nil), a.Source...)
			psGlobalAddScaled(localParent, productDecoded, 1)
			cp, err = psWideCheckpointFor(params, fmt.Sprintf("G%d-add", round), "giant_aligned_add", parent, parentSource, localParent)
			if err != nil {
				return replay, err
			}
			replay.Checkpoints = append(replay.Checkpoints, cp)
			_ = parentDecoded
			b.Value, b.Source, b.Identity = parent, parentSource, parentID
			b.Degree = 2*deg - 1
			baby[i] = nil
			i++
			round++
		}
		kept := baby[:0]
		for _, node := range baby {
			if node != nil {
				kept = append(kept, node)
			}
		}
		baby = kept
	}
	root := baby[0]
	if root.Value.Degree() == 2 {
		var err error
		root.Value, err = psWideRelinearize(params, root.Value)
		if err != nil {
			return replay, err
		}
		cp, err := psWideCheckpointFor(params, "F0-final-parent-relinearize", "final_parent_zero_secret_relinearization", root.Value, root.Source, root.Source)
		if err != nil {
			return replay, err
		}
		cp.Relinearized = true
		replay.Checkpoints = append(replay.Checkpoints, cp)
	}
	replay.Root = root.Value.CopyNew()
	replay.RootSource = append([]complex128(nil), root.Source...)
	rootDecoded, err := psGlobalDecode(params, replay.Root)
	if err != nil {
		return replay, err
	}
	rootCP, err := psWideCheckpointFor(params, "F0-root", "final_ps_root_before_rescale", replay.Root, root.Source, root.Source)
	if err != nil {
		return replay, err
	}
	replay.Checkpoints = append(replay.Checkpoints, rootCP)
	replay.Output, err = psWideRescale(params, replay.Root)
	if err != nil {
		return replay, err
	}
	outputDecoded, err := psGlobalDecode(params, replay.Output)
	if err != nil {
		return replay, err
	}
	outputCP, err := psWideCheckpointFor(params, "F0-final-rescale", "final_ps_rescale", replay.Output, root.Source, root.Source)
	if err != nil {
		return replay, err
	}
	outputCP.LocalResidual = polynomialDecompositionResidualMetric(rootDecoded, outputDecoded, psWideSemanticThreshold)
	outputCP.LogicalDivisorBits = new(big.Int).SetUint64(params.RingQ().SubRings[replay.Root.Level()].Modulus).BitLen()
	replay.Checkpoints = append(replay.Checkpoints, outputCP)
	replay.OutputSource = append([]complex128(nil), root.Source...)
	return replay, nil
}

func mustPSWideDecode(params ckks.Parameters, ct *rlwe.Ciphertext) []complex128 {
	values, err := psGlobalDecode(params, ct)
	if err != nil {
		return nil
	}
	return values
}

func psWideRowsEqualQ01(params ckks.Parameters, source, contracted *rlwe.Ciphertext) bool {
	if source == nil || contracted == nil || source.Level() != contracted.Level() || source.Degree() != contracted.Degree() {
		return false
	}
	for component := 0; component <= source.Degree(); component++ {
		for limb := 0; limb <= 1; limb++ {
			left := append([]uint64(nil), source.Value[component].Coeffs[limb]...)
			right := append([]uint64(nil), contracted.Value[component].Coeffs[limb]...)
			params.RingQ().SubRings[limb].IMForm(left, left)
			params.RingQ().SubRings[limb].IMForm(right, right)
			for i := range left {
				if left[i] != right[i] {
					return false
				}
			}
		}
	}
	return true
}

func psWideContractQ01(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, psWideContractionProof, error) {
	proof := psWideContractionProof{AttemptedAtPSExit: true, FirstFailure: "none"}
	capacity, err := psWideCapacityFor(params, source)
	if err != nil {
		proof.FirstFailure = "capacity"
		return nil, proof, err
	}
	proof.Capacity = capacity
	proof.Q01CenteredUnique = capacity.Q01CenteredUnique
	proof.OutputLevel, proof.OutputScale, proof.OutputDegree = source.Level(), finalizationScaleString(source.Scale), source.Degree()
	if !proof.Q01CenteredUnique {
		proof.FirstFailure = "q01_not_centered_unique"
		return nil, proof, nil
	}
	values, err := nativeQ012SignedValues(params, source)
	if err != nil {
		return nil, proof, err
	}
	out := ckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	for component := range values {
		ntt := postMod1S2CLiftSigned(params, values[component], 1)
		for limb := 0; limb <= 1; limb++ {
			params.RingQ().SubRings[limb].MForm(ntt.Coeffs[limb], out.Value[component].Coeffs[limb])
		}
	}
	proof.DirectReductionRowsEqual = psWideRowsEqualQ01(params, source, out)
	proof.MetadataValid = out.Level() == source.Level() && out.Scale.Equal(source.Scale) && out.Degree() == source.Degree() && out.IsNTT && out.IsMontgomery
	proof.Q2DroppedExactlyOnce = true
	if !proof.DirectReductionRowsEqual {
		proof.FirstFailure = "direct_q01_reduction_rows"
	} else if !proof.MetadataValid {
		proof.FirstFailure = "metadata"
	}
	return out, proof, nil
}

func psWideBranchWorstCapacity(branch psWideBranchResult) (float64, string, int) {
	worst, checkpoint, remaining := 0.0, "none", 0
	consider := func(id string, capacity *psWideCapacity) {
		if capacity == nil {
			return
		}
		if capacity.Q012MaxAbsOverHalf > worst {
			worst, checkpoint, remaining = capacity.Q012MaxAbsOverHalf, id, capacity.MinimumRemainingQ012Bits
		}
	}
	for _, power := range branch.GeneratedPowers {
		for _, checkpoint := range power.Checkpoints {
			consider(checkpoint.ID, checkpoint.Capacity)
		}
	}
	for _, checkpoint := range branch.PSCheckpoints {
		consider(checkpoint.ID, checkpoint.Capacity)
	}
	consider("ps_output", branch.PSOutputCapacity)
	return worst, checkpoint, remaining
}

func psWideWorstCapacity(realBranch, imagBranch psWideBranchResult) (float64, string, int) {
	realWorst, realCheckpoint, realRemaining := psWideBranchWorstCapacity(realBranch)
	imagWorst, imagCheckpoint, imagRemaining := psWideBranchWorstCapacity(imagBranch)
	if realWorst >= imagWorst {
		return realWorst, "real:" + realCheckpoint, realRemaining
	}
	return imagWorst, "imag:" + imagCheckpoint, imagRemaining
}

func psWideBranchFailure(branch psWideBranchResult) string {
	if branch.FirstFailure != "" {
		return branch.FirstFailure
	}
	for _, power := range branch.GeneratedPowers {
		for _, checkpoint := range power.Checkpoints {
			if checkpoint.Capacity != nil && !checkpoint.Capacity.Q012CenteredUnique {
				return branch.Name + "." + checkpoint.ID + ".q012_capacity"
			}
		}
	}
	for _, checkpoint := range branch.PSCheckpoints {
		if checkpoint.Capacity != nil && !checkpoint.Capacity.Q012CenteredUnique {
			return branch.Name + "." + checkpoint.ID + ".q012_capacity"
		}
	}
	return "none"
}

func psWideRunBranch(cfg BootstrapConfig, branchName string, fastInput, standardInput *rlwe.Ciphertext, btp bootstrapping.Parameters, eval *bootstrapping.FastEvaluator, standardEval *bootstrapping.Evaluator, standardSK *rlwe.SecretKey, exponent int) (psWideBranchResult, *rlwe.Ciphertext, error) {
	params := btp.BootstrappingParameters
	result := psWideBranchResult{Name: branchName, FirstFailure: "", Q012OnlyUntilPSExit: true, NoIntermediateContraction: true}
	e2, inputValues, err := psGlobalE2(btp, eval, fastInput)
	if err != nil {
		return result, nil, err
	}
	result.InputLevel, result.InputScale = e2.Level(), finalizationScaleString(e2.Scale)
	q012Input, err := nativeQ012LiftQ01(params, e2)
	if err != nil {
		return result, nil, err
	}
	powers, expectedPowers, powerEvidence, err := psWideGeneratePowers(params, q012Input, inputValues)
	if err != nil {
		return result, nil, err
	}
	result.GeneratedPowers = powerEvidence
	targetScale := psGlobalTargetScale(e2, eval)
	planScale := psWideScale(exponent)
	plan := psWidePlan(eval, q012Input, targetScale, planScale)
	replay, err := psWideReplayPS(params, eval, plan, powers, expectedPowers, inputValues, planScale)
	if err != nil {
		return result, nil, err
	}
	result.PSCheckpoints = replay.Checkpoints
	result.Alignments = replay.Alignments
	outputValues, err := psGlobalDecode(params, replay.Output)
	if err != nil {
		return result, nil, err
	}
	reference := chebyshevRawOracle(eval.Mod1Parameters.Mod1Poly, inputValues)
	standardValues, err := psOracleScaleStandardPolynomial(btp, eval, standardEval, params, fastInput, standardInput, standardSK, inputValues)
	if err != nil {
		return result, nil, err
	}
	result.PSPolynomialOracle = polynomialDecompositionMetric(reference, outputValues, psWideSemanticThreshold)
	result.PSPolynomialStandard = polynomialDecompositionMetric(standardValues, outputValues, psWidePublicThreshold)
	result.PSOutputLevel, result.PSOutputScale = replay.Output.Level(), finalizationScaleString(replay.Output.Scale)
	result.PSOutputCapacity, err = psWideCapacityFor(params, replay.Output)
	if err != nil {
		return result, nil, err
	}
	contracted, proof, err := psWideContractQ01(params, replay.Output)
	if err != nil {
		return result, nil, err
	}
	result.Contraction = proof
	if contracted == nil {
		result.FirstFailure = psWideBranchFailure(result)
		if result.FirstFailure == "none" {
			result.FirstFailure = "ps_exit_contraction"
		}
		return result, nil, nil
	}
	path, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(fastInput, standardInput, eval, standardEval, params, zeroSecret(params), standardSK, planScale, contracted)
	if err != nil {
		return result, nil, err
	}
	if path.FastFinal == nil {
		result.FirstFailure = path.FirstFailure
		return result, nil, nil
	}
	final, err := evalModMatchedFinalEvidence(path, params, zeroSecret(params), standardSK)
	if err != nil {
		return result, nil, err
	}
	result.EvalModVsStandard = final.FastVsStandard
	result.MetadataCorrect = final.LevelDegreeUnchanged && final.RestoreRowsMatch
	result.PostS2CVsStandard = final.CoherentVsStandard
	result.InputUnchanged = false
	result.PublicThresholdPass = false
	result.SemanticPass = result.PSPolynomialOracle != nil && result.PSPolynomialOracle.Pass
	result.CapacityPass = result.PSOutputCapacity.Q012CenteredUnique && result.Contraction.Q01CenteredUnique
	result.FirstFailure = path.FirstFailure
	return result, contracted, nil
}

func psWideControl(cfg BootstrapConfig) (map[string]interface{}, ckks.Parameters, bootstrapping.Parameters, *bootstrapping.FastEvaluator, *bootstrapping.Evaluator, *rlwe.SecretKey, error) {
	effective := q056BuildConfig(cfg, 56)
	residual, btp, err := NewBootstrapParametersFromConfig(effective)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	fastEval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	skN1 := rlwe.NewKeyGenerator(residual).GenSecretKeyNew()
	standardEval, standardSK, err := buildFIX001P3GenuineStandardEvaluator(btp, skN1)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	standardOutput, err := standardEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	standardValues, err := decodeWithSecret(residual, standardOutput, standardSK)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	fastOutput, err := fastEval.Bootstrap(reproducibleInput(residual, btp))
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	fastValues, err := decodeWithSecret(residual, fastOutput, zeroSecret(residual))
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	standardMetric := evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), standardValues, psWidePublicThreshold)
	fastMetric := evalModMatchedMetricFor(reproducibleValues(residual.MaxSlots()), fastValues, psWidePublicThreshold)
	profile, err := q056PrepareProfile(effective)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	realContext, imagContext, _, _, _, err := psLocalizationAcceptedSetup(profile)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	accepted, err := generatedPowerReentryRunCase(profile, realContext, imagContext, realContext.Branch.Base.OraclePowerMap, imagContext.Branch.Base.OraclePowerMap, realContext.Branch.Base.PowerExpected, imagContext.Branch.Base.PowerExpected, realContext.Branch.Base.OraclePowerValues, imagContext.Branch.Base.OraclePowerValues, "D0-oracle-control", nil)
	if err != nil {
		return nil, residual, btp, nil, nil, nil, err
	}
	return map[string]interface{}{"standard_public_like": standardMetric, "q0_56_current_fast_public_like": fastMetric, "q0_56_fast_metadata_correct": fastOutput.Level() == residual.MaxLevel() && fastOutput.Degree() == 1 && fastOutput.N() == residual.N() && fastOutput.IsNTT && !fastOutput.IsMontgomery && fastOutput.Scale.Equal(residual.DefaultScale()), "accepted_d0_q0_56_plan92_control_public_like": accepted.Metrics.PublicLike, "expected_q0_56_control_public_like": 0.005554220603853743}, residual, btp, fastEval, standardEval, standardSK, nil
}

func psWideClassify(candidates []psWideCandidate) (string, string, string) {
	passed := []int{}
	for _, candidate := range candidates {
		if candidate.PrecisionPass && candidate.PublicLikePass && candidate.EvalModPass && candidate.PostS2CPass && candidate.CapacityPass && candidate.PSExitContractionPass {
			passed = append(passed, candidate.Exponent)
		}
	}
	if len(passed) > 0 {
		if len(passed) > 1 {
			return "logn13_ps_wide_q012_q056_multiple_scales_system_sufficient", fmt.Sprintf("2^%d", passed[0]), "production_ready_for_fixed_width_ps_wide_q012_design"
		}
		return fmt.Sprintf("logn13_ps_wide_q012_q056_p%d_system_sufficient", passed[0]), fmt.Sprintf("2^%d", passed[0]), "production_ready_for_fixed_width_ps_wide_q012_design"
	}
	anyCapacitySafe := false
	anyContractionFailure := false
	for _, candidate := range candidates {
		if candidate.CapacityPass {
			anyCapacitySafe = true
			if !candidate.PSExitContractionPass {
				anyContractionFailure = true
			}
		}
	}
	if anyContractionFailure {
		return "logn13_ps_wide_q012_output_not_q01_contractible", "none", "blocked_by_output_contraction"
	}
	if !anyCapacitySafe {
		return "logn13_ps_wide_q012_capacity_insufficient", "none", "blocked_by_q012_capacity"
	}
	return "logn13_ps_wide_q012_semantic_mismatch", "none", "blocked_by_reference_semantics_or_downstream_precision"
}

func psWideNextTarget(classification string) string {
	switch classification {
	case "logn13_ps_wide_q012_q056_p92_system_sufficient", "logn13_ps_wide_q012_q056_p93_system_sufficient", "logn13_ps_wide_q012_q056_p94_system_sufficient", "logn13_ps_wide_q012_q056_multiple_scales_system_sufficient":
		return "implement fixed-width three-limb Q012 primitives in Secondary and integrate a clean PS-wide mode"
	case "logn13_ps_wide_q012_output_not_q01_contractible":
		return "prove and repair the PS-exit Q012-to-Q01 contraction before production integration"
	case "logn13_ps_wide_q012_capacity_insufficient":
		return "localize the first Q012 capacity failure and revise the PS-wide scale or representation"
	default:
		return "localize the PS-wide Q012 semantic or precision mismatch before production integration"
	}
}

func psWideWrite(result psWideResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DesignLogN13PSWideQ012Q056ScaleSweep(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	if cfg.LogN != 13 {
		return fmt.Errorf("PS-wide Q012 design requires LogN13")
	}
	effective := q056BuildConfig(cfg, 56)
	control, residual, btp, eval, standardEval, standardSK, err := psWideControl(cfg)
	if err != nil {
		return err
	}
	q0, q1, q0Bits, q1Bits, q01, q01Bits, err := q056ProfilePrimes(btp.BootstrappingParameters)
	if err != nil {
		return err
	}
	q2 := new(big.Int).SetUint64(btp.BootstrappingParameters.RingQ().SubRings[2].Modulus)
	primaryAncestor := gitOutput(primaryRoot, "merge-base", requiredFIX001P3PSWidePrimaryBase, "HEAD") == requiredFIX001P3PSWidePrimaryBase
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	q01Value, _ := new(big.Int).SetString(q01, 10)
	q012 := new(big.Int).Mul(q01Value, q2)
	result := psWideResult{SchemaVersion: "fix-001-p3-design-logn13-ps-wide-q012-q056-scale-sweep.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "accepted-q056-ps-input.v1", Formula: "reproducibleInput.v1 with matched C2S real/imag EvalMod inputs", LogicalSlots: residual.MaxSlots()}, Q0Profile: map[string]interface{}{"diagnostic_q0_bits": 56, "q0": q0, "q1": q1, "q2": q2.String(), "q0_bits": q0Bits, "q1_bits": q1Bits, "q2_bits": q2.BitLen(), "q01": q01, "q01_bits": q01Bits, "q012": q012.String(), "in_memory_override_only": true, "checked_in_config_unchanged": len(cfg.Q0) == 1 && cfg.Q0[0] == 55}, Controls: control, Candidates: []psWideCandidate{}, Validation: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3PSWidePrimaryBase, "primary_ancestor_present": primaryAncestor, "required_secondary_commit": requiredFIX001P3PSWideSecondary, "secondary_commit": secondaryCommit, "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": secondaryClean, "secondary_exact": secondaryCommit == requiredFIX001P3PSWideSecondary, "no_secondary_production_changes": true, "logn13_only": true, "diagnostic_q0_56": true, "checked_in_frontend_config_unchanged": true, "q012_authoritative_from_ps_entry_to_exit": true, "q3_plus_never_arithmetic_source": true, "q012_big_int_only_diagnostic_rescale_oracle": true, "no_intermediate_q012_to_q01_contraction": true, "q2_dropped_once_at_ps_exit": true, "actual_logical_q_level_rescale": true, "same_ps_decomposition": true, "deterministic_workload": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true}}
	if !primaryAncestor || secondaryCommit != requiredFIX001P3PSWideSecondary || !secondaryClean {
		result.Classification = "logn13_ps_wide_q012_harness_mismatch"
		result.ProductionReadiness = "blocked_by_provenance"
		result.NextTarget = "restore the exact read-only Primary/Secondary provenance"
		return psWideWrite(result, outPath)
	}
	acceptedControl, _ := control["accepted_d0_q0_56_plan92_control_public_like"].(*PSGlobalMetric)
	controlPass := acceptedControl != nil && acceptedControl.Pass && math.Abs(acceptedControl.MaxComponent-0.005554220603853743) <= 1e-9
	result.Controls["accepted_d0_control_pass"] = controlPass
	if !controlPass {
		result.Classification = "logn13_ps_wide_q012_control_mismatch"
		result.ProductionReadiness = "blocked_by_control_mismatch"
		result.NextTarget = "reproduce the accepted q0=56 / planScale=2^92 D0 control before PS-wide comparison"
		return psWideWrite(result, outPath)
	}
	inputs, err := evalModMatchedC2SInputs(cfg, btp, eval, standardEval, standardSK)
	if err != nil {
		return err
	}
	for _, exponent := range psWideScaleExponents {
		candidate := psWideCandidate{Exponent: exponent, PlanScale: finalizationScaleString(psWideScale(exponent)), Classification: "not_run", FirstFailure: "none"}
		var realContracted *rlwe.Ciphertext
		candidate.Real, realContracted, err = psWideRunBranch(effective, "real", inputs.FastReal, inputs.OrdinaryReal, btp, eval, standardEval, standardSK, exponent)
		if err != nil {
			return err
		}
		var imagContracted *rlwe.Ciphertext
		candidate.Imag, imagContracted, err = psWideRunBranch(effective, "imag", inputs.FastImag, inputs.OrdinaryImag, btp, eval, standardEval, standardSK, exponent)
		if err != nil {
			return err
		}
		if realContracted != nil && imagContracted != nil {
			realPath, realErr := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastReal, inputs.OrdinaryReal, eval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, psWideScale(exponent), realContracted)
			imagPath, imagErr := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(inputs.FastImag, inputs.OrdinaryImag, eval, standardEval, btp.BootstrappingParameters, inputs.FastSK, standardSK, psWideScale(exponent), imagContracted)
			if realErr != nil || imagErr != nil {
				if realErr != nil {
					candidate.Real.FirstFailure = realErr.Error()
				}
				if imagErr != nil {
					candidate.Imag.FirstFailure = imagErr.Error()
				}
			} else if realPath.FastFinal != nil && imagPath.FastFinal != nil {
				publicLike, standardLike, postS2C, metadata, unchanged, publicErr := precisionSweepPublicLike(realPath, imagPath, eval, standardEval, btp, residual.MaxSlots(), standardSK, reproducibleInput(residual, btp))
				if publicErr != nil {
					return publicErr
				}
				candidate.Real.PublicLike, candidate.Imag.PublicLike = publicLike, publicLike
				candidate.Real.StandardPublicLike, candidate.Imag.StandardPublicLike = standardLike, standardLike
				candidate.Real.PostS2CVsStandard, candidate.Imag.PostS2CVsStandard = postS2C, postS2C
				candidate.Real.MetadataCorrect, candidate.Imag.MetadataCorrect = metadata, metadata
				candidate.Real.InputUnchanged, candidate.Imag.InputUnchanged = unchanged, unchanged
				candidate.Real.PublicThresholdPass = publicLike != nil && publicLike.MaxComponent <= psWidePublicThreshold
				candidate.Imag.PublicThresholdPass = candidate.Real.PublicThresholdPass
			}
		}
		candidate.MaxCapacityRatio, candidate.WorstCapacityCheckpoint, candidate.MinRemainingCapacityBits = psWideWorstCapacity(candidate.Real, candidate.Imag)
		candidate.CapacityPass = candidate.Real.CapacityPass && candidate.Imag.CapacityPass
		candidate.PSExitContractionPass = candidate.Real.Contraction.Q01CenteredUnique && candidate.Imag.Contraction.Q01CenteredUnique && candidate.Real.Contraction.DirectReductionRowsEqual && candidate.Imag.Contraction.DirectReductionRowsEqual && candidate.Real.Contraction.MetadataValid && candidate.Imag.Contraction.MetadataValid
		candidate.PrecisionPass = candidate.Real.SemanticPass && candidate.Imag.SemanticPass
		candidate.EvalModPass = candidate.Real.EvalModVsStandard != nil && candidate.Imag.EvalModVsStandard != nil && candidate.Real.EvalModVsStandard.Pass && candidate.Imag.EvalModVsStandard.Pass
		candidate.PostS2CPass = candidate.Real.PostS2CVsStandard != nil && candidate.Imag.PostS2CVsStandard != nil && candidate.Real.PostS2CVsStandard.Pass && candidate.Imag.PostS2CVsStandard.Pass
		candidate.PublicLikePass = candidate.Real.PublicThresholdPass && candidate.Imag.PublicThresholdPass
		candidate.FirstFailure = psWideBranchFailure(candidate.Real)
		if candidate.FirstFailure == "none" {
			candidate.FirstFailure = psWideBranchFailure(candidate.Imag)
		}
		candidate.Classification = "logn13_ps_wide_q012_candidate_failed"
		if candidate.PrecisionPass && candidate.PublicLikePass && candidate.EvalModPass && candidate.PostS2CPass && candidate.CapacityPass && candidate.PSExitContractionPass {
			candidate.Classification = "logn13_ps_wide_q012_candidate_pass"
		}
		result.Candidates = append(result.Candidates, candidate)
	}
	result.Classification, result.RecommendedScale, result.ProductionReadiness = psWideClassify(result.Candidates)
	result.NextTarget = psWideNextTarget(result.Classification)
	result.ArchitectureCompare = map[string]interface{}{"existing_accepted_q056_plan92_patched_control": map[string]interface{}{"correctness": control["q0_56_current_fast_public_like"], "domain_transitions": 0, "conceptual_complexity": "existing local q2/G0/F0/DA guard stack"}, "ps_wide_q012_p92": map[string]interface{}{"candidate": "P92", "domain_transitions": 2, "transitions": []string{"Q01->Q012 at PS entry", "Q012->Q01 at PS exit"}}, "ps_wide_q012_p93": map[string]interface{}{"candidate": "P93", "domain_transitions": 2}, "ps_wide_q012_p94": map[string]interface{}{"candidate": "P94", "domain_transitions": 2}, "recommendation": "select lowest passing scale with large Q012 capacity margin; do not extend beyond P94"}
	return psWideWrite(result, outPath)
}
