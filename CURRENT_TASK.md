# Current Task

Task: QPREFIX-PERF-OPT-001
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-OPT-001-TRANSACTIONAL-RESCALE-STAGING.md`

Task class:
`P — Performance Repair`

Accepted framework:
- Primary R1 summary: `3e75a8db349f7c155db82a9ee9bc94afe6d7be98`
- Secondary R1: `4783c641cee2df5f504b8033905a62a82971da28`
- status: `DIAGNOSTIC_FRAMEWORK_REPAIR_READY`

Accepted diagnosis:
- `GENERATED_POWER_RESCALE_DOMINANT`
- Rescale explains `97.7205%` of generated-power regression
- candidate preflight contributes about `9.267 ms` across twelve generated-power Rescales

Goal:
Preserve transactional failure-before-mutation while computing exact Q-prefix Rescale results only once into evaluator-owned staging, then commit only after all components validate.

Do not change:
- centered CRT / rounding semantics;
- per-step capacity checks;
- Q-prefix authority or `QPrefixWidth(Level)`;
- P93/polynomial schedule;
- NTT/Montgomery semantics;
- F/full-RNS policy.

Mandatory performance floors:
- q0=55 P93 Count-1 E2E median >=10% faster;
- representative rows4 Rescale median >=20% faster;
- no material allocation regression.

Use the reusable `scripts/fastdiag` framework for post-change structural attribution. No ad-hoc overlay.

Write:
`results/QPREFIX-PERF-OPT-001-summary.md`

Return one:
- `TRANSACTIONAL_RESCALE_STAGING_READY`
- `TRANSACTIONAL_RESCALE_STAGING_CORRECT_BUT_NO_WIN`
- `TRANSACTIONAL_RESCALE_STAGING_BLOCKED`

Then report `READY_FOR_WEB_REVIEW`.
