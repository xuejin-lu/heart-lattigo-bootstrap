# Current Task

Task: QPREFIX-IMPL-004-REVIEW-FIX
Status: READY_FOR_CODEX

Accepted implementation candidate:
Secondary `fast-qprefix` commit
`18f4da53f03acd065d18550a8f2106725462896f`

Review result:
Most of QPREFIX-IMPL-004 is correct:
- q0123 fixed-width CRT/Rescale capability is present;
- public Rescale remains legacy-authority activated;
- current C2S/DFT q3 poison regression passes;
- SameLift, transactionality, logical q_ell divisor, and full regression all pass.

One semantic blocker remains:

`DropLevelCanonical` currently accepts targetLevel > 3, but q0123 does not determine
the canonical centered representative modulo the full logical Q_target when q4+ are part
of that modulus.

Revised authoritative spec:
`specs/QPREFIX-IMPL-004-RESCALE-LEVEL-TRANSITIONS.md`
at Primary commit:
`a7bf5f2c68662f66c7b955b75a566e5eff4e0816`

Required fix only:
- `DropLevelCanonical` must explicitly reject targetLevel > 3;
- add focused regression(s), e.g. 5->4 canonical request rejects transactionally;
- preserve existing successful 3->2, 2->1, 1->0 canonical tests;
- do not change SameLift;
- do not expand scope.

Validation:
- focused level-transition tests;
- `go test ./schemes/ckks/fast -count=1`;
- `go test ./circuits/ckks/dft -count=1`;
- `go test ./... `;
- `git diff --check`;
- gofmt.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
