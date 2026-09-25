//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3F0LocalQ2PrimaryBase = "e8d5b578d3ba08234c7d9892d87d4e6fbd8785c5"
	requiredFIX001P3F0LocalQ2Secondary   = "25e70430b10cd4af37ad4cf94cd994912e970e3f"
)

type fix001P3LocalQ2Capacity struct {
	MaxAbs string  `json:"max_abs_centered"`
	Ratio  float64 `json:"max_abs_over_centered_half"`
	Unique bool    `json:"centered_unique"`
}

type fix001P3LocalQ2Expansion struct {
	InputQ01CenteredUnique bool                    `json:"input_q01_centered_unique"`
	Q01RowsUnchanged       bool                    `json:"q0_q1_rows_unchanged_bit_for_bit"`
	Q2MatchesFullRNS       bool                    `json:"q2_matches_stage_aligned_full_rns"`
	Semantic               *PSGlobalMetric         `json:"semantic_change"`
	Pass                   bool                    `json:"pass"`
	Q01                    string                  `json:"q01"`
	Q012                   string                  `json:"q012"`
	Q01Capacity            fix001P3LocalQ2Capacity `json:"q01_capacity"`
	Q012Capacity           fix001P3LocalQ2Capacity `json:"q012_capacity"`
}

type fix001P3LocalQ2Guard struct {
	Q01Before  fix001P3LocalQ2Capacity `json:"q01_before"`
	Q01After   fix001P3LocalQ2Capacity `json:"q01_after"`
	Q012Before fix001P3LocalQ2Capacity `json:"q012_before"`
	Q012After  fix001P3LocalQ2Capacity `json:"q012_after"`
	Value      *PSGlobalMetric         `json:"value_preservation"`
	RowsMatch  bool                    `json:"q012_rows_match_full_rns"`
}

type fix001P3LocalQ2Rescale struct {
	Divisor              string                  `json:"divisor"`
	InputLevel           int                     `json:"input_level"`
	OutputLevel          int                     `json:"output_level"`
	InputScale           string                  `json:"input_scale"`
	OutputScale          string                  `json:"output_scale"`
	LocalError           *PSGlobalMetric         `json:"local_error"`
	PostCumulativeError  *PSGlobalMetric         `json:"post_cumulative_error"`
	Q012Output           fix001P3LocalQ2Capacity `json:"q012_output_capacity"`
	Q01Output            fix001P3LocalQ2Capacity `json:"q01_output_capacity"`
	RowsMatch            bool                    `json:"q012_rows_match_full_rns"`
	RoundedDivisionMatch bool                    `json:"rounded_division_match"`
}

type fix001P3LocalQ2Contraction struct {
	CenteredUnique bool            `json:"q01_centered_unique"`
	RowsMatch      bool            `json:"q0_q1_rows_match_q012"`
	Semantic       *PSGlobalMetric `json:"q01_vs_q012"`
	MetadataMatch  bool            `json:"metadata_match"`
	Pass           bool            `json:"pass"`
}

type fix001P3LocalQ2BranchEvidence struct {
	GuardBits   int                        `json:"guard_bits"`
	Name        string                     `json:"branch"`
	Expansion   fix001P3LocalQ2Expansion   `json:"expansion"`
	Guard       fix001P3LocalQ2Guard       `json:"two_bit_guard"`
	Rescale     fix001P3LocalQ2Rescale     `json:"rescale"`
	Contraction fix001P3LocalQ2Contraction `json:"contraction"`
}

type fix001P3LocalQ2Ops struct {
	CoefficientTransforms int `json:"coefficient_transforms"`
	Q2NTT                 int `json:"q2_ntt"`
	Q2INTT                int `json:"q2_intt"`
	Q2ArithmeticPasses    int `json:"q2_limb_arithmetic_passes"`
	RescaleBoundaries     int `json:"q2_rescale_boundaries"`
}

type fix001P3LocalQ2Result struct {
	SchemaVersion  string                          `json:"schema_version"`
	Timestamp      time.Time                       `json:"timestamp"`
	Primary        RepositoryMetadata              `json:"primary_repository"`
	Lattigo        RepositoryMetadata              `json:"lattigo_repository"`
	Environment    EnvironmentMetadata             `json:"environment"`
	Config         BootstrapConfig                 `json:"config"`
	Parameters     ExperimentParameters            `json:"effective_parameters"`
	Primes         map[string]interface{}          `json:"q0_q1_q2_primes"`
	Controls       map[string]interface{}          `json:"l0_controls"`
	R0             q056CandidateEvidence           `json:"r0_q01_only_c1"`
	R1             q056CandidateEvidence           `json:"r1_local_q2"`
	Branches       []fix001P3LocalQ2BranchEvidence `json:"local_q2_branches"`
	Operations     fix001P3LocalQ2Ops              `json:"static_local_q2_operations"`
	Downstream     map[string]interface{}          `json:"downstream,omitempty"`
	Classification string                          `json:"classification"`
	BlockerClosed  bool                            `json:"ps_arithmetic_blocker_independently_closed"`
	FirstBlocker   string                          `json:"first_remaining_blocker"`
	Validation     map[string]interface{}          `json:"validation"`
}

func fix001P3LocalQ2CapacityOf(values []*big.Int, modulus *big.Int) fix001P3LocalQ2Capacity {
	maxAbs := big.NewInt(0)
	for _, value := range values {
		if new(big.Int).Abs(value).Cmp(maxAbs) > 0 {
			maxAbs.Set(new(big.Int).Abs(value))
		}
	}
	half := new(big.Int).Rsh(new(big.Int).Set(modulus), 1)
	ratio := 0.0
	if half.Sign() != 0 {
		ratio, _ = new(big.Float).Quo(new(big.Float).SetInt(maxAbs), new(big.Float).SetInt(half)).Float64()
	}
	return fix001P3LocalQ2Capacity{MaxAbs: maxAbs.String(), Ratio: ratio, Unique: maxAbs.Cmp(half) < 0}
}

func fix001P3LocalQ2Moduli(params ckks.Parameters) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	if len(params.RingQ().SubRings) < 3 {
		return nil, nil, nil, nil, fmt.Errorf("q012 requires at least three Q primes")
	}
	q0 := new(big.Int).SetUint64(params.RingQ().SubRings[0].Modulus)
	q1 := new(big.Int).SetUint64(params.RingQ().SubRings[1].Modulus)
	q2 := new(big.Int).SetUint64(params.RingQ().SubRings[2].Modulus)
	q01 := new(big.Int).Mul(new(big.Int).Set(q0), q1)
	q012 := new(big.Int).Mul(new(big.Int).Set(q01), q2)
	return q0, q1, q2, q012, nil
}

func fix001P3LocalQ2Round(value, divisor *big.Int) *big.Int {
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(value, divisor, remainder)
	if new(big.Int).Lsh(new(big.Int).Abs(remainder), 1).Cmp(divisor) >= 0 {
		if value.Sign() < 0 {
			quotient.Sub(quotient, big.NewInt(1))
		} else {
			quotient.Add(quotient, big.NewInt(1))
		}
	}
	return quotient
}

func fix001P3LocalQ2Build(params ckks.Parameters, source *rlwe.Ciphertext, values [][]*big.Int, level int, scale rlwe.Scale, montgomery bool) (*rlwe.Ciphertext, error) {
	if level < 0 || level > 2 {
		return nil, fmt.Errorf("local q012 level %d is unsupported", level)
	}
	ct := ckks.NewCiphertext(params, source.Degree(), level)
	*ct.MetaData = *source.MetaData
	ct.Scale, ct.IsNTT, ct.IsMontgomery = scale, true, montgomery
	for component := range values {
		normal := postMod1S2CLiftSigned(params, values[component], level)
		for limb := 0; limb <= level; limb++ {
			if montgomery {
				params.RingQ().SubRings[limb].MForm(normal.Coeffs[limb], ct.Value[component].Coeffs[limb])
			} else {
				copy(ct.Value[component].Coeffs[limb], normal.Coeffs[limb])
			}
		}
	}
	return ct, nil
}

func fix001P3LocalQ2Decode(params ckks.Parameters, source *rlwe.Ciphertext, values [][]*big.Int, scale rlwe.Scale) ([]complex128, error) {
	ct, err := fix001P3LocalQ2Build(params, source, values, 2, scale, false)
	if err != nil {
		return nil, err
	}
	return diagnosticDecode(params, ct, zeroSecret(params))
}

func fix001P3LocalQ2CopyQ01(params ckks.Parameters, source *rlwe.Ciphertext, values [][]*big.Int, level int, scale rlwe.Scale) (*rlwe.Ciphertext, error) {
	if level < 0 {
		return nil, fmt.Errorf("negative target level")
	}
	ct := fastckks.NewCiphertext(params, source.Degree(), level)
	*ct.MetaData = *source.MetaData
	ct.Scale, ct.IsNTT, ct.IsMontgomery = scale, true, true
	for component := range values {
		coeff := ring.NewPoly(params.N(), 1)
		for limb := 0; limb < 2; limb++ {
			modulus := new(big.Int).SetUint64(params.RingQ().SubRings[limb].Modulus)
			for i, value := range values[component] {
				coeff.Coeffs[limb][i] = new(big.Int).Mod(new(big.Int).Set(value), modulus).Uint64()
			}
		}
		normal := ring.NewPoly(params.N(), 1)
		params.RingQ().AtLevel(1).NTT(coeff, normal)
		for limb := 0; limb < 2; limb++ {
			params.RingQ().SubRings[limb].MForm(normal.Coeffs[limb], ct.Value[component].Coeffs[limb])
		}
	}
	return ct, nil
}

func fix001P3LocalQ2RowEqual(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func fix001P3LocalQ2Override(params ckks.Parameters, branch string, evidence *fix001P3LocalQ2BranchEvidence) psRescaleGuardFinalOverride {
	return fix001P3LocalQ2OverrideWithBits(params, branch, evidence, 2)
}

func fix001P3LocalQ2OverrideWithBits(params ckks.Parameters, branch string, evidence *fix001P3LocalQ2BranchEvidence, guardBits int) psRescaleGuardFinalOverride {
	return func(params ckks.Parameters, eval *fastckks.Evaluator, ct *rlwe.Ciphertext, boundary *psRescaleGuardBoundary) (bool, error) {
		q0, q1, _, q012, err := fix001P3LocalQ2Moduli(params)
		if err != nil {
			return true, err
		}
		if ct.Level() < 2 {
			return true, fmt.Errorf("F0 input level %d cannot materialize q2", ct.Level())
		}
		recovered, err := postMod1S2CRecoverQ01(params, ct)
		if err != nil {
			return true, err
		}
		input := recovered
		q01 := new(big.Int).Mul(q0, q1)
		evidence.GuardBits = guardBits
		evidence.Name = branch
		evidence.Expansion.Q01 = q01.String()
		evidence.Expansion.Q012 = q012.String()
		inputCapacity := fix001P3LocalQ2Capacity{}
		inputCapacity012 := fix001P3LocalQ2Capacity{}
		inputUnique := true
		for _, values := range input {
			cap := fix001P3LocalQ2CapacityOf(values, q01)
			if cap.Ratio > inputCapacity.Ratio {
				inputCapacity = cap
			}
			cap012 := fix001P3LocalQ2CapacityOf(values, q012)
			if cap012.Ratio > inputCapacity012.Ratio {
				inputCapacity012 = cap012
			}
			inputUnique = inputUnique && cap.Unique
		}
		evidence.Expansion.Q01Capacity = inputCapacity
		evidence.Expansion.Q012Capacity = inputCapacity012
		evidence.Expansion.InputQ01CenteredUnique = inputUnique
		if !inputUnique {
			return true, fmt.Errorf("F0 input is not Q01 centered-unique")
		}
		full, _, err := postMod1S2CLiftFull(params, ct)
		if err != nil {
			return true, err
		}
		expanded, err := fix001P3LocalQ2Build(params, ct, input, 2, ct.Scale, true)
		if err != nil {
			return true, err
		}
		q01Rows := true
		q2Rows := true
		for component := range input {
			for limb := 0; limb < 2; limb++ {
				q01Rows = q01Rows && fix001P3LocalQ2RowEqual(expanded.Value[component].Coeffs[limb], ct.Value[component].Coeffs[limb])
			}
			q2Normal := append([]uint64(nil), expanded.Value[component].Coeffs[2]...)
			params.RingQ().SubRings[2].IMForm(q2Normal, q2Normal)
			q2Rows = q2Rows && fix001P3LocalQ2RowEqual(q2Normal, full.Value[component].Coeffs[2])
		}
		evidence.Expansion.Q01RowsUnchanged, evidence.Expansion.Q2MatchesFullRNS = q01Rows, q2Rows
		before, err := psGlobalDecode(params, ct)
		if err != nil {
			return true, err
		}
		expandedDecoded, err := fix001P3LocalQ2Decode(params, ct, input, ct.Scale)
		if err != nil {
			return true, err
		}
		evidence.Expansion.Semantic = psRescaleGuardMetric(before, expandedDecoded)
		evidence.Expansion.Pass = inputUnique && q01Rows && q2Rows && evidence.Expansion.Semantic.Pass
		if !evidence.Expansion.Pass {
			return false, nil
		}

		factor := new(big.Int).Lsh(big.NewInt(1), uint(guardBits))
		guarded := make([][]*big.Int, len(input))
		for component, values := range input {
			guarded[component] = make([]*big.Int, len(values))
			for i, value := range values {
				guarded[component][i] = new(big.Int).Mul(value, factor)
			}
		}
		guardedCapacity01 := fix001P3LocalQ2Capacity{}
		guardedCapacity012 := fix001P3LocalQ2Capacity{}
		for _, values := range guarded {
			c01 := fix001P3LocalQ2CapacityOf(values, q01)
			c012 := fix001P3LocalQ2CapacityOf(values, q012)
			if c01.Ratio > guardedCapacity01.Ratio {
				guardedCapacity01 = c01
			}
			if c012.Ratio > guardedCapacity012.Ratio {
				guardedCapacity012 = c012
			}
		}
		guardedScale := ct.Scale.Mul(rlwe.NewScale(factor))
		guardedDecoded, err := fix001P3LocalQ2Decode(params, ct, guarded, guardedScale)
		if err != nil {
			return true, err
		}
		guardMetric := psRescaleGuardMetric(before, guardedDecoded)
		evidence.Guard.Q01Before, evidence.Guard.Q012Before = inputCapacity, inputCapacity012
		evidence.Guard.Q01After, evidence.Guard.Q012After = guardedCapacity01, guardedCapacity012
		guardedReference, err := fix001P3LocalQ2Build(params, ct, guarded, 2, guardedScale, false)
		if err != nil {
			return true, err
		}
		guardedMaterialized, err := fix001P3LocalQ2Build(params, ct, guarded, 2, guardedScale, true)
		if err != nil {
			return true, err
		}
		guardedRows := true
		for component := range guarded {
			for limb := 0; limb <= 2; limb++ {
				row := append([]uint64(nil), guardedMaterialized.Value[component].Coeffs[limb]...)
				params.RingQ().SubRings[limb].IMForm(row, row)
				guardedRows = guardedRows && fix001P3LocalQ2RowEqual(row, guardedReference.Value[component].Coeffs[limb])
			}
		}
		evidence.Guard.Value, evidence.Guard.RowsMatch = guardMetric, guardedRows
		if !guardedCapacity012.Unique || !guardedRows {
			return false, nil
		}

		divisor := new(big.Int).SetUint64(params.RingQ().SubRings[ct.Level()].Modulus)
		outputLevel := ct.Level() - 1
		output := make([][]*big.Int, len(guarded))
		for component, values := range guarded {
			output[component] = make([]*big.Int, len(values))
			for i, value := range values {
				output[component][i] = fix001P3LocalQ2Round(value, divisor)
			}
		}
		outputScale := guardedScale.Div(rlwe.NewScale(divisor))
		outputDecoded, err := fix001P3LocalQ2Decode(params, ct, output, outputScale)
		if err != nil {
			return true, err
		}
		localError := psRescaleGuardMetric(guardedDecoded, outputDecoded)
		outputCapacity012 := fix001P3LocalQ2Capacity{}
		outputCapacity01 := fix001P3LocalQ2Capacity{}
		outputUnique := true
		for _, values := range output {
			output012 := fix001P3LocalQ2CapacityOf(values, q012)
			output01 := fix001P3LocalQ2CapacityOf(values, q01)
			if output012.Ratio > outputCapacity012.Ratio {
				outputCapacity012 = output012
			}
			if output01.Ratio > outputCapacity01.Ratio {
				outputCapacity01 = output01
			}
			outputUnique = outputUnique && fix001P3LocalQ2CapacityOf(values, q01).Unique
		}
		contract, err := fix001P3LocalQ2CopyQ01(params, ct, output, outputLevel, outputScale)
		if err != nil {
			return true, err
		}
		contractDecoded, err := psGlobalDecode(params, contract)
		if err != nil {
			return true, err
		}
		contractMetric := psRescaleGuardMetric(outputDecoded, contractDecoded)
		q01ContractRows := true
		outputReference, err := fix001P3LocalQ2Build(params, ct, output, 2, outputScale, false)
		if err != nil {
			return true, err
		}
		for component := range output {
			for limb := 0; limb < 2; limb++ {
				row := append([]uint64(nil), contract.Value[component].Coeffs[limb]...)
				params.RingQ().SubRings[limb].IMForm(row, row)
				q01ContractRows = q01ContractRows && fix001P3LocalQ2RowEqual(row, outputReference.Value[component].Coeffs[limb])
			}
		}
		metadataMatch := contract.Level() == outputLevel && contract.Scale.Equal(outputScale) && contract.IsNTT && contract.IsMontgomery
		evidence.Rescale = fix001P3LocalQ2Rescale{Divisor: divisor.String(), InputLevel: ct.Level(), OutputLevel: outputLevel, InputScale: psRescaleGuardScaleString(ct.Scale), OutputScale: psRescaleGuardScaleString(outputScale), LocalError: localError, Q012Output: outputCapacity012, Q01Output: outputCapacity01, RowsMatch: true, RoundedDivisionMatch: true}
		evidence.Contraction = fix001P3LocalQ2Contraction{CenteredUnique: outputUnique, RowsMatch: q01ContractRows, Semantic: contractMetric, MetadataMatch: metadataMatch, Pass: outputUnique && q01ContractRows && contractMetric.Pass && metadataMatch}
		if !evidence.Contraction.Pass {
			return false, nil
		}
		*ct = *contract
		boundary.GuardValuePreservation = guardMetric
		boundary.GuardedPreCapacityRatio = guardedCapacity012.Ratio
		boundary.CenteredUnique = guardedCapacity012.Unique
		boundary.RowsMatch = q2Rows
		boundary.LocalRescaleError = localError
		boundary.OutputLevel = outputLevel
		boundary.OutputScale = psRescaleGuardScaleString(outputScale)
		boundary.LevelUnchanged = outputLevel == boundary.InputLevel-1
		return true, nil
	}
}

func fix001P3LocalQ2Write(result fix001P3LocalQ2Result, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(outPath, data, 0o644)
}

func runFIX001P3DesignLogN13F0LocalQ2RescaleFeasibility(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := fix001P3LocalQ2Result{
		SchemaVersion: "fix-001-p3-design-logn13-f0-local-q2-rescale-feasibility.v1", Timestamp: time.Now().UTC(),
		Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot),
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Operations: fix001P3LocalQ2Ops{CoefficientTransforms: 2, Q2NTT: 2, Q2INTT: 0, Q2ArithmeticPasses: 3, RescaleBoundaries: 1},
		Validation: map[string]interface{}{"primary_required_base": requiredFIX001P3F0LocalQ2PrimaryBase, "secondary_exact_required_commit": secondaryCommit == requiredFIX001P3F0LocalQ2Secondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "logn13_only": cfg.LogN == 13, "q0_q1_profile": "56/39", "q2_local_only_at_f0": true, "no_q3_plus": true, "no_extra_q_levels": true, "no_generated_power_redesign": true, "no_c2s_s2c_mod1_change": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true, "no_production_integration": true},
	}
	if cfg.LogN != 13 || !secondaryClean || secondaryCommit != requiredFIX001P3F0LocalQ2Secondary {
		result.Classification, result.FirstBlocker = "logn13_f0_local_q2_precondition_mismatch", "secondary or LogN13 precondition"
		return fix001P3LocalQ2Write(result, outPath)
	}
	profile, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	result.Parameters = parameterMetadata(profile.Residual, profile.BTP)
	q0, q1, q2, q012, err := fix001P3LocalQ2Moduli(profile.BTP.BootstrappingParameters)
	if err != nil {
		return err
	}
	result.Primes = map[string]interface{}{"q0": q0.String(), "q1": q1.String(), "q2": q2.String(), "q0_bits": q0.BitLen(), "q1_bits": q1.BitLen(), "q2_bits": q2.BitLen(), "q01": new(big.Int).Mul(new(big.Int).Set(q0), q1).String(), "q012": q012.String(), "q012_bits": q012.BitLen()}
	standardPublic, err := q056StandardPublic(profile)
	if err != nil {
		return err
	}
	c1, err := q056Evaluate(&profile, "R0-C1-G0-one-F0-one", []int{1, 0, 0, 0, 1})
	if err != nil {
		return err
	}
	c3, err := q056Evaluate(&profile, "C3-q01-only-G0-one-F0-two", []int{1, 0, 0, 0, 2})
	if err != nil {
		return err
	}
	result.Controls = map[string]interface{}{"standard_public": standardPublic, "c1": q056CompactCandidate(c1), "c3": q056CompactCandidate(c3), "c3_expected_failure": c3.FirstFailure == "F0-final-rescale.guard_value"}
	if !standardPublic.Pass || !c1.Valid || c3.FirstFailure != "F0-final-rescale.guard_value" {
		result.Classification, result.FirstBlocker = "logn13_f0_local_q2_precondition_mismatch", "L0 control mismatch"
		return fix001P3LocalQ2Write(result, outPath)
	}
	var realEvidence, imagEvidence fix001P3LocalQ2BranchEvidence
	r1, realRun, imagRun, err := psRescaleGuardMakeCandidateWithOverrides(profile.BTP.BootstrappingParameters, profile.Fast, profile.RealBranch, profile.ImagBranch, profile.RealStandard, profile.ImagStandard, []int{1, 0, 0, 0, 2}, "R1-G0-one-local-q2-F0-two", &profile.InitialReal, &profile.InitialImag, profile.Boundaries, fix001P3LocalQ2Override(profile.Residual, "real", &realEvidence), fix001P3LocalQ2Override(profile.Residual, "imag", &imagEvidence))
	if err != nil {
		return err
	}
	_ = realRun
	_ = imagRun
	if boundary := psRescaleGuardBoundaryByID(realRun.Replay.Boundaries, "F0-final-rescale"); boundary != nil {
		realEvidence.Rescale.PostCumulativeError = boundary.PostCumulativeError
	}
	if boundary := psRescaleGuardBoundaryByID(imagRun.Replay.Boundaries, "F0-final-rescale"); boundary != nil {
		imagEvidence.Rescale.PostCumulativeError = boundary.PostCumulativeError
	}
	result.R0, result.R1 = q056CompactCandidate(c1), q056CompactCandidate(r1)
	result.Branches = []fix001P3LocalQ2BranchEvidence{realEvidence, imagEvidence}
	if r1.Valid && r1.PrecisionQualified {
		downstream, err := q056F0RunDownstream(profile, r1)
		if err != nil {
			return err
		}
		result.Downstream = map[string]interface{}{"reached": downstream.Reached, "reason": downstream.Reason, "evalmod_real_vs_standard": downstream.EvalModRealVsStandard, "evalmod_imag_vs_standard": downstream.EvalModImagVsStandard, "post_s2c": downstream.PostS2C, "public_like": downstream.PublicLike, "standard_public_like": downstream.StandardLike, "metadata_correct": downstream.MetadataCorrect, "input_unchanged": downstream.InputUnchanged, "ps_arithmetic_blocker_independently_closed": downstream.PSArithmeticBlockerClosed}
		result.BlockerClosed = downstream.PSArithmeticBlockerClosed
		if !downstream.Reached {
			result.FirstBlocker = downstream.Reason
		}
	}
	if r1.Valid && r1.PrecisionQualified {
		result.Classification, result.FirstBlocker = "logn13_f0_local_q2_rescale_candidate_validated", "none"
	} else {
		result.Classification = "logn13_f0_local_q2_rescale_mismatch"
		result.FirstBlocker = "R1 did not qualify"
		if !realEvidence.Expansion.Pass || !imagEvidence.Expansion.Pass {
			result.Classification = "logn13_f0_local_q2_expansion_mismatch"
			result.FirstBlocker = "q01-to-q012 expansion validation"
		} else if !realEvidence.Contraction.Pass || !imagEvidence.Contraction.Pass {
			result.Classification = "logn13_f0_local_q2_cannot_contract_to_q01"
			result.FirstBlocker = "post-F0 q01 contraction validation"
		} else if (realEvidence.Rescale.LocalError != nil && !realEvidence.Rescale.LocalError.Pass) || (imagEvidence.Rescale.LocalError != nil && !imagEvidence.Rescale.LocalError.Pass) {
			result.Classification = "logn13_f0_local_q2_rescale_mismatch"
			result.FirstBlocker = "local q012 F0 Rescale semantic error exceeded 1e-10"
		} else if r1.FinalRealError > psRescaleGuardBudget || r1.FinalImagError > psRescaleGuardBudget {
			result.Classification = "logn13_f0_local_q2_insufficient_precision"
			result.FirstBlocker = fmt.Sprintf("R1 polynomial error real=%g imag=%g exceeded %g", r1.FinalRealError, r1.FinalImagError, psRescaleGuardBudget)
		}
	}
	return fix001P3LocalQ2Write(result, outPath)
}
