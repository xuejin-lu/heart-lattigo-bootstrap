# Current Task

Task: QPREFIX-PERF-MEASURE-LOGN16-001
Status: READY_FOR_CODEX

Task class:
`M — Performance Measurement`

Temporarily suspended:
`FAST-STANDARD-NUMERICAL-DIAG-002`

Authoritative spec:
`specs/QPREFIX-PERF-MEASURE-LOGN16-001-STANDARD-VS-FAST.md`

Goal:
Measure matched `N=2^16` / LogN16 bootstrapping timing for genuine Standard versus current Q-prefix Fast, including complete Bootstrap and per-stage timing.

Frozen refs:
- Primary harness: `2b1022e1820e612e9209094dd672be15d06a8d13`
- Standard Secondary: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Fast Secondary: `82601ea2517edc14784c9da250426169a1b221c7`

Config:
`configs/bootstrap_config.logN16.json`

Protocol:
- temporary detached worktrees;
- warmup 1;
- repetitions 7;
- `-stages`;
- same host;
- no production changes.

Required outputs:
- `results/QPREFIX-PERF-MEASURE-LOGN16-001-standard.json`
- `results/QPREFIX-PERF-MEASURE-LOGN16-001-fast.json`
- `results/QPREFIX-PERF-MEASURE-LOGN16-001-summary.md`

Return:
`LOGN16_STANDARD_FAST_TIMING_READY`
then `READY_FOR_WEB_REVIEW`.


Web-review amendment:
- Standard detached harness run is explicitly authorized to use `-tags lattigo_standard`.
- Run `go test -tags lattigo_standard ./...` before the Standard measurement.
- Standard measurement uses `go run -tags lattigo_standard . ...`.
- Fast measurement remains untagged.
- This tag only selects the Primary harness's built-in Standard stubs and excludes Fast-only diagnostics; it does not authorize arithmetic or config changes.


Web-review amendment 2:
- Supersede prior current-Primary/build-tag execution attempts.
- Use detached measurement harness Primary commit `8186f50e7b591b7f76b39fb89b47c32ad1cc1410` for BOTH Standard and Fast.
- Do not use `lattigo_standard`.
- Standard Secondary remains `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast Secondary remains `82601ea2517edc14784c9da250426169a1b221c7`.
- Run `go test ./...` and then the existing `go run . -stages -config configs/bootstrap_config.logN16.json -warmup 1 -repetitions 7 ...` in each detached harness.
- If current Q-prefix does not compile against this proven clean public harness without patching, stop BLOCKED.
