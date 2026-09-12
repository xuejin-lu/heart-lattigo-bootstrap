# FIX-001-P3-DIAG-CHEBYSHEV-ORACLE-BASIS — Resolve the plaintext oracle construction mismatch

## Purpose

The completed LogN13 PS global-semantics diagnostic stopped with:

`FIRST_SUPPORTED_CAUSE = ps_plaintext_oracle_construction_mismatch`

Canonical `2^91` evidence:

- generated powers pass;
- baby checks pass;
- giant checks pass;
- authoritative pre-final q0/q1 hash matches the prior normalization run;
- PS-root source-backed error is about `2.60e-7`;
- the current direct whole-polynomial oracle reports about `0.0114439786`;
- PS-root plaintext construction versus that direct oracle differs by about `0.0114440390`.

The current diagnostic runner constructs:

- PS/root expectations from `PatersonStockmeyerPolynomial(...).Value` and explicit Chebyshev powers;
- the direct oracle using `Mod1Poly.Evaluate(e2Decoded)`.

Pinned Lattigo source establishes two important semantics:

1. `bignum.Polynomial.Evaluate` automatically applies `ChangeOfBasis()` when `Basis == Chebyshev`, mapping the supplied input from polynomial interval `[A,B]` to the Chebyshev variable before evaluating the recurrence.
2. CKKS polynomial evaluator documentation states that, for Chebyshev input, the caller must apply the change of basis to the ciphertext **before** polynomial evaluation; the generic homomorphic evaluator treats the supplied ciphertext as the power-basis variable.
3. Fast and Standard Mod1 both perform the caller-side Mod1/Chebyshev preprocessing before invoking polynomial evaluation.

Therefore the previous `poly.Evaluate(e2Decoded)` may be applying an interval change-of-basis to a value that is already in the homomorphic evaluator's Chebyshev-variable domain.

This task must determine the exact plaintext meaning of `e2Decoded`, establish one authoritative direct oracle, and then re-evaluate the pre/post-final-Rescale semantic status using that oracle.

Diagnostic only. LogN13 only.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`c3e0500e872eed33fa16b5426eda47bf0bc96675`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Relevant pinned source semantics:

- `utils/bignum/polynomial.go`
- `circuits/common/polynomial/polynomial.go`
- `circuits/common/polynomial/power_basis.go`
- `circuits/ckks/polynomial/polynomial_evaluator.go`
- `circuits/ckks/mod1/mod1_evaluator.go`
- `circuits/ckks/mod1/fast.go`

Do not modify Secondary production code.

Do not full-fetch any large `results/*.json` artifact.

---

# Scope lock

Do not:

- run LogN16;
- benchmark;
- run Gate 4/5 production experiments;
- start EXP-003;
- change polynomial coefficients, Mod1 parameters, PS schedule, generated powers, internal candidate scale, or semantic threshold;
- modify production Lattigo;
- resume target-scale promotion in this task;
- assume either the PS-root oracle or `Polynomial.Evaluate` oracle is correct before the basis-domain comparison is completed.

The only questions are:

1. What mathematical variable does `e2Decoded` represent at the exact input to the polynomial evaluator?
2. Which plaintext oracle reproduces the polynomial evaluator's intended Chebyshev semantics?
3. Under that corrected oracle, are the pre-final and post-final Fast Rescale states actually within `1e-2`?

---

# Source-backed formulas

For `Mod1Poly` with Chebyshev basis and interval `[A,B]`, record:

`scalar = 2 / (B - A)`

`constant = (-B - A) / (B - A)`

and the library `Polynomial.Evaluate(x)` semantics:

`z = scalar*x + constant`

`P(x) = sum_k c_k T_k(z)`.

Also record the actual Fast/Standard Mod1 preprocessing applied immediately before polynomial evaluation, including:

- scale reinterpretation;
- offset added to ciphertext;
- `IntervalShrinkFactor`;
- resulting plaintext-domain relation to the polynomial's Chebyshev variable.

Do not infer this from comments alone: numerically validate the relation on the exact canonical input.

---

# Canonical probe

Use only candidate internal scale `2^91` initially.

Reproduce the exact canonical path through the same point used by the completed global-semantics diagnostic.

Require:

- generated powers pass historical checks;
- pre-final root q0/q1 hashes match the authoritative prior root;
- PS-root source error remains in the observed `~2.6e-7` regime.

If not, stop with:

`CHEBYSHEV_ORACLE_PRECONDITION_MISMATCH`.

Do not run `2^86` in this task unless specifically required to resolve an oracle ambiguity.

---

# Part 1 — identify the variable domain

Let `z = e2Decoded` be the decoded value passed as power-basis `X` to the homomorphic polynomial evaluator.

For the exact same slots, construct the inverse interval-map candidate:

`x = (z - constant) / scalar`.

Record compact ranges for both `z` and `x`:

- min/max real component;
- min/max imaginary component;
- maximum absolute value;
- deterministic hash.

Then validate generated powers against explicit recurrence in `z`:

- `T0(z)=1`;
- `T1(z)=z`;
- `Tn(z)=2*T_a(z)*T_b(z)-T_|a-b|(z)` following the same `SplitDegree` structure.

If ciphertext `T1/T2/...` match recurrence in `z`, that proves the homomorphic power basis is using `z` directly and not applying a hidden second interval mapping.

---

# Part 2 — construct four plaintext oracles

For every slot construct these independent values.

## O1 — historical direct oracle

`O1 = Mod1Poly.Evaluate(z)`

This intentionally preserves the previous diagnostic behavior and therefore includes `Polynomial.Evaluate`'s automatic Chebyshev `ChangeOfBasis`.

## O2 — raw Chebyshev-variable oracle

Evaluate the original `Mod1Poly.Coeffs` directly at `z` with an explicit high-precision Chebyshev recurrence:

`O2 = sum_k c_k*T_k(z)`

Do **not** call `Polynomial.Evaluate` for O2.

Do **not** apply `ChangeOfBasis`.

## O3 — inverse-map then library Evaluate

Construct:

`x = (z - constant)/scalar`

then:

`O3 = Mod1Poly.Evaluate(x)`.

Because `Polynomial.Evaluate` maps `x -> z` internally, O3 should agree with O2 if the domain interpretation is correct.

## O4 — PS decomposition plaintext root

Reconstruct the PS root from the exact `PatersonStockmeyerPolynomial` decomposition using explicit source coefficients and explicit Chebyshev powers in `z`, independent of ciphertext-decoded child values.

This is the corrected/retained form of the previous PS-root plaintext construction.

---

# Part 3 — oracle agreement matrix

Compute pairwise max component errors:

- `O1 vs O2`;
- `O1 vs O3`;
- `O1 vs O4`;
- `O2 vs O3`;
- `O2 vs O4`;
- `O3 vs O4`.

Use a much tighter plaintext-oracle agreement threshold than the CKKS semantic threshold. Prefer high-precision construction and require agreement at or below `1e-10` where float conversion permits; if numerical precision requires a different bound, record and justify it explicitly.

Expected decisive case:

- O2/O3/O4 agree tightly;
- O1 differs by approximately the historical `0.011444...`.

If observed, classify:

`plaintext_oracle_double_change_of_basis_confirmed`.

This means the prior direct oracle was invalid for `z=e2Decoded`.

Alternative classifications:

- `plaintext_ps_root_construction_mismatch` — O2/O3 agree, O4 does not;
- `plaintext_inverse_mapping_mismatch` — O2/O4 agree, O3 does not;
- `plaintext_chebyshev_variable_assumption_mismatch` — generated powers do not correspond to recurrence in `z`;
- `plaintext_oracle_unresolved_disagreement` — no supported equivalence class emerges.

Do not continue to Fast arithmetic conclusions unless an authoritative plaintext oracle is established.

---

# Part 4 — authoritative oracle selection

If O2/O3/O4 agree tightly, designate O2 as the authoritative direct oracle for the already-preprocessed Chebyshev variable `z` because it is explicit and does not hide an interval transform.

Record:

`AUTHORITATIVE_POLY_ORACLE = raw_chebyshev_on_preprocessed_z`

Also explicitly mark the old construction:

`Mod1Poly.Evaluate(e2Decoded)`

as invalid for this checkpoint because it applies the interval map again.

Do not globally change unrelated uses of `Polynomial.Evaluate`; this conclusion is checkpoint/domain-specific.

---

# Part 5 — re-evaluate the historical failure boundary

Only after establishing an authoritative oracle, evaluate the exact ciphertext states against it.

## C0 — pre-final-Rescale root

Use the exact authoritative q0/q1 root state that previously hash-matched.

Record:

- corrected whole-polynomial max component error;
- old O1-based error for comparison;
- q0/q1 hashes;
- Level/Degree/Scale.

## C1 — post-final Fast Rescale

Run the same actual final Fast Rescale used in the prior diagnostics.

Record:

- corrected whole-polynomial max component error;
- old O1-based error for comparison;
- output Level/Degree/Scale;
- q0/q1 hashes.

Interpretation:

### Case A — both C0 and C1 pass `1e-2`

`FIRST_SUPPORTED_CAUSE = historical_finalization_failure_was_plaintext_oracle_basis_error`

The previous `~0.0114439` failure was caused by the invalid direct oracle, not by the PS root or final Fast Rescale.

Stop here. Do not perform target-scale promotion in this task.

### Case B — C0 passes but C1 fails

`FIRST_SUPPORTED_CAUSE = corrected_oracle_reopens_final_fast_rescale_boundary`

This re-establishes final Rescale as a supported failure boundary. Stop; a later task may resume the internal Rescale diagnostic.

### Case C — C0 already fails

`FIRST_SUPPORTED_CAUSE = corrected_oracle_pre_final_semantic_failure`

The oracle mismatch was real but not sufficient to explain the pre-final failure. Stop; do not inspect final Rescale internals.

---

# Artifact discipline

Create compact artifacts only:

- `results/FIX-001-P3-DIAG-CHEBYSHEV-ORACLE-BASIS-logN13.json`
- `results/FIX-001-P3-DIAG-CHEBYSHEV-ORACLE-BASIS-logN13-summary.json`

The summary must stay human-reviewable.

Store only:

- provenance;
- `[A,B]`, scalar, constant, IntervalShrinkFactor, Mod1 offset;
- compact `z` / inverse-mapped `x` ranges and hashes;
- generated-power domain validation;
- O1/O2/O3/O4 pairwise error matrix;
- authoritative-oracle decision;
- C0/C1 corrected and historical errors;
- first supported cause;
- tests/clean-state evidence.

Do not store full slot vectors, coefficient arrays, operation traces, or repeated index lists.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- focused diagnostic tests pass;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Secondary changes;
- Primary artifacts are compact, committed, and pushed normally;
- both worktrees are clean;
- no LogN16;
- no benchmark;
- no Gate 4/5 production run;
- no EXP-003;
- no target-scale promotion.

The deliverable is a source-backed resolution of the Chebyshev plaintext-oracle domain, followed by one corrected pre/post-final-Rescale semantic check.