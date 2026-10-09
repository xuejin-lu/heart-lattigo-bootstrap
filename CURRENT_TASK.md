# Current Task

Task: FAST-DROPIN-API-LIFECYCLE-AUDIT-003
Status: READY_FOR_CODEX
Task class: bounded identical-frontend API compatibility audit (no production changes)

**Executable spec:** `specs/FAST-DROPIN-API-LIFECYCLE-AUDIT-003.md`
**Long-term architecture goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`
**Accepted numerical result:** `results/FAST-ZERO-SECRET-E-METRICS-RECOVERY-002-summary.md`
**Independent review:** `results/FAST-ZERO-SECRET-E-METRICS-RECOVERY-002-web-review.md`

## Web scientific review

`FAST-ZERO-SECRET-E-METRICS-RECOVERY-002` is **ACCEPTED as a bounded LogN13 zero-secret numerical experiment**. Secondary `463d494627b2e9e2bfac51aefe7f4ecb3493b68e` produced identical decoded Fast E0 and E32 outputs (RMSE=0); Fast E32 vs original RMSE `1.1907546782784477e-9`; Fast E32 vs native Standard E32 RMSE `4.914687044299207e-9`. Exactly two new Fast Bootstrap calls; no new Standard calls; errors and reference documented with matched source hashes.

Do not claim universal numerical accuracy, cryptographic security, or measured speedup. The Fast diagnostic input was `c1=0`, and the unchanged real frontend lifecycle has not yet been demonstrated.

## Active bounded next step

Audit whether the **same unchanged frontend program** using the Standard public CKKS API and exactly the same CKKS parameter configuration can compile and execute its keygen/encrypt/evaluator/decrypt lifecycle against pinned genuine Standard and Fast dependency source. Investigate `rlwe.NewEncryptor`, `GenEvaluationKeys`, `bootstrapping.NewEvaluator`, `rlwe.NewDecryptor` and representation invariants. First record the actual API/semantic mismatch, if any. This is a static and inexpensive compilation/runtime-preflight task; **no production Fast/Standard code changes and default zero Bootstrap calls**. Do not use the special Fast c1=0 harness as evidence of drop-in compatibility. If any uncertainty requires mathematical design, stop for Web review.

Write a compact Primary report and return `FAST_DROPIN_LIFECYCLE_AUDITED`, `FAST_DROPIN_LIFECYCLE_PARTIAL` or `FAST_DROPIN_LIFECYCLE_BLOCKED`, followed by `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.

`FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**.
