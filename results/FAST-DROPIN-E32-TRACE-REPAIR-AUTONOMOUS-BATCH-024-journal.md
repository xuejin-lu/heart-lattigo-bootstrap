# Batch 024 journal — E32 Rescale trace repair

Status: `BATCH_IN_PROGRESS` — S1–S3 zero-call gates complete; S4 not started.

## Frozen provenance and budget

- Primary synchronized `main`: `4a879e532feabd4b521891f6e99c43a2855bc143`.
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
| S4 isolated E32 diagnostic trace | not authorized until S3 passes; max two | 0 |
| S5 aggregate report and Web handoff | pending | 0 |
