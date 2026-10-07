# Current Task

Task: FAST-STANDARD-COMPARISON-EVIDENCE-002
Status: READY_FOR_CODEX
Task class: I — separate formal paired comparisons from synthetic historical diagnostic evidence

**Authoritative executable spec:** `specs/FAST-STANDARD-COMPARISON-EVIDENCE-002.md`

## Accepted predecessor

`FAST-STANDARD-INPUT-PROVENANCE-001` is accepted for its bounded input construction and preflight only (Primary implementation `37b73ad58983bd7689999847ed5c59c4a1e14fed`, report `56c1793348d33f4a075746cc69b5dd95a5dabeac`).

## Purpose and boundaries

The formal paired `cmd/perfprobe compare` still requires legacy `fastdiag numerical`, which uses a special `c0=encoded-message,c1=0` input for its in-tree Standard evaluator. That diagnostic must not be mandatory or treated as formal genuinely encrypted Standard evidence.

Repair only the formal comparison evidence boundary and focused tests. Do not change Standard or Fast cryptographic production code, run a performance benchmark, or modify/delete historical result files.

The old `FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**; its legacy execution instructions are not authorized.

## Required handoff

Execute only the referenced new task, self-review, run targeted tests, commit/push Primary if safe, and report `FORMAL_COMPARISON_EVIDENCE_BOUNDARY_READY` or `FORMAL_COMPARISON_EVIDENCE_BOUNDARY_BLOCKED`, followed by `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW` as applicable.
