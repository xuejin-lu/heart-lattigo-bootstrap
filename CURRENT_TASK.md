# Current Task

Task: FAST-DROPIN-BOOTSTRAP-PERFORMANCE-ATTRIBUTION-AUTONOMOUS-BATCH-021
Status: READY_FOR_CODEX
Mode: S1–S4 BOUNDED AUTONOMOUS BOOTSTRAP COLD/WARM ATTRIBUTION
Web review only after S4 or immediate scientific/unsafe STOP.

**Executable spec:** `specs/FAST-DROPIN-BOOTSTRAP-PERFORMANCE-ATTRIBUTION-AUTONOMOUS-BATCH-021.md`
**Accepted Batch020 review:** `results/FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-web-review.md`
**Workflow:** both AGENTS.md, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, Secondary FAST_QPREFIX_SPEC.md.

## Why this task
Batch020 S1–S5 engineering execution and numerical gates accepted. Pre-Bootstrap evaluation-only cheap median: Standard 6.101624 ms, Fast 3.177875 ms (5 samples). One-shot public Bootstrap: Standard 379.015 ms versus Fast 574.403 ms, BUT this comparison is not a fair steady-state result because Standard `NewEvaluator` eagerly builds circuit data while Fast `Bootstrap` lazily builds it on its first call. Likewise first-call allocations 144 MB Standard vs 961 MB Fast conflate initialization boundaries. `NO_SAFE_OPT_CANDIDATE` justified skipping optional S4 under the prior charter.

## Batch021 S1–S4
S1: source-backed exact cold/warm measurement design and matched same-source instrumentation, with prior accepted 019/020 math/params/API unchanged; safe checkpoint.
S2: **at most two real E32 Standard Bootstrap + two Fast Bootstrap total across whole Batch**. First-call and second-call in same backend lifecycle, same logical input, no retries or extra warmups; independent native numerical oracle for every call. Report evaluator-construction time separately.
S3: independent source-backed attribution of cold/warm wall and Go allocation differences, no further Bootstrap calls or library rewrite; scientific classification.
S4: immutable compact 021 report/evidence, focused tests/vet, safe Primary-only push, one Web scientific review.

**Freeze:** unchanged public application frontend, original LogN13 E32 full Q/P, 4096 slots, A/B Scale2^45, C exact Scale=q5, q-prefix capacity/rows, original max-error1e-6; genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` read-only and Fast `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` read-only. No Secondary optimization in 021, no LogN16/bench sweeps. Never interpret one warm measurement per backend as a statistically stable speedup. Resume journal without repeating expensive calls; do not invent Batch022.
