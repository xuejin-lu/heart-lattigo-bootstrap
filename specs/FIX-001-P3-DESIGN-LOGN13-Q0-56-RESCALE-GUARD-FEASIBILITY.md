# FIX-001-P3-DESIGN-LOGN13-Q0-56-RESCALE-GUARD-FEASIBILITY

## Purpose

Test the smallest parameter widening that is already inside the current Fast Rescale implementation's documented domain:

- current diagnostic profile: q0 ~= 55 bits, q1 ~= 39 bits;
- candidate profile: q0 ~= 56 bits, q1 ~= 39 bits;
- all other LogQ entries unchanged.

The specific question is:

> Does one extra q0 bit provide enough Q01 centered headroom to make the previously-invalid G0 one-bit Rescale guard valid, and together with the useful final-Rescale guard reduce oracle-power PS error below the polynomial budget?

This is a diagnostic/design task only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`d9d540a306f09298dbcb1a4960bd2900bd7a3a53`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

## Accepted evidence

From the previous Rescale-guard task:

- fixed all-`2^92` oracle-power PS baseline max error: about `3.3025e-8`;
- polynomial budget: `1.2e-8`;
- G0 baseline pre-Rescale centered-Q01 capacity ratio: about `0.89729`;
- therefore a physical one-bit guard would require roughly `2 * 0.89729 > 1` of the current centered range and cannot preserve the represented value under the current Q01;
- G0 one-bit guard reduced local Rescale error from about `2.35e-8` to about `1.23e-8`, but was invalid because value preservation/scale alignment failed;
- F0 one-bit guard was valid and reduced final max error to about `2.5509e-8`;
- G1/G2/G3 one-bit guards were valid but individually less useful;
- classification: `logn13_ps_rescale_guard_insufficient_precision`.

Current Fast Rescale source explicitly supports the domain:

- q0 <= 56 bits;
- q1 <= 39 bits;
- Q01 < 2^95.

Therefore this task tests only the minimal `q0: 55 -> 56` widening. It does not redesign the Rescale implementation and does not introduce q2.

---

# Scope lock

LogN13 only.

Use validated oracle powers for PS qualification.

Keep fixed:

- polynomial coefficients, degree and Chebyshev basis;
- PS decomposition;
- all `plan.Value` scales at the accepted common `2^92` baseline;
- same level topology and number of Rescales;
- same relinearization points;
- q1 and every Q modulus after q1 unchanged in requested bit size;
- P chain unchanged;
- C2S/S2C/Mod1 algorithm unchanged.

Do not:

- modify Secondary production code;
- redesign generated powers;
- use q2+ maintained arithmetic;
- widen q1;
- widen q0 beyond 56 bits;
- add Q levels;
- change polynomial/mixed scales;
- benchmark;
- run LogN16, Gate 4/5, EXP-003;
- production-integrate a candidate.

Primary diagnostic parameter construction/helpers/tests are allowed.

---

# Q0 — reproduce current control

With the current 55/39 profile, reproduce the accepted oracle-power fixed-`2^92` baseline and guard facts:

- no-guard max error near `3.3025e-8`;
- G0 baseline capacity ratio near `0.89729`;
- G0 one-bit guard invalid/value-not-preserving;
- F0 one-bit guard valid, final max error near `2.5509e-8`.

If not, stop:

`logn13_q0_56_guard_precondition_mismatch`.

---

# Q1 — construct the minimal widened profile

Construct a diagnostic LogN13 parameter profile differing only by requested q0 bit size:

- q0: 56 bits;
- q1: 39 bits;
- all later Q bit sizes identical to the accepted profile;
- P identical.

Record the actual generated q0/q1 primes and bit lengths.

Require:

- q0 and q1 are distinct NTT-compatible primes selected by the normal parameter builder;
- Fast Rescale domain validator accepts them without Secondary code changes;
- Q01 product remains within the existing documented implementation domain;
- genuine Standard bootstrap on the widened profile remains correct with public max-component error <= `1e-2`;
- ordinary no-guard Fast/oracle-power PS executes with unchanged topology.

If parameter construction or current Fast Rescale rejects the profile, classify:

`logn13_q0_56_existing_backend_not_supported`.

Do not patch Secondary in this task.

---

# Q2 — quantify headroom change

For the widened profile, measure the same five PS Rescale boundaries:

- G0, G1, G2, G3, F0-final.

For each record compactly:

- unguarded pre-Rescale centered capacity ratio;
- one-bit-guard projected and actual capacity ratio;
- guard represented-value preservation error;
- Fast q0/q1 vs stage-aligned full-RNS row equality;
- local Rescale error before/after guard.

The key hypothesis is that G0 one-bit guard becomes centered-unique and value-preserving under widened Q01.

If G0 remains invalid despite centered capacity, identify the first exact non-capacity failure.

---

# Q3 — bounded guard candidates on widened q0

Using the widened profile and fixed `2^92` plan, evaluate exactly these candidates first:

- W0: no guards;
- W1: G0 one-bit guard only;
- W2: F0 one-bit guard only;
- W3: G0 + F0 one-bit guards.

Then, only if W3 is valid but still above budget, allow one bounded greedy pass over G1/G2/G3, at most one guard bit each, accepting an action only if it strictly lowers final max(real, imag) error while preserving all contracts.

Do not use more than one guard bit per boundary in this task.

For every candidate require:

- physical integer guard multiplication plus matching Scale multiplication;
- pre-Rescale value preservation <= `1e-10`;
- centered uniqueness;
- correct downstream physical scale alignment;
- Fast/full-RNS q0/q1 row equality where exact;
- no extra level consumption.

---

# Q4 — PS qualification

A widened-q0 candidate qualifies only if:

- real polynomial error <= `1.2e-8`;
- imag polynomial error <= `1.2e-8`;
- all capacity/value/alignment contracts pass;
- same level topology is preserved.

If one qualifies, classify:

`logn13_q0_56_rescale_guard_candidate_validated`.

Select the valid candidate with the fewest guarded boundaries, then lowest final max error.

Do not production-integrate it here.

If none qualifies, classify exactly one:

- `logn13_q0_56_rescale_guard_insufficient_precision`
- `logn13_q0_56_rescale_guard_noncapacity_semantic_failure`

Record the best valid candidate and first remaining blocker.

---

# Q5 — downstream oracle-power sufficiency

Only for a Q4-qualified candidate:

- derive normalized Mod1 K/schedule from the actual polynomial output Scale;
- run normalized DoubleAngle using valid widened-q0 Fast semantics;
- run unchanged Fast S2C;
- run supported one-input unpack/finalization/public-like path.

Record:

- EvalMod real/imag vs genuine Standard widened-profile reference;
- post-S2C semantic error;
- public-like max-component error;
- metadata correctness.

Target:

`public_like_max_component_error <= 1e-2`.

This uses oracle powers and proves only whether the PS arithmetic blocker can be closed with the minimal q0 widening.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-Q0-56-RESCALE-GUARD-FEASIBILITY-summary.json`

Include only:

- provenance;
- 55/39 control facts;
- actual widened q0/q1 primes and bit lengths;
- Standard widened-profile public result;
- widened five-boundary capacity/local-error table;
- W0-W3 results;
- optional accepted bounded greedy steps;
- selected/best candidate;
- final real/imag polynomial errors;
- downstream metrics if reached;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize full vectors, coefficient arrays, RNS rows, large traces or exhaustive rejected candidates.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*Q0.*56.*Guard` pass;
- Primary `go test ./...` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no q1 widening;
- no q2+ arithmetic;
- no generated-power redesign;
- no mixed/common plan-scale redesign;
- no production integration;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.