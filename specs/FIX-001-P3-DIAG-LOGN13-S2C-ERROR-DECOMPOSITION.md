# FIX-001-P3-DIAG-LOGN13-S2C-ERROR-DECOMPOSITION

## Purpose

Explain why the best currently valid arithmetic candidate still misses the public-like threshold after S2C.

Accepted evidence at Primary commit `30e2e602576cf2c8a3c91bbb4d8d724c78bc027d`:

- fixed candidate uses oracle powers;
- q0/q1/q2 profile = 56/39/40;
- G0 local-q2 guard = 2 bits;
- F0 local-q2 guard = 3 bits;
- all 3 DA rounds use the accepted bounded local-q2 pattern and contract after each round;
- EvalMod vs Standard improves materially to about `6.2613e-5` real and `4.6478e-5` imag;
- post-S2C vs Standard is still about `3.3385298628089955e-4`;
- public-like error is `0.010683237260415292`;
- finalization is effectively a fixed ×32 restore, so system success requires approximately:

  `post_s2c_max_component <= 0.01 / 32 = 3.125e-4`.

The current post-S2C result therefore misses the true downstream requirement by only about 6.8%.

This task is diagnostic only. Do not retune G0/F0/DA, do not change S2C production code, and do not modify Secondary.

---

## Required provenance

Primary required base:

`30e2e602576cf2c8a3c91bbb4d8d724c78bc027d`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use the exact accepted G0=2 oracle-power candidate and unchanged workload/reference.

Keep fixed:

- corrected C2S input/workload;
- PS candidate and common plan scale;
- G0/F0 guard schedules;
- all DA local-q2 behavior;
- S2C matrices, factorization, Levels, Scaling, encoding scales and operation order;
- finalization/public-like scaling.

Do not:

- tune any guard bit;
- change q0/q1/q2 parameters;
- extend q2 into S2C;
- change S2C arithmetic;
- redesign generated powers;
- modify Secondary production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary diagnostic mirrors/helpers are allowed.

---

# D0 — reproduce current best candidate

Reproduce compactly:

- EvalMod real/imag vs genuine Standard widened-profile reference;
- post-S2C Fast vs Standard max-component error near `3.3385298628089955e-4`;
- public-like max-component error near `0.010683237260415292`;
- final metadata valid;
- no DA capacity failure.

If mismatch, stop:

`logn13_s2c_error_decomposition_precondition_mismatch`.

---

# D1 — construct a stage-aligned full-RNS S2C mirror

Using the exact same S2C mathematical matrix/factorization/scaling as production, construct a diagnostic full-RNS S2C mirror `T_N` that can be applied to either the Fast EvalMod output or the genuine Standard EvalMod output.

The mirror must preserve:

- identical S2C group boundaries;
- identical plaintext matrix values;
- identical logical Levels/Scales;
- equivalent packing/recombination semantics;
- no artificial metadata-only relabeling.

Validate mirror fidelity by applying `T_N` to the genuine Standard EvalMod input and comparing to genuine Standard S2C output.

Target mirror-fidelity error: `<= 1e-10` where exact/stage-aligned comparison is meaningful; otherwise report the measured numerical floor and explain the exact source.

If the mirror is not trustworthy, classify:

`logn13_s2c_full_rns_mirror_mismatch`.

---

# D2 — exact causal decomposition

Define:

- `E_F`: accepted Fast EvalMod output before S2C;
- `E_S`: genuine Standard EvalMod output before S2C;
- `T_F`: production Fast S2C;
- `T_N`: stage-aligned full-RNS S2C mirror;
- `S_F = T_F(E_F)`;
- `N_F = T_N(E_F)`;
- `N_S = T_N(E_S)`;
- `S_S = genuine Standard S2C(E_S)`.

Compute vector-wise differences internally and report compact max-component metrics for:

1. **Fast S2C implementation effect**: `S_F - N_F`;
2. **propagated EvalMod-input effect**: `N_F - N_S`;
3. **full-RNS mirror fidelity**: `N_S - S_S`;
4. **observed total**: `S_F - S_S`.

Verify vector closure:

`(S_F-N_F) + (N_F-N_S) + (N_S-S_S) = S_F-S_S`

with closure residual <= `1e-10` relative to the decoded slot domain, or the smallest justified numerical tolerance if exact floating decode prevents `1e-10`.

Do not infer causality from max norms alone; compute closure from the underlying vectors and serialize only compact metrics.

---

# D3 — S2C group-by-group amplification

For the propagated-input path only (`E_F` vs `E_S` through `T_N`), record paired error after:

- S2C input;
- group 0 output;
- group 1 output;
- group 2/final S2C output.

At every group record:

- max-component error;
- worst index/component;
- empirical amplification relative to the previous stage;
- intended full-RNS Q01 centered-capacity ratio;
- whether q0/q1 rows from production Fast S2C match the stage-aligned mirror where an exact row comparison is meaningful.

The goal is to identify whether the remaining `~3.34e-4` is:

- mostly pre-existing EvalMod error transformed by mathematically correct S2C;
- an intrinsic Fast S2C arithmetic floor;
- or concentrated in one S2C group.

---

# D4 — linear-delta confirmation

Because S2C is linear, form the decoded EvalMod delta:

`Delta_E = E_F - E_S`.

Apply the full-RNS S2C linear map to the delta directly, with the same matrix/scaling semantics, to obtain `T_N(Delta_E)`.

Compare it to `N_F - N_S`.

Require agreement within the justified decode tolerance.

This confirms that any amplification attributed to S2C is genuine linear transformation of the input error rather than an unrelated state mismatch.

Record:

- input delta max-component;
- output delta max-component;
- empirical total amplification;
- worst output index/component.

---

# D5 — decision classification

Choose exactly one primary classification:

### If Fast S2C implementation effect is negligible

If `S_F-N_F` is <= 10% of the current post-S2C total error and mirror fidelity is negligible, classify:

`logn13_s2c_error_dominated_by_evalmod_input_propagation`

Record the first S2C group with the dominant amplification.

### If Fast S2C implementation effect is substantial

If `S_F-N_F` contributes >10% of current post-S2C total error, classify:

`logn13_s2c_fast_arithmetic_precision_blocker`

Record the first group where Fast diverges materially from `T_N`.

### If mirror/reference construction is the blocker

Use:

`logn13_s2c_full_rns_mirror_mismatch`.

No design fix is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-S2C-ERROR-DECOMPOSITION-summary.json`

Include only:

- provenance;
- D0 control metrics;
- mirror-fidelity metric;
- four-way causal decomposition metrics;
- closure residual;
- compact per-group amplification table;
- linear-delta confirmation;
- current required post-S2C threshold `3.125e-4`;
- current gap to threshold;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize slot vectors, coefficient arrays, full matrices, complete RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*S2C.*Error.*Decomposition` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no candidate retuning;
- no q-parameter changes;
- no generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.