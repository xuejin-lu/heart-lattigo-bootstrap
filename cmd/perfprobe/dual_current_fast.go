//go:build perf_fast

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func decodeDualCurrentStage(backend *publicBackend, name string, ct *rlwe.Ciphertext) ([]complex128, string, error) {
	if backend == nil {
		return nil, "", fmt.Errorf("Fast stage decoder is unavailable")
	}
	if name == "bootstrap_output" {
		name = "final_public_output"
	}
	adapter := fastBackend{residual: backend.residual, bootstrappingParams: backend.full}
	return adapter.DecodeStage(ct, name)
}

func auditDualCurrentC2SGroup(_ *publicBackend, _ *rlwe.Ciphertext) (dualCurrentC2SAudit, error) {
	return dualCurrentC2SAudit{
		Applicable: false,
		Reason:     "The pinned Fast source does not expose genuine-Standard C2S linear-transform internals; the public CoeffsToSlots replay is recorded independently.",
	}, nil
}
