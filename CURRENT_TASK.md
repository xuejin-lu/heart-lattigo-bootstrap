# Current Task

Task: QPREFIX-PERF-DIAG-002
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-002-MATCHED-Q55-STAGE-CLOSURE.md`

Task class:
`D — Performance Diagnosis`

Accepted parent result:
`results/QPREFIX-PERF-DIAG-001-summary.md`

Parent classification:
`UNEXPLAINED_PERFORMANCE_REGRESSION`

Historical baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Q-prefix-v2 production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Goal:
Freshly rerun the exact matched q0=55 P93 baseline and candidate, capture non-overlapping in-context Bootstrap stage timings on both immutable SHAs, and account for the observed E2E latency delta directly.

Key rule:
Do not infer the matched 3.79x regression from q0=56 stage microbenchmarks. Measure baseline and candidate under the same q0=55 workload.

Do not modify Secondary production code.
Do not optimize kernels.
Do not change parameters, schedules, or `QPrefixWidth(Level)`.
Do not reopen F/full-RNS fallback or architecture.

Write:
`results/QPREFIX-PERF-DIAG-002-summary.md`

Return one:
- `MATCHED_STAGE_ATTRIBUTION_CLOSED`
- `MATCHED_STAGE_ATTRIBUTION_UNCLOSED`
- `MATCHED_BASELINE_REPLAY_BLOCKED`

Also report:
- `TOP_DELTA_STAGE=<stage name>`
- `RESIDUAL_FRACTION=<value>`

Then report `READY_FOR_WEB_REVIEW`.
