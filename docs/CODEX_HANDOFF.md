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

## Fixed architecture

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- intended PS-wide Q012
- intended planScale = `2^93`
- deterministic 4096-slot workload

No parameter tuning unless a later synchronized spec explicitly authorizes it.

## Current accepted evidence

The `1e-2` production milestone is not end-to-end passed.

Fresh historical P93 reference replay:
- EvalMod real ≈ `1.6351e-4`
- EvalMod imag ≈ `1.3360e-4`
- post-S2C proxy ≈ `3.0329e-4`
- public-like ≈ `9.7054e-3`

P93 reference vs production forced-path arithmetic:
- T2 difference ~ `6.94e-9`
- T16 ~ `9.1e-7`
- PS final ~ `5.0e-7`
- DoubleAngle ~ `1e-14`
- no material reference-vs-production divergence.

Latest C2S diagnostic:
- production ModUp vs controlled ModUp = `0`
- actual production Fast C2S vs controlled Fast C2S = `0`
- Fast C2S vs aligned Standard = `0`
- ordinary Standard C2S differs only ~ `1e-13`

Therefore C2S is not the blocker.

Do **not** trust the latest task's final classification `P93_STANDARD_REFERENCE_SEMANTICS_MISMATCH` as authoritative because its R0 replay expected `0.0223/0.0206` but its ordinary-Standard comparison measured about `0.1339/0.1364` and the runner failed to stop on that mismatch.

The strongest current clue is a direct Fast-vs-Fast difference from the same C2S input:
- actual `FastEvaluator.EvalMod` vs forced explicit P93 path:
  - real ≈ `0.01824300375`
  - imag ≈ `0.01657172301`

The next task must compare those two Fast EvalMod paths directly, without using Standard as the primary boundary.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-ACTUAL-EVALMOD-VS-FORCED-P93-PATH.md`

The task must:
1. prove identical C2S input for both Fast EvalMod paths;
2. reproduce direct actual-vs-forced-P93 Fast output gap;
3. audit the effective public `FastEvaluator.EvalMod` configuration/planScale read-only;
4. stage-align the two Fast paths;
5. test only the exact effective production configuration against forced P93;
6. stop after classification; no repair.

## Two-word workflow

1. Codex: user says only `開始`.
2. Codex synchronizes Primary, executes current spec, validates, commits and pushes Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator reviews and prepares the next task.

Only exceptional repository/provenance/safety failures interrupt this workflow.

## Current prohibitions

- no dirty Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no C2S repair
- no PS repair
- no EvalMod repair
- no S2C repair
- no finalizer repair
- no q tuning
- no open-ended scale sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
