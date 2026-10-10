//go:build !perf_fast && !perf_standard

package main

import (
	"errors"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func newBackend(bootstrapping.Parameters, ckks.Parameters) (backendAdapter, error) {
	return nil, errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}

func newInputBackend(bootstrapping.Parameters, ckks.Parameters) (backendAdapter, error) {
	return nil, errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}

func publicBackendName() string { return "unselected" }

func validatePublicEphemeralWeight(int) error {
	return errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}

func publicEvaluatorDispatch(*bootstrapping.Evaluator) (string, error) {
	return "", errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}
