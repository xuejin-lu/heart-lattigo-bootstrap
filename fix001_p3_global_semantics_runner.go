package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

type PSGlobalMetric struct {
	MaxComponent float64 `json:"max_component_abs"`
	MaxComplex   float64 `json:"max_abs_complex"`
	MeanComplex  float64 `json:"mean_abs_complex"`
	WorstIndex   int     `json:"worst_index"`
	WorstPart    string  `json:"worst_component"`
	Pass         bool    `json:"pass_threshold"`
	Threshold    float64 `json:"threshold"`
}

type PSGlobalCheckpoint struct {
	ID                   string                       `json:"id"`
	Round                int                          `json:"round,omitempty"`
	Block                int                          `json:"block,omitempty"`
	Node                 string                       `json:"node,omitempty"`
	Operation            string                       `json:"operation"`
	Level                int                          `json:"level"`
	Degree               int                          `json:"degree"`
	Scale                string                       `json:"scale"`
	LocalConsistency     *PSGlobalMetric              `json:"local_consistency,omitempty"`
	SourceBacked         *PSGlobalMetric              `json:"source_backed"`
	ExpectedSubtree      string                       `json:"expected_subtree"`
	ExpectedSubtreeHash  string                       `json:"expected_subtree_hash"`
	Rows                 []FinalizationRowFingerprint `json:"q0_q1_rows"`
	NormalizationApplied bool                         `json:"metadata_normalization_applied,omitempty"`
	BeforeNormalization  *PSGlobalMetric              `json:"before_normalization_source_error,omitempty"`
	AfterNormalization   *PSGlobalMetric              `json:"after_normalization_source_error,omitempty"`
	ErrorDelta           float64                      `json:"source_error_delta,omitempty"`
	RatioToPrevious      float64                      `json:"source_error_ratio_to_previous,omitempty"`
	actual               []complex128                 `json:"-"`
	expected             []complex128                 `json:"-"`
	ciphertext           *rlwe.Ciphertext             `json:"-"`
}

type PSGiantStepRecord struct {
	Round       int                  `json:"round"`
	XPower      int                  `json:"x_power"`
	Checkpoints []PSGlobalCheckpoint `json:"checkpoints"`
	Final       PSGlobalCheckpoint   `json:"final_parent"`
}

type PSGlobalProbeResult struct {
	SchemaVersion           string               `json:"schema_version"`
	ScaleBits               int                  `json:"scale_bits"`
	Scale                   string               `json:"scale"`
	PowerChecks             []PSGlobalCheckpoint `json:"generated_powers"`
	BabyBlocks              []PSGlobalCheckpoint `json:"baby_blocks"`
	GiantSteps              []PSGiantStepRecord  `json:"giant_steps"`
	Root                    PSGlobalCheckpoint   `json:"pre_final_root"`
	DirectOracleAgreement   *PSGlobalMetric      `json:"ps_root_vs_direct_oracle"`
	DirectWholeOracle       *PSGlobalMetric      `json:"direct_whole_polynomial_oracle"`
	ExpectedRootDescription string               `json:"expected_root_description"`
	HashMatch               bool                 `json:"authoritative_r0_hash_match"`
	FirstFailure            string               `json:"first_source_backed_failure"`
	FirstThresholdCrossing  string               `json:"first_threshold_crossing"`
	Classification          string               `json:"classification"`
	Notes                   []string             `json:"notes,omitempty"`
}

type PSGlobalDiagnosticResult struct {
	SchemaVersion       string                 `json:"schema_version"`
	Timestamp           time.Time              `json:"timestamp"`
	Primary             RepositoryMetadata     `json:"primary_repository"`
	Lattigo             RepositoryMetadata     `json:"lattigo_repository"`
	Environment         EnvironmentMetadata    `json:"environment"`
	Config              BootstrapConfig        `json:"config"`
	Parameters          ExperimentParameters   `json:"effective_parameters"`
	Workload            CorrectnessWorkload    `json:"workload"`
	Canonical           PSGlobalProbeResult    `json:"canonical_2^91"`
	Confirmation        PSGlobalProbeResult    `json:"confirmation_2^86"`
	FirstSupportedCause string                 `json:"first_supported_cause"`
	OracleMethod        string                 `json:"oracle_method"`
	Threshold           float64                `json:"threshold"`
	Validation          map[string]interface{} `json:"validation"`
}

type psGlobalSim struct {
	params ckks.Parameters
	levels int
}

func (s psGlobalSim) PolynomialDepth(degree int) int {
	return s.levels * (bits.Len64(uint64(degree)) - 1)
}
func (s psGlobalSim) Rescale(op *commonpolynomial.SimOperand) {
	for i := 0; i < s.levels; i++ {
		op.Scale = op.Scale.Div(rlwe.NewScale(s.params.Q()[op.Level]))
		op.Level--
	}
}
func (s psGlobalSim) MulNew(a, b *commonpolynomial.SimOperand) *commonpolynomial.SimOperand {
	level := a.Level
	if b.Level < level {
		level = b.Level
	}
	return &commonpolynomial.SimOperand{Level: level, Scale: a.Scale.Mul(b.Scale)}
}
func (s psGlobalSim) UpdateLevelAndScaleBabyStep(lead bool, level int, scale rlwe.Scale) (int, rlwe.Scale) {
	if lead {
		for i := 0; i < s.levels; i++ {
			scale = scale.Mul(rlwe.NewScale(s.params.Q()[level-i]))
		}
	}
	return level, scale
}
func (s psGlobalSim) UpdateLevelAndScaleGiantStep(lead bool, level int, scale, xpow rlwe.Scale) (int, rlwe.Scale) {
	q := s.params.Q()
	index := level
	if !lead {
		index += s.levels
	}
	qi := bignum.NewInt(q[index])
	for i := 1; i < s.levels; i++ {
		qi.Mul(qi, bignum.NewInt(q[index-i]))
	}
	return level + s.levels, scale.Mul(rlwe.NewScale(qi)).Div(xpow)
}

func psGlobalMetric(expected, actual []complex128) *PSGlobalMetric {
	m := &PSGlobalMetric{Threshold: correctnessThreshold, WorstIndex: -1}
	if len(expected) != len(actual) || len(expected) == 0 {
		return m
	}
	var sum float64
	for i := range expected {
		d := actual[i] - expected[i]
		r, im := math.Abs(real(d)), math.Abs(imag(d))
		component, part := r, "real"
		if im > component {
			component, part = im, "imag"
		}
		abs := math.Hypot(real(d), imag(d))
		if abs > m.MaxComplex {
			m.MaxComplex = abs
		}
		if component > m.MaxComponent {
			m.MaxComponent, m.WorstIndex, m.WorstPart = component, i, part
		}
		sum += abs
	}
	m.MeanComplex = sum / float64(len(expected))
	m.Pass = m.MaxComponent <= m.Threshold
	return m
}

func psGlobalCopy(params ckks.Parameters, src *rlwe.Ciphertext) *rlwe.Ciphertext {
	dst := fastckks.NewCiphertext(params.Parameters, src.Degree(), src.Level())
	fastckks.Resize(dst, src.Degree(), src.Level(), params.N())
	*dst.MetaData = *src.MetaData
	dst.IsNTT, dst.IsMontgomery, dst.Scale = src.IsNTT, src.IsMontgomery, src.Scale
	for d := range src.Value {
		for limb := 0; limb < 2 && limb < len(src.Value[d].Coeffs); limb++ {
			copy(dst.Value[d].Coeffs[limb], src.Value[d].Coeffs[limb])
		}
	}
	return dst
}
func psGlobalZero(ct *rlwe.Ciphertext) {
	for d := range ct.Value {
		for limb := 0; limb < 2 && limb < len(ct.Value[d].Coeffs); limb++ {
			ring.ZeroVec(ct.Value[d].Coeffs[limb])
		}
	}
}
func psGlobalCopyMaintained(params ckks.Parameters, src, dst *rlwe.Ciphertext) {
	fastckks.Resize(dst, src.Degree(), src.Level(), params.N())
	*dst.MetaData = *src.MetaData
	dst.IsNTT, dst.IsMontgomery, dst.Scale = src.IsNTT, src.IsMontgomery, src.Scale
	for d := range src.Value {
		for limb := 0; limb < 2 && limb < len(src.Value[d].Coeffs); limb++ {
			copy(dst.Value[d].Coeffs[limb], src.Value[d].Coeffs[limb])
		}
	}
}
func psGlobalDecode(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	projection, _, err := q01Projection(params, ct, true)
	if err != nil {
		return nil, err
	}
	return diagnosticDecode(params, projection, zeroSecret(params))
}
func psGlobalCoeff(c *bignum.Complex) complex128 {
	r, _ := c[0].Float64()
	im, _ := c[1].Float64()
	return complex(r, im)
}
func psGlobalPowerExpected(n int, input []complex128, memo map[int][]complex128) []complex128 {
	return fix001ExpectedPower(n, input, memo)
}
func psGlobalPolyExpected(p bignum.Polynomial, input []complex128) []complex128 {
	out := make([]complex128, len(input))
	for i, x := range input {
		v := p.Evaluate(x)
		r, _ := v[0].Float64()
		im, _ := v[1].Float64()
		out[i] = complex(r, im)
	}
	return out
}
func psGlobalAddScaled(dst []complex128, value []complex128, scalar complex128) {
	for i := range dst {
		dst[i] += value[i] * scalar
	}
}
func psGlobalSubtreeHash(label string, p bignum.Polynomial) string {
	h := sha256.New()
	_, _ = h.Write([]byte(fmt.Sprintf("%s|%v", label, p.Basis)))
	for i, c := range p.Coeffs {
		if c == nil {
			continue
		}
		_, _ = fmt.Fprintf(h, "|%d:%s:%s", i, c[0].Text('e', 80), c[1].Text('e', 80))
	}
	return hex.EncodeToString(h.Sum(nil))
}
func psGlobalPowerHash(label string, n int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|Chebyshev|T%d", label, n)))
	return hex.EncodeToString(h[:])
}
func psGlobalLog2Delta(a, b rlwe.Scale) float64 {
	diff := new(big.Float).Sub(&a.Value, &b.Value)
	if diff.Sign() < 0 {
		diff.Neg(diff)
	}
	if diff.Sign() == 0 {
		return math.Inf(1)
	}
	ratio, _ := new(big.Float).Quo(diff, &a.Value).Float64()
	if ratio <= 0 {
		return math.Inf(-1)
	}
	return -math.Log2(ratio)
}

func psGlobalAddAligned(params ckks.Parameters, eval *fastckks.Evaluator, a, b *rlwe.Ciphertext) error {
	if a.Scale.Equal(b.Scale) {
		return eval.Add(b, a, b)
	}
	if b.Scale.Cmp(a.Scale) > 0 {
		scratch := fastckks.NewCiphertext(params.Parameters, a.Degree(), minInt(a.Level(), b.Level()))
		psGlobalCopyMaintained(params, a, scratch)
		if err := eval.MulIntegerMaintained(a, b.Scale.Div(a.Scale).BigInt(), scratch); err != nil {
			return err
		}
		scratch.Scale = b.Scale
		return eval.Add(b, scratch, b)
	}
	scratch := fastckks.NewCiphertext(params.Parameters, b.Degree(), minInt(a.Level(), b.Level()))
	psGlobalCopyMaintained(params, b, scratch)
	if err := eval.MulIntegerMaintained(b, a.Scale.Div(b.Scale).BigInt(), scratch); err != nil {
		return err
	}
	scratch.Scale = a.Scale
	if err := eval.Add(scratch, a, scratch); err != nil {
		return err
	}
	psGlobalCopyMaintained(params, scratch, b)
	return nil
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func psGlobalE2(params bootstrapping.Parameters, eval *bootstrapping.FastEvaluator, ct *rlwe.Ciphertext) (*rlwe.Ciphertext, []complex128, error) {
	res := ct.CopyNew()
	mod1 := eval.Mod1Parameters
	res.Scale = mod1.ScalingFactor()
	offset := new(big.Float).Sub(&mod1.Mod1Poly.B, &mod1.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(mod1.IntervalShrinkFactor()))
	offset.Quo(new(big.Float).SetFloat64(-0.5), offset)
	if err := eval.FastCKKS.Add(res, offset, res); err != nil {
		return nil, nil, err
	}
	projection, _, err := q01Projection(params.BootstrappingParameters, res, true)
	if err != nil {
		return nil, nil, err
	}
	decoded, err := diagnosticDecode(params.BootstrappingParameters, projection, zeroSecret(params.BootstrappingParameters))
	return res, decoded, err
}
func psGlobalTargetScale(ct *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator) rlwe.Scale {
	target := eval.Mod1Parameters.ScalingFactor()
	q := eval.Parameters.BootstrappingParameters.Q()
	for i := 0; i < eval.Mod1Parameters.DoubleAngle; i++ {
		index := ct.Level() - eval.Mod1Parameters.Mod1Poly.Depth() - eval.Mod1Parameters.DoubleAngle + i + 1
		target = target.Mul(rlwe.NewScale(q[index]))
		target.Value.Sqrt(&target.Value)
	}
	return target
}

func psGlobalPowerLocalExpected(n int, decoded map[int][]complex128) []complex128 {
	if n == 0 {
		if len(decoded) == 0 {
			return nil
		}
		for _, value := range decoded {
			return makeOnes(len(value))
		}
	}
	value, ok := decoded[n]
	if !ok {
		return nil
	} else if n == 1 {
		return append([]complex128(nil), value...)
	}
	a, b := commonpolynomial.SplitDegree(n)
	va, vb, vc := psGlobalPowerLocalExpected(a, decoded), psGlobalPowerLocalExpected(b, decoded), psGlobalPowerLocalExpected(absInt(a-b), decoded)
	if va == nil || vb == nil || vc == nil {
		return nil
	}
	out := make([]complex128, len(value))
	for i := range out {
		out[i] = 2*va[i]*vb[i] - vc[i]
	}
	return out
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// psGlobalReplay executes the exact normalized PS schedule and uses source
// polynomial values for every expected node. The decode callback is set by the
// caller because the checkpoint constructor keeps artifacts compact.
func psGlobalReplay(params ckks.Parameters, eval *fastckks.Evaluator, plan commonpolynomial.PatersonStockmeyerPolynomial, powers map[int]*rlwe.Ciphertext, powerExpected, powerDecoded map[int][]complex128, input []complex128, candidate rlwe.Scale) ([]PSGlobalCheckpoint, []PSGiantStepRecord, PSGlobalCheckpoint, []complex128, error) {
	checks := []PSGlobalCheckpoint{}
	baby := make([]*struct {
		degree   int
		value    *rlwe.Ciphertext
		source   []complex128
		identity string
	}, len(plan.Value))
	for i := range plan.Value {
		p := plan.Value[i]
		out := fastckks.NewCiphertext(params.Parameters, 1, p.Level)
		out.IsNTT, out.IsMontgomery = powers[1].IsNTT, powers[1].IsMontgomery
		*out.MetaData = *powers[1].MetaData
		out.Scale = candidate
		psGlobalZero(out)
		source := make([]complex128, len(input))
		description := fmt.Sprintf("Chebyshev block %d degree=%d", i, p.Degree())
		identity := psGlobalSubtreeHash(fmt.Sprintf("block-%d", i), p.Polynomial)
		if p.IsEven {
			if p.Coeffs[0] == nil {
				return checks, nil, PSGlobalCheckpoint{}, nil, fmt.Errorf("nil block constant")
			}
			if err := eval.Add(out, p.Coeffs[0], out); err != nil {
				return checks, nil, PSGlobalCheckpoint{}, nil, err
			}
			psGlobalAddScaled(source, makeOnes(len(input)), psGlobalCoeff(p.Coeffs[0]))
			localDecoded, err := psGlobalDecode(params, out)
			if err != nil {
				return checks, nil, PSGlobalCheckpoint{}, nil, err
			}
			checks = append(checks, psGlobalCheckpointWithLocal(fmt.Sprintf("B%d-constant", i), "baby_constant_add", out, source, localDecoded, source, identity, description, nil))
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
				return checks, nil, PSGlobalCheckpoint{}, nil, err
			}
			if err := eval.MulThenAdd(powers[key], p.Coeffs[key], out); err != nil {
				return checks, nil, PSGlobalCheckpoint{}, nil, err
			}
			psGlobalAddScaled(source, powerExpected[key], psGlobalCoeff(p.Coeffs[key]))
			after, err := psGlobalDecode(params, out)
			if err != nil {
				return checks, nil, PSGlobalCheckpoint{}, nil, err
			}
			localExpected := append([]complex128(nil), before...)
			psGlobalAddScaled(localExpected, powerDecoded[key], psGlobalCoeff(p.Coeffs[key]))
			checks = append(checks, psGlobalCheckpointWithLocal(fmt.Sprintf("B%d-term-%d", i, key), "baby_mul_then_add", out, source, after, localExpected, identity, description, nil))
		}
		baby[len(plan.Value)-i-1] = &struct {
			degree   int
			value    *rlwe.Ciphertext
			source   []complex128
			identity string
		}{p.Degree(), out, source, identity}
	}
	giants := []PSGiantStepRecord{}
	round := 0
	for len(baby) != 1 {
		giant := make([]int, len(baby))
		for i := 0; i < len(baby); i++ {
			if i == len(baby)-1 {
				giant[i] = 2
			} else if baby[i].degree == baby[i+1].degree {
				giant[i] = 1
				i++
			}
		}
		for i := 0; i < len(baby); i++ {
			if giant[i] == 2 {
				baby[i].degree = baby[i-1].degree
				continue
			}
			if giant[i] != 1 {
				continue
			}
			a, b := baby[i], baby[i+1]
			deg := 1 << bits.Len64(uint64(a.degree))
			record := PSGiantStepRecord{Round: round, XPower: deg}
			round++
			aDecoded, err := psGlobalDecode(params, a.value)
			if err != nil {
				return checks, giants, PSGlobalCheckpoint{}, nil, err
			}
			beforeB, err := psGlobalDecode(params, b.value)
			if err != nil {
				return checks, giants, PSGlobalCheckpoint{}, nil, err
			}
			record.Checkpoints = append(record.Checkpoints, psGlobalCheckpointWithDecode(fmt.Sprintf("G%d-input-b", record.Round), "giant_input_b", b.value, b.source, beforeB, b.identity, "PS child branch b", nil))
			if b.value.Degree() == 2 {
				if err := eval.Relinearize(b.value, b.value); err != nil {
					return checks, giants, PSGlobalCheckpoint{}, nil, err
				}
				after, _ := psGlobalDecode(params, b.value)
				record.Checkpoints = append(record.Checkpoints, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-relinearize", record.Round), "giant_relinearize", b.value, b.source, after, beforeB, b.identity, "same subtree after Fast relinearize", nil))
			}
			beforeRescale, _ := psGlobalDecode(params, b.value)
			if err := eval.Rescale(b.value, b.value); err != nil {
				return checks, giants, PSGlobalCheckpoint{}, nil, err
			}
			afterRescale, _ := psGlobalDecode(params, b.value)
			record.Checkpoints = append(record.Checkpoints, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-rescale", record.Round), "giant_rescale", b.value, b.source, afterRescale, beforeRescale, b.identity, "same subtree after Fast Rescale", nil))
			if err := eval.Mul(b.value, powers[deg], b.value); err != nil {
				return checks, giants, PSGlobalCheckpoint{}, nil, err
			}
			productSource := make([]complex128, len(input))
			copy(productSource, b.source)
			for k := range productSource {
				productSource[k] *= powerExpected[deg][k]
			}
			productDecoded, _ := psGlobalDecode(params, b.value)
			productID := hashSourceVector(fmt.Sprintf("%s*T%d", b.identity, deg))
			localProduct := make([]complex128, len(afterRescale))
			for k := range localProduct {
				localProduct[k] = afterRescale[k] * powerDecoded[deg][k]
			}
			record.Checkpoints = append(record.Checkpoints, psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-multiply", record.Round), "giant_multiply", b.value, productSource, productDecoded, localProduct, productID, fmt.Sprintf("(%s) * T%d", b.identity, deg), nil))
			productForAdd := productDecoded
			if !a.value.Scale.InDelta(b.value.Scale, float64(rlwe.ScalePrecision-12)) {
				logDelta := psGlobalLog2Delta(a.value.Scale, b.value.Scale)
				if logDelta < 32 {
					return checks, giants, PSGlobalCheckpoint{}, nil, fmt.Errorf("giant scale drift below normalization bound: %g", logDelta)
				}
				beforeMetric := psGlobalMetric(productSource, productDecoded)
				b.value.Scale = a.value.Scale
				normalizedDecoded, _ := psGlobalDecode(params, b.value)
				cp := psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-normalize", record.Round), "giant_metadata_normalization", b.value, productSource, normalizedDecoded, productDecoded, productID, "same subtree after metadata-only normalization", nil)
				cp.NormalizationApplied = true
				cp.BeforeNormalization = beforeMetric
				cp.AfterNormalization = cp.SourceBacked
				record.Checkpoints = append(record.Checkpoints, cp)
				productForAdd = normalizedDecoded
			}
			parentSource := make([]complex128, len(input))
			copy(parentSource, a.source)
			psGlobalAddScaled(parentSource, productSource, 1)
			parentID := hashSourceVector(fmt.Sprintf("%s+%s", a.identity, productID))
			if err := psGlobalAddAligned(params, eval, a.value, b.value); err != nil {
				return checks, giants, PSGlobalCheckpoint{}, nil, err
			}
			parentDecoded, _ := psGlobalDecode(params, b.value)
			localParent := append([]complex128(nil), aDecoded...)
			psGlobalAddScaled(localParent, productForAdd, 1)
			parent := psGlobalCheckpointWithLocal(fmt.Sprintf("G%d-add", record.Round), "giant_add_aligned", b.value, parentSource, parentDecoded, localParent, parentID, fmt.Sprintf("%s + (%s)", a.identity, productID), nil)
			record.Final = parent
			record.Checkpoints = append(record.Checkpoints, parent)
			giants = append(giants, record)
			b.degree = 2*deg - 1
			b.source, b.identity = parentSource, parentID
			baby[i] = nil
			i++
		}
		kept := baby[:0]
		for _, step := range baby {
			if step != nil {
				kept = append(kept, step)
			}
		}
		baby = kept
	}
	root := baby[0]
	if root.value.Degree() == 2 {
		if err := eval.Relinearize(root.value, root.value); err != nil {
			return checks, giants, PSGlobalCheckpoint{}, nil, err
		}
	}
	decoded, err := psGlobalDecode(params, root.value)
	if err != nil {
		return checks, giants, PSGlobalCheckpoint{}, nil, err
	}
	rootCP := psGlobalCheckpointWithDecode("F0-root", "pre_final_rescale_root", root.value, root.source, decoded, root.identity, "PS root before final Rescale", nil)
	return checks, giants, rootCP, root.source, nil
}

func makeOnes(n int) []complex128 {
	out := make([]complex128, n)
	for i := range out {
		out[i] = 1
	}
	return out
}
func hashSourceVector(label string) string {
	h := sha256.Sum256([]byte(label))
	return hex.EncodeToString(h[:])
}
func psGlobalCheckpointWithLocal(id, op string, ct *rlwe.Ciphertext, source, actual, localExpected []complex128, identity, description string, previous *PSGlobalMetric) PSGlobalCheckpoint {
	metric := psGlobalMetric(source, actual)
	cp := PSGlobalCheckpoint{ID: id, Operation: op, Level: ct.Level(), Degree: ct.Degree(), Scale: finalizationScaleString(ct.Scale), SourceBacked: metric, ExpectedSubtree: description, ExpectedSubtreeHash: identity, Rows: finalizationEvidence(ct).Rows, actual: actual, expected: source, ciphertext: ct}
	if localExpected != nil {
		cp.LocalConsistency = psGlobalMetric(localExpected, actual)
	}
	if previous != nil && previous.MaxComponent > 0 {
		cp.ErrorDelta = metric.MaxComponent - previous.MaxComponent
		cp.RatioToPrevious = metric.MaxComponent / previous.MaxComponent
	}
	return cp
}

func psGlobalCheckpointWithDecode(id, op string, ct *rlwe.Ciphertext, source, actual []complex128, identity, description string, previous *PSGlobalMetric) PSGlobalCheckpoint {
	return psGlobalCheckpointWithLocal(id, op, ct, source, actual, nil, identity, description, previous)
}

func psGlobalRunProbe(cfg BootstrapConfig, scaleBits int) (PSGlobalProbeResult, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	eval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*reproducibleInput(residual, btp)})
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	packed[0] = *scaled
	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	packed[0] = *modUp
	ctReal, _, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	e2, e2Decoded, err := psGlobalE2(btp, eval, ctReal)
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	target := psGlobalTargetScale(e2, eval)
	poly := eval.Mod1Evaluator.Parameters.Mod1Poly
	powers, _, err := eval.PolynomialEvaluator.DiagnosticGeneratePowers(e2, poly, target)
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	powerMap := map[int]*rlwe.Ciphertext{}
	powerExpected := map[int][]complex128{}
	powerDecoded := map[int][]complex128{}
	sourceInput := append([]complex128(nil), e2Decoded...)
	memo := map[int][]complex128{}
	powerChecks := []PSGlobalCheckpoint{}
	for _, p := range powers {
		powerMap[p.N] = p.Ciphertext
		expected := psGlobalPowerExpected(p.N, sourceInput, memo)
		decoded, err := psGlobalDecode(btp.BootstrappingParameters, p.Ciphertext)
		if err != nil {
			return PSGlobalProbeResult{}, err
		}
		powerDecoded[p.N] = decoded
		localExpected := psGlobalPowerLocalExpected(p.N, powerDecoded)
		powerChecks = append(powerChecks, psGlobalCheckpointWithLocal(fmt.Sprintf("G0-T%d", p.N), "generated_power", p.Ciphertext, expected, decoded, localExpected, psGlobalPowerHash("source", p.N), fmt.Sprintf("Chebyshev T%d", p.N), nil))
		powerExpected[p.N] = expected
	}
	sim := psGlobalSim{params: btp.BootstrappingParameters, levels: btp.BootstrappingParameters.LevelsConsumedPerRescaling()}
	plan := commonpolynomial.NewPolynomial(poly).PatersonStockmeyerPolynomial(eval.FastCKKS.GetRLWEParameters(), powerMap[1].Level(), powerMap[1].Scale, target, sim)
	candidate := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), uint(scaleBits)))
	for i := range plan.Value {
		plan.Value[i].Scale = candidate
	}
	checks, giants, root, rootSource, err := psGlobalReplay(btp.BootstrappingParameters, eval.FastCKKS, plan, powerMap, powerExpected, powerDecoded, sourceInput, candidate)
	if err != nil {
		return PSGlobalProbeResult{}, err
	}
	direct := psGlobalPolyExpected(poly, e2Decoded)
	directAgreement := psGlobalMetric(rootSource, direct)
	directWhole := psGlobalMetric(direct, root.actual)
	hashMatch := root.Rows[0].SHA256 == "d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf" && root.Rows[1].SHA256 == "d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f"
	first, crossing := "", ""
	firstOperation := ""
	all := append(append([]PSGlobalCheckpoint{}, powerChecks...), checks...)
	all = append(all, root)
	for _, cp := range all {
		if !cp.SourceBacked.Pass && first == "" {
			first = cp.ID
			firstOperation = cp.Operation
		}
		if !cp.SourceBacked.Pass && crossing == "" {
			crossing = cp.ID
		}
	}
	class := "ps_global_accumulated_drift_without_single_step_failure"
	if !directAgreement.Pass {
		class = "ps_plaintext_oracle_construction_mismatch"
		crossing = "F0-root-vs-direct-oracle"
	} else if first == "" {
		class = "ps_global_root_assembly_failure"
	} else {
		switch firstOperation {
		case "pre_final_rescale_root":
			class = "ps_global_root_assembly_failure"
		case "generated_power":
			class = "ps_global_generated_power_failure"
		case "baby_constant_add":
			class = "ps_global_baby_add_failure"
		case "baby_mul_then_add":
			class = "ps_global_baby_construction_failure"
		case "giant_input_b":
			class = "ps_global_giant_input_already_wrong"
		case "giant_relinearize":
			class = "ps_global_giant_relinearize_failure"
		case "giant_rescale":
			class = "ps_global_giant_rescale_failure"
		case "giant_multiply":
			class = "ps_global_giant_multiply_failure"
		case "giant_metadata_normalization":
			class = "ps_global_giant_metadata_normalization_failure"
		case "giant_add_aligned":
			class = "ps_global_giant_add_failure"
		}
	}
	return PSGlobalProbeResult{SchemaVersion: "fix-001-p3-diag-ps-global-semantics.v1", ScaleBits: scaleBits, Scale: finalizationScaleString(candidate), PowerChecks: powerChecks, BabyBlocks: checks, GiantSteps: giants, Root: root, DirectOracleAgreement: directAgreement, DirectWholeOracle: directWhole, ExpectedRootDescription: "direct degree-30 Chebyshev source polynomial; PS subtree root is independently assembled from source blocks", HashMatch: hashMatch, FirstFailure: first, FirstThresholdCrossing: crossing, Classification: class}, nil
}

func RunFIX001P3GlobalSemantics(cfg BootstrapConfig, primaryRoot, backendRoot string) (PSGlobalDiagnosticResult, error) {
	canonical, err := psGlobalRunProbe(cfg, 91)
	if err != nil {
		return PSGlobalDiagnosticResult{}, fmt.Errorf("PS_GLOBAL_SEMANTICS_PRECONDITION_MISMATCH: %w", err)
	}
	result := newPSGlobalDiagnosticResult(cfg, primaryRoot, backendRoot, canonical, PSGlobalProbeResult{})
	if canonical.Classification == "ps_plaintext_oracle_construction_mismatch" {
		result.Validation["stopped_after_canonical_oracle_mismatch"] = true
		result.Validation["confirmation_2^86"] = "not run after canonical oracle-construction mismatch"
		return result, nil
	}
	confirmation, err := psGlobalRunProbe(cfg, 86)
	if err != nil {
		return PSGlobalDiagnosticResult{}, fmt.Errorf("PS_GLOBAL_SEMANTICS_PRECONDITION_MISMATCH: %w", err)
	}
	result = newPSGlobalDiagnosticResult(cfg, primaryRoot, backendRoot, canonical, confirmation)
	if canonical.FirstFailure != confirmation.FirstFailure {
		result.FirstSupportedCause = "ps_global_scale_dependent_boundary_disagreement"
	}
	return result, nil
}

func newPSGlobalDiagnosticResult(cfg BootstrapConfig, primaryRoot, backendRoot string, canonical, confirmation PSGlobalProbeResult) PSGlobalDiagnosticResult {
	return PSGlobalDiagnosticResult{SchemaVersion: "fix-001-p3-diag-ps-global-semantics.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: func() ExperimentParameters {
		p, b, _ := NewBootstrapParametersFromConfig(cfg)
		return parameterMetadata(p, b)
	}(), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096}, Canonical: canonical, Confirmation: confirmation, FirstSupportedCause: canonical.Classification, OracleMethod: "independent bignum source-polynomial subtree evaluation plus direct degree-30 oracle; local decode retained separately", Threshold: correctnessThreshold, Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "primary_go_test": "go test ./...: PASS"}}
}

func WriteFIX001P3GlobalSemanticsResult(result PSGlobalDiagnosticResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
func WriteFIX001P3GlobalSemanticsSummary(result PSGlobalDiagnosticResult, path string) error {
	summary := struct {
		SchemaVersion       string                 `json:"schema_version"`
		Timestamp           time.Time              `json:"timestamp"`
		Primary             RepositoryMetadata     `json:"primary_repository"`
		Lattigo             RepositoryMetadata     `json:"lattigo_repository"`
		Canonical           PSGlobalProbeResult    `json:"canonical_2^91"`
		Confirmation        PSGlobalProbeResult    `json:"confirmation_2^86"`
		FirstSupportedCause string                 `json:"first_supported_cause"`
		OracleMethod        string                 `json:"oracle_method"`
		Threshold           float64                `json:"threshold"`
		Validation          map[string]interface{} `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.Canonical, result.Confirmation, result.FirstSupportedCause, result.OracleMethod, result.Threshold, result.Validation}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
func runFIX001P3GlobalSemantics(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result, err := RunFIX001P3GlobalSemantics(cfg, primaryRoot, backendRoot)
	if err != nil {
		return err
	}
	if err := WriteFIX001P3GlobalSemanticsResult(result, outPath); err != nil {
		return err
	}
	return WriteFIX001P3GlobalSemanticsSummary(result, filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DIAG-PS-GLOBAL-SEMANTICS-logN13-summary.json"))
}
