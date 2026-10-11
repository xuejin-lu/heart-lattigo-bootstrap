package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	logN16PublicE32Profile = "logn16-e32-public-native"
	logN16PublicE32Mode    = "public-native"
	logN16FastInputKind    = "fast_zero_secret_public_encrypt_v1"
	logN16FastConstructor  = "rlwe.NewEncryptor(full, extended bootstrap secret).EncryptNew; pinned Fast CKKS zero-secret capability yields observed c1=0"
	logN16FastDispatch     = "bootstrapping.NewEvaluator public wrapper; Fast dispatch selected by pinned zero-secret capability"
)

var logN16PublicCheckpointNames = []string{
	"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_level5",
	"mulrelin_level5", "rescale_q5_level4", "rotate_level4", "drop_level0",
}

// validatePublicE32LogN16Fixture is the explicit v2 path. The historical
// LogN13 validator remains frozen so a new profile cannot silently inherit its
// backend pin, dimensions, or vector semantics.
func validatePublicE32LogN16Fixture(
	manifestPath, vectorsPath, fixturePrimaryCommit, measurementSourceHash string,
) ([]byte, publicE32FixtureManifest, []byte, publicE32Vectors, error) {
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("read LogN16 E32 fixture manifest: %w", err)
	}
	var manifest publicE32FixtureManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("decode LogN16 E32 fixture manifest: %w", err)
	}
	if err := validateLogN16E32Manifest(manifest, fixturePrimaryCommit, measurementSourceHash); err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, err
	}
	fixtureDir := filepath.Dir(manifestPath)
	for _, item := range []struct{ name, digest string }{
		{manifest.ParametersFile, manifest.ParametersSHA256},
		{manifest.CiphertextFile, manifest.CiphertextSHA256},
	} {
		if filepath.Base(item.name) != item.name || item.name == "" || item.name == "." || len(item.digest) != sha256.Size*2 {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("LogN16 fixture component path or SHA is invalid: %q", item.name)
		}
		data, err := os.ReadFile(filepath.Join(fixtureDir, item.name))
		if err != nil {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("read LogN16 fixture component %s: %w", item.name, err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != item.digest {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("LogN16 fixture component %s SHA-256 mismatch", item.name)
		}
	}
	vectorsBytes, err := os.ReadFile(vectorsPath)
	if err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("read LogN16 public-native vectors: %w", err)
	}
	var vectors publicE32Vectors
	if err := json.Unmarshal(vectorsBytes, &vectors); err != nil {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("decode LogN16 public-native vectors: %w", err)
	}
	if vectors.SchemaVersion != "fast-standard-public-native-vectors.v1" || vectors.Mode != "public-native" ||
		vectors.Backend != "fast" || vectors.BackendCommit != e32LogN16ProductionFastCommit ||
		vectors.PrimaryCommit != fixturePrimaryCommit || vectors.SourceSHA256 != measurementSourceHash ||
		vectors.ConfigSHA256 != manifest.ConfigSHA256 || vectors.QPSHA256 != manifest.QPSHA256 ||
		vectors.InputSHA256 != manifest.InputSHA256 || vectors.WorkloadSHA256 != manifest.WorkloadSHA256 ||
		vectors.EphemeralSecretWeight != 32 || len(vectors.BootstrapOutputs) != 0 || len(vectors.Checkpoints) != len(logN16PublicCheckpointNames) {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, errors.New("LogN16 vectors do not match the zero-Bootstrap current-Fast public-native fixture contract")
	}
	for _, name := range logN16PublicCheckpointNames {
		values, ok := vectors.Checkpoints[name]
		if !ok || len(values) != 1<<15 || !finiteE32Vectors(values) {
			return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, fmt.Errorf("LogN16 preflight checkpoint %s is absent, malformed, or non-finite", name)
		}
	}
	drop := vectors.Checkpoints["drop_level0"]
	if fingerprintE32Vectors(drop) != manifest.DecodedSHA256 {
		return nil, publicE32FixtureManifest{}, nil, publicE32Vectors{}, errors.New("LogN16 fixture decoded hash does not match its held drop_level0 vector")
	}
	return manifestBytes, manifest, vectorsBytes, vectors, nil
}

func validateLogN16E32Manifest(manifest publicE32FixtureManifest, fixturePrimaryCommit, measurementSourceHash string) error {
	wantScale := "3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13"
	if manifest.SchemaVersion != "fast-public-e32-trace-fixture.v2" || manifest.Profile != logN16PublicE32Profile ||
		manifest.Backend != "fast" || manifest.BackendCommit != e32LogN16ProductionFastCommit ||
		manifest.PrimaryCommit != fixturePrimaryCommit || manifest.PrimarySourceSHA256 != measurementSourceHash ||
		manifest.ConfigSHA256 != logN16ConfigSHA || manifest.QPSHA256 != logN16QPSHA || manifest.InputSHA256 != logN16InputSHA ||
		manifest.WorkloadSHA256 != logN16WorkloadSHA || manifest.Mode != logN16PublicE32Mode ||
		manifest.InputKind != logN16FastInputKind || manifest.InputConstructor != logN16FastConstructor || manifest.PublicDispatch != logN16FastDispatch ||
		manifest.BootstrapBudget != 0 || manifest.BootstrapCalls != 0 || manifest.NumericalGate != 1e-6 ||
		manifest.EphemeralSecretWeight != 32 || manifest.State.Level != 0 || manifest.State.Degree != 1 ||
		manifest.State.Scale != wantScale || manifest.State.QPrefixRows != 1 || manifest.State.PrefixQ == "" ||
		!logN16PrefixMatchesEffectiveQ(manifest.State.PrefixQ, manifest.Parameters.QPrimes) ||
		!validSHA256Hex(manifest.DecodedSHA256) || !validSHA256Hex(manifest.CiphertextFingerprintSHA256) ||
		!validSHA256Hex(manifest.ParametersSHA256) || !validSHA256Hex(manifest.CiphertextSHA256) ||
		!validSHA256Hex(manifest.PrimarySourceSHA256) || len(fixturePrimaryCommit) != 40 ||
		!manifest.IsNTT || manifest.IsMontgomery || manifest.C1Nonzero || len(manifest.PhysicalRowLengths) != 2 ||
		!validLogN16PhysicalRows(manifest.PhysicalRowLengths) || !validLogN16EffectiveParameters(manifest.Parameters) {
		return errors.New("fixture manifest fails the frozen current-Fast LogN16/E32 native Level0 v2 contract")
	}
	if fixturePrimaryCommit == "" || measurementSourceHash == "" {
		return errors.New("LogN16 fixture validation requires pinned Primary commit and measurement-source fingerprint")
	}
	return nil
}

func logN16PrefixMatchesEffectiveQ(prefixQ string, qPrimes []string) bool {
	return len(qPrimes) > 0 && prefixQ == qPrimes[0]
}

func validSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validLogN16PhysicalRows(rows [][]int) bool {
	if len(rows) != 2 {
		return false
	}
	for _, component := range rows {
		if len(component) != 1 || component[0] != 1<<16 {
			return false
		}
	}
	return true
}

func validLogN16EffectiveParameters(p perfmeasure.EffectiveParameters) bool {
	return p.EphemeralSecretWeight == 32 && p.LogN == 16 && p.LogSlots == 15 && p.InputSlots == 1<<15 && p.RingN == 1<<16 &&
		p.DefaultScale == "3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13" &&
		p.Mod1Scale == 60 && p.Mod1Degree == 30 && p.DoubleAngle == 3 && p.K == 16 && p.CircuitOrder == 0 && p.QPrefixRowsAtMax == 4 &&
		len(p.QPrimes) > 0 && p.QPrimes[0] == "36028797019488257"
}

func validatePublicE32LogN16RawTrace(raw publicE32RawTrace, manifest publicE32FixtureManifest, vectors publicE32Vectors,
	primaryCommit, fixturePrimaryCommit, diagnosticHead, manifestHash string, reference task005ReferenceInfo) error {
	if raw.SchemaVersion != "fastdiag.public-e32.trace.v1" || raw.TraceStatus != "TRACE_VALIDATED" || raw.TraceValidationError != "" ||
		raw.RawEvidenceFile == "" || filepath.Base(raw.RawEvidenceFile) != raw.RawEvidenceFile || len(raw.RawEvidenceSHA256) != sha256.Size*2 ||
		raw.Profile != logN16PublicE32Profile || raw.Mode != "test-only-public-e32-fast" || raw.PrimaryCommit != primaryCommit ||
		raw.FixturePrimaryCommit != fixturePrimaryCommit || raw.FixturePrimaryCommit != manifest.PrimaryCommit || raw.PrimarySourceSHA256 != manifest.PrimarySourceSHA256 ||
		raw.ProductionFastCommit != e32LogN16ProductionFastCommit || raw.DiagnosticHead != diagnosticHead || raw.ConfigSHA256 != manifest.ConfigSHA256 ||
		raw.QPSHA256 != manifest.QPSHA256 || raw.InputSHA256 != manifest.InputSHA256 || raw.WorkloadSHA256 != manifest.WorkloadSHA256 ||
		raw.FixtureManifestSHA256 != manifestHash || raw.CiphertextSHA256 != manifest.CiphertextSHA256 || raw.ExpectedInputSHA256 != manifest.DecodedSHA256 ||
		raw.FastReferenceVectorsSHA256 != reference.VectorsSHA256 || raw.FastReferenceResultSHA256 != reference.ResultSHA256 ||
		raw.CallBudget != 2 || raw.ActualCalls != 2 || raw.NumericalGate != 1e-6 || raw.GoVersion != "go1.26.4" || raw.OS != "darwin" || raw.Arch != "arm64" ||
		raw.NumCPU != 10 || raw.GOMAXPROCS != 10 || raw.CPUProfileFile == "" || raw.CPUProfileSHA256 == "" || raw.HeapProfileFile == "" || raw.HeapProfileSHA256 == "" {
		return errors.New("LogN16 trace provenance or exact two-call budget does not match the v2 fixture/reference contract")
	}
	if !raw.ProductionSourceIdentical || raw.DiagnosticTestAdapterCommit != diagnosticHead || len(raw.DiagnosticTestAdapterCommit) != 40 ||
		!validSHA256Hex(raw.DiagnosticTestAdapterPatchSHA256) {
		return errors.New("LogN16 trace lacks verified production-source identity or isolated test-only adapter provenance")
	}
	if len(raw.Runs) != 2 || raw.Runs[0].Index != 1 || raw.Runs[0].Phase != "first_cold_bootstrap" || raw.Runs[1].Index != 2 || raw.Runs[1].Phase != "warm_bootstrap_01" {
		return errors.New("LogN16 trace must contain exactly one cold and one traced warm Bootstrap output")
	}
	for _, run := range raw.Runs {
		if run.ElapsedNS <= 0 || run.DecodedSHA256 == "" || run.Level != 1 || run.Degree != 1 || run.QPrefixRows != 2 || run.Scale != manifest.State.Scale ||
			!run.FastReferenceAvailable || !finiteNonNegativeE32(run.OracleRMSE) || !finiteNonNegativeE32(run.OracleMaxComplex) ||
			!finiteNonNegativeE32(run.FastReferenceWorstRMSE) || !finiteNonNegativeE32(run.FastReferenceWorstMax) ||
			run.OracleMaxComplex > raw.NumericalGate || run.FastReferenceWorstMax > raw.NumericalGate {
			return fmt.Errorf("LogN16 trace output %s fails its decoded oracle/state/reference gate", run.Phase)
		}
	}
	if len(vectors.Checkpoints) != len(logN16PublicCheckpointNames) || len(reference.Vectors.BootstrapOutputs) != 6 {
		return errors.New("LogN16 trace is missing current preflight checkpoints or validated Task005 output references")
	}
	return validatePublicE32Events(raw.Events)
}

func finiteNonNegativeE32(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
