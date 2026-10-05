# FAST-STANDARD-NUMERICAL-DIAG-003 — Normalized EvalMod Semantic Alignment

## Status

Executable numerical diagnosis task.

## Task class

`D — Numerical Correctness Diagnosis`

## Accepted parent

Accepted clean reproduction:
- `results/FAST-STANDARD-NUMERICAL-DIAG-002-summary.md`
- measurement Primary: `4fd11a5aa557a9e91c0473c94f76b9bfc41a9e36` (clean)
- result commit: `c379930340dcfcdcfafa4437c5bcbd4356c6c036`
- Secondary: `97c1c6174e0d7ef781de6d8dadce5c2869a53496` (clean)

Accepted outer-stage facts on the canonical LogN13 workload:
- Standard Bootstrap SNR: `144.71339028083412 dB`
- Fast Bootstrap SNR: `-0.17249085395397162 dB`
- final Fast-vs-Standard complex RMSE: `0.02089221106956351`
- C2S real/imag remain near Standard (~2.4e-14 RMSE)
- first material outer stage remains `evalmod_real`
- combined-branch S2C amplification: `2.8284271247428943`

These outer-stage results remain accepted.

## Web-review finding that supersedes the prior internal classification

The prior internal classification
`CURRENT_FAST_FIRST_MATERIAL_EVALMOD_NORMALIZATION`
is **not accepted as a causal localization yet**.

Reason:

The current Fast normalized LogN13 EvalMod intentionally changes its internal semantic representation by a power-of-two normalization.

In Secondary `circuits/ckks/mod1/fast.go`, the normalized path:

1. produces a polynomial result with a working scale;
2. computes `kIn`;
3. changes metadata from the working scale to a coherent scale without multiplying ciphertext coefficients;
4. tracks `currentExponent = kIn`;
5. implements DoubleAngle using exponent-aware multipliers and constants;
6. finally multiplies by `2^currentExponent` to restore the normalized representation.

Therefore, after the coherent-scale metadata transition, raw decoded Fast values are intentionally smaller than the corresponding Standard semantic values by a known power-of-two factor.

A direct comparison

[
\mathrm{Decode}(F_i) \quad \text{vs} \quad \mathrm{Decode}(S_i)
]

at those internal normalized checkpoints is not a common-semantic comparison.

The previous internal trace compared those raw decoded values directly, so the apparent jump to approximately 0 dB stage-reference SNR at the coherent-scale transition can be a representation artifact rather than numerical corruption.

The stage-level conclusion "the material error originates somewhere inside EvalMod" remains valid. The internal sub-stage cause must be re-established with aligned semantics.

## Goal

Re-run the EvalMod internal bisect with mathematically aligned Fast and Standard semantic values.

Answer:

1. After correcting for the intentional Fast power-of-two normalized representation, where is the first observable divergence inside EvalMod?
2. Where is the first material divergence?
3. Does the polynomial output remain only a small error, or does the aligned DoubleAngle recurrence materially amplify it?
4. Is the actual problem:
   - polynomial/generated-power arithmetic;
   - normalized exponent schedule;
   - DoubleAngle arithmetic;
   - final normalized restore;
   - public scale reset;
   - or still unclosed?
5. Only after this aligned diagnosis, what is the single next causal counterfactual?

Do not execute a repair/counterfactual in this task.

## Canonical workload

Reuse the exact accepted DIAG-002 workload:
- fingerprint `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`;
- LogN=13;
- LogSlots=12;
- 4096 complex slots;
- q0=55;
- unchanged Q/P chain;
- Mod1 degree 30;
- DoubleAngle 3;
- EvalMod log scale 60;
- LogMessageRatio 10.

Fail if the fingerprint changes.

## Frozen production state

Secondary production arithmetic is frozen at:

`97c1c6174e0d7ef781de6d8dadce5c2869a53496`

No production arithmetic change is authorized.

Primary diagnostic code may be changed only to correct semantic alignment and reporting.

## Required semantic-alignment model

For every internal EvalMod checkpoint, distinguish:

- **raw decoded value**: what ordinary CKKS decoding returns from the current metadata;
- **representation exponent**: the power-of-two normalization carried by Fast at that checkpoint;
- **aligned semantic value**: the Fast decoded value mapped back to the same mathematical quantity represented by Standard.

Do not compare raw decoded values when the two implementations intentionally represent different normalized quantities.

### Source-derived alignment

Derive the alignment factor from the actual production equations in `evaluateNormalizedLogN13`; do not hard-code an unexplained correction merely to improve agreement.

At minimum, document and test the expected alignment through:

1. polynomial output / before coherent transition;
2. after coherent-scale transition;
3. before each DoubleAngle round;
4. immediately after each raw square/multiply;
5. after exponent multiplier and constant;
6. after each Rescale;
7. before final normalized restore;
8. after final normalized restore;
9. before public scale reset;
10. after public scale reset.

For q0=55 the current source observes:
- plan bits = 91;
- working exponent = 31;
- initial `kIn = 29`;
- DoubleAngle multiplier exponent = 30.

These are observed production facts, not presumed correct values.

### Alignment equations

Let a Fast checkpoint decode to (f_i), Standard to (s_i), and let (r_i) denote the source-derived Fast representation exponent at that exact checkpoint.

Where the Fast representation means the same mathematical value divided by (2^{r_i}), compare:

[
\tilde f_i = 2^{r_i} f_i
]

against (s_i).

For checkpoints immediately after a raw square and before the source's exponent-restoring multiplier, the alignment exponent may be (2r_i) rather than (r_i).

Do not assume one exponent for an entire round. Derive the exact exponent for each checkpoint from the source algebra.

If a checkpoint cannot be assigned an unambiguous source-derived semantic mapping, mark it `NOT_COMPARABLE` rather than forcing a comparison.

## Required invariance checks

Before using the corrected internal bisect for classification, prove:

1. Applying the source-derived alignment to the Fast value immediately after the coherent-scale metadata transition reproduces the Fast value immediately before that transition within numerical noise.
2. The alignment operation itself introduces no material error.
3. At each DoubleAngle checkpoint, an equivalent plaintext/symbolic recurrence using the same normalization exponent agrees with the aligned interpretation.
4. Fast and Standard source-faithful replays still reproduce the actual public EvalMod outputs with RMSE <= 1e-12.
5. The accepted outer-stage measurements remain unchanged.

## Metrics

For each aligned comparable checkpoint report:

[
D_i^{aligned}=\mathrm{RMSE}(\tilde F_i,S_i)
]

and max complex difference.

Also compute aligned stage-reference SNR using Standard as signal and (	ilde F_i-S_i) as error.

For sequential checkpoints in the same semantic path:

[
A_i^{aligned}=\frac{D_i^{aligned}}{D_{parent}^{aligned}}
]

and aligned (Delta\mathrm{SNR}).

Do not use raw unaligned (D_i), (A_i), or SNR for causal classification after the normalized representation begins.

Raw values may be retained only in a clearly labeled diagnostic column.

## Thresholds

Reuse DIAG-002 thresholds unchanged:

Observable:
[
\Delta_{obs}=10^{-8}
]

Material:
[
\Delta_{mat}=0.003640730397217224
]

A first-material classification requires the previous **aligned comparable** checkpoint to be below (Delta_{mat}) and the current aligned checkpoint to be at or above it.

## Classification

Return exactly one:

- `CURRENT_FAST_ALIGNED_FIRST_MATERIAL_EVALMOD_POLYNOMIAL`
- `CURRENT_FAST_ALIGNED_FIRST_MATERIAL_EVALMOD_EXPONENT_SCHEDULE`
- `CURRENT_FAST_ALIGNED_FIRST_MATERIAL_EVALMOD_DOUBLE_ANGLE`
- `CURRENT_FAST_ALIGNED_FIRST_MATERIAL_EVALMOD_FINAL_RESTORE`
- `CURRENT_FAST_ALIGNED_FIRST_MATERIAL_EVALMOD_PUBLIC_SCALE_RESET`
- `CURRENT_FAST_ALIGNED_EVALMOD_DIVERGENCE_UNCLOSED`

Do not classify the metadata-only coherent-scale transition itself as numerical corruption solely because raw decoded values differ by the intentional normalization factor.

An exponent-schedule classification requires evidence that the source-derived normalized recurrence uses an exponent inconsistent with the Standard-equivalent mathematical recurrence, not merely that raw representations differ.

## Required report

Write:

`results/FAST-STANDARD-NUMERICAL-DIAG-003-summary.md`

Include:

- clean provenance;
- accepted DIAG-002 outer-stage facts;
- explanation of why the old raw internal comparison was semantically misaligned;
- per-checkpoint raw representation exponent;
- raw decoded comparison where useful, clearly labeled non-causal;
- aligned RMSE / max diff / SNR;
- first aligned observable checkpoint;
- first aligned material checkpoint;
- replay verification;
- exact classification;
- exactly one next bounded causal counterfactual;
- limitations.

Required explicit fields:

- `ALIGNED_FIRST_OBSERVABLE_CHECKPOINT=...`
- `ALIGNED_FIRST_OBSERVABLE_MAX_DIFF=...`
- `ALIGNED_FIRST_MATERIAL_CHECKPOINT=...`
- `ALIGNED_FIRST_MATERIAL_MAX_DIFF=...`
- `ALIGNED_CLASSIFICATION=...`
- `NEXT_BOUNDED_COUNTERFACTUAL=...`

## Validation

Run:

- focused Primary fastdiag tests;
- new semantic-alignment unit tests;
- `go test ./cmd/fastdiag ./internal/numericalmetrics`;
- Secondary `go test ./schemes/ckks/fast -count=1`;
- relevant bootstrapping / Mod1 focused tests;
- `git diff --check`.

Secondary must remain clean and unchanged.

## Completion

Return:

`FAST_STANDARD_NORMALIZED_ALIGNMENT_READY`

plus:
- aligned classification;
- first aligned observable/material checkpoints and gaps;
- whether the previous coherent-scale "divergence" disappears after semantic alignment;
- exactly one next bounded counterfactual.

Then:

`READY_FOR_WEB_REVIEW`.
