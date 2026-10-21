# FIXED-SIGMA-CHAIN-002 — two-step noise compensation report

**Status: BATCH_COMPLETE_READY_FOR_WEB_REVIEW.** CP-A and CP-B passed their scoped gates. CP-C completed three matched Standard/Fast+σ two-operation chains and is classified `RMSE_ONLY`. CP-D completed three sequential Fast fixed-σ captures plus v3 `compare-public` validation. No Bootstrap was invoked; authorized Bootstrap spend remains `0/0`.

## Workload and frozen conditions

LogN13 / LogSlots12 / 4096 slots / E=32, canonical A/B/C workload, Scale `2^45`. Stage 1 is Level5→4 with q5 divisor; Stage 2 is Level4→3 with q4 divisor and re-encodes the same C message at Scale q4. Both outputs remain at Scale `2^45` and use authoritative q0..q3 rows. Fixed coefficient σ is `4.004280012097432`; each completed Fast MulRelin+Rescale stage receives exactly one explicit diagnostic injection. No generic Rescale hook or Bootstrap behavior is covered.

## Stagewise numerical results

CP-B has one provenance-clean matched baseline pair. CP-C has three matched Standard/Fast+σ pairs. RMS/variance/max below are the **new-operation residual** computed from the actual decoded stage inputs; total-oracle errors are kept separate in the journal and compact pair artifacts.

| Stage | Genuine Standard CP-B RMS / variance / max | Original Fast CP-B RMS / variance / max | Standard CP-C RMS range | Fast+σ CP-C RMS / variance / max ranges |
|---|---:|---:|---:|---:|
| Stage 1: Level5→4 | `1.03186e-11 / 1.06401e-22 / 4.23940e-11` | `7.43728e-13 / 5.53131e-25 / 2.20541e-12` | `1.03448–1.03930e-11` | `1.02955–1.03860e-11 / 1.05904–1.07854e-22 / 2.74403–3.06560e-11` |
| Stage 2: Level4→3 | `1.03435e-11 / 1.06978e-22 / 5.91370e-11` | `7.39085e-13 / 5.46152e-25 / 2.13309e-12` | `1.01831–1.04529e-11` | `1.03316–1.03770e-11 / 1.06737–1.07683e-22 / 2.76059–3.06130e-11` |

Across the three CP-C pairs, the Fast+σ residual RMS is close to Standard, while residual variances differ and Fast+σ maximum residuals remain lower. Cross-backend covariance and coefficient moments are recorded per stage in `journal.md` and `cp-c-pair-01.json` through `cp-c-pair-03.json`. Evidence supports only `RMSE_ONLY`; it does not establish Gaussian, tail, or distribution equivalence. The frozen σ was not retuned.

| Stage | Standard CP-B total-oracle RMS | Original Fast CP-B total-oracle RMS | Standard CP-C total-oracle RMS range | Fast+σ CP-C total-oracle RMS range |
|---|---:|---:|---:|---:|
| After Stage 1 | `1.03217e-11` | `7.44207e-13` | `1.03438–1.03947e-11` | `1.02957–1.03858e-11` |
| After Stage 2 | `1.03649e-11` | `7.39395e-13` | `1.01945–1.04587e-11` | `1.03325–1.03807e-11` |

Stage 2 total-oracle error includes propagated Stage 1 error; it is not substituted for the Stage 2 incremental residual.

## Injection and attribution

Each CP-C chain had two injections and zero Bootstrap calls. Stage 1/2 seeds were `(20261025, 20261026)`, `(20261027, 20261028)`, and `(20261029, 20261030)`; all used the same σ. Empirical coefficient RMS values ranged `3.98859–4.02602`; each strict capacity check covered 8192 coefficients across four active Q rows. CP-C outer helper observations ranged `1.048–1.212 ms` per injection.

CP-D measured five phases inside the existing opt-in helper: sampling/hash/moments, coefficient mapping plus four NTTs, c0 candidate copy/add, centered Q-prefix capacity validation, and c0 commit. Median phase-sum time was `1.118 ms` / `1.163 ms` for Stage 1 / Stage 2, compared with Mul+Rescale medians of `2.165 ms` / `1.876 ms`; per-run phase-sum ratios were `50.63–54.64%` / `57.66–69.32%`. Outer helper timers include about `0.20–0.26 ms` instrumentation and wrapper overhead. The capacity checker dominated phase time at about `66–68%`; the four main measured phases each allocated about `71 KB`, `262 KB`, `262 KB`, and `262 KB`. These are bounded observations with process-wide allocation snapshots, not an optimization benchmark. The checker remains mandatory and unchanged.

## Provenance and evidence

- CP-B accepted artifact: [`cp-b-pair-01.json`](cp-b-pair-01.json), SHA-256 `95a858ad8c893242ffcb1b18c19e0a98f078601b4f1272231ada7d62f7f970cb`. Pairs 02/03 were excluded for stale binary/runtime Primary provenance; CP-B's three-pair allowance is exhausted.
- CP-C compact pair artifacts: `cp-c-pair-01.json` SHA-256 `0f61f4e2e9945f8179a2ccaa29013b9f8761805a817b684f25e1763dbb64d5cf`; `cp-c-pair-02.json` SHA-256 `248f9f96a7acd3eb522c53a9450ac94f88c23dd7c6f8e14139d47c2ce52f9f71`; `cp-c-pair-03.json` SHA-256 `d11a1dcede135f5fd9bd77bb98f45eb064c07786aa891ba183572206159b637c`.
- CP-C binary SHA-256: genuine Standard `a2fbf1a0c94536b2bc1b5bd67999567bae05be1382638e7a6b5c7320fa002e32`; Fast `426f7b57e5f5e3223093248d06fc227a79c397ea4b9575aa103e9f899080103a`. Primary build was clean commit `718018cafdd6e2c7c86e6a14342e44b99598e07d`; Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; CP-C Fast pin `b5c55e0b82dd6ad4a333d079141d33745edcb379`.
- CP-D measurement Primary commit: `3420c3f0a8e6b398a1c4164b1dc17ffd0dd66c88`; measurement-source SHA-256 `409856ddd4f5bcd4696922d03ee4b0aa49587f1ac144af83b1f4e2a28d08fecd`; Secondary instrumentation and accepted chain pin: `1e7828293c36f9845e30723ba4e9500d84bb50d2`.
- CP-D binary SHA-256: genuine Standard `d3036b90681b59425a7d36ca542e6364d188024df8fea4124aeae79619ec17bc`; Fast `0370cb66aa519567243f0a2dc5dd442076ff4e0dcfa38f4c40a4fdc9d3eb3a7c`. Three v3 pair artifacts are committed as `cp-d-pair-01.json`, `cp-d-pair-02.json`, and `cp-d-pair-03.json`; hashes and raw result/vector provenance are in `journal.md`.
- CP-D compact raw result docs are also committed as `cp-d-standard-reference.json` and `cp-d-fast-run-01.json` through `cp-d-fast-run-03.json`, preserving every per-phase time/allocation observation. Decoded vectors remain outside Git.
- CP-C raw vectors/results remain outside Git under `/private/tmp/fixed-sigma-chain-002-cp-c-v2-*`; they were size-checked before targeted summary extraction. No multi-megabyte vectors are committed.
- CP-B and CP-C each used 0 Bootstrap calls; no LogN16, large benchmark, or teammate BTS was run.

See [`journal.md`](journal.md) for exact per-pair means, variances, maxima, covariance, injection seeds/timings, provenance exclusions, tests, and measurement limits.
