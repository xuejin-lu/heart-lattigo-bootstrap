package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	commonpolynomial "github.com/tuneinsight/lattigo/v6/circuits/common/polynomial"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestValidatePublicE32FixturePinsHeldInputAndFastVectors(t *testing.T) {
	root := t.TempDir()
	parameters := []byte("serialized public E32 parameters")
	ciphertext := []byte("serialized held Level0 ciphertext")
	require.NoError(t, os.WriteFile(filepath.Join(root, "parameters.bin"), parameters, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ciphertext.bin"), ciphertext, 0o600))
	parametersSum, ciphertextSum := sha256.Sum256(parameters), sha256.Sum256(ciphertext)
	decoded := make([]complex128, 1<<12)
	input := make([]publicE32Vector, len(decoded))
	for index := range decoded {
		input[index] = publicE32Vector{Real: real(decoded[index]), Imag: imag(decoded[index])}
	}
	vectors := publicE32Vectors{
		SchemaVersion: "fast-standard-public-native-repeatability-vectors.v1", Mode: "public-native-repeatability",
		Backend: "fast", BackendCommit: e32ProductionFastCommit, PrimaryCommit: "primary",
		SourceSHA256: "source", ConfigSHA256: "config", QPSHA256: "qp", InputSHA256: "input", WorkloadSHA256: "workload",
		Checkpoints:      map[string][]publicE32Vector{"drop_level0": input},
		BootstrapOutputs: map[string][]publicE32Vector{},
	}
	for _, name := range []string{"bootstrap_first_cold", "bootstrap_warm_01", "bootstrap_warm_02", "bootstrap_warm_03", "bootstrap_warm_04", "bootstrap_warm_05"} {
		vectors.BootstrapOutputs[name] = input
	}
	vectorsBytes, err := json.Marshal(vectors)
	require.NoError(t, err)
	vectorsPath := filepath.Join(root, "vectors.json")
	require.NoError(t, os.WriteFile(vectorsPath, vectorsBytes, 0o600))
	manifest := publicE32FixtureManifest{
		SchemaVersion: "fast-public-e32-trace-fixture.v1", Profile: "logn13-e32-public-native", Backend: "fast",
		BackendCommit: e32ProductionFastCommit, PrimaryCommit: "primary", PrimarySourceSHA256: "source",
		ConfigSHA256: "config", QPSHA256: "qp", InputSHA256: "input", WorkloadSHA256: "workload",
		ParametersFile: "parameters.bin", ParametersSHA256: hex.EncodeToString(parametersSum[:]),
		CiphertextFile: "ciphertext.bin", CiphertextSHA256: hex.EncodeToString(ciphertextSum[:]),
		DecodedSHA256: perfmeasure.Fingerprint(decoded), State: publicE32CipherState{Level: 0, Degree: 1, QPrefixRows: 1, Scale: "2^45", PrefixQ: "q0"},
		NumericalGate: 1e-6, EphemeralSecretWeight: 32,
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	manifestPath := filepath.Join(root, "manifest.json")
	require.NoError(t, os.WriteFile(manifestPath, manifestBytes, 0o600))
	gotManifestBytes, gotManifest, gotVectorsBytes, gotVectors, err := validatePublicE32Fixture(manifestPath, vectorsPath, "primary", "source")
	require.NoError(t, err)
	require.Equal(t, manifestBytes, gotManifestBytes)
	require.Equal(t, manifest.CiphertextSHA256, gotManifest.CiphertextSHA256)
	require.Equal(t, vectorsBytes, gotVectorsBytes)
	require.Len(t, gotVectors.BootstrapOutputs, 6)
	_, _, _, _, err = validatePublicE32Fixture(manifestPath, vectorsPath, "primary", "different-source")
	require.ErrorContains(t, err, "measurement-source fingerprint")

	vectors.BackendCommit = "wrong-fast-commit"
	vectorsBytes, err = json.Marshal(vectors)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(vectorsPath, vectorsBytes, 0o600))
	_, _, _, _, err = validatePublicE32Fixture(manifestPath, vectorsPath, "primary", "source")
	require.ErrorContains(t, err, "do not match")
}

func TestPublicE32TraceReservationsPrecedeSubprocess(t *testing.T) {
	base := filepath.Join(t.TempDir(), "trace")
	launched := false
	err := runReservedPublicE32Trace(base, func() error {
		launched = true
		for index := 1; index <= 2; index++ {
			path := fmt.Sprintf("%s.bootstrap-attempt-%02d.json", base, index)
			data, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			var reservation struct {
				SchemaVersion string `json:"schema_version"`
				Index         int    `json:"attempt_index"`
				Budget        int    `json:"bootstrap_budget"`
				Status        string `json:"status"`
			}
			require.NoError(t, json.Unmarshal(data, &reservation))
			require.Equal(t, index, reservation.Index)
			require.Equal(t, 2, reservation.Budget)
			require.Equal(t, "irrevocably_reserved_before_call", reservation.Status)
		}
		return nil
	})
	require.NoError(t, err)
	require.True(t, launched)

	blockedBase := filepath.Join(t.TempDir(), "blocked")
	secondPath := blockedBase + ".bootstrap-attempt-02.json"
	require.NoError(t, os.WriteFile(secondPath, []byte("preserve"), 0o600))
	launched = false
	err = runReservedPublicE32Trace(blockedBase, func() error { launched = true; return nil })
	require.Error(t, err)
	require.False(t, launched)
	firstBytes, readErr := os.ReadFile(blockedBase + ".bootstrap-attempt-01.json")
	require.NoError(t, readErr, "the first token remains irrevocably consumed after partial reservation")
	require.NotEmpty(t, firstBytes)
	secondBytes, readErr := os.ReadFile(secondPath)
	require.NoError(t, readErr)
	require.Equal(t, "preserve", string(secondBytes))
}

func TestReadProfileArtifactChecksPathAndDigest(t *testing.T) {
	dir := t.TempDir()
	data := []byte("scoped profile fixture")
	path := filepath.Join(dir, "warm-cpu.pprof")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	sum := sha256.Sum256(data)
	gotPath, err := readProfileArtifact(dir, filepath.Base(path), hex.EncodeToString(sum[:]))
	require.NoError(t, err)
	require.Equal(t, path, gotPath)
	_, err = readProfileArtifact(dir, "../warm-cpu.pprof", hex.EncodeToString(sum[:]))
	require.ErrorContains(t, err, "filename")
	_, err = readProfileArtifact(dir, filepath.Base(path), strings.Repeat("0", 64))
	require.ErrorContains(t, err, "SHA-256 mismatch")
}

func TestValidatePublicE32EventsChecksHierarchyAndRequiredPowers(t *testing.T) {
	events := publicE32TestEvents()
	require.NoError(t, validatePublicE32Events(events))

	missingParent := append([]Event(nil), events...)
	missingParent[len(missingParent)-1].ParentSequence = 999
	require.ErrorContains(t, validatePublicE32Events(missingParent), "missing parent")

	missingPower := append([]Event(nil), events...)
	for index := range missingPower {
		if missingPower[index].Scope == "power" && missingPower[index].Power != nil && *missingPower[index].Power == 16 {
			missingPower = append(missingPower[:index], missingPower[index+1:]...)
			break
		}
	}
	require.ErrorContains(t, validatePublicE32Events(missingPower), "missing generated Chebyshev power T16")

	duplicateSequence := append([]Event(nil), events...)
	duplicateSequence[len(duplicateSequence)-1].Sequence = duplicateSequence[0].Sequence
	require.ErrorContains(t, validatePublicE32Events(duplicateSequence), "duplicate fastdiag event sequence")
}

func TestValidatePublicE32EventsRecursivePowerTree(t *testing.T) {
	// Real Fast recursion starts T16 before generating its child T8.
	events := publicE32TestEvents()
	var t16Sequence, t8Sequence uint64
	for i := range events {
		if events[i].Scope == "power" && events[i].Name == "power" && events[i].Power != nil {
			switch *events[i].Power {
			case 16:
				if t16Sequence == 0 {
					t16Sequence = events[i].Sequence
				}
			case 8:
				if t8Sequence == 0 {
					t8Sequence = events[i].Sequence
				}
			}
		}
	}
	require.NotZero(t, t16Sequence)
	require.NotZero(t, t8Sequence)
	var t8Parent uint64
	for _, event := range events {
		if event.Sequence == t8Sequence {
			t8Parent = event.ParentSequence
		}
	}
	require.Equal(t, t16Sequence, t8Parent)
	require.NoError(t, validatePublicE32Events(events))

	wrongScope := append([]Event(nil), events...)
	for i := range wrongScope {
		if wrongScope[i].Sequence == t8Sequence {
			wrongScope[i].ParentSequence = 5 // unrelated Bootstrap stage
		}
	}
	require.ErrorContains(t, validatePublicE32Events(wrongScope), "non-power ancestor")

	cycle := append([]Event(nil), events...)
	for i := range cycle {
		if cycle[i].Scope == "power" && cycle[i].Name == "power" && cycle[i].Power != nil && *cycle[i].Power == 16 && cycle[i].Sequence == t16Sequence {
			cycle[i].ParentSequence = t8Sequence // T16 -> T8 -> T16
		}
	}
	require.ErrorContains(t, validatePublicE32Events(cycle), "cyclic ancestry")
}

func TestValidatePublicE32EventsRejectsDuplicatePowerWithinRoot(t *testing.T) {
	events := publicE32TestEvents()
	var generatedRoot Event
	var duplicate Event
	var maxSequence uint64
	for _, event := range events {
		if event.Sequence > maxSequence {
			maxSequence = event.Sequence
		}
		if event.Scope == "power" && event.Name == "generated_powers" && generatedRoot.Sequence == 0 {
			generatedRoot = event
		}
		if event.Scope == "power" && event.Name == "power" && event.Power != nil && *event.Power == 2 && duplicate.Sequence == 0 {
			duplicate = event
		}
	}
	require.NotZero(t, generatedRoot.Sequence)
	require.NotZero(t, duplicate.Sequence)
	duplicate.Sequence = maxSequence + 1
	duplicate.ParentSequence = generatedRoot.Sequence
	events = append(events, duplicate)

	require.ErrorContains(t, validatePublicE32Events(events), "duplicate generated power T2")
}

func TestValidatePublicE32RawTraceRequiresExactMatchedTwoCalls(t *testing.T) {
	manifest := publicE32FixtureManifest{
		SchemaVersion: "fast-public-e32-trace-fixture.v1", Profile: "logn13-e32-public-native", Backend: "fast",
		BackendCommit: e32ProductionFastCommit, PrimaryCommit: "primary", PrimarySourceSHA256: "source",
		ConfigSHA256: "config", QPSHA256: "qp", InputSHA256: "input", WorkloadSHA256: "workload",
		CiphertextSHA256: "ciphertext", DecodedSHA256: "decoded", State: publicE32CipherState{Scale: "scale"},
	}
	vectors := publicE32Vectors{BootstrapOutputs: map[string][]publicE32Vector{
		"bootstrap_first_cold": make([]publicE32Vector, 1<<12),
		"bootstrap_warm_01":    make([]publicE32Vector, 1<<12),
		"bootstrap_warm_02":    make([]publicE32Vector, 1<<12),
		"bootstrap_warm_03":    make([]publicE32Vector, 1<<12),
		"bootstrap_warm_04":    make([]publicE32Vector, 1<<12),
		"bootstrap_warm_05":    make([]publicE32Vector, 1<<12),
	}}
	raw := publicE32RawTrace{
		SchemaVersion: "fastdiag.public-e32.trace.v1", TraceStatus: "TRACE_VALIDATED", Profile: "logn13-e32-public-native", Mode: "test-only-public-e32-fast",
		PrimaryCommit: "current-primary", FixturePrimaryCommit: "primary", PrimarySourceSHA256: "source", ProductionFastCommit: e32ProductionFastCommit,
		RawEvidenceFile: "raw-trace-unverified.json", RawEvidenceSHA256: strings.Repeat("c", 64),
		DiagnosticHead: "diagnostic", ConfigSHA256: "config", QPSHA256: "qp", InputSHA256: "input", WorkloadSHA256: "workload",
		FixtureManifestSHA256: "manifest", CiphertextSHA256: "ciphertext", ExpectedInputSHA256: "decoded",
		CallBudget: 2, ActualCalls: 2, CPUProfileFile: "warm-cpu.pprof", CPUProfileSHA256: strings.Repeat("a", 64),
		HeapProfileFile: "post-warm-heap.pprof", HeapProfileSHA256: strings.Repeat("b", 64),
		NumericalGate: 1e-6, GoVersion: "go1.26.4", OS: "darwin", Arch: "arm64", NumCPU: 10, GOMAXPROCS: 10,
		GOGC: "runtime-default", GOMEMLIMIT: "runtime-default", GODEBUG: "runtime-default",
		Runs: []publicE32Run{
			{Index: 1, Phase: "first_cold_bootstrap", DecodedSHA256: "cold", OracleMaxComplex: 0, FastReferenceWorstMax: 0, Level: 1, Degree: 1, QPrefixRows: 2, Scale: "scale"},
			{Index: 2, Phase: "warm_bootstrap_01", DecodedSHA256: "warm", OracleMaxComplex: 0, FastReferenceWorstMax: 0, Level: 1, Degree: 1, QPrefixRows: 2, Scale: "scale"},
		},
		Events: publicE32TestEvents(),
	}
	require.NoError(t, validatePublicE32RawTrace(raw, manifest, vectors, "current-primary", "primary", "diagnostic", "manifest"))
	raw.ActualCalls = 3
	require.ErrorContains(t, validatePublicE32RawTrace(raw, manifest, vectors, "current-primary", "primary", "diagnostic", "manifest"), "exact 2-call budget")
}

func TestValidatePublicE32RawCertificationBindsUnverifiedEvents(t *testing.T) {
	unverified := publicE32RawTrace{
		TraceStatus: "TRACE_UNVERIFIED", TraceValidationError: "raw event capture is not certified",
		PrimaryCommit: "current-primary", FixturePrimaryCommit: "fixture-primary", ProductionFastCommit: e32ProductionFastCommit,
		DiagnosticHead: "diagnostic", CallBudget: 2, ActualCalls: 2,
		Runs:   []publicE32Run{{Index: 1, Phase: "cold"}, {Index: 2, Phase: "warm"}},
		Events: publicE32TestEvents(),
	}
	rawBytes, err := json.Marshal(unverified)
	require.NoError(t, err)
	certified := unverified
	certified.TraceStatus = "TRACE_VALIDATED"
	certified.TraceValidationError = ""
	certified.RawEvidenceFile = "raw-trace-unverified.json"
	rawSHA := sha256.Sum256(rawBytes)
	certified.RawEvidenceSHA256 = hex.EncodeToString(rawSHA[:])
	require.NoError(t, validatePublicE32RawCertification(unverified, certified, rawBytes))

	certified.Events = append([]Event(nil), certified.Events...)
	certified.Events[0].ElapsedNS++
	require.ErrorContains(t, validatePublicE32RawCertification(unverified, certified, rawBytes), "changes captured raw runs or event records")
}

func TestValidateE32InstrumentationOnlyPatchRejectsArithmeticChanges(t *testing.T) {
	patch := `diff --git a/rescale.go b/rescale.go
@@ -1,0 +2,4 @@
+var materializationSpan fastdiag.Span
+if fastdiag.Enabled && fastdiag.Selected(fastdiag.Rescale) {
+materializationSpan = fastdiag.Begin(fastdiag.Rescale, "materialization", wholeSpan.Sequence(), fastdiag.Input(op0, sourceRows).Merge(fastdiag.InPlace(op0 == opOut)).Merge(fastdiag.RepetitionCount(len(op0.Value))))
+}
@@ -5,0 +10,4 @@
+var restoreSpan fastdiag.Span
+if fastdiag.Enabled && fastdiag.Selected(fastdiag.Rescale) {
+restoreSpan = fastdiag.Begin(fastdiag.Rescale, "ntt_montgomery_restore", materializationSpan.Sequence(), rescaleDiagFields(op0.Level(), targetLevel, sourceRows, targetRows, fmt.Sprintf("c%d", component), op0 == opOut))
+}
@@ -12,0 +18,3 @@
+if fastdiag.Enabled && fastdiag.Selected(fastdiag.Rescale) {
+restoreSpan.End(fastdiag.Fields{})
+}
@@ -20,0 +29 @@
+materializationSpan.End(fastdiag.Output(opOut, targetRows))
`
	require.NoError(t, validateE32InstrumentationOnlyPatch(patch))
	require.ErrorContains(t, validateE32InstrumentationOnlyPatch(patch+"+opOut.Scale = op0.Scale\n"), "non-instrumentation")
	require.ErrorContains(t, validateE32InstrumentationOnlyPatch(patch+"-existingArithmetic()\n"), "removes existing source")
}

func TestVerifyE32ProductionSourceIdentityAgainstDiagnosticCommit(t *testing.T) {
	secondaryRoot := os.Getenv("FASTDIAG_SECONDARY_ROOT")
	if secondaryRoot == "" {
		t.Skip("set FASTDIAG_SECONDARY_ROOT to verify the committed Batch024 diagnostic source delta")
	}
	diagnosticHead, err := gitOutput(secondaryRoot, "rev-parse", "HEAD")
	require.NoError(t, err)
	productionHash, diagnosticHash, paths, instrumentationOnly, err := verifyE32ProductionSourceIdentity(secondaryRoot, diagnosticHead)
	require.NoError(t, err)
	require.NotEqual(t, productionHash, diagnosticHash, "diagnostic source digest must reflect the separate trace-only spans")
	require.Equal(t, []string{e32AuthorizedInstrumentationPath}, paths)
	require.True(t, instrumentationOnly)
}

func TestAnalyzePublicE32EventsPreservesMeasuredClosureAndBounds(t *testing.T) {
	levelOut := 7
	events := []Event{
		{Scope: "stage", Name: "bootstrap", Sequence: 1, ElapsedNS: 1000},
		{Scope: "stage", Name: "coeffs_to_slots", Sequence: 2, ParentSequence: 1, ElapsedNS: 700},
		{Scope: "power", Name: "generated_powers", Sequence: 3, ParentSequence: 2, ElapsedNS: 750},
		{Scope: "stage", Name: "evalmod_real", Sequence: 4, ParentSequence: 1, ElapsedNS: 400},
		{Scope: "rescale", Name: "rescale", Sequence: 5, ElapsedNS: 200},
		{Scope: "rescale", Name: "preflight", Sequence: 6, ParentSequence: 5, ElapsedNS: 80},
		{Scope: "rescale", Name: "materialization", Sequence: 7, ParentSequence: 5, ElapsedNS: 100},
		{Scope: "power", Name: "generated_powers", Sequence: 8, ElapsedNS: 500},
		{Scope: "power", Name: "generated_powers", Sequence: 9, ParentSequence: 2, ElapsedNS: 25, LevelOut: &levelOut},
	}
	analysis, err := analyzePublicE32Events(events)
	require.NoError(t, err)
	require.Len(t, analysis.Closures, len(events), "legacy per-event closures remain available")
	require.Equal(t, float64(-100), analysis.RootUnattributedNS, "root closure must preserve negative residuals")
	require.Equal(t, float64(-0.1), analysis.RootUnattributedFraction)
	var rescaleClosure, rescaleChildClosure, parentClosure publicE32EventClosure
	for _, closure := range analysis.Closures {
		if closure.Sequence == 5 {
			rescaleClosure = closure
		}
		if closure.Sequence == 7 {
			rescaleChildClosure = closure
		}
		if closure.Sequence == 2 {
			parentClosure = closure
		}
	}
	require.Equal(t, float64(20), rescaleClosure.UnattributedNS)
	require.Equal(t, float64(1), rescaleClosure.InclusiveRootShare,
		"per-event inclusive shares are relative to their own independent root")
	require.InDelta(t, 0.5, rescaleChildClosure.InclusiveRootShare, 1e-12,
		"nested Rescale shares use the Rescale root, not the Bootstrap root, and are not counted twice")
	require.Equal(t, float64(-75), parentClosure.UnattributedNS, "overlapping child timing must remain visible")
	require.Len(t, analysis.PositiveExclusivePareto, 3, "independent Bootstrap, Rescale and Power roots remain separate")
	var rescaleRootGroup *publicE32ParetoGroup
	for i := range analysis.PositiveExclusivePareto {
		if analysis.PositiveExclusivePareto[i].RootSequence == 5 {
			rescaleRootGroup = &analysis.PositiveExclusivePareto[i]
		}
	}
	require.NotNil(t, rescaleRootGroup)
	require.Equal(t, "rescale", rescaleRootGroup.RootScope)
	require.Len(t, rescaleRootGroup.Entries, 3)
	var rescaleRootEntry *publicE32ParetoEntry
	for i := range rescaleRootGroup.Entries {
		if rescaleRootGroup.Entries[i].Sequence == 5 {
			rescaleRootEntry = &rescaleRootGroup.Entries[i]
		}
	}
	require.NotNil(t, rescaleRootEntry)
	require.NotNil(t, rescaleRootEntry.RootShare)
	require.InDelta(t, 0.1, *rescaleRootEntry.RootShare, 1e-12, "Rescale share is relative only to its own root, not Bootstrap")
	rootExclusiveTotals := make(map[uint64]float64)
	powerRootSequences := make(map[uint64]bool)
	var rescaleTypeSummary *publicE32EventTypeSummary
	var repeatedPowerTypeSummary *publicE32EventTypeSummary
	for i := range analysis.EventTypeSummaries {
		summary := &analysis.EventTypeSummaries[i]
		rootExclusiveTotals[summary.TreeRootSequence] += summary.ExclusiveTotalNS
		if summary.TreeRootSequence == 5 && summary.Scope == "rescale" && summary.Name == "rescale" {
			rescaleTypeSummary = summary
		}
		if summary.Scope == "power" && summary.Name == "generated_powers" {
			powerRootSequences[summary.TreeRootSequence] = true
			if summary.TreeRootSequence == 1 && summary.ParentType == "stage/coeffs_to_slots" {
				repeatedPowerTypeSummary = summary
			}
		}
	}
	require.NotNil(t, rescaleTypeSummary)
	require.Equal(t, uint64(5), rescaleTypeSummary.TreeRootSequence)
	require.Equal(t, 1, rescaleTypeSummary.Count)
	require.NotNil(t, rescaleTypeSummary.ExclusiveRootShare)
	require.InDelta(t, 0.1, *rescaleTypeSummary.ExclusiveRootShare, 1e-12,
		"type summary shares are local to their own independent event-tree root")
	require.Equal(t, map[uint64]bool{1: true, 8: true}, powerRootSequences,
		"same event type under separate roots must remain separate summaries")
	require.NotNil(t, repeatedPowerTypeSummary)
	require.Equal(t, 2, repeatedPowerTypeSummary.Count,
		"event type aggregation ignores per-occurrence metadata while preserving count")
	require.Equal(t, float64(775), repeatedPowerTypeSummary.ExclusiveTotalNS)
	require.Equal(t, float64(1000), rootExclusiveTotals[1],
		"exclusive parent-minus-child accounting must not double count nested inclusive time")
	require.Equal(t, float64(200), rootExclusiveTotals[5])
	require.Equal(t, float64(500), rootExclusiveTotals[8])
	require.Len(t, analysis.PositiveExclusiveByType, 3,
		"type Pareto groups must remain partitioned by independent event root")
	var rescaleTypePareto *publicE32EventTypeParetoGroup
	for i := range analysis.PositiveExclusiveByType {
		if analysis.PositiveExclusiveByType[i].RootSequence == 5 {
			rescaleTypePareto = &analysis.PositiveExclusiveByType[i]
		}
	}
	require.NotNil(t, rescaleTypePareto)
	var rescaleTypeEntry *publicE32EventTypeParetoEntry
	for i := range rescaleTypePareto.Entries {
		if rescaleTypePareto.Entries[i].Scope == "rescale" && rescaleTypePareto.Entries[i].Name == "rescale" {
			rescaleTypeEntry = &rescaleTypePareto.Entries[i]
		}
	}
	require.NotNil(t, rescaleTypeEntry)
	require.Equal(t, float64(20), rescaleTypeEntry.ExclusiveTotalNS)
	require.NotNil(t, rescaleTypeEntry.RootShare)
	require.InDelta(t, 0.1, *rescaleTypeEntry.RootShare, 1e-12,
		"type Pareto shares are relative only to their own event-tree root")
	var evalmod publicE32AmdahlScenario
	for _, scenario := range analysis.AmdahlScenarios {
		if scenario.Stage == "evalmod_real" {
			evalmod = scenario
		}
	}
	require.Equal(t, 0.4, evalmod.ObservedRootFraction)
	require.NotNil(t, evalmod.TwoXStageOverallSpeedup)
	require.InDelta(t, 1.25, *evalmod.TwoXStageOverallSpeedup, 1e-12)
}

func publicE32TestEvents() []Event {
	var events []Event
	stageNames := []string{"bootstrap", "pack_n1_to_n2", "scale_down", "mod_up_trace", "coeffs_to_slots", "evalmod_real", "evalmod_imag", "slots_to_coeffs", "unpack_n2_to_n1", "public_finalization"}
	for index, name := range stageNames {
		sequence, parent := uint64(index+1), uint64(0)
		if name != "bootstrap" {
			parent = 1
		}
		events = append(events, Event{Scope: "stage", Name: name, Sequence: sequence, ParentSequence: parent})
	}
	sequence := uint64(len(events) + 1)
	for range 2 {
		rootSequence := sequence
		sequence++
		events = append(events, Event{Scope: "power", Name: "generated_powers", Sequence: rootSequence})
		generated := make(map[int]bool)
		var addPower func(int, uint64)
		addPower = func(power int, parent uint64) {
			if power <= 1 || generated[power] {
				return
			}
			generated[power] = true
			a, b := commonpolynomial.SplitDegree(power)
			powerSequence := sequence
			sequence++
			p, splitA, splitB := power, a, b
			events = append(events, Event{Scope: "power", Name: "power", Sequence: powerSequence, ParentSequence: parent, Power: &p, SplitA: &splitA, SplitB: &splitB})
			addPower(a, powerSequence)
			addPower(b, powerSequence)
			for _, phase := range []string{"copy_workspace", "mul_relin", "chebyshev_doubling", "recurrence_correction", "rescale"} {
				events = append(events, Event{Scope: "power", Name: phase, Sequence: sequence, ParentSequence: powerSequence})
				sequence++
			}
		}
		addPower(16, rootSequence)
		addPower(6, rootSequence)
	}
	levelIn, levelOut, rowsIn, rowsOut := 4, 3, 4, 4
	makeRescaleEvent := func(name, component string, parent uint64, elapsed int64) Event {
		levelInCopy, levelOutCopy, rowsInCopy, rowsOutCopy := levelIn, levelOut, rowsIn, rowsOut
		return Event{Scope: "rescale", Name: name, Component: component, Sequence: sequence, ParentSequence: parent, ElapsedNS: elapsed,
			LevelIn: &levelInCopy, LevelOut: &levelOutCopy, RowsIn: &rowsInCopy, RowsOut: &rowsOutCopy}
	}
	rescaleRoot := sequence
	events = append(events, makeRescaleEvent("rescale", "", 0, 100))
	sequence++
	preflight := sequence
	events = append(events, makeRescaleEvent("preflight", "", rescaleRoot, 80))
	sequence++
	for _, component := range []string{"c0", "c1"} {
		events = append(events, makeRescaleEvent("prefix_to_coefficient", component, preflight, 10))
		sequence++
	}
	loopSequences := make(map[string]uint64)
	for _, component := range []string{"c0", "c1"} {
		loopSequences[component] = sequence
		events = append(events, makeRescaleEvent("coefficient_loop", component, preflight, 40))
		sequence++
	}
	for _, component := range []string{"c0", "c1"} {
		for _, childName := range []string{"reconstruct_center_round_capacity", "residue_materialization"} {
			events = append(events, makeRescaleEvent(childName, component, loopSequences[component], 15))
			sequence++
		}
	}
	materialization := sequence
	events = append(events, makeRescaleEvent("materialization", "", rescaleRoot, 20))
	sequence++
	for _, component := range []string{"c0", "c1"} {
		events = append(events, makeRescaleEvent("ntt_montgomery_restore", component, materialization, 10))
		sequence++
	}
	return events
}
