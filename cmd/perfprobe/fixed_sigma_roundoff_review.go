package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"reflect"
)

const fixedSigmaRoundoffReviewSchema = "fixed-sigma-direct-mul-roundoff-review.v1"

type fixedSigmaRoundoffReview struct {
	SchemaVersion       string                               `json:"schema_version"`
	Status              string                               `json:"status"`
	Classification      string                               `json:"classification"`
	ManifestSHA256      string                               `json:"manifest_sha256"`
	PriorAnalysisSHA256 string                               `json:"prior_analysis_sha256"`
	PriorAnalysisStatus string                               `json:"prior_analysis_status"`
	Arithmetic          fixedSigmaRoundoffArithmetic         `json:"arithmetic"`
	Pairs               []fixedSigmaRoundoffPair             `json:"pairs"`
	FailureMaskAnalysis fixedSigmaRoundoffFailureMaskSummary `json:"old_gamma64_failure_mask_analysis"`
}

type fixedSigmaRoundoffArithmetic struct {
	InputInterpretation string   `json:"input_interpretation"`
	ExactMethod         string   `json:"high_precision_method"`
	FirstOrderModel     string   `json:"first_order_model_residual"`
	ClosureIdentity     string   `json:"full_identity_closure"`
	UnitRoundoff        string   `json:"unit_roundoff"`
	AddSubBound         string   `json:"complex_add_sub_bound_l1"`
	MultiplyBound       string   `json:"complex_multiply_bound_l1"`
	FullBound           string   `json:"full_closure_bound_derivation"`
	Assumptions         []string `json:"bound_assumptions"`
}

type fixedSigmaRoundoffPair struct {
	Workload                        string                   `json:"workload"`
	StageSeeds                      [2]int64                 `json:"requested_stage_seeds"`
	BaselineOrigin                  string                   `json:"baseline_origin,omitempty"`
	BaselineSourceSeeds             [2]int64                 `json:"baseline_source_stage_seeds"`
	BaselinePrimaryCommit           string                   `json:"baseline_primary_commit"`
	BaselinePrimarySourceSHA256     string                   `json:"baseline_primary_measurement_source_sha256"`
	BaselineBackendCommit           string                   `json:"baseline_backend_commit"`
	BaselineBinarySHA256            string                   `json:"baseline_binary_sha256"`
	BaselineResultSHA256            string                   `json:"baseline_result_sha256"`
	BaselineVectorsSHA256           string                   `json:"baseline_vectors_sha256"`
	ONOFFPrimaryCommit              string                   `json:"on_off_primary_commit"`
	ONOFFPrimarySourceSHA256        string                   `json:"on_off_primary_measurement_source_sha256"`
	ONOFFBackendCommit              string                   `json:"on_off_backend_commit"`
	ONOFFBinarySHA256               string                   `json:"on_off_binary_sha256"`
	ONOFFResultSHA256               string                   `json:"on_off_result_sha256"`
	ONOFFVectorsSHA256              string                   `json:"on_off_vectors_sha256"`
	OldGamma64FailureCount          int                      `json:"old_gamma64_failure_count"`
	OldGamma64FailureMaskSHA256     string                   `json:"old_gamma64_failure_mask_sha256"`
	OldGamma64FirstFailureIndices   []int                    `json:"old_gamma64_first_failure_indices"`
	OldGamma64PassIndices           []int                    `json:"old_gamma64_pass_indices"`
	OldGamma64MaxResidual           float64                  `json:"old_gamma64_max_residual"`
	OldGamma64MaxTolerance          float64                  `json:"old_gamma64_max_tolerance"`
	OldGamma64MinFailureRatio       float64                  `json:"old_gamma64_min_failure_ratio"`
	OldGamma64MaxPassRatio          float64                  `json:"old_gamma64_max_pass_ratio"`
	OldGamma64PassMaxProductAbs     float64                  `json:"old_gamma64_pass_max_abs_x_times_c"`
	OldGamma64FailureMinProductAbs  float64                  `json:"old_gamma64_failure_min_abs_x_times_c"`
	CExactZeroSlots                 int                      `json:"decoded_c_exact_zero_slots"`
	Complex128Closure               fixedSigmaVectorAnalysis `json:"complex128_full_closure_residual"`
	Complex128MaxClosureL1          string                   `json:"complex128_max_closure_l1"`
	HighPrecisionExactClosureZero   bool                     `json:"high_precision_exact_closure_zero"`
	HighPrecisionIdentityMismatch   int                      `json:"high_precision_identity_mismatch_slots"`
	HighPrecisionFirstOrderResidual fixedSigmaVectorAnalysis `json:"high_precision_first_order_unexplained_residual"`
	Complex128FirstOrderResidual    fixedSigmaVectorAnalysis `json:"complex128_first_order_unexplained_residual"`
	FirstOrderMaxFloat64VsExactL1   string                   `json:"first_order_max_complex128_vs_exact_l1"`
	DerivedBoundViolationCount      int                      `json:"derived_bound_violation_count"`
	DerivedBoundMaxL1               string                   `json:"derived_bound_max_l1"`
	MaxClosureToBoundRatio          string                   `json:"max_closure_to_bound_ratio"`
	MaxIntermediateL1               map[string]string        `json:"max_intermediate_magnitude_l1"`
	MaxFloat64RoundingErrorL1       map[string]string        `json:"max_complex128_rounding_error_l1_by_intermediate"`
	NewBoundFailureMaskSHA256       string                   `json:"derived_bound_violation_mask_sha256"`
}

type fixedSigmaRoundoffFailureMaskSummary struct {
	AllOldFailureMasksIdentical bool  `json:"all_old_failure_masks_identical"`
	IntersectionFailureSlots    int   `json:"intersection_failure_slots"`
	UnionFailureSlots           int   `json:"union_failure_slots"`
	DifferentIndicesFirst       []int `json:"different_failure_indices_first"`
	AllPairsShareSamePassSlots  bool  `json:"all_pairs_share_same_pass_slots"`
}

type exactComplex struct {
	re *big.Rat
	im *big.Rat
}

func runFixedSigmaPropagationRoundoffReview(args []string) error {
	flags := flag.NewFlagSet("perfprobe review-fixed-sigma-roundoff", flag.ContinueOnError)
	manifestPath := flags.String("manifest", "", "captured direct-ctct ablation manifest")
	priorAnalysisPath := flags.String("prior-analysis", "", "unchanged original analysis.json")
	outPath := flags.String("out", "", "new independent high-precision review JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *manifestPath == "" || *priorAnalysisPath == "" || *outPath == "" {
		return errors.New("--manifest, --prior-analysis and --out are required")
	}
	report, err := buildFixedSigmaPropagationRoundoffReview(*manifestPath, *priorAnalysisPath)
	if err != nil {
		return err
	}
	if err := writeExclusiveJSON(*outPath, report); err != nil {
		return err
	}
	fmt.Printf("fixed-sigma roundoff review pairs=%d old_gamma64_failures=%d/%d/%d/%d derived_bound_violations=%d result=%s\n",
		len(report.Pairs), report.Pairs[0].OldGamma64FailureCount, report.Pairs[1].OldGamma64FailureCount,
		report.Pairs[2].OldGamma64FailureCount, report.Pairs[3].OldGamma64FailureCount,
		totalRoundoffBoundViolations(report), *outPath)
	return nil
}

func buildFixedSigmaPropagationRoundoffReview(manifestPath, priorAnalysisPath string) (fixedSigmaRoundoffReview, error) {
	var manifest fixedSigmaAblationManifest
	if err := readJSONBounded(manifestPath, &manifest, fixedSigmaAblationManifestMaxBytes); err != nil {
		return fixedSigmaRoundoffReview{}, fmt.Errorf("read captured manifest: %w", err)
	}
	if manifest.SchemaVersion != fixedSigmaDirectMulAblationManifestSchema || len(manifest.Workloads) != 2 {
		return fixedSigmaRoundoffReview{}, errors.New("roundoff review requires the frozen two-workload direct-ctct manifest")
	}
	var prior fixedSigmaAblationComparison
	if err := readJSONBounded(priorAnalysisPath, &prior, 16<<20); err != nil {
		return fixedSigmaRoundoffReview{}, fmt.Errorf("read prior analysis: %w", err)
	}
	if prior.SchemaVersion != "fixed-sigma-direct-ctct-ablation-analysis.v1" ||
		prior.Status != "FAIL_PROPAGATION_DECOMPOSITION_ROUNDOFF_CHECK" ||
		prior.Classification != "BATCH_BLOCKED_NEEDS_WEB_REVIEW" ||
		prior.BootstrapCalls != 0 || prior.FastChainRuns != 16 || prior.StandardChainRuns != 2 ||
		len(prior.Workloads) != 2 {
		return fixedSigmaRoundoffReview{}, errors.New("prior analysis is not the preserved completed 16-Fast/2-Standard blocked result")
	}
	manifestSHA, err := sha256File(manifestPath)
	if err != nil {
		return fixedSigmaRoundoffReview{}, err
	}
	priorSHA, err := sha256File(priorAnalysisPath)
	if err != nil {
		return fixedSigmaRoundoffReview{}, err
	}
	report := fixedSigmaRoundoffReview{
		SchemaVersion:       fixedSigmaRoundoffReviewSchema,
		Status:              "OFFLINE_HIGH_PRECISION_RECOMPUTE_COMPLETE",
		Classification:      "RMS_SCREEN_PASS_DECOMPOSITION_ROUNDOFF_REVIEW_REQUIRED",
		ManifestSHA256:      manifestSHA,
		PriorAnalysisSHA256: priorSHA,
		PriorAnalysisStatus: prior.Status,
		Arithmetic: fixedSigmaRoundoffArithmetic{
			InputInterpretation: "Every decoded JSON float is treated as its exact IEEE-754 binary64 value; this review bounds only the subsequent complex128 algebra, not CKKS/decode error.",
			ExactMethod:         "math/big.Rat exact rational arithmetic over the binary64 inputs; algebraic identity and first-order remainder are evaluated without intermediate rounding.",
			FirstOrderModel:     "U = P1 - d1*C. U is the Stage2 operation-residual change r2a-r2b and is not a floating-point closure error or a claim about a specific CKKS primitive.",
			ClosureIdentity:     "E = P1 - (d1*C + (r2a-r2b)); the exact-rational identity is zero for every slot when all terms use the same decoded binary64 inputs.",
			UnitRoundoff:        "u=2^-53 for round-to-nearest binary64; the old gate used 2^-52 as u and gamma_64 over only final small terms.",
			AddSubBound:         "For L(z)=|Re(z)|+|Im(z)|, each complex add/sub absolute error is bounded by u*(L(a)+L(b)), absent overflow/underflow.",
			MultiplyBound:       "The componentwise two-product-plus-add/sub complex multiply has L1 error <= gamma_2*L(a)*L(b), gamma_2=2u/(1-2u), absent overflow/underflow.",
			FullBound:           "Per slot: eP=u(L(Yon)+L(Yoff)); eD=u(L(Xon)+L(Xoff)); ePred=L(C)*eD+gamma_2*L(Dfl)*L(C); eA=gamma_2*L(Xon)*L(C)+u(L(Yon)+L((Xon*C)fl)); eB is analogous; eDiff=u(L(rAfl)+L(rBfl)); eRhs=ePred+eA+eB+eDiff+u(L(predfl)+L(diffFl)); Ebound=eP+eRhs+u(L(P1fl)+L(rhsfl)). The exact rational L1 envelope is compared against the L1 norm of the complex128 closure residual.",
			Assumptions: []string{
				"The error bound follows the existing operation order in computeFixedSigmaAblationPropagation: subtract P1 and d1, multiply d1*C and X*C, subtract r2a/r2b, subtract their difference, add the RHS, then close with the final subtraction.",
				"All input and intermediate values are checked finite; captured magnitudes are normal-range or exact zero, so the standard relative roundoff model's overflow/underflow exclusions hold.",
				"L1 bounds are conservative upper bounds on complex Euclidean magnitude; the bound is operation-derived, not a tuned multiplier or a change to the original gate.",
			},
		},
	}
	priorByWorkload := make(map[string]fixedSigmaAblationWorkloadResult, len(prior.Workloads))
	for _, workload := range prior.Workloads {
		priorByWorkload[workload.Name] = workload
	}
	for _, workload := range manifest.Workloads {
		priorWorkload, ok := priorByWorkload[workload.Name]
		if !ok || workload.CVariant != workload.Name || len(workload.SeedPairs) != 2 || len(priorWorkload.SeedPairs) != 2 {
			return fixedSigmaRoundoffReview{}, fmt.Errorf("workload %s is not a complete frozen direct-ctct pair set", workload.Name)
		}
		standard, err := loadFixedSigmaAblationArtifact(workload.Standard)
		if err != nil {
			return fixedSigmaRoundoffReview{}, fmt.Errorf("workload %s Standard provenance: %w", workload.Name, err)
		}
		for _, seedPair := range workload.SeedPairs {
			oldPair, ok := priorPairBySeed(priorWorkload, seedPair.Stage1Seed)
			if !ok {
				return fixedSigmaRoundoffReview{}, fmt.Errorf("prior analysis is missing %s seed %d", workload.Name, seedPair.Stage1Seed)
			}
			recomputed, err := analyzeFixedSigmaAblationSeedPair(standard, workload.CVariant, seedPair)
			if err != nil {
				return fixedSigmaRoundoffReview{}, fmt.Errorf("revalidate %s seed %d captures: %w", workload.Name, seedPair.Stage1Seed, err)
			}
			if !reflect.DeepEqual(recomputed.ArtifactSHA256, oldPair.ArtifactSHA256) ||
				!reflect.DeepEqual(recomputed.Propagation, oldPair.Propagation) ||
				recomputed.BaselineOrigin != oldPair.BaselineOrigin {
				return fixedSigmaRoundoffReview{}, fmt.Errorf("%s seed %d does not reproduce the preserved analysis inputs and complex128 result", workload.Name, seedPair.Stage1Seed)
			}
			baseRef := seedPair.Arms["OFF/OFF"]
			if seedPair.BaselineReuse != nil {
				baseRef = seedPair.Arms["OFF/OFF"]
			}
			base, err := loadFixedSigmaAblationArtifact(baseRef)
			if err != nil {
				return fixedSigmaRoundoffReview{}, fmt.Errorf("%s seed %d OFF/OFF vectors: %w", workload.Name, seedPair.Stage1Seed, err)
			}
			onOff, err := loadFixedSigmaAblationArtifact(seedPair.Arms["ON/OFF"])
			if err != nil {
				return fixedSigmaRoundoffReview{}, fmt.Errorf("%s seed %d ON/OFF vectors: %w", workload.Name, seedPair.Stage1Seed, err)
			}
			pair, err := reviewFixedSigmaPropagationPair(workload.Name, seedPair.Stage1Seed, recomputed.BaselineOrigin, base, onOff)
			if err != nil {
				return fixedSigmaRoundoffReview{}, fmt.Errorf("%s seed %d offline arithmetic: %w", workload.Name, seedPair.Stage1Seed, err)
			}
			report.Pairs = append(report.Pairs, pair)
		}
	}
	if len(report.Pairs) != 4 {
		return fixedSigmaRoundoffReview{}, fmt.Errorf("roundoff review produced %d pairs, want four", len(report.Pairs))
	}
	report.FailureMaskAnalysis, err = compareRoundoffFailureMasks(report.Pairs)
	if err != nil {
		return fixedSigmaRoundoffReview{}, err
	}
	if totalRoundoffBoundViolations(report) == 0 {
		report.Status = "DERIVED_ROUNDOFF_ENVELOPE_COVERS_ALL_COMPLEX128_CLOSURES"
	} else {
		report.Status = "DERIVED_ROUNDOFF_ENVELOPE_VIOLATIONS_FOUND"
	}
	return report, nil
}

func priorPairBySeed(workload fixedSigmaAblationWorkloadResult, seed int64) (fixedSigmaAblationSeedPairResult, bool) {
	for _, pair := range workload.SeedPairs {
		if pair.Stage1Seed == seed {
			return pair, true
		}
	}
	return fixedSigmaAblationSeedPairResult{}, false
}

func reviewFixedSigmaPropagationPair(workload string, stage1Seed int64, baselineOrigin string, baseline, onOff loadedFixedSigmaAblationArtifact) (fixedSigmaRoundoffPair, error) {
	const slots = 1 << 12
	get := func(artifact loadedFixedSigmaAblationArtifact, name string) ([]complex128, error) {
		values, ok := artifact.vectors.Checkpoints[name]
		if !ok || len(values) != slots {
			return nil, fmt.Errorf("checkpoint %s is missing or has %d values", name, len(values))
		}
		decoded := decodeComplexValues(values)
		if !allFinite(decoded) {
			return nil, fmt.Errorf("checkpoint %s contains a non-finite value", name)
		}
		return decoded, nil
	}
	xOff, err := get(baseline, "chain_stage1_output_level4")
	if err != nil {
		return fixedSigmaRoundoffPair{}, err
	}
	xOn, err := get(onOff, "chain_stage1_output_level4")
	if err != nil {
		return fixedSigmaRoundoffPair{}, err
	}
	yOff, err := get(baseline, "chain_stage2_output_level3")
	if err != nil {
		return fixedSigmaRoundoffPair{}, err
	}
	yOn, err := get(onOff, "chain_stage2_output_level3")
	if err != nil {
		return fixedSigmaRoundoffPair{}, err
	}
	c, err := get(baseline, "encrypt_c_q4_level4")
	if err != nil {
		return fixedSigmaRoundoffPair{}, err
	}
	cOn, err := get(onOff, "encrypt_c_q4_level4")
	if err != nil {
		return fixedSigmaRoundoffPair{}, err
	}
	if !complexVectorsEqual(c, cOn) {
		return fixedSigmaRoundoffPair{}, errors.New("decoded C input differs between OFF/OFF and ON/OFF")
	}
	p1 := subtractComplexVectors(yOn, yOff)
	delta := subtractComplexVectors(xOn, xOff)
	prediction := multiplyComplexVectors(delta, c)
	productOn := multiplyComplexVectors(xOn, c)
	productOff := multiplyComplexVectors(xOff, c)
	r2a := subtractComplexVectors(yOn, productOn)
	r2b := subtractComplexVectors(yOff, productOff)
	residualDifference := subtractComplexVectors(r2a, r2b)
	rhs := addComplexVectors(prediction, residualDifference)
	closure := subtractComplexVectors(p1, rhs)
	firstOrder := subtractComplexVectors(p1, prediction)
	if len(closure) != slots || !allFinite(closure) || !allFinite(firstOrder) {
		return fixedSigmaRoundoffPair{}, errors.New("complex128 recomputation produced invalid vectors")
	}
	for _, vector := range [][]complex128{xOff, xOn, yOff, yOn, c, p1, delta, prediction, productOn, productOff, r2a, r2b, residualDifference, rhs, closure, firstOrder} {
		for _, value := range vector {
			if math.IsInf(real(value), 0) || math.IsNaN(real(value)) || math.IsInf(imag(value), 0) || math.IsNaN(imag(value)) {
				return fixedSigmaRoundoffPair{}, errors.New("captured roundoff envelope includes a non-finite intermediate")
			}
			if complexHasSubnormalComponent(value) {
				return fixedSigmaRoundoffPair{}, errors.New("captured roundoff envelope includes a subnormal component; standard relative-error assumptions do not apply")
			}
		}
	}
	const oldUnitRoundoff = 0x1p-52
	oldGamma64 := 64 * oldUnitRoundoff / (1 - 64*oldUnitRoundoff)
	oldFailures := make([]bool, slots)
	oldFailureIndices := make([]int, 0, slots)
	oldPassIndices := make([]int, 0, slots)
	newFailures := make([]bool, slots)
	maxOldResidual, maxOldTolerance := 0.0, 0.0
	minOldFailureRatio, maxOldPassRatio := math.Inf(1), 0.0
	maxPassProductAbs, minFailureProductAbs := 0.0, math.Inf(1)
	cExactZeroSlots := 0
	maxComplexL1 := new(big.Rat)
	maxDerivedBound := new(big.Rat)
	maxRatio := new(big.Rat)
	maxModelDifference := new(big.Rat)
	maxIntermediate := map[string]*big.Rat{}
	maxRounding := map[string]*big.Rat{}
	modelExactValues := make([]complex128, slots)
	identityMismatches, derivedFailures := 0, 0
	for index := 0; index < slots; index++ {
		oldResidual := cmplx.Abs(closure[index])
		oldTolerance := oldGamma64 * (cmplx.Abs(p1[index]) + cmplx.Abs(prediction[index]) + cmplx.Abs(r2a[index]) + cmplx.Abs(r2b[index]))
		if oldResidual > maxOldResidual {
			maxOldResidual = oldResidual
		}
		if oldTolerance > maxOldTolerance {
			maxOldTolerance = oldTolerance
		}
		oldFailures[index] = oldResidual > oldTolerance
		oldRatio := math.Inf(1)
		if oldTolerance > 0 {
			oldRatio = oldResidual / oldTolerance
		}
		if oldFailures[index] {
			oldFailureIndices = append(oldFailureIndices, index)
			if oldRatio < minOldFailureRatio {
				minOldFailureRatio = oldRatio
			}
			for _, product := range []complex128{productOn[index], productOff[index]} {
				magnitude := cmplx.Abs(product)
				if magnitude < minFailureProductAbs {
					minFailureProductAbs = magnitude
				}
			}
		} else {
			oldPassIndices = append(oldPassIndices, index)
			if oldRatio > maxOldPassRatio {
				maxOldPassRatio = oldRatio
			}
			for _, product := range []complex128{productOn[index], productOff[index]} {
				magnitude := cmplx.Abs(product)
				if magnitude > maxPassProductAbs {
					maxPassProductAbs = magnitude
				}
			}
		}

		eXOff, eXOn := exactFromComplex(xOff[index]), exactFromComplex(xOn[index])
		eYOff, eYOn, eC := exactFromComplex(yOff[index]), exactFromComplex(yOn[index]), exactFromComplex(c[index])
		if c[index] == 0 {
			cExactZeroSlots++
		}
		eP1 := exactSub(eYOn, eYOff)
		eDelta := exactSub(eXOn, eXOff)
		eProductOn, eProductOff := exactMul(eXOn, eC), exactMul(eXOff, eC)
		eR2A, eR2B := exactSub(eYOn, eProductOn), exactSub(eYOff, eProductOff)
		ePrediction := exactMul(eDelta, eC)
		eFirstOrder := exactSub(eP1, ePrediction)
		eResidualDifference := exactSub(eR2A, eR2B)
		eClosure := exactSub(eP1, exactAdd(ePrediction, eResidualDifference))
		if !exactEqual(eFirstOrder, eResidualDifference) || !exactIsZero(eClosure) {
			identityMismatches++
		}
		modelExactValues[index] = exactToComplex128(eFirstOrder)

		actuals := map[string]complex128{
			"x_on": xOn[index], "x_off": xOff[index], "y_on": yOn[index], "y_off": yOff[index], "c": c[index],
			"p1": p1[index], "d1": delta[index], "x_on_times_c": productOn[index], "x_off_times_c": productOff[index],
			"d1_times_c": prediction[index], "r2a": r2a[index], "r2b": r2b[index],
			"r2a_minus_r2b": residualDifference[index], "decomposition_rhs": rhs[index], "closure": closure[index],
		}
		for name, value := range actuals {
			updateMaxRat(maxIntermediate, name, exactL1(exactFromComplex(value)))
		}
		rounding := map[string]*big.Rat{
			"p1":                      exactL1(exactSub(exactFromComplex(p1[index]), eP1)),
			"d1":                      exactL1(exactSub(exactFromComplex(delta[index]), eDelta)),
			"x_on_times_c":            exactL1(exactSub(exactFromComplex(productOn[index]), eProductOn)),
			"x_off_times_c":           exactL1(exactSub(exactFromComplex(productOff[index]), eProductOff)),
			"r2a":                     exactL1(exactSub(exactFromComplex(r2a[index]), eR2A)),
			"r2b":                     exactL1(exactSub(exactFromComplex(r2b[index]), eR2B)),
			"d1_times_c":              exactL1(exactSub(exactFromComplex(prediction[index]), ePrediction)),
			"r2a_minus_r2b":           exactL1(exactSub(exactFromComplex(residualDifference[index]), exactSub(eR2A, eR2B))),
			"decomposition_rhs":       exactL1(exactSub(exactFromComplex(rhs[index]), exactAdd(ePrediction, exactSub(eR2A, eR2B)))),
			"first_order_unexplained": exactL1(exactSub(exactFromComplex(firstOrder[index]), eFirstOrder)),
			"full_closure":            exactL1(exactFromComplex(closure[index])),
		}
		for name, value := range rounding {
			updateMaxRat(maxRounding, name, value)
		}
		modelDiff := exactL1(exactSub(exactFromComplex(firstOrder[index]), eFirstOrder))
		if modelDiff.Cmp(maxModelDifference) > 0 {
			maxModelDifference.Set(modelDiff)
		}

		eP := exactAddError(eYOn, eYOff)
		eD := exactAddError(eXOn, eXOff)
		ePred := exactAddRat(exactMulRat(exactL1(eC), eD), exactComplexMulError(exactFromComplex(delta[index]), eC))
		eA := exactAddRat(exactComplexMulError(eXOn, eC), exactAddError(eYOn, exactFromComplex(productOn[index])))
		eB := exactAddRat(exactComplexMulError(eXOff, eC), exactAddError(eYOff, exactFromComplex(productOff[index])))
		eDiff := exactAddError(exactFromComplex(r2a[index]), exactFromComplex(r2b[index]))
		eRhs := exactAddRat(exactAddRat(exactAddRat(ePred, eA), exactAddRat(eB, eDiff)), exactAddError(exactFromComplex(prediction[index]), exactFromComplex(residualDifference[index])))
		eBound := exactAddRat(exactAddRat(eP, eRhs), exactAddError(exactFromComplex(p1[index]), exactFromComplex(rhs[index])))
		residualL1 := exactL1(exactFromComplex(closure[index]))
		if eBound.Cmp(maxDerivedBound) > 0 {
			maxDerivedBound.Set(eBound)
		}
		if residualL1.Cmp(eBound) > 0 {
			newFailures[index] = true
			derivedFailures++
		}
		if eBound.Sign() > 0 {
			ratio := new(big.Rat).Quo(residualL1, eBound)
			if ratio.Cmp(maxRatio) > 0 {
				maxRatio.Set(ratio)
			}
		}
		if residualL1.Cmp(maxComplexL1) > 0 {
			maxComplexL1.Set(residualL1)
		}
	}
	newFailureIndices := make([]int, 0, derivedFailures)
	for index, failed := range newFailures {
		if failed {
			newFailureIndices = append(newFailureIndices, index)
		}
	}
	maskHash := func(mask []bool) string {
		bits := make([]byte, (len(mask)+7)/8)
		for i, bit := range mask {
			if bit {
				bits[i/8] |= byte(1 << (i % 8))
			}
		}
		digest := sha256.Sum256(bits)
		return hex.EncodeToString(digest[:])
	}
	firstFailures := oldFailureIndices
	if len(firstFailures) > 16 {
		firstFailures = firstFailures[:16]
	}
	return fixedSigmaRoundoffPair{
		Workload: workload, StageSeeds: [2]int64{stage1Seed, stage1Seed + 1},
		BaselineOrigin:        baselineOrigin,
		BaselineSourceSeeds:   baseline.doc.FixedSigmaChain.StageSeeds,
		BaselinePrimaryCommit: baseline.doc.PrimaryCommit, BaselinePrimarySourceSHA256: baseline.doc.PrimarySourceSHA256,
		BaselineBackendCommit: baseline.doc.BackendCommit,
		BaselineBinarySHA256:  baseline.doc.BinarySHA256,
		BaselineResultSHA256:  baseline.resultSHA, BaselineVectorsSHA256: baseline.vectorsSHA,
		ONOFFPrimaryCommit: onOff.doc.PrimaryCommit, ONOFFPrimarySourceSHA256: onOff.doc.PrimarySourceSHA256,
		ONOFFBackendCommit: onOff.doc.BackendCommit,
		ONOFFBinarySHA256:  onOff.doc.BinarySHA256,
		ONOFFResultSHA256:  onOff.resultSHA, ONOFFVectorsSHA256: onOff.vectorsSHA,
		OldGamma64FailureCount: len(oldFailureIndices), OldGamma64FailureMaskSHA256: maskHash(oldFailures),
		OldGamma64FirstFailureIndices: append([]int(nil), firstFailures...), OldGamma64PassIndices: oldPassIndices,
		OldGamma64MaxResidual: maxOldResidual, OldGamma64MaxTolerance: maxOldTolerance,
		OldGamma64MinFailureRatio: minOldFailureRatio, OldGamma64MaxPassRatio: maxOldPassRatio,
		OldGamma64PassMaxProductAbs: maxPassProductAbs, OldGamma64FailureMinProductAbs: minFailureProductAbs,
		CExactZeroSlots:   cExactZeroSlots,
		Complex128Closure: summarizeFixedSigmaVector(closure), Complex128MaxClosureL1: rationalUp(maxComplexL1),
		HighPrecisionExactClosureZero: identityMismatches == 0, HighPrecisionIdentityMismatch: identityMismatches,
		HighPrecisionFirstOrderResidual: summarizeFixedSigmaVector(modelExactValues),
		Complex128FirstOrderResidual:    summarizeFixedSigmaVector(firstOrder), FirstOrderMaxFloat64VsExactL1: rationalUp(maxModelDifference),
		DerivedBoundViolationCount: derivedFailures, DerivedBoundMaxL1: rationalUp(maxDerivedBound),
		MaxClosureToBoundRatio: rationalUp(maxRatio), MaxIntermediateL1: ratMapUp(maxIntermediate),
		MaxFloat64RoundingErrorL1: ratMapUp(maxRounding), NewBoundFailureMaskSHA256: maskHash(newFailures),
	}, nil
}

func totalRoundoffBoundViolations(report fixedSigmaRoundoffReview) int {
	total := 0
	for _, pair := range report.Pairs {
		total += pair.DerivedBoundViolationCount
	}
	return total
}

func exactFromFloat(value float64) *big.Rat {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return new(big.Rat)
	}
	return new(big.Rat).SetFloat64(value)
}

func exactFromComplex(value complex128) exactComplex {
	return exactComplex{re: exactFromFloat(real(value)), im: exactFromFloat(imag(value))}
}

func exactAddRat(a, b *big.Rat) *big.Rat { return new(big.Rat).Add(a, b) }
func exactSubRat(a, b *big.Rat) *big.Rat { return new(big.Rat).Sub(a, b) }
func exactMulRat(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }
func exactAbsRat(a *big.Rat) *big.Rat    { return new(big.Rat).Abs(a) }

func exactAdd(a, b exactComplex) exactComplex {
	return exactComplex{re: exactAddRat(a.re, b.re), im: exactAddRat(a.im, b.im)}
}

func exactSub(a, b exactComplex) exactComplex {
	return exactComplex{re: exactSubRat(a.re, b.re), im: exactSubRat(a.im, b.im)}
}

func exactMul(a, b exactComplex) exactComplex {
	realPart := exactSubRat(exactMulRat(a.re, b.re), exactMulRat(a.im, b.im))
	imagPart := exactAddRat(exactMulRat(a.re, b.im), exactMulRat(a.im, b.re))
	return exactComplex{re: realPart, im: imagPart}
}

func exactEqual(a, b exactComplex) bool { return a.re.Cmp(b.re) == 0 && a.im.Cmp(b.im) == 0 }
func exactIsZero(a exactComplex) bool   { return a.re.Sign() == 0 && a.im.Sign() == 0 }
func exactL1(a exactComplex) *big.Rat   { return exactAddRat(exactAbsRat(a.re), exactAbsRat(a.im)) }

func exactToComplex128(a exactComplex) complex128 {
	r, _ := a.re.Float64()
	i, _ := a.im.Float64()
	return complex(r, i)
}

func roundoffUnit() *big.Rat {
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 53))
}

func roundoffGamma2() *big.Rat {
	u := roundoffUnit()
	twoU := exactMulRat(big.NewRat(2, 1), u)
	return new(big.Rat).Quo(twoU, exactSubRat(big.NewRat(1, 1), twoU))
}

func exactAddError(a, b exactComplex) *big.Rat {
	return exactMulRat(roundoffUnit(), exactAddRat(exactL1(a), exactL1(b)))
}

func exactComplexMulError(a, b exactComplex) *big.Rat {
	return exactMulRat(roundoffGamma2(), exactMulRat(exactL1(a), exactL1(b)))
}

func updateMaxRat(dst map[string]*big.Rat, name string, value *big.Rat) {
	if current, ok := dst[name]; !ok || value.Cmp(current) > 0 {
		dst[name] = new(big.Rat).Set(value)
	}
}

func ratMapUp(values map[string]*big.Rat) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = rationalUp(value)
	}
	return result
}

func rationalUp(value *big.Rat) string {
	f := new(big.Float).SetPrec(256).SetMode(big.ToPositiveInf)
	f.SetRat(value)
	return f.Text('g', 20)
}

func compareRoundoffFailureMasks(pairs []fixedSigmaRoundoffPair) (fixedSigmaRoundoffFailureMaskSummary, error) {
	const slots = 1 << 12
	if len(pairs) == 0 {
		return fixedSigmaRoundoffFailureMaskSummary{}, errors.New("no failure masks to compare")
	}
	passMasks := make([][]bool, len(pairs))
	failureMasks := make([][]bool, len(pairs))
	for pairIndex, pair := range pairs {
		if pair.OldGamma64FailureCount+len(pair.OldGamma64PassIndices) != slots {
			return fixedSigmaRoundoffFailureMaskSummary{}, fmt.Errorf("pair %d failure and pass counts do not cover %d slots", pairIndex, slots)
		}
		passMasks[pairIndex] = make([]bool, slots)
		failureMasks[pairIndex] = make([]bool, slots)
		for _, index := range pair.OldGamma64PassIndices {
			if index < 0 || index >= slots || passMasks[pairIndex][index] {
				return fixedSigmaRoundoffFailureMaskSummary{}, fmt.Errorf("pair %d has invalid or duplicate pass slot %d", pairIndex, index)
			}
			passMasks[pairIndex][index] = true
		}
		for index := 0; index < slots; index++ {
			failureMasks[pairIndex][index] = !passMasks[pairIndex][index]
		}
	}
	identical, samePass := true, true
	intersection, union := 0, 0
	different := make([]int, 0)
	for index := 0; index < slots; index++ {
		allFail, anyFail := true, false
		for pairIndex := range pairs {
			failed := failureMasks[pairIndex][index]
			allFail = allFail && failed
			anyFail = anyFail || failed
			if failed != failureMasks[0][index] {
				identical = false
			}
			if passMasks[pairIndex][index] != passMasks[0][index] {
				samePass = false
			}
		}
		if allFail {
			intersection++
		}
		if anyFail {
			union++
		}
		if anyFail && !allFail && len(different) < 32 {
			different = append(different, index)
		}
	}
	return fixedSigmaRoundoffFailureMaskSummary{
		AllOldFailureMasksIdentical: identical, IntersectionFailureSlots: intersection,
		UnionFailureSlots: union, DifferentIndicesFirst: different, AllPairsShareSamePassSlots: samePass,
	}, nil
}

func complexHasSubnormalComponent(value complex128) bool {
	const smallestNormal = 0x1p-1022
	for _, component := range []float64{real(value), imag(value)} {
		if component != 0 && math.Abs(component) < smallestNormal {
			return true
		}
	}
	return false
}
