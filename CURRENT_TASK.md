# Current Task

Task: FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003
Status: READY_FOR_CODEX
Task class: I — direct Secondary cleanup and numerical regression

**Authoritative executable spec:** `specs/FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003.md`

Direct user order: **delete the entire legacy balanced/pre-scale generated-power mechanism** from Secondary `xuejin-lu/lattigo` `fast-qprefix`. No exploratory A/B, no new workaround, no residual disabled branch/dead helper or tests asserting an obsolete algorithm.

Remove `balancedScheduleFor`, `balancedFactorPair`, `integerSqrt`, factor/scale constants, scratch fields/helpers, operand-side integer-scaling/rescale branch and stale tests. One direct source-level multiply for all generated powers; Chebyshev recurrence correction before the **single** result Rescale; Monomial internal math/metadata preserved. Keep strict Q-prefix capacity, row contraction and Standard-equivalent Level/Scale.

Baseline corrected Secondary: `75ef5dbe7bbf7d3947fb2b9fb232c4a56f05c948`. Primary `cmd/fastdiag` pin and compact summary must reflect resulting Secondary SHA. LogN13 and LogN16 numerical regression/tests before committing result. Normal safe Secondary and Primary push after tests only.

**Supersedes/suspends** `FAST-STANDARD-PERF-REBASELINE-003` until this deletion is independently accepted; do NOT run 7x perf yet. Report `BALANCED_PRESCALE_REMOVED` or concrete `BALANCED_PRESCALE_REMOVAL_BLOCKED`, then `READY_FOR_WEB_REVIEW`.
