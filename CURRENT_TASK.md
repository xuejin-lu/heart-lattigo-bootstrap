# Current Task

Task: FAST-ZERO-SECRET-E-METRICS-RECOVERY-002
Status: READY_FOR_CODEX
Task class: E — bounded evidence recovery only (two new Fast Bootstrap calls)

**Executable spec:** `specs/FAST-ZERO-SECRET-E-METRICS-RECOVERY-002.md`
**Long-term objective:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Web review of predecessor

`FAST-ZERO-SECRET-E-PASSTHROUGH-001` is accepted **only as PARTIAL**: the Secondary implementation `c0ac40b736aee441dd32a85ff6cd40b2a9be4b4e` permits E=0/32 zero-secret Fast Bootstrap and Secondary fix `463d494627b2e9e2bfac51aefe7f4ecb3493b68e` serializes exact-zero SNR. Original E0/E32 Bootstrap and decode executed, but aggregate JSON did not persist; no numerical results may be inferred. The genuine Standard E32 reference was recorded with RMSE `4.80507250213e-9`.

Primary partial evidence is already pushed at `d1e054dec3a833394091fc56d86d8e9780aade99` (verified as fast-forward from `e884658aebcf7bf7ceffc1be76d679a5cf0b84c1`).

## New bounded work

Safely synchronize both repositories. **First verify the external pinned Standard E32 reference and targeted serialization tests**. If its exact per-slot file is unavailable/unverifiable, stop without any Bootstrap; do not reconstruct it. Otherwise, under Secondary fixed `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`, run **exactly one new Fast E=0 and one new Fast E=32** public LogN13 Bootstrap using the existing diagnostic, capture JSON and report true numerical differences. Zero Standard Bootstrap, no parameter tuning, no performance benchmark, no LogN16, no production-code edits.

Write a new concise Primary result summary, preserve the old partial report. Return `FAST_E_METRICS_RECOVERED_PENDING_WEB_REVIEW`, `FAST_E_METRICS_RECOVERY_PARTIAL`, or a precise blocked classification. Then `READY_FOR_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.

`FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**.
