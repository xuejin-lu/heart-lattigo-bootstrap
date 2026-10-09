# Current Task

Task: FAST-DROPIN-KEYLESS-MULRELIN-006
Status: READY_FOR_CODEX
Task class: M-approved narrow invariant / bounded I implementation and E preflight

**Executable spec:** `specs/FAST-DROPIN-KEYLESS-MULRELIN-006.md`
**Independent Web review of 005:** `results/FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005-web-review.md`
**Long-term goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Web review and accepted boundary

`FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005` was accepted as `PARTIAL`, not as proof of optimized Fast primitive execution. The same unchanged public CKKS frontend passed Add/MulRelin/Rescale/Rotate numerical oracles with native `ckks.NewEvaluator` in both builds. Fast Zero-Secret c1 remained zero, but all four checkpoints used generic CKKS arithmetic; native KeySwitch rejects Fast-layout keys and the dedicated `schemes/ckks/fast` evaluator was never selected. Evidence: Primary report `9b48bfbb1096b41827bed3d280bdb3163726b60e`, Fast Secondary `eb7793f1e449702590cdba3d9603261623354f80`.

## New bounded implementation

Implement just **one real optimized operation** behind the existing public `*ckks.Evaluator`: Fast-only **keyless Zero-Secret MulRelin/MulRelinNew** for fully materialized logical Q limbs and freshly simulated c1=0 degree-one inputs. Algebraically `(a0,0)*(b0,0)=(a0*b0,0,0)` so no RelinearizationKey or GadgetProduct is required. Preserve all active Q rows, metadata/Level/Scale/degree, correct NTT/Montgomery behavior, aliasing safety and input guards. Never zero normal nonzero-c1 ciphertexts. Do not change public signatures, genuine Standard, sibling fast.Q-prefix evaluator, other schemes or KeyLayoutFast safety guards. Use the same prior frontend source `tools/fast-dropin-ckks-primitive-api-audit-005/main.go`, source SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`, unchanged in both dependency runs, only if source-confirmed safe.

Required: targeted tests including independent proof that no relin key is *accessed* on Fast, negative nonzero-c1 safety, then an inexpensive same-source Standard/Fast oracle comparison. **Zero Bootstrap calls, zero benchmarks, no tuning**, no general drop-in claim.

Write `results/FAST-DROPIN-KEYLESS-MULRELIN-006-summary.md`; report `FAST_DROPIN_KEYLESS_MULRELIN_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYLESS_MULRELIN_PARTIAL` or `FAST_DROPIN_KEYLESS_MULRELIN_BLOCKED` and Web review status. Safe ff-only commits/pushes per AGENTS; stop for architectural ambiguity.

`FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.
