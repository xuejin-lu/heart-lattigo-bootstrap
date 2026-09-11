# EXP-002-C-DIAG — Bootstrap Stage Numerical Localization

## Purpose

Localize the first bootstrap stage at which the Fast backend diverges numerically from Standard for the exact formal LogN13 workload.

This task exists because EXP-002-C established that:

- Standard `5dbffb...` is numerically correct on LogN13 under the generated-secret oracle.
- Fast `6f1c957...` and Fast `ce79b861...` both fail under the authoritative zero-secret oracle with approximately the same final error.
- ordinary decrypt/decode works for both secret views, so this is now a real numerical-path problem rather than an output-recovery incompatibility.

This task is diagnostic only.

Do not fix Lattigo, do not run LogN16, do not optimize performance, and do not begin EXP-003.

## Exact commits

Primary diagnostic runner base:

- `3eaf4321328132518625c8e3aaebb90015428d35`

Standard backend:

- `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Fast backend used by EXP-001-P:

- `6f1c957f95d0dfef6bf73bc01413a7ab55cfb8dc`

Fast backend used by EXP-002:

- `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Start with `ce79b861...` for localization because it is the current Fast branch head and has the public stage API required by this experiment.

Only after the first-divergence stage is identified on `ce79b861...`, optionally verify that the same stage fails on `6f1c957...` if the needed public stage API exists there without modifying that pinned commit.

Do not change either pinned Fast commit during this task.

## Fixed profile

Use only:

- `configs/bootstrap_config.logN13.json`

Effective profile:

- LogN = 13
- N = 8192
- full 4096 complex slots
- `ModUpThenEncode`
- current formal parameter set
- `LogMessageRatio = 10`
- `K = 16`
- `Mod1Degree = 30`
- `DoubleAngle = 3`
- `Mod1InvDegree = 0`

Do not change any profile parameter during stage localization.

## Fixed input

Use exactly the same deterministic logical input already used by the timing and correctness harness:

```text
real = ((i % 7) - 3) / 16
imag = ((i % 5) - 2) / 32
value = real + imag*i
```

Use all 4096 slots.

## Stage pipeline to inspect

Run the same sequential public bootstrap stage API already used by EXP-002:

1. `PackAndSwitchN1ToN2`
2. `ScaleDown`
3. `ModUp`
4. `CoeffsToSlots`
5. `EvalMod` real
6. `EvalMod` imag if non-nil
7. `SlotsToCoeffs`
8. `UnpackAndSwitchN2ToN1`

For the current full-slot LogN13 profile, packing, ring-degree switching, and Trace are expected to be logically trivial, but they remain explicit boundaries in the diagnostic record.

## Core diagnostic rule

Do not compare only metadata.

At every stage boundary, capture enough numerical information to determine whether Fast and Standard still represent the same logical quantity.

The objective is:

> Find the earliest stage whose output ceases to agree with the corresponding Standard stage output.

Do not assume in advance that EvalMod is the failing stage.

## Secret semantics

For Standard stage outputs, use the generated residual secret key when a normal residual-domain decrypt/decode is semantically valid.

For Fast stage outputs, use an all-zero secret key when a normal decrypt/decode is semantically valid.

Do not use generated-secret Fast decode as the authoritative oracle.

Do not import `schemes/ckks/fast` into Primary merely to decode internals.

## Important: not every intermediate stage is directly CKKS-slot-decodable

A stage output may not represent ordinary user slots in the same domain/scale/layout as the original input.

Therefore use a stage-appropriate comparison strategy.

Priority order:

1. Prefer ordinary public decrypt + CKKS decode when the stage output is a logical CKKS slot vector.
2. If a stage output is in coefficient/DFT intermediate representation, compare Standard and Fast at the same public-stage boundary using backend-neutral polynomial/ciphertext data that is mathematically comparable.
3. Normalize representation differences that are metadata/domain-only and explicitly documented, such as Montgomery form, before comparison.
4. Do not invent a semantic inverse transform solely for the diagnostic runner unless it is already available through the ordinary public API.
5. Never compare stale/dormant Fast limbs `q2...` as authoritative data. Fast authoritative arithmetic state is q0/q1.

The diagnostic artifact must clearly state what quantity is being compared at each boundary.

## Required stage comparisons

### 0. Input sanity

Before Bootstrap stages:

- confirm Standard and Fast runs start from the identical logical input formula;
- confirm effective parameter identity;
- decode the initial constructed input through the meaningful oracle and verify it represents the intended values.

### 1. PackAndSwitchN1ToN2

Record whether the boundary is trivial for this profile.

If N1=N2 and one ciphertext/full slots make this logically trivial, prove that from effective parameters and record it. A numerical transform comparison is optional if the output is literally unchanged except storage/copy semantics.

### 2. ScaleDown

Compare Standard and Fast outputs at the ScaleDown boundary.

At minimum compare:

- level
- scale
- domain flags
- authoritative q0 coefficient/NTT representation after normalizing Montgomery form if required
- decoded/logical values if ordinary decode is valid at this point
- `errScale` from Standard and Fast

The current implementation already has unit-level evidence that Fast ScaleDown can match Standard for LogMessageRatio=10. This formal diagnostic should verify the exact experiment input/profile, not rely only on the unit test.

### 3. ModUp

Compare Standard and Fast after complete ModUp including Trace.

For this full-slot profile, also record:

- trace gap
- number of trace automorphism iterations
- whether Trace is actually trivial

Compare authoritative q0/q1 values after representation normalization.

Do not compare dormant Fast q2... limbs as correctness evidence.

### 4. CoeffsToSlots

Compare both outputs independently:

- real branch
- imaginary branch

Record:

- level
- scale
- domain flags
- a backend-neutral numerical comparison of corresponding Standard/Fast branch contents
- decoded values if the public representation is valid for CKKS decode at this boundary

This is a critical boundary because the DFT implementation showed the largest performance speedup in EXP-002.

### 5. EvalMod real / imag

Compare Standard and Fast outputs separately for real and imaginary branches.

Record full numerical errors between corresponding branch outputs.

This is also a critical boundary because existing Fast EvalMod numerical tests do not formally cover the exact `LogMessageRatio=10`, K16, degree30, DoubleAngle3 experiment profile at strict `1e-2` accuracy.

Do not change any parameter during this task.

### 6. SlotsToCoeffs

Compare Standard and Fast outputs immediately after SlotsToCoeffs and before public final unpack/finalization.

Where appropriate, normalize Montgomery representation before numerical comparison.

### 7. Final public output

Reconfirm the already observed result:

- Standard generated-secret decode vs input
- Fast zero-secret decode vs input
- Fast zero-secret decode vs Standard generated-secret decode

This final comparison is expected to fail and serves as consistency evidence with EXP-002-C.

## Numerical metrics

Where stage quantities are directly comparable as complex values, compute:

```text
max_abs_complex
mean_abs_complex
rmse_complex
max_abs_real
max_abs_imag
max_component_abs
```

Record maximum-error slot/index.

Where comparing polynomial/RNS data rather than decoded complex slots, compute exact residue equality where expected and/or a centered numeric delta under q0/q1 with an explicitly documented normalization.

Do not force slot-error metrics onto an intermediate representation where they are not mathematically meaningful.

## First-divergence classification

For each stage, classify:

```text
MATCH
SMALL_NUMERICAL_DRIFT
DIVERGED
NOT_DIRECTLY_COMPARABLE
TRIVIAL_BOUNDARY
```

Use `MATCH` for exact/near-exact agreement consistent with that stage's expected arithmetic.

Use `DIVERGED` only when the difference is clearly too large to be explained by normal CKKS approximation or representation normalization.

The summary must identify one of:

```text
first_divergence_stage = <stage name>
```

or, if the evidence is insufficient:

```text
first_divergence_stage = unresolved
```

Do not force a conclusion if a stage cannot be compared correctly.

## Output artifacts

Create diagnostic artifacts only after the runner is finalized and clean.

Suggested files:

```text
results/EXP-002C-DIAG-logN13-standard.json
results/EXP-002C-DIAG-logN13-fast.json
results/EXP-002C-DIAG-logN13-summary.json
```

The summary must contain:

- exact Primary commit
- exact Standard/Fast commits
- clean-state identity checks
- exact effective profile
- stage-by-stage comparison method
- stage-by-stage metrics/classification
- first divergence stage or unresolved
- evidence explaining why the classification is valid
- confirmation that no LogN16 run occurred
- confirmation that no Lattigo code was modified

## Implementation constraints

- Keep Primary backend-agnostic.
- Use ordinary bootstrapping public stage API.
- No Fast imports in Primary.
- No backend-specific arithmetic path in Primary.
- It is acceptable for result interpretation to label the known backend commit as Standard/Fast.
- No Lattigo modifications.
- No parameter sweep.
- No performance optimization.
- No EXP-003.
- No LogN16.
- Do not overwrite EXP-001/EXP-002 timing evidence or EXP-002-C runner history.

## Validation

At minimum:

```text
go test ./...
```

must pass against:

- Standard `5dbffb...`
- Fast `ce79b861...`

If diagnostic logic adds normalization or comparison helpers, add focused Primary tests for them.

## Stop conditions

Stop immediately and report if:

- a public stage cannot be compared without adding Fast-specific Primary semantics;
- the diagnostic requires modifying a pinned Lattigo commit;
- repository identity/clean provenance cannot be maintained;
- a comparison method would be mathematically invalid for that stage.

Otherwise stop after the LogN13 first-divergence evidence is committed.

Do not fix the detected bug in the same task.

## Completion criteria

This diagnostic task is COMPLETE when:

1. the exact LogN13 formal profile is used unchanged;
2. Standard `5dbffb...` and Fast `ce79b861...` are compared through the same public stage sequence;
3. every relevant stage is classified with a documented valid comparison method;
4. the earliest supported divergence is identified, or explicitly marked unresolved with the blocking reason;
5. final output failure is reproduced consistently;
6. no LogN16 experiment is run;
7. no Lattigo code is modified;
8. no optimization or parameter sweep is started;
9. diagnostic artifacts and summary are committed;
10. `CURRENT_TASK.md` is marked COMPLETE only after the diagnostic evidence commit.

After this task, stop for independent review before any bug fix task is created.