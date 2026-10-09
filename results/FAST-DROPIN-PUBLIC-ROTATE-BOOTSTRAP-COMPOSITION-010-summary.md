# FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010

**Classification:** `FAST_DROPIN_COMPOSITION_BLOCKED`

**Handoff:** `NEEDS_WEB_REVIEW`

## Scope and provenance

This was a bounded public API composition check only. No Bootstrap, benchmark, warmup, retry, LogN16/CNN run, or Secondary source edit was performed. Primary began at `0c49274a0242f818889a7e0784aecec6f2544691`. The pinned Standard checkout was clean at `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Secondary `fast-qprefix` was clean at `00ac70ba136d190fa31bbb26c2f51d003a221634`.

The frozen LogN13 E32 config hash is `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`. The deterministic 4096-slot complex input hash is `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`. Both backends derive the same rotation element, `5`, for `k=1`, and the element is in the Bootstrap evaluation-key plan. Actual bootstrap Q/P prime-bit arrays and hashes are in the adjacent evidence JSON.

## Preflight results

Fast preflight passed on the final shared frontend source SHA `d6b6e4d83d1e2a06b5c9be4abfc4ce7bd9ef0b03cca96d62eccb8b092262d02d`. Its public `EncryptNew` input and `RotateNew` output were both degree 1, Level 0, log₂Scale 45, NTT/non-Montgomery, with one fully backed active q0 row in each component. Both had zero nonzero c1 coefficients. The Rotate instrumentation observed zero `GetGaloisKey` and zero `GetGaloisKeysList` calls. The generated Bootstrap keyset had 30 Fast-layout Galois keys and included the requested rotation.

Fast immediate decrypt RMSE was `7.449975091875706e-13` (max complex error `2.101955261562136e-12`, SNR `208.78410083248798 dB`). Rotated output versus the cleartext slot oracle had RMSE `7.44997342660252e-13` (max complex error `2.1019564094557047e-12`, SNR `208.78410277402207 dB`). The decoded rotated-array SHA-256 is recorded in the JSON evidence.

Standard used genuine secret-key `EncryptNew`; c1 was nonzero, and checks before Rotate confirmed the requested Galois element was included in `bootstrapping.Parameters.GenEvaluationKeys(sk)`. The first failure was the ordinary public `ckks.NewEvaluator(residual, bootstrapEvaluationKeys).RotateNew(level0Ciphertext, 1)` call. Standard panicked with:

```text
eval.GadgetProductLazy: ctQP.LevelP()=-1 < gadgetCt.LevelP()=4
```

The Standard Bootstrap rotation key carries P level 4, while the residual evaluator/ciphertext has no P basis (`LevelP=-1`). The process panicked before it could serialize a Standard JSON record; the exact frontend SHA for that run was therefore not captured. The subsequent reflection-only harness correction used to inspect Fast key layout did not alter key generation or Rotate logic, but means the final Fast preflight SHA cannot be claimed as identical to the unrecorded Standard-run source hash.

## Stop decision

Per the spec, the Standard preflight failure stops the composition gate. **Bootstrap calls: Standard 0, Fast 0, total 0.** No key-basis workaround, alternate rotation key, parameter change, or fallback was attempted. There is no Standard/Fast numerical comparison and no Bootstrap output. The observed P-basis incompatibility needs Web architecture review before selecting a different public key plan or changing the contract.

The runner package compiled successfully under both pinned dependency workspaces after the harness correction. It has no package-local Go tests. Compact machine-readable evidence is in `results/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010-evidence.json`; the temporary Fast raw preflight, including vectors, remains outside the repository.
