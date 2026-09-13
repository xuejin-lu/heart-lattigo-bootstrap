# FIX-001-P3-DIAG-LOGN13-QUANTIZATION-AWARE-EVALMOD-DECOMPOSITION

## Purpose

Resume the blocked residual EvalMod causal decomposition without pretending that a mathematical slot oracle can be injected into a finite-scale CKKS state with zero error.

Accepted evidence from Primary commit `dbb2e96fe2b6366d6870ccf827dda87668141feb`:

- accepted executed path: LogN13, oracle powers, q0/q1/q2 = 56/39/40;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all 3 DoubleAngle rounds use the validated bounded local-q2 pattern;
- all capacity/row/rounded-division/contraction/final metadata contracts pass;
- current system metrics remain:
  - EvalMod real vs Standard `6.261306805777143e-5`;
  - EvalMod imag vs Standard `4.647793870173522e-5`;
  - post-S2C `3.3385298628089955e-4`;
  - public-like `0.010683237260415292`;
- prior S2C decomposition proved:
  - Fast S2C implementation effect = 0;
  - full-RNS mirror fidelity = 0;
  - closure residual = 0;
  - EvalMod delta -> post-S2C amplification ~5.332x;
- high-precision oracle sweep at 80/96/128/160 bits produced the same deterministic materialization floor:
  - real ~`6.2916043e-10`;
  - imag ~`7.0719838e-10`;
  - reporting-only complex conversion floor ~`5.55e-17`.

The DA-input CKKS Scale is approximately:

`Delta = 1.374389534718183583021165637421335497639e11`.

Thus one coefficient-domain quantization unit is roughly `1/Delta ~= 7.28e-12`. For N=8192, a slot-domain transform of independent coefficient quantization has a natural order around `sqrt(N)/Delta ~= 6.59e-10`, which is the same order as the measured oracle encode/decode floor.

Therefore the previous `<=1e-10` injection gate was too strict for this fixed CKKS representation. The oracle materialization floor must be modeled as a separate causal component, not treated as a failed diagnostic implementation.

Diagnostic only. No production arithmetic changes are authorized.

---

## Required provenance

Primary required base:

`dbb2e96fe2b6366d6870ccf827dda87668141feb`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Keep the accepted executed candidate fixed:

- oracle powers;
- q0/q1/q2 56/39/40;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all-round DA local-q2 design;
- polynomial coefficients, interval, scalar, offset, PS decomposition and plan scale;
- S2C matrix/factorization/scaling;
- final x32 restore/public-like semantics.

Do not:

- retune guards, scales, K values or constants;
- alter q parameters;
- modify Secondary production code;
- modify S2C production code;
- redesign generated powers;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary-only diagnostic arbitrary-precision helpers/full-RNS mirrors are allowed.

---

# Q0 — reproduce accepted control

Reproduce:

- actual Fast polynomial/PS output `P_F` at DA input;
- DA-input metadata: Level 7, accepted Scale, degree/dimensions/NTT semantics;
- EvalMod real/imag control;
- post-S2C control;
- public-like control;
- all accepted exact arithmetic/capacity contracts.

If mismatch, stop:

`logn13_quantization_aware_decomposition_precondition_mismatch`.

---

# Q1 — establish canonical CKKS oracle state

Use at least 160-bit source polynomial/oracle arithmetic and a 160-bit CKKS encoder.

Define:

- `P_O`: ideal mathematical polynomial oracle in slot space before CKKS materialization;
- `P_Q`: the canonical CKKS state obtained by encoding `P_O` at exactly the accepted `P_F` DA-input Level/Scale/metadata and then decoding that exact integer/RNS state back to high-precision slots.

`P_Q` is the best representable reference state for this fixed CKKS metadata. Its difference from `P_O` is **representation quantization**, not oracle implementation failure.

Require:

- deterministic integer/RNS materialization across repeated encodes;
- identical logical Level/Scale/degree/dimensions/NTT metadata to `P_F`;
- reporting conversion floor <= `1e-12`;
- 80/96/128/160-bit encoder results already known to converge to the same integer materialization; optionally re-check only 80 and 160 bits to prove stability, do not repeat a large sweep unless required by implementation.

Record:

- `P_Q - P_O` real/imag max-component error;
- `P_F - P_Q` real/imag max-component error;
- `P_F - P_O` real/imag max-component error;
- `1/Delta`;
- `sqrt(N)/Delta`;
- measured max injection divided by `sqrt(N)/Delta`.

Do not require `P_Q-P_O <=1e-10`.

If materialization is non-deterministic or depends materially on encoder precision >53 bits, stop:

`logn13_canonical_ckks_oracle_materialization_unstable`.

---

# Q2 — define DA evaluators

Use the exact accepted normalized DoubleAngle recurrence and schedule.

Define:

### `DA_F`
Actual accepted Fast/local-q2 DA implementation.

### `DA_N`
Stage-aligned normalized full-RNS CKKS DA mirror using the same integer arithmetic, constants, Levels, Scales and Rescale divisors as the accepted design, without q0/q1 truncation ambiguity.

### `DA_I`
High-precision ideal mathematical DoubleAngle recurrence with the exact same normalized transformation but no CKKS integer rounding/rescale quantization.

Validate:

- `DA_F(P_F)` rows agree with `DA_N(P_F)` at all exact accepted checkpoints;
- ideal recurrence uses source-derived constants/K schedule;
- no metadata-only relabeling is introduced in the full-RNS mirror.

---

# Q3 — five-component EvalMod causal decomposition

Construct:

- `E_F = DA_F(P_F)`;
- `N_F = DA_N(P_F)`;
- `N_Q = DA_N(P_Q)`;
- `I_Q = DA_I(P_Q)`;
- `I_O = DA_I(P_O)`;
- `E_S = genuine Standard EvalMod output`.

Compute underlying vector differences and serialize only compact metrics for:

1. **Fast/local-q2 implementation effect**
   - `A = E_F - N_F`

2. **PS arithmetic excess over canonical representable oracle**
   - `B = N_F - N_Q`

3. **DA CKKS arithmetic/rounding effect**
   - `C = N_Q - I_Q`

4. **mandatory DA-input CKKS representation quantization effect**
   - `D = I_Q - I_O`

5. **ideal mathematical/reference alignment floor**
   - `E = I_O - E_S`

Observed total:

`T = E_F - E_S`.

Verify vector closure:

`A + B + C + D + E = T`.

Require closure residual <= `1e-10` or the rigorously measured arbitrary-precision reporting floor if smaller.

Do not infer causality from max norms alone; cancellation/direction matters.

---

# Q4 — DA checkpoint attribution

At DA input and every round checkpoint:

- input;
- after square;
- after multiplier;
- after constant;
- after Rescale;

for rounds 0,1,2, record compact metrics for:

- `DA_N(P_F) - DA_N(P_Q)` : propagated PS arithmetic excess;
- `DA_N(P_Q) - DA_I(P_Q)` : DA CKKS arithmetic component;
- `DA_I(P_Q) - DA_I(P_O)` : canonical representation quantization component;
- worst slot/component for each;
- accepted capacity/row/rounded-division flags.

Goal: locate the first stage where B/C/D becomes materially important.

---

# Q5 — downstream-sensitive counterfactuals

Use the previously validated full-RNS S2C linear mirror and unchanged final x32 restore.

Evaluate these decoded/stage-aligned cases against genuine Standard:

### C0 — actual

`E_F -> S2C -> restore`.

Must reproduce public-like near `0.010683237260415292`.

### C_PS — eliminate PS arithmetic excess only

`N_Q -> S2C -> restore`.

Interpretation:

> The polynomial/PS path is corrected to the best state canonically representable at the existing DA-input Level/Scale; DA CKKS arithmetic and mandatory encoding quantization remain.

### C_DA — eliminate DA arithmetic only

Construct `I_F = DA_I(P_F)` and run:

`I_F -> S2C -> restore`.

### C_BOTH_CKKS — eliminate PS excess and DA arithmetic, retain mandatory CKKS input quantization

`I_Q -> S2C -> restore`.

### C_MATH — ideal mathematical polynomial and ideal DA

`I_O -> S2C -> restore`.

This removes even the fixed DA-input CKKS representation quantization and provides the mathematical floor of the current polynomial/S2C/finalization chain.

For every case record:

- EvalMod-domain max-component vs Standard;
- post-S2C max-component vs Standard;
- public-like max-component vs Standard;
- pass/fail against `1e-2`.

Use underlying vectors. Do not estimate by subtracting scalar norms.

---

# Q6 — S2C projection of causal components

Using the validated full-RNS S2C linear map, independently project:

- `T(A)` implementation;
- `T(B)` PS excess;
- `T(C)` DA arithmetic;
- `T(D)` representation quantization;
- `T(E)` ideal/reference floor.

Record:

- max-component of each after S2C;
- worst output index/component;
- vector closure residual after S2C;
- contribution of each component at the actual worst public-like slot/component;
- public-like value if each component is removed individually from the actual vector.

This is the primary decision evidence because the current failure is downstream-direction-sensitive.

---

# Q7 — decision classification

Choose exactly one primary classification.

### PS fix alone sufficient

If `C_PS <= 1e-2` and `C_DA > 1e-2`:

`logn13_evalmod_residual_ps_input_blocker`

### DA fix alone sufficient

If `C_DA <= 1e-2` and `C_PS > 1e-2`:

`logn13_evalmod_residual_da_arithmetic_blocker`

### Either single fix sufficient

If both `C_PS` and `C_DA` pass:

`logn13_evalmod_residual_multiple_single_fix_options`

Record which has the larger public-like margin and lower implementation scope.

### Joint PS + DA fix required

If `C_PS` and `C_DA` fail but `C_BOTH_CKKS` passes:

`logn13_evalmod_residual_joint_ps_da_blocker`

### Fixed CKKS representation itself blocks the target

If `C_BOTH_CKKS` fails but `C_MATH` passes:

`logn13_evalmod_residual_ckks_representation_floor`

This means the accepted DA-input Level/Scale representation is itself part of the system blocker and must be redesigned before further guard tuning.

### Mathematical chain floor

If `C_MATH` fails:

`logn13_evalmod_residual_mathematical_chain_floor`

### Fast implementation mismatch

If `A` is materially nonzero despite accepted exact row/capacity contracts:

`logn13_evalmod_residual_fast_implementation_mismatch`

No design fix is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-QUANTIZATION-AWARE-EVALMOD-DECOMPOSITION-summary.json`

Include only:

- provenance;
- Q0 control;
- canonical-oracle quantization metrics and theoretical scale quantities;
- five-component EvalMod decomposition;
- closure residual;
- compact DA checkpoint attribution table;
- C0/C_PS/C_DA/C_BOTH_CKKS/C_MATH table;
- S2C-projected component metrics;
- required post-S2C threshold `3.125e-4`;
- public-like threshold `1e-2`;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize slot vectors, coefficients, full RNS rows, matrices or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*Quantization.*Aware.*EvalMod` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no candidate retuning;
- no q/S2C/generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.