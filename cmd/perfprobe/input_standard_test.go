//go:build perf_standard

package main

import (
	"path/filepath"
	"testing"

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
