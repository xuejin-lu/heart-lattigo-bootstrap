# Fast CKKS drop-in evaluator — reuse-first architecture inventory

**Date:** 2026-10-09
**Status:** SOURCE-VERIFIED INVENTORY / proposed integration direction, **NOT an authorized code rewrite**.
**Fast source:** `xuejin-lu/lattigo` `fast-qprefix@93ab7ecc5a864fd9bbacea55f0af730941552691`.
**Standard source (immutable reference):** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Authority:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`; Secondary `docs/FAST_QPREFIX_SPEC.md`; `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A.

## Goal, and what was misdiagnosed

The fixed end-goal is one **identical frontend CKKS source and all identical public API calls and parameter values**, switching only Lattigo library dependency. In the Fast dependency, internal zero-secret numerical simulation should use existing optimized Fast computation rather than native key switching where mathematically valid. This is **not** secure encryption or a claim that Fast equals Q-prefix.

Previously, the ordinary public `ckks.NewEvaluator` not selecting the optimized evaluator was interpreted as an absent MulRelin/Rotate implementation. This was wrong: the core algorithms already exist under `schemes/ckks/fast`. Implementations 006/006A established a useful bounded keyless full-Q-authoritative public MulRelin and fail-closed safeguards, but **do not establish that the prior Fast Q-prefix kernel is reused**. The proposed one-off public Rotate implementation (007) was suspended **before coding** to avoid further duplication.

## Source-backed capability inventory

| Capability | Ordinary public API in Fast fork | Existing explicit Fast implementation | Gap |
|---|---|---|---|
| Evaluator construction | `schemes/ckks/evaluator.go:16-31`: concrete `*ckks.Evaluator` embeds `*rlwe.Evaluator` | `schemes/ckks/fast/evaluator.go:18-50`: separate concrete `*fast.Evaluator` with own scratch/index cache | `INTEGRATION_DISPATCH`, incompatible concrete return types |
| Add/Sub | Native `ckks.Evaluator.Add`; full-active-Q ordinary arithmetic | `schemes/ckks/fast/evaluator_ntt.go:22-183`, maintained prefix path (comments identify q0/q1 default) | dispatch + input representation and operand overload differences |
| MulRelin | Fast fork already has keyless public pilot `schemes/ckks/evaluator_fast_zero_secret.go`, but full-active-Q via generic multiply kernel | `schemes/ckks/fast/evaluator_ntt.go:209-323`: distinct keyless Fast prefix implementation, plus `MulRelinElementQPrefixRows` in `fast/evaluator.go` | **reuse of preexisting optimized kernel not proven**; 006A safety constraints retained |
| Rescale | ordinary public native CKKS rescale; 005/006A reported its numerical correctness on full-Q inputs, not fast dispatch | `schemes/ckks/fast/rescale.go:138-188`: prefix-aware exact logical top-q division and maintained-row authority | `REPRESENTATION_CONTRACT` and dispatch; do not assume it is interchangeable with full-RNS |
| Rotate | `schemes/ckks/evaluator.go:1190-1210`: public `Rotate` invokes native RLWE `Automorphism` and Galois KeySwitch | `schemes/ckks/fast/evaluator.go:239-358`; `fast/automorphism.go`: keyless Fast automorphism/rotation, separate method argument order; existing default can process limited maintained rows | `INTEGRATION_DISPATCH` plus all-active-Q/compact-Q authority mismatch |
| Fast ciphertext storage | `ckks.NewCiphertext` and public secret-key `EncryptNew` currently return full-backed zero-c1 simulated ciphertext | `schemes/ckks/fast/ciphertext.go` allocates logical Level with **only prefix backing** and nil dormant rows | **representation gap**; cannot send compact ciphertext into any Standard full-RNS op |
| Public bootstrap | `circuits/ckks/bootstrapping/evaluator.go` contains internal `fast *FastEvaluator` compatibility construction path | existing dedicated optimized Bootstrap engine | useful precedent for API-preserving internal delegation, **not proof** that arithmetic package can import fast |
| Key and safety | `KeyLayoutFast` rejected by normal `rlwe` KeySwitch; 006A full-Q MulRelin fails closed on unsupported pair | dedicated Fast keys/zero-secret semantics | must never conceal native KeySwitch fallback or assume layout interoperability |

## Non-negotiable package/API facts

- `fast/evaluator.go` imports `schemes/ckks`; importing sibling `schemes/ckks/fast` from `schemes/ckks` would form a **Go import cycle**.
- Public `ckks.NewEvaluator(parameters, evk)` returns concrete `*ckks.Evaluator`. A factory returning `*fast.Evaluator` would break existing type contracts, even if operations had similar names.
- Explicit `fast.Evaluator.Rotate(ctIn, ctOut, k)` vs public `ckks.Evaluator.Rotate(ctIn, k, ctOut)` is a real signature mismatch. The explicit Fast evaluator intentionally lacks methods needed for the generic scheme evaluator interface (`fast/evaluator_ntt.go:13-20`), to prevent implicit native KeySwitch fallback.
- Secondary Q-prefix v2 spec defines maintained width `w_Q(ell)=min(ell+1,4)`, and logical Level remains unchanged. Some currently exported Fast ordinary kernels declare q0/q1 maintained authority; this **must be reconciled with per-producer actual active rows and strict `2B<S_Q` capacity rule** before any automatic use.
- Public zero-secret EncryptNew of the Fast fork outputs logically valid `(c0,0)` with full backing in the tested secret-key lifecycle; **that does not automatically prove conversion into compact Q-prefix** is safe. A full-length row can still be stale if an earlier Fast kernel did not update it.

## Architecture alternatives

**A. Direct `ckks` → `fast` import / direct concrete evaluator swap.** REJECT: cyclic Go package dependency and public signature incompatibility.

**B. Registration in `fast.init()` or frontend-specific `import _ ".../fast"`.** REJECT as a default: the frontend is not permitted to import Fast, and there is no guarantee the package is linked into a standalone ordinary CKKS caller. The existing Bootstrap package importing Fast cannot be assumed by every frontend.

**C. Continue hand-writing a separate version of each Fast operation inside `ckks`.** REJECT as project default: duplicates existing kernels, risks divergent Q-prefix authority, breaks reuse-first gate. A narrowly proven specialized exception requires explicit Web approval.

**D. Move/extract reusable, **import-neutral** Fast kernel/state machinery into a shared lower-level CKKS-internal package, preserving wrappers for (i) existing `fast.Evaluator` users and (ii) the ordinary `*ckks.Evaluator` Fast-only adapter.** **PREFERRED TARGET, CONDITIONAL ON TRANSITIVE DEPENDENCY AND REPRESENTATION AUDIT.** A prospective location is `schemes/ckks/internal/fastcore`, importable by siblings under `schemes/ckks` but **must not import either parent `schemes/ckks` or child `schemes/ckks/fast`**. Any needed parameter data must come from a safe ring/RLWE-level provider or import-neutral types; scratch, keyless/NTT operations, and prefix preconditions must be proved reusable, not recopied. This is an architecture *hypothesis* rather than authorization to create/move that package.

**E. Invert package layering by relocating all Fast logic into `schemes/ckks` and retaining `fast` facade.** Possible but potentially a broader refactor; only consider if D is impractical under inspected dependencies, with a costed comparison. Must not collapse established Bootstrap/test call sites.

## Required pre-implementation proof

1. A complete **transitive import and helper dependency map** for one representative existing Fast core operation (initially Rotate/Automorphism because public 007 is paused) and for Rescale/MulRelin at API-signature level. Identify symbols requiring `ckks.Parameters`, `fast.Evaluator` internal scratch, `fast.NewCiphertext`, and producer-authoritative Q rows; identify whether an import-neutral extraction is actually small.
2. Build a **contract matrix** for all-active-Q and compact Q-prefix at Level=0/1/2/3 and Level>=4, recording actual maintained versus required source/destination rows and explicit failure/materialization boundaries. Treat actual current code, not historical assumptions, as decisive.
3. Show compile-valid Go package directions / exact provider interfaces without changing the public `*ckks.Evaluator` constructor signature, with a **reuse claim tested by execution instrumentation rather than only a green oracle**.
4. Constrain the first integration to one source-proven representative algorithm and select an already existing kernel, not write a substitute. Preserve original explicit `fast` callers and key-layout guards. A feasibility contradiction must be reported, not masked with full-RNS fallbacks or interface widening.
5. Consider a dispatch smoke test that proves the same existing kernel is invoked through both entrypoints, verifies correct error behavior on compact/stale input and proves ordinary Standard stays unaffected. **Do not run performance benchmarks until correctness is established.**

The next action is **read-only feasibility task 008**. Do not execute suspended `007`, modify algorithm kernels or build an import-neutral extraction before 008's evidence is independently reviewed.
