# FAST-STORAGE-002 — Fast Ciphertext Container and Basis Conversion Boundaries

## Status

Executable Codex task.

## Task class

`I — Implementation`

The mathematics/architecture for this task is already frozen by GPT Web in Secondary `docs/FAST_CKKS_SPEC.md`, especially Sections 4.11–4.15 and the Level-0 entry refinement.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`

Accepted storage foundation:
- `cc5028c872a89aff05ca43aa7e6f8c4269fcf8b5`

Architecture refinements:
- `3c3fe59f24fd9e80ab03ca566a39336c53b6c121`
- `aa99f85499899af53af87b569cec48d8ce4236c6`

## Purpose

Create the physically separate Fast ciphertext container and exact basis-conversion boundaries required before widened Fast storage can be integrated into Bootstrap.

This task is **not** a Bootstrap integration task.

It must prove:

1. logical CKKS metadata can remain independent from physical Fast storage width;
2. Level-0 logical-q ciphertext coefficients can be imported into the private F basis by canonical centered lifting;
3. an authoritative Fast lift can be exported back into arbitrary logical Q levels by coefficient-wise reduction modulo logical q_i;
4. NTT/Montgomery domain conversions across bases are performed correctly and never by row reinterpretation.

## Frozen container model

Create a CKKS-Fast-private container, conceptually:

```text
FastCiphertext
  LogicalLevel
  Scale
  Degree
  LogN / N
  IsNTT
  IsMontgomery
  ActiveStorageWidth
  Value []ring.Poly   // rows are modulo f_i, not q_i
```

Exact field names may improve, but these semantics are mandatory.

### C1 — LogicalLevel is explicit

`LogicalLevel` must be stored independently.

Do not define logical Level from `ring.Poly.Level()`.

For a Fast storage polynomial:

```text
poly.Level() == ActiveStorageWidth - 1
```

is a physical storage fact only.

### C2 — Storage rows are F-only

Every `Value[d].Coeffs[i]` in the Fast container is interpreted modulo private `f_i`.

No ordinary Lattigo `ringQ.SubRings[i]` may be used on these rows.

Use the accepted `fastStorageBasis` / storage subrings.

### C3 — Public logical metadata remains CKKS metadata

The container must preserve at least:

- logical Level;
- CKKS Scale;
- ciphertext degree;
- NTT state;
- Montgomery state;
- plaintext batching/dimension metadata needed for later public reconstruction.

It must not mutate the frontend parameter object.

### C4 — Width independent from level

For example this must be valid:

```text
LogicalLevel = 5
ActiveStorageWidth = 3
```

and also:

```text
LogicalLevel = 1
ActiveStorageWidth = 3
```

No API may infer one from the other.

## Required constructors / structural helpers

Provide a minimal API conceptually equivalent to:

```text
NewFastCiphertext(params, degree, logicalLevel, storageWidth)
CopyNew()
ResizeDegree(...)
SetLogicalLevel(...)
StorageWidth()
N()
Degree()
```

Do not overbuild serialization in this task.

Changing logical Level must not resize storage rows.

Changing storage width must be a separate explicit operation and must obey storage-capacity correctness where reconstruction is required.

## Boundary A — Level-0 LogicalQ -> FastStorage import

The production-oriented import implemented in this task is restricted to logical Level 0.

Input:
- ordinary `*rlwe.Ciphertext`;
- logical Level 0;
- Standard ring;
- matching N;
- positive Scale;
- degree arbitrary but tested at least for degree 1;
- source may be coefficient or NTT domain;
- source may be ordinary or Montgomery form only if conversion support is implemented correctly.

Preferred first implementation may reject Montgomery input if that keeps the boundary exact and explicit. Do not silently reinterpret it.

For each component/coefficient:

1. convert source q0 row to ordinary coefficient-domain canonical residue;
2. choose:
   [
   C = Center_{q_0}(c) in (-q_0/2,q_0/2];
   ]
3. prove:
   [
   |C| < S_w/2;
   ]
4. reduce C into active Fast storage primes;
5. transform to the requested Fast representation domain using Fast storage subrings.

The imported Fast value has:

```text
LogicalLevel = 0
Scale = source.Scale
Degree = source.Degree()
ActiveStorageWidth = explicit requested width
```

No hidden q1/q2 logical rows are materialized.

### Import equality contract

For every coefficient:

[
X_F = Center_{q_0}(c).
]

This is exact integer equality at the entry boundary, not merely congruence.

## Boundary B — FastStorage -> LogicalQ export

Implement exact export from Fast container at logical Level `ell` to an ordinary `rlwe.Ciphertext`.

For each authoritative Fast coefficient lift X:

[
r_i = X mod q_i,
qquad 0 le i le ell.
]

Required:

- output logical Level equals `FastCiphertext.LogicalLevel`;
- output Scale equals Fast Scale;
- degree preserved;
- public CKKS parameter identity unchanged;
- output domain state explicitly defined and tested.

For this task, prefer exporting to ordinary non-Montgomery coefficient domain first, then optionally NTT if requested through an explicit API parameter.

Do not directly copy Fast NTT rows into logical NTT rows.

## Cross-basis domain rule

Basis conversion must conceptually use:

[
NTT_{source}
	o
INTT_{source}
	o
integer/reduction
	o
NTT_{target}.
]

If source is Montgomery:
- IMForm before integer reconstruction.

If target is Montgomery:
- reduce to target modulus first, then MForm.

No row reinterpretation.

## Generic high-level import policy

Do **not** make generic high-level Q->F import part of production API in this task.

A diagnostic helper may exist only if it:
- canonicalizes modulo full `Q_level`;
- checks every coefficient against Fast storage capacity;
- rejects insufficient capacity.

But this is optional.

The required production-oriented entry is Level 0 only.

## Required tests

### T1 — Structural level/width independence

Construct:

```text
LogicalLevel=5, Width=3
LogicalLevel=1, Width=3
LogicalLevel=5, Width=2
```

Verify:
- `LogicalLevel` remains exact;
- storage poly row count follows width only;
- changing logical level leaves storage bytes/rows untouched.

### T2 — Level-0 coefficient-domain import

For synthetic degree-1 logical q0 ciphertext coefficients:
- positive;
- negative centered representatives;
- zero;
- boundary-near values.

Import to widths 1/2/3 and verify exact integer equality after Fast reconstruction.

### T3 — Level-0 NTT-domain import

Create q0 coefficient-domain data, NTT it with logical q0, import, and verify reconstructed Fast coefficients equal the original centered integers.

This proves no NTT-coordinate reinterpretation occurred.

### T4 — Export to logical Level 0

Import Level-0 q0 -> F -> export q0.

Verify exact q0 residue equality and metadata preservation.

### T5 — Export to higher logical level

Take a known Fast lift X and set synthetic `LogicalLevel > 0`.

Export and verify for every logical row:

[
row_i[k] = X_k mod q_i.
]

### T6 — Fast NTT -> logical coefficient export

Place Fast container in Fast NTT domain; export to logical coefficient domain; verify exact residues.

### T7 — optional logical NTT export

If implemented, verify coefficient export followed by logical NTT equals direct requested NTT export.

### T8 — Montgomery rejection or correctness

Either:
- explicitly reject unsupported Montgomery source/target states and test rejection;

or:
- support them and prove exact round-trip semantics.

No ambiguous partial support.

### T9 — capacity guard

Construct an artificial Level-0 logical modulus / source case or direct boundary helper where the chosen centered representative does not fit requested Fast width.

Verify import rejects rather than wraps.

For the project's actual q0 and width 1/2/3, record that capacity passes.

### T10 — metadata preservation

Verify:
- Scale;
- degree;
- batching metadata;
- dimensions;
- bit-reversed flag;
- explicit logical Level.

### T11 — current production regression

Run existing Fast/q012 tests and Bootstrap-targeted tests sufficient to prove this standalone container did not change current production behavior.

## Performance rule

This task prioritizes correctness.

However:
- constructors may cache/reuse `fastStorageBasis` if clean;
- do not place `math/big` per-coefficient work into a future hot path unnecessarily;
- fixed-width storage helpers should be used where practical.

No performance benchmark gate yet.

## Files

Prefer a small CKKS-Fast-local surface such as:

- `schemes/ckks/fast/storage_ciphertext.go`
- `schemes/ckks/fast/storage_conversion.go`
- corresponding tests.

Do not modify generic `core/rlwe.Ciphertext` layout.

## Prohibitions

- No Bootstrap pipeline wiring.
- No change to `FastEvaluator.bootstrapCore`.
- No change to current `ScaleDown`.
- No replacement of current `ModUp`.
- No Fast Add/Sub/Mul integration.
- No Rescale integration.
- No KeySwitch/Relinearize/Rotate work.
- No key-generation changes.
- No public parameter mutation.
- No reinterpretation of f_i rows as q_i rows.
- No CNN/application changes.

## Validation

Required:

1. new container/conversion tests;
2. existing `schemes/ckks/fast` tests;
3. relevant bootstrapping Fast tests;
4. `git diff --check`;
5. broader `go test ./...` when practical.

## Success classification

`FAST_STORAGE_002_CONTAINER_BOUNDARIES_ACCEPTED_CANDIDATE`

Codex handoff status:

`READY_FOR_WEB_REVIEW`

## Completion report

Report:

- container fields/API;
- Level/width independence evidence;
- Level-0 q0 -> F import evidence for widths 1/2/3;
- F -> logical Q export evidence;
- NTT boundary evidence;
- Montgomery policy;
- capacity guard evidence;
- regressions;
- changed files;
- Primary/Secondary commits and push status.
