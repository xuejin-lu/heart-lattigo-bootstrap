package perfmeasure

import (
	"math/big"
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

func TestExplicitEphemeralSecretWeightIsPreserved(t *testing.T) {
	cfg, _, err := LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, params, effective, err := ParametersFromConfigWithE(cfg, 32)
	if err != nil {
		t.Fatal(err)
	}
	if params.EphemeralSecretWeight != 32 || effective.EphemeralSecretWeight != 32 {
		t.Fatalf("explicit E was not preserved: parameters=%d effective=%d", params.EphemeralSecretWeight, effective.EphemeralSecretWeight)
	}

	if _, _, _, err := ParametersFromConfigWithE(cfg, -1); err == nil {
		t.Fatal("negative E was accepted")
	}
}

func TestBatch019PublicWorkloadUsesCanonicalInputAndCapacityGates(t *testing.T) {
	a, b, c, err := PublicMulRescaleWorkload(1 << 12)
	if err != nil {
		t.Fatal(err)
	}
	if got := Fingerprint(a); got != "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285" {
		t.Fatalf("public workload A fingerprint=%s", got)
	}
	w1, err := PublicMulRescaleWorkloadFingerprint(a, b, c, "q5")
	if err != nil {
		t.Fatal(err)
	}
	w2, err := PublicMulRescaleWorkloadFingerprint(a, b, c, "q5")
	if err != nil || w1 != w2 {
		t.Fatalf("workload fingerprint was not deterministic: %s %v", w2, err)
	}
	cfg, _, err := LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, params, _, err := ParametersFromConfigWithE(cfg, 32)
	if err != nil {
		t.Fatal(err)
	}
	workloadHash, err := PublicMulRescaleWorkloadFingerprint(a, b, c, new(big.Int).SetUint64(params.BootstrappingParameters.Q()[5]).String())
	if err != nil {
		t.Fatal(err)
	}
	if want := "00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74"; workloadHash != want {
		t.Fatalf("Batch019 workload fingerprint=%s, want %s", workloadHash, want)
	}
	plan, err := BoundPublicMulRescaleWorkload(a, b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Q0 == "" || plan.Q0123 == "" || plan.TerminalQ0Status != PublicCapacitySufficient || plan.TerminalQ0Deficit != "" || plan.MulDivisor != new(big.Int).SetUint64(params.BootstrappingParameters.Q()[5]).String() {
		t.Fatalf("incomplete Batch019 capacity plan: %+v", plan)
	}
	if _, err := BoundPublicMulRescaleWorkload(a[:1], b, c, params.BootstrappingParameters.Q(), params.BootstrappingParameters.N()); err == nil {
		t.Fatal("mismatched workload vectors were accepted")
	}
}

func TestLogN16PublicCapacityAssessmentRetainsStrictBoundWarning(t *testing.T) {
	config, _, err := LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN16.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, params, effective, err := ParametersFromConfigWithE(config, 32)
	if err != nil {
		t.Fatal(err)
	}
	a, b, c, err := PublicMulRescaleWorkload(effective.InputSlots)
	if err != nil {
		t.Fatal(err)
	}
	q := params.BootstrappingParameters.Q()
	plan, err := AssessPublicMulRescaleWorkload(a, b, c, q, params.BootstrappingParameters.N())
	if err != nil {
		t.Fatalf("Q0123 sufficient-condition assessment rejected the canonical LogN16 workload: %v", err)
	}
	if plan.Rescale != "18169085704944493" || plan.Q0 != "36028797019488257" ||
		plan.TerminalQ0Status != PublicCapacityNotProven || plan.TerminalQ0Deficit != "309374390400729" {
		t.Fatalf("LogN16 sufficient-condition warning changed: %+v", plan)
	}
	if _, err := BoundPublicMulRescaleWorkload(a, b, c, q, params.BootstrappingParameters.N()); err == nil ||
		err.Error() != "terminal Level0 q0 capacity failure: B=18169085704944493 q0=36028797019488257 deficit=309374390400729" {
		t.Fatalf("historical strict API no longer fails with the exact terminal warning: %v", err)
	}
	workloadHash, err := PublicMulRescaleWorkloadFingerprint(a, b, c, new(big.Int).SetUint64(q[5]).String())
	if err != nil {
		t.Fatal(err)
	}
	if want := "1413c0f0edc3c6b52d7a71e4714efe70280854e1208f87640b741682b65fa2a9"; workloadHash != want {
		t.Fatalf("canonical LogN16 workload SHA-256=%s, want %s", workloadHash, want)
	}
}
