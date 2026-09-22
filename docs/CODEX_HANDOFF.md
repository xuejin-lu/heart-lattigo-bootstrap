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

## Orchestrator static audit completed

Read:

- `docs/FIX-001-P3-LOGN13-P93-SCALE-STATIC-AUDIT.md`
- `results/FIX-001-P3-LOGN13-P93-SCALE-STATIC-AUDIT-summary.json`

Static classification:

`STATIC_SCALE_AUDIT_NO_CAUSAL_DEFECT_FOUND_RUNTIME_ROUNDING_CHECKS_REQUIRED`

Important source conclusions already made by the orchestrator:

- `Scale.BigInt()` rounds to nearest; it does not truncate.
- ScaleDown's rounding contract matches Standard.
- ModUp's `round(scale)` physical multiplier + floating metadata scale also exists in Genuine Standard.
- EvalMod's public DefaultScale reset and final residual DefaultScale reset match Standard exactly.
- P93 `planScale=2^93` is a planner/intermediate scale, not the public ciphertext scale.
- normalized Mod1's virtual-exponent recurrence is algebraically consistent.
- no Secondary fix is authorized from static analysis.

## Current Codex task: measurement only

`specs/FIX-001-P3-MEASURE-LOGN13-P93-SCALE-RUNTIME-FACTS.md`

Codex must only measure:

1. ScaleDown rounded-ratio mismatch.
2. ModUp exact vs Float64 ratio, rounded scalar, and semantic distortion.
3. Executed PS `addAligned/subAligned` scale ratios and `BigInt()` mismatch.
4. Executed PS `InDelta` metadata-snap magnitudes.
5. Exact normalized Mod1 scale / virtual-exponent ledger.
6. C2S compressed restore factors and semantic residuals.
7. Frozen exact-E2E control.

Codex must not:
- decide whether any result is correct;
- propose or implement a fix;
- modify Secondary.

The result must be a compact facts-only JSON ending with:
`MEASUREMENT_COMPLETE`.

## Workflow

1. User says `開始`.
2. Codex executes the measurement-only task.
3. Codex commits/pushes Primary measurements only.
4. User returns to ChatGPT and says `review`.
5. ChatGPT/orchestrator performs the correctness judgment.

## Secondary protection

- no Secondary edits
- no Secondary commit/push
- no reset/stash/clean/discard
- no parameter tuning
- no scale changes
- no precision tightening
