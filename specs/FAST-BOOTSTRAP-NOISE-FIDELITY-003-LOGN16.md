# FAST-BOOTSTRAP-NOISE-FIDELITY-003-LOGN16

**Status:** `ACTIVE_USER_AUTHORIZED` — 2026-10-27. This is an independent LogN16/E32 experiment. It does not reopen, reset, borrow from, or modify the 001 or 002 budgets or captures.

## Goal

Extend the accepted fixed-sigma Fast Bootstrap fidelity and Cold/Warm cost comparison from LogN13 to LogN16 using the existing Primary measurement platform. Compare genuine Standard (S), Fast Noise OFF (F), and Fast with fixed-sigma Noise ON (FN), including precision, runtime, allocations, and hook overhead. This is exploratory evidence, not a security-equivalence claim or proof that coefficient noise models Standard CKKS noise.

## Frozen profile and provenance

- Primary: `xuejin-lu/heart-lattigo-bootstrap`, `main`.
- Genuine Standard: `xuejin-lu/lattigo@5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast/FN: `xuejin-lu/lattigo:experiment/fast-bootstrap-noise-fidelity`, starting from `49b65dd927780cec0d7e05772a3df8554ef58005`; record the final task commit. Do not modify or substitute `fast-qprefix`.
- The completed pre-experiment LogN16 Noise Hook/test extension is pinned at Secondary commit `9e540b30898cea74f0fa4acdf03477264344db79`; this is the exact 003 Fast/FN source pin unless a later authorized task checkpoint supersedes it.
- The same Primary measurement source, public CKKS API sequence, config, effective Q/P, workload, held original input, and plaintext oracle are used in all arms. Standard is compiled against the genuine Standard checkout. F and FN use one identical Fast binary; the explicit Noise switch is the only runtime difference.
- Frozen profile: LogN=16, N=65536, LogSlots=15, slots=32768, E=32, Default Scale=`2^45`, Mod1 Scale=`2^60`, degree=30, DoubleAngle=3, K=16, `log_message_ratio=10`; sigma=`4.004280012097432`; root seed=`20261027`.
- Recompute config/input/workload/QP and complete effective parameter fingerprints from the constructed profile and compare to the frozen Task005 LogN16/E32 values. Do not copy historical parameters in place of the current effective values. A mismatch stops before any Bootstrap reservation.
- Preserve full source, build, binary SHA, environment, parameter, level/scale, Q-prefix, slot, input/oracle, and output provenance. Keep decoded vectors external when large and retain byte size and SHA-256 in Git evidence. Never commit complete Noise coefficient vectors.

## Secondary Noise Hook scope

Only extend the already accepted explicit hook from its frozen LogN13 research profile to an explicit allowlist of LogN13 and LogN16. Do not change the historical coefficient-only helper or ordinary Fast arithmetic. Before any Bootstrap, test N=65536 NTT+Montgomery coefficient-domain injection, actual N16 Q/P parameter shape, Level 12 / Scale `2^60`, Q-prefix active rows, strict capacity behavior, left/right independent sampling, square alias safety, input immutability, default-off bit-exactness, and exclusion of ct×pt/scalar/rotation. Reject LogN14/15 and mismatched N explicitly. Preserve the strict `CheckQPrefixAdditionCapacity` gate; do not turn a capacity failure into a pass.

Run focused no-full-Bootstrap N16 Chebyshev power 2/4/16 and required Mod1 regressions. These are narrowly selected tests and must not call `Bootstrap`.

## LogN16 capacity policy

Use only the frozen Task005 policy `logn16-original-e32-observed-native-oracle-v1`: recompute and report the conservative terminal q0 sufficient-bound classification as `NOT_PROVEN`; require the complete exact-profile zero-Bootstrap Standard/Fast native stage-oracle pair, level/scale continuity, source and input fingerprints, finite full-slot output, and all Task005 checks. The historical conservative q0 result is not proof of overflow and is not a capacity proof. Do not relabel it `PASS`, weaken the Noise Hook's strict per-injection Q-prefix check, alter sigma/Scale/Q/P/input magnitude, or infer universal CNN support.

Any actual representation, capacity, Scale, mathematical, backend-dispatch, or source incompatibility stops as `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with current evidence preserved. No task-authorized call may proceed while a required gate is failed or unproven beyond this exact observed-oracle policy.

## Existing platform reuse

Use only the maintained `cmd/perfprobe`, `cmd/fastdiag`/`scripts/fastdiag`, `internal/perfmeasure`, and `internal/numericalmetrics` owners. Extend the existing `public-repeatability` / `runPublicBootstrapAttempts` path and six-sample `compare-public` triplet narrowly for this task. Preserve every 001/002/Task005 mode, pin, artifact, and acceptance rule. Do not add a runner, timer, statistics package, noise sampler, event-tree validator, or alternate comparator.

The task-scoped CLI contracts are `perfprobe --noise-fidelity-003-logn16`, `compare-public --noise-fidelity-003-logn16-preflight`, and the existing six-sample `compare-public-triplet --noise-fidelity-003-logn16`. They are explicit LogN16/E32 entry points; they do not broaden historical pilot or 002 pins.

## Budget and execution order

Create exactly 18 immutable attempt IDs in `results/FAST-BOOTSTRAP-NOISE-FIDELITY-003-LOGN16/attempt-ledger.json`:

| Order | Arm | Calls | Phases | Maximum |
|---:|---|---:|---|---:|
| 1 | FN | 6 | one cold, five warm | 6 |
| 2 | S | 6 | one cold, five warm | 6 |
| 3 | F | 6 | one cold, five warm | 6 |
| | **Total** | **18** | | **18** |

Each arm uses one evaluator lifetime and a fresh independent copy of the same held input for each call. No hidden warmup, retry, extra Bootstrap, BTS, LogN16 auxiliary workload, or unbudgeted profiling call is allowed. Reserve each slot immediately before its call; a failure permanently consumes its slot and stops the batch. Complete, validate, commit, and push one arm checkpoint before starting the next.

Before the first call: finish all targeted no-Bootstrap tests; commit and push the Secondary hook and Primary task-scoped platform integration; build clean Standard/Fast binaries from pinned clean source; verify module replacement, build tags, `go version -m`, VCS metadata, clean state, and binary SHA (F and FN must share one SHA); produce a fresh exact-profile zero-call Standard/Fast preflight pair; verify all 32768 slots, effective Q/P/parameter fingerprints, Scale/Level, plaintext oracle and exact Task005 capacity policy; freeze/push the plan and all 18 reservations. No Bootstrap may run until every gate passes.

## Required measurements and report

For each of 18 outputs, persist output SHA, input-relative RMSE, global bits `-log2(RMSE)`, MaxAbs, p99 and p99.9 of complex slot error magnitude, finite and numerical-gate observation, elapsed ns, allocation bytes/count, and any Noise events. p99/p99.9 refer to slot error magnitude, not latency. Keep finite/structural/provenance/scale/level/capacity checks strict; final `1e-6` is observational only and a miss must remain a miss.

Report Cold separately; retain five individual Warm measurements and report mean, median, sample standard deviation, min/max, and allocation summaries. Include evaluator-construction phases, construction plus Cold cost, Fast vs Standard speedup, FN vs F added ns/ratio and allocation deltas, per-phase signed/absolute precision-bit gaps from Standard, paired output RMSE, and whether Noise moved precision closer. Report hook subphase costs without summing overlapping timers into Bootstrap time. Preserve all output and evidence hashes.

Derive and report for each arm the full effective LogN/N/LogSlots/slots, residual and Bootstrap ring N, E and secret-key mode, Q/P counts and every prime bit size/fingerprint, maximum Level, q0 bits, active Q-prefix rows and physical backing, Default/Mod1 scales, Bootstrap input/output Level, observed internal FN Level/Scale range, and Mod1 degree/DoubleAngle/K/config. Explicitly mark internal Standard/F observations unavailable when the public boundary does not expose them; do not fabricate symmetry.

After all three arms, use the existing comparator to produce a compact summary and compare directly against immutable archived 002 LogN13 results. State differing ring/slot/modulus profiles so no cross-dimension equivalence is implied. Preserve full raw result files, journal, summary, ledger, SHA inventory, and failure evidence.

## Completion

Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only after all 18 calls complete, the existing comparator recomputes every sample, final coding self-review and targeted validation pass, and task evidence is committed and normally pushed to Primary `main` and the authorized Secondary experiment branch. Any gate failure or new scientific/architecture decision returns `BATCH_BLOCKED_NEEDS_WEB_REVIEW`; preserve all prior state and never retry. No Bootstrap occurs in unit tests.
