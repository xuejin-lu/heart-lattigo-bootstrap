package perfmeasure

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
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

func TestLoadConfigRejectsRemovedLegacyFields(t *testing.T) {
	for _, field := range []string{"q_circuit_slots", "q_eval_mod", "log_bsgs_ratio"} {
		t.Run(field, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(`{"`+field+`":1}`), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := LoadConfig(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
				t.Fatalf("LoadConfig error = %v, want unknown legacy field %q", err, field)
			}
		})
	}
}

func TestChangingConfiguredPrimeScaleChangesEffectiveModulus(t *testing.T) {
	cfg, _, err := LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, base, _, err := ParametersFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	changed := cfg
	changed.QCoeffsToSlots = append([]int(nil), cfg.QCoeffsToSlots...)
	changed.QCoeffsToSlots[0]--
	_, modified, _, err := ParametersFromConfig(changed)
	if err != nil {
		t.Fatal(err)
	}
	baseBits := base.BootstrappingParameters.LogQi()
	changedBits := modified.BootstrappingParameters.LogQi()
	const firstCoeffsToSlotsIndex = 2 + 3 + 8
	if len(baseBits) != len(changedBits) || baseBits[firstCoeffsToSlotsIndex] == changedBits[firstCoeffsToSlotsIndex] {
		t.Fatalf("configured C2S prime scale did not alter its effective Q prime: base=%v changed=%v", baseBits, changedBits)
	}
}

func TestSharedParameterBuilderDerivesLogN13Through16FromEffectiveParameters(t *testing.T) {
	cfg, _, err := LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}

	for logN := 13; logN <= 16; logN++ {
		t.Run(fmt.Sprintf("logn%d", logN), func(t *testing.T) {
			profileConfig := cfg
			profileConfig.LogN = logN
			profileConfig.LogSlots = -1

			residual, bootstrap, effective, err := ParametersFromConfigWithE(profileConfig, 32)
			if err != nil {
				t.Fatalf("shared parameter construction rejected LogN%d: %v", logN, err)
			}
			if effective.LogN != bootstrap.BootstrappingParameters.LogN() || effective.LogN != logN ||
				effective.LogSlots != bootstrap.CoeffsToSlotsParameters.LogSlots || effective.LogSlots != logN-1 ||
				effective.InputSlots != 1<<effective.LogSlots || effective.RingN != 1<<logN ||
				effective.ResidualRingN != residual.N() || effective.EphemeralSecretWeight != 32 {
				t.Fatalf("effective profile is not derived from constructed parameters: %+v", effective)
			}
			if len(effective.QPrimes) != len(bootstrap.BootstrappingParameters.Q()) ||
				len(effective.PPrimes) != len(bootstrap.BootstrappingParameters.P()) ||
				len(effective.QChainBits) != len(effective.QPrimes) || len(effective.PBits) != len(effective.PPrimes) {
				t.Fatalf("effective Q/P metadata does not describe the constructed profile: %+v", effective)
			}
			defaultScale := residual.DefaultScale()
			if effective.DefaultScale != defaultScale.Value.Text('e', 80) {
				t.Fatalf("effective scale=%s, constructed residual scale=%s", effective.DefaultScale, defaultScale.Value.Text('e', 80))
			}
			if got, err := EffectiveProfileName(effective); err != nil || got != fmt.Sprintf("logn%d", logN) {
				t.Fatalf("effective profile name=%q err=%v", got, err)
			}
			if got, err := ResolveProfileName("", effective); err != nil || got != fmt.Sprintf("logn%d", logN) {
				t.Fatalf("empty profile label did not resolve from effective parameters: %q err=%v", got, err)
			}
			if _, err := ResolveProfileName("logn13", effective); logN != 13 && err == nil {
				t.Fatalf("mismatched profile label was accepted for effective LogN%d", logN)
			}
			q, p := bootstrap.BootstrappingParameters.Q(), bootstrap.BootstrappingParameters.P()
			qpHash := QPFingerprint(q, p)
			if qpHash == "" || qpHash != QPFingerprint(q, p) {
				t.Fatalf("Q/P fingerprint is not deterministic: %q", qpHash)
			}
			reorderedQ := append([]uint64(nil), q...)
			reorderedQ[0], reorderedQ[1] = reorderedQ[1], reorderedQ[0]
			if QPFingerprint(reorderedQ, p) == qpHash {
				t.Fatal("Q/P fingerprint did not bind prime order")
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

func TestPublicFixedSigmaChainFingerprintBindsBothRescaleScales(t *testing.T) {
	a, b, c, err := PublicMulRescaleWorkload(8)
	if err != nil {
		t.Fatal(err)
	}
	base, err := PublicFixedSigmaChainFingerprint(a, b, c, "35184372088832", "35000000000001", "35000000000003")
	if err != nil {
		t.Fatal(err)
	}
	changedQ4, err := PublicFixedSigmaChainFingerprint(a, b, c, "35184372088832", "35000000000001", "35000000000005")
	if err != nil {
		t.Fatal(err)
	}
	changedQ5, err := PublicFixedSigmaChainFingerprint(a, b, c, "35184372088832", "35000000000007", "35000000000003")
	if err != nil {
		t.Fatal(err)
	}
	if base == changedQ4 || base == changedQ5 || changedQ4 == changedQ5 {
		t.Fatal("fixed-sigma chain fingerprint did not bind both effective divisors")
	}
	if _, err := PublicFixedSigmaChainFingerprint(a, b, c, "", "1", "2"); err == nil {
		t.Fatal("fixed-sigma chain fingerprint accepted a missing default Scale")
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
