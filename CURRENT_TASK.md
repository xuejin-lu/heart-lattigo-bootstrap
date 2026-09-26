# Current Task

Task: QPREFIX-IMPL-004
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-004-RESCALE-LEVEL-TRANSITIONS.md`

Task class:
`I — Implementation`

Accepted prerequisites:
- QPREFIX-IMPL-001 at `6553491f9fb9b964c8fd0d743f3a302de83d0b54`.
- QPREFIX-IMPL-002 at `91baa6a4655e10fe2460a399633fa54a03a35318`.
- QPREFIX-IMPL-003 at `c8b591a30c05a2261de8d0181d7b3f64bc58169b`.

Implement only Q-prefix-v2 Rescale and explicit Level-transition semantics:
- q0123 centered reconstruction;
- Rescale source width = QPrefixWidth(level);
- divisor always logical q_level;
- target prefix shrinks naturally with new Level;
- explicit SameLift vs Canonical DropLevel APIs;
- strict target-capacity checks and transactional failure.

Only Rescale/DropLevel become prefix-policy authoritative in this milestone.
Do not globally activate q0123 in other production wrappers.

Do not touch ModUp/Trace, DFT/LinearTransform, EvalMod/PS/DA, Bootstrap orchestration, `fast-ckks`, or introduce F.

Run focused tests, accepted C2S Rescale regression, `go test ./schemes/ckks/fast`, relevant DFT/Bootstrap tests, `go test ./...`, `git diff --check`, gofmt.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW` if a real semantic ambiguity is found.
