# FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013

**Status:** READY_FOR_CODEX
**Task type:** Web-approved autonomous bounded E/I coverage batch; one Web review at batch completion or genuine blocker.
**Previous Web acceptance:** `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-web-review.md`.
**Pinned genuine Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Pinned Fast Secondary:** `00ac70ba136d190fa31bbb26c2f51d003a221634`.
**Authority:** `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A reuse-first, §4B batches; `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`.

## Goal and limits

Finish **one milestone** of public CKKS arithmetic drop-in evidence using **identical frontend source in both dependency builds**; reuse the earlier 005 oracle fixture/profiles and 006A/009 tests when relevant, rather than recreate arithmetic kernels. Validate `Add/Sub`, `MulRelin` supported overloads, `Rescale`, and one composed operation chain under Fast P0 zero-secret **full-active-Q** representation. Preserve intentional insecurity and original public API/signatures/parameters. The explicit C0 compact-prefix implementation is **not** implicitly interchangeable with P0; no stale/dormant row reads or native unsafe fallback.

This batch is **Primary harness/tests/evidence only**. No changes to genuine Standard or Secondary Lattigo production or tests, no modifications of old frozen runners/artifacts, no full-CKKS retrofit, no compact-Q materialization architecture change. Codex may autonomously repair **only mechanical Primary harness and reporting errors** within one bounded repair pass for each checkpoint, with independent numerical oracles and transactional state tests.

## Continuous authorized checkpoints

**A. Audit and fixtures — cheap, read-only preflight.** Safe sync via AGENTS, verify both pinned Lattigo workspaces clean and expected commits, read current spec and the prior 005/006A/009/012 summaries, plus exact existing current public API signatures. Make a minimal source map for each *already implemented* primitive, distinguish generic public evaluator from optimized explicit Fast and shared kernel, and select fixed small deterministic multi-slot complex inputs with two suitable Levels and an independently computed cleartext oracle. Prefer reuse of 005 frontend and known profile (LogN13, four Q, one P, 16 slots) where compatible. Preserve exact Q/P identity and source hash matching between Standard/Fast.

**B. Public Add/Sub.** Using unchanged shared frontend source and identical CKKS parameters in both dependency builds, exercise `Add`, `AddNew`, `Sub` and `SubNew` as API supports; include ciphertext/ciphertext and a supported plaintext or scalar overload if existing API permits. Check decoded complex oracle, input immutability, output Level/Scale/LogDimensions, c1 status, full active Q rows and ring-domain consistency. **Do not assert a specialized Fast kernel was used unless dispatch instrumentation/source evidence proves it**; ordinary generic full-Q calculation on valid P0 ciphertext may count only as numeric public API compatibility.

**C. Public MulRelin.** Exercise ciphertext×ciphertext `MulRelin`/`MulRelinNew` on valid Fast zero-c1 full-Q inputs and genuine Standard ciphertexts with required native relin keys; check branch selection or keyset lookup tracking (Fast no native relin-key lookup; Standard ordinary relin behavior), independent cleartext product/square oracle, c1 invariant, degree/Level/Scale, input/output aliasing where supported, and negative invalid c1 with **actual** normal-layout relin key. Investigate plaintext/scalar operand variants separately based on actual public signatures: preserve error on unsupported operand rather than inventing feature support. Do not rewrite the 006A full-Q branch or claim it shares the explicit prefix MulRelin core when it does not.

**D. Public Rescale + composition.** Test isolated public Rescale at a source-proven suitable high Level, verifying top-q removal (Level decrement and Scale divided by actual logical qLevel), independent plaintext decoded comparison and no dormant-row consumption. Test RescaleTo only if supported by current pinned public API; otherwise record `unsupported` not failed. Execute one modest arithmetic sequence, e.g. Add → MulRelin → Rescale → Rotate, with predetermined scale-compatible inputs and explicit oracle at each step, in identical Standard/Fast frontend. Use existing appropriate valid evaluation keys; preflight Q/P layout against each evaluator to avoid repeating 010 error. Preserve full-active-Q P0 and existing Fast zero-secret fail-closed behavior. **Do not** force the sequence through compact C0 or Bootstrap.

**E. Milestone audit and handoff.** Produce a compact matrix: operation/overload, Standard and Fast source-identical test coverage, plaintext oracle RMSE/max, actual Fast path (generic full-Q vs special keyless/shared fastcore), c1/Level/Scale/backing checks, unsupported cases, source and parameter/input hashes, and unresolved P0/C0 gaps. Record attempts, clean repo SHAs and at least one reproducible source/provenance artifact. Highlight whether old 005 coverage can be reused, what is new, and priority for next scientific design; no performance claim. Final Codex coding self-review, package tests and go vet under both workspaces; `git diff --check`.

## Execution cadence and costs

- Codex completes A→B→C→D→E **without intermediate Web review** if each gate passes. Per checkpoint: execute + oracle + focused tests + coding self-review + up to one local repair; journal state, commit safe checkpoints and continue. On session/context ending, leave a resumable journal and `BATCH_IN_PROGRESS`, then next `開始` resumes. No new math or unsupported architecture chosen autonomously.
- **Expensive-budget cap: zero Bootstrap calls, zero benchmarks, zero LogN16, zero performance sweeps.** Cheap deterministic tests can run as needed for mechanical fixes, but do not tune parameters/oracle thresholds for green results.
- **STOP** `BATCH_BLOCKED_NEEDS_WEB_REVIEW` for: differing immutable parameters/source between backends, unanticipated API behavior requiring Secondary edits, unexpected numerical failure after bounded repair, unsafe native KeySwitch fallback, missing or stale active rows, unapproved P0/C0 conversion, or mathematical conflict. Record the first failure, do not work around it by silently changing the workload or writing duplicated kernels.
- If a distinct overload is definitively unsupported but the charter explicitly allows documenting it as such, label it unsupported and continue; otherwise new architectural ambiguity is a STOP. This keeps minor unimplemented variants from causing meaningless emergencies without hiding real defects.

## Deliverables and classification

- New Primary `tools/fast-dropin-public-primitives-batch-013/` source/test fixture(s) with one identical frontend, compact `results/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013-journal.md`, final summary/coverage matrix and compact JSON evidence (do not commit 4096-slot arrays or secret keys).
- Preserve 012 results and the user's accepted 006A/009 production code. Only Primary normal fast-forward commits/pushes under AGENTS; Secondary and genuine Standard remain clean/pinned.
- Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` and `READY_FOR_WEB_REVIEW` only at E success, `BATCH_IN_PROGRESS` for a resumable session boundary, or `BATCH_BLOCKED_NEEDS_WEB_REVIEW` / `NEEDS_WEB_REVIEW` for a true blocker.
- Formal performance baseline `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED until separately authorized matched performance and representation-proven correctness.
