# FAST-ZERO-SECRET-E-PASSTHROUGH-001

**Classification: `FAST_E_PASSTHROUGH_PARTIAL`**
**Handoff: `NEEDS_WEB_REVIEW`**

## Scope and provenance

- Primary: `main`, clean at `e884658aebcf7bf7ceffc1be76d679a5cf0b84c1` during this measurement.
- Secondary implementation: `xuejin-lu/lattigo`, `fast-qprefix`, implementation commit `c0ac40b736aee441dd32a85ff6cd40b2a9be4b4e`.
- Secondary diagnostic serialization repair: `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`; this changes only diagnostic SNR representation/test coverage, not production arithmetic. It is the current pushed Secondary HEAD.
- Genuine Standard reference: clean detached checkout `5dbffbdea05394de2ca3a432ed5318aa832e3f40` at `/private/tmp/fast-standard-rebaseline-002/lattigo-standard`.
- Measurement environment: Go `go1.26.4`, `darwin/arm64`; canonical config `configs/bootstrap_config.logN13.json`; report timestamp `2026-10-08T18:57:43Z`.
- Secondary is clean on `fast-qprefix` at `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`; Standard checkout is clean and detached at the pinned commit above.
- Canonical LogN13 profile hashes: config `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; original input `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; effective parameters `f151442a4e08e1ebf8b7bb515fdd075a3748298bee1f3e2b990a6ad99aeca693`; Q primes `30066778ef1caea959b3357b0a70ff792fe8582ba9545f6b53c8df2a1a6cb568`; P primes `70451211d27cd1e6bf092aa3c0752a63147c34842211dcd70d64d50755d318f7`.

## Implementation evidence

Fast Bootstrap now accepts only `EphemeralSecretWeight` 0 or 32 and preserves the caller-supplied value. Other values remain rejected. The existing public `GenEvaluationKeys` / `NewEvaluator` compatibility dispatch selects the Fast path; it does not construct the Standard CKKS evaluator. At E=32, generated Fast-compatible key material contains Dense/Sparse keys, but the Fast Bootstrap path does not pass through the Standard Dense/Sparse KeySwitch implementation. The focused dispatch test also confirms that Fast-compatible construction succeeds without either key.

The Fast diagnostic input is the canonical deterministic LogN13 message encoded as a plaintext-like ciphertext with `c1=0`; it is diagnostic-only and is not a normal encryption or security-equivalent path. The Standard reference used native `EncryptNew`, one public E=32 Bootstrap, and matching-secret `DecryptNew`/decode.

## Measurements

The one allowed genuine Standard E=32 run completed:

| Backend / E | Calls | Output metadata | Complex RMSE vs original | Max complex | Max real | Max imag | SNR |
|---|---:|---|---:|---:|---:|---:|---:|
| Standard E=32 | 1 | Level 1, Scale `2^45`, degree 1, N=8192, 2 components, NTT=true, Montgomery=false, 4096 slots | `4.80507250213e-9` | `1.44564238201e-8` | `1.27025209518e-8` | `1.40677870030e-8` | `132.593198348 dB` |

The tagged Fast run executed exactly one public Bootstrap for E=0 and one for E=32; both returned successfully and passed decode/finite-value checks. Aggregate JSON serialization then failed because the zero-error SNR case produced `+Inf`, which Go's JSON encoder rejects. Thus no Fast-vs-original, E32-vs-E0, or Fast-E32-vs-Standard metrics were persisted. The tagged test was repaired to represent zero-error SNR as `snr_status=infinite_zero_error` with no numeric `snr_db`; a pure serialization test passes. **No Fast Bootstrap was rerun**, preserving the one-call-per-E limit. The numeric Fast comparison remains incomplete and is not inferred from the successful calls.

## Validation

- `go -C /Users/xuejinlu/Developer/xuejin-lu/lattigo test ./circuits/ckks/bootstrapping -run 'TestFastBootstrapEphemeralWeightPassthrough|TestNewEvaluatorFastDispatchDoesNotRequireDenseSparseKeys' -count=1` — pass.
- `go -C /Users/xuejinlu/Developer/xuejin-lu/lattigo test ./schemes/ckks/fast -count=1` — pass.
- `go -C /Users/xuejinlu/Developer/xuejin-lu/lattigo test -tags fast_ephemeral_diag ./circuits/ckks/bootstrapping -run '^TestFastEPassZeroErrorSNRIsSerializable$' -count=1` — pass; no Bootstrap invoked.
- Tagged diagnostic runner compile-only (`-run '^$'`) — pass.
- Pinned Standard runner compile-only — pass; one Standard E=32 run completed as above.
- Fast diagnostic attempt — two calls completed; aggregate output was not saved due to the serialization defect. No retry was performed.
- `git diff --check` — pass for Secondary changes.

The task is partial because Fast numerical metrics were not preserved and the task explicitly permits only one Fast Bootstrap per E. The next measurement requires independent authorization; this report makes no correctness, security-equivalence, or performance claim.
