# BTS-FAST-NOISE-FIDELITY-009

Status: `ALL_THREE_ARMS_COMPLETE_AND_VALIDATED`

Input SHA-256: `5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373`; LogN16, E32, original BTS Q30/P5 and identical plaintext oracle. Fast OFF and Fast+Noise use commit `be9c42ff23662300769b77f1eead3f12fc53adef` and binary `df4b9c9b0dbccb370d081d239f4d0a986c769bdbd077bf83669f93a978de883e`.

## Three-arm comparison

Times are milliseconds; precision is global `-log2(RMSE)`, with the separate median per-slot precision shown explicitly. Allocation values are mean bytes/count over five warm calls.

| 指標 | Standard | Fast OFF | Fast+Noise |
|---|---:|---:|---:|
| cold cold (ms) | 72073.439 | 7000.037 | 6055.017 |
| warm mean (ms) | 156530.870 | 6666.347 | 8620.939 |
| warm median (ms) | 150167.095 | 5304.018 | 5543.626 |
| warm sample stddev (ms) | 21391.469 | 3253.187 | 6266.852 |
| warm min / max (ms) | 141071.666 / 193556.814 | 4914.079 / 12473.909 | 5219.604 / 19718.835 |
| warm speedup vs Standard | 1.000× | 28.312× | 27.088× |
| warm allocations mean bytes | 3471629563 | 5300960242 | 5609481398 |
| warm allocations mean count | 52561 | 29956111 | 29958592 |
| cold allocation bytes / count | 3706589952 / 53355 | 5330651744 / 29956276 | 5638742840 / 29958749 |
| input-relative RMSE (cold) | 1.632446e-04 | 5.708477e-06 | 3.594987e-05 |
| global RMSE precision bits (cold) | 12.5807 | 17.4185 | 14.7637 |
| median per-slot precision bits (cold) | 13.5613 | 18.0053 | 15.3357 |
| warm global RMSE precision bits (mean across runs) | 12.5807 | 17.4185 | 14.7684 |
| warm median per-slot precision bits (mean across runs) | 13.5613 | 18.0053 | 15.3358 |
| cold MaxAbs / p99 / p99.9 | 1.210e-03 / 5.239e-04 / 7.820e-04 | 2.525e-05 / 1.476e-05 / 1.925e-05 | 1.515e-04 / 9.226e-05 / 1.183e-04 |
| Noise events (all calls) | 0 | 0 | 84 |

Evaluator/BTS construction: Standard 39.813 s, Fast OFF 44.444 s, Fast+Noise 42.434 s. Construction plus Cold: 111.886 s, 51.444 s, 48.489 s.

Fast+Noise overhead vs Fast OFF: cold -945019708 ns (0.8650x); warm median +239607833 ns (1.0452x); warm allocation means +308521157 bytes and +2481 allocations.

Across five warm outputs, mean global RMSE precision is Standard 12.5807 bits, Fast OFF 17.4185 bits, Fast+Noise 14.7684 bits. Noise moved closer to Standard: true (distance 4.8378 → 2.1877 bits).

## Individual precision and execution captures

| Arm | Phase | Time (ms) | Allocated bytes | Allocations | RMSE | Global bits | Slot-median bits | MaxAbs | p99 | p99.9 | Output SHA-256 | Noise events |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---:|
| Standard | first_cold_bootstrap | 72073.439 | 3706589952 | 53355 | 1.632446e-04 | 12.5807 | 13.5613 | 1.210e-03 | 5.239e-04 | 7.820e-04 | `163349917e656fe69ec303f968c708acc5f555a5f531c59792644abab4eec091` | 0 |
| Standard | warm_bootstrap_01 | 143254.461 | 3591279936 | 53025 | 1.632446e-04 | 12.5807 | 13.5613 | 1.210e-03 | 5.239e-04 | 7.820e-04 | `163349917e656fe69ec303f968c708acc5f555a5f531c59792644abab4eec091` | 0 |
| Standard | warm_bootstrap_02 | 141071.666 | 3389417752 | 52243 | 1.632446e-04 | 12.5807 | 13.5613 | 1.210e-03 | 5.239e-04 | 7.820e-04 | `163349917e656fe69ec303f968c708acc5f555a5f531c59792644abab4eec091` | 0 |
| Standard | warm_bootstrap_03 | 193556.814 | 3372110720 | 52177 | 1.632446e-04 | 12.5807 | 13.5613 | 1.210e-03 | 5.239e-04 | 7.820e-04 | `163349917e656fe69ec303f968c708acc5f555a5f531c59792644abab4eec091` | 0 |
| Standard | warm_bootstrap_04 | 154604.316 | 3618560232 | 53169 | 1.632446e-04 | 12.5807 | 13.5613 | 1.210e-03 | 5.239e-04 | 7.820e-04 | `163349917e656fe69ec303f968c708acc5f555a5f531c59792644abab4eec091` | 0 |
| Standard | warm_bootstrap_05 | 150167.095 | 3386779176 | 52192 | 1.632446e-04 | 12.5807 | 13.5613 | 1.210e-03 | 5.239e-04 | 7.820e-04 | `163349917e656fe69ec303f968c708acc5f555a5f531c59792644abab4eec091` | 0 |
| Fast OFF | first_cold_bootstrap | 7000.037 | 5330651744 | 29956276 | 5.708477e-06 | 17.4185 | 18.0053 | 2.525e-05 | 1.476e-05 | 1.925e-05 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` | 0 |
| Fast OFF | warm_bootstrap_01 | 5304.018 | 5301800368 | 29956119 | 5.708477e-06 | 17.4185 | 18.0053 | 2.525e-05 | 1.476e-05 | 1.925e-05 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` | 0 |
| Fast OFF | warm_bootstrap_02 | 4914.079 | 5300747472 | 29956097 | 5.708477e-06 | 17.4185 | 18.0053 | 2.525e-05 | 1.476e-05 | 1.925e-05 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` | 0 |
| Fast OFF | warm_bootstrap_03 | 5482.214 | 5300748968 | 29956103 | 5.708477e-06 | 17.4185 | 18.0053 | 2.525e-05 | 1.476e-05 | 1.925e-05 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` | 0 |
| Fast OFF | warm_bootstrap_04 | 5157.512 | 5300747616 | 29956098 | 5.708477e-06 | 17.4185 | 18.0053 | 2.525e-05 | 1.476e-05 | 1.925e-05 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` | 0 |
| Fast OFF | warm_bootstrap_05 | 12473.909 | 5300756784 | 29956137 | 5.708477e-06 | 17.4185 | 18.0053 | 2.525e-05 | 1.476e-05 | 1.925e-05 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` | 0 |
| Fast+Noise | first_cold_bootstrap | 6055.017 | 5638742840 | 29958749 | 3.594987e-05 | 14.7637 | 15.3357 | 1.515e-04 | 9.226e-05 | 1.183e-04 | `9046bab9e52bb5e071126b4b5a77d69102b7313d9e04dcce5eec3ef941d09b53` | 14 |
| Fast+Noise | warm_bootstrap_01 | 5219.604 | 5609895256 | 29958599 | 3.583658e-05 | 14.7682 | 15.3483 | 1.545e-04 | 9.240e-05 | 1.176e-04 | `598e5a081d4612f0ab23e25cc16698191ab9a4b9636c5678f9e2f8667a662d3c` | 14 |
| Fast+Noise | warm_bootstrap_02 | 5257.090 | 5608858888 | 29958580 | 3.585559e-05 | 14.7674 | 15.3363 | 1.522e-04 | 9.335e-05 | 1.195e-04 | `291ca21ad6809cd8bc8541c4d4d8f24816f28699448f9d6b7bf76b80c1f54c6c` | 14 |
| Fast+Noise | warm_bootstrap_03 | 5543.626 | 5608826120 | 29958579 | 3.579111e-05 | 14.7700 | 15.3358 | 1.571e-04 | 9.175e-05 | 1.188e-04 | `8c9540ef7769986a4104a3afb624bd6de7d5d49be4761389887e756b0aadb608` | 14 |
| Fast+Noise | warm_bootstrap_04 | 7365.540 | 5610474712 | 29958618 | 3.578751e-05 | 14.7702 | 15.3358 | 1.508e-04 | 9.172e-05 | 1.171e-04 | `98780412c0d4d090aca56b58b2a7bc8dd60237189d9aba3f45b1e70200ac2ee0` | 14 |
| Fast+Noise | warm_bootstrap_05 | 19718.835 | 5609352016 | 29958585 | 3.589105e-05 | 14.7660 | 15.3229 | 1.527e-04 | 9.178e-05 | 1.197e-04 | `4495209aa7cbc62bf2f9d5a76fe8023e29349af00b652f85a01dcc623794030f` | 14 |

BTS-008 historical warm median (separate older capture): Standard 154536.049 ms, Fast 5044.745 ms, 30.633x. This is context only and is not part of the 009 comparison.
