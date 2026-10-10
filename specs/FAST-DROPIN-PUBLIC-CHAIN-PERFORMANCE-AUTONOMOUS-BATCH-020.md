# FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020 — Autonomous Performance Sprint

**Status:** READY_FOR_CODEX
**Purpose:** This is a genuine **multi-milestone autonomous batch**, not a single stopwatch task with four routine gates. Finish S1→S5 without intermediate Web handoffs; one Web scientific review at the sprint end, or immediate STOP on new mathematical/architecture issues.
**Previous accepted result:** `results/FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019-web-review.md`.
**True Standard:** original Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (read-only).
**Original Fast performance baseline:** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` (freeze as baseline even if S4 later produces an optional new Fast candidate).
**Reuse-before-reinventing:** `results/QPREFIX-PERF-DIAG-007-summary.md`, `results/QPREFIX-PERF-OPT-004-summary.md`, `results/FAST-STANDARD-PERF-REBASELINE-002-comparison.md` and existing benchmark tools. Historical workload/pins differ, so historical speed ratios are **not comparable** to Batch019 public API.

## Global frozen mathematical/API contract

The **same CKKS public application source** must compile with genuine Standard or Fast library dependency (only the library build/checkout changes), with no Fast-specific API or runtime selection. Reuse the accepted 019 exact E32 LogN13 full Bootstrap Q/P primes, 4096 slots, default Scale 2^45, A/B Scale 2^45, C's legal per-ciphertext Scale = actual q5, input values, keyplan, ordinary methods:
`EncryptNew(A,B,C) → AddNew → MulRelinNew → Rescale(logical q5) → RotateNew → DropLevelNew(4) → public Bootstrap → native Level1 DecryptNew/Decode` (and native Level0 preflight).
Maintain original 1e-6 numeric max-error gate, exact Level/Scale, strict 2B<S_Q capacity, zero-c1 in Fast, q4+ dormancy after compact public operations, no full-Q fallback, genuine randomized nonzero-c1 Standard lifecycle, E=32. Fast is an **insecure zero-secret simulation**, not security preserving.

Measurement probes and Fast-only Q-prefix observer may exist in separate diagnostic/build-tag code, but NEVER as a different app math/API path, and NEVER inside timed kernel sections. If measurements require editing common frontend source, both compiled backends must run the identical modified source and report its exact SHA.

**Batch-wide expensive cap:** ONE actual Standard E32 Bootstrap + ONE actual Fast E32 Bootstrap **total**, performed only in S5 after all cheap validation. Zero expensive warmup, retry, sweep or hidden calls; no LogN16 and no CNN extrapolation. A single Bootstrap time is one observation, NOT a statistically stable speedup.

**Implementation authority:** Primary can add common benchmark harness/tests and compact reporting. Secondary initially read-only; S4 MAY authorize exactly ONE profiler-backed **mechanical**, same-math optimization restricted to a four-file allowlist below. Original Standard is never editable. Scientific maths/representation/API/algorithm change is *always* a Web STOP, never silently authorized within 020. Safe Git and self-review policy from both AGENTS and workflow `4A/4B apply. Never reset, force-push, stash, clean, discard other people's work. After any meaningful subtask: test, self-review, at most one repair, safe atomic commit and normal fast-forward push. Continue automatically after passing each gate.

## S1 — FAIR reusable measurement harness (first independently useful milestone)

1. Safely sync refs and read accepted 019 runner/review, current code and historical diagnostic benchmark frameworks. Freeze environment and measurement methodology before measuring: actual Go version, CPU/OS/arch, GOMAXPROCS, GC policy, source hashes, effective Q/P, inputs, both pinned backends, initialization conditions and process isolation.
2. Build or reuse ONE same-source public operation harness. Wall-clock stage timers: GenEvaluationKeys/keygen, Encrypt A/B/C, Add, MulRelin, Rescale, Rotate, DropLevel, native Decode, Bootstrap, total eval-only, process total. Exclude compilation, stdin hold/wait, probes and diagnostic capacity observers. Distinguish setup vs evaluation.
3. Fair memory signals: stage allocation counters, approximate physical q-backed bytes, and peak RSS only if OS measurement is reliable. Do NOT conflate q-row count, allocated bytes and RSS. Avoid high-cost probes in timed sections; if necessary run allocation probes in an untimed lane.
4. Test provenance and source identity, focused go test/vet, one source review; commit Primary harness. Journal `S1_COMPLETE`, continue to S2 without Web review.

## S2 — Original-pins cheap Standard/Fast baseline (second milestone)

1. With frozen 019 inputs and exact public operators, make **five** bounded cheap repetitions per backend of each primitive and the non-Bootstrap composition. Time outside key generation/encryption when reporting eval-only; report individual raw samples, median and dispersion (not only a ratio). Document cold/warm preparation and data reuse. NO Bootstrap here.
2. In every paired test gate check Standard genuine native ciphertext/oracle and Fast zero-c1/compact four rows, exact Level/Scale/q5 Rescale, final 2B<q0, no dormant reads/native fallback, native residual Decode; do not weaken thresholds for speed.
3. Report stage total, total allocations and memory metrics. Save frozen compact `S2-original-baseline` in 020 journal/evidence or a separate small subartifact, including Original Fast pin 2d6145d... and genuine Standard SHA. Self-review, commit Primary, continue to S3.

## S3 — Diagnose highest current evaluation bottleneck (third milestone)

1. Attribute the S2 stage wall/alloc bottlenecks; keep keygen/encryption separate from eval-only and carefully distinguish microbench from full execution. Reuse historic QPREFIX-PERF-DIAG methods without copying its old percentages into current data.
2. If needed, take one bounded cheap-only CPU/heap profile of the highest currently material stage, zero Bootstrap. Distinguish flat vs cumulative pprof shares (do not sum overlapping stacks). Record exact source/function and how current public wrapper dispatches into existing fastcore; no speculative math rewrite.
3. Publish `S3-ATTRIBUTION` and choose autonomously:
   - `MECHANICAL_OPT_CANDIDATE` iff a **single** high-impact, source-backed optimization uses existing wrapper/core arithmetic unchanged and needs **no new math, ownership semantics or Q-prefix contract**; specify exact file/function, source/metric motivation, low-risk change and before/after tests.
   - `NO_SAFE_OPT_CANDIDATE` otherwise. This is a **successful diagnostic outcome**: automatically SKIP S4 and continue S5 with original pinned Fast. No filler changes, no mandatory Web handoff.
4. Commit attribution/provenance to Primary at a coherent checkpoint. Do not spend Bootstrap here.

## S4 — OPTIONAL: exactly one mechanically safe Fast improvement (fourth milestone)

S4 is only authorized if S3 yields `MECHANICAL_OPT_CANDIDATE`. The **only** Secondary production file allowlist is:
- `schemes/ckks/internal/fastcore/mul.go`
- `schemes/ckks/internal/fastcore/rescale.go`
- `schemes/ckks/evaluator_fast_zero_secret.go`
- `schemes/ckks/evaluator_fast_zero_secret_automorphism.go`
Plus tests targeting changed source. One candidate; no new package, key layout change or reimplementation of existing math.

1. Permitted category: mechanically remove demonstrably redundant **representation-neutral** allocation/copy or reuse existing scratch *without changing aliasing/input/output ownership* or any checks. Not permitted: change CRT/INTT/NTT/division, Scale/Level, key semantics, memory authority, optimized limb count, capacity guard, fallback routing, API, or numerical tolerance. If candidate requires such change, record its classification and SKIP S4 as `NEEDS_FUTURE_WEB_ARCHITECTURE`, continue S5 unchanged; do not implement it.
2. One implementation, self-review, max one local repair. Run targeted negative/alias/capacity tests and 019 cheap pre-Bootstrap Standard/Fast oracle; original Standard untouched. On unexpected math/numerical failure after bounded repair, STOP for Web without Bootstrap. Do not hide rejected changes.
3. Benchmark frozen original Fast commit vs candidate at least five bounded cheap trials each on same stage and fixed source. Retain candidate only if ALL: targeted median stage improvement **≥5%**, eval-only chain median regression **≤3%**, no material alloc/representation regressions, exact oracle and API invariants pass. These are conservative engineering gates, not statistical significance.
4. If accepted, normal fast-forward commit/push Secondary `fast-qprefix` and record **both original and new Fast SHA**. The S5 **one Fast Bootstrap** runs on the accepted candidate and is not a before/after Bootstrap study. If correct but below improvement gates, reverse ONLY task-owned candidate diff under explicit review, never reset/stash/clean; retain negative evidence, leave Secondary on pinned original Fast and continue S5. If worktree provenance unsafe, STOP.

## S5 — Paired one-shot Bootstrap, synthesis and ONE Web handoff (fifth milestone)

1. With all cheap numerical/provenance gates passed, run ONE actual public E32 Standard Bootstrap and ONE Fast Bootstrap (original pin or accepted S4 candidate) on their actual valid held 019 chain outputs; same host, source, Q/P, E32, Scale and oracle. Never rerun Bootstrap just for favorable data. Time public wall-clock Bootstrap call and record allocations/RSS where reliable; distinguish one-shot cold-state uncertainty.
2. Verify native DecryptNew/Decode at residual level, direct paired RMSE/max, q-prefix layout/keys, original 1e-6 threshold. A candidate whose Bootstrap now fails is a genuine scientific blocker; report budget spent, no retry or reinterpretation.
3. Final `results/FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-{summary.md,journal.md,evidence.json}` shall include separate **S2 original baseline, S3 attribution, optional S4 candidate trial/before-after, S5 one-shot Bootstrap**, timing samples and dispersion, alloc/RSS methods and limitations, all SHAs, verdict (`MEASURED_BASELINE_ONLY`, `NO_SAFE_OPT_CANDIDATE`, or `MECHANICAL_OPT_ACCEPTED`), prioritized research follow-up without auto-starting 021.
4. Focused go test/vet/diff/source review, Primary final safe commit/push and Secondary clean at original or accepted candidate SHA. Terminal status `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only if actual matched numerical and measurement gates passed; otherwise `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with exact first failure and consumed expensive budget.

**Scientific STOP boundaries:** any newly invented CKKS math or altered CRT/Q-prefix/Scale/Level/key contract; native high-Level compact decrypt assumption; missing optimized dispatch/full-RNS fallback; genuine Standard error; unsafe repository; changed E32 or input; extra Bootstrap/warmup; more than one candidate. Continue independently across successful S1–S5, not one Web review per minor stage.
