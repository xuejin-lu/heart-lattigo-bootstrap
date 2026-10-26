package main

import (
	"encoding/json"
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
		Backend: "fast", BackendCommit: publicBootstrapNoiseFastSHA, BinarySHA256: string(make([]byte, 64)),
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
	prepareTripletFixtureMetadata(&fast, &fastVectors, "fast", publicBootstrapNoiseFastSHA, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	prepareTripletFixtureMetadata(&noise, &noiseVectors, "fast", publicBootstrapNoiseFastSHA, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	bootstrapInput, err := publicOracleCheckpoints(standard.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	canonical := bootstrapInput["drop_level0"]
	installPublicTripletOutput(t, &standard, &standardVectors, canonical, 1e-9)
	installPublicTripletOutput(t, &fast, &fastVectors, canonical, 1e-8)
	installPublicTripletOutput(t, &noise, &noiseVectors, canonical, 2e-9)
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
	if result.Status != "TRIPLET_COMPLETE_EXPLORATORY" || len(result.Arms) != 3 || len(result.Pairwise) != 3 ||
		result.Arms["S"].SuccessfulSlots != 1<<12 || result.Arms["F"].BootstrapTimeNS != 100 || result.NoiseRootSeed == nil || *result.NoiseRootSeed != 20261027 {
		t.Fatalf("triplet summary is incomplete: %+v", result)
	}
	if result.Arms["F"].Error.RMSE <= result.Arms["FN"].Error.RMSE || result.Arms["FN"].CtCtMultiplyCalls != 1 ||
		result.Arms["FN"].InjectedPolynomials != 2 || result.Arms["FN"].SigmaCoefficient != publicBootstrapNoiseSigma {
		t.Fatalf("triplet summary lost fixed-sigma or error metrics: %+v", result.Arms)
	}
	if result.PrecisionDistanceImprovement == nil || *result.PrecisionDistanceImprovement <= 0 {
		t.Fatalf("synthetic FN should be closer in absolute precision bits: %+v", result)
	}

	noise.BinarySHA256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	if err := validatePublicBootstrapTripletCompatibility(standard, fast, noise, standardVectors, fastVectors, noiseVectors); err == nil {
		t.Fatal("triplet accepted Fast and Fast+Noise artifacts from different binaries")
	}
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
	doc.SchemaVersion, doc.Mode, doc.Status = "fast-standard-public-native-bootstrap-once.v1", "public-native-bootstrap-once", "PASS_PUBLIC_BOOTSTRAP_ONCE"
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
