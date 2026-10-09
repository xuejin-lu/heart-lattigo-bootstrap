# FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012

**Status:** READY_FOR_CODEX
**Class:** Web-approved bounded autonomous E/I batch under `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4B.
**Objective:** finish the existing 011 corrected-key-plan integration, validate one additional public chained operation and produce a source-backed gap inventory before ONE Web milestone review.
**Pinned Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Pinned Fast Secondary:** `00ac70ba136d190fa31bbb26c2f51d003a221634`.
**Canonical bootstrap config SHA:** `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`.
**Input SHA:** `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.
**Architecture:** P0 public zero-secret full-active-Q, C0 explicit compact-Q, NO implicit conversion. One identical frontend source for pinned genuine Standard and Fast dependency. Zero-secret is insecure simulation.

## Workflow

Once a user says `開始`, safely sync both repositories, read refreshed `AGENTS.md` and `CURRENT_TASK.md`, and complete the following checkpoints **sequentially without an intermediate Web handoff** if successful. Each checkpoint includes implementation/execution, a targeted test, coding self-review, at most one local repair pass, commit/push when safe, and a compact journal in `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-journal.md`. Continue automatically to the next authorized checkpoint. If the Codex context/session must end, leave a resumable journal and `BATCH_IN_PROGRESS`; next `開始` resumes rather than repeating expensive experiments.

### A — Correct 010 key-plan mismatch (supersedes 011 handoff cadence)

Execute the existing `specs/FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011.md` under this batch's autonomy/cost limits. The original 010 frontend used full Bootstrap Galois keys generated over Q/P to rotate a Level0 ciphertext with a residual evaluator that has P Level -1. Create a **new** identical-source frontend using `ckks.NewEvaluator(btpParams.BootstrappingParameters, keys.MemEvaluationKeySet)` in both pinned Standard and Fast; leave historical 004/010 source and evidence unchanged. Before expensive operations prove residual Q is a prime-exact prefix of bootstrapping Q, same ring degree/type and secret extension. Record failed preflight as a hard batch stop if compatibility is still wrong. No changes to either Lattigo backend.

### B — Demonstrate correct no-Bootstrap public composition and parity

From the fixed source and exact frozen canonical input/profile run the unchanged Standard and Fast frontend preflights: native keygen/EncryptNew → ordinary public RotateNew → public DecryptNew/Decode with correct nonconstant multi-slot cleartext rotation oracle. Verify key provenance and Q/P compatibility, Level0/Scale, metadata, full active q0 backing, Standard native c1, Fast c1=0/no GaloisKey lookups. Capture exact pre-execution shared frontend SHA **even on panic**, config/input/Q/P hashes, RMSE/max/SNR and (if both runs succeed) direct Standard-vs-Fast decoded delta. **Both pass** or stop; do not change the user-facing profile/parameter values.

### C — Single bounded public Bootstrap composition

If A and B pass, execute at most **one genuine Standard Bootstrap and one Fast zero-secret Bootstrap** on the pre-rotated Level0 ciphertext, with the same identical frontend source and original E32 configuration. Verify output Level/Scale/slots/c1, independent rotated-cleartext oracle, numerical RMSE/max/SNR, direct backend delta and actual Bootstrap invocation count. No retry, warmup, benchmarking or parameter tuning. **Total expensive-call budget for this entire batch: 1 Standard + 1 Fast Bootstrap maximum**, regardless of subprocess/session restarts. Journal the used budget and never repeat a consumed call. If a bootstrap fails, preserve exact error/provenance and stop for Web; do not unilaterally change algorithms.

### D — Read-only end-to-end coverage and next-wave recommendations

Only if C succeeds, create a compact audit **without editing Secondary production** identifying what is and is not proven by same-source public CKKS execution: EncryptNew, Add, MulRelin, Rescale, Rotate, Bootstrap, DecryptNew; note any full-Q/C0 mixing and keyplan boundaries; distinguish numerical oracle success from execution of optimized Fast kernels and from proven runtime speedup. Include targeted *cheap* Go tests/compilation to verify the final submitted runner but **no extra Bootstrap**. Recommend the next 2–4 related I/E tasks to advance the original drop-in goal and eventual compact-Q performance. Do not automatically implement new unreviewed math/representation architecture.

## Autonomy and STOP boundaries

- Autonomous task-to-task progression is authorized for A→B→C→D **only** after each prerequisite passes. Codex handles routine harness defects, compile/vet fixes, reflection/logging schema problems and test adjustments via one bounded self-review repair. **Do not change math, ciphertext semantics, security assumptions, Q-prefix width/authority, Q/P configuration, numeric oracle tolerances, pinned refs or dependency versions.**
- Scope limited to **Primary frontend test harness + compact evidence, journal and aggregate results**. Secondary, genuine Standard and explicit Fast kernels remain unchanged. No backend API/algorithm modification or public production application changes.
- If a compatibility issue remains after local repair, a previously unknown library/source inconsistency requires crypto/architectural judgment, a numerical oracle fails, there is a backend divergence or tests require broader changes, stop immediately `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with specific evidence and decision request. No silent Standard native downgrade, arbitrary fallback or manual c1 mutation.
- If all A-D pass, `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`. Do not treat Codex self-review as independent scientific acceptance. The user asks Web for **one** milestone review then.
- No independent Web review between A,B,C,D, no filler subtask/commit, and no repeated historical experiment merely to satisfy cadence. Preserve all old tasks/results.

## Final artifacts and safe pushes

`results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-journal.md` with checkpoint statuses, minimal source hashes, per-backend attempt counts and consumed Bootstrap budget; `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-summary.md` and compact machine-readable evidence (no secret key dumps); and one identical-source 011/012 frontend under `tools/`. Existing `011` spec remains the authoritative key-plan math and execution constraint, modified here only by explicitly approved continuous handoff.

Use normal fast-forward commits/pushes under AGENTS. After completing A→D, report `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` / `READY_FOR_WEB_REVIEW`. If interrupted and resumable, `BATCH_IN_PROGRESS`, and report next authorized checkpoint. If blocked, `BATCH_BLOCKED_NEEDS_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.
