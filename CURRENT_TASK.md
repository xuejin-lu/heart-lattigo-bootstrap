# Current Task

Task: BUILD-ISO-001
Status: READY_FOR_CODEX

Specification:
`specs/BUILD-ISO-001-STANDARD-FAST-LINUX-BUILD-ISOLATION.md`

Task class:
`I — Implementation`

Scope:
Primary-only build isolation so the same benchmark frontend can produce:
- `bootstrap-standard-linux-amd64` using `../lattigo@main`
- `bootstrap-fast-linux-amd64` using `../lattigo@fast-ckks`

This task must not modify:
- research architecture;
- mathematics;
- Fast-CKKS arithmetic or storage logic;
- CKKS/Bootstrap parameters;
- benchmark/experiment semantics;
- Secondary source.

Use the minimal build-tag + Standard-stub + build-script solution defined in the spec.

Codex should follow the bounded implementation -> self-review -> one repair pass -> final validation workflow, commit/push Primary only, then return `READY_FOR_WEB_REVIEW`.
