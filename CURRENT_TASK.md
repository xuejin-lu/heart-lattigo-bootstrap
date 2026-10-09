# Current Task

Task: FAST-DROPIN-KEYLESS-ROTATE-007
Status: READY_FOR_CODEX
Task class: narrowly bounded M-approved Fast zero-secret Rotate implementation and no-Bootstrap numerical preflight

**Executable spec:** `specs/FAST-DROPIN-KEYLESS-ROTATE-007.md`
**Independent prior review:** `results/FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A-web-review.md`
**Long-term goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Accepted predecessor

`FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A` is **ACCEPTED**. Secondary `93ab7ecc5a864fd9bbacea55f0af730941552691` now unconditionally rejects unsupported Fast ciphertext/ciphertext MulRelin, including nonzero c1 even with a valid Standard-layout relinearization key. Tests verify zero relin key lookups and transactional rejection. The valid keyless MulRelin path is unchanged and its fixed-profile Fast RMSE was approximately `5.32e-15`. The full standalone numeric regression preceded a final nil-input guard; final Secondary committed source passed focused package tests. Primary report `23ee62d01116d84a2575981d126db1cecca07068`.

## Current scope

Add **one** transparent Fast CKKS zero-secret `ckks.Evaluator.Rotate/RotateNew` branch, using an ordinary full-active-Q ring automorphism on `(c0,0)` without Galois key lookup or native KeySwitch. Preserve the same public API, signature, frontend, parameters, Level/Scale/metadata and all active Q rows. Reject invalid inputs and compact/missing Q backing without mutation; do not silently fall back to native Standard rotation or use only q0/q1. In-place aliasing, k=0/+1/-1, two valid levels, and independent plaintext slot oracles are required.

Only Fast Secondary implementation and tests; same unchanged frontend for one Fast numerical regression; reuse pinned genuine Standard results. **Zero Bootstrap, zero benchmarks, no performance claims or security claims**. Do not expand to RotateHoisted, Conjugate, generic Automorphism, optimized Rescale, public-key EncryptNew or Q-prefix compact inputs.

Read and execute the referenced spec after AGENTS safe sync. Report `FAST_DROPIN_KEYLESS_ROTATE_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYLESS_ROTATE_PARTIAL`, or `FAST_DROPIN_KEYLESS_ROTATE_BLOCKED`, plus `READY_FOR_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.

`FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.
