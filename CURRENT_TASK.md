# Current Task

Task: DIAG-FRAMEWORK-001-R1
Status: READY_FOR_CODEX

Specification:
`specs/DIAG-FRAMEWORK-001-R1-RESCALE-EVENT-CLOSURE.md`

Task class:
`I — Diagnostic Infrastructure Repair`

Parent implementation:
- Primary `9662cd15d1a8c54e3ee618731f6f397fd45ca213`
- Primary summary `f832b8ed1788a7b59f29dde9919e5eee3bcf774c`
- Secondary `6312e8b9a982a405709a5124041eec569b0ae980`

Independent Web review result:
The reusable framework is broadly sound, but Rescale subphase tracing currently emits duplicate same-semantic reconstruction events and can double-count deepest child closure.

Goal:
Repair only the diagnostic event hierarchy so one physical interval is represented once at each accounting level and nested closure is mathematically unambiguous.

Do not modify Rescale arithmetic.
Do not remove/fuse preflight.
Do not optimize production performance.
Do not change Q-prefix policy, generated-power schedule, or F/full-RNS architecture.

Write:
`results/DIAG-FRAMEWORK-001-R1-summary.md`

Return one:
- `DIAGNOSTIC_FRAMEWORK_REPAIR_READY`
- `DIAGNOSTIC_FRAMEWORK_REPAIR_FAILED`

Then report `READY_FOR_WEB_REVIEW`.
