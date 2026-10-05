# Current Task

Task: FAST-STANDARD-NUMERICAL-FIX-001
Status: READY_FOR_CODEX

Task class:
`C — Numerical Correctness Repair`

Authoritative spec:
`specs/FAST-STANDARD-NUMERICAL-FIX-001-REMOVE-NORMALIZED-EVALMOD.md`

Architecture decision:
The Fast-only normalized LogN13 EvalMod schedule is a historical workaround and must be removed from the production path.

Fast Q-prefix EvalMod should follow genuine Standard Mod1 mathematics and Scale/Level progression, while using Fast Q-prefix storage/arithmetic primitives.

Do not:
- tune the normalized exponents;
- run another semantic-alignment diagnosis to preserve the normalized algorithm;
- reintroduce full-RNS/Standard fallback;
- remove retained sampled error/noise;
- change Standard production arithmetic.

Current key semantics must be described correctly:
- the large `a` mask contribution is removed/elided;
- the public-key target is approximately `(e_pk, 0)`;
- sampled error is retained.

Use existing Q0123 capacity checks as assertions. If the Standard-equivalent schedule actually violates `2B < S_Q`, stop and report the exact first capacity deficit; do not invent a new workaround.

Required output:
`results/FAST-STANDARD-NUMERICAL-FIX-001-summary.md`

Required completion token:
`FAST_STANDARD_EVALMOD_FIX_READY`

Then:
`READY_FOR_WEB_REVIEW`.
