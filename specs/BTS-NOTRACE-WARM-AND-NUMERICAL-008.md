# BTS-NOTRACE-WARM-AND-NUMERICAL-008

Status: `AUTHORIZED_AUTONOMOUS_BATCH`  
Class: E + minimum necessary I (measurement-only)  
Primary baseline: `f3eb665e35f15b91d5f290626fbb7a3bb360a95c`  
Genuine Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (immutable)  
Fast Secondary source: `ee308c9743000b5f281a84a871e7c4bd59b176c4` (immutable)  
Original BTS source manifest SHA-256: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`

## User-authorized resource-policy amendment — 2026-10-13

The user explicitly removes the former GPT-Web-imposed hard gate that treated low unused RAM, low free Swap, a `memory_pressure` reading, or the historical Task007 memorystatus termination as sufficient by themselves to prohibit launching Standard. Those values and the prior termination remain mandatory observations to record and consider operationally, but they are **not an automatic `RESOURCE_BLOCKED` decision** and do not establish a deterministic memory minimum.

Before measurement, retain the ordinary Git synchronization, source/config/backend provenance, build and zero-Bootstrap adapter/test checks. Record a fresh bounded host-resource snapshot for context, then proceed with the authorized Standard-first measurement even when the advisory readings resemble the earlier snapshots. Do not disable, evade, or tune around macOS memory protection, change BTS parameters, reduce the workload, or substitute a backend. If a real process/build/Bootstrap/decode failure occurs, preserve its exact log and phase, count every reserved/entered call under the existing immutable budget, stop without retrying the same binary/input, and report the actual failure. If Standard fails, do not start Fast because the required comparable Standard-first series is incomplete.

This amendment changes only how resource telemetry is interpreted; it does not enlarge either backend's call budget or waive any source, compile, numerical, provenance, or measurement requirement. The three pre-amendment `RESOURCE_BLOCKED` records remain immutable historical evidence. The user authorized this amendment and resumption of the existing Batch; no new Batch or budget reset is created.

Before any BTS process is launched, finish the workload adapter in the maintained `cmd/perfprobe` / `internal/perfmeasure` platform and pass its zero-Bootstrap tests. The adapter may use a Go overlay of the `external/BTS/main.go` entrypoint only; all original BTS source/config files remain byte-identical. It must create one `NewBTS` instance and one original encrypted input, then call the same `bts.Bootstrap(input.CopyNew())` operation exactly six times (first cold, next five warm) in the same process, with each attempt durably reserved before invocation. The shared attempt-budget helper and existing `internal/numericalmetrics`/`perfprobe` comparison logic own budgeting and aggregation; do not add another runner or metrics engine. Tests use fake callbacks/synthetic vectors and must never call real `bts.Bootstrap`.

## User decision and single primary objective

Complete the **original teammate BTS project's** **No-Trace Cold/Warm performance table**, and harvest the same successful runs' decoded 32,768-slot vectors to complete a genuine Standard-vs-Fast numerical comparison. **Do not prioritize numerical comparison over table completion**, but capture vectors opportunistically in the same executions, without extra Bootstrap for metrics. This is not the earlier standalone public-native LogN16 benchmark in Notion.

BTS original fixed effective parameters (verify from constructed runtime): LogN16, N=65,536, slots=32,768, SlotLimbs=15, BTSModePlain, seed=0, amp=1, input L3/2^32, target returned L18/2^32, E=32 if confirmed by original effective parameters. Do not change parameters, input, native key lifecycle, application CKKS operations, or original `external/BTS` files.

Required table shape, exactly matching Notion:
| 環維度 | CKKS slots | 實作 | Cold (ms) | Warm 中位數 (ms) | Warm 倍率 |
|---|---:|---|---:|---:|---:|
| LogN16 (N=65,536) | 32,768 | Standard | new measured value | new measured value | 1× |
| LogN16 (N=65,536) | 32,768 | Fast | new measured value | new measured value | Standard warm median / Fast warm median |

For each backend, **one first cold Bootstrap + five subsequent warm Bootstrap calls** with same constructed evaluator/keys/parameter profile and logically identical original input, without recreating the evaluator between warm calls. Ensure every call starts with an independent, correctly initialized ciphertext of the same input/metadata and does not inadvertently reuse a mutated prior output. Stop on numerical/domain/metadata mismatch; do not call a failed run a valid sample. Do not substitute earlier historical cold timings into this fresh Cold/Warm table. Keep prior Standard 139653 ms and Fast 7369 ms as labelled historical single cold observations in a separate note only. No stage/power/Rescale traces or warm-up calls outside the six authorized calls per backend.

## Mandatory reuse and preflight

Follow authoritative Primary `AGENTS.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`, and active spec after the required fetch + safe fast-forward. Read only necessary ranges of large files; check byte size before opening logs/csv/artifacts. Inventory/reuse `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics` and Task007 measurement-only overlay/checkpoint. Extend existing maintained measurement infrastructure minimally if the original BTS workload lacks a repeatable adapter. No new independent stopwatch/metrics framework, CLI/runner, CRT or experiment engine. Test zero-Bootstrap path first. Unchanged BTS frontend must use ordinary public APIs and independent clean pinned Standard/Fast Lattigo builds; no backend-specific CKKS math branch in original BTS.

The earlier original Secondary checkout at `0b922dc...` and its five dirty tracked files are historical context only; the user later confirmed those edits were upstreamed and cleared. The current observed `fast-qprefix` checkout is clean at `ee308c9743000b5f281a84a871e7c4bd59b176c4`, equal to `origin/fast-qprefix`. Continue to leave the original Secondary checkout untouched and use separate clean isolated checkouts for linked backends. Do not modify either backend source, Standard implementation, Q-prefix rules, or Fast mathematical operations in this experiment.

Validate exact `external/BTS` manifest and `go version -m` source/dependency provenance and both binary SHA-256 before any expensive call. Source/measurement-only overlay may add output capture and repetition orchestration **outside** original BTS arithmetic source with proven identical per-call operation sequence; reference the measurement adapter SHA. Each attempt's `NewBTS` construction time is recorded independently and excluded from Bootstrap time. Use monotonic wall clock; synchronize sample boundaries, units, and method on both backend builds. Do not claim warm if process/evaluator actually re-created.

## Standard exit 137 observation and actual-failure policy

Task007 Standard process exited 137 after `NewBTS` and before `[enc]`; bounded unified-log evidence later identified a memorystatus termination. **Do not assume OOM solely from 137.** Record accessible macOS signal/memorystatus or other process-failure evidence, memory-pressure/resource readings, process limits and disk constraints. A successful prior Standard full Bootstrap in historical Task001 (2m19.653s) proves it has run at least once but does not prove present resource sufficiency. Distinguish direct evidence from inference. Resource readings are observational context, not a numeric pass/fail gate: low unused RAM, low free Swap, any `memory_pressure` percentage, or the historical termination alone must not block the authorized launch.

Before formal samples, complete the ordinary zero-Bootstrap compile, source/provenance, adapter, output-path and test gates. Then run Standard first using the unchanged profile and OS safeguards. Do not disable or evade macOS memory protection. If the real process is terminated or an actual execution/validation error occurs, preserve the complete log and evidence, mark the relevant reservation consumed, stop the Standard series without retrying the same binary/input, and do not start Fast. Do not infer an execution failure from telemetry alone.

Benchmark comparability requirement: **run both backends on the same machine, CPU/threading configuration, Go version, execution environment, and run policy**. If Standard must move to another machine, run both on that machine or label timings `CROSS_ENVIRONMENT_NOT_COMPARABLE`, and withhold a formal speedup; still allow separately labelled numerical comparison of equal-input vectors.

## Explicit expensive-call budget and autonomous sequence

The *new Batch008* allocates **at most 6 Standard Bootstrap calls and at most 6 Fast Bootstrap calls**, counting cold + 5 warm for each, and **at most one measurement process per backend**. Resource telemetry is recorded but does not gate these authorized calls. Previous batch reservations/calls are immutable, never reused or reset. No retries to replace a failed sample and no additional calls beyond the task budget. No extra bootstrap hidden in tests, validators, vector capture or perfprobe defaults. If one backend fails partway, retain all valid samples with exact `INCOMPLETE` status and do not fabricate a median or paired speedup. Stopping with partial evidence is preferable to retries.

Run Standard first, then Fast, without a resource-telemetry veto. No source repair to Standard or parameter changes. On success, capture each Standard decoded vector and its checkpoint before proceeding to the next warm call, without extra Bootstrap. Then run the Fast series, using the proven Task007 native zero-secret projection, and capture each decoded vector. Verify input identity before calling any result paired. A real process termination or execution/decode failure stops the series and prevents Fast from starting when Standard has not completed.

Prepare/reserve durable journal entries *before each call*; append outcome, monotonic timing, backend/commit/binary, source manifest, runtime device/Go/GOMAXPROCS/E, effective Q/P/slots/Levels/Scale, snapshot/vector SHA, exit code and validation result after each call. Never overwrite Batch007 artifacts. If interrupted, resume accounting from task-scoped journal; never assume a fresh session resets calls.

## Numeric comparison as a free co-product

Compare the same canonical input X, the newly obtained Standard native vector S, and saved/new valid Fast projected vector F. Preserve strict distinctions: Standard full-Q native decode; Fast Level3 Q0123 projection of logical Level18. Both modes must use identical original X, slot order, scale metadata, and 32,768 finite real slots. Verify vector lengths, hashes and complete/finite metadata. The existing Batch007 Fast vector `results/BTS-FAST-DECODE-007/fast-attempt-01-output.f64le` (SHA `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b`) may be used for the numerical pair when source/input/config identity is verified; **never** substitute it as a new Batch008 warm sample. Its input SHA is `5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373`.

Using shared `internal/numericalmetrics`, report MaxAbsDiff=max|F-S|, MeanAbsDiff=mean|F-S|, RMSE=sqrt(mean((F-S)^2)), Standard/Fast max and mean absolute errors versus X, largest-difference slot, hashes, scientific notation, log10 and decimal order `10^floor(log10(metric))` (special-case zero). Record batch and environment provenance. Do not introduce arbitrary numeric acceptance thresholds, suppress nonfinite/large differences, or pretend the intentional Fast zero-secret E32 simulation is secure or formally equivalent to genuine Standard.

## Completion, reporting, and Git

Output `results/BTS-NOTRACE-WARM-AND-NUMERICAL-008/{summary.md,journal.json,evidence.json}` and small bounded raw-sample logs/vector metadata; store complete binary vectors/snapshots only if needed and within repository limits. Explicitly distinguish final `TABLE_COMPLETE`, `STANDARD_FAST_OUTPUT_COMPARED`, `PARTIAL_MEASUREMENTS`, `RESOURCE_BLOCKED`, `NEEDS_WEB_REVIEW`. Keep a copy of the exact six-column table, individual 1+5 sample timings, median, warm ratio, environment proof, and numerical metrics or honest unavailability in summary. Cite earlier Notion benchmark as a separate workload, not part of this table.

Run targeted zero-Bootstrap tests, go vet, diff check, source manifests, and a bounded coding self-review. No full suites that secretly trigger Bootstrap. Commit and normally push changes only to the Primary main after safe startup preflight; do not mutate Secondary. No pull/reset/stash/rebase/force push, no overwrite of unrelated local work. A new unsupported mathematical semantic or confirmed environment impossibility stops with precise evidence, not indefinite theoretical analysis.

**Primary goal: finish the teammate original BTS No-Trace table; numerical output comparison is generated by those same calls.**
