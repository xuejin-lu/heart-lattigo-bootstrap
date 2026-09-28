# Current Task

Task: QPREFIX-IMPL-009
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-009-PERFORMANCE-RELEASE-GATE.md`

Task class:
`V — Validation / Release Gate`

Accepted production candidate:
QPREFIX-IMPL-008 at
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Historical pre-F comparison point:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Review result for QPREFIX-IMPL-008:
- PASS;
- explicit 1..4-row N1<->N2 conversion works;
- explicit packing/unpacking preserves caller authority;
- BootstrapMany derives input/output rows from actual Levels;
- public Level 0/1 remains q0/q01;
- finalization produces ordinary NTT/non-Montgomery public ciphertexts;
- dormant-row isolation, Standard/public correctness, full regression, and structural performance guards pass.

Goal:
Run the final Q-prefix v2 release gate.

Do not modify production code unless a hard release bug is found and reported first.

Required:
- isolate pre-F baseline in a detached worktree;
- run matched same-code benchmarks;
- run a byte-identical temporary q0=55 P93 public-API benchmark on baseline and current;
- run current q0=56 P93 Fast count1/count3 and matched Standard count1;
- rerun semantic, capacity, structure, fallback, and public-output gates;
- report exact performance ratios and one final release status.

Important:
- q0=55 baseline and q0=56 production are not directly comparable as pure code changes;
- preserve that distinction in the report;
- no optimization campaign;
- no F/full-RNS fallback;
- no `fast-ckks` modification.

Final status must be exactly one of:
- `QPREFIX_V2_RELEASE_PASS`
- `QPREFIX_V2_PERFORMANCE_REVIEW`
- `QPREFIX_V2_RELEASE_FAIL`

Commit/push only task/report artifacts if authorized by the spec; do not change production source.
Then report `READY_FOR_WEB_REVIEW`.
