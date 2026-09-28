# QPREFIX-PERF-DIAG-005 — Low-Overhead Power Attribution

## Status

Executable diagnostic task.

## Task class

`D — Performance Diagnosis`

## Parent

- `results/QPREFIX-PERF-DIAG-004-summary.md`
- Primary result commit: `672ce1b578db1147ba9bb4cb9834ffe357a50af3`
- Post-opt Secondary production ref: `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

DIAG-004 established a trustworthy stage-only picture:

- diagnostics-off Count-1 median: about `78.994 ms`;
- stage-only trace parent: about `80.279 ms` (+1.63%);
- EvalMod real: about `30.064 ms`;
- EvalMod imaginary: about `29.867 ms`;
- C2S: about `10.717 ms`;
- S2C: about `9.134 ms`.

However, its `CURRENT_RESCALE_STILL_DOMINANT` classification relied on `stage,power,rescale` all-scope tracing, whose Bootstrap parent rose to about `145.225 ms` (+~81% versus stage-only / diagnostics-off).

Source review confirms that enabling the `rescale` scope adds per-coefficient timers inside `rescaleQPrefixComponent`. Those timers execute inside the outer `scope=power,name=rescale` span and therefore inflate the measured power-category Rescale duration.

This task determines the current generated-power category distribution **without enabling the deep Rescale scope**.

## Goal

Measure current generated-power timing with low instrumentation perturbation and decide whether Rescale is still the dominant production-side generated-power category after QPREFIX-PERF-OPT-001.

No production optimization is allowed.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`

Secondary:
- active branch `fast-qprefix`
- production ref to measure: `d50ff4db757d4a2b9922937a4e7f316fd3f286b9` or current code-equivalent control-plane HEAD.

No Secondary production edits.

## 1. Hard constraints

Do not:

- enable `rescale` scope for the authoritative power-category measurement;
- add one-off overlays;
- modify production source;
- change Q-prefix policy, parameters, P93 schedule, polynomial schedule, or generated-power DAG;
- use the DIAG-004 all-scope `96.51%` number as an input assumption.

Use only the permanent reusable fastdiag framework.

## 2. Authoritative diagnostics-off baseline

Rerun current q0=55 P93 Count-1 diagnostics-off benchmark:

- one warmup;
- at least 7 measured samples;
- `-benchmem`.

Report median and raw samples.

This is the production-speed reference.

## 3. Power-only trace

Run:

```sh
./scripts/fastdiag trace \
  --profile p93-q55 \
  --trace power \
  --warmup 1 \
  --repetitions 7
```

The `rescale` scope MUST be disabled.

Report:

- Bootstrap wall time recorded by the harness;
- both `generated_powers` parent spans;
- direct child categories:
  - copy/workspace;
  - integer scaling;
  - relinearize;
  - mul;
  - mul_relin;
  - rescale;
  - Chebyshev doubling;
  - recurrence correction;
  - residual;
- each generated power `{2,3,4,6,8,16}`.

For every run verify:

[
T_{generated powers}
approx
sum T_{direct categories}+residual
]

without summing nested per-power parents twice.

Compute:

[
S_R = rac{T_{power/rescale}}{T_{generated powers}}.
]

Return:

`POWER_ONLY_RESCALE_SHARE=<fraction>`

## 4. Stage + power trace

Run:

```sh
./scripts/fastdiag trace \
  --profile p93-q55 \
  --trace stage,power \
  --warmup 1 \
  --repetitions 7
```

Again, `rescale` scope MUST be disabled.

Purpose:

- place generated powers inside the current EvalMod stage context;
- measure instrumentation interaction between stage and power scopes;
- verify whether generated powers remain the largest sub-contributor to EvalMod under low perturbation.

Report:

- Bootstrap parent;
- EvalMod real/imag;
- generated-power real/imag parent spans;
- power-category Rescale spans;
- closure.

## 5. Instrumentation-overhead gate

Calculate overhead relative to diagnostics-off E2E for:

- power-only;
- stage+power.

Use median wall times from the harness.

Interpretation:

- if power-only overhead <= 10%, it may be used for current category ranking;
- if power-only overhead > 10% but <= 20%, ranking is provisional and must be corroborated by focused diagnostics-off benchmarks;
- if power-only overhead > 20%, return attribution unclosed.

Do not use all-scope `stage,power,rescale` for this gate.

## 6. Focused diagnostics-off corroboration

Use or add **test/benchmark-only** focused benchmarks if needed, without modifying production arithmetic, to measure the actual production Rescale calls used by generated powers.

At minimum compare, diagnostics disabled:

- one representative rows4 in-place Rescale at the relevant high Level;
- one representative generated-power MulRelin;
- one representative integer-scale operation.

The fixture must be outside the timed region and match the p93-q55 row/Level regime.

This section is mandatory if power-only overhead exceeds 10%; optional otherwise.

Do not add deep per-coefficient timers.

## 7. Decision

Return exactly one:

- `LOW_OVERHEAD_RESCALE_DOMINANCE_CONFIRMED`
- `LOW_OVERHEAD_RESCALE_DOMINANCE_REJECTED`
- `LOW_OVERHEAD_POWER_ATTRIBUTION_UNCLOSED`

Use `CONFIRMED` only if:

1. power-only overhead is <=10% (or focused diagnostics-off corroboration clearly agrees), and
2. current Rescale accounts for >=50% of current generated-power time, and
3. generated powers remain a major contributor inside current EvalMod.

Use `REJECTED` if the low-overhead measurement shows the old all-scope result materially overstated Rescale and another category or mixed set now dominates.

Use `UNCLOSED` if instrumentation perturbation prevents a reliable decision.

Additionally report:

- `POWER_ONLY_OVERHEAD=<fraction>`
- `POWER_ONLY_RESCALE_SHARE=<fraction>`
- `GENERATED_POWER_SHARE_OF_EVALMOD=<fraction>`

## 8. DIAG-004 reinterpretation

Explicitly state whether the DIAG-004 `CURRENT_RESCALE_STILL_DOMINANT` classification is:

- confirmed;
- superseded;
- or unresolved.

Do not alter the historical DIAG-004 report.

The all-scope `96.51%` number must be labeled:

`DEEP_TRACE_PERTURBED_REFERENCE_ONLY`

in this new report.

## 9. Required artifact

Write:

`results/QPREFIX-PERF-DIAG-005-summary.md`

Include:

- diagnostics-off raw samples;
- power-only raw/median table;
- stage+power raw/median table;
- category closure;
- overhead calculation;
- optional focused diagnostics-off corroboration;
- decision;
- reinterpretation of DIAG-004;
- recommended next **investigation target only**.

## 10. Completion

No Secondary production commit.

Primary may commit only the result artifact and any strictly test/benchmark-only helper required by this task.

Preserve existing unrelated Primary test debt without modification.

Verify both authoritative worktrees clean after completion.

Return:

`POST_OPT_LOW_OVERHEAD_ATTRIBUTION_READY`

plus the classification and metrics above, then:

`READY_FOR_WEB_REVIEW`.
