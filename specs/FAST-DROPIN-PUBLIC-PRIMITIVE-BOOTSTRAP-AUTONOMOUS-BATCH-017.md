# FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017

**Status:** READY_FOR_CODEX
**Mode:** bounded four-checkpoint autonomous integration experiment, mandatory Web scientific review.
**Prior accepted scientific evidence:** `results/FAST-DROPIN-ABOVE-CAP-NUMERICAL-AUTONOMOUS-BATCH-016-web-review.md`, 015R and 012 reviews.
**Baseline:** untouched genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Fast:** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
**Cost ceiling for ALL A–D:** maximum **ONE Standard + ONE Fast actual Bootstrap**, zero retries, zero benchmarks, zero LogN16, zero parameter sweeps.
**Current authority:** both `AGENTS.md`, Primary `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, Secondary `docs/FAST_QPREFIX_SPEC.md`.

## Single objective

Demonstrate an **unchanged ordinary CKKS public frontend** (no Fast-specific application imports, constructors, flags, manual c1 patches, custom measurement adapters) that performs a **genuine nontrivial composition** of accepted `Add(New)`, degree-one ciphertext `MulRelin(New)`, `Rescale`, `Rotate(New)` and then `bootstrapping.Evaluator.Bootstrap`, with ordinary public key generation and native `DecryptNew/Decode`, under the original E=32 Bootstrap profile. Change **only the Lattigo dependency** between genuine Standard and zero-secret Fast; preserve original E32, actual Q/P, LogN, Scale and the identical frontend/config/input hashes.

This is the next **bounded low-residual-Level** bootstrap integration milestone. Batch 016's Level-5 four-row compact numerical proof used a special measurement-only decoder; **do not** treat that as proof that ordinary full-Q `DecryptNew` accepts Level>3 compact ciphertexts. The current milestone uses a compatible residual input/output profile so native public decryption may be evaluated with full authoritative active rows. No claim of generalized high-Level compact decrypt or fully unrestricted drop-in follows from passing.

## Checkpoint A — exact reuse, key-plan and depth preflight (zero Bootstrap)

1. Safely sync both repos per `AGENTS.md`. Independently inspect accepted 012 runner, exact canonical Bootstrap config/key plan/E32, existing 015 public Add→MulRelin→Rescale→Rotate code, present Secondary `bootstrapping.Parameters.GenEvaluationKeys`, `bootstrapping.NewEvaluator`, public `Bootstrap`, Fast mode, and contract for input Level, scale and expected output. Use the same original baseline pins without editing genuine Standard.
2. Construct a small, mathematically valid CKKS public *pre-Bootstrap* chain in the frozen Bootstrap residual parameter profile (likely low residual Level); verify whether `MulRelin` can be followed by `Rescale` without making Bootstrap input Level/Scale incompatible. Choose source-backed chain order, possibly additional valid operations, and a deterministic nonconstant 16/4096-slot input consistent with canonical 012 supported parameters. Do NOT silently change canonical 012 profile or E32. The goal is **all four named primitive classes plus Bootstrap**, but source-backed unsupported depth must cause STOP/explicit reduced-coverage proposal, not hidden omission.
3. Verify exact Q/P prefixes, N, P-level Galois-key layout, key generation parameters and values, documented Fast key dispatch, and both backend public native lifecycle. E32 must stay visible. Observe current Fast `GenEvaluationKeys` shortcut conditions and never make a new frontend special-case to select Fast keys.
4. Prepare one byte-identical common frontend executable without Fast-only imports; compare source/profile/actual Q/P/input hashes. Fast/Standard module selection may happen through build/workspace configuration only. Keep diagnostic run code separate from public frontend. **Zero Bootstrap calls in A.**
5. If existing 012 profile cannot accommodate the proposed pre-Bootstrap chain, or key layout cannot be established without changing the frozen profile, STOP `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with concrete Level/Scale/key-plan evidence. Do not burn expensive calls.

## Checkpoint B — cheap public primitive preflight (zero Bootstrap)

1. Run one matched public primitive-chain preflight for Standard and Fast with native keygen/Encode/EncryptNew, Add/MulRelin/Rescale/Rotate, native DecryptNew/Decode at supported residual levels. Verify nontrivial cleartext oracle after each operation, finite RMSE/max, exact Level/Scale/slots, proper Standard nonzero c1, Fast zero c1, no hidden full-RNS fallback when rows are dormant, and correct selection of shared existing Fast cores.
2. Confirm Bootstrap input conforms to each backend's public input contract (not merely that preflight decrypted). Record equality of common frontend, parameter, Q/P, E, input hashes and key-plan compatibility; require real residual output of the preflight chain to become **the actual input ciphertext** to public Bootstrap.
3. If expected residual-input Level cannot be achieved, numerical gating fails, key layout differs or genuine Standard fails, stop with exact first cause; no tweaks to params, E, tolerances, secret distribution, backend code, or alternate frontend fast selector.

## Checkpoint C — one-shot public Bootstrap composition

1. Only after A+B pass, execute **at most one Standard and one Fast actual Bootstrap** on the exact preflight-validated chain outputs. No warmup/retry/repeat/benchmark, even if a subprocess fails; journal call budget before invocation, preserving recoverable attempt status.
2. Verify original public `GenEvaluationKeys`, `NewEvaluator`, `Bootstrap`, native public `DecryptNew/Decode` and original identical frontend signatures without Fast-specific calls. Compare expected transformed cleartext, genuinely encrypted Standard, Fast numerical outputs (RMSE/max and SNR when well-defined), E32, Scale, Level, slots, authoritative Q rows, and keys consulted. No security equivalence claim.
3. If a **source-proven missing Public adapter** appears, classify `PUBLIC_BOOTSTRAP_INTEGRATION_GAP` and STOP for Web architecture review **without modifying Secondary in this task**. Do not silently revert to old explicit Fast bootstrap constructor or materialize dormant rows in the hot path.

## Checkpoint D — evidence, review and safe handoff

1. Emit new Primary-only dedicated runner/test and `results/FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017-{summary.md,journal.md,evidence.json}` with exact hashes/backend refs/profile/E/key-layout, numeric oracles, direct paired error/SNR, source-backed dispatch, Bootstrap attempt counter, and first failure if any. Avoid large CSVs, full decoded vectors, secrets, and unnecessary logs.
2. Focused Go tests/vet, `git diff --check`, self-review and ordinary fast-forward commit/push of Primary. Preserve earlier reports and Secondary exactly; genuine Standard remains original and read-only.
3. `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only if the whole unchanged public chain and both Bootstrap results genuinely pass. Otherwise `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with exact test/code/provenance evidence and honest consumed budget. Do NOT continue to 018 autonomously.

## STOP conditions

Any changed or unsupported math/depth/capacity contract; no single valid canonical E32 profile; mismatched public frontend hashes; missing original Standard operations/key plan; hidden Fast-specific frontend or special measurement-only decoder; failing plaintext oracle after one bounded repair; generic/full-RNS fallback on dormant Fast rows; any unsafe Git action; exhausted one-call Bootstrap budget. No Secondary edits or benchmarks in 017.
