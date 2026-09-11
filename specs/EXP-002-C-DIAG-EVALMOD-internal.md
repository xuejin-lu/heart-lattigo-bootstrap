# EXP-002-C-DIAG-EVALMOD — Localize the first failing operation inside Fast EvalMod(real)

## Purpose

EXP-002-C-DIAG-Q01 established, under the formal LogN13 workload, that:

- Q1 `ModUp`: Standard q0/q1 vs Fast q0/q1 matches exactly;
- Q2 `CoeffsToSlots` real/imag: both pass the fixed `1e-2` component threshold;
- Q3 `EvalMod(real)`: first supported divergence, max component error about `0.120238`;
- therefore the first supported failing Bootstrap stage is `EvalMod(real)`.

This task must localize the **first failing operation inside Fast EvalMod(real)**.

Do not repair the bug in this task.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected primary remote base when this spec is authored:

`f26eb5879579110a8ee43d41c932719c608d25aa`

Fast Lattigo baseline:

`ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Standard reference backend remains:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md`. Read `docs/FAST_CKKS_SPEC.md` before touching Secondary.

---

## Repository policy for this task

The production Fast implementation must not be modified.

A **temporary diagnostic-only Go test file** may be created in the Secondary Lattigo worktree solely to access the public/exported Fast Mod1 evaluator components and reproduce the formal LogN13 Q2 real branch.

Explicit exception to the default Secondary commit/push rule for this task:

- do **not** commit or push the temporary Secondary diagnostic file;
- after raw evidence has been written to `/tmp` and copied into Primary, delete only the temporary file created by this task;
- verify Secondary returns to the exact clean `fast-ckks` commit `ce79b861...`;
- never reset/stash/discard unrelated or pre-existing work;
- if Secondary is not clean at startup, stop.

Primary may commit the diagnostic result JSON/summary and any backend-neutral result-analysis helper required to serialize results. Do not move Fast implementation logic into Primary.

---

## Fixed workload

Reproduce exactly the existing formal LogN13 workload through Q2 real:

- `LogN = 13`
- full 4096 slots
- same parameters as `configs/bootstrap_config.logN13.json`
- `LogMessageRatio = 10`
- `K = 16`
- `Mod1Degree = 30`
- `DoubleAngle = 3`
- `Mod1InvDegree = 0`
- same deterministic application input
- same `ModUpThenEncode` circuit order

Run the Fast path only until the real branch immediately after `CoeffsToSlots` (Q2 real), then manually reproduce the exact operations of `mod1.FastEvaluator.EvaluateNew` on independent copies/checkpoints.

Do not run LogN16.

---

## Source of truth for operation order

Use the current Fast implementation at:

`circuits/ckks/mod1/fast.go`

and compare its intended operation sequence against the pinned Standard implementation:

`circuits/ckks/mod1/mod1_evaluator.go`

Do not infer operation order from memory.

At the pinned Fast baseline the expected high-level sequence is:

1. clone Q2-real input;
2. save `inputScale`;
3. scale reinterpretation: `res.Scale = mod1Params.ScalingFactor()`;
4. compute `targetScale`;
5. cosine/Chebyshev offset `Add`;
6. polynomial evaluation;
7. for each DoubleAngle iteration:
   - `MulRelin(res,res,res)`;
   - `Add(res,res,res)`;
   - scalar offset `Add(res,-sqrt2pi,res)`;
   - `Rescale(res,res)`;
8. restore `res.Scale = inputScale`.

Confirm this from source before proceeding.

---

## Diagnostic principle: local operation oracle

Do **not** compare every internal checkpoint directly to the original application input.

For each operation, decode the actual Fast q0/q1 state before the operation, apply the intended mathematical operation to those decoded values in plaintext, and compare that expected vector with the decoded actual q0/q1 state after the operation.

This prevents earlier CKKS approximation error from being blamed on a later operation.

Use the authoritative Fast zero secret.

For every Fast checkpoint:

- make an isolated q0/q1 diagnostic copy;
- IMForm q0/q1 on the copy only when needed;
- preserve the source Scale exactly;
- set diagnostic level to 1 as in EXP-002-C-DIAG-Q01;
- ordinary decrypt + CKKS decode;
- prove source q0/q1 rows are unchanged by the diagnostic projection.

Do not read dormant q2+ rows.

---

## Parameter sanity before operation checks

Record the exact effective Fast Mod1 parameters used by the formal run:

- LevelQ
- LogDefaultScale / ScalingFactor
- MessageRatio
- QDiff
- K
- DoubleAngle
- IntervalShrinkFactor
- Sqrt2Pi
- Mod1 polynomial degree/depth/basis
- Mod1 polynomial interval A/B
- a stable hash of all non-nil Mod1 polynomial coefficients
- target-scale Q indices used by each DoubleAngle planning step

Compare these to the corresponding Standard formal-run parameter metadata wherever accessible without altering Standard source.

If a meaningful parameter differs before arithmetic begins, classify `MOD1_PARAMETER_MISMATCH` and stop arithmetic causal claims.

---

## Required internal checkpoints

Use only the **Q2 real** branch for causal localization in this task.

### E0 — `q2_real_input`

Record/decode the Fast Q2-real input before Mod1.

This must be consistent with the Q2-real evidence from EXP-002-C-DIAG-Q01 within the fixed threshold.

### E1 — `scale_reinterpretation`

After changing only metadata:

`res.Scale = mod1Params.ScalingFactor()`

Expected plaintext values are obtained from E0 using the exact ratio implied by the old vs new scale, because coefficients are unchanged.

Check:

- q0/q1 rows remain bit-exact;
- only Scale metadata changes;
- decoded values match the scale-ratio expectation.

### E2 — `cosine_offset`

After the scalar offset `Add` used before the Chebyshev evaluation.

Compute the exact offset from current Mod1 parameters using the same formula as source.

Expected for each slot:

`E2_expected = E1_actual + offset`

Compare E2 actual vs expected.

### E3 — `polynomial_evaluation`

After `PolynomialEvaluator.Evaluate(...)`, before DoubleAngle.

Expected values must be computed by evaluating the exact current `Mod1Poly` in its declared Chebyshev basis on **E2 actual decoded values**.

Prefer an existing Lattigo/bignum polynomial plaintext-evaluation helper if one exists. Inspect source and use its semantics. Do not silently substitute monomial-basis Horner evaluation for a Chebyshev polynomial.

If no reliable source-backed plaintext polynomial evaluator can be used, stop with `POLYNOMIAL_PLAINTEXT_ORACLE_UNAVAILABLE` rather than inventing one silently.

### E4.x — DoubleAngle iteration 0

Split iteration 0 into four checkpoints:

- E4.1 after `MulRelin`: expected = previous_actual²
- E4.2 after doubling Add: expected = 2 * previous_actual
- E4.3 after scalar offset: expected = previous_actual - current `sqrt2pi`
- E4.4 after Rescale: logical expected = previous_actual (rescale should preserve decoded logical value within CKKS error while changing level/scale)

### E5.x — DoubleAngle iteration 1

Same four sub-checkpoints and local expected-value rules.

### E6.x — DoubleAngle iteration 2

Same four sub-checkpoints and local expected-value rules.

### E7 — `restore_input_scale_metadata`

After only:

`res.Scale = inputScale`

Record coefficient hashes and decoded effect implied by the metadata scale change.

This is diagnostic completeness; causal localization should normally have occurred earlier if Q3 is already wrong.

---

## Numerical metrics

For each local operation comparison record all 4096 slots and:

- max_abs_complex
- mean_abs_complex
- rmse_complex
- max_abs_real
- max_abs_imag
- max_component_abs
- max-error indices
- sample count

Fixed local pass threshold:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

Do not loosen it.

Also record source/output:

- Level
- Scale exact string + log2
- IsNTT
- IsMontgomery
- q0/q1 row hashes for c0/c1

---

## Classification

Select the first failed local operation in execution order.

Allowed `FIRST_SUPPORTED_CAUSE` values:

- `mod1_parameter_mismatch`
- `scale_reinterpretation`
- `cosine_offset_add`
- `polynomial_evaluation`
- `double_angle_0_mulrelin`
- `double_angle_0_doubling_add`
- `double_angle_0_scalar_offset`
- `double_angle_0_rescale`
- corresponding iteration 1 values
- corresponding iteration 2 values
- `restore_input_scale_metadata`
- `unresolved_evalmod_internal`

If E0 does not reproduce Q2-real evidence, classify `DIAGNOSTIC_INPUT_MISMATCH` and stop.

If a local plaintext oracle is unavailable or invalid before the first failure can be established, classify `unresolved_evalmod_internal` and state exactly where/why.

---

## Cross-check against whole Q3

After manually replaying all operations, compare the final manual output against ordinary public `EvalMod(real)` output.

They must match within the fixed threshold and have consistent Level/Scale metadata.

If not, classify `MANUAL_REPLAY_MISMATCH` and do not make a root-cause claim from the replay.

Also compare the ordinary Q3 output against the prior EXP-002-C-DIAG-Q01 Q3-real evidence to ensure the formal divergence is reproduced.

---

## Required artifacts

Write diagnostic raw output to `/tmp` first.

After Secondary has been restored clean, copy only result artifacts into Primary:

- `results/EXP-002C-DIAG-EVALMOD-logN13-fast.json`
- `results/EXP-002C-DIAG-EVALMOD-logN13-summary.json`

Summary must include:

- formal effective Mod1 parameters;
- all internal checkpoint pass/fail metrics;
- first supported cause;
- manual replay vs public EvalMod cross-check;
- confirmation that Secondary temporary diagnostic source was removed and Secondary HEAD/worktree returned to clean `ce79b861...`.

---

## Validation

Before finishing:

- run the focused temporary diagnostic successfully;
- remove only the temporary Secondary diagnostic file created by this task;
- verify Secondary `HEAD = ce79b861...` and worktree clean;
- run `go test ./...` on clean Fast baseline after removal;
- do not modify or commit Lattigo production source;
- do not run LogN16;
- do not start EXP-003.

Primary may commit/push its result artifacts and mark the task COMPLETE.

---

## Non-goals

Do not:

- fix Fast Mod1 yet;
- change polynomial evaluator, FastCKKS arithmetic, Rescale, MulRelin, or parameters;
- perform a LogMessageRatio/K/degree/DoubleAngle sweep;
- benchmark performance;
- run LogN16;
- start EXP-003.

The only deliverable is the **first failing operation inside EvalMod(real)**.