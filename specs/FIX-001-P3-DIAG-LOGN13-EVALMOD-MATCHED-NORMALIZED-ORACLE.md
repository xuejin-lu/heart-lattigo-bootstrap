# FIX-001-P3-DIAG-LOGN13-EVALMOD-MATCHED-NORMALIZED-ORACLE

## Purpose

Diagnose the remaining independent LogN13 Fast EvalMod matched-input divergence after the C2S precision blocker has been resolved.

Accepted current state:

- Primary base: `3e4bb1bf4ba877ea3516292486e6dd12b331f832`
- Secondary exact clean: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- C2S group 0 accepted diagnostic design: compressed matrix `k=4`, ordinary Rescale, physical+metadata restore
- C2S group 1 accepted diagnostic design: compressed matrix `k=2`, ordinary Rescale, physical+metadata restore
- groups 2 and 3 execute once with ordinary production operations
- final split-only replay is validated
- final C2S real error vs ordinary genuine Standard: approximately `9.21e-14`
- final C2S imag error vs ordinary genuine Standard: approximately `9.44e-14`
- `C2S_PRECISION_BLOCKER_RESOLVED_FOR_LOGN13=true`

The remaining historical issue is independent:

- Fast EvalMod differs materially from genuine Standard EvalMod even on semantically matched inputs;
- previous raw Standard DoubleAngle q0/q1 square comparison was invalid after the genuine Standard full-RNS square exceeded centered `Q01/2` capacity;
- historical normalized DoubleAngle design was separately validated against a stage-aligned normalized full-RNS reference.

This task must establish a valid causal decomposition of EvalMod without using an invalid raw Standard q0/q1 projection.

No production fix is authorized in this task.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- modify C2S production code;
- integrate the accepted group-0/group-1 C2S compression into production;
- change Fast Mod1 production behavior;
- change polynomial coefficients, Mod1 parameters, target-scale schedule, DoubleAngle count, or final Scale contract;
- use a raw genuine-Standard DoubleAngle q0/q1 projection as an exact oracle after centered `Q01/2` uniqueness is lost;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- diagnose S2C, packing, unpacking, or finalization;
- start a production repair even if the cause becomes obvious.

Primary-only diagnostic helpers are allowed.

---

# Core oracle model

Use four logically distinct paths. Do not collapse them into one comparison.

## F — Fast production path

The actual current Fast Mod1 LogN13 path from Secondary:

1. metadata Scale normalization to `ScalingFactor()`;
2. source-defined cosine/Chebyshev offset;
3. `EvaluateWithPlanScale(..., targetScale, 2^91)`;
4. normalized LogN13 recurrence with metadata-only initial `K=2^29` normalization;
5. three normalized DoubleAngle rounds;
6. physical final `K=2^29` restore;
7. final metadata Scale reset to original Mod1 input Scale.

## N — normalized full-RNS stage-aligned mirror

A full-RNS reference implementing the **same compressed polynomial plan and same normalized recurrence as F**, but with full-RNS exact coefficients available for capacity reasoning.

This is the authoritative exact-row/q0q1 oracle for F.

Reuse the previously validated normalized full-RNS reference methodology from the accepted normalized DoubleAngle design. Do not substitute genuine Standard ordinary DoubleAngle ciphertext rows for N.

## C — coherent ordinary-DoubleAngle full-RNS path

Start from the same compressed-plan polynomial output used by N, but convert it to the ordinary DoubleAngle representation by the previously validated coherent promotion:

- physical coefficients `× M`;
- metadata Scale `× M`;
- `M = 2^29` derived from `targetScale / compressedPolynomialScale`;
- semantic plaintext value unchanged.

Then run the **ordinary full-RNS Standard DoubleAngle recurrence** for three rounds and final Scale reset.

C is semantic-only relative to N because independent Rescale histories need not have identical rows.

C isolates the semantic effect of the compressed polynomial plan from the normalized-recurrence transformation.

## S — genuine Standard EvalMod path

The ordinary genuine Standard Mod1 implementation:

1. same Scale normalization;
2. same source-defined offset;
3. ordinary full-RNS polynomial evaluation at Standard `targetScale`;
4. ordinary DoubleAngle recurrence;
5. final Scale reset.

S is the authoritative public semantic contract.

---

# E0 — construct corrected matched EvalMod inputs

Use the accepted corrected C2S diagnostic pipeline from the current Primary base:

1. common ModUp state;
2. group 0 `k=4` compressed + Rescale + physical+metadata restore;
3. group 1 `k=2` compressed + Rescale + physical+metadata restore;
4. groups 2 and 3 each exactly once;
5. split-only replay, without another DFT call.

Produce:

- Fast corrected C2S real input at Mod1 LevelQ;
- Fast corrected C2S imag input at Mod1 LevelQ;
- aligned full-RNS Standard real/imag semantic references;
- ordinary genuine Standard C2S real/imag outputs for external semantic comparison.

The real path is the primary localization path. The imag path may be final-confirmation-only unless it exhibits a different cause.

For the real path require before EvalMod:

- Level = Mod1 `LevelQ` (expected 12);
- Degree 1;
- Scale compatible with the C2S output contract;
- Fast NTT/Montgomery state valid;
- centered q0/q1 input capacity safe;
- corrected Fast vs ordinary genuine Standard semantic delta <= current derived C2S budget, expected around `1e-13`;
- no C2S factor is executed again.

If this fails:

`logn13_evalmod_matched_input_precondition_failure`.

---

# E1 — build the N full-RNS stage-aligned input

From the authoritative Fast q0/q1 corrected real input, construct a full-RNS zero-secret/stage-aligned reference N using the previously accepted centered-CRT reconstruction discipline:

1. convert authoritative q0/q1 to coefficient domain correctly;
2. leave Montgomery form correctly before centered CRT;
3. centered reconstruct exact integer coefficients from q0/q1;
4. redistribute those exact centered coefficients to all required Q limbs;
5. restore the representation/domain expected by the full-RNS diagnostic operations;
6. preserve Level, Degree, Scale, LogDimensions and semantic meaning.

Require at N input:

- exact q0/q1 rows agree with F after representation normalization;
- decoded F vs N semantic error = 0 or numerical roundoff only;
- N vs genuine Standard matched input semantic error <= derived C2S budget;
- N full-RNS exact coefficients are a valid unique lift of F q0/q1 at this boundary.

Do not create N by copying stale higher Fast limbs.

---

# E2 — Scale normalization and offset

Replay Scale normalization and source-defined Mod1 offset independently for F, N and S.

At checkpoints:

- `evalmod_input`
- `normalize_scale`
- `apply_offset`

record compactly:

- metadata;
- F q0/q1 capacity;
- N full-RNS exact capacity vs `Q01/2`;
- F q0/q1 exact rows vs N q0/q1 rows;
- F vs N semantic error;
- N vs S semantic error.

F and N must remain exact stage-aligned at these operations.

If F differs from N while N capacity is safe:

`logn13_evalmod_fast_prefix_primitive_mismatch`.

---

# E3 — polynomial-stage causal split

Compute the same source-defined `targetScale` as both Fast and Standard production code.

## N compressed-plan polynomial

Evaluate the Mod1 polynomial in full RNS using the same polynomial, same preprocessed input semantics and the same Fast LogN13 plan scale:

`planScale = 2^91`.

This must mirror the production Fast `EvaluateWithPlanScale` algorithm sufficiently to act as the stage-aligned full-RNS reference.

## F compressed-plan polynomial

Run the actual Fast production polynomial path.

## S Standard polynomial

Run the genuine Standard full-RNS polynomial evaluation at `targetScale`.

Record:

- polynomial output Level and Scale for F/N/S;
- F q0/q1 vs N q0/q1 exact row match;
- N full-RNS capacity vs `Q01/2`;
- F vs N semantic error;
- N compressed polynomial semantic output vs S Standard polynomial semantic output;
- Standard polynomial output capacity only as full-RNS evidence; do not require its q0/q1 rows to be a valid exact oracle when capacity is not unique.

### Correct plaintext polynomial oracle

Also evaluate the polynomial mathematically from decoded preprocessed input values using the already established basis convention:

`raw_chebyshev_on_preprocessed_z`.

Do not double-apply Chebyshev change of basis.

Record:

- N compressed polynomial vs plaintext oracle;
- S Standard polynomial vs plaintext oracle;
- F vs plaintext oracle.

This plaintext oracle is semantic-only; it is not an exact ciphertext-row oracle.

If F differs from N while N centered capacity is safe:

`logn13_evalmod_fast_polynomial_primitive_mismatch`.

Do not yet classify a semantic N-vs-S difference as the root cause until the causal hybrid paths below are completed.

---

# E4 — normalized recurrence N and Fast path F

From the compressed polynomial outputs, replay all three normalized rounds in F and N using the already validated source-defined recurrence:

For each round derive, do not blindly hardcode:

- current `K_i` exponent;
- next `K_{i+1}` exponent;
- `A_i = 2*K_i^2/K_{i+1}`;
- normalized constant `C_i = c_i/K_{i+1}`;
- expected next Scale after the ordinary one-level Rescale.

Expected historical LogN13 values may be `K_i=2^29` and `A_i=2^30`, but source-derived values are authoritative.

For rounds 0, 1, 2 record at:

- normalized round input;
- square;
- after integer A multiplier;
- after normalized constant;
- post-Rescale.

At every N checkpoint:

- measure full-RNS exact centered capacity vs `Q01/2` before using q0/q1 row equality;
- require F q0/q1 exact rows == N q0/q1 rows while capacity is unique;
- record F-vs-N semantic error.

If an N checkpoint itself exceeds `Q01/2`, do not project it and continue pretending it is an exact oracle. Classify:

`logn13_evalmod_normalized_reference_capacity_failure`.

If N is safe but F rows differ:

`logn13_evalmod_fast_normalized_primitive_mismatch`.

### Semantic mapping to ordinary y-space

At each normalized post-Rescale state, compare the mapped semantic value

`y_i = K_i * z_i`

in plaintext/decoded space against the corresponding genuine Standard ordinary DoubleAngle semantic state.

This is the valid cross-algorithm comparison.

Do **not** require N rows to equal S rows.

Do **not** q0/q1-project a genuine Standard raw square after its full-RNS coefficient magnitude exceeds `Q01/2`.

Record explicitly:

`RAW_STANDARD_DA_Q01_PROJECTION_DISPOSITION = invalid_after_centered_capacity_loss`.

---

# E5 — coherent ordinary-DA path C from the compressed polynomial

From an independent full-RNS copy of N's compressed polynomial output:

1. derive `M` from `targetScale / compressedPolynomialScale`;
2. physically multiply coefficients by `M`;
3. multiply metadata Scale by `M`;
4. verify semantic value is unchanged;
5. run the ordinary full-RNS Standard DoubleAngle recurrence for three rounds;
6. apply the ordinary final Scale reset.

This is the historical coherent path generalized to the current matched C2S input.

Do not use q0/q1 capacity to reject C merely because ordinary full-RNS intermediate coefficients exceed `Q01/2`; C is intentionally a full-RNS semantic oracle.

Record:

- compressed polynomial N vs coherent promoted C semantic equality before DA;
- N normalized final output vs C coherent final output;
- per-round mapped N `K_i*z_i` vs C ordinary `y_i` semantic deltas;
- final N vs C max-component error.

If F matches N exactly but N vs C alone exceeds the local semantic threshold materially enough to explain the final F-vs-S failure, classify:

`logn13_evalmod_normalized_recurrence_semantic_divergence`.

Otherwise N-vs-C is considered the normalization transformation cost and remains a measured term in the causal decomposition.

---

# E6 — compressed-polynomial causal isolation: C versus S

Compare the coherent ordinary-DA path C against genuine Standard EvalMod S.

Because C and S both use the ordinary full-RNS DoubleAngle recurrence, their principal algorithmic difference before DA is the polynomial evaluation representation/planning:

- C starts from the compressed-plan polynomial output;
- S starts from the Standard target-scale polynomial output.

Record:

- compressed polynomial vs Standard polynomial semantic delta before DA;
- C vs S semantic delta after DA round 0;
- after DA round 1;
- after DA round 2;
- final after Scale reset;
- amplification factor from polynomial-stage delta to final C-vs-S delta when denominator is nonzero.

If:

- F matches N stage-aligned;
- N and C are semantically equivalent within the final local threshold;
- C vs S fails the final local threshold;

then authoritative classification:

`logn13_evalmod_compressed_polynomial_semantic_divergence`.

This means q0/q1 Fast primitives and the normalized recurrence are not the primary cause; the compressed polynomial plan/precision/path is sufficient to produce the matched-input EvalMod failure.

---

# E7 — final restore and Scale reset

Validate the normalized path finalization independently:

1. before final physical K restore, measure N full-RNS exact capacity and prospective `K*B < Q01/2` bound;
2. physical restore in F and N using the same integer K;
3. require exact q0/q1 row equality F vs N;
4. require Level/Degree/domain unchanged;
5. record semantic N vs C and N vs S before Scale reset;
6. perform metadata-only final Scale reset to original Mod1 input Scale;
7. verify rows unchanged across reset;
8. compare final decoded F/N/C/S semantics.

If only the restore/reset diverges:

- `logn13_evalmod_final_restore_mismatch`, or
- `logn13_evalmod_final_scale_reset_mismatch`.

---

# E8 — authoritative matched-input final metrics

For corrected C2S real input record final outputs:

- F vs N;
- N vs C;
- C vs S;
- F vs S;
- N vs plaintext-oracle-derived expected result when available;
- S vs plaintext-oracle-derived expected result when available.

Use `1e-2` as the final local EvalMod semantic pass/fail threshold for this diagnostic, while preserving raw measured errors at all stages.

For imag input, at minimum run final F vs S and confirm whether the same classification is consistent. If imag exhibits a materially different stage/cause, report it separately and do not force one classification across both inputs.

If corrected matched inputs cause F vs S <= `1e-2` for both real and imag, classify:

`logn13_evalmod_matched_input_divergence_not_reproduced_after_c2s_correction`.

Do not assume the historical divergence must reproduce.

---

# E9 — required causal accounting

The summary must expose the final error decomposition conceptually as:

`F - S = (F - N) + (N - C) + (C - S)`

This is a semantic attribution identity, not a claim that max-absolute scalar metrics add algebraically.

Record each measured term separately:

- `fast_implementation_effect = F_vs_N`;
- `normalized_recurrence_effect = N_vs_C`;
- `compressed_polynomial_effect = C_vs_S`;
- `total_fast_vs_standard = F_vs_S`.

Do not infer causality from a single intermediate error magnitude when the corresponding hybrid final path has not been measured.

---

# Required classification

Choose exactly one primary classification for the real matched-input path:

- `logn13_evalmod_matched_input_precondition_failure`
- `logn13_evalmod_fast_prefix_primitive_mismatch`
- `logn13_evalmod_fast_polynomial_primitive_mismatch`
- `logn13_evalmod_normalized_reference_capacity_failure`
- `logn13_evalmod_fast_normalized_primitive_mismatch`
- `logn13_evalmod_normalized_recurrence_semantic_divergence`
- `logn13_evalmod_compressed_polynomial_semantic_divergence`
- `logn13_evalmod_final_restore_mismatch`
- `logn13_evalmod_final_scale_reset_mismatch`
- `logn13_evalmod_matched_input_divergence_not_reproduced_after_c2s_correction`
- `logn13_evalmod_semantic_cause_not_yet_isolated`

Also record:

- first failing stage/checkpoint or `none`;
- whether C2S precision blocker remains resolved (`true` expected);
- whether raw Standard DA q0/q1 projection was avoided (`true` required);
- whether Secondary was unchanged (`true` required).

---

# Artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-DIAG-LOGN13-EVALMOD-MATCHED-NORMALIZED-ORACLE-summary.json`

Include only:

- provenance and evaluator proof;
- corrected C2S matched-input evidence;
- targetScale / planScale / compressed polynomial Scale;
- compact polynomial semantic metrics;
- N full-RNS capacity summaries for normalized checkpoints;
- exact F-vs-N row-match booleans and first mismatch only;
- per-round semantic N-vs-C and C-vs-S metrics;
- final F/N/C/S metrics;
- real/imag final confirmation;
- causal-effect summary;
- classification and validation flags.

Do not serialize:

- full slot vectors;
- coefficient vectors;
- q-limb arrays;
- PS operation traces;
- repeated index arrays;
- raw polynomial coefficient dumps;
- thousands of checkpoint objects.

If detailed data is useful locally, keep it out of the committed summary.

---

# Validation

Before completion require:

- focused tests matching `TestFIX001P3.*EvalMod.*Matched.*Normalized` pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- no production fix;
- no C2S production integration;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- raw genuine-Standard DoubleAngle q0/q1 projection is never used as an exact oracle after centered capacity loss;
- Primary result committed and normally fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and `origin/main` synchronized.