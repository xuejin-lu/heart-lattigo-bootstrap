//go:build !lattigo_standard

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

const fix001P3PublicFinalizationThreshold = 1e-2

type fix001P3PublicFinalizationMetric struct {
	MaxComponentAbs float64 `json:"max_component_abs"`
	MaxAbsComplex   float64 `json:"max_abs_complex"`
	MeanAbsComplex  float64 `json:"mean_abs_complex"`
	WorstIndex      int     `json:"worst_index"`
	WorstComponent  string  `json:"worst_component"`
	PassThreshold   bool    `json:"pass_threshold"`
	Threshold       float64 `json:"threshold"`
}

type fix001P3PublicFinalizationCheckpoint struct {
	Name       string                            `json:"name"`
	Method     string                            `json:"method"`
	Ciphertext FinalizationCiphertextEvidence    `json:"ciphertext"`
	VsStandard *fix001P3PublicFinalizationMetric `json:"vs_standard,omitempty"`
	VsInput    *fix001P3PublicFinalizationMetric `json:"vs_input,omitempty"`
}

type fix001P3PublicFinalizationScale struct {
	F1Scale             string                            `json:"f1_scale"`
	DefaultScale        string                            `json:"residual_default_scale"`
	Ratio               string                            `json:"f1_over_default_ratio"`
	InverseRatio        string                            `json:"default_over_f1_ratio"`
	RatioFloat64        float64                           `json:"ratio_float64"`
	InverseRatioFloat64 float64                           `json:"inverse_ratio_float64"`
	F2Pass              bool                              `json:"f2_pass"`
	F3Pass              bool                              `json:"f3_pass"`
	MetadataPrediction  *fix001P3PublicFinalizationMetric `json:"metadata_only_scaling_prediction,omitempty"`
}

type fix001P3PublicFinalizationResult struct {
	SchemaVersion       string                                `json:"schema_version"`
	Timestamp           time.Time                             `json:"timestamp"`
	Primary             RepositoryMetadata                    `json:"primary_repository"`
	Lattigo             RepositoryMetadata                    `json:"lattigo_repository"`
	Environment         EnvironmentMetadata                   `json:"environment"`
	Config              BootstrapConfig                       `json:"config"`
	Parameters          ExperimentParameters                  `json:"effective_parameters"`
	Workload            CorrectnessWorkload                   `json:"workload"`
	SecondaryDiffStat   string                                `json:"secondary_dirty_diff_stat"`
	SecondaryDiffSHA256 string                                `json:"secondary_dirty_diff_sha256"`
	SecondaryDirty      bool                                  `json:"secondary_dirty"`
	FixedProfile        map[string]interface{}                `json:"fixed_profile"`
	Step0               map[string]interface{}                `json:"step0_boundary"`
	F0                  *fix001P3PublicFinalizationCheckpoint `json:"f0_raw_post_s2c_pre_unpack,omitempty"`
	F1                  *fix001P3PublicFinalizationCheckpoint `json:"f1_after_unpack,omitempty"`
	F2                  *fix001P3PublicFinalizationCheckpoint `json:"f2_imform_only,omitempty"`
	RoundTrip           *FinalizationRoundTrip                `json:"imform_mform_roundtrip,omitempty"`
	F3                  *fix001P3PublicFinalizationCheckpoint `json:"f3_scale_restoration,omitempty"`
	F4                  *fix001P3PublicFinalizationCheckpoint `json:"f4_official_fast_output,omitempty"`
	UnpackBoundary      *FinalizationBoundaryComparison       `json:"unpack_boundary,omitempty"`
	Scale               *fix001P3PublicFinalizationScale      `json:"scale_evidence,omitempty"`
	F3VsF4              *fix001P3PublicFinalizationMetric     `json:"f3_vs_f4,omitempty"`
	F3F4MetadataEqual   bool                                  `json:"f3_f4_metadata_equal"`
	Classification      string                                `json:"classification"`
	FirstSupportedCause string                                `json:"first_supported_cause"`
	Validation          map[string]interface{}                `json:"validation"`
}

func fix001P3PublicFinalizationCompactMetric(metrics CorrectnessMetrics) *fix001P3PublicFinalizationMetric {
	return &fix001P3PublicFinalizationMetric{
		MaxComponentAbs: metrics.MaxComponentAbs, MaxAbsComplex: metrics.MaxAbsComplex, MeanAbsComplex: metrics.MeanAbsComplex,
		WorstIndex: firstMetricIndex(metrics.MaxComponentIndices), WorstComponent: metricWorstComponent(metrics),
		PassThreshold: metrics.PassThreshold, Threshold: metrics.Threshold,
	}
}

func fix001P3PublicFinalizationMetricFor(reference, actual []complex128) (*fix001P3PublicFinalizationMetric, error) {
	metrics, err := compareComplexVectors(reference, actual, fix001P3PublicFinalizationThreshold)
	if err != nil {
		return nil, err
	}
	return fix001P3PublicFinalizationCompactMetric(metrics), nil
}

func firstMetricIndex(indices []int) int {
	if len(indices) == 0 {
		return -1
	}
	return indices[0]
}

func metricWorstComponent(metrics CorrectnessMetrics) string {
	if metrics.MaxComponentAbs == metrics.MaxAbsReal && metrics.MaxComponentAbs >= metrics.MaxAbsImag {
		return "real"
	}
	if metrics.MaxComponentAbs == metrics.MaxAbsImag {
		return "imag"
	}
	return "complex"
}

func makeFIX001P3PublicFinalizationCheckpoint(name, method string, ct *rlwe.Ciphertext, standard, input []complex128, decode func(*rlwe.Ciphertext) ([]complex128, error)) (*fix001P3PublicFinalizationCheckpoint, []complex128, error) {
	if ct == nil {
		return nil, nil, fmt.Errorf("%s ciphertext is nil", name)
	}
	values, err := decode(ct)
	if err != nil {
		return nil, nil, fmt.Errorf("decode %s: %w", name, err)
	}
	vsStandard, err := fix001P3PublicFinalizationMetricFor(standard, values)
	if err != nil {
		return nil, nil, fmt.Errorf("compare %s vs Standard: %w", name, err)
	}
	vsInput, err := fix001P3PublicFinalizationMetricFor(input, values)
	if err != nil {
		return nil, nil, fmt.Errorf("compare %s vs input: %w", name, err)
	}
	return &fix001P3PublicFinalizationCheckpoint{Name: name, Method: method, Ciphertext: finalizationEvidence(ct), VsStandard: vsStandard, VsInput: vsInput}, values, nil
}

func fix001P3PublicFinalizationRunFastCore(eval *bootstrapping.FastEvaluator, input *rlwe.Ciphertext) (*rlwe.Ciphertext, func(*rlwe.Ciphertext) ([]rlwe.Ciphertext, error), error) {
	packed, ctxtN1, ctxtN2, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, nil, fmt.Errorf("PackAndSwitchN1ToN2: %w", err)
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return nil, nil, fmt.Errorf("ScaleDown: %w", err)
	}
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return nil, nil, fmt.Errorf("ModUp: %w", err)
	}
	ctReal, ctImag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return nil, nil, fmt.Errorf("CoeffsToSlots: %w", err)
	}
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return nil, nil, fmt.Errorf("EvalMod real: %w", err)
	}
	if ctImag != nil {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return nil, nil, fmt.Errorf("EvalMod imag: %w", err)
		}
	}
	core, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return nil, nil, fmt.Errorf("SlotsToCoeffs: %w", err)
	}
	unpack := func(source *rlwe.Ciphertext) ([]rlwe.Ciphertext, error) {
		return eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*source.CopyNew()}, ctxtN1, ctxtN2)
	}
	return core, unpack, nil
}

func fix001P3PublicFinalizationRunStandardCore(eval *bootstrapping.Evaluator, input *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	packed, _, _, err := eval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return nil, fmt.Errorf("Standard PackAndSwitchN1ToN2: %w", err)
	}
	scaled, _, err := eval.ScaleDown(&packed[0])
	if err != nil {
		return nil, fmt.Errorf("Standard ScaleDown: %w", err)
	}
	modUp, err := eval.ModUp(scaled)
	if err != nil {
		return nil, fmt.Errorf("Standard ModUp: %w", err)
	}
	ctReal, ctImag, err := eval.CoeffsToSlots(modUp)
	if err != nil {
		return nil, fmt.Errorf("Standard CoeffsToSlots: %w", err)
	}
	ctReal, err = eval.EvalMod(ctReal)
	if err != nil {
		return nil, fmt.Errorf("Standard EvalMod real: %w", err)
	}
	if ctImag != nil {
		ctImag, err = eval.EvalMod(ctImag)
		if err != nil {
			return nil, fmt.Errorf("Standard EvalMod imag: %w", err)
		}
	}
	core, err := eval.SlotsToCoeffs(ctReal, ctImag)
	if err != nil {
		return nil, fmt.Errorf("Standard SlotsToCoeffs: %w", err)
	}
	return core, nil
}

func fix001P3PublicFinalizationDiffFingerprint(root string) (string, string, bool, error) {
	status := gitOutput(root, "status", "--porcelain")
	command := exec.Command("git", "-C", root, "diff", "--no-ext-diff")
	diff, err := command.Output()
	if err != nil {
		return "", "", false, err
	}
	hash := sha256.Sum256(diff)
	return gitOutput(root, "diff", "--stat"), hex.EncodeToString(hash[:]), status != "", nil
}

func fix001P3PublicFinalizationScaleRatio(a, b string) (string, string, float64, float64, error) {
	av, ok := new(big.Float).SetPrec(256).SetString(a)
	if !ok {
		return "", "", 0, 0, fmt.Errorf("parse scale %q", a)
	}
	bv, ok := new(big.Float).SetPrec(256).SetString(b)
	if !ok {
		return "", "", 0, 0, fmt.Errorf("parse scale %q", b)
	}
	ratio := new(big.Float).SetPrec(256).Quo(av, bv)
	inverse := new(big.Float).SetPrec(256).Quo(bv, av)
	ratio64, _ := ratio.Float64()
	inverse64, _ := inverse.Float64()
	return ratio.Text('e', 80), inverse.Text('e', 80), ratio64, inverse64, nil
}

func fix001P3PublicFinalizationWrite(result fix001P3PublicFinalizationResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll("results", 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func runFIX001P3DiagPublicFinalization(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	diffStat, diffHash, secondaryDirty, err := fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return fmt.Errorf("secondary dirty diff fingerprint: %w", err)
	}
	effective := q056BuildConfig(cfg, 56)
	profile, err := q056PrepareProfile(effective)
	if err != nil {
		return fmt.Errorf("prepare q0=56 profile: %w", err)
	}
	input := reproducibleInput(profile.Residual, profile.BTP)
	inputValues := reproducibleValues(profile.Residual.MaxSlots())
	planScale := precisionSweepScale(93)
	realPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastReal, profile.Inputs.OrdinaryReal, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, nil)
	if err != nil {
		return fmt.Errorf("matched real P93 reference path: %w", err)
	}
	imagPath, err := evalModMatchedRunPathWithPlanScaleAndFastPolynomial(profile.Inputs.FastImag, profile.Inputs.OrdinaryImag, profile.Fast, profile.Standard, profile.BTP.BootstrappingParameters, profile.Inputs.FastSK, profile.StandardSK, planScale, nil)
	if err != nil {
		return fmt.Errorf("matched imag P93 reference path: %w", err)
	}
	if realPath.StandardFinal == nil || imagPath.StandardFinal == nil {
		return fmt.Errorf("matched P93 reference path did not produce Standard final branches")
	}
	standardCore, err := profile.Standard.SlotsToCoeffs(realPath.StandardFinal.CopyNew(), imagPath.StandardFinal.CopyNew())
	if err != nil {
		return fmt.Errorf("matched Standard SlotsToCoeffs reference: %w", err)
	}
	standardCoreValues, err := decodeWithSecret(profile.BTP.BootstrappingParameters, standardCore, profile.StandardSK)
	if err != nil {
		return fmt.Errorf("decode Standard core: %w", err)
	}
	standardPublic, err := profile.Standard.Bootstrap(input.CopyNew())
	if err != nil {
		return fmt.Errorf("Standard public Bootstrap: %w", err)
	}
	standardPublicValues, err := decodeWithSecret(profile.Residual, standardPublic, profile.StandardSK)
	if err != nil {
		return fmt.Errorf("decode Standard public output: %w", err)
	}
	core, unpack, err := fix001P3PublicFinalizationRunFastCore(profile.Fast, input)
	if err != nil {
		return err
	}
	coreValues, err := psGlobalDecode(profile.BTP.BootstrappingParameters, core)
	if err != nil {
		return fmt.Errorf("decode Fast F0 core: %w", err)
	}
	preFinalMetric, err := fix001P3PublicFinalizationMetricFor(standardCoreValues, coreValues)
	if err != nil {
		return err
	}
	official, err := profile.Fast.Bootstrap(input.CopyNew())
	if err != nil {
		return fmt.Errorf("official Fast public Bootstrap: %w", err)
	}
	officialValues, err := decodeWithSecret(profile.Residual, official, zeroSecret(profile.Residual))
	if err != nil {
		return fmt.Errorf("decode Fast F4 output: %w", err)
	}
	publicMetric, err := fix001P3PublicFinalizationMetricFor(standardPublicValues, officialValues)
	if err != nil {
		return err
	}
	diffStat, diffHash, secondaryDirty, err = fix001P3PublicFinalizationDiffFingerprint(backendRoot)
	if err != nil {
		return fmt.Errorf("secondary provenance after run: %w", err)
	}

	result := fix001P3PublicFinalizationResult{
		SchemaVersion: "fix-001-p3-diag-public-finalization.v1", Timestamp: time.Now().UTC(),
		Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()},
		Config:      effective, Parameters: parameterMetadata(profile.Residual, profile.BTP),
		Workload:          CorrectnessWorkload{Identifier: "reproducibleInput.v1", Formula: "real=((i%7)-3)/16; imag=((i%5)-2)/32; value=real+imag*i", LogicalSlots: profile.Residual.MaxSlots()},
		SecondaryDiffStat: diffStat, SecondaryDiffSHA256: diffHash, SecondaryDirty: secondaryDirty,
		FixedProfile: map[string]interface{}{"log_n": 13, "q0_bits": 56, "q1_bits_approx": 39, "q2_bits_approx": 40, "ps_authority": "Q012-wide", "plan_scale": "2^93", "slots": profile.Residual.MaxSlots(), "threshold": fix001P3PublicFinalizationThreshold},
		Step0:        map[string]interface{}{"reference_method": "matched evalMod P93 StandardFinal -> Standard SlotsToCoeffs for pre-finalization; genuine Standard Bootstrap for public output", "pre_finalization_vs_standard_core": preFinalMetric, "official_public_vs_standard": publicMetric, "pre_finalization_reproduced": preFinalMetric.PassThreshold, "public_failure_reproduced": !publicMetric.PassThreshold},
		Validation:   map[string]interface{}{"logn13_only": true, "no_secondary_source_modification": true, "no_parameter_tuning": true, "no_t2_repair": true, "no_ps_repair": true, "no_power_replacement_sweep": true, "no_logn16": true, "no_gate_4_or_5": true, "no_exp003": true},
	}

	f0, _, err := makeFIX001P3PublicFinalizationCheckpoint("F0", "raw post-S2C before UnpackAndSwitchN2ToN1", core, standardCoreValues, inputValues, func(ct *rlwe.Ciphertext) ([]complex128, error) {
		return psGlobalDecode(profile.BTP.BootstrappingParameters, ct)
	})
	if err != nil {
		return err
	}
	result.F0 = f0
	unpacked, err := unpack(core)
	if err != nil {
		return fmt.Errorf("F1 UnpackAndSwitchN2ToN1: %w", err)
	}
	if len(unpacked) != 1 {
		return fmt.Errorf("F1 unpack returned %d ciphertexts", len(unpacked))
	}
	f1ct := &unpacked[0]
	f1, _, err := makeFIX001P3PublicFinalizationCheckpoint("F1", "after UnpackAndSwitchN2ToN1 before public finalizer", f1ct, standardCoreValues, inputValues, func(ct *rlwe.Ciphertext) ([]complex128, error) { return psGlobalDecode(profile.Residual, ct) })
	if err != nil {
		return err
	}
	result.F1 = f1
	result.UnpackBoundary = ptrFinalizationBoundaryComparison(finalizationBoundaryComparison(core, f1ct))

	f2ct := f1ct.CopyNew()
	if err := finalizationTransformMaintained(f2ct, profile.Residual, false); err != nil {
		return fmt.Errorf("F2 IMForm-only: %w", err)
	}
	f2ct.IsMontgomery = false
	f2, f2Values, err := makeFIX001P3PublicFinalizationCheckpoint("F2", "IMForm only; preserve F1 Scale", f2ct, standardCoreValues, inputValues, func(ct *rlwe.Ciphertext) ([]complex128, error) {
		return decodeWithSecret(profile.Residual, ct, zeroSecret(profile.Residual))
	})
	if err != nil {
		return err
	}
	result.F2 = f2

	roundTripCT := f1ct.CopyNew()
	if err := finalizationTransformMaintained(roundTripCT, profile.Residual, false); err != nil {
		return err
	}
	if err := finalizationTransformMaintained(roundTripCT, profile.Residual, true); err != nil {
		return err
	}
	roundTripRows := compareFinalizationRows(f1ct, roundTripCT)
	result.RoundTrip = &FinalizationRoundTrip{Rows: roundTripRows, ExactEqual: finalizationRowsEqual(roundTripRows), Classification: "MATCH"}
	if !result.RoundTrip.ExactEqual {
		result.RoundTrip.Classification = "MONTGOMERY_ROUNDTRIP_FAILURE"
	}

	f3ct := f1ct.CopyNew()
	if err := finalizationTransformMaintained(f3ct, profile.Residual, false); err != nil {
		return fmt.Errorf("F3 IMForm: %w", err)
	}
	f3ct.IsMontgomery = false
	f3ct.Scale = profile.Residual.DefaultScale()
	f3, f3Values, err := makeFIX001P3PublicFinalizationCheckpoint("F3", "IMForm plus ResidualParameters.DefaultScale", f3ct, standardPublicValues, inputValues, func(ct *rlwe.Ciphertext) ([]complex128, error) {
		return decodeWithSecret(profile.Residual, ct, zeroSecret(profile.Residual))
	})
	if err != nil {
		return err
	}
	result.F3 = f3

	ratio, inverse, ratio64, inverse64, err := fix001P3PublicFinalizationScaleRatio(finalizationScaleString(f1ct.Scale), finalizationScaleString(profile.Residual.DefaultScale()))
	if err != nil {
		return err
	}
	result.Scale = &fix001P3PublicFinalizationScale{F1Scale: finalizationScaleString(f1ct.Scale), DefaultScale: finalizationScaleString(profile.Residual.DefaultScale()), Ratio: ratio, InverseRatio: inverse, RatioFloat64: ratio64, InverseRatioFloat64: inverse64, F2Pass: f2.VsStandard.PassThreshold, F3Pass: f3.VsStandard.PassThreshold}
	if result.Scale.F2Pass && !result.Scale.F3Pass {
		predicted := make([]complex128, len(f2Values))
		for i := range predicted {
			predicted[i] = f2Values[i] * complex(ratio64, 0)
		}
		prediction, err := fix001P3PublicFinalizationMetricFor(f3Values, predicted)
		if err != nil {
			return err
		}
		result.Scale.MetadataPrediction = prediction
	}

	f4, _, err := makeFIX001P3PublicFinalizationCheckpoint("F4", "ordinary FastEvaluator.Bootstrap public output", official, standardPublicValues, inputValues, func(ct *rlwe.Ciphertext) ([]complex128, error) {
		return decodeWithSecret(profile.Residual, ct, zeroSecret(profile.Residual))
	})
	if err != nil {
		return err
	}
	result.F4 = f4
	f3VsF4, err := fix001P3PublicFinalizationMetricFor(f3Values, officialValues)
	if err != nil {
		return err
	}
	result.F3VsF4 = f3VsF4
	result.F3F4MetadataEqual = finalizationMetadataEqual(f3ct, official)

	switch {
	case !preFinalMetric.PassThreshold || publicMetric.PassThreshold:
		result.Classification, result.FirstSupportedCause = "PRE_FINALIZATION_REPRODUCTION_CONFLICT", "pre-finalization/public boundary was not reproduced"
	case !result.UnpackBoundary.ExactEqual:
		result.Classification, result.FirstSupportedCause = "UNPACK_BOUNDARY_CHANGED_VALUE", "UnpackAndSwitchN2ToN1"
	case !result.RoundTrip.ExactEqual || !result.F2.VsStandard.PassThreshold:
		result.Classification, result.FirstSupportedCause = "MONTGOMERY_REPRESENTATION_BOUNDARY", "IMForm representation boundary"
	case result.F2.VsStandard.PassThreshold && !result.F3.VsStandard.PassThreshold && result.F3VsF4.PassThreshold && result.F3F4MetadataEqual:
		result.Classification, result.FirstSupportedCause = "FINAL_SCALE_RESTORATION", "finalizeFastPublicCiphertext scale restoration"
	default:
		result.Classification, result.FirstSupportedCause = "UNRESOLVED_PUBLIC_PATH_MISMATCH", "F3/F4 public finalization mismatch"
	}
	return fix001P3PublicFinalizationWrite(result, outPath)
}

func ptrFinalizationBoundaryComparison(value FinalizationBoundaryComparison) *FinalizationBoundaryComparison {
	return &value
}
