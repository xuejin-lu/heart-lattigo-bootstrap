# BTS-FAST-PUBLIC-API-COVERAGE-004 — durable batch journal

- Status: `BATCH_IN_PROGRESS`; checkpoints A–C passed; attempt 01 launched and running.
- Latest Primary checkpoint: `d8ed1def2b918f47e2908452fca4dc03ec1b2fa5`, `main`, equal to `origin/main` after A–C report push.
- Secondary implementation: `dcb46a0028af56e6caaa6f3daf8c3b49b53349bb`, normally pushed to official `origin/fast-qprefix`; independent checkout is clean and `HEAD=origin/fast-qprefix`.
- Original Secondary checkout: `fast-qprefix` at `0b922dc48315128a27db811a5f237e0a6c9813d6`; five authorized historical dirty files remain untouched. Tracked-diff SHA-256 remains `a052b913ee6e575d88b3cff5416ac21477e563bbbcd3536ad7014ab2f4f64a60`.
- Task 003 is immutable: its one failed Bootstrap attempt, raw log, summary, and journal were not modified. Batch 004 actual Bootstrap calls: **1 / 2**; attempt 01 is running and must not be relaunched.

## Checkpoint A–C

The source-backed API audit, shared-kernel repair, regression coverage, tests, known full-suite baseline comparison, and retained full-Q/QP hot-path classification are in [the compact Batch summary](BTS-FAST-PUBLIC-API-COVERAGE-004-summary.md).

### BTS input manifest

The unchanged source/config manifest includes all `external/BTS` `*.go`, `go.mod`, `go.sum`, and `run.sh` files (13 files; intentionally excludes the pre-existing `btstest` executable). Sorted per-file SHA-256 manifest digest:

`53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`

The pre-existing `external/BTS/btstest` is 11,277,285 bytes and was not used or overwritten. `external/BTS/` is not part of the Primary Git index.

## Checkpoint D — durable reservations

A started invocation consumes its reservation even if it fails; never rerun the same binary. Attempt 02 is allowed only under the narrow charter condition of a new explicit ordinary public API dispatch gap, after a distinct Secondary fix/test/commit/push, distinct binary SHA, and a separately pushed reservation.

### Attempt 01

`LAUNCHED — RUNNING`; reserved at `2026-10-12T05:04:20Z`, process launched at `2026-10-12T06:18:03Z`.

- Source/config inputs: 13 unchanged `external/BTS` files (`*.go`, `go.mod`, `go.sum`, `run.sh`); sorted per-file SHA-256 manifest digest `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`. Rechecked immediately before reservation. The existing executable `external/BTS/btstest` (11,277,285 bytes; SHA-256 `0908513a6df542922e5355964af8695bf109fe70611663e89d2f9a5eddbae813`) is excluded and was not overwritten or executed.
- Fast backend: Secondary `fast-qprefix` commit **and fetched official `origin/fast-qprefix` HEAD** `dcb46a0028af56e6caaa6f3daf8c3b49b53349bb`; independent checkout clean.
- Alternate modfile: `/private/tmp/bts004-fast-build.4E7vjB/fast.mod`; its only replacement is `github.com/tuneinsight/lattigo/v6 => /private/tmp/lattigo-bts-003-working.HnjCjy/repo`.
- Build working directory: `/Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/external/BTS`.
- Exact build command: `GOWORK=off GOCACHE=/private/tmp/bts004-gocache.ZwwTt1 go build -mod=mod -modfile=/private/tmp/bts004-fast-build.4E7vjB/fast.mod -o /private/tmp/bts004-fast-build.4E7vjB/bts-fast .`
- Binary SHA-256: `ac65e791232a0935400baa81eaee777ab32f7074544dd7e706923e5bbf28bd5d`.
- Exact No-Trace run command (same working directory; default original BTS flags): `set -o pipefail; GOWORK=off /private/tmp/bts004-fast-build.4E7vjB/bts-fast 2>&1 | tee /Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/results/BTS-FAST-PUBLIC-API-COVERAGE-004-attempt-01.log`
- Unique full raw log: `results/BTS-FAST-PUBLIC-API-COVERAGE-004-attempt-01.log`.
- Budget: one attempt used of maximum two; actual invocations `1/2`. No Standard, warmup, Trace, or benchmark invocation. Do not relaunch this binary.
- This reservation must be committed and normally pushed to `origin/main` and verified before launch.

### Attempt 02

`NOT_RESERVED — NOT_RUN`
