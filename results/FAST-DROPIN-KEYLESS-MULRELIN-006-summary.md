# FAST-DROPIN-KEYLESS-MULRELIN-006

Classification: `FAST_DROPIN_KEYLESS_MULRELIN_COMPLETE_PENDING_WEB_REVIEW`; handoff: `READY_FOR_WEB_REVIEW`.

## Bounded result

Implemented a Fast-only, keyless Zero-Secret `MulRelin` / `MulRelinNew` route behind the unchanged public `ckks.NewEvaluator(params, evk)` API. The route is selected only by the existing Fast CKKS parameter capability and only handles degree-one ciphertext pairs with fully materialized active logical-Q rows and zero `c1` in every active row.

The route reuses the existing CKKS multiplication kernel with relinearization disabled, computes into scratch to preserve valid output aliasing, then returns `(a0*b0, 0)` with degree 1, exact product Scale, minimum supported Level, and propagated CKKS dimensions. It never reads a RelinearizationKey or calls `GadgetProduct`. Compact/missing active rows are rejected before native fallback. Nonzero-`c1` ciphertexts retain the existing native key-switch route; if its key is absent, the operation fails before mutating the destination. Unsupported Montgomery and non-NTT forms are rejected transactionally.

No Standard source, Q-prefix arithmetic, Bootstrap, key generation, frontend, configuration, or other scheme was changed. This does not establish general Fast evaluator transparency, compact-Q-prefix multiplication, security equivalence, or a speedup. The public `Rescale` and `Rotate` checkpoints below used the existing generic CKKS path on fully materialized rows; they are not claimed as optimized Fast dispatch.

## Source and validation map

- Public dispatch and API documentation: Secondary `schemes/ckks/evaluator.go:717-779`.
- Fast capability detection, active-row/c1 guards, scratch product, and metadata handling: Secondary `schemes/ckks/evaluator_fast_zero_secret.go:9-136`.
- Public-API numerical, no-key, level/scale/dimension, aliasing, negative-c1, compact-row, out-of-range-level, and unsupported-representation tests: Secondary `schemes/ckks/fast/keyless_mulrelin_test.go:14-203`.
- A nil evaluation-key set succeeds for the zero-secret operation, independently showing that the key is not accessed. Existing CKKS tests with ordinary nonzero-`c1` ciphertexts and standard-layout keys still pass. Existing Fast-key/native-consumer rejection tests remain covered by the passing Fast package suite.

Focused validation passed:

```text
GOCACHE=/private/tmp/fast-dropin-keyless-mulrelin-gocache go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe -count=1
ok   github.com/tuneinsight/lattigo/v6/schemes/ckks
ok   github.com/tuneinsight/lattigo/v6/schemes/ckks/fast
ok   github.com/tuneinsight/lattigo/v6/core/rlwe
git diff --cached --check
```

`go test ./...` was not required by this bounded task and was not run.

## Same-source Standard/Fast oracle run

The unchanged frontend was executed once per backend. Each checkpoint is decoded through the public `DecryptNew`/`Decode` lifecycle and compared with its cleartext oracle. All metrics are finite. “Max” is maximum complex absolute error. `c1` is nonzero coefficients / total active coefficients.

| Checkpoint | Level / Scale | Standard RMSE / max | Fast RMSE / max | Standard c1 / Fast c1 |
|---|---|---:|---:|---:|
| EncryptNew, Add operand A | 3 / `2^45` | `5.096e-13 / 8.758e-13` | `4.307e-14 / 8.564e-14` | `32768/32768 / 0/32768` |
| EncryptNew, Add operand B | 3 / `2^45` | `6.335e-13 / 1.170e-12` | `4.806e-14 / 9.500e-14` | `32768/32768 / 0/32768` |
| Add | 3 / `2^45` | `8.221e-13 / 1.813e-12` | `6.366e-14 / 9.928e-14` | `32768/32768 / 0/32768` |
| EncryptNew, MulRelin operand A | 3 / `2^45` | `4.077e-13 / 7.159e-13` | `4.307e-14 / 8.564e-14` | `32768/32768 / 0/32768` |
| EncryptNew, MulRelin operand B | 3 / `2^45` | `4.060e-13 / 8.716e-13` | `4.806e-14 / 9.500e-14` | `32768/32768 / 0/32768` |
| MulRelin | 3 / `2^90` | `4.853e-14 / 9.607e-14` | `5.321e-15 / 1.212e-14` | `32768/32768 / 0/32768` |
| Rescale | 2 / `2^50.999999441053866` | `4.772e-14 / 8.984e-14` | `5.509e-15 / 1.249e-14` | `24576/24576 / 0/24576` |
| Rotate left 1 | 2 / `2^50.999999441053866` | `5.156e-14 / 9.090e-14` | `5.508e-15 / 1.249e-14` | `24576/24576 / 0/24576` |

The Fast output retained `c1=0` through every measured checkpoint. The numerical runs are correctness evidence for this fixed profile, not direct timing or an optimized-dispatch claim for Add, Rescale, or Rotate.

## Provenance

- Primary source commit at measurement time: `ff824a20bbc82635e39eb51f1f5cf0957d26dd53`, clean.
- Standard Lattigo: pinned unmodified commit `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean.
- Fast Secondary: `6ce15cf8b949c8a89f610bcca5ab506dd3f560f8` on `fast-qprefix`, pushed and clean.
- Shared frontend: `tools/fast-dropin-ckks-primitive-api-audit-005/main.go`, SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`.
- Profile: LogN=13; `LogQ=[55,39,40,39]`; `LogP=[60]`; initial Level=3; default Scale=`2^45`; LogSlots=4; secret Hamming weight=192; Rotate left 1. Profile SHA-256 `8a2b9ee73d1beee7759171b0f8ea08a554c07392b666baed7f0212a3664fa91a`.
- Generated Q primes: `[36028797018652673,549755731969,1099511480321,549756026881]`; P prime: `[1152921504606830593]`.
- Deterministic input SHA-256: `b46752705bcda9226eefbf6f19aa5d75e03e74226f029f7e6db5694050d257a4`.
- Environment: Go `go1.26.4`, `darwin/arm64`. One execution per backend; Bootstrap calls: 0; benchmarks: 0.

The runs used the ordinary `ckks.NewKeyGenerator` and public `ckks.NewEvaluator` in both dependency trees. The Fast run did not construct `schemes/ckks/fast.Evaluator`; the new zero-secret `MulRelin` capability was selected internally by the Fast CKKS parameter type. Raw JSON evidence remains in `/private/tmp` and is not committed.
