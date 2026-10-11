package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	task006SummaryFile        = "FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.json"
	task006RawSHA256          = "9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d"
	task006SidecarSHA256      = "f598b481b776448dcc78ff80083610e9376cee12451716ddb88fd414abfc756b"
	task006LedgerSHA256       = "e752dbaf4aee28229e14cbcbc85c6c8c9f4b4467c0da3b5dbf29577b4e28672d"
	task006AdapterPatchSHA256 = "3ad46f934d597472caac0f0810719425dec75bde0064342cc0b05278510b50ae"
	task006PrimaryCommit      = "aff322665f6d178f159408d66eae39a2823815f5"
	task006FastCommit         = "2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f"
	task006AdapterCommit      = "649da85757a97dc8c60b5f9cc031634f39ad4a40"
	task006AdapterPatchFile   = "cmd/fastdiag/testdata/logn16-public-e32-test.patch.b64"
)

type task006OfflineSummaryInput struct {
	SchemaVersion   string `json:"schema_version"`
	Task            string `json:"task"`
	Classification  string `json:"classification"`
	TraceStatus     string `json:"trace_status"`
	TraceValidation struct {
		EventCount    int    `json:"event_count"`
		FailureReason string `json:"failure_reason"`
		FailurePhase  string `json:"failure_phase"`
	} `json:"trace_validation"`
	Provenance struct {
		PrimaryCommit             string `json:"primary_commit"`
		PrimarySourceSHA256       string `json:"primary_measurement_source_sha256"`
		SecondaryProductionCommit string `json:"secondary_production_commit"`
		DiagnosticAdapterCommit   string `json:"diagnostic_test_adapter_commit"`
		DiagnosticPatchSHA256     string `json:"diagnostic_test_adapter_patch_sha256"`
		ConfigSHA256              string `json:"config_sha256"`
		QPSHA256                  string `json:"qp_sha256"`
		InputSHA256               string `json:"canonical_input_sha256"`
		WorkloadSHA256            string `json:"workload_sha256"`
		ManifestSHA256            string `json:"fixture_manifest_sha256"`
		VectorsSHA256             string `json:"fixture_vectors_sha256"`
		CiphertextSHA256          string `json:"fixture_ciphertext_sha256"`
		Task005VectorsSHA256      string `json:"task005_fast_reference_vectors_sha256"`
		Task005ResultSHA256       string `json:"task005_fast_reference_result_sha256"`
	} `json:"provenance"`
	Evidence struct {
		RawPath       string `json:"raw_unverified_trace_path"`
		RawSHA256     string `json:"raw_unverified_trace_sha256"`
		SidecarPath   string `json:"failure_artifact_path"`
		SidecarSHA256 string `json:"failure_artifact_sha256"`
		LedgerPath    string `json:"attempt_ledger_path"`
		LedgerSHA256  string `json:"attempt_ledger_sha256"`
		Journals      []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"reservation_journals"`
	} `json:"evidence"`
	BootstrapBudget struct {
		Task005Reserved   int `json:"task005_reserved"`
		Task005Actual     int `json:"task005_actual"`
		Task006Authorized int `json:"task006_authorized"`
		Task006Reserved   int `json:"task006_reserved"`
		Task006Actual     int `json:"task006_actual"`
		RetryCount        int `json:"retry_count"`
	} `json:"bootstrap_budget"`
	Runs []struct {
		Phase         string  `json:"phase"`
		Level         int     `json:"level"`
		Degree        int     `json:"degree"`
		Scale         string  `json:"scale"`
		Rows          int     `json:"q_prefix_rows"`
		OracleRMSE    float64 `json:"oracle_complex_rmse"`
		OracleMax     float64 `json:"oracle_max_complex_difference"`
		ReferenceRMSE float64 `json:"task005_fast_reference_worst_rmse"`
		ReferenceMax  float64 `json:"task005_fast_reference_worst_max_complex_difference"`
		GatePassed    bool    `json:"output_gate_passed"`
	} `json:"task006_fast_only_runs"`
}

type task006OfflineLedger struct {
	SchemaVersion         string `json:"schema_version"`
	Task                  string `json:"task"`
	CompletionStatus      string `json:"completion_status"`
	Task005Actual         int    `json:"task005_actual_calls_not_reused"`
	Task005Reserved       int    `json:"task005_reserved_calls_not_reused"`
	Task006Actual         int    `json:"task006_actual_calls"`
	Task006ActualObserved bool   `json:"task006_actual_calls_observed"`
	Task006Budget         int    `json:"task006_bootstrap_budget"`
	Task006Reserved       int    `json:"task006_reserved_calls"`
	Reservations          []struct {
		AttemptIndex int    `json:"attempt_index"`
		Phase        string `json:"phase"`
		Status       string `json:"status"`
		JournalFile  string `json:"journal_file"`
		JournalSHA   string `json:"journal_sha256"`
	} `json:"attempts"`
}

type task006OfflineJournalInput struct {
	SchemaVersion   string `json:"schema_version"`
	AttemptIndex    int    `json:"attempt_index"`
	BootstrapBudget int    `json:"bootstrap_budget"`
	Phase           string `json:"phase"`
	Status          string `json:"status"`
	ReservedAt      string `json:"reserved_at"`
}

type task006OfflineRunCheck struct {
	Phase         string  `json:"phase"`
	Level         int     `json:"level"`
	Degree        int     `json:"degree"`
	ScaleLog2     float64 `json:"scale_log2"`
	QPrefixRows   int     `json:"q_prefix_rows"`
	OracleRMSE    float64 `json:"recorded_oracle_rmse"`
	OracleMax     float64 `json:"recorded_oracle_max_complex_difference"`
	ReferenceRMSE float64 `json:"recorded_task005_fast_reference_rmse"`
	ReferenceMax  float64 `json:"recorded_task005_fast_reference_max_complex_difference"`
	Pass          bool    `json:"recorded_gate_passed"`
}

type task006PowerRoot struct {
	Branch      string  `json:"branch"`
	Sequence    uint64  `json:"root_sequence"`
	InclusiveNS float64 `json:"inclusive_ns"`
	ChildrenNS  float64 `json:"immediate_children_ns"`
	ResidualNS  float64 `json:"signed_unattributed_ns"`
}

type task006StageTiming struct {
	Name        string  `json:"stage"`
	Count       int     `json:"count"`
	InclusiveNS float64 `json:"inclusive_ns"`
	Share       float64 `json:"share_of_bootstrap_root"`
}

type task006RescaleRootTiming struct {
	Sequence    uint64  `json:"root_sequence"`
	EventCount  int     `json:"tree_event_count"`
	InclusiveNS float64 `json:"inclusive_ns"`
	ChildrenNS  float64 `json:"immediate_children_ns"`
	ResidualNS  float64 `json:"signed_unattributed_ns"`
}

type task006OfflineAnalysis struct {
	RootElapsedNS            float64                    `json:"bootstrap_root_elapsed_ns"`
	RootImmediateChildrenNS  float64                    `json:"bootstrap_direct_children_ns"`
	RootUnattributedNS       float64                    `json:"bootstrap_unattributed_ns"`
	RootUnattributedFraction float64                    `json:"bootstrap_unattributed_fraction"`
	EventTypeRootRecords     int                        `json:"event_type_root_record_count"`
	PositiveExclusiveGroups  int                        `json:"positive_exclusive_group_count"`
	AmdahlScenarioCount      int                        `json:"amdahl_scenario_count"`
	BootstrapStages          []task006StageTiming       `json:"nonoverlapping_bootstrap_direct_stages"`
	RescaleDiagnosticRoots   []task006RescaleRootTiming `json:"independent_rescale_diagnostic_roots"`
}

type task006OfflineEvidence struct {
	SchemaVersion             string                          `json:"schema_version"`
	GeneratedAt               time.Time                       `json:"generated_at"`
	Task                      string                          `json:"task"`
	Classification            string                          `json:"classification"`
	OriginalClassification    string                          `json:"original_classification"`
	OriginalTraceStatus       string                          `json:"original_trace_status"`
	OriginalRawRelabeled      bool                            `json:"original_raw_relabelled"`
	NewBootstrapCalls         int                             `json:"new_bootstrap_calls"`
	PrimaryAnalyzer           RepositoryMetadata              `json:"primary_analyzer_repository"`
	SecondaryProductionSource RepositoryMetadata              `json:"secondary_production_repository"`
	RawPath                   string                          `json:"raw_path"`
	RawSHA256                 string                          `json:"raw_sha256"`
	RawBytes                  int                             `json:"raw_bytes"`
	SidecarPath               string                          `json:"failure_sidecar_path"`
	SidecarSHA256             string                          `json:"failure_sidecar_sha256"`
	SummaryPath               string                          `json:"task006_summary_path"`
	SummarySHA256             string                          `json:"task006_summary_sha256"`
	LedgerPath                string                          `json:"attempt_ledger_path"`
	LedgerSHA256              string                          `json:"attempt_ledger_sha256"`
	AttemptJournals           []publicE32OfflineAttempt       `json:"attempt_journals"`
	Profile                   string                          `json:"profile"`
	Mode                      string                          `json:"mode"`
	CapturedPrimaryCommit     string                          `json:"captured_primary_commit"`
	CapturedPrimarySourceSHA  string                          `json:"primary_measurement_source_sha256"`
	EffectiveParameters       perfmeasure.EffectiveParameters `json:"effective_parameters"`
	ProductionFastCommit      string                          `json:"production_fast_commit"`
	DiagnosticAdapterCommit   string                          `json:"diagnostic_test_adapter_commit"`
	DiagnosticPatchSHA256     string                          `json:"diagnostic_test_adapter_patch_sha256"`
	DiagnosticPatchMatches    bool                            `json:"diagnostic_test_adapter_patch_matches"`
	DiagnosticCommitAvailable bool                            `json:"diagnostic_commit_object_available_in_current_clone"`
	ConfigSHA256              string                          `json:"config_sha256"`
	QPSHA256                  string                          `json:"qp_sha256"`
	InputSHA256               string                          `json:"input_sha256"`
	WorkloadSHA256            string                          `json:"workload_sha256"`
	FixtureManifestSHA256     string                          `json:"fixture_manifest_sha256"`
	FixtureVectorsSHA256      string                          `json:"fixture_vectors_sha256"`
	CiphertextSHA256          string                          `json:"ciphertext_sha256"`
	Task005ReferenceHashes    []string                        `json:"task005_reference_hashes"`
	CallBudget                int                             `json:"captured_call_budget"`
	ActualCalls               int                             `json:"captured_actual_calls"`
	EventCount                int                             `json:"event_count"`
	StageCounts               map[string]int                  `json:"stage_counts"`
	IndependentRootCount      int                             `json:"independent_root_count"`
	GeneratedPowerRoots       []task006PowerRoot              `json:"generated_power_branch_roots"`
	RootClosures              []publicE32OfflineRootClosure   `json:"independent_root_closures"`
	MetadataCoverage          map[string]int                  `json:"event_metadata_field_coverage"`
	RecursivePowerAncestry    []uint64                        `json:"recursive_power_ancestry_sequences"`
	RunChecks                 []task006OfflineRunCheck        `json:"recorded_output_checks"`
	Analysis                  *task006OfflineAnalysis         `json:"root_partitioned_analysis,omitempty"`
	ValidationAnomalies       []string                        `json:"validation_anomalies,omitempty"`
	Limitations               []string                        `json:"limitations"`
}

func replayPublicE32Task006Offline(primaryRoot, secondaryRoot string, primaryMeta RepositoryMetadata, opts publicE32OfflineOptions) (task006OfflineEvidence, string, error) {
	var doc task006OfflineEvidence
	addAnomaly := func(format string, args ...any) {
		doc.ValidationAnomalies = append(doc.ValidationAnomalies, fmt.Sprintf(format, args...))
	}
	readExpected := func(path, expected string) ([]byte, string, error) {
		data, digest, err := readFileSHA256(path)
		if err != nil {
			return nil, "", err
		}
		if digest != expected {
			return nil, digest, fmt.Errorf("immutable Task006 evidence SHA-256 mismatch for %s: got %s want %s", path, digest, expected)
		}
		return data, digest, nil
	}

	summaryBytes, summarySHA, err := readFileSHA256(opts.task006Summary)
	if err != nil {
		return doc, "", fmt.Errorf("read Task006 summary: %w", err)
	}
	var summary task006OfflineSummaryInput
	if err := json.Unmarshal(summaryBytes, &summary); err != nil {
		return doc, "", fmt.Errorf("decode Task006 summary: %w", err)
	}
	if summary.SchemaVersion != "fastdiag.logn16.task006-summary.v1" || summary.Task != "FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006" {
		return doc, "", errors.New("input summary is not the frozen Task006 LogN16 trace record")
	}

	rawBytes, rawSHA, err := readExpected(opts.rawPath, task006RawSHA256)
	if err != nil {
		return doc, "", err
	}
	var raw publicE32RawTrace
	if err := json.Unmarshal(rawBytes, &raw); err != nil {
		return doc, "", fmt.Errorf("decode Task006 raw trace: %w", err)
	}
	sidecarBytes, sidecarSHA, err := readExpected(opts.sidecarPath, task006SidecarSHA256)
	if err != nil {
		return doc, "", err
	}
	var sidecar publicE32FailureArtifact
	if err := json.Unmarshal(sidecarBytes, &sidecar); err != nil {
		return doc, "", fmt.Errorf("decode Task006 failure sidecar: %w", err)
	}
	ledgerBytes, ledgerSHA, err := readExpected(opts.task006Ledger, task006LedgerSHA256)
	if err != nil {
		return doc, "", err
	}
	var ledger task006OfflineLedger
	if err := json.Unmarshal(ledgerBytes, &ledger); err != nil {
		return doc, "", fmt.Errorf("decode Task006 attempt ledger: %w", err)
	}

	doc = task006OfflineEvidence{
		SchemaVersion: "fastdiag.public-e32.task006-offline.v1", GeneratedAt: time.Now().UTC(),
		Task: summary.Task, OriginalClassification: summary.Classification, OriginalTraceStatus: raw.TraceStatus,
		OriginalRawRelabeled: false, NewBootstrapCalls: 0,
		PrimaryAnalyzer: primaryMeta, RawPath: opts.rawPath, RawSHA256: rawSHA, RawBytes: len(rawBytes),
		SidecarPath: opts.sidecarPath, SidecarSHA256: sidecarSHA, SummaryPath: opts.task006Summary, SummarySHA256: summarySHA,
		LedgerPath: opts.task006Ledger, LedgerSHA256: ledgerSHA,
		Profile: raw.Profile, Mode: raw.Mode, CapturedPrimaryCommit: raw.PrimaryCommit,
		CapturedPrimarySourceSHA: raw.PrimarySourceSHA256, ProductionFastCommit: raw.ProductionFastCommit,
		DiagnosticAdapterCommit: raw.DiagnosticTestAdapterCommit, DiagnosticPatchSHA256: raw.DiagnosticTestAdapterPatchSHA256,
		ConfigSHA256: raw.ConfigSHA256, QPSHA256: raw.QPSHA256, InputSHA256: raw.InputSHA256,
		WorkloadSHA256: raw.WorkloadSHA256, FixtureManifestSHA256: raw.FixtureManifestSHA256,
		FixtureVectorsSHA256: summary.Provenance.VectorsSHA256, CiphertextSHA256: raw.CiphertextSHA256,
		Task005ReferenceHashes: []string{raw.FastReferenceVectorsSHA256, raw.FastReferenceResultSHA256},
		CallBudget:             raw.CallBudget, ActualCalls: raw.ActualCalls, EventCount: len(raw.Events),
		StageCounts: make(map[string]int), MetadataCoverage: map[string]int{},
		AttemptJournals: make([]publicE32OfflineAttempt, 0, 2),
		Limitations: []string{
			"Analysis reuses the maintained Batch025 recursive event validator and root-partitioned analyzer; it does not rerun Bootstrap, decrypt, or change the immutable raw trace.",
			"The historical diagnostic adapter commit object is absent from the current Secondary clone. Its recorded commit and patch digest match Task006 metadata, and the committed patch digest is independently verified; the event attribution is certified from captured trace data, not a fresh source-tree rebuild.",
			"Task005 reference vector/result hashes are recorded provenance; the historical temporary vector payloads are not embedded in this repository and are not reconstructed here.",
		},
	}

	if filepath.Base(opts.task006Summary) != task006SummaryFile || normalizedEvidencePath(primaryRoot, summary.Evidence.RawPath) != normalizedEvidencePath(primaryRoot, opts.rawPath) || summary.Evidence.RawSHA256 != rawSHA ||
		normalizedEvidencePath(primaryRoot, summary.Evidence.SidecarPath) != normalizedEvidencePath(primaryRoot, opts.sidecarPath) || summary.Evidence.SidecarSHA256 != sidecarSHA ||
		normalizedEvidencePath(primaryRoot, summary.Evidence.LedgerPath) != normalizedEvidencePath(primaryRoot, opts.task006Ledger) || summary.Evidence.LedgerSHA256 != ledgerSHA {
		addAnomaly("Task006 summary evidence paths/hashes do not bind the supplied immutable artifacts")
	}
	if summary.Classification != "TRACE_COLLECTED_UNVERIFIED" || summary.TraceStatus != "TRACE_UNVERIFIED" || raw.TraceStatus != "TRACE_UNVERIFIED" || raw.TraceValidationError == "" {
		addAnomaly("Task006 original trace classification/status does not remain explicitly unverified")
	}
	if raw.SchemaVersion != "fastdiag.public-e32.trace.v1" || raw.Profile != "logn16-e32-public-native" || raw.Mode != "test-only-public-e32-fast" ||
		raw.PrimaryCommit != task006PrimaryCommit || raw.FixturePrimaryCommit != task006PrimaryCommit || raw.ProductionFastCommit != task006FastCommit ||
		raw.DiagnosticHead != task006AdapterCommit || raw.DiagnosticTestAdapterCommit != task006AdapterCommit ||
		raw.DiagnosticTestAdapterPatchSHA256 != task006AdapterPatchSHA256 || !raw.ProductionSourceIdentical ||
		raw.ConfigSHA256 != summary.Provenance.ConfigSHA256 || raw.QPSHA256 != summary.Provenance.QPSHA256 ||
		raw.InputSHA256 != summary.Provenance.InputSHA256 || raw.WorkloadSHA256 != summary.Provenance.WorkloadSHA256 ||
		raw.FixtureManifestSHA256 != summary.Provenance.ManifestSHA256 || raw.CiphertextSHA256 != summary.Provenance.CiphertextSHA256 ||
		raw.FastReferenceVectorsSHA256 != summary.Provenance.Task005VectorsSHA256 || raw.FastReferenceResultSHA256 != summary.Provenance.Task005ResultSHA256 {
		addAnomaly("Task006 raw trace source/profile/config/QP/input/workload/reference pins disagree with the compact summary")
	}
	if raw.CallBudget != 2 || raw.ActualCalls != 2 || len(raw.Runs) != 2 || raw.NumericalGate != 1e-6 || len(raw.Events) != 539 {
		addAnomaly("Task006 raw trace does not preserve the recorded 2-call / 539-event / 1e-6 contract")
	}
	if summary.TraceValidation.EventCount != len(raw.Events) || summary.TraceValidation.FailurePhase != "event_tree_validation" ||
		summary.TraceValidation.FailureReason != "power event is not nested under generated_powers" {
		addAnomaly("Task006 original failure note does not match the immutable raw event count and known obsolete direct-child rejection")
	}
	if sidecar.SchemaVersion != "fastdiag.public-e32.failure.v1" || sidecar.TraceStatus != "TRACE_UNVERIFIED" ||
		sidecar.FailurePhase != summary.TraceValidation.FailurePhase || sidecar.FailureReason != summary.TraceValidation.FailureReason ||
		sidecar.RawEvidenceSHA256 != rawSHA || filepath.Base(sidecar.RawEvidenceFile) != filepath.Base(opts.rawPath) ||
		sidecar.PrimaryCommit != raw.PrimaryCommit || sidecar.ProductionFastCommit != raw.ProductionFastCommit ||
		sidecar.DiagnosticHead != raw.DiagnosticHead || sidecar.CallBudget != raw.CallBudget || sidecar.ActualCalls != raw.ActualCalls {
		addAnomaly("Task006 failure sidecar does not bind the exact raw trace and its recorded first failure")
	}
	if ledger.SchemaVersion != "fastdiag.logn16.task006-attempt-ledger.v1" || ledger.Task != summary.Task ||
		ledger.Task005Reserved != 16 || ledger.Task005Actual != 16 || ledger.Task006Budget != 2 || ledger.Task006Reserved != 2 ||
		ledger.Task006Actual != 2 || !ledger.Task006ActualObserved || len(ledger.Reservations) != 2 || ledger.CompletionStatus != "TRACE_FAILED_RESERVED_NO_RETRY" {
		addAnomaly("Task006 attempt ledger does not preserve the recorded Task005 16/16 and Task006 2/2 no-retry accounting")
	}
	if summary.BootstrapBudget.Task005Reserved != ledger.Task005Reserved || summary.BootstrapBudget.Task005Actual != ledger.Task005Actual ||
		summary.BootstrapBudget.Task006Authorized != ledger.Task006Budget || summary.BootstrapBudget.Task006Reserved != ledger.Task006Reserved ||
		summary.BootstrapBudget.Task006Actual != ledger.Task006Actual || summary.BootstrapBudget.RetryCount != 0 {
		addAnomaly("Task006 compact summary and immutable call ledger disagree")
	}

	for index, path := range []string{opts.attemptOne, opts.attemptTwo} {
		data, digest, readErr := readFileSHA256(path)
		if readErr != nil {
			addAnomaly("attempt-%02d reservation journal unavailable: %v", index+1, readErr)
			continue
		}
		var attempt task006OfflineJournalInput
		if err := json.Unmarshal(data, &attempt); err != nil {
			addAnomaly("attempt-%02d reservation journal JSON is invalid: %v", index+1, err)
			continue
		}
		wantPhase := "diagnostic_fast_cold"
		if index == 1 {
			wantPhase = "diagnostic_fast_traced_warm"
		}
		if attempt.SchemaVersion != "perfprobe-bootstrap-attempt.v1" || attempt.AttemptIndex != index+1 || attempt.BootstrapBudget != 2 ||
			attempt.Phase != wantPhase || attempt.Status != "irrevocably_reserved_before_call" || attempt.ReservedAt == "" {
			addAnomaly("attempt-%02d reservation journal fails the irrevocable pre-call contract", index+1)
		}
		if index >= len(summary.Evidence.Journals) || normalizedEvidencePath(primaryRoot, summary.Evidence.Journals[index].Path) != normalizedEvidencePath(primaryRoot, path) || summary.Evidence.Journals[index].SHA256 != digest {
			addAnomaly("attempt-%02d journal path/SHA-256 does not match Task006 summary", index+1)
		}
		if index >= len(ledger.Reservations) || ledger.Reservations[index].AttemptIndex != index+1 || ledger.Reservations[index].Phase != wantPhase ||
			ledger.Reservations[index].Status != attempt.Status || ledger.Reservations[index].JournalSHA != digest || filepath.Base(ledger.Reservations[index].JournalFile) != filepath.Base(path) {
			addAnomaly("attempt-%02d reservation is not bound consistently by Task006 ledger", index+1)
		}
		doc.AttemptJournals = append(doc.AttemptJournals, publicE32OfflineAttempt{
			Path: path, SHA256: digest, AttemptIndex: attempt.AttemptIndex, BootstrapBudget: attempt.BootstrapBudget,
			Phase: attempt.Phase, Status: attempt.Status, ReservedAt: attempt.ReservedAt,
		})
	}
	if len(summary.Evidence.Journals) != 2 || len(doc.AttemptJournals) != 2 {
		addAnomaly("found %d Task006 reservation journals; want exactly two", len(doc.AttemptJournals))
	}

	if err := gitRun(primaryRoot, "merge-base", "--is-ancestor", task006PrimaryCommit, primaryMeta.Commit); err != nil {
		addAnomaly("captured Primary commit %s is not an ancestor of current analyzer %s", task006PrimaryCommit, primaryMeta.Commit)
	}
	primarySourceSHA, sourceErr := primaryMeasurementSourceFingerprint(primaryRoot, raw.PrimaryCommit)
	if sourceErr != nil {
		addAnomaly("captured Primary measurement source fingerprint is unavailable: %v", sourceErr)
	} else if primarySourceSHA != raw.PrimarySourceSHA256 || primarySourceSHA != summary.Provenance.PrimarySourceSHA256 {
		addAnomaly("captured Primary source fingerprint %s differs from raw/summary", primarySourceSHA)
	}
	cfgPath := filepath.Join(primaryRoot, "configs", "bootstrap_config.logN16.json")
	cfg, configBytes, configErr := perfmeasure.LoadConfig(cfgPath)
	if configErr != nil {
		addAnomaly("cannot load the pinned LogN16 config for effective-parameter verification: %v", configErr)
	} else {
		if hex.EncodeToString(configBytes[:]) != raw.ConfigSHA256 {
			addAnomaly("Task006 config file digest does not match the recorded config SHA-256")
		}
		_, bootstrapParams, effective, paramsErr := perfmeasure.ParametersFromConfigWithE(cfg, 32)
		if paramsErr != nil {
			addAnomaly("cannot derive Task006 effective CKKS parameters from the pinned config: %v", paramsErr)
		} else {
			doc.EffectiveParameters = effective
			qpHash := perfmeasure.QPFingerprint(bootstrapParams.BootstrappingParameters.Q(), bootstrapParams.BootstrappingParameters.P())
			inputHash := perfmeasure.Fingerprint(perfmeasure.DeterministicInput(effective.InputSlots))
			valuesA, valuesB, valuesC, workloadErr := perfmeasure.PublicMulRescaleWorkload(effective.InputSlots)
			workloadHash := ""
			if workloadErr == nil && len(bootstrapParams.BootstrappingParameters.Q()) > 5 {
				workloadHash, workloadErr = perfmeasure.PublicMulRescaleWorkloadFingerprint(valuesA, valuesB, valuesC, new(big.Int).SetUint64(bootstrapParams.BootstrappingParameters.Q()[5]).String())
			}
			if effective.LogN != 16 || effective.LogSlots != 15 || effective.InputSlots != 1<<15 || effective.EphemeralSecretWeight != 32 ||
				qpHash != raw.QPSHA256 || inputHash != raw.InputSHA256 || workloadErr != nil || workloadHash != raw.WorkloadSHA256 {
				addAnomaly("effective LogN16/E32 config, Q/P, canonical input, or workload fingerprints disagree with Task006 capture")
			}
		}
	}
	secondaryMeta, metaErr := repositoryMetadata(secondaryRoot, "fast-qprefix")
	if metaErr != nil {
		addAnomaly("cannot read Secondary production source metadata: %v", metaErr)
	} else {
		doc.SecondaryProductionSource = secondaryMeta
		if secondaryMeta.Dirty || secondaryMeta.Commit != task006FastCommit {
			addAnomaly("current Secondary must remain clean at Task006 production pin %s; got %s dirty=%t", task006FastCommit, secondaryMeta.Commit, secondaryMeta.Dirty)
		}
		if _, err := gitOutput(secondaryRoot, "cat-file", "-e", task006AdapterCommit+"^{commit}"); err == nil {
			doc.DiagnosticCommitAvailable = true
		}
	}
	patchBytes, patchReadErr := os.ReadFile(filepath.Join(primaryRoot, task006AdapterPatchFile))
	if patchReadErr != nil {
		addAnomaly("cannot read committed Task006 test-only adapter patch: %v", patchReadErr)
	} else {
		decodedPatch, decodeErr := base64.StdEncoding.DecodeString(strings.TrimSpace(string(patchBytes)))
		if decodeErr != nil {
			addAnomaly("committed Task006 test-only adapter patch is not valid base64: %v", decodeErr)
		} else {
			sum := sha256.Sum256(decodedPatch)
			doc.DiagnosticPatchMatches = hex.EncodeToString(sum[:]) == task006AdapterPatchSHA256
			if !doc.DiagnosticPatchMatches {
				addAnomaly("committed Task006 test-only adapter patch SHA-256 does not match recorded provenance")
			}
		}
	}

	validateTask006RawRuns(raw, summary, &doc, addAnomaly)
	if err := validatePublicE32Events(raw.Events); err != nil {
		for _, line := range strings.Split(err.Error(), "\n") {
			if strings.TrimSpace(line) != "" {
				doc.ValidationAnomalies = append(doc.ValidationAnomalies, "event structure: "+line)
			}
		}
	}
	doc.RootClosures = publicE32OfflineRootClosures(raw.Events)
	doc.IndependentRootCount = len(doc.RootClosures)
	if branches, branchErr := publicE32GeneratedPowerBranchRoots(raw.Events); branchErr != nil {
		addAnomaly("cannot project generated-power roots: %v", branchErr)
	} else {
		seenBranches := make(map[string]bool, len(branches))
		for _, branch := range branches {
			root := findTask006RootClosure(doc.RootClosures, branch.Root.Sequence)
			if root.Sequence == 0 || seenBranches[branch.Branch] {
				addAnomaly("generated-power branch %q has no unique independent root closure", branch.Branch)
				continue
			}
			seenBranches[branch.Branch] = true
			doc.GeneratedPowerRoots = append(doc.GeneratedPowerRoots, task006PowerRoot{
				Branch: branch.Branch, Sequence: branch.Root.Sequence,
				InclusiveNS: root.InclusiveNS, ChildrenNS: root.ChildrenNS, ResidualNS: root.SignedResidual,
			})
		}
		if len(branches) != 2 || !seenBranches["real"] || !seenBranches["imag"] {
			addAnomaly("Task006 generated-power projection must contain exactly one real and one imag root")
		}
	}
	for _, event := range raw.Events {
		if event.Scope == "stage" && event.Name != "bootstrap" {
			doc.StageCounts[event.Name]++
		}
		countTask006EventMetadata(event, doc.MetadataCoverage)
	}
	if analysis, analysisErr := analyzePublicE32Events(raw.Events); analysisErr != nil {
		addAnomaly("Batch025 root-partitioned analyzer rejected the Task006 event tree: %v", analysisErr)
	} else {
		rootSequence := task006BootstrapRootSequence(analysis.Closures)
		stages, stageErr := task006BootstrapStageTimings(analysis, rootSequence)
		if stageErr != nil {
			addAnomaly("cannot project nonoverlapping direct Bootstrap stage timings: %v", stageErr)
		}
		rescaleRoots := make([]task006RescaleRootTiming, 0, 35)
		for _, closure := range doc.RootClosures {
			if closure.Scope == "rescale" {
				rescaleRoots = append(rescaleRoots, task006RescaleRootTiming{
					Sequence: closure.Sequence, EventCount: closure.EventCount,
					InclusiveNS: closure.InclusiveNS, ChildrenNS: closure.ChildrenNS, ResidualNS: closure.SignedResidual,
				})
			}
		}
		if len(rescaleRoots) != 35 {
			addAnomaly("Task006 expected 35 independent Rescale diagnostic roots, got %d", len(rescaleRoots))
		}
		doc.Analysis = &task006OfflineAnalysis{
			RootElapsedNS: analysis.RootElapsedNS, RootImmediateChildrenNS: analysis.RootImmediateChildrenNS,
			RootUnattributedNS: analysis.RootUnattributedNS, RootUnattributedFraction: analysis.RootUnattributedFraction,
			EventTypeRootRecords: len(analysis.EventTypeSummaries), PositiveExclusiveGroups: len(analysis.PositiveExclusiveByType),
			AmdahlScenarioCount: len(analysis.AmdahlScenarios), BootstrapStages: stages, RescaleDiagnosticRoots: rescaleRoots,
		}
	}
	doc.RecursivePowerAncestry = task006RecursivePowerAncestry(raw.Events)
	if len(doc.RecursivePowerAncestry) == 0 {
		addAnomaly("known Task006 recursive power ancestry T16→T8→T4→T2 is absent")
	}
	if doc.Analysis != nil && len(doc.ValidationAnomalies) == 0 {
		doc.Classification = "OFFLINE_RECURSIVE_TRACE_REVALIDATED"
	} else {
		doc.Classification = "OFFLINE_PARTIAL_UNVERIFIED"
	}
	return doc, renderPublicE32Task006OfflineJournal(doc), nil
}

func findTask006RootClosure(closures []publicE32OfflineRootClosure, sequence uint64) publicE32OfflineRootClosure {
	for _, closure := range closures {
		if closure.Sequence == sequence {
			return closure
		}
	}
	return publicE32OfflineRootClosure{}
}

func task006BootstrapRootSequence(closures []publicE32EventClosure) uint64 {
	for _, closure := range closures {
		if closure.ParentSequence == 0 && closure.Scope == "stage" && closure.Name == "bootstrap" {
			return closure.Sequence
		}
	}
	return 0
}

func task006BootstrapStageTimings(analysis publicE32TraceAnalysis, rootSequence uint64) ([]task006StageTiming, error) {
	if rootSequence == 0 || analysis.RootElapsedNS <= 0 {
		return nil, errors.New("missing positive Bootstrap root timing")
	}
	rows := make([]task006StageTiming, 0, 10)
	seen := make(map[string]bool)
	var directStageTotal float64
	for _, summary := range analysis.EventTypeSummaries {
		if summary.TreeRootSequence != rootSequence || summary.Scope != "stage" || summary.ParentType != "stage/bootstrap" || summary.Name == "bootstrap" {
			continue
		}
		if summary.Count != 1 || seen[summary.Name] || summary.InclusiveMedianNS <= 0 {
			return nil, fmt.Errorf("direct stage %q has count=%d duplicate=%t or invalid timing", summary.Name, summary.Count, seen[summary.Name])
		}
		seen[summary.Name] = true
		share := summary.InclusiveMedianNS / analysis.RootElapsedNS
		rows = append(rows, task006StageTiming{Name: summary.Name, Count: summary.Count, InclusiveNS: summary.InclusiveMedianNS, Share: share})
		directStageTotal += summary.InclusiveMedianNS
	}
	for _, required := range []string{"coeffs_to_slots", "evalmod_real", "evalmod_imag", "slots_to_coeffs", "public_finalization"} {
		if !seen[required] {
			return nil, fmt.Errorf("required direct Bootstrap stage %q is missing", required)
		}
	}
	if directStageTotal > analysis.RootElapsedNS {
		return nil, fmt.Errorf("direct Bootstrap stage spans total %.0f ns exceed root %.0f ns; nonoverlap is not established", directStageTotal, analysis.RootElapsedNS)
	}
	return rows, nil
}

func normalizedEvidencePath(root, path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(absolute)
}

func validateTask006RawRuns(raw publicE32RawTrace, summary task006OfflineSummaryInput, doc *task006OfflineEvidence, add func(string, ...any)) {
	wantRawPhases := []string{"first_cold_bootstrap", "warm_bootstrap_01"}
	wantSummaryPhases := []string{"first_cold_bootstrap", "warm_bootstrap_01_traced"}
	if len(raw.Runs) != 2 || len(summary.Runs) != 2 {
		add("Task006 output metadata count is raw=%d summary=%d; want two each", len(raw.Runs), len(summary.Runs))
		return
	}
	for index := range raw.Runs {
		run, recorded := raw.Runs[index], summary.Runs[index]
		scaleLog2, ok := recordedTask006ScaleLog2(run.Scale)
		if run.Index != index+1 || run.Phase != wantRawPhases[index] || recorded.Phase != wantSummaryPhases[index] ||
			run.Level != recorded.Level || run.Degree != recorded.Degree || run.QPrefixRows != recorded.Rows || !ok || math.Abs(scaleLog2-45) > 1e-9 || recorded.Scale != "2^45" ||
			run.OracleRMSE != recorded.OracleRMSE || run.OracleMaxComplex != recorded.OracleMax || run.FastReferenceWorstRMSE != recorded.ReferenceRMSE ||
			run.FastReferenceWorstMax != recorded.ReferenceMax || run.OracleMaxComplex > raw.NumericalGate || run.FastReferenceWorstMax > raw.NumericalGate || !recorded.GatePassed {
			add("Task006 run %d raw and compact output metadata disagree or fail recorded state/numerical gates", index+1)
		}
		doc.RunChecks = append(doc.RunChecks, task006OfflineRunCheck{
			Phase: recorded.Phase, Level: recorded.Level, Degree: recorded.Degree, ScaleLog2: scaleLog2,
			QPrefixRows: recorded.Rows, OracleRMSE: recorded.OracleRMSE, OracleMax: recorded.OracleMax,
			ReferenceRMSE: recorded.ReferenceRMSE, ReferenceMax: recorded.ReferenceMax,
			Pass: recorded.GatePassed && recorded.OracleMax <= raw.NumericalGate && recorded.ReferenceMax <= raw.NumericalGate,
		})
	}
	rootFound := false
	for _, event := range raw.Events {
		if event.Scope == "stage" && event.Name == "bootstrap" && event.ParentSequence == 0 {
			rootFound = true
			if event.LevelOut == nil || *event.LevelOut != raw.Runs[1].Level || event.DegreeOut == nil || *event.DegreeOut != raw.Runs[1].Degree ||
				event.RowsOut == nil || *event.RowsOut != raw.Runs[1].QPrefixRows || event.ScaleLog2Out == nil || math.Abs(*event.ScaleLog2Out-45) > 1e-9 {
				add("Task006 traced Bootstrap root output state does not match its recorded warm output")
			}
		}
	}
	if !rootFound {
		add("Task006 trace has no root Bootstrap event for output metadata comparison")
	}
}

func recordedTask006ScaleLog2(scale string) (float64, bool) {
	value, _, err := big.ParseFloat(scale, 10, 256, big.ToNearestEven)
	if err != nil || value.Sign() <= 0 {
		return 0, false
	}
	got, _ := value.Float64()
	return math.Log2(got), !math.IsNaN(got) && !math.IsInf(got, 0)
}

func countTask006EventMetadata(event Event, counts map[string]int) {
	for _, item := range []struct {
		key string
		ok  bool
	}{
		{"level_in", event.LevelIn != nil}, {"level_out", event.LevelOut != nil},
		{"rows_in", event.RowsIn != nil}, {"rows_out", event.RowsOut != nil},
		{"scale_log2_in", event.ScaleLog2In != nil}, {"scale_log2_out", event.ScaleLog2Out != nil},
		{"degree_in", event.DegreeIn != nil}, {"degree_out", event.DegreeOut != nil},
	} {
		if item.ok {
			counts[item.key]++
		}
	}
}

func task006RecursivePowerAncestry(events []Event) []uint64 {
	bySequence := make(map[uint64]Event, len(events))
	for _, event := range events {
		bySequence[event.Sequence] = event
	}
	leaf, exists := bySequence[63]
	if !exists || leaf.Power == nil || *leaf.Power != 2 {
		return nil
	}
	chain := []uint64{leaf.Sequence}
	for leaf.ParentSequence != 0 {
		leaf, exists = bySequence[leaf.ParentSequence]
		if !exists {
			return nil
		}
		chain = append(chain, leaf.Sequence)
		if leaf.Scope == "power" && leaf.Name == "generated_powers" {
			break
		}
	}
	if len(chain) != 5 || chain[len(chain)-1] != 59 {
		return nil
	}
	return chain
}

func renderPublicE32Task006OfflineSummary(doc task006OfflineEvidence) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Task006 LogN16/E32 offline recursive trace revalidation\n\n- Classification: `%s`\n- Original classification/status: `%s` / `%s` (unchanged)\n- New Bootstrap calls: **%d**\n- Event count: **%d**; independent event roots: **%d**\n- Generated-power branch roots: `%v`\n- Raw SHA-256: `%s`\n- Failure-sidecar SHA-256: `%s`\n- Task006 ledger SHA-256: `%s`\n- Captured Fast production / diagnostic adapter commits: `%s` / `%s`\n- Adapter patch digest matches committed patch: **%t**\n- Historical adapter commit object currently available: **%t**\n\n", doc.Classification, doc.OriginalClassification, doc.OriginalTraceStatus, doc.NewBootstrapCalls, doc.EventCount, doc.IndependentRootCount, doc.GeneratedPowerRoots, doc.RawSHA256, doc.SidecarSHA256, doc.LedgerSHA256, doc.ProductionFastCommit, doc.DiagnosticAdapterCommit, doc.DiagnosticPatchMatches, doc.DiagnosticCommitAvailable)
	if doc.Analysis != nil {
		fmt.Fprintf(&out, "- Bootstrap root timing: inclusive %.0f ns; direct children %.0f ns; signed residual %.0f ns (%.2f%%)\n- Existing analyzer detail counts: event-type/root records %d; positive-exclusive groups %d; Amdahl scenarios %d\n\n", doc.Analysis.RootElapsedNS, doc.Analysis.RootImmediateChildrenNS, doc.Analysis.RootUnattributedNS, 100*doc.Analysis.RootUnattributedFraction, doc.Analysis.EventTypeRootRecords, doc.Analysis.PositiveExclusiveGroups, doc.Analysis.AmdahlScenarioCount)
		out.WriteString("## Nonoverlapping direct Bootstrap stages\n\nThe table contains only direct child stage spans from the Bootstrap root. Nested power and Rescale instrumentation roots are reported separately below and are not added to this wall-time table.\n\n| Stage | Count | Inclusive (ns) | Inclusive (ms) | Share of Bootstrap root |\n|---|---:|---:|---:|---:|\n")
		for _, stage := range doc.Analysis.BootstrapStages {
			fmt.Fprintf(&out, "| %s | %d | %.0f | %.6f | %.2f%% |\n", stage.Name, stage.Count, stage.InclusiveNS, stage.InclusiveNS/1e6, 100*stage.Share)
		}
		out.WriteString("\n## Independent generated-power and Rescale diagnostic roots\n\nThese are separate instrumentation trees that overlap Bootstrap wall-time measurement; their durations must not be summed with each other or with direct stage spans.\n\nGenerated-power branches:\n\n| Branch | Root sequence | Inclusive (ns) | Inclusive (ms) | Direct children (ns) | Signed residual (ns) |\n|---|---:|---:|---:|---:|---:|\n")
		for _, root := range doc.GeneratedPowerRoots {
			fmt.Fprintf(&out, "| %s | %d | %.0f | %.6f | %.0f | %.0f |\n", root.Branch, root.Sequence, root.InclusiveNS, root.InclusiveNS/1e6, root.ChildrenNS, root.ResidualNS)
		}
		out.WriteString("\nRescale diagnostic roots:\n\n| Root sequence | Events | Inclusive (ns) | Inclusive (ms) | Direct children (ns) | Signed residual (ns) |\n|---:|---:|---:|---:|---:|---:|\n")
		for _, root := range doc.Analysis.RescaleDiagnosticRoots {
			fmt.Fprintf(&out, "| %d | %d | %.0f | %.6f | %.0f | %.0f |\n", root.Sequence, root.EventCount, root.InclusiveNS, root.InclusiveNS/1e6, root.ChildrenNS, root.ResidualNS)
		}
		out.WriteByte('\n')
	}
	out.WriteString("## Recorded numerical outputs (metadata only; no decrypt repeated)\n\n| Call | Level | Scale log2 | Degree | Q-prefix rows | Oracle RMSE | Oracle max | Task005 Fast-reference max | Gate |\n|---|---:|---:|---:|---:|---:|---:|---:|---|\n")
	for _, run := range doc.RunChecks {
		fmt.Fprintf(&out, "| %s | %d | %.0f | %d | %d | %.9g | %.9g | %.9g | %t |\n", run.Phase, run.Level, run.ScaleLog2, run.Degree, run.QPrefixRows, run.OracleRMSE, run.OracleMax, run.ReferenceMax, run.Pass)
	}
	out.WriteString("\n## Recursive power ancestry\n\nThe corrected shared validator accepts the captured recurrence chain `generated_powers#59 → T16#60 → T8#61 → T4#62 → T2#63`; the old direct-child-only rejection is not reinstated. Stage timings and event-type summaries are partitioned by independent roots, so nested inclusive times are not treated as disjoint work.\n\n")
	if len(doc.ValidationAnomalies) > 0 {
		out.WriteString("## Validation anomalies\n")
		for _, anomaly := range doc.ValidationAnomalies {
			fmt.Fprintf(&out, "- %s\n", anomaly)
		}
		out.WriteByte('\n')
	}
	out.WriteString("## Limitations\n")
	for _, limitation := range doc.Limitations {
		fmt.Fprintf(&out, "- %s\n", limitation)
	}
	return out.String()
}

func renderPublicE32Task006OfflineJournal(doc task006OfflineEvidence) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Task006 offline analysis journal\n\n- Generated: `%s`\n- Classification: `%s`\n- Original raw remains `%s`; relabeled: `%t`\n- New Bootstrap calls: `%d`\n- Captured events analyzed by maintained Batch025 analyzer: `%d`\n- Raw / sidecar / ledger SHA-256: `%s` / `%s` / `%s`\n- Task006 attempt accounting preserved: Task005 %d/%d, Task006 %d/%d, retries 0\n- Generated-power branch roots: `%v`\n- Recursive power ancestry sequences (leaf-to-root): `%v`\n- Root-partitioned analysis available: `%t`\n\n", doc.GeneratedAt.Format(time.RFC3339), doc.Classification, doc.OriginalTraceStatus, doc.OriginalRawRelabeled, doc.NewBootstrapCalls, doc.EventCount, doc.RawSHA256, doc.SidecarSHA256, doc.LedgerSHA256, 16, 16, doc.ActualCalls, doc.CallBudget, doc.GeneratedPowerRoots, doc.RecursivePowerAncestry, doc.Analysis != nil)
	if doc.Analysis != nil {
		fmt.Fprintf(&out, "- Root timing (inclusive/direct children/signed residual): %.0f / %.0f / %.0f ns; event-type/root records: %d; positive-exclusive groups: %d; Amdahl scenarios: %d\n", doc.Analysis.RootElapsedNS, doc.Analysis.RootImmediateChildrenNS, doc.Analysis.RootUnattributedNS, doc.Analysis.EventTypeRootRecords, doc.Analysis.PositiveExclusiveGroups, doc.Analysis.AmdahlScenarioCount)
	}
	fmt.Fprintf(&out, "- Validation anomalies: %d\n", len(doc.ValidationAnomalies))
	for _, anomaly := range doc.ValidationAnomalies {
		fmt.Fprintf(&out, "  - %s\n", anomaly)
	}
	return out.String()
}
