# Current Task

Task: FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013
Status: READY_FOR_CODEX
Mode: APPROVED AUTONOMOUS FIVE-CHECKPOINT BATCH
Web review: ONCE after milestone E, or immediately on genuine scientific blocker

**Runnable batch spec:** `specs/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013.md`
**Accepted 012 Web review:** `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-web-review.md`
**Workflow/reuse guard:** `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/§4B; `AGENTS.md`
**Long-term end-goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Accepted predecessor and boundary

`012` is **ACCEPTED as a bounded, correct same-source public Rotate→Bootstrap numerical composition**: frozen LogN13 E32 input, genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` vs insecure Fast zero-secret `00ac70ba136d190fa31bbb26c2f51d003a221634`; at most one actual Bootstrap call per build was made, with matching frontend/config/input hashes. Bootstrap output oracle RMSE Standard `4.885854924910455e-9`, Fast `1.1814120448423705e-9`; direct backend RMSE `4.954605783029611e-9`. This does not establish full drop-in, security, compact-Q public interoperability or speedup. Do not rerun 012 Bootstrap.

## Current uninterrupted A→E charter

A: reuse-first public Add/MulRelin/Rescale source/API inventory and frozen same-frontend test profile.
B: same-source Add/Sub variants and plaintext-oracle checks.
C: same-source MulRelin/MulRelinNew with keyless Fast path and native Standard key evidence; unsupported scalar/plaintext overloads classified explicitly; negative fail-closed controls.
D: same-source Rescale and a short Add→MulRelin→Rescale→Rotate public composition with correct Level/Scale/full-Q and numerical oracles. No bootstrap.
E: aggregate coverage/oracle/dispatch/representation matrix, journal, self-review, two-backend tests/vet and ONE Web milestone handoff.

Codex may automatically advance to the next checkpoint **after** the previous acceptance gate passes and its own bounded implementation review/repair; do not seek Web review between A-E. If Codex session ends, journal `BATCH_IN_PROGRESS`; subsequent `開始` safely syncs and resumes. Stop `BATCH_BLOCKED_NEEDS_WEB_REVIEW` only for math/architecture uncertainty, unsupported changes, persistent oracle failure, unsafe key fallback or provenance inconsistency. **No Secondary or Standard modifications, no duplicated Fast arithmetic kernels, no implicit P0↔C0 conversion.**

Fast is intentionally insecure zero-secret numeric simulation. Frozen P0 ordinary public CKKS full-Q is distinct from explicit C0 compact Q-prefix. **Zero Bootstrap, zero benchmark** and no LogN16 for this batch. Formal performance baseline `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.
