# Batch 024 — E32 Rescale trace repair

Status: BATCH_BLOCKED_NEEDS_WEB_REVIEW.

## Outcome

- S1–S3 source, instrumentation and zero-call gates passed. The Primary
  isolated-clone path-alias fix also passed focused tests, vet, and the
  zero-Bootstrap E32 fixture validation.
- The only Secondary production delta was guarded event-span instrumentation
  in `schemes/ckks/internal/fastcore/rescale.go`, around the existing
  RescaleWorkspace.ApplyRows materialization and actual per-row NTT/MForm
  restore work; arithmetic and operation order were not changed. Primary
  analysis retained per-event interfaces and added event-type summaries
  using parent-minus-direct-child elapsed time, partitioned by event-tree root.
- The authorized S4 session consumed exactly **2/2** Fast-only Bootstrap
  calls: one cold initialization and one traced warm call. No Standard or
  formal-performance Bootstrap was run; there was no retry.
- Both decoded outputs passed the frozen native oracle and matched the held
  Fast repeatability reference: oracle RMSE 1.1795785978778852e-9,
  oracle max complex difference 5.738790672737443e-8 (gate 1e-6), and
  Fast-reference RMSE/max difference 0.
- Structural validation failed on the first nonconforming power parent:
  event sequence 61, T8, has parent sequence 60 (T16); the validator
  requires generated powers to be direct children of generated_powers
  sequence 59. The raw trace is therefore TRACE_UNVERIFIED; no certified
  trace, event attribution, Pareto ranking, or Amdahl estimate is claimed.
- The observed cold/warm elapsed values are retained in the raw evidence only
  and are not a performance result. Batch023's separately accepted 4.225x
  observation remains historical and is not inferred or re-measured here.

## Provenance and preserved evidence

- Primary harness at diagnostic time: 4d77c511b27ea303fc5a65870dff1404011d1807.
- Secondary diagnostic commit: 4f2557062cb5c1ffb9a671bc3df67401fd7092b4.
- Original production Fast pin: 2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac;
  its detached pinned checkout was clean at final read-only verification.
- Genuine Standard pin: 5dbffbdea05394de2ca3a432ed5318aa832e3f40; the commit
  object exists, but its registered detached checkout is unavailable/prunable,
  so a fresh worktree status could not be collected. No Standard run or source
  modification occurred in this final session.
- Authentic Batch023 six-output vector SHA-256:
  63ad89ce9ccc70a0b737397810d4f4a15d8a1ed65e82ccd3af0435d1df4d6881.
  Its manifest SHA-256 is
  5bc19b1eb6e67997ca3d19f39e6fd781bf93b27ad305a1be733f0c883611b847.
- Raw, owner-only, unverified trace (539 events; 177,696 bytes), kept outside
  the repository:
  /var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2533882906/raw-trace-unverified.json
  (SHA-256 f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74).
- Failure sidecar SHA-256:
  59a872b2df18ca32859f4753650c3f1e14633dd79d1358b283204c4445b3f907.
  Both immutable attempt journals record irrevocably_reserved_before_call.

Batch023's 14/14 calls remain spent and were not reused. No Secondary
production code was changed in this final diagnostic session. The exact
stop, call ledger, fixture hashes, and raw evidence are recorded in the
companion journal and compact evidence.json.

## Validation and handoff

- go test ./cmd/fastdiag -count=1 — PASS.
- go vet ./cmd/fastdiag — PASS.
- E32 fixture validation with FASTDIAG_VALIDATE_ONLY=1 — PASS, zero calls.
- git diff --check — PASS.
- S4 event-tree validation — FAIL as described above; stop for Web review.

No further Bootstrap, benchmark, Standard run, or retry is authorized by this
batch. The trace failure is preserved for independent review; do not alter the
validator or rerun without a new decision.
