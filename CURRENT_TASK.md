# Current Task

Task: FAST-STORAGE-008
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STORAGE-008-C2S-WIDTH2-EXPERIMENT.md`

Task class:
`E — Experiment / bounded implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: experiment implementation target.

Accepted prerequisites:
- FAST-STORAGE-007 at `1a8018efda159f1dc9ab7078ce80a40c6db28e71`.
- Explicit stage-local width experiment rule at `04396c94a3bb063e93b00bb80a500dfcc2891e21`.

Implement only the standalone C2S stage-width experiment:
- explicit safe 3->2 contraction;
- width-parameterized private-F plaintext mirrors and LinearTransform;
- real LogN13 width-2 and width-3 C2S factor-chain capacity checks;
- exact semantic comparison against logical Fast;
- factor-0 and full four-factor chain benchmarks for Logical Fast / F3 / F2.

Production Bootstrap/C2S routing and the fixed production width-3 policy must remain unchanged.

Classification:
- width-2 capacity blocked;
- width-2 valid but not competitive;
- or width-2 promising,
according to the spec.

Commit/push Secondary and report `READY_FOR_WEB_REVIEW`, unless an architecture/math conflict requires `NEEDS_WEB_REVIEW`.
