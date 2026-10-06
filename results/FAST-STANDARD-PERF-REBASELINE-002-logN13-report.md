# FAST-STANDARD-PERF-REBASELINE-002 — LogN13

- Classification: `FAST_STANDARD_NUMERICAL_CLOSE`
- Measurement Primary SHA: `68558a82af6faffd304fa0f14bbc4c82fa190176`
- Fast SHA: `5117fc57949647182f476dc5952099c706b9f869`
- Genuine Standard SHA: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Config: `configs/bootstrap_config.logN13.json` (`919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`)
- Input: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; metadata SHA `9171dac4a79b97f7f916001d0ea5805f6137711527d364f8dfcd8e71ed2764f6`
- Effective LogN / LogSlots / slots: 13 / 12 / 4096
- q0 target / actual prime bits: 55 / 55 (prime `36028797018652673`)
- Threshold: `0.01` per real/imag component
- Environment: `go1.26.4`, `darwin/arm64`, CPU `arm64`

## Genuine Standard / Fast bootstrap SNR

SNR uses the reusable `numericalmetrics.Compare(preDecoded, postDecoded)` implementation. Each Standard trial uses a fresh secret/evaluation key pair and decrypts both pre/post values; Fast uses its current c0-direct path.

| Backend | Trial | Status | SNR dB | Noise RMSE |
|---|---:|---|---:|---:|
| Standard | 1 | FINITE | 144.71339 | 1.19039296e-09 |
| Standard | 2 | FINITE | 144.71339 | 1.19039296e-09 |
| Standard | 3 | FINITE | 144.71339 | 1.19039296e-09 |
| Fast | 1 | FINITE | 144.71075 | 1.19075486e-09 |
| Fast | 2 | FINITE | 144.71075 | 1.19075486e-09 |

## Full public Bootstrap timing

| Backend | Median ms | Mean ms | Min ms | Max ms | Samples ns | Alloc B/op | Allocs/op |
|---|---:|---:|---:|---:|---|---:|---:|
| Standard | 296.456 | 299.297 | 290.058 | 318.509 | `[295063250 293706500 290058417 318509125 303084208 296455791 298202250]` | 91864720 | 34638 |
| Fast | 67.754 | 67.882 | 67.630 | 68.647 | `[67877250 67903250 68646875 67754125 67706292 67656042 67630042]` | 7499173 | 7903 |

Median speedup: `4.3755x`.

Stage timings are separate single-sample public-stage replays and are not summed into the full Bootstrap time.

| Backend | Stage | Elapsed ms | Available | Reason |
|---|---|---:|---|---|
| Standard | pack_and_switch | 0.001 | true |  |
| Standard | scale_down | 0.015 | true |  |
| Standard | mod_up | 2.352 | true |  |
| Standard | coeffs_to_slots | 140.336 | true |  |
| Standard | evalmod_real | 65.097 | true |  |
| Standard | evalmod_imag | 63.844 | true |  |
| Standard | slots_to_coeffs | 33.759 | true |  |
| Fast | pack_and_switch | 0.006 | true |  |
| Fast | scale_down | 0.011 | true |  |
| Fast | mod_up | 0.434 | true |  |
| Fast | coeffs_to_slots | 10.157 | true |  |
| Fast | evalmod_real | 24.178 | true |  |
| Fast | evalmod_imag | 24.129 | true |  |
| Fast | slots_to_coeffs | 8.821 | true |  |

## Actual pinned-baseline output comparison

- Median Fast-vs-Standard complex RMSE across 6 trial pairs: `1.44680895e-10`
- Maximum complex difference: `4.6752606e-10`
- Max real / imag component difference: `4.16540372e-10` / `3.94031588e-10`
- Real / imag components over threshold: 0 / 0

| Fast trial | Standard trial | Fast-vs-Standard RMSE | Max complex diff | Median precision bits | Fast-vs-original RMSE |
|---:|---:|---:|---:|---:|---:|
| 1 | 1 | 1.44680895e-10 | 4.6752606e-10 | 32.9732 | 1.19075468e-09 |
| 1 | 2 | 1.44680895e-10 | 4.6752606e-10 | 32.9732 | 1.19075468e-09 |
| 1 | 3 | 1.44680895e-10 | 4.6752606e-10 | 32.9732 | 1.19075468e-09 |
| 2 | 1 | 1.44680895e-10 | 4.6752606e-10 | 32.9732 | 1.19075468e-09 |
| 2 | 2 | 1.44680895e-10 | 4.6752606e-10 | 32.9732 | 1.19075468e-09 |
| 2 | 3 | 1.44680895e-10 | 4.6752606e-10 | 32.9732 | 1.19075468e-09 |

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
| s2c | 1 / 45.0000; 1 / 45.0000 | 1.44680895e-10 | 4.6752606e-10 | 163.018973 dB / FINITE | evalmod_real+evalmod_imag (joint) | n/a | n/a | true |
| final_public_output | 1 / 45.0000; 1 / 45.0000 | 1.44680895e-10 | 4.6752606e-10 | 163.018973 dB / FINITE | s2c | 1 | 0 | true |

## Detailed EvalMod / capacity evidence

Fastdiag classification: `FAST_STANDARD_NUMERICAL_CLOSE`. Q-prefix capacity gate: strict `2B < S_Q`, 56 checkpoints, 0 failures; first failure `none`. Raw fastdiag SHA-256: `356c168a8d5dbbe480a2361ccb287f2ee38867ff773cdd033a9960998d1e15dd` (`results/FAST-STANDARD-PERF-REBASELINE-002-LogN13-fastdiag-raw.json`). Compact PS generated-power, polynomial, per-round DoubleAngle, stage, and capacity evidence is included under `fastdiag_stage_lockstep`; the full raw fastdiag artifact is separate.

Both exact pinned worktrees recorded nine public checkpoint states/vectors. Direct decoded-vector comparisons are emitted only when the representation uses the same maintained Q-prefix rows; high-level internal EvalMod/PS/DoubleAngle comparisons use fastdiag's explicit common-Q projection and are retained separately. Its Standard replay evaluator is the Standard API in the pinned Fast source tree, not a second build from the genuine-Standard SHA.

Deep stage replay provenance: Primary `4bba6b5cf403aa7ecd79005b78f974b11f07fafc` (dirty=false); Secondary `5117fc57949647182f476dc5952099c706b9f869` (dirty=false). This is the detailed fastdiag replay source, separate from the matched public timing harness SHA `68558a82af6faffd304fa0f14bbc4c82fa190176`.

First material checkpoint: `none`, max complex difference `n/a` (material threshold `0.0036407304`; observable threshold `1e-08`). PS plan: degree 30, base 8, input level 12, scale log2 60.000000, 5 blocks.

### Generated Chebyshev powers (real/imag branches)

RMSE/max difference is from the fastdiag common-Q replay; rows with a level/scale mismatch are explicitly not a strict matched-state comparison.

| Branch | Power | Level Fast/Standard | Scale log2 Fast/Standard | RMSE | Max complex diff | State match |
|---|---:|---:|---:|---:|---:|---|
| real | T1 | 12 / 12 | 60.000000 / 60.000000 | 2.36743836e-17 | 9.19403442e-17 | matched |
| real | T2 | 11 / 11 | 60.000000 / 60.000000 | 2.81859203e-17 | 2.22044605e-16 | matched |
| real | T3 | 10 / 10 | 60.000000 / 60.000000 | 7.05411629e-17 | 2.77555756e-16 | matched |
| real | T4 | 10 / 10 | 60.000000 / 60.000000 | 5.97913661e-17 | 3.33066907e-16 | matched |
| real | T6 | 9 / 9 | 60.000000 / 60.000000 | 5.77951966e-17 | 2.22044605e-16 | matched |
| real | T8 | 9 / 9 | 60.000000 / 60.000000 | 1.16032125e-16 | 3.33066907e-16 | matched |
| real | T16 | 8 / 8 | 60.000000 / 60.000000 | 3.45980584e-16 | 1.11022302e-15 | matched |
| imag | T1 | 12 / 12 | 60.000000 / 60.000000 | 2.38793221e-17 | 9.54097912e-17 | matched |
| imag | T2 | 11 / 11 | 60.000000 / 60.000000 | 3.20806611e-17 | 2.22044605e-16 | matched |
| imag | T3 | 10 / 10 | 60.000000 / 60.000000 | 7.12357957e-17 | 2.70616862e-16 | matched |
| imag | T4 | 10 / 10 | 60.000000 / 60.000000 | 6.50463477e-17 | 3.33066907e-16 | matched |
| imag | T6 | 9 / 9 | 60.000000 / 60.000000 | 5.93873657e-17 | 3.33066907e-16 | matched |
| imag | T8 | 9 / 9 | 60.000000 / 60.000000 | 1.243931e-16 | 4.4408921e-16 | matched |
| imag | T16 | 8 / 8 | 60.000000 / 60.000000 | 3.56737774e-16 | 1.11022302e-15 | matched |

### Polynomial, DoubleAngle and EvalMod internal checkpoints

These rows use the same fastdiag common-Q replay and keep real/imag as separate branches. Dᵢ is Fast-vs-Standard complex RMSE; Aᵢ and ΔSNRᵢ preserve their undefined-status labels.

| Checkpoint | Level Fast/Standard | Scale log2 Fast/Standard | Maintained/authority Q rows | Dᵢ RMSE | Real RMSE | Imag RMSE | Aᵢ | ΔSNRᵢ dB |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| real/polynomial_input | 12 / 12 | 60.000000 / 60.000000 | 4 / 13 | 2.36743836e-17 | 2.36743836e-17 | 2.53276373e-24 | 1 (FINITE) | 0 (FINITE) |
| real/polynomial_output | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 2.07413741e-16 | 2.07413741e-16 | 4.57539721e-24 | 8.76110417 (FINITE) | 15.1084514 (FINITE) |
| real/before_double_angle_round_0 | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 2.07413741e-16 | 2.07413741e-16 | 4.57539721e-24 | 1 (FINITE) | 0 (FINITE) |
| real/double_angle_round_0_after_rescale | 6 / 6 | 60.000000 / 60.000000 | 4 / 7 | 5.99225819e-16 | 5.99225819e-16 | 1.4146657e-23 | 1.0017309 (FINITE) | -0.0150214501 (FINITE) |
| real/double_angle_round_1_after_rescale | 5 / 5 | 60.000000 / 60.000000 | 4 / 6 | 1.37789524e-15 | 1.37789524e-15 | 3.22793821e-23 | 0.998955541 (FINITE) | 0.00907679315 (FINITE) |
| real/double_angle_round_2_after_rescale | 4 / 4 | 60.000000 / 60.000000 | 4 / 5 | 1.55491182e-15 | 1.55491182e-15 | 3.73888958e-23 | 0.999797425 (FINITE) | 0.00175971804 (FINITE) |
| real/evalmod_output_after_public_scale_reset | 4 / 4 | 45.000000 / 45.000000 | 4 / 5 | 5.09513504e-11 | 5.09513504e-11 | 1.22515934e-18 | 32 (FINITE) | 0 (FINITE) |
| imag/polynomial_input | 12 / 12 | 60.000000 / 60.000000 | 4 / 13 | 2.38793221e-17 | 2.38793221e-17 | 2.5328232e-24 | 1 (FINITE) | 0 (FINITE) |
| imag/polynomial_output | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 2.14152298e-16 | 2.14152298e-16 | 4.50467795e-24 | 8.96810627 (FINITE) | 14.9056139 (FINITE) |
| imag/before_double_angle_round_0 | 7 / 7 | 60.000000 / 60.000000 | 4 / 8 | 2.14152298e-16 | 2.14152298e-16 | 4.50467795e-24 | 1 (FINITE) | 0 (FINITE) |
| imag/double_angle_round_0_after_rescale | 6 / 6 | 60.000000 / 60.000000 | 4 / 7 | 6.04147182e-16 | 6.04147182e-16 | 1.43063348e-23 | 1.00089162 (FINITE) | -0.00774105337 (FINITE) |
| imag/double_angle_round_1_after_rescale | 5 / 5 | 60.000000 / 60.000000 | 4 / 6 | 1.39166346e-15 | 1.39166346e-15 | 3.34270641e-23 | 1.00026231 (FINITE) | -0.00227805418 (FINITE) |
| imag/double_angle_round_2_after_rescale | 4 / 4 | 60.000000 / 60.000000 | 4 / 5 | 1.56726578e-15 | 1.56726578e-15 | 3.71650195e-23 | 0.99974046 (FINITE) | 0.00225463075 (FINITE) |
| imag/evalmod_output_after_public_scale_reset | 4 / 4 | 45.000000 / 45.000000 | 4 / 5 | 5.13561651e-11 | 5.13561651e-11 | 1.21782336e-18 | 32 (FINITE) | 0 (FINITE) |

## Interpretation limits

Input is c0=encoded-message,c1=0. Standard outputs are decrypted with fresh Standard keys; Fast outputs use current direct-c0 zero-a semantics. This is a specialized numerical/performance comparator, not a security-equivalent encrypted-input or RLWE-noise experiment.
The detailed internal PS/DoubleAngle replay was performed with Fast's direct diagnostic hooks and the Standard evaluator API compiled from the pinned Fast source tree. Exact pinned worktrees recorded public checkpoint states, but raw decoded values from different maintained Q-prefix widths are not compared directly; those checkpoints are marked NOT_COMPARABLE and the common-Q fastdiag replay is reported separately.
