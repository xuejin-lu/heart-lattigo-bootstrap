# BTS-NOTRACE-WARM-AND-NUMERICAL-008 — resource preflight

**Status: `RESOURCE_BLOCKED`**

**Bootstrap calls entered in Batch 008: Standard 0/6, Fast 0/6.** No call was reserved, and no measurement process was launched.

## Required No-Trace Cold/Warm table

| 環維度 | CKKS slots | 實作 | Cold (ms) | Warm 中位數 (ms) | Warm 倍率 |
|---|---:|---|---:|---:|---:|
| LogN16 (N=65,536) | 32,768 | Standard | 未測量 | 未測量 | 未計算 |
| LogN16 (N=65,536) | 32,768 | Fast | 未測量 | 未測量 | 未計算 |

The table is intentionally incomplete: the required Standard-first resource gate failed before any Batch 008 attempt. No historical timing was substituted. Task007's Standard process was killed by kernel memorystatus after `NewBTS`; Task001's 139.653 s Standard cold observation and Task007's 7.369 s Fast single cold observation remain historical single-call notes only and are not samples for this table.

## Resource stop evidence

- Host: macOS 26.6.2, Darwin 25.6.0, arm64; physical memory 16 GiB.
- At the preflight snapshot, swap was 5,120 MiB total, 3,748.38 MiB used, 1,371.62 MiB free. `memory_pressure` reported system-wide free percentage decreasing from 55% to 45% during the inspection window.
- Disk had 25 GiB available. Process/file limits were not near exhaustion (`maxproc` soft/hard 2,666/4,000; file descriptors 1,048,575).
- Read-only unified-log evidence at `2026-10-12 23:54:23.317` records kernel `memorystatus: killing largest compressed process bts-standard-decode-007 [66842] 25608 MB`. The Task007 Standard log ends immediately after `NewBTS()` (1m30.373s), before `[enc]` and before `bts.Bootstrap()`, with process exit 137. This is direct evidence of a kernel memory-pressure kill for that prior process, not an inference from exit code alone.
- The available execution host is this 16 GiB Mac; no larger-memory host is configured in this session. Therefore the identical Standard workload cannot be safely relaunched here. Per the Batch008 Standard-first rule, Fast measurement must not begin when Standard cannot safely launch.

The result is `RESOURCE_BLOCKED`, not an assertion that BTS is mathematically unsupported. Resume only after a capable, comparable execution environment is available and a fresh resource preflight passes.

## Numerical comparison

`STANDARD_FAST_OUTPUT_COMPARED` was not reached. There is no Batch008 Standard vector. The Task007 Fast vector may later be considered only for the paired numerical artifact after re-verifying input/source/config identity; it is not a Batch008 performance sample. Its previously recorded input/vector hashes remain in the immutable Task007 result. Max difference, RMSE, mean absolute difference, per-backend error metrics, and speedup are unavailable here.

## Reuse, provenance, and integrity

- Read and followed the maintained-platform contract. `cmd/perfprobe` already has an explicit six-call `public-repeatability` policy, but it measures the separate public-native CKKS fixture, not the teammate's original BTS. Task007's existing measurement-only BTS overlay is the relevant output-capture prior art. No new runner, timer, stats tool, build, or overlay was created because the Standard resource stop was reached first.
- Original `external/BTS` source/config manifest was recomputed over the 13 `*.go`, `go.mod`, `go.sum`, and `run.sh` files: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda` (matches the frozen task pin). The pre-existing executable is excluded. No original BTS file was changed.
- Pinned Genuine Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Pinned Fast source: `ee308c9743000b5f281a84a871e7c4bd59b176c4`. Both existing isolated checkouts were clean at their exact commits. The original Secondary checkout remains untouched at `0b922dc48315128a27db811a5f237e0a6c9813d6` with its five pre-existing dirty files.
- Primary startup sync completed by fast-forward to `bccddf0972dd0e4f54594808d384b1dfb34b6ea4`; the worktree was clean before this result artifact was created.
- No Batch008 Bootstrap reservation, process, build, or cryptographic operation occurred. Historical Batch006/007 records and budgets remain unchanged.

## Validation

- Resource/source preflight: completed; Standard gate: blocked.
- `go test ./internal/numericalmetrics -count=1`: passed; this package test invokes no Bootstrap.
- `go vet ./internal/numericalmetrics`: passed.
- `python3 scripts/task_worktree_policy.py ...`: passed for the three exact task result paths; `python3 -m unittest discover -s scripts -p 'test_task_worktree_policy.py' -v`: 2 tests passed.
- No BTS adapter test was run because the Standard resource gate blocked before adapter creation.
- JSON syntax (`jq empty`) and final staged `git diff --check`: passed before commit.

## Recovery resource preflight — 2026-10-13 (UTC 2026-10-12T19:18:45Z)

The previous `RESOURCE_BLOCKED` record is retained; this is a fresh resume check, not a reset of the Batch or its counters. Primary was clean on `main` at `1727c88a9db00be41e127fd91cd9f94ed30e2a3b`, equal to fetched `origin/main`. Secondary was clean on `fast-qprefix` at `ee308c9743000b5f281a84a871e7c4bd59b176c4`, equal to fetched `origin/fast-qprefix`. The earlier journal note about five dirty Secondary files describes the previous run only; no such dirty files exist in this current checkout.

Fresh local resource readings: 16 GiB physical RAM; `top` reported 15 GiB used, 3,598 MiB in the compressor and 87 MiB unused; swap 3,072 MiB total, 1,942.94 MiB used and 1,129.06 MiB free; `memory_pressure` reported 63% and then 62% system-wide free; load averages 2.02/2.26/2.32 with 84.23% CPU idle; disk available 25 GiB; `maxproc` 2,666/4,000 and open-file limit 1,048,575. The largest observed resident process was WebKit WebContent at 2,454,576 KiB; the remaining top processes included another WebKit process at 929,680 KiB and Codex Renderer at 745,216 KiB.

**Decision: remain `RESOURCE_BLOCKED`.** The improved `memory_pressure` percentage and low CPU load are recorded, but the contemporaneous physical-memory/swap headroom is very small. Combined with the directly recorded prior kernel termination of the matching Standard setup at a 25,608 MB footprint, this does not provide a safe basis to launch the Standard-first series on this 16 GiB host. This is a safety assessment, not a claim that 25,608 MB is the workload's deterministic minimum. No measurement process or binary was started; Standard and Fast reservations/entered calls remain 0/6 each. Fast was not run because Standard must be first. No BTS source or backend code was changed.

The exact live readings and both repository synchronization checks were appended to `journal.json` and `evidence.json`. A later resume must perform another fresh resource preflight; historical Batch 008 calls remain 0/6 and are not reset or consumed by this check.
