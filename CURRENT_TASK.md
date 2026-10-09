# Current Task

Task: FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005
Status: READY_FOR_CODEX
Task class: bounded unchanged-frontend CKKS primitive API / semantic audit (no production code changes)

**Executable spec:** `specs/FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005.md`
**Long-term objective:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`
**Accepted prior stage:** `results/FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004-web-review.md`

## Web review of predecessor

`FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004` is accepted for **one specified secret-key EncryptNew → Bootstrap → DecryptNew numeric lifecycle** using one *identical frontend source* with library-only dependency replacement (LogN13, E32). Secondary `eb7793f1e449702590cdba3d9603261623354f80`, Primary result `19ebcb888cd4e806f40df24e785ac75ec5d85d8e`. Fast output vs original RMSE `1.1907546782784477e-9`; Fast vs genuine Standard RMSE `4.821647086585007e-9`. This is not evidence for general CKKS Add/MulRelin/Rotate/Rescale APIs, public-key EncryptNew, overall speedup or real encryption security.

## Current bounded task

Audit **the same ordinary public CKKS primitive frontend code** compiled against both genuine Standard and Fast dependencies. Source-check `ckks.NewEvaluator` and Fast key paths; test Add/MulRelin/Rescale/Rotate only within sound representation boundaries with plaintext numeric oracles. Do not silently run Standard KeySwitch on Fast-layout evaluation keys. Stop and report the first mismatch; **no production code modification, no Bootstrap, no benchmarking**. Write compact Primary report and return `FAST_DROPIN_PRIMITIVE_AUDITED`, `FAST_DROPIN_PRIMITIVE_PARTIAL` or `FAST_DROPIN_PRIMITIVE_BLOCKED` plus Web review status.

`FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.
