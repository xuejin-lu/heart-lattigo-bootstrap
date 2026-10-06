# FAST-STANDARD-PERF-REBASELINE-002 — LogN16

- Classification: `FAST_NUMERICAL_QUALITY_DEGRADED`
- Measurement Primary SHA: `68558a82af6faffd304fa0f14bbc4c82fa190176`
- Fast SHA: `5117fc57949647182f476dc5952099c706b9f869`
- Genuine Standard SHA: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Config: `configs/bootstrap_config.logN16.json` (`16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02`)
- Input: `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf`; metadata SHA `58b699119d1ca2d73cfe8812e6ae776f77d4267ff4c25e67229d7b140a6cf44f`
- Effective LogN / LogSlots / slots: 16 / 15 / 32768
- q0 target / actual prime bits: 55 / 56 (prime `36028797019488257`)
- Threshold: `0.01` per real/imag component
- Environment: `go1.26.4`, `darwin/arm64`, CPU `arm64`

## Genuine Standard / Fast bootstrap SNR

SNR uses the reusable `numericalmetrics.Compare(preDecoded, postDecoded)` implementation. Each Standard trial uses a fresh secret/evaluation key pair and decrypts both pre/post values; Fast uses its current c0-direct path.

| Backend | Trial | Status | SNR dB | Noise RMSE |
|---|---:|---|---:|---:|
| Standard | 1 | FINITE | 130.947666 | 5.80818105e-09 |
| Standard | 2 | FINITE | 130.947666 | 5.80818105e-09 |
| Standard | 3 | FINITE | 130.947666 | 5.80818105e-09 |
| Fast | 1 | FINITE | -0.000461910319 | 0.020485498 |
| Fast | 2 | FINITE | -0.000461910319 | 0.020485498 |

## Full public Bootstrap timing

| Backend | Median ms | Mean ms | Min ms | Max ms | Samples ns | Alloc B/op | Allocs/op |
|---|---:|---:|---:|---:|---|---:|---:|
| Standard | 3827.880 | 3934.338 | 3294.619 | 5074.047 | `[4053977208 3294618916 3413033458 4150301750 3827880083 3726507417 5074047208]` | 849679900 | 40299 |
| Fast | 675.102 | 678.169 | 674.359 | 689.027 | `[675572334 674817917 674716084 683590041 674358500 675102042 689027416]` | 57363056 | 14066 |

Median speedup: `5.6701x`.

Stage timings are separate single-sample public-stage replays and are not summed into the full Bootstrap time.

| Backend | Stage | Elapsed ms | Available | Reason |
|---|---|---:|---|---|
| Standard | pack_and_switch | 0.006 | true |  |
| Standard | scale_down | 0.072 | true |  |
| Standard | mod_up | 123.561 | true |  |
| Standard | coeffs_to_slots | 5390.337 | true |  |
| Standard | evalmod_real | 2169.692 | true |  |
| Standard | evalmod_imag | 1306.977 | true |  |
| Standard | slots_to_coeffs | 1562.317 | true |  |
| Fast | pack_and_switch | 0.038 | true |  |
| Fast | scale_down | 0.058 | true |  |
| Fast | mod_up | 4.137 | true |  |
| Fast | coeffs_to_slots | 104.089 | true |  |
| Fast | evalmod_real | 235.546 | true |  |
| Fast | evalmod_imag | 235.339 | true |  |
| Fast | slots_to_coeffs | 95.889 | true |  |

## Actual pinned-baseline output comparison

- Median Fast-vs-Standard complex RMSE across 6 trial pairs: `0.020485498`
- Maximum complex difference: `0.046281023`
- Max real / imag component difference: `0.0312505497` / `0.0341370377`
- Real / imag components over threshold: 138780 / 30306

| Fast trial | Standard trial | Fast-vs-Standard RMSE | Max complex diff | Median precision bits | Fast-vs-original RMSE |
|---:|---:|---:|---:|---:|---:|
| 1 | 1 | 0.020485498 | 0.046281023 | 5.7622 | 0.020485498 |
| 1 | 2 | 0.020485498 | 0.046281023 | 5.7622 | 0.020485498 |
| 1 | 3 | 0.020485498 | 0.046281023 | 5.7622 | 0.020485498 |
| 2 | 1 | 0.020485498 | 0.046281023 | 5.7622 | 0.020485498 |
| 2 | 2 | 0.020485498 | 0.046281023 | 5.7622 | 0.020485498 |
| 2 | 3 | 0.020485498 | 0.046281023 | 5.7622 | 0.020485498 |

## Topology-aware public stage comparisons

`C2S real` and `C2S imag` are parallel children of ModUp. EvalMod real/imag compare only with their matching C2S branch. S2C amplification uses the joint real+imag EvalMod error representation; final public output compares with S2C.

| Checkpoint | L / scale (Fast; Standard) | Fast-vs-Standard RMSE | Max complex diff | Stage-reference SNR | Parent | Aᵢ | ΔSNRᵢ | Metadata match |
|---|---|---:|---:|---:|---|---:|---:|---|
| input | 0 / 45.0000; 0 / 45.0000 | 0 | 0 | POSITIVE_INFINITY |  | n/a | n/a | true |
| scale_down | 0 / 45.0000; 0 / 45.0000 | 0 | 0 | POSITIVE_INFINITY | input | n/a | n/a | true |
| mod_up | 16 / 50.0000; 16 / 50.0000 | n/a | n/a | NOT_COMPARABLE | scale_down | n/a | n/a | true |

  Not comparable: Fast maintains 4 Q rows while Standard uses 17; decoded values require a common-Q projection
| c2s_real | 12 / 50.0000; 12 / 50.0000 | n/a | n/a | NOT_COMPARABLE | mod_up | n/a | n/a | true |

  Not comparable: Fast maintains 4 Q rows while Standard uses 13; decoded values require a common-Q projection
| c2s_imag | 12 / 50.0000; 12 / 50.0000 | n/a | n/a | NOT_COMPARABLE | mod_up | n/a | n/a | true |

  Not comparable: Fast maintains 4 Q rows while Standard uses 13; decoded values require a common-Q projection
| evalmod_real | 4 / 45.0000; 4 / 45.0000 | n/a | n/a | NOT_COMPARABLE | c2s_real | n/a | n/a | true |

  Not comparable: Fast maintains 4 Q rows while Standard uses 5; decoded values require a common-Q projection
| evalmod_imag | 4 / 45.0000; 4 / 45.0000 | n/a | n/a | NOT_COMPARABLE | c2s_imag | n/a | n/a | true |

  Not comparable: Fast maintains 4 Q rows while Standard uses 5; decoded values require a common-Q projection
| s2c | 1 / 45.0000; 1 / 45.0000 | 0.020485498 | 0.046281023 | -0.000462 dB / FINITE | evalmod_real+evalmod_imag (joint) | n/a | n/a | true |
| final_public_output | 1 / 45.0000; 1 / 45.0000 | 0.020485498 | 0.046281023 | -0.000462 dB / FINITE | s2c | 1 | 0 | true |

## Detailed EvalMod / capacity evidence

Fastdiag classification: `FAST_NUMERICAL_QUALITY_DEGRADED`. Q-prefix capacity gate: strict `2B < S_Q`, 56 checkpoints, 0 failures; first failure `none`. Raw fastdiag SHA-256: `716a3a6addc8bb08649d4d891ab8cd73f052e3c22f4723e72c5cd839de1e3554` (`results/FAST-STANDARD-PERF-REBASELINE-002-LogN16-fastdiag-raw.json`). Compact PS generated-power, polynomial, per-round DoubleAngle, stage, and capacity evidence is included under `fastdiag_stage_lockstep`; the full raw fastdiag artifact is separate.

Both exact pinned worktrees recorded nine public checkpoint states/vectors. Direct decoded-vector comparisons are emitted only when the representation uses the same maintained Q-prefix rows; high-level internal EvalMod/PS/DoubleAngle comparisons use fastdiag's explicit common-Q projection and are retained separately. Its Standard replay evaluator is the Standard API in the pinned Fast source tree, not a second build from the genuine-Standard SHA.

Deep stage replay provenance: Primary `4bba6b5cf403aa7ecd79005b78f974b11f07fafc` (dirty=false); Secondary `5117fc57949647182f476dc5952099c706b9f869` (dirty=false). This is the detailed fastdiag replay source, separate from the matched public timing harness SHA `68558a82af6faffd304fa0f14bbc4c82fa190176`.

First material checkpoint: `evalmod_real`, max complex difference `0.00932384135` (material threshold `0.0036407304`; observable threshold `1e-08`). PS plan: degree 30, base 8, input level 12, scale log2 60.000000, 5 blocks.

### Generated Chebyshev powers (real/imag branches)

RMSE/max difference is from the fastdiag common-Q replay; rows with a level/scale mismatch are explicitly not a strict matched-state comparison.

| Branch | Power | Level Fast/Standard | Scale log2 Fast/Standard | RMSE | Max complex diff | State match |
|---|---:|---:|---:|---:|---:|---|
| real | T1 | 12 / 12 | 60.000000 / 60.000000 | 0 | 0 | matched |
| real | T2 | 11 / 11 | 60.000000 / 60.000000 | 3.06256424e-10 | 1.11354237e-09 | matched |
| real | T3 | 10 / 10 | 60.000000 / 60.000000 | 9.78484187e-09 | 3.56683356e-08 | level/scale mismatch |
| real | T4 | 10 / 10 | 60.000000 / 60.000000 | 1.22494867e-09 | 4.43085035e-09 | matched |
| real | T6 | 9 / 9 | 60.000000 / 60.000000 | 2.75272105e-09 | 1.0008754e-08 | level/scale mismatch |
| real | T8 | 9 / 9 | 60.000000 / 60.000000 | 4.88861559e-09 | 1.77479896e-08 | matched |
| real | T16 | 8 / 8 | 60.000000 / 60.000000 | 1.94011724e-08 | 7.04728877e-08 | matched |
| imag | T1 | 12 / 12 | 60.000000 / 60.000000 | 0 | 0 | matched |
| imag | T2 | 11 / 11 | 60.000000 / 60.000000 | 3.04223095e-10 | 1.11698195e-09 | matched |
| imag | T3 | 10 / 10 | 60.000000 / 60.000000 | 9.72187939e-09 | 3.57782345e-08 | level/scale mismatch |
| imag | T4 | 10 / 10 | 60.000000 / 60.000000 | 1.21633397e-09 | 4.44460146e-09 | matched |
| imag | T6 | 9 / 9 | 60.000000 / 60.000000 | 2.73444304e-09 | 1.00396682e-08 | level/scale mismatch |
| imag | T8 | 9 / 9 | 60.000000 / 60.000000 | 4.85558171e-09 | 1.78028853e-08 | matched |
| imag | T16 | 8 / 8 | 60.000000 / 60.000000 | 1.92708696e-08 | 7.06907567e-08 | matched |

### Polynomial, DoubleAngle and EvalMod internal checkpoints

These rows use the same fastdiag common-Q replay and keep real/imag as separate branches. Dᵢ is Fast-vs-Standard complex RMSE; Aᵢ and ΔSNRᵢ preserve their undefined-status labels.

| Checkpoint | Level Fast/Standard | Scale log2 Fast/Standard | Maintained/authority Q rows | Dᵢ RMSE | Real RMSE | Imag RMSE | Aᵢ | ΔSNRᵢ dB |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| real/polynomial_input | 12 / 12 | 60.000000 / 60.000000 | 4 / 13 | 0 | 0 | 0 | n/a (UNDEFINED_ZERO_OVER_ZERO) | n/a (UNDEFINED_INFINITY_MINUS_INFINITY) |
| real/polynomial_output | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 9.54849068e-09 | 9.54849068e-09 | 2.45598143e-24 | n/a (UNDEFINED_ZERO_PREVIOUS_D) | n/a (NEGATIVE_INFINITY) |
| real/before_double_angle_round_0 | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 9.54849068e-09 | 9.54849068e-09 | 2.45598143e-24 | 1 (FINITE) | 0 (FINITE) |
| real/double_angle_round_0_after_rescale | 6 / 6 | 60.000000 / 60.000000 | 4 / 7 | 2.97711887e-08 | 2.97711887e-08 | 7.6407563e-24 | 1 (FINITE) | 1.59076308e-10 (FINITE) |
| real/double_angle_round_1_after_rescale | 5 / 5 | 60.000000 / 60.000000 | 4 / 6 | 6.94906774e-08 | 6.94906774e-08 | 1.78653881e-23 | 1 (FINITE) | 5.32622835e-11 (FINITE) |
| real/double_angle_round_2_after_rescale | 4 / 4 | 60.000000 / 60.000000 | 4 / 5 | 7.84118323e-08 | 7.84118323e-08 | 2.02409096e-23 | 1 (FINITE) | 2.07866276e-13 (FINITE) |
| real/evalmod_output_after_public_scale_reset | 4 / 4 | 45.000000 / 45.000000 | 4 / 5 | 0.00256939892 | 0.00256939892 | 6.63254127e-19 | 32 (FINITE) | 0 (FINITE) |
| imag/polynomial_input | 12 / 12 | 60.000000 / 60.000000 | 4 / 13 | 0 | 0 | 0 | n/a (UNDEFINED_ZERO_OVER_ZERO) | n/a (UNDEFINED_INFINITY_MINUS_INFINITY) |
| imag/polynomial_output | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 9.48363083e-09 | 9.48363083e-09 | 2.45689728e-24 | n/a (UNDEFINED_ZERO_PREVIOUS_D) | n/a (NEGATIVE_INFINITY) |
| imag/before_double_angle_round_0 | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 9.48363083e-09 | 9.48363083e-09 | 2.45689728e-24 | 1 (FINITE) | 0 (FINITE) |
| imag/double_angle_round_0_after_rescale | 6 / 6 | 60.000000 / 60.000000 | 4 / 7 | 2.95689625e-08 | 2.95689625e-08 | 7.6136882e-24 | 1 (FINITE) | 4.56651605e-10 (FINITE) |
| imag/double_angle_round_1_after_rescale | 5 / 5 | 60.000000 / 60.000000 | 4 / 6 | 6.90186496e-08 | 6.90186496e-08 | 1.78655445e-23 | 1 (FINITE) | -3.15765192e-11 (FINITE) |
| imag/double_angle_round_2_after_rescale | 4 / 4 | 60.000000 / 60.000000 | 4 / 5 | 7.78792064e-08 | 7.78792064e-08 | 2.01416736e-23 | 1 (FINITE) | -2.47820102e-13 (FINITE) |
| imag/evalmod_output_after_public_scale_reset | 4 / 4 | 45.000000 / 45.000000 | 4 / 5 | 0.00255194584 | 0.00255194584 | 6.60002361e-19 | 32 (FINITE) | 0 (FINITE) |

## Interpretation limits

Input is c0=encoded-message,c1=0. Standard outputs are decrypted with fresh Standard keys; Fast outputs use current direct-c0 zero-a semantics. This is a specialized numerical/performance comparator, not a security-equivalent encrypted-input or RLWE-noise experiment.
The detailed internal PS/DoubleAngle replay was performed with Fast's direct diagnostic hooks and the Standard evaluator API compiled from the pinned Fast source tree. Exact pinned worktrees recorded public checkpoint states, but raw decoded values from different maintained Q-prefix widths are not compared directly; those checkpoints are marked NOT_COMPARABLE and the common-Q fastdiag replay is reported separately.
