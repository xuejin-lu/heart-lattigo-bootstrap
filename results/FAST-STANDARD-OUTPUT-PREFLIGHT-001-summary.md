# FAST-STANDARD-OUTPUT-PREFLIGHT-001

**Classification: `OUTPUT_PREFLIGHT_NUMERICAL_CONCERN`**
**Output accuracy: `UNASSESSED`**
**Review: `READY_FOR_WEB_REVIEW`**

All four one-operation public Bootstrap runs completed. Both backend inputs passed the existing `1e-6` pre-Bootstrap input-quality gate; outputs had valid metadata, expected decoded slot counts, finite values, and computed metrics. Each run performed exactly one Bootstrap, with no timing, warmup, stage replay, or speedup measurement.

The Standard output-vs-original errors substantially exceed the descriptive `1e-2` component cutoff, so the result is classified as a numerical concern for Web interpretation. This cutoff is not a pass/fail gate. No algorithmic cause is inferred, and output quality remains unassessed.

## Provenance and matched profile checks

- Primary harness commit: `59899ef084380a4fc01d2582798e5113aac2409c`, clean at each run.
- Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, `perf_standard`, clean detached source `/private/tmp/fast-standard-rebaseline-002/lattigo-standard`.
- Fast: `5feb44917fca40c93abec6def6f26bc81a82c536`, `perf_fast`, clean detached source `/private/tmp/fast-standard-input-provenance-001-lattigo-fast`.
- `verifyCompiledBackendSource` confirmed the compiled Lattigo replacement path for each invocation; each run also verified the requested Secondary HEAD and clean state.
- Go `go1.26.4`, `darwin/arm64`; all four records use the same Primary SHA and machine environment.
- Same-profile Standard/Fast original-vector and config SHA-256 values match. A full sorted JSON comparison of `effective_parameters` was identical within each profile:

| Profile | Slots | Config SHA-256 | Original SHA-256 | Canonical effective-parameter SHA-256 |
|---|---:|---|---|---|
| LogN13 | 4,096 | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` | `45c5efad1af6bbeab61e1f00a296a2a5c4ed3347356b9508ff9c6eb96cd02c32` |
| LogN16 | 32,768 | `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02` | `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf` | `b6f919bc439b413050262106e459f1377261d5a9129b3282a91f74b99218be6e` |

The equality check compared each profile's complete sorted `effective_parameters` JSON object; the summary stores its canonical SHA-256 plus the full Q/P bit-chain and operational parameter fields. LogN13 q0 is 55 bits; LogN16 q0 is 56 bits.

## Input and output evidence

Standard used fresh native `GenSecretKeyNew` / `EncryptNew`, public Bootstrap, and same-key decrypt/decode. Fast used the explicitly insecure `fast_zero_secret_direct_encoded_v1` direct-encoded zero-secret simulation, public Fast Bootstrap, and c0-direct decode; it is not secure public-key encryption. Input maximum deviations were below `1e-6` in all runs.

Output metadata in every run: Level 1, Scale log2 45, Degree 1, NTT=true, Montgomery=false, two components. Decoded slots match the profile (4,096 or 32,768). “Over cutoff” counts are descriptive real/imag component counts at `1e-2`.

### Public output versus original input

| Profile | Backend | Input max deviation | Complex RMSE | Max complex | Max real | Max imag | SNR (dB) | Real / imag over cutoff |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| LogN13 | Standard | `2.229618e-11` | `2.195800` | `3.105585` | `3.104650` | `3.104778` | `-40.604659` | 4,048 / 4,055 |
| LogN13 | Fast | `2.101955e-12` | `1.190755e-9` | `5.805793e-8` | `1.648099e-9` | `5.805750e-8` | `144.710751` | 0 / 0 |
| LogN16 | Standard | `7.549669e-11` | `10.348459` | `16.116659` | `15.938221` | `15.764809` | `-54.069045` | 32,744 / 32,747 |
| LogN16 | Fast | `6.442118e-12` | `5.808193e-9` | `4.708805e-7` | `1.552243e-8` | `4.708494e-7` | `130.947647` | 0 / 0 |

### Pre-Bootstrap versus post-Bootstrap

| Profile | Backend | Complex RMSE | Max complex | Max real | Max imag | SNR (dB) |
|---|---|---:|---:|---:|---:|---:|
| LogN13 | Standard | `2.195800` | `3.105585` | `3.104650` | `3.104778` | `-40.604659` |
| LogN13 | Fast | `1.190755e-9` | `5.805821e-8` | `1.647693e-9` | `5.805779e-8` | `144.710750` |
| LogN16 | Standard | `10.348459` | `16.116659` | `15.938221` | `15.764809` | `-54.069045` |
| LogN16 | Fast | `5.808181e-9` | `4.708794e-7` | `1.552307e-8` | `4.708484e-7` | `130.947666` |

The raw per-slot vectors and ciphertexts were not serialized. The machine-readable summary contains aggregate metrics, metadata, provenance, and fingerprints only.

## Validation

- `go test ./cmd/perfprobe ./internal/perfmeasure`
- `go test -tags perf_standard ./cmd/perfprobe`
- `go test -tags perf_fast ./cmd/perfprobe`
- `git diff --check`
- Four separate `go run` output-smoke executions using the pinned source worktrees: Standard/Fast × LogN13/LogN16. Each was invoked with `--warmup=0 --repetitions=1 --standard-trials=1`; output-smoke recorded one public Bootstrap and `timing_performed=false`.

No Secondary source was changed. No benchmark campaign, parameter tuning, numerical diagnosis, or historical result deletion was performed.
