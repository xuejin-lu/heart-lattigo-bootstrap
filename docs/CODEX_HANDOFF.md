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
- Historical P93 polynomial accuracy is valid and within budget.
- Current production polynomial error is ~`5e-7`.
- DoubleAngle is not the blocker.
- q01-vs-q012 Rescale/scaling A/B is not causal.

## Schedule causality is proven

Primary commit:
`dd074420d81d2cd9ba73437864b76094ef735c59`

Current balanced pre-Rescale generated-power scheduling is causally responsible for the T3 precision regression.

Current balanced T3:
- implementation residual ~`1.86e-7..2.23e-7`.

Current fixed-width Q012 post-product schedule on the exact same T1/T2:
- residual ~`1e-16`;
- Q012 centered capacity safe.

Historical/big-int Q012 post-product oracle:
- residual ~`1e-16`;
- agrees with the same semantics.

Bounded T6 relevance:
- real `6.2948e-8 -> 1.7304e-8`
- imag `5.2192e-8 -> 1.8343e-8`
- both re-enter the polynomial budget.

The summary field `r7.authorized=false` was a diagnostic bookkeeping bug: the runner initialized it false and never assigned it true. Its individual criteria and classification all passed. Use the actual evidence, not that stale boolean.

## Historical schedule

Historical P93 generated powers T2/T3/T4/T6/T8/T16 all used:

[
	ext{direct multiply at full scales}
ightarrow
	ext{relinearize}
ightarrow
	ext{full Chebyshev recurrence}
ightarrow
	ext{single post-product Rescale}.
]

Historical evidence shows all required powers were Q012-safe.

## Current task

`specs/FIX-001-P3-PROD-LOGN13-Q012-GENERATED-POWER-POSTPRODUCT-SCHEDULE-REPAIR.md`

This task is explicitly authorized to make a **minimal local Secondary production modification** in:

`circuits/ckks/polynomial/fast.go`

for the already-proven LogN13/P93/Q012 generated-power domain.

Required repaired ordering:

[
oxed{	ext{Mul/MulRelin}ightarrow	ext{Chebyshev recurrence}ightarrow	ext{single Rescale}}
]

Do not merely force the old non-balanced branch if that branch rescales before recurrence subtraction.

Preserve existing balanced behavior as fallback outside the proven Q012 domain.

After repair validate:

1. T2/T3/T4/T6/T8/T16 semantic accuracy and Q012 capacity;
2. polynomial output <= `3.716228023823462e-8`;
3. DA2 <= `3.0517578125e-7`;
4. internal final <= `3.125e-4`;
5. public EvalMod <= `1e-2`;
6. exact full LogN13 E2E <= `1e-2`;
7. relevant Secondary and Primary tests.

## Secondary handling

This task may modify Secondary locally.

At completion:
- **do not commit Secondary**
- **do not push Secondary**
- record new dirty diff stat
- record new dirty SHA-256 fingerprint.

Primary evidence may be committed/pushed according to `AGENTS.md`.

If exact E2E passes, the next orchestrator task may authorize final Secondary commit/push after diff review.

## Two-word workflow

1. Codex: `開始`
2. Codex executes repair + validation, commits/pushes Primary evidence only.
3. ChatGPT Web: `review`
4. Orchestrator reviews and decides whether Secondary may be committed.

## Prohibitions

- no Secondary reset/stash/clean/discard
- no Secondary commit
- no Secondary push
- no q/planScale tuning
- no public Scale-contract modification
- no Rescale primitive rewrite
- no C2S/S2C/finalizer repair
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
