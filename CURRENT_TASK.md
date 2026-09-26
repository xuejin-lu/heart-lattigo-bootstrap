# Current Task

Task: QPREFIX-AUDIT-002
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-AUDIT-002-C2S-RAW-RESCALE.md`

Task class:
`M/E — Focused evidence completion`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: evidence/report.
- Secondary `xuejin-lu/lattigo@fast-qprefix`: current-branch C2S audit target.

Accepted prerequisite:
- QPREFIX-AUDIT-001 at `494e7b7b672cbedb4eb842887d464df2ea0cb7f9`.

Close only the first audit gap:
- reproduce current-branch C2S groups 0..3;
- capture raw LinearTransform output before each Rescale;
- measure exact maintained-prefix coefficient bounds;
- record each logical divisor and predicted/observed Rescale bound;
- record restore propagation;
- verify strict capped Q-prefix capacity;
- prove equivalence to the combined production C2S helper.

Do not modify production behavior, `fast-ckks`, parameters, schedules, EvalMod, or S2C.

Return one of:
- `QPREFIX_C2S_CAPACITY_PROVEN`
- `QPREFIX_C2S_CAPACITY_FAIL`
- `QPREFIX_C2S_MEASUREMENT_CONFLICT`
- `QPREFIX_C2S_EVIDENCE_INCOMPLETE`.
