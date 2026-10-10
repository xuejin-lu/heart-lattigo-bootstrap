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

- `internal/perfmeasure/profile.go`: `ParametersFromConfig` historically hard-coded `EphemeralSecretWeight=0`. For the accepted Public API E32 profile it MUST be a supported **explicit** parameter (E=32) with exact Q/P provenance and matching Standard/Fast configuration. Do **not** quietly mutate frozen original experiments or change historical data; legacy E=0 only under a specifically labelled legacy diagnostic profile.
- `cmd/perfprobe/input.go`: the historical Fast `fast_zero_secret_direct_encoded_v1` input is encoded `c0` with manually zeroed `c1`; that is **not** the accepted current public `rlwe.NewEncryptor(...).EncryptNew` lifecycle. Remove the direct-c0-only requirement from **formal public mode**. Keep it, if needed, only as an explicitly selected, labelled *legacy diagnostic* mode.
- `cmd/perfprobe/backend_fast.go`: historical `bootstrapping.NewFastEvaluator` bypasses the currently tested ordinary public `GenEvaluationKeys→bootstrapping.NewEvaluator→Bootstrap` dispatch. Public mode MUST use the ordinary public constructor in both backends and the same application operation calls; backend/build adapters may select dependency/workspace only.
- `cmd/fastdiag/numerical.go` and other fixed P93 tests can pin E=0, earlier Fast commit or direct-c0 assumptions. Keep their historical meaning and regression fixtures intact, but add a separate named current E32/public lifecycle profile and source-backed semantics before advertising current public integration support.
- Existing CLI constraints (such as `perfprobe` insisting on `warmup>=1, repetitions>=7, standard-trials>=3`) are unsuitable for a charter allowing at most one or two expensive calls. Provide **explicit bounded research-run policy** and fail-closed attempt journal; never accidentally expend calls due to old warmup defaults.
- Existing default profile limits and hard-coded backend commits must not forbid a valid **approved pinned** newer task; provenance must still fail closed on mismatched specified pins, Q/P/inputs or dirty checkout. No silent fallback to an unpinned/latest Standard baseline.
- **Important semantic distinction:** Fast `c1=0` may remain the correct property of the **current intentional zero-secret backend mode**. What is obsolete is forcing every input to be manually constructed/direct-encoded `c0` or assuming **E must equal 0** as universal tool policy. Do not add a fake nonzero Fast c1, alter Fast zero-secret mathematical semantics, or claim secure FHE. The Secondary `AGENTS.md` explicitly describes zero-secret as a **mode**, not a permanent invariant.

## 3. Current accepted formal public workload

The Batch019–020 E32 public lifecycle is the acceptance fixture: original genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, accepted Fast `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, LogN13/E32, exact canonical Q/P and deterministic 4096-slot message, A/B Scale `2^45`, C Scale `q5`, Public `EncryptNew → AddNew → MulRelinNew → Rescale(q5) → RotateNew → DropLevelNew(4) → E32 Bootstrap → native low-Level DecryptNew/Decode`. Supported high-Level compact output only q0..q3, fixed width `w_Q(ell)=min(ell+1,4)`; target q0 Drop needs independent `2B<q0`. This fixture is NOT authorization to treat arbitrary high-Level compact `DecryptNew` as supported.

A correct formal comparison needs **genuine original Standard native encryption/decryption**, the same public application source, actual parameters and plaintext, real Fast public dispatch, approved zero-secret mode, matching oracle/method/call budget, numerical and row/capacity gates. Standard ciphertext coefficients/keys need not equal Fast. Build-tag adapters are measurement backend plumbing, not permission to change application algorithms.

## 4. Platform repair precedence and anti-regression

1. When the current tool cannot represent accepted public semantics, create a **platform-repair integration task first**; do not side-step via another standalone stopwatch/metric implementation.
2. Prefer extending `perfprobe` / `perfmeasure` public-mode parameters, native lifecycle and composable stage measurements; reuse `fastdiag` diagnostics only when hooks and parameters genuinely support the profile. Do not silently call legacy `NewFastEvaluator` and label it the same public API.
3. Allow *explicit* profile modes (e.g. `legacy-direct-e0` versus `public-e32`) with versioned schemas and names, not silent changes in meaning. Existing historical outputs/config fingerprints remain immutable. Mode selection must not be a runtime Fast-vs-Standard branch inside production CKKS application source.
4. Regression tests must cover: historical diagnostics still labelled and intact; E32 exact parameter/QP/ciphertext lifecycle; formal input native provenance; Fast zero-c1 mode correctly observed rather than manually forced; public dispatch; absence of missing vectors/NaN/Inf false pass; default warmup call budget fail-closed; matched tags/SHAs/scale/levels; separately labelled cold construction/first/warm Bootstrap; zero expensive Bootstrap in unit tests.
5. A chat handoff must explicitly include **existing platform inventory, current unsupported legacy assumptions, repaired status and remaining gaps**, not merely a list of last Batch commits. If the platform is broken, repair it **before** more performance exploration.

## 5. Workflow reference

Primary `AGENTS.md` mandates this file; `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B mandates reusable architecture and bounded independent Batch reviews. Code in Secondary remains authoritative for Fast mathematical semantics; no platform-doc text authorizes changing backend crypto mathematics. The active task pointer is `CURRENT_TASK.md`.

**Current migration gate:** Batch021 is a shared measurement-platform refurbishment task. It supersedes the previously drafted cold/warm-only Batch021; cold/warm observations resume only after platform public E32 mode is repaired and Web accepts the compatibility evidence.
