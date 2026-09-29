# Current Task

Task: FAST-STANDARD-NUMERICAL-DIAG-002
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STANDARD-NUMERICAL-DIAG-002-CURRENT-QPREFIX-LOCKSTEP.md`

Task class:
`D — Numerical Correctness Diagnosis`

Accepted parent:
- `results/FAST-STANDARD-NUMERICAL-001-summary.md`
- classification `FAST_NUMERICAL_QUALITY_DEGRADED`
- current Secondary branch `fast-qprefix`

Accepted numerical facts:
- Fast complex RMSE `0.020892211083414026`
- Standard complex RMSE median `1.1903936621664996e-9`
- Fast median precision `5.7435154 bits`
- Standard median precision `30.7954521 bits`
- Fast-vs-Standard max complex diff `0.03640730397217224`

Goal:
Localize the first material numerical divergence between current Fast and genuine Standard on the canonical P93 q0=55 workload.

Required stage lockstep:
input -> ScaleDown -> ModUp -> C2S -> EvalMod -> S2C -> final output.

If EvalMod is first material, bisect the current normalized/polynomial/DoubleAngle path and audit exact scales.

Important:
- historical fast-ckks findings are reference only;
- do not change plan scale or production arithmetic;
- no performance optimization in this task.

Write:
`results/FAST-STANDARD-NUMERICAL-DIAG-002-summary.md`

Return:
`FAST_STANDARD_NUMERICAL_BISECT_READY`
plus required classification/checkpoints, then `READY_FOR_WEB_REVIEW`.
