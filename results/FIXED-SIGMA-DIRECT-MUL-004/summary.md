# FIXED-SIGMA-DIRECT-MUL-004 — direct-mul experiment result summary

**PLANNED_NOT_RUN** — no numerical results yet. This is a durable reporting template and must never be presented as a completed experiment.

**Question:** Does the σ calibrated on the historical `(A+B)×C` fixture transfer to genuinely direct `A×B` and `(A×B)×C` with no Add?

Frozen σ: `4.004280012097432`; genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; active Fast initial `1e7828293c36f9845e30723ba4e9500d84bb50d2`. New inputs `direct-base`, `direct-c4` must be fingerprinted separately from 001–003.

| Workload / metric | Standard | Fast OFF/OFF | Fast ON/OFF | Fast OFF/ON | Fast ON/ON |
|---|---|---|---|---|---|
| Direct-base Stage1 incremental RMS, variance, max | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Direct-base Stage2 incremental RMS, variance, max | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Direct-base Stage2 total oracle error | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Direct-c4 Stage1 and Stage2 metrics | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Propagated `P1`, exact residual decomposition and tails | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Actual Level/Scale, strict capacity and built source | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |

`SCOPED_RMS_TRANSFER_SCREEN_PASS` only if two stagewise relative RMS discrepancies to genuine Standard are within predeclared ±10% for **both workloads** and all tested ON/ON seed arms; this is a descriptive screen, not statistical equivalence. Always separately report variance, signed bias, p95/p99/p99.9 and maximum mismatch, and any failure to meet the band. No tuning σ after viewing results.

**Authorized budgets:** Fast 0/16 used, Standard 0/2 used, Bootstrap 0/0. Raw results, source/binary provenance and hash-indexed compact evidence: NOT_RUN.
