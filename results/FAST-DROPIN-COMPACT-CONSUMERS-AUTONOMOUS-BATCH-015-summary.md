# FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015

**Result:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`

**Scope:** compact public Add → MulRelin → Rescale → Rotate integration; no Bootstrap, benchmark, or LogN16 execution.

## Review findings and minimal repairs

1. **RescaleTo no-op / row authority.** The ordinary public CKKS `Rescale` and `RescaleTo` wrappers pass `QPrefixWidth(Level)`, i.e. the fixed production policy `w_Q(Level)=min(Level+1,4)`. A narrower width is not produced by the supported ordinary production path, so the review's Level-3 / one-row example is not evidence of a public production arithmetic bug. However, the explicitly row-parameterized Fast APIs and `fastcore.RescaleCore` promise exact source-row authority. Their no-op path previously copied policy-width rows instead of the caller-authorized rows; the actual Rescale path also left newly allocated rows outside a narrow source authority looking materialized. The minimal repair now copies/retains only `sourceRows` (or `min(sourceRows,target policy width)` after a true Rescale) and clears backing above that authority. No adaptive limb-selection algorithm was added. Level-0 true Rescale remains rejected; a no-op at Level 3 preserves Level and Scale.
2. **`MulRelinNew` malformed inputs.** The Fast ciphertext-pair path now validates nil, degree-one component count, metadata, logical row shape, configured Level, and fixed-prefix backing before calling `Level()` or allocating compact output. Negative tests cover missing components, nil metadata, empty polynomial/q0 storage, and inconsistent component rows. `MulRelin` rejection tests confirm both inputs and a pre-existing output remain unchanged.
3. **Stale Rescale/Q-prefix comments.** Rescale comments now describe `w_Q(Level)=min(Level+1,4)` and distinguish fixed production policy from explicit source-row authority; stale “full/complete active-Q” wording at the affected public Rotate and MulRelin boundaries was corrected as well.

## Reuse and dispatch evidence

- Public Fast `ckks.Evaluator.MulRelin` / `MulRelinNew` and explicit Fast MulRelin use the same import-neutral `fastcore.MulWorkspace`; instrumentation tests assert shared-core dispatch. The supported public pair remains degree-one, NTT, non-Montgomery, canonical prefix rows, and zero `c1`; output has degree one, product Scale, and the minimum logical Level. It does not use a relinearization key, KeySwitch, GadgetProduct, or full-RNS fallback.
- Public CKKS and explicit Fast Rescale now call the same `fastcore.RescaleWorkspace`, preserving the existing bounded centered-CRT implementation, logical top-q divisor, per-transition capacity checks, and transactional output.
- Public Rotate uses the shared import-neutral automorphism core with fixed Q-prefix authority. Existing Add/Sub shared-core behavior remains intact. Level-5 tests keep q4/q5 dormant through the composed consumers; this is structural-only evidence, not a numeric claim.

## Matched numerical result

Same Primary runner, deterministic plaintext inputs, LogN13 / Q bits `[55,39,40,39]` / P bits `[60]` / scale `2^45` / `LogSlots=4`, and native Standard keygen-encrypt-decrypt lifecycle were used for both backends. Numerical decoding is limited to the accepted Levels 1 and 3. The separate Level-5 fixture is structural-only.

| Checkpoint | Level | log2 Scale (both) | Standard plaintext RMSE | Fast plaintext RMSE | Fast-vs-Standard RMSE / max |
|---|---:|---:|---:|---:|---:|
| Add | 1 | 45 | 8.5701e-13 | 6.3661e-14 | 0 / 0 |
| MulRelin | 1 | 90 | 1.1877e-13 | 1.2108e-14 | 0 / 0 |
| Rescale | 0 | 51.00000021 | 1.1914e-13 | 1.2091e-14 | 0 / 0 |
| Rotate | 0 | 51.00000021 | 1.2218e-13 | 1.2091e-14 | 0 / 0 |
| Add | 3 | 45 | 6.1461e-13 | 6.3661e-14 | 0 / 0 |
| MulRelin | 3 | 90 | 1.1789e-13 | 1.2108e-14 | 0 / 0 |
| Rescale | 2 | 50.99999944 | 1.1760e-13 | 1.2038e-14 | 0 / 0 |
| Rotate | 2 | 50.99999944 | 1.1719e-13 | 1.2037e-14 | 0 / 0 |

All eight checkpoints passed the runner's state and cleartext-oracle gates; Standard/Fast Level, degree, and Scale progression match. The paired decoded values were identical at the recorded float precision. Fast direct relinearization/Galois retrieval counts were both zero. `GetGaloisKeysList` was observed once during evaluator construction (the pinned RLWE evaluator enumerates the list when initialized); it was not a per-Rotate key retrieval. Bootstrap call count: zero.

### Provenance

- Primary source revision during both runs: `d097a3f638431ed862e8f0da34956480e348c937`; Primary was dirty because this task's runner/journal were not yet committed. The runner source SHA-256 is recorded in the evidence artifact. This is recorded honestly, not represented as a clean Primary run.
- Genuine Standard: clean detached checkout `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast Secondary: clean, pushed `fast-qprefix` commit `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- Profile SHA-256: `f1d7944ea3febd9731cfd846af71c8596089893d014b729e489fc694d27f7ebc`.
- Effective Q/P SHA-256: `41c103fb3338c8fca82ab0c5b2e4dfd17cf1affc6d7295e082a6ba4a6e372324`.
- Deterministic input SHA-256: `296447e0602fcaec7839fd302ec783dd0994418aad18757c20cfe9d38323cd24`.

## Validation

- Focused Secondary tests passed in `schemes/ckks/internal/fastcore`, `schemes/ckks`, and `schemes/ckks/fast`, including row-capacity/CRT oracles, public and explicit shared-core dispatch, malformed-input transactionality, the compact arithmetic chain, and the Level-5 structural chain. The exact-name allowlist excluded LogN16 tests.
- `GOWORK=off go vet ./schemes/ckks/internal/fastcore ./schemes/ckks ./schemes/ckks/fast` passed.
- `git diff --check` passed before commit.
- Matched runner: `GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go run ./tools/fast-dropin-compact-consumers-batch-015` (Standard) and `GOWORK=off go run ./tools/fast-dropin-compact-consumers-batch-015` (Fast); combined evidence is in `FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-evidence.json`.
- No Bootstrap, benchmark, or LogN16 run was performed in this batch.

## Commits and handoff

- Secondary `fast-qprefix`: `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, normally pushed; worktree clean.
- Primary runner and result artifacts are the current task's authorized deliverables; their commit SHA and push state are reported in the final handoff.
