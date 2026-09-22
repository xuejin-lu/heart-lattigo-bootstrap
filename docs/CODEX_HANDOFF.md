# Fast-CKKS Codex Handoff

## Repository state

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- authoritative committed SHA:
  `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- local/remote synchronized
- worktree must remain clean.

## Frozen milestone

The LogN13/P93/Q012 `1e-2` exact-E2E milestone is complete and frozen.

Fast exact E2E:
`0.009705381393898434`

Genuine Standard:
`5.830057349387463e-8`

No precision-tightening phase has started.

## Current state: no Codex action

The previously created scale-audit spec:

`specs/FIX-001-P3-AUDIT-LOGN13-P93-SCALE-SEMANTICS-END-TO-END.md`

is preserved as an audit design/decision record but is **superseded as an executable Codex task**.

ChatGPT/orchestrator is now performing the static and mathematical scale-semantics audit directly from committed source and accepted evidence.

Do not run that spec when the user says `開始`.

A future Codex task may be created only for narrowly scoped runtime measurements that cannot be established from static source review. Such a task must be explicitly routed in `CURRENT_TASK.md`.

## Secondary protection

Until a new explicit task exists:

- no Secondary edits
- no Secondary commit/push
- no reset/stash/clean/discard
- no parameter tuning
- no scale changes
- no precision tightening
