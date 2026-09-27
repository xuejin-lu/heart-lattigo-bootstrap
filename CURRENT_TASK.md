# Current Task

Task: QPREFIX-IMPL-006
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-006-LINEARTRANSFORM-DFT-C2S-S2C.md`

Task class:
`I — Implementation`

Accepted prerequisite:
QPREFIX-IMPL-005 at
`08b36b0b730dcc57eb98466594796c10bffbfdb9`

Review result for QPREFIX-IMPL-005:
- PASS;
- midpoint rule corrected;
- ModUp materializes exact QPrefixWidth;
- q0123 remains coherent through scale alignment, Trace, and Montgomery conversion;
- ScaleDown remains pre-ModUp legacy-authority;
- full regressions pass;
- focused performance guard passes.

Goal:
Migrate LinearTransform and DFT/C2S to explicit Q-prefix execution.

Critical activation rule:
- production C2S MUST now preserve q0123 from the QPREFIX-IMPL-005 ModUp boundary;
- S2C must become q-prefix-capable, but production S2C must NOT assume q0123 until EvalMod (QPREFIX-IMPL-007) provides it;
- capability and production activation remain separate.

Do not migrate EvalMod/PS/DoubleAngle, packing/N1-N2, parameters, or `fast-ckks`.

Run all validation and benchmarks required by the spec.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
