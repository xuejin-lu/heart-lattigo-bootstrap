# Current Task

Task: QPREFIX-IMPL-005
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-005-LEVEL0-MODUP-TRACE.md`

Task class:
`I — Implementation`

Accepted prerequisite:
QPREFIX-IMPL-004 at
`18f4da53f03acd065d18550a8f2106725462896f`

Goal:
Migrate the Bootstrap entry boundary
`ScaleDown -> Level-0 ModUp -> scale alignment -> Trace`
to Q-prefix v2.

Critical rules:
- Level-0 canonical midpoint is positive at `r=q0>>1`;
- ModUp materializes exactly `QPrefixWidth(targetLevel)` rows;
- Level >=3 ModUp therefore creates authoritative q0123;
- scale alignment and Trace inside the same ModUp boundary must process the same explicit width;
- final Montgomery conversion covers all authoritative rows;
- ScaleDown remains pre-ModUp legacy-authority in this milestone;
- DFT/C2S remains legacy q012 until QPREFIX-IMPL-006.

Do not modify DFT/LinearTransform production routing, EvalMod/PS/DA, packing/N1-N2, parameters, or `fast-ckks`.

Run all validation and benchmarks required by the spec.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
