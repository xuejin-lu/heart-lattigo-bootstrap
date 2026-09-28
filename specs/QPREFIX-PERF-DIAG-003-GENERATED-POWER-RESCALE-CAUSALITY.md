# QPREFIX-PERF-DIAG-003 — Matched Generated-Power Rescale Causality

## Status

Executable Codex diagnostic task.

## Task class

`D — Performance Diagnosis`

## Parent result

Accepted Web review input:

- Primary result commit: `71964eb7192d05971933d07913518b1cc5dde083`
- Result: `results/QPREFIX-PERF-DIAG-002-summary.md`
- Classification: `MATCHED_STAGE_ATTRIBUTION_CLOSED`
- `TOP_DELTA_STAGE=EvalMod real`
- `RESIDUAL_FRACTION=0.003343`

Fresh matched q0=55 evidence:

- baseline Count-1 median: `28.114291 ms`
- candidate Count-1 median: `116.130333 ms`
- E2E delta: `+88.016042 ms`
- EvalMod real delta: `+34.329583 ms`
- EvalMod imaginary delta: `+34.098749 ms`
- within EvalMod real, generated-power construction delta: `+20.998209 ms`
  (`5.725250 -> 26.723459 ms`)

This task explains that generated-power delta before any optimization is authorized.

## Repositories and immutable comparison points

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- spec/result only.

Secondary:
- `xuejin-lu/lattigo`
- production code read-only.

Historical q0=55 baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Q-prefix-v2 production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Use isolated detached worktrees or equivalent. Do not move/rewrite the authoritative `fast-qprefix` worktree.

## Source fact to verify, not assume

Web review observed an important implementation difference that must be checked against runtime evidence:

Historical q01 `rescaleN` performs the q01 coefficient reconstruction/division/materialization path once per component.

Q-prefix-v2 `rescaleNQPrefix` performs:

1. a full preflight call to `rescaleQPrefixComponent(..., dst=nil, ...)`;
2. then a full materialization call to `rescaleQPrefixComponent(..., dst=&opOut..., ...)`.

This design preserves failure-before-mutation semantics, but appears to repeat coefficient-domain conversion + CRT reconstruction + division for in-place Rescale.

Do not treat this source observation as the conclusion. Measure its contribution.

## Scientific question

For the exact matched q0=55 P93 EvalMod-real generated-power phase:

[
Delta T_{powers}
=
T_{candidate,powers}-T_{baseline,powers}
approx 20.998209	ext{ ms}.
]

Determine how much of this delta comes from:

- Rescale total;
- candidate Rescale preflight pass;
- candidate Rescale materialization pass;
- multiplication / relinearization;
- integer scaling used by balanced pre-Rescale;
- copies / workspace preparation;
- Chebyshev doubling;
- recurrence correction / aligned subtraction;
- other orchestration.

The task must distinguish:

[
	ext{Q-prefix width cost}
]

from

[
	ext{extra work caused by the two-pass transactional Rescale implementation}.
]

No production repair is allowed yet.

## 1. Hard constraints

Do not:

- modify committed Secondary production code;
- optimize Rescale;
- remove the preflight pass;
- weaken failure-before-mutation semantics;
- change q0=55 parameters;
- change P93 polynomial degree/schedule;
- change generated-power recursion or power set;
- change `QPrefixWidth(Level)`;
- introduce narrower production authority;
- introduce F or Standard/full-RNS fallback;
- use q0=56 as a substitute for this matched q0=55 analysis.

Temporary overlay/test-only instrumentation is allowed and must be removed before completion.

## 2. Reproduce the matched generated-power phase

Use the same q0=55 P93 workload fingerprint accepted in QPREFIX-PERF-DIAG-002.

Before detailed timing, confirm both immutable SHAs still reach:

- the same EvalMod-real input Level/Scale;
- the same polynomial degree and Chebyshev basis;
- the same generated power set;
- the same input checksum or equivalent deterministic fingerprint.

Record any difference and stop if the workloads are not semantically matched.

## 3. Exact runtime schedule

For both baseline and candidate, record the actual generated-power DAG and runtime branch taken for every generated power.

At minimum report powers:

`{2,3,4,6,8,16}`

For each generated power report:

- split `(a,b)`;
- input levels;
- input scales;
- authoritative row count;
- lazy/non-lazy path;
- balanced pre-Rescale selected: yes/no;
- post-product schedule selected: yes/no;
- number of operand copies;
- number of integer-scale multiplies;
- number of Relinearize calls;
- number of Mul / MulRelin calls;
- number of Rescale calls;
- number of doubling Adds;
- correction type:
  - subtract one;
  - aligned subtraction;
  - none.

The operation-count table must be derived from actual runtime instrumentation, not source inspection alone.

If baseline and candidate execute different logical schedules, isolate that schedule difference explicitly before attributing kernel speed.

## 4. In-context generated-power timing

Inside the matched EvalMod-real invocation, time generated-power work with non-overlapping categories.

For each SHA, aggregate over the complete generated-power phase:

1. copy / workspace preparation;
2. balanced integer scaling;
3. Relinearize;
4. Mul / MulRelin;
5. Rescale total;
6. Chebyshev doubling;
7. recurrence correction / aligned subtraction;
8. residual orchestration.

Use at least 5 measured EvalMod-real runs after warmup.

Report raw runs and medians.

The sum of non-overlapping generated-power categories should cover at least 95% of the measured generated-power phase. If it does not, report the residual and do not over-attribute.

## 5. Per-power timing

For each power in `{2,3,4,6,8,16}`, report baseline and candidate median total time plus candidate-minus-baseline delta.

Identify whether the slowdown is distributed approximately uniformly across powers or concentrated in specific powers/levels/row widths.

For each Rescale executed while constructing that power, record:

- source Level;
- source rows;
- target Level;
- target rows;
- in-place versus out-of-place;
- baseline/candidate duration.

## 6. Candidate Rescale internal pass timing

For Q-prefix-v2 candidate only, instrument each generated-power Rescale into these non-overlapping pieces:

- preflight:
  - prefix-to-coefficient conversion;
  - CRT reconstruction / centering / rounded division / capacity check;
- materialization:
  - prefix-to-coefficient conversion;
  - CRT reconstruction / centering / rounded division / capacity check;
  - residue rematerialization + NTT/Montgomery restore;
- call/orchestration residual.

You may group reconstruction+centering+division if separating them would distort the measurement.

Report:

[
T_{preflight},
T_{materialization},
T_{rescale}.
]

Check that:

[
T_{preflight}+T_{materialization}+T_{residual}
approx T_{rescale}.
]

Do not add nested component timings twice.

## 7. Historical baseline Rescale comparison

Instrument the corresponding historical q01 Rescale calls in the same generated-power workload.

At minimum separate:

- coefficient-domain conversion;
- q01 CRT + rounded division;
- residue materialization + NTT/Montgomery restore;
- orchestration residual.

The goal is a directly matched comparison:

[
T_{candidate,rescale}
-
T_{baseline,rescale}.
]

Do not substitute current-candidate rows2 capability benchmarks for the historical baseline.

## 8. Causal accounting

Compute:

[
Delta T_{powers}
=
T_{candidate,powers}-T_{baseline,powers}.
]

Then estimate directly from matched timing:

[
Delta T_{powers}
=
Delta T_{rescale}
+
Delta T_{mul/relin}
+
Delta T_{integer-scale}
+
Delta T_{copy}
+
Delta T_{add/correction}
+
Delta T_{other}.
]

Within candidate Rescale additionally report the share of the generated-power delta represented by preflight time:

[
S_{preflight}
=
rac{T_{candidate,rescale,preflight}}{Delta T_{powers}}.
]

This is an attribution metric, not an immediately reclaimable speedup claim.

Also calculate a **counterfactual lower-bound timing estimate only**:

[
T_{candidate,powers}^{no repeated preflight}
=
T_{candidate,powers}-T_{candidate,rescale,preflight}.
]

Clearly label it as arithmetic accounting only. It is not permission to remove preflight, because a correct repair must preserve transactional failure-before-mutation behavior and may need staging/copy cost.

## 9. Interpretation gate

Return one primary classification:

- `GENERATED_POWER_RESCALE_DOMINANT`
- `GENERATED_POWER_MUL_RELIN_DOMINANT`
- `GENERATED_POWER_MIXED_COST`
- `GENERATED_POWER_CAUSALITY_UNCLOSED`

Use `GENERATED_POWER_RESCALE_DOMINANT` only if matched Rescale delta explains at least 60% of `Delta T_powers`.

Additionally report:

- `RESCALE_DELTA_SHARE=<fraction>`
- `PREFLIGHT_TIME_SHARE_OF_POWER_DELTA=<fraction>`
- `GENERATED_POWER_RESIDUAL_FRACTION=<fraction>`

If the two-pass preflight is large, state that it is a **measured optimization candidate**, not yet a correctness-safe repair.

## 10. No repair in this task

Do not implement:

- single-pass Rescale;
- staged result buffers;
- changed transactional semantics;
- specialized q0123 CRT;
- fewer active rows;
- schedule changes.

A later Web-authored implementation spec will decide whether and how to optimize after this causal attribution closes.

## 11. Required result artifact

Write:

`results/QPREFIX-PERF-DIAG-003-summary.md`

Include:

- provenance;
- matched workload fingerprint;
- runtime generated-power DAG;
- operation counts;
- per-power timing;
- generated-power non-overlapping timing table;
- matched baseline/candidate Rescale table;
- candidate preflight/materialization table;
- causal delta accounting;
- counterfactual arithmetic estimate;
- classification and remaining uncertainty.

## 12. Validation and completion

Run the task-specific focused tests needed to prove instrumentation did not alter outputs.

The existing documented unrelated Primary debt
`TestFIX001P3GenuineStandardPublicVsStagedConsistency` /
`P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`
must not be changed or bypassed.

Before completion verify:

- all temporary instrumentation removed;
- temporary detached worktrees removed when safe;
- authoritative Secondary production source unchanged;
- Primary worktree clean after the result commit;
- authoritative Secondary `fast-qprefix` worktree clean.

Commit and normal fast-forward push only the Primary result artifact under the standing workflow rules.

Then report:

`READY_FOR_WEB_REVIEW`.
