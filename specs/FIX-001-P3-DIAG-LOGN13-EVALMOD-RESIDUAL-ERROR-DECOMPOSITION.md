# FIX-001-P3-DIAG-LOGN13-EVALMOD-RESIDUAL-ERROR-DECOMPOSITION

## Purpose

Identify which remaining EvalMod error source must be reduced to close the final LogN13 public-like gap.

Accepted evidence at Primary commit `02d472f7e5c02709bafd5ef67586a8337684c4f3`:

- accepted oracle-power candidate uses q0/q1/q2 = 56/39/40;
- G0 local-q2 guard = 2 bits;
- F0 local-q2 guard = 3 bits;
- all three DA rounds use validated bounded local-q2 arithmetic and contract after each round;
- EvalMod vs genuine Standard: about `6.261306805777143e-5` real and `4.647793870173522e-5` imag;
- Fast S2C implementation effect = 0;
- full-RNS S2C mirror fidelity = 0;
- S2C linear-delta agreement passes (`~2.39e-13`);
- total EvalMod-delta -> post-S2C amplification is about `5.3320x`;
- post-S2C error = `3.3385298628089955e-4`;
- required post-S2C threshold = `3.125e-4`;
- remaining post-S2C gap = `2.1352986280899543e-5`;
- public-like error = `0.010683237260415292`.

S2C itself is therefore closed as an implementation blocker. This task decomposes the residual EvalMod error into:

1. polynomial/PS input semantic error propagated through DA;
2. normalized DA CKKS arithmetic/rounding error;
3. Fast/local-q2 implementation effect;
4. residual reference/model-alignment floor.

Diagnostic only. No design changes are authorized.

---

## Required provenance

Primary required base:

`02d472f7e5c02709bafd5ef67586a8337684c4f3`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use the exact accepted G0=2/F0=3 oracle-power candidate and the accepted all-round local-q2 DA schedule.

Keep fixed:

- corrected C2S workload/input;
- polynomial, basis, PS decomposition and plan scale;
- G0/F0 guard schedules;
- DA K schedule, constants and Levels;
- all local-q2 activation/contraction rules;
- S2C matrices/factorization/scaling;
- final ×32/public-like restore semantics.

Do not:

- retune guards/scales/K/constants;
- modify Secondary production code;
- redesign generated powers;
- alter q parameters;
- change S2C;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary diagnostic full-RNS/high-precision mirrors and stage-aligned oracle injection are allowed.

---

# E0 — reproduce accepted control

Reproduce compactly:

- accepted polynomial output before DA;
- EvalMod real/imag vs genuine Standard;
- post-S2C error;
- public-like error;
- final metadata;
- no capacity/row/rounded-division failure.

If mismatch, stop:

`logn13_evalmod_residual_decomposition_precondition_mismatch`.

---

# E1 — define stage-aligned polynomial references

Define decoded/stage-aligned polynomial states before DoubleAngle:

- `P_F`: actual accepted Fast polynomial/PS output (oracle powers, G0=2, F0=3);
- `P_O`: source-exact mathematical polynomial oracle evaluated on the same preprocessed input, injected/materialized at the same logical DA input Level/Scale without introducing material semantic error.

Validate `P_O` injection/materialization against a high-precision decoded oracle.

Target injection error: `<= 1e-10` in the decoded slot domain. If not achievable because of an unavoidable encoding floor, report and justify the measured floor; it must be at least 10x below the current `P_F-P_O` error.

Record:

- `P_F-P_O` real/imag max-component error;
- worst index/component;
- input Level/Scale metadata.

If the oracle injection is not trustworthy, classify:

`logn13_evalmod_residual_oracle_injection_mismatch`.

---

# E2 — define three DA evaluators

Use the exact accepted normalized DA recurrence/schedule.

### `DA_F`

Actual accepted Fast/local-q2 DA implementation.

### `DA_N`

Stage-aligned normalized full-RNS DA mirror using the same CKKS integer arithmetic, constants, rescale divisors, Levels and Scales as the accepted Fast design, but without q0/q1 truncation ambiguity.

Require Fast/full-RNS row equality at all exact checkpoints for `DA_F(P_F)` vs `DA_N(P_F)`.

### `DA_I`

High-precision ideal real/complex recurrence implementing the same mathematical normalized DoubleAngle transformation with no CKKS rounding/rescale quantization.

Validate the recurrence/constants/K schedule against source semantics before using it for causal attribution.

---

# E3 — four-state causal chain

Construct:

- `E_F = DA_F(P_F)`
- `N_F = DA_N(P_F)`
- `N_O = DA_N(P_O)`
- `I_O = DA_I(P_O)`
- `E_S = genuine Standard EvalMod output`

Report compact vector max-component metrics for:

1. **Fast/local-q2 implementation effect**: `E_F - N_F`;
2. **PS-input propagation effect**: `N_F - N_O`;
3. **normalized DA arithmetic/rounding effect on oracle input**: `N_O - I_O`;
4. **ideal/reference alignment floor**: `I_O - E_S`;
5. **observed total**: `E_F - E_S`.

Compute closure from underlying vectors:

`(E_F-N_F) + (N_F-N_O) + (N_O-I_O) + (I_O-E_S) = E_F-E_S`.

Require closure residual <= `1e-10` in decoded slot domain, or report the smallest justified floating/high-precision floor.

Do not infer additive percentages solely from max norms because components may cancel directionally.

---

# E4 — stage-by-stage DA attribution

For both `P_F` and `P_O`, record compact decoded error at:

- DA input;
- round0 after square / multiplier / constant / rescale;
- round1 corresponding checkpoints;
- round2 corresponding checkpoints;
- final EvalMod output.

At each checkpoint record:

- `DA_N(P_F) - DA_N(P_O)` max-component error (propagated PS-input component);
- `DA_N(P_O) - DA_I(P_O)` max-component error (DA arithmetic component);
- worst index/component;
- exact capacity/row-agreement flags already required by the accepted design.

Goal: identify the first DA checkpoint where each component becomes materially significant.

---

# E5 — downstream-sensitive counterfactuals

Use the already-validated full-RNS S2C linear mirror and final ×32 restore to evaluate counterfactual system sufficiency.

Construct these four downstream cases in decoded/stage-aligned form:

### H0 — actual accepted candidate

`S2C(E_F)` then final restore.

Must reproduce public-like `~0.01068323726`.

### H_PS — perfect polynomial/PS semantics, same DA arithmetic

`S2C(N_O)` then final restore.

Question answered:

> If PS/polynomial input error were eliminated but the current normalized DA CKKS arithmetic remained unchanged, would the final public-like threshold pass?

### H_DA — current PS semantics, ideal DA arithmetic

Compute `I_F = DA_I(P_F)`, then `S2C(I_F)` and final restore.

Question answered:

> If DA arithmetic/rounding were eliminated but current PS output semantics remained unchanged, would the final public-like threshold pass?

### H_BOTH — ideal polynomial input + ideal DA

`S2C(I_O)` then final restore.

This is the diagnostic floor for the current mathematical polynomial/S2C/finalization chain.

For H0/H_PS/H_DA/H_BOTH record:

- EvalMod-domain max-component vs Standard;
- post-S2C max-component vs Standard;
- public-like max-component vs Standard;
- pass/fail against `1e-2`.

The S2C mirror/final restore must be the same validated linear mapping used in the prior decomposition.

---

# E6 — S2C-sensitive component projection

Pass each causal EvalMod component through the full-RNS S2C linear map independently:

- `T(E_F-N_F)`
- `T(N_F-N_O)`
- `T(N_O-I_O)`
- `T(I_O-E_S)`

Record max-component metrics and vector closure after S2C.

This determines which EvalMod component actually contributes to the output direction that misses the `3.125e-4` post-S2C threshold.

Also record the amount by which eliminating each component alone would reduce the actual worst public-like output, using vector counterfactuals rather than norm subtraction.

---

# E7 — decision classification

Choose exactly one primary classification based on the counterfactual system outcomes.

### PS fix alone sufficient

If H_PS passes `<=1e-2` while H_DA does not:

`logn13_evalmod_residual_ps_input_blocker`

### DA fix alone sufficient

If H_DA passes while H_PS does not:

`logn13_evalmod_residual_da_arithmetic_blocker`

### Either alone sufficient

If both H_PS and H_DA independently pass:

`logn13_evalmod_residual_multiple_single_fix_options`

Record which has the larger public-like margin.

### Neither alone sufficient, both together sufficient

If H_PS and H_DA both fail but H_BOTH passes:

`logn13_evalmod_residual_joint_ps_da_blocker`

### Even idealized current chain insufficient

If H_BOTH fails:

`logn13_evalmod_residual_mathematical_chain_floor`

### Implementation mismatch

If `E_F-N_F` is materially nonzero despite exact checkpoint contracts:

`logn13_evalmod_residual_fast_implementation_mismatch`

No implementation/design fix is authorized in this diagnostic task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-EVALMOD-RESIDUAL-ERROR-DECOMPOSITION-summary.json`

Include only:

- provenance;
- E0 controls;
- P_F vs P_O metric and oracle-injection floor;
- four-state EvalMod causal decomposition;
- closure residual;
- compact DA checkpoint attribution table;
- H0/H_PS/H_DA/H_BOTH table;
- S2C-projected component metrics;
- required post-S2C threshold `3.125e-4`;
- public-like threshold `1e-2`;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize slot vectors, coefficient arrays, complete RNS rows, matrices or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*EvalMod.*Residual.*Decomposition` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no candidate retuning;
- no q-parameter/S2C changes;
- no generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.