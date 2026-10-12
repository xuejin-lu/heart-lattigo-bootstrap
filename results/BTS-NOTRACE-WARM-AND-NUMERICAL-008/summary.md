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
