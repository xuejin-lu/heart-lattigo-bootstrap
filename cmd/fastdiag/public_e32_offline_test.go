package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePublicE32OfflineArgsRequiresFrozenInputs(t *testing.T) {
	args := []string{
		"--raw", "raw.json", "--sidecar", "failure.json", "--manifest", "manifest.json",
		"--vectors", "vectors.json", "--attempt-01", "attempt-01.json", "--attempt-02", "attempt-02.json",
		"--out-prefix", "results/offline",
	}
	opts, err := parsePublicE32OfflineArgs(args)
	require.NoError(t, err)
	require.Equal(t, "raw.json", opts.rawPath)
	require.Equal(t, "results/offline", opts.outPrefix)

	_, err = parsePublicE32OfflineArgs(args[:len(args)-2])
	require.ErrorContains(t, err, "--out-prefix")
	_, err = parsePublicE32OfflineArgs(append(args, "--trace", "all"))
	require.Error(t, err, "offline mode must reject online-trace flags")

	task006Args := []string{
		"--raw", "raw.json", "--sidecar", "failure.json", "--task006-summary", "task006.json", "--task006-ledger", "ledger.json",
		"--attempt-01", "attempt-01.json", "--attempt-02", "attempt-02.json", "--out-prefix", "results/task006-offline",
	}
	task006Opts, err := parsePublicE32OfflineArgs(task006Args)
	require.NoError(t, err)
	require.Equal(t, "task006.json", task006Opts.task006Summary)
	_, err = parsePublicE32OfflineArgs(append(task006Args, "--manifest", "batch025.json"))
	require.ErrorContains(t, err, "不接受 Batch025")
}

func TestReadPublicE32OfflineRawRejectsTampering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raw.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"tampered":true}`), 0o600))
	_, _, err := readPublicE32OfflineRaw(path)
	require.ErrorContains(t, err, "frozen raw trace integrity STOP")
}

func TestPublicE32OfflineEventOmissionIsReportedTogether(t *testing.T) {
	events := publicE32TestEvents()
	filtered := make([]Event, 0, len(events)-2)
	for _, event := range events {
		if event.Scope == "stage" && event.Name == "evalmod_imag" {
			continue
		}
		if event.Scope == "rescale" && event.Name == "ntt_montgomery_restore" && event.Component == "c1" {
			continue
		}
		filtered = append(filtered, event)
	}
	problems := strings.Join(collectPublicE32EventAnomalies(filtered), "\n")
	require.Contains(t, problems, "stage count mismatch")
	require.Contains(t, problems, "Rescale materialization")
}

func TestPublicE32OfflineOutputBudgetInconsistencyIsReported(t *testing.T) {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	raw := publicE32RawTrace{
		TraceStatus: "TRACE_UNVERIFIED", TraceValidationError: "raw event capture is not certified; a separate certified result is written only after all gates pass",
		Profile: "logn13-e32-public-native", Mode: "test-only-public-e32-fast", CallBudget: 3, ActualCalls: 2,
		NumericalGate: 1e-6, Runs: []publicE32Run{{Index: 1}, {Index: 2}},
	}
	validatePublicE32OfflineTraceStatus(raw, add)
	require.Contains(t, strings.Join(problems, "\n"), "budget/output metadata is inconsistent")
}

func TestPublicE32OfflineClassificationNeverCertifiesOrMutatesRaw(t *testing.T) {
	rawStatus := "TRACE_UNVERIFIED"
	classification := publicE32OfflineClassification(rawStatus, 0, true)
	require.Equal(t, "OFFLINE_REVALIDATED_TRACE", classification)
	require.Equal(t, "TRACE_UNVERIFIED", rawStatus, "offline replay cannot mutate the source raw status")
	require.Equal(t, "OFFLINE_PARTIAL_UNVERIFIED", publicE32OfflineClassification("TRACE_VALIDATED", 0, true),
		"a fake source certification must fail closed")
	require.Equal(t, "OFFLINE_PARTIAL_UNVERIFIED", publicE32OfflineClassification("TRACE_UNVERIFIED", 1, true))
}

func TestPublicE32OfflineReplayFrozenCaptureWithoutBootstrap(t *testing.T) {
	rawPath := os.Getenv("FASTDIAG_BATCH025_RAW")
	if rawPath == "" {
		t.Skip("set FASTDIAG_BATCH025_RAW and companion paths to revalidate the preserved offline fixture")
	}
	root, err := findPrimaryRoot(".")
	require.NoError(t, err)
	secondaryRoot := os.Getenv("FASTDIAG_BATCH025_SECONDARY_ROOT")
	require.NotEmpty(t, secondaryRoot)
	primaryMeta, err := repositoryMetadata(root, "")
	require.NoError(t, err)
	opts := publicE32OfflineOptions{
		rawPath: rawPath, sidecarPath: os.Getenv("FASTDIAG_BATCH025_SIDECAR"),
		manifestPath: os.Getenv("FASTDIAG_BATCH025_MANIFEST"), vectorsPath: os.Getenv("FASTDIAG_BATCH025_VECTORS"),
		attemptOne: os.Getenv("FASTDIAG_BATCH025_ATTEMPT_01"), attemptTwo: os.Getenv("FASTDIAG_BATCH025_ATTEMPT_02"),
	}
	for name, value := range map[string]string{
		"sidecar": opts.sidecarPath, "manifest": opts.manifestPath, "vectors": opts.vectorsPath,
		"attempt-01": opts.attemptOne, "attempt-02": opts.attemptTwo,
	} {
		require.NotEmpty(t, value, "missing FASTDIAG_BATCH025_%s path", strings.ToUpper(strings.ReplaceAll(name, "-", "_")))
	}
	doc, journal, err := replayPublicE32Offline(root, secondaryRoot, primaryMeta, opts)
	require.NoError(t, err)
	require.Equal(t, "OFFLINE_REVALIDATED_TRACE", doc.Classification)
	require.Equal(t, "TRACE_UNVERIFIED", doc.RawTraceStatus)
	require.False(t, doc.OriginalRawRelabeled)
	require.Equal(t, 0, doc.NewBootstrapCalls)
	require.Equal(t, 539, doc.EventCount)
	require.Empty(t, doc.ValidationAnomalies)
	require.Contains(t, journal, "New Standard/Fast Bootstrap calls: **0**")
}

func TestTask006OfflineReplayCapturedTraceWithoutBootstrap(t *testing.T) {
	rawPath := os.Getenv("FASTDIAG_TASK006_RAW")
	if rawPath == "" {
		t.Skip("set FASTDIAG_TASK006_RAW and FASTDIAG_TASK006_SIDECAR to revalidate the preserved 539-event Task006 capture")
	}
	root, err := findPrimaryRoot(".")
	require.NoError(t, err)
	secondaryRoot, err := resolveSecondaryRoot(root)
	require.NoError(t, err)
	primaryMeta, err := repositoryMetadata(root, "")
	require.NoError(t, err)
	opts := publicE32OfflineOptions{
		rawPath: rawPath, sidecarPath: os.Getenv("FASTDIAG_TASK006_SIDECAR"),
		task006Summary: filepath.Join(root, "results", task006SummaryFile),
		task006Ledger:  filepath.Join(root, "results", "FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006-attempt-ledger.json"),
		attemptOne:     filepath.Join(root, "results", "FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.json.diagnostic-fast.bootstrap-attempt-01.json"),
		attemptTwo:     filepath.Join(root, "results", "FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.json.diagnostic-fast.bootstrap-attempt-02.json"),
	}
	require.NotEmpty(t, opts.sidecarPath)
	doc, journal, err := replayPublicE32Task006Offline(root, secondaryRoot, primaryMeta, opts)
	require.NoError(t, err)
	require.Empty(t, doc.ValidationAnomalies)
	require.Equal(t, "OFFLINE_RECURSIVE_TRACE_REVALIDATED", doc.Classification)
	require.Equal(t, "TRACE_UNVERIFIED", doc.OriginalTraceStatus)
	require.False(t, doc.OriginalRawRelabeled)
	require.Zero(t, doc.NewBootstrapCalls)
	require.Equal(t, 539, doc.EventCount)
	require.Equal(t, []task006PowerRoot{{Branch: "real", Sequence: 59}, {Branch: "imag", Sequence: 279}}, doc.GeneratedPowerRoots)
	require.Equal(t, []uint64{63, 62, 61, 60, 59}, doc.RecursivePowerAncestry)
	require.Equal(t, 16, doc.EffectiveParameters.LogN)
	require.Equal(t, 15, doc.EffectiveParameters.LogSlots)
	require.True(t, doc.DiagnosticPatchMatches)
	require.NotNil(t, doc.Analysis)
	require.Equal(t, 539, doc.Analysis.EventTypeRootRecords)
	encoded, err := json.Marshal(doc)
	require.NoError(t, err)
	require.Less(t, len(encoded), 64*1024, "offline evidence must remain compact and omit per-event analyzer arrays")
	require.NotContains(t, string(encoded), "positive_exclusive_top10_by_event_type")
	require.Contains(t, journal, "New Bootstrap calls: `0`")
}
