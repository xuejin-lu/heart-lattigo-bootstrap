# QPREFIX-PERF-DIAG-002 — Matched q0=55 Stage-Delta Closure

## Status

Executable Codex diagnostic task.

## Task class

`D — Performance Diagnosis`

## Parent result

Accepted Web review input:

- Primary result commit: `18e3290124ab7fb4a0a5af1f8ecbd33bee8c4474`
- Result: `results/QPREFIX-PERF-DIAG-001-summary.md`
- Classification: `UNEXPLAINED_PERFORMANCE_REGRESSION`

The prior task established:

- matched q0=55 historical public-API baseline: about 30.67 ms;
- matched q0=55 Q-prefix-v2 candidate: about 116.39 ms;
- observed excess: about 85.72 ms;
- most rows4/rows2 kernels are near 2x;
- Rescale is about 2.1–2.4x and centered CRT about 4.39x;
- one-bit guard is not active on the current public P93 path;
- current q0=56 stage microbenchmarks do not quantitatively explain the matched q0=55 3.79x regression.

This task exists to close that exact evidence gap before any optimization.

## Repositories and immutable comparison points

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- task/spec/result only.

Secondary:
- `xuejin-lu/lattigo`
- production code is read-only for this task.

Historical baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Q-prefix-v2 production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Current Secondary branch pointer may be newer only because of task-pointer commits. Do not substitute a newer production source for either immutable comparison SHA.

## Scientific question

For the exact matched q0=55 P93 workload, determine where the candidate's additional end-to-end latency is spent.

Define:

[
Delta T_{E2E}=T_{candidate}-T_{baseline}.
]

For each non-overlapping Bootstrap stage `i`:

[
Delta T_i=T_{candidate,i}-T_{baseline,i}.
]

The target is an additive accounting in which:

[
sum_i Delta T_i approx Delta T_{E2E}.
]

This must be measured from the matched q0=55 baseline and candidate directly. Do not infer it from q0=56 microbenchmarks or rows2/rows4 capability ratios.

## 1. Hard constraints

Do not:

- modify Secondary production code;
- optimize any kernel;
- alter `QPrefixWidth(Level)`;
- narrow candidate authority below the Q-prefix-v2 constitution;
- change the P93 schedule;
- change q0 bit size or other matched parameters;
- change input data between baseline and candidate;
- introduce F;
- introduce Standard/full-RNS fallback;
- weaken correctness/capacity gates;
- use the current q0=56 profile as a substitute for the matched q0=55 experiment.

Temporary test-only / benchmark-only instrumentation is allowed. It must be removed before completion.

If comparable matched instrumentation cannot be constructed without changing production semantics, stop and report `MATCHED_BASELINE_REPLAY_BLOCKED`.

## 2. Checkout discipline

Do not move or rewrite the authoritative `fast-qprefix` worktree to perform the comparison.

Use clean, temporary detached worktrees or another equivalently isolated mechanism for the two immutable Secondary SHAs.

Record for both runs:

- exact Secondary SHA;
- dirty/clean state before measurement;
- CPU;
- Go version;
- GOOS/GOARCH;
- GOMAXPROCS;
- benchmark command;
- timestamp.

After measurement, remove temporary instrumentation and temporary worktrees when safe. The authoritative Secondary branch must remain unchanged except for its task-pointer commit.

## 3. Exact matched workload

Recover the exact q0=55 P93 public-API harness used for the QPREFIX-IMPL-009 matched result.

Baseline and candidate must use:

- the same Primary-side workload/configuration;
- the same q0=55 parameter profile;
- the same input vector;
- the same slot count;
- the same Bootstrap count;
- the same public API boundary;
- the same benchmark duration/count policy;
- the same machine/session where practical.

Before timing, emit a compact fingerprint proving the two runs are matched:

- q0 bit length and value;
- LogN / LogSlots;
- degree / DoubleAngle;
- Q-chain bit-size sequence;
- input length and deterministic input checksum or equivalent compact fingerprint;
- public Bootstrap entry point;
- N1/N2 relation.

If any fingerprint differs, stop. Do not compare unmatched runs.

## 4. Uninstrumented E2E replay

First rerun the exact matched q0=55 public-API benchmark on both immutable SHAs without stage timers.

Measure at minimum:

- Count 1;
- Count 3.

Use:

- one warmup;
- at least 5 measured runs for Count 1;
- at least 3 measured runs for Count 3;
- `-benchmem`;
- raw runs and medians;
- no cherry-picking.

Report:

- latency;
- B/op;
- allocs/op;
- candidate/baseline ratio;
- `Delta T_{E2E}`.

The fresh replay need not reproduce 30.670 ms / 116.390 ms bit-for-bit, but it must be consistent enough to show that the same substantial regression is still present. If not, classify the discrepancy before continuing.

## 5. In-context non-overlapping stage trace

Instrument the full matched q0=55 Bootstrap invocation so stage durations are captured inside the same end-to-end execution.

Use these non-overlapping top-level stages:

1. public pack / N1→N2 boundary;
2. ScaleDown;
3. ModUp including its production Trace work;
4. C2S total;
5. EvalMod real total;
6. EvalMod imaginary total;
7. S2C total;
8. N2→N1 / unpack boundary;
9. public finalization.

Do not separately add nested timers such as Trace to ModUp, DFT groups to C2S/S2C, or polynomial/DoubleAngle to EvalMod when computing the top-level sum.

For both baseline and candidate report, per run:

- total instrumented Bootstrap time;
- every stage duration;
- sum of non-overlapping stages;
- uncovered residual.

Use at least 5 measured Count-1 traced runs per SHA after warmup.

Also report instrumentation overhead by comparing instrumented versus uninstrumented Count-1 medians. If timing hooks add more than 5% overhead to either implementation, state that explicitly and do not claim fine-grained closure from the affected trace without justification.

## 6. Delta table and closure calculation

For every top-level stage report:

- baseline median;
- candidate median;
- candidate/baseline ratio;
- `Delta T_i`;
- percentage of the measured `Delta T_{E2E}`.

Then calculate:

[
R=Delta T_{E2E}-sum_iDelta T_i.
]

and:

[
residual_fraction=rac{|R|}{|Delta T_{E2E}|}.
]

Primary closure target:

[
residual_fraction le 0.10.
]

Also report the stage-sum coverage of each implementation separately.

Do not hide a negative stage delta or fold it into another category.

## 7. One bounded drill-down only

Identify the single top-level stage with the largest positive `Delta T_i`.

Drill down only that stage on both matched q0=55 SHAs.

Use the smallest semantically equivalent substage mapping available.

If the largest stage is EvalMod, measure at minimum:

- generated-power construction;
- PS baby steps;
- PS giant-step merges;
- final PS Rescale / result finalization;
- DoubleAngle total and each round when practical.

If the largest stage is C2S or S2C, measure per DFT group.

If the largest stage is ModUp, separate basis raise / scale alignment from Trace where the source permits.

If the largest stage is packing/unpacking, separate copy versus ring-degree conversion where the source permits.

Do not drill down a second stage in this task. The purpose is to identify the first measured dominant delta, not to profile the entire program again.

## 8. Allocation evidence

The previous task lacked a matched baseline allocation comparison.

For the fresh matched q0=55 uninstrumented E2E replay, report:

- baseline B/op and allocs/op;
- candidate B/op and allocs/op;
- absolute delta;
- ratio.

For the single drilled-down dominant stage, report B/op / allocs/op when the test harness can do so without changing semantics.

Do not classify memory as causal solely from allocation counts. Use timing plus allocation evidence.

## 9. Interpretation rules

This task may conclude that a stage is the dominant measured contributor only from the fresh matched q0=55 delta table.

Do not use:

- current q0=56 stage timings;
- rows2/3/4 capability ratios;
- isolated CRT ratios;

as substitutes for matched stage deltas.

Those prior measurements may be used only to interpret a stage after the matched delta identifies it.

Do not call mandatory width cost an implementation bug.

Do not authorize an optimization in this task.

## 10. Required result artifact

Write:

`results/QPREFIX-PERF-DIAG-002-summary.md`

Include:

- provenance and immutable SHAs;
- matched-workload fingerprint;
- uninstrumented Count-1/Count-3 replay;
- matched B/op / allocs/op;
- in-context top-level stage table for both SHAs;
- stage delta table;
- additive closure math;
- instrumentation overhead;
- one bounded dominant-stage drill-down;
- interpretation;
- remaining uncertainty.

Return exactly one primary status:

- `MATCHED_STAGE_ATTRIBUTION_CLOSED`
- `MATCHED_STAGE_ATTRIBUTION_UNCLOSED`
- `MATCHED_BASELINE_REPLAY_BLOCKED`

Additionally report:

- `TOP_DELTA_STAGE=<stage name>`
- `RESIDUAL_FRACTION=<value>`

## 11. Completion / repository state

No Secondary production commit is permitted.

Primary may commit only the diagnostic result artifact required by this task.

The existing documented Primary test debt
`TestFIX001P3GenuineStandardPublicVsStagedConsistency` /
`P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`
must not be modified, bypassed, or relabeled by this task.

Before completion verify:

- temporary instrumentation removed;
- temporary worktrees removed when safe;
- authoritative Primary worktree clean after result commit;
- authoritative Secondary `fast-qprefix` worktree clean;
- Secondary production source unchanged.

Then normal fast-forward push the Primary result under the standing workflow rules and report:

`READY_FOR_WEB_REVIEW`.
