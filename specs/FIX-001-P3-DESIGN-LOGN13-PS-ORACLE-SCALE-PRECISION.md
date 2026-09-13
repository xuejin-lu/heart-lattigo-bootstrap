# FIX-001-P3-DESIGN-LOGN13-PS-ORACLE-SCALE-PRECISION

## Purpose

Resolve the **Fast PS arithmetic blocker** that was independently proven by
`FIX-001-P3-DIAG-LOGN13-POLYNOMIAL-PRECISION-DECOMPOSITION`.

This task deliberately holds generated-power semantics near exact by using the validated oracle-power injection path. It asks one narrow question:

> What is the smallest common PS plan scale that makes Fast PS arithmetic alone satisfy the current compressed-polynomial budget, and where does q0/q1 capacity stop us?

This is a design/diagnostic task only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`08e8c3b150d029fc961925a93b79437cb54ddb5b`

Secondary required exact commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

Follow repository instructions and standing safe-push rules.

---

## Accepted evidence

The decomposition task established:

- classification: `logn13_polynomial_precision_joint_power_and_ps_blocker`;
- source/plaintext polynomial oracle vs genuine Standard: ~`3e-15` real/imag;
- `H_actual` (ideal PS, actual powers):
  - real ~`4.9391e-7`;
  - imag ~`5.0296e-7`;
- `H_oracle`: ~`3e-15`;
- `F_actual` production polynomial:
  - real ~`5.4574e-7`;
  - imag ~`5.2953e-7`;
- `F_oracle` (actual Fast PS, oracle powers, plan scale `2^91`):
  - real ~`6.1848e-8`;
  - imag ~`6.2675e-8`;
- decomposition closure `<5e-38`;
- oracle power injection error `<7e-16`;
- all oracle-injected q0/q1 rows matched full-RNS and were centered-unique;
- first PS arithmetic budget crossing: `F0-final-rescale`;
- current polynomial design budget:
  - `B = 1.2e-8`;
- worst measured PS capacity ratio in the decomposition: ~`0.2627` at `B4-constant`.

Therefore generated powers and PS arithmetic are both real blockers. This task addresses **only the PS arithmetic blocker**.

Do not re-run power-generation redesign here.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- change generated-power construction;
- use actual imperfect generated powers for candidate qualification;
- change polynomial coefficients, degree, basis or PS decomposition;
- implement mixed/per-block PS scales;
- change C2S or S2C;
- change Mod1 degree, DoubleAngle count, bootstrap parameters or Q/P chains;
- add q2+ maintained arithmetic;
- consume extra Q levels;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- integrate a candidate into production.

Primary diagnostic helpers/tests are allowed.

---

# P0 — exact baseline reproduction

Reproduce the accepted oracle-power decomposition baseline at plan scale `2^91`:

- D1 oracle alignment <= `1e-10` and approximately roundoff;
- oracle injection <= `1e-10` and approximately roundoff;
- oracle q0/q1 rows match full-RNS at every injected power;
- all injected powers centered-unique;
- `F_oracle` real/imag near `6.18e-8 / 6.27e-8`;
- first cumulative PS budget crossing remains `F0-final-rescale`.

If not, stop with:

`logn13_ps_oracle_scale_precondition_mismatch`.

---

# P1 — deterministic oracle-power common-scale sweep

Using the **same validated oracle-injected powers** and the exact same PS decomposition, evaluate common PS plan scales:

`2^91, 2^92, 2^93, 2^94, 2^95, 2^96`.

`2^91` is the control.

For each exponent:

- change only the common PS plan scale supplied to the existing Fast polynomial evaluator diagnostic path;
- keep all oracle power ciphertext Level/Scale/domain metadata equal to production generated-power metadata;
- keep coefficients, baby/giant grouping, operation ordering, relinearization points and level topology unchanged;
- use the actual Fast q0/q1 arithmetic path;
- use a stage-aligned full-RNS mirror where exact projection is valid.

Do not reject a candidate from Scale metadata alone. Use exact centered coefficients and Q01/2.

---

# P2 — PS-only capacity and primitive checkpoints

For every candidate, checkpoint enough of the actual PS path to determine the first real capacity or primitive failure.

At minimum include:

- baby constant injection for every block;
- baby `MulThenAdd` outputs;
- giant multiply inputs;
- giant pre-Rescale states;
- giant post-Rescale states;
- giant multiply-by-power outputs;
- aligned additions;
- final relinearization if present;
- `F0-final-rescale` input;
- `F0-final-rescale` output.

At each internal checkpoint compute:

- Level;
- exact Scale/log2;
- exact full-RNS centered max coefficient;
- Q01/2;
- capacity ratio;
- outside count;
- centered uniqueness;
- Fast q0/q1 vs full-RNS row equality while uniqueness holds;
- local arithmetic semantic error against the exact numerical operation on decoded actual inputs;
- cumulative oracle-path semantic error.

Artifact must remain compact: only serialize first failure, worst capacity checkpoint/ratio, first budget crossing and final metrics per candidate.

Candidate hard-failure classes:

- `ps_oracle_scalar_capacity_failure`
- `ps_oracle_baby_capacity_failure`
- `ps_oracle_giant_capacity_failure`
- `ps_oracle_final_rescale_capacity_failure`
- `ps_oracle_primitive_mismatch`
- `ps_oracle_stage_pass`.

Once centered uniqueness is lost, stop exact q0/q1 semantic attribution for that candidate.

---

# P3 — PS arithmetic precision qualification

For every capacity-safe candidate compare final Fast oracle-power polynomial output against:

1. validated source/plaintext polynomial oracle;
2. genuine Standard polynomial output on the same corrected-C2S input;
3. the stage-aligned full-RNS candidate mirror.

Record real and imag independently.

A candidate is PS-precision-qualified only if:

- Fast vs stage-aligned full-RNS is zero/roundoff where exact;
- real polynomial error vs genuine Standard <= `1.2e-8`;
- imag polynomial error vs genuine Standard <= `1.2e-8`;
- no extra Q level is consumed;
- all authoritative q0/q1 checkpoints remain centered-unique.

Select the **lowest exponent** satisfying all conditions.

If one qualifies, classify the PS-only result:

`logn13_ps_oracle_common_scale_candidate_validated`.

Do not integrate it yet.

---

# P4 — quantify the F0-final-rescale scaling law

Because `F0-final-rescale` was the first budget crossing at `2^91`, explicitly measure for every capacity-safe scale candidate:

- local error immediately before `F0-final-rescale`;
- local error introduced by `F0-final-rescale` itself;
- cumulative error immediately after it;
- final polynomial error;
- output Scale after final Rescale.

Determine whether the error approximately follows inverse output-scale behavior.

For successive candidate exponents record:

`error_ratio = E(k) / E(k+1)`.

This is diagnostic evidence only; do not force a theoretical `2x` expectation if measurements disagree.

---

# P5 — downstream sufficiency intervention for qualifying candidates

For every PS-precision-qualified candidate, continue the oracle-power candidate through the existing **candidate-derived normalized Mod1 schedule** rather than assuming production `K=2^29`.

Derive K from the candidate compressed polynomial Scale exactly as in the accepted common-scale sweep discipline:

- target Mod1 physical scale remains source-defined (~`2^60`);
- derive candidate K from targetScale / polynomialWorkingScale;
- derive normalized DoubleAngle multipliers/constants from that K schedule;
- require centered q0/q1 capacity throughout normalized DA;
- require Fast vs normalized full-RNS q0/q1 equality where exact;
- use genuine Standard only as semantic comparison, never as an invalid raw q0/q1 oracle after capacity loss.

Then run unchanged Fast S2C and supported one-input unpack/finalization/public-like path.

Record:

- EvalMod real/imag vs genuine Standard;
- post-S2C semantic error;
- public-like max-component error;
- metadata contract.

Authoritative downstream target:

`public_like_max_component_error <= 1e-2`.

This is an oracle-power intervention. It proves whether the PS arithmetic blocker can be closed independently; it does **not** claim production is fixed while generated powers remain imperfect.

---

# P6 — if no common scale qualifies

If none of `2^91 ... 2^96` satisfies the PS budget, classify exactly one supported outcome:

- `logn13_ps_oracle_common_scale_blocked_by_capacity`
- `logn13_ps_oracle_common_scale_insufficient_precision`

Record:

- largest capacity-safe exponent;
- best real/imag F_oracle polynomial error;
- first capacity-limiting checkpoint, if any;
- exact `F0-final-rescale` local/cumulative error at the best candidate;
- whether precision would require an exponent beyond the measured safe range according to observed scaling.

Only a later task may redesign `F0-final-rescale`, use a local higher-scale/final normalization strategy, or introduce mixed/per-block PS scaling.

---

# Required artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-DESIGN-LOGN13-PS-ORACLE-SCALE-PRECISION-summary.json`

Include only:

- provenance;
- baseline reproduction;
- one row per scale exponent containing:
  - exponent;
  - first failure/pass;
  - worst capacity checkpoint/ratio;
  - first cumulative budget crossing;
  - `F0-final-rescale` pre/local/post errors;
  - final real/imag polynomial errors;
  - output Scale;
  - downstream EvalMod/public-like metrics if qualified;
- selected exponent if any;
- classification;
- first limiting checkpoint;
- validation flags.

Do not serialize full vectors, coefficient arrays, RNS rows, complete traces or repeated hashes.

---

# Validation

Before completion require:

- focused tests matching `TestFIX001P3.*PS.*Oracle.*Scale.*Precision` pass;
- Primary `go test ./...` passes;
- Secondary remains exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no production integration;
- no generated-power redesign;
- no mixed/per-block scale implementation;
- no C2S/S2C changes;
- no bootstrap parameter retuning;
- no extra level consumption for a successful candidate;
- no invalid raw Standard DA q0/q1 oracle;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.