# Current Task

Task: FAST-STANDARD-INPUT-PROVENANCE-001
Status: READY_FOR_CODEX
Task class: I — native Standard encrypted-input / explicit Fast simulation input provenance

**Authoritative executable spec:** `specs/FAST-STANDARD-INPUT-PROVENANCE-001.md`

## Purpose

Repair and validate the input-construction boundary for a future formal Standard-vs-Fast Bootstrap comparison. Standard must use genuine pinned Standard key generation, encryption, Bootstrap-compatible input, and matching-key decryption/decoding. Fast input must be explicit about its current intentionally insecure zero-secret/direct-encoded model.

This is **input provenance preflight only**, not a performance rebaseline or numerical-fidelity acceptance campaign. Preserve production code, the pinned Standard implementation, and existing results.

## Superseded / blocked task

`FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED** and is **not authorized for execution**. Its old specification still mandates a plaintext-like `c1=0` Standard input; do not follow it or run any old benchmark command. Only the new input-provenance task is active.

Do not run seven-repetition timing campaigns or publish speedup/SNR claims. Do not independently start another task.

## Required handoff

Implement and test only the new specification, self-review, commit/push Primary if safe, and report `INPUT_PROVENANCE_READY`, `INPUT_PROVENANCE_PARTIAL`, or `INPUT_PROVENANCE_BLOCKED`, followed by `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW` as applicable.
