# FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008

**Status:** READY_FOR_CODEX
**Task class:** M / architecture evidence collection, **READ-ONLY** (no production edits and no numerical performance experiments).
**Required context:** `docs/FAST_DROPIN_EVALUATOR_REUSE_ARCHITECTURE.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A, `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`, Secondary `docs/FAST_QPREFIX_SPEC.md`.
**Fast Secondary baseline:** `93ab7ecc5a864fd9bbacea55f0af730941552691`.
**Genuine Standard reference:** pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

## Why this replaces 007

The explicit `schemes/ckks/fast.Evaluator` already contains keyless Rotate, MulRelin, Rescale and Add, but the normal public `ckks.NewEvaluator` returns a concrete native CKKS evaluator and never directly invokes the explicit Fast kernels. Implementation 006A added a safe limited keyless full-Q public MulRelin without reusing the older Fast prefix implementation. Repeating that approach for Rotate would duplicate kernels and potentially violate the different full-Q and compact-Q authority contracts. **The suspended task `FAST-DROPIN-KEYLESS-ROTATE-007` MUST NOT RUN.**

## Objective

Determine whether the preferred shared, import-neutral kernel module can be formed by **extracting/moving actual existing Fast code**, avoiding both an import cycle and a frontend Fast-only constructor; determine the smallest safe integration slice. Return an evidence-based recommendation and **no production code**.

### Exact audit

1. Safe-sync both repositories per AGENTS, verify clean worktrees and source SHA. Read current task/spec and the 3 authoritative documents.
2. For existing Fast **Rotate / Automorphism**, trace `schemes/ckks/fast/evaluator.go`, `fast/automorphism.go`, any `fast/qprefix`/helpers, `schemes/ckks/evaluator.go`, `core/rlwe/evaluator_automorphism.go`, `ring.AutomorphismNTTIndex`. Draw a source-level call/dependency graph: imports, helper symbols, `fast.Evaluator` state/caches/scratch, ring vs CKKS dependencies, and method-signature adaptations needed to share the **same implementation**. Compare known source paths for MulRelin and Rescale at API/interface level but do not scope-creep into detailed kernel rewrites.
3. Test **feasibility only** of an import-neutral package such as `schemes/ckks/internal/fastcore`, not an assumption that this specific layout works. It must not import `schemes/ckks` or `schemes/ckks/fast` and must be usable from both packages under Go `internal` import rules. If dependence on `ckks.Parameters`, types, non-exported Fast scratch or ring operations prevents a minimal extraction, identify exact cycles/symbols and compare an alternative (e.g. an import-neutral provider type or package layering change). No new production package, no copy of arithmetic logic, no Go import cycle workaround by frontend `_ import`.
4. Produce a **representation contract table** for logical Levels 0,1,2,3, and >=4: Q-prefix policy width `min(level+1,4)`; `ckks.NewCiphertext` full logical Q backing; `fast.NewCiphertext` compact backing; current legacy q0/q1 operational restrictions; NTT/Montgomery requirements, metadata, c1-zero and aliasing invariants, required input/output authoritative rows. Identify exactly when passing a full-Q input into a Fast prefix kernel leaves stale high rows and must fail closed.
5. Recommend the smallest safe reused-kernel architecture and first integration slice, with source-backed precise method signatures/public returns. Compare **at least** (a) import-neutral extraction, (b) clean package layering alternative, (c) any safe direct low-level ring primitive reuse if extraction infeasible. Explicitly distinguish sharing the existing **Fast kernel** from calling an entirely new or native generic ring algorithm. Explain how to preserve optimized `fast.Evaluator` call sites and current 006A fail-closed behavior without a duplicate core.
6. Specify **future proof tests only**: same unchanged frontend; native Standard reference; no eval-key lookup in Fast zero-secret; matched Level/Scale/slots; positive and negative c1/compact-row cases; assert/instrument that the **same moved Fast kernel** is invoked from both public and existing explicit Fast call sites; compile-time import-cycle validation; test public Bootstrap regressions as a future gate. Provide decision labels `REUSE_FEASIBLE`, `REUSE_REQUIRES_REDESIGN`, or `REUSE_BLOCKED`, with exact evidence.

## Explicit prohibitions

- **Zero modifications to Secondary code**, production/test code or existing Primary frontend/tests. No new algorithm implementation, no `Rotate` or `Rescale` patch, no interface widening.
- No task execution of suspended 007, no planned broad refactor implementation. No Bootstrap, benchmark, CNN, LogN16, or parameter sweep.
- Do not weaken Q-prefix centered-uniqueness `2B<S_Q`, leave Q active high rows stale, or silently fall back to native full-RNS key-switching.
- Do not treat Web's preferred `internal/fastcore` as established; critically validate it. Do not suggest duplicate code unless it has a source-proven necessity.

## Output and review handoff

Write `results/FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008-summary.md` in Primary with mapped imports/symbols/contract matrix, ranked architectural options and an exact next implementation decision point. This may contain pseudocode but **no source diff**. Keep existing artifacts unchanged; safe normal commit/push of report only; self-review the analysis against the new reuse-first gate.

Report `REUSE_FEASIBLE`, `REUSE_REQUIRES_REDESIGN`, or `REUSE_BLOCKED`, and `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`. Leave `FAST-STANDARD-PERF-REBASELINE-003` blocked. Web will decide the *single* next implementation contract after independently reviewing the evidence.
