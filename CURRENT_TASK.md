# Current Task

Task: QPREFIX-PERF-OPT-002
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-OPT-002-FIXED-WIDTH-DIVISION-DEDUP.md`

Task class:
`P — Performance Repair`

Accepted parent:
- `results/QPREFIX-PERF-DIAG-006-summary.md`
- classification `RESCALE_MIXED_KERNELS`
- top isolated phase: fixed-width reconstruct / round / capacity = `43.51%`
- current production arithmetic baseline `d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

Accepted pprof evidence:
- `math/bits.Div64` = 27.40% flat;
- domain transforms are also material;
- `mod192By64`, CRT reconstruction, and signed residue work are hot.

Goal:
Perform a bounded exact-arithmetic cleanup:
- remove provably redundant leading Div64 operations in fixed-width modular reduction;
- review quotient/remainder duplication in roundedMagnitude192;
- reuse already prepared invariant prefix products in the Rescale reconstruction path where exact;
- preserve all semantics and transactionality.

Do not introduce reciprocal approximations, unsafe, assembly, width-policy changes, or schedule changes.

Write:
`results/QPREFIX-PERF-OPT-002-summary.md`

Return one:
- `FIXED_WIDTH_DIVISION_DEDUP_READY`
- `FIXED_WIDTH_DIVISION_DEDUP_CORRECT_BUT_NO_WIN`
- `FIXED_WIDTH_DIVISION_DEDUP_BLOCKED`

Then report `READY_FOR_WEB_REVIEW`.
