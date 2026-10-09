# Independent Web review — FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013

**Disposition:** ACCEPT AS BOUNDED P0 PUBLIC-API NUMERICAL COVERAGE; NOT FAST Q-PREFIX DROP-IN COMPLETION.
**Reviewed:** 2026-10-10 (Asia/Taipei).
**Primary submitted commit:** `bd8993752d39eec748e891edd2d2aeb32f226e0d`.
**Pinned genuine Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Pinned Fast implementation used by 013:** `00ac70ba136d190fa31bbb26c2f51d003a221634`.
**Post-013 Fast constitutional amendment:** `9489925c803be09e49a840e6828b005cdeb1faca`, `AGENTS.md` only.

## Independently verified GitHub facts

- The 013 Primary commit contains only a new runner, its tests and three evidence/report artifacts; it does not modify either Lattigo implementation.
- Inspected `013-summary.md` and the submitted source map. Reported 22 numerical checkpoints and two negative controls on identical frontend source for Standard and Fast; focused Go tests and vet passed *as reported by Codex*, not independently re-executed by Web.
- Existing public `ckks.NewEvaluator` in the Fast fork still embeds `rlwe.NewEvaluator`; it does **not** construct `schemes/ckks/fast.NewEvaluator`.
- Public Fast P0 `Add/Sub`, generic scalar/plaintext multiplication and `Rescale` use full-active-Q paths. Public Fast `MulRelin` for supported ciphertext pairs uses a zero-secret no-key branch; public Fast `Rotate` reaches the shared `fastcore.ApplyRows` with all active Q rows. No claim that ordinary public calls reached optimized compact-Q arithmetic is justified.
- The evidence is scoped to LogN13 P0 with four active Q residues. No Bootstrap, benchmark, or LogN16 execution occurred in 013. The numeric result is not evidence of Q-prefix speed.

## User-confirmed architecture correction (2026-10-10)

The **end goal** is a single unchanged ordinary CKKS frontend. Swapping the Lattigo dependency to Fast shall make the *same public API* execute the already-proven Fast zero-secret numeric algorithms and Q-prefix data policy automatically. Requiring special `fast.NewEvaluator`, API calls, runtime flags, or frontend modifications is not acceptable.

A separate pinned, unmodified original Standard Lattigo checkout supplies the formal correctness/performance baseline. There is **no requirement to retain a parallel Normal/Standard implementation inside the Fast fork solely for comparison**; this is explicitly amended in Secondary `AGENTS.md` at `9489925c...`.

P0 full-active-Q and C0 compact-prefix were accepted in 008/009 only as a **transitional** correctness bridge. Their coexistence is **not** a permanent architecture requirement. The now-authorized priority is to route public calls to the *existing* Fast kernels and their proven Q-prefix lifecycle; do not extend or duplicate the P0 full-Q arithmetic merely to collect more numerical coverage.

## Pre-established mathematical authority

Secondary `docs/FAST_QPREFIX_SPEC.md` governs logical Level/Scale, the actual q-prefix `min(Level+1, 4)` capacity policy, authoritative centered lift `2B < S_Q`, reconstruction for logical-top-`q_Level` Rescale, ModUp, materialization, and dormant-row prohibitions. The existing explicit Fast implementation and accepted tests are technical assets. These are **not open-ended new algorithm-design problems**. A genuinely unsupported operation or provable bound failure should be reported as a counterexample; neither a Standard full-RNS fallback nor weakened proof is acceptable.

The integration barrier to resolve is Go package dependency and ABI: `fast/evaluator.go` imports parent `ckks`; `ckks.NewEvaluator` returns concrete `*ckks.Evaluator`; public and explicit Fast Rotate argument order differs. Direct parent import of child `fast` would create an import cycle. A private import-neutral Fast engine/adapter with reused kernel bodies is a valid bounded integration direction, retaining public signatures.

## Decision

- Accept 013 as a limited historical test result; freeze its evidence.
- Do not commission further P0-only public arithmetic coverage as a substitute for integration.
- Approve next bounded `FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014` charter. It targets one compact-Q public-API arithmetic composition without frontend changes and with proof of actual existing Fast kernel dispatch, not a new CKKS algorithm or immediate full Bootstrapping claim.
- No old Bootstrap or performance experiment should be rerun as part of the initial adapter integration.
