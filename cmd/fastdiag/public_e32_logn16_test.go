package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestValidateLogN16E32ManifestFailsClosed(t *testing.T) {
	base := validLogN16FixtureManifest()
	require.NoError(t, validateLogN16E32Manifest(base, base.PrimaryCommit, base.PrimarySourceSHA256))
	tests := []struct {
		name   string
		mutate func(*publicE32FixtureManifest)
	}{
		{"historical schema", func(m *publicE32FixtureManifest) { m.SchemaVersion = "fast-public-e32-trace-fixture.v1" }},
		{"historical profile", func(m *publicE32FixtureManifest) { m.Profile = "logn13-e32-public-native" }},
		{"wrong Fast pin", func(m *publicE32FixtureManifest) { m.BackendCommit = e32ProductionFastCommit }},
		{"wrong Primary source", func(m *publicE32FixtureManifest) { m.PrimarySourceSHA256 = strings.Repeat("c", 64) }},
		{"nonzero budget", func(m *publicE32FixtureManifest) { m.BootstrapBudget = 1 }},
		{"nonzero calls", func(m *publicE32FixtureManifest) { m.BootstrapCalls = 1 }},
		{"wrong dimensions", func(m *publicE32FixtureManifest) { m.PhysicalRowLengths[1][0]-- }},
		{"prefix q does not match effective q0", func(m *publicE32FixtureManifest) { m.State.PrefixQ = "36028797019488259" }},
		{"non-NTT input", func(m *publicE32FixtureManifest) { m.IsNTT = false }},
		{"Montgomery input", func(m *publicE32FixtureManifest) { m.IsMontgomery = true }},
		{"nonzero c1", func(m *publicE32FixtureManifest) { m.C1Nonzero = true }},
		{"wrong effective profile", func(m *publicE32FixtureManifest) { m.Parameters.LogN = 13 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := base
			manifest.PhysicalRowLengths = [][]int{{1 << 16}, {1 << 16}}
			test.mutate(&manifest)
			require.Error(t, validateLogN16E32Manifest(manifest, base.PrimaryCommit, base.PrimarySourceSHA256))
		})
	}
}

func TestValidatePublicE32LogN16FixtureBindsZeroCallVectorsAndHashes(t *testing.T) {
	root := t.TempDir()
	parameters, ciphertext := []byte("parameters"), []byte("held native level0 ciphertext")
	require.NoError(t, os.WriteFile(filepath.Join(root, "parameters.bin"), parameters, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ciphertext.bin"), ciphertext, 0o600))
	parametersSHA, ciphertextSHA := sha256.Sum256(parameters), sha256.Sum256(ciphertext)
	decoded := make([]complex128, 1<<15)
	input := make([]publicE32Vector, len(decoded))
	decodedSHA := perfmeasure.Fingerprint(decoded)
	manifest := validLogN16FixtureManifest()
	manifest.DecodedSHA256 = decodedSHA
	manifest.ParametersFile, manifest.ParametersSHA256 = "parameters.bin", hex.EncodeToString(parametersSHA[:])
	manifest.CiphertextFile, manifest.CiphertextSHA256 = "ciphertext.bin", hex.EncodeToString(ciphertextSHA[:])
	manifest.CiphertextFingerprintSHA256 = strings.Repeat("d", 64)
	manifest.State.PrefixQ = "36028797019488257"
	vectors := publicE32Vectors{
		SchemaVersion: "fast-standard-public-native-vectors.v1", Mode: "public-native", Backend: "fast",
		BackendCommit: e32LogN16ProductionFastCommit, PrimaryCommit: manifest.PrimaryCommit, SourceSHA256: manifest.PrimarySourceSHA256,
		ConfigSHA256: logN16ConfigSHA, QPSHA256: logN16QPSHA, InputSHA256: logN16InputSHA,
		WorkloadSHA256: logN16WorkloadSHA, EphemeralSecretWeight: 32,
		Checkpoints: map[string][]publicE32Vector{},
	}
	for _, name := range logN16PublicCheckpointNames {
		vectors.Checkpoints[name] = input
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	vectorsBytes, err := json.Marshal(vectors)
	require.NoError(t, err)
	manifestPath, vectorsPath := filepath.Join(root, "manifest.json"), filepath.Join(root, "vectors.json")
	require.NoError(t, os.WriteFile(manifestPath, manifestBytes, 0o600))
	require.NoError(t, os.WriteFile(vectorsPath, vectorsBytes, 0o600))

	gotManifestBytes, gotManifest, gotVectorsBytes, gotVectors, err := validatePublicE32LogN16Fixture(manifestPath, vectorsPath, manifest.PrimaryCommit, manifest.PrimarySourceSHA256)
	require.NoError(t, err)
	require.Equal(t, manifestBytes, gotManifestBytes)
	require.Equal(t, manifest.CiphertextSHA256, gotManifest.CiphertextSHA256)
	require.Equal(t, vectorsBytes, gotVectorsBytes)
	require.Len(t, gotVectors.Checkpoints, 8)
	for _, omittedField := range []string{"bootstrap_budget", "actual_bootstrap_calls"} {
		var fields map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(manifestBytes, &fields))
		delete(fields, omittedField)
		malformed, err := json.Marshal(fields)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(manifestPath, malformed, 0o600))
		_, _, _, _, err = validatePublicE32LogN16Fixture(manifestPath, vectorsPath, manifest.PrimaryCommit, manifest.PrimarySourceSHA256)
		require.ErrorContains(t, err, "must explicitly record")
	}
	require.NoError(t, os.WriteFile(manifestPath, manifestBytes, 0o600))

	vectors.SchemaVersion = "fast-standard-public-native-repeatability-vectors.v1"
	vectorsBytes, err = json.Marshal(vectors)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(vectorsPath, vectorsBytes, 0o600))
	_, _, _, _, err = validatePublicE32LogN16Fixture(manifestPath, vectorsPath, manifest.PrimaryCommit, manifest.PrimarySourceSHA256)
	require.ErrorContains(t, err, "zero-Bootstrap current-Fast")
}

func TestVerifyCurrentFastTestOnlyCloneAgainstIsolatedAdapter(t *testing.T) {
	root := os.Getenv("FASTDIAG_LOGN16_DIAGNOSTIC_ROOT")
	if root == "" {
		t.Skip("set FASTDIAG_LOGN16_DIAGNOSTIC_ROOT to audit an isolated test-only adapter commit")
	}
	head, err := gitOutput(root, "rev-parse", "HEAD")
	require.NoError(t, err)
	productionHash, diagnosticHash, paths, err := verifyCurrentFastTestOnlyClone(root, head)
	require.NoError(t, err)
	require.Equal(t, productionHash, diagnosticHash)
	require.Equal(t, []string{logN16E32TestAdapterPath}, paths)
}

func TestTask006AttemptLedgerBindsReservationsAndTask005Exhaustion(t *testing.T) {
	root := t.TempDir()
	outputJSON := filepath.Join(root, "trace.json")
	baseline := &publicE32Task005Baseline{ReservedCalls: 16, ActualCalls: 16, AttemptLedgerSHA256: strings.Repeat("a", 64)}
	journalBase := outputJSON + ".diagnostic-fast"
	require.NoError(t, reservePublicE32TraceAttempts(journalBase))
	ledgerPath, ledgerSHA, err := writePublicE32Task006AttemptLedger(outputJSON, 2, 2, true, "FAST_LOGN16_STAGE_TRACE_VERIFIED", baseline)
	require.NoError(t, err)
	require.Equal(t, publicE32Task006AttemptLedgerPath(outputJSON), ledgerPath)
	require.Len(t, ledgerSHA, 64)
	data, err := os.ReadFile(ledgerPath)
	require.NoError(t, err)
	var ledger publicE32Task006AttemptLedger
	require.NoError(t, json.Unmarshal(data, &ledger))
	require.Equal(t, 16, ledger.Task005ActualCalls)
	require.Equal(t, 2, ledger.Task006ReservedCalls)
	require.Equal(t, 2, ledger.Task006ActualCalls)
	require.True(t, ledger.Task006ActualCallsObserved)
	require.Len(t, ledger.Reservations, 2)
	require.WithinDuration(t, time.Now(), ledger.Reservations[0].ReservedAt, time.Minute)
	require.Equal(t, "FAST_LOGN16_STAGE_TRACE_VERIFIED", ledger.CompletionStatus)
}

func validLogN16FixtureManifest() publicE32FixtureManifest {
	return publicE32FixtureManifest{
		SchemaVersion: "fast-public-e32-trace-fixture.v2", Profile: logN16PublicE32Profile, Backend: "fast",
		BackendCommit: e32LogN16ProductionFastCommit, PrimaryCommit: strings.Repeat("a", 40), PrimarySourceSHA256: strings.Repeat("b", 64),
		ConfigSHA256: logN16ConfigSHA, QPSHA256: logN16QPSHA, InputSHA256: logN16InputSHA,
		WorkloadSHA256: logN16WorkloadSHA, ParametersFile: "parameters.bin", CiphertextFile: "ciphertext.bin",
		ParametersSHA256: strings.Repeat("e", 64), CiphertextSHA256: strings.Repeat("f", 64), DecodedSHA256: strings.Repeat("1", 64),
		CiphertextFingerprintSHA256: strings.Repeat("2", 64), State: publicE32CipherState{
			Level: 0, Degree: 1, Scale: "3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13",
			QPrefixRows: 1, PrefixQ: "36028797019488257",
		}, PhysicalRowLengths: [][]int{{1 << 16}, {1 << 16}}, IsNTT: true,
		Mode: logN16PublicE32Mode, InputKind: logN16FastInputKind, InputConstructor: logN16FastConstructor, PublicDispatch: logN16FastDispatch,
		NumericalGate: 1e-6, EphemeralSecretWeight: 32,
		Parameters: perfmeasure.EffectiveParameters{
			EphemeralSecretWeight: 32, LogN: 16, LogSlots: 15, InputSlots: 1 << 15, RingN: 1 << 16,
			DefaultScale: "3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13",
			Mod1Scale:    60, Mod1Degree: 30, DoubleAngle: 3, K: 16, CircuitOrder: 0, QPrefixRowsAtMax: 4,
			QPrimes: []string{"36028797019488257"},
		},
	}
}
