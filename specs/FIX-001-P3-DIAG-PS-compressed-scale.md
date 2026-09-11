# FIX-001-P3-DIAG-PS-COMPRESSED — q0/q1-safe internal PS scale feasibility

## Purpose

The formal P3 PS first-divergence diagnosis at Primary commit

`8222fd9f6addc479bf87626ee4c24efdc9287fd9`

established:

- all formal generated powers pass under Secondary `61607bb4bb82591009ce768d9a3773bed1497565`;
- PS baby blocks 0 and 1 pass;
- the first failure is baby block 2, `constant_add`, coefficient index 0;
- block-2 accumulator Level = 9;
- block-2 accumulator Scale ≈ `2^120`;
- constant ≈ `0.03665425157863994`;
- local semantic error ≈ `0.03665425827255296`;
- `q0*q1/2 = 9903518838475962595646701568`;
- the actual q0/q1 centered residue remains inside the centered interval, but the **nominal encoded constant** is about `4.92e6 * (q0*q1/2)`;
- therefore the high-scale scalar constant cannot be uniquely represented by only q0/q1.

The production Fast scalar `Add` encodes a scalar directly at the ciphertext's current Scale. Standard PS can do this at ~`2^120` because it retains the full RNS basis; Fast cannot.

Do **not** special-case only this constant. The same ~`2^120` baby accumulator scale also governs subsequent non-integer coefficient `MulThenAdd` terms, so this task must test whether the whole Fast PS evaluation can instead use a lower q0/q1-safe **internal representation scale**, while preserving the public target scale at the end.

This is diagnostic/feasibility work only. Do not modify production Lattigo.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when authored:

`8222fd9f6addc479bf87626ee4c24efdc9287fd9`

Secondary repository: `xuejin-lu/lattigo`

Exact P3 production commit:

`61607bb4bb82591009ce768d9a3773bed1497565`

Pinned Standard reference:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md` and Secondary `docs/FAST_CKKS_SPEC.md`.

Both worktrees must begin clean. Do not reset/stash/discard unrelated work.

Parent FIX-001/P3 remains incomplete.

---

# Core hypothesis

The Standard PS plan uses baby-block scales around `2^120` so that a final Rescale yields the requested polynomial output scale around `2^60`.

For the two-limb Fast backend, that internal scale is unnecessarily large and causes scalar coefficients to alias modulo `q0*q1`.

Instead evaluate the PS structure at a lower internal scale `S_fast` satisfying:

1. baby-block scalar constants/terms fit q0/q1 with margin;
2. non-integer coefficient encoding still has enough precision;
3. giant-step recurrence remains scale-stable;
4. final Fast result can be safely promoted back to the original target Scale using integer multiplication + metadata scale update, without an additional level transition.

For the formal profile:

- power scales are approximately `2^60`;
- choose candidate baby/giant internal scales near `2^90`;
- non-integer coefficient encoding scale becomes approximately `2^30`;
- giant step `Rescale(b)` gives approximately `2^30`, then multiplication by an approximately `2^60` power returns to approximately `2^90`;
- final Rescale yields approximately `2^30`;
- a final integer promotion of approximately `2^30` can restore the public target near `2^60` while preserving the represented value.

This task must prove or reject that complete schedule on the exact formal degree-30 polynomial.

---

# Preconditions

Reproduce the exact formal LogN13 real-branch E2 state.

Require before interpretation:

- E2 exact historical reproduction;
- corrected `T0=1` oracle;
- actual P3 generated powers all pass <= `1e-2`:
  - T1
  - T2
  - T3
  - T4
  - T6
  - T8
  - T16;
- ordinary P3 public Fast polynomial evaluation reproduces the known whole-polynomial failure around `0.7909176408709179`.

If any precondition differs, stop with:

`COMPRESSED_PS_PRECONDITION_MISMATCH`.

---

# Diagnostic access policy

No committed Secondary production changes.

Temporary, uncommitted, build-tagged helpers in package `polynomial` are allowed to:

- obtain the actual Standard/common PS plan;
- obtain the actual P3 generated powers;
- execute the real Fast baby/giant primitives with **diagnostically overridden accumulator scales**;
- snapshot maintained q0/q1 state.

Do not replace Fast arithmetic with plaintext evaluation as the tested path.

Plaintext/source-backed evaluation is only the oracle.

Write temporary raw evidence to `/tmp`; remove every helper before completion and restore exact clean Secondary `61607bb...`.

---

# Part 1 — record the unmodified plan and scale problem

Archive the exact formal PS plan:

- degree/base;
- baby block count;
- each block Level/Scale/degree/parity;
- non-nil coefficients;
- reversal order;
- giant-step pairing schedule;
- public targetScale.

For every scalar constant or coefficient term that would be encoded at a plan baby scale, compute a **nominal scalar representation bound**:

`abs(coefficient) * planScale`

and compare with `Q01/2`.

This is not a full ciphertext product bound; it is specifically the scalar-encoding feasibility check.

Record which plan scalar injections are intrinsically non-unique under q0/q1.

The historical block-2 constant must reproduce the `~4.92e6` over-capacity ratio.

---

# Part 2 — deterministic candidate internal scales

Let:

`Qhalf = floor(q0*q1/2)`.

Define candidate internal scale bits from q0/q1 only, not from ciphertext contents:

- `floor(log2(Qhalf)) - 2`
- `floor(log2(Qhalf)) - 3`
- `floor(log2(Qhalf)) - 4`
- `floor(log2(Qhalf)) - 5`
- `floor(log2(Qhalf)) - 6`
- `floor(log2(Qhalf)) - 7`

For the formal profile these should cover approximately `2^91 ... 2^86`; explicitly include exact `2^90` even if duplicated by the rule.

Do not use arbitrary floating scale values when an exact power-of-two candidate can be represented.

For each candidate `S_fast`, before arithmetic compute:

### Constant feasibility
For each baby-block constant:

`abs(c0) * S_fast < Qhalf`

### Coefficient precision
For each non-integer `coefficient * power[key]` term, estimate the scalar encoding scale actually needed by current Fast `MulThenAdd`:

`coeffScale ~= S_fast / power[key].Scale`

Record minimum coefficient-scale log2 across all terms.

A candidate is rejected before replay if either:

- any known scalar constant nominal encoding exceeds `Qhalf`; or
- minimum non-integer coefficient encoding scale is below `2^20`.

The `2^20` floor is diagnostic and fixed for this experiment; do not silently relax it.

---

# Part 3 — compressed baby-step replay

For each surviving `S_fast`, replay **all** baby blocks with the actual plan Levels and actual P3 powers, but replace each baby accumulator's plan Scale with the candidate common internal scale `S_fast`.

Do not alter levels or coefficients.

Use the real production operations:

- zero accumulator;
- scalar `Add` for constant;
- `MulThenAdd(power[key], coefficient, accumulator)` in exact production key order.

For every operation:

- compare against the local semantic oracle from actual decoded inputs;
- record metadata;
- record q0/q1 centered capacity for c0;
- for scalar injection record nominal encoded magnitude / Qhalf;
- require <= `1e-2`.

A candidate fails immediately if any baby operation fails.

Classify per candidate:

- `baby_constant_capacity_failure`
- `baby_multhenadd_capacity_failure`
- `baby_coefficient_precision_failure`
- `baby_arithmetic_failure`
- `baby_all_pass`.

---

# Part 4 — compressed giant-step replay

Only for candidates where **all baby blocks pass**.

Replay the exact PS reversal and giant-step pairing schedule using the compressed baby outputs.

For each `evaluateMonomial` equivalent:

1. optional actual Fast Relinearize(b);
2. actual Fast Rescale(b);
3. actual Fast Mul(b, xpow, b);
4. scale-closeness check against `a`;
5. actual Fast `addAligned(a,b)`.

Do not force metadata to match if the real operations do not naturally produce compatible scales; record the mismatch.

At every step use local semantic oracle and q0/q1 capacity accounting where a centered CRT Rescale is involved.

Expected scale invariant to test:

`S_fast --Rescale(~2^60)--> ~S_fast/2^60 --Mul(power ~2^60)--> ~S_fast`.

Classify the first failure per candidate:

- `giant_relinearize_failure`
- `giant_rescale_capacity_failure`
- `giant_rescale_failure`
- `giant_multiply_capacity_failure`
- `giant_multiply_failure`
- `giant_scale_alignment_failure`
- `giant_addaligned_failure`
- `giant_all_pass`.

---

# Part 5 — compressed finalization

Only for candidates with all giant steps passing.

Run the same final production sequence:

1. optional final Relinearize;
2. final Fast Rescale.

Compare the post-final-Rescale semantic value against the source-backed degree-30 polynomial oracle.

Record resulting:

- Level;
- Scale exact/log2;
- component error;
- q0/q1 capacity.

At this point the value must pass <= `1e-2` even though Scale is expected to be lower than the public target.

If it fails, classify:

- `final_rescale_capacity_failure`
- `final_rescale_failure`
- `compressed_polynomial_semantic_failure`.

---

# Part 6 — safe final target-scale promotion

For a candidate whose compressed polynomial value passes, restore the original public target Scale without consuming another level.

Let:

`scaleRatio = targetScale / compressedOutput.Scale`.

Record exact high-precision ratio.

Choose integer:

`M = nearest positive integer(scaleRatio)`

using high-precision arithmetic.

On an independent copy:

1. `MulIntegerMaintained(result, M, result)`;
2. `result.Scale *= M`;
3. require resulting Scale is within existing CKKS scale tolerance of public `targetScale`;
4. if close, normalize metadata to exact `targetScale`;
5. decode and compare again to polynomial oracle.

Before accepting promotion, prove q0/q1 safety:

- recover centered c0 coefficients before promotion;
- bound/compute coefficient magnitudes after multiplying by M;
- require every maintained component remains uniquely representable under Qhalf where required by subsequent Fast semantics;
- record max Qhalf ratio.

If promotion capacity fails:

`final_integer_promotion_capacity_failure`.

If semantic value changes beyond threshold:

`final_integer_promotion_precision_failure`.

If promoted result passes:

`compressed_ps_candidate_pass`.

---

# Part 7 — choose the preferred candidate

Among candidates that pass the complete polynomial + final promotion path, choose deterministically:

1. highest `S_fast` that remains capacity-safe at every observed scalar/ciphertext checkpoint;
2. if tied, smallest whole-polynomial max component error;
3. require minimum coefficient encoding scale >= `2^20`.

Record the safety headroom:

- max observed `abs(centered coefficient)/(Qhalf)`;
- max nominal scalar encoding ratio;
- minimum coefficient encoding log2;
- final promotion capacity ratio.

Do not infer a universal input-independent bound from this one workload. This is a formal-profile feasibility result.

---

# Part 8 — optional control: exact historical plan scale

As a consistency check, replay the existing plan scale through the same diagnostic harness and require it reproduces the known block-2 constant failure.

This validates that the harness itself is capable of exposing the historical bug.

---

# Required overall classification

Choose one:

## Case A

`FIRST_SUPPORTED_CAUSE = compressed_ps_scale_validated`

At least one deterministic lower internal scale completes:

- all baby blocks;
- all giant steps;
- final Rescale;
- final target-scale integer promotion;
- whole degree-30 polynomial <= `1e-2`.

## Case B

`FIRST_SUPPORTED_CAUSE = no_safe_internal_scale_with_precision_floor`

No candidate simultaneously satisfies q0/q1 capacity and >=`2^20` coefficient encoding precision.

## Case C

`FIRST_SUPPORTED_CAUSE = compressed_baby_steps_still_fail`

At least one candidate passes prechecks but no candidate gets through all baby blocks.

## Case D

`FIRST_SUPPORTED_CAUSE = compressed_giant_steps_fail`

Baby blocks can be made correct but giant-step semantics fail.

## Case E

`FIRST_SUPPORTED_CAUSE = compressed_finalization_fail`

Baby + giant pass but final Rescale/promotion fails.

Do not repair production code in this task.

---

# Required artifacts

After restoring Secondary exact/clean at `61607bb...`, create in Primary:

- `results/FIX-001-P3-DIAG-PS-COMPRESSED-logN13-fast.json`
- `results/FIX-001-P3-DIAG-PS-COMPRESSED-logN13-summary.json`

Summary must include:

- exact provenance;
- historical block-2 failure reproduction;
- exact candidate scale list;
- per-candidate precheck table;
- per-candidate baby/giant/finalization outcome;
- selected preferred candidate if any;
- all relevant capacity/precision headroom;
- whole-polynomial error before final promotion;
- whole-polynomial error after final promotion;
- exact final target Scale comparison;
- first supported cause;
- Primary/Secondary test and clean-state evidence.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- Secondary `go test ./...` passes after temporary helpers are removed;
- Secondary remains exact `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Lattigo code modification;
- Primary artifacts committed/pushed normally;
- both worktrees clean;
- no formal Gate 4/5 production repair run;
- no LogN16;
- no benchmark;
- no EXP-003.

---

# Non-goals

Do not:

- special-case block 2 or coefficient 0;
- modify scalar Add/MulThenAdd in production;
- change the source polynomial;
- change Mod1 degree/K/DoubleAngle/LogScale;
- change power generation;
- add q2+ or auxiliary limbs;
- lower the user-visible/public target Scale permanently;
- loosen `1e-2`;
- run LogN16;
- benchmark;
- start EXP-003.

The deliverable is whether a lower q0/q1-safe internal PS scale can preserve the exact formal polynomial semantics and restore the original public target Scale at the boundary.