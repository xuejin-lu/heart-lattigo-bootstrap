# FIX-001-P3-DIAG-P93-PRODUCTION-C2S-FIRST-MATERIAL-DIVERGENCE

## Purpose

The preceding diagnostic at Primary commit `b47ea3a3a56aa1f3c037a2af36c8cdde46930064` freshly replayed the historical P93 reference and compared it with the current dirty production implementation.

It established:

- fresh historical P93 reference replay succeeds:
  - real EvalMod vs Standard: `0.0001635104343335875`
  - imag EvalMod vs Standard: `0.00013359983063118935`
  - post-S2C proxy: `0.0003032931504531125`
  - public-like: `0.009705381393898434` (passes the current `1e-2` system milestone)
- reference-vs-production P93 PS/DoubleAngle arithmetic has no material divergence:
  - first observable mismatch: real T2 around `6.94e-9`
  - T16 around `9.1e-7`
  - PS final around `5.0e-7`
  - DoubleAngle aligned differences around `1e-14`
  - classification: `P93_NO_MATERIAL_REFERENCE_PRODUCTION_DIVERGENCE`
- current actual full production core still shows:
  - EvalMod real vs matched Standard: `0.022317936759951508`
  - EvalMod imag vs matched Standard: `0.020569287028685726`
- but when EvalMod is started from the controlled/stage-aligned Fast C2S diagnostic inputs, the same production EvalMod implementation gives only:
  - real: `0.004183019582623534`
  - imag: `0.004261561142162278`

Therefore the current evidence does **not** support PS/EvalMod arithmetic as the primary source of the full-core `~0.02` gap.

The important path distinction is:

1. **actual production core**
   `Pack -> ScaleDown -> ModUp -> Fast CoeffsToSlots -> Fast EvalMod`

versus

2. **controlled matched-C2S diagnostic path**
   `Pack -> ScaleDown -> ModUp -> source-faithful compressed/aligned C2S diagnostic -> Fast EvalMod`.

This task must determine whether the missing error is introduced by the **actual production Fast C2S path**, and if so locate the first material C2S group/boundary.

Do not repair anything in this task.

---

## Repository safety

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Follow `AGENTS.md` startup preflight, synchronize `main`, then re-read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary

Repository: `xuejin-lu/lattigo`

Required local production state:

- branch: `fast-ckks`
- committed HEAD: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact branch/HEAD/dirty fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed profile

Keep exactly:

- LogN13
- q0 = 56-bit effective profile
- q1 approximately 39 bits
- q2 approximately 40 bits
- P93 / planScale = `2^93`
- deterministic 4096-slot workload
- same Pack/ScaleDown/ModUp
- same C2S matrices, levels and factorization
- same EvalMod parameters
- current diagnostic system milestone `1e-2`

No tuning.

---

# R0 — reproduce the two C2S-to-EvalMod paths

Using the unchanged dirty production state, reproduce in one run:

## Path P — actual production

Starting from the same `reproducibleInput`:

`PackAndSwitchN1ToN2 -> ScaleDown -> ModUp -> FastEvaluator.CoeffsToSlots -> FastEvaluator.EvalMod`

Record:

- actual production ModUp output
- production C2S real/imag output
- final production EvalMod real/imag

Expected full-core EvalMod magnitudes:

- real ~ `0.02231793676`
- imag ~ `0.02056928703`

## Path D — controlled matched-C2S diagnostic

Use the established `evalModMatchedC2SInputs` machinery and its controlled Fast C2S outputs, then run the same P93 Fast EvalMod implementation.

Record:

- matched diagnostic ModUp input
- diagnostic Fast C2S real/imag
- final diagnostic-path EvalMod real/imag

Expected stage-aligned magnitudes:

- real ~ `0.00418301958`
- imag ~ `0.00426156114`

Also retain:

- aligned Standard C2S real/imag
- ordinary Standard C2S real/imag

If either path fails to reproduce current committed evidence while provenance is unchanged, stop with:

`P93_C2S_BOUNDARY_REPLAY_CONFLICT`.

---

# R1 — prove whether the divergence already exists at C2S output

Before examining groups, compare the actual semantic vectors directly.

Required comparisons:

1. production ModUp input vs diagnostic Fast ModUp input
2. production Fast C2S real vs diagnostic Fast C2S real
3. production Fast C2S imag vs diagnostic Fast C2S imag
4. diagnostic Fast C2S vs aligned Standard C2S
5. diagnostic Fast C2S vs ordinary Standard C2S
6. production Fast C2S vs aligned Standard C2S
7. production Fast C2S vs ordinary Standard C2S

For every comparison record:

- max component
- max complex
- mean complex
- worst slot/component
- Level/Scale/Degree
- metadata equality
- q0/q1 capacity state where meaningful

The goal is to distinguish:

- input mismatch before C2S;
- actual Fast C2S implementation divergence;
- Standard-reference-definition mismatch.

If production and diagnostic ModUp inputs are not semantically aligned, classify upstream before any C2S claim.

---

# R2 — source-faithful C2S stage trace

If R1 shows production and diagnostic ModUp inputs align but C2S outputs differ materially, trace the C2S transform group-by-group.

Read the current local dirty Secondary implementation to determine the exact production `CoeffsToSlots` operation order, but do not modify it.

Build diagnostic tracing in Primary or temporary code only.

The trace must be source-faithful to the actual current dirty production path and must distinguish it from the existing controlled compressed/aligned diagnostic path.

At minimum compare, where applicable:

1. C2S input / ModUp output
2. group 0 pre-transform
3. group 0 post-transform
4. group 0 post-Rescale
5. group 0 restore/compression boundary
6. group 1 post-transform
7. group 1 post-Rescale
8. group 1 restore/compression boundary
9. group 2 post-transform / post-Rescale
10. group 3 post-transform / post-Rescale
11. conjugate/split boundary
12. final real output
13. final imag output

For each stage compare:

- actual production source-faithful path
- controlled diagnostic Fast path
- stage-aligned Standard path where semantically meaningful

Record compactly:

- production-vs-diagnostic max component
- production-vs-aligned-Standard max component
- diagnostic-vs-aligned-Standard max component
- Level/Scale/Degree
- row hash equality only where exact equality is mathematically required
- capacity
- amplification from prior stage

Do not commit slot vectors or RNS dumps.

---

# R3 — downstream causal confirmation

If C2S output divergence is found, prove that it explains the difference between the two EvalMod measurements.

Let:

- `C_P` = actual production C2S output
- `C_D` = controlled diagnostic Fast C2S output
- `E(C)` = the same production P93 EvalMod map

Compute:

- `E(C_P)`
- `E(C_D)`
- semantic delta `E(C_P) - E(C_D)`

Compare this delta with the observed difference between:

- full-core production EvalMod error (~0.022 / ~0.0206)
- stage-aligned EvalMod behavior (~0.00418 / ~0.00426)

Do not subtract max norms as if they were vectors. Use actual decoded vectors for closure/agreement.

Record:

- C2S input delta max component
- EvalMod output delta max component
- empirical amplification
- vector closure/agreement residual

Target agreement floor: `<=1e-10` when supported by the same deterministic run, otherwise report the justified measured floor.

---

# R4 — first observable and first material C2S divergence

Report separately:

## First observable divergence

Earliest source-faithful C2S checkpoint above the deterministic numerical/replay floor.

## First material divergence

Earliest checkpoint where production-vs-controlled-diagnostic difference reaches at least 10% of the final actual-vs-diagnostic C2S output difference, unless a stricter justified criterion is available.

For the transition around the first material divergence report:

- incoming difference
- outgoing difference
- amplification
- Level/Scale/Degree
- capacity
- exact production operation involved

Do not infer causality from the first tiny row mismatch alone.

---

# R5 — reconcile Standard references

Because prior diagnostics used both aligned and ordinary Standard C2S paths, produce a compact reference table for:

1. aligned Standard C2S
2. ordinary Standard C2S
3. controlled diagnostic Fast C2S
4. actual production Fast C2S

For real and imag state:

- pairwise max-component semantic differences
- intended role of each object
- which Standard object is the valid comparison for each diagnostic question

Then state explicitly which reference must be used for:

- Fast C2S implementation fidelity
- Fast EvalMod stage fidelity
- full production core correctness
- final public E2E correctness

This is required to prevent another `0.004` vs `0.022` reference mix-up.

---

# Decision classification

Choose exactly one:

## A — `P93_C2S_BOUNDARY_REPLAY_CONFLICT`

One of the two known C2S-to-EvalMod paths does not reproduce.

## B — `P93_PRE_C2S_INPUT_DIVERGENCE`

Production and controlled diagnostic paths already differ materially at ModUp/C2S input.

Next task localizes Pack/ScaleDown/ModUp.

## C — `P93_FAST_C2S_GROUP0_DIVERGENCE`

Inputs align; first material divergence is in group 0.

## D — `P93_FAST_C2S_GROUP1_DIVERGENCE`

First material divergence is in group 1.

## E — `P93_FAST_C2S_GROUP2_OR_GROUP3_DIVERGENCE`

First material divergence is in group 2 or group 3.

## F — `P93_FAST_C2S_SPLIT_DIVERGENCE`

Matrix groups align; first material divergence appears at conjugate/split/final real-imag construction.

## G — `P93_STANDARD_REFERENCE_SEMANTICS_MISMATCH`

Actual and diagnostic Fast C2S outputs are sufficiently aligned; the apparent full-core gap is caused by comparing against different Standard semantics/reference objects.

## H — `P93_C2S_TRACE_ALIGNMENT_INVALID`

The source-faithful production trace cannot be aligned reliably with the diagnostic/reference path.

Stop after classification. No repair in this task.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-PRODUCTION-C2S-FIRST-MATERIAL-DIVERGENCE-summary.json`

Include only:

- Primary provenance
- Secondary branch/HEAD/dirty fingerprint
- fixed profile
- R0 path reproduction
- R1 boundary comparisons
- compact R2 group trace
- R3 downstream vector causal confirmation
- first observable/material divergence
- R5 Standard-reference table
- one classification A-H
- explicit confirmation that Secondary was not modified

Detailed vectors and row traces remain under `/tmp`.

---

## Validation

Run:

- focused C2S diagnostic test(s)
- directly affected Primary tests
- `git diff --check`

On successful completion commit/push only Primary diagnostic code and compact evidence under `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no q/planScale tuning
- no PS repair
- no EvalMod repair
- no S2C repair
- no finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is the first material **production C2S/reference-boundary** divergence, not a fix.
