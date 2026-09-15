# FIX-001-P3-DIAG-P93-S2C-ATTRIBUTION-AND-REFERENCE-RECONCILIATION

## Purpose

The public-finalization diagnostic at Primary commit `94791163745d3271acabb17984e967eec8fc3c0d` ruled out the public finalization boundary as the current supported cause.

Current committed evidence for the intentionally dirty P93 production candidate shows:

- F0 raw post-S2C vs the matched P93 Standard-core reference: max component `0.0402736269192105`;
- F0 -> F1 unpack: maintained rows and metadata exact;
- IMForm -> MForm: bit-exact;
- F1 Scale = residual DefaultScale = `3.5184372088832e13`, ratio = `1`;
- F3 and official F4 are numerically identical;
- official Fast public output vs the ordinary public reference remains about `0.2219142512701697` max component;
- classification: `PRE_FINALIZATION_REPRODUCTION_CONFLICT`.

Therefore the current evidence does **not** support `FINAL_SCALE_RESTORATION` and does not support modifying the public finalizer.

At the same time, the accepted P93 design sweep previously established a P93 public-like proxy near `0.0097053814`, below the `1e-2` system target. That accepted diagnostic used the P93 reference/design arithmetic and ordinary Standard S2C as a system-level proxy; it was not the current dirty production Fast public path.

A historical S2C error-decomposition task also established a validated methodology using a stage-aligned full-RNS S2C mirror to distinguish:

1. upstream EvalMod semantic error propagated/amplified by the linear S2C transform; from
2. error introduced by Fast S2C arithmetic itself.

This task must reconcile these facts on the **current dirty P93 candidate**.

The only questions are:

1. Can the accepted P93 reference proxy (`~0.0097053814`) still be reproduced under its original methodology?
2. Is the current production F0 error (`~0.0402736`) explained predominantly by the production EvalMod input being transformed by an otherwise-correct S2C, or does the current Fast S2C implementation add a material independent error?

Do not repair anything in this task.

---

## Repository state and safety

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Safely synchronize `main` according to `AGENTS.md`, then re-read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

The latest accepted public-finalization evidence is:

`results/FIX-001-P3-DIAG-PUBLIC-FINALIZATION-summary.json`

Do not use the previous `/tmp` handoff result as authoritative evidence where it conflicts with this committed result.

### Secondary

Repository: `xuejin-lu/lattigo`

Expected local state:

- branch: `fast-ckks`;
- committed HEAD: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`;
- intentionally dirty fixed-width Q012 / PS-wide P93 production candidate;
- expected dirty-diff SHA-256 fingerprint from the last completed diagnostic:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`.

On entry:

1. inspect branch, HEAD, `git status --short`, and `git diff --stat`;
2. recompute SHA-256 of the same deterministic dirty diff representation used by the previous diagnostic;
3. require the fingerprint to equal the expected value above;
4. if branch/HEAD/fingerprint differs or the worktree is unexpectedly clean, stop and report `SECONDARY_HANDOFF_STATE_MISMATCH`;
5. do not reset, stash, clean, discard, checkout-overwrite, rebase, or pull across the dirty worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed architecture / workload

Keep fixed exactly:

- LogN13 only;
- q0 = 56-bit profile;
- q1 approximately 39 bits;
- q2 approximately 40 bits;
- PS-wide Q012 authoritative arithmetic;
- planScale = `2^93`;
- q3+ are not arithmetic sources;
- Q012 -> Q01 contraction only at PS exit;
- existing bounded downstream local-q2 behavior;
- same deterministic 4096-slot workload;
- same C2S inputs/matrices;
- same EvalMod degree, K, DoubleAngle count, message ratio, circuit order;
- same S2C matrix/factorization/scaling;
- component threshold `1e-2` for the system-level target.

Do not tune any of these values.

---

# R0 — reproduce the two reference facts before attribution

This step prevents another comparison between mismatched references.

## R0-A — accepted P93 design proxy

Using the same diagnostic/reference methodology as:

`FIX-001-P3-DESIGN-LOGN13-PS-WIDE-Q012-Q056-SCALE-SWEEP`

re-run **only P93** (`planScale=2^93`), not P92/P94 and not a new sweep.

Use the same accepted P93 reference/design arithmetic and the same ordinary Standard S2C proxy path that produced the accepted P93 public-like result.

Record:

- exact P93 reference EvalMod real and imag stage errors under that methodology;
- post-DoubleAngle real/imag reference semantics;
- reference post-S2C/public-like max component;
- reference provenance and exact helper/path used.

Expected historical public-like magnitude is approximately `0.0097053814`.

Do not require bit-identical floating output to the historical number, but the reproduced value must remain on the same side of the fixed `1e-2` gate and be numerically consistent with the historical result.

If the accepted P93 proxy cannot be reproduced, stop immediately with:

`P93_REFERENCE_REPRODUCTION_CONFLICT`

and explain the exact methodological/provenance difference. Do not proceed to blame production S2C or EvalMod.

## R0-B — current dirty production boundary

In the same diagnostic run and against the same matched reference semantics, reproduce:

- production Fast EvalMod real difference;
- production Fast EvalMod imag difference;
- production Fast post-S2C/F0 difference.

The prior local/committed evidence suggests EvalMod differences around `0.00418` / `0.00426` and F0 around `0.0402736`, but the current run is authoritative.

If F0 does not reproduce consistently with the committed public-finalization diagnostic, stop with:

`P93_PRODUCTION_BOUNDARY_REPRODUCTION_CONFLICT`.

---

# R1 — define stage-aligned objects

Use the same definitions throughout the task:

- `E_R_real`, `E_R_imag`: accepted P93 reference/design EvalMod outputs immediately before S2C;
- `E_F_real`, `E_F_imag`: current dirty production Fast EvalMod outputs immediately before S2C;
- `T_R`: the ordinary full-RNS/Standard S2C transform used by the accepted P93 proxy, with the exact same S2C matrix, factorization, group order, Levels and scaling;
- `T_F`: the current dirty production Fast S2C path;
- `S_R = T_R(E_R)`;
- `N_F = T_R(Mirror(E_F))`, where `Mirror(E_F)` is a stage-faithful full-RNS diagnostic representation of the production EvalMod semantics;
- `S_F = T_F(E_F)`.

The mirror is diagnostic only and must not alter Secondary.

Do not compare objects with different logical scale/level/representation semantics without first proving the conversion is faithful.

---

# R2 — validate the full-RNS S2C mirror

Reuse/adapt the methodology already established by:

`FIX-001-P3-DIAG-LOGN13-S2C-ERROR-DECOMPOSITION`

Construct a stage-aligned full-RNS mirror of the S2C transform that preserves:

- exact S2C group boundaries;
- same plaintext matrices;
- same matrix encoding/scaling;
- same logical Levels/Scales;
- same factor order and recombination semantics;
- no metadata-only relabeling to force a match.

First apply the mirror to the accepted P93 reference input and compare against `S_R`.

Require mirror fidelity `<= 1e-10` in decoded max-component difference when the existing diagnostic machinery supports that floor. If a larger numerical floor is intrinsic to the conversion, report and justify it explicitly before using the mirror.

If mirror fidelity is not trustworthy, stop with:

`P93_S2C_FULL_RNS_MIRROR_MISMATCH`.

Also prove that construction of `Mirror(E_F)` preserves the production Fast EvalMod logical values within the same justified floor before using it for attribution.

---

# R3 — causal S2C decomposition on current P93 production state

Compute vector differences internally and serialize only compact metrics.

Required terms:

1. **upstream-input propagation through reference S2C**

   `N_F - S_R`

   This answers how much output difference is caused simply by feeding the current production EvalMod semantics into the accepted/reference S2C transform.

2. **Fast S2C implementation effect**

   `S_F - N_F`

   This isolates any additional error introduced by current Fast S2C arithmetic after controlling for its input.

3. **observed total**

   `S_F - S_R`

Verify vector closure:

`(S_F - N_F) + (N_F - S_R) = S_F - S_R`

Use underlying vectors for closure, not only maxima. Require closure residual `<= 1e-10` or the smallest justified numerical floor established in R2.

For each term report:

- max component;
- max complex;
- worst slot index/component;
- relative contribution to observed total where meaningful.

Do not call S2C causal merely because the total grows from about `0.004` to about `0.040`; S2C is a linear transform and may legitimately amplify input error.

---

# R4 — group-by-group attribution

Run both paths stage-aligned through S2C groups.

At:

- S2C input;
- after group 0;
- after group 1;
- after group 2/final output;

record compact metrics for both:

### Input-propagation path

`T_R(E_F)` versus `T_R(E_R)`

Record:

- max component;
- empirical amplification from prior stage;
- worst slot/component.

### Implementation-effect path

`T_F(E_F)` versus stage-aligned `T_R(E_F)`

Record:

- max component;
- first group where the difference becomes material;
- q0/q1 row-hash equality when exact comparison is mathematically meaningful;
- Level/Scale/Degree metadata equality.

Do not introduce q2 into production S2C or modify its implementation for this diagnostic.

---

# R5 — linear-delta confirmation

Because S2C is linear, form the semantic input deltas:

- `Delta_real = E_F_real - E_R_real`;
- `Delta_imag = E_F_imag - E_R_imag`.

Apply the same reference S2C linear map to those deltas using the established diagnostic method and compare the result with `N_F - S_R`.

Record:

- EvalMod input delta max component for real and imag;
- S2C output delta max component;
- empirical amplification;
- agreement residual between direct delta propagation and `N_F - S_R`.

This check is required before classifying upstream EvalMod propagation as causal.

---

# R6 — reconcile the accepted P93 proxy with current production

Produce a compact reconciliation table with exactly these rows:

1. accepted/reproduced P93 reference EvalMod -> reference S2C (`S_R`);
2. current production EvalMod -> reference S2C (`N_F`);
3. current production EvalMod -> Fast production S2C (`S_F`);
4. current official public output (for context only; finalization was already ruled out in the preceding task).

For each row state:

- comparison reference;
- max-component error;
- pass/fail vs `1e-2` only when that threshold is semantically applicable;
- what changed relative to the preceding row.

This table must make it impossible to confuse the accepted P93 design proxy with the dirty production public path again.

---

# Decision classification

Choose exactly one primary classification.

## A — `P93_REFERENCE_REPRODUCTION_CONFLICT`

Use if the original accepted P93 proxy cannot be reproduced under its documented methodology.

Stop. The next task must repair/reconcile the reference methodology, not production.

## B — `P93_EVALMOD_ERROR_AMPLIFIED_BY_S2C`

Use only if all are true:

- accepted P93 proxy reproduces;
- mirror fidelity passes;
- linear-delta confirmation passes;
- `S_F - N_F` is negligible compared with total output difference (use the historical 10% contribution criterion unless a stricter, clearly justified threshold is available);
- `N_F - S_R` explains the dominant current F0 divergence.

Conclusion: current Fast S2C is not the primary blocker; the dirty production EvalMod semantics are insufficiently precise for downstream S2C even if their local error is below `1e-2`.

Do not repair EvalMod in this task.

## C — `P93_FAST_S2C_IMPLEMENTATION_DIVERGENCE`

Use if:

- accepted P93 proxy reproduces;
- mirror fidelity passes;
- `S_F - N_F` contributes materially (>10% of observed total) or crosses the system gate when `N_F` would otherwise pass.

Record the first S2C group where implementation divergence becomes material.

Do not repair S2C in this task.

## D — `P93_MIXED_EVALMOD_AND_S2C_BLOCKERS`

Use only if both upstream propagation and Fast S2C implementation effect are independently material and neither can reasonably be treated as secondary under the 10% attribution criterion.

## E — `P93_S2C_FULL_RNS_MIRROR_MISMATCH`

Use if the stage-aligned mirror cannot be validated sufficiently for causal interpretation.

## F — `P93_PRODUCTION_BOUNDARY_REPRODUCTION_CONFLICT`

Use if the current dirty production F0 result cannot reproduce the committed `~0.0402736` boundary while provenance/fingerprint are unchanged.

---

## Stop rule

This is diagnostic only.

Once one classification above is established, stop. Do not implement a repair in the same run.

In particular:

- if classification B, do not immediately change planScale, degree, q sizes, DoubleAngle, or S2C; the next task will decide the smallest EvalMod precision repair;
- if classification C, do not immediately patch an S2C group; the next task will localize/repair the first proven Fast S2C divergence;
- if classification A/E/F, resolve the evidence/methodology conflict first.

---

## Required artifact

Create a compact human-reviewable artifact:

`results/FIX-001-P3-DIAG-P93-S2C-ATTRIBUTION-AND-REFERENCE-RECONCILIATION-summary.json`

Include only:

- Primary commit/provenance;
- Secondary branch, committed HEAD, dirty status, diff stat and exact dirty fingerprint;
- fixed P93 architecture identity;
- R0 accepted-reference reproduction metrics;
- R0 current production EvalMod/F0 metrics;
- mirror fidelity and production-input mirror fidelity;
- three-term causal decomposition and vector-closure residual;
- group-by-group compact attribution;
- linear-delta confirmation;
- four-row reconciliation table;
- one classification A-F;
- explicit statement that Secondary production source was not modified.

Do not commit full slot vectors, coefficient arrays, matrices, RNS rows, or huge traces. Detailed temporary evidence may remain under `/tmp`.

---

## Tests / validation

Run at least:

- the new focused test(s) for this diagnostic;
- directly affected Primary tests;
- `git diff --check`.

Primary-only diagnostic helper/test changes are allowed if narrowly scoped. On successful completion, commit and push the compact Primary evidence and required Primary diagnostic code under `AGENTS.md` standing safe-push rules.

Do not modify, commit, or push Secondary.

No new performance timing is required.

---

## Prohibitions

- no Secondary production source changes;
- no q0/q1/q2 tuning;
- no planScale tuning or new scale sweep;
- no P92/P94 reruns except reading existing committed evidence;
- no polynomial-degree/K/DoubleAngle changes;
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
- no reset/stash/clean/discard of the intentional Secondary dirty worktree.

The deliverable is a reconciled causal attribution, not a repair.