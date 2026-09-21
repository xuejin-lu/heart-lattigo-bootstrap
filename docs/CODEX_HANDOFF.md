# Fast-CKKS Codex Handoff

This is the current operational handoff for a fresh Codex chat.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed base `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- local worktree intentionally dirty.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the Secondary dirty worktree.

## Startup

When the user says `開始`, follow `AGENTS.md` preflight, synchronize Primary, then read synchronized `CURRENT_TASK.md` and its exact spec.

## Accepted evidence

- Genuine Standard exact E2E: `5.830057349387463e-8`.
- Fast/Genuine Standard C2S difference: ~`1e-13`.
- Fast final internal pre-public EvalMod error:
  - real `0.004183019582623534`
  - imag `0.004261561142162278`.
- shared public DefaultScale contract maps internal error to exactly 32x public error.
- internal target corresponding to public `1e-2`: `3.125e-4`.

## Rejected prior classifications

Do not use:
- old polynomial/pre-DA classification from `0c9987fa...`;
- DA0 classification from `469c2417...`.

The lockstep trace still showed nonphysical pre-Rescale semantics:
large apparent error before Rescale followed by micro-error immediately after Rescale.

## Strong current causal clue

From commit `469c24171b0b480830eac52dc7ec4a5814f1059d`:

Real:
- DA2 post-Rescale error `4.084980061155408e-6`
- internal final error `0.004183019582623534`
- ratio ≈ `1024 = 2^10`.

Imag:
- DA2 post-Rescale error `4.161680802892462e-6`
- internal final error `0.004261561142162278`
- ratio ≈ `1024 = 2^10`.

Current Fast restore uses virtual exponent `e=27`, raw Scale near `2^60`, then materializes `2^27` and resets Scale to caller `2^50`.

The mathematically required factor for preserving logical semantics across both virtual exponent and Scale transition is:

[
F^* = 2^e cdot S_{out}/S_{in},
]

which is expected near `2^17`, making the current `2^27` factor too large by `2^10=1024`.

## Current task

`specs/FIX-001-P3-DIAG-P93-FINAL-MAINTAINED-RESTORE-FACTOR-CAUSAL-PROOF.md`

It must:
1. reproduce the exact 1024 amplification;
2. derive the required restore factor from runtime scales;
3. apply a diagnostic-only counterfactual factor on a clone;
4. test internal EvalMod against Genuine Standard;
5. keep the real public DefaultScale contract unchanged;
6. run exact LogN13 E2E;
7. source-audit the restore formula;
8. stop after classification, with no Secondary production change.

## Two-word workflow

1. Codex: `開始`
2. Codex executes current spec, validates, commits/pushes only Primary evidence.
3. ChatGPT Web: `review`
4. Orchestrator reviews and prepares the next task.

## Prohibitions

- no Secondary source modification
- no Secondary commit/push
- no destructive Git operation on dirty Secondary
- no q/planScale tuning
- no public Scale-contract modification
- no C2S/S2C/finalizer repair
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
