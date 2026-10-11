# CKKS Measurement Platform Generalization 001

Status: `READY_FOR_WEB_REVIEW`

## Outcome

The maintained Primary platform now derives shared profile labels and metadata from constructed CKKS/bootstrapping parameters, uses centralized ordered Q/P fingerprints, and exercises one parameter-construction path at LogN13–16. No production CKKS arithmetic or application operation sequence changed.

| LogN | Shared effective-parameter construction | Frozen public-native contract |
|---:|---|---|
| 13 | PASS | Available |
| 14 | PASS | Explicitly `UNSUPPORTED_EFFECTIVE_PROFILE`: no frozen config/Q/P/input/workload contract |
| 15 | PASS | Explicitly `UNSUPPORTED_EFFECTIVE_PROFILE`: no frozen config/Q/P/input/workload contract |
| 16 | PASS | Available under its existing observed-capacity gate; this task claims no new numerical result |

The maintained owners remain `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, and `internal/numericalmetrics`. The frontend remains backend-transparent; this task did not select a backend in application code.

## Task006 trace

The existing offline `fastdiag` path revalidated the preserved 539-event LogN16/E32 trace using the Batch025 recursive event validator, analyzer, and root closures:

- New Bootstrap calls: **0**. Original Task006 accounting remains Task005 16/16 and Task006 2/2; no retry or budget reuse.
- New assessment: `OFFLINE_RECURSIVE_TRACE_REVALIDATED`; original artifact remains `TRACE_UNVERIFIED` and was not relabeled.
- 38 independent event roots; generated-power branch roots are real #59 and imag #279. The recursive ancestry is T16 #60 → T8 #61 → T4 #62 → T2 #63.
- Bootstrap root inclusive/direct-child/signed residual time: 1,060,517,791 / 1,060,469,666 / 48,125 ns.
- Both recorded outputs passed the original 1e-6 max-complex gate: Level 1, Scale 2^45, Degree 1, 2 Q-prefix rows; oracle RMSE 5.82126707846636e-9 and max complex difference 4.6694655714754984e-7.
- Raw SHA-256: `9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d`; sidecar SHA-256: `f598b481b776448dcc78ff80083610e9376cee12451716ddb88fd414abfc756b`.

The compact Task006 analyzer evidence and journal are in [offline summary](CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-task006-offline-summary.md), [offline journal](CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-task006-offline-journal.md), and [offline evidence JSON](CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-task006-offline-evidence.json). The JSON is 17,639 bytes and omits per-event analyzer arrays. Limitations: the historical diagnostic commit object is absent from the current Secondary clone, though the recorded commit/patch digest is verified; Task005 reference vectors are not embedded.

## Retired duplicate software and retained regressions

Removed **162 root-level Go files** and **26 `tools/` Go files** (old root CLI/config builder, FIX-001/P93/Q01 runners/tests, duplicated Fast measurement/oracle/statistics code, and standalone Batch CLIs/adapters). No current maintained package or script referenced these files. Configs, specs, results, attempt journals, hashes, and historical conclusions remain intact. The exact 188 deleted source paths are recorded in [summary JSON](CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-summary.json) and the implementation commit.

Independent Primary regressions now live with their maintained owners:
- `internal/perfmeasure`: legacy-field rejection and configured prime-scale sensitivity.
- `cmd/perfprobe`: public CKKS left-rotation convention.

CRT/Q0123, Rescale, row-authority, Bootstrap-boundary, and zero-secret/keyless safety regressions remain in the unchanged Secondary Fast/bootstrapping test suites. Historical runner classification tests and obsolete Q01/P93 workaround tests were retired with their executable code.

## Validation

Passed:
- `go test ./internal/perfmeasure -count=1`
- `go test ./internal/numericalmetrics -count=1`
- Focused `cmd/perfprobe` effective-profile, unsupported-profile, and migrated regression tests.
- Focused `cmd/fastdiag` recursive ancestry, independent roots, invalid metadata, analyzer closure, and Task006 offline replay tests.
- `go vet ./internal/perfmeasure ./internal/numericalmetrics ./cmd/perfprobe ./cmd/fastdiag`
- `git diff --check`

A repository-wide `go test ./...` was intentionally not run: the task spec requires audited focused tests and excludes tests that may execute real cryptographic workloads. No Bootstrap or performance campaign was started.
