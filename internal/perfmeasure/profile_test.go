package perfmeasure

import (
	"path/filepath"
	"testing"
)

func TestCanonicalLogN13Fingerprint(t *testing.T) {
	if got, want := Fingerprint(DeterministicInput(1<<12)), "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"; got != want {
		t.Fatalf("input fingerprint = %s, want %s", got, want)
	}
}

func TestEffectiveConfigProfiles(t *testing.T) {
	for _, tc := range []struct {
		name, config                  string
		logN, logSlots, slots, q0Bits int
	}{
		{name: "logn13", config: "bootstrap_config.logN13.json", logN: 13, logSlots: 12, slots: 1 << 12, q0Bits: 55},
		{name: "logn16", config: "bootstrap_config.logN16.json", logN: 16, logSlots: 15, slots: 1 << 15, q0Bits: 56},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _, err := LoadConfig(filepath.Join("..", "..", "configs", tc.config))
			if err != nil {
				t.Fatal(err)
			}
			residual, params, effective, err := ParametersFromConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if effective.LogN != tc.logN || effective.LogSlots != tc.logSlots || effective.InputSlots != tc.slots {
				t.Fatalf("unexpected effective profile: %+v", effective)
			}
			if residual.LogN() != tc.logN || len(effective.QChainBits) != params.BootstrappingParameters.MaxLevel()+1 {
				t.Fatalf("unexpected effective parameters: residual LogN=%d, profile=%+v", residual.LogN(), effective)
			}
			if effective.Q0Bits != tc.q0Bits {
				t.Fatalf("effective q0 bit size=%d, want=%d (config target=%d)", effective.Q0Bits, tc.q0Bits, cfg.Q0[0])
			}
			if effective.Q0Target != cfg.Q0[0] || len(effective.QPrimes) != len(effective.QChainBits) || len(effective.PPrimes) != len(effective.PBits) {
				t.Fatalf("effective modulus provenance is incomplete: %+v", effective)
			}
			if cfg.Mod1Degree != 30 || cfg.Mod1DoubleAngle != 3 || cfg.Mod1K != 16 || cfg.LogMessageRatio != 10 {
				t.Fatalf("config does not meet the frozen EvalMod profile: %+v", cfg)
			}
		})
	}
}
