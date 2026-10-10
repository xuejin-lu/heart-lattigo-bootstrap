package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const e32ProductionFastCommit = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac"
const e32FixturePrimaryCommit = "6e938918442c409faa6d32e159c7a3f7a1041f7a"

const e32AuthorizedInstrumentationPath = "schemes/ckks/internal/fastcore/rescale.go"

type publicE32FixtureManifest struct {
	SchemaVersion         string               `json:"schema_version"`
	Profile               string               `json:"profile"`
	Backend               string               `json:"backend"`
	BackendCommit         string               `json:"backend_commit"`
	PrimaryCommit         string               `json:"primary_commit"`
	PrimarySourceSHA256   string               `json:"primary_measurement_source_sha256"`
	ConfigSHA256          string               `json:"config_sha256"`
	QPSHA256              string               `json:"qp_sha256"`
	InputSHA256           string               `json:"canonical_input_sha256"`
	WorkloadSHA256        string               `json:"workload_sha256"`
	ParametersFile        string               `json:"parameters_file"`
	ParametersSHA256      string               `json:"parameters_sha256"`
	CiphertextFile        string               `json:"ciphertext_file"`
	CiphertextSHA256      string               `json:"ciphertext_sha256"`
	DecodedSHA256         string               `json:"decoded_sha256"`
	State                 publicE32CipherState `json:"ciphertext_state"`
	NumericalGate         float64              `json:"max_complex_numerical_gate"`
	EphemeralSecretWeight int                  `json:"ephemeral_secret_weight"`
}

type publicE32CipherState struct {
	Level       int    `json:"level"`
	Degree      int    `json:"degree"`
	Scale       string `json:"scale"`
	QPrefixRows int    `json:"q_prefix_rows"`
	PrefixQ     string `json:"prefix_q"`
}

type publicE32Vector struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

type publicE32Vectors struct {
	SchemaVersion    string                       `json:"schema_version"`
	Mode             string                       `json:"mode"`
	Backend          string                       `json:"backend"`
	BackendCommit    string                       `json:"backend_commit"`
	PrimaryCommit    string                       `json:"primary_commit"`
	SourceSHA256     string                       `json:"primary_measurement_source_sha256"`
	ConfigSHA256     string                       `json:"config_sha256"`
	QPSHA256         string                       `json:"qp_sha256"`
	InputSHA256      string                       `json:"input_sha256"`
	WorkloadSHA256   string                       `json:"workload_sha256"`
	Checkpoints      map[string][]publicE32Vector `json:"checkpoints"`
	BootstrapOutputs map[string][]publicE32Vector `json:"bootstrap_outputs"`
}

type publicE32Run struct {
	Index                  int     `json:"index"`
	Phase                  string  `json:"phase"`
	ElapsedNS              int64   `json:"elapsed_ns"`
	DecodedSHA256          string  `json:"decoded_sha256"`
	OracleRMSE             float64 `json:"oracle_complex_rmse"`
	OracleMaxComplex       float64 `json:"oracle_max_complex_difference"`
	FastReferenceWorstRMSE float64 `json:"fast_reference_worst_rmse"`
	FastReferenceWorstMax  float64 `json:"fast_reference_worst_max_complex_difference"`
	Level                  int     `json:"level"`
	Degree                 int     `json:"degree"`
	Scale                  string  `json:"scale"`
	QPrefixRows            int     `json:"q_prefix_rows"`
}

type publicE32RawTrace struct {
	SchemaVersion         string         `json:"schema_version"`
	Timestamp             time.Time      `json:"timestamp"`
	TraceStatus           string         `json:"trace_status"`
	TraceValidationError  string         `json:"trace_validation_error,omitempty"`
	RawEvidenceFile       string         `json:"raw_unverified_file,omitempty"`
	RawEvidenceSHA256     string         `json:"raw_unverified_sha256,omitempty"`
	Profile               string         `json:"profile"`
	Mode                  string         `json:"mode"`
	PrimaryCommit         string         `json:"primary_commit"`
	FixturePrimaryCommit  string         `json:"fixture_primary_commit"`
	PrimarySourceSHA256   string         `json:"primary_measurement_source_sha256"`
	ProductionFastCommit  string         `json:"production_fast_commit"`
	DiagnosticHead        string         `json:"diagnostic_head"`
	ConfigSHA256          string         `json:"config_sha256"`
	QPSHA256              string         `json:"qp_sha256"`
	InputSHA256           string         `json:"input_sha256"`
	WorkloadSHA256        string         `json:"workload_sha256"`
	FixtureManifestSHA256 string         `json:"fixture_manifest_sha256"`
	CiphertextSHA256      string         `json:"ciphertext_sha256"`
	ExpectedInputSHA256   string         `json:"expected_input_decoded_sha256"`
	CallBudget            int            `json:"call_budget"`
	ActualCalls           int            `json:"actual_calls"`
	CPUProfileFile        string         `json:"warm_cpu_profile_file"`
	CPUProfileSHA256      string         `json:"warm_cpu_profile_sha256"`
	HeapProfileFile       string         `json:"post_warm_heap_profile_file"`
	HeapProfileSHA256     string         `json:"post_warm_heap_profile_sha256"`
	NumericalGate         float64        `json:"max_complex_gate"`
	GoVersion             string         `json:"go_version"`
	OS                    string         `json:"os"`
	Arch                  string         `json:"arch"`
	NumCPU                int            `json:"num_cpu"`
	GOMAXPROCS            int            `json:"gomaxprocs"`
	GOGC                  string         `json:"gogc"`
	GOMEMLIMIT            string         `json:"gomemlimit"`
	GODEBUG               string         `json:"godebug"`
	Runs                  []publicE32Run `json:"runs"`
	Events                []Event        `json:"warm_traced_events"`
}

type publicE32FailureArtifact struct {
	SchemaVersion        string    `json:"schema_version"`
	Timestamp            time.Time `json:"timestamp"`
	TraceStatus          string    `json:"trace_status"`
	FailurePhase         string    `json:"failure_phase"`
	FailureReason        string    `json:"failure_reason"`
	RawEvidenceFile      string    `json:"raw_unverified_file"`
	RawEvidenceSHA256    string    `json:"raw_unverified_sha256"`
	PrimaryCommit        string    `json:"primary_commit"`
	FixturePrimaryCommit string    `json:"fixture_primary_commit"`
	ProductionFastCommit string    `json:"production_fast_commit"`
	DiagnosticHead       string    `json:"diagnostic_head"`
	CallBudget           int       `json:"call_budget"`
	ActualCalls          int       `json:"actual_calls"`
}

type publicE32TraceDocument struct {
	SchemaVersion             string                 `json:"schema_version"`
	Timestamp                 time.Time              `json:"timestamp"`
	Profile                   string                 `json:"profile"`
	Trace                     []string               `json:"trace"`
	Primary                   RepositoryMetadata     `json:"primary_repository"`
	Secondary                 RepositoryMetadata     `json:"diagnostic_secondary_repository"`
	ProductionFastCommit      string                 `json:"production_fast_commit"`
	FixturePrimaryCommit      string                 `json:"fixture_primary_commit"`
	ProductionGoSourceSHA256  string                 `json:"production_go_source_sha256"`
	DiagnosticGoSourceSHA256  string                 `json:"diagnostic_go_source_sha256"`
	ProductionSourceIdentical bool                   `json:"production_go_source_identical"`
	InstrumentationOnlyDelta  bool                   `json:"authorized_instrumentation_only_delta"`
	ProductionSourceDiffPaths []string               `json:"production_source_diff_paths"`
	FixtureManifestSHA256     string                 `json:"fixture_manifest_sha256"`
	CiphertextSHA256          string                 `json:"ciphertext_sha256"`
	FastVectorsSHA256         string                 `json:"fast_vectors_sha256"`
	RawTracePath              string                 `json:"raw_trace_path"`
	RawTraceSHA256            string                 `json:"raw_trace_sha256"`
	CertifiedTracePath        string                 `json:"certified_trace_path"`
	CertifiedTraceSHA256      string                 `json:"certified_trace_sha256"`
	FailureArtifactPath       string                 `json:"failure_artifact_path"`
	CPUProfilePath            string                 `json:"warm_cpu_profile_path"`
	CPUProfileSHA256          string                 `json:"warm_cpu_profile_sha256"`
	CPUProfileAttribution     string                 `json:"cpu_profile_symbol_attribution"`
	HeapProfilePath           string                 `json:"post_warm_heap_profile_path"`
	HeapProfileSHA256         string                 `json:"post_warm_heap_profile_sha256"`
	CallBudget                int                    `json:"call_budget"`
	ActualCalls               int                    `json:"actual_calls"`
	NumericalGate             float64                `json:"max_complex_gate"`
	Environment               EnvironmentMetadata    `json:"environment"`
	GCEnvironment             publicE32GCEnvironment `json:"gc_environment"`
	Runs                      []publicE32Run         `json:"runs"`
	EventMedians              []EventMedian          `json:"event_medians"`
	Analysis                  publicE32TraceAnalysis `json:"event_analysis"`
}

type publicE32GCEnvironment struct {
	GOGC       string `json:"gogc"`
	GOMEMLIMIT string `json:"gomemlimit"`
	GODEBUG    string `json:"godebug"`
}

func tracePublicE32(primaryRoot, secondaryRoot string, primary RepositoryMetadata, opts options, outputs outputPaths) (publicE32TraceDocument, error) {
	if runtime.Version() != "go1.26.4" || runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" ||
		runtime.NumCPU() != 10 || runtime.GOMAXPROCS(0) != 10 || cpuModel() != "Apple M4" {
		return publicE32TraceDocument{}, fmt.Errorf("public E32 trace environment mismatch: Go=%s OS/Arch=%s/%s CPU=%s NumCPU=%d GOMAXPROCS=%d; require go1.26.4 darwin/arm64 Apple M4 NumCPU=10 GOMAXPROCS=10",
			runtime.Version(), runtime.GOOS, runtime.GOARCH, cpuModel(), runtime.NumCPU(), runtime.GOMAXPROCS(0))
	}
	if primary.Dirty {
		return publicE32TraceDocument{}, errors.New("public E32 trace requires a clean Primary worktree")
	}
	if primary.Commit == "" {
		return publicE32TraceDocument{}, errors.New("public E32 trace requires Primary commit provenance")
	}
	primaryBranch, err := gitOutput(primaryRoot, "branch", "--show-current")
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	if primaryBranch != "main" {
		return publicE32TraceDocument{}, fmt.Errorf("public E32 trace requires Primary main, got %q", primaryBranch)
	}
	remotePrimary, err := gitOutput(primaryRoot, "rev-parse", "origin/main")
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	if err := gitRun(primaryRoot, "merge-base", "--is-ancestor", remotePrimary, primary.Commit); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("Primary HEAD %s is not a fast-forward descendant of origin/main %s", primary.Commit, remotePrimary)
	}
	branch, err := gitOutput(secondaryRoot, "branch", "--show-current")
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	if branch != "fast-qprefix" {
		return publicE32TraceDocument{}, fmt.Errorf("public E32 diagnostic extension requires Secondary fast-qprefix, got %q", branch)
	}
	remoteSecondary, err := gitOutput(secondaryRoot, "rev-parse", "origin/fast-qprefix")
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	secondaryHead, err := gitOutput(secondaryRoot, "rev-parse", "HEAD")
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	if remoteSecondary != secondaryHead {
		return publicE32TraceDocument{}, fmt.Errorf("Secondary diagnostic commit %s is not the pushed origin/fast-qprefix HEAD %s", secondaryHead, remoteSecondary)
	}
	if err := requireCleanWorktree(secondaryRoot); err != nil {
		return publicE32TraceDocument{}, err
	}
	secondary, err := repositoryMetadata(secondaryRoot, branch)
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	productionHash, diagnosticHash, sourceDiffPaths, instrumentationOnlyDelta, err := verifyE32ProductionSourceIdentity(secondaryRoot, secondary.Commit)
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	measurementSourceHash, err := primaryMeasurementSourceFingerprint(primaryRoot, primary.Commit)
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	manifestBytes, manifest, vectorsBytes, vectors, err := validatePublicE32Fixture(opts.e32Manifest, opts.fastVectors, e32FixturePrimaryCommit, measurementSourceHash)
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	if err := gitRun(primaryRoot, "merge-base", "--is-ancestor", manifest.PrimaryCommit, primary.Commit); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("held E32 fixture source commit %s is not an ancestor of current Primary harness %s", manifest.PrimaryCommit, primary.Commit)
	}
	manifestSum := sha256.Sum256(manifestBytes)
	manifestHash := hex.EncodeToString(manifestSum[:])
	vectorsSum := sha256.Sum256(vectorsBytes)
	vectorsHash := hex.EncodeToString(vectorsSum[:])

	journalBase := outputs.json + ".diagnostic-fast"
	tempRoot, err := os.MkdirTemp("", "fastdiag-public-e32-")
	if err != nil {
		return publicE32TraceDocument{}, err
	}
	cachePath := filepath.Join(tempRoot, "gocache")
	if err := os.Mkdir(cachePath, 0o700); err != nil {
		return publicE32TraceDocument{}, err
	}
	profileDir := filepath.Join(tempRoot, "profiles")
	if err := os.Mkdir(profileDir, 0o700); err != nil {
		return publicE32TraceDocument{}, err
	}
	diagnosticRoot := filepath.Join(tempRoot, "secondary-diagnostic")
	if err := createPublicE32IsolatedClone(secondaryRoot, diagnosticRoot, secondary.Commit); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("create task-local pinned diagnostic clone (temp retained at %s): %w", tempRoot, err)
	}
	diagnosticMeta, err := repositoryMetadata(diagnosticRoot, "diagnostic-detached:"+secondary.Commit[:12])
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("inspect isolated diagnostic checkout %s: %w", diagnosticRoot, err)
	}
	if diagnosticMeta.Dirty {
		return publicE32TraceDocument{}, fmt.Errorf("isolated diagnostic checkout unexpectedly dirty; preserved at %s", diagnosticRoot)
	}
	productionHash, diagnosticHash, sourceDiffPaths, instrumentationOnlyDelta, err = verifyE32ProductionSourceIdentity(diagnosticRoot, diagnosticMeta.Commit)
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("isolated diagnostic checkout fails production-source identity: %w", err)
	}
	rawPath := filepath.Join(tempRoot, "raw-trace-unverified.json")
	certifiedPath := filepath.Join(tempRoot, "raw-trace-certified.json")
	failurePath := filepath.Join(tempRoot, "trace-failure.json")
	if err := runPublicE32TraceTest(diagnosticRoot, certifiedPath, rawPath, failurePath, cachePath, profileDir, manifestPath(opts.e32Manifest), manifestPath(opts.fastVectors), diagnosticMeta.Commit, primary.Commit, true); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("zero-call E32 fixture compatibility validation failed before diagnostic reservations; no Bootstrap call was launched; preserve checkout %s and artifacts %s: %w", diagnosticRoot, tempRoot, err)
	}
	if err := runReservedPublicE32Trace(journalBase, func() error {
		return runPublicE32TraceTest(diagnosticRoot, certifiedPath, rawPath, failurePath, cachePath, profileDir, manifestPath(opts.e32Manifest), manifestPath(opts.fastVectors), diagnosticMeta.Commit, primary.Commit, false)
	}); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("E32 tracer failed; both diagnostic attempt tokens remain reserved; preserve checkout %s, raw unverified trace %s, failure record %s, and artifacts %s: %w", diagnosticRoot, rawPath, failurePath, tempRoot, err)
	}
	rawBytes, err := os.ReadFile(rawPath)
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("read E32 raw trace in preserved temp directory %s: %w", tempRoot, err)
	}
	var unverified, raw publicE32RawTrace
	if err := json.Unmarshal(rawBytes, &unverified); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("decode E32 unverified raw trace in preserved temp directory %s: %w", tempRoot, err)
	}
	certifiedBytes, err := os.ReadFile(certifiedPath)
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("read separately certified E32 trace in preserved temp directory %s: %w", tempRoot, err)
	}
	if err := json.Unmarshal(certifiedBytes, &raw); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("decode separately certified E32 trace in preserved temp directory %s: %w", tempRoot, err)
	}
	if err := validatePublicE32RawCertification(unverified, raw, rawBytes); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("raw/certified E32 trace linkage failed in preserved temp directory %s: %w", tempRoot, err)
	}
	if err := validatePublicE32RawTrace(raw, manifest, vectors, primary.Commit, manifest.PrimaryCommit, secondary.Commit, manifestHash); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("invalid E32 trace artifact in preserved temp directory %s: %w", tempRoot, err)
	}
	analysis, err := analyzePublicE32Events(raw.Events)
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("analyze E32 event closure in preserved temp directory %s: %w", tempRoot, err)
	}
	gcEnvironment := publicE32GCEnvironment{GOGC: environmentSetting("GOGC"), GOMEMLIMIT: environmentSetting("GOMEMLIMIT"), GODEBUG: environmentSetting("GODEBUG")}
	if raw.GOGC != gcEnvironment.GOGC || raw.GOMEMLIMIT != gcEnvironment.GOMEMLIMIT || raw.GODEBUG != gcEnvironment.GODEBUG {
		return publicE32TraceDocument{}, fmt.Errorf("trace GC environment differs from parent process: trace=%+v parent=%+v", publicE32GCEnvironment{raw.GOGC, raw.GOMEMLIMIT, raw.GODEBUG}, gcEnvironment)
	}
	cpuProfilePath, err := readProfileArtifact(profileDir, raw.CPUProfileFile, raw.CPUProfileSHA256)
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("validate scoped warm CPU profile in preserved temp directory %s: %w", tempRoot, err)
	}
	heapProfilePath, err := readProfileArtifact(profileDir, raw.HeapProfileFile, raw.HeapProfileSHA256)
	if err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("validate post-warm heap profile in preserved temp directory %s: %w", tempRoot, err)
	}
	if err := os.RemoveAll(diagnosticRoot); err != nil {
		return publicE32TraceDocument{}, fmt.Errorf("clean task-local diagnostic clone %s: %w", diagnosticRoot, err)
	}
	profileAttribution := inspectE32CPUProfile(cpuProfilePath)
	rawSum := sha256.Sum256(rawBytes)
	certifiedSum := sha256.Sum256(certifiedBytes)
	doc := publicE32TraceDocument{
		SchemaVersion: "fastdiag.public-e32.summary.v1", Timestamp: raw.Timestamp, Profile: publicE32Profile,
		Trace: []string{"stage", "power", "rescale"}, Primary: primary, Secondary: diagnosticMeta,
		ProductionFastCommit: e32ProductionFastCommit, FixturePrimaryCommit: manifest.PrimaryCommit, ProductionGoSourceSHA256: productionHash,
		DiagnosticGoSourceSHA256: diagnosticHash, ProductionSourceIdentical: productionHash == diagnosticHash,
		InstrumentationOnlyDelta: instrumentationOnlyDelta, ProductionSourceDiffPaths: sourceDiffPaths,
		FixtureManifestSHA256: manifestHash, CiphertextSHA256: manifest.CiphertextSHA256,
		FastVectorsSHA256: vectorsHash, RawTracePath: rawPath, RawTraceSHA256: hex.EncodeToString(rawSum[:]),
		CertifiedTracePath: certifiedPath, CertifiedTraceSHA256: hex.EncodeToString(certifiedSum[:]), FailureArtifactPath: failurePath,
		CPUProfilePath: cpuProfilePath, CPUProfileSHA256: raw.CPUProfileSHA256,
		CPUProfileAttribution: profileAttribution,
		HeapProfilePath:       heapProfilePath, HeapProfileSHA256: raw.HeapProfileSHA256,
		CallBudget: raw.CallBudget, ActualCalls: raw.ActualCalls, NumericalGate: raw.NumericalGate,
		Environment:   EnvironmentMetadata{GoVersion: raw.GoVersion, OS: raw.OS, Arch: raw.Arch, CPU: cpuModel(), NumCPU: raw.NumCPU, GOMAXPROCS: raw.GOMAXPROCS},
		GCEnvironment: gcEnvironment,
		Runs:          raw.Runs, EventMedians: aggregateSingleTraceEventTypes(raw.Events), Analysis: analysis,
	}
	return doc, nil
}

func manifestPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func primaryMeasurementSourceFingerprint(root, commit string) (string, error) {
	paths := []string{"cmd/perfprobe", "internal/perfmeasure", "internal/numericalmetrics", "go.mod", "go.sum"}
	h := sha256.New()
	for _, path := range paths {
		object, err := gitOutput(root, "rev-parse", commit+":"+path)
		if err != nil {
			return "", fmt.Errorf("resolve Primary measurement source fingerprint %s at %s: %w", path, commit, err)
		}
		_, _ = h.Write([]byte(path + "\x00" + object + "\x00"))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func environmentSetting(name string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return "runtime-default"
}

func validatePublicE32Fixture(manifestPath, vectorsPath, fixturePrimaryCommit, measurementSourceHash string) ([]byte, publicE32FixtureManifest, []byte, publicE32Vectors, error) {
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("read E32 fixture manifest: %w", err)
	}
	var manifest publicE32FixtureManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("decode E32 fixture manifest: %w", err)
	}
	if manifest.SchemaVersion != "fast-public-e32-trace-fixture.v1" || manifest.Profile != "logn13-e32-public-native" ||
		manifest.Backend != "fast" || manifest.BackendCommit != e32ProductionFastCommit || manifest.PrimaryCommit != fixturePrimaryCommit ||
		manifest.State.Level != 0 || manifest.State.Degree != 1 || manifest.State.QPrefixRows != 1 || manifest.State.PrefixQ == "" ||
		manifest.NumericalGate != 1e-6 || manifest.EphemeralSecretWeight != 32 || manifest.DecodedSHA256 == "" ||
		manifest.PrimarySourceSHA256 == "" || manifest.ConfigSHA256 == "" || manifest.QPSHA256 == "" || manifest.InputSHA256 == "" ||
		manifest.WorkloadSHA256 == "" || manifest.ParametersSHA256 == "" || manifest.CiphertextSHA256 == "" {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, errors.New("fixture manifest fails the frozen public LogN13/E32/Fast Level0 provenance contract")
	}
	if manifest.PrimarySourceSHA256 != measurementSourceHash {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("current Primary measurement-source fingerprint %s differs from held fixture %s", measurementSourceHash, manifest.PrimarySourceSHA256)
	}
	fixtureDir := filepath.Dir(manifestPath)
	for _, item := range []struct{ name, digest string }{{manifest.ParametersFile, manifest.ParametersSHA256}, {manifest.CiphertextFile, manifest.CiphertextSHA256}} {
		if filepath.Base(item.name) != item.name || item.name == "." || item.name == "" {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("fixture filename is not a local basename: %q", item.name)
		}
		data, err := os.ReadFile(filepath.Join(fixtureDir, item.name))
		if err != nil {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("read fixture component %s: %w", item.name, err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != item.digest {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("fixture component %s SHA-256 mismatch", item.name)
		}
	}
	vectorsBytes, err := os.ReadFile(vectorsPath)
	if err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("read uninstrumented Fast vectors: %w", err)
	}
	var vectors publicE32Vectors
	if err := json.Unmarshal(vectorsBytes, &vectors); err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("decode uninstrumented Fast vectors: %w", err)
	}
	if vectors.SchemaVersion != "fast-standard-public-native-repeatability-vectors.v1" || vectors.Mode != "public-native-repeatability" ||
		vectors.Backend != "fast" || vectors.BackendCommit != e32ProductionFastCommit || vectors.PrimaryCommit != fixturePrimaryCommit ||
		vectors.SourceSHA256 != manifest.PrimarySourceSHA256 || vectors.ConfigSHA256 != manifest.ConfigSHA256 ||
		vectors.QPSHA256 != manifest.QPSHA256 || vectors.InputSHA256 != manifest.InputSHA256 || vectors.WorkloadSHA256 != manifest.WorkloadSHA256 {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, errors.New("uninstrumented Fast vectors do not match the E32 trace fixture provenance")
	}
	if len(vectors.BootstrapOutputs) != 6 {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("Fast repeatability vectors contain %d Bootstrap outputs, want exactly six", len(vectors.BootstrapOutputs))
	}
	for _, name := range []string{"bootstrap_first_cold", "bootstrap_warm_01", "bootstrap_warm_02", "bootstrap_warm_03", "bootstrap_warm_04", "bootstrap_warm_05"} {
		values := vectors.BootstrapOutputs[name]
		if len(values) != 1<<12 {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("Fast vector %s is absent or has the wrong slot count", name)
		}
		for index, value := range values {
			if math.IsNaN(value.Real) || math.IsInf(value.Real, 0) || math.IsNaN(value.Imag) || math.IsInf(value.Imag, 0) {
				return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("Fast vector %s has a non-finite slot at %d", name, index)
			}
		}
	}
	input, ok := vectors.Checkpoints["drop_level0"]
	if !ok || len(input) != 1<<12 {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, errors.New("Fast vectors do not contain the exact held drop_level0 input")
	}
	decoded := make([]complex128, len(input))
	for index, value := range input {
		if math.IsNaN(value.Real) || math.IsInf(value.Real, 0) || math.IsNaN(value.Imag) || math.IsInf(value.Imag, 0) {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("drop_level0 vector contains a non-finite slot at %d", index)
		}
		decoded[index] = complex(value.Real, value.Imag)
	}
	if perfmeasure.Fingerprint(decoded) != manifest.DecodedSHA256 {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, errors.New("fixture decoded-input hash does not match the held drop_level0 Fast vector")
	}
	return manifestBytes, manifest, vectorsBytes, vectors, nil
}

func verifyE32ProductionSourceIdentity(root, diagnosticHead string) (productionHash, diagnosticHash string, changedPaths []string, instrumentationOnly bool, err error) {
	if _, err := gitOutput(root, "cat-file", "-e", diagnosticHead+":circuits/ckks/bootstrapping/fastdiag_public_e32_test.go"); err != nil {
		return "", "", nil, false, errors.New("diagnostic Secondary commit does not contain the test-only public E32 adapter")
	}
	productionFiles, err := gitOutput(root, "ls-tree", "-r", "--name-only", e32ProductionFastCommit)
	if err != nil {
		return "", "", nil, false, fmt.Errorf("list pinned Fast production sources: %w", err)
	}
	diagnosticFiles, err := gitOutput(root, "ls-tree", "-r", "--name-only", diagnosticHead)
	if err != nil {
		return "", "", nil, false, fmt.Errorf("list diagnostic checkout sources: %w", err)
	}
	productionSources, err := productionGoSourceBlobs(root, e32ProductionFastCommit, productionFiles)
	if err != nil {
		return "", "", nil, false, err
	}
	diagnosticSources, err := productionGoSourceBlobs(root, diagnosticHead, diagnosticFiles)
	if err != nil {
		return "", "", nil, false, err
	}
	productionHash = hashProductionGoSources(productionSources)
	diagnosticHash = hashProductionGoSources(diagnosticSources)
	allPaths := make(map[string]struct{}, len(productionSources)+len(diagnosticSources))
	for path := range productionSources {
		allPaths[path] = struct{}{}
	}
	for path := range diagnosticSources {
		allPaths[path] = struct{}{}
	}
	for path := range allPaths {
		if productionSources[path] != diagnosticSources[path] {
			changedPaths = append(changedPaths, path)
		}
	}
	sort.Strings(changedPaths)
	if len(changedPaths) != 1 || changedPaths[0] != e32AuthorizedInstrumentationPath {
		return productionHash, diagnosticHash, changedPaths, false, fmt.Errorf("diagnostic production Go source differs outside the one authorized trace-instrumentation file: %v", changedPaths)
	}
	patch, err := gitOutput(root, "diff", "--no-ext-diff", "--unified=0", e32ProductionFastCommit, diagnosticHead, "--", e32AuthorizedInstrumentationPath)
	if err != nil {
		return productionHash, diagnosticHash, changedPaths, false, fmt.Errorf("inspect authorized Rescale instrumentation source delta: %w", err)
	}
	if err := validateE32InstrumentationOnlyPatch(patch); err != nil {
		return productionHash, diagnosticHash, changedPaths, false, err
	}
	for _, path := range []string{"go.mod", "go.sum"} {
		productionBlob, productionErr := gitOutput(root, "rev-parse", e32ProductionFastCommit+":"+path)
		diagnosticBlob, diagnosticErr := gitOutput(root, "rev-parse", diagnosticHead+":"+path)
		if productionErr != nil || diagnosticErr != nil || productionBlob != diagnosticBlob {
			return productionHash, diagnosticHash, changedPaths, false, fmt.Errorf("diagnostic checkout module manifest %s differs from frozen Fast pin", path)
		}
	}
	return productionHash, diagnosticHash, changedPaths, true, nil
}

func productionGoSourceBlobs(root, commit, fileList string) (map[string]string, error) {
	sources := make(map[string]string)
	for _, path := range strings.Split(fileList, "\n") {
		if path == "" || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			continue
		}
		blob, err := gitOutput(root, "rev-parse", commit+":"+path)
		if err != nil {
			return nil, fmt.Errorf("resolve production source %s at %s: %w", path, commit, err)
		}
		sources[path] = blob
	}
	return sources, nil
}

func hashProductionGoSources(sources map[string]string) string {
	entries := make([]string, 0, len(sources))
	for path, blob := range sources {
		entries = append(entries, path+" "+blob)
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}

func validateE32InstrumentationOnlyPatch(patch string) error {
	want := map[string]int{
		"var materializationSpan fastdiag.Span":                        1,
		"if fastdiag.Enabled && fastdiag.Selected(fastdiag.Rescale) {": 3,
		"materializationSpan = fastdiag.Begin(fastdiag.Rescale, \"materialization\", wholeSpan.Sequence(), fastdiag.Input(op0, sourceRows).Merge(fastdiag.InPlace(op0 == opOut)).Merge(fastdiag.RepetitionCount(len(op0.Value))))": 1,
		"var restoreSpan fastdiag.Span": 1,
		"restoreSpan = fastdiag.Begin(fastdiag.Rescale, \"ntt_montgomery_restore\", materializationSpan.Sequence(), rescaleDiagFields(op0.Level(), targetLevel, sourceRows, targetRows, fmt.Sprintf(\"c%d\", component), op0 == opOut))": 1,
		"restoreSpan.End(fastdiag.Fields{})":                          1,
		"materializationSpan.End(fastdiag.Output(opOut, targetRows))": 1,
		"}": 3,
	}
	seen := make(map[string]int, len(want))
	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "-") {
			return fmt.Errorf("authorized Rescale instrumentation patch removes existing source: %s", line)
		}
		if !strings.HasPrefix(line, "+") {
			continue
		}
		code := strings.TrimSpace(strings.TrimPrefix(line, "+"))
		if _, ok := want[code]; !ok {
			return fmt.Errorf("authorized Rescale patch contains non-instrumentation source addition: %s", code)
		}
		seen[code]++
	}
	for code, count := range want {
		if seen[code] != count {
			return fmt.Errorf("authorized Rescale instrumentation patch has %d occurrences of %q, want %d", seen[code], code, count)
		}
	}
	return nil
}

func reservePublicE32TraceAttempts(journalBase string) error {
	if journalBase == "" {
		return errors.New("E32 trace requires a durable attempt-journal base")
	}
	for index, phase := range []string{"diagnostic_fast_cold", "diagnostic_fast_traced_warm"} {
		reservation := struct {
			SchemaVersion string    `json:"schema_version"`
			Index         int       `json:"attempt_index"`
			Budget        int       `json:"bootstrap_budget"`
			Phase         string    `json:"phase"`
			ReservedAt    time.Time `json:"reserved_at"`
			Status        string    `json:"status"`
		}{"perfprobe-bootstrap-attempt.v1", index + 1, 2, phase, time.Now().UTC(), "irrevocably_reserved_before_call"}
		path := fmt.Sprintf("%s.bootstrap-attempt-%02d.json", journalBase, index+1)
		data, err := json.MarshalIndent(reservation, "", "  ")
		if err != nil {
			return err
		}
		if err := writeNewFile(path, append(data, '\n'), 0o600); err != nil {
			return fmt.Errorf("irrevocably reserve both diagnostic Fast attempts before launch: %w", err)
		}
	}
	return nil
}

func runReservedPublicE32Trace(journalBase string, run func() error) error {
	if run == nil {
		return errors.New("E32 trace subprocess callback is required")
	}
	if err := reservePublicE32TraceAttempts(journalBase); err != nil {
		return err
	}
	return run()
}

func runPublicE32TraceTest(secondaryRoot, certifiedPath, rawPath, failurePath, cachePath, profileDir, manifestPath, vectorsPath, diagnosticHead, primaryHead string, validateOnly bool) error {
	args := []string{"test", "-tags", "fastdiag", "./circuits/ckks/bootstrapping", "-run", "^TestFastDiagPublicE32Trace$", "-count=1"}
	command := exec.Command("go", args...)
	command.Dir = secondaryRoot
	validateOnlySetting := "0"
	if validateOnly {
		validateOnlySetting = "1"
	}
	command.Env = updateEnv(os.Environ(), map[string]string{
		"FASTDIAG_TRACE":                   "stage,power,rescale",
		"FASTDIAG_PUBLIC_E32_MANIFEST":     manifestPath,
		"FASTDIAG_PUBLIC_E32_FAST_VECTORS": vectorsPath,
		"FASTDIAG_DIAGNOSTIC_HEAD":         diagnosticHead,
		"FASTDIAG_PRIMARY_HEAD":            primaryHead,
		"FASTDIAG_OUTPUT":                  certifiedPath,
		"FASTDIAG_RAW_OUTPUT":              rawPath,
		"FASTDIAG_FAILURE_OUTPUT":          failurePath,
		"FASTDIAG_PROFILE_DIR":             profileDir,
		"FASTDIAG_VALIDATE_ONLY":           validateOnlySetting,
		"GOCACHE":                          cachePath,
		"GOMAXPROCS":                       fmt.Sprint(runtime.GOMAXPROCS(0)),
	})
	var output strings.Builder
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return fmt.Errorf("test-only public E32 tracer failed: %w\n%s", err, strings.TrimSpace(output.String()))
	}
	return nil
}

func validatePublicE32RawTrace(raw publicE32RawTrace, manifest publicE32FixtureManifest, vectors publicE32Vectors, primaryCommit, fixturePrimaryCommit, diagnosticHead, manifestHash string) error {
	if raw.SchemaVersion != "fastdiag.public-e32.trace.v1" || raw.TraceStatus != "TRACE_VALIDATED" || raw.TraceValidationError != "" ||
		raw.RawEvidenceFile == "" || filepath.Base(raw.RawEvidenceFile) != raw.RawEvidenceFile || len(raw.RawEvidenceSHA256) != sha256.Size*2 ||
		raw.Profile != "logn13-e32-public-native" || raw.Mode != "test-only-public-e32-fast" ||
		raw.PrimaryCommit != primaryCommit || raw.FixturePrimaryCommit != fixturePrimaryCommit || raw.FixturePrimaryCommit != manifest.PrimaryCommit || raw.PrimarySourceSHA256 != manifest.PrimarySourceSHA256 ||
		raw.ProductionFastCommit != e32ProductionFastCommit || raw.DiagnosticHead != diagnosticHead || raw.ConfigSHA256 != manifest.ConfigSHA256 ||
		raw.QPSHA256 != manifest.QPSHA256 || raw.InputSHA256 != manifest.InputSHA256 || raw.WorkloadSHA256 != manifest.WorkloadSHA256 ||
		raw.FixtureManifestSHA256 != manifestHash || raw.CiphertextSHA256 != manifest.CiphertextSHA256 || raw.ExpectedInputSHA256 != manifest.DecodedSHA256 ||
		raw.CallBudget != 2 || raw.ActualCalls != 2 || raw.NumericalGate != 1e-6 ||
		raw.GoVersion != "go1.26.4" || raw.OS != "darwin" || raw.Arch != "arm64" || raw.NumCPU != 10 || raw.GOMAXPROCS != 10 ||
		raw.CPUProfileFile == "" || raw.CPUProfileSHA256 == "" || raw.HeapProfileFile == "" || raw.HeapProfileSHA256 == "" ||
		raw.GOGC == "" || raw.GOMEMLIMIT == "" || raw.GODEBUG == "" {
		return errors.New("trace artifact provenance or exact 2-call budget is inconsistent with its fixture")
	}
	if len(raw.Runs) != 2 || raw.Runs[0].Index != 1 || raw.Runs[0].Phase != "first_cold_bootstrap" || raw.Runs[1].Index != 2 || raw.Runs[1].Phase != "warm_bootstrap_01" {
		return errors.New("trace artifact must contain exactly one cold and one traced warm public Bootstrap result")
	}
	for _, run := range raw.Runs {
		if run.ElapsedNS < 0 || run.DecodedSHA256 == "" || run.Level != 1 || run.Degree != 1 || run.QPrefixRows != 2 || run.Scale != manifest.State.Scale ||
			run.OracleRMSE < 0 || run.OracleMaxComplex < 0 || run.FastReferenceWorstRMSE < 0 || run.FastReferenceWorstMax < 0 ||
			math.IsNaN(run.OracleRMSE) || math.IsNaN(run.OracleMaxComplex) || math.IsNaN(run.FastReferenceWorstRMSE) || math.IsNaN(run.FastReferenceWorstMax) ||
			math.IsInf(run.OracleRMSE, 0) || math.IsInf(run.OracleMaxComplex, 0) || math.IsInf(run.FastReferenceWorstRMSE, 0) || math.IsInf(run.FastReferenceWorstMax, 0) ||
			run.OracleMaxComplex > raw.NumericalGate || run.FastReferenceWorstMax > raw.NumericalGate {
			return fmt.Errorf("trace output phase %s fails its decoded numerical/state gate", run.Phase)
		}
	}
	if len(vectors.BootstrapOutputs) != 6 {
		return errors.New("matched uninstrumented Fast repeatability evidence changed during trace validation")
	}
	return validatePublicE32Events(raw.Events)
}

func validatePublicE32RawCertification(unverified, certified publicE32RawTrace, rawBytes []byte) error {
	if unverified.TraceStatus != "TRACE_UNVERIFIED" || unverified.TraceValidationError == "" {
		return errors.New("raw event file must remain explicitly TRACE_UNVERIFIED with a validation-status note")
	}
	if certified.TraceStatus != "TRACE_VALIDATED" || certified.TraceValidationError != "" {
		return errors.New("separate certified file must be marked TRACE_VALIDATED only after all gates pass")
	}
	if certified.RawEvidenceFile != "raw-trace-unverified.json" || filepath.Base(certified.RawEvidenceFile) != certified.RawEvidenceFile {
		return errors.New("certified file does not point to the expected unverified raw evidence basename")
	}
	rawSum := sha256.Sum256(rawBytes)
	if certified.RawEvidenceSHA256 != hex.EncodeToString(rawSum[:]) {
		return errors.New("certified file raw-evidence SHA-256 does not match the preserved unverified trace")
	}
	if unverified.ActualCalls != certified.ActualCalls || unverified.CallBudget != certified.CallBudget ||
		unverified.PrimaryCommit != certified.PrimaryCommit || unverified.FixturePrimaryCommit != certified.FixturePrimaryCommit || unverified.ProductionFastCommit != certified.ProductionFastCommit ||
		unverified.PrimarySourceSHA256 != certified.PrimarySourceSHA256 || unverified.Profile != certified.Profile || unverified.Mode != certified.Mode ||
		unverified.DiagnosticHead != certified.DiagnosticHead || unverified.ConfigSHA256 != certified.ConfigSHA256 ||
		unverified.QPSHA256 != certified.QPSHA256 || unverified.InputSHA256 != certified.InputSHA256 ||
		unverified.WorkloadSHA256 != certified.WorkloadSHA256 || unverified.FixtureManifestSHA256 != certified.FixtureManifestSHA256 ||
		unverified.CiphertextSHA256 != certified.CiphertextSHA256 || unverified.ExpectedInputSHA256 != certified.ExpectedInputSHA256 ||
		unverified.NumericalGate != certified.NumericalGate || unverified.GoVersion != certified.GoVersion || unverified.OS != certified.OS ||
		unverified.Arch != certified.Arch || unverified.NumCPU != certified.NumCPU || unverified.GOMAXPROCS != certified.GOMAXPROCS ||
		unverified.GOGC != certified.GOGC || unverified.GOMEMLIMIT != certified.GOMEMLIMIT || unverified.GODEBUG != certified.GODEBUG ||
		unverified.CPUProfileFile != certified.CPUProfileFile || unverified.HeapProfileFile != certified.HeapProfileFile {
		return errors.New("unverified and certified E32 artifacts disagree on source, fixture, or call-budget provenance")
	}
	if len(unverified.Runs) != len(certified.Runs) || len(unverified.Events) != len(certified.Events) {
		return errors.New("unverified and certified E32 artifacts disagree on captured run/event counts")
	}
	if !reflect.DeepEqual(unverified.Runs, certified.Runs) || !reflect.DeepEqual(unverified.Events, certified.Events) {
		return errors.New("certified E32 artifact changes captured raw runs or event records")
	}
	return nil
}

func readProfileArtifact(dir, name, wantSHA string) (string, error) {
	if name == "" || filepath.Base(name) != name || len(wantSHA) != sha256.Size*2 {
		return "", errors.New("profile artifact filename or SHA-256 is invalid")
	}
	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != wantSHA {
		return "", fmt.Errorf("profile %s SHA-256 mismatch", name)
	}
	return path, nil
}

func validatePublicE32Events(events []Event) error {
	if len(events) == 0 {
		return errors.New("traced warm call contains no fastdiag events")
	}
	sequences := make(map[uint64]Event, len(events))
	children := make(map[uint64][]Event, len(events))
	var root *Event
	var stages []string
	generatedPowers := make(map[uint64]bool)
	powers := make(map[int]bool)
	rescaleCount := 0
	for _, event := range events {
		if event.Sequence == 0 || event.ElapsedNS < 0 {
			return errors.New("fastdiag event has zero sequence or negative elapsed time")
		}
		if event.Scope != "stage" && event.Scope != "power" && event.Scope != "rescale" {
			return fmt.Errorf("unexpected fastdiag scope %q", event.Scope)
		}
		if _, exists := sequences[event.Sequence]; exists {
			return fmt.Errorf("duplicate fastdiag event sequence %d", event.Sequence)
		}
		sequences[event.Sequence] = event
		children[event.ParentSequence] = append(children[event.ParentSequence], event)
		if event.Scope == "stage" {
			if event.Name == "bootstrap" {
				if root != nil || event.ParentSequence != 0 {
					return errors.New("fastdiag event tree must have exactly one root Bootstrap stage")
				}
				copy := event
				root = &copy
			} else {
				stages = append(stages, event.Name)
			}
		}
		if event.Scope == "power" && event.Name == "generated_powers" {
			generatedPowers[event.Sequence] = true
		}
		if event.Scope == "power" && event.Name == "power" && event.Power != nil {
			if event.SplitA == nil || event.SplitB == nil {
				return fmt.Errorf("generated power T%d has no recurrence split", *event.Power)
			}
			if powers[*event.Power] {
				return fmt.Errorf("duplicate generated power T%d", *event.Power)
			}
			powers[*event.Power] = true
		}
		if event.Scope == "rescale" && event.Name == "rescale" {
			rescaleCount++
		}
	}
	if root == nil {
		return errors.New("missing Bootstrap event root")
	}
	for _, event := range events {
		if event.ParentSequence != 0 {
			if _, exists := sequences[event.ParentSequence]; !exists {
				return fmt.Errorf("event %d references missing parent %d", event.Sequence, event.ParentSequence)
			}
		}
	}
	// A generated Chebyshev power can recursively generate lower powers.
	// Thus T8 may be a child of T16, which is itself under generated_powers.
	// Follow the actual ancestry instead of requiring every power to be a
	// direct child of generated_powers.
	for _, event := range events {
		if event.Scope != "power" || event.Name != "power" {
			continue
		}
		if event.Power == nil || event.SplitA == nil || event.SplitB == nil {
			return fmt.Errorf("power event %d lacks generated-power fields", event.Sequence)
		}
		ancestorSequence := event.ParentSequence
		seen := map[uint64]bool{event.Sequence: true}
		for {
			if ancestorSequence == 0 {
				return fmt.Errorf("power event T%d has no generated_powers ancestor", *event.Power)
			}
			if seen[ancestorSequence] {
				return fmt.Errorf("power event T%d has cyclic ancestry", *event.Power)
			}
			seen[ancestorSequence] = true
			ancestor, found := sequences[ancestorSequence]
			if !found {
				return fmt.Errorf("power event T%d has missing ancestor %d", *event.Power, ancestorSequence)
			}
			if ancestor.Scope != "power" {
				return fmt.Errorf("power event T%d has non-power ancestor %d", *event.Power, ancestorSequence)
			}
			if ancestor.Name == "generated_powers" {
				if !generatedPowers[ancestorSequence] {
					return fmt.Errorf("power event T%d has an unregistered root", *event.Power)
				}
				break
			}
			if ancestor.Name != "power" || ancestor.Power == nil || *ancestor.Power <= *event.Power {
				return fmt.Errorf("power event T%d has invalid recursive ancestor %d", *event.Power, ancestorSequence)
			}
			ancestorSequence = ancestor.ParentSequence
		}
	}
	wantStages := []string{"pack_n1_to_n2", "scale_down", "mod_up_trace", "coeffs_to_slots", "evalmod_real"}
	for _, name := range stages {
		if name == "evalmod_imag" {
			wantStages = append(wantStages, name)
		}
	}
	wantStages = append(wantStages, "slots_to_coeffs", "unpack_n2_to_n1", "public_finalization")
	if len(stages) != len(wantStages) {
		return fmt.Errorf("stage count mismatch: got %v want %v", stages, wantStages)
	}
	for index, name := range stages {
		if name != wantStages[index] {
			return fmt.Errorf("stage %d is %q, want %q", index, name, wantStages[index])
		}
		var stage Event
		for _, event := range events {
			if event.Scope == "stage" && event.Name == name {
				stage = event
				break
			}
		}
		if stage.ParentSequence != root.Sequence {
			return fmt.Errorf("stage %s is not a direct child of Bootstrap", name)
		}
	}
	for _, power := range []int{2, 3, 4, 6, 8, 16} {
		if !powers[power] {
			return fmt.Errorf("missing generated Chebyshev power T%d", power)
		}
	}
	if rescaleCount == 0 {
		return errors.New("trace has no Rescale parent events")
	}
	for _, event := range events {
		if event.Scope != "rescale" || event.Name != "rescale" {
			continue
		}
		passes := make(map[string]bool)
		for _, child := range children[event.Sequence] {
			passes[child.Name] = child.Scope == "rescale"
		}
		if len(passes) != 2 || !passes["preflight"] || !passes["materialization"] {
			return fmt.Errorf("Rescale event %d must have exactly preflight and materialization children", event.Sequence)
		}
	}
	return nil
}

func renderPublicE32Summary(doc publicE32TraceDocument) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Public E32 Fast diagnostic trace\n\n- Primary harness: `%s` (`%s`, dirty=%t)\n- Held fixture source Primary: `%s`\n- Diagnostic Secondary: `%s` (`%s`, dirty=%t)\n", doc.Primary.Commit, doc.Primary.Ref, doc.Primary.Dirty, doc.FixturePrimaryCommit, doc.Secondary.Commit, doc.Secondary.Ref, doc.Secondary.Dirty)
	fmt.Fprintf(&out, "- Frozen formal Fast production pin: `%s`\n- Original production Go-source SHA-256: `%s`\n- Diagnostic production Go-source SHA-256: `%s`\n- Production source identical: `%t`; authorized instrumentation-only delta: `%t`; changed production paths: `%v`\n- E32 trace budget: %d reserved / %d actual calls\n- Numerical gate: %.3g\n- Environment: `%s`, `%s/%s`, CPU `%s`, NumCPU `%d`, GOMAXPROCS `%d`; GC `%s`, limit `%s`, debug `%s`\n- Raw unverified trace (outside Git): `%s` (SHA-256 `%s`)\n- Separately certified trace (outside Git): `%s` (SHA-256 `%s`)\n- Failure sidecar path (created only on failure): `%s`\n- CPU profile (warm traced call only): `%s` (SHA-256 `%s`)\n- Heap profile (post-warm in-use snapshot, not allocation attribution): `%s` (SHA-256 `%s`)\n\n## CPU profile symbol views\n\n%s\n\n", doc.ProductionFastCommit, doc.ProductionGoSourceSHA256, doc.DiagnosticGoSourceSHA256, doc.ProductionSourceIdentical, doc.InstrumentationOnlyDelta, doc.ProductionSourceDiffPaths, doc.CallBudget, doc.ActualCalls, doc.NumericalGate, doc.Environment.GoVersion, doc.Environment.OS, doc.Environment.Arch, doc.Environment.CPU, doc.Environment.NumCPU, doc.Environment.GOMAXPROCS, doc.GCEnvironment.GOGC, doc.GCEnvironment.GOMEMLIMIT, doc.GCEnvironment.GODEBUG, doc.RawTracePath, doc.RawTraceSHA256, doc.CertifiedTracePath, doc.CertifiedTraceSHA256, doc.FailureArtifactPath, doc.CPUProfilePath, doc.CPUProfileSHA256, doc.HeapProfilePath, doc.HeapProfileSHA256, doc.CPUProfileAttribution)
	fmt.Fprintf(&out, "## Root closure and event topology\n\n- Bootstrap root: %.3f ms; direct children: %.3f ms; signed unknown/unattributed residual: %.3f ms (%.2f%%). Failure-status time is not separately emitted by this event schema, so no failure share is inferred. Residuals are measured as parent minus immediate children; no normalization or forced 100%% closure is applied.\n\n", doc.Analysis.RootElapsedNS/1e6, doc.Analysis.RootImmediateChildrenNS/1e6, doc.Analysis.RootUnattributedNS/1e6, doc.Analysis.RootUnattributedFraction*100)
	out.WriteString("| Seq | Parent seq | Scope | Event | Component | Inclusive (ms) | Direct children (ms) | Unattributed/exclusive (ms) | Event residual (%) |\n|---:|---:|---|---|---|---:|---:|---:|---:|\n")
	for _, closure := range doc.Analysis.Closures {
		fmt.Fprintf(&out, "| %d | %d | %s | %s | %s | %.3f | %.3f | %.3f | %.2f%% |\n", closure.Sequence, closure.ParentSequence, closure.Scope, closure.Name, displayOrDash(closure.Component), closure.InclusiveNS/1e6, closure.ImmediateChildrenNS/1e6, closure.UnattributedNS/1e6, closure.UnattributedFraction*100)
	}
	out.WriteString("\n## Positive exclusive-time Pareto by independent event root\n\nStage, Power and Rescale root trees are reported separately because Power/Rescale roots are not children of Bootstrap stages; their percentages must not be summed across trees.\n")
	for _, group := range doc.Analysis.PositiveExclusivePareto {
		fmt.Fprintf(&out, "\n### `%s/%s` root #%d (%.3f ms)\n\n| Rank | Scope | Event | Component | Exclusive (ms) | Share of this event-tree root |\n|---:|---|---|---|---:|---:|\n", group.RootScope, group.RootName, group.RootSequence, group.RootElapsedNS/1e6)
		for _, entry := range group.Entries {
			fmt.Fprintf(&out, "| %d | %s | %s | %s | %.3f | %s |\n", entry.Rank, entry.Scope, entry.Name, displayOrDash(entry.Component), entry.ExclusiveNS/1e6, formatE32Percent(entry.RootShare))
		}
	}
	out.WriteString("\n## Bounded Amdahl scenarios\n\nThese are hypothetical stage-only speedups computed from each observed inclusive stage/root fraction, not predicted end-to-end performance. The stage-eliminated value is an upper bound under the same simplified model.\n\n| Stage | Observed root fraction | Stage 2× | Stage 4× | If eliminated (upper bound) |\n|---|---:|---:|---:|---:|\n")
	for _, scenario := range doc.Analysis.AmdahlScenarios {
		fmt.Fprintf(&out, "| %s | %.2f%% | %s | %s | %s |\n", scenario.Stage, scenario.ObservedRootFraction*100, formatE32OptionalFloat(scenario.TwoXStageOverallSpeedup), formatE32OptionalFloat(scenario.FourXStageOverallSpeedup), formatE32OptionalFloat(scenario.StageEliminatedUpperBound))
	}
	out.WriteString("\n## Event medians and hierarchy\n\n`Samples` counts same event/parent-type occurrences within this single traced warm call; these are not repeated-run medians. `Parent key` preserves nesting instead of summing nested inclusive values as disjoint work.\n\n| Parent key | Scope | Event | Power | Component | Elapsed (ms) | Samples |\n|---|---|---|---|---|---:|---:|\n")
	for _, event := range doc.EventMedians {
		power := formatPower(event.Power, event.SplitA, event.SplitB)
		component := event.Component
		if component == "" {
			component = "—"
		}
		fmt.Fprintf(&out, "| %s | %s | %s | %s | %s | %.3f | %d |\n", displayOrDash(event.ParentKey), event.Scope, event.Name, power, component, event.MedianNS/1e6, event.Count)
	}
	return out.String()
}

func displayOrDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func formatE32OptionalFloat(value *float64) string {
	if value == nil {
		return "not defined"
	}
	return fmt.Sprintf("%.3fx", *value)
}

func formatE32Percent(value *float64) string {
	if value == nil {
		return "not defined"
	}
	return fmt.Sprintf("%.2f%%", *value*100)
}

func inspectE32CPUProfile(path string) string {
	var outputs []string
	for _, view := range []struct {
		name string
		args []string
	}{{"flat", []string{"-top", "-nodecount=12"}}, {"cumulative", []string{"-top", "-cum", "-nodecount=12"}}} {
		args := append([]string{"tool", "pprof"}, view.args...)
		args = append(args, path)
		command := exec.Command("go", args...)
		output, err := command.CombinedOutput()
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("%s: unavailable: go tool pprof: %v (%s)", view.name, err, truncateE32ProfileOutput(strings.TrimSpace(string(output)))))
			continue
		}
		outputs = append(outputs, view.name+":\n"+strings.TrimSpace(string(output)))
	}
	return strings.Join(outputs, "\n")
}

func truncateE32ProfileOutput(value string) string {
	const limit = 400
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
