# FIX-001-P3-DESIGN-LOGN13-C2S-GROUP1-COMPRESSED-LINEAR

## Purpose

Validate a compressed representation for **LogN13 C2S group 1** using the same design principle already validated for group 0:

- keep the mathematical DFT diagonals unchanged;
- re-encode the group-1 matrix at a lower plaintext encoding scale;
- execute the complete group-1 LinearTransform;
- perform the ordinary one-level Rescale;
- restore physical coefficients and metadata Scale by the same exact power of two;
- compare against genuine full-RNS Standard semantics.

This is diagnostic/design validation only. Do not modify Secondary production code.

Accepted predecessor:

- Primary result commit: `5c1c0d9a5e0a599370cc152faab56d91649518f8`
- Secondary exact clean: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- group-0 accepted design: `k_safe_full_group=4`
- group-1 classification: `logn13_c2s_group1_outer_accumulation_q01_alias`
- first unsafe checkpoint: `outer_partial_3840`
- worst complete-group ratio: `1.0439473580664484`
- Fast residues still match Standard reduced modulo q0/q1 at the unsafe checkpoint
- diagnostic recommendation: `k_min_full_group=1`, `k_safe_full_group=2`
- group-1 post-Rescale semantic error: approximately `4.880298e-4`
- derived downstream-sensitive C2S budget: approximately `1.953125e-5`

---

## Fixed provenance and scope

Primary required base:

`5c1c0d9a5e0a599370cc152faab56d91649518f8`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- LogN13 only
- deterministic `reproducibleInput.v1`
- C2S schedule `[1,1,1,1]`
- group 0 must use the already validated `k=4` compressed + ordinary Rescale + physical+metadata restore design
- design target: **group 1 / matrix index 1 only**

Do not:

- modify Secondary;
- modify production DFT matrix generation;
- modify Fast LinearTransform, Fast Rescale, EvalMod, packing, or finalization;
- change mathematical diagonal values;
- change BSGS structure, rotations, LevelQ/LevelP, workload, or DFT factorization;
- implement a production fix;
- internally diagnose groups 2 or 3 in this task;
- run LogN16, benchmark, Gate 4/5, or EXP-003.

---

# D0 — reproduce accepted aligned group-1 input

Reproduce:

- genuine Standard evaluator proof;
- Standard E2E <= `1e-2`;
- ModUp Fast-vs-Standard delta <= `1e-9`;
- group-0 `k=4` compressed/restored output within derived C2S budget of original Standard group-0 post-Rescale;
- group-1 aligned input Fast-vs-full-RNS Standard semantic delta <= derived C2S budget;
- group-1 aligned input q0/q1 residues exact after representation normalization;
- group-1 original transform structure exactly matches predecessor.

If not:

`logn13_c2s_group1_compression_baseline_mismatch`.

---

# D1 — candidate set

Test exactly:

`k in {1,2,3}`

Rationale:

- predecessor complete-group worst ratio `1.043947...` implies theoretical minimum `k=1`;
- `k=2` is the one-extra-bit safety candidate;
- `k=3` provides one additional measured margin and helps quantify precision cost.

Do not assume the predecessor recommendation is valid before measurement.

For each candidate, re-encode the **same mathematical group-1 diagonal values** at:

`compressedMatrixScale = originalGroup1MatrixScale / 2^k`.

Must preserve:

- diagonal values mathematically unchanged;
- diagonal index set;
- BSGS `N1`;
- giant/baby rotation structure;
- LevelQ/LevelP;
- LogDimensions.

Re-encode from unencoded/source diagonal values using Lattigo encoding. Do not divide already-encoded residues modulo q0/q1.

---

# D2 — complete internal replay for each candidate

For each `k`, replay all source-order group-1 checkpoints with:

1. genuine full-RNS compressed Standard reference;
2. Fast q0/q1 compressed candidate.

Include:

- baby rotations;
- every diagonal product;
- every inner partial accumulation;
- giant rotations;
- every outer partial accumulation;
- completed LinearTransform.

For every checkpoint record compactly:

- exact full-RNS max coefficient;
- ratio to `Q01/2`;
- outside count;
- centered q0/q1 uniqueness;
- Fast residues vs Standard reduced mod q0/q1;
- semantic metric only while centered uniqueness is valid.

A candidate fails capacity if any checkpoint exceeds centered q0/q1 uniqueness.

If capacity is safe but residue rows disagree:

`logn13_c2s_group1_compression_linear_arithmetic_mismatch`.

---

# D3 — ordinary Rescale

For every complete-group-safe candidate:

- perform exactly one ordinary group-1 Rescale;
- consume the same production logical dropped modulus;
- do not change Rescale implementation;
- record pre/post levels, scales, dropped modulus, capacity, residue equality, and semantics.

---

# D4 — physical + metadata restore

After Rescale multiply both physical coefficients and metadata Scale by exact integer `2^k`.

Requirements:

- no extra level consumed;
- no metadata-only promotion;
- preserve domain flags and LogDimensions;
- restored Scale must match the original group-1 post-Rescale downstream contract within normal CKKS metadata tolerance.

Record capacity before/after restore.

---

# D5 — authoritative semantic comparisons

For each safe restored candidate compare:

1. compressed Standard restored vs original uncompressed genuine Standard group-1 post-Rescale;
2. Fast compressed restored vs compressed Standard restored;
3. Fast compressed restored vs original uncompressed genuine Standard group-1 post-Rescale.

Require each relevant max-component error <= derived C2S budget.

If compressed Standard itself exceeds budget:

`logn13_c2s_group1_compression_encoding_precision_insufficient`.

If compressed Standard passes but Fast fails:

`logn13_c2s_group1_compression_fast_mismatch`.

---

# D6 — choose measured full-group k

Derive only from measured candidates:

- `k_min_full_group`: smallest tested k whose every group-1 LinearTransform checkpoint is centered-q0/q1 unique and whose restored semantics pass budget;
- `k_safe_full_group = k_min_full_group + 1` only if that candidate was actually tested and passes all requirements.

Expected accepted classification:

`logn13_c2s_group1_compressed_linear_design_validated`

Otherwise:

`logn13_c2s_group1_compressed_linear_design_not_validated`.

---

# D7 — downstream boundary probe

Only after a valid `k_safe_full_group` exists:

- run group 0 with validated `k=4` design;
- run group 1 with validated new group-1 safety candidate;
- run existing unmodified production operations for groups 2 and 3;
- finish C2S real/imag split;
- compare against genuine Standard using the same derived C2S budget.

Record only:

- group-1 corrected post-Rescale semantic delta;
- final C2S real/imag deltas;
- first later group boundary over budget, if any;
- whether correcting groups 0+1 is sufficient for full C2S budget.

Do **not** internally diagnose the later group in this task.

Possible downstream dispositions:

- `groups0_1_fix_satisfies_c2s_budget`
- `groups0_1_fix_valid_but_next_budget_breach_group2`
- `groups0_1_fix_valid_but_next_budget_breach_group3`
- `groups0_1_fix_valid_but_final_real_imag_split_breach`.

---

# D8 — preserve EvalMod separation

Record explicitly:

- `C2S_ERROR_ALONE_SUFFICIENT_FOR_FINAL_FAILURE=true` from accepted evidence;
- `FAST_EVALMOD_MATCHED_INPUT_DIVERGENCE_REMAINS_UNRESOLVED=true`.

Do not execute or modify EvalMod.

---

# Artifact

Create one compact artifact:

`results/FIX-001-P3-DESIGN-LOGN13-C2S-GROUP1-COMPRESSED-LINEAR-summary.json`

Include only compact evidence:

- provenance;
- baseline reproduction;
- candidate summaries for k=1,2,3;
- per-candidate worst capacity ratio and first unsafe checkpoint;
- residue-match booleans;
- Rescale/restore metadata;
- semantic comparisons;
- measured `k_min_full_group` / `k_safe_full_group`;
- downstream probe disposition;
- final classification.

Do not serialize full slot vectors, coefficient vectors, encoded diagonals, or redundant full checkpoint traces in the summary. If detailed checkpoint evidence is needed, keep it in a separate raw artifact and keep the summary human-reviewable.

---

# Validation

Before completion require:

- focused `TestFIX001P3.*C2S.*Group1.*Compressed` tests pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- Primary committed;
- ordinary fast-forward push to `origin/main` is allowed under `AGENTS.md` standing safe-push authorization if all safety conditions hold;
- Primary worktree clean and `origin/main` synchronized after push;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no production fix.