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
- PS-wide Q012
- planScale = `2^93`
- deterministic 4096-slot workload

No tuning unless a later synchronized spec explicitly authorizes it.

## Current accepted evidence

The `1e-2` production milestone is not yet end-to-end passed.

Fresh P93 reference:
- EvalMod real ≈ `1.6351e-4`
- EvalMod imag ≈ `1.3360e-4`
- post-S2C ≈ `3.0329e-4`
- public-like ≈ `9.7054e-3`

C2S is excluded:
- production ModUp vs controlled ModUp = `0`
- actual production Fast C2S vs controlled Fast C2S = `0`
- ordinary Standard C2S differs only ~`1e-13`.

Actual-vs-forced P93 Fast EvalMod diagnostic:
- identical C2S inputs proven by semantics, metadata and active q0/q1/q2 row hashes;
- effective production planScale is also `2^93`;
- PS generated powers/rows align;
- all three DoubleAngle rounds align;
- internal final scale reset aligns exactly at input scale `2^50`;
- then public wrapper changes only Scale metadata to residual DefaultScale `2^45`;
- q0/q1/q2 rows remain equal;
- direct decoded Fast-vs-Fast gap becomes:
  - real ≈ `0.01824300375`
  - imag ≈ `0.01657172301`.
- scale ratio is exactly `2^50 / 2^45 = 32`.

Therefore the broad prior classification `P93_PRODUCTION_POST_PS_DOUBLEANGLE_DIVERGENCE` should be refined: the supported first material boundary is the public EvalMod scale metadata reset, not PS or DoubleAngle arithmetic.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-EVALMOD-PUBLIC-SCALE-RESET-CAUSAL-PROOF.md`

The task must:
1. reproduce the metadata-only `2^50 -> 2^45` public EvalMod reset;
2. prove rows stay unchanged;
3. on diagnostic clones only, restore Scale to `2^50`;
4. prove corrected actual output matches forced P93;
5. replay full downstream S2C/unpack/finalization with only that metadata correction;
6. determine whether the full current LogN13 production pipeline returns to <= `1e-2`;
7. audit the exact source assignment but do not repair Secondary.

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
- no coefficient correction
- no q/planScale tuning
- no PS/C2S/S2C/finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

Only diagnostic clone metadata Scale correction is allowed in the current task.
