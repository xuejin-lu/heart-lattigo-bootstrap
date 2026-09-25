# Current Task

Task: FAST-STORAGE-006
Status: READY_FOR_CODEX_DIRTY_RECOVERY

Specification:
`specs/FAST-STORAGE-006-PRIVATE-F-TRACE-FOUNDATION-SALVAGE.md`

Task class:
`I — Implementation / controlled dirty-worktree recovery`

Decision:
- FAST-INTEGRATION-003 production residency is rejected on performance.
- Do not weaken the 20% gate.
- Preserve reusable private-F scalar/normalized-Trace foundation only.
- Restore production ModUp exactly to accepted FAST-INTEGRATION-002 semantics.

Known Secondary state:
- committed HEAD remains `775e8901ffdb7496b345be852af04bd0e0b03c96`;
- FAST-INTEGRATION-003 left 7 dirty files intentionally uncommitted;
- those known 003 changes may be reconciled under the recovery authority in the spec.

Codex must inspect the dirty diff, preserve only standalone foundation work, revert rejected production wiring, run the required tests, commit/push the salvaged Secondary result, and return `READY_FOR_WEB_REVIEW`.

If any dirty change is unrelated to FAST-INTEGRATION-003, stop with `NEEDS_WEB_REVIEW`.
