# FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013 journal

**State:** BATCH_COMPLETE_READY_FOR_WEB_REVIEW
**Primary base:** e7dbe218c53b7f92df26da80f9a099c8b24e434a
**Standard pin:** 5dbffbdea05394de2ca3a432ed5318aa832e3f40 (clean)
**Fast pin:** 00ac70ba136d190fa31bbb26c2f51d003a221634 (clean)

| Checkpoint | State | Evidence / verification | Cost / attempt notes |
|---|---|---|---|
| A — audit and fixtures | Complete | Reused 005 LogN13 profile and deterministic A/B vectors; read 005, 006A, 009 and accepted 012 evidence. Verified identical public signatures and source paths; selected Levels 3 and 2 with exact Q/P fingerprints. | Read-only; no Bootstrap or benchmark. |
| B — Add/Sub | Pass | Same source at both pins; c/c Add/Sub, AddNew/SubNew, scalar/vector overloads and op0 alias checks. Cleartext, input immutability, Level/Scale/Degree, dimensions, domain, c1, and active-row checks passed. | No backend edits. |
| C — MulRelin | Pass | c/c MulRelin/MulRelinNew, product, square, alias, scalar/vector, unsupported-type rejection, and Fast invalid-c1 transactionality. Standard: native key lookups; Fast: zero relin lookups. | One local repair pass fixed a runner error-variable shadowing defect in the negative control; final paired runs and tests were rerun. No production failure. |
| D — Rescale + composition | Pass | Isolated Rescale and supported RescaleTo each consumed q3 once (Level 3→2, Scale/q3); checked Add→MulRelin→Rescale→Rotate at every link with plaintext oracles and full-active-Q backing. | Zero Bootstrap; no C0 or dormant-row path. |
| E — matrix, review, validation | Complete | Combined 24 paired checkpoints, matching source/test/profile/QP/input hashes, backend commits, path and key-lookup evidence. Added compact summary and aggregate JSON. Standard/Fast package tests and vet passed; final diff review and git diff --check completed before commit. | One Standard and one Fast final runner execution; no benchmark, LogN16, or Bootstrap. |

## Resume / repository boundaries

- Primary was clean on main before implementation, fetched and fast-forwarded to origin/main; synchronized HEAD and remote were e7dbe218c53b7f92df26da80f9a099c8b24e434a.
- Secondary fast-qprefix was clean, fetched and already up to date at 00ac70ba136d190fa31bbb26c2f51d003a221634; Standard workspace was clean at its pin. No Secondary or Standard files changed.
- CURRENT_TASK.md, specs, parameters, benchmarks, and production code were not edited.
- Deliverables are the new tools/fast-dropin-public-primitives-batch-013/ runner/tests and the three matching results/ artifacts.
- Expensive-task budget used: 0 Bootstrap, 0 benchmarks, 0 LogN16. No further task is authorized by this batch.
- Final action: review exact Primary diff, run git diff --check, commit only the listed deliverables, then ordinary fast-forward push to origin/main if the standing safe-push conditions still hold.
