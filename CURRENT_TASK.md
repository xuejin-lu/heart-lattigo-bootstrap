# Current Task

Task: FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011
Status: READY_FOR_CODEX
Class: bounded E integration and frontend evaluation-key pairing fix only

**Executable spec:** `specs/FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011.md`
**Web review:** `results/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010-web-review.md`
**Original goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## 010 decision and root cause

Task 010 is **BLOCKED with useful Fast preflight evidence**, not an indictment of Fast arithmetic or Standard Rotate. Fast public keyless Level0 Rotate passed a 4096-slot complex oracle with RMSE `7.44997342660252e-13`, zero c1, and no GaloisKey lookups. **No Bootstrap calls were made**, and no Standard-vs-Fast comparison exists. Standard panicked because the 010 test runner passed **Bootstrap-parameter-generated Galois keys (P Level 4)** into an evaluator built from the P-free **residual** parameters (P Level -1). The Galois element's presence alone does not imply Q/P compatibility. Codex's Fast reflection-harness fix required one additional preflight run; provenance is documented.

## Authorized corrective task

Create **new identical-source 011 Primary frontend** to pair `ckks.NewEvaluator(btpParams.BootstrappingParameters, keys.MemEvaluationKeySet)` with the existing Bootstrap-generated keys in **both** pinned Standard and Fast builds, on the original residual-generated Level0 ciphertext. Preflight Q-prefix equality, ring degree, Standard-ring type, same-secret extension in pinned Standard code, exact config/input/Q/P hashes and unchanged Scale/Level. Only if both no-Bootstrap preflights pass may the run execute at most one E32 Bootstrap per backend on the rotated ciphertext. Log all results, including early failure/panic, before deciding any library repair. **Do not alter either Lattigo library or existing 004/010 evidence.**

Pinned genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
Pinned Fast Secondary: `00ac70ba136d190fa31bbb26c2f51d003a221634`.
Old performance task `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.

Return `FAST_DROPIN_KEYPLAN_COMPOSITION_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYPLAN_COMPOSITION_PARTIAL`, or `FAST_DROPIN_KEYPLAN_COMPOSITION_BLOCKED` plus `READY_FOR_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.
