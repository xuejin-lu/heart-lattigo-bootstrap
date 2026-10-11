# CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001

Status: `READY_FOR_CODEX` — user-authorized autonomous integration task. This task supersedes the active pointer to `FAST-LOGN16-TRACE-OFFLINE-RECOVERY-007`; preserve that historical spec and all Task005/006 evidence.

## Objective

Consolidate Primary CKKS measurement and diagnostics onto the existing shared platform; make common profile metadata and algorithms depend on effective CKKS parameters rather than a LogN13/LogN16 allowlist; repair/revalidate the already captured Task006 recursive LogN16 trace using the Batch025-accepted analyzer; and retire obsolete duplicate root/`tools/` runner software without losing independent mathematical or security regression coverage.

The user has authorized implementation, tests, and deletion of confirmed superseded software within this scope. Stop only for a CKKS math/security semantic decision, a persistent unexplained correctness failure, unavailable/mismatched immutable evidence that cannot be safely recaptured under this spec, or unsafe Git conditions.

## Maintained platform and architecture

The only maintained Primary measurement owners are:

- `cmd/perfprobe`: paired public Standard/Fast measurement and comparisons;
- `cmd/fastdiag` plus `scripts/fastdiag`: diagnostics, trace ingestion/analysis and reporting;
- `internal/perfmeasure`: effective parameters, shared workload and provenance;
- `internal/numericalmetrics`: decoded numerical metrics.

Extend these existing owners. Do not create a separate runner, event-tree validator, statistical aggregator, metric implementation, compatibility CLI, or duplicate profile calculator. Secondary may be inspected read-only; do not modify or push Secondary.

The same CKKS application operation source and ordinary public APIs must be built independently against genuine Standard and Fast dependencies. No runtime backend selector or backend-specific arithmetic branch may enter the frontend. No production CKKS arithmetic or security/key semantics may change in this task.

## Effective parameter contract

1. Derive recorded LogN, LogSlots, slots, Scale, exact Q/P primes, E, and profile fingerprints from the effective constructed CKKS and bootstrapping parameters, not only JSON hints or a finite set of hard-coded LogN names.
2. Apply the same shared parameter/profile logic to LogN13, LogN14, LogN15, and LogN16 test profiles. Preserve all frozen historical fixture identities and hashes as profile-specific checks; do not rewrite their meaning.
3. If configuration construction, a capacity/numerical gate, or a selected backend does not support an effective profile, return an explicit unsupported/not-proven status with the exact failing field/gate. Do not fabricate support, switch backend, change parameters, or weaken a threshold.
4. Retain deterministic-input fingerprints and versioned schemas so historical results remain auditable.

## Task006 trace recovery

The Task006 compact report records a 539-event raw Fast LogN16/E32 trace and an unverified source status. Expected immutable evidence, subject to rechecking Task006 metadata:

- raw path: `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2617861684/raw-trace-unverified.json`;
- raw SHA-256: `9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d`;
- failure sidecar SHA-256: `f598b481b776448dcc78ff80083610e9376cee12451716ddb88fd414abfc756b`.

Use the existing Batch025 `cmd/fastdiag` recursive event validation, analysis, and independent-root closure machinery. The known trace chain `generated_powers#59 → T16#60 → T8#61 → T4#62 → T2#63` is valid recursive ancestry; do not reinstate a direct-child-only rule or create another validator. Preserve the original Task006 raw/report/call journals byte-for-byte and keep the original `TRACE_UNVERIFIED` label; assign any new offline assessment only to a separate Task result. Validate provenance, stage tree, real/imag power roots, recurrence ancestry, Rescale children, row/Level/Scale fields, event timing/closures, and recorded numerical output metadata without claiming a new decrypt.

Prefer offline recovery. The Task006 raw currently exists at the expected path/hash, so the default allowance is zero new Bootstrap calls. If the original raw is later confirmed absent, first validate the corrected existing trace path using zero-call tests. Only then this task authorizes one new trace-recovery capture: at most one cold and one traced-warm Fast Bootstrap, reserved/journaled before launch, numerical max-complex gate `1e-6`, no retries. If a present raw file has a hash mismatch, preserve it and stop for evidence review; do not overwrite it or treat a mismatch as permission to recapture. This exception does not authorize additional Task005 performance measurements or a Standard Bootstrap campaign. Preserve explicit STOP conditions and never reuse Task005/006 reservation tokens.

## Historical tool retirement

1. Inventory root-level runners/CLI/adapters and `tools/` programs, with their current references, build/test dependencies, and capabilities already provided by the maintained platform.
2. Retire code that has been superseded or is only a historical command implementation. Do not add shims to keep retired CLIs working. Git history is sufficient for old executable code; preserve specs, configs, result artifacts, provenance, hashes, and published conclusions.
3. Before removing old tests, identify independent mathematical, input-integrity, backend-provenance, capacity, row-authority, and security/key-semantics regression value. Move required tests into existing maintained packages and shared test helpers. Do not port whole historical experiment runners merely to retain their old test harness.
4. Do not remove current Q0123 arithmetic, Rescale, NTT, Chebyshev, Bootstrap boundary, numerical-metric, or safety regression coverage. Do not delete historical result/spec artifacts as cleanup.

Regression migration boundary: Primary config strictness and effective-prime sensitivity belong in `internal/perfmeasure`; public rotation orientation belongs in `cmd/perfprobe`. Historical CRT/Rescale/Q-prefix/capacity/row-authority/zero-secret security tests stay with the maintained Secondary arithmetic and Bootstrap packages (including the Q0123/Batch027 and keyless zero-secret regression suites); do not copy those backend tests into experiment CLIs. Tests that only assert historical runner classifications, JSON schemas, or obsolete Q01/P93 workarounds are retired with those runners.

## Checkpoints and implementation order

This is one approved autonomous integration task. Complete sequentially without waiting for per-checkpoint review:

1. **Parameter generalization:** identify hard-coded profile/LogN gates in shared code; centralize effective-parameter derivation and explicit unsupported reasons; test the common path across LogN13–16 while preserving frozen LogN13/16 fingerprints.
2. **Task006 offline recovery:** reuse the Batch025 analyzer and make the Task006 evidence compatible through the existing diagnostic platform; produce a compact summary, journal, and evidence artifact without copying raw events or vectors into Git. If the trace is unrecoverable, follow the bounded capture rule above or report the precise blocker.
3. **Tool audit and retirement:** complete the dependency inventory, migrate unique regression tests, and delete confirmed duplicate/superseded runners, adapters, and obsolete root/`tools/` CLI code. Preserve unrelated configs and artifacts.
4. **Self-review and final validation:** check scope and provenance, run focused low-cost tests and vet for maintained packages, `git diff --check`, verify no new runner/analyzer was introduced, and report all unsupported parameter profiles explicitly.

## Verification

- Run focused tests for `internal/perfmeasure`, `internal/numericalmetrics`, `cmd/perfprobe`, and `cmd/fastdiag`, including parameterized LogN13–16 fixtures, unsupported-profile behavior, recursive T16→T8 ancestry, cycles/orphans/duplicates, independent real/imag roots, and non-overlapping parent/child timing attribution.
- Audit test bodies before invoking suites. Do not accidentally execute real Bootstrap or benchmark tests as part of a low-cost regression run. If an existing test performs real crypto work, select safe zero-call tests or replace only its harness seam with a non-crypto fixture; do not weaken its correctness oracle.
- Run `go vet` on maintained packages and `git diff --check`. A repository-wide `go test ./...` is not required unless the complete remaining test inventory is confirmed not to trigger Bootstrap, benchmark campaigns, or other expensive experiments.
- Do not rerun Task005 performance measurements. Do not change CKKS mathematics, production backend source, thresholds, parameters, or frontend operations.

## Results and Git

Write compact Primary results:

- `results/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-summary.md`
- `results/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-summary.json`
- `results/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-journal.md`
- `results/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-task006-offline-summary.md`
- `results/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-task006-offline-journal.md`
- `results/CKKS-MEASUREMENT-PLATFORM-GENERALIZATION-001-task006-offline-evidence.json` (compact projection of the existing offline analyzer output)

Record effective-profile support by LogN, Task006 raw/sidecar provenance and offline classification, exact retired source paths and migrated regression tests, commands/results, and Bootstrap counts. Do not include raw event arrays. Preserve the Task006 and Task005 files unchanged.

Commit the durable task spec/rules/pointer before implementation, then commit implementation/results. Push Primary `main` with ordinary fast-forward only under `AGENTS.md`; never rewrite or force-push history. End `READY_FOR_WEB_REVIEW`, or `NEEDS_WEB_REVIEW` with the precise unresolved mathematical/security/evidence blocker.

## Commands

Use focused package commands after auditing test bodies, for example:

```sh
go test ./internal/perfmeasure ./internal/numericalmetrics -count=1
go test ./cmd/fastdiag -run '^(TestValidatePublicE32EventsRecursivePowerTree|TestPublicE32Offline.*|TestAnalyzePublicE32Events.*)$' -count=1
go vet ./internal/perfmeasure ./internal/numericalmetrics ./cmd/perfprobe ./cmd/fastdiag
git diff --check
```

The test selector is illustrative; update it to the final migrated test names and explicitly omit any real-Bootstrap/performance test unless separately authorized by the bounded trace exception.
