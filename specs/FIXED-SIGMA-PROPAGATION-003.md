# FIXED-SIGMA-PROPAGATION-003 — Frozen-σ propagation ablation, guarded cost improvement and Bootstrap source map

**Status:** `AUTHORIZED_AUTONOMOUS_BATCH` but **not yet run**. Codex starts only on the user's next `開始`.
**Canonical primary:** `xuejin-lu/heart-lattigo-bootstrap` `main`; **secondary:** `xuejin-lu/lattigo` `fast-qprefix`.
**Input reviewed pins:** Primary `25de6e43386cc16baf9f81bf455356b2da49ad1b`, Secondary `1e7828293c36f9845e30723ba4e9500d84bb50d2`, genuine original Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Actual runtime clean SHA, binary module identity and experiment sources must be recorded. Prior: `results/FIXED-SIGMA-CHAIN-002/web-review.md`.
**Hard budget:** `0` genuine Standard Bootstrap, `0` Fast Bootstrap. **NO LogN16, large benchmark, teammate BTS, multiparty, historical A=0, or operation-level Bootstrap invocation** (including tests that invoke Bootstrap).
**Frozen sigma:** exactly `4.004280012097432` **coefficient-domain signed rounded Gaussian**, fresh deterministic seed per completed logical operation, same σ at both stages; **no refit** to obtain favorable outcomes.

## Scientific question

The prior two-stage chain proved two genuine sequential `ct×ct MulRelinNew → Rescale` operations and two explicit injections with near-Standard *incremental* RMS under one repeated A/B/C workload. It **did not quantitatively isolate** propagation of the stage-1 artificial error, because the stage-2 residual `r_2 = z_2 − Decode(x_2)⊙Decode(y_2)` subtracts earlier error from the predictor.

This batch isolates the transmitted noise using an **ablation with otherwise identical original inputs, RNG state and public operation sequence**:
- `OFF/OFF`: both injections disabled (Fast baseline).
- `ON/OFF`: inject only immediately after stage 1 (trace its impact on stage-2 pre-injection output).
- `OFF/ON`: inject only after stage 2 (new stage-2 error only).
- `ON/ON`: exactly one fresh injection after each of two operations (current model).

The Standard reference has **no injected artificial noise**. For matched runs and stage2 decoded values, compute:
`P1 = Decode(Stage2_ON/OFF) − Decode(Stage2_OFF/OFF)`.
The first-order expectation is approximately `P1 ≈ Decode(Stage1_Injected − Stage1_Base) ⊙ C_decoded` **only after accounting for equivalent operand encoding, operation rounding and stage2 input**, with nonlinearity/correlation residual shown separately. Compare `ON/ON − OFF/ON` to P1 for a second consistency check; the two contexts differ in rounding state and must not be assumed identical. Use a separate oracle for the final total error and record operation incremental RMS, variance, signed complex mean, Max Error and seeds at each stage.

## Checkpoints, ordered and gated

### CP-A — Safe startup + reuse/proof preflight (no edits to CKKS arithmetic)

Read Primary `AGENTS.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md`, `docs/MEASUREMENT_PLATFORM.md`, prior 001/002 specs, reviews, journal. Execute proper git status/fetch/safe ff-only in existing Primary and Secondary worktrees, honor Secondary `AGENTS.md` and `docs/FAST_QPREFIX_SPEC.md`. Never clone redundant repository, `git reset`, stash/clean, force push, amend/rebase or overwrite unknown dirty state. Check artifact sizes *before* reading; no full fetch of large vector dumps.

Map the exact existing `cmd/perfprobe/fixed_sigma_chain.go` stage-1/2 source and `InjectFixedSigmaCoefficientNoiseDiagnostic` so each ablation configuration has an unambiguous post-Rescale boundary. Confirm the proposed scaled/perturbed second operand remains a genuine ct×ct operation at the accepted q4 Scale / Level4 and strict capacity. Write short invariants to new journal. Stop `BATCH_BLOCKED_NEEDS_WEB_REVIEW` if a novel math/Scale/Q-prefix decision is needed.

### CP-B — Small zero-Bootstrap ablation with truly different inputs

Use existing Primary perfprobe and `internal/numericalmetrics`, not a new runner. Extend the **research-only** harness options or task-specific fixture to select each of four named ablation arms; the application-facing Standard/Fast CKKS call order must remain identical and backend-neutral. The injection diagnostic is Fast-only build-adapter opt-in and disabled outside this fixture.

Use at most **two accepted input workloads**: (i) original canonical A/B/C, (ii) one predeclared, moderately stronger C input (e.g. unit-magnitude or scaled diagnostic values) chosen *before examining outputs*, with same effective params/Scale, to expose propagation. Record exact input hashes and strict centered capacity; do not tune σ or backfit message scale. If the stronger C violates numerical/representation gates, record failure and do not invent replacements. For each workload take at most two fixed seed pairs across four Fast arms (no more than **16 total Fast chain executions**); the maximum genuine Standard reference is **two chains**, one per distinct workload, independently compiled against genuine Standard. Existing accepted 002 results may be reused instead of unnecessary duplicative canonical baseline.

Use matching plaintext oracle and compare actual decoded vectors, stage1 injected delta, stage2 propagated delta P1, stage2 op-incremental residual, final total error, tail and covariance. Preserve binary SHA, vcs.revision/modified, Q/P + input SHA, code SHA and exact run count. Do not treat slots as independent repeated trials. Record a result table even on falsification. No automatic claim of Gaussianity.

### CP-C — Guarded cost reduction of capacity checker (conditional)

ONLY if CP-A/B have passed without new scientific decisions, explore reusing buffers/inverse-NTT/centered CRT intermediates inside the existing **Fast-only diagnostic** library path. Strict fail-closed centered prefix bounds, exact signed coefficient-to-residue agreement, all active rows, c1/dormant row preservation, meta/Scale invariants and transactional commit must remain **exactly equivalent**.

First specify equivalence and adversarial oracle tests: valid, boundary-near-half, invalid overflow, deliberately malformed q_i residue, wrong NTT/coefficient representation, Level3/4, sigma=0. If any optimization would alter Q-prefix mathematics, remove checks, change safe bound or generalized CKKS semantics, **stop and return to Web review without implementing**. If only low-risk allocation/scratch reuse is possible, implement smallest change, compare oracle parity on bounded tests, and use same existing profiling to measure before/after. If too risky, leave code unchanged and report suggested optimizations. No tuning σ, no parallel non-deterministic benchmark sweep.

### CP-D — Read-only Bootstrap coverage map + durable delivery

Without running Bootstrap or editing its algorithm, inspect **bounded source excerpts** for the existing true Bootstrap evaluator/poly code and map every known category of ct×ct, deferred Relin/Rescale, ct×pt, standalone Rescale and public evaluator dispatch. Note where the desired “inject after every complete three-stage multiplication” could and **could not** be guaranteed. If no single automatic semantics-preserving public hook exists, state that plainly; suggest a future explicit wrapper/operation event contract, not a magic one-line global hook. **Do not activate automatic Bootstrap injection** in this batch.

Update `results/FIXED-SIGMA-PROPAGATION-003/journal.md`, `summary.md`, bounded hash-listed evidence. Preserve 001/002 completed artifacts unchanged. Normal safe task-scoped commit/push Primary and Secondary (only if gated CP-C changed library). Final status `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` or `BATCH_BLOCKED_NEEDS_WEB_REVIEW`; stop on conflict or unapproved mathematical semantic change.

## Gates / no shortcuts

- Full Bootstrap calls: **0/0**, and no tests covertly calling it.
- Distinguish `r_op` (incremental) from total oracle error. The first-stage injected error's transmitted effect must be tested by ablation; never infer its magnitude from the Stage2 incremental RMS alone.
- Keep `σ=4.004280012097432` as-is across all arms, Levels and inputs; record `FIXED_SIGMA_NOT_TRANSFERABLE` or `INSUFFICIENT_EVIDENCE` if observed. Do not selectively hide invalid tails or data.
- Run source/provenance checks on actual binaries, not only strings from run manifests; retain genuine Standard original Lattigo vs Fast zero-secret and no security claims.
- Maintain same frontend source and only research harness backend adapters, reusing existing metrics, no new runner/duplicate framework.
- Do not silently extend helper to nonzero c1, unsupported Level/Scale/PK, arbitrary Rescale, entire Bootstrap, or user-facing production API. Time attribution is a diagnostic, not formal speed benchmark.
