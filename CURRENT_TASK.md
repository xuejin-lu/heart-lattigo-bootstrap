# Current Task

Task: QPREFIX-IMPL-002
Status: READY_FOR_CODEX

Architecture plan:
`docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

Accepted prerequisite:
- QPREFIX-IMPL-001 at Secondary commit `6553491f9fb9b964c8fd0d743f3a302de83d0b54`.

Task class:
`I — Implementation`

Goal:
Connect compact ciphertext lifecycle and evaluator-owned storage allocation to the single Q-prefix v2 policy.

Frozen policy:
[
w_Q(ell)=min(ell+1,4)
]

Scope:
- constructors for Fast-owned ciphertexts/temporaries;
- copy/copy-new behavior;
- resize/degree growth where it controls physical row backing;
- evaluator scratch/buffer allocation;
- output allocation helpers;
- preserve logical Level header independently from physically maintained rows;
- dormant rows must remain unallocated/stale and must not become authoritative.

Acceptance:
- Level 0/1/2/3/high-Level Fast-owned temporaries have exactly policy-width N-sized backing rows;
- copy/alias/degree-growth behavior preserves maintained rows exactly;
- shrinking logical Level across 3->2, 2->1, 1->0 changes physical maintained width structurally but does not invent representative semantics;
- no dormant-row reads/writes;
- existing production semantics remain unchanged because arithmetic producers/consumers are not yet generalized;
- focused lifecycle tests plus relevant fast package regressions and `go test ./...`.

Out of scope:
- arithmetic primitive widening;
- Rescale semantics;
- ModUp;
- DFT/EvalMod/Bootstrap;
- F;
- adaptive width;
- `fast-ckks`.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
