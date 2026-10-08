# Current Task

Task: FAST-STANDARD-EPHEMERAL-REPLICATION-002
Status: READY_FOR_CODEX
Task class: E / I — small independent-key LogN13 native Standard replication (diagnostic only)

**Authoritative executable spec:** `specs/FAST-STANDARD-EPHEMERAL-REPLICATION-002.md`

## Accepted review and open issue

`FAST-STANDARD-EPHEMERAL-ABLATION-001` demonstrated an E=0-to-E=32 reduction in native Standard output RMSE **within two single-input pairs**, with only EphemeralSecretWeight changed and matched native ciphertext/secret within each pair (result `5791c45dee6f4a647d6a7a9fe7e26da9eb61832a`). E=32 remains diagnostic-only.

However, prior independently sampled **LogN13 E=0** RMSE was `2.195800` and the ablation E=0 RMSE was `9.3551e-6` for the same frozen profile. This large cross-run discrepancy is unresolved; do not claim reliable numerical validity or causation for that variation.

## Active bounded question

For LogN13 only, run **three independently generated genuine Standard key/ciphertext pairs**; in each pair compare the **same** native encrypted ciphertext under E=0 and E=32, one public Bootstrap per weight (six total). Report all per-pair output-versus-original RMSE/SNR, source/provenance, and ranges. Do not run Fast, LogN16, extra trials, benchmarks, any parameter sweep, or production-code changes.

The old `FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**.

## Handoff

Perform only the referenced diagnostic, verify all provenance and input gates, run focused tests and report `E0_VARIABILITY_REPRODUCED`, `E0_VARIABILITY_NOT_REPRODUCED_IN_SMALL_SAMPLE`, or a precise blocked/partial classification; commit/push Primary if safe and return `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
