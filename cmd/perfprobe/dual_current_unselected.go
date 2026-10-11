//go:build !perf_fast && !perf_standard

package main

import (
	"errors"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func auditDualCurrentC2SGroup(_ *publicBackend, _ *rlwe.Ciphertext) (dualCurrentC2SAudit, error) {
	return dualCurrentC2SAudit{}, errors.New("build perfprobe with exactly one of -tags=perf_fast or -tags=perf_standard")
}
