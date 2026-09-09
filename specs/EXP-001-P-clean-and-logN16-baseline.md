# EXP-001-P — Clean Baseline Matrix + N=65536

## Goal

Turn EXP-001 into a clean, reproducible coarse-grained baseline matrix and formally add the N=65536 case required by the architecture work.

This task must preserve all existing EXP-001 raw results. Do not overwrite or delete them.

The task answers two coarse-grained questions only:

1. Can the Standard/Fast full-bootstrap comparison be reproduced with both repositories clean (`dirty=false`)?
2. Under the same current compatibility workload, what is the full-bootstrap performance at both LogN=13 and LogN=16 (N=65536)?

Do not add stage timing or primitive timing yet. Those belong to later tasks.

## Fixed backend commits

Use the same pinned backends as EXP-001:

Standard:
- `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Fast:
- `d5429b1c368a6828ef1bb8a229e3f57458ad6491`

Use detached-HEAD switching and restore the original secondary branch afterward, following the safe procedure in EXP-001 and `AGENTS.md`.

## Preserve EXP-001

Do not modify or overwrite:

- `results/EXP-001-standard.json`
- `results/EXP-001-fast.json`
- `results/EXP-001-summary.json`

They are historical first-pass baseline data and must remain available.

## Clean-provenance requirement

Before every formal measurement:

- Primary must be on synchronized `main`.
- Primary `git status --short` must be empty.
- Secondary `git status --short` must be empty.
- The secondary HEAD must equal the pinned backend commit being measured.

If either repository is dirty, do not record a formal baseline.

To avoid the output files themselves making Primary dirty during measurement, write raw experiment output to a temporary location outside the repository first, e.g. `/tmp/...`. Only after measurement is complete and the raw JSON records `primary_repository.dirty=false` and `lattigo_repository.dirty=false` may it be copied into `results/` for archival and committed later.

Do not change the harness to lie about Git status. Fix the experiment procedure, not the metadata.

## Configuration profiles

Create and keep two explicit active config files.

### Profile A — LogN=13

Recommended path:

`configs/bootstrap_config.logN13.json`

Use the current compatibility baseline values already represented by the active example config, with:

- `log_n = 13`
- `log_slots = -1`

For Standard CKKS this means N=8192 and maximum slots = 4096.

### Profile B — LogN=16 / N=65536

Create:

`configs/bootstrap_config.logN16.json`

Start from the same active current compatibility profile and change only:

- `log_n = 16`

Keep:

- `log_slots = -1`
- `log_default_scale = 45`
- `secret_hamming = 192`
- `q0 = [55]`
- `q_slots_to_coeffs = [39,39,39]`
- `q_coeffs_to_slots = [56,56,56,56]`
- `p = [61,61,61,61,61]`
- `slots_to_coeffs_dft_levels = [1,1,1]`
- `coeffs_to_slots_dft_levels = [1,1,1,1]`
- `mod1_log_scale = 60`
- `mod1_degree = 30`
- `mod1_double_angle = 3`
- `mod1_k = 16`
- `log_message_ratio = 10`
- `mod1_inv_degree = 0`

With a Standard ring and `log_slots=-1`, this profile is intended to exercise N=65536 and 32768 logical slots.

Do not silently reduce LogSlots for the N=65536 formal profile.

## Historical N=65536 reference

Create a short documentation file, for example:

`docs/LEGACY_N65536_REFERENCE.md`

Record the historical `legacy-hardware-model` test contract accurately:

`TestSlimBootstrapperHWN65536` performed:

```go
cfg := DefaultBootstrapConfig()
cfg.LogN = 16
```

and otherwise inherited the old default values:

- LogDefaultScale 45
- SecretHamming 192
- Q0 [55]
- QSlotsToCoeffs [39,39,39]
- QCircuitSlots [45]
- QEvalMod [60,60,60,60,60,60,60,60]
- QCoeffsToSlots [56,56,56,56]
- P [61,61,61,61,61]
- SlotsToCoeffsDFT [1,1,1]
- CoeffsToSlotsDFT [1,1,1,1]
- LogBSGSRatio 1
- LogSlots -1
- Mod1LogScale 60
- Mod1Degree 30
- Mod1DoubleAngle 3
- Mod1K 16
- LogMessageRatio 10
- Mod1InvDegree 0

Also document the important incompatibility:

The modern experiment profile intentionally preserves many high-level numerical settings, but it is **not** construction-identical to the legacy builder. The legacy path used a full residual/full Q chain and `DecodeThenModUp`; the current Standard/Fast compatibility path uses a two-prime residual and `ModUpThenEncode`, with the bootstrapping Q chain derived by the current public Lattigo parameter builder.

Do not claim that the new LogN16 profile exactly reproduces the old hardware model.

## Formal measurement matrix

Run four formal full-bootstrap experiments:

1. LogN13 Standard
2. LogN13 Fast
3. LogN16 Standard
4. LogN16 Fast

For every run use:

- `repetitions = 10`
- `warmup = 3`
- the same measurement boundary as EXP-001
- complete `eval.Bootstrap(ct)` only inside the timed section

Use the same command structure for both backends; only config path and output filename may vary.

Example:

```text
go run . -config <profile>.json -repetitions=10 -warmup=3 -out /tmp/<raw-result>.json
```

## Archive filenames

After validating clean metadata, archive without overwriting previous EXP-001 data:

```text
results/EXP-001P-logN13-standard.json
results/EXP-001P-logN13-fast.json
results/EXP-001P-logN13-summary.json
results/EXP-001P-logN16-standard.json
results/EXP-001P-logN16-fast.json
results/EXP-001P-logN16-summary.json
```

Also create:

`results/EXP-001P-matrix-summary.json`

The matrix summary should contain for each LogN:

- N
- LogSlots / logical slot count
- primary commit
- Standard backend commit
- Fast backend commit
- effective-parameter identity result
- clean-state checks
- Standard median/mean/min/max elapsed time
- Fast median/mean/min/max elapsed time
- median speedup
- median allocation bytes
- median allocation count
- environment metadata

## Identity checks

For each Standard/Fast pair, require before reporting speedup:

- Primary commit identical
- Primary dirty=false on both runs
- Secondary dirty=false on both runs
- config identical between the pair
- effective parameters identical between the pair
- environment identical
- repetitions/warmup identical
- expected N matches profile
- expected LogSlots matches profile
- all output levels expected
- backend commits equal the pinned Standard/Fast SHAs

If any check fails, do not publish a speedup for that pair.

## Validation

Before formal runs:

```text
go test ./...
```

must succeed against both pinned backends as appropriate.

No Lattigo source modification is expected.

## Relationship to existing Lattigo LogN16 benchmark

The Fast Lattigo repository already contains a complete-bootstrap benchmark that includes `LogN=16`, but that benchmark uses `LogSlots=4` and a speed-oriented profile. It is historical engineering evidence, but it is not a substitute for this formal N=65536 / `log_slots=-1` experiment profile.

Do not overwrite or reinterpret the old benchmark results.

## Future coarse-to-fine roadmap

After EXP-001-P passes:

### EXP-002 — Bootstrap stage breakdown

Measure Standard/Fast at LogN13 and LogN16 for major stages, preserving raw data separately:

- ScaleDown
- ModUp / Trace as applicable
- CoeffsToSlots
- EvalMod
- SlotsToCoeffs
- packing/finalization where applicable
- full Bootstrap as the parent total

### EXP-003 — CKKS primitive breakdown

Measure Standard/Fast at relevant LogN values for reusable primitives, including at least:

- Add / Sub
- ciphertext × ciphertext Mul
- MulRelin
- Relinearize
- Rescale / RescaleTo
- scalar/plaintext multiplication
- Automorphism / rotation
- linear transform
- NTT / INTT or other ring primitives where useful

Existing Fast unit tests should be reused as correctness evidence where appropriate, but experiment results must be archived in the Primary harness rather than left only as transient benchmark console output.

## Stop condition

EXP-001-P passes when:

- previous EXP-001 data remains untouched,
- clean LogN13 Standard/Fast raw results are archived,
- clean full-slot LogN16/N65536 Standard/Fast raw results are archived,
- pairwise identity checks pass,
- summaries are stored,
- the historical N65536 parameter relationship is documented without claiming false equivalence,
- Secondary is restored cleanly to its original branch/ref.

Do not automatically begin stage or primitive profiling after this task.