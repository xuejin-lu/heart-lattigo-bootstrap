# FAST-STANDARD-EPHEMERAL-ABLATION-001

**Classification: `EPHEMERAL_WEIGHT_HYPOTHESIS_SUPPORTED`**
**Output accuracy: `UNASSESSED` pending independent Web review**

## Controlled protocol and provenance

The paired diagnostic changed only `bootstrapping.Parameters.EphemeralSecretWeight` from `0` to `32`. Within each profile, one generated Standard secret and one native `EncryptNew` ciphertext were shared by both evaluators; the exact ciphertext was copied for each public Bootstrap and verified unchanged afterward. Each weight performed exactly one Bootstrap. The input had nonzero `c1` and was decrypted with the matching Standard secret before and after Bootstrap. `SecretHamming=192`, `K=16`, all configured Q/P primes, and all other parameters were held fixed. There was no Fast run, warmup, timing, benchmark, sweep, or formal comparison.

- Primary runner commit: `d654fc39bb2c5fbe41018fac6e3dce4cb20b450f`, clean during both runs.
- Genuine Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean detached worktree at `/private/tmp/fast-standard-rebaseline-002/lattigo-standard`; compiled module replacement was verified against that path.
- Environment: Go `go1.26.4`, `darwin/arm64`; diagnostic existed only in a `perf_standard && ephemeral_diag` test build.
- Source evidence at the pinned Standard commit: `circuits/ckks/bootstrapping/keys.go` returns without encapsulation keys when weight is zero; nonzero weight generates the dense-to-sparse and sparse-to-dense keys using a sparse secret. `parameters_literal.go` defines the Standard default ephemeral weight as 32. Runtime key-presence evidence matched this behavior for both profiles.
- No key/ciphertext material, key fingerprints, or per-slot vectors are in this report. Compact JSON measurements were written outside the repository under `/private/tmp`.

| Profile | Slots | Input max deviation | Config SHA-256 | Original SHA-256 | Effective parameters SHA-256 | Q primes SHA-256 | P primes SHA-256 | Q/P bits |
|---|---:|---:|---|---|---|---|---|---|
| LogN13 | 4,096 | `2.59440524e-11` | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` | `f151442a4e08e1ebf8b7bb515fdd075a3748298bee1f3e2b990a6ad99aeca693` | `30066778ef1caea959b3357b0a70ff792fe8582ba9545f6b53c8df2a1a6cb568` | `70451211d27cd1e6bf092aa3c0752a63147c34842211dcd70d64d50755d318f7` | Q `[55,39,40,39,40,60,60,61,60,60,60,61,61,56,57,56,56]`; P `[61,61,62,61,62]` |
| LogN16 | 32,768 | `7.93226244e-11` | `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02` | `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf` | `9286a69c1c3184fbad9d9a7d3590de7aed34a819fcdb301f7d552e85fa76e10` | `2b2fadb04da2a823b2c479e63c2a771de8242aa2a91d3bcb6405d7025536989b` | `b212ae9f5739e7e70a78d2a4202f58fb34f05d760d72d047a66c263bcfc091e0` | Q `[56,39,39,39,39,60,61,60,61,61,60,60,61,57,56,56,57]`; P `[61,61,62,61,62]` |

## Bootstrap output versus original message

Each output had Level 1, Scale log2 45, Degree 1, NTT=true, Montgomery=false, two components, and the expected decoded slot count. SNR status was finite in all four observations.

| Profile | E | Dense→sparse / sparse→dense keys | Bootstrap calls | Complex RMSE | Max complex | Max real | Max imag | SNR (dB) |
|---|---:|---|---:|---:|---:|---:|---:|---:|
| LogN13 | 0 | absent / absent | 1 | `9.355135785595e-6` | `2.276041695758e-5` | `2.271743366533e-5` | `2.089757570216e-5` | `66.806195` |
| LogN13 | 32 | present / present | 1 | `4.835051363184e-9` | `1.491374744237e-8` | `1.232375763938e-8` | `1.489540921509e-8` | `132.539175` |
| LogN16 | 0 | absent / absent | 1 | `10.683906250769` | `17.717852403489` | `17.558789115316` | `17.518458128112` | `-54.346133` |
| LogN16 | 32 | present / present | 1 | `3.999112507605e-8` | `1.433616002053e-7` | `1.349181891312e-7` | `1.275886178462e-7` | `114.189196` |

The E=32 output's complex RMSE was lower than E=0 by `99.9483166%` for LogN13 and `99.99999963%` for LogN16. The direct E32-vs-E0 complex RMSE was `9.35512256e-6` (max complex `2.27588018e-5`) for LogN13, and `10.68390625` (max complex `17.71785245`) for LogN16. This supports the specified controlled ephemeral-weight hypothesis for these paired inputs; it does not establish that the full historical Standard failure is explained or generalize across key/ciphertext randomness.

Notably, the LogN13 E=0 result here (`9.3551e-6`) differs substantially from the predecessor output preflight's Standard LogN13 RMSE (`2.195800`). This single paired diagnostic does not resolve that cross-run discrepancy. E=32 remains diagnostic-only and is not promoted into the formal Standard profile; numerical quality remains unassessed pending Web review.

## Validation

- `go test -modfile=/private/tmp/fast-standard-rebaseline-002/standard.mod ./cmd/perfprobe ./internal/perfmeasure` — pass.
- `go test -modfile=/private/tmp/fast-standard-rebaseline-002/standard.mod -tags perf_standard ./cmd/perfprobe ./internal/perfmeasure` — pass.
- Tagged LogN13 and LogN16 paired runs using `-tags 'perf_standard ephemeral_diag'` — pass; exactly two public Bootstrap calls per profile total.
- `go test -modfile=/private/tmp/fast-standard-rebaseline-002/standard.mod -tags perf_standard ./cmd/perfprobe -list '^TestFastStandardEphemeralAblation$'` — the diagnostic test is absent from the ordinary `perf_standard` build.
- `git diff --check` — pass.

No Primary production arithmetic, shared profile, official config, formal comparison route, benchmark route, or Secondary source was modified.
