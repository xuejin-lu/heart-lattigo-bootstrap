# Current Task

Task: QPREFIX-PERF-DIAG-007
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-007-POST-OPT002-RESCALE-REPROFILE.md`

Task class:
`D — Performance Diagnosis`

Accepted parent:
- `results/QPREFIX-PERF-OPT-002-summary.md`
- Secondary implementation `6930cf6cb3c71ce139a1eb42eede7be335b7174c`
- classification `FIXED_WIDTH_DIVISION_DEDUP_READY`

Accepted production effect:
- rows4 fixed-width phase 12.23% faster;
- full rows4 Rescale 6.75% faster;
- q0=55 P93 Count-1 6.11% faster;
- allocation count stable.

Important:
DIAG-006 CPU-profile shares are pre-OPT002 and are no longer authoritative.

Goal:
Freshly reprofile the current rows4 Rescale and determine whether remaining cost is now division-dominated, transform-dominated, or mixed.

No Secondary production modification.

Write:
`results/QPREFIX-PERF-DIAG-007-summary.md`

Return:
`POST_OPT002_RESCALE_REPROFILE_READY`

plus:
- classification;
- `CURRENT_TOP_FLAT_SYMBOL=...`
- `CURRENT_TOP_FLAT_SHARE=...`
- `CURRENT_FIXEDWIDTH_SHARE=...`
- `CURRENT_TRANSFORM_SHARE=...`
- `CURRENT_PHASE_CLOSURE=...`

Then report `READY_FOR_WEB_REVIEW`.
