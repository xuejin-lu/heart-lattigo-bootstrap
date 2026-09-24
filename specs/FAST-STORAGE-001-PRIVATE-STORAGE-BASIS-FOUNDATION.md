# FAST-STORAGE-001 — Fast Private Storage Basis Foundation

## Status

Executable Codex task.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`

Authoritative Secondary architecture:
- `docs/FAST_CKKS_SPEC.md`
- current constitution baseline includes commit `51c9d6a252e06daa339036c52fd9f50d79e568ea`

Accepted Primary observability foundation:
- `6f728f19a6300da339541e3f67259e873b9f942e`

## Purpose

Create the backend-private Fast storage-basis foundation required by the widened-storage architecture.

This task must establish a real software distinction between:

```text
LogicalQ[i]    = frontend CKKS q_i
FastStorage[i] = backend-private f_i
```

without yet wiring the new storage basis into ciphertext arithmetic, Rescale, KeySwitch, Relinearize, Rotation, Bootstrap, or application code.

The goal is to prove that the private storage representation itself is valid and reusable before production integration starts.

## Architectural authority

The implementation must follow Secondary `docs/FAST_CKKS_SPEC.md` Section 4.

The permanent rules include:

1. logical `Q` owns CKKS Level / Scale / Rescale semantics;
2. Fast storage `F` owns representation capacity only;
3. active storage width is independent of logical Level;
4. Fast storage primes are backend-private and must not alter public CKKS parameters;
5. centered uniqueness is the representation correctness condition;
6. storage contraction is representation-only;
7. ModUp canonicalization and production Rescale integration are future tasks.

Historical q0/q1/q012 code remains current implementation evidence. Do not silently reinterpret it as the new architecture.

## Scope

### Secondary modifications are authorized

Create the private-storage foundation in `schemes/ckks/fast` and, only if strictly necessary for clean construction/testing, narrowly use existing `ring` public APIs.

Prefer CKKS-specific Fast code. Do not refactor generic ring infrastructure.

### Primary modifications

Primary may receive only task pointer / documentation updates in this task.

Do not integrate runtime measurement hooks into production yet. The accepted Primary FAST-OBS-001 framework will be consumed in a later integration task.

## Required storage prime policy

Define one fixed shared storage basis:

[
F=(f_0,f_1,f_2)
]

with exactly three distinct near-60-bit primes.

Requirements:

- each prime is odd and prime;
- each prime is (<2^{60});
- each prime is NTT-friendly for Standard ring with maximum supported `LogN=16`;
- therefore each must satisfy at least
  [
  f_iequiv1pmod{2^{17}};
  ]
- using the stronger
  [
  f_iequiv1pmod{2^{18}}
  ]
  is acceptable and preferred if the selected fixed set already satisfies it;
- the same exact constants must be reused across all supported Standard-ring LogN values in this project;
- constants must be validated by Lattigo `ring.IsPrime`, `ring.NewSubRing`, and NTT constant generation for the supported range.

The exact numerical constants may be selected during this task, but must be deterministic, documented, and tested. Do not regenerate different primes per run or per LogN.

## Required API / data model

Create a small CKKS-Fast-private abstraction representing the storage basis.

Names may be improved, but conceptually provide:

```text
FastStorageBasis
  primes[3]
  width
  product / product bit length
  subrings or validation helpers
```

The abstraction must make it impossible or at least structurally awkward to confuse storage-prime lookup with logical `ringQ.SubRings[level].Modulus`.

Prefer explicit APIs such as:

```text
FastStoragePrimes()
NewFastStorageBasis(logN)
StoragePrime(index)
StorageProduct(width)
Validate(...)
```

Do not overload existing `q_i`-named helpers.

## Required representation helpers

Implement correctness-first helpers for coefficients represented in the Fast private basis.

At minimum:

### S1 — Encode lifted integer to storage residues

Given signed integer `X` and width `w in {1,2,3}`, return canonical residues:

[
Xmod f_0,ldots,Xmod f_{w-1}.
]

For the production-facing foundation, use fixed-width arithmetic where practical.

A test/reference `math/big` path is acceptable and encouraged for cross-checking.

### S2 — Centered reconstruct from storage residues

Given storage residues at width `w`, reconstruct:

[
Xin(-S_w/2,S_w/2],
qquad
S_w=prod_{i=0}^{w-1}f_i.
]

The implementation must support all widths 1, 2, and 3.

### S3 — Exact round-trip

For any representable `X` satisfying:

[
|X|<S_w/2,
]

require:

[
Decode(Encode(X))=X.
]

### S4 — Capacity metadata

Expose enough basis metadata to compute:

- active width;
- product bit length;
- centered-capacity bit length/log2;
- whether a given magnitude can be uniquely represented.

This must not depend on logical CKKS Level.

## Fixed-width design requirement

Three near-60-bit moduli imply a product near 180 bits.

The implementation may reuse/adapt the existing internal `uint192` machinery, but must not keep the new private storage foundation semantically tied to q0/q1/q2.

If existing `uint192` helpers are reused:

- rename or wrap them so the new public/internal API is storage-generic rather than q012-specific where practical;
- avoid a broad historical refactor in this task;
- preserve existing q012 behavior unchanged.

Do not delete the old q012 path.

## NTT / SubRing validation

For each selected `f_i`:

1. verify `ring.IsPrime(f_i)`;
2. verify the required congruence for maximum supported LogN;
3. construct `ring.NewSubRing(N, f_i)`;
4. generate NTT constants using the existing SubRing API;
5. perform NTT -> INTT round-trip tests for representative polynomials.

Required LogN coverage:

- at least LogN 13;
- LogN 16;
- and one smaller supported Standard-ring LogN.

The same fixed primes must pass all of them.

## Separation tests: LogicalQ versus FastStorage

Add explicit tests proving the new basis is not the logical chain.

At minimum:

1. construct ordinary CKKS parameters and read a logical `q_i`;
2. construct Fast storage basis and read `f_i`;
3. verify storage API never derives active width from logical Level;
4. verify storage product/capacity remains unchanged when the synthetic logical Level changes;
5. verify no public CKKS parameter object is mutated.

Do not require `f_i != q_i` for every possible frontend parameter by mathematical theorem, but for the project's tested profile the chosen private primes should be demonstrably independent constants.

## Required property tests

### T1 — Prime validity

All three fixed primes:
- prime;
- distinct;
- (<2^{60});
- satisfy required NTT congruence.

### T2 — SubRing / NTT validity

For each tested LogN and each storage prime:
- SubRing construction succeeds;
- NTT constant generation succeeds;
- coefficient-domain round trip through NTT/INTT succeeds exactly modulo the storage prime.

### T3 — Width-1 round trip

Test positive, negative, zero, and near-boundary values.

### T4 — Width-2 round trip

Same.

### T5 — Width-3 round trip

Same, including values above 128-bit magnitude so the 180-bit path is genuinely exercised.

### T6 — Strict centered boundary

For each width:
- (2|X|<S_w) accepted;
- boundary/non-unique value rejected or shown not exact-round-trippable under the uniqueness contract.

### T7 — Independent logical level

Change synthetic logical level metadata without changing storage width; verify storage representation and capacity do not change.

### T8 — Fixed-across-LogN

Construct basis for multiple LogN values and verify the exact same three storage prime constants are used.

### T9 — Reference cross-check

Cross-check fixed-width centered reconstruction against an independent `math/big` CRT/reference implementation over many deterministic or pseudo-random values.

### T10 — Existing q012 regression

Run targeted existing q012/Fast tests sufficient to show this foundation did not alter historical production behavior.

## API boundary

This task must **not**:

- change current `maintainedLimbCount`;
- replace q0/q1/q2 rows in existing ciphertexts;
- change Fast Add/Sub/Mul;
- change Fast Rescale;
- change ModUp;
- change ScaleDown;
- change DFT / EvalMod;
- change KeySwitch / Relinearize / Rotate;
- change Bootstrap orchestration;
- change key generation;
- add application/runtime mode selectors;
- change public CKKS parameters.

The task creates foundation code only.

## Expected files

Prefer a small coherent surface such as:

- `schemes/ckks/fast/storage_basis.go`
- `schemes/ckks/fast/storage_basis_test.go`

and the minimum helper changes needed to reuse fixed-width arithmetic safely.

Do not perform a repository-wide rename of q012 internals in this task.

## Validation

Required:

1. new storage-basis tests;
2. relevant existing `schemes/ckks/fast` targeted tests;
3. `git diff --check`;
4. broader package tests when practical.

Any pre-existing documented unrelated failure must be reported but does not authorize unrelated repair.

## Completion classification

Success classification:

`FAST_STORAGE_001_FOUNDATION_ACCEPTED_CANDIDATE`

Codex itself should report:

`READY_FOR_WEB_REVIEW`

after the bounded implementation/self-review/repair cycle.

## Completion report

Report:

- exact three storage prime constants;
- congruence and bit lengths;
- LogN values validated;
- NTT round-trip evidence;
- storage width 1/2/3 round-trip evidence;
- fixed-width vs `math/big` cross-check;
- existing q012 regression result;
- files changed;
- Primary and Secondary commits;
- push results;
- any pre-existing unrelated test debt.
