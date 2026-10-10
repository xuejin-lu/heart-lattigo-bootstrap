# Batch 020 execution journal

## Startup and frozen provenance

- Primary safe startup sync: `main`, initially clean at `c3096bd69c7379f737a8e19b6acf0242fc7ffac2`; fetched and fast-forwarded to `75e31d57fbee6b1b634246b18a0fcfb73e200adc`.
- Secondary: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean and unchanged after its own fetch / `merge --ff-only`.
- Genuine Standard checkout: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, detached, clean, read-only.
- Frozen application/profile/workload contract: accepted Batch 019 LogN13/E32 public chain, exact Q/P, default Scale, per-input C Scale=`q5`, and fixed `1e-6` maximum-error gate. No parameters, backend production code, or Standard source changed.
- Expensive Bootstrap budget at this checkpoint: Standard `0/1`, Fast `0/1`.

## S1 — reusable measurement harness: COMPLETE

Reused the accepted Batch 019 same-public-source runner and added only measurement instrumentation in Primary. Both backend builds compile the same common source; the build-tagged capacity adapter remains a separate observer and is not in timed operations. Added per-stage wall samples and allocation deltas for key generation, evaluation-key generation, evaluator/codec initialization, Encrypt A/B/C, Add, MulRelin, Rescale, Rotate, DropLevel, native decode, and the later one-shot Bootstrap/post-Bootstrap decode. Heap-stat snapshots occur before and after—not inside—the timed interval. `eval_only_stage_sum_wall_ns` is the sum of only the six timed public-chain stages from Add through native decode, excluding capacity observation/hashing. `process_active_wall_ns_excluding_stdin_wait` excludes the interactive one-shot gate interval. Each checkpoint also reports approximate q-backed bytes as the sum of coefficient-slice capacities × 8; this excludes object headers, keys, evaluator state, and allocator overhead and is not RSS.

Environment for the upcoming measurements is frozen to Go `go1.26.4`, `darwin/arm64`, Apple M4 / 10 CPU cores, `GOMAXPROCS=1`, Go's default GC with no `GOGC` or `GOMEMLIMIT` override, identical process-isolated runs, and a build-before-run workflow so compilation is excluded. The primary harness records Go/OS/arch, CPU count, GOMAXPROCS, and GC-related environment overrides. Peak RSS will be recorded only through macOS `/usr/bin/time -l` process high-water output, separately from q-backed bytes and stage allocations.

Validation before measurement:

- Fast focused `go test ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -count=1`: PASS.
- Standard focused `go test -tags lattigo_standard ... -count=1` against the pinned workspace: PASS.
- Fast and Standard focused `go vet`: PASS.
- `git diff --check`: PASS.
- No Bootstrap was invoked. No S2 repetitions have been counted yet.

Next: commit/push this Primary-only S1 harness checkpoint, then execute exactly five cheap process-isolated samples per backend for S2. Preserve each raw sample; do not call Bootstrap before S5.
