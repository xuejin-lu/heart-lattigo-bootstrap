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
		result.NoiseColdTimeOverheadRatio <= 0 || result.NoiseWarmMedianAllocOverheadBytes != 0 {
		t.Fatalf("exploratory metrics/Noise accounting missing: %+v", result)
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
	phase := []bootstrapNoisePhaseEvidence{{Name: "sampling_hash_moments", ElapsedNS: 10}, {Name: "coefficient_to_qprefix_ntt", ElapsedNS: 20}, {Name: "candidate_c0_add", ElapsedNS: 30}, {Name: "qprefix_centered_capacity_check", ElapsedNS: 40}, {Name: "commit_c0", ElapsedNS: 5}}
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
