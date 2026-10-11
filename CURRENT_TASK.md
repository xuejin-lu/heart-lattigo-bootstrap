# CURRENT TASK — LogN16 capacity-safe native E32 performance, inherited budget

Task: FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005
Status: READY_FOR_CODEX
Authoritative specification: `specs/FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.md`

## Web review disposition

Task 004 completed as `CURRENT_PERFORMANCE_BLOCKED` / `PUBLIC_NATIVE_LOGN16_PROFILE_BLOCKED`, with **0/16** full Bootstrap attempts spent. This is not a proven coefficient overflow: `internal/perfmeasure.BoundPublicMulRescaleWorkload` found a conservative centered-capacity bound `B=18169085704944493` for original LogN16 public-native E32 workload; `q0=36028797019488257`, so `2B-q0=309374390400729` (about 0.859% of q0). Keep the old failure and guard intact. No proof yet of current LogN16 Fast kernel runtime or speed.

## New safe experiment and authorization

Create a **separately named** LogN16/E32 public-native workload `logn16-e32-c15over16-v1` in our existing `perfprobe`/`perfmeasure` platform, with original A/B and C scaled exactly by `15/16`. Keep Q/P, Scale, E32, full identical public CKKS frontend and original genuine Standard/Fast pins unchanged; change only the named benchmark's input C **for both backends equally**, and regenerate its own workload SHA/oracle. Never alter the frozen original LogN13/E32 fixture or original failing LogN16 case, silently bypass the `2B<q0` gate, or claim these two LogN16 workloads are identical.

First require zero Bootstrap capacity recomputation, canonical hashes, full matched original Standard and Fast native public E32 preflights, all eight checkpoints, dispatch/kernel evidence and focused tests. If any preflight fails or Fast optimized execution cannot be established, STOP and spend **zero** calls; no Secondary changes.

Only AFTER those gates pass: **TRANSFER, not add to, task 004's unspent total allowance**. Across 004+005 maximum stays 16 full Bootstrap calls: LogN16 Standard 1 cold+5 warm, Fast 1 cold+5 warm (12); LogN13 frozen original Standard 1 cold+1 warm, Fast 1 cold+1 warm (4). Durable task-wide cumulative reservation journal, failures consume calls, no implicit warmups or repetitions, no retries, no extra profiler Bootstrap. Reuse existing timers/alloc counters and `fastdiag`, measure paired native numerical oracles plus speed; report exact Fast kernel path and construction vs cold vs warm cost. No mathematics, backend, parameter or optimization modifications authorized.

Produce new `results/FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.md` and `.json`, preserve 002/003/004, safe push Primary only and READY_FOR_WEB_REVIEW/NEEDS_WEB_REVIEW. Follow latest AGENTS/workflow and backend transparency contract. User need only type `開始` in Codex.
