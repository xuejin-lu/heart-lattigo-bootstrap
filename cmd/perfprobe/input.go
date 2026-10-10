package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/cmplx"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	pinnedStandardSHA         = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	pinnedFastSHA             = "5feb44917fca40c93abec6def6f26bc81a82c536"
	standardNativeInputKind   = "standard_native_rlwe_encrypt_v1"
	fastDirectInputKind       = "fast_zero_secret_direct_encoded_v1"
	standardNativeConstructor = "ckks.NewEncoder.Encode -> rlwe.NewEncryptor(residual, generated Standard secret).EncryptNew"
	fastDirectConstructor     = "ckks.NewEncoder.Encode -> Fast test-helper-equivalent c0 copy + c1 zero (simulation, not encryption)"
	inputQualityLimit         = 1e-6
)

func validatePinnedBackend(backend, commit string) error {
	want := ""
	switch backend {
	case "standard":
		want = pinnedStandardSHA
	case "fast":
		want = pinnedFastSHA
	default:
		return fmt.Errorf("unsupported backend %q", backend)
	}
	if commit != want {
		return fmt.Errorf("%s source SHA %s differs from input-contract pin %s", backend, commit, want)
	}
	return nil
}

type inputRecord struct {
	Kind                string          `json:"kind"`
	Constructor         string          `json:"constructor"`
	State               ciphertextState `json:"ciphertext_state"`
	MetadataSHA256      string          `json:"metadata_sha256"`
	C1SHA256            string          `json:"c1_q0_sha256"`
	C1Nonzero           bool            `json:"c1_nonzero"`
	PreDecodedSHA256    string          `json:"pre_decoded_sha256"`
	MaxComplexDeviation float64         `json:"max_complex_deviation"`
	InputQualityLimit   float64         `json:"input_quality_limit"`
}

type inputSmokeDocument struct {
	SchemaVersion               string                          `json:"schema_version"`
	Timestamp                   time.Time                       `json:"timestamp"`
	Profile                     string                          `json:"profile"`
	Backend                     string                          `json:"backend"`
	BuildTag                    string                          `json:"build_tag"`
	PrimaryCommit               string                          `json:"primary_commit"`
	BackendCommit               string                          `json:"backend_commit"`
	BackendRef                  string                          `json:"backend_ref"`
	BackendCheckoutRef          string                          `json:"backend_checkout_ref"`
	BackendSource               string                          `json:"backend_source"`
	ConfigPath                  string                          `json:"config_path"`
	ConfigSHA256                string                          `json:"config_sha256"`
	OriginalSHA256              string                          `json:"original_sha256"`
	Parameters                  perfmeasure.EffectiveParameters `json:"effective_parameters"`
	GoVersion                   string                          `json:"go_version"`
	OS                          string                          `json:"os"`
	Arch                        string                          `json:"arch"`
	FreshKeyEvidence            string                          `json:"fresh_key_evidence"`
	FastPublicBootstrapAccepted bool                            `json:"fast_public_bootstrap_accepted"`
	Trials                      []inputRecord                   `json:"input_trials"`
	Limitations                 []string                        `json:"limitations"`
}

func prepareAndValidateInput(backend backendAdapter, residual ckks.Parameters, params bootstrapping.Parameters, values []complex128, logSlots int) (*rlwe.Ciphertext, []complex128, inputRecord, error) {
	if backend.Name() == "standard" && backend.SecretKeyForTrial() == nil {
		return nil, nil, inputRecord{}, fmt.Errorf("Standard input has no generated matching secret key")
	}
	ct, err := backend.PrepareInput(values, logSlots)
	if err != nil {
		return nil, nil, inputRecord{}, fmt.Errorf("prepare %s input: %w", backend.Name(), err)
	}
	decoded, record, err := validateInputCiphertext(backend, residual, params, values, logSlots, ct)
	return ct, decoded, record, err
}

func validateInputCiphertext(backend backendAdapter, residual ckks.Parameters, params bootstrapping.Parameters, values []complex128, logSlots int, ct *rlwe.Ciphertext) ([]complex128, inputRecord, error) {
	if ct == nil || ct.Level() != 0 || ct.Degree() != 1 || !ct.IsNTT || ct.IsMontgomery || ct.LogDimensions.Rows != 0 || ct.LogDimensions.Cols != logSlots || !ct.Scale.Equal(residual.DefaultScale()) {
		return nil, inputRecord{}, fmt.Errorf("%s input violates residual Level-0 ciphertext metadata contract", backend.Name())
	}
	if len(ct.Value) != 2 || len(ct.Value[1].Coeffs) != 1 || len(ct.Value[1].Coeffs[0]) != residual.N() {
		return nil, inputRecord{}, fmt.Errorf("%s input c1 q0 row is unavailable", backend.Name())
	}
	c1Hash := sha256.New()
	var encoded [8]byte
	var c1Nonzero bool
	for _, coefficient := range ct.Value[1].Coeffs[0] {
		binary.LittleEndian.PutUint64(encoded[:], coefficient)
		_, _ = c1Hash.Write(encoded[:])
		c1Nonzero = c1Nonzero || coefficient != 0
	}
	switch backend.Name() {
	case "standard":
		if backend.InputKind() != standardNativeInputKind || !c1Nonzero {
			return nil, inputRecord{}, fmt.Errorf("Standard formal input is not a native nontrivial-c1 encryption")
		}
	case "fast":
		if backend.InputKind() != fastDirectInputKind || c1Nonzero {
			return nil, inputRecord{}, fmt.Errorf("Fast simulation input does not have declared zero-a semantics")
		}
	default:
		return nil, inputRecord{}, fmt.Errorf("unsupported input backend %q", backend.Name())
	}
	decoded, err := backend.Decode(ct)
	if err != nil {
		return nil, inputRecord{}, fmt.Errorf("decode %s input: %w", backend.Name(), err)
	}
	if len(decoded) != len(values) || len(values) != 1<<logSlots {
		return nil, inputRecord{}, fmt.Errorf("%s decoded slot count=%d, original=%d, expected=%d", backend.Name(), len(decoded), len(values), 1<<logSlots)
	}
	var maxDeviation float64
	for i, value := range decoded {
		if math.IsNaN(real(value)) || math.IsNaN(imag(value)) || math.IsInf(real(value), 0) || math.IsInf(imag(value), 0) {
			return nil, inputRecord{}, fmt.Errorf("%s decoded slot %d is non-finite", backend.Name(), i)
		}
		maxDeviation = math.Max(maxDeviation, cmplx.Abs(value-values[i]))
	}
	if maxDeviation > inputQualityLimit {
		return nil, inputRecord{}, fmt.Errorf("%s input pre-Bootstrap max complex deviation %.9g exceeds %.9g", backend.Name(), maxDeviation, inputQualityLimit)
	}
	state, err := backendStateOf(backend, ct, params.BootstrappingParameters.Q())
	if err != nil {
		return nil, inputRecord{}, err
	}
	metadata, err := json.Marshal(state)
	if err != nil {
		return nil, inputRecord{}, err
	}
	metadataHash := sha256.Sum256(metadata)
	return decoded, inputRecord{
		Kind: backend.InputKind(), Constructor: backend.InputConstructor(), State: state,
		MetadataSHA256: hex.EncodeToString(metadataHash[:]), C1SHA256: hex.EncodeToString(c1Hash.Sum(nil)),
		C1Nonzero: c1Nonzero, PreDecodedSHA256: perfmeasure.Fingerprint(decoded),
		MaxComplexDeviation: maxDeviation, InputQualityLimit: inputQualityLimit,
	}, nil
}

func verifyCompiledBackendSource(expected string) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("Go build info is unavailable")
	}
	for _, dependency := range info.Deps {
		if dependency.Path == "github.com/tuneinsight/lattigo/v6" {
			if dependency.Replace == nil {
				return fmt.Errorf("compiled Lattigo has no local source replacement")
			}
			actual, err := filepath.EvalSymlinks(dependency.Replace.Path)
			if err != nil {
				return err
			}
			wanted, err := filepath.EvalSymlinks(expected)
			if err != nil {
				return err
			}
			if actual != wanted {
				return fmt.Errorf("compiled Lattigo source %s differs from pinned source %s", actual, wanted)
			}
			return nil
		}
	}
	return fmt.Errorf("compiled Lattigo dependency is absent from Go build info")
}

func runInputSmoke(opts cliOptions, configHash string, residual ckks.Parameters, params bootstrapping.Parameters, effective perfmeasure.EffectiveParameters, values []complex128) error {
	if err := verifyCompiledBackendSource(opts.secondaryRoot); err != nil {
		return err
	}
	_, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		return fmt.Errorf("Primary provenance: %w", err)
	}
	secondaryRoot, secondaryCommit, secondaryRef, err := cleanSecondaryState(opts.secondaryRoot, opts.backendCommit)
	if err != nil {
		return fmt.Errorf("Secondary provenance: %w", err)
	}
	backend, err := newInputBackend(params, residual)
	if err != nil {
		return err
	}
	if err := validatePinnedBackend(backend.Name(), secondaryCommit); err != nil {
		return err
	}
	count := 1
	if backend.Name() == "standard" {
		count = 2
	}
	trials := make([]inputRecord, 0, count)
	var firstKey *rlwe.SecretKey
	var firstInput *rlwe.Ciphertext
	for i := 0; i < count; i++ {
		if i > 0 {
			backend, err = newInputBackend(params, residual)
			if err != nil {
				return err
			}
		}
		ct, _, record, err := prepareAndValidateInput(backend, residual, params, values, effective.LogSlots)
		if err != nil {
			return fmt.Errorf("input smoke trial %d: %w", i+1, err)
		}
		if backend.Name() == "standard" {
			key := backend.SecretKeyForTrial()
			if i == 0 {
				firstKey = key
			} else if key == nil || key.Equal(firstKey) || record.C1SHA256 == trials[0].C1SHA256 {
				return fmt.Errorf("Standard trial %d did not produce an independent key and c1", i+1)
			}
		}
		if i == 0 {
			firstInput = ct
		}
		trials = append(trials, record)
	}
	accepted := false
	if opts.fastBootstrapAcceptance {
		if backend.Name() != "fast" {
			return fmt.Errorf("--fast-bootstrap-acceptance is only valid for the Fast simulation adapter")
		}
		fullBackend, err := newBackend(params, residual)
		if err != nil {
			return err
		}
		guardedBackend := &budgetedBackend{backendAdapter: fullBackend, budget: &bootstrapBudget{limit: opts.bootstrapBudget, journalBase: opts.out}}
		setBootstrapPhase(guardedBackend, "fast_public_acceptance_smoke")
		if _, err := guardedBackend.Bootstrap(firstInput.CopyNew()); err != nil {
			return fmt.Errorf("Fast public Bootstrap rejected prepared input: %w", err)
		}
		accepted = true
	}
	keyEvidence := "not applicable: Fast direct-encoded zero-a simulation has no encryption key"
	if backend.Name() == "standard" {
		keyEvidence = "two independently generated Standard secret keys compared in memory; separate EncryptNew calls; distinct c1 q0 SHA-256"
	}
	doc := inputSmokeDocument{
		SchemaVersion: "fast-standard-input-provenance.v1", Timestamp: time.Now().UTC(),
		Profile: opts.profile, Backend: backend.Name(), BuildTag: "perf_" + backend.Name(),
		PrimaryCommit: primaryCommit, BackendCommit: secondaryCommit, BackendRef: opts.backendRef, BackendCheckoutRef: secondaryRef,
		BackendSource: secondaryRoot, ConfigPath: opts.config, ConfigSHA256: configHash,
		OriginalSHA256: perfmeasure.Fingerprint(values), Parameters: effective,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
		FreshKeyEvidence: keyEvidence, FastPublicBootstrapAccepted: accepted, Trials: trials,
		Limitations: []string{"Input preflight only: no timed campaign, SNR, or output accuracy measurement."},
	}
	if accepted {
		doc.Limitations = append(doc.Limitations, "One Fast public Bootstrap call checked input acceptance only; its output was not evaluated.")
	}
	if backend.Name() == "fast" {
		doc.Limitations = append(doc.Limitations, "Fast input is an intentionally insecure zero-secret/zero-a direct-encoded simulation, not native or secure public-key encryption.")
	}
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		return err
	}
	fmt.Printf("%s %s input-smoke=%s trials=%d max-deviation=%.9g\n", opts.profile, backend.Name(), opts.out, len(trials), trials[0].MaxComplexDeviation)
	return nil
}
