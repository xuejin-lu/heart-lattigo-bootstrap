//go:build perf_standard && ephemeral_diag

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const ephemeralAblationStandardCommit = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"

type ephemeralAblationRun struct {
	Weight                  int                      `json:"ephemeral_secret_weight"`
	DenseToSparseKeyPresent bool                     `json:"dense_to_sparse_key_present"`
	SparseToDenseKeyPresent bool                     `json:"sparse_to_dense_key_present"`
	PublicBootstrapCalls    int                      `json:"public_bootstrap_calls"`
	OutputMetadata          outputCiphertextMetadata `json:"output_metadata"`
	OutputAgainstOriginal   outputComparison         `json:"output_against_original"`
}

type ephemeralAblationDocument struct {
	SchemaVersion                  string                 `json:"schema_version"`
	Timestamp                      time.Time              `json:"timestamp"`
	Task                           string                 `json:"task"`
	Profile                        string                 `json:"profile"`
	PrimaryCommit                  string                 `json:"primary_commit"`
	PrimaryDirty                   bool                   `json:"primary_dirty"`
	StandardCommit                 string                 `json:"standard_commit"`
	StandardSource                 string                 `json:"standard_source"`
	StandardRef                    string                 `json:"standard_ref"`
	StandardDirty                  bool                   `json:"standard_dirty"`
	BuildTag                       string                 `json:"build_tag"`
	ConfigPath                     string                 `json:"config_path"`
	ConfigSHA256                   string                 `json:"config_sha256"`
	OriginalSHA256                 string                 `json:"original_sha256"`
	EffectiveParametersSHA256      string                 `json:"effective_parameters_sha256"`
	QPrimesSHA256                  string                 `json:"q_primes_sha256"`
	PPrimesSHA256                  string                 `json:"p_primes_sha256"`
	LogN                           int                    `json:"log_n"`
	LogSlots                       int                    `json:"log_slots"`
	Slots                          int                    `json:"slots"`
	SecretHammingWeight            int                    `json:"secret_hamming_weight"`
	Mod1K                          int                    `json:"mod1_k"`
	Mod1Degree                     int                    `json:"mod1_degree"`
	QChainBits                     []int                  `json:"q_chain_bits"`
	PBits                          []int                  `json:"p_bits"`
	InputKind                      string                 `json:"input_kind"`
	InputConstructor               string                 `json:"input_constructor"`
	InputNontrivialC1              bool                   `json:"input_nontrivial_c1"`
	InputMaxComplexDeviation       float64                `json:"input_max_complex_deviation"`
	SameSecretAndCiphertextForPair bool                   `json:"same_secret_and_ciphertext_for_pair"`
	Runs                           []ephemeralAblationRun `json:"runs"`
	E32VsE0                        outputComparison       `json:"e32_vs_e0"`
	E32RMSEReductionPercent        float64                `json:"e32_rmse_reduction_percent"`
	GoVersion                      string                 `json:"go_version"`
	OS                             string                 `json:"os"`
	Arch                           string                 `json:"arch"`
	Limitations                    []string               `json:"limitations"`
}

// This test is the sole entry point for the diagnostic. It only exists in a
// perf_standard && ephemeral_diag test build, never in the normal harness.
func TestFastStandardEphemeralAblation(t *testing.T) {
	profile := os.Getenv("FAST_STANDARD_EPHEMERAL_PROFILE")
	if profile == "" {
		t.Skip("set FAST_STANDARD_EPHEMERAL_PROFILE=logn13 (logn16 is only for a discriminating LogN13 result)")
	}
	if profile != "logn13" && profile != "logn16" {
		t.Fatalf("unsupported diagnostic profile %q", profile)
	}
	standardRoot := os.Getenv("FAST_STANDARD_EPHEMERAL_STANDARD_ROOT")
	if standardRoot == "" {
		t.Fatal("FAST_STANDARD_EPHEMERAL_STANDARD_ROOT is required")
	}
	outputPath := os.Getenv("FAST_STANDARD_EPHEMERAL_OUT")
	if outputPath == "" || !filepath.IsAbs(outputPath) {
		t.Fatal("FAST_STANDARD_EPHEMERAL_OUT must be an absolute path outside the repository")
	}

	primaryRoot, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		t.Fatalf("Primary provenance: %v", err)
	}
	if err := verifyCompiledBackendSource(standardRoot); err != nil {
		t.Fatalf("compiled Standard source: %v", err)
	}
	standardPath, standardCommit, standardRef, err := cleanSecondaryState(standardRoot, ephemeralAblationStandardCommit)
	if err != nil {
		t.Fatalf("pinned Standard provenance: %v", err)
	}
	if standardRef != standardCommit {
		t.Fatalf("pinned Standard source must be detached at its exact commit; ref=%q commit=%q", standardRef, standardCommit)
	}
	relativeOutput, err := filepath.Rel(primaryRoot, outputPath)
	if err != nil || !hasParentPrefix(relativeOutput) {
		t.Fatalf("diagnostic output must be outside Primary repository: %s", outputPath)
	}

	configPath := filepath.Join(primaryRoot, "configs", "bootstrap_config."+profile+".json")
	cfg, configSum, err := perfmeasure.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load canonical %s config: %v", profile, err)
	}
	if cfg.LogN != expectedLogN(profile) || cfg.SecretHamming != 192 || cfg.Mod1K != 16 {
		t.Fatalf("canonical diagnostic invariants changed: LogN=%d H=%d K=%d", cfg.LogN, cfg.SecretHamming, cfg.Mod1K)
	}
	residual, baseParams, effective, err := perfmeasure.ParametersFromConfig(cfg)
	if err != nil {
		t.Fatalf("construct canonical parameters: %v", err)
	}
	if baseParams.EphemeralSecretWeight != 0 {
		t.Fatalf("frozen config must construct E=0, got %d", baseParams.EphemeralSecretWeight)
	}
	variantParams := baseParams
	variantParams.EphemeralSecretWeight = 32
	if variantParams.EphemeralSecretWeight != 32 || baseParams.EphemeralSecretWeight != 0 ||
		!variantParams.BootstrappingParameters.Equal(&baseParams.BootstrappingParameters) ||
		!variantParams.ResidualParameters.Equal(&baseParams.ResidualParameters) {
		t.Fatal("diagnostic variants differ beyond EphemeralSecretWeight")
	}

	values := perfmeasure.DeterministicInput(effective.InputSlots)
	keygen := rlwe.NewKeyGenerator(baseParams.BootstrappingParameters)
	secret := keygen.GenSecretKeyNew()
	inputBackend := &standardBackend{residual: residual, bootstrappingParams: baseParams.BootstrappingParameters, secret: secret}
	input, err := inputBackend.PrepareInput(values, effective.LogSlots)
	if err != nil {
		t.Fatalf("native Standard EncryptNew: %v", err)
	}
	preDecoded, inputEvidence, err := validateInputCiphertext(inputBackend, residual, baseParams, values, effective.LogSlots, input)
	if err != nil {
		t.Fatalf("validate native Standard input: %v", err)
	}
	if !inputEvidence.C1Nonzero || inputEvidence.Kind != standardNativeInputKind || inputEvidence.MaxComplexDeviation > inputQualityLimit {
		t.Fatalf("input is synthetic, trivial-c1, or inaccurate: %+v", inputEvidence)
	}
	inputSnapshot := input.CopyNew()
	if len(preDecoded) != effective.InputSlots {
		t.Fatalf("decoded input slots=%d, expected=%d", len(preDecoded), effective.InputSlots)
	}

	runs := make([]ephemeralAblationRun, 0, 2)
	outputs := make([][]complex128, 0, 2)
	for _, diagnosticParams := range []bootstrapping.Parameters{baseParams, variantParams} {
		keys, _, err := diagnosticParams.GenEvaluationKeys(secret)
		if err != nil {
			t.Fatalf("generate E=%d Standard evaluation keys: %v", diagnosticParams.EphemeralSecretWeight, err)
		}
		denseToSparsePresent := keys.EvkDenseToSparse != nil
		sparseToDensePresent := keys.EvkSparseToDense != nil
		wantPresent := diagnosticParams.EphemeralSecretWeight == 32
		if denseToSparsePresent != wantPresent || sparseToDensePresent != wantPresent {
			t.Fatalf("E=%d encapsulation key evidence dense->sparse=%t sparse->dense=%t, want both=%t",
				diagnosticParams.EphemeralSecretWeight, denseToSparsePresent, sparseToDensePresent, wantPresent)
		}
		evaluator, err := bootstrapping.NewEvaluator(diagnosticParams, keys)
		if err != nil {
			t.Fatalf("construct E=%d Standard evaluator: %v", diagnosticParams.EphemeralSecretWeight, err)
		}
		calls := 0
		calls++
		output, err := evaluator.Bootstrap(input.CopyNew())
		if err != nil {
			t.Fatalf("E=%d public Bootstrap: %v", diagnosticParams.EphemeralSecretWeight, err)
		}
		if output == nil {
			t.Fatalf("E=%d public Bootstrap returned a nil ciphertext", diagnosticParams.EphemeralSecretWeight)
		}
		if calls != 1 {
			t.Fatalf("E=%d public Bootstrap calls=%d, expected exactly one", diagnosticParams.EphemeralSecretWeight, calls)
		}
		if !input.Equal(inputSnapshot) {
			t.Fatalf("E=%d Bootstrap modified the shared source ciphertext", diagnosticParams.EphemeralSecretWeight)
		}
		decoded, err := inputBackend.Decode(output)
		if err != nil {
			t.Fatalf("E=%d decrypt/decode with matching Standard secret: %v", diagnosticParams.EphemeralSecretWeight, err)
		}
		if len(decoded) != effective.InputSlots || !allFinite(decoded) {
			t.Fatalf("E=%d output has invalid slot count or non-finite decoded values", diagnosticParams.EphemeralSecretWeight)
		}
		state := stateOf(output, baseParams.BootstrappingParameters.Q(), output.Level()+1)
		metadata := snapshotOutputMetadata(output, state, len(decoded))
		if err := validateOutputCiphertext(output, metadata, baseParams.BootstrappingParameters.MaxLevel(), effective.LogSlots); err != nil {
			t.Fatalf("E=%d output metadata: %v", diagnosticParams.EphemeralSecretWeight, err)
		}
		if !metadata.IsNTT || metadata.IsMontgomery || metadata.ComponentCount != 2 {
			t.Fatalf("E=%d output representation is unexpected: NTT=%t Montgomery=%t components=%d",
				diagnosticParams.EphemeralSecretWeight, metadata.IsNTT, metadata.IsMontgomery, metadata.ComponentCount)
		}
		runs = append(runs, ephemeralAblationRun{
			Weight:                  diagnosticParams.EphemeralSecretWeight,
			DenseToSparseKeyPresent: denseToSparsePresent, SparseToDenseKeyPresent: sparseToDensePresent,
			PublicBootstrapCalls: calls, OutputMetadata: metadata, OutputAgainstOriginal: makeOutputComparison(values, decoded),
		})
		outputs = append(outputs, decoded)
	}
	if len(runs) != 2 || runs[0].Weight != 0 || runs[1].Weight != 32 || !input.Equal(inputSnapshot) {
		t.Fatal("controlled pair did not reuse the same secret and exact native ciphertext in E=0, E=32 order")
	}
	if runs[0].PublicBootstrapCalls != 1 || runs[1].PublicBootstrapCalls != 1 {
		t.Fatalf("controlled pair must execute exactly one Bootstrap per weight: %d, %d", runs[0].PublicBootstrapCalls, runs[1].PublicBootstrapCalls)
	}

	effectiveJSON, err := json.Marshal(effective)
	if err != nil {
		t.Fatal(err)
	}
	qJSON, err := json.Marshal(effective.QPrimes)
	if err != nil {
		t.Fatal(err)
	}
	pJSON, err := json.Marshal(effective.PPrimes)
	if err != nil {
		t.Fatal(err)
	}
	configHash := hex.EncodeToString(configSum[:])
	doc := ephemeralAblationDocument{
		SchemaVersion: "fast-standard-ephemeral-ablation.v1", Timestamp: time.Now().UTC(),
		Task: "FAST-STANDARD-EPHEMERAL-ABLATION-001", Profile: profile,
		PrimaryCommit: primaryCommit, PrimaryDirty: false,
		StandardCommit: standardCommit, StandardSource: standardPath, StandardRef: standardRef, StandardDirty: false,
		BuildTag: "perf_standard && ephemeral_diag", ConfigPath: configPath, ConfigSHA256: configHash,
		OriginalSHA256: perfmeasure.Fingerprint(values), EffectiveParametersSHA256: sha256Hex(effectiveJSON),
		QPrimesSHA256: sha256Hex(qJSON), PPrimesSHA256: sha256Hex(pJSON),
		LogN: effective.LogN, LogSlots: effective.LogSlots, Slots: effective.InputSlots,
		SecretHammingWeight: cfg.SecretHamming, Mod1K: effective.K, Mod1Degree: effective.Mod1Degree,
		QChainBits: effective.QChainBits, PBits: effective.PBits,
		InputKind: inputEvidence.Kind, InputConstructor: inputEvidence.Constructor,
		InputNontrivialC1: inputEvidence.C1Nonzero, InputMaxComplexDeviation: inputEvidence.MaxComplexDeviation,
		SameSecretAndCiphertextForPair: true, Runs: runs,
		E32VsE0:   makeOutputComparison(outputs[0], outputs[1]),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
		Limitations: []string{
			"Diagnostic-only native Standard contrast; E=32 is not promoted to the formal profile and is not compared with Fast.",
			"One generated Standard secret and one EncryptNew ciphertext are reused; exactly one public Bootstrap is run per weight.",
			"No timing, warmup, benchmark, parameter sweep, ciphertext/key material, or key fingerprint is recorded.",
			"Output accuracy remains unassessed pending independent Web review.",
		},
	}
	rmse0 := runs[0].OutputAgainstOriginal.Error.ComplexRMSE
	rmse32 := runs[1].OutputAgainstOriginal.Error.ComplexRMSE
	if !finiteAblation(rmse0) || !finiteAblation(rmse32) || rmse0 <= 0 {
		t.Fatalf("invalid controlled output RMSE values E=0 %.9g E=32 %.9g", rmse0, rmse32)
	}
	doc.E32RMSEReductionPercent = 100 * (rmse0 - rmse32) / rmse0
	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("encode compact diagnostic evidence: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		t.Fatalf("create external output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write external diagnostic evidence: %v", err)
	}
	t.Logf("profile=%s primary=%s standard=%s input_sha256=%s input_maxdev=%.9g E0_RMSE=%.9g E32_RMSE=%.9g reduction=%.4f%% calls=%d/%d keys=(%t,%t)/(%t,%t) output=%s",
		profile, primaryCommit, standardCommit, doc.OriginalSHA256, doc.InputMaxComplexDeviation,
		rmse0, rmse32, doc.E32RMSEReductionPercent, runs[0].PublicBootstrapCalls, runs[1].PublicBootstrapCalls,
		runs[0].DenseToSparseKeyPresent, runs[0].SparseToDenseKeyPresent,
		runs[1].DenseToSparseKeyPresent, runs[1].SparseToDenseKeyPresent, outputPath)

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("diagnostic artifact missing: %v", err)
	}
}

func expectedLogN(profile string) int {
	if profile == "logn13" {
		return 13
	}
	return 16
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func finiteAblation(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func hasParentPrefix(path string) bool {
	return path == ".." || len(path) > 3 && path[:3] == ".."+string(os.PathSeparator)
}
