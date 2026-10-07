//go:build perf_fast

package main

import (
	"path/filepath"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestFastDirectEncodedSimulationInput(t *testing.T) {
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
			backend, err := newInputBackend(params, residual)
			if err != nil {
				t.Fatal(err)
			}
			_, _, evidence, err := prepareAndValidateInput(backend, residual, params, perfmeasure.DeterministicInput(effective.InputSlots), effective.LogSlots)
			if err != nil {
				t.Fatal(err)
			}
			if evidence.Kind != fastDirectInputKind || evidence.C1Nonzero || evidence.State.Level != 0 || evidence.State.LogCols != effective.LogSlots {
				t.Fatalf("Fast simulation input evidence: %+v", evidence)
			}
		})
	}
}
