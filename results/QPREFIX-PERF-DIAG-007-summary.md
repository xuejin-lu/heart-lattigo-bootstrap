# QPREFIX-PERF-DIAG-007 — Post-OPT002 Rescale Reprofile

**Classification:** `POST_OPT_RESCALE_MIXED`

**Required metrics** (phase shares use the sum of the four independently timed phase medians as denominator):

```text
CURRENT_TOP_FLAT_SYMBOL=math/bits.Div64
CURRENT_TOP_FLAT_SHARE=0.2055
CURRENT_FIXEDWIDTH_SHARE=0.3655
CURRENT_TRANSFORM_SHARE=0.4189
CURRENT_PHASE_CLOSURE=1.0892
```

## Provenance and scope

- Primary `main`: `1dbaae014552df09bd305052bb7ac8bdf1276862`.
- Secondary `fast-qprefix`: `21ae64a3304287b1d96706acf2390973b04bb193`; its sole change relative to OPT-002 commit `6930cf6cb3c71ce139a1eb42eede7be335b7174c` was `CURRENT_TASK.md`. Production source is unchanged from the accepted OPT-002 implementation.
- Go `go1.26.4`, `darwin/arm64`, Apple M4. P93 `LogN=13`, `q0=55`; full rows4, degree-1 ciphertext, NTT and non-Montgomery representation.
- No production source, Q-prefix policy, width, schedule, or Rescale transactionality was changed. A temporary same-package test-only phase/per-row benchmark was used and removed after measurement; Secondary is clean.

## Current full rows4 Rescale

Command:

```text
go test ./schemes/ckks/fast -run '^$' -bench '^BenchmarkFastRescaleQPrefixRows4LogN13P93$' -benchtime=100ms -benchmem -count=7
```

The existing benchmark performs one untimed evaluator warmup before `b.ResetTimer`. Seven raw samples (`ns/op`):

`1115320, 1130417, 1127798, 1139612, 1130242, 1126415, 1138981`

Median: **1,130,242 ns/op**; every sample reported **664 B/op, 18 allocs/op**.

## Fresh CPU profile

Focused rows4-only profile command (no fastdiag tracing):

```text
go test ./schemes/ckks/fast -run '^$' -bench '^BenchmarkFastRescaleQPrefixRows4LogN13P93$' -benchtime=8s -benchmem -cpuprofile=/tmp/qprefix-diag007.OqD1rx/rows4-cpu.pprof -o=/tmp/qprefix-diag007.OqD1rx/fast.test
```

The benchmark ran **8,612 iterations** at 1.124471 ms/op, with 670 B/op and 18 allocs/op. pprof recorded **9.49 s** of CPU samples over 10.09 s (94.06%); this exceeds the minimum profile duration and is limited to the focused Rescale benchmark.

### Flat top

| Symbol | Flat | Flat share | Cumulative share |
|---|---:|---:|---:|
| `math/bits.Div64` | 1.95 s | **20.55%** | 20.55% |
| `ring.nttUnrolled16Lazy` | 1.44 s | 15.17% | 18.02% |
| `ring.inttLazyUnrolled16` | 1.29 s | 13.59% | 20.34% |
| `(*Evaluator).rescaleQPrefixComponent` | 1.19 s | 12.54% | 76.92% |
| `ring.MRedLazy` (inline) | 0.88 s | 9.27% | 9.27% |
| `fast.mod192By64` | 0.55 s | 5.80% | 19.60% |
| `fast.crtQ0123Prepared` | 0.27 s | 2.85% | 14.01% |
| `ring.mulscalarmontgomeryvec` | 0.23 s | 2.42% | 3.06% |
| `fast.reconstructQPrefix` | 0.22 s | 2.32% | 16.33% |
| `fast.crtQ01` | 0.20 s | 2.11% | 2.74% |
| `fast.crtQ012Prepared` | 0.15 s | 1.58% | 6.53% |
| `fast.centeredQPrefix` | 0.12 s | 1.26% | 1.37% |
| `fast.roundedMagnitude192` | 0.08 s | 0.84% | 3.69% |

`math/bits.Div64` remains the largest single flat symbol. NTT and INTT kernels together contribute **28.76% flat**; their focused inclusive stack is **3.64 s / 38.36%**, including butterfly/reduction work (`MRedLazy`). Thus transforms are now comparable to the fixed-width coefficient path, and exceed it in the independently timed phase split, but neither group meets the 45% dominance threshold with profile corroboration.

### Cumulative and focused attribution

The cumulative top is nested and must not be summed: `rescaleQPrefixComponent` 76.92%; source `prefixToCoefficientRows` / INTT stack 23.39%; `mod192By64` 19.60%; `inttLazyUnrolled16` 20.34%; `nttUnrolled16Lazy` 18.02%; `signedResidue192` 16.97%; `reconstructQPrefix` 16.33%; `crtQ0123Prepared` 14.01%; `crtQ012Prepared` 6.53%; and `roundedMagnitude192` 3.69% cumulative.

pprof focused stacks measured **2.00 s / 21.07%** for the named reconstruction/center/round/prepared-CRT helper call paths, and **3.64 s / 38.36%** for `inttLazyUnrolled16|nttUnrolled16Lazy`. These are profile filters, not additive phase partitions. The coefficient-component function also carries loop, residue-load/write, and capacity-check work not represented by the selected helper-only filter.

No distinct flat symbols were material for `add192`/`mul192`/comparison helpers; they inline into prepared CRT/coefficient-loop callers. The focused source listing attributed 350 ms to residue reads and 350 ms to staged writes in `rescaleQPrefixComponent`; `runtime.memmove`/`memclr` did not appear as material flat nodes. The profile was non-Montgomery: `IMForm` and `MForm` branches were not exercised. `MRedLazy` samples arise in NTT butterfly kernels; they are included in the transform-stack focus.

## Refreshed phase split

Seven samples per phase, `-benchtime=100ms -benchmem -count=7`; fixtures and buffers were prepared before timing. Each phase processes both ciphertext components. The source transform is INTT over four active q rows with no `IMForm` for this non-Montgomery fixture; restore is NTT over four target rows with no `MForm`.

| Phase | Seven raw samples (`ns/op`) | Median (`µs/op`) | Share of phase sum |
|---|---|---:|---:|
| Source `prefixToCoefficientRows` / INTT rows4 | `287103, 273285, 274227, 280171, 271840, 271381, 273364` | 273.364 | 22.21% |
| Fixed-width reconstruct / center / round / capacity rows4 | `451524, 445024, 447502, 447999, 449917, 451511, 451051` | 449.917 | **36.55%** |
| Signed residue staging rows4 | `265612, 267169, 264801, 265418, 264820, 265058, 266284` | 265.418 | 21.56% |
| Commit-domain restore / NTT rows4 | `242091, 242367, 248232, 242079, 242367, 242501, 242225` | 242.367 | 19.69% |
| **Phase sum** | — | **1,231.066** | 100% |
| Full production API rows4 | `1115320, 1130417, 1127798, 1139612, 1130242, 1126415, 1138981` | **1,130.242** | — |

Closure is `1,231,066 / 1,130,242 = 1.0892`, inside the spec's 0.80–1.20 attribution band. Independent microbench contexts sum to 108.92% of full API time; phase shares above are normalized to the phase sum and are not claimed as disjoint fractions of full API latency. Source+restore transform phases total 515.731 µs, **41.89%** of phase sum; the fixed-width phase is **36.55%**. Relative to full API median, their raw ratios are 45.63% and 39.81%, respectively, but the closure caveat prevents using those ratios alone as dominance attribution.

### Per-row domain-transform measurements

Each sample transforms the corresponding q row for both ciphertext components; all reported 0 B/op and 0 allocs/op.

| Row | Source INTT median (`µs/op`) | Commit NTT median (`µs/op`) |
|---|---:|---:|
| q0 | 67.799 | 60.597 |
| q1 | 67.798 | 60.581 |
| q2 | 67.829 | 60.721 |
| q3 | 67.836 | 60.828 |

Per-row costs are nearly flat across q0–q3 for this profile; four-row totals closely match the corresponding aggregate phase measurements. These are ordinary non-Montgomery transforms, not a claim about an additional Montgomery conversion.

## Remaining fixed-width division evidence

Seven-sample median `ns/op`, 0 B/op and 0 allocs/op:

| Helper | Current median |
|---|---:|
| `mod128By64` | 2.665 |
| `mod192By64` | 6.078 |
| Prepared rows3 CRT | 12.88 |
| Prepared rows4 CRT | 34.31 |
| `roundedMagnitude192` source `/` + `%` | 6.460 |
| Exact `bits.Div64` candidate for rounded magnitude | 7.430 |
| Isolated one required lower-limb `bits.Div64` | 2.217 |
| Isolated two-limb `bits.Div64` chain | 5.125 |

The source `/`+`%` rounded helper remains faster than the exact `bits.Div64` candidate. The lower-limb chain is measurable but is a small isolated cost; these microbenchmarks do not establish an additional production speedup opportunity by themselves. No reciprocal, Barrett, unsafe, or assembly alternative was tested or implemented.

## `PRE_OPT002_ARCHIVED_REFERENCE`

DIAG-006 values are historical only: `math/bits.Div64` 27.40% flat; `ring.inttLazyUnrolled16` 16.33%; `ring.nttUnrolled16Lazy` 11.43%; fixed-width phase 43.51%. They predate OPT-002 and are not current attribution. In this fresh profile, Div64 fell to 20.55% flat, the NTT kernel rose to 15.17%, the INTT kernel to 13.59%, and the refreshed fixed-width phase is 36.55% of phase sum while source+restore transforms are 41.89%. The post-change shift is descriptive profile evidence, not causal speedup arithmetic.

## Classification and next bounded candidate

`POST_OPT_RESCALE_MIXED`: the fixed-width phase (36.55% of phase sum) and combined transform phases (41.89%) are both material but below 45%; fresh pprof corroborates both, and phase closure/profile duration are adequate. Div64 is still the top flat symbol, while transform stacks collectively are larger than any one arithmetic symbol.

Exactly one next candidate: **source-domain transform batching in `prefixToCoefficientRows` for q0–q3 across c0/c1**. Evidence: source INTT is the larger of the two transform phases at 273.364 µs/op (22.21% of phase sum), and pprof reports the INTT stack at 23.39% cumulative. Any later optimization must preserve the exact coefficient residues presented to CRT, transform only the active prefix, leave the input untouched, and apply `IMForm` exactly when the source is Montgomery-form. No batching or production change was made here.

Uncertainty: attribution applies to the specified q0=55, LogN13, rows4, NTT/non-Montgomery benchmark. It does not rank the Montgomery path, other widths/levels, or end-to-end Bootstrap costs. Flat/cumulative profile nodes overlap by design; phase microbench contexts differ from the full API context. A fresh profile of any later implementation is required before reclassifying the remaining hotspots.

## Validation and repository state

- `go test ./schemes/ckks/fast -run '^TestFastRescaleFixedWidthPhaseFixtureMatchesProductionAndStandard$' -count=1` — **PASS**.
- `go test ./schemes/ckks/fast -count=1` — **PASS**.
- All seven-sample benchmark groups — **PASS**; the fixed-width phase fixture had already been checked against actual Fast and Standard outputs.
- `git diff --check` — **PASS** in both repositories after removing the temporary Secondary benchmark helper.
- Secondary production ref: `21ae64a3304287b1d96706acf2390973b04bb193`; no Secondary production change, commit, or push. Secondary worktree clean.
- Primary worktree had only this result artifact pending.
