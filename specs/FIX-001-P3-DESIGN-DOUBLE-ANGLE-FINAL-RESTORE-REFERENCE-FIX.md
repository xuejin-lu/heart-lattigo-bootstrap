# FIX-001-P3-DESIGN-DOUBLE-ANGLE-FINAL-RESTORE-REFERENCE-FIX

## Purpose

The corrected normalized DoubleAngle run completed all three LogN13 rounds successfully:

- Fast and normalized full-RNS reference matched exactly at every aligned input / square / A / constant / post-Rescale checkpoint;
- all three square-capacity bounds passed;
- all three pre-Rescale exact capacity checks passed with outside count 0;
- semantic errors remained far below `1e-2`.

The run then classified:

`normalized_double_angle_final_restoration_failure`

because after multiplying the normalized Fast state by `K3`, the harness required its q0/q1 rows to match `rCoherent`, the separately evaluated ordinary DoubleAngle recurrence.

Independent review finds that exact-row requirement invalid. The normalized and ordinary paths perform different integer representatives and different Rescale rounding histories. They are required to be semantically equivalent, not bitwise identical.

The correct exact-row oracle for the final `K3` restore is the stage-aligned normalized full-RNS reference `rNorm`, which had matched Fast exactly through the end of round 2. The correct reference operation is therefore:

`rNormRestored = K3 * rNorm`

with Scale metadata unchanged.

`rCoherent` and `rExactTarget` remain semantic oracles only.

This task corrects only the final-restore oracle and reruns the same LogN13 normalized design. No Secondary production changes.

---

## Fixed provenance

Primary base when authored:

`7ee7195378414ccba3e92e8ee7ff37eaea9dce05`

Secondary exact:

`61607bb4bb82591009ce768d9a3773bed1497565`

Threshold:

`1e-2`

Expected `K3`:

`536870912 = 2^29`

Previous classification disposition:

`superseded_by_wrong_final_restore_reference`

---

# Scope lock

Do not:

- change the normalized recurrence;
- change `K_i`, `A_i`, constants, Scale schedule, target Scale, threshold, or Mod1 parameters;
- change Secondary production code;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- continue into later bootstrap stages;
- require normalized-path rows to equal ordinary/coherent-path rows after independent Rescale histories.

---

# F0 — reproduce validated three-round prefix

Rerun the same three-round normalized recurrence using the corrected stage-aligned reference harness from the prior task.

Require for rounds 0, 1, 2:

- input rows match normalized full-RNS reference;
- square rows match;
- after-A rows match;
- after-constant rows match;
- post-Rescale rows match;
- square bound passes;
- pre-Rescale exact capacity outside count = 0;
- all semantic checks <= `1e-2`.

If any previously validated checkpoint regresses, stop with:

`normalized_double_angle_final_reference_precondition_mismatch`.

At the end of round 2 require:

`fastBeforeRestore q0/q1 rows == rNorm q0/q1 rows`.

This is the exact stage-aligned starting point for final restoration.

---

# F1 — prospective K3 capacity

Convert the end-of-round-2 normalized state to the corrected q0/q1 coefficient-domain centered representation.

Let:

`B = max |c0 coefficient|`.

Because final restoration is a linear integer multiplication by `K3`, compute exactly:

`B_restore_bound = K3 * B`.

Require:

`K3 * B < Q01/2`.

Record:

- `B`;
- `K3`;
- `K3*B`;
- `Q01/2`;
- ratio `(K3*B)/(Q01/2)`;
- outside/capacity disposition.

If this fails, classify:

`normalized_double_angle_final_restore_capacity_failure`.

Do not infer safety from post-wrap residues alone.

---

# F2 — correct exact-row final restore oracle

From independent copies of the aligned end-of-round-2 states:

## Fast

Execute:

`FastCKKS.MulIntegerMaintained(fastBeforeRestore, K3, fastRestored)`

Scale metadata must remain unchanged.

## Normalized full-RNS reference

Execute the exact same integer multiplication in full RNS:

`rNormRestored = normalizedFullIntegerMultiply(rNorm, K3)`

Scale metadata must remain unchanged.

Require exact q0/q1 row equality:

`fastRestored == rNormRestored` on c0/c1 q0/q1 rows.

Also require:

- Level unchanged;
- Degree unchanged;
- NTT/Montgomery flags unchanged;
- Scale unchanged;
- c1 remains zero under the zero-secret diagnostic contract.

If rows differ, classify:

`normalized_double_angle_final_restore_primitive_mismatch`.

This is the only valid exact-row test for `K3` restoration.

---

# F3 — ordinary/coherent path is semantic-only

Do **not** require `rNormRestored` or `fastRestored` rows to equal `rCoherent` or `rExactTarget`.

Instead decode at the common pre-reset Scale and record semantic differences:

1. Fast restored vs normalized full-RNS restored;
2. normalized full-RNS restored vs ordinary coherent recurrence `rCoherent`;
3. ordinary coherent recurrence vs exact-target recurrence `rExactTarget`;
4. Fast restored vs `rCoherent`;
5. Fast restored vs `rExactTarget`.

Require all semantically relevant max-component errors <= `1e-2`.

The previously observed Fast-vs-coherent difference around `1.4e-5` is acceptable historical evidence, not a required exact value.

If exact normalized rows match but semantic equivalence to the ordinary/exact-target path fails, classify:

`normalized_double_angle_final_restore_semantic_failure`.

---

# F4 — explain the old row mismatch

Record explicitly:

`OLD_RCOHERENT_ROW_EQUALITY_DISPOSITION = invalid_due_to_independent_rescale_rounding_histories`

For evidence, record compactly that:

- normalized Fast path and `rNorm` were exact row matches through all three rounds;
- `rCoherent` was generated by the ordinary recurrence with separate Rescale operations;
- semantic equivalence remains within threshold despite row inequality.

Do not attempt to force the two independently rounded paths into bitwise equality.

---

# F5 — source-defined final metadata reset

After F2/F3 pass, capture pre-reset rows and Scale `S_DA`.

Set:

`fastRestored.Scale = original Mod1 input Scale`.

Perform the equivalent metadata reset on the normalized reference used for semantic comparison.

Require:

- Fast q0/q1 rows unchanged across reset;
- normalized-reference q0/q1 rows unchanged across reset;
- Level/Degree unchanged;
- final Scale equals original Mod1 input Scale exactly.

Because this is metadata-only, expected decoded values transform by:

`S_DA / S_input`.

Verify final decoded semantics against the independently transformed expected plaintext with max-component error <= `1e-2`.

Do not require row equality against the independently rounded `rCoherent` path.

---

# Required classification

Choose exactly one:

- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_recurrence_validated_after_final_reference_fix`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_final_reference_precondition_mismatch`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_final_restore_capacity_failure`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_final_restore_primitive_mismatch`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_final_restore_semantic_failure`
- `FIRST_SUPPORTED_CAUSE = normalized_double_angle_final_scale_reset_failure`

Also record:

- `PREVIOUS_CLASSIFICATION_DISPOSITION = superseded_by_wrong_final_restore_reference`;
- `OLD_RCOHERENT_ROW_EQUALITY_DISPOSITION = invalid_due_to_independent_rescale_rounding_histories`;
- first failing checkpoint or `none`.

A validated result means the LogN13 normalized DoubleAngle design is ready for a separate production-integration task. It does **not** authorize production changes in this task.

---

# Artifacts

Create compact artifacts only:

- `results/FIX-001-P3-DESIGN-DOUBLE-ANGLE-FINAL-RESTORE-REFERENCE-FIX-logN13.json`
- `results/FIX-001-P3-DESIGN-DOUBLE-ANGLE-FINAL-RESTORE-REFERENCE-FIX-logN13-summary.json`

No full coefficient arrays, slot vectors, trace dumps, or large mismatch lists.

---

# Validation

Before completion:

- focused final-restore-reference tests pass;
- Primary `go test ./...` passes;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no Secondary production changes;
- artifacts committed and pushed;
- both worktrees clean;
- no LogN16, benchmark, Gate 4/5, EXP-003, or later bootstrap stages.