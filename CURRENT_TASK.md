# Current Task

Task: QPREFIX-IMPL-004
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-004-RESCALE-LEVEL-TRANSITIONS.md`

Task class:
`I — Implementation`

Accepted prerequisites:
- QPREFIX-IMPL-001: `6553491f9fb9b964c8fd0d743f3a302de83d0b54`
- QPREFIX-IMPL-002: `91baa6a4655e10fe2460a399633fa54a03a35318`
- QPREFIX-IMPL-003: `c8b591a30c05a2261de8d0181d7b3f64bc58169b`

QPREFIX-IMPL-003 review result:
- PASS;
- explicit width 1..4 kernels are validated;
- production wrappers remain on legacy authoritative-row selection;
- poisoned q3 rows are isolated and not promoted accidentally;
- full regression passed.

Implement Q0123 Rescale and explicit Level-transition semantics exactly as specified.

Critical rules:
- Level >=3 Rescale source authority becomes q0123;
- divisor is always logical q_ell, not highest maintained q;
- 3->2, 2->1, 1->0 physically contract only after rounded division;
- structural Resize remains semantics-free;
- SameLift and Canonical DropLevel are distinct explicit operations;
- no dormant rows, F, or Standard full-RNS fallback.

Run all validation required by the spec.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
