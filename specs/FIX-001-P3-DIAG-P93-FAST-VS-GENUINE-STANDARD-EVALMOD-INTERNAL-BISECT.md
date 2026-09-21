# FIX-001-P3-DIAG-P93-FAST-VS-GENUINE-STANDARD-EVALMOD-INTERNAL-BISECT

## Purpose

The preceding reference-reconciliation task at Primary commit `aef1cb2acb6c4fa511479e314b044121377e8d83` established the authoritative Standard hierarchy:

- Genuine Standard public Bootstrap exact E2E:
  `5.830057349387463e-8`;
- Genuine Standard staged pipeline equals public Bootstrap exactly for the deterministic workload;
- matched diagnostic Standard C2S objects equal Genuine Standard C2S;
- matched diagnostic Standard EvalMod differs from Genuine Standard EvalMod only in public Scale metadata:
  - coefficient rows equal;
  - matched/internal Scale = `2^50`;
  - Genuine Standard public EvalMod Scale = `2^45`;
  - ratio = 32;
- therefore the matched Standard object is valid as an **internal/stage oracle before the public Scale reset**, but not as a public exact-E2E oracle.

This also revises an earlier interpretation:

> The public `2^50 -> 2^45` EvalMod Scale reset is **not a Fast-specific production bug**. Genuine Standard uses the same public Scale contract.

Likewise, the final public DefaultScale restoration is part of the Standard public contract and must not be "repaired" merely to preserve the internal `2^50` diagnostic semantics.

The true current Fast-vs-Genuine-Standard boundary is:

- C2S real/imag: ~`9e-14` semantic difference;
- Fast corrected/internal EvalMod vs Genuine Standard internal/public-equivalent:
  internal error about `0.00418/0.00426`;
- after the common public Scale reset by factor 32, public Fast-vs-Genuine-Standard EvalMod error is about `0.1339/0.1364`;
- final Fast exact E2E remains ~`0.188575`.

The current diagnostic milestone is `1e-2` at the public/E2E level. Since the public EvalMod Scale contract changes `2^50 -> 2^45`, an internal pre-public error budget corresponding to `1e-2` is approximately:

[
1e-2 / 32 = 3.125e-4.
]

The current internal Fast EvalMod error ~`4.2e-3` therefore remains materially too large.

This task must locate the first material **internal arithmetic/semantic divergence** between current Fast EvalMod and Genuine Standard EvalMod, before the shared public Scale reset.

No production repair in this task.

---

## Repository safety

### Primary

Repository:
`xuejin-lu/heart-lattigo-bootstrap`

Follow `AGENTS.md` mandatory startup preflight, synchronize `main`, then read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary

Repository:
`xuejin-lu/lattigo`

Required local state:

- branch: `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- exact expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed profile

Keep fixed:

- LogN13
- same deterministic 4096-slot workload
- same bootstrap parameters
- q0 = 56-bit effective Fast profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012 Fast candidate
- Fast planScale = `2^93`
- same polynomial degree/K/DoubleAngle count/message ratio
- Genuine Standard implementation/configuration unchanged
- public diagnostic milestone = `1e-2`
- derived internal pre-public target = `3.125e-4`

No tuning.

---

# R0 — establish equivalent C2S inputs

From the same deterministic original message:

1. run the Genuine Standard staged path through CoeffsToSlots;
2. run the current Fast path through CoeffsToSlots.

For real and imag record:

- decoded semantic difference;
- Level/Scale/Degree;
- representation flags;
- source lineage.

Expected semantic difference:
approximately `1e-13`.

Require max-component C2S semantic difference <= `1e-10`.

If not, stop:

`P93_FAST_VS_STANDARD_C2S_PRECONDITION_CONFLICT`.

Do not require coefficient-row equality across Standard and Fast representations.

---

# R1 — source-faithful internal EvalMod traces

Trace the **actual current implementations**, not a proxy-only reconstruction.

## Standard trace

Read and follow the Genuine Standard public EvalMod implementation source.

Capture the logical stages immediately before the public wrapper Scale reset.

## Fast trace

Read and follow the current dirty Fast public EvalMod implementation source.

Capture equivalent logical stages immediately before its public wrapper Scale reset.

Primary diagnostic code may source-faithfully replay operations only when necessary to expose checkpoints. It must prove replay alignment with the actual public implementation at the pre-public boundary.

For both real and imag capture, when semantically present:

1. EvalMod input
2. metadata normalization / scaling-factor assignment
3. offset application
4. polynomial evaluator input
5. generated-power checkpoints sufficient to align Standard and Fast
6. PS baby-step / giant-step boundaries sufficient for first-divergence localization
7. polynomial root before final polynomial Rescale
8. polynomial output after Rescale
9. any internal Q012 -> Q01 contraction on the Fast side
10. pre-DoubleAngle state
11. DoubleAngle round 0:
    - square
    - multiplier/constant
    - post-Rescale
12. DoubleAngle round 1 post-Rescale
13. DoubleAngle round 2 post-Rescale
14. internal final restore, immediately **before** public Scale reset
15. public output after shared DefaultScale reset (context only)

Do not call matched/manual Standard objects the Genuine Standard path unless replay alignment to the actual Standard public implementation is proven.

---

# R2 — semantic comparison rules

Standard and Fast may use different internal representation flags. Compare decoded semantics, not raw row equality, unless row equality is mathematically meaningful within one implementation.

At every aligned checkpoint report:

- Standard decoded max magnitude;
- Fast decoded max magnitude;
- Fast-vs-Standard max-component difference;
- max-complex difference;
- mean-complex difference;
- worst slot/component;
- Level/Scale/Degree for each side;
- capacity status on Fast q0/q1/q2 where relevant;
- whether the comparison is semantically aligned.

If a checkpoint cannot be aligned, mark it `not_comparable`; do not fabricate a metric.

---

# R3 — first observable and first material internal divergence

Determine:

## First observable divergence

Earliest aligned internal checkpoint above deterministic/replay numerical floor.

## First material divergence

Use both criteria and report whichever becomes active first:

1. >= 10% of the current final internal EvalMod gap;
2. >= `3.125e-4`, the internal error budget corresponding to public `1e-2` after the shared factor-32 Scale contract.

Current internal final context:

- real ≈ `0.004183019582623534`
- imag ≈ `0.004261561142162278`.

For the transition around first material divergence report:

- incoming Fast-vs-Standard difference;
- outgoing difference;
- amplification factor;
- exact operation;
- Level/Scale/Degree before/after;
- Fast q012 capacity;
- whether metadata or coefficients changed.

Do not automatically blame T2 or any generated power unless the material criterion is met.

---

# R4 — prove the factor-32 public mapping

From the aligned internal final state immediately before public Scale reset:

- compute Fast-vs-Standard internal semantic difference;
- then apply each implementation's actual public DefaultScale contract;
- compute public Fast-vs-Standard difference.

Verify numerically that:

[
E_{public} approx 32 E_{internal}
]

for real and imag, within justified floating-point tolerance.

Expected context:

- internal real ≈ `0.00418301958`
- public real ≈ `0.1338566`
- internal imag ≈ `0.00426156114`
- public imag ≈ `0.1363700`.

This confirms that the shared public Scale reset amplifies an existing internal Fast error rather than creating a Fast-specific bug.

---

# R5 — reconcile with historical P93 design evidence

The historical/fresh P93 design diagnostic reported internal-like Fast-vs-Standard EvalMod errors:

- real ≈ `0.0001635104343335875`
- imag ≈ `0.00013359983063118935`.

Do not assume those values use the same exact internal Standard oracle.

Audit their reference lineage and answer:

1. Are those historical values comparable to the Genuine Standard internal pre-public oracle established in this task?
2. If yes, where does the current production path diverge from the historical P93 design path enough to explain `~0.0042` vs `~0.00016`?
3. If no, explicitly retire those values as a different proxy metric for this purpose.

Only perform checkpoint comparison to historical P93 design if semantic equivalence is proven.

Do not rerun P92/P94.

---

# R6 — source audit

Record the exact Standard and Fast source functions/files for:

- public EvalMod wrapper;
- internal Mod1 evaluator;
- polynomial evaluator call;
- target scale derivation;
- plan scale / decomposition strategy;
- Q012/Q01 behavior;
- DoubleAngle arithmetic;
- internal final scale restore;
- public DefaultScale reset.

Produce a compact Standard-vs-Fast configuration table.

Pay special attention to any source-level differences in:

- targetScale;
- polynomial input scale;
- coefficient encoding/scaling;
- planScale;
- Rescale placement;
- DoubleAngle multiplier exponents;
- constant scaling;
- level schedule.

Do not modify source.

---

# Decision classification

Choose exactly one:

## A — `P93_FAST_VS_STANDARD_C2S_PRECONDITION_CONFLICT`

C2S inputs are not semantically aligned enough for EvalMod comparison.

## B — `P93_FAST_EVALMOD_PREPROCESSING_DIVERGENCE`

First material internal divergence occurs before polynomial evaluation.

## C — `P93_FAST_EVALMOD_POLYNOMIAL_DIVERGENCE`

First material divergence occurs inside polynomial/PS evaluation.

Record earliest material checkpoint.

## D — `P93_FAST_EVALMOD_DOUBLEANGLE_DIVERGENCE`

Polynomial outputs align within budget; first material divergence occurs in DoubleAngle.

## E — `P93_FAST_EVALMOD_INTERNAL_RESTORE_DIVERGENCE`

Arithmetic aligns through DoubleAngle; first material divergence occurs at internal restore before public reset.

## F — `P93_FAST_EVALMOD_PUBLIC_RESET_ONLY`

Internal Fast-vs-Standard difference is already <= `3.125e-4`, and only the common public reset makes it exceed `1e-2`.

This classification would imply the internal arithmetic already meets the milestone budget. Do not select it unless the measured internal difference actually satisfies the budget.

## G — `P93_HISTORICAL_P93_REFERENCE_NOT_COMPARABLE`

Current Fast-vs-Genuine-Standard localization succeeds, but the historical P93 `~1e-4` evidence uses a non-equivalent oracle and cannot be used as the target.

Include the actual current first material divergence as a secondary field.

## H — `P93_FAST_VS_STANDARD_EVALMOD_TRACE_ALIGNMENT_INVALID`

Standard and Fast internal traces cannot be semantically aligned reliably enough.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-FAST-VS-GENUINE-STANDARD-EVALMOD-INTERNAL-BISECT-summary.json`

Keep it compact. Include:

- Primary provenance;
- Secondary branch/HEAD/dirty fingerprint;
- R0 C2S precondition;
- compact internal checkpoint table;
- first observable divergence;
- first material divergence;
- factor-32 mapping proof;
- historical-P93 comparability result;
- Standard-vs-Fast source configuration table;
- one classification A-H;
- explicit confirmation Secondary was not modified.

No full slot vectors, coefficient dumps, or large repeated row-hash tables.

---

## Validation

Run:

- focused internal EvalMod bisect test;
- Standard source-faithful replay/public alignment test;
- Fast source-faithful replay/public alignment test;
- directly affected Primary tests;
- `go test ./...`;
- `git diff --check`.

On success, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification;
- no Secondary commit/push;
- no destructive operation on dirty Secondary;
- no metadata "repair";
- no coefficient correction;
- no q/planScale tuning;
- no P92/P94 sweep;
- no production PS repair;
- no C2S/S2C/finalizer repair;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.

The deliverable is the first trustworthy internal Fast-vs-Genuine-Standard EvalMod divergence under the real public semantics, not a fix.
