# Current Task

Task: QPREFIX-PERF-OPT-004
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-OPT-004-BARRETT-FIXED-WIDTH-REDUCTION.md`

Task class:
`P — Performance Repair`

Accepted parent:
- `results/QPREFIX-PERF-OPT-003-summary.md`
- OPT-003 decision `SOURCE_INTT_BATCHING_CORRECT_BUT_NO_WIN`
- accepted production implementation remains `6930cf6cb3c71ce139a1eb42eede7be335b7174c`

Fresh current evidence:
- `math/bits.Div64` remains largest single flat symbol at 20.55%;
- fixed-width phase 36.55%;
- signed residue staging 21.56%;
- transform loop reordering showed no win.

Goal:
Evaluate exact Barrett-Horner modular reduction using Lattigo's existing Barrett primitives to reduce hot Div64 work in Rescale.

Important:
- feasibility benchmark first;
- no production change unless mod192 improves >=15% and signed-residue or rows4 CRT improves >=8%;
- exact equality against current helpers and BigInt;
- no approximate reciprocal, unsafe, assembly, concurrency, width-policy or schedule changes.

Write:
`results/QPREFIX-PERF-OPT-004-summary.md`

Return one:
- `BARRETT_FIXED_WIDTH_REDUCTION_READY`
- `BARRETT_FIXED_WIDTH_REDUCTION_CORRECT_BUT_NO_WIN`
- `BARRETT_FIXED_WIDTH_REDUCTION_BLOCKED`

Then report `READY_FOR_WEB_REVIEW`.
