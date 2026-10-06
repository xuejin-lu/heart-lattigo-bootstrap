package main

import (
	"path/filepath"
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestDeterministicLogN13InputKeepsCanonicalFingerprint(t *testing.T) {
	if got := perfmeasure.Fingerprint(perfmeasure.DeterministicInput(1 << 12)); got != canonicalLogN13InputSHA256 {
		t.Fatalf("LogN13 input SHA-256 = %s, want %s", got, canonicalLogN13InputSHA256)
	}
}

func TestEffectiveProfilesComeFromTheirConfigFiles(t *testing.T) {
	for _, test := range []struct {
		name, file            string
		logN, logSlots, slots int
	}{
		{name: "logn13", file: "bootstrap_config.logN13.json", logN: 13, logSlots: 12, slots: 1 << 12},
		{name: "logn16", file: "bootstrap_config.logN16.json", logN: 16, logSlots: 15, slots: 1 << 15},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, _, err := perfmeasure.LoadConfig(filepath.Join("..", "..", "configs", test.file))
			if err != nil {
				t.Fatal(err)
			}
			residual, params, effective, err := perfmeasure.ParametersFromConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if effective.LogN != test.logN || effective.LogSlots != test.logSlots || effective.InputSlots != test.slots {
				t.Fatalf("unexpected effective dimensions: %+v", effective)
			}
			if len(effective.QChainBits) != params.BootstrappingParameters.MaxLevel()+1 || len(effective.PBits) != len(params.BootstrappingParameters.P()) {
				t.Fatalf("effective chain metadata does not match constructed parameters: %+v", effective)
			}
			if residual.LogN() != test.logN || cfg.Mod1Degree != 30 || cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 {
				t.Fatalf("config did not preserve required effective inputs: cfg=%+v residual LogN=%d", cfg, residual.LogN())
			}
		})
	}
}
