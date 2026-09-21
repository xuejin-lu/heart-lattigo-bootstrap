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
- local worktree is intentionally dirty.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the Secondary dirty worktree.

## Startup rule

When the user says `開始`, follow `AGENTS.md` mandatory preflight first, then read synchronized `CURRENT_TASK.md` and its exact spec. `CURRENT_TASK.md` is authoritative.

## Current accepted evidence

Authoritative Standard:
- Genuine Standard public/staged exact E2E = `5.830057349387463e-8`.
- matched diagnostic Standard is only a local/internal proxy from EvalMod onward.

Current Fast:
- C2S real/imag align to Genuine Standard at ~`1e-13`.
- final internal pre-public Fast-vs-Genuine-Standard EvalMod gap:
  - real `0.004183019582623534`
  - imag `0.004261561142162278`.
- shared public DefaultScale contract maps this to exactly 32x public error.
- public `1e-2` therefore corresponds to internal target `3.125e-4`.

## Important rejection of the previous task classification

Primary commit:
`0c9987fa2ebaaf924e8a6b4cd85cafe1d53a3406`

Its endpoint controls are useful, but the claimed first material `pre_double_angle` / polynomial classification is not accepted.

Why:
1. the old trace let Fast advance through DoubleAngle while the paired Standard comparison object stayed at polynomial output;
2. Fast normalized LogN13 uses a deferred maintained scalar / virtual exponent;
3. raw decode immediately after metadata-only coherent-scale changes is not in the same semantic coordinate system as Standard.

The nonphysical pattern
`~5e-7 -> 0.779 -> ~7e-7`
therefore reflects invalid checkpoint pairing/coordinates, not a proven production causal boundary.

No production repair should be based on that classification.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-EVALMOD-LOCKSTEP-LOGICAL-SEMANTICS-TRACE.md`

The task must:
1. explicitly document the previous trace defect;
2. run Standard and Fast in true lockstep logical stages;
3. track Fast virtual maintained-scalar exponent;
4. canonicalize Fast logical semantics with the virtual exponent before comparison;
5. compare Standard and Fast only at equivalent logical checkpoints;
6. reproduce the accepted final internal gap and factor-32 public mapping;
7. identify the first true observable/material divergence;
8. keep the committed summary compact (target <=300 JSON lines);
9. stop after classification; no production repair.

## Two-word workflow

1. Codex: user says only `開始`.
2. Codex syncs Primary, executes current spec, validates, commits and pushes Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator independently reviews and prepares the next task.

## Current prohibitions

- no Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no production repair
- no metadata repair
- no coefficient correction
- no q/planScale tuning
- no P92/P94 sweep
- no C2S/S2C/finalizer work
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
