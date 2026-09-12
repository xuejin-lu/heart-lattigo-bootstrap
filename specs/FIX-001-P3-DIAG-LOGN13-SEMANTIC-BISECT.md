# FIX-001-P3-DIAG-LOGN13-SEMANTIC-BISECT

## Purpose

The public LogN13 Fast Bootstrap boundary is structurally correct but fails end-to-end plaintext semantics.

Accepted predecessor:

- Primary: `47150c02b7d82e3f766c853e405c2d5cc445d45d`
- Secondary `fast-ckks`: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- classification: `logn13_e2e_semantic_failure`
- first failing checkpoint: `E5: public_bootstrap_plaintext_semantics`
- public Bootstrap rows exactly match the independently constructed unpack/finalization oracle
- one-element BootstrapMany exactly matches Bootstrap
- end-to-end `max_component_abs = 0.21536374987425272` > `1e-2`

Therefore the remaining problem is not unpack/finalization. The goal of this task is to find the **first production stage whose decoded semantics diverge from an independent full-RNS Standard control**.

This is a diagnostic-only Primary task. Do not modify Secondary production code.

---

# Fixed provenance

Primary required base:

`47150c02b7d82e3f766c853e405c2d5cc445d45d`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- `configs/bootstrap_config.logN13.json`
- LogN = 13
- LogSlots = 12
- LogDefaultScale = 45
- residual PrecisionMode = `PREC64`
- Residual MaxLevel = 1
- CircuitOrder = `ModUpThenEncode`
- EphemeralSecretWeight = 0
- N1 = N2 = 8192
- threshold = `1e-2`

---

# D0 — explicitly dispose of the errScale hypothesis

Source facts:

- Fast ScaleDown computes `errScale` with the same definition as Standard ScaleDown.
- Standard `Evaluator.Evaluate` applies the `diffScale = input.Scale / (output.Scale * errScale)` physical correction only when iterative bootstrapping is enabled or residual PrecisionMode is `PREC128`.
- This accepted LogN13 profile has LogDefaultScale 45 and therefore `PREC64`; IterationsParameters is nil.

Record:

- residual precision mode;
- iterations parameters present/absent;
- Fast ScaleDown `errScale`;
- Standard ScaleDown `errScale`;
- `STANDARD_DIFFSCALE_CORRECTION_APPLICABLE = false`.

Do **not** apply the PREC128/iterative correction in this task.

---

# D1 — construct a genuine full-RNS Standard control

Do not call `btp.GenEvaluationKeys(skN1)` blindly. On the current `fast-ckks` branch that API has a compatibility shortcut which can return `fastCompatible` keys for exactly this kind of profile, causing `bootstrapping.NewEvaluator` to dispatch back into Fast.

Instead construct Standard evaluation keys in Primary using the source-defined ordinary Standard path, without setting the private Fast compatibility marker.

For this exact N1=N2, EphemeralSecretWeight=0 profile:

1. Generate `skN1` under residual parameters.
2. Extend the same small secret to the full bootstrap Q/P basis exactly as ordinary `GenEvaluationKeys` does when N1=N2:
   - allocate `skN2 := rlwe.NewSecretKey(paramsN2)`;
   - use `rlwe.ExtendBasisSmallNormAndCenterNTTMontgomery` to extend residual secret Q into bootstrap Q and P.
3. Generate under `skN2`:
   - relinearization key;
   - all `btp.GaloisElements(paramsN2)` keys;
   - complex-conjugation Galois key.
4. Build a public `bootstrapping.EvaluationKeys` with the resulting `MemEvaluationKeySet`; N1/N2 ring-switch keys are nil because N1=N2; encapsulation keys are nil because EphemeralSecretWeight=0.
5. Call `bootstrapping.NewEvaluator(btp, standardKeys)`.

Prove this is not the Fast compatibility path:

- embedded Standard CKKS evaluator is non-nil;
- `DFTEvaluator` is non-nil;
- Standard `Mod1Evaluator` is non-nil;
- ordinary Standard stage calls are used.

If a genuine Standard evaluator cannot be constructed, classify:

`logn13_semantic_bisect_standard_control_construction_failure`.

---

# D2 — Standard end-to-end control must pass first

Before stage comparison, run the genuine Standard evaluator on an independent copy of the same `reproducibleInput.v1`.

The input is the existing deterministic direct CKKS encoding used by Primary; do not change the workload.

Decode the Standard public output under the proper residual secret and compare against `reproducibleValues`.

Require:

- Standard Bootstrap succeeds;
- final Level = residual MaxLevel;
- final Scale = residual default Scale;
- `max_component_abs <= 1e-2`;
- no NaN/Inf.

If Standard itself fails the semantic threshold, stop and classify:

`logn13_semantic_bisect_standard_control_failure`.

Do not use a failing Standard run as an oracle.

---

# D3 — representation-safe semantic projection helper

Intermediate Fast ciphertexts may maintain only q0/q1 and may be Montgomery while higher limbs are stale/dormant. Never ordinary-decrypt a high-level Fast ciphertext directly.

Create a Primary-only diagnostic helper which produces a decode-safe maintained projection:

1. target level = `min(source.Level(), 1)`;
2. copy only maintained q0/q1 rows for every ciphertext component;
3. preserve Scale and LogDimensions;
4. preserve NTT state;
5. if source is Montgomery, IMForm only copied maintained rows and mark projected ciphertext non-Montgomery;
6. do not copy/read limbs >=2.

For the Standard side, project to the same maintained level before decoding so comparisons use the same q0/q1 modulus window.

Decode:

- Fast projection with the zero secret matching current Fast semantics;
- Standard projection with the appropriate Standard secret (`skN1` at the residual boundary, extended `skN2` inside the bootstrap ring).

This helper is diagnostic only and must not alter production ciphertexts.

---

# D4 — run the same production stages side by side

Use independent copies of the exact same deterministic input.

Fast side: current production `FastEvaluator`.

Standard side: D1 genuine Standard evaluator.

Run in source order and record compact semantic checkpoints.

## D4.0 input

Verify both direct input views decode to `reproducibleValues` within `1e-2` before any bootstrap operation.

If not:

`logn13_semantic_bisect_input_precondition_failure`.

## D4.1 PackAndSwitchN1ToN2

Run both source-defined pack calls and retain their own packing contexts.

For this profile N1=N2 and one full-slot ciphertext, record that no ring-degree switch is expected, but verify semantics rather than assuming identity.

Compare:

- Fast vs Standard projected decoded semantics;
- each vs original plaintext where meaningful;
- Level / Scale / LogDimensions / representation metadata.

## D4.2 ScaleDown

Run Standard and Fast `ScaleDown` independently.

Record both returned `errScale` values and output metadata.

Compare:

- Fast projected semantics vs Standard projected semantics;
- Fast and Standard semantics vs original plaintext.

If this is the first failed checkpoint:

`logn13_semantic_first_divergence_scaledown`.

## D4.3 ModUp

Run production ModUp on both sides.

Compare q0/q1-projected decoded semantics and metadata.

If first failure:

`logn13_semantic_first_divergence_modup`.

## D4.4 CoeffsToSlots

Run production CoeffsToSlots on both sides using their respective evaluators and generated matrices.

Require both have the same real/imag branch structure for full packing.

Compare separately:

- Fast real vs Standard real decoded semantics;
- Fast imag vs Standard imag decoded semantics;
- Level / Scale / LogDimensions.

Do not compare C2S outputs directly to the original plaintext; C2S changes representation. Standard output is the semantic oracle here.

First-failure classifications:

- `logn13_semantic_first_divergence_c2s_real`
- `logn13_semantic_first_divergence_c2s_imag`.

## D4.5 EvalMod

Run production EvalMod independently on real and imag for both sides.

Compare Fast vs Standard projected decoded semantics separately for real and imag.

First-failure classifications:

- `logn13_semantic_first_divergence_evalmod_real`
- `logn13_semantic_first_divergence_evalmod_imag`.

Also record whether Fast production Mod1 still matches the previously accepted normalized q0/q1 evidence; this is compatibility evidence only and does not override a Standard semantic divergence.

## D4.6 SlotsToCoeffs / core output

Run production SlotsToCoeffs on both sides.

Compare:

- Fast vs Standard projected decoded semantics;
- Standard core semantics vs original plaintext;
- Fast core semantics vs original plaintext;
- Level / Scale / LogDimensions.

First-failure classification:

`logn13_semantic_first_divergence_s2c_core`.

Stop localization at the earliest failed stage. Later stages may be omitted once the first divergence is proven.

---

# D5 — comparison metric

For every semantic comparison record only compact aggregate metrics:

- max_component_abs;
- max_abs_complex;
- mean_abs_complex;
- worst index/component;
- threshold;
- pass/fail.

Use threshold `1e-2` for first-divergence classification.

Also record both sides' Level, Scale, LogDimensions, NTT/Montgomery flags and maintained level used for decoding.

Do not store full decoded slot vectors.

---

# D6 — interpretation rules

The first failing checkpoint is authoritative for the next task.

Examples:

- ScaleDown/ModUp failure => diagnose Fast scale/basis handling before DFT.
- C2S failure => next task localizes actual LogN13 Fast DFT factor/group semantics and q0/q1 capacity.
- EvalMod failure after C2S passes => next task revisits production normalized Mod1 against genuine Standard semantics rather than a Fast-derived local oracle.
- S2C failure after EvalMod passes => revisit S2C from genuine Standard semantic inputs.

If every internal Fast stage matches Standard within threshold but the prior public E5 still fails, classify:

`logn13_semantic_bisect_no_internal_divergence_oracle_mismatch`

and explicitly compare the genuine Standard final plaintext semantics with the prior E5 reference definition before any production fix.

---

# Required classifications

Choose exactly one:

- `logn13_semantic_bisect_standard_control_construction_failure`
- `logn13_semantic_bisect_standard_control_failure`
- `logn13_semantic_bisect_input_precondition_failure`
- `logn13_semantic_first_divergence_scaledown`
- `logn13_semantic_first_divergence_modup`
- `logn13_semantic_first_divergence_c2s_real`
- `logn13_semantic_first_divergence_c2s_imag`
- `logn13_semantic_first_divergence_evalmod_real`
- `logn13_semantic_first_divergence_evalmod_imag`
- `logn13_semantic_first_divergence_s2c_core`
- `logn13_semantic_bisect_no_internal_divergence_oracle_mismatch`

Record first failing checkpoint/stage, or `none` only for the final no-divergence classification.

---

# Artifact

Create one compact artifact:

`results/FIX-001-P3-DIAG-LOGN13-SEMANTIC-BISECT-summary.json`

Include only:

- provenance;
- proof Standard evaluator is genuinely Standard;
- Standard end-to-end control metric;
- PREC64 / errScale correction disposition;
- per-stage compact metadata + semantic metrics up to first divergence;
- classification;
- tests/scope confirmations.

No slot arrays, coefficient arrays, operation traces, or huge hashes.

---

# Scope lock

Do not:

- modify Secondary production code;
- modify parameters/workload;
- implement a fix in this task;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- use `btp.GenEvaluationKeys` as the Standard-control key constructor without proving it did not dispatch to Fast compatibility;
- treat an evaluator with nil Standard DFT/Mod1 fields as a Standard oracle.

---

# Validation

Before completion require:

- focused `TestFIX001P3...SemanticBisect` test passes;
- Primary `go test ./...` passes;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary changes;
- Primary artifact/support code committed and pushed;
- Primary worktree clean and remote synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003.