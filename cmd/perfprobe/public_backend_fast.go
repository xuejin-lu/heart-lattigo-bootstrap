//go:build perf_fast

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func publicBackendName() string { return "fast" }

func validatePublicEphemeralWeight(weight int) error {
	if weight != 0 && weight != 32 {
		return fmt.Errorf("pinned Fast source supports explicit EphemeralSecretWeight 0 or 32, got %d", weight)
	}
	return nil
}

func publicEvaluatorDispatch(eval *bootstrapping.Evaluator) (string, error) {
	if eval == nil || eval.Evaluator != nil {
		return "", fmt.Errorf("public bootstrapping.NewEvaluator did not select the pinned Fast dispatch")
	}
	return "bootstrapping.NewEvaluator public wrapper; Fast dispatch selected by pinned zero-secret capability", nil
}

func publicMulRelinRescaleDispatchEvidence() string {
	return "ordinary ckks.Evaluator.MulRelinNew -> Fast zero-secret MulCore.ApplyRows; ordinary ckks.Evaluator.Rescale -> shared Q-prefix RescaleCore.ApplyRows; no Standard fallback"
}

func injectPublicFixedSigmaNoise(evaluator *ckks.Evaluator, ct *rlwe.Ciphertext, sigma float64, seed int64) (fixedSigmaNoiseEvidence, error) {
	result, err := evaluator.InjectFixedSigmaCoefficientNoiseDiagnostic(ct, sigma, seed)
	evidence := fixedSigmaNoiseEvidence{
		SigmaCoefficient:           result.Sigma,
		SamplingSeed:               result.Seed,
		CoefficientSampleCount:     result.SampleCount,
		CoefficientMean:            result.CoefficientMean,
		CoefficientVariance:        result.CoefficientVariance,
		CoefficientRMS:             result.CoefficientRMS,
		CoefficientMaxAbs:          result.CoefficientMaxAbs,
		CoefficientSHA256:          result.CoefficientSHA256,
		SamplingAlgorithm:          result.SamplingAlgorithm,
		QPrefixRows:                result.Capacity.Rows,
		QPrefixCheckedCoefficients: result.Capacity.CheckedCoefficients,
		QPrefixMaxAbs:              result.Capacity.MaxAbs,
		QPrefixPrefixProduct:       result.Capacity.PrefixProduct,
		QPrefixStrictFit:           result.Capacity.StrictFit,
		DispatchEvidence:           result.DispatchEvidence,
	}
	return evidence, err
}
