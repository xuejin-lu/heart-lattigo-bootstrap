# Current Task

Task: QPREFIX-PERF-DIAG-001
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-DIAG-001-ATTRIBUTION.md`

Task class:
`D — Performance Diagnosis`

QPREFIX-IMPL-009 result:
`QPREFIX_V2_PERFORMANCE_REVIEW`

Recorded result:
`results/QPREFIX-IMPL-009-PERFORMANCE-REVIEW.md`

Production candidate remains:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Historical baseline:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Critical interpretation:
- matched q0=55 historical baseline generally uses q01 at high Level because legacy `MaintainedLimbCount` selects 3 rows only for q0 bit length 56;
- current Q-prefix v2 production uses exact constitution `rows=QPrefixWidth(Level)`, hence q0123 at Level>=3;
- the observed ~3.8x regression is therefore a system-policy migration cost comparison, not proof of an inefficient q012->q0123 implementation.

Goal:
Attribute the slowdown into:
- mandatory width cost;
- CRT/Rescale cost;
- PS/polynomial cost;
- guard/DoubleAngle;
- copy/allocation/memory traffic;
- DFT/packing/ring-degree overhead.

Do not modify production code.
Do not narrow production authority below `QPrefixWidth(Level)`.
Do not reopen F/full-RNS fallback or architecture.

Temporary benchmark/test instrumentation is allowed only as specified and must be removed.

Write:
`results/QPREFIX-PERF-DIAG-001-summary.md`

Return one:
- `WIDTH_COST_DOMINANT`
- `RESCALE_COST_DOMINANT`
- `MEMORY_COPY_COST_DOMINANT`
- `MIXED_COST`
- `UNEXPLAINED_PERFORMANCE_REGRESSION`

Then report `READY_FOR_WEB_REVIEW`.
