# Current Task

Task: QPREFIX-PERF-DIAG-006
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-006-RESCALE-KERNEL-ATTRIBUTION.md`

Task class:
`D — Performance Diagnosis`

Accepted parent:
- `results/QPREFIX-PERF-DIAG-005-summary.md`
- decision `LOW_OVERHEAD_RESCALE_DOMINANCE_CONFIRMED`
- current production ref `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

Accepted current facts:
- power-only overhead about 3.73%;
- generated-power Rescale share about 93.74%;
- generated powers about 60.53% of combined EvalMod.

Goal:
Attribute the remaining rows4 Q-prefix Rescale cost to production kernels using diagnostics-off focused benchmarks plus CPU profiling, without per-coefficient tracing.

No Secondary production arithmetic changes.

Write:
`results/QPREFIX-PERF-DIAG-006-summary.md`

Return:
`RESCALE_KERNEL_ATTRIBUTION_READY`

plus:
- classification;
- `TOP_RESCALE_KERNEL=...`
- `TOP_RESCALE_SHARE=...`
- `ROWS4_ROWS2_FULL_RATIO=...`
- `PHASE_CLOSURE_RATIO=...`

Then report `READY_FOR_WEB_REVIEW`.
