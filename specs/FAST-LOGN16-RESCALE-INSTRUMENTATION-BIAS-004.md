# FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004

Status: READY_FOR_CODEX
Class: bounded empirical measurement / profiling-validity study (E/I), **not optimization implementation**.
Authority: AGENTS.md, docs/MEASUREMENT_PLATFORM.md, docs/RESEARCH_ENGINEERING_WORKFLOW.md.
Previous accepted evidence: FAST-LOGN16-EVALMOD-BOTTLENECK-003 @ 6271aae2fee2cdfb286f112f309e0a8e0a9e90ec.
Production Fast pin for historical comparison: Secondary fast-qprefix @ 2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f.

## Why this gate is necessary

The Task003 source audit found 28 EvalMod Rescale roots of roughly 25 ms each under a 1,060.518 ms **traced** LogN16/E32 Fast Bootstrap (EvalMod 71.4485%). Crucial unaccounted confounder: `schemes/ckks/internal/fastcore/rescale.go` invokes `fastdiag.StartTimer()` and `ElapsedNS()` inside every coefficient iteration *when tracing is selected*. For the 28 EvalMod Rescales, 2 components x N=65,536 entails 3,670,016 coefficient iterations and up to 14,680,064 clock-related function calls (two starts plus two elapsed calls per iteration). This may significantly perturb observed coefficient-loop, Rescale, and EvalMod timing; it is not yet measured how much. Task005 uninstrumented Fast warm median 663.878 ms and Task006 traced warm 1060.783 ms are NOT matched trace-on/off measurements; their difference cannot be called timer overhead.

Another confounder: the existing `BenchmarkFastRescaleLogN16` in `schemes/ckks/fast/rescale_test.go` uses an alternate LogN16 Q bit-chain (q0=54, q1=38), not the exact Task006 production Q/P chain. Do not mistake this for a Task006-matched benchmark.

**Research question:** Does the purported Q0123 coefficient-loop hotspot remain a hotspot when instrumenting cost is removed, with source-bound LogN16 four-row input, or is the trace overhead driving its apparent share? Decide whether to pursue CRT-loop optimization only after this gate.

## Checkpoint A — audit and configuration

1. Mandatory safe Git sync, preserve local work. Size-check files before parsing, especially Task006 raw, JSON/CSV, or per-slot vectors. Inspect Task003 report and Task002 compact 539-event evidence without changing either or the original raw TRACE_UNVERIFIED classification.
2. Audit exact `fastcore.RescaleWorkspace.ApplyRows/rescaleComponent` calls and build-tag `internal/fastdiag/{enabled,disabled}.go` clock implementation; document timer count per component and emitted/nested spans. Audit existing benchmarks `BenchmarkFastRescaleLogN16`, `BenchmarkFastRescaleQPrefixRows4LogN13P93`, Batch027 helpers/fixtures; classify what each does and doesn't match. Reuse `testing.B`, package microbenchmarks and existing metrics/oracles. No new benchmark program/framework, event-tree validator or custom timer.
3. Establish a valid **matched** LogN16 Rescale input fixture: two components, N=65536, authoritative q0..q3 rows, relevant EvalMod Level12->11 and Scale approximately 2^120->2^60, exact effective Task006 Q chain (hash-check against committed metadata) and documented NTT/Montgomery/in-place flags. Check centered capacity and numerical semantics using existing oracles before timing. If direct Task006 ciphertext unavailable, use a clearly labelled synthetic but shape-matched valid coefficient input; don't claim it is the production distribution. If exact Q chain cannot be reproduced safely and cheaply, state the blocker and use unmatched existing benchmarks only as exploratory controls, never a formal match.

## Checkpoint B — matched instrumentation perturbation study

1. Use the **same input/parameters and identical Rescale arithmetic source** in three measurement situations where technically feasible:
   A. production compiled without `fastdiag` tag;
   B. compiled with `fastdiag` tag but Rescale tracing mask disabled;
   C. compiled with `fastdiag` tag and `fastdiag.ConfigureCSV("rescale")` enabled.
   For A/B/C record build tags, source/blob or commit pins, fixture hash, effective parameters, output Level/Scale/rows, correctness oracle, ns/op, allocs/op, finite sample counts. Control compiler/setup/caches, benchmark loop placement and output aliasing; test that the timer/events are enabled only in C. Use identical workload, bounded repetitions and adequate warm-up; prefer e.g. 5 independent low-cost samples per setting, no endless retries or expensive campaign. Report noisy/heterogeneous distributions instead of cherry-picking.
2. Use existing benchmark functions or a minimal focused test/benchmark case in the maintained Secondary test package, **not a new independent tool**. If a test-only extension is required, run it in a task-owned isolated worktree/copy while keeping production Secondary `fast-qprefix` unchanged and clean; store a small, auditable patch/digest or reproducible exact instructions in Primary as needed, not a new maintenance subsystem. Do not modify/push Secondary production arithmetic or committed branch just to collect evidence. Standard may be consulted as a semantic oracle, not used as a comparable trace-performance baseline without separately pinned isolated build.
3. Do not run any full Bootstrap, reconstruct a 32768-slot historical vector, or perform large parameter sweeps. If local experiment environment is not capable of the matched benchmark, report an exact `MATCHED_FIXTURE_UNAVAILABLE` or `MEASUREMENT_INCONCLUSIVE` blocker rather than substituting an unrelated benchmark or writing a new measurement infrastructure.

## Checkpoint C — decision and delivery

- Separate clock/collector overhead, build-tag-only overhead and uninstrumented Rescale cost. An observed fastdiag slowdown under a matched microbenchmark is an **instrumentation perturbation**, not necessarily a measured Bootstrap slowdown. Do not ascribe Task005-vs-Task006 delta to instrumentation without a matched E2E control.
- Provide observed CRT/round/coefficient-loop share only where measurement supports it; avoid addition of overlapping root times. Give an evidence-backed yes/no/inconclusive for whether current Q0123 Rescale still merits the next optimization implementation. If yes, identify one constrained improvement to test with relevant oracle/safety gates; if not, recommend a different measured bottleneck. Do not implement any new CKKS arithmetic optimization in Task004.
- Produce compact `results/FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004-summary.md` with source/fixture provenance, commands, experimental A/B/C results, median/spread, interpretation, limitations and **one next hypothesis**. Optional small JSON only if beneficial. Tests/vet/diff checks focused and cheap. No new full Bootstrap calls, Secondary production stays pinned/clean, Primary safe commit/push. Finish READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW only for genuine evidence/math/security/Git blockers.
