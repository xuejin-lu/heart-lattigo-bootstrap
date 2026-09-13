# FIX-001-P3-DESIGN-LOGN13-G0-LOCAL-Q2-GUARD-SWEEP

## Purpose

Close the remaining ~7.7% LogN13 public-like gap by targeting the largest still-unimproved PS Rescale boundary: G0.

Accepted system path at Primary commit `07e4e8e1c4a9fca8c0f4504f4fda4099a3d883de`:

- q0/q1/q2 profile = 56/39/40 bits;
- oracle powers;
- common PS plan scale `2^92`;
- G0 guard = 1 bit under q0/q1;
- F0 bounded local-q2 guard = 3 bits;
- all three DA rounds use bounded local-q2 from post-square through multiplier/constant/Rescale and contract after each round;
- all q012 expansion, rows, rounded division, contraction, Level/Scale metadata contracts pass;
- final EvalMod max-component: real `0.00010200712264148441`, imag `0.00009681523035057725`;
- post-S2C max-component `0.0003364984531681451`;
- public-like max-component `0.01076789220080715`, slightly above the system threshold `1e-2`.

The public-like/post-S2C ratio is essentially 32, so the finalization path is behaving as an expected scale restoration rather than introducing a new independent semantic defect. To pass the system target, post-S2C error needs only a modest reduction to approximately `<= 3.125e-4`.

Historical G0 evidence under q0/q1=56/39:

- unguarded G0 pre-Rescale Q01 capacity ratio ~`0.44864535`;
- one-bit G0 guard ratio ~`0.89729071`, valid;
- G0 local Rescale error dropped from ~`2.3489e-8` to ~`1.2273e-8` with one guard bit;
- a second physical guard bit would require about `1.7946` of Q01 half-range and therefore cannot be represented uniquely in q0/q1, but should be trivial under Q012 headroom.

This task tests bounded local-q2 at G0 with a small guard sweep, while keeping the already validated F0 and DA local-q2 system path fixed.

Diagnostic/design only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`07e4e8e1c4a9fca8c0f4504f4fda4099a3d883de`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use validated oracle powers.

Keep fixed:

- q0/q1/q2 = existing 56/39/40 profile;
- common plan scale `2^92`;
- polynomial coefficients, basis and PS decomposition;
- F0 local-q2 guard = 3 bits;
- DA rounds 0..2 bounded local-q2 design exactly as accepted in `07e4e8e...`;
- normalized DA schedule and constants unchanged;
- S2C and finalization/public-like path unchanged;
- same logical Levels and Rescale counts.

Only variable in this task:

- G0 guard bits, tested under a bounded local-q2 region.

Do not:

- retune F0;
- retune DA K/multipliers/constants;
- change plan scale;
- redesign generated powers;
- widen q0/q1;
- use q3+;
- add Q levels;
- modify Secondary production code;
- benchmark;
- run LogN16, Gate4/5 or EXP003;
- production-integrate.

---

# G0 — reproduce accepted system control

Reproduce the current accepted oracle-power system path with ordinary q0/q1 G0 one-bit guard:

- polynomial output near the accepted k=3 values;
- all F0/DA local-q2 contracts pass;
- EvalMod real/imag near accepted values;
- post-S2C near `0.0003364984531681451`;
- public-like near `0.01076789220080715`;
- final metadata valid.

If not, stop:

`logn13_g0_local_q2_guard_sweep_precondition_mismatch`.

---

# G1 — bounded q01 -> q012 expansion before G0 guard

At the exact G0 pre-Rescale guard point:

1. require the unguarded intended value to be Q01-centered-unique;
2. reconstruct exact centered coefficients from q0/q1;
3. materialize q2 from those coefficients;
4. preserve q0/q1 bit-for-bit;
5. place q2 in the correct NTT/Montgomery representation.

Require:

- semantic change <= `1e-10`;
- q2 row matches the stage-aligned full-RNS mirror;
- q0/q1 rows remain unchanged.

If not, classify:

`logn13_g0_local_q2_expansion_mismatch`.

---

# G2 — G0 local-q2 guard sweep

Test exactly these G0 guard-bit counts under Q012-authoritative semantics:

`2, 3, 4`.

For each candidate k:

1. physically multiply q0/q1/q2 residues by `2^k`;
2. multiply Scale by `2^k`;
3. require represented-value preservation <= `1e-10` under Q012;
4. require Q012 centered uniqueness;
5. allow Q01 alias inside the bounded local-q2 G0 region;
6. perform the unchanged logical G0 Rescale with the same divisor and Level transition;
7. require exact rounded-division equality and q012 row equality versus stage-aligned full-RNS;
8. after Rescale require Q01 centered uniqueness;
9. contract immediately back to q0/q1 with semantic equality <= `1e-10` and unchanged intended metadata.

Stop increasing k if Q012 capacity or post-Rescale Q01 contraction fails.

Record for each k:

- Q01/Q012 pre/post-guard capacity ratios;
- local G0 Rescale rounding error versus ideal arithmetic (diagnostic only);
- post-G0 cumulative polynomial error;
- full final polynomial real/imag/max error after the fixed F0 k=3 path.

Do not reject a candidate solely because a local ideal-arithmetic rounding metric exceeds `1e-10` when full-RNS rounded division and rows match exactly.

---

# G3 — full downstream evaluation for every valid candidate

For each valid G0 guard candidate k = 2..4, run the already-validated downstream path unchanged:

1. fixed F0 local-q2 guard k=3;
2. fixed all-three-round DA bounded local-q2 design;
3. unchanged S2C;
4. unchanged one-input finalization/public-like path.

Record:

- final polynomial real/imag/max error;
- final EvalMod real/imag max-component error;
- post-S2C max-component error;
- public-like max-component error;
- final metadata correctness;
- all F0/DA q012 contract pass flags.

System success criterion:

`public_like_max_component_error <= 1e-2`.

Select the **smallest G0 guard-bit count** that meets the system criterion with all contracts valid.

Do not require the historical `1.2e-8` local polynomial heuristic budget to pass; record it separately as a diagnostic field.

---

# G4 — decision

If any candidate passes the system target, classify:

`logn13_g0_local_q2_guard_downstream_validated`

and record:

`PS_DA_ARITHMETIC_SYSTEM_SUFFICIENT = true`.

The next remaining blocker is then the separately-proven generated-power semantic error. Do not production-integrate in this task.

If no valid G0 candidate passes, classify exactly one:

- `logn13_g0_local_q2_guard_insufficient_downstream_precision`
- `logn13_g0_local_q2_guard_capacity_failure`
- `logn13_g0_local_q2_guard_contraction_failure`

Record the best valid candidate and exact first remaining blocker.

---

# Static cost accounting

Do not benchmark.

For the selected/best candidate record:

- G0 q01->q012 coefficient-transform count;
- q2 NTT/INTT count;
- q2 limb arithmetic passes;
- confirmation that q2 exists only for the bounded G0 guard/Rescale region and is discarded immediately afterward;
- total static local-q2 regions when combined with the fixed F0 and DA design.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-G0-LOCAL-Q2-GUARD-SWEEP-summary.json`

Include only:

- provenance;
- G0 system control;
- one compact row per tested G0 guard k=2..4;
- G0 expansion/row/rescale/contraction pass flags;
- polynomial/EvalMod/post-S2C/public-like errors for each valid k;
- selected smallest passing k or best valid candidate;
- final metadata;
- static q2 cost summary;
- classification;
- `PS_DA_ARITHMETIC_SYSTEM_SUFFICIENT`;
- first remaining blocker;
- validation flags.

Do not serialize vectors, coefficient arrays, full RNS rows, or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*G0.*Local.*Q2.*Guard.*Sweep` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no generated-power redesign;
- no retuning outside G0 guard bits;
- no q3+;
- no q0/q1 widening;
- no extra Q levels;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.