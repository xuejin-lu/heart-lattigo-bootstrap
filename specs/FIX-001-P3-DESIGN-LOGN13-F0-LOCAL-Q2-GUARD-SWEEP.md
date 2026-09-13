# FIX-001-P3-DESIGN-LOGN13-F0-LOCAL-Q2-GUARD-SWEEP

## Purpose

Correct the acceptance criterion from the prior local-q2 feasibility task and test whether a slightly larger F0 guard inside the already-valid local q012 region closes the PS arithmetic blocker.

Accepted evidence from commit `73b0ac1aeaafa42650ef79260c2c5412354ac558`:

- q0/q1/q2 = 56/39/40 bits;
- q01->q012 expansion passes exactly;
- local-q2 F0 two-bit guard preserves value;
- q012 rows match stage-aligned full-RNS;
- rounded division matches full-RNS;
- q2->q01 contraction passes exactly;
- R1 polynomial error: real `1.3848946989192257e-8`, imag `1.1974903735278986e-8`;
- the reported local F0 Rescale semantic error (`~3.4e-9` real, `~3.1e-9` imag) is therefore normal CKKS rounding relative to ideal real arithmetic, not a Fast/full-RNS primitive mismatch.

The previous `<=1e-10` local-Rescale semantic threshold must NOT be used as a validity requirement in this task.

The validity requirements are instead:

1. exact q012 row agreement with stage-aligned full-RNS;
2. exact rounded-division agreement;
3. value-preserving guard expansion;
4. valid q2->q01 contraction;
5. final polynomial error within the global budget.

---

## Required provenance

Primary required base:

`73b0ac1aeaafa42650ef79260c2c5412354ac558`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use validated oracle powers.

Keep fixed:

- q0/q1/q2 = existing 56/39/40 profile;
- G0 one-bit guard;
- polynomial coefficients/degree/basis/PS decomposition;
- common `2^92` plan scales;
- same logical levels, Rescale count, relinearization points;
- local q2 only at F0;
- immediate contraction back to q0/q1 after F0;
- C2S/S2C/Mod1 algorithms unchanged.

Do not:

- modify Secondary production code;
- redesign generated powers;
- widen q0/q1;
- use q3+;
- extend q2 outside the F0 local region;
- add Q levels;
- benchmark;
- run LogN16, Gate 4/5, EXP-003;
- production-integrate a candidate.

Primary diagnostic helpers/tests are allowed.

---

# S0 — reproduce controls

Reproduce:

- q01-only control G0=1/F0=1 max error near `1.241886349e-8`;
- local-q2 G0=1/F0=2 result real near `1.384894699e-8`, imag near `1.197490374e-8`;
- local-q2 expansion, row matching, rounded division and contraction all pass.

If not, stop:

`logn13_f0_local_q2_guard_sweep_precondition_mismatch`.

---

# S1 — local-q2 F0 guard sweep

Keep G0 guard fixed at 1 bit.

Within the local q012 F0 region, test total F0 guard bits:

`2, 3, 4, 5, 6`.

For each candidate:

1. expand q01 -> q012 exactly as in the accepted feasibility task;
2. physically multiply q0/q1/q2 residues by `2^k`;
3. multiply Scale by `2^k`;
4. require value preservation <= `1e-10` before Rescale;
5. require centered uniqueness under Q012;
6. allow Q01 uniqueness to be lost inside the local region;
7. perform the same F0 rounded Rescale using q012-authoritative exact semantics;
8. require q012 rows and rounded division to match stage-aligned full-RNS exactly;
9. require post-Rescale output centered-unique under Q01;
10. contract immediately to q0/q1 and require semantic equality <= `1e-10`.

Stop increasing k once Q012 uniqueness or post-Rescale Q01 contraction fails.

Do not reject a candidate solely because its local Rescale semantic error versus ideal real arithmetic exceeds `1e-10`.

---

# S2 — precision qualification

For every valid k, record:

- Q012 capacity ratio before and after guard;
- local F0 Rescale rounding error versus ideal arithmetic;
- final real polynomial error;
- final imag polynomial error;
- final max error;
- q012/full-RNS row equality;
- rounded-division equality;
- contraction validity;
- level/scale metadata.

A candidate qualifies iff:

- real error <= `1.2e-8`;
- imag error <= `1.2e-8`;
- all exact q012/full-RNS and contraction contracts pass;
- no extra level is consumed.

Select the smallest guard-bit count k that qualifies.

Success classification:

`logn13_f0_local_q2_guard_candidate_validated`.

If none qualifies, classify exactly one:

- `logn13_f0_local_q2_guard_insufficient_precision`
- `logn13_f0_local_q2_guard_capacity_failure`
- `logn13_f0_local_q2_guard_contraction_failure`.

---

# S3 — downstream oracle-power sufficiency

Only for the smallest qualifying k:

- derive normalized Mod1 K/schedule from the actual polynomial output Scale;
- run normalized DoubleAngle using ordinary q0/q1 Fast semantics after contraction;
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

Oracle powers remain in use, so generated-power error is still a separate blocker.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-F0-LOCAL-Q2-GUARD-SWEEP-summary.json`

Include only:

- provenance;
- S0 controls;
- one row for each tested guard bit count;
- selected smallest qualifying k, if any;
- final real/imag/max polynomial errors;
- downstream metrics if reached;
- classification;
- blocker-closed flag;
- first remaining blocker;
- validation flags.

Do not serialize vectors, coefficient arrays, complete RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*F0.*Local.*Q2.*Guard.*Sweep` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- q2 maintained only inside F0;
- no q3+ arithmetic;
- no q0/q1 widening;
- no generated-power redesign;
- no production integration;
- no C2S/S2C/Mod1 algorithm change;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.