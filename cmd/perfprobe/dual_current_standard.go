//go:build perf_standard

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func auditDualCurrentC2SGroup(backend *publicBackend, modUp *rlwe.Ciphertext) (dualCurrentC2SAudit, error) {
	if backend == nil || backend.bootstrapEval == nil || backend.bootstrapEval.DFTEvaluator == nil || modUp == nil {
		return dualCurrentC2SAudit{}, fmt.Errorf("genuine Standard C2S observer is incomplete")
	}
	matrix := backend.bootstrapEval.C2SDFTMatrix
	if len(matrix.Levels) == 0 || len(matrix.Matrices) < 2 || matrix.Levels[0] < 1 || len(matrix.Matrices) <= matrix.Levels[0] {
		return dualCurrentC2SAudit{}, fmt.Errorf("C2S matrix lacks a first group followed by a next matrix")
	}
	q := backend.full.Q()
	if modUp.Level() != backend.full.MaxLevel() || modUp.Level()+1 != len(q) {
		return dualCurrentC2SAudit{}, fmt.Errorf("ModUp level=%d active rows=%d, expected full Standard Q level=%d rows=%d", modUp.Level(), modUp.Level()+1, backend.full.MaxLevel(), len(q))
	}
	if backend.full.LogN() == 16 && (len(q) != 16 || modUp.Level() != 15) {
		return dualCurrentC2SAudit{}, fmt.Errorf("LogN16 genuine Standard full-Q preflight expected q0..q15 at level 15; got QCount=%d Level=%d", len(q), modUp.Level())
	}
	before, err := dualCurrentAuditShape(backend, "mod_up_before_c2s_group0", modUp)
	if err != nil {
		return dualCurrentC2SAudit{}, err
	}
	groupSize := matrix.Levels[0]
	out := ckks.NewCiphertext(backend.full, 1, matrix.LevelQ)
	current := modUp
	for idx := 0; idx < groupSize; idx++ {
		if err := backend.bootstrapEval.DFTEvaluator.LTEvaluator.Evaluate(current, matrix.Matrices[idx], out); err != nil {
			return dualCurrentC2SAudit{}, fmt.Errorf("first C2S group matrix %d: %w", idx, err)
		}
		current = out
	}
	if err := backend.bootstrapEval.Rescale(out, out); err != nil {
		return dualCurrentC2SAudit{}, fmt.Errorf("first C2S group Rescale: %w", err)
	}
	if out.Level() != modUp.Level()-1 {
		return dualCurrentC2SAudit{}, fmt.Errorf("first C2S group Rescale level=%d, expected %d", out.Level(), modUp.Level()-1)
	}
	afterRescale, err := dualCurrentAuditShape(backend, "after_c2s_group0_rescale", out)
	if err != nil {
		return dualCurrentC2SAudit{}, err
	}
	if err := backend.bootstrapEval.DFTEvaluator.LTEvaluator.Evaluate(out, matrix.Matrices[groupSize], out); err != nil {
		return dualCurrentC2SAudit{}, fmt.Errorf("next C2S matrix %d after group Rescale: %w", groupSize, err)
	}
	afterNext, err := dualCurrentAuditShape(backend, "after_next_c2s_matrix", out)
	if err != nil {
		return dualCurrentC2SAudit{}, err
	}
	return dualCurrentC2SAudit{Applicable: true, MatrixGroupSize: groupSize, Before: before, AfterRescale: afterRescale, AfterNextMatrix: afterNext}, nil
}
