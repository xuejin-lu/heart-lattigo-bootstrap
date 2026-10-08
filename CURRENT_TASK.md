# Current Task

Task: FAST-ZERO-SECRET-E-PASSTHROUGH-001
Status: READY_FOR_CODEX
Task class: bounded Secondary Fast implementation plus numerical smoke, followed by independent Web review

**Authoritative executable spec:** `specs/FAST-ZERO-SECRET-E-PASSTHROUGH-001.md`

**Durable user objective:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`.

## Direction and superseded task

The goal is a drop-in Fast numerical-validation backend: keep frontend source, Standard public APIs and parameters unchanged, switch only the Lattigo library version, and internally use intentionally insecure zero-secret execution with sound elimination of unnecessary KeySwitch. Q-prefix is only one optional optimization.

The previous `FAST-STANDARD-EPHEMERAL-REPLICATION-002` is **SUPERSEDED / DO NOT RUN**. No more Standard E0 key-variability experiments under that task.

## Active bounded milestone

In Secondary `xuejin-lu/lattigo` branch `fast-qprefix`, accept frontend E=0 and E=32 without modifying its public value, preserve zero-secret semantics, omit Dense/Sparse KeySwitch, and perform exactly one LogN13 Fast Bootstrap for each E. Measure actual output vs original and, if matched, genuine Standard E32 as reference. Do not assume result accuracy or speedup before seeing numbers. At most three public Bootstrap runs total, no performance benchmark, no tuning, no LogN16, no broad drop-in retrofit this task.

Do not modify the pinned Standard implementation or Primary frontend/production harness.

The old `FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**.

## Handoff

Safe-sync, implement this scope with tests, bounded self-review and permitted Secondary/Primary report commits; return measured errors, code/source evidence and `FAST_E_PASSTHROUGH_RUN_COMPLETE`, `FAST_E_PASSTHROUGH_PARTIAL` or `FAST_E_PASSTHROUGH_BLOCKED`, then `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
