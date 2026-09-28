# QPREFIX-PERF-DIAG-006 — Rescale Kernel Attribution

**Classification:** `RESCALE_MIXED_KERNELS`\
**TOP_RESCALE_KERNEL:** fixed-width reconstruct / round / capacity phase\
**TOP_RESCALE_SHARE:** `0.4351` of canonical full rows4 Rescale median\
**ROWS4_ROWS2_FULL_RATIO:** `2.199x` (`WIDTH_COUNTERFACTUAL_ONLY`)\
**PHASE_CLOSURE_RATIO:** `1.128`

## Scope and provenance

This is a diagnostics-off LogN13/P93 q0=55 measurement. No production arithmetic, width policy, parameter, generated-power schedule, or benchmark production path was changed. Secondary's temporary same-package fixture/benchmarks were removed after measurement; no Secondary commit was made.

- Primary `main`: `7bf70f10c3d0f03a140d04484bb5b4dad04ed80b` at measurement start.
- Secondary `fast-qprefix`: `96e5b11eedc9b5c3f582421fe1c12ffb683e9b44`.
- Secondary production-code comparison against required control `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`: only `CURRENT_TASK.md` differs; production source is code-equivalent.
- Machine reported by Go benchmark: `darwin/arm64`, Apple M4.

The rows2 fixture's source values were generated under rows2 authority and reused unchanged for rows4. This is a kernel-width counterfactual only; it makes no claim that rows2 is policy-eligible at the active P93 states. Both non-Montgomery (the existing benchmark's representation) and Montgomery-matched controls were measured.

## Authoritative full rows4 Rescale

Diagnostics-off command:

```text
go test ./schemes/ckks/fast -run '^$' -bench '^BenchmarkFastRescaleQPrefixRows4LogN13P93$' -benchtime=100ms -benchmem -count=7
```

Seven raw samples (`ns/op`): `1244084, 1321834, 1224771, 1204722, 1259439, 1220883, 1214449`.

- Median: **1.224771 ms/op**
- Allocation result: **664–665 B/op, 18 allocs/op**
- Median ordering was computed over all seven samples; none was dropped.

## CPU profile

The focused 5-second profile used only `BenchmarkFastRescaleQPrefixRows4LogN13P93`, with diagnostics disabled:

```text
go test ./schemes/ckks/fast -run '^$' -bench '^BenchmarkFastRescaleQPrefixRows4LogN13P93$' -benchtime=5s -benchmem -cpuprofile=/tmp/qpdiag006-rows4-cpu.pprof -o /tmp/qpdiag006-fast.test
```

It collected 4,827 iterations (reported 1.244929 ms/op; 676 B/op; 18 allocs/op; approximately 5.51 seconds of profile samples). This fixture is non-Montgomery, so the profile does not measure explicit IMForm/MForm work. Go 1.26.4 did not provide `go tool pprof`; the cached `github.com/google/pprof` CLI was used for `-top` and focused `-list` inspection of the same profile.

Selected pprof top entries (flat / cumulative sample share):

| Function | Flat | Cumulative |
|---|---:|---:|
| `math/bits.Div64` | 27.40% | 27.59% |
| `ring.inttLazyUnrolled16` | 16.33% | 21.42% |
| `ring.nttUnrolled16Lazy` | 11.43% | 16.33% |
| `(*Evaluator).rescaleQPrefixComponent` | 7.99% | 73.87% |
| `ring.MRedLazy` | 7.62% | 8.53% |
| `fast.mod192By64` | 4.90% | 23.59% |

Focused `-list` results included cumulative shares of 23.23% for `prefixToCoefficientRows`, 19.06% for `reconstructQPrefix`, 18.33% for `signedResidue192`, and 3.09% for `roundedMagnitude192`. These inclusive/call-tree figures overlap and must not be summed. In particular, division helpers are shared/inlined across exact arithmetic and residue work, so pprof alone does not support an exclusive per-category partition. It does corroborate material domain-transform and fixed-width arithmetic cost.

## Phase benchmarks and closure

Same-package test-only fixture phases were checked row-by-row against both actual Fast production Rescale output and the Standard Rescale oracle, for rows2/rows4 and Montgomery false/true. All phase loops reported 0 B/op and 0 allocs/op. The rows2/rows4 medians below use the matched non-Montgomery fixture; each is the median of seven `-benchtime=100ms` samples.

| Phase | rows2 median (ms) | rows4 median (ms) | rows4/rows2 | rows4 phase / canonical full |
|---|---:|---:|---:|---:|
| Prefix → coefficient rows (INTT; IMForm only if applicable) | 0.137312 | 0.276784 | 2.016x | 22.60% |
| Fixed-width reconstruct / center / round / capacity | 0.220440 | 0.532853 | 2.417x | 43.51% |
| Signed residue staging / target-row writes | 0.176640 | 0.323381 | 1.831x | 26.40% |
| NTT / Montgomery restore | 0.123319 | 0.248345 | 2.014x | 20.28% |
| Full production API (matched width fixture) | 0.555357 | 1.221488 | **2.199x** | — |

Rows4 phase medians sum to **1.381363 ms**. Dividing by the authoritative canonical full median (1.224771 ms) gives **1.128 closure**, within the spec's 0.80–1.20 phase-ranking band. No values were rescaled to force closure. The four independent phase timings total 112.8% because they run in separate microbenchmarks/cache contexts; the percentages are coarse comparisons to full Rescale, not a normalized additive partition.

Montgomery-true matched results support the same ranking: rows2 / rows4 full medians were 0.572113 / 1.276900 ms (2.232x); phase medians were prefix 0.145696 / 0.290373 ms, fixed-width 0.222433 / 0.526455 ms, staging 0.178196 / 0.319296 ms, and restore 0.138241 / 0.278439 ms. Its phase closure was 110.8%. The fixed-width phase remained largest but below 45% of full Rescale.

## Generated-power sensitivity arithmetic

DIAG-005 reports **34.525831 ms** combined generated-power Rescale time per Bootstrap and **36.826541 ms** for the generated-power parent. The following arithmetic applies each focused rows4 phase's fraction of the canonical full Rescale median to the combined Rescale total, then hypothetically reduces that phase by 25% or 50%:

| Candidate phase | 25% reduction: upper-bound impact | 50% reduction: upper-bound impact |
|---|---:|---:|
| Prefix/domain conversion | 1.951 ms (5.30% of parent) | 3.901 ms (10.59%) |
| Fixed-width reconstruct / round / capacity | 3.755 ms (10.20%) | 7.510 ms (20.39%) |
| Residue staging | 2.279 ms (6.19%) | 4.558 ms (12.38%) |
| NTT / Montgomery restore | 1.750 ms (4.75%) | 3.500 ms (9.51%) |

These are sensitivity bounds, not speedup predictions. The arithmetic assumes the focused Rescale's phase distribution represents the generated-power calls; that distribution was not measured per call, so the values should not be interpreted as expected Bootstrap savings. The rows are independent hypotheticals and are not additive.

## Classification and bounded next candidate

`RESCALE_MIXED_KERNELS`: the largest measured phase is fixed-width reconstruct/round/capacity at **43.51%**, below the required 45% dominance threshold. Domain conversion, residue staging, and restore are also material; the CPU profile independently shows INTT/NTT kernels and exact-division helpers among the largest samples. Attribution is sufficiently closed for a mixed classification, but not for naming a single dominant subsystem.

Exactly one bounded next candidate: **investigate/optimize the fixed-width coefficient loop in `rescaleQPrefixComponent`**, specifically its `reconstructQPrefix` → `centeredQPrefix` → sequential `roundedMagnitude192` / target-capacity path. This is the largest isolated phase and the pprof list places its helpers on the hot call path. Preserve exact centered CRT interpretation for every supported prefix width, sequential rounded-division semantics for each removed modulus, the target-width capacity guard, and byte-for-byte output residues plus metadata/transactionality behavior. No optimization was implemented here.

## Validation and repository state

- `go test ./schemes/ckks/fast -run '^TestFastRescaleKernelPhaseFixtureMatchesProductionAndStandard$' -count=1` — **PASS** for rows2/rows4 and Montgomery false/true.
- `go test ./schemes/ckks/fast -count=1` — **PASS**; includes existing Rescale correctness/oracle and transactionality tests.
- Both seven-sample focused benchmark commands — **PASS**.
- `git diff --check` in Secondary after removing the temporary fixture — **PASS**; no Secondary changes remain.
- Secondary worktree: **clean**, branch `fast-qprefix` at `96e5b11eedc9b5c3f582421fe1c12ffb683e9b44`; no production files or commits changed.
- Primary `go test ./...` was not run; it is not required for this measurement-only task. The documented unrelated `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT` remains untouched.
