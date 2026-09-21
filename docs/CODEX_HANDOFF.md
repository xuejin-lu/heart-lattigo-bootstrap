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

Fresh Genuine Standard is now authoritative:
- public Bootstrap exact E2E = `5.830057349387463e-8`;
- staged Standard equals public Standard exactly for the deterministic workload.

Matched diagnostic Standard:
- C2S real/imag exactly match Genuine Standard;
- from EvalMod onward, matched objects use a manually normalized stage oracle;
- matched EvalMod rows equal Genuine Standard EvalMod rows, but matched/internal Scale is `2^50` while Genuine Standard public EvalMod Scale is `2^45`;
- therefore matched Standard is valid only as an internal/stage proxy, not as the public exact-E2E oracle.

This revises the previous metadata interpretation:
- the public `2^50 -> 2^45` EvalMod Scale reset is part of the Genuine Standard public contract, not a Fast-specific bug;
- the final public DefaultScale restoration is likewise part of the Standard public contract;
- prior diagnostic "corrections" that preserved `2^50` were useful only to expose internal Fast arithmetic error against a matched internal oracle.

Current Fast-vs-Genuine Standard:
- C2S real/imag ≈ `1e-13` difference;
- first material divergence occurs at EvalMod;
- Fast internal/matched error ≈ `0.00418/0.00426`;
- common public Scale contract magnifies this internal error by ~32;
- public Fast-vs-Genuine Standard EvalMod error is ~`0.1339/0.1364`;
- final Fast exact E2E remains ~`0.188575`.

For public `1e-2`, the corresponding internal pre-public EvalMod budget is approximately `3.125e-4`.

Do not repair Scale metadata now. The next blocker is the actual internal Fast EvalMod arithmetic/semantics.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-FAST-VS-GENUINE-STANDARD-EVALMOD-INTERNAL-BISECT.md`

The task must:
1. prove Fast and Genuine Standard C2S inputs align;
2. source-faithfully trace actual Standard and Fast EvalMod internals before the shared public Scale reset;
3. compare preprocessing, polynomial/PS, DoubleAngle, and internal restore semantics;
4. identify first observable and first material Fast-vs-Genuine-Standard internal divergence;
5. prove the common public reset maps internal error to approximately 32x public error;
6. determine whether historical P93 ~1e-4 evidence is comparable to this Genuine Standard internal oracle;
7. stop after classification; no production repair.

## Two-word workflow

1. Codex: user says only `開始`.
2. Codex synchronizes Primary, executes current spec, validates, commits and pushes Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator reviews and prepares the next task.

## Current prohibitions

- no dirty Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no metadata repair
- no coefficient correction
- no q/planScale tuning
- no P92/P94 sweep
- no production PS/C2S/S2C/finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
