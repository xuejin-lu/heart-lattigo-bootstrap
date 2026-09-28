# QPREFIX-PERF-DIAG-006 — Rescale Kernel Attribution

## Status

Executable diagnostic task.

## Task class

`D — Performance Diagnosis`

## Accepted parents

Low-overhead power attribution:
- `results/QPREFIX-PERF-DIAG-005-summary.md`
- Primary commit: `5a39f8ba560666f3c549a78cf4c9e33df89d2c72`
- decision: `LOW_OVERHEAD_RESCALE_DOMINANCE_CONFIRMED`

Current production Secondary ref:
`d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

Accepted current facts:
- diagnostics-off q0=55 P93 Count-1: about `78.5 ms`;
- power-only trace overhead: about `3.73%`;
- generated-power Rescale share: about `93.74%`;
- generated powers: about `60.53%` of combined EvalMod real+imag.

Therefore Rescale is the confirmed next investigation target.

## Goal

Identify which **production Rescale kernel(s)** dominate the remaining rows4 Q-prefix Rescale cost without using deep per-coefficient tracing.

This task must answer:

1. How much CPU time is spent in domain conversion (INTT/IMForm and NTT/MForm)?
2. How much is spent in fixed-width CRT reconstruction / centering / rounded division / capacity arithmetic?
3. How much is spent in signed residue staging / memory traffic?
4. How much is validation/orchestration?
5. What is the rows4 versus rows2 cost ratio under the same current implementation?
6. Which exact function or bounded group should be considered for the next optimization task?

No production optimization is allowed here.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`

Secondary:
- active branch `fast-qprefix`
- production arithmetic ref `d50ff4db757d4a2b9922937a4e7f316fd3f286b9` or current code-equivalent control-plane HEAD.

## 1. Hard constraints

Do not:

- modify Secondary production arithmetic;
- enable deep `rescale` trace for authoritative attribution;
- add per-coefficient timers;
- add one-off source overlays;
- change Q-prefix width/policy;
- change generated-power/P93 schedules;
- change parameters;
- introduce F or Standard/full-RNS fallback.

Test-only benchmark helpers are allowed.

## 2. Authoritative full-Rescale benchmark

Use the existing focused production benchmark:

`BenchmarkFastRescaleQPrefixRows4LogN13P93`

or its exact current equivalent.

Run diagnostics disabled with:

- setup/warmup outside timed region;
- enough iterations to obtain stable CPU-profile samples;
- `-benchmem`;
- at least 7 ordinary benchmark samples for latency distribution.

Report:

- raw ns/op or ms/op;
- median;
- B/op;
- allocs/op.

This is the authoritative full rows4 Rescale cost.

## 3. CPU profile

Collect a CPU profile from the focused rows4 Rescale benchmark only.

Use a sufficient benchmark duration (for example `-benchtime=5s` or longer if needed) so the profile has meaningful samples.

Produce:

- `go tool pprof -top`;
- cumulative and flat percentages;
- focused `-list` output for relevant Rescale functions when useful.

At minimum classify sampled CPU time into:

### A. Source-domain conversion
- INTT;
- IMForm;
- directly related ring-domain conversion helpers.

### B. Fixed-width exact arithmetic
- `crtQ01`, `crtQ012`, `crtQ0123`;
- `reconstructQPrefix`;
- `centeredQPrefix`;
- `roundedMagnitude192`;
- 192-bit multiply/divide/compare helpers used by the hot loop;
- per-step capacity arithmetic.

### C. Residue staging
- `signedResidue192`;
- target-row writes;
- immediately associated fixed-width modulus work.

### D. Commit domain restore
- NTT;
- MForm;
- directly related ring-domain restore helpers.

### E. Validation / orchestration / other
- shape validation;
- metadata/scale;
- scratch checks;
- all remaining production work.

Do not force a category when pprof cannot distinguish it; report unresolved samples explicitly.

## 4. Diagnostics-off phase microbenchmarks

CPU sampling may merge or inline small helpers, so add same-package **test/benchmark-only** microbenchmarks as necessary.

Required phase benchmarks for the same LogN13/P93 q0=55 rows4 regime:

1. `prefix_to_coefficient_rows4`
   - NTT input -> coefficient rows;
   - include IMForm exactly when production does.

2. `reconstruct_round_capacity_rows4`
   - coefficient rows already prepared outside timing;
   - execute the exact production fixed-width reconstruct/center/round/capacity logic over N coefficients;
   - do not include residue writes.

3. `residue_staging_rows4`
   - magnitudes/signs or equivalent valid fixed-width inputs prepared outside timing;
   - write exact target residues for production target row count.

4. `ntt_montgomery_restore_rows4`
   - staged coefficient rows prepared outside timing;
   - execute production NTT + MForm behavior.

5. `full_rescale_rows4`
   - same production API benchmark for reference.

Test-only helper code must call the same production primitives rather than invent mathematically different arithmetic.

Validate the phase fixture against one real production Rescale output so the benchmark inputs and row counts are representative.

## 5. Coarse phase closure

Use the phase benchmark medians to form an approximate decomposition:

[
T_{phase sum}
=
T_{prefix}
+
T_{fixedwidth}
+
T_{residue}
+
T_{restore}.
]

Compare it with full Rescale median.

Report:

[
closure_ratio = T_{phase sum}/T_{full}.
]

Because independent microbenchmarks differ in cache context, exact 100% closure is not required.

Interpretation:
- 0.80–1.20: adequate for phase ranking;
- outside that range: use pprof as primary evidence and explain the microbenchmark context mismatch.

Do not rescale numbers artificially to force closure.

## 6. Rows2 counterfactual measurement

Using the same current implementation and same parameter family, add a strictly test-only rows2 matched benchmark where the controlled input is valid under two-row authority.

Measure:

- full Rescale rows2;
- fixed-width reconstruction/rounding rows2;
- domain conversion rows2;
- commit restore rows2.

Report:

[
R_{4/2}=T_{rows4}/T_{rows2}
]

for full Rescale and each comparable phase.

This is a **kernel-width counterfactual only**.

Do not infer that production may safely use rows2 at the current P93 states. Q-prefix authority remains unchanged.

Label all rows2 comparisons:

`WIDTH_COUNTERFACTUAL_ONLY`.

## 7. Generated-power impact arithmetic

Using the accepted low-overhead DIAG-005 power-only data:

- combined generated-power Rescale (approx 34.5 ms) per Bootstrap;
- current generated-power parent (approx 36.8 ms).

Combine the focused full-Rescale improvement opportunity only arithmetically.

For each candidate phase, report the upper-bound Bootstrap impact if that phase were hypothetically reduced by:

- 25%;
- 50%.

This is sensitivity analysis only, not an implementation speedup prediction.

Do not assume all Rescale calls have identical phase distribution; state the limitation.

## 8. Classification

Return exactly one:

- `RESCALE_FIXED_WIDTH_ARITH_DOMINANT`
- `RESCALE_DOMAIN_TRANSFORM_DOMINANT`
- `RESCALE_RESIDUE_STAGING_DOMINANT`
- `RESCALE_MIXED_KERNELS`
- `RESCALE_KERNEL_ATTRIBUTION_UNCLOSED`

Use a dominant classification only if one category has at least 45% of attributable focused Rescale CPU/time and is corroborated by both pprof and phase microbenchmark evidence.

Use `MIXED` if no category reaches 45% but attribution is otherwise credible.

Use `UNCLOSED` if profiling/inlining/benchmark mismatch prevents a reliable ranking.

Also report:

- `TOP_RESCALE_KERNEL=<function-or-category>`
- `TOP_RESCALE_SHARE=<fraction>`
- `ROWS4_ROWS2_FULL_RATIO=<ratio>`
- `PHASE_CLOSURE_RATIO=<ratio>`

## 9. Candidate next optimization target

Name exactly one bounded **investigation/optimization candidate** based on the evidence.

Examples:
- specialized rows4 CRT reconstruction;
- roundedMagnitude192/division kernel;
- combined reconstruct+residue loop;
- domain-transform batching;
- NTT/MForm restore.

Do not implement it in this task.

The candidate must identify:
- exact function(s);
- why it is hot;
- expected semantic invariants that a later optimization must preserve.

## 10. Validation

Run:

- all new focused Secondary benchmark/test fixtures;
- relevant Rescale correctness tests;
- BigInt oracle tests;
- transactionality tests;
- `git diff --check`.

No Secondary production source commit is permitted.

If test-only benchmark code is added to Secondary, it may be committed only if it is clearly reusable diagnostics infrastructure/benchmark code and does not change normal package behavior. Otherwise keep it temporary and remove before completion.

Primary full `go test ./...` is not required for this measurement-only task; preserve known unrelated debt unchanged.

## 11. Required artifact

Write:

`results/QPREFIX-PERF-DIAG-006-summary.md`

Include:

- provenance;
- full rows4 Rescale raw benchmark data;
- CPU profile top table;
- phase microbenchmarks;
- closure;
- rows2 counterfactual;
- sensitivity arithmetic;
- classification;
- exact next bounded optimization candidate;
- uncertainty/limitations.

## 12. Completion

Return:

`RESCALE_KERNEL_ATTRIBUTION_READY`

plus:
- classification;
- `TOP_RESCALE_KERNEL=...`
- `TOP_RESCALE_SHARE=...`
- `ROWS4_ROWS2_FULL_RATIO=...`
- `PHASE_CLOSURE_RATIO=...`

Then report:

`READY_FOR_WEB_REVIEW`.
