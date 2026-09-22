# FIX-001-P3-DIAG-P93-T3-PRIMITIVE-SEMANTIC-RESIDUAL-LOCALIZATION

## Purpose

The preceding T3 causal decomposition at Primary commit

`bd1a001017ea5f32e46d5fcebc3808fd4d3bbc07`

is accepted.

It established for current dirty production vs historical clean P93:

### Real

- T1 regression: `0`
- T2 regression: `6.940818919609626e-9`
- T3 observed regression: `2.2233773797064593e-7`
- T1/T2 input-error propagation into T3:
  `2.16898193849957e-10`
- net T3 implementation-regression component:
  `2.225546361644959e-7`

### Imag

- T1 regression: `0`
- T2 regression: `5.797987645550506e-9`
- T3 observed regression: `1.855698303146469e-7`
- input-error propagation:
  `1.8118786332399495e-10`
- net T3 implementation-regression:
  `1.857510181779709e-7`

Vector closure passes to numerical floor.

Therefore the T3 regression is overwhelmingly an **implementation regression**, not normal propagation of T1/T2 input error.

The exact recurrence is:

[
T_3=2T_2T_1-T_1.
]

The active T3 path uses a balanced schedule and the current dirty source changes on the executed path include:

- maintained q2 copies/workspace;
- q0/q1/q2 `MulIntegerMaintained`;
- maintained q0/q1/q2 multiplication;
- Q012 rescale.

However, the preceding task could only identify `T3-final` as the first stable material boundary. It did **not** causally identify which primitive creates the implementation residual.

This task must source-faithfully replay the current T3 generation from the exact current T1/T2 inputs and compare every **semantically stable** primitive boundary against a plaintext oracle.

No production repair in this task.

---

## Repository safety

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact branch/HEAD/fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary production worktree.

No Secondary production source modification, commit, or push is authorized.

---

## Fixed controls

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012
- planScale = `2^93`
- deterministic 4096-slot workload
- same current production T1/T2 inputs
- T3 recurrence:
  [
  2T_2T_1-T_1
  ]
- polynomial-output budget:
  `3.716228023823462e-8`
- measured T3 implementation residual:
  roughly `1.86e-7..2.23e-7`

No parameter tuning.

---

# R0 — source-faithful current T3 shadow replay

Build a diagnostic-only T3 shadow in Primary using the same exported Fast evaluator primitives and the exact current production T1/T2 ciphertexts.

It must reproduce the executed `fastPowerBasis.genPowerInternal(3,...)` path exactly:

1. identify `SplitDegree(3)`;
2. select current production T1/T2;
3. align to common level;
4. derive the same balanced factor pair;
5. make maintained left/right copies;
6. apply the same maintained integer factors;
7. apply the same operand Rescales;
8. perform the same Mul/MulRelin/lazy behavior;
9. apply Chebyshev doubling;
10. apply the same T1 subtraction/alignment;
11. produce T3 endpoint.

Do not change any production parameter or primitive.

Require shadow T3 endpoint vs actual production T3:

- decoded max component <= `1e-10`, preferably exact within deterministic floor;
- Level/Scale/Degree agreement;
- representation metadata agreement.

If not, stop:

`P93_T3_SHADOW_REPLAY_CONFLICT`.

---

# R1 — define plaintext semantic oracle at each stable stage

Decode the exact current T1/T2 inputs:

[
x_1=P_1,qquad x_2=P_2.
]

Define plaintext expectations:

## Input

[
L_{in}=x_2,qquad R_{in}=x_1.
]

## Balanced left/right post-Rescale

Maintained integer scaling plus matching Scale metadata plus Rescale is intended to preserve plaintext semantics:

[
L_{post}=x_2,
qquad
R_{post}=x_1.
]

## Product

[
M=x_2x_1.
]

## Chebyshev doubling

[
D=2x_2x_1.
]

## Final subtraction

[
T_3^{oracle}=2x_2x_1-x_1.
]

Use actual decoded current T1/T2 vectors; do not substitute historical inputs.

---

# R2 — stable-boundary residual trace

For real and imag, capture and compare the current shadow path at:

1. T2 input
2. T1 input
3. left maintained copy before factor — only if semantic decode is stable
4. right maintained copy before factor — only if stable
5. left post-`MulIntegerMaintained + Scale metadata + Rescale`
6. right post-`MulIntegerMaintained + Scale metadata + Rescale`
7. product after Mul/MulRelin, before Chebyshev doubling, if semantically decodable
8. after Chebyshev doubling
9. after T1 subtraction/alignment
10. final shadow T3 endpoint

Do **not** use pre-Rescale maintained-scalar states for causal norms unless their semantic coordinate is explicitly validated.

For each valid checkpoint record:

- actual decoded semantic vs plaintext expectation;
- max component residual;
- max complex residual;
- mean complex residual;
- worst slot/component;
- Level/Scale/Degree;
- q012 capacity;
- comparison-valid flag.

The metric here is:

[
E_{stage}=|C_{stage}-O_{stage}|_{infty, component}.
]

This is an absolute implementation residual against the recurrence semantics, not current-vs-historical difference.

---

# R3 — historical clean validation of the same stable oracle

The historical clean P93 actual T3 endpoint already matches the plaintext recurrence at roughly `1e-16`.

For the stable primitive boundaries identified in R2, obtain historical clean evidence only as needed to validate that the plaintext oracle is the correct semantic interpretation.

Use isolated temporary reference tooling/worktrees. Do not touch the dirty production worktree.

At minimum require:

- historical left/right post-Rescale residual negligible relative to current T3 implementation residual;
- historical product/doubling/final subtraction residual negligible wherever those states are captured.

If historical instrumentation cannot expose a stable boundary, the current plaintext semantic equation may still be used if directly justified by CKKS operation semantics and validated at adjacent boundaries.

Do not force raw-coordinate comparison.

---

# R4 — locate the first material implementation residual

Define materiality relative to the measured production T3 implementation residual.

For each branch:

[
E_{impl,T3}approx
|P_3-O_P|.
]

A stable primitive boundary is material if its residual reaches at least:

[
0.1	imes E_{impl,T3}.
]

Identify the first material residual among:

- left operand post-Rescale
- right operand post-Rescale
- product
- doubling
- subtraction/alignment

For the transition where the residual first becomes material report:

- incoming residual;
- outgoing residual;
- amplification;
- exact primitive(s);
- scale transition;
- capacity;
- row/metadata facts needed for interpretation.

Do not use the final polynomial budget as the primitive-level materiality threshold.

---

# R5 — special decomposition if operand post-Rescale is first material boundary

If either balanced operand post-Rescale is the first material boundary, isolate:

[
	ext{copy}
ightarrow
	ext{MulIntegerMaintained}
ightarrow
	ext{Scale assignment}
ightarrow
	ext{Rescale}.
]

Because pre-Rescale decode may be unstable, use two causal checks:

1. **post-Rescale semantic residual** against unchanged input semantics;
2. a diagnostic-only q01-vs-q012 primitive shadow, if available without modifying Secondary production, to identify whether the regression is associated with:
   - maintained q2 integer scaling;
   - Q012 rescale;
   - their interaction.

This is diagnostic-only. No repair.

Do not perform an open-ended variant matrix.

---

# R6 — special decomposition if product is first material boundary

If both operand post-Rescale residuals are below material threshold but product becomes material, compare:

[
C_{mul}
]

against:

[
O_{mul}=L_{decoded}R_{decoded}.
]

Then audit the executed dirty multiplication path:

- Mul vs MulRelin
- maintained q012 limb behavior
- degree/lazy behavior
- any immediate metadata normalization.

A diagnostic q01-vs-q012 multiplication shadow is allowed only if it is a single bounded A/B needed to establish causality.

No repair.

---

# R7 — special decomposition if subtraction/alignment is first material boundary

If product and doubling remain accurate but final subtraction becomes material, isolate `subAligned`:

- scale ratio derivation;
- integer ratio rounding;
- `MulIntegerMaintained`;
- Scale reassignment;
- subtraction.

Compare the result against:

[
2x_2x_1-x_1.
]

No repair.

---

# R8 — final-polynomial relevance remains separate

Even if a primitive causing T3 implementation residual is proven, do not yet claim it is sufficient to explain the final polynomial regression.

Record only:

- T3 feeds T6;
- T6 is consumed by the active even PS plan;
- final causal sufficiency still requires a later T3/T6/PS counterfactual or propagation proof.

No replacement experiment in this task.

---

# Decision classification

Choose exactly one:

## A — `P93_T3_SHADOW_REPLAY_CONFLICT`

Diagnostic shadow does not reproduce production T3.

## B — `P93_T3_OPERAND_RESCALE_REGRESSION`

First material implementation residual appears in balanced left/right operand post-Rescale.

State left/right/both and whether q012 rescale interaction is supported.

## C — `P93_T3_MULTIPLICATION_REGRESSION`

Operands remain semantically accurate; product is first material residual.

## D — `P93_T3_DOUBLING_REGRESSION`

Product is accurate; Chebyshev doubling first becomes material.

## E — `P93_T3_SUBTRACTION_ALIGNMENT_REGRESSION`

Doubling is accurate; T1 subtraction/alignment first becomes material.

## F — `P93_T3_PRIMITIVE_MIXED_REGRESSION`

Multiple primitive boundaries contribute materially without one dominant first source.

## G — `P93_T3_PRIMITIVE_SEMANTIC_ALIGNMENT_INVALID`

Stable semantic oracle cannot be validated sufficiently.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-T3-PRIMITIVE-SEMANTIC-RESIDUAL-LOCALIZATION-summary.json`

Keep <=300 pretty-printed JSON lines.

Include:

- provenance
- R0 shadow replay
- R1 semantic equations
- compact stable residual table
- historical oracle validation
- first material primitive boundary
- conditional R5/R6/R7 evidence
- source audit
- R8 relevance statement
- one classification A-G
- explicit confirmation Secondary production was not modified.

No full vectors, coefficient arrays, or unbounded trace dumps.

---

## Validation

Run:

- T3 shadow-vs-production replay test
- stable semantic-oracle tests
- historical oracle validation where instrumented
- first-material-boundary test
- any single bounded q01/q012 A/B test triggered by classification
- directly affected Primary tests
- `go test ./...`
- `git diff --check`

On success commit/push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no T3 repair
- no T2 repair
- no generated-power replacement sweep
- no PS repair
- no q/planScale tuning
- no DoubleAngle/C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is the first semantically stable T3 primitive that creates the proven implementation residual, not a fix.
