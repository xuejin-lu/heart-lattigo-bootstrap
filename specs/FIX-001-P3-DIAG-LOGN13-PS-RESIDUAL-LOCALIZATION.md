# FIX-001-P3-DIAG-LOGN13-PS-RESIDUAL-LOCALIZATION

## Purpose

Localize the smallest practical Polynomial/Paterson–Stockmeyer correction that can close the remaining LogN13 public-like gap.

Accepted evidence at Primary commit `c33dba976f13ef27c1cb35199a05c442acc36934`:

- fixed executed path is LogN13 with oracle powers;
- q0/q1/q2 = 56/39/40;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all three DoubleAngle rounds use the accepted bounded local-q2 path;
- Fast/local-q2 implementation effect at EvalMod is exactly zero versus the stage-aligned full-RNS mirror;
- S2C implementation effect is zero and its mirror fidelity is zero;
- actual public-like error is `0.010683237260415292`;
- canonical CKKS polynomial oracle `P_Q` is deterministic at the same DA-input Level/Scale;
- `P_F - P_Q` is approximately:
  - real `7.432464310674902e-9`;
  - imag `5.729599017456621e-9`;
- replacing only the PS output with canonical `P_Q` while keeping the current CKKS DoubleAngle arithmetic (`C_PS`) gives public-like error about `3.91e-4`, comfortably below `1e-2`;
- removing DA CKKS arithmetic instead (`C_DA`) does **not** help and breaks alignment with the genuine Standard reference;
- classification from the accepted quantization-aware decomposition is `logn13_evalmod_residual_ps_input_blocker`.

Important interpretation:

> The next fix must move the actual PS output toward the canonical CKKS state `P_Q` while preserving the current DA arithmetic exactly. Do not attempt to make DoubleAngle more "ideal".

This task is diagnostic/localization only. It must identify the smallest PS block/operation worth changing next; it must not change Secondary production code.

---

## Required provenance

Primary required base:

`c33dba976f13ef27c1cb35199a05c442acc36934`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Keep fixed:

- oracle powers;
- polynomial coefficients/basis/interval/scalar/offset;
- PS decomposition;
- common plan scale `2^92`;
- q0/q1/q2 profile 56/39/40;
- G0 guard=2 and F0 guard=3 for the control path;
- all accepted DA local-q2 behavior;
- S2C and final restore behavior;
- genuine Standard widened-profile reference.

Do not:

- change DA arithmetic, K schedule, constants or local-q2 policy;
- retune global plan scale;
- change q parameters;
- redesign generated powers;
- modify Secondary production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary diagnostic stage capture, high-precision oracle construction, full-RNS mirrors and counterfactual checkpoint injection are allowed.

---

# P0 — reproduce the accepted control

Reproduce compactly:

- actual PS output `P_F` at DA input;
- canonical CKKS oracle `P_Q` at identical logical metadata;
- `P_F-P_Q` real/imag metrics;
- actual EvalMod/post-S2C/public-like metrics;
- `C_PS` public-like result from `P_Q` through the unchanged current DA/S2C/finalization path;
- all exact capacity/row/rescale contracts.

If the control differs materially from commit `c33dba...`, stop:

`logn13_ps_residual_localization_precondition_mismatch`.

---

# P1 — reconstruct the actual PS DAG and named checkpoints

Derive the exact PS execution DAG from the fixed Secondary source and the existing Primary diagnostics. Do not infer block order from labels alone.

At minimum identify the source-backed checkpoints corresponding to the previously used logical regions:

- G0;
- G1;
- G2;
- G3;
- F0/final accumulation;

and, within each region where present:

- accumulator before term;
- multiply / MulThenAdd input;
- immediately after multiply;
- immediately after scalar/plaintext add;
- immediately before Rescale;
- immediately after Rescale;
- any metadata scale-alignment/relabel operation.

Record for every checkpoint:

- source file/function and a compact operation label;
- Level;
- exact Scale;
- degree;
- whether q2 is active;
- whether a Rescale occurs next.

Do not serialize source code or large traces.

---

# P2 — canonical stage oracle for each PS checkpoint

For each named checkpoint, construct a high-precision mathematical value for the exact partial PS expression represented at that point, using the same oracle powers and exact polynomial coefficients.

Materialize it into a **canonical CKKS checkpoint state** at the exact actual checkpoint Level/Scale/metadata, using the accepted quantization-aware method:

- arbitrary-precision polynomial evaluation;
- >53-bit encoder;
- same Level/Scale/dimensions/NTT metadata;
- deterministic integer/RNS materialization.

Call the actual checkpoint `X_F` and canonical representable checkpoint `X_Q`.

For each checkpoint record:

- `X_F-X_Q` max-component error;
- mean error;
- worst slot/component;
- canonical materialization floor `X_Q-X_O`;
- error-to-floor ratio.

Do not reject a checkpoint merely because the materialization floor exceeds `1e-10`; use the quantization-aware interpretation established in the accepted task.

If the partial-expression oracle cannot be matched unambiguously to a source PS checkpoint, mark that checkpoint unsupported and continue with the remaining source-proven checkpoints. Do not guess.

---

# P3 — incremental residual growth

Using underlying vectors, compute how the canonical residual changes from checkpoint to checkpoint.

For consecutive supported checkpoints `i -> j`, report:

- actual residual vector at `i`: `R_i = X_F_i - X_Q_i`;
- actual residual vector at `j`: `R_j = X_F_j - X_Q_j`;
- max-component growth;
- mean growth;
- cosine/correlation or another direction-preserving similarity metric between `R_i` propagated through the exact mathematical suffix to `j` and `R_j`;
- whether the operation introduces a new dominant residual direction versus merely amplifying an existing one.

The purpose is to distinguish:

- error introduced by one operation;
- harmless amplification of earlier error;
- cancellation between operations.

Do not rank blocks using scalar max norms alone.

---

# P4 — suffix-replay checkpoint counterfactuals

For every supported PS checkpoint, perform a **checkpoint reset counterfactual**:

1. replace the actual checkpoint state `X_F` with its canonical CKKS state `X_Q`;
2. run the remaining PS suffix using the unchanged accepted Fast diagnostic arithmetic and oracle powers;
3. run the unchanged accepted DA local-q2 path;
4. run the validated S2C/finalization path;
5. compare to the genuine Standard reference.

For each reset record:

- final PS `P_F-P_Q` after suffix replay;
- EvalMod max-component vs Standard;
- post-S2C max-component vs Standard;
- public-like max-component vs Standard;
- pass/fail at `1e-2`;
- final metadata/contract pass.

This is the primary causal test.

Interpretation:

- if resetting checkpoint `k` passes but resetting the immediately previous checkpoint does not, the region between them is strongly implicated;
- if many early resets pass but later resets do not, identify the **latest** passing reset because it gives the smallest correction scope;
- if only the final PS-output reset passes, classify the residual as distributed/joint rather than inventing a single dominant block.

---

# P5 — single-operation counterfactual drill-down

Take only the smallest implicated PS region from P4.

Within that region, for each source-backed operation checkpoint:

- replace only that operation's output with its canonical CKKS checkpoint state;
- replay the remaining PS suffix and full downstream path unchanged;
- record public-like result and exact contracts.

Do not test unrelated blocks.

If one operation reset alone yields `public_like <= 1e-2`, identify it as the minimal causal operation.

If no individual operation passes but the whole-region reset passes, classify the region as a joint arithmetic blocker.

---

# P6 — required correction magnitude at the final PS boundary

Estimate how much of the **final PS residual direction** must be removed to cross the real system threshold, without changing DA.

Use decoded/stage-aligned vector counterfactuals between actual `P_F` and canonical `P_Q`:

`P(alpha) = P_F + alpha * (P_Q - P_F)`, for `alpha in [0,1]`.

Use a bounded dyadic sweep first:

- 0;
- 1/16;
- 1/8;
- 1/4;
- 1/2;
- 1.

If the pass/fail transition lies between two tested values, bisect only that interval to estimate the smallest passing alpha to within `1/64`.

Important:

- materialize/interpolate in a way that preserves the accepted DA-input Level/Scale and quantify any interpolation materialization floor;
- alpha=0 must reproduce the actual control within the measured interpolation floor;
- alpha=1 must reproduce C_PS within the measured interpolation floor;
- if those endpoint controls fail, discard the alpha experiment rather than reporting a misleading threshold.

Record:

- smallest passing alpha;
- equivalent required reduction fraction of `P_F-P_Q`;
- public-like margin at the smallest passing alpha.

This is a design target, not a production fix.

---

# P7 — decision classification

Choose exactly one primary classification.

### One PS operation is sufficient

`logn13_ps_residual_single_operation_blocker`

Record:

- block;
- exact source-backed operation;
- smallest passing alpha from P6;
- downstream margin.

### One PS block is sufficient but no single operation is

`logn13_ps_residual_single_block_joint_arithmetic_blocker`

Record the block and its operation set.

### Residual is distributed across multiple PS blocks

`logn13_ps_residual_multi_block_blocker`

Use this only if no single block reset passes but a broader/final reset does.

### Final PS reset itself does not reproduce prior C_PS success

`logn13_ps_residual_counterfactual_mismatch`

Stop and report the mismatch.

No production fix is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-PS-RESIDUAL-LOCALIZATION-summary.json`

Include only:

- provenance;
- P0 controls;
- compact PS DAG/checkpoint table;
- compact actual-vs-canonical checkpoint residual table;
- incremental residual-growth table;
- checkpoint-reset downstream table;
- operation-reset table for the one implicated region only;
- alpha sensitivity table and smallest passing alpha if valid;
- classification;
- first remaining blocker;
- recommended next design target;
- validation flags.

Do not serialize slot vectors, coefficient arrays, powers, matrices, full RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*PS.*Residual.*Localization` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no DA/S2C/q/generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.