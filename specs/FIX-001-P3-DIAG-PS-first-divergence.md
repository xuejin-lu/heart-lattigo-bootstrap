# FIX-001-P3-DIAG-PS — Localize first Paterson–Stockmeyer divergence

## Purpose

P3 productionized balanced two-sided pre-Rescale Fast Chebyshev power generation at Secondary commit:

`61607bb4bb82591009ce768d9a3773bed1497565`

Formal LogN13 validation establishes:

- Gate 1 PASS:
  - T2 max component ≈ `5.748576947794959e-9`
  - T3 max component ≈ `1.8383125176268944e-7`
- Gate 2 PASS:
  - all generated powers `T1,T2,T3,T4,T6,T8,T16` pass
- Gate 3 FAIL:
  - whole degree-30 polynomial max component ≈ `0.7909176408709179`
  - threshold `1e-2`

Therefore the first remaining supported failure is downstream of power generation, inside the Fast Paterson–Stockmeyer polynomial evaluation path.

This task must identify the **first specific PS operation** where a correct local semantic input becomes incorrect.

Diagnostic only. Do not repair production Lattigo.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when authored:

`026004033a3b85fd91cd8ee5a13b9555ee4665ad`

Secondary repository: `xuejin-lu/lattigo`

Exact P3 production commit:

`61607bb4bb82591009ce768d9a3773bed1497565`

Pinned Standard reference:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md` and Secondary `docs/FAST_CKKS_SPEC.md`.

Both worktrees must begin clean. Do not reset/stash/discard unrelated work.

Parent FIX-001/P3 remains incomplete.

---

# Formal workload

Use the exact existing LogN13 Mod1 real-branch formal input:

- LogN = 13
- all 4096 slots
- `configs/bootstrap_config.logN13.json`
- same deterministic application input
- same Q2 real branch after CoeffsToSlots
- same E1 scale reinterpretation
- same E2 scalar offset
- corrected Chebyshev oracle with `T0=1`
- exact Secondary `61607bb...`

Before PS interpretation require:

1. E2 reproduces historical formal E2;
2. actual generated powers all pass the corrected source-backed power oracle at `1e-2`:
   - T1
   - T2
   - T3
   - T4
   - T6
   - T8
   - T16
3. actual whole Fast degree-30 polynomial reproduces the known Gate-3 failure.

If these do not hold, stop with `PS_DIAGNOSTIC_PRECONDITION_MISMATCH`.

Do not run LogN16.

---

# Source path under diagnosis

Use the real P3 production code in:

`circuits/ckks/polynomial/fast.go`

Specifically:

```text
evaluatePlan
  -> evaluateBabyStep for each plan block
     -> optional constant Add
     -> repeated MulThenAdd(power[key], coefficient, accumulator)
  -> repeated giant-step evaluateMonomial
     -> optional Relinearize(b)
     -> Rescale(b)
     -> Mul(b, xpow, b)
     -> addAligned(a,b)
  -> optional final Relinearize
  -> final Rescale
```

The task must invoke the actual production planning and arithmetic path. Do not replace PS with Horner or another polynomial decomposition.

---

# Diagnostic access policy

Do not modify committed production source.

Temporary, uncommitted, build-tagged Secondary helpers are allowed only to expose/snapshot internal PS state from the real production workspace.

Preferred helper location:

`circuits/ckks/polynomial/*_diag_test.go` or equivalent build-tagged diagnostic file in package `polynomial`.

Requirements:

- run actual `commonPoly.PatersonStockmeyerPolynomial(...)` planner;
- run actual P3 `generatePowers(...)`;
- use actual Fast evaluator operations (`MulThenAdd`, `Rescale`, `Mul`, `addAligned`, etc.);
- deep-copy only for evidence/oracle comparison;
- do not reimplement the production PS algorithm as the tested path;
- any manual replay must cross-check its final rows/metadata against the corresponding real production intermediate;
- remove all temporary Secondary helpers before completion;
- Secondary must return clean at exact `61607bb...`.

Write temporary evidence to `/tmp` until Secondary is restored clean.

---

# Local semantic oracle principle

Do not compare every intermediate only against the ideal original application input.

At each operation, construct the expected output from the **decoded actual inputs immediately before that operation**.

This distinguishes a local arithmetic failure from small errors inherited from already-valid powers/accumulators.

For all semantic checkpoints use all 4096 slots and fixed threshold:

`max_component_abs <= 1e-2`.

Record max real, max imag, max complex, mean abs complex, RMSE, and worst indices.

---

# Part 1 — archive the actual PS plan

Record the exact `PatersonStockmeyerPolynomial` plan generated from the formal degree-30 polynomial:

- degree;
- base;
- number of initial baby-step blocks;
- each block's:
  - original plan index;
  - polynomial degree;
  - Level;
  - exact Scale + log2;
  - IsEven / IsOdd;
  - every non-nil coefficient index and exact coefficient value;
- reversal/order mapping used by `evaluatePlan` (`ws.babySteps[split-i-1]`);
- each giant-step round:
  - current block list/degrees;
  - pairing decisions (`giantSteps` values);
  - `deg = 1 << bits.Len64(even.Degree)`;
  - chosen `xpow` power index.

Historical evidence suggested base 8 / five initial blocks; record the actual synchronized plan rather than hardcoding that assumption.

---

# Part 2 — validate every baby block operation-by-operation

For each original `evaluateBabyStep` block, recreate the production execution on an independently-owned diagnostic accumulator while preserving exact production metadata and operation order.

Cross-check the diagnostic final block against the actual production baby-block snapshot.

## B0 — initialized zero accumulator

Record:

- Level = `poly.Level`
- Scale = `poly.Scale`
- Degree = 1
- NTT/Montgomery flags
- maintained q0/q1 row hashes

Decoded semantic value must be zero within numerical noise where decoding is valid.

## B1 — constant term when `poly.IsEven`

Execute the same production:

`eval.Add(out, poly.Coeffs[0], out)`

Local oracle:

`expected_after = expected_before + c0`

Record before/after metadata and metrics.

If this is first failure:

`FIRST_SUPPORTED_CAUSE = baby_constant_add`

and identify block index.

## B2 — each `MulThenAdd`

For each actual non-nil key visited in descending production order:

```text
eval.MulThenAdd(powers[key], poly.Coeffs[key], out)
```

Before each call record:

### Power term
- key n
- Level
- Scale exact + log2
- Degree
- decoded actual `T_n` vector
- power oracle pass status

### Accumulator
- Level
- Scale exact + log2
- Degree
- decoded actual accumulator

### Coefficient
- exact source-backed coefficient
- integer vs non-integer
- real/complex status

### MulThenAdd branch metadata
Using the actual Fast implementation, record:

- `power.Scale.Cmp(accumulator.Scale)` direction;
- whether accumulator promotion occurs;
- exact promotion ratio before `.BigInt()` if applicable;
- integer ratio actually used;
- ratio integrality / truncation error;
- whether non-integer coefficient invokes `coefficientScale(level)`;
- exact coefficientScale;
- predicted term scale and output scale.

Then execute actual `MulThenAdd`.

Local semantic oracle:

`expected_after = accumulator_actual_before + coefficient * power_actual_before`

where complex multiplication uses the source coefficient exactly/high precision before conversion to complex128 for comparison.

If first failure occurs here classify:

`FIRST_SUPPORTED_CAUSE = baby_mul_then_add`

and record:

- block index;
- coefficient key;
- exact internal branch;
- last passing accumulator;
- first failing accumulator.

Do not continue causal attribution past the first failed operation, though already-produced later evidence may be archived.

---

# Part 3 — baby-block final oracle

For every baby block, independently compute its plaintext block polynomial on the decoded actual formal E2/powers using the plan coefficients.

Compare to final actual baby block.

Require all blocks <= `1e-2` before moving causally to giant steps.

If an operation-level trace appeared to pass but block-level oracle fails, classify:

`FIRST_SUPPORTED_CAUSE = baby_step_unresolved`

and report discrepancy.

---

# Part 4 — giant-step operation trace

Only if **all baby blocks pass**.

Follow the exact `evaluatePlan` pairing rounds and `evaluateMonomial` operations.

For each paired `(even=a, odd=b, xpow)` record actual inputs before mutation.

Define decoded actual vectors:

- `A` = actual `a` before the giant step;
- `B` = actual `b` before the giant step;
- `X` = actual generated `xpow` power.

The local semantic target after the whole giant step is:

`B_final_expected = A + B * X`

subject to the planner's block semantics.

Trace separately:

## G1 — optional Relinearize(b)

If b degree 2, execute actual relinearization.

Under current authoritative zero-secret Fast semantics, compare decoded before/after logical value.

If first failure:

`FIRST_SUPPORTED_CAUSE = giant_relinearize`

## G2 — Rescale(b)

Execute actual Fast Rescale.

Local oracle: represented decoded value must remain `B` within `1e-2`.

Record:

- pre/post Level;
- pre/post Scale;
- divisor;
- q0/q1 capacity evidence for pre-Rescale c0:
  - max centered coefficient;
  - `Q01/2` ratio;
  - outside count.

If failure and capacity crossing is proven:

`FIRST_SUPPORTED_CAUSE = giant_rescale_q01_capacity`

Otherwise:

`FIRST_SUPPORTED_CAUSE = giant_rescale`

## G3 — Mul(b, xpow, b)

Execute actual Fast `Mul`.

Local oracle:

`expected = B_after_rescale_actual * X_actual`.

Because multiplication may produce a high-scale degree-2 value whose immediate q0/q1 decoded interpretation can be unsafe, distinguish:

- semantic slot comparison when representation is valid;
- coefficient capacity accounting before any later centered-CRT Rescale.

Record output Level/Scale/Degree and whether future Rescale would face Q01 ambiguity.

If direct supported failure:

`FIRST_SUPPORTED_CAUSE = giant_multiply`

## G4 — `addAligned(a,b)`

Before call record exact Scale relation.

Trace the actual `addAligned` branch:

- equal-scale direct Add; or
- ratio computation;
- exact ratio;
- `.BigInt()` ratio;
- integer multiplication scratch;
- metadata Scale assignment;
- final Add and copy-back.

Local oracle after alignment/add:

`expected = A_actual + B_times_X_actual`.

If first failure classify the narrowest supported cause:

- `giant_addaligned_scale_promotion`
- `giant_addaligned_metadata_scale_assignment`
- `giant_addaligned_addition`
- `giant_addaligned_unresolved`

---

# Part 5 — finalization trace

Only if all baby and giant steps pass.

When one block remains:

## F1 — optional final Relinearize

Compare decoded before/after.

Failure:

`FIRST_SUPPORTED_CAUSE = ps_final_relinearize`

## F2 — final Rescale

Before Rescale record coefficient-domain q0/q1 capacity statistics exactly as in prior capacity diagnostics.

Execute actual final Rescale and compare represented value before/after.

If capacity crossing causally explains failure:

`FIRST_SUPPORTED_CAUSE = ps_final_rescale_q01_capacity`

Else if Rescale itself fails semantic oracle:

`FIRST_SUPPORTED_CAUSE = ps_final_rescale`

## F3 — final whole-polynomial cross-check

Require the traced final result to match ordinary public `FastEvaluator.Evaluate(...)` maintained rows/metadata (bit-exact where possible) and reproduce the known whole-polynomial error.

If trace and public output disagree:

`FIRST_SUPPORTED_CAUSE = ps_diagnostic_inconsistency`

---

# Capacity checks for scalar/coefficient operations

If first failure occurs inside a baby-step scalar operation, perform coefficient-domain evidence sufficient to determine whether scalar scaling/promotion itself violates q0/q1 centered uniqueness.

At minimum for the immediately failing operation record:

- Q01/2;
- accumulator c0 max centered coefficient before operation;
- power c0 max centered coefficient;
- scalar/integer multiplier(s) introduced by `MulThenAdd`;
- nominal post-multiply coefficient magnitude where exact reconstruction is possible;
- whether the subsequent operation requires centered CRT interpretation/Rescale.

Do not blame capacity from scale metadata alone.

---

# Standard/source-backed context

The causal diagnosis is primarily local Fast-vs-mathematical-oracle.

Where practical also record the pinned Standard planner metadata for the same block/operation, but do not require internal Standard ciphertext state to match Fast because representations differ.

Use Standard only as supporting evidence, not as a substitute for local semantic checks.

---

# Required classification

Choose the **earliest evidence-supported** one from:

- `baby_constant_add`
- `baby_mul_then_add`
- `baby_step_unresolved`
- `giant_relinearize`
- `giant_rescale_q01_capacity`
- `giant_rescale`
- `giant_multiply`
- `giant_addaligned_scale_promotion`
- `giant_addaligned_metadata_scale_assignment`
- `giant_addaligned_addition`
- `giant_addaligned_unresolved`
- `ps_final_relinearize`
- `ps_final_rescale_q01_capacity`
- `ps_final_rescale`
- `ps_diagnostic_inconsistency`

Also record exact block/round/key identifiers so the next repair can be narrow.

Do not implement the repair in this task.

---

# Required artifacts

After all temporary Secondary helpers are removed and Secondary is restored clean at `61607bb...`, create:

- `results/FIX-001-P3-DIAG-PS-logN13-fast.json`
- `results/FIX-001-P3-DIAG-PS-logN13-summary.json`

Raw artifact should include operation-by-operation evidence up to and beyond the first failing checkpoint as reasonably useful.

Summary must include:

- exact Secondary commit;
- E2 / all-power preconditions;
- actual PS plan;
- baby block pass/fail table;
- first failing baby coefficient operation if applicable;
- giant-step round table if reached;
- finalization checkpoints if reached;
- whole polynomial reproduced error;
- first supported cause;
- last passing / first failing operation;
- relevant scale-ratio/capacity evidence;
- tests and clean-state evidence.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- Secondary clean `go test ./...` passes after temporary helpers removed;
- Secondary remains exact `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Lattigo modification;
- Primary artifacts committed/pushed normally;
- both worktrees clean at end;
- no Gate-4/5 repair attempt;
- no LogN16;
- no benchmark;
- no EXP-003.

---

# Non-goals

Do not:

- modify power generation;
- revert balanced scheduling;
- repair MulThenAdd/addAligned/PS in this task;
- change polynomial degree or Mod1 parameters;
- lower scales;
- add q2+ or auxiliary limbs;
- loosen threshold;
- run LogN16;
- benchmark;
- start EXP-003.

The deliverable is the first precise failing Paterson–Stockmeyer operation after all formal generated powers have been proven correct.