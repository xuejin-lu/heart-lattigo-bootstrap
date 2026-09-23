# Current Task

Task: FAST-OBS-001
Status: READY_FOR_CODEX

Specification:
`specs/FAST-OBS-001-UNIFIED-FAST-STATE-INVARIANT-MEASUREMENT.md`

Purpose:
Build the unified Fast observability/oracle infrastructure before widened-storage implementation.

Constraints:
- Primary only.
- Secondary is read-only.
- No Fast arithmetic semantic changes.
- No 60/60/60 production implementation.
- No precision tuning.

Secondary constitution authority:
`xuejin-lu/lattigo@89e71b2bce3a4b063343cbea9178828a32afbcf4`
`docs/FAST_CKKS_SPEC.md`

When Codex is started, it should sync safely, read AGENTS.md, this task pointer, the exact spec, and the Secondary constitution, then implement/test/commit/push the Primary task.
