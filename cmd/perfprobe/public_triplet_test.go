package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestPublicBootstrapAttemptSpecsOneCallHasNoImplicitWarmup(t *testing.T) {
	attempts, err := publicBootstrapAttemptSpecs(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(attempts) != 1 || attempts[0].phase != "single_bootstrap" || attempts[0].name != "bootstrap_once" {
		t.Fatalf("one-call budget expands to %#v", attempts)
	}
	for _, budget := range []int{0, 3, 4, 5, 7} {
		if _, err := publicBootstrapAttemptSpecs(budget); err == nil {
			t.Errorf("unsupported exact budget %d accepted", budget)
		}
	}
}

func TestPublicBootstrapOnceBudgetAndReservationAreExactlyOne(t *testing.T) {
	input := publicTestCiphertext(t)
	base := filepath.Join(t.TempDir(), "one-call-run")
	budget := &bootstrapBudget{limit: 1, journalBase: base}
	calls := 0
	call := func(callInput *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
		calls++
		return callInput.CopyNew(), nil
	}
	validate := func(name string, _ *rlwe.Ciphertext) (publicCheckpoint, []complex128, error) {
		return publicCheckpoint{Name: name}, make([]complex128, 1<<12), nil
	}
	results, outputs, phases, err := runPublicBootstrapAttempts(input, budget, call, validate)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || budget.attempts != 1 || len(results) != 1 || len(outputs) != 1 || len(phases) != 1 ||
		results[0].Phase != "single_bootstrap" || results[0].Output.Name != "bootstrap_once" {
		t.Fatalf("single-call budget expanded unexpectedly: calls=%d attempts=%d results=%d outputs=%d phases=%d", calls, budget.attempts, len(results), len(outputs), len(phases))
	}
	if _, err := os.Stat(base + ".bootstrap-attempt-01.json"); err != nil {
		t.Fatalf("single attempt reservation missing: %v", err)
	}
	if _, err := os.Stat(base + ".bootstrap-attempt-02.json"); !os.IsNotExist(err) {
		t.Fatalf("unexpected second reservation: %v", err)
	}
	if _, _, _, err := runPublicBootstrapAttempts(input, &bootstrapBudget{limit: 1, journalBase: base}, call, validate); err == nil || calls != 1 {
		t.Fatalf("existing attempt ledger was reused: err=%v calls=%d", err, calls)
	}
}

func TestPublicBootstrapOnceCLIRejectsBudgetDriftOrMultipleModes(t *testing.T) {
	valid := cliOptions{mode: "public-native", bootstrapBudget: 1, publicBootstrapPilot: true, preflightPair: "passing-pair.json"}
	if err := validateExecutionLimits(valid); err != nil {
		t.Fatalf("valid one-call CLI contract rejected: %v", err)
	}
	for _, invalid := range []cliOptions{
		{mode: "public-native", bootstrapBudget: 2, publicBootstrapPilot: true, preflightPair: "passing-pair.json"},
		{mode: "public-native", bootstrapBudget: 3, publicBootstrapPilot: true, publicBootstrap: true, preflightPair: "passing-pair.json"},
	} {
		if err := validateExecutionLimits(invalid); err == nil {
			t.Fatalf("invalid one-call CLI contract accepted: %+v", invalid)
		}
	}
}

func TestPublicBootstrapNoiseSnapshotRequiresTwoIndependentCapacityCheckedSides(t *testing.T) {
	snapshot := bootstrapNoiseSnapshotEvidence{
		Enabled: true, Sigma: publicBootstrapNoiseSigma, RootSeed: 27, CtCtMultiplyCalls: 1, InjectedPolynomials: 2,
		Events: []bootstrapNoiseEventEvidence{{
			Operation: 1, Level: 3, Scale: "2^45", MultiplyCompleted: true,
			Left:  bootstrapNoiseTestPolynomial(11),
			Right: bootstrapNoiseTestPolynomial(12),
		}},
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBootstrapNoiseSnapshot(encoded, publicBootstrapNoiseSigma, 27, 1<<13, []uint64{3, 5, 7, 11}); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	snapshot.Events[0].Right.Seed = snapshot.Events[0].Left.Seed
	encoded, err = json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBootstrapNoiseSnapshot(encoded, publicBootstrapNoiseSigma, 27, 1<<13, []uint64{3, 5, 7, 11}); err == nil {
		t.Fatal("same left/right side seed accepted")
	}
}

func TestPublicTripletSummaryUsesGlobalRMSEPrecisionAndTails(t *testing.T) {
	doc := publicNativeDocument{
		Backend: "fast", BackendCommit: publicBootstrapNoisePilotFastSHA, BinarySHA256: string(make([]byte, 64)),
		BootstrapCalls: 1, InputSlots: 4, BootstrapInputSHA256: "same-input",
		BootstrapResults: []publicBootstrapResult{{
			Phase: "single_bootstrap", Output: publicCheckpoint{DecodedSHA256: "decoded"},
			Timing: phaseTiming{Phase: "single_bootstrap", ElapsedNS: 1234, AllocBytes: 4096, Allocs: 7, Samples: 1, Available: true},
		}},
	}
	output := []complex128{1, 2i, 3, 4}
	input := []complex128{0, 0, 0, 0}
	arm, err := summarizePublicTripletArm(doc, output, input)
	if err != nil {
		t.Fatal(err)
	}
	wantRMSE := math.Sqrt(7.5)
	if math.Abs(arm.Error.RMSE-wantRMSE) > 1e-15 || arm.Error.PrecisionBits == nil ||
		math.Abs(*arm.Error.PrecisionBits+math.Log2(wantRMSE)) > 1e-15 || arm.Error.Tails.P99 != 4 || arm.Error.Tails.P999 != 4 {
		t.Fatalf("unexpected shared-metric summary: %+v", arm.Error)
	}
	if arm.BootstrapTimeNS != 1234 || arm.AllocBytes != 4096 || arm.Allocs != 7 {
		t.Fatalf("Bootstrap timing/allocation evidence was not retained: %+v", arm)
	}
	zero, err := summarizePublicTripletArm(doc, input, input)
	if err != nil {
		t.Fatal(err)
	}
	if zero.Error.PrecisionBits != nil || zero.Error.PrecisionBitsStatus != "UNDEFINED_ZERO_RMSE" {
		t.Fatalf("zero-RMSE precision did not have explicit undefined status: %+v", zero.Error)
	}
}

func TestPublicBootstrapOncePairArtifactRequiresOneMatchedOutputAndCommonPlaintext(t *testing.T) {
	standard, standardVectors := makePublicBootstrapPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	standard, standardVectors = promotePublicBootstrapFixtureToOnce(t, standard, standardVectors)
	fast, fastVectors = promotePublicBootstrapFixtureToOnce(t, fast, fastVectors)
	if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err != nil {
		t.Fatalf("valid single-call pair rejected: %v", err)
	}
	fastVectors.BootstrapPlaintextInput[0].Real += 1
	fastVectors.BootstrapInputSHA256 = perfmeasure.Fingerprint(decodeComplexValues(fastVectors.BootstrapPlaintextInput))
	fast.BootstrapInputSHA256 = fastVectors.BootstrapInputSHA256
	if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("pair accepted different canonical Bootstrap plaintexts")
	}
}

func TestRunPublicTripletCompareUsesSameOriginalInputAndValidatesSharedFastBinary(t *testing.T) {
	standard, standardVectors := makePublicBootstrapPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	noise, noiseVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	standard, standardVectors = promotePublicBootstrapFixtureToOnce(t, standard, standardVectors)
	fast, fastVectors = promotePublicBootstrapFixtureToOnce(t, fast, fastVectors)
	noise, noiseVectors = promotePublicBootstrapFixtureToOnce(t, noise, noiseVectors)
	prepareTripletFixtureMetadata(&standard, &standardVectors, "standard", publicStandardSHA, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	prepareTripletFixtureMetadata(&fast, &fastVectors, "fast", publicBootstrapNoisePilotFastSHA, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	prepareTripletFixtureMetadata(&noise, &noiseVectors, "fast", publicBootstrapNoisePilotFastSHA, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	bootstrapInput, err := publicOracleCheckpoints(standard.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	canonical := bootstrapInput["drop_level0"]
	// All three outputs stay finite, but their MaxAbs deliberately exceeds the
	// observational 1e-6 gate. The comparator must preserve the measurements
	// and report the gate miss instead of rejecting the triplet.
	installPublicTripletOutput(t, &standard, &standardVectors, canonical, 2e-6)
	installPublicTripletOutput(t, &fast, &fastVectors, canonical, 4e-6)
	installPublicTripletOutput(t, &noise, &noiseVectors, canonical, 1.5e-6)
	noiseSnapshot := bootstrapNoiseSnapshotEvidence{
		Enabled: true, Sigma: publicBootstrapNoiseSigma, RootSeed: 20261027, CtCtMultiplyCalls: 1, InjectedPolynomials: 2,
		Events: []bootstrapNoiseEventEvidence{{
			Operation: 1, Level: 3, Scale: "2^45", MultiplyCompleted: true,
			Left: bootstrapNoiseTestPolynomial(51), Right: bootstrapNoiseTestPolynomial(52),
		}},
	}
	qPrimes, err := parseEffectiveQPrimes(noise.Parameters)
	if err != nil {
		t.Fatal(err)
	}
	prefixProduct := big.NewInt(1)
	for _, prime := range qPrimes[:noiseSnapshot.Events[0].Left.Capacity.Rows] {
		prefixProduct.Mul(prefixProduct, new(big.Int).SetUint64(prime))
	}
	noiseSnapshot.Events[0].Left.Capacity.PrefixProduct = prefixProduct.String()
	noiseSnapshot.Events[0].Right.Capacity.PrefixProduct = prefixProduct.String()
	noise.BootstrapNoiseEvidence, err = json.Marshal(noiseSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{
		filepath.Join(t.TempDir(), "s.json"), filepath.Join(t.TempDir(), "sv.json"),
		filepath.Join(t.TempDir(), "f.json"), filepath.Join(t.TempDir(), "fv.json"),
		filepath.Join(t.TempDir(), "fn.json"), filepath.Join(t.TempDir(), "fnv.json"), filepath.Join(t.TempDir(), "triplet.json"),
	}
	for _, item := range []struct {
		path  string
		value any
	}{
		{paths[0], standard}, {paths[1], standardVectors}, {paths[2], fast}, {paths[3], fastVectors},
		{paths[4], noise}, {paths[5], noiseVectors},
	} {
		if err := writeExclusiveJSON(item.path, item.value); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{
		"--standard=" + paths[0], "--standard-vectors=" + paths[1], "--fast=" + paths[2], "--fast-vectors=" + paths[3],
		"--fast-noise=" + paths[4], "--fast-noise-vectors=" + paths[5], "--out=" + paths[6],
	}
	if err := runPublicTripletCompare(args); err != nil {
		t.Fatalf("valid synthetic three-arm captures rejected: %v", err)
	}
	var result publicTripletDocument
	if err := readJSON(paths[6], &result); err != nil {
		t.Fatal(err)
	}
	if result.SchemaVersion != "fast-bootstrap-noise-fidelity-triplet.v2" || result.Status != "TRIPLET_COMPLETE_EXPLORATORY" || len(result.Arms) != 3 || len(result.Pairwise) != 3 ||
		result.Arms["S"].SuccessfulSlots != 1<<12 || result.Arms["F"].BootstrapTimeNS != 100 || result.NoiseRootSeed == nil || *result.NoiseRootSeed != 20261027 {
		t.Fatalf("triplet summary is incomplete: %+v", result)
	}
	if result.Arms["F"].Error.RMSE <= result.Arms["FN"].Error.RMSE || result.Arms["FN"].CtCtMultiplyCalls != 1 ||
		result.Arms["FN"].InjectedPolynomials != 2 || result.Arms["FN"].SigmaCoefficient != publicBootstrapNoiseSigma {
		t.Fatalf("triplet summary lost fixed-sigma or error metrics: %+v", result.Arms)
	}
	for _, name := range []string{"S", "F", "FN"} {
		arm := result.Arms[name]
		if arm.NumericalGatePass || arm.NumericalGate != publicNumericalGate || arm.NumericalGateMaxAbs <= publicNumericalGate ||
			arm.Error.Stats.Status != "FINITE" || arm.Error.RMSE <= 0 || arm.Error.PrecisionBits == nil {
			t.Errorf("arm %s did not preserve finite above-gate observations and report the numerical gate miss: %+v", name, arm)
		}
	}
	if result.PrecisionDistanceImprovement == nil || *result.PrecisionDistanceImprovement <= 0 {
		t.Fatalf("synthetic FN should be closer in absolute precision bits: %+v", result)
	}

	noise.BackendCommit, noiseVectors.BackendCommit = publicBootstrapNoiseFastSHA, publicBootstrapNoiseFastSHA
	if err := validatePublicBootstrapTripletArm(noise, noiseVectors, "fast", true); err == nil {
		t.Fatal("one-call triplet accepted the historical CP-B pin instead of its exact pilot source")
	}
	noise.BackendCommit, noiseVectors.BackendCommit = publicBootstrapNoisePilotFastSHA, publicBootstrapNoisePilotFastSHA
	noise.BinarySHA256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	if err := validatePublicBootstrapTripletCompatibility(standard, fast, noise, standardVectors, fastVectors, noiseVectors); err == nil {
		t.Fatal("triplet accepted Fast and Fast+Noise artifacts from different binaries")
	}
}

func TestPublicBootstrapHistoricalModeStillRejectsOracleGateMiss(t *testing.T) {
	standard, standardVectors := makePublicBootstrapPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	oracles, err := publicOracleCheckpoints(standard.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	values := decodeComplexValues(standardVectors.BootstrapOutputs["bootstrap_first_cold"])
	values[0] += 2e-6
	standardVectors.BootstrapOutputs["bootstrap_first_cold"] = encodeVector(values)
	checkpoint := &standard.BootstrapResults[0].Output
	checkpoint.DecodedSHA256 = perfmeasure.Fingerprint(values)
	checkpoint.Oracle = compareVectors(oracles["drop_level0"], values, publicNumericalGate)
	checkpoint.OracleSNR = numericalmetrics.Compare(oracles["drop_level0"], values)
	checkpoint.OraclePass = checkpoint.Oracle.MaxComplexDifference <= publicNumericalGate
	if checkpoint.OraclePass {
		t.Fatal("historical-mode fixture did not exceed the strict numerical gate")
	}
	if err := validatePublicBootstrapPairArtifacts(standard, fast, standardVectors, fastVectors); err == nil {
		t.Fatal("legacy two-call Bootstrap mode accepted an output above its strict numerical gate")
	}
}

func TestNoiseFidelity002SixSampleTripletIsExploratoryAndKeepsLegacyGateStrict(t *testing.T) {
	preflightPairPath, preflightPairSHA := makeNoiseFidelity002PreflightPairFixture(t)
	standard, standardVectors := makePublicBootstrapPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	noise, noiseVectors := makePublicBootstrapPairArtifactFixture(t, "fast")
	standard, standardVectors = promoteToNoiseFidelity002Fixture(t, standard, standardVectors, "standard", publicStandardSHA, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 2e-6, false)
	fast, fastVectors = promoteToNoiseFidelity002Fixture(t, fast, fastVectors, "fast", publicNoiseFidelity002FastSHA, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 4e-6, false)
	noise, noiseVectors = promoteToNoiseFidelity002Fixture(t, noise, noiseVectors, "fast", publicNoiseFidelity002FastSHA, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 1.5e-6, true)
	standard.PreflightPairSHA256, fast.PreflightPairSHA256, noise.PreflightPairSHA256 = preflightPairSHA, preflightPairSHA, preflightPairSHA
	noise.BootstrapNoiseEvidence = makeNoiseFidelity002Snapshot(t, noise.Parameters)
	paths := make([]string, 7)
	for index := range paths {
		paths[index] = filepath.Join(t.TempDir(), fmt.Sprintf("artifact-%d.json", index))
	}
	for _, item := range []struct {
		path  string
		value any
	}{
		{paths[0], standard}, {paths[1], standardVectors}, {paths[2], fast}, {paths[3], fastVectors},
		{paths[4], noise}, {paths[5], noiseVectors},
	} {
		if err := writeExclusiveJSON(item.path, item.value); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{
		"--standard=" + paths[0], "--standard-vectors=" + paths[1], "--fast=" + paths[2], "--fast-vectors=" + paths[3],
		"--fast-noise=" + paths[4], "--fast-noise-vectors=" + paths[5], "--preflight-pair=" + preflightPairPath,
		"--out=" + paths[6], "--noise-fidelity-002",
	}
	if err := runPublicTripletCompare(args); err != nil {
		t.Fatalf("valid synthetic 1-cold/5-warm triplet rejected: %v", err)
	}
	var result publicTripletFidelity002Document
	if err := readJSON(paths[6], &result); err != nil {
		t.Fatal(err)
	}
	if result.SchemaVersion != "fast-bootstrap-noise-fidelity-002-triplet.v1" || result.TaskID != publicNoiseFidelity002TaskID ||
		len(result.Arms) != 3 || len(result.Arms["S"].Samples) != 6 || len(result.Arms["F"].Warm) != 5 ||
		result.Arms["S"].WarmElapsedNS.Count != 5 {
		t.Fatalf("six-sample report is incomplete: %+v", result)
	}
	if result.Arms["S"].Samples[0].NumericalGatePass || result.Arms["S"].Samples[0].Error.PrecisionBits == nil ||
		result.Arms["F"].Samples[0].Error.RMSE <= result.Arms["FN"].Samples[0].Error.RMSE ||
		result.NoiseEvidence.RootSeed != publicNoiseFidelity002RootSeed || result.NoiseEvidence.TotalInjectionSubphaseNS == 0 ||
		result.NoiseColdTimeOverheadRatio <= 0 || result.NoiseWarmMedianAllocOverheadBytes != 0 ||
		result.NoiseEvidence.InjectionSubphaseElapsed["commit_c0"] != 0 {
		t.Fatalf("exploratory metrics/Noise accounting missing: %+v", result)
	}
	copyNoiseDocument := func(mutate func(*bootstrapNoiseSnapshotEvidence)) publicNativeDocument {
		t.Helper()
		var snapshot bootstrapNoiseSnapshotEvidence
		if err := json.Unmarshal(noise.BootstrapNoiseEvidence, &snapshot); err != nil {
			t.Fatal(err)
		}
		mutate(&snapshot)
		encoded, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		copy := noise
		copy.BootstrapNoiseEvidence = encoded
		return copy
	}
	zeroNonCommitPhase := copyNoiseDocument(func(snapshot *bootstrapNoiseSnapshotEvidence) {
		snapshot.Events[0].Left.Phases[0].ElapsedNS = 0
	})
	if _, err := summarizePublicTriplet002Noise(zeroNonCommitPhase); err == nil {
		t.Fatal("Noise evidence accepted a zero-duration non-commit subphase")
	}
	perOperationMemStats := copyNoiseDocument(func(snapshot *bootstrapNoiseSnapshotEvidence) {
		snapshot.Events[0].Left.Phases[4].MemoryStatsCollected = true
	})
	if _, err := summarizePublicTriplet002Noise(perOperationMemStats); err == nil {
		t.Fatal("Noise evidence accepted per-operation MemStats collection")
	}
	badFastPair := fast
	badFastPair.PreflightPairSHA256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	if err := validateNoiseFidelity002TripletPreflight(preflightPairPath, standard, badFastPair, noise); err == nil {
		t.Fatal("six-sample comparator accepted captures that name a different preflight pair")
	}
	if len(result.Arms["FN"].EvaluatorConstructionPhases) != 2 || result.Arms["FN"].ConstructionPlusColdElapsedNS <= result.Arms["FN"].Cold.ElapsedNS ||
		result.Arms["FN"].NoiseHookConfiguration == nil || len(result.PairwiseOutputsByPhase) != 6 || result.PrecisionComparisonByPhase[0].NoiseReducedAbsoluteGap == nil ||
		result.NoiseImprovedPrecisionSamples != 6 {
		t.Fatalf("construction costs were not retained: %+v", result.Arms["FN"])
	}
	badFast := fast
	badFast.ExecutionPhases = append(append([]phaseTiming(nil), fast.ExecutionPhases...), phaseTiming{Phase: "bootstrap_noise_hook_configuration", Available: true, Samples: 1, ElapsedNS: 1})
	if err := validatePublicTripletCompatibility002(standard, badFast, noise, standardVectors, fastVectors, noiseVectors); err == nil {
		t.Fatal("six-sample comparator accepted a Noise hook configuration phase on the OFF arm")
	}
	badNoise := noise
	badNoise.ExecutionPhases = append(append([]phaseTiming(nil), noise.ExecutionPhases...), noise.ExecutionPhases[len(noise.ExecutionPhases)-1])
	if err := validatePublicTripletCompatibility002(standard, fast, badNoise, standardVectors, fastVectors, noiseVectors); err == nil {
		t.Fatal("six-sample comparator accepted duplicate Noise hook configuration phases")
	}
	badConstruction := standard
	badConstruction.ExecutionPhases = append(append([]phaseTiming(nil), standard.ExecutionPhases...), standard.ExecutionPhases[0])
	if _, err := summarizePublicTriplet002Arm(badConstruction, standardVectors, decodeComplexValues(standardVectors.BootstrapPlaintextInput)); err == nil {
		t.Fatal("six-sample comparator accepted duplicate evaluator construction phases")
	}

	// The same finite above-gate data is not accepted under the legacy public
	// repeatability mode, preserving its strict numerical acceptance contract.
	legacyStandard, legacyStandardVectors := standard, standardVectors
	legacyFast, legacyFastVectors := fast, fastVectors
	legacyStandard.TaskID, legacyFast.TaskID = "", ""
	legacyStandardVectors.TaskID, legacyFastVectors.TaskID = "", ""
	legacyStandard.Mode, legacyFast.Mode = "public-native-repeatability", "public-native-repeatability"
	legacyStandard.SchemaVersion, legacyFast.SchemaVersion = "fast-standard-public-native-repeatability.v1", "fast-standard-public-native-repeatability.v1"
	legacyStandard.Status, legacyFast.Status = "PASS_PUBLIC_BOOTSTRAP_1COLD_5WARM", "PASS_PUBLIC_BOOTSTRAP_1COLD_5WARM"
	legacyStandardVectors.Mode, legacyFastVectors.Mode = legacyStandard.Mode, legacyFast.Mode
	legacyStandardVectors.SchemaVersion, legacyFastVectors.SchemaVersion = "fast-standard-public-native-repeatability-vectors.v1", "fast-standard-public-native-repeatability-vectors.v1"
	if err := validatePublicBootstrapPairArtifacts(legacyStandard, legacyFast, legacyStandardVectors, legacyFastVectors); err == nil {
		t.Fatal("legacy repeatability validator accepted above-gate output under the new task's observational exception")
	}
}

func makeNoiseFidelity002PreflightPairFixture(t *testing.T) (path, digest string) {
	t.Helper()
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	standard.BinarySHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	fast.BinarySHA256 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	fast.BackendCommit, fastVectors.BackendCommit = publicNoiseFidelity002FastSHA, publicNoiseFidelity002FastSHA
	standard.PrimaryCommit, fast.PrimaryCommit = publicStandardSHA, publicStandardSHA
	standardVectors.PrimaryCommit, fastVectors.PrimaryCommit = publicStandardSHA, publicStandardSHA
	standard.PrimaryEvidenceCommit, fast.PrimaryEvidenceCommit = publicFixedSigmaFastSHA, publicFixedSigmaFastSHA
	standardVectors.PrimaryEvidenceCommit, fastVectors.PrimaryEvidenceCommit = publicFixedSigmaFastSHA, publicFixedSigmaFastSHA
	standard.TaskID, fast.TaskID = publicNoiseFidelity002TaskID, publicNoiseFidelity002TaskID
	standardVectors.TaskID, fastVectors.TaskID = publicNoiseFidelity002TaskID, publicNoiseFidelity002TaskID
	paths := make([]string, 5)
	for index := range paths {
		paths[index] = filepath.Join(t.TempDir(), fmt.Sprintf("002-preflight-%d.json", index))
	}
	for _, item := range []struct {
		path  string
		value any
	}{
		{paths[0], standard}, {paths[1], fast}, {paths[2], standardVectors}, {paths[3], fastVectors},
	} {
		if err := writeExclusiveJSON(item.path, item.value); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{
		"--standard=" + paths[0], "--fast=" + paths[1], "--standard-vectors=" + paths[2],
		"--fast-vectors=" + paths[3], "--out=" + paths[4], "--noise-fidelity-002-preflight",
	}
	if err := runPublicCompare(args); err != nil {
		t.Fatalf("construct valid binary-bound task preflight fixture: %v", err)
	}
	data, err := os.ReadFile(paths[4])
	if err != nil {
		t.Fatal(err)
	}
	return paths[4], fmt.Sprintf("%x", sha256.Sum256(data))
}

func TestNoiseFidelity002ZeroCallPreflightPinsAndBindsBothBinaries(t *testing.T) {
	if err := validatePublicBackendPinForTaskMode("fast", publicNoiseFidelity002FastSHA, "logn13", false, true); err != nil {
		t.Fatalf("exact accepted 002 Fast source pin rejected: %v", err)
	}
	if err := validatePublicBackendPinForTaskMode("fast", publicNoiseFidelity002FastSHA, "logn16", false, true); err == nil {
		t.Fatal("002 Fast source pin accepted outside the frozen LogN13 profile")
	}
	if err := validatePublicBackendPinForTaskMode("fast", publicNoiseFidelity002FastSHA, "logn13", true, true); err == nil {
		t.Fatal("002 Fast source pin leaked into the historical one-call pilot mode")
	}
	if err := validatePublicBackendPinForTaskMode("fast", publicBootstrapNoisePilotFastSHA, "logn13", false, true); err == nil {
		t.Fatal("002 repeatability mode accepted the historical pilot source pin")
	}
	standard, standardVectors := makePublicPairArtifactFixture(t, "standard")
	fast, fastVectors := makePublicPairArtifactFixture(t, "fast")
	standard.BinarySHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	fast.BinarySHA256 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	fast.BackendCommit, fastVectors.BackendCommit = publicNoiseFidelity002FastSHA, publicNoiseFidelity002FastSHA
	standard.PrimaryCommit, fast.PrimaryCommit = publicStandardSHA, publicStandardSHA
	standardVectors.PrimaryCommit, fastVectors.PrimaryCommit = publicStandardSHA, publicStandardSHA
	standard.PrimaryEvidenceCommit, fast.PrimaryEvidenceCommit = publicFixedSigmaFastSHA, publicFixedSigmaFastSHA
	standardVectors.PrimaryEvidenceCommit, fastVectors.PrimaryEvidenceCommit = publicFixedSigmaFastSHA, publicFixedSigmaFastSHA
	for _, doc := range []*publicNativeDocument{&standard, &fast} {
		doc.TaskID = publicNoiseFidelity002TaskID
	}
	standardVectors.TaskID, fastVectors.TaskID = publicNoiseFidelity002TaskID, publicNoiseFidelity002TaskID
	paths := make([]string, 5)
	for index := range paths {
		paths[index] = filepath.Join(t.TempDir(), fmt.Sprintf("preflight-%d.json", index))
	}
	for _, item := range []struct {
		path  string
		value any
	}{
		{paths[0], standard}, {paths[1], fast}, {paths[2], standardVectors}, {paths[3], fastVectors},
	} {
		if err := writeExclusiveJSON(item.path, item.value); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{
		"--standard=" + paths[0], "--fast=" + paths[1], "--standard-vectors=" + paths[2],
		"--fast-vectors=" + paths[3], "--out=" + paths[4], "--noise-fidelity-002-preflight",
	}
	if err := runPublicCompare(args); err != nil {
		t.Fatalf("valid task-scoped zero-call pair rejected: %v", err)
	}
	profile, err := publicProfileFor("logn13")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := validatePublicPreflightGateForTaskMode(paths[4], standard.PrimaryCommit, standard.PrimarySourceSHA256,
		profile, publicNoiseFidelity002FastSHA, standard.ConfigSHA256, standard.QPSHA256, standard.InputSHA256,
		standard.WorkloadSHA256, false, true, "fast", fast.BinarySHA256); err != nil {
		t.Fatalf("valid task/binary-bound pair rejected by Fast run gate: %v", err)
	}
	if _, err := validatePublicPreflightGateForTaskMode(paths[4], standard.PrimaryCommit, standard.PrimarySourceSHA256,
		profile, publicNoiseFidelity002FastSHA, standard.ConfigSHA256, standard.QPSHA256, standard.InputSHA256,
		standard.WorkloadSHA256, false, true, "fast", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"); err == nil {
		t.Fatal("task preflight gate accepted a Fast binary different from its paired binary")
	}
	if err := validatePublicBackendPinForBootstrapOnce("fast", publicNoiseFidelity002FastSHA, "logn13", true); err == nil {
		t.Fatal("new task-specific Fast pin leaked into the historical one-call pilot mode")
	}
	fast.BackendCommit, fastVectors.BackendCommit = publicFastSHA, publicFastSHA
	if err := validatePublicPairArtifactsForTaskMode(standard, fast, standardVectors, fastVectors, false, true); err == nil {
		t.Fatal("task preflight accepted a Fast source commit outside its exact 002 pin")
	}
}

func TestNoiseFidelity002CLIContractRejectsModeAndBudgetDrift(t *testing.T) {
	preflight := cliOptions{
		mode: "public-native", profile: "logn13", ephemeralSecretWeight: 32,
		bootstrapBudget: 0, noiseFidelity002: true,
	}
	if err := validateNoiseFidelity002Options(preflight); err != nil {
		t.Fatalf("valid zero-call task preflight rejected: %v", err)
	}

	repeatability := preflight
	repeatability.publicRepeatability = true
	repeatability.bootstrapBudget = 6
	repeatability.preflightPair = "passing-pair.json"
	if err := validateNoiseFidelity002Options(repeatability); err != nil {
		t.Fatalf("valid six-call task mode rejected: %v", err)
	}

	noise := repeatability
	noise.bootstrapNoiseEnabled = true
	noise.bootstrapNoiseSeedGiven = true
	noise.bootstrapNoiseRootSeed = publicNoiseFidelity002RootSeed
	if publicBackendName() == "fast" {
		if err := validateNoiseFidelity002Options(noise); err != nil {
			t.Fatalf("valid Fast+Noise task mode rejected: %v", err)
		}
	} else if err := validateNoiseFidelity002Options(noise); err == nil {
		t.Fatal("Fast+Noise mode was accepted by a non-Fast build")
	}

	badNoiseSeed := noise
	badNoiseSeed.bootstrapNoiseRootSeed++
	badBudget := repeatability
	badBudget.bootstrapBudget = 5
	badProfile := repeatability
	badProfile.profile = "logn16"
	legacyHook := cliOptions{
		mode: "public-native", profile: "logn13", ephemeralSecretWeight: 32, bootstrapBudget: 1,
		publicBootstrapPilot: true, bootstrapNoiseEnabled: true, bootstrapNoiseSeedGiven: true,
		bootstrapNoiseRootSeed: publicNoiseFidelity002RootSeed,
	}
	if publicBackendName() == "fast" {
		if err := validateNoiseFidelity002Options(legacyHook); err != nil {
			t.Fatalf("historical one-call Noise mode lost its prior acceptance contract: %v", err)
		}
	} else if err := validateNoiseFidelity002Options(legacyHook); err == nil {
		t.Fatal("historical one-call Fast Noise mode was accepted by a non-Fast build")
	}
	legacyHookMissingPilot := legacyHook
	legacyHookMissingPilot.publicBootstrapPilot = false
	conflictingPreflight := preflight
	conflictingPreflight.bootstrapOncePreflight = true
	for name, candidate := range map[string]cliOptions{
		"wrong-root-seed":               badNoiseSeed,
		"wrong-call-budget":             badBudget,
		"wrong-profile":                 badProfile,
		"historical-hook-without-pilot": legacyHookMissingPilot,
		"mixed-one-call-preflight":      conflictingPreflight,
	} {
		if err := validateNoiseFidelity002Options(candidate); err == nil {
			t.Errorf("invalid %s options were accepted", name)
		}
	}
}

func TestNoiseFidelity003LogN16PinsPreflightAndCLIContract(t *testing.T) {
	if err := validatePublicBackendPinForTaskID("standard", publicStandardSHA, "logn16", false, publicNoiseFidelity003TaskID); err != nil {
		t.Fatalf("genuine Standard LogN16 pin rejected: %v", err)
	}
	if err := validatePublicBackendPinForTaskID("fast", publicNoiseFidelity003FastSHA, "logn16", false, publicNoiseFidelity003TaskID); err != nil {
		t.Fatalf("exact 003 Fast source pin rejected: %v", err)
	}
	for _, test := range []struct{ commit, profile string }{
		{publicNoiseFidelity002FastSHA, "logn16"},
		{"9e540b30898cea74f0fa4acdf03477264344db79", "logn16"}, // historical captures are not the current source pin
		{publicNoiseFidelity003FastSHA, "logn13"},
		{publicNoiseFidelity003FastSHA, "logn14"},
	} {
		if err := validatePublicBackendPinForTaskID("fast", test.commit, test.profile, false, publicNoiseFidelity003TaskID); err == nil {
			t.Errorf("003 accepted an unapproved source/profile pair %s/%s", test.commit, test.profile)
		}
	}
	if err := validatePublicBackendPinForTaskID("fast", publicNoiseFidelity003FastSHA, "logn16", true, publicNoiseFidelity003TaskID); err == nil {
		t.Fatal("003 repeatability pin leaked into the one-call pilot mode")
	}

	preflight := cliOptions{mode: "public-native", profile: "logn16", ephemeralSecretWeight: 32, bootstrapBudget: 0, noiseFidelity003: true}
	if err := validateNoiseFidelityOptions(preflight); err != nil {
		t.Fatalf("valid 003 zero-call preflight rejected: %v", err)
	}
	repeatability := preflight
	repeatability.publicRepeatability, repeatability.bootstrapBudget, repeatability.preflightPair = true, 6, "pair.json"
	if err := validateNoiseFidelityOptions(repeatability); err != nil {
		t.Fatalf("valid 003 six-call measurement contract rejected: %v", err)
	}
	noise := repeatability
	noise.bootstrapNoiseEnabled, noise.bootstrapNoiseSeedGiven = true, true
	noise.bootstrapNoiseRootSeed = publicNoiseFidelity003RootSeed
	if publicBackendName() == "fast" {
		if err := validateNoiseFidelityOptions(noise); err != nil {
			t.Fatalf("valid 003 Fast+Noise contract rejected: %v", err)
		}
	} else if err := validateNoiseFidelityOptions(noise); err == nil {
		t.Fatal("Fast+Noise contract was accepted by a non-Fast build")
	}
	wrong := noise
	wrong.bootstrapNoiseRootSeed++
	if err := validateNoiseFidelityOptions(wrong); err == nil {
		t.Fatal("003 accepted a changed root seed")
	}
	wrong = repeatability
	wrong.profile = "logn13"
	if err := validateNoiseFidelityOptions(wrong); err == nil {
		t.Fatal("003 accepted a non-LogN16 profile")
	}
	wrong = repeatability
	wrong.noiseFidelity002 = true
	if err := validateNoiseFidelityOptions(wrong); err == nil {
		t.Fatal("003 accepted a mixed 002 task contract")
	}
}

func TestNoiseFidelity003LogN16PreflightRequiresObservedOracleAndExactBinaries(t *testing.T) {
	profile, err := publicProfileFor("logn16")
	if err != nil {
		t.Fatal(err)
	}
	const (
		standardBinary = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		fastBinary     = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		sourceSHA      = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	)
	pair := publicPairDocument{
		SchemaVersion: "fast-standard-public-native-pair.v1", TaskID: publicNoiseFidelity003TaskID,
		Status: "PASS", Profile: "logn16", ProfileKind: publicProfileKindFrozen, NumericalGate: publicNumericalGate,
		FastCommit: publicNoiseFidelity003FastSHA, StandardCommit: publicStandardSHA,
		FastBinarySHA256: fastBinary, StandardBinarySHA256: standardBinary,
		PrimaryCommit: publicStandardSHA, PrimaryEvidenceCommit: publicBootstrapNoisePilotFastSHA, SourceSHA256: sourceSHA,
		InputSHA256: profile.InputSHA256, WorkloadSHA256: profile.WorkloadSHA256, ConfigSHA256: profile.ConfigSHA256, QPSHA256: profile.QPSHA256,
		CapacityEvidenceVersion: publicCapacityEvidenceVersion, CapacityPolicy: publicCapacityPolicyLogN16Observed,
		ConservativeBoundStatus: perfmeasure.PublicCapacityNotProven, ObservedNativeOracleStatus: publicObservedOraclePassed,
		E: 32, BootstrapCalls: 0, MatchedEnvironment: true,
		StandardLevel0Continuity: vectorMetrics{}, FastLevel0Continuity: vectorMetrics{},
	}
	for _, name := range []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"} {
		pair.Checkpoints = append(pair.Checkpoints, publicPairCheckpoint{Name: name, Pass: true, StateMatched: true, Metrics: vectorMetrics{}})
	}
	path := filepath.Join(t.TempDir(), "003-preflight-pair.json")
	if err := writeExclusiveJSON(path, pair); err != nil {
		t.Fatal(err)
	}
	if _, err := validatePublicPreflightGateForTaskID(path, publicStandardSHA, sourceSHA, profile, publicNoiseFidelity003FastSHA,
		profile.ConfigSHA256, profile.QPSHA256, profile.InputSHA256, profile.WorkloadSHA256, false, publicNoiseFidelity003TaskID, "fast", fastBinary); err != nil {
		t.Fatalf("valid exact-profile 003 preflight rejected: %v", err)
	}
	if _, err := validatePublicPreflightGateForTaskID(path, publicStandardSHA, sourceSHA, profile, publicNoiseFidelity003FastSHA,
		profile.ConfigSHA256, profile.QPSHA256, profile.InputSHA256, profile.WorkloadSHA256, false, publicNoiseFidelity003TaskID, "fast", standardBinary); err == nil {
		t.Fatal("003 preflight accepted a binary other than the paired Fast binary")
	}
	pair.ConservativeBoundStatus = perfmeasure.PublicCapacitySufficient
	badPath := filepath.Join(t.TempDir(), "bad-pair.json")
	if err := writeExclusiveJSON(badPath, pair); err != nil {
		t.Fatal(err)
	}
	if _, err := validatePublicPreflightGateForTaskID(badPath, publicStandardSHA, sourceSHA, profile, publicNoiseFidelity003FastSHA,
		profile.ConfigSHA256, profile.QPSHA256, profile.InputSHA256, profile.WorkloadSHA256, false, publicNoiseFidelity003TaskID, "fast", fastBinary); err == nil {
		t.Fatal("003 preflight relabeled the frozen conservative q0 NOT_PROVEN result")
	}
}

func TestNoiseFidelity003LogN16TripletCompatibilityIsTaskAndProfileBound(t *testing.T) {
	profile, err := publicProfileFor("logn16")
	if err != nil {
		t.Fatal(err)
	}
	config, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN16.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, _, effective, err := perfmeasure.ParametersFromConfigWithE(config, 32)
	if err != nil {
		t.Fatal(err)
	}
	oracles, err := publicOracleCheckpoints(profile.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	input := encodeVector(oracles["drop_level0"])
	inputSHA := perfmeasure.Fingerprint(decodeComplexValues(input))
	const (
		fastBinary = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		sourceSHA  = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		preflight  = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	)
	mode := "public-native-noise-fidelity-003-logn16-repeatability"
	var docs [3]publicNativeDocument
	var vectors [3]publicVectorDocument
	for i, backend := range []string{"standard", "fast", "fast"} {
		commit := publicStandardSHA
		binary := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		if backend == "fast" {
			commit, binary = publicNoiseFidelity003FastSHA, fastBinary
		}
		docs[i] = publicNativeDocument{
			TaskID: publicNoiseFidelity003TaskID, Mode: mode, BootstrapBudget: 6, BootstrapCalls: 6, Profile: "logn16",
			EphemeralSecretWeight: 32, PrimaryCommit: publicStandardSHA, PrimaryEvidenceCommit: publicBootstrapNoisePilotFastSHA,
			PrimarySourceSHA256: sourceSHA, Backend: backend, BackendCommit: commit, BinarySHA256: binary,
			Parameters: effective, ConfigSHA256: profile.ConfigSHA256, QPSHA256: profile.QPSHA256,
			InputSHA256: profile.InputSHA256, WorkloadSHA256: profile.WorkloadSHA256, BootstrapInputSHA256: inputSHA,
			PreflightPairSHA256: preflight, CapacityPolicy: publicCapacityPolicyLogN16Observed,
			ExecutionPhases: []phaseTiming{
				{Phase: "public_ckks_evaluator_construction", ElapsedNS: 100, Samples: 1, Available: true},
				{Phase: "public_bootstrapping_evaluator_construction", ElapsedNS: 200, Samples: 1, Available: true},
			},
		}
		if i == 2 {
			docs[i].ExecutionPhases = append(docs[i].ExecutionPhases, phaseTiming{Phase: "bootstrap_noise_hook_configuration", ElapsedNS: 10, Samples: 1, Available: true})
		}
		vectors[i] = publicVectorDocument{
			TaskID: publicNoiseFidelity003TaskID, Mode: mode, BootstrapInputSHA256: inputSHA,
			ConfigSHA256: profile.ConfigSHA256, QPSHA256: profile.QPSHA256, InputSHA256: profile.InputSHA256,
			WorkloadSHA256: profile.WorkloadSHA256, BootstrapPlaintextInput: input,
		}
	}
	if err := validatePublicTripletCompatibilityTask(docs[0], docs[1], docs[2], vectors[0], vectors[1], vectors[2], publicNoiseFidelity003TaskID); err != nil {
		t.Fatalf("valid exact-profile 003 triplet provenance rejected: %v", err)
	}
	docs[1].BackendCommit = publicNoiseFidelity002FastSHA
	if err := validatePublicTripletCompatibilityTask(docs[0], docs[1], docs[2], vectors[0], vectors[1], vectors[2], publicNoiseFidelity003TaskID); err == nil {
		t.Fatal("003 triplet accepted the 002 Fast source pin")
	}
	docs[1].BackendCommit = publicNoiseFidelity003FastSHA
	docs[1].Parameters.LogN = 15
	if err := validatePublicTripletCompatibilityTask(docs[0], docs[1], docs[2], vectors[0], vectors[1], vectors[2], publicNoiseFidelity003TaskID); err == nil {
		t.Fatal("003 triplet accepted mismatched LogN dimensions")
	}
}

func TestNoiseFidelity003LogN16ComparatorConsumesSixSamplesPerArm(t *testing.T) {
	standard, standardVectors := makeNoiseFidelity003BaseArtifacts(t, "standard")
	fast, fastVectors := makeNoiseFidelity003BaseArtifacts(t, "fast")
	preflightPaths := []string{
		filepath.Join(t.TempDir(), "standard.json"), filepath.Join(t.TempDir(), "fast.json"),
		filepath.Join(t.TempDir(), "standard-vectors.json"), filepath.Join(t.TempDir(), "fast-vectors.json"),
		filepath.Join(t.TempDir(), "pair.json"),
	}
	for _, item := range []struct {
		path  string
		value any
	}{
		{preflightPaths[0], standard}, {preflightPaths[1], fast},
		{preflightPaths[2], standardVectors}, {preflightPaths[3], fastVectors},
	} {
		if err := writeExclusiveJSON(item.path, item.value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runPublicCompare([]string{
		"--standard=" + preflightPaths[0], "--fast=" + preflightPaths[1],
		"--standard-vectors=" + preflightPaths[2], "--fast-vectors=" + preflightPaths[3],
		"--out=" + preflightPaths[4], "--noise-fidelity-003-logn16-preflight",
	}); err != nil {
		t.Fatalf("valid synthetic zero-call LogN16 pair rejected: %v", err)
	}
	pairData, err := os.ReadFile(preflightPaths[4])
	if err != nil {
		t.Fatal(err)
	}
	preflightSHA := fmt.Sprintf("%x", sha256.Sum256(pairData))

	standard, standardVectors = promoteNoiseFidelity003Capture(t, standard, standardVectors, "standard", preflightSHA, 2e-6, false)
	fastBase, fastBaseVectors := fast, fastVectors
	fast, fastVectors = promoteNoiseFidelity003Capture(t, fastBase, fastBaseVectors, "fast", preflightSHA, 4e-6, false)
	noise, noiseVectors := promoteNoiseFidelity003Capture(t, fastBase, fastBaseVectors, "fast", preflightSHA, 1.5e-6, true)
	noise.BootstrapNoiseEvidence = makeNoiseFidelity003Snapshot(t, noise.Parameters)

	paths := []string{
		filepath.Join(t.TempDir(), "s.json"), filepath.Join(t.TempDir(), "sv.json"),
		filepath.Join(t.TempDir(), "f.json"), filepath.Join(t.TempDir(), "fv.json"),
		filepath.Join(t.TempDir(), "fn.json"), filepath.Join(t.TempDir(), "fnv.json"),
		filepath.Join(t.TempDir(), "triplet.json"),
	}
	for _, item := range []struct {
		path  string
		value any
	}{
		{paths[0], standard}, {paths[1], standardVectors}, {paths[2], fast}, {paths[3], fastVectors},
		{paths[4], noise}, {paths[5], noiseVectors},
	} {
		if err := writeExclusiveJSON(item.path, item.value); err != nil {
			t.Fatal(err)
		}
	}
	if err := runPublicTripletCompare([]string{
		"--standard=" + paths[0], "--standard-vectors=" + paths[1],
		"--fast=" + paths[2], "--fast-vectors=" + paths[3],
		"--fast-noise=" + paths[4], "--fast-noise-vectors=" + paths[5],
		"--preflight-pair=" + preflightPaths[4], "--out=" + paths[6], "--noise-fidelity-003-logn16",
	}); err != nil {
		t.Fatalf("valid synthetic six-sample LogN16 triplet rejected: %v", err)
	}
	var result publicTripletFidelity002Document
	if err := readJSON(paths[6], &result); err != nil {
		t.Fatal(err)
	}
	if result.SchemaVersion != "fast-bootstrap-noise-fidelity-003-logn16-triplet.v1" || result.TaskID != publicNoiseFidelity003TaskID ||
		result.Profile != "logn16" || result.NoiseRootSeed != publicNoiseFidelity003RootSeed || len(result.Arms["S"].Samples) != 6 ||
		len(result.Arms["F"].Samples) != 6 || len(result.Arms["FN"].Samples) != 6 || result.Arms["FN"].Cold.NumericalGatePass {
		t.Fatalf("unexpected 003 six-sample result contract: schema=%s task=%s profile=%s", result.SchemaVersion, result.TaskID, result.Profile)
	}
}

func makeNoiseFidelity003BaseArtifacts(t *testing.T, backend string) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	config, configSHA, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN16.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, _, effective, err := perfmeasure.ParametersFromConfigWithE(config, 32)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := publicProfileFor("logn16")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", configSHA) != profile.ConfigSHA256 {
		t.Fatal("LogN16 config fixture hash differs from frozen profile")
	}
	oracles, err := publicOracleCheckpoints(profile.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	a, b, c, err := perfmeasure.PublicMulRescaleWorkload(profile.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	qPrimes, err := parseEffectiveQPrimes(effective)
	if err != nil {
		t.Fatal(err)
	}
	capacity, err := assessPublicCapacityForProfile(profile, a, b, c, qPrimes, effective.RingN)
	if err != nil {
		t.Fatal(err)
	}
	continuity, err := comparePublicLevel0Continuity(oracles["rotate_level4"], oracles["drop_level0"])
	if err != nil {
		t.Fatal(err)
	}
	backendCommit, buildTag := publicNoiseFidelity003FastSHA, "perf_fast"
	if backend == "standard" {
		backendCommit, buildTag = publicStandardSHA, "perf_standard"
	}
	inputKind, constructor := publicInputContract(backend)
	checkpoints := make([]publicCheckpoint, 0, 8)
	checkpointVectors := make(map[string][]complexValue, 8)
	for _, name := range []string{"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5", "mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0"} {
		level := map[string]int{"encrypt_a_level5": 5, "encrypt_b_level5": 5, "encrypt_c_q5_level5": 5, "add_level5": 5, "mulrelin_level5": 5, "rescale_q5_level4": 4, "rotate_level4": 4, "drop_level0": 0}[name]
		activeRows, physicalRows := min(level+1, 4), level+1
		compact := backend == "fast" && name != "encrypt_a_level5" && name != "encrypt_b_level5" && name != "encrypt_c_q5_level5"
		if backend == "fast" && !compact {
			physicalRows = len(effective.QPrimes)
		}
		rowLengths, rowHashes := make([][]int, 2), make([][]string, 2)
		for component := 0; component < 2; component++ {
			rowLengths[component] = make([]int, physicalRows)
			rowHashes[component] = make([]string, activeRows)
			for row := range rowLengths[component] {
				if row < activeRows || (backend == "fast" && !compact) {
					rowLengths[component][row] = effective.RingN
				}
			}
			for row := range rowHashes[component] {
				rowHashes[component][row] = fmt.Sprintf("%s-%d-%d", name, component, row)
			}
		}
		decodePath := "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)"
		if backend == "standard" {
			decodePath = "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode"
		} else if name == "drop_level0" {
			decodePath = "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode"
		}
		scale, err := expectedPublicCheckpointScale(effective, name)
		if err != nil {
			t.Fatal(err)
		}
		values := oracles[name]
		checkpoints = append(checkpoints, publicCheckpoint{
			Name: name, State: ciphertextState{Level: level, Degree: 1, Scale: scale, QPrefixRows: activeRows},
			DecodePath: decodePath, PhysicalRowLengths: rowLengths, RowSHA256: rowHashes,
			FastCompactPrefix: compact, C1Nonzero: backend == "standard", DecodedSHA256: perfmeasure.Fingerprint(values),
			Oracle: compareVectors(values, values, publicNumericalGate), OracleSNR: numericalmetrics.Compare(values, values), OraclePass: true,
		})
		checkpointVectors[name] = encodeVector(values)
	}
	primarySourceSHA := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	binarySHA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if backend == "fast" {
		binarySHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	}
	doc := publicNativeDocument{
		SchemaVersion: "fast-standard-public-native-preflight.v1", TaskID: publicNoiseFidelity003TaskID,
		Mode: "public-native", Status: "PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN", Profile: "logn16", ProfileKind: publicProfileKindFrozen,
		Backend: backend, BackendCommit: backendCommit, BackendClean: true, PrimaryClean: true, BinarySHA256: binarySHA,
		PrimaryCommit: publicStandardSHA, PrimaryEvidenceCommit: publicBootstrapNoisePilotFastSHA, PrimarySourceSHA256: primarySourceSHA,
		BuildTag: buildTag, ConfigSHA256: profile.ConfigSHA256, QPSHA256: profile.QPSHA256,
		InputSHA256: profile.InputSHA256, WorkloadSHA256: profile.WorkloadSHA256,
		Parameters: effective, EphemeralSecretWeight: 32, InputSlots: profile.InputSlots, InputKind: inputKind, InputConstructor: constructor,
		InputC1Nonzero: backend == "standard", NumericalGate: publicNumericalGate,
		CapacityEvidenceVersion: publicCapacityEvidenceVersion, CapacityPolicy: profile.CapacityPolicy,
		ConservativeBoundStatus: capacity.TerminalQ0Status, ObservedNativeOracleStatus: publicObservedOraclePassed,
		Level0Continuity: continuity, Capacity: capacity, BootstrapBudget: 0, BootstrapCalls: 0, Checkpoints: checkpoints,
	}
	vectors := publicVectorDocument{
		SchemaVersion: "fast-standard-public-native-vectors.v1", TaskID: publicNoiseFidelity003TaskID, Mode: "public-native",
		Backend: backend, BackendCommit: backendCommit, PrimaryCommit: doc.PrimaryCommit, PrimaryEvidenceCommit: doc.PrimaryEvidenceCommit,
		SourceSHA256: primarySourceSHA, ConfigSHA256: profile.ConfigSHA256, QPSHA256: profile.QPSHA256,
		InputSHA256: profile.InputSHA256, WorkloadSHA256: profile.WorkloadSHA256, E: 32, Checkpoints: checkpointVectors,
	}
	return doc, vectors
}

func promoteNoiseFidelity003Capture(t *testing.T, doc publicNativeDocument, vectors publicVectorDocument, backend, preflightSHA string, offset float64, noiseEnabled bool) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	doc.SchemaVersion, doc.Mode, doc.Status = "fast-standard-public-native-noise-fidelity-003-logn16-repeatability.v1", "public-native-noise-fidelity-003-logn16-repeatability", "OBSERVED_PUBLIC_BOOTSTRAP_1COLD_5WARM"
	doc.TaskID, doc.BootstrapBudget, doc.BootstrapCalls, doc.PreflightPairSHA256 = publicNoiseFidelity003TaskID, 6, 6, preflightSHA
	oracles, err := publicOracleCheckpoints(doc.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	input := oracles["drop_level0"]
	doc.BootstrapInputSHA256 = perfmeasure.Fingerprint(input)
	doc.ExecutionPhases = []phaseTiming{
		{Phase: "public_ckks_evaluator_construction", ElapsedNS: 100, Samples: 1, Available: true},
		{Phase: "public_bootstrapping_evaluator_construction", ElapsedNS: 200, Samples: 1, Available: true},
	}
	if noiseEnabled {
		doc.ExecutionPhases = append(doc.ExecutionPhases, phaseTiming{Phase: "bootstrap_noise_hook_configuration", ElapsedNS: 10, Samples: 1, Available: true})
	}
	bootstrapInput, err := expectedPublicCheckpointScale(doc.Parameters, "bootstrap_first_cold")
	if err != nil || bootstrapInput == "" {
		t.Fatalf("Bootstrap output Scale fixture: %v", err)
	}
	wants, err := publicBootstrapAttemptSpecs(6)
	if err != nil {
		t.Fatal(err)
	}
	doc.BootstrapResults = make([]publicBootstrapResult, 0, len(wants))
	vectors.TaskID, vectors.Mode = publicNoiseFidelity003TaskID, doc.Mode
	vectors.SchemaVersion = "fast-standard-public-native-noise-fidelity-003-logn16-repeatability-vectors.v1"
	vectors.BootstrapInputSHA256, vectors.BootstrapPlaintextInput = doc.BootstrapInputSHA256, encodeVector(input)
	vectors.Backend, vectors.BackendCommit = backend, doc.BackendCommit
	vectors.PrimaryCommit, vectors.PrimaryEvidenceCommit, vectors.SourceSHA256 = doc.PrimaryCommit, doc.PrimaryEvidenceCommit, doc.PrimarySourceSHA256
	vectors.BootstrapOutputs = make(map[string][]complexValue, len(wants))
	for index, want := range wants {
		output := append([]complex128(nil), input...)
		output[0] += complex(offset+float64(index)*1e-9, 0)
		rowLengths, rowHashes := [][]int{{doc.Parameters.RingN, doc.Parameters.RingN}, {doc.Parameters.RingN, doc.Parameters.RingN}}, [][]string{{"q0", "q1"}, {"q0", "q1"}}
		decodePath := "rlwe.NewDecryptor(residual parameters, generated secret).DecryptNew -> ckks.NewEncoder.Decode"
		if backend == "fast" {
			decodePath = "rlwe.NewDecryptor(residual parameters, matching zero-secret-mode secret).DecryptNew -> ckks.NewEncoder.Decode"
		}
		checkpoint := publicCheckpoint{
			Name: want.name, State: ciphertextState{Level: 1, Degree: 1, Scale: bootstrapInput, QPrefixRows: 2},
			DecodePath: decodePath, PhysicalRowLengths: rowLengths, RowSHA256: rowHashes,
			FastCompactPrefix: backend == "fast", C1Nonzero: backend == "standard", DecodedSHA256: perfmeasure.Fingerprint(output),
			Oracle: compareVectors(input, output, publicNumericalGate), OracleSNR: numericalmetrics.Compare(input, output),
			OraclePass: compareVectors(input, output, publicNumericalGate).MaxComplexDifference <= publicNumericalGate,
		}
		doc.BootstrapResults = append(doc.BootstrapResults, publicBootstrapResult{
			Phase: want.phase, InputCiphertextSHA256: "held-n16-bootstrap-input-ciphertext-sha",
			Timing: phaseTiming{Phase: want.phase, ElapsedNS: int64(1000 + index*100), AllocBytes: uint64(1 << 20), Allocs: 100, Samples: 1, Available: true}, Output: checkpoint,
		})
		vectors.BootstrapOutputs[want.name] = encodeVector(output)
	}
	return doc, vectors
}

func makeNoiseFidelity003Snapshot(t *testing.T, parameters perfmeasure.EffectiveParameters) json.RawMessage {
	t.Helper()
	qPrimes, err := parseEffectiveQPrimes(parameters)
	if err != nil {
		t.Fatal(err)
	}
	product := big.NewInt(1)
	for _, prime := range qPrimes[:4] {
		product.Mul(product, new(big.Int).SetUint64(prime))
	}
	phase := []bootstrapNoisePhaseEvidence{
		{Name: "sampling_hash_moments", ElapsedNS: 10}, {Name: "coefficient_to_qprefix_ntt", ElapsedNS: 20},
		{Name: "candidate_c0_add", ElapsedNS: 30}, {Name: "qprefix_centered_capacity_check", ElapsedNS: 40}, {Name: "commit_c0", ElapsedNS: 0},
	}
	leftSeed := expectedBootstrapNoiseSamplingSeed(publicNoiseFidelity003RootSeed, 1, false)
	rightSeed := expectedBootstrapNoiseSamplingSeed(publicNoiseFidelity003RootSeed, 1, true)
	left, right := bootstrapNoiseTestPolynomial(leftSeed), bootstrapNoiseTestPolynomial(rightSeed)
	mod1Scale := rlwe.NewScale(math.Exp2(60))
	for _, side := range []*bootstrapNoisePolynomialEvidence{&left, &right} {
		side.Level = 12
		side.Scale = mod1Scale.Value.Text('g', -1)
		side.SampleCount = 1 << 16
		side.Capacity.CheckedCoefficients = 1 << 16
		side.Capacity.PrefixProduct = product.String()
		side.SamplingAlgorithm = "deterministic Gaussian test fixture"
		side.DispatchEvidence = "NTT and MForm-mapped coefficient noise"
		side.CoefficientMean = 0
		side.CoefficientVariance = publicBootstrapNoiseSigma * publicBootstrapNoiseSigma
		side.CoefficientMaxAbs = 12
		side.Phases = phase
	}
	encoded, err := json.Marshal(bootstrapNoiseSnapshotEvidence{
		Enabled: true, Sigma: publicBootstrapNoiseSigma, RootSeed: publicNoiseFidelity003RootSeed,
		CtCtMultiplyCalls: 1, InjectedPolynomials: 2,
		Events: []bootstrapNoiseEventEvidence{{Operation: 1, Level: 12, Scale: left.Scale, Left: left, Right: right, MultiplyCompleted: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func TestPrimaryTaskIdentityAllowsEvidenceOnlyCommitAfterCleanCodeBuild(t *testing.T) {
	const sourceCommit = "ca9fc012e68d23cf1573e4f0ff26e2ddc05fa0af"
	const evidenceCommit = "e0e6ade64a75b1a19874533fbdcb3d2daca5f818"
	identity, err := primaryMeasurementIdentityFromVCS(".", evidenceCommit, sourceCommit, "false")
	if err != nil {
		t.Fatalf("clean source build with evidence-only descendant rejected: %v", err)
	}
	if identity.SourceCommit != sourceCommit || identity.EvidenceCommit != evidenceCommit || identity.SourceSHA256 == "" {
		t.Fatalf("source/evidence commit identities were conflated: %+v", identity)
	}
	if _, err := primaryMeasurementIdentityFromVCS(".", evidenceCommit, sourceCommit, "true"); err == nil {
		t.Fatal("task mode accepted a binary built from a modified source tree")
	}
}

func promoteToNoiseFidelity002Fixture(t *testing.T, doc publicNativeDocument, vectors publicVectorDocument, backend, commit, binarySHA string, offset float64, noiseEnabled bool) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	doc, vectors = promotePublicBootstrapFixtureToRepeatability(t, doc, vectors)
	doc.SchemaVersion, doc.Mode, doc.Status = "fast-standard-public-native-noise-fidelity-002-repeatability.v1", "public-native-noise-fidelity-002-repeatability", "OBSERVED_PUBLIC_BOOTSTRAP_1COLD_5WARM"
	doc.TaskID, doc.BootstrapBudget, doc.BootstrapCalls = publicNoiseFidelity002TaskID, 6, 6
	doc.PrimaryCommit, vectors.PrimaryCommit = publicStandardSHA, publicStandardSHA
	doc.PrimaryEvidenceCommit, vectors.PrimaryEvidenceCommit = publicFixedSigmaFastSHA, publicFixedSigmaFastSHA
	doc.BootstrapInputSHA256 = publicBootstrapInputSHA
	doc.PreflightPairSHA256 = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	doc.BinarySHA256, doc.Backend, doc.BackendCommit = binarySHA, backend, commit
	doc.ExecutionPhases = []phaseTiming{
		{Phase: "public_ckks_evaluator_construction", ElapsedNS: 1000, AllocBytes: 100, Allocs: 2, Samples: 1, Available: true},
		{Phase: "public_bootstrapping_evaluator_construction", ElapsedNS: 2000, AllocBytes: 200, Allocs: 3, Samples: 1, Available: true},
	}
	if noiseEnabled {
		doc.ExecutionPhases = append(doc.ExecutionPhases, phaseTiming{Phase: "bootstrap_noise_hook_configuration", ElapsedNS: 50, AllocBytes: 10, Allocs: 1, Samples: 1, Available: true})
	}
	bootstrapInput, err := publicOracleCheckpoints(doc.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	input := bootstrapInput["drop_level0"]
	vectors.TaskID, vectors.BootstrapInputSHA256 = publicNoiseFidelity002TaskID, doc.BootstrapInputSHA256
	vectors.BootstrapPlaintextInput = encodeVector(input)
	vectors.Backend, vectors.BackendCommit = backend, commit
	vectors.SchemaVersion, vectors.Mode = "fast-standard-public-native-noise-fidelity-002-repeatability-vectors.v1", doc.Mode
	for index := range doc.BootstrapResults {
		result := &doc.BootstrapResults[index]
		output := append([]complex128(nil), input...)
		output[0] += complex(offset+float64(index)*1e-9, 0)
		vectors.BootstrapOutputs[result.Output.Name] = encodeVector(output)
		result.Output.DecodedSHA256 = perfmeasure.Fingerprint(output)
		result.Output.Oracle = compareVectors(input, output, publicNumericalGate)
		result.Output.OracleSNR = numericalmetrics.Compare(input, output)
		result.Output.OraclePass = result.Output.Oracle.MaxComplexDifference <= publicNumericalGate
	}
	return doc, vectors
}

func makeNoiseFidelity002Snapshot(t *testing.T, parameters perfmeasure.EffectiveParameters) json.RawMessage {
	t.Helper()
	phase := []bootstrapNoisePhaseEvidence{{Name: "sampling_hash_moments", ElapsedNS: 10}, {Name: "coefficient_to_qprefix_ntt", ElapsedNS: 20}, {Name: "candidate_c0_add", ElapsedNS: 30}, {Name: "qprefix_centered_capacity_check", ElapsedNS: 40}, {Name: "commit_c0", ElapsedNS: 0}}
	left, right := bootstrapNoiseTestPolynomial(expectedBootstrapNoiseSamplingSeed(publicNoiseFidelity002RootSeed, 1, false)), bootstrapNoiseTestPolynomial(expectedBootstrapNoiseSamplingSeed(publicNoiseFidelity002RootSeed, 1, true))
	left.SamplingAlgorithm, right.SamplingAlgorithm = "deterministic Gaussian test fixture", "deterministic Gaussian test fixture"
	left.DispatchEvidence, right.DispatchEvidence = "NTT transform and MForm-mapped Montgomery strict capacity test fixture", "NTT transform and MForm-mapped Montgomery strict capacity test fixture"
	left.CoefficientMean, right.CoefficientMean = 0, 0
	left.CoefficientVariance, right.CoefficientVariance = publicBootstrapNoiseSigma*publicBootstrapNoiseSigma, publicBootstrapNoiseSigma*publicBootstrapNoiseSigma
	left.CoefficientMaxAbs, right.CoefficientMaxAbs = 12, 12
	left.Phases, right.Phases = phase, phase
	qPrimes, err := parseEffectiveQPrimes(parameters)
	if err != nil {
		t.Fatal(err)
	}
	product := big.NewInt(1)
	for _, prime := range qPrimes[:left.Capacity.Rows] {
		product.Mul(product, new(big.Int).SetUint64(prime))
	}
	left.Capacity.PrefixProduct, right.Capacity.PrefixProduct = product.String(), product.String()
	encoded, err := json.Marshal(bootstrapNoiseSnapshotEvidence{
		Enabled: true, Sigma: publicBootstrapNoiseSigma, RootSeed: publicNoiseFidelity002RootSeed,
		CtCtMultiplyCalls: 1, InjectedPolynomials: 2,
		Events: []bootstrapNoiseEventEvidence{{Operation: 1, Level: 3, Scale: "2^45", Left: left, Right: right, MultiplyCompleted: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func prepareTripletFixtureMetadata(doc *publicNativeDocument, vectors *publicVectorDocument, backend, commit, binarySHA string) {
	doc.Backend, doc.BackendCommit, doc.BinarySHA256 = backend, commit, binarySHA
	doc.PrimaryCommit, doc.PrimarySourceSHA256 = "primary-task-commit", "source-fingerprint"
	doc.PrimaryClean, doc.BackendClean = true, true
	vectors.Backend, vectors.BackendCommit = backend, commit
	vectors.PrimaryCommit, vectors.SourceSHA256 = doc.PrimaryCommit, doc.PrimarySourceSHA256
}

func installPublicTripletOutput(t *testing.T, doc *publicNativeDocument, vectors *publicVectorDocument, input []complex128, offset float64) {
	t.Helper()
	values := append([]complex128(nil), input...)
	values[0] += complex(offset, 0)
	encoded := encodeVector(values)
	vectors.BootstrapOutputs["bootstrap_once"] = encoded
	result := &doc.BootstrapResults[0]
	result.Output.DecodedSHA256 = perfmeasure.Fingerprint(values)
	result.Output.Oracle = compareVectors(input, values, publicNumericalGate)
	result.Output.OracleSNR = numericalmetrics.Compare(input, values)
	result.Output.OraclePass = result.Output.Oracle.MaxComplexDifference <= publicNumericalGate
}

func bootstrapNoiseTestPolynomial(seed int64) bootstrapNoisePolynomialEvidence {
	return bootstrapNoisePolynomialEvidence{
		Sigma: publicBootstrapNoiseSigma, Seed: seed, Level: 3, Scale: "2^45", SampleCount: 1 << 13, CoefficientRMS: publicBootstrapNoiseSigma,
		CoefficientSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Capacity:          bootstrapNoiseCapacityEvidence{Rows: 4, CheckedCoefficients: 1 << 13, PrefixProduct: "1155", StrictFit: true},
	}
}

func promotePublicBootstrapFixtureToOnce(t *testing.T, doc publicNativeDocument, vectors publicVectorDocument) (publicNativeDocument, publicVectorDocument) {
	t.Helper()
	oracles, err := publicOracleCheckpoints(doc.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	bootstrapInput := oracles["drop_level0"]
	doc.SchemaVersion, doc.Mode, doc.Status = "fast-standard-public-native-bootstrap-once.v2", "public-native-bootstrap-once", publicBootstrapOnceObservedStatus
	doc.BootstrapBudget, doc.BootstrapCalls = 1, 1
	doc.BootstrapInputSHA256 = perfmeasure.Fingerprint(bootstrapInput)
	result := doc.BootstrapResults[0]
	result.Phase, result.Timing.Phase, result.Output.Name = "single_bootstrap", "single_bootstrap", "bootstrap_once"
	doc.BootstrapResults = []publicBootstrapResult{result}
	vectors.SchemaVersion, vectors.Mode = "fast-standard-public-native-bootstrap-once-vectors.v1", "public-native-bootstrap-once"
	vectors.BootstrapInputSHA256 = doc.BootstrapInputSHA256
	vectors.BootstrapPlaintextInput = encodeVector(bootstrapInput)
	values := vectors.BootstrapOutputs["bootstrap_first_cold"]
	vectors.BootstrapOutputs = map[string][]complexValue{"bootstrap_once": values}
	return doc, vectors
}
