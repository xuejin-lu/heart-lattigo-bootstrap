package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	batch025RawSHA256             = "f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74"
	batch025RawSize               = 177696
	batch025SidecarSHA256         = "59a872b2df18ca32859f4753650c3f1e14633dd79d1358b283204c4445b3f907"
	batch025VectorsSHA256         = "63ad89ce9ccc70a0b737397810d4f4a15d8a1ed65e82ccd3af0435d1df4d6881"
	batch025ManifestSHA256        = "5bc19b1eb6e67997ca3d19f39e6fd781bf93b27ad305a1be733f0c883611b847"
	batch025CapturedPrimaryCommit = "4d77c511b27ea303fc5a65870dff1404011d1807"
	batch025DiagnosticCommit      = "4f2557062cb5c1ffb9a671bc3df67401fd7092b4"
)

type publicE32OfflineOptions struct {
	rawPath      string
	sidecarPath  string
	manifestPath string
	vectorsPath  string
	attemptOne   string
	attemptTwo   string
	outPrefix    string
}

type publicE32AttemptJournal struct {
	SchemaVersion   string `json:"schema_version"`
	AttemptIndex    int    `json:"attempt_index"`
	BootstrapBudget int    `json:"bootstrap_budget"`
	Phase           string `json:"phase"`
	Status          string `json:"status"`
	ReservedAt      string `json:"reserved_at"`
}

type publicE32OfflineAttempt struct {
	Path            string `json:"path"`
	SHA256          string `json:"sha256"`
	AttemptIndex    int    `json:"attempt_index"`
	BootstrapBudget int    `json:"bootstrap_budget"`
	Phase           string `json:"phase"`
	Status          string `json:"status"`
	ReservedAt      string `json:"reserved_at"`
}

type publicE32OfflineOutput struct {
	Phase                    string  `json:"phase"`
	VectorName               string  `json:"matched_vector_name"`
	DecodedSHA256            string  `json:"recorded_decoded_sha256"`
	MatchedVectorSHA256      string  `json:"recomputed_vector_sha256"`
	DecodedHashMatches       bool    `json:"decoded_hash_matches_vector"`
	ElapsedNS                int64   `json:"recorded_elapsed_ns"`
	Level                    int     `json:"level"`
	Degree                   int     `json:"degree"`
	Scale                    string  `json:"scale"`
	QPrefixRows              int     `json:"q_prefix_rows"`
	OracleRMSE               float64 `json:"recorded_oracle_rmse"`
	OracleMaxComplex         float64 `json:"recorded_oracle_max_complex_difference"`
	FastReferenceWorstRMSE   float64 `json:"recorded_fast_reference_worst_rmse"`
	FastReferenceWorstMax    float64 `json:"recorded_fast_reference_worst_max_complex_difference"`
	NumericalMetadataPass    bool    `json:"recorded_numerical_metadata_passes_gate"`
	NativeDecryptReperformed bool    `json:"native_decrypt_reperformed"`
}

type publicE32OfflineProfile struct {
	FileName               string `json:"file_name,omitempty"`
	RecordedSHA256         string `json:"raw_recorded_sha256,omitempty"`
	ObservedSHA256         string `json:"available_file_sha256,omitempty"`
	FileAvailable          bool   `json:"file_available"`
	HashBoundByRawMetadata bool   `json:"hash_bound_by_raw_metadata"`
	UsedForAttribution     bool   `json:"used_for_attribution"`
	Disposition            string `json:"disposition"`
}

type publicE32OfflineRootClosure struct {
	Sequence       uint64  `json:"root_sequence"`
	Scope          string  `json:"root_scope"`
	Name           string  `json:"root_name"`
	EventCount     int     `json:"tree_event_count"`
	InclusiveNS    float64 `json:"root_inclusive_ns"`
	ChildrenNS     float64 `json:"immediate_children_ns"`
	SignedResidual float64 `json:"signed_unattributed_ns"`
}

type publicE32OfflineDocument struct {
	SchemaVersion             string                        `json:"schema_version"`
	GeneratedAt               time.Time                     `json:"generated_at"`
	CapturedAt                time.Time                     `json:"captured_at"`
	CapturedGoVersion         string                        `json:"captured_go_version"`
	CapturedOS                string                        `json:"captured_os"`
	CapturedArch              string                        `json:"captured_arch"`
	CapturedNumCPU            int                           `json:"captured_num_cpu"`
	CapturedGOMAXPROCS        int                           `json:"captured_gomaxprocs"`
	Classification            string                        `json:"classification"`
	RawTraceStatus            string                        `json:"original_raw_trace_status"`
	OriginalRawRelabeled      bool                          `json:"original_raw_relabelled"`
	NewBootstrapCalls         int                           `json:"new_bootstrap_calls"`
	RawPath                   string                        `json:"raw_path"`
	RawSHA256                 string                        `json:"raw_sha256"`
	RawBytes                  int                           `json:"raw_bytes"`
	SidecarPath               string                        `json:"failure_sidecar_path"`
	SidecarSHA256             string                        `json:"failure_sidecar_sha256"`
	ManifestPath              string                        `json:"fixture_manifest_path"`
	ManifestSHA256            string                        `json:"fixture_manifest_sha256"`
	VectorsPath               string                        `json:"fast_vectors_path"`
	VectorsSHA256             string                        `json:"fast_vectors_sha256"`
	Profile                   string                        `json:"profile"`
	Mode                      string                        `json:"mode"`
	CapturedPrimaryCommit     string                        `json:"captured_primary_commit"`
	FixturePrimaryCommit      string                        `json:"fixture_primary_commit"`
	PrimarySourceSHA256       string                        `json:"primary_measurement_source_sha256"`
	ProductionFastCommit      string                        `json:"production_fast_commit"`
	DiagnosticSecondaryCommit string                        `json:"diagnostic_secondary_commit"`
	PrimaryAnalyzer           RepositoryMetadata            `json:"primary_analyzer_repository"`
	SecondarySource           RepositoryMetadata            `json:"secondary_source_repository"`
	ProductionSourceSHA256    string                        `json:"production_go_source_sha256,omitempty"`
	DiagnosticSourceSHA256    string                        `json:"diagnostic_go_source_sha256,omitempty"`
	ChangedProductionPaths    []string                      `json:"diagnostic_production_source_diff_paths,omitempty"`
	InstrumentationOnlyDelta  bool                          `json:"authorized_instrumentation_only_delta"`
	ConfigSHA256              string                        `json:"config_sha256"`
	QPSHA256                  string                        `json:"qp_sha256"`
	InputSHA256               string                        `json:"input_sha256"`
	WorkloadSHA256            string                        `json:"workload_sha256"`
	CiphertextSHA256          string                        `json:"ciphertext_sha256"`
	CallBudget                int                           `json:"captured_call_budget"`
	ActualCalls               int                           `json:"captured_actual_calls"`
	NumericalGate             float64                       `json:"captured_max_complex_gate"`
	Attempts                  []publicE32OfflineAttempt     `json:"attempt_journals"`
	Outputs                   []publicE32OfflineOutput      `json:"captured_output_checks"`
	CPUProfile                publicE32OfflineProfile       `json:"cpu_profile_evidence"`
	HeapProfile               publicE32OfflineProfile       `json:"heap_profile_evidence"`
	EventCount                int                           `json:"event_count"`
	StageCounts               map[string]int                `json:"stage_counts"`
	IndependentRootCount      int                           `json:"independent_root_count"`
	NegativeExclusiveCount    int                           `json:"negative_exclusive_event_count"`
	EventMedians              []EventMedian                 `json:"event_type_medians_descriptive_only"`
	RootClosures              []publicE32OfflineRootClosure `json:"independent_root_closures"`
	Analysis                  *publicE32TraceAnalysis       `json:"root_partitioned_event_analysis,omitempty"`
	ValidationAnomalies       []string                      `json:"validation_anomalies,omitempty"`
	Warnings                  []string                      `json:"warnings,omitempty"`
}

func parsePublicE32OfflineArgs(args []string) (publicE32OfflineOptions, error) {
	flags := flag.NewFlagSet("fastdiag offline-e32", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	opts := publicE32OfflineOptions{}
	flags.StringVar(&opts.rawPath, "raw", "", "凍結的 TRACE_UNVERIFIED raw JSON")
	flags.StringVar(&opts.sidecarPath, "sidecar", "", "failure sidecar JSON")
	flags.StringVar(&opts.manifestPath, "manifest", "", "E32 fixture manifest JSON")
	flags.StringVar(&opts.vectorsPath, "vectors", "", "凍結 Fast repeatability vectors JSON")
	flags.StringVar(&opts.attemptOne, "attempt-01", "", "Batch024 attempt-01 journal JSON")
	flags.StringVar(&opts.attemptTwo, "attempt-02", "", "Batch024 attempt-02 journal JSON")
	flags.StringVar(&opts.outPrefix, "out-prefix", "", "結果檔前綴（產生 -summary.md、-journal.md、-evidence.json）")
	if err := flags.Parse(args); err != nil {
		return publicE32OfflineOptions{}, err
	}
	if flags.NArg() != 0 {
		return publicE32OfflineOptions{}, fmt.Errorf("offline-e32 不接受位置參數：%s", strings.Join(flags.Args(), " "))
	}
	for name, value := range map[string]string{
		"--raw": opts.rawPath, "--sidecar": opts.sidecarPath, "--manifest": opts.manifestPath,
		"--vectors": opts.vectorsPath, "--attempt-01": opts.attemptOne, "--attempt-02": opts.attemptTwo,
		"--out-prefix": opts.outPrefix,
	} {
		if strings.TrimSpace(value) == "" {
			return publicE32OfflineOptions{}, fmt.Errorf("offline-e32 需要 %s", name)
		}
	}
	if filepath.Clean(opts.attemptOne) == filepath.Clean(opts.attemptTwo) {
		return publicE32OfflineOptions{}, errors.New("attempt-01 與 attempt-02 必須是不同檔案")
	}
	return opts, nil
}

func executePublicE32Offline(args []string) error {
	opts, err := parsePublicE32OfflineArgs(args)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	primaryRoot, err := findPrimaryRoot(cwd)
	if err != nil {
		return err
	}
	if err := verifyRepository(primaryRoot, "heart-lattigo-bootstrap"); err != nil {
		return fmt.Errorf("Primary repo 驗證失敗：%w", err)
	}
	primaryBranch, err := gitOutput(primaryRoot, "branch", "--show-current")
	if err != nil {
		return err
	}
	if primaryBranch != "main" {
		return fmt.Errorf("offline-e32 要求 Primary main，目前為 %q", primaryBranch)
	}
	if err := requireCleanWorktree(primaryRoot); err != nil {
		return fmt.Errorf("offline-e32 要求乾淨 Primary worktree：%w", err)
	}
	secondaryRoot, err := resolveSecondaryRoot(primaryRoot)
	if err != nil {
		return err
	}
	if err := verifyRepository(secondaryRoot, "lattigo"); err != nil {
		return fmt.Errorf("Secondary repo 驗證失敗：%w", err)
	}
	paths, err := publicE32OfflineOutputPaths(opts)
	if err != nil {
		return err
	}
	primaryMeta, err := repositoryMetadata(primaryRoot, "")
	if err != nil {
		return err
	}
	doc, journal, err := replayPublicE32Offline(primaryRoot, secondaryRoot, primaryMeta, opts)
	if err != nil {
		return err
	}
	if err := writeJSON(paths.evidence, doc); err != nil {
		return fmt.Errorf("寫入離線 evidence 失敗：%w", err)
	}
	if err := writeNewFile(paths.summary, []byte(renderPublicE32OfflineSummary(doc)), 0o644); err != nil {
		return fmt.Errorf("寫入離線 summary 失敗：%w", err)
	}
	if err := writeNewFile(paths.journal, []byte(journal), 0o644); err != nil {
		return fmt.Errorf("寫入離線 journal 失敗：%w", err)
	}
	fmt.Printf("分類：%s\nSummary：%s\nJournal：%s\nEvidence：%s\n", doc.Classification, paths.summary, paths.journal, paths.evidence)
	return nil
}

type publicE32OfflinePaths struct {
	summary  string
	journal  string
	evidence string
}

func publicE32OfflineOutputPaths(opts publicE32OfflineOptions) (publicE32OfflinePaths, error) {
	prefix, err := filepath.Abs(opts.outPrefix)
	if err != nil {
		return publicE32OfflinePaths{}, err
	}
	if filepath.Ext(prefix) != "" {
		return publicE32OfflinePaths{}, errors.New("--out-prefix 不可含副檔名")
	}
	paths := publicE32OfflinePaths{
		summary: prefix + "-summary.md", journal: prefix + "-journal.md", evidence: prefix + "-evidence.json",
	}
	inputs := []string{opts.rawPath, opts.sidecarPath, opts.manifestPath, opts.vectorsPath, opts.attemptOne, opts.attemptTwo}
	for _, output := range []string{paths.summary, paths.journal, paths.evidence} {
		absOutput, _ := filepath.Abs(output)
		for _, input := range inputs {
			absInput, _ := filepath.Abs(input)
			if filepath.Clean(absOutput) == filepath.Clean(absInput) {
				return publicE32OfflinePaths{}, fmt.Errorf("輸出路徑與輸入證據重疊：%s", output)
			}
		}
		if _, err := os.Stat(output); err == nil {
			return publicE32OfflinePaths{}, fmt.Errorf("拒絕覆寫既有離線輸出：%s", output)
		} else if !errors.Is(err, os.ErrNotExist) {
			return publicE32OfflinePaths{}, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(paths.evidence), 0o755); err != nil {
		return publicE32OfflinePaths{}, err
	}
	return paths, nil
}

func replayPublicE32Offline(primaryRoot, secondaryRoot string, primaryMeta RepositoryMetadata, opts publicE32OfflineOptions) (publicE32OfflineDocument, string, error) {
	rawBytes, rawSHA, err := readPublicE32OfflineRaw(opts.rawPath)
	if err != nil {
		return publicE32OfflineDocument{}, "", err
	}

	var raw publicE32RawTrace
	if err := json.Unmarshal(rawBytes, &raw); err != nil {
		return publicE32OfflineDocument{}, "", fmt.Errorf("decode frozen raw trace: %w", err)
	}
	doc := publicE32OfflineDocument{
		SchemaVersion: "fastdiag.public-e32.offline-salvage.v1", GeneratedAt: time.Now().UTC(),
		CapturedAt: raw.Timestamp, CapturedGoVersion: raw.GoVersion, CapturedOS: raw.OS,
		CapturedArch: raw.Arch, CapturedNumCPU: raw.NumCPU, CapturedGOMAXPROCS: raw.GOMAXPROCS,
		Classification: "OFFLINE_PARTIAL_UNVERIFIED", RawTraceStatus: raw.TraceStatus,
		OriginalRawRelabeled: false, NewBootstrapCalls: 0,
		RawPath: opts.rawPath, RawSHA256: rawSHA, RawBytes: len(rawBytes),
		Profile: raw.Profile, Mode: raw.Mode, CapturedPrimaryCommit: raw.PrimaryCommit,
		FixturePrimaryCommit: raw.FixturePrimaryCommit, PrimarySourceSHA256: raw.PrimarySourceSHA256,
		ProductionFastCommit: raw.ProductionFastCommit, DiagnosticSecondaryCommit: raw.DiagnosticHead,
		ConfigSHA256: raw.ConfigSHA256, QPSHA256: raw.QPSHA256, InputSHA256: raw.InputSHA256,
		WorkloadSHA256: raw.WorkloadSHA256, CiphertextSHA256: raw.CiphertextSHA256,
		PrimaryAnalyzer: primaryMeta, CallBudget: raw.CallBudget, ActualCalls: raw.ActualCalls,
		NumericalGate: raw.NumericalGate, EventCount: len(raw.Events), StageCounts: make(map[string]int),
		Attempts: make([]publicE32OfflineAttempt, 0, 2), Outputs: make([]publicE32OfflineOutput, 0, len(raw.Runs)),
		EventMedians: aggregateSingleTraceEventTypes(raw.Events), Warnings: []string{
			"OFFLINE_REVALIDATED_TRACE is a separate analysis classification; the immutable source raw remains TRACE_UNVERIFIED.",
			"Recorded numerical metrics are metadata only; this offline replay performs no native decrypt, CKKS operation, key generation, or Bootstrap.",
			"Event-type medians are descriptive only. Shares and Pareto are partitioned by independent event-tree root.",
		},
	}
	for _, event := range raw.Events {
		if event.Scope == "stage" && event.Name != "bootstrap" {
			doc.StageCounts[event.Name]++
		}
	}
	addAnomaly := func(format string, args ...any) {
		doc.ValidationAnomalies = append(doc.ValidationAnomalies, fmt.Sprintf(format, args...))
	}

	sidecarBytes, sidecarSHA, sidecarErr := readFileSHA256(opts.sidecarPath)
	if sidecarErr != nil {
		addAnomaly("failure sidecar unavailable: %v", sidecarErr)
	} else {
		doc.SidecarPath, doc.SidecarSHA256 = opts.sidecarPath, sidecarSHA
		if sidecarSHA != batch025SidecarSHA256 {
			addAnomaly("failure sidecar SHA-256 %s does not match frozen value", sidecarSHA)
		}
		var sidecar publicE32FailureArtifact
		if err := json.Unmarshal(sidecarBytes, &sidecar); err != nil {
			addAnomaly("failure sidecar JSON is invalid: %v", err)
		} else {
			validatePublicE32OfflineSidecar(raw, sidecar, rawSHA, addAnomaly)
		}
	}

	manifestBytes, manifestSHA, manifestErr := readFileSHA256(opts.manifestPath)
	if manifestErr != nil {
		addAnomaly("fixture manifest unavailable: %v", manifestErr)
	} else {
		doc.ManifestPath, doc.ManifestSHA256 = opts.manifestPath, manifestSHA
		if manifestSHA != batch025ManifestSHA256 {
			addAnomaly("fixture manifest SHA-256 %s does not match frozen value", manifestSHA)
		}
	}
	vectorsBytes, vectorsSHA, vectorsErr := readFileSHA256(opts.vectorsPath)
	if vectorsErr != nil {
		addAnomaly("Fast vectors unavailable: %v", vectorsErr)
	} else {
		doc.VectorsPath, doc.VectorsSHA256 = opts.vectorsPath, vectorsSHA
		if vectorsSHA != batch025VectorsSHA256 {
			addAnomaly("Fast vectors SHA-256 %s does not match frozen value", vectorsSHA)
		}
	}

	var manifest publicE32FixtureManifest
	var vectors publicE32Vectors
	if len(manifestBytes) > 0 {
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			addAnomaly("fixture manifest JSON is invalid: %v", err)
		}
	}
	if len(vectorsBytes) > 0 {
		if err := json.Unmarshal(vectorsBytes, &vectors); err != nil {
			addAnomaly("Fast vectors JSON is invalid: %v", err)
		}
	}
	if len(manifestBytes) > 0 && len(vectorsBytes) > 0 {
		measurementHash, hashErr := primaryMeasurementSourceFingerprint(primaryRoot, raw.PrimaryCommit)
		if hashErr != nil {
			addAnomaly("captured Primary measurement source fingerprint unavailable: %v", hashErr)
		} else if measurementHash != raw.PrimarySourceSHA256 || measurementHash != manifest.PrimarySourceSHA256 {
			addAnomaly("captured Primary source fingerprint does not match raw/manifest: computed=%s raw=%s manifest=%s", measurementHash, raw.PrimarySourceSHA256, manifest.PrimarySourceSHA256)
		}
		_, _, _, _, fixtureErr := validatePublicE32Fixture(opts.manifestPath, opts.vectorsPath, e32FixturePrimaryCommit, measurementHash)
		if fixtureErr != nil {
			addAnomaly("held E32 fixture consistency check failed: %v", fixtureErr)
		}
		validatePublicE32OfflineProvenance(raw, manifest, vectors, doc.ManifestSHA256, addAnomaly)
		if err := gitRun(primaryRoot, "merge-base", "--is-ancestor", manifest.PrimaryCommit, raw.PrimaryCommit); err != nil {
			addAnomaly("fixture Primary commit %s is not an ancestor of captured Primary commit %s", manifest.PrimaryCommit, raw.PrimaryCommit)
		}
		if err := gitRun(primaryRoot, "merge-base", "--is-ancestor", raw.PrimaryCommit, primaryMeta.Commit); err != nil {
			addAnomaly("captured Primary commit %s is not an ancestor of analyzer commit %s", raw.PrimaryCommit, primaryMeta.Commit)
		}
	}
	validatePublicE32OfflineOutputMetadata(raw, manifest, vectors, &doc, addAnomaly)

	for index, path := range []string{opts.attemptOne, opts.attemptTwo} {
		data, digest, readErr := readFileSHA256(path)
		if readErr != nil {
			addAnomaly("attempt-%02d journal unavailable: %v", index+1, readErr)
			continue
		}
		var attempt publicE32AttemptJournal
		if err := json.Unmarshal(data, &attempt); err != nil {
			addAnomaly("attempt-%02d journal JSON is invalid: %v", index+1, err)
			continue
		}
		doc.Attempts = append(doc.Attempts, publicE32OfflineAttempt{
			Path: path, SHA256: digest, AttemptIndex: attempt.AttemptIndex,
			BootstrapBudget: attempt.BootstrapBudget, Phase: attempt.Phase,
			Status: attempt.Status, ReservedAt: attempt.ReservedAt,
		})
		wantPhase := "diagnostic_fast_cold"
		if index == 1 {
			wantPhase = "diagnostic_fast_traced_warm"
		}
		if attempt.SchemaVersion != "perfprobe-bootstrap-attempt.v1" || attempt.AttemptIndex != index+1 ||
			attempt.BootstrapBudget != 2 || attempt.Phase != wantPhase || attempt.Status != "irrevocably_reserved_before_call" || attempt.ReservedAt == "" {
			addAnomaly("attempt-%02d journal does not match its frozen reservation contract", index+1)
		}
	}
	if len(doc.Attempts) != 2 {
		addAnomaly("found %d/2 required attempt journals", len(doc.Attempts))
	}

	secondaryBranch, branchErr := gitOutput(secondaryRoot, "branch", "--show-current")
	if branchErr != nil {
		addAnomaly("cannot read Secondary branch: %v", branchErr)
	} else if secondaryBranch != "fast-qprefix" {
		addAnomaly("Secondary branch is %q, want fast-qprefix", secondaryBranch)
	}
	doc.SecondarySource, err = repositoryMetadata(secondaryRoot, secondaryBranch)
	if err != nil {
		addAnomaly("cannot read Secondary metadata: %v", err)
		doc.SecondarySource = RepositoryMetadata{}
	} else {
		if doc.SecondarySource.Commit != batch025DiagnosticCommit || raw.DiagnosticHead != doc.SecondarySource.Commit {
			addAnomaly("Secondary source HEAD %s does not match frozen diagnostic commit %s", doc.SecondarySource.Commit, batch025DiagnosticCommit)
		}
		if doc.SecondarySource.Dirty {
			addAnomaly("Secondary source worktree is dirty")
		}
		remoteHead, remoteErr := gitOutput(secondaryRoot, "rev-parse", "origin/fast-qprefix")
		if remoteErr != nil || remoteHead != doc.SecondarySource.Commit {
			addAnomaly("Secondary origin/fast-qprefix does not match checked source HEAD: remote=%s error=%v", remoteHead, remoteErr)
		}
		productionHash, diagnosticHash, changed, instrumentationOnly, sourceErr := verifyE32ProductionSourceIdentity(secondaryRoot, raw.DiagnosticHead)
		if sourceErr != nil {
			addAnomaly("Secondary production-source identity audit failed: %v", sourceErr)
		} else {
			doc.ProductionSourceSHA256, doc.DiagnosticSourceSHA256 = productionHash, diagnosticHash
			doc.ChangedProductionPaths, doc.InstrumentationOnlyDelta = changed, instrumentationOnly
		}
	}

	validatePublicE32OfflineTraceStatus(raw, addAnomaly)
	if eventErr := validatePublicE32Events(raw.Events); eventErr != nil {
		for _, line := range strings.Split(eventErr.Error(), "\n") {
			if strings.TrimSpace(line) != "" {
				doc.ValidationAnomalies = append(doc.ValidationAnomalies, "event structure: "+line)
			}
		}
	}
	if analysis, analysisErr := analyzePublicE32Events(raw.Events); analysisErr != nil {
		addAnomaly("root-partitioned event analysis unavailable: %v", analysisErr)
	} else {
		doc.Analysis = &analysis
		doc.RootClosures = publicE32OfflineRootClosures(raw.Events)
		doc.IndependentRootCount = len(doc.RootClosures)
		for _, closure := range analysis.Closures {
			if closure.UnattributedNS < 0 {
				doc.NegativeExclusiveCount++
			}
		}
	}
	if doc.RawTraceStatus != "TRACE_UNVERIFIED" {
		addAnomaly("original raw status unexpectedly changed to %q", doc.RawTraceStatus)
	}
	if doc.CPUProfile, err = inspectPublicE32OfflineProfile(opts.rawPath, raw.CPUProfileFile, raw.CPUProfileSHA256); err != nil {
		addAnomaly("CPU profile evidence: %v", err)
	}
	if doc.HeapProfile, err = inspectPublicE32OfflineProfile(opts.rawPath, raw.HeapProfileFile, raw.HeapProfileSHA256); err != nil {
		addAnomaly("heap profile evidence: %v", err)
	}
	doc.Classification = publicE32OfflineClassification(doc.RawTraceStatus, len(doc.ValidationAnomalies), doc.Analysis != nil)
	return doc, renderPublicE32OfflineJournal(doc), nil
}

func publicE32OfflineClassification(rawStatus string, anomalyCount int, analysisAvailable bool) string {
	if rawStatus == "TRACE_UNVERIFIED" && anomalyCount == 0 && analysisAvailable {
		return "OFFLINE_REVALIDATED_TRACE"
	}
	return "OFFLINE_PARTIAL_UNVERIFIED"
}

func readPublicE32OfflineRaw(path string) ([]byte, string, error) {
	data, digest, err := readFileSHA256(path)
	if err != nil {
		return nil, "", err
	}
	if digest != batch025RawSHA256 || len(data) != batch025RawSize {
		return nil, digest, fmt.Errorf("frozen raw trace integrity STOP: SHA-256=%s bytes=%d, want SHA-256=%s bytes=%d", digest, len(data), batch025RawSHA256, batch025RawSize)
	}
	return data, digest, nil
}

func validatePublicE32OfflineTraceStatus(raw publicE32RawTrace, add func(string, ...any)) {
	if raw.TraceStatus != "TRACE_UNVERIFIED" || raw.TraceValidationError == "" {
		add("source raw must remain TRACE_UNVERIFIED with its original status note")
	}
	if raw.TraceValidationError != "raw event capture is not certified; a separate certified result is written only after all gates pass" {
		add("source raw status note differs from its frozen unverified note")
	}
	if raw.Profile != "logn13-e32-public-native" || raw.Mode != "test-only-public-e32-fast" {
		add("source raw profile/mode mismatch: %s/%s", raw.Profile, raw.Mode)
	}
	if raw.PrimaryCommit != batch025CapturedPrimaryCommit || raw.FixturePrimaryCommit != e32FixturePrimaryCommit ||
		raw.ProductionFastCommit != e32ProductionFastCommit || raw.DiagnosticHead != batch025DiagnosticCommit {
		add("source raw commit pins do not match the frozen E32 trace provenance")
	}
	if raw.CallBudget != 2 || raw.ActualCalls != 2 || len(raw.Runs) != 2 || raw.NumericalGate != 1e-6 {
		add("source raw budget/output metadata is inconsistent: budget=%d calls=%d runs=%d gate=%g", raw.CallBudget, raw.ActualCalls, len(raw.Runs), raw.NumericalGate)
	}
}

func validatePublicE32OfflineSidecar(raw publicE32RawTrace, sidecar publicE32FailureArtifact, rawSHA string, add func(string, ...any)) {
	if sidecar.SchemaVersion != "fastdiag.public-e32.failure.v1" || sidecar.TraceStatus != "TRACE_UNVERIFIED" || sidecar.FailurePhase != "event_tree_validation" ||
		sidecar.FailureReason != "power event is not nested under generated_powers" {
		add("failure sidecar does not preserve the expected original first live-validator failure")
	}
	if filepath.Base(sidecar.RawEvidenceFile) != "raw-trace-unverified.json" || sidecar.RawEvidenceSHA256 != rawSHA {
		add("failure sidecar does not bind the exact frozen raw SHA-256")
	}
	if sidecar.PrimaryCommit != raw.PrimaryCommit || sidecar.FixturePrimaryCommit != raw.FixturePrimaryCommit ||
		sidecar.ProductionFastCommit != raw.ProductionFastCommit || sidecar.DiagnosticHead != raw.DiagnosticHead ||
		sidecar.CallBudget != raw.CallBudget || sidecar.ActualCalls != raw.ActualCalls {
		add("failure sidecar source/call-budget metadata differs from the raw trace")
	}
}

func validatePublicE32OfflineProvenance(raw publicE32RawTrace, manifest publicE32FixtureManifest, vectors publicE32Vectors, manifestSHA string, add func(string, ...any)) {
	if raw.FixtureManifestSHA256 != manifestSHA || raw.FixtureManifestSHA256 != batch025ManifestSHA256 {
		add("raw fixture-manifest SHA does not match frozen manifest")
	}
	if raw.PrimarySourceSHA256 != manifest.PrimarySourceSHA256 || vectors.SourceSHA256 != manifest.PrimarySourceSHA256 ||
		raw.ConfigSHA256 != manifest.ConfigSHA256 || vectors.ConfigSHA256 != manifest.ConfigSHA256 ||
		raw.QPSHA256 != manifest.QPSHA256 || vectors.QPSHA256 != manifest.QPSHA256 ||
		raw.InputSHA256 != manifest.InputSHA256 || vectors.InputSHA256 != manifest.InputSHA256 ||
		raw.WorkloadSHA256 != manifest.WorkloadSHA256 || vectors.WorkloadSHA256 != manifest.WorkloadSHA256 {
		add("raw, manifest and frozen vectors disagree on source/config/QP/input/workload fingerprints")
	}
	if manifest.Profile != "logn13-e32-public-native" || manifest.Backend != "fast" || manifest.BackendCommit != e32ProductionFastCommit ||
		manifest.PrimaryCommit != e32FixturePrimaryCommit || manifest.EphemeralSecretWeight != 32 ||
		manifest.NumericalGate != 1e-6 || manifest.State.Level != 0 || manifest.State.Degree != 1 || manifest.State.QPrefixRows != 1 {
		add("fixture manifest violates frozen public LogN13/E32/Fast Level0 contract")
	}
	if vectors.SchemaVersion != "fast-standard-public-native-repeatability-vectors.v1" || vectors.Mode != "public-native-repeatability" ||
		vectors.Backend != "fast" || vectors.BackendCommit != e32ProductionFastCommit || vectors.PrimaryCommit != e32FixturePrimaryCommit {
		add("frozen six-output vectors violate their recorded public Fast provenance")
	}
	if raw.CiphertextSHA256 != manifest.CiphertextSHA256 || raw.ExpectedInputSHA256 != manifest.DecodedSHA256 {
		add("raw captured input/ciphertext identity differs from the frozen fixture")
	}
}

func validatePublicE32OfflineOutputMetadata(raw publicE32RawTrace, manifest publicE32FixtureManifest, vectors publicE32Vectors, doc *publicE32OfflineDocument, add func(string, ...any)) {
	if len(raw.Runs) != 2 {
		add("captured output count is %d, want two matched outputs", len(raw.Runs))
	}
	wantPhases := []string{"first_cold_bootstrap", "warm_bootstrap_01"}
	wantVectors := []string{"bootstrap_first_cold", "bootstrap_warm_01"}
	for index, run := range raw.Runs {
		vectorName := "unexpected"
		if index < len(wantVectors) {
			vectorName = wantVectors[index]
		}
		if index >= len(wantPhases) || run.Index != index+1 || run.Phase != wantPhases[index] {
			add("captured output %d has unexpected index/phase %d/%q", index+1, run.Index, run.Phase)
		}
		vectorValues, found := vectors.BootstrapOutputs[vectorName]
		vectorHash := ""
		if found && len(vectorValues) == 1<<12 {
			decoded := make([]complex128, len(vectorValues))
			for slot, value := range vectorValues {
				if math.IsNaN(value.Real) || math.IsInf(value.Real, 0) || math.IsNaN(value.Imag) || math.IsInf(value.Imag, 0) {
					add("matched vector %s has non-finite slot %d", vectorName, slot)
				}
				decoded[slot] = complex(value.Real, value.Imag)
			}
			vectorHash = perfmeasure.Fingerprint(decoded)
		} else {
			add("matched vector %s is absent or not exactly 4096 slots", vectorName)
		}
		hashMatches := vectorHash != "" && vectorHash == run.DecodedSHA256
		if !hashMatches {
			add("captured output %s decoded SHA does not match its frozen repeatability vector", run.Phase)
		}
		metricsPass := run.ElapsedNS >= 0 && run.Level == 1 && run.Degree == 1 && run.QPrefixRows == 2 &&
			run.Scale == manifest.State.Scale && finiteNonNegative(run.OracleRMSE) && finiteNonNegative(run.OracleMaxComplex) &&
			finiteNonNegative(run.FastReferenceWorstRMSE) && finiteNonNegative(run.FastReferenceWorstMax) &&
			run.OracleMaxComplex <= raw.NumericalGate && run.FastReferenceWorstMax <= raw.NumericalGate
		if !metricsPass {
			add("captured output %s fails Level/Scale/Degree/Q-prefix/finite numerical metadata gate", run.Phase)
		}
		doc.Outputs = append(doc.Outputs, publicE32OfflineOutput{
			Phase: run.Phase, VectorName: vectorName, DecodedSHA256: run.DecodedSHA256,
			MatchedVectorSHA256: vectorHash, DecodedHashMatches: hashMatches,
			ElapsedNS: run.ElapsedNS, Level: run.Level, Degree: run.Degree, Scale: run.Scale,
			QPrefixRows: run.QPrefixRows, OracleRMSE: run.OracleRMSE, OracleMaxComplex: run.OracleMaxComplex,
			FastReferenceWorstRMSE: run.FastReferenceWorstRMSE, FastReferenceWorstMax: run.FastReferenceWorstMax,
			NumericalMetadataPass: metricsPass && hashMatches, NativeDecryptReperformed: false,
		})
	}
}

func finiteNonNegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func inspectPublicE32OfflineProfile(rawPath, fileName, recordedSHA string) (publicE32OfflineProfile, error) {
	profile := publicE32OfflineProfile{FileName: fileName, RecordedSHA256: recordedSHA}
	if fileName == "" || filepath.Base(fileName) != fileName {
		profile.Disposition = "unavailable_or_invalid_filename"
		return profile, nil
	}
	path := filepath.Join(filepath.Dir(rawPath), "profiles", fileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		profile.Disposition = "profile_bytes_unavailable"
		return profile, nil
	}
	if err != nil {
		return profile, err
	}
	sum := sha256.Sum256(data)
	profile.ObservedSHA256 = hex.EncodeToString(sum[:])
	profile.FileAvailable = true
	profile.HashBoundByRawMetadata = recordedSHA != "" && recordedSHA == profile.ObservedSHA256
	profile.UsedForAttribution = false
	if recordedSHA == "" {
		profile.Disposition = "file_present_but_original_raw_has_no_digest; excluded_from_attribution"
	} else if !profile.HashBoundByRawMetadata {
		profile.Disposition = "recorded_digest_mismatch; excluded_from_attribution"
		return profile, fmt.Errorf("profile %s SHA-256 differs from raw metadata", fileName)
	} else {
		profile.Disposition = "hash_bound_to_raw_metadata_but_not_reanalyzed"
	}
	return profile, nil
}

func publicE32OfflineRootClosures(events []Event) []publicE32OfflineRootClosure {
	bySequence := make(map[uint64]Event, len(events))
	children := make(map[uint64][]Event, len(events))
	for _, event := range events {
		bySequence[event.Sequence] = event
		if event.ParentSequence != 0 {
			children[event.ParentSequence] = append(children[event.ParentSequence], event)
		}
	}
	var roots []Event
	for _, event := range events {
		if event.ParentSequence == 0 {
			roots = append(roots, event)
		}
	}
	closures := make([]publicE32OfflineRootClosure, 0, len(roots))
	for _, root := range roots {
		count := 0
		for _, event := range events {
			current := event
			for steps := 0; current.ParentSequence != 0 && steps <= len(events); steps++ {
				parent, exists := bySequence[current.ParentSequence]
				if !exists {
					break
				}
				current = parent
			}
			if current.Sequence == root.Sequence {
				count++
			}
		}
		var childNS float64
		for _, child := range children[root.Sequence] {
			childNS += float64(child.ElapsedNS)
		}
		inclusive := float64(root.ElapsedNS)
		closures = append(closures, publicE32OfflineRootClosure{
			Sequence: root.Sequence, Scope: root.Scope, Name: root.Name, EventCount: count,
			InclusiveNS: inclusive, ChildrenNS: childNS, SignedResidual: inclusive - childNS,
		})
	}
	return closures
}

func readFileSHA256(path string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func renderPublicE32OfflineSummary(doc publicE32OfflineDocument) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Batch025 — Fast E32 trace offline salvage\n\n- Classification: **`%s`**\n- Source raw status: `%s` (relabelled=%t)\n- New Bootstrap calls: **%d**\n- Raw: `%s` (%d bytes, SHA-256 `%s`)\n- Sidecar: `%s` (SHA-256 `%s`)\n- Fixture manifest: `%s` (SHA-256 `%s`)\n- Six-output Fast vectors: `%s` (SHA-256 `%s`)\n\n", doc.Classification, doc.RawTraceStatus, doc.OriginalRawRelabeled, doc.NewBootstrapCalls, doc.RawPath, doc.RawBytes, doc.RawSHA256, doc.SidecarPath, doc.SidecarSHA256, doc.ManifestPath, doc.ManifestSHA256, doc.VectorsPath, doc.VectorsSHA256)
	fmt.Fprintf(&out, "## Provenance\n\n- Captured Primary commit: `%s`; fixture Primary commit: `%s`; current offline analyzer: `%s` (`%s`, dirty=%t).\n- Captured measurement source SHA-256: `%s`; config `%s`; Q/P `%s`; input `%s`; workload `%s`.\n- Production Fast pin: `%s`; diagnostic Secondary: `%s` (current `%s`, `%s`, dirty=%t).\n- Diagnostic-vs-production source delta: `%t`, paths `%v`.\n- Reservation journals: %d/2; captured actual calls %d/%d. Reservation entries prove spent tokens; the raw holds the two output records.\n\n", doc.CapturedPrimaryCommit, doc.FixturePrimaryCommit, doc.PrimaryAnalyzer.Commit, doc.PrimaryAnalyzer.Ref, doc.PrimaryAnalyzer.Dirty, doc.PrimarySourceSHA256, doc.ConfigSHA256, doc.QPSHA256, doc.InputSHA256, doc.WorkloadSHA256, doc.ProductionFastCommit, doc.DiagnosticSecondaryCommit, doc.SecondarySource.Commit, doc.SecondarySource.Ref, doc.SecondarySource.Dirty, doc.InstrumentationOnlyDelta, doc.ChangedProductionPaths, len(doc.Attempts), doc.ActualCalls, doc.CallBudget)
	out.WriteString("## Preserved output metadata\n\nThese are checks of already-recorded metadata and vector hashes, not a new native decrypt or numerical experiment.\n\n| Phase | Elapsed (ms) | Level | Scale | Degree | Q-prefix rows | Oracle RMSE | Oracle max | Fast-reference max | Vector SHA match | Gate metadata |\n|---|---:|---:|---|---:|---:|---:|---:|---:|---|---|\n")
	for _, run := range doc.Outputs {
		fmt.Fprintf(&out, "| %s | %.3f | %d | %s | %d | %d | %.4g | %.4g | %.4g | %t | %t |\n", run.Phase, float64(run.ElapsedNS)/1e6, run.Level, run.Scale, run.Degree, run.QPrefixRows, run.OracleRMSE, run.OracleMaxComplex, run.FastReferenceWorstMax, run.DecodedHashMatches, run.NumericalMetadataPass)
	}
	fmt.Fprintf(&out, "\nCaptured numerical gate: `%.3g`. The Fast native output's reported decoded hashes match the corresponding held repeatability vectors; raw records report oracle RMSE %.4g and max complex difference %.4g. No numerical metric was recomputed from a new decrypt.\n\n", doc.NumericalGate, firstOfflineOutputMetric(doc.Outputs, true), firstOfflineOutputMetric(doc.Outputs, false))

	out.WriteString("## Event structure and timing\n\nThe full per-event inclusive/exclusive closures, root-partitioned Pareto, and conditional Amdahl scenarios are in the sibling `-evidence.json` artifact under `root_partitioned_event_analysis`. Negative signed residuals indicate observed child timing exceeds its parent; they are retained rather than normalized. Independent Power and Rescale trees are never combined with Bootstrap-root percentages.\n\n")
	fmt.Fprintf(&out, "- Events: %d; independent roots: %d; events with negative exclusive residual: %d.\n- Stage counts: `%s`.\n\n", doc.EventCount, doc.IndependentRootCount, doc.NegativeExclusiveCount, formatPublicE32StageCounts(doc.StageCounts))
	out.WriteString("### Descriptive event-type medians (no shares)\n\n| Scope | Event | Component | Parent type | Count | Median (ms) |\n|---|---|---|---|---:|---:|\n")
	for _, stat := range doc.EventMedians {
		if stat.Scope != "stage" && stat.Scope != "power" && stat.Scope != "rescale" {
			continue
		}
		fmt.Fprintf(&out, "| %s | %s | %s | %s | %d | %.3f |\n", stat.Scope, stat.Name, displayOrDash(stat.Component), displayOrDash(stat.ParentKey), stat.Count, stat.MedianNS/1e6)
	}
	out.WriteString("\nThese medians are descriptive event-type summaries only; they do not sum nested durations or form cross-root percentages. Consult the root-partitioned evidence for exact per-root Pareto and closure.\n\n")
	if doc.Analysis != nil {
		fmt.Fprintf(&out, "Bootstrap root closure: %.3f ms inclusive, %.3f ms direct children, signed residual %+.3f ms. Amdahl scenarios are conditional mathematical bounds on measured inclusive stage fractions, not speedup predictions.\n\n", doc.Analysis.RootElapsedNS/1e6, doc.Analysis.RootImmediateChildrenNS/1e6, doc.Analysis.RootUnattributedNS/1e6)
	}
	out.WriteString("## Profile evidence\n\n")
	fmt.Fprintf(&out, "- CPU: `%s`, file available=%t, raw-bound hash=%t, disposition `%s`.\n- Heap: `%s`, file available=%t, raw-bound hash=%t, disposition `%s`.\n\n", doc.CPUProfile.ObservedSHA256, doc.CPUProfile.FileAvailable, doc.CPUProfile.HashBoundByRawMetadata, doc.CPUProfile.Disposition, doc.HeapProfile.ObservedSHA256, doc.HeapProfile.FileAvailable, doc.HeapProfile.HashBoundByRawMetadata, doc.HeapProfile.Disposition)
	out.WriteString("## Validation anomalies\n\n")
	if len(doc.ValidationAnomalies) == 0 {
		out.WriteString("None. The separate offline analysis passes the strict frozen-source, output-metadata, full event-tree and root-partition gates. The original raw artifact remains `TRACE_UNVERIFIED`.\n\n")
	} else {
		for _, anomaly := range doc.ValidationAnomalies {
			fmt.Fprintf(&out, "- %s\n", anomaly)
		}
		out.WriteString("\n")
	}
	out.WriteString("## Interpretation boundary\n\nThis is a Fast-only offline analysis of the already captured 539-event trace. It does not establish a Standard timing comparison, a speedup, secure-CKKS equivalence, or a new Bootstrap result. Batch023's historical 4.225× uninstrumented observation remains separate and is not inferred from this trace.\n")
	return out.String()
}

func formatPublicE32StageCounts(counts map[string]int) string {
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	var out strings.Builder
	for i, name := range names {
		if i != 0 {
			out.WriteString(", ")
		}
		fmt.Fprintf(&out, "%s=%d", name, counts[name])
	}
	return out.String()
}

func firstOfflineOutputMetric(outputs []publicE32OfflineOutput, rmse bool) float64 {
	if len(outputs) == 0 {
		return 0
	}
	if rmse {
		return outputs[0].OracleRMSE
	}
	return outputs[0].OracleMaxComplex
}

func renderPublicE32OfflineJournal(doc publicE32OfflineDocument) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Batch025 offline salvage journal\n\n- Status: `%s`\n- Generated: `%s`\n- New Standard/Fast Bootstrap calls: **0**\n- Original raw SHA-256: `%s`\n\n", doc.Classification, doc.GeneratedAt.Format(time.RFC3339Nano), doc.RawSHA256)
	out.WriteString("## Checkpoints\n\n| Checkpoint | Result | Evidence boundary |\n|---|---|---|\n")
	fmt.Fprintf(&out, "| O1 immutable evidence/source audit | %s | Raw/sidecar/manifest/vectors hashes, source pins, attempt journals and source identity recorded in evidence JSON. |\n", offlineCheckpoint(doc, "O1"))
	fmt.Fprintf(&out, "| O2 offline replay and full-tree validation | %s | %d events; original raw remains `%s`; no trace/runner/Bootstrap call. |\n", offlineCheckpoint(doc, "O2"), doc.EventCount, doc.RawTraceStatus)
	fmt.Fprintf(&out, "| O3 timing and closure analysis | %s | %d independent roots; separate-root Pareto and signed closures in evidence JSON. |\n", offlineCheckpoint(doc, "O3"), doc.IndependentRootCount)
	fmt.Fprintf(&out, "| O4 artifacts | %s | Summary, journal, evidence and platform documentation are task outputs; output files are create-only. |\n\n", offlineCheckpoint(doc, "O4"))
	out.WriteString("## Attempt accounting\n\n")
	for _, attempt := range doc.Attempts {
		fmt.Fprintf(&out, "- Attempt %02d `%s`: `%s`, budget=%d, phase=`%s`, journal SHA-256 `%s`.\n", attempt.AttemptIndex, attempt.Status, attempt.ReservedAt, attempt.BootstrapBudget, attempt.Phase, attempt.SHA256)
	}
	fmt.Fprintf(&out, "\nCaptured budget/outputs: %d/%d calls; newly invoked calls: 0.\n", doc.ActualCalls, doc.CallBudget)
	if len(doc.ValidationAnomalies) == 0 {
		out.WriteString("\nNo offline validation anomalies. The raw capture is immutable and still unverified as a live trace artifact; only this separate offline replay receives its classification.\n")
	} else {
		out.WriteString("\n## Remaining anomalies\n\n")
		for _, anomaly := range doc.ValidationAnomalies {
			fmt.Fprintf(&out, "- %s\n", anomaly)
		}
	}
	return out.String()
}

func offlineCheckpoint(doc publicE32OfflineDocument, checkpoint string) string {
	if len(doc.ValidationAnomalies) != 0 || doc.Analysis == nil {
		return "PARTIAL_UNVERIFIED"
	}
	return "PASS"
}
