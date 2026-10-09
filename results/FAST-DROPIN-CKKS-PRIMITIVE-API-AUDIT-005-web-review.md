# Independent Web Review — FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005

**Disposition: ACCEPT the bounded audit as PARTIAL, not optimized Fast CKKS primitive validation.**

## Evidence reviewed

- Primary report committed `9b48bfbb1096b41827bed3d280bdb3163726b60e`; Secondary Fast remained `eb7793f1e449702590cdba3d9603261623354f80`; pinned unmodified Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- The same frontend source (`tools/fast-dropin-ckks-primitive-api-audit-005/main.go`, SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`) ran with identical input/profile. Reported final profile LogN13, four Q primes, one P prime, initial level 3, scale 2^45, 16 slots. No Bootstrap calls, no performance measurements.
- Add/MulRelin/Rescale/Rotate passed independent cleartext oracles under native `ckks.NewEvaluator`, including correct metadata and Fast c1=0 checkpoints. Fast-vs-cleartext RMSE was ~6.37e-14 for Add, ~5.32e-15 for MulRelin, ~5.51e-15 after Rescale and Rotate. These do NOT measure `schemes/ckks/fast.Evaluator`.
- Source: `schemes/ckks/evaluator.go` in Standard and Fast was identical. `ckks.NewEvaluator` returns concrete `*ckks.Evaluator` embedding `*rlwe.Evaluator`. `schemes/ckks/fast` imports `schemes/ckks`, so a direct import from ckks to fast would create a package import cycle. The explicit Fast evaluator is not a drop-in Go concrete type/signature replacement. Generic native `MulRelin` requests a relinearization key and invokes `GadgetProduct`; native rotation similarly uses Galois KeySwitch. Dedicated `KeyLayoutFast` keys are guarded against use by generic native KeySwitch, correctly so.
- The audit's P-free initial trial was invalid for Standard key switching and excluded; the report records a P-containing final comparable profile. The temporary full output vectors were not uploaded; this review independently inspected committed report/source semantics and summary, not the machine-local full vectors or actual test execution.

## Architecture decision (bounded next step)

The public Go API must continue returning `*ckks.Evaluator` from `ckks.NewEvaluator(params, evk)`; a wholesale direct dispatch to `*fast.Evaluator` is **not** a valid source-compatible approach. This is not a reason to alter the original Standard public API.

For an *intentionally insecure*, CKKS-specific zero-secret simulation, a **newly encrypted** degree-1 ciphertext with `c1 = 0` represents its message by `c0`. Given two such inputs,
`(a0,0) * (b0,0) = (a0*b0,0,0)`,
so a **relinearized degree-one output** can be `(a0*b0, 0)`, without relinearization-key generation/use for that operation. This is a sufficient arithmetic reason to add a bounded keyless `MulRelin` branch behind the existing public `*ckks.Evaluator` in the Fast fork, rather than routing into native `GadgetProduct`. It is **not** license to zero `c1` from a native encrypted nonzero-c1 input, change secret/key security semantics, or silently bypass KeySwitch for other operations.

The first implementation must preserve **all logical active Q limbs**, output degree=1, exact logical Level and CKKS Scale product, NTT/Montgomery representation correctness, metadata, and ciphertext aliasing guarantees, and must reject any nonzero-c1 input *before mutation*. This is a full-Q-authority correctness pilot, **not yet a Q-prefix partial-row speedup**. Keep any existing compact Q-prefix input separate and reject/unhandled until its row authority can be proven; do not read or treat dormant residues as authoritative.

Implement the internal Fast-only branch in `schemes/ckks` using the existing Fast CKKS marker/capability; do not attempt to import sibling `schemes/ckks/fast` from `schemes/ckks`. Preserve unchanged frontend method names/signatures and leave the pinned Standard untouched. This pilot does not settle Add/Rotate/Rescale optimized dispatch, future integration with Q-prefix kernels, or Fast-vs-Standard runtime speedup.

Next authorized bounded task: `FAST-DROPIN-KEYLESS-MULRELIN-006`. No full Bootstrap or benchmark is required for this stage.
