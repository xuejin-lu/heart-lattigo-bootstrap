# Current Task

Task: FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009
Status: READY_FOR_CODEX
Class: M-approved representation contracts / bounded I shared-core extraction and public Rotate adapter / E numerical check

**Executable spec:** `specs/FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009.md`
**Web review and decision:** `results/FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008-web-review.md`
**Existing-source inventory:** `docs/FAST_DROPIN_EVALUATOR_REUSE_ARCHITECTURE.md`
**Long-term goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Architectural decision

The 008 feasibility report is **accepted** (`REUSE_REQUIRES_REDESIGN` correctly identifies representation mismatch). Web freezes **P0 public full-active-Q** simulation and **C0 explicit compact-Q** Fast contract. Do not implicitly mix. Extract/share the **existing** ring-level Fast Automorphism/Rotate kernel through an import-neutral internal module usable by both ordinary public `*ckks.Evaluator` and explicit `fast.Evaluator` without cycle. Public Fast zero-secret Rotate should process **all Level+1 valid Q rows**, including Level>=4, with zero c1 and zero Galois key lookup, or reject invalid inputs before mutation. Existing explicit Fast callers retain their narrower known-authoritative row policies, their Q-prefix scratch and dormant residue rules. Preserve pinned Standard source, `ckks.NewEvaluator` signature, prior 006A keyless MulRelin safety and external parameters.

**Suspended `FAST-DROPIN-KEYLESS-ROTATE-007` is superseded; DO NOT RUN IT.** 009 is a reuse/extraction task, not permission to invent a second automorphism implementation. Demonstrate two call paths sharing one moved function, with correctness and keyless evidence, plus a single unchanged-frontend Fast regression. No Bootstrap or benchmarks. All constraints, stop rules and deliverables are in the linked spec.

**Secondary baseline:** `93ab7ecc5a864fd9bbacea55f0af730941552691`.
**Pinned Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Formal performance rebaseline:** `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.

Return `FAST_SHARED_AUTOMORPHISM_BRIDGE_COMPLETE_PENDING_WEB_REVIEW`, `FAST_SHARED_AUTOMORPHISM_BRIDGE_PARTIAL` or `FAST_SHARED_AUTOMORPHISM_BRIDGE_BLOCKED` plus `READY_FOR_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.
