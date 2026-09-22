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
- Required polynomial-output budget for public `1e-2`: `3.716228023823462e-8`.
- Historical P93 polynomial error is valid and within budget.
- Current production polynomial error is ~`5e-7`.
- DoubleAngle is not the blocker.

## T3 implementation regression is proven

Primary commit:
`bd1a001017ea5f32e46d5fcebc3808fd4d3bbc07`

T3 regression is >99.9% implementation effect, not propagated T1/T2 error.

## First primitive boundary

Primary commit:
`a4872ef14b4bde43aa6c8c307b9a7d6c2b533daf`

A source-faithful T3 shadow reproduces production exactly.

The first material semantic residual is the **right balanced operand post-Rescale**:

Real:
- left post-Rescale residual `1.771577862186291e-8`
- right post-Rescale residual `1.1105472362549218e-7`

Imag:
- left post-Rescale residual `1.606223765104886e-8`
- right post-Rescale residual `9.276714673864261e-8`

Capacity is extremely safe; this is not centered-CRT overflow.

The previous spec required a bounded q01-vs-q012 A/B when operand Rescale became the first material boundary, but that A/B was not executed. Therefore do not repair production yet.

## Current task

`specs/FIX-001-P3-DIAG-P93-T3-Q012-RESCALE-CAUSAL-AB.md`

It must:

1. reproduce the current right-operand residual;
2. compare current q2-aware maintained scaling against a q01-only scaling shadow on the exact same input;
3. run current Q012 Rescale vs clean q01-authoritative Rescale on the exact same pre-Rescale ciphertext;
4. run q01 scaling + q01 Rescale as the third bounded control;
5. repeat the same bounded check on the left operand;
6. feed the causally supported operand path back into T3 and measure endpoint residual;
7. decide whether Q012 Rescale, q2-scaling/rescale interaction, or neither is causal;
8. state whether a later minimal production repair is authorized;
9. stop without modifying Secondary.

## Two-word workflow

1. Codex: `開始`
2. Codex executes current spec, validates, commits/pushes Primary evidence.
3. ChatGPT Web: `review`
4. Orchestrator reviews and prepares the next task.

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no Rescale production repair
- no q2 production disablement
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
