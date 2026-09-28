# QPREFIX-PERF-DIAG-001 — Performance Attribution

## Classification

`UNEXPLAINED_PERFORMANCE_REGRESSION`

The evidence identifies a large Rescale/CRT-bearing hot path and a broad cost from increasing high-level authority from q01 to q0123. It does **not** quantitatively close the matched historical ~3.8x end-to-end gap: the historical baseline was not re-run under the current stage instrumentation, and baseline allocation/copy measurements are unavailable. The residual therefore remains an attribution question for Web review, not a production-fix authorization.

## Scope and provenance

- Production candidate source: `xuejin-lu/lattigo@f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`.
- Read-only Secondary checkout used for tests/benchmarks: `fast-qprefix@bc230b916ff7a598aba00eeaa60e834d02bad51c`; relative to the production candidate, the only intervening file is `CURRENT_TASK.md`.
- Historical baseline: `40532b4dce5c7eeae2db5b0b6f21be64801ce923`.
- Historical matched q0=55 result source: `results/QPREFIX-IMPL-009-PERFORMANCE-REVIEW.md`.
- Machine: Apple M4; Go `go1.26.4`; `darwin/arm64`, Darwin kernel `25.6.0`; `GOMAXPROCS=10` (benchmark suffix `-10`).
- No production source, tests, parameters, or Secondary files were changed. Benchmark/test helpers were supplied only through a Go overlay and removed before completion.

### Historical width path (q0=55)

At the baseline, `MaintainedLimbCount` returns 3 only when `bits.Len64(q0)==56`; the matched q0=55 profile therefore selects q01 at every operational level >=1. Level 0 starts with q0. This is confirmed by the baseline selector and its callers in `schemes/ckks/fast/q012.go`, `circuits/ckks/bootstrapping/fast_modup.go`, `circuits/ckks/polynomial/fast.go`, and the Fast evaluator/DFT wrappers.

| Historical baseline P93 stage | Actual authority at q0=55 |
|---|---|
| Input / ScaleDown at Level 0 | q0 |
| ModUp output and Trace | q01 (the basis lift starts from q0 and materializes q1; no q2 branch for q0=55) |
| C2S groups | q01 |
| polynomial powers / PS | q01; the Q012-specific schedule predicate is false |
| DoubleAngle | q01 |
| S2C groups / public output | q01 at Levels >=1 |

This baseline is **not Q012**. The historical 3.79x ratio is a policy migration comparison, not a four-row implementation-only comparison.

### Current q0=56 Q-prefix-v2 path

The production constitution remains `rows=min(Level+1,4)`; no width was changed.

| Stage | Level(s) | Authority | Median elapsed |
|---|---:|---|---:|
| public pack / N1→N2 (this fixture has N1=N2) | 0 | q0 | 5.286 µs |
| ScaleDown | 0 | q0 | 12.621 µs |
| ModUp | 0→16 | q0123 at Level 16 | 0.457 ms |
| C2S outputs | 15,14,13,12 | q0123 | 15.660 ms total |
| EvalMod input / output | 12→4 | q0123 throughout | 40.518 ms real; 40.653 ms imag |
| S2C groups | 4→3→2→1 | q0123, q0123, q012 | 18.980 ms total |
| N2→N1 / unpack (N1=N2 here) | 1 | q01 output | 10.982 µs |
| public finalization | 1 | q01 | 11.039 µs |

C2S has one BSGS factor per group, with restore exponents 4, 2, 0, 0. S2C group inputs are respectively Level 4/rows 4, Level 3/rows 4, Level 2/rows 3. The production ModUp Trace at `LogSlots=12` is an identity-sized trace in this fixture; its measured median is 8.998 µs. EvalMod’s polynomial input is Level 12/rows 4; the degree-30 polynomial consumes five levels to Level 7, and the three DoubleAngle rounds run from Levels 7, 6, and 5, producing Levels 6, 5, and 4, all at rows 4.

## End-to-end and stage timing

Matched historical q0=55 public-API results from QPREFIX-IMPL-009:

| Batch | Baseline | Candidate | Ratio |
|---:|---:|---:|---:|
| Count 1 | 30.670 ms | 116.390 ms | 3.79x |
| Count 3 | 89.891 ms | 348.166 ms | 3.87x |

Fresh current q0=56 candidate P93 full Bootstrap, 3 runs each (raw `ns/op`, `B/op`, `allocs/op`):

| Batch | Raw latency runs (ms) | Median | Raw B/op | Raw allocs/op |
|---:|---|---:|---|---|
| Count 1 | 112.410334, 108.298250, 108.622875 | 108.623 ms | 7,373,264; 7,372,888; 7,373,008 | 12,905; 12,899; 12,902 |
| Count 3 | 334.322708, 336.297750, 337.708709 | 336.298 ms | 22,119,056; 22,118,680; 22,122,288 | 38,697; 38,691; 38,707 |

The fresh q0=56 run is a distinct parameter profile; it is not substituted into the matched q0=55 ratio above.

Current q0=56 stage microbenchmarks (`GOMAXPROCS=10`, five 50 ms runs; raw times shown in ms, except µs-marked entries):

| Stage | Median | Five raw runs | B/op; allocs/op |
|---|---:|---|---:|
| PackN1ToN2 | 5.286 µs | 5.366, 5.286, 6.269, 4.532, 4.717 µs | 132,096; 15 |
| ScaleDown | 0.012621 | 0.012338, 0.013774, 0.012974, 0.012621, 0.012297 | 3,489; 79 |
| ModUp basis only | 0.429844 | 0.431845, 0.424401, 0.433285, 0.411411, 0.429844 | 395,808; 59 |
| Trace at production LogSlots | 8.998 µs | 8.998, 10.477, 8.364, 8.416, 9.017 µs | 0; 0 |
| ModUp total | 0.457352 | 0.457352, 0.457268, 0.449337, 0.569192, 0.516195 | 395,808; 59 |
| C2S total | 15.659958 | 16.776136, 15.484406, 24.886653, 15.659958, 15.565833 | 1,067,256; 575 |
| C2S Group 0 (Level 16, rows 4) | 3.526599 | 3.576615, 3.526599, 3.495039, 3.545357, 3.487315 | 4,096; 123 |
| C2S Group 1 (Level 15, rows 4) | 4.038140 | 3.918857, 4.038140, 4.130723, 4.055485, 3.978190 | 4,576; 143 |
| C2S Group 2 (Level 14, rows 4) | 3.983719 | 3.942051, 3.917539, 3.983719, 4.011143, 4.036649 | 3,960; 117 |
| C2S Group 3 (Level 13, rows 4) | 3.939033 | 3.987342, 3.939033, 3.903851, 3.908649, 4.607929 | 3,896; 117 |
| Polynomial P93 | 32.449750 | 33.208020, 31.520917, 34.536521, 32.449750, 32.096979 | 1,820,064 median; 4,808–4,811 |
| DoubleAngle round 0 | 2.599353 | 2.602627, 2.575133, 2.539989, 2.599353, 2.637314 | 1,728; 63 |
| DoubleAngle round 1 | 2.770774 | 2.789770, 2.771921, 2.770774, 2.742002, 2.769881 | 1,728; 63 |
| DoubleAngle round 2 | 2.782191 | 2.777198, 2.746694, 2.782191, 2.918733, 2.999682 | 1,728; 63 |
| EvalMod real total | 40.518208 | 51.951083, 40.084479, 40.518208, 40.035979, 41.133125 | 2,418,888–2,419,264; 5,778–5,784 |
| EvalMod imag total | 40.653354 | 40.653354, 40.036688, 40.111396, 45.713458, 40.904396 | 2,418,888–2,419,076; 5,778–5,781 |
| S2C total | 18.980208 | 15.789417, 31.962861, 18.980208, 20.504573, 17.112861 | 541,392–541,426; 572 |
| S2C Group 0 (Level 4, rows 4) | 5.471842 | 5.471842, 5.013090, 5.590804, 5.258450, 5.810227 | 6,152; 199 |
| S2C Group 1 (Level 3, rows 4) | 4.789712 | 9.309375, 8.783417, 4.789712, 4.675344, 4.648417 | 137,160; 201 |
| S2C Group 2 (Level 2, rows 3) | 2.783555 | 2.561892, 2.584996, 2.783555, 2.888396, 3.099562 | 134,424; 119 |
| Unpack N2ToN1 | 10.982 µs | 9.058, 10.982, 12.688, 11.834, 10.258 µs | 263,168; 16 |
| Public finalization | 11.039 µs | 10.972, 11.039, 10.985, 11.086, 11.111 µs | 0; 0 |

For an additive stage check, sum only the non-overlapping public stages: pack + ScaleDown + ModUp total + C2S total + EvalMod real + EvalMod imag + S2C total + unpack + finalization. The sum of their medians is **116.309 ms**, 7.686 ms (7.08%) above the fresh full-bootstrap Count-1 median (108.623 ms). The totals are separately calibrated microbenchmarks, not a serialized trace; this positive drift is retained as measurement variance. Do **not** add ModUp basis/Trace to ModUp total, DFT groups to C2S/S2C totals, or polynomial/DoubleAngle to EvalMod totals. These are nested subcomponent timings. The C2S group median sum is 15.488 ms versus 15.660 ms total. The S2C group timings cover one real-path group input; S2C total also combines the real/imaginary path and is not the sum of those three group measurements.

## Rows 2/3/4 capability scaling

All three explicit widths were measured on the current candidate using the same LogN13 Level-3 small signed coefficient fixture. Rows 2/3/4 are **capability measurements**, not an alternative production policy. Timings are medians in µs; each cell also gives B/op and allocs/op. Ratios are derived from medians; the expected row-linear reference is 2.00x for rows4/rows2.

| Operation | rows2 median (B/op, allocs) | rows3 median (B/op, allocs) | rows4 median (B/op, allocs) | T3/T2 | T4/T3 | T4/T2 |
|---|---:|---:|---:|---:|---:|---:|
| Add | 9.225 (0,0) | 13.753 (0,0) | 18.495 (0,0) | 1.49 | 1.34 | 2.00 |
| Sub | 10.564 (0,0) | 15.980 (0,0) | 20.876 (0,0) | 1.51 | 1.31 | 1.98 |
| Integer Mul | 13.426 (48,6) | 19.703 (72,9) | 27.112 (96,12) | 1.47 | 1.38 | 2.02 |
| Scalar Mul | 14.326 (920,35) | 20.889 (960,40) | 28.177 (1,000,45) | 1.46 | 1.35 | 1.97 |
| Pointwise Mul | 38.041 (64,1) | 60.392 (64,1) | 77.289 (64,1) | 1.59 | 1.28 | 2.03 |
| MulRelin | 28.794 (64,1) | 49.195 (64,1) | 58.217 (64,1) | 1.71 | 1.18 | 2.02 |
| Relinearize | 3.449 (0,0) | 5.907 (0,0) | 8.723 (0,0) | 1.71 | 1.48 | 2.53 |
| Rescale | 1,064.696 (664,18) | 1,867.957 (664,18) | 2,596.169 (664,18) | 1.75 | 1.39 | 2.44 |
| Rotate | 21.022 (168,8) | 32.326 (168,8) | 41.163 (168,8) | 1.54 | 1.27 | 1.96 |
| Trace (LogSlots 10) | 76.664 (920,33) | 115.640 (1,000,36) | 156.307 (1,080,39) | 1.51 | 1.35 | 2.04 |
| Poly multiply + merge | 37.879 (64,1) | 60.486 (64,1) | 75.482 (64,1) | 1.60 | 1.25 | 1.99 |
| One-bit scalar guard (isolated) | 664.063 (2,384,76) | 1,162.580 (2,448,84) | 1,769.999 (2,512,92) | 1.75 | 1.52 | 2.67 |
| LinearTransform representative BSGS | 295.435 (1,984,77) | 456.989 (1,984,77) | 606.949 (1,984,77) | 1.55 | 1.33 | 2.05 |
| N1→N2, LogN12→13 | 220.392 (393,408,12) | 322.544 (590,144,16) | 449.744 (786,816,20) | 1.46 | 1.39 | 2.04 |
| N2→N1, LogN13→12 | 237.784 (393,408,12) | 331.379 (590,144,16) | 438.106 (786,816,20) | 1.39 | 1.32 | 1.84 |
| Pack N1/N2, batch 3 | 567.244 (3,678,048,154) | 793.582 (4,071,520,162) | 1,046.210 (4,464,864,170) | 1.40 | 1.32 | 1.84 |
| Unpack N2/N1, batch 3 | 547.290 (3,153,920,156) | 791.359 (3,547,392,164) | 1,014.696 (3,940,736,172) | 1.45 | 1.28 | 1.85 |

Five raw `ns/op` runs per operation, rows2 / rows3 / rows4 (same order as the table; no run dropped):

| Operation | rows2 raw | rows3 raw | rows4 raw |
|---|---|---|---|
| Add | 10,202; 9,153; 9,182; 9,468; 9,225 | 13,854; 13,753; 13,830; 13,753; 13,695 | 18,495; 18,332; 18,272; 18,687; 18,622 |
| Sub | 10,601; 10,520; 10,594; 10,564; 10,532 | 15,809; 15,950; 15,980; 16,117; 16,428 | 21,072; 20,876; 20,950; 20,758; 20,739 |
| Integer Mul | 14,000; 13,426; 13,811; 13,282; 13,196 | 19,703; 19,775; 19,655; 19,614; 23,918 | 26,735; 26,833; 27,112; 27,479; 28,814 |
| Scalar Mul | 14,253; 14,326; 14,782; 14,568; 14,171 | 20,803; 20,889; 20,585; 21,236; 22,307 | 30,130; 28,177; 27,644; 27,539; 36,033 |
| Pointwise Mul | 38,041; 37,863; 38,622; 38,059; 37,868 | 61,743; 60,392; 60,707; 59,110; 59,805 | 76,978; 76,289; 77,514; 77,289; 77,681 |
| MulRelin | 28,480; 28,630; 29,021; 28,794; 28,945 | 51,854; 58,271; 44,478; 44,884; 49,195 | 58,217; 57,848; 58,445; 58,218; 57,933 |
| Relinearize | 3,443; 3,527; 3,449; 3,391; 3,519 | 5,700; 5,673; 5,936; 5,907; 6,832 | 9,288; 8,731; 8,676; 8,510; 8,723 |
| Rescale | 1,043,472; 1,064,696; 1,179,709; 1,052,768; 1,098,063 | 1,900,569; 1,875,402; 1,830,030; 1,867,957; 1,808,676 | 2,574,658; 2,687,638; 2,561,371; 2,596,169; 2,654,254 |
| Rotate | 21,260; 21,241; 21,022; 20,692; 20,681 | 32,326; 38,211; 32,299; 31,849; 34,863 | 45,799; 40,723; 40,681; 41,177; 41,163 |
| Trace | 76,664; 77,904; 77,237; 76,336; 75,765 | 117,899; 115,000; 113,764; 115,640; 116,777 | 155,779; 158,703; 156,307; 157,550; 153,645 |
| Poly multiply + merge | 57,043; 38,684; 37,762; 37,831; 37,879 | 57,298; 81,811; 63,881; 60,486; 57,172 | 75,162; 75,681; 75,482; 75,074; 75,768 |
| One-bit scalar guard | 665,675; 672,478; 664,063; 658,228; 658,854 | 1,162,580; 1,149,041; 1,378,601; 1,164,620; 1,134,740 | 1,769,999; 1,761,708; 1,754,970; 1,770,742; 1,804,389 |
| LinearTransform BSGS | 296,951; 295,136; 295,435; 294,781; 321,456 | 444,982; 456,989; 540,164; 455,557; 457,560 | 615,429; 606,949; 602,525; 605,653; 620,027 |
| N1→N2 | 232,761; 220,392; 230,246; 216,843; 218,348 | 336,728; 328,346; 319,393; 322,544; 320,583 | 449,744; 426,775; 518,171; 437,457; 465,832 |
| N2→N1 | 260,619; 237,784; 227,587; 243,098; 236,234 | 330,950; 342,887; 347,347; 330,296; 331,379 | 438,196; 438,918; 432,412; 438,106; 428,937 |
| Pack batch 3 | 656,380; 573,828; 566,119; 552,252; 567,244 | 1,030,578; 788,130; 798,459; 792,953; 793,582 | 1,178,897; 1,046,210; 1,113,584; 1,037,238; 1,019,516 |
| Unpack batch 3 | 537,407; 546,972; 576,309; 558,227; 547,290 | 812,840; 791,021; 791,007; 865,426; 791,359 | 1,492,229; 1,022,507; 1,008,941; 1,014,696; 1,006,358 |

Most rows4/rows2 kernel ratios cluster around 1.84–2.05. The clear super-linear capability outliers are Rescale (2.44x), standalone Relinearize (2.53x), and the isolated scalar guard (2.67x). The guard is **not active** in the current public P93 path: `circuits/ckks/mod1/fast.go` calls the unguarded `EvaluateWithPlanScaleQPrefixRows`; current production guard call count is zero. The 2.67x guard number is capability-only and contributes zero to this bootstrap’s measured latency.

## Rescale attribution

This is a same-candidate, same-N, same-small-signed-coefficient-distribution comparison using explicit rows2/3/4; it models q01/q012/q0123 arithmetic widths but is not a checkout of three historical implementations. Five raw runs are `ns/op`; substeps marked “single component” operate on one ciphertext polynomial. `PreflightPass` and `MaterializationPass` are separate internal passes and overlap conceptually with the full call; do not sum them with full Rescale.

| Rescale phase | rows2 median | rows3 median | rows4 median | rows4/rows2 |
|---|---:|---:|---:|---:|
| INTT + IMForm (single component) | 72.791 µs | 127.890 µs | 151.637 µs | 2.08x |
| Centered CRT reconstruction (single component) | 76.087 µs | 185.559 µs | 334.286 µs | 4.39x |
| Rounded division (single component) | 26.178 µs | 26.253 µs | 27.132 µs | 1.04x |
| Rematerialize residues + NTT/MForm (single component) | 174.990 µs | 250.757 µs | 248.350 µs | 1.42x |
| Preflight pass (single component) | 200.571 µs | 322.402 µs | 514.863 µs | 2.57x |
| Materialization pass (single component) | 351.988 µs | 569.136 µs | 763.123 µs | 2.17x |
| Full two-component ciphertext Rescale | 1.251902 ms | 1.807355 ms | 2.644862 ms | 2.11x |

Raw five-run `ns/op`, rows2 / rows3 / rows4:

| Phase | rows2 raw | rows3 raw | rows4 raw |
|---|---|---|---|
| INTT + IMForm | 72,948; 73,037; 72,786; 72,643; 72,791 | 139,917; 127,890; 127,474; 130,366; 122,671 | 228,384; 149,716; 153,881; 151,637; 151,537 |
| Centered CRT | 77,540; 76,046; 74,924; 76,087; 77,448 | 185,559; 201,528; 187,222; 183,365; 185,436 | 477,303; 328,339; 336,580; 334,286; 333,102 |
| Rounded division | 26,117; 26,492; 26,261; 25,961; 26,178 | 26,751; 26,744; 26,253; 26,148; 26,015 | 27,132; 26,724; 27,255; 26,702; 28,236 |
| Rematerialize + NTT/MForm | 164,081; 179,969; 174,006; 174,990; 179,859 | 233,922; 267,265; 295,428; 250,757; 248,301 | 261,695; 242,398; 255,095; 248,350; 243,630 |
| Preflight pass | 202,195; 292,656; 200,135; 200,571; 200,157 | 322,402; 321,843; 321,492; 322,735; 324,079 | 509,307; 508,660; 514,863; 530,889; 536,747 |
| Materialization pass | 385,319; 351,826; 347,618; 366,618; 351,988 | 567,036; 567,578; 570,304; 569,136; 643,560 | 772,784; 763,123; 764,244; 759,676; 752,308 |
| Full Rescale | 1,139,256; 1,213,742; 1,251,902; 1,312,495; 1,349,224 | 1,810,088; 1,802,563; 1,798,880; 1,807,355; 1,957,177 | 2,508,913; 2,625,408; 2,651,258; 2,644,862; 2,683,521 |

The four-row CRT kernel itself grows 4.39x against two rows, while full Rescale grows about 2.11x in this isolated substep harness (2.44x in the independent width-scaling harness). The repeated conversion/reconstruction/division/rematerialization passes are material: row4 preflight+materialization consume about 1.278 ms per component, before accounting for call orchestration. Division alone is nearly width-insensitive. This is evidence for a Rescale/CRT cost center, not proof that CRT arithmetic by itself explains the full-bootstrap regression.

## EvalMod / PS operation-count attribution

The active LogN13 P93 profile uses polynomial degree 30, polynomial depth 5, `planScale=2^93`, Level 12 input, rows 4. An overlay-only phase benchmark used the same CosDiscrete degree-30 coefficient-generation formula and sparse even-index pattern as the P93 profile, nominal `LogQ` input schedule `[56,39,39,39,39,60×8,56×4]`, and deterministic small signed coefficient residues converted to NTT/Montgomery form. The independently generated NTT primes are fixture values, not a byte-identical copy of the bootstrap parameter object. This is a schedule/kernel microfixture, not an encrypted semantic input. Actual encoded-input stage timings are reported above.

| PS phase | Median | Five raw runs | B/op; allocs/op |
|---|---:|---|---:|
| Generated powers | 14.815687 ms | 14.999948, 14.815687, 14.691468, 14.651250, 15.319928 ms | 47,696–47,730; 1,728 |
| PS baby steps | 530.243 µs | 561.193, 540.264, 522.848, 530.243, 517.224 µs | 13,376–13,392; 534 |
| PS giant-step merges | 8.245119 ms | 8.226940, 8.306512, 8.245119, 8.239661, 8.321744 ms | 1,127,824; 820 |
| final PS Rescale (no final RLWE relinearization in this fixture) | 2.027112 ms | 2.087107, 1.936667, 2.027112, 2.045671, 1.981742 ms | 672–680; 18 |
| full polynomial evaluator | 25.998230 ms | 26.678208, 25.963000, 25.355104, 25.998230, 32.334167 ms | 1,820,040–1,820,348; 4,808–4,812 |

The four phase medians sum to 25.618161 ms, 0.380069 ms (1.46%) below the full evaluator median. The small residual is plan/capacity observation, result ownership clone, and benchmark boundary overhead; phase boundaries do not materially overlap.

Operation counts from the exact P93 plan/source:

| Phase | Operation count per EvalMod |
|---|---|
| Power generation | 6 generated powers `{2,3,4,6,8,16}`; 6 `MulRelin`, 6 Chebyshev doubling `Add`, 6 `Rescale`; 5 subtract-one `AddScalar`, plus 1 aligned subtraction for T3. The T3 alignment can include an integer scale-alignment multiply if metadata scales differ. |
| PS baby | 5 blocks with degrees `[2,3,7,7,7]`; 5 workspace `zeroQPrefix`; 11 nonconstant `MulThenAdd` (all 11 coefficients non-integer); 5 constant `AddScalar`; guard operations 0. |
| PS giant | 4 merges; each performs 1 `Rescale`, 1 pointwise `Mul`, and 1 aligned `Add`; no RLWE ciphertext relinearization in these four steps. |
| Final PS | 1 final `Rescale`; 0 final ciphertext relinearizations for this degree-one result; one output Q-prefix clone/copy. |
| DoubleAngle | 3 rounds per EvalMod; each round performs 1 `MulRelin`, 1 integer `Mul`, 1 scalar `Add`, and 1 `Rescale`. |

Thus the polynomial + DoubleAngle path has 14 `Rescale` calls per EvalMod (11 in power/PS plus 3 DoubleAngle), 6 power `MulRelin`, 11 PS `MulThenAdd`, 4 giant-step pointwise `Mul`, 4 giant `Add`, 5 baby constant `AddScalar`, and 3 DoubleAngle `MulRelin` + 3 integer `Mul` + 3 scalar `Add`. The direct public EvalMod calls this once for real and once for imaginary. Multiplying 14 by the isolated rows4 full-Rescale median (~2.60 ms) gives a rough 36 ms per channel, near the observed ~40.5 ms; this is a cross-harness upper-order estimate, not an exact per-call profiler attribution.

The plan has 16 nonzero even Chebyshev coefficients at indices 0,2,…,30, all non-integer. Generated power map is `{1,2,3,4,6,8,16}`. No full arrays or coefficient dumps are included.

## Memory / allocation evidence

- Full q0=56 Count-1 bootstrap: 7.373 MB/op, ~12.9k allocs/op; Count-3: 22.119 MB/op, ~38.7k allocs/op.
- One EvalMod: ~2.419 MB/op, 5,778–5,784 allocs/op.
- Polynomial evaluator: ~1.820 MB/op, 4,808–4,812 allocs/op; PS giant phase alone: 1.128 MB/op, 820 allocs/op.
- C2S: ~1.067 MB/op, 575 allocs/op; S2C: ~0.541 MB/op, 572 allocs/op.
- Explicit N1→N2 / N2→N1 capability conversion grows 393,408→590,144→786,816 B/op for rows2/3/4. Batch-3 Pack grows 3.678→4.072→4.465 MB/op; batch-3 Unpack grows 3.154→3.547→3.941 MB/op.
- The 4-row ciphertext row payload is `N*8=65,536` bytes per row at N=8192; one polynomial across four Q rows is 262,144 bytes, and a degree-one ciphertext’s two polynomials are 524,288 bytes. This is the approximate row-payload size; it excludes metadata and temporary CRT/scratch storage.
- Workspace source-level traffic includes one Q-prefix input copy into `x1`, reused power/baby buffers, five baby-output zero operations, one returned Q-prefix result clone, and scale-alignment scratch only where scales differ. There is no retained per-operation full trace or heap profile.

The current Q-prefix runs do allocate/copy materially, but there is no same-machine matched q0=55 baseline `B/op`/copy counter, and the current Copy/Pack stages are small relative to EvalMod. The available evidence therefore does not support classifying memory traffic as the dominant explanation. Do not infer “zero copy cost” from this: its *incremental baseline contribution* is unmeasured.

## Slowdown decomposition (first-order estimate only)

Observed matched ratio is 3.79x, i.e. 85.72 ms of additional latency over the 30.67 ms historical baseline. The measured rows4/rows2 medians are approximately 1.84–2.05x for most core kernels and 2.11–2.44x for full Rescale (depending on the isolated harness); centered CRT alone is 4.39x. A deliberately broad first-order envelope is:

| Term | Estimate of the 85.72 ms excess | Basis / caveat |
|---|---:|---|
| `T_row-linear` | about 30.67 ms (~36% of the excess) | A 2x width-only model applied to the 30.67 ms baseline. This is a kernel-level envelope, not a measured full-baseline replay. |
| `T_CRT/Rescale-extra` above row-linear | about 1.7–10.8 ms (~2–13%) | Uses full-Rescale rows4/rows2=2.11–2.44 versus the row-linear 2.0, with an assumed 50–80% Rescale-bearing share of the baseline. The share and parameter mix are not directly measured on baseline; range is intentionally wide. |
| `T_copy/allocation-extra` | not identified | Current B/op is measured; matched baseline B/op is unavailable. No numeric credit is taken. |
| `T_other + unresolved residual` | at least about 44–53 ms (~51–62%) before any unknown copy delta | Arithmetic remainder after the above range; it includes any copy/allocation delta, q0/profile differences, harness variance, and unmeasured implementation overhead. |

Consequently the row-width and CRT/Rescale evidence plausibly explains roughly 38–49% of the observed multiplicative excess, not “most” of the 3.79x ratio. A quantitative `T_candidate = T_row-linear + T_CRT/Rescale-extra + T_copy/allocation-extra + T_other` closure is not available from current evidence. This is why the selected status is `UNEXPLAINED_PERFORMANCE_REGRESSION`, while Rescale/CRT is the strongest measured hotspot in the current candidate.

## Optimization candidates (not implemented)

1. **Rescale preflight/materialization kernel** — `schemes/ckks/fast/rescale_qprefix.go:rescaleQPrefixComponent`. Evidence: rows4 full Rescale 2.645 ms; preflight+materialization are both full coefficient passes. Mechanism: reduce duplicate coefficient-domain reconstruction while preserving failure-before-mutation semantics. Potential: up to ~20–40% of isolated rows4 Rescale time if a safe pass can be removed/fused; projected end-to-end benefit is uncertain and not claimed. Correctness risk: high (transactional behavior, signed rounding, dormant-row/capacity guarantees). Preserves `QPrefixWidth(Level)`.
2. **Fixed-width centered CRT kernel** — `schemes/ckks/fast/q012.go:reconstructQPrefix` / `centeredQPrefix`. Evidence: rows4 centered CRT 334 µs per component versus 76 µs rows2 (4.39x), inside full Rescale 2.645 ms. Mechanism: improve fixed-width reconstruction/centering without changing active rows. Potential: low single-digit to ~10% end-to-end only if it improves repeated Rescale calls; needs matched profiling. Correctness risk: very high (CRT canonical interval and signed residue exactness). Preserves `QPrefixWidth(Level)`.
3. **Polynomial workspace ownership/copy path** — `circuits/ckks/polynomial/fast.go:reset`, baby buffers, `cloneQPrefixResult`. Evidence: 1.820 MB and ~4.8k allocations per polynomial evaluation; giant phase 1.128 MB. Mechanism: reduce scratch/clone allocation while retaining independent public result ownership. Potential: likely low single-digit end-to-end; no matched baseline memory proof. Correctness risk: aliasing and reuse lifetime. Preserves `QPrefixWidth(Level)`.

No candidate was implemented, no parameter or schedule was changed, and no narrower production policy is proposed.

## Validation and limitations

Commands:

```bash
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test ./circuits/ckks/bootstrapping -run '^$' -bench '^(BenchmarkFastBootstrapManyLogN13P93Count1|BenchmarkFastBootstrapManyLogN13P93Count3)$' -benchmem -benchtime=100ms -count=3
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag_stage_overlay.json ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkQPDiagP93StageAttribution$' -benchmem -benchtime=50ms -count=5
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag_stage_overlay.json ./circuits/ckks/polynomial -run '^$' -bench '^BenchmarkQPDiagP93PSPhases$' -benchmem -benchtime=50ms -count=5 -v
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag_stage_overlay.json ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkQPDiagWidthScaling$' -benchmem -benchtime=50ms -count=5
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag_stage_overlay.json ./schemes/ckks/fast -run '^$' -bench '^BenchmarkQPDiagRescaleAttribution$' -benchmem -benchtime=50ms -count=5
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag_stage_overlay.json ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkQPDiagP93StageAttribution$' -benchmem -benchtime=1x -count=1 -v
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test ./schemes/ckks/fast ./circuits/ckks/bootstrapping -count=1
env GOCACHE=/private/tmp/qpdiag-go-cache-20260928 GOMAXPROCS=10 go test ./...
```

Results: Secondary focused tests passed (`schemes/ckks/fast`, `circuits/ckks/bootstrapping`). Primary `go test ./...` had exactly the already-documented unrelated failure `TestFIX001P3GenuineStandardPublicVsStagedConsistency` / `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT` (documented in `docs/CODEX_HANDOFF.md` and prior QPREFIX audit results); this task changed no code in that path and did not modify/bypass the test. Temporary overlay/test files were removed. No LogN16 or later experiment was run.

## Conclusion

Repeated four-row Rescale/CRT is the strongest measured cost center: 14 Rescales are scheduled per EvalMod, both EvalMod channels run, and each measured rows4 Rescale is about 2.1–2.4x the rows2 analog. Most other rows4 kernels are close to row-linear; the isolated one-bit guard is not active. The matched q0=55 historical baseline uses q01, whereas Q-prefix-v2 uses q0123 at high levels. Yet available width/Rescale data accounts for at most about half of the observed multiplicative excess under a broad estimate, and baseline allocation/stage data is missing. No causal optimization claim should be accepted without the independent Web review.

`READY_FOR_WEB_REVIEW`
