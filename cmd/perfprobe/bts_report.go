package main

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const bts008ExpectedInputSHA = "5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373"

type btsMetricValue struct {
	Value            float64  `json:"value"`
	Scientific       string   `json:"scientific"`
	Log10            *float64 `json:"log10"`
	OrderOfMagnitude string   `json:"order_of_magnitude"`
}

type btsReportedMetrics struct {
	Count      int            `json:"count"`
	MaxAbs     btsMetricValue `json:"max_abs"`
	MeanAbs    btsMetricValue `json:"mean_abs"`
	RMSE       btsMetricValue `json:"rmse"`
	WorstIndex int            `json:"worst_index"`
}

type btsSampleNumerics struct {
	AttemptIndex    int                `json:"attempt_index"`
	Phase           string             `json:"phase"`
	FastVsStandard  btsReportedMetrics `json:"fast_vs_standard"`
	StandardVsInput btsReportedMetrics `json:"standard_vs_input"`
	FastVsInput     btsReportedMetrics `json:"fast_vs_input"`
}

type btsBackendPerformance struct {
	Backend                string  `json:"backend"`
	NewBTSNS               int64   `json:"new_bts_elapsed_ns"`
	ColdBootstrapNS        int64   `json:"cold_bootstrap_ns"`
	WarmBootstrapNS        []int64 `json:"warm_bootstrap_ns"`
	WarmMedianNS           float64 `json:"warm_median_ns"`
	ColdMilliseconds       float64 `json:"cold_ms"`
	WarmMedianMilliseconds float64 `json:"warm_median_ms"`
}

type btsReportDocument struct {
	SchemaVersion         string                `json:"schema_version"`
	Task                  string                `json:"task"`
	GeneratedAt           time.Time             `json:"generated_at"`
	Status                string                `json:"status"`
	Standard              btsBackendPerformance `json:"standard"`
	Fast                  btsBackendPerformance `json:"fast"`
	FastWarmSpeedup       float64               `json:"fast_warm_speedup"`
	MatchedEnvironment    bool                  `json:"matched_environment"`
	InputVectorSHA256     string                `json:"input_vector_sha256"`
	StandardVectorSHA256s []string              `json:"standard_vector_sha256s"`
	FastVectorSHA256s     []string              `json:"fast_vector_sha256s"`
	StandardMaxError      btsReportedMetrics    `json:"standard_vs_input_all_samples"`
	FastMaxError          btsReportedMetrics    `json:"fast_vs_input_all_samples"`
	FastVsStandard        btsReportedMetrics    `json:"fast_vs_standard_all_samples"`
	Samples               []btsSampleNumerics   `json:"sample_numerics"`
	StandardProvenance    btsRunProvenance      `json:"standard_provenance"`
	FastProvenance        btsRunProvenance      `json:"fast_provenance"`
	PlaintextLevelNote    string                `json:"plaintext_level_note"`
	SecurityCaveat        string                `json:"security_caveat"`
}

type btsRunProvenance struct {
	PrimaryCommit      string `json:"primary_commit"`
	BackendCommit      string `json:"backend_commit"`
	BinarySHA256       string `json:"binary_sha256"`
	BTSManifestSHA256  string `json:"bts_source_manifest_sha256"`
	MeasurementAdapter string `json:"measurement_adapter_sha256"`
	GoVersion          string `json:"go_version"`
	OS                 string `json:"os"`
	Arch               string `json:"arch"`
	CPU                string `json:"cpu"`
	GOMAXPROCS         int    `json:"gomaxprocs"`
}

func runBTSReport(args []string) error {
	flags := flag.NewFlagSet("perfprobe bts-report", flag.ContinueOnError)
	standardPath := flags.String("standard-run", "", "completed Standard Batch008 run manifest")
	fastPath := flags.String("fast-run", "", "completed Fast Batch008 run manifest")
	outJSON := flags.String("out-json", "", "new compact JSON report path")
	outMarkdown := flags.String("out-markdown", "", "new Markdown report path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *standardPath == "" || *fastPath == "" || *outJSON == "" || *outMarkdown == "" {
		return errors.New("--standard-run, --fast-run, --out-json, and --out-markdown are required")
	}
	var standard, fast perfmeasure.BTSRunDocument
	if err := readJSON(*standardPath, &standard); err != nil {
		return fmt.Errorf("read Standard run: %w", err)
	}
	if err := readJSON(*fastPath, &fast); err != nil {
		return fmt.Errorf("read Fast run: %w", err)
	}
	if err := validateBTSRun(&standard, "standard", bts008StandardCommit); err != nil {
		return fmt.Errorf("validate Standard run: %w", err)
	}
	if err := validateBTSRun(&fast, "fast", bts008FastCommit); err != nil {
		return fmt.Errorf("validate Fast run: %w", err)
	}
	if err := validateMatchedBTSRuns(standard, fast); err != nil {
		return err
	}

	standardInput, err := readBTSVector(standard.Input)
	if err != nil {
		return fmt.Errorf("read Standard input: %w", err)
	}
	fastInput, err := readBTSVector(fast.Input)
	if err != nil {
		return fmt.Errorf("read Fast input: %w", err)
	}
	if !equalFloatVectors(standardInput, fastInput) {
		return errors.New("Standard and Fast original input vectors differ")
	}
	allStandard, allFast, allInput := make([]float64, 0, 6*32768), make([]float64, 0, 6*32768), make([]float64, 0, 6*32768)
	standardHashes, fastHashes := make([]string, 0, 6), make([]string, 0, 6)
	comparison := btsReportDocument{
		SchemaVersion: "bts-notrace-warm-and-numerical-008-report.v1",
		Task:          "BTS-NOTRACE-WARM-AND-NUMERICAL-008", GeneratedAt: time.Now().UTC(), Status: "TABLE_COMPLETE_AND_OUTPUTS_COMPARED",
		Standard: summarizeBTSPerformance(standard), Fast: summarizeBTSPerformance(fast),
		MatchedEnvironment: true, InputVectorSHA256: standard.Input.SHA256,
		StandardProvenance: btsProvenance(standard), FastProvenance: btsProvenance(fast),
		PlaintextLevelNote: "Genuine Standard native plaintext remains at logical output Level 18; Fast zero-secret Q-prefix projection decodes at plaintext Level 3 from a logical Level 18 ciphertext. These are intentionally distinct observation semantics.",
		SecurityCaveat:     "Fast is the intentionally insecure zero-secret experimental backend; numerical proximity is not security or full-RNS semantic equivalence.",
	}
	comparison.FastWarmSpeedup = comparison.Standard.WarmMedianNS / comparison.Fast.WarmMedianNS
	for index := 0; index < 6; index++ {
		standardSample, fastSample := standard.Samples[index], fast.Samples[index]
		standardVector, err := readBTSVector(perfmeasure.BTSVectorArtifact{Path: standardSample.VectorPath, Encoding: standardSample.VectorEncoding, SHA256: standardSample.VectorSHA256, Slots: standardSample.DecodedSlots, FiniteSlots: standardSample.FiniteDecodedSlots})
		if err != nil {
			return fmt.Errorf("read Standard sample %d: %w", index+1, err)
		}
		fastVector, err := readBTSVector(perfmeasure.BTSVectorArtifact{Path: fastSample.VectorPath, Encoding: fastSample.VectorEncoding, SHA256: fastSample.VectorSHA256, Slots: fastSample.DecodedSlots, FiniteSlots: fastSample.FiniteDecodedSlots})
		if err != nil {
			return fmt.Errorf("read Fast sample %d: %w", index+1, err)
		}
		cross := numericalmetrics.CompareRealVectors(standardVector, fastVector)
		standardError := numericalmetrics.CompareRealVectors(standardInput, standardVector)
		fastError := numericalmetrics.CompareRealVectors(fastInput, fastVector)
		if cross.Status != "FINITE" || standardError.Status != "FINITE" || fastError.Status != "FINITE" {
			return fmt.Errorf("sample %d produced non-comparable numeric metrics: cross=%s standard=%s fast=%s", index+1, cross.Status, standardError.Status, fastError.Status)
		}
		comparison.Samples = append(comparison.Samples, btsSampleNumerics{
			AttemptIndex: index + 1, Phase: standardSample.Phase,
			FastVsStandard: reportMetrics(cross), StandardVsInput: reportMetrics(standardError), FastVsInput: reportMetrics(fastError),
		})
		allStandard = append(allStandard, standardVector...)
		allFast = append(allFast, fastVector...)
		allInput = append(allInput, standardInput...)
		standardHashes = append(standardHashes, standardSample.VectorSHA256)
		fastHashes = append(fastHashes, fastSample.VectorSHA256)
	}
	comparison.StandardVectorSHA256s, comparison.FastVectorSHA256s = standardHashes, fastHashes
	comparison.FastVsStandard = reportMetrics(numericalmetrics.CompareRealVectors(allStandard, allFast))
	comparison.StandardMaxError = reportMetrics(numericalmetrics.CompareRealVectors(allInput, allStandard))
	comparison.FastMaxError = reportMetrics(numericalmetrics.CompareRealVectors(allInput, allFast))
	if err := writeExclusiveJSON(*outJSON, comparison); err != nil {
		return err
	}
	if err := writeExclusiveText(*outMarkdown, renderBTSReport(comparison)); err != nil {
		return err
	}
	fmt.Printf("BTS report complete: %s; warm speedup %.4fx\n", *outMarkdown, comparison.FastWarmSpeedup)
	return nil
}

func validateBTSRun(run *perfmeasure.BTSRunDocument, backend, commit string) error {
	if run.SchemaVersion != perfmeasure.BTSNoTrace008SchemaVersion || run.Task != "BTS-NOTRACE-WARM-AND-NUMERICAL-008" || run.Status != "COMPLETE" {
		return fmt.Errorf("schema/task/status mismatch: %q/%q/%q", run.SchemaVersion, run.Task, run.Status)
	}
	if run.Backend != backend || run.BackendCommit != commit || run.BinarySHA256 == "" || run.MeasurementAdapterSHA256 == "" || run.PrimaryCommit == "" {
		return errors.New("backend, pinned provenance, or binary identity mismatch")
	}
	if run.BTSManifestSHA256 != bts008SourceManifestSHA || run.MeasurementProcesses != 1 || run.BootstrapBudget != 6 || run.BootstrapReservations != 6 || len(run.Samples) != 6 {
		return errors.New("source manifest, process count, six-call budget, or sample count mismatch")
	}
	p := run.Parameters
	if p.LogN != 16 || p.RingN != 1<<16 || p.Slots != 1<<15 || p.SlotLimbs != 15 || p.Mode != "BTSModePlain" || p.Seed != 0 || p.Amplitude != 1 || p.EphemeralH != 32 || p.InputLevel != 3 || p.InputScale != math.Pow(2, 32) || p.OutputLevel != 18 || p.OutputScale != math.Pow(2, 32) || len(p.QPrimes) == 0 || len(p.PPrimes) == 0 {
		return fmt.Errorf("effective BTS parameters do not match frozen profile: %+v", p)
	}
	if run.Input.Encoding != "IEEE754-f64-little-endian" || run.Input.SHA256 != bts008ExpectedInputSHA || run.Input.Slots != 1<<15 || run.Input.FiniteSlots != 1<<15 || run.Input.Level != 3 || run.Input.Scale != math.Pow(2, 32) {
		return errors.New("original input identity/shape/metadata mismatch")
	}
	for index, sample := range run.Samples {
		wantPhase := fmt.Sprintf("warm_bootstrap_%02d", index)
		if index == 0 {
			wantPhase = "first_cold_bootstrap"
		}
		if sample.AttemptIndex != index+1 || sample.Phase != wantPhase || sample.BootstrapElapsedNS <= 0 || sample.InputLevel != 3 || sample.InputScale != math.Pow(2, 32) {
			return fmt.Errorf("sample %d has invalid phase, timing, or input metadata", index+1)
		}
		if sample.OutputCiphertextLevel != 18 || sample.OutputCiphertextScale != math.Pow(2, 32) || sample.DecodedSlots != 1<<15 || sample.FiniteDecodedSlots != 1<<15 {
			return fmt.Errorf("sample %d ciphertext output or decoded slot metadata mismatch", index+1)
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

func validateMatchedBTSRuns(standard, fast perfmeasure.BTSRunDocument) error {
	if standard.PrimaryCommit != fast.PrimaryCommit || standard.BTSManifestSHA256 != fast.BTSManifestSHA256 || standard.MeasurementAdapterSHA256 != fast.MeasurementAdapterSHA256 {
		return errors.New("Primary/source/adapter provenance differs between Standard and Fast")
	}
	if standard.Parameters.LogN != fast.Parameters.LogN || standard.Parameters.RingN != fast.Parameters.RingN || standard.Parameters.Slots != fast.Parameters.Slots || standard.Parameters.SlotLimbs != fast.Parameters.SlotLimbs || standard.Parameters.Mode != fast.Parameters.Mode || standard.Parameters.Seed != fast.Parameters.Seed || standard.Parameters.Amplitude != fast.Parameters.Amplitude || standard.Parameters.EphemeralH != fast.Parameters.EphemeralH || standard.Parameters.InputLevel != fast.Parameters.InputLevel || standard.Parameters.InputScale != fast.Parameters.InputScale || standard.Parameters.OutputLevel != fast.Parameters.OutputLevel || standard.Parameters.OutputScale != fast.Parameters.OutputScale || !equalUint64s(standard.Parameters.QPrimes, fast.Parameters.QPrimes) || !equalUint64s(standard.Parameters.PPrimes, fast.Parameters.PPrimes) {
		return errors.New("effective Standard and Fast Q/P/profile parameters differ")
	}
	if standard.Input.SHA256 != fast.Input.SHA256 || standard.Input.Slots != fast.Input.Slots || standard.Input.Level != fast.Input.Level || standard.Input.Scale != fast.Input.Scale {
		return errors.New("Standard and Fast input identities or metadata differ")
	}
	a, b := standard.Environment, fast.Environment
	if a.GoVersion != b.GoVersion || a.OS != b.OS || a.Arch != b.Arch || a.CPU != b.CPU || a.NumCPU != b.NumCPU || a.GOMAXPROCS != b.GOMAXPROCS {
		return errors.New("Standard and Fast execution environments are not matched")
	}
	return nil
}

func summarizeBTSPerformance(run perfmeasure.BTSRunDocument) btsBackendPerformance {
	warm := make([]int64, 0, 5)
	for _, sample := range run.Samples[1:] {
		warm = append(warm, sample.BootstrapElapsedNS)
	}
	ordered := append([]int64(nil), warm...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	median := float64(ordered[len(ordered)/2])
	return btsBackendPerformance{
		Backend: run.Backend, NewBTSNS: run.NewBTSNS,
		ColdBootstrapNS: run.Samples[0].BootstrapElapsedNS, WarmBootstrapNS: warm,
		WarmMedianNS: median, ColdMilliseconds: float64(run.Samples[0].BootstrapElapsedNS) / 1e6,
		WarmMedianMilliseconds: median / 1e6,
	}
}

func btsProvenance(run perfmeasure.BTSRunDocument) btsRunProvenance {
	return btsRunProvenance{
		PrimaryCommit: run.PrimaryCommit, BackendCommit: run.BackendCommit,
		BinarySHA256: run.BinarySHA256, BTSManifestSHA256: run.BTSManifestSHA256,
		MeasurementAdapter: run.MeasurementAdapterSHA256, GoVersion: run.Environment.GoVersion,
		OS: run.Environment.OS, Arch: run.Environment.Arch, CPU: run.Environment.CPU,
		GOMAXPROCS: run.Environment.GOMAXPROCS,
	}
}

func readBTSVector(artifact perfmeasure.BTSVectorArtifact) ([]float64, error) {
	if artifact.Encoding != "IEEE754-f64-little-endian" || artifact.Slots <= 0 || artifact.FiniteSlots != artifact.Slots {
		return nil, errors.New("unsupported vector encoding or invalid slot metadata")
	}
	data, err := os.ReadFile(artifact.Path)
	if err != nil {
		return nil, err
	}
	if len(data) != artifact.Slots*8 {
		return nil, fmt.Errorf("vector byte length=%d, expected=%d", len(data), artifact.Slots*8)
	}
	if digest := digestBytes(data); digest != artifact.SHA256 {
		return nil, fmt.Errorf("vector SHA-256=%s, metadata says %s", digest, artifact.SHA256)
	}
	values := make([]float64, artifact.Slots)
	for index := range values {
		values[index] = math.Float64frombits(binary.LittleEndian.Uint64(data[index*8:]))
		if math.IsNaN(values[index]) || math.IsInf(values[index], 0) {
			return nil, fmt.Errorf("vector contains non-finite slot %d", index)
		}
	}
	return values, nil
}

func reportMetrics(metric numericalmetrics.RealVectorError) btsReportedMetrics {
	return btsReportedMetrics{
		Count: metric.Count, MaxAbs: magnitude(metric.MaxAbs), MeanAbs: magnitude(metric.MeanAbs),
		RMSE: magnitude(metric.RMSE), WorstIndex: metric.WorstIndex,
	}
}

func magnitude(value float64) btsMetricValue {
	result := btsMetricValue{Value: value, Scientific: fmt.Sprintf("%.8e", value)}
	if value == 0 {
		result.OrderOfMagnitude = "ZERO"
		return result
	}
	log10 := math.Log10(value)
	result.Log10 = &log10
	result.OrderOfMagnitude = fmt.Sprintf("10^%d", int(math.Floor(log10)))
	return result
}

func renderBTSReport(report btsReportDocument) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# BTS-NOTRACE-WARM-AND-NUMERICAL-008 — measured results\n\nStatus: `%s`\n\n", report.Status)
	output.WriteString("| 環維度 | CKKS slots | 實作 | Cold (ms) | Warm 中位數 (ms) | Warm 倍率 |\n|---|---:|---|---:|---:|---:|\n")
	fmt.Fprintf(&output, "| LogN16 (N=65,536) | 32,768 | Standard | %.3f | %.3f | 1× |\n", report.Standard.ColdMilliseconds, report.Standard.WarmMedianMilliseconds)
	fmt.Fprintf(&output, "| LogN16 (N=65,536) | 32,768 | Fast | %.3f | %.3f | %.4f× |\n\n", report.Fast.ColdMilliseconds, report.Fast.WarmMedianMilliseconds, report.FastWarmSpeedup)
	fmt.Fprintf(&output, "NewBTS initialization (separate from Bootstrap): Standard %.6f s; Fast %.6f s.\n\n", float64(report.Standard.NewBTSNS)/1e9, float64(report.Fast.NewBTSNS)/1e9)
	fmt.Fprintf(&output, "Input SHA-256: `%s`. Warm speedup uses Standard warm median / Fast warm median.\n\n", report.InputVectorSHA256)
	output.WriteString("## Decoded numerical comparison\n\nAll six same-index sample comparisons are included below; index pairing is descriptive because backend key/encryption randomness is independent. Metrics are over 32,768 real slots per sample.\n\n")
	output.WriteString("| Attempt | Phase | Fast vs Standard MaxAbs / RMSE / MeanAbs | Standard vs X MaxAbs | Fast vs X MaxAbs |\n|---:|---|---|---|---|\n")
	for _, sample := range report.Samples {
		fmt.Fprintf(&output, "| %d | %s | %s / %s / %s | %s | %s |\n",
			sample.AttemptIndex, sample.Phase,
			sample.FastVsStandard.MaxAbs.Scientific, sample.FastVsStandard.RMSE.Scientific, sample.FastVsStandard.MeanAbs.Scientific,
			sample.StandardVsInput.MaxAbs.Scientific, sample.FastVsInput.MaxAbs.Scientific)
	}
	output.WriteString("\n### Aggregate over the six recorded sample vectors\n\n")
	fmt.Fprintf(&output, "- Fast vs Standard: MaxAbs %s (log10 %s; %s), RMSE %s (log10 %s; %s), MeanAbs %s (log10 %s; %s), worst slot %d of the concatenated six-vector set.\n",
		report.FastVsStandard.MaxAbs.Scientific, logText(report.FastVsStandard.MaxAbs), report.FastVsStandard.MaxAbs.OrderOfMagnitude,
		report.FastVsStandard.RMSE.Scientific, logText(report.FastVsStandard.RMSE), report.FastVsStandard.RMSE.OrderOfMagnitude,
		report.FastVsStandard.MeanAbs.Scientific, logText(report.FastVsStandard.MeanAbs), report.FastVsStandard.MeanAbs.OrderOfMagnitude,
		report.FastVsStandard.WorstIndex)
	fmt.Fprintf(&output, "- Standard vs X: MaxAbs %s (log10 %s; %s), RMSE %s (log10 %s; %s), MeanAbs %s (log10 %s; %s).\n",
		report.StandardMaxError.MaxAbs.Scientific, logText(report.StandardMaxError.MaxAbs), report.StandardMaxError.MaxAbs.OrderOfMagnitude,
		report.StandardMaxError.RMSE.Scientific, logText(report.StandardMaxError.RMSE), report.StandardMaxError.RMSE.OrderOfMagnitude,
		report.StandardMaxError.MeanAbs.Scientific, logText(report.StandardMaxError.MeanAbs), report.StandardMaxError.MeanAbs.OrderOfMagnitude)
	fmt.Fprintf(&output, "- Fast vs X: MaxAbs %s (log10 %s; %s), RMSE %s (log10 %s; %s), MeanAbs %s (log10 %s; %s).\n\n",
		report.FastMaxError.MaxAbs.Scientific, logText(report.FastMaxError.MaxAbs), report.FastMaxError.MaxAbs.OrderOfMagnitude,
		report.FastMaxError.RMSE.Scientific, logText(report.FastMaxError.RMSE), report.FastMaxError.RMSE.OrderOfMagnitude,
		report.FastMaxError.MeanAbs.Scientific, logText(report.FastMaxError.MeanAbs), report.FastMaxError.MeanAbs.OrderOfMagnitude)
	output.WriteString("Fast plaintext Level 3 is a Q-prefix projection of the logical Level 18 output; Standard plaintext is native Level 18. Fast remains intentionally insecure zero-secret simulation; numerical similarity is not security equivalence.\n")
	return output.String()
}

func logText(value btsMetricValue) string {
	if value.Log10 == nil {
		return "N/A (zero)"
	}
	return fmt.Sprintf("%.6f", *value.Log10)
}

func equalUint64s(left, right []uint64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalFloatVectors(left, right []float64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if math.Float64bits(left[index]) != math.Float64bits(right[index]) {
			return false
		}
	}
	return true
}

func digestBytes(data []byte) string {
	sum := sha256Sum(data)
	return fmt.Sprintf("%x", sum)
}

func sha256Sum(data []byte) [32]byte { return sha256.Sum256(data) }
