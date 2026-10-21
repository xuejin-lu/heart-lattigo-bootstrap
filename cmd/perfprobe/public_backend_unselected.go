//go:build !perf_fast && !perf_standard

package main

import (
	"errors"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func publicMulRelinRescaleDispatchEvidence() string {
	return "backend build tag not selected"
}

func injectPublicFixedSigmaNoise(*ckks.Evaluator, *rlwe.Ciphertext, float64, int64) (fixedSigmaNoiseEvidence, error) {
	return fixedSigmaNoiseEvidence{}, errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}
