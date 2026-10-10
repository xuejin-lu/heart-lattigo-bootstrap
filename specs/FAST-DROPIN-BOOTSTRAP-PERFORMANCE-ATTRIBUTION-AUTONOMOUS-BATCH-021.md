# FAST-DROPIN-BOOTSTRAP-PERFORMANCE-ATTRIBUTION-AUTONOMOUS-BATCH-021

Status: READY_FOR_CODEX. Autonomous S1–S4 diagnostic milestones; ONE final Web review or scientific STOP.

## Objective
Correct the **cold-vs-warm Bootstrap attribution** gap documented at `results/FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-web-review.md`. The 020 one-shot timings Standard 379 ms/Fast 574 ms are not steady-state comparable: Standard constructs circuit data in `bootstrapping.NewEvaluator`; Fast builds circuit data lazily inside its first `Bootstrap`. Preserve the original 020 raw reports; no reinterpretation of them as stable speedups.

## Immutable experiment
True untouched Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Fast untouched `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`. No Secondary production edits. Reuse the accepted 019/020 same-source Public CKKS frontend, exact LogN13 E32 and Q/P configuration, 4096 slots, input hashes, C Scale=q5, original `Add→MulRelin→Rescale(q5)→Rotate→DropLevel→Public Bootstrap→DecryptNew`, strict capacity/row/Level/Scale, native plaintext oracle and 1e-6 error gate. The intentional Fast zero-secret semantics remain insecure. Reuse existing Primary runner and profiling tools; no second math pipeline.

**Total expensive budget across S1–S4: MAX TWO Standard and TWO Fast real Bootstrap calls**, including any failed or warmup calls. No hidden retries or extra warmups, no LogN16, broad sweeps or GPU extrapolation. These four total calls are a diagnostic ceiling; if a single source-backed approach gives comparable results with fewer calls, prefer fewer. The first and second calls MUST be attributed as cold and warmed respectively. Do not call twice on output ciphertext unless confirmed same logical input/baseline profile; preserve a clone/fresh prepared identical input for both invocations. Independently verify each numerical oracle and input fingerprint. Do not compare source/hardware mismatches.

## S1: instrumentation and method, no Bootstrap
1. Safe sync, inspect both AGENTS, original 019/020 evidence, current Secondary `circuits/ckks/bootstrapping/{evaluator,fast_bootstrap}.go`. Explain source-proven initialization asymmetry before measuring. Capture exact source/config/Q/P/input/compiled dependency identities, hardware, GOMAXPROCS and GC policy.
2. Design low-impact same-source timing/alloc measurement split: construction (`NewEvaluator`), first Bootstrap, second Bootstrap, and sum construction+first. Optionally use separate diagnostic-only observed lazy build marker if a supported API permits it, but **do not move initialization with a Fast-only app call** or change public behavior. Record `runtime.MemStats` allocation deltas distinctly from peak RSS. No Fast-only observer within timed Bootstrap interval.
3. Add tests for one-shot budget accounting, source equivalence and lifecycle labels. Save S1 commit/journal and automatically continue.

## S2: two-call per-backend bounded matched lifecycle
1. Same process/evaluator/backend within each backend pair, fixed conditions; first public Bootstrap call unavoidably cold, second warmed (Fast initialized). Genuine Standard evaluated with matching same-input semantics. EXACTLY at most two calls per backend, no retry after any actual invocation.
2. Capture init+first-call wall/alloc and steady second-call wall/alloc; retain raw per-call metrics. After each call, native public `DecryptNew/Decode`, direct Standard/Fast comparison where valid, Level/Scale/capacity gates and 4096-slot original oracle. Reject empty/NaN pair metrics. Be explicit about reusing the original logical input state, not bootstrapping already-bootstrapped ciphertext by accident.
3. Persist S2 preliminary evidence with consumed call count and SHA, independent of later optional profile. Advance without intermediate Web review if numerical/provenance passes.

## S3: attribution, cheap-only after call budget
1. Analyze method-consistent new cold/warm wall and allocation deltas; distinguish evaluator creation from lazy circuit setup and the actual steady public Bootstrap. Compare raw allocation objects/bytes to source-backed initialization path, note GC, cache effects and uncertainty.
2. Only use non-Bootstrap profiling/replay or already collected cheap data to localize remaining cost. No extra Bootstrap beyond S2. Report whether Fast is faster/slower as **observed**, not a stable population estimate based on one warm call.
3. Classify `INIT_BOUNDARY_EXPLAINS_MUCH`, `WARMED_FAST_STILL_SLOW`, `INSUFFICIENT_EVIDENCE`, or evidence-appropriate precise variant. No optimization implementation or code rewrite in this batch. Continue to S4.

## S4: comprehensive report and Web handoff
1. Write `results/FAST-DROPIN-BOOTSTRAP-PERFORMANCE-ATTRIBUTION-AUTONOMOUS-BATCH-021-{summary.md,journal.md,evidence.json}` with all calls, source/hardware hashes, exact init/cold/warm times, allocations, numeric RMSE/max, diagnostics and uncertainty. Keep 020 results immutable.
2. Focused Standard/Fast tests, vet, diff check and Codex self-review. Primary normal fast-forward commits and push, Secondary and genuine Standard remain pinned/clean. Terminal `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` on all gates; on numerical discrepancy, changing mathematical contract, unsafe Git or exhausted budget return `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with exact blocker.
3. Do not start Batch022 autonomously. Web will select any subsequent algorithmic optimization, other parameter family or statistics plan.

This four-stage package is one autonomous milestone, not four separate Web reviews. Never claim cryptographic equivalence or stable speedup from these few runs.
