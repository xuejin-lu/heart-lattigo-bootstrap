# FIX-001-P3-DESIGN-LOGN13-C2S-GROUP0-COMPRESSED-LINEAR

## Purpose

Validate a **diagnostic-only compressed representation** for LogN13 C2S group 0 that prevents q0/q1 alias during the LinearTransform while preserving the original C2S mathematical function and the downstream scale contract.

Accepted predecessor:

- Primary result commit: `7059d11da93d1bc79e76ce4f1fd05f1dfbe1e8d9`
- Secondary exact clean: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- classification: `logn13_c2s_group0_single_term_q01_alias`
- first unsafe checkpoint: `giant_0_diagonal_0`
- first-term full-RNS ratio to `Q01/2`: `2.4330520098231347`
- first-term Fast residues match genuine Standard reduced mod q0/q1
- completed original group-0 LinearTransform full-RNS ratio: `7.260077050804978`
- completed original group-0 Fast-vs-Standard semantic error: `0.04690968130449508`
- group-0 Rescale incremental semantic error: approximately `-3.15e-15`
- derived downstream-sensitive C2S budget: approximately `1.953125e-5`

Important distinction:

- predecessor `k_min=2`, `k_safe=3` were derived only from the **first unsafe single term**;
- for the **entire completed group-0 LinearTransform**, the observed ratio `7.260077...` implies:
  - full-group minimum theoretical compression exponent `ceil(log2(7.260077...)) = 3`;
  - one additional safety-bit candidate = `4`.

This task must test candidates `k = 2, 3, 4`. Do not assume any candidate is valid before measurement.

Diagnostic/design validation only. Do not modify Secondary production code.

---

## Fixed provenance and scope

Primary required base:

`7059d11da93d1bc79e76ce4f1fd05f1dfbe1e8d9`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- LogN13 only
- deterministic `reproducibleInput.v1`
- C2S matrix schedule `[1,1,1,1]`
- design target: **group 0 / matrix index 0 only**
- original group-0 matrix Scale approximately `2^56`
- original group input Scale approximately `2^50`

Do not:

- modify Secondary;
- modify production DFT matrix generation;
- modify Fast LinearTransform, Fast Rescale, EvalMod, packing, or finalization;
- change C2S mathematical diagonal values;
- change BSGS structure, rotations, matrix factorization, LevelQ/LevelP, or workload;
- implement a production fix;
- run LogN16, benchmark, Gate 4/5, or EXP-003.

---

# D0 — reproduce accepted baseline

Reproduce the predecessor baseline:

- genuine Standard evaluator proof;
- ModUp Fast-vs-Standard semantic delta <= `1e-9`;
- original group-0 BSGS structure unchanged;
- original completed Standard full-RNS group-0 LinearTransform capacity ratio near `7.260077...` and not q0/q1 unique;
- original Fast-vs-Standard group-0 LinearTransform error near `0.0469096813`;
- original group-0 Rescale incremental error negligible compared with LinearTransform error.

If baseline does not reproduce:

`logn13_c2s_group0_compression_baseline_mismatch`.

---

# D1 — construct compressed matrices without changing DFT mathematics

For each candidate `k in {2,3,4}` construct a **Primary-only diagnostic group-0 matrix** as follows.

Use the same group-0 unencoded mathematical diagonals as the production C2S matrix. Re-encode those same values at:

`compressedMatrixScale = originalMatrixScale / 2^k`.

Mandatory invariants:

- mathematical diagonal complex values are unchanged;
- diagonal index set unchanged;
- BSGS `N1` unchanged;
- giant/baby rotation structure unchanged;
- LevelQ, LevelP, LogDimensions unchanged;
- only the plaintext **encoding scale / physical coefficient magnitude** is reduced.

Do **not** obtain the compressed matrix by modularly dividing already-encoded plaintext residues. Re-encode from the source diagonal values using the existing Lattigo linear-transformation encoder so CKKS rounding is well-defined.

Record for each candidate:

- k;
- original and compressed matrix Scale/log2 Scale;
- same-structure booleans;
- diagonal count and BSGS structure summary.

---

# D2 — full group-0 internal capacity replay for each k

For each candidate replay the complete group-0 LinearTransform in exact source order using:

1. a genuine full-RNS compressed Standard reference;
2. the Fast q0/q1 compressed candidate.

Replay the **entire group**, not only until the predecessor first-failure checkpoint.

At all source-defined checkpoints record compactly:

- rotations;
- individual diagonal products;
- BSGS inner partial sums;
- giant rotations;
- outer partial sums;
- completed LinearTransform.

For each materialized term/partial sum record:

- Standard full-RNS `max_abs_exact_coefficient`;
- ratio to `Q01/2`;
- outside count;
- centered q0/q1 uniqueness;
- Fast q0/q1 residue match against Standard reduced modulo q0/q1.

Semantic comparison is authoritative only while full-RNS centered q0/q1 uniqueness holds.

A candidate fails capacity if **any** checkpoint exceeds centered q0/q1 uniqueness.

Expected diagnostic behavior to verify, not hard-code:

- `k=2` may make the first term safe but is expected to remain insufficient for the completed group because baseline completed ratio / 4 is > 1;
- `k=3` is the theoretical minimum from the observed completed-group ratio;
- `k=4` is the one-extra-bit safety candidate.

---

# D3 — compressed LinearTransform semantic correctness

For every candidate whose complete group remains q0/q1 unique:

- compare Fast compressed LinearTransform against genuine compressed Standard full-RNS semantics;
- require Fast q0/q1 rows to match the genuine Standard result reduced to q0/q1 after representation normalization;
- record max-component, max-complex and mean-complex error.

The compressed Fast transform must be modularly exact and semantically consistent with its compressed full-RNS reference.

If capacity is safe but residues disagree:

`logn13_c2s_group0_compression_linear_arithmetic_mismatch`.

---

# D4 — one ordinary group Rescale

For each complete-group-safe candidate, apply exactly one ordinary group Rescale using the same production logical dropped modulus as group 0.

Do not change the Rescale implementation.

Record:

- pre/post levels;
- pre/post scales;
- dropped modulus;
- capacity before Rescale;
- Fast-vs-compressed-Standard semantic and q0/q1 agreement after Rescale.

The purpose is to prove that reduced matrix encoding scale does not break the fixed one-level DFT schedule. Fast `Rescale` is source-defined to consume `LevelsConsumedPerRescaling()` rather than dynamically choosing a level from the Scale.

---

# D5 — representation restore after Rescale

After the compressed group-0 Rescale, restore the original downstream representation contract by the exact power of two `2^k`.

For the Fast candidate and its full-RNS compressed Standard reference:

1. multiply ciphertext physical coefficients by integer `2^k`;
2. multiply ciphertext metadata Scale by the same `2^k`;
3. do not consume a level;
4. do not change LogDimensions or domain flags.

This physical+metadata multiplication is a **semantic no-op**: it restores physical scale headroom and metadata without changing decoded plaintext values.

Do not use metadata-only promotion.

Record capacity before and after restore. The post-Rescale state is expected to have ample capacity, but this must be measured.

Require restored metadata to match the original baseline group-0 post-Rescale contract:

- same Level;
- same Scale within normal CKKS metadata tolerance;
- same LogDimensions;
- same NTT/Montgomery expectations for each backend.

---

# D6 — compare restored candidate against original genuine Standard group-0 output

The authoritative functional oracle is the **original uncompressed genuine Standard group-0 output after its ordinary Rescale**.

For each complete-group-safe restored candidate record:

### Compressed Standard restored vs original Standard

This isolates the precision cost of lower matrix encoding scale and restore.

Require max-component semantic error <= derived downstream-sensitive C2S budget.

### Fast compressed restored vs compressed Standard restored

Require max-component semantic error <= derived budget and, when q0/q1 uniqueness holds, exact projected q0/q1 agreement.

### Fast compressed restored vs original Standard

Require max-component semantic error <= derived C2S budget.

If the compressed Standard reference itself exceeds the budget relative to original Standard:

`logn13_c2s_group0_compression_encoding_precision_insufficient`.

If Standard compressed/restored passes but Fast does not:

`logn13_c2s_group0_compression_fast_mismatch`.

---

# D7 — choose full-group minimum and safety candidate

Derive from measured complete-group checkpoints:

- `k_min_full_group`: smallest tested k for which every group-0 LinearTransform checkpoint is q0/q1 unique and restored semantic error vs original Standard is within budget;
- `k_safe_full_group`: `k_min_full_group + 1` only if that candidate was actually tested and also passes all requirements.

Do not reuse predecessor `k_min/k_safe` labels without the `_full_group` qualifier.

Expected candidates are only `{2,3,4}` for this task.

Possible accepted classification:

`logn13_c2s_group0_compressed_linear_design_validated`

with both `k_min_full_group` and `k_safe_full_group` recorded.

If none pass:

`logn13_c2s_group0_compressed_linear_design_not_validated`.

---

# D8 — downstream probe after corrected group 0

Only after a valid `k_safe_full_group` candidate exists, perform one diagnostic downstream probe:

- start from the same common ModUp state;
- use the validated compressed+restore design for group 0 only;
- use the existing unmodified production matrices/operations for groups 1–3;
- complete C2S real/imag split;
- compare against genuine Standard C2S output using the existing derived downstream-sensitive budget.

This probe is **not** authorization to diagnose or fix later groups in this task.

Record only:

- final C2S real and imag deltas;
- first later group boundary whose post-Rescale semantic delta exceeds the derived C2S budget, if any;
- whether correcting group 0 alone is sufficient for the complete C2S budget.

Classify the downstream disposition separately, for example:

- `group0_fix_alone_satisfies_c2s_budget`
- `group0_fix_valid_but_next_budget_breach_group1`
- `group0_fix_valid_but_next_budget_breach_group2`
- `group0_fix_valid_but_next_budget_breach_group3`
- `group0_fix_valid_but_final_real_imag_split_breach`

Do not inspect internals of the later failing group in this task.

---

# D9 — preserve EvalMod separation

Record explicitly:

- `C2S_ERROR_ALONE_SUFFICIENT_FOR_FINAL_FAILURE` from the accepted causal evidence;
- `FAST_EVALMOD_MATCHED_INPUT_DIVERGENCE_REMAINS_UNRESOLVED = true`.

Do not touch EvalMod.

---

# Artifact

Create one compact artifact:

`results/FIX-001-P3-DESIGN-LOGN13-C2S-GROUP0-COMPRESSED-LINEAR-summary.json`

Include:

- provenance;
- baseline reproduction;
- candidate `{2,3,4}` summaries;
- per-candidate max capacity ratio and first unsafe checkpoint if any;
- modular residue-match booleans;
- pre/post-Rescale and restore metadata;
- restored semantic comparisons;
- `k_min_full_group` and `k_safe_full_group`;
- downstream probe disposition;
- final classification.

Do not serialize full coefficient vectors, slot vectors, encoded diagonals, or repeated trace arrays.

---

# Validation

Before completion require:

- focused `TestFIX001P3.*C2S.*Group0.*Compressed` tests pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- Primary diagnostic/design support and artifact committed/pushed;
- Primary worktree clean and `origin/main` synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no production fix.