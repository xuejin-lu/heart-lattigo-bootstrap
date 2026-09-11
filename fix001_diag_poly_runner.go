//go:build fix001diag

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

type Fix001PowerRecord struct {
	N          int                            `json:"n"`
	A          int                            `json:"a"`
	B          int                            `json:"b"`
	C          int                            `json:"c"`
	Ciphertext FinalizationCiphertextEvidence `json:"ciphertext"`
	Projection Q01ProjectionEvidence          `json:"q01_projection"`
	Decoded    []CorrectnessValue             `json:"decoded"`
	Expected   []CorrectnessValue             `json:"expected"`
	Metrics    CorrectnessMetrics             `json:"metrics"`
	Pass       bool                           `json:"pass"`
}

type Fix001Plan struct {
	Degree    int               `json:"degree"`
	Base      int               `json:"base"`
	Level     int               `json:"level"`
	Scale     string            `json:"scale"`
	ScaleLog2 float64           `json:"scale_log2"`
	Blocks    []Fix001PlanBlock `json:"blocks"`
}

type Fix001PlanBlock struct {
	Degree    int     `json:"degree"`
	Level     int     `json:"level"`
	Scale     string  `json:"scale"`
	ScaleLog2 float64 `json:"scale_log2"`
}

type Fix001WholePolynomial struct {
	Ciphertext             FinalizationCiphertextEvidence `json:"ciphertext"`
	Decoded                []CorrectnessValue             `json:"decoded"`
	Expected               []CorrectnessValue             `json:"expected"`
	VsExactOracle          CorrectnessMetrics             `json:"vs_exact_oracle"`
	VsPriorE3Actual        CorrectnessMetrics             `json:"vs_prior_e3_actual"`
	ReproducedPriorFailure bool                           `json:"reproduced_prior_failure"`
}

type Fix001PublicQ3 struct {
	Ciphertext    FinalizationCiphertextEvidence `json:"ciphertext"`
	Decoded       []CorrectnessValue             `json:"decoded"`
	VsStandardQ01 CorrectnessMetrics             `json:"vs_standard_q01"`
	Pass          bool                           `json:"pass"`
}

type Fix001DiagPolyResult struct {
	SchemaVersion             string                          `json:"schema_version"`
	Timestamp                 time.Time                       `json:"timestamp"`
	Primary                   RepositoryMetadata              `json:"primary_repository"`
	Lattigo                   RepositoryMetadata              `json:"lattigo_repository"`
	Environment               EnvironmentMetadata             `json:"environment"`
	Config                    BootstrapConfig                 `json:"config"`
	Parameters                ExperimentParameters            `json:"effective_parameters"`
	Workload                  CorrectnessWorkload             `json:"workload"`
	Backend                   string                          `json:"backend"`
	BackendFastGoFileBlob     string                          `json:"backend_fast_go_file_blob"`
	RepairedCorrectionMarker  string                          `json:"repaired_correction_marker"`
	E2Decoded                 []CorrectnessValue              `json:"e2_decoded"`
	E2VsHistorical            CorrectnessMetrics              `json:"e2_vs_historical"`
	E2InputMatch              bool                            `json:"e2_input_match"`
	GeneratedPowers           []Fix001PowerRecord             `json:"generated_powers"`
	PowerOracleValidation     bool                            `json:"power_oracle_validation"`
	FailingPowers             []int                           `json:"failing_powers"`
	FirstFailingPowerN        int                             `json:"first_failing_power_n"`
	FirstFailingPowerDeps     []int                           `json:"first_failing_power_dependencies,omitempty"`
	FirstFailingPowerMetadata *FinalizationCiphertextEvidence `json:"first_failing_power_metadata,omitempty"`
	T2Comparison              Fix001T2Comparison              `json:"t2_comparison"`
	WholePolynomial           Fix001WholePolynomial           `json:"whole_polynomial"`
	PublicQ3                  Fix001PublicQ3                  `json:"public_q3_real"`
	Plan                      Fix001Plan                      `json:"paterson_stockmeyer_plan"`
	FirstSupportedCause       string                          `json:"first_supported_cause"`
	StandardReferencePath     string                          `json:"standard_reference_path"`
	HistoricalFastCommit      string                          `json:"historical_fast_commit"`
	RepairedFastCommit        string                          `json:"repaired_fast_commit"`
	SecondaryWorktreeClean    bool                            `json:"secondary_worktree_clean"`
	TemporarySourceRemoved    bool                            `json:"temporary_secondary_source_removed"`
	Threshold                 float64                         `json:"threshold"`
	Notes                     []string                        `json:"notes,omitempty"`
}

type Fix001T2Comparison struct {
	RepairedMetrics           CorrectnessMetrics `json:"repaired_metrics"`
	HistoricalMaxComponent    float64            `json:"historical_max_component"`
	AbsoluteImprovement       float64            `json:"absolute_improvement"`
	Pass                      bool               `json:"pass"`
	Level                     int                `json:"level"`
	Scale                     string             `json:"scale"`
	PostRescaleSemanticOracle bool               `json:"post_rescale_semantic_oracle"`
}

type Fix001PowerSummary struct {
	N         int                 `json:"n"`
	A         int                 `json:"a"`
	B         int                 `json:"b"`
	C         int                 `json:"c"`
	Ciphertext FinalizationCiphertextEvidence `json:"ciphertext"`
	Metrics   CorrectnessMetrics `json:"metrics"`
	Pass      bool               `json:"pass"`
}

type fix001HistoricalPoly struct {
	E2Decoded []CorrectnessValue `json:"e2_decoded"`
}

type fix001HistoricalEvalMod struct {
	Checkpoints []struct {
		Name    string             `json:"name"`
		Decoded []CorrectnessValue `json:"decoded"`
	} `json:"checkpoints"`
}

func fix001ScaleString(scale rlwe.Scale) string {
	return scale.Value.Text('e', rlwe.ScalePrecisionLog10)
}

func fix001ExpectedPower(n int, x []complex128, memo map[int][]complex128) []complex128 {
	if value, ok := memo[n]; ok {
		return value
	}
	if n == 1 {
		memo[n] = append([]complex128(nil), x...)
		return memo[n]
	}
	a, b := commonpolynomial.SplitDegree(n)
	va := fix001ExpectedPower(a, x, memo)
	vb := fix001ExpectedPower(b, x, memo)
	c := a - b
	if c < 0 {
		c = -c
	}
	vc := make([]complex128, len(x))
	if c != 0 {
		vc = fix001ExpectedPower(c, x, memo)
	}
	out := make([]complex128, len(x))
	for i := range out {
		out[i] = 2*va[i]*vb[i] - vc[i]
	}
	memo[n] = out
	return out
}

func fix001TargetScale(ct *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator) rlwe.Scale {
	target := eval.Mod1Parameters.ScalingFactor()
	qi := eval.Parameters.BootstrappingParameters.Q()
	for i := 0; i < eval.Mod1Parameters.DoubleAngle; i++ {
		index := ct.Level() - eval.Mod1Parameters.Mod1Poly.Depth() - eval.Mod1Parameters.DoubleAngle + i + 1
		target = target.Mul(rlwe.NewScale(qi[index]))
		target.Value.Sqrt(&target.Value)
	}
	return target
}

func fix001E2(params bootstrapping.Parameters, eval *bootstrapping.FastEvaluator, ct *rlwe.Ciphertext) (*rlwe.Ciphertext, []complex128, error) {
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

func fix001Plan(plan polynomial.DiagnosticPlan) Fix001Plan {
	result := Fix001Plan{Degree: plan.Degree, Base: plan.Base, Level: plan.Level, Scale: fix001ScaleString(plan.Scale), ScaleLog2: plan.Scale.Log2(), Blocks: make([]Fix001PlanBlock, len(plan.Blocks))}
	for i, block := range plan.Blocks {
		result.Blocks[i] = Fix001PlanBlock{Degree: block.Degree, Level: block.Level, Scale: fix001ScaleString(block.Scale), ScaleLog2: block.Scale.Log2()}
	}
	return result
}

func fix001LoadE2(path string) ([]complex128, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value fix001HistoricalPoly
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	result := make([]complex128, len(value.E2Decoded))
	for i, v := range value.E2Decoded {
		result[i] = complex(v.Real, v.Imag)
	}
	return result, nil
}

func fix001LoadPriorE3(path string) ([]complex128, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value fix001HistoricalEvalMod
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	for _, checkpoint := range value.Checkpoints {
		if checkpoint.Name == "E3_polynomial_evaluation" {
			result := make([]complex128, len(checkpoint.Decoded))
			for i, v := range checkpoint.Decoded {
				result[i] = complex(v.Real, v.Imag)
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("historical EvalMod artifact has no E3 checkpoint")
}

func RunFIX001DiagPoly(cfg BootstrapConfig, primaryRoot, backendRoot, historicalPolyPath, historicalEvalModPath, standardQ01Path string) (Fix001DiagPolyResult, error) {
	residual, btp, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	eval, err := bootstrapping.NewFastEvaluator(btp)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*reproducibleInput(residual, btp)})
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	packed[0] = *scaled
	modUp, err := eval.ModUp(&packed[0])
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	packed[0] = *modUp
	ctReal, _, err := eval.CoeffsToSlots(&packed[0])
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	e2, e2Decoded, err := fix001E2(btp, eval, ctReal)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	historicalE2, err := fix001LoadE2(historicalPolyPath)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	e2Metrics, err := compareComplexVectors(historicalE2, e2Decoded, correctnessThreshold)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	if !e2Metrics.PassThreshold {
		return Fix001DiagPolyResult{}, fmt.Errorf("DIAGNOSTIC_INPUT_MISMATCH: E2 max component=%g", e2Metrics.MaxComponentAbs)
	}

	targetScale := fix001TargetScale(e2, eval)
	powers, plan, err := eval.PolynomialEvaluator.DiagnosticGeneratePowers(e2, eval.Mod1Parameters.Mod1Poly, targetScale)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	sort.Slice(powers, func(i, j int) bool { return powers[i].N < powers[j].N })
	expectedMemo := map[int][]complex128{}
	powerRecords := make([]Fix001PowerRecord, 0, len(powers))
	failing := []int{}
	firstSupported := 0
	var firstMeta *FinalizationCiphertextEvidence
	var firstDeps []int
	for _, snapshot := range powers {
		a, b, c := 0, 0, 0
		if snapshot.N > 1 {
			a, b = commonpolynomial.SplitDegree(snapshot.N)
			c = a - b
			if c < 0 {
				c = -c
			}
		}
		projection, evidence, err := q01Projection(btp.BootstrappingParameters, snapshot.Ciphertext, true)
		if err != nil {
			return Fix001DiagPolyResult{}, err
		}
		decoded, err := diagnosticDecode(btp.BootstrappingParameters, projection, zeroSecret(btp.BootstrappingParameters))
		if err != nil {
			return Fix001DiagPolyResult{}, err
		}
		expected := fix001ExpectedPower(snapshot.N, e2Decoded, expectedMemo)
		metrics, err := compareComplexVectors(expected, decoded, correctnessThreshold)
		if err != nil {
			return Fix001DiagPolyResult{}, err
		}
		pass := metrics.PassThreshold
		if !pass {
			failing = append(failing, snapshot.N)
		}
		if !pass && firstSupported == 0 {
			depsPass := true
			for _, dep := range []int{a, b, c} {
				if dep > 1 {
					depExpected := fix001ExpectedPower(dep, e2Decoded, expectedMemo)
					depDecoded := make([]complex128, 0)
					for _, prior := range powerRecords {
						if prior.N == dep {
							for _, value := range prior.Decoded {
								depDecoded = append(depDecoded, complex(value.Real, value.Imag))
							}
							break
						}
					}
					if len(depDecoded) != len(depExpected) {
						depsPass = false
					} else {
						depMetrics, _ := compareComplexVectors(depExpected, depDecoded, correctnessThreshold)
						depsPass = depsPass && depMetrics.PassThreshold
					}
				}
			}
			if depsPass {
				firstSupported = snapshot.N
				metadata := finalizationEvidence(snapshot.Ciphertext)
				firstMeta = &metadata
				firstDeps = []int{a, b, c}
			}
		}
		powerRecords = append(powerRecords, Fix001PowerRecord{N: snapshot.N, A: a, B: b, C: c, Ciphertext: finalizationEvidence(snapshot.Ciphertext), Projection: evidence, Decoded: correctnessValues(decoded), Expected: correctnessValues(expected), Metrics: metrics, Pass: pass})
	}

	wholeCT := e2.CopyNew()
	whole, err := eval.PolynomialEvaluator.Evaluate(wholeCT, eval.Mod1Evaluator.Parameters.Mod1Poly, targetScale)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	wholeProjection, _, err := q01Projection(btp.BootstrappingParameters, whole, true)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	wholeDecoded, err := diagnosticDecode(btp.BootstrappingParameters, wholeProjection, zeroSecret(btp.BootstrappingParameters))
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	wholeExpected := make([]complex128, len(e2Decoded))
	poly := eval.Mod1Evaluator.Parameters.Mod1Poly
	for i, x := range e2Decoded {
		value := poly.Evaluate(x)
		realValue, _ := value[0].Float64()
		imagValue, _ := value[1].Float64()
		wholeExpected[i] = complex(realValue, imagValue)
	}
	wholeExact, err := compareComplexVectors(wholeExpected, wholeDecoded, correctnessThreshold)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	priorE3, err := fix001LoadPriorE3(historicalEvalModPath)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	wholePrior, err := compareComplexVectors(priorE3, wholeDecoded, correctnessThreshold)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	publicOut, err := eval.EvalMod(e2.CopyNew())
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	publicProjection, _, err := q01Projection(btp.BootstrappingParameters, publicOut, true)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	publicDecoded, err := diagnosticDecode(btp.BootstrappingParameters, publicProjection, zeroSecret(btp.BootstrappingParameters))
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	standardData, err := os.ReadFile(standardQ01Path)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	var standard Q01DiagnosticResult
	if err := json.Unmarshal(standardData, &standard); err != nil {
		return Fix001DiagPolyResult{}, err
	}
	standardStage, err := q01ReferenceStage(standard, "Q3_eval_mod_real")
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	standardBranch, err := q01ReferenceBranch(standardStage, "real")
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}
	standardQ3 := finalizationValues(standardBranch.StandardQ01Decoded)
	publicMetrics, err := compareComplexVectors(standardQ3, publicDecoded, correctnessThreshold)
	if err != nil {
		return Fix001DiagPolyResult{}, err
	}

	t2Metrics := CorrectnessMetrics{}
	for _, power := range powerRecords {
		if power.N == 2 {
			t2Metrics = power.Metrics
			break
		}
	}
	historicalT2 := 0.9995117188159196
	cause := ""
	if firstSupported != 0 {
		cause = "repaired_power_generation_t2_still_fails"
		if firstSupported != 2 {
			cause = "repaired_power_generation"
		}
	} else if !wholeExact.PassThreshold {
		cause = "paterson_stockmeyer_accumulation_or_scale_alignment"
	} else if !publicMetrics.PassThreshold {
		cause = "post_polynomial_evalmod_logic"
	} else {
		cause = "diagnostic_inconsistency"
	}
	result := Fix001DiagPolyResult{SchemaVersion: "fix-001-diag-poly.v1", Timestamp: time.Now().UTC(), Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg, Parameters: parameterMetadata(residual, btp), Workload: CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: residual.MaxSlots()}, Backend: "Fast repaired formal LogN13", BackendFastGoFileBlob: gitOutput(backendRoot, "rev-parse", "HEAD:circuits/ckks/polynomial/fast.go"), RepairedCorrectionMarker: "powers captured after repaired post-Rescale Chebyshev correction ordering in fast.go", E2Decoded: correctnessValues(e2Decoded), E2VsHistorical: e2Metrics, E2InputMatch: true, GeneratedPowers: powerRecords, PowerOracleValidation: true, FailingPowers: failing, FirstFailingPowerN: firstSupported, FirstFailingPowerDeps: firstDeps, FirstFailingPowerMetadata: firstMeta, T2Comparison: Fix001T2Comparison{RepairedMetrics: t2Metrics, HistoricalMaxComponent: historicalT2, AbsoluteImprovement: historicalT2 - t2Metrics.MaxComponentAbs, Pass: t2Metrics.PassThreshold, Level: func() int {
		for _, p := range powerRecords {
			if p.N == 2 {
				return p.Ciphertext.Level
			}
		}
		return 0
	}(), Scale: func() string {
		for _, p := range powerRecords {
			if p.N == 2 {
				return p.Ciphertext.Scale
			}
		}
		return ""
	}(), PostRescaleSemanticOracle: t2Metrics.PassThreshold}, WholePolynomial: Fix001WholePolynomial{Ciphertext: finalizationEvidence(whole), Decoded: correctnessValues(wholeDecoded), Expected: correctnessValues(wholeExpected), VsExactOracle: wholeExact, VsPriorE3Actual: wholePrior, ReproducedPriorFailure: !wholeExact.PassThreshold}, PublicQ3: Fix001PublicQ3{Ciphertext: finalizationEvidence(publicOut), Decoded: correctnessValues(publicDecoded), VsStandardQ01: publicMetrics, Pass: publicMetrics.PassThreshold}, Plan: fix001Plan(plan), FirstSupportedCause: cause, StandardReferencePath: standardQ01Path, HistoricalFastCommit: "ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2", RepairedFastCommit: "87be78ff3c591932699aba63d3be46ca306a6eea", SecondaryWorktreeClean: false, TemporarySourceRemoved: false, Threshold: correctnessThreshold}
	_ = ctxtN1
	_ = ctxtN2
	return result, nil
}

func WriteFIX001DiagPolyResult(result Fix001DiagPolyResult, path string) error {
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

func WriteFIX001DiagPolySummary(result Fix001DiagPolyResult, path string) error {
	powers := make([]Fix001PowerSummary, len(result.GeneratedPowers))
	for i, power := range result.GeneratedPowers {
		powers[i] = Fix001PowerSummary{N: power.N, A: power.A, B: power.B, C: power.C, Ciphertext: power.Ciphertext, Metrics: power.Metrics, Pass: power.Pass}
	}
	summary := struct {
		SchemaVersion           string                    `json:"schema_version"`
		Timestamp               time.Time                 `json:"timestamp"`
		Backend                 string                    `json:"backend"`
		Primary                 RepositoryMetadata        `json:"primary_repository"`
		Lattigo                 RepositoryMetadata        `json:"lattigo_repository"`
		RepairedFastCommit      string                    `json:"repaired_fast_commit"`
		HistoricalFastCommit    string                    `json:"historical_fast_commit"`
		E2InputMatch            bool                      `json:"e2_input_match"`
		E2VsHistorical          CorrectnessMetrics        `json:"e2_vs_historical"`
		GeneratedPowers         []Fix001PowerSummary      `json:"generated_powers"`
		FailingPowers           []int                     `json:"failing_powers"`
		T2Comparison            Fix001T2Comparison        `json:"t2_comparison"`
		WholePolynomial         Fix001WholePolynomial    `json:"whole_polynomial"`
		PublicQ3                Fix001PublicQ3           `json:"public_q3_real"`
		Plan                    Fix001Plan               `json:"paterson_stockmeyer_plan"`
		FirstSupportedCause     string                    `json:"first_supported_cause"`
		Threshold               float64                   `json:"threshold"`
		PowerOracleValidation   bool                      `json:"power_oracle_validation"`
		SecondaryWorktreeClean  bool                      `json:"secondary_worktree_clean"`
		TemporarySourceRemoved  bool                      `json:"temporary_secondary_source_removed"`
	}{SchemaVersion: result.SchemaVersion, Timestamp: result.Timestamp, Backend: result.Backend, Primary: result.Primary, Lattigo: result.Lattigo, RepairedFastCommit: result.RepairedFastCommit, HistoricalFastCommit: result.HistoricalFastCommit, E2InputMatch: result.E2InputMatch, E2VsHistorical: result.E2VsHistorical, GeneratedPowers: powers, FailingPowers: result.FailingPowers, T2Comparison: result.T2Comparison, WholePolynomial: result.WholePolynomial, PublicQ3: result.PublicQ3, Plan: result.Plan, FirstSupportedCause: result.FirstSupportedCause, Threshold: result.Threshold, PowerOracleValidation: result.PowerOracleValidation, SecondaryWorktreeClean: result.SecondaryWorktreeClean, TemporarySourceRemoved: result.TemporarySourceRemoved}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil { return err }
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
	return os.WriteFile(path, data, 0o644)
}

func runFIX001DiagPoly(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	result, err := RunFIX001DiagPoly(cfg, primaryRoot, backendRoot, "results/EXP-002C-DIAG-POLY-logN13-fast.json", "results/EXP-002C-DIAG-EVALMOD-logN13-fast.json", "results/FIX-001-DIAG-Q01-logN13-standard.json")
	if err != nil {
		return err
	}
	if err := WriteFIX001DiagPolyResult(result, outPath); err != nil { return err }
	return WriteFIX001DiagPolySummary(result, filepath.Join(filepath.Dir(outPath), "FIX-001-DIAG-POLY-logN13-summary.json"))
}
