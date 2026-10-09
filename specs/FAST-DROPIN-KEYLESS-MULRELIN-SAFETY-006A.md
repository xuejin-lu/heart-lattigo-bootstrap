# FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A

**Status:** READY_FOR_CODEX
**Task type:** bounded I repair against approved invariant (no architecture redesign).
**Decision:** `results/FAST-DROPIN-KEYLESS-MULRELIN-006-web-review.md`; original `specs/FAST-DROPIN-KEYLESS-MULRELIN-006.md`.
**Primary current baseline:** `3cf7e0a5d230148f30ff3b6a0c36d7637450ed80`.
**Secondary current baseline:** `6ce15cf8b949c8a89f610bcca5ab506dd3f560f8`.
**Pinned Standard (do not modify):** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

## Specific contract violation to fix

In Fast marked CKKS parameters, `*ckks.Evaluator.MulRelin/MulRelinNew` with ciphertext/ciphertext operands must only execute the keyless simulation when the supported zero-secret preconditions hold. Current code returns `handled=false` on c1 nonzero, then falls through to ordinary native KeySwitch **if** a Standard-layout RelinearizationKey is available. This is contrary to the originally frozen requirement that nonzero-c1 inputs be rejected without mutation, regardless of keys. With no key the path already fails, but that test is insufficient.

**Frozen correction:**
1. For a Fast-marked evaluator and ciphertext/ciphertext MulRelin inputs, **unconditionally reject** any active nonzero c1, unsupported ciphertext degree, unverified/missing active-Q backing, or unsupported input NTT/Montgomery representation, independent of whether any regular or Fast evaluation key is available. Error must be clear and inputs/output unmodified. Never fall through to native MulRelin/GadgetProduct for an unsupported ciphertext/ciphertext pair. A ciphertext with native nonzero c1 must never be forcibly zeroed.
2. For supported fresh degree-one zero-secret inputs `(a0,0),(b0,0)`, preserve the already implemented keyless full-active-Q product `(a0*b0,0)`, original Level/Scale/logDimensions/c1-zero semantics, alias-safe scratch behavior, and no relin key access. No change to algorithms, ring arithmetic, Q-prefix width, or output format.
3. Keep `ckks.NewEvaluator`, `MulRelin`, and `MulRelinNew` public signatures unchanged; do not alter generic non-Fast RLWE/BFV/BGV semantics, genuine pinned Standard, `schemes/ckks/fast` optimized arithmetic, EncryptNew, Bootstrap or KeyLayoutFast rejection guards. Constrain this fail-closed behavior to the Fast-specific ciphertext/ciphertext MulRelin simulation. Preserve valid existing scalar/plaintext overload behavior; if new overload ambiguity requires redesign, stop for Web review rather than broadening.
4. Reject bad inputs *before* mutating any aliased output, including destination metadata/Scale. Check cases with a normal Standard-layout relin key, with nil key set, and (where a lightweight existing fixture permits) Fast layout keys. A standard-layout key must not re-enable native fallback. Failures must not panic.

## Minimal tests and demonstration

- Add a targeted negative test in Secondary using `ckks.NewEvaluator(params, normalEvalKeys)` with an **actual available normal-layout relinearization key** (generated with ordinary CKKS public keygen API), and a nonzero-c1 ciphertext/ciphertext pair. Verify descriptive error and byte-identical inputs/output (including metadata) and that no native GadgetProduct path is entered. An intentionally modified c1 from a fresh simulated ciphertext is sufficient; optionally add a native encrypted nonzero-c1 fixture without changing production code.
- Repeat bad-input check with `nil` eval keys, and verify unsupported higher-degree ciphertext is rejected **even when a normal relin key exists**. Test `MulRelinNew` returning a clean error as well as `MulRelin` destination transactional semantics.
- Keep passing Level 0/3, squaring, aliasing, complete-Q backing and nil-key keyless tests; include a supported keyless product with a non-nil ordinary eval keyset to prove the optimized branch is selected regardless of key availability.
- Run `go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe -count=1`, `git diff --check`. Perform one bounded self-review/repair pass. No `go test ./...` requirement, no Bootstrap, benchmark, LogN16 or performance tuning.
- Use the **same unmodified** prior frontend `tools/fast-dropin-ckks-primitive-api-audit-005/main.go` (SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`) for a single Fast dependency numerical regression if preflight indicates no source/metadata problem. The pinned Standard and prior comparison remain valid; do not repeat Standard computation unless an actual change to shared frontend/provenance requires it. No frontend source changes or Fast-specific selection flags.

## Safe workflow and output

- Start with clean Primary/Secondary and `AGENTS.md` single fetch/ff-only sync, reread latest task/spec. If unexpected conflict, STOP; no reset/rebase/force push.
- Modify only Secondary `schemes/ckks/evaluator.go`, `schemes/ckks/evaluator_fast_zero_secret.go` as needed and focused tests. Primary deliver a new concise `results/FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A-summary.md`; preserve all older reports. State exact source SHA, tests, positive/negative behaviors and whether a native standard-layout key was present in failing test.
- After bounded self-review, safe commits/ff pushes and clean worktrees; classify `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_PARTIAL`, or `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_BLOCKED`, followed by `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
- Do not claim universal drop-in, runtime speedup, cryptographic security or optimized Rotate/Rescale. Formal `FAST-STANDARD-PERF-REBASELINE-003` stays BLOCKED.
