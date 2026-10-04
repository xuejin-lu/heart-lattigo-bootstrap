# Current Task

Task: FAST-STANDARD-NUMERICAL-DIAG-002
Status: READY_FOR_CODEX

Task class:
`D — Numerical Correctness Diagnosis`

Authoritative spec:
`specs/FAST-STANDARD-NUMERICAL-DIAG-002-CURRENT-QPREFIX-LOCKSTEP.md`

Accepted parent:
- `results/FAST-STANDARD-NUMERICAL-001-summary.md`
- Primary result commit: `c338482bf4066050702622497722237746ac26ef`
- classification: `FAST_NUMERICAL_QUALITY_DEGRADED`

Goal:
Identify the first material numerical divergence between current Q-prefix Fast and genuine Standard along the Bootstrap pipeline.

Required numerical instrumentation:
- preserve RMSE / max-diff / precision diagnostics;
- add end-to-end decoded **Bootstrap SNR** for Standard and Fast using actual decoded pre-Bootstrap vs post-Bootstrap vectors;
- add per-checkpoint (D_i), (A_i), Standard-reference stage SNR, and (Delta\mathrm{SNR}_i);
- distinguish Bootstrap SNR from Fast-vs-Standard stage reference SNR;
- do not interpret either as RLWE security noise or noise budget.

Primary diagnosis:
- input;
- ScaleDown;
- ModUp;
- C2S real/imag;
- EvalMod real/imag;
- S2C;
- final public output;
- if EvalMod is first material, continue the internal EvalMod bisect defined in the spec.

No production repair in this task.
No performance optimization in this task.
Diagnostic-only reusable instrumentation is allowed and must be behavior-neutral.

Required output:
`results/FAST-STANDARD-NUMERICAL-DIAG-002-summary.md`

Required completion token:
`FAST_STANDARD_NUMERICAL_BISECT_READY`

Then:
`READY_FOR_WEB_REVIEW`
