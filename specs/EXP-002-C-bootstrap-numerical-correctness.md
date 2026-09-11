# EXP-002-C — End-to-End Bootstrap Numerical Correctness Closure

## Purpose

Close the numerical-correctness evidence gap before beginning EXP-003 primitive profiling.

EXP-001 / EXP-001-P and EXP-002 established performance and stage-timing evidence, but their formal outputs did not archive a full-slot numerical comparison of Standard and Fast bootstrap results.

This task answers:

> For the exact full-slot LogN13 and LogN16 workload used by the formal performance experiments, do the Standard and Fast backends preserve the same logical values within CKKS-appropriate approximation error, under each backend's actual ciphertext secret semantics?

Do not optimize performance, do not add primitive profiling, and do not begin EXP-003.

## Important correction after the first halted attempt

The first EXP-002-C attempt used the generated residual secret key `sk` as the only decryption oracle for both Standard and Fast. That is not a valid Fast correctness oracle.

The current Fast backend deliberately uses zero-secret evaluation-key semantics. Existing Fast bootstrap tests explicitly validate the public output using an all-zero secret key through the ordinary `rlwe.NewDecryptor` + `ckks.NewEncoder.Decode` path.

Therefore:

- Standard's meaningful decryption oracle is the generated residual secret key.
- Fast's meaningful decryption oracle is an all-zero residual secret key.
- A generated-key decode of a Fast output is a diagnostic mismatch, not by itself a numerical-correctness failure.
- A zero-key decode of a Standard output is likewise diagnostic only.

The Primary runner must still remain backend-agnostic. It must **always compute both decryption views for every backend** and must not branch on Fast/Standard identity.

Do not silently decode Fast `c0` directly. Use the ordinary public decrypt-and-decode API for both views.

## Existing halted evidence

The first attempt observed at LogN13:

- Standard with generated-secret decode: `max_abs_complex ≈ 5.75e-8` — numerically correct.
- Fast `6f1c957...` with generated-secret decode: large mismatch, including `max_abs_real ≈ 0.18750034` and `max_abs_imag ≈ 0.06266213`.
- Fast `ce79b861...` with generated-secret decode: same class of mismatch.

These Fast generated-secret results are retained as diagnostic evidence only. They must not be interpreted as proving the Fast arithmetic is wrong until the formal zero-secret oracle is evaluated.

No result from the halted attempt should be committed as a completed EXP-002-C artifact unless regenerated under the corrected runner and clean provenance rules below.

## Fixed profiles

Use the exact formal profiles already present:

- `configs/bootstrap_config.logN13.json`
- `configs/bootstrap_config.logN16.json`

They correspond to:

- `ModUpThenEncode`
- full slots
- LogN13: N=8192, 4096 logical complex slots
- LogN16: N=65536, 32768 logical complex slots

Do not change either config.

## Backend commits to validate

Standard:

- `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Fast used by EXP-001-P:

- `6f1c957f95d0dfef6bf73bc01413a7ab55cfb8dc`

Fast used by EXP-002:

- `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Validate both exact Fast commits so correctness evidence maps directly to the completed performance experiments.

Use detached-HEAD switching and safe restore. Do not modify the pinned Lattigo commits.

## Fixed logical input

Use exactly the deterministic full-slot workload already used by `reproducibleInput`:

```text
real = ((i % 7) - 3) / 16
imag = ((i % 5) - 2) / 32
value = real + imag*i
```

Use all logical slots. No sampling, no random replacement workload.

## Backend-agnostic correctness runner

The Primary repository must not import Fast packages, inspect Fast key layouts, branch on backend commit, or add a Fast/Standard execution selector.

For every run, use the same code path:

1. Build parameters from the selected existing config.
2. Generate residual secret `sk` through the ordinary key generator.
3. Generate bootstrap evaluation keys through ordinary `btpParams.GenEvaluationKeys(sk)`.
4. Construct ordinary `bootstrapping.NewEvaluator(...)`.
5. Construct the deterministic full-slot input.
6. Execute exactly one complete `eval.Bootstrap(input)`.
7. Build two decryption keys:
   - `generated_secret`: the original generated `sk`.
   - `zero_secret`: a residual-parameter secret key whose Q/P values are all exactly zero.
8. For the same bootstrap output, run ordinary `rlwe.NewDecryptor(...).DecryptNew(...)` with each key.
9. Decode both plaintexts with ordinary `ckks.NewEncoder(...).Decode(...)`.
10. Archive both full decoded vectors and their error metrics versus the original input.

There must be no backend-specific branch in these steps.

## Oracle interpretation

The raw result must contain both views for every backend:

```text
decoded.generated_secret
decoded.zero_secret
```

and metrics:

```text
errors.generated_secret_vs_input
errors.zero_secret_vs_input
```

For formal pass/fail interpretation:

- Standard correctness = `generated_secret` view.
- Fast correctness = `zero_secret` view.

The opposite view remains archived as diagnostic evidence but does not control pass/fail.

This interpretation belongs in the result-summary logic based on the known experimental backend identity/commit; the actual bootstrap/decrypt execution code must still compute both views unconditionally.

## Formal gate before the full matrix

Before running LogN16, first run the corrected LogN13 gate:

1. Standard `5dbffb...`
2. Fast `6f1c957...`
3. Fast `ce79b861...`

Require:

- Standard generated-secret view passes.
- Fast zero-secret view passes for both Fast commits.
- Fast zero-secret decoded vectors agree with Standard generated-secret decoded vector within the fixed threshold.

If either Fast zero-secret view fails at LogN13, stop and report. That would indicate a real numerical path problem requiring stage-level diagnosis; do not proceed to LogN16 or EXP-003.

If the LogN13 gate passes, run the corresponding three LogN16 cases.

## Formal six-run matrix

1. LogN13 Standard `5dbffb...`
2. LogN13 Fast EXP-001-P `6f1c957...`
3. LogN13 Fast EXP-002 `ce79b861...`
4. LogN16 Standard `5dbffb...`
5. LogN16 Fast EXP-001-P `6f1c957...`
6. LogN16 Fast EXP-002 `ce79b861...`

This is numerical correctness, not timing. One complete bootstrap per backend/profile is sufficient.

Do not publish elapsed time from these runs as performance evidence.

## Clean provenance

Before each formal run:

- Primary is clean and at the same finalized correctness-runner commit.
- Secondary is clean and detached at the exact pinned backend commit.
- write output to `/tmp` first.
- raw metadata records both repositories `dirty=false`.

Only after identity checks pass may artifacts be copied into `results/`.

## Raw artifacts

Archive:

```text
results/EXP-002C-logN13-standard.json
results/EXP-002C-logN13-fast-exp001p.json
results/EXP-002C-logN13-fast-exp002.json
results/EXP-002C-logN16-standard.json
results/EXP-002C-logN16-fast-exp001p.json
results/EXP-002C-logN16-fast-exp002.json
```

Each raw result must contain at least:

- schema version and timestamp
- Primary/Lattigo commit/ref/dirty metadata
- environment, config, effective parameters
- deterministic workload identifier/formula
- logical slot count
- output level/scale/domain metadata
- complete full-slot decoded vector for `generated_secret`
- complete full-slot decoded vector for `zero_secret`
- error metrics for both views versus input

Do not truncate LogN16 outputs.

## Numerical metrics

For complex vectors `a` and `b`, calculate at least:

```text
max_abs_complex = max_i |a_i - b_i|
mean_abs_complex = mean_i |a_i - b_i|
rmse_complex = sqrt(mean_i |a_i - b_i|^2)
max_abs_real = max_i |Re(a_i) - Re(b_i)|
max_abs_imag = max_i |Im(a_i) - Im(b_i)|
max_component_abs = max(max_abs_real, max_abs_imag)
```

Record the slot indices producing maximum complex/real/imag errors.

Optionally record:

```text
precision_bits = -log2(max_abs_complex)
```

Use a JSON-safe representation when the error is exactly zero.

## Required formal comparisons

For each profile:

### Standard correctness

Compare:

```text
Standard decoded.generated_secret vs original input
```

### Fast correctness

Compare:

```text
Fast EXP-001-P decoded.zero_secret vs original input
Fast EXP-002   decoded.zero_secret vs original input
```

### Fast vs Standard logical equivalence

Compare slot-by-slot:

```text
Fast EXP-001-P decoded.zero_secret
    vs
Standard decoded.generated_secret

Fast EXP-002 decoded.zero_secret
    vs
Standard decoded.generated_secret
```

These pairwise comparisons must be computed from archived decoded vectors, not inferred from their separate input-error summaries.

Also retain the non-authoritative opposite-key metrics:

```text
Standard decoded.zero_secret vs input
Fast decoded.generated_secret vs input
```

as diagnostic evidence of the differing secret semantics.

## Acceptance threshold

Use the existing implementation-level numerical contract:

```text
max_abs_real <= 1e-2
max_abs_imag <= 1e-2
```

Require it for:

- Standard generated-secret vs input
- each Fast zero-secret vs input
- each Fast zero-secret vs Standard generated-secret

Do not weaken the threshold after observing results.

If a meaningful comparison fails, archive the failure evidence, stop, and report. Do not proceed to EXP-003.

## Summaries

Create:

```text
results/EXP-002C-logN13-summary.json
results/EXP-002C-logN16-summary.json
results/EXP-002C-matrix-summary.json
```

Each summary must include:

- exact commits
- identity/clean-state checks
- slot count
- both decryption views and which one is authoritative for that backend
- all authoritative input-reference metrics
- Fast-vs-Standard metrics
- diagnostic opposite-key metrics
- fixed threshold
- per-comparison pass/fail
- overall profile pass/fail
- raw artifact paths

The matrix summary must make explicit that Standard and Fast use different secret semantics in the current intentionally insecure Fast architecture.

## Implementation constraints

- Keep Primary bootstrap execution backend-agnostic.
- Always compute both decryption views; no `if Fast` execution branch.
- Do not import `schemes/ckks/fast` into Primary.
- Do not alter timing paths or any EXP-001/EXP-002 timing artifact.
- Do not change configs or bootstrapping parameters.
- Do not modify Lattigo for this corrected oracle experiment.
- Do not start noise-fidelity research.

## Validation

After the corrected Primary runner is finalized, run:

```text
go test ./...
```

against all three pinned backend commits.

Add focused Primary tests for dual-decryption metric/serialization logic if required.

## Completion criteria

EXP-002-C passes only when:

1. one finalized clean Primary runner produced all six formal runs;
2. LogN13 validates all 4096 slots;
3. LogN16 validates all 32768 slots;
4. Standard generated-secret output passes vs input in both profiles;
5. Fast `6f1c957...` zero-secret output passes vs input and vs Standard in both profiles;
6. Fast `ce79b861...` zero-secret output passes vs input and vs Standard in both profiles;
7. actual errors/outlier indices and both secret views are archived;
8. previous timing evidence remains unchanged;
9. both repositories are restored safely;
10. `CURRENT_TASK.md` is marked `COMPLETE` only after results are committed.

Only after independent review marks EXP-002-C PASS may the project advance to EXP-003.

## Stop condition

Stop after EXP-002-C completion. Do not start EXP-003 automatically.