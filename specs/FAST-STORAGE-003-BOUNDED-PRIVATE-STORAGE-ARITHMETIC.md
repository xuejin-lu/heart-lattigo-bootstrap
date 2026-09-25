# FAST-STORAGE-003 — Bounded Private-Storage Arithmetic Foundation

## Status

Executable Codex task.

## Task class

`I — Implementation`

All mathematical/architectural decisions required by this task are already frozen by GPT Web in Secondary `docs/FAST_CKKS_SPEC.md`, especially Sections 4.2, 4.4, 4.5, and 4.6.

Codex owns implementation and coding-quality review only. If current source evidence appears to require changing a formula, capacity invariant, logical CKKS semantics, or the Q/F architecture, stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`
- orchestration/spec only; no experiment-code change is expected for this task.

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- implementation target.

Accepted prerequisite:
- FAST-STORAGE-002 implementation `3b57a52b311397e0e1cf8298027782eec20d4ffc`

Current accepted container/conversion files include:
- `schemes/ckks/fast/storage_ciphertext.go`
- `schemes/ckks/fast/storage_basis.go`
- `schemes/ckks/fast/storage_conversion.go`

## Purpose

Turn the accepted private-F container into a small, mathematically bounded arithmetic substrate.

This task must add only:

1. proven per-component coefficient-infinity bounds to `FastCiphertext`;
2. exact private-storage width planning;
3. exact storage expansion to a wider F basis;
4. standalone private-storage Add/Sub;
5. standalone raw private-storage ciphertext Mul.

This is still **not** production Bootstrap integration.

Do not wire these operations into the existing Fast evaluator, Bootstrap, Rescale, KeySwitch, Relinearize, Rotate, or application path.

---

# 1. Frozen state invariant

For ciphertext degree `d`, a Fast ciphertext carries a bound vector

[
mathbf B=(B_0,ldots,B_d)
]

such that for every component:

[
|X_j|_infty le B_j.
]

For active private-storage width `w`, let

[
S_w=prod_{i=0}^{w-1} f_i.
]

Every valid authoritative component must satisfy the strict uniqueness condition:

[
oxed{2B_j<S_w}
]

for every `j`.

Bounds are coefficient-domain semantic facts. They do not change merely because a polynomial is transformed between coefficient and NTT domains.

Bounds must be non-negative exact integers. No floating-point approximation is allowed for capacity decisions.

Implementation representation is a coding choice:
- `big.Int` metadata is acceptable;
- checked fixed-width metadata is acceptable if it cannot silently overflow.

Any overflow/representation failure must become an explicit error, never a wrapped bound.

---

# 2. Bound lifecycle in FastCiphertext

Extend `FastCiphertext` with private per-component bound state.

Exact field/type names are implementation choices, but the following behavior is mandatory.

## Constructor

`NewFastCiphertext(...)` creates mathematical zero components, therefore:

[
B_j=0
]

for every allocated component.

## Copy

`CopyNew()` must deep-copy bound metadata. No aliased mutable integer state.

## ResizeDegree

When increasing degree:
- existing component bounds are preserved;
- new zero-filled components receive bound zero.

When decreasing degree:
- truncated component bounds are removed consistently with truncated value components.

## Logical metadata changes

`SetLogicalLevel` and `SetScale` do not alter bounds.

## Read access

Provide a minimal read-only/copying accessor conceptually equivalent to:

`ComponentBounds()`

It must not let callers mutate internal bound state through aliasing.

Do not add a broad public arbitrary-bound setter merely for convenience.

---

# 3. Bound provenance at the accepted Level-0 import

Update `ImportLevel0` so imported Fast ciphertexts receive a proven bound vector.

For each component, after obtaining the canonical centered q0 lift

[
C=Center_{q_0}(c),
]

compute the exact observed maximum:

[
B_j = max_k |C_{j,k}|.
]

This scan is already on the import boundary and is acceptable.

The bound must describe the represented coefficient polynomial even if the requested target storage domain is NTT.

The existing exact import semantics remain unchanged:

[
X_F=Center_{q_0}(c).
]

Do not change logical q0 semantics, Scale, or the accepted conversion architecture.

---

# 4. Required storage width

For a bound vector `B`, define:

[
w_{req}(mathbf B)
=
minleft{
win{1,2,3}:
2B_j<S_w orall j
ight}.
]

Implement a small exact planner/helper.

If no width in `{1,2,3}` satisfies the condition, return an explicit capacity error.

Strict inequality is mandatory.

Useful existing source:
- `fastStorageBasis.products`
- `centeredCapacity`
- fixed storage primes.

Do not infer required width from logical CKKS Level.

---

# 5. Exact storage expansion

Provide an explicit representation-only operation conceptually equivalent to:

[
A_w ightarrow A_{w'}
qquad
w' > w.
]

Required properties:

- only widths 1, 2, 3 are valid;
- contraction is not part of this task;
- logical Level unchanged;
- Scale unchanged;
- degree unchanged;
- bound vector unchanged;
- represented lifted integer `X` unchanged exactly;
- existing input state must already satisfy `2B_j<S_w`;
- target result must satisfy `2B_j<S_{w'}`;
- Montgomery Fast storage may be rejected explicitly;
- both coefficient and NTT storage domains must be handled correctly.

For NTT-domain expansion, do not reinterpret NTT coordinates across moduli.

The semantic path is:

[
F_{source}	ext{-NTT}
ightarrow
F_{source}	ext{-coeff}
ightarrow
	ext{centered integer }X
ightarrow
Xmod f_{new}
ightarrow
F_{new}	ext{-NTT}.
]

Existing authoritative rows may be copied when safe, but every newly populated row must represent the same coefficient-domain lift.

No hidden logical q_i row may be introduced.

---

# 6. Common arithmetic compatibility contract

The standalone storage arithmetic in this task must require:

- non-nil inputs;
- matching CKKS parameter identity/value;
- matching ring degree N;
- Standard ring type;
- non-Montgomery Fast storage;
- valid storage rows and valid bound invariant;
- matching representation domain where the operation requires it.

Do not call ordinary logical-`RingQ` arithmetic on private F rows.

Do not mutate either input.

On any validation/capacity failure:
- return an error;
- leave all inputs byte-for-byte/metadata-equivalent to their pre-call state;
- do not return a partially valid output.

This is the transactional failure rule.

---

# 7. Add/Sub contract

For inputs `X` and `Y`, missing higher-degree components are mathematical zero.

Output degree:

[
d_Z=max(d_X,d_Y).
]

Output logical level:

[
ell_Z=min(ell_X,ell_Y).
]

Initial arithmetic contract requires equal CKKS Scale:

[
Delta_X=Delta_Y.
]

Otherwise reject.

Require matching coefficient/NTT domain. Add/Sub may operate in either domain as long as both inputs match.

Per-component bounds:

[
oxed{
B_{Z,j}le B_{X,j}+B_{Y,j}
}
]

with missing input component bound interpreted as zero.

Plan the operation width as:

[
oxed{
w_{op}
=
maxleft(
w_X,,
w_Y,,
w_{req}(mathbf B_Z)
ight)
}
]

This deliberately forbids implicit contraction.

If an input has smaller width than `w_op`, expand a copy or otherwise perform an equivalent non-mutating exact expansion before arithmetic.

The output storage width is `w_op`.

Residue arithmetic must be performed modulo the corresponding `f_i`.

Output Scale equals the common input Scale.

Preserve compatible plaintext metadata. If metadata required to interpret the ciphertext is incompatible between operands, reject explicitly rather than guessing.

---

# 8. Raw ciphertext multiplication contract

This task implements **raw multiplication only**.

No Relinearize.
No KeySwitch.
No Rescale.
No Scale normalization.

Require both operands in Fast NTT domain for this initial implementation.

Output:

[
d_Z=d_X+d_Y,
]

[
ell_Z=min(ell_X,ell_Y),
]

[
Delta_Z=Delta_XDelta_Y.
]

For each output ciphertext component:

[
Z_k
=
sum_{i=max(0,k-d_Y)}^{min(d_X,k)}
X_istar Y_{k-i}
]

in

[
R=mathbb Z[X]/(X^N+1).
]

The mandatory conservative bound is:

[
oxed{
B_{Z,k}
le
N
sum_i
B_{X,i}B_{Y,k-i}
}
]

computed with exact integer arithmetic before executing the polynomial multiplication.

For degree-one by degree-one:

[
B_{Z,0}le NB_{X,0}B_{Y,0},
]

[
B_{Z,1}le N(B_{X,0}B_{Y,1}+B_{X,1}B_{Y,0}),
]

[
B_{Z,2}le NB_{X,1}B_{Y,1}.
]

Use the same width rule:

[
w_{op}
=
maxleft(
w_X,,
w_Y,,
w_{req}(mathbf B_Z)
ight).
]

If no private width can prove unique reconstruction, fail **before** returning/mutating arithmetic state.

At each active private prime, implement the ordinary ciphertext-component convolution using that storage subring's NTT-domain modular multiplication/addition primitives.

Do not fold Relinearization into this result.

---

# 9. Scope/API placement

Prefer a small CKKS-Fast-local surface, for example:

- extending `storage_ciphertext.go` for bound metadata/lifecycle;
- one small arithmetic/width file such as `storage_arithmetic.go`;
- corresponding focused tests.

Exact file split is a coding choice.

Do not replace or rewrite the existing historical `fast_add.go`, `fast_mul.go`, or production evaluator path in this task.

The new foundation may coexist unused by production until a later integration task.

---

# 10. Required tests

## T1 — bound lifecycle

Verify:
- zero constructor bounds;
- `CopyNew` deep-copy behavior;
- `ResizeDegree` grow/shrink behavior;
- logical Level/Scale metadata changes do not alter bounds.

## T2 — exact import bounds

Use Level-0 inputs with positive, negative, zero, and boundary-near centered q0 representatives.

For each ciphertext component verify:

[
B_j=max_k|Center_{q_0}(c_{j,k})|.
]

Repeat for coefficient input and NTT input.

## T3 — required-width thresholds

Using synthetic bound vectors around the exact centered capacities, prove:
- width 1 selection;
- width 2 selection;
- width 3 selection;
- strict-boundary rejection;
- >width3 capacity rejection.

No approximate bit-length-only decision is sufficient.

## T4 — exact expansion

Test at least:
- 1 -> 2;
- 1 -> 3;
- 2 -> 3;

in coefficient domain and NTT domain.

Reconstruct before/after and prove exact lifted-integer equality for every tested component/coefficient.

Verify logical Level, Scale, degree, bounds, and domain state are preserved.

## T5 — Add/Sub exactness

Use small synthetic lifts where expected integer results are easy to compute.

Cover:
- same width;
- mixed widths requiring expansion;
- degree mismatch with missing components interpreted as zero;
- coefficient domain;
- NTT domain.

Decode output and compare exact integer Add/Sub results.

Verify output bound vector and `w_op`.

## T6 — Add/Sub logical metadata

Verify:
- output logical Level is `min`;
- equal Scale is preserved;
- unequal Scale rejects;
- input ciphertexts remain unchanged.

## T7 — raw Mul exactness

Use small degree-1 NTT-domain ciphertext polynomials.

Compute an independent coefficient-domain negacyclic integer reference and verify decoded output equals it exactly.

Verify degree 2 output and Scale product.

## T8 — multiplication bound formula

For degree-1 x degree-1, verify the three component bounds exactly match the conservative formulas above.

Also include at least one mixed/nonuniform bound-vector case so the implementation cannot accidentally use one global scalar bound.

## T9 — multiplication width planning

Construct cases where the proven output requires:
- width 1;
- expansion to width 2;
- expansion to width 3.

Verify no implicit contraction when an input already has a wider basis.

## T10 — transactional capacity failure

Construct Add/Sub or Mul metadata/data whose proven output bound does not fit width 3.

Verify:
- explicit capacity error;
- both inputs unchanged;
- no partially valid output is exposed.

## T11 — domain/compatibility rejection

Verify explicit rejection for:
- Montgomery Fast storage;
- Mul with non-NTT operands;
- incompatible parameters/N;
- incompatible required metadata;
- malformed storage/bound invariant.

## T12 — logical export congruence after arithmetic

For representative successful Add/Sub/Mul outputs, export through the accepted Fast -> LogicalQ boundary and verify logical residues agree with the same integer reference modulo each active logical `q_i`.

This confirms arithmetic preserves C1 while operating only on F rows.

## T13 — regressions

Run:
- new storage arithmetic tests;
- all existing `schemes/ckks/fast` tests;
- relevant existing CKKS/bootstrapping tests sufficient to prove this unused foundation did not change production behavior;
- `git diff --check`;
- broader `go test ./...` when practical, classifying any pre-existing unrelated debt precisely.

---

# 11. Performance rule

This is still a correctness foundation, not a speed gate.

However:

- do not use per-coefficient `math/big` inside Add/Sub/Mul hot residue arithmetic;
- exact `big.Int` for a tiny per-component bound planner is acceptable;
- exact expansion may reconstruct coefficients because expansion is an explicit representation transition;
- avoid unnecessary width-3 work when `w_op` proves width 1 or 2 is enough.

No benchmark acceptance threshold is introduced in this task.

---

# 12. Prohibitions

Do not:

- wire this foundation into production Bootstrap;
- modify Primary benchmark semantics;
- change Logical Q or frontend CKKS parameters;
- use F primes as logical CKKS moduli;
- infer logical Level from storage width;
- implement storage contraction;
- implement Rescale;
- implement ModUp changes;
- implement KeySwitch;
- implement Relinearize;
- implement Rotate/Automorphism changes;
- modify key generation;
- rewrite the existing production Fast evaluator;
- change mathematical bounds or relax strict capacity checks;
- add a hidden Standard/full-RNS fallback.

---

# 13. Success classification

Candidate classification:

`FAST_STORAGE_003_BOUNDED_ARITHMETIC_ACCEPTED_CANDIDATE`

Successful Codex handoff:

`READY_FOR_WEB_REVIEW`

## Completion report

Report at minimum:

- exact Secondary commit;
- changed files;
- bound representation and lifecycle;
- exact import-bound evidence;
- width planner threshold evidence;
- expansion evidence;
- Add/Sub exactness and width evidence;
- raw Mul exactness and bound evidence;
- transactional capacity-failure evidence;
- logical export congruence evidence;
- regression tests;
- confirmation that production Bootstrap/evaluator paths and Primary experiment semantics were not modified.
