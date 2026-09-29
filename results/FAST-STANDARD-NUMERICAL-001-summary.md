# Fast vs Standard numerical reference — p93-q55

- Classification: **FAST_NUMERICAL_QUALITY_DEGRADED**
- Timestamp: `2026-09-29T05:28:34Z`
- Threshold: `0.01`
- Primary: `7541d0517c949593c0f53fb2d1e229791d70c2ef` (`main`, dirty=true)
- Secondary: `bde527336d25ab7e01692706dc0d185a51d4abe2` (`fast-qprefix`, dirty=false)
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

## Limitations

- Fast output is decoded directly from c0 under the current Fast zero-secret mode; each Standard output is decrypted with its trial secret key.
- The diagnostic uses the canonical deterministic plaintext-like input c0=encoded message, c1=0; it does not characterize encrypted-input noise or security.
- This is a numerical correctness diagnostic, not a timing benchmark or an independent scientific acceptance decision.

The JSON artifact contains aggregate metrics and worst-slot evidence only; decoded slot arrays are intentionally not persisted.
