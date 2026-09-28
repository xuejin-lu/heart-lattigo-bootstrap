# Current Task

Task: DIAG-FRAMEWORK-001
Status: READY_FOR_CODEX

Specification:
`specs/DIAG-FRAMEWORK-001-REUSABLE-TRACE.md`

Task class:
`I — Diagnostic Infrastructure`

Accepted parent result:
`results/QPREFIX-PERF-DIAG-003-summary.md`

Parent classification:
`GENERATED_POWER_RESCALE_DOMINANT`

Accepted facts:
- matched generated-power Rescale delta share: `0.977205`
- candidate preflight time share of power delta: `0.443635`
- repeated temporary overlay instrumentation has now been used successfully at stage, power, and Rescale depth.

Goal:
Convert the proven stage/power/rescale diagnostic patterns into a reusable, opt-in, compile-time-gated framework so future debugging does not require rebuilding one-off timing overlays.

This task does not optimize Rescale.

Required user-facing workflow:
- trace current checkout;
- compare two Secondary refs;
- select `stage`, `power`, `rescale` scopes;
- structured JSON plus concise Markdown summary.

Do not change:
- Q-prefix arithmetic;
- transactional Rescale semantics;
- `QPrefixWidth(Level)`;
- P93 schedule;
- F/full-RNS policy.

Write:
`results/DIAG-FRAMEWORK-001-summary.md`

Return one:
- `REUSABLE_DIAGNOSTIC_FRAMEWORK_READY`
- `DIAGNOSTIC_FRAMEWORK_NEEDS_REPAIR`
- `DIAGNOSTIC_FRAMEWORK_BLOCKED`

Then report `READY_FOR_WEB_REVIEW`.
