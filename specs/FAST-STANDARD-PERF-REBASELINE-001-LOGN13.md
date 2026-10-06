# Web-review amendment — MUST reuse canonical numerical framework and report SNR

**This amendment is authoritative and supersedes any weaker “numerical screening optional” wording below.**

## A. Two linked tracks: timings and decoded-domain numerical quality

**The previous timing harness (`8186f50e...`) is only an optional historical starting point for matched timing, not a replacement for our canonical numerical framework.** Existing numerical infrastructure in current Primary:

- `cmd/fastdiag/numerical.go`: `numericalReference`, `fastStandardP93Parameters`, `fastStandardP93Values`, `numericalOutput`, stage-lockstep and threshold/classification gates.
- `internal/numericalmetrics/snr.go`: authoritative complex decoded-domain SNR definition.
- Accepted artifact: `results/FAST-STANDARD-NUMERICAL-FIX-001-GENERATED-POWER-ORDER.json` and `.md`.

For this task, **fresh current-commit SNR and RMSE are mandatory**, not just a reprint of historical `144.71 dB`. Run the canonical `cmd/fastdiag numerical` workflow on the clean corrected Secondary `fast-qprefix` (pin SHA, record exact Primary SHA). Save the fresh raw JSON and generated report under a new task-specific filename, e.g.:

- `results/FAST-STANDARD-PERF-REBASELINE-001-numerical.json`
- `results/FAST-STANDARD-PERF-REBASELINE-001-numerical.md`

Do not overwrite the accepted historical numerical result.

Run at least the prior 2 Fast executions / 3 genuine Standard key trials as currently implemented. Follow the existing CLI and validate required backend, clean worktree, input SHA, effective LogN13 / q0=55 / LogSlots12, output contract, classification and capacity. If the numerical harness fails to compile or run under the frozen corrected Fast commit, report BLOCKED; do not silently substitute old measurements.

## B. Mandatory SNR definition and outputs

Use the **existing** `numericalmetrics.Compare(preDecoded, postDecoded)` definition. For the mode-specific decoded pre-Bootstrap reference vector `x` and post-Bootstrap observation `y`:

`P_signal = mean(|x_i|²)`, `P_error = mean(|y_i-x_i|²)`,
`SNR_dB = 10 log10(P_signal / P_error)` when finite.

- Standard: use the genuine Standard decryption path for each Standard pre- and post-Bootstrap ciphertext; do not decode arbitrary `c0` as if it were generally equivalent to Standard decryption.
- Fast: use the accepted current zero-a/error-retaining Fast decoded-domain method with accurate semantics.
- Record `signal_power`, `noise_power`, `noise_rmse`, `snr_db`, `status` for each mode. Preserve POSITIVE_INFINITY / UNDEFINED_ZERO_SIGNAL / NOT_COMPARABLE status honestly rather than injecting a noise floor.
- Report `STANDARD_BOOTSTRAP_SNR_DB`, `FAST_BOOTSTRAP_SNR_DB`, and `FAST_MINUS_STANDARD_SNR_DB`. Do **not** describe this SNR as RLWE noise budget, cryptographic security, or a proof of secure encryption.
- Also report Fast-to-Standard complex output RMSE and max diff, Fast/Standard complex RMSE against the original message, median precision bits, the final Level/Scale contract and numerical classification.
- Retain stage-lockstep, EvalMod/PS/DoubleAngle checkpoint evidence including stage-reference SNR and its parent-topology-dependent ΔSNR. No need to invent redundant stage SNR experiments; reuse diagnostic output.

Accepted prior baseline for regression reference only (not a fresh measurement):
- Standard SNR `144.713390 dB`
- Fast SNR `144.710750 dB`
- final Fast-vs-Standard RMSE `1.44680895443e-10`
- classification `FAST_STANDARD_NUMERICAL_CLOSE`.

Any newly observed materially degraded numerical classification must cause `LOGN13_CORRECTED_PERF_REBASELINE_NUMERICAL_FAIL` (not an efficiency PASS).

## C. SNR and timing workload must not be silently conflated

Because `configs/bootstrap_config.logN13.json` and the canonical `fastStandardP93Parameters()` are **not automatically identical**, explicitly compare the **generated effective parameters and input digests** of the canonical numerical and timed workload:

1. If truly identical: link current SNR/precision and timing in one matched-workload results table.
2. If different: **do not attach canonical SNR to timed latency as though they describe the same workload**. Either:
   - use a single, identical cross-backend benchmark harness with the exact canonical numerical parameters and deterministic values (and code/commit pinned) for both timing and off-timer numerical measurements; or
   - report them as two clearly named distinct workloads. In the latter case a joined “speed and SNR on the same workload” result is **not achieved**; do not return `LOGN13_CORRECTED_PERF_REBASELINE_READY` as a combined quality/performance result.
3. If a shared harness extension is necessary, permit **Primary-only** shared measurement/diagnostic code, identical for both backends, to collect SNR and actual decoded output. Keep it out of timed `Bootstrap` calls. It must not change the production Secondary or change backend-specific arithmetic. Ensure source/harness commit provenance; report any cross-version incompatibility rather than inventing fallback algorithms.

## D. Measurement boundaries / reporting gate

- Keygen, input encoding, pre-Bootstrap decode, output decode, SNR/RMSE/precision computation, JSON/report output are **all outside the timed Bootstrap region**.
- Warmup and measured Bootstrap iterations remain genuine `Bootstrap` calls, not stage replay timing disguised as full end-to-end.
- Need both raw timing and fresh SNR artifacts. Missing/NOT_COMPARABLE SNR is **not** a completed matched performance+quality result.
- Summary must contain a table with `Standard/Fast`: full Bootstrap median (ms), Bootstrap SNR (dB/status), original-message RMSE, median precision bits; plus the Fast-vs-Standard complex RMSE and relative speedup.
- Explicitly state whether the exact timed and numerical input/parameter fingerprints match.

The task is complete only if both timing and numerical gates succeed; otherwise state the precise blocker and do not make up results.

---

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
