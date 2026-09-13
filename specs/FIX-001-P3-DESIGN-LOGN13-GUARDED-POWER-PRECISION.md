# FIX-001-P3-DESIGN-LOGN13-GUARDED-POWER-PRECISION

## Purpose

Determine whether the remaining LogN13 EvalMod/public Bootstrap precision failure can be solved by increasing **generated Chebyshev power precision** while preserving the existing polynomial/Mod1 scale contract.

This task follows two accepted results.

### Accepted downstream causal result

- Primary accepted base: `72cc6a44d4ce25b11f1e77cbd77d018f21e4aa68`
- Secondary production base: `25e70430b10cd4af37ad4cf94cd994912e970e3f`
- Fast C2S is resolved (~`9e-14`).
- Fast S2C implementation effect is zero against a valid stage-aligned full-RNS reference.
- unpack/finalization effects are zero.
- Fast EvalMod semantic delta (~`4.5e-3`) is amplified by S2C to public error ~`0.246`.
- local EvalMod design target: `1e-4`.
- derived polynomial delta budget for that target is approximately `1.2e-8`.

### Accepted common plan-scale sweep

Primary commit:

`7017ef8dc567ef13487f923770c5aa2d0414f296`

Classification:

`logn13_evalmod_common_scale_insufficient_precision`

Key result:

- `2^91` polynomial vs Standard ~`5.4574e-7`;
- `2^92` polynomial vs Standard ~`4.9546e-7`;
- the improvement is far too small;
- `2^93+` reaches invalid/capacity-loss behavior, with first limiting checkpoint recorded as `round0.after_multiplier_capacity`;
- no precision-qualified common plan-scale candidate exists.

Do not repeat that sweep.

### Older PS global-semantics evidence

At canonical `2^91`, generated Chebyshev powers already carry source-backed errors before PS accumulation, including approximately:

- `T2`: `5.75e-9`
- `T3`: `1.84e-7`
- `T4`: `2.72e-8`
- `T6`: `5.13e-8`
- `T8`: `1.08e-7`
- `T16`: `4.31e-7`

By contrast early baby-block coefficient accumulation errors were around `1e-12 ... 1e-10`.

Therefore the next supported hypothesis is:

> the current balanced pre-rescale generated-power schedule throws away too much fixed-point precision before multiplication; the resulting power error is then amplified by Chebyshev recurrence and dominates the final compressed polynomial precision.

This task tests a **guard-bit power multiplication** while keeping the generated-power output scale restored to the existing contract (~`2^60`).

No production code change is authorized.

---

# Core design idea

Current balanced power multiplication approximately does:

1. multiply left/right operands by balanced integer factors whose product is approximately the current dropped modulus `q`;
2. Rescale each operand by `q`;
3. multiply the two reduced-scale operands;
4. continue Chebyshev recurrence.

For ~`2^60` inputs and ~`2^60` q, the two post-Rescale operands are only around `2^30` scale, so the product returns to ~`2^60` but operand rounding precision is limited.

For a guard exponent `g >= 0`, test balanced factors satisfying approximately:

`f_left * f_right ~= q * 2^g`.

Then after pre-rescaling the operands, the multiplication result is approximately at:

`S_out * 2^g`

rather than `S_out`.

For Chebyshev recurrence:

`T_{a+b} = 2*T_a*T_b - T_|a-b|`.

Apply the Chebyshev doubling while still in the guarded high-scale domain, then perform a deterministic centered coefficient-domain rounded division by `2^g` on maintained q0/q1, with metadata Scale divided by the same factor, **without consuming a level**. Only after that restored-scale normalization apply the subtraction term.

Thus the public/generated-power Scale contract remains unchanged while multiplication gets additional guard precision.

This is a design experiment, not yet a production primitive.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- change polynomial coefficients;
- change production PS plan scale (`2^91` remains fixed in this experiment);
- change C2S or S2C;
- change Mod1 degree/K/DoubleAngle/parameters;
- change Q/P chains;
- add q2+ maintained arithmetic;
- permanently change generated-power output Scale;
- consume an extra Q level;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- integrate a successful design in this task.

Primary diagnostic/design helpers are allowed. Secondary must remain exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`.

---

# G0 — reproduce baseline

Using corrected production C2S and deterministic `reproducibleInput.v1`, reproduce at `g=0`:

- current generated powers and their source-backed errors;
- current compressed polynomial error ~`5.4574e-7` for real branch;
- current EvalMod real/imag errors ~`4.5663e-3` / `4.4350e-3`;
- current Fast public/public-like failure ~`0.246`;
- genuine Standard control passes.

If baseline does not reproduce, classify:

`logn13_guarded_power_precondition_mismatch`.

---

# G1 — candidate guard sweep

Test deterministic integer guard exponents:

`g = 0, 1, 2, ..., 12`.

`g=0` is the exact production-control schedule.

Do not assume all candidates are feasible. Reject only from measured/source-derived centered capacity evidence.

For each multiplication in actual generated-power construction, derive balanced positive integer factors approximating:

`f_left * f_right ~= q * 2^g`

with the same principles as the current near-square balancing:

1. minimize product error relative to the target product;
2. then minimize max(f_left, f_right);
3. then deterministic lexical tie-break.

Use high-precision integers for target derivation and record the chosen factors.

The candidate must retain existing generated-power Level topology and final restored Scale topology.

---

# G2 — exact guarded normalization primitive for diagnosis

Implement only in Primary diagnostic code the candidate operation:

`round_div_pow2_maintained(ct, g)`.

Requirements:

1. no level consumption;
2. apply to every maintained ciphertext component required by current Fast semantics;
3. convert q0/q1 NTT -> coefficient domain;
4. remove Montgomery representation correctly before centered CRT interpretation;
5. reconstruct each coefficient uniquely from q0/q1;
6. require `abs(x) < Q01/2` before division;
7. compute nearest-integer signed rounding of `x / 2^g` using exact integer arithmetic and a documented tie convention;
8. map the rounded signed result back to q0/q1 residues;
9. restore NTT and original Montgomery/public internal domain as appropriate;
10. set metadata Scale = prior Scale / `2^g`;
11. preserve Degree, Level, LogDimensions and all unrelated metadata.

For `g=0` the helper must be an exact no-op.

Do not use float64 coefficient conversion.

Build a full-RNS stage-aligned mirror of the same guarded operation and require exact q0/q1 row equality whenever centered uniqueness is valid.

---

# G3 — guarded generated-power replay

Replay the real Fast generated-power dependency graph used by the degree-30 Chebyshev PS plan.

At minimum cover all generated powers actually used:

- T1
- T2
- T3
- T4
- T6
- T8
- T16

For each nontrivial generated power and each candidate g, checkpoint:

1. left/right input;
2. after integer factor multiplication, before each pre-Rescale;
3. after each pre-Rescale;
4. raw guarded multiplication result;
5. after Chebyshev doubling in guarded domain;
6. after rounded `/2^g` normalization;
7. after recurrence subtraction/alignment;
8. final generated power.

At every valid checkpoint record compactly:

- Level;
- Scale/log2;
- exact full-RNS max centered coefficient;
- Q01/2;
- capacity ratio;
- outside count;
- q0/q1 centered uniqueness;
- Fast vs full-RNS q0/q1 row equality where exact;
- semantic error vs source-backed Chebyshev Tn oracle.

If a pre-Rescale scaled operand already exceeds centered Q01/2, the candidate fails there; do not let modular wrap continue and then claim a later result is meaningful.

If raw product/doubled product exceeds Q01/2, fail there before rounded normalization.

Per-candidate first-failure classes:

- `guard_factor_pre_rescale_capacity_failure`
- `guard_pre_rescale_primitive_mismatch`
- `guard_product_capacity_failure`
- `guard_round_div_primitive_mismatch`
- `guard_recurrence_alignment_failure`
- `guard_generated_power_semantic_failure`
- `guard_generated_powers_pass`.

---

# G4 — generated-power precision qualification

For each candidate that completes all powers, report max source-backed error for every Tn and especially T16.

Use these diagnostic guidance targets:

- preferred `T16 <= 2e-8`;
- all other generated powers <= `2e-8` where attainable.

These are guidance, not final acceptance criteria.

A candidate may proceed even if one power exceeds `2e-8`; the authoritative question is complete polynomial/EvalMod/public behavior.

Select all capacity-safe candidates for downstream replay; do not prematurely select only the smallest T16 error.

---

# G5 — polynomial replay with fixed production PS contract

For every capacity-safe guard candidate:

- use its guarded generated powers;
- keep production common PS plan scale exactly `2^91`;
- keep source polynomial, PS decomposition, baby/giant Levels, coefficients and operation ordering unchanged;
- replay real Fast baby/giant arithmetic;
- keep final compressed polynomial output Scale contract unchanged from current production.

Compare final compressed polynomial against:

1. authoritative plaintext/source polynomial oracle;
2. genuine Standard polynomial output from the same corrected C2S input;
3. a suitable stage-aligned full-RNS mirror of the guarded-power PS path.

Record:

- polynomial max-component error vs genuine Standard;
- Fast vs guarded full-RNS mirror effect;
- capacity headroom throughout PS.

Precision guidance target:

`compressed_polynomial_vs_standard <= 1.2e-8`.

Again, downstream actual correctness is authoritative.

If generated powers improve but polynomial remains near `5e-7`, record where PS introduces the new dominant error rather than assuming the guard design failed.

---

# G6 — unchanged normalized DoubleAngle and local EvalMod

Because guarded normalization restores every generated-power Scale to the existing production contract, keep the existing production normalized Mod1 schedule unchanged:

- polynomial plan scale `2^91`;
- compressed polynomial output ~`2^31`;
- K schedule `2^29`;
- normalized recurrence multiplier `2^30` where source-derived current production uses it;
- three DoubleAngle rounds;
- final physical+metadata restore;
- final public bootstrap DefaultScale metadata behavior unchanged.

Replay actual Fast-style normalized DoubleAngle using the guarded polynomial result.

Require the established exact-oracle discipline:

- Fast q0/q1 vs normalized full-RNS mirror exact where valid;
- no raw genuine-Standard DA q0/q1 projection after centered capacity loss.

Record real/imag final EvalMod vs genuine Standard.

Precision-qualified candidate target:

- real <= `1e-4`;
- imag <= `1e-4`.

---

# G7 — downstream S2C/public-like validation

For every precision-qualified candidate, run unchanged Fast S2C and the supported one-input unpack/finalization/public representation diagnostic path.

Record:

- Fast S2C implementation effect vs valid stage-aligned full-RNS mirror;
- post-S2C difference vs genuine Standard path;
- final public-like max-component error vs original `reproducibleInput.v1` message;
- Level/Scale/Degree/NTT/Montgomery contract.

Authoritative acceptance:

`public_like_max_component_error <= 1e-2`.

If possible with the diagnostic harness, also run an equivalent end-to-end candidate replay that matches the public `Bootstrap` semantic comparison exactly. Do not modify Secondary production just to obtain it.

---

# G8 — candidate choice

If one or more guard candidates pass public-like `1e-2`, choose deterministically:

1. smallest guard exponent `g` that passes;
2. then smallest public-like max-component error;
3. require all q0/q1 authoritative checkpoints centered-unique;
4. require no extra level consumption;
5. require generated-power final Scale/Level contract identical to production;
6. require polynomial PS plan scale remains `2^91`.

Classify:

`logn13_guarded_power_precision_candidate_validated`.

Do not integrate in this task.

---

# G9 — if no guard candidate passes

If no `g=0..12` candidate succeeds, identify the supported limit.

Classify exactly one:

- `logn13_guarded_power_blocked_by_pre_rescale_capacity`
- `logn13_guarded_power_blocked_by_product_capacity`
- `logn13_guarded_power_precision_improved_but_insufficient`
- `logn13_guarded_power_not_polynomial_bottleneck`

Record:

- largest capacity-safe g;
- best T16 error;
- best polynomial error;
- best EvalMod real/imag errors;
- first limiting checkpoint;
- whether remaining dominant error is generated-power, PS accumulation, normalized DA transform, or another demonstrated stage.

Only after this result may a later task decide whether mixed/per-block PS scales are actually necessary.

---

# Required artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-DESIGN-LOGN13-GUARDED-POWER-PRECISION-summary.json`

Include only:

- provenance;
- baseline metrics;
- one compact row per guard g containing:
  - first failure or pass;
  - worst capacity checkpoint/ratio;
  - selected factor summaries;
  - T2/T3/T4/T6/T8/T16 max errors;
  - polynomial error if reached;
  - EvalMod real/imag error if reached;
  - public-like error if reached;
- selected candidate if any;
- causal disposition;
- classification;
- first limiting checkpoint;
- validation flags.

Do not serialize full vectors, coefficient arrays, RNS rows, complete operation traces, or repeated hash collections.

---

# Validation

Before completion require:

- focused tests matching `TestFIX001P3.*Guarded.*Power.*Precision` pass;
- Primary `go test ./...` passes;
- Secondary remains exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production modification;
- no production integration;
- no PS plan-scale change;
- no C2S/S2C change;
- no Mod1/parameter retuning;
- no extra level consumption in successful candidates;
- no invalid raw Standard DA q0/q1 oracle;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.