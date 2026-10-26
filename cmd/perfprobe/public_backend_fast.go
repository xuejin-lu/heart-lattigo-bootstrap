//go:build perf_fast

package main

import (
	"encoding/json"
	"fmt"
	"reflect"

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

func configurePublicBootstrapNoise(evaluator *bootstrapping.Evaluator, sigma float64, rootSeed int64) error {
	if evaluator == nil {
		return fmt.Errorf("fixed-sigma Bootstrap hook requires an initialized public evaluator")
	}
	return evaluator.EnableFixedSigmaBootstrapNoiseDiagnostic(sigma, rootSeed)
}

func publicBootstrapNoiseSnapshot(evaluator *bootstrapping.Evaluator) json.RawMessage {
	if evaluator == nil {
		return nil
	}
	encoded, err := json.Marshal(evaluator.FixedSigmaBootstrapNoiseDiagnosticSnapshot())
	if err != nil {
		return nil
	}
	return encoded
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
		PhaseMeasurements:          fixedSigmaNoisePhaseMeasurements(result),
	}
	return evidence, err
}

func fixedSigmaNoisePhaseMeasurements(evidence any) []fixedSigmaNoisePhaseEvidence {
	value := reflect.ValueOf(evidence)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil
	}
	phases := value.FieldByName("Phases")
	if !phases.IsValid() || phases.Kind() != reflect.Slice {
		return nil
	}
	measurements := make([]fixedSigmaNoisePhaseEvidence, phases.Len())
	for index := 0; index < phases.Len(); index++ {
		phase := phases.Index(index)
		if phase.Kind() != reflect.Struct {
			return nil
		}
		name, elapsed, allocBytes, allocs := phase.FieldByName("Name"), phase.FieldByName("ElapsedNS"), phase.FieldByName("AllocBytes"), phase.FieldByName("Allocs")
		if !name.IsValid() || !elapsed.IsValid() || !allocBytes.IsValid() || !allocs.IsValid() ||
			name.Kind() != reflect.String || elapsed.Kind() != reflect.Int64 ||
			allocBytes.Kind() != reflect.Uint64 || allocs.Kind() != reflect.Uint64 {
			return nil
		}
		measurements[index] = fixedSigmaNoisePhaseEvidence{
			Name: name.String(), ElapsedNS: elapsed.Int(),
			AllocBytes: allocBytes.Uint(), Allocs: allocs.Uint(),
		}
	}
	return measurements
}
