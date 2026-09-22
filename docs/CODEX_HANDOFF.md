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

## q01-vs-q012 primitive A/B result

Primary commit:
`48a42226182088167eebdc822b5ffbe00d629efa`

Accepted:
- production q0/q1 rows after maintained integer scaling are identical to q01-only shadow;
- current Q012 Rescale, clean q01 Rescale, and q01 scaling + q01 Rescale all produce the same balanced-operand residual;
- substituting q01 operand paths into T3 does not improve T3;
- therefore q01-vs-q012 scaling/rescale implementation differences are not causal.

Do not repair Q012 Rescale.

## Important schedule provenance correction

Historical P93 design at Primary `41993cd03f85ec5f2de54dc21ade5ec812090d50` did not use production balanced pre-Rescale generated-power scheduling.

Historical design T3 schedule:
1. direct Q012 multiply at full operand scales;
2. relinearize;
3. Chebyshev recurrence;
4. one post-product Q012 Rescale.

Current production T3 schedule:
1. balanced factor pair;
2. scale each operand;
3. Rescale each operand to about `2^30`;
4. multiply;
5. recurrence;
6. no final generated-power Rescale in the balanced branch.

Historical T3 final recurrence residual is ~`1e-16`.
Current balanced operand Rescales introduce ~`1e-8..1e-7` residuals before multiplication.

Therefore historical P93 remains a valid design target, but its generated-power arithmetic schedule is not the same as current production.

## Current task

`specs/FIX-001-P3-DIAG-P93-T3-BALANCED-VS-POSTPRODUCT-SCHEDULE-CAUSAL-AB.md`

It must:

1. record the schedule provenance correction explicitly;
2. reproduce current balanced T3 control;
3. build a diagnostic current-primitives post-product T3 from the exact same current T1/T2;
4. compare it to a historical/big-int Q012 post-product oracle;
5. determine whether balanced pre-Rescale scheduling itself causes the T3 precision regression;
6. quantify the precision mechanism;
7. if T3 recovers >=90%, perform one bounded T6 relevance check;
8. state whether a later minimal production schedule repair is authorized;
9. stop without modifying Secondary.

## Two-word workflow

1. Codex: `開始`
2. Codex executes current spec, validates, commits/pushes Primary evidence.
3. ChatGPT Web: `review`
4. Orchestrator reviews and prepares next task.

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no production schedule repair
- no generated-power sweep
- no full PS replacement
- no q/planScale tuning
- no DoubleAngle/C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
