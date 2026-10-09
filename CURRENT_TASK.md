# Current Task

Task: FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014
Status: READY_FOR_CODEX
Mode: APPROVED BOUNDED AUTONOMOUS FOUR-CHECKPOINT INTEGRATION BATCH
Web review: ONCE after checkpoint D, or immediately on genuine scientific/representation blocker

**Runnable spec:** `specs/FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014.md`
**Accepted preceding milestone:** `results/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013-web-review.md`
**Durable end goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`
**Workflow:** `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B and both repositories' `AGENTS.md`
**Secondary branch:** `xuejin-lu/lattigo fast-qprefix`; amended constitution commit `9489925c803be09e49a840e6828b005cdeb1faca`.
**Genuine Standard baseline:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (separate unmodified checkout).

## Architecture decision from Web / user

**Only switch Lattigo dependency; application frontend source, public CKKS API, parameters, E, and call signatures must remain unchanged.** The Fast dependency must internally use the existing zero-secret Fast/Q-prefix kernel implementation. No special Fast constructor or public runtime selector. No separate Normal/Standard implementation required *inside* Fast for comparing results: external unmodified Standard pin is the baseline.

Existing P0 full-active-Q public numerical coverage (013) is an accepted transitional test result, not the endpoint. C0 explicit Q-prefix kernels and proven Rescale CRT/materialization/Bootstrap mathematical rules already exist; prefer bridging and reusing them. The next task must prove real optimized Fast execution behind ordinary public calls, **not** reimplement or retest P0-only generic arithmetic. Do not create an unsafe Standard fallback.

## Four autonomous checkpoints

A: current source/API/import-cycle/representation map; locate minimum import-neutral Fast adapter extraction.
B: implement and test one public API -> existing compact Fast kernel bridge without frontend change; verify dispatch and oracle.
C: minimally compose the bridged operation with other existing public primitives and identify first real interoperability gap, repairing only already-proven internal adapter issues.
D: compact evidence + coding self-review + relevant tests/vet + safe commit/push and one Web milestone handoff.

Details, STOP conditions, allowed repository scope and budget are in the runnable spec. Codex may proceed A→D on passing scoped gates without intermediate Web approval; it must STOP on a new math/representation decision, unexplained oracle failure, dangerous fallback or unsafe Git state.

**Budget:** zero Bootstrap, zero benchmark, zero LogN16. Preserve existing 012/013 artifacts and frozen Standard implementation. Do not start later tasks on Codex's own initiative.
