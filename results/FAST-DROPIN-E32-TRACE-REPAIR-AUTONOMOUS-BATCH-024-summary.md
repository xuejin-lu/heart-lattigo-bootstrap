# Batch 024 — E32 trace preflight blocker

Status: `BATCH_BLOCKED_NEEDS_WEB_REVIEW`.

The Secondary Rescale instrumentation and S1–S3 zero-call gates remain
complete. The Primary event-analysis repair is committed as
`b39ae8519a98e88a902d7868924ae29eb721531c` and pushed to `origin/main`.
Existing per-event closure/Pareto fields and renderer/test interfaces are
preserved; additive event-type summaries use parent-minus-direct-children
exclusive time and keep each independent event-tree root separate.

S4 stopped during fixture validation, before attempt reservation, diagnostic
worktree creation, or Bootstrap. The manifest expects
`fast-standard-public-native-repeatability-vectors.v1` in
`public-native-repeatability` mode with six `bootstrap_outputs`. The supplied
`fast-vectors.json` is `fast-standard-public-native-vectors.v1` in
`public-native` mode and has no `bootstrap_outputs`. Its backend, source,
config, Q/P, input, and workload hashes match the manifest, but it is not the
required repeatability artifact. The runner reported:

> `uninstrumented Fast vectors do not match the E32 trace fixture provenance`

No raw trace or attempt journal was created. Batch024 remains at **0/2
reserved calls and 0 Bootstrap calls**. Batch023's **14/14 calls remain spent**
and were not reused. No retry or alternate fixture was attempted.

The compact machine-readable provenance is in
`results/FAST-DROPIN-E32-TRACE-REPAIR-AUTONOMOUS-BATCH-024-evidence.json`.
Resume requires Web review to identify the authoritative matching
repeatability vectors artifact and clarify whether S4 may be invoked after
this pre-reservation validator failure.
