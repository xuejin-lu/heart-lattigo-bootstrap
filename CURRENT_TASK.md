# Current Task

Task: FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019
Status: READY_FOR_CODEX
Mode: APPROVED FOUR-CHECKPOINT BOUNDED FEASIBILITY-FIRST HIGH-LEVEL MULTIPLICATION TO PUBLIC E32 BOOTSTRAP
Review: ONE independent Web review at completion, or immediate science/architecture stop

**Runnable charter:** `specs/FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019.md`
**Accepted previous review:** `results/FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018-web-review.md`
**Durable target:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`
**Workflow:** both `AGENTS.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, Secondary `docs/FAST_QPREFIX_SPEC.md`

## Accepted results and remaining gap

Batch 018 ACCEPTED for bounded Level5 Fast compact public `EncryptNew→AddNew→RotateNew→DropLevelNew(5)→native Level0 DecryptNew→E32 Public Bootstrap→native Decode` using the exact 017 E32 full-domain generated Q/P without modification. Pair pre-Bootstrap RMSE `1.5656982925854002e-11`, max `7.13774172628691e-11`; post-Bootstrap RMSE `4.963927268857064e-9`, max `3.93386413412319e-8`, each backend one Bootstrap. Fast output q4/q5 dormant at Level5 and zero-c1. Target Level0 q0 had independent `2B<q0` gate. **018 did not run high-Level MulRelin or Rescale.**

Batch 016 separately proved Level5 MulRelin/Rescale with measured high-Level CRT and numeric oracle under **another six-Q profile**, not the canonical E32 Bootstrap chain. Batch 019 addresses this gap, not full CKKS or measured acceleration.

## A→D autonomously, preserving STOP conditions

A: mandatory read-only exact q5, Scale and coefficient-capacity feasibility; note that two initial 2^45 scales lead to `2^90/q5 ≈ 2^30`, not canonical 2^45. Only source-backed public per-ciphertext Scale and regular operations may resolve, without changing frozen Q/P, E32 or default Scale. STOP with evidence if impossible.
B: if A passes, identical frontend Standard/Fast Level5 public Add/MulRelin/Rescale/Rotate/Drop to residual Level0, native decoder, independent bounds and no dormant q4+ read.
C: if B passes, max one genuine Standard + one Fast public E32 Bootstrap on held B outputs, with native oracle and direct paired metrics.
D: compact results, focused go tests/vet/self-review, safe Primary-only commit/push; one Web review. No Secondary/genuine Standard changes.

**Pins:** true Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` separate read-only; Fast `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` read-only.
**Budget:** one Standard + one Fast Bootstrap maximum in batch, only if both preflights pass; zero benchmark, LogN16 or sweeps. No automatic 020.
