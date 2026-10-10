# FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019

**Status:** READY_FOR_CODEX. **Mode:** bounded four checkpoints with Web review.
**Authority:** Primary + Secondary AGENTS.md, Primary research workflow §4A/4B and end-goal; Secondary FAST_QPREFIX_SPEC.md.
**Predecessor:** `results/FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018-web-review.md`.
**Genuine Standard pinned:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` untouched.
**Fast pinned:** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
**Batch cost ceiling:** max 1 Standard and 1 Fast Bootstrap only AFTER cheap matched preflights, zero retry/benchmark/LogN16/sweeps. Keep Secondary and genuine Standard read-only.

## Specific question

Can the **same original public CKKS frontend** under the exact frozen canonical E32 Batch017/018 generated Q/P profile chain successfully compose genuinely above-cap Level5 `Add → MulRelin → Rescale(q5) → Rotate → (mathematically justified public Level drop to canonical residual 0) → E32 Bootstrap → native DecryptNew/Decode` in genuine Standard and intentionally insecure Fast merely by switching dependency?

Batch016 proved Level5 compact MulRelin/Rescale numeric results under a different six-Q profile and diagnostic decoder. Batch018 proved same frozen E32 Level5 **Add/Rotate only**, justified `DropLevelNew(5)`, then public Bootstrap. This 019 batch must not conflate those proofs.

## Mathematical stop gates (A before code/test)

1. Safely sync current repositories; read source, 016, 017, 018 reviews/results/runners, exact canonical E32 full Bootstrap Q chain. Preserve Q/P, LogN13, E=32, default Scale2^45 and original public signatures; Standard is genuine encrypted, Fast zero-secret insecure.
2. Derive the **exact planned CKKS Scale after a Level5 ciphertext×ciphertext MulRelin and logical-top-`q5` Rescale**: for both input Scales default 2^45, after Rescale it is `2^90/q5`; q5 is ~2^60 in this canonical profile, so this is ~2^30 and **not the required 2^45**. Do not conceal or tune this difference. Examine whether a legitimately chosen *per-input* CKKS Scale (as already used for C in Batch017), an approved mathematically valid additional Rescale, or existing normal public `SetScale` can meet the same canonical Level0 / Scale2^45 constraint without altering original parameter/default scale. Calculate the independent propagation bounds and precision implications before considering a run. If no justified configuration exists within original frozen profile and supported public operations, STOP `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with exact scale/capacity evidence, no Bootstrap.
3. For the exact public chain chosen, prove source-backed output authority and conservative bounds for **all** intermediate operations in Q-prefix q0..q3 and at intended final q0 after DropLevel. Multiplication bound must account for N=8192 negacyclic convolution and zero-c1; `Rescale` must use true logical `q5` even though dormant q5 has no backing. Require `2B<S_Q` at each stage and `2B<q0` at terminal Level0. A DropLevel is NOT Rescale, not a magic Scale repair, and not permissible without numeric bound. Check actual Q/P and bootstrapping key domain; no hidden full-RNS fallback or Fast-only application API.
4. Reuse existing Fast kernels and previous bounded evidence; no invention of CRT/mul/rescale math, no modification of Secondary, no parameter sweeping. A negative feasibility determination is a valid and required science outcome.

## B — cheap identical-frontend public arithmetic preflight (only if A passes)

1. Build a dedicated Primary-only runner/test from previously accepted source, with identical public source and keys for Standard and Fast (module/tag swaps only); native full-profile `EncryptNew`, public `AddNew`, `MulRelinNew`, `Rescale`, `RotateNew`, appropriate **proven** DropLevel to Level0; native decrypt/oracle only at residual Level0 (and later Bootstrap output Level1).
2. Verify source/config/QP/input SHA identity, actual Level/Scale chain, correct plaintext numerical oracle and Fast-vs-Standard direct difference (finite nonempty full slots), compact four-row Fast operations above cap, zero-c1, no q4+ read, logical q5 divisor, capacity observer and independently bounded input/progression. Existing authorized diagnostic adapter may be used for measurements off to the side only, never application path.
3. If any mismatch / insufficient capacity / wrong scale / hidden fallback or generic CKKS incompatibility exists, STOP without Bootstrap or backend edits. No broadened tolerance to force PASS.

## C — one-shot Bootstrap (only if A and B succeed)

1. Journal call count before each invocation; perform max **1 Standard + 1 Fast actual public Bootstrap**, with original E32 and exact q0 residual Scale2^45 / same keyset as prior 018, on the *held* preflight output. No retries, extra calls or changes after spending quota.
2. Native `DecryptNew/Decode` output and cleartext oracle, direct paired RMSE/max/SNR, output Level/Scale, rows, key accesses and provenance; fixed prior `1e-6` max-error gate. Document genuine Standard vs intentionally insecure Fast numeric semantics.

## D — compact evidence and handoff

1. Commit a bounded Primary runner/test only if mathematical feasibility allows and compact `results/FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019-{summary.md,journal.md,evidence.json}`; if A blocks, record exact scale/capacity/source stop and **0 Bootstrap** rather than fabricated numeric output.
2. Focused go test/vet/diff checks and self-review; commit and ordinary fast-forward push Primary, leave Secondary and Standard pinned clean and unmodified.
3. Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only on all gates; else `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with precise first cause and call budget. Do NOT autonomously start 020.

**Scope:** This 019 experiment still does not establish all public CKKS overloads, high-level compact native DecryptNew, performance/speedup, or cryptographic security. Do not claim more than the tested chain.
