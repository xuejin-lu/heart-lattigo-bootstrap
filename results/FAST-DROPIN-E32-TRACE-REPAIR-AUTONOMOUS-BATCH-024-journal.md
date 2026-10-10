# Batch 024 journal — E32 Rescale trace repair

Status: `BATCH_BLOCKED_NEEDS_WEB_REVIEW` — S1–S3 complete; the authorized S4 session consumed 2/2 Fast-only calls and stopped at event-tree validation. S5 evidence and documentation are complete; awaiting Web review.

## Frozen provenance and budget

- Primary synchronized `main`: `4a879e532feabd4b521891f6e99c43a2855bc143`.
- Current Primary repair commit: `b39ae8519a98e88a902d7868924ae29eb721531c` (pushed; clean and equal to `origin/main`).
- Secondary synchronized `fast-qprefix`: `f7eb9f88d0c877331e62287508c541a1e1147bdc` (clean; matches origin).
- Secondary Batch024 diagnostic commit: `4f2557062cb5c1ffb9a671bc3df67401fd7092b4` (pushed to and verified equal with `origin/fast-qprefix`; clean).
- Genuine Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`: clean detached worktree.
- Formal original Fast pin `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`: clean detached worktree.
- New Batch024 Bootstrap calls: **0**. Batch023's 14/14 reservations remain spent and will not be reused.

## S1 source map — zero Bootstrap calls

- `schemes/ckks/internal/fastcore/rescale.go`,
  `(*RescaleWorkspace).ApplyRows` (base line 153): emits `rescale` parent;
  opens `preflight` at base line 223, delegates coefficient reconstruction,
  rounding/capacity, and residue staging to `rescaleComponent`, then ends
  preflight. It subsequently performs out-of-place compact resize, per-row
  NTT and optional MForm, metadata/scale/domain update, in-place compaction,
  and Q-prefix retention, then ends the parent. No `materialization` or
  `ntt_montgomery_restore` event surrounds that real post-preflight work.
- `rescaleComponent` emits per-component `prefix_to_coefficient` and
  `coefficient_loop` spans and `reconstruct_center_round_capacity` /
  `residue_materialization` records. These are the existing preflight tree;
  do not move or duplicate their arithmetic.
- Existing P93 validator
  `circuits/ckks/bootstrapping/fastdiag_trace_test.go` around base lines
  199–256 requires both direct Rescale children and per-component restore
  events. The E32 validator in `fastdiag_public_e32_test.go` around base line
  492 imposes the same contract. Web review source inspection confirms the
  producer omission; the previous failure is not an arithmetic failure.
- E32 test currently validates event structure before writing its artifact
  (`fastdiag_public_e32_test.go`, base line 328 before artifact serialization).
  Primary `cmd/fastdiag/public_e32.go` preserves the task temp directory on
  test failure, but the raw trace file does not yet exist at that point.
  Batch024 will preserve an owner-only `TRACE_UNVERIFIED` raw file first, then
  write a separate certified artifact only after all required gates pass.
- Prior CPU profile exists at
  `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-4268150718/profiles/warm-cpu.pprof`;
  SHA-256 matches Batch023 record:
  `c12b7e3bcb63fe935e5750573f5019a4cb102640328f90b5d63cbebc35755162`.
  Local `go tool pprof` is unavailable (`go: no such tool "pprof"`), so no
  profile symbol attribution is claimed from this prior artifact.
- Reuse boundary: keep the existing `internal/fastdiag` `Begin`/`End`, event
  model and P93/E32 validators; only add the authorized guarded spans at the
  existing `ApplyRows` commit boundary. Keep Primary's current `cmd/fastdiag`
  fixture, attempt reservation, JSON provenance and event aggregation; extend
  raw-preservation/closure handling there rather than adding a new runner.

## S2 implementation — ZERO Bootstrap calls

- Secondary production delta is confined to the approved `rescale.go` path and
  consists only of guarded span declarations/begins/ends. No existing source
  lines are removed or moved. `materialization` encloses compact output resize,
  per-component NTT/optional MForm, metadata and Scale/domain state, in-place
  compaction, and Q-prefix retention. One nested
  `ntt_montgomery_restore` span wraps each real component transform loop.
- Direct real `ApplyRows` regression fixtures cover authoritative widths 1–4,
  both Montgomery states, in-place/out-of-place output, no-trace versus traced
  exact ciphertext equality, metadata/domain state, expected per-component
  event hierarchy, measured parent/direct-child closure, and capacity-preflight
  failure without fabricated commit events.
- E32 now atomically saves an owner-only `TRACE_UNVERIFIED` raw artifact before
  event structural assertions. A separate owner-only certified artifact is
  written only after output oracle, event tree, profiles and input-integrity
  gates pass. Failure sidecars preserve the exact structural validator reason,
  raw artifact hash/path, current and fixture Primary commits, diagnostic and
  production Fast refs, and reserved/actual call counts. The Primary adapter
  reports all preserved paths on subprocess failure and verifies raw/certified
  hash and captured-event linkage.
- Source provenance distinguishes fixture-origin Primary
  `6e938918442c409faa6d32e159c7a3f7a1041f7a` from the current harness commit.
  The held Batch023 measurement source SHA remains
  `b63fef5bbbdfe1e679b2e50ab66e0a5c43afefc526bd2228b16b7541ae01ea63`; a
  read-only recomputation over `cmd/perfprobe`, `internal/perfmeasure`,
  `internal/numericalmetrics`, `go.mod`, and `go.sum` matches exactly.
- Primary `cmd/fastdiag` tests and `go vet` pass. Secondary regular and
  `fastdiag`-tagged Rescale/E32 focused tests, package compile-only checks and
  vet pass. No P93 end-to-end test was run: `TestFastDiagP93Q55Trace` performs
  real Bootstrap calls; the old P93 event validator is exercised using a
  synthetic complete rescale tree instead. Original Batch023 held-input
  `FASTDIAG_VALIDATE_ONLY=1` passed with zero artifacts and zero calls.

## S3 zero-call gates

- The Primary verifier now hashes and compares every non-test Go production
  blob against the frozen Fast pin, allows only the exact
  `schemes/ckks/internal/fastcore/rescale.go` delta, checks that its unified
  diff is exactly the approved span-only additions (no removed lines or
  arithmetic), and verifies `go.mod`/`go.sum` blob identity. Test-only files do
  not enter either production-source digest.
- Primary E32 analysis records parent keys, all measured inclusive/direct-child
  closures and signed unattributed residuals; it never normalizes closure.
  Positive-exclusive Pareto entries are grouped by each independent event-tree
  root (Bootstrap, generated powers, Rescale) and ranked from measured
  parent-minus-child time; root shares are local to that tree and are never
  added across unparented/overlapping roots. Stage-only 2x/4x and
  elimination-bound Amdahl scenarios use observed inclusive stage/root shares
  and are explicitly hypothetical. CPU profile symbol attribution is attempted
  separately in flat and cumulative views.
- No attempt tokens were created or spent. Batch024 Bootstrap calls remain
  **0**. Existing Batch023 attempt journals remain untouched.
- Primary's integration test against the committed/pushed Secondary HEAD
  passed: it computed separate production and diagnostic Go-tree hashes,
  reported exactly `schemes/ckks/internal/fastcore/rescale.go`, and accepted
  the unified diff only after verifying all authorized span additions with no
  removed arithmetic or other production-source changes.
- The final Secondary commit passed a second exact Batch023 held-input
  `FASTDIAG_VALIDATE_ONLY=1` invocation using Primary HEAD
  `4a879e532feabd4b521891f6e99c43a2855bc143`; no output/profile files were
  created. Frozen genuine Standard and formal-original-Fast detached worktrees
  remain clean at their pinned HEADs.

## Checkpoints

| Checkpoint | State | Bootstrap calls |
|---|---|---:|
| S1 source/provenance audit | complete | 0 |
| S2 regression-first span and raw evidence implementation | complete | 0 |
| S3 enabled/disabled, committed-source identity, and zero-call validation | complete | 0 |
| S4 isolated E32 diagnostic trace | executed; raw trace preserved, structural validation failed; no retry | 2 |
| S5 aggregate report and Web handoff | complete; blocked for independent Web review | 2 |

## S4 zero-call preflight stop

- Invoked the fixed `logn13-e32-public` trace command only after Primary and
  Secondary were clean, synchronized, and pinned; the runner stopped in
  `validatePublicE32Fixture` before creating its isolated diagnostic checkout,
  attempt journals, or launching any Bootstrap.
- The held manifest is `fast-public-e32-trace-fixture.v1` and expects a
  `fast-standard-public-native-repeatability-vectors.v1` artifact in mode
  `public-native-repeatability`, including exactly six
  `bootstrap_outputs`.
- The supplied `fast-vectors.json` is instead
  `fast-standard-public-native-vectors.v1`, mode `public-native`, and contains
  no `bootstrap_outputs`. Backend/ref, Primary source, config, Q/P, input, and
  workload hashes match the manifest, but the schema/mode and required output
  set do not. The exact validator error is
  `uninstrumented Fast vectors do not match the E32 trace fixture provenance`.
- Manifest SHA-256:
  `5bc19b1eb6e67997ca3d19f39e6fd781bf93b27ad305a1be733f0c883611b847`.
  Supplied vector SHA-256:
  `d0334680ac337d8d507a0074750e677dcf0d0ccdda47f17b7219c2655c5b0fc1`.
- Batch024 attempt reservations: **0/2**; Bootstrap calls: **0**. Both
  attempt-journal paths, the raw/certified trace artifacts, and the requested
  summary JSON remain absent. Batch023's **14/14** calls remain spent and were
  not reused. No retry or alternate fixture was attempted.
- Resume requires Web review to identify the authoritative matching
  repeatability vectors artifact and decide whether a fresh S4 invocation is
  permitted under the no-retry rule. Until then, do not launch Bootstrap.

## 2026-10-11 — Web-authorized fixture recovery; worktree creation stop

- Primary safely synchronized `main` to `3413374b3bc830934a00a1ae609fcb4295b90430`;
  `origin/main` matched. Secondary safely synchronized `fast-qprefix` to
  `4f2557062cb5c1ffb9a671bc3df67401fd7092b4`; its remote matched. Both
  worktrees were clean.
- Found the authentic Batch023 Fast repeatability vectors at
  `/private/tmp/fast-dropin-batch023-final.3VxTOp/fast-repeatability-vectors.json`,
  SHA-256 `63ad89ce9ccc70a0b737397810d4f4a15d8a1ed65e82ccd3af0435d1df4d6881`.
  Metadata matches the held manifest: repeatability schema/mode, original Fast
  commit `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, fixture Primary commit
  `6e938918442c409faa6d32e159c7a3f7a1041f7a`, six named 4096-slot finite
  Bootstrap outputs, and matching source/config/QP/input/workload hashes.
  Current Primary measurement-source fingerprint recomputed to
  `b63fef5bbbdfe1e679b2e50ab66e0a5c43afefc526bd2228b16b7541ae01ea63`.
- Re-ran the Secondary test-only fixture gate with
  `FASTDIAG_VALIDATE_ONLY=1`:
  `go test -tags=fastdiag ./circuits/ckks/bootstrapping -run '^TestFastDiagPublicE32Trace$' -count=1`
  — PASS, zero Bootstrap calls. A Primary compile-only check with the
  task-owned writable `GOCACHE=/private/tmp/batch024-go-cache.TQR9yT` also
  passed: `go test ./cmd/fastdiag -run '^$' -count=1`.
- The authorized outer trace command reached Primary fixture and production
  source identity validation. Exact invocation:
  `GOCACHE=/private/tmp/batch024-go-cache.TQR9yT go run ./cmd/fastdiag trace --profile=logn13-e32-public --trace=stage,power,rescale --e32-manifest=/private/tmp/fast-dropin-batch023-final.3VxTOp/fast-e32-trace.manifest.json --fast-vectors=/private/tmp/fast-dropin-batch023-final.3VxTOp/fast-repeatability-vectors.json --out=/private/tmp/fastdiag-e32-fixture-validation.gNDA1K/batch024-e32-trace.json`.
  It then failed at
  `git worktree add --detach /var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2472111079/secondary-diagnostic 4f2557062cb5c1ffb9a671bc3df67401fd7092b4` with
  `fatal: could not create directory of '.git/worktrees/secondary-diagnostic1': Operation not permitted`.
  The runner creates the worktree before launching the Secondary test or
  reserving attempts; source order and filesystem checks confirm this failure
  was before both. The selected output base and both Batch024 reservation
  files remain absent. No raw/certified trace exists.
- Final ledger remains Batch024 **0/2 reserved, 0 Bootstrap calls**; Batch023
  **14/14 spent, 0 reused**. After this worktree-metadata failure, no
  escalation, alternate worktree path or repeat trace invocation was
  attempted. Primary/Secondary worktrees remain clean.
- **STOP:** `BATCH_BLOCKED_NEEDS_WEB_REVIEW`. The exact blocker is sandbox denial
  writing Secondary Git worktree metadata, not the Go build cache or fixture.
  Do not launch another trace until Web provides a new decision.

## 2026-10-11 — path-alias gate passed; S4 calls consumed; trace unverified

- Primary startup synchronization completed on clean `main`; `origin/main` and
  local HEAD were `4d77c511b27ea303fc5a65870dff1404011d1807`. The first
  unprivileged fetch and ff-only merge hit Git metadata write denial; formal
  permission escalation succeeded for both. No task or source files were
  changed by synchronization.
- Re-read the synchronized task/spec and latest clone-path Web review.
  `go test ./cmd/fastdiag -count=1`, `go vet ./cmd/fastdiag`, and
  `git diff --check` passed. The Secondary validate-only E32 fixture test
  passed with zero Bootstrap calls. The authentic six-output Batch023 vector
  artifact and held manifest hashes remained verified; the task-owned
  independent clone was pinned to diagnostic commit
  `4f2557062cb5c1ffb9a671bc3df67401fd7092b4`.
- The single authorized S4 invocation reserved attempt 1
  (`diagnostic_fast_cold`) and attempt 2 (`diagnostic_fast_traced_warm`)
  before each call. Both records are `irrevocably_reserved_before_call`.
  Exactly two Fast-only calls completed; this exhausts the Batch024 budget.
  No Standard call, formal performance run, retry, or further Bootstrap ran.
- Both decoded calls passed the native oracle: RMSE
  `1.1795785978778852e-9`, maximum complex difference
  `5.738790672737443e-8` against gate `1e-6`. Their Fast-reference RMSE and
  maximum complex difference were both zero. These output checks do not
  certify the event trace.
- Raw trace retained outside Git:
  `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2533882906/raw-trace-unverified.json`,
  177,696 bytes, 539 events, SHA-256
  `f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74`.
  Failure sidecar:
  `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2533882906/trace-failure.json`,
  SHA-256 `59a872b2df18ca32859f4753650c3f1e14633dd79d1358b283204c4445b3f907`.
- First structural mismatch: event sequence `61` is power `T8` with parent
  sequence `60`, which is power `T16`; expected direct parent is the
  `generated_powers` event sequence `59`. Validator error:
  `power event is not nested under generated_powers`. Trace status is
  `TRACE_UNVERIFIED`; no certified result or stage/power/rescale attribution,
  Pareto ranking, Amdahl estimate, or performance claim is made. Observed
  elapsed values remain raw diagnostics only.
- Batch023 remains 14/14 spent and untouched; its separately accepted 4.225x
  observation is historical and is not a Batch024 result. Original Standard
  pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and original production Fast
  pin `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` remain read-only. Secondary
  diagnostic source was not modified in this final session.
- Final read-only repository check: Secondary `fast-qprefix` is clean at
  `4f2557062cb5c1ffb9a671bc3df67401fd7092b4`, equal to
  `origin/fast-qprefix`; the original Fast pinned checkout is clean at
  `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, and the task-local diagnostic
  clone is clean. The genuine Standard commit object is present, but its old
  registered detached checkout is unavailable/prunable, so its current
  worktree cleanliness could not be rechecked. No Standard code or run was
  touched in this final session.
- Final status: `BATCH_BLOCKED_NEEDS_WEB_REVIEW`. S5 report, compact evidence,
  and source-supported measurement-platform update are complete. Do not retry,
  change event semantics, or spend further calls without Web direction.
