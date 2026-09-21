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

## Final restore factor hypothesis rejected

Primary commit:
`58fd5ecc2f4ac4063da23736895fe3ed7ddc7c42`

Accepted:
- DA2 post-Rescale to internal-final error ratio is ~1024;
- runtime formula `2^e * S_out/S_in` yields ~`2^17`;
- current restore uses `2^27`.

But diagnostic counterfactual `2^17` did not solve the problem:
- internal error only improved to ~`0.00376/0.00377`;
- public error remained ~`0.120`;
- exact E2E remained ~`0.1875`.

Therefore do not repair final restore factor.

Also, the ~1024 error scaling is consistent with the common Standard final Scale restoration from roughly `2^60` to `2^50`, so it is not itself evidence of a Fast-specific bug.

## Stable current checkpoints

Use semantically stable post-Rescale checkpoints only:

Real:
- polynomial output ~`4.986e-7`
- DA0 post-Rescale ~`1.549e-6`
- DA1 post-Rescale ~`3.615e-6`
- DA2 post-Rescale ~`4.085e-6`

Imag:
- polynomial output ~`5.078e-7`
- DA0 post-Rescale ~`1.578e-6`
- DA1 post-Rescale ~`3.683e-6`
- DA2 post-Rescale ~`4.162e-6`

Pre-Rescale virtual-coordinate metrics are not causal evidence.

The DA2 post-Rescale budget corresponding to public `1e-2` is approximately:
`3.0517578125e-7`.

## Current task

`specs/FIX-001-P3-DIAG-P93-DOUBLEANGLE-PLAINTEXT-CAUSAL-DECOMPOSITION.md`

It must:
1. build and validate the exact plaintext DoubleAngle recurrence against Genuine Standard;
2. propagate the Fast polynomial-output semantic error through that plaintext recurrence;
3. compare propagated error with actual Fast post-Rescale DA outputs;
4. decompose polynomial-error propagation vs Fast DA implementation effect;
5. verify vector closure each round;
6. determine whether DA2 error is propagation-dominated, implementation-dominated, or mixed;
7. derive the polynomial-output error budget needed for public `1e-2`;
8. stop after classification, no production repair.

## Two-word workflow

1. Codex: `開始`
2. Codex executes current spec, validates, commits/pushes only Primary evidence.
3. ChatGPT Web: `review`
4. Orchestrator reviews and prepares the next task.

## Prohibitions

- no Secondary source modification
- no Secondary commit/push
- no destructive Git operation on dirty Secondary
- no production repair
- no q/planScale tuning
- no public Scale-contract modification
- no P92/P94 sweep
- no C2S/S2C/finalizer work
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
