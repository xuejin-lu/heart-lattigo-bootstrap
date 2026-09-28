# Current Task

Task: QPREFIX-PERF-DIAG-004
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-004-POST-OPT-HOTSPOT-REFRESH.md`

Task class:
`D — Performance Diagnosis`

Accepted parent:
- Primary summary `results/QPREFIX-PERF-OPT-001-summary.md`
- Secondary implementation `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`
- classification `TRANSACTIONAL_RESCALE_STAGING_READY`

Goal:
Refresh the current bottleneck after the successful Rescale staging repair. The old pre-opt Rescale-dominant conclusion is no longer authoritative for the post-opt candidate.

Use only the reusable fastdiag framework for supported tracing/compare. No ad-hoc overlays and no Secondary production changes.

Required outputs:
- current diagnostics-off q0=55 P93 E2E;
- stage-only current ranking and closure;
- current power/rescale structural trace;
- pre-opt vs post-opt fastdiag comparison;
- fresh current bottleneck classification.

Write:
`results/QPREFIX-PERF-DIAG-004-summary.md`

Return:
`POST_OPT_HOTSPOT_REFRESHED`

plus:
- `CURRENT_TOP_STAGE=<stage>`
- `CURRENT_POWER_TOP_CATEGORY=<category>`
- `CURRENT_BOTTLENECK_CLASS=<classification>`

Then report `READY_FOR_WEB_REVIEW`.
