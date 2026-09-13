# FIX-001-P3-DESIGN-LOGN13-EVALMOD-PRECISION-SCALE-SWEEP

## Purpose

Design the smallest capacity-safe change that can reduce the remaining LogN13 EvalMod semantic error enough for public Bootstrap correctness.

This task follows the accepted causal diagnosis:

- Primary required base: `72cc6a44d4ce25b11f1e77cbd77d018f21e4aa68`
- Secondary exact required base: `25e70430b10cd4af37ad4cf94cd994912e970e3f` on `fast-ckks`
- production C2S compression `[4,2,0,0]` is correct;
- production C2S real/imag error is approximately `9e-14`;
- Fast S2C implementation effect is zero against a valid stage-aligned full-RNS reference;
- unpack effect = 0;
- finalization effect = 0;
- current Fast EvalMod vs genuine Standard is approximately:
  - real `4.566288964195857e-3`
  - imag `4.434999466167917e-3`;
- S2C amplifies that semantic difference empirically by approximately `53.93449x` on `reproducibleInput.v1`;
- resulting Fast public max-component error is approximately `0.24628047174869752`;
- classification: `logn13_evalmod_semantic_difference_amplified_by_s2c`.

Previous matched-input causal decomposition also established:

- Fast implementation vs normalized full-RNS mirror: `F-N = 0`;
- normalized recurrence effect is small, approximately `2.6e-5` at final EvalMod output;
- the dominant current term is the compressed-polynomial/coherent path vs genuine Standard, approximately `4.59e-3` after DoubleAngle;
- current compressed polynomial output differs from genuine Standard polynomial output by approximately `5.46e-7` before DoubleAngle;
- current production Fast polynomial plan scale is exactly `2^91`;
- current compressed polynomial output scale is approximately `2^31`;
- current normalized LogN13 production recurrence uses `K≈2^29` to restore a coherent scale near `2^60`.

The new evidence changes the precision requirement. A local EvalMod error below `1e-2` is not sufficient. Using the measured S2C transfer as guidance:

`1e-2 / 53.93449 ≈ 1.85e-4`.

For design margin, use:

`EVALMOD_LOCAL_DESIGN_TARGET = 1e-4`.

The current polynomial-to-final-EvalMod amplification is approximately:

`4.589e-3 / 5.457e-7 ≈ 8.4e3`.

Therefore a useful polynomial-stage design target is approximately:

`POLYNOMIAL_DESIGN_TARGET = 2e-8`.

These local targets are guidance. **Authoritative acceptance is the candidate's actual downstream public-like error after S2C, which must be <= `1e-2`.**

The first question is deliberately narrow:

> Can a higher common Fast PS plan scale than `2^91` recover enough polynomial precision while remaining q0/q1 capacity-safe through polynomial evaluation, normalized DoubleAngle, and final restore?

If yes, prefer that simple solution. If no, identify the exact limiting PS checkpoint that forces a later mixed/per-block scale design.

No production change is authorized in this task.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- change C2S compression;
- change S2C;
- change polynomial coefficients;
- change Mod1 degree, K, DoubleAngle count, interval shrink, target scale, Q/P parameters, or bootstrap configuration;
- add q2+ maintained arithmetic;
- implement a mixed/per-block production schedule in this task;
- change public thresholds;
- use raw genuine-Standard DoubleAngle q0/q1 projection after centered capacity loss;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003.

Primary diagnostic/design helpers are allowed. Secondary must remain exact and clean.

---

# S0 — reproduce fixed production input and baselines

Use the actual production pipeline through corrected C2S on deterministic `reproducibleInput.v1`.

Require:

- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- production C2S compression active with `[4,2,0,0]`;
- C2S real/imag vs genuine Standard <= current C2S budget;
- current production `2^91` EvalMod real/imag errors reproduce approximately the accepted `4.5e-3` range;
- current public Fast failure reproduces > `1e-2`;
- genuine Standard public Bootstrap passes.

If not, classify:

`logn13_evalmod_precision_sweep_precondition_mismatch`.

---

# S1 — derive and record precision budgets

From fresh reproduced metrics compute and record:

- public threshold = `1e-2`;
- measured current real/imag EvalMod deltas;
- measured current downstream S2C amplification ratio;
- derived no-margin local EvalMod budget = `public_threshold / measured_amplification`;
- fixed design local target = `1e-4`;
- current compressed-polynomial vs Standard polynomial delta;
- current coherent-after-DA vs Standard-after-DA delta;
- measured polynomial-to-EvalMod amplification ratio;
- derived approximate polynomial delta budget for `1e-4` local EvalMod;
- fixed polynomial design target = `2e-8`.

Do not treat the empirical amplification ratios as universal operator norms. They are design evidence for this exact LogN13 workload.

---

# S2 — deterministic common plan-scale candidate sweep

Test exact power-of-two common Fast PS plan scales:

`2^91, 2^92, 2^93, 2^94, 2^95, 2^96, 2^97, 2^98, 2^99, 2^100`.

`2^91` is the production baseline/control.

Do not silently skip a candidate merely because its exponent looks larger than q0/q1. Capacity is about actual centered encoded/intermediate coefficients, not Scale metadata alone.

For each candidate:

1. use the same source polynomial and same actual C2S-corrected Mod1 input;
2. call/replay the real Fast polynomial PS algorithm with the candidate as common `planScale`;
3. keep generated-power logic unchanged;
4. keep Levels and PS decomposition unchanged;
5. use the established authoritative Chebyshev semantic oracle `raw_chebyshev_on_preprocessed_z`;
6. build/reuse a stage-aligned full-RNS normalized mirror for exact capacity and q0/q1 row checks where unique.

No production source edit is permitted to inject these candidates.

---

# S3 — polynomial candidate checkpoints and capacity

For every candidate, collect compact evidence at the real PS checkpoints sufficient to locate the first loss of capacity or precision.

At minimum cover:

- each baby-block final accumulator;
- each giant-step pre-Rescale state;
- each giant-step post-Rescale state;
- each giant-step multiply/add-aligned result;
- final pre-Rescale polynomial state;
- final post-Rescale compressed polynomial output.

For scalar constant / coefficient injections also compute the nominal encoded scalar magnitude when meaningful.

At each checkpoint record:

- Level;
- Scale/log2;
- max exact centered coefficient for the full-RNS mirror;
- `Q01/2`;
- capacity ratio;
- outside count;
- whether centered q0/q1 reconstruction is unique;
- Fast q0/q1 vs stage-aligned full-RNS row-match boolean while unique;
- semantic error vs the correct local/plaintext oracle where meaningful.

Rules:

- if a full-RNS checkpoint exceeds centered `Q01/2`, do not q0/q1-project it as an exact oracle;
- mark candidate first capacity failure and stop exact-row reasoning after that invalid point;
- do not infer safety from wrapped Fast residues alone.

Per-candidate polynomial classifications may include:

- `polynomial_plan_capacity_failure`
- `polynomial_fast_vs_full_rns_mismatch`
- `polynomial_semantic_precision_failure`
- `polynomial_stage_pass`.

---

# S4 — generalized normalized DoubleAngle for each surviving candidate

A higher plan scale will generally change the compressed polynomial output Scale. Do **not** assume the production `K=2^29` remains correct.

For each polynomial-surviving candidate, derive from source semantics:

`targetScale` = same Standard/source-defined Mod1 target scale.

Let compressed output scale be `W`.

Derive the nearest power-of-two normalization factor:

`K0 ≈ targetScale / W`.

Use high-precision arithmetic and record exponent/value/error.

Then replay the already validated normalized recurrence generically:

`z_i = y_i / K_i`

`z_{i+1} = A_i * z_i^2 - C_i`

where

`A_i = 2*K_i^2/K_{i+1}`

`C_i = c_i/K_{i+1}`.

Derive `K_i`, `A_i`, and constants from the candidate Scale schedule; do not hardcode historical `29/30` exponents.

For each round require:

- valid Level/Scale transition;
- normalized full-RNS centered capacity safe at exact-row checkpoints;
- Fast q0/q1 rows equal normalized full-RNS reference while unique;
- mapped semantic `K_i*z_i` remains close to the coherent ordinary full-RNS path.

At final restore:

- derive final `K`;
- prove prospective `K * B < Q01/2` from coefficient-domain exact capacity;
- perform physical+metadata restore;
- verify exact Fast-vs-normalized q0/q1 rows;
- apply source-defined final metadata Scale reset.

Candidate failure classes:

- `normalized_candidate_scale_derivation_failure`
- `normalized_candidate_capacity_failure`
- `normalized_candidate_primitive_mismatch`
- `normalized_candidate_semantic_failure`
- `normalized_candidate_pass`.

---

# S5 — candidate local EvalMod precision

For every candidate that completes polynomial + normalized DoubleAngle, compare final candidate EvalMod output against:

1. normalized full-RNS mirror — exact/stage-aligned effect;
2. coherent ordinary-DA path from the same candidate compressed polynomial output — normalization-transform effect;
3. genuine Standard EvalMod on the same corrected C2S input — authoritative matched-input semantic delta.

For real and imag record max-component / max-complex / mean-complex.

Require for a **precision-qualified** candidate:

- Fast vs normalized mirror = zero or numerical roundoff;
- real Fast vs genuine Standard <= `1e-4`;
- imag Fast vs genuine Standard <= `1e-4`.

Also record candidate compressed polynomial vs Standard polynomial delta and whether it meets the guidance target `2e-8`.

The `2e-8` polynomial target is not mandatory if the actual downstream candidate passes.

---

# S6 — actual downstream S2C/public-like validation per precision-qualified candidate

For each candidate meeting the local `1e-4` EvalMod target, run the actual unchanged Fast S2C path on the candidate real/imag outputs.

Then reproduce the same supported one-input unpack/finalization/public representation path used by the current diagnostic harness, without changing production code.

Record:

- Fast S2C implementation effect vs a valid stage-aligned full-RNS S2C mirror;
- post-S2C semantic error against the corresponding genuine Standard path;
- final public-like max-component error against original deterministic message;
- Level/Scale/Degree/domain contract.

Authoritative candidate acceptance requires:

`public_like_max_component_error <= 1e-2`.

If local `1e-4` passes but public-like output fails, keep the candidate rejected and record actual downstream amplification.

---

# S7 — preferred common-scale candidate

If one or more candidates pass the full public-like path, choose deterministically:

1. **lowest plan-scale exponent** that passes public-like `1e-2`;
2. among equal exponent variants, lowest public-like max-component error;
3. require every exact-capacity checkpoint used by Fast q0/q1 semantics to remain centered-unique;
4. require no new level consumption;
5. require current C2S/S2C/parameter contracts unchanged.

Choosing the lowest passing scale avoids unnecessary coefficient growth and preserves maximum q0/q1 headroom.

Overall classification:

`logn13_evalmod_common_plan_scale_candidate_validated`.

Record selected exponent, local EvalMod errors, polynomial error, minimum capacity headroom, and public-like error.

Do **not** integrate it into Secondary in this task.

---

# S8 — if no common-scale candidate passes

If no candidate in `2^91 ... 2^100` is both capacity-safe and public-like correct, do not invent a production fix.

Identify the earliest limiting mechanism among candidates that improve precision:

- exact PS scalar injection capacity;
- baby accumulator capacity;
- giant pre-Rescale capacity;
- giant multiplication capacity;
- final polynomial capacity;
- normalized DoubleAngle capacity;
- insufficient precision improvement despite capacity headroom.

Record the highest precision-achieving safe common scale and its metrics.

Then classify one of:

- `logn13_evalmod_common_scale_blocked_by_ps_capacity`
- `logn13_evalmod_common_scale_blocked_by_normalized_da_capacity`
- `logn13_evalmod_common_scale_insufficient_precision`

Only if this happens may the **next** task consider a mixed/per-block scale schedule. Do not implement or fully design that schedule here.

---

# Required artifact

Create one compact artifact:

`results/FIX-001-P3-DESIGN-LOGN13-EVALMOD-PRECISION-SCALE-SWEEP-summary.json`

Include only:

- provenance;
- reproduced baseline metrics;
- derived budgets;
- one compact row per candidate with:
  - exponent;
  - polynomial output scale/error;
  - minimum/worst q0/q1 capacity ratio and checkpoint;
  - first failing checkpoint/classification if any;
  - derived K schedule summary;
  - real/imag final EvalMod errors if reached;
  - public-like error if reached;
- preferred candidate if any;
- classification;
- first failing/limiting checkpoint;
- validation flags.

Do not serialize full slots, coefficients, RNS limbs, PS traces, repeated hashes, or large per-operation arrays.

---

# Validation

Before completion require:

- focused tests matching `TestFIX001P3.*EvalMod.*Precision.*Scale.*Sweep` pass;
- `go test ./...` passes in Primary;
- Secondary remains exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary modification;
- no production fix;
- no C2S/S2C change;
- no parameter retuning;
- no invalid raw Standard DoubleAngle q0/q1 oracle;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- compact artifact committed;
- Primary ordinary fast-forward push under standing safe-push authorization;
- Primary worktree clean and synchronized.