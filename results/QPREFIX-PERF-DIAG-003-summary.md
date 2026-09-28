# QPREFIX-PERF-DIAG-003 — Generated-Power Rescale Causality

**Primary classification:** `GENERATED_POWER_RESCALE_DOMINANT`\
**RESCALE_DELTA_SHARE:** `0.977205` (97.7205%)\
**PREFLIGHT_TIME_SHARE_OF_POWER_DELTA:** `0.443635` (44.3635%)\
**GENERATED_POWER_RESIDUAL_FRACTION:** `0.000423` (0.0423%; median orchestration-residual delta / median power delta)

## Provenance and matched workload

- Historical baseline: Secondary `40532b4dce5c7eeae2db5b0b6f21be64801ce923`.
- Q-prefix-v2 candidate: Secondary `f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`.
- The authoritative Secondary checkout remained clean on `fast-qprefix` at `0c2ce7bcbd7b46f0f745509cd85040af3db5c85d`. No Secondary production change or commit was made.
- Primary was clean on `main` at `ce2083dd28ef86e37d8e14f5a91826be232da4ac` before this result artifact was added.
- Runtime: Apple M4, `go1.26.4`, `darwin/arm64`.
- Measurements were strictly sequential: the baseline's warmup plus five samples completed before the candidate started. Each SHA ran in its own detached worktree with an isolated temporary `GOCACHE`; no baseline/candidate timing processes overlapped.
- Both worktrees used the accepted QPREFIX-PERF-DIAG-002 q0=55, P93 real EvalMod workload: LogN=13, LogSlots=12, `LogQ=[55,39]`, `P=[61]*5`, degree-30 Chebyshev polynomial, `DoubleAngle=3`, K=16, and the same deterministic 4096-value input.
- Input Count-1 SHA-256 on both: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.
- Both reached EvalMod-real input Level 12, Scale `1.125899906843e+15`, degree 1, polynomial degree 30, basis 1, `DoubleAngle=3`, and q-chain bit lengths `[55 39 40 39 40 60 60 61 60 60 60 61 61 56 57 56 56]`.
- The representation hashes differ as expected because the authoritative row counts differ: baseline input rows=2 / hash `41c5255a27e3d811774c6b6c72a2b9d951cc29a97082c038dfb94736449ca296`; candidate rows=4 / hash `cec3fa9936688ee1e8bb55910b234706f4afc6e81453bc30ae75a953c8fe931b`. These are not workload-input differences.
- The six repeated EvalMod-real outputs were stable within each SHA and match the uninstrumented reference hashes: baseline `97394c8de6c9664a12efbd4e7d358edb25cf273bc76852f90b39c4baf7e83b1b`; candidate `317219272521afe5fa44fd7795d8acfd38187f36f8593c9b82e09eb6a6cfb26c`.

## Runtime generated-power DAG and operation counts

Both SHA runs generated exactly `{2,3,4,6,8,16}` using the same `(a,b)` splits, levels, and schedule branches. All generated powers were non-lazy, selected balanced pre-Rescale (`true`), and did not select the post-product schedule (`false`). The two input scales for every row below were both `1.152922e+18` in both implementations. Baseline/candidate authoritative rows were 2/4 at every listed level.

| Power n | Split (a,b) | Input levels | Output level | Operand copies | Integer-scale multiplies | Relinearize | Mul | MulRelin | Rescale | Doubling Add | Correction |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 2 | (1,1) | (12,12) | 11 | 2 | 2 | 0 | 0 | 1 | 2 | 1 | subtract one |
| 3 | (1,2) | (12,11) | 10 | 2 | 2 | 0 | 0 | 1 | 2 | 1 | aligned subtraction |
| 4 | (2,2) | (11,11) | 10 | 2 | 2 | 0 | 0 | 1 | 2 | 1 | subtract one |
| 6 | (3,3) | (10,10) | 9 | 2 | 2 | 0 | 0 | 1 | 2 | 1 | subtract one |
| 8 | (4,4) | (10,10) | 9 | 2 | 2 | 0 | 0 | 1 | 2 | 1 | subtract one |
| 16 | (8,8) | (9,9) | 8 | 2 | 2 | 0 | 0 | 1 | 2 | 1 | subtract one |

Counts above come from runtime hooks. The workspace-preparation timing category includes the two operand copies and one output power-buffer preparation; the explicit operand-copy count remains two.

## In-context generated-power timing

Times are milliseconds except where noted. Each raw row is one measured run after the warmup. Category timings are non-overlapping; `Residual` is phase elapsed minus the measured categories.

| Run | T powers B/C (ms) | Rescale B/C (ms) | Copy/workspace B/C (µs) | Integer scale B/C (µs) | Mul/MulRelin B/C (µs) | Double B/C (µs) | Correction B/C (µs) | Residual B/C (µs) |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 5.913291 / 26.865667 | 5.149457 / 25.664542 | 71.957 / 122.042 | 196.045 / 314.625 | 185.332 / 357.748 | 52.999 / 97.959 | 56.043 / 100.832 | 201.458 / 207.919 |
| 2 | 5.997625 / 27.019250 | 5.258292 / 25.805458 | 73.751 / 123.748 | 170.625 / 317.500 | 185.126 / 361.083 | 53.417 / 97.250 | 56.666 / 103.917 | 199.748 / 210.294 |
| 3 | 5.976209 / 26.972458 | 5.215668 / 25.628957 | 72.334 / 139.500 | 180.916 / 396.168 | 186.583 / 354.500 | 54.000 / 96.916 | 58.750 / 123.208 | 207.958 / 233.209 |
| 4 | 5.948083 / 26.855000 | 5.211751 / 25.588791 | 69.870 / 129.917 | 169.625 / 327.289 | 179.124 / 383.417 | 50.584 / 97.251 | 59.500 / 101.250 | 207.629 / 227.085 |
| 5 | 6.030416 / 26.792000 | 5.307831 / 25.587123 | 69.081 / 126.043 | 168.583 / 320.499 | 181.042 / 359.166 | 52.332 / 99.999 | 58.000 / 101.500 | 193.547 / 197.670 |

Five-run medians and candidate-minus-baseline deltas:

| Non-overlapping category | Baseline median (ms) | Candidate median (ms) | Delta (ms) |
|---|---:|---:|---:|
| Copy / workspace | 0.071957 | 0.126043 | +0.054086 |
| Balanced integer scaling | 0.170625 | 0.320499 | +0.149874 |
| Relinearize | 0.000000 | 0.000000 | +0.000000 |
| Mul / MulRelin | 0.185126 | 0.359166 | +0.174040 |
| Rescale total | 5.215668 | 25.628957 | +20.413289 |
| Chebyshev doubling | 0.052999 | 0.097251 | +0.044252 |
| Recurrence correction | 0.058000 | 0.101500 | +0.043500 |
| Orchestration residual | 0.201458 | 0.210294 | +0.008836 |
| **Generated-power phase** | **5.976209** | **26.865667** | **+20.889458** |

The measured categories cover 96.52% of baseline and 99.23% of candidate phase time at the median. The differences of independently computed category medians sum to 20.887877 ms, 1.581 µs below the phase-median delta; this is median non-additivity, not an unmeasured operation. The median residual delta is 8.836 µs, or 0.0423% of the power delta.

| Power | Baseline median (ms) | Candidate median (ms) | Delta (ms) |
|---:|---:|---:|---:|
| 2 | 1.072375 | 4.980208 | +3.907833 |
| 3 | 1.022875 | 4.608291 | +3.585416 |
| 4 | 0.937667 | 4.151500 | +3.213833 |
| 6 | 1.055250 | 4.961958 | +3.906708 |
| 8 | 0.922542 | 4.063750 | +3.141208 |
| 16 | 0.918666 | 4.043792 | +3.125126 |

Per-power times are the measured power-call spans after their split operands were available. Slowdown is distributed across all six powers (about 3.13–3.91 ms added per power), rather than isolated to one power. Level and row authority match that distribution: candidate source rows are four versus baseline two at every Rescale.

## Matched Rescale timing and row authority

Each generated power performs two in-place Rescales. Durations below are the five-run median for the pair of calls constructing that power. Baseline arithmetic includes q01 CRT, rounded division, and the per-coefficient q0/q1 residue writes; those writes share the coefficient loop with reconstruction and division. Candidate materialization arithmetic includes reconstruction, rounded division, capacity-safe residue writes; NTT/Montgomery restore is separately timed.

| Power | Source Level → target Level | Source→target rows (baseline / candidate) | Calls, in-place B/C | Baseline total (ms) | Baseline convert (ms) | Baseline q01 arithmetic + residue (ms) | Baseline NTT/MForm (ms) | Baseline other (µs) | Candidate total (ms) | Candidate preflight (ms) | Candidate materialization (ms) |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 2 | 12→11 | 2→2 / 4→4 | 2, yes / 2, yes | 0.947124 | 0.301416 | 0.353709 | 0.285125 | 3.833 | 4.768958 | 1.785625 | 2.983042 |
| 3 | 11→10 | 2→2 / 4→4 | 2, yes / 2, yes | 0.874875 | 0.298418 | 0.289334 | 0.283959 | 1.958 | 4.362166 | 1.578583 | 2.783292 |
| 4 | 11→10 | 2→2 / 4→4 | 2, yes / 2, yes | 0.813291 | 0.299292 | 0.227626 | 0.285041 | 2.003 | 3.969875 | 1.404042 | 2.560042 |
| 6 | 10→9 | 2→2 / 4→4 | 2, yes / 2, yes | 0.939333 | 0.299625 | 0.353210 | 0.284750 | 2.001 | 4.770250 | 1.759667 | 2.976125 |
| 8 | 10→9 | 2→2 / 4→4 | 2, yes / 2, yes | 0.811666 | 0.301875 | 0.216250 | 0.281916 | 2.334 | 3.882292 | 1.361625 | 2.511959 |
| 16 | 9→8 | 2→2 / 4→4 | 2, yes / 2, yes | 0.805791 | 0.302584 | 0.214834 | 0.284833 | 1.999 | 3.860999 | 1.356708 | 2.500124 |

## Candidate preflight/materialization and causal accounting

Across the twelve generated-power Rescale calls, the five-run medians are:

| Run after warmup | Preflight (ms) | Materialization (ms) | Rescale total (ms) | Residual (ms) |
|---:|---:|---:|---:|---:|
| 1 | 9.289373 | 16.373209 | 25.664542 | 0.001960 |
| 2 | 9.339876 | 16.462874 | 25.805458 | 0.002708 |
| 3 | 9.255793 | 16.371459 | 25.628957 | 0.001705 |
| 4 | 9.257959 | 16.327294 | 25.588791 | 0.003538 |
| 5 | 9.267292 | 16.317875 | 25.587123 | 0.001956 |

Five-run medians:

| Candidate Rescale part | Median (ms) |
|---|---:|
| Preflight total | 9.267292 |
| └ prefix-to-coefficient conversion | 3.516502 |
| └ reconstruction / centering / rounded division / capacity check | 5.770292 |
| Materialization total | 16.371459 |
| └ prefix-to-coefficient conversion | 3.496124 |
| └ reconstruction / rounding / residue materialization | 9.528586 |
| └ NTT / Montgomery restore | 3.286794 |
| Rescale-call residual (`total − preflight − materialization`) | 0.001960 |
| Rescale total | 25.628957 |

For each raw Rescale call, `total = preflight + materialization + residual` by construction. Independently computed medians do not sum exactly: median preflight + median materialization + median residual is 25.640711 ms, 11.754 µs above the median total. The pass subcomponent medians are likewise independent; per-run preflight/materialization residuals account for their small gaps.

```text
Delta T_powers = 26.865667 − 5.976209 = 20.889458 ms
Delta T_rescale = 25.628957 − 5.215668 = 20.413289 ms
RESCALE_DELTA_SHARE = Delta T_rescale / Delta T_powers = 0.977205
PREFLIGHT_TIME_SHARE_OF_POWER_DELTA = 9.267292 / 20.889458 = 0.443635
GENERATED_POWER_RESIDUAL_FRACTION = (0.210294 − 0.201458) / 20.889458 = 0.000423
```

The counterfactual arithmetic-only estimate with repeated candidate preflight subtracted is `26.865667 − 9.267292 = 17.598375 ms`. This is not a safe speedup claim and does not authorize removing preflight: transactional failure-before-mutation semantics must remain intact. The observed two-pass preflight is a measured optimization candidate for a later design task, not a correctness-safe repair.

The classification gate is met: the directly matched Rescale delta explains 97.72% of `Delta T_powers`, above the 60% threshold. Per-category median differences leave only a 1.581 µs accounting closure difference from median non-additivity.

## Focused validation and cleanup

The instrumented harness was run at both immutable SHAs. Its input checksums and every repeated EvalMod output hash match the accepted uninstrumented references recorded above. Focused polynomial and Rescale regression tests passed at both SHAs, including the candidate's transactional Q-prefix-bound check. Commands (each executed in the corresponding detached worktree):

```sh
GOCACHE=/private/tmp/qpdiag003-baseline-cache.fTRt9Z go test -v ./circuits/ckks/bootstrapping -run '^TestQPDiag003MatchedPowerPhase$' -count=1
GOCACHE=/private/tmp/qpdiag003-candidate-cache.FlK3vb go test -v ./circuits/ckks/bootstrapping -run '^TestQPDiag003MatchedPowerPhase$' -count=1

GOCACHE=/private/tmp/qpdiag003-baseline-cache.fTRt9Z go test ./circuits/ckks/polynomial ./schemes/ckks/fast -run 'TestFastPolynomial(PowerBasisQPrefixRowsOracle|NonPowerOfTwoChebyshevOracle|RepresentativeDegree30)$|TestFastRescale(MatchesStandardAtHigherLevels|SupportsBothNTTRepresentations|MatchesBigIntCenteredCRTOracleAtEveryPrefixWidth)$' -count=1
GOCACHE=/private/tmp/qpdiag003-candidate-cache.FlK3vb go test ./circuits/ckks/polynomial ./schemes/ckks/fast -run 'TestFastPolynomial(PowerBasisQPrefixRowsOracle|NonPowerOfTwoChebyshevOracle|RepresentativeDegree30)$|TestFastRescale(MatchesStandardAtHigherLevels|SupportsBothNTTRepresentations|MatchesBigIntCenteredCRTOracleAtEveryPrefixWidth)$|TestCommitQPrefixBoundsIsTransactional$' -count=1
```

All four commands exited 0. Temporary harnesses and timing hooks were deleted, both temporary detached worktrees were clean before removal, and both worktrees and their temporary Go caches were removed. The authoritative Secondary remained unchanged and clean at `fast-qprefix` HEAD `0c2ce7bcbd7b46f0f745509cd85040af3db5c85d`. No parameter, schedule, `QPrefixWidth(Level)`, production source, or transactional behavior was changed. This diagnosis does not establish a pure row-width-only counterfactual: the measured implementation difference couples four-row authority with the two-pass preflight/materialization path.
