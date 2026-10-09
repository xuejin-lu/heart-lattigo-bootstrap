# Current Task

Task: FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A
Status: READY_FOR_CODEX
Task class: narrowly bounded I repair after independent scientific review

**Executable spec:** `specs/FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A.md`
**Review evidence:** `results/FAST-DROPIN-KEYLESS-MULRELIN-006-web-review.md`
**Long-term goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Web decision

`FAST-DROPIN-KEYLESS-MULRELIN-006` is accepted **only as PARTIAL**, notwithstanding passing numeric metrics. Secondary `6ce15cf8b949c8a89f610bcca5ab506dd3f560f8` implements a real keyless public CKKS MulRelin with c1=0, full active Q backing, degree-one inputs, Scale/Level preservation and an independent no-relin-key test. Fixed same-source Fast MulRelin RMSE `5.320587702687035e-15` (Standard `4.853e-14`). The original 006 implementation **does not fail closed when c1 is nonzero but normal RelinearizationKey is available**: its helper returns `handled=false` and generic native KeySwitch may run. This violates the original accepted zero-secret input contract.

## Current targeted repair

For Fast-marked `ckks.Evaluator.MulRelin/MulRelinNew` ciphertext/ciphertext operations only, reject any nonzero c1, unsupported degrees/representation, missing active Q backing or other unsupported inputs **without mutation** and **regardless of whether Standard-layout relin keys exist**. Do not silently execute the native Standard KeySwitch as a fallback. Keep the existing passing keyless path and unchanged frontend API/config. The main negative test must give the Fast evaluator a real native-layout key to prove fallback is impossible. No edits to Standard or separate Fast Q-prefix kernels, no Bootstrap or benchmark.

Follow `AGENTS.md` safe sync, execute focused tests and one bounded self-review/repair pass, then commit/push authorized source and concise Primary summary. Return `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_PARTIAL` or `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_BLOCKED` and Web review status. `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.
