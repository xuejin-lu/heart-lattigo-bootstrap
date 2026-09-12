# FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-REFERENCE-FIX

## Purpose

The previous normalized DoubleAngle design diagnostic ended with:

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_arithmetic_mismatch`

at LogN13 round 0.

That classification is **superseded pending a corrected rerun** because independent review found two diagnostic-harness defects:

1. the Fast ciphertext was mutated through `square -> multiply A -> subtract constant` before row comparisons were made, so the reported `first_square_row_mismatch` did not compare the Fast square checkpoint against the full-RNS square checkpoint;
2. the full-RNS reference square copied input metadata but did not update `product.Scale = input.Scale * input.Scale`, so subsequent reference constant encoding used the wrong Scale.

The prior run nevertheless showed that Fast semantic checkpoints remained extremely accurate through square / multiplier / constant, so the normalized recurrence itself has not been disproved.

This task fixes only the reference harness and reruns the exact same LogN13 normalized-recurrence design experiment through all three DoubleAngle rounds if each checkpoint passes.

No Secondary production changes.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`9fc323dd0a3aa97622b52a7a34a5f9eaf8cb6dc7`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Authoritative polynomial oracle:

`raw_chebyshev_on_preprocessed_z`

Threshold:

`1e-2`

LogN13 DoubleAngle count:

`3`

Do not modify Secondary production code.

---

# Scope lock

Do not:

- change the normalized recurrence mathematics;
- change the working Scale selection;
- change `K_i`, `A_i`, constants, target Scale, Mod1 parameters, or threshold;
- modify Secondary `MulRelin`, `Add`, `Rescale`, key generation, or `mod1/fast.go`;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- continue into CoeffsToSlots or later bootstrap stages;
- treat the previous `normalized_double_angle_arithmetic_mismatch` as authoritative.

The only goal is to determine whether the previously proposed normalized recurrence is valid once the full-RNS oracle and stage alignment are corrected.

---

# H0 — reproduce the two harness defects in code review evidence

Before rerunning, record compact source-backed evidence that the previous harness did both of the following:

### H0.1 stale checkpoint object

Previous code mutated one `fastNormalized` through:

1. `MulRelin`;
2. `MulIntegerMaintained`;
3. `Add(-constant)`;

and only afterward compared that same object to:

- `fullSquare`;
- `fullAfterA`;
- `fullAfterConstant`.

Therefore the prior `first_square_row_mismatch` was not a square-vs-square comparison.

### H0.2 missing squared Scale in full-RNS reference

Previous `normalizedFullSquareStep` copied input metadata into `product` but failed to set:

`product.Scale = input.Scale.Mul(input.Scale)`.

Because `normalizedFullSubtractConstant` converts the plaintext constant into an integer using `input.Scale`, every reference constant checkpoint after the square used the wrong scale encoding.

Record:

`PREVIOUS_CLASSIFICATION_DISPOSITION = superseded_by_reference_harness_error`

---

# H1 — fix full-RNS square metadata

In the diagnostic reference helper only:

After exact full-RNS coefficient multiplication, set the product metadata exactly as the corresponding Fast multiplication does.

At minimum require:

- Degree = 2 for the raw product;
- Level = input Level;
- `IsNTT` unchanged;
- `IsMontgomery` unchanged;
- `Scale = input.Scale * input.Scale`;
- LogDimensions preserved consistently with the input.

When truncating the zero-secret degree-2 product to the degree-1 reference, preserve the **squared Scale**.

Add a focused unit test that fails if the reference square retains the input Scale.

---

# H2 — immutable Fast stage snapshots

For every DoubleAngle round, take independent snapshots immediately after each actual Fast operation:

- `fastInput` — before square;
- `fastSquare` — immediately after `MulRelin`;
- `fastAfterA` — immediately after `MulIntegerMaintained(A_i)`;
- `fastAfterConstant` — immediately after `Add(-c_i/K_{i+1})`;
- `fastPostRescale` — immediately after `Rescale`.

Never use a later-mutated ciphertext to represent an earlier checkpoint.

The actual working ciphertext may remain in-place, but checkpoint evidence must use immutable copies.

Add a focused test proving that modifying the live ciphertext after `fastSquare` does not change the stored square snapshot.

---

# H3 — exact stage-aligned full-RNS reference

For each round construct:

- `refInput` corresponding exactly to `fastInput`;
- `refSquare` from `refInput * refInput`, with squared Scale metadata;
- `refAfterA = A_i * refSquare`, Scale unchanged from `refSquare`;
- `refAfterConstant = refAfterA - c_i/K_{i+1}`, encoded using the **current squared Scale**;
- `refPostRescale` using the same dropped logical modulus schedule.

For each corresponding Fast/reference pair require exact q0/q1 row equality:

1. input;
2. square;
3. after A;
4. after constant / pre-Rescale;
5. post-Rescale.

If a row mismatch occurs, stop at the **first truly aligned stage** and record:

- round;
- checkpoint;
- component / limb / index;
- Fast value;
- reference value;
- Fast Scale;
- reference Scale;
- representation state.

Use classifications:

- `normalized_double_angle_square_arithmetic_mismatch`
- `normalized_double_angle_multiplier_arithmetic_mismatch`
- `normalized_double_angle_constant_arithmetic_mismatch`
- `normalized_double_angle_rescale_arithmetic_mismatch`

Do not use the generic old classification when a stage can be identified.

---

# H4 — semantic oracle remains independent

Keep the high-precision normalized recurrence oracle independent from ciphertext/reference rows.

For each round preserve semantic checks for:

- input `z_i`;
- square `z_i^2`;
- multiplied value `A_i z_i^2`;
- constant-adjusted `z_{i+1}` before Rescale;
- post-Rescale `z_{i+1}`.

Threshold remains `1e-2`.

The previous round-0 values near `1e-15` or smaller are useful historical evidence but must be recomputed.

If exact rows match but semantic output fails, classify:

`normalized_double_angle_semantic_failure`.

---

# H5 — capacity checks

Keep the same capacity discipline:

### Before square

Use the authoritative q0/q1 coefficient-domain view and conservative negacyclic bound:

`N * B^2 < Q01/2`.

If false:

`normalized_double_angle_square_capacity_failure`.

### Before Fast Rescale

Use the corrected full-RNS exact coefficient reference after A and constant.

Require every exact c0 coefficient used by the Fast q0/q1 reconstruction to satisfy:

`|x| < Q01/2`.

If false:

`normalized_double_angle_pre_rescale_capacity_failure`.

Do not infer capacity from post-wrap q0/q1 centered representatives alone.

---

# H6 — run all three rounds when valid

If round 0 passes aligned rows, semantic checks, and capacity, continue immediately through rounds 1 and 2 using the exact same normalized recurrence design.

For every round record compactly:

- Level in/out;
- Scale in / squared / post-Rescale;
- `K_i`, `K_{i+1}`;
- `A_i`;
- normalized constant;
- semantic max component error at each checkpoint;
- square bound ratio;
- pre-Rescale exact capacity ratio/outside count;
- exact row-match booleans at each aligned checkpoint.

Do not serialize coefficient arrays or slot vectors.

---

# H7 — final restoration and Mod1 scale reset

Only if all three normalized DoubleAngle rounds pass:

1. execute the same final normalized-value restoration specified by the previous design task;
2. verify it against the coherent full-RNS reference;
3. verify semantic equivalence to the independently iterated ordinary DoubleAngle recurrence;
4. then perform the source-defined final metadata-only reset to the original Mod1 input Scale;
5. verify q0/q1 rows do not change across the metadata-only reset;
6. verify final decoded semantics against the expected reset-scale plaintext.

This task still stops before later bootstrap stages.

---

# Required final classification

Choose exactly one:

- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_recurrence_validated_after_reference_fix`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_square_capacity_failure`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_pre_rescale_capacity_failure`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_square_arithmetic_mismatch`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_multiplier_arithmetic_mismatch`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_constant_arithmetic_mismatch`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_rescale_arithmetic_mismatch`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_semantic_failure`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_final_restoration_failure`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_reference_fix_precondition_mismatch`

Also record:

`PREVIOUS_CLASSIFICATION_DISPOSITION = superseded_by_reference_harness_error`

and the first failing round/checkpoint or `none`.

---

# Artifacts

Create compact artifacts:

- `results/FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-REFERENCE-FIX-logN13.json`
- `results/FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-REFERENCE-FIX-logN13-summary.json`

Summary must remain human-reviewable.

Store only:

- provenance;
- H0 harness-error disposition;
- corrected scale bookkeeping evidence;
- compact per-round aligned checkpoints;
- capacity ratios/outside counts;
- final restoration/reset evidence if reached;
- classification;
- tests and clean-state evidence.

Do not store full coefficient vectors, slot vectors, operation traces, or large mismatch lists.

---

# Validation

Before completion:

- focused normalized-reference-fix tests pass;
- Primary `go test ./...` passes;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no Secondary production changes;
- artifacts committed and pushed normally;
- Primary and Secondary worktrees clean;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no later bootstrap stages.

The deliverable is whether the normalized DoubleAngle recurrence survives a correctly staged and correctly scaled full-RNS reference comparison.