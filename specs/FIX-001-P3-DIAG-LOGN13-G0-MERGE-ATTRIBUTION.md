# FIX-001-P3-DIAG-LOGN13-G0-MERGE-ATTRIBUTION

## Purpose

Determine whether the final LogN13 PS residual that becomes removable at `G0-add` is caused primarily by:

1. the G0 parent branch `a`;
2. the G0 product branch `b * T8`;
3. the G0 scale-alignment/add semantics themselves;
4. or a joint cancellation/interaction of both branches.

Accepted evidence at Primary commit `13d54221f7feac63c860b3bb6cf81227ceb65240`:

- actual public-like error = `0.010683237260415292`;
- canonical final PS replacement (`C_PS`) = `0.0003910748686204446`;
- only about `0.0458984375` of the final residual direction must be removed to cross `1e-2`;
- `G0-multiply` checkpoint reset still fails downstream (`public-like ~0.01326`);
- `G0-add` checkpoint reset passes downstream (`public-like ~0.00526`);
- resetting the whole G0-add output is broad: it replaces all upstream residual arriving from both merge operands, so it does not by itself prove the Add operation is the root cause;
- current Fast source has a special `planScaleOverride` path in `fastPolynomialWorkspace.evaluateMonomial`:

  ```go
  if ws.planScaleOverride && !a.Scale.Equal(b.Scale) {
      b.Scale = a.Scale
      return eval.Add(b, a, b)
  }
  ```

  i.e. metadata-only scale relabeling is performed before the merge;
- `rlwe.Scale.BigInt()` rounds to nearest integer; do not treat it as a truncation-to-floor mechanism in this task.

This task is diagnostic only. It must identify the smallest causal side of the G0 merge. No production changes are authorized.

---

## Required provenance

Primary required base:

`13d54221f7feac63c860b3bb6cf81227ceb65240`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Keep fixed:

- oracle powers;
- q0/q1/q2 = 56/39/40;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DA local-q2 behavior;
- common plan scale `2^92`;
- polynomial coefficients/PS decomposition;
- S2C/final restore behavior;
- genuine Standard widened-profile reference.

Do not:

- retune guards, scales, K values or constants;
- change q parameters;
- redesign generated powers;
- modify Secondary production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary-only diagnostic state capture, canonical checkpoint construction, full-RNS mirrors and branch replacement are allowed.

---

# M0 — reproduce accepted control

Reproduce:

- actual PS final residual vs canonical;
- actual EvalMod/post-S2C/public-like;
- G0-multiply reset downstream failure;
- G0-add reset downstream pass;
- smallest passing final residual alpha `0.0458984375` within the prior interpolation tolerance.

If any control differs materially, stop:

`logn13_g0_merge_attribution_precondition_mismatch`.

---

# M1 — capture exact G0 merge operands

At the source-backed G0 merge immediately before `G0-add`, capture separate diagnostic snapshots:

- `A_F`: actual parent branch `a` before addition;
- `B_F_raw`: actual `b` after G0 rescale and before multiplication by `T8`;
- `P_F`: actual product branch after `b * T8` and immediately before addition;
- `O_F`: actual current G0-add output.

Record for each:

- Level;
- exact Scale;
- Degree;
- NTT/Montgomery state;
- q0/q1 intended centered-capacity ratio;
- if a q2 diagnostic mirror exists, Q012 ratio;
- decoded max magnitude.

For the merge record:

- `A_F.Scale`;
- `P_F.Scale`;
- exact ratio both ways;
- `Log2Delta`;
- which branch current `planScaleOverride` relabels;
- exact metadata value before and after the relabel.

Do not infer values from block-level planned scales; capture the live source operands.

---

# M2 — canonical operand states

Construct canonical CKKS states at the exact native metadata of each operand:

- `A_Q`: canonical state for the exact partial PS expression represented by `A_F`;
- `P_Q`: canonical state for the exact partial PS expression represented by `P_F`.

Use the already accepted quantization-aware method:

- high-precision mathematical partial expression;
- deterministic >53-bit CKKS materialization;
- exact same Level/Scale/degree/dimensions/NTT metadata as the corresponding actual operand.

Record:

- `A_F-A_Q`;
- `P_F-P_Q`;
- canonical materialization floors;
- worst slot/component;
- direction similarity between both residual vectors.

If either partial-expression mapping is ambiguous from the source PS DAG, stop rather than guessing:

`logn13_g0_merge_operand_oracle_mismatch`.

---

# M3 — isolate merge-alignment semantics

Using the **same actual operand values** `A_F` and `P_F`, compare these merge implementations without changing any upstream arithmetic:

### S0 current Fast merge

Exactly reproduce the current `planScaleOverride` metadata-only relabel + Fast Add.

### S1 full-RNS mirror of current Fast merge

Perform the same metadata relabel and addition under stage-aligned full-RNS storage.

Purpose: prove whether q0/q1 Fast storage itself contributes anything at the merge.

### S2 Standard CKKS Add semantics

Using stage-aligned full-RNS clones of the same operand states and their **native pre-merge scales**, apply the genuine Standard `ckks.Evaluator.Add` behavior and record its resulting scale/value.

Do not silently substitute a different PS schedule. This is only an operator-semantics comparison on the captured G0 operands.

### S3 ideal decoded addition reference

Decode `A_F` and `P_F` at their own native scales and add the represented values directly in high precision. This is not an executable CKKS candidate; it is the semantic reference for the merge.

For S0/S1/S2 record:

- output vs S3 semantic error;
- output vs canonical G0 merge checkpoint;
- exact scale metadata;
- q0/q1/full-RNS row relationships where meaningful;
- capacity.

Interpretation:

- if S0 == S1 but S2 is materially closer to S3/canonical, the issue is Fast merge scale policy, not q0/q1 storage;
- if S0 ~= S2, the merge itself is not the primary cause and upstream branch residual must dominate;
- if S0 differs from S1, classify Fast merge implementation mismatch.

No production modification is authorized here.

---

# M4 — branch replacement counterfactuals

Replay the PS suffix from immediately before G0-add with the current accepted Fast merge and all downstream arithmetic unchanged.

Evaluate exactly four cases:

### C00 actual

`merge(A_F, P_F)`.

Must reproduce actual downstream public-like `~0.01068323726`.

### CA parent-only correction

`merge(A_Q, P_F)`.

### CP product-only correction

`merge(A_F, P_Q)`.

### CAP both branches canonical

`merge(A_Q, P_Q)`.

For each case record:

- G0-add output vs canonical G0 checkpoint;
- final PS output vs canonical `P_Q_final`;
- EvalMod real/imag vs Standard;
- post-S2C max-component vs Standard;
- public-like max-component vs message;
- pass/fail at `1e-2`;
- all capacity/row/metadata contracts.

Use the exact current merge semantics first. Do not mix S2 semantics into these branch-causality cases.

---

# M5 — branch residual sensitivity

For each branch independently, interpolate only that branch toward canonical while leaving the other branch actual:

- `A(alpha) = A_F + alpha*(A_Q-A_F)`;
- `P(alpha) = P_F + alpha*(P_Q-P_F)`.

Use diagnostic materialization at the native branch metadata.

Test alpha:

- 0;
- 1/16;
- 1/8;
- 1/4;
- 1/2;
- 1.

If a branch has a pass/fail transition, bisect to within `1/64` and record the smallest passing alpha.

Endpoint requirements:

- alpha=0 reproduces C00 within measured interpolation floor;
- alpha=1 reproduces CA or CP respectively.

If endpoint controls fail, invalidate that branch sensitivity result.

This determines whether the real system needs only a tiny improvement on one branch or a broad branch rewrite.

---

# M6 — decision classification

Choose exactly one primary classification.

### Parent branch alone is sufficient

If `CA <= 1e-2` and `CP > 1e-2`:

`logn13_g0_merge_parent_branch_blocker`

Record the smallest passing parent alpha.

### Product branch alone is sufficient

If `CP <= 1e-2` and `CA > 1e-2`:

`logn13_g0_merge_product_branch_blocker`

Record the smallest passing product alpha.

### Either branch alone is sufficient

If both CA and CP pass:

`logn13_g0_merge_multiple_single_branch_options`

Prefer the branch with lower required alpha and smaller source scope.

### Both branches are jointly required

If CA and CP fail but CAP passes:

`logn13_g0_merge_joint_branch_blocker`

### Merge scale policy itself is the actionable blocker

If S0/S1 agree, S2 is materially closer to the semantic/canonical merge, and switching only the merge policy (with actual A_F/P_F) produces a downstream pass in a diagnostic replay:

`logn13_g0_merge_scale_alignment_blocker`

A downstream S2-policy replay may be added only after S2 operator fidelity is established; keep all other arithmetic unchanged.

### Fast q0/q1 merge implementation mismatch

If S0 and S1 differ materially:

`logn13_g0_merge_fast_implementation_mismatch`

### Neither branch nor merge explains prior G0-add reset

`logn13_g0_merge_attribution_mismatch`

No production fix is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-G0-MERGE-ATTRIBUTION-summary.json`

Include only:

- provenance;
- M0 controls;
- live G0 operand metadata/scales/capacities;
- A_F-A_Q and P_F-P_Q metrics;
- S0/S1/S2/S3 merge-semantics comparison;
- C00/CA/CP/CAP downstream table;
- branch alpha sensitivity table(s);
- classification;
- first remaining blocker;
- recommended next design target;
- validation flags.

Do not serialize slot vectors, coefficient arrays, powers, complete RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*G0.*Merge.*Attribution` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no accepted arithmetic candidate changes;
- no DA/S2C/q/generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.