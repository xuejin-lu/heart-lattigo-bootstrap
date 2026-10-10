# Batch 020 execution journal

## Startup and frozen provenance

- Primary safe startup sync: `main`, initially clean at `c3096bd69c7379f737a8e19b6acf0242fc7ffac2`; fetched and fast-forwarded to `75e31d57fbee6b1b634246b18a0fcfb73e200adc`.
- Secondary: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean and unchanged after its own fetch / `merge --ff-only`.
- Genuine Standard checkout: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, detached, clean, read-only.
- Frozen application/profile/workload contract: accepted Batch 019 LogN13/E32 public chain, exact Q/P, default Scale, per-input C Scale=`q5`, and fixed `1e-6` maximum-error gate. No parameters, backend production code, or Standard source changed.
- Expensive Bootstrap budget at this checkpoint: Standard `0/1`, Fast `0/1`.

## S1 — reusable measurement harness: COMPLETE

Reused the accepted Batch 019 same-public-source runner and added only measurement instrumentation in Primary. Both backend builds compile the same common source; the build-tagged capacity adapter remains a separate observer and is not in timed operations. Added per-stage wall samples and allocation deltas for key generation, evaluation-key generation, evaluator/codec initialization, Encrypt A/B/C, Add, MulRelin, Rescale, Rotate, DropLevel, native decode, and the later one-shot Bootstrap/post-Bootstrap decode. Heap-stat snapshots occur before and after—not inside—the timed interval. `eval_only_stage_sum_wall_ns` is the sum of only the six timed public-chain stages from Add through native decode, excluding capacity observation/hashing. `process_active_wall_ns_excluding_stdin_wait` excludes the interactive one-shot gate interval. Each checkpoint also reports approximate q-backed bytes as the sum of coefficient-slice capacities × 8; this excludes object headers, keys, evaluator state, and allocator overhead and is not RSS.

Environment was frozen to Go `go1.26.4`, `darwin/arm64`, Apple M4 / 10 CPU cores, `GOMAXPROCS=1`, Go's default GC with no `GOGC` or `GOMEMLIMIT` override, identical process-isolated runs, and a build-before-run workflow so compilation is excluded. The primary harness records Go/OS/arch, CPU count, GOMAXPROCS, and GC-related environment overrides. Peak RSS collection was attempted with macOS `/usr/bin/time -l`, but sandbox denial of `sysctl kern.clockrate` made that tool exit nonzero. `time -p` works but does not report RSS, so peak RSS is omitted rather than inferred from q-backed bytes or Go allocations.

Validation before measurement:

- Fast focused `go test ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -count=1`: PASS.
- Standard focused `go test -tags lattigo_standard ... -count=1` against the pinned workspace: PASS.
- Fast and Standard focused `go vet`: PASS.
- `git diff --check`: PASS.
- No Bootstrap was invoked. An initial harness smoke process was excluded from S2 because its runtime `GOMAXPROCS` was 10; it also surfaced the `/usr/bin/time -l` sandbox limitation above.

## S2 — original-pins cheap baseline: COMPLETE

- Five valid process-isolated samples per backend ran with identical committed Primary source, config/input/profile, Go version, host and `GOMAXPROCS=1`; each Standard/Fast pair passed all provenance checks and the unchanged `1e-6` direct numerical gate. All ten pre-Bootstrap native oracles passed; total Bootstrap calls remain Standard `0/1`, Fast `0/1`.
- Full raw stage wall samples, median/range, allocation medians, paired RMSE/max, compact row-capacity byte estimates, provenance hashes and limitations are preserved in `FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-S2-original-baseline.md`.
- Descriptive eval-only stage-sum medians: Standard `6.102 ms`, Fast `3.178 ms`; highest Fast stage is Rescale at `1.730 ms`. This does not establish Bootstrap speedup or statistical significance.
- No reliable peak RSS value was available. Allocation counters and q-backed ciphertext bytes are kept as separate measurements.

Next: source-backed S3 attribution of Fast's Rescale bottleneck using current Secondary source and reusable historical diagnostics; no Bootstrap before S5.

## S3 — current Rescale attribution: COMPLETE

- Current public Fast dispatch and `fastcore.RescaleWorkspace` were reviewed at the frozen, clean Secondary pin `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- S2 identifies Fast Rescale as the largest timed eval-only stage (`1.730 ms` median). A bounded 8-second CPU profile of the existing rows4 LogN13 q0=55 Rescale benchmark measured `1,278,368 ns/op`, `680 B/op`, and `20 allocs/op`; the profile is a max-logical-level attribution proxy, not the public-chain q5 microbenchmark.
- Profile attribution points at required NTT/INTT and centered-CRT/round/capacity work; source review found no demonstrably redundant representation-neutral allocation/copy. Batched staging would retain the required transforms and arithmetic while changing intermediate liveness/cache behavior, with no evidence of a high-impact mechanical gain.
- Verdict: `NO_SAFE_OPT_CANDIDATE`. Optional S4 is skipped. No Secondary source was modified; S5 remains on the frozen original Fast commit. Detailed evidence is in `FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-S3-attribution.md`.
- Cumulative Bootstrap budget remains Standard `0/1`, Fast `0/1`.
