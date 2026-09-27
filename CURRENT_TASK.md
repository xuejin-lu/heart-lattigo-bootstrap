# Current Task

Task: QPREFIX-IMPL-007-REVIEW-FIX
Status: READY_FOR_CODEX

Accepted implementation candidate:
Secondary `fast-qprefix` HEAD
`980be10e7ad8d14df4541526c05b4bcb44f17996`

Review result:
QPREFIX-IMPL-007 is mathematically and operationally accepted except for one API-contract regression.

Accepted parts:
- q0123 authority propagates through EvalMod Level 12 -> 4;
- polynomial workspace / generated powers / PS baby+giant / guard / DoubleAngle all use explicit rows;
- accepted P93 schedule semantics are unchanged;
- all reported strict capacity checkpoints pass;
- one-bit q0123 contraction oracle passes;
- final EvalMod output establishes q0123 authority at Level 4;
- production S2C activation from that proven output is correct;
- public correctness/full regression/performance guards pass;
- Q-prefix capacity observer has zero reconstruction cost when disabled.

Review blocker:
The exported legacy polynomial/Mod1 entry points were changed to infer authority with `QPrefixWidth(Level)`.

This violates the migration rule:
`kernel capability != production activation`.

A caller that still owns only legacy q012 authority may have an allocated/stale q3 row. The legacy exported API must not promote that row merely because backing exists.

Required bounded fix:

1. Restore legacy authority selection for exported polynomial wrappers:
   - `polynomial.FastEvaluator.Evaluate`
   - `EvaluateWithPlanScale`
   - `EvaluateWithPlanScaleFinalParentOneBitScalarGuard`
   or preserve their pre-007 equivalent contract.

2. Add explicit-row polynomial entry points (names are coding choice), e.g.:
   - `EvaluateQPrefixRows(..., rows)`
   - `EvaluateWithPlanScaleQPrefixRows(..., rows)`
   - guarded explicit-row equivalent.

3. Restore legacy authority behavior for exported `mod1.FastEvaluator.EvaluateNew`.

4. Add an explicit-row Mod1 entry point, e.g.:
   `EvaluateNewQPrefixRows(ct, rows)`.

5. Production Bootstrap `FastEvaluator.EvalMod` must compute/own:
   `rows = QPrefixWidth(ctIn.Level())`
   and call only the explicit-row Mod1 path.

6. The explicit-row path must retain all currently accepted 007 behavior:
   - q0123 P93 authority;
   - capacity checkpoints;
   - guard;
   - DoubleAngle;
   - final q0123 output;
   - production S2C activation.

7. Add compatibility poison regressions:
   - legacy polynomial wrapper: q3-only poison must not affect q012 result / must not become authority;
   - explicit rows=4 polynomial path: q3-only poison must affect q3 result;
   - legacy Mod1 `EvaluateNew`: q3-only poison must not affect q012 semantic result;
   - explicit Mod1 path used by Bootstrap: q3 is consumed and propagated.

Do not redesign mathematics, schedule, capacity logic, S2C logic, or any other subsystem.

Validation:
- focused polynomial and Mod1 compatibility tests;
- existing QPREFIX-IMPL-007 capacity/oracle tests;
- public Bootstrap generated-secret and zero-secret controls;
- `go test ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/dft ./circuits/ckks/bootstrapping -count=1`;
- `go test ./... -count=1`;
- `git diff --check`;
- gofmt.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
