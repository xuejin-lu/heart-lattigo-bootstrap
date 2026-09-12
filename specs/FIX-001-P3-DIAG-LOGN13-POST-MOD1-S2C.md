# FIX-001-P3-DIAG-LOGN13-POST-MOD1-S2C

## Purpose

Production LogN13 normalized Fast Mod1 is now integrated and validated.

Accepted production state:

- Primary: `7745583290cd9d8c8febd1bc8ec1adb0982b1951`
- Secondary `fast-ckks`: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- classification: `logn13_normalized_fast_mod1_production_integrated`
- production Mod1 output: Level 4, Degree 1, NTT/Montgomery, Scale restored to the original Mod1 input Scale
- production output q0/q1 rows match the accepted normalized oracle and a freshly generated normalized full-RNS oracle exactly
- semantic threshold `1e-2` passes

The actual Fast bootstrap core is source-ordered as:

`ScaleDown -> ModUp -> CoeffsToSlots -> EvalMod -> SlotsToCoeffs`

The next unvalidated production boundary is therefore the real `DFTEvaluator.SlotsToCoeffsNew` call after production EvalMod.

This task resumes the real LogN13 bootstrap only through **SlotsToCoeffs / bootstrapCore output**. It is diagnostic/validation only. Do not modify Secondary production code unless the task first proves a concrete Secondary defect and the current spec explicitly authorizes a fix; this spec does not authorize such a fix.

---

# Fixed provenance

Primary required base:

`7745583290cd9d8c8febd1bc8ec1adb0982b1951`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Branch:

`fast-ckks`

Configuration remains the accepted reproducible LogN13 configuration:

- LogN = 13
- LogSlots = 12
- Residual MaxLevel = 1
- CircuitOrder = `ModUpThenEncode`
- CoeffsToSlots factorization = four groups
- SlotsToCoeffs factorization = three groups (`[1,1,1]`)
- Mod1 degree = 30
- DoubleAngle = 3
- K = 16
- LogMessageRatio = 10
- threshold = `1e-2`

Do not alter the workload or parameter construction.

---

# Scope lock

Do not:

- modify Secondary production code;
- modify Fast Mod1 or polynomial code;
- modify DFT implementation;
- modify Fast arithmetic primitives;
- change parameters, workload, or stage order;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- run public unpack/finalization as the system under test;
- claim full Bootstrap success from this task.

The task stops at the **bootstrapCore output immediately after production SlotsToCoeffs**.

---

# S0 — reproduce the real production prefix

Use the ordinary reproducible LogN13 frontend and actual production Fast evaluators.

Execute the same source path as `bootstrapCore` through:

1. Pack/Switch to the bootstrap ring only if needed to reproduce the exact core input used by the current harness;
2. production `ScaleDown`;
3. production `ModUp`;
4. production `DFTEvaluator.CoeffsToSlotsNew`;
5. production `EvalMod` on `ctReal` and, when present, on `ctImag`.

Do not substitute the old diagnostic recurrence for production EvalMod.

Require the production EvalMod output(s) to reproduce the already accepted production Mod1 evidence before continuing.

For the dense full-slot LogN13 profile, explicitly record whether `ctImag` is present after CoeffsToSlots and validate both branches before SlotsToCoeffs.

If the accepted Mod1 boundary cannot be reproduced, stop with:

`post_mod1_s2c_precondition_mismatch`.

---

# S1 — production SlotsToCoeffs is the system under test

Call exactly:

`eval.DFTEvaluator.SlotsToCoeffsNew(ctReal, ctImag, eval.S2CDFTMatrix)`

on independent copies of the validated production EvalMod output(s).

Record compactly:

- input Level / Degree / Scale / LogDimensions;
- `ctImag` present or nil;
- S2C matrix `LevelQ`;
- `S2CDFTMatrix.Levels`;
- number of factor matrices;
- output Level / Degree / Scale / LogDimensions;
- NTT/Montgomery flags;
- q0/q1 row hashes;
- input immutability.

For the accepted profile, source structure predicts three group Rescales from Level 4 to Level 1. Require final Level = 1.

Do not hard-code a final Scale value from memory. Derive the expected scale from the matrix plan / full-RNS reference and require equality with that source-backed value.

If the production public call returns an error, classify with the first source-backed failing operation when possible; otherwise classify:

`post_mod1_s2c_production_call_failure`.

---

# S2 — fresh full-RNS reference input from q0/q1

Do not use dormant higher Fast limbs as reference data.

For each validated EvalMod input (`ctReal`, and `ctImag` if present):

1. take only authoritative q0/q1;
2. convert correctly from NTT/Montgomery to coefficient-domain canonical residues;
3. reconstruct the centered integer coefficient from q0/q1;
4. prove/record the input coefficient capacity needed for unique centered reconstruction;
5. freshly reduce the reconstructed integer coefficients into every active full-RNS modulus;
6. rebuild an independent full-RNS NTT ciphertext with matching metadata and c1 semantics.

Never copy limbs >=2 from the Fast ciphertext into the reference.

If q0/q1 centered uniqueness cannot be justified at this boundary, classify:

`post_mod1_s2c_input_capacity_unresolved`.

---

# S3 — Standard/full-RNS SlotsToCoeffs reference

Build an independent Standard DFT evaluator for the same `S2CDFTMatrix` using the matrix's required Galois elements.

Evaluate the freshly lifted full-RNS reference with Standard `SlotsToCoeffsNew`.

The reference must use:

- the same matrix generated by the bootstrap parameters;
- the same LevelQ / Levels / Scaling / LogSlots / format;
- a freshly constructed full-RNS input from S2;
- no stale Fast higher limbs.

Normalize representation only for comparison (for example IMForm the Fast result when comparing against non-Montgomery Standard rows); do not alter mathematical Scale or Level.

At final S2C output require:

- Level equality;
- Scale equality;
- LogDimensions equality;
- Degree equality;
- exact q0/q1 row equality after representation normalization.

If final rows match exactly, record:

`S2C_FINAL_ROWS_MATCH = true`.

---

# S4 — group-by-group localization and capacity

Even if the public production S2C succeeds, independently replay the same factor/group sequence for diagnostic evidence so that a mismatch can be localized.

Source contract:

- every matrix factor performs a linear transform;
- each `S2CDFTMatrix.Levels` group ends with exactly one Rescale;
- accepted profile has three groups.

For each group record only compact evidence:

- group index;
- input Level / Scale;
- matrix indices in the group;
- Fast pre-Rescale row hashes;
- full-RNS reference pre-Rescale row hashes;
- exact-row match before Rescale when meaningful;
- full-RNS exact coefficient max magnitude before Rescale;
- `Q01/2`;
- exact capacity ratio / outside count;
- Fast post-Rescale Level / Scale / row hashes;
- reference post-Rescale Level / Scale / row hashes;
- exact q0/q1 match after Rescale.

Capacity oracle must use full-RNS exact coefficient reconstruction before the Fast Rescale. Do not infer no-alias from a post-wrap q0/q1 centered representative.

If a group has exact coefficients outside `Q01/2`, classify:

`post_mod1_s2c_group_capacity_failure`

and stop at the first such group.

If capacity passes but Fast/reference first differ before a group Rescale, classify:

`post_mod1_s2c_linear_transform_mismatch`.

If pre-Rescale rows/capacity pass but post-Rescale rows differ, classify:

`post_mod1_s2c_rescale_mismatch`.

---

# S5 — semantic check

Decode/project the final production S2C output using the existing q0/q1-authoritative zero-secret diagnostic decode path.

Compare against the independently generated Standard/full-RNS S2C reference semantics.

Require max-component error <= `1e-2`.

Exact q0/q1 row equality is the stronger preferred check when capacity is established; semantic agreement is still required as a separate sanity check.

If rows are exact but semantic comparison fails, classify:

`post_mod1_s2c_semantic_failure`.

---

# S6 — core-output boundary

On success, the output should be the same state that `bootstrapCore` returns immediately before `BootstrapMany` performs `UnpackAndSwitchN2ToN1`.

Record:

- Level = 1;
- Degree = 1;
- NTT/Montgomery state;
- Scale;
- LogDimensions;
- q0/q1 row hashes;
- exact match to the independent full-RNS S2C reference;
- semantic error;
- input immutability;
- Secondary provenance.

Do not invoke public unpack/finalization after this accepted boundary in this task.

Successful classification:

`FIRST_SUPPORTED_CAUSE = logn13_post_mod1_s2c_core_output_validated`

This success authorizes a separate next task for `UnpackAndSwitchN2ToN1` + public finalization / end-to-end LogN13 Bootstrap. It does not authorize LogN16 or benchmarking.

---

# Required classifications

Choose exactly one:

- `logn13_post_mod1_s2c_core_output_validated`
- `post_mod1_s2c_precondition_mismatch`
- `post_mod1_s2c_input_capacity_unresolved`
- `post_mod1_s2c_production_call_failure`
- `post_mod1_s2c_group_capacity_failure`
- `post_mod1_s2c_linear_transform_mismatch`
- `post_mod1_s2c_rescale_mismatch`
- `post_mod1_s2c_final_metadata_mismatch`
- `post_mod1_s2c_semantic_failure`

Also record first failing group/checkpoint or `none`.

---

# Artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-DIAG-LOGN13-POST-MOD1-S2C-summary.json`

Do not create full coefficient arrays, slot vectors, repeated rotation arrays, or operation traces.

The summary should contain only:

- provenance;
- production prefix confirmation;
- S2C matrix metadata;
- compact group checkpoints;
- capacity ratios/outside counts;
- final row hashes;
- semantic metric;
- classification;
- tests / scope confirmations.

---

# Validation

Before completion require:

- focused `TestFIX001P3...` test for this S2C task passes;
- Primary `go test ./...` passes;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- Primary artifact and supporting diagnostic code committed/pushed;
- Primary worktree clean and remote synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no unpack/finalization after the S2C core-output boundary.