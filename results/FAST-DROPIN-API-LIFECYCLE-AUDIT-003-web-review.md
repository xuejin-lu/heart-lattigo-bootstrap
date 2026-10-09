# Web Review — FAST-DROPIN-API-LIFECYCLE-AUDIT-003

**Disposition: ACCEPT as a correctly bounded PARTIAL audit; zero Bootstrap executions were appropriate.** This is not a successful full drop-in CKKS lifecycle.

## Reviewed evidence

- Primary result `441ee378690de9ca8567520c8fedf9ad00aaa1af`; source frontend `tools/fast-dropin-api-lifecycle-audit/main.go` from `7a87502f4b09205fe5d711e8fc998d0a12ddc9e2`, same source/hash in Standard and Fast builds.
- Standard pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Fast production pinned `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`, with later Secondary HEAD `68c3da8ec18222f985ebb5c43f6a8495b853cfc2` affecting only AGENTS startup synchronization. Both source builds and runtime preflight succeeded with E=32 and identical canonical input/Q/P/config hashes.
- The same frontend called ordinary RLWE KeyGen, secret-key EncryptNew, GenEvaluationKeys, bootstrapping.NewEvaluator and DecryptNew. Ordinary native EncryptNew produced nonzero c1 (8192 of 8192 coefficients) in both builds; decrypted RMSE versus input was approximately 8.3e-12. Fast keys were marked compatible and Fast NewEvaluator selected the separate Fast engine. Neither build performed Bootstrap.
- Source check: `core/rlwe/encryptor.go` still performs genuine RLWE encryption with the supplied nonzero secret; `circuits/ckks/bootstrapping/keys.go` conditionally picks Fast key material; `evaluator.go` conditionally picks Fast Bootstrap, which interprets the input as zero-secret and omits Dense/Sparse KeySwitch. Current Fast public input preflight does **not** prove that ordinary nonzero-c1 RLWE ciphertext represents its original plaintext under zero-secret interpretation.

## Mathematical determination for next step

For degree-1 native Standard RLWE ciphertext, plaintext information is in `c0 + c1*s` (under the appropriate ring convention and scaling). With actual `s != 0`, interpreting it as zero-secret changes the effective plaintext to `c0`, which generally is **not** the original message. Therefore **passing a native nonzero-c1 ciphertext unchanged to Fast is not numerically sound**, even if structural validation accepts the bytes. Do not merely set an already-encrypted `c1` to zero: that discards `c1*s` and changes the message.

For the intentionally insecure Fast *simulation* backend, a suitable degree-one input representation is `c0 = encoded plaintext`, `c1 = 0`, correct public CKKS metadata and logical Level/Scale. This matches the previously accepted P93 direct-input diagnostic. Existing `rlwe.NewDecryptor` computes `c0 + c1*s` and thus gives `c0` for an output whose c1 is **actually zero**, irrespective of the sampled secret key. The internal effective secret may be zero without requiring every SecretKey object to be all-zero; however, any nonzero Fast output c1 requires an explicit semantics decision and must not be silently ignored.

The correct next implementation boundary is **before ciphertext generation** (CKKS-specific internal Fast mode in the public NewEncryptor/EncryptNew path), not a destructive conversion from normally encrypted ciphertext at Bootstrap entry. Preserve **the same frontend source and API signatures** in both library builds, and do not silently change N/Q/P/Scale/E. Prevent Fast simulation activation in BFV/BGV/ordinary RLWE consumers. The exact CKKS-only dispatch should be based on current source evidence and tested for non-CKKS isolation, not an unguarded global `rlwe.Encryptor` rewrite.

## Remaining limits

- The current audit frontend explicitly asserts nonzero `c1`, which is an audit instrumentation assumption and must **not** be used as the next identical-frontend acceptance program once Fast encryption intentionally emits `c1=0`. Create a separate shared frontend with no assumption about `c1` and no Fast-only APIs.
- Public-key EncryptNew, independent CKKS arithmetic chains and post-Bootstrap DecryptNew have not been validated.
- No speedup, security-equivalence, universal accuracy or complete drop-in claims supported.

Next action: bounded CKKS-scoped Fast zero-secret encryption bridge, focused no-Bootstrap tests first, and (only after preflight) one paired Standard/Fast public Bootstrap through a single unchanged frontend, without numerical cherry-picking.
