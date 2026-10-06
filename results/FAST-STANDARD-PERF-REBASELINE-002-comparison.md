# FAST-STANDARD-PERF-REBASELINE-002 — matched profile comparison

**Joint outcome: `DUAL_LOGN13_LOGN16_PERF_SNR_NUMERICAL_FAIL`.** Both profiles completed with matched committed harness, environment, parameters and inputs. LogN13 passes the 1e-2 output component threshold; LogN16 does not. The strict Q-prefix capacity audit passes all 56 checkpoints in both profiles, so the LogN16 failure is numerical quality, not a capacity-gate failure.

## Provenance and measurement contract

- Matched public timing/numerical harness Primary commit: `68558a82af6faffd304fa0f14bbc4c82fa190176` (clean detached paired worktrees).
- Fast Secondary: `5117fc57949647182f476dc5952099c706b9f869` (`fast-qprefix`); genuine Standard Secondary: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Both profiles ran on `go1.26.4`, `darwin/arm64`, CPU `arm64`, with one untimed warmup and seven full public Bootstrap timing repetitions per backend. Numerical trials: three independent genuine Standard key trials and two Fast executions per profile.
- The input is the same deterministic `c0=encoded-message, c1=0` workload within each pair. Standard pre/post outputs were decrypted and decoded with fresh Standard keys; Fast used the current direct-c0 zero-a path. This is not a security-equivalent encrypted-input or RLWE-noise comparison.
- Topology-aware SNR / amplification and detailed PS, generated-power, EvalMod, DoubleAngle and Q-prefix audit are in each profile report and numerical JSON. The deep fastdiag replay was clean at Primary `4bba6b5cf403aa7ecd79005b78f974b11f07fafc` with the same pinned Fast commit; `cmd/fastdiag` and `internal/numericalmetrics` are unchanged between that diagnostic SHA and the matched timing harness SHA. It uses the Standard evaluator API built from the pinned Fast source tree, not a separate genuine-Standard build. Raw public checkpoints with different maintained Q-row counts are marked not comparable; common-Q replay is reported separately.
- Full compact-profile reports: [`LogN13`](FAST-STANDARD-PERF-REBASELINE-002-logN13-report.md) and [`LogN16`](FAST-STANDARD-PERF-REBASELINE-002-logN16-report.md). Full fastdiag JSON source artifacts: [`LogN13 raw`](FAST-STANDARD-PERF-REBASELINE-002-logN13-fastdiag-raw.json), SHA-256 `356c168a8d5dbbe480a2361ccb287f2ee38867ff773cdd033a9960998d1e15dd`; [`LogN16 raw`](FAST-STANDARD-PERF-REBASELINE-002-logN16-fastdiag-raw.json), SHA-256 `716a3a6addc8bb08649d4d891ab8cd73f052e3c22f4723e72c5cd839de1e3554`.

## Results

| Profile | Standard / Fast median ms | Speedup | Standard / Fast / Fast−Standard SNR dB | Standard / Fast original RMSE | Fast−Standard RMSE | Median precision bits | Alloc B/op Standard / Fast | Allocs/op Standard / Fast | Threshold violations real / imag | Capacity failures / checkpoints | Classification |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| LogN13, q0=55 (55 actual bits) | 296.456 / 67.754 | 4.3755× | 144.713390 / 144.710750 / −0.002640 | 1.19039e−9 / 1.19075e−9 | 1.44681e−10 | 32.9732 | 91,864,720 / 7,499,173 | 34,638 / 7,903 | 0 / 0 | 0 / 56 | `FAST_STANDARD_NUMERICAL_CLOSE` |
| LogN16, q0 target=55 (56 actual bits) | 3827.880 / 675.102 | 5.6701× | 130.947666 / −0.000462 / −130.948128 | 5.80819e−9 / 0.0204855 | 0.0204855 | 5.7622 | 849,679,900 / 57,363,056 | 40,299 / 14,066 | 138,780 / 30,306 | 0 / 56 | `FAST_NUMERICAL_QUALITY_DEGRADED` |

LogN16 first material checkpoint is `evalmod_real`: max complex difference `0.00932384135` versus material threshold `0.00364073040`. Final Fast-vs-Standard complex RMSE is `0.02048549798` (> `1e-2`), maximum complex difference is `0.04628102297`, and real/imag coordinate violations are `138780 / 30306`. Do not classify this profile as numerically restored; no threshold was weakened.

LogN13 config SHA-256: `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; input SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.

LogN16 config SHA-256: `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02`; input SHA-256: `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf`.

## Historical result — not comparable

The prior `6.671×` LogN16 speedup used old Fast commit `82601ea2517edc14784c9da250426169a1b221c7`. It is retained only as historical context and is not the corrected Fast result above.
