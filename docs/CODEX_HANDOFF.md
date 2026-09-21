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

Never reset, stash, clean, discard, checkout-overwrite, rebase, or otherwise destroy the Secondary dirty worktree.

## Startup rule

When the user says `開始`, first follow the mandatory startup preflight in `AGENTS.md`:

1. safely synchronize Primary `main`;
2. re-read `AGENTS.md`;
3. re-read this file;
4. re-read the freshly synchronized `CURRENT_TASK.md`;
5. read the exact spec referenced by `CURRENT_TASK.md`;
6. then execute only that task.

`CURRENT_TASK.md` after safe remote synchronization is authoritative if any older historical document disagrees.

## Current architecture

The active LogN13 Fast-CKKS candidate is fixed at:

- q0 = 56-bit profile;
- q1 approximately 39 bits;
- q2 approximately 40 bits;
- PS-wide Q012 authoritative arithmetic;
- planScale = `2^93`;
- q3+ are not arithmetic sources;
- Q012 -> Q01 contraction once at PS exit;
- existing bounded downstream local-q2 behavior;
- same deterministic 4096-slot workload.

Do not retune these parameters unless a later synchronized spec explicitly authorizes it.

## Current correctness status

The `1e-2` milestone is **not yet end-to-end passed** by the current dirty production path.

Accepted/reference evidence:
- historical accepted P93 design proxy public-like error ≈ `0.0097053814`, which is below `1e-2`.

Current dirty production evidence:
- EvalMod real ≈ `0.0041830196`;
- EvalMod imag ≈ `0.0042615611`;
- raw post-S2C/F0 vs matched P93 Standard-core reference ≈ `0.0402736269`, which fails `1e-2`;
- official Fast public output max-component error ≈ `0.2219142513`, which fails `1e-2`.

The latest public-finalization diagnostic ruled out the public finalizer as the supported cause:
- F0 -> F1 unpack exact;
- IMForm <-> MForm round-trip bit-exact;
- F1 Scale / residual DefaultScale = 1;
- F3 equals official F4.

Therefore the remaining immediate question is upstream of public finalization: reconcile the accepted P93 proxy with the current production path and determine whether the post-S2C failure is dominated by propagated EvalMod error or by Fast S2C implementation divergence.

## Current task

The authoritative task is the one referenced by synchronized `CURRENT_TASK.md`.

At this handoff revision, it is:

`specs/FIX-001-P3-DIAG-P93-S2C-ATTRIBUTION-AND-REFERENCE-RECONCILIATION.md`

This task is diagnostic only.

It must:
1. reproduce the accepted P93 reference proxy under its original methodology;
2. reproduce the current dirty production EvalMod and F0 boundary;
3. validate a stage-aligned full-RNS S2C mirror;
4. separate:
   - upstream EvalMod error propagated through reference S2C; from
   - Fast S2C implementation effect;
5. perform group-by-group and linear-delta confirmation;
6. commit/push a compact Primary summary and stop at the specified classification.

Do not repair Secondary in this task.

## Two-word operating workflow

The intended user workflow is:

1. In Codex, user says only: `開始`.
2. Codex safely synchronizes Primary, reads the current task/spec, performs the task, runs required validation, and when authorized/safe commits and pushes the compact Primary evidence.
3. The user returns to ChatGPT Web and says only: `review`.
4. The orchestrator reviews the newest Primary evidence independently and either:
   - accepts the result and prepares the next spec/`CURRENT_TASK.md`; or
   - identifies an evidence gap and prepares the smallest next diagnostic.

Do not require the user to manually restate the task or paste old conversational context.

If an exceptional safety stop occurs before Codex can safely record/push Primary evidence (for example Primary itself is dirty, remote sync fails, or repository provenance is unsafe), report that exception clearly to the user. Otherwise preserve the two-word workflow.

## Prohibitions for the current task

- no Secondary production source changes;
- no Secondary commit/push;
- no q0/q1/q2 tuning;
- no planScale sweep;
- no T2 repair;
- no PS repair;
- no generated-power replacement sweep;
- no S2C production repair;
- no public-finalization repair;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign;
- no destructive operation on the intentional Secondary dirty worktree.
