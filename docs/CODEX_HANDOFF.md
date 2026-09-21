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

When the user says `開始`, follow `AGENTS.md` mandatory preflight first:

1. safely synchronize Primary `main`;
2. re-read `AGENTS.md`;
3. re-read this file;
4. re-read synchronized `CURRENT_TASK.md`;
5. read the exact referenced spec;
6. execute only that task.

`CURRENT_TASK.md` is authoritative after synchronization.

## Fixed current architecture

- LogN13
- q0 = 56-bit effective/diagnostic profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012
- planScale = `2^93`
- q3+ are not PS arithmetic sources
- Q012 -> Q01 once at PS exit
- deterministic 4096-slot workload

Do not retune parameters unless a later synchronized spec explicitly authorizes it.

## Current correctness status

The `1e-2` milestone is not end-to-end passed by current production.

Historical P93 design/reference path:
- EvalMod real ≈ `1.6351e-4`
- EvalMod imag ≈ `1.3360e-4`
- post-S2C proxy ≈ `3.0329e-4`
- public-like ≈ `9.7054e-3` (passes `1e-2`)

Important: this historical candidate did not pass every strict internal precision criterion; treat it only as a design/reference path that reached the `1e-2` system milestone.

Current dirty production evidence from commit `35b06d1c2ef67c1962f02971dcef225049fbc692`:
- EvalMod real ≈ `0.02231793676`
- EvalMod imag ≈ `0.02056928703`
- post-S2C/F0 ≈ `0.04027362692`
- official public output ≈ `0.22191425127`

The same-input S2C decomposition showed:
- Fast S2C implementation effect = `0`
- upstream EvalMod propagation accounts for the observed F0 difference
- linear-delta residual ≈ `9.29e-12`

Therefore do not repair or re-debug S2C now.

The previous diagnostic did not literally rerun the historical P93 reference; it loaded the committed artifact. The next task must first perform a fresh isolated historical P93 replay, then localize the first material reference-vs-production divergence through PS/EvalMod.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision it points to:

`specs/FIX-001-P3-DIAG-P93-EVALMOD-REFERENCE-VS-PRODUCTION-FIRST-DIVERGENCE.md`

This task is diagnostic only.

It must:
1. create isolated temporary detached reference worktrees and freshly rerun historical P93 methodology;
2. leave the existing dirty Secondary worktree untouched;
3. reproduce current production EvalMod/F0;
4. compare fresh reference vs production checkpoint-by-checkpoint;
5. report first observable and first material divergence;
6. commit/push compact Primary evidence and stop without repair.

## Two-word operating workflow

Normal workflow:

1. Codex: user says only `開始`.
2. Codex syncs Primary, reads current spec, executes, validates, commits and pushes authorized Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator independently reviews GitHub evidence and prepares the next spec/`CURRENT_TASK.md`.

Only exceptional repository/provenance/safety failures should interrupt this workflow.

## Current prohibitions

- no dirty Secondary production source modification;
- no Secondary production commit/push;
- no destructive Git operation on dirty Secondary;
- no S2C repair;
- no q/planScale tuning;
- no T2 repair merely because it is early;
- no PS repair;
- no replacement sweep;
- no finalizer repair;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.
