# QPREFIX-PERF-MEASURE-LOGN16-001 — Standard vs Q-prefix Fast

Classification: `LOGN16_STANDARD_FAST_TIMING_READY`

## Provenance and protocol

- Measurement harness Primary commit: `8186f50e7b591b7f76b39fb89b47c32ad1cc1410` (used identically for both runs).
- Current control-plane Primary commit when measurements began: `716f9afeb478b64762b7ecb78b1bda0bb3740693`.
- Standard Secondary: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Q-prefix Fast Secondary: `82601ea2517edc14784c9da250426169a1b221c7`.
- Host: Apple M4, macOS `darwin/arm64`, 10 logical CPUs; Go `go1.26.4`.
- Config: `configs/bootstrap_config.logN16.json`; warmup 1, measured repetitions 7, `-stages`.
- Both runs used default Go build tags. Each backend passed `go test ./...` before timing.
- Standard measurement: `go run . -stages -config configs/bootstrap_config.logN16.json -warmup 1 -repetitions 7 -out "$BASE_STD/standard-logN16.json"`.
- Fast measurement: `go run . -stages -config configs/bootstrap_config.logN16.json -warmup 1 -repetitions 7 -out "$BASE_FAST/fast-qprefix-logN16.json"`.
- Measurement timestamps (UTC): Standard `2026-09-29T06:04:06.111341Z`; Fast `2026-09-29T06:06:31.741621Z`.

## Matched-run validation

The two artifacts have the same harness commit, host, config, effective parameters, warmup and repetition count. Both runs contain all nine required stages, seven present measurements per stage, and 7/7 staged-versus-full output metadata checks passed (level, scale, and domain).

- Config LogN 16, default scale (2^{45}), full slots (`log_slots=-1`), Mod1 degree 30, DoubleAngle 3, K 16, LogMessageRatio 10.
- Effective residual parameters: LogN 16, MaxLevel 1.
- Effective bootstrap parameters: LogN 16, MaxLevel 16, LogSlots 15.
- Effective Q-chain bit lengths: `[56, 39, 39, 39, 39, 60, 61, 60, 61, 61, 60, 60, 61, 57, 56, 56, 57]`.
- Effective P-chain bit lengths: `[61, 61, 62, 61, 62]`.
- Circuit order: `ModUpThenEncode`.
- Backend path was verified at the frozen Secondary commits: Standard `NewEvaluator` constructs the regular CKKS evaluator; Fast-compatible evaluation keys select the Fast path in `NewEvaluator`, which delegates to `FastEvaluator`.

## Full Bootstrap

- Standard raw times, rep 1→7 (ms): 4814.973708, 4788.739917, 4922.946084, 4728.043334, 4712.632500, 4818.683833, 4846.002000.
- Fast raw times, rep 1→7 (ms): 723.213333, 721.726167, 710.232208, 714.831750, 713.131791, 734.959208, 735.294333.
- Standard median: **4814.973708 ms**.
- Fast median: **721.726167 ms**.
- Speedup (Standard median / Fast median): **6.671×**.
- Median time saved: **4093.247541 ms**.

## Per-stage comparison

Speedup is Standard median / Fast median. Saved time is Standard median − Fast median; negative values mean the Fast stage was slower.

| Stage | Standard median (ms) | Fast median (ms) | Speedup | Saved (ms) |
|---|---:|---:|---:|---:|
| `pack_and_switch_n1_to_n2` | 0.015417 | 0.156958 | 0.098× | -0.141541 |
| `scale_down` | 0.088000 | 0.074084 | 1.188× | 0.013916 |
| `mod_up` | 29.870208 | 4.552292 | 6.562× | 25.317916 |
| `coeffs_to_slots` | 2693.195083 | 110.042167 | 24.474× | 2583.152916 |
| `eval_mod_real` | 684.611750 | 249.075084 | 2.749× | 435.536666 |
| `eval_mod_imag` | 677.138500 | 246.628209 | 2.746× | 430.510291 |
| `slots_to_coeffs` | 770.363916 | 107.628834 | 7.158× | 662.735082 |
| `unpack_and_switch_n2_to_n1` | 0.002875 | 0.393875 | 0.007× | -0.391000 |
| `full_bootstrap` | 4814.973708 | 721.726167 | 6.671× | 4093.247541 |

### Standard per-stage samples and distribution

| Stage | 7 raw runs (ms, rep 1→7) | Median | Mean | Min | Max |
|---|---:|---:|---:|---:|---:|
| `pack_and_switch_n1_to_n2` | 0.014417, 0.016959, 0.015417, 0.015208, 0.008792, 0.015958, 0.027458 | 0.015417 | 0.016316 | 0.008792 | 0.027458 |
| `scale_down` | 0.138208, 0.083458, 0.093000, 0.081708, 0.088000, 0.079958, 1.310541 | 0.088000 | 0.267839 | 0.079958 | 1.310541 |
| `mod_up` | 24.278833, 24.775083, 35.809208, 43.895750, 34.132750, 22.036208, 29.870208 | 29.870208 | 30.685434 | 22.036208 | 43.895750 |
| `coeffs_to_slots` | 2827.077208, 2705.826709, 2693.195083, 2565.538750, 2582.448042, 2865.889541, 2636.184000 | 2693.195083 | 2696.594190 | 2565.538750 | 2865.889541 |
| `eval_mod_real` | 684.611750, 679.645000, 938.864083, 654.342542, 633.312000, 698.626583, 692.329209 | 684.611750 | 711.675881 | 633.312000 | 938.864083 |
| `eval_mod_imag` | 615.521125, 611.850833, 639.359375, 677.997292, 705.232917, 705.266958, 677.138500 | 677.138500 | 661.766714 | 611.850833 | 705.266958 |
| `slots_to_coeffs` | 723.942459, 717.704625, 750.632584, 791.398416, 770.363916, 790.317167, 793.928458 | 770.363916 | 762.612518 | 717.704625 | 793.928458 |
| `unpack_and_switch_n2_to_n1` | 0.006375, 0.005458, 0.000959, 0.001167, 0.006167, 0.002875, 0.001208 | 0.002875 | 0.003458 | 0.000959 | 0.006375 |
| `full_bootstrap` | 4814.973708, 4788.739917, 4922.946084, 4728.043334, 4712.632500, 4818.683833, 4846.002000 | 4814.973708 | 4804.574482 | 4712.632500 | 4922.946084 |

### Fast per-stage samples and distribution

| Stage | 7 raw runs (ms, rep 1→7) | Median | Mean | Min | Max |
|---|---:|---:|---:|---:|---:|
| `pack_and_switch_n1_to_n2` | 0.167041, 0.786250, 0.149500, 0.150375, 0.150458, 0.156958, 4.237459 | 0.156958 | 0.828292 | 0.149500 | 4.237459 |
| `scale_down` | 0.075459, 0.081500, 0.068666, 0.074084, 0.069959, 0.073417, 0.162167 | 0.074084 | 0.086465 | 0.068666 | 0.162167 |
| `mod_up` | 5.864667, 6.428833, 4.505458, 4.552292, 4.490666, 4.548792, 5.393416 | 4.552292 | 5.112018 | 4.490666 | 6.428833 |
| `coeffs_to_slots` | 120.918709, 109.240167, 107.965208, 114.093084, 106.429667, 110.042167, 116.408333 | 110.042167 | 112.156762 | 106.429667 | 120.918709 |
| `eval_mod_real` | 249.279250, 256.492292, 246.286750, 246.926667, 246.078000, 255.788542, 249.075084 | 249.075084 | 249.989512 | 246.078000 | 256.492292 |
| `eval_mod_imag` | 246.088000, 257.699209, 246.016958, 245.814417, 246.628209, 253.178292, 259.057583 | 246.628209 | 250.640381 | 245.814417 | 259.057583 |
| `slots_to_coeffs` | 106.392417, 107.852875, 105.274291, 111.412458, 116.793083, 107.628834, 106.138833 | 107.628834 | 108.784684 | 105.274291 | 116.793083 |
| `unpack_and_switch_n2_to_n1` | 0.370459, 0.393875, 1.162292, 0.384209, 0.386750, 0.399458, 0.419292 | 0.393875 | 0.502334 | 0.370459 | 1.162292 |
| `full_bootstrap` | 723.213333, 721.726167, 710.232208, 714.831750, 713.131791, 734.959208, 735.294333 | 721.726167 | 721.912684 | 710.232208 | 735.294333 |

## Full-bootstrap allocations

| Backend | Median bytes/op | Median allocs/op |
|---|---:|---:|
| Standard | 848929168 | 40282 |
| Q-prefix Fast | 57362536 | 14064 |

Standard/Fast allocation ratios: **14.799×** bytes and **2.864×** allocations.

## EvalMod aggregate

Per-repetition total is `eval_mod_real + eval_mod_imag`; the median is computed from these paired totals, not by summing independent stage medians.

- Standard per-repetition totals: 1300.132875, 1291.495833, 1578.223458, 1332.339834, 1338.544917, 1403.893541, 1369.467709 ms
- Standard EvalMod total median: **1338.544917 ms**.
- Fast per-repetition totals: 495.367250, 514.191501, 492.303708, 492.741084, 492.706209, 508.966834, 508.132667 ms
- Fast EvalMod total median: **495.367250 ms**.
- EvalMod total speedup: **2.702×**; median saved time **843.177667 ms**.

## Limitations

This is a timing comparison, not a LogN16 numerical-quality validation. The accepted LogN13 evidence classifies Fast numerical quality as degraded; successful timing and metadata checks do not establish LogN16 correctness. Very short pack/unpack/scale-down stage timings are sensitive to timer and scheduling noise; interpret those ratios cautiously. No performance threshold was specified.
