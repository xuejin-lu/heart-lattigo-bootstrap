# FAST-BOOTSTRAP-NOISE-FIDELITY-001 journal

## Activation and budget

- Activated by explicit user instruction on 2026-10-27; the remote `origin/main` pointer was still `FIXED-SIGMA-DIRECT-MUL-004` and the remote spec still said `PLANNED`. The local task pointer/spec were already changed to this task by the preceding activation step and are preserved as user-authorized task changes. No remote files were checked out over them; Primary HEAD equals `origin/main` at `67d1f572aa3ae9f0d2b28031123145b411fcfc3e`.
- CP-A/CP-B are authorized. CP-C authorizes at most one successful full Bootstrap per arm, three total, no retry or implicit warmup. No Bootstrap, LogN16 workload, BTS, or large benchmark has been run in this task.
- Primary source checkout: `main`, clean except for the user-authorized local activation changes and this journal. Secondary implementation checkout: `/private/tmp/lattigo-fast-bootstrap-noise-fidelity`, branch `experiment/fast-bootstrap-noise-fidelity`, at `1e7828293c36f9845e30723ba4e9500d84bb50d2`. The ordinary `fast-qprefix` worktree remains clean and at the same frozen baseline; no edits were made there. Both origins were fetched successfully using the approved escalation after normal Git metadata writes returned `Operation not permitted`. Remote and local task branches were equal; no merge was needed.
- Genuine Standard pin: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Frozen Fast control: `1e7828293c36f9845e30723ba4e9500d84bb50d2`. Profile candidate: LogN13/E32/Scale 2^45, subject to exact effective-parameter and input preflight. Fixed sigma: `4.004280012097432`; not tuned.

## CP-A — source and platform review (no Bootstrap)

### Reuse inventory

- Primary maintained owners found: `cmd/perfprobe`, `cmd/fastdiag`/`scripts/fastdiag`, `internal/perfmeasure`, and `internal/numericalmetrics`. `perfmeasure.ParametersFromConfigWithE` builds the explicit E32 parameters and owns effective Q/P/profile metadata and deterministic input fingerprints. `numericalmetrics.SummarizeComplexVector` provides RMSE/max/percentile summaries. `perfprobe` owns public-native lifecycle, source/build provenance, preflight, Bootstrap reservation, output decode and comparison; no new runner or metric implementation is planned. `fastdiag` remains Fast-only internal attribution and is not a substitute for the public Bootstrap measurement.
- Existing `perfprobe --public-bootstrap` accepts exactly two calls per backend (one cold and one warm); `--public-repeatability` accepts six. `publicBootstrapAttemptSpecs` rejects a budget of one. This task authorizes one call per arm and no hidden warmup, so CP-B must extend the existing `perfprobe` owner with an explicit single-call pilot mode and zero-call regression tests; using the current two-call path would exceed authorization.
- `perfprobe.validatePublicBackendPin` currently recognizes the frozen `publicFixedSigmaChainFastSHA` constant `1e7828293c36f9845e30723ba4e9500d84bb50d2` only through other workflow assumptions, not as an allowed public-native Fast pin. CP-B must add a narrowly scoped approved LogN13 pin for this task, leaving other profiles and genuine Standard validation fail-closed.

### Fast source map and coverage

- Public `bootstrapping.NewEvaluator` selects the Fast implementation only for the Fast-compatible evaluation-key marker (`circuits/ckks/bootstrapping/evaluator.go`). `bootstrapping.NewFastEvaluator` creates one `schemes/ckks/fast.Evaluator`; the Fast polynomial evaluator and Fast Mod1 evaluator share it (`circuits/ckks/bootstrapping/fast_scaledown.go`, `fast_bootstrap.go`).
- `schemes/ckks/fast.Evaluator.MulElementQPrefixRows` and `MulRelinElementQPrefixRows` both reach `mulElementRows` and then the existing `fastcore.MulWorkspace.ApplyRows`. `ApplyRows` distinguishes degree-one × degree-one ciphertext multiplication from degree-one × degree-zero plaintext multiplication. This shared Fast boundary covers lazy polynomial products and relinearized polynomial products (`circuits/ckks/polynomial/fast.go`), as well as Mod1 DoubleAngle squaring (`circuits/ckks/mod1/fast.go`). These are the actual Fast Bootstrap ct×ct call sites found in production sources; `MulThenAdd` use in polynomial evaluation is scalar accumulation, not a missed ct×ct path.
- The Mod1 square passes `res` and `res.El()` which alias the same underlying ciphertext. H1 explicitly requires two independent local copies and independent side seeds, so preserving the core's pointer-identity square optimization would be incorrect in the enabled path; OFF must continue to pass the original pointers unchanged.
- A narrow hook at the Fast evaluator's shared `mulElementRows` boundary is the candidate interception point. The Fast subpackage already imports parent `schemes/ckks`, so the existing parent helper can be reused without a new import cycle; a helper evaluator can be initialized from the same Fast parameters. `op1.Degree()==0` remains a strict ct×pt exclusion.

### Noise helper, representation, and capacity

- Existing `schemes/ckks/evaluator_fast_fixed_sigma_noise.go` samples a length-N rounded signed Gaussian vector, reduces the same coefficients into each active Q row, NTT-transforms each row, creates a candidate c0, calls the existing `fastcore.RescaleWorkspace.CheckQPrefixAdditionCapacity`, and commits only c0 after success. The operation preserves c1 and metadata and is transactional with respect to its target ciphertext.
- That helper currently accepts only LogN13, Level3/4, Scale 2^45, four rows, NTT and non-Montgomery ciphertexts. Those limits describe the old post-operation fixture and do not prove the internal Bootstrap hook's varying levels/scales. H1 specifies coefficient-domain c0 addition with unchanged metadata; CP-B must generalize only the helper's profile gate to the actual per-call Q-prefix width while retaining NTT/non-Montgomery/c1=0/canonical-residue checks and the existing centered-capacity check. Unit tests must cover multiple Levels and scales, and fail without mutating inputs when capacity/shape validation fails. No Bootstrap path has been executed to sample levels/scales.
- For a right operand represented as `*rlwe.Element[ring.Poly]`, an enabled ct×ct hook must make a ciphertext wrapper from a deep copy of that element (including its metadata), not mutate `op1` or `op0`, then inject independently into each local c0. The core can then consume the two local copies. All this is explicit in the current H1 contract; no ciphertext storage or Q-prefix policy change is authorized.

### CP-A checks and status

- Primary `git fetch origin` and Secondary `git fetch origin experiment/fast-bootstrap-noise-fidelity fast-qprefix` succeeded after ordinary commands returned exact `.git/*/FETCH_HEAD: Operation not permitted`; the formal escalation succeeded. Branch, ancestry and worktree checks show no remote divergence and no unauthorized Secondary edits.
- Primary task path guard accepted only `CURRENT_TASK.md` and `specs/FAST-BOOTSTRAP-NOISE-FIDELITY-001.md`. `python3 -m unittest discover -s scripts -p 'test_task_worktree_policy.py' -v`: 2 tests passed.
- All source review and these checks were non-Bootstrap. No Standard or Fast Bootstrap attempt reservation/call has been made.
- CP-A outcome: source paths and reusable capacity helper are identified. CP-B may proceed only by preserving the explicit coefficient-domain H1, dynamic active-prefix check, default-off direct path, and one-call public measurement boundary. Any capacity rejection during one of the three authorized calls is a terminal no-retry failure under the task charter.

## CP-B — minimal hook and one-call measurement support (no Bootstrap)

### User steering and scope decision

- The user confirmed the research priority is matching final Bootstrap precision with minimal Fast changes, not rebuilding a full CKKS noise model. The proposed patch remains limited to the one Fast `mulElementRows` ct×ct boundary, explicit default-off configuration, reusing the existing Gaussian/RNS/NTT and Q-prefix capacity helper, and the maintained Primary `perfprobe` path. No production application branch, arithmetic redesign, or Standard modification was introduced.
- The Primary task activation was versioned and pushed at `cf19fbbc1dfd893bdb9679909ef4593391be351f`; this records the active task pointer/spec, not a Bootstrap capture. The activation decision does not change `FIXED-SIGMA-DIRECT-MUL-004` artifacts.
- Secondary branch instructions were checked against the active task exception. The task explicitly selects `experiment/fast-bootstrap-noise-fidelity`; the branch `AGENTS.md` permits an explicitly authorized branch exception. The original `fast-qprefix` branch/worktree was not changed.

### Minimum implementation and coverage

- Secondary implementation was committed and pushed as `b81f54c2e0f98dc2b4a7474fb333fbae79be8c63`, parented by the frozen Fast control `1e7828293c36f9845e30723ba4e9500d84bb50d2`. It factors the existing coefficient injection body, adds only a Bootstrap-specific LogN13 helper gate for positive scales and active Q-prefix levels, and routes the explicit hook through `mulElementRows` only when both operands have degree one. The parent `ckks` package already imports no Fast package; `fast` already imports parent `ckks`, so helper reuse adds no import cycle. The receiver mismatch is handled by constructing a parent `ckks.Evaluator` only at opt-in configuration time.
- The enabled route copies both ciphertext operands before injection, including the aliased Mod1 square case, derives separate deterministic seeds by operation and side, and sends those temporary operands into the unchanged Fast `MulCore.ApplyRows`. `ct×pt` and scalar multiplication remain outside the degree-one/degree-one gate. Rotation uses its separate Automorphism path. OFF enters the original core directly without copies or sampling.
- The inherited helper's historical post-operation `Level3/4, Scale=2^45` gate remains intact. The new Bootstrap-only helper checks actual logical level, active Q-prefix backing, NTT/non-Montgomery state, canonical c0 residues, zero c1, positive Scale, and existing centered-Q-prefix capacity before committing c0. It leaves metadata and c1 unchanged. The helper reports each operand's Level/Scale, sampled moments/hash, and capacity evidence.
- Bootstrap evaluator setup and hooks were added only as explicit methods on the existing Fast evaluator and ordinary public Bootstrap wrapper. The Standard adapter always rejects enabling the hook and serializes no noise evidence. No genuine Standard arithmetic/source was modified.

### Primary platform extension

- `cmd/perfprobe` now has a distinct exact `--public-bootstrap-once` mode requiring `--bootstrap-budget=1` plus the already-passing zero-call preflight pair. It creates exactly one `single_bootstrap` reservation/call, no hidden warmup or repeat, and preserves exclusive per-attempt journaling/failure capture. Fixed-sigma mode is accepted only for Fast, the one-call pilot, LogN13, and an explicit root seed; sigma is frozen in code at `4.004280012097432`.
- The existing public-native vectors now persist the exact canonical plaintext entering Bootstrap (the cleartext oracle after the frozen A/B/C Mul+Rescale+Rotation workload) and its SHA. One-call Bootstrap outputs are checked against that same plaintext, rather than against each arm's separately decrypted pre-Bootstrap ciphertext.
- A `compare-public-triplet` subcommand within the maintained `perfprobe` owner validates all three arm artifacts, exact Standard/Fast source pins, equal effective parameters/workload/host/preflight/plaintext, and identical Fast/Fast+Noise binary SHA. It uses `numericalmetrics` for global input-relative RMSE, `-log2(RMSE)`, MaxAbs/p99/p99.9, and output pairwise errors; records the measured single-call time/allocations, Fast speed ratios and noise overhead; and explicitly reports peak memory as unavailable. Zero/nonfinite precision has an explicit undefined status. This comparison is exploratory and does not turn the ±2-bit screen into an acceptance criterion.
- The Fast noise evidence validator checks exactly two injections per recorded ct×ct event, distinct left/right seeds, fixed sigma, completed operation status, per-side sample count and coefficient SHA, active prefix rows, strict-capacity status, and recomputed prefix products against the effective Q primes. A post-call evidence failure is persisted as `STOPPED_NO_RETRY`.

### Focused verification

- Primary, Fast-tagged targeted tests passed with `GOWORK=off`, the isolated modfile pointing at the Secondary task worktree, and `GOCACHE=/private/tmp/gocache-fast-bootstrap-noise-fidelity`:
  - `TestPublicBootstrapAttemptSpecsOneCallHasNoImplicitWarmup`
  - `TestPublicBootstrapOnceBudgetAndReservationAreExactlyOne`
  - `TestPublicBootstrapOnceCLIRejectsBudgetDriftOrMultipleModes`
  - `TestPublicBootstrapNoiseSnapshotRequiresTwoIndependentCapacityCheckedSides`
  - `TestPublicTripletSummaryUsesGlobalRMSEPrecisionAndTails`
  - `TestPublicBootstrapOncePairArtifactRequiresOneMatchedOutputAndCommonPlaintext`
  - `TestRunPublicTripletCompareUsesSameOriginalInputAndValidatesSharedFastBinary`
  - `TestPublicBootstrapPairRequiresBoundedNativeOutputs`
- The corresponding `perf_standard` Primary package build compiled with `-run '^$'`; it executed no Bootstrap test. Secondary targeted tests passed in `schemes/ckks`, `schemes/ckks/fast`, and `circuits/ckks/bootstrapping`: existing-helper dynamic level/scale/capacity cases; default-off bit-exact path; independent seeds; in-place alias-square equivalence; `ct×pt`/scalar/rotation exclusion; active prefix levels; Montgomery rejection without operand mutation; and public evaluator hook configuration. All tests use synthetic ciphertext operations or config only; none calls Bootstrap.
- Two CP-B regressions were caught by pure tests and fixed before this checkpoint: the new one-call output name was missing from the public checkpoint scale resolver, and a first snapshot decoder assumed flattened capacity JSON instead of the helper's nested `qprefix_capacity` object. Final focused suites above pass with the actual nested schema. A temporary compile attempt to compare `Scale.Value` with nil was rejected because `big.Float` is a value field; that guard was removed, leaving the positive-sign check.
- `git diff --check` passed for both repositories. The ordinary Secondary fetch encountered a Git metadata `Operation not permitted` error at the linked worktree's `FETCH_HEAD`; the approved escalation fetch succeeded and showed the remote task branch remained exactly at its parent commit before commit. Normal push had a transient SSH DNS resolution failure; the approved normal fast-forward push succeeded. No reset, stash, rebase, force push, or Fast control branch mutation occurred.

### CP-B gate and next authorization boundary

- CP-B code and focused regression gates pass. The committed Fast implementation pin for both F and FN is now `b81f54c2e0f98dc2b4a7474fb333fbae79be8c63`; the Primary allowlist is pinned to this exact commit for LogN13 only. Genuine Standard remains pinned to `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- No real Bootstrap call, LogN16 workload, BTS, or large benchmark has been run. CP-C's charter budget remains exactly one successful call in each of S/F/FN, maximum three total, no retry. Before any of those calls, Primary must commit/push this measurement/comparator gate, build against clean exact source pins, verify `go version -m` and binary SHA, run only zero-Bootstrap public-native preflights and their comparator, then version an immutable three-slot attempt ledger. Any provenance, effective capacity, or preflight mismatch stops CP-C.

### Minimal-change feasibility review and planned surface

User steering reaffirmed the primary goal as matching full-Bootstrap precision with a small Fast-only change, and requested an explicit file/impact review before broader implementation. The review found no unresolved import-cycle, receiver, initialization, or Level/Scale issue requiring a broad rewrite. The small CP-B implementation below is the approved surface; it does not alter the Fast multiplication algorithm or genuine Standard.

#### Secondary Fast implementation surface (already committed on the task branch)

- `schemes/ckks/evaluator_fast_fixed_sigma_noise.go`: factor/reuse the existing Gaussian coefficient sampler, Q reduction/NTT injection, and centered Q-prefix capacity check; preserve the old helper's historical profile gate and add the narrowly scoped Bootstrap diagnostic entry point.
- `schemes/ckks/fast/fixed_sigma_bootstrap_noise.go`, `schemes/ckks/fast/evaluator.go`, and `schemes/ckks/fast/evaluator_ntt.go`: add a nil-by-default hook at the shared Fast ct×ct `mulElementRows` boundary. When explicitly enabled, copy both inputs, inject separately seeded noise into the copies, and call the unchanged `MulCore.ApplyRows`; when disabled, go directly to the unchanged core route. The degree-one/degree-one gate excludes ct×pt; scalar accumulation and Rotation do not enter this hook.
- `circuits/ckks/bootstrapping/evaluator.go` and `circuits/ckks/bootstrapping/fast_bootstrap.go`: pass the opt-in configuration/snapshot through the ordinary Bootstrap evaluator while rejecting the feature on Standard. No application-side Fast arithmetic branch was added.
- `schemes/ckks/evaluator_fast_fixed_sigma_noise_test.go`, `schemes/ckks/fast/fixed_sigma_bootstrap_noise_test.go`, and `circuits/ckks/bootstrapping/fast_bootstrap_noise_config_test.go`: retain focused helper, alias-square, OFF-equivalence, operation-exclusion, and public-wrapper regression coverage. These tests do not call Bootstrap.

Impact is limited to explicitly enabled Fast Bootstrap ct×ct multiplication calls and their temporary operands. In particular, Mod1's in-place square is copied into independent left/right temporaries before the separate draws, so the caller's original ciphertext is not contaminated. The same Fast binary is intended for F and FN; the only difference is the explicit hook setting.

#### Primary measurement surface (uncommitted at time of this review)

- `cmd/perfprobe/main.go`, `cmd/perfprobe/public_native.go`, `cmd/perfprobe/public_backend_fast.go`, and `cmd/perfprobe/public_backend_standard.go`: extend the existing public-native owner with an exact one-call/no-warmup mode, Fast-only hook configuration/evidence, and canonical plaintext input evidence. Standard refuses the Fast-only hook.
- `cmd/perfprobe/public_triplet.go` and `cmd/perfprobe/public_triplet_test.go`: add a three-artifact comparison inside the existing `perfprobe`, reusing `internal/numericalmetrics`; verify common source/config/effective parameters/host/input and shared Fast/FN binary before reporting input-relative RMSE, precision bits, tails, pairwise errors, time, allocations, and descriptive speed ratios. Peak memory is marked unavailable because the maintained timing API does not expose it. This is not a new runner or statistics package.
- `results/FAST-BOOTSTRAP-NOISE-FIDELITY-001/journal.md`: record the review, implementation evidence, and future stop conditions.

Primary production impact is limited to the existing measurement/adapter boundary; shared application workload and ordinary CKKS API remain unchanged. CP-B targeted Fast and Standard compile checks pass, the exact task-path guard passes, and `git diff --check` passes. No new scientific or representation decision was introduced beyond H1 as already stated in the activated spec. No Bootstrap has run; the three-call CP-C ledger and zero-call provenance preflights remain prerequisites before the separately budgeted pilot.

The final Primary regression pass also deliberately changed only the synthetic Fast+Noise artifact's binary SHA and confirmed that the triplet comparator rejects the pair; this guards the same-binary requirement without invoking CKKS operations.

### CP-C pre-call build and source checks

- The CP-B Primary checkpoint `ec43fe112441f26ad22480e7025db90c24b42e32` is pushed and clean. The task Fast worktree is clean at `b81f54c2e0f98dc2b4a7474fb333fbae79be8c63`; the genuine Standard source is a newly created, clean detached worktree at the pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. The root `fast-qprefix` checkout remains unchanged.
- A preliminary Fast and Standard build confirmed `go version -m` carries Primary `vcs.revision=ec43fe112441f26ad22480e7025db90c24b42e32`, `vcs.modified=false`, `go1.26.4`, `darwin/arm64`, and the correct respective local Lattigo replacement path. Those preliminary Binary SHA values are not designated for a capture: the attempt-ledger commit below will change the Primary VCS revision, so final binaries must be rebuilt and their metadata/hash rechecked from the post-ledger clean source before any preflight or Bootstrap call.
- No public-native preflight or Bootstrap call has yet run. CP-C remains at zero calls; the next action is to version the immutable 3-slot authorization/budget ledger, then rebuild and run only zero-Bootstrap preflights.
