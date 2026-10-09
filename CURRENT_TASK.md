# Current Task

Task: FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R
Status: READY_FOR_CODEX
Mode: APPROVED PRIMARY-ONLY EVIDENCE REPAIR (THREE CHECKPOINTS)
Review: ONE independent Web acceptance after repair or immediate genuine blocker

**Runnable spec:** `specs/FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R.md`
**Web review / root cause:** `results/FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-web-review.md`
**Frozen original report:** `results/FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-{summary.md,journal.md,evidence.json}` — never overwrite.
**Workflow:** `AGENTS.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B

## User goal and review classification

Batch 015 completed the selected compact public CKKS Add → MulRelin → Rescale → Rotate Fast Q-prefix core integration. However, the Primary evidence aggregation **lost raw decoded vectors through Go slice aliasing**, then wrongly accepted two empty arrays as RMSE=0. The eight saved direct `fast_vs_standard_rmse=0` and max-error=0 values are invalid; original plaintext-oracle results remain bounded historical evidence, not independent Web reruns.

**This task repairs evidence only.** Preserve all Secondary CKKS algorithms/kernels and genuine Standard baseline unchanged. Freeze frontend, Q/P, E, deterministic inputs, thresholds and levels. Correct the aliasing and empty/mismatch/nonfinite guards; add regression tests; run one fresh genuine Standard and one Fast matched runner; publish corrected standalone 015R evidence with real paired RMSE/max.

## Sequential checkpoints

A: inspect source; regression-first unit tests for `summaryOf` immutability and `compare` nonempty/shape; fix combiner without modifying workload.
B: two small matched runs (Standard/Fast) and corrected compact numerical report with exact provenance; do not overwrite 015 artifacts.
C: focused tests, vet, diff/self-review, safe Primary-only commit/push, final Web handoff.

**Pins:** Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` separate clean; Fast Secondary `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` branch `fast-qprefix`, read-only.
**Cost:** zero Bootstrap, zero benchmarks, zero LogN16. No Secondary modification or push. No subsequent batch without a new Web authorization.

**Expected terminal:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only if corrected eight real pairs and original individual plaintext gates pass; else `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with exact failure.
