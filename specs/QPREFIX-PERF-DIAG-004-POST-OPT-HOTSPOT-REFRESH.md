# QPREFIX-PERF-DIAG-004 — Post-Optimization Hotspot Refresh

## Status

Executable diagnostic task.

## Task class

`D — Performance Diagnosis`

## Accepted parent

Production optimization:
- Primary summary: `results/QPREFIX-PERF-OPT-001-summary.md`
- Primary commit: `3fa3a7605691ffe471eb47c3eea98ea3ce023c5a`
- Secondary implementation: `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`
- classification: `TRANSACTIONAL_RESCALE_STAGING_READY`

Accepted production-speed effect:
- q0=55 P93 Count-1 E2E median: about `120.331 -> 81.610 ms` (`32.2%` faster);
- representative rows4 Rescale median: about `1.922 -> 1.207 ms` (`37.2%` faster);
- allocations materially flat.

The previous diagnosis that Rescale dominated the generated-power regression was made before this repair and must not be reused as the current bottleneck conclusion.

## Goal

Refresh the performance attribution after the successful transactional-staging repair and identify the **current** dominant production bottleneck.

This task must answer:

1. What are the current non-overlapping Bootstrap stage costs?
2. Which current stage consumes the largest share of diagnostics-off E2E time?
3. Inside the current generated-power path, how much time remains in Rescale versus the other power categories?
4. How much of the pre-opt -> post-opt improvement is actually explained by the Rescale repair?
5. What should be investigated next, based on fresh evidence only?

Do not optimize production code in this task.

## Repositories and refs

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`

Secondary:
- active branch `fast-qprefix`

Pre-optimization comparison ref:
`74cb73dcff6c552cda0671ed7faea897b448fbbd`

Post-optimization candidate:
`d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

Both refs contain the reusable fastdiag hooks and are eligible for framework comparison.

Historical q01 baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Historical q01 numbers from QPREFIX-PERF-DIAG-002 may be cited only as archived context. Do not pretend they are a same-session comparison in this task.

## 1. Hard constraints

Do not:

- modify Secondary production code;
- add task-specific timing overlays;
- change Rescale;
- change `QPrefixWidth(Level)`;
- alter parameters, polynomial schedule, P93 schedule, or generated-power DAG;
- introduce private-F or Standard/full-RNS fallback;
- use all-scope trace latency as the authoritative production E2E number.

Use the reusable `scripts/fastdiag` framework for supported traces.

## 2. Diagnostics-off authoritative E2E refresh

On the current post-opt candidate, rerun the ordinary q0=55 P93 benchmark with diagnostics disabled.

Measure:

- Count 1;
- Count 3 if the existing benchmark supports it without new production instrumentation.

For Count 1 use at least:

- one warmup;
- 7 measured samples;
- `-benchmem`.

Report raw samples, median latency, B/op, and allocs/op.

If Count 3 is unavailable in the current permanent benchmark, do not create an overlay solely for it; report that limitation.

This diagnostics-off result is the authoritative current E2E timing.

## 3. Stage-only current trace

Run the reusable framework on the current candidate with:

```sh
./scripts/fastdiag trace \
  --profile p93-q55 \
  --trace stage \
  --warmup 1 \
  --repetitions 7
```

Use the exact supported CLI spelling if it differs.

For each non-overlapping stage report median time:

- bootstrap parent;
- pack / N1->N2;
- ScaleDown;
- ModUp including Trace;
- C2S;
- EvalMod real;
- EvalMod imaginary;
- S2C;
- unpack / N2->N1;
- public finalization.

Calculate:

[
stage_share_i = rac{T_i}{sum_j T_j}
]

and stage-sum closure against the traced Bootstrap parent.

Because `stage` scope previously showed negligible overhead, this trace may be used to rank current top-level stages. Still report measured stage-only overhead versus diagnostics-off E2E.

Return:

`CURRENT_TOP_STAGE=<stage>`

## 4. Current deep structural trace

Run current candidate with:

```sh
./scripts/fastdiag trace \
  --profile p93-q55 \
  --trace stage,power,rescale \
  --warmup 1 \
  --repetitions 5
```

This all-scope trace is structural only because its overhead is known to be large.

Report, for current candidate:

### Generated powers

- parent generated-power median;
- each generated power `{2,3,4,6,8,16}`;
- copy/workspace;
- balanced integer scaling;
- Mul/MulRelin;
- Rescale;
- doubling;
- recurrence correction;
- residual.

### Rescale

- total Rescale;
- preflight compute-and-stage;
- materialization commit;
- prefix-to-coefficient;
- reconstruction/center/round/capacity;
- residue staging;
- NTT/Montgomery restore.

Do not interpret all-scope wall time as production E2E.

Return:

`CURRENT_POWER_TOP_CATEGORY=<category>`

and:

`CURRENT_RESCALE_SHARE_OF_POWER=<fraction>`

## 5. Pre-opt versus post-opt framework comparison

Use the permanent compare command:

```sh
./scripts/fastdiag compare \
  --baseline 74cb73dcff6c552cda0671ed7faea897b448fbbd \
  --candidate d50ff4db757d4a2b9922937a4e7f316fd3f286b9 \
  --profile p93-q55 \
  --trace stage,power,rescale
```

Use at least 5 repetitions after warmup.

Report:

- top-level stage deltas;
- generated-power category deltas;
- Rescale preflight/materialization changes;
- closure residuals.

This comparison is for **where the repair saved time**, not for absolute production-speed authority.

Calculate what fraction of the measured traced improvement is assigned to:

- EvalMod real;
- EvalMod imaginary;
- C2S/S2C;
- generated-power Rescale;
- other.

Do not double-count nested events.

## 6. Fresh current bottleneck decision

The next bottleneck must be selected from the post-opt **current** absolute stage/category costs, not from the old pre-opt delta ranking.

Classify the current state as exactly one:

- `CURRENT_RESCALE_STILL_DOMINANT`
- `CURRENT_EVALMOD_NONRESCALE_DOMINANT`
- `CURRENT_DFT_DOMINANT`
- `CURRENT_MIXED_HOTSPOTS`
- `POST_OPT_ATTRIBUTION_UNCLOSED`

Use these rules:

### CURRENT_RESCALE_STILL_DOMINANT

Only if current Rescale accounts for at least 50% of current generated-power time **and** generated powers remain the dominant measured contributor inside the top EvalMod stage.

### CURRENT_EVALMOD_NONRESCALE_DOMINANT

If EvalMod remains the top stage but Rescale no longer meets the dominance condition, and non-Rescale EvalMod work is the largest clear remaining contributor.

### CURRENT_DFT_DOMINANT

If C2S or S2C becomes the largest current top-level stage or their combined cost clearly dominates the remaining nontrivial cost.

### CURRENT_MIXED_HOTSPOTS

If no single category explains at least 40% of the current nontrivial runtime and two or more categories are comparable.

### POST_OPT_ATTRIBUTION_UNCLOSED

If trace closure or overhead makes the ranking unreliable.

Do not authorize the next production optimization inside this diagnostic task.

## 7. Historical context

For context only, include the archived q01 stage medians from QPREFIX-PERF-DIAG-002:

- C2S about `4.09 ms`;
- EvalMod real about `9.85 ms`;
- EvalMod imaginary about `9.82 ms`;
- S2C about `4.11 ms`;
- total about `28.1 ms`.

If current candidate is compared numerically to these values, label the comparison:

`ARCHIVED_CROSS_SESSION_REFERENCE_ONLY`

Do not compute causal percentages from this cross-session comparison.

## 8. Framework validation

This task is also a real-use validation of the reusable framework.

Record:

- whether `trace --trace stage` succeeds;
- whether `trace --trace stage,power,rescale` succeeds;
- whether `compare` succeeds for the two hook-enabled refs;
- whether closure is mathematically consistent;
- whether any event-schema ambiguity reappears.

If the framework itself blocks valid attribution, return `POST_OPT_ATTRIBUTION_UNCLOSED` and report the exact framework defect. Do not silently patch it in this task.

## 9. Required artifact

Write:

`results/QPREFIX-PERF-DIAG-004-summary.md`

Include:

- provenance;
- current diagnostics-off E2E raw samples;
- stage-only current table and closure;
- deep current power/Rescale table;
- pre-opt/post-opt compare table;
- current bottleneck classification;
- archived historical q01 context with explicit cross-session label;
- recommended **investigation target**, not an implementation plan;
- remaining uncertainty.

## 10. Completion

No Secondary production commit is allowed.

Primary may commit only the result artifact required by this task.

Preserve the existing unrelated Primary:

`TestFIX001P3GenuineStandardPublicVsStagedConsistency` /
`P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`

without modification or bypass.

Verify Primary and Secondary worktrees are clean after completion.

Return:

`POST_OPT_HOTSPOT_REFRESHED`

plus:

- `CURRENT_TOP_STAGE=<stage>`
- `CURRENT_POWER_TOP_CATEGORY=<category>`
- `CURRENT_BOTTLENECK_CLASS=<classification>`

Then report:

`READY_FOR_WEB_REVIEW`.
