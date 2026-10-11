# CURRENT TASK — Generalize and consolidate the CKKS measurement platform

Task: `CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001`
Status: `READY_FOR_CODEX` (user-authorized autonomous integration task)
Authoritative spec: `specs/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001.md`

Extend only the maintained Primary platform (`cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics`) so generic profile metadata derives from effective CKKS parameters and common logic covers LogN13–16, with unsupported backend/profile cases explicit. Revalidate the existing Task006 539-event trace offline by reusing Batch025's recursive event-tree analyzer. Audit and retire confirmed superseded root/`tools/` runner, adapter, and CLI code while preserving independent math/security regression tests and all historical results/specs/configs.

No CKKS production arithmetic or security semantics changes; the frontend remains backend-transparent. Task005 performance measurements are not repeated. Task006's existing raw trace is preferred; only if immutable raw evidence is unavailable, the spec permits a separately journaled, no-retry capture of at most one cold and one traced-warm Fast Bootstrap after zero-call validation.

Execute the sequential checkpoints in the authoritative spec autonomously, commit the spec/rules/pointer before implementation, then commit and safely fast-forward push the completed Primary work. Secondary remains read-only. Stop for an actual math/security decision, unresolved correctness/evidence failure, or unsafe Git state.
