# CURRENT TASK — Batch025 Zero-call Fast E32 Trace Recovery

Task: FAST-DROPIN-E32-TRACE-OFFLINE-SALVAGE-AUTONOMOUS-BATCH-025
Status: READY_FOR_CODEX
Mode: REUSE-EXISTING-FASTDIAG OFFLINE RAW EVIDENCE ANALYSIS, O1–O4

**Spec:** `specs/FAST-DROPIN-E32-TRACE-OFFLINE-SALVAGE-AUTONOMOUS-BATCH-025.md`
**Web review:** `results/FAST-DROPIN-E32-TRACE-REPAIR-AUTONOMOUS-BATCH-024-web-review.md`
**Permanent rules:** both AGENTS.md, Primary docs/MEASUREMENT_PLATFORM.md, RESEARCH_ENGINEERING_WORKFLOW.md §4A/4B.

Batch024 fully spent two/2 Fast-only E32 calls: oracle correct (RMSE 1.17958e-9; worst error 5.73879e-8). Raw 539-event trace preserved, SHA256 f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74. First validator mismatch event 61 T8 child of event 60 T16, event 59 generated_powers: SOURCE-CORRECT recursive generation. Web directly corrected Primary `validatePublicE32Events` to accept source-backed ancestor chains; regression tests added. Original raw remains TRACE_UNVERIFIED and cannot be rewritten.

O1: safe sync, source, immutable raw/sidecar/fixture and 2/2 journals verified (0 new calls).
O2: existing `cmd/fastdiag` offline replay with strict full-tree/provenance and synthetic tests (0 new calls).
O3: source-grounded internal time Pareto, Stage/Power/Rescale and closure from same saved raw only, or explicit all anomalies (0 new calls).
O4: honest OFFLINE_REVALIDATED_TRACE or OFFLINE_PARTIAL_UNVERIFIED compact report, tests/vet, safe Primary commit/push, ONE Web scientific review.

**ABSOLUTE expensive cap: 0 Standard and 0 Fast Bootstrap across this batch.** No fresh keygen, CKKS loop, new experiment, Secondary math/test changes, new benchmark runner, or alteration to original evidence. Historical Batch023 14/14 and Batch024 2/2 remain spent. Unavailable old Standard pinned worktree is not a blocker for offline Fast-only analysis; do not clean/prune unrelated stale worktrees.
