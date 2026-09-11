//go:build fix001t3diag

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type FIX001T3Checkpoint struct {
	Name         string                         `json:"name"`
	Operation    string                         `json:"operation"`
	SemanticSafe bool                           `json:"semantic_safe"`
	Ciphertext   FinalizationCiphertextEvidence `json:"ciphertext"`
	Projection   Q01ProjectionEvidence          `json:"q01_projection"`
	Decoded      []CorrectnessValue             `json:"decoded"`
	Expected     []CorrectnessValue             `json:"expected,omitempty"`
	Metrics      *CorrectnessMetrics            `json:"metrics,omitempty"`
}

type FIX001T3AlignmentTrace struct {
	LeftOutput            FinalizationCiphertextEvidence  `json:"left_output"`
	CorrectionT1          FinalizationCiphertextEvidence  `json:"correction_t1"`
	AlignmentProduct      FinalizationCiphertextEvidence  `json:"alignment_product,omitempty"`
	AlignedCorrection     *FinalizationCiphertextEvidence `json:"aligned_correction,omitempty"`
	ManualSubtraction     FinalizationCiphertextEvidence  `json:"manual_subtraction"`
	ProductionSubtraction FinalizationCiphertextEvidence  `json:"production_subtraction"`
	RatioDirection        int                             `json:"ratio_direction"`
	RatioExact            string                          `json:"ratio_exact"`
	RatioInteger          string                          `json:"ratio_integer"`
	RatioAbsDifference    string                          `json:"ratio_abs_difference"`
	RatioRelativeError    string                          `json:"ratio_relative_error"`
	RatioIntegral         bool                            `json:"ratio_integral"`
	AlignedCorrectionVsT1 CorrectnessMetrics              `json:"aligned_correction_vs_t1"`
	ManualVsExpected      CorrectnessMetrics              `json:"manual_vs_expected"`
	ProductionVsExpected  CorrectnessMetrics              `json:"production_vs_expected"`
	ManualVsProduction    CorrectnessMetrics              `json:"manual_vs_production"`
	RowsBitExact          bool                            `json:"rows_bit_exact"`
}

type FIX001T3Bypass struct {
	Present    bool                           `json:"present"`
	Correction FinalizationCiphertextEvidence `json:"correction"`
	Output     FinalizationCiphertextEvidence `json:"output"`
	Decoded    []CorrectnessValue             `json:"decoded"`
	Expected   []CorrectnessValue             `json:"expected"`
	Metrics    CorrectnessMetrics             `json:"metrics"`
	Pass       bool                           `json:"pass"`
}

type FIX001T3Capacity struct {
	Q0Q1CenteredHalf              string  `json:"q0_q1_centered_half"`
	AlignmentIntegerRatio         string  `json:"alignment_integer_ratio"`
	T1DecodedMaxMagnitude         float64 `json:"t1_decoded_max_magnitude"`
	NominalScaledEncodingEstimate float64 `json:"nominal_scaled_encoding_estimate"`
	CapacityOverflowClaim         bool    `json:"capacity_overflow_claim"`
	Note                          string  `json:"note"`
}

type FIX001T3Result struct {
	SchemaVersion          string                 `json:"schema_version"`
	Timestamp              time.Time              `json:"timestamp"`
	Primary                RepositoryMetadata     `json:"primary_repository"`
	Lattigo                RepositoryMetadata     `json:"lattigo_repository"`
	Environment            EnvironmentMetadata    `json:"environment"`
	Config                 BootstrapConfig        `json:"config"`
	Parameters             ExperimentParameters   `json:"effective_parameters"`
	Workload               CorrectnessWorkload    `json:"workload"`
	RepairedFastCommit     string                 `json:"repaired_fast_commit"`
	CorrectedT0Oracle      bool                   `json:"corrected_t0_oracle"`
	E2Decoded              []CorrectnessValue     `json:"e2_decoded"`
	E2VsHistorical         CorrectnessMetrics     `json:"e2_vs_historical"`
	T3Split                []int                  `json:"t3_split"`
	ProductionMode         string                 `json:"production_mode"`
	Checkpoints            []FIX001T3Checkpoint   `json:"checkpoints"`
	Alignment              FIX001T3AlignmentTrace `json:"subaligned_trace"`
	Bypass                 *FIX001T3Bypass        `json:"bypass_control,omitempty"`
	Capacity               FIX001T3Capacity       `json:"capacity_sanity"`
	FirstSupportedCause    string                 `json:"first_supported_cause"`
	Threshold              float64                `json:"threshold"`
	SecondaryWorktreeClean bool                   `json:"secondary_worktree_clean"`
	TemporarySourceRemoved bool                   `json:"temporary_secondary_source_removed"`
	Notes                  []string               `json:"notes"`
}

func fix001T3Checkpoint(name, operation string, params ckks.Parameters, ct *rlwe.Ciphertext, expected []complex128, semanticSafe bool) (FIX001T3Checkpoint, []complex128, error) {
	projection, evidence, err := q01Projection(params, ct, true)
	if err != nil {
		return FIX001T3Checkpoint{}, nil, err
	}
	decoded, err := diagnosticDecode(params, projection, zeroSecret(params))
	if err != nil {
		return FIX001T3Checkpoint{}, nil, err
	}
	checkpoint := FIX001T3Checkpoint{Name: name, Operation: operation, SemanticSafe: semanticSafe, Ciphertext: finalizationEvidence(ct), Projection: evidence, Decoded: correctnessValues(decoded)}
	if expected != nil {
		metrics, err := compareComplexVectors(expected, decoded, correctnessThreshold)
		if err != nil {
			return FIX001T3Checkpoint{}, nil, err
		}
		checkpoint.Expected = correctnessValues(expected)
		checkpoint.Metrics = &metrics
	}
	return checkpoint, decoded, nil
}

func fix001T3VectorMul(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] * b[i]
	}
	return out
}
func fix001T3VectorScale(a []complex128, scalar complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = scalar * a[i]
	}
	return out
}
func fix001T3VectorSub(a, b []complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range out {
		out[i] = a[i] - b[i]
	}
	return out
}

func fix001T3EqualScaleCorrection(params ckks.Parameters, reference *rlwe.Ciphertext, values []complex128) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, reference.Level())
	*pt.MetaData = *reference.MetaData
	pt.Scale = reference.Scale
	pt.IsNTT = true
	pt.IsMontgomery = false
	pt.LogDimensions = ring.Dimensions{Cols: reference.LogDimensions.Cols, Rows: reference.LogDimensions.Rows}
	if err := ckks.NewEncoder(params).Encode(values, pt); err != nil {
		return nil, err
	}
	ct := fastckks.NewCiphertext(params, 1, reference.Level())
	*ct.MetaData = *reference.MetaData
	ct.IsNTT = true
	ct.IsMontgomery = true
	for limb := 0; limb < 2; limb++ {
		copy(ct.Value[0].Coeffs[limb], pt.Value.Coeffs[limb])
		params.RingQ().SubRings[limb].MForm(ct.Value[0].Coeffs[limb], ct.Value[0].Coeffs[limb])
		ct.Value[1].Coeffs[limb] = make([]uint64, params.N())
	}
	return ct, nil
}

func fix001T3E2(params bootstrapping.Parameters, eval *bootstrapping.FastEvaluator, ct *rlwe.Ciphertext) (*rlwe.Ciphertext, []complex128, error) {
	res := ct.CopyNew()
	mod1Params := eval.Mod1Parameters
	res.Scale = mod1Params.ScalingFactor()
	offset := new(big.Float).Sub(&mod1Params.Mod1Poly.B, &mod1Params.Mod1Poly.A)
	offset.Mul(offset, new(big.Float).SetFloat64(mod1Params.IntervalShrinkFactor()))
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

func fix001T3LoadE2(path string) ([]complex128, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value struct {
		E2Decoded []CorrectnessValue `json:"e2_decoded"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	result := make([]complex128, len(value.E2Decoded))
	for i, item := range value.E2Decoded {
		result[i] = complex(item.Real, item.Imag)
	}
	return result, nil
}

func RunFIX001T3(cfg BootstrapConfig, primaryRoot, backendRoot string) (FIX001T3Result, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return FIX001T3Result{}, err
	}
	eval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return FIX001T3Result{}, err
	}
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*reproducibleInput(residual, btp)})
	if err != nil {
		return FIX001T3Result{}, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return FIX001T3Result{}, err
	}
	packed[0] = *scaled
	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return FIX001T3Result{}, err
	}
	packed[0] = *modUp
	ctReal, _, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return FIX001T3Result{}, err
	}
	e2, e2Decoded, err := fix001T3E2(btp, eval, ctReal)
	if err != nil {
		return FIX001T3Result{}, err
	}
	historicalE2, err := fix001T3LoadE2("results/EXP-002C-DIAG-POLY-logN13-fast.json")
	if err != nil {
		return FIX001T3Result{}, err
	}
	e2Metrics, err := compareComplexVectors(historicalE2, e2Decoded, correctnessThreshold)
	if err != nil {
		return FIX001T3Result{}, err
	}
	if !e2Metrics.PassThreshold {
		return FIX001T3Result{}, fmt.Errorf("DIAGNOSTIC_INPUT_MISMATCH: E2 max component=%g", e2Metrics.MaxComponentAbs)
	}
	trace, err := eval.PolynomialEvaluator.DiagnosticReplayT3(e2)
	if err != nil {
		return FIX001T3Result{}, err
	}
	zero := zeroSecret(btp.BootstrappingParameters)
	memo := make(map[int][]complex128)
	x := fix001ExpectedPower(1, e2Decoded, memo)
	y := fix001ExpectedPower(2, e2Decoded, memo)
	productExpected := fix001T3VectorMul(x, y)
	doubledExpected := fix001T3VectorScale(productExpected, 2)
	correctedExpected := fix001T3VectorSub(doubledExpected, x)
	checkpoints := make([]FIX001T3Checkpoint, 0, 7)
	for _, item := range []struct {
		name, operation string
		ct              *rlwe.Ciphertext
		expected        []complex128
		safe            bool
	}{
		{"T3-0-T1", "production power basis input", trace.T1, x, true},
		{"T3-0-T2", "production power basis repaired T2", trace.T2, y, true},
		{"T3-1-multiply", "strict MulRelin(T1,T2)", trace.Multiply, nil, false},
		{"T3-2-double", "Add(out,out,out)", trace.Doubled, nil, false},
		{"T3-3-post-rescale", "Rescale before c>0 correction", trace.PostRescale, doubledExpected, true},
		{"T3-4-production", "actual production subAligned result", trace.ProductionT3, correctedExpected, true},
		{"T3-4-independent-subtraction", "independent production-equivalent subAligned replay", trace.ManualSubtraction, correctedExpected, true},
	} {
		checkpoint, _, err := fix001T3Checkpoint(item.name, item.operation, btp.BootstrappingParameters, item.ct, item.expected, item.safe)
		if err != nil {
			return FIX001T3Result{}, err
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	postMetrics := *checkpoints[4].Metrics
	productionMetrics := *checkpoints[5].Metrics
	manualMetrics := *checkpoints[6].Metrics
	alignedProjection, alignedEvidence, err := q01Projection(btp.BootstrappingParameters, trace.AlignedCorrection, true)
	if err != nil {
		return FIX001T3Result{}, err
	}
	alignedDecoded, err := diagnosticDecode(btp.BootstrappingParameters, alignedProjection, zero)
	if err != nil {
		return FIX001T3Result{}, err
	}
	alignedMetrics, err := compareComplexVectors(x, alignedDecoded, correctnessThreshold)
	if err != nil {
		return FIX001T3Result{}, err
	}
	_ = alignedEvidence
	leftEvidence := finalizationEvidence(trace.PostRescale)
	correctionEvidence := finalizationEvidence(trace.T1)
	alignmentEvidence := finalizationEvidence(trace.AlignmentProduct)
	alignedFinal := finalizationEvidence(trace.AlignedCorrection)
	manualEvidence := finalizationEvidence(trace.ManualSubtraction)
	productionEvidence := finalizationEvidence(trace.ProductionSubtraction)
	manualProduction, err := compareComplexVectors(finalizationValues(checkpoints[6].Decoded), finalizationValues(checkpoints[5].Decoded), correctnessThreshold)
	if err != nil {
		return FIX001T3Result{}, err
	}
	rowsExact := finalizationRowsEqual(compareFinalizationRows(trace.ManualSubtraction, trace.ProductionSubtraction))
	traceResult := FIX001T3AlignmentTrace{LeftOutput: leftEvidence, CorrectionT1: correctionEvidence, AlignmentProduct: alignmentEvidence, ManualSubtraction: manualEvidence, ProductionSubtraction: productionEvidence, RatioDirection: trace.RatioDirection, RatioExact: trace.RatioExact, RatioInteger: trace.RatioInteger, RatioAbsDifference: trace.RatioAbsDifference, RatioRelativeError: trace.RatioRelativeError, RatioIntegral: trace.RatioIntegral, AlignedCorrectionVsT1: alignedMetrics, ManualVsExpected: manualMetrics, ProductionVsExpected: productionMetrics, ManualVsProduction: manualProduction, RowsBitExact: rowsExact}
	if trace.AlignedCorrection != nil {
		traceResult.AlignedCorrection = &alignedFinal
	}

	var bypass *FIX001T3Bypass
	if postMetrics.PassThreshold {
		correction, err := fix001T3EqualScaleCorrection(btp.BootstrappingParameters, trace.PostRescale, x)
		if err != nil {
			return FIX001T3Result{}, err
		}
		bypassOut := trace.PostRescale.CopyNew()
		if err := eval.FastCKKS.Sub(bypassOut, correction, bypassOut); err != nil {
			return FIX001T3Result{}, err
		}
		projection, _, err := q01Projection(btp.BootstrappingParameters, bypassOut, true)
		if err != nil {
			return FIX001T3Result{}, err
		}
		decoded, err := diagnosticDecode(btp.BootstrappingParameters, projection, zero)
		if err != nil {
			return FIX001T3Result{}, err
		}
		metrics, err := compareComplexVectors(correctedExpected, decoded, correctnessThreshold)
		if err != nil {
			return FIX001T3Result{}, err
		}
		bypass = &FIX001T3Bypass{Present: true, Correction: finalizationEvidence(correction), Output: finalizationEvidence(bypassOut), Decoded: correctnessValues(decoded), Expected: correctnessValues(correctedExpected), Metrics: metrics, Pass: metrics.PassThreshold}
	}
	q01q1 := new(big.Int).Mul(new(big.Int).SetUint64(btp.BootstrappingParameters.RingQ().SubRings[0].Modulus), new(big.Int).SetUint64(btp.BootstrappingParameters.RingQ().SubRings[1].Modulus))
	q01q1.Rsh(q01q1, 1)
	maxT1 := 0.0
	for _, value := range x {
		maxT1 = math.Max(maxT1, math.Hypot(real(value), imag(value)))
	}
	ratioFloat, _, _ := big.NewFloat(0).SetPrec(256).Parse(trace.RatioInteger, 10)
	nominal := maxT1
	if ratioFloat != nil {
		ratioValue, _ := ratioFloat.Float64()
		nominal *= ratioValue
	}
	cause := "subaligned_unresolved"
	if !postMetrics.PassThreshold {
		cause = "t3_multiply_double_or_rescale"
	} else if !alignedMetrics.PassThreshold {
		cause = "subaligned_scale_promotion"
	} else if !productionMetrics.PassThreshold {
		cause = "subaligned_subtraction"
	} else {
		cause = "diagnostic_inconsistency"
	}
	return FIX001T3Result{SchemaVersion: "fix-001-diag-t3.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, RepairedFastCommit: "87be78ff3c591932699aba63d3be46ca306a6eea", CorrectedT0Oracle: true, E2Decoded: correctnessValues(e2Decoded), E2VsHistorical: e2Metrics, T3Split: []int{trace.A, trace.B, trace.C}, ProductionMode: trace.Mode, Checkpoints: checkpoints, Alignment: traceResult, Bypass: bypass, Capacity: FIX001T3Capacity{Q0Q1CenteredHalf: q01q1.String(), AlignmentIntegerRatio: trace.RatioInteger, T1DecodedMaxMagnitude: maxT1, NominalScaledEncodingEstimate: nominal, CapacityOverflowClaim: false, Note: "Estimate is diagnostic only; no capacity overflow is claimed from this evidence."}, FirstSupportedCause: cause, Threshold: correctnessThreshold, SecondaryWorktreeClean: false, TemporarySourceRemoved: false, Notes: []string{"Corrected T0=1 oracle is used.", "This is diagnostic only; no production Lattigo source was changed.", "No PS repair or interpretation is made in this task."}}, nil
}

func WriteFIX001T3Result(result FIX001T3Result, path string) error {
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

func WriteFIX001T3Summary(result FIX001T3Result, path string) error {
	type checkpoint struct {
		Name         string                         `json:"name"`
		Operation    string                         `json:"operation"`
		SemanticSafe bool                           `json:"semantic_safe"`
		Ciphertext   FinalizationCiphertextEvidence `json:"ciphertext"`
		Metrics      *CorrectnessMetrics            `json:"metrics,omitempty"`
	}
	checks := make([]checkpoint, len(result.Checkpoints))
	for i, item := range result.Checkpoints {
		checks[i] = checkpoint{Name: item.Name, Operation: item.Operation, SemanticSafe: item.SemanticSafe, Ciphertext: item.Ciphertext, Metrics: item.Metrics}
	}
	summary := struct {
		SchemaVersion          string                 `json:"schema_version"`
		Timestamp              time.Time              `json:"timestamp"`
		RepairedFastCommit     string                 `json:"repaired_fast_commit"`
		CorrectedT0Oracle      bool                   `json:"corrected_t0_oracle"`
		E2VsHistorical         CorrectnessMetrics     `json:"e2_vs_historical"`
		T3Split                []int                  `json:"t3_split"`
		ProductionMode         string                 `json:"production_mode"`
		Checkpoints            []checkpoint           `json:"checkpoints"`
		Alignment              FIX001T3AlignmentTrace `json:"subaligned_trace"`
		Bypass                 *FIX001T3Bypass        `json:"bypass_control,omitempty"`
		Capacity               FIX001T3Capacity       `json:"capacity_sanity"`
		FirstSupportedCause    string                 `json:"first_supported_cause"`
		Threshold              float64                `json:"threshold"`
		SecondaryWorktreeClean bool                   `json:"secondary_worktree_clean"`
		TemporarySourceRemoved bool                   `json:"temporary_secondary_source_removed"`
		Notes                  []string               `json:"notes"`
	}{result.SchemaVersion, result.Timestamp, result.RepairedFastCommit, result.CorrectedT0Oracle, result.E2VsHistorical, result.T3Split, result.ProductionMode, checks, result.Alignment, result.Bypass, result.Capacity, result.FirstSupportedCause, result.Threshold, result.SecondaryWorktreeClean, result.TemporarySourceRemoved, result.Notes}
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

func runFIX001T3(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result, err := RunFIX001T3(cfg, primaryRoot, backendRoot)
	if err != nil {
		return err
	}
	if err := WriteFIX001T3Result(result, outPath); err != nil {
		return err
	}
	return WriteFIX001T3Summary(result, filepath.Join(filepath.Dir(outPath), "FIX-001-DIAG-T3-logN13-summary.json"))
}
