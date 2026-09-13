# FIX-001-P3-DIAG-LOGN13-C2S-FINAL-SPLIT-PROBE-FIX

## Purpose

Correct and re-run the LogN13 C2S downstream probe after the accepted group-0 and group-1 compressed designs.

The predecessor group-1 compressed design itself is accepted, but its downstream disposition

`groups0_1_fix_valid_but_final_real_imag_split_breach`

is **not accepted** because the diagnostic harness appears to execute C2S groups 2 and 3 twice before the final real/imag split.

Accepted predecessor:

- Primary result commit: `2c2e69717fbdaa38500968756a4392e156be478e`
- Secondary exact clean: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- group 0 accepted safety design: `k=4`
- group 1 accepted minimum/safety design: `k_min_full_group=1`, `k_safe_full_group=2`
- corrected group-1 post-Rescale error: approximately `3.67e-13`
- derived C2S budget: approximately `1.953125e-5`

Important harness defect to verify:

The predecessor downstream probe manually executed original C2S matrix groups 2 and 3, then constructed a `dft.Matrix` containing `Matrices[2:]` and called `FastEvaluator.CoeffsToSlotsNew` on the already group-3-completed ciphertext. Source inspection shows `CoeffsToSlotsNew -> CoeffsToSlots -> eval.dft(...)` before Conjugate/Sub/Mul/Add in split mode. Therefore that call executes groups 2 and 3 again before splitting.

This task is diagnostic/harness correction only. Do not modify Secondary production code.

---

## Fixed provenance and scope

Primary required base:

`2c2e69717fbdaa38500968756a4392e156be478e`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- LogN13 only
- deterministic `reproducibleInput.v1`
- C2S groups `[1,1,1,1]`
- group 0 diagnostic corrected design: compressed matrix `k=4`, ordinary Rescale, physical+metadata restore
- group 1 diagnostic corrected design: compressed matrix `k=2`, ordinary Rescale, physical+metadata restore
- groups 2 and 3: unchanged original production matrices and operations, each executed exactly once

Do not:

- modify Secondary;
- modify production DFT matrix generation;
- modify Fast LinearTransform, Fast Rescale, Conjugate, Add/Sub/Mul, EvalMod, packing, or finalization;
- implement a production fix;
- run LogN16, benchmark, Gate 4/5, or EXP-003;
- diagnose EvalMod in this task.

---

# P0 — reproduce accepted corrected prefix

Reproduce:

1. genuine Standard evaluator proof;
2. Standard end-to-end bootstrap <= `1e-2`;
3. common ModUp state;
4. accepted group-0 `k=4` compressed + Rescale + restore;
5. accepted group-1 `k=2` compressed + Rescale + restore.

Require at the restored group-1 boundary:

- Fast and aligned full-RNS Standard metadata compatible;
- capacity safely within centered q0/q1 range;
- Fast residues match genuine full-RNS Standard reduced mod q0/q1;
- semantic delta <= derived C2S budget;
- corrected group-1 error consistent with predecessor (~`3.67e-13`).

If this does not reproduce:

`logn13_c2s_split_probe_prefix_mismatch`.

---

# P1 — execute groups 2 and 3 exactly once

Starting from the corrected/restored group-1 state, replay original production groups 2 and 3 once, in source order:

For each group:

1. original production LinearTransform;
2. exactly one ordinary Rescale.

Maintain aligned paths:

- Fast q0/q1 path;
- genuine full-RNS Standard path.

At each post-LinearTransform and post-Rescale boundary record compactly:

- level and Scale;
- Fast q0/q1 capacity;
- Standard full-RNS capacity vs `Q01/2`;
- centered uniqueness;
- Fast residue match against Standard reduced mod q0/q1;
- semantic Fast-vs-Standard error when the q0/q1 projection is a valid oracle.

Do not inspect internal group-2/group-3 BSGS checkpoints unless a boundary unexpectedly fails the already observed downstream-budget behavior.

After group 3, call this aligned state `zV_fast` / `zV_standard`.

Require:

- groups 2 and 3 have each executed exactly once;
- no further DFT factor remains before split;
- post-group3 error <= derived C2S budget.

If group 2 or group 3 post-Rescale now exceeds the budget, classify the earliest boundary and stop split attribution.

---

# P2 — prove predecessor downstream probe double-applied the DFT tail

Record source-grounded control-flow evidence:

- predecessor probe manually replayed groups 2 and 3;
- predecessor then passed the already group-3-completed ciphertext to `FastEvaluator.CoeffsToSlotsNew` with a matrix containing `Matrices[2:]` / `Levels[2:]`;
- Fast `CoeffsToSlots` in split mode calls `eval.dft(zV, matrices, zV)` before Conjugate/Sub/Mul/Add;
- therefore predecessor final real/imag numbers were measured after a second execution of groups 2 and 3.

Record:

`PREDECESSOR_FINAL_SPLIT_DISPOSITION_VALID=false`

and

`PREDECESSOR_DOUBLE_DFT_TAIL_CONFIRMED=true`

if the control flow matches.

Do not reproduce the bad path unless a minimal confirmation is needed; do not use it as an oracle.

---

# P3 — split-only replay from completed zV

Do **not** call `CoeffsToSlotsNew`, `CoeffsToSlots`, or any function that invokes DFT factors.

Replay only the post-DFT split primitives in exactly the Fast source order on `zV_fast`, with a genuine Standard full-RNS replay on `zV_standard`:

1. `conj = Conjugate(zV)`
2. `imag_tmp = zV - conj`
3. `imag = imag_tmp * (-1i)`
4. `real = conj + zV`

For each checkpoint record:

- metadata;
- Fast q0/q1 capacity;
- Standard full-RNS capacity vs `Q01/2`;
- Fast residue equality to Standard reduced mod q0/q1;
- semantic Fast-vs-Standard error while centered q0/q1 uniqueness is valid;
- input immutability where relevant.

Use exact checkpoint names:

- `split_input_zv`
- `split_conjugate`
- `split_imag_sub`
- `split_imag_mul_minus_i`
- `split_real_add`

Classify the earliest safe mismatch, if any:

- `logn13_c2s_final_split_conjugate_mismatch`
- `logn13_c2s_final_split_imag_sub_mismatch`
- `logn13_c2s_final_split_imag_mul_mismatch`
- `logn13_c2s_final_split_real_add_mismatch`

If an exact full-RNS split checkpoint exceeds centered q0/q1 uniqueness while Fast remains modulo-correct, classify the corresponding split alias rather than decoding an invalid q0/q1 oracle.

---

# P4 — authoritative final C2S comparison

Compare split-only `real` and `imag` against the genuine Standard C2S outputs produced by the ordinary full-RNS `CoeffsToSlots` path from the same common pre-C2S state.

Record for real and imag:

- max-component absolute error;
- max-complex absolute error;
- mean-complex error;
- worst slot/component;
- pass/fail against derived C2S budget.

Also compare split-only Fast against aligned split-only full-RNS Standard.

If all groups and split primitives pass and final real/imag are <= budget, authoritative classification:

`logn13_c2s_corrected_groups0_1_final_split_validated`

and disposition:

`groups0_1_compression_sufficient_for_c2s_budget`.

If final comparison fails despite all aligned split checkpoints passing, classify:

`logn13_c2s_final_reference_mismatch_requires_diagnosis`

and record enough compact evidence to identify the reference mismatch. Do not begin a new diagnosis in this task.

---

# P5 — preserve EvalMod separation

Record explicitly:

- `C2S_ERROR_ALONE_SUFFICIENT_FOR_FINAL_FAILURE` according to prior causal history;
- `FAST_EVALMOD_MATCHED_INPUT_DIVERGENCE_REMAINS_UNRESOLVED=true`.

If corrected C2S now passes the derived budget, record:

`C2S_PRECISION_BLOCKER_RESOLVED_FOR_LOGN13=true`

but do not claim full bootstrap correctness because the independent Fast EvalMod matched-input divergence remains unresolved.

Do not call EvalMod in this task.

---

# Artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-DIAG-LOGN13-C2S-FINAL-SPLIT-PROBE-FIX-summary.json`

Include only:

- provenance;
- corrected prefix checks;
- group-2/group-3 boundary summaries;
- predecessor double-DFT proof booleans;
- five split checkpoint summaries;
- final real/imag comparisons;
- C2S blocker disposition;
- EvalMod unresolved flag;
- classification and validation flags.

Do not serialize full slot vectors, coefficient vectors, encoded diagonals, BSGS traces, or repeated per-index arrays. Keep this summary materially smaller than the predecessor design artifact.

---

# Validation

Before completion require:

- focused `TestFIX001P3.*C2S.*Split.*Probe` tests pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- no production fix;
- no EvalMod work;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- Primary task result committed and, under standing safe-push authorization, normally fast-forward pushed to `origin/main` without a separate confirmation when all repository safety conditions hold;
- Primary worktree clean and `origin/main` synchronized.