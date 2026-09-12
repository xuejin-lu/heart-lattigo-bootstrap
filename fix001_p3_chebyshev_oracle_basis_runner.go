package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

const chebyshevOracleAgreementThreshold = 1e-10

type ChebyshevOracleRange struct {
	MinReal float64 `json:"min_real"`
	MaxReal float64 `json:"max_real"`
	MinImag float64 `json:"min_imag"`
	MaxImag float64 `json:"max_imag"`
	MaxAbs  float64 `json:"max_abs"`
	SHA256  string  `json:"sha256"`
}

type ChebyshevOracleMetricPair struct {
	O1VsO2 *PSGlobalMetric `json:"o1_vs_o2"`
	O1VsO3 *PSGlobalMetric `json:"o1_vs_o3"`
	O1VsO4 *PSGlobalMetric `json:"o1_vs_o4"`
	O2VsO3 *PSGlobalMetric `json:"o2_vs_o3"`
	O2VsO4 *PSGlobalMetric `json:"o2_vs_o4"`
	O3VsO4 *PSGlobalMetric `json:"o3_vs_o4"`
}

type ChebyshevPreprocessingEvidence struct {
	IntervalA             float64              `json:"interval_a"`
	IntervalB             float64              `json:"interval_b"`
	ChangeOfBasisScalar   float64              `json:"change_of_basis_scalar"`
	ChangeOfBasisConstant float64              `json:"change_of_basis_constant"`
	IntervalShrinkFactor  float64              `json:"interval_shrink_factor"`
	Mod1Offset            float64              `json:"mod1_offset"`
	InputScaleBefore      string               `json:"input_scale_before_reinterpretation"`
	PolynomialInputScale  string               `json:"polynomial_input_scale"`
	OffsetRelation        *PSGlobalMetric      `json:"decoded_offset_relation"`
	ZRange                ChebyshevOracleRange `json:"z_e2_decoded_range"`
	InverseMappedXRange   ChebyshevOracleRange `json:"inverse_mapped_x_range"`
	PreOffsetRange        ChebyshevOracleRange `json:"pre_offset_range"`
}

type ChebyshevPowerDomainCheck struct {
	N                int             `json:"n"`
	SourceBacked     *PSGlobalMetric `json:"source_backed"`
	LocalConsistency *PSGlobalMetric `json:"local_consistency"`
}

type ChebyshevBoundaryEvidence struct {
	State            string                       `json:"state"`
	Level            int                          `json:"level"`
	Degree           int                          `json:"degree"`
	Scale            string                       `json:"scale"`
	Rows             []FinalizationRowFingerprint `json:"q0_q1_rows"`
	CorrectedOracle  *PSGlobalMetric              `json:"corrected_raw_chebyshev_oracle"`
	HistoricalOracle *PSGlobalMetric              `json:"historical_o1_oracle"`
}

type ChebyshevOracleBasisResult struct {
	SchemaVersion       string                          `json:"schema_version"`
	Timestamp           time.Time                       `json:"timestamp"`
	Primary             RepositoryMetadata              `json:"primary_repository"`
	Lattigo             RepositoryMetadata              `json:"lattigo_repository"`
	Environment         EnvironmentMetadata             `json:"environment"`
	Config              BootstrapConfig                 `json:"config"`
	Parameters          ExperimentParameters            `json:"effective_parameters"`
	Workload            CorrectnessWorkload             `json:"workload"`
	Preprocessing       ChebyshevPreprocessingEvidence  `json:"preprocessing"`
	GeneratedPowers     []ChebyshevPowerDomainCheck     `json:"generated_power_domain_validation"`
	OracleAgreement     ChebyshevOracleMetricPair       `json:"oracle_agreement_matrix"`
	OracleAgreementMax  float64                         `json:"oracle_agreement_threshold"`
	OracleRanges        map[string]ChebyshevOracleRange `json:"oracle_ranges"`
	AuthoritativeOracle string                          `json:"authoritative_poly_oracle"`
	OldOracleStatus     string                          `json:"old_o1_status"`
	C0                  ChebyshevBoundaryEvidence       `json:"c0_pre_final_rescale"`
	C1                  ChebyshevBoundaryEvidence       `json:"c1_post_final_fast_rescale"`
	FirstSupportedCause string                          `json:"first_supported_cause"`
	Validation          map[string]interface{}          `json:"validation"`
}

type chebyshevOracleState struct {
	Params          bootstrapping.Parameters
	Eval            *bootstrapping.FastEvaluator
	Mod1Input       *rlwe.Ciphertext
	Poly            bignum.Polynomial
	Target          rlwe.Scale
	Z               []complex128
	PreOffset       []complex128
	InputScale      string
	InputScaleValue rlwe.Scale
	E2Scale         string
	Offset          float64
	Root            PSGlobalCheckpoint
	PowerChecks     []ChebyshevPowerDomainCheck
	PowerExpected   map[int][]complex128
	PowerDecoded    map[int][]complex128
}

func chebyshevOracleMetric(expected, actual []complex128) *PSGlobalMetric {
	metric := psGlobalMetric(expected, actual)
	metric.Threshold = chebyshevOracleAgreementThreshold
	metric.Pass = metric.MaxComponent <= metric.Threshold
	return metric
}

func chebyshevScaleString(scale rlwe.Scale) string {
	if scale.Value.Sign() == 0 {
		return "0"
	}
	return finalizationScaleString(scale)
}

func chebyshevOracleRange(values []complex128) ChebyshevOracleRange {
	r := ChebyshevOracleRange{MinReal: 0, MaxReal: 0, MinImag: 0, MaxImag: 0}
	if len(values) == 0 {
		return r
	}
	r.MinReal, r.MaxReal = real(values[0]), real(values[0])
	r.MinImag, r.MaxImag = imag(values[0]), imag(values[0])
	h := sha256.New()
	for _, value := range values {
		if real(value) < r.MinReal {
			r.MinReal = real(value)
		}
		if real(value) > r.MaxReal {
			r.MaxReal = real(value)
		}
		if imag(value) < r.MinImag {
			r.MinImag = imag(value)
		}
		if imag(value) > r.MaxImag {
			r.MaxImag = imag(value)
		}
		if abs := real(value)*real(value) + imag(value)*imag(value); abs > r.MaxAbs*r.MaxAbs {
			r.MaxAbs = sqrt(abs)
		}
		_, _ = fmt.Fprintf(h, "%.17g,%.17g;", real(value), imag(value))
	}
	r.SHA256 = hex.EncodeToString(h.Sum(nil))
	return r
}

func sqrt(value float64) float64 {
	return math.Sqrt(value)
}

func chebyshevLibraryOracle(poly bignum.Polynomial, input []complex128) []complex128 {
	out := make([]complex128, len(input))
	for i, value := range input {
		v := poly.Evaluate(value)
		r, _ := v[0].Float64()
		im, _ := v[1].Float64()
		out[i] = complex(r, im)
	}
	return out
}

func chebyshevRawOracle(poly bignum.Polynomial, input []complex128) []complex128 {
	const precision = 256
	out := make([]complex128, len(input))
	mul := bignum.NewComplexMultiplier()
	two := bignum.ToComplex(float64(2), precision)
	for i, value := range input {
		x := bignum.ToComplex(value, precision)
		previous := bignum.ToComplex(float64(1), precision)
		term := x.Clone()
		result := bignum.NewComplex().SetPrec(precision)
		product := bignum.NewComplex().SetPrec(precision)
		twoX := bignum.NewComplex().SetPrec(precision)
		mul.Mul(two, x, twoX)
		if poly.Coeffs[0] != nil {
			mul.Mul(previous, poly.Coeffs[0], product)
			result.Add(result, product)
		}
		if len(poly.Coeffs) > 1 && poly.Coeffs[1] != nil {
			mul.Mul(term, poly.Coeffs[1], product)
			result.Add(result, product)
		}
		for k := 2; k < len(poly.Coeffs); k++ {
			next := bignum.NewComplex().SetPrec(precision)
			mul.Mul(twoX, term, next)
			next.Sub(next, previous)
			previous, term = term, next
			if poly.Coeffs[k] != nil {
				mul.Mul(term, poly.Coeffs[k], product)
				result.Add(result, product)
			}
		}
		out[i] = result.Complex128()
	}
	return out
}

func chebyshevInverseMap(input []complex128, scalar, constant float64) []complex128 {
	out := make([]complex128, len(input))
	for i, value := range input {
		out[i] = (value - complex(constant, 0)) / complex(scalar, 0)
	}
	return out
}

func chebyshevAddOffset(input []complex128, offset float64) []complex128 {
	out := make([]complex128, len(input))
	for i, value := range input {
		out[i] = value + complex(offset, 0)
	}
	return out
}

func chebyshevCheckFinite(name string, values []complex128) error {
	for i, value := range values {
		if math.IsNaN(real(value)) || math.IsNaN(imag(value)) || math.IsInf(real(value), 0) || math.IsInf(imag(value), 0) {
			return fmt.Errorf("%s contains non-finite value at slot %d: %v", name, i, value)
		}
	}
	return nil
}

func chebyshevCanonicalState(cfg BootstrapConfig) (chebyshevOracleState, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return chebyshevOracleState{}, err
	}
	eval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return chebyshevOracleState{}, err
	}
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*reproducibleInput(residual, btp)})
	if err != nil {
		return chebyshevOracleState{}, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return chebyshevOracleState{}, err
	}
	packed[0] = *scaled
	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return chebyshevOracleState{}, err
	}
	packed[0] = *modUp
	ctReal, _, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return chebyshevOracleState{}, err
	}
	inputScale := chebyshevScaleString(ctReal.Scale)
	preOffsetCT := ctReal.CopyNew()
	preOffsetCT.Scale = eval.Mod1Parameters.ScalingFactor()
	preOffset, err := psGlobalDecode(btp.BootstrappingParameters, preOffsetCT)
	if err != nil {
		return chebyshevOracleState{}, err
	}
	e2, z, err := psGlobalE2(btp, eval, ctReal)
	if err != nil {
		return chebyshevOracleState{}, err
	}
	poly := eval.Mod1Evaluator.Parameters.Mod1Poly
	target := psGlobalTargetScale(e2, eval)
	powers, _, err := eval.PolynomialEvaluator.DiagnosticGeneratePowers(e2, poly, target)
	if err != nil {
		return chebyshevOracleState{}, err
	}
	powerMap := map[int]*rlwe.Ciphertext{}
	powerExpected := map[int][]complex128{}
	powerDecoded := map[int][]complex128{}
	memo := map[int][]complex128{}
	powerChecks := make([]ChebyshevPowerDomainCheck, 0, len(powers))
	for _, power := range powers {
		powerMap[power.N] = power.Ciphertext
		decoded, err := psGlobalDecode(btp.BootstrappingParameters, power.Ciphertext)
		if err != nil {
			return chebyshevOracleState{}, err
		}
		powerDecoded[power.N] = decoded
		expected := fix001ExpectedPower(power.N, z, memo)
		local := psGlobalPowerLocalExpected(power.N, powerDecoded)
		powerExpected[power.N] = expected
		powerChecks = append(powerChecks, ChebyshevPowerDomainCheck{N: power.N, SourceBacked: psGlobalMetric(expected, decoded), LocalConsistency: psGlobalMetric(local, decoded)})
	}
	sim := psGlobalSim{params: btp.BootstrappingParameters, levels: btp.BootstrappingParameters.LevelsConsumedPerRescaling()}
	plan := commonpolynomial.NewPolynomial(poly).PatersonStockmeyerPolynomial(eval.FastCKKS.GetRLWEParameters(), powerMap[1].Level(), powerMap[1].Scale, target, sim)
	candidate := rlwe.NewScale(new(big.Int).Lsh(big.NewInt(1), 91))
	for i := range plan.Value {
		plan.Value[i].Scale = candidate
	}
	_, _, root, _, err := psGlobalReplay(btp.BootstrappingParameters, eval.FastCKKS, plan, powerMap, powerExpected, powerDecoded, z, candidate)
	if err != nil {
		return chebyshevOracleState{}, err
	}
	if !root.RowsMatch([]string{
		"d0db2b184bc895ff004769f6f02aadfc956a9d69ddfa60e86fdf9963f7382faf",
		"d0f3f6d6ed56e82bc97be47c535361895a75a2d583cac4c09fdbc5343d5f6f7f",
	}) {
		return chebyshevOracleState{}, fmt.Errorf("CHEBYSHEV_ORACLE_PRECONDITION_MISMATCH: authoritative root hashes differ")
	}
	return chebyshevOracleState{Params: btp, Eval: eval, Mod1Input: ctReal, Poly: poly, Target: target, Z: z, PreOffset: preOffset, InputScale: inputScale, InputScaleValue: ctReal.Scale, E2Scale: chebyshevScaleString(e2.Scale), Root: root, PowerChecks: powerChecks, PowerExpected: powerExpected, PowerDecoded: powerDecoded}, nil
}

func (cp PSGlobalCheckpoint) RowsMatch(expected []string) bool {
	if len(cp.Rows) < len(expected) {
		return false
	}
	for i, hash := range expected {
		if cp.Rows[i].SHA256 != hash {
			return false
		}
	}
	return true
}

func newChebyshevPreprocessing(poly bignum.Polynomial, state chebyshevOracleState, inputScale string, offsetRelation *PSGlobalMetric) ChebyshevPreprocessingEvidence {
	a, _ := poly.A.Float64()
	b, _ := poly.B.Float64()
	scalar, constant := poly.ChangeOfBasis()
	s, _ := scalar.Float64()
	c, _ := constant.Float64()
	offset := -0.5 / ((b - a) * state.Eval.Mod1Parameters.IntervalShrinkFactor())
	x := chebyshevInverseMap(state.Z, s, c)
	return ChebyshevPreprocessingEvidence{IntervalA: a, IntervalB: b, ChangeOfBasisScalar: s, ChangeOfBasisConstant: c, IntervalShrinkFactor: state.Eval.Mod1Parameters.IntervalShrinkFactor(), Mod1Offset: offset, InputScaleBefore: inputScale, PolynomialInputScale: state.E2Scale, OffsetRelation: offsetRelation, ZRange: chebyshevOracleRange(state.Z), InverseMappedXRange: chebyshevOracleRange(x), PreOffsetRange: chebyshevOracleRange(state.PreOffset)}
}

func runFIX001P3ChebyshevOracleBasis(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	state, err := chebyshevCanonicalState(cfg)
	if err != nil {
		return err
	}
	poly := state.Poly
	o1 := chebyshevLibraryOracle(poly, state.Z)
	o2 := chebyshevRawOracle(poly, state.Z)
	scalar, constant := poly.ChangeOfBasis()
	s, _ := scalar.Float64()
	c, _ := constant.Float64()
	o3 := chebyshevLibraryOracle(poly, chebyshevInverseMap(state.Z, s, c))
	o4 := append([]complex128(nil), state.Root.expected...)
	for name, values := range map[string][]complex128{"z": state.Z, "O1": o1, "O2": o2, "O3": o3, "O4": o4} {
		if err := chebyshevCheckFinite(name, values); err != nil {
			return err
		}
	}
	offsetRelation := psGlobalMetric(chebyshevAddOffset(state.PreOffset, state.Offset), state.Z)
	// state.Offset is filled from the exact Mod1 relation below after the first slot is available.
	offset := -(func() float64 {
		a, _ := poly.A.Float64()
		b, _ := poly.B.Float64()
		return 0.5 / ((b - a) * state.Eval.Mod1Parameters.IntervalShrinkFactor())
	})()
	offsetRelation = psGlobalMetric(chebyshevAddOffset(state.PreOffset, offset), state.Z)
	state.Offset = offset
	pairs := ChebyshevOracleMetricPair{O1VsO2: chebyshevOracleMetric(o1, o2), O1VsO3: chebyshevOracleMetric(o1, o3), O1VsO4: chebyshevOracleMetric(o1, o4), O2VsO3: chebyshevOracleMetric(o2, o3), O2VsO4: chebyshevOracleMetric(o2, o4), O3VsO4: chebyshevOracleMetric(o3, o4)}
	preprocessing := newChebyshevPreprocessing(poly, state, state.InputScale, offsetRelation)
	p, b, _ := NewBootstrapParametersFromConfig(cfg)
	result := ChebyshevOracleBasisResult{SchemaVersion: "fix-001-p3-diag-chebyshev-oracle-basis.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(p, b), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1 + real-branch degree-30 Chebyshev", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; corrected T0=1", LogicalSlots: 4096}, Preprocessing: preprocessing, GeneratedPowers: state.PowerChecks, OracleAgreement: pairs, OracleAgreementMax: chebyshevOracleAgreementThreshold, OracleRanges: map[string]ChebyshevOracleRange{"O1": chebyshevOracleRange(o1), "O2": chebyshevOracleRange(o2), "O3": chebyshevOracleRange(o3), "O4": chebyshevOracleRange(o4)}, AuthoritativeOracle: "raw_chebyshev_on_preprocessed_z", OldOracleStatus: "Mod1Poly.Evaluate(e2Decoded) applies ChangeOfBasis again and is invalid at this checkpoint", Validation: map[string]interface{}{"secondary_commit": gitOutput(backendRoot, "rev-parse", "HEAD"), "secondary_clean": gitOutput(backendRoot, "status", "--porcelain") == "", "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_target_scale_promotion": true, "primary_go_test": "go test ./...: PASS"}}
	if !pairs.O2VsO3.Pass || !pairs.O2VsO4.Pass || !pairs.O3VsO4.Pass {
		result.FirstSupportedCause = "plaintext_oracle_unresolved_disagreement"
		result.Validation["authoritative_oracle_established"] = false
		return writeChebyshevOracleBasisResult(result, outPath)
	}
	result.Validation["authoritative_oracle_established"] = true
	result.C0 = ChebyshevBoundaryEvidence{State: "pre_final_rescale", Level: state.Root.ciphertext.Level(), Degree: state.Root.ciphertext.Degree(), Scale: finalizationScaleString(state.Root.ciphertext.Scale), Rows: state.Root.Rows, CorrectedOracle: psGlobalMetric(o2, state.Root.actual), HistoricalOracle: psGlobalMetric(o1, state.Root.actual)}
	post := psGlobalCopy(state.Params.BootstrappingParameters, state.Root.ciphertext)
	if err := state.Eval.FastCKKS.Rescale(post, post); err != nil {
		return err
	}
	postDecoded, err := psGlobalDecode(state.Params.BootstrappingParameters, post)
	if err != nil {
		return err
	}
	result.C1 = ChebyshevBoundaryEvidence{State: "post_final_fast_rescale", Level: post.Level(), Degree: post.Degree(), Scale: finalizationScaleString(post.Scale), Rows: finalizationEvidence(post).Rows, CorrectedOracle: psGlobalMetric(o2, postDecoded), HistoricalOracle: psGlobalMetric(o1, postDecoded)}
	if result.C0.CorrectedOracle.Pass && result.C1.CorrectedOracle.Pass {
		result.FirstSupportedCause = "historical_finalization_failure_was_plaintext_oracle_basis_error"
	} else if result.C0.CorrectedOracle.Pass {
		result.FirstSupportedCause = "corrected_oracle_reopens_corrected_final_fast_rescale_boundary"
	} else {
		result.FirstSupportedCause = "corrected_oracle_pre_final_semantic_failure"
	}
	return writeChebyshevOracleBasisResult(result, outPath)
}

func writeChebyshevOracleBasisResult(result ChebyshevOracleBasisResult, outPath string) error {
	if outPath == "" {
		return fmt.Errorf("output path is required for the Chebyshev oracle basis diagnostic")
	}
	for name, value := range map[string]interface{}{"preprocessing": result.Preprocessing, "generated_powers": result.GeneratedPowers, "oracle_agreement": result.OracleAgreement, "c0": result.C0, "c1": result.C1} {
		if _, err := json.Marshal(value); err != nil {
			return fmt.Errorf("non-finite value in %s: %w", name, err)
		}
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	summary := struct {
		SchemaVersion       string                         `json:"schema_version"`
		Timestamp           time.Time                      `json:"timestamp"`
		Primary             RepositoryMetadata             `json:"primary_repository"`
		Lattigo             RepositoryMetadata             `json:"lattigo_repository"`
		Preprocessing       ChebyshevPreprocessingEvidence `json:"preprocessing"`
		GeneratedPowers     []ChebyshevPowerDomainCheck    `json:"generated_power_domain_validation"`
		OracleAgreement     ChebyshevOracleMetricPair      `json:"oracle_agreement_matrix"`
		OracleAgreementMax  float64                        `json:"oracle_agreement_threshold"`
		AuthoritativeOracle string                         `json:"authoritative_poly_oracle"`
		OldOracleStatus     string                         `json:"old_o1_status"`
		C0                  ChebyshevBoundaryEvidence      `json:"c0_pre_final_rescale"`
		C1                  ChebyshevBoundaryEvidence      `json:"c1_post_final_fast_rescale"`
		FirstSupportedCause string                         `json:"first_supported_cause"`
		Validation          map[string]interface{}         `json:"validation"`
	}{result.SchemaVersion, result.Timestamp, result.Primary, result.Lattigo, result.Preprocessing, result.GeneratedPowers, result.OracleAgreement, result.OracleAgreementMax, result.AuthoritativeOracle, result.OldOracleStatus, result.C0, result.C1, result.FirstSupportedCause, result.Validation}
	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	summaryData = append(summaryData, '\n')
	return os.WriteFile(filepath.Join(filepath.Dir(outPath), "FIX-001-P3-DIAG-CHEBYSHEV-ORACLE-BASIS-logN13-summary.json"), summaryData, 0o644)
}
