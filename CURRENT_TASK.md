# Current Task

Task: FAST-STORAGE-003
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STORAGE-003-BOUNDED-PRIVATE-STORAGE-ARITHMETIC.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisite:
`FAST-STORAGE-002` at Secondary commit
`3b57a52b311397e0e1cf8298027782eec20d4ffc`.

Implement only the bounded private-storage arithmetic foundation defined by the spec:
- per-component proven coefficient bounds;
- exact width planner;
- exact F-basis expansion;
- standalone Add/Sub;
- standalone raw Mul.

Do not wire into production Bootstrap/Evaluator and do not implement Rescale, ModUp, KeySwitch, Relinearize, Rotate, storage contraction, or application changes.

Codex must follow the normal startup sync and bounded implementation -> self-review -> one repair pass -> validation workflow, commit/push Secondary `fast-ckks`, then report `READY_FOR_WEB_REVIEW`.
