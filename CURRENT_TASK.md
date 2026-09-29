# Current Task

Task: FAST-STANDARD-NUMERICAL-001
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STANDARD-NUMERICAL-001-P93-REFERENCE.md`

Task class:
`C — Numerical Correctness Validation`

Purpose:
Pause performance work and measure numerical quality on the actual accepted P93 q0=55 / LogN13 / LogSlots12 / 4096-slot workload.

Compare:
- original message;
- current Fast Bootstrap;
- Standard Bootstrap across >=3 independent Standard key trials.

Important:
Current fastdiag only validates Fast diagnostics-on vs diagnostics-off and compares Fast git revisions. It does not yet quantify P93 Fast-vs-Standard decoded error.

Required output:
- Fast vs original metrics;
- Standard vs original metrics;
- Fast vs Standard metrics;
- Standard-to-Standard variability;
- precision-bit comparison;
- 1e-2 threshold audit;
- classification.

Preferred reusable command:
`./scripts/fastdiag numerical --profile p93-q55 --standard-trials 3`

No production arithmetic changes.

Write:
`results/FAST-STANDARD-NUMERICAL-001-summary.md`

Return:
`FAST_STANDARD_NUMERICAL_REFERENCE_READY`

plus required metrics/classification from the spec, then `READY_FOR_WEB_REVIEW`.
