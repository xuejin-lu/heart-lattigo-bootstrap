# Current Task

Task: FAST-STORAGE-004
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STORAGE-004-LOGICAL-Q-RESCALE-OVER-PRIVATE-F.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisite:
`FAST-STORAGE-003` at Secondary commit
`b8305a7e3d4ff15591a2249e97a54ad0b3311dde`.

Implement only the standalone one-step private-storage Rescale defined by the spec:
- exact rounded division by logical `q_ell`;
- Level -> Level-1;
- Scale -> Scale/q_ell;
- exact post-Rescale bound propagation;
- unchanged storage width;
- coefficient and NTT private-F support;
- independent Standard-logical oracle validation.

Do not modify historical production `Evaluator.Rescale`, do not wire into Bootstrap, and do not implement contraction, RescaleTo, ModUp, KeySwitch, Relinearize, Rotate, or application changes.

Codex must follow the normal startup sync and bounded implementation -> self-review -> one repair pass -> validation workflow, commit/push Secondary `fast-ckks`, then report `READY_FOR_WEB_REVIEW`.
