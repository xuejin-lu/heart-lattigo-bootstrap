# FIX-001-DIAG-T3 — Localize the repaired T3 c>0 correction failure

## Purpose

After correcting the Primary Chebyshev oracle, the formal repaired-Fast power replay now establishes:

- formal E2 reproduction: exact;
- repaired `T2`: PASS, max component ≈ `0.0004882811772376483`;
- repaired `T3`: FAIL, max component ≈ `0.031234736649481946`;
- T3 dependencies `(T1,T2,T1)` are individually valid;
- first supported cause remains inside repaired Fast power generation, now specifically the first `c>0` Chebyshev correction;
- whole polynomial and public Q3(real) remain wrong.

This task must answer exactly:

> For formal `T3 = 2*T1*T2 - T1`, is the first failure introduced before the correction term, or by the post-Rescale `subAligned(..., T1)` correction path? If `subAligned` is responsible, which internal scale-alignment step first changes the logical value incorrectly?

Diagnostic only. Do not repair production Lattigo yet.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when this spec is authored:

`263bea3bdd92672a06c5c089501d2a4a017c0f9d`

Repaired Fast backend:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Standard reference, if needed only for supporting same-stage comparison:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md`. Secondary must begin clean. Do not reset/stash/discard unrelated work.

The parent FIX-001 remains incomplete.

---

## Formal workload

Reproduce the exact corrected formal path:

- LogN = 13
- all 4096 slots
- `configs/bootstrap_config.logN13.json`
- same deterministic application input
- same Q2 real branch
- same E1 scale reinterpretation
- same E2 offset
- repaired Fast exactly `87be78ff...`
- corrected Chebyshev plaintext oracle with `T0=1`

Require E2 to match the corrected historical E2 vector exactly or within the existing fixed `1e-2` threshold before interpretation.

Use the actual repaired production power-generation path to obtain formal `T1` and `T2`. Require both to pass the corrected power oracle before proceeding.

Do not run LogN16.

---

## Confirm actual T3 production mode

Before manually replaying T3, inspect the synchronized repaired source in:

`circuits/ckks/polynomial/fast.go`

Record whether formal T3 is generated with:

- lazy `Mul`, retaining degree 2; or
- strict `MulRelin`, producing degree 1.

Do not assume `MulRelin` from the mathematical recurrence. Reproduce the exact production path, including any required prior relinearization of operands.

Record the actual T3 dependency split from `SplitDegree(3)`:

- `a`
- `b`
- `c = |a-b|`

Expected mathematical recurrence is:

`T3 = 2*T_a*T_b - T_c = 2*T1*T2 - T1`.

---

## Diagnostic access policy

Do not commit production Lattigo changes.

Temporary, uncommitted, build-tagged Secondary diagnostic helpers are allowed only if needed to access the real Fast polynomial workspace or unexported `subAligned` path.

Preferred approach:

- use the real repaired `fastPowerBasis` / workspace implementation;
- snapshot actual T1/T2 and the production T3 result;
- use a temporary helper in package `polynomial` to expose a diagnostic T3 replay and `subAligned` trace;
- do not modify existing production `.go` files.

Any temporary helper that reproduces `subAligned` internals must cross-check its final result against the actual production `ws.subAligned(...)` result:

- maintained q0/q1 rows bit-exact when possible;
- Level/Scale/Degree equal;
- decoded values equal within numerical noise.

After evidence is written to `/tmp`, remove all temporary Secondary files and restore exact clean `87be78ff...`.

---

## Local-oracle principle

Each step must be judged from the **actual decoded logical value before that step**, not from the ideal application input.

Let:

- `X = decoded actual T1`
- `Y = decoded actual repaired T2`

Both must already pass their own corrected Chebyshev oracles.

Then the expected local T3 values are:

- product: `X * Y`
- doubled product: `2 * X * Y`
- post-Rescale: still `2 * X * Y`
- corrected T3: `2 * X * Y - X`

This isolates T3-local arithmetic from the small existing T1/T2 approximation error.

Fixed pass threshold remains:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

Do not loosen it.

---

## Required checkpoints

### T3-0 — inputs

Record actual production snapshots for:

- T1
- repaired T2

For each:

- Degree
- Level
- Scale exact string + log2
- IsNTT / IsMontgomery
- q0/q1 hashes for every component
- all 4096 decoded values
- corrected oracle metrics

Require both to pass.

### T3-1 — multiply

Execute the exact production T3 multiplication mode using T1/T2.

Record metadata and maintained hashes.

Do not make a semantic claim from an unsafe high-scale pre-Rescale q0/q1 projection unless capacity is independently valid.

### T3-2 — double

Apply the exact production doubling `Add(out,out,out)`.

Record metadata/hashes.

Again, do not infer logical failure solely from pre-Rescale decoding.

### T3-3 — Rescale before correction

Apply the exact repaired production Rescale.

Now decode the result through the validated q0/q1 diagnostic projection.

Compare against:

`2 * T1_actual * T2_actual`

for all 4096 slots.

If this exceeds `1e-2`, classify:

`FIRST_SUPPORTED_CAUSE = t3_multiply_double_or_rescale`

and do not blame `subAligned`.

Still record subsequent already-produced evidence if available, but stop causal interpretation there.

### T3-4 — actual production `subAligned(postRescale, T1)`

If T3-3 passes, invoke the real production `subAligned` path used by repaired power generation.

Decode result and compare against:

`T3_expected = T3_3_actual - T1_actual`

If this fails, continue into the internal alignment trace below.

---

## Mandatory `subAligned` metadata trace

Before performing the correction, record:

### Left/output operand

- Level
- Scale exact string
- scale log2
- Degree
- q0/q1 hashes

### T1 correction operand

- Level
- Scale exact string
- scale log2
- Degree
- q0/q1 hashes

Then record:

- `out.Scale.Cmp(T1.Scale)` direction;
- exact high-precision ratio selected by production logic;
- decimal ratio with sufficient precision;
- `.BigInt()` result used by production `subAligned`;
- absolute difference `ratio_exact - ratio_integer`;
- relative ratio error;
- whether ratio is exactly integral.

Do not use float64 to decide exact integrality.

---

## `subAligned` internal step replay

Using independent copies, replay the exact branch selected by production `subAligned` and snapshot each step.

The current repaired implementation may perform some or all of:

1. allocate/resize `scaleScratch` at `min(out.Level, T1.Level)`;
2. copy metadata/domain flags;
3. compute integer alignment ratio with `.BigInt()`;
4. `MulIntegerMaintained(smallerScaleOperand, ratio, scaleScratch)`;
5. assign `scaleScratch.Scale = largerScale`;
6. `Sub(...)`;
7. possible `copyMaintainedElement(...)` back into `out`.

Do not assume branch direction; follow actual scale comparison.

For every intermediate scratch/result record:

- Level
- Scale
- Degree
- q0/q1 hashes
- decoded vector when representation is safe/valid

---

## Correction-operand semantic check

The most important internal check is whether the aligned correction operand still represents the same logical T1 value.

After the production-equivalent integer multiplication + metadata scale assignment, decode the aligned T1 scratch and compare it against original `T1_actual`.

Record full metrics.

Allowed interpretation:

### If aligned scratch already differs from T1

`FIRST_SUPPORTED_CAUSE = subaligned_scale_promotion`

Also identify whether evidence points specifically to:

- non-integral ratio rounded/truncated by `.BigInt()`;
- integer multiplication itself;
- metadata scale assignment after integer multiplication;
- level reduction/copy interaction.

Do not collapse these if the trace can distinguish them.

### If aligned scratch matches T1 but subtraction output fails

`FIRST_SUPPORTED_CAUSE = subaligned_subtraction`

Then compare actual `eval.Sub` result against coefficient-wise maintained q0/q1 subtraction on an independent copy to determine whether Fast `Sub` or workspace alias/copy behavior is implicated.

---

## Bypass control

If T3-3 passes but production `subAligned` fails, add one diagnostic-only bypass control to prove the intended mathematical correction itself is feasible under q0/q1.

Construct a zero-secret correction ciphertext representing `T1_actual` at exactly:

- T3-3 Level;
- T3-3 Scale;
- same NTT/Montgomery domain;
- c0 = encoded T1 logical values;
- c1 = 0.

Subtract this equal-scale correction from an independent T3-3 copy using ordinary Fast `Sub`.

Compare against:

`T3_3_actual - T1_actual`.

If this bypass passes while production `subAligned` fails, record strong evidence that the issue is scale/level alignment rather than the q0/q1 subtraction primitive or the Chebyshev recurrence itself.

This bypass is diagnostic only and must not become production code in this task.

---

## Capacity sanity

For the selected alignment branch, record whether any integer multiplication performed by `subAligned` itself risks violating the q0/q1 centered uniqueness bound.

At minimum record:

- q0*q1/2;
- alignment integer ratio;
- decoded T1 magnitude range;
- resulting nominal scaled encoding magnitude estimate where meaningful.

Do not claim capacity overflow unless supported by integer/high-precision evidence.

---

## Required classification

Choose the earliest evidence-supported outcome:

- `t3_multiply_double_or_rescale`
- `subaligned_scale_promotion`
- `subaligned_integer_multiply`
- `subaligned_metadata_scale_assignment`
- `subaligned_level_alignment`
- `subaligned_subtraction`
- `subaligned_unresolved`
- `diagnostic_inconsistency`

If a more specific internal classification is not cleanly distinguishable, use `subaligned_unresolved` and report the last passing/first failing checkpoint.

Do not repair anything yet.

---

## Required artifacts

Write temporary output to `/tmp` while any Secondary diagnostic helper exists.

After Secondary is restored clean, create in Primary:

- `results/FIX-001-DIAG-T3-logN13-fast.json`
- `results/FIX-001-DIAG-T3-logN13-summary.json`

Summary must include:

- exact repaired Fast commit;
- T1/T2 input validity;
- actual T3 production mode (lazy/strict, Mul/MulRelin);
- T3-3 post-Rescale metrics;
- production T3-4 metrics;
- exact scale ratio + BigInt ratio;
- aligned-correction-vs-T1 metrics;
- bypass-control metrics when applicable;
- first supported cause;
- confirmation that corrected T0 oracle is used;
- final clean-state/test evidence.

---

## Validation

Before completion:

- Primary `go test ./...` passes;
- repaired Secondary `go test ./...` passes after all temporary helpers are removed;
- Secondary remains exact `87be78ff...` and clean;
- Primary artifacts committed/pushed normally;
- no production Lattigo source changes;
- no LogN16;
- no benchmark;
- no EXP-003.

---

## Non-goals

Do not:

- modify/revert FIX-001;
- implement a `subAligned` repair;
- alter power planner/lazy mode;
- change Mod1 parameters or scales;
- add q2+;
- loosen threshold;
- inspect/fix Paterson–Stockmeyer yet;
- run LogN16;
- benchmark performance;
- start EXP-003.

The deliverable is a precise localization of the first failing operation in formal repaired `T3`.