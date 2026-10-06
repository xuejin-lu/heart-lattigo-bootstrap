# Current Task

Task: FAST-STANDARD-PERF-REBASELINE-002
Status: READY_FOR_CODEX
Task class: M — Matched Performance + Numerical (SNR)

**Authoritative executable spec**:
`specs/FAST-STANDARD-PERF-REBASELINE-002-LOGN13-LOGN16.md`

Supersedes for execution:
`specs/FAST-STANDARD-PERF-REBASELINE-001-LOGN13.md` (historical only).

Accepted numerical repair:
- Fast Secondary production `5117fc57949647182f476dc5952099c706b9f869`
- accepted LogN13 `FAST_STANDARD_NUMERICAL_CLOSE`; numerical report stage classifier `CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`

**MANDATORY BOTH profiles:** LogN13 (LogSlots12 / q0=55) AND LogN16 (LogSlots15 / effective q0=verify), each with **fresh** matched Standard/Fast latency, full Bootstrap decoded-domain SNR, complex RMSE, precision bits, stage timing / numerical checkpoint diagnosis, memory allocations, actual effective parameter/input fingerprint and truthful capacity/semantic limits. Two separate reports plus joint comparison.

Reuse/extend Primary `cmd/fastdiag/numerical.go` and `internal/numericalmetrics/snr.go`, do not replace the canonical SNR definition. LogN16 is NOT already supported by the current hardcoded p93-q55 diagnostic: create a profile-aware Primary-only extension. Keep all numerical computations **outside** timed Bootstrap. Do not modify Secondary production.

Historical pre-fix LogN16 6.671× speedup is not current result.

Return status:
`DUAL_LOGN13_LOGN16_PERF_SNR_READY`,
`DUAL_LOGN13_LOGN16_PERF_SNR_PARTIAL`,
`DUAL_LOGN13_LOGN16_PERF_SNR_NUMERICAL_FAIL`, or
`DUAL_LOGN13_LOGN16_PERF_SNR_BLOCKED`.

Then `READY_FOR_WEB_REVIEW`.
