# Fast vs Standard numerical reference — p93-q55

- Classification: **FAST_NUMERICAL_QUALITY_DEGRADED**
- Timestamp: `2026-10-05T10:29:36Z`
- Threshold: `0.01`
- Primary: `4fd11a5aa557a9e91c0473c94f76b9bfc41a9e36` (`main`, dirty=false)
- Secondary: `97c1c6174e0d7ef781de6d8dadce5c2869a53496` (`fast-qprefix`, dirty=false)
- Environment: `go1.26.4`, `darwin/arm64`, CPU `Apple M4`, NumCPU `10`, GOMAXPROCS `10`

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
| Fast 1 | real | 1.688692e-02 | 1.949464e-02 | 1.673427e-02 | 3.184576e-02 | 3.324075e-02 | 3.582357e-02 | 3398 |
| Fast 1 | imag | 6.510326e-03 | 7.512877e-03 | 6.462624e-03 | 1.244487e-02 | 1.369237e-02 | 1.637619e-02 | 2032 |
| Fast 1 | complex | 1.897382e-02 | 2.089221e-02 | 1.866507e-02 | 3.276182e-02 | 3.440657e-02 | 3.640730e-02 | 2583 |
| Fast 2 | real | 1.688692e-02 | 1.949464e-02 | 1.673427e-02 | 3.184576e-02 | 3.324075e-02 | 3.582357e-02 | 3398 |
| Fast 2 | imag | 6.510326e-03 | 7.512877e-03 | 6.462624e-03 | 1.244487e-02 | 1.369237e-02 | 1.637619e-02 | 2032 |
| Fast 2 | complex | 1.897382e-02 | 2.089221e-02 | 1.866507e-02 | 3.276182e-02 | 3.440657e-02 | 3.640730e-02 | 2583 |
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
| Fast 1 | real | 3398 | `0.02734375-0.0078125i` | `-0.0084798177-0.010239219i` | `-0.035823568-0.0024267193i` |
| Fast 1 | imag | 2032 | `0.00390625+0.01171875i` | `-0.00029943571-0.004657441i` | `-0.0042056857-0.016376191i` |
| Fast 1 | complex | 2583 | `0.03125-0.009765625i` | `-0.0036998661+0.00043234843i` | `-0.034949866+0.010197973i` |
| Fast 2 | real | 3398 | `0.02734375-0.0078125i` | `-0.0084798177-0.010239219i` | `-0.035823568-0.0024267193i` |
| Fast 2 | imag | 2032 | `0.00390625+0.01171875i` | `-0.00029943571-0.004657441i` | `-0.0042056857-0.016376191i` |
| Fast 2 | complex | 2583 | `0.03125-0.009765625i` | `-0.0036998661+0.00043234843i` | `-0.034949866+0.010197973i` |
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
| 1 | 1 | real | 1.688692e-02 | 1.949464e-02 | 3.184576e-02 | 3.324075e-02 | 3.582357e-02 | 3398 | MATCH |
| 1 | 1 | imag | 6.510326e-03 | 7.512877e-03 | 1.244487e-02 | 1.369237e-02 | 1.637619e-02 | 2032 | MATCH |
| 1 | 1 | complex | 1.897382e-02 | 2.089221e-02 | 3.276182e-02 | 3.440656e-02 | 3.640730e-02 | 2583 | MATCH |
| 1 | 2 | real | 1.688692e-02 | 1.949464e-02 | 3.184576e-02 | 3.324075e-02 | 3.582357e-02 | 3398 | MATCH |
| 1 | 2 | imag | 6.510326e-03 | 7.512877e-03 | 1.244487e-02 | 1.369237e-02 | 1.637619e-02 | 2032 | MATCH |
| 1 | 2 | complex | 1.897382e-02 | 2.089221e-02 | 3.276182e-02 | 3.440656e-02 | 3.640730e-02 | 2583 | MATCH |
| 1 | 3 | real | 1.688692e-02 | 1.949464e-02 | 3.184576e-02 | 3.324075e-02 | 3.582357e-02 | 3398 | MATCH |
| 1 | 3 | imag | 6.510326e-03 | 7.512877e-03 | 1.244487e-02 | 1.369237e-02 | 1.637619e-02 | 2032 | MATCH |
| 1 | 3 | complex | 1.897382e-02 | 2.089221e-02 | 3.276182e-02 | 3.440656e-02 | 3.640730e-02 | 2583 | MATCH |

### Fast-vs-Standard worst-slot examples

| Fast | Standard | Component | Slot | Original | Fast | Standard | Fast−Standard |
|---:|---:|---|---:|---|---|---|---|
| 1 | 1 | real | 3398 | `0.02734375-0.0078125i` | `-0.0084798177-0.010239219i` | `0.027343749-0.0078125065i` | `-0.035823567-0.0024267129i` |
| 1 | 1 | imag | 2032 | `0.00390625+0.01171875i` | `-0.00029943571-0.004657441i` | `0.0039062499+0.011718749i` | `-0.0042056856-0.01637619i` |
| 1 | 1 | complex | 2583 | `0.03125-0.009765625i` | `-0.0036998661+0.00043234843i` | `0.03125-0.0097656251i` | `-0.034949866+0.010197974i` |
| 1 | 2 | real | 3398 | `0.02734375-0.0078125i` | `-0.0084798177-0.010239219i` | `0.027343749-0.0078125065i` | `-0.035823567-0.0024267129i` |
| 1 | 2 | imag | 2032 | `0.00390625+0.01171875i` | `-0.00029943571-0.004657441i` | `0.0039062499+0.011718749i` | `-0.0042056856-0.01637619i` |
| 1 | 2 | complex | 2583 | `0.03125-0.009765625i` | `-0.0036998661+0.00043234843i` | `0.03125-0.0097656251i` | `-0.034949866+0.010197974i` |
| 1 | 3 | real | 3398 | `0.02734375-0.0078125i` | `-0.0084798177-0.010239219i` | `0.027343749-0.0078125065i` | `-0.035823567-0.0024267129i` |
| 1 | 3 | imag | 2032 | `0.00390625+0.01171875i` | `-0.00029943571-0.004657441i` | `0.0039062499+0.011718749i` | `-0.0042056856-0.01637619i` |
| 1 | 3 | complex | 2583 | `0.03125-0.009765625i` | `-0.0036998661+0.00043234843i` | `0.03125-0.0097656251i` | `-0.034949866+0.010197974i` |

## Precision bits

Precision is computed per slot as `-log2(max(|output-original|, 1e-30))` using complex magnitude error; quantiles use linear interpolation over sorted samples.

| Output | Mean | Median | p05 | Minimum | Minimum slot | Slot error | CKKS helper median L2 |
|---|---:|---:|---:|---:|---:|---:|---:|
| Fast 1 | 5.9277 | 5.7435 | 4.9318 | 4.7796 | 2583 | 3.640730e-02 | 5.4011 |
| Fast 2 | 5.9277 | 5.7435 | 4.9318 | 4.7796 | 2583 | 3.640730e-02 | 5.4011 |
| Standard 1 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |
| Standard 2 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |
| Standard 3 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |

Median across Fast trials: `5.7435` bits; median across Standard trial medians: `30.7955` bits (drop `25.0519` bits).

## Standard-to-Standard variability

- Per-slot Standard mean was computed across all trials; SHA-256: `0af5cdd54baa7f9eb28d53f5048a8856e1dfe15e3473057cc11626eab0479df0`. Sample means: slot 0=(-3.125000e-02-1.171881e-02i); slot 2048=(-4.749342e-10+3.906249e-03i)
- RMS spread around per-slot means (real / imag / complex): `0.000000e+00` / `0.000000e+00` / `0.000000e+00`
- Maximum spread from per-slot mean: `0.000000e+00` at slot 0 (trial 1)
- Maximum pairwise Standard difference: `0.000000e+00` at slot 0 (trials 1 vs 2)

## 1e-2 threshold audit

| Comparison | Real > threshold | Imag > threshold | Combined fraction |
|---|---:|---:|---:|
| Fast vs original (1 comparisons) | 2872 / 4096 (70.117188%) | 903 / 4096 (22.045898%) | 3775 / 8192 (46.081543%) |
| Standard vs original (3 comparisons) | 0 / 12288 (0.000000%) | 0 / 12288 (0.000000%) | 0 / 24576 (0.000000%) |
| Fast vs Standard (3 comparisons) | 8616 / 12288 (70.117188%) | 2709 / 12288 (22.045898%) | 11325 / 24576 (46.081543%) |

Median Fast complex RMSE: `2.089221e-02`; median Standard complex RMSE: `1.190394e-09`; median Fast-vs-Standard complex RMSE: `2.089221e-02`; max Fast-vs-Standard complex difference: `3.640730e-02`.

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
| Fast Q-prefix 1 | 4.194882e-04 | 2.048141e-02 | 4.364845e-04 | 2.089221e-02 | -0.172491 / FINITE | 7.449975e-13 |
| Fast Q-prefix 2 | 4.194882e-04 | 2.048141e-02 | 4.364845e-04 | 2.089221e-02 | -0.172491 / FINITE | 7.449975e-13 |

`STANDARD_BOOTSTRAP_SNR_DB=144.713390 / FINITE`
`FAST_BOOTSTRAP_SNR_DB=-0.172491 / FINITE`

## Stage reference SNR and divergence

`D_i` is Fast-vs-genuine-Standard complex RMSE. Amplification and SNR deltas follow the explicit stage topology, not table adjacency: C2S real/imag are parallel children of ModUp; each EvalMod branch uses its matching C2S branch; S2C uses a joint EvalMod real+imag reference; final public output may use S2C.

| Checkpoint | Fast state | Standard state | D_i RMSE | A_i parent | A_i | Stage reference SNR (dB / status) | ΔSNR_i parent | ΔSNR_i (dB / status) | Max complex diff |
|---|---|---|---:|---|---:|---:|---|---:|---:|
| input | L0 / 45.000000 / d1 / true / false / 1(+1 maintained) | L0 / 45.000000 / d1 / true / false / 1(+0 maintained) | 0.000000e+00 | — | — (NOT_COMPARABLE) | +Inf / POSITIVE_INFINITY | — | — (NOT_COMPARABLE) | 0.000000e+00 |
| scale_down | L0 / 45.000000 / d1 / true / false / 1(+1 maintained) | L0 / 45.000000 / d1 / true / false / 1(+0 maintained) | 0.000000e+00 | input | — (UNDEFINED_ZERO_OVER_ZERO) | +Inf / POSITIVE_INFINITY | input | — (UNDEFINED_INFINITY_MINUS_INFINITY) | 0.000000e+00 |
| mod_up | L16 / 50.000000 / d1 / true / true / 4(+4 maintained) | L16 / 50.000000 / d1 / true / false / 17(+0 maintained) | 0.000000e+00 | scale_down | — (UNDEFINED_ZERO_OVER_ZERO) | +Inf / POSITIVE_INFINITY | scale_down | — (UNDEFINED_INFINITY_MINUS_INFINITY) | 0.000000e+00 |
| c2s_real | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.396436e-14 | mod_up | — (UNDEFINED_ZERO_PREVIOUS_D) | 175.432910 / FINITE | mod_up | — (NEGATIVE_INFINITY) | 9.185075e-14 |
| c2s_imag | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.409568e-14 | mod_up | — (UNDEFINED_ZERO_PREVIOUS_D) | 175.358744 / FINITE | mod_up | — (NEGATIVE_INFINITY) | 9.768351e-14 |
| evalmod_real | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.422415e-03 | c2s_real | 3.097272e+11 (FINITE) | -0.201277 / FINITE | c2s_real | -1.756342e+02 (FINITE) | 2.033136e-02 |
| evalmod_imag | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.350434e-03 | c2s_imag | 3.050520e+11 (FINITE) | -0.143334 / FINITE | c2s_imag | -1.755021e+02 (FINITE) | 2.259150e-02 |
| s2c | L1 / 45.000000 / d1 / true / true / 2(+2 maintained) | L1 / 45.000000 / d1 / true / false / 2(+0 maintained) | 2.089221e-02 | — | n/a (combined branches) (NOT_APPLICABLE_COMBINED_BRANCH) | -0.172491 / FINITE | evalmod_real+evalmod_imag (joint) | 9.742762e-13 (FINITE) | 3.640730e-02 |
| final_public_output | L1 / 45.000000 / d1 / true / false / 2(+2 maintained) | L1 / 45.000000 / d1 / true / false / 2(+0 maintained) | 2.089221e-02 | s2c | 1.000000e+00 (FINITE) | -0.172491 / FINITE | s2c | 0.000000e+00 (FINITE) | 3.640730e-02 |

`LARGEST_RAW_AMPLIFICATION_CHECKPOINT=evalmod_real`
`LARGEST_RAW_AMPLIFICATION_FACTOR=3.097272e+11`
`LARGEST_SNR_DROP_CHECKPOINT=evalmod_real`
`LARGEST_SNR_DROP_DB=-1.756342e+02`

- First observable: `evalmod_real`, max complex diff `2.033136e-02` (threshold `1.000e-08`)
- First material: `evalmod_real`, max complex diff `2.033136e-02` (threshold `0.00364073039722`)
- Final Fast-vs-Standard RMSE: `0.0208922110696`
- Combined-branch EvalMod error RMSE: `7.386512e-03`
- Combined-branch S2C amplification: `2.828427e+00` (status `FINITE`; `D_s2c / RMSE(concat(EvalMod real, EvalMod imag))`)
- Current classification: `CURRENT_FAST_FIRST_MATERIAL_EVALMOD_NORMALIZATION`

`FIRST_OBSERVABLE_CHECKPOINT=evalmod_real`
`FIRST_OBSERVABLE_MAX_DIFF=2.033136e-02`
`FIRST_MATERIAL_CHECKPOINT=evalmod_real`
`FIRST_MATERIAL_MAX_DIFF=2.033136e-02`
`FINAL_FAST_STANDARD_RMSE=0.0208922110696`
`COMBINED_BRANCH_S2C_AMPLIFICATION_FACTOR=2.828427e+00`
`S2C_AMPLIFICATION_FACTOR=2.828427e+00 (combined-branch compatibility alias)`

## EvalMod internal bisect

Standard is decrypted with the genuine Standard secret key; Fast is decoded through the validated Q-prefix c0 path after projecting both branches to common authoritative Q rows. Replay verification compares each source-faithful replay output with the actual public EvalMod output.

| Checkpoint | Fast L / log2(scale) / degree / NTT / Montgomery / rows | Standard L / log2(scale) / degree / NTT / Montgomery / rows | D_i RMSE | A_i | Stage reference SNR (dB / status) | ΔSNR_i (dB / status) | Max complex diff |
|---|---|---|---:|---:|---:|---:|---:|---:|
| real/input_before_normalization | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.396436e-14 | — (NO_PREVIOUS_CHECKPOINT) | 175.432910 / FINITE | — (NO_PREVIOUS_CHECKPOINT) | 9.185075e-14 |
| real/after_normalization | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.340270e-17 | 9.765625e-04 (FINITE) | 175.432910 / FINITE | 0.000000e+00 (FINITE) | 8.969800e-17 |
| real/after_chebyshev_offset | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.367438e-17 | 1.011609e+00 (FINITE) | 296.390827 / FINITE | 1.209579e+02 (FINITE) | 9.194034e-17 |
| real/polynomial_input | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.367438e-17 | 1.000000e+00 (FINITE) | 296.390827 / FINITE | 0.000000e+00 (FINITE) | 9.194034e-17 |
| real/generated_power_polynomial_output | L7 / 31.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.760408e-08 | 1.165989e+09 (FINITE) | 149.016565 / FINITE | -1.473743e+02 (FINITE) | 7.565266e-08 |
| real/before_coherent_scale_transition | L7 / 31.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.760408e-08 | 1.000000e+00 (FINITE) | 149.016565 / FINITE | 0.000000e+00 (FINITE) | 7.565266e-08 |
| real/after_coherent_scale_transition | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 7.794737e-01 | 2.823763e+07 (FINITE) | 0.000000 / FINITE | -1.490166e+02 (FINITE) | 7.794738e-01 |
| real/before_double_angle_round_0 | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 7.794737e-01 | 1.000000e+00 (FINITE) | 0.000000 / FINITE | 0.000000e+00 (FINITE) | 7.794738e-01 |
| real/double_angle_round_0_after_multiply | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 6.075792e-01 | 7.794737e-01 (FINITE) | 0.000000 / FINITE | -1.617873e-08 (FINITE) | 6.075793e-01 |
| real/double_angle_round_0_after_constant | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 5.835397e-01 | 9.604339e-01 (FINITE) | 0.000000 / FINITE | 1.617873e-08 (FINITE) | 5.835399e-01 |
| real/double_angle_round_0_after_rescale | L6 / 60.000000 / d1 / true / true / 4(+4 maintained) | L6 / 60.000000 / d1 / true / false / 7(+0 maintained) | 5.835397e-01 | 1.000000e+00 (FINITE) | 0.000000 / FINITE | 0.000000e+00 (FINITE) | 5.835399e-01 |
| real/double_angle_round_1_after_multiply | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 3.405185e-01 | 5.835397e-01 (FINITE) | 0.000000 / FINITE | -1.617873e-08 (FINITE) | 3.405188e-01 |
| real/double_angle_round_1_after_constant | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 2.820948e-01 | 8.284271e-01 (FINITE) | 0.000000 / FINITE | 1.617871e-08 (FINITE) | 2.820953e-01 |
| real/double_angle_round_1_after_rescale | L5 / 60.000000 / d1 / true / true / 4(+4 maintained) | L5 / 60.000000 / d1 / true / false / 6(+0 maintained) | 2.820948e-01 | 1.000000e+00 (FINITE) | 0.000000 / FINITE | 9.643275e-16 (FINITE) | 2.820953e-01 |
| real/double_angle_round_2_after_multiply | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 7.957747e-02 | 2.820948e-01 (FINITE) | 0.000000 / FINITE | -1.617871e-08 (FINITE) | 7.957777e-02 |
| real/double_angle_round_2_after_constant | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 2.213255e-07 | 2.781258e-06 (FINITE) | -0.000000 / FINITE | -3.069136e-10 (FINITE) | 6.274740e-07 |
| real/double_angle_round_2_after_rescale | L4 / 60.000000 / d1 / true / true / 4(+4 maintained) | L4 / 60.000000 / d1 / true / false / 5(+0 maintained) | 2.213255e-07 | 1.000000e+00 (FINITE) | -0.000000 / FINITE | -3.433006e-12 (FINITE) | 6.274740e-07 |
| real/evalmod_output_before_public_scale_reset | L4 / 50.000000 / d1 / true / true / 4(+4 maintained) | L4 / 50.000000 / d1 / true / false / 5(+0 maintained) | 2.319505e-04 | 1.048006e+03 (FINITE) | -0.201277 / FINITE | -2.012768e-01 (FINITE) | 6.353551e-04 |
| real/evalmod_output_after_public_scale_reset | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 7.422415e-03 | 3.200000e+01 (FINITE) | -0.201277 / FINITE | 0.000000e+00 (FINITE) | 2.033136e-02 |

Replay verification RMSE by mode: `real.fast=0.000000e+00 (verified=true)`; `real.standard=0.000000e+00 (verified=true)`

## Polynomial plan and generated-power evidence

- Fast diagnostic PS plan: degree `30`, base `8`, level `12`, target scale log2 `60.000000`, exact `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18`, blocks `5`.

| Power n | Level | Scale log2 | Fast maintained rows | RMSE vs reference | Max complex diff | Reference provenance |
|---:|---:|---:|---:|---:|---:|---|
| 1 | 12 | 60.000000 | 4 | 2.367438e-17 | 9.194034e-17 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| 2 | 11 | 60.000000 | 4 | 8.605597e-10 | 2.753504e-09 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| 3 | 10 | 60.000000 | 4 | 2.749948e-08 | 8.800716e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| 4 | 10 | 60.000000 | 4 | 3.456590e-09 | 9.801409e-09 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| 6 | 9 | 60.000000 | 4 | 7.733163e-09 | 2.484335e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| 8 | 9 | 60.000000 | 4 | 1.379921e-08 | 3.912483e-08 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |
| 16 | 8 | 60.000000 | 4 | 5.476583e-08 | 1.552693e-07 | plaintext Chebyshev recurrence from genuine Standard polynomial-input decode; not Standard ciphertext power |

## EvalMod exact-scale audit

The q0=55 path retains plan scale 2^91; these values are observed, not altered by the diagnostic. Exact scale strings are retained even when float64 projection is non-finite.

| Checkpoint | Level | Scale log2 | Exact scale | Plan bits | Working bits | Target log2 / exact | kIn | multiplier | round |
|---|---:|---:|---|---:|---:|---|---:|---:|---:|
| evalmod_real/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| evalmod_real/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| evalmod_imag/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| evalmod_imag/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| evalmod/fixed_q0_55_normalized_plan | 12 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 91 | 31 | 60.000000 / `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0 | 0 | 0 |
| real/input_before_normalization/fast | 12 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/input_before_normalization/standard | 12 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/after_normalization/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/after_normalization/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/after_chebyshev_offset/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/after_chebyshev_offset/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/polynomial_input/fast | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/polynomial_input/standard | 12 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/generated_power_polynomial_output/fast | 7 | 31.000000 | `2.14748364800030517391860489413660047419373055647882651071796378250411407861975022e+09` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/generated_power_polynomial_output/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/before_coherent_scale_transition/fast | 7 | 31.000000 | `2.14748364800030517391860489413660047419373055647882651071796378250411407861975022e+09` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/before_coherent_scale_transition/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/normalized_coherent_scale_transition | 7 | 60.000000 | `1.15292150460701081500000002328278014916002058853905509749893099069595336914062500e+18` | 91 | 31 | 60.000000 / `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 29 | 30 | 0 |
| real/after_coherent_scale_transition/fast | 7 | 60.000000 | `1.15292150460701081500000002328278014916002058853905509749893099069595336914062500e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/after_coherent_scale_transition/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/before_double_angle_round_0/fast | 7 | 60.000000 | `1.15292150460701081500000002328278014916002058853905509749893099069595336914062500e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/before_double_angle_round_0/standard | 7 | 60.000000 | `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_after_multiply/fast | 7 | 120.000000 | `1.32922799578529365991659370321340006700781250000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_after_multiply/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_after_constant/fast | 7 | 120.000000 | `1.32922799578529365991659370321340006700781250000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_after_constant/standard | 7 | 120.000000 | `1.32922799578502448935052568629877657665625000000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_after_rescale/fast | 6 | 60.000000 | `1.15292150460702719700000004679833410202449694992310469388030469417572021484375000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_after_rescale/standard | 6 | 60.000000 | `1.15292150460679372874999999910506875969357955258320913571878918446600437164306641e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_0_fast | 6 | 60.000000 | `1.15292150460702719700000004679833410202449694992310469388030469417572021484375000e+18` | 91 | 31 | 0.000000 / `` | 29 | 30 | 0 |
| real/double_angle_round_1_after_multiply/fast | 6 | 120.000000 | `1.32922799578533143423677070180728834101562500000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_1_after_multiply/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_1_after_constant/fast | 6 | 120.000000 | `1.32922799578533143423677070180728834101562500000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_1_after_constant/standard | 6 | 120.000000 | `1.32922799578479309310463470718675763237500000000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_1_after_rescale/fast | 5 | 60.000000 | `1.15292150460730572100000016088279155768094204337348429589837905950844287872314453e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_1_after_rescale/standard | 5 | 60.000000 | `1.15292150460683878449999999997089972214852487963909766222059261053800582885742188e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_1_fast | 5 | 60.000000 | `1.15292150460730572100000016088279155768094204337348429589837905950844287872314453e+18` | 91 | 31 | 0.000000 / `` | 29 | 30 | 1 |
| real/double_angle_round_2_after_multiply/fast | 5 | 120.000000 | `1.32922799578597366685506937772979005721093750000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_2_after_multiply/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_2_after_constant/fast | 5 | 120.000000 | `1.32922799578597366685506937772979005721093750000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_2_after_constant/standard | 5 | 120.000000 | `1.32922799578489698459079708630633676800000000000000000000000000000000000000000000e+36` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_2_after_rescale/fast | 4 | 60.000000 | `1.15292150460778084900000051756956064732506476999684963402614812366664409637451172e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_2_after_rescale/standard | 4 | 60.000000 | `1.15292150460684697600000000000000000000000000000000000000000000000000000000000000e+18` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/double_angle_round_2_fast | 4 | 60.000000 | `1.15292150460778084900000051756956064732506476999684963402614812366664409637451172e+18` | 91 | 31 | 0.000000 / `` | 29 | 30 | 2 |
| real/evalmod_output_before_public_scale_reset/fast | 4 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/evalmod_output_before_public_scale_reset/standard | 4 | 50.000000 | `1.12589990684262400000000000000000000000000000000000000000000000000000000000000000e+15` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/evalmod_output_after_public_scale_reset/fast | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |
| real/evalmod_output_after_public_scale_reset/standard | 4 | 45.000000 | `3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13` | 0 | 0 | 0.000000 / `` | 0 | 0 | 0 |

## S2C attribution

The first material divergence is already present at EvalMod, so S2C is not the originating stage. The S2C amplification is reported separately as a combined-branch metric: the S2C error RMSE divided by the joint RMSE of concatenated EvalMod real and imag semantic errors. It is not an ordinary sequential `A_i`. The S2C ΔSNR parent is the joint EvalMod real+imag reference, not the preceding table row.


## Historical-reference reconciliation and next bounded counterfactual

`HISTORICAL_FAST_CKKS_REFERENCE_ONLY`: older q0=56/dirty-tree diagnosis is not used to classify this current q0=55 run. Current evidence is reproduced on the synchronized current Fast Q-prefix branch; the selected plan scale remains 2^91.

Next causal experiment (exactly one; not executed): **test-only coherent-scale exponent correction at the first diverging normalization transition, holding every other stage fixed.**

## Limitations

- Fast output is decoded directly from c0 under the current Fast zero-secret mode; each Standard output is decrypted with its trial secret key.
- The diagnostic uses the canonical deterministic plaintext-like input c0=encoded message, c1=0; it does not characterize encrypted-input noise or security.
- This is a numerical correctness diagnostic, not a timing benchmark or an independent scientific acceptance decision.

The JSON artifact contains aggregate metrics and worst-slot evidence only; decoded slot arrays are intentionally not persisted.
