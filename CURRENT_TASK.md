# Current Task

Task: QPREFIX-IMPL-007
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-007-EVALMOD-PS-DOUBLEANGLE.md`

Task class:
`I — Implementation`

Accepted prerequisite:
QPREFIX-IMPL-006 at
`2aaec605951b4469cef6db10c28453422204595e`

Review result for QPREFIX-IMPL-006:
- PASS;
- production C2S now preserves q0123 authority;
- strict q0123 C2S capacity checkpoints pass;
- direct and BSGS explicit-row LinearTransform paths pass;
- S2C is q-prefix-capable but production remains on current EvalMod q012 authority;
- full regressions and performance guard pass.

Goal:
Migrate EvalMod, Paterson–Stockmeyer polynomial evaluation, one-bit guard, and DoubleAngle to explicit Q-prefix authority.

Critical rules:
- accepted P93 EvalMod enters at Level 12 and exits at Level 4, so q0123 authority must remain continuous throughout;
- preserve the accepted polynomial/PS/scale/DoubleAngle schedule;
- do not interpret historical Q012 schedule names as a three-row storage requirement;
- migrate every workspace/power/PS/guard/Rescale operation that mutates the EvalMod state;
- collect local EvalMod/PS/DoubleAngle strict capacity evidence;
- only after EvalMod output q0123 is proven may production S2C be activated with rows=4.

Do not redesign the polynomial approximation or schedule.
Do not modify packing/N1-N2/public boundary integration.
Do not introduce F/full-RNS fallback or modify `fast-ckks`.

Run all validation and benchmarks required by the spec.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
