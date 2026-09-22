# FIX-001-P3-PROD-LOGN13-Q012-GENERATED-POWER-POSTPRODUCT-SCHEDULE-REPAIR

## Purpose

The diagnostic sequence has now causally identified the dominant generated-power precision failure.

Accepted evidence:

### Historical P93 design

Primary reference:
`41993cd03f85ec5f2de54dc21ade5ec812090d50`

Historical generated powers use:

[
	ext{direct Q012 multiply}
ightarrow
	ext{relinearize}
ightarrow
	ext{Chebyshev recurrence}
ightarrow
	ext{single post-product Q012 Rescale}.
]

For P93, all required generated powers:

[
T_2,T_3,T_4,T_6,T_8,T_{16}
]

were Q012-capacity safe and had near-oracle semantic accuracy.

Historical P93 polynomial output vs Genuine Standard is already reconciled and within the required polynomial budget:

- real `1.8570138426987626e-8`
- imag `1.6219059983946238e-8`

Required polynomial budget for the current public `1e-2` milestone:

[
3.716228023823462e-8.
]

### Current production failure

Current production uses balanced pre-Rescale generated-power scheduling:

[
	ext{scale operands}
ightarrow
	ext{Rescale operands to about }2^{30}
ightarrow
	ext{multiply}
ightarrow
	ext{recurrence}.
]

This introduces operand quantization residuals around `1e-8..1e-7`.

T3 causal decomposition proved >99.9% of its regression is implementation/schedule effect, not propagation of T1/T2 input error.

### q01/q012 primitive A/B

Primary:
`48a42226182088167eebdc822b5ffbe00d629efa`

Current Q012 Rescale vs clean q01 Rescale is not causal.
Do not repair Rescale.

### Schedule causal A/B

Primary:
`dd074420d81d2cd9ba73437864b76094ef735c59`

From the exact same current T1/T2:

- current balanced T3 residual:
  ~`1.86e-7..2.23e-7`
- current-primitives post-product T3 residual:
  ~`1e-16`
- historical/big-int post-product oracle:
  ~`1e-16`
- current-primitives post-product path is Q012-capacity safe;
- bounded T6 relevance improves:
  - real `6.2948e-8 -> 1.7304e-8`
  - imag `5.2192e-8 -> 1.8343e-8`

Both corrected T6 branches re-enter the polynomial budget.

Therefore:

[
oxed{	ext{balanced pre-Rescale generated-power scheduling is causal}}
]

and current fixed-width Q012 primitives are sufficient for the post-product schedule.

This task is authorized to implement the smallest production repair in the **already-proven LogN13/P93/Q012 generated-power domain**, then validate the complete pipeline.

---

## Repository safety

### Primary

Repository:
`xuejin-lu/heart-lattigo-bootstrap`

Follow `AGENTS.md` startup preflight, synchronize `main`, then read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary

Repository:
`xuejin-lu/lattigo`

Required starting state:

- branch: `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- required starting dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact match **before modification**.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, or reconstruct the Secondary worktree.

### Secondary modification authorization

This task explicitly authorizes a minimal local Secondary modification for the generated-power schedule repair.

Allowed production source scope:

- `circuits/ckks/polynomial/fast.go`

Allowed Secondary test scope, only if needed:

- directly relevant polynomial Fast tests.

Do not modify unrelated Secondary files.

### Secondary commit/push prohibition

Even if the repair passes:

- **do not commit Secondary**
- **do not push Secondary**

Keep the repaired Secondary worktree dirty.

At task completion record:

- new dirty diff stat
- new dirty diff SHA-256.

Primary diagnostic/tests/evidence may be committed and pushed according to `AGENTS.md`.

---

# Fixed target domain

The repair is not a generic rewrite of every Fast polynomial mode.

It is for the already-proven production candidate:

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- P93 / `planScale = 2^93`
- Q012-authoritative generated powers
- `LevelsConsumedPerRescaling() == 1`
- deterministic 4096-slot validation workload

Preserve the existing balanced schedule as fallback outside the proven Q012 domain.

Do not broaden the repair beyond the minimum predicate justified by the actual source architecture.

The implementation should prefer an architectural predicate such as:

- active maintained-limb count is 3 at the generated-power level;
- the current plan is the plan-scale/Q012 production path;
- LogN13/current supported fixed profile where required;

rather than an unconditional global schedule replacement.

If the exact source does not expose enough context to make this predicate safely, stop and report rather than broadening behavior silently.

---

# R0 — pre-change production replay

Before editing Secondary, reproduce the accepted current controls:

- T3 balanced residual;
- T6 current residual;
- polynomial output:
  - real ~`4.986e-7`
  - imag ~`5.078e-7`
- exact current Fast E2E baseline.

Record the actual baseline fresh in this run.

Also preserve the starting Secondary fingerprint.

---

# R1 — implement the generated-power schedule repair

In the proven Q012 domain, generated-power recurrence must follow this ordering:

1. recursively produce powers (T_a,T_b);
2. align them to the common level;
3. **do not balanced-pre-Rescale either operand**;
4. multiply at full operand scales:
   - `Mul` for lazy path when semantically permitted;
   - `MulRelin` for strict path;
5. if required, relinearize at the same logical point as the historical/proven schedule;
6. perform Chebyshev recurrence at the full product scale:
   [
   T_n=2T_aT_b-T_{|a-b|}
   ]
   or (-1) when (a=b);
7. only after the full recurrence is formed, perform **one Rescale**;
8. store the resulting generated power.

Critical:

> Do not implement the repair by merely forcing the old non-balanced branch if that branch performs Rescale before the Chebyshev subtraction.

The repaired Q012 path must be semantically:

[
oxed{	ext{multiply}ightarrow	ext{full recurrence}ightarrow	ext{Rescale}}
]

matching the causally validated A/B and historical design.

Preserve:

- lazy/strict degree semantics;
- relinearization contract;
- scale alignment semantics for the difference power;
- maintained q0/q1/q2 representation;
- existing fallback behavior outside the repaired domain.

No q/planScale changes.

---

# R2 — source-level regression tests

Add the smallest tests necessary to prove:

1. repaired Q012 path selects post-product scheduling;
2. fallback/non-Q012 path still uses existing behavior;
3. Chebyshev subtraction occurs before the repaired generated-power Rescale;
4. output Level/Scale/Degree match the historical post-product contract;
5. no extra Rescale occurs.

Do not overfit tests to one slot vector only.

---

# R3 — generated-power semantic validation

For real and imag, run the repaired production path and validate:

[
T_2,T_3,T_4,T_6,T_8,T_{16}.
]

For each record:

- Level/Scale/Degree;
- semantic error vs exact Chebyshev plaintext power oracle;
- q01 capacity;
- q012 capacity;
- remaining Q012 bits.

All must remain Q012 centered-unique.

Primary accuracy requirement:

Every generated power must be below:

[
3.716228023823462e-8.
]

Expected from historical design is much better, approximately `1e-14` or below.

Also compare repaired powers vs historical P93 outputs where semantically aligned.

---

# R4 — polynomial output validation

Run the full repaired P93 polynomial evaluation.

Compare against Genuine Standard internal polynomial output.

Record real/imag:

- max component
- max complex
- mean complex
- worst slot/component
- Level/Scale/Degree
- PS-exit Q012-to-Q01 contraction evidence.

Required:

[
E_{poly}le3.716228023823462e-8.
]

Historical target context:

- real ~`1.857e-8`
- imag ~`1.622e-8`.

If polynomial still fails budget, stop before claiming system sufficiency and localize the new first material checkpoint using existing diagnostic infrastructure.

Do not tune.

---

# R5 — DoubleAngle / internal EvalMod validation

Using the real Genuine Standard internal oracle:

Record repaired Fast-vs-Genuine Standard:

- polynomial output
- DA0 post-Rescale
- DA1 post-Rescale
- DA2 post-Rescale
- internal final before public reset.

Use only the previously validated stable semantic checkpoints.

Required DA2 pre-final-restore budget:

[
E_{DA2}le3.0517578125e-7.
]

Required internal final budget:

[
E_{internal}le3.125e-4.
]

Do not change the Standard public Scale contract.

---

# R6 — public EvalMod validation

Apply the unchanged Genuine Standard-compatible public DefaultScale reset.

Compare repaired Fast public EvalMod vs Genuine Standard public EvalMod.

Record real/imag max-component errors.

Required:

[
E_{public,EvalMod}le1e-2.
]

Do not preserve diagnostic (2^{50}) metadata through the public boundary.

---

# R7 — exact full LogN13 E2E

Run the complete actual repaired production bootstrap:

[
	ext{Pack}
ightarrow
	ext{ScaleDown}
ightarrow
	ext{ModUp}
ightarrow
	ext{C2S}
ightarrow
	ext{repaired Fast EvalMod}
ightarrow
	ext{S2C}
ightarrow
	ext{Unpack}
ightarrow
	ext{public finalization}.
]

Decode against the original deterministic message.

Use:

[
E_{max}
=
max_i
{
|Re(hat m_i-m_i)|,
|Im(hat m_i-m_i)|
}.
]

Record:

- max component E2E
- max complex
- mean complex
- worst slot/component
- final Level/Scale/Degree.

Primary milestone:

[
oxed{E_{max}le1e-2}
]

Also report Genuine Standard exact E2E control.

Do not use matched proxy metrics as E2E.

---

# R8 — no-regression validation

Run at minimum:

## Secondary

- directly affected Fast polynomial tests
- directly affected Fast evaluator/rescale tests
- `go test ./circuits/ckks/polynomial/...`
- `go test ./schemes/ckks/fast/...`
- broader `go test ./...` if feasible in the existing repository environment

## Primary

- generated-power semantic tests
- P93 polynomial reference tests
- DoubleAngle decomposition controls
- exact E2E test
- `go test ./...`
- `git diff --check`

If an unrelated pre-existing test fails, distinguish it explicitly from repair-caused failure.

---

# R9 — final Secondary dirty-state evidence

At completion record:

- Secondary branch
- committed HEAD
- new dirty diff stat
- new dirty diff SHA-256
- exact files modified by this task on top of the prior dirty state.

Do not commit or push Secondary.

This new fingerprint is required for the next orchestrator review.

---

# Decision classification

Choose exactly one:

## A — `P93_POSTPRODUCT_REPAIR_IMPLEMENTATION_BLOCKED`

The proven domain cannot be selected safely without broadening unrelated behavior.

## B — `P93_POSTPRODUCT_REPAIR_GENERATED_POWER_FAIL`

Repair builds, but one or more generated powers fail precision/capacity.

Record first failing power.

## C — `P93_POSTPRODUCT_REPAIR_POLYNOMIAL_FAIL`

Generated powers pass, but polynomial output remains above `3.716228023823462e-8`.

Record first downstream material checkpoint.

## D — `P93_POSTPRODUCT_REPAIR_EVALMOD_FAIL`

Polynomial passes, but DA2/internal/public EvalMod fails its derived budget.

## E — `P93_POSTPRODUCT_REPAIR_E2E_FAIL`

Public EvalMod passes, but exact full E2E remains > `1e-2`.

Record first downstream boundary.

## F — `P93_POSTPRODUCT_REPAIR_1E2_SYSTEM_PASS`

All hold:

1. all required generated powers pass precision and capacity;
2. polynomial output <= `3.716228023823462e-8`;
3. DA2 <= `3.0517578125e-7`;
4. internal final <= `3.125e-4`;
5. public EvalMod <= `1e-2`;
6. exact full LogN13 E2E <= `1e-2`;
7. validation tests pass;
8. no unrelated production parameters changed.

If F is selected, the next task may commit/push the repaired Secondary candidate after final diff review.

---

## Required Primary artifact

Create:

`results/FIX-001-P3-PROD-LOGN13-Q012-GENERATED-POWER-POSTPRODUCT-SCHEDULE-REPAIR-summary.json`

Keep <=350 pretty-printed JSON lines.

Include:

- Primary provenance
- starting Secondary provenance/fingerprint
- exact Secondary source changes made by this task
- repaired-domain predicate
- generated-power table
- polynomial metrics
- stable DA/internal metrics
- public EvalMod metrics
- exact E2E
- tests
- ending Secondary dirty fingerprint
- one classification A-F
- explicit statement that Secondary was not committed/pushed.

No full vectors or coefficient dumps.

---

## Prohibitions

- no Secondary reset/stash/clean/discard
- no Secondary commit
- no Secondary push
- no q/planScale tuning
- no public Scale-contract modification
- no Rescale primitive rewrite
- no C2S/S2C/finalizer repair
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

This task is the minimal causally authorized generated-power schedule repair plus complete correctness validation.
