# Current Task

Task: QPREFIX-PERF-DIAG-005
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-005-LOW-OVERHEAD-POWER-ATTRIBUTION.md`

Task class:
`D — Performance Diagnosis`

Accepted parent:
- `results/QPREFIX-PERF-DIAG-004-summary.md`
- Primary commit `672ce1b578db1147ba9bb4cb9834ffe357a50af3`
- Secondary production ref `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

Important review finding:
DIAG-004's stage-only ranking is accepted, but its `96.51%` Rescale share came from all-scope deep tracing with ~81% Bootstrap overhead. Enabling `rescale` scope adds per-coefficient timer work inside the outer power/rescale span.

Goal:
Re-measure current generated-power category costs with `power` and `stage,power` only, leaving the deep `rescale` scope disabled.

No Secondary production modification.
No one-off overlays.

Write:
`results/QPREFIX-PERF-DIAG-005-summary.md`

Return:
`POST_OPT_LOW_OVERHEAD_ATTRIBUTION_READY`

plus one:
- `LOW_OVERHEAD_RESCALE_DOMINANCE_CONFIRMED`
- `LOW_OVERHEAD_RESCALE_DOMINANCE_REJECTED`
- `LOW_OVERHEAD_POWER_ATTRIBUTION_UNCLOSED`

and:
- `POWER_ONLY_OVERHEAD=<fraction>`
- `POWER_ONLY_RESCALE_SHARE=<fraction>`
- `GENERATED_POWER_SHARE_OF_EVALMOD=<fraction>`

Then report `READY_FOR_WEB_REVIEW`.
