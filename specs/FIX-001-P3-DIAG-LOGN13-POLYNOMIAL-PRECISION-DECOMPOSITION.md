# FIX-001-P3-DIAG-LOGN13-POLYNOMIAL-PRECISION-DECOMPOSITION

## Purpose

Isolate the remaining ~`5e-7` LogN13 compressed-polynomial error into two independent causes:

1. semantic error already present in the generated Chebyshev powers supplied to Paterson-Stockmeyer (PS);
2. arithmetic/rounding error introduced by Fast PS baby/giant/final-Rescale evaluation itself.

Do not tune another scale schedule until this decomposition is measured.

---

## Fixed accepted provenance

Primary repository:

`xuejin-lu/heart-lattigo-bootstrap`

Required Primary base:

`571cd81854ec8b1834c33162cc40810d832d3cd3`

Secondary repository:

`xuejin-lu/lattigo`

Required exact Secondary commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

Follow both repositories' `AGENTS.md`, `CURRENT_TASK.md`, and Secondary `docs/FAST_CKKS_SPEC.md` where applicable.

This task is diagnostic only. No Secondary production modification is authorized.

---

## Accepted background

The following are already accepted and must not be re-litigated:

- production C2S compression `[4,2,0,0]` is correct;
- C2S Fast-vs-Standard real/imag are approximately `9e-14`;
- Fast normalized Mod1 implementation matches its valid normalized full-RNS mirror;
- Fast S2C implementation effect is zero against a valid stage-aligned full-RNS mirror;
- unpack/finalization effects are zero;
- current production EvalMod real/imag deltas are approximately `4.566e-3 / 4.435e-3`;
- S2C amplifies the EvalMod semantic delta into public error ~`0.246`;
- local EvalMod design target is `1e-4`;
- derived compressed-polynomial-vs-Standard design budget is approximately `1.2e-8`;
- common plan-scale sweep `2^91 ... 2^100` did not produce a precision-qualified candidate;
- guarded generated-power sweep `g=0...12` improved some low-degree powers but did not materially reduce the final polynomial/EvalMod error;
- best guarded polynomial error remained ~`5.19e-7` and no public-like candidate passed.

Current production PS plan scale remains exactly `2^91` in this task.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- change production power generation;
- change production PS plan scale;
- implement guarded powers;
- implement mixed/per-block PS scales;
- change polynomial coefficients or basis;
- change C2S or S2C;
- change Mod1 degree/K/DoubleAngle/parameters;
- change Q/P chains;
- add q2+ maintained arithmetic;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- integrate any hypothetical oracle path into production.

Primary diagnostic helpers/tests are allowed.

---

# D0 — reproduce exact baseline

Use deterministic `reproducibleInput.v1` and the actual production path through corrected C2S.

Require before interpretation:

- genuine Standard public Bootstrap passes `1e-2`;
- current Fast public Bootstrap reproduces failure near `0.246`;
- production `2^91` compressed polynomial real error vs genuine Standard reproduces near `5.46e-7`;
- imag reproduces near `5.30e-7`;
- production EvalMod real/imag reproduce near the accepted `4.5e-3` range;
- Secondary is exact/clean at required commit.

If not, stop with:

`logn13_polynomial_precision_decomposition_precondition_mismatch`.

---

# D1 — oracle sanity before decomposition

The previous diagnostic history has used source/plaintext Chebyshev oracles. Before using any plaintext/source oracle in this task, prove it is aligned to the **exact current corrected-C2S input and current polynomial semantics**.

For real and imag independently:

1. take the exact decoded input entering polynomial evaluation;
2. evaluate the source polynomial using the authoritative source-backed Chebyshev definition and interval/preprocessing semantics;
3. compare to genuine Standard polynomial evaluator output on the same input.

Require max-component discrepancy <= `1e-10` (prefer numerical roundoff much smaller).

If this check fails:

- do not use the source oracle for attribution;
- record the mismatch and stop with:

`logn13_polynomial_oracle_alignment_mismatch`.

Do not silently accept a source oracle that differs from genuine Standard by `1e-3` or `1e-2`.

---

# D2 — capture actual generated-power semantics

For each branch (real and imag), capture the actual production generated powers used by the degree-30 PS plan:

- T1
- T2
- T3
- T4
- T6
- T8
- T16

For each power record compactly:

- Level;
- exact Scale/log2;
- decoded semantic max-component error vs source-exact `T_n(x)`;
- centered q0/q1 capacity ratio;
- centered uniqueness;
- no full vectors in artifact.

This is measurement only; do not regenerate powers differently.

---

# D3 — construct four causal paths

Build the following four logically distinct paths for each branch.

## Path A — `H_actual`: ideal numerical PS fed by actual Fast power semantics

Decode the actual production Fast powers from D2.

Replay the **exact common/source PS decomposition and operation order numerically**, without CKKS rounding/modular arithmetic, using those decoded power values and the exact source coefficients.

Requirements:

- same baby blocks;
- same reversal order;
- same giant-step dependency structure;
- same polynomial coefficients;
- mathematical operations only;
- use enough numerical precision that replay error is <= `1e-12` against an independently evaluated equivalent expression where applicable.

This path answers:

> If PS arithmetic were perfect but it received the current imperfect power values, how wrong would the polynomial be?

Compare `H_actual` to the validated source/Standard polynomial reference.

Call this vector residual:

`POWER_SEMANTIC_EFFECT`.

## Path B — `H_oracle`: ideal numerical PS fed by exact source powers

Feed source-exact `T_n(x)` values into the same numerical PS replay.

Require `H_oracle` matches the D1 source/Standard polynomial reference <= `1e-10`.

This validates the numerical PS replay itself.

## Path C — `F_actual`: actual Fast PS fed by actual production powers

This is the production polynomial path at plan scale `2^91`.

Compare `F_actual` against `H_actual`.

Call the vector residual:

`FAST_PS_ARITHMETIC_EFFECT_ON_ACTUAL_POWERS`.

## Path D — `F_oracle`: actual Fast PS fed by oracle-encoded source-exact powers

Construct diagnostic-only zero-secret Fast ciphertext powers carrying source-exact `T_n(x)` semantics at the **same Level, Scale, degree, dimensions, NTT/Montgomery contract** as the corresponding production generated powers.

Preferred construction discipline:

1. encode source-exact power slot values into a valid full-RNS plaintext/ciphertext representation at the exact production power metadata;
2. establish correct domain/Montgomery representation from source code, not assumption;
3. project/copy authoritative q0/q1 rows into a degree-1 Fast diagnostic ciphertext;
4. set c1 consistently with zero-secret Fast semantics;
5. never copy stale dormant limbs and then use them as arithmetic authority;
6. verify decoded oracle-injected power error <= `1e-10`;
7. verify a stage-aligned full-RNS representation agrees on q0/q1 rows while centered-unique.

Then run the **actual unchanged Fast PS baby/giant/finalization arithmetic** at production plan scale `2^91` using those oracle powers.

Compare `F_oracle` against `H_oracle`.

Call the residual:

`FAST_PS_ARITHMETIC_FLOOR_WITH_ORACLE_POWERS`.

If a valid oracle power ciphertext cannot be represented under the production power metadata/q0q1 uniqueness contract, stop with:

`logn13_oracle_power_injection_capacity_or_domain_failure`.

---

# D4 — vector-level decomposition closure

Do not add max-error scalars and call that a decomposition. The maxima may occur at different slots.

For each slot/component, establish the vector identity:

`F_actual - Reference`

=

`(F_actual - H_actual)`

+

`(H_actual - H_oracle)`

+

`(H_oracle - Reference)`.

The last term should be numerical roundoff from D1/D3.

Record only compact norms in the artifact:

- total Fast polynomial residual;
- power semantic residual;
- Fast PS arithmetic residual on actual powers;
- oracle/reference residual;
- max decomposition-closure residual.

Require closure residual <= `1e-10`.

Also record:

`F_actual - F_oracle`

as the actual Fast-PS response to replacing all generated powers by oracle powers.

---

# D5 — stage-level PS arithmetic localization with oracle powers

Use Path D (`F_oracle`) to find whether Fast PS arithmetic itself can satisfy the `~1.2e-8` polynomial budget when power inputs are near-exact.

At minimum checkpoint the actual production PS operations:

- each baby block after constant injection;
- each baby coefficient `MulThenAdd` result;
- each giant-step optional relinearization;
- each giant-step pre-Rescale input;
- post-Rescale result;
- multiply by generated power;
- scale alignment;
- add-aligned result;
- final optional relinearization;
- final pre-Rescale state;
- final post-Rescale compressed polynomial.

For each checkpoint compute two different errors:

### Local arithmetic effect

Compare actual Fast operation output against the exact numerical operation applied to the **decoded actual inputs of that operation**.

This isolates rounding/primitive error introduced by that operation alone.

### Cumulative oracle-path effect

Compare actual Fast checkpoint semantics against the mathematically expected subtree/checkpoint value derived from oracle powers.

This shows accumulated PS arithmetic error.

Record Level/Scale and centered q0/q1 capacity at each checkpoint internally, but artifact should include only:

- first checkpoint where cumulative error exceeds `1.2e-8`;
- local error there;
- cumulative error there;
- worst PS capacity ratio/checkpoint;
- final oracle-power Fast PS error.

Do not dump vectors or every checkpoint to JSON.

---

# D6 — generated-power influence analysis

Only as a sensitivity diagnostic, not a proposed production path:

Starting from numerical Path A (`H_actual`), replace one generated power at a time with its exact source value while leaving the others at their actual decoded values:

- T2
- T3
- T4
- T6
- T8
- T16

Re-evaluate the ideal numerical PS expression and record the resulting final polynomial error vs reference.

For each power report:

`error_reduction = baseline_H_actual_error - one_power_replaced_error`.

This identifies which power errors materially propagate through the actual PS decomposition.

Because the power set is recurrence-correlated, do **not** interpret these one-at-a-time replacements as additive causal contributions.

Also evaluate:

- all powers actual;
- all powers oracle.

Optionally, if one or two powers clearly dominate, perform an actual Fast Path-D-style single-power oracle substitution as a confirmation, but do not expand scope into a combinatorial search.

---

# D7 — sufficiency intervention downstream

If `F_oracle` final compressed polynomial error vs genuine Standard is <= `1.2e-8`, feed that actual ciphertext result into the **unchanged existing normalized DoubleAngle → S2C → supported one-input unpack/finalization/public-like diagnostic path**.

Record:

- EvalMod real/imag error vs genuine Standard;
- post-S2C semantic difference;
- public-like max-component error vs deterministic message;
- metadata contract.

This answers whether fixing generated-power precision alone would be sufficient end-to-end while leaving Fast PS arithmetic unchanged.

If `F_oracle` does not meet polynomial budget, do not force downstream and do not claim generated powers alone are sufficient.

---

# D8 — classification logic

Let `B = 1.2e-8` be the current polynomial design budget.

Use the measured paths, not intuition.

## Case A — generated powers are the blocker

Conditions:

- `H_actual` error > B;
- `F_oracle` error <= B;
- decomposition closure passes.

Classify:

`logn13_polynomial_precision_generated_power_blocker`.

If D7 public-like also passes `1e-2`, record:

`GENERATED_POWER_FIX_SUFFICIENCY = confirmed`.

If D7 fails, record insufficiency explicitly.

## Case B — Fast PS arithmetic is the blocker

Conditions:

- `H_actual` error <= B;
- `F_oracle` error > B.

Classify:

`logn13_polynomial_precision_ps_arithmetic_blocker`.

Record the first D5 checkpoint exceeding budget.

## Case C — both independently exceed budget

Conditions:

- `H_actual` error > B;
- `F_oracle` error > B.

Classify:

`logn13_polynomial_precision_joint_power_and_ps_blocker`.

Record both magnitudes and first PS arithmetic checkpoint.

## Case D — neither isolated path explains the production residual

Conditions:

- both isolated paths <= B but `F_actual` remains > B; or
- vector decomposition does not close despite valid oracles.

Classify:

`logn13_polynomial_precision_interaction_not_isolated`.

Do not design a fix in this case.

Oracle/precondition failures use their explicit classifications above.

---

# Required artifact

Create one compact artifact:

`results/FIX-001-P3-DIAG-LOGN13-POLYNOMIAL-PRECISION-DECOMPOSITION-summary.json`

Include only:

- provenance;
- baseline reproduction;
- D1 oracle sanity metrics;
- actual power errors for real/imag;
- four-path final errors:
  - H_actual
  - H_oracle
  - F_actual
  - F_oracle;
- compact vector residual norms and decomposition closure;
- first oracle-power PS checkpoint exceeding budget, if any;
- worst PS capacity ratio/checkpoint;
- one-at-a-time power sensitivity table;
- D7 downstream intervention metrics if reached;
- classification;
- first supported blocker/checkpoint;
- validation flags.

Do not serialize:

- full slots;
- coefficient arrays;
- RNS rows;
- complete operation traces;
- repeated hashes;
- large per-index tables.

---

# Validation

Before completion require:

- focused tests matching `TestFIX001P3.*Polynomial.*Precision.*Decomposition` pass;
- Primary `go test ./...` passes;
- Secondary remains exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no production integration;
- no guarded-power production change;
- no mixed/per-block scale implementation;
- no PS plan-scale change;
- no C2S/S2C change;
- no Mod1/parameter retuning;
- no invalid source oracle accepted without D1 sanity;
- no invalid raw Standard DoubleAngle q0/q1 projection;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.