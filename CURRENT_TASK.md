# Current Task

Task: FAST-STANDARD-OUTPUT-PREFLIGHT-001
Status: READY_FOR_CODEX
Task class: I + E — bounded genuine-Standard/Fast public Bootstrap output preflight

**Authoritative executable spec:** `specs/FAST-STANDARD-OUTPUT-PREFLIGHT-001.md`

## Accepted predecessors

- `FAST-STANDARD-INPUT-PROVENANCE-001`: native Standard Encrypt/Decrypt and explicit Fast simulation input, independently smoke-validated.
- `FAST-STANDARD-COMPARISON-EVIDENCE-002`: formal paired report no longer depends on synthetic Standard `fastdiag` and reports unassessed numerical quality.

## Purpose and limits

For LogN13 and LogN16, run **one** native encrypted Standard public Bootstrap and **one** explicitly insecure Fast simulated-input public Bootstrap per profile, and report each actual decoded output against the original plaintext. Confirm provenance, valid finite outputs and metadata. Do not classify output-quality pass/fail from the old descriptive `1e-2` cutoff.

No seven-repetition timing campaign, speedup claim, Secondary edits, production arithmetic changes, parameter tuning, or historical result deletion.

`FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**; do not execute the old specification.

## Handoff

Implement only the referenced output-preflight task, run focused tests and four bounded smoke executions, commit/push Primary if safe, and return the defined outcome with `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
