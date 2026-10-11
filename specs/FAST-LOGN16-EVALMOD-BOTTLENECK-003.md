# FAST-LOGN16-EVALMOD-BOTTLENECK-003

Status: READY_FOR_CODEX
Class: performance-forensics / source audit; **no new mathematical design or CKKS production changes**.
Authority: AGENTS.md; docs/MEASUREMENT_PLATFORM.md; docs/RESEARCH_ENGINEERING_WORKFLOW.md.
Previous: CKKS-MEASUREMENT-PLATFORM-ACCEPTANCE-REPAIR-002, accepted for its scoped engineering goals at 8f1848aacf32bf9c6199eb57a9b9f447dd031772.

## Research purpose

We need to accelerate genuine public Fast CKKS Bootstrap for the fingerprint-CNN pipeline, not grow diagnostic infrastructure. A single already captured LogN16/E32 traced warm Bootstrap took 1,060,517,791 ns. Its disjoint direct-stage spans include EvalMod real 375,138,083 ns and imag 382,586,417 ns (71.45% combined), CoeffsToSlots 152,656,958 ns, SlotsToCoeffs 145,065,667 ns. Generated-powers diagnostic roots separately report real 160,157,292 ns and imag 163,584,750 ns; 35 separate Rescale roots exist. These instrumentations overlap the same execution: DO NOT sum the independent power/rescale root totals into Bootstrap or infer an uninstrumented speedup.

Objective: attribute the expensive EvalMod work to actual operations/kernels with source-backed evidence, assess whether Rescale/NTT/CRT/power recurrence or another shared kernel dominates, and recommend **one prioritized, testable optimization hypothesis** with cost model and correctness gate. Do not claim a causal root purely from 71.45% stage wall time.

## Checkpoint A — Existing evidence, not new tooling

1. Safe-sync and preserve local work according to AGENTS. Confirm pins: Primary accepted 8f1848a, Secondary fast-qprefix 2cb8bc7 (verify effective local/remote refs). Inspect sizes before reading any JSON/CSV, trace, or log; use bounded JSON selections, never dump arrays of 32768 slots. Preserve original Task005/006 results and Task002 accepted outputs byte-for-byte.
2. Reuse Task002 compact stage breakdown, its root-separated Task006 evidence, source Task006 539-event raw trace if still available and validated by its SHA, and prior source-backed Batch025/026/027 Rescale microbench methodology and existing perfprobe/fastdiag facilities. Do not create a third analyzer, standalone benchmark runner, new stats library, or per-LogN adapter. If raw unavailable, work from committed compact evidence and explicitly describe limits; do not silently recollect it.
3. Write an evidence table containing observed nine direct stages (nonoverlapping), real/imag power roots, all 35 Rescale roots with proper scope/group, count/cost distributions, and distinct measured vs inferred vs hypothesis flags. Examine Amdahl constraints without summing overlapping timers. Distinguish cold vs traced warm vs Task005 uninstrumented warm median.

## Checkpoint B — Source-backed CPU/time attribution

1. Trace the existing Secondary EvalMod path, Chebyshev power-generation/PS arithmetic, Rescale, CRT/Q-prefix and NTT/INTT implementations to existing measured event types. Document exact file/function pointers and operand sizes, q-prefix widths, Level/Scale transitions. Inspect file size before opening large source. Do not conflate Standard full-Q or q0/q1 theoretical complexity with measured current Fast code paths.
2. Attribute high-cost operations only when source events and measurements support it. If per-operation timings are unavailable, show the uncertainty explicitly. Reuse preexisting focused microbench/profiling hooks for any needed low-cost comparison (no Bootstrap), with finite samples, clean pins and input-bound evidence; do not construct a new timer or copy metrics. Do not perform entire CNN, perf campaigns, or large parameter sweeps.
3. Identify at most 2-3 distinct optimization hypotheses. For each give expected affected fraction, asymptotic operation/limb count, feasible numerical/representation invariants, minimal focused regression, and potential pitfalls. Recommend just one candidate for a later authorized implementation. No claimed 10x E2E gain from an unproven stage-only speedup.

## Checkpoint C — Deliver research result and handoff

- Preserve Fast intentionally insecure zero-secret semantics as a simulation; do not suggest production security equivalence. No Secondary or Primary CKKS production mathematical changes, no API/frontend changes, no Standard source changes. Do not silently reenable Task006's retired unsafe live LogN16 trace adapter; this task is offline attribution.
- Zero new Bootstrap calls; no old performance reruns. If microbench needed, use existing library benches only and journal finite invocation/count, machine, source pins and samples; source-only analysis is acceptable when evidence sufficient.
- Produce a **compact, research-useful** Markdown report and JSON only if structured data materially helps; put breakdown, formula, bottleneck candidate, evidence, falsification gate and next step in results/FAST-LOGN16-EVALMOD-BOTTLENECK-003-summary.md. Do not overshrink away the metrics relevant to the research. Reuse existing test/metrics code; focused verification and diff check. Safe fast-forward Primary commit/push, Secondary read-only; READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW only for genuine math/provenance/Git blockers.
