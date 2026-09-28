# QPREFIX-PERF-DIAG-004 — Post-Optimization Hotspot Refresh

**Classification:** `CURRENT_RESCALE_STILL_DOMINANT`

## Provenance

- Primary: `60f9a747df359e44a8db426e9cae4cc8906863e8`, `main`, clean during measurements.
- Pre-optimization Secondary ref: `74cb73dcff6c552cda0671ed7faea897b448fbbd`.
- Post-optimization production candidate: `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`.
- Secondary was safely fast-forwarded to `0946485e7601bb4e58cc4ad0dec474dce24469fd` before measurement. The only difference between that control-plane commit and `d50ff4d` is `CURRENT_TASK.md`; no production source changed. Current traces therefore use the current branch at `0946485e`, code-equivalent to the pinned candidate. The framework comparison uses the exact pinned refs above.
- Profile: `p93-q55`, LogN13, LogSlots12, 4096 input values, polynomial degree 30, DoubleAngle 3. Stage/deep traces and both compare refs used the same input SHA-256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.
- Environment: Go `go1.26.4`, macOS `darwin/arm64`, Apple M4, 10 CPUs, `GOMAXPROCS=10`.

## Diagnostics-off E2E refresh

The existing `BenchmarkFastDiagP93Q55Count1` performs one untimed Bootstrap warmup before resetting the timer. Seven independent one-iteration samples were collected with `-benchmem`:

```text
78.185625, 81.514083, 78.994459, 78.174708, 78.424542, 79.208083, 81.301833 ms/op
```

Median: **78.994459 ms/op**. Every sample reported **7,615,224 B/op** and **15,951 allocs/op**. This is the authoritative diagnostics-off E2E result; trace wall times below are not substitutes for it.

The repository has a Count-3 `BenchmarkFastBootstrapManyLogN13P93Count3`, but its existing fixture uses residual `LogQ={56,39}` rather than the required q0=55 profile. No matching q0=55 Count-3 benchmark exists, so it was not treated as a comparable measurement and no new benchmark was added.

## Current non-overlapping stages

Stage-only trace: 1 warmup, 7 measured repetitions. Shares use the sum of non-overlapping child-stage medians (`80.257876 ms`) as the denominator.

| Stage | Median (ms) | Share |
|---|---:|---:|
| pack / N1→N2 | 0.005542 | 0.01% |
| ScaleDown | 0.011417 | 0.01% |
| ModUp including Trace | 0.439167 | 0.55% |
| C2S | 10.716542 | 13.35% |
| EvalMod real | 30.063916 | 37.46% |
| EvalMod imaginary | 29.866583 | 37.21% |
| S2C | 9.134167 | 11.38% |
| unpack / N2→N1 | 0.010375 | 0.01% |
| public finalization | 0.010167 | 0.01% |

The Bootstrap parent median was `80.279334 ms`; child-median closure residual was `0.021458 ms` (`0.027%` of the parent). Relative to diagnostics-off E2E, stage-only tracing overhead was about **1.63%**. All seven runs passed the framework's numerical check. `evalmod_real` is the largest stage, narrowly ahead of `evalmod_imag`.

## Current generated-power and Rescale structure

The `stage,power,rescale` trace used 1 warmup and 5 measured repetitions. Its Bootstrap parent median was `145.225 ms`, about 81% above the stage-only parent, confirming that all-scope timing is structural rather than production E2E evidence. Within that same trace, generated-power parents were `34.205 ms` per EvalMod call (`68.474 ms` total across the real and imaginary calls); deep-traced EvalMod stage medians were about `57.2 ms` each.

The inclusive per-call power spans (median over the real/imaginary observations) were:

| Generated power | Median span (ms) |
|---:|---:|
| 2 | 5.942 |
| 3 | 5.794 |
| 4 | 11.535 |
| 6 | 11.703 |
| 8 | 16.980 |
| 16 | 22.442 |

These power spans are hierarchical/inclusive and must not be summed. The direct category spans below are aggregated across both EvalMod calls per Bootstrap; the residual is calculated per run from the generated-power parent minus its measured category spans.

| Direct generated-power category | Median total (ms / Bootstrap) |
|---|---:|
| copy / workspace | 0.252 |
| balanced integer scaling | 0.637 |
| Mul | no standalone event emitted (`0`) |
| MulRelin | 0.726 |
| Rescale | 66.062 |
| Chebyshev doubling | 0.200 |
| recurrence correction | 0.203 |
| residual / unclassified | 0.372 |

Current generated-power Rescale share was **96.51%** of generated-power parent time (median of per-run ratios). Thus Rescale remains the dominant power category by a wide margin.

The detailed `rescale` scope covers all Bootstrap Rescale calls, not only calls made while generating powers. It recorded 47 Rescale parent events per run:

| All-Bootstrap Rescale timing | Median total (ms / Bootstrap) |
|---|---:|
| Rescale total | 130.044 |
| preflight / compute-and-stage | 116.911 |
| materialization commit | 13.091 |
| prefix-to-coefficient | 13.935 |
| coefficient loop | 102.790 |
| reconstruct / center / round / capacity | 31.958 |
| residue staging | 19.655 |
| NTT / Montgomery restore | 13.013 |

Rescale total closes against preflight plus materialization within `0.032 ms`; preflight closes against prefix conversion plus coefficient loop within `0.080 ms`. The measured reconstruction and residue-staging subphases leave about `51.095 ms` of coefficient-loop time unassigned to those two events. Materialization is small now; the remaining aggregate Rescale time is predominantly preflight.

## Pre-optimization versus post-optimization framework comparison

Reusable compare: baseline `74cb73dcff6c552cda0671ed7faea897b448fbbd`, candidate `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`, same profile/fingerprint, 1 warmup and 5 repetitions. Both refs reported `READY`; all runs matched numerically.

| Non-overlapping stage | Baseline (ms) | Candidate (ms) | Δ (ms) | Share of traced Bootstrap improvement |
|---|---:|---:|---:|---:|
| Bootstrap parent | 212.116 | 150.697 | -61.420 | 100% |
| EvalMod real | 85.286 | 59.495 | -25.791 | 42.0% |
| EvalMod imaginary | 84.810 | 59.473 | -25.337 | 41.3% |
| C2S | 23.003 | 17.036 | -5.967 | 9.7% |
| S2C | 18.026 | 13.851 | -4.175 | 6.8% |

Other stage changes and parent closure account for the remaining approximately `0.150 ms` (`0.2%`). Bootstrap parent closure residual was `0.176 ms` (`0.286%`). The two generated-power parent deltas were `-15.495 ms` and `-15.342 ms`; their child closures were `0.066 ms` and `-0.027 ms` respectively (both below `0.5%` of their parent deltas).

The direct generated-power `power/rescale` spans decreased from a combined **100.049 ms** to **68.780 ms**, a `31.269 ms` reduction: **50.9%** of the measured `61.420 ms` all-scope Bootstrap improvement. A non-overlapping decomposition of that improvement is:

| Attribution bucket | Improvement (ms) | Fraction |
|---|---:|---:|
| generated-power Rescale | 31.269 | 50.9% |
| remaining EvalMod work (EvalMod stage delta less generated-power Rescale) | 19.859 | 32.3% |
| C2S + S2C | 10.142 | 16.5% |
| other stages + parent closure | 0.150 | 0.2% |

This is event-delta attribution, not proof that the Rescale change alone caused every saved nanosecond. Across **all** Bootstrap Rescale events, total time changed `197.475 → 134.949 ms`; preflight share changed `34.77% → 89.39%`, while materialization share changed `65.21% → 10.57%`. Those scope-wide values include non-power Rescales and are not added again to the stage attribution above.

## Framework validation, classification, and limits

- `trace --trace stage`: succeeded; stage closure is tight and measured overhead is 1.63% versus diagnostics-off E2E.
- `trace --trace stage,power,rescale`: succeeded; the large all-scope overhead is explicitly treated as structural only.
- `compare`: succeeded for both hook-enabled refs; matching workload fingerprint and numerical checks; top-level Bootstrap and generated-power parent closures are consistent.
- Event-schema limitation: low-level `scope=rescale` root events have no `parent_sequence` linking them to a `scope=power` Rescale event. Their preflight/materialization subphases therefore cannot be isolated specifically to generated-power calls. The report keeps those all-Bootstrap totals separate from the direct generated-power Rescale category; this limits phase-level attribution but does not block the stage or power-category ranking.
- Historical q01 context only: `C2S ≈4.09 ms`, EvalMod real `≈9.85 ms`, EvalMod imaginary `≈9.82 ms`, `S2C ≈4.11 ms`, total `≈28.1 ms`, labeled `ARCHIVED_CROSS_SESSION_REFERENCE_ONLY`; no causal percentages are computed from it.
- Existing uncertainty: the coefficient-loop subphases are not exhaustive, and the deep trace has substantial instrumentation overhead. No parameter, schedule, or implementation change is recommended or authorized here.

**Fresh investigation target (not an implementation plan):** inspect the remaining Q-prefix Rescale preflight/coefficient-loop time and close its currently unassigned phase share before selecting any next optimization.

## Commands

```sh
go test ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkFastDiagP93Q55Count1$' -benchtime=1x -benchmem -count=7
./scripts/fastdiag trace --profile p93-q55 --trace stage --warmup 1 --repetitions 7
./scripts/fastdiag trace --profile p93-q55 --trace stage,power,rescale --warmup 1 --repetitions 5
./scripts/fastdiag compare --baseline 74cb73dcff6c552cda0671ed7faea897b448fbbd --candidate d50ff4db757d4a2b9922937a4e7f316fd3f286b9 --profile p93-q55 --trace stage,power,rescale --warmup 1 --repetitions 5
go test ./cmd/fastdiag
```

`go test ./cmd/fastdiag` passed. No Secondary production code was modified. The known unrelated Primary debt `TestFIX001P3GenuineStandardPublicVsStagedConsistency` / `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT` remains documented and untouched; the full Primary suite was not rerun for this measurement-only task.
