# FIX-001-P3-DIAG-P93-T3-GENERATED-POWER-CAUSAL-DECOMPOSITION

## Purpose

The preceding task at Primary commit

`e23337b8087cb8fec28d857be3bff2efe78f60e0`

is accepted.

It proved:

1. the historical/design Standard polynomial comparator is exactly equivalent to the Genuine Standard internal polynomial oracle for the deterministic workload;
2. fresh historical P93 polynomial accuracy is genuinely:
   - real `1.8570138426987626e-8`
   - imag `1.6219059983946238e-8`;
3. therefore historical P93 already satisfies the current derived polynomial-output budget:
   [
   E_{poly,target}=3.716228023823462e-8;
   ]
4. current dirty production polynomial output is about:
   - real `4.986373945969902e-7`
   - imag `5.07761187318323e-7`;
5. current-vs-historical polynomial regression vector closure passes;
6. the first generated-power checkpoint exceeding the derived budget is `T3-final`:
   - T2 current-vs-historical real: `6.940818919609626e-9` (within budget)
   - T3 current-vs-historical real: `2.2233773797064593e-7` (fails budget)
   - apparent growth ≈ 32x;
   - capacity remains safe.

However:

> A generated power exceeding the final polynomial-output budget does **not** by itself prove that power is the causal source of the final polynomial regression.

PS coefficients and later combinations can attenuate, amplify, or cancel intermediate errors.

Therefore this task must first causally decompose the T3 regression itself and determine whether it is:

- expected propagation of T1/T2 input error through the Chebyshev recurrence; or
- additional error introduced by the current production T3 generation path.

No production repair in this task.

---

## Repository safety

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary production:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact branch/HEAD/fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty production worktree.

No Secondary source modification, commit, or push is authorized.

Historical clean reference:
- Secondary commit `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- use isolated detached temporary worktree under `/tmp` when needed.

---

## Fixed controls

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012
- planScale = `2^93`
- deterministic 4096-slot workload
- same Mod1 polynomial input
- polynomial budget:
  `3.716228023823462e-8`

No tuning.

---

# R0 — reproduce T1/T2/T3 endpoints

For both real and imag, freshly reproduce in the same diagnostic run:

## Historical clean P93

- T1
- T2-final
- T3-final

## Current dirty production

- T1
- T2-final
- T3-final

Record decoded semantics and compact metadata.

Require current-vs-historical endpoint consistency with existing evidence.

At minimum:

- T2 real ~`6.94e-9`
- T3 real ~`2.22e-7`

If endpoints fail to reproduce, stop:

`P93_T3_REPLAY_CONFLICT`.

---

# R1 — prove the actual T3 recurrence

Read the common polynomial splitting logic and current/historical Fast source.

Record:

- `SplitDegree(3)` result;
- Chebyshev recurrence actually implemented;
- whether T3 is mathematically:
  [
  T_3 = 2T_2T_1-T_1
  ]
  for the active path;
- lazy/relinearization mode;
- balanced schedule selected yes/no;
- common level;
- balanced factor pair if selected;
- expected output scale.

If the implementation uses a different mathematically equivalent recurrence, use the actual recurrence in all later equations.

---

# R2 — plaintext T3 oracle validation

Build a plaintext semantic oracle using the exact recurrence established in R1.

For historical clean P93:

[
O_H = 2H_2H_1-H_1
]

For current production inputs:

[
O_P = 2P_2P_1-P_1.
]

Compare:

- historical actual T3 (H_3) vs (O_H);
- production actual T3 (P_3) vs (O_P).

Because CKKS/rescale introduces implementation noise, measure the oracle floor rather than assuming exact zero.

The historical oracle residual must be small relative to the observed current-vs-historical T3 regression and small enough for causal decomposition.

If not, stop:

`P93_T3_PLAINTEXT_ORACLE_MISMATCH`.

---

# R3 — causal decomposition of T3 regression

Define vectors:

[
Delta_{input}=O_P-O_H
]

[
Delta_{impl,P}=P_3-O_P
]

[
Delta_{impl,H}=H_3-O_H
]

and observed regression:

[
Delta_{obs}=P_3-H_3.
]

Then:

[
Delta_{obs}
=
Delta_{input}
+
(Delta_{impl,P}-Delta_{impl,H}).
]

Verify vector closure before reducing to norms.

Report for real and imag:

- propagated T1/T2 input contribution;
- net implementation-regression contribution;
- observed T3 regression;
- closure residual;
- contribution fractions.

Closure target:
<= `1e-10`, or justified oracle/replay floor.

Classification thresholds:

## Input-propagation dominated

Implementation-regression contribution <=10% of observed T3 regression.

## Implementation dominated

Implementation-regression contribution >50%.

## Mixed

Otherwise.

Do not infer causality from max-norm subtraction; use vector decomposition.

---

# R4 — if input-propagation dominated, recurse one level to T2

Only if R3 is input-propagation dominated:

Validate the active T2 recurrence, expected to be Chebyshev:

[
T_2=2T_1^2-1.
]

Build the analogous plaintext oracle for historical and production T1.

Decompose T2 regression into:

- T1 input propagation;
- historical implementation residual;
- production implementation residual.

Report whether T2 regression is implementation-generated even though its magnitude is below the final polynomial budget.

Stop after this T2 decomposition; do not repair.

---

# R5 — if T3 implementation effect is material, localize the first stable primitive boundary

Only if R3 is implementation-dominated or mixed with a material implementation component.

Trace both historical clean P93 and current production T3 source-faithfully through stable semantic boundaries.

At minimum:

1. T1 input
2. T2 input
3. common-level alignment
4. balanced schedule decision
5. balanced left operand after maintained integer scaling
6. balanced left post-Rescale
7. balanced right operand after maintained integer scaling
8. balanced right post-Rescale
9. product after Mul/MulRelin
10. Chebyshev doubling
11. subtraction/alignment of T1 difference power
12. T3-final

If a pre-Rescale semantic coordinate is not directly comparable, mark it `not_semantically_stable` and compare the next stable post-Rescale boundary.

For each stable checkpoint report:

- current-vs-historical max component
- max complex
- Level/Scale/Degree
- selected factors
- capacity
- comparison-valid flag.

Identify first operation where the **implementation-regression component** becomes material.

Do not use the final polynomial budget blindly as a primitive threshold; compare against the measured T3 implementation-regression magnitude from R3.

---

# R6 — source diff audit

Read-only compare the historical clean source and current dirty source relevant to T3:

At minimum inspect dirty changes affecting:

- `circuits/ckks/polynomial/fast.go`
- maintained integer multiplication
- multiplication/relinearization
- Rescale
- scale alignment helpers

Record only source changes that are on the executed T3 path.

For each relevant dirty change state:

- file/function
- behavioral difference
- whether R5 evidence points to that operation
- whether it is only a candidate or causally supported.

Do not propose a repair unless supported by R3/R5.

---

# R7 — establish whether T3 is likely relevant to final polynomial regression

Do **not** claim T3 is the final polynomial blocker merely because it is the first budget-breaking generated power.

Report a bounded relevance assessment using existing PS plan information:

- which baby/giant steps consume T3;
- coefficients/plan locations involving T3;
- whether the first current-vs-historical PS accumulation divergence above the polynomial budget is downstream of a T3-dependent branch;
- whether T3's observed regression can plausibly contribute at the measured final polynomial regression scale.

This is a dependency/scale check, not a replacement experiment.

Classification may still state `T3 causal contribution to final output not yet proven`.

---

# Decision classification

Choose exactly one primary classification:

## A — `P93_T3_REPLAY_CONFLICT`

Historical/current T3 endpoints do not reproduce.

## B — `P93_T3_PLAINTEXT_ORACLE_MISMATCH`

T3 recurrence oracle cannot support causal decomposition.

## C — `P93_T3_INPUT_PROPAGATION_DOMINATED`

Current-vs-historical T3 regression is >=90% explained by T1/T2 input propagation.

Include R4 T2 decomposition.

## D — `P93_T3_IMPLEMENTATION_REGRESSION`

>50% of T3 regression is current T3 implementation effect.

Include first stable primitive boundary from R5.

## E — `P93_T3_MIXED_REGRESSION`

Both input propagation and implementation regression materially contribute.

## F — `P93_T3_CAUSAL_DECOMPOSITION_ALIGNMENT_INVALID`

Vector/reference alignment is insufficient for causal interpretation.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-T3-GENERATED-POWER-CAUSAL-DECOMPOSITION-summary.json`

Keep <=300 pretty-printed JSON lines.

Include:
- provenance
- R0 T1/T2/T3 endpoint replay
- R1 recurrence/schedule
- R2 oracle fidelity
- R3 vector decomposition
- R4 T2 decomposition if triggered
- R5 primitive localization if triggered
- R6 relevant dirty-source audit
- R7 final-polynomial relevance
- one classification A-F
- explicit statement Secondary production worktree was not modified.

No full vectors or coefficient arrays.

---

## Validation

Run:
- T3 endpoint replay test
- T3 plaintext-oracle test
- vector-closure test
- conditional T2 oracle/decomposition test
- conditional T3 primitive trace test
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
- no DoubleAngle repair
- no C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is a causal decomposition of the first generated-power regression, not a fix.
