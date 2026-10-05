# FAST-STANDARD-NUMERICAL-REPORT-001 — Close-state classifier

## Status

Executable reporting-only repair.

## Task class

`R — Diagnostic / Reporting Repair`

## Accepted parent result

Accepted numerical repair result:

- Secondary production commit: `5117fc57949647182f476dc5952099c706b9f869`
- Primary result commit: `88e6e6badc0973b64bb051a0ce7fc51cf3f95e75`
- top-level classification: `FAST_STANDARD_NUMERICAL_CLOSE`
- final Fast-vs-Standard complex RMSE: `1.44680895443e-10`
- max Fast-vs-Standard complex difference: `4.675261e-10`
- Standard Bootstrap SNR: `144.713390 dB`
- Fast Bootstrap SNR: `144.710750 dB`
- no coordinate above `1e-2`
- no first observable checkpoint at threshold `1e-8`
- no first material checkpoint
- all 56 Q-prefix capacity checks passed

These numerical results are accepted. Do not rerun or modify Secondary production arithmetic merely to address this reporting task.

## Problem

`summarizeNumericalStageLockstep()` initializes:

`Classification = "CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED"`

and only replaces it when a first material checkpoint is found.

Therefore a healthy run with:
- no first observable checkpoint;
- no first material checkpoint;
- a comparable final public output well below the observable threshold;

incorrectly retains the legacy divergence label.

This is a reporting state-machine bug.

## Required behavior

Keep existing causal classifications when a first material checkpoint exists.

Add an explicit healthy/close stage-lockstep terminal classification:

`CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`

Use it only when all of the following are true:

1. `FirstObservable == ""`;
2. `FirstMaterial == ""`;
3. `final_public_output` exists and is comparable;
4. final Fast-vs-Standard metrics exist and are finite;
5. final max complex difference is strictly below `numericalObservableThreshold`.

Do not infer healthy status merely because `FirstObservable` is empty when the final output is unavailable or non-comparable.

If no material checkpoint exists but an observable checkpoint does exist, do not return the healthy label. Preserve an explicit non-healthy classification, e.g.:
`CURRENT_FAST_OBSERVABLE_NON_MATERIAL_DIVERGENCE`.

If the diagnostic cannot establish either a valid healthy terminal state or an existing causal classification, retain:
`CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED`.

## Required tests

Add focused unit tests for at least:

1. no observable divergence + comparable final output below threshold =>
   `CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`;
2. observable but non-material divergence =>
   `CURRENT_FAST_OBSERVABLE_NON_MATERIAL_DIVERGENCE`;
3. first material EvalMod remains the existing EvalMod classification;
4. missing/non-comparable final output does not get healthy classification;
5. current accepted canonical result renders without contradictory stage classification.

Run:
- `go test ./cmd/fastdiag ./internal/numericalmetrics`
- `git diff --check`

## Scope

Primary reporting/diagnostic code only.

Do not:
- modify Secondary;
- change numerical thresholds;
- modify production arithmetic;
- change the accepted numerical measurements;
- weaken classification gates.

## Result

Regenerate or patch the canonical numerical result/report only as necessary so that its stage-lockstep classification is consistent with the accepted measurement.

The expected stage-lockstep classification for the accepted result is:

`CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`

The top-level classification remains:

`FAST_STANDARD_NUMERICAL_CLOSE`.

## Completion

Return:

`FAST_STANDARD_NUMERICAL_REPORT_FIX_READY`

then:
- stage-lockstep classification;
- top-level classification;
- confirmation that numerical values are unchanged.

Then:

`READY_FOR_WEB_REVIEW`.
