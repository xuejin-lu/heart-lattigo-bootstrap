package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const bts009NoiseEvidenceLimit = 16 << 20

var bts009OriginalQ = []uint64{
	4398044938241, 4293918721, 4296540161, 4291952641, 4298506241,
	4299685889, 4300210177, 4300865537, 4289462273, 4301651969,
	4288806913, 4302176257, 4302569473, 4286709761, 4304535553,
	4286054401, 4284874753, 4306894849, 4284088321, 4398043496449,
	4398042972161, 4398042185729, 4398052147201, 4398052540417,
	4398053588993, 4398054506497, 4398054899713, 4398055555073,
	4398058045441, 4398034845697,
}

var bts009OriginalP = []uint64{
	4398063157249, 4398029340673, 4398065909761, 4398028029953,
	4398070628353,
}

type bts009NoiseSide struct {
	Seed              int64  `json:"sampling_seed"`
	SampleCount       int    `json:"coefficient_sample_count"`
	CoefficientSHA256 string `json:"coefficient_sha256"`
	Capacity          struct {
		Rows                int    `json:"rows"`
		CheckedCoefficients int    `json:"checked_coefficients"`
		MaxAbs              string `json:"max_abs"`
		PrefixProduct       string `json:"prefix_product"`
		StrictFit           bool   `json:"strict_fit"`
	} `json:"qprefix_capacity"`
}

type bts009NoiseEvent struct {
	Operation         uint64          `json:"operation"`
	Level             int             `json:"level"`
	Scale             string          `json:"scale"`
	Left              bts009NoiseSide `json:"left"`
	Right             bts009NoiseSide `json:"right"`
	MultiplyCompleted bool            `json:"multiply_completed"`
	Failure           string          `json:"failure,omitempty"`
}

type bts009NoiseCall struct {
	SchemaVersion  string             `json:"schema_version"`
	Task           string             `json:"task"`
	Phase          string             `json:"phase"`
	AttemptIndex   int                `json:"attempt_index"`
	Sigma          float64            `json:"sigma_coefficient"`
	RootSeed       int64              `json:"root_seed"`
	OperationStart uint64             `json:"operation_start"`
	OperationEnd   uint64             `json:"operation_end"`
	Events         []bts009NoiseEvent `json:"events"`
}

type bts009SampleReport struct {
	AttemptIndex   int                                `json:"attempt_index"`
	Phase          string                             `json:"phase"`
	BootstrapNS    int64                              `json:"bootstrap_elapsed_ns"`
	AllocatedBytes uint64                             `json:"allocated_bytes"`
	Allocations    uint64                             `json:"allocation_count"`
	OutputSHA256   string                             `json:"output_sha256"`
	NoiseEvents    int                                `json:"noise_event_count,omitempty"`
	Quality        numericalmetrics.RealVectorQuality `json:"input_relative_quality"`
}

type bts009ArmReport struct {
	Arm                    string                                 `json:"arm"`
	Backend                string                                 `json:"backend"`
	NoiseMode              string                                 `json:"noise_mode"`
	BackendCommit          string                                 `json:"backend_commit"`
	BinarySHA256           string                                 `json:"binary_sha256"`
	EvaluatorAndBTSBuildNS int64                                  `json:"new_bts_construction_ns"`
	Cold                   bts009SampleReport                     `json:"cold_sample"`
	Warm                   []bts009SampleReport                   `json:"warm_samples"`
	WarmTime               perfmeasure.BTSRepeatedStatistics      `json:"warm_time_ns"`
	WarmAllocatedBytes     perfmeasure.BTSRepeatedStatistics      `json:"warm_allocated_bytes"`
	WarmAllocationCount    perfmeasure.BTSRepeatedStatistics      `json:"warm_allocation_count"`
	ConstructionPlusColdNS int64                                  `json:"construction_plus_cold_ns"`
	WarmSpeedupVsStandard  float64                                `json:"warm_speedup_vs_standard"`
	TotalNoiseEvents       int                                    `json:"total_noise_events"`
	NoiseEvidence          []perfmeasure.BTSNoiseEvidenceArtifact `json:"noise_evidence_artifacts,omitempty"`
	Provenance             btsRunProvenance                       `json:"provenance"`
}

type bts009NoiseOverhead struct {
	ColdExtraNS                  int64   `json:"cold_extra_ns"`
	ColdMultiplier               float64 `json:"cold_multiplier"`
	WarmMedianExtraNS            float64 `json:"warm_median_extra_ns"`
	WarmMedianMultiplier         float64 `json:"warm_median_multiplier"`
	WarmAllocatedBytesMeanExtra  float64 `json:"warm_allocated_bytes_mean_extra"`
	WarmAllocationCountMeanExtra float64 `json:"warm_allocation_count_mean_extra"`
}

type bts009PrecisionComparison struct {
	StandardWarmMeanBits        *float64 `json:"standard_warm_mean_global_rmse_bits,omitempty"`
	FastOffWarmMeanBits         *float64 `json:"fast_off_warm_mean_global_rmse_bits,omitempty"`
	FastNoiseWarmMeanBits       *float64 `json:"fast_noise_warm_mean_global_rmse_bits,omitempty"`
	FastOffDistanceToStandard   *float64 `json:"fast_off_distance_to_standard_bits,omitempty"`
	FastNoiseDistanceToStandard *float64 `json:"fast_noise_distance_to_standard_bits,omitempty"`
	NoiseMovedCloserToStandard  bool     `json:"noise_moved_global_precision_closer_to_standard"`
	Status                      string   `json:"status"`
}

type bts009ReportDocument struct {
	SchemaVersion       string                    `json:"schema_version"`
	Task                string                    `json:"task"`
	GeneratedAt         time.Time                 `json:"generated_at"`
	Status              string                    `json:"status"`
	InputSHA256         string                    `json:"input_sha256"`
	MatchedEnvironment  bool                      `json:"matched_environment"`
	Standard            bts009ArmReport           `json:"standard"`
	FastOff             bts009ArmReport           `json:"fast_off"`
	FastNoise           bts009ArmReport           `json:"fast_noise"`
	NoiseOverhead       bts009NoiseOverhead       `json:"noise_overhead_vs_fast_off"`
	PrecisionComparison bts009PrecisionComparison `json:"precision_proximity"`
	Historical008       struct {
		StandardWarmMedianMS float64 `json:"standard_warm_median_ms"`
		FastWarmMedianMS     float64 `json:"fast_warm_median_ms"`
		WarmSpeedup          float64 `json:"warm_speedup"`
		Note                 string  `json:"note"`
	} `json:"historical_bts_008_comparison"`
}

func runBTS009Report(standardPath, fastOffPath, fastNoisePath, outJSON, outMarkdown string) error {
	var standard, fastOff, fastNoise perfmeasure.BTSRunDocument
	for _, input := range []struct {
		path string
		run  *perfmeasure.BTSRunDocument
		arm  string
	}{{standardPath, &standard, "Standard"}, {fastOffPath, &fastOff, "Fast OFF"}, {fastNoisePath, &fastNoise, "Fast+Noise"}} {
		if err := readJSON(input.path, input.run); err != nil {
			return fmt.Errorf("read %s run: %w", input.arm, err)
		}
	}
	if err := validateBTS009Run(&standard, "standard", "off"); err != nil {
		return fmt.Errorf("validate Standard run: %w", err)
	}
	if err := validateBTS009Run(&fastOff, "fast", "off"); err != nil {
		return fmt.Errorf("validate Fast OFF run: %w", err)
	}
	if err := validateBTS009Run(&fastNoise, "fast", "on"); err != nil {
		return fmt.Errorf("validate Fast+Noise run: %w", err)
	}
	if err := validateMatchedBTS009Runs(standard, fastOff, fastNoise); err != nil {
		return err
	}
	input, err := readBTSVector(standard.Input)
	if err != nil {
		return fmt.Errorf("read BTS original input: %w", err)
	}
	for _, run := range []perfmeasure.BTSRunDocument{fastOff, fastNoise} {
		other, err := readBTSVector(run.Input)
		if err != nil {
			return fmt.Errorf("read %s BTS original input: %w", run.NoiseMode, err)
		}
		if !equalFloatVectors(input, other) {
			return errors.New("all BTS-009 arms must use the same original plaintext vector")
		}
	}
	standardArm, err := summarizeBTS009Arm("Standard", standard, input)
	if err != nil {
		return err
	}
	fastArm, err := summarizeBTS009Arm("Fast OFF", fastOff, input)
	if err != nil {
		return err
	}
	noiseArm, err := summarizeBTS009Arm("Fast+Noise", fastNoise, input)
	if err != nil {
		return err
	}
	standardMedian := standardArm.WarmTime.Median
	fastArm.WarmSpeedupVsStandard = standardMedian / fastArm.WarmTime.Median
	noiseArm.WarmSpeedupVsStandard = standardMedian / noiseArm.WarmTime.Median
	if fastArm.WarmTime.Median <= 0 || noiseArm.WarmTime.Median <= 0 || fastArm.Cold.BootstrapNS <= 0 || noiseArm.Cold.BootstrapNS <= 0 {
		return errors.New("non-positive Bootstrap duration prevents speed and overhead ratios")
	}
	overhead := bts009NoiseOverhead{
		ColdExtraNS:                  int64(noiseArm.Cold.BootstrapNS - fastArm.Cold.BootstrapNS),
		ColdMultiplier:               float64(noiseArm.Cold.BootstrapNS) / float64(fastArm.Cold.BootstrapNS),
		WarmMedianExtraNS:            noiseArm.WarmTime.Median - fastArm.WarmTime.Median,
		WarmMedianMultiplier:         noiseArm.WarmTime.Median / fastArm.WarmTime.Median,
		WarmAllocatedBytesMeanExtra:  noiseArm.WarmAllocatedBytes.Mean - fastArm.WarmAllocatedBytes.Mean,
		WarmAllocationCountMeanExtra: noiseArm.WarmAllocationCount.Mean - fastArm.WarmAllocationCount.Mean,
	}
	report := bts009ReportDocument{
		SchemaVersion: "bts-fast-noise-fidelity-009-report.v1", Task: bts009Task,
		GeneratedAt: time.Now().UTC(), Status: "ALL_THREE_ARMS_COMPLETE_AND_VALIDATED",
		InputSHA256: standard.Input.SHA256, MatchedEnvironment: true,
		Standard: standardArm, FastOff: fastArm, FastNoise: noiseArm, NoiseOverhead: overhead,
	}
	report.Historical008.StandardWarmMedianMS = 154536.049
	report.Historical008.FastWarmMedianMS = 5044.745
	report.Historical008.WarmSpeedup = report.Historical008.StandardWarmMedianMS / report.Historical008.FastWarmMedianMS
	report.Historical008.Note = "BTS-008 historical baseline only; different task capture and source pins, not substituted into BTS-009."
	report.PrecisionComparison = compareBTS009Precision(standardArm.Warm, fastArm.Warm, noiseArm.Warm)
	if err := writeExclusiveJSON(outJSON, report); err != nil {
		return err
	}
	if err := writeExclusiveText(outMarkdown, renderBTS009Report(report)); err != nil {
		return err
	}
	fmt.Printf("BTS-009 tri-arm report complete: %s; Fast Noise warm overhead %.4fx\n", outMarkdown, overhead.WarmMedianMultiplier)
	return nil
}

func validateBTS009Run(run *perfmeasure.BTSRunDocument, backend, noiseMode string) error {
	wantCommit, err := btsBackendPin(bts009Task, backend)
	if err != nil {
		return err
	}
	if run.SchemaVersion != perfmeasure.BTSFastNoise009SchemaVersion || run.Task != bts009Task || run.Status != "COMPLETE" {
		return fmt.Errorf("schema/task/status mismatch: %q/%q/%q", run.SchemaVersion, run.Task, run.Status)
	}
	if run.Backend != backend || run.BackendCommit != wantCommit || run.BinarySHA256 == "" || run.MeasurementAdapterSHA256 == "" || run.PrimaryCommit == "" {
		return errors.New("backend, source pin, or binary provenance mismatch")
	}
	if run.NoiseMode != noiseMode || run.BTSManifestSHA256 != bts008SourceManifestSHA || run.MeasurementProcesses != 1 || run.BootstrapBudget != 6 || run.BootstrapReservations != 6 || len(run.Samples) != 6 {
		return errors.New("Noise mode, original BTS manifest, process count, six-call budget, or sample count mismatch")
	}
	if backend == "standard" && (run.NoiseHookSupported || run.NoiseHookEnabled) {
		return errors.New("genuine Standard run unexpectedly exposes or enables the Fast-only Noise Hook")
	}
	if backend == "fast" && (!run.NoiseHookSupported || run.NoiseHookEnabled != (noiseMode == "on")) {
		return errors.New("Fast Noise Hook support/enabled state does not match the selected arm")
	}
	if run.ProfileSHA256 == "" || run.ArmConfigSHA256 == "" || run.Circuit == "" {
		return errors.New("BTS-009 configuration fingerprints or circuit description are missing")
	}
	if noiseMode == "on" {
		if run.NoiseSigma != 4.004280012097432 || run.NoiseRootSeed != 20261027 || len(run.NoiseEvidence) != 6 {
			return errors.New("Fast+Noise frozen sigma, root seed, or six-call evidence mismatch")
		}
		if err := validateBTS009NoiseEvidence(run); err != nil {
			return err
		}
	} else if run.NoiseSigma != 0 || run.NoiseRootSeed != 0 || len(run.NoiseEvidence) != 0 {
		return errors.New("Noise-OFF run contains Noise configuration or event evidence")
	}
	p := run.Parameters
	if p.LogN != 16 || p.RingN != 1<<16 || p.Slots != 1<<15 || p.SlotLimbs != 15 || p.Mode != "BTSModePlain" || p.Seed != 0 || p.Amplitude != 1 || p.EphemeralH != 32 || p.MaxLevel != 29 || p.QCount != 30 || p.PCount != 5 || p.InputLevel != 3 || p.InputScale != math.Pow(2, 32) || p.Mod1Level != 26 || p.Mod1Scale != math.Pow(2, 42) || p.Mod1Degree != 31 || p.Mod1PSDepth != 5 || p.DoubleAngle != 3 || p.OutputLevel != 18 || p.OutputScale != math.Pow(2, 32) || !equalUint64s(p.QPrimes, bts009OriginalQ) || !equalUint64s(p.PPrimes, bts009OriginalP) {
		return fmt.Errorf("effective BTS parameters differ from the frozen profile: %+v", p)
	}
	if run.Input.Encoding != "IEEE754-f64-little-endian" || run.Input.SHA256 != bts008ExpectedInputSHA || run.Input.Slots != 1<<15 || run.Input.FiniteSlots != 1<<15 || run.Input.Level != 3 || run.Input.Scale != math.Pow(2, 32) {
		return errors.New("original input identity, shape, or metadata mismatch")
	}
	for index, sample := range run.Samples {
		wantPhase := fmt.Sprintf("warm_bootstrap_%02d", index)
		if index == 0 {
			wantPhase = "first_cold_bootstrap"
		}
		if sample.AttemptIndex != index+1 || sample.Phase != wantPhase || sample.BootstrapElapsedNS <= 0 || sample.AllocatedBytes == 0 || sample.Allocations == 0 || sample.InputLevel != 3 || sample.InputScale != math.Pow(2, 32) {
			return fmt.Errorf("sample %d has invalid phase, timing, allocations, or input metadata", index+1)
		}
		if sample.OutputCiphertextLevel != 18 || sample.OutputCiphertextScale != math.Pow(2, 32) || sample.DecodedSlots != 1<<15 || sample.FiniteDecodedSlots != 1<<15 {
			return fmt.Errorf("sample %d output ciphertext or decoded slots mismatch", index+1)
		}
		if backend == "standard" && (sample.DecodedPlaintextLevel != 18 || sample.DecodedPlaintextScale != math.Pow(2, 32)) {
			return fmt.Errorf("Standard sample %d is not a native Level18 plaintext decode", index+1)
		}
		if backend == "fast" && (sample.DecodedPlaintextLevel != 3 || sample.DecodedPlaintextScale != math.Pow(2, 32)) {
			return fmt.Errorf("Fast sample %d is not the specified Level3 prefix projection", index+1)
		}
	}
	return nil
}

func validateBTS009NoiseEvidence(run *perfmeasure.BTSRunDocument) error {
	var nextOperation uint64 = 1
	for index, artifact := range run.NoiseEvidence {
		if artifact.AttemptIndex != index+1 || artifact.Phase != run.Samples[index].Phase || artifact.Bytes <= 0 || artifact.Bytes > bts009NoiseEvidenceLimit {
			return fmt.Errorf("Noise evidence artifact %d has invalid attempt/phase/size metadata", index+1)
		}
		info, err := os.Stat(artifact.Path)
		if err != nil {
			return fmt.Errorf("stat Noise evidence artifact %d: %w", index+1, err)
		}
		if info.Size() != artifact.Bytes {
			return fmt.Errorf("Noise evidence artifact %d size=%d, manifest=%d", index+1, info.Size(), artifact.Bytes)
		}
		digest, err := sha256File(artifact.Path)
		if err != nil || digest != artifact.SHA256 {
			return fmt.Errorf("Noise evidence artifact %d SHA-256 mismatch: %s (%v)", index+1, digest, err)
		}
		var evidence bts009NoiseCall
		if err := readJSON(artifact.Path, &evidence); err != nil {
			return fmt.Errorf("read Noise evidence artifact %d: %w", index+1, err)
		}
		if evidence.SchemaVersion != "bts-fast-noise-fidelity-009-noise-call.v1" || evidence.Task != bts009Task || evidence.AttemptIndex != index+1 || evidence.Phase != artifact.Phase || evidence.Sigma != 4.004280012097432 || evidence.RootSeed != 20261027 {
			return fmt.Errorf("Noise evidence artifact %d does not match the frozen call identity", index+1)
		}
		if evidence.OperationStart != nextOperation || evidence.OperationEnd < evidence.OperationStart ||
			evidence.OperationEnd != evidence.OperationStart+uint64(len(evidence.Events))-1 ||
			artifact.OperationStart != evidence.OperationStart || artifact.OperationEnd != evidence.OperationEnd ||
			len(evidence.Events) != artifact.EventCount || artifact.CtCtMultiplyCalls != uint64(len(evidence.Events)) || artifact.InjectedPolynomials != uint64(len(evidence.Events))*2 {
			return fmt.Errorf("Noise evidence artifact %d operation range or event count mismatch", index+1)
		}
		for eventIndex, event := range evidence.Events {
			if event.Operation != nextOperation+uint64(eventIndex) || event.Level < 0 || event.Level > 29 || event.Scale == "" || !event.MultiplyCompleted || event.Failure != "" {
				return fmt.Errorf("Noise event %d in attempt %d is incomplete or malformed", eventIndex+1, index+1)
			}
			var scale big.Float
			if _, ok := scale.SetString(event.Scale); !ok || scale.Sign() <= 0 {
				return fmt.Errorf("Noise event %d in attempt %d has an invalid or non-positive Scale", eventIndex+1, index+1)
			}
			for sideName, side := range map[string]bts009NoiseSide{"left": event.Left, "right": event.Right} {
				decoded, decodeErr := hex.DecodeString(side.CoefficientSHA256)
				if side.SampleCount != 1<<16 || side.CoefficientSHA256 == "" || decodeErr != nil || len(decoded) != 32 ||
					!bts009CapacityEvidenceFits(event.Level, side) || side.Capacity.CheckedCoefficients != 1<<16 ||
					side.Seed != bts009ExpectedNoiseSeed(evidence.RootSeed, event.Operation, sideName) {
					return fmt.Errorf("Noise event %d attempt %d %s sample/capacity evidence is invalid", eventIndex+1, index+1, sideName)
				}
			}
			if event.Left.Seed == event.Right.Seed || event.Left.CoefficientSHA256 == event.Right.CoefficientSHA256 {
				return fmt.Errorf("Noise event %d attempt %d did not use independent operand streams", eventIndex+1, index+1)
			}
		}
		nextOperation = evidence.OperationEnd + 1
	}
	return nil
}

func bts009CapacityEvidenceFits(level int, side bts009NoiseSide) bool {
	rows := min(level+1, 4)
	if level < 0 || rows < 1 || !side.Capacity.StrictFit || side.Capacity.Rows != rows {
		return false
	}
	maxAbs, ok := new(big.Int).SetString(side.Capacity.MaxAbs, 10)
	if !ok || maxAbs.Sign() < 0 {
		return false
	}
	prefixProduct, ok := new(big.Int).SetString(side.Capacity.PrefixProduct, 10)
	if !ok || prefixProduct.Cmp(bts009QPrefixProduct(rows)) != 0 {
		return false
	}
	twiceMaxAbs := new(big.Int).Lsh(new(big.Int).Set(maxAbs), 1)
	return twiceMaxAbs.Cmp(prefixProduct) < 0
}

func bts009QPrefixProduct(rows int) *big.Int {
	if rows < 1 || rows > 4 {
		return new(big.Int)
	}
	product := big.NewInt(1)
	for index := 0; index < rows; index++ {
		product.Mul(product, new(big.Int).SetUint64(bts009OriginalQ[index]))
	}
	return product
}

func validateMatchedBTS009Runs(standard, fastOff, fastNoise perfmeasure.BTSRunDocument) error {
	if standard.PrimaryCommit != fastOff.PrimaryCommit || standard.PrimaryCommit != fastNoise.PrimaryCommit ||
		standard.BTSManifestSHA256 != fastOff.BTSManifestSHA256 || standard.BTSManifestSHA256 != fastNoise.BTSManifestSHA256 ||
		standard.MeasurementAdapterSHA256 != fastOff.MeasurementAdapterSHA256 || standard.MeasurementAdapterSHA256 != fastNoise.MeasurementAdapterSHA256 ||
		standard.ProfileSHA256 != fastOff.ProfileSHA256 || standard.ProfileSHA256 != fastNoise.ProfileSHA256 ||
		standard.Circuit != fastOff.Circuit || standard.Circuit != fastNoise.Circuit {
		return errors.New("Primary/source/measurement adapter provenance differs across arms")
	}
	if fastOff.BackendCommit != fastNoise.BackendCommit || fastOff.BinarySHA256 != fastNoise.BinarySHA256 {
		return errors.New("Fast OFF and Fast+Noise must use the exact same source commit and Binary SHA-256")
	}
	for _, candidate := range []perfmeasure.BTSRunDocument{fastOff, fastNoise} {
		if !sameBTS009Parameters(standard.Parameters, candidate.Parameters) ||
			standard.Input.SHA256 != candidate.Input.SHA256 || standard.Input.Slots != candidate.Input.Slots ||
			standard.Input.Level != candidate.Input.Level || standard.Input.Scale != candidate.Input.Scale {
			return errors.New("effective parameters or original input metadata differ across arms")
		}
		a, b := standard.Environment, candidate.Environment
		if a.GoVersion != b.GoVersion || a.OS != b.OS || a.Arch != b.Arch || a.CPU != b.CPU || a.NumCPU != b.NumCPU || a.GOMAXPROCS != b.GOMAXPROCS {
			return errors.New("execution environments are not matched across arms")
		}
	}
	return nil
}

func sameBTS009Parameters(a, b perfmeasure.BTSMeasurementParameters) bool {
	return a.LogN == b.LogN && a.RingN == b.RingN && a.Slots == b.Slots && a.SlotLimbs == b.SlotLimbs &&
		a.Mode == b.Mode && a.Seed == b.Seed && a.Amplitude == b.Amplitude && a.EphemeralH == b.EphemeralH &&
		a.MaxLevel == b.MaxLevel && a.QCount == b.QCount && a.PCount == b.PCount && a.Mod1Level == b.Mod1Level &&
		a.Mod1Scale == b.Mod1Scale && a.Mod1Degree == b.Mod1Degree && a.Mod1PSDepth == b.Mod1PSDepth && a.DoubleAngle == b.DoubleAngle &&
		a.InputLevel == b.InputLevel && a.InputScale == b.InputScale && a.OutputLevel == b.OutputLevel && a.OutputScale == b.OutputScale &&
		equalUint64s(a.QPrimes, b.QPrimes) && equalUint64s(a.PPrimes, b.PPrimes)
}

func summarizeBTS009Arm(name string, run perfmeasure.BTSRunDocument, input []float64) (bts009ArmReport, error) {
	arm := bts009ArmReport{
		Arm: name, Backend: run.Backend, NoiseMode: run.NoiseMode,
		BackendCommit: run.BackendCommit, BinarySHA256: run.BinarySHA256,
		EvaluatorAndBTSBuildNS: run.NewBTSNS, ConstructionPlusColdNS: run.NewBTSNS + run.Samples[0].BootstrapElapsedNS,
		Provenance: btsProvenance(run), NoiseEvidence: append([]perfmeasure.BTSNoiseEvidenceArtifact(nil), run.NoiseEvidence...),
	}
	warmTimes, warmBytes, warmAllocations := make([]int64, 0, 5), make([]float64, 0, 5), make([]float64, 0, 5)
	for index, sample := range run.Samples {
		artifact := perfmeasure.BTSVectorArtifact{
			Path: sample.VectorPath, Encoding: sample.VectorEncoding, SHA256: sample.VectorSHA256,
			Slots: sample.DecodedSlots, FiniteSlots: sample.FiniteDecodedSlots,
		}
		vector, err := readBTSVector(artifact)
		if err != nil {
			return bts009ArmReport{}, fmt.Errorf("read %s sample %d: %w", name, index+1, err)
		}
		quality := numericalmetrics.CompareRealVectorQuality(input, vector)
		if quality.Status != "FINITE" {
			return bts009ArmReport{}, fmt.Errorf("%s sample %d has non-comparable numerical metrics: %s", name, index+1, quality.Reason)
		}
		reportSample := bts009SampleReport{
			AttemptIndex: sample.AttemptIndex, Phase: sample.Phase, BootstrapNS: sample.BootstrapElapsedNS,
			AllocatedBytes: sample.AllocatedBytes, Allocations: sample.Allocations,
			OutputSHA256: sample.VectorSHA256, Quality: quality,
		}
		if run.NoiseMode == "on" {
			reportSample.NoiseEvents = run.NoiseEvidence[index].EventCount
			arm.TotalNoiseEvents += reportSample.NoiseEvents
		}
		if index == 0 {
			arm.Cold = reportSample
			continue
		}
		arm.Warm = append(arm.Warm, reportSample)
		warmTimes = append(warmTimes, sample.BootstrapElapsedNS)
		warmBytes = append(warmBytes, float64(sample.AllocatedBytes))
		warmAllocations = append(warmAllocations, float64(sample.Allocations))
	}
	arm.WarmTime = perfmeasure.SummarizeBTSRepeatedSamples(warmTimes)
	arm.WarmAllocatedBytes = perfmeasure.SummarizeBTSRepeatedValues(warmBytes)
	arm.WarmAllocationCount = perfmeasure.SummarizeBTSRepeatedValues(warmAllocations)
	if len(arm.Warm) != 5 || arm.WarmTime.Count != 5 {
		return bts009ArmReport{}, fmt.Errorf("%s report requires exactly five warm captures", name)
	}
	return arm, nil
}

func renderBTS009Report(report bts009ReportDocument) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# BTS-FAST-NOISE-FIDELITY-009\n\nStatus: `%s`\n\n", report.Status)
	fmt.Fprintf(&output, "Input SHA-256: `%s`; LogN16, E32, original BTS Q30/P5 and identical plaintext oracle. Fast OFF and Fast+Noise use commit `%s` and binary `%s`.\n\n",
		report.InputSHA256, report.FastOff.BackendCommit, report.FastOff.BinarySHA256)
	output.WriteString("## Three-arm comparison\n\n")
	output.WriteString("Times are milliseconds; precision is global `-log2(RMSE)`, with the separate median per-slot precision shown explicitly. Allocation values are mean bytes/count over five warm calls.\n\n")
	output.WriteString("| 指標 | Standard | Fast OFF | Fast+Noise |\n|---|---:|---:|---:|\n")
	writeArm := func(label string) {
		fmt.Fprintf(&output, "| %s cold (ms) | %.3f | %.3f | %.3f |\n", label, float64(report.Standard.Cold.BootstrapNS)/1e6, float64(report.FastOff.Cold.BootstrapNS)/1e6, float64(report.FastNoise.Cold.BootstrapNS)/1e6)
	}
	writeArm("cold")
	fmt.Fprintf(&output, "| warm mean (ms) | %.3f | %.3f | %.3f |\n", report.Standard.WarmTime.Mean/1e6, report.FastOff.WarmTime.Mean/1e6, report.FastNoise.WarmTime.Mean/1e6)
	fmt.Fprintf(&output, "| warm median (ms) | %.3f | %.3f | %.3f |\n", report.Standard.WarmTime.Median/1e6, report.FastOff.WarmTime.Median/1e6, report.FastNoise.WarmTime.Median/1e6)
	fmt.Fprintf(&output, "| warm sample stddev (ms) | %.3f | %.3f | %.3f |\n", report.Standard.WarmTime.SampleStandardDeviation/1e6, report.FastOff.WarmTime.SampleStandardDeviation/1e6, report.FastNoise.WarmTime.SampleStandardDeviation/1e6)
	fmt.Fprintf(&output, "| warm min / max (ms) | %.3f / %.3f | %.3f / %.3f | %.3f / %.3f |\n", report.Standard.WarmTime.Min/1e6, report.Standard.WarmTime.Max/1e6, report.FastOff.WarmTime.Min/1e6, report.FastOff.WarmTime.Max/1e6, report.FastNoise.WarmTime.Min/1e6, report.FastNoise.WarmTime.Max/1e6)
	fmt.Fprintf(&output, "| warm speedup vs Standard | 1.000× | %.3f× | %.3f× |\n", report.FastOff.WarmSpeedupVsStandard, report.FastNoise.WarmSpeedupVsStandard)
	fmt.Fprintf(&output, "| warm allocations mean bytes | %.0f | %.0f | %.0f |\n", report.Standard.WarmAllocatedBytes.Mean, report.FastOff.WarmAllocatedBytes.Mean, report.FastNoise.WarmAllocatedBytes.Mean)
	fmt.Fprintf(&output, "| warm allocations mean count | %.0f | %.0f | %.0f |\n", report.Standard.WarmAllocationCount.Mean, report.FastOff.WarmAllocationCount.Mean, report.FastNoise.WarmAllocationCount.Mean)
	fmt.Fprintf(&output, "| cold allocation bytes / count | %d / %d | %d / %d | %d / %d |\n", report.Standard.Cold.AllocatedBytes, report.Standard.Cold.Allocations, report.FastOff.Cold.AllocatedBytes, report.FastOff.Cold.Allocations, report.FastNoise.Cold.AllocatedBytes, report.FastNoise.Cold.Allocations)
	fmt.Fprintf(&output, "| input-relative RMSE (cold) | %.6e | %.6e | %.6e |\n", report.Standard.Cold.Quality.RMSE, report.FastOff.Cold.Quality.RMSE, report.FastNoise.Cold.Quality.RMSE)
	fmt.Fprintf(&output, "| global RMSE precision bits (cold) | %s | %s | %s |\n", formatPrecision(report.Standard.Cold.Quality.GlobalRMSEPrecisionBits), formatPrecision(report.FastOff.Cold.Quality.GlobalRMSEPrecisionBits), formatPrecision(report.FastNoise.Cold.Quality.GlobalRMSEPrecisionBits))
	fmt.Fprintf(&output, "| median per-slot precision bits (cold) | %s | %s | %s |\n", formatPrecision(report.Standard.Cold.Quality.MedianPerSlotPrecisionBits), formatPrecision(report.FastOff.Cold.Quality.MedianPerSlotPrecisionBits), formatPrecision(report.FastNoise.Cold.Quality.MedianPerSlotPrecisionBits))
	fmt.Fprintf(&output, "| warm global RMSE precision bits (mean across runs) | %.4f | %.4f | %.4f |\n", meanPrecisionBits(report.Standard.Warm), meanPrecisionBits(report.FastOff.Warm), meanPrecisionBits(report.FastNoise.Warm))
	fmt.Fprintf(&output, "| warm median per-slot precision bits (mean across runs) | %.4f | %.4f | %.4f |\n", meanSlotPrecisionBits(report.Standard.Warm), meanSlotPrecisionBits(report.FastOff.Warm), meanSlotPrecisionBits(report.FastNoise.Warm))
	fmt.Fprintf(&output, "| cold MaxAbs / p99 / p99.9 | %.3e / %.3e / %.3e | %.3e / %.3e / %.3e | %.3e / %.3e / %.3e |\n",
		report.Standard.Cold.Quality.MaxAbs, report.Standard.Cold.Quality.P99, report.Standard.Cold.Quality.P999,
		report.FastOff.Cold.Quality.MaxAbs, report.FastOff.Cold.Quality.P99, report.FastOff.Cold.Quality.P999,
		report.FastNoise.Cold.Quality.MaxAbs, report.FastNoise.Cold.Quality.P99, report.FastNoise.Cold.Quality.P999)
	fmt.Fprintf(&output, "| Noise events (all calls) | 0 | 0 | %d |\n\n", report.FastNoise.TotalNoiseEvents)
	fmt.Fprintf(&output, "Evaluator/BTS construction: Standard %.3f s, Fast OFF %.3f s, Fast+Noise %.3f s. Construction plus Cold: %.3f s, %.3f s, %.3f s.\n\n",
		float64(report.Standard.EvaluatorAndBTSBuildNS)/1e9, float64(report.FastOff.EvaluatorAndBTSBuildNS)/1e9, float64(report.FastNoise.EvaluatorAndBTSBuildNS)/1e9,
		float64(report.Standard.ConstructionPlusColdNS)/1e9, float64(report.FastOff.ConstructionPlusColdNS)/1e9, float64(report.FastNoise.ConstructionPlusColdNS)/1e9)
	fmt.Fprintf(&output, "Fast+Noise overhead vs Fast OFF: cold %+d ns (%.4fx); warm median %+0.f ns (%.4fx); warm allocation means %+0.f bytes and %+0.f allocations.\n\n",
		report.NoiseOverhead.ColdExtraNS, report.NoiseOverhead.ColdMultiplier, report.NoiseOverhead.WarmMedianExtraNS, report.NoiseOverhead.WarmMedianMultiplier,
		report.NoiseOverhead.WarmAllocatedBytesMeanExtra, report.NoiseOverhead.WarmAllocationCountMeanExtra)
	if report.PrecisionComparison.Status == "FINITE_GLOBAL_RMSE_PRECISION" {
		fmt.Fprintf(&output, "Across five warm outputs, mean global RMSE precision is Standard %.4f bits, Fast OFF %.4f bits, Fast+Noise %.4f bits. Noise moved closer to Standard: %t (distance %.4f → %.4f bits).\n\n",
			*report.PrecisionComparison.StandardWarmMeanBits, *report.PrecisionComparison.FastOffWarmMeanBits, *report.PrecisionComparison.FastNoiseWarmMeanBits,
			report.PrecisionComparison.NoiseMovedCloserToStandard, *report.PrecisionComparison.FastOffDistanceToStandard, *report.PrecisionComparison.FastNoiseDistanceToStandard)
	} else {
		fmt.Fprintf(&output, "Precision proximity to Standard is not comparable: %s.\n\n", report.PrecisionComparison.Status)
	}
	output.WriteString("## Individual precision and execution captures\n\n| Arm | Phase | Time (ms) | Allocated bytes | Allocations | RMSE | Global bits | Slot-median bits | MaxAbs | p99 | p99.9 | Output SHA-256 | Noise events |\n|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---:|\n")
	for _, arm := range []bts009ArmReport{report.Standard, report.FastOff, report.FastNoise} {
		for _, sample := range append([]bts009SampleReport{arm.Cold}, arm.Warm...) {
			fmt.Fprintf(&output, "| %s | %s | %.3f | %d | %d | %.6e | %s | %s | %.3e | %.3e | %.3e | `%s` | %d |\n",
				arm.Arm, sample.Phase, float64(sample.BootstrapNS)/1e6, sample.AllocatedBytes, sample.Allocations,
				sample.Quality.RMSE, formatPrecision(sample.Quality.GlobalRMSEPrecisionBits), formatPrecision(sample.Quality.MedianPerSlotPrecisionBits),
				sample.Quality.MaxAbs, sample.Quality.P99, sample.Quality.P999, sample.OutputSHA256, sample.NoiseEvents)
		}
	}
	fmt.Fprintf(&output, "\nBTS-008 historical warm median (separate older capture): Standard %.3f ms, Fast %.3f ms, %.3fx. This is context only and is not part of the 009 comparison.\n",
		report.Historical008.StandardWarmMedianMS, report.Historical008.FastWarmMedianMS, report.Historical008.WarmSpeedup)
	return output.String()
}

func formatPrecision(value *float64) string {
	if value == nil {
		return "∞"
	}
	return fmt.Sprintf("%.4f", *value)
}

func meanPrecisionBits(samples []bts009SampleReport) float64 {
	var total float64
	for _, sample := range samples {
		if sample.Quality.GlobalRMSEPrecisionBits == nil {
			return math.Inf(1)
		}
		total += *sample.Quality.GlobalRMSEPrecisionBits
	}
	return total / float64(len(samples))
}

func meanSlotPrecisionBits(samples []bts009SampleReport) float64 {
	var total float64
	for _, sample := range samples {
		if sample.Quality.MedianPerSlotPrecisionBits == nil {
			return math.Inf(1)
		}
		total += *sample.Quality.MedianPerSlotPrecisionBits
	}
	return total / float64(len(samples))
}

func compareBTS009Precision(standard, fastOff, fastNoise []bts009SampleReport) bts009PrecisionComparison {
	standardBits := finiteMeanGlobalPrecision(standard)
	fastOffBits := finiteMeanGlobalPrecision(fastOff)
	fastNoiseBits := finiteMeanGlobalPrecision(fastNoise)
	result := bts009PrecisionComparison{
		StandardWarmMeanBits: standardBits, FastOffWarmMeanBits: fastOffBits, FastNoiseWarmMeanBits: fastNoiseBits,
		Status: "FINITE_GLOBAL_RMSE_PRECISION",
	}
	if standardBits == nil || fastOffBits == nil || fastNoiseBits == nil {
		result.Status = "NOT_COMPARABLE_EXACT_ZERO_RMSE"
		return result
	}
	fastOffDistance, fastNoiseDistance := math.Abs(*fastOffBits-*standardBits), math.Abs(*fastNoiseBits-*standardBits)
	result.FastOffDistanceToStandard, result.FastNoiseDistanceToStandard = &fastOffDistance, &fastNoiseDistance
	result.NoiseMovedCloserToStandard = fastNoiseDistance < fastOffDistance
	return result
}

func finiteMeanGlobalPrecision(samples []bts009SampleReport) *float64 {
	if len(samples) == 0 {
		return nil
	}
	var sum float64
	for _, sample := range samples {
		if sample.Quality.GlobalRMSEPrecisionBits == nil {
			return nil
		}
		sum += *sample.Quality.GlobalRMSEPrecisionBits
	}
	mean := sum / float64(len(samples))
	if math.IsNaN(mean) || math.IsInf(mean, 0) {
		return nil
	}
	return &mean
}

func bts009ExpectedNoiseSeed(root int64, operation uint64, side string) int64 {
	var input [17]byte
	binary.LittleEndian.PutUint64(input[0:8], uint64(root))
	binary.LittleEndian.PutUint64(input[8:16], operation)
	if side == "left" {
		input[16] = 0
	} else {
		input[16] = 1
	}
	digest := sha256.Sum256(input[:])
	return int64(binary.LittleEndian.Uint64(digest[:8]))
}
