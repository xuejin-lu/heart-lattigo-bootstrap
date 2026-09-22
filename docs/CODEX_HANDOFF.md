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
- worktree clean.

## Frozen milestone

The LogN13/P93/Q012 `1e-2` exact-E2E milestone remains complete and frozen.

Fast exact E2E:
`0.009705381393898434`

Genuine Standard:
`5.830057349387463e-8`

No precision-tightening phase has started.

## Scaling-factor audit complete

Static/mathematical audit:

- `docs/FIX-001-P3-LOGN13-P93-SCALE-STATIC-AUDIT.md`
- `results/FIX-001-P3-LOGN13-P93-SCALE-STATIC-AUDIT-summary.json`

Runtime facts:

- `results/FIX-001-P3-MEASURE-LOGN13-P93-SCALE-RUNTIME-FACTS-summary.json`

Final audit:

- `docs/FIX-001-P3-LOGN13-P93-SCALE-AUDIT-FINAL.md`
- `results/FIX-001-P3-LOGN13-P93-SCALE-AUDIT-FINAL-summary.json`

Final classification:

`SCALE_AUDIT_ALL_CONTRACTS_VERIFIED`

No causal scaling-factor defect was found.

Specifically ruled out as material scale causes:

- ScaleDown rounding
- ModUp Float64/round scalar selection
- PS BigInt alignment
- PS metadata snapping
- C2S restore factors
- normalized Mod1 virtual exponent selection
- public DefaultScale resets

No Secondary scaling repair is authorized by this audit.

## Test-suite maintenance debt

Primary full `go test ./...` currently encounters:

`TestFIX001P3GenuineStandardPublicVsStagedConsistency`

with historical classification:

`P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`

This is not a current Standard numerical failure.

The historical runner has an obsolete provenance guard requiring the old dirty Secondary handoff state before it executes the Standard comparison. The finalized Secondary is now committed at `40532b4d...` and clean.

Treat this as Primary diagnostic maintenance debt, separate from scale correctness.

## Current task

None.

Status:

`SCALE_AUDIT_COMPLETE`

No Codex action is pending.

## Protection

Until an explicit new phase is started:

- no Secondary edits
- no parameter tuning
- no scale changes
- no precision tightening
- no threshold relaxation
