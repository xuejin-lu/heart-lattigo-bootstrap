# Current Task

Task: QPREFIX-IMPL-003
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-003-PREFIX-COMPLETE-PRIMITIVE-KERNELS.md`

Task class:
`I — Implementation`

Accepted prerequisite:
- QPREFIX-IMPL-001 at `6553491f9fb9b964c8fd0d743f3a302de83d0b54`.
- QPREFIX-IMPL-002 at `91baa6a4655e10fe2460a399633fa54a03a35318`.

Implement only prefix-complete primitive kernels with explicit validated row counts up to four q rows.

Important transitional rule:
- do not globally activate q2/q3 consumption in production wrappers yet;
- existing wrappers keep their accepted legacy authoritative-row behavior unless the spec explicitly proves otherwise;
- width-4 capability must be exercised by focused synthetic/oracle tests;
- poisoned higher rows must prove no accidental promotion.

Do not modify Rescale, ModUp, ScaleDown, DFT/LinearTransform production routing, EvalMod/PS/DA, Bootstrap orchestration, or `fast-ckks`.

Run focused tests, `go test ./schemes/ckks/fast`, relevant indirect tests, `go test ./...`, `git diff --check`, gofmt.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
