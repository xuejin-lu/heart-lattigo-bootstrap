# DIAG-FRAMEWORK-001 — Reusable Fast-CKKS Diagnostic Trace Framework

## Status

Executable implementation task.

## Task class

`I — Diagnostic Infrastructure`

## Motivation

QPREFIX-PERF-DIAG-001/002/003 repeatedly used temporary overlays and disposable timing hooks to answer:

- where an end-to-end regression lives;
- which generated-power subphase dominates;
- which Rescale internal pass dominates.

The experiments succeeded, but the instrumentation was removed each time.

This task converts only the proven recurring needs into a reusable opt-in framework so future debugging does not require rebuilding one-off timing code.

This task is infrastructure, not a performance repair.

## Accepted evidence motivating the framework

Accepted result:
`results/QPREFIX-PERF-DIAG-003-summary.md`

Key facts:

- matched E2E stage attribution closed to 0.3343% residual;
- EvalMod real/imag account for most of the regression;
- generated-power Rescale explains 97.7205% of the generated-power delta;
- candidate Rescale preflight alone occupies 9.267292 ms across the twelve generated-power Rescales;
- the existing one-off overlay methodology proved useful at three nested depths.

## Repositories

Primary:
`xuejin-lu/heart-lattigo-bootstrap@main`

Secondary:
`xuejin-lu/lattigo@fast-qprefix`

Both repositories may be modified by this task within the bounded infrastructure scope below.

## Goal

Provide a reusable diagnostic workflow with a simple user-facing switch:

```sh
./scripts/fastdiag trace --trace stage
./scripts/fastdiag trace --trace stage,power,rescale
./scripts/fastdiag compare --baseline <sha> --candidate <sha> --trace stage,power,rescale
```

Exact CLI spelling may differ slightly if required by shell/Go conventions, but the final interface must remain this simple in spirit.

The framework must support the three trace depths already proven useful:

1. `stage`
2. `power`
3. `rescale`

Do not generalize beyond demonstrated requirements in this task.

## Architecture

### A. Production-default behavior

Normal Fast-CKKS builds and tests must remain diagnostics-off.

Diagnostic instrumentation must be compile-time gated with a dedicated Go build tag:

`fastdiag`

The normal build must compile diagnostic branches away.

Preferred pattern:

```go
// normal build
const Enabled = false

// fastdiag build
const Enabled = true
```

Production call sites may contain:

```go
if fastdiag.Enabled {
    ...
}
```

but diagnostic-disabled builds must not:

- call `time.Now`;
- allocate diagnostic records;
- acquire diagnostic locks;
- read diagnostic environment variables in hot paths;
- serialize output;
- alter arithmetic or metadata semantics.

If compiler inspection or benchmark evidence shows the disabled branch is not eliminated, stop and redesign before completion.

### B. Shared trace support

Implement a small internal Secondary package, preferably:

`internal/fastdiag`

or another cycle-safe internal location.

It must provide:

- compile-time `Enabled`;
- trace-scope filtering;
- reset/start/end or span API;
- in-memory event collection;
- stable event records;
- deterministic JSON export or equivalent structured output.

Do not add a third-party tracing dependency.

### C. Trace scopes

Support exactly these top-level selectable scopes:

- `stage`
- `power`
- `rescale`

`all` may be an alias for all three.

Unknown scope names must fail clearly.

Trace selection may use `FASTDIAG_TRACE` internally, but the user-facing wrapper must set it automatically.

## Event schema

Use one reusable structured event schema rather than task-specific structs.

Each event must contain where applicable:

- `scope`
- `name`
- `elapsed_ns`
- `sequence`
- `level_in`
- `level_out`
- `rows_in`
- `rows_out`
- `scale_log2_in`
- `scale_log2_out`
- `in_place`
- `power`
- `component`
- `count` or another compact repetition indicator if useful
- optional parent/span identifier if needed for non-overlapping closure.

Do not serialize ciphertext coefficients, secret material, keys, or full polynomial arrays.

## Required instrumentation

### 1. Stage trace

Instrument the actual Fast public Bootstrap execution path, not a separately reconstructed stage pipeline.

Record non-overlapping spans compatible with the accepted QPREFIX-PERF-DIAG-002 decomposition:

- pack / N1→N2
- ScaleDown
- ModUp including production Trace
- C2S
- EvalMod real
- EvalMod imaginary
- S2C
- N2→N1 / unpack
- public finalization
- full Bootstrap parent span

The stage trace must permit automatic stage-sum closure against the parent span.

### 2. Power trace

Instrument Fast Chebyshev generated-power construction.

For each runtime generated power record:

- power n;
- split if practical;
- source/output levels;
- authoritative rows;
- total power span.

Also record the non-overlapping categories proven in PERF-DIAG-003:

- copy/workspace;
- balanced integer scaling;
- Relinearize;
- Mul/MulRelin;
- Rescale;
- Chebyshev doubling;
- recurrence correction/aligned subtraction;
- residual derivable by parent minus children.

Do not change the power DAG or planner.

### 3. Rescale trace

Instrument Q-prefix Rescale with:

- whole Rescale call;
- preflight pass;
- materialization pass.

Within those passes record, where the existing source boundary permits without duplicating work:

- prefix-to-coefficient conversion;
- reconstruction / centering / rounded division / capacity validation;
- residue materialization;
- NTT/Montgomery restore.

Record source/target Level, rows, component, and in-place status.

Do not alter transactional behavior.

## User-facing control plane

Add a reusable Primary wrapper:

`scripts/fastdiag`

It must have at least:

### Trace current checkout

```sh
./scripts/fastdiag trace --trace stage,power,rescale
```

Behavior:

- verify Primary and Secondary paths;
- refuse unsafe dirty-state operations if it needs temporary worktrees;
- run the diagnostic-enabled harness;
- perform warmup;
- collect a configured number of repetitions;
- write structured JSON under a requested output path or a safe temporary path;
- print a concise human-readable summary.

### Compare two Secondary refs

```sh
./scripts/fastdiag compare \
  --baseline <sha-or-ref> \
  --candidate <sha-or-ref> \
  --trace stage,power,rescale
```

Behavior:

- create isolated detached temporary worktrees without moving the authoritative Secondary checkout;
- use isolated `GOCACHE`;
- run sequentially, never concurrently;
- use the same diagnostic workload/configuration;
- collect raw runs;
- remove task-created temporary worktrees/caches on normal completion;
- preserve them or report paths on unexpected failure when needed for debugging;
- produce one comparison JSON and one concise Markdown summary.

No force/reset/stash/destructive Git operation is allowed.

## Comparison analysis

For matching event keys, automatically report:

- baseline median;
- candidate median;
- ratio;
- delta;
- contribution to parent delta when a parent exists.

For `stage` calculate:

[
R=Delta T_{E2E}-sum_iDelta T_i
]

and residual fraction.

For `power` calculate generated-power parent/child closure.

For `rescale` report preflight/materialization shares.

The tool must preserve negative deltas.

Do not automatically declare a causal conclusion such as "bug" or "safe optimization". It reports evidence only.

## Workload/profile

For this first implementation support the already-proven matched diagnostic profile:

`p93-q55`

It must reproduce the accepted QPREFIX-PERF-DIAG-002/003 workload fingerprint:

- LogN=13;
- LogSlots=12;
- q0=55 profile;
- degree-30 Chebyshev / P93 path;
- DoubleAngle=3;
- deterministic 4096-value input used by those tasks.

Expose it as:

`--profile p93-q55`

Do not add a generic profile DSL in this task.

The framework should be structurally extendable by adding another named profile later.

## Historical-ref compatibility boundary

The framework is primarily forward-reusable.

For refs that predate the diagnostic hook call sites, the compare command may report:

`DIAGNOSTIC_HOOKS_UNAVAILABLE_AT_REF`

rather than dynamically patch arbitrary old source.

Do not reintroduce ad-hoc overlay generation merely to support every historical commit.

The accepted old q0=55 measurements remain in PERF-DIAG-002/003.

## Output schema

Use versioned schemas, e.g.:

- `fastdiag.trace.v1`
- `fastdiag.compare.v1`

Include:

- timestamp;
- Primary commit/ref/dirty state;
- Secondary commit/ref/dirty state;
- Go version;
- OS/arch/CPU/GOMAXPROCS;
- profile;
- selected trace scopes;
- warmup/repetitions;
- workload fingerprint;
- raw events/runs;
- computed medians/deltas/closure.

Keep raw artifacts bounded and human-reviewable.

## Tests

### Secondary

Add focused tests for:

- normal build compiles with diagnostics disabled;
- `fastdiag` build compiles;
- scope parser;
- event nesting/closure;
- stage events emitted in expected order;
- power events for `{2,3,4,6,8,16}` under p93-q55;
- Rescale preflight/materialization nesting;
- trace collection does not change output metadata or numerical result relative to diagnostics-off within existing tolerances.

### Primary

Add tests for:

- CLI argument validation;
- compare aggregation math;
- median/delta/ratio;
- negative deltas;
- closure calculation;
- dirty authoritative Secondary refusal;
- unknown trace scope/profile;
- cleanup behavior using a test-safe temporary repository fixture where practical.

## Performance non-regression gate

This is mandatory.

On the ordinary build without `-tags fastdiag`, rerun the current candidate q0=55 Count-1 Bootstrap benchmark.

Compare against a same-session pre-task measurement.

Diagnostics-off median latency must not regress by more than 2%.

If the observed difference is >2%, repeat enough runs to distinguish noise. If a reproducible >2% regression remains, the task fails and the instrumentation architecture must be revised.

Also inspect allocation metrics; diagnostics-off must add zero persistent per-operation diagnostic allocations.

## Diagnostic-enabled overhead

Measure the same workload with:

- diagnostics off;
- `--trace stage`;
- `--trace stage,power,rescale`.

Report overhead.

No hard performance threshold is imposed for the enabled mode in this first task, but the overhead must be measured and documented so future timing interpretation is calibrated.

## Documentation

Update Primary README or a dedicated:

`docs/FAST_DIAGNOSTICS.md`

with:

- purpose;
- normal-build zero-impact model;
- three trace scopes;
- example trace command;
- example compare command;
- output locations/schema;
- interpretation boundaries.

Do not document task-specific PERF-DIAG history as part of the API.

## Prohibitions

Do not:

- optimize Rescale;
- remove/fuse preflight;
- change Q-prefix arithmetic;
- change `QPrefixWidth(Level)`;
- change P93 schedule;
- reopen F/full-RNS fallback;
- add application-facing Fast/Normal runtime selectors;
- move Fast arithmetic semantics into Primary;
- retain temporary task-specific overlay files;
- add a general tracing dependency.

## Completion artifacts

Primary:

- reusable control plane / aggregation code;
- tests;
- docs;
- `results/DIAG-FRAMEWORK-001-summary.md`.

Secondary:

- compile-time-gated reusable trace support;
- bounded trace call sites;
- focused tests.

The summary must report:

- exact commits;
- diagnostics-off performance before/after;
- enabled overhead by trace scope;
- example trace output;
- example compare output;
- tests;
- any unsupported historical-ref behavior;
- confirmation that arithmetic/transactional semantics were unchanged.

## Acceptance result

Return one:

- `REUSABLE_DIAGNOSTIC_FRAMEWORK_READY`
- `DIAGNOSTIC_FRAMEWORK_NEEDS_REPAIR`
- `DIAGNOSTIC_FRAMEWORK_BLOCKED`

Then report:

`READY_FOR_WEB_REVIEW`.
