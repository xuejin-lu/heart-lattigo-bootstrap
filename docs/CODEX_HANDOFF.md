# Fast-CKKS Codex Handoff

This is the current operational handoff for starting a fresh Codex chat.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch: `main`

Secondary:
- `xuejin-lu/lattigo`
- active branch: `fast-ckks`
- committed base: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- local worktree is intentionally dirty with the current fixed-width Q012 / PS-wide P93 production candidate.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the Secondary dirty worktree.

## Startup rule

When the user says `開始`, follow `AGENTS.md` mandatory preflight first, then read synchronized `CURRENT_TASK.md` and its exact spec. `CURRENT_TASK.md` is authoritative.

## Current accepted evidence

The `1e-2` milestone is not yet exact-E2E passed.

Two separate metadata bugs are causally proven:

1. public Fast EvalMod resets Scale from internal/input-restored `2^50` to residual DefaultScale `2^45`;
2. downstream Fast public finalization resets corrected Scale back to residual DefaultScale.

Diagnostic-only metadata preservation gives:
- corrected EvalMod real ≈ `0.00418301958`
- corrected EvalMod imag ≈ `0.00426156114`
- corrected post-S2C/F0 vs matched diagnostic Standard ≈ `0.00693482035`

But even preserving both proven metadata corrections through final decode gives exact E2E ≈ `0.18857544537` vs original message, still far above `1e-2`.

Therefore the remaining blocker is not explained by the two factor-32 metadata resets.

A critical reference issue must now be resolved:
- matched diagnostic Standard core suggests local/proxy error ≈ `0.0069`;
- exact Fast corrected E2E vs original message is ≈ `0.1886`;
- Genuine Standard historical exact E2E is around `5.83e-8`.

Do not treat the matched diagnostic Standard core as an exact E2E oracle until its lineage is reconciled with a fresh Genuine Standard full-bootstrap run.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-GENUINE-STANDARD-VS-MATCHED-REFERENCE-RECONCILIATION.md`

The task must:
1. freshly run Genuine Standard public Bootstrap and a staged Standard pipeline from the same deterministic input;
2. establish their exact E2E baseline and agreement;
3. reconstruct the exact matched diagnostic Standard reference lineage;
4. compare matched vs Genuine Standard at C2S, EvalMod, S2C/core and final-equivalent stages where valid;
5. explicitly distinguish stage/proxy metrics from exact E2E;
6. compare the metadata-corrected Fast branch against Genuine Standard;
7. identify the first valid material Fast-vs-Genuine-Standard divergence;
8. stop after classification; no production repair.

## Two-word workflow

1. Codex: user says only `開始`.
2. Codex synchronizes Primary, executes current spec, validates, commits and pushes Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator reviews and prepares the next task.

Only exceptional repository/provenance/safety failures interrupt this workflow.

## Current prohibitions

- no dirty Secondary source modification
- no Secondary commit/push
- no destructive Git operation on dirty Secondary
- no new Fast production repair
- no coefficient correction
- no q/planScale tuning
- no PS/C2S/S2C/finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
