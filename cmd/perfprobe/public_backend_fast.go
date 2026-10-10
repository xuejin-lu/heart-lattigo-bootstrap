//go:build perf_fast

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
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
