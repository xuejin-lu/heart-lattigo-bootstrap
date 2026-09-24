# Current Task

Task: FAST-STORAGE-001
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STORAGE-001-PRIVATE-STORAGE-BASIS-FOUNDATION.md`

Task class:
`I — Implementation`

Purpose:
Create the backend-private fixed three-prime Fast storage-basis foundation, independent of logical CKKS Q, before any production arithmetic integration.

Authorized repositories:
- Primary: task/documentation coordination.
- Secondary `xuejin-lu/lattigo@fast-ckks`: storage foundation implementation and tests.

Important prohibitions:
- no current ciphertext/evaluator integration;
- no Fast Rescale/ModUp/Bootstrap changes;
- no KeySwitch/Relinearize/Rotate changes;
- no parameter tuning;
- no application/CNN changes.

Accepted prerequisite:
FAST-OBS-001 at `6f728f19a6300da339541e3f67259e873b9f942e`.

Secondary architecture authority:
`docs/FAST_CKKS_SPEC.md`

Codex should follow the bounded implementation → self-review → one repair pass → final validation workflow, then return `READY_FOR_WEB_REVIEW`.
