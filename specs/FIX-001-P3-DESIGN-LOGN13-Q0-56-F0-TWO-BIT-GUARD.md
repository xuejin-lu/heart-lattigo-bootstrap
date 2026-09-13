# FIX-001-P3-DESIGN-LOGN13-Q0-56-F0-TWO-BIT-GUARD

## Purpose

Close the remaining oracle-power PS precision gap using the smallest newly-enabled guard schedule under the already validated 56/39-bit q0/q1 profile.

The accepted 56/39 result established:

- G0 one-bit guard is now valid;
- F0 one-bit guard is valid;
- combined W3 (`G0=1 bit`, `F0=1 bit`) gives max polynomial error `1.2418863493124377e-8`;
- budget is `1.2e-8`;
- W3 therefore misses by only about `4.19e-10`;
- under 56/39, F0 unguarded capacity ratio is about `0.19486845`, one-bit ratio about `0.38973690`, so a two-bit F0 guard is projected near `0.77947380`, still centered-unique;
- a third F0 guard would be projected beyond 1 and is out of scope.

This task tests whether `G0=1 bit` plus `F0=2 bits` is sufficient, with no other design change.

This is diagnostic/design only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`29562ce5998c77a27584ed1f8251ab1a5b446c06`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use the validated oracle-power path.

Keep fixed:

- q0 requested/actual bit length 56;
- q1 bit length 39;
- all later Q/P entries unchanged;
- polynomial coefficients, degree and Chebyshev basis;
- PS decomposition;
- all plan.Value scales at common `2^92`;
- same levels, Rescale count and relinearization points;
- no extra Q levels.

Do not:

- modify Secondary production code;
- change generated powers;
- add q2+ maintained arithmetic;
- widen q0 beyond 56 or widen q1;
- use common/mixed-scale redesign;
- change C2S/S2C/Mod1 algorithms;
- benchmark;
- run LogN16, Gate 4/5 or EXP-003;
- production-integrate a candidate.

Primary diagnostic helpers/tests are allowed.

---

# T0 — reproduce accepted 56/39 W3

Reproduce:

- genuine Standard widened-profile public error <= `1e-2`;
- W3 guard schedule: G0=1, F0=1, all others 0;
- W3 real error near `1.185708509e-8`;
- W3 imag error near `1.241886349e-8`;
- W3 max error near `1.241886349e-8`;
- all value-preservation/capacity/alignment/row-equality/level contracts pass.

If not, stop:

`logn13_q0_56_f0_two_bit_precondition_mismatch`.

---

# T1 — exact F0 two-bit feasibility

At F0-final-rescale, test a physical two-bit guard (`g=4`) under the 56/39 profile.

Require:

1. physically multiply maintained q0/q1 residues by 4;
2. multiply Scale metadata by exactly 4;
3. represented value before Rescale unchanged <= `1e-10`;
4. exact centered-Q01 capacity ratio < 1;
5. Fast q0/q1 rows equal the stage-aligned full-RNS mirror where exact;
6. unchanged level consumption;
7. downstream scale alignment remains physical/valid; no metadata-only relabel.

Record:

- unguarded, one-bit and two-bit F0 capacity ratios;
- F0 local Rescale error for 0/1/2 guard bits;
- F0 post-Rescale cumulative error for 0/1/2 guard bits.

Do not test 3 F0 guard bits.

---

# T2 — bounded candidate set

Evaluate exactly these oracle-power candidates under 56/39:

- C0: no guards;
- C1: G0=1, F0=1 (accepted W3 control);
- C2: F0=2 only;
- C3: G0=1, F0=2.

All other boundaries remain unguarded.

For each record:

- real polynomial error;
- imag polynomial error;
- max error;
- capacity-safe;
- guard-value-preserving;
- scale-alignment-safe;
- Fast/full-RNS row equality;
- no-extra-level flag.

No greedy search and no additional guard combinations in this task.

---

# T3 — qualification

Candidate C3 qualifies if both:

- real polynomial error <= `1.2e-8`;
- imag polynomial error <= `1.2e-8`;

and every semantic/capacity/level contract passes.

If qualified, classify:

`logn13_q0_56_f0_two_bit_guard_candidate_validated`.

If not, classify exactly one:

- `logn13_q0_56_f0_two_bit_guard_capacity_failure`
- `logn13_q0_56_f0_two_bit_guard_alignment_failure`
- `logn13_q0_56_f0_two_bit_guard_insufficient_precision`

Record the first remaining blocker.

---

# T4 — downstream oracle-power sufficiency

Only if C3 qualifies:

- derive normalized Mod1 K/schedule from C3's actual polynomial output Scale;
- run normalized DoubleAngle using valid 56/39 Fast semantics;
- run unchanged Fast S2C;
- run supported one-input unpack/finalization/public-like path.

Record:

- EvalMod real/imag vs genuine Standard widened-profile reference;
- post-S2C semantic error;
- public-like max-component error;
- final metadata correctness;
- input unchanged.

Target:

`public_like_max_component_error <= 1e-2`.

If this passes, record:

`PS_ARITHMETIC_BLOCKER_INDEPENDENTLY_CLOSED = true`

This still uses oracle powers, so generated-power error remains a separate production blocker.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-Q0-56-F0-TWO-BIT-GUARD-summary.json`

Include only:

- provenance;
- T0 control;
- F0 0/1/2-bit capacity and local-error table;
- C0-C3 compact candidate rows;
- selected C3 result;
- downstream metrics if reached;
- classification;
- PS arithmetic blocker closed flag;
- first remaining blocker;
- validation flags.

Do not serialize vectors, coefficient arrays, RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*Q0.*56.*F0.*Two.*Bit.*Guard` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no q0/q1 widening beyond 56/39;
- no q2+ arithmetic;
- no generated-power redesign;
- no production integration;
- no C2S/S2C/Mod1 algorithm change;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.