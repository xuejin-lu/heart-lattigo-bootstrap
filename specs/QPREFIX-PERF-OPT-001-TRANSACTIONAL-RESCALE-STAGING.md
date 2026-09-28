# QPREFIX-PERF-OPT-001 — Transactional Rescale Staging

## Status

Executable production optimization task.

## Task class

`P — Performance Repair`

## Accepted parents

Diagnostic framework:
- Primary R1 summary: `3e75a8db349f7c155db82a9ee9bc94afe6d7be98`
- Secondary R1: `4783c641cee2df5f504b8033905a62a82971da28`
- status: `DIAGNOSTIC_FRAMEWORK_REPAIR_READY`

Performance diagnosis:
- `results/QPREFIX-PERF-DIAG-003-summary.md`
- classification: `GENERATED_POWER_RESCALE_DOMINANT`
- generated-power Rescale delta share: `0.977205`
- candidate preflight time across twelve generated-power Rescales: about `9.267 ms`
- arithmetic-only counterfactual after removing repeated preflight work: about `26.866 -> 17.598 ms` for the real generated-power phase.

## Problem

Current Q-prefix Rescale preserves transactional failure-before-mutation by executing the expensive coefficient reconstruction path twice:

1. preflight:
   - prefix-to-coefficient conversion;
   - CRT reconstruction;
   - centered interpretation;
   - rounded division;
   - capacity validation;
   - then discards the computed result;
2. materialization:
   - repeats prefix-to-coefficient conversion;
   - repeats CRT reconstruction / rounding;
   - writes target residues;
   - NTT/Montgomery restore.

The first pass already proves whether the exact output is valid. Recomputing the exact result in the second pass is the measured dominant avoidable cost.

## Goal

Preserve the exact transactional contract while computing the exact rescaled value only once.

Preferred design:

```text
all input components
      |
      v
transactional staging pass
  prefix -> coeff
  CRT / center / round / capacity
  write final target residues into evaluator-owned staging
      |
      | only if every component succeeds
      v
commit pass
  NTT / MForm staged residues
  write opOut
  commit metadata / scale / resize
```

If any component fails validation, neither `op0` nor `opOut` may have observable mutation.

## 1. Hard semantic constraints

Do not change:

- centered CRT semantics;
- rounding rule;
- per-step capacity checks;
- divisor sequence;
- target-row contraction;
- q-prefix authority;
- `QPrefixWidth(Level)`;
- public Fast APIs;
- P93 schedule;
- polynomial schedule;
- failure-before-mutation behavior;
- NTT/Montgomery output representation;
- scale update semantics;
- F/full-RNS policy.

Do not specialize this task only to generated powers. Repair the shared Q-prefix Rescale implementation correctly.

## 2. Evaluator-owned reusable staging

Extend `fastRescaleScratch` with reusable coefficient-domain staging sufficient to hold the final target-prefix residues for every ciphertext component that must be committed atomically.

Requirements:

- staging is evaluator-owned, not allocated afresh for each coefficient or each Rescale call;
- after warmup, the production Rescale path must not introduce a new per-call allocation proportional to N or row count;
- support every ciphertext degree currently accepted by Fast Rescale;
- do not assume only component 0 or degree-one unless the existing public contract already enforces it;
- staging may grow on first use if necessary, but must then be reusable.

Use an implementation that is easy to audit for alias safety.

## 3. Single exact-computation pass

Refactor the current component routine so the expensive exact arithmetic executes once per component:

1. convert the authoritative source prefix to coefficient domain once;
2. reconstruct the centered integer once per coefficient;
3. apply all requested sequential rounded divisors;
4. perform the existing capacity check after every contraction;
5. write the final signed target-prefix residues into that component's staging buffer.

Do not write `opOut` during this pass.

Do not perform NTT/MForm during this pass unless the result is still entirely evaluator-owned and cannot affect caller-visible state.

## 4. Atomic commit

Only after all components stage successfully:

- resize `opOut` when needed;
- NTT target rows from staged coefficient residues;
- restore Montgomery form when required;
- write the output;
- update metadata and scale;
- perform final in-place resize.

For `op0 == opOut`, all source data must have been consumed into staging before any source row is overwritten.

For `op0 != opOut`, a failure before commit must leave the previous `opOut` contents and metadata unchanged.

There must be no second CRT reconstruction / rounded-division pass during commit.

## 5. Transactional tests

Add explicit tests that force a capacity failure after at least one earlier component has already staged successfully.

Test both:

### In-place

`op0 == opOut`

On error assert byte-for-byte / field-for-field preservation of:

- coefficient backing for every component/row;
- Level/storage shape;
- Scale;
- metadata/domain flags.

### Out-of-place

`op0 != opOut`

Initialize `opOut` with nontrivial sentinel contents and metadata.

On error assert both:

- `op0` unchanged;
- `opOut` unchanged.

The test must prove that staged work is not externally visible.

## 6. Correctness regression suite

At minimum rerun and pass the relevant existing tests covering:

- BigInt centered-CRT oracle at every prefix width;
- sequential `RescaleTo` contractions;
- in-place and out-of-place behavior;
- both ordinary and Montgomery NTT representations;
- higher-level Standard comparisons;
- dormant-residue independence;
- Q-prefix capacity failure behavior.

Add a direct equality test proving the optimized path is bit-exact against the pre-optimization algorithm for representative valid q01/q012/q0123/qprefix inputs when practical.

Do not weaken an existing test or tolerance.

## 7. Fastdiag event semantics

Update the reusable Rescale trace to reflect the physical algorithm, not the old two-computation implementation.

Keep the stable top-level concepts:

```text
rescale
├── preflight
└── materialization
```

After this repair:

- `preflight` means transactional compute-and-stage;
- `materialization` means commit of already-staged results.

Expected structure may become:

```text
rescale
├── preflight
│   ├── prefix_to_coefficient
│   └── coefficient_loop
│       ├── reconstruct_center_round_capacity
│       └── residue_materialization
└── materialization
    └── ntt_montgomery_restore
```

per component.

Do not emit fictional second reconstruction events merely for compatibility.

Update the tagged event-shape tests accordingly.

## 8. Use the reusable framework

This is the first production repair that must use `scripts/fastdiag` rather than ad-hoc overlays for post-change attribution.

Use the pre-change Secondary ref:

`4783c641cee2df5f504b8033905a62a82971da28`

and the final candidate ref with:

```sh
./scripts/fastdiag compare \
  --baseline 4783c641cee2df5f504b8033905a62a82971da28 \
  --candidate <candidate> \
  --profile p93-q55 \
  --trace stage,power,rescale
```

Because all-scope diagnostics have high overhead, use this result for structural attribution and closure, not as the authoritative production-speed number.

Do not build a task-specific overlay.

## 9. Authoritative production performance measurement

Performance acceptance must come from diagnostics-OFF ordinary builds.

Before modifying production Rescale, record a same-session baseline at the current Secondary HEAD.

After the implementation, rerun the same commands/session policy.

### A. q0=55 P93 Count-1 E2E

Use the existing ordinary-build P93 q0=55 benchmark with:

- warmup;
- at least 5 measured samples;
- `-benchmem`;
- sequential baseline then candidate measurement where practical.

Report raw samples and medians.

### B. Focused Q-prefix Rescale benchmark

Add or use a benchmark for representative q0=55, rows4, high-level Rescale corresponding to the generated-power path.

The benchmark must avoid timing input fixture construction inside the measured region.

Report:

- ns/op;
- B/op;
- allocs/op;
- in-place and/or out-of-place mode clearly.

## 10. Performance gates

Primary success gate:

- ordinary q0=55 P93 Count-1 E2E median improves by at least **10%** versus the same-session pre-change baseline.

Focused Rescale gate:

- representative rows4 Rescale median improves by at least **20%**.

Allocation gate:

- no material per-operation allocation regression;
- after warmup, staging must be reused;
- report any B/op or allocs/op change explicitly.

These are acceptance floors, not optimization targets.

If correctness passes but either performance floor is missed, do not hide it. Return the no-win classification below.

## 11. Expected but non-binding interpretation

Based on PERF-DIAG-003, removing repeated preflight computation should reclaim roughly the measured preflight work, but the exact speedup is not guaranteed.

Do not present the old arithmetic counterfactual as a guaranteed result.

This task does not attempt to optimize:

- the remaining q0123/rows4 CRT kernel;
- 192-bit arithmetic;
- NTT;
- DFT;
- PS schedule;
- DoubleAngle;
- row width.

Those can be evaluated separately after this repair.

## 12. Full validation

Run:

### Secondary

- focused Rescale / polynomial / bootstrap tests;
- tagged fastdiag tests;
- `go test ./...` if no repository-wide unrelated blocker exists.

### Primary

- `go test ./cmd/fastdiag`;
- wrapper trace/compare smoke tests;
- `go test ./...` and preserve the already-documented
  `TestFIX001P3GenuineStandardPublicVsStagedConsistency` /
  `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`
  as unrelated debt if it remains the only failure.

Run `git diff --check`.

## 13. Required result artifact

Write:

`results/QPREFIX-PERF-OPT-001-summary.md`

Include:

- exact before/after commits;
- design description;
- proof of transactional staging;
- tests;
- ordinary E2E before/after raw samples and medians;
- focused Rescale before/after;
- B/op / allocs/op;
- fastdiag compare summary;
- whether the old repeated CRT/rounding pass is absent;
- remaining performance bottleneck after the repair.

## 14. Completion classification

Return exactly one:

- `TRANSACTIONAL_RESCALE_STAGING_READY`
- `TRANSACTIONAL_RESCALE_STAGING_CORRECT_BUT_NO_WIN`
- `TRANSACTIONAL_RESCALE_STAGING_BLOCKED`

Use `READY` only if correctness/transactionality gates pass and both performance floors pass.

Commit and normal fast-forward push the Secondary implementation and Primary result under the standing workflow.

Then report:

`READY_FOR_WEB_REVIEW`.
