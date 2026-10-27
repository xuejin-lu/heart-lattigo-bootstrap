# FAST-BOOTSTRAP-NOISE-FIDELITY-002 journal

Date: 2026-10-27
Status: `BATCH_IN_PROGRESS`
Actual Bootstrap calls: `0/18`
Reservations: `0/18`

## Authorization and preservation

The user authorized a new, independent 18-call study on 2026-10-27. It compares genuine Standard, Fast with the hook OFF, and the same Fast binary with the hook ON, each using one Cold plus five Warm calls. The fixed settings are LogN13/E32, Bootstrap Scale 2^45, sigma `4.004280012097432`, root seed `20261027`, and the accepted canonical input/workload. Run order is FN, S, F. No retry is permitted; any call or strict structural/provenance/capacity failure stops the rest of the batch.

The previous 001 task remains immutable: its task ledger records S=1, F=1, FN=1 failed before injection, total 3/3; all 001 captures and the accepted Montgomery repair commit remain unchanged. New attempts belong only to this directory and ledger.

## CP-A reuse inventory

- Primary preflight started on `main`, clean at `ca9fc012e68d23cf1573e4f0ff26e2ddc05fa0af`; `git fetch origin` succeeded via formal permission escalation after the normal `.git/FETCH_HEAD: Operation not permitted` error. Fetched `origin/main` equals local HEAD; no divergence or local-only commits.
- Secondary experiment worktree is clean on `experiment/fast-bootstrap-noise-fidelity` at `49b65dd927780cec0d7e05772a3df8554ef58005`; fetched remote equals local. Root `fast-qprefix` remains clean and unchanged at `1e7828293c36f9845e30723ba4e9500d84bb50d2`.
- Read synchronized `AGENTS.md`, `CURRENT_TASK.md`, 001 spec, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md`, Secondary branch `AGENTS.md`, and `docs/FAST_QPREFIX_SPEC.md`.
- Reuse finding: `runPublicBootstrapAttempts` already supports six named calls and creates/fingerprints an independent copy of the same held ciphertext for each; `perfmeasure.BootstrapAttemptBudget` reserves per-call and rejects retries. `public-repeatability` already sets the exact six-call budget. Evaluator setup phases and each call's elapsed time/allocation bytes/allocation count are already measured.
- Gap: `--bootstrap-fixed-sigma-noise` is currently rejected with `--public-repeatability`; the existing task pin only covers the old one-call path. The existing `compare-public-triplet` validates one call per arm and reports one timing. Repeatability output validation currently treats the final `1e-6` MaxAbs oracle as a hard gate.
- Reuse finding: Secondary commit `49b65dd…` already provides the accepted Montgomery-compatible NTT/MForm Noise helper. Its per-operation snapshot records operation order, independent side seeds, sigma, coefficient moments/RMS/max/hash, Level/Scale, strict Q-prefix evidence, and completion without embedding full coefficient vectors. No Secondary code change is currently needed.
- Planned minimal gap repair is Primary-only: add an explicit 002-only preflight/repeatability contract with exact Fast pin `49b65dd…`; allow the existing hook in the existing six-call loop; make only this task's final numerical threshold observational; extend the existing triplet comparator to validate and summarize six outputs/call timings per arm. All old one-call, generic, and legacy pin/gate semantics remain unchanged.

## Frozen run plan

1. Implement the task-scoped `perfprobe` integration and comparator; add fake-call/schema/pin/observational-gate regressions.
2. Run focused no-full-Bootstrap Chebyshev power-16 and Fast Mod1/polynomial tests, then the minimum relevant Primary tests. Never run broad tests that can invoke complete Bootstrap.
3. Freeze/commit the final Primary source, build clean Standard and Fast binaries from the same Primary source commit, verify `go version -m`, exact VCS commits, `vcs.modified=false`, binary SHAs, and identical Fast/FN binary identity.
4. Run only the zero-call LogN13/E32 Standard/Fast preflight, require all strict pre-Bootstrap oracles/capacity/provenance to pass, compare it with existing `compare-public`, and freeze the fresh pair and all input/profile hashes.
5. Commit/push the exact plan, source pins, binary identities, preflight evidence, and zeroed 18-slot ledger before any Bootstrap call.
6. Measure FN one Cold + five Warm, validate and push; then S six, validate and push; then F six, validate and push. Each invocation uses the same evaluator for its six calls and an independent copy of the same held pre-Bootstrap ciphertext per call. No extra warmup or hidden Bootstrap.
7. Run the existing triplet comparator extended for six samples; save compact result, all per-call metrics and output hashes, warm timing summaries, Noise event evidence, and final report. Keep decoded vectors external with path/size/SHA if large; do not store full Noise coefficient arrays.

## Attempt log

The versioned ledger begins with 18 planned slots and zero reservations/calls. Each slot changes status only as an attempt is actually reserved and consumed. Per-call `perfmeasure` reservation/result sidecars are authoritative for invocation order and timing. No source build or measurement has started yet. Next: implement the explicitly scoped Primary integration and regression tests.
