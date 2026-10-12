# BTS-FAST-PUBLIC-API-COVERAGE-004 — durable batch journal

- Status: `BATCH_IN_PROGRESS`; checkpoints A–C passed; checkpoint D not started.
- Charter/pointer bootstrap: Primary `a582dcecf01d55a8a74b734d612f4c5356d46460`, `main`, equal to `origin/main` at last check.
- Secondary implementation: `dcb46a0028af56e6caaa6f3daf8c3b49b53349bb`, normally pushed to official `origin/fast-qprefix`; independent checkout is clean and `HEAD=origin/fast-qprefix`.
- Original Secondary checkout: `fast-qprefix` at `0b922dc48315128a27db811a5f237e0a6c9813d6`; five authorized historical dirty files remain untouched. Tracked-diff SHA-256 remains `a052b913ee6e575d88b3cff5416ac21477e563bbbcd3536ad7014ab2f4f64a60`.
- Task 003 is immutable: its one failed Bootstrap attempt, raw log, summary, and journal were not modified. Batch 004 Bootstrap calls: **0 / 2**; no reservation exists.

## Checkpoint A–C

The source-backed API audit, shared-kernel repair, regression coverage, tests, known full-suite baseline comparison, and retained full-Q/QP hot-path classification are in [the compact Batch summary](BTS-FAST-PUBLIC-API-COVERAGE-004-summary.md).

### BTS input manifest

The unchanged source/config manifest includes all `external/BTS` `*.go`, `go.mod`, `go.sum`, and `run.sh` files (13 files; intentionally excludes the pre-existing `btstest` executable). Sorted per-file SHA-256 manifest digest:

`53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`

The pre-existing `external/BTS/btstest` is 11,277,285 bytes and was not used or overwritten. `external/BTS/` is not part of the Primary Git index.

## Checkpoint D — durable reservations

No attempt is currently reserved or launched. Before attempt 01, append a reservation here and push it to Primary `origin/main`. It must record, together:

- attempt index and UTC reservation time;
- the source manifest digest above;
- exact Secondary backend commit and verified official remote HEAD;
- task-local alternate modfile path and exact `replace` target;
- exact build and No-Trace run commands;
- resulting executable SHA-256;
- unique complete stdout/stderr log path;
- cumulative invocation count `0 -> 1` of maximum 2.

Do not run the executable until that reservation commit is pushed and remote-verified. A started invocation consumes the reservation even if it fails; never rerun the same binary. Attempt 02 is allowed only under the narrow charter condition of a new explicit ordinary public API dispatch gap, after a distinct Secondary fix/test/commit/push, distinct binary SHA, and a separately pushed reservation.

### Attempt 01

`NOT_RESERVED — NOT_RUN`

### Attempt 02

`NOT_RESERVED — NOT_RUN`
