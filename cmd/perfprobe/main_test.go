package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestDeterministicLogN13InputKeepsCanonicalFingerprint(t *testing.T) {
	if got := perfmeasure.Fingerprint(perfmeasure.DeterministicInput(1 << 12)); got != canonicalLogN13InputSHA256 {
		t.Fatalf("LogN13 input SHA-256 = %s, want %s", got, canonicalLogN13InputSHA256)
	}
}

func TestBootstrapBudgetFailsClosedAndConsumesFailedAttempts(t *testing.T) {
	journalBase := filepath.Join(t.TempDir(), "bounded-run")
	budget := &bootstrapBudget{limit: 1, journalBase: journalBase}
	calls := 0
	call := func() (*rlwe.Ciphertext, error) {
		calls++
		return nil, errors.New("simulated failed call")
	}
	if _, err := budget.invoke("unit-test", call); err == nil || err.Error() != "simulated failed call" {
		t.Fatalf("first bounded attempt error=%v", err)
	}
	if budget.attempts != 1 || calls != 1 {
		t.Fatalf("failed attempt was not consumed: attempts=%d calls=%d", budget.attempts, calls)
	}
	if _, err := os.Stat(journalBase + ".bootstrap-attempt-01.json"); err != nil {
		t.Fatalf("irrevocable pre-call attempt reservation missing: %v", err)
	}
	if _, err := budget.invoke("over-budget", call); err == nil {
		t.Fatal("budget allowed a second actual invocation")
	}
	if budget.attempts != 1 || calls != 1 {
		t.Fatalf("exhausted budget invoked operation: attempts=%d calls=%d", budget.attempts, calls)
	}
	restarted := &bootstrapBudget{limit: 1, journalBase: journalBase}
	if _, err := restarted.invoke("retry", call); err == nil {
		t.Fatal("new process budget reused an existing output journal and retried the call")
	}
	if calls != 1 {
		t.Fatalf("journal collision still invoked operation: calls=%d", calls)
	}
}

func TestBootstrapBudgetSupportsExplicitTwoAttemptPolicy(t *testing.T) {
	journalBase := filepath.Join(t.TempDir(), "two-attempt-run")
	budget := &bootstrapBudget{limit: 2, journalBase: journalBase}
	calls := 0
	call := func() (*rlwe.Ciphertext, error) {
		calls++
		return nil, nil
	}
	for i, suffix := range []string{"01", "02"} {
		i++
		if _, err := budget.invoke("bounded-two-call-test", call); err != nil {
			t.Fatalf("bounded attempt %d failed: %v", i, err)
		}
		journal := journalBase + ".bootstrap-attempt-" + suffix + ".json"
		if _, err := os.Stat(journal); err != nil {
			t.Fatalf("attempt %d journal missing: %v", i, err)
		}
	}
	if budget.attempts != 2 || calls != 2 {
		t.Fatalf("explicit two-attempt budget consumed attempts=%d calls=%d", budget.attempts, calls)
	}
	if _, err := budget.invoke("over-budget", call); err == nil {
		t.Fatal("two-attempt policy allowed a third invocation")
	}
	if calls != 2 {
		t.Fatalf("over-budget invocation executed underlying operation: calls=%d", calls)
	}
}

func TestPublicNativeModeIsExplicitlyZeroBootstrap(t *testing.T) {
	if err := validateExecutionLimits(cliOptions{mode: "public-native", bootstrapBudget: 0}); err != nil {
		t.Fatalf("zero-call public-native preflight rejected: %v", err)
	}
	if err := validateExecutionLimits(cliOptions{mode: "public-native", bootstrapBudget: 1}); err == nil {
		t.Fatal("public-native preflight accepted a nonzero Bootstrap allowance")
	}
	if err := validateExecutionLimits(cliOptions{inputSmoke: true}); err != nil {
		t.Fatalf("legacy input-only smoke unexpectedly requires Bootstrap budget: %v", err)
	}
	if err := validateExecutionLimits(cliOptions{inputSmoke: true, fastBootstrapAcceptance: true, bootstrapBudget: 1}); err != nil {
		t.Fatalf("explicit one-call legacy acceptance budget rejected: %v", err)
	}
}

func TestPublicPairProvenanceRequiresPinnedCoverageAndCleanSources(t *testing.T) {
	checkpointNames := []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"}
	makeArtifact := func(backend string) (publicNativeDocument, publicVectorDocument) {
		commit, tag := publicStandardSHA, "perf_standard"
		if backend == "fast" {
			commit, tag = publicFastSHA, "perf_fast"
		}
		checkpoints := make([]publicCheckpoint, 0, len(checkpointNames))
		vectors := make(map[string][]complexValue, len(checkpointNames))
		for _, name := range checkpointNames {
			compact := backend == "fast" && name != "encrypt_a_level5" && name != "encrypt_b_level5" && name != "encrypt_c_q5_level5"
			decodePath := "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode"
			if backend == "fast" && name != "drop_level0" {
				decodePath = "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)"
			}
			checkpoints = append(checkpoints, publicCheckpoint{Name: name, DecodePath: decodePath, FastCompactPrefix: compact, OraclePass: true, Oracle: vectorMetrics{MaxComplexDifference: 0}})
			vectors[name] = []complexValue{{Real: 0.25}}
		}
		params := perfmeasure.EffectiveParameters{EphemeralSecretWeight: 32, InputSlots: 1 << 12}
		doc := publicNativeDocument{
			SchemaVersion: "fast-standard-public-native-preflight.v1", Mode: "public-native", Status: "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN",
			Profile: "logn13", Backend: backend, BackendCommit: commit, BackendClean: true,
			PrimaryCommit: "primary-sha", PrimaryClean: true, PrimarySourceSHA256: "source-sha", BuildTag: tag,
			ConfigSHA256: publicConfigSHA, QPSHA256: publicQPSHA, InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA,
			Parameters: params, EphemeralSecretWeight: 32, InputSlots: 1 << 12, BootstrapBudget: 0, BootstrapCalls: 0,
			Checkpoints: checkpoints,
		}
		vectorDoc := publicVectorDocument{
			SchemaVersion: "fast-standard-public-native-vectors.v1", Mode: "public-native", Backend: backend,
			BackendCommit: commit, PrimaryCommit: "primary-sha", SourceSHA256: "source-sha", ConfigSHA256: publicConfigSHA,
			QPSHA256: publicQPSHA, InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA, E: 32, Checkpoints: vectors,
		}
		return doc, vectorDoc
	}
	standard, standardVectors := makeArtifact("standard")
	fast, fastVectors := makeArtifact("fast")
	if err := validatePublicPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("canonical public pair rejected: %v", err)
	}

	missing := fast
	missing.Checkpoints = append([]publicCheckpoint(nil), fast.Checkpoints[:len(fast.Checkpoints)-1]...)
	if err := validatePublicPairArtifacts(standard, missing, standardVectors, fastVectors); err == nil {
		t.Fatal("public comparator accepted a pair missing a canonical checkpoint")
	}
	dirty := standard
	dirty.PrimaryClean = false
	if err := validatePublicPairArtifacts(dirty, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("public comparator accepted dirty Primary provenance")
	}
	wrongPin := fast
	wrongPin.BackendCommit = "unapproved-fast-commit"
	if err := validatePublicPairArtifacts(standard, wrongPin, standardVectors, fastVectors); err == nil {
		t.Fatal("public comparator accepted an unapproved Fast backend pin")
	}
}

func TestEffectiveProfilesComeFromTheirConfigFiles(t *testing.T) {
	for _, test := range []struct {
		name, file            string
		logN, logSlots, slots int
	}{
		{name: "logn13", file: "bootstrap_config.logN13.json", logN: 13, logSlots: 12, slots: 1 << 12},
		{name: "logn16", file: "bootstrap_config.logN16.json", logN: 16, logSlots: 15, slots: 1 << 15},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", test.file))
			if err != nil {
				t.Fatal(err)
			}
			residual, params, effective, err := perfmeasure.ParametersFromConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if effective.LogN != test.logN || effective.LogSlots != test.logSlots || effective.InputSlots != test.slots {
				t.Fatalf("unexpected effective dimensions: %+v", effective)
			}
			if len(effective.QChainBits) != params.BootstrappingParameters.MaxLevel()+1 || len(effective.PBits) != len(params.BootstrappingParameters.P()) {
				t.Fatalf("effective chain metadata does not match constructed parameters: %+v", effective)
			}
			if residual.LogN() != test.logN || cfg.Mod1Degree != 30 || cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 {
				t.Fatalf("config did not preserve required effective inputs: cfg=%+v residual LogN=%d", cfg, residual.LogN())
			}
		})
	}
}
