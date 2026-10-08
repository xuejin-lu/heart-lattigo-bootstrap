# Current Task

Task: FAST-STANDARD-EPHEMERAL-ABLATION-001
Status: READY_FOR_CODEX
Task class: E / I — controlled native Standard ephemeral-secret ablation (diagnostic only)

**Authoritative executable spec:** `specs/FAST-STANDARD-EPHEMERAL-ABLATION-001.md`

## Web review of predecessor

`FAST-STANDARD-OUTPUT-PREFLIGHT-001` mechanically completed four single-operation Bootstrap runs (Primary harness `59899ef084380a4fc01d2582798e5113aac2409c`, evidence `eebd68a15eb34f0fa90da06b4f642f94def0d33c`). **Standard genuine-encryption output correctness is not accepted:** RMSE LogN13 `2.195800`, LogN16 `10.348459`. Fast simulation showed smaller errors but is intentionally insecure and not security-equivalent. Formal numerical quality is still `UNASSESSED`.

## Now: one controlled scientific question

With pinned unmodified genuine Standard source, the exact same Standard secret, native-encrypted ciphertext, workload, all Q/P primes, and K=16, compare `EphemeralSecretWeight=0` to `32` in a **diagnostic-only** LogN13 pair. Only if this discriminates should the analogous LogN16 pair run. No Fast execution, no sweeping, no formal comparison, no timing benchmark, and no alteration of the frozen official configs.

The original `FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**. Do not resume it.

## Handoff

Implement the small opt-in Primary diagnostic as specified, test, report the exact numerical contrast, self-review and commit/push Primary when safe; return the classification and `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
