# FIX-001-P2-DIAG — Balanced pre-Rescale feasibility for formal T3

## Purpose

`FIX-001-P2` attempted to avoid the proven T3 q0/q1 multiplication overflow by moving the planner-required Rescale to one operand before multiplication.

The production implementation at Secondary commit `ceb5482...` added a heuristic gate:

```text
preRescaleThreshold = rescaleFactor * 2^10
preRescale = chosen.Scale > preRescaleThreshold
```

For the formal T3 path, T1/T2 scales are both approximately `2^60` while the common-level rescale divisor is also approximately 60–61 bits. The gate is therefore false, so formal T3 falls back to the old unsafe `multiply -> double -> Rescale` schedule and reproduces the exact historical error.

Simply removing the gate is not a valid repair: one-sided pre-Rescale would reduce a `~2^60` operand by a `~2^60` modulus to scale near one, destroying precision.

This task must test a more principled q0/q1-only schedule **without modifying production code**:

> safely multiply each operand by a bounded integer, rescale both operands in parallel at the same common logical level, then multiply the two post-Rescale operands so the final product scale remains near the planner's expected post-product scale without ever forming the original high-scale product.

This is a feasibility/diagnostic task, not a repair task.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when authored:

`f5666f124fcc227cce9af3880b46b6b8f0bbb22d`

Secondary repository: `xuejin-lu/lattigo`

Current Secondary HEAD:

`f2b89ed0ade9efb5f99fce3949ffeba87cd3d2fd`

Production P2 implementation commit within that history:

`ceb5482c1f5e5ffaccd4d09ec835ea6c1bd4f04b`

Prior repaired Fast base:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Standard reference remains:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Both worktrees must be clean before starting. Follow both repositories' `AGENTS.md`.

Do not revert or amend the existing incomplete P2 commits in this task.

---

## Formal workload

Use exactly the same formal LogN13 real-branch E2/T1/T2 state as the corrected diagnostics:

- LogN = 13
- 4096 slots
- `configs/bootstrap_config.logN13.json`
- same deterministic input
- same Q2 real branch
- same E1 scale reinterpretation
- same E2 offset
- corrected Chebyshev oracle with `T0=1`

Require:

- E2 matches historical corrected vector;
- T1 passes;
- T2 passes.

Use the actual formal production T1/T2 ciphertexts produced at current Secondary HEAD.

Do not run LogN16.

---

# Background identity

Old desired power transition at common level `L` is:

```text
out = Rescale(left * right)
```

with expected scale approximately:

`Delta_target = Delta_left * Delta_right / q_L`.

A one-sided pre-Rescale gives one operand scale approximately `Delta/q_L ~ 1`, which is too small.

Instead consider two parallel pre-Rescales:

```text
left'  = Rescale(m1 * left)
right' = Rescale(m2 * right)
out    = left' * right'
```

where integer multiplication also updates the operand metadata scale by the same integer factor before Rescale.

Then:

`Scale(left')  = Delta_left  * m1 / q_L`

`Scale(right') = Delta_right * m2 / q_L`

and:

`Scale(out) = Delta_left * Delta_right * m1*m2 / q_L^2`.

To reproduce the old target scale, choose:

`m1*m2 ~= q_L`.

Both Rescales occur in parallel copies from logical level L to L-1, so the resulting product remains at level L-1. Logical depth consumption is still one level even though two Rescale operations execute.

---

# Part 1 — prove why current P2 gate does not fire

Record for formal T3:

- T1 Level and exact Scale;
- T2 Level and exact Scale;
- commonLevel;
- exact logical rescale divisor `q_L`;
- `rescaleFactor` used by current P2 implementation;
- current `preRescaleThreshold = rescaleFactor * 2^10`;
- chosen operand under current deterministic rule;
- exact comparison result;
- current `preRescale` boolean.

Require evidence that current formal path has `preRescale=false` before proceeding.

If current production unexpectedly takes `preRescale=true`, stop with `P2_PATH_MISMATCH`.

---

# Part 2 — recover centered q0/q1 coefficients and safe multiplier bounds

Reuse the validated coefficient recovery from `FIX-001-DIAG-T3-CAPACITY`.

For each actual T1/T2 maintained component, at minimum c0 and preferably all components:

- convert q0/q1 to ordinary coefficient domain;
- centered-CRT reconstruct signed coefficients;
- record max absolute coefficient.

Let:

`Qhalf = floor(q0*q1/2)`.

For each operand/component compute the maximum positive integer multiplier that keeps every coefficient strictly in the centered unique range:

`Mmax = floor((Qhalf-1) / maxAbsCoeff)`.

Record exact integer Mmax values.

For semantic zero-secret correctness, c0 is authoritative. However also report c1 bounds and whether a proposed schedule is safe for all maintained components.

Do not use float64 for bound decisions.

---

# Part 3 — mathematical feasibility of balanced factors

For c0 first evaluate the necessary condition:

`Mmax_left * Mmax_right >= q_L`.

If false, then no pair of integer pre-scale multipliers can simultaneously:

1. keep both multiplied operands uniquely representable under q0/q1, and
2. satisfy `m1*m2 ~= q_L` closely enough to preserve the old target scale.

Classify immediately:

`BALANCED_PRE_RESCALE_FEASIBLE = false`

and do not execute an unsafe arithmetic prototype.

Also report the same condition using the stricter all-components Mmax values.

If c0 feasibility is true, continue.

---

# Part 4 — choose a deterministic safe factor pair

Choose integers `(m1,m2)` satisfying:

- `1 <= m1 <= Mmax_left`
- `1 <= m2 <= Mmax_right`
- product is as close as practical to `q_L`
- both post-Rescale operand scales remain comfortably above a minimum precision floor.

Preferred selection:

1. target `sqrt(q_L)` balance;
2. clamp by each operand's Mmax;
3. choose the complementary integer by high-precision nearest division;
4. search a small deterministic neighborhood if needed to minimize relative product error while satisfying both bounds.

Do not perform an unbounded scan.

Record:

- m1, m2;
- product `m1*m2`;
- signed difference from `q_L`;
- relative product error;
- exact predicted post-Rescale scales;
- exact predicted product scale;
- old planner target scale;
- predicted product-scale / target-scale ratio.

Minimum diagnostic precision floor:

- each post-Rescale operand Scale should be >= `2^20` unless evidence justifies a lower threshold.

If no factor pair meets both capacity and precision floor, classify balanced schedule infeasible.

---

# Part 5 — arithmetic prototype on independent copies

Only if Part 3/4 establishes feasibility.

Do **not** modify production polynomial code.

Create independent maintained copies of formal T1 and T2 at exact commonLevel.

For left copy:

1. `MulIntegerMaintained(left, m1, left)`
2. `left.Scale *= m1`
3. verify multiplied pre-Rescale coefficients remain strictly within Qhalf
4. `Rescale(left,left)`

For right copy:

1. `MulIntegerMaintained(right, m2, right)`
2. `right.Scale *= m2`
3. verify multiplied pre-Rescale coefficients remain strictly within Qhalf
4. `Rescale(right,right)`

Require both outputs at `commonLevel-1` (or the parameterized equivalent).

Decode both post-Rescale copies through validated q0/q1 projection and compare against their original actual T1/T2 decoded vectors.

Record full 4096-slot metrics.

Each must pass `1e-2` before proceeding.

If either operand loses unacceptable precision, classify:

`FIRST_SUPPORTED_CAUSE = balanced_operand_rescale_precision_failure`.

---

# Part 6 — balanced T3 product prototype

Using the two validated post-Rescale copies:

1. strict `MulRelin(left', right')`
2. Chebyshev doubling `Add(out,out,out)`
3. **do not Rescale again**

The output level must already be the old planned post-product level.

Decode and compare against local oracle:

`2 * T1_actual * T2_actual`.

Require component threshold `1e-2`.

Also perform coefficient-capacity accounting on the new product c0 and report:

- max absolute coefficient;
- Qhalf ratio;
- outside-Q01 count.

If product still crosses q0/q1 capacity, classify:

`FIRST_SUPPORTED_CAUSE = balanced_product_still_crosses_q01_capacity`.

If product stays in range but numerical output fails, classify:

`FIRST_SUPPORTED_CAUSE = balanced_product_arithmetic_failure`.

---

# Part 7 — planner-scale normalization control

The actual balanced product scale will generally differ slightly from the old exact planner target because `q_L` is prime and `m1*m2` need not equal it exactly.

Create two independent diagnostic interpretations:

### Actual-scale interpretation

Decode using the actual computed balanced product Scale.

### Planner-scale interpretation

Copy the ciphertext and set metadata Scale to the exact old planned target:

`Delta_left * Delta_right / q_L`.

Do not modify coefficients.

Decode again.

Record:

- actual-scale metrics vs semantic oracle;
- planner-scale metrics vs semantic oracle;
- difference between the two decoded vectors.

Planner-scale normalization is acceptable for a future repair only if it remains <= `1e-2` and the scale-ratio error is demonstrably tiny.

---

# Part 8 — full T3 correction prototype

If the doubled balanced product passes, apply the existing repaired c>0 post-product correction path against formal T1:

`T3 = doubledBalancedProduct - T1`

using the same scale-alignment mechanism production would use.

Decode all slots and compare against corrected local oracle:

`2*T1_actual*T2_actual - T1_actual`.

Require component threshold `1e-2`.

This task may diagnose scale-alignment failure but must not repair it.

---

# Part 9 — T2 control

Prototype the same balanced strategy for formal T2 (`T1*T1`) only if doing so is required by the prospective general rule.

Compare against current repaired T2 behavior.

A future production rule must not regress T2.

If a general balanced rule is unnecessary for T2 and would worsen it, record evidence supporting a deterministic scheduling condition based on scale/capacity-safe static metadata rather than ciphertext-content scans.

Do not introduce a production heuristic in this task.

---

# Classification

Select one:

## Case A — safe factors do not exist

`FIRST_SUPPORTED_CAUSE = balanced_pre_rescale_mathematically_infeasible`

This means q0/q1-only multiplication at the formal schedule likely requires a different architecture (for example a temporary auxiliary modulus or a different polynomial/scale schedule). Do not implement such an architecture here.

## Case B — factors exist but operand precision fails

`FIRST_SUPPORTED_CAUSE = balanced_operand_rescale_precision_failure`

## Case C — operands pass but product still exceeds capacity

`FIRST_SUPPORTED_CAUSE = balanced_product_still_crosses_q01_capacity`

## Case D — product capacity is safe but arithmetic/oracle fails

`FIRST_SUPPORTED_CAUSE = balanced_product_arithmetic_failure`

## Case E — product passes but correction fails

`FIRST_SUPPORTED_CAUSE = balanced_t3_correction_failure`

## Case F — complete balanced T3 passes

`FIRST_SUPPORTED_CAUSE = balanced_pre_rescale_validated`

Record whether actual-scale and exact-planner-scale interpretations both pass.

---

# Required artifacts

Use `/tmp` while any temporary Secondary helper exists.

After Secondary is restored exact/clean at `f2b89ed...`, create in Primary:

- `results/FIX-001-P2-DIAG-BALANCED-logN13-fast.json`
- `results/FIX-001-P2-DIAG-BALANCED-logN13-summary.json`

Summary must include:

- exact provenance;
- proof current P2 preRescale gate is false;
- T1/T2 coefficient maxima;
- safe multiplier maxima;
- feasibility products;
- selected m1/m2 if feasible;
- predicted and actual post-Rescale scales;
- operand preservation metrics;
- balanced product capacity statistics;
- doubled-product metrics;
- actual-scale vs planner-scale metrics;
- full T3 metrics;
- classification;
- clean/test state.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- Secondary `go test ./...` passes after temporary helpers are removed;
- Secondary exact HEAD remains `f2b89ed0ade9efb5f99fce3949ffeba87cd3d2fd`;
- Secondary worktree clean;
- no production Lattigo changes;
- artifacts committed/pushed normally in Primary;
- no LogN16;
- no benchmark;
- no EXP-003.

---

# Non-goals

Do not:

- remove the current heuristic gate in production;
- implement balanced scheduling in production;
- revert P2 commits;
- add q2+ or auxiliary limbs;
- change Mod1 parameters;
- lower global Mod1 scale;
- loosen thresholds;
- run LogN16;
- benchmark;
- start EXP-003.

The deliverable is a mathematical and empirical answer to whether balanced two-sided pre-Rescale can preserve formal T3 semantics within the q0/q1-only architecture.