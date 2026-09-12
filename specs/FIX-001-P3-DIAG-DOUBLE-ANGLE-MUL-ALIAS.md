# FIX-001-P3-DIAG-DOUBLE-ANGLE-MUL-ALIAS — Isolate the first DoubleAngle square failure

## Purpose

The completed LogN13 DoubleAngle diagnostic established:

- canonical restored P4 / D0 is semantically correct (`~2.6648e-7` max component error);
- D0 coefficient-domain q0/q1 capacity is valid;
- the first failing operation is Round 0 `D1.0 FastCKKS.MulRelin`;
- D1 max component error is `0.6075795944054425` at threshold `1e-2`;
- D1's **post-operation** centered q0/q1 representative has ratio `0.5568217129919162` and outside count `0`.

Independent source review shows that the current Fast architecture is explicitly zero-secret for these diagnostics:

- Primary diagnostic decoding uses `zeroSecret(params)`;
- Fast key material is generated from zeroed secret keys;
- `FastTruncateDegree2To1` documents the Fast contract `s = 0`;
- Fast `MulRelin` for degree-one operands computes `c0`, `c1`, omits `c2`, and relies on that zero-secret contract.

At the canonical P4 state, the maintained `c1` is expected to be exactly zero. Therefore the current evidence does **not** yet support blaming relinearization or omitted `c2`.

A stronger hypothesis is q0/q1 multiplication alias:

- P4 coefficient-domain `c0` maximum is about `8.9867e17`;
- the first square is performed only modulo q0 and q1;
- the D1 capacity ratio is measured **after modular multiplication and centered reconstruction**, so a wrapped product can still report ratio `< 1`;
- therefore D1 post-operation capacity is not a proof that the exact pre-modular negacyclic product was uniquely representable in `[-Q01/2,Q01/2)`.

This task must distinguish:

1. q0/q1 modular multiplication alias caused by the physically restored target Scale;
2. a genuine Fast multiplication primitive defect unrelated to capacity;
3. a relinearization/truncation defect.

Diagnostic only. LogN13 only. Do not modify Secondary production code.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`7dd1ed3086e94cbb94c6e01d7a648999d60ace1f`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Authoritative polynomial oracle:

`raw_chebyshev_on_preprocessed_z`

Semantic threshold:

`1e-2`

Canonical internal polynomial candidate:

`2^91`

Known target promotion:

`M = 536870912 = 2^29`

Do not modify Secondary production code in this task.

---

# Scope lock

Do not:

- change `Mul`, `MulRelin`, `Relinearize`, Rescale, key generation, or target-scale production code;
- run lower PS candidates;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- continue to DoubleAngle doubling/offset/Rescale after the failing square;
- redesign the DoubleAngle scale schedule in this task;
- call Standard/full-RNS arithmetic on stale or dormant Fast limbs;
- use direct CRT of NTT rows as coefficient-domain evidence;
- infer exact multiplication safety from a post-wrap centered q0/q1 representative.

The only question is:

> Does the Round-0 square fail because the exact coefficient-domain negacyclic product cannot be represented uniquely by q0/q1 at the physically restored target Scale, or does Fast multiplication/relinearization fail even when alias is excluded?

---

# Mandatory reproduction

Reproduce both canonical polynomial states from the already validated `2^91` path:

## State L — low/compressed state

The exact post-final-Fast-Rescale polynomial result **before** target-scale integer promotion.

Expected approximately:

- Scale `S ≈ 2.147483648e9`;
- semantic error `~2.6648e-7`;
- authoritative plaintext `Y0`.

## State H — high/restored state

The exact P4 state after:

- integer promotion by `M=2^29`;
- matched Scale update;
- exact target metadata normalization.

Expected approximately:

- Scale `T ≈ 1.1529215046069e18`;
- semantic error `~2.6648e-7`;
- same authoritative plaintext `Y0`;
- Level 7, Degree 1;
- known P4 q0/q1 hashes from the prior diagnostic.

Require L and H to decode to the same `Y0` within the existing threshold and reproduce prior hashes/scales where pinned.

If not:

`DOUBLE_ANGLE_MUL_ALIAS_PRECONDITION_MISMATCH`.

---

# A0 — establish the actual zero-secret multiplication contract

For both L and H, using the corrected coefficient-domain view:

1. verify every maintained q0/q1 coefficient of `c1` is exactly zero;
2. verify NTT/Montgomery maintained `c1` rows are also exactly zero;
3. record zero counts / nonzero counts compactly;
4. record that Primary semantic decoding uses the zero-secret diagnostic path;
5. record source provenance for Fast `MulRelin` / `FastTruncateDegree2To1` zero-secret semantics.

If H `c1` is not exactly zero, stop with:

`double_angle_mul_zero_secret_precondition_mismatch`.

Do not assume `c2` is irrelevant unless this check passes.

---

# A1 — split multiplication from relinearization at State H

Starting from independent copies of H:

## H-MUL

Execute:

`FastCKKS.Mul(H, H, degree2Out)`

Do **not** relinearize.

Record:

- Level / Degree / Scale;
- q0/q1 hashes for c0/c1/c2;
- whether c1 and c2 are exactly zero;
- zero-secret decoded semantic error against `Y0^2`;
- corrected coefficient-domain post-modular centered representatives for all components, clearly labelled as **post-modular only**.

Because the Fast semantic secret is zero, degree-two decoding is expected to depend only on c0. If H-MUL already fails against `Y0^2`, the failure exists before relinearization.

## H-REL

From an independent H-MUL product execute:

`FastCKKS.Relinearize(product, relinOut)`.

Also independently execute direct:

`FastCKKS.MulRelin(H, H, directRelinOut)`.

Require:

- H-REL and direct MulRelin q0/q1 c0/c1 rows match exactly;
- semantic outputs agree;
- with c1/c2 zero, relinearization does not change c0.

Classify a relinearization problem only if H-MUL is semantically correct but H-REL/direct MulRelin is not.

---

# A2 — low-scale control square

Starting from independent copies of L, execute the same operations:

1. `FastCKKS.Mul(L,L)`;
2. zero-secret decode against `Y0^2`;
3. `Relinearize` and direct `MulRelin` comparison.

Before multiplication, construct the authoritative coefficient-domain c0 vector `x` and compute:

`B = max_i |x_i|`.

Use the rigorous negacyclic-product absolute bound:

`productBound = N * B^2`.

Record exact big-integer values for:

- `N`;
- `B`;
- `N*B^2`;
- `Q01/2`;
- ratio `(N*B^2)/(Q01/2)`.

If:

`N*B^2 < Q01/2`

then q0/q1 uniqueness is mathematically guaranteed for every exact negacyclic product coefficient of L.

Require L-MUL semantic error <= `1e-2` if this bound passes.

This is the key control: the same Fast pointwise q0/q1 multiplication machinery should work when alias is rigorously impossible.

---

# A3 — fresh-q2 alias witness for State H

Do not use H's dormant q2 row.

Construct a **fresh independent q2 product oracle** from the authoritative coefficient-domain c0 of H:

1. obtain exact centered coefficient-domain c0 values from corrected q0/q1 reconstruction;
2. allocate fresh q2-only diagnostic storage;
3. reduce each exact coefficient modulo q2 into that fresh storage;
4. apply the correct q2 NTT representation required for ring multiplication;
5. square modulo q2 using ring operations;
6. inverse NTT back to a fresh coefficient-domain q2 product row.

Separately obtain the H Fast q0/q1 square product c0 and convert it correctly to a centered coefficient-domain q0/q1 candidate `r_j` for every coefficient.

For every coefficient j compare:

`r_j mod q2`

with:

`fresh_q2_product[j]`.

Rationale:

- the Fast q0/q1 product candidate `r_j` is congruent to the exact integer product modulo `Q01`;
- fresh q2 independently gives the exact product modulo q2;
- q2 is coprime to q0 and q1;
- if `r_j mod q2 != fresh_q2_product[j]`, then `r_j` cannot equal the exact integer product coefficient;
- since `r_j` is the unique centered representative modulo Q01, such a mismatch proves that the exact product lies outside the centered Q01 uniqueness interval and has aliased modulo q0/q1.

Record compactly:

- q2;
- mismatch count;
- first mismatch index;
- first centered q01 candidate;
- candidate mod q2;
- fresh q2 product residue;
- optional hash of the mismatch bitmap/index set, but no full arrays.

### Required interpretation

If mismatch count > 0:

`H_Q01_ALIAS_WITNESS = true`.

If mismatch count == 0, do **not** automatically claim no alias; continue to A4 because equality modulo one auxiliary prime is not a universal no-alias proof.

---

# A4 — high-scale product bound / no-alias disposition

For H compute the same conservative bound:

`N * B_H^2`.

If this bound is already `< Q01/2`, then alias is rigorously impossible. In that case a failing H-MUL must be a primitive/representation bug rather than capacity alias.

If the bound is `>= Q01/2`, it is only inconclusive—not proof of alias.

Use A3 fresh-q2 mismatch as the positive alias witness when available.

If A3 has no mismatch and the conservative bound is inconclusive, classify the capacity question as unresolved rather than guessing.

---

# A5 — distinguish physical scale restoration from multiplication machinery

Compare L and H:

- both represent the same `Y0`;
- H coefficients should be the physically promoted image of L under `M=2^29` modulo q0/q1;
- H Scale is correspondingly promoted so plaintext semantics are preserved.

Record:

- L and H max coefficient magnitude;
- their ratio;
- expected promotion factor M;
- low-scale bound result;
- H fresh-q2 alias witness;
- L-MUL vs H-MUL semantic error.

The decisive expected pattern is:

- L square: rigorous no-alias bound passes and semantic square passes;
- H square: semantic square fails near the prior D1 failure;
- H fresh-q2 witness has mismatches;
- H `Mul` already fails before `Relinearize`;
- `Relinearize(Mul(H,H))` equals direct `MulRelin(H,H)`.

If all hold, this proves the first DoubleAngle failure is not a relinearization bug; it is q0/q1 multiplication alias introduced by carrying the Standard-style physical target Scale into the q0/q1-only square.

---

# Required classification

Choose exactly one:

## Case A — alias confirmed

`FIRST_SUPPORTED_CAUSE = double_angle_mul_q01_alias_confirmed`

Requirements:

- H-MUL fails before relinearization;
- fresh-q2 witness has mismatch count > 0;
- H-REL equals direct MulRelin;
- L no-alias bound passes;
- L square is semantically correct.

Also record:

`ROOT_MECHANISM = physical_target_scale_promotion_exceeds_q01_square_uniqueness_capacity`

This result authorizes a later **design task** for a Fast-specific DoubleAngle internal scale schedule. It does not authorize implementing that design in this task.

## Case B — primitive failure without alias

`FIRST_SUPPORTED_CAUSE = double_angle_mul_primitive_failure_without_alias`

Use only if alias is rigorously excluded for H but H-MUL still fails.

## Case C — relinearization failure

`FIRST_SUPPORTED_CAUSE = double_angle_relinearization_truncation_failure`

Use only if H-MUL passes against `Y0^2` and H-REL/direct MulRelin fails.

## Case D — low-scale control also fails

`FIRST_SUPPORTED_CAUSE = double_angle_mul_primitive_failure_precedes_target_scale_alias_question`

Use if the rigorous low-scale no-alias bound passes but L-MUL itself fails.

## Case E — alias unresolved

`FIRST_SUPPORTED_CAUSE = double_angle_mul_capacity_unresolved`

Use if H-MUL fails, H no-alias bound is inconclusive, and the fresh-q2 witness gives no mismatch.

## Case F — precondition mismatch

`FIRST_SUPPORTED_CAUSE = double_angle_mul_alias_precondition_mismatch`

---

# Artifact discipline

Create compact artifacts only:

- `results/FIX-001-P3-DIAG-DOUBLE-ANGLE-MUL-ALIAS-logN13.json`
- `results/FIX-001-P3-DIAG-DOUBLE-ANGLE-MUL-ALIAS-logN13-summary.json`

Summary must remain human-reviewable.

Store only:

- provenance;
- L/H scales, hashes, semantics;
- c1/c2 zero checks;
- Mul vs MulRelin vs Relinearize compact semantic evidence;
- L/H `B`, `N*B^2`, Q01/2 and bound disposition;
- fresh-q2 witness metadata and mismatch count/first mismatch;
- root mechanism / classification;
- tests and clean-state evidence.

Do not serialize full coefficient vectors, slot vectors, mismatch-index arrays, operation traces, or large hash collections.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- focused Mul-alias diagnostic tests pass;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no Secondary production changes;
- compact artifacts are committed and pushed normally;
- both worktrees clean;
- no lower candidate sweep;
- no DoubleAngle operation after the first square;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003.

The deliverable is a source- and representation-correct answer to whether the first LogN13 DoubleAngle square fails from q0/q1 modular alias caused by physical target-scale restoration, rather than from relinearization.