# Current Task

Task: QPREFIX-IMPL-009
Status: READY_FOR_CODEX

Authoritative specification:
`specs/QPREFIX-IMPL-009-PERFORMANCE-RELEASE-GATE.md`

Accepted production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Historical pre-F comparison point:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

008 re-review result:
- production public-boundary code remains accepted;
- BootstrapMany derives input authority from input logical Level and output authority again from core-output logical Level;
- explicit packing/ring-degree kernels may support narrower rows for compatibility/testing, but that is capability only;
- the Q-prefix v2 production constitution is exact:
  `rows = QPrefixWidth(Level) = min(Level+1,4)`;
- "internal q0123" is valid only when current logical Level >= 3;
- Level 2 must be q012, Level 1 q01, Level 0 q0;
- production must neither under-maintain nor over-maintain authority relative to logical Level.

Clarified specs:
- QPREFIX-IMPL-008 spec updated at Primary commit `a9f5e43ff4c5c968218f8a3d903f2e3d35ed3d2b`;
- QPREFIX-IMPL-009 release gate updated at Primary commit `3ace6637ac793db1e246a1abfd9ccd2612abab1b`.

Critical release-gate correction:
Do not merely prove `rows <= 4`.
For every production stage and every logical-Level transition prove:
`authoritative rows == QPrefixWidth(logical Level)`.

Examples:
- L0 -> q0
- L1 -> q01
- L2 -> q012
- L>=3 -> q0123
- 3->2: q0123 -> q012
- 2->1: q012 -> q01
- 1->0: q01 -> q0

Proceed with the final validation/performance gate under this exact invariant.
Do not modify production source unless a hard release bug is found and reported first.

Final status must be one of:
- `QPREFIX_V2_RELEASE_PASS`
- `QPREFIX_V2_PERFORMANCE_REVIEW`
- `QPREFIX_V2_RELEASE_FAIL`
