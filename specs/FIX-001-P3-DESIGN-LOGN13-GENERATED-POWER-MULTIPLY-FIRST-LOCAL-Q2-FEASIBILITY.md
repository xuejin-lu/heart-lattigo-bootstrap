# FIX-001-P3-DESIGN-LOGN13-GENERATED-POWER-MULTIPLY-FIRST-LOCAL-Q2-FEASIBILITY

## Goal

Determine whether the remaining LogN13 generated-power error is caused by the current balanced **pre-Rescale operands -> multiply** schedule, and test the smallest capacity-safe alternative:

> temporarily extend the generated-power multiplication to q0/q1/q2, multiply first at the native high operand scales, apply the Chebyshev doubling/correction schedule in the correct order, perform one centered rounded Rescale by the native current-level divisor, and return to q0/q1.

This is a **design-feasibility** task. It must first perform real canonical-reset attribution; the previous task collected 88 checkpoint errors but explicitly deferred canonical reset/replay.

Do not production-integrate in this task.

---

# Accepted evidence / motivation

Primary accepted generated-power re-entry result:

`e7e677c5e6d78e8fdea888e57a55455215e7c064`

Secondary exact baseline:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Classification:

`logn13_generated_power_multi_power_blocker`

Fixed D0 oracle-power control:

- q0/q1/q2 = 56/39/40;
- common plan scale = `2^92`;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DoubleAngle local-q2 rounds;
- B4 final-parent T2 one-bit scalar guard;
- unchanged G0 merge / S2C / finalization;
- public-like = `0.005554220603853743`.

Generated-power evidence:

- `SG(T2)` passes: ~`0.00598004`;
- `SG(T3)` passes: ~`0.00555422`;
- `SG(T4)` fails: ~`0.02018555`;
- `SG(T6)` fails: ~`0.01378813`;
- `SG(T8)` fails: ~`0.17488021`;
- `SG(T16)` fails: ~`0.02204507`;
- all-generated `G_ALL` fails: ~`0.22247498`;
- no single oracle rescue isolates the all-generated failure;
- cumulative `[2]`, `[2,3]` pass; `[2,3,4]` is first cumulative failure.

For causal powers T4/T6/T8/T16:

- Fast F == full-RNS mirror N at completed power checkpoints;
- q0/q1 rows match;
- centered uniqueness passes;
- completed-power q01 capacity ratios are only ~`5.6e-11` to `5.8e-11`;
- N-Q / F-Q is nonzero, from ~`5e-8` up to ~`9e-7`.

Therefore the blocker is **not** current q0/q1 storage mismatch and **not** completed-power q01 capacity. It is generated-power arithmetic schedule / rounding accumulation.

T4 checkpoint evidence is especially important:

- input residual ~`6.9e-9`;
- after each balanced operand Rescale ~`1.43e-8`;
- raw product ~`2.87e-8`;
- after Chebyshev doubling ~`5.73e-8`;
- completed T4 ~`5.73e-8`.

The current Secondary source uses a balanced capacity-safe schedule:

1. choose integer factors near sqrt(current q-level divisor);
2. multiply left/right operands by those integer factors;
3. Rescale **both operands separately**;
4. multiply the two lower-scale operands;
5. Chebyshev double;
6. relabel output Scale to the planned target;
7. apply the Chebyshev recurrence correction.

Git history shows pre-Rescale scheduling was introduced specifically to avoid q0/q1 capacity failure. Therefore this task must **not** blindly revert to q01 multiply-first.

---

# Repository provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required base/ancestor:

`e7e677c5e6d78e8fdea888e57a55455215e7c064`

Start only from clean `main`. Fetch and ff-only pull, then re-read fresh `AGENTS.md`, `CURRENT_TASK.md`, and this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Branch: `fast-ckks`, clean.

**No Secondary production modification is authorized in this task.**

Primary-only diagnostic/design helpers may be added.

---

# Scope lock

LogN13 only.

Every system replay must keep the accepted D0 stack fixed:

- q0/q1/q2 = 56/39/40;
- planScale `2^92`;
- same polynomial and PS split;
- same T2 scalar guard;
- same G0=2 and F0=3 local-q2 behavior;
- same all-round DA local-q2 behavior;
- same S2C and finalization;
- same widened Standard reference;
- same workload/input.

Do not:

- change q parameters;
- widen q0/q1 in this task;
- change coefficients;
- change plan scale;
- redesign PS/G0/DA/S2C;
- sweep arbitrary guard bits;
- modify Secondary production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5, or EXP003.

---

# A0 — reproduce controls

Reproduce at minimum:

- D0 oracle public-like `0.005554220603853743`;
- SG(T4) ~`0.0201855534`;
- SG(T6) ~`0.0137881289`;
- G_ALL ~`0.2224749802`;
- T4 completed F-Q ~`5.73e-8` real / ~`5.75e-8` imag;
- F-N = 0 for T4/T6 under the existing schedule.

If these controls materially mismatch, stop:

`logn13_generated_power_schedule_precondition_mismatch`.

---

# A1 — derive the actual power dependency graph

From the current source/runtime, emit the exact dependency graph for required powers `[2,3,4,6,8,16]`.

Do not assume dependencies from names alone.

For every generated power record:

- split `(a,b)` selected by `commonpolynomial.SplitDegree`;
- whether the operation is lazy;
- input degree/level/scale;
- balanced vs unbalanced branch;
- current-level Rescale divisor bit length;
- output degree/level/scale.

Use this graph to distinguish:

- **root schedule error** created by the power's own operation;
- **inherited error** already present in its parents;
- **descendant amplification**.

The key question is whether T8/T16 are primarily descendants of T4 error and whether T6 is a separate root branch.

---

# A2 — real canonical-reset attribution

The previous task recorded checkpoint distances only. This task must perform actual state reset and replay.

For T4 first, and then T6 if T4 does not explain T6:

## A2.1 parent reset

Construct canonical CKKS parent states at the **same native metadata** as the actual generated parents.

Cases:

- `P0`: actual parents -> current generated operation (baseline);
- `PL`: canonical left only, actual right;
- `PR`: actual left, canonical right only;
- `PB`: canonical both parents.

Execute the unchanged current balanced generation operation and record the completed child power residual vs canonical.

Then insert that completed child into a dependency-consistent power map, regenerate all descendants that depend on it using the current generated schedule, and replay full fixed D0 PS/DA/S2C/finalization.

Record public-like for each reset case.

## A2.2 operation-boundary reset

For T4 current balanced path, reset to canonical at the actual operation metadata at these source-backed boundaries:

1. after integer factor multiplication, before operand Rescale;
2. after left Rescale only;
3. after right Rescale only;
4. after both operand Rescales;
5. raw product;
6. after Chebyshev doubling;
7. after recurrence correction / completed T4.

For each reset, replay the remaining T4 operations, regenerate dependency descendants consistently, then replay the fixed D0 system.

The canonical state must represent the mathematically correct **partial expression at that exact native Level/Scale/metadata**, not merely a decoded-slot replacement at a convenient scale.

Required output:

- latest reset that still fails system threshold;
- first reset that passes system threshold;
- local completed T4 residual and public-like for every reset.

This determines the smallest causal operation interval.

## A2.3 intrinsic-vs-inherited decomposition

For each causal root power identified (at least T4; T6 if independent):

- `actual-input/current-schedule`;
- `canonical-input/current-schedule`;
- `actual-input/canonical-output`;
- `canonical-input/canonical-output`.

This must quantify the power's **intrinsic schedule error** separately from inherited parent error.

Do not infer this from checkpoint norms alone.

---

# A3 — current balanced schedule precision ledger

For the causal root power(s), produce a compact error ledger for current balanced scheduling:

- parent residual before factor multiplication;
- after factor multiplication;
- after each operand Rescale;
- raw product;
- after Chebyshev doubling;
- recurrence correction;
- completed output.

Also record actual scales at every step.

For T4 specifically verify the observed pattern near:

- input Scale ~`2^60`;
- each pre-rescaled operand Scale ~`2^30`;
- product output back near `2^60`.

State explicitly whether each operand Rescale contributes independent CKKS rounding/quantization.

---

# A4 — q01 multiply-first capacity-only oracle

Before executing any q01 multiply-first candidate, compute exact intended physical capacity.

Candidate mathematical schedule for a strict generated power at common level `L`:

1. use the native high-scale parents at level `L`;
2. multiply / MulRelin first;
3. apply Chebyshev doubling if required;
4. apply the correct recurrence ordering required to match Standard semantics;
5. perform **one** rounded Rescale by `Q[L]`;
6. output at level `L-1` with the same target Scale as the planned generated power.

For every pre-Rescale checkpoint compute from exact intended centered coefficient values:

- max `|x|`;
- current Q01/2;
- `max|x|/(Q01/2)`;
- outside count;
- minimum number of Q01 bits required for centered uniqueness:
  `ceil(log2(2*max|x| + 1))` (or mathematically equivalent exact integer computation).

Do this at least for T4 and T6. If dependency graph shows T8/T16 require larger physical range, include them.

### Rule

If any required q01 multiply-first checkpoint has `|x| >= Q01/2`, classify q01 multiply-first as capacity-invalid and **do not execute it modulo q01**.

This capacity evidence is also the quantitative bridge for the future independent research topic “Fast q0/q1 parameter design”; do not change parameters in this task.

---

# A5 — temporary-q2 multiply-first feasibility

Run only if A4 shows q01 multiply-first is capacity-invalid or if source-history constraints justify q2 protection.

Use q0/q1/q2 solely as a **temporary representation container** for generated-power multiplication. q2 is not a new Rescale divisor.

For each selected causal root power:

1. start from parents whose current q0/q1 state is centered-unique;
2. recover the exact centered parent coefficients independently from q0/q1;
3. lift/reconstruct q2 exactly;
4. restore the correct NTT/Montgomery representation for q0/q1/q2;
5. execute the native multiply or MulRelin according to the actual generation plan using q0/q1/q2-authoritative arithmetic or an exact full-RNS-equivalent diagnostic implementation;
6. apply Chebyshev doubling in the same arithmetic representation;
7. apply recurrence correction in the Standard-consistent order;
8. check **Q012 centered uniqueness before every rounded division / contraction that requires it**;
9. perform one centered symmetric rounded Rescale by the actual current-level divisor `Q[L]`;
10. independently verify the rounded quotient against a big.Int/full-RNS oracle;
11. reconstruct q0/q1 (and q2 only while still needed);
12. ensure the post-Rescale q0/q1 state is centered-unique;
13. restore the same output Level/Scale/degree/NTT/Montgomery metadata expected by the generated-power plan.

Important:

- q2 is only extra capacity; the Rescale divisor remains `Q[L]`;
- do not use modular inverse division as a substitute for centered rounded Rescale;
- do not assume Q012 fits 128 bits; q0/q1/q2 here can exceed 128 total bits, so diagnostic big.Int / genuine RNS arithmetic is acceptable;
- compare q0/q1/q2 rows against an independent full-RNS mirror wherever meaningful.

Required capacity evidence:

- Q01 ratios from A4;
- Q012 ratios for the same multiply-first checkpoints;
- outside counts;
- row agreement;
- post-Rescale q01 uniqueness.

If Q012 itself is not unique at a required pre-Rescale checkpoint, stop that candidate:

`logn13_generated_power_multiply_first_q012_capacity_blocker`.

---

# A6 — candidate comparison on root powers

For each causal root power (expected at least T4 and possibly T6), compare:

### B — current balanced pre-Rescale schedule

Current Secondary behavior.

### M — multiply-first with temporary q2 and one Rescale

A5 candidate.

At identical planned output metadata record:

- completed power vs canonical max component / mean / worst index;
- improvement factor;
- F/full-RNS row agreement;
- capacity contracts;
- number of operand Rescales before multiplication;
- total rounded Rescales used for that power;
- decoded semantic agreement.

Do not require arbitrary `1e-10` local perfection. System threshold decides sufficiency.

---

# A7 — dependency-aware system replay

If A6 shows the temporary-q2 multiply-first schedule is valid and improves the causal root powers, replay real generated powers under the fixed D0 system.

Use the candidate only at source-backed causal root/descendant nodes justified by A1/A2; do not silently alter unrelated power nodes.

Run staged cases:

1. corrected T4 branch only, regenerate its descendants from corrected T4;
2. corrected T6 branch only if independent;
3. corrected T4 + T6 roots, regenerate all dependent T8/T16 consistently;
4. all required generated powers with the accepted candidate schedule wherever causally justified.

Record for every case:

- generated power completed residuals;
- polynomial residual;
- EvalMod real/imag;
- post-S2C;
- public-like;
- public margin to `1e-2`;
- all capacity/metadata contracts.

Primary system pass criterion:

`public_like <= 1e-2`.

If the all-generated candidate passes, do not production-integrate in this task.

---

# A8 — relation to future q0/q1 parameter research

Add a compact evidence section only; do not alter q values.

Record:

- current actual q0/q1 bit lengths;
- current Q01 bit length;
- minimum Q01 bit length required by multiply-first T4/T6/T8/T16 checkpoints from A4;
- extra bits beyond current Q01 required, if any;
- whether temporary Q012 clears the same bound.

This section must distinguish:

- **current blocker**: generated-power schedule/rounding under the fixed D0 profile;
- **future architecture question**: whether a permanent two-limb Fast representation should use wider q0/q1 so multiply-first can be safe without temporary q2.

Do not claim wider q0/q1 automatically fixes precision unless A4/A6 evidence demonstrates the mechanism.

---

# Classification

Choose exactly one primary classification:

- `logn13_generated_power_pre_rescale_quantization_blocker`
- `logn13_generated_power_multiply_first_q01_capacity_blocker`
- `logn13_generated_power_multiply_first_q012_capacity_blocker`
- `logn13_generated_power_local_q2_multiply_first_system_sufficient`
- `logn13_generated_power_local_q2_multiply_first_insufficient`
- `logn13_generated_power_schedule_attribution_mismatch`

If a more precise source-operation interval is found, include it as a separate field, not by inventing an unrelated classification.

Recommended-next-target must be one of:

- production design/integration of the validated generated-power schedule, only if system sufficient;
- a narrower generated-power arithmetic design task if insufficient;
- q0/q1 parameter architecture study only if evidence shows capacity remains the limiting issue after schedule attribution.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-GENERATED-POWER-MULTIPLY-FIRST-LOCAL-Q2-FEASIBILITY-summary.json`

Include:

- provenance;
- A0 controls;
- dependency graph;
- canonical-reset table for T4/T6 roots;
- intrinsic-vs-inherited decomposition;
- current balanced precision ledger;
- q01 multiply-first capacity-only table;
- minimum required Q01 bits;
- q012 multiply-first capacity/contracts;
- B vs M local comparison;
- dependency-aware system replay table;
- classification;
- first remaining blocker;
- recommended next target;
- validation flags.

Do not serialize full slot vectors, coefficient arrays, or full RNS rows.

---

# Validation

Primary:

- focused tests matching `TestFIX001P3.*Generated.*Power.*Multiply.*First.*Q2` (or a close source-backed equivalent);
- `go test ./...`;
- `git diff --check`;
- artifact no NaN/Inf;
- D0 reproduction exact within established deterministic tolerance;
- canonical resets are actual reset+replay, not checkpoint-distance-only;
- q01 unsafe candidates are never executed after capacity failure;
- q012 candidate uses independent rounded-divide/full-RNS oracle;
- fixed D0 stack preserved in every system replay;
- Secondary remains exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean, unmodified;
- no q parameter change;
- no production integration;
- no LogN16, benchmark, Gate4/5, EXP003;
- final Primary worktree clean after ordinary ff push.
