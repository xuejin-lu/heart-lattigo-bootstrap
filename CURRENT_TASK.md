# Current Task

Task: QPREFIX-IMPL-007-REVIEW-FIX-2
Status: READY_FOR_CODEX

Accepted candidate:
Secondary `fast-qprefix` commit
`4c6d7f207ddbaa531eb6d1010ffa6cd8ab08eace`

Review result:
The legacy-vs-explicit polynomial/Mod1 API contract fix is correct.

One final production handoff blocker remains:

`bootstrapping.FastEvaluator.SlotsToCoeffs(...)` now selects `QPrefixWidth`, but the actual `bootstrapCore(...)` does not use that wrapper. It still calls:

`eval.DFTEvaluator.SlotsToCoeffsNew(ctReal, ctImag, eval.S2CDFTMatrix)`

and `dft.FastEvaluator.SlotsToCoeffsNew(...)` still selects legacy authority with `MaintainedLimbCount`.

Therefore full `Bootstrap` / `BootstrapMany` still enters production S2C with legacy q012 authority, despite QPREFIX-IMPL-007 requiring q0123 activation after EvalMod proves q0123 output authority.

Required bounded fix:

1. Change the actual `bootstrapCore` EvalMod->S2C handoff to use explicit authority from the EvalMod output:
   `rows = QPrefixWidth(ctReal.Level())`.

2. Prefer one of:
   - call the already-correct `bootstrapping.FastEvaluator.SlotsToCoeffs(ctReal, ctImag)`; or
   - call `DFTEvaluator.SlotsToCoeffsNewQPrefixRows(..., rows)` explicitly.

3. Do not globally change `dft.FastEvaluator.SlotsToCoeffsNew`; it remains a compatibility/legacy-authority wrapper.

4. Add a production-core regression that exercises the real `bootstrapCore` / `Bootstrap` or `BootstrapMany` call graph and proves S2C consumes q3 after EvalMod.
   A direct helper-only S2C test is not sufficient.

5. Preserve:
   - all existing 007 capacity evidence;
   - legacy polynomial/Mod1 wrapper compatibility;
   - q0123 EvalMod output contract;
   - public generated-secret and zero-secret correctness;
   - no q4+ reads.

No mathematics or schedule changes.

Validation:
- focused real Bootstrap/BootstrapMany S2C authority regression;
- existing QPREFIX-IMPL-007 tests;
- `go test ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/dft ./circuits/ckks/bootstrapping -count=1`;
- `go test ./... -count=1`;
- `git diff --check`;
- gofmt.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.

Do not start QPREFIX-IMPL-008 yet.
