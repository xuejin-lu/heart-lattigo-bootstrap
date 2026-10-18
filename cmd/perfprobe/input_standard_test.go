//go:build perf_standard

package main

import (
	"path/filepath"
	"testing"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestStandardNativeInputOriginAndRoundTrip(t *testing.T) {
	for _, profile := range []string{"logN13", "logN16"} {
		t.Run(profile, func(t *testing.T) {
			cfg, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config."+profile+".json"))
			if err != nil {
				t.Fatal(err)
			}
			residual, params, effective, err := perfmeasure.ParametersFromConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			values := perfmeasure.DeterministicInput(effective.InputSlots)
			first, err := newInputBackend(params, residual)
			if err != nil {
				t.Fatal(err)
			}
			_, _, a, err := prepareAndValidateInput(first, residual, params, values, effective.LogSlots)
			if err != nil {
				t.Fatal(err)
			}
			second, err := newInputBackend(params, residual)
			if err != nil {
				t.Fatal(err)
			}
			_, _, b, err := prepareAndValidateInput(second, residual, params, values, effective.LogSlots)
			if err != nil {
				t.Fatal(err)
			}
			if first.SecretKeyForTrial().Equal(second.SecretKeyForTrial()) || a.C1SHA256 == b.C1SHA256 || !a.C1Nonzero || !b.C1Nonzero {
				t.Fatal("Standard inputs did not use independent generated keys and nontrivial c1")
			}
			// A synthetic c0-copy/c1-zero artifact must fail even when carrying
			// otherwise valid CKKS metadata and the Standard adapter's secret.
			plain, err := perfmeasure.EncodeInputPlaintext(residual, effective.LogSlots, values)
			if err != nil {
				t.Fatal(err)
			}
			fake := ckks.NewCiphertext(residual, 1, 0)
			*fake.MetaData = *plain.MetaData
			fake.Value[0].Copy(plain.Value)
			fake.Value[1].Zero()
			fake.IsNTT, fake.IsMontgomery = plain.IsNTT, plain.IsMontgomery
			if _, _, err := validateInputCiphertext(first, residual, params, values, effective.LogSlots, fake); err == nil {
				t.Fatal("Standard formal preflight accepted synthetic zero-c1 input")
			}
		})
	}
}

func TestStandardAZeroBootstrapKeyAuditChecksQAndP(t *testing.T) {
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: 10, LogQ: []int{55, 45}, LogP: []int{60}, LogDefaultScale: 45,
	})
	if err != nil {
		t.Fatal(err)
	}
	kgen := rlwe.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyNew()
	rlk := kgen.GenRelinearizationKeyNew(sk)
	gk := kgen.GenGaloisKeyNew(params.GaloisElement(3), sk)
	denseToSparse := kgen.GenEvaluationKeyNew(sk, sk)
	sparseToDense := kgen.GenEvaluationKeyNew(sk, sk)
	keySet := rlwe.NewMemEvaluationKeySet(rlk, gk)
	backend := &standardBackend{bootstrapKeys: &bootstrapping.EvaluationKeys{
		EvkDenseToSparse: denseToSparse, EvkSparseToDense: sparseToDense, MemEvaluationKeySet: keySet,
	}}
	if err := verifyAZeroBootstrapKeys(backend); err != nil {
		t.Fatalf("valid zero-a evaluation key set rejected: %v", err)
	}

	gk.Value[0][0][1].P.Coeffs[0][0] = 1
	if err := verifyAZeroBootstrapKeys(backend); err == nil {
		t.Fatal("nonzero P residue in a was not rejected before Bootstrap")
	}
}

func TestStandardAZeroDiagnosticUsesNativeEncryptNewInput(t *testing.T) {
	for _, profile := range []string{"logN13", "logN16"} {
		t.Run(profile, func(t *testing.T) {
			cfg, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config."+profile+".json"))
			if err != nil {
				t.Fatal(err)
			}
			residual, params, effective, err := perfmeasure.ParametersFromConfigWithE(cfg, 32)
			if err != nil {
				t.Fatal(err)
			}
			if effective.EphemeralSecretWeight != 32 {
				t.Fatalf("effective E=%d, want 32", effective.EphemeralSecretWeight)
			}
			values := perfmeasure.DeterministicInput(effective.InputSlots)
			backend, err := newInputBackend(params, residual)
			if err != nil {
				t.Fatal(err)
			}
			input, decoded, evidence, err := prepareAndValidateInputMode(backend, residual, params, values, effective.LogSlots, true)
			if err != nil {
				t.Fatal(err)
			}
			c1Nonzero, err := isCiphertextComponentNonzero(input, 1)
			if err != nil {
				t.Fatal(err)
			}
			if c1Nonzero || evidence.C1Nonzero || evidence.Kind != standardAZeroInputKind || evidence.Constructor != standardAZeroConstructor {
				t.Fatalf("A=0 input evidence is inconsistent: c1_nonzero=%t evidence=%+v", c1Nonzero, evidence)
			}
			if len(decoded) != effective.InputSlots || evidence.MaxComplexDeviation > inputQualityLimit {
				t.Fatalf("native A=0 input decode failed its slot/quality gate: slots=%d evidence=%+v", len(decoded), evidence)
			}
		})
	}
}
