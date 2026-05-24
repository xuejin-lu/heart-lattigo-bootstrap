package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

// BootstrapConfig is the single place to set the hardware bootstrapping test
// parameters used by tests and by the functional-unit usage report.
type BootstrapConfig struct {
	LogN             int   `json:"log_n"`
	LogDefaultScale  int   `json:"log_default_scale"`
	SecretHamming    int   `json:"secret_hamming"`
	Q0               []int `json:"q0"`
	QSlotsToCoeffs   []int `json:"q_slots_to_coeffs"`
	QCircuitSlots    []int `json:"q_circuit_slots"`
	QEvalMod         []int `json:"q_eval_mod"`
	QCoeffsToSlots   []int `json:"q_coeffs_to_slots"`
	P                []int `json:"p"`
	SlotsToCoeffsDFT []int `json:"slots_to_coeffs_dft_levels"`
	CoeffsToSlotsDFT []int `json:"coeffs_to_slots_dft_levels"`
	LogBSGSRatio     int   `json:"log_bsgs_ratio"`
	LogSlots         int   `json:"log_slots"`
	Mod1LogScale     int   `json:"mod1_log_scale"`
	Mod1Degree       int   `json:"mod1_degree"`
	Mod1DoubleAngle  int   `json:"mod1_double_angle"`
	Mod1K            int   `json:"mod1_k"`
	LogMessageRatio  int   `json:"log_message_ratio"`
	Mod1InvDegree    int   `json:"mod1_inv_degree"`
}

func DefaultBootstrapConfig() BootstrapConfig {
	return BootstrapConfig{
		LogN:             13,
		LogDefaultScale:  45,
		SecretHamming:    192,
		Q0:               []int{55},
		QSlotsToCoeffs:   []int{39, 39, 39},
		QCircuitSlots:    []int{45},
		QEvalMod:         []int{60, 60, 60, 60, 60, 60, 60, 60},
		QCoeffsToSlots:   []int{56, 56, 56, 56},
		P:                []int{61, 61, 61, 61, 61},
		SlotsToCoeffsDFT: []int{1, 1, 1},
		CoeffsToSlotsDFT: []int{1, 1, 1, 1},
		LogBSGSRatio:     1,
		LogSlots:         -1,
		Mod1LogScale:     60,
		Mod1Degree:       30,
		Mod1DoubleAngle:  3,
		Mod1K:            16,
		LogMessageRatio:  10,
		Mod1InvDegree:    0,
	}
}

func NewBootstrapParametersFromConfig(cfg BootstrapConfig) (ckks.Parameters, bootstrapping.Parameters, error) {
	if cfg.LogN <= 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("LogN must be positive")
	}
	if cfg.LogDefaultScale <= 0 {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("LogDefaultScale must be positive")
	}
	if len(cfg.QCircuitSlots) == 0 {
		cfg.QCircuitSlots = []int{cfg.LogDefaultScale}
	}

	logQ := append([]int{}, cfg.Q0...)
	logQ = append(logQ, cfg.QSlotsToCoeffs...)
	logQ = append(logQ, cfg.QCircuitSlots...)
	logQ = append(logQ, cfg.QEvalMod...)
	logQ = append(logQ, cfg.QCoeffsToSlots...)

	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            cfg.LogN,
		LogQ:            logQ,
		LogP:            cfg.P,
		LogDefaultScale: cfg.LogDefaultScale,
		Xs:              ring.Ternary{H: cfg.SecretHamming},
	})
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, fmt.Errorf("ckks parameters: %w", err)
	}

	logSlots := cfg.LogSlots
	if logSlots < 0 {
		logSlots = params.LogMaxSlots()
	}

	coeffsToSlots := dft.MatrixLiteral{
		Type:         dft.HomomorphicEncode,
		Format:       dft.RepackImagAsReal,
		LogSlots:     logSlots,
		LevelQ:       params.MaxLevelQ(),
		LevelP:       params.MaxLevelP(),
		LogBSGSRatio: cfg.LogBSGSRatio,
		Levels:       append([]int{}, cfg.CoeffsToSlotsDFT...),
	}

	mod1Params := mod1.ParametersLiteral{
		LevelQ:          params.MaxLevel() - coeffsToSlots.Depth(true),
		LogScale:        cfg.Mod1LogScale,
		Mod1Type:        mod1.CosDiscrete,
		Mod1Degree:      cfg.Mod1Degree,
		DoubleAngle:     cfg.Mod1DoubleAngle,
		K:               cfg.Mod1K,
		LogMessageRatio: cfg.LogMessageRatio,
		Mod1InvDegree:   cfg.Mod1InvDegree,
	}

	slotsToCoeffs := dft.MatrixLiteral{
		Type:         dft.HomomorphicDecode,
		LogSlots:     logSlots,
		LogBSGSRatio: cfg.LogBSGSRatio,
		LevelP:       params.MaxLevelP(),
		Levels:       append([]int{}, cfg.SlotsToCoeffsDFT...),
	}
	slotsToCoeffs.LevelQ = len(slotsToCoeffs.Levels)

	return params, bootstrapping.Parameters{
		ResidualParameters:      params,
		BootstrappingParameters: params,
		SlotsToCoeffsParameters: slotsToCoeffs,
		Mod1ParametersLiteral:   mod1Params,
		CoeffsToSlotsParameters: coeffsToSlots,
		EphemeralSecretWeight:   0,
		CircuitOrder:            bootstrapping.DecodeThenModUp,
	}, nil
}
