# Independent Web Review — FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004

**Scientific disposition: ACCEPT the bounded secret-key EncryptNew → Bootstrap → DecryptNew drop-in simulation lifecycle.** This is not blanket compatibility, cryptographic equivalence or speedup acceptance.

## Evidence inspected

- Primary result `19ebcb888cd4e806f40df24e785ac75ec5d85d8e`, shared frontend `tools/fast-dropin-zero-secret-encrypt-bridge-004/main.go` committed `9fe762063af07df92375b0a3464c22430aff906f`, source SHA-256 `f81b4222acc77cd48557fca4626606a42be50bb70bfcaae0600d39c031280e32`.
- Secondary Fast code `eb7793f1e449702590cdba3d9603261623354f80`; pinned genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Primary/Secondary current heads were checked, and the recorded source changed only the described code/test surfaces.
- Secondary `schemes/ckks/params.go` adds an unexported-provider-interface capability method on CKKS parameters; `core/rlwe/encryptor.go` uses it to select a **secret-key-only** Fast Encrypt path. That path copies encoded plaintext Q rows to c0, zeros c1, clones metadata and preserves the public constructor/method names. Native ordinary RLWE and BGV retain their tested nonzero-c1 encryption.
- Fast bootstrap input gate now rejects a nonzero active c1 explicitly. The same shared frontend links against either library at build time, uses normal CKKS public API methods and reports immediate and post-bootstrap numeric checks.
- Both backends reportedly ran **exactly one** public E32 Bootstrap on LogN13, same config, input and Q/P prime hashes, no warmups, with no frontend-specific branches or manual c1 patching. The actual temporary full decoded arrays/logs were not uploaded; this Web review verified committed report, recorded checks/hashes, and source, not independent mathematical recomputation of the 4096 slots.

## Measured result (fixed LogN13, E32, 4096 slots)

| Output vs original | Standard | Fast |
|---|---:|---:|
| Input c1 nonzero coefficients | 8192 | 0 |
| Immediate DecryptNew/Decode RMSE | 8.247557280558041e-12 | 7.449975091875706e-13 |
| Bootstrap calls | 1 | 1 |
| Bootstrap output c1 nonzero | 16384 | 0 |
| Bootstrap RMSE | 4.697396790468602e-9 | 1.1907546782784477e-9 |
| Maximum complex error | 1.3779606222747244e-8 | 5.805793207821362e-8 |
| Bootstrap SNR | 132.79005231297143 dB | 144.7107513180764 dB |

Fast vs Standard decoded RMSE = **4.821647086585007e-9**; maximum complex error = **4.901316703636267e-8**; SNR = **132.56328882843212 dB**.

## Accepted invariant and limitations

For the tested secret-key CKKS path, Fast's intentionally insecure virtual zero-secret input represents `m` as `(Encode(m),0)`. Regular `rlwe.NewDecryptor(params,sk).DecryptNew` is meaningful when all c1 residues are zero, even though the public `sk` object itself was sampled normally. This is a numerical simulation, NOT a secure RLWE encryption procedure.

This confirms the unchanged frontend *for a single Bootstrap lifecycle*, NOT arbitrary CKKS application workloads. The following remain **unproven**:
- `rlwe.NewEncryptor(params,pk)` using public keys, `EncryptZeroNew`, non-secret-key encryption overloads or ciphertext exchange.
- Generic `ckks.NewEvaluator` with `Add`, `MulRelin`, `Rotate`, `Rescale` and their proper Fast zero-secret semantics; currently that public constructor still directly builds the native `rlwe.Evaluator`, not the explicit `schemes/ckks/fast.Evaluator`. Fast layout evaluation keys must never silently pass through ordinary Standard KeySwitch operations.
- Other N/Q/P/logSlots/circuit orders, multi-operator CNN chains and measured end-to-end speedup.
- The new CKKS parameter capability acts on all `ckks.Parameters` passed to `rlwe.NewEncryptor` in the Fast fork; it is intentionally global for this library build, not a general per-evaluator secure/fast runtime switch. Preserve any normal-implementation coexistence invariants, and record exceptions if further operations conflict with them.
- Fast E32 currently generates specialized Dense/Sparse keys despite eliding their use; startup/memory costs remain unmeasured.

**Next task:** bounded same-source Add/MulRelin/Rotate/Rescale **call-path and semantic audit**, using fresh independent ciphertexts, no Bootstrap, before authorizing a compatibility implementation. Do not claim general application drop-in now.
