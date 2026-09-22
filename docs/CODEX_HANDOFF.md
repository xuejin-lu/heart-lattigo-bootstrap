# Fast-CKKS Codex Handoff

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
- Fast/Genuine C2S difference: ~`1e-13`.
- DoubleAngle implementation effect is <1%; DA error is polynomial-error propagation dominated.
- Required polynomial-output budget for public `1e-2`: `3.716228023823462e-8`.
- Historical P93 polynomial error is genuine and within budget:
  - real `1.8570138426987626e-8`
  - imag `1.6219059983946238e-8`.
- Current production polynomial error is ~`5e-7`.

## T3 causal decomposition accepted

Primary commit:
`bd1a001017ea5f32e46d5fcebc3808fd4d3bbc07`

For current vs historical:

Real:
- T1 regression = 0
- T2 regression = `6.940818919609626e-9`
- observed T3 regression = `2.2233773797064593e-7`
- T1/T2 propagation contribution = `2.16898193849957e-10`
- implementation-regression contribution = `2.225546361644959e-7`

Imag:
- T1 regression = 0
- T2 regression = `5.797987645550506e-9`
- observed T3 regression = `1.855698303146469e-7`
- input propagation = `1.8118786332399495e-10`
- implementation-regression = `1.857510181779709e-7`

Vector closure passes to numerical floor.

Therefore T3 regression is overwhelmingly an implementation regression, not propagated T1/T2 error.

However, the preceding task could only identify `T3-final` as the first stable material boundary. Dirty source candidates on the executed path include:
- maintained q2 copies/workspace
- q012 `MulIntegerMaintained`
- q012 multiplication
- Q012 rescale

None is yet causally identified.

Do not repair T3 yet.

## Current task

`specs/FIX-001-P3-DIAG-P93-T3-PRIMITIVE-SEMANTIC-RESIDUAL-LOCALIZATION.md`

It must:
1. build a source-faithful diagnostic shadow of current T3 using the exact current T1/T2 inputs;
2. reproduce actual production T3;
3. compare semantically stable boundaries against the exact plaintext recurrence:
   [
   T_3=2T_2T_1-T_1;
   ]
4. locate the first material implementation residual among:
   - balanced operands post-Rescale
   - product
   - doubling
   - subtraction/alignment;
5. validate historical clean semantics at the relevant stable boundaries;
6. perform at most one bounded q01-vs-q012 primitive A/B if required by the observed first boundary;
7. stop after classification; no repair.

## Two-word workflow

1. Codex: `開始`
2. Codex executes current spec, validates, commits/pushes Primary evidence.
3. ChatGPT Web: `review`
4. Orchestrator reviews and prepares next task.

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no T3/T2 repair
- no generated-power replacement sweep
- no PS repair
- no q/planScale tuning
- no DoubleAngle/C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
