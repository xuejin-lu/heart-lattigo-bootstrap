# DIAG-FRAMEWORK-001-R1 — Rescale Event Closure Repair

## Status

Executable bounded repair task.

## Task class

`I — Diagnostic Infrastructure Repair`

## Parent implementation

Primary:
- implementation: `9662cd15d1a8c54e3ee618731f6f397fd45ca213`
- summary: `f832b8ed1788a7b59f29dde9919e5eee3bcf774c`

Secondary:
- implementation: `6312e8b9a982a405709a5124041eec569b0ae980`

Parent reported:
`REUSABLE_DIAGNOSTIC_FRAMEWORK_READY`

Independent Web review found one framework defect in Rescale subphase accounting.

## Defect

In `schemes/ckks/fast/rescale_qprefix.go`, each Rescale component currently emits:

1. a span named `reconstruct_center_round_capacity` using `Begin(...).End(...)`;
2. another event with the same name and same parent using `Record(..., reconstructNS)`.

Therefore the same logical reconstruction region appears twice as sibling events.

Primary aggregation intentionally treats repeated same-key events as separate occurrences (`#1`, `#2`). Parent closure therefore counts both, which can overstate child time and makes the deepest Rescale closure semantically ambiguous.

This is a diagnostic-schema defect only. It is not evidence of arithmetic failure and must not trigger any Fast-CKKS algorithm change.

## Required repair

Use an explicit nested hierarchy for the coefficient loop.

Preferred event tree:

```text
rescale
├── preflight
│   ├── prefix_to_coefficient [per component]
│   └── coefficient_loop      [per component]
│       └── reconstruct_center_round_capacity
└── materialization
    ├── prefix_to_coefficient [per component]
    ├── coefficient_loop      [per component]
    │   ├── reconstruct_center_round_capacity
    │   └── residue_materialization
    └── ntt_montgomery_restore [per component]
```

The exact internal helper names may differ, but the semantic rules are mandatory:

- one physical interval must have one parent-level accounting event;
- detailed child events may subdivide that interval;
- parent closure must never sum both a whole interval and its internal details as siblings;
- `reconstruct_center_round_capacity` must no longer be emitted twice for the same component/pass;
- negative or residual time must remain visible rather than being hidden.

The existing per-coefficient accumulated timers may remain if needed to separate reconstruction from residue materialization, but they must be parented beneath a single `coefficient_loop` span.

## Scope

Secondary changes allowed only in:

- `internal/fastdiag` if a tiny generic nesting helper is genuinely needed;
- Rescale diagnostic call sites;
- diagnostic-focused tests.

Primary changes allowed only in:

- aggregation/tests if required to validate nested closure;
- summary artifact.

Do not redesign the CLI or schema version unless unavoidable.

## Mandatory tests

### Secondary event-shape test

Extend the tagged Rescale trace validation so it asserts, for every traced Rescale:

- exactly one `preflight` child;
- exactly one `materialization` child;
- each pass contains the expected number of component-level `prefix_to_coefficient` and `coefficient_loop` events;
- each `coefficient_loop` has exactly one `reconstruct_center_round_capacity` child;
- materialization coefficient loops additionally have exactly one `residue_materialization` child;
- no duplicate same-semantic sibling event exists.

Do not merely test that required names are present.

### Primary closure test

Add a synthetic nested-event test proving:

- `rescale` closure uses only `preflight + materialization`;
- pass closure uses `prefix_to_coefficient + coefficient_loop (+ restore)`;
- coefficient-loop closure uses reconstruction and residue children;
- nested grandchildren are not also summed into the grandparent.

If the current generic aggregator incorrectly sums grandchildren or duplicate logical intervals, repair it narrowly.

## Re-run framework smoke tests

Run:

- normal Secondary focused tests;
- `-tags fastdiag` trace tests;
- Primary `go test ./cmd/fastdiag`;
- wrapper `trace --trace stage,power,rescale`.

Generate one sample trace and verify the Rescale tree is structurally unambiguous.

## Normal-build performance guard

This repair must not weaken the existing zero-impact model.

Confirm ordinary build still has:

- no diagnostic allocations;
- no timing calls in the disabled path;
- no reproducible >2% latency regression versus the accepted framework baseline.

A full new long benchmark campaign is unnecessary unless a regression appears.

## Diagnostic overhead

Because this repair changes only event hierarchy, record the enabled `stage,power,rescale` overhead once.

Do not attempt to optimize diagnostic-enabled overhead in this repair.

## Hard prohibitions

Do not:

- change Rescale arithmetic;
- remove/fuse preflight;
- change transactional failure-before-mutation behavior;
- change Q-prefix width or policy;
- change generated-power schedule;
- optimize production performance;
- reopen F/full-RNS fallback.

## Completion

Write:

`results/DIAG-FRAMEWORK-001-R1-summary.md`

Report:

- repaired event hierarchy;
- closure test results;
- example Rescale event tree;
- normal-build non-regression evidence;
- enabled overhead;
- exact Primary/Secondary commits.

Return one:

- `DIAGNOSTIC_FRAMEWORK_REPAIR_READY`
- `DIAGNOSTIC_FRAMEWORK_REPAIR_FAILED`

Then report:

`READY_FOR_WEB_REVIEW`.
