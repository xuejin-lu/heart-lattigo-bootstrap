# EXP-002-C — End-to-End Bootstrap Numerical Correctness Closure

## Purpose

Close the numerical-correctness evidence gap before beginning EXP-003 primitive profiling.

EXP-001 / EXP-001-P and EXP-002 already established trustworthy performance and stage-timing evidence, but their formal experiment outputs primarily checked successful execution, output metadata, parameter identity, and matched workload. They did not archive a formal full-slot numerical comparison of decoded Standard and Fast bootstrap outputs.

This task answers one question only:

> For the exact full-slot LogN13 and LogN16 workload used by the formal performance experiments, does a complete Fast bootstrap preserve the same logical values as Standard and the original input within CKKS-appropriate approximation error?

Do not optimize performance, do not add primitive profiling, and do not begin EXP-003 in this task.

## Relationship to previous experiments

This is a correctness companion to the existing performance evidence. It must not overwrite, reinterpret, or rerun the timing datasets from EXP-001 / EXP-001-P / EXP-002.

Use the exact formal configuration profiles already present:

- `configs/bootstrap_config.logN13.json`
- `configs/bootstrap_config.logN16.json`

These correspond to the same current compatibility workload used by the formal experiments:

- `ModUpThenEncode`
- full slots (`log_slots = -1` in the source config)
- LogN13: N=8192, 4096 logical complex slots
- LogN16: N=65536, 32768 logical complex slots

Do not change either config for this task.

## Backend commits to validate

Validate the exact implementation versions that underpin the formal performance claims.

Standard baseline:

- `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Fast backend used by the completed EXP-001-P / LogN16 baseline:

- `6f1c957f95d0dfef6bf73bc01413a7ab55cfb8dc`

Fast backend used by EXP-002 stage profiling:

- `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

The EXP-002 Fast commit is a descendant of the EXP-001-P Fast commit, but both exact commits must be validated so the archived correctness evidence maps directly to both completed experiment phases.

Use detached-HEAD backend switching and the safe restore procedure from `AGENTS.md`. Do not modify either pinned Lattigo commit.

## Fixed input workload

The numerical input must be exactly the same deterministic logical workload already used by the Primary timing harness in `reproducibleInput`.

For every logical slot index `i`:

```text
real = ((i % 7) - 3) / 16
imag = ((i % 5) - 2) / 32
value = real + imag*i
```

Use all logical slots for the profile.

Do not introduce random values, reduced slot counts, sampled-only correctness, or a new synthetic workload. The point of this task is to validate the exact logical data pattern used by the speed experiments.

## Correctness execution path

Add a Primary correctness mode/runner that uses the ordinary public Lattigo API and remains backend-agnostic.

The Primary repository must not import Fast packages, inspect Fast key layouts, check Fast commit identities to branch execution, or add `--fast` / `--standard` behavior.

For each run, the ordinary path should be conceptually:

1. Build the same residual and bootstrapping parameters from the selected existing config.
2. Generate `sk` with the ordinary residual-parameter key generator.
3. Generate bootstrapping evaluation keys through ordinary `btpParams.GenEvaluationKeys(sk)`.
4. Construct the ordinary `bootstrapping.NewEvaluator(...)`.
5. Construct the existing deterministic full-slot input ciphertext using the same logical values as `reproducibleInput`.
6. Execute one complete `eval.Bootstrap(input)`.
7. Recover the logical plaintext output through one backend-neutral public path.
8. Decode all logical slots.
9. Compare the decoded output against the deterministic original values.
10. Archive the decoded output and error metrics.

### Output recovery requirement

Prefer the normal public decrypt-and-decode path for both backends:

```text
rlwe.NewDecryptor(residualParams, sk).DecryptNew(output)
ckks.NewEncoder(residualParams).Decode(...)
```

This must be attempted and validated with the same Primary code against Standard and both Fast commits.

If the current Fast backend cannot produce a logically correct result through this ordinary backend-neutral recovery path, **stop and report the exact incompatibility**. Do not silently decode Fast `c0` directly, do not add a backend selector, and do not teach the Primary repository Fast-specific ciphertext semantics merely to make the test pass.

A backend compatibility change, if truly required, must be reviewed as a separate Lattigo task before this correctness experiment is resumed.

## Formal run matrix

Run exactly these six formal correctness runs with one finalized Primary source commit:

1. LogN13 Standard `5dbffb...`
2. LogN13 Fast EXP-001-P `6f1c957...`
3. LogN13 Fast EXP-002 `ce79b861...`
4. LogN16 Standard `5dbffb...`
5. LogN16 Fast EXP-001-P `6f1c957...`
6. LogN16 Fast EXP-002 `ce79b861...`

This is a numerical-correctness experiment, not a timing benchmark. One complete bootstrap per backend/profile is sufficient because the goal is to compare the exact full-slot logical transformation, not to estimate runtime variance.

Do not report or reuse elapsed time from these runs as performance evidence.

## Clean provenance

Use the same clean-provenance discipline as the formal speed experiments.

Before every formal run:

- Primary is clean and at the same finalized correctness-runner commit.
- Secondary is clean and detached at the exact pinned backend commit.
- write formal output to `/tmp` first so creating the output does not make Primary dirty.
- raw output must record `primary_repository.dirty=false` and `lattigo_repository.dirty=false`.

Only after all run identity checks pass may outputs be copied into `results/` and committed.

## Raw correctness artifacts

Archive one raw JSON per formal run:

```text
results/EXP-002C-logN13-standard.json
results/EXP-002C-logN13-fast-exp001p.json
results/EXP-002C-logN13-fast-exp002.json
results/EXP-002C-logN16-standard.json
results/EXP-002C-logN16-fast-exp001p.json
results/EXP-002C-logN16-fast-exp002.json
```

Each raw file must contain at least:

- schema version
- timestamp
- Primary repository commit/ref/dirty state
- Lattigo repository commit/ref/dirty state
- environment metadata
- config
- effective parameters
- deterministic input workload identifier/formula
- logical slot count
- complete decoded output for all logical slots, represented losslessly enough for recomputation of the stored floating-point error metrics
- output level
- output scale
- output NTT/Montgomery metadata if available
- error metrics against the original deterministic input

Do not truncate the LogN16 output to a sample. Full-slot evidence is required.

## Required numerical metrics

For two complex vectors `a` and `b`, calculate at least:

```text
max_abs_complex = max_i |a_i - b_i|
mean_abs_complex = mean_i |a_i - b_i|
rmse_complex = sqrt(mean_i |a_i - b_i|^2)
max_abs_real = max_i |Re(a_i) - Re(b_i)|
max_abs_imag = max_i |Im(a_i) - Im(b_i)|
max_component_abs = max(max_abs_real, max_abs_imag)
```

Also record the slot index/indices producing the maximum component and complex errors so an outlier can be inspected directly.

Optionally report an approximate precision indicator:

```text
precision_bits = -log2(max_abs_complex)
```

If `max_abs_complex == 0`, store a JSON-safe explicit representation such as `null` plus a reason instead of emitting NaN/Inf.

## Comparisons that must be archived

For each profile, compute:

### A. Backend output vs original input

- Standard vs input
- Fast EXP-001-P vs input
- Fast EXP-002 vs input

### B. Fast vs Standard

- Fast EXP-001-P vs Standard
- Fast EXP-002 vs Standard

The Fast-vs-Standard comparison must be computed slot-by-slot from the archived decoded outputs, not inferred from separate summary metrics.

## Acceptance threshold

Use the existing implementation-level Fast bootstrap numerical contract as the initial formal acceptance threshold:

```text
max_abs_real <= 1e-2
max_abs_imag <= 1e-2
```

Require this threshold for:

- Standard vs original input
- each Fast backend vs original input
- each Fast backend vs Standard

Do not weaken the threshold after observing the data. If any comparison fails, archive the evidence, mark the profile failed, and stop for review.

The summary must still report the actual measured errors even when they are much smaller than the threshold.

## Pair/profile summaries

Create:

```text
results/EXP-002C-logN13-summary.json
results/EXP-002C-logN16-summary.json
results/EXP-002C-matrix-summary.json
```

Each profile summary must include:

- exact Primary correctness-runner commit
- exact Standard commit
- exact Fast EXP-001-P commit
- exact Fast EXP-002 commit
- config/effective-parameter identity checks
- clean-state checks
- slot count
- output metadata checks
- all input-reference numerical metrics
- both Fast-vs-Standard numerical metric sets
- threshold used
- pass/fail for each comparison
- overall profile pass/fail
- references to the raw result files

The matrix summary must make it obvious whether the exact backend commits behind EXP-001-P and EXP-002 both pass at LogN13 and LogN16.

## Implementation constraints

- Keep the Primary frontend backend-agnostic.
- Reuse existing parameter/config/input construction where practical instead of duplicating it.
- Do not alter the timing paths or timing result files from EXP-001 / EXP-002.
- Do not modify the existing raw timing measurements or summaries.
- Do not change bootstrapping parameters.
- Do not modify Lattigo unless a real public-API compatibility gap is discovered; if one is discovered, stop and report before making that secondary change.
- Do not add noise-fidelity research beyond what is necessary to decode and compare the existing correctness workload.

## Validation

After the Primary correctness runner is finalized, verify it compiles and its tests pass against all three pinned backend commits.

At minimum:

```text
go test ./...
```

must pass with:

- Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Fast `6f1c957f95d0dfef6bf73bc01413a7ab55cfb8dc`
- Fast `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Add focused Primary tests for the error-metric calculation and JSON summary logic if new code is introduced for them.

## Completion criteria

EXP-002-C passes only when:

1. all six formal full-slot correctness runs are archived from one clean finalized Primary correctness-runner commit;
2. all backend/config/environment identity checks pass;
3. LogN13 validates all 4096 logical complex slots;
4. LogN16 validates all 32768 logical complex slots;
5. Standard vs input passes the fixed numerical threshold for both profiles;
6. Fast `6f1c957...` vs input and vs Standard passes for both profiles;
7. Fast `ce79b861...` vs input and vs Standard passes for both profiles;
8. actual error magnitudes and outlier indices are archived;
9. no previous timing evidence has been modified;
10. both repositories are restored safely after the run.

Only after independent review marks EXP-002-C PASS should the project advance to EXP-003 primitive profiling.

## Stop condition

Stop after EXP-002-C results are committed and `CURRENT_TASK.md` is marked `COMPLETE`.

Do not start EXP-003 automatically.