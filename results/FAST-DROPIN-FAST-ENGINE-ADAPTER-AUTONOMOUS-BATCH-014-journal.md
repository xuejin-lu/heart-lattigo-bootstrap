# FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014 journal

**State:** BATCH_BLOCKED_NEEDS_WEB_REVIEW
**Primary base:** 2c08ca374e4b868ac00fc29e1ae548424189d8e2
**Fast Secondary base:** 9489925c803be09e49a840e6828b005cdeb1faca (clean)
**Fast Secondary result:** 0c92e7ed071a1be6eb258474811db470e4463df7 (pushed, clean)
**Genuine Standard pin:** 5dbffbdea05394de2ca3a432ed5318aa832e3f40 (clean)
**Cost budget:** 0 Bootstrap, 0 benchmarks, 0 LogN16.

## Checkpoint A — source-backed adapter map

- The unchanged public constructor is `ckks.NewEvaluator(parameters, evk) *ckks.Evaluator`; `ckks.Evaluator` embeds `*rlwe.Evaluator`. Fast capability is selected by the parameter-provider marker `FastCKKSZeroSecretSimulation()`.
- Public signatures are `Add(op0, op1, opOut)`, `Sub(op0, op1, opOut)`, `MulRelin(op0, op1, opOut)`, `Rescale(op0, opOut)`, and `Rotate(op0, k, opOut)`, with the existing `rlwe.Ciphertext` / `rlwe.Operand` types. Explicit Fast signatures include `Add/Sub(op0, op1, opOut)` and bounded `AddQPrefixRows/SubQPrefixRows(..., rows)`; `fast.NewEvaluator(params)` returns `*fast.Evaluator`.
- Public `ckks.NewCiphertext` and RLWE allocation make full active-Q backing. Explicit `fast.NewCiphertext` preserves logical Level but allocates only `min(Level+1,4)` Q rows; `fast.Resize` contracts higher rows to nil. The existing Fast add/sub kernel validates metadata, scale/domain, dimensions and exactly the requested prefix before arithmetic, and preserves rows above that prefix as non-authoritative.
- Ordinary `rlwe.EncryptNew` detects the Fast-only parameter capability and copies the encoded plaintext to c0 with c1 zero; the genuine Standard pin uses its ordinary encryption path. Ordinary `DecryptNew` is the existing RLWE decryptor and is not a compact-row decoder.
- `schemes/ckks/fast` imports parent `schemes/ckks`, so parent CKKS cannot import the Fast child. Both already import/use `schemes/ckks/internal/fastcore` for the shared automorphism. Directly extracting the existing add/sub row kernel into this import-neutral package allows both evaluator facades to call the same implementation without changing public signatures or duplicating arithmetic.
- **B selection:** use public ciphertext/ciphertext `Add` as the first adapter bridge. It is the smallest existing Q-prefix kernel with direct modular-oracle tests; `Sub` can exercise the same shared core in checkpoint C. Keep unsupported operand forms fail-closed at the compact boundary rather than silently entering the full-Q implementation.
- **Level boundary:** the frozen numerical profile is LogN13 with exactly four Q primes, so no logical Level above the four-row cap is reachable without changing the profile. A separate Level-5 structural test exercises capped compaction and downstream public-consumer rejection; it deliberately makes no decoded numerical claim because the ordinary full-active-Q decoder would require missing rows.

## Checkpoint progress

| Checkpoint | State | Commit / verification | Notes |
|---|---|---|---|
| A — source map and selection | Complete | Primary base `2c08ca3`; Fast base `9489925`; Standard `5dbffbde` | No implementation edits during audit. |
| B — first public bridge | Pass | Fast `0c92e7ed`; affected Secondary tests/vet pass | Public `AddNew`/`SubNew` dispatch to the same import-neutral `fastcore.AddSubCore` used by explicit Fast Add/Sub. Levels 1/3 decode against the plaintext oracle; Level-5 compact shape is tested structurally. |
| C — narrow composition | Partial; stopped at first boundary | Paired numerical evidence in `...-evidence.json` | Add→Sub recovery passes at Levels 1/3. At Level 5, first downstream public consumer `MulRelin` rejects compact input at missing active row q4; separate Rescale and Rotate controls also reject q4. Classification: `API_ADAPTER`. No consumer repair or math change attempted. |
| D — evidence and handoff | Complete | Primary artifacts staged for commit; Secondary `0c92e7ed` pushed | Fast/Standard runner and vet, Secondary affected-package tests/vet, diff checks and self-review completed. Batch remains blocked pending Web decision on the first compact consumer boundary. |

## Paired run provenance and numerical evidence

- Primary source base during both runs: `2c08ca374e4b868ac00fc29e1ae548424189d8e2`, dirty only with this batch's runner/test/journal/evidence changes. Shared runner SHA-256 `3da11cd491fbc0da8887d0bd3c2f44f9e3f9aa076074ffd3d612a0b034e6ea93`; shared test SHA-256 `52884f33138df8ddde84f73056c8b576b390c25ec3bdd83f3a2a7e02e28b5742`.
- Fast: `0c92e7ed071a1be6eb258474811db470e4463df7`, `fast-qprefix`, clean. Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean. Go `go1.26.4`, `darwin/arm64`.
- Frozen profile: LogN13, LogQ `[55,39,40,39]`, LogP `[60]`, scale `2^45`, 16 slots, Levels 1 and 3. Profile SHA-256 `4ac3a9eb4a1662395dbda550ca9469f89989ba7c1ee3ca26148f20a593e2a78b`; effective Q/P SHA-256 `41c103fb3338c8fca82ab0c5b2e4dfd17cf1affc6d7295e082a6ba4a6e372324`; deterministic plaintext pair SHA-256 `296447e0602fcaec7839fd302ec783dd0994418aad18757c20cfe9d38323cd24`.
- Both builds use the same public runner and `ckks.NewEvaluator`; both generate ordinary keys and encrypt with `rlwe.NewEncryptor(...).EncryptNew`. Standard ciphertexts have nonzero c1; Fast inputs and Add/Sub outputs have zero c1. All Level/Scale/degree and active-row checks passed. No evaluation-key lookup or Bootstrap was used.
- Paired Fast-vs-Standard decoded metrics (complex RMSE / maximum complex error): `AddNew L1` 6.284978314e-13 / 1.379871789e-12; `SubNew L1` 7.881538765e-13 / 1.072852153e-12; Add→Sub recovery L1 4.979864223e-13 / 9.314964055e-13; `AddNew L3` 5.115951271e-13 / 9.616961147e-13; `SubNew L3` 8.178353689e-13 / 1.608593107e-12; Add→Sub recovery L3 5.240723238e-13 / 8.623256543e-13. All outputs retained Scale `2^45`, Degree 1, Levels 1/3; backing was `[2,2]` and `[4,4]` rows per component respectively.

## First compact downstream boundary

The Level-5 public Add output preserves logical Level 5 while backing only q0..q3 for each component (`[4,4]`); q4 is the first absent row, q5 also remains dormant. The next tested public consumer in the sequence is `ckks.Evaluator.MulRelin` (`schemes/ckks/evaluator.go:840-849`), which dispatches to `mulRelinFastCKKSZeroSecret`; its active-row validator (`schemes/ckks/evaluator_fast_zero_secret.go:117-139`) requires materialized rows through logical Level 5 and returns at q4 before arithmetic/key lookup/output mutation. Separate Rescale and Rotate boundary controls likewise reject q4. This is classified `API_ADAPTER`: the current public CKKS wrapper does not route compact authority into a compatible Q-prefix product/rescale/rotation path. The task stops here; no new numerical semantics or repair is inferred.
