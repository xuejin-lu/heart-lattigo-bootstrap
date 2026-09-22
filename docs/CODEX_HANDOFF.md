# Fast-CKKS Codex Handoff

## Repositories

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

Do not rewrite or modify the finalized candidate during the current audit.

## Frozen milestone baseline

The LogN13/P93/Q012 `1e-2` milestone is complete and frozen.

Authoritative Fast exact E2E:
`0.009705381393898434`

Genuine Standard control:
`5.830057349387463e-8`

Do not start precision tightening.

## Current task

`specs/FIX-001-P3-AUDIT-LOGN13-P93-SCALE-SEMANTICS-END-TO-END.md`

This is a post-milestone integrity audit of scaling-factor semantics only.

It must audit the full scale lineage and distinguish:

- metadata Scale
- physical coefficient scalar
- CKKS Rescale divisor
- Mod1 ScalingFactor
- targetScale
- P93 planScale
- virtual/deferred scalar
- integer scale-alignment ratios
- DFT restore powers
- public DefaultScale contracts

No Secondary modification is authorized.

## High-risk audit points

Mandatory checks include:

1. ModUp:
   physical scalar uses `round(scale)` while metadata multiplies the original floating `scale`;
   quantify exact mismatch and semantic impact.

2. polynomial add/sub alignment:
   `Scale ratio -> BigInt()`;
   audit integerization error for every executed alignment.

3. `InDelta` metadata snapping:
   quantify every metadata-only equality assignment after approximate scale comparison.

4. Mod1:
   keep `ScalingFactor`, `targetScale`, `planScale=2^93`, working scale, virtual exponent, coherent scale, DA scale, and public scale logically separate.

5. DFT:
   verify every group Rescale and restore-plan power.

6. public boundary:
   verify EvalMod/public-final DefaultScale resets against Genuine Standard contract.

## Required artifacts

- `results/FIX-001-P3-AUDIT-LOGN13-P93-SCALE-SEMANTICS-END-TO-END-summary.json`
- `docs/FIX-001-P3-LOGN13-P93-SCALE-LEDGER.md`

No long narrative; the ledger should be compact and table-driven.

## Workflow

1. User says `開始` to Codex.
2. Codex syncs Primary and executes the audit.
3. Secondary must remain exactly at `40532b4d...` and clean.
4. Codex validates and commits/pushes Primary audit evidence only.
5. User returns here and says `review`.
6. Orchestrator independently reviews warnings/failures before any later phase.

## Prohibitions

- no Secondary modification
- no scale fix
- no precision tightening
- no parameter tuning
- no q/planScale changes
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
