# FIX-001-P3-DIAG-LOGN13-EVALMOD-S2C-AMPLIFICATION

## Purpose

Diagnose the remaining LogN13 public Bootstrap semantic failure **after production C2S compression has been integrated and validated**.

Current accepted production state:

- Primary required base: `f3161a96b7b3d6d1250abc9e7ea834f4ca98e393`
- Secondary required base: `25e70430b10cd4af37ad4cf94cd994912e970e3f` on `fast-ckks`
- production Fast C2S restore plan: `[4,2,0,0]`
- production C2S real vs genuine Standard max-component error: `9.209497856162152e-14`
- production C2S imag vs genuine Standard max-component error: `9.444563729178734e-14`
- C2S budget: `1.9531249403614602e-5`, passed
- genuine Standard public Bootstrap passes
- Fast public Bootstrap metadata is correct but semantic max-component error is approximately `0.24628047174869752`, failing `1e-2`

Previously accepted matched-input EvalMod diagnostic showed:

- Fast production EvalMod `F` matches the normalized full-RNS stage-aligned mirror `N` exactly;
- `F` vs genuine Standard EvalMod `S` was about `4.57e-3` real / `4.43e-3` imag, below the **local** `1e-2` threshold;
- the largest causal term was compressed-polynomial/coherent-vs-Standard, around `4.59e-3`;
- raw genuine-Standard DoubleAngle q0/q1 projection is invalid after centered `Q01/2` capacity loss and must never be used as an exact oracle.

The new public E2E result proves that a local EvalMod threshold of `1e-2` is not sufficient evidence that EvalMod is harmless downstream. This task must determine whether the approximately `4.5e-3` EvalMod semantic difference is amplified by SlotsToCoeffs into the approximately `0.246` public error, or whether Fast S2C/public boundary introduces a separate error.

This is a **diagnostic-only** task. Do not change production behavior.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- retune C2S compression;
- retune EvalMod polynomial plan;
- modify S2C;
- modify packing/unpacking/finalization;
- change parameters;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- implement a fix even if the cause is identified.

Primary diagnostic helpers/tests are allowed.

---

# D0 — reproduce current production facts

Using the current production Secondary commit and deterministic `reproducibleInput.v1`, reproduce:

1. genuine Standard public Bootstrap semantic error <= `1e-2`;
2. production Fast C2S real/imag vs genuine Standard <= the derived C2S budget;
3. production Fast public Bootstrap semantic error > `1e-2`, expected near `0.246` max-component;
4. public metadata contract still correct:
   - residual Level correct;
   - Degree 1;
   - NTT true;
   - non-Montgomery public boundary;
   - Scale = residual DefaultScale;
   - input unchanged.

If these no longer reproduce, classify:

`logn13_post_c2s_public_failure_precondition_mismatch`.

---

# D1 — capture actual production EvalMod outputs

From the **actual production C2S outputs** produced by Secondary `25e704...`, capture real and imag at Mod1 input, then run actual production Fast EvalMod and genuine Standard EvalMod.

For each real/imag path record:

- input Level/Scale/Degree/domain;
- Fast EvalMod output Level/Scale/Degree/domain;
- genuine Standard EvalMod output Level/Scale/Degree/domain;
- Fast vs genuine Standard semantic max-component / max-complex / mean-complex errors;
- Fast output q0/q1 centered capacity;
- appropriate normalized full-RNS stage-aligned capacity and exact-row match evidence if reused.

Expected historical matched-input values are approximately:

- real `4.5663e-3`;
- imag `4.4350e-3`.

Do not fail the task merely because those are below `1e-2`; the purpose is downstream causal attribution.

Also record explicitly:

`LOCAL_EVALMOD_1E2_THRESHOLD_DISPOSITION = insufficient_for_e2e_causality`.

---

# D2 — Fast S2C implementation oracle on Fast EvalMod output

Construct a valid full-RNS stage-aligned reference from each Fast EvalMod output using the established centered-CRT q0/q1 lift discipline:

1. use authoritative maintained q0/q1 only;
2. correct NTT/INTT and Montgomery/IMForm handling before centered reconstruction;
3. require centered uniqueness before treating the lift as exact;
4. redistribute centered coefficients to all required Q limbs;
5. preserve Level, Scale, Degree, LogDimensions and semantic meaning;
6. never copy stale dormant Fast limbs.

Call the lifted real/imag pair `N_mod_real`, `N_mod_imag`.

Run:

- actual Fast `SlotsToCoeffs` on the original Fast EvalMod real/imag;
- a full-RNS stage-aligned S2C mirror on `N_mod_real` / `N_mod_imag` using the same mathematical S2C matrices, group schedule and ordinary Rescale sequence.

For each S2C group boundary and final S2C output record compactly:

- Level/Scale;
- Fast q0/q1 capacity;
- normalized/full-RNS capacity vs `Q01/2`;
- q0/q1 row equality Fast vs stage-aligned full-RNS reference while uniqueness holds;
- first row mismatch only;
- final semantic difference where meaningful.

If the full-RNS reference loses centered q0/q1 uniqueness, stop using exact-row projection after that point and classify accordingly rather than decoding an invalid oracle.

If the reference remains safe and Fast differs from it, classify the earliest supported cause:

`logn13_fast_s2c_primitive_or_rescale_mismatch`.

If Fast matches the stage-aligned full-RNS S2C reference through all groups, record:

`FAST_S2C_IMPLEMENTATION_EFFECT = zero_or_numerical_roundoff`.

---

# D3 — downstream semantic transfer of Fast EvalMod error

Now isolate the semantic effect of the Fast-vs-Standard EvalMod difference through the same S2C transform.

Construct two full-RNS semantic paths at the same Mod1 output metadata:

## H_F — hybrid path from Fast EvalMod semantics

Use decoded Fast EvalMod real/imag values to construct valid full-RNS ciphertext inputs at the same Level/Scale/layout suitable for ordinary full-RNS S2C evaluation. A fresh encoding/encryption under the genuine Standard diagnostic secret is allowed here because this path is a **semantic transfer oracle**, not an exact-row oracle.

Run ordinary genuine-Standard/full-RNS S2C on those Fast EvalMod semantics.

## H_S — genuine Standard path

Run ordinary genuine Standard S2C on the genuine Standard EvalMod real/imag outputs from the same matched C2S input.

Compare after S2C:

- `H_F` vs `H_S` semantic max-component / max-complex / mean-complex error;
- each against the expected original deterministic message semantics after the complete S2C/public representation is restored;
- the actual Fast S2C output vs `H_F` where representation permits a valid semantic comparison.

The critical question is whether `H_F` reproduces most of the approximately `0.246` public failure while Fast S2C itself matches its stage-aligned reference.

Record:

`evalmod_input_difference_real`

`evalmod_input_difference_imag`

`post_s2c_difference_due_to_evalmod_semantics`

and a descriptive empirical amplification ratio when denominators are nonzero:

`post_s2c_max_component_delta / max(real_evalmod_delta, imag_evalmod_delta)`.

This ratio is descriptive only; do not present it as a universal operator-norm guarantee.

If the hybrid path from Fast EvalMod semantics reproduces the public-scale failure while the Fast S2C implementation effect is negligible, classify:

`logn13_evalmod_semantic_difference_amplified_by_s2c`.

---

# D4 — alternative: S2C-specific semantic failure

If Fast S2C does **not** match the valid stage-aligned full-RNS S2C reference on the same Fast EvalMod input, distinguish:

- linear transform mismatch;
- Rescale mismatch;
- combination of real/imag (`imag * i + real`) mismatch;
- capacity alias at a specific group;
- final S2C metadata/representation mismatch.

Use the earliest valid checkpoint.

Do not blame EvalMod merely because its local output differs from Standard if an independent S2C mismatch is sufficient to explain the public failure.

Allowed specific classifications:

- `logn13_s2c_input_combine_mismatch`
- `logn13_s2c_group0_mismatch`
- `logn13_s2c_group1_mismatch`
- `logn13_s2c_group2_mismatch`
- `logn13_s2c_q01_capacity_alias`
- `logn13_s2c_final_representation_mismatch`

---

# D5 — unpack/finalization boundary check

Because the current LogN13 configuration has `N1=N2=8192` and one deterministic public input, packing/unpacking should not be assumed to cause the semantic failure, but verify it compactly.

Capture:

1. actual Fast core output immediately after S2C;
2. output after `UnpackAndSwitchN2ToN1`;
3. output after `finalizeFastPublicCiphertext`;
4. public `Bootstrap` output.

Record:

- Level/Scale/Degree/domain flags;
- q0/q1 rows or hashes where exact comparison is valid;
- semantic error against original message at each representation where decoding is valid;
- whether finalization changes physical rows;
- whether finalization only performs maintained-limb IMForm + metadata reset as expected;
- whether public output equals the manually finalized path.

If S2C output is already wrong and later boundaries preserve it, record boundary effect as zero.

If the first new error appears after S2C, classify:

- `logn13_unpack_semantic_mismatch`, or
- `logn13_finalization_semantic_mismatch`.

---

# D6 — causal accounting

The final summary must explicitly separate at least these effects:

1. `c2s_effect` — expected resolved / ~1e-13;
2. `evalmod_local_difference` — current Fast vs genuine Standard at Mod1 output;
3. `fast_s2c_implementation_effect` — Fast S2C vs stage-aligned full-RNS S2C on the same Fast EvalMod input;
4. `evalmod_difference_after_s2c` — hybrid Fast-EvalMod-semantics S2C vs genuine Standard EvalMod+S2C;
5. `unpack_effect`;
6. `finalization_effect`;
7. `total_public_fast_error`.

Do not infer causality from the local `1e-2` EvalMod threshold. The authoritative correctness criterion is the public E2E threshold `1e-2`.

If evidence supports it, state explicitly:

`PREVIOUS_EVALMOD_LOCAL_PASS_DISPOSITION = insufficient_local_threshold_for_downstream_e2e`.

---

# Required primary classification

Choose exactly one:

- `logn13_post_c2s_public_failure_precondition_mismatch`
- `logn13_evalmod_semantic_difference_amplified_by_s2c`
- `logn13_fast_s2c_primitive_or_rescale_mismatch`
- `logn13_s2c_input_combine_mismatch`
- `logn13_s2c_group0_mismatch`
- `logn13_s2c_group1_mismatch`
- `logn13_s2c_group2_mismatch`
- `logn13_s2c_q01_capacity_alias`
- `logn13_s2c_final_representation_mismatch`
- `logn13_unpack_semantic_mismatch`
- `logn13_finalization_semantic_mismatch`
- `logn13_post_c2s_public_failure_not_yet_isolated`

Record first failing stage/checkpoint or `none`.

---

# Artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-DIAG-LOGN13-EVALMOD-S2C-AMPLIFICATION-summary.json`

Include only:

- Primary/Secondary provenance;
- reproduced public failure;
- C2S resolved metrics;
- production EvalMod real/imag metrics;
- compact S2C stage-aligned capacity/row-match evidence;
- hybrid `H_F` / `H_S` post-S2C semantic metrics;
- empirical downstream amplification metric;
- unpack/finalization effect summaries;
- causal accounting;
- classification and validation flags.

Do not serialize full slot vectors, coefficient arrays, Q-limb dumps, matrix contents, or long traces.

---

# Validation

Before completion require:

- focused tests matching `TestFIX001P3.*EvalMod.*S2C.*Amplification` pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary modification;
- no production fix;
- no C2S retuning;
- no EvalMod retuning;
- no S2C modification;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no invalid raw Standard DoubleAngle q0/q1 oracle;
- compact artifact committed;
- Primary normally fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.