# BTS-NOTRACE-WARM-AND-NUMERICAL-008 — resumed measurement

**Status: `TABLE_COMPLETE_AND_OUTPUTS_COMPARED`**

**Bootstrap calls entered in Batch 008: Standard 6/6, Fast 6/6.** The three pre-amendment resource-stop records remain in `journal.json` / `evidence.json`; no budget was reset.

## Required No-Trace Cold/Warm table

| 環維度 | CKKS slots | 實作 | Cold (ms) | Warm 中位數 (ms) | Warm 倍率 |
|---|---:|---|---:|---:|---:|
| LogN16 (N=65,536) | 32,768 | Standard | 103327.916 | 154536.049 | 1× |
| LogN16 (N=65,536) | 32,768 | Fast | 5889.669 | 5044.745 | 30.6331× |

Both table rows are new Batch008 measurements on the same Mac. The earlier resource assessments remain historical evidence but are no longer launch gates. No historical timing was substituted: Task001's 139.653 s Standard cold observation and Task007's 7.369 s Fast single cold observation remain historical single-call notes only and are not samples for this table.

## Historical resource-stop evidence — three records retained

- Host: macOS 26.6.2, Darwin 25.6.0, arm64; physical memory 16 GiB.
- At the preflight snapshot, swap was 5,120 MiB total, 3,748.38 MiB used, 1,371.62 MiB free. `memory_pressure` reported system-wide free percentage decreasing from 55% to 45% during the inspection window.
- Disk had 25 GiB available. Process/file limits were not near exhaustion (`maxproc` soft/hard 2,666/4,000; file descriptors 1,048,575).
- Read-only unified-log evidence at `2026-10-12 23:54:23.317` records kernel `memorystatus: killing largest compressed process bts-standard-decode-007 [66842] 25608 MB`. The Task007 Standard log ends immediately after `NewBTS()` (1m30.373s), before `[enc]` and before `bts.Bootstrap()`, with process exit 137. This is direct evidence of a kernel memory-pressure kill for that prior process, not an inference from exit code alone.
- At the time of the original pre-amendment decision, the available host was this 16 GiB Mac and the historical evidence was treated as a stop condition. That decision is preserved as historical context and was superseded by the user-authorized policy amendment below.

These readings and the prior kernel termination remain observations, not proof of a deterministic memory requirement. The initial and two recovery `RESOURCE_BLOCKED` decisions remain preserved as historical records; the user-authorized amendment below supersedes their former hard-gate policy for this existing Batch.

## Numerical comparison

`STANDARD_FAST_OUTPUT_COMPARED` was reached using the same original input fingerprint and all six newly decoded vectors per backend. Full per-sample metrics and aggregate metrics, including scientific notation, `log10`, and decimal order, are in [numerical-report.md](numerical-report.md) and [comparison.json](comparison.json). Standard native decode is plaintext L18; Fast uses the intentionally insecure zero-secret Q-prefix projection at plaintext L3 from a logical L18 ciphertext. This is an observed real-slot numerical comparison, not security or full-RNS semantic equivalence.

## Reuse, provenance, and integrity

- Read and followed the maintained-platform contract. `cmd/perfprobe` already has an explicit six-call `public-repeatability` policy, but it measures the separate public-native CKKS fixture, not the teammate's original BTS. Task007's existing measurement-only BTS overlay informed the bounded adapter now implemented inside the maintained platform.
- Original `external/BTS` source/config manifest was recomputed over the 13 `*.go`, `go.mod`, `go.sum`, and `run.sh` files: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda` (matches the frozen task pin). The pre-existing executable is excluded. No original BTS file was changed.
- Pinned Genuine Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Pinned Fast source: `ee308c9743000b5f281a84a871e7c4bd59b176c4`. Both measurement checkouts were clean at their exact commits. The old note about five Secondary dirty files is historical; the user confirmed those edits were upstreamed and the current `fast-qprefix` worktree is clean at `ee308c9`.
- Primary startup sync completed by fast-forward to `bccddf0972dd0e4f54594808d384b1dfb34b6ea4`; the worktree was clean before this result artifact was created.
- At the original pre-amendment checkpoint, no Batch008 Bootstrap reservation or measurement process had occurred. The current binaries are built, but the measurement result below is authoritative for subsequent attempt accounting. Historical Batch006/007 records and budgets remain unchanged.

## Validation

- Resource readings: recorded as advisory; the user-authorized policy amendment permits proceeding.
- `GOCACHE=/private/tmp/bts008-gocache.OBnTDE go test ./cmd/perfprobe -run 'Test(BTS|BootstrapBudget|ZeroBootstrapBudget|PublicRepeatabilityUsesSixIndependentCopiesAndReservations)' -count=1`: passed; all Bootstrap callbacks were fakes and no BTS Bootstrap was invoked.
- `GOCACHE=/private/tmp/bts008-gocache.OBnTDE go test ./internal/perfmeasure ./internal/numericalmetrics -count=1`: passed.
- `GOCACHE=/private/tmp/bts008-gocache.OBnTDE go vet ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics`: passed.
- Standard/Fast adapter binaries compiled against their pinned clean checkouts; module provenance and binary SHA-256 checks passed.
- `python3 scripts/task_worktree_policy.py ...`: passed for the three exact task result paths; `python3 -m unittest discover -s scripts -p 'test_task_worktree_policy.py' -v`: 2 tests passed.
- JSON syntax (`jq empty`) and final staged `git diff --check`: passed before commit.

## User-authorized policy amendment and resumed preflight — 2026-10-13

RAM unused, swap availability, `memory_pressure`, and Task007's historical memorystatus termination are retained as observations and are no longer sufficient to auto-block Standard. The current host snapshot: macOS 26.6.2 / Darwin 25.6.0 arm64, 16 GiB physical RAM, `memory_pressure` 49% free, `vm_stat` 7,424 free 16-KiB pages and 352,621 compressor-occupied pages, 23 GiB available disk. `sysctl vm.swapusage`, `top`, and `ps` returned `Operation not permitted` in this environment; those values are reported unavailable rather than guessed. This did not stop the authorized run and no OS memory protection was bypassed.

Adapter SHA-256: `b88f8397f74764775a563f0a72ed3c1ce4fa43761b6cfc3edab1da393add3ccd`. Genuine Standard binary SHA-256: `f3d4f89558abbe1665313b13fea54ade8bf291802cad448a6d70bc9f25934720` (Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`). Fast binary SHA-256: `16d6f8010f6c272b66d0b27524f3508c069174b0b231a81218a4d5ddf2fc0d3a` (Lattigo `ee308c9743000b5f281a84a871e7c4bd59b176c4`). Both use the unchanged 13-file BTS manifest `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`.

## Standard No-Trace Cold/Warm result

One Standard measurement process completed all six reserved Bootstrap calls. `NewBTS()` took 72.000814 s separately. Cold Bootstrap took 103327.916 ms; warm samples were 154536.049, 157511.626, 167255.374, 145448.560, and 133207.716 ms, with median 154536.049 ms. All returned ciphertexts were L18 / Scale 2³² and native Standard plaintext decodes were L18 / Scale 2³² with 32,768 finite slots.

All six decoded vectors had SHA-256 `d934d832c4ae4383268c948fb31e43b38b6127965c9da468881dad1820106c3b`. Standard vs original input: max absolute error `1.15430948e-3`, mean absolute error `1.16583762e-4`, worst slot 27,650. The complete raw log, six `f64le` vectors, run manifest, and per-call reservation/result records are under `standard/`.

## Fast No-Trace Cold/Warm result and paired numerical comparison

One Fast measurement process completed all six reserved calls. `NewBTS()` took 43.319619 s; cold Bootstrap took 5,889.669 ms; warm samples were 5,044.745, 4,956.468, 5,161.159, 4,996.886, and 10,526.021 ms, with median 5,044.745 ms. Warm speedup is 30.6331× (Standard warm median / Fast warm median). All ciphertexts were L18 / Scale 2³²; Fast projection plaintexts were L3 / Scale 2³² with 32,768 finite real slots.

All six Fast vectors had SHA-256 `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b`. Fast vs original input: max absolute error `2.52521213e-5`, mean absolute error `4.53363389e-6`, worst slot 18,078.

Aggregate over six 32,768-slot paired runs (196,608 real values):

- Fast vs Standard: MaxAbs `1.15945733e-3` (`log10=-2.935745`, `10^-3`), RMSE `1.62793260e-4` (`log10=-3.788364`, `10^-4`), MeanAbs `1.16704845e-4` (`log10=-3.932911`, `10^-4`), worst slot 27,650.
- Standard vs original input: MaxAbs `1.15430948e-3` (`log10=-2.937678`, `10^-3`), RMSE `1.62755050e-4` (`log10=-3.788466`, `10^-4`), MeanAbs `1.16583762e-4` (`log10=-3.933362`, `10^-4`).
- Fast vs original input: MaxAbs `2.52521213e-5` (`log10=-4.597702`, `10^-5`), RMSE `5.70847687e-6` (`log10=-5.243480`, `10^-6`), MeanAbs `4.53363389e-6` (`log10=-5.343554`, `10^-6`).

Each run's six repeated vectors was bitwise identical by SHA. Fast's projection plaintext Level 3 differs intentionally from the Genuine Standard native plaintext Level 18; no security or full-RNS equivalence is implied. Full per-attempt metrics are in [numerical-report.md](numerical-report.md); raw logs, run manifests, vectors, and per-call journals remain under `standard/` and `fast/`.

## Recovery resource preflight — 2026-10-13 (UTC 2026-10-12T19:18:45Z)

The previous `RESOURCE_BLOCKED` record is retained; this is a fresh resume check, not a reset of the Batch or its counters. Primary was clean on `main` at `1727c88a9db00be41e127fd91cd9f94ed30e2a3b`, equal to fetched `origin/main`. Secondary was clean on `fast-qprefix` at `ee308c9743000b5f281a84a871e7c4bd59b176c4`, equal to fetched `origin/fast-qprefix`. The earlier journal note about five dirty Secondary files describes the previous run only; no such dirty files exist in this current checkout.

Fresh local resource readings: 16 GiB physical RAM; `top` reported 15 GiB used, 3,598 MiB in the compressor and 87 MiB unused; swap 3,072 MiB total, 1,942.94 MiB used and 1,129.06 MiB free; `memory_pressure` reported 63% and then 62% system-wide free; load averages 2.02/2.26/2.32 with 84.23% CPU idle; disk available 25 GiB; `maxproc` 2,666/4,000 and open-file limit 1,048,575. The largest observed resident process was WebKit WebContent at 2,454,576 KiB; the remaining top processes included another WebKit process at 929,680 KiB and Codex Renderer at 745,216 KiB.

**Decision: remain `RESOURCE_BLOCKED`.** The improved `memory_pressure` percentage and low CPU load are recorded, but the contemporaneous physical-memory/swap headroom is very small. Combined with the directly recorded prior kernel termination of the matching Standard setup at a 25,608 MB footprint, this does not provide a safe basis to launch the Standard-first series on this 16 GiB host. This is a safety assessment, not a claim that 25,608 MB is the workload's deterministic minimum. No measurement process or binary was started; Standard and Fast reservations/entered calls remain 0/6 each. Fast was not run because Standard must be first. No BTS source or backend code was changed.

The exact live readings and both repository synchronization checks were appended to `journal.json` and `evidence.json`. A later resume must perform another fresh resource preflight; historical Batch 008 calls remain 0/6 and are not reset or consumed by this check.

## Second recovery resource preflight — 2026-10-13 (UTC 2026-10-12T21:57:53Z)

This is a new live check after the user's retry request. Primary remained clean on `main` at `895101fc59d8fcf4d90471062a8766e756dcbed9`, equal to `origin/main`; Secondary remained clean on `fast-qprefix` at `ee308c9743000b5f281a84a871e7c4bd59b176c4`, equal to `origin/fast-qprefix`.

The new macOS 26.6.2 / Darwin 25.6.0 arm64 snapshot reported 16 GiB RAM; `top` showed 15 GiB used, 5,444 MiB compressor and 190 MiB unused. Swap was 4,096 MiB total, 2,956.31 MiB used and 1,139.69 MiB free. Two `memory_pressure` reads both reported 50% system-wide free. `top` reported load averages 3.97/2.69/2.23 and 45.50% CPU idle. Disk available was 24 GiB; `maxproc` was 2,666/4,000 and open-file limit 1,048,575. The largest observed resident process was WebKit WebContent at 1,074,480 KiB, followed by another WebKit process at 684,944 KiB and Codex Renderer at 548,192 KiB.

**Decision remains `RESOURCE_BLOCKED`.** This retry's live snapshot shows very little unused physical memory, substantial compression and only about 1.11 GiB free swap; it is less favorable than the prior recovery snapshot. Together with the prior direct memorystatus termination during the matching Standard setup, there is not a safe basis for launching Standard's six-call series. No measurement process or binary was started, no reservation was made, and neither backend's call count changed (Standard 0/6, Fast 0/6). Fast remains unstarted because Standard must run first. No application, BTS, or backend source was modified.
