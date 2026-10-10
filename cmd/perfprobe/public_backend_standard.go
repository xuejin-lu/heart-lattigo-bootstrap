//go:build perf_standard

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
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
