# Batch 020 — S2 original-pins cheap baseline

**Status: `S2_COMPLETE` — five valid, process-isolated paired samples; zero Bootstrap calls.** This is a descriptive measurement, not a statistically stable speedup claim.

## Frozen provenance and method

- Primary runner: `c486cb2b229f4cc6c0b77f9be2b85066c3b6c213`, clean for all ten runs.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean/read-only.
- Original Fast: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean/read-only.
- Go `go1.26.4`, `darwin/arm64`, Apple M4 / 10 CPU cores, `GOMAXPROCS=1`; `GOGC` and `GOMEMLIMIT` unset (Go defaults). Same compiled Primary source and deterministic input/profile in both builds. Compile was completed before timing. Each of five trials per backend used a separate process, fresh key material, same fixed input, no within-process warmup or state reuse; Standard ran first in each pair. Timing/allocation probes excluded capacity-observer hashing.
- `shared_frontend_sha256=fc059c1fb0b724f6f77734b5ba5e10e177c5fa6db85a7a188f61abb791f08d30`
- `config_sha256=919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`
- `workload_sha256=00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74`
- `effective_profile_sha256=862d233af63b18a46e0a2cb5e7f74cda93a9dcab6c04fc2f4b1e425c424a7044`
- `effective_qp_sha256=1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`

All runs reported `preflight_passed_bootstrap_not_started`, `bootstrap_calls=0`, and matching provenance. The unchanged native Standard lifecycle had nonzero c1; Fast retained zero c1. Fast's existing Q-prefix capacity observer passed all supported checkpoints; no dormant-row/full-RNS fallback was used. Every Standard and Fast cleartext pre-Bootstrap oracle passed the fixed `1e-6` maximum-error gate.

## Stage wall samples

Values are raw nanoseconds for trials 1–5. `median [min–max]` reports center and dispersion; allocation column is median bytes / objects per call. Allocation snapshots were taken outside the timed interval. Key generation and encryption remain separate from the eval-only chain.

### Genuine Standard

| Stage | Raw wall samples (ns) | Median [min–max] ns | Median alloc bytes / objects |
|---|---|---:|---:|
| key_generation | 133583, 674833, 651000, 633792, 830458 | 651000 [133583–830458] | 200864 / 250 |
| evaluation_key_generation | 416825042, 435906541, 419585750, 417804916, 414832292 | 417804916 [414832292–435906541] | 395172480 / 94240 |
| ckks_evaluator_init | 1014167, 1126000, 996042, 1071250, 1094333 | 1071250 [996042–1126000] | 2719992 / 16610 |
| bootstrap_evaluator_init | 415587416, 410207542, 387978250, 395321875, 406866083 | 406866083 [387978250–415587416] | 710581384 / 14232870 |
| codec_init | 131125, 111250, 73792, 52334, 74750 | 74750 [52334–131125] | 608392 / 46 |
| encrypt_a | 1711666, 1846875, 2168084, 1850416, 2351208 | 1850416 [1711666–2351208] | 4920576 / 21547 |
| encrypt_b | 1705541, 1877792, 2276125, 1733708, 2360583 | 1877792 [1705541–2360583] | 4854872 / 21543 |
| encrypt_c | 1715709, 1790375, 2113666, 1696125, 2292625 | 1790375 [1696125–2292625] | 4854872 / 21543 |
| add | 136834, 326416, 1452083, 349875, 1405541 | 349875 [136834–1452083] | 1575328 / 62 |
| mul_relin | 3283166, 3221583, 3558791, 3057125, 13017209 | 3283166 [3057125–13017209] | 5586888 / 445 |
| rescale | 410833, 492458, 400041, 371584, 416500 | 410833 [371584–492458] | 788552 / 48 |
| rotate | 1949833, 1886917, 1917750, 1783333, 1969500 | 1917750 [1783333–1969500] | 1060472 / 277 |
| drop_level | 24000, 31583, 49042, 42334, 78333 | 42334 [24000–78333] | 655824 / 17 |
| native_decode | 136500, 142667, 163084, 150625, 130834 | 142667 [130834–163084] | 329144 / 26 |

Eval-only stage-sum samples: `5941166, 6101624, 7540791, 5754876, 17017917` ns; median `6101624` ns, range `5754876–17017917` ns. Stage allocation total: `9996208 bytes / 875 objects` per trial. The short-chain wall range includes visible outliers; keep the raw samples and do not infer significance.

Across all instrumented setup, encryption, and evaluation stages, median per-process allocation was `1133909640 bytes / 14409524 objects`.

### Original Fast

| Stage | Raw wall samples (ns) | Median [min–max] ns | Median alloc bytes / objects |
|---|---|---:|---:|
| key_generation | 225500, 839667, 813416, 823042, 818459 | 818459 [225500–839667] | 200864 / 252 |
| evaluation_key_generation | 848469250, 865380917, 861084334, 884465125, 869759042 | 865380917 [848469250–884465125] | 578444224 / 421084 |
| ckks_evaluator_init | 1017000, 1886584, 1834416, 1838292, 1275208 | 1834416 [1017000–1886584] | 5408688 / 16663 |
| bootstrap_evaluator_init | 2624250, 2593375, 2617917, 2659208, 2481334 | 2617917 [2481334–2659208] | 8852352 / 32084 |
| codec_init | 85459, 76000, 104958, 73417, 107084 | 85459 [73417–107084] | 608392 / 46 |
| encrypt_a | 1198167, 1203042, 1128708, 1245750, 1171333 | 1198167 [1128708–1245750] | 4918056 / 21535 |
| encrypt_b | 1137291, 1144833, 1103959, 1187125, 1115458 | 1137291 [1103959–1187125] | 4852368 / 21532 |
| encrypt_c | 1130375, 1137417, 1099833, 1155292, 1114791 | 1130375 [1099833–1155292] | 4852368 / 21532 |
| add | 156291, 872083, 864916, 779250, 817459 | 817459 [156291–872083] | 525376 / 18 |
| mul_relin | 268334, 272083, 288708, 295833, 291792 | 288708 [268334–295833] | 525504 / 19 |
| rescale | 1724250, 1714084, 1735375, 1754875, 1729625 | 1729625 [1714084–1754875] | 657248 / 41 |
| rotate | 120833, 131625, 130334, 132083, 120166 | 130334 [120166–132083] | 591600 / 30 |
| drop_level | 47125, 49500, 51500, 51292, 48708 | 49500 [47125–51500] | 524752 / 15 |
| native_decode | 132542, 147292, 142000, 164542, 152583 | 147292 [132542–164542] | 329016 / 25 |

Eval-only stage-sum samples: `2449375, 3186667, 3212833, 3177875, 3160333` ns; median `3177875` ns, range `2449375–3212833` ns. Stage allocation total: `3153496 bytes / 148 objects` per trial (trial 5: `3153624 bytes / 149 objects`).

Across all instrumented setup, encryption, and evaluation stages, median per-process allocation was `611290840 bytes / 534878 objects`.

## Paired numerical and storage gates

| Trial | Standard-vs-Fast pre-Bootstrap RMSE | Max complex difference | Gate |
|---:|---:|---:|---|
| 1 | 1.46203107655e-11 | 6.81138397995e-11 | PASS |
| 2 | 1.43288267233e-11 | 6.91904100577e-11 | PASS |
| 3 | 1.45509715048e-11 | 6.28708689093e-11 | PASS |
| 4 | 1.45172349443e-11 | 5.84609944353e-11 | PASS |
| 5 | 1.45448913764e-11 | 6.50102241513e-11 | PASS |

All use 4096 decoded slots and the unchanged maximum-error threshold `1e-6`. Standard native oracle RMSE samples were `1.4600229225241158e-11, 1.4316599218857392e-11, 1.453833779462367e-11, 1.4499933969659996e-11, 1.4530727577144727e-11`; its max-error range was `5.86840482580165e-11–6.924536582235276e-11`. Fast native oracle RMSE/max were stable at `7.442066238955308e-13 / 2.1986015127059466e-12` for all five trials.

Approximate ciphertext q-backed slice-capacity bytes (not total process memory; rows × coefficients × 8, excluding object/allocator/key memory):

| Backend | Encrypt inputs L5 | Add/MulRelin | Rescale/Rotate L4 | DropLevel L0 |
|---|---:|---:|---:|---:|
| Standard | 786432 | 786432 | 655360 | 131072 |
| Fast | 786432 | 524288 | 524288 | 131072 |

Peak RSS is **not reported**: macOS `/usr/bin/time -l` failed while querying `sysctl kern.clockrate` under the sandbox; `/usr/bin/time -p` worked but provides no RSS. This is kept distinct from the q-backed estimate and Go allocation counters.

## Bounded interpretation

On these five cheap, non-Bootstrap observations, the Fast eval-only stage-sum median was `3.178 ms` versus Standard `6.102 ms`. Fast's highest measured eval stage was Rescale (`1.730 ms`, about 54% of its stage-sum median); Standard's were MulRelin (`3.283 ms`) and Rotate (`1.918 ms`). Fast's Add timing showed high range (`0.156–0.872 ms`) and Standard had large Add/MulRelin outliers; the samples are retained as-is. The full active pre-Bootstrap process medians were Standard `874.004 ms` and Fast `913.683 ms`, dominated by different key/evaluator initialization costs. These are descriptive stage observations only; no Bootstrap timing or end-to-end speedup has been measured.

An initial instrumentation smoke process was excluded because it ran with `GOMAXPROCS=10` and `/usr/bin/time -l` exited on the sandbox `sysctl` denial. It made zero Bootstrap calls. The five listed samples were all rerun with and recorded `GOMAXPROCS=1`; no RSS substitute was inferred.
