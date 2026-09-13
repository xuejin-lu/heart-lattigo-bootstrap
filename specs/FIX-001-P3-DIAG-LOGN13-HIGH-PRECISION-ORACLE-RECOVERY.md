# FIX-001-P3-DIAG-LOGN13-HIGH-PRECISION-ORACLE-RECOVERY

## Purpose

Recover a trustworthy stage-aligned polynomial oracle for the residual EvalMod causal decomposition, then immediately resume the PS-vs-DA counterfactual analysis if oracle injection qualifies.

Accepted evidence from Primary commit `591c8fa5dabada8ac3e2577f5fc512a2fab85e3c`:

- accepted executed system path uses oracle powers, q0/q1/q2 = 56/39/40;
- intended guard schedule is G0 local-q2 = 2 bits, F0 local-q2 = 3 bits;
- all 3 DA rounds use the validated bounded local-q2 path and contract after each round;
- current system metrics remain:
  - EvalMod real vs Standard `6.261306805777143e-5`;
  - EvalMod imag vs Standard `4.647793870173522e-5`;
  - post-S2C `3.3385298628089955e-4`;
  - public-like `0.010683237260415292`;
- prior S2C decomposition proved:
  - Fast S2C implementation effect = 0;
  - full-RNS mirror fidelity = 0;
  - closure residual = 0;
  - EvalMod-delta -> post-S2C amplification ~5.332x;
- previous residual decomposition stopped at E1 because P_O materialization error was too large:
  - real oracle injection floor `6.291603904529097e-10`;
  - imag oracle injection floor `7.071984109430218e-10`;
  - P_F-P_O real `7.45649086919542e-9`;
  - P_F-P_O imag `5.5101831986092975e-9`.

Secondary source at the fixed commit supports arbitrary-precision CKKS encoding:

- `ckks.NewEncoder(parameters, precision)` uses `float64/complex128` only when precision <=53;
- precision >53 uses `*big.Float` / `*bignum.Complex` for slot FFT/encoding.

This task uses that existing capability only for Primary diagnostics. Do not modify Secondary.

---

## Required provenance

Primary required base:

`591c8fa5dabada8ac3e2577f5fc512a2fab85e3c`

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
- all-round DA local-q2 design unchanged;
- same polynomial coefficients, interval, scalar, offset, PS decomposition and plan scale;
- same S2C mirror and final x32 restore semantics.

Do not:

- retune any arithmetic candidate;
- alter q parameters;
- change G0/F0/DA behavior;
- redesign generated powers;
- modify Secondary production code;
- modify S2C production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary-only diagnostic high-precision code is allowed.

---

# H0 — fix diagnostic provenance/metadata consistency

The previous artifact serialized an internally inconsistent candidate summary: the object name indicated G0 local-q2 guard=2 while some legacy fields still showed a 1-bit G0 schedule and `scale_alignment_safe=false / valid=false / first_failure=G0-add`.

Before any new decomposition:

1. inspect the actual executed diagnostic path at the accepted base;
2. confirm the real executed guard schedule was G0=2, F0=3 and all-round DA local-q2;
3. confirm the full downstream path used for the reported EvalMod/post-S2C/public-like metrics completed all required exact capacity/row/rounded-division contracts;
4. rebuild diagnostic provenance from the actually executed schedule rather than reusing a stale q01-only candidate object.

If the inconsistency is serialization-only, correct the new artifact/test metadata and proceed without changing arithmetic.

If the actual execution path itself does not match G0=2/F0=3, stop:

`logn13_high_precision_oracle_control_provenance_mismatch`.

---

# H1 — end-to-end high-precision polynomial oracle

Construct P_O without passing through `float64` or `complex128` before final diagnostic decoding.

Requirements:

1. represent the preprocessed polynomial input in arbitrary precision;
2. evaluate the exact same source polynomial/Chebyshev semantics in arbitrary precision end-to-end;
3. preserve the exact interval/scalar/offset and coefficient values used by the accepted source path;
4. materialize slots using `[]*bignum.Complex` (or equivalent arbitrary-precision representation supported by the fixed Secondary source);
5. use `ckks.NewEncoder(parameters, precision)` with precision >53;
6. materialize at the same DA-input metadata as P_F:
   - Level 7;
   - exact accepted Scale near `1.374389534718183583021165637421335497639e+11`;
   - same dimensions, NTT state and degree semantics;
7. do not use a higher ciphertext Scale and later rescale as a substitute;
8. do not convert the oracle to complex128 before `Encode/Embed`.

Diagnostic decode to complex128 for final error reporting is allowed after materialization because that floor is far below 1e-10; if used, verify this explicitly with a higher-precision decode comparison.

---

# H2 — precision sweep

Test encoder/oracle arithmetic precisions exactly:

- 80 bits;
- 96 bits;
- 128 bits;
- 160 bits.

For each precision and for real/imag branches record:

- oracle source precision;
- encoder precision;
- P_O encode/decode injection error vs the high-precision source oracle;
- P_F-P_O error;
- Level/Scale/NTT metadata match;
- whether the same integer plaintext/ciphertext materialization is deterministic across repeated runs.

Qualification requires BOTH branches:

- injection max-component <= `1e-10`;
- injection error <= 0.1 * corresponding P_F-P_O error;
- exact stage metadata match.

Select the smallest precision that qualifies.

If no tested precision qualifies, classify:

`logn13_high_precision_oracle_materialization_still_insufficient`

and stop. Do not weaken the causal-decomposition threshold in this task.

---

# H3 — high-precision decode sanity

For the selected oracle precision:

- decode P_O once to complex128;
- decode/evaluate the same state in arbitrary precision where supported;
- quantify the reporting-only conversion floor.

Require reporting conversion floor <= `1e-12`.

If not, use arbitrary-precision differences for the causal metrics and serialize decimal summaries only.

---

# H4 — resume residual EvalMod causal decomposition

Only after H2 qualifies, resume the blocked decomposition using the selected high-precision P_O.

Define:

- `P_F`: actual accepted Fast polynomial/PS output;
- `P_O`: qualified high-precision oracle materialized at identical DA input metadata;
- `DA_F`: actual accepted Fast/local-q2 DA implementation;
- `DA_N`: stage-aligned normalized full-RNS CKKS DA mirror with the exact same integer arithmetic, constants, Levels, Scales and Rescale divisors;
- `DA_I`: high-precision ideal mathematical DoubleAngle recurrence with no CKKS quantization;
- `E_S`: genuine Standard EvalMod output.

Construct:

- `E_F = DA_F(P_F)`;
- `N_F = DA_N(P_F)`;
- `N_O = DA_N(P_O)`;
- `I_O = DA_I(P_O)`.

Report vector-wise compact metrics for:

1. Fast/local-q2 implementation effect: `E_F - N_F`;
2. PS-input propagation effect: `N_F - N_O`;
3. normalized DA CKKS arithmetic/rounding effect: `N_O - I_O`;
4. ideal/reference alignment floor: `I_O - E_S`;
5. observed total: `E_F - E_S`.

Require vector closure residual <= `1e-10` or the rigorously measured arbitrary-precision numerical floor if smaller.

At DA input and every round square/multiplier/constant/rescale checkpoint, record compact max-component values for:

- propagated PS-input component;
- DA arithmetic component;
- worst index/component;
- exact row/capacity contract flags.

---

# H5 — counterfactual system sufficiency

Use the already-validated full-RNS S2C linear mirror and unchanged final x32 restore.

Evaluate exactly:

### C0 actual
`DA_F(P_F)` -> S2C -> restore.

Must reproduce public-like near `0.010683237260415292`.

### C_PS perfect polynomial, current DA CKKS arithmetic
`DA_N(P_O)` -> S2C -> restore.

### C_DA current polynomial, ideal DA arithmetic
Construct `I_F = DA_I(P_F)`, then S2C -> restore.

### C_BOTH ideal polynomial + ideal DA
`DA_I(P_O)` -> S2C -> restore.

For each record:

- EvalMod-domain max-component vs Standard;
- post-S2C max-component vs Standard;
- public-like max-component vs Standard;
- pass/fail against `1e-2`.

Also project each causal EvalMod component independently through the validated S2C linear map and compute vector closure after S2C.

Do not use norm subtraction to estimate counterfactual improvement; evaluate the underlying vectors.

---

# H6 — decision classification

Choose exactly one primary classification:

- `logn13_evalmod_residual_ps_input_blocker`
  - C_PS passes, C_DA fails.

- `logn13_evalmod_residual_da_arithmetic_blocker`
  - C_DA passes, C_PS fails.

- `logn13_evalmod_residual_multiple_single_fix_options`
  - both C_PS and C_DA independently pass; record which has larger margin.

- `logn13_evalmod_residual_joint_ps_da_blocker`
  - C_PS and C_DA fail, C_BOTH passes.

- `logn13_evalmod_residual_mathematical_chain_floor`
  - C_BOTH fails.

- `logn13_evalmod_residual_fast_implementation_mismatch`
  - E_F-N_F is materially nonzero despite accepted exact contracts.

If H2 fails first, use only:

`logn13_high_precision_oracle_materialization_still_insufficient`.

No design fix is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-HIGH-PRECISION-ORACLE-RECOVERY-summary.json`

Include only:

- provenance and corrected executed-candidate metadata;
- H2 precision-sweep table;
- selected precision;
- injection/source/reporting floors;
- P_F-P_O real/imag;
- resumed causal decomposition if reached;
- compact DA checkpoint attribution if reached;
- C0/C_PS/C_DA/C_BOTH table if reached;
- S2C-projected causal metrics if reached;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize slot vectors, coefficients, full RNS rows, matrices or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*High.*Precision.*Oracle` pass;
- if H4 is reached, focused residual-decomposition tests also pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no accepted arithmetic candidate changes;
- no q/S2C/generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.