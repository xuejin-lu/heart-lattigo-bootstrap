//go:build perf_standard

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
)

func verifyAZeroBootstrapKeys(backend backendAdapter) error {
	standard, ok := backend.(*standardBackend)
	if !ok || standard.bootstrapKeys == nil {
		return fmt.Errorf("A=0 diagnostic did not receive the generated Standard Bootstrap key set")
	}
	keys := standard.bootstrapKeys
	if keys.EvkDenseToSparse == nil || keys.EvkSparseToDense == nil {
		return fmt.Errorf("E=32 Bootstrap key set is missing dense/sparse encapsulation keys")
	}
	for _, item := range []struct {
		name string
		evk  *rlwe.EvaluationKey
	}{
		{name: "N1-to-N2", evk: keys.EvkN1ToN2},
		{name: "N2-to-N1", evk: keys.EvkN2ToN1},
		{name: "real-to-complex", evk: keys.EvkRealToCmplx},
		{name: "complex-to-real", evk: keys.EvkCmplxToReal},
		{name: "dense-to-sparse", evk: keys.EvkDenseToSparse},
		{name: "sparse-to-dense", evk: keys.EvkSparseToDense},
	} {
		if item.evk != nil {
			if err := verifyAZeroEvaluationKey(item.name, item.evk); err != nil {
				return err
			}
		}
	}
	relinearizationKey, err := keys.GetRelinearizationKey()
	if err != nil {
		return fmt.Errorf("read Bootstrap relinearization key: %w", err)
	}
	if err := verifyAZeroEvaluationKey("relinearization", &relinearizationKey.EvaluationKey); err != nil {
		return err
	}
	for _, galEl := range keys.GetGaloisKeysList() {
		galoisKey, err := keys.GetGaloisKey(galEl)
		if err != nil {
			return fmt.Errorf("read Bootstrap Galois key %d: %w", galEl, err)
		}
		if err := verifyAZeroEvaluationKey(fmt.Sprintf("Galois-%d", galEl), &galoisKey.EvaluationKey); err != nil {
			return err
		}
	}
	return nil
}

func verifyAZeroEvaluationKey(name string, evk *rlwe.EvaluationKey) error {
	if evk == nil || len(evk.Value) == 0 || len(evk.Value[0]) == 0 || evk.Degree() != 1 {
		return fmt.Errorf("Bootstrap key %s is nil or not an uncompressed degree-1 key", name)
	}
	for i := range evk.Value {
		if len(evk.Value[i]) == 0 {
			return fmt.Errorf("Bootstrap key %s has an empty decomposition row %d", name, i)
		}
		for j := range evk.Value[i] {
			if len(evk.Value[i][j]) != 2 {
				return fmt.Errorf("Bootstrap key %s entry (%d,%d) has %d components, want 2", name, i, j, len(evk.Value[i][j]))
			}
			for _, domain := range []struct {
				name string
				pol  ring.Poly
			}{
				{name: "Q", pol: evk.Value[i][j][1].Q},
				{name: "P", pol: evk.Value[i][j][1].P},
			} {
				for level := range domain.pol.Coeffs {
					for coefficientIndex, coefficient := range domain.pol.Coeffs[level] {
						if coefficient != 0 {
							return fmt.Errorf("Bootstrap key %s entry (%d,%d) has nonzero a residue in %s level %d coefficient %d", name, i, j, domain.name, level, coefficientIndex)
						}
					}
				}
			}
		}
	}
	return nil
}
