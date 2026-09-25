# FAST-STORAGE-004 — Logical-Q Rescale over Private-F Storage

## Status

Executable Codex task.

## Task class

`I — Implementation`

The mathematical and architectural decisions for this task are already frozen by GPT Web in Secondary `docs/FAST_CKKS_SPEC.md`, especially Sections 4.2, 4.5, 4.8, 4.11, 4.14, and 4.16.

Codex owns implementation and coding-quality review only.

If current source evidence appears to require changing:
- the Rescale divisor;
- the rounding rule;
- the post-Rescale bound formula;
- Logical-Q/F separation;
- Level/Scale semantics;
- or the strict centered-capacity invariant,

stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`
- orchestration/spec only; no experiment-code changes are expected.

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- implementation target.

Accepted prerequisite:
- FAST-STORAGE-003 implementation `b8305a7e3d4ff15591a2249e97a54ad0b3311dde`

Current private-F foundation:
- explicit logical Level independent of storage width;
- fixed private storage basis F;
- per-component proven coefficient bounds;
- exact width planner;
- exact storage expansion;
- standalone Add/Sub/raw Mul.

## Purpose

Implement the first operation that changes the authoritative lifted integer while preserving the Logical-Q/F separation:

[
X longmapsto Y=operatorname{Round}(X/q_ell).
]

This task implements **one-step standalone Fast-storage Rescale only**.

It must prove that:
- the physical rounded division uses the frontend logical modulus `q_ell`;
- private storage primes `f_i` are never used as CKKS Rescale divisors;
- logical Level and Scale evolve exactly according to CKKS semantics;
- the authoritative private-F lift is updated exactly;
- the bound vector is updated by the frozen theorem;
- storage width remains independent of logical Level.

This is still **not** production Bootstrap/Evaluator integration.

---

# 1. Frozen one-step Rescale transition

For a valid `FastCiphertext` at logical Level

[
ellge 1,
]

let the authoritative component lifts be `X_j`, with

[
|X_j|_inftyle B_j.
]

Let

[
q=q_ell
]

be the **logical CKKS modulus from the public parameter chain at the current logical Level**.

The output must represent

[
oxed{
Y_j=operatorname{Round}(X_j/q)
}
]

coefficient-wise.

The logical metadata transition is exactly

[
oxed{
ell'=ell-1
}
]

and

[
oxed{
Delta'=Delta/q.
}
]

Degree is unchanged.

The active private storage width is unchanged in this task:

[
oxed{
w'=w.
}
]

No storage contraction is allowed here.

---

# 2. Mandatory logical divisor rule

The divisor is:

[
q_ell=	ext{ct.params.Q()[ct.LogicalLevel()]}.
]

It is never:
- `f0`;
- `f1`;
- `f2`;
- the last active private storage prime;
- or a value inferred from `StorageWidth()`.

The same logical `q_ell` must be used for:

1. coefficient rounded division;
2. Scale division.

This is the central correctness gate.

Do not copy the historical assumption from the old q0/q1/q012 Rescale path that storage-row modulus identity implies logical-divisor identity.

The historical `schemes/ckks/fast/rescale.go` is reference/current-production evidence only. Do not rewire or replace it in this task.

---

# 3. Rounding rule

Every logical `q_i` is an odd CKKS prime.

For an integer coefficient `x`, define nearest-integer division:

[
operatorname{Round}(x/q).
]

Because `q` is odd, exact half-integer ties cannot occur for integer `x`.

An equivalent signed-magnitude rule is:

[
left|operatorname{Round}(x/q)ight|
=
leftlfloor
rac{|x|+(q-1)/2}{q}
ightfloor
]

with the original sign preserved for nonzero quotient.

The implementation must be exact for the full private-storage centered range.

Do not convert coefficients through `float64`.

---

# 4. Frozen output bound

For each component:

[
oxed{
B'_j
=
leftlfloor
rac{B_j+(q-1)/2}{q}
ightfloor
}.
]

Compute this with exact integer arithmetic.

The output bound remains provenance metadata and must satisfy:

[
2B'_j<S_w
]

for the unchanged storage width.

Since Rescale reduces magnitude, a valid input should imply the output fits the same width; nevertheless validate the resulting bound/state rather than relying on an unchecked assumption.

Do not scan output coefficients to replace this conservative theorem with a smaller observed bound.

---

# 5. API shape

Provide a minimal standalone API conceptually equivalent to:

`FastStorageRescale(op *FastCiphertext) (*FastCiphertext, error)`

Requirements:
- non-mutating input;
- one logical level per call;
- returns a new `FastCiphertext`;
- nil on failure;
- no `opOut` aliasing API is required yet;
- no `RescaleTo` loop is required yet.

Exact function/file naming may improve, but keep the surface small and Fast-private.

Prefer a new focused file such as:

- `schemes/ckks/fast/storage_rescale.go`
- `schemes/ckks/fast/storage_rescale_test.go`

Do not modify the public Standard CKKS API.

---

# 6. Input contract

Require:

- non-nil valid `FastCiphertext`;
- logical Level >= 1;
- valid private storage width 1/2/3;
- valid per-component bound invariant;
- positive Scale;
- Standard ring parameters;
- non-Montgomery Fast storage;
- valid private F rows.

Support and preserve both:
- coefficient-domain private-F storage;
- NTT-domain private-F storage.

The output domain must equal the input domain.

Montgomery private-F input may be explicitly rejected.

---

# 7. Exact physical algorithm

For every ciphertext component:

## Coefficient-domain input

For every coefficient:

1. reconstruct the unique centered authoritative private-F lift `X` from the active storage rows;
2. compute exactly:
   [
   Y=operatorname{Round}(X/q_ell);
   ]
3. reduce `Y` modulo every currently active private storage prime `f_i`;
4. write those residues into the output.

## NTT-domain input

The semantic path must be:

[
F	ext{-NTT}
	o
F	ext{-coefficient residues}
	o
X
	o
operatorname{Round}(X/q_ell)
	o
F	ext{-coefficient residues}
	o
F	ext{-NTT}.
]

Do not divide NTT coordinates directly.

Do not copy/reinterpret logical-Q NTT machinery over private-F rows.

For this correctness foundation, an explicit per-component INTT/NTT round trip is acceptable.

---

# 8. Fixed-width arithmetic preference

Current private storage uses at most three ~60-bit primes and already has `storageUint192` / `storageInteger`.

Prefer an exact fixed-width signed-magnitude divide-by-uint64 helper for the per-coefficient hot loop.

Conceptually:

`roundStorageIntegerByUint64(value storageInteger, divisor uint64)`

or equivalent.

Requirements:
- exact across the supported 192-bit magnitude representation;
- no per-coefficient `math/big` allocation in the Rescale hot loop;
- quotient and remainder computed without silent overflow;
- rounding increment exact;
- zero has non-negative canonical sign.

A tiny per-component `big.Int` calculation for bound metadata is acceptable.

Do not import the historical q0/q1-specific bit-size restrictions from `rescale.go`; the new private-F representation already uniquely reconstructs the bounded lift. The logical divisor is simply the current CKKS `q_ell`.

---

# 9. Metadata transition

The output must preserve all compatible ciphertext/plaintext metadata except the fields intentionally changed by Rescale.

Mandatory:

- `LogicalLevel = input.LogicalLevel() - 1`;
- `Scale = input.Scale() / q_ell`;
- degree unchanged;
- `StorageWidth` unchanged;
- `IsNTT` unchanged;
- `IsMontgomery = false`;
- batching/dimension/bit-reversal metadata preserved;
- CKKS parameter object/value unchanged;
- component bounds replaced by the frozen post-Rescale bounds.

Rescale must not infer or change storage width from the new logical Level.

---

# 10. Transactional failure

Any validation failure must:

- return an error;
- return no partially valid output;
- leave the input unchanged.

Examples:
- logical Level 0;
- invalid/malformed bound state;
- malformed storage rows;
- Montgomery storage;
- invalid/non-positive Scale;
- impossible internal arithmetic condition.

---

# 11. Required independent semantic oracle

The implementation must not be validated only by self-consistency of private-F encode/decode.

For a representative valid Fast ciphertext at logical Level `ell >= 1`:

1. export the input through the accepted Fast -> LogicalQ boundary at Level `ell`;
2. run ordinary Standard CKKS one-step Rescale on that logical ciphertext;
3. run the new `FastStorageRescale` on the original Fast ciphertext;
4. export the Fast output to LogicalQ at Level `ell-1`;
5. compare logical residues and Scale/Level semantics.

The theorem in `docs/FAST_CKKS_SPEC.md` guarantees agreement modulo `Q_{ell-1}` even when the Fast authoritative lift differs from the canonical logical representative by a multiple of `Q_ell`.

This oracle is specifically intended to detect accidental division by a private `f_i`.

---

# 12. Required tests

## T1 — simple signed rounded division

Use direct private-F lifts around:
- 0;
- +/-1;
- +/-q;
- +/- (q-1)/2;
- +/- (q+1)/2;
- `k*q +/- delta` for several small k/delta values.

Verify exact nearest-integer quotient.

Include negative values.

## T2 — fixed-width helper versus big.Int oracle

For many deterministic values spanning:
- low magnitudes;
- width-1 range;
- width-2 range;
- width-3 range;
- values near centered-capacity limits;

compare the fixed-width rounded division result to an independent `big.Int` reference.

Test multiple logical q divisors from actual CKKS parameter chains.

## T3 — logical divisor is independent of storage width

Construct the same authoritative lift and logical Level under widths 1, 2, and 3 where capacity allows.

Rescale each.

Verify all produce the same integer quotient and that the divisor is `q_ell`, not `f_{w-1}`.

At least one test must use a logical `q_ell` numerically different from every active `f_i`.

## T4 — coefficient-domain Rescale exactness

For synthetic multi-component Fast ciphertexts:
- decode input X;
- compute independent integer `Round(X/q_ell)`;
- run FastStorageRescale;
- decode output;
- verify exact integer equality coefficient-by-coefficient.

## T5 — NTT-domain Rescale exactness

Repeat T4 with private-F NTT input.

Verify output remains NTT and decodes to the same expected quotient.

This proves no NTT-coordinate division/reinterpretation occurred.

## T6 — bound transition

For nonuniform component bounds verify exactly:

[
B'_j=
leftlfloor
rac{B_j+(q_ell-1)/2}{q_ell}
ightfloor.
]

Do not collapse to one global bound.

## T7 — metadata transition

Verify:
- Level decrements exactly by one;
- Scale divides by the same logical q;
- degree unchanged;
- storage width unchanged;
- domain unchanged;
- public CKKS parameters unchanged;
- plaintext metadata preserved.

Include a case such as:

`LogicalLevel=3, StorageWidth=1/2/3`

where valid, to prove level/width independence.

## T8 — no implicit contraction

Create a width-3 input whose post-Rescale bound would fit width 1.

Output must still have width 3.

Storage contraction is a separate future operation.

## T9 — Standard logical oracle

Use the required oracle workflow:
- Fast export at Level ell;
- Standard CKKS Rescale;
- FastStorageRescale + export at Level ell-1.

Compare all active logical rows and Scale/Level metadata.

Cover at least:
- degree 1;
- more than one logical level/divisor;
- coefficient and NTT private-F inputs when practical.

## T10 — congruence with noncanonical lift

Construct a Fast authoritative lift of the form:

[
X=c+kQ_ell
]

for nonzero k, while remaining within private-F capacity.

Verify after FastStorageRescale that export agrees modulo `Q_{ell-1}` with Standard Rescale of `c mod Q_ell`.

This directly tests the theorem rather than only canonical-import cases.

## T11 — rejection and transactional behavior

Verify:
- Level 0 rejects;
- Montgomery rejects;
- malformed bounds reject;
- malformed rows reject;
- non-positive Scale rejects;
- input snapshot remains unchanged after every failure.

## T12 — chained arithmetic -> Rescale

Build a small valid path:

`FastStorageMul -> FastStorageRescale`

and verify:
- Mul output bound feeds Rescale without rescanning;
- degree remains raw-Mul degree;
- Scale changes from `Delta_X*Delta_Y` to `Delta_X*Delta_Y/q_ell`;
- exported logical result matches the independent integer/logical reference.

Do not relinearize.

## T13 — regressions

Run:
- new storage Rescale tests;
- all `schemes/ckks/fast` tests;
- relevant bootstrapping tests;
- `go test ./...`;
- `git diff --check`.

Any new production-path regression is blocking.

---

# 13. Performance rule

This is a correctness foundation with hot-path-aware implementation.

Required:
- no per-coefficient `math/big` in the new Rescale loop;
- reuse per-component temporary coefficient rows where practical;
- no work on dormant logical-Q residues;
- no hidden Standard/full-RNS fallback.

Not required yet:
- benchmark threshold;
- scratch-pool refactor;
- production evaluator integration;
- fused Mul+Rescale optimization.

---

# 14. Prohibitions

Do not:

- modify or replace historical production `Evaluator.Rescale` / `RescaleTo` behavior in this task;
- wire the new container Rescale into Bootstrap;
- implement storage contraction;
- implement multi-step `RescaleTo`;
- implement ModUp;
- implement KeySwitch;
- implement Relinearize;
- implement Rotate/Automorphism changes;
- modify key generation;
- change Add/Sub/Mul semantics from FAST-STORAGE-003;
- use `f_i` as a logical divisor;
- infer logical Level from storage width;
- mutate frontend CKKS parameters;
- add a Standard/full-RNS fallback;
- add CNN/application changes.

---

# 15. Success classification

Candidate classification:

`FAST_STORAGE_004_LOGICAL_RESCALE_ACCEPTED_CANDIDATE`

Successful Codex handoff:

`READY_FOR_WEB_REVIEW`

## Completion report

Report at minimum:

- exact Secondary commit;
- changed files;
- one-step API;
- fixed-width rounded-division implementation;
- evidence that the divisor is logical `q_ell`;
- coefficient/NTT exactness evidence;
- bound transition evidence;
- no-contraction evidence;
- Standard logical oracle evidence;
- noncanonical-lift theorem evidence;
- chained Mul -> Rescale evidence;
- regression tests;
- confirmation that historical production Rescale, Bootstrap, KeySwitch/Relinearize/Rotate, and Primary experiment semantics were not modified.
