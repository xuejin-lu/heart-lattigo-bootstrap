package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
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

func TestSharedPerformanceProfileParametersMatchExperimentConfigBuilder(t *testing.T) {
	for _, name := range []string{"bootstrap_config.logN13.json", "bootstrap_config.logN16.json"} {
		t.Run(name, func(t *testing.T) {
			cfg, err := LoadBootstrapConfig(filepath.Join("configs", name))
			if err != nil {
				t.Fatal(err)
			}
			wantResidual, wantBTP, err := NewBootstrapParametersFromConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			gotResidual, gotBTP, _, err := perfmeasure.ParametersFromConfig(perfmeasure.Config(cfg))
			if err != nil {
				t.Fatal(err)
			}
			if !wantResidual.Equal(&gotResidual) {
				t.Fatal("shared measurement residual parameters differ from the canonical experiment builder")
			}
			if !wantBTP.BootstrappingParameters.Equal(&gotBTP.BootstrappingParameters) ||
				!reflect.DeepEqual(wantBTP.CoeffsToSlotsParameters, gotBTP.CoeffsToSlotsParameters) ||
				!reflect.DeepEqual(wantBTP.SlotsToCoeffsParameters, gotBTP.SlotsToCoeffsParameters) ||
				!reflect.DeepEqual(wantBTP.Mod1ParametersLiteral, gotBTP.Mod1ParametersLiteral) {
				t.Fatal("shared measurement effective parameters differ from the canonical experiment builder")
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
