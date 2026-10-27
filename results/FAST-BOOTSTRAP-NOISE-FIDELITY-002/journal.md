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

The versioned ledger began with 18 planned slots and zero reservations/calls. Each slot changes status only as an attempt is actually reserved and consumed. Per-call `perfmeasure` reservation/result sidecars are authoritative for invocation order and timing.

## CP-B implementation and focused validation

Date: 2026-10-27

The minimal integration is implemented locally in the maintained `cmd/perfprobe` and existing triplet comparator. No new runner, statistics package, Noise sampler, or Secondary production arithmetic was added.

- Added the explicit `--noise-fidelity-002` mode for the exact LogN13/E32 zero-call preflight and existing six-call `--public-repeatability` path. The 002 Fast source pin is exactly `49b65dd…`; its preflight pair binds both executable SHA-256 values and the task id. The established Standard and Fast pins for all other modes remain unchanged.
- Kept the historical one-call `--public-bootstrap-once` Noise option and pin behavior. Self-review caught and repaired an initial validator change that would have rejected that legacy mode; regression coverage now checks both the historical acceptance path and the task-specific six-call mode.
- Attached the existing Fast Noise Hook to the same evaluator's six existing calls. Every call still uses a fresh copy of the held ciphertext; there is no extra warmup or Bootstrap call. The fixed task seed is enforced at the CLI, and the comparator verifies every recorded operation/side seed against the accepted deterministic derivation.
- For 002 only, the final `1e-6` output threshold is recorded as an observational pass/fail field. Output structure, finiteness, 4096 slots, output hashes, recomputed metrics, preflight, source/binary identity, Scale, representation, and strict Q-prefix evidence remain validation requirements. Legacy repeatability retains its strict gate.
- Extended the existing triplet comparator for exact 1 Cold + 5 Warm artifacts per arm. It reports all individual metrics/timings/allocations, warm mean/median/sample standard deviation/min/max, construction and construction-plus-cold costs, pairwise output errors, precision-bit gaps to Standard, and net Fast+Noise overhead. Global `-log2(RMSE)` is labeled separately from p99/p99.9 error tails. Noise vectors remain absent; event coefficient summaries, seeds, operation order, hashes, and capacity proofs are retained.
- Tightened report validation to reject duplicate/missing evaluator-construction phases, duplicate hook configuration, or hook configuration on an OFF arm.
- A provenance review identified a commit-order cycle: the task requires the fresh preflight pair committed before measured calls, but the older runner pins `primary_commit` to the current HEAD, so committing the pair would invalidate it. For 002 only, the runner now takes `primary_commit` from the clean binary's Go `vcs.revision`, records the current clean HEAD as `primary_evidence_commit`, requires the source revision to be an ancestor, and recomputes the maintained source fingerprint at both commits. This permits later result-only commits without rebuilding and does not change older modes. The triplet comparator now requires the actual pair file, revalidates its exact SHA/task pins, and checks the Standard/Fast executable SHA values against it.

Focused tests passed without invoking a complete Bootstrap:

- Primary Fast-tagged `cmd/perfprobe` synthetic/task validation:
  `go test -modfile=/private/tmp/fast-bootstrap-noise-fidelity-002.mod -tags perf_fast ./cmd/perfprobe -run 'TestNoiseFidelity002|TestPrimaryTaskIdentityAllowsEvidenceOnlyCommitAfterCleanCodeBuild|TestPublicBootstrapNoiseSnapshotRequiresTwoIndependentCapacityCheckedSides|TestPublicTripletSummaryUsesGlobalRMSEPrecisionAndTails' -count=1` — PASS. This includes a zero-call pair fixture with raw pair hash, both binary SHA bindings, and rejection of a mismatched pair reference.
- Primary Standard-tagged CLI/comparator validation:
  `go test -modfile=/private/tmp/fast-bootstrap-noise-fidelity-002.mod -tags perf_standard ./cmd/perfprobe -run 'TestNoiseFidelity002|TestPrimaryTaskIdentityAllowsEvidenceOnlyCommitAfterCleanCodeBuild' -count=1` — PASS. These are synthetic validator tests using the local module replacement, not a formal Standard capture.
- Secondary Fast Chebyshev power 2/power 4/power 16 Montgomery Noise tests — PASS; four ct×ct products in power 16, eight injections, strict capacity, input immutability, and NTT/Montgomery output were checked. Power-16 regression patch is retained at `evidence/secondary-fast-power16-regression.patch` (SHA-256 `bd98219b3688743bb70f69951b4c2a1e5bcd62d0ec94145dfdda572e2474a7de`, 2,330 bytes). The test-only patch was removed from the Secondary checkout after the run so the frozen binary source remains clean at `49b65dd…`.
- Secondary LogN13/E32 `TestFastEvalModP93ExplicitQPrefixConsumesQ3` Mod1 path — PASS.
- The task source-identity test verified that clean code commit `ca9fc012…` and its result-only descendant `e0e6ade…` have identical maintained-source fingerprints, and that `vcs.modified=true` is rejected.
- An additional pre-existing `TestFastMod1FormalDegree30PolynomialRegression` was attempted and failed in its purported Standard half when run inside the Fast checkout. That fixture is not genuine Standard: it converts Fast plaintext-like `c1=0` state, then invokes Standard-named evaluator APIs in the Fast module, which dispatch to Fast zero-secret arithmetic and fail on missing active `q4` row. A temporary fixture correction also failed the expected nonzero-`c1` assertion because the Fast capability intentionally remains `c1=0`; all such fixture edits were reverted. This failure does not exercise the accepted Fast P93/E32 Mod1 path; the targeted test above passed. No production arithmetic was changed.

At that earlier checkpoint, Primary code and focused results were still uncommitted, and no binaries or fresh 002 preflight existed. The next section records completion of those gates; actual Bootstrap reservations/calls remain zero.

## Clean builds and fresh zero-call preflight

Date: 2026-10-27

The final Primary measurement source was committed at `9f3f53ff9639866605ab0dd13da0431cb4c2775f`. Its maintained-source fingerprint is `894668dd7efef96e2cf9d17680881be43c14bedf96f2279e5ac552a88f369f54`. Clean source checkouts were used for the binaries: genuine Standard Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and Fast `experiment/fast-bootstrap-noise-fidelity@49b65dd927780cec0d7e05772a3df8554ef58005`.

- Standard binary: `/private/tmp/perfprobe-standard-002`, SHA-256 `df76ea83d25c3f428bb4567c6d34bb32af1434189b5918d6246484e3549bdbc8`, `perf_standard`, darwin/arm64, 13,539,874 bytes.
- Shared Fast/FN binary: `/private/tmp/perfprobe-fast-002`, SHA-256 `311fdca496368f42102192c39c0835454a0c41c3bd1bc4064b32ad4b117745c7`, `perf_fast`, darwin/arm64, 14,222,274 bytes. Fast OFF and Fast+Noise will use this exact same binary.
- Both binaries report Go `vcs.revision=9f3f53ff9639866605ab0dd13da0431cb4c2775f`, `vcs.modified=false`, the expected build tag and module source in `go version -m`. Temporary modfiles and their checksums are recorded in `preflight/manifest.json`; the repository `go.mod` was not changed. The builds exited successfully. Go emitted an optional module-stat-cache write warning (`operation not permitted`) for its default cache; this did not affect binary creation, VCS metadata, SHA verification, or execution.
- Environment: Apple M4, darwin/arm64, Go `go1.26.4`, 10 CPUs / `GOMAXPROCS=10`; default GOGC/GODEBUG/GOMEMLIMIT.

The fresh Standard and Fast preflights both returned `PASS_ZERO_BOOTSTRAP_PREBOOTSTRAP_CHAIN`; the existing `compare-public` flow produced `PASS`. Actual Bootstrap calls and reservations remain `0/18`. The pair records 4096 slots, LogN13/E32, effective Bootstrap Scale `2^45`, exact Q/P fingerprint, original-input/workload/oracle fingerprints, matching environment, eight checkpoints, and the pinned Standard/Fast source and binary identities. Pair SHA-256: `37ab35c0eb724c381f7b6a78e5ebf17cc391708a596907a8f22e6e59a3cb0ae0` (15,833 bytes).

Versioned compact documents:

- `preflight/standard.json`: 29,423 bytes, SHA-256 `cfef27f713cc7c3dc2b2d1e5b8dd3bad44852d353dbbe85e1344d33bd37a8248`.
- `preflight/fast.json`: 29,765 bytes, SHA-256 `76e97e37e1135aad13ae3f7bd5eee8f09cb14ab3508ac9ed310fba0147490b71`.
- `preflight/pair.json`: 15,833 bytes, SHA-256 above.
- `preflight/manifest.json` captures the source/build/module/environment identities and all artifact hashes. Both decoded 4096-slot preflight vectors are intentionally outside Git: Standard `/private/tmp/fast-bootstrap-noise-fidelity-002-preflight/standard-vectors.json`, 4,165,467 bytes, SHA-256 `ce90902d3b66cf56442db7cc6fccd9d330ffb2c6f9f08a30977080d3482f7f47`; Fast `/private/tmp/fast-bootstrap-noise-fidelity-002-preflight/fast-vectors.json`, 4,159,849 bytes, SHA-256 `66641219e5c9cec380c3a985da01c051cfbf931173a28b1a7bcfda2615427e12`.

All manifest file sizes and hashes were rechecked against the files, including both binaries and the external vectors. No output-precision threshold was used to waive any preflight structural, capacity, provenance, finite-slot, or oracle checks. No Bootstrap was reserved or invoked. The plan, manifest, ledger, and preflight artifacts must be committed and pushed before the first FN reservation.
