# Current Task

Task: FAST-STANDARD-PERF-REBASELINE-003
Status: READY_FOR_CODEX
Task class: E — Post-Chebyshev matched performance and independent Standard numerical confirmation

**Executable spec**: `specs/FAST-STANDARD-PERF-REBASELINE-003-POST-CHEBYSHEV.md`

Prior `FAST-QPREFIX-CHEBYSHEV-ORDER-002` is Web-reviewed as **production Chebyshev order restored in both LogN13 and LogN16 for the existing deterministic zero-a numerical protocol**; Secondary repaired and pushed `75ef5dbe7bbf7d3947fb2b9fb232c4a56f05c948` and Primary summary `results/FAST-QPREFIX-CHEBYSHEV-ORDER-002-summary.md`.

**What remains**: remeasure latency / allocations + verify paired numerical output against an **independently built genuine Standard** from `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, in BOTH LogN13 and LogN16. The earlier `fastdiag` Standard evaluator is Standard API compiled from the Fast tree, not the independent historical Standard commit. Do not mislabel; do not infer new speedups from pre-repair 4.3755x/5.6701x.

Reuse the frozen `cmd/perfprobe` and `internal/perfmeasure` implementation, `perf_fast`/`perf_standard` tags, existing `cmd/fastdiag` checkpoints and `internal/numericalmetrics` SNR. Same effective parameters / input within each profile. One untimed warmup + 7 full Bootstrap timing runs per backend/profile, fresh numerical trial pairs, compact reports. **No Secondary production modifications, no new harness, no speculative algorithm changes, no thresholds weakening.**

Codex must safe-sync per AGENTS, perform bounded cycle, push authorized Primary results if checks pass and end with one spec-defined classification then `READY_FOR_WEB_REVIEW`.
