# Independent Web Review — FAST-DROPIN-KEYLESS-MULRELIN-006

**Decision: PARTIAL / FIX REQUIRED before full acceptance.** The keyless zero-secret path and narrow numerical evidence are supported; the nonzero-c1 fallback violates the frozen fail-closed contract and must be repaired. Do not call the overall task completed or claim performance speedup.

## Provenance reviewed

- Secondary production commit `6ce15cf8b949c8a89f610bcca5ab506dd3f560f8`, child of `eb7793f1e449702590cdba3d9603261623354f80`. Modified only `schemes/ckks/evaluator.go`, added `schemes/ckks/evaluator_fast_zero_secret.go`, and `schemes/ckks/fast/keyless_mulrelin_test.go`.
- Primary report `3cf7e0a5d230148f30ff3b6a0c36d7637450ed80`, child of `ff824a20bbc82635e39eb51f1f5cf0957d26dd53`. Primary and Secondary remote HEAD matched those commits at review.
- Same frontend `tools/fast-dropin-ckks-primitive-api-audit-005/main.go`, reported SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`, unchanged in both dependency runs. Standard pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. The result report records passing focused Go package tests; review was of committed code/report rather than execution on user's machine or recomputation from temporary full vectors.

## Confirmed bounded success

- For fully materialized, active-Q, NTT non-Montgomery degree-one Fast zero-secret inputs `(a0,0)` and `(b0,0)`, the newly gated public `ckks.Evaluator.MulRelin/MulRelinNew` route uses `eval.mulRelin(..., false, scratch)`, avoiding `CheckAndGetRelinearizationKey` and `GadgetProduct` inside the handled path; then emits degree-one `(a0*b0,0)`, retaining logical Scale/Level. A `nil` eval-key-set test establishes it does not need access to a relin key on the accepted path.
- Source tests cover Level 0 and 3, fresh public EncryptNew, source/target aliasing, squaring, metadata, no-key, basic malformed/backing/unsupported-representation cases; the summary reports the requested packages passed. Existing native Fast key-layout KeySwitch rejection guards were left in place.
- The fixed same-source, LogN13 four-Q/one-P/16-slot numerical run reported Fast MulRelin complex RMSE `5.320587702687035e-15`, maximum `1.2117592599745943e-14` vs genuine Standard `4.853e-14` RMSE, `9.607e-14` max; zero c1 throughout checkpoints. No Bootstrap or benchmark. These values establish neither general CKKS compatibility nor speedup.

## Explicit blocker (source-proven)

The prior spec `specs/FAST-DROPIN-KEYLESS-MULRELIN-006.md` requires **reject any nonzero-c1 input before mutation**, not merely fail when an eval key happens to be missing. Yet in `schemes/ckks/evaluator_fast_zero_secret.go` the Fast helper returns `(handled=false, nil)` on detecting an active nonzero c1. In `schemes/ckks/evaluator.go`, the caller checks whether a normal RelinearizationKey is available and, if it is, falls through to the ordinary native `eval.mulRelin(..., true,...)` / `GadgetProduct`. Thus a Fast-marked, nonzero-c1 pair with a valid standard-layout key can proceed down an unintended native full-Q KeySwitch path instead of being unconditionally rejected. The existing negative test supplies `ckks.NewEvaluator(params,nil)`, so it proves only the **no-key** failure, not fail-closed behavior with an available key.

The public constructor must not silently reinterpret native nonzero-c1 ciphertexts as zero-secret nor silently accept the native key-switch fallback when an input violates this deliberately narrow Fast simulation contract. This is an implementation-contract bug, not evidence invalidating the already observed keyless c1-zero numerical data.

Additionally, the current `MulRelin` dispatch is limited to operand type `*rlwe.Ciphertext` and degree-one operands; unsupported ciphertext degree paths can still enter normal code. Confirm fail-closed handling for ciphertext/ciphertext operands outside the explicit pilot contract, independent of key availability, while retaining existing conventional scalar/plaintext operations. Do not broaden scope to all CKKS overloads.

The structural row check verifies coefficient backing presence/length, not the independent mathematical correctness of arbitrarily stale full-length residues. Results should remain scoped to freshly generated, known fully materialized inputs.

## Decision

- Accept **bounded keyless branch, tests, and numerical evidence provisionally**.
- Overall classification: `FAST_DROPIN_KEYLESS_MULRELIN_PARTIAL` until unconditional input-boundary rejection is implemented and tested with and without available normal-layout keys.
- Authorize a minimal fail-closed follow-up `FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A`. Preserve the originally passing keyless path, leave pinned Standard source and independent Fast Q-prefix evaluator unchanged; no Bootstrap or benchmark. Only after its verification may subsequent primitives (e.g. Rotate) be considered.
