# FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010

**Status:** READY_FOR_CODEX
**Task:** Bounded E system-integration test, **no Secondary production modifications**.
**Authority:** `results/FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009-web-review.md`; `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`.
**Fast pinned:** `00ac70ba136d190fa31bbb26c2f51d003a221634`. **Genuine Standard pinned:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

## Scientific purpose

009 proved shared Fast Automorphism core through public Rotate/RotateNew under P0, while 004 separately proved an ordinary public E32 Bootstrap lifecycle. Test whether **the same unchanged ordinary CKKS frontend** can perform a depth-free public key generation / EncryptNew → Rotate → Bootstrap → DecryptNew sequence on both genuine Standard and Fast, using the **canonical 004 LogN13 E32 Bootstrap profile**, without special Fast constructors, hand-built ciphertexts, parameters altered per backend, or a native KeySwitch fallback in Fast.

This is a deliberately narrow interoperability smoke test. It does not imply general Rescale/MulRelin→Bootstrap support, compact-Q compatibility, CNN, overall speedup or secure Fast encryption.

## Gate before expensive operations

1. Safely synchronize Primary main and Secondary fast-qprefix per AGENTS and confirm clean/pinned refs. Read prior 004 source/config and 009 results, 009 web review, Secondary Fast Q-prefix spec.
2. Check the actual bootstrap residual Level and pre/post Rotate contracts. Prior 004 profile has residual max Level 1, input Bootstrap at Level 0. The new shared public P0 Rotate supports Level 0 via explicit `rows=1` **only if current source proves that all parameter/domain conditions are valid**. Verify source-based actual `params.GaloisElement(k)`, CKKS Encode/EncryptNew/Rotate/DecryptNew validity, and that Bootstrap consumes the rotated result. No experimental bypass of failing preflight.
3. The frontend should be one new Primary `tools/` command, compiled **unchanged** under pinned Standard and Fast dependency workspaces. It may reuse 004 config parser/helpers by non-duplicative source-level refactoring **only if this does not edit or change the frozen 004 artifact**; otherwise a small isolated test harness is appropriate. Use identical LogN13/E32 actual Q/P primes, slots, Scale, config and input SHA, and a nonconstant complex-valued vector so a wrong rotation fails. Only library workspace selection differs.
4. Generate genuine native evaluation keys from public API in Standard and Fast according to the **same frontend code**. Use public `ckks.NewEvaluator`, `RotateNew`, public `bootstrapping.NewEvaluator`, and public `Bootstrap`. No manual ciphertext-component mutation, no explicit Fast namespace import, no zero-secret constructor in frontend. Standard uses genuinely encrypted native c1!=0 input; Fast zero-secret c1 must remain zero.
5. **First run separate preflight-only mode** under each dependency: Encode → EncryptNew → Rotate (+ optionally Add with same Scale, if source-proven) → DecryptNew/Decode; inspect Level, Scale, degree, complete active Q backing, c1 counts, relative slot oracle, and representation/domain. In Fast the Rotate call must exercise the common `fastcore` without any native GaloisKey lookup; cite source and if feasible instrument via focused package test, not by changing frontend. If preflights fail, stop with exact first mismatch and **zero Bootstrap**. Do not adjust input/profile or substitute Standard result.
6. Only if both preflights pass, run **at most one genuine Standard Bootstrap and one Fast Bootstrap** on their pre-rotated input using identical frontend and canonical 004 profile (E32). Record preflight and post-Bootstrap decoded complex RMSE, max error, SNR vs the correct rotated cleartext oracle; directly compare decoded Standard vs Fast values if full arrays are available. Require Fast output c1=0; Standard native c1 remains governed by its ordinary encryption, not forced zero. Do not invent a universal accuracy threshold; report failures honestly.
7. Preserve exact frontend/config/input SHA, pinned repository refs and worktree flags, generated Q/P prime hashes, active Levels/Scale/slot dimensions, ciphertext c1 statistics, key/layout provenance, call counts and errors. Prefer a compact committed JSON evidence artifact under `results/` if size is reasonable (no secrets), as prior temporary full outputs were unreviewable remotely. At minimum commit complete relevant metadata, metrics and output array hash in report.
8. No benchmarks, warmups, retries, performance comparison, LogN16, CNN, arbitrary fallback, or Secondary production edits. If a source inconsistency or unavoidable failure requires an architecture decision, stop and return `NEEDS_WEB_REVIEW` without modifying Fast code.

## Deliverables

- Primary new shared frontend and `results/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010-summary.md`, committed/pushed after Codex's bounded self-review. Only Primary may change.
- Classification: `FAST_DROPIN_COMPOSITION_COMPLETE_PENDING_WEB_REVIEW` when both preflights and bounded Bootstrap runs completed and numerical comparisons recorded; `FAST_DROPIN_COMPOSITION_PARTIAL` for a correctly bounded preflight but no complete Bootstrap; or `FAST_DROPIN_COMPOSITION_BLOCKED` for an architectural contradiction. Return `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
- Preserve accepted P0 full-Q/C0 compact-Q boundaries, no false general end-to-end CKKS compatibility claim. `FAST-STANDARD-PERF-REBASELINE-003` remains blocked.
