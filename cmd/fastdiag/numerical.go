package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/bits"
	"math/cmplx"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
)

const (
	fastStandardInputSHA256 = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
	numericalThresholdValue = 1e-2
	precisionErrorFloor     = 1e-30
)

type decodedNumericalRun struct {
	output     NumericalOutput
	values     []complex128
	preValues  []complex128
	ciphertext *rlwe.Ciphertext
}

func numericalReference(secondaryRoot string, primary RepositoryMetadata, opts options) (NumericalDocument, error) {
	branch, err := gitOutput(secondaryRoot, "branch", "--show-current")
	if err != nil {
		return NumericalDocument{}, err
	}
	if branch != "fast-qprefix" {
		return NumericalDocument{}, fmt.Errorf("numerical mode 要求 authoritative Secondary branch fast-qprefix，目前是 %q", branch)
	}
	if err := requireCleanWorktree(secondaryRoot); err != nil {
		return NumericalDocument{}, err
	}
	secondary, err := repositoryMetadata(secondaryRoot, "")
	if err != nil {
		return NumericalDocument{}, err
	}
	if secondary.Dirty {
		return NumericalDocument{}, fmt.Errorf("Secondary 在 provenance 擷取後變為 dirty，拒絕執行 numerical comparison")
	}

	params, residual, err := fastStandardP93Parameters()
	if err != nil {
		return NumericalDocument{}, err
	}
	values := fastStandardP93Values()
	inputHash := fastStandardInputFingerprint(values)
	if inputHash != fastStandardInputSHA256 {
		return NumericalDocument{}, fmt.Errorf("canonical P93 input fingerprint 不符：%s", inputHash)
	}
	if len(values) != 4096 || params.BootstrappingParameters.LogN() != 13 || params.CoeffsToSlotsParameters.LogSlots != 12 || bits.Len64(params.BootstrappingParameters.Q()[0]) != 55 {
		return NumericalDocument{}, fmt.Errorf("canonical P93 parameters 不符：slots=%d LogN=%d LogSlots=%d q0bits=%d", len(values), params.BootstrappingParameters.LogN(), params.CoeffsToSlotsParameters.LogSlots, bits.Len64(params.BootstrappingParameters.Q()[0]))
	}
	input, err := fastStandardEncodedInput(residual, params.CoeffsToSlotsParameters.LogSlots, values)
	if err != nil {
		return NumericalDocument{}, err
	}
	preFastDecoded, err := fastStandardDecodeFast(residual, input)
	if err != nil {
		return NumericalDocument{}, fmt.Errorf("decode Fast pre-Bootstrap input: %w", err)
	}

	fastEval, err := bootstrapping.NewFastEvaluator(params)
	if err != nil {
		return NumericalDocument{}, fmt.Errorf("建立 Fast evaluator：%w", err)
	}
	fastRuns := make([]decodedNumericalRun, 0, 2)
	for i := 0; i < 2; i++ {
		out, err := fastEval.Bootstrap(input.CopyNew())
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("Fast Bootstrap execution %d：%w", i+1, err)
		}
		decoded, err := fastStandardDecodeFast(residual, out)
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("decode Fast execution %d：%w", i+1, err)
		}
		record, err := numericalOutput(i+1, residual, values, decoded, out, input, preFastDecoded)
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("Fast execution %d metrics：%w", i+1, err)
		}
		fastRuns = append(fastRuns, decodedNumericalRun{output: record, values: decoded, preValues: preFastDecoded, ciphertext: out})
	}

	standardRuns := make([]decodedNumericalRun, 0, opts.standardTrials)
	standardKeys := make([]*rlwe.SecretKey, 0, opts.standardTrials)
	keyTrialEvidence := make([]NumericalKeyTrial, 0, opts.standardTrials)
	var standardEvalForLockstep *bootstrapping.Evaluator
	var standardSecretForLockstep *rlwe.SecretKey
	for i := 0; i < opts.standardTrials; i++ {
		keygen := rlwe.NewKeyGenerator(params.BootstrappingParameters)
		sk := keygen.GenSecretKeyNew()
		if sk.LevelP() < 0 {
			return NumericalDocument{}, fmt.Errorf("Standard trial %d secret key has no P basis; this could select the Fast-compatible key path", i+1)
		}
		unique := true
		for _, priorKey := range standardKeys {
			if sk.Equal(priorKey) {
				unique = false
				break
			}
		}
		if !unique {
			return NumericalDocument{}, fmt.Errorf("Standard trial %d generated a secret key identical to an earlier trial", i+1)
		}
		standardKeys = append(standardKeys, sk)
		keys, _, err := params.GenEvaluationKeys(sk)
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("Standard trial %d evaluation keys：%w", i+1, err)
		}
		standardEval, err := bootstrapping.NewEvaluator(params, keys)
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("建立 Standard evaluator for trial %d：%w", i+1, err)
		}
		if i == 0 {
			standardEvalForLockstep, standardSecretForLockstep = standardEval, sk
		}
		prePlain := rlwe.NewDecryptor(residual, sk).DecryptNew(input)
		preStandardDecoded := make([]complex128, residual.MaxSlots())
		if err := ckks.NewEncoder(residual).Decode(prePlain, preStandardDecoded); err != nil {
			return NumericalDocument{}, fmt.Errorf("decode Standard pre-Bootstrap trial %d: %w", i+1, err)
		}
		if err := validateNumericalVector(preStandardDecoded); err != nil {
			return NumericalDocument{}, fmt.Errorf("Standard pre-Bootstrap trial %d decoded vector: %w", i+1, err)
		}
		out, err := standardEval.Bootstrap(input.CopyNew())
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("Standard Bootstrap trial %d：%w", i+1, err)
		}
		decryptor := rlwe.NewDecryptor(residual, sk)
		plain := decryptor.DecryptNew(out)
		decoded := make([]complex128, residual.MaxSlots())
		if err := ckks.NewEncoder(residual).Decode(plain, decoded); err != nil {
			return NumericalDocument{}, fmt.Errorf("decode Standard trial %d：%w", i+1, err)
		}
		if err := validateNumericalVector(decoded); err != nil {
			return NumericalDocument{}, fmt.Errorf("Standard trial %d decoded vector：%w", i+1, err)
		}
		record, err := numericalOutput(i+1, residual, values, decoded, out, input, preStandardDecoded)
		if err != nil {
			return NumericalDocument{}, fmt.Errorf("Standard trial %d metrics：%w", i+1, err)
		}
		standardRuns = append(standardRuns, decodedNumericalRun{output: record, values: decoded, preValues: preStandardDecoded, ciphertext: out})
		keyTrialEvidence = append(keyTrialEvidence, NumericalKeyTrial{
			Index: i + 1, FreshKeyGenerator: true, FreshEvaluationKeys: true,
			SecretKeyLevelP: sk.LevelP(), UniqueSecretKey: unique,
			EvaluatorPath: "Standard GenEvaluationKeys with P-capable secret key",
		})
	}
	if standardEvalForLockstep == nil || standardSecretForLockstep == nil {
		return NumericalDocument{}, fmt.Errorf("Standard lockstep evaluator/key was not retained from trial 1")
	}
	stageLockstep, err := runNumericalStageLockstep(
		fastEval, standardEvalForLockstep, standardSecretForLockstep,
		params, input, values,
		fastRuns[0].preValues, standardRuns[0].preValues,
		fastRuns[0].values, standardRuns[0].values,
		fastRuns[0].ciphertext, standardRuns[0].ciphertext,
	)
	if err != nil {
		return NumericalDocument{}, fmt.Errorf("Fast/Standard stage lockstep: %w", err)
	}

	fastRecords := make([]NumericalOutput, len(fastRuns))
	fastThresholds := make([]NumericalThresholdAudit, len(fastRuns))
	for i := range fastRuns {
		fastRecords[i], fastThresholds[i] = fastRuns[i].output, fastRuns[i].output.Threshold
	}
	standardRecords := make([]NumericalOutput, len(standardRuns))
	standardVectors := make([][]complex128, len(standardRuns))
	standardThresholds := make([]NumericalThresholdAudit, len(standardRuns))
	for i := range standardRuns {
		standardRecords[i], standardVectors[i], standardThresholds[i] = standardRuns[i].output, standardRuns[i].values, standardRuns[i].output.Threshold
	}
	analysisFastRuns, analysisFastRecords, analysisFastThresholds := fastRuns, fastRecords, fastThresholds
	determinism := numericalFastDeterminism(fastRuns[0].values, fastRuns[1].values)
	if determinism.BitwiseIdentical {
		analysisFastRuns = fastRuns[:1]
		analysisFastRecords = fastRecords[:1]
		analysisFastThresholds = fastThresholds[:1]
	}

	pairwise := make([]NumericalPairwiseComparison, 0, len(analysisFastRuns)*len(standardRuns))
	pairwiseThresholds := make([]NumericalThresholdAudit, 0, cap(pairwise))
	for fastIndex, fastRun := range analysisFastRuns {
		for standardIndex, standardRun := range standardRuns {
			metrics, threshold := numericalComparison(fastRun.values, standardRun.values)
			metadata := compareNumericalMetadata(fastRun.output, standardRun.output)
			pairwise = append(pairwise, NumericalPairwiseComparison{
				FastTrialIndex: fastIndex + 1, StandardTrialIndex: standardIndex + 1,
				Metrics: metrics, WorstExamples: pairwiseWorstExamples(values, fastRun.values, standardRun.values, metrics),
				Threshold: threshold, Metadata: metadata,
			})
			pairwiseThresholds = append(pairwiseThresholds, threshold)
		}
	}

	spread := numericalStandardSpread(standardVectors)
	fastAggregate := aggregateNumericalOutputs(analysisFastRecords)
	standardAggregate := aggregateNumericalOutputs(standardRecords)
	fastOriginalThreshold := aggregateNumericalThresholds(analysisFastThresholds)
	standardOriginalThreshold := aggregateNumericalThresholds(standardThresholds)
	fastStandardThreshold := aggregateNumericalThresholds(pairwiseThresholds)
	checks, classification := classifyFastStandard(
		fastAggregate, standardAggregate, fastOriginalThreshold, fastStandardThreshold,
		analysisFastRecords, standardRecords, pairwise,
		maxMagnitude(values),
	)

	return NumericalDocument{
		SchemaVersion: "fastdiag.numerical.v1", Timestamp: time.Now().UTC(), Profile: opts.profile,
		Primary: primary, Secondary: secondary,
		Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
			CPU: cpuModel(), NumCPU: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0)},
		Workload: NumericalWorkload{
			FingerprintSHA256: inputHash, InputSlots: len(values), LogN: params.BootstrappingParameters.LogN(),
			LogSlots:   params.CoeffsToSlotsParameters.LogSlots,
			QChainBits: numericalPrimeBits(params.BootstrappingParameters.Q()), PBits: numericalPrimeBits(params.BootstrappingParameters.P()),
			Q0Bits: bits.Len64(params.BootstrappingParameters.Q()[0]), PolynomialDegree: params.Mod1ParametersLiteral.Mod1Degree,
			DoubleAngle: params.Mod1ParametersLiteral.DoubleAngle, EvalModLogScale: params.Mod1ParametersLiteral.LogScale,
			LogMessageRatio: params.Mod1ParametersLiteral.LogMessageRatio,
		},
		Execution: map[string]int{"fast_executions": len(fastRuns), "fast_effective_numerical_trials": len(analysisFastRuns), "standard_key_trials": len(standardRuns), "slots_per_trial": len(values)},
		Threshold: numericalThresholdValue, PrecisionErrorFloor: precisionErrorFloor,
		StandardRMSENearZeroFloor: standardRMSENearZeroFloor(maxMagnitude(values)),
		FastDeterminism:           determinism, FastTrials: fastRecords, StandardTrials: standardRecords,
		StandardKeyTrials: keyTrialEvidence,
		FastAggregate:     fastAggregate, StandardAggregate: standardAggregate,
		FastVsOriginalThreshold: fastOriginalThreshold, StandardVsOriginalThreshold: standardOriginalThreshold,
		FastVsStandard: pairwise, FastVsStandardThreshold: fastStandardThreshold,
		StandardToStandardVariability: spread, StageLockstep: &stageLockstep,
		Classification: classification, ClassificationChecks: checks,
		Limitations: []string{
			"Fast output is decoded directly from c0 under the current Fast zero-secret mode; each Standard output is decrypted with its trial secret key.",
			"The diagnostic uses the canonical deterministic plaintext-like input c0=encoded message, c1=0; it does not characterize encrypted-input noise or security.",
			"This is a numerical correctness diagnostic, not a timing benchmark or an independent scientific acceptance decision.",
		},
	}, nil
}

func fastStandardP93Parameters() (bootstrapping.Parameters, ckks.Parameters, error) {
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: 13, LogQ: []int{55, 39}, LogDefaultScale: 45, Xs: ring.Ternary{H: 192},
	})
	if err != nil {
		return bootstrapping.Parameters{}, ckks.Parameters{}, err
	}
	logN, logSlots, zero, evalModScale, degree, doubleAngle := 13, 12, 0, 60, 30, 3
	k, logMessageRatio := 16, 10
	params, err := bootstrapping.NewParametersFromLiteral(residual, bootstrapping.ParametersLiteral{
		LogN: &logN, LogSlots: &logSlots, LogP: []int{61, 61, 61, 61, 61},
		SlotsToCoeffsFactorizationDepthAndLogScales: [][]int{{39}, {39}, {39}},
		CoeffsToSlotsFactorizationDepthAndLogScales: [][]int{{56}, {56}, {56}, {56}},
		EvalModLogScale: &evalModScale, Mod1Degree: &degree, DoubleAngle: &doubleAngle,
		K: &k, LogMessageRatio: &logMessageRatio, Mod1InvDegree: &zero,
		EphemeralSecretWeight: &zero,
	})
	if err != nil {
		return bootstrapping.Parameters{}, ckks.Parameters{}, err
	}
	params.CircuitOrder = bootstrapping.ModUpThenEncode
	params.ResidualParameters = residual
	return params, residual, nil
}

func fastStandardP93Values() []complex128 {
	values := make([]complex128, 1<<12)
	for j := range values {
		values[j] = complex(float64(j%17-8)/256, float64((3*j)%13-6)/512)
	}
	return values
}

func fastStandardEncodedInput(params ckks.Parameters, logSlots int, values []complex128) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, 0)
	pt.IsNTT, pt.IsMontgomery = true, false
	pt.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err := ckks.NewEncoder(params).Encode(values, pt); err != nil {
		return nil, err
	}
	ct := ckks.NewCiphertext(params, 1, 0)
	*ct.MetaData = *pt.MetaData
	ct.Value[0].Copy(pt.Value)
	ct.Value[1].Zero()
	ct.IsNTT, ct.IsMontgomery = pt.IsNTT, pt.IsMontgomery
	return ct, nil
}

func fastStandardDecodeFast(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	if ct == nil || ct.Degree() < 0 || ct.Level() < 0 || ct.Level() > params.MaxLevel() {
		return nil, fmt.Errorf("Fast output has invalid ciphertext metadata")
	}
	pt := ckks.NewPlaintext(params, ct.Level())
	*pt.MetaData = *ct.MetaData
	pt.Value.Copy(ct.Value[0])
	pt.IsNTT, pt.IsMontgomery = ct.IsNTT, ct.IsMontgomery
	decoded := make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(pt, decoded); err != nil {
		return nil, err
	}
	return decoded, validateNumericalVector(decoded)
}

func numericalOutput(index int, residual ckks.Parameters, original, decoded []complex128, output, input *rlwe.Ciphertext, preBootstrapDecoded []complex128) (NumericalOutput, error) {
	if err := validateNumericalVector(decoded); err != nil {
		return NumericalOutput{}, err
	}
	if err := validateNumericalVector(preBootstrapDecoded); err != nil {
		return NumericalOutput{}, fmt.Errorf("pre-Bootstrap decoded values: %w", err)
	}
	metrics, threshold := numericalComparison(original, decoded)
	preAgainstOriginal, _ := numericalComparison(original, preBootstrapDecoded)
	precision := numericalPrecisionBits(original, decoded)
	helper := ckks.GetPrecisionStats(residual, ckks.NewEncoder(residual), nil, original, decoded, 0, false)
	metadata := numericalMetadata(output)
	contract := NumericalPublicContract{
		LevelMatchesResidualMax:  output.Level() == residual.MaxLevel(),
		ScaleMatchesDefault:      output.Scale.Equal(residual.DefaultScale()),
		DimensionsMatchCanonical: output.LogDimensions == input.LogDimensions,
		Status:                   "PASS",
	}
	if !contract.LevelMatchesResidualMax || !contract.ScaleMatchesDefault || !contract.DimensionsMatchCanonical {
		contract.Status = "SEMANTIC_MISMATCH"
	}
	return NumericalOutput{
		Index: index, Metadata: metadata, PublicContract: contract, VsOriginal: metrics,
		Threshold: threshold, PrecisionBits: precision, CKKSHelper: numericalCKKSHelper(helper),
		BootstrapSNR: numericalmetrics.Compare(preBootstrapDecoded, decoded),
		PreBootstrapVsCanonicalOriginalComplexRMSE: preAgainstOriginal.Complex.RMSE,
	}, nil
}

func numericalMetadata(ct *rlwe.Ciphertext) NumericalMetadata {
	scale := ct.Scale.Float64()
	scaleFinite := !math.IsNaN(scale) && !math.IsInf(scale, 0)
	if !scaleFinite {
		scale = 0
	}
	return NumericalMetadata{
		Level: ct.Level(), Degree: ct.Degree(), Scale: scale, ScaleExact: ct.Scale.Value.Text('e', 80), ScaleFloat64Finite: scaleFinite, ScaleLog2: ct.Scale.Log2(),
		IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		LogDimensions: NumericalDimensions{Rows: ct.LogDimensions.Rows, Cols: ct.LogDimensions.Cols},
	}
}

func numericalCKKSHelper(stats ckks.PrecisionStats) NumericalCKKSPrecisionHelper {
	return NumericalCKKSPrecisionHelper{
		Minimum:   NumericalPrecisionVector{Real: stats.MINLog2Prec.Real, Imag: stats.MINLog2Prec.Imag, L2: stats.MINLog2Prec.L2},
		Maximum:   NumericalPrecisionVector{Real: stats.MAXLog2Prec.Real, Imag: stats.MAXLog2Prec.Imag, L2: stats.MAXLog2Prec.L2},
		Average:   NumericalPrecisionVector{Real: stats.AVGLog2Prec.Real, Imag: stats.AVGLog2Prec.Imag, L2: stats.AVGLog2Prec.L2},
		Median:    NumericalPrecisionVector{Real: stats.MEDLog2Prec.Real, Imag: stats.MEDLog2Prec.Imag, L2: stats.MEDLog2Prec.L2},
		StdDev:    NumericalPrecisionVector{Real: stats.STDLog2Prec.Real, Imag: stats.STDLog2Prec.Imag, L2: stats.STDLog2Prec.L2},
		Log2Scale: stats.Log2Scale,
	}
}

func numericalComparison(want, have []complex128) (NumericalMetricSet, NumericalThresholdAudit) {
	realErrors := make([]float64, len(want))
	imagErrors := make([]float64, len(want))
	complexErrors := make([]float64, len(want))
	for i := range want {
		difference := have[i] - want[i]
		realErrors[i] = math.Abs(real(difference))
		imagErrors[i] = math.Abs(imag(difference))
		complexErrors[i] = cmplx.Abs(difference)
	}
	return NumericalMetricSet{
		Real:    numericalErrorMetric(realErrors, want, have),
		Imag:    numericalErrorMetric(imagErrors, want, have),
		Complex: numericalErrorMetric(complexErrors, want, have),
	}, numericalThresholdAudit(realErrors, imagErrors, numericalThresholdValue)
}

func numericalErrorMetric(errors []float64, original, output []complex128) NumericalErrorMetric {
	ordered := append([]float64(nil), errors...)
	sort.Float64s(ordered)
	worst := 0
	for i := 1; i < len(errors); i++ {
		if errors[i] > errors[worst] {
			worst = i
		}
	}
	var sum, sumSquares float64
	for _, value := range errors {
		sum += value
		sumSquares += value * value
	}
	difference := output[worst] - original[worst]
	return NumericalErrorMetric{
		MeanAbsoluteError: sum / float64(len(errors)), RMSE: math.Sqrt(sumSquares / float64(len(errors))),
		P50: numericalQuantile(ordered, 0.50), P95: numericalQuantile(ordered, 0.95),
		P99: numericalQuantile(ordered, 0.99), Max: ordered[len(ordered)-1], WorstSlot: worst,
		WorstReference: numericalComplex(original[worst]), WorstOutput: numericalComplex(output[worst]),
		WorstDifference: numericalComplex(difference),
	}
}

func pairwiseWorstExamples(original, fast, standard []complex128, metrics NumericalMetricSet) NumericalPairwiseWorstExamples {
	return NumericalPairwiseWorstExamples{
		Real:    pairwiseWorstExample(metrics.Real.WorstSlot, original, fast, standard),
		Imag:    pairwiseWorstExample(metrics.Imag.WorstSlot, original, fast, standard),
		Complex: pairwiseWorstExample(metrics.Complex.WorstSlot, original, fast, standard),
	}
}

func pairwiseWorstExample(slot int, original, fast, standard []complex128) NumericalPairwiseWorstExample {
	return NumericalPairwiseWorstExample{
		Slot: slot, Original: numericalComplex(original[slot]), Fast: numericalComplex(fast[slot]),
		Standard: numericalComplex(standard[slot]), Difference: numericalComplex(fast[slot] - standard[slot]),
	}
}

func numericalQuantile(sortedValues []float64, probability float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	position := probability * float64(len(sortedValues)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return sortedValues[lower]
	}
	weight := position - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func numericalThresholdAudit(realErrors, imagErrors []float64, threshold float64) NumericalThresholdAudit {
	result := NumericalThresholdAudit{Threshold: threshold, Comparisons: 1}
	result.RealCoordinates, result.ImagCoordinates = len(realErrors), len(imagErrors)
	for _, difference := range realErrors {
		if difference > threshold {
			result.RealExceeding++
		}
	}
	for _, difference := range imagErrors {
		if difference > threshold {
			result.ImagExceeding++
		}
	}
	result.TotalExceeding = result.RealExceeding + result.ImagExceeding
	result.TotalCoordinates = result.RealCoordinates + result.ImagCoordinates
	if result.RealCoordinates > 0 {
		result.RealFraction = float64(result.RealExceeding) / float64(result.RealCoordinates)
	}
	if result.ImagCoordinates > 0 {
		result.ImagFraction = float64(result.ImagExceeding) / float64(result.ImagCoordinates)
	}
	if result.TotalCoordinates > 0 {
		result.TotalFraction = float64(result.TotalExceeding) / float64(result.TotalCoordinates)
	}
	return result
}

func aggregateNumericalThresholds(audits []NumericalThresholdAudit) NumericalThresholdAudit {
	result := NumericalThresholdAudit{Threshold: numericalThresholdValue, Comparisons: len(audits)}
	for _, audit := range audits {
		result.RealExceeding += audit.RealExceeding
		result.RealCoordinates += audit.RealCoordinates
		result.ImagExceeding += audit.ImagExceeding
		result.ImagCoordinates += audit.ImagCoordinates
	}
	result.TotalExceeding = result.RealExceeding + result.ImagExceeding
	result.TotalCoordinates = result.RealCoordinates + result.ImagCoordinates
	if result.RealCoordinates > 0 {
		result.RealFraction = float64(result.RealExceeding) / float64(result.RealCoordinates)
	}
	if result.ImagCoordinates > 0 {
		result.ImagFraction = float64(result.ImagExceeding) / float64(result.ImagCoordinates)
	}
	if result.TotalCoordinates > 0 {
		result.TotalFraction = float64(result.TotalExceeding) / float64(result.TotalCoordinates)
	}
	return result
}

func numericalPrecisionBits(want, have []complex128) NumericalPrecisionBits {
	values := make([]float64, len(want))
	precision := make([]float64, len(want))
	var sum float64
	minimum := math.Inf(1)
	minimumSlot := 0
	for i := range want {
		values[i] = cmplx.Abs(have[i] - want[i])
		precision[i] = -math.Log2(math.Max(values[i], precisionErrorFloor))
		sum += precision[i]
		if precision[i] < minimum {
			minimum, minimumSlot = precision[i], i
		}
	}
	ordered := append([]float64(nil), precision...)
	sort.Float64s(ordered)
	return NumericalPrecisionBits{
		ErrorFloor: precisionErrorFloor, MeanBits: sum / float64(len(precision)),
		MedianBits: numericalQuantile(ordered, 0.50), P05Bits: numericalQuantile(ordered, 0.05),
		MinimumBits: minimum, MinimumSlot: minimumSlot, MinimumSlotError: values[minimumSlot],
	}
}

func compareNumericalMetadata(fast, standard NumericalOutput) NumericalMetadataComparison {
	differences := make([]string, 0, 5)
	contractOK := fast.PublicContract.Status == "PASS" && standard.PublicContract.Status == "PASS"
	semanticMismatch := false
	if fast.Metadata.Level != standard.Metadata.Level {
		differences = append(differences, "level differs")
		semanticMismatch = true
	}
	if !almostEqualFloat(fast.Metadata.ScaleLog2, standard.Metadata.ScaleLog2, 1e-12) {
		differences = append(differences, "scale differs")
		semanticMismatch = true
	}
	if fast.Metadata.LogDimensions != standard.Metadata.LogDimensions {
		differences = append(differences, "LogDimensions differs")
		semanticMismatch = true
	}
	if fast.Metadata.Degree != standard.Metadata.Degree {
		differences = append(differences, "degree differs (backend/key-path representation)")
	}
	if fast.Metadata.IsNTT != standard.Metadata.IsNTT {
		differences = append(differences, "IsNTT differs (representation)")
	}
	if fast.Metadata.IsMontgomery != standard.Metadata.IsMontgomery {
		differences = append(differences, "IsMontgomery differs (representation)")
	}
	classification := "MATCH"
	if semanticMismatch || !contractOK {
		classification = "SEMANTIC_MISMATCH"
	} else if len(differences) != 0 {
		classification = "EXPECTED_REPRESENTATION_OR_KEY_PATH_DIFFERENCE"
	}
	return NumericalMetadataComparison{Classification: classification, Differences: differences, PublicContractOK: contractOK}
}

func numericalFastDeterminism(first, second []complex128) NumericalFastDeterminism {
	result := NumericalFastDeterminism{ComparedRuns: 2, BitwiseIdentical: len(first) == len(second)}
	for i := range first {
		if math.Float64bits(real(first[i])) != math.Float64bits(real(second[i])) || math.Float64bits(imag(first[i])) != math.Float64bits(imag(second[i])) {
			result.BitwiseIdentical = false
		}
		difference := cmplx.Abs(first[i] - second[i])
		if difference > result.MaxComplexDiff {
			result.MaxComplexDiff, result.MaxDiffSlot = difference, i
		}
	}
	return result
}

func numericalStandardSpread(trials [][]complex128) NumericalStandardSpread {
	result := NumericalStandardSpread{TrialCount: len(trials)}
	if len(trials) == 0 || len(trials[0]) == 0 {
		return result
	}
	slots := len(trials[0])
	means := make([]complex128, slots)
	for trialIndex, trial := range trials {
		for slot := range means {
			means[slot] += (trial[slot] - means[slot]) / complex(float64(trialIndex+1), 0)
		}
	}
	result.PerSlotMeanFingerprintSHA256 = numericalVectorFingerprint(means)
	var realSquares, imagSquares, complexSquares float64
	maxSpread := -1.0
	for trialIndex, trial := range trials {
		for slot, value := range trial {
			difference := value - means[slot]
			r, im, magnitude := real(difference), imag(difference), cmplx.Abs(difference)
			realSquares += r * r
			imagSquares += im * im
			complexSquares += magnitude * magnitude
			if magnitude > maxSpread {
				maxSpread, result.MaxSpreadSlot, result.MaxSpreadTrialIndex = magnitude, slot, trialIndex+1
				result.MaxSpreadMean, result.MaxSpreadOutput = numericalComplex(means[slot]), numericalComplex(value)
			}
		}
	}
	denominator := float64(len(trials) * slots)
	result.RMSRealSpread = math.Sqrt(realSquares / denominator)
	result.RMSImagSpread = math.Sqrt(imagSquares / denominator)
	result.RMSComplexSpread = math.Sqrt(complexSquares / denominator)
	result.MaxComplexSpread = maxSpread

	maxPair := -1.0
	for a := 0; a < len(trials); a++ {
		for b := a + 1; b < len(trials); b++ {
			for slot := 0; slot < slots; slot++ {
				difference := cmplx.Abs(trials[a][slot] - trials[b][slot])
				if difference > maxPair {
					maxPair = difference
					result.MaxPairwiseComplexDiff, result.MaxPairwiseSlot = difference, slot
					result.MaxPairwiseTrialA, result.MaxPairwiseTrialB = a+1, b+1
					result.MaxPairwiseValueA = numericalComplex(trials[a][slot])
					result.MaxPairwiseValueB = numericalComplex(trials[b][slot])
				}
			}
		}
	}
	selectedSlots := map[int]bool{0: true, slots / 2: true, result.MaxSpreadSlot: true}
	for slot := 0; slot < slots; slot++ {
		if selectedSlots[slot] {
			result.PerSlotMeanSamples = append(result.PerSlotMeanSamples, NumericalMeanSlotSample{Slot: slot, Mean: numericalComplex(means[slot])})
		}
	}
	return result
}

func aggregateNumericalOutputs(outputs []NumericalOutput) NumericalAggregate {
	result := NumericalAggregate{TrialCount: len(outputs)}
	result.Real = aggregateNumericalMetric(outputs, func(output NumericalOutput) NumericalErrorMetric { return output.VsOriginal.Real })
	result.Imag = aggregateNumericalMetric(outputs, func(output NumericalOutput) NumericalErrorMetric { return output.VsOriginal.Imag })
	result.Complex = aggregateNumericalMetric(outputs, func(output NumericalOutput) NumericalErrorMetric { return output.VsOriginal.Complex })
	medianBits, meanBits := make([]float64, len(outputs)), make([]float64, len(outputs))
	for i, output := range outputs {
		medianBits[i], meanBits[i] = output.PrecisionBits.MedianBits, output.PrecisionBits.MeanBits
	}
	result.MedianOfTrialMedianBits = numericalMedian(medianBits)
	result.MedianOfTrialMeanBits = numericalMedian(meanBits)
	return result
}

func aggregateNumericalMetric(outputs []NumericalOutput, get func(NumericalOutput) NumericalErrorMetric) NumericalMetricAggregate {
	mae, rmse := make([]float64, len(outputs)), make([]float64, len(outputs))
	p50, p95, p99, maxValues := make([]float64, len(outputs)), make([]float64, len(outputs)), make([]float64, len(outputs)), make([]float64, len(outputs))
	for i, output := range outputs {
		metric := get(output)
		mae[i], rmse[i] = metric.MeanAbsoluteError, metric.RMSE
		p50[i], p95[i], p99[i], maxValues[i] = metric.P50, metric.P95, metric.P99, metric.Max
	}
	return NumericalMetricAggregate{
		MeanAbsoluteError: numericalMedian(mae), RMSE: numericalMedian(rmse),
		P50: numericalMedian(p50), P95: numericalMedian(p95), P99: numericalMedian(p99), Max: numericalMedian(maxValues),
	}
}

func classifyFastStandard(fast, standard NumericalAggregate, fastOriginalThreshold, fastStandardThreshold NumericalThresholdAudit, fastTrials, standardTrials []NumericalOutput, pairwise []NumericalPairwiseComparison, maxOriginalMagnitude float64) (NumericalClassificationChecks, string) {
	metadataOK := true
	for _, output := range fastTrials {
		metadataOK = metadataOK && output.PublicContract.Status == "PASS"
	}
	for _, output := range standardTrials {
		metadataOK = metadataOK && output.PublicContract.Status == "PASS"
	}
	for _, comparison := range pairwise {
		metadataOK = metadataOK && comparison.Metadata.Classification != "SEMANTIC_MISMATCH"
	}
	precisionDrop := standard.MedianOfTrialMedianBits - fast.MedianOfTrialMedianBits
	standardRMSEFloor := standardRMSENearZeroFloor(maxOriginalMagnitude)
	standardNearZero := standard.Complex.RMSE <= standardRMSEFloor
	checks := NumericalClassificationChecks{
		FastVsOriginalNoCoordinateExceeds: fastOriginalThreshold.TotalExceeding == 0,
		FastVsStandardNoCoordinateExceeds: fastStandardThreshold.TotalExceeding == 0,
		MedianPrecisionDropWithinTwoBits:  precisionDrop <= 2,
		FastRMSEWithinFourXOrNearZero:     standardNearZero || fast.Complex.RMSE <= 4*standard.Complex.RMSE,
		PublicMetadataContractRespected:   metadataOK, StandardRMSENearZero: standardNearZero,
		FastToStandardMedianPrecisionDrop: precisionDrop,
	}
	if !standardNearZero && standard.Complex.RMSE > 0 {
		ratio := fast.Complex.RMSE / standard.Complex.RMSE
		checks.FastToStandardComplexRMSERatio = &ratio
	}
	classification := "FAST_NUMERICAL_QUALITY_DEGRADED"
	if checks.FastVsOriginalNoCoordinateExceeds && checks.FastVsStandardNoCoordinateExceeds &&
		checks.MedianPrecisionDropWithinTwoBits && checks.FastRMSEWithinFourXOrNearZero && checks.PublicMetadataContractRespected {
		classification = "FAST_STANDARD_NUMERICAL_CLOSE"
	}
	return checks, classification
}

func standardRMSENearZeroFloor(maxOriginalMagnitude float64) float64 {
	return 64 * (math.Nextafter(1, 2) - 1) * math.Max(1, maxOriginalMagnitude)
}

func fastStandardInputFingerprint(values []complex128) string {
	hash := sha256.New()
	var encoded [16]byte
	for _, value := range values {
		binary.LittleEndian.PutUint64(encoded[:8], math.Float64bits(real(value)))
		binary.LittleEndian.PutUint64(encoded[8:], math.Float64bits(imag(value)))
		_, _ = hash.Write(encoded[:])
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func numericalVectorFingerprint(values []complex128) string {
	return fastStandardInputFingerprint(values)
}

func numericalPrimeBits(primes []uint64) []int {
	result := make([]int, len(primes))
	for i, prime := range primes {
		result[i] = bits.Len64(prime)
	}
	return result
}

func maxMagnitude(values []complex128) float64 {
	maximum := 0.0
	for _, value := range values {
		maximum = math.Max(maximum, cmplx.Abs(value))
	}
	return maximum
}

func validateNumericalVector(values []complex128) error {
	if len(values) != 4096 {
		return fmt.Errorf("decoded slot count is %d, want 4096", len(values))
	}
	for i, value := range values {
		if math.IsNaN(real(value)) || math.IsNaN(imag(value)) || math.IsInf(real(value), 0) || math.IsInf(imag(value), 0) {
			return fmt.Errorf("decoded slot %d is not finite", i)
		}
	}
	return nil
}

func numericalComplex(value complex128) NumericalComplex {
	return NumericalComplex{Real: real(value), Imag: imag(value)}
}

func numericalMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	return numericalQuantile(ordered, 0.5)
}

func almostEqualFloat(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}

func renderNumericalSummary(doc NumericalDocument) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Fast vs Standard numerical reference — %s\n\n", doc.Profile)
	fmt.Fprintf(&out, "- Classification: **%s**\n- Timestamp: `%s`\n- Threshold: `%.3g`\n", doc.Classification, doc.Timestamp.Format(time.RFC3339), doc.Threshold)
	fmt.Fprintf(&out, "- Primary: `%s` (`%s`, dirty=%t)\n- Secondary: `%s` (`%s`, dirty=%t)\n", doc.Primary.Commit, doc.Primary.Ref, doc.Primary.Dirty, doc.Secondary.Commit, doc.Secondary.Ref, doc.Secondary.Dirty)
	fmt.Fprintf(&out, "- Environment: `%s`, `%s/%s`, CPU `%s`, NumCPU `%d`, GOMAXPROCS `%d`\n", doc.Environment.GoVersion, doc.Environment.OS, doc.Environment.Arch, doc.Environment.CPU, doc.Environment.NumCPU, doc.Environment.GOMAXPROCS)

	out.WriteString("\n## Canonical workload\n\n| Field | Value |\n|---|---:|\n")
	fmt.Fprintf(&out, "| Input SHA-256 | `%s` |\n| Slots | %d |\n| LogN / LogSlots | %d / %d |\n| q0 bits | %d |\n| Q-chain bits | `%v` |\n| P bits | `%v` |\n| Mod1 degree / DoubleAngle | %d / %d |\n| EvalMod log scale / LogMessageRatio | %d / %d |\n| Fast executions / effective numerical trials / Standard key trials | %d / %d / %d |\n", doc.Workload.FingerprintSHA256, doc.Workload.InputSlots, doc.Workload.LogN, doc.Workload.LogSlots, doc.Workload.Q0Bits, doc.Workload.QChainBits, doc.Workload.PBits, doc.Workload.PolynomialDegree, doc.Workload.DoubleAngle, doc.Workload.EvalModLogScale, doc.Workload.LogMessageRatio, doc.Execution["fast_executions"], doc.Execution["fast_effective_numerical_trials"], doc.Execution["standard_key_trials"])
	out.WriteString("\nInput is the existing deterministic plaintext-like construction `c0=encoded message, c1=0`; both Fast and every Standard trial receive the same encoded ciphertext coefficients.\n")

	out.WriteString("\n## Output metadata\n\n| Output | Level | Degree | Scale (log2) | IsNTT | IsMontgomery | LogDimensions | Public contract |\n|---|---:|---:|---:|---|---|---|---|\n")
	for _, output := range doc.FastTrials {
		writeNumericalMetadataRow(&out, fmt.Sprintf("Fast %d", output.Index), output)
	}
	for _, output := range doc.StandardTrials {
		writeNumericalMetadataRow(&out, fmt.Sprintf("Standard %d", output.Index), output)
	}
	fmt.Fprintf(&out, "\nFast repeated execution bitwise identical: **%t**; maximum complex difference `%.6e` at slot %d.\n", doc.FastDeterminism.BitwiseIdentical, doc.FastDeterminism.MaxComplexDiff, doc.FastDeterminism.MaxDiffSlot)
	if len(doc.StandardKeyTrials) > 0 {
		out.WriteString("\nStandard key-trial path validation:\n\n| Trial | Fresh keygen/evaluation keys | secret key LevelP | Unique secret | Evaluator path |\n|---:|---|---:|---|---|\n")
		for _, trial := range doc.StandardKeyTrials {
			fmt.Fprintf(&out, "| %d | %t / %t | %d | %t | %s |\n", trial.Index, trial.FreshKeyGenerator, trial.FreshEvaluationKeys, trial.SecretKeyLevelP, trial.UniqueSecretKey, trial.EvaluatorPath)
		}
	}

	out.WriteString("\n## Against original message\n\n| Output | Component | MAE | RMSE | p50 | p95 | p99 | max | worst slot |\n|---|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, output := range doc.FastTrials {
		writeNumericalMetricRows(&out, fmt.Sprintf("Fast %d", output.Index), output.VsOriginal)
	}
	for _, output := range doc.StandardTrials {
		writeNumericalMetricRows(&out, fmt.Sprintf("Standard %d", output.Index), output.VsOriginal)
	}
	writeAggregateRows(&out, "Standard trial median", doc.StandardAggregate)
	out.WriteString("\n### Worst-slot examples vs original\n\n| Output | Component | Slot | Original | Output | Difference |\n|---|---|---:|---|---|---|\n")
	for _, output := range doc.FastTrials {
		writeAgainstOriginalWorstRows(&out, "Fast", output)
	}
	for _, output := range doc.StandardTrials {
		writeAgainstOriginalWorstRows(&out, "Standard", output)
	}

	out.WriteString("\n## Fast vs Standard, per trial pair\n\n| Fast | Standard | Component | MAE | RMSE | p95 | p99 | max | worst slot | Metadata |\n|---:|---:|---|---:|---:|---:|---:|---:|---:|---|\n")
	for _, pair := range doc.FastVsStandard {
		for _, row := range []struct {
			name   string
			metric NumericalErrorMetric
		}{{"real", pair.Metrics.Real}, {"imag", pair.Metrics.Imag}, {"complex", pair.Metrics.Complex}} {
			fmt.Fprintf(&out, "| %d | %d | %s | %.6e | %.6e | %.6e | %.6e | %.6e | %d | %s |\n", pair.FastTrialIndex, pair.StandardTrialIndex, row.name, row.metric.MeanAbsoluteError, row.metric.RMSE, row.metric.P95, row.metric.P99, row.metric.Max, row.metric.WorstSlot, pair.Metadata.Classification)
		}
	}
	out.WriteString("\n### Fast-vs-Standard worst-slot examples\n\n| Fast | Standard | Component | Slot | Original | Fast | Standard | Fast−Standard |\n|---:|---:|---|---:|---|---|---|---|\n")
	for _, pair := range doc.FastVsStandard {
		for _, row := range []struct {
			name    string
			example NumericalPairwiseWorstExample
		}{{"real", pair.WorstExamples.Real}, {"imag", pair.WorstExamples.Imag}, {"complex", pair.WorstExamples.Complex}} {
			fmt.Fprintf(&out, "| %d | %d | %s | %d | %s | %s | %s | %s |\n", pair.FastTrialIndex, pair.StandardTrialIndex, row.name, row.example.Slot, formatNumericalComplex(row.example.Original), formatNumericalComplex(row.example.Fast), formatNumericalComplex(row.example.Standard), formatNumericalComplex(row.example.Difference))
		}
	}

	out.WriteString("\n## Precision bits\n\nPrecision is computed per slot as `-log2(max(|output-original|, 1e-30))` using complex magnitude error; quantiles use linear interpolation over sorted samples.\n\n| Output | Mean | Median | p05 | Minimum | Minimum slot | Slot error | CKKS helper median L2 |\n|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, output := range doc.FastTrials {
		writePrecisionRow(&out, "Fast", output)
	}
	for _, output := range doc.StandardTrials {
		writePrecisionRow(&out, "Standard", output)
	}
	fmt.Fprintf(&out, "\nMedian across Fast trials: `%.4f` bits; median across Standard trial medians: `%.4f` bits (drop `%.4f` bits).\n", doc.FastAggregate.MedianOfTrialMedianBits, doc.StandardAggregate.MedianOfTrialMedianBits, doc.ClassificationChecks.FastToStandardMedianPrecisionDrop)

	spread := doc.StandardToStandardVariability
	fmt.Fprintf(&out, "\n## Standard-to-Standard variability\n\n- Per-slot Standard mean was computed across all trials; SHA-256: `%s`. Sample means: %s\n- RMS spread around per-slot means (real / imag / complex): `%.6e` / `%.6e` / `%.6e`\n- Maximum spread from per-slot mean: `%.6e` at slot %d (trial %d)\n- Maximum pairwise Standard difference: `%.6e` at slot %d (trials %d vs %d)\n", spread.PerSlotMeanFingerprintSHA256, formatMeanSamples(spread.PerSlotMeanSamples), spread.RMSRealSpread, spread.RMSImagSpread, spread.RMSComplexSpread, spread.MaxComplexSpread, spread.MaxSpreadSlot, spread.MaxSpreadTrialIndex, spread.MaxPairwiseComplexDiff, spread.MaxPairwiseSlot, spread.MaxPairwiseTrialA, spread.MaxPairwiseTrialB)

	out.WriteString("\n## 1e-2 threshold audit\n\n| Comparison | Real > threshold | Imag > threshold | Combined fraction |\n|---|---:|---:|---:|\n")
	writeThresholdRow(&out, "Fast vs original", doc.FastVsOriginalThreshold)
	writeThresholdRow(&out, "Standard vs original", doc.StandardVsOriginalThreshold)
	writeThresholdRow(&out, "Fast vs Standard", doc.FastVsStandardThreshold)
	fmt.Fprintf(&out, "\nMedian Fast complex RMSE: `%.6e`; median Standard complex RMSE: `%.6e`; median Fast-vs-Standard complex RMSE: `%.6e`; max Fast-vs-Standard complex difference: `%.6e`.\n", doc.FastAggregate.Complex.RMSE, doc.StandardAggregate.Complex.RMSE, medianPairwiseComplexRMSE(doc.FastVsStandard), maximumPairwiseComplexDifference(doc.FastVsStandard))

	out.WriteString("\n## Classification checks\n\n")
	fmt.Fprintf(&out, "- Fast vs original has no coordinate above 1e-2: `%t`\n- Fast vs Standard has no coordinate above 1e-2: `%t`\n- Median precision degradation is at most 2 bits: `%t`\n- Fast complex RMSE is within 4x Standard, or Standard RMSE is within the documented near-zero floor: `%t`\n- Public Bootstrap Level/Scale/dimensions contract is respected: `%t`\n- Standard near-zero floor: `%.6e`\n", doc.ClassificationChecks.FastVsOriginalNoCoordinateExceeds, doc.ClassificationChecks.FastVsStandardNoCoordinateExceeds, doc.ClassificationChecks.MedianPrecisionDropWithinTwoBits, doc.ClassificationChecks.FastRMSEWithinFourXOrNearZero, doc.ClassificationChecks.PublicMetadataContractRespected, doc.StandardRMSENearZeroFloor)

	renderNumericalStageEvidence(&out, doc)

	out.WriteString("\n## Limitations\n\n")
	for _, limitation := range doc.Limitations {
		fmt.Fprintf(&out, "- %s\n", limitation)
	}
	out.WriteString("\nThe JSON artifact contains aggregate metrics and worst-slot evidence only; decoded slot arrays are intentionally not persisted.\n")
	return out.String()
}

func renderNumericalStageEvidence(out *strings.Builder, doc NumericalDocument) {
	if doc.StageLockstep == nil {
		return
	}
	stage := doc.StageLockstep
	out.WriteString("\n## Decoded-domain Bootstrap SNR\n\nSNR here is numerical distortion, not RLWE security noise or a noise budget. Each Bootstrap SNR uses that mode's actual decoded pre-Bootstrap vector as the signal reference and its decoded post-Bootstrap output as the observation.\n\n| Mode / trial | Signal power | Signal RMS | Error power | Pre→Post RMSE | Bootstrap SNR (dB / status) | Pre-Bootstrap vs canonical original RMSE |\n|---|---:|---:|---:|---:|---:|---:|\n")
	for _, output := range doc.StandardTrials {
		fmt.Fprintf(out, "| Standard %d | %s | %s | %s | %s | %s | %.6e |\n", output.Index, formatSNRMetric(output.BootstrapSNR.SignalPower), formatSNRMetric(output.BootstrapSNR.SignalRMS), formatSNRMetric(output.BootstrapSNR.NoisePower), formatSNRMetric(output.BootstrapSNR.NoiseRMSE), formatSNRDB(output.BootstrapSNR), output.PreBootstrapVsCanonicalOriginalComplexRMSE)
	}
	for _, output := range doc.FastTrials {
		fmt.Fprintf(out, "| Fast Q-prefix %d | %s | %s | %s | %s | %s | %.6e |\n", output.Index, formatSNRMetric(output.BootstrapSNR.SignalPower), formatSNRMetric(output.BootstrapSNR.SignalRMS), formatSNRMetric(output.BootstrapSNR.NoisePower), formatSNRMetric(output.BootstrapSNR.NoiseRMSE), formatSNRDB(output.BootstrapSNR), output.PreBootstrapVsCanonicalOriginalComplexRMSE)
	}
	if len(doc.StandardTrials) > 0 {
		fmt.Fprintf(out, "\n`STANDARD_BOOTSTRAP_SNR_DB=%s`\n", formatSNRDB(doc.StandardTrials[0].BootstrapSNR))
	}
	if len(doc.FastTrials) > 0 {
		fmt.Fprintf(out, "`FAST_BOOTSTRAP_SNR_DB=%s`\n", formatSNRDB(doc.FastTrials[0].BootstrapSNR))
	}

	out.WriteString("\n## Stage reference SNR and divergence\n\n`D_i` is Fast-vs-genuine-Standard complex RMSE. Amplification and SNR deltas follow the explicit stage topology, not table adjacency: C2S real/imag are parallel children of ModUp; each EvalMod branch uses its matching C2S branch; S2C uses a joint EvalMod real+imag reference; final public output may use S2C.\n\n| Checkpoint | Fast state | Standard state | D_i RMSE | A_i parent | A_i | Stage reference SNR (dB / status) | ΔSNR_i parent | ΔSNR_i (dB / status) | Max complex diff |\n|---|---|---|---:|---|---:|---:|---|---:|---:|\n")
	for _, checkpoint := range stage.Checkpoints {
		if !checkpoint.Comparable || checkpoint.FastVsStandard == nil {
			continue
		}
		aValue := formatOptionalFloat(checkpoint.Amplification)
		if checkpoint.AmplificationStatus == "NOT_APPLICABLE_COMBINED_BRANCH" {
			aValue = "n/a (combined branches)"
		}
		fmt.Fprintf(out, "| %s | %s | %s | %s | %s | %s (%s) | %s | %s | %s (%s) | %.6e |\n", checkpoint.Name,
			formatInternalState(checkpoint.Fast), formatInternalState(checkpoint.Standard), formatOptionalFloat(checkpoint.D),
			formatMetricParent(checkpoint.AmplificationParent), aValue, checkpoint.AmplificationStatus,
			formatSNRDB(checkpoint.StageReferenceSNR), formatMetricParent(checkpoint.DeltaSNRParent),
			formatOptionalFloat(checkpoint.DeltaSNRDB), checkpoint.DeltaSNRStatus, checkpoint.FastVsStandard.Complex.Max)
	}
	fmt.Fprintf(out, "\n`LARGEST_RAW_AMPLIFICATION_CHECKPOINT=%s`\n`LARGEST_RAW_AMPLIFICATION_FACTOR=%s`\n`LARGEST_SNR_DROP_CHECKPOINT=%s`\n`LARGEST_SNR_DROP_DB=%s`\n",
		stage.LargestRawAmplificationCheckpoint, formatOptionalFloat(stage.LargestRawAmplificationFactor), stage.LargestSNRDropCheckpoint, formatOptionalFloat(stage.LargestSNRDropDB))
	fmt.Fprintf(out, "\n- First observable: `%s`, max complex diff `%s` (threshold `%.3e`)\n- First material: `%s`, max complex diff `%s` (threshold `%.12g`)\n- Final Fast-vs-Standard RMSE: `%.12g`\n- Combined-branch EvalMod error RMSE: `%s`\n- Combined-branch S2C amplification: `%s` (status `%s`; `D_s2c / RMSE(concat(EvalMod real, EvalMod imag))`)\n- Current classification: `%s`\n",
		stage.FirstObservable, formatOptionalFloat(stage.FirstObservableMaxDiff), stage.ObservableThreshold,
		stage.FirstMaterial, formatOptionalFloat(stage.FirstMaterialMaxDiff), stage.MaterialThreshold,
		stage.FinalFastStandardRMSE, formatOptionalFloat(stage.CombinedBranchEvalModErrorRMSE),
		formatOptionalFloat(stage.CombinedBranchS2CAmplification), stage.CombinedBranchS2CAmplificationStatus, stage.Classification)
	fmt.Fprintf(out, "\n`FIRST_OBSERVABLE_CHECKPOINT=%s`\n`FIRST_OBSERVABLE_MAX_DIFF=%s`\n`FIRST_MATERIAL_CHECKPOINT=%s`\n`FIRST_MATERIAL_MAX_DIFF=%s`\n`FINAL_FAST_STANDARD_RMSE=%.12g`\n`COMBINED_BRANCH_S2C_AMPLIFICATION_FACTOR=%s`\n`S2C_AMPLIFICATION_FACTOR=%s (combined-branch compatibility alias)`\n",
		stage.FirstObservable, formatOptionalFloat(stage.FirstObservableMaxDiff), stage.FirstMaterial, formatOptionalFloat(stage.FirstMaterialMaxDiff), stage.FinalFastStandardRMSE,
		formatOptionalFloat(stage.CombinedBranchS2CAmplification), formatOptionalFloat(stage.CombinedBranchS2CAmplification))

	if len(stage.EvalModInternal) > 0 {
		out.WriteString("\n## EvalMod direct stage-equivalence replay\n\nEach Fast checkpoint is decoded from authoritative Q-prefix rows; Standard is decrypted with the genuine Standard secret key. Comparisons use the same mathematical scale directly, with no representation-specific power-of-two alignment. Replay verification compares the manual stage sequence with each actual public EvalMod output.\n\n| Checkpoint | Fast L / log2(scale) / degree / NTT / Montgomery / rows | Standard L / log2(scale) / degree / NTT / Montgomery / rows | D_i RMSE | A_i | Stage reference SNR (dB / status) | ΔSNR_i (dB / status) | Max complex diff |\n|---|---|---|---:|---:|---:|---:|---:|---:|\n")
		for _, checkpoint := range stage.EvalModInternal {
			if checkpoint.Comparable && checkpoint.FastVsStandard != nil {
				fmt.Fprintf(out, "| %s | %s | %s | %s | %s (%s) | %s | %s (%s) | %.6e |\n", checkpoint.Name,
					formatInternalState(checkpoint.Fast), formatInternalState(checkpoint.Standard), formatOptionalFloat(checkpoint.D),
					formatOptionalFloat(checkpoint.Amplification), checkpoint.AmplificationStatus, formatSNRDB(checkpoint.StageReferenceSNR),
					formatOptionalFloat(checkpoint.DeltaSNRDB), checkpoint.DeltaSNRStatus, checkpoint.FastVsStandard.Complex.Max)
			} else {
				fmt.Fprintf(out, "| %s | — | — | — | — | NOT_COMPARABLE: %s |\n", checkpoint.Name, checkpoint.NotComparableReason)
			}
		}
		out.WriteString("\nReplay verification RMSE by mode: ")
		replayKeys := make([]string, 0, len(stage.EvalModReplayRMSE))
		for key := range stage.EvalModReplayRMSE {
			replayKeys = append(replayKeys, key)
		}
		sort.Strings(replayKeys)
		for i, key := range replayKeys {
			if i > 0 {
				out.WriteString("; ")
			}
			fmt.Fprintf(out, "`%s=%.6e (verified=%t)`", key, stage.EvalModReplayRMSE[key], stage.EvalModReplayVerified[key])
		}
		out.WriteString("\n")
	}
	if len(stage.EvalModCapacityAudit) > 0 {
		out.WriteString("\n## Standard-equivalent Fast EvalMod Q-prefix capacity audit\n\nBounds are exact observed centered coefficient maxima by ciphertext component. Every checkpoint requires strict `2B < S_Q`; execution stops at the first failed guard.\n\n| Checkpoint | Level | Rows | Exact prefix product S_Q | Degree | MaxAbs by component | Strict `2B < S_Q` |\n|---|---:|---:|---|---:|---|---:|\n")
		for _, audit := range stage.EvalModCapacityAudit {
			fmt.Fprintf(out, "| %s | %d | %d | `%s` | %d | `%v` | %t |\n", audit.Checkpoint, audit.Level, audit.Rows, audit.PrefixProduct, audit.Degree, audit.MaxAbs, audit.StrictFit)
		}
	}

	if stage.PolynomialPlan != nil {
		plan := stage.PolynomialPlan
		fmt.Fprintf(out, "\n## Polynomial plan and generated-power evidence\n\n- Fast diagnostic PS plan (%s branch): degree `%d`, base `%d`, level `%d`, target scale log2 `%.6f`, exact `%s`, blocks `%d`.\n", plan.Branch, plan.Degree, plan.Base, plan.Level, plan.ScaleLog2, plan.ScaleExact, plan.BlockCount)
		if len(stage.GeneratedPowerEvidence) > 0 {
			out.WriteString("\nEach Standard power is generated by the genuine Standard `PowerBasis.GenPower` using the Standard CKKS evaluator; comparisons decrypt Standard ciphertexts and decode Fast through authoritative Q-prefix rows.\n\n| Branch | Power | Fast Level / Scale log2 | Standard Level / Scale log2 | Fast rows | Standard rows | Level/Scale match | Complex RMSE Fast-vs-Standard | Max complex diff | Reference provenance |\n|---|---:|---|---|---:|---:|---|---:|---:|---|\n")
			for _, power := range stage.GeneratedPowerEvidence {
				fmt.Fprintf(out, "| %s | T%d | %d / %.6f | %d / %.6f | %d | %d | %t | %.6e | %.6e | %s |\n", power.Branch, power.Power, power.FastLevel, power.FastScaleLog2, power.StandardLevel, power.StandardScaleLog2, power.FastMaintainedRows, power.StandardAuthorityRows, power.LevelScaleMatch, power.RMSEFastStandard, power.MaxComplexDiffFastStandard, power.ReferenceKind)
			}
		}
	}

	if len(stage.ScaleAudit) > 0 {
		out.WriteString("\n## EvalMod exact-scale audit\n\nExact scale strings are retained alongside the observed logical Level at each direct Standard/Fast replay checkpoint. The target scale follows the Standard Mod1 Q schedule.\n\n| Checkpoint | Level | Scale log2 | Exact scale | Target log2 / exact | DoubleAngle round |\n|---|---:|---:|---|---|---:|\n")
		for _, audit := range stage.ScaleAudit {
			if !strings.HasPrefix(audit.Checkpoint, "evalmod") && !strings.HasPrefix(audit.Checkpoint, "real/") && !strings.HasPrefix(audit.Checkpoint, "imag/") {
				continue
			}
			fmt.Fprintf(out, "| %s | %d | %.6f | `%s` | %.6f / `%s` | %d |\n", audit.Checkpoint, audit.Level, audit.ScaleLog2, audit.ScaleExact, audit.TargetScaleLog2, audit.TargetScaleExact, audit.DoubleAngleRound)
		}
	}

	out.WriteString("\n## S2C attribution\n\nThe S2C amplification is reported separately as a combined-branch metric: the S2C error RMSE divided by the joint RMSE of concatenated EvalMod real and imag semantic errors. It is not an ordinary sequential `A_i`. The S2C ΔSNR parent is the joint EvalMod real+imag reference, not the preceding table row.\n\n")
}

func formatOptionalFloat(value *float64) string {
	if value == nil {
		return "—"
	}
	return fmt.Sprintf("%.6e", *value)
}

func formatMetricParent(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func formatSNRMetric(value *float64) string {
	if value == nil {
		return "—"
	}
	return fmt.Sprintf("%.6e", *value)
}

func formatSNRDB(value numericalmetrics.SNR) string {
	if value.SNRDB != nil {
		return fmt.Sprintf("%.6f / %s", *value.SNRDB, value.Status)
	}
	switch value.Status {
	case numericalmetrics.PositiveInfinity:
		return "+Inf / POSITIVE_INFINITY"
	case numericalmetrics.UndefinedZeroSignal:
		return "undefined / UNDEFINED_ZERO_SIGNAL"
	case numericalmetrics.NotComparable:
		return "— / NOT_COMPARABLE"
	default:
		return "— / " + string(value.Status)
	}
}

func formatInternalState(state NumericalStageState) string {
	metadata := state.Metadata
	return fmt.Sprintf("L%d / %.6f / d%d / %t / %t / %d(+%d maintained)", metadata.Level, metadata.ScaleLog2, metadata.Degree, metadata.IsNTT, metadata.IsMontgomery, state.AuthorityRows, state.MaintainedRows)
}

func writeNumericalMetadataRow(out *strings.Builder, name string, output NumericalOutput) {
	m := output.Metadata
	fmt.Fprintf(out, "| %s | %d | %d | %.6f | %t | %t | (%d,%d) | %s |\n", name, m.Level, m.Degree, m.ScaleLog2, m.IsNTT, m.IsMontgomery, m.LogDimensions.Rows, m.LogDimensions.Cols, output.PublicContract.Status)
}

func writeNumericalMetricRows(out *strings.Builder, name string, metrics NumericalMetricSet) {
	for _, row := range []struct {
		name   string
		metric NumericalErrorMetric
	}{{"real", metrics.Real}, {"imag", metrics.Imag}, {"complex", metrics.Complex}} {
		fmt.Fprintf(out, "| %s | %s | %.6e | %.6e | %.6e | %.6e | %.6e | %.6e | %d |\n", name, row.name, row.metric.MeanAbsoluteError, row.metric.RMSE, row.metric.P50, row.metric.P95, row.metric.P99, row.metric.Max, row.metric.WorstSlot)
	}
}

func writeAgainstOriginalWorstRows(out *strings.Builder, kind string, output NumericalOutput) {
	for _, row := range []struct {
		name   string
		metric NumericalErrorMetric
	}{{"real", output.VsOriginal.Real}, {"imag", output.VsOriginal.Imag}, {"complex", output.VsOriginal.Complex}} {
		fmt.Fprintf(out, "| %s %d | %s | %d | %s | %s | %s |\n", kind, output.Index, row.name, row.metric.WorstSlot, formatNumericalComplex(row.metric.WorstReference), formatNumericalComplex(row.metric.WorstOutput), formatNumericalComplex(row.metric.WorstDifference))
	}
}

func formatNumericalComplex(value NumericalComplex) string {
	return fmt.Sprintf("`%.8g%+.8gi`", value.Real, value.Imag)
}

func writeAggregateRows(out *strings.Builder, name string, aggregate NumericalAggregate) {
	for _, row := range []struct {
		name   string
		metric NumericalMetricAggregate
	}{{"real", aggregate.Real}, {"imag", aggregate.Imag}, {"complex", aggregate.Complex}} {
		fmt.Fprintf(out, "| %s | %s | %.6e | %.6e | %.6e | %.6e | %.6e | %.6e | median of trial maxima |\n", name, row.name, row.metric.MeanAbsoluteError, row.metric.RMSE, row.metric.P50, row.metric.P95, row.metric.P99, row.metric.Max)
	}
}

func writeThresholdRow(out *strings.Builder, name string, audit NumericalThresholdAudit) {
	fmt.Fprintf(out, "| %s (%d comparisons) | %d / %d (%.6f%%) | %d / %d (%.6f%%) | %d / %d (%.6f%%) |\n", name, audit.Comparisons, audit.RealExceeding, audit.RealCoordinates, audit.RealFraction*100, audit.ImagExceeding, audit.ImagCoordinates, audit.ImagFraction*100, audit.TotalExceeding, audit.TotalCoordinates, audit.TotalFraction*100)
}

func writePrecisionRow(out *strings.Builder, kind string, output NumericalOutput) {
	fmt.Fprintf(out, "| %s %d | %.4f | %.4f | %.4f | %.4f | %d | %.6e | %.4f |\n", kind, output.Index, output.PrecisionBits.MeanBits, output.PrecisionBits.MedianBits, output.PrecisionBits.P05Bits, output.PrecisionBits.MinimumBits, output.PrecisionBits.MinimumSlot, output.PrecisionBits.MinimumSlotError, output.CKKSHelper.Median.L2)
}

func formatMeanSamples(samples []NumericalMeanSlotSample) string {
	parts := make([]string, len(samples))
	for i, sample := range samples {
		parts[i] = fmt.Sprintf("slot %d=(%.6e%+.6ei)", sample.Slot, sample.Mean.Real, sample.Mean.Imag)
	}
	return strings.Join(parts, "; ")
}

func medianPairwiseComplexRMSE(pairs []NumericalPairwiseComparison) float64 {
	values := make([]float64, len(pairs))
	for i, pair := range pairs {
		values[i] = pair.Metrics.Complex.RMSE
	}
	return numericalMedian(values)
}

func maximumPairwiseComplexDifference(pairs []NumericalPairwiseComparison) float64 {
	maximum := 0.0
	for _, pair := range pairs {
		maximum = math.Max(maximum, pair.Metrics.Complex.Max)
	}
	return maximum
}
