# FIX-001-P3-DIAG-P93-POLYNOMIAL-REFERENCE-RECONCILIATION-AND-PS-REGRESSION-LOCALIZATION

## Purpose

The preceding DoubleAngle causal decomposition at Primary commit

`a6b541577eaa3ed33238a4e907fd0c727dc7035f`

is accepted.

It established that Fast DoubleAngle is **not** the dominant blocker:

- vector closure passes to numerical floor;
- Fast DA implementation effect is only about 0.36%–0.54% of observed post-Rescale DA error;
- more than 99% of DA2 error is explained by normal propagation of the polynomial-output error.

The empirically required polynomial-output accuracy for the current public `1e-2` milestone is approximately:

[
E_{poly,target}approx3.72	imes10^{-8}.
]

Current dirty production polynomial output is approximately:

- real `4.986373945969902e-7`
- imag `5.07761187318323e-7`

and therefore fails that derived budget by about 13–14x.

Historical fresh P93 design evidence recorded polynomial errors:

- real `1.8570138426987626e-8`
- imag `1.6219059983946238e-8`

which are already below the newly derived budget.

However, those historical values were measured against the older matched/design Standard polynomial reference. Before treating them as proof that the P93 design already meets the Genuine Standard polynomial target, this task must reconcile that polynomial reference with the **Genuine Standard internal polynomial output**, then localize the current production regression relative to the historical P93 design.

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
- required dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact branch/HEAD/fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed controls

Authoritative:

- Genuine Standard exact E2E:
  `5.830057349387463e-8`
- Fast/Genuine C2S:
  ~`1e-13`
- DoubleAngle implementation contribution:
  <1% of observed DA error
- derived polynomial-output budget:
  `3.716228023823462e-8` (use the stricter imag/worst value)
- current production polynomial error:
  ~`5e-7`
- historical P93 design polynomial error:
  ~`1.6e-8..1.9e-8`

No q/planScale tuning.

---

# R0 — Genuine Standard polynomial oracle

Using the same deterministic input and Genuine Standard evaluator, capture the Standard internal polynomial output:

- after Mod1 preprocessing/offset;
- after Standard polynomial evaluator;
- before any DoubleAngle operation.

Record for real and imag:

- Level
- Scale
- Degree
- decoded semantic summary
- source lineage

Prove the Standard replay aligns with actual Genuine Standard EvalMod internal execution.

---

# R1 — fresh historical P93 design replay at polynomial boundary

Reuse the isolated clean-reference methodology from the earlier P93 replay:

Historical Primary reference:
`41993cd03f85ec5f2de54dc21ade5ec812090d50`

Historical clean Secondary:
`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Use detached temporary worktrees under `/tmp`; never touch the dirty production Secondary.

Freshly replay P93 only.

Capture:

- P93 design Fast polynomial output real/imag;
- the historical/design Standard polynomial comparator used for the recorded `~1.8e-8` metric.

Do not merely read the old JSON.

---

# R2 — reconcile historical Standard polynomial comparator with Genuine Standard

Compare:

[
S_{hist-poly}
]

vs

[
S_{genuine-poly}.
]

Required:

- max component
- max complex
- mean complex
- worst slot/component
- Level/Scale/Degree
- exact input lineage

If their semantic difference is <= `1e-10`, classify the historical polynomial metric as directly comparable to Genuine Standard.

If not, stop with:

`P93_HISTORICAL_POLYNOMIAL_REFERENCE_MISMATCH`.

Do not use the old `1.8e-8` value as a target until this passes.

---

# R3 — verify historical P93 design against Genuine Standard polynomial oracle

If R2 passes, compare fresh historical P93 Fast polynomial output directly against Genuine Standard polynomial output.

Expected context:

- real ~`1.857e-8`
- imag ~`1.622e-8`

Require both:

[
E_{poly}le3.716228023823462e-8.
]

If fresh historical P93 no longer meets the budget, classify:

`P93_HISTORICAL_DESIGN_POLYNOMIAL_BUDGET_CONFLICT`.

If it passes, establish:

> the historical P93 design arithmetic already had sufficient polynomial accuracy for the current public `1e-2` target.

---

# R4 — current production polynomial regression

From the unchanged dirty production path, capture the actual current Fast polynomial output at the same semantic boundary.

Compare:

1. current production Fast vs Genuine Standard;
2. current production Fast vs historical P93 Fast design;
3. historical P93 Fast design vs Genuine Standard.

Record real/imag metrics.

Expected production-vs-historical design context:
about `5.03e-7` at final polynomial/PS output.

Define regression vector:

[
Delta_{reg}=F_{prod-poly}-F_{hist-poly}.
]

Verify vector closure:

[
(F_{prod-poly}-F_{hist-poly})+
(F_{hist-poly}-S_{genuine-poly})
=
F_{prod-poly}-S_{genuine-poly}.
]

Closure target <= `1e-10` or justified numerical floor.

---

# R5 — PS/generated-power localization

Only after R2–R4 pass, compare current production vs historical P93 design at semantically equivalent PS checkpoints.

At minimum, where valid:

- polynomial input
- T2
- T3
- T4
- T6
- T8
- T16
- each material baby/giant-step boundary available in both implementations
- final polynomial root before final Rescale
- final polynomial output after Rescale

Do not force raw row equality if the historical design representation differs.

For each comparable checkpoint record:

- production-vs-historical semantic difference
- production-vs-Genuine Standard metric if available
- historical-vs-Genuine Standard metric if available
- Level/Scale/Degree
- current production capacity
- comparison-valid flag

Use this materiality threshold:

[
3.716228023823462e-8.
]

The first checkpoint exceeding this threshold is the **first budget-breaking regression**, not automatically the causal source-code bug.

Historical context only:
- T2 difference ~`6e-9`
- T3 difference ~`1.9e-7..2.2e-7`
- T4/T6 ~`5e-8..6e-8`
- T8 ~`2.3e-7`
- T16 ~`9.1e-7`

Fresh measurement is authoritative.

---

# R6 — distinguish generated-power regression from PS accumulation regression

For the first budget-breaking checkpoint, determine whether it is:

## Generated-power boundary

A generated power itself already exceeds the polynomial budget.

or

## PS accumulation boundary

Generated powers remain within budget, but a baby/giant-step accumulation first exceeds it.

Do not automatically blame T3 merely because it is historically the first large generated-power difference.

Record:

- incoming semantic error
- outgoing semantic error
- amplification
- exact mathematical operation
- current Fast capacity

No repair or replacement sweep.

---

# R7 — target implication

Using the accepted DoubleAngle directional gain (~8.212), report:

- historical P93 polynomial error projected through plaintext DA;
- current production polynomial error projected through plaintext DA;
- whether the historical design is predicted to meet the DA2 pre-restore budget;
- whether current production is predicted to fail.

This is plaintext-only validation; no tuning.

---

# Decision classification

Choose exactly one:

## A — `P93_HISTORICAL_POLYNOMIAL_REFERENCE_MISMATCH`

Historical/design Standard polynomial comparator is not equivalent to Genuine Standard polynomial output.

## B — `P93_HISTORICAL_DESIGN_POLYNOMIAL_BUDGET_CONFLICT`

Reference reconciliation passes, but fresh historical P93 design no longer meets the derived polynomial budget.

## C — `P93_PRODUCTION_GENERATED_POWER_REGRESSION`

Historical P93 meets the budget, current production does not, and the first budget-breaking regression is a generated-power checkpoint.

Record the exact power.

## D — `P93_PRODUCTION_PS_ACCUMULATION_REGRESSION`

Generated powers remain within budget, but a PS baby/giant-step accumulation first breaks the budget.

## E — `P93_PRODUCTION_POLYNOMIAL_REGRESSION_UNLOCALIZED`

Historical P93 meets the budget and current production fails, but comparable checkpoints are insufficient to localize the first budget-breaking operation.

## F — `P93_POLYNOMIAL_REGRESSION_REPLAY_CONFLICT`

Current production or historical reference replay fails to reproduce accepted polynomial endpoints.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-POLYNOMIAL-REFERENCE-RECONCILIATION-AND-PS-REGRESSION-LOCALIZATION-summary.json`

Keep compact:
- <=300 pretty-printed JSON lines
- no full vectors
- no coefficient arrays
- no large repeated row-hash tables

Include:
- provenance
- Genuine Standard polynomial oracle
- fresh historical P93 replay
- historical-vs-Genuine reconciliation
- historical/current polynomial budget status
- regression vector closure
- compact PS checkpoint table
- first budget-breaking regression
- generated-power vs accumulation classification
- one classification A-F

---

## Validation

Run:
- Genuine Standard polynomial replay alignment test
- fresh historical P93 polynomial replay test
- reference reconciliation test
- regression vector-closure test
- focused PS localization test
- directly affected Primary tests
- `go test ./...`
- `git diff --check`

On success, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no q/planScale tuning
- no generated-power replacement sweep
- no PS repair
- no DoubleAngle repair
- no C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is proof that historical P93 polynomial accuracy is or is not sufficient against Genuine Standard, plus localization of the current production regression.
