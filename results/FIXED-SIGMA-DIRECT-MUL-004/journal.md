# FIXED-SIGMA-DIRECT-MUL-004 — durable experiment journal

**Status:** `BATCH_IN_PROGRESS`. CP-A and CP-B remain complete. F5 and F5-R1 remain preserved as pre-chain launch failures. Web authorized exactly one additional F5-R2 launch; its binary provenance and corrected command passed prelaunch checks, and the durable ledger now has its one-time launch reservation (`8/18` Fast launches, `5/16` actual chains). The process result is pending.
**Prior:** `FIXED-SIGMA-PROPAGATION-003` results remain unchanged; Web review in `../FIXED-SIGMA-PROPAGATION-003/web-review.md`.
**Authoritative spec:** `../../specs/FIXED-SIGMA-DIRECT-MUL-004.md`.

## Frozen decisions

- Replace `(A+B)×C×C` only in a **new versioned fixture** with `A×B×C`; no `AddNew` in the new measured workload. Original 001–003 files/fixtures remain valid historical evidence.
- Two full ct×ct MulRelin→Rescale operations, Stage1 Level5→4, Stage2 Level4→3, both output Scale `2^45`. B Stage1 `q5` Scale; independently defined C Stage2 `q4` Scale, q values from runtime constructed Q chain.
- New deterministic C formula and `direct-base`/`direct-c4` workloads predefined in spec. Both require new SHA/fingerprint, preflight conservative and runtime strict Q-prefix capacity.
- Frozen signed coefficient `σ=4.004280012097432`, not refit to direct multiplication. Four arms OFF/OFF, ON/OFF, OFF/ON, ON/ON, distinct deterministic seeds per stage.
- Genuine original Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (read-only); Secondary active `fast-qprefix@1e7828293c36f9845e30723ba4e9500d84bb50d2`. Primary initial reviewed base `574a1aee5ce5b0e329096095de3eda1b2df7bd1b`.

## Bounded counters (persist across sessions)

| Metric | Used | Maximum |
|---|---:|---:|
| Fast actual two-stage chain executions completed | 5 | 16 |
| Fast launch attempts recorded, including F5 and F5-R1 failures | 8 | 18 |
| Original authorized F5 pre-chain CLI failure | 1 | 1 |
| F5-R1 pre-chain binary provenance failure | 1 | permanently recorded |
| F5 recovery launches authorized / launch-reserved | 2 / 2 | 2 |
| Replacement F5-R2 | LAUNCH_COMMITTED; actual chain result pending | 1 |
| Genuine Standard two-stage chains completed | 2 | 2 |
| Standard Bootstrap calls | 0 | 0 |
| Fast Bootstrap calls | 0 | 0 |
| LogN16 workload runs | 0 | 0 |
| BTS | 0 | 0 |

| Checkpoint | Status |
|---|---|
| CP-A new-workload source/math/capacity gates | COMPLETE |
| CP-B direct A×B→×C baseline | COMPLETE |
| CP-C fixed σ four-arm experiment | IN_PROGRESS; F5-R2 authorized and preflighted, not launched |
| CP-D full analysis/provenance/commit | NOT_STARTED |

The paired numerical analysis and aggregate stage/tail metrics remain **NOT_RUN** because neither F5 nor F5-R1 entered the ciphertext chain; F5-R2's reserved process has not yet returned. The seven validated captures and their SHA-256 index remain preserved under `captures/`; F5 and F5-R1 are separate launch attempts and zero actual chains. Bootstrap remains 0/0.

## CP-A — completed source, scale and capacity preflight

### Safe repository state and source provenance

- Primary started on clean `main` at `574a1aee5ce5b0e329096095de3eda1b2df7bd1b`. After recording its empty status, `git fetch origin` found the authorized task pointer update; safe ff-only synchronization completed to `5f9174a806d6a89800a1935d2e03258bf6852682`, now equal to `origin/main`. No local-only commits or uncommitted files were present before synchronization. The only remote changes were the 004 pointer/spec/journal/summary and the 003 review. Previously accepted 003 evidence hashes were checked and remained unchanged.
- Secondary started clean on `fast-qprefix` at `1e7828293c36f9845e30723ba4e9500d84bb50d2`. After a normal fetch was denied at `.git/FETCH_HEAD`, the formal Git metadata permission retry succeeded. `origin/fast-qprefix` was identical to local HEAD (0 ahead/0 behind); no merge or Secondary source modification was needed. `AGENTS.md` and `docs/FAST_QPREFIX_SPEC.md` were reread after sync. The genuine Standard ref is `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; its prior temporary directory was absent, so CP-B must create a new isolated detached read-only checkout before compiling.
- Measurement reuse inventory: the maintained Primary `cmd/perfprobe/fixed_sigma_chain.go` already owns public native encryption, complete `MulRelinNew→Rescale` stages, post-stage Fast injection, decoded incremental/total residuals and artifact provenance; `fixed_sigma_ablation.go` already owns bounded paired arm loading, shared numerical summaries, and propagation vectors. Those are the extension points. The old fixture is Add-specific and its `PublicCapacityPlan` covers `(A+B)×C` only, so it cannot serve as evidence for this new circuit. New shared `internal/perfmeasure` workload and capacity helpers are the smallest required extension; no new runner or statistics implementation is planned.
- Fast dispatch source map and architecture were already reviewed in 003 at the same unchanged Secondary commit. The current task keeps the ordinary public CKKS calls: `ckks.Evaluator.MulRelinNew` (`schemes/ckks/evaluator.go`) and `Rescale` (`schemes/ckks/evaluator.go`), dispatched internally to Fast `MulCore.ApplyRows` (`schemes/ckks/internal/fastcore/mul.go`) and Q-prefix `RescaleCore.ApplyRows` (`schemes/ckks/internal/fastcore/rescale.go`). Prior 003 Web review confirmed these were the active paths; this task does not re-audit or alter the Fast algorithm.
- Durable Q-prefix rule: centered-lift authority requires strict `2B < S_Q(Level)`, with `S_Q(Level)=product(q0..q_min(Level,3))`; the four rows used here are actual logical Q residues. No private-F storage or dormant row is used in the plan.

### Frozen direct workload and scale schedule

The direct circuit is `Encrypt(A, Level5, Δ)`, `Encrypt(B, Level5, q5)`, `MulRelinNew(A,B)`, `Rescale(q5)` to Level4/Δ, then `Encrypt(C, Level4, q4)`, `MulRelinNew(X,C)`, `Rescale(q4)` to Level3/Δ. There is no Add operation in this fixture. `Δ=2^45`; q4/q5 come from the constructed E32 LogN13 parameter object. Therefore `(Δ×q5)/q5=Δ` and `(Δ×q4)/q4=Δ` exactly in integer-scale arithmetic. Actual effective values: q4=`549756174337`, q5=`1152921504606830593`; effective Q has 17 entries q0…q16, and the experiment uses only levels 5,4,3.

Inputs were generated and fingerprinted before any ciphertext capture:

| Identity | SHA-256 |
|---|---|
| Config `configs/bootstrap_config.logN13.json` | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` |
| Effective ordered Q/P | `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b` |
| A (`DeterministicInput(4096)`) | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` |
| B (maintained `PublicMulRescaleWorkload(4096)` B values) | `200fdf50e784980cab99289e9c2bbe1930cc43b2ca84df09ac6cde38745dc20f` |
| direct-base C | `4276aa7ef5c475febdab3d4b4fc1e2284f4b4091e2f648f245bf6477aa2caeb7` |
| direct-base combined input identity | `956b256d4a5e60aa4a052968311a7a0ef3fce081f8a3916068cdf1d98afd1d25` |
| direct-base workload identity | `bdafeed50fb2aa9c6ab72c4e10797b9f7dbaeaf5d46dd65bdcb4c92dfcde6d03` |
| direct-c4 C (`4×C`) | `766d36b04d9240867b28f034e2ded7f57e2af8439f90eb1eb7604fa7efba1165` |
| direct-c4 combined input identity | `cce5d432ca1cefcc6ca2d70f37f9e50c463783b5895560fff8b13728fc1a109e` |
| direct-c4 workload identity | `6ed6318094e75d3e8481611d9ede971aebad8deeae5763496d1365274f345786` |

The message formula is the exact spec formula for every `i=0..4095`; C differs from A. The two variants share identical A and B and retain the same q4/q5 scales and parameters. Seeds remain the spec-frozen stage1 pairs `(20261101,20261103)` for direct-base and `(20261105,20261107)` for direct-c4, with each corresponding stage2 seed exactly stage1+1. The descriptive Stagewise RMS screening band remains ±10%; no results were observed during this freeze.

### Strict Q-prefix capacity plan

Capacity uses the shared exact encoded-input coefficient bound `ceil(2×Scale×max|slot|)+1`, negacyclic bound `N×B_X×B_Y`, and nearest-integer Rescale bound `floor((B+floor((q−1)/2))/q)`. It covers plaintext-polynomial centered-lift bounds, not Standard encryption/key-switch noise. The Fast zero-secret measurement mode has c1=0 and uses the encoded plaintext c0; each run must still record the runtime injection helper's independent transactional capacity evidence. At Levels 5/4/3 the authoritative active rows are q0..q3 (4 rows), capacity `q0q1q2q3=5986308565615587353347023386369277282933624144412673`.

| Check | direct-base bound | direct-c4 bound | Active prefix / result |
|---|---:|---:|---|
| Encrypt A at Δ | `2348557866436` | `2348557866436` | 4 rows, strict fit |
| Encrypt B at q5 | `59140416434924413` | `59140416434924413` | 4 rows, strict fit |
| Stage1 MulRelin | `1137825302466909215393461879341056` | same | Level5, 4 rows, strict fit |
| Stage1 Rescale(q5) | `986906131874893` | same | Level4, 4 rows, strict fit |
| Encrypt C at q4 | `27999779486` | `111999117939` | Level4, 4 rows, strict fit |
| Stage2 MulRelin | `226370798107674539948810223616` | `905483192390274484633645277184` | Level4, 4 rows, strict fit |
| Stage2 Rescale(q4) | `411765813054624214` | `1647063252144966668` | Level3, 4 rows, strict fit |

All eight tested `2B < q0q1q2q3` predicates pass for each variant. Neither workload is classified `UNSUPPORTED_CAPACITY`; no input was changed to make it fit.

### Verification and incident notes

- Passed the task-scoped worktree policy unit tests: `python3 -m unittest discover -s scripts -p 'test_task_worktree_policy.py' -v` (2/2).
- Passed only `GOCACHE=/tmp/fixed-sigma-direct-mul-004-gocache go test ./internal/perfmeasure -run '^TestPublicDirectMulWorkloadAndCapacityPreflight$' -count=1 -v`. It constructed the exact LogN13/E32 public parameters, checked both frozen vectors, scales, Q-prefix rows and strict capacities, and printed the hashes/plans above. No ciphertext-chain run, Bootstrap, LogN16 workload, BTS or benchmark was invoked.
- Development-only test correction: the first focused test expected the full effective Q chain to end at q5. The constructed LogN13 bootstrap parameter object correctly has q0…q16; the test was corrected to require q4/q5 availability and then passed. This was a test assertion defect, not a parameter or capacity failure. The first Go invocation also could not read the user's shared build cache; rerunning the same exact test with a task-local `/tmp` cache succeeded.
- CP-A did not expose an unresolved mathematical or architecture choice. The independently authorized direct fixture uses the existing public call path; the conservative bound composition follows the existing helper's documented bound model and the spec's strict Q-prefix condition.
- Counters at CP-A completion: Fast chain runs `0/16`; genuine Standard chains `0/2`; Standard/Fast Bootstrap `0/0`; LogN16 workload `0`; BTS `0`.

## CP-B/CP-C pre-capture attempt ledger — frozen before any chain call

The first Fast `OFF/OFF` capture for each workload and seed pair 1 is also that workload's CP-B Fast baseline. It is counted once and reused as the matching CP-C arm; **no additional Fast baseline chain** will be launched. This is how the CP-B gate fits inside the fixed CP-C ceiling of 16 Fast chain invocations. Standard references are separate: one native Standard `OFF/OFF` chain per workload (2 total). Together the full planned matrix is exactly 2 Standard + 16 Fast calls, with no warmup, retry, replacement sample, or hidden chain invocation. Every launch below consumes its budget slot even if the process exits with an error. A failed launch stops the batch; it is not retried.

Before each launch, this ledger entry will be changed from `NOT_LAUNCHED` to `LAUNCH_COMMITTED` and both exclusive output paths will be recorded. Once launched, it will be marked `PASS` or `FAILED_CONSUMED`; a failed chain will not be replaced. In the table, `captures/...` resolves during execution to `tmp/FIXED-SIGMA-DIRECT-MUL-004/captures/...`; validated evidence is copied to the result directory's `captures/` and hashes are reconciled here. The runner requires a clean committed Primary source and embeds its Git revision, so active attempt state and exclusive output files live under the repository's pre-existing ignored `tmp/FIXED-SIGMA-DIRECT-MUL-004/` scratch directory. The local ledger is reconciled into this journal and the checked-in evidence directory before handoff.

| Slot | Workload | Backend / arm | Stage1 / Stage2 seed | Budget state | Result / vectors |
|---:|---|---|---|---|---|
| S1 | direct-base | genuine Standard / OFF/OFF | 20261101 / 20261102 | PASS_VALIDATED | `captures/standard-direct-base-s1.result.json` / `captures/standard-direct-base-s1.vectors.json` |
| S2 | direct-c4 | genuine Standard / OFF/OFF | 20261105 / 20261106 | PASS_VALIDATED | `captures/standard-direct-c4-s1.result.json` / `captures/standard-direct-c4-s1.vectors.json` |
| F1 | direct-base | Fast / OFF/OFF | 20261101 / 20261102 | PASS_VALIDATED (CP-B gate) | `captures/fast-direct-base-s1-off-off.result.json` / `captures/fast-direct-base-s1-off-off.vectors.json` |
| F2 | direct-base | Fast / ON/OFF | 20261101 / 20261102 | PASS_VALIDATED | `captures/fast-direct-base-s1-on-off.result.json` / `captures/fast-direct-base-s1-on-off.vectors.json` |
| F3 | direct-base | Fast / OFF/ON | 20261101 / 20261102 | PASS_VALIDATED | `captures/fast-direct-base-s1-off-on.result.json` / `captures/fast-direct-base-s1-off-on.vectors.json` |
| F4 | direct-base | Fast / ON/ON | 20261101 / 20261102 | PASS_VALIDATED | `captures/fast-direct-base-s1-on-on.result.json` / `captures/fast-direct-base-s1-on-on.vectors.json` |
| F5 | direct-base | Fast / OFF/OFF | 20261103 / 20261104 | FAILED_CONSUMED; CLI rejected before chain | no result or vectors created |
| F6 | direct-base | Fast / ON/OFF | 20261103 / 20261104 | NOT_LAUNCHED | `captures/fast-direct-base-s2-on-off.result.json` / `captures/fast-direct-base-s2-on-off.vectors.json` |
| F7 | direct-base | Fast / OFF/ON | 20261103 / 20261104 | NOT_LAUNCHED | `captures/fast-direct-base-s2-off-on.result.json` / `captures/fast-direct-base-s2-off-on.vectors.json` |
| F8 | direct-base | Fast / ON/ON | 20261103 / 20261104 | NOT_LAUNCHED | `captures/fast-direct-base-s2-on-on.result.json` / `captures/fast-direct-base-s2-on-on.vectors.json` |
| F9 | direct-c4 | Fast / OFF/OFF | 20261105 / 20261106 | PASS_VALIDATED (CP-B gate) | `captures/fast-direct-c4-s1-off-off.result.json` / `captures/fast-direct-c4-s1-off-off.vectors.json` |
| F10 | direct-c4 | Fast / ON/OFF | 20261105 / 20261106 | NOT_LAUNCHED | `captures/fast-direct-c4-s1-on-off.result.json` / `captures/fast-direct-c4-s1-on-off.vectors.json` |
| F11 | direct-c4 | Fast / OFF/ON | 20261105 / 20261106 | NOT_LAUNCHED | `captures/fast-direct-c4-s1-off-on.result.json` / `captures/fast-direct-c4-s1-off-on.vectors.json` |
| F12 | direct-c4 | Fast / ON/ON | 20261105 / 20261106 | NOT_LAUNCHED | `captures/fast-direct-c4-s1-on-on.result.json` / `captures/fast-direct-c4-s1-on-on.vectors.json` |
| F13 | direct-c4 | Fast / OFF/OFF | 20261107 / 20261108 | NOT_LAUNCHED | `captures/fast-direct-c4-s2-off-off.result.json` / `captures/fast-direct-c4-s2-off-off.vectors.json` |
| F14 | direct-c4 | Fast / ON/OFF | 20261107 / 20261108 | NOT_LAUNCHED | `captures/fast-direct-c4-s2-on-off.result.json` / `captures/fast-direct-c4-s2-on-off.vectors.json` |
| F15 | direct-c4 | Fast / OFF/ON | 20261107 / 20261108 | NOT_LAUNCHED | `captures/fast-direct-c4-s2-off-on.result.json` / `captures/fast-direct-c4-s2-off-on.vectors.json` |
| F16 | direct-c4 | Fast / ON/ON | 20261107 / 20261108 | NOT_LAUNCHED | `captures/fast-direct-c4-s2-on-on.result.json` / `captures/fast-direct-c4-s2-on-on.vectors.json` |

At ledger freeze: Standard `0/2`, Fast `0/16`, Bootstrap `0/0`, LogN16 `0`, BTS `0`. Tests below are validation-only and do not invoke the chain. The test selection has passed independently under `perf_standard` and `perf_fast`: `go test ./cmd/perfprobe -run '^TestFixedSigmaDirectMul|TestFixedSigmaChain' -count=1` with each build tag, using task-local `GOCACHE=/tmp/fixed-sigma-direct-mul-004-gocache`. No package-wide test command was used.

### CP-B implementation review before capture

- The maintained `perfprobe` command now selects the separately fingerprinted `direct-ctct-v1` fixture. The common runner creates independent A/B/C ciphertexts, makes only the two specified public `MulRelinNew→Rescale` calls, and explicitly rejects wrong Level, Scale, degree, NTT, or Montgomery state. Checkpoints carry NTT-domain evidence; Fast active input q0..q3 hashes must recompute from the checkpoint rows, and compact output metadata must show dormant backing rows empty.
- The comparator now dispatches direct artifacts through the direct schema validator and direct Standard/Fast pair validator. It reports the direct strict capacity plan and preserves the Add fixture's existing validation route. Artifact validation recomputes inputs, workload hash, seven plaintext oracles, residual and total-error vectors, scales, dispatch, seeds, injection and capacity evidence. Direct inputs and operation schedule cannot silently fall back to the old Add fixture.
- Coding self-review found that the first comparator draft still routed direct artifacts through the legacy Add-specific validator and would have omitted the direct capacity plan from the analysis output. Both defects were corrected before any chain invocation. A representation assertion was also added for every direct checkpoint; runtime rejects MulRelin/Rescale results outside degree-1 NTT non-Montgomery state.
- After those changes, the focused perfprobe test selection passed under each isolated build tag: `GOCACHE=/tmp/fixed-sigma-direct-mul-004-gocache go test -tags=perf_standard ./cmd/perfprobe -run '^TestFixedSigmaDirectMul|TestFixedSigmaChain' -count=1` and the corresponding `-tags=perf_fast` command. `git diff --check` passed. A transient test compile failure caught a duplicated row-evidence block while wiring the direct validator; the duplicate was removed, and both tagged test commands then passed. This was a coding-only failure and consumed no experiment chain budget.
- The above records were frozen before captures. Capture outcomes and the stop incident are documented below; the initial `0/2` Standard and `0/16` Fast table entry was the pre-capture state only.

## CP-B/CP-C capture checkpoint and stop incident

- CP-B passed using the two native Standard `OFF/OFF` references (S1/S2) and the original Fast `OFF/OFF` references (F1/F9). The paired artifacts matched the frozen public workload/configuration identities, effective parameters, levels/scales, plaintext-oracle checks, strict Q-prefix capacity evidence and expected Standard/Fast dispatch invariants. The full recomputed ablation analysis was not run.
- The first `direct-base` seed pair completed all four Fast arms: F1–F4, seeds 20261101/20261102. For `direct-c4`, only its first-pair Fast `OFF/OFF` reference F9 completed. This is not a complete two-pair CP-C matrix.
- Before finalizing the baseline validators, the scratch pre-next-call checker initially used three stale schema assumptions: it matched evaluator-construction text containing “bootstrap” as if it were a Bootstrap call, expected `canonical_input_sha256` instead of `input_sha256`, and expected `q0123` instead of `q0123_product`. These checker-only assertions were corrected against the maintained artifact schema before continuing. No multiplication chain was relaunched to repair a result, and Bootstrap count remained zero.
- F5 was marked `LAUNCH_COMMITTED` and the process exited 1 before entering the chain. The invocation mistakenly included the positional word `mul-rescale-chain`; this binary accepts global flags and the boolean `--mul-rescale-chain` flag. Go's `flag.Parse` stopped at that positional argument, so the required flags were not parsed and `run()` returned `perfprobe: --config, --out, --backend-commit, --backend-ref, and --secondary-root are required`. The source parser/required-flag check is in `cmd/perfprobe/main.go` around lines 210–255. Neither F5 output file exists; no MulRelin, Rescale, noise injection, or Bootstrap call occurred in F5.
- At the time of the incident, the pre-ruling ledger policy counted every launched slot as consumed and stopped the batch. F5 remains `FAILED_CONSUMED`; it is not changed into a successful or unlaunched record. The incident checkpoint had 5 actual Fast chains, 6 launch attempts and 2/2 Standard chains. Bootstrap 0/0, LogN16 0, BTS 0.
- Seven validated capture pairs (2 Standard + 5 Fast; 55,915,277 bytes total) are preserved in `captures/`. `captures/SHA256SUMS` indexes each result/vector file; the durable copy of the attempt ledger is `attempt-ledger.json`. Ledger SHA-256 is `70fc601b82497ef19a9c4678c9b02d053f9bfa0ea4374af824ed61699ccf7401`; checksum-index SHA-256 is `8245793a1162c756659bddc2193a6d0e251a5f5da634b6a36788986847873545`. Binary SHA-256: Standard `2a44e3f89d055d4a70b294d70b9b265edee3e9f5451c5877a3491f1f2ae2a83c`; Fast `a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff`. Both binaries were built from Primary source commit `977d853541cf59b37e42cdfc1e635bc0cab7862d` with the intended isolated backend module sources recorded in the pre-capture review.
- No aggregate stagewise RMSE, propagation decomposition, tail statistic or frozen ±10% screen was reported at the blocked checkpoint because the matched matrix was incomplete. The fixed σ and original seeds remained unchanged.

## Web ruling and authorized resumption — 2026-10-22

- The user relayed independent Web approval classified `AUTHORIZED_ONE_PRECHAIN_REPLACEMENT`. The ruling confirms F5 failed at CLI parsing before ciphertext work and is not an actual chain execution. The original F5 attempt remains permanently in the attempt history.
- Exactly one additional linked launch record, `F5-R1`, is authorized: same `direct-base` workload, `OFF/OFF` arm, stage seeds `20261103`/`20261104`, frozen σ, E32 LogN13 parameters, Fast source commit and direct-mul source implementation. No new Standard call is authorized.
- Effective accounting is 5/16 actual Fast chains completed, 6/17 Fast launch attempts recorded, and 1/1 known pre-chain parse failure (the original F5). F5-R1 plus the 10 remaining original Fast slots can add at most 11 actual Fast chains. Any additional failure stops immediately; no second replacement is authorized.
- Recovery preflight: Primary began clean on `main@8f203f8beba7fcc06a988c712272bce73c8b51e3`, fetched `origin`, verified `origin/main` was the ancestor, and verified the three ahead commits were task-owned. The three commits were pushed normally; after push local `HEAD` and `origin/main` both equaled `8f203f8beba7fcc06a988c712272bce73c8b51e3`. Normal push first failed DNS resolution for `github.com`; one formal network-permission retry succeeded with fast-forward `5f9174a..8f203f8`.
- Secondary was checked on `fast-qprefix`, clean at `1e7828293c36f9845e30723ba4e9500d84bb50d2`; `git fetch origin` succeeded through the formal metadata-permission retry, and `origin/fast-qprefix` matched exactly. Secondary source remains unchanged.
- The capture matrix and pair validator are pinned to Primary source commit `977d853541cf59b37e42cdfc1e635bc0cab7862d`; the Web ruling commit changes only task documents. Fast capture runs must build and execute from a clean checkout at the pinned Primary revision, with valid Go VCS metadata, and the pinned Secondary source. Keep the original 7 validated captures and SHA index unchanged.
- Before F5-R1, run the no-operation CLI validation with the correct global flag layout and an intentionally invalid C variant. The expected direct-mul variant validation error must occur before config loading/parameter construction; confirm no output files or chain counters change. Only then launch the corrected exact F5-R1 command, with `--mul-rescale-chain` among flags and no positional command word.
- The authorization checkpoint attempt ledger SHA-256 is `dc050c051d6f5ffa57cdbcf456c730bc415bd30b99b228f3138a9c0016af8072`; its accounting is 6/17 Fast launches and 5/16 actual Fast chains, with one authorized replacement slot unused. The seven-pair capture checksum index SHA-256 remains `8245793a1162c756659bddc2193a6d0e251a5f5da634b6a36788986847873545`; all 14 indexed files verify.
- In an isolated temporary copy of the operational ledger, `start F5-R1` was exercised without touching the durable ledger. At that prelaunch checkpoint it recorded launch 7/17 while retaining 5/16 completed chains; the durable ledger was then `NOT_LAUNCHED` for F5-R1.

## F5-R1 consumed failure and stop — 2026-10-22

- The required CLI preflight used the corrected global flag layout and an intentionally invalid C variant. It exited with the expected validation message `direct-ctct-v1 requires --fixed-sigma-c-variant=direct-base or direct-c4` before config loading or parameter construction. No result/vector files were created and the durable ledger hash did not change. The preflight binary was built from Primary `5d1688a27cb57bd5582b8b4e48efa3a6d9446f03`, SHA-256 `0e5a5c22e49565b7af39e4f6a32d5ae08d83662515fbe55e9ff4011c3c1aced4`.
- To retain the captured pair's Primary source pin, the F5-R1 runner was built from a clean detached Primary checkout at `977d853541cf59b37e42cdfc1e635bc0cab7862d` using the same `perf_fast` tag and pinned Secondary `1e7828293c36f9845e30723ba4e9500d84bb50d2`. Its binary SHA-256 was `3e6ab7cfb1d55d75641c7792cd0c4bb845fdc5fc8356b152fa37648f2c008056`; `go version -m` showed no VCS metadata settings. The direct-chain measurement-source fingerprint in the existing captures is `544356df4747d51d9bc5d4acbfdaca02888ba9f417d2130629a3573f96ca236c`; the committed measurement source at the pinned revision was left unchanged.
- F5-R1 used the frozen `direct-base`, `OFF/OFF`, stage seeds `20261103`/`20261104`, E32 LogN13 configuration, and corrected flags with no positional `mul-rescale-chain`. It exited 1 with `perfprobe: Primary binary provenance: binary vcs.revision="" does not match Primary HEAD="977d853541cf59b37e42cdfc1e635bc0cab7862d"`. The failure occurred at `debug.ReadBuildInfo` provenance validation before encryption, MulRelin, Rescale, or noise injection. Neither F5-R1 result nor vector output exists.
- The attempt ledger preserves original F5 as a failed launch, marks F5-R1 `FAILED_CONSUMED`, and records `actual_chain_executed=false`. Counts are now Fast launches 7/17, actual chains 5/16, two pre-chain failures (one authorized F5 CLI parse failure and one unexpected F5-R1 provenance failure), one of one replacement launch consumed, Standard 2/2, Bootstrap 0/0, LogN16 0, and BTS 0. No additional Fast, Standard, Bootstrap, LogN16, or BTS run was made.
- The new failure is outside the one-time Web authorization. CP-C and CP-D stop here; no comparator or aggregate statistics were produced. The original seven capture pairs remain intact: all 14 files pass `captures/SHA256SUMS` (index SHA-256 `8245793a1162c756659bddc2193a6d0e251a5f5da634b6a36788986847873545`). Failure-checkpoint ledger SHA-256 is `d4fce29902cc009fb4b497abf7199c20961dd6ec8c0ed14c184d5c1817ae8e30`. The detached Primary `977d853` checkout and Secondary worktree remain clean and unchanged.

## F5-R2 Web approval and provenance recovery — 2026-10-22

- Web approved exactly one additional F5 recovery launch, `F5-R2`, after classifying F5 and F5-R1 as pre-chain CLI/build-provenance stops. F5 and F5-R1 remain unchanged in the ledger and both remain `actual_chain_executed=false`. This authorization changes only the Fast launch ceiling from 17 to 18; actual Fast chains remain capped at 16. At resume, 7 Fast launches and 5 actual chains were recorded, so at most 11 additional chains remain. No successful Fast/Standard sample is rerun; Standard remains 2/2; Bootstrap, LogN16, BTS and multiparty remain 0.
- Before this work, Primary `main@13eb7948165813d5a679714466b21bd88fe18250` was clean and one local checkpoint ahead of `origin/main`; its diff contained only this task's attempt ledger, journal and summary. It was pushed normally, and a follow-up fetch verified both local `HEAD` and `origin/main` equal `13eb7948165813d5a679714466b21bd88fe18250`. Secondary `fast-qprefix` was clean and equal to `origin/fast-qprefix` at `1e7828293c36f9845e30723ba4e9500d84bb50d2` after fetch. No Secondary source was edited.
- The valid historical Fast binary `a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff` was not present in the searched Primary task build directory or `/private/tmp`. The existing `3e6ab7cf...` and `ffb4355e...` candidate builds lacked Go VCS settings and were not used for any new chain.
- A clean independent Primary clone was created at `/private/tmp/fixed-sigma-direct-mul-004-build-source`, detached at `977d853541cf59b37e42cdfc1e635bc0cab7862d`. Its ordinary `go.mod` replacement `../lattigo` resolves through `/private/tmp/lattigo` to `/Users/xuejinlu/Developer/xuejin-lu/lattigo`, whose branch, clean state and HEAD were verified as `fast-qprefix`, clean, and `1e7828293c36f9845e30723ba4e9500d84bb50d2`. Built the `perf_fast` binary with `GOWORK=off`, the task-local `GOCACHE`, and `-buildvcs=true`, without `-modfile`. `go version -m` confirms `vcs.revision=977d853541cf59b37e42cdfc1e635bc0cab7862d`, `vcs.modified=false`, and `-tags=perf_fast`; the candidate SHA-256 is `bde01f61c36f70dc70939eb12d72f3a42ce07a925942904939b54e516cafae17`. The clone remains clean. The binary has not run a chain.
- Using that binary, the correctly positioned frozen F5-R2 flags plus an intentionally invalid `--fixed-sigma-c-variant=invalid` returned exit 1 with `direct-ctct-v1 requires --fixed-sigma-c-variant=direct-base or direct-c4`. The check created no result or vector files. It was a CLI-only preflight and did not change the durable attempt ledger or consume an experiment launch.
- F5-R2 is recorded as authorized but `NOT_LAUNCHED`, with the original `direct-base`, `OFF/OFF`, seeds `20261103`/`20261104`, σ and source pins. The ledger also records binary SHA, VCS metadata, measurement-source fingerprint, Fast source, and the CLI-only preflight outcome. A scratch-copy exercise of the existing task ledger helper accepted F5-R2 at 8/18 launches and 5/16 chains while preserving F5/F5-R1 and leaving the durable ledger unlaunched; the helper now targets the authoritative result ledger through `FIXED_SIGMA_LEDGER_PATH` and gates F5-R2 on the new authorization plus linked pre-chain history. Helper SHA-256: `99af06be672021432ac904871218c480d4a10500f629dc244c941bfa3f2181a5`.
- This authorization checkpoint ledger SHA-256 is `094950169062f78071ccbefaab79a3b4cc02bd801dd6f77526eeb3877d1e2957`. Reuse binary SHA `bde01f61...` unchanged for the remaining Fast arms so every capture in a seed pair has matching binary/source provenance. F5-R2 may launch only after this checkpoint is committed and pushed. If F5-R2 or any later arm fails unexpectedly, stop with no further retry.
- The mandatory Primary task-path checker reported `TASK_PATHS_ONLY`; its two regression tests passed. `git diff --check` passed. Only the four authorized task files are modified.

## F5-R2 launch reservation — 2026-10-22

- After the authorization checkpoint `d0e849ceda72d55079c098cb8edf01b748ae7a33` was pushed and verified equal to `origin/main`, the final prelaunch gate passed: Primary clone `977d853...` clean, Secondary `1e782829...` clean, candidate binary SHA and Go VCS metadata matched, and both output paths were absent.
- The existing ledger helper's `start F5-R2` action wrote `LAUNCH_COMMITTED`, preserving `actual_chain_executed=null` until the result is classified. It reports Fast launch attempts `8/18` and actual chains before execution `5/16`. This is the sole F5-R2 launch reservation; no process result or ciphertext chain has yet been recorded at this checkpoint.
- The frozen invocation uses the candidate binary in the clean Primary clone, `--mode=public-native --profile=logn13 --ephemeral-secret-weight=32`, the pinned LogN13 config, Fast commit/ref/root `1e782829...` / `fast-qprefix` / the clean Secondary path, `--bootstrap-budget=0 --mul-rescale-chain`, workload `direct-ctct-v1`, arm `OFF/OFF`, C variant `direct-base`, and seed `20261103`. No positional `mul-rescale-chain` argument is present.
