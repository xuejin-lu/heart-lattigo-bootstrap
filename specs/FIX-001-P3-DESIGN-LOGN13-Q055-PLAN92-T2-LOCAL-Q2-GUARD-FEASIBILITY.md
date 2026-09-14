# FIX-001-P3-DESIGN-LOGN13-Q055-PLAN92-T2-LOCAL-Q2-GUARD-FEASIBILITY

## Goal

Determine whether the apparent `q0=56` production precondition is actually caused by the **final-parent B4 T2 one-bit scalar guard remaining q0/q1-only** under `planScale=2^92`.

Primary hypothesis:

> `q0=55 + planScale=2^92` fails because the guarded scalar accumulation transient exceeds centered Q01 capacity before the divide-by-two contraction. If that single guard is promoted to temporary q0/q1/q2 authority, the fixed checked-in frontend configuration (`q0=55`) should recover the accepted D0 system result without changing the frontend config.

This is a Primary-only design-feasibility/attribution task.

**Do not modify Secondary production code in this task.**

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required ancestor:

`ce335898c233eb3b08984206dc5772a0681fba12`

Start from clean `main`, fetch + ff-only pull, then re-read fresh:

- `AGENTS.md`
- `CURRENT_TASK.md`
- this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Required branch: `fast-ckks`

Required state: clean.

Secondary is read-only in this task.

---

# Correct interpretation of the previous closure matrix

The previous matrix produced:

- A: q0=55, plan91 -> FAIL, public-like `0.01076789220080715`;
- B: q0=55, plan92 -> FAIL catastrophically, public-like `3662.9000547146406`;
- C: q0=56, plan91 -> FAIL, public-like `0.01076789220080715`;
- D: q0=56, plan92 -> PASS, public-like `0.005554220603853743`.

Therefore the previous primary classification `logn13_d0_q0_56_width_precondition` is **not logically sufficient**. Only D passes, so the observed matrix establishes a q0/planScale interaction or joint precondition until the interaction is localized.

Two harness defects are accepted and must be corrected here:

1. classification precedence bug:
   - the previous switch tested `(C || D) && !A && !B` before the D-only interaction case;
   - therefore D-only success was incorrectly captured as `q0_56_width_precondition`;
2. first-divergence bug:
   - any downstream system failure was hard-coded to `public_finalization`;
   - B actually already fails at EvalMod (`~102.299`) and post-S2C (`~114.466`).

Do not reuse either faulty classification rule.

---

# Source-backed reason for testing the T2 scalar guard

Current Secondary production one-bit scalar guard:

`schemes/ckks/fast/scalar_guard.go`

Semantics:

1. multiply accumulator by 2;
2. double accumulator Scale;
3. execute ordinary q0/q1 `MulThenAdd`;
4. centered rounded divide by 2 using **q0/q1 CRT only**;
5. restore native metadata/scale.

Therefore this guard is capacity-sensitive to Q01 **before** the contraction.

Historical accepted T2-guard feasibility used the q0=56 diagnostic profile and planScale=2^92. The final-parent T2 scalar operation must now be re-evaluated under the real q0=55 profile.

Expected qualitative scaling:

- reducing q0 by roughly one bit approximately halves Q01;
- the same physical guarded transient therefore approximately doubles its Q01 capacity ratio;
- a q0=56 transient around one-half of Q01/2 would become approximately capacity-bound or capacity-invalid under q0=55.

This task must prove or reject that with exact intended physical coefficients. Do not rely on wrapped q01 representatives.

---

# Fixed system stack

Use the checked-in LogN13 frontend configuration unchanged:

`configs/bootstrap_config.logN13.json`

Required:

`q0 = 55`

Use:

`planScale = 2^92`

Keep the complete accepted D0 arithmetic stack fixed:

- accepted C2S compression;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DoubleAngle local-q2 rounds;
- generated-power temporary-q012 multiply-first for T2/T3/T4/T6/T8/T16;
- same polynomial coefficients/PS split;
- same S2C/finalization;
- same deterministic workload/reference.

The **only candidate change** is the final-parent B4 T2 scalar guard implementation.

Do not change q parameters, plan scale, coefficients, generated-power schedule, G0/F0/DA guards, S2C, finalization, or frontend config.

---

# T0 — reproduce the q0/plan matrix classification correctly

Reproduce or import with source-backed checks the four previous pass/fail outcomes:

- A=false
- B=false
- C=false
- D=true

Classification regression requirement:

- this pattern must be interpreted as `q0_plan_interaction`, not `q0_width_only`.

First-divergence regression requirement:

- B must be reported as failing by EvalMod or earlier, not `public_finalization`;
- A/C may remain late precision/public-threshold failures if their upstream stages are otherwise healthy.

Add focused assertions so the old switch-order bug and hard-coded first-divergence bug cannot recur.

---

# T1 — B control: q0=55, plan92, current q01 T2 guard

Reproduce case B using the complete fixed stack and the existing q01-only one-bit scalar guard.

Required control metrics:

- generated-power local residuals remain near canonical;
- generated-power q012 proofs remain valid;
- EvalMod real/imag reproduce the catastrophic semantic mismatch order of magnitude (~102);
- post-S2C reproduces ~114;
- public-like reproduces ~3662.9.

If B does not reproduce, stop:

`logn13_q055_plan92_t2_guard_precondition_mismatch`.

---

# T2 — exact final-parent T2 guard capacity attribution

Instrument the real and imag final-parent `B4-term-2` scalar operation under **B: q0=55, plan92**.

At minimum capture these states:

1. native accumulator before guard promotion;
2. promoted accumulator after multiply-by-2;
3. scalar product / encoded contribution before accumulation;
4. intended physical post-add value before divide-by-two;
5. wrapped/reduced q01 representative of that post-add value;
6. exact centered divide-by-two quotient;
7. contracted final state.

For each relevant coefficient-domain state compute independently:

- max intended `|x|`;
- Q01/2;
- Q012/2;
- intended/Q01 ratio;
- intended/Q012 ratio;
- q01 outside_count based on intended value;
- q012 outside_count;
- minimum required centered-uniqueness bits;
- reduced q01 representative ratio separately.

### Hard attribution rule

If intended physical post-add `|x| >= Q01/2` while Q012 remains unique, classify the current q01 guard as a transient-capacity blocker.

Do not infer safety from the wrapped q01 centered representative.

Also run the same instrumentation for D (`q0=56, plan92`) as a control and compare ratios.

Expected qualitative relationship:

`B_q01_ratio ~= 2 * D_q01_ratio`

within the actual generated modulus ratio, not by a hard-coded factor 2.

---

# T3 — genuine temporary-q2 T2 scalar guard candidate

Implement a Primary-only diagnostic candidate for the same `B4-term-2` operation with **q0/q1/q2 authoritative only for the guard boundary**.

Required semantics:

1. start from the actual q01-authoritative accumulator and T2 power;
2. prove both are Q01 centered-unique at guard entry;
3. recover exact centered integer coefficients from q01;
4. lift exactly to q2;
5. keep q0/q1/q2 authoritative in correct NTT/Montgomery form;
6. multiply accumulator by 2 in q012;
7. update Scale exactly as the one-bit guard requires;
8. execute the same scalar `MulThenAdd` semantics in q012, including source-equivalent scalar encoding/rounding;
9. require Q012 centered uniqueness for the intended physical post-add state;
10. perform centered symmetric rounded divide by 2 in q012;
11. prove the quotient is centered-unique in Q01;
12. contract to q0/q1;
13. restore exactly the same native Level/Scale/degree/NTT/Montgomery metadata expected by the existing guard.

Do not:

- use full-RNS arithmetic as the producer;
- read q3+ dormant rows;
- inject decoded/canonical outputs;
- change the scalar coefficient;
- change guard selection;
- change planScale.

Full-RNS/big.Int may be used only as independent oracle evidence.

---

# T4 — row/rounding proof for local-q2 guard

For real and imag compare native-q012 candidate against an independent full-RNS/big.Int oracle at:

- promoted accumulator;
- scalar contribution;
- post-add;
- post-divide-by-two;
- final q01 contraction.

Record q0/q1/q2 exact row equality and metadata equality.

The divide-by-two oracle must use symmetric nearest integer rounding with the same tie/sign convention as the accepted guard.

Require:

- Q012 unique before contraction-sensitive operations;
- exact rounded quotient rows;
- post-divide Q01 unique;
- exact q0/q1 contraction rows.

If the q012 guard arithmetic itself diverges:

`logn13_q055_plan92_t2_local_q2_guard_operation_mismatch`.

---

# T5 — B system replay with only the T2 guard replaced

Replay the complete q0=55 / plan92 D0 stack, changing only:

`final-parent T2 q01 one-bit scalar guard`

into:

`final-parent T2 temporary-q2 one-bit scalar guard`.

Record:

- guarded B4 residual real/imag;
- polynomial residual if available;
- EvalMod real/imag;
- post-S2C;
- public-like;
- public margin to `1e-2`;
- final metadata/contracts.

System pass condition:

`public_like <= 1e-2`.

Preferred target:

- result should be in the same numeric regime as accepted D (`0.005554220603853743`), though exact equality is not required unless all profile-dependent effects cancel.

If B passes, this proves q0=56 is **not** a required frontend parameter change for correctness; the actual precondition is that plan92 must not use a q01-only T2 guard whose transient exceeds Q01.

---

# T6 — D control with q012 guard

Run D (`q0=56, plan92`) with the same temporary-q2 T2 guard candidate.

Because D's q01 guard is expected to be capacity-safe, require the q012 guard to be semantically equivalent to the accepted q01 guard after contraction:

- exact or source-explained q0/q1 row agreement;
- same metadata;
- same system result within existing diagnostic tolerance.

This prevents the local-q2 guard from becoming a profile-specific semantic hack.

---

# T7 — production-readiness conclusion

Choose the conclusion strictly from evidence.

## If B passes with only local-q2 T2 guard

Set:

`production_ready_fixed_q055_plan92_with_t2_local_q2`

The next production task may keep the frontend config unchanged and integrate the full D0 stack under q0=55, including:

- internal normalized LogN13 planScale 91 -> 92;
- generated-power temporary-q2 multiply-first;
- G0/F0/DA local-q2 guards;
- final-parent T2 temporary-q2 one-bit scalar guard;
- accepted C2S compression;
- no q0 frontend change.

## If B remains failing

Identify the next earliest genuine divergence after the corrected T2 guard and stop. Do not infer q0=56 is required unless all q0=55 alternatives supported by the already-accepted local-q2 architecture are exhausted with source-backed evidence.

---

# Classification

Choose exactly one:

- `logn13_q055_plan92_t2_q01_guard_capacity_blocker`
- `logn13_q055_plan92_t2_local_q2_guard_system_sufficient`
- `logn13_q055_plan92_t2_local_q2_guard_operation_mismatch`
- `logn13_q055_plan92_t2_local_q2_guard_numeric_insufficient`
- `logn13_q0_plan_interaction_not_t2_guard`
- `logn13_q055_plan92_t2_guard_precondition_mismatch`

If the q01 capacity blocker is confirmed and the q012 replacement also makes the system pass, use the **system-sufficient** classification as primary and record the q01 capacity blocker as the causal finding.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-Q055-PLAN92-T2-LOCAL-Q2-GUARD-FEASIBILITY-summary.json`

Include:

- provenance;
- corrected previous-matrix interpretation;
- classification-regression evidence;
- B q01-guard control;
- D q01-guard control;
- real/imag guard capacity tables;
- reduced-vs-intended q01 distinction;
- native-q012 guard row proof;
- divide-by-two oracle proof;
- post-q01 contraction proof;
- B full-system replay after q012 guard;
- D equivalence replay after q012 guard;
- production readiness;
- classification;
- first remaining blocker;
- recommended next target;
- validation flags.

Do not serialize full slot arrays, full coefficient vectors, or full RNS rows.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*Q055.*Plan92.*T2.*Local.*Q2.*Guard`;
- `go test ./...` in Primary;
- `git diff --check`;
- artifact contains no NaN/Inf;
- checked-in LogN13 config remains unchanged at q0=55;
- Secondary remains exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean, unmodified;
- previous matrix D-only pass is classified as interaction, not q0-width-only;
- B first divergence is not hard-coded to public finalization;
- B current q01 guard failure is reproduced before candidate replacement;
- intended transient capacity is computed from physical q012/full-RNS-consistent values, not wrapped q01 representatives;
- q012 local guard is genuine q0/q1/q2-authoritative arithmetic;
- post-divide Q01 uniqueness proven before q2 drop;
- no q parameter change;
- no production integration;
- no LogN16;
- no benchmark;
- no Gate4/5;
- no EXP003;
- final Primary worktree clean after ordinary ff push.

---

# Recommended next target rules

If B passes:

`production integration of the complete LogN13 D0 stack under fixed q0=55 with internal planScale92 and temporary-q2 T2 guard`

If B still fails after a correct q012 T2 guard:

`localize the first remaining q0=55/plan92 divergence before any frontend parameter change`
