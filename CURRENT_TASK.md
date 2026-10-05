# Current Task

Task: FAST-STANDARD-NUMERICAL-REPORT-001
Status: READY_FOR_CODEX

Task class:
`R — Diagnostic / Reporting Repair`

Authoritative spec:
`specs/FAST-STANDARD-NUMERICAL-REPORT-001-CLOSE-CLASSIFIER.md`

Accepted numerical state:
- Secondary production commit: `5117fc57949647182f476dc5952099c706b9f869`
- Primary accepted result commit: `88e6e6badc0973b64bb051a0ce7fc51cf3f95e75`
- top-level classification: `FAST_STANDARD_NUMERICAL_CLOSE`
- final Fast-vs-Standard RMSE: `1.44680895443e-10`
- Standard Bootstrap SNR: `144.713390 dB`
- Fast Bootstrap SNR: `144.710750 dB`
- no first observable/material checkpoint;
- all 56 Q-prefix capacity checks pass.

Goal:
Fix only the legacy stage-lockstep classifier so a healthy run with no observable divergence does not retain `CURRENT_FAST_NUMERICAL_DIVERGENCE_UNCLOSED`.

Expected accepted stage classification:
`CURRENT_FAST_NO_OBSERVABLE_DIVERGENCE`

Do not modify Secondary or production arithmetic.
Do not change thresholds or accepted measurements.

Required completion token:
`FAST_STANDARD_NUMERICAL_REPORT_FIX_READY`

Then:
`READY_FOR_WEB_REVIEW`.
