# FIX-001-P3-DIAG-PS-NORMALIZE — Validate giant-step metadata scale normalization

## Purpose

The compressed-scale diagnostic at Primary commit:

`605d864d02b87793530e6bf53e3d67cd800c7503`

using exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

proved:

- candidate internal scales `2^91 ... 2^86` all pass prechecks;
- all five baby blocks pass for every candidate;
- minimum coefficient encoding precision stays about `2^31 ... 2^26`, above the fixed `2^20` floor;
- first giant-step `Rescale(b)` semantic oracle passes;
- first giant-step `Mul(b,T8)` semantic oracle passes;
- every candidate stops only because the product Scale fails the production `Scale.InDelta(a.Scale, ScalePrecision-12)` gate;
- example at candidate `2^91`:
  - `a.Scale = 2.475880078570760549798248448e27`
  - `b.Scale = 2.47588007856713654229323309002828985927e27`
  - relative mismatch ~`1.46e-12` (~`2^-39.3`)
- current gate requires relative mismatch <= about `2^-116` because `Scale.InDelta(x,116)` means `-log2(relative error) >= 116`;
- the arithmetic values themselves are still correct at the failing checkpoint.

This task must test whether the compressed PS schedule can proceed correctly by **normalizing metadata Scale only** after a semantically-correct giant-step product, instead of rejecting the tiny deterministic scale drift.

Diagnostic only. Do not modify production Lattigo.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`605d864d02b87793530e6bf53e3d67cd800c7503`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Pinned Standard reference:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Both worktrees must start clean. Follow both repositories' `AGENTS.md`.

Parent FIX-001/P3 remains incomplete.

---

# Core hypothesis

For compressed Fast PS, after:

```text
b = Rescale(b)
b = Mul(b, xpow)
```

the represented value is already numerically correct, but the actual metadata Scale differs from the sibling accumulator `a.Scale` by a tiny deterministic relative factor caused by the compressed internal schedule and real RNS divisors.

If:

1. local semantic oracle for the product passes;
2. relative Scale drift is sufficiently small;
3. q0/q1 capacity remains safe;

then changing **only**:

`b.Scale = a.Scale`

should perturb the interpreted value only by the same tiny relative factor and allow the existing `addAligned`/Add semantics to continue.

This task must prove or reject that through the complete PS tree and final target-scale restoration.

Do not simply relax the production `InDelta` gate in this task.

---

# Formal workload

Use the exact LogN13 real-branch formal polynomial workload and corrected `T0=1` oracle.

Require:

- E2 exact historical reproduction;
- P3 powers `T1,T2,T3,T4,T6,T8,T16` all pass <= `1e-2`;
- historical ordinary P3 polynomial still fails around `0.7909176408709179`;
- compressed-scale historical control reproduces the first giant-step Scale mismatch.

If not, stop with:

`NORMALIZATION_PRECONDITION_MISMATCH`.

Do not run LogN16.

---

# Candidate internal scales

Use the same exact set:

- `2^91`
- `2^90`
- `2^89`
- `2^88`
- `2^87`
- `2^86`

Do not add candidates unless needed only as clearly labeled supplemental evidence.

---

# Diagnostic normalization rule

At each compressed giant-step after actual Fast:

```text
Rescale(b)
Mul(b, xpow, b)
```

first validate the product **before normalization**.

Record:

- `a.Scale`
- actual `b.Scale`
- exact relative difference
- `Log2Delta(a.Scale,b.Scale)`
- local semantic product error
- output Level/Degree
- q0/q1 c0 capacity ratio / outside count

Normalization is allowed diagnostically only if:

1. local product semantic error <= `1e-2`;
2. `Log2Delta(a.Scale,b.Scale) >= 32`;
3. no immediately-required q0/q1 centered interpretation is already invalid.

The fixed diagnostic bound `2^-32` is intentionally much looser than Standard's metadata bookkeeping tolerance but still far below the experiment's `1e-2` semantic threshold.

Do not silently change this bound.

If allowed:

- deep-copy/snapshot the product;
- set only `b.Scale = a.Scale`;
- do not modify q0/q1 coefficients;
- decode immediately after metadata normalization;
- compare normalized result against the same local semantic product oracle;
- record the difference between pre-normalization and post-normalization decoded vectors.

Require normalized semantic error <= `1e-2`.

If not allowed because Log2Delta < 32, classify that candidate as scale drift too large.

---

# Part 1 — reproduce compressed baby blocks

For every candidate:

- replay all five baby blocks exactly as in the prior compressed diagnostic;
- require all baby operations pass;
- archive max baby error and capacity headroom.

If any candidate that previously passed now fails, stop with diagnostic inconsistency.

---

# Part 2 — complete all giant steps with normalization

For each candidate with passing baby blocks, follow the exact PS reversal/pairing schedule.

At every paired giant step:

1. optional actual Fast Relinearize(b);
2. actual Fast Rescale(b);
3. local semantic oracle for Rescale;
4. actual Fast Mul(b,xpow,b);
5. local semantic oracle for product;
6. evaluate normalization eligibility using fixed `Log2Delta >= 32`;
7. if eligible, metadata-normalize `b.Scale = a.Scale` only;
8. local semantic oracle immediately after normalization;
9. execute the existing actual `addAligned(a,b)` or direct Add path as production would after equal scales;
10. local semantic oracle for `A + B*X`.

For every step record:

- round index;
- block/even/odd indices;
- xpow key;
- Levels/Degrees;
- pre/post scales;
- Log2Delta;
- normalization applied yes/no;
- max semantic errors;
- q0/q1 capacity evidence before every Rescale.

Stop causal interpretation at the first failure for each candidate, but complete other candidates independently.

Classify candidate giant outcome as one of:

- `giant_normalization_drift_too_large`
- `giant_rescale_failure`
- `giant_multiply_failure`
- `giant_metadata_normalization_failure`
- `giant_add_failure`
- `giant_all_pass`.

---

# Part 3 — finalization

For candidates where all giant steps pass:

1. optional final Relinearize;
2. record pre-final-Rescale q0/q1 capacity;
3. actual Fast final Rescale;
4. compare against source-backed degree-30 polynomial oracle.

Require <= `1e-2`.

Record resulting compressed output Scale.

---

# Part 4 — final public target-scale restoration

For each candidate whose compressed final polynomial passes, restore the original public targetScale using the same method proposed previously:

Let:

`ratio = targetScale / compressedOutput.Scale`.

Choose:

`M = nearest positive integer(ratio)`

with high-precision arithmetic.

On an independent maintained copy:

1. prove multiplying maintained q0/q1 coefficients by `M` remains within required centered uniqueness for subsequent semantics;
2. `MulIntegerMaintained(result,M,result)`;
3. `result.Scale *= M`;
4. require resulting Scale has adequate closeness to public targetScale;
5. normalize metadata to exact public targetScale only if closeness is explicitly recorded and acceptable;
6. compare all 4096 slots to the degree-30 polynomial oracle.

Require <= `1e-2`.

If integer promotion is unsafe or inaccurate, record that as the first finalization failure.

---

# Part 5 — preferred candidate

Among candidates that complete:

- all baby blocks;
- all normalized giant steps;
- final Rescale;
- final target-scale restoration;
- whole polynomial <= `1e-2`;

choose preferred candidate deterministically:

1. highest internal scale;
2. all giant-step `Log2Delta >= 32`;
3. all q0/q1 capacity checks safe;
4. minimum coefficient encoding precision >= `2^20`;
5. then smallest whole-polynomial max component error.

Record minimum giant-step Log2Delta and maximum semantic perturbation caused by metadata normalization.

---

# Required overall classification

Choose one:

## Case A

`FIRST_SUPPORTED_CAUSE = compressed_ps_with_scale_normalization_validated`

At least one candidate passes the complete polynomial plus final target-scale restoration.

## Case B

`FIRST_SUPPORTED_CAUSE = giant_scale_drift_exceeds_normalization_bound`

All viable candidates encounter `Log2Delta < 32` before completion.

## Case C

`FIRST_SUPPORTED_CAUSE = giant_metadata_normalization_semantic_failure`

Scale drift is small enough but metadata-only normalization itself causes semantic error > `1e-2`.

## Case D

`FIRST_SUPPORTED_CAUSE = compressed_giant_arithmetic_failure_after_normalization`

Normalization works but a later giant Rescale/Mul/Add fails.

## Case E

`FIRST_SUPPORTED_CAUSE = compressed_finalization_failure_after_normalization`

All giant steps pass but final Rescale or target-scale restoration fails.

Do not implement production repair in this task.

---

# Required artifacts

After restoring exact clean Secondary `61607bb...`, commit to Primary:

- `results/FIX-001-P3-DIAG-PS-NORMALIZE-logN13-fast.json`
- `results/FIX-001-P3-DIAG-PS-NORMALIZE-logN13-summary.json`

Summary must include:

- exact provenance;
- candidate table;
- historical mismatch control;
- per-giant-step Scale / Log2Delta table;
- pre/post normalization semantic errors;
- q0/q1 capacity evidence;
- final Rescale result where reached;
- target-scale restoration result where reached;
- preferred candidate if any;
- first supported cause;
- test/clean-state evidence.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- Secondary `go test ./...` passes after all temporary helpers are removed;
- Secondary remains exact `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Lattigo changes;
- Primary artifacts committed/pushed normally;
- both worktrees clean;
- no formal production Gate 4/5;
- no LogN16;
- no benchmark;
- no EXP-003.

---

# Non-goals

Do not:

- relax production `Scale.InDelta` directly;
- modify `Scale.InDelta` globally;
- modify Standard Lattigo;
- change source polynomial or Mod1 parameters;
- change P3 power generation;
- add q2+ or auxiliary limbs;
- permanently lower public target Scale;
- loosen semantic threshold;
- run LogN16;
- benchmark;
- start EXP-003.

The deliverable is whether metadata-only giant-step scale normalization makes the q0/q1-safe compressed PS schedule complete and numerically correct end to end.