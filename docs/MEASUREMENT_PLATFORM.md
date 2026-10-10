# Measurement Platform — durable single source of truth

**Status:** ACTIVE ARCHITECTURAL RULE. This document must survive chat/session handoffs and govern all future CKKS Standard/Fast performance, attribution, and numerical comparison tasks. **Read with Primary `AGENTS.md` before a measurement-related spec, not just after a profiler problem appears.**

## 1. The existing measurement platform MUST be reused

| Owner / code | Actual reusable capabilities | Intended role |
|---|---|---|
| `cmd/perfprobe/` | Genuine Standard vs Fast adapter separation, multi-run public Bootstrap wall/alloc measurement, warmups, stage timing, formal input provenance, decoded metrics, pairwise `compare`, summary documents | **Default entry point for comparable Standard/Fast performance measurements** |
| `cmd/fastdiag/` and `scripts/fastdiag` | `trace/compare/numerical` orchestration, named stage/power/rescale tracing, provenance, paired diagnostics, existing safe worktree/ref handling | **Default diagnostic/attribution platform**, not a disposable example |
| `internal/perfmeasure/` | Parameter construction, exact Q/P metadata, canonical workload input/fingerprint and effective-profile evidence | Shared config/provenance and workload helpers |
| `internal/numericalmetrics/` | Complex-vector SNR/RMSE, finite/special-case handling | Shared numeric metrics; do not duplicate incorrectly |
| Root `fast_measurement*.go`, `stage_runner.go`; Secondary `internal/fastdiag/` and focused benchmarks | Existing light/full capacity/CRT trace, invariant checks, internal stage tracing, profiling hooks, regression fixtures | Reuse by scope, preserve measurement-only separation |
| `results/QPREFIX-PERF-DIAG-001..007`, `QPREFIX-PERF-OPT-001..004`, `FAST-STANDARD-PERF-REBASELINE-002-*` | Historical measured evidence, attribution methodology and known failed optimization candidates | Prior art **within this project**, not current apples-to-apples numbers |

**Preflight checklist (mandatory):** inventory these locations and read the relevant current implementation and tests; map each requested metric or oracle to reusable functions; document only true capability gaps and their minimum extension points. New one-off `tools/batchNNN/*` timing/metrics/stats code is **prohibited by default** when the same capability already exists. A workload-specific adapter or small fixture may be added only when it delegates measurement/provenance/metrics to these existing modules; any unavoidable new helper should land in the shared platform, with tests and a documented reason. Reuse existing numerical checks; never infer successful comparison from empty vectors.

## 2. Legacy assumptions are NOT the current research contract

The existing platform contains valuable functionality alongside **obsolete assumptions**. Repair the platform instead of building around it.

- `internal/perfmeasure/profile.go`: `ParametersFromConfig` historically hard-coded `EphemeralSecretWeight=0`. The tool MUST accept an **explicit experiment E**, not force zero. For the frozen accepted public workload select E=32 identically for Standard and Fast; other E values require an explicit separate profile and backend acceptance checks. Currently Fast public source admits E=0 or E=32 (not arbitrary E). E is a Standard ephemeral-secret/KeySwitch parameter, whereas Fast zero-secret execution bypasses that dense/sparse KeySwitch arithmetic. Do **not** quietly mutate frozen original experiments or change historical data; legacy E=0 only under a specifically labelled legacy diagnostic profile.
- `cmd/perfprobe/input.go`: the historical Fast `fast_zero_secret_direct_encoded_v1` input is encoded `c0` with manually zeroed `c1`; that is **not** the accepted current public `rlwe.NewEncryptor(...).EncryptNew` lifecycle. Remove the direct-c0-only requirement from **formal public mode**. Keep it, if needed, only as an explicitly selected, labelled *legacy diagnostic* mode. The old tool's compatibility promise is reproducibility of **historic recorded experiments**, not a new Fast requirement or permission to force E=0.
- `cmd/perfprobe/backend_fast.go`: historical `bootstrapping.NewFastEvaluator` bypasses the currently tested ordinary public `GenEvaluationKeys→bootstrapping.NewEvaluator→Bootstrap` dispatch. Public mode MUST use the ordinary public constructor in both backends and the same application operation calls; backend/build adapters may select dependency/workspace only.
- `cmd/fastdiag/numerical.go` and other fixed P93 tests can pin E=0, earlier Fast commit or direct-c0 assumptions. Keep their historical meaning and regression fixtures intact, but add a separate named current E32/public lifecycle profile and source-backed semantics before advertising current public integration support.
- Existing CLI constraints (such as `perfprobe` insisting on `warmup>=1, repetitions>=7, standard-trials>=3`) are unsuitable for a charter allowing at most one or two expensive calls. Provide **explicit bounded research-run policy** and fail-closed attempt journal; never accidentally expend calls due to old warmup defaults.
- Existing default profile limits and hard-coded backend commits must not forbid a valid **approved pinned** newer task; provenance must still fail closed on mismatched specified pins, Q/P/inputs or dirty checkout. No silent fallback to an unpinned/latest Standard baseline.
- **Important semantic distinction:** Fast `c1=0` may remain the correct property of the **current intentional zero-secret backend mode**. The Fast arithmetic does not need E=0: it skips the ephemeral-secret KeySwitch, whereas Standard may use E to determine key switching/cost. Preserve and validate the frontend E rather than replacing it with zero. Current source-level Fast acceptance is **E in {0,32} only** until explicitly extended; do not assert support for all possible E. What is obsolete is forcing every input to be manually constructed/direct-encoded `c0` or assuming **E must equal 0** as universal tool policy. Do not add a fake nonzero Fast c1, alter Fast zero-secret mathematical semantics, or claim secure FHE. The Secondary `AGENTS.md` explicitly describes zero-secret as a **mode**, not a permanent invariant.

## 3. CRITICAL: the platform lives in PRIMARY, and Standard stays pristine

The real repository layout is **two separate modules**:
- Primary: `xuejin-lu/heart-lattigo-bootstrap` with independent Go module `github.com/xuejin-lu/heart-lattigo-bootstrap`. All `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure` and `internal/numericalmetrics` source and test code remains here, never in Lattigo.
- Secondary: `github.com/tuneinsight/lattigo/v6` dependency compiled from **two clean, detached, immutable checkouts**: genuine Standard pin and Fast production pin (doc-only changes on active Fast branch do not change this frozen code pin).

Primary `go.mod` historically has `replace github.com/tuneinsight/lattigo/v6 => ../lattigo`. That path controls which **library** is linked: it does NOT move or delete the Primary measurement tools. The correct automation should create two **task-owned isolated temporary Go modfile/build configurations**, each with an explicit `replace ... => <absolute pinned checkout>`, then build the *same Primary source* into distinct Standard and Fast binaries. For example, use Go's `-modfile=<task-temp-alternate.mod>` and `GOWORK=off` as appropriate. Do **not** rewrite the committed Primary `go.mod`, swap the live `../lattigo` branch between executions, alter pristine Standard source, or allow build tags to change the application math. Probe `go list -m`, module replacement paths, git SHAs and source hashes before measurement; freeze/verify output binaries' module provenance and both executable identities. The exact safe build approach must be validated by tests before taking formal samples.

**Measure at two layers and never equate them:**
1. **Comparable public boundary:** Common Primary timers/alloc counters around **both** original Standard and Fast public methods (`NewEvaluator`, `Bootstrap`, crypto lifecycle, public primitive stages). This needs **no** private Standard hooks. Output metrics, phase labels, numeric oracle, provenance and resources are comparable once cold/warm setup is normalized.
2. **Fast internal attribution:** Secondary `internal/fastdiag` and Fast-specific stage/power/rescale hooks do **not** exist in the unmodified original Standard commit. These are Fast-only diagnostics. Original Standard may expose some public stage methods, but replaying them is not automatically the same measurement as the in-circuit `Bootstrap` hot path. Label unavailable internal stages and any separately replayed public stages as noncomparable, **never** patch Standard or invent equivalent internal hooks to fake symmetry.

Historical `perfprobe` already supports backend adapters and timed public Bootstrap, but its pinned Fast SHA and explicit `NewFastEvaluator` must be repaired for accepted current public experiments. Update adapters in Primary and share measurement/metrics code. The default benchmark CLI historically invokes warmups, repetition and additional numerical Bootstrap trials after timing; all of those count against a task's expensive-call budget. Ensure a zero-call preflight path can run, and future small-cap timing cannot silently invoke those legacy implicit operations.

## 4. Current accepted formal public workload

The Batch019–020 E32 public lifecycle is the acceptance fixture: original genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, accepted Fast `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, LogN13/E32, exact canonical Q/P and deterministic 4096-slot message, A/B Scale `2^45`, C Scale `q5`, Public `EncryptNew → AddNew → MulRelinNew → Rescale(q5) → RotateNew → DropLevelNew(4) → E32 Bootstrap → native low-Level DecryptNew/Decode`. Supported high-Level compact output only q0..q3, fixed width `w_Q(ell)=min(ell+1,4)`; target q0 Drop needs independent `2B<q0`. This fixture is NOT authorization to treat arbitrary high-Level compact `DecryptNew` as supported.

A correct formal comparison needs **genuine original Standard native encryption/decryption**, the same public application source, actual parameters and plaintext, real Fast public dispatch, approved zero-secret mode, matching oracle/method/call budget, numerical and row/capacity gates. Standard ciphertext coefficients/keys need not equal Fast. Build-tag adapters are measurement backend plumbing, not permission to change application algorithms.

## 5. Platform repair precedence and anti-regression

1. When the current tool cannot represent accepted public semantics, create a **platform-repair integration task first**; do not side-step via another standalone stopwatch/metric implementation.
2. Prefer extending `perfprobe` / `perfmeasure` public-mode parameters, native lifecycle and composable stage measurements; reuse `fastdiag` diagnostics only when hooks and parameters genuinely support the profile. Do not silently call legacy `NewFastEvaluator` and label it the same public API.
3. Allow *explicit* profile/lifecycle modes (e.g. `legacy-direct-e0` versus `public-native` with a separate E parameter, including the accepted E=32 profile) with versioned schemas and names, not silent changes in meaning. Existing historical outputs/config fingerprints remain immutable. Mode selection must not be a runtime Fast-vs-Standard branch inside production CKKS application source.
4. Regression tests must cover: historical diagnostics still labelled and intact; E32 exact parameter/QP/ciphertext lifecycle; formal input native provenance; Fast zero-c1 mode correctly observed rather than manually forced; public dispatch; absence of missing vectors/NaN/Inf false pass; default warmup call budget fail-closed; matched tags/SHAs/scale/levels; separately labelled cold construction/first/warm Bootstrap; zero expensive Bootstrap in unit tests.
5. A chat handoff must explicitly include **existing platform inventory, current unsupported legacy assumptions, repaired status and remaining gaps**, not merely a list of last Batch commits. If the platform is broken, repair it **before** more performance exploration.

## 6. Workflow reference

Primary `AGENTS.md` mandates this file; `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B mandates reusable architecture and bounded independent Batch reviews. Code in Secondary remains authoritative for Fast mathematical semantics; no platform-doc text authorizes changing backend crypto mathematics. The active task pointer is `CURRENT_TASK.md`.

## 7. Batch021 repair status — 2026-10-10

Batch021 repaired the existing Primary measurement platform in place. It does not authorize or claim a Bootstrap result or timing measurement.

| Capability | Current state | Reusable implementation / evidence |
|---|---|---|
| Explicit E profile | Available; `ParametersFromConfig` remains the historical E=0 entry point, while `ParametersFromConfigWithE` records the selected E. The formal frozen fixture is E=32; pinned Fast accepts only E=0 or E=32. | `internal/perfmeasure/profile.go`; regression tests in `profile_test.go` |
| Formal public pre-Bootstrap lifecycle | Available for the frozen LogN13/E32 fixture in the existing `cmd/perfprobe` command. It uses ordinary `EncryptNew`, `AddNew`, `MulRelinNew`, `Rescale(q5)`, `RotateNew`, `DropLevelNew(4)`, native Standard decrypt, and Fast authoritative-prefix observation with a native terminal Level0 decrypt. | `cmd/perfprobe/public_native.go`; Batch021 paired artifact |
| Fast public evaluator dispatch | Constructor dispatch is verified through ordinary `bootstrapping.NewEvaluator`; the actual `Bootstrap` method was not called. | Build-tag adapters in `public_backend_fast.go` / `public_backend_standard.go` |
| Bootstrap attempt budget | Explicit process-wide limit, fail-closed before each invocation, and exclusive per-attempt reservation journal. The Batch021 acceptance path fixes the budget to zero. Unit tests cover zero, one, two, failed-call consumption, restart collision, and over-budget refusal without invoking a cryptographic Bootstrap. | `cmd/perfprobe/main.go`; `main_test.go` |
| Fast compact rows | The native Fast `EncryptNew` fixture initially retains its full physical rows (as in accepted Batch019 evidence); compact `AddNew` and subsequent Fast outputs must have dormant rows above `w_Q(Level)=min(Level+1,4)`. | `public_native.go`; per-checkpoint physical row lengths in Batch021 evidence |
| Public Bootstrap acceptance/output and cold/warm timing | Not verified in Batch021 (its budget was zero); subsequently verified in Batch022 with exactly one cold and one warm call per backend. | See Batch022 status below |
| Fast internal stage/power attribution for formal E32 | Existing `cmd/fastdiag`, `internal/fastdiag`, and Secondary tracing remain reusable for their supported Fast diagnostic profiles, but their fixed P93/E0 assumptions and internal-only checkpoint topology are not interchangeable with this E32 public lifecycle. No Standard internal trace was fabricated and no Secondary code was changed. | Reuse existing Fast-only diagnostics only when a task explicitly accepts their profile and non-comparable label |

### Batch021 reuse map and migration boundary

- `cmd/perfprobe/main.go`: retained the existing timing engine, stage replay, result writer, and backend adapter boundary; extended invocation limits, phase labels, and immutable attempt reservations. Public-native does not call the legacy Bootstrap timing path.
- `cmd/perfprobe/input.go`: retained the existing input provenance/validation helpers and explicit legacy direct-c0 diagnostic. The historical Fast input contract pin remains accepted, alongside only the task-approved Fast SHA; arbitrary SHAs still fail closed.
- `cmd/perfprobe/compare.go`: retained the existing numerical pair-comparison machinery and v2 legacy support; v3 records must be explicitly labeled `legacy-diagnostic`. Public E32 checkpoint comparison is a separate mode in the same command and rejects mismatched pins, source/config/input/workload hashes, checkpoint coverage, non-finite/empty vectors, and state/numerical mismatches.
- `internal/perfmeasure/`: reused configuration loading, deterministic input and fingerprinting. The only new workload helper is the shared Batch019 A/B/C public fixture plus exact coefficient-bound/Q0123/q0 capacity gates; its workload SHA is checked against the accepted Batch019 value.
- `internal/numericalmetrics/`: remains the common SNR/RMSE implementation for existing reports. Batch021's checkpoint comparator reuses existing `compareVectors`; it does not introduce another metrics package.
- `cmd/fastdiag` and Secondary `internal/fastdiag`: left unchanged. The E32 zero-Bootstrap task cannot claim their historical P93/E0 in-circuit traces as E32 Standard/Fast stage evidence; that is a documented diagnostic gap, not a reason to duplicate or silently relabel traces.

The dual-checkout preflight compiled and ran the same committed Primary measurement source against clean detached Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and Fast `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` checkouts through isolated temporary modfiles. The original Primary `go.mod` was not changed. Both public E32 runs recorded zero actual Bootstrap calls and passed all eight pre-Bootstrap checkpoints at the fixed `1e-6` gate; see `results/FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021-summary.md`.

**Historical Batch021 boundary:** at the time, its zero-call budget left public Bootstrap acceptance and timing unverified. That one-call follow-up was superseded by Batch022's explicitly authorized two-call-per-backend cold/warm measurement below; do not repeat the one-call smoke against the same batch budget.

## 8. Batch022 public-native cold/warm status — 2026-10-10

Batch022 extended this same `cmd/perfprobe` path; it did not add a standalone runner or change production arithmetic. The frozen LogN13/E32 public lifecycle is now verified through ordinary `bootstrapping.NewEvaluator(...).Bootstrap` for both approved implementations. Primary source is `37002ff3bbd97889cf0ec8108ff2cd527ab8d3d5`, genuine Standard is `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, and Fast is `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.

### Integrity and bounded execution

The shared `compare-public` evidence boundary now binds each canonical checkpoint to its separate decoded-vector artifact: exactly 4096 finite complex samples, matching `DecodedSHA256`, recomputed plaintext-oracle metrics/SNR, exact checkpoint coverage, metadata, row authority, pinned provenance, and Q0123/q0 capacity evidence. Cold/warm outputs receive the same hash/finite/oracle/state checks and are compared directly between Standard and Fast at the fixed `1e-6` maximum-complex gate. Unit tests use fake calls only; no test invokes cryptographic Bootstrap.

The existing `public-native` zero-call mode remains available and still requires `--bootstrap-budget=0`. For this profile, first create clean Standard and Fast artifacts and pass them through `compare-public`. Only then may the explicit measurement mode be run, with its passing pair digest supplied and `--bootstrap-budget=2`. It performs exactly one cold and one warm call on independent copies of the same held Level0 ciphertext from `EncryptNew → AddNew → MulRelinNew → Rescale(q5) → RotateNew → DropLevelNew(4)`. Each reservation is written exclusively before invocation; pre-existing result/vector/journal paths fail before measurement. There are no implicit warmups, repetitions, or retries.

Example shape (use task-owned isolated modfiles and exact pinned checkout paths):

```sh
GOWORK=off go run -modfile=/tmp/standard.mod -tags=perf_standard ./cmd/perfprobe \
  --mode=public-native --profile=logn13 --config=configs/bootstrap_config.logN13.json \
  --out=/tmp/standard-preflight.json --vectors-out=/tmp/standard-preflight-vectors.json \
  --backend-commit=5dbffbdea05394de2ca3a432ed5318aa832e3f40 \
  --backend-ref=5dbffbdea05394de2ca3a432ed5318aa832e3f40 \
  --secondary-root=/path/to/clean-standard-checkout --ephemeral-secret-weight=32 --bootstrap-budget=0

GOWORK=off go run -modfile=/tmp/standard.mod ./cmd/perfprobe compare-public \
  --standard=/tmp/standard-preflight.json --fast=/tmp/fast-preflight.json \
  --standard-vectors=/tmp/standard-preflight-vectors.json \
  --fast-vectors=/tmp/fast-preflight-vectors.json --out=/tmp/preflight-pair.json

GOWORK=off go run -modfile=/tmp/standard.mod -tags=perf_standard ./cmd/perfprobe \
  --mode=public-native --public-bootstrap --preflight-pair=/tmp/preflight-pair.json \
  --profile=logn13 --config=configs/bootstrap_config.logN13.json \
  --out=/tmp/standard-bootstrap.json --vectors-out=/tmp/standard-bootstrap-vectors.json \
  --backend-commit=5dbffbdea05394de2ca3a432ed5318aa832e3f40 \
  --backend-ref=5dbffbdea05394de2ca3a432ed5318aa832e3f40 \
  --secondary-root=/path/to/clean-standard-checkout --ephemeral-secret-weight=32 --bootstrap-budget=2
```

Fast uses the equivalent explicit invocations below, with the isolated modfile replaced by the pinned Fast checkout. After both backends complete, `compare-public` is run again on the Bootstrap result/vector artifacts. Never reuse the same output prefix after any attempt reservation; an interrupted/failed reservation consumes that attempt.

```sh
GOWORK=off go run -modfile=/tmp/fast.mod -tags=perf_fast ./cmd/perfprobe \
  --mode=public-native --profile=logn13 --config=configs/bootstrap_config.logN13.json \
  --out=/tmp/fast-preflight.json --vectors-out=/tmp/fast-preflight-vectors.json \
  --backend-commit=2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac \
  --backend-ref=2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac \
  --secondary-root=/path/to/clean-fast-checkout --ephemeral-secret-weight=32 --bootstrap-budget=0

GOWORK=off go run -modfile=/tmp/fast.mod -tags=perf_fast ./cmd/perfprobe \
  --mode=public-native --public-bootstrap --preflight-pair=/tmp/preflight-pair.json \
  --profile=logn13 --config=configs/bootstrap_config.logN13.json \
  --out=/tmp/fast-bootstrap.json --vectors-out=/tmp/fast-bootstrap-vectors.json \
  --backend-commit=2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac \
  --backend-ref=2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac \
  --secondary-root=/path/to/clean-fast-checkout --ephemeral-secret-weight=32 --bootstrap-budget=2
```

### Batch022 measured result and remaining boundary

The zero-call preflight passed all eight canonical checkpoints with matching Level/Scale/Degree and matched Go/OS/CPU/runtime environment. Maximum pre-Bootstrap Fast-vs-Standard difference was below `7.8e-11`. Exactly two actual calls per backend were reserved and completed. Both outputs were 4096-slot native Level1 decryptions at Scale `2^45`, and every self-oracle and direct paired comparison passed. Cold and warm Fast-vs-Standard RMSE were both `4.7262101606249815e-9`; maximum complex difference was `4.489543437647087e-8`, below `1e-6`.

The one cold and one warm sample are descriptive, not a stable speedup estimate. The pinned Fast source explicitly initializes its circuit lazily on first Bootstrap (`fast_bootstrap.go:18–40`), whereas Standard constructs its circuit data inside `NewEvaluator` (`evaluator.go:169–178`). This explains why the observed Fast constructor was short and its first call much larger than its warm call; report the phases separately and do not present the first call as steady-state latency. Fast's E32 in-circuit/internal stage tracing remains unavailable for formal Standard-comparable attribution. Fast E=32 is still the intentionally insecure zero-secret simulation mode; this result makes no security-equivalence claim.

Compact result, journal, and provenance/aggregate evidence are in `results/FAST-DROPIN-PUBLIC-NATIVE-COLD-WARM-AUTONOMOUS-BATCH-022-{summary.md,journal.md,evidence.json}`. Raw 4096-slot vectors and temporary alternate modfiles remain outside the repository.

## Batch024 E32 trace status — final diagnostic outcome (2026-10-11)

S1–S3 and the final zero-call gates passed. After the Web-approved fix to
identify an isolated diagnostic clone by filesystem identity rather than
lexical path strings, `go test ./cmd/fastdiag -count=1`,
`go vet ./cmd/fastdiag`, `git diff --check`, and the Secondary
`FASTDIAG_VALIDATE_ONLY=1` E32 fixture test passed. The verified held fixture
uses the Batch023 six-output repeatability vectors (SHA-256
`63ad89ce9ccc70a0b737397810d4f4a15d8a1ed65e82ccd3af0435d1df4d6881`) and
manifest SHA-256
`5bc19b1eb6e67997ca3d19f39e6fd781bf93b27ad305a1be733f0c883611b847`.

The approved Secondary production delta was limited to guarded event spans in
`schemes/ckks/internal/fastcore/rescale.go`, around the existing
RescaleWorkspace.ApplyRows materialization and actual per-row NTT/MForm
restore work; it did not alter arithmetic or execution order. The Primary
analysis repair retained per-event closure/Pareto and renderer/test
interfaces, adding event-type summaries based on parent-minus-direct-child
elapsed time and partitioned by event-tree root.

The one authorized S4 session reserved and completed exactly two Fast-only
Bootstrap calls: one cold initialization and one traced warm call. Both
decoded outputs passed the native oracle (RMSE `1.1795785978778852e-9`,
maximum complex difference `5.738790672737443e-8` against `1e-6`) and matched
the held Fast reference. Structural validation then failed at the first
nonconforming power parent: event sequence 61 (`T8`) points to sequence 60
(`T16`), whereas the validator requires direct parent sequence 59
(`generated_powers`). The 539-event raw trace is retained outside Git with
SHA-256
`f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74`; it is
`TRACE_UNVERIFIED`, with no certified event attribution, Pareto/Amdahl
analysis, or performance claim. The two calls exhaust Batch024's budget;
there was no retry. Batch023's separately accepted 4.225x observation remains
historical and is not a Batch024 finding. Its 14/14 calls were not reused.

Original Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and original
production Fast `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` remain frozen; the
diagnostic checkout was pinned to Secondary commit
`4f2557062cb5c1ffb9a671bc3df67401fd7092b4`. No Secondary production code was
changed during the final diagnostic session. The observed elapsed values are
raw diagnostic data only and must not be used as a speedup estimate.

Final state: `BATCH_BLOCKED_NEEDS_WEB_REVIEW`. See the compact Batch024
summary, journal, and evidence JSON for full provenance, attempt journals,
failure sidecar hash, and the exact first structural mismatch.

Checkout-status qualification: the original Standard commit object
`5dbffbdea05394de2ca3a432ed5318aa832e3f40` is present, but its registered
detached worktree is currently unavailable/prunable, so a fresh clean-status
check could not be collected. No Standard run or source modification occurred
in this final session. The original Fast pinned checkout and the task-local
diagnostic clone were both clean at final read-only verification.
