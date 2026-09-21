# FIX-001-P3-DIAG-P93-ACTUAL-EVALMOD-VS-FORCED-P93-PATH

## Purpose

The preceding C2S diagnostic at Primary commit `3cf4c4f427837b0b85587d8e62666bd0d239eef4` produced two different kinds of evidence.

### Accepted evidence

The C2S boundary is not the current blocker:

- production ModUp vs controlled ModUp: semantic difference `0`;
- production Fast C2S real vs controlled Fast C2S real: `0`;
- production Fast C2S imag vs controlled Fast C2S imag: `0`;
- production/controlled Fast C2S vs aligned Standard: `0`;
- ordinary Standard C2S differs only at roughly `1e-13`.

Therefore do not modify or further localize C2S.

### Rejected classification

The preceding task classified `P93_STANDARD_REFERENCE_SEMANTICS_MISMATCH`, but its R0 replay contract was not satisfied:

- the spec expected the prior full-core context near real `0.02231793676`, imag `0.02056928703`;
- the new ordinary-Standard comparison instead measured roughly real `0.13385662664`, imag `0.13636995655`;
- the runner did not stop on that mismatch.

So the classification must not be treated as authoritative.

### Strong new clue

On an identical C2S input, two **Fast EvalMod** execution paths differ materially:

1. actual production:
   `FastEvaluator.EvalMod(ct)`

2. explicit diagnostic P93 path:
   `evalModMatchedRunPathWithPlanScaleAndFastPolynomial(..., planScale=2^93, ...)`

The preceding artifact measured actual-production-vs-forced-P93 output differences of approximately:

- real: `0.018243003747759622`
- imag: `0.01657172300885934`

This task must determine exactly where these two Fast EvalMod paths diverge and whether the production `FastEvaluator.EvalMod` is actually executing the intended P93 PS-wide Q012 design.

Do not repair Secondary in this task.

---

## Repository safety

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Follow `AGENTS.md` startup preflight, synchronize `main`, then read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary

Repository: `xuejin-lu/lattigo`

Required local state:

- branch: `fast-ckks`
- committed HEAD: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- exact expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter this dirty worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed profile

Keep fixed:

- LogN13
- q0 = 56-bit effective profile
- q1 approximately 39 bits
- q2 approximately 40 bits
- intended production design = PS-wide Q012
- intended planScale = `2^93`
- deterministic 4096-slot workload
- same C2S output for both paths
- same EvalMod polynomial / K / DoubleAngle count / message ratio
- `1e-2` remains a diagnostic system milestone, not final research acceptance

No tuning.

---

# R0 — establish one identical EvalMod input

Produce the current actual production Fast C2S real/imag once.

Clone the resulting ciphertexts so both EvalMod paths receive semantically and structurally identical inputs.

Before running EvalMod, prove for each branch:

- decoded semantic difference = `0`;
- q0/q1/q2 active row hashes equal;
- Level equal;
- Scale equal;
- Degree equal;
- NTT/Montgomery metadata equal.

If this cannot be proven, stop:

`P93_EVALMOD_IDENTICAL_INPUT_PRECONDITION_FAILURE`.

Do not use different C2S constructors for the two EvalMod paths after this point.

---

# R1 — execute the two Fast EvalMod paths from the identical clone

## Path A — actual production

Run exactly the current dirty Secondary public production call:

`profile.Fast.EvalMod(input.CopyNew())`

Record final real/imag semantics and metadata.

## Path B — forced P93 diagnostic

Starting from the identical cloned input, run the established explicit P93 diagnostic path with:

- `planScale = 2^93`;
- same intended PS-wide Q012 design;
- same polynomial and DoubleAngle parameters.

Record final real/imag semantics and metadata.

Required direct comparison:

`PathA - PathB`

using decoded vectors, without involving any Standard evaluator.

Historical clue to reproduce:

- real max component approximately `0.01824300375`;
- imag max component approximately `0.01657172301`.

If the direct Fast-vs-Fast gap does not reproduce materially, stop:

`P93_FAST_EVALMOD_PATH_REPLAY_CONFLICT`.

This direct Fast-vs-Fast comparison is the authoritative boundary for this task.

---

# R2 — read-only production configuration audit

Inspect the current dirty Secondary source implementing the public `FastEvaluator.EvalMod` call.

Do not modify it.

Record exactly:

- function chain entered by `FastEvaluator.EvalMod`;
- effective polynomial evaluator entry point;
- whether a planScale is explicitly passed;
- if so, its exact value;
- if not, how the plan scale is derived;
- PS planner/decomposition identity;
- whether Q012-wide authority is used for the full PS phase;
- where Q012 -> Q01 contraction occurs;
- preprocessing operations before polynomial evaluation;
- post-polynomial normalization/offset operations;
- each DoubleAngle round;
- any scale restore/reset.

Compare this configuration with Path B.

Produce a compact configuration-difference table.

Do not infer values from comments; record runtime or source-grounded values.

---

# R3 — stage-aligned Fast-vs-Fast trace

Trace Path A and Path B from the identical input.

No Standard ciphertext is needed.

Required checkpoints, when present in both paths:

1. EvalMod input
2. normalization / scale normalization
3. offset application
4. polynomial/PS entry
5. generated powers T2/T3/T4/T6/T8/T16
6. PS baby/giant-step outputs sufficient to identify the first material divergence
7. final PS root before final Rescale
8. final PS Rescale
9. PS-exit contraction
10. pre-DoubleAngle state
11. DA round 0 after square
12. DA round 0 after multiplier/constant
13. DA round 0 post-Rescale
14. DA round 1 post-Rescale
15. DA round 2 post-Rescale / final EvalMod output

For each aligned checkpoint record:

- Path A vs Path B max component;
- max complex;
- worst slot/component;
- Level/Scale/Degree;
- active q0/q1/q2 row-hash equality where mathematically meaningful;
- capacity status;
- amplification from previous checkpoint.

Detailed vectors remain under `/tmp`.

---

# R4 — effective planScale experiment without tuning

This is identification, not a sweep.

Determine the effective Path A planScale from R2.

Then run at most these bounded confirmations from the identical input:

1. explicit Path B with `2^93`;
2. explicit diagnostic path with the exact effective Path A planScale/configuration found in R2.

The purpose is to test whether the actual production output can be reproduced by explicitly instantiating its effective configuration.

Do not test unrelated scales.

Require direct vector comparison.

If explicit reproduction of Path A configuration yields Path A output within `1e-10` (or a justified deterministic floor), this proves a configuration/integration difference rather than an unexplained arithmetic difference.

---

# R5 — reconcile prior EvalMod numbers

Produce a compact table with exactly these rows:

1. actual public Fast EvalMod output vs forced P93 Fast output;
2. forced P93 Fast output vs its matched P93 Standard diagnostic reference;
3. actual public Fast output vs that same matched P93 Standard diagnostic reference;
4. actual public Fast output vs ordinary Standard EvalMod output;
5. historical fresh P93 design/reference EvalMod context.

For every row clearly state:

- both objects being compared;
- whether the comparison is Fast-vs-Fast or Fast-vs-Standard;
- max-component real/imag;
- whether `1e-2` is semantically applicable;
- provenance.

Do not label different references with the same name "full-core".

---

# Decision classification

Choose exactly one:

## A — `P93_EVALMOD_IDENTICAL_INPUT_PRECONDITION_FAILURE`

The two Fast paths did not start from identical C2S ciphertext semantics/metadata.

## B — `P93_FAST_EVALMOD_PATH_REPLAY_CONFLICT`

The previously observed actual-vs-forced-P93 direct gap does not reproduce.

## C — `P93_PRODUCTION_PLAN_SCALE_NOT_P93`

The public production `FastEvaluator.EvalMod` uses a different effective planScale/configuration, and explicit reproduction of that configuration reproduces Path A.

## D — `P93_PRODUCTION_PS_ARCHITECTURE_NOT_P93_Q012`

PlanScale is P93 but the public production call does not execute the intended full PS-wide Q012 architecture.

Record the first architectural difference.

## E — `P93_PRODUCTION_PREPROCESSING_DIVERGENCE`

The first material Fast-vs-Fast divergence occurs before PS.

## F — `P93_PRODUCTION_PS_ARITHMETIC_DIVERGENCE`

Configuration matches, but the first material Fast-vs-Fast divergence occurs inside PS.

## G — `P93_PRODUCTION_POST_PS_DOUBLEANGLE_DIVERGENCE`

PS outputs align materially; divergence first appears after PS / in DoubleAngle or restore.

## H — `P93_FAST_EVALMOD_TRACE_ALIGNMENT_INVALID`

The public production and forced-P93 paths cannot be aligned reliably enough for causal interpretation.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-ACTUAL-EVALMOD-VS-FORCED-P93-PATH-summary.json`

Include only:

- Primary provenance;
- Secondary branch/HEAD/dirty fingerprint;
- identical-input proof;
- Path A and Path B final direct comparison;
- configuration-difference table;
- compact stage trace;
- bounded effective-config reproduction;
- five-row prior-number reconciliation;
- one classification A-H;
- explicit statement that Secondary was not modified.

No full vectors or huge RNS dumps.

---

## Validation

Run:

- focused tests for identical-input proof and Fast-vs-Fast diagnostic;
- directly affected Primary tests;
- `git diff --check`.

On completion commit/push only Primary diagnostic code + compact evidence under `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification;
- no Secondary commit/push;
- no destructive operation on dirty Secondary;
- no C2S repair;
- no PS repair;
- no EvalMod repair;
- no S2C repair;
- no finalizer repair;
- no open-ended scale sweep;
- no q tuning;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.

The deliverable is the first causal difference between the **actual public Fast EvalMod path** and the **forced P93 Fast path**, not a repair.
