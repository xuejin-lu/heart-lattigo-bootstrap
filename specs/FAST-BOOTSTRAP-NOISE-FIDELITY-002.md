# FAST-BOOTSTRAP-NOISE-FIDELITY-002 — Six-sample Bootstrap fidelity and cost study

**Status:** `ACTIVE_USER_AUTHORIZED` — 2026-10-27. This is a new task with an independent 18-call ledger. It does not reopen, reset, or modify `FAST-BOOTSTRAP-NOISE-FIDELITY-001` or its exhausted 3/3 Bootstrap budget.

## Goal

Using the already accepted Montgomery-correct fixed-sigma Noise Hook, compare full Bootstrap output precision and Cold/Warm performance for genuine Standard, Fast with Noise OFF, and the same Fast binary with Noise ON. Quantify Noise's time and allocation overhead. Keep the frontend workload and measurement platform fixed; make only the minimum Primary harness/comparator changes needed to use the existing six-call repeatability path.

This is exploratory numerical evidence, not a security-equivalence claim or proof that the sampled coefficient model reproduces Standard CKKS noise.

## Source and workload contract

- Primary control plane: `xuejin-lu/heart-lattigo-bootstrap`, `main`.
- Genuine Standard Lattigo: `xuejin-lu/lattigo@5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast Lattigo: `xuejin-lu/lattigo:experiment/fast-bootstrap-noise-fidelity@49b65dd927780cec0d7e05772a3df8554ef58005`, the accepted Montgomery representation repair. Do not modify or use `fast-qprefix` as the experiment branch.
- Standard, Fast OFF, and Fast+Noise use the same Primary application/perfprobe source, frozen LogN13/E32/Bootstrap Scale 2^45 parameters, exact Q/P, canonical original message, and Bootstrap plaintext oracle. Standard is built against the genuine Standard source. Fast OFF and Fast+Noise must use the exact same Fast executable SHA; only the explicit runtime hook option differs.
- Reuse the accepted canonical fixture fingerprints, then require a fresh zero-call pair bound to this task's exact sources before measurement: config SHA-256 `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; Q/P SHA-256 `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`; original input SHA-256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; workload SHA-256 `00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74`; Bootstrap plaintext SHA-256 `c32c0d7b731c0fcd8d82fb9c1182a43a62d120f761c090ce21dc764e9aa1e703`. The preflight must recompute and match each fingerprint; any mismatch stops before a Bootstrap reservation.
- Fixed Noise sigma is `4.004280012097432`; root seed is `20261027`. No tuning, resampling, or changing the seed is allowed.

## Reuse inventory and minimum integration

Verified existing owners:

1. `cmd/perfprobe/public_native.go:runPublicBootstrapAttempts` and `publicBootstrapAttemptSpecs(6)` already execute one `first_cold_bootstrap` plus five named warm calls, on one evaluator lifetime, each using a fresh `input.CopyNew()` whose fingerprint is checked against the same held ciphertext. `perfmeasure.BootstrapAttemptBudget` durably reserves each invocation before calling Bootstrap and rejects reuse/retry.
2. `cmd/perfprobe/main.go` already exposes `--public-repeatability` and enforces an exact budget of six. It currently rejects the fixed-sigma Bootstrap hook in repeatability mode.
3. `cmd/perfprobe/public_backend_fast.go` already configures the library hook and serializes its snapshot. The accepted Secondary hook records each operation in order, independent left/right seeds, sigma, coefficient summary/hash, Level/Scale, strict Q-prefix capacity, and completion; the coefficient vector itself is not serialized.
4. `cmd/perfprobe/public_triplet.go` and `internal/numericalmetrics` already validate three-arm provenance and compute global input-relative RMSE, `-log2(RMSE)`, MaxAbs/p99/p99.9, pairwise vector errors, timing, and gate observations, but the validator/comparator currently accepts only one-call Bootstrap artifacts.
5. `public-native` preflight and `compare-public` already verify genuine Standard lifecycle, exact effective parameters/input/source/build identity, slot count, capacity, decode path, and strict pre-Bootstrap oracle checks. The accepted 001 pin `715ef585…` and historical pins must remain unchanged for their existing modes; 002 gets a separate exact Fast source pin `49b65dd…` restricted to the explicit 002 mode.

Implement only a task-scoped 002 repeatability/preflight mode in the existing `perfprobe` owner, attach the existing Fast hook to the 6-call path, and extend the existing triplet comparator for six samples. Preserve every legacy mode and its gate/pin behavior. Do not create a runner, timer, statistics package, noise sampler, alternate comparator, or Secondary arithmetic implementation.

## Frozen measurement design and budget

Each arm runs exactly one Cold and five Warm Bootstrap calls in one process/evaluator lifecycle, with no hidden warmup, retry, or extra Bootstrap:

| Order | Arm | Hook | Cold | Warm | Maximum calls |
|---:|---|---|---:|---:|---:|
| 1 | Fast+Noise (FN) | ON | 1 | 5 | 6 |
| 2 | Genuine Standard (S) | OFF/not available | 1 | 5 | 6 |
| 3 | Fast (F) | OFF | 1 | 5 | 6 |
| | **Total** | | **3** | **15** | **18** |

The measurement runner's first call is the sole Cold sample. Each of the five following calls is Warm; no `runtime.GC`, hidden warmup, or extra CKKS call is inserted between or before them. Every call is individually reserved and its failure permanently consumes that slot. Any Bootstrap error, invalid evidence, capacity failure, non-finite data, or provenance mismatch stops the entire batch; do not retry and do not start later arms. The task-wide `attempt-ledger.json` has exactly 18 immutable attempt IDs and tracks planned/reserved/consumed/succeeded/failed states. Existing per-call `perfmeasure.BootstrapAttemptBudget` sidecars remain the invocation authority.

The order FN→S→F is fixed so the first post-repair full Bootstrap checks the exact Montgomery path before other arms. Complete, validate, commit, and push each arm's result checkpoint before reserving the next arm. A failed FN Cold call ends the batch with at most one consumed call; no S/F runs follow.

## Build and pre-execution gates — zero Bootstrap

1. Reuse `cmd/perfprobe`, `internal/perfmeasure`, `internal/numericalmetrics`, existing `compare-public`, existing triplet comparator, and the current attempt-budget implementation.
2. Add task-scoped exact pin acceptance for Fast commit `49b65dd927780cec0d7e05772a3df8554ef58005`; do not widen generic Fast allowlists or alter 001/single-call/repeatability historical pin rules.
3. Add regression tests accepting this pin only for 002, rejecting a wrong pin/profile/mode, preserving the old mode's acceptance/rejection behavior, allowing the explicit observational output mode only for 002, and retaining exact budgets/call phases.
4. Run focused Fast Chebyshev power-16 and necessary Fast Mod1/polynomial tests at the accepted profile. Tests must not call complete Bootstrap. Keep strict capacity checks; any failure blocks the task and cannot be bypassed.
5. Commit the final Primary measurement source before binary builds. Build clean, source-pinned Standard and Fast binaries from that exact Primary source commit and clean Secondary checkouts. Verify `go version -m` VCS metadata, module replacement/source identity, build tags, binary SHA-256, `vcs.modified=false`, OS/architecture/host, and equal Fast/FN binary SHA. Do not rebuild after later evidence-only Primary commits.
6. Run the fresh task-scoped zero-call Standard/Fast preflight, then existing `compare-public` with strict pre-Bootstrap checks. Freeze its digest, canonical inputs, exact effective LogN/E/Scale/Q/P, both source commits, binaries, and environment in the ledger/journal. Preflight actual Bootstrap count must be zero.
7. Commit/push the frozen plan, exact 18-slot ledger, binaries/provenance, and passing zero-call pair before any measured arm starts.

## Exploratory output validation

For 002 only, final Bootstrap MaxAbs against `1e-6` is observational: a finite, correctly shaped, correctly sourced output is retained even when it exceeds the threshold. Record `NumericalGatePass` for every output; never label an observed miss as a pass. Preflight oracles and all structural/data checks remain strict: source/binary/preflight identity, exact input/plaintext fingerprints, Q/P/Scale/Level, output degree and NTT/Montgomery contract, Q-prefix capacity, decode path, 4096 finite slots, output SHA, metrics recomputation, Noise event order/seeds/count/hash/Level/Scale/capacity, and exact 1 Cold + 5 Warm result coverage. This observational exception is task-scoped and must not change legacy or 001 semantics.

## Required result schema and analysis

The existing triplet comparator must accept the exact six outputs for each of S/F/FN and validate phase names in order (`first_cold_bootstrap`, `warm_bootstrap_01`…`warm_bootstrap_05`). For each of the 18 outputs, report its output SHA-256, input-relative global RMSE, global precision bits `-log2(RMSE)`, MaxAbs, p99, p99.9, finite/gate status, elapsed ns, allocation bytes, and allocation count. Undefined zero/non-finite precision must be explicit; all output samples must be finite to complete the task. Preserve global precision bits as distinct from any per-slot precision summary.

Report Cold and Warm precision separately. For the five Warm timing values per arm report all five observations plus mean, median, sample standard deviation (`n-1`), minimum, maximum, and mean allocation bytes/count; retain each individual warm allocation result. Report Cold time/allocations separately. Report evaluator construction phases (`public_ckks_evaluator_construction` and `public_bootstrapping_evaluator_construction`), their sum, and construction-sum + Cold Bootstrap elapsed cost. If Fast+Noise hook configuration has a measured phase, report it separately.

Compare per corresponding Cold/Warm sample: Fast and Fast+Noise global precision bits against Standard; report signed and absolute bit gaps and whether Noise reduced the absolute gap. Report Standard/Fast/Fast+Noise output pairwise RMSE separately. Report cold and warm-median Fast speedup vs Standard and FN overhead vs F in ns and ratio, plus paired allocation-byte/count deltas. State whether Noise improved precision and the measured time/allocation cost. No causal claim beyond these observations.

Persist compact raw result documents, per-call attempt reservations/results, triplet JSON, compact summary, journal, source/config/input/profile/binary provenance, environment, and every output SHA under `results/FAST-BOOTSTRAP-NOISE-FIDELITY-002/`. Keep large decoded vectors outside Git if their size warrants it, with exact path, byte size, SHA-256, and their per-sample hashes in Git. Never serialize or commit full Noise coefficient vectors; retain coefficient summaries and SHA-256 values only. Do not place secret material in artifacts.

## Completion and stop conditions

Success status: `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only if all 18 authorized calls complete once, all six outputs per arm pass structural/provenance/finite checks, the six-sample comparator recomputes all metrics, and the final evidence is committed/pushed. The numeric `1e-6` observational threshold is not an acceptance gate for these final outputs.

Stop immediately as `BATCH_BLOCKED_NEEDS_WEB_REVIEW` on any actual call failure, missing attempt evidence, capacity/representation/scale/source mismatch, wrong Fast pin or binary identity, non-finite/malformed output, or any necessary new mathematical/architecture choice. Preserve every prior result/reservation and do not retry. No LogN16, BTS, large benchmark, alternate sigma, or modification to 001 or `fast-qprefix` is authorized.

## Existing task preservation

`FAST-BOOTSTRAP-NOISE-FIDELITY-001` remains an immutable 3/3-call historical pilot with S/F success and the pre-injection FN failure. Its captures, `attempt-ledger.json`, repair record, and result hashes remain unchanged. This task uses only its own spec, directory, and 18-call ledger.
