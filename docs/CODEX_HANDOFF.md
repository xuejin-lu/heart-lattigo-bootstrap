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
- Fast/Genuine Standard C2S difference: ~`1e-13`.
- shared public DefaultScale contract maps internal EvalMod error to exactly 32x public error.
- DoubleAngle causal decomposition at commit `a6b541577eaa3ed33238a4e907fd0c727dc7035f` is accepted.

### DoubleAngle result

Fast DA implementation effect is negligible:
- round-0 contribution about 0.5%
- round-1 contribution about 0.5%
- round-2 contribution about 0.4%
- vector closure passes to numerical floor.

Therefore DA2 error is >99% explained by normal propagation of the polynomial-output error.

Do not repair DoubleAngle.

### Derived polynomial accuracy requirement

To satisfy the current public `1e-2` milestone, the polynomial-output Fast-vs-Genuine-Standard error must be approximately:

`<= 3.716228023823462e-8`.

Current dirty production polynomial error:
- real ~`4.986373945969902e-7`
- imag ~`5.07761187318323e-7`

so current production misses the budget by about 13–14x.

Historical fresh P93 design evidence:
- real polynomial error ~`1.8570138426987626e-8`
- imag polynomial error ~`1.6219059983946238e-8`

which would be sufficient **if** its historical Standard polynomial comparator is semantically identical to the Genuine Standard internal polynomial oracle.

That equivalence has not yet been proven.

## Current task

`specs/FIX-001-P3-DIAG-P93-POLYNOMIAL-REFERENCE-RECONCILIATION-AND-PS-REGRESSION-LOCALIZATION.md`

It must:
1. capture the Genuine Standard internal polynomial output;
2. freshly replay the historical P93 design in isolated temporary worktrees;
3. prove or reject equivalence of historical/design Standard polynomial comparator and Genuine Standard polynomial output;
4. if equivalent, verify historical P93 meets the derived `3.716e-8` budget;
5. compare current dirty production polynomial output against historical P93 and Genuine Standard;
6. localize the first budget-breaking production regression through generated powers / PS accumulation;
7. stop after classification; no production repair.

## Two-word workflow

1. Codex: `開始`
2. Codex executes current spec, validates, commits/pushes Primary evidence.
3. ChatGPT Web: `review`
4. Orchestrator reviews and prepares next task.

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no q/planScale tuning
- no generated-power replacement sweep
- no PS repair
- no DoubleAngle repair
- no C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
