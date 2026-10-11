//go:build perf_fast

package main

import "github.com/tuneinsight/lattigo/v6/core/rlwe"

func auditDualCurrentC2SGroup(_ *publicBackend, _ *rlwe.Ciphertext) (dualCurrentC2SAudit, error) {
	return dualCurrentC2SAudit{
		Applicable: false,
		Reason:     "The pinned Fast source does not expose genuine-Standard C2S linear-transform internals; the public CoeffsToSlots replay is recorded independently.",
	}, nil
}
