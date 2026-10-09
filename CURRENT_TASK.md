# Current Task

Task: FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012
Status: READY_FOR_CODEX
Mode: EXPLICITLY AUTHORIZED MULTI-CHECKPOINT AUTONOMOUS BATCH
Review: ONE Web milestone review on completion, or immediate NEEDS_WEB_REVIEW on a hard blocker

**Executable batch charter:** `specs/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012.md`
**Prior approved key-plan math:** `specs/FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011.md`
**010 Web failure diagnosis:** `results/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010-web-review.md`
**Durable autonomy/workflow authority:** `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4B; `AGENTS.md`.

## Purpose

The user authorized Codex to implement, test, self-review, repair and **continue across related subtasks without a Web handoff on each small change**. Web review remains compulsory **at the batch milestone** or when a genuine architecture/math/provenance blocker requires independent arbitration.

This batch covers a strict ordered chain:
- **A**: fix Standard `RotateNew` evaluator/key QP mismatch in a **new identical-source shared frontend**, using the Bootstrap Parameters evaluator with existing Bootstrap evaluation keys in BOTH pinned Standard/Fast builds.
- **B**: validate both ordinary public Rotate preflights, genuine Standard native encryption and Fast zero-secret input, independent plaintext oracle and full provenance.
- **C**: only if both B preflights pass, execute at most **ONE genuine Standard + ONE Fast E32 Bootstrap** on their rotated ciphertexts; zero repeats.
- **D**: if C passes, produce a concise read-only end-to-end compatibility/gap inventory and milestone handoff.

**After each passed checkpoint, Codex automatically proceeds to the next permitted checkpoint** after coding self-review, affected tests and safe journal commits. No need to return for Web between A, B, C and D. Journal under `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-journal.md`; if session ends, return `BATCH_IN_PROGRESS` and resume from journal on next `開始` without repeating consumed Bootstrap work. Stop `BATCH_BLOCKED_NEEDS_WEB_REVIEW` when any math, architecture, parameter, key-layout or numerical decision is unapproved or a required gate fails; report evidence, do not improvise.

Pinned genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Pinned Fast Secondary `00ac70ba136d190fa31bbb26c2f51d003a221634`. Both libraries unmodified. Existing 004/010 source and reports immutable. Original zero-secret simulation is **not secure HE**. Formal `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.

At completed batch, return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` and `READY_FOR_WEB_REVIEW`. Before batch completion, `BATCH_IN_PROGRESS` means continue autonomously or await the next Codex session, **not a request for Web acceptance**.
