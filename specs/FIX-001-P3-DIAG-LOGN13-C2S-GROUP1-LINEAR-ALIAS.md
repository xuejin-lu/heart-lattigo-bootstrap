# FIX-001-P3-DIAG-LOGN13-C2S-GROUP1-LINEAR-ALIAS

## Purpose

Diagnose the first downstream C2S budget breach that remains after the accepted group-0 compressed design.

Accepted predecessor:

- Primary result commit: `302efbaef8ab15e0e7b09ee40911921753495ee0`
- Secondary exact clean commit: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- accepted classification: `logn13_c2s_group0_compressed_linear_design_validated`
- `k_min_full_group = 3`
- `k_safe_full_group = 4`
- group-0 `k=4` restored Fast vs original genuine Standard max-component error: approximately `9.78e-13`
- derived downstream-sensitive C2S budget: approximately `1.953125e-5`
- first later boundary over budget: `group1_post_rescale`
- Fast EvalMod matched-input divergence remains unresolved and is out of scope here.

This task is **diagnostic only**. Do not implement a production fix.

---

## Fixed provenance and scope

Primary required base:

`302efbaef8ab15e0e7b09ee40911921753495ee0`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- LogN13 only
- deterministic `reproducibleInput.v1`
- C2S matrix schedule `[1,1,1,1]`
- diagnose **C2S group 1 / matrix index 1 only**
- group 0 must first use the already validated diagnostic design:
  - re-encode group-0 mathematical diagonals at original matrix Scale / `2^4`
  - run group-0 LinearTransform
  - ordinary one-level Rescale
  - physical + metadata restore by `2^4`
- groups 2–3 are not to be internally diagnosed in this task.

Do not:

- modify Secondary;
- modify production DFT matrix generation;
- modify Fast LinearTransform, Fast Rescale, EvalMod, packing, finalization, or bootstrap scheduling;
- change BSGS ratios, matrix factorization, LevelQ/LevelP, workload, or C2S mathematical diagonals;
- implement a compressed group-1 design yet;
- run LogN16, benchmark, Gate 4/5, or EXP-003.

---

# G1-0 — reproduce accepted preconditions

Before localizing group 1, reproduce all of the following:

1. genuine Standard evaluator proof is present and the Fast compatibility path is absent;
2. Standard end-to-end LogN13 error <= `1e-2`;
3. ModUp Fast-vs-Standard semantic delta <= `1e-9`;
4. validated group-0 `k=4` compressed+Rescale+restore path succeeds;
5. group-0 restored Fast vs original genuine Standard max-component error <= derived C2S budget;
6. group-0 restored metadata matches the original production group-0 post-Rescale contract;
7. when the existing unmodified group-1 production operation is applied next, `group1_post_rescale` exceeds the derived C2S budget.

If these do not reproduce:

`logn13_c2s_group1_precondition_mismatch`.

---

# G1-1 — construct aligned group-1 inputs

The group-1 diagnostic must avoid input asymmetry.

Construct two aligned inputs at the group-1 boundary:

### Fast input

Use the validated group-0 `k=4` Fast compressed+Rescale+physical+metadata-restore result.

### Full-RNS reference input

Use the corresponding genuine full-RNS compressed Standard group-0 result with the same ordinary Rescale and physical+metadata restore.

Mandatory checks before group 1:

- same logical Level;
- same Scale within normal metadata tolerance;
- same LogDimensions;
- same degree;
- Fast q0/q1 rows exactly equal the full-RNS reference reduced modulo q0/q1 after representation normalization;
- semantic difference <= derived C2S budget;
- full-RNS group-1 input centered-q0/q1 capacity is measured and recorded.

Also keep the **original uncompressed genuine Standard group-1 input/output** as a semantic oracle for later functional comparison. Do not use its q0/q1 projection after capacity is lost.

If aligned input construction fails:

`logn13_c2s_group1_input_alignment_mismatch`.

---

# G1-2 — record actual group-1 transform structure

From production C2S matrix index 1, record compactly:

- matrix index = 1;
- `N1`;
- direct vs BSGS path;
- sorted diagonal indices;
- number of diagonals;
- if BSGS: giant groups, baby rotations, diagonals per giant group;
- matrix Scale and log2 Scale;
- LevelQ;
- LevelP;
- LogDimensions;
- aligned input Level and Scale.

Replay group 1 in the exact source-defined Fast LinearTransform order.

---

# G1-3 — genuine full-RNS exact replay and aligned Fast modular replay

Replay the **entire group-1 LinearTransform**, not merely until the first unsafe checkpoint.

Use:

1. genuine full-RNS Standard replay from the aligned full-RNS group-1 input;
2. aligned Fast q0/q1 replay from the aligned Fast group-1 input.

For every source-defined checkpoint record compactly:

- checkpoint order;
- checkpoint name/type;
- giant index if applicable;
- baby rotation if applicable;
- diagonal if applicable;
- Standard full-RNS exact coefficient capacity versus `Q01/2`;
- Fast q0/q1 centered representative capacity;
- whether Standard full-RNS centered q0/q1 reconstruction is unique;
- whether Fast residues exactly equal Standard full-RNS reduced modulo q0/q1.

Required checkpoints:

- baby rotations;
- each individual diagonal product;
- each BSGS inner partial sum;
- each giant rotation;
- each outer partial sum;
- completed group-1 LinearTransform.

No coefficient arrays or slot arrays in the artifact.

---

# G1-4 — oracle rule after capacity loss

If the genuine full-RNS exact value at a checkpoint satisfies:

`max |x| >= Q01/2`,

then:

- set `FULL_RNS_CENTERED_Q01_UNIQUE=false`;
- do **not** decode centered q0/q1 projection as genuine Standard semantics;
- continue checking whether Fast residues equal genuine full-RNS values reduced modulo q0/q1.

If capacity is lost but residues still match, classify the Fast primitive at that checkpoint as modularly correct; the failure is representational alias, not arithmetic corruption.

---

# G1-5 — semantic comparisons while safe

At checkpoints where the genuine full-RNS value is still centered-q0/q1 unique:

- compare aligned Fast semantic values against genuine full-RNS aligned-reference semantics;
- require q0/q1 rows to match exactly after representation normalization;
- record max-component/max-complex/mean-complex error.

Possible safe-region failures:

- rotation mismatch;
- diagonal product mismatch;
- inner accumulation mismatch;
- giant rotation mismatch;
- outer accumulation mismatch.

Do not attribute any mismatch to capacity if the full-RNS exact value is still within `Q01/2`.

---

# G1-6 — ordinary group-1 Rescale

After the completed group-1 LinearTransform, apply the same ordinary one-level Rescale used by production.

Record:

- pre/post Level;
- pre/post Scale;
- dropped modulus;
- Standard full-RNS pre-Rescale capacity;
- Fast q0/q1 pre-Rescale capacity;
- post-Rescale capacity for both;
- Fast residues vs aligned Standard reduced modulo q0/q1;
- Fast vs aligned Standard semantic error after Rescale when the oracle is valid.

Separately compare the aligned full-RNS group-1 post-Rescale result against the **original uncompressed genuine Standard group-1 post-Rescale output**. This measures only the tiny inherited precision effect of the validated group-0 compression and must remain below the derived C2S budget.

---

# G1-7 — authoritative earliest classification

The first supported cause is the earliest source-order checkpoint satisfying either:

1. full-RNS value is still centered-q0/q1 unique but Fast disagrees; or
2. full-RNS value first loses centered-q0/q1 uniqueness while Fast remains exactly modulo-correct.

Use exactly one primary classification from:

- `logn13_c2s_group1_precondition_mismatch`
- `logn13_c2s_group1_input_alignment_mismatch`
- `logn13_c2s_group1_rotation_mismatch`
- `logn13_c2s_group1_diagonal_product_mismatch`
- `logn13_c2s_group1_single_term_q01_alias`
- `logn13_c2s_group1_inner_accumulation_mismatch`
- `logn13_c2s_group1_inner_accumulation_q01_alias`
- `logn13_c2s_group1_giant_rotation_mismatch`
- `logn13_c2s_group1_outer_accumulation_mismatch`
- `logn13_c2s_group1_outer_accumulation_q01_alias`
- `logn13_c2s_group1_linear_modularly_correct_but_final_q01_alias`
- `logn13_c2s_group1_rescale_mismatch`
- `logn13_c2s_group1_linear_no_internal_divergence`.

Record the exact first failing checkpoint.

---

# G1-8 — full-group compression requirement if alias is confirmed

If any representational alias is confirmed, do **not** derive a design only from the first unsafe checkpoint.

Because the complete full-RNS group-1 replay is required, calculate:

`R_max = max over all group-1 LinearTransform checkpoints of (maxExact / (Q01/2))`.

Then derive:

- `k_min_full_group = max(0, ceil(log2(R_max)))`;
- `k_safe_full_group = k_min_full_group + 1`.

Record:

- first unsafe checkpoint and its ratio;
- worst-capacity checkpoint and `R_max`;
- whether first alias occurs at a single term, inner accumulation, giant rotation, outer accumulation, or only the completed transform;
- whether compression would be required before plaintext multiplication or only before/within accumulation.

These values are **diagnostic recommendations only**. Do not implement group-1 compression in this task.

---

# G1-9 — preserve downstream separation

Record explicitly:

- accepted group-0 design remains validated;
- `C2S_ERROR_ALONE_SUFFICIENT_FOR_FINAL_FAILURE = true`;
- `FAST_EVALMOD_MATCHED_INPUT_DIVERGENCE_REMAINS_UNRESOLVED = true`;
- no EvalMod work was performed.

Do not inspect group 2 or group 3 internals.

---

# Artifact

Create one compact artifact:

`results/FIX-001-P3-DIAG-LOGN13-C2S-GROUP1-LINEAR-ALIAS-summary.json`

Include only:

- provenance;
- precondition reproduction;
- aligned group-1 input summary;
- actual group-1 transform structure;
- compact ordered checkpoint summaries;
- first failing checkpoint;
- first supported cause;
- worst full-group capacity ratio;
- `k_min_full_group` / `k_safe_full_group` if alias confirmed;
- group-1 pre/post-Rescale summary;
- aligned-reference vs original-Standard semantic comparison;
- validation booleans.

Do not serialize full vectors, coefficient arrays, encoded matrices, or repetitive hashes.

---

# Validation

Before completion require:

- focused `TestFIX001P3.*C2S.*Group1.*Linear` tests pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- Primary diagnostic support and compact artifact committed;
- do not push without explicit push authorization if the repository safety policy requires it;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no production fix.