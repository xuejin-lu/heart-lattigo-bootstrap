# FAST-ZERO-SECRET-E-METRICS-RECOVERY-002

**Classification:** `FAST_E_METRICS_RECOVERED_PENDING_WEB_REVIEW`
**Handoff:** `READY_FOR_WEB_REVIEW`

## Provenance

- Experiment Primary commit: `fb0d2141287ede644ab33166d4f1e065b4f5a298` (clean `main`).
- Fast backend: Secondary `fast-qprefix` commit `463d494627b2e9e2bfac51aefe7f4ecb3493b68e` (clean before and after; unchanged).
- Standard reference: clean, unmodified Standard commit `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, genuine E=32 lifecycle (`GenSecretKeyNew`, native `EncryptNew` with nonzero `c1`, one public Bootstrap, matching-secret `DecryptNew`/decode). Temporary reference artifact SHA-256: `6157a3692b0b6ab84611c47cc53c807c064774bffe23d20cd91aeaae3ce98330`.
- Reference validation: schema `fast-zero-secret-e-passthrough-standard-reference.v1`, `standard_dirty=false`, 4,096 finite decoded slots, expected output Level 1 / Scale `2^45` / degree 1 / N=8192 / 2 components / NTT=true / Montgomery=false.
- Profile hashes (all matched the diagnostic's pinned constants and reference): config `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; input `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; effective parameters `f151442a4e08e1ebf8b7bb515fdd075a3748298bee1f3e2b990a6ad99aeca693`; Q primes `30066778ef1caea959b3357b0a70ff792fe8582ba9545f6b53c8df2a1a6cb568`; P primes `70451211d27cd1e6bf092aa3c0752a63147c34842211dcd70d64d50755d318f7`.
- Runtime: Go `go1.26.4`, `darwin/arm64`. Fast diagnostic input was `diagnostic_plaintext_like_c1_zero`; this is diagnostic-only and is not evidence of a normal encrypted Fast lifecycle or security equivalence.

## Captured Fast run

Exactly one tagged diagnostic invocation completed: one public Fast Bootstrap at E=0 and one at E=32 (2 total; no Standard Bootstrap was run). Both returned and decoded 4,096 finite slots. Output metadata for each: Level 1, Scale `2^45`, degree 1, N=8192, 2 components, NTT=true, Montgomery=false. Dense/sparse conversion keys were absent at E=0 and present at E=32.

| Comparison | Complex RMSE | Max complex | Max real | Max imag | SNR |
|---|---:|---:|---:|---:|---:|
| Fast E=0 vs original | `1.1907546782784477e-9` | `5.805793207821362e-8` | `1.6480985430555872e-9` | `5.8057504432468265e-8` | `144.7107513180764 dB` |
| Fast E=32 vs original | `1.1907546782784477e-9` | `5.805793207821362e-8` | `1.6480985430555872e-9` | `5.8057504432468265e-8` | `144.7107513180764 dB` |
| Fast E=32 vs Fast E=0 | `0` | `0` | `0` | `0` | `infinite_zero_error` (numeric `snr_db` intentionally absent) |
| Fast E=32 vs genuine Standard E=32 | `4.914687044299207e-9` | `5.030262706326469e-8` | `1.2665915774534575e-8` | `4.976561657384082e-8` | `132.3972798514185 dB` |
| Standard E=32 vs original (existing reference) | `4.8050725021293596e-9` | `1.445642382010722e-8` | `1.2702520951757279e-8` | `1.4067787003010923e-8` | `132.59319834768257 dB` |

The E=0 and E=32 Fast decoded outputs were exactly equal for this diagnostic input/profile. This observation does not establish encryption/noise equivalence, broad API completeness, or numerical-accuracy acceptance. No execution blockers occurred; the captured JSON passed validation for schema, provenance hashes, two runs, one call per E, expected metadata, and finite required metrics.

## Verified commands

- `go test ./circuits/ckks/bootstrapping -run '^(TestFastBootstrapEphemeralWeightPassthrough|TestNewEvaluatorFastDispatchDoesNotRequireDenseSparseKeys)$' -count=1` — pass.
- `go test -tags fast_ephemeral_diag ./circuits/ckks/bootstrapping -run '^TestFastEPassZeroErrorSNRIsSerializable$' -count=1` — pass.
- `go test -tags fast_ephemeral_diag ./circuits/ckks/bootstrapping -run '^$' -count=1` — pass (compile-only; no tests run).
- One execution of `go test -tags fast_ephemeral_diag ./circuits/ckks/bootstrapping -run '^TestFastEphemeralPassthroughP93$' -count=1 -v` with the required pinned-backend, canonical-config, and external-reference environment variables — pass; exactly two Fast Bootstrap calls total.

`FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**. This result is pending independent Web review and makes no speedup or drop-in-completeness claim.
