package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBootstrapConfigRejectsRemovedLegacyFields(t *testing.T) {
	for _, field := range []string{"q_circuit_slots", "q_eval_mod", "log_bsgs_ratio"} {
		t.Run(field, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			data := []byte(`{"` + field + `": 1}`)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadBootstrapConfig(path); err == nil {
				t.Fatalf("LoadBootstrapConfig accepted removed legacy field %q", field)
			}
		})
	}
}

func TestConfigPrimeScaleChangesEffectiveParameters(t *testing.T) {
	base := DefaultBootstrapConfig()
	_, baseBTP, err := NewBootstrapParametersFromConfig(base)
	if err != nil {
		t.Fatal(err)
	}

	changed := base
	changed.QCoeffsToSlots = append([]int(nil), base.QCoeffsToSlots...)
	changed.QCoeffsToSlots[0]--
	_, changedBTP, err := NewBootstrapParametersFromConfig(changed)
	if err != nil {
		t.Fatal(err)
	}

	baseQBits := baseBTP.BootstrappingParameters.LogQi()
	changedQBits := changedBTP.BootstrappingParameters.LogQi()
	if len(baseQBits) != len(changedQBits) {
		t.Fatalf("changing one prime scale changed Q count: %d != %d", len(baseQBits), len(changedQBits))
	}
	// The first CoeffsToSlots prime is appended after residual, SlotsToCoeffs,
	// and EvalMod primes. Its effective bit length must follow the config.
	const firstCoeffsToSlotsIndex = 2 + 3 + 8
	if baseQBits[firstCoeffsToSlotsIndex] == changedQBits[firstCoeffsToSlotsIndex] {
		t.Fatalf("changing q_coeffs_to_slots did not change the effective prime bit length")
	}
}
