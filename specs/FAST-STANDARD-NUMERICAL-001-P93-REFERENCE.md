# FAST-STANDARD-NUMERICAL-001 — P93 Fast vs Standard Numerical Reference

## Status

Executable correctness / diagnostic-framework task.

## Task class

`C — Numerical Correctness Validation`

This task intentionally pauses performance optimization.

## Accepted production state

Primary latest accepted result:
- `results/QPREFIX-PERF-OPT-004-summary.md`
- Primary result commit: `57e92e9bc35fd4d576d23cca12cb6d70c0428f06`
- OPT-004 classification: `BARRETT_FIXED_WIDTH_REDUCTION_CORRECT_BUT_NO_WIN`

Secondary:
- active branch `fast-qprefix`
- current control-plane HEAD `e3a7e04dc90f06128268510bcd31b08c2fbca176`
- accepted production arithmetic remains `6930cf6cb3c71ce139a1eb42eede7be335b7174c`

No production code changed in OPT-003 or OPT-004.

## Motivation

The current reusable fastdiag framework checks numerical stability only as:

```text
Fast diagnostics-off
vs
Fast diagnostics-on
```

and `fastdiag compare` compares timing/event structure between two Fast git revisions.

The repository separately contains the small correctness test:

`TestFastBootstrapMatchesStandardDecodedReference`

which compares decoded Fast vs Standard values with a fixed `1e-2` tolerance.

What is missing is an authoritative numerical comparison on the actual accepted:

- P93;
- q0=55;
- LogN=13;
- LogSlots=12;
- 4096-slot;
- Count-1 fastdiag workload.

## Goal

Measure, report, and make reusable the numerical-quality relationship among:

1. original encoded message;
2. current Fast Bootstrap output;
3. Standard Bootstrap output.

The result must answer quantitatively:

- How accurate is Fast versus the original message?
- How accurate is Standard versus the original message?
- How far is Fast from Standard?
- Does Fast materially lose numerical precision relative to Standard?
- Is the existing `1e-2` pass criterion hiding a much larger quality gap?

No performance optimization is allowed.

## 1. Hard constraints

Do not modify:

- Fast or Standard production arithmetic;
- Q-prefix policy;
- parameters;
- bootstrap polynomial schedule;
- P93 schedule;
- key semantics;
- public production APIs.

Allowed changes:
- reusable diagnostic/test code;
- Primary fastdiag command/framework code;
- Secondary test-only numerical fixture code.

Do not mix this task with performance benchmarking.

## 2. Canonical P93 input

Reuse the exact existing `fastDiagP93Fixture` message vector and fingerprint:

`d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`

The canonical workload contains 4096 complex slots and must keep the existing:
- LogN=13;
- LogSlots=12;
- q0=55;
- polynomial degree;
- double-angle;
- Q/P chain;
- message values.

Fail if the fingerprint changes.

## 3. Apples-to-apples Fast and Standard execution

Use the same canonical message values and the same deterministic plaintext-like level-0 ciphertext construction already used by the Fast tests:

```text
c0 = encoded plaintext
c1 = 0
```

This ensures both evaluators receive mathematically the same input message/ciphertext coefficients.

### Fast

- construct `NewFastEvaluator(params)`;
- run `Fast Bootstrap`;
- decode the Fast output in the same manner as current Fast tests.

### Standard

For each Standard trial:
- generate a fresh Standard secret key from the same bootstrapping parameters;
- generate the matching Standard evaluation keys;
- construct `NewEvaluator(params, keys)`;
- feed the same canonical plaintext-like ciphertext;
- run Standard `Bootstrap`;
- decrypt with that trial's secret key;
- decode with the same residual CKKS parameters.

Do not compare raw ciphertext coefficients between Fast and Standard; the representations/key/noise are not expected to be identical.

## 4. Standard trial count

Run at least **3 independent Standard key/evaluation-key trials** on the same canonical message.

Reason:
Standard evaluation-key noise is key/randomness dependent; a single trial is not enough to distinguish a stable Fast-vs-Standard offset from one Standard-noise realization.

Record each trial separately.

Fast may be run once if repeated Fast executions are deterministic. Verify this with at least two Fast executions; if they differ numerically, record multiple Fast trials too.

## 5. Metadata comparison

For each Fast/Standard output report:

- Level;
- Degree;
- Scale;
- IsNTT;
- IsMontgomery;
- LogDimensions.

Classify metadata differences as:
- expected representation/key-path difference;
- or semantic mismatch.

Do not require raw metadata equality if Standard legitimately retains a different ciphertext degree/representation, but Level and Scale must be checked against the public Bootstrap contract.

## 6. Numerical metrics

For decoded vectors (x) (original), (f) (Fast), and (s_j) (Standard trial j), compute all metrics over **all 4096 slots**.

### Fast vs original

For real, imaginary, and complex magnitude error:
- mean absolute error;
- RMSE;
- p50 absolute error;
- p95;
- p99;
- maximum;
- worst slot index and values.

### Standard vs original

Same metrics for every Standard trial, plus median across Standard trials.

### Fast vs Standard

For every Standard trial, compute:

[
d_i = f_i - s_{j,i}
]

and report:
- real MAE/RMSE/p95/p99/max;
- imaginary MAE/RMSE/p95/p99/max;
- complex (|d_i|) MAE/RMSE/p95/p99/max;
- worst slot index and `original / Fast / Standard / difference`.

### Precision bits

For each output compute an error-based precision statistic:

[
p_i = -log_2(max(|y_i-x_i|,epsilon))
]

with a documented sufficiently small numerical floor (epsilon) only to avoid (log(0)).

Report:
- mean;
- median;
- p05;
- minimum slot precision.

Also use existing CKKS precision helpers such as `GetPrecisionStats` where applicable, but do not substitute them for the explicit metrics above.

## 7. Standard-to-Standard variability

Across the >=3 Standard trials, measure decoded spread:

- per-slot Standard trial mean;
- Standard trial standard deviation or RMS spread;
- maximum pairwise Standard-vs-Standard decoded difference.

This establishes how much of Fast-vs-Standard separation could simply be ordinary Standard evaluation-key noise.

## 8. Existing 1e-2 criterion audit

Explicitly report:

- number and fraction of Fast-vs-original real/imag coordinates exceeding `1e-2`;
- Standard-vs-original exceeding `1e-2`;
- Fast-vs-Standard exceeding `1e-2`.

Do **not** stop after this threshold check.

The report must answer whether the actual errors are:
- orders of magnitude below `1e-2`;
- near the threshold;
- or exceeding it.

## 9. Classification

Return exactly one:

- `FAST_STANDARD_NUMERICAL_CLOSE`
- `FAST_NUMERICAL_QUALITY_DEGRADED`
- `FAST_STANDARD_NUMERICAL_UNCLOSED`

### CLOSE

Use only if all are true:

1. no Fast-vs-original real/imag coordinate exceeds the existing `1e-2` correctness bound;
2. no Fast-vs-Standard real/imag coordinate exceeds `1e-2`;
3. Fast median precision is no more than **2 bits worse** than the median Standard-trial precision;
4. Fast complex RMSE is no more than **4x** the median Standard complex RMSE, unless Standard RMSE is so close to zero that the ratio is numerically unstable—in that case use the precision-bit criterion and absolute errors;
5. metadata/public Bootstrap contract is respected.

The 2-bit / 4x conditions are the same degradation scale expressed in logarithmic/linear form; report both.

### DEGRADED

Use if the measurements are reliable but any CLOSE quality condition materially fails.

### UNCLOSED

Use if Standard P93 cannot execute under truly comparable parameters/input, decode is invalid, key semantics make comparison non-comparable, or the measurement framework cannot establish reliable metrics.

Do not weaken thresholds after seeing the data.

## 10. Reusable fastdiag support

Extend the reusable framework with a numerical-reference mode, preferred CLI:

```sh
./scripts/fastdiag numerical \
  --profile p93-q55 \
  --standard-trials 3
```

Exact flag spelling may follow existing parser conventions.

The generated JSON/Markdown must contain:
- canonical workload fingerprint;
- Fast repository commit;
- environment;
- Fast numerical metrics;
- every Standard trial's metrics;
- Standard aggregate metrics;
- Fast-vs-Standard metrics;
- Standard-to-Standard variability;
- classification.

This mode is **not a timing benchmark**.

Do not change existing `trace` or `compare` schemas/behavior except for shared refactoring that preserves compatibility.

If adding this mode cleanly would require disproportionate framework churn, implement the numerical test/artifact first and report the CLI framework limitation; do not block the correctness measurement itself.

## 11. Permanent regression test

Add a bounded permanent P93 numerical regression test if runtime is acceptable for ordinary Secondary `go test ./...`.

If generating three Standard key sets is too expensive for the ordinary suite:
- keep a smaller permanent Fast-vs-Standard smoke test;
- keep the full 3-trial P93 comparison behind an explicit diagnostic test/command rather than making ordinary CI excessively slow.

Do not silently omit the full P93 measurement.

## 12. Required result artifact

Write:

`results/FAST-STANDARD-NUMERICAL-001-summary.md`

Include:

- exact Primary/Secondary provenance;
- canonical fingerprint;
- parameter table;
- execution/decode method;
- Fast repeat determinism result;
- Standard trial key count;
- full metric tables;
- worst-slot examples;
- threshold audit;
- precision-bit comparison;
- Standard-to-Standard variability;
- classification;
- limitations.

Include machine-readable JSON artifact if the framework supports it.

## 13. Validation

Run:
- new focused numerical tests;
- `go test ./cmd/fastdiag` in Primary if framework changes;
- relevant Secondary Fast/Standard bootstrap tests;
- Secondary `go test ./...` if practical;
- `git diff --check`.

No production arithmetic changes are permitted.

## 14. Completion

Return:

`FAST_STANDARD_NUMERICAL_REFERENCE_READY`

plus:
- classification;
- `FAST_COMPLEX_RMSE=...`
- `STANDARD_COMPLEX_RMSE_MEDIAN=...`
- `FAST_STANDARD_COMPLEX_RMSE_MEDIAN=...`
- `FAST_MEDIAN_PRECISION_BITS=...`
- `STANDARD_MEDIAN_PRECISION_BITS=...`
- `FAST_STANDARD_MAX_COMPLEX_DIFF=...`
- `FAST_VS_STANDARD_OVER_1E2_COUNT=...`

Then:

`READY_FOR_WEB_REVIEW`.
