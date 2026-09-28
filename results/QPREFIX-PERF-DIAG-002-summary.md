# QPREFIX-PERF-DIAG-002 — Matched q0=55 Stage-Delta Closure

**Primary classification:** `MATCHED_STAGE_ATTRIBUTION_CLOSED`\
**TOP_DELTA_STAGE:** `EvalMod real`\
**RESIDUAL_FRACTION:** `0.003343` (0.3343%)

## Provenance and matched workload

- Historical baseline: Secondary `40532b4dce5c7eeae2db5b0b6f21be64801ce923`, detached and clean before measurement.
- Q-prefix-v2 candidate: Secondary `f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`, detached and clean before measurement.
- The authoritative Secondary checkout stayed on `fast-qprefix`; no production code or branch was changed. Temporary test-only timing hooks were added only in disposable detached worktrees and removed after measurement.
- Machine/runtime: Apple M4, `go1.26.4`, `darwin/arm64`, `GOMAXPROCS=10`.
- Both runs use the same Primary-side LogN13/P93 profile, q0=55 (`q0=36028797018652673`), `LogSlots=12`, degree 1, `DoubleAngle=3`, `P=[61]*5`, and public entry `BootstrapMany`. `N1=N2=8192`.
- Q-chain bit sizes: `[55 39 40 39 40 60 60 61 60 60 60 61 61 56 57 56 56]`.
- Input length: 4096. Deterministic input is `complex(((j+3*i)%17-8)/256, ((3*j+i)%13-6)/512)` using float64 components. Count-1 SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; Count-3 SHA-256: `642fc6dca6a9a8d2acfadf14017a3438fd17e4c0836422e06f3150f795027051`.
- Clean uninstrumented replay windows (UTC): baseline `2026-09-28T09:06:17Z`–`2026-09-28T09:08:26Z`; candidate `2026-09-28T09:08:33Z`–`2026-09-28T09:08:49Z`.
- E2E commands (each run in its corresponding clean detached worktree; cache path matched the SHA):

  ```sh
  GOCACHE=/private/tmp/qpdiag002-cache-baseline-clean GOMAXPROCS=10 \
    go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag002/overlay-e2e.json \
    ./circuits/ckks/bootstrapping -run '^$' \
    -bench '^BenchmarkQPDiagMatchedP93Count(1|3)$' \
    -benchtime=1x -count=5 -benchmem -v

  GOCACHE=/private/tmp/qpdiag002-cache-candidate-clean GOMAXPROCS=10 \
    go test -overlay=/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/.qpdiag002/overlay-e2e.json \
    ./circuits/ckks/bootstrapping -run '^$' \
    -bench '^BenchmarkQPDiagMatchedP93Count(1|3)$' \
    -benchtime=1x -count=5 -benchmem -v
  ```

  The harness performs one warmup for each case; the five reported measurements are all retained below. The overlay and harness were temporary and have been removed.

## Uninstrumented matched E2E replay

Each row is one measured run: latency in ms / B/op / allocs/op.

| Bootstrap count | Baseline runs | Candidate runs |
|---|---|---|
| 1 | 28.208375 / 5,257,048 / 18,658<br>28.040125 / 5,257,048 / 18,658<br>28.063917 / 5,257,048 / 18,658<br>28.114291 / 5,257,096 / 18,659<br>28.150958 / 5,257,048 / 18,658 | 115.919417 / 7,615,224 / 15,951<br>116.130333 / 7,615,600 / 15,957<br>117.091542 / 7,615,224 / 15,951<br>115.942000 / 7,615,224 / 15,951<br>116.603750 / 7,615,224 / 15,951 |
| 3 | 83.966375 / 15,769,208 / 55,958<br>84.018625 / 15,769,208 / 55,958<br>84.184167 / 15,769,208 / 55,958<br>84.820375 / 15,769,328 / 55,961<br>84.286791 / 15,769,208 / 55,958 | 347.097792 / 22,845,736 / 47,848<br>347.265042 / 22,848,904 / 47,857<br>351.690666 / 22,848,528 / 47,851<br>351.239083 / 22,845,688 / 47,847<br>348.166000 / 22,845,688 / 47,847 |

| Count | Baseline median (ms; B/op; allocs/op) | Candidate median (ms; B/op; allocs/op) | Candidate / baseline | Delta |
|---|---:|---:|---:|---:|
| 1 | 28.114291; 5,257,048; 18,658 | 116.130333; 7,615,224; 15,951 | 4.1307× latency; 1.4486× bytes; 0.8549× allocs | +88.016042 ms; +2,358,176 B/op; −2,707 allocs/op |
| 3 | 84.184167; 15,769,208; 55,958 | 348.166000; 22,845,736; 47,848 | 4.1358× latency; 1.4488× bytes; 0.8551× allocs | +263.981833 ms; +7,076,528 B/op; −8,110 allocs/op |

The fresh replay retains a substantial, consistent regression at both batch counts. Candidate allocation volume rises about 44.9% while allocation count falls about 14.5%; allocation counts alone do not establish causality.

## In-context non-overlapping stage trace

Five measured Count-1 runs per SHA after warmup, in the same matched public Bootstrap invocation. Durations are ms. Stage order is `pack`, `ScaleDown`, `ModUp`, `C2S`, `EvalMod real`, `EvalMod imag`, `S2C`, `unpack`, `finalize`. `sum` includes only those nine non-overlapping stages; `residual = total − sum`. Nested timers were excluded from this sum.

| Run | Total | pack | ScaleDown | ModUp | C2S | EvalMod real | EvalMod imag | S2C | unpack | finalize | sum | residual |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| **Baseline 1** | 28.222666 | 0.005792 | 0.011791 | 0.250792 | 4.077125 | 9.897083 | 9.844500 | 4.111542 | 0.009583 | 0.010166 | 28.218374 | 0.004292 |
| **Baseline 2** | 28.017375 | 0.006500 | 0.013458 | 0.252334 | 4.079209 | 9.817834 | 9.703125 | 4.107792 | 0.018958 | 0.009875 | 28.009085 | 0.008290 |
| **Baseline 3** | 28.088583 | 0.007125 | 0.012750 | 0.240875 | 4.097500 | 9.852583 | 9.776791 | 4.074167 | 0.011166 | 0.010458 | 28.083415 | 0.005168 |
| **Baseline 4** | 28.156542 | 0.006709 | 0.012167 | 0.242000 | 4.131209 | 9.851458 | 9.819334 | 4.059167 | 0.012583 | 0.010083 | 28.144710 | 0.011832 |
| **Baseline 5** | 28.215709 | 0.006667 | 0.012292 | 0.244292 | 4.090708 | 9.878708 | 9.844291 | 4.111042 | 0.010792 | 0.011125 | 28.209917 | 0.005792 |
| **Candidate 1** | 116.011667 | 0.007542 | 0.014583 | 0.499417 | 15.185292 | 44.182166 | 43.918083 | 12.175292 | 0.011584 | 0.010000 | 116.003959 | 0.007708 |
| **Candidate 2** | 115.805916 | 0.006042 | 0.013875 | 0.441833 | 15.160167 | 44.161875 | 43.856667 | 12.138250 | 0.010708 | 0.010291 | 115.799708 | 0.006208 |
| **Candidate 3** | 116.774166 | 0.006125 | 0.012959 | 0.443042 | 15.447291 | 44.479125 | 44.227250 | 12.131708 | 0.010375 | 0.010125 | 116.768000 | 0.006166 |
| **Candidate 4** | 115.304125 | 0.006333 | 0.013041 | 0.441916 | 15.100208 | 44.009083 | 43.701000 | 12.001916 | 0.014416 | 0.010333 | 115.298246 | 0.005879 |
| **Candidate 5** | 116.820125 | 0.007916 | 0.014917 | 0.449209 | 15.156334 | 44.470041 | 44.566875 | 12.126500 | 0.011250 | 0.010208 | 116.813250 | 0.006875 |

## Stage deltas and closure

Stage values below are medians of each stage across the five in-context runs. Percent contribution uses the uninstrumented Count-1 `Delta T_E2E = 88.016042 ms`; the negative packing delta is preserved.

| Stage | Baseline median (ms) | Candidate median (ms) | Candidate / baseline | `Delta T_i` (ms) | Share of E2E delta |
|---|---:|---:|---:|---:|---:|
| pack / N1→N2 | 0.006667 | 0.006333 | 0.950× | −0.000334 | −0.0004% |
| ScaleDown | 0.012292 | 0.013875 | 1.129× | +0.001583 | 0.0018% |
| ModUp (including production Trace) | 0.244292 | 0.443042 | 1.814× | +0.198750 | 0.2258% |
| C2S | 4.090708 | 15.160167 | 3.706× | +11.069459 | 12.5766% |
| EvalMod real | 9.852583 | 44.182166 | 4.484× | +34.329583 | 39.0038% |
| EvalMod imaginary | 9.819334 | 43.918083 | 4.473× | +34.098749 | 38.7415% |
| S2C | 4.107792 | 12.131708 | 2.953× | +8.023916 | 9.1164% |
| N2→N1 / unpack | 0.011166 | 0.011250 | 1.008× | +0.000084 | 0.0001% |
| public finalization | 0.010166 | 0.010208 | 1.004× | +0.000042 | 0.0000% |
| **Sum of stage deltas** |  |  |  | **+87.721832** | **99.6657%** |

Closure uses the fresh uninstrumented Count-1 medians:

```text
Delta T_E2E = 116.130333 − 28.114291 = 88.016042 ms
R = Delta T_E2E − sum(Delta T_i) = 88.016042 − 87.721832 = 0.294210 ms
residual_fraction = abs(R) / abs(Delta T_E2E) = 0.003343 (0.3343%)
```

Median per-run stage-sum coverage of the instrumented total is 99.9580% for baseline (`28.144710 / 28.156542`) and 99.9934% for candidate (`116.003959 / 116.011667`).

Instrumentation overhead, comparing the in-context traced total median to the clean uninstrumented Count-1 median, was +0.042251 ms (+0.1503%) for baseline and −0.118666 ms (−0.1022%) for candidate. The absolute observed difference is below 0.16% for both, well below the 5% limit; no positive overhead was measurable for candidate at this run variance.

## Single bounded drill-down: EvalMod real

This is the only stage drilled down, selected because it has the largest positive top-level delta. Values are medians in ms; each substage was measured within EvalMod real. DoubleAngle rounds are nested in its total and must not be added again.

| EvalMod-real substage | Baseline median | Candidate median | Delta | Candidate / baseline |
|---|---:|---:|---:|---:|
| generated powers | 5.725250 | 26.723459 | +20.998209 | 4.668× |
| PS baby steps | 0.274959 | 0.520167 | +0.245208 | 1.892× |
| PS giant-step merges | 1.897333 | 8.465625 | +6.568292 | 4.462× |
| final PS Rescale | 0.396250 | 1.993125 | +1.596875 | 5.030× |
| PS result finalization | 0.010083 | 0.022875 | +0.012792 | 2.269× |
| DoubleAngle total | 1.353084 | 6.233125 | +4.880041 | 4.607× |
| └ round 0 (nested) | 0.451917 | 2.085792 | +1.633875 | 4.615× |
| └ round 1 (nested) | 0.450083 | 2.081250 | +1.631167 | 4.624× |
| └ round 2 (nested) | 0.450750 | 2.072458 | +1.621708 | 4.598× |

The non-overlapping substage total is 9.656959 ms baseline vs 43.958376 ms candidate (`+34.301417 ms`), leaving 0.028166 ms (0.0820% of the EvalMod-real delta) within that stage unassigned at this drilldown depth. Generated-power construction is the largest measured substage contribution; this is diagnosis only and does not authorize optimization.

Separate EvalMod-real allocation benchmark input was formed by the same ScaleDown→ModUp→C2S path and reached Level 12, Scale `1.125900e15`, degree 1. The evaluator clones the input, so it did not mutate the shared input. Five-run B/op and allocs/op medians: baseline `1,793,048 B/op`, `8,544 allocs/op`; candidate `2,540,120 B/op`, `7,307 allocs/op` (`+747,072 B/op`, +41.7%; −1,237 allocs/op, −14.5%). The isolated timing series was variable and is not used in stage closure.

## Interpretation and remaining uncertainty

The fresh matched q0=55 replay reproduces an approximately 4.13× latency regression at both Count 1 and Count 3. The in-context accounting assigns 99.6657% of the Count-1 excess to measured top-level stages, with a 0.3343% absolute residual fraction. `EvalMod real` is the largest positive single-stage delta (39.0038% of the E2E excess), narrowly ahead of `EvalMod imaginary`; C2S and S2C are also material contributors. Within the one permitted drilldown, generated powers account for the largest EvalMod-real substage delta.

The work isolates where this matched run spends additional time; it does not determine whether any stage cost is avoidable or authorize a code change. Remaining uncertainty is ordinary run-to-run scheduling/thermal variation on a single Apple M4 host and the limited five-sample set. No other stage was drilled down, and no parameter, schedule, kernel, or production source was changed.

## Temporary artifacts and repository scope

All temporary test harnesses, overlays, timing hooks, and four task-created Secondary worktrees have been removed. The authoritative Secondary remains clean at `fast-qprefix` HEAD `1a4fd3d08a2c666f3203b857adb65ee11a9627e8`; no Secondary production commit was created. The only persistent Primary change for this task is this result artifact.
