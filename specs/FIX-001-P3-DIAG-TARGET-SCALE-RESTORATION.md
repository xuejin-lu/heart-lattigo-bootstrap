# FIX-001-P3-DIAG-TARGET-SCALE-RESTORATION — Validate LogN13 public target-scale restoration

## Purpose

The Chebyshev-oracle diagnostic established for canonical compressed internal scale `2^91`:

- `e2Decoded` is already the preprocessed Chebyshev variable `z`;
- authoritative polynomial oracle is `raw_chebyshev_on_preprocessed_z`;
- the historical `~0.0114439` failure came from applying `Polynomial.Evaluate` to `z` and therefore applying the interval change-of-basis a second time;
- corrected pre-final polynomial error is about `2.60e-7`;
- corrected post-final Fast Rescale error is about `2.66e-7`;
- both pass the fixed semantic threshold `1e-2`.

Therefore the next unresolved boundary in the original compressed-PS plan is **restoring the public target Scale after the compressed polynomial result**.

This task must validate that restoration for the canonical highest candidate `2^91` only.

Diagnostic only. LogN13 only.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`4397843a0028a00fe4d4a4d98b7ce552e2c13b36`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Authoritative polynomial oracle:

`raw_chebyshev_on_preprocessed_z`

Semantic threshold:

`1e-2`

Both worktrees must start clean and follow their `AGENTS.md` startup rules.

Do not modify Secondary production code.

---

# Scope lock

Do not:

- run `2^90 ... 2^86` unless a later task explicitly requests them;
- run LogN16;
- benchmark;
- run Gate 4/5 production experiments;
- start EXP-003;
- modify production Lattigo;
- change polynomial coefficients, Mod1 parameters, PS schedule, generated powers, metadata-normalization rule, or semantic threshold;
- use the historical O1 / `Mod1Poly.Evaluate(e2Decoded)` oracle;
- proceed into the Mod1 DoubleAngle chain in this task;
- implement the validated restoration as production behavior.

The only question is:

> Can the already-correct canonical `2^91` compressed polynomial output be restored to the public target Scale without violating q0/q1 centered-capacity assumptions or semantic correctness?

---

# Mandatory reproduction

Reproduce the exact canonical LogN13 `2^91` path through the actual final Fast Rescale.

Before attempting target-scale restoration require:

1. authoritative pre-final q0/q1 hashes match the known canonical state;
2. corrected pre-final error remains in the `~2.60e-7` regime;
3. corrected post-final Fast Rescale error remains in the `~2.66e-7` regime and <= `1e-2`;
4. authoritative oracle is explicitly recorded as `raw_chebyshev_on_preprocessed_z`;
5. no historical O1-based failure is used for control flow.

If not, stop with:

`TARGET_SCALE_RESTORATION_PRECONDITION_MISMATCH`.

---

# Restoration construction

Let:

- `S = compressedOutput.Scale` after the actual final Fast Rescale;
- `T = public targetScale` required by the polynomial-evaluation schedule.

Using high-precision arithmetic compute:

`R = T / S`.

Choose:

`M = nearest positive integer(R)`.

Record:

- exact/high-precision `S`;
- exact/high-precision `T`;
- `R`;
- chosen integer `M`;
- relative difference between `M` and `R`;
- `Log2Delta(S*M, T)`.

Require `M > 0`.

Do not silently floor, ceil, clamp, or choose another integer merely to make the diagnostic pass.

---

# P0 — baseline compressed output

Snapshot the exact post-final-Rescale ciphertext before promotion.

Record:

- corrected semantic max component error;
- Level/Degree/Scale;
- q0/q1 residue hashes;
- per-component centered q0/q1 reconstruction capacity:
  - maximum absolute centered coefficient;
  - `maxAbs / (q0*q1/2)`;
  - outside count.

Require corrected semantic error <= `1e-2`.

---

# P1 — prospective centered-capacity proof

Before mutating the ciphertext, independently reconstruct every maintained q0/q1 coefficient as a centered integer under `Q01 = q0*q1`.

For every component/coefficient compute the exact mathematical product:

`M * centeredCoefficient`.

Require:

`|M * centeredCoefficient| < Q01/2`

for every maintained coefficient needed by the restored result.

Record only compact evidence:

- maximum absolute pre-promotion centered magnitude;
- maximum absolute prospective promoted magnitude;
- maximum prospective capacity ratio;
- outside count;
- worst component/index/value tuple.

If this fails, stop with:

`target_scale_restoration_centered_capacity_failure`.

Do not rely on modular wraparound as proof of future q0/q1 centered semantics.

---

# P2 — integer coefficient promotion

On an independent maintained copy execute exactly:

`MulIntegerMaintained(result, M, result)`.

Do not change Scale yet.

Decode immediately with the existing q0/q1 diagnostic decode.

Because Scale is still `S`, the expected plaintext is:

`M * authoritativePolynomialOracle`.

Compare against that scaled plaintext oracle.

Require <= `1e-2 * max(1,M)` only for this intentionally M-scaled intermediate comparison, and additionally record the normalized error divided by `M`.

Prefer an equivalent relative/normalized formulation that avoids overflow when reporting metrics.

If the normalized semantic error exceeds `1e-2`, classify:

`target_scale_restoration_integer_promotion_failure`.

Also verify the resulting maintained residues equal an independent modular multiplication of the P0 q0/q1 residues by `M`.

---

# P3 — matched Scale update

Without changing coefficients again, set:

`result.Scale = S * M`.

Decode against the unchanged authoritative polynomial oracle.

This coefficient-and-scale promotion should be semantically identity-preserving.

Require corrected semantic error <= `1e-2`.

Record semantic perturbation relative to P0.

If it fails, classify:

`target_scale_restoration_matched_scale_failure`.

---

# P4 — exact public target metadata normalization

Compare `S*M` with exact public target `T`.

Metadata-only normalization to exact `T` is allowed diagnostically only if:

`Log2Delta(S*M, T) >= 32`.

This uses the same fixed `2^-32` diagnostic metadata-drift bound previously used for giant-step normalization.

If the bound is not met, classify:

`target_scale_restoration_integer_approximation_too_coarse`.

If allowed:

1. snapshot P3;
2. change **only** `result.Scale = T`;
3. do not modify q0/q1 coefficients;
4. decode immediately;
5. compare with the authoritative polynomial oracle;
6. measure decoded-vector perturbation between P3 and P4.

Require final corrected semantic error <= `1e-2`.

If it fails, classify:

`target_scale_restoration_metadata_normalization_failure`.

---

# P5 — restored-state invariants

For a passing P4 state require:

- Scale equals exact public target `T`;
- Level and Degree are unchanged from P0 except for no unexpected mutation;
- q0/q1 coefficient hashes are unchanged from P2/P3 through P4 metadata normalization;
- centered-capacity evidence remains valid;
- corrected semantic error <= `1e-2`;
- no historical O1 oracle participates in pass/fail.

Do not execute DoubleAngle.

---

# Required classification

Choose exactly one:

## Case A

`FIRST_SUPPORTED_CAUSE = compressed_ps_2p91_target_scale_restoration_validated`

P0 through P5 pass.

This means the diagnostic compressed PS path for the highest candidate `2^91` is numerically valid through:

- baby blocks;
- normalized giant steps;
- final Fast Rescale;
- restoration to exact public target Scale.

It does **not** yet authorize production integration or LogN16.

## Case B

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_centered_capacity_failure`

Prospective multiplication by `M` violates the required q0/q1 centered uniqueness interval.

## Case C

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_integer_promotion_failure`

`MulIntegerMaintained` does not match the independently checked expected multiplication semantics.

## Case D

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_matched_scale_failure`

Coefficient multiplication is locally correct, but matched `Scale *= M` does not restore the original polynomial semantics within threshold.

## Case E

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_integer_approximation_too_coarse`

Nearest integer promotion does not bring `S*M` sufficiently close to `T` for the fixed `Log2Delta >= 32` metadata-normalization rule.

## Case F

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_metadata_normalization_failure`

The integer approximation is sufficiently close, but metadata-only normalization to exact `T` causes semantic error > `1e-2`.

## Case G

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_precondition_mismatch`

The corrected canonical state cannot be reproduced.

---

# Why only `2^91`

`2^91` is the highest candidate in the fixed candidate set and has already passed the corrected polynomial path through final Rescale.

The prior deterministic preference rule chooses the **highest internal scale** among otherwise valid candidates.

Therefore:

- if `2^91` passes restoration, it is already the preferred candidate and lower candidates are unnecessary for selection;
- if `2^91` fails restoration, stop and report the precise boundary instead of automatically broadening the run.

A later task may decide whether lower candidates are scientifically useful after seeing that failure mode.

---

# Artifact discipline

Create compact artifacts only:

- `results/FIX-001-P3-DIAG-TARGET-SCALE-RESTORATION-logN13.json`
- `results/FIX-001-P3-DIAG-TARGET-SCALE-RESTORATION-logN13-summary.json`

Summary must remain human-reviewable.

Store:

- exact provenance;
- `S`, `T`, `R`, `M`, relative integer-approximation error, `Log2Delta`;
- P0-P5 checkpoint table;
- compact capacity evidence;
- independent modular multiplication check;
- semantic errors and metadata-normalization perturbation;
- exact final classification;
- tests and clean-state evidence.

Do not store full slot vectors, coefficient arrays, repeated indices, or operation traces.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- focused diagnostic tests pass;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Secondary changes;
- Primary compact artifacts are committed and pushed normally;
- both worktrees are clean;
- no lower candidate sweep;
- no LogN16;
- no benchmark;
- no Gate 4/5 production run;
- no EXP-003;
- no DoubleAngle execution.

The deliverable is whether canonical `2^91` can be restored from its correct compressed output Scale to the exact public target Scale while preserving q0/q1 safety and corrected polynomial semantics.