# FIX-001-P3-DESIGN-LOGN13-GENERATED-POWER-NATIVE-Q012-MULTIPLY-FIRST-FEASIBILITY

## Goal

Validate the smallest **actually capacity-correct** generated-power multiply-first design for LogN13:

> temporarily make q0/q1/q2 authoritative during generated-power `multiply -> Chebyshev double -> recurrence correction -> one rounded Rescale`, then contract back to authoritative q0/q1 only after the post-Rescale value is again centered-unique in Q01.

This task exists because the immediately preceding native-q01 experiment proved that q01-only multiply-first wraps **before Rescale**, even though the reduced q01 centered representative looks small afterward.

This is a Primary-only design-feasibility task. Do not modify Secondary production code.

---

# Corrected interpretation of the previous task

Accepted Primary result:

`4caed4ca2cf9de7a45a2f526ec2e3f7301c9e7eb`

Previous classification:

`logn13_generated_power_native_q01_rescale_mismatch`

Observed first blocker:

`real.T2.post_rescale`

Important correction:

- native q01 arithmetic and the full-RNS mirror had identical q0/q1 rows through `real.T2.recurrence_corrected`;
- native Fast q01 Rescale matched the exact symmetric rounded quotient of the **q01 centered representative**;
- the full-RNS mirror quotient differed by ~`1.152358554734027335e18` at the largest discrepancy;
- therefore Fast q01 Rescale itself is not presently implicated.

The root cause is an earlier **transient capacity wrap hidden by q01 reduction**.

For T2:

- q01 centered reconstruction at raw product looked small:
  - `max_abs ~= 1.8341235208318604e26`
  - apparent q01 ratio ~= `0.0092599588`;
- but the q012/full-RNS-consistent reconstruction at the same raw-product stage was:
  - `max_abs ~= 3.2451868871293437e32`;
- current Q01/2 is approximately:
  - `1.9807037677037493e28`;
- therefore the intended physical raw-product magnitude is about `1.6384e4 * (Q01/2)`.

At T2 recurrence-corrected:

- q012/full-RNS-consistent `max_abs ~= 1.32857895840749e36`;
- this is about `6.7e7 * (Q01/2)`.

Therefore the previous q01-only `ratio < 1` table was measuring an already-wrapped representative and must **not** be used as transient-capacity proof.

This is exactly the standing centered-CRT rule:

> a later centered representative can look safe after wrap; transient capacity must be checked against the intended physical integer before reduction.

---

# Accepted system evidence from the earlier multiply-first task

Accepted full-RNS diagnostic multiply-first result:

Primary commit:

`904f0460e36b69add5491f7c7006c72153970789`

System result:

- D0 oracle control public-like: `0.005554220603853743`;
- old balanced generated powers G_ALL: `0.22247498019233428`;
- full-RNS multiply-first all-generated candidate: `0.005554220603853743` PASS;
- T4 local residual improved roughly `5.73e-8 -> 2.77e-8`;
- T6 local residual improved roughly `6.29e-8 -> 4.17e-8`.

Accepted q012 capacity evidence from that design task:

- q0/q1/q2 profile = 56/39/40;
- Q012 total bit length ~135 bits;
- all examined multiply-first checkpoints were centered-unique in Q012;
- maximum Q012 ratio was approximately `0.0002439022`;
- therefore temporary q2 had very large safety margin under the accepted deterministic LogN13 workload.

The prior successful candidate, however, used full-RNS diagnostic arithmetic and only used q012 as evidence. This task must now prove a genuine **three-authoritative-limb** implementation model.

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required ancestor:

`4caed4ca2cf9de7a45a2f526ec2e3f7301c9e7eb`

Start from clean `main`, fetch + ff-only pull, then re-read fresh:

- `AGENTS.md`
- `CURRENT_TASK.md`
- this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Branch: `fast-ckks`, clean.

**No Secondary production changes are authorized in this task.**

Primary-only diagnostic/design helpers are allowed.

---

# Scope lock

LogN13 only.

Every system replay must keep the accepted D0 stack fixed:

- q0/q1/q2 = 56/39/40;
- common plan scale = `2^92`;
- same polynomial and PS split;
- same B4 final-parent T2 one-bit scalar guard;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DoubleAngle local-q2 rounds;
- same S2C;
- same finalization;
- same widened Standard reference;
- same deterministic workload/input.

Do not:

- change q parameters;
- widen q0/q1;
- change coefficients;
- change PS/G0/DA/S2C design;
- change plan scale;
- production-integrate;
- modify Secondary;
- run LogN16, benchmark, Gate4/5, EXP003.

---

# R0 — reproduce controls and explicitly correct the q01 capacity interpretation

Reproduce:

1. D0 oracle public-like `0.005554220603853743`;
2. old balanced G_ALL `0.22247498019233428`;
3. full-RNS all-generated multiply-first `0.005554220603853743`;
4. native-q01 first blocker at `real.T2.post_rescale`;
5. native q01 exact rounded quotient passes;
6. T2 pre-Rescale native/full q0/q1 rows match;
7. q012/full-RNS T2 raw-product and recurrence-corrected magnitudes from the prior task.

Create an explicit `capacity_interpretation_correction` section containing:

- `q01_reduced_representative_ratio`;
- `intended_physical_ratio_against_q01`;
- `q012_ratio`;
- `q01_wrap_occurred_before_rescale`.

For T2 raw product compute:

`intended_physical_ratio_against_q01 = max_abs_from_q012_or_full_rns / (Q01/2)`.

Expected: much greater than 1 (roughly `1.6e4`).

For T2 recurrence-corrected expected: much greater than 1 (roughly `6.7e7`).

If the correction cannot be reproduced, stop:

`logn13_generated_power_q012_precondition_mismatch`.

---

# R1 — exact three-limb representation contract

Define a Primary-only diagnostic representation in which **only q0/q1/q2 are authoritative**.

Requirements:

- q0, q1, q2 rows must represent the same centered integer polynomial coefficient while Q012 uniqueness holds;
- higher limbs q3+ are ignored/dormant and must never be read as arithmetic truth;
- NTT/Montgomery state must be explicit and preserved/restored correctly;
- support every ciphertext component needed by the zero-secret Fast generated-power path;
- no decoded-slot reconstruction is allowed as the arithmetic mechanism;
- big.Int is allowed for independent diagnostic CRT/rounding proof because Q012 is ~135 bits and does not fit in 128 bits.

Provide helpers for:

1. q01 -> exact centered integer recovery when Q01 unique;
2. exact lift to q2;
3. q012 -> exact centered integer recovery;
4. coefficient-domain q012 symmetric rounded division by an arbitrary positive divisor;
5. reconstruction of quotient residues to q0/q1/q2;
6. contraction from q012 to q01 only after proving post-operation Q01 uniqueness.

---

# R2 — native q012 multiply-first generated-power operation

Implement the candidate in Primary diagnostic code without using full-RNS arithmetic as the producer.

For each required generated power:

- T2=(1,1)
- T3=(1,2)
- T4=(2,2)
- T6=(3,3)
- T8=(4,4)
- T16=(8,8)

perform:

1. take q01-authoritative parent inputs;
2. prove parent Q01 centered uniqueness;
3. recover parent centered integer coefficients from q01;
4. lift parent values to q2 exactly;
5. convert q0/q1/q2 to the correct NTT/Montgomery representation;
6. execute the generated-power multiply or square using **only q0/q1/q2-authoritative arithmetic**;
7. apply Chebyshev doubling in q012;
8. apply recurrence correction before Rescale:
   - subtract constant 1 at the correct scale when c=0;
   - otherwise perform source-equivalent scale alignment and subtract the required difference power;
9. keep q0/q1/q2 authoritative throughout;
10. perform exactly one rounded Rescale by the actual current-level divisor `Q[L]`;
11. restore output Level/Scale/degree/NTT/Montgomery metadata expected by the generated-power planner;
12. prove post-Rescale Q01 centered uniqueness;
13. contract back to authoritative q0/q1;
14. discard q2 authority before returning the completed generated power.

Important:

- q2 is a temporary representation limb, **not** the Rescale divisor;
- do not call Standard full-RNS arithmetic on stale q3+ limbs;
- do not materialize a full-RNS result and then merely copy q0/q1/q2;
- full-RNS is oracle only.

---

# R3 — intended physical capacity proof before every reduction-sensitive step

For every generated power and branch, record two distinct quantities:

## A. Reduced representative

The centered CRT representative seen from q01 after modular arithmetic.

## B. Intended physical integer

The q012/full-RNS-consistent centered integer before any contraction to q01.

At minimum record for:

- parents;
- raw product;
- doubled state;
- recurrence-aligned state;
- recurrence-corrected state;
- post-Rescale state.

For each checkpoint compute:

- `max_abs_intended`;
- `Q01/2`;
- `Q012/2`;
- intended/Q01 ratio;
- intended/Q012 ratio;
- q01 outside count based on intended value, **not reduced q01 representative**;
- q012 outside count;
- minimum bits required for intended centered uniqueness.

Rules:

- Q01 wrap is allowed while q2 is active;
- Q012 must remain centered-unique at every step where intended integer identity matters;
- after Rescale, Q01 must become centered-unique before q2 is dropped.

If Q012 fails at any required pre-Rescale checkpoint, stop:

`logn13_generated_power_native_q012_capacity_blocker`.

If post-Rescale Q01 is not unique, stop:

`logn13_generated_power_native_q012_contraction_blocker`.

---

# R4 — row-by-row native-q012 vs full-RNS oracle equivalence

Use the previously accepted full-RNS multiply-first schedule only as an independent oracle.

For T2/T3/T4/T6/T8/T16, compare native q012 against full-RNS reduced to q0/q1/q2 at matching checkpoints:

1. parent left/right;
2. raw product;
3. doubled state;
4. recurrence-aligned state;
5. recurrence-corrected state;
6. post-Rescale output before q2 contraction;
7. final contracted q01 output.

Record separately for q0, q1, q2:

- exact row equality after representation normalization;
- first mismatching component/limb/index;
- metadata equality;
- decoded semantic difference where meaningful.

Expected contract:

- q012 rows exactly match full-RNS oracle rows for limbs 0..2 at every checkpoint;
- post-Rescale q01 rows match the full-RNS oracle q0/q1 exactly.

If first divergence occurs before Rescale despite Q012 uniqueness, classify:

`logn13_generated_power_native_q012_operation_mismatch`.

---

# R5 — rounded Rescale proof in q012

For each generated power:

1. independently recover exact centered Q012 integer immediately before Rescale;
2. divisor is the actual `Q[L]` from the level being consumed;
3. compute symmetric nearest rounded signed quotient with big.Int;
4. reduce expected quotient to q0/q1/q2;
5. compare exact rows against native q012 Rescale output;
6. compare native q012 output against full-RNS Rescale oracle limbs 0..2;
7. prove post-Rescale Q01 uniqueness;
8. contract to q01 and prove exact q0/q1 row equality against oracle.

Record:

- divisor value/bit length;
- pre-Q012 uniqueness;
- expected quotient max abs;
- q0/q1/q2 quotient row equality;
- full-RNS q0/q1/q2 equality;
- post-Q01 ratio;
- contraction validity.

If q012 rounded division disagrees with the independent quotient:

`logn13_generated_power_native_q012_rescale_mismatch`.

---

# R6 — all-generated system replay

Generate the complete real power map using only the new native-q012 temporary representation for the multiply-first generated-power nodes.

No oracle power replacement except the normal T1/input basis.

Replay the fixed D0 downstream exactly:

- PS polynomial;
- accepted B4 T2 scalar guard;
- G0 local-q2=2;
- F0 local-q2=3;
- all accepted DA local-q2 rounds;
- S2C;
- finalization.

Record:

- completed T2/T3/T4/T6/T8/T16 local residuals;
- polynomial residual;
- EvalMod real/imag;
- post-S2C;
- public-like;
- public margin to `1e-2`;
- difference to the accepted full-RNS multiply-first system result;
- all capacity/contracts.

System success:

`public_like <= 1e-2`.

Preferred equivalence:

`abs(native_q012_public_like - 0.005554220603853743) <= 1e-10`.

If system passes but differs beyond the equivalence tolerance, localize the first q012/full-RNS row divergence before recommending production integration.

---

# R7 — architecture evidence

Add a compact section for the independent future q0/q1 parameter-design study.

Record:

- current q0 bits = 56;
- current q1 bits = 39;
- q2 bits = 40;
- Q01 bit length ~95;
- Q012 bit length ~135;
- maximum intended physical bits required before Rescale across all generated powers;
- maximum intended/Q01 ratio;
- maximum intended/Q012 ratio;
- maximum post-Rescale Q01 ratio;
- whether two limbs would need to be widened to avoid temporary q2 for this exact schedule.

Interpret carefully:

- current 56/39 q01 is **not** capacity-safe for multiply-first transient values, even when its reduced centered residue looks small;
- temporary q2 is justified if R3/R4/R5 pass;
- do not claim 56/39 is generally wrong for all Fast arithmetic;
- do not redesign q parameters in this task.

---

# Classification

Choose exactly one primary classification:

- `logn13_generated_power_native_q012_multiply_first_system_sufficient`
- `logn13_generated_power_native_q012_capacity_blocker`
- `logn13_generated_power_native_q012_contraction_blocker`
- `logn13_generated_power_native_q012_operation_mismatch`
- `logn13_generated_power_native_q012_rescale_mismatch`
- `logn13_generated_power_native_q012_numeric_insufficient`
- `logn13_generated_power_q012_precondition_mismatch`

If system sufficient, recommended next target:

`production integration of temporary-q2 generated-power multiply-first schedule in Secondary`.

Do not production-integrate in this task.

---

# Required compact artifact

Create:

`results/FIX-001-P3-DESIGN-LOGN13-GENERATED-POWER-NATIVE-Q012-MULTIPLY-FIRST-FEASIBILITY-summary.json`

Include:

- provenance;
- R0 controls;
- explicit q01 capacity-interpretation correction;
- dependency graph;
- intended-vs-reduced capacity table;
- q012/full-RNS row-equivalence table;
- q012 rounded-Rescale proof table;
- post-Rescale q01 contraction table;
- local power residual table;
- all-generated system metrics;
- architecture evidence;
- classification;
- first remaining blocker;
- recommended next target;
- validation flags.

Do not serialize full slots, full coefficient vectors, or complete RNS rows.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*Generated.*Power.*Native.*Q012.*Multiply.*First`;
- `go test ./...`;
- `git diff --check`;
- artifact has no NaN/Inf;
- D0 exact control reproduced;
- old balanced G_ALL reproduced;
- full-RNS multiply-first control reproduced;
- native-q01 T2 mismatch reproduced and correctly reinterpreted as pre-Rescale wrap evidence;
- intended physical capacity is never inferred solely from reduced q01 representative;
- Q012 uniqueness checked before every rounded divide/contraction-sensitive step;
- post-Rescale Q01 uniqueness proven before q2 drop;
- q012 arithmetic candidate is genuinely three-authoritative-limb, not full-RNS output materialization;
- Secondary remains exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean, unmodified;
- no q parameter changes;
- no production integration;
- no LogN16, benchmark, Gate4/5, EXP003;
- final Primary worktree clean after ordinary ff push.
