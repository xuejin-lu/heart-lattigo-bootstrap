# QPREFIX-PERF-OPT-003 — Row-Major Source INTT Batching

## Status

Executable production optimization task.

## Task class

`P — Performance Repair`

## Accepted parent

Post-OPT002 reprofile:
- `results/QPREFIX-PERF-DIAG-007-summary.md`
- Primary commit: `7460f1cc4ff1ef6fcd3cfd26994ccaec7147b0bd`
- classification: `POST_OPT_RESCALE_MIXED`

Current production arithmetic:
- Secondary implementation base: `6930cf6cb3c71ce139a1eb42eede7be335b7174c`
- current control-plane descendant: `21ae64a3304287b1d96706acf2390973b04bb193`

Accepted current evidence:
- full rows4 Rescale median: about `1.130242 ms`;
- source `prefixToCoefficientRows` / INTT: `273.364 us` (22.21% of phase sum);
- commit NTT restore: `242.367 us` (19.69%);
- source+restore transforms: 41.89%;
- fresh pprof INTT stack: 23.39% cumulative;
- current source transform loop is component-major:
  - c0: q0,q1,q2,q3
  - c1: q0,q1,q2,q3.

The Lattigo `Ring.INTT` helper itself loops rows serially, so merely replacing the explicit row loop with `Ring.INTT` is not considered batching and is not the target.

## Goal

Improve source-domain conversion locality by changing the production preflight transform order from component-major to row-major across ciphertext components:

```text
current:
  c0 q0 q1 q2 q3
  c1 q0 q1 q2 q3

candidate:
  q0 c0 c1
  q1 c0 c1
  q2 c0 c1
  q3 c0 c1
```

This should allow each SubRing's INTT roots/twiddle data to be reused immediately across components.

No parallel goroutines, new NTT algorithm, assembly, or unsafe code are authorized.

## 1. Hard semantic constraints

Do not change:

- active Q-prefix rows;
- `QPrefixWidth(Level)`;
- INTT/IMForm mathematical operations;
- source ciphertext contents;
- centered CRT / rounding / capacity semantics;
- transactional failure-before-mutation;
- evaluator public APIs;
- output NTT/Montgomery representation;
- P93/generated-power schedule;
- F/full-RNS policy.

The candidate must produce bit-identical coefficient-domain residues to the current component-major source conversion.

## 2. Reuse existing evaluator-owned staging

Do not allocate a new per-Rescale coefficient tensor.

Use the already evaluator-owned:

`scratch.staged[component]`

as the coefficient-domain destination for each source component during preflight.

Required sequence:

1. validate all components/rows as today;
2. ensure all staged component polys exist;
3. transform the full source prefix for every component into its staged poly, using row-major order;
4. after all source-domain conversions succeed, run the exact coefficient reconstruction/round/capacity/staging logic for each component;
5. commit only after every component succeeds, preserving the existing transactionality contract.

In-place reuse of the staged polynomial for both coefficient-domain source and final target residues is allowed only if each coefficient's source residues are loaded into local scalars before any target-row write for that coefficient.

Prove this invariant in tests.

## 3. New bounded helper

Preferred shape:

`prefixToCoefficientComponents(...)`

or equivalently named helper.

Inputs should include:

- ringQ;
- source component polys;
- active source row count;
- NTT/Montgomery flags;
- destination staged component polys.

The helper must:

- validate backing/shape before mutation;
- loop rows outermost;
- loop components innermost;
- execute the same `SubRing.INTT` and conditional `IMForm` operations as the current source path;
- not touch inactive rows.

Do not modify generic Ring INTT implementation.

## 4. Restructure coefficient pass only as necessary

The current `rescaleQPrefixComponent` combines:

- source conversion;
- coefficient reconstruction;
- residue staging.

Refactor narrowly so source conversion can happen once for all components before coefficient reconstruction.

Do not rewrite the coefficient arithmetic.

It is acceptable to split into helpers such as:

- source conversion batch;
- staged coefficient-to-rescaled-residue pass.

Avoid unrelated cleanup.

## 5. Transactionality and alias safety

Add explicit tests for:

### In-place valid result
- `op0 == opOut`;
- bit-exact against the accepted pre-OPT003 implementation.

### Out-of-place valid result
- output bit-exact;
- input unchanged.

### Failure after batched conversion
Force a capacity failure after all source components have already been INTT-converted into evaluator-owned staging.

Assert:
- `op0` unchanged;
- `opOut` unchanged;
- metadata, Level, Scale, IsNTT, IsMontgomery unchanged.

### Higher-degree ciphertext
At least one degree >1 case to prove batching handles all staged components rather than assuming c0/c1 only.

## 6. Montgomery path

The accepted DIAG-007 profile was non-Montgomery, but production supports Montgomery input.

Add correctness tests proving the row-major helper performs:

`INTT -> IMForm`

for every active row/component exactly when `IsMontgomery=true`.

No performance win is required on the Montgomery fixture, but no regression in correctness is allowed.

## 7. Diagnostic-event semantics

Update fastdiag Rescale events to reflect the physical implementation.

The old component-scoped `prefix_to_coefficient` events may be replaced by one batched event under `preflight`, e.g.:

`prefix_to_coefficient_batch`

with fields indicating:
- source rows;
- component count;
- in-place state.

Do not fabricate per-component wall-clock spans around work that is actually row-major/interleaved.

Update tagged event-shape tests accordingly.

The normal `power` trace contract must remain unaffected.

## 8. Feasibility benchmark before accepting production change

Before finalizing, benchmark both exact strategies in a same-package test-only harness:

- current component-major source conversion;
- candidate row-major component-batched conversion.

Use:
- LogN13;
- P93 q0=55;
- rows4;
- degree-one c0/c1;
- diagnostics disabled;
- inputs prepared outside timing;
- >=7 samples.

Also benchmark rows2 as `WIDTH_COUNTERFACTUAL_ONLY`.

If row-major source conversion does not improve rows4 median by at least 3%, do not retain the production change. Return the no-win classification.

## 9. Authoritative performance

If the feasibility gate passes, rerun same-session before/after:

### A. source conversion phase
- rows4;
- >=7 samples;
- diagnostics off.

### B. full rows4 Rescale
`BenchmarkFastRescaleQPrefixRows4LogN13P93`
- >=7 samples;
- `-benchmem`.

### C. q0=55 P93 Count-1 E2E
`BenchmarkFastDiagP93Q55Count1`
- >=7 samples;
- `-benchmem`.

### D. low-overhead power trace
`./scripts/fastdiag trace --profile p93-q55 --trace power --warmup 1 --repetitions 7`

Do not use deep Rescale tracing for performance acceptance.

## 10. Performance gates

Return `SOURCE_INTT_BATCHING_READY` only if:

- source rows4 phase improves >= **5%**;
- full rows4 Rescale improves >= **2%**;
- diagnostics-off P93 E2E improves >= **1%**;
- no material allocation regression;
- all correctness/transactionality gates pass.

Return `SOURCE_INTT_BATCHING_CORRECT_BUT_NO_WIN` if correctness passes but a performance floor fails.

If the initial feasibility comparison is <3%, revert/omit the production source change and report the no-win result rather than committing churn.

Return `SOURCE_INTT_BATCHING_BLOCKED` if semantics cannot be preserved.

## 11. No concurrency in this task

Do not introduce:

- goroutines;
- worker pools;
- GOMAXPROCS-dependent scheduling;
- SIMD/assembly;
- unsafe;
- changes to ring NTT kernels.

If row-major locality is insufficient, concurrency/vectorization requires a separate task.

## 12. Validation

Run:

- new exact source-conversion equivalence tests;
- relevant Rescale oracle/transactionality tests;
- tagged fastdiag tests;
- `go test ./schemes/ckks/fast -count=1`;
- Secondary `go test ./...`;
- `git diff --check`.

Primary full suite remains optional for this production task; preserve known unrelated Primary debt.

## 13. Required artifact

Write:

`results/QPREFIX-PERF-OPT-003-summary.md`

Include:

- exact Secondary implementation commit, or explicit no-win/no-production-change state;
- source loop-order design;
- alias/transactionality proof;
- Montgomery and higher-degree tests;
- feasibility before/after source-phase samples;
- full Rescale samples;
- P93 E2E samples;
- allocations;
- low-overhead power trace;
- fastdiag event-shape update;
- remaining hotspot after change.

## 14. Completion

If successful, commit/push Secondary implementation normally.
Primary commits/pushes the result artifact.

Return exactly one:

- `SOURCE_INTT_BATCHING_READY`
- `SOURCE_INTT_BATCHING_CORRECT_BUT_NO_WIN`
- `SOURCE_INTT_BATCHING_BLOCKED`

Then report:

`READY_FOR_WEB_REVIEW`.
