# FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009

**Status:** READY_FOR_CODEX
**Class:** M-approved architecture / bounded I extraction+adapter / E correctness validation.
**Web decision:** `results/FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008-web-review.md`.
**Reuse inventory:** `docs/FAST_DROPIN_EVALUATOR_REUSE_ARCHITECTURE.md`.
**Starting Fast Secondary:** `93ab7ecc5a864fd9bbacea55f0af730941552691`.
**Pinned genuine Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (do not modify).

## Purpose: reuse, not reimplementation

The previous explicit Fast evaluator **already has** Automorphism/Rotate code. This task must extract the existing ring-level automorphism implementation and allow BOTH (a) the original explicit `fast.Evaluator` and (b) unchanged ordinary `ckks.NewEvaluator`/ `Rotate`/`RotateNew` in the Fast fork to call **the same moved kernel**, with no Go import cycle and no frontend changes. The suspended older task `FAST-DROPIN-KEYLESS-ROTATE-007` MUST NOT execute.

## Frozen representation contracts

- **P0 public full-Q:** known-valid secret-key Fast simulation `EncryptNew` ciphertexts, and results of source-proven full-Q public primitive chains, have degree=1, active c1=0, and **ALL logical active Q residues q0..qLevel authoritative**. Public `Rotate(ct,k,out)` must transform **all Level+1 rows**, including Level>=4, preserving Level, Scale, dimensions, flags and c1=0. Use `rows=level+1`; this may be greater than `MaxQPrefixWidth=4`. Full-Q baseline is not yet a compact-Q speedup.
- **C0 explicit compact-Q:** keep the existing Fast evaluator and its legacy producer row selection (usually q0/q1, sometimes q012); preserve nil dormant rows, scratch/caching policy and Q-prefix bounds. It uses **the identical moved kernel** with exactly its existing authoritative rows. Do not widen all Fast producers to 4/full-Q as a side effect.
- There is **NO implicit conversion** between P0/C0. Do not let an ordinary public Rotate read/write a compact ciphertext with absent backing; reject transactionally. Presence of backing alone is not generic proof of non-stale data; initial P0 acceptance is bounded to source-proven public paths, and future mixing limitations must be documented.
- For c1=0, ring automorphism on c0 yields rotated encoded message and c1 stays zero; no Galois/KeySwitch needed. Inputs with any nonzero active c1 must be rejected without mutation even when a real Standard-layout Galois key exists. This is deliberately insecure numerical simulation, not secure HE.
- Keep pinned Standard unchanged. Keep `ckks.NewEvaluator` return type `*ckks.Evaluator`, API signatures and CKKS parameters unchanged. Preserve earlier 006A keyless MulRelin and all existing KeyLayoutFast rejection guards.

## Bounded implementation

1. Safely sync Primary main and Secondary fast-qprefix per AGENTS; verify clean state, read this spec, Web decision and Secondary `docs/FAST_QPREFIX_SPEC.md`. Examine actual source before editing. Unexpected conflicts require `NEEDS_WEB_REVIEW`.
2. Build a small import-neutral internal package (preferred `schemes/ckks/internal/fastcore`; use alternative only with documented evidence), accessible from both `schemes/ckks` and `schemes/ckks/fast` under Go internal import rules. It MUST NOT import either parent `schemes/ckks` or child `schemes/ckks/fast`. **Move/adapt the existing** automorphism coefficient/NTT row algorithm, validation, cache and scratch where necessary; avoid copying it to create a second kernel. Allow explicit validated `rows` up to `level+1` for P0 while retaining C0-specific <=4 policy in the explicit Fast wrapper (do not globally relax other Q-prefix validators). Keep cache/scratch evaluator-owned and concurrency expectations unchanged.
3. Rewire the **original** `fast.Evaluator.Automorphism` and `AutomorphismQPrefixRows` to delegate to the extracted common kernel, preserving their public method signatures and row-selection/metadata/aliasing behavior. Tests must prove existing explicit Fast path still uses its narrower row contract without touching dormant rows.
4. Add a Fast-capability-gated branch to ordinary public `ckks.Evaluator.Rotate` and `RotateNew` (not other primitives). Map k to the same CKKS Galois element, call the shared kernel over **ALL active full-Q rows**, and preserve output/metadata including in-place aliasing. No Galois key lookup; no native RLWE Automorphism or GadgetProduct. Reject unsupported degree, Standard/CI ring mismatch, nil metadata, nonzero c1, absent/malformed active Q, invalid domain/representation, or other unverified states **before writes**; never fall back to native KeySwitch. `RotateNew` nil input must return an error rather than panic.
5. Source-proven API preservation: identical frontend source/parameters with only library dependency changed. Do not alter `schemes/ckks/fast` Add/MulRelin/Rescale algorithms, public MulRelin 006A or Bootstrap. Do not introduce Fast-only imports to the frontend, hidden registration, build flags or custom ct constructors.

## Mandatory evidence and tests

- Compile both `schemes/ckks` and `schemes/ckks/fast` to prove legal shared dependency direction; inspect source/diff and instrument or test **both entrypoints call the same moved core function** (not merely produce numerically equal results).
- Positive Fast public secret-key EncryptNew, public Rotate/RotateNew with c1=0, real/complex nonconstant multiple slots: k=0, +1, -1; at least two levels including **Level>=4** (requires all rows, not q0/q1-only), custom LogSlots, alias/no-alias; compare each to independent plaintext rotation oracles and check Level/Scale/slots/metadata/c1.
- Test zero Galois key accesses using nil eval keys **and** a tracking keyset with a real native Standard-layout Galois key. Negative c1 !=0 with key available, missing active row, compact Level>=4 input, unsupported ring/degree/representation/nil: errors and byte-identical input/output; no key accesses or KeySwitch. Existing Fast compact wrapper regression must prove only the selected rows are touched, no dormant-row read/writes.
- Reuse immutable Primary frontend `tools/fast-dropin-ckks-primitive-api-audit-005/main.go` SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e` for **one Fast dependency** public Add→MulRelin→Rescale→Rotate numerical regression, if the final modified source passes preflight. Previously pinned Standard result may be reused with matched profile/input/source hashes. No frontend changes; no metric tuning.
- `go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe -count=1`, `git diff --check`, bounded self-review/one repair pass. If tests change source after numerical run, transparently report provenance and rerun affected focused tests.
- **No Bootstrap, benchmarks, LogN16, CNN, broad KeyGen edits, unbounded refactor, or changes to Standard.** Keep `FAST-STANDARD-PERF-REBASELINE-003` blocked.

## Deliverables

Secondary code/tests, safe commit/push; Primary `results/FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009-summary.md` reporting source and commit hashes, core reuse instrumentation, representation rows and negative cases, exact metrics, tests, limitations and final worktree states. This is an internal integration pilot, **not** broad CKKS optimization or secure-encryption claim.

Return `FAST_SHARED_AUTOMORPHISM_BRIDGE_COMPLETE_PENDING_WEB_REVIEW`, `FAST_SHARED_AUTOMORPHISM_BRIDGE_PARTIAL` or `FAST_SHARED_AUTOMORPHISM_BRIDGE_BLOCKED`; then `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`. If even moving the existing kernel requires an unapproved math/representation change, stop for Web rather than replacing it with a new algorithm.
