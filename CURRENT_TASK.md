# Current Task

Task: QPREFIX-PERF-DIAG-003
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-003-GENERATED-POWER-RESCALE-CAUSALITY.md`

Task class:
`D — Performance Diagnosis`

Accepted parent result:
`results/QPREFIX-PERF-DIAG-002-summary.md`

Parent classification:
`MATCHED_STAGE_ATTRIBUTION_CLOSED`

Parent facts:
- `TOP_DELTA_STAGE=EvalMod real`
- `RESIDUAL_FRACTION=0.003343`
- EvalMod real delta: `+34.329583 ms`
- generated-power delta inside EvalMod real: `+20.998209 ms`

Historical q0=55 baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Q-prefix-v2 production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Goal:
Attribute the matched q0=55 generated-power slowdown, especially the contribution of Rescale and the candidate's two-pass preflight/materialization implementation.

Do not modify Secondary production code.
Do not optimize Rescale or change transactional semantics.
Do not change parameters, schedules, generated-power DAG, or `QPrefixWidth(Level)`.
Do not reopen F/full-RNS fallback.

Write:
`results/QPREFIX-PERF-DIAG-003-summary.md`

Return one:
- `GENERATED_POWER_RESCALE_DOMINANT`
- `GENERATED_POWER_MUL_RELIN_DOMINANT`
- `GENERATED_POWER_MIXED_COST`
- `GENERATED_POWER_CAUSALITY_UNCLOSED`

Also report:
- `RESCALE_DELTA_SHARE=<fraction>`
- `PREFLIGHT_TIME_SHARE_OF_POWER_DELTA=<fraction>`
- `GENERATED_POWER_RESIDUAL_FRACTION=<fraction>`

Then report `READY_FOR_WEB_REVIEW`.
