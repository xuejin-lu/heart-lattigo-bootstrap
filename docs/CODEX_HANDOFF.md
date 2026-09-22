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
- Required polynomial-output budget for current public `1e-2`: `3.716228023823462e-8`.

### Historical P93 polynomial reference is now reconciled

Primary commit:
`e23337b8087cb8fec28d857be3bff2efe78f60e0`

Accepted:
- historical/design Standard polynomial comparator equals Genuine Standard polynomial oracle exactly for the deterministic workload;
- fresh historical P93 polynomial error:
  - real `1.8570138426987626e-8`
  - imag `1.6219059983946238e-8`
- historical P93 therefore meets the derived polynomial budget;
- current production polynomial error:
  - real `4.986373945969902e-7`
  - imag `5.07761187318323e-7`
- regression vector closure passes.

### First budget-breaking generated-power checkpoint

Current production vs historical P93:

- T2 real: `6.940818919609626e-9` — within budget
- T3 real: `2.2233773797064593e-7` — first budget break
- capacity remains safe.

This proves a generated-power regression is present, but **does not yet prove T3 implementation itself is causal** for final polynomial regression.

Do not repair T3 yet.

## Current task

`specs/FIX-001-P3-DIAG-P93-T3-GENERATED-POWER-CAUSAL-DECOMPOSITION.md`

It must:

1. freshly reproduce historical/current T1/T2/T3;
2. establish the exact T3 Chebyshev recurrence and active schedule;
3. build plaintext T3 oracles for historical and current inputs;
4. decompose observed T3 regression into:
   - propagation of T1/T2 input error;
   - net T3 implementation-regression effect;
5. verify vector closure;
6. if input-propagation dominated, recurse one level into T2;
7. if implementation effect is material, localize the first stable T3 primitive boundary;
8. audit only dirty source changes actually on the executed T3 path;
9. assess T3 relevance to final PS output without replacement experiments;
10. stop after classification; no repair.

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
- no DoubleAngle repair
- no C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
