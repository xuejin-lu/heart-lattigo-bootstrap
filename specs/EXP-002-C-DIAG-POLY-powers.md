# EXP-002-C-DIAG-POLY — Split Fast polynomial failure into power generation vs Paterson–Stockmeyer accumulation

## Purpose

The formal LogN13 correctness diagnosis has now localized the first failing operation to the Fast Mod1 polynomial evaluation:

- E1 scale reinterpretation: PASS
- E2 cosine/Chebyshev offset: PASS
- E3 polynomial evaluation: FAIL
- E3 max component error in the previous formal run: approximately `0.790917606881237`
- manual EvalMod replay matches the public EvalMod path

This task must answer one narrower question:

> Does the formal degree-30 Chebyshev polynomial first become wrong while generating the Chebyshev power basis, or only later during Paterson–Stockmeyer accumulation / scale alignment?

This is a diagnostic task only. Do not repair the polynomial evaluator yet.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when this spec is authored:

`7eae25265627ac5ab428b247b5ecdf4036a9ac2a`

Fast Lattigo baseline:

`ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Standard backend is not required for the causal power-vs-PS split because the previous Q01 task already established that the Fast input to EvalMod is valid. If Standard is used for a supporting cross-check, keep it pinned at:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow Primary and Secondary `AGENTS.md`. Read `docs/FAST_CKKS_SPEC.md` before temporary Secondary diagnostic work.

---

## Formal workload must remain unchanged

Reproduce the exact formal LogN13 path through E2:

- LogN = 13
- full 4096 slots
- same `configs/bootstrap_config.logN13.json`
- LogDefaultScale = 45
- LogMessageRatio = 10
- Mod1LogScale = 60
- Mod1Degree = 30
- DoubleAngle = 3
- K = 16
- Mod1InvDegree = 0
- same deterministic application input
- same `ModUpThenEncode` circuit order

Use the **actual Q2 real branch**, then perform the same E1 scale reinterpretation and E2 scalar offset as the pinned Fast Mod1 implementation.

Do not replace E2 with an arbitrary synthetic polynomial input.

Before continuing, verify the newly reproduced E2 decoded vector agrees with the previous EXP-002-C-DIAG-EVALMOD E2 evidence within the fixed `1e-2` component threshold. If not, stop with `DIAGNOSTIC_INPUT_MISMATCH`.

---

## Source-backed implementation split

At the pinned Fast baseline, inspect and use the real implementation in:

`circuits/ckks/polynomial/fast.go`

The relevant split is:

1. `workspace.reset(...)`
2. `workspace.generatePowers(...)`
   - recursive Chebyshev power generation
   - `Mul` / `MulRelin`
   - possible `Relinearize`
   - `Rescale`
   - Chebyshev identity `2*T_a*T_b - T_|a-b|`
   - `subAligned(...)` when needed
3. Paterson–Stockmeyer plan construction
4. `workspace.evaluatePlan(...)`
   - baby-step coefficient accumulation (`MulThenAdd`)
   - giant-step multiply/merge
   - `addAligned(...)`
   - final relinearization/rescale

Do not infer this structure from memory; confirm it from the synchronized source before running.

---

## Temporary diagnostic access policy

The production Fast implementation must remain unchanged in Git history.

For this task only, it is explicitly permitted to create temporary, uncommitted, diagnostic-only Secondary files to expose snapshots of the real polynomial workspace. Prefer a build-tagged file such as:

`circuits/ckks/polynomial/zz_exp002c_diag_powers.go`

with a dedicated build tag such as `exp002cdiag`.

The temporary helper may expose a narrowly bounded diagnostic API that:

- invokes the existing `workspace.reset` and `workspace.generatePowers` implementation;
- returns deep copies of every generated `workspace.powers[n]` after power generation is complete;
- returns the generated power indices and metadata needed for diagnosis;
- optionally returns the actual Paterson–Stockmeyer plan metadata, but must not replace the implementation with a reimplementation.

A temporary diagnostic test/harness may also be added under `circuits/ckks/bootstrapping` to reconstruct the exact formal E2 input and call this diagnostic API.

Rules:

- do not commit or push temporary Secondary diagnostic files;
- do not change existing production `.go` files;
- do not use reset/stash/discard to clean unrelated work;
- Secondary must be clean before starting;
- after evidence is written to `/tmp`, delete only files created by this task;
- verify Secondary returns exactly to clean `fast-ckks` HEAD `ce79b861...`;
- run clean `go test ./...` after removing diagnostic files.

If the required snapshots cannot be obtained without modifying an existing production source file, stop and report before doing so.

---

## Power projection / decode

Fast generated powers may have degree 1 or 2 and logical levels greater than 1.

For each power snapshot, build an isolated q0/q1 diagnostic ciphertext:

- diagnostic level = 1;
- preserve the source ciphertext degree (including degree 2 when present);
- copy q0 and q1 for **every existing ciphertext component**;
- copy Scale, LogSlots, NTT and relevant plaintext metadata;
- apply `IMForm` to q0/q1 on the diagnostic copy only when the source is Montgomery;
- mark the diagnostic copy non-Montgomery after normalization;
- decrypt using the authoritative all-zero bootstrap secret;
- ordinary CKKS decode all 4096 slots;
- never read q2+;
- prove the original power snapshot is unchanged by projection.

Do not silently drop c2 merely because the current secret is zero; preserve the actual ciphertext degree in the diagnostic copy.

---

## Plaintext Chebyshev power oracle

Let `x_i` be the actual decoded E2 slot value for slot `i`.

The generated Fast power `powers[n]` is intended to represent the Chebyshev basis value `T_n(x)`.

Compute the expected value using the Chebyshev identity used by the implementation:

- `T_0(x) = 1`
- `T_1(x) = x`
- for the implementation split `n = a + b`:
  - `T_n(x) = 2*T_a(x)*T_b(x) - T_|a-b|(x)`

Use the same `commonpolynomial.SplitDegree(n)` dependency structure when recording the diagnostic graph.

For additional oracle validation, where practical compare this recurrence against Lattigo's source-backed `bignum.Polynomial` Chebyshev evaluation for representative slots/powers. If these disagree beyond numerical floating-point noise, stop with `POWER_PLAINTEXT_ORACLE_INVALID`.

The expected power must be computed from the **actual E2 decoded values**, not the ideal original application input.

---

## Required power evidence

After the real `generatePowers(...)` completes, enumerate every power present in the actual workspace map, including power 1.

For every generated `n`, record:

- `n`
- dependency split `(a,b)` from `SplitDegree(n)` when n > 1
- `c = |a-b|`
- source Degree
- source Level
- source Scale exact string + log2
- IsNTT
- IsMontgomery
- q0/q1 row hashes for every component
- all 4096 decoded q0/q1-projection slots
- plaintext expected `T_n(E2_actual)` for all 4096 slots
- comparison metrics

Metrics:

- max_abs_complex
- mean_abs_complex
- rmse_complex
- max_abs_real
- max_abs_imag
- max_component_abs
- max-error indices
- sample count

Fixed pass criterion:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

Do not loosen the threshold.

---

## Dependency-aware failure classification

Do not simply sort powers by numeric n and call the first failing n the cause.

A power `T_n` is a supported first failing power when:

- its own q0/q1 decoded result fails the fixed threshold; and
- every power dependency used to construct it that exists in the generated dependency graph is still within threshold.

Record all failing powers, but identify the earliest dependency-supported failing node(s).

Allowed classification:

### Case A — at least one dependency-supported power fails

`FIRST_SUPPORTED_CAUSE = power_generation`

Also record:

- `first_failing_power_n`
- `(a,b,c)` dependency
- whether the resulting snapshot is degree 1 or degree 2
- its Level/Scale before PS evaluation

Do **not** repair it and do not yet attribute the failure to Mul/Relin/Rescale/subAligned unless this task directly proves that sub-operation.

### Case B — all generated powers pass, final polynomial still fails

Run the ordinary real Fast `PolynomialEvaluator.Evaluate(E2, Mod1Poly, targetScale)` on an independent E2 copy and confirm it reproduces the previous E3 failure (> `1e-2`).

If every generated power passes but final polynomial evaluation fails:

`FIRST_SUPPORTED_CAUSE = paterson_stockmeyer_accumulation_or_scale_alignment`

Record the actual PS plan metadata:

- number of baby-step polynomial blocks;
- each block's degree, planned Level, planned Scale;
- giant-step merge structure if available through the temporary diagnostic helper;
- final planned vs actual Level/Scale.

Do not inspect/fix individual baby/giant steps in this task; that becomes the next bounded diagnostic.

### Case C — generated powers pass and final polynomial now passes

`FIRST_SUPPORTED_CAUSE = diagnostic_inconsistency`

Stop. Reconcile with EXP-002-C-DIAG-EVALMOD before proceeding.

### Case D — power oracle cannot be validated

`FIRST_SUPPORTED_CAUSE = unresolved_power_oracle`

State exactly which power/oracle check prevents interpretation.

---

## Required whole-polynomial cross-check

Regardless of Case A/B, run the actual Fast polynomial evaluator on an independent E2 input and record:

- output Level/Scale;
- max component error vs exact Mod1 polynomial plaintext evaluation;
- whether it reproduces the prior E3 formal failure.

Use the same source-backed plaintext polynomial oracle already validated in EXP-002-C-DIAG-EVALMOD.

If the whole-polynomial failure is not reproduced, do not make a causal claim from the power snapshots.

---

## Required artifacts

Write diagnostic output to `/tmp` while Secondary diagnostic files exist.

Only after Secondary has been restored clean, copy result artifacts into Primary:

- `results/EXP-002C-DIAG-POLY-logN13-fast.json`
- `results/EXP-002C-DIAG-POLY-logN13-summary.json`

Summary must contain:

- E2 reproduction check;
- actual generated power set;
- per-power dependency graph and metrics;
- all failing powers;
- dependency-supported first failing power(s), if any;
- whole-polynomial reproduction metrics;
- power-generation vs PS classification;
- PS plan metadata when Case B applies;
- confirmation that temporary Secondary files were removed;
- final Secondary HEAD/worktree state.

---

## Validation

Before completion:

- temporary diagnostic run succeeds with the dedicated build tag;
- all temporary Secondary files created by this task are removed;
- Secondary HEAD remains `ce79b861...`;
- Secondary worktree is clean;
- clean Fast `go test ./...` passes after removal;
- Primary `go test ./...` passes;
- Primary result artifacts are committed and pushed by normal fast-forward;
- no production Lattigo source commit exists;
- LogN16 is not run;
- EXP-003 is not started.

---

## Non-goals

Do not:

- fix the polynomial evaluator;
- change Fast arithmetic;
- change Mod1 parameters or polynomial coefficients;
- change Paterson–Stockmeyer planning;
- sweep polynomial degree, K, LogMessageRatio, or scale;
- benchmark performance;
- run LogN16;
- start EXP-003.

The only deliverable is a supported split between **power generation** and **Paterson–Stockmeyer accumulation/scale alignment**, with the first failing generated Chebyshev power identified when applicable.