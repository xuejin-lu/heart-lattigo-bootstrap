# Batch 018 execution journal

## Startup and pinned inputs

- Primary synchronized `main` to `origin/main=9baab4f602cd83f0beef2aa3bf572aa07465c0fe`; clean at startup.
- Secondary synchronized `fast-qprefix` to `origin/fast-qprefix=2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; clean, read-only for this batch.
- Genuine Standard checkout: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean and read-only.
- Frozen config: `configs/bootstrap_config.logN13.json`, SHA-256 `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`.
- Batch-wide Bootstrap budget: Standard `0/1`, Fast `0/1` consumed at this checkpoint.

## Checkpoint A — read-only profile, key-domain, depth and capacity audit

**Status: PASS; zero Bootstrap calls.** The residual parameter set alone is only `MaxLevel=1`, so it is not a legal Level-5 source. The existing frozen profile nevertheless defines a legitimate full-chain source: `btpParams.BootstrappingParameters` has `MaxLevel=16`, E32, LogN13 and the same actual Q/P chain used by the canonical 017 evaluator. `NewParametersFromLiteral` copies the residual Q as the exact prefix before appending circuit primes. Public `GenEvaluationKeys(sk)` extends the same residual secret into the full bootstrap Q/P domain; Standard source uses `ExtendBasisSmallNormAndCenterNTTMontgomery`, and Fast's public dispatcher reaches `GenFastEvaluationKeys`, which applies the corresponding basis extension. The public arithmetic evaluator can therefore be bound to the full Bootstrap parameters and returned extended key, while the exact q0 key/domain is preserved at the eventual residual boundary.

The bounded path selected for B is:

`full-profile Level-5 EncryptNew(A,B) → public AddNew → public RotateNew(1) → public DropLevelNew(5 levels) → Level-0 native DecryptNew/Decode → public E32 Bootstrap`.

`AddNew` and `RotateNew` execute above the four-row Q-prefix cap at logical Level 5. Fast public Add/Rotate dispatch to the existing shared Q-prefix cores, whose output authority is q0..q3; q4/q5 and higher rows remain dormant. `DropLevelNew` is the existing ordinary CKKS projection (no Scale change or division). The output is level 0, backed only by q0, and has the same exact Scale `2^45` accepted by the 017 public Bootstrap input. Standard `Bootstrap` uses that same q0 and full keyset; Fast's existing public Bootstrap gate allows the resulting Level 0 input (its residual maximum is 1). No parameter, profile, default Scale, key-generation API, or frontend-specific Fast behavior is introduced.

The pinned Fast public-dispatch source was checked read-only at Secondary `2d6145d7`: `schemes/ckks/evaluator.go` routes eligible zero-secret `AddNew` into `addSubFastCKKSZeroSecret`, which applies the shared `fastcore.AddSubWorkspace`; `RotateNew` allocates compact output and routes through `rotateFastCKKSZeroSecret`, which applies the shared `fastcore.AutomorphismWorkspace`. The Add fallback explicitly requires full active rows and rejects compact input; it is not a silent materialization path. The paired runner additionally records row lengths/hashes and invokes the existing Fast Q-prefix capacity observer after each relevant public operation. The Standard build uses the same frontend APIs with its ordinary evaluator.

The conservative independent coefficient bound uses the previously accepted 016 encoder bound `ceil(2*max(|slot|)*Scale+1)`, then Add's `B'=B_A+B_B`; the automorphism/rotation bound is unchanged. For the canonical 017 deterministic 4096-slot inputs at Scale `2^45`:

| Quantity | Bound / capacity |
|---|---:|
| `B_A` | `2,348,557,866,436` |
| `B_B` | `1,804,822,278,899` |
| Add and Rotate `B` | `4,153,380,145,335` |
| Exact target q0 | `36,028,797,018,652,673` |
| Strict target test | `2B = 8,306,760,290,670 < q0` — pass |
| Level-5 authoritative q0..q3 product | `5,986,308,565,615,587,353,347,023,386,369,277,282,933,624,144,412,673` |

Thus the same centered integer lift fits the target q0; the Level-5→0 DropLevel does not rely on dormant q4+ or on a diagnostic observer alone. This independent `B` bound applies to Fast's zero-secret encoded c0 representation; it is **not** asserted as a bound on either component of genuine randomized Standard RLWE ciphertexts. Standard is checked through native public decryption/decoding and its cleartext oracle. Scale stays exactly `2^45`, satisfying the frozen Level-0 Bootstrap scale contract. Native public decryption is performed only at the supported final Level 0 (and post-Bootstrap residual Level 1), not on compact high-Level Fast ciphertexts. Existing 017 evidence already demonstrated that ordinary native `DecryptNew/Decode` and public E32 Bootstrap accept the exact canonical q0/Scale and key lifecycle at the residual boundary.

The separate accepted 016 profile is not substituted: it has six Q primes `[55,39,40,39,39,39]`, one P prime `[60]`, logical Level 5, and 16 slots; 017 has the frozen generated 17-prime Bootstrap Q chain, five P primes, residual Level 1 / Bootstrap Level 16, and 4096 slots. This audit composes within 017's own full Bootstrap Q domain; it does not transplant 016's profile or its measurement-only decoder.

### Checkpoint log

| Checkpoint | Status | Bootstrap calls | Next action |
|---|---|---:|---|
| A — read-only exact profile/Level/Scale/CRT feasibility | PASS | 0 | Build and run matched zero-Bootstrap B preflight |
| B — matched public high-Level compact preflight | PASS | 0 | Pair validator passed; only then proceed to the one-shot Bootstrap pair |
| C — one-shot public Bootstrap pair | PASS | Standard 1/1, Fast 1/1 | Both native output oracles and fixed paired error gate passed; no retries |
| D — artifacts, self-review and safe Primary handoff | PASS | Standard 1/1, Fast 1/1 | Batch complete; deliver for independent Web review |

### Runner self-review correction before accepted B evidence

The first paired runner processes reached the explicit Bootstrap wait gate, but the pair validator incorrectly required the full-profile Level-5 `EncryptNew` inputs themselves to have compact backing. The pinned Fast public encryption legitimately produced six active logical rows; the authorized Q-prefix operations are the step that must leave q4+ dormant. Both processes were explicitly aborted at the gate before any Bootstrap call (`Standard=0/1`, `Fast=0/1`). The validator was corrected to require the existing observer and zero-c1 contract on the encryption inputs, and to require compact backing only on the public Add/Rotate/DropLevel outputs. A focused regression test covers this distinction. These aborted raw runs are not accepted measurement evidence; the matched B preflight will be rerun from the corrected clean commit.

### Checkpoint B — accepted matched preflight

**Status: PASS; zero Bootstrap calls.** Both runs used the same clean Primary code commit `86af97092ba666a4a16ab2e7c4dde861ee669a96`, config/workload/frontend/profile/QP hashes, canonical 4096-slot A/B inputs, full E32 key lifecycle, and ordinary public CKKS API chain. Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and Fast `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` were clean. Both residual Level-0 native decrypt/decode cleartext oracles passed the fixed max-error `1e-6` gate. Fast-vs-Standard pre-Bootstrap comparison: complex RMSE `1.5656982925854002e-11`, max `7.13774172628691e-11`.

Fast public `EncryptNew` inputs at full-profile Level 5 have six logical rows; the operation outputs prove the bounded policy: Add and Rotate remain Level 5 / Scale `2^45`, authoritative rows 4, with q4+ backing lengths zero; `DropLevelNew(5)` is Level 0 / unchanged Scale, with one q0 row. Authoritative c1 stayed zero. The existing Fast observer ran at each checkpoint. Independent bounds: `B_A=2348557866436`, `B_B=1804822278899`, and post-Add/Rotate `B=4153380145335`; `2B=8306760290670 < q0=36028797018652673`. The Level-5 q0..q3 prefix product was `5986308565615587353347023386369277282933624144412673`. The independent coefficient bound is for Fast's zero-secret c0 representation only; no such component bound is claimed for Standard ciphertexts.

### Checkpoint C — one-shot public Bootstrap pair

**Status: PASS.** The accepted B processes were held at their local one-shot gates. Standard was called once and passed native post-Bootstrap decryption at Level 1 / Scale `2^45`: cleartext RMSE `4.880485998172208e-9`, max `1.8458745857066216e-8`, SNR `134.4577446217342 dB`. Only after that success, Fast was called once on its in-memory B output and passed: RMSE `1.1829277433501467e-9`, max `5.764906238946737e-8`, SNR `146.76784169515037 dB`. Both outputs had 4096 decoded slots; Fast retained two authoritative rows and zero c1.

Final Fast-vs-Standard Bootstrap output comparison: complex RMSE `4.963927268857064e-9`, max `3.93386413412319e-8`, pass under the unchanged `1e-6` max-error gate. Total actual Bootstrap calls: Standard `1/1`, Fast `1/1`; no benchmark, sweep, LogN16, retry, or Secondary edit was performed. Compact paired evidence is in `FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018-evidence.json`.

The runner provenance commit is `86af97092ba666a4a16ab2e7c4dde861ee669a96`; Primary was clean during both accepted backend runs. Secondary remained clean and unchanged at `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`. Final task status: `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`; artifact commit/push state is included in the completion report.
