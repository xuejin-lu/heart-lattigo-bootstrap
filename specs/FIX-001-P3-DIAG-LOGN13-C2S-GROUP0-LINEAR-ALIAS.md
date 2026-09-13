# FIX-001-P3-DIAG-LOGN13-C2S-GROUP0-LINEAR-ALIAS

## Purpose

The previous precision-budget diagnostic proved that the first downstream-sensitive C2S precision breach occurs inside **group 0 LinearTransform**, not in the group-0 Rescale.

Accepted predecessor:

- Primary result commit: `da13958650fbf923ffd5049f741610a8e3143c53`
- Secondary exact clean: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- classification: `logn13_c2s_precision_first_breach_group0_linear_transform`
- ModUp Fast-vs-Standard semantic delta: `0`
- derived C2S real budget: approximately `1.953125e-5`
- group 0 after LinearTransform Fast-vs-Standard max-component error: `0.04690968130449508`
- group 0 after Rescale: `0.046909681304491926`
- Rescale incremental error: approximately `-3.15e-15`
- group 0 matrix Scale: approximately `2^56`
- group 0 dropped modulus: 56 bits

Critical capacity evidence at the group-0 LinearTransform output:

- Fast q0/q1 centered representative ratio to `Q01/2`: `0.9999435648779011`, outside count 0;
- genuine Standard full-RNS exact coefficient ratio to `Q01/2`: `7.260077050804978`, outside count 4726;
- therefore the Standard full-RNS output is **not uniquely representable by centered q0/q1**, even though Fast still has a modular q0/q1 residue.

Fast source `schemes/ckks/fast/linear_transform.go` shows the transform is q0/q1-only and consists of:

- rotation / automorphism;
- plaintext diagonal multiplication in NTT/Montgomery q0/q1;
- modular q0/q1 additions for inner/outer accumulation;
- output Scale = input Scale × matrix Scale;
- no Rescale until the surrounding DFT group completes.

The next task must determine whether group-0 divergence is:

1. a primitive mismatch while exact values are still within q0/q1 capacity; or
2. modularly correct q0/q1 arithmetic whose exact full-RNS term/partial-sum grows beyond centered-CRT uniqueness, causing unavoidable alias before Rescale.

Diagnostic only. Do not modify Secondary production code.

---

## Fixed provenance and scope

Primary required base:

`da13958650fbf923ffd5049f741610a8e3143c53`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- `configs/bootstrap_config.logN13.json`
- LogN = 13
- LogSlots = 12
- C2S levels `[1,1,1,1]`
- investigate **group 0 / matrix index 0 only**
- use deterministic `reproducibleInput.v1`

Do not:

- modify Secondary;
- modify matrix generation, matrix scale, BSGS ratio, or workload;
- modify Fast LinearTransform, rotation, Rescale, or EvalMod;
- run later C2S groups except where needed to reproduce predecessor preconditions;
- run LogN16, benchmark, Gate 4/5, or EXP-003;
- implement a fix.

---

# L0 — reproduce predecessor preconditions

Reproduce:

- genuine Standard evaluator proof;
- genuine Standard E2E <= `1e-2`;
- ModUp Fast-vs-Standard semantic delta <= `1e-9`;
- group-0 LinearTransform final Fast-vs-Standard error near predecessor `0.0469096813`;
- group-0 Rescale adds negligible error relative to the LinearTransform;
- group-0 Standard full-RNS output exceeds centered q0/q1 uniqueness while Fast modular q0/q1 output exists.

If these do not reproduce, classify:

`logn13_c2s_group0_linear_precondition_mismatch`.

---

# L1 — record the actual transform structure

From the actual group-0 matrix object, record compactly:

- `matrix.N1`;
- direct path or BSGS path;
- number of diagonals in `matrix.Vec`;
- sorted diagonal indices;
- if BSGS: number of giant groups, baby-rotation set, and diagonal count per giant group;
- matrix Scale / LevelQ / LevelP / LogDimensions;
- group input Scale and Level.

Do not dump encoded diagonal coefficient arrays.

The diagnostic replay must follow the exact source-defined order, including BSGS grouping and rotations.

---

# L2 — independent full-RNS term oracle

Construct an independent **genuine full-RNS** replay of group-0 LinearTransform using the same Standard group input and the same encoded matrix diagonals.

For every source-defined diagonal contribution, materialize the full-RNS contribution before it is accumulated.

For a direct path, each contribution is:

`term_d = plaintextDiagonal_d * Rotate(input, d)`.

For a BSGS path, preserve the exact source convention:

- baby rotation by `i`;
- multiply by encoded `Vec[j+i]`;
- accumulate inner terms;
- giant rotation by `j`;
- accumulate outer terms.

Use genuine Standard/full-RNS operations or a source-faithful diagnostic equivalent. Do not derive the reference from Fast q0/q1 dormant limbs.

At every term/partial-sum checkpoint record the exact full-RNS coefficient-domain capacity against:

`Q01/2`.

Record compactly:

- max absolute exact coefficient;
- ratio to `Q01/2`;
- outside count;
- whether centered q0/q1 reconstruction would be unique.

---

# L3 — Fast modular replay aligned to the same checkpoints

Replay the same group-0 transform using the actual Fast source-defined operations and capture immutable q0/q1 checkpoints aligned with L2.

Required checkpoint classes:

### Rotations

For every distinct baby/direct rotation used by group 0:

- compare decoded Fast rotation semantics against genuine Standard rotation semantics;
- require <= derived C2S budget while the input itself is common;
- when the Standard rotated state is q0/q1-unique, also compare exact q0/q1 rows after representation normalization.

If the first safe checkpoint already fails, classify:

`logn13_c2s_group0_rotation_mismatch`.

### Single diagonal products

For every diagonal term before accumulation:

- compare Fast q0/q1 residue against the genuine full-RNS term reduced modulo q0/q1;
- decode semantics when the full-RNS term is centered-q0/q1 unique;
- record exact term capacity.

If the first term is within capacity but Fast modular residues disagree:

`logn13_c2s_group0_diagonal_product_mismatch`.

If a single exact term is the first object to exceed `Q01/2` while Fast residues still match modulo q0/q1:

`logn13_c2s_group0_single_term_q01_alias`.

### Inner partial sums (BSGS only)

After every source-order inner accumulation, record:

- full-RNS exact capacity;
- Fast-vs-full-RNS modulo-q0/q1 residue match;
- decoded semantic comparison only if centered uniqueness still holds.

If the first capacity loss occurs here and residues remain modulo-correct:

`logn13_c2s_group0_inner_accumulation_q01_alias`.

If capacity is safe but residues disagree:

`logn13_c2s_group0_inner_accumulation_mismatch`.

### Giant rotations (BSGS only)

For each completed inner sum before outer accumulation:

- apply the source-defined giant rotation on both sides;
- if the full-RNS input to that rotation is still q0/q1 unique, require semantic agreement and exact projected-row agreement;
- if uniqueness has already been lost, only test modular residue consistency and do not treat centered projection as semantic oracle.

If a safe giant rotation first fails:

`logn13_c2s_group0_giant_rotation_mismatch`.

### Outer partial sums

After every outer accumulation, record full-RNS exact capacity and Fast modulo-q0/q1 residue agreement.

If the first capacity loss occurs during outer accumulation while residues remain modulo-correct:

`logn13_c2s_group0_outer_accumulation_q01_alias`.

If capacity is safe but residues disagree:

`logn13_c2s_group0_outer_accumulation_mismatch`.

---

# L4 — distinguish modular correctness from semantic uniqueness

This distinction is mandatory.

For any checkpoint whose genuine full-RNS exact coefficient satisfies:

`max |x| >= Q01/2`,

do **not** decode a q0/q1-centered projection and call it the Standard semantic oracle.

Instead record separately:

1. `FULL_RNS_CENTERED_Q01_UNIQUE = false`;
2. whether Fast q0/q1 rows equal the genuine full-RNS result reduced modulo q0/q1.

If the exact value exceeds centered capacity and residues still match, the arithmetic primitive is modularly correct; the failure mechanism is representational alias, not primitive arithmetic error.

This rule overrides any misleading post-wrap centered value that happens to lie inside `[-Q01/2,Q01/2)`.

---

# L5 — first authoritative failure

Stop causal localization at the earliest source-order checkpoint satisfying one of:

- safe exact value + Fast modular/semantic mismatch;
- exact capacity first exceeds q0/q1 centered uniqueness while Fast remains modulo-correct.

Choose exactly one classification:

- `logn13_c2s_group0_linear_precondition_mismatch`
- `logn13_c2s_group0_rotation_mismatch`
- `logn13_c2s_group0_diagonal_product_mismatch`
- `logn13_c2s_group0_single_term_q01_alias`
- `logn13_c2s_group0_inner_accumulation_mismatch`
- `logn13_c2s_group0_inner_accumulation_q01_alias`
- `logn13_c2s_group0_giant_rotation_mismatch`
- `logn13_c2s_group0_outer_accumulation_mismatch`
- `logn13_c2s_group0_outer_accumulation_q01_alias`
- `logn13_c2s_group0_linear_modularly_correct_but_final_q01_alias`
- `logn13_c2s_group0_linear_no_internal_divergence`

For the final alias classification, require that all earlier terms/partial sums remain unique and the first loss occurs only at the completed transform output.

Record `first_failing_checkpoint` precisely, including diagonal / giant-group / partial-sum index where applicable.

---

# L6 — compression requirement for the next design task

If any alias classification is reached, derive a **diagnostic-only minimum power-of-two compression exponent** from the largest exact full-RNS coefficient observed up to the first unsafe checkpoint.

Let:

`R = maxExact / (Q01/2)`.

Compute:

`k_min = ceil(log2(R))`

so division of the physical coefficients by `2^k_min` would bring the observed maximum to <= `Q01/2` in the same checkpoint.

Also record a one-bit safety candidate:

`k_safe = k_min + 1`.

This is **not** authorization to change the matrix Scale or production implementation. It is only evidence for a later normalized/compressed LinearTransform design.

Record:

- worst checkpoint;
- worst exact coefficient ratio;
- `k_min`;
- `k_safe`;
- whether the required compression appears before multiplication or only before accumulation.

Do not implement compression in this task.

---

# L7 — preserve downstream context

Record:

- derived C2S budget from the predecessor;
- group-0 completed LinearTransform semantic error;
- group-0 Rescale incremental error;
- `C2S_ERROR_ALONE_SUFFICIENT_FOR_FINAL_FAILURE = true`;
- `FAST_EVALMOD_MATCHED_INPUT_DIVERGENCE_REMAINS_UNRESOLVED = true`.

The next task after this one must address only the earliest confirmed group-0 mechanism. Do not mix the C2S fix with EvalMod work.

---

# Artifact

Create one compact artifact only:

`results/FIX-001-P3-DIAG-LOGN13-C2S-GROUP0-LINEAR-ALIAS-summary.json`

Include:

- provenance and scope validation;
- actual group-0 transform structure;
- compact ordered checkpoint list up to the first authoritative failure;
- capacity metrics;
- modular residue-match booleans;
- semantic metrics only where the reference is valid;
- first failing checkpoint and classification;
- compression requirement if alias confirmed.

Do not serialize:

- full slot vectors;
- coefficient arrays;
- full plaintext diagonal arrays;
- repeated per-coefficient traces;
- giant hash collections.

---

# Validation

Before completion require:

- focused `TestFIX001P3.*C2S.*Group0.*Linear` tests pass;
- Primary `go test ./...` passes;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- Primary diagnostic support/artifact committed and pushed;
- Primary worktree clean and `origin/main` synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no production fix.