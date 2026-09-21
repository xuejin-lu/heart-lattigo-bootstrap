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

When the user says `開始`, follow `AGENTS.md` mandatory preflight first, then read synchronized `CURRENT_TASK.md` and the exact referenced spec. `CURRENT_TASK.md` is authoritative after synchronization.

## Fixed architecture

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012
- planScale = `2^93`
- q3+ are not PS arithmetic sources
- Q012 -> Q01 once at PS exit
- deterministic 4096-slot workload

No parameter tuning unless a later synchronized spec explicitly authorizes it.

## Current evidence

The `1e-2` production milestone is still not end-to-end passed.

Fresh historical P93 reference replay succeeded:
- EvalMod real ≈ `1.6351e-4`
- EvalMod imag ≈ `1.3360e-4`
- post-S2C proxy ≈ `3.0329e-4`
- public-like ≈ `9.7054e-3`

Reference-vs-production P93 PS/DoubleAngle trace found no material arithmetic divergence:
- T2 observable difference ~ `6.94e-9`
- T16 ~ `9.1e-7`
- PS final ~ `5.0e-7`
- DoubleAngle differences ~ `1e-14`
- classification: `P93_NO_MATERIAL_REFERENCE_PRODUCTION_DIVERGENCE`

But two current production EvalMod measurements differ by path:
- actual full production core: real `0.02231793676`, imag `0.02056928703`
- controlled matched-C2S diagnostic input: real `0.00418301958`, imag `0.00426156114`

This means the next blocker is before EvalMod arithmetic, most likely at the actual production CoeffsToSlots/reference boundary. Do not repair PS, EvalMod, S2C, or finalization now.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-PRODUCTION-C2S-FIRST-MATERIAL-DIVERGENCE.md`

The task must:
1. reproduce both actual-production and controlled-diagnostic C2S-to-EvalMod paths;
2. compare ModUp and C2S boundaries directly;
3. source-faithfully trace actual production C2S group-by-group if inputs align;
4. prove downstream EvalMod delta attribution with vectors;
5. reconcile aligned vs ordinary Standard reference semantics;
6. stop after one classification; no repair.

## Two-word workflow

1. Codex: user says only `開始`.
2. Codex syncs Primary, executes current spec, validates, commits and pushes authorized Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator reviews evidence and prepares the next task.

Only exceptional repository/provenance/safety failures should interrupt this workflow.

## Current prohibitions

- no dirty Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no q/planScale tuning
- no PS repair
- no EvalMod repair
- no S2C repair
- no finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
