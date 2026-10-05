# FAST-STANDARD-NUMERICAL-FIX-001 — LogN13 q0=55

- FIX-001 classification: `FAST_STANDARD_EVALMOD_IMPROVED_BUT_NOT_RESTORED`
- Numerical harness classification: `FAST_NUMERICAL_QUALITY_DEGRADED`
- Timestamp: `2026-10-05T21:02:36Z`
- Threshold: `0.01`
- Primary: `bbdbebf741f3e4070e4f30bbd676ed8f64ec1835` (`main`, dirty=false)
- Secondary: `5c4fae3adb8b547a6da16463ba2071627982e156` (`fast-qprefix`, dirty=false)
- Environment: `go1.26.4`, `darwin/arm64`, CPU `arm64`, NumCPU `10`, GOMAXPROCS `10`

## FIX-001 repair outcome

The Fast-only normalized LogN13 production path and its special scale/DoubleAngle compensation were removed. Fast Mod1 now uses the generic Standard-equivalent normalization, target scale, Chebyshev offset, `y <- 2y²-c` recurrence, and Rescale progression, with Q-prefix operations retained. No key construction or sampled-error behavior was intentionally changed; the repair touched only Secondary `circuits/ckks/mod1/fast.go` and its formal degree-30 regression fixture in `circuits/ckks/mod1/fast_test.go`.

The canonical result is a partial numerical improvement, not a successful repair under the FIX-001 gates:

| Gate / metric | Before | After | Result |
|---|---:|---:|---|
| Final Fast-vs-Standard complex RMSE | 0.02089221106956351 | 0.02065572177060881 | Only 1.01145× improvement (1.13% reduction); required ≥1000×, so FAIL |
| Fast Bootstrap SNR | -0.17249085395397162 dB | -0.07361025788092074 dB | +0.0988805961 dB, but remains negative; FAIL |
| Standard Bootstrap SNR | 144.71339028083412 dB | 144.71339028083412 dB | Reference |
| EvalMod real Fast-vs-Standard RMSE / max | prior material divergence | 0.007326518157745949 / 0.020409129969526655 | Still order-1e-2 maximum divergence; FAIL |
| EvalMod imag Fast-vs-Standard RMSE / max | prior material divergence | 0.007279206148367247 / 0.02215177363770851 | Still order-1e-2 maximum divergence; FAIL |
| Public metadata contract | — | Level 1, degree 1, scale 2^45, NTT=true | PASS |
| Repeated Fast determinism | — | Bitwise identical; max difference 0 | PASS |

The first outer material checkpoint is `evalmod_real` (RMSE 7.3265e-3) after C2S real agrees at 2.3964e-14 RMSE. The direct stage replay reports matching Level/Scale progression: polynomial output RMSE 2.7227e-8, DoubleAngle round 2 after Rescale 2.2359e-7, then EvalMod output before public-scale reset 2.2895e-4 and after reset 7.3265e-3. The imaginary branch follows the same pattern (2.7051e-8, 2.2214e-7, 2.2748e-4, 7.2792e-3). S2C/final output reaches RMSE 2.0656e-2. These measurements locate where the observed divergence grows; they do not by themselves establish a causal explanation.

The Standard-equivalent schedule passed all **56/56** Q-prefix capacity checks (28 per real/imag branch), each using 4 authoritative rows and exact prefix product `5986308565615587353347023386369277282933624144412673`; every `strict 2B < S_Q(level)` check is true. The largest observed component bound is `2385407501657849389216817951508074643`, at `imag/ps-baby-3`, logical level 9. No capacity failure blocked this schedule.

Source changes and verification:

- Secondary production/test files: `circuits/ckks/mod1/fast.go`, `circuits/ckks/mod1/fast_test.go`; source audit found no remaining `normalizedLogN13` production dispatch/helper or its special compensation.
- Primary diagnostic files: `cmd/fastdiag/numerical.go`, `cmd/fastdiag/numerical_evalmod.go`, `cmd/fastdiag/numerical_lockstep.go`, `cmd/fastdiag/numerical_model.go`, `cmd/fastdiag/numerical_test.go`.
- Secondary: `go test ./schemes/ckks/fast -count=1`; `go test ./circuits/ckks/mod1 -count=1`; `go test ./circuits/ckks/bootstrapping -run 'Fast' -count=1` — all passed.
- Primary: `go test ./cmd/fastdiag ./internal/numericalmetrics` — passed.
- Canonical validation: `go run ./cmd/fastdiag numerical --standard-trials=3 --out results/FAST-STANDARD-NUMERICAL-FIX-001-summary.json` — completed; real/imag direct replay verified against public EvalMod with zero replay RMSE.
- `git diff --check` — passed. The numerical artifact records clean provenance: Primary `bbdbebf741f3e4070e4f30bbd676ed8f64ec1835`, Secondary `5c4fae3adb8b547a6da16463ba2071627982e156`.

The exact per-stage RMSE, metadata, and every capacity observation follow below. All three numerical correctness gates above remain unmet, so this result is handed off for independent review without inventing a further arithmetic workaround.

## Canonical workload

| Field | Value |
|---|---:|
| Input SHA-256 | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` |
| Slots | 4096 |
| LogN / LogSlots | 13 / 12 |
| q0 bits | 55 |
| Q-chain bits | `[55 39 40 39 40 60 60 61 60 60 60 61 61 56 57 56 56]` |
| P bits | `[61 61 62 61 62]` |
| Mod1 degree / DoubleAngle | 30 / 3 |
| EvalMod log scale / LogMessageRatio | 60 / 10 |
| Fast executions / effective numerical trials / Standard key trials | 2 / 1 / 3 |

Input is the existing deterministic plaintext-like construction `c0=encoded message, c1=0`; both Fast and every Standard trial receive the same encoded ciphertext coefficients.

## Output metadata

| Output | Level | Degree | Scale (log2) | IsNTT | IsMontgomery | LogDimensions | Public contract |
|---|---:|---:|---:|---|---|---|---|
| Fast 1 | 1 | 1 | 45.000000 | true | false | (0,12) | PASS |
| Fast 2 | 1 | 1 | 45.000000 | true | false | (0,12) | PASS |
| Standard 1 | 1 | 1 | 45.000000 | true | false | (0,12) | PASS |
| Standard 2 | 1 | 1 | 45.000000 | true | false | (0,12) | PASS |
| Standard 3 | 1 | 1 | 45.000000 | true | false | (0,12) | PASS |

Fast repeated execution bitwise identical: **true**; maximum complex difference `0.000000e+00` at slot 0.

Standard key-trial path validation:

| Trial | Fresh keygen/evaluation keys | secret key LevelP | Unique secret | Evaluator path |
|---:|---|---:|---|---|
| 1 | true / true | 4 | true | Standard GenEvaluationKeys with P-capable secret key |
| 2 | true / true | 4 | true | Standard GenEvaluationKeys with P-capable secret key |
| 3 | true / true | 4 | true | Standard GenEvaluationKeys with P-capable secret key |

## Against original message

| Output | Component | MAE | RMSE | p50 | p95 | p99 | max | worst slot |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| Fast 1 | real | 1.670180e-02 | 1.929232e-02 | 1.614410e-02 | 3.152006e-02 | 3.216654e-02 | 3.516046e-02 | 0 |
| Fast 1 | imag | 6.399537e-03 | 7.380051e-03 | 6.268313e-03 | 1.202404e-02 | 1.260912e-02 | 1.573171e-02 | 0 |
| Fast 1 | complex | 1.875408e-02 | 2.065572e-02 | 1.849151e-02 | 3.237330e-02 | 3.366437e-02 | 3.851941e-02 | 0 |
| Fast 2 | real | 1.670180e-02 | 1.929232e-02 | 1.614410e-02 | 3.152006e-02 | 3.216654e-02 | 3.516046e-02 | 0 |
| Fast 2 | imag | 6.399537e-03 | 7.380051e-03 | 6.268313e-03 | 1.202404e-02 | 1.260912e-02 | 1.573171e-02 | 0 |
| Fast 2 | complex | 1.875408e-02 | 2.065572e-02 | 1.849151e-02 | 3.237330e-02 | 3.366437e-02 | 3.851941e-02 | 0 |
| Standard 1 | real | 3.585202e-10 | 4.468538e-10 | 3.044938e-10 | 8.730562e-10 | 1.117017e-09 | 1.583010e-09 | 2411 |
| Standard 1 | imag | 3.982300e-10 | 1.103340e-09 | 3.119795e-10 | 9.183425e-10 | 1.288455e-09 | 5.805451e-08 | 0 |
| Standard 1 | complex | 6.009943e-10 | 1.190394e-09 | 5.365932e-10 | 1.130107e-09 | 1.480100e-09 | 5.805496e-08 | 0 |
| Standard 2 | real | 3.585202e-10 | 4.468538e-10 | 3.044938e-10 | 8.730562e-10 | 1.117017e-09 | 1.583010e-09 | 2411 |
| Standard 2 | imag | 3.982300e-10 | 1.103340e-09 | 3.119795e-10 | 9.183425e-10 | 1.288455e-09 | 5.805451e-08 | 0 |
| Standard 2 | complex | 6.009943e-10 | 1.190394e-09 | 5.365932e-10 | 1.130107e-09 | 1.480100e-09 | 5.805496e-08 | 0 |
| Standard 3 | real | 3.585202e-10 | 4.468538e-10 | 3.044938e-10 | 8.730562e-10 | 1.117017e-09 | 1.583010e-09 | 2411 |
| Standard 3 | imag | 3.982300e-10 | 1.103340e-09 | 3.119795e-10 | 9.183425e-10 | 1.288455e-09 | 5.805451e-08 | 0 |
| Standard 3 | complex | 6.009943e-10 | 1.190394e-09 | 5.365932e-10 | 1.130107e-09 | 1.480100e-09 | 5.805496e-08 | 0 |
| Standard trial median | real | 3.585202e-10 | 4.468538e-10 | 3.044938e-10 | 8.730562e-10 | 1.117017e-09 | 1.583010e-09 | median of trial maxima |
| Standard trial median | imag | 3.982300e-10 | 1.103340e-09 | 3.119795e-10 | 9.183425e-10 | 1.288455e-09 | 5.805451e-08 | median of trial maxima |
| Standard trial median | complex | 6.009943e-10 | 1.190394e-09 | 5.365932e-10 | 1.130107e-09 | 1.480100e-09 | 5.805496e-08 | median of trial maxima |

### Worst-slot examples vs original

| Output | Component | Slot | Original | Output | Difference |
|---|---|---:|---|---|---|
| Fast 1 | real | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `0.035160459+0.015731715i` |
| Fast 1 | imag | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `0.035160459+0.015731715i` |
| Fast 1 | complex | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `0.035160459+0.015731715i` |
| Fast 2 | real | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `0.035160459+0.015731715i` |
| Fast 2 | imag | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `0.035160459+0.015731715i` |
| Fast 2 | complex | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `0.035160459+0.015731715i` |
| Standard 1 | real | 2411 | `0.0234375-0.001953125i` | `0.023437498-0.001953125i` | `-1.5830098e-09-3.8274081e-11i` |
| Standard 1 | imag | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.280182e-10-5.805451e-08i` |
| Standard 1 | complex | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.280182e-10-5.805451e-08i` |
| Standard 2 | real | 2411 | `0.0234375-0.001953125i` | `0.023437498-0.001953125i` | `-1.5830098e-09-3.8274081e-11i` |
| Standard 2 | imag | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.280182e-10-5.805451e-08i` |
| Standard 2 | complex | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.280182e-10-5.805451e-08i` |
| Standard 3 | real | 2411 | `0.0234375-0.001953125i` | `0.023437498-0.001953125i` | `-1.5830098e-09-3.8274081e-11i` |
| Standard 3 | imag | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.280182e-10-5.805451e-08i` |
| Standard 3 | complex | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.280182e-10-5.805451e-08i` |

## Fast vs Standard, per trial pair

| Fast | Standard | Component | MAE | RMSE | p95 | p99 | max | worst slot | Metadata |
|---:|---:|---|---:|---:|---:|---:|---:|---:|---|
| 1 | 1 | real | 1.670180e-02 | 1.929232e-02 | 3.152006e-02 | 3.216654e-02 | 3.516046e-02 | 0 | MATCH |
| 1 | 1 | imag | 6.399537e-03 | 7.380051e-03 | 1.202404e-02 | 1.260912e-02 | 1.573177e-02 | 0 | MATCH |
| 1 | 1 | complex | 1.875408e-02 | 2.065572e-02 | 3.237330e-02 | 3.366437e-02 | 3.851943e-02 | 0 | MATCH |
| 1 | 2 | real | 1.670180e-02 | 1.929232e-02 | 3.152006e-02 | 3.216654e-02 | 3.516046e-02 | 0 | MATCH |
| 1 | 2 | imag | 6.399537e-03 | 7.380051e-03 | 1.202404e-02 | 1.260912e-02 | 1.573177e-02 | 0 | MATCH |
| 1 | 2 | complex | 1.875408e-02 | 2.065572e-02 | 3.237330e-02 | 3.366437e-02 | 3.851943e-02 | 0 | MATCH |
| 1 | 3 | real | 1.670180e-02 | 1.929232e-02 | 3.152006e-02 | 3.216654e-02 | 3.516046e-02 | 0 | MATCH |
| 1 | 3 | imag | 6.399537e-03 | 7.380051e-03 | 1.202404e-02 | 1.260912e-02 | 1.573177e-02 | 0 | MATCH |
| 1 | 3 | complex | 1.875408e-02 | 2.065572e-02 | 3.237330e-02 | 3.366437e-02 | 3.851943e-02 | 0 | MATCH |

### Fast-vs-Standard worst-slot examples

| Fast | Standard | Component | Slot | Original | Fast | Standard | Fast−Standard |
|---:|---:|---|---:|---|---|---|---|
| 1 | 1 | real | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 1 | imag | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 1 | complex | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 2 | real | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 2 | imag | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 2 | complex | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 3 | real | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 3 | imag | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |
| 1 | 3 | complex | 0 | `-0.03125-0.01171875i` | `0.003910459+0.0040129648i` | `-0.03125-0.011718808i` | `0.035160459+0.015731773i` |

## Precision bits

Precision is computed per slot as `-log2(max(|output-original|, 1e-30))` using complex magnitude error; quantiles use linear interpolation over sorted samples.

| Output | Mean | Median | p05 | Minimum | Minimum slot | Slot error | CKKS helper median L2 |
|---|---:|---:|---:|---:|---:|---:|---:|
| Fast 1 | 5.9479 | 5.7570 | 4.9491 | 4.6983 | 0 | 3.851941e-02 | 5.4528 |
| Fast 2 | 5.9479 | 5.7570 | 4.9491 | 4.6983 | 0 | 3.851941e-02 | 5.4528 |
| Standard 1 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |
| Standard 2 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |
| Standard 3 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |

Median across Fast trials: `5.7570` bits; median across Standard trial medians: `30.7955` bits (drop `25.0385` bits).

## Standard-to-Standard variability

- Per-slot Standard mean was computed across all trials; SHA-256: `0af5cdd54baa7f9eb28d53f5048a8856e1dfe15e3473057cc11626eab0479df0`. Sample means: slot 0=(-3.125000e-02-1.171881e-02i); slot 2048=(-4.749342e-10+3.906249e-03i)
- RMS spread around per-slot means (real / imag / complex): `0.000000e+00` / `0.000000e+00` / `0.000000e+00`
- Maximum spread from per-slot mean: `0.000000e+00` at slot 0 (trial 1)
- Maximum pairwise Standard difference: `0.000000e+00` at slot 0 (trials 1 vs 2)

## 1e-2 threshold audit

| Comparison | Real > threshold | Imag > threshold | Combined fraction |
|---|---:|---:|---:|
| Fast vs original (1 comparisons) | 2890 / 4096 (70.556641%) | 857 / 4096 (20.922852%) | 3747 / 8192 (45.739746%) |
| Standard vs original (3 comparisons) | 0 / 12288 (0.000000%) | 0 / 12288 (0.000000%) | 0 / 24576 (0.000000%) |
| Fast vs Standard (3 comparisons) | 8670 / 12288 (70.556641%) | 2571 / 12288 (20.922852%) | 11241 / 24576 (45.739746%) |

Median Fast complex RMSE: `2.065572e-02`; median Standard complex RMSE: `1.190394e-09`; median Fast-vs-Standard complex RMSE: `2.065572e-02`; max Fast-vs-Standard complex difference: `3.851943e-02`.

## Classification checks

- Fast vs original has no coordinate above 1e-2: `false`
- Fast vs Standard has no coordinate above 1e-2: `false`
- Median precision degradation is at most 2 bits: `false`
- Fast complex RMSE is within 4x Standard, or Standard RMSE is within the documented near-zero floor: `false`
- Public Bootstrap Level/Scale/dimensions contract is respected: `true`
- Standard near-zero floor: `1.421085e-14`

## Decoded-domain Bootstrap SNR

SNR here is numerical distortion, not RLWE security noise or a noise budget. Each Bootstrap SNR uses that mode's actual decoded pre-Bootstrap vector as the signal reference and its decoded post-Bootstrap output as the observation.

| Mode / trial | Signal power | Signal RMS | Error power | Pre→Post RMSE | Bootstrap SNR (dB / status) | Pre-Bootstrap vs canonical original RMSE |
|---|---:|---:|---:|---:|---:|---:|
| Standard 1 | 4.194882e-04 | 2.048141e-02 | 1.417035e-18 | 1.190393e-09 | 144.713390 / FINITE | 7.449975e-13 |
| Standard 2 | 4.194882e-04 | 2.048141e-02 | 1.417035e-18 | 1.190393e-09 | 144.713390 / FINITE | 7.449975e-13 |
| Standard 3 | 4.194882e-04 | 2.048141e-02 | 1.417035e-18 | 1.190393e-09 | 144.713390 / FINITE | 7.449975e-13 |
| Fast Q-prefix 1 | 4.194882e-04 | 2.048141e-02 | 4.266588e-04 | 2.065572e-02 | -0.073610 / FINITE | 7.449975e-13 |
| Fast Q-prefix 2 | 4.194882e-04 | 2.048141e-02 | 4.266588e-04 | 2.065572e-02 | -0.073610 / FINITE | 7.449975e-13 |

`STANDARD_BOOTSTRAP_SNR_DB=144.713390 / FINITE`
`FAST_BOOTSTRAP_SNR_DB=-0.073610 / FINITE`

## Stage reference SNR and divergence

`D_i` is Fast-vs-genuine-Standard complex RMSE. Amplification and SNR deltas follow the explicit stage topology, not table adjacency: C2S real/imag are parallel children of ModUp; each EvalMod branch uses its matching C2S branch; S2C uses a joint EvalMod real+imag reference; final public output may use S2C.

| Checkpoint | Fast state | Standard state | D_i RMSE | A_i parent | A_i | Stage reference SNR (dB / status) | ΔSNR_i parent | ΔSNR_i (dB / status) | Max complex diff |
|---|---|---|---:|---|---:|---:|---|---:|---:|
| input | L0 / 45.000000 / d1 / true / false / 1(+1 maintained) | L0 / 45.000000 / d1 / true / false / 1(+0 maintained) | 0.000000e+00 | — | — (NOT_COMPARABLE) | +Inf / POSITIVE_INFINITY | — | — (NOT_COMPARABLE) | 0.000000e+00 |
| scale_down | L0 / 45.000000 / d1 / true / false / 1(+1 maintained) | L0 / 45.000000 / d1 / true / false / 1(+0 maintained) | 0.000000e+00 | input | — (UNDEFINED_ZERO_OVER_ZERO) | +Inf / POSITIVE_INFINITY | input | — (UNDEFINED_INFINITY_MINUS_INFINITY) | 0.000000e+00 |
| mod_up | L16 / 50.000000 / d1 / true / true / 4(+4 maintained) | L16 / 50.000000 / d1 / true / false / 17(+0 maintained) | 0.000000e+00 | scale_down | — (UNDEFINED_ZERO_OVER_ZERO) | +Inf / POSITIVE_INFINITY | scale_down | — (UNDEFINED_INFINITY_MINUS_INFINITY) | 0.000000e+00 |
| c2s_real | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.396436e-14 | mod_up | — (UNDEFINED_ZERO_PREVIOUS_D) | 175.432910 / FINITE | mod_up | — (NEGATIVE_INFINITY) | 9.185075e-14 |
| c2s_imag | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.409568e-14 | mod_up | — (UNDEFINED_ZERO_PREVIOUS_D) | 175.358744 / FINITE | mod_up | — (NEGATIVE_INFINITY) | 9.768351e-14 |
| evalmod_real | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.326518e-03 | c2s_real | 3.057255e+11 (FINITE) | -0.088325 / FINITE | c2s_real | -1.755212e+02 (FINITE) | 2.040913e-02 |
| evalmod_imag | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.279206e-03 | c2s_imag | 3.020959e+11 (FINITE) | -0.058754 / FINITE | c2s_imag | -1.754175e+02 (FINITE) | 2.215177e-02 |
| s2c | L1 / 45.000000 / d1 / true / true / 2(+2 maintained) | L1 / 45.000000 / d1 / true / false / 2(+0 maintained) | 2.065572e-02 | — | n/a (combined branches) (NOT_APPLICABLE_COMBINED_BRANCH) | -0.073610 / FINITE | evalmod_real+evalmod_imag (joint) | -8.836126e-12 (FINITE) | 3.851943e-02 |
| final_public_output | L1 / 45.000000 / d1 / true / false / 2(+2 maintained) | L1 / 45.000000 / d1 / true / false / 2(+0 maintained) | 2.065572e-02 | s2c | 1.000000e+00 (FINITE) | -0.073610 / FINITE | s2c | 0.000000e+00 (FINITE) | 3.851943e-02 |

`LARGEST_RAW_AMPLIFICATION_CHECKPOINT=evalmod_real`
`LARGEST_RAW_AMPLIFICATION_FACTOR=3.057255e+11`
`LARGEST_SNR_DROP_CHECKPOINT=evalmod_real`
`LARGEST_SNR_DROP_DB=-1.755212e+02`

- First observable: `evalmod_real`, max complex diff `2.040913e-02` (threshold `1.000e-08`)
- First material: `evalmod_real`, max complex diff `2.040913e-02` (threshold `0.00364073039722`)
- Final Fast-vs-Standard RMSE: `0.0206557217706`
- Combined-branch EvalMod error RMSE: `7.302900e-03`
- Combined-branch S2C amplification: `2.828427e+00` (status `FINITE`; `D_s2c / RMSE(concat(EvalMod real, EvalMod imag))`)
- Current classification: `FAST_STANDARD_FIRST_MATERIAL_EVALMOD_DOUBLE_ANGLE`

`FIRST_OBSERVABLE_CHECKPOINT=evalmod_real`
`FIRST_OBSERVABLE_MAX_DIFF=2.040913e-02`
`FIRST_MATERIAL_CHECKPOINT=evalmod_real`
`FIRST_MATERIAL_MAX_DIFF=2.040913e-02`
`FINAL_FAST_STANDARD_RMSE=0.0206557217706`
`COMBINED_BRANCH_S2C_AMPLIFICATION_FACTOR=2.828427e+00`
`S2C_AMPLIFICATION_FACTOR=2.828427e+00 (combined-branch compatibility alias)`

## EvalMod direct stage-equivalence replay

Each Fast checkpoint is decoded from authoritative Q-prefix rows; Standard is decrypted with the genuine Standard secret key. Comparisons use the same mathematical scale directly, with no representation-specific power-of-two alignment. Replay verification compares the manual stage sequence with each actual public EvalMod output.

| Checkpoint | Fast L / log2(scale) / degree / NTT / Montgomery / rows | Standard L / log2(scale) / degree / NTT / Montgomery / rows | D_i RMSE | A_i | Stage reference SNR (dB / status) | ΔSNR_i (dB / status) | Max complex diff |
|---|---|---|---:|---:|---:|---:|---:|---:|
| real/input_before_normalization | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.396436e-14 | — (NO_PREVIOUS_CHECKPOINT) | 175.432910 / FINITE | — (NO_PREVIOUS_CHECKPOINT) | 9.185075e-14 |
| real/after_normalization | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.340270e-17 | 9.765625e-04 (FINITE) | 175.432910 / FINITE | 0.000000e+00 (FINITE) | 8.969800e-17 |
| real/after_chebyshev_offset | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.367438e-17 | 1.011609e+00 (FINITE) | 296.390827 / FINITE | 1.209579e+02 (FINITE) | 9.194034e-17 |
| real/polynomial_input | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.367438e-17 | 1.000000e+00 (FINITE) | 296.390827 / FINITE | 0.000000e+00 (FINITE) | 9.194034e-17 |
| real/polynomial_output | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.722707e-08 | 1.150064e+09 (FINITE) | 149.136012 / FINITE | -1.472548e+02 (FINITE) | 7.584522e-08 |
| real/before_double_angle_round_0 | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.722707e-08 | 1.000000e+00 (FINITE) | 149.136012 / FINITE | 0.000000e+00 (FINITE) | 7.584522e-08 |
| real/double_angle_round_0_after_multiply | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 4.244556e-08 | 1.558947e+00 (FINITE) | 143.115412 / FINITE | -6.020600e+00 (FINITE) | 1.182387e-07 |
| real/double_angle_round_0_after_add | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 8.489112e-08 | 2.000000e+00 (FINITE) | 143.115412 / FINITE | 0.000000e+00 (FINITE) | 2.364774e-07 |
| real/double_angle_round_0_after_constant | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 8.489112e-08 | 1.000000e+00 (FINITE) | 136.744162 / FINITE | -6.371251e+00 (FINITE) | 2.364774e-07 |
| real/double_angle_round_0_after_rescale | L6 / 60.000000 / d1 / true / true / 4(+4 maintained) | L6 / 60.000000 / d1 / true / false / 7(+0 maintained) | 8.489112e-08 | 1.000000e+00 (FINITE) | 136.744162 / FINITE | 4.667982e-10 (FINITE) | 2.364774e-07 |
| real/double_angle_round_1_after_multiply | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 9.907468e-08 | 1.167079e+00 (FINITE) | 130.723562 / FINITE | -6.020600e+00 (FINITE) | 2.759878e-07 |
| real/double_angle_round_1_after_add | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.981494e-07 | 2.000000e+00 (FINITE) | 130.723562 / FINITE | 0.000000e+00 (FINITE) | 5.519757e-07 |
| real/double_angle_round_1_after_constant | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.981494e-07 | 1.000000e+00 (FINITE) | 123.068048 / FINITE | -7.655514e+00 (FINITE) | 5.519757e-07 |
| real/double_angle_round_1_after_rescale | L5 / 60.000000 / d1 / true / true / 4(+4 maintained) | L5 / 60.000000 / d1 / true / false / 6(+0 maintained) | 1.981494e-07 | 1.000000e+00 (FINITE) | 123.068048 / FINITE | -5.549339e-11 (FINITE) | 5.519757e-07 |
| real/double_angle_round_2_after_multiply | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 1.117938e-07 | 5.641896e-01 (FINITE) | 117.047448 / FINITE | -6.020600e+00 (FINITE) | 3.114186e-07 |
| real/double_angle_round_2_after_add | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 2.235876e-07 | 2.000000e+00 (FINITE) | 117.047448 / FINITE | 0.000000e+00 (FINITE) | 6.228372e-07 |
| real/double_angle_round_2_after_constant | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 2.235876e-07 | 1.000000e+00 (FINITE) | -0.088325 / FINITE | -1.171358e+02 (FINITE) | 6.228372e-07 |
| real/double_angle_round_2_after_rescale | L4 / 60.000000 / d1 / true / true / 4(+4 maintained) | L4 / 60.000000 / d1 / true / false / 5(+0 maintained) | 2.235876e-07 | 1.000000e+00 (FINITE) | -0.088325 / FINITE | -1.988899e-11 (FINITE) | 6.228372e-07 |
| real/evalmod_output_before_public_scale_reset | L4 / 50.000000 / d1 / true / true / 4(+4 maintained) | L4 / 50.000000 / d1 / true / false / 5(+0 maintained) | 2.289537e-04 | 1.024000e+03 (FINITE) | -0.088325 / FINITE | 0.000000e+00 (FINITE) | 6.377853e-04 |
| real/evalmod_output_after_public_scale_reset | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.326518e-03 | 3.200000e+01 (FINITE) | -0.088325 / FINITE | 0.000000e+00 (FINITE) | 2.040913e-02 |
| imag/input_before_normalization | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.409568e-14 | — (NO_PREVIOUS_CHECKPOINT) | 175.358744 / FINITE | — (NO_PREVIOUS_CHECKPOINT) | 9.768351e-14 |
| imag/after_normalization | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.353093e-17 | 9.765625e-04 (FINITE) | 175.358744 / FINITE | 0.000000e+00 (FINITE) | 9.539405e-17 |
| imag/after_chebyshev_offset | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.387932e-17 | 1.014806e+00 (FINITE) | 296.315960 / FINITE | 1.209572e+02 (FINITE) | 9.540979e-17 |
| imag/polynomial_input | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.387932e-17 | 1.000000e+00 (FINITE) | 296.315960 / FINITE | 0.000000e+00 (FINITE) | 9.540979e-17 |
| imag/polynomial_output | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.705124e-08 | 1.132831e+09 (FINITE) | 149.192285 / FINITE | -1.471237e+02 (FINITE) | 8.232131e-08 |
| imag/before_double_angle_round_0 | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.705124e-08 | 1.000000e+00 (FINITE) | 149.192285 / FINITE | 0.000000e+00 (FINITE) | 8.232131e-08 |
| imag/double_angle_round_0_after_multiply | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 4.217146e-08 | 1.558947e+00 (FINITE) | 143.171685 / FINITE | -6.020600e+00 (FINITE) | 1.283346e-07 |
| imag/double_angle_round_0_after_add | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 8.434292e-08 | 2.000000e+00 (FINITE) | 143.171685 / FINITE | 0.000000e+00 (FINITE) | 2.566692e-07 |
| imag/double_angle_round_0_after_constant | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 8.434292e-08 | 1.000000e+00 (FINITE) | 136.800435 / FINITE | -6.371251e+00 (FINITE) | 2.566692e-07 |
| imag/double_angle_round_0_after_rescale | L6 / 60.000000 / d1 / true / true / 4(+4 maintained) | L6 / 60.000000 / d1 / true / false / 7(+0 maintained) | 8.434292e-08 | 1.000000e+00 (FINITE) | 136.800435 / FINITE | -8.859047e-11 (FINITE) | 2.566692e-07 |
| imag/double_angle_round_1_after_multiply | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 9.843488e-08 | 1.167079e+00 (FINITE) | 130.779835 / FINITE | -6.020600e+00 (FINITE) | 2.995532e-07 |
| imag/double_angle_round_1_after_add | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.968698e-07 | 2.000000e+00 (FINITE) | 130.779835 / FINITE | 0.000000e+00 (FINITE) | 5.991064e-07 |
| imag/double_angle_round_1_after_constant | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.968698e-07 | 1.000000e+00 (FINITE) | 123.124321 / FINITE | -7.655514e+00 (FINITE) | 5.991064e-07 |
| imag/double_angle_round_1_after_rescale | L5 / 60.000000 / d1 / true / true / 4(+4 maintained) | L5 / 60.000000 / d1 / true / false / 6(+0 maintained) | 1.968698e-07 | 1.000000e+00 (FINITE) | 123.124321 / FINITE | 3.132072e-11 (FINITE) | 5.991064e-07 |
| imag/double_angle_round_2_after_multiply | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 1.110719e-07 | 5.641896e-01 (FINITE) | 117.103721 / FINITE | -6.020600e+00 (FINITE) | 3.380092e-07 |
| imag/double_angle_round_2_after_add | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 2.221437e-07 | 2.000000e+00 (FINITE) | 117.103721 / FINITE | 0.000000e+00 (FINITE) | 6.760185e-07 |
| imag/double_angle_round_2_after_constant | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 2.221437e-07 | 1.000000e+00 (FINITE) | -0.058754 / FINITE | -1.171625e+02 (FINITE) | 6.760185e-07 |
| imag/double_angle_round_2_after_rescale | L4 / 60.000000 / d1 / true / true / 4(+4 maintained) | L4 / 60.000000 / d1 / true / false / 5(+0 maintained) | 2.221437e-07 | 1.000000e+00 (FINITE) | -0.058754 / FINITE | -6.269929e-12 (FINITE) | 6.760185e-07 |
| imag/evalmod_output_before_public_scale_reset | L4 / 50.000000 / d1 / true / true / 4(+4 maintained) | L4 / 50.000000 / d1 / true / false / 5(+0 maintained) | 2.274752e-04 | 1.024000e+03 (FINITE) | -0.058754 / FINITE | 0.000000e+00 (FINITE) | 6.922429e-04 |
| imag/evalmod_output_after_public_scale_reset | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.279206e-03 | 3.200000e+01 (FINITE) | -0.058754 / FINITE | 0.000000e+00 (FINITE) | 2.215177e-02 |

Replay verification RMSE by mode: `imag.fast=0.000000e+00 (verified=true)`; `imag.standard=0.000000e+00 (verified=true)`; `real.fast=0.000000e+00 (verified=true)`; `real.standard=0.000000e+00 (verified=true)`

## Standard-equivalent Fast EvalMod Q-prefix capacity audit

Bounds are exact observed centered coefficient maxima by ciphertext component. Every checkpoint requires strict `2B < S_Q`; execution stops at the first failed guard.

| Checkpoint | Level | Rows | Exact prefix product S_Q | Degree | MaxAbs by component | Strict `2B < S_Q` |
|---|---:|---:|---|---:|---|---:|
| real/evalmod-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[617940038 0]` | true |
| real/polynomial-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014399071389638 0]` | true |
| real/generated-power-1 | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014399071389638 0]` | true |
| real/generated-power-2 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1152358554585939901 0]` | true |
| real/generated-power-3 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[54025606050744262 0]` | true |
| real/generated-power-4 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1150670254550073347 0]` | true |
| real/generated-power-6 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1147858252521176499 0]` | true |
| real/generated-power-8 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1143925296175874037 0]` | true |
| real/generated-power-16 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1117077065111486493 0]` | true |
| real/ps-baby-0 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[145443220898665513360621908076 0]` | true |
| real/ps-baby-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[21627059961492905227666666683298 0]` | true |
| real/ps-baby-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[68144292540158768030199855671138822 0]` | true |
| real/ps-baby-3 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2385407501425870769186807017308299057 0]` | true |
| real/ps-baby-4 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1396742617111005567665963028050393156 0]` | true |
| real/ps-giant-0 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[970051628347912179981297018585441010 0 0]` | true |
| real/ps-giant-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[21772219182953280960878315289677 0 0]` | true |
| real/ps-giant-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[68165894871434360306094563946573437 0 0]` | true |
| real/ps-giant-3 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[1036098239049013362596824538197159012 0 0]` | true |
| real/polynomial-before-final-rescale | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1036098239049013362596824538197159012 0]` | true |
| real/polynomial-after-final-rescale | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969348404341 0]` | true |
| real/evalmod-polynomial-output | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969348404341 0]` | true |
| real/double-angle-0-before | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969348404341 0]` | true |
| real/double-angle-0-after-rescale | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775424730852156 0]` | true |
| real/double-angle-1-before | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775424730852156 0]` | true |
| real/double-angle-1-after-rescale | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233153285604551 0]` | true |
| real/double-angle-2-before | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233153285604551 0]` | true |
| real/double-angle-2-after-rescale | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2011786824 0]` | true |
| real/evalmod-output | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2011786824 0]` | true |
| imag/evalmod-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[745777719 0]` | true |
| imag/polynomial-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014398155867732 0]` | true |
| imag/generated-power-1 | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014398155867732 0]` | true |
| imag/generated-power-2 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1152358554653048787 0]` | true |
| imag/generated-power-3 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[54025602988787284 0]` | true |
| imag/generated-power-4 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1150670254550073347 0]` | true |
| imag/generated-power-6 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1147858253024329351 0]` | true |
| imag/generated-power-8 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1143925296175874037 0]` | true |
| imag/generated-power-16 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1117077065111486493 0]` | true |
| imag/ps-baby-0 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[145443220900143143012885635766 0]` | true |
| imag/ps-baby-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[21627059961654216038542900758120 0]` | true |
| imag/ps-baby-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[68144292541247008759964546778590094 0]` | true |
| imag/ps-baby-3 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2385407501657849389216817951508074643 0]` | true |
| imag/ps-baby-4 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1396742617571849720318535593306611848 0]` | true |
| imag/ps-giant-0 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[970051628117236525794663143096215341 0 0]` | true |
| imag/ps-giant-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[21772219183116893112263649511193 0 0]` | true |
| imag/ps-giant-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[68165894872522763473251312028137963 0 0]` | true |
| imag/ps-giant-3 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[1036098238819392272725661445305298570 0 0]` | true |
| imag/polynomial-before-final-rescale | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1036098238819392272725661445305298570 0]` | true |
| imag/polynomial-after-final-rescale | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969149239794 0]` | true |
| imag/evalmod-polynomial-output | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969149239794 0]` | true |
| imag/double-angle-0-before | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969149239794 0]` | true |
| imag/double-angle-0-after-rescale | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775424109878064 0]` | true |
| imag/double-angle-1-before | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775424109878064 0]` | true |
| imag/double-angle-1-after-rescale | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233151836152494 0]` | true |
| imag/double-angle-2-before | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233151836152494 0]` | true |
| imag/double-angle-2-after-rescale | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2011786720 0]` | true |
| imag/evalmod-output | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2011786720 0]` | true |

## Polynomial plan and generated-power evidence

- Fast diagnostic PS plan (real branch): degree `30`, base `8`, level `12`, target scale log2 `60.000000`, exact `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18`, blocks `5`.

| Branch | Power n | Level | Scale log2 | Fast maintained rows | RMSE vs reference | Max complex diff | Reference provenance |
|---|---:|---:|---:|---:|---:|---:|---|
| real | 1 | 12 | 60.000000 | 4 | 2.367438e-17 | 9.194034e-17 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| real | 2 | 11 | 60.000000 | 4 | 8.605597e-10 | 2.753504e-09 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| real | 3 | 10 | 60.000000 | 4 | 2.749948e-08 | 8.800716e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| real | 4 | 10 | 60.000000 | 4 | 3.456590e-09 | 9.801409e-09 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| real | 6 | 9 | 60.000000 | 4 | 7.733163e-09 | 2.484335e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| real | 8 | 9 | 60.000000 | 4 | 1.379921e-08 | 3.912483e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| real | 16 | 8 | 60.000000 | 4 | 5.476583e-08 | 1.552693e-07 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 1 | 12 | 60.000000 | 4 | 2.387932e-17 | 9.540979e-17 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 2 | 11 | 60.000000 | 4 | 8.582712e-10 | 2.644256e-09 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 3 | 10 | 60.000000 | 4 | 2.742569e-08 | 8.450668e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 4 | 10 | 60.000000 | 4 | 3.445871e-09 | 1.039698e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 6 | 9 | 60.000000 | 4 | 7.706842e-09 | 2.410964e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 8 | 9 | 60.000000 | 4 | 1.375666e-08 | 4.150245e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| imag | 16 | 8 | 60.000000 | 4 | 5.459747e-08 | 1.647056e-07 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |

## EvalMod exact-scale audit

Exact scale strings are retained alongside the observed logical Level at each direct Standard/Fast replay checkpoint. The target scale follows the Standard Mod1 Q schedule.

| Checkpoint | Level | Scale log2 | Exact scale | Target log2 / exact | DoubleAngle round |
|---|---:|---:|---|---|---:|
| evalmod_real/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| evalmod_real/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| evalmod_imag/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| evalmod_imag/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| evalmod/standard_target_scale | 12 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 60.000000 / `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0 |
| real/input_before_normalization/fast | 12 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| real/input_before_normalization/standard | 12 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| real/after_normalization/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/after_normalization/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/after_chebyshev_offset/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/after_chebyshev_offset/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/polynomial_input/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/polynomial_input/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/polynomial_output/fast | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| real/polynomial_output/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| real/before_double_angle_round_0/fast | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| real/before_double_angle_round_0/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_multiply/fast | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_multiply/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_add/fast | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_add/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_constant/fast | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_constant/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_rescale/fast | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_0_after_rescale/standard | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_0_fast | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_multiply/fast | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_multiply/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_add/fast | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_add/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_constant/fast | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_constant/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_rescale/fast | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_1_after_rescale/standard | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_1_fast | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0.000000 / `` | 1 |
| real/double_angle_round_2_after_multiply/fast | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_multiply/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_add/fast | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_add/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_constant/fast | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_constant/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_rescale/fast | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_2_after_rescale/standard | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| real/double_angle_round_2_fast | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 2 |
| real/evalmod_output_before_public_scale_reset/fast | 4 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| real/evalmod_output_before_public_scale_reset/standard | 4 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| real/evalmod_output_after_public_scale_reset/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| real/evalmod_output_after_public_scale_reset/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| imag/input_before_normalization/fast | 12 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| imag/input_before_normalization/standard | 12 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| imag/after_normalization/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/after_normalization/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/after_chebyshev_offset/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/after_chebyshev_offset/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/polynomial_input/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/polynomial_input/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/polynomial_output/fast | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| imag/polynomial_output/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| imag/before_double_angle_round_0/fast | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| imag/before_double_angle_round_0/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_multiply/fast | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_multiply/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_add/fast | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_add/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_constant/fast | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_constant/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_rescale/fast | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_after_rescale/standard | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_0_fast | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_multiply/fast | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_multiply/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_add/fast | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_add/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_constant/fast | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_constant/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_rescale/fast | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_after_rescale/standard | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_1_fast | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0.000000 / `` | 1 |
| imag/double_angle_round_2_after_multiply/fast | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_multiply/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_add/fast | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_add/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_constant/fast | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_constant/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_rescale/fast | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_after_rescale/standard | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 0 |
| imag/double_angle_round_2_fast | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0.000000 / `` | 2 |
| imag/evalmod_output_before_public_scale_reset/fast | 4 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| imag/evalmod_output_before_public_scale_reset/standard | 4 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0.000000 / `` | 0 |
| imag/evalmod_output_after_public_scale_reset/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |
| imag/evalmod_output_after_public_scale_reset/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0.000000 / `` | 0 |

## S2C attribution

The S2C amplification is reported separately as a combined-branch metric: the S2C error RMSE divided by the joint RMSE of concatenated EvalMod real and imag semantic errors. It is not an ordinary sequential `A_i`. The S2C ΔSNR parent is the joint EvalMod real+imag reference, not the preceding table row.


## Limitations

- Fast output is decoded directly from c0 under the current Fast zero-secret mode; each Standard output is decrypted with its trial secret key.
- The diagnostic uses the canonical deterministic plaintext-like input c0=encoded message, c1=0; it does not characterize encrypted-input noise or security.
- This is a numerical correctness diagnostic, not a timing benchmark or an independent scientific acceptance decision.

The JSON artifact contains aggregate metrics and worst-slot evidence only; decoded slot arrays are intentionally not persisted.
