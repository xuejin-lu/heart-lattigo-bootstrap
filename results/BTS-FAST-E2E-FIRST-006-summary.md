# BTS-FAST-E2E-FIRST-006 — outcome

Status: `NEEDS_WEB_REVIEW` (2026-10-12)

The scoped Fast public `Sub` repairs were implemented, tested, committed, and pushed. Three authorized Fast BTS calls were entered. The third `bts.Bootstrap` returned a ciphertext, but the unchanged BTS program then panicked in the ordinary RLWE decryptor before decoding or exporting vectors. The original BTS end-to-end run therefore did not complete, and no Standard/Fast numerical comparison is available. The three-call budget is exhausted; no retry was made.

## Fast backend changes and checks

- Policy amendment: Secondary `c65ac76dc534a398504c292a6a9cda055516c53b` records the experimental-only `CAPACITY_UNPROVEN_EXPERIMENTAL` waiver. It does not alter Q-prefix or dormant-row rules and does not extend to Standard, secure CKKS, or nonzero-secret paths.
- `75b9cab0c000cc07b8ba0604384bea5178dbd33e`: public different-Scale `Sub` using the shared scalar and Add/Sub Q-prefix cores.
- `3b2fef14357d127899d97523d6543fe69f476842`: accepts the BTS L25 output with its low-Scale L26 right operand, computing only at output Level.
- `e05283690fcb0262540a2f979b570dc979f49682`: supports the actual zero-secret degree-1 T3 product while retaining degree-2/degree-1 coverage.
- Focused CKKS Add/Sub/T3 tests, `go test ./schemes/ckks/internal/fastcore -count=1`, `go test ./schemes/ckks/fast -count=1`, affected-package `go vet`, and `git diff --check` passed after the final repair. A separate scalar-Add test failure was reproduced unchanged at pristine `d1b6dc9963b0df95a7c149306fb62660e1020f7b`; it is unrelated to this Sub repair.

## No-Trace Fast attempts

All builds used the unchanged `external/BTS` source/parameters and a measurement-only overlay that exports original plaintext and decoded real slots after the original decode. Source manifest SHA-256: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`. The overlay did not alter CKKS operations, parameters, or order. No output vector was produced by any attempt.

| Attempt | Fast commit / binary SHA-256 | `NewBTS` | Outcome |
|---|---|---:|---|
| 1 | `75b9cab0c000cc07b8ba0604384bea5178dbd33e` / `3b41770457e04960bc8a8505c1d29fd6b05afe652a4bf6c02dba6e5a7e44cc53` | 46.952 s | `EvalSine → genPower T[6] → Sub`: compact input missing active q4; Bootstrap returned error. |
| 2 | `3b2fef14357d127899d97523d6543fe69f476842` / `1d01bbdd152275fd2bdc2289e55273551d8a8945966827b21a2be9777c3b7d74` | 44.883 s | Same guard failure. Source inspection showed Fast zero-secret `MulRelinNew` produces degree 1; the helper had only accepted degree 2. |
| 3 | `e05283690fcb0262540a2f979b570dc979f49682` / `1ec5c97ae145e6becbd002d631594a63f7a72a71ee8bc040366926eac971ddb8` | 43.814 s | `bts.Bootstrap` returned a ciphertext. Original `rlwe.Decryptor.Decrypt` then panicked in `ring.addvec` because c1 q4 backing had length 0. No verified output Level/Scale, decoded slots, or Bootstrap elapsed time was emitted. |

Raw logs and immutable call accounting are in `results/BTS-FAST-E2E-FIRST-006/` and `results/BTS-FAST-E2E-FIRST-006-journal.json`. Fast budget: 3/3 entered, 1 Bootstrap returned a ciphertext, 0 decoded BTS outputs. The adapter vector files are absent because the original decode did not return.

## Numerical comparison and remaining blocker

The earlier `BTS-LOGN16-NOTRACE-BASELINE-001` report contains aggregate Genuine Standard errors only (max `1.708e-3`, mean `1.172e-4`), not a 32,768-slot decoded vector. Its source/parameter profile matches, but it is insufficient for the requested pairwise comparison. The one newly authorized Standard call was not used because Fast produced no decodable output. No `MaxAbsDiff`, RMSE, `MeanAbsDiff`, or Fast-vs-plaintext error is claimed.

The next decision is whether/how the ordinary decrypt/decode path may safely consume the compact Fast zero-c1 representation without reading unavailable dormant rows or changing the BTS frontend. That follow-up requires Web review and a new bounded Fast-call authorization; no fourth Fast call is included here. Fast remains an intentionally insecure zero-secret simulation; capacity is `CAPACITY_UNPROVEN_EXPERIMENTAL`, not proven.
