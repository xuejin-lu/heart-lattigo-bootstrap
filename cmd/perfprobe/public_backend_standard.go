//go:build perf_standard

package main

import (
	"encoding/json"
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func publicBackendName() string { return "standard" }

func validatePublicEphemeralWeight(weight int) error {
	if weight < 0 {
		return fmt.Errorf("Standard EphemeralSecretWeight must be non-negative, got %d", weight)
	}
	return nil
}

func publicEvaluatorDispatch(eval *bootstrapping.Evaluator) (string, error) {
	if eval == nil || eval.Evaluator == nil {
		return "", fmt.Errorf("public bootstrapping.NewEvaluator did not construct the genuine Standard evaluator")
	}
	return "bootstrapping.NewEvaluator genuine Standard evaluator", nil
}

func publicMulRelinRescaleDispatchEvidence() string {
	return "ordinary ckks.Evaluator.MulRelinNew and Rescale on the isolated genuine Standard source"
}

func configurePublicBootstrapNoise(*bootstrapping.Evaluator, float64, int64) error {
	return fmt.Errorf("fixed-sigma Bootstrap injection is Fast-only")
}

func publicBootstrapNoiseSnapshot(*bootstrapping.Evaluator) json.RawMessage { return nil }

func injectPublicFixedSigmaNoise(*ckks.Evaluator, *rlwe.Ciphertext, float64, int64) (fixedSigmaNoiseEvidence, error) {
	return fixedSigmaNoiseEvidence{}, fmt.Errorf("fixed-sigma injection is Fast-only")
}
