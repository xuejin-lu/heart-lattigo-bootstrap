//go:build !perf_fast && !perf_standard

package main

import (
	"errors"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func decodeDualCurrentStage(_ *publicBackend, _ string, _ *rlwe.Ciphertext) ([]complex128, string, error) {
	return nil, "", errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}

func auditDualCurrentC2SGroup(_ *publicBackend, _ *rlwe.Ciphertext) (dualCurrentC2SAudit, error) {
	return dualCurrentC2SAudit{}, errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}
