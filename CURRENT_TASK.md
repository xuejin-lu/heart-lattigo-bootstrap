# Current Task

Task: FAST-STORAGE-005
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STORAGE-005-LEVEL0-MODUP-CANONICALIZATION.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisite:
- FAST-STORAGE-004 at `1b9ecd7973505cac1c4a1673a9578260a761950f`.
- Fixed-width-3 initial production policy at Secondary docs commit `d9919f9c080e0dfa731746f5c447f93633ae2f36`.

Implement only the standalone Level-0 private-F ModUp canonicalization boundary:
- require Level 0, width 3;
- canonicalize `X` to `Center_q0(X mod q0)`;
- raise only logical Level metadata to an explicit target level;
- preserve Scale, degree, domain, metadata, and width 3;
- reset per-component bounds to the exact observed canonical magnitude;
- validate by logical-Q export and Rescale -> ModUp chaining.

Do not modify historical production `FastEvaluator.modUpBasis`/`ModUp`, do not wire into Bootstrap, and do not implement Trace, scale alignment, Montgomery conversion, KeySwitch, Relinearize, Rotate, contraction, or adaptive width.

Codex must follow the normal startup sync and bounded implementation -> self-review -> one repair pass -> validation workflow, commit/push Secondary `fast-ckks`, then report `READY_FOR_WEB_REVIEW`.
