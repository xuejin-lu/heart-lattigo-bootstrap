# FAST-OBS-001 — Unified Fast State & Invariant Measurement Framework

## Status

Executable Codex task.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- **read-only for this task**

Authoritative Secondary architecture:
- `docs/FAST_CKKS_SPEC.md`
- constitution commit: `89e71b2bce3a4b063343cbea9178828a32afbcf4`

## Purpose

Build the reusable observability/oracle infrastructure **before** implementing the widened Fast storage basis.

This task must make future debugging answer, from one common trace format:

1. What logical CKKS state was intended?
2. What Fast storage state represented it?
3. Was the lifted integer uniquely representable?
4. Was logical congruence preserved?
5. Did Scale/Level follow the constitution?
6. What is the first failing checkpoint/invariant?
7. At semantic model boundaries, what signal/error statistics are available for later FHE-aware quantization or noise-aware training?

Do not change Fast arithmetic semantics in this task.

## Architectural authority

The implementation must follow the durable constitution in Secondary `docs/FAST_CKKS_SPEC.md`, especially:

- logical `Q` owns CKKS Level/Scale/Rescale semantics;
- Fast private storage basis owns representation capacity only;
- `X ≡ c_logical (mod Q_level)`;
- centered uniqueness `|X| < S_A/2`;
- logical Level and ActiveStorageWidth are independent;
- Rescale uses logical `q_level`;
- storage contraction is representation-only;
- ModUp canonicalizes to the logical centered representative before extension.

If current historical q0/q1/q012 code conflicts with the constitution, treat the historical code as current implementation evidence, not as target semantics.

## Scope

Primary only.

Create reusable measurement/oracle code and tests. It may use synthetic integers and current logical parameter metadata. Do not require the future 60/60/60 production representation to exist yet.

Do not edit Secondary.

## Measurement modes

Define one stable enum/string mode:

- `OFF`
- `LIGHT`
- `FULL`

Semantics:

### OFF

No diagnostic reconstruction, no trace aggregation, no semantic decoding. Call sites should be able to use a no-op collector.

### LIGHT

Cheap metadata/capacity ledger only. Intended fields include:

- operation/checkpoint ID;
- stage ID;
- logical Level;
- logical Scale / log2 Scale;
- ciphertext degree when relevant;
- NTT/Montgomery flags when relevant;
- ActiveStorageWidth;
- active storage modulus bit lengths/product bit length when known;
- proven or measured `max_abs_X` bit length when available;
- centered capacity bit length;
- `headroom_bits`;
- operation-specific logical divisor or expected transition metadata.

LIGHT must not require full per-coefficient JSON dumps.

### FULL

Everything in LIGHT plus exact/reference checks where inputs are available:

- exact centered CRT reconstruction;
- active-residue consistency;
- logical congruence;
- exact storage round-trip;
- exact contraction identity;
- exact rounded logical Rescale oracle;
- ModUp canonicalization oracle;
- semantic/reference error summaries;
- first-divergence classification.

FULL may scan all coefficients but must store only compact summaries plus a small deterministic sample/index set by default.

## Orthogonal research purpose

Do **not** add a fourth measurement mode.

Instead, support optional collectors/fields for two consumers:

1. `debug_verification`
2. `ml_calibration`

The same checkpoint schema should serve both.

### ML calibration design rule

Model training / quantization does **not** need every low-level homomorphic multiplication as a direct training sample.

The future default semantic capture granularity should be model/operator boundaries such as:

- convolution output;
- polynomial/nonlinear activation output;
- pooling / packing boundary where semantically meaningful;
- fully connected output;
- Bootstrap input/output;
- optionally the major Bootstrap sub-stages.

Primitive-level `Mul / Rescale / Rotate / KeySwitch` checkpoints belong in FULL traces for learning and validating the cryptographic error model, not as the default model-quantization dataset.

Reserve schema fields sufficient for future semantic boundaries:

- `model_layer_id`
- `model_operator`
- `tensor_role`
- `branch` or channel identifier when applicable
- signal statistics
- error/residual statistics

Do not integrate a CNN or fingerprint model in this task.

## Why max|X| alone is not enough for ML calibration

For representation/capacity debugging, `max_abs_X` and `headroom_bits` are primary metrics.

For future quantization/noise-aware training, reserve compact distribution summaries. At minimum the schema must permit:

- count;
- min/max;
- mean;
- standard deviation;
- RMS;
- absolute max;
- quantiles (at least p50, p90, p99, p99.9 where sample count permits);
- the same summary for error/residual;
- optional small deterministic sample values/indices.

Do not assume FHE error is IID Gaussian. The framework must allow empirical, layer-conditioned residual distributions to be retained later.

Raw full tensors or all polynomial coefficients must not be emitted by default.

## Required common schema

Use typed Go structs rather than ad-hoc `map[string]interface{}` for the reusable core schema.

Suggested conceptual shape:

```text
FastMeasurementRun
  schema_version
  provenance
  mode
  purpose[]
  checkpoints[]
  summary

FastCheckpoint
  operation_id
  stage
  sequence
  before_state?
  after_state?
  logical
  storage
  capacity
  invariants
  semantic?
  ml_calibration?
  notes?

FastMeasurementSummary
  checkpoint_count
  first_failed_checkpoint
  first_failed_invariant
  minimum_headroom_bits
  minimum_headroom_checkpoint
  e2e_error?
```

Names may be improved, but the semantics above are required.

The output JSON must be deterministic in structure and versioned.

## Required mathematical helpers / exact oracles

Implement slow, correctness-first helpers using `math/big`. These are diagnostic/reference code, not production Fast arithmetic.

### O1 — Centered CRT

Given pairwise-coprime storage moduli and residues, reconstruct the unique centered representative

[
X \in (-S/2,S/2], \qquad S=\prod_i f_i.
]

Return enough metadata to determine capacity and sign.

### O2 — Residue consistency

Given `X`, verify each active residue equals

[
X \bmod f_i.
]

Handle negative `X` canonically.

### O3 — Logical congruence

Given `X`, logical reference `c`, and

[
Q_\ell=\prod_{i=0}^{\ell} q_i,
]

verify

[
X\equiv c\pmod{Q_\ell}.
]

Do not require raw integer equality between different valid lifts.

### O4 — Centered uniqueness / headroom

Given a proven/measured magnitude bound `B` and active storage product `S`, compute:

[
|X| < S/2
]

and

[
headroom\_bits
=
\log_2(S/2)-\log_2(B)
]

with an exact/integer-safe pass/fail test. Floating-point log2 is presentation only.

Handle `B=0` explicitly.

### O5 — Logical Rescale oracle

For signed integer `X` and positive odd logical divisor `q`, compute exact CKKS-style nearest rounded division

[
Y=\operatorname{Round}(X/q)
]

with the same tie convention required by current Fast/Standard semantics.

This oracle must be independent of any storage modulus.

### O6 — Storage contraction oracle

Given `X` and target storage basis, permit contraction only if the target basis can uniquely represent `X`.

When legal:

[
CRT_{target}(X \bmod f_i)=X
]

must hold exactly.

Contraction changes no logical Level and no Scale.

### O7 — ModUp canonicalization oracle

Given `X` and logical modulus product `Q_level`, compute:

[
C=Center_{Q_level}(X \bmod Q_level).
]

For Level 0 this becomes centered modulo logical `q0`.

This is the representative that future ModUp extension must use.

## Required invariant identifiers

Use stable machine-readable IDs for at least:

- `logical_congruence`
- `residue_consistency`
- `centered_unique`
- `scale_transition`
- `logical_level_transition`
- `storage_identity`
- `rescale_integer_result`
- `modup_canonicalization`

The summary's `first_failed_invariant` must use these IDs.

## Required first-divergence logic

Given an ordered list of checkpoints, return the earliest checkpoint with a failed invariant.

If no invariant fails, report none.

Do not conflate:
- invariant failure;
- semantic error threshold failure;
- missing/unavailable measurement.

Represent these separately.

## Required statistics helper

Add a reusable compact statistics helper for future semantic/ML calibration data.

Given a numeric vector, produce the distribution summary described above without retaining the whole vector in JSON.

A deterministic bounded sample/index helper is allowed.

Do not add external dependencies unless already present and clearly justified.

## Required tests

At minimum:

### T1 — Centered CRT round-trip

Random/deterministic values inside the centered range for 1-, 2-, and 3-modulus synthetic bases.

Verify exact round-trip.

Include negative values and boundary-near values.

### T2 — Residue consistency

Corrupt exactly one residue and verify the invariant fails.

### T3 — Logical congruence permits different lifts

Construct:

[
X=c+kQ
]

for positive and negative `k`.

Verify logical congruence passes even when `X != c`.

### T4 — Capacity/headroom

Test:
- safe positive margin;
- exact/near centered boundary;
- unsafe value;
- zero magnitude.

### T5 — Rescale theorem

For multiple logical chains and random/deterministic signed `X=c+kQ_level`, verify:

[
Round(X/q_level)
\equiv
Round(c/q_level)
\pmod{Q_{level-1}}.
]

This is a constitution test and must not depend on storage primes being equal to logical primes.

### T6 — Contraction

Verify:
- legal 3→2 exact identity;
- legal 2→1 exact identity;
- illegal contraction is rejected when the target centered range is insufficient.

### T7 — ModUp canonicalization

Show that two different lifts congruent modulo the old logical basis canonicalize to the same centered logical representative.

Also show that naively extending the arbitrary lifts would differ, documenting why canonicalization is required.

### T8 — First divergence

Create an ordered synthetic trace with several passing checkpoints and one failed invariant. Verify the summary identifies the exact first failure.

### T9 — OFF/LIGHT/FULL behavior

Verify:
- OFF produces no expensive oracle work / no checkpoints when collector is no-op;
- LIGHT can record metadata/headroom without exact CRT details;
- FULL records exact invariant results.

### T10 — Statistics

Verify deterministic statistics and quantiles on a known vector; verify no full raw vector is serialized by default.

## Performance / allocation requirement

This is infrastructure, not a benchmark task.

Still:
- OFF must be a true no-op path suitable for production integration later.
- LIGHT should avoid `big.Int` reconstruction unless the caller explicitly supplies a bound requiring it.
- FULL may be slow.

Do not optimize FULL prematurely.

## Files

Prefer a small coherent set such as:

- `fast_measurement.go`
- `fast_measurement_oracle.go`
- `fast_measurement_stats.go`
- matching `*_test.go`

Exact names may differ.

Do not create a new one-off FIX-001 runner for this framework.

## Existing diagnostics

Do not delete or rewrite historical diagnostic runners in this task.

They remain evidence.

This task creates the reusable foundation that later tasks can progressively adopt.

## Prohibitions

- No Secondary edits.
- No widened 60/60/60 production implementation.
- No Fast arithmetic semantic changes.
- No parameter tuning.
- No precision tightening.
- No threshold relaxation.
- No fingerprint/CNN integration.
- No assumption that FHE residual is Gaussian.
- No raw all-coefficient/all-activation JSON dump by default.
- No replacement of historical result files.

## Acceptance

This task passes only if:

1. all required oracle/property tests pass;
2. common typed schema exists;
3. OFF/LIGHT/FULL semantics are tested;
4. first-divergence summary is tested;
5. ML-calibration-compatible compact distribution schema exists;
6. no Secondary files changed;
7. Primary changes are committed and pushed normally.

## Completion report

Report concisely:

- files changed;
- tests run and results;
- schema version;
- brief example of one FULL checkpoint;
- confirmation that Secondary is unchanged;
- Primary commit SHA and push result.
