# Current Task

Task: FAST-INTEGRATION-002
Status: READY_FOR_CODEX

Specification:
`specs/FAST-INTEGRATION-002-FUSED-MODUP-BRIDGE-OPTIMIZATION.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisite:
- FAST-INTEGRATION-001 at `31efadc693559217b48d3e76a2e3655b9e6cd14d`.
- Fusion architecture clarification at `dc698e2d99a488f0de2cf4f3207b09ef94521303`.

Implement only the fused Level-0 ModUp production bridge optimization:
- replace the hot-path `ImportLevel0 -> FastStorageModUpLevel0 -> ExportToCompactLogical` chain with a fused canonical q0 -> maintained logical-q conversion;
- preserve exact semantics using the unfused private-F path as the test oracle;
- keep dormant logical rows nil;
- preserve downstream Trace/Montgomery/DFT/EvalMod/packing/Rescale behavior;
- meet the performance gates in the spec.

Do not migrate downstream arithmetic, change CKKS math, delete standalone private-F APIs, add contraction/adaptive width, or broaden into general pooling/refactoring.

Codex must follow the normal startup sync and bounded implementation -> self-review -> at most one repair -> validation workflow, commit/push Secondary `fast-ckks`, then report `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW` according to the performance gates.
