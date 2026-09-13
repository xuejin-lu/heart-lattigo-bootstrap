# FIX-001-P3-DESIGN-LOGN13-PS-RESCALE-GUARD-PRECISION

## Purpose

Continue the LogN13 PS precision investigation after common-scale and mixed-scale designs both failed.

This task tests a narrower hypothesis:

> Keep the mathematically valid PS plan fixed, but improve CKKS rounding precision only at PS Rescale boundaries using value-preserving pre-Rescale guard factors.

The goal is to determine whether the PS arithmetic blocker can be closed without widening q0/q1, adding q2, changing the polynomial plan, or modifying generated powers.

This is diagnostic/design only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`1a41918bba9754eb30b43b6ce5ad6ee6682c5cc7`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

## Accepted evidence

From previous accepted tasks:

- polynomial budget: `1.2e-8`;
- production/generated-power error is a separate blocker and is excluded here with validated oracle powers;
- common oracle-power PS at `2^91` gives about `6.27e-8` max error;
- common oracle-power PS at `2^92` improves to about `3.30e-8`;
- common `2^93` is not a valid precision path: first semantic divergence appears at `B3-term-2` and final error jumps to about `1.56e-2`;
- at that `2^93` boundary, scalar and resulting ciphertext remain centered-unique and Fast/full-RNS q0/q1 rows match, so the divergence is not established as q0/q1 capacity alias;
- mixed-scale best valid candidate `[92,92,93,92,92]` improves only to about `2.54e-8` and still fails budget;
- mixed-scale classification: `logn13_ps_oracle_mixed_scale_insufficient_precision`;
- at valid `2^92`, first remaining crossing is in the giant/final Rescale path (`G0-rescale` / final rounding evidence).

Therefore do not raise all plan.Value scales again and do not introduce another mixed-scale search.

---

# Core guard semantics

For a ciphertext representing value `m` with scale `S`, a guard factor `g = 2^k` is value-preserving if the physical coefficients are multiplied by `g` and metadata scale is also multiplied by `g`:

`(g * round(m*S)) / (g*S) ~= m`.

Immediately before a Rescale by modulus `q`, this produces output scale approximately:

`g*S/q`

instead of `S/q`, reducing the semantic effect of integer rounding if centered capacity remains valid.

A guard is not a metadata-only relabel. The residues must be physically multiplied by the exact integer guard.

---

# Scope lock

LogN13 only.

Use validated oracle powers for all PS candidate qualification.

Keep:

- polynomial coefficients unchanged;
- Chebyshev basis unchanged;
- PS decomposition unchanged;
- plan.Value scales fixed to the accepted common `2^92` baseline;
- same levels and same number of Rescales;
- same relinearization locations;
- no extra Q levels.

Do not:

- modify Secondary production code;
- redesign generated powers;
- raise common/mixed block plan scales beyond the fixed `2^92` baseline;
- change C2S/S2C;
- change Mod1 parameters;
- retune Q/P chain;
- widen q0/q1;
- maintain q2+;
- benchmark;
- run LogN16, Gate 4/5, or EXP-003;
- production-integrate a candidate.

Primary diagnostic helpers/tests are allowed.

---

# R0 — baseline reproduction

Reproduce the accepted oracle-power all-`2^92` baseline:

- max polynomial error about `3.3025e-8`;
- valid semantics/capacity;
- Fast vs stage-aligned full-RNS q0/q1 rows match;
- same PS topology and Levels;
- Secondary exact/clean.

If not, stop with:

`logn13_ps_rescale_guard_precondition_mismatch`.

---

# R1 — enumerate actual PS Rescale boundaries

Identify every Rescale executed by the real fixed `2^92` PS evaluation, in execution order.

Give stable IDs, e.g. giant-step Rescales and final Rescale. Do not assume only one giant Rescale.

For each boundary record compactly:

- input Level and Scale;
- output Level and Scale;
- pre-Rescale cumulative semantic error;
- local Rescale semantic error;
- post-Rescale cumulative semantic error;
- exact centered Q01 capacity ratio immediately before Rescale;
- first downstream checkpoint where this Rescale's error remains observable.

Use exact full-RNS/stage-aligned numerical reference where valid.

---

# R2 — single-boundary one-bit guards

For each PS Rescale boundary independently, test a one-bit guard `g=2` applied only immediately before that boundary.

Implementation semantics:

1. physically multiply maintained q0/q1 residues by integer 2;
2. multiply ciphertext Scale metadata by exactly 2;
3. verify represented value before Rescale is unchanged to <= `1e-10`;
4. perform the unchanged Fast Rescale;
5. carry the resulting true output Scale forward; do not reset it by metadata-only relabeling;
6. use only exact physical integer scale alignment downstream.

For each candidate record:

- guard boundary ID;
- guarded pre-Rescale capacity ratio;
- local Rescale error before vs after guard;
- final real/imag polynomial error;
- whether downstream scale alignment stayed semantically valid;
- no-extra-level flag.

Reject a candidate if any guarded pre-Rescale coefficient loses centered uniqueness or any downstream alignment requires metadata-only relabeling.

---

# R3 — deterministic combined guard schedule

Start from the unguarded fixed `2^92` baseline.

Use the R2 results to construct a bounded greedy schedule:

- candidate action: add one guard bit to one Rescale boundary;
- a boundary may receive up to 3 guard bits total (`g <= 8`), only if centered capacity remains valid;
- at each step choose the valid action producing the largest strict decrease in final `max(real, imag)` polynomial error;
- deterministic tie-break: earlier execution-order boundary first;
- stop when polynomial budget is reached, no valid action improves error, or 8 accepted guard-bit actions total have been used.

Do not perform exhaustive combinatorial search.

At every accepted step require:

- pre-guard and post-guard represented-value equality <= `1e-10` before Rescale;
- guarded centered uniqueness;
- Fast q0/q1 vs full-RNS row match where exact;
- physically correct downstream scale alignment;
- unchanged Level consumption.

---

# R4 — classify precision vs capacity

If a guard schedule reaches both:

- real error <= `1.2e-8`;
- imag error <= `1.2e-8`;

classify:

`logn13_ps_oracle_rescale_guard_candidate_validated`.

Select the candidate with:

1. fewest total guard bits;
2. fewest guarded boundaries;
3. lowest final max error;
4. deterministic execution-order tie-break.

Do not integrate yet.

If no schedule qualifies, classify exactly one:

- `logn13_ps_rescale_guard_blocked_by_q01_capacity`
- `logn13_ps_rescale_guard_blocked_by_scale_alignment`
- `logn13_ps_rescale_guard_insufficient_precision`

Record the best valid schedule and first limiting boundary.

This classification is the decision point for whether a later task should investigate widened q0/q1 or local q0/q1/q2 maintenance.

---

# R5 — downstream oracle-power sufficiency

Only if R4 validates a PS candidate:

- derive normalized Mod1 K/schedule from the candidate's actual polynomial output Scale;
- run valid normalized DoubleAngle;
- run unchanged Fast S2C;
- run supported one-input unpack/finalization/public-like path.

Record:

- EvalMod real/imag vs genuine Standard;
- post-S2C semantic error;
- public-like max-component error;
- metadata correctness.

Target:

`public_like_max_component_error <= 1e-2`.

This still uses oracle powers and proves only that the PS arithmetic blocker can be independently closed.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-PS-RESCALE-GUARD-PRECISION-summary.json`

Include only:

- provenance;
- R0 baseline;
- ordered Rescale-boundary table;
- one-bit single-boundary results;
- accepted greedy guard steps;
- selected/best schedule;
- worst guarded capacity ratio/checkpoint;
- final real/imag polynomial errors;
- downstream metrics if reached;
- classification;
- first limiting boundary;
- validation flags.

Do not serialize vectors, coefficient arrays, complete RNS rows, exhaustive rejected candidates, or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*PS.*Rescale.*Guard.*Precision` pass;
- Primary `go test ./...` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no generated-power redesign;
- no q0/q1 parameter widening;
- no q2+ arithmetic;
- no production integration;
- no C2S/S2C change;
- no bootstrap parameter retuning;
- no extra Q levels;
- no metadata-only scale relabeling;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.