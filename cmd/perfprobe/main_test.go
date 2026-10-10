package main

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
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

func TestZeroBootstrapBudgetDoesNotReserveOrInvoke(t *testing.T) {
	journalBase := filepath.Join(t.TempDir(), "zero-budget")
	budget := &bootstrapBudget{limit: 0, journalBase: journalBase}
	calls := 0
	if _, err := budget.invoke("must-not-run", func() (*rlwe.Ciphertext, error) {
		calls++
		return nil, nil
	}); err == nil {
		t.Fatal("zero-call budget accepted an invocation")
	}
	if budget.attempts != 0 || calls != 0 {
		t.Fatalf("zero-call budget changed state: attempts=%d calls=%d", budget.attempts, calls)
	}
	if _, err := os.Stat(journalBase + ".bootstrap-attempt-01.json"); !os.IsNotExist(err) {
		t.Fatalf("zero-call budget wrote a reservation: stat error=%v", err)
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
	if err := validateExecutionLimits(cliOptions{mode: "public-native", publicBootstrap: true, preflightPair: "preflight.json", bootstrapBudget: 2}); err != nil {
		t.Fatalf("exact two-call public-native cold/warm budget rejected: %v", err)
	}
	for _, budget := range []int{0, 1, 3} {
		if err := validateExecutionLimits(cliOptions{mode: "public-native", publicBootstrap: true, preflightPair: "preflight.json", bootstrapBudget: budget}); err == nil {
			t.Fatalf("public-native cold/warm mode accepted budget=%d, want exactly 2", budget)
		}
	}
	if err := validateExecutionLimits(cliOptions{mode: "public-native", publicBootstrap: true, bootstrapBudget: 2}); err == nil {
		t.Fatal("public-native cold/warm mode accepted a missing zero-call preflight pair")
	}
	if err := validateExecutionLimits(cliOptions{mode: "public-native", preflightPair: "preflight.json", bootstrapBudget: 0}); err == nil {
		t.Fatal("zero-call mode accepted an unused preflight-pair path")
	}
	if err := validateExecutionLimits(cliOptions{mode: "legacy-diagnostic", publicBootstrap: true, preflightPair: "preflight.json", bootstrapBudget: 2}); err == nil {
		t.Fatal("legacy mode accepted public-native Bootstrap flags")
	}
}

func TestPublicBootstrapAttemptsUseIndependentCopiesOfOneHeldInput(t *testing.T) {
	input := publicTestCiphertext(t)
	baseFingerprint, err := publicCiphertextFingerprint(input)
	if err != nil {
		t.Fatal(err)
	}
	budget := &bootstrapBudget{limit: 2, journalBase: filepath.Join(t.TempDir(), "bootstrap")}
	calls := 0
	var copyFingerprints []string
	call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
		calls++
		fingerprint, err := publicCiphertextFingerprint(callInput)
		if err != nil {
			return nil, err
		}
		copyFingerprints = append(copyFingerprints, fingerprint)
		callInput.Value[0].Coeffs[0][0]++ // A fake backend mutates only the per-call copy.
		return callInput.CopyNew(), nil
	}
	validate := func(name string, _ *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
		values := make([]complex128, 1<<12)
		for i := range values {
			values[i] = complex(float64(i)/4096, -float64(i)/8192)
		}
		return publicCheckpoint{Name: name}, values, nil
	}
	results, outputs, phases, err := runTwoPublicBootstrapAttempts(input, budget, call, validate)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || budget.attempts != 2 || len(results) != 2 || len(outputs) != 2 || len(phases) != 2 {
		t.Fatalf("cold/warm attempts calls=%d reserved=%d results=%d outputs=%d phases=%d", calls, budget.attempts, len(results), len(outputs), len(phases))
	}
	if copyFingerprints[0] != baseFingerprint || copyFingerprints[1] != baseFingerprint || results[0].InputCiphertextSHA256 != results[1].InputCiphertextSHA256 || results[0].InputCiphertextSHA256 != baseFingerprint {
		t.Fatal("cold and warm did not receive independent copies of the same held ciphertext")
	}
	afterFingerprint, err := publicCiphertextFingerprint(input)
	if err != nil {
		t.Fatal(err)
	}
	if afterFingerprint != baseFingerprint {
		t.Fatal("fake Bootstrap mutation escaped its per-call ciphertext copy")
	}
	if _, err := os.Stat(budget.journalBase + ".bootstrap-attempt-01.json"); err != nil {
		t.Fatalf("cold reservation missing: %v", err)
	}
	if _, err := os.Stat(budget.journalBase + ".bootstrap-attempt-02.json"); err != nil {
		t.Fatalf("warm reservation missing: %v", err)
	}
}

func TestPublicCiphertextFingerprintBindsDomainAndDimensions(t *testing.T) {
	ciphertext := publicTestCiphertext(t)
	baseline, err := publicCiphertextFingerprint(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	nttVariant := ciphertext.CopyNew()
	nttVariant.IsNTT = !nttVariant.IsNTT
	nttFingerprint, err := publicCiphertextFingerprint(nttVariant)
	if err != nil {
		t.Fatal(err)
	}
	if nttFingerprint == baseline {
		t.Fatal("ciphertext fingerprint ignored NTT domain metadata")
	}
	dimensionsVariant := ciphertext.CopyNew()
	dimensionsVariant.LogDimensions.Cols++
	dimensionsFingerprint, err := publicCiphertextFingerprint(dimensionsVariant)
	if err != nil {
		t.Fatal(err)
	}
	if dimensionsFingerprint == baseline {
		t.Fatal("ciphertext fingerprint ignored logical dimensions")
	}
}

func TestPublicNativeOutputPreflightRejectsAliasesAndExistingArtifacts(t *testing.T) {
	root := t.TempDir()
	result := filepath.Join(root, "result.json")
	vectors := filepath.Join(root, "vectors.json")
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: result, vectorsOut: vectors, publicBootstrap: true}); err != nil {
		t.Fatalf("fresh output paths rejected: %v", err)
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: result, vectorsOut: filepath.Join(root, ".", "result.json"), publicBootstrap: true}); err == nil {
		t.Fatal("aliased result/vector output paths accepted")
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: filepath.Join(root, "missing", "result.json"), vectorsOut: vectors, publicBootstrap: true}); err == nil {
		t.Fatal("output path with a missing parent directory accepted")
	}
	if err := os.WriteFile(result+".bootstrap-attempt-02.json", []byte("reserved"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: result, vectorsOut: vectors, publicBootstrap: true}); err == nil {
		t.Fatal("existing warm-attempt journal accepted before measurement")
	}
}

func TestPublicBootstrapGateRequiresExactPassingZeroCallPair(t *testing.T) {
	wantNames := []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"}
	base := publicPairDocument{
		SchemaVersion: "fast-standard-public-native-pair.v1", Status: "PASS", NumericalGate: publicNumericalGate,
		FastCommit: publicFastSHA, StandardCommit: publicStandardSHA, PrimaryCommit: "primary-sha", SourceSHA256: "source-sha",
		InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA, ConfigSHA256: publicConfigSHA, QPSHA256: publicQPSHA,
		E: 32, BootstrapCalls: 0, MatchedEnvironment: true,
	}
	for _, name := range wantNames {
		base.Checkpoints = append(base.Checkpoints, publicPairCheckpoint{
			Name: name, Pass: true, StateMatched: true, Metrics: vectorMetrics{MaxComplexDifference: 0},
		})
	}
	tests := []struct {
		name   string
		mutate func(*publicPairDocument)
		valid  bool
	}{
		{name: "canonical passing zero-call pair", valid: true},
		{name: "contains bootstrap comparisons", mutate: func(pair *publicPairDocument) {
			pair.Bootstrap = append(pair.Bootstrap, publicPairBootstrapComparison{Phase: "first_cold_bootstrap"})
		}},
		{name: "negative error metric", mutate: func(pair *publicPairDocument) {
			pair.Checkpoints[0].Metrics.MaxComplexDifference = -1
		}},
		{name: "failed checkpoint", mutate: func(pair *publicPairDocument) {
			pair.Checkpoints[0].Pass = false
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pair := base
			pair.Checkpoints = append([]publicPairCheckpoint(nil), base.Checkpoints...)
			if test.mutate != nil {
				test.mutate(&pair)
			}
			path := filepath.Join(t.TempDir(), "preflight-pair.json")
			if err := writeExclusiveJSON(path, pair); err != nil {
				t.Fatal(err)
			}
			_, err := validatePublicPreflightGate(path, "primary-sha", "source-sha", publicConfigSHA, publicQPSHA, publicInputSHA, publicWorkloadSHA)
			if test.valid && err != nil {
				t.Fatalf("canonical zero-call pair rejected: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("invalid preflight pair accepted")
			}
		})
	}
}

func TestPublicBootstrapAttemptsStopAfterFirstOutputGateFailure(t *testing.T) {
	input := publicTestCiphertext(t)
	journalBase := filepath.Join(t.TempDir(), "bootstrap-fail-closed")
	budget := &bootstrapBudget{limit: 2, journalBase: journalBase}
	calls := 0
	call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
		calls++
		return callInput.CopyNew(), nil
	}
	validate := func(string, *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
		return publicCheckpoint{}, nil, errors.New("simulated first-output gate failure")
	}
	_, _, _, err := runTwoPublicBootstrapAttempts(input, budget, call, validate)
	if err == nil || !strings.Contains(err.Error(), "simulated first-output gate failure") {
		t.Fatalf("failed first output gate not returned: %v", err)
	}
	if calls != 1 || budget.attempts != 1 {
		t.Fatalf("failure incorrectly continued: calls=%d reserved=%d", calls, budget.attempts)
	}
	if _, err := os.Stat(journalBase + ".bootstrap-attempt-01.json"); err != nil {
		t.Fatalf("spent cold reservation missing: %v", err)
	}
	if _, err := os.Stat(journalBase + ".bootstrap-attempt-02.json"); !os.IsNotExist(err) {
		t.Fatalf("warm attempt was reserved after first-call failure: stat error=%v", err)
	}
}

func TestPublicBootstrapAttemptErrorReportsWarmReservationPhase(t *testing.T) {
	input := publicTestCiphertext(t)
	journalBase := filepath.Join(t.TempDir(), "warm-reservation-collision")
	if err := os.WriteFile(journalBase+".bootstrap-attempt-02.json", []byte("already reserved"), 0o600); err != nil {
		t.Fatal(err)
	}
	budget := &bootstrapBudget{limit: 2, journalBase: journalBase}
	calls := 0
	call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
		calls++
		return callInput.CopyNew(), nil
	}
	validate := func(name string, _ *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
		values := make([]complex128, 1<<12)
		return publicCheckpoint{Name: name}, values, nil
	}
	_, _, _, err := runTwoPublicBootstrapAttempts(input, budget, call, validate)
	var attemptErr *publicBootstrapAttemptError
	if !errors.As(err, &attemptErr) || attemptErr.Phase != "later_warm_bootstrap" {
		t.Fatalf("warm reservation failure phase=%v, want later_warm_bootstrap (error=%v)", attemptErr, err)
	}
	if calls != 1 || budget.attempts != 1 {
		t.Fatalf("warm reservation collision invoked extra work: calls=%d reserved=%d", calls, budget.attempts)
	}
}

func TestPublicBootstrapPairRequiresBoundedNativeOutputs(t *testing.T) {
	standard, standardVectors := makePublicBootstrapPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("valid bounded cold/warm output pair rejected: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*publicNativeDocument, *publicVectorDocument)
	}{
		{
			name: "output decoded hash mismatch",
			mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
				doc.BootstrapResults[0].Output.DecodedSHA256 = "tampered"
			},
		},
		{
			name: "output vector wrong count",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				vectors.BootstrapOutputs["bootstrap_first_cold"] = vectors.BootstrapOutputs["bootstrap_first_cold"][:1]
			},
		},
		{
			name: "cold and warm input mismatch",
			mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
				doc.BootstrapResults[1].InputCiphertextSHA256 = "different-input"
			},
		},
		{
			name: "duplicate phase",
			mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
				doc.BootstrapResults[1].Phase = doc.BootstrapResults[0].Phase
			},
		},
		{
			name: "native decode declaration changed",
			mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
				doc.BootstrapResults[0].Output.DecodePath = "Fast c0 projection"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := clonePublicNativeDocument(standard)
			mutated.BootstrapResults = append([]publicBootstrapResult(nil), standard.BootstrapResults...)
			vectors := clonePublicVectorDocument(standardVectors)
			test.mutate(&mutated, &vectors)
			if err := validatePublicBootstrapPairArtifacts(mutated, fast, vectors, fastVectors); err == nil {
				t.Fatal("public Bootstrap comparator accepted invalid cold/warm evidence")
			}
		})
	}
}

func publicTestCiphertext(t *testing.T) *rlwe.Ciphertext {
	t.Helper()
	config, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	residual, _, _, err := perfmeasure.ParametersFromConfigWithE(config, 32)
	if err != nil {
		t.Fatal(err)
	}
	return ckks.NewCiphertext(residual, 1, 0)
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
		Parameters: params, EphemeralSecretWeight: 32, InputSlots: 1 << 12, NumericalGate: publicNumericalGate, BootstrapBudget: 0, BootstrapCalls: 0,
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

func makePublicBootstrapPairArtifactFixture(t *testing.T, backend string) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	doc, vectors := makePublicPairArtifactFixture(t, backend)
	doc.SchemaVersion, doc.Mode, doc.Status = "fast-standard-public-native-bootstrap.v1", "public-native-bootstrap", "PASS_PUBLIC_BOOTSTRAP_COLD_WARM"
	doc.BootstrapBudget, doc.BootstrapCalls, doc.PreflightPairSHA256 = 2, 2, "preflight-pair-sha"
	vectors.SchemaVersion, vectors.Mode = "fast-standard-public-native-bootstrap-vectors.v1", "public-native-bootstrap"
	vectors.BootstrapOutputs = make(map[string][]complexValue, 2)
	preBootstrap := decodeComplexValues(vectors.Checkpoints["drop_level0"])
	params, _ := publicFixtureParameters(t)
	for _, item := range []struct {
		phase string
		name  string
	}{
		{phase: "first_cold_bootstrap", name: "bootstrap_first_cold"},
		{phase: "later_warm_bootstrap", name: "bootstrap_later_warm"},
	} {
		rowLengths, rowHashes := make([][]int, 2), make([][]string, 2)
		for component := 0; component < 2; component++ {
			rowLengths[component] = []int{1 << 13, 1 << 13}
			rowHashes[component] = []string{"q0-row-sha", "q1-row-sha"}
		}
		scale, err := expectedPublicCheckpointScale(params, item.name)
		if err != nil {
			t.Fatal(err)
		}
		checkpoint := publicCheckpoint{
			Name: item.name, State: ciphertextState{Level: 1, Degree: 1, Scale: scale, QPrefixRows: 2},
			DecodePath:         "rlwe.NewDecryptor(residual parameters, generated secret).DecryptNew -> ckks.NewEncoder.Decode",
			PhysicalRowLengths: rowLengths, RowSHA256: rowHashes, FastCompactPrefix: backend == "fast", C1Nonzero: backend == "standard",
			DecodedSHA256: perfmeasure.Fingerprint(preBootstrap), OraclePass: true,
			Oracle: compareVectors(preBootstrap, preBootstrap, publicNumericalGate), OracleSNR: numericalmetrics.Compare(preBootstrap, preBootstrap),
		}
		if backend == "fast" {
			checkpoint.DecodePath = "rlwe.NewDecryptor(residual parameters, matching zero-secret-mode secret).DecryptNew -> ckks.NewEncoder.Decode"
		}
		doc.BootstrapResults = append(doc.BootstrapResults, publicBootstrapResult{
			Phase: item.phase, InputCiphertextSHA256: "held-level0-ciphertext-sha",
			Timing: phaseTiming{Phase: item.phase, ElapsedNS: 100, Samples: 1, Available: true}, Output: checkpoint,
		})
		vectors.BootstrapOutputs[item.name] = encodeVector(preBootstrap)
	}
	return doc, vectors
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
	clone.BootstrapOutputs = make(map[string][]complexValue, len(doc.BootstrapOutputs))
	for name, values := range doc.BootstrapOutputs {
		clone.BootstrapOutputs[name] = append([]complexValue(nil), values...)
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
