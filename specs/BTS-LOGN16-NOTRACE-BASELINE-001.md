# BTS-LOGN16-NOTRACE-BASELINE-001

Status: NEEDS_WEB_REVIEW
Class: bounded experiment / measurement (E)
Authority: user-approved task request, `AGENTS.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md`, and `docs/RESEARCH_ENGINEERING_WORKFLOW.md`.

## Objective

Compile the exact same, unmodified research-lab BTS Go module into two independent executables, linked respectively to genuine Standard Lattigo and the specified Fast Lattigo commit. Run the original BTS program once per backend with its original defaults, and preserve the complete No-Trace console output and numerical results. This is a one-sample-per-backend descriptive baseline, not a repeated benchmark campaign or a security-equivalence claim.

## Immutable workload and backend pins

- BTS source: `external/BTS/`; do not edit, format, add, commit, push, or otherwise alter this directory. Preserve its original `go.mod`, `go.sum`, and every `.go` file.
- Use `main.go` defaults without arguments or changed environment modes: LogN16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0. Preserve its input generation, CKKS parameters, key setup, public API calls, operation order, decode and error calculations.
- Standard source: pristine original Lattigo commit `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast source: `fast-qprefix` commit `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`.
- Build two distinct binaries from the same BTS directory and source using isolated temporary `-modfile` files, `GOWORK=off`, and local module replacements. Do not alter Primary's or BTS's committed `go.mod`, either backend checkout, or runtime selection logic.
- Verify module replacement and each binary's build info/source commit before execution. Both builds must use the ordinary public CKKS API from the unchanged BTS frontend; Fast's zero-secret behavior remains internal to Fast Lattigo.

## Bounded execution and stop rules

- Produce one verified binary per backend. The task budget is exactly one `bts.Bootstrap` invocation per backend, not one successful process launch: if a process is conclusively observed to have exited before the `[NewBTS]` completion marker (which precedes the only Bootstrap call in the unchanged `main.go`), a single resume launch is permitted to complete initialization and make that one call. Record the interrupted pre-Bootstrap launch. If the process state or whether Bootstrap began is ambiguous, stop rather than relaunch. Preserve complete stdout/stderr for every launch by backend under `results/`.
- Do not run warmups, repeated trials, traces, `fastdiag`, Rescale microbenchmarks, Bootstrap diagnostics, or other Bootstrap calls. `NewBTS` setup and the single `bts.Bootstrap` call are timed by the existing original `main.go`; preserve these timings as printed.
- A compile/environment issue may be corrected only in the temporary modfile/build setup and before any Bootstrap attempt. A verified interruption before Bootstrap may be resumed under the preceding rule. Do not change BTS or backend production code to get a passing result.
- If a process reaches `bts.Bootstrap` and errors, panics, or is interrupted, preserve its exact log and count the attempt as consumed. Never rerun that backend. Report the failing API/location when the original log makes it identifiable. A failure requiring Fast mathematical/API semantics is a stop for Web review, not an invitation to patch Secondary.
- Never change parameters or skip a computation. No source or production changes are authorized in Secondary.

## Evidence and report

Produce a compact Markdown comparison report under `results/`, with `No Trace` in the title and table. Include:

- exact Primary/BTS source identity and SHA-256 manifest for the BTS `go.mod`, `go.sum`, and `.go` files (exclude pre-existing executable/build artifacts);
- exact Standard and Fast commit/source identities, clean status, module replacements and executable `go version -m` provenance;
- Go version, OS/architecture, environment, exact build and run commands, and confirmation that the committed BTS source/config/input were untouched;
- per backend: whether `NewBTS` completed and its printed initialization time; whether the first Bootstrap completed and its printed time; input/output Level and Scale; max absolute error, mean absolute error, and max-error slot; exact error or panic if failed;
- performance ratio only when both runs complete and both original decoded outputs are finite and satisfy an explicitly available numerical acceptance criterion. No new tolerance may be invented. If no acceptance threshold exists, report the raw timing pair and state why a correctness-qualified speedup is withheld;
- limitations: one observation per backend is descriptive; Fast zero-secret simulation is intentionally insecure and is not security-equivalent to Standard.

Use only the original BTS timing and error calculations; do not add a runner or a duplicate numerical-metrics implementation. The maintained Primary measurement platform does not own this external custom application workload; preserve its source unchanged and use its existing output rather than forcing it into a different public-native harness.

## Completion

Both exact backend runs and provenance checks were completed, but Fast panicked in C2S before producing a decoded output. Do not call this a successful paired baseline or claim a speedup. The one-call-per-backend budget is exhausted. Preserve the evidence and stop for Web review before any diagnosis that would alter BTS semantics or Secondary production code. Commit and normally fast-forward push only Primary task records, report, and raw run logs; never include `external/BTS/` or change Secondary.
