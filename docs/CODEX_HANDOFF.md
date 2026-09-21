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

## Fixed architecture

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012
- planScale = `2^93`
- deterministic 4096-slot workload

No tuning unless a later synchronized spec explicitly authorizes it.

## Current accepted evidence

The `1e-2` production milestone is not yet end-to-end passed.

### EvalMod public scale-reset causal proof

Primary evidence commit:
`00bb90c25137a2d570ac11c728d081fbc96fc125`

Accepted:
- public Fast EvalMod changes only Scale metadata from internal/input-restored `2^50` to residual DefaultScale `2^45`;
- q0/q1/q2 rows remain unchanged;
- ratio = 32;
- diagnostic metadata-only correction `Scale := original C2S input Scale` reproduces forced-P93 output exactly;
- corrected EvalMod vs matched Standard:
  - real ≈ `0.00418301958`
  - imag ≈ `0.00426156114`
  - both pass `1e-2`;
- corrected post-S2C/F0 ≈ `0.00693482035`, also passes `1e-2`;
- uncorrected post-S2C/F0 ≈ `0.04027362692`.

However:
- corrected final public-like ≈ `0.22191425194`;
- uncorrected final public-like ≈ `0.22191425194`;
- corrected-minus-uncorrected final vector = exactly 0.

Therefore the EvalMod metadata correction is causal and effective through S2C, but a later downstream boundary collapses the corrected and uncorrected semantic branches back together.

Do not reopen PS, C2S, EvalMod arithmetic, or S2C.

## Current task

Read synchronized `CURRENT_TASK.md`.

At this revision:

`specs/FIX-001-P3-DIAG-P93-POST-S2C-SCALE-COLLAPSE-CAUSAL-PROOF.md`

The task must:
1. reproduce corrected and uncorrected post-S2C branches;
2. trace both through unpack/switch-back/finalization;
3. identify the first boundary where the branches semantically converge;
4. classify that convergence as metadata-only or arithmetic;
5. if metadata-only, apply a diagnostic clone-only metadata counterfactual;
6. run exact final E2E against the original deterministic message using the Standard-baseline max-component metric;
7. determine whether the combined proven metadata corrections are sufficient for the current `1e-2` milestone;
8. audit the exact downstream source assignment but do not repair Secondary.

## Two-word workflow

1. Codex: user says only `開始`.
2. Codex synchronizes Primary, executes current spec, validates, commits and pushes Primary evidence.
3. ChatGPT Web: user says only `review`.
4. Orchestrator reviews and prepares the next task.

Only exceptional repository/provenance/safety failures interrupt this workflow.

## Current prohibitions

- no dirty Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no coefficient correction
- no q/planScale tuning
- no PS/C2S/S2C production repair
- no production finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

Only diagnostic clone metadata preservation at the proven downstream convergence boundary is allowed.
