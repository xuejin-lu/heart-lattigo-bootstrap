# Current Task

Task: QPREFIX-AUDIT-003
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-AUDIT-003-EVALMOD-PS-DA-CAPACITY.md`

Task class:
`M/E — Focused evidence completion`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: evidence/report.
- Secondary `xuejin-lu/lattigo@fast-qprefix`: current-branch EvalMod/PS/DA audit target.

Accepted prerequisites:
- QPREFIX-AUDIT-001 at `494e7b7b672cbedb4eb842887d464df2ea0cb7f9`.
- QPREFIX-AUDIT-002 accepted at `78ce1599ac9a118fdd8b246d8bb52d4d4dbaeb7f`.

Audit only current-branch EvalMod / generated powers / Paterson-Sockmeyer / DoubleAngle:
- exact observed bounds;
- conservative recurrence bounds;
- every logical Rescale divisor and predicted/observed bound;
- strict Q-prefix-v2 capacity;
- relevant lower-Level crossing semantics;
- public Bootstrap semantic guard.

Do not modify production behavior, `fast-ckks`, parameters, polynomial schedule, S2C, or introduce F.

Return one of:
- `QPREFIX_EVALMOD_CAPACITY_PROVEN`
- `QPREFIX_EVALMOD_CAPACITY_FAIL`
- `QPREFIX_EVALMOD_PROOF_TOO_LOOSE`
- `QPREFIX_EVALMOD_MEASUREMENT_CONFLICT`
- `QPREFIX_EVALMOD_EVIDENCE_INCOMPLETE`.
