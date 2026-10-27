# FAST-BOOTSTRAP-NOISE-FIDELITY-002 journal

Date: 2026-10-27
Status: `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`
Actual Bootstrap calls: `18/18`
Reservations: `18/18`

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


## FN arm capture — complete and validated

Date: 2026-10-27. The first measured arm used the pre-pinned Fast binary with Noise enabled. Exactly six Bootstrap calls were reserved and returned in the same evaluator lifetime: one Cold followed by five Warm. No warmup or retry was performed. All six calls used independent copies of the same held Level0 ciphertext (ciphertext fingerprint `e512061ab0cc1ca3ebab443ab8d03ae021286124ade946d97fc51bc216a0c95f`), and all output checks completed with 4096 finite slots, Level 1, degree 1, NTT non-Montgomery output, and Scale 2^45. The exploratory `1e-6` gate passed for these six samples; each per-sample gate remains explicitly recorded in the raw document.

- Capture document `raw/fn.json`: 824394 bytes, SHA-256 `a38bf6548b208aabc0721763b56fcb69bbaf10738300159f3a4d6a3e94391612`.
- Decoded vectors remain outside Git at `/private/tmp/fast-bootstrap-noise-fidelity-002-captures/fn-vectors.json`: 6519727 bytes, SHA-256 `758c35cd5bc0fae72d98358d9c6919b72d7f11e4ad8f8506ccbcc6dc950d72ce`. The document contains the 4096-slot bootstrap plaintext input hash, six decoded sample hashes, and complete metrics inputs.
- Each of the six exclusive reservation files and six result sidecars is preserved under `raw/`; `attempt-ledger.json` binds each slot to its reservation/result sidecar hashes and decoded output hash. All reservations are `irrevocably_reserved_before_call`; all result sidecars are `RETURNED`.
- `bootstrap_noise_evidence`: root seed `20261027`, sigma `4.004280012097432`, 156 ct×ct multiply hooks and 312 polynomial injections across six calls. Operations are ordered 1 through 156. The event evidence validates distinct deterministic left/right seeds, all operations completed, and all 312 strict Q-prefix capacity checks passed.

| # | Phase | Elapsed ns | Allocated bytes | Allocations | RMSE | Global precision bits (`-log2(RMSE)`) | MaxAbs | Decoded output SHA-256 |
|---:|---|---:|---:|---:|---:|---:|---:|---|
| 1 | first_cold_bootstrap | 648866292 | 1033675696 | 19286917 | 6.20791478276e-09 | 27.263244 | 5.82553954106e-08 | `4b4783ba1b16383271f1fc55a4aaa84d4ca3865851b6e738d34c79e75ee24095` |
| 2 | warm_bootstrap_01 | 131644667 | 79606120 | 12509 | 6.17040057222e-09 | 27.271989 | 5.80076097502e-08 | `d1ea874ced6a6cac559f3fc8784b5f663e81794890e0c46c8375ff6f75ecd2ad` |
| 3 | warm_bootstrap_02 | 131107125 | 79638888 | 12509 | 6.20086156417e-09 | 27.264884 | 5.22368164053e-08 | `930ace6f98b746878a6f1f5d47a0a4a51eee4dabc7f265649c8028706a8411ba` |
| 4 | warm_bootstrap_03 | 132045208 | 79573304 | 12507 | 6.14152732555e-09 | 27.278755 | 5.90075198002e-08 | `dc97e7cc524eb0980672ca612e0172997fc0bd5ce5445a72e79fabf964aad45a` |
| 5 | warm_bootstrap_04 | 144226541 | 79573304 | 12507 | 6.21945942331e-09 | 27.260564 | 6.74194736557e-08 | `c73f4d3c1134a1254ac98f5cbcddc28dedcc9e4998f6e7600273e7440647b810` |
| 6 | warm_bootstrap_05 | 133289917 | 79707656 | 12519 | 6.23092904695e-09 | 27.257906 | 5.97318770758e-08 | `c1afe15266282f73a397a0533ac275a117eb7269971133df4adc8924c10acad5` |

Cold timing/allocation: 648866292 ns, 1033675696 bytes, 19286917 allocations. Warm timing (ns): `131644667, 131107125, 132045208, 144226541, 133289917`; mean 134462691.600, median 132045208.000, sample SD 5517097.792, min 131107125, max 144226541. Warm mean allocation: 79619854.4 bytes and 12510.2 allocations.

Evaluator construction phases: CKKS evaluator 1097375 ns / 6195696 bytes / 16682 allocations; bootstrapping evaluator 2654542 ns / 9640480 bytes / 32101 allocations. Combined: 3751917 ns / 15836176 bytes / 48783 allocations; construction plus Cold Bootstrap: 652618209 ns. Hook configuration phase is separately recorded in `raw/fn.json`.

Noise hook evidence is in the raw document, not expanded into separate files: root seed `20261027`, σ `4.004280012097432`, 156 ordered ct×ct events (operations 1–156), 312 sampled polynomial injections; every multiply completed, every left/right Q-prefix check reported `strict_fit=true`, and every event records the actual independent side seeds, Level/Scale, coefficient moments/max/SHA-256, and transformation/capacity evidence. No complete coefficient vector is serialized.

The user authorized FN→S→F and requires each arm checkpoint to be committed/pushed before reserving the next arm. At this checkpoint the task ledger is 6/18 calls; Standard/Fast remain 0/6 each. Proceed only after this FN evidence is committed and pushed.


## Standard arm capture — complete and validated

Date: 2026-10-27. The genuine Standard binary, built against Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, completed exactly one Cold plus five Warm calls. All six calls used independent copies of the same held ciphertext fingerprint `bad8ac3f290de3408d9d8f9f52b29ddaacb15722ef48a7bd55d3304361883dc2` and the identical bootstrap plaintext input SHA-256 `c32c0d7b731c0fcd8d82fb9c1182a43a62d120f761c090ce21dc764e9aa1e703`. The native Standard lifecycle is confirmed by `input_c1_nonzero=true`; no Fast backend was used. All outputs have 4096 finite slots, Level 1, degree 1, NTT non-Montgomery representation, Scale 2^45, and passed the exploratory 1e-6 observation gate. The six decoded output hashes are identical in this capture; all six raw samples remain represented and were not deduplicated.

- Capture document `raw/standard.json`: 45192 bytes, SHA-256 `8bff544b68f0c42306427c369a7150e3077ad0847a94574e1665da25230684e7`.
- Decoded vectors remain outside Git at `/private/tmp/fast-bootstrap-noise-fidelity-002-captures/standard-vectors.json`: 6525586 bytes, SHA-256 `db8ded36ed46fd0aabb66a81fe8de459709bb8af06a413de059ebb1096b94d81`.
- Six reservation and six result sidecars are preserved under `raw/`, with paths and SHA-256 values bound by `attempt-ledger.json`; all were reserved and returned exactly once. Noise was not configured for Standard.

| # | Phase | Elapsed ns | Allocated bytes | Allocations | RMSE | Global precision bits (`-log2(RMSE)`) | MaxAbs | Decoded output SHA-256 |
|---:|---|---:|---:|---:|---:|---:|---:|---|
| 1 | first_cold_bootstrap | 325961917 | 143968736 | 36696 | 4.91010460178e-09 | 27.601599 | 1.46721660358e-08 | `ef6b65fffdd08061b2f8f93689df37dbf9380e71e0a49642e71002860d62fa18` |
| 2 | warm_bootstrap_01 | 309807250 | 91944920 | 35027 | 4.91010460178e-09 | 27.601599 | 1.46721660358e-08 | `ef6b65fffdd08061b2f8f93689df37dbf9380e71e0a49642e71002860d62fa18` |
| 3 | warm_bootstrap_02 | 307756208 | 91674088 | 35005 | 4.91010460178e-09 | 27.601599 | 1.46721660358e-08 | `ef6b65fffdd08061b2f8f93689df37dbf9380e71e0a49642e71002860d62fa18` |
| 4 | warm_bootstrap_03 | 326458917 | 91874760 | 35043 | 4.91010460178e-09 | 27.601599 | 1.46721660358e-08 | `ef6b65fffdd08061b2f8f93689df37dbf9380e71e0a49642e71002860d62fa18` |
| 5 | warm_bootstrap_04 | 321371375 | 91730064 | 35004 | 4.91010460178e-09 | 27.601599 | 1.46721660358e-08 | `ef6b65fffdd08061b2f8f93689df37dbf9380e71e0a49642e71002860d62fa18` |
| 6 | warm_bootstrap_05 | 320322208 | 91664504 | 35002 | 4.91010460178e-09 | 27.601599 | 1.46721660358e-08 | `ef6b65fffdd08061b2f8f93689df37dbf9380e71e0a49642e71002860d62fa18` |

Cold timing/allocation: 325961917 ns, 143968736 bytes, 36696 allocations. Warm timing (ns): `309807250, 307756208, 326458917, 321371375, 320322208`; mean 317143191.600, median 320322208.000, sample SD 8010897.580, min 307756208, max 326458917. Warm mean allocation: 91777667.2 bytes and 35016.2 allocations.

Evaluator construction phases: CKKS evaluator 984375 ns / 2719992 bytes / 16610 allocations; bootstrapping evaluator 353882000 ns / 710670992 bytes / 14232896 allocations. Combined: 354866375 ns / 713390984 bytes / 14249506 allocations; construction plus Cold Bootstrap: 680828292 ns.

The six Standard outputs have global RMSE `4.91010460178e-09`, corresponding to 27.601599 global precision bits; this is distinct from the output's median per-slot precision field. Full output hashes and vectors are retained for later comparator recomputation.

The task ledger is now 12/18 calls. FN and Standard are each 6/6 successful; Fast OFF remains 0/6 and must not start until this checkpoint is committed and pushed.


## Fast OFF arm capture — complete and validated

Date: 2026-10-27. This arm used the same Fast binary as FN (`311fdca496368f42102192c39c0835454a0c41c3bd1bc4064b32ad4b117745c7`), with the Noise Hook explicitly OFF. It completed exactly one Cold plus five Warm calls. The output document validates the same original plaintext SHA-256 `c32c0d7b731c0fcd8d82fb9c1182a43a62d120f761c090ce21dc764e9aa1e703`, canonical input, Q/P and workload fingerprints as the preflight and other arms; Fast's expected zero-secret `c1=0` mode is recorded. All six outputs are finite, 4096-slot, Level 1/degree 1, NTT non-Montgomery, Scale 2^45, and structurally valid. No Noise evidence is present, as expected.

- Capture document `raw/fast.json`: 45583 bytes, SHA-256 `b5c25af9e1e6f54ae2696be29630e2441c65f80c96325edc4d8e77c897a1d2fb`.
- Decoded vectors remain outside Git at `/private/tmp/fast-bootstrap-noise-fidelity-002-captures/fast-vectors.json`: 6520359 bytes, SHA-256 `e266aeba87a80f73f0c49e7b11e30e990c837acc4d3a25d089ba68efad99082e`.
- Six reservation and six result sidecars are under `raw/`; the ledger records paths and SHA-256 values. All six results returned and passed finite/data validation. The six decoded output hashes are identical in this capture, so the samples remain as separate timings rather than being deduplicated.

| # | Phase | Elapsed ns | Allocated bytes | Allocations | RMSE | Global precision bits (`-log2(RMSE)`) | MaxAbs | Decoded output SHA-256 |
|---:|---|---:|---:|---:|---:|---:|---:|---|
| 1 | first_cold_bootstrap | 561848875 | 961512376 | 19282518 | 1.17957184064e-09 | 29.659090 | 5.7387742391e-08 | `31cb3bb8b9f7377918ddd7cae659e9851230bd903acdba69bd32ca4a273bc934` |
| 2 | warm_bootstrap_01 | 70035916 | 7506168 | 8107 | 1.17957184064e-09 | 29.659090 | 5.7387742391e-08 | `31cb3bb8b9f7377918ddd7cae659e9851230bd903acdba69bd32ca4a273bc934` |
| 3 | warm_bootstrap_02 | 71019167 | 7506624 | 8115 | 1.17957184064e-09 | 29.659090 | 5.7387742391e-08 | `31cb3bb8b9f7377918ddd7cae659e9851230bd903acdba69bd32ca4a273bc934` |
| 4 | warm_bootstrap_03 | 70619708 | 7506168 | 8107 | 1.17957184064e-09 | 29.659090 | 5.7387742391e-08 | `31cb3bb8b9f7377918ddd7cae659e9851230bd903acdba69bd32ca4a273bc934` |
| 5 | warm_bootstrap_04 | 70844125 | 7506168 | 8107 | 1.17957184064e-09 | 29.659090 | 5.7387742391e-08 | `31cb3bb8b9f7377918ddd7cae659e9851230bd903acdba69bd32ca4a273bc934` |
| 6 | warm_bootstrap_05 | 73143792 | 7506168 | 8107 | 1.17957184064e-09 | 29.659090 | 5.7387742391e-08 | `31cb3bb8b9f7377918ddd7cae659e9851230bd903acdba69bd32ca4a273bc934` |

Cold timing/allocation: 561848875 ns, 961512376 bytes, 19282518 allocations. Warm timing (ns): `70035916, 71019167, 70619708, 70844125, 73143792`; mean 71132541.600, median 70844125.000, sample SD 1183929.064, min 70035916, max 73143792. Warm mean allocation: 7506259.2 bytes and 8108.6 allocations.

Evaluator construction phases: CKKS evaluator 1313500 ns / 6195696 bytes / 16682 allocations; bootstrapping evaluator 2651666 ns / 9640608 bytes / 32102 allocations. Combined: 3965166 ns / 15836304 bytes / 48784 allocations; construction plus Cold Bootstrap: 565814041 ns.

This completes the authorized 18/18 calls: FN 6/6, genuine Standard 6/6, and Fast OFF 6/6; no retries, warmups, Bootstrap extras or failures occurred. The ledger records all slots as succeeded and the comparator is the only analysis step remaining.

## Comparator boundary repair — no experiment calls

The first offline run of the existing `compare-public-triplet --noise-fidelity-002` failed before it wrote an output file. It reported: `Noise event 3 phase commit_c0 is absent or collected per-operation MemStats`. Inspection of the unchanged FN capture showed all five named phases in the required order, with `memory_stats_collected=false`. Of 312 recorded left/right `commit_c0` phase measurements, 103 are `0 ns`; all 1,248 other subphase measurements are positive, and all 312 `commit_c0` entries are present. Raw events, seeds, coefficient summaries/hashes, and strict capacity evidence are intact. No Bootstrap was executed by this comparator attempt, and the capture files were not modified.

This is a comparator validation defect: timer granularity can record a present `commit_c0` operation as 0 ns. The narrowly scoped repair accepts 0 ns only for a present, correctly ordered `commit_c0` phase; it continues to reject zero-duration sampling/transform/add/capacity phases and any per-operation MemStats collection. Regression tests exercise those cases in both `perf_standard` and `perf_fast` builds and pass. No Secondary helper or arithmetic changed, and no Bootstrap was repeated. The analysis binary will be rebuilt from the repaired Primary comparator source; its source revision and binary SHA will be recorded separately from the already-frozen measurement binaries.


## Final comparison — BATCH_COMPLETE_READY_FOR_WEB_REVIEW

The accepted 18-call budget is fully consumed: FN 6/6, genuine Standard 6/6, Fast OFF 6/6; all reservations returned successfully exactly once. No hidden warmup, retry, or extra Bootstrap occurred. The final existing comparator validated all 18 finite 4096-slot outputs and all task/source/binary/preflight/input identities. It recorded the `1e-6` MaxAbs observation for every output; all 18 happened to pass. The measured run used Apple M4 / darwin-arm64, Go `go1.26.4`, 10 CPUs and `GOMAXPROCS=10`, with matched host/runtime configuration.

Final comparator artifact: `triplet.json`, 858,526 bytes, SHA-256 `40dad6e49fe5d8c06adbec6c3bdd33af8ea2d25674668f9db43b762d0ba5435d`. It contains all 18 per-sample RMSE, global `-log2(RMSE)` precision bits, MaxAbs, p99/p99.9, numerical-gate flag, elapsed time, allocation bytes/count, and decoded output SHA-256; it also contains paired phase comparisons and pairwise output RMSE. The three decoded-vector files remain external at the exact paths and hashes recorded in `attempt-ledger.json`; the comparator read and recomputed all 18 outputs.

The table separates the cold observation from the five warm observations. Warm time and allocation lists retain each call; timing summaries are mean / median / sample standard deviation (`n-1`). Precision bits are global `-log2(RMSE)`, not median per-slot precision. Tail columns are MaxAbs / p99 / p99.9.

| Arm | Cold time | Cold alloc bytes / count | Five Warm times (ms) | Warm mean / median / SD (ms) | Five Warm alloc bytes (MB); allocation counts | Cold RMSE / global bits | Warm median RMSE / global bits | Cold MaxAbs / p99 / p99.9 | Warm median MaxAbs / p99 / p99.9 |
|---|---:|---:|---|---:|---|---:|---:|---:|---:|
| S | 325.962 ms | 143,968,736 / 36,696 | `309.807, 307.756, 326.459, 321.371, 320.322` ms | 317.143 / 320.322 / 8.011 ms | `91.945, 91.674, 91.875, 91.730, 91.665` MB; `35027, 35005, 35043, 35004, 35002` | 4.9101e-09 / 27.6016 | 4.9101e-09 / 27.6016 | 1.46722e-08 / 1.05619e-08 / 1.32257e-08 | 1.46722e-08 / 1.05619e-08 / 1.32257e-08 |
| F | 561.849 ms | 961,512,376 / 19,282,518 | `70.036, 71.019, 70.620, 70.844, 73.144` ms | 71.133 / 70.844 / 1.184 ms | `7.506, 7.507, 7.506, 7.506, 7.506` MB; `8107, 8115, 8107, 8107, 8107` | 1.17957e-09 / 29.6591 | 1.17957e-09 / 29.6591 | 5.73877e-08 / 1.54085e-09 / 6.67157e-09 | 5.73877e-08 / 1.54085e-09 / 6.67157e-09 |
| FN | 648.866 ms | 1,033,675,696 / 19,286,917 | `131.645, 131.107, 132.045, 144.227, 133.290` ms | 134.463 / 132.045 / 5.517 ms | `79.606, 79.639, 79.573, 79.573, 79.708` MB; `12509, 12509, 12507, 12507, 12519` | 6.20791e-09 / 27.2632 | 6.20086e-09 / 27.2649 | 5.82554e-08 / 1.29862e-08 / 1.53686e-08 | 5.90075e-08 / 1.32674e-08 / 1.6704e-08 |

Warm timing summaries (mean / median / sample SD / min–max) are Standard `317.143 / 320.322 / 8.011 / 307.756–326.459 ms`, Fast `71.133 / 70.844 / 1.184 / 70.036–73.144 ms`, and Fast+Noise `134.463 / 132.045 / 5.517 / 131.107–144.227 ms`.

| Construction cost | Standard | Fast OFF | Fast+Noise |
|---|---:|---:|---:|
| CKKS + Bootstrapping evaluator construction (ns / bytes / allocations) | 354,866,375 / 713,390,984 / 14,249,506 | 3,965,166 / 15,836,304 / 48,784 | 3,751,917 / 15,836,176 / 48,783 |
| Construction + Cold Bootstrap (ns / bytes / allocations) | 680,828,292 / 857,359,720 / 14,286,202 | 565,814,041 / 977,348,680 / 19,331,302 | 652,618,209 / 1,049,511,872 / 19,335,700 |
| Noise hook configuration (ns / bytes / allocations) | n/a | n/a | 820,875 / 4,224,968 / 16,643 |

Fast warm median was 4.522x faster than Standard; Fast+Noise warm median was 2.426x faster than Standard. The Cold Bootstrap-only ratios were `0.580x` for Fast and `0.502x` for Fast+Noise; both Fast cold calls were slower than Standard's cold call on this host, while construction costs are reported separately above.

Against Standard, Fast OFF precision bits were `29.6591` Cold and `29.6591` at the Warm median, a positive gap of `2.0575` bits. Fast+Noise was `27.2632` Cold and `27.2649` at the Warm median, below Standard by about `0.3384` Cold and `0.3296` bits at the Warm median. Noise reduced the absolute precision-bit gap in all six paired phases (from `2.0575` to `0.3384` bits Cold, and from `2.0575` to `0.3296` bits on the first Warm sample). Thus Noise moved Fast closer to Standard here, while ending slightly below Standard's precision; it did not match Standard exactly.

Compared with Fast OFF, Fast+Noise added `87.017 ms` to Cold (`1.155x`, or 15.5% higher) and `61.201 ms` to Warm median (`1.864x`, or 86.4% higher). Median Warm allocation increased by 72,099,952 bytes and 4,402 allocations; Cold increased by 72,163,320 bytes and 4,399 allocations. Across six calls the hook recorded 156 ct×ct multiply events and 312 polynomial injections at σ `4.004280012097432`, root seed `20261027`; all q-prefix capacity checks passed. Aggregate injection subphase time was 361,862,854 ns. Per-injection MemStats collection was disabled; allocation overhead is the paired full-call measurement.

Comparator repair/build provenance: source commit `6168bf69b846a4e0a810bb5f834c575865dbdc06`, analysis binary `/private/tmp/perfprobe-standard-002-comparator-6168bf6`, SHA-256 `d037678409bb5a6cdbb0e46131685183ab6a82e8cd0e447ec167918be9d13a88`, `vcs.modified=false`, `perf_standard`, darwin/arm64, Go `go1.26.4`. The focused regression passed under both `perf_standard` and `perf_fast`; it uses synthetic comparator fixtures and invokes no Bootstrap. The build succeeded with an optional Go module stat-cache write warning (`operation not permitted`); VCS metadata and binary hash were verified. Measurement binaries and their frozen source revision remain unchanged.

The completed result is exploratory and host/profile-specific. It supports the observed conclusion that fixed-sigma Noise moved Fast's global precision bits closer to genuine Standard in all six paired phases, with a clear increase in measured runtime and allocations. It does not establish cryptographic-noise equivalence or a causal model beyond this controlled three-arm run.
