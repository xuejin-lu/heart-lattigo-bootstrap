# QPREFIX-PERF-DIAG-001 — Attribution of Q-prefix v2 Performance Regression

## Status

Executable Codex diagnostic task.

## Task class

`D — Performance Diagnosis`

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- task/spec/results only.

Secondary:
- `xuejin-lu/lattigo@fast-qprefix`
- production code is read-only for this task.

Production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Historical baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Reference result:
`results/QPREFIX-IMPL-009-PERFORMANCE-REVIEW.md`

## Purpose

Explain the approximately 3.8x matched q0=55 P93 slowdown before any optimization is attempted.

Do not modify production code.

The key question is:

[
T_{candidate}/T_{baseline}approx3.8
]

How much is explained by:

1. mandatory authority-width increase from historical q01/q012 to Q-prefix-v2 q0123;
2. q0123 fixed-width CRT / Rescale reconstruction;
3. polynomial/PS workspace and copy traffic;
4. one-bit scalar guard;
5. DoubleAngle;
6. DFT/C2S/S2C row-width expansion;
7. allocations/scratch/copy effects;
8. packing/ring-degree/public-boundary overhead?

The output is attribution, not optimization.

---

# 1. Preserve architecture

The Q-prefix v2 constitution is fixed:

[
rows=QPrefixWidth(Level)=min(Level+1,4).
]

Do not propose or implement a narrower production width in this task.

Historical narrow q01/q012 behavior is a comparison baseline only.

Do not reopen F.

---

# 2. Measurement discipline

Use the same machine/environment as QPREFIX-IMPL-009 where possible.

Record:
- CPU;
- Go version;
- GOOS/GOARCH;
- GOMAXPROCS;
- exact SHAs;
- benchmark commands.

Use:
- warmup;
- `-benchmem`;
- at least 5 runs for sub-100ms components;
- at least 3 runs for full P93 Bootstrap;
- medians;
- raw runs.

Do not cherry-pick.

---

# 3. Stage timing decomposition

Instrument with benchmark-only/test-only timing hooks or isolated benchmark helpers.

Measure current candidate P93 q0=56 and, where API-compatible, matched q0=55 for:

- public pack + N1->N2;
- ScaleDown;
- ModUp;
- Trace;
- each C2S group;
- total C2S;
- EvalMod total;
- polynomial total;
- generated-power construction;
- PS baby-step total;
- PS giant-step total;
- one-bit scalar guard;
- each DoubleAngle round;
- total DoubleAngle;
- each S2C group;
- total S2C;
- N2->N1 + unpack;
- finalization;
- full Bootstrap.

Report the sum of measured stages versus end-to-end timing.

If stages overlap due to fused APIs, state the overlap explicitly rather than double-counting.

---

# 4. Width-scaling microbenchmarks

On the current candidate, use explicit-row APIs to compare the same operation at rows 2, 3, and 4 where semantically valid as a microbenchmark.

These are **capability microbenchmarks only**, not production states.

At minimum:
- Add/Sub;
- scalar/integer Mul;
- pointwise Mul;
- MulRelin/relinearize;
- Rescale;
- automorphism/rotation;
- LinearTransform representative BSGS group;
- polynomial representative multiply/merge;
- Trace;
- ring-degree conversion;
- pack/unpack.

For each operation report:

[
T_3/T_2,quad T_4/T_3,quad T_4/T_2
]

plus B/op and allocs/op.

The expected row-linear baseline is roughly proportional to row count:

[
4/2=2.
]

Operations with substantially super-linear `T_4/T_2` are prime optimization suspects.

Do not interpret rows=2 at high Level as valid production policy.

---

# 5. Rescale attribution

Rescale is the most likely non-linear cost center because q0123 requires fixed-width reconstruction/division.

Measure:
- legacy q01 source;
- q012 source;
- q0123 source;

under the same N and comparable coefficient distribution.

Separate:
- coefficient-domain conversion;
- centered CRT reconstruction;
- logical q_ell division/rounding;
- re-materialization;
- NTT/Montgomery restoration.

If practical, add benchmark-only internal substep timers.

Report whether q0123 Rescale is near row-linear or significantly super-linear versus q01/q012.

---

# 6. Polynomial / PS attribution

For the accepted P93 schedule, measure:
- power-generation phase;
- baby-step phase;
- giant-step phase;
- final PS merge;
- final Rescale;
- guard;
- DoubleAngle.

Count:
- number of Add/Sub;
- integer/scalar Mul;
- Mul;
- MulRelin;
- MulThenAdd;
- Rescale;
- copies;
- workspace zero/copy operations.

Estimate a first-order predicted slowdown using operation counts times measured row-scaling ratios.

Compare predicted slowdown to observed EvalMod slowdown.

If the prediction accounts for most of the observed ~4x, report that clearly.

---

# 7. Copy/allocation attribution

Measure bytes copied or at least row-polynomial copy counts for:
- workspace clone/copy;
- generated powers;
- PS scratch;
- DFT scratch;
- packing/unpacking.

Report:
- B/op;
- allocs/op;
- approximate N*8-byte row traffic per major stage.

Determine whether latency growth is dominated by arithmetic or memory traffic.

---

# 8. Baseline-path characterization

For the historical q0=55 P93 harness, record the actual authoritative width per major stage.

At minimum:
- ModUp;
- C2S;
- polynomial/PS;
- DoubleAngle;
- S2C.

Do not assume "baseline Q012".

Explicitly record whether each stage is:
- q01;
- q012;
- another legacy local-width special case.

This characterization is mandatory for interpreting the 3.8x ratio.

---

# 9. Current-candidate path characterization

Record the current production width per stage under the constitution:

- Level 0 -> q0;
- Level 1 -> q01;
- Level 2 -> q012;
- Level >=3 -> q0123.

For the P93 path, provide the actual Level/rows timeline alongside timing.

---

# 10. Attribution model

Construct a simple decomposition:

[
T_{candidate}
=
T_{row-linear}
+
T_{CRT/Rescale-extra}
+
T_{copy/allocation-extra}
+
T_{other}.
]

Estimate each term from measured evidence.

Report:
- percentage of slowdown explained by width alone;
- percentage explained by Rescale/CRT;
- percentage explained by copying/allocation;
- unexplained residual.

Do not overstate precision; use ranges when appropriate.

---

# 11. Optimization candidates

After attribution only, identify up to five candidates.

For each:
- exact function/path;
- evidence;
- expected mechanism;
- estimated possible gain;
- risk to correctness/architecture;
- whether optimization preserves exact `QPrefixWidth(Level)` policy.

Do not implement them.

Do not recommend narrowing production authority below the constitution in this task.

---

# 12. No production changes

Allowed:
- benchmark-only temporary files;
- test-only instrumentation;
- Primary diagnostic report.

Forbidden:
- production code edits;
- schedule changes;
- parameter changes;
- architecture changes;
- F;
- full-RNS fallback.

Temporary files must be removed before completion.

---

# 13. Completion report

Write Primary:

`results/QPREFIX-PERF-DIAG-001-summary.md`

Include:

- environment;
- baseline path widths;
- candidate path widths;
- stage timing table;
- rows 2/3/4 scaling table;
- Rescale substep table;
- EvalMod/PS operation-count attribution;
- memory/allocation attribution;
- slowdown decomposition;
- top optimization candidates;
- conclusion.

Return one diagnostic status:

- `WIDTH_COST_DOMINANT`
- `RESCALE_COST_DOMINANT`
- `MEMORY_COPY_COST_DOMINANT`
- `MIXED_COST`
- `UNEXPLAINED_PERFORMANCE_REGRESSION`

Then report:
`READY_FOR_WEB_REVIEW`.
