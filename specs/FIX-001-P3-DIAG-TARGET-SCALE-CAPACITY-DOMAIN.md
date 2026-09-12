# FIX-001-P3-DIAG-TARGET-SCALE-CAPACITY-DOMAIN — Correct the q0/q1 capacity oracle domain

## Purpose

The previous diagnostic `FIX-001-P3-DIAG-TARGET-SCALE-RESTORATION` stopped with:

`target_scale_restoration_centered_capacity_failure`

for canonical internal scale `2^91`, with:

- corrected pre-final polynomial error about `2.6027e-7`;
- corrected post-final Fast Rescale error about `2.6648e-7`;
- `M = 536870912 = 2^29`;
- `Log2Delta(S*M,T) = 43.1671...`;
- reported P0 capacity ratio about `0.999871`;
- reported P1 prospective ratio about `5.368e8`;
- reported P1 outside count `8192`.

Independent review found that the capacity evidence was constructed in the wrong representation domain:

1. `q01Projection(..., true)` copies q0/q1 and performs `IMForm`, but preserves `IsNTT=true` and performs no INTT.
2. `targetScaleCenteredRows` then directly applies CRT to each row index and interprets those NTT-domain residues as centered coefficient-domain integers.
3. Secondary `schemes/ckks/fast/partial_ntt.go` explicitly states that `FastPartialINTT` produces canonical coefficient-domain q0/q1 residues suitable for the CRT precondition once Montgomery form is also removed.

Therefore the prior P0/P1 capacity numbers are not valid coefficient-domain centered-capacity evidence.

This task must correct only that diagnostic oracle, reproduce the exact canonical `2^91` state, and determine whether the target-scale restoration actually passes or fails once capacity is measured in the proper coefficient domain.

Diagnostic only. LogN13 only.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`735bc5ea0279dd19e8b2c7437260d585cb433fb4`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Authoritative polynomial oracle:

`raw_chebyshev_on_preprocessed_z`

Semantic threshold:

`1e-2`

Canonical candidate:

`2^91`

Do not modify Secondary production code.

Do not full-read large historical result JSON files.

---

# Superseded evidence

The following prior evidence must be retained only as historical diagnostic output and must **not** participate in pass/fail:

- P0 capacity ratio `0.9998711804284713`;
- P1 prospective capacity ratio `536801752.51914996`;
- P1 outside count `8192`;
- classification `target_scale_restoration_centered_capacity_failure`.

Label these values explicitly as:

`invalid_ntt_domain_capacity_view`

until/unless a corrected coefficient-domain calculation independently reproduces the same conclusion.

Do not delete or rewrite the historical artifact.

---

# Scope lock

Do not:

- run lower internal scales `2^90 ... 2^86`;
- run LogN16;
- benchmark;
- run Gate 4/5 production experiments;
- start EXP-003;
- enter DoubleAngle;
- modify production Lattigo;
- alter polynomial coefficients, PS schedule, Mod1 parameters, target Scale, final Fast Rescale, metadata-normalization rule, or semantic threshold;
- use NTT-domain CRT magnitudes as coefficient-domain capacity evidence;
- accept the prior capacity classification without reproducing it in the corrected domain.

The canonical question is:

> After transforming the maintained q0/q1 rows to canonical coefficient-domain non-Montgomery residues, does multiplication by the required target-scale integer `M=2^29` remain uniquely representable under `Q01=q0*q1`, and if so does the full target-scale restoration pass P2–P5?

---

# Mandatory reproduction

Reproduce the exact `2^91` path through the actual final Fast Rescale.

Require:

1. authoritative pre-final q0/q1 hashes match the known canonical state;
2. corrected pre-final semantic error remains in the `~2.60e-7` regime;
3. corrected post-final semantic error remains in the `~2.66e-7` regime and <= `1e-2`;
4. post-final Scale `S` reproduces approximately `2.147483648e9`;
5. public target `T` reproduces approximately `1.1529215046069e18`;
6. nearest positive integer `M` remains `536870912`;
7. `Log2Delta(S*M,T)` remains >= 32.

If any of these fail, stop with:

`TARGET_SCALE_CAPACITY_DOMAIN_PRECONDITION_MISMATCH`.

---

# C0 — prove the representation state

Before any capacity computation, record the post-final ciphertext representation metadata:

- `IsNTT`;
- `IsMontgomery`;
- Level;
- Degree;
- Scale;
- q0/q1 row hashes.

The task must explicitly state whether the historical capacity helper consumed NTT-domain rows.

If the historical input was not NTT-domain, stop with:

`TARGET_SCALE_CAPACITY_DOMAIN_ASSUMPTION_MISMATCH`.

---

# C1 — retain the historical NTT-domain view for comparison only

On an isolated copy, reproduce the previous helper behavior exactly:

- q0/q1 projection;
- Montgomery normalization as previously done;
- direct per-index CRT without INTT.

Store its compact P0/P1 ratios and outside count under:

`historical_ntt_domain_view`.

This view is diagnostic comparison only.

It must never set the final classification.

---

# C2 — construct the authoritative coefficient-domain q0/q1 view

Starting from an independent copy of the exact post-final ciphertext:

1. copy only authoritative q0/q1 rows;
2. preserve the original representation long enough to mirror production domain conversion;
3. if `IsNTT == true`, apply Secondary `fast.FastPartialINTT` to q0/q1;
4. after INTT, if the source representation was Montgomery, apply `IMForm` separately to q0 and q1;
5. mark the diagnostic copy logically as coefficient-domain and non-Montgomery;
6. do not touch or read limbs >=2;
7. do not mutate the source ciphertext.

The conversion order should mirror the production Fast Rescale semantics:

`FastPartialINTT -> IMForm (when needed) -> centered CRT`.

For every component and coefficient index, CRT-reconstruct q0/q1 and center into:

`[-Q01/2, Q01/2)`.

This C2 view is the only authoritative capacity domain in this task.

---

# C3 — coefficient-domain round-trip validation

Before trusting C2, prove the transform itself is not corrupting q0/q1.

On an independent copy:

1. start from the original maintained NTT/Montgomery q0/q1 rows;
2. perform the C2 conversion to coefficient/non-Montgomery form;
3. re-enter Montgomery form if required by the original representation;
4. apply `FastPartialNTT` to q0/q1;
5. compare resulting q0/q1 rows with the original source rows exactly.

Require zero mismatches.

Also add a focused unit test using a deterministic nontrivial q0/q1 polynomial proving:

- direct CRT of NTT row indices generally differs from CRT after INTT;
- `FastPartialINTT -> FastPartialNTT` round-trip restores q0/q1 exactly under the appropriate representation conversions.

If round-trip fails, stop with:

`target_scale_capacity_domain_transform_failure`.

---

# C4 — corrected P0 capacity

Using only the C2 coefficient-domain data, compute compact P0 evidence:

- maximum absolute centered coefficient per component;
- overall maximum absolute centered coefficient;
- `maxAbs / (Q01/2)`;
- outside count;
- worst component/index/value.

Require P0 itself to satisfy:

`|x| < Q01/2`

for all maintained coefficients.

If not:

`FIRST_SUPPORTED_CAUSE = target_scale_post_rescale_centered_capacity_failure`

Stop. Do not attempt promotion.

---

# C5 — corrected prospective P1 capacity

For every authoritative coefficient-domain centered value `x`, compute using unbounded integer arithmetic:

`x_promoted = M*x`

with fixed:

`M = 536870912`.

Require:

`|M*x| < Q01/2`

for every maintained coefficient required by the promoted state.

Record compact evidence:

- maximum `|x|`;
- maximum `|M*x|`;
- maximum prospective ratio;
- outside count;
- worst tuple.

### If corrected C5 fails

Classify:

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_centered_capacity_failure_confirmed`

This is the first point at which the earlier capacity failure becomes accepted as a real Fast-CKKS limitation.

Stop. Do not run P2–P5.

### If corrected C5 passes

The historical capacity stop was a harness error. Continue immediately to P2–P5 below in the same task.

Do not create another task merely to resume the already-specified restoration sequence.

---

# P2 — actual integer promotion

Only if C5 passes.

Starting from the original post-final NTT/Montgomery ciphertext copy, execute exactly:

`MulIntegerMaintained(result, M, result)`.

Do not change Scale yet.

Verify both:

1. q0/q1 residues equal independent modular multiplication of the original rows by `M`;
2. decoded semantics at original Scale match `M * authoritativePolynomialOracle`, with normalized semantic error <= `1e-2`.

Additionally create a C2-style coefficient-domain view of the promoted ciphertext and verify it agrees with the modular image of the expected promoted coefficient data.

If not:

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_integer_promotion_failure`

---

# P3 — matched Scale

Set only:

`Scale = S*M`.

Decode against the unchanged authoritative polynomial oracle.

Require:

- semantic error <= `1e-2`;
- no coefficient mutation caused by metadata change;
- perturbation from P0 recorded compactly.

If not:

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_matched_scale_failure`

---

# P4 — exact target metadata normalization

Require:

`Log2Delta(S*M,T) >= 32`.

Then change only metadata:

`Scale = T`.

Decode against the authoritative polynomial oracle.

Require:

- semantic error <= `1e-2`;
- q0/q1 hashes unchanged from P2/P3;
- decoded perturbation from P3 recorded.

If not:

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_metadata_normalization_failure`

If the `Log2Delta` precondition unexpectedly fails:

`FIRST_SUPPORTED_CAUSE = target_scale_restoration_integer_approximation_too_coarse`

---

# P5 — final restored-state invariants

Require:

- exact public target Scale `T`;
- Level and Degree preserved as intended;
- corrected semantic error <= `1e-2`;
- q0/q1 hashes stable through P3/P4 metadata-only normalization;
- corrected coefficient-domain capacity evidence retained;
- no O1 / double-change-of-basis oracle participates in pass/fail.

If all pass:

`FIRST_SUPPORTED_CAUSE = compressed_ps_2p91_target_scale_restoration_validated_after_capacity_domain_fix`

This validates LogN13 through exact target-scale restoration only.

It does not authorize DoubleAngle, LogN16, benchmark, Gate 4/5, EXP-003, or production integration.

---

# Required diagnostic classification

In addition to `FIRST_SUPPORTED_CAUSE`, record exactly one diagnostic-oracle disposition:

- `CAPACITY_ORACLE_DISPOSITION = historical_ntt_domain_capacity_invalid_corrected_domain_passed`
- `CAPACITY_ORACLE_DISPOSITION = historical_ntt_domain_capacity_invalid_corrected_domain_also_fails`
- `CAPACITY_ORACLE_DISPOSITION = capacity_domain_transform_unresolved`
- `CAPACITY_ORACLE_DISPOSITION = capacity_domain_precondition_mismatch`

The historical `target_scale_restoration_centered_capacity_failure` classification is superseded unless C5 independently confirms it in coefficient domain.

---

# Artifact discipline

Create compact artifacts only:

- `results/FIX-001-P3-DIAG-TARGET-SCALE-CAPACITY-DOMAIN-logN13.json`
- `results/FIX-001-P3-DIAG-TARGET-SCALE-CAPACITY-DOMAIN-logN13-summary.json`

Summary must remain human-reviewable.

Store only:

- provenance;
- C0 representation metadata;
- historical NTT-domain compact ratios/outside count;
- coefficient-domain transform/round-trip status;
- corrected C4/C5 compact capacity evidence;
- `S`, `T`, `M`, `Log2Delta`;
- P2–P5 compact checkpoint evidence when executed;
- oracle disposition;
- first supported cause;
- tests/clean-state evidence.

Do not serialize full coefficient vectors, slot vectors, repeated indices, operation traces, or large hash collections.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- focused capacity-domain tests pass;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Secondary changes;
- compact artifacts committed/pushed normally;
- both worktrees clean;
- no lower-scale sweep;
- no DoubleAngle;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003.

The deliverable is a representation-correct answer to whether `2^91` target-scale restoration is actually capacity-safe under q0/q1.