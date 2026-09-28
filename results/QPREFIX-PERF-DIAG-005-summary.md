# QPREFIX-PERF-DIAG-005 — Low-Overhead Power Attribution

**Decision:** `LOW_OVERHEAD_RESCALE_DOMINANCE_CONFIRMED`

**Required metrics**

- `POWER_ONLY_OVERHEAD=0.037253` (3.73%)
- `POWER_ONLY_RESCALE_SHARE=0.937425` (93.74%)
- `GENERATED_POWER_SHARE_OF_EVALMOD=0.605260` (60.53%)

## Provenance and controlled workload

- Primary: `e0398fc330e63bf2342ab81aa20673c0260e1a0c`, `main`, clean during measurement.
- Secondary measured code: `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`.
- Secondary branch at measurement: `fast-qprefix`, `297e872a45aaa46d878fb47859ddb3457cbc0ee0`, clean. Relative to the specified production ref, this control-plane-only commit changes only `CURRENT_TASK.md`; production code is code-equivalent. No Secondary production source was modified.
- Profile: `p93-q55`, LogN=13, LogSlots=12, batch=1, P93, polynomial degree 30, DoubleAngle=3.
- Input SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` (4096 values; identical across traces).
- Environment: Go `go1.26.4`, `darwin/arm64`, 10 CPUs, `GOMAXPROCS=10`.
- Both trace commands used 1 warmup and 7 measured repetitions. `power` and `stage,power` were the only trace scopes used; deep `rescale` scope was disabled throughout.

## Diagnostics-off E2E baseline

Command:

```sh
go test ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkFastDiagP93Q55Count1$' -benchtime=1x -benchmem -count=7
```

The benchmark performs one untimed Bootstrap warmup. Seven measured samples:

| Sample | ms/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 1 | 79.130333 | 7,615,224 | 15,951 |
| 2 | 78.322792 | 7,615,224 | 15,951 |
| 3 | 78.465250 | 7,615,224 | 15,951 |
| 4 | 78.046833 | 7,615,224 | 15,951 |
| 5 | 78.442250 | 7,618,456 | 15,961 |
| 6 | 83.101958 | 7,615,224 | 15,951 |
| 7 | 78.512625 | 7,615,224 | 15,951 |

Median: **78.465250 ms/op**. Allocations are effectively unchanged; one sample had 3,232 additional B/op and 10 additional allocs/op. This diagnostics-off median is the production-speed reference for both overhead calculations.

## Power-only trace

Command:

```sh
./scripts/fastdiag trace --profile p93-q55 --trace power --warmup 1 --repetitions 7 --out /tmp/qpdiag005-power.json
```

Bootstrap wall samples (ms): `81.388291, 81.129167, 81.255833, 81.368541, 81.518042, 81.537375, 81.779208`.

Harness median: **81.388291 ms**; overhead versus diagnostics-off is `(81.388291 / 78.465250) - 1 = 3.7253%`. This is below the 10% category-ranking gate.

The two `generated_powers` parent spans had medians of **18.457750 ms (real call)** and **18.365583 ms (imaginary call)**. The median of their per-run combined duration was **36.826541 ms**. Per-run parent closure uses the two parent spans and direct category events only; nested per-power parent spans are not added again.

Direct category totals are medians of per-run Bootstrap totals across both EvalMod calls:

| Direct category | Median (ms / Bootstrap) |
|---|---:|
| Copy / workspace | 0.232585 |
| Integer scaling | 0.624836 |
| Mul (standalone event) | 0 — no standalone event emitted |
| Relinearize (standalone event) | 0 — not emitted separately |
| MulRelin | 0.703415 |
| Rescale | 34.525831 |
| Chebyshev doubling | 0.192000 |
| Recurrence correction | 0.195210 |
| Residual / unclassified | 0.355247 |

The residual is calculated independently in each run as generated-power parent time minus the listed direct events. It ranged from **0.343214 to 0.370411 ms**. Reconstructed parent closure had a maximum absolute accounting delta of **0 ns** across all seven runs. Separate medians need not add exactly because each category is summarized independently.

Per-generated-power results (medians of per-run totals across real and imaginary calls):

| Power | Inclusive power span (ms) | Copy / workspace (ms) | Integer scaling (ms) | MulRelin (ms) | Rescale (ms) | Chebyshev doubling (ms) | Recurrence correction (ms) |
|---:|---:|---:|---:|---:|---:|---:|---:|
| 2 | 6.472 | 0.043 | 0.104 | 0.117 | 6.097 | 0.032 | 0.017 |
| 3 | 6.113 | 0.040 | 0.104 | 0.117 | 5.649 | 0.032 | 0.113 |
| 4 | 12.040 | 0.037 | 0.104 | 0.116 | 5.215 | 0.032 | 0.016 |
| 6 | 13.144 | 0.037 | 0.104 | 0.117 | 6.657 | 0.032 | 0.017 |
| 8 | 17.585 | 0.037 | 0.104 | 0.116 | 5.135 | 0.032 | 0.016 |
| 16 | 23.656 | 0.037 | 0.104 | 0.117 | 5.721 | 0.032 | 0.017 |

The inclusive power spans are hierarchical: larger powers include nested lower-power calls and must not be summed across rows. No standalone `Mul` or `Relinearize` event is emitted; the framework reports the combined `MulRelin` event.

`POWER_ONLY_RESCALE_SHARE` is the median of per-run `power/rescale` totals divided by per-run `generated_powers` parent totals: **93.7425%** (per-run range **93.6923%–93.7767%**).

## Stage + power trace

Command:

```sh
./scripts/fastdiag trace --profile p93-q55 --trace stage,power --warmup 1 --repetitions 7 --out /tmp/qpdiag005-stage-power.json
```

Bootstrap wall samples (ms): `80.589917, 80.425834, 80.644625, 80.632542, 80.551334, 80.486917, 80.528416`.

| Measurement | Median (ms) |
|---|---:|
| Harness Bootstrap wall | 80.551334 |
| Stage Bootstrap parent | 80.542417 |
| EvalMod real | 30.184209 |
| EvalMod imaginary | 29.987792 |
| Generated-power parent in real call | 18.241541 |
| Generated-power parent in imaginary call | 18.179500 |
| Combined generated-power parent (median of per-run totals) | 36.407375 |
| Power-category Rescale (median of per-run totals) | 34.116000 |

Harness-wall overhead versus diagnostics-off is **2.6586%**. The median of per-run `generated_powers / (evalmod_real + evalmod_imag)` ratios is **60.5260%**, so generated powers remain a major contributor inside the current EvalMod stages. The per-run Rescale share of generated-power time is **93.7000%** (range **93.6435%–93.7363%**).

For each run, the sum of direct power-category events plus the explicit residual closes exactly to the two generated-power parent spans; the residual median was **0.347544 ms** (range **0.337037–0.367880 ms**). The trace's stage Bootstrap parent median closes with the harness wall within the observed sub-millisecond instrumentation/measurement difference. All seven runs passed the reusable framework's numerical comparison against the diagnostics-off replay and remained within the existing `1e-2` decoded-slot tolerance.

## Decision and DIAG-004 reinterpretation

`LOW_OVERHEAD_RESCALE_DOMINANCE_CONFIRMED` meets all specified conditions:

1. Power-only overhead is 3.73%, below 10%.
2. Power-category Rescale is 93.74% of generated-power time, well above 50%.
3. Generated powers account for 60.53% of combined EvalMod time in the stage+power trace.

DIAG-004's `CURRENT_RESCALE_STILL_DOMINANT` classification is **confirmed** by this low-overhead measurement. Its all-scope `96.51%` figure is retained only as `DEEP_TRACE_PERTURBED_REFERENCE_ONLY`; the current power-only estimate is lower, but still shows Rescale dominance. No DIAG-004 historical artifact was modified.

## Next investigation target (not an implementation recommendation)

Investigate the remaining production-side generated-power Rescale cost with focused diagnostics-off measurements that separate preflight from materialization without per-coefficient timers. No production optimization, parameter change, or arithmetic change was made or is proposed by this diagnostic.

## Validation and scope

- Both `fastdiag trace` commands completed with all seven numerical checks passing; both used the same input fingerprint and left deep `rescale` tracing disabled.
- `go test ./cmd/fastdiag` — **PASS**.
- No focused microbenchmarks were added or run because power-only overhead is below the specification's 10% threshold and category attribution is closed.
- Full Primary `go test ./...` was not run; the known unrelated `TestFIX001P3GenuineStandardPublicVsStagedConsistency` / `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT` debt remains untouched.
- No Secondary production code, tests, parameters, benchmark code, or one-off overlays were modified.
