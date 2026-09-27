# Current Task

Task: QPREFIX-IMPL-008
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-008-PUBLIC-BOOTSTRAP-BOUNDARY.md`

Task class:
`I — Implementation`

Accepted prerequisite:
QPREFIX-IMPL-007 final production handoff at
`74c058ad59655f2a47efcb4faf1cf38324bd6137`

Review result for QPREFIX-IMPL-007:
- PASS;
- legacy polynomial/Mod1 wrappers preserve MaintainedLimbCount authority;
- explicit EvalMod path carries q0123 from Level 12 to Level 4;
- real bootstrapCore now uses q-prefix-aware S2C;
- strict capacity, public correctness, full regression, and performance guards pass.

Goal:
Integrate the public Bootstrap boundary:
- packing/unpacking;
- N1<->N2 ring-degree conversion;
- BootstrapMany;
- finalization/public output.

Critical rules:
- public Residual MaxLevel remains <=1, so public authority is q0/q01;
- internal q0123 is created only after ModUp and must not leak into or be required by public input/output;
- ring-degree conversion must gain explicit 1..4-row capability because the current generic coefficient mapper silently caps at two rows;
- legacy wrappers remain legacy-authority;
- production BootstrapMany passes explicit public rows and must never infer authority from backing;
- no full-RNS fallback.

Do not change circuit mathematics, schedules, parameter chains, or `fast-ckks`.

Run all validation and benchmarks required by the spec.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
