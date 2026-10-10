# Current Task

Task: FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018
Status: READY_FOR_CODEX
Mode: APPROVED FOUR-CHECKPOINT BOUNDED FEASIBILITY-FIRST INTEGRATION BATCH
Web review: once at batch completion or immediately on a genuine source/math/profile incompatibility

**Runnable charter:** `specs/FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018.md`
**Accepted predecessor:** `results/FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017-web-review.md`
**Goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`
**Workflow:** Primary/Secondary `AGENTS.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, Secondary `docs/FAST_QPREFIX_SPEC.md`

## Current evidence

Batch 017 accepted: genuine unchanged public CKKS EncryptNew → AddNew → MulRelinNew → Rescale → RotateNew → Bootstrap → native DecryptNew/Decode on canonical LogN13 residual-Level1→0, E32 with 4096 slots, using original genuine Standard and intentional zero-secret Fast. One Bootstrap each, final paired RMSE `4.997862453389243e-9`, max `4.326414527528976e-8`. No speed claim.

Batch 016 accepted separately: different frozen six-Q LogN13 profile; true compact-Q high-Level5 Add/MulRelin/Rescale/Rotate numerical comparison passed using a **Fast-only measurement-boundary centered CRT decoder**, not native public DecryptNew. The 016 profile is NOT identical to canonical 017 residual inputs; the two successes are not yet one E32 high-Level compact→Bootstrap public workflow.

## Batch 018 checkpoints

A: mandatory **READ-ONLY exact profile/Level/Scale/CRT feasibility audit**. If canonical E32 residual profile does not permit an above-cap source ciphertext to feed its public Bootstrap without new profile/architecture, STOP with explicit evidence and no Bootstrap attempts; **do not** force/change Q/P.
B: only if A passes, same public frontend Standard/Fast high-Level compact→residual-Level0 matched cheap preflight; ensure legitimate Level/Scale/capacity, normal low-Level DecryptNew and no dormant/full-RNS fallback.
C: only after B passes, max one original Standard and one Fast E32 public Bootstrap; no retries.
D: compact results/report, self-review/test as appropriate, safe Primary-only commit/push, Web review.

**Pins:** Standard original `5dbffbdea05394de2ca3a432ed5318aa832e3f40` independent clean checkout; Fast `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` read-only. Verify current refs before work.
**Cost:** zero benchmarks/LogN16/sweeps; at most 1 Standard+1 Fast Bootstrap ONLY if A/B passing gates. No Secondary edits, no new math or hidden materialization in application workflow. On scientific blocker return `BATCH_BLOCKED_NEEDS_WEB_REVIEW`; no unilateral 019.
