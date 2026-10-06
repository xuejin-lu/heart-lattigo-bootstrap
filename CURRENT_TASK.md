# Current Task

Task: FAST-STANDARD-PERF-REBASELINE-001
Status: READY_FOR_CODEX
Task class: M — Matched Performance Measurement

Authoritative spec:
`specs/FAST-STANDARD-PERF-REBASELINE-001-LOGN13.md`

Completed and accepted:
- `FAST-STANDARD-NUMERICAL-FIX-001` — `FAST_STANDARD_NUMERICAL_CLOSE`
- `FAST-STANDARD-NUMERICAL-REPORT-001` — `CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`
- Primary report fix: `fd759d96144d20c4725febb367a8828b8c069e08`
- Fast Secondary: `5117fc57949647182f476dc5952099c706b9f869`

Next goal:
Benchmark corrected Fast vs genuine Standard **LogN13/q0=55** under an identical frozen measurement harness and effective parameters, preserving numerical checks and recording raw stage/full timing. Do not infer anything from historical pre-fix LogN16 timings. No production arithmetic changes permitted.

Required terminal token: `LOGN13_CORRECTED_PERF_REBASELINE_READY`, `LOGN13_CORRECTED_PERF_REBASELINE_BLOCKED`, or `LOGN13_CORRECTED_PERF_REBASELINE_NUMERICAL_FAIL`

Then `READY_FOR_WEB_REVIEW`.
