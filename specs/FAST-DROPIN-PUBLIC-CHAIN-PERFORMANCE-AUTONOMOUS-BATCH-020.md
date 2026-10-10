# FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020

**Status:** READY_FOR_CODEX
**Class:** Bounded experimental performance qualification, A→D with one Web review.
**Accepted predecessor:** `results/FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019-web-review.md`.
**Exact backends:** unmodified true Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; unchanged Fast `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
**User goal:** unchanged public CKKS frontend; dependency checkout only; Fast intentionally insecure zero-secret; quantify **actual time/memory** savings without altering mathematical semantics.

## Strict existing-workload freeze

Reuse the accepted 019 exact canonical E32 LogN13 full-Q/P profile, actual primes, 4096 slots, A/B default Scale 2^45, C legal per-ciphertext Scale=q5, and public `EncryptNew→AddNew→MulRelinNew→Rescale(q5)→RotateNew→DropLevelNew(4)→public E32 Bootstrap→native low-Level DecryptNew/Decode`. Same deterministic input, public keyplan, algorithms, order and numerical thresholds for genuine Standard and Fast. No loss of original CKKS equivalence gates, no switching to explicit fast evaluator in frontend, no changes to Secondary or original Standard checkout.

This is the **first empirical performance baseline** and must distinguish compile/cold-process, key generation, encryption, arithmetic, Bootstrap, final native decrypt and cost of any diagnostic capacity observer. A numerical correctness result from prior batches is not proof of speed.

## A — benchmark provenance and fairness source audit (no expensive call)

1. Sync safely and check pinned head/worktrees/AGENTS; read full accepted 019 source/evidence, 019 Web review, existing performance scripts/instrumentation. Adopt 019 source and math as measurement authority; do not silently rewrite frontend to optimize either backend.
2. Design one **common source-identified benchmark frontend** (can add instrumentation around the 019 computation, or a thin separate benchmark harness calling shared common implementation) whose CKKS operation calls/params remain literally identical in both builds. Instrument low-overhead monotonic wall timing at **well-defined boundaries**: keygen+public GenEvaluationKeys; encryption A/B/C; Add; MulRelin; Rescale; Rotate; DropLevel; entire Bootstrap; native decrypt; total excluding compilation and user-input hold time. If timing source instrumentation necessarily changes shared source, the *same changed common source* must be used on both dependencies, and its hash recorded.
3. Design a fair memory metric (Go total allocated bytes per stage and peak process RSS if reliably obtainable) with identical run options; distinguish allocation from retained physical backed q rows. Same OS/arch/Go version, GOMAXPROCS, process isolation, GC policy/cold/warm state and CPU configuration where measurable. Do not compare compiled/test preparation against hot function times. Explicitly document if true peak RSS is unavailable.
4. Separate and quantify or disable *diagnostic-only* `fastckks.NewEvaluator.ObserveQPrefixCapacity` during timed primitive/Bootstrap sections. The diagnostic adapter must not cause Fast-only timing work. E32 key generation is potentially expensive: do not include it in an alleged eval-only speedup. Decide the narrow warmup/repetition policy before measurements.
5. **STOP** and write a measurement-design blocker if identical-core/source/no-probe-timing is impossible with Primary-only edits. No backend edits authorized, no invented comparative timing.

## B — cheap matched non-Bootstrap timing and oracle checks

1. Create dedicated Primary-only runner/tests, reusing frozen 019 inputs/params and exact public function calls. Use one original Standard and one Fast process under documented controlled settings. On a small predetermined bounded count (e.g. 3–5 if cheap), measure non-Bootstrap primitive stage times (Add, MulRelin, Rescale, Rotate, DropLevel, supported final native Decode), separating public keygen and encryption costs. Do not alter input values, E or capacity policy. If repeats require new keys/ct or temperature management, document and ensure both baselines are treated identically; do not conflate diagnostic overhead.
2. Each run must pass exact established numerical Level/Scale, q-row authority, zero-c1, q0 capacity and plaintext-oracle gates. Keep per-backend raw samples bounded, record median/min/max and individual samples; do not label microbench stage ratio as whole-workflow ratio.
3. Prove real Fast compact outputs and shared kernels, no secret Native Standard fallback; record physical row count/allocated backing separately from runtime bytes. No anomalous input/test adjustment.

## C — bounded Bootstrap timing

1. Only if B passes and fairness A passes, execute **at most one real Standard and one real Fast E32 Bootstrap** in the performance report unless A explicitly demonstrates safe repeatable budget without exceeding task authority (default remains one per backend). Time `Bootstrap` wall clock on the actual held preflight output after a stable and documented preparation phase. Verify original fixed numerical oracles after each call.
2. Record stage duration (whole public Bootstrap call) and process allocations/RSS under same setup. Distinguish cold one-shot variability from repeatable timing: with one call per backend **label only a one-shot observation, not an established speedup**. Do not extrapolate LogN16, GPU, CNN throughput or repeated Bootstrap speed ratios.
3. If Bootstrap fails or startup conditions are asymmetric, STOP, report spent budget and raw metrics; no retry and no change of parameters. Do not run independent hidden warmup Bootstraps.

## D — quantitative reporting and Web handoff

1. Write new `results/FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-{summary.md,journal.md,evidence.json}`, new dedicated runner/tests (Primary only). Include exact hashes, controlled hardware/Go/env/GOMAXPROCS/GC, full timing samples with units (ns/op, ms/op etc), median and empirical ratios as **measured**, numerical gates, memory metrics with acquisition method, measurement overhead warnings and any repeatability limitations.
2. Focused `go test`, `go vet`, self-review and normal fast-forward push Primary. Preserve all 019 artifacts. Secondary and genuine Standard source untouched.
3. Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only for valid fair matched numeric performance evidence, otherwise `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with first precise experimental blocker and consumed Bootstrap budget. Do not start 021 autonomously.

**Strict safety/cost:** No algorithm/parameter/Scale/key-layout changes; no Fast library modifications; no benchmark sweep; no LogN16. Scope bounded to LogN13 E32; no claiming security or production acceleration from zero-secret simulation. If keygen or memory instrumentation dominates, separate the observations honestly rather than hiding them.
