# Fast vs Standard numerical reference — p93-q55

- Classification: **FAST_STANDARD_NUMERICAL_CLOSE**
- Timestamp: `2026-10-05T21:40:00Z`
- Threshold: `0.01`
- Primary: `12464780dc5670a5b50ad4a9466c48d2c0c9437a` (`main`, dirty=false)
- Secondary: `5117fc57949647182f476dc5952099c706b9f869` (`fast-qprefix`, dirty=false)
- Environment: `go1.26.4`, `darwin/arm64`, CPU `arm64`, NumCPU `10`, GOMAXPROCS `10`

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
| Fast 1 | real | 3.584928e-10 | 4.480629e-10 | 3.031758e-10 | 8.855674e-10 | 1.129725e-09 | 1.648099e-09 | 3491 |
| Fast 1 | imag | 3.976396e-10 | 1.103239e-09 | 3.078608e-10 | 9.107840e-10 | 1.280953e-09 | 5.805750e-08 | 0 |
| Fast 1 | complex | 6.006255e-10 | 1.190755e-09 | 5.386957e-10 | 1.125122e-09 | 1.475781e-09 | 5.805793e-08 | 0 |
| Fast 2 | real | 3.584928e-10 | 4.480629e-10 | 3.031758e-10 | 8.855674e-10 | 1.129725e-09 | 1.648099e-09 | 3491 |
| Fast 2 | imag | 3.976396e-10 | 1.103239e-09 | 3.078608e-10 | 9.107840e-10 | 1.280953e-09 | 5.805750e-08 | 0 |
| Fast 2 | complex | 6.006255e-10 | 1.190755e-09 | 5.386957e-10 | 1.125122e-09 | 1.475781e-09 | 5.805793e-08 | 0 |
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
| Fast 1 | real | 3491 | `-0.0078125+0.00390625i` | `-0.0078125016+0.0039062502i` | `-1.6480985e-09+1.6071746e-10i` |
| Fast 1 | imag | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.2283687e-10-5.8057504e-08i` |
| Fast 1 | complex | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.2283687e-10-5.8057504e-08i` |
| Fast 2 | real | 3491 | `-0.0078125+0.00390625i` | `-0.0078125016+0.0039062502i` | `-1.6480985e-09+1.6071746e-10i` |
| Fast 2 | imag | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.2283687e-10-5.8057504e-08i` |
| Fast 2 | complex | 0 | `-0.03125-0.01171875i` | `-0.03125-0.011718808i` | `-2.2283687e-10-5.8057504e-08i` |
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
| 1 | 1 | real | 8.143229e-11 | 1.031707e-10 | 2.047674e-10 | 2.681273e-10 | 4.165404e-10 | 1853 | MATCH |
| 1 | 1 | imag | 7.981245e-11 | 1.014316e-10 | 2.007375e-10 | 2.732208e-10 | 3.940316e-10 | 937 | MATCH |
| 1 | 1 | complex | 1.267167e-10 | 1.446809e-10 | 2.525531e-10 | 3.154815e-10 | 4.675261e-10 | 937 | MATCH |
| 1 | 2 | real | 8.143229e-11 | 1.031707e-10 | 2.047674e-10 | 2.681273e-10 | 4.165404e-10 | 1853 | MATCH |
| 1 | 2 | imag | 7.981245e-11 | 1.014316e-10 | 2.007375e-10 | 2.732208e-10 | 3.940316e-10 | 937 | MATCH |
| 1 | 2 | complex | 1.267167e-10 | 1.446809e-10 | 2.525531e-10 | 3.154815e-10 | 4.675261e-10 | 937 | MATCH |
| 1 | 3 | real | 8.143229e-11 | 1.031707e-10 | 2.047674e-10 | 2.681273e-10 | 4.165404e-10 | 1853 | MATCH |
| 1 | 3 | imag | 7.981245e-11 | 1.014316e-10 | 2.007375e-10 | 2.732208e-10 | 3.940316e-10 | 937 | MATCH |
| 1 | 3 | complex | 1.267167e-10 | 1.446809e-10 | 2.525531e-10 | 3.154815e-10 | 4.675261e-10 | 937 | MATCH |

### Fast-vs-Standard worst-slot examples

| Fast | Standard | Component | Slot | Original | Fast | Standard | Fast−Standard |
|---:|---:|---|---:|---|---|---|---|
| 1 | 1 | real | 1853 | `-0.03125+0.00390625i` | `-0.031250001+0.0039062502i` | `-0.03125+0.0039062503i` | `-4.1654037e-10-1.2165336e-10i` |
| 1 | 1 | imag | 937 | `-0.0234375-0.005859375i` | `-0.0234375-0.0058593747i` | `-0.0234375-0.005859375i` | `2.5163411e-10+3.9403159e-10i` |
| 1 | 1 | complex | 937 | `-0.0234375-0.005859375i` | `-0.0234375-0.0058593747i` | `-0.0234375-0.005859375i` | `2.5163411e-10+3.9403159e-10i` |
| 1 | 2 | real | 1853 | `-0.03125+0.00390625i` | `-0.031250001+0.0039062502i` | `-0.03125+0.0039062503i` | `-4.1654037e-10-1.2165336e-10i` |
| 1 | 2 | imag | 937 | `-0.0234375-0.005859375i` | `-0.0234375-0.0058593747i` | `-0.0234375-0.005859375i` | `2.5163411e-10+3.9403159e-10i` |
| 1 | 2 | complex | 937 | `-0.0234375-0.005859375i` | `-0.0234375-0.0058593747i` | `-0.0234375-0.005859375i` | `2.5163411e-10+3.9403159e-10i` |
| 1 | 3 | real | 1853 | `-0.03125+0.00390625i` | `-0.031250001+0.0039062502i` | `-0.03125+0.0039062503i` | `-4.1654037e-10-1.2165336e-10i` |
| 1 | 3 | imag | 937 | `-0.0234375-0.005859375i` | `-0.0234375-0.0058593747i` | `-0.0234375-0.005859375i` | `2.5163411e-10+3.9403159e-10i` |
| 1 | 3 | complex | 937 | `-0.0234375-0.005859375i` | `-0.0234375-0.0058593747i` | `-0.0234375-0.005859375i` | `2.5163411e-10+3.9403159e-10i` |

## Precision bits

Precision is computed per slot as `-log2(max(|output-original|, 1e-30))` using complex magnitude error; quantiles use linear interpolation over sorted samples.

| Output | Mean | Median | p05 | Minimum | Minimum slot | Slot error | CKKS helper median L2 |
|---|---:|---:|---:|---:|---:|---:|---:|
| Fast 1 | 30.9412 | 30.7898 | 29.7273 | 24.0379 | 0 | 5.805793e-08 | 31.1191 |
| Fast 2 | 30.9412 | 30.7898 | 29.7273 | 24.0379 | 0 | 5.805793e-08 | 31.1191 |
| Standard 1 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |
| Standard 2 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |
| Standard 3 | 30.9329 | 30.7955 | 29.7209 | 24.0380 | 0 | 5.805496e-08 | 31.1129 |

Median across Fast trials: `30.7898` bits; median across Standard trial medians: `30.7955` bits (drop `0.0056` bits).

## Standard-to-Standard variability

- Per-slot Standard mean was computed across all trials; SHA-256: `0af5cdd54baa7f9eb28d53f5048a8856e1dfe15e3473057cc11626eab0479df0`. Sample means: slot 0=(-3.125000e-02-1.171881e-02i); slot 2048=(-4.749342e-10+3.906249e-03i)
- RMS spread around per-slot means (real / imag / complex): `0.000000e+00` / `0.000000e+00` / `0.000000e+00`
- Maximum spread from per-slot mean: `0.000000e+00` at slot 0 (trial 1)
- Maximum pairwise Standard difference: `0.000000e+00` at slot 0 (trials 1 vs 2)

## 1e-2 threshold audit

| Comparison | Real > threshold | Imag > threshold | Combined fraction |
|---|---:|---:|---:|
| Fast vs original (1 comparisons) | 0 / 4096 (0.000000%) | 0 / 4096 (0.000000%) | 0 / 8192 (0.000000%) |
| Standard vs original (3 comparisons) | 0 / 12288 (0.000000%) | 0 / 12288 (0.000000%) | 0 / 24576 (0.000000%) |
| Fast vs Standard (3 comparisons) | 0 / 12288 (0.000000%) | 0 / 12288 (0.000000%) | 0 / 24576 (0.000000%) |

Median Fast complex RMSE: `1.190755e-09`; median Standard complex RMSE: `1.190394e-09`; median Fast-vs-Standard complex RMSE: `1.446809e-10`; max Fast-vs-Standard complex difference: `4.675261e-10`.

## Classification checks

- Fast vs original has no coordinate above 1e-2: `true`
- Fast vs Standard has no coordinate above 1e-2: `true`
- Median precision degradation is at most 2 bits: `true`
- Fast complex RMSE is within 4x Standard, or Standard RMSE is within the documented near-zero floor: `true`
- Public Bootstrap Level/Scale/dimensions contract is respected: `true`
- Standard near-zero floor: `1.421085e-14`

## Decoded-domain Bootstrap SNR

SNR here is numerical distortion, not RLWE security noise or a noise budget. Each Bootstrap SNR uses that mode's actual decoded pre-Bootstrap vector as the signal reference and its decoded post-Bootstrap output as the observation.

| Mode / trial | Signal power | Signal RMS | Error power | Pre→Post RMSE | Bootstrap SNR (dB / status) | Pre-Bootstrap vs canonical original RMSE |
|---|---:|---:|---:|---:|---:|---:|
| Standard 1 | 4.194882e-04 | 2.048141e-02 | 1.417035e-18 | 1.190393e-09 | 144.713390 / FINITE | 7.449975e-13 |
| Standard 2 | 4.194882e-04 | 2.048141e-02 | 1.417035e-18 | 1.190393e-09 | 144.713390 / FINITE | 7.449975e-13 |
| Standard 3 | 4.194882e-04 | 2.048141e-02 | 1.417035e-18 | 1.190393e-09 | 144.713390 / FINITE | 7.449975e-13 |
| Fast Q-prefix 1 | 4.194882e-04 | 2.048141e-02 | 1.417897e-18 | 1.190755e-09 | 144.710750 / FINITE | 7.449975e-13 |
| Fast Q-prefix 2 | 4.194882e-04 | 2.048141e-02 | 1.417897e-18 | 1.190755e-09 | 144.710750 / FINITE | 7.449975e-13 |

`STANDARD_BOOTSTRAP_SNR_DB=144.713390 / FINITE`
`FAST_BOOTSTRAP_SNR_DB=144.710750 / FINITE`

## Stage reference SNR and divergence

`D_i` is Fast-vs-genuine-Standard complex RMSE. Amplification and SNR deltas follow the explicit stage topology, not table adjacency: C2S real/imag are parallel children of ModUp; each EvalMod branch uses its matching C2S branch; S2C uses a joint EvalMod real+imag reference; final public output may use S2C.

| Checkpoint | Fast state | Standard state | D_i RMSE | A_i parent | A_i | Stage reference SNR (dB / status) | ΔSNR_i parent | ΔSNR_i (dB / status) | Max complex diff |
|---|---|---|---:|---|---:|---:|---|---:|---:|
| input | L0 / 45.000000 / d1 / true / false / 1(+1 maintained) | L0 / 45.000000 / d1 / true / false / 1(+0 maintained) | 0.000000e+00 | — | — (NOT_COMPARABLE) | +Inf / POSITIVE_INFINITY | — | — (NOT_COMPARABLE) | 0.000000e+00 |
| scale_down | L0 / 45.000000 / d1 / true / false / 1(+1 maintained) | L0 / 45.000000 / d1 / true / false / 1(+0 maintained) | 0.000000e+00 | input | — (UNDEFINED_ZERO_OVER_ZERO) | +Inf / POSITIVE_INFINITY | input | — (UNDEFINED_INFINITY_MINUS_INFINITY) | 0.000000e+00 |
| mod_up | L16 / 50.000000 / d1 / true / true / 4(+4 maintained) | L16 / 50.000000 / d1 / true / false / 17(+0 maintained) | 0.000000e+00 | scale_down | — (UNDEFINED_ZERO_OVER_ZERO) | +Inf / POSITIVE_INFINITY | scale_down | — (UNDEFINED_INFINITY_MINUS_INFINITY) | 0.000000e+00 |
| c2s_real | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.396436e-14 | mod_up | — (UNDEFINED_ZERO_PREVIOUS_D) | 175.432910 / FINITE | mod_up | — (NEGATIVE_INFINITY) | 9.185075e-14 |
| c2s_imag | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.409568e-14 | mod_up | — (UNDEFINED_ZERO_PREVIOUS_D) | 175.358744 / FINITE | mod_up | — (NEGATIVE_INFINITY) | 9.768351e-14 |
| evalmod_real | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 5.095135e-11 | c2s_real | 2.126130e+03 (FINITE) | 163.066514 / FINITE | c2s_real | -1.236640e+01 (FINITE) | 1.602983e-10 |
| evalmod_imag | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 5.135617e-11 | c2s_imag | 2.131344e+03 (FINITE) | 162.971074 / FINITE | c2s_imag | -1.238767e+01 (FINITE) | 1.754174e-10 |
| s2c | L1 / 45.000000 / d1 / true / true / 2(+2 maintained) | L1 / 45.000000 / d1 / true / false / 2(+0 maintained) | 1.446809e-10 | — | n/a (combined branches) (NOT_APPLICABLE_COMBINED_BRANCH) | 163.018973 / FINITE | evalmod_real+evalmod_imag (joint) | 2.949497e-04 (FINITE) | 4.675261e-10 |
| final_public_output | L1 / 45.000000 / d1 / true / false / 2(+2 maintained) | L1 / 45.000000 / d1 / true / false / 2(+0 maintained) | 1.446809e-10 | s2c | 1.000000e+00 (FINITE) | 163.018973 / FINITE | s2c | 0.000000e+00 (FINITE) | 4.675261e-10 |

`LARGEST_RAW_AMPLIFICATION_CHECKPOINT=evalmod_imag`
`LARGEST_RAW_AMPLIFICATION_FACTOR=2.131344e+03`
`LARGEST_SNR_DROP_CHECKPOINT=evalmod_imag`
`LARGEST_SNR_DROP_DB=-1.238767e+01`

- First observable: ``, max complex diff `—` (threshold `1.000e-08`)
- First material: ``, max complex diff `—` (threshold `0.00364073039722`)
- Final Fast-vs-Standard RMSE: `1.44680895443e-10`
- Combined-branch EvalMod error RMSE: `5.115416e-11`
- Combined-branch S2C amplification: `2.828331e+00` (status `FINITE`; `D_s2c / RMSE(concat(EvalMod real, EvalMod imag))`)
- Current classification: `CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`

`FIRST_OBSERVABLE_CHECKPOINT=`
`FIRST_OBSERVABLE_MAX_DIFF=—`
`FIRST_MATERIAL_CHECKPOINT=`
`FIRST_MATERIAL_MAX_DIFF=—`
`FINAL_FAST_STANDARD_RMSE=1.44680895443e-10`
`COMBINED_BRANCH_S2C_AMPLIFICATION_FACTOR=2.828331e+00`
`S2C_AMPLIFICATION_FACTOR=2.828331e+00 (combined-branch compatibility alias)`

## EvalMod direct stage-equivalence replay

Each Fast checkpoint is decoded from authoritative Q-prefix rows; Standard is decrypted with the genuine Standard secret key. Comparisons use the same mathematical scale directly, with no representation-specific power-of-two alignment. Replay verification compares the manual stage sequence with each actual public EvalMod output.

| Checkpoint | Fast L / log2(scale) / degree / NTT / Montgomery / rows | Standard L / log2(scale) / degree / NTT / Montgomery / rows | D_i RMSE | A_i | Stage reference SNR (dB / status) | ΔSNR_i (dB / status) | Max complex diff |
|---|---|---|---:|---:|---:|---:|---:|---:|
| real/input_before_normalization | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.396436e-14 | — (NO_PREVIOUS_CHECKPOINT) | 175.432910 / FINITE | — (NO_PREVIOUS_CHECKPOINT) | 9.185075e-14 |
| real/after_normalization | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.340270e-17 | 9.765625e-04 (FINITE) | 175.432910 / FINITE | 0.000000e+00 (FINITE) | 8.969800e-17 |
| real/after_chebyshev_offset | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.367438e-17 | 1.011609e+00 (FINITE) | 296.390827 / FINITE | 1.209579e+02 (FINITE) | 9.194034e-17 |
| real/polynomial_input | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.367438e-17 | 1.000000e+00 (FINITE) | 296.390827 / FINITE | 0.000000e+00 (FINITE) | 9.194034e-17 |
| real/polynomial_output | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.074137e-16 | 8.761104e+00 (FINITE) | 311.499279 / FINITE | 1.510845e+01 (FINITE) | 6.661338e-16 |
| real/before_double_angle_round_0 | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.074137e-16 | 1.000000e+00 (FINITE) | 311.499279 / FINITE | 0.000000e+00 (FINITE) | 6.661338e-16 |
| real/double_angle_round_0_after_multiply | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 3.072959e-16 | 1.481560e+00 (FINITE) | 305.920924 / FINITE | -5.578355e+00 (FINITE) | 9.992007e-16 |
| real/double_angle_round_0_after_add | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 6.145917e-16 | 2.000000e+00 (FINITE) | 305.920924 / FINITE | 0.000000e+00 (FINITE) | 1.998401e-15 |
| real/double_angle_round_0_after_constant | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 5.981904e-16 | 9.733134e-01 (FINITE) | 299.784619 / FINITE | -6.136305e+00 (FINITE) | 1.998401e-15 |
| real/double_angle_round_0_after_rescale | L6 / 60.000000 / d1 / true / true / 4(+4 maintained) | L6 / 60.000000 / d1 / true / false / 7(+0 maintained) | 5.992258e-16 | 1.001731e+00 (FINITE) | 299.769597 / FINITE | -1.502145e-02 (FINITE) | 1.998401e-15 |
| real/double_angle_round_1_after_multiply | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 6.906716e-16 | 1.152607e+00 (FINITE) | 293.857383 / FINITE | -5.912214e+00 (FINITE) | 2.220446e-15 |
| real/double_angle_round_1_after_add | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.381343e-15 | 2.000000e+00 (FINITE) | 293.857383 / FINITE | 0.000000e+00 (FINITE) | 4.440892e-15 |
| real/double_angle_round_1_after_constant | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.379336e-15 | 9.985468e-01 (FINITE) | 286.214500 / FINITE | -7.642883e+00 (FINITE) | 4.329870e-15 |
| real/double_angle_round_1_after_rescale | L5 / 60.000000 / d1 / true / true / 4(+4 maintained) | L5 / 60.000000 / d1 / true / false / 6(+0 maintained) | 1.377895e-15 | 9.989555e-01 (FINITE) | 286.223577 / FINITE | 9.076793e-03 (FINITE) | 4.329870e-15 |
| real/double_angle_round_2_after_multiply | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 7.776965e-16 | 5.644090e-01 (FINITE) | 280.199599 / FINITE | -6.023978e+00 (FINITE) | 2.442491e-15 |
| real/double_angle_round_2_after_add | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 1.555393e-15 | 2.000000e+00 (FINITE) | 280.199599 / FINITE | 0.000000e+00 (FINITE) | 4.884981e-15 |
| real/double_angle_round_2_after_constant | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 1.555227e-15 | 9.998932e-01 (FINITE) | 163.064754 / FINITE | -1.171348e+02 (FINITE) | 4.896305e-15 |
| real/double_angle_round_2_after_rescale | L4 / 60.000000 / d1 / true / true / 4(+4 maintained) | L4 / 60.000000 / d1 / true / false / 5(+0 maintained) | 1.554912e-15 | 9.997974e-01 (FINITE) | 163.066514 / FINITE | 1.759718e-03 (FINITE) | 4.891915e-15 |
| real/evalmod_output_before_public_scale_reset | L4 / 50.000000 / d1 / true / true / 4(+4 maintained) | L4 / 50.000000 / d1 / true / false / 5(+0 maintained) | 1.592230e-12 | 1.024000e+03 (FINITE) | 163.066514 / FINITE | 0.000000e+00 (FINITE) | 5.009321e-12 |
| real/evalmod_output_after_public_scale_reset | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 5.095135e-11 | 3.200000e+01 (FINITE) | 163.066514 / FINITE | 0.000000e+00 (FINITE) | 1.602983e-10 |
| imag/input_before_normalization | L12 / 50.000000 / d1 / true / true / 4(+4 maintained) | L12 / 50.000000 / d1 / true / false / 13(+0 maintained) | 2.409568e-14 | — (NO_PREVIOUS_CHECKPOINT) | 175.358744 / FINITE | — (NO_PREVIOUS_CHECKPOINT) | 9.768351e-14 |
| imag/after_normalization | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.353093e-17 | 9.765625e-04 (FINITE) | 175.358744 / FINITE | 0.000000e+00 (FINITE) | 9.539405e-17 |
| imag/after_chebyshev_offset | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.387932e-17 | 1.014806e+00 (FINITE) | 296.315960 / FINITE | 1.209572e+02 (FINITE) | 9.540979e-17 |
| imag/polynomial_input | L12 / 60.000000 / d1 / true / true / 4(+4 maintained) | L12 / 60.000000 / d1 / true / false / 13(+0 maintained) | 2.387932e-17 | 1.000000e+00 (FINITE) | 296.315960 / FINITE | 0.000000e+00 (FINITE) | 9.540979e-17 |
| imag/polynomial_output | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.141523e-16 | 8.968106e+00 (FINITE) | 311.221574 / FINITE | 1.490561e+01 (FINITE) | 7.771561e-16 |
| imag/before_double_angle_round_0 | L7 / 60.000000 / d1 / true / true / 4(+4 maintained) | L7 / 60.000000 / d1 / true / false / 8(+0 maintained) | 2.141523e-16 | 1.000000e+00 (FINITE) | 311.221574 / FINITE | 0.000000e+00 (FINITE) | 7.771561e-16 |
| imag/double_angle_round_0_after_multiply | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 3.120767e-16 | 1.457265e+00 (FINITE) | 305.786831 / FINITE | -5.434743e+00 (FINITE) | 1.110223e-15 |
| imag/double_angle_round_0_after_add | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 6.241534e-16 | 2.000000e+00 (FINITE) | 305.786831 / FINITE | 0.000000e+00 (FINITE) | 2.220446e-15 |
| imag/double_angle_round_0_after_constant | L7 / 120.000000 / d1 / true / true / 4(+4 maintained) | L7 / 120.000000 / d1 / true / false / 8(+0 maintained) | 6.036090e-16 | 9.670843e-01 (FINITE) | 299.706294 / FINITE | -6.080538e+00 (FINITE) | 2.109424e-15 |
| imag/double_angle_round_0_after_rescale | L6 / 60.000000 / d1 / true / true / 4(+4 maintained) | L6 / 60.000000 / d1 / true / false / 7(+0 maintained) | 6.041472e-16 | 1.000892e+00 (FINITE) | 299.698553 / FINITE | -7.741053e-03 (FINITE) | 2.109424e-15 |
| imag/double_angle_round_1_after_multiply | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 6.967847e-16 | 1.153336e+00 (FINITE) | 293.780843 / FINITE | -5.917709e+00 (FINITE) | 2.331468e-15 |
| imag/double_angle_round_1_after_add | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.393569e-15 | 2.000000e+00 (FINITE) | 293.780843 / FINITE | 0.000000e+00 (FINITE) | 4.662937e-15 |
| imag/double_angle_round_1_after_constant | L6 / 120.000000 / d1 / true / true / 4(+4 maintained) | L6 / 120.000000 / d1 / true / false / 7(+0 maintained) | 1.391299e-15 | 9.983705e-01 (FINITE) | 286.139495 / FINITE | -7.641348e+00 (FINITE) | 4.773959e-15 |
| imag/double_angle_round_1_after_rescale | L5 / 60.000000 / d1 / true / true / 4(+4 maintained) | L5 / 60.000000 / d1 / true / false / 6(+0 maintained) | 1.391663e-15 | 1.000262e+00 (FINITE) | 286.137217 / FINITE | -2.278054e-03 (FINITE) | 4.773959e-15 |
| imag/double_angle_round_2_after_multiply | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 7.838092e-16 | 5.632175e-01 (FINITE) | 280.131596 / FINITE | -6.005621e+00 (FINITE) | 2.664535e-15 |
| imag/double_angle_round_2_after_add | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 1.567618e-15 | 2.000000e+00 (FINITE) | 280.131596 / FINITE | 0.000000e+00 (FINITE) | 5.329071e-15 |
| imag/double_angle_round_2_after_constant | L5 / 120.000000 / d1 / true / true / 4(+4 maintained) | L5 / 120.000000 / d1 / true / false / 6(+0 maintained) | 1.567673e-15 | 1.000035e+00 (FINITE) | 162.968820 / FINITE | -1.171628e+02 (FINITE) | 5.355400e-15 |
| imag/double_angle_round_2_after_rescale | L4 / 60.000000 / d1 / true / true / 4(+4 maintained) | L4 / 60.000000 / d1 / true / false / 5(+0 maintained) | 1.567266e-15 | 9.997405e-01 (FINITE) | 162.971074 / FINITE | 2.254631e-03 (FINITE) | 5.353314e-15 |
| imag/evalmod_output_before_public_scale_reset | L4 / 50.000000 / d1 / true / true / 4(+4 maintained) | L4 / 50.000000 / d1 / true / false / 5(+0 maintained) | 1.604880e-12 | 1.024000e+03 (FINITE) | 162.971074 / FINITE | 0.000000e+00 (FINITE) | 5.481793e-12 |
| imag/evalmod_output_after_public_scale_reset | L4 / 45.000000 / d1 / true / true / 4(+4 maintained) | L4 / 45.000000 / d1 / true / false / 5(+0 maintained) | 5.135617e-11 | 3.200000e+01 (FINITE) | 162.971074 / FINITE | 0.000000e+00 (FINITE) | 1.754174e-10 |

Replay verification RMSE by mode: `imag.fast=0.000000e+00 (verified=true)`; `imag.standard=0.000000e+00 (verified=true)`; `real.fast=0.000000e+00 (verified=true)`; `real.standard=0.000000e+00 (verified=true)`

## Standard-equivalent Fast EvalMod Q-prefix capacity audit

Bounds are exact observed centered coefficient maxima by ciphertext component. Every checkpoint requires strict `2B < S_Q`; execution stops at the first failed guard.

| Checkpoint | Level | Rows | Exact prefix product S_Q | Degree | MaxAbs by component | Strict `2B < S_Q` |
|---|---:|---:|---|---:|---|---:|
| real/evalmod-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[617940038 0]` | true |
| real/polynomial-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014399071389638 0]` | true |
| real/generated-power-1 | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014399071389638 0]` | true |
| real/generated-power-2 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1152358554617929345 0]` | true |
| real/generated-power-3 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[54025605026444462 0]` | true |
| real/generated-power-4 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1150670254407469943 0]` | true |
| real/generated-power-6 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1147858252707227554 0]` | true |
| real/generated-power-8 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1143925295611478874 0]` | true |
| real/generated-power-16 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1117077062881759653 0]` | true |
| real/ps-baby-0 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[145443220899369869427772687336 0]` | true |
| real/ps-baby-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[21627059961569798817453269322886 0]` | true |
| real/ps-baby-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[68144292540391981978043715792120622 0]` | true |
| real/ps-baby-3 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2385407501396353137439299281866268627 0]` | true |
| real/ps-baby-4 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1396742617284346584302592688043329408 0]` | true |
| real/ps-giant-0 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[970051626977544046192490768092224172 0 0]` | true |
| real/ps-giant-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[21772219183013335532733117574682 0 0]` | true |
| real/ps-giant-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[68165894871656975479794746703610630 0 0]` | true |
| real/ps-giant-3 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[1036098237547029464191419262806787988 0 0]` | true |
| real/polynomial-before-final-rescale | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1036098237547029464191419262806787988 0]` | true |
| real/polynomial-after-final-rescale | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671968045640977 0]` | true |
| real/evalmod-polynomial-output | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671968045640977 0]` | true |
| real/double-angle-0-before | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671968045640977 0]` | true |
| real/double-angle-0-after-rescale | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775420668974808 0]` | true |
| real/double-angle-1-before | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775420668974808 0]` | true |
| real/double-angle-1-after-rescale | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233143804554676 0]` | true |
| real/double-angle-2-before | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233143804554676 0]` | true |
| real/double-angle-2-after-rescale | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[9887040613 0]` | true |
| real/evalmod-output | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[9887040613 0]` | true |
| imag/evalmod-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[745777719 0]` | true |
| imag/polynomial-entry | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014398155867732 0]` | true |
| imag/generated-power-1 | 12 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[18014398155867732 0]` | true |
| imag/generated-power-2 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1152358554675149467 0]` | true |
| imag/generated-power-3 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[54025602282560937 0]` | true |
| imag/generated-power-4 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1150670254636238673 0]` | true |
| imag/generated-power-6 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1147858253221538266 0]` | true |
| imag/generated-power-8 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1143925296524766975 0]` | true |
| imag/generated-power-16 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1117077066506406626 0]` | true |
| imag/ps-baby-0 | 11 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[145443220900629764413562617966 0]` | true |
| imag/ps-baby-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[21627059961707339832189485338480 0]` | true |
| imag/ps-baby-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[68144292541778178863066431562476050 0]` | true |
| imag/ps-baby-3 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[2385407501819324605448410428613551551 0]` | true |
| imag/ps-baby-4 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1396742617752725238508475630899385808 0]` | true |
| imag/ps-giant-0 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[970051628818439733967835463786165985 0 0]` | true |
| imag/ps-giant-1 | 10 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[21772219183180886824212972386577 0 0]` | true |
| imag/ps-giant-2 | 9 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[68165894873060586266473885168290017 0 0]` | true |
| imag/ps-giant-3 | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 2 | `[1036098239603590524285184654646812329 0 0]` | true |
| imag/polynomial-before-final-rescale | 8 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[1036098239603590524285184654646812329 0]` | true |
| imag/polynomial-after-final-rescale | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969829423352 0]` | true |
| imag/evalmod-polynomial-output | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969829423352 0]` | true |
| imag/double-angle-0-before | 7 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[898671969829423352 0]` | true |
| imag/double-angle-0-after-rescale | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775426230620454 0]` | true |
| imag/double-angle-1-before | 6 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[672775426230620454 0]` | true |
| imag/double-angle-1-after-rescale | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233156786317838 0]` | true |
| imag/double-angle-2-before | 5 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[325233156786317838 0]` | true |
| imag/double-angle-2-after-rescale | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[11932443579 0]` | true |
| imag/evalmod-output | 4 | 4 | `5986308565615587353347023386369277282933624144412673` | 1 | `[11932443579 0]` | true |

## Polynomial plan and generated-power evidence

- Fast diagnostic PS plan (real branch): degree `30`, base `8`, level `12`, target scale log2 `60.000000`, exact `1.15292150460689408087499999518512994174077078771145643543150072218850255012512207e+18`, blocks `5`.

Each Standard power is generated by the genuine Standard `PowerBasis.GenPower` using the Standard CKKS evaluator; comparisons decrypt Standard ciphertexts and decode Fast through authoritative Q-prefix rows.

| Branch | Power | Fast Level / Scale log2 | Standard Level / Scale log2 | Fast rows | Standard rows | Level/Scale match | Complex RMSE Fast-vs-Standard | Max complex diff | Reference provenance |
|---|---:|---|---|---:|---:|---|---:|---:|---|
| real | T1 | 12 / 60.000000 | 12 / 60.000000 | 4 | 13 | true | 2.367438e-17 | 9.194034e-17 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| real | T2 | 11 / 60.000000 | 11 / 60.000000 | 4 | 12 | true | 2.818592e-17 | 2.220446e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| real | T3 | 10 / 60.000000 | 10 / 60.000000 | 4 | 11 | true | 7.054116e-17 | 2.775558e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| real | T4 | 10 / 60.000000 | 10 / 60.000000 | 4 | 11 | true | 5.979137e-17 | 3.330669e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| real | T6 | 9 / 60.000000 | 9 / 60.000000 | 4 | 10 | true | 5.779520e-17 | 2.220446e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| real | T8 | 9 / 60.000000 | 9 / 60.000000 | 4 | 10 | true | 1.160321e-16 | 3.330669e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| real | T16 | 8 / 60.000000 | 8 / 60.000000 | 4 | 9 | true | 3.459806e-16 | 1.110223e-15 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T1 | 12 / 60.000000 | 12 / 60.000000 | 4 | 13 | true | 2.387932e-17 | 9.540979e-17 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T2 | 11 / 60.000000 | 11 / 60.000000 | 4 | 12 | true | 3.208066e-17 | 2.220446e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T3 | 10 / 60.000000 | 10 / 60.000000 | 4 | 11 | true | 7.123580e-17 | 2.706169e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T4 | 10 / 60.000000 | 10 / 60.000000 | 4 | 11 | true | 6.504635e-17 | 3.330669e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T6 | 9 / 60.000000 | 9 / 60.000000 | 4 | 10 | true | 5.938737e-17 | 3.330669e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T8 | 9 / 60.000000 | 9 / 60.000000 | 4 | 10 | true | 1.243931e-16 | 4.440892e-16 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |
| imag | T16 | 8 / 60.000000 | 8 / 60.000000 | 4 | 9 | true | 3.567378e-16 | 1.110223e-15 | genuine Standard PowerBasis.GenPower using Standard CKKS Evaluator |

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
