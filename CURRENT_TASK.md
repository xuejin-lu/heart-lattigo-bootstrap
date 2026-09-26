# Current Task

Task: QPREFIX-AUDIT-001
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-AUDIT-001-FULL-BOOTSTRAP-CAPACITY.md`

Task class:
`M/E — Architecture audit + reproducible measurement`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/report.
- Secondary `xuejin-lu/lattigo@fast-qprefix`: audit target.

Secondary branch point:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Authoritative Q-prefix v2 architecture:
`docs/FAST_QPREFIX_SPEC.md` at `c9fb900314a1c82f2cd21320554d3d47fd7dbf00`.

Audit only. Do not modify production behavior.

Question to answer:
Does the current LogN13/P93 Bootstrap profile fit the capped, naturally shrinking maintained Q-prefix

[
q_0,ldots,q_{min(ell,3)}
]

at every required authoritative-lift boundary?

Produce the compact JSON + markdown report required by the spec and classify:
- `QPREFIX_V2_FULL_PROFILE_CAPACITY_PROVEN`,
- `QPREFIX_V2_FAILS_AT_STAGE`, or
- `QPREFIX_V2_AUDIT_EVIDENCE_INCOMPLETE`.

Do not modify `fast-ckks`, introduce F primes, implement q0123 production arithmetic, retune parameters, or start optimization.
