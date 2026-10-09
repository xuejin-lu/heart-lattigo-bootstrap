# FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014

**Classification:** `API_ADAPTER` — compact Add/Sub bridge passes; first downstream compact consumer remains unsupported.
**Batch status:** `BATCH_BLOCKED_NEEDS_WEB_REVIEW`
**Bootstrap / benchmark / LogN16:** 0 / 0 / 0

## Outcome

The ordinary public `ckks.Evaluator.AddNew` / `SubNew` ciphertext path now dispatches eligible Fast zero-secret ciphertexts to the exact same import-neutral `fastcore.AddSubCore` used by the existing explicit Fast Q-prefix Add/Sub API. The public signatures and shared frontend are unchanged. Ordinary Fast `EncryptNew` supplies the zero-c1 inputs; the adapter does not use Fast constructors, hand-patch ciphertexts, or change Standard arithmetic.

At the frozen LogN13 profile, Standard and Fast Add/Sub plus Add→Sub recovery passed decoded plaintext oracles at Levels 1 and 3. A Level-5 structural test proves compact Add/Sub composition preserves logical Level 5 while materializing only q0..q3. The first downstream public consumer, `MulRelin`, rejects that compact ciphertext at the first absent active row q4. Separate Rescale and Rotate controls also reject q4. This batch therefore demonstrates a real public Fast primitive bridge, **not** complete compact-Q CKKS composition or drop-in Bootstrap readiness. Work stopped at the first boundary; no new consumer implementation or mathematical rule was guessed.

## Architecture and reuse map

| Public/API boundary | Observed implementation |
|---|---|
| `ckks.NewEvaluator(params, evk)` | Same public type/signature; Fast capability is detected through the existing parameter marker without importing `ckks/fast` into parent `ckks`. |
| Public `Add` / `Sub` ciphertext pairs | Eligible Standard-ring, configured-NTT, equal-scale/batching zero-secret operands dispatch through shared `fastcore.AddSubCore`; exact prefix rows are validated before writes. Nonzero-c1 full-backed inputs retain the generic CKKS path; compact inputs are rejected before any generic full-Q read. |
| Public `AddNew` / `SubNew` | Fast-eligible inputs receive compact output backing directly while preserving logical Level and ordinary CKKS metadata. Other accepted inputs retain the normal full-backed allocator. |
| Explicit Fast `AddQPrefixRows` / `SubQPrefixRows` | Calls the same shared core; no second Add/Sub arithmetic implementation remains. Fast compact `NewCiphertext`/`Resize` now share the neutral compact allocator. |
| Public `MulRelin`, `Rescale`, `Rotate` after compact Level 5 | Not end-to-end compatible yet. Each stops at q4 before using missing/dormant rows; first boundary is `MulRelin`. |

The import-neutral layer avoids the existing `schemes/ckks/fast` → parent `schemes/ckks` cycle. The public `ckks` evaluator depends only on `schemes/ckks/internal/fastcore`; the separate genuine Standard checkout contains no Fast capability marker and retains native Standard behavior.

## Matched provenance

- Primary source base during measurements: `2c08ca374e4b868ac00fc29e1ae548424189d8e2`; dirty only from this batch's runner, tests, journal, and evidence at measurement time.
- Fast Secondary: `0c92e7ed071a1be6eb258474811db470e4463df7`, `fast-qprefix`, clean and pushed to `origin/fast-qprefix`.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, pinned separate checkout, clean.
- Runtime: `go1.26.4`, `darwin/arm64`.
- Frozen parameters: LogN 13, LogQ `[55,39,40,39]`, LogP `[60]`, default scale `2^45`, 16 slots; Levels `[1,3]`. The profile has exactly four Q primes, so no Level above the four-row cap is reachable without changing the frozen workload.
- Profile SHA-256: `4ac3a9eb4a1662395dbda550ca9469f89989ba7c1ee3ca26148f20a593e2a78b`.
- Effective Q/P SHA-256: `41c103fb3338c8fca82ab0c5b2e4dfd17cf1affc6d7295e082a6ba4a6e372324`.
- Deterministic input pair SHA-256: `296447e0602fcaec7839fd302ec783dd0994418aad18757c20cfe9d38323cd24`.
- Identical runner SHA-256: `3da11cd491fbc0da8887d0bd3c2f44f9e3f9aa076074ffd3d612a0b034e6ea93`; test SHA-256: `52884f33138df8ddde84f73056c8b576b390c25ec3bdd83f3a2a7e02e28b5742`.
- Lifecycle on both dependencies: ordinary CKKS key generation, plaintext encoding, `rlwe.NewEncryptor(...).EncryptNew`, public `ckks.NewEvaluator`, and ordinary `DecryptNew`/Decode. Standard c1 was nonzero; Fast EncryptNew and successful Add/Sub results had zero c1. No evaluation-key lookup occurred.

## Numerical oracle matrix

Each entry is Fast-vs-Standard decoded complex RMSE / max complex difference. All compared checkpoints retained the same Level, Degree 1, and Scale `2^45`; per-component active rows were `[2,2]` at Level 1 and `[4,4]` at Level 3.

| Checkpoint | Level | Standard plaintext RMSE | Fast plaintext RMSE | Fast-vs-Standard RMSE / max |
|---|---:|---:|---:|---:|
| AddNew | 1 | 6.2741e-13 | 6.3661e-14 | 6.2850e-13 / 1.3799e-12 |
| SubNew | 1 | 7.8883e-13 | 6.5412e-14 | 7.8815e-13 / 1.0729e-12 |
| Add→Sub recovery | 1 | 5.0157e-13 | 4.3074e-14 | 4.9799e-13 / 9.3150e-13 |
| AddNew | 3 | 5.3276e-13 | 6.3661e-14 | 5.1160e-13 / 9.6170e-13 |
| SubNew | 3 | 8.0037e-13 | 6.5412e-14 | 8.1784e-13 / 1.6086e-12 |
| Add→Sub recovery | 3 | 5.2792e-13 | 4.3074e-14 | 5.2407e-13 / 8.6233e-13 |

The Level-5 compact-path test is structural-only: it validates prefix row contents, preservation of rows above the prefix, unchanged logical Level, and fail-closed consumers. It intentionally makes no ordinary decoded numerical claim above the four-row prefix.

## First unsupported compact consumer

At logical Level 5, public Fast Add returns component backing for q0..q3 only; q4 is the first absent active row and q5 is also dormant. The next public operation in the tested chain is `ckks.Evaluator.MulRelin` at `schemes/ckks/evaluator.go:840-849`. Its Fast branch enters `mulRelinFastCKKSZeroSecret`; `validateFastCKKSZeroSecretMulRows` at `schemes/ckks/evaluator_fast_zero_secret.go:117-139` requires rows through logical Level 5 and reports component 0 q4 as not fully materialized before arithmetic, key lookup, or output mutation. This is an `API_ADAPTER` boundary: the existing public wrapper does not route the compact object into a compatible Q-prefix product path.

Rescale and Rotate were separately checked against the same compact object and also fail closed at q4. They are not reported as sequential successes after MulRelin. No q4/q5 materialization, full-RNS fallback, Bootstrap, or consumer repair was attempted.

## Validation and handoff

Passed for both pinned dependency selections:

```text
GOWORK=off go test ./tools/fast-dropin-fast-engine-adapter-batch-014 -count=1
GOWORK=off go vet ./tools/fast-dropin-fast-engine-adapter-batch-014
GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go test ./tools/fast-dropin-fast-engine-adapter-batch-014 -count=1
GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go vet ./tools/fast-dropin-fast-engine-adapter-batch-014
GOWORK=off go test ./schemes/ckks/internal/fastcore ./schemes/ckks ./schemes/ckks/fast -count=1
GOWORK=off go vet ./schemes/ckks/internal/fastcore ./schemes/ckks ./schemes/ckks/fast
git diff --check
```

The focused Secondary tests instrument the shared core to prove public Fast AddNew/SubNew dispatch at Levels 1 and 3, check decoded plaintexts, and prove Level-5 compaction plus transactional rejection by MulRelin/Rescale/Rotate. The explicit Fast Add/Sub tests continue to pass. The paired runner contains no Bootstrap call.

Compact combined evidence: `results/FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014-evidence.json` (19,867 bytes; aggregate metrics only, no decoded vectors). Detailed progress and exact command provenance: `results/FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014-journal.md`.

**Web review request:** decide the next bounded adapter for the first compact consumer (`MulRelin`) or whether the current public Add/Sub bridge should be accepted as an isolated integration milestone. This batch does not authorize a follow-on implementation.
