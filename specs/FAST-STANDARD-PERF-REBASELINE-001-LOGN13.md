# FAST-STANDARD-PERF-REBASELINE-001 — Corrected LogN13 Standard-vs-Fast Timing

## Task
M — Matched performance measurement, no production changes.

## Motivation
The numerical correctness repair is accepted:
- Fast Secondary commit `5117fc57949647182f476dc5952099c706b9f869` (or newer exact fast-qprefix tip if code unchanged).
- Primary reporting commit `fd759d96144d20c4725febb367a8828b8c069e08`.
- Canonical LogN13/q0=55/Q0123: Fast vs Standard RMSE `1.44680895443e-10`, top-level `FAST_STANDARD_NUMERICAL_CLOSE`, 56/56 capacity checks pass.
- Prior LogN16 6.671× speedup was recorded on **older** Fast Secondary `82601ea2517edc14784c9da250426169a1b221c7`. Never present it as current corrected performance.

## Goal
Produce a trustworthy matched-parameter and matched-machine **LogN13** end-to-end Bootstrap latency comparison of current corrected Fast vs genuine Standard, and stage/allocations where feasible, with correctness screening of both outputs.

Do not re-optimize or modify any production arithmetic.

## Source identities
- Fast: pin exact committed Secondary SHA `5117fc57949647182f476dc5952099c706b9f869`. If branch tip differs only in task-pointer metadata, document it; do not silently substitute changed production code.
- Standard: pin genuine Standard Secondary `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, or identify a more appropriate genuine upstream baseline **before execution**, explaining the choice and ensuring no Fast arithmetic in Standard.
- Use a single byte-identical Primary measurement harness commit for both backends. The historical proven cross-backend harness `8186f50e7b591b7f76b39fb89b47c32ad1cc1410` is an available starting point. Confirm it compiles unchanged with both exact Secondary refs; if not, stop as BLOCKED instead of modifying benchmarks differently for each backend. A **single** new common measurement harness commit is allowed only if required for both, with its exact SHA recorded and separately reviewed.
- Both Primary detached worktrees must use identical harness commit, config and commands.

## Parameters and input
- `configs/bootstrap_config.logN13.json` is a candidate, NOT automatic proof of equivalence with the canonical numerical experiment. Record **effective** LogN, LogSlots, Q-chain/P-chain, q0, default scale, Mod1 degree/DoubleAngle/K/log-message ratio, input slots and input digest.
- Require LogN13 and q0=55; do not call a measurement canonical numerical-profile matched unless all effective parameters and input fingerprint also match accepted numerical artifact.
- Genuine Standard and corrected Fast must be run under **identical effective parameters and input**.
- Never mix LogN13 and LogN16, q0=55 and q0=56, or old and new Secondary commits in a headline speedup.
- State explicitly if timing input uses `c0=encoded plaintext-like message,c1=0`: this is a specialized numerical/performance comparator, **not** a general encrypted-input security or functionality test. Compare Standard and Fast honestly under the current zero-a/error-retaining Fast semantics and do not imply equal security.

## Execution
1. Fetch refs; ensure both authoritative worktrees are clean. Do not reset, stash, or checkout authoritative branches.
2. Create two isolated, detached sibling Primary+Secondary worktree pairs so `go.mod` local `replace => ../lattigo` resolves to the correct backend.
3. Verify backend dispatch: Standard public `NewEvaluator` uses genuine Standard, Fast path actually uses FastEvaluator. Confirm Primary benchmark entry point invokes a full public `Bootstrap` for both.
4. Verify identical effective config, input, execution environment, and measurement boundaries.
5. Keygen, key preparation, setup and input preparation are **outside** timed regions; record whether clones/allocations are counted. Do one warmup and >=7 measured iterations per backend when practical. Record raw runs, median, mean, min/max, B/op and allocs/op when instrumentation supports them. Do not cherry-pick.
6. If stages are supported without source changes, record per-stage time and EvalMod real+imag paired sums. Otherwise report stages unavailable rather than making up values.
7. Check both outputs against each other and original canonical input for finite values, expected Level/Scale and numerical quality. If the historical harness does not perform strong numerical screening, explicitly mark timing correctness unverified and do not use it as a release result; do not silently add backend-specific logic.
8. Run appropriate build/tests for each frozen backend/harness. Use the proven no-tag harness setup; avoid historical failed `lattigo_standard` build tag patching. If incompatibility prevents matched measurement, stop and give concrete compile failure.
9. Remove only clean task-created worktrees. Commit/push results in Primary; no Secondary changes.

## Outputs
- `results/FAST-STANDARD-PERF-REBASELINE-001-standard.json`
- `results/FAST-STANDARD-PERF-REBASELINE-001-fast.json`
- `results/FAST-STANDARD-PERF-REBASELINE-001-summary.md`

Summary must include:
- full SHA provenance, host/CPU/Go/OS, input fingerprint, config and effective parameter matching;
- raw timings, median full-bootstrap latency and **Standard median / Fast median speedup**;
- memory allocations, stages/paired EvalMod when available;
- output numerical check and caveats;
- difference from old LogN16 6.671× measurement explicitly framed as **not directly comparable**;
- whether a fresh LogN16 correctness+timing campaign is now justified.

## Stop / acceptance
- `LOGN13_CORRECTED_PERF_REBASELINE_READY` only if identities match, numerical screening passes, and full timings are trustworthy.
- `LOGN13_CORRECTED_PERF_REBASELINE_BLOCKED` for build/compatibility or unmatched backend/config, with precise failure.
- `LOGN13_CORRECTED_PERF_REBASELINE_NUMERICAL_FAIL` if fresh output numerical checks fail.
- Do not modify production to force PASS.

Return the status and `READY_FOR_WEB_REVIEW` with concise Chinese summary.
