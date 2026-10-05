# Current Task

Task: FAST-STANDARD-NUMERICAL-DIAG-003
Status: READY_FOR_CODEX

Task class:
`D — Numerical Correctness Diagnosis`

Authoritative spec:
`specs/FAST-STANDARD-NUMERICAL-DIAG-003-NORMALIZED-SEMANTIC-ALIGNMENT.md`

Accepted parent:
- `results/FAST-STANDARD-NUMERICAL-DIAG-002-summary.md`
- measurement Primary: `4fd11a5aa557a9e91c0473c94f76b9bfc41a9e36` (clean)
- result commit: `c379930340dcfcdcfafa4437c5bcbd4356c6c036`
- Secondary: `97c1c6174e0d7ef781de6d8dadce5c2869a53496` (clean)

Accepted outer-stage facts:
- C2S real/imag remain near Standard;
- first material outer stage is EvalMod;
- Standard Bootstrap SNR = 144.71339028083412 dB;
- Fast Bootstrap SNR = -0.17249085395397162 dB;
- final Fast-vs-Standard RMSE = 0.02089221106956351;
- combined-branch S2C amplification = 2.8284271247428943.

Web-review correction:
The prior internal claim that the coherent-scale transition itself is the first material numerical corruption is not accepted. The Fast normalized EvalMod intentionally carries a power-of-two-normalized internal semantic representation; raw decoded Fast values after that transition are not directly comparable with raw Standard decoded values.

Goal:
Re-run the internal EvalMod bisect with source-derived semantic alignment and identify the true first aligned observable/material divergence.

Do not modify production arithmetic.
Do not execute the coherent-scale exponent counterfactual yet.
Do not tune plan scale.

Required output:
`results/FAST-STANDARD-NUMERICAL-DIAG-003-summary.md`

Required completion token:
`FAST_STANDARD_NORMALIZED_ALIGNMENT_READY`

Then:
`READY_FOR_WEB_REVIEW`.
