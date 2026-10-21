package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestVerifyCompiledBackendSourceResolvesRelativeReplacePath(t *testing.T) {
	expected, err := filepath.Abs(filepath.Join("..", "..", "..", "lattigo"))
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyCompiledBackendSource(expected); err != nil {
		t.Fatalf("relative build-info replacement did not resolve to the pinned Secondary source: %v", err)
	}
}

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
	if err := validateExecutionLimits(cliOptions{mode: "public-native", publicRepeatability: true, preflightPair: "preflight.json", bootstrapBudget: 6}); err != nil {
		t.Fatalf("exact six-call public-native repeatability budget rejected: %v", err)
	}
	for _, budget := range []int{0, 1, 3} {
		if err := validateExecutionLimits(cliOptions{mode: "public-native", publicBootstrap: true, preflightPair: "preflight.json", bootstrapBudget: budget}); err == nil {
			t.Fatalf("public-native cold/warm mode accepted budget=%d, want exactly 2", budget)
		}
	}
	if err := validateExecutionLimits(cliOptions{mode: "public-native", publicBootstrap: true, bootstrapBudget: 2}); err == nil {
		t.Fatal("public-native cold/warm mode accepted a missing zero-call preflight pair")
	}
	for _, budget := range []int{0, 2, 5, 7} {
		if err := validateExecutionLimits(cliOptions{mode: "public-native", publicRepeatability: true, preflightPair: "preflight.json", bootstrapBudget: budget}); err == nil {
			t.Fatalf("public-native repeatability mode accepted budget=%d, want exactly 6", budget)
		}
	}
	if err := validateExecutionLimits(cliOptions{mode: "public-native", preflightPair: "preflight.json", bootstrapBudget: 0}); err == nil {
		t.Fatal("zero-call mode accepted an unused preflight-pair path")
	}
	if err := validateExecutionLimits(cliOptions{mode: "legacy-diagnostic", publicBootstrap: true, preflightPair: "preflight.json", bootstrapBudget: 2}); err == nil {
		t.Fatal("legacy mode accepted public-native Bootstrap flags")
	}
}

func TestStandardAZeroDiagnosticRequiresExactlyOneE32OutputSmokeCall(t *testing.T) {
	valid := cliOptions{
		mode: "standard-a-zero-diagnostic", ephemeralSecretWeight: 32,
		outputSmoke: true, bootstrapBudget: 1,
	}
	if err := validateExecutionLimits(valid); err != nil {
		t.Fatalf("valid one-call E32 diagnostic rejected: %v", err)
	}
	for _, mutate := range []func(*cliOptions){
		func(o *cliOptions) { o.bootstrapBudget = 0 },
		func(o *cliOptions) { o.bootstrapBudget = 2 },
		func(o *cliOptions) { o.outputSmoke = false },
		func(o *cliOptions) { o.ephemeralSecretWeight = 0 },
		func(o *cliOptions) { o.fastBootstrapAcceptance = true },
		func(o *cliOptions) { o.vectorsOut = "vectors.json" },
	} {
		candidate := valid
		mutate(&candidate)
		if err := validateExecutionLimits(candidate); err == nil {
			t.Fatalf("invalid A=0 diagnostic options accepted: %+v", candidate)
		}
	}
}

func TestPublicRepeatabilityUsesSixIndependentCopiesAndReservations(t *testing.T) {
	input := publicTestCiphertext(t)
	baseFingerprint, err := publicCiphertextFingerprint(input)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(t.TempDir(), "six-attempt-run")
	budget := &bootstrapBudget{limit: 6, journalBase: base}
	calls := 0
	var inputFingerprints []string
	call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
		calls++
		fingerprint, err := publicCiphertextFingerprint(callInput)
		if err != nil {
			return nil, err
		}
		inputFingerprints = append(inputFingerprints, fingerprint)
		callInput.Value[0].Coeffs[0][0]++
		return callInput.CopyNew(), nil
	}
	validate := func(name string, _ *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
		values := make([]complex128, 1<<12)
		return publicCheckpoint{Name: name}, values, nil
	}
	results, outputs, phases, err := runPublicBootstrapAttempts(input, budget, call, validate)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 6 || budget.attempts != 6 || len(results) != 6 || len(outputs) != 6 || len(phases) != 6 {
		t.Fatalf("repeatability calls=%d reserved=%d results=%d outputs=%d phases=%d", calls, budget.attempts, len(results), len(outputs), len(phases))
	}
	for index := range results {
		if inputFingerprints[index] != baseFingerprint || results[index].InputCiphertextSHA256 != baseFingerprint {
			t.Fatalf("attempt %d did not receive an independent copy of the held input", index+1)
		}
		wantPhase, wantName := fmt.Sprintf("warm_bootstrap_%02d", index), fmt.Sprintf("bootstrap_warm_%02d", index)
		if index == 0 {
			wantPhase, wantName = "first_cold_bootstrap", "bootstrap_first_cold"
		}
		if results[index].Phase != wantPhase || results[index].Output.Name != wantName {
			t.Fatalf("attempt %d phase/output=(%s,%s), want (%s,%s)", index+1, results[index].Phase, results[index].Output.Name, wantPhase, wantName)
		}
		if _, err := os.Stat(fmt.Sprintf("%s.bootstrap-attempt-%02d.json", base, index+1)); err != nil {
			t.Fatalf("attempt %d reservation missing: %v", index+1, err)
		}
	}
	if _, err := budget.invoke("over-budget", func() (*rlwe.Ciphertext, error) { return call(input.CopyNew()) }); err == nil || calls != 6 {
		t.Fatalf("six-call budget allowed extra invocation: err=%v calls=%d", err, calls)
	}
	restarted := &bootstrapBudget{limit: 6, journalBase: base}
	if _, _, _, err := runPublicBootstrapAttempts(input, restarted, call, validate); err == nil || calls != 6 {
		t.Fatalf("existing six-call journal was reused after restart: err=%v calls=%d", err, calls)
	}
}

func TestPublicRepeatabilityStopsAtEveryFailedOrInterruptedAttempt(t *testing.T) {
	for failAt := 1; failAt <= 6; failAt++ {
		t.Run(fmt.Sprintf("call_failure_%d", failAt), func(t *testing.T) {
			input := publicTestCiphertext(t)
			base := filepath.Join(t.TempDir(), "six-failure")
			budget := &bootstrapBudget{limit: 6, journalBase: base}
			calls := 0
			call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
				calls++
				if calls == failAt {
					return nil, errors.New("simulated Bootstrap process/call failure")
				}
				return callInput.CopyNew(), nil
			}
			validate := func(name string, output *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
				return publicCheckpoint{Name: name}, make([]complex128, 1<<12), nil
			}
			_, _, _, err := runPublicBootstrapAttempts(input, budget, call, validate)
			if err == nil || calls != failAt || budget.attempts != failAt {
				t.Fatalf("failure at %d did not stop exactly at the reserved call: calls=%d reserved=%d err=%v", failAt, calls, budget.attempts, err)
			}
			for index := 1; index <= 6; index++ {
				path := fmt.Sprintf("%s.bootstrap-attempt-%02d.json", base, index)
				_, statErr := os.Stat(path)
				if index <= failAt && statErr != nil {
					t.Fatalf("spent attempt %d is missing its reservation: %v", index, statErr)
				}
				if index > failAt && !os.IsNotExist(statErr) {
					t.Fatalf("attempt %d was reserved after failure at %d: %v", index, failAt, statErr)
				}
			}
			resumed := &bootstrapBudget{limit: 6, journalBase: base}
			if _, _, _, resumeErr := runPublicBootstrapAttempts(input, resumed, call, validate); resumeErr == nil || calls != failAt {
				t.Fatalf("restart reused a spent six-call journal: calls=%d err=%v", calls, resumeErr)
			}
		})
	}

	input := publicTestCiphertext(t)
	base := filepath.Join(t.TempDir(), "interrupted-six-run")
	if err := os.WriteFile(base+".bootstrap-attempt-04.json", []byte("interrupted reservation"), 0o600); err != nil {
		t.Fatal(err)
	}
	budget := &bootstrapBudget{limit: 6, journalBase: base}
	calls := 0
	call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
		calls++
		return callInput.CopyNew(), nil
	}
	validate := func(name string, _ *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
		return publicCheckpoint{Name: name}, make([]complex128, 1<<12), nil
	}
	if _, _, _, err := runPublicBootstrapAttempts(input, budget, call, validate); err == nil || calls != 3 || budget.attempts != 3 {
		t.Fatalf("interrupted attempt reservation was reused or exceeded: calls=%d reserved=%d err=%v", calls, budget.attempts, err)
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
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: result, vectorsOut: vectors, publicBootstrap: true, bootstrapBudget: 2}); err != nil {
		t.Fatalf("fresh output paths rejected: %v", err)
	}
	fixtureBase := filepath.Join(root, "fast-e32-trace")
	ciphertextPath, parametersPath, manifestPath := publicTraceFixtureArtifactPaths(fixtureBase)
	if filepath.Base(ciphertextPath) != "fast-e32-trace.ciphertext.bin" ||
		filepath.Base(parametersPath) != "fast-e32-trace.parameters.bin" ||
		filepath.Base(manifestPath) != "fast-e32-trace.manifest.json" {
		t.Fatalf("unexpected E32 fixture artifact paths: %q %q %q", ciphertextPath, parametersPath, manifestPath)
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: filepath.Join(root, "preflight.json"), vectorsOut: filepath.Join(root, "preflight-vectors.json"), traceFixtureOut: fixtureBase}); err != nil {
		t.Fatalf("fresh E32 trace fixture paths rejected: %v", err)
	}
	if err := os.WriteFile(ciphertextPath, []byte("reserved"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: filepath.Join(root, "second-preflight.json"), vectorsOut: filepath.Join(root, "second-vectors.json"), traceFixtureOut: fixtureBase}); err == nil {
		t.Fatal("existing E32 held-ciphertext fixture path was accepted")
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: result, vectorsOut: filepath.Join(root, ".", "result.json"), publicBootstrap: true, bootstrapBudget: 2}); err == nil {
		t.Fatal("aliased result/vector output paths accepted")
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: filepath.Join(root, "missing", "result.json"), vectorsOut: vectors, publicBootstrap: true, bootstrapBudget: 2}); err == nil {
		t.Fatal("output path with a missing parent directory accepted")
	}
	if err := os.WriteFile(result+".bootstrap-attempt-02.json", []byte("reserved"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ensurePublicNativeOutputsAvailable(cliOptions{out: result, vectorsOut: vectors, publicBootstrap: true, bootstrapBudget: 2}); err == nil {
		t.Fatal("existing warm-attempt journal accepted before measurement")
	}
}

func TestTraceFixtureRequestIsVersionedAndLogN16IsCurrentFastOnly(t *testing.T) {
	for _, tc := range []struct {
		name, profile string
		wantSchema    string
		wantProfile   string
		wantBackend   string
	}{
		{name: "logn13", profile: "logn13", wantSchema: "fast-public-e32-trace-fixture.v1", wantProfile: "logn13-e32-public-native", wantBackend: publicFastSHA},
		{name: "logn16", profile: "logn16", wantSchema: "fast-public-e32-trace-fixture.v2", wantProfile: "logn16-e32-public-native", wantBackend: publicCurrentFastSHA},
	} {
		t.Run(tc.name, func(t *testing.T) {
			contract, err := publicTraceFixtureContractFor(tc.profile)
			if err != nil {
				t.Fatal(err)
			}
			if contract.SchemaVersion != tc.wantSchema || contract.Profile != tc.wantProfile || contract.BackendCommit != tc.wantBackend {
				t.Fatalf("fixture contract = %+v, want schema=%q profile=%q backend=%q", contract, tc.wantSchema, tc.wantProfile, tc.wantBackend)
			}
			opts := cliOptions{mode: "public-native", profile: tc.profile, backendCommit: tc.wantBackend, ephemeralSecretWeight: 32, bootstrapBudget: 0}
			if err := validateTraceFixtureRequest(opts); err != nil {
				t.Fatalf("current Fast E32 zero-call fixture request rejected: %v", err)
			}
			for name, mutate := range map[string]func(*cliOptions){
				"wrong_fast_pin": func(o *cliOptions) {
					o.backendCommit = publicCurrentFastSHA
					if tc.wantBackend == publicCurrentFastSHA {
						o.backendCommit = publicFastSHA
					}
				},
				"wrong_e":   func(o *cliOptions) { o.ephemeralSecretWeight = 0 },
				"budget":    func(o *cliOptions) { o.bootstrapBudget = 1 },
				"bootstrap": func(o *cliOptions) { o.publicBootstrap = true },
			} {
				bad := opts
				mutate(&bad)
				if err := validateTraceFixtureRequest(bad); err == nil {
					t.Errorf("%s request unexpectedly accepted", name)
				}
			}
		})
	}
	if _, err := publicTraceFixtureContractFor("unsupported"); err == nil {
		t.Fatal("unsupported profile unexpectedly has a trace-fixture contract")
	}
}

func TestLogN16TraceFixtureEmitsExplicitZeroCallCounts(t *testing.T) {
	manifest := publicTraceFixtureManifest{BootstrapBudget: 0, BootstrapCalls: 0}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"bootstrap_budget", "actual_bootstrap_calls"} {
		if got := string(fields[name]); got != "0" {
			t.Errorf("fixture %s = %q, want explicit numeric zero", name, got)
		}
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

func TestPublicBootstrapGateBindsTheCurrentFastLogN13Pin(t *testing.T) {
	pair := publicPairDocument{
		SchemaVersion: "fast-standard-public-native-pair.v1", Status: "PASS", NumericalGate: publicNumericalGate,
		FastCommit: publicCurrentFastSHA, StandardCommit: publicStandardSHA, PrimaryCommit: "primary-sha", SourceSHA256: "source-sha",
		InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA, ConfigSHA256: publicConfigSHA, QPSHA256: publicQPSHA,
		E: 32, BootstrapCalls: 0, MatchedEnvironment: true,
	}
	for _, name := range []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"} {
		pair.Checkpoints = append(pair.Checkpoints, publicPairCheckpoint{
			Name: name, Pass: true, StateMatched: true, Metrics: vectorMetrics{MaxComplexDifference: 0},
		})
	}
	path := filepath.Join(t.TempDir(), "current-fast-preflight-pair.json")
	if err := writeExclusiveJSON(path, pair); err != nil {
		t.Fatal(err)
	}
	if _, err := validatePublicPreflightGateForProfile(path, "primary-sha", "source-sha", "logn13", publicCurrentFastSHA,
		publicConfigSHA, publicQPSHA, publicInputSHA, publicWorkloadSHA); err != nil {
		t.Fatalf("current pinned Fast preflight pair rejected: %v", err)
	}
}

func TestPublicBootstrapGateRequiresVersionedLogN16ObservedOracleEvidence(t *testing.T) {
	profile, err := publicProfileFor("logn16")
	if err != nil {
		t.Fatal(err)
	}
	base := publicPairDocument{
		SchemaVersion: "fast-standard-public-native-pair.v1", Status: "PASS", NumericalGate: publicNumericalGate,
		FastCommit: publicCurrentFastSHA, StandardCommit: publicStandardSHA, PrimaryCommit: "primary-sha", SourceSHA256: "source-sha",
		InputSHA256: publicLogN16InputSHA, WorkloadSHA256: publicLogN16WorkloadSHA,
		ConfigSHA256: publicLogN16ConfigSHA, QPSHA256: publicLogN16QPSHA,
		CapacityEvidenceVersion: publicCapacityEvidenceVersion, CapacityPolicy: publicCapacityPolicyLogN16Observed,
		ConservativeBoundStatus: perfmeasure.PublicCapacityNotProven, ObservedNativeOracleStatus: publicObservedOraclePassed,
		StandardLevel0Continuity: vectorMetrics{}, FastLevel0Continuity: vectorMetrics{},
		E: 32, BootstrapCalls: 0, MatchedEnvironment: true,
	}
	for _, name := range []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"} {
		base.Checkpoints = append(base.Checkpoints, publicPairCheckpoint{Name: name, Pass: true, StateMatched: true, Metrics: vectorMetrics{MaxComplexDifference: 0}})
	}
	tests := []struct {
		name   string
		mutate func(*publicPairDocument)
		valid  bool
	}{
		{name: "complete observed evidence", valid: true},
		{name: "missing evidence version", mutate: func(pair *publicPairDocument) { pair.CapacityEvidenceVersion = "" }},
		{name: "unproven result hidden", mutate: func(pair *publicPairDocument) { pair.ConservativeBoundStatus = perfmeasure.PublicCapacitySufficient }},
		{name: "oracle status failed", mutate: func(pair *publicPairDocument) { pair.ObservedNativeOracleStatus = "NOT_RUN" }},
		{name: "Level0 continuity exceeded", mutate: func(pair *publicPairDocument) {
			pair.FastLevel0Continuity.MaxComplexDifference = 2 * publicNumericalGate
		}},
		{name: "workload hash mismatch", mutate: func(pair *publicPairDocument) { pair.WorkloadSHA256 = publicWorkloadSHA }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pair := base
			pair.Checkpoints = append([]publicPairCheckpoint(nil), base.Checkpoints...)
			if test.mutate != nil {
				test.mutate(&pair)
			}
			path := filepath.Join(t.TempDir(), "logn16-preflight-pair.json")
			if err := writeExclusiveJSON(path, pair); err != nil {
				t.Fatal(err)
			}
			_, err := validatePublicNativeBootstrapPreflight(path, "primary-sha", "source-sha", "logn16", "standard", publicStandardSHA,
				profile.ConfigSHA256, profile.QPSHA256, profile.InputSHA256, profile.WorkloadSHA256)
			if test.valid && err != nil {
				t.Fatalf("complete LogN16 observed-oracle pair rejected: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("incomplete/forged LogN16 observed-oracle pair accepted")
			}
		})
	}
	path := filepath.Join(t.TempDir(), "valid-logn16-preflight-pair.json")
	if err := writeExclusiveJSON(path, base); err != nil {
		t.Fatal(err)
	}
	for _, backend := range []struct{ name, commit string }{
		{name: "standard", commit: publicStandardSHA},
		{name: "fast", commit: publicCurrentFastSHA},
	} {
		if _, err := validatePublicNativeBootstrapPreflight(path, "primary-sha", "source-sha", "logn16", backend.name, backend.commit,
			profile.ConfigSHA256, profile.QPSHA256, profile.InputSHA256, profile.WorkloadSHA256); err != nil {
			t.Errorf("%s build rejected the paired current Fast preflight pin: %v", backend.name, err)
		}
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
	currentFast, currentFastVectors := fast, fastVectors
	currentFast.BackendCommit = publicCurrentFastSHA
	currentFastVectors.BackendCommit = publicCurrentFastSHA
	if err := validatePublicPairArtifacts(standard, currentFast, standardVectors, currentFastVectors); err != nil {
		t.Fatalf("current Fast LogN13 pair rejected: %v", err)
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

func TestFixedSigmaFastPinIsTaskScopedToLogN13(t *testing.T) {
	if err := validatePublicBackendPin("fast", publicFixedSigmaFastSHA, "logn13"); err != nil {
		t.Fatalf("official fixed-sigma LogN13 Fast pin rejected: %v", err)
	}
	if err := validatePublicBackendPin("fast", publicFixedSigmaFastSHA, "logn16"); err == nil {
		t.Fatal("fixed-sigma LogN13 pin was accepted for LogN16")
	}
	if err := validatePublicBackendPin("fast", publicFixedSigmaNoiseFastSHA, "logn13"); err != nil {
		t.Fatalf("fixed-sigma helper LogN13 Fast pin rejected: %v", err)
	}
	if err := validatePublicBackendPin("fast", publicFixedSigmaNoiseFastSHA, "logn16"); err == nil {
		t.Fatal("fixed-sigma helper pin was accepted for LogN16")
	}
}

func TestFixedSigmaChainFastPinIsTaskScoped(t *testing.T) {
	if err := validatePublicFixedSigmaChainBackendPin("fast", publicFixedSigmaChainFastSHA, "logn13"); err != nil {
		t.Fatalf("fixed-sigma chain LogN13 candidate pin rejected: %v", err)
	}
	if err := validatePublicFixedSigmaChainBackendPin("fast", publicFixedSigmaChainFastSHA, "logn16"); err == nil {
		t.Fatal("fixed-sigma chain candidate pin was accepted for LogN16")
	}
	if err := validatePublicBackendPin("fast", publicFixedSigmaChainFastSHA, "logn13"); err == nil {
		t.Fatal("task-scoped chain candidate pin was accepted by the generic public backend gate")
	}
	if err := validatePublicFixedSigmaChainBackendPin("standard", publicFixedSigmaChainFastSHA, "logn13"); err == nil {
		t.Fatal("Fast chain candidate pin was accepted for Standard")
	}
	candidate := cliOptions{
		mode: "public-native", profile: "logn13", backendCommit: publicFixedSigmaChainFastSHA,
		bootstrapBudget: 0, mulRescaleChain: true, fixedSigmaNoise: true,
		fixedSigmaNoiseSigma: 4.004280012097432, fixedSigmaNoiseSeed: 20261025,
		fixedSigmaSigmaProvided: true, fixedSigmaSeedProvided: true,
	}
	if err := validateFixedSigmaNoiseOptions(candidate); (publicBackendName() == "fast") != (err == nil) {
		t.Fatalf("chain-scoped fixed-sigma option validation mismatch for %s build: %v", publicBackendName(), err)
	}
}

func TestFixedSigmaChainBinaryProvenanceRequiresExactCleanSource(t *testing.T) {
	const commit = "0123456789abcdef0123456789abcdef01234567"
	clean, err := validateFixedSigmaChainBinaryProvenance(commit, commit, "false")
	if err != nil || !clean {
		t.Fatalf("matching clean binary provenance rejected: clean=%v err=%v", clean, err)
	}
	for _, tc := range []struct {
		name       string
		build      string
		vcsChanged string
	}{
		{name: "stale revision", build: "89abcdef0123456789abcdef0123456789abcdef", vcsChanged: "false"},
		{name: "modified source", build: commit, vcsChanged: "true"},
		{name: "missing revision", build: "", vcsChanged: "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := validateFixedSigmaChainBinaryProvenance(commit, tc.build, tc.vcsChanged); err == nil {
				t.Fatal("binary provenance mismatch was accepted")
			}
		})
	}
}

func TestFixedSigmaChainV3RequiresCompleteOrderedPhaseMeasurements(t *testing.T) {
	phases := []fixedSigmaNoisePhaseEvidence{
		{Name: "sampling_hash_moments", ElapsedNS: 1, AllocBytes: 2, Allocs: 3},
		{Name: "coefficient_to_qprefix_ntt", ElapsedNS: 4, AllocBytes: 5, Allocs: 6},
		{Name: "candidate_c0_add", ElapsedNS: 7, AllocBytes: 8, Allocs: 9},
		{Name: "qprefix_centered_capacity_check", ElapsedNS: 10, AllocBytes: 11, Allocs: 12},
		{Name: "commit_c0", ElapsedNS: 13, AllocBytes: 14, Allocs: 15},
	}
	if !validFixedSigmaNoisePhaseMeasurements(phases) {
		t.Fatal("complete ordered CP-D phase measurements rejected")
	}
	if validFixedSigmaNoisePhaseMeasurements(phases[:4]) {
		t.Fatal("incomplete CP-D phase measurements accepted")
	}
	phases[2].Name, phases[3].Name = phases[3].Name, phases[2].Name
	if validFixedSigmaNoisePhaseMeasurements(phases) {
		t.Fatal("out-of-order CP-D phase measurements accepted")
	}
	phases[2].Name, phases[3].Name = phases[3].Name, phases[2].Name
	phases[1].ElapsedNS = -1
	if validFixedSigmaNoisePhaseMeasurements(phases) {
		t.Fatal("negative CP-D phase duration accepted")
	}
	if !validFixedSigmaChainSchema("fast-standard-fixed-sigma-two-step-chain.v3") ||
		!validFixedSigmaChainEvidenceSchema("fixed-sigma-two-step-chain.v3") ||
		fixedSigmaChainVectorSchema("fast-standard-fixed-sigma-two-step-chain.v3") != "fast-standard-fixed-sigma-two-step-chain-vectors.v3" {
		t.Fatal("v3 CP-D chain schemas are not wired together")
	}
}

func TestFixedSigmaNoiseOptInRequiresExactFastZeroBootstrapContract(t *testing.T) {
	valid := cliOptions{
		mode: "public-native", profile: "logn13", backendCommit: publicFixedSigmaNoiseFastSHA,
		bootstrapBudget: 0, fixedSigmaNoise: true, fixedSigmaNoiseSigma: 4.004280012097432,
		fixedSigmaNoiseSeed: 20261022, fixedSigmaSigmaProvided: true, fixedSigmaSeedProvided: true,
	}
	err := validateFixedSigmaNoiseOptions(valid)
	if publicBackendName() == "fast" && err != nil {
		t.Fatalf("valid fixed-sigma Fast opt-in rejected: %v", err)
	}
	if publicBackendName() != "fast" && err == nil {
		t.Fatal("fixed-sigma opt-in was accepted outside the Fast build")
	}

	tests := []struct {
		name string
		edit func(*cliOptions)
	}{
		{name: "missing sigma flag", edit: func(opts *cliOptions) { opts.fixedSigmaSigmaProvided = false }},
		{name: "missing seed flag", edit: func(opts *cliOptions) { opts.fixedSigmaSeedProvided = false }},
		{name: "zero sigma", edit: func(opts *cliOptions) { opts.fixedSigmaNoiseSigma = 0 }},
		{name: "nonfinite sigma", edit: func(opts *cliOptions) { opts.fixedSigmaNoiseSigma = math.Inf(1) }},
		{name: "Bootstrap budget", edit: func(opts *cliOptions) { opts.bootstrapBudget = 1 }},
		{name: "LogN16", edit: func(opts *cliOptions) { opts.profile = "logn16" }},
		{name: "wrong helper commit", edit: func(opts *cliOptions) { opts.backendCommit = publicFixedSigmaFastSHA }},
		{name: "Bootstrap enabled", edit: func(opts *cliOptions) { opts.publicBootstrap = true }},
		{name: "trace export", edit: func(opts *cliOptions) { opts.traceFixtureOut = "trace.json" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := valid
			test.edit(&mutated)
			if validateFixedSigmaNoiseOptions(mutated) == nil {
				t.Fatal("invalid fixed-sigma opt-in was accepted")
			}
		})
	}
	if validateFixedSigmaNoiseOptions(cliOptions{fixedSigmaSigmaProvided: true}) == nil {
		t.Fatal("sigma argument without the explicit opt-in was accepted")
	}
}

func TestMulRescaleChainRequiresZeroBootstrapPublicNativeMode(t *testing.T) {
	valid := cliOptions{mode: "public-native", mulRescaleChain: true, bootstrapBudget: 0}
	if err := validateFixedSigmaNoiseOptions(valid); err != nil {
		t.Fatalf("valid public-native chain mode rejected: %v", err)
	}
	if err := validateExecutionLimits(valid); err != nil {
		t.Fatalf("valid zero-Bootstrap chain budget rejected: %v", err)
	}
	for _, invalid := range []cliOptions{
		{mode: "legacy-diagnostic", mulRescaleChain: true, bootstrapBudget: 0},
		{mode: "public-native", mulRescaleChain: true, publicBootstrap: true, bootstrapBudget: 2},
		{mode: "public-native", mulRescaleChain: true, traceFixtureOut: "trace", bootstrapBudget: 0},
	} {
		if err := validateExecutionLimits(invalid); err == nil {
			t.Fatalf("invalid chain execution limits were accepted: %+v", invalid)
		}
	}
	if err := validateFixedSigmaNoiseOptions(cliOptions{mode: "legacy-diagnostic", mulRescaleChain: true}); err == nil {
		t.Fatal("chain mode accepted a non-public-native application path")
	}
}

func TestFixedSigmaChainDispatchEvidenceIsBackendSpecific(t *testing.T) {
	standard := fixedSigmaChainDispatchEvidence("standard")
	fast := fixedSigmaChainDispatchEvidence("fast")
	if standard == "" || fast == "" || standard == fast {
		t.Fatalf("backend dispatch evidence is not distinct: standard=%q fast=%q", standard, fast)
	}
	if fixedSigmaChainDispatchEvidence("unknown") != "" {
		t.Fatal("unknown backend received dispatch evidence")
	}
}

func TestPublicOperationResidualEvidenceBindsDecodedInputsAndOutput(t *testing.T) {
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	standard.BinarySHA256 = strings.Repeat("1", 64)
	fast.BinarySHA256 = strings.Repeat("2", 64)
	fast.BackendCommit = publicFixedSigmaFastSHA
	fastVectors.BackendCommit = publicFixedSigmaFastSHA
	attachPublicOperationResidualEvidence(t, &standard, &standardVectors)
	attachPublicOperationResidualEvidence(t, &fast, &fastVectors)
	if err := validatePublicPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("task-scoped residual evidence rejected: %v", err)
	}
	corrupted := standardVectors
	corrupted.MulRelinRescaleIncrementalResidual = append([]complexValue(nil), standardVectors.MulRelinRescaleIncrementalResidual...)
	corrupted.MulRelinRescaleIncrementalResidual[0].Real += 1
	if err := validatePublicOperationResidualEvidence(standard, corrupted); err == nil {
		t.Fatal("residual validator accepted a modified incremental vector")
	}
}

func TestPublicOperationResidualValidatesSingleFixedSigmaInjectionEvidence(t *testing.T) {
	doc, vectors := makePublicPairArtifactFixture(t, "fast")
	attachPublicOperationResidualEvidence(t, &doc, &vectors)
	preNoise := decodeComplexValues(vectors.Checkpoints["rescale_q5_level4"])
	x := decodeComplexValues(vectors.Checkpoints["add_level5"])
	y := decodeComplexValues(vectors.Checkpoints["encrypt_c_q5_level5"])
	oracles, err := publicOracleCheckpoints(doc.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	preResidual, _, _, _, err := makePublicOperationResidual(x, y, preNoise, oracles["rescale_q5_level4"])
	if err != nil {
		t.Fatal(err)
	}
	seed, sigma := int64(20261022), 4.004280012097432
	capacityProduct := big.NewInt(1)
	for _, raw := range doc.Parameters.QPrimes[:4] {
		prime, parseErr := strconv.ParseUint(raw, 10, 64)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		capacityProduct.Mul(capacityProduct, new(big.Int).SetUint64(prime))
	}
	evidence := &fixedSigmaNoiseEvidence{
		SigmaCoefficient: sigma, SamplingSeed: seed, CoefficientSampleCount: doc.Parameters.RingN,
		CoefficientMean: 0, CoefficientVariance: sigma * sigma, CoefficientRMS: sigma, CoefficientMaxAbs: 17,
		CoefficientSHA256: strings.Repeat("a", 64), SamplingAlgorithm: "Go math/rand NormFloat64 rounded with math.Round",
		QPrefixRows: 4, QPrefixCheckedCoefficients: doc.Parameters.RingN, QPrefixMaxAbs: "123",
		QPrefixPrefixProduct: capacityProduct.String(), QPrefixStrictFit: true,
		DispatchEvidence: "existing fastcore centered-Q-prefix addition-capacity check",
	}
	zeroDelta := make([]complex128, len(preNoise))
	deltaStats := numericalmetrics.SummarizeComplexVector(zeroDelta)
	residual := doc.MulRelinRescaleResidual
	residual.LogicalNoiseInjections = 1
	residual.SigmaCoefficient = &sigma
	residual.SigmaStatus = "FROZEN_CP_B_CANDIDATE_DIAGNOSTIC_ONLY"
	residual.SamplingSeedStatus = "SEEDED_GO_MATH_RAND_RECORDED"
	residual.SamplingSeed = &seed
	residual.FixedSigmaNoise = evidence
	residual.PreNoiseOutputSHA256 = perfmeasure.Fingerprint(preNoise)
	residual.PreNoiseStatistics = &preResidual.Statistics
	residual.DecodedNoiseDelta = &deltaStats
	residual.InjectionElapsedNS = 100
	vectors.MulRelinRescalePreNoiseOutput = encodeVector(preNoise)
	vectors.MulRelinRescaleNoiseDelta = encodeVector(zeroDelta)
	if err := validatePublicOperationResidualEvidence(doc, vectors); err != nil {
		t.Fatalf("single fixed-sigma evidence rejected: %v", err)
	}
	corrupted := vectors
	corrupted.MulRelinRescaleNoiseDelta = append([]complexValue(nil), vectors.MulRelinRescaleNoiseDelta...)
	corrupted.MulRelinRescaleNoiseDelta[0].Real = 1
	if err := validatePublicOperationResidualEvidence(doc, corrupted); err == nil {
		t.Fatal("residual validator accepted a forged decoded injection delta")
	}
}

func attachPublicOperationResidualEvidence(t *testing.T, doc *publicNativeDocument, vectors *publicVectorDocument) {
	t.Helper()
	oracles, err := publicOracleCheckpoints(doc.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	residual, total, residualVector, totalVector, err := makePublicOperationResidual(
		decodeComplexValues(vectors.Checkpoints["add_level5"]),
		decodeComplexValues(vectors.Checkpoints["encrypt_c_q5_level5"]),
		decodeComplexValues(vectors.Checkpoints["rescale_q5_level4"]),
		oracles["rescale_q5_level4"],
	)
	if err != nil {
		t.Fatal(err)
	}
	residual.ElapsedNS = 10
	doc.MulRelinRescaleResidual = residual
	doc.MulRelinRescaleTotalError = total
	vectors.MulRelinRescaleIncrementalResidual = residualVector
	vectors.MulRelinRescaleTotalError = totalVector
}

func TestMakePublicOperationResidualSubtractsDecodedInputProduct(t *testing.T) {
	x := []complex128{2 + 1i, 3 - 2i}
	y := []complex128{4 - 1i, -1 + 2i}
	z := []complex128{x[0]*y[0] + 0.5i, x[1]*y[1] - 0.25}
	oracle := []complex128{x[0] * y[0], x[1] * y[1]}
	residual, total, residualVector, totalVector, err := makePublicOperationResidual(x, y, z, oracle)
	if err != nil {
		t.Fatal(err)
	}
	if residual.Classification != "DIAGNOSTIC_ONLY" || residual.Statistics.Count != 2 || residual.Statistics.MeanReal != -0.125 ||
		residual.Statistics.MeanImag != 0.25 || math.Abs(residual.Statistics.RMS-math.Sqrt(0.15625)) > 1e-15 {
		t.Fatalf("unexpected decoded incremental residual: %+v", residual)
	}
	if total.Status != "FINITE" || total.RMS != residual.Statistics.RMS || len(residualVector) != 2 || len(totalVector) != 2 {
		t.Fatalf("unexpected total-error summary or vectors: total=%+v residual=%v totalVector=%v", total, residualVector, totalVector)
	}
	if _, _, _, _, err := makePublicOperationResidual(x, y[:1], z, oracle); err == nil {
		t.Fatal("mismatched decoded vectors were accepted")
	}
}

func TestLogN16PublicNativeProfileUsesOnlyObservedOracleCapacityPolicy(t *testing.T) {
	profile, err := publicProfileFor("logn16")
	if err != nil {
		t.Fatal(err)
	}
	if profile.InputSlots != 1<<15 || profile.RingN != 1<<16 || profile.ConfigSHA256 != publicLogN16ConfigSHA || profile.InputSHA256 != publicLogN16InputSHA || profile.WorkloadSHA256 != publicLogN16WorkloadSHA || profile.QPSHA256 != publicLogN16QPSHA {
		t.Fatalf("LogN16 contract differs from its frozen canonical profile: %+v", profile)
	}
	if err := validatePublicBackendPin("fast", publicCurrentFastSHA, "logn16"); err != nil {
		t.Fatalf("current Fast source rejected for LogN16: %v", err)
	}
	if err := validatePublicBackendPin("fast", publicFastSHA, "logn16"); err == nil {
		t.Fatal("historical LogN13 Fast source accepted for LogN16")
	}
	config, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN16.json"))
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
	plan, err := assessPublicCapacityForProfile(profile, a, b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N())
	if err != nil {
		t.Fatalf("exact LogN16 observed-oracle profile rejected its Q0123-passing bound assessment: %v", err)
	}
	if profile.CapacityPolicy != publicCapacityPolicyLogN16Observed || plan.TerminalQ0Status != perfmeasure.PublicCapacityNotProven || plan.TerminalQ0Deficit != "309374390400729" {
		t.Fatalf("LogN16 does not carry the pinned NOT_PROVEN sufficient-bound evidence: profile=%+v plan=%+v", profile, plan)
	}
	if err := validatePublicCapacityForProfile(plan, profile); err != nil {
		t.Fatalf("versioned LogN16 capacity evidence rejected: %v", err)
	}
	if err := validatePublicCapacity(plan); err == nil {
		t.Fatal("strict default capacity validator accepted the LogN16 NOT_PROVEN terminal condition")
	}
	logn13, err := publicProfileFor("logn13")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := assessPublicCapacityForProfile(logn13, a, b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N()); err == nil {
		t.Fatal("strict LogN13 profile accepted the LogN16 NOT_PROVEN capacity case")
	}
	_, err = perfmeasure.BoundPublicMulRescaleWorkload(a, b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N())
	const wantBlocker = "terminal Level0 q0 capacity failure: B=18169085704944493 q0=36028797019488257 deficit=309374390400729"
	if err == nil || err.Error() != wantBlocker {
		t.Fatalf("LogN16 E32 workload capacity result = %v, want exact blocker %q", err, wantBlocker)
	}
	parameterized, err := publicNativeProfileForEffectiveConfig("logn16", effective, "noncanonical-config-sha")
	if err != nil {
		t.Fatal(err)
	}
	if parameterized.Kind != publicProfileKindParameterized || parameterized.CapacityPolicy != publicCapacityPolicyStrict {
		t.Fatalf("noncanonical LogN16 config inherited a frozen observed-oracle exception: %+v", parameterized)
	}
	if _, err := assessPublicCapacityForProfile(parameterized, a, b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N()); err == nil || err.Error() != wantBlocker {
		t.Fatalf("parameterized LogN16 did not stop at its explicit strict capacity limit: %v", err)
	}
}

func TestPublicNativeProfilesResolveFromEffectiveLogN(t *testing.T) {
	for _, logN := range []int{13, 14, 15, 16} {
		name := fmt.Sprintf("logn%d", logN)
		effective := perfmeasure.EffectiveParameters{
			LogN: logN, LogSlots: logN - 1, InputSlots: 1 << (logN - 1), RingN: 1 << logN,
			EphemeralSecretWeight: 32,
		}
		profile, err := publicNativeProfileForEffective(name, effective)
		if err != nil || profile.Kind != publicProfileKindParameterized || profile.LogN != logN || profile.InputSlots != effective.InputSlots {
			t.Fatalf("parameterized LogN%d profile was not derived from effective parameters: profile=%+v err=%v", logN, profile, err)
		}
		mismatchedLabel := "logn14"
		if logN == 14 {
			mismatchedLabel = "logn13"
		}
		if _, err := publicNativeProfileForEffective(mismatchedLabel, effective); err == nil || !strings.Contains(err.Error(), "does not match effective profile") {
			t.Fatalf("profile label inconsistent with effective LogN was not rejected: %v", err)
		}
	}
}

func TestPublicNativeProfileConstructionUsesEffectiveParametersAcrossLogN13To16(t *testing.T) {
	primaryRoot := filepath.Join("..", "..")
	base, _, err := perfmeasure.LoadConfig(filepath.Join(primaryRoot, "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, logN := range []int{13, 14, 15, 16} {
		cfg := base
		cfg.LogN, cfg.LogSlots = logN, logN-1
		_, params, effective, err := perfmeasure.ParametersFromConfigWithE(cfg, 32)
		if err != nil {
			t.Fatalf("construct LogN%d test parameters: %v", logN, err)
		}
		profile, err := publicNativeProfileForEffective(nameForLogN(logN), effective)
		if err != nil {
			t.Fatalf("resolve LogN%d profile: %v", logN, err)
		}
		if profile.Kind != publicProfileKindParameterized || params.BootstrappingParameters.LogN() != logN ||
			profile.LogN != effective.LogN || profile.InputSlots != effective.InputSlots || profile.RingN != effective.RingN {
			t.Fatalf("LogN%d profile does not follow constructed parameters: profile=%+v effective=%+v", logN, profile, effective)
		}
	}
}

func TestPublicNativeFrozenProfilesRequireTheirCanonicalConfigHash(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, profileName := range []string{"logn13", "logn16"} {
		configPath := filepath.Join(root, "configs", "bootstrap_config."+profileName+".json")
		cfg, sum, err := perfmeasure.LoadConfig(configPath)
		if err != nil {
			t.Fatal(err)
		}
		_, _, effective, err := perfmeasure.ParametersFromConfigWithE(cfg, 32)
		if err != nil {
			t.Fatal(err)
		}
		canonicalHash := fmt.Sprintf("%x", sum)
		frozen, err := publicNativeProfileForEffectiveConfig(profileName, effective, canonicalHash)
		if err != nil || frozen.Kind != publicProfileKindFrozen {
			t.Fatalf("canonical %s fixture did not remain frozen: profile=%+v err=%v", profileName, frozen, err)
		}
		parameterized, err := publicNativeProfileForEffectiveConfig(profileName, effective, "noncanonical-config-sha")
		if err != nil || parameterized.Kind != publicProfileKindParameterized || parameterized.CapacityPolicy != publicCapacityPolicyStrict {
			t.Fatalf("noncanonical %s config inherited historical fixture policy: profile=%+v err=%v", profileName, parameterized, err)
		}
	}
}

func TestParameterizedPublicPairValidatesAgainstItsEffectiveMetadata(t *testing.T) {
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	standard.ProfileKind, fast.ProfileKind = publicProfileKindParameterized, publicProfileKindParameterized
	if err := validatePublicPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("parameterized pair with matching effective metadata was rejected: %v", err)
	}
	fast.ConfigSHA256 = "different-config"
	if err := validatePublicPairArtifacts(standard, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("parameterized pair with mismatched config provenance was accepted")
	}
}

func nameForLogN(logN int) string { return fmt.Sprintf("logn%d", logN) }

func TestPublicRotateLeftMatchesCKKSSlotConvention(t *testing.T) {
	input := []complex128{1 + 2i, 3 + 4i, 5 + 6i, 7 + 8i}
	want := []complex128{3 + 4i, 5 + 6i, 7 + 8i, 1 + 2i}
	got := publicRotateLeft(input, 1)
	if len(got) != len(want) {
		t.Fatalf("rotation length=%d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("rotation[%d]=%v, want %v", index, got[index], want[index])
		}
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
			name: "corrupted Level0 output",
			mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
				values := vectors.Checkpoints["drop_level0"]
				values[0].Real += 0.25
				vectors.Checkpoints["drop_level0"] = values
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

func TestPublicPairRejectsForgedObservedOracleEvidence(t *testing.T) {
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	tests := []struct {
		name   string
		mutate func(*publicNativeDocument, *publicVectorDocument)
	}{
		{name: "changed capacity policy", mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
			doc.CapacityPolicy = publicCapacityPolicyLogN16Observed
		}},
		{name: "changed observed status", mutate: func(doc *publicNativeDocument, _ *publicVectorDocument) {
			doc.ObservedNativeOracleStatus = "NOT_RUN"
		}},
		{name: "backend vector commit mismatch", mutate: func(_ *publicNativeDocument, vectors *publicVectorDocument) {
			vectors.BackendCommit = "different-backend-commit"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := clonePublicNativeDocument(fast)
			vectors := clonePublicVectorDocument(fastVectors)
			test.mutate(&mutated, &vectors)
			if err := validatePublicPairArtifacts(standard, mutated, standardVectors, vectors); err == nil {
				t.Fatal("public comparator accepted forged or mismatched observed-oracle evidence")
			}
		})
	}
}

func TestLogN16ObservedOracleRejectsForgedLevel0ContinuityMetric(t *testing.T) {
	doc, vectors := makePublicPairArtifactFixture(t, "fast")
	doc.Level0Continuity.MaxComplexDifference = publicNumericalGate
	if err := validatePublicLevel0ContinuityEvidence(publicCapacityPolicyLogN16Observed, doc, vectors); err == nil {
		t.Fatal("LogN16 observed-oracle validator accepted a forged Level4-to-Level0 continuity metric")
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
	continuity, err := comparePublicLevel0Continuity(rotated, rotated)
	if err != nil {
		t.Fatal(err)
	}
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
		CapacityEvidenceVersion: publicCapacityEvidenceVersion, CapacityPolicy: publicCapacityPolicyStrict,
		ConservativeBoundStatus: capacity.TerminalQ0Status, ObservedNativeOracleStatus: publicStrictOraclePassed,
		Level0Continuity: continuity,
		InputC1Nonzero:   backend == "standard",
		Capacity:         capacity,
		Checkpoints:      checkpoints,
	}
	vectorDoc := publicVectorDocument{
		SchemaVersion: "fast-standard-public-native-vectors.v1", Mode: "public-native", Backend: backend,
		BackendCommit: commit, PrimaryCommit: "primary-sha", SourceSHA256: "source-sha", ConfigSHA256: publicConfigSHA,
		QPSHA256: publicQPSHA, InputSHA256: publicInputSHA, WorkloadSHA256: publicWorkloadSHA, E: 32, Checkpoints: vectors,
	}
	return doc, vectorDoc
}

func TestPublicRepeatabilityPairAcceptsExactlySixMatchedOutputs(t *testing.T) {
	standard, standardVectors := makePublicBootstrapPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	standard, standardVectors = promotePublicBootstrapFixtureToRepeatability(t, standard, standardVectors)
	fast, fastVectors = promotePublicBootstrapFixtureToRepeatability(t, fast, fastVectors)
	if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("valid six-sample repeatability pair rejected: %v", err)
	}
	fast.BootstrapResults = append([]publicBootstrapResult(nil), fast.BootstrapResults...)
	fast.BootstrapResults[5].InputCiphertextSHA256 = "different-held-input"
	if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("six-sample pair accepted a changed held-input fingerprint")
	}
}

func promotePublicBootstrapFixtureToRepeatability(t *testing.T, doc publicNativeDocument, vectors publicVectorDocument) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	doc.SchemaVersion, doc.Mode, doc.Status = "fast-standard-public-native-repeatability.v1", "public-native-repeatability", "PASS_PUBLIC_BOOTSTRAP_1COLD_5WARM"
	doc.BootstrapBudget, doc.BootstrapCalls = 6, 6
	base := doc.BootstrapResults[0]
	baseValues := vectors.BootstrapOutputs["bootstrap_first_cold"]
	doc.BootstrapResults = make([]publicBootstrapResult, 0, 6)
	vectors.SchemaVersion, vectors.Mode = "fast-standard-public-native-repeatability-vectors.v1", "public-native-repeatability"
	vectors.BootstrapOutputs = make(map[string][]complexValue, 6)
	attempts, err := publicBootstrapAttemptSpecs(6)
	if err != nil {
		t.Fatal(err)
	}
	for _, attempt := range attempts {
		result := base
		result.Phase = attempt.phase
		result.Timing.Phase = attempt.phase
		result.Output.Name = attempt.name
		doc.BootstrapResults = append(doc.BootstrapResults, result)
		vectors.BootstrapOutputs[attempt.name] = append([]complexValue(nil), baseValues...)
	}
	return doc, vectors
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
