package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type BootstrapUsageReportOptions struct {
	ConfigPath string
	OutputPath string
	Format     string
	ByStep     bool
	Verbose    bool
}

type BootstrapUsageReport struct {
	Config     BootstrapConfig        `json:"config"`
	Parameters BootstrapParameterInfo `json:"parameters"`
	Stages     []BootstrapStageUsage  `json:"stages,omitempty"`
	Total      FunctionalUnitUsage    `json:"total"`
}

type BootstrapParameterInfo struct {
	LogN                  int   `json:"log_n"`
	N                     int   `json:"n"`
	LogMaxSlots           int   `json:"log_max_slots"`
	MaxSlots              int   `json:"max_slots"`
	LogDefaultScale       int   `json:"log_default_scale"`
	QPrimeBitLengths      []int `json:"q_prime_bit_lengths"`
	QPrimeCount           int   `json:"q_prime_count"`
	QMaxLevel             int   `json:"q_max_level"`
	QTotalBitLength       int   `json:"q_total_bit_length"`
	PPrimeBitLengths      []int `json:"p_prime_bit_lengths"`
	PPrimeCount           int   `json:"p_prime_count"`
	PMaxLevel             int   `json:"p_max_level"`
	PTotalBitLength       int   `json:"p_total_bit_length"`
	SlotsToCoeffsLevelQ   int   `json:"slots_to_coeffs_level_q"`
	SlotsToCoeffsLevelP   int   `json:"slots_to_coeffs_level_p"`
	SlotsToCoeffsDepth    int   `json:"slots_to_coeffs_depth"`
	CoeffsToSlotsLevelQ   int   `json:"coeffs_to_slots_level_q"`
	CoeffsToSlotsLevelP   int   `json:"coeffs_to_slots_level_p"`
	CoeffsToSlotsDepth    int   `json:"coeffs_to_slots_depth"`
	EvalModLevelQ         int   `json:"eval_mod_level_q"`
	EvalModLogScale       int   `json:"eval_mod_log_scale"`
	EvalModDegree         int   `json:"eval_mod_degree"`
	EvalModDoubleAngle    int   `json:"eval_mod_double_angle"`
	EvalModK              int   `json:"eval_mod_k"`
	EvalModLogMessageRate int   `json:"eval_mod_log_message_ratio"`
}

type BootstrapStageUsage struct {
	Name  string              `json:"name"`
	Usage FunctionalUnitUsage `json:"usage"`
}

func RunBootstrapUsageReport(opts BootstrapUsageReportOptions) (BootstrapUsageReport, error) {
	cfg := DefaultBootstrapConfig()
	if opts.ConfigPath != "" {
		data, err := os.ReadFile(opts.ConfigPath)
		if err != nil {
			return BootstrapUsageReport{}, fmt.Errorf("read config: %w", err)
		}
		if err = json.Unmarshal(data, &cfg); err != nil {
			return BootstrapUsageReport{}, fmt.Errorf("parse config: %w", err)
		}
	}

	params, btpParams, err := NewBootstrapParametersFromConfig(cfg)
	if err != nil {
		return BootstrapUsageReport{}, err
	}

	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()
	encoder := ckks.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)

	evk, _, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return BootstrapUsageReport{}, fmt.Errorf("generate bootstrapping keys: %w", err)
	}

	refEval, err := bootstrapping.NewEvaluator(btpParams, evk)
	if err != nil {
		return BootstrapUsageReport{}, fmt.Errorf("new bootstrapping evaluator: %w", err)
	}

	hwEval := NewSlimBootstrapperHW(refEval)
	hwEval.Verbose = opts.Verbose
	hwEval.C2S, hwEval.S2C, err = ConfigureNaiveDFTMatrices(params, encoder, kgen, sk, refEval)
	if err != nil {
		return BootstrapUsageReport{}, err
	}

	values := make([]complex128, params.MaxSlots())
	for i := 0; i < minInt(8, len(values)); i++ {
		values[i] = complex(math.Sin(float64(i+1))/4, 0)
	}

	pt := ckks.NewPlaintext(params, btpParams.SlotsToCoeffsParameters.LevelQ)
	if err = encoder.Encode(values, pt); err != nil {
		return BootstrapUsageReport{}, fmt.Errorf("encode input: %w", err)
	}

	ct, err := encryptor.EncryptNew(pt)
	if err != nil {
		return BootstrapUsageReport{}, fmt.Errorf("encrypt input: %w", err)
	}

	report := BootstrapUsageReport{
		Config:     cfg,
		Parameters: bootstrapParameterInfo(params, btpParams),
	}

	if opts.ByStep {
		var total FunctionalUnitUsage
		record := func(name string, f func() error) error {
			ResetFunctionalUnitUsage()
			err := f()
			usage := CurrentFunctionalUnitUsage()
			ClearFunctionalUnitUsage()
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			report.Stages = append(report.Stages, BootstrapStageUsage{Name: name, Usage: usage})
			total = total.Add(usage)
			return nil
		}

		if err = record("SlotsToCoeffs", func() error {
			ct, err = hwEval.SlotsToCoeffs(ct, nil)
			return err
		}); err != nil {
			return BootstrapUsageReport{}, err
		}
		if err = record("ScaleDown", func() error {
			ct, _, err = hwEval.ScaleDown(ct)
			return err
		}); err != nil {
			return BootstrapUsageReport{}, err
		}
		if err = record("ModUp", func() error {
			ct, err = hwEval.ModUp(ct)
			return err
		}); err != nil {
			return BootstrapUsageReport{}, err
		}

		var ctReal, ctImag *rlwe.Ciphertext
		if err = record("CoeffsToSlots", func() error {
			ctReal, ctImag, err = hwEval.CoeffsToSlots(ct)
			return err
		}); err != nil {
			return BootstrapUsageReport{}, err
		}
		if err = record("EvalMod(real)", func() error {
			ctReal, err = hwEval.EvalMod(ctReal)
			return err
		}); err != nil {
			return BootstrapUsageReport{}, err
		}
		if ctImag != nil {
			if err = record("EvalMod(imag)", func() error {
				ctImag, err = hwEval.EvalMod(ctImag)
				return err
			}); err != nil {
				return BootstrapUsageReport{}, err
			}
			if err = record("Recombine", func() error {
				eval := newHardwareCKKSEvaluator(hwEval)
				if err = eval.Mul(ctImag, 1i, ctImag); err != nil {
					return err
				}
				return eval.Add(ctReal, ctImag, ctReal)
			}); err != nil {
				return BootstrapUsageReport{}, err
			}
		}
		report.Total = total
	} else {
		ResetFunctionalUnitUsage()
		_, err = hwEval.Bootstrap(ct)
		report.Total = CurrentFunctionalUnitUsage()
		ClearFunctionalUnitUsage()
		if err != nil {
			return BootstrapUsageReport{}, fmt.Errorf("bootstrap: %w", err)
		}
	}

	return report, nil
}

func WriteBootstrapUsageReport(report BootstrapUsageReport, outputPath, format string) error {
	if outputPath == "" {
		return nil
	}
	if format == "" {
		format = inferReportFormat(outputPath)
	}
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(outputPath, append(data, '\n'), 0o644)
	case "csv":
		file, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		return writeUsageCSV(file, report)
	default:
		return fmt.Errorf("unsupported report format %q", format)
	}
}

func PrintBootstrapUsageReport(report BootstrapUsageReport) {
	fmt.Printf("Parameters: LogN=%d N=%d slots=%d Q=%d bits (%d primes) P=%d bits (%d primes)\n",
		report.Parameters.LogN,
		report.Parameters.N,
		report.Parameters.MaxSlots,
		report.Parameters.QTotalBitLength,
		report.Parameters.QPrimeCount,
		report.Parameters.PTotalBitLength,
		report.Parameters.PPrimeCount)
	if len(report.Stages) > 0 {
		for _, stage := range report.Stages {
			fmt.Printf("%-16s %s\n", stage.Name+":", stage.Usage.String())
		}
	}
	fmt.Printf("%-16s %s\n", "TOTAL:", report.Total.String())
}

func ConfigureNaiveDFTMatrices(
	params ckks.Parameters,
	encoder *ckks.Encoder,
	kgen *rlwe.KeyGenerator,
	sk *rlwe.SecretKey,
	eval *bootstrapping.Evaluator,
) (dft.Matrix, dft.Matrix, error) {
	c2s := eval.CoeffsToSlotsParameters
	c2s.LogBSGSRatio = -1
	s2c := eval.SlotsToCoeffsParameters
	s2c.LogBSGSRatio = -1

	c2sMatrix, err := dft.NewMatrixFromLiteral(params, c2s, encoder)
	if err != nil {
		return dft.Matrix{}, dft.Matrix{}, fmt.Errorf("naive C2S matrix: %w", err)
	}
	s2cMatrix, err := dft.NewMatrixFromLiteral(params, s2c, encoder)
	if err != nil {
		return dft.Matrix{}, dft.Matrix{}, fmt.Errorf("naive S2C matrix: %w", err)
	}

	addKeysForMatrix := func(mat dft.Matrix) {
		for _, m := range mat.Matrices {
			slots := 1 << m.LogDimensions.Cols
			for k := range m.Vec {
				rot := k & (slots - 1)
				if rot == 0 {
					continue
				}
				galEl := params.GaloisElement(rot)
				if _, err := eval.MemEvaluationKeySet.GetGaloisKey(galEl); err == nil {
					continue
				}
				eval.MemEvaluationKeySet.GaloisKeys[galEl] = kgen.GenGaloisKeyNew(galEl, sk)
			}
		}
	}

	addKeysForMatrix(c2sMatrix)
	addKeysForMatrix(s2cMatrix)
	return c2sMatrix, s2cMatrix, nil
}

func bootstrapParameterInfo(params ckks.Parameters, btp bootstrapping.Parameters) BootstrapParameterInfo {
	q := params.Q()
	p := params.P()
	return BootstrapParameterInfo{
		LogN:                  params.LogN(),
		N:                     params.N(),
		LogMaxSlots:           params.LogMaxSlots(),
		MaxSlots:              params.MaxSlots(),
		LogDefaultScale:       int(params.DefaultScale().Log2()),
		QPrimeBitLengths:      bitLengths(q),
		QPrimeCount:           len(q),
		QMaxLevel:             params.MaxLevelQ(),
		QTotalBitLength:       productBitLength(q),
		PPrimeBitLengths:      bitLengths(p),
		PPrimeCount:           len(p),
		PMaxLevel:             params.MaxLevelP(),
		PTotalBitLength:       productBitLength(p),
		SlotsToCoeffsLevelQ:   btp.SlotsToCoeffsParameters.LevelQ,
		SlotsToCoeffsLevelP:   btp.SlotsToCoeffsParameters.LevelP,
		SlotsToCoeffsDepth:    btp.SlotsToCoeffsParameters.Depth(false),
		CoeffsToSlotsLevelQ:   btp.CoeffsToSlotsParameters.LevelQ,
		CoeffsToSlotsLevelP:   btp.CoeffsToSlotsParameters.LevelP,
		CoeffsToSlotsDepth:    btp.CoeffsToSlotsParameters.Depth(true),
		EvalModLevelQ:         btp.Mod1ParametersLiteral.LevelQ,
		EvalModLogScale:       btp.Mod1ParametersLiteral.LogScale,
		EvalModDegree:         btp.Mod1ParametersLiteral.Mod1Degree,
		EvalModDoubleAngle:    btp.Mod1ParametersLiteral.DoubleAngle,
		EvalModK:              btp.Mod1ParametersLiteral.K,
		EvalModLogMessageRate: btp.Mod1ParametersLiteral.LogMessageRatio,
	}
}

func productBitLength(moduli []uint64) int {
	x := big.NewInt(1)
	for _, qi := range moduli {
		x.Mul(x, new(big.Int).SetUint64(qi))
	}
	return x.BitLen()
}

func bitLengths(moduli []uint64) []int {
	out := make([]int, len(moduli))
	for i, qi := range moduli {
		out[i] = new(big.Int).SetUint64(qi).BitLen()
	}
	return out
}

func inferReportFormat(path string) string {
	switch filepath.Ext(path) {
	case ".csv":
		return "csv"
	default:
		return "json"
	}
}

func writeUsageCSV(file *os.File, report BootstrapUsageReport) error {
	w := csv.NewWriter(file)
	defer w.Flush()
	if err := w.Write([]string{
		"stage", "ntt_forward", "ntt_inverse", "ntt_total", "base_conversion",
		"automorphism", "double_prime_scaling", "ewu_tensor", "ewu_acc_q",
		"ewu_acc_p", "ewu_mod_d", "ewu_mad", "ewu_total", "double_rns_add", "double_rns_mul",
	}); err != nil {
		return err
	}

	writeRow := func(stage string, u FunctionalUnitUsage) error {
		return w.Write([]string{
			stage,
			strconv.FormatUint(u.NTTForward, 10),
			strconv.FormatUint(u.NTTInverse, 10),
			strconv.FormatUint(u.NTTTotal(), 10),
			strconv.FormatUint(u.BaseConversion, 10),
			strconv.FormatUint(u.Automorphism, 10),
			strconv.FormatUint(u.DoublePrimeScale, 10),
			strconv.FormatUint(u.ElementWise[OpTensor], 10),
			strconv.FormatUint(u.ElementWise[OpAccQ], 10),
			strconv.FormatUint(u.ElementWise[OpAccP], 10),
			strconv.FormatUint(u.ElementWise[OpModD], 10),
			strconv.FormatUint(u.ElementWise[OpMAD], 10),
			strconv.FormatUint(u.ElementWiseTotal(), 10),
			strconv.FormatUint(u.DoubleRNSAdd, 10),
			strconv.FormatUint(u.DoubleRNSMul, 10),
		})
	}

	for _, stage := range report.Stages {
		if err := writeRow(stage.Name, stage.Usage); err != nil {
			return err
		}
	}
	return writeRow("TOTAL", report.Total)
}
