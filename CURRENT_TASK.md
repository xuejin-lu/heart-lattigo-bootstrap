# Current Task

Task: FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008
Status: READY_FOR_CODEX
Task class: architecture feasibility investigation (READ-ONLY; no code changes)

**Executable spec:** `specs/FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008.md`
**Source-backed architecture inventory:** `docs/FAST_DROPIN_EVALUATOR_REUSE_ARCHITECTURE.md`
**Required workflow gate:** `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A
**Original goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Why this replaces 007

The older `schemes/ckks/fast.Evaluator` already implements Add/MulRelin/Rescale/Rotate, whereas `ckks.NewEvaluator` still builds and runs generic/native CKKS arithmetic. Existing `006A` corrects the fail-closed keyless full-Q public MulRelin boundary but **does not prove reusable Fast Q-prefix core execution**. The previously proposed `FAST-DROPIN-KEYLESS-ROTATE-007` was suspended before implementation and **must not be executed**, because another public Rotate implementation could duplicate existing Fast code and mishandle authority of compact Q-prefix residues.

## Current authorized work

Using **current Secondary source** pinned initially at `93ab7ecc5a864fd9bbacea55f0af730941552691`, map the full import/helper/API graph for existing Fast Rotate/Automorphism and compare MulRelin/Rescale entrypoint constraints. Test feasibility of extracting/moving a shared import-neutral Fast kernel reachable from both ordinary `*ckks.Evaluator` and the existing explicit Fast evaluator, **without frontend changes, import cycles or duplicate algorithms**. The candidate `schemes/ckks/internal/fastcore` is a hypothesis, not a forced implementation. Produce a Q-prefix versus full-Q representation contract and rank safe architecture options. No implementation, no tests/benchmarks/Bootstrap, no changes to Secondary. Primary report only; return `REUSE_FEASIBLE` / `REUSE_REQUIRES_REDESIGN` / `REUSE_BLOCKED` and `READY_FOR_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.

Preserve pinned genuine Standard, accepted 006A fixes and all original APIs. Formal `FAST-STANDARD-PERF-REBASELINE-003` benchmark remains BLOCKED.
