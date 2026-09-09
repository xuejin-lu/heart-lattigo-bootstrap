package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

// BootstrapConfig is the experiment's single parameter source of truth.
// Fields removed from the legacy hardware/DecodeThenModUp configuration are
// intentionally not represented here; LoadBootstrapConfig rejects them.
type BootstrapConfig struct {
	LogN             int   `json:"log_n"`
	LogDefaultScale  int   `json:"log_default_scale"`
	SecretHamming    int   `json:"secret_hamming"`
	Q0               []int `json:"q0"`
	QSlotsToCoeffs   []int `json:"q_slots_to_coeffs"`
	QCoeffsToSlots   []int `json:"q_coeffs_to_slots"`
	P                []int `json:"p"`
	SlotsToCoeffsDFT []int `json:"slots_to_coeffs_dft_levels"`
	CoeffsToSlotsDFT []int `json:"coeffs_to_slots_dft_levels"`
	LogSlots         int   `json:"log_slots"`
	Mod1LogScale     int   `json:"mod1_log_scale"`
	Mod1Degree       int   `json:"mod1_degree"`
	Mod1DoubleAngle  int   `json:"mod1_double_angle"`
	Mod1K            int   `json:"mod1_k"`
	LogMessageRatio  int   `json:"log_message_ratio"`
	Mod1InvDegree    int   `json:"mod1_inv_degree"`
	Repetitions      int   `json:"repetitions"`
	Warmup           int   `json:"warmup"`
}

func DefaultBootstrapConfig() BootstrapConfig {
	return BootstrapConfig{
		LogN:             13,
		LogDefaultScale:  45,
		SecretHamming:    192,
		Q0:               []int{55},
		QSlotsToCoeffs:   []int{39, 39, 39},
		QCoeffsToSlots:   []int{56, 56, 56, 56},
		P:                []int{61, 61, 61, 61, 61},
		SlotsToCoeffsDFT: []int{1, 1, 1},
		CoeffsToSlotsDFT: []int{1, 1, 1, 1},
		LogSlots:         -1,
		Mod1LogScale:     60,
		Mod1Degree:       30,
		Mod1DoubleAngle:  3,
		Mod1K:            16,
		LogMessageRatio:  10,
		Mod1InvDegree:    0,
		Repetitions:      3,
		Warmup:           1,
	}
}

func LoadBootstrapConfig(path string) (BootstrapConfig, error) {
	cfg := DefaultBootstrapConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return BootstrapConfig{}, fmt.Errorf("read config: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return BootstrapConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func factorization(depths, scales []int) ([][]int, error) {
	if len(depths) != len(scales) || len(depths) == 0 {
		return nil, fmt.Errorf("DFT levels and prime scales must have the same non-zero length")
	}
	result := make([][]int, len(depths))
	for i, depth := range depths {
		if depth <= 0 || scales[i] <= 0 {
			return nil, fmt.Errorf("DFT depth and prime scale must be positive at index %d", i)
		}
		result[i] = make([]int, depth)
		for j := range result[i] {
			result[i][j] = scales[i]
		}
	}
	return result, nil
}

// NewBootstrapParametersFromConfig adapts the historical configuration to the
// bounded public contract shared by Standard and Fast. The first two Q primes
// form the residual ciphertext; the bootstrapping package then constructs the
// full circuit Q chain from the same DFT and EvalMod intent.
func NewBootstrapParametersFromConfig(cfg BootstrapConfig) (ckks.Parameters, bootstrapping.Parameters, error) {
	if cfg.LogN <= 0 || cfg.LogDefaultScale <= 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("log_n and log_default_scale must be positive")
	}
	if len(cfg.Q0) != 1 || len(cfg.QSlotsToCoeffs) == 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("configuration must provide q0 and at least one q_slots_to_coeffs prime")
	}
	if len(cfg.P) == 0 || cfg.SecretHamming <= 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("configuration must provide P primes and a positive secret hamming weight")
	}

	residualLogQ := []int{cfg.Q0[0], cfg.QSlotsToCoeffs[0]}
	residual, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            cfg.LogN,
		LogQ:            residualLogQ,
		LogDefaultScale: cfg.LogDefaultScale,
		Xs:              ring.Ternary{H: cfg.SecretHamming},
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("residual CKKS parameters: %w", err)
	}

	c2s, err := factorization(cfg.CoeffsToSlotsDFT, cfg.QCoeffsToSlots)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("coeffs-to-slots configuration: %w", err)
	}
	s2c, err := factorization(cfg.SlotsToCoeffsDFT, cfg.QSlotsToCoeffs)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("slots-to-coeffs configuration: %w", err)
	}

	logSlots := cfg.LogSlots
	if logSlots < 0 {
		logSlots = residual.LogMaxSlots()
	}
	if logSlots < 0 || logSlots > residual.LogMaxSlots() {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("log_slots=%d is outside [0,%d]", logSlots, residual.LogMaxSlots())
	}

	logN, evalModScale, mod1Degree, doubleAngle, k, logMessageRatio, invDegree :=
		cfg.LogN, cfg.Mod1LogScale, cfg.Mod1Degree, cfg.Mod1DoubleAngle, cfg.Mod1K, cfg.LogMessageRatio, cfg.Mod1InvDegree
	effectiveLiteral := bootstrapping.ParametersLiteral{
		LogN:     &logN,
		LogP:     append([]int(nil), cfg.P...),
		Xs:       ring.Ternary{H: cfg.SecretHamming},
		LogSlots: &logSlots,
		CoeffsToSlotsFactorizationDepthAndLogScales: c2s,
		SlotsToCoeffsFactorizationDepthAndLogScales: s2c,
		EvalModLogScale:       &evalModScale,
		EphemeralSecretWeight: intPtr(0),
		Mod1Type:              mod1.CosDiscrete,
		LogMessageRatio:       &logMessageRatio,
		K:                     &k,
		Mod1Degree:            &mod1Degree,
		DoubleAngle:           &doubleAngle,
		Mod1InvDegree:         &invDegree,
	}
	btp, err := bootstrapping.NewParametersFromLiteral(residual, effectiveLiteral)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("bootstrapping parameters: %w", err)
	}
	btp.CircuitOrder = bootstrapping.ModUpThenEncode
	return residual, btp, nil
}

func intPtr(value int) *int { return &value }
