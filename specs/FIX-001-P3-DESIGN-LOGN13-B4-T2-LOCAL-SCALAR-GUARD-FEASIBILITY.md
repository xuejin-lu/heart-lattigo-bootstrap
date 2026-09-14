# FIX-001-P3-DESIGN-LOGN13-B4-T2-LOCAL-SCALAR-GUARD-FEASIBILITY

## Purpose

Test the smallest evidence-backed arithmetic design that can reduce the LogN13 B4 parent-branch scalar-quantization residual without changing the polynomial, PS split, DoubleAngle, S2C, q parameters, or generated powers.

Accepted evidence at Primary commit `490d3f3e6402f52b8c31d62a51b3e4e862954fc3`:

- final classification: `logn13_b4_single_scalar_operation_blocker`;
- B4 completed output is exactly the G0 parent branch;
- actual public-like error: `0.010683237260415292`;
- canonical G0-parent replacement gives public-like `0.0032351181652570046`;
- only `0.07421875` of the G0-parent residual direction must be removed to pass `1e-2`;
- B4 final parent residual: `1.800410931451779e-10`;
- current Fast B4 and same-schedule full-RNS B4 decode identically (`F-N = 0`);
- genuine Standard B4 at the same native schedule has the same final residual vs canonical;
- therefore the residual is not a Fast q0/q1 implementation mismatch and not a Standard-vs-Fast semantic mismatch;
- B5 scalar-rounding ledger closes the B4 residual to about `1e-15`;
- dominant B4 scalar coefficient quantization errors are approximately:
  - T6: `-7.566812359365216e-11`;
  - T4: `-3.744174804752284e-12`;
  - T2: `-1.084943476202816e-10`;
  - constant: negligible at this scale;
- correcting only the T2 coefficient contribution at the existing scalar grid does **not** remove the accumulated B4 residual; the successful B4-term-2 checkpoint reset replaces the whole post-operation accumulator state;
- current B4 final q01 centered-capacity ratio is about `0.262698128`, so a one-bit physical guard is plausibly feasible while a two-bit guard is near/over the Q01 uniqueness boundary.

The design under test is a **one-bit local fixed-point guard around the final B4 T2 scalar accumulation**:

1. preserve the represented accumulator value while physically/metadata scaling it by 2;
2. execute the existing T2 `MulThenAdd` at the resulting doubled scalar encoding scale;
3. perform a centered rounded divide-by-2 contraction back to the original B4 Level/Scale;
4. continue the existing PS / G0 / DA / S2C / finalization path unchanged.

This is a design-feasibility task. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`490d3f3e6402f52b8c31d62a51b3e4e862954fc3`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Keep fixed:

- oracle powers;
- q0/q1/q2 = 56/39/40;
- common plan scale `2^92`;
- polynomial coefficients and PS decomposition;
- all B4 contributions except arithmetic representation of the final T2 accumulation;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DA local-q2 behavior;
- current G0 merge semantics;
- S2C and final restore behavior;
- genuine Standard widened-profile reference.

Do not:

- change any coefficient value mathematically;
- hardcode `rounded_integer + 1` or another coefficient-specific bias as a production candidate;
- change global plan scale;
- change q parameters;
- redesign generated powers;
- alter G0/DA/S2C arithmetic;
- modify Secondary production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary-only diagnostic guarded arithmetic helpers are allowed.

---

# G0 — reproduce the accepted control

Reproduce:

- completed B4 actual-vs-canonical residual `~1.800410931451779e-10`;
- actual EvalMod real/imag;
- actual post-S2C;
- actual public-like `~0.010683237260415292`;
- CA / canonical-parent public-like `~0.0032351181652570046`;
- B4 final q01 capacity ratio `~0.262698128`;
- all accepted final metadata/capacity/row contracts.

If materially different, stop:

`logn13_b4_t2_scalar_guard_precondition_mismatch`.

---

# G1 — capture the native T2 operation

At the B4 state immediately after T4 and immediately before T2 `MulThenAdd`, record:

- accumulator Level/Scale/degree/NTT/Montgomery state;
- T2 power Level/Scale;
- exact high-precision T2 coefficient;
- current scalar encoding scale;
- current rounded scalar integer real/imag;
- current effective represented coefficient;
- current coefficient quantization error;
- accumulator q01 centered-capacity ratio;
- current post-T2 q01 centered-capacity ratio.

The diagnostic must source these values from the actual executed path, not duplicate constants from the previous artifact.

---

# G2 — define the one-bit guarded T2 operation

Let the native pre-T2 accumulator be `A`, with scale `S_A`.

Construct diagnostic guarded accumulator `A_g` as follows:

1. multiply every authoritative q0/q1 maintained coefficient of `A` by 2 using the current NTT/Montgomery-compatible integer multiplication path;
2. set `A_g.Scale = 2 * S_A`;
3. preserve Level, degree, dimensions, NTT and Montgomery metadata.

This step must preserve the represented value of `A`.

Then execute the unchanged T2 scalar `MulThenAdd` using:

- the same T2 power;
- the same mathematical T2 coefficient;
- the current Fast scalar conversion/rounding code;
- `A_g` as the accumulator.

Because the accumulator scale is doubled, the actual scalar encoding scale should increase by approximately 2x. Record the exact scale and rounded integer actually used; do not force an expected integer.

Call the guarded post-add state `Y_g`.

---

# G3 — guard capacity and value-preservation contract

Before any contraction, establish intended centered physical values, not merely reduced representatives.

Require at each guarded checkpoint:

- after accumulator x2 promotion;
- after guarded T2 scalar add;

that every authoritative q01 coefficient has a unique intended centered representative:

`|x| < Q01/2`.

Record:

- max absolute intended centered coefficient;
- `max_abs / (Q01/2)`;
- outside count;
- q0/q1 row agreement with a stage-aligned full-RNS/integer oracle for the same guarded arithmetic.

If the one-bit guarded T2 output is not Q01-unique/value-preserving, stop:

`logn13_b4_t2_one_bit_scalar_guard_capacity_blocker`.

Do not hide a transient wrap behind a safe reduced representative.

Also compute the hypothetical two-bit (`x4`) capacity ratio from the exact intended state. This is **capacity evidence only**; do not execute the x4 candidate unless it is strictly Q01-unique. No q2 extension is authorized in this task.

---

# G4 — centered rounded divide-by-2 contraction

Contract `Y_g` back to the native B4 T2 output representation at the **same Level** and original scale `S_A`.

Required semantics per coefficient:

1. recover the exact centered integer represented by q0/q1 using CRT;
2. divide by 2 with deterministic nearest-integer rounding symmetric for positive/negative values;
3. reconstruct q0/q1 residues of the rounded quotient;
4. restore the original NTT/Montgomery representation;
5. set output Scale back to `S_A`;
6. preserve Level, degree and dimensions.

Do not use modular inverse-of-2 as a substitute for rounded integer division unless every coefficient is proven exactly divisible by 2.

Record:

- count/fraction of even vs odd centered coefficients before contraction;
- maximum contraction rounding error in integer units;
- exact q0/q1 reconstruction match against an independent big.Int oracle;
- output q01 uniqueness/capacity;
- Level/Scale/degree/NTT/Montgomery equality with the native B4 T2 output metadata.

If reconstruction or rounded division differs from the independent oracle, stop:

`logn13_b4_t2_scalar_guard_contraction_mismatch`.

---

# G5 — local precision result

Compare the guarded-and-contracted B4 output `B4_G1` with the deterministic canonical B4 checkpoint.

Record real/imag:

- max-component error;
- max-abs complex error;
- mean error;
- worst slot/component;
- improvement factor vs native B4 residual;
- direction similarity of `(B4_G1 - B4_Q)` vs native `(B4_F - B4_Q)`.

Also record the guarded T2 effective coefficient quantization error after its doubled encoding scale.

The goal is not `<=1e-10` per se. The system threshold decides sufficiency.

---

# G6 — full downstream replay

Feed `B4_G1` into the unchanged remainder of the accepted pipeline:

- remaining PS / G0 merge;
- accepted DA local-q2 path;
- validated S2C path;
- unchanged final restore.

Record:

- final PS residual vs canonical;
- EvalMod real/imag vs genuine Standard;
- post-S2C max-component vs Standard;
- public-like max-component vs message;
- final Level/Scale/degree;
- all capacity/row/rounded-division/contraction flags.

Primary pass condition:

`public_like <= 1e-2`.

Also record public margin:

`1e-2 - public_like`.

If G1 passes with valid arithmetic contracts, do **not** broaden the arithmetic design in this task.

---

# G7 — conditional second candidate only if G1 is valid but insufficient

Run this section only if:

- G1 is capacity-safe;
- G4 contraction is exact according to its oracle;
- all downstream contracts pass;
- but public-like remains `>1e-2`.

Then test one additional evidence-backed candidate:

### G2 candidate: one-bit guards on both B4 T6 and B4 T2

Apply the exact same local guard / rounded-divide-by-2 procedure independently around:

1. T6 scalar accumulation;
2. T2 scalar accumulation;

with T4 left unchanged.

Each guarded operation must contract back to its native Level/Scale before the next native B4 operation.

This candidate is justified because T6 and T2 are the two dominant scalar ledger errors. Do not add guards to constant/T4.

Record the same local and downstream metrics.

If G1 already passes, skip this section and mark `not_needed_minimal_candidate_passed`.

---

# G8 — decision classification

Choose exactly one primary classification.

### One-bit T2 guard is system-sufficient

If G1 passes all arithmetic contracts and public-like <= `1e-2`:

`logn13_b4_t2_one_bit_scalar_guard_system_sufficient`

Recommended next step: production integration of this exact local arithmetic mechanism only, followed by full public-like validation.

### T2-only insufficient, T6+T2 sufficient

If G1 is valid but fails public threshold and conditional G2 passes:

`logn13_b4_t6_t2_one_bit_scalar_guards_system_sufficient`

### Guard arithmetic is valid but still insufficient

If all tested evidence-backed guard candidates are valid but public-like remains > `1e-2`:

`logn13_b4_local_scalar_guard_insufficient_precision`

### One-bit T2 capacity blocker

`logn13_b4_t2_one_bit_scalar_guard_capacity_blocker`

### Rounded contraction mismatch

`logn13_b4_t2_scalar_guard_contraction_mismatch`

No production modification is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-B4-T2-LOCAL-SCALAR-GUARD-FEASIBILITY-summary.json`

Include only:

- provenance;
- G0 controls;
- native T2 operation metadata/rounding;
- one-bit guarded scale/rounded scalar integer;
- guarded capacity table;
- x4 capacity-only evidence;
- contraction evidence;
- B4 local precision result;
- G1 downstream result;
- conditional G2 result or explicit skip reason;
- classification;
- first remaining blocker;
- recommended next target;
- validation flags.

Do not serialize slot vectors, coefficient arrays, complete RNS rows, matrices or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*B4.*T2.*Scalar.*Guard` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no q/G0/DA/S2C/generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.