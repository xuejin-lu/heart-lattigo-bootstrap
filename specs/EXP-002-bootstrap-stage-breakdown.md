# EXP-002 — Bootstrap Stage Breakdown

## Goal

Explain where the end-to-end Standard-vs-Fast bootstrap speedup comes from by measuring the major bootstrapping stages separately at both LogN=13 and LogN=16.

This task is stage-level profiling only. Do not begin CKKS primitive profiling yet; that belongs to EXP-003.

## Fixed experiment profiles

Reuse the existing clean profiles from EXP-001-P:

- `configs/bootstrap_config.logN13.json`
- `configs/bootstrap_config.logN16.json`

Use:

- LogN13: N=8192, log_slots=12, 4096 logical slots
- LogN16: N=65536, log_slots=15, 32768 logical slots
- warmup=3
- repetitions=10

Do not change the profile between Standard and Fast.

## Backend identities

Standard baseline:

- `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Fast baseline at task start:

- `6f1c957f95d0dfef6bf73bc01413a7ab55cfb8dc`

If a Fast compatibility-only change is required to expose the same public stage API, make the smallest change on `fast-ckks`, push it, and use that new exact Fast commit for every Fast measurement in this task.

Do not modify Standard behavior.

## Architecture rule

The Primary harness must remain backend-agnostic.

It may call only the ordinary public `circuits/ckks/bootstrapping.Evaluator` API. Do not import `schemes/ckks/fast`, do not add `--fast`, do not inspect q0/q1 internals, and do not branch on backend identity.

If the ordinary stage methods do not currently delegate to Fast when the evaluator was constructed through the Fast-compatible standard API, fix that compatibility boundary in the Fast Lattigo branch rather than adding backend-specific code to Primary.

Before editing, inspect the exact current signatures and behavior of:

- `ScaleDown`
- `ModUp`
- `CoeffsToSlots`
- `EvalMod`
- `SlotsToCoeffs`
- `PackAndSwitchN1ToN2`
- `UnpackAndSwitchN2ToN1`
- `Bootstrap`

Do not invent a parallel profiling API if the existing public methods can be safely delegated.

## Stage pipeline

For the current `ModUpThenEncode` workload, profile the real pipeline in execution order:

1. `PackAndSwitchN1ToN2`
2. `ScaleDown`
3. `ModUp` (including Trace where the implementation performs it)
4. `CoeffsToSlots`
5. `EvalMod` on the real branch
6. `EvalMod` on the imaginary branch when non-nil
7. `SlotsToCoeffs`
8. `UnpackAndSwitchN2ToN1`
9. complete `Bootstrap` as an independent parent total

Also report useful aggregate labels:

- `eval_mod_total = eval_mod_real + eval_mod_imag`
- `stage_sum = sum of separately timed pipeline stages`
- `unattributed_overhead = full_bootstrap_median - stage_sum_median` only as a diagnostic; do not assume medians are perfectly additive.

For the current full-slot Standard-ring profiles, explicitly record whether packing, Trace, or ring switching is effectively active or trivial. Do not silently omit a stage just because it is cheap.

## Measurement method

The stage measurements must use the actual sequential outputs of the preceding stage, not independently fabricated stage inputs.

For each measured repetition:

1. Start from the same deterministic base ciphertext construction used by the existing harness.
2. Keep input-copy/setup work outside stage timers.
3. Execute the pipeline in order.
4. Time each public stage call around only that call.
5. Record allocation bytes and allocation count for each stage using the same methodology for Standard and Fast.
6. Verify the stage output metadata needed by the next stage before proceeding.
7. Run a separate complete `Bootstrap` measurement rather than using the sum of stages as the full-bootstrap result.

Warm-up must execute the same stage pipeline before recorded repetitions.

Do not include parameter generation, key generation, evaluator construction, config loading, JSON serialization, or output-file writes in stage timing.

## Mutation / fairness rule

Many bootstrap stage methods mutate ciphertexts. Ensure Standard and Fast both receive equivalent fresh starting state for every recorded repetition.

Do not reuse an already-mutated ciphertext across repetitions.

The timing harness itself must be identical across backends.

## Correctness checks

Before accepting timing data, verify for every profile/backend:

- all stages complete without error;
- stage levels/scales/domains are consistent with the next stage contract;
- complete pipeline output level matches the existing full-bootstrap baseline;
- the final staged pipeline result and ordinary complete `Bootstrap` result agree to the extent already supported by the current experiment correctness contract;
- Standard and Fast use identical config/effective parameters;
- Primary commit is identical within each Standard/Fast pair;
- machine/Go/environment match;
- both repositories record `dirty=false` during formal measurement.

This task is about speed attribution, not new noise-fidelity analysis. Do not broaden the correctness oracle beyond what is needed to show the staged path is the same workload as the existing full bootstrap.

## Results

Preserve all EXP-001 / EXP-001-P / EXP-001P2 files.

Archive new raw data under distinct names, for example:

```text
results/EXP-002-logN13-standard.json
results/EXP-002-logN13-fast.json
results/EXP-002-logN16-standard.json
results/EXP-002-logN16-fast.json
results/EXP-002-logN13-summary.json
results/EXP-002-logN16-summary.json
results/EXP-002-matrix-summary.json
```

Do not overwrite earlier baselines.

Each raw result must include per repetition and per stage:

- elapsed_ns
- alloc_bytes
- allocs
- input level
- output level
- input scale where practical
- output scale where practical
- whether the stage produced one or two ciphertext branches when applicable

Also include provenance:

- Primary commit
- Lattigo commit/ref
- dirty flags
- config
- effective parameters
- Go/OS/arch/CPU
- warmup/repetitions
- timestamp

## Summary statistics

For every stage and backend report at least:

- median elapsed time
- mean elapsed time
- min/max elapsed time
- median allocation bytes
- median allocation count

For every comparable stage report:

```text
stage_speedup = Standard median / Fast median
```

For each LogN also report:

- full Bootstrap median Standard/Fast
- full Bootstrap speedup
- percentage of Standard full-bootstrap time attributed to each Standard stage
- percentage of Fast full-bootstrap time attributed to each Fast stage
- top 3 stages by absolute time in Standard
- top 3 stages by absolute time in Fast
- top 3 contributors to absolute time saved (`Standard median - Fast median`)

Do not infer causality beyond measured stage timing.

## Expected scientific question

The final report should make it easy to answer:

> At N=8192 and N=65536, which bootstrap stages account for most of the total runtime, and which stages account for most of Fast's end-to-end speedup?

## Tests

Primary:

```text
go test ./...
```

Standard and Fast backend checkouts must both support the same Primary stage runner source.

If Secondary changes are required for Fast stage delegation, run the relevant Fast packages at minimum:

```text
go test ./schemes/ckks/fast
go test ./circuits/ckks/bootstrapping
go test ./circuits/ckks/dft
go test ./circuits/ckks/mod1
go test ./circuits/ckks/polynomial
```

## Stop condition

EXP-002 passes when clean, reproducible stage-level Standard/Fast measurements for both LogN13 and LogN16 are archived, the same Primary stage runner works against both backends, the complete-bootstrap totals remain consistent with the EXP-001P2 order of magnitude, and the major sources of time and time savings can be identified from stored data.

Do not automatically begin EXP-003 primitive profiling after completion.