# FIXED-SIGMA-CHAIN-002 — durable experiment journal

**Status:** `PLANNED_NOT_RUN`
**Date created:** 2026-10-22 (Asia/Taipei)
**Prior reviewed result:** `FIXED-SIGMA-MUL-NOISE-001`: `ACCEPTED_SCOPED_RMSE_ONLY` (not Gaussian/tail equivalence).
**Frozen coefficient σ:** `4.004280012097432` from prior calibration; not to be changed.
**Initial accepted refs:** Primary `167bd0c9c1d61a6f2ff01e37f62565eb123de5dd`; Secondary `9ef0315feaf59dbee03bf0852b72c76fdeb6b800`; genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Runtime commits/binaries must be recorded separately.
**Bootstrap budget:** Standard 0/0, Fast 0/0, used 0; prior Batch budgets are immutable.

| Checkpoint | Status | Notes |
|---|---|---|
| CP-A source/math Level3 gate | PASS_SCOPED | source map/theory/scale/row-authority gate recorded below; no Secondary edit |
| CP-B two-step baseline | IMPLEMENTED_UNMEASURED | shared perfprobe adapter and vector-recomputed pair validator; ≤3 matched pairs reserved |
| CP-C opt-in fixed-σ two-step chain | NOT_STARTED | only if prior gates pass; 2 explicit injections |
| CP-D cost attribution + report | NOT_STARTED | cheap diagnostics; no Bootstrap |

No implementation, measurement, σ adjustment or performance optimization has been run under this new task as of creation. Preserve checkpoints, source hashes, task-scoped budgets and safety rules.

## 2026-10-22 CP-A — source reuse and two-step math gate

**Status: `PASS_SCOPED`; CP-B may proceed.** This reuses the accepted 001 source map and its isolated-operation theory; it does not repeat the full evaluator audit. No CKKS operation, Bootstrap, helper injection, or source edit was performed for this gate.

**Reused source map:** Primary `cmd/perfprobe/public_native.go` owns the native inputs, ordinary public `AddNew`/`MulRelinNew`/`Rescale`, decoded checkpoints, provenance, and `makePublicOperationResidual`; `public_backend_standard.go` and `public_backend_fast.go` only supply build-isolated backend dispatch evidence and the Fast-only opt-in helper. `internal/perfmeasure` constructs effective E32 parameters, exact Q/P metadata, deterministic inputs, and fingerprints; `internal/numericalmetrics` already provides decoded complex-vector statistics. Its current public-native fixture has one Level5→4 MulRelin+Rescale and one residual slot, so the minimum CP-B extension belongs in this maintained path to represent two ordered residuals/checkpoints; no new runner or metric engine is needed.

In Secondary, the 001-audited ordinary `ckks.Evaluator.MulRelinNew` path dispatches to the Fast zero-secret `MulCore.ApplyRows`; ordinary `Rescale` uses the shared Q-prefix `RescaleCore.ApplyRows`. `schemes/ckks/evaluator_fast_fixed_sigma_noise.go` is an explicit diagnostic helper, while `schemes/ckks/internal/fastcore/rescale.go:CheckQPrefixAdditionCapacity` validates a proposed signed c0 addition against the centered lift and verifies each active-row sum. `schemes/ckks/fast/mul_capacity_test.go` already exercises multiplication capacity at Level 3. This is reuse evidence, not a claim that the Level 3 injection profile is currently accepted by the helper.

**Candidate ordinary CKKS schedule and Scale proof:** Reuse the 001 operands for stage 1: `x=AddNew(EncryptNew(A), EncryptNew(B))` at Level 5/Scale `Δ=2^45`, and `y=EncryptNew(C)` at Level 5/Scale `q5`. Then `Scale(MulRelin(x,y))=Δ·q5`; ordinary `Rescale` divides by logical top modulus `q5`, producing Level 4/Scale `Δ`, as already observed in 001. For stage 2, encrypt the same original message `C` at Level 4 with explicit ordinary CKKS Scale `q4`, multiply the stage-1 ciphertext by that ciphertext, and Rescale by logical top modulus `q4`: `(Δ·q4)/q4=Δ`, producing Level 3/Scale `2^45`. This keeps the message values and operation order backend-matched and tests the same coefficient-noise-to-slot gain at both outputs. CP-B must derive `q4`/`q5` from the constructed effective Q chain, assert actual product/output Scale and Levels, and stop if either backend does not produce the exact expected state; the algebra does not substitute for that runtime check.

At Level 4 and Level 3 the Q-prefix rule gives the same four authoritative rows `q0..q3`; `q4` is the logical Rescale divisor, not an active stored row. The 4→3 transition retains q0..q3, so it neither contracts the prefix nor authorizes reading dormant q4. The existing Fast multiplication/Rescale bound checks and the strict centered-capacity checker remain required for each concrete operation/output. The explicit fixed-σ helper currently rejects Level 3; its Level 4 condition must remain intact until the later gated CP-C change. The only proposed boundary is a manually invoked injection after the complete, source-visible stage-1 or stage-2 `MulRelinNew → Rescale` pair, never a generic Rescale hook.

**Theory retained from 001:** for each stage use `r_i = z_i − Decode(x_i)⊙Decode(y_i)` with the actual decoded ciphertext inputs at that stage. This subtracts prior input error propagation from the incremental operation residual. Separately, compare each decoded output with its plaintext oracle; stage 2 total error includes stage-1 error/injection propagated through multiplication by the second plaintext, plus stage-2 residual. With equal output Scale `Δ`, the first-order coefficient-noise model predicts `Var(slot)=N·Var(η)/Δ²` for both injections (`N=8192`); this is a moment model, not Gaussian/tail equivalence. The baseline residual may combine key-switch, Rescale rounding, encoding, and decode effects; it must not be named wholly as relinearization noise.

**Existing helper cost inventory for CP-D:** seeded `NormFloat64` sampling, integer rounding, coefficient hash and online moments; coefficient-to-residue construction and four per-row NTTs; c0 copy plus four-row NTT addition; candidate metadata/slice construction; then input and candidate prefix-to-coefficient conversions, signed Q-prefix CRT reconstruction/centering for every coefficient, capacity comparison, and residue-agreement validation. The helper allocates the sampled coefficient slice, NTT noise polynomial, copied c0 candidate, candidate ciphertext slice, and coefficient-domain checker output (plus setup/arithmetic temporaries). CP-D should attribute only with existing profiling/timing utilities and must preserve the strict checker.

**Gate decision:** no new Scale, Level, row-authority, or manual-boundary ambiguity is identified in the authorized explicit two-operation schedule. CP-B is authorized to verify the exact constructed Level/Scale states, native per-stage oracles, Fast Q-prefix gates, and provenance. Any actual scale mismatch, unsupported level, capacity failure, or Fast/Standard schedule mismatch is a stop condition, not a reason to change scales or weaken a gate. Bootstrap spend remains `0/0`.

### CP-B implementation gate — ready for bounded baseline capture

The maintained `cmd/perfprobe` now has an explicit `--mul-rescale-chain` public-native path that records 9 decoded checkpoints (including both MulRelin products and both Rescale outputs), original A/B/C plaintext vectors, exact q4/q5 Scale/divisor schedule, per-step `makePublicOperationResidual` statistics, total plaintext-oracle error, effective Q/P and input/workload fingerprints, build/source/binary provenance, and an existing `compare-public` chain branch that recomputes residuals and oracle metrics from the bounded vector artifact. It has no Bootstrap invocation path and rejects nonzero Bootstrap budgets or Bootstrap/trace flags. Standard/Fast arithmetic remains the same shared application call sequence; only the existing build-tag backend adapter differs.

Implementation verification before any CP-B capture: `go test -tags perf_fast ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics` passed; the isolated genuine Standard modfile focused `perf_standard` tests for the chain limits, residual evidence, and chain fingerprint passed; `go vet -tags perf_fast ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics` passed. An initial compile-only check required a task-local writable Go build cache because the default user cache was not writable; no source/cache outside this task's temp path was changed.

CP-B capture budget remains at most three Standard/Fast matched pairs, with zero Bootstrap. No pair has run under this journal yet. Preserve raw per-slot vectors outside Git where possible; commit only compact paired summaries, hashes, provenance, and any failure evidence.
