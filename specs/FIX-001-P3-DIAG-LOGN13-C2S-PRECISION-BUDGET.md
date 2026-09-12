# FIX-001-P3-DIAG-LOGN13-C2S-PRECISION-BUDGET

## Purpose

The accepted LogN13 end-to-end failure is now known to contain two independent effects:

1. Fast CoeffsToSlots (C2S) already differs semantically from genuine Standard C2S before EvalMod.
2. Fast EvalMod also differs from genuine Standard EvalMod on matched semantic inputs.

The next task must address the **earliest independently sufficient cause first**: C2S precision.

Accepted evidence from Primary `c0a058ac2e5caafa9cdc9eb63e68cd43269c329f`:

- genuine Standard E2E `max_component_abs = 5.754406148805735e-8`;
- Fast-vs-Standard C2S real `max_component_abs = 0.0003098126393859551`;
- Fast-vs-Standard C2S imag `max_component_abs = 0.00026367655745716595`;
- re-encoding the Fast C2S real semantics into genuine Standard is accurate to `6.47e-13`;
- genuine Standard EvalMod on Fast C2S semantics differs from genuine Standard native EvalMod by `0.15862407053340882`;
- observed local amplification ratio is `511.99999731385975`;
- therefore the old intermediate `1e-2` threshold is not a sufficient C2S acceptance criterion for this downstream nonlinear circuit.

For the observed real branch, an empirical local input-error budget that maps to the final `1e-2` target is:

`C2S_budget = 1e-2 / 511.99999731385975 ≈ 1.953125010246812e-5`.

The current real C2S error is about 15.86 times larger than this budget.

This task is diagnostic only. It must locate the first C2S factor/group operation that exceeds the downstream-sensitive precision budget and determine whether the dominant increment is introduced by the linear transform or by the group Rescale.

Do not modify Secondary production code in this task.

---

## Fixed provenance

Primary required base:

`c0a058ac2e5caafa9cdc9eb63e68cd43269c329f`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- `configs/bootstrap_config.logN13.json`
- LogN = 13
- full packing, LogSlots = 12
- CircuitOrder = `ModUpThenEncode`
- C2S factorization levels = `[1,1,1,1]`
- four C2S factor groups
- final correctness threshold = `1e-2`

Use the same deterministic `reproducibleInput.v1`.

---

# C0 — preserve causal disposition from the prior task

Record explicitly:

- `PREVIOUS_RAW_STANDARD_SQUARE_PROJECTION_DISPOSITION = invalid_as_q01_oracle_after_centered_capacity_loss`;
- the prior `da_round0_square` Standard full-RNS exact coefficient capacity exceeded `Q01/2` (`outside_count = 99`), therefore a q0/q1 projection of that raw Standard square is not a valid exact semantic oracle;
- Fast normalized square itself remained within q0/q1 centered capacity.

Do not continue using the prior raw-Standard-square q0/q1 projection to localize EvalMod in this task.

Also preserve:

- `MATCHED_INPUT_ASYMMETRY = false`;
- Fast EvalMod matched-input divergence remains an unresolved **separate downstream issue**.

This task does not attempt to solve that downstream issue.

---

# C1 — genuine Standard control

Reuse the already validated genuine full-RNS Standard evaluator construction from the semantic-bisect task.

Require proof:

- ordinary CKKS evaluator non-nil;
- DFT evaluator non-nil;
- Mod1 evaluator non-nil;
- Fast compatibility path not used.

Reproduce genuine Standard E2E pass (`<=1e-2`) before using it as a control.

Failure classification:

`logn13_c2s_precision_standard_control_failure`.

---

# C2 — reproduce identical ModUp input semantics

Run both Fast and genuine Standard through:

`input -> PackAndSwitchN1ToN2 -> ScaleDown -> ModUp`

using independent copies of the same deterministic input.

Before entering C2S require:

- Fast vs Standard projected semantic error <= `1e-9`;
- Level / Scale / LogDimensions agree;
- no input mutation;
- same four-group C2S matrix schedule.

The existing semantic bisect observed exact semantic equality at ModUp; reproduce this.

Failure classification:

`logn13_c2s_precision_modup_precondition_failure`.

---

# C3 — downstream-sensitive C2S budget

Do not use `1e-2` as the C2S stage acceptance threshold.

Recompute the local downstream sensitivity from the current run using genuine Standard EvalMod:

1. obtain genuine Standard native C2S real output;
2. obtain Fast native C2S real decoded semantics;
3. re-encode Fast C2S real semantics into a genuine Standard ciphertext with reconstruction error <= `1e-6`;
4. run genuine Standard EvalMod on both Standard-native and reconstructed-Fast semantic inputs;
5. define:

`amplification = output_delta / input_delta`

using max-component metrics when input delta is non-zero;
6. define empirical local C2S budget:

`budget = 1e-2 / amplification`.

Record the measured values.

Historical expected values are approximately:

- amplification `≈512`;
- budget `≈1.953125e-5`.

Do not hard-code them as universal constants; derive them from the current run.

If the sensitivity reconstruction is not trustworthy, classify:

`logn13_c2s_precision_budget_construction_failure`.

---

# C4 — replay C2S factor groups stage by stage

The production Fast C2S source applies, for each factor group:

1. one `LinearTransform` for this LogN13 `[1,1,1,1]` factorization;
2. exactly one `Rescale` after the group.

The genuine Standard DFT follows the same factor-group schedule.

Create Primary-only diagnostic replay helpers which use the **actual production matrices** and the actual source-defined arithmetic paths.

For group `g = 0..3`, capture immutable checkpoints:

- `group_g_input`;
- `group_g_after_linear_transform`;
- `group_g_after_rescale`.

For both Fast and Standard record compactly:

- Level;
- Scale / log2 Scale;
- LogDimensions;
- NTT/Montgomery flags;
- maintained q0/q1 capacity evidence where centered reconstruction is used;
- Fast-vs-Standard semantic max-component / max-complex / mean-complex;
- cumulative error relative to the common ModUp semantic input where meaningful.

Do not store slot arrays, coefficient arrays, or operation traces.

---

# C5 — representation-safe comparison

For every high-level Fast checkpoint, use the established maintained q0/q1 projection discipline:

- copy only q0/q1;
- IMForm copied maintained limbs when required;
- never read dormant limbs as authoritative.

For Standard checkpoints, compare decoded semantics using the genuine Standard secret and full-RNS ciphertext, and additionally provide a q0/q1-projected semantic view only when centered reconstruction is valid.

Do **not** declare a Standard projection oracle valid merely because its reduced residues exist.

If a Standard checkpoint exceeds q0/q1 centered uniqueness, retain the full-RNS semantic decode as the oracle and mark q0/q1 exact-row comparison unavailable.

---

# C6 — identify the first budget breach

For each checkpoint compute Fast-vs-Standard semantic max-component error.

The first checkpoint whose cumulative error exceeds the derived C2S budget is the authoritative first C2S precision breach.

Classify one of:

- `logn13_c2s_precision_first_breach_group0_linear`
- `logn13_c2s_precision_first_breach_group0_rescale`
- `logn13_c2s_precision_first_breach_group1_linear`
- `logn13_c2s_precision_first_breach_group1_rescale`
- `logn13_c2s_precision_first_breach_group2_linear`
- `logn13_c2s_precision_first_breach_group2_rescale`
- `logn13_c2s_precision_first_breach_group3_linear`
- `logn13_c2s_precision_first_breach_group3_rescale`

If all internal checkpoints stay within budget but final real/imag split causes the breach, use:

- `logn13_c2s_precision_first_breach_real_imag_split`.

If final Fast C2S is within the derived budget, use:

- `logn13_c2s_precision_budget_satisfied`.

Record first failing checkpoint or `none`.

---

# C7 — incremental error attribution

At the first breached group, quantify separately:

- error before the group's LinearTransform;
- error immediately after LinearTransform;
- error immediately after group Rescale;
- incremental error introduced by LinearTransform;
- incremental error introduced by Rescale.

Also record source-relevant evidence:

- matrix scale for the group;
- dropped logical modulus for the Rescale;
- Fast input/output Scale;
- Standard input/output Scale;
- whether Fast and Standard pre-Rescale semantics were already outside budget;
- whether capacity uniqueness is valid before Fast Rescale.

The result must make the next task unambiguous:

- LinearTransform-dominant => next task investigates Fast q0/q1 LinearTransform precision / rotation / scalar accumulation.
- Rescale-dominant => next task investigates Fast DFT group Rescale rounding / divisor reconstruction.
- final split-dominant => next task investigates conjugation / real-imag post-processing.

Do not implement the fix yet.

---

# C8 — verify causal sufficiency

Using the final Fast C2S real semantic output and genuine Standard EvalMod, reproduce that the C2S error alone is sufficient to violate the final `1e-2` budget.

Record:

- C2S real input delta;
- Standard-native vs Standard-on-Fast EvalMod output delta;
- amplification ratio;
- derived budget;
- `C2S_ERROR_ALONE_SUFFICIENT_FOR_FINAL_FAILURE = true/false`.

Expected from predecessor: true.

This is required so the task does not regress to the incorrect conclusion that C2S is acceptable merely because `3e-4 < 1e-2`.

---

# Artifact

Create one compact artifact only:

`results/FIX-001-P3-DIAG-LOGN13-C2S-PRECISION-BUDGET-summary.json`

Include:

- provenance;
- genuine Standard proof;
- ModUp precondition;
- derived amplification and C2S budget;
- group-by-group compact checkpoint metrics through the first budget breach;
- incremental error attribution at first breach;
- causal-sufficiency result;
- classification;
- validation/scope confirmations.

No full vectors, coefficient arrays, traces, or large repeated hash sets.

---

# Scope lock

Do not:

- modify Secondary production code;
- modify C2S matrices or parameters;
- modify workload;
- implement a precision fix;
- modify Fast EvalMod / polynomial / DoubleAngle code;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- resume unpack/finalization diagnostics.

---

# Validation

Before completion require:

- focused `TestFIX001P3.*C2S.*Precision` tests pass;
- Primary `go test ./...` passes;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- Primary support code/artifact committed and pushed;
- Primary worktree clean and remote synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003.