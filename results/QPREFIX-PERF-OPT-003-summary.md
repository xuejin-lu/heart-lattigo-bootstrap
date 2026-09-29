# QPREFIX-PERF-OPT-003 — Row-Major Source INTT Batching

## Decision

`SOURCE_INTT_BATCHING_CORRECT_BUT_NO_WIN`

The row-major candidate was bit-identical to the existing component-major conversion, but its rows4 median was 0.36% slower. This is below the required 3% feasibility improvement, so no production change was retained. Full Rescale, E2E, and power-trace before/after measurements were not run because the feasibility gate failed.

## Provenance and implementation state

- Primary task pointer: `QPREFIX-PERF-OPT-003`, accepted parent Primary commit `7460f1cc4ff1ef6fcd3cfd26994ccaec7147b0bd`.
- Secondary branch: `fast-qprefix`, current HEAD `2018ae7c74151d7b4de8e0c0bf393d96e18b8181`.
- No Secondary production commit was created and no Secondary push was made. HEAD `2018ae7c` updates the control-plane task pointer; `schemes/ckks/fast/rescale_qprefix.go` remains unchanged from the accepted production implementation at `6930cf6cb3c71ce139a1eb42eede7be335b7174c`.
- Secondary `fast-qprefix` worktree is clean.

## Candidate and correctness

The test-only candidate used the proposed row-major order (`q0: c0,c1; q1: c0,c1; ...`) and performed the same per-row `INTT` and optional `IMForm` operations as the current per-component helper. It prevalidated all source and destination component backing before transforming. The current Rescale path, transactionality, staged result writes, arithmetic, Q-prefix policy, and diagnostic event schema were not changed.

A temporary same-package equivalence test compared every coefficient in rows2 and rows4 for the same LogN13 P93, q0=55, degree-one NTT input. All residues were bit-identical. The temporary harness was removed after collecting the feasibility samples. Existing tests also passed for both NTT representations, higher-degree staging, and capacity-failure transactionality. Since the candidate was not retained, there is no new production aliasing behavior or new alias proof to claim.

## Feasibility benchmark

Command:

```bash
go test ./schemes/ckks/fast -run '^$' \
  -bench '^BenchmarkFastRescaleSourceINTTOrderOPT003$' \
  -benchmem -count=7
```

The default build used `internal/fastdiag/disabled.go`; diagnostics were off. LogN13 P93 input generation and output allocation occurred outside timing. The same degree-one c0/c1 input was used for both strategies. Rows2 is counterfactual only.

| Rows / strategy | 7 samples (ns/op) | Median | Allocs |
|---|---|---:|---:|
| rows4 component-major | 273023, 276575, 273135, 272705, 282918, 272808, 273526 | 273135 | 0 B/op, 0 allocs/op |
| rows4 row-major | 273103, 273102, 278147, 274123, 274507, 275340, 273224 | 274123 | 0 B/op, 0 allocs/op |
| rows2 component-major (counterfactual) | 137509, 136191, 135950, 136031, 135783, 136407, 135955 | 136031 | 0 B/op, 0 allocs/op |
| rows2 row-major (counterfactual) | 135911, 135878, 136461, 138215, 136244, 136134, 136091 | 136134 | 0 B/op, 0 allocs/op |

Rows4 candidate delta: `+0.36%` time (slower), versus the required `-3%` or better. Rows2 candidate delta: `+0.08%` time (slower); it does not affect the decision.

## Conditional measurements and remaining hotspot

The feasibility gate rejected the candidate, so the spec's conditional same-session full rows4 Rescale, diagnostics-off P93 E2E, and low-overhead power-trace before/after runs were not performed. The accepted parent evidence remains the baseline: full rows4 Rescale about `1.130242 ms`; source INTT `273.364 us` (22.21% of phase sum); commit NTT restore `242.367 us` (19.69%); combined transforms 41.89%; INTT stack 23.39% cumulative in pprof.

No hotspot was changed. Source INTT remains the largest measured individual Rescale phase, with commit NTT restore the next significant transform phase.

## Validation

- Temporary `TestFastRescaleSourceINTTRowMajorMatchesComponentMajor`: passed for rows2 and rows4; then the temporary harness was removed.
- `go test ./schemes/ckks/fast -count=1`: passed.
- `go test -tags fastdiag ./schemes/ckks/fast ./circuits/ckks/bootstrapping ./internal/fastdiag -count=1`: passed, including tagged Rescale event-shape coverage.
- `go test ./...` in Secondary: passed.
- `git diff --check` in Secondary: passed.
- Secondary `fast-qprefix` worktree: clean; no production code modified.
