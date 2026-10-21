# FIXED-SIGMA-CHAIN-002 — two-step noise compensation report

**Status: NOT_RUN.** Future Codex fills this report with genuine tested values, hashes, source/binary provenance, math gates and limitations. No performance or scientific success is pre-asserted.

## Stagewise planned comparison

| Stage / metric | Genuine Standard | Original Fast | Fast + fixed σ |
|---|---|---|---|
| 1: Level5→4, incremental RMS/variance/max | NOT_RUN | NOT_RUN | NOT_RUN |
| 1: decoded total oracle error | NOT_RUN | NOT_RUN | NOT_RUN |
| 2: Level4→3, incremental RMS/variance/max | NOT_RUN | NOT_RUN | NOT_RUN |
| 2: propagated total oracle error | NOT_RUN | NOT_RUN | NOT_RUN |
| Logical complete ct×ct operations | NOT_RUN | NOT_RUN | NOT_RUN |
| Noise injections | 0 | 0 | NOT_RUN (target exactly 2) |
| Frozen σ | N/A | N/A | `4.004280012097432` (NOT_RUN) |
| Overhead components and timing | N/A | N/A | NOT_RUN |

Bootstrap calls: 0 authorized, 0 performed as of task creation. Do not use old Single-Op result as a two-step measurement.
