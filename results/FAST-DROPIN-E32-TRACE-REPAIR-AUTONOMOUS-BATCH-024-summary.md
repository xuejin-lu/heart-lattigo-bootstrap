# Batch 024 — E32 trace blocked before reservation

Status: `BATCH_BLOCKED_NEEDS_WEB_REVIEW`.

The Secondary Rescale instrumentation and S1–S3 zero-call gates remain
complete. The Primary event-analysis repair is committed as
`b39ae8519a98e88a902d7868924ae29eb721531c` and pushed to `origin/main`.
Existing per-event closure/Pareto fields and renderer/test interfaces are
preserved; additive event-type summaries use parent-minus-direct-children
exclusive time and keep each independent event-tree root separate.

The first S4 invocation stopped during fixture validation, before attempt
reservation, diagnostic worktree creation, or Bootstrap. The manifest expects
`fast-standard-public-native-repeatability-vectors.v1` in
`public-native-repeatability` mode with six `bootstrap_outputs`. The supplied
`fast-vectors.json` is `fast-standard-public-native-vectors.v1` in
`public-native` mode and has no `bootstrap_outputs`. Its backend, source,
config, Q/P, input, and workload hashes match the manifest, but it is not the
required repeatability artifact. The runner reported:

> `uninstrumented Fast vectors do not match the E32 trace fixture provenance`

No raw trace or attempt journal was created. Batch024 remains at **0/2
reserved calls and 0 Bootstrap calls** at that checkpoint. Batch023's **14/14
calls remain spent** and were not reused. No retry or alternate fixture was
attempted during that initial invocation.

The compact machine-readable provenance is in
`results/FAST-DROPIN-E32-TRACE-REPAIR-AUTONOMOUS-BATCH-024-evidence.json`.

## Web-authorized recovery and current stop (2026-10-11)

The authoritative six-output artifact was found at
`/private/tmp/fast-dropin-batch023-final.3VxTOp/fast-repeatability-vectors.json`
(SHA-256 `63ad89ce9ccc70a0b737397810d4f4a15d8a1ed65e82ccd3af0435d1df4d6881`).
It has the required repeatability schema/mode, all six named 4096-slot outputs
are finite, and its Fast, Primary, measurement-source, config, Q/P, input and
workload provenance matches the held manifest. The manifest hash remains
`5bc19b1eb6e67997ca3d19f39e6fd781bf93b27ad305a1be733f0c883611b847`; the
current Primary measurement-source fingerprint matches the recorded fixture
fingerprint `b63fef5bbbdfe1e679b2e50ab66e0a5c43afefc526bd2228b16b7541ae01ea63`.
The Secondary `FASTDIAG_VALIDATE_ONLY=1` test passed without Bootstrap.

The outer CLI compiled with the Web-authorized private cache
`/private/tmp/batch024-go-cache.TQR9yT`; `GOCACHE=... go test ./cmd/fastdiag
-run '^$' -count=1` also passed. The identical trace invocation passed Primary
fixture/source identity validation, then stopped while creating its isolated
Secondary diagnostic worktree. Git reported:

> `fatal: could not create directory of '.git/worktrees/secondary-diagnostic1': Operation not permitted`

This occurred before the Secondary test subprocess and before either token was
reserved. The selected output base and both `bootstrap-attempt-01/02` journal
paths remain absent; no raw/certified trace was produced. Current Batch024
ledger: **0/2 reserved, 0 Bootstrap calls**. Batch023 remains **14/14 spent**
and unchanged. No permission escalation or alternate worktree path was tried.

Status remains `BATCH_BLOCKED_NEEDS_WEB_REVIEW`; do not retry this diagnostic
without a new Web decision. Full provenance and the precise stop phase are in
the compact evidence JSON and journal.
