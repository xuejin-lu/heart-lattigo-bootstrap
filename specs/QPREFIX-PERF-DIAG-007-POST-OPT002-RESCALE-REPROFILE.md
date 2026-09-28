# QPREFIX-PERF-DIAG-007 — Post-OPT002 Rescale Reprofile

## Status

Executable diagnostic task.

## Task class

`D — Performance Diagnosis`

## Accepted parent

Production optimization:
- Primary summary: `results/QPREFIX-PERF-OPT-002-summary.md`
- Primary commit: `5efdcc17d9daf42400ac970a4c7c1a10dc87ac58`
- Secondary implementation: `6930cf6cb3c71ce139a1eb42eede7be335b7174c`
- classification: `FIXED_WIDTH_DIVISION_DEDUP_READY`

Accepted production effect:
- rows4 fixed-width phase: `532.672 -> 467.548 us` (12.23% faster);
- full rows4 Rescale: `1.206153 -> 1.124711 ms` (6.75% faster);
- q0=55 P93 Count-1: `79.169688 -> 74.330000 ms` (6.11% faster);
- allocation count stable.

The DIAG-006 CPU-profile percentages predate this optimization and must not be reused as current attribution.

## Goal

Refresh the current rows4 Rescale CPU profile after fixed-width division deduplication and determine the next bounded optimization target.

This task must answer:

1. Is `math/bits.Div64` still the largest flat CPU hotspot?
2. Have INTT/NTT now overtaken fixed-width arithmetic?
3. Which exact current function/group should be optimized next?
4. Is the remaining Rescale cost still arithmetic-dominated, transform-dominated, or mixed?

No production optimization is allowed.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`

Secondary:
- active branch `fast-qprefix`
- production ref `6930cf6cb3c71ce139a1eb42eede7be335b7174c` or current control-plane-only descendant.

## 1. Hard constraints

Do not:

- modify Secondary production arithmetic;
- add per-coefficient timers;
- add ad-hoc source overlays;
- change Q-prefix width/policy;
- change Rescale transactionality;
- change P93/generated-power schedule;
- introduce reciprocal approximations, unsafe, assembly, or F/full-RNS fallback.

Reusable benchmark/test-only helpers may be used.

## 2. Fresh authoritative full-Rescale benchmark

Rerun diagnostics-off:

`BenchmarkFastRescaleQPrefixRows4LogN13P93`

with:

- one warmup where applicable;
- at least 7 measured samples;
- `-benchmem`;
- same P93 q0=55 regime.

Report raw samples, median, B/op, allocs/op.

This establishes the current absolute full-Rescale baseline.

## 3. Fresh CPU profile

Collect a new CPU profile from the focused rows4 Rescale benchmark only.

Use enough benchmark duration for stable sampling (minimum 5s; extend if sample count is weak).

Report:

- pprof flat top;
- pprof cumulative top;
- focused `-list` views where useful.

At minimum identify current samples attributable to:

### Fixed-width arithmetic
- `math/bits.Div64`;
- `mod128By64`;
- `mod192By64`;
- `crtQ01`;
- `crtQ012Prepared`;
- `crtQ0123Prepared`;
- `roundedMagnitude192`;
- `signedResidue192`;
- helper multiply/add/compare routines.

### Source-domain conversion
- INTT kernels;
- IMForm if present.

### Commit-domain restore
- NTT kernels;
- MForm if present.

### Memory/copy
- row copies;
- staged residue writes;
- runtime memmove/memclr if materially sampled.

### Other
- orchestration/validation and unresolved samples.

Do not force an inlined sample into a named helper if pprof cannot support that attribution.

## 4. Post-opt phase refresh

Rerun the reusable DIAG-006 phase benchmarks on the new production commit:

- prefix-to-coefficient rows4;
- reconstruct/round/capacity rows4;
- residue staging rows4;
- NTT/Montgomery restore rows4;
- full Rescale rows4.

At least 7 samples per phase.

Report medians and recompute:

[
closure_ratio = rac{T_{prefix}+T_{fixedwidth}+T_{residue}+T_{restore}}{T_{full}}.
]

Also report each phase share of the phase sum.

Do not compare current shares to DIAG-006 without labeling those old numbers pre-OPT002.

## 5. Remaining-division focused evidence

Rerun helper benchmarks for:

- `mod128By64`;
- `mod192By64`;
- `roundedMagnitude192`;
- rows3/rows4 prepared CRT reconstruction.

Additionally benchmark the remaining required lower-limb `bits.Div64` chains in isolation if a reusable helper already permits it or a test-only benchmark can do so without changing production code.

The purpose is to estimate whether a further exact division optimization has enough headroom to justify a dedicated task.

Do not implement reciprocal or Barrett reduction here.

## 6. Domain-transform focused evidence

Benchmark the production-equivalent rows4:

- source INTT(+IMForm as applicable);
- commit NTT(+MForm as applicable).

Use fixtures prepared outside timing.

Where feasible, also report per-row transform cost to determine whether rows4 cost is close to linear in active rows.

This is measurement only; do not change transform scheduling or authority.

## 7. Classification

Return exactly one:

- `POST_OPT_RESCALE_DIVISION_DOMINANT`
- `POST_OPT_RESCALE_TRANSFORM_DOMINANT`
- `POST_OPT_RESCALE_MIXED`
- `POST_OPT_RESCALE_ATTRIBUTION_UNCLOSED`

Use `DIVISION_DOMINANT` only if fixed-width division/reconstruction evidence accounts for at least 45% of attributable current full-Rescale cost and pprof corroborates it.

Use `TRANSFORM_DOMINANT` only if source+commit domain transforms together account for at least 45% and pprof corroborates it.

Use `MIXED` if neither reaches 45% but attribution is credible.

Use `UNCLOSED` if phase closure or profile quality is insufficient.

Also report:

- `CURRENT_TOP_FLAT_SYMBOL=<symbol>`
- `CURRENT_TOP_FLAT_SHARE=<fraction>`
- `CURRENT_FIXEDWIDTH_SHARE=<fraction>`
- `CURRENT_TRANSFORM_SHARE=<fraction>`
- `CURRENT_PHASE_CLOSURE=<ratio>`

## 8. Next bounded candidate

Name exactly one next optimization candidate based on fresh evidence.

The candidate must be one of these bounded families:

- remaining exact division/mod-reduction kernel;
- CRT reconstruction composition;
- signed residue reduction/staging;
- source-domain transform batching;
- commit-domain transform batching;
- another precisely identified current hotspot.

Do not write an implementation plan beyond naming:
- exact function(s)/phase;
- measured evidence;
- semantic invariants a later optimization must preserve.

## 9. Historical comparison

Include DIAG-006 pre-OPT002 profile values only as:

`PRE_OPT002_ARCHIVED_REFERENCE`

At minimum note:
- bits.Div64 27.40% flat;
- INTT 16.33%;
- NTT 11.43%;
- fixed-width phase 43.51%.

Then state how the fresh profile shifted.

Do not compute speedup causality from profile percentages alone.

## 10. Validation

Run:

- relevant focused benchmark/test fixtures;
- `go test ./schemes/ckks/fast -count=1`;
- `git diff --check`.

No Secondary production commit is allowed.

Primary full `go test ./...` is not required for this measurement-only task; preserve unrelated debt unchanged.

## 11. Required artifact

Write:

`results/QPREFIX-PERF-DIAG-007-summary.md`

Include:

- provenance;
- current full Rescale raw samples;
- fresh pprof top/cumulative evidence;
- refreshed phase table and closure;
- remaining division helper evidence;
- transform evidence;
- archived pre-OPT002 comparison;
- classification;
- exact next bounded optimization candidate;
- uncertainties.

## 12. Completion

Return:

`POST_OPT002_RESCALE_REPROFILE_READY`

plus:
- classification;
- `CURRENT_TOP_FLAT_SYMBOL=...`
- `CURRENT_TOP_FLAT_SHARE=...`
- `CURRENT_FIXEDWIDTH_SHARE=...`
- `CURRENT_TRANSFORM_SHARE=...`
- `CURRENT_PHASE_CLOSURE=...`

Then report:

`READY_FOR_WEB_REVIEW`.
