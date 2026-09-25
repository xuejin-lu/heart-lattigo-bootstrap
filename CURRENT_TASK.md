# Current Task

Task: FAST-INTEGRATION-003
Status: READY_FOR_CODEX

Specification:
`specs/FAST-INTEGRATION-003-PRIVATE-F-RESIDENT-MODUP-TRACE.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisites:
- FAST-INTEGRATION-002 at `57ffb88744c82778c0a9392ecab394e19f712a3d`.
- Private-F normalized Trace theorem at `8f823fdb9464c2738f71c30d156ce574098d8605`.

Implement only the first private-F resident production segment:
- Level-0 import into width-3 private-F;
- logical ModUp canonical boundary in F;
- integer scale alignment in F;
- normalized Trace in F using unnormalized automorphism sum followed by exact normalization;
- compact LogicalQ export only after Trace;
- preserve existing Montgomery conversion and downstream C2S/EvalMod/S2C.

Do not pre-multiply private-F Trace by modular gap inverse.
Do not add LogicalQ/full-RNS fallback for capacity failures.
Do not migrate downstream stages or production Rescale.

Codex must follow the normal startup sync and bounded implementation -> self-review -> at most one repair -> validation workflow, commit/push Secondary `fast-ckks`, then report `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
