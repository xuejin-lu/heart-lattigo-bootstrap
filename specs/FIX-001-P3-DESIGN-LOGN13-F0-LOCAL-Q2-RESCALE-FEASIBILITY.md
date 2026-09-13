# FIX-001-P3-DESIGN-LOGN13-F0-LOCAL-Q2-RESCALE-FEASIBILITY

## Purpose

Test the smallest three-limb design that directly targets the remaining PS arithmetic blocker.

Accepted prior evidence shows:

- q0/q1 = 56/39 bits makes the G0 one-bit guard valid;
- G0=1 + F0=1 gives max polynomial error `1.2418863493124377e-8`, only slightly above the `1.2e-8` budget;
- F0=2 alone is valid under q0/q1 and improves local F0 Rescale error;
- but combining G0=1 + F0=2 makes `F0-final-rescale.guard_value` fail catastrophically, with final semantic error about 1.0;
- therefore two-limb scale/guard tuning is considered exhausted for this PS blocker.

This task tests a local three-limb expansion only at `F0-final-rescale`:

> Run the rest of PS with maintained q0/q1 exactly as before. After the valid G0 one-bit guard path reaches F0, reconstruct a temporary q2 residue from the exact centered q0/q1 coefficient representation, perform the F0 two-bit guard and Rescale using q0/q1/q2 as the authoritative CRT domain, then drop back to q0/q1 immediately after Rescale if the post-Rescale coefficients are again centered-unique in Q01.

This is diagnostic/design only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`e8d5b578d3ba08234c7d9892d87d4e6fbd8785c5`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

## Fixed parameter profile

Use the already validated widened profile:

- q0 requested/actual bit length: 56;
- q1 bit length: 39;
- q2 is the existing third Q modulus from the same parameter chain, expected around 40 bits;
- all later Q and P moduli unchanged;
- no new modulus and no extra Q level.

Record the actual q0, q1, q2 primes and bit lengths.

The local three-limb CRT modulus is:

`Q012 = q0 * q1 * q2`.

Do not change the modulus chain itself.

---

# Scope lock

LogN13 only.

Use validated oracle powers so generated-power error remains excluded.

Keep fixed:

- polynomial coefficients, degree, basis and PS decomposition;
- all plan.Value scales at common `2^92`;
- q0/q1 = 56/39 profile;
- G0 one-bit guard;
- F0 two-bit guard target;
- same logical Levels, same Rescale count, same relinearization points;
- C2S/S2C/Mod1 algorithm unchanged.

Do not:

- modify Secondary production code;
- redesign generated powers;
- maintain q2 outside the explicitly bounded F0 local region;
- use q3+;
- widen q0/q1 further;
- add Q levels;
- change common/mixed plan scales;
- benchmark;
- run LogN16, Gate 4/5 or EXP-003;
- production-integrate a candidate.

Primary diagnostic helpers/tests may implement local q012 materialization and exact/big-integer CRT arithmetic.

---

# L0 — reproduce accepted controls

Reproduce under q0/q1=56/39:

1. C1 control: G0=1, F0=1 is valid and max error near `1.241886349e-8`.
2. C3 control: G0=1, F0=2 in q0/q1-only mode fails at `F0-final-rescale.guard_value` with catastrophic semantic divergence.
3. Secondary exact/clean.

If any control mismatches, stop:

`logn13_f0_local_q2_precondition_mismatch`.

---

# L1 — define exact q01 -> q012 expansion at F0

Immediately before applying the F0 two-bit guard on the valid G0=1 path:

1. require the unguarded F0 input to be centered-unique under Q01;
2. convert q0/q1 from NTT to coefficient representation using the existing partial transform semantics;
3. reconstruct the exact centered integer coefficient `x` in `(-Q01/2, Q01/2)` from q0/q1;
4. set q2 residue to `x mod q2`;
5. transform the new q2 row into the same NTT/Montgomery representation as q0/q1;
6. keep q0/q1 unchanged bit-for-bit.

Validate expansion against the stage-aligned full-RNS oracle:

- q0/q1 rows unchanged exactly;
- reconstructed q2 row equals the full-RNS q2 row wherever the Q01 centered-uniqueness premise holds;
- represented semantic value changes by <= `1e-10`.

If q2 cannot be reconstructed exactly from a Q01-unique F0 input, stop:

`logn13_f0_local_q2_expansion_mismatch`.

Do not use dormant/stale q2 contents as input evidence.

---

# L2 — q012-authoritative two-bit F0 guard

On the expanded q012 F0 input:

1. physically multiply q0/q1/q2 residues by 4;
2. multiply ciphertext Scale by exactly 4;
3. require represented value preservation <= `1e-10` under centered Q012 interpretation;
4. record exact pre-guard and post-guard Q01 and Q012 centered-capacity ratios;
5. allow Q01 uniqueness to be lost inside this local region, provided Q012 remains centered-unique;
6. require q012 rows to match a stage-aligned full-RNS reference.

This is the central purpose of the task: Q012 is temporarily authoritative at F0, so Q01 wrap inside the local region is allowed and must not be misclassified as failure.

---

# L3 — q012-authoritative F0 Rescale

Perform the same logical F0 Rescale divisor and level transition as the existing PS plan, but compute the rounded centered division using Q012-authoritative coefficients.

Diagnostic implementation may use exact `big.Int` centered CRT/division to avoid fixed-width assumptions.

Requirements:

- same divisor/modulus as the ordinary F0 Rescale;
- same one-level consumption;
- same CKKS rounded division semantics;
- q0/q1/q2 output residues match the stage-aligned full-RNS reference;
- output Scale is exactly the physical input Scale divided by the same dropped modulus;
- no metadata-only relabeling.

Record:

- local F0 Rescale semantic error;
- post-F0 cumulative error;
- Q012 pre/post capacity ratio;
- Q01 post-Rescale capacity ratio.

---

# L4 — drop q2 immediately after F0

After the q012 F0 Rescale, require the output to be centered-unique again under Q01.

If true:

- discard q2;
- retain q0/q1 unchanged;
- continue with the existing q0/q1 Fast representation.

Validate:

- q0/q1-only decoded value after dropping q2 equals the q012/full-RNS value <= `1e-10`;
- no change in Level, Scale, NTT/Montgomery metadata beyond ordinary F0 Rescale semantics.

If post-F0 output is not Q01-centered-unique, classify:

`logn13_f0_local_q2_cannot_contract_to_q01`.

Do not extend the q2 region beyond F0 in this task.

---

# L5 — polynomial qualification

Evaluate exactly two oracle-power candidates:

- R0: accepted q0/q1-only C1 control (`G0=1, F0=1`);
- R1: `G0=1`, then local-q2 `F0=2`, then immediate contraction to q0/q1.

R1 qualifies if:

- real polynomial error <= `1.2e-8`;
- imag polynomial error <= `1.2e-8`;
- local q012 expansion/guard/Rescale/contraction contracts all pass;
- same logical level topology is preserved.

If qualified, classify:

`logn13_f0_local_q2_rescale_candidate_validated`.

If not, classify exactly one:

- `logn13_f0_local_q2_insufficient_precision`
- `logn13_f0_local_q2_expansion_mismatch`
- `logn13_f0_local_q2_rescale_mismatch`
- `logn13_f0_local_q2_cannot_contract_to_q01`.

---

# L6 — downstream oracle-power sufficiency

Only if R1 qualifies:

- derive normalized Mod1 K/schedule from R1's actual polynomial output Scale;
- run normalized DoubleAngle with ordinary q0/q1 Fast semantics after q2 contraction;
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

If passed, record:

`PS_ARITHMETIC_BLOCKER_INDEPENDENTLY_CLOSED = true`.

Remember: oracle powers are still used, so generated-power error remains a separate blocker.

---

# Performance accounting for later design choice

Do not benchmark in this task, but record static operation scope:

- number of coefficient transforms required to enter/exit local q2;
- number of q2 NTT/INTT operations;
- number of q2 limb arithmetic passes;
- whether q2 exists only for one Rescale boundary.

This allows a later production task to estimate why local q2 should cost much less than maintaining three limbs through all C2S/S2C/PS operations.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-F0-LOCAL-Q2-RESCALE-FEASIBILITY-summary.json`

Include only:

- provenance;
- actual q0/q1/q2 primes and bit lengths;
- L0 controls;
- F0 pre-expansion Q01 capacity;
- q01->q012 expansion validation;
- Q01 and Q012 capacity before/after two-bit guard;
- q012 Rescale local/cumulative error;
- post-Rescale Q01 contraction validation;
- R0/R1 final real/imag polynomial errors;
- downstream metrics if reached;
- static local-q2 operation counts;
- classification;
- blocker-closed flag;
- first remaining blocker;
- validation flags.

Do not serialize vectors, coefficient arrays, complete RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*F0.*Local.*Q2.*Rescale` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- q2 maintained only inside the F0 local region;
- no q3+ arithmetic;
- no q0/q1 widening beyond 56/39;
- no generated-power redesign;
- no production integration;
- no C2S/S2C/Mod1 algorithm change;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.