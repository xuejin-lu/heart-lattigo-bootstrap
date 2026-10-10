# Independent Web Scientific Review — FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020

**Date:** 2026-10-10 Asia/Taipei.
**Decision:** **ACCEPT THE AUTONOMOUS S1–S5 ENGINEERING EXECUTION, PRE-BOOTSTRAP NUMERICAL GATES AND DESCRIPTIVE RAW MEASUREMENTS; DO NOT ACCEPT THE 379 ms vs 574 ms VALUES AS A FAIR STEADY-STATE BOOTSTRAP ALGORITHM PERFORMANCE COMPARISON.** The measurement contains a **source-proven asymmetric initialization boundary**, not merely sample-count uncertainty. Classification: `BATCH020_EXECUTION_ACCEPTED_BOOTSTRAP_PERF_UNATTRIBUTED`.
**Original Fast remains unchanged:** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
**True Standard:** unmodified original `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Primary submitted:** verified `main@5b3f50e293270a3440f6251db08e638cd78308fb`.
**Primary S5 source:** `2bf9893a266d188d28197a2aef5848057d7ac9af`; instrumentation introduced at `c486cb2b229f4cc6c0b77f9be2b85066c3b6c213`.

## Evidence inspected

Web inspected submitted `020-summary.md`, `020-journal.md`, `020-evidence.json`, `020-S2-original-baseline.md`, `020-S3-attribution.md`, the instrumented Primary original 019 public runner under `tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019/main.go`, and the exact pinned Secondary Bootstrap sources. Verified Primary and Secondary remote branch HEADs. Web **did not rerun tests** or independently access uncommitted raw 4096-slot decoded samples, local worktrees, or expensive Bootstraps.

The S1–S5 autonomous charter was followed: five matched non-Bootstrap process-isolated pairs, one bounded current-pin Rescale CPU profile, source-backed `NO_SAFE_OPT_CANDIDATE`, S4 intentionally skipped, exactly one Standard and one Fast Bootstrap, and complete Primary results. No Secondary production commit. Initial closed-stdin Standard process exited before any Bootstrap and did not spend call budget. Focused Standard/Fast Go tests/vet are reported passing.

## Accepted bounded numerical and early performance evidence

Same standard public CKKS frontend and source/config/QP/input hashes; LogN13 E32, 4096 slots, A/B default Scale2^45, C per-ciphertext actual Scale=q5, high-level `Add→MulRelin→Rescale→Rotate→DropLevel→Bootstrap`. Genuine Standard randomized encryption vs intentional insecure Fast zero-secret, correct Q-prefix authority and native supported-level decryption. S5 paired Bootstrap output: RMSE `4.914355209707142e-9`, max error `4.2372353589552495e-8` under unchanged `1e-6` gate; each backend called Bootstrap exactly once.

Five cheap, non-Bootstrap evaluation-only stage-sum wall observations: Standard median `6.101624 ms`, Fast median `3.177875 ms` (**1.9207×** descriptive ratio); Fast `Rescale` median `1.729625 ms` was its largest stage, versus genuine Standard MulRelin and Rotate. In this test setting: Go 1.26.4, macOS arm64 Apple M4, `GOMAXPROCS=1`; keygen/evaluation-keygen/evaluator init/encryption costs are substantially different and NOT included in eval-only ratio. The two groups each have just five samples with outliers, no inferential significance. q-backed byte estimates are NOT RSS; peak RSS measurement failed in this sandbox, and no RSS conclusion is accepted.

S3 current Fast Rescale source attribution supports required NTT/INTT, CRT/round/capacity work, and **no obvious mechanically removable no-math copy**. Therefore `NO_SAFE_OPT_CANDIDATE` is a justified engineering decision **within Batch020's narrow S4 permission**. It is NOT a claim there is no optimization opportunity in Fast Bootstrap or any library code.

## HIGH-PRIORITY MEASUREMENT CAVEAT: asymmetric lazy initialization

The S5 records:

| Phase | Genuine Standard | Fast zero-secret |
|---|---:|---:|
| `bootstrapping.NewEvaluator` stage | 406.866 ms / 710.581 MB allocated / ~14.233M objects | 2.618 ms / 8.852 MB / ~32k objects |
| First `Bootstrap()` public call | 379.015 ms / 143.878 MB / 36,648 objects | 574.403 ms / 960.703 MB / 19.282M objects |
| **Init + first-call observed sum** | **785.881 ms / 854.459 MB** | **577.021 ms / 969.556 MB** |

**Source-backed proof of apples-to-oranges boundaries:** in Secondary `circuits/ckks/bootstrapping/evaluator.go`, Standard `NewEvaluator` calls `eval.initialize(btpParams)` which executes `buildBootstrapCircuitData` before returning. For Fast, public `NewEvaluator` returns `NewFastEvaluator`; `fast_bootstrap.go` documents `ensureFastBootstrapCircuit` as a *lazy* build and calls it from `FastEvaluator.BootstrapMany` after public input validation. Its first `Bootstrap` therefore includes DFT/Mod1 circuit-plan/matrix creation, while Standard's first `Bootstrap` does not. `captureStage` wraps the public operation (wall + Go `MemStats.TotalAlloc/Mallocs` deltas), so this difference directly contaminates the purported pure-Bootstrap timing and allocations.

Consequences:
1. `Fast 574ms vs Standard 379ms` is a **truthful cold public-method first-call observation**, but **NOT a fair optimized Bootstrap steady-state cost comparison**, and not evidence that Fast core is 1.52× slower.
2. The striking `19.28M vs 36.6k` first-call object count likewise includes different initialization lifecycles; do not diagnose Fast arithmetic based on this comparison.
3. Aggregating both phases shows an opposite **single-run** order in time (Fast 577ms vs Standard 786ms), but it is not a repeatable end-to-end speedup. Full runtime also includes very different evaluation-key generation and encryption costs.
4. Stage timing, allocation deltas and actual restored-CRT/fast-diagnostic overhead should be kept conceptually distinct. **No numeric results are invalidated** by this timing-method issue.

**Review:** Batch020 subtask execution is accepted; numerical gates pass. The bootstrap-performance research goal is **incomplete attributions** until a properly matched steady-state measurement or separately identified initialization/circuit-evaluation cost is obtained. Do not rewrite historical submitted 020 raw reports. This Web review is the corrective annotation.

## Next research authorization

Create a bounded autonomous `FAST-DROPIN-BOOTSTRAP-PERFORMANCE-ATTRIBUTION-BATCH-021` to 1) preserve 020 raw historical evidence, 2) instrument truly comparable initialization-aware cold/warm lifecycle, 3) measure at most **two** genuine Standard and **two** Fast Bootstrap calls total (first cold, second warmed), with original 019 numerical gates, same public frontend/source and untouched backends, 4) produce source-backed stage and Go allocation attribution isolating expensive Fast first-call matrix build, 5) decide whether optimized kernel performance is credible, without an algorithm/library change in this 021. All expensive calls must be budgeted and journaled; any need for a new primitive, changed math, unsafe GC/profiling mismatch or more runs must STOP for Web.
