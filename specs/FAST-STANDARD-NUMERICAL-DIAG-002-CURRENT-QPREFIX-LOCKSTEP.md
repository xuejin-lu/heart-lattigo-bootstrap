# FAST-STANDARD-NUMERICAL-DIAG-002 — Current Q-prefix Stage Lockstep Bisect

## Status

Executable numerical diagnosis task.

## Task class

`D — Numerical Correctness Diagnosis`

## Accepted parent

- `results/FAST-STANDARD-NUMERICAL-001-summary.md`
- Primary result commit: `c338482bf4066050702622497722237746ac26ef`
- classification: `FAST_NUMERICAL_QUALITY_DEGRADED`

Current accepted production state:
- Secondary active branch: `fast-qprefix`
- control-plane HEAD: `bde527336d25ab7e01692706dc0d185a51d4abe2`
- production arithmetic baseline remains `6930cf6cb3c71ce139a1eb42eede7be335b7174c`

Accepted numerical facts on canonical P93 q0=55 / LogN13 / LogSlots12 / 4096-slot input:
- Fast complex RMSE vs original: `0.020892211083414026`
- Standard complex RMSE median: `1.1903936621664996e-9`
- Fast-vs-Standard complex RMSE median: `0.02089221106956351`
- Fast median precision: `5.7435154 bits`
- Standard median precision: `30.7954521 bits`
- Fast-vs-Standard max complex diff: `0.03640730397217224`
- 46.0815% of Fast-vs-Standard real/imag coordinates exceed `1e-2`.

This task pauses performance work.

## Historical context — reference only

Older `fast-ckks` investigations observed:
- genuine Standard remained near exact;
- first material Fast-vs-Standard divergence appeared in EvalMod;
- an internal bisect once localized a large amplification near normalized/pre-double-angle semantics;
- S2C then amplified upstream EvalMod error.

Those results used older source/branch/parameter details (including q0=56 cases and dirty diagnostic trees). They are **not** current evidence and must be labeled:

`HISTORICAL_FAST_CKKS_REFERENCE_ONLY`

Do not reuse their classification without current-branch reproduction.

Also note:
- current q0=55 Fast path selects `legacyNormalizedLogN13PlanScaleBits = 91`;
- old q0=55 Plan92 experiments did not establish a complete precision repair and found multiple blockers.

Therefore this task must diagnose first. It must not change plan scale in production.

## Goal

On the current clean `fast-qprefix` production implementation, identify the **first material numerical divergence** between Fast and genuine Standard along the Bootstrap pipeline.

Answer:

1. Are Fast and Standard already materially different after C2S?
2. If not, where inside EvalMod does the gap first become material?
3. How much does S2C amplify or transform the EvalMod gap?
4. Is the current error primarily:
   - pre-EvalMod;
   - polynomial/generated-power;
   - normalized scale transition;
   - DoubleAngle;
   - post-EvalMod/S2C;
   - or still unclosed?
5. What is the single next bounded causal counterfactual?

No production repair in this task.

## Canonical workload

Reuse exactly the numerical-001 workload:
- fingerprint `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`;
- LogN=13;
- LogSlots=12;
- 4096 complex slots;
- q0=55;
- same Q/P chain;
- Mod1 degree 30;
- DoubleAngle 3;
- EvalMod log scale 60;
- LogMessageRatio 10.

Fail if the fingerprint changes.

## Apples-to-apples execution

Use:
- current Fast evaluator;
- genuine Standard evaluator with real Standard evaluation keys;
- same deterministic plaintext-like input `c0=encoded message,c1=0`.

Use one Standard key trial for stage lockstep after confirming it reproduces the numerical-001 Standard decoded result within numerical noise. The numerical-001 3-trial result already established Standard-key variation was zero for this workload.

Do not compare raw ciphertext coefficients across Fast/Standard unless transformed into a common semantic representation.

## Common semantic comparison rule

At every checkpoint compare **decoded semantic values**, not raw ciphertext words.

For Standard intermediates:
- decrypt with the Standard secret key;
- decode with metadata appropriate for the checkpoint.

For Fast intermediates:
- where the zero-secret representation permits direct c0 decode, use the existing validated Fast diagnostic decode;
- otherwise use an independently justified semantic decoding path.

If a checkpoint cannot be meaningfully decoded due to representation/scale semantics, report it as non-decodable and use a mathematically aligned alternative checkpoint. Do not fabricate comparisons.

For every comparable checkpoint report:
- Level;
- Scale log2;
- Degree;
- IsNTT;
- IsMontgomery;
- authority rows / maintained rows where relevant;
- complex RMSE Fast-vs-Standard;
- max complex diff;
- median precision vs original if semantically meaningful.

## Materiality thresholds

Use two thresholds:

### Observable
[
Delta_{obs}=10^{-8}
]

### Material
Define from final accepted Fast-vs-Standard max complex gap:
[
Delta_{mat}=0.1	imes 0.03640730397217224
approx 0.003640730397217224
]

A checkpoint is first material divergence only when:
- previous comparable checkpoint < (Delta_{mat});
- current comparable checkpoint >= (Delta_{mat}).

Report first observable divergence separately.

Do not change thresholds after seeing results.

## Stage-level lockstep

Required checkpoints:

1. input
2. after ScaleDown
3. after ModUp
4. after C2S real
5. after C2S imag
6. after EvalMod real
7. after EvalMod imag
8. after S2C
9. final public Bootstrap output

At each stage run the actual public/production stage implementations as much as possible.

Do not replace Standard with a hand-written "matched reference" unless separately labeled. Genuine Standard is authoritative.

## EvalMod internal bisect

If first material stage is EvalMod (real or imag), bisect current Fast and Standard through semantically aligned milestones.

At minimum attempt:

1. EvalMod input before normalization
2. after normalization / scale reinterpretation
3. after Chebyshev offset
4. polynomial input
5. generated-power / polynomial-evaluation output
6. immediately before normalized coherent-scale transition
7. immediately after coherent-scale transition
8. before DoubleAngle round 0
9. after round 0 multiply
10. after round 0 rescale
11. after rounds 1 and 2 similarly where needed
12. EvalMod output before public scale reset
13. EvalMod output after public scale reset

Use current production codepaths.

Where Standard does not expose the identical internal boundary, add diagnostic-only snapshot hooks or construct a source-faithful reference only if exact equivalence is proven.

Do not mutate production arithmetic.

## Generated-power inspection

If polynomial/generated powers are implicated:
- use existing `DiagnosticGeneratePowers` for Fast;
- compare semantically aligned generated powers or polynomial milestones against genuine Standard/reference;
- report power number, Level, Scale, rows, RMSE/max diff.

Do not enable expensive performance tracing; timing is irrelevant here.

## Scale semantics audit

At every EvalMod checkpoint record exact scale values.

Specifically audit current q0=55 normalized path:
- selected plan bits;
- working scale;
- target scale;
- coherent scale;
- `kIn`;
- per-round multiplier exponents;
- post-rescale scale.

The current source's q0=55 selection of plan scale 91 is an observed implementation fact, not a presumed cause.

Do not change it in this task.

## S2C amplification

If EvalMod output already differs:
- feed actual Fast EvalMod output through Fast S2C;
- feed actual Standard EvalMod output through Standard S2C;
- measure before/after gap.

Where safe, add diagnostic cross-feed counterfactuals:
- Standard-like EvalMod semantic input through Fast S2C;
- Fast-like EvalMod semantic input through a matched linear S2C reference.

Purpose:
determine whether S2C is causal source or mostly an amplifier.

No production mutation.

## SNR instrumentation — required

This diagnosis must extend the reusable numerical/debug framework with explicit decoded-domain SNR metrics.

### Terminology

Do not use the bare term `SNR` without naming the reference.

There are two different metrics:

1. **Bootstrap SNR** — within one execution mode, compare the decoded semantic vector immediately before Bootstrap with the decoded final Bootstrap output.
2. **Stage reference SNR** — at a semantically comparable checkpoint, use genuine Standard as the reference signal and Fast-minus-Standard as the error/noise.

These are numerical signal-to-distortion metrics. They are **not** RLWE security noise, remaining noise budget, or a security claim.

The canonical numerical-001 workload uses plaintext-like `c0=encoded message, c1=0`. Therefore the reported Bootstrap SNR measures numerical distortion introduced by Bootstrap on this workload; it does not characterize encryption noise.

### A. End-to-end Bootstrap SNR

For each mode separately, decode the actual ciphertext entering Bootstrap and the final output through the mode-appropriate validated semantic recovery path.

Let

[
x_i^{\mathrm{pre}}
=
\text{decoded slot }i\text{ immediately before Bootstrap},
]

[
x_i^{\mathrm{post}}
=
\text{decoded slot }i\text{ after Bootstrap},
]

and

[
e_i=x_i^{\mathrm{post}}-x_i^{\mathrm{pre}}.
]

Define

[
P_{\mathrm{signal}}
=
\frac{1}{n}\sum_i |x_i^{\mathrm{pre}}|^2,
]

[
P_{\mathrm{noise}}
=
\frac{1}{n}\sum_i |e_i|^2,
]

[
\boxed{
\mathrm{SNR}_{\mathrm{bootstrap,dB}}
=
10\log_{10}
\frac{P_{\mathrm{signal}}}{P_{\mathrm{noise}}}
}
]

or equivalently

[
\mathrm{SNR}_{\mathrm{bootstrap,dB}}
=
20\log_{10}
\frac{\mathrm{RMS}(x^{\mathrm{pre}})}
{\mathrm{RMSE}(x^{\mathrm{post}},x^{\mathrm{pre}})}.
]

Required for:
- current Fast Q-prefix;
- genuine Standard.

Also report:
- signal power;
- signal RMS;
- noise/error power;
- noise/error RMSE;
- Bootstrap SNR in dB;
- pre-Bootstrap decoded-vs-canonical-original RMSE as a sanity guard.

Do not silently substitute the canonical original message for the actual decoded pre-Bootstrap vector in the authoritative Bootstrap-SNR result.

### B. Stage reference SNR

At every semantically comparable Fast-vs-Standard checkpoint (i), let:
- (F_i) be the decoded Fast semantic vector;
- (S_i) be the decoded genuine-Standard semantic vector.

Retain the existing divergence metric:

[
\boxed{
D_i=\mathrm{RMSE}(F_i,S_i)
}
]

and define the stage-to-stage raw error amplification:

[
\boxed{
A_i=\frac{D_i}{D_{i-1}}
}
]

when both adjacent checkpoints are comparable and (D_{i-1}>0).

Define Standard-reference signal power:

[
P_i
=
\frac{1}{n}\sum_j |S_{i,j}|^2,
]

and normalized stage reference SNR:

[
\boxed{
\mathrm{SNR}_{i,\mathrm{dB}}
=
10\log_{10}
\frac{P_i}{D_i^2}
}
]

because (D_i^2) is the mean squared complex Fast-minus-Standard error.

Also define:

[
\boxed{
\Delta\mathrm{SNR}_i
=
\mathrm{SNR}_i-\mathrm{SNR}_{i-1}
}
]

when adjacent stage SNRs are comparable.

Interpretation:
- large (A_i): raw Fast-vs-Standard error expanded strongly at stage (i);
- large negative (Delta\mathrm{SNR}_i): stage (i) materially worsened error relative to the signal magnitude;
- neither metric alone proves the stage is causal; use them together with the first-observable/material-divergence rules and bounded counterfactuals.

This normalized SNR view is important because decoded signal magnitude may change substantially across C2S, EvalMod, and S2C; raw RMSE amplification alone can be misleading.

### C. Complex-vector convention

For complex CKKS slots use:

[
|z|^2=(\Re z)^2+(\Im z)^2.
]

The primary SNR is one complex-vector SNR. Do not average separate real-SNR and imag-SNR values into a pseudo-complex SNR.

Real/imag diagnostic SNRs may be emitted only as clearly secondary fields if useful.

### D. Reusable metric schema

Add one reusable metric type to the Primary numerical diagnostic model, conceptually containing:

- `signal_power`;
- `signal_rms`;
- `noise_power`;
- `noise_rmse`;
- finite `snr_db` when defined;
- explicit status/reason for non-finite or non-comparable cases.

Required edge handling:
- zero noise and nonzero signal => mathematically (+\infty); do not emit JSON NaN/Inf;
- zero signal => SNR undefined;
- non-decodable/non-semantic checkpoint => not comparable;
- no arbitrary epsilon floor to manufacture a finite SNR.

Prefer a nullable finite numeric `snr_db` plus an explicit status such as:
- `FINITE`;
- `POSITIVE_INFINITY`;
- `UNDEFINED_ZERO_SIGNAL`;
- `NOT_COMPARABLE`.

### E. Required tests

Add focused tests that prove:

1. a synthetic known signal/error pair produces the expected dB value;
2. scaling both signal and error by the same nonzero constant leaves SNR unchanged;
3. `noise_rmse` equals the existing complex RMSE for the same pair;
4. exact equality is handled without writing JSON NaN/Inf;
5. zero-signal input is handled explicitly;
6. existing numerical JSON remains parseable and existing RMSE/precision values are unchanged.

No production arithmetic changes are authorized by this SNR instrumentation.

### F. Required reporting

The final numerical diagnosis summary must include:

#### End-to-end table

| Mode | Pre→Post RMSE | Signal RMS | Bootstrap SNR (dB) |
|---|---:|---:|---:|
| Standard | ... | ... | ... |
| Fast Q-prefix | ... | ... | ... |

#### Stage table

For every comparable checkpoint include at least:

| Checkpoint | (D_i) RMSE | (A_i) | Stage reference SNR (dB) | (Delta\mathrm{SNR}_i) (dB) | Max complex diff |
|---|---:|---:|---:|---:|---:|

Report explicitly:
- `STANDARD_BOOTSTRAP_SNR_DB=...`
- `FAST_BOOTSTRAP_SNR_DB=...`
- `LARGEST_RAW_AMPLIFICATION_CHECKPOINT=...`
- `LARGEST_RAW_AMPLIFICATION_FACTOR=...`
- `LARGEST_SNR_DROP_CHECKPOINT=...`
- `LARGEST_SNR_DROP_DB=...`

Do not introduce an SNR pass/fail threshold in this task. SNR is diagnostic evidence, not a new acceptance criterion.


## Required classification

Return exactly one:

- `CURRENT_FAST_FIRST_MATERIAL_C2S`
- `CURRENT_FAST_FIRST_MATERIAL_EVALMOD_POLYNOMIAL`
- `CURRENT_FAST_FIRST_MATERIAL_EVALMOD_NORMALIZATION`
- `CURRENT_FAST_FIRST_MATERIAL_EVALMOD_DOUBLE_ANGLE`
- `CURRENT_FAST_FIRST_MATERIAL_S2C`
- `CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED`

Also report:
- `FIRST_OBSERVABLE_CHECKPOINT=...`
- `FIRST_OBSERVABLE_MAX_DIFF=...`
- `FIRST_MATERIAL_CHECKPOINT=...`
- `FIRST_MATERIAL_MAX_DIFF=...`
- `FINAL_FAST_STANDARD_RMSE=...`
- `S2C_AMPLIFICATION_FACTOR=...` if defined.

## Next bounded counterfactual

Name exactly one next causal experiment.

Examples:
- q0=55 plan-scale 91→92 test-only counterfactual;
- coherent-scale exponent correction;
- one generated-power replacement;
- one DoubleAngle round semantic correction;
- one S2C scale/representation correction.

The candidate must be chosen from **current evidence**, not historical intuition.

Do not implement it here.

## Validation

Run:
- focused diagnostic tests;
- relevant Fast/Standard bootstrap tests;
- `go test ./schemes/ckks/fast -count=1`;
- `go test ./circuits/ckks/bootstrapping -run <relevant focused tests> -count=1`;
- Primary diagnostic/framework tests if changed;
- `git diff --check`.

Secondary production source must remain unchanged. Diagnostic-only hooks/tests may be committed only if reusable and behavior-neutral; otherwise remove them.

## Required artifact

Write:

`results/FAST-STANDARD-NUMERICAL-DIAG-002-summary.md`

Include:
- provenance;
- canonical fingerprint;
- stage table;
- internal EvalMod table if applicable;
- scale audit;
- generated-power evidence if applicable;
- S2C attribution;
- first observable/material divergence;
- historical-reference reconciliation;
- classification;
- exactly one next bounded counterfactual;
- limitations.

## Completion

Return:

`FAST_STANDARD_NUMERICAL_BISECT_READY`

plus:
- classification;
- first observable/material checkpoints and gaps;
- final RMSE;
- S2C amplification if defined;
- next bounded counterfactual.

Then:

`READY_FOR_WEB_REVIEW`.
