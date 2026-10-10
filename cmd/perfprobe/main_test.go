package main

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestDeterministicLogN13InputKeepsCanonicalFingerprint(t *testing.T) {
	if got := perfmeasure.Fingerprint(perfmeasure.DeterministicInput(1 << 12)); got != canonicalLogN13InputSHA256 {
		t.Fatalf("LogN13 input SHA-256 = %s, want %s", got, canonicalLogN13InputSHA256)
	}
}

func TestLegacyFastModeAcceptsHistoricalAndApprovedBatch021PinsOnly(t *testing.T) {
	for _, pin := range []string{pinnedFastSHA, publicFastSHA} {
		if err := validatePinnedBackend("fast", pin); err != nil {
			t.Fatalf("approved Fast pin %s rejected: %v", pin, err)
		}
	}
	if err := validatePinnedBackend("fast", "unapproved-fast-sha"); err == nil {
		t.Fatal("legacy Fast mode accepted an unapproved source pin")
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
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
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
	duplicate := fast
	duplicate.Checkpoints = append(append([]publicCheckpoint(nil), fast.Checkpoints...), fast.Checkpoints[0])
	if err := validatePublicPairArtifacts(standard, duplicate, standardVectors, fastVectors); err == nil {
		t.Fatal("public comparator accepted a duplicate checkpoint")
	}
}

func TestPublicPairRejectsDecodedVectorIntegrityFailures(t *testing.T) {
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	tests := []struct {
		name   string
		mutate func(*publicNativeDocument, *publicVectorDocument)
	}{
		{
			name: "checkpoint decoded hash mismatch",
			mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
				doc.Checkpoints[0].DecodedSHA256 = "tampered-decoded-hash"
			},
		},
		{
			name: "substituted separate vector artifact",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				values := vectors.Checkpoints["encrypt_a_level5"]
				values[0].Real += 0.25
				vectors.Checkpoints["encrypt_a_level5"] = values
			},
		},
		{
			name: "one-slot vector",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				vectors.Checkpoints["encrypt_a_level5"] = vectors.Checkpoints["encrypt_a_level5"][:1]
			},
		},
		{
			name: "wrong vector count",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				vectors.Checkpoints["encrypt_a_level5"] = vectors.Checkpoints["encrypt_a_level5"][:(1<<12)-1]
			},
		},
		{
			name: "oversized vector",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				vectors.Checkpoints["encrypt_a_level5"] = append(vectors.Checkpoints["encrypt_a_level5"], complexValue{})
			},
		},
		{
			name: "non-finite vector",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				values := vectors.Checkpoints["encrypt_a_level5"]
				values[0].Imag = math.NaN()
				vectors.Checkpoints["encrypt_a_level5"] = values
			},
		},
		{
			name: "missing vector checkpoint",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				delete(vectors.Checkpoints, "drop_level0")
			},
		},
		{
			name: "unexpected vector checkpoint",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				vectors.Checkpoints["unexpected"] = append([]complexValue(nil), vectors.Checkpoints["encrypt_a_level5"]...)
			},
		},
		{
			name: "stale stored oracle metrics",
			mutate: func(doc *publicNativeDocument, vectors *publicVectorDocument) {
				values := vectors.Checkpoints["encrypt_a_level5"]
				values[0].Real += 1e-3
				vectors.Checkpoints["encrypt_a_level5"] = values
				decoded := decodeComplexValues(values)
				doc.Checkpoints[0].DecodedSHA256 = perfmeasure.Fingerprint(decoded)
			},
		},
		{
			name: "source-supplied oracle pass cannot mask failure",
			mutate: func(doc *publicNativeDocument, vectors *publicVectorDocument) {
				values := vectors.Checkpoints["encrypt_a_level5"]
				values[0].Real += 1e-3
				vectors.Checkpoints["encrypt_a_level5"] = values
				doc.Checkpoints[0].DecodedSHA256 = perfmeasure.Fingerprint(decodeComplexValues(values))
				doc.Checkpoints[0].OraclePass = true
				doc.Checkpoints[0].Oracle = vectorMetrics{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := clonePublicNativeDocument(standard)
			vectors := clonePublicVectorDocument(standardVectors)
			test.mutate(&mutated, &vectors)
			if err := validatePublicPairArtifacts(mutated, fast, vectors, fastVectors); err == nil {
				t.Fatal("public comparator accepted invalid checkpoint vector evidence")
			}
		})
	}
}

func TestPublicPairRejectsForgedStateRowsAndCapacity(t *testing.T) {
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	tests := []struct {
		name   string
		mutate func(*publicNativeDocument)
	}{
		{
			name: "checkpoint level",
			mutate: func(doc *publicNativeDocument) {
				doc.Checkpoints[0].State.Level = 4
			},
		},
		{
			name: "Fast dormant row",
			mutate: func(doc *publicNativeDocument) {
				doc.Checkpoints[3].PhysicalRowLengths[0][4] = 1 << 13
			},
		},
		{
			name: "terminal q0 capacity",
			mutate: func(doc *publicNativeDocument) {
				doc.Capacity.Q0 = "15"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := clonePublicNativeDocument(fast)
			test.mutate(&mutated)
			if err := validatePublicPairArtifacts(standard, mutated, standardVectors, fastVectors); err == nil {
				t.Fatal("public comparator accepted forged Level/row/capacity evidence")
			}
		})
	}
}

func makePublicPairArtifactFixture(t *testing.T, backend string) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	commit, tag := publicStandardSHA, "perf_standard"
	if backend == "fast" {
		commit, tag = publicFastSHA, "perf_fast"
	}
	a, b, c, err := perfmeasure.PublicMulRescaleWorkload(1 << 12)
	if err != nil {
		t.Fatal(err)
	}
	add := publicAdd(a, b)
	product := publicMul(add, c)
	rotated := publicRotateLeft(product, 1)
	params, capacity := publicFixtureParameters(t)
	oracles := map[string][]complex128{
		"encrypt_a_level5":    a,
		"encrypt_b_level5":    b,
		"encrypt_c_q5_level5": c,
		"add_level5":          add,
		"mulrelin_level5":     product,
		"rescale_q5_level4":   product,
		"rotate_level4":       rotated,
		"drop_level0":         rotated,
	}
	checkpoints := make([]publicCheckpoint, 0, len(oracles))
	vectors := make(map[string][]complexValue, len(oracles))
	for _, name := range []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"} {
		oracle := oracles[name]
		compact := backend == "fast" && name != "encrypt_a_level5" && name != "encrypt_b_level5" && name != "encrypt_c_q5_level5"
		level := map[string]int{
			"encrypt_a_level5": 5, "encrypt_b_level5": 5, "encrypt_c_q5_level5": 5,
			"add_level5": 5, "mulrelin_level5": 5, "rescale_q5_level4": 4,
			"rotate_level4": 4, "drop_level0": 0,
		}[name]
		activeRows, physicalRows := min(level+1, 4), level+1
		if backend == "fast" && !compact {
			physicalRows = 17
		}
		rowLengths, rowHashes := make([][]int, 2), make([][]string, 2)
		for component := range rowLengths {
			rowLengths[component] = make([]int, physicalRows)
			rowHashes[component] = make([]string, activeRows)
			for row := range rowLengths[component] {
				if row < activeRows || (backend == "fast" && !compact) {
					rowLengths[component][row] = 1 << 13
				}
			}
			for row := range rowHashes[component] {
				rowHashes[component][row] = "row-sha"
			}
		}
		decodePath := "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode"
		if backend == "fast" && name != "drop_level0" {
			decodePath = "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)"
		}
		scale, err := expectedPublicCheckpointScale(params, name)
		if err != nil {
			t.Fatal(err)
		}
		checkpoints = append(checkpoints, publicCheckpoint{
			Name: name, State: ciphertextState{Level: level, Degree: 1, Scale: scale, QPrefixRows: activeRows},
			DecodePath: decodePath, PhysicalRowLengths: rowLengths, RowSHA256: rowHashes,
			FastCompactPrefix: compact, C1Nonzero: backend == "standard",
			DecodedSHA256: perfmeasure.Fingerprint(oracle), OraclePass: true,
			Oracle: compareVectors(oracle, oracle, publicNumericalGate), OracleSNR: numericalmetrics.Compare(oracle, oracle),
		})
		vectors[name] = encodeVector(oracle)
	}
	doc := publicNativeDocument{
		SchemaVersion: "fast-standard-public-native-preflight.v1", Mode: "public-native", Status: "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN",
		Profile: "logn13", Backend: backend, BackendCommit: commit, BackendClean: true,
		PrimaryCommit: "primary-sha", PrimaryClean: true, PrimarySourceSHA256: "source-sha", BuildTag: tag,
		ConfigSHA256: publicConfigSHA, QPSHA256: publicQPSHA, InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA,
		Parameters: params, EphemeralSecretWeight: 32, InputSlots: 1 << 12, BootstrapBudget: 0, BootstrapCalls: 0,
		InputC1Nonzero: backend == "standard",
		Capacity:       capacity,
		Checkpoints:    checkpoints,
	}
	vectorDoc := publicVectorDocument{
		SchemaVersion: "fast-standard-public-native-vectors.v1", Mode: "public-native", Backend: backend,
		BackendCommit: commit, PrimaryCommit: "primary-sha", SourceSHA256: "source-sha", ConfigSHA256: publicConfigSHA,
		QPSHA256: publicQPSHA, InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA, E: 32, Checkpoints: vectors,
	}
	return doc, vectorDoc
}

func publicFixtureParameters(t *testing.T) (perfmeasure.EffectiveParameters, perfmeasure.PublicCapacityPlan) {
	t.Helper()
	configPath := filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json")
	config, _, err := perfmeasure.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	_, params, effective, err := perfmeasure.ParametersFromConfigWithE(config, 32)
	if err != nil {
		t.Fatal(err)
	}
	a, b, c, err := perfmeasure.PublicMulRescaleWorkload(effective.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err := perfmeasure.BoundPublicMulRescaleWorkload(a, b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N())
	if err != nil {
		t.Fatal(err)
	}
	return effective, capacity
}

func clonePublicNativeDocument(doc publicNativeDocument) publicNativeDocument {
	doc.Checkpoints = append([]publicCheckpoint(nil), doc.Checkpoints...)
	for index := range doc.Checkpoints {
		checkpoint := &doc.Checkpoints[index]
		checkpoint.PhysicalRowLengths = cloneIntRows(checkpoint.PhysicalRowLengths)
		checkpoint.RowSHA256 = cloneStringRows(checkpoint.RowSHA256)
	}
	return doc
}

func cloneIntRows(rows [][]int) [][]int {
	clone := make([][]int, len(rows))
	for index, row := range rows {
		clone[index] = append([]int(nil), row...)
	}
	return clone
}

func cloneStringRows(rows [][]string) [][]string {
	clone := make([][]string, len(rows))
	for index, row := range rows {
		clone[index] = append([]string(nil), row...)
	}
	return clone
}

func clonePublicVectorDocument(doc publicVectorDocument) publicVectorDocument {
	clone := doc
	clone.Checkpoints = make(map[string][]complexValue, len(doc.Checkpoints))
	for name, values := range doc.Checkpoints {
		clone.Checkpoints[name] = append([]complexValue(nil), values...)
	}
	return clone
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
