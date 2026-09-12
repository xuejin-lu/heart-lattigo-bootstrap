# FIX-001-P3-DIAG-LOGN13-EVALMOD-CAUSAL-LOCALIZE

## Purpose

Stage-wise semantic bisect established the first threshold failure at `EvalMod(real)`:

- Primary predecessor: `da9db6c8efa9fec30243241bbb58c0da3afe18f5`
- Secondary exact clean: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- genuine Standard E2E: `max_component_abs = 5.754406148805735e-8`
- Fast-vs-Standard C2S real: `3.098126393859551e-4` (passes 1e-2)
- Fast-vs-Standard C2S imag: `2.6367655745716595e-4` (passes 1e-2)
- Fast-vs-Standard EvalMod real: `0.12614267221866593` (fails 1e-2)

This does **not yet prove** the Fast EvalMod implementation is the root cause. Mod1 is nonlinear and can amplify a small upstream C2S semantic difference. This task first separates:

1. **C2S-input sensitivity amplified by Standard EvalMod**, versus
2. **Fast EvalMod implementation divergence on matched semantic input**.

Only if (2) is established should the task continue inside EvalMod to localize the first failing internal checkpoint.

Diagnostic only. No Secondary production modification.

---

# Fixed provenance and scope

Primary required base:

`da9db6c8efa9fec30243241bbb58c0da3afe18f5`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration remains `configs/bootstrap_config.logN13.json`.

Do not:

- modify Secondary;
- implement a fix;
- change workload or parameters;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- revisit unpack/finalization.

Threshold for causal classification remains `1e-2`.

---

# C0 — reproduce predecessor boundary

Reproduce the genuine-Standard/Fast production prefix only through `CoeffsToSlots` for the real branch using the same constructor and workload accepted in the predecessor.

Require:

- genuine Standard proof still holds (`Evaluator`, DFT, Mod1 all non-nil; Fast compatibility path not used);
- Standard E2E control still passes <=1e-2;
- Fast-vs-Standard C2S-real metric remains approximately the predecessor result and below 1e-2;
- native production Fast EvalMod(real) vs native Standard EvalMod(real) reproduces failure >1e-2.

If not:

`logn13_evalmod_causal_precondition_mismatch`.

Keep the original native Fast and native Standard C2S-real ciphertexts immutable for later controls.

---

# C1 — build semantic matched-input constructors

The comparison must match **decoded Mod1 input values**, not ciphertext randomness or secret-key representation.

## C1-A Standard ciphertext from Fast C2S semantic values

1. Decode the native Fast C2S-real input through the already accepted q0/q1 zero-secret projection helper.
2. Using bootstrap CKKS parameters, encode those complex values at:
   - Level = native C2S input Level (expected 12),
   - Scale = native C2S input Scale (expected 2^50),
   - same LogDimensions.
3. Encrypt under the genuine Standard bootstrap secret `skN2` using ordinary Standard encryption.
4. Decode immediately with `skN2` and compare against the source Fast-decoded vector.

Require reconstruction error <= `1e-6` max component. Record it.

Call this ciphertext:

`standard_from_fast_semantic_input`.

## C1-B Fast zero-secret ciphertext from Standard C2S semantic values

1. Decode the native genuine Standard C2S-real input under `skN2`.
2. Encode those values with bootstrap parameters at the same Level/Scale/LogDimensions.
3. Construct a Fast degree-one ciphertext whose maintained semantic state is:
   - c0 = encoded plaintext q0/q1 rows;
   - c1 = zero q0/q1 rows;
   - NTT = true;
   - Montgomery = true after applying MForm to maintained q0/q1 rows;
   - only q0/q1 authoritative.
4. Decode its non-Montgomery q0/q1 projection with zero secret and compare against the source Standard-decoded vector.

Require reconstruction error <= `1e-6` max component. Record it.

Call this ciphertext:

`fast_from_standard_semantic_input`.

If either semantic reconstruction cannot meet the bound:

`logn13_evalmod_matched_input_construction_failure`.

Do not use full dormant Fast limbs as authoritative.

---

# C2 — four EvalMod executions

Run exactly these four independent calls:

A. `STD_NATIVE = Standard EvalMod(native Standard C2S real)`

B. `FAST_NATIVE = Fast EvalMod(native Fast C2S real)`

C. `STD_ON_FAST = Standard EvalMod(standard_from_fast_semantic_input)`

D. `FAST_ON_STD = Fast EvalMod(fast_from_standard_semantic_input)`

Record compact metadata and semantic output vectors only through aggregate comparison metrics; do not store full slots.

Required comparisons:

1. `FAST_NATIVE vs STD_NATIVE` — reproduce predecessor divergence.
2. `FAST_NATIVE vs STD_ON_FAST` — **same Fast semantic input**, different implementation.
3. `STD_NATIVE vs FAST_ON_STD` — **same Standard semantic input**, different implementation.
4. `STD_NATIVE vs STD_ON_FAST` — Standard-only measurement of how much the original C2S semantic delta is amplified.
5. `FAST_NATIVE vs FAST_ON_STD` — Fast-only input sensitivity context.

---

# C3 — causal classification gate

Use the two matched-input implementation comparisons as the primary gate:

- `FAST_NATIVE vs STD_ON_FAST`
- `STD_NATIVE vs FAST_ON_STD`

## Case A — upstream C2S sensitivity explains the divergence

If both matched-input implementation comparisons pass <=1e-2, while:

- native `FAST_NATIVE vs STD_NATIVE` fails >1e-2, and
- `STD_NATIVE vs STD_ON_FAST` also fails >1e-2,

classify:

`logn13_evalmod_divergence_explained_by_c2s_input_sensitivity`

Stop. Do not enter internal EvalMod replay.

Record amplification ratio where denominator is predecessor C2S-real max-component delta and numerator is Standard native-vs-STD_ON_FAST max-component delta.

This classification means the next task belongs to C2S accuracy/capacity/rounding, even though the C2S delta itself was below the global 1e-2 threshold.

## Case B — Fast EvalMod implementation divergence survives matched input

If either matched-input implementation comparison fails >1e-2, classify provisionally:

`fast_evalmod_matched_input_divergence`

and continue to C4 using the **Standard semantic input pair**:

- Standard side: native genuine Standard C2S real;
- Fast side: `fast_from_standard_semantic_input`.

This removes upstream C2S semantic difference from internal localization.

If the two matched-input directions conflict materially (one passes, one fails), continue C4 but record:

`MATCHED_INPUT_ASYMMETRY = true`

and preserve both metrics.

---

# C4 — internal EvalMod localization on matched Standard semantics

Reproduce the source-defined Mod1 steps manually in Primary diagnostics only. Do not modify production evaluators.

Use:

- genuine Standard Mod1 parameters/evaluator components;
- current production Fast Mod1 parameters/evaluator components;
- Standard native C2S-real input;
- Fast reconstructed zero-secret ciphertext encoding the **same Standard-decoded C2S-real values**.

Derive all scales/constants from source/parameters; do not hardcode except the already production-defined LogN13 plan scale 2^91 when invoking the same Fast polynomial plan.

## C4.0 normalize + cosine offset

For both sides:

- copy input;
- save original input Scale;
- set Scale to `Mod1Parameters.ScalingFactor()`;
- apply the exact source cosine offset.

Compare decoded semantics.

If first failure >1e-2:

`logn13_evalmod_first_internal_divergence_normalize_offset`.

## C4.1 polynomial output before normalized metadata promotion

Standard:

- clone/source-equivalent Mod1 polynomial for `EvaluateNew(..., scaling=1)`;
- call genuine Standard `PolynomialEvaluator.Evaluate` with source-derived targetScale.

Fast:

- call production-equivalent `PolynomialEvaluator.EvaluateWithPlanScale` with:
  - same Mod1 polynomial;
  - same targetScale;
  - plan scale = 2^91.

Important: compare the decoded polynomial values **before** Fast metadata-only K promotion. Different ciphertext Scale values are allowed; decoded polynomial semantics should represent the same y value.

Record Level, Scale, metadata, q0/q1 capacity context, and semantic metric.

If first failure >1e-2:

`logn13_evalmod_first_internal_divergence_polynomial`.

## C4.2 normalized initial state z0

Derive `K0` from the actual production scale ratio exactly as Fast source does.

Production expectation is currently exponent 29, but derive first and record.

Fast source performs metadata-only promotion so decoded Fast value becomes `z0 = y0 / K0`.

Do not mutate the Standard ciphertext merely to force equal Scale. Compare numerically:

`FastDecoded(z0) * K0` vs `StandardDecoded(y0)`.

If first failure >1e-2:

`logn13_evalmod_first_internal_divergence_initial_normalization`.

## C4.3 DoubleAngle rounds 0..2

Replay source operations independently on both sides.

Standard recurrence:

`y_{i+1} = 2*y_i^2 - c_i`, followed by Standard Rescale.

Fast normalized recurrence:

`z_{i+1} = A_i*z_i^2 - C_i`, followed by Fast Rescale,

where derive from actual scales:

- `A_i = 2*K_i^2/K_{i+1}`
- `C_i = c_i/K_{i+1}`.

At each round record compact semantic checkpoints:

1. **square**: compare `Fast(z_i^2) * K_i^2` vs Standard `y_i^2`;
2. **multiplier/doubling**: compare Fast after A multiplication times `K_{i+1}` vs Standard after doubling;
3. **constant**: compare Fast after normalized constant times `K_{i+1}` vs Standard after Standard constant;
4. **post-rescale**: compare Fast decoded `z_{i+1} * K_{i+1}` vs Standard decoded `y_{i+1}`.

Also record full-RNS Standard coefficient-domain max magnitude before each Fast-sensitive multiplication/rescale where feasible, Q01/2, ratio and outside count. Capacity failure is independently meaningful.

First-failure classifications:

- `logn13_evalmod_first_internal_divergence_da_round0_square`
- `logn13_evalmod_first_internal_divergence_da_round0_multiplier`
- `logn13_evalmod_first_internal_divergence_da_round0_constant`
- `logn13_evalmod_first_internal_divergence_da_round0_rescale`
- same four forms for round1 and round2.

If exact-capacity evidence shows an operand exceeds centered Q01 uniqueness before a semantic mismatch, classify instead:

`logn13_evalmod_internal_q01_capacity_failure`

and record first round/checkpoint.

## C4.4 final K restore and input-scale reset

After round2, derive final K exactly as production source.

Fast physically multiplies by K and then resets metadata Scale to original input Scale. Standard source resets Scale to original input Scale after ordinary DoubleAngle.

Compare:

- immediately after Fast physical K restore vs Standard final recurrence semantic value;
- after both sides' final Scale reset.

Classifications:

- `logn13_evalmod_first_internal_divergence_final_restore`
- `logn13_evalmod_first_internal_divergence_final_scale_reset`.

If all internal matched-input checkpoints pass <=1e-2 but public Fast/Standard matched-input EvalMod still fails, classify:

`logn13_evalmod_internal_replay_oracle_mismatch`.

---

# C5 — interpretation

A success of this diagnostic is **not** a passing bootstrap. It is a supported causal localization.

Exactly one final classification must be emitted from:

- `logn13_evalmod_causal_precondition_mismatch`
- `logn13_evalmod_matched_input_construction_failure`
- `logn13_evalmod_divergence_explained_by_c2s_input_sensitivity`
- `logn13_evalmod_first_internal_divergence_normalize_offset`
- `logn13_evalmod_first_internal_divergence_polynomial`
- `logn13_evalmod_first_internal_divergence_initial_normalization`
- `logn13_evalmod_first_internal_divergence_da_round0_square`
- `logn13_evalmod_first_internal_divergence_da_round0_multiplier`
- `logn13_evalmod_first_internal_divergence_da_round0_constant`
- `logn13_evalmod_first_internal_divergence_da_round0_rescale`
- same four `da_round1_*`
- same four `da_round2_*`
- `logn13_evalmod_internal_q01_capacity_failure`
- `logn13_evalmod_first_internal_divergence_final_restore`
- `logn13_evalmod_first_internal_divergence_final_scale_reset`
- `logn13_evalmod_internal_replay_oracle_mismatch`

Also record:

- provisional matched-input verdict;
- first failing internal checkpoint or `none`;
- whether matched-input asymmetry occurred.

---

# Artifact

Create only:

`results/FIX-001-P3-DIAG-LOGN13-EVALMOD-CAUSAL-LOCALIZE-summary.json`

Keep compact and human-reviewable:

- provenance;
- predecessor reproduction;
- matched-input reconstruction errors;
- five C2 comparison metrics;
- causal verdict;
- internal checkpoints only if C4 was entered;
- capacity ratios/outside counts only where relevant;
- classification;
- tests/scope confirmations.

No slot arrays, coefficient arrays, full traces, or large repeated index lists.

---

# Validation

Require before completion:

- focused `TestFIX001P3.*EvalMod.*Causal` test passes;
- Primary `go test ./...` passes;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary changes;
- no production fix;
- Primary artifact/supporting diagnostic committed and pushed;
- Primary worktree clean and remote synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003.